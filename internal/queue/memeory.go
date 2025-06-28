package queue

import "sync"

type Queue struct {
	totalQueued int
	number      int
	elements    []string
	mu          sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		totalQueued: 0,
		number:      0,
		elements:    make([]string, 0),
	}
}

func (q *Queue) Enqueue(url string, crawled *CrawledSet) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if crawled.Contains(url) {
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

func (q *Queue) Dequeue() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	url := q.elements[0]
	q.elements = q.elements[1:]
	q.number--
	return url
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.number
}

func (q *Queue) TotalQueued() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.totalQueued
}
