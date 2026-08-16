package scraper

import (
	"testing"
)

func TestParseGMPValue(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"₹ 15", 15},
		{"15", 15},
		{"-", 0},
		{"", 0},
		{"₹ 120.5", 120.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseGMPValue(tt.input)
			if result != tt.expected {
				t.Errorf("parseGMPValue(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParsePremiumPercent(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"25.50%", 25.5},
		{"25.50", 25.5},
		{"-", 0},
		{"", 0},
		{"30.2% (Estimated)", 30.2}, // The text sometimes has extra text separated by a space
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePremiumPercent(tt.input)
			if result != tt.expected {
				t.Errorf("parsePremiumPercent(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
