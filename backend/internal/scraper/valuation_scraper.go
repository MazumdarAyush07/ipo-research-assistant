package scraper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ValuationData struct {
	IssuePrice    float64
	PriceBandLow  float64
	PriceBandHigh float64
	MarketCap     float64
	PE            float64
	PB            float64
	Sector        string
	ListingDate   string
	AboutCompany  string
}

func FetchValuationData(ctx context.Context, url string) (ValuationData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ValuationData{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ValuationData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ValuationData{}, fmt.Errorf("chittorgarh returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ValuationData{}, err
	}

	var data ValuationData

	// Parse tables for Issue Price, Market Cap, P/E, P/B
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() >= 2 {
			label := strings.TrimSpace(tds.Eq(0).Text())
			value := strings.TrimSpace(tds.Eq(1).Text())

			// Examples:
			// Issue Price: ₹197 to ₹206 per share -> extract 206
			// Market Cap (Rs Cr.): 869.60 -> extract 869.60
			// P/E (x): 35.5 -> extract 35.5

			labelLower := strings.ToLower(label)

			if strings.Contains(labelLower, "issue price") || strings.Contains(labelLower, "price band") {
				// E.g., "₹197 to ₹206 per share"
				parts := strings.Split(value, "to")
				if len(parts) == 2 {
					data.PriceBandLow = extractFloat(parts[0])
					data.PriceBandHigh = extractFloat(parts[1])
					data.IssuePrice = data.PriceBandHigh
				} else {
					data.PriceBandLow = extractFloat(value)
					data.PriceBandHigh = extractFloat(value)
					data.IssuePrice = data.PriceBandHigh
				}
				log.Printf("DEBUG SCRAPE: IssuePrice matched. value=%q -> %f", value, data.IssuePrice)
			} else if strings.Contains(labelLower, "market cap") {
				log.Printf("DEBUG SCRAPE: MarketCap matched. label=%q, value=%q", label, value)
				val := extractFloat(value)
				if val > 0 {
					data.MarketCap = val
				}
			} else if strings.Contains(labelLower, "p/e") || strings.Contains(labelLower, "pe ") {
				val := extractFloat(value)
				if val > 0 {
					data.PE = val
				}
			} else if strings.Contains(labelLower, "p/b") || strings.Contains(labelLower, "pb ") {
				val := extractFloat(value)
				if val > 0 {
					data.PB = val
				}
			} else if strings.Contains(labelLower, "listing date") {
				// "Listing Date" usually looks like "August 24, 2026" or similar
				data.ListingDate = value
			} else if strings.Contains(labelLower, "industry") || strings.Contains(labelLower, "sector") {
				data.Sector = value
			}
		}
	})

	// Fallback to searching the full document text for Industry/Sector if missing from table
	if data.Sector == "" {
		fullText := doc.Text()
		lines := strings.Split(fullText, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "industry:") || strings.HasPrefix(lower, "industry :") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					data.Sector = strings.TrimSpace(parts[1])
					break
				}
			}
			if strings.HasPrefix(lower, "sector:") || strings.HasPrefix(lower, "sector :") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					data.Sector = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Capture text for AI inference (up to 5000 characters)
	fullText := doc.Text()
	if len(fullText) > 5000 {
		data.AboutCompany = fullText[:5000]
	} else {
		data.AboutCompany = fullText
	}

	return data, nil
}

func extractFloat(s string) float64 {
	// Clean up formatting
	s = strings.ReplaceAll(s, "₹", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "Cr.", "")
	s = strings.ReplaceAll(s, "Cr", "")
	s = strings.TrimSpace(s)
	// Some fields may have extra text like " per share", split by space and take first
	parts := strings.Split(s, " ")
	if len(parts) > 0 {
		val, _ := strconv.ParseFloat(parts[0], 64)
		return val
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
