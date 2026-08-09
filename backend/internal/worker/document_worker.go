package worker

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
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
	
	// PRE-CLEANUP: If a corrupted file exists, delete it so we force a fresh download
	if info, err := os.Stat(destPath); err == nil && info.Size() > 0 {
		if !isValidPDF(destPath) && !isZipFile(destPath) {
			log.Printf("Existing file %s is corrupted (neither PDF nor ZIP). Deleting it...", destPath)
			os.Remove(destPath)
		}
	}

	log.Printf("Downloading DRHP to %s", destPath)

	// Download PDF or ZIP (require at least 100KB)
	err = downloader.DownloadPDF(drhpUrl, destPath, 100*1024)
	if err != nil {
		return fmt.Errorf("failed to download DRHP for %s: %v", ipo.Name, err)
	}

	// POST-VALIDATION: Check if what we downloaded is actually valid
	if !isValidPDF(destPath) && !isZipFile(destPath) {
		os.Remove(destPath)
		return fmt.Errorf("downloaded file is neither a valid PDF nor a ZIP archive")
	}

	// Bulletproof check: Does the file start with ZIP magic bytes?
	if isZipFile(destPath) {
		log.Printf("File is actually a ZIP archive. Extracting PDF...")
		tempZipPath := filepath.Join("../storage", slug, "temp_drhp.zip")
		
		// Rename the downloaded file to a temp zip
		if err := os.Rename(destPath, tempZipPath); err != nil {
			return fmt.Errorf("failed to rename zip file: %v", err)
		}
		
		// Extract the largest PDF from the zip into destPath (drhp.pdf)
		err = extractLargestPDFFromZip(tempZipPath, destPath)
		if err != nil {
			return fmt.Errorf("failed to extract PDF from zip: %v", err)
		}
		
		// Cleanup the temp zip
		os.Remove(tempZipPath)
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

// extractLargestPDFFromZip opens a zip archive, finds the largest .pdf file, and extracts it to destPdfPath
func extractLargestPDFFromZip(zipPath string, destPdfPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var largestFile *zip.File
	var maxSize uint64 = 0

	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".pdf") {
			if f.UncompressedSize64 > maxSize {
				maxSize = f.UncompressedSize64
				largestFile = f
			}
		}
	}

	if largestFile == nil {
		return fmt.Errorf("no PDF found inside zip")
	}

	rc, err := largestFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(destPdfPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// isZipFile checks if a file starts with the ZIP magic bytes "PK\x03\x04"
func isZipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return false
	}
	return string(buf) == "PK\x03\x04"
}

// isValidPDF checks if a file contains the PDF magic bytes "%PDF-" within the first 512 bytes
func isValidPDF(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	return strings.Contains(string(buf[:n]), "%PDF-")
}
