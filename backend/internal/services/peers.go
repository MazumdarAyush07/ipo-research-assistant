package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/redis/go-redis/v9"
)

// PeerConfig maps sector names to a list of NSE ticker symbols (without ".NS").
// For Screener.in, we use the symbol directly (e.g., "TRENT", not "TRENT.NS").
type PeerConfig map[string][]string

type PeerData struct {
	Ticker    string  `json:"ticker"`
	Name      string  `json:"name"`
	PE        float64 `json:"pe"`
	PB        float64 `json:"pb"`
	MarketCap float64 `json:"market_cap"` // in Crores
}

type PeerService struct {
	RedisClient *redis.Client
	Config      PeerConfig
}

func NewPeerService(redisClient *redis.Client, configPath string) (*PeerService, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not open peers config: %w", err)
	}
	defer file.Close()

	var config PeerConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("could not decode peers config: %w", err)
	}

	return &PeerService{
		RedisClient: redisClient,
		Config:      config,
	}, nil
}

func (s *PeerService) FetchPeersForSector(ctx context.Context, sector string) ([]PeerData, error) {
	tickers, ok := s.Config[sector]
	if !ok || len(tickers) == 0 {
		return nil, fmt.Errorf("no peers found for sector: %s", sector)
	}

	var results []PeerData
	for _, ticker := range tickers {
		data, err := s.fetchPeerData(ctx, ticker)
		if err != nil {
			log.Printf("Failed to fetch peer data for %s: %v", ticker, err)
			continue
		}
		results = append(results, data)
		// Polite delay to avoid hammering screener.in
		time.Sleep(500 * time.Millisecond)
	}

	return results, nil
}

func (s *PeerService) fetchPeerData(ctx context.Context, ticker string) (PeerData, error) {
	// Strip ".NS" suffix if present (config might have either format)
	cleanTicker := strings.TrimSuffix(strings.TrimSpace(ticker), ".NS")
	cacheKey := fmt.Sprintf("cache:peers:%s", cleanTicker)

	// Check Redis Cache first
	if s.RedisClient != nil {
		cachedData, err := s.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var pd PeerData
			if json.Unmarshal([]byte(cachedData), &pd) == nil {
				return pd, nil
			}
		}
	}

	// Fetch from Screener.in
	pd, err := scrapeScreenerIn(ctx, cleanTicker)
	if err != nil {
		return PeerData{}, fmt.Errorf("screener.in scrape failed for %s: %w", cleanTicker, err)
	}

	// Cache in Redis for 24 hours
	if s.RedisClient != nil {
		jsonPd, _ := json.Marshal(pd)
		s.RedisClient.Set(ctx, cacheKey, jsonPd, 24*time.Hour)
	}

	return pd, nil
}

// scrapeScreenerIn scrapes PE, Market Cap, and Book Value from screener.in for a given NSE symbol.
func scrapeScreenerIn(ctx context.Context, symbol string) (PeerData, error) {
	url := fmt.Sprintf("https://www.screener.in/company/%s/", symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return PeerData{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.google.com/")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return PeerData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return PeerData{}, fmt.Errorf("screener.in returned status %d for %s", resp.StatusCode, symbol)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return PeerData{}, err
	}

	pd := PeerData{Ticker: symbol}

	// Extract company name from <h1> or <title>
	pd.Name = strings.TrimSpace(doc.Find("h1").First().Text())
	if pd.Name == "" {
		title := doc.Find("title").Text()
		if idx := strings.Index(title, "|"); idx > 0 {
			pd.Name = strings.TrimSpace(title[:idx])
		}
	}

	// Parse the top-ratios section
	doc.Find("#top-ratios li").Each(func(i int, sel *goquery.Selection) {
		label := cleanText(sel.Find("span.name").Text())
		value := cleanText(sel.Find("span.nowrap").Text())
		if value == "" {
			// fallback: last span child
			sel.Find("span").Each(func(j int, sp *goquery.Selection) {
				v := cleanText(sp.Text())
				if v != "" {
					value = v
				}
			})
		}

		switch {
		case strings.Contains(label, "Market Cap"):
			pd.MarketCap = parseIndianNumber(value)
		case strings.EqualFold(label, "Stock P/E"):
			pd.PE = parseFloat(value)
		case strings.Contains(label, "Book Value"):
			// PB = Current Price / Book Value — we store book value, frontend can compute PB
			// For now store as PB field with approximate ratio using price if available
			pd.PB = parseFloat(value)
		}
	})

	if pd.PE == 0 && pd.MarketCap == 0 {
		return PeerData{}, fmt.Errorf("no data parsed from screener.in for %s", symbol)
	}

	return pd, nil
}

var spaceRegex = regexp.MustCompile(`\s+`)

func cleanText(s string) string {
	s = spaceRegex.ReplaceAllString(strings.TrimSpace(s), " ")
	// Remove currency symbols
	s = strings.ReplaceAll(s, "₹", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.TrimSpace(s)
	return s
}

// parseIndianNumber handles values like "1,39,316" (Crores) → 139316.0
func parseIndianNumber(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "Cr.", "")
	s = strings.ReplaceAll(s, "Cr", "")
	s = strings.TrimSpace(s)
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	val, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return val
}

func (s *PeerService) parseYahooResponse(ctx context.Context, _ interface{}, ticker, cacheKey string) (PeerData, error) {
	// Kept for interface compatibility — not used anymore
	return PeerData{}, fmt.Errorf("not implemented")
}
