package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadPDF downloads a file from the given URL and saves it to destPath.
// It ensures that the downloaded file is at least minSizeBytes, returning an error otherwise.
func DownloadPDF(url string, destPath string, minSizeBytes int64) error {
	// Check if file already exists
	if info, err := os.Stat(destPath); err == nil && info.Size() > minSizeBytes {
		return nil // Already downloaded
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %v", url, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code fetching %s: %d", url, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "pdf") && !strings.Contains(strings.ToLower(contentType), "octet-stream") && !strings.Contains(strings.ToLower(contentType), "zip") {
		return fmt.Errorf("URL %s is not a valid document, got Content-Type: %s", url, contentType)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", destPath, err)
	}
	defer out.Close()

	// Stream response body to file
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", destPath, err)
	}

	// Validate file size
	if written < minSizeBytes {
		_ = os.Remove(destPath) // clean up bad file
		return fmt.Errorf("downloaded file too small: %d bytes (expected at least %d)", written, minSizeBytes)
	}

	return nil
}
