package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type GMPData struct {
	GMPAmount      float64
	PremiumPercent float64
}

// FetchGMPData scrapes the GMP from Chittorgarh or InvestorGain.
// Since Chittorgarh puts everything on a single IPO page, we can search the GMP list page.
func FetchGMPData(ctx context.Context, ipoName string) (*GMPData, error) {
	url := "https://ipowatch.in/ipo-grey-market-premium-latest-ipo-gmp/"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find("title").Text())
	rows := doc.Find("figure.wp-block-table table tbody tr, table tbody tr").Length()
	fmt.Printf("IPOWatch GMP Scraper - Title: %s | Rows: %d\n", title, rows)

	var foundGMP *GMPData
	
	doc.Find("figure.wp-block-table table tbody tr, table tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// First column usually contains the IPO Name
		nameCol := s.Find("td").Eq(0).Text()
		
		normalizedScraped := strings.ToLower(strings.ReplaceAll(nameCol, " ", ""))
		normalizedTarget := strings.ToLower(strings.ReplaceAll(ipoName, " ", ""))
		normalizedTarget = strings.ReplaceAll(normalizedTarget, "ipo", "")

		if strings.Contains(normalizedScraped, normalizedTarget) {
			// Found it!
			// On IPOWatch: Col 0: Name, Col 1: GMP, Col 2: Trend, Col 3: Price, Col 4: Est Listing
			gmpCol := s.Find("td").Eq(1).Text() 
			estPremiumCol := s.Find("td").Eq(4).Text()

			gmpAmount := parseGMPValue(gmpCol)
			premiumPercent := parsePremiumPercent(estPremiumCol)

			foundGMP = &GMPData{
				GMPAmount:      gmpAmount,
				PremiumPercent: premiumPercent,
			}
			return false // break
		}
		return true // continue
	})

	if foundGMP != nil {
		return foundGMP, nil
	}
	
	return nil, fmt.Errorf("GMP not found for IPO: %s", ipoName)
}

func parseGMPValue(s string) float64 {
	// e.g. "₹ 15", "15", "-"
	clean := strings.TrimSpace(strings.ReplaceAll(s, "₹", ""))
	if clean == "-" || clean == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(clean, 64)
	return val
}

func parsePremiumPercent(s string) float64 {
	// e.g. "₹168 (14.49%)", "25.50%", "-"
	
	// If it has parenthesis, extract the value inside
	idx1 := strings.Index(s, "(")
	idx2 := strings.Index(s, "%")
	
	var clean string
	if idx1 != -1 && idx2 != -1 && idx2 > idx1 {
		clean = s[idx1+1 : idx2]
	} else {
		clean = strings.TrimSpace(strings.ReplaceAll(s, "%", ""))
		if idx := strings.Index(clean, " "); idx != -1 {
			clean = clean[:idx]
		}
	}

	clean = strings.ReplaceAll(clean, "₹", "")
	if clean == "-" || clean == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(clean, 64)
	return val
}
