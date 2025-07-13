package main

import (
	"fmt"
	"time"

	"webcrawler/internal/api"
	"webcrawler/internal/config"
	"webcrawler/internal/crawler"
	"webcrawler/internal/queue"
	"webcrawler/internal/robots"
	"webcrawler/internal/stats"
	"webcrawler/internal/storage"
)

func main() {
	cfg := config.Load()

	db := storage.NewMongoDB(cfg.DBAccess, cfg.MongoURI)
	db.Connect()

	crawled := queue.NewCrawledSet()
	q := queue.NewQueue()
	robotsChecker := robots.NewRobotsChecker(cfg.UserAgent)
	crawlerStats := stats.NewCrawlerStats()

	// Start API server in goroutine
	apiServer := api.NewAPIServer(db, crawlerStats, crawled, q)
	go apiServer.Start("8080")

	ticker := time.NewTicker(1 * time.Minute)
	done := make(chan bool)

	// Tick every minute and broadcast stats
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				crawlerStats.Update(crawled, q, t)
				apiServer.BroadcastStats()
			}
		}
	}()

	// Also broadcast stats every 5 seconds for more responsive UI
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				apiServer.BroadcastStats()
			}
		}
	}()

	q.Enqueue(cfg.SeedURL, crawled)
	url := q.Dequeue()
	crawled.Add(url)
	c := make(chan []byte)

	// Check robots.txt before fetching
	if allowed, crawlDelay := robotsChecker.IsAllowed(url); allowed {
		if crawlDelay > 0 {
			time.Sleep(crawlDelay)
		}
		go crawler.FetchPage(url, c)

		content := <-c
		crawler.ParsePage(url, content, q, crawled, db, robotsChecker)
	} else {
		fmt.Printf("Robots.txt disallows crawling: %s\n", url)
		c <- []byte("") // Send empty content to continue flow
	}

	for q.Size() > 0 && crawled.Size() < 5000 {
		url := q.Dequeue()
		crawled.Add(url)

		// Check robots.txt before fetching
		allowed, crawlDelay := robotsChecker.IsAllowed(url)
		if !allowed {
			fmt.Printf("Robots.txt disallows crawling: %s\n", url)
			continue
		}

		// Respect crawl delay
		if crawlDelay > 0 {
			time.Sleep(crawlDelay)
		}

		go crawler.FetchPage(url, c)
		content := <-c
		if len(content) == 0 {
			continue
		}
		crawler.ParsePage(url, content, q, crawled, db, robotsChecker)
	}

	ticker.Stop()
	done <- true
	db.Disconnect()
	fmt.Println("\n------------------CRAWLER STATS------------------")
	fmt.Printf("Total queued: %d\n", q.TotalQueued())
	fmt.Printf("To be crawled (Queue) size: %d\n", q.Size())
	fmt.Printf("Crawled size: %d\n", crawled.Size())
	crawlerStats.Print()
}
