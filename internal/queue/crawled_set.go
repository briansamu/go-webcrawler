package queue

import (
	"sync"
	"webcrawler/internal/utils"
)

type CrawledSet struct {
	data   map[uint64]bool
	number int
	mu     sync.Mutex
}

func NewCrawledSet() *CrawledSet {
	return &CrawledSet{
		data: make(map[uint64]bool),
	}
}

func (c *CrawledSet) Add(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[utils.HashUrl(url)] = true
	c.number++
}

func (c *CrawledSet) Contains(url string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[utils.HashUrl(url)]
}

func (c *CrawledSet) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.number
}
