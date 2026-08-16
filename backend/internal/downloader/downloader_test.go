package downloader

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadPDF_Success(t *testing.T) {
	// Create a dummy server that serves "PDF" bytes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		// Write exactly 100 bytes
		data := make([]byte, 100)
		for i := range data {
			data[i] = 'A'
		}
		w.Write(data)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "test.pdf")

	// Call downloader with a minimum size of 50 bytes (less than 100)
	err := DownloadPDF(ts.URL, destPath, 50)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Verify file was written and size is correct
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("File not found: %v", err)
	}
	if info.Size() != 100 {
		t.Errorf("Expected 100 bytes, got %d", info.Size())
	}
}

func TestDownloadPDF_TooSmall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("too small"))
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "test.pdf")

	// Call downloader with a minimum size of 50 bytes
	err := DownloadPDF(ts.URL, destPath, 50)
	if err == nil {
		t.Fatalf("Expected error for file being too small, but got success")
	}

	// Verify file was deleted
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		t.Errorf("Expected file to be deleted because it was too small, but it still exists")
	}
}

func TestDownloadPDF_BadContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>bad type</html>"))
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "test.pdf")

	err := DownloadPDF(ts.URL, destPath, 1)
	if err == nil {
		t.Fatalf("Expected error for bad content type, got success")
	}
}
