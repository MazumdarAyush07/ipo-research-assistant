package scraper

import (
	"fmt"
	"testing"
	"time"
)

func TestParsePeriod(t *testing.T) {
	year := 2026
	open, close := parsePeriod("24 Aug - 27 Aug", year)
	if open != "24-Aug-2026" {
		t.Errorf("Expected '24-Aug-2026', got '%s'", open)
	}
	if close != "27-Aug-2026" {
		t.Errorf("Expected '27-Aug-2026', got '%s'", close)
	}
}

func TestParsePeriodInvalid(t *testing.T) {
	open, close := parsePeriod("invalid", 2026)
	if open != "" || close != "" {
		t.Errorf("Expected empty strings for invalid period, got '%s' and '%s'", open, close)
	}
}

func TestFetchUpcomingIPOs(t *testing.T) {
	// Test JSON struct mapping with real field names from the Chittorgarh API
	items := []chittorgarhIPO{
		{
			IpoNewsTitle:         "TechNova Solutions IPO",
			IssueCategory:        "Mainline",
			IpoPeriod:            "10 Aug - 12 Aug",
			UrlrewriteFolderName: "technova-solutions-ipo",
		},
		{
			IpoNewsTitle:         "GreenFuture Energy IPO",
			IssueCategory:        "SME",
			IpoPeriod:            "01 Jul - 03 Jul",
			UrlrewriteFolderName: "greenfuture-energy-ipo",
		},
	}

	year := time.Now().Year()

	// Test first item (Mainline -> MAINBOARD)
	open0, close0 := parsePeriod(items[0].IpoPeriod, year)
	if open0 != fmt.Sprintf("10-Aug-%d", year) {
		t.Errorf("Expected '10-Aug-%d', got '%s'", year, open0)
	}
	if close0 != fmt.Sprintf("12-Aug-%d", year) {
		t.Errorf("Expected '12-Aug-%d', got '%s'", year, close0)
	}

	expectedURL0 := chittorgarhBaseURL + "technova-solutions-ipo" + "/"
	actualURL0 := chittorgarhBaseURL + items[0].UrlrewriteFolderName + "/"
	if actualURL0 != expectedURL0 {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL0, actualURL0)
	}

	// Test second item (SME)
	if items[1].IssueCategory != "SME" {
		t.Errorf("Expected 'SME', got '%s'", items[1].IssueCategory)
	}

	if items[0].IpoNewsTitle != "TechNova Solutions IPO" {
		t.Errorf("Expected 'TechNova Solutions IPO', got '%s'", items[0].IpoNewsTitle)
	}
}
