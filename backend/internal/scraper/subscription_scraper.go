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

type SubscriptionData struct {
	QIB    float64
	NII    float64
	Retail float64
	Total  float64
}

// FetchSubscriptionData scrapes Chittorgarh's specific IPO page for subscription numbers.
func FetchSubscriptionData(ctx context.Context, sourceUrl string) (*SubscriptionData, error) {
	if sourceUrl == "" {
		return nil, fmt.Errorf("no source URL provided")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", sourceUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

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

	data := &SubscriptionData{}

	// Chittorgarh typically displays subscription in a table where rows contain categories like "QIB", "NII", "Retail"
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		cols := s.Find("td")
		if cols.Length() >= 2 {
			category := strings.ToLower(cols.Eq(0).Text())
			valueStr := cols.Eq(1).Text() // usually times subscribed is in the second column or later depending on table
			
			// Try to find the specific columns containing the "times"
			// Usually Chittorgarh subscription table has columns: Category | Subscription Status
			// So let's extract the number.
			val := parseTimesSubscribed(valueStr)

			if strings.Contains(category, "qualified") || strings.Contains(category, "qib") {
				data.QIB = val
			} else if strings.Contains(category, "non-institutional") || strings.Contains(category, "nii") {
				data.NII = val
			} else if strings.Contains(category, "retail") {
				data.Retail = val
			} else if strings.Contains(category, "total") {
				data.Total = val
			}
		}
	})

	// If everything is 0, it means we probably didn't find the table or the IPO is not open yet
	if data.QIB == 0 && data.NII == 0 && data.Retail == 0 && data.Total == 0 {
		return nil, fmt.Errorf("subscription data not found or zero on page")
	}

	return data, nil
}

func parseTimesSubscribed(s string) float64 {
	// e.g. "1.52x", "1.52", "0.00"
	clean := strings.TrimSpace(strings.ReplaceAll(s, "x", ""))
	clean = strings.ReplaceAll(clean, "Times", "")
	clean = strings.TrimSpace(clean)
	
	if clean == "-" || clean == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(clean, 64)
	return val
}
