package crawler

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func FetchPage(url string, c chan []byte) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var content string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		// chromedp.Sleep(1500*time.Millisecond), // Wait for hydration
		chromedp.OuterHTML(`html`, &content, chromedp.ByQuery),
	)

	if err != nil {
		fmt.Println("Error fetching page:", err)
		c <- []byte("")
		return
	}

	c <- []byte(content)
}
