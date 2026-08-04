//go:build ignore

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate("https://www.chittorgarh.com/ipo/ipo_dashboard.asp"),
		chromedp.WaitVisible(`table`, chromedp.ByQuery),
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("page.html", []byte(htmlContent), 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Saved to page.html")
}
