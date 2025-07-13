package stats

import (
	"fmt"
	"time"

	"webcrawler/internal/queue"
)

type CrawlerStats struct {
	pagesPerMinute        string // 0 0 \n 1 100
	crawledRatioPerMinute string
	startTime             time.Time
}

func NewCrawlerStats() *CrawlerStats {
	return &CrawlerStats{
		pagesPerMinute:        "0 0\n",
		crawledRatioPerMinute: "0 0\n",
		startTime:             time.Now(),
	}
}

func (stats *CrawlerStats) Update(crawled *queue.CrawledSet, queue *queue.Queue, t time.Time) {
	stats.pagesPerMinute = fmt.Sprintf("%f %d\n", time.Since(stats.startTime).Minutes(), crawled.Size())
	stats.crawledRatioPerMinute = fmt.Sprintf("%f %f\n", time.Since(stats.startTime).Minutes(), float64(crawled.Size())/float64(queue.Size()))
}

func (stats *CrawlerStats) Print() {
	fmt.Println("Pages crawled per minute:")
	fmt.Println(stats.pagesPerMinute)
	fmt.Println("Crawl to Queued Ratio per minute:")
	fmt.Println(stats.crawledRatioPerMinute)
}

func (stats *CrawlerStats) GetStartTime() time.Time {
	return stats.startTime
}

func (stats *CrawlerStats) GetCurrentStats() map[string]interface{} {
	return map[string]interface{}{
		"startTime":             stats.startTime,
		"pagesPerMinute":        stats.pagesPerMinute,
		"crawledRatioPerMinute": stats.crawledRatioPerMinute,
		"uptime":                time.Since(stats.startTime),
	}
}
