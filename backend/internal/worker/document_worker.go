package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/downloader"
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/hibiken/asynq"
)

const TaskDownloadDocuments = "document:download"

type DownloadDocumentsPayload struct {
	IPOID int64 `json:"ipo_id"`
}

// HandleDownloadDocumentsTask processes a document download job.
func (p *Processor) HandleDownloadDocumentsTask(ctx context.Context, t *asynq.Task) error {
	var payload DownloadDocumentsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Fetch IPO from DB to get the SourceUrl
	ipo, err := p.Queries.GetIPO(ctx, payload.IPOID)
	if err != nil {
		return fmt.Errorf("failed to fetch IPO %d: %v", payload.IPOID, err)
	}

	if !ipo.SourceUrl.Valid || ipo.SourceUrl.String == "" {
		log.Printf("IPO %s has no SourceUrl, skipping download", ipo.Name)
		return nil
	}

	// Create a safe slug for the folder name
	slug := strings.ToLower(strings.ReplaceAll(ipo.Name, " ", "-"))
	slug = strings.ReplaceAll(slug, ".", "")

	log.Printf("Scraping detail page for %s: %s", ipo.Name, ipo.SourceUrl.String)
	drhpUrl, err := findDRHPLink(ipo.SourceUrl.String)
	if err != nil {
		log.Printf("Could not find DRHP link for %s: %v", ipo.Name, err)
		return err // Will be retried by Asynq
	}

	destPath := filepath.Join("../storage", slug, "drhp.pdf")
	log.Printf("Downloading DRHP to %s", destPath)

	// Download PDF (require at least 100KB)
	err = downloader.DownloadPDF(drhpUrl, destPath, 100*1024)
	if err != nil {
		return fmt.Errorf("failed to download DRHP for %s: %v", ipo.Name, err)
	}

	// Insert record into documents table
	_, err = p.Queries.CreateDocument(ctx, models.CreateDocumentParams{
		IpoID:        ipo.ID,
		Type:         "DRHP",
		FilePath:     destPath,
		DownloadedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		log.Printf("Failed to record document in DB for %s: %v", ipo.Name, err)
	}

	log.Printf("Successfully downloaded DRHP for IPO ID: %d", payload.IPOID)

	// Now enqueue the parsing job for this document
	parsePayload, _ := json.Marshal(ParseDocumentPayload{
		IPOID:    payload.IPOID,
		FilePath: destPath,
	})
	parseTask := asynq.NewTask(TaskParseDocument, parsePayload, asynq.MaxRetry(3))

	if _, err := p.AsynqClient.Enqueue(parseTask); err != nil {
		log.Printf("Failed to enqueue parse task for IPO %d: %v", payload.IPOID, err)
	}

	return nil
}

// findDRHPLink fetches the Chittorgarh detail page and searches for the DRHP PDF link
func findDRHPLink(url string) (string, error) {
	// For simplicity, we can use standard goquery here, since the links are usually in static HTML
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %v", url, err)
	}

	var pdfUrl string
	var fallbackUrl string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		href, exists := s.Attr("href")
		if exists && (strings.Contains(text, "drhp") || strings.Contains(text, "rhp") || strings.Contains(text, "prospectus")) {
			if strings.Contains(href, "/report/") || strings.Contains(href, "/keyword/") || strings.Contains(href, "/book-chapter/") {
				return // skip sidebar and glossary links
			}
			log.Printf("Found candidate link: text='%s', href='%s'", text, href)
			if strings.HasPrefix(href, "/") {
				href = "https://www.chittorgarh.com" + href
			}
			
			lowerHref := strings.ToLower(href)
			if strings.Contains(lowerHref, ".pdf") || strings.Contains(lowerHref, ".zip") {
				pdfUrl = href
			} else if fallbackUrl == "" {
				fallbackUrl = href
			}
		}
	})

	if pdfUrl == "" {
		pdfUrl = fallbackUrl
	}

	if pdfUrl == "" {
		return "", fmt.Errorf("no DRHP link found on the page")
	}

	if strings.Contains(pdfUrl, "sebi.gov.in") && strings.HasSuffix(pdfUrl, ".html") {
		log.Printf("Following SEBI HTML link: %s", pdfUrl)
		sebiPdf, err := extractSebiPDF(pdfUrl)
		if err != nil {
			return "", fmt.Errorf("failed to extract PDF from SEBI page: %v", err)
		}
		pdfUrl = sebiPdf
	}

	return pdfUrl, nil
}

// extractSebiPDF fetches a SEBI HTML page and extracts the embedded PDF link from the iframe src
func extractSebiPDF(url string) (string, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return "", err
	}

	var pdfUrl string
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && strings.Contains(src, "?file=") {
			parts := strings.Split(src, "?file=")
			if len(parts) > 1 {
				pdfUrl = parts[1]
			}
		}
	})

	if pdfUrl == "" {
		return "", fmt.Errorf("no iframe with ?file= found on SEBI page")
	}

	return pdfUrl, nil
}
