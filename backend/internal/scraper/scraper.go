package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type ScrapedIPO struct {
	Name         string
	ExchangeType string
	OpenDate     string
	CloseDate    string
	SourceUrl    string
}

// IPOSource defines an interface for fetching IPOs
type IPOSource interface {
	FetchUpcomingIPOs(ctx context.Context, url string) ([]ScrapedIPO, error)
}

// ChromedpSource uses a headless browser to render JS-hydrated tables
type ChromedpSource struct{}

func (s *ChromedpSource) FetchUpcomingIPOs(ctx context.Context, url string) ([]ScrapedIPO, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Timeout to prevent hanging
	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		// Wait for the table to be rendered by Next.js
		chromedp.WaitVisible(`table`, chromedp.ByQuery),
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp failed to fetch page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	return parseHTMLTable(doc), nil
}

// APISource acts as a fallback using a local JSON file (or a real API later)
type APISource struct {
	FallbackFilePath string
}

func (s *APISource) FetchUpcomingIPOs(ctx context.Context, url string) ([]ScrapedIPO, error) {
	log.Println("Using APISource fallback...")
	data, err := os.ReadFile(s.FallbackFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fallback data: %w", err)
	}

	var ipos []ScrapedIPO
	if err := json.Unmarshal(data, &ipos); err != nil {
		return nil, fmt.Errorf("failed to parse fallback JSON: %w", err)
	}

	log.Printf("Fallback API returned %d IPOs", len(ipos))
	return ipos, nil
}

// FetchUpcomingIPOs is the main entry point for the processor.
// It tries Chromedp first, and falls back to APISource if it fails.
func FetchUpcomingIPOs(ctx context.Context, url string) ([]ScrapedIPO, error) {
	primary := &ChromedpSource{}
	fallback := &APISource{FallbackFilePath: "fallback_data.json"}

	ipos, err := primary.FetchUpcomingIPOs(ctx, url)
	if err == nil && len(ipos) > 0 {
		log.Printf("Successfully scraped %d IPOs using Headless Chrome", len(ipos))
		return ipos, nil
	}

	if err != nil {
		log.Printf("Primary scraper failed: %v. Falling back to API...", err)
	} else {
		log.Printf("Primary scraper returned 0 IPOs. Falling back to API just in case...")
	}

	return fallback.FetchUpcomingIPOs(ctx, url)
}

// parseHTMLTable extracts IPO data from the loaded goquery document
func parseHTMLTable(doc *goquery.Document) []ScrapedIPO {
	var ipos []ScrapedIPO
	
	// Identify the main table on the report page by checking for the "Issue Category" header
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		isIPOTable := false
		table.Find("th").Each(func(j int, th *goquery.Selection) {
			if strings.Contains(strings.ToLower(strings.TrimSpace(th.Text())), "issue category") {
				isIPOTable = true
			}
		})

		if !isIPOTable {
			return
		}

		table.Find("tr").Each(func(k int, row *goquery.Selection) {
			cols := row.Find("td")
			// The report table has around 10 columns
			if cols.Length() >= 5 {
				name := strings.TrimSpace(cols.Eq(0).Text())
				sourceUrl, _ := cols.Eq(0).Find("a").Attr("href")
				exchangeType := strings.ToUpper(strings.TrimSpace(cols.Eq(1).Text()))
				openDate := strings.TrimSpace(cols.Eq(3).Text())
				closeDate := strings.TrimSpace(cols.Eq(4).Text())

				if name == "" {
					return
				}

				ipos = append(ipos, ScrapedIPO{
					Name:         name,
					ExchangeType: exchangeType,
					OpenDate:     openDate,
					CloseDate:    closeDate,
					SourceUrl:    sourceUrl,
				})
			}
		})
	})
	
	return ipos
}
