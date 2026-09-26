package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type GMPData struct {
	GMPAmount      float64
	PremiumPercent float64
	PriceBand      float64 // The issue price from the ipowatch table (Col 3)
	ListingDate    string  // Extracted from the ipowatch detail page
	AboutCompany   string  // Extracted from the ipowatch detail page
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

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	htmlStr := string(bodyBytes)

	// ipowatch.in often has deeply nested formatting tags (e.g. <mark>, <strong>) that exceed 
	// the golang.org/x/net/html 512 max depth limit. We strip them out before parsing.
	reTags := regexp.MustCompile(`(?i)</?(strong|b|span|mark|em|i)[^>]*>`)
	cleanHTML := reTags.ReplaceAllString(htmlStr, "")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(cleanHTML))
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
		normalizedScraped = strings.ReplaceAll(normalizedScraped, "ipo", "")
		
		normalizedTarget := strings.ToLower(strings.ReplaceAll(ipoName, " ", ""))
		normalizedTarget = strings.ReplaceAll(normalizedTarget, "ipo", "")

		// Strip common suffixes that might be present in the DB but omitted by IPOWatch
		for _, suffix := range []string{"ltd", "limited", "company", "inc", "(india)"} {
			normalizedScraped = strings.ReplaceAll(normalizedScraped, suffix, "")
			normalizedTarget = strings.ReplaceAll(normalizedTarget, suffix, "")
		}

		// A match is found if either string contains the other (e.g. 'tempsensinstruments' is contained within 'tempsensinstruments(india)')
		// We add a basic length check to prevent false positives from tiny substrings
		if len(normalizedScraped) > 4 && len(normalizedTarget) > 4 && (strings.Contains(normalizedScraped, normalizedTarget) || strings.Contains(normalizedTarget, normalizedScraped)) {
			// Found it!
			// On IPOWatch: Col 0: Name, Col 1: GMP, Col 2: Trend, Col 3: Price Band, Col 4: Est Listing, Col 5: Date, Col 6: Type
			gmpCol := s.Find("td").Eq(1).Text()
			priceBandCol := s.Find("td").Eq(3).Text()
			estPremiumCol := s.Find("td").Eq(4).Text()

			gmpAmount := parseGMPValue(gmpCol)
			premiumPercent := parsePremiumPercent(estPremiumCol)
			priceBand := parseGMPValue(priceBandCol)

			var listingDate, aboutCompany string
			detailURL, exists := s.Find("td").Eq(0).Find("a").Attr("href")
			if exists && detailURL != "" {
				listingDate, aboutCompany, _ = fetchIPOWatchListingDate(ctx, detailURL)
			}

			foundGMP = &GMPData{
				GMPAmount:      gmpAmount,
				PremiumPercent: premiumPercent,
				PriceBand:      priceBand,
				ListingDate:    listingDate,
				AboutCompany:   aboutCompany,
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

func fetchIPOWatchListingDate(ctx context.Context, url string) (string, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("status code %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	htmlStr := string(bodyBytes)
	
	reTags := regexp.MustCompile(`(?i)</?(strong|b|span|mark|em|i)[^>]*>`)
	cleanHTML := reTags.ReplaceAllString(htmlStr, "")
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(cleanHTML))
	if err != nil {
		return "", "", err
	}

	var listingDate string
	doc.Find("figure.wp-block-table table tbody tr, table tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		label := strings.TrimSpace(s.Find("td").Eq(0).Text())
		if strings.Contains(strings.ToLower(label), "ipo listing date") {
			listingDate = strings.TrimSpace(s.Find("td").Eq(1).Text())
			return false // break
		}
		return true // continue
	})

	var aboutText string
	doc.Find("h2").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		if strings.HasPrefix(text, "about ") {
			// Find the next sibling p tags and grab their text
			nextP := s.Next()
			for nextP.Is("p") {
				aboutText += nextP.Text() + "\n"
				nextP = nextP.Next()
			}
			return false // break
		}
		return true // continue
	})

	// Fallback
	if aboutText == "" {
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 150 { // Only grab substantial paragraphs
				aboutText += text + "\n"
			}
		})
	}

	if len(aboutText) > 4000 {
		aboutText = aboutText[:4000]
	}

	return listingDate, aboutText, nil
}
