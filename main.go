package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func ParseHtml(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	return body
}

func main() {
	pageBody := ParseHtml("https://samuinteractive.com")
	fmt.Println(string(pageBody))
}

// Add page to queue
// Get page body
// Parse page body for links
// Add links to queue
// Add page to visited
// Repeat until no more pages to visit
