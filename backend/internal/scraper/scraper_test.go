package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFetchUpcomingIPOs(t *testing.T) {
	// Mock Chittorgarh HTML Response (matching the new report format)
	mockHTML := `
	<html>
	<body>
		<table class="table">
			<tr>
				<th>Company</th>
				<th>Issue Category</th>
				<th>Pricing Method</th>
				<th>Opening Date</th>
				<th>Closing Date</th>
			</tr>
			<tr>
				<td>TechNova Solutions Ltd</td>
				<td>Mainboard</td>
				<td>Bookbuilding</td>
				<td>10-Aug-2026</td>
				<td>12-Aug-2026</td>
			</tr>
			<tr>
				<td>GreenFuture Energy</td>
				<td>SME</td>
				<td>Bookbuilding</td>
				<td>01-Jul-2026</td>
				<td>03-Jul-2026</td>
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
	if ipos[0].OpenDate != "10-Aug-2026" {
		t.Errorf("Expected OpenDate '10-Aug-2026', got %s", ipos[0].OpenDate)
	}

	// Verify SME Parsing
	if ipos[1].Name != "GreenFuture Energy" {
		t.Errorf("Expected name 'GreenFuture Energy', got %s", ipos[1].Name)
	}
	if ipos[1].ExchangeType != "SME" {
		t.Errorf("Expected ExchangeType 'SME', got %s", ipos[1].ExchangeType)
	}
}
