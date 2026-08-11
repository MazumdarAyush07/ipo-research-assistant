package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type PeerConfig map[string][]string

type PeerData struct {
	Ticker    string  `json:"ticker"`
	Name      string  `json:"name"`
	PE        float64 `json:"pe"`
	PB        float64 `json:"pb"`
	MarketCap float64 `json:"market_cap"`
}

type PeerService struct {
	RedisClient *redis.Client
	Config      PeerConfig
	
	// Yahoo Finance Auth
	crumb       string
	cookies     []*http.Cookie
	authMutex   sync.Mutex
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

	ps := &PeerService{
		RedisClient: redisClient,
		Config:      config,
	}
	
	// Initialize Yahoo Auth
	_ = ps.refreshYahooAuth(context.Background())

	return ps, nil
}

func (s *PeerService) refreshYahooAuth(ctx context.Context) error {
	s.authMutex.Lock()
	defer s.authMutex.Unlock()

	client := &http.Client{Timeout: 10 * time.Second}
	
	// 1. Get cookies from fc.yahoo.com
	req1, _ := http.NewRequestWithContext(ctx, "GET", "https://fc.yahoo.com", nil)
	req1.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp1, err := client.Do(req1)
	if err != nil {
		return err
	}
	defer resp1.Body.Close()
	s.cookies = resp1.Cookies()

	// 2. Get crumb from getcrumb
	req2, _ := http.NewRequestWithContext(ctx, "GET", "https://query1.finance.yahoo.com/v1/test/getcrumb", nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	for _, c := range s.cookies {
		req2.AddCookie(c)
	}
	
	resp2, err := client.Do(req2)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	
	crumbBytes, _ := io.ReadAll(resp2.Body)
	s.crumb = string(crumbBytes)
	
	return nil
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
	}

	return results, nil
}

func (s *PeerService) fetchPeerData(ctx context.Context, ticker string) (PeerData, error) {
	cacheKey := fmt.Sprintf("cache:peers:%s", ticker)
	
	// Check Redis Cache
	if s.RedisClient != nil {
		cachedData, err := s.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var pd PeerData
			if json.Unmarshal([]byte(cachedData), &pd) == nil {
				return pd, nil
			}
		}
	}

	// Fetch from Yahoo Finance
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s&crumb=%s", ticker, s.crumb)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	s.authMutex.Lock()
	for _, c := range s.cookies {
		req.AddCookie(c)
	}
	s.authMutex.Unlock()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return PeerData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		// Retry once after refreshing auth
		_ = s.refreshYahooAuth(ctx)
		return s.fetchPeerDataRetry(ctx, ticker)
	}

	if resp.StatusCode != http.StatusOK {
		return PeerData{}, fmt.Errorf("yahoo finance returned status %d", resp.StatusCode)
	}

	return s.parseYahooResponse(ctx, resp.Body, ticker, cacheKey)
}

func (s *PeerService) fetchPeerDataRetry(ctx context.Context, ticker string) (PeerData, error) {
	cacheKey := fmt.Sprintf("cache:peers:%s", ticker)
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s&crumb=%s", ticker, s.crumb)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	s.authMutex.Lock()
	for _, c := range s.cookies {
		req.AddCookie(c)
	}
	s.authMutex.Unlock()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return PeerData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PeerData{}, fmt.Errorf("yahoo finance retry returned status %d", resp.StatusCode)
	}
	
	return s.parseYahooResponse(ctx, resp.Body, ticker, cacheKey)
}

func (s *PeerService) parseYahooResponse(ctx context.Context, bodyReader io.Reader, ticker, cacheKey string) (PeerData, error) {
	body, _ := io.ReadAll(bodyReader)
	
	var yfResp struct {
		QuoteResponse struct {
			Result []struct {
				ShortName  string  `json:"shortName"`
				TrailingPE float64 `json:"trailingPE"`
				PriceToBook float64 `json:"priceToBook"`
				MarketCap   float64 `json:"marketCap"`
			} `json:"result"`
		} `json:"quoteResponse"`
	}

	if err := json.Unmarshal(body, &yfResp); err != nil {
		return PeerData{}, err
	}

	if len(yfResp.QuoteResponse.Result) == 0 {
		return PeerData{}, fmt.Errorf("no quote data returned for %s", ticker)
	}

	res := yfResp.QuoteResponse.Result[0]
	pd := PeerData{
		Ticker:    ticker,
		Name:      res.ShortName,
		PE:        res.TrailingPE,
		PB:        res.PriceToBook,
		MarketCap: res.MarketCap,
	}

	// Cache in Redis for 24 hours
	if s.RedisClient != nil {
		jsonPd, _ := json.Marshal(pd)
		s.RedisClient.Set(ctx, cacheKey, jsonPd, 24*time.Hour)
	}

	return pd, nil
}
