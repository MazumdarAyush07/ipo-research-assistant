package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFetchUpcomingIPOs(t *testing.T) {
	// Mock Chittorgarh HTML Response (matching ipo_dashboard.asp structure)
	mockHTML := `
	<html>
	<body>
		<table class="table">
			<tr>
				<th>Company</th>
				<th>Issue Date</th>
			</tr>
			<tr>
				<td>TechNova Solutions Ltd</td>
				<td>10 - 12 Aug</td>
			</tr>
			<tr>
				<td>GreenFuture Energy SME</td>
				<td>01 - 03 Jul</td>
			</tr>
		</table>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(mockHTML))
	if err != nil {
		t.Fatalf("Failed to parse mock HTML: %v", err)
	}

	// Execute Scraper parsing logic directly
	ipos := parseHTMLTable(doc)

	if len(ipos) != 2 {
		t.Fatalf("Expected 2 IPOs, got %d", len(ipos))
	}

	// Verify Mainboard Parsing
	if ipos[0].Name != "TechNova Solutions Ltd" {
		t.Errorf("Expected name 'TechNova Solutions Ltd', got %s", ipos[0].Name)
	}
	if ipos[0].ExchangeType != "MAINBOARD" {
		t.Errorf("Expected ExchangeType 'MAINBOARD', got %s", ipos[0].ExchangeType)
	}

	// Verify SME Parsing
	if ipos[1].Name != "GreenFuture Energy SME" {
		t.Errorf("Expected name 'GreenFuture Energy SME', got %s", ipos[1].Name)
	}
	if ipos[1].ExchangeType != "SME" {
		t.Errorf("Expected ExchangeType 'SME', got %s", ipos[1].ExchangeType)
	}
}
