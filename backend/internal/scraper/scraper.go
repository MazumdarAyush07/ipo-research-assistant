package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ScrapedIPO holds the parsed IPO data from any source.
type ScrapedIPO struct {
	Name         string
	ExchangeType string
	OpenDate     string
	CloseDate    string
	SourceUrl    string
}

// chittorgarhIPO maps the exact JSON fields from the webnodejs list-read API.
type chittorgarhIPO struct {
	ID                  int    `json:"id"`                    // e.g. 2510
	IpoNewsTitle        string `json:"ipo_news_title"`        // e.g. "Skyways Air IPO"
	IssueCategory       string `json:"issue_category"`        // "Mainline" or "SME"
	IpoPeriod           string `json:"ipo_period"`            // e.g. "24 Aug - 27 Aug"
	UrlrewriteFolderName string `json:"urlrewrite_folder_name"` // e.g. "skyways-air-ipo"
}

// chittorgarhResponse is the top-level JSON wrapper for list-read.
type chittorgarhResponse struct {
	IpoDropDownList []chittorgarhIPO `json:"ipoDropDownList"`
}

// ChittorgarhAPISource fetches IPO data from the Chittorgarh webnodejs JSON API.
// Endpoint discovered from network requests by https://www.chittorgarh.com/report/ipo-in-india-list-main-board-sme/82/all/
type ChittorgarhAPISource struct{}

const (
	chittorgarhListReadURL = "https://webnodejs.chittorgarh.com/cloud/ipo/list-read"
	chittorgarhBaseURL     = "https://www.chittorgarh.com/ipo/"
)

func (s *ChittorgarhAPISource) FetchUpcomingIPOs(ctx context.Context, _ string) ([]ScrapedIPO, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", chittorgarhListReadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", "https://www.chittorgarh.com")
	req.Header.Set("Referer", "https://www.chittorgarh.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var apiResp chittorgarhResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	log.Printf("Chittorgarh API returned %d IPOs", len(apiResp.IpoDropDownList))

	var ipos []ScrapedIPO
	for _, item := range apiResp.IpoDropDownList {
		// Parse "24 Aug - 27 Aug" into open and close dates
		// We reconstruct them as "24-Aug-YYYY" using the current year
		currentYear := time.Now().Year()
		openDate, closeDate := parsePeriod(item.IpoPeriod, currentYear)

		sourceUrl := ""
		if item.UrlrewriteFolderName != "" {
			// Real Chittorgarh URLs end with the ID, e.g. /ipo/skyways-air-ipo/2510/
			sourceUrl = fmt.Sprintf("%s%s/%d/", chittorgarhBaseURL, item.UrlrewriteFolderName, item.ID)
		}

		// Normalise "Mainline" -> "MAINBOARD" to match our DB schema
		exchangeType := strings.ToUpper(item.IssueCategory)
		if exchangeType == "MAINLINE" {
			exchangeType = "MAINBOARD"
		}

		ipos = append(ipos, ScrapedIPO{
			Name:         item.IpoNewsTitle,
			ExchangeType: exchangeType,
			OpenDate:     openDate,
			CloseDate:    closeDate,
			SourceUrl:    sourceUrl,
		})
	}

	return ipos, nil
}

// parsePeriod parses "24 Aug - 27 Aug" into ("24-Aug-2026", "27-Aug-2026").
func parsePeriod(period string, year int) (openDate, closeDate string) {
	parts := strings.Split(period, " - ")
	if len(parts) != 2 {
		return "", ""
	}

	// Each part is like "24 Aug" — convert to "24-Aug-YYYY"
	openDate = strings.ReplaceAll(strings.TrimSpace(parts[0]), " ", "-") + fmt.Sprintf("-%d", year)
	closeDate = strings.ReplaceAll(strings.TrimSpace(parts[1]), " ", "-") + fmt.Sprintf("-%d", year)
	return openDate, closeDate
}

// APISource acts as a last-resort fallback using a local JSON file.
type APISource struct {
	FallbackFilePath string
}

func (s *APISource) FetchUpcomingIPOs(ctx context.Context, url string) ([]ScrapedIPO, error) {
	log.Println("Using local JSON fallback...")
	data, err := os.ReadFile(s.FallbackFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fallback data: %w", err)
	}

	var ipos []ScrapedIPO
	if err := json.Unmarshal(data, &ipos); err != nil {
		return nil, fmt.Errorf("failed to parse fallback JSON: %w", err)
	}

	log.Printf("Fallback JSON returned %d IPOs", len(ipos))
	return ipos, nil
}

// FetchUpcomingIPOs is the main entry point for the processor.
func FetchUpcomingIPOs(ctx context.Context, url string) ([]ScrapedIPO, error) {
	primary := &ChittorgarhAPISource{}
	fallback := &APISource{FallbackFilePath: "internal/scraper/fallback_data.json"}

	ipos, err := primary.FetchUpcomingIPOs(ctx, url)
	if err == nil && len(ipos) > 0 {
		log.Printf("Successfully fetched %d IPOs from Chittorgarh JSON API", len(ipos))
		return ipos, nil
	}

	if err != nil {
		log.Printf("Chittorgarh API failed: %v. Falling back to local file...", err)
	} else {
		log.Printf("Chittorgarh API returned 0 IPOs. Falling back to local file...")
	}

	return fallback.FetchUpcomingIPOs(ctx, url)
}
