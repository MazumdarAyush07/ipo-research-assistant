package worker

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

	// ALWAYS create the directory immediately so the audit script knows the IPO exists
	// even if the download ultimately fails.
	storageDir := filepath.Join("../storage", slug)
	os.MkdirAll(storageDir, os.ModePerm)

	log.Printf("Scraping detail page for %s: %s", ipo.Name, ipo.SourceUrl.String)
	drhpUrl, err := findDRHPLink(ipo.SourceUrl.String)
	if err != nil {
		log.Printf("Could not find DRHP link for %s: %v. Aborting job (no retries).", ipo.Name, err)
		return nil // Return nil so Asynq does not retry this job
	}

	destPath := filepath.Join(storageDir, "drhp.pdf")
	
	// PRE-CLEANUP: If a corrupted file exists, delete it so we force a fresh download
	if info, err := os.Stat(destPath); err == nil && info.Size() > 0 {
		if !isValidPDF(destPath) && !isZipFile(destPath) {
			log.Printf("Existing file %s is corrupted (neither PDF nor ZIP). Deleting it...", destPath)
			os.Remove(destPath)
		}
	}

	log.Printf("Downloading DRHP to %s", destPath)

	// Download PDF or ZIP (require at least 500KB)
	err = downloader.DownloadPDF(drhpUrl, destPath, 500*1024)
	if err != nil {
		log.Printf("Failed to download DRHP for %s: %v. Aborting job (no retries).", ipo.Name, err)
		return nil // Do not retry on download size failures
	}

	// POST-VALIDATION: Check if what we downloaded is actually valid
	if !isValidPDF(destPath) && !isZipFile(destPath) {
		os.Remove(destPath)
		log.Printf("Downloaded file is neither a valid PDF nor a ZIP archive. Aborting job.")
		return nil
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
			log.Printf("Failed to extract PDF from zip: %v. Aborting job.", err)
			return nil
		}
		
		// Cleanup the temp zip
		os.Remove(tempZipPath)
	}

	// PHASE 6 INTEGRITY CHECK: Call python sidecar /validate to ensure it has >50 pages and no Unexpected EOF
	if err := validatePDFWithPython(destPath); err != nil {
		os.Rename(destPath, filepath.Join(storageDir, "bad.pdf")) // save the bad file for inspection
		log.Printf("PDF integrity validation failed: %v. Saved as bad.pdf. Aborting job.", err)
		return nil
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

	log.Printf("Successfully downloaded DRHP for IPO ID: %d. Standing by for manual parse trigger.", payload.IPOID)

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
		
		isValidText := strings.Contains(text, "drhp") || strings.Contains(text, "rhp") || 
					   strings.Contains(text, "prospectus") || strings.Contains(text, "offer document")

		if exists && isValidText {
			if strings.Contains(href, "/report/") || strings.Contains(href, "/keyword/") || strings.Contains(href, "/book-chapter/") {
				return // skip sidebar and glossary links
			}
			badKeywords := []string{"addendum", "corrigendum", "placement", "abridged", "notice", "checklist"}
			for _, bad := range badKeywords {
				if strings.Contains(text, bad) || strings.Contains(strings.ToLower(href), bad) {
					return // skip bad links
				}
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
		// DEBUG: Print all a tags to see what we missed
		log.Printf("DEBUG: No valid link found. Listing all links on page:")
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			href, _ := s.Attr("href")
			if href != "" && (strings.Contains(href, ".pdf") || strings.Contains(href, ".zip")) {
				log.Printf("DEBUG LINK - Text: '%s', Href: '%s'", text, href)
			}
		})
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

// extractLargestPDFFromZip opens a zip archive and extracts the most relevant PDF to destPdfPath.
// If multiple PDFs exist (e.g. RHP and DRHP), it prefers RHP -> DRHP -> Largest PDF.
func extractLargestPDFFromZip(zipPath string, destPdfPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var targetFile *zip.File
	var rhpFile *zip.File
	var drhpFile *zip.File
	
	var largestFile *zip.File
	var maxSize uint64 = 0

	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".pdf") {
			lowerName := strings.ToLower(f.Name)
			if strings.Contains(lowerName, "gid") || strings.Contains(lowerName, "form") || 
			   strings.Contains(lowerName, "checklist") || strings.Contains(lowerName, "certificate") || 
			   strings.Contains(lowerName, "notice") {
				continue // skip bad files in zip
			}
			
			if strings.Contains(lowerName, "drhp") {
				drhpFile = f
			} else if strings.Contains(lowerName, "rhp") {
				rhpFile = f
			}

			if f.UncompressedSize64 > maxSize {
				maxSize = f.UncompressedSize64
				largestFile = f
			}
		}
	}

	// Priority: RHP > DRHP > Largest PDF
	if rhpFile != nil {
		targetFile = rhpFile
	} else if drhpFile != nil {
		targetFile = drhpFile
	} else if largestFile != nil {
		targetFile = largestFile
	} else {
		return fmt.Errorf("no valid PDF found inside zip")
	}

	rc, err := targetFile.Open()
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

type ValidateResponse struct {
	Valid     bool   `json:"valid"`
	PageCount int    `json:"page_count"`
	Error     string `json:"error"`
}

func validatePDFWithPython(filePath string) error {
	parserUrl := os.Getenv("PDF_PARSER_URL")
	if parserUrl == "" {
		parserUrl = "http://pdf-parser:8000"
	}

	payload := map[string]string{"file_path": filePath}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(parserUrl+"/validate", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to call python validation endpoint: %v", err)
	}
	defer resp.Body.Close()

	var valResp ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&valResp); err != nil {
		return fmt.Errorf("failed to decode validation response: %v", err)
	}

	if !valResp.Valid {
		return fmt.Errorf("%s (page count: %d)", valResp.Error, valResp.PageCount)
	}
	return nil
}
