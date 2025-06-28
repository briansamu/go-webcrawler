package main

import (
	"fmt"
	"time"

	"webcrawler/internal/config"
	"webcrawler/internal/crawler"
	"webcrawler/internal/queue"
	"webcrawler/internal/stats"
	"webcrawler/internal/storage"
)

func main() {
	cfg := config.Load()

	db := storage.NewMongoDB(cfg.DBAccess, cfg.MongoURI)
	db.Connect()

	crawled := queue.NewCrawledSet()
	q := queue.NewQueue()

	ticker := time.NewTicker(1 * time.Minute)
	done := make(chan bool)
	crawlerStats := stats.NewCrawlerStats()

	// Tick every minute
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				crawlerStats.Update(crawled, q, t)
			}
		}
	}()

	q.Enqueue(cfg.SeedURL, crawled)
	url := q.Dequeue()
	crawled.Add(url)
	c := make(chan []byte)

	go crawler.FetchPage(url, c)

	content := <-c
	crawler.ParsePage(url, content, q, crawled, db)

	for q.Size() > 0 && crawled.Size() < 5000 {
		url := q.Dequeue()
		crawled.Add(url)

		go crawler.FetchPage(url, c)
		content := <-c
		if len(content) == 0 {
			continue
		}
		go crawler.ParsePage(url, content, q, crawled, db)
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
