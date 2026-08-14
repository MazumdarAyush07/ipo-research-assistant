package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url := "https://www.chittorgarh.com/report/ipo-in-india-list-main-board-sme/82/all/"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error fetching: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing html: %v", err)
	}

	count := 0
	doc.Find("table tr").Each(func(i int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() >= 5 {
			name := strings.TrimSpace(cols.Eq(0).Text())
			if name != "" {
				count++
				if count <= 5 {
					log.Printf("Found IPO: %s | Open: %s | Close: %s", name, strings.TrimSpace(cols.Eq(3).Text()), strings.TrimSpace(cols.Eq(4).Text()))
				}
			}
		}
	})
	log.Printf("Total IPOs extracted: %d", count)
}
