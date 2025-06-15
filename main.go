package main

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/html"
)

type Queue struct {
	totalQueued int
	number      int
	elements    []string
	mu          sync.Mutex
}

func (q *Queue) enqueue(url string, crawled *CrawledSet) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if crawled.contains(url) {
		return
	}

	for _, existing := range q.elements {
		if existing == url {
			return
		}
	}

	q.elements = append(q.elements, url)
	q.totalQueued++
	q.number++
}

func (q *Queue) dequeue() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	url := q.elements[0]
	q.elements = q.elements[1:]
	q.number--
	return url
}

func (q *Queue) size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.number
}

type CrawledSet struct {
	data   map[uint64]bool
	number int
	mu     sync.Mutex
}

func (c *CrawledSet) add(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[hashUrl(url)] = true
	c.number++
}

func (c *CrawledSet) contains(url string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[hashUrl(url)]
}

func (c *CrawledSet) size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.number
}

type DatabaseConnection struct {
	access     bool
	uri        string
	client     *mongo.Client
	collection *mongo.Collection
}

func (db *DatabaseConnection) connect() {
	if db.access {
		db.uri = os.Getenv("MONGO_URI")
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(db.uri))
		if err != nil {
			panic(err)
		}
		db.client = client
		db.collection = db.client.Database("webcrawler").Collection("pages")
		filter := bson.D{{}}
		// Deletes all documents in the collection
		db.collection.DeleteMany(context.TODO(), filter)
	}
}

func (db *DatabaseConnection) disconnect() {
	if db.access {
		db.client.Disconnect(context.TODO())
		db.access = false
	}
}

func (db *DatabaseConnection) insertPage(page Page) {
	if db.access {
		db.collection.InsertOne(context.TODO(), page)
	}
}

type Page struct {
	Url     string
	Title   string
	Content string
}

func hashUrl(url string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(url))
	return h.Sum64()
}

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			if len(a.Val) == 0 || !strings.HasPrefix(a.Val, "http") {
				ok = false
				href = a.Val
				return ok, href
			}
			href = a.Val
			ok = true
		}
	}
	return ok, href
}

func fetchPage(url string, c chan []byte) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var content string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond), // Wait for hydration
		chromedp.OuterHTML(`html`, &content, chromedp.ByQuery),
	)

	if err != nil {
		fmt.Println("Error fetching page:", err)
		c <- []byte("")
		return
	}

	c <- []byte(content)
}

func parsePage(currUrl string, content []byte, q *Queue, crawled *CrawledSet, db *DatabaseConnection) {
	z := html.NewTokenizer(bytes.NewReader(content))
	tokenCount := 0
	pageContentLength := 0
	body := false
	page := Page{Url: currUrl, Title: "", Content: ""}

	for {
		if z.Next() == html.ErrorToken || tokenCount > 500 {
			if crawled.size() < 1000 {
				db.insertPage(page)
			}
			return
		}
		t := z.Token()
		if t.Type == html.StartTagToken {
			if t.Data == "body" {
				body = true
			}
			if t.Data == "javascript" || t.Data == "script" || t.Data == "style" {
				z.Next()
				continue
			}
			if t.Data == "title" {
				z.Next()
				title := z.Token().Data
				page.Title = title
				fmt.Printf("Count: %d | %s -> %s\n", crawled.size(), currUrl, title)
			}
			if t.Data == "a" {
				ok, href := getHref(t)
				if !ok {
					continue
				}
				if crawled.contains(href) {
					continue
				} else {
					q.enqueue(href, crawled)
				}
			}
		}
		if body && t.Type == html.TextToken && pageContentLength < 500 {
			page.Content += strings.TrimSpace(t.Data)
			pageContentLength += len(t.Data)
		}
		tokenCount++
	}
}

func main() {
	dbAccess := true

	if godotenv.Load() != nil {
		fmt.Println("Error loading .env file")
		dbAccess = false
	}

	db := DatabaseConnection{access: dbAccess, uri: "", client: nil, collection: nil}
	db.connect()

	crawled := CrawledSet{data: make(map[uint64]bool)}
	seed := os.Getenv("SEED_URL")
	queue := Queue{totalQueued: 0, number: 0, elements: make([]string, 0)}

	ticker := time.NewTicker(1 * time.Minute)
	done := make(chan bool)
	crawlerStats := CrawlerStats{pagesPerMinute: "0 0\n", crawledRatioPerMinute: "0 0\n", startTime: time.Now()}

	// Tick every minute
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				crawlerStats.update(&crawled, &queue, t)
			}
		}
	}()
	queue.enqueue(seed, &crawled)
	url := queue.dequeue()
	crawled.add(url)
	c := make(chan []byte)

	go fetchPage(url, c)

	content := <-c
	parsePage(url, content, &queue, &crawled, &db)

	for queue.size() > 0 && crawled.size() < 5000 {
		url := queue.dequeue()
		crawled.add(url)

		go fetchPage(url, c)
		content := <-c
		if len(content) == 0 {
			continue
		}
		go parsePage(url, content, &queue, &crawled, &db)
	}
	ticker.Stop()
	done <- true
	db.disconnect()
	fmt.Println("\n------------------CRAWLER STATS------------------")
	fmt.Printf("Total queued: %d\n", queue.totalQueued)
	fmt.Printf("To be crawled (Queue) size: %d\n", queue.size())
	fmt.Printf("Crawled size: %d\n", crawled.size())
	crawlerStats.print()
}

type CrawlerStats struct {
	pagesPerMinute        string // 0 0 \n 1 100
	crawledRatioPerMinute string
	startTime             time.Time
}

func (stats *CrawlerStats) update(crawled *CrawledSet, queue *Queue, t time.Time) {
	stats.pagesPerMinute = fmt.Sprintf("%f %d\n", t.Sub(stats.startTime).Minutes(), crawled.size())
	stats.crawledRatioPerMinute = fmt.Sprintf("%f %f\n", t.Sub(stats.startTime).Minutes(), float64(crawled.size())/float64(queue.size()))
}

func (stats *CrawlerStats) print() {
	fmt.Println("Pages crawled per minute:")
	fmt.Println(stats.pagesPerMinute)
	fmt.Println("Crawl to Queued Ratio per minute:")
	fmt.Println(stats.crawledRatioPerMinute)
}
