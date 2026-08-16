package scraper

import (
	"testing"
)

func TestParseTimesSubscribed(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1.52x", 1.52},
		{"1.52", 1.52},
		{"0.00", 0.0},
		{"-", 0.0},
		{"", 0.0},
		{"  15.5 Times  ", 15.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseTimesSubscribed(tt.input)
			if result != tt.expected {
				t.Errorf("parseTimesSubscribed(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
