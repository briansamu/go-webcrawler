package crawler

import (
	"bytes"
	"fmt"
	"strings"

	"webcrawler/internal/models"
	"webcrawler/internal/queue"
	"webcrawler/internal/storage"
	"webcrawler/internal/utils"

	"golang.org/x/net/html"
)

func ParsePage(currUrl string, content []byte, q *queue.Queue, crawled *queue.CrawledSet, db *storage.MongoDB) {
	z := html.NewTokenizer(bytes.NewReader(content))
	tokenCount := 0
	pageContentLength := 0
	body := false
	page := models.Page{Url: currUrl, Title: "", Content: ""}

	for {
		if z.Next() == html.ErrorToken || tokenCount > 25000 {
			if crawled.Size() < 1000 {
				db.InsertPage(page)
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
				fmt.Printf("Count: %d | %s -> %s\n", crawled.Size(), currUrl, title)
			}
			if t.Data == "a" {
				ok, href := utils.GetHref(t, currUrl)
				if !ok {
					continue
				}
				if crawled.Contains(href) {
					continue
				} else {
					q.Enqueue(href, crawled)
				}
			}
		}
		if body && t.Type == html.TextToken && pageContentLength < 15000 {
			page.Content += strings.TrimSpace(t.Data)
			pageContentLength += len(t.Data)
		}
		tokenCount++
	}
}
