package services

import (
	"database/sql"
	"math"
	"testing"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
)

func TestCalculateMetrics(t *testing.T) {
	// Helper to create string easily
	ns := func(s string) sql.NullString {
		return sql.NullString{String: s, Valid: true}
	}

	tests := []struct {
		name       string
		financials []models.Financial
		wantErr    bool
		check      func(*testing.T, *Metrics)
	}{
		{
			name:       "empty data",
			financials: []models.Financial{},
			wantErr:    true,
		},
		{
			name: "single year",
			financials: []models.Financial{
				{
					Year:        2023,
					Revenue:     ns("100.0"),
					Pat:         ns("20.0"),
					Ebitda:      ns("30.0"),
					TotalAssets: ns("200.0"),
					TotalDebt:   ns("50.0"),
					Equity:      ns("150.0"),
				},
			},
			wantErr: false,
			check: func(t *testing.T, m *Metrics) {
				if m.EBITDAMargin != 0.3 {
					t.Errorf("expected EBITDAMargin 0.3, got %v", m.EBITDAMargin)
				}
				if m.PATMargin != 0.2 {
					t.Errorf("expected PATMargin 0.2, got %v", m.PATMargin)
				}
				if m.AssetTurnover != 0.5 {
					t.Errorf("expected AssetTurnover 0.5, got %v", m.AssetTurnover)
				}
				if m.ROE != 20.0/150.0 {
					t.Errorf("expected ROE %v, got %v", 20.0/150.0, m.ROE)
				}
				if m.DebtToEquity != 50.0/150.0 {
					t.Errorf("expected DebtToEquity %v, got %v", 50.0/150.0, m.DebtToEquity)
				}
				if m.ROCE != 30.0/200.0 {
					t.Errorf("expected ROCE %v, got %v", 30.0/200.0, m.ROCE)
				}
				if m.RevenueCAGR3Y != 0.0 {
					t.Errorf("expected RevenueCAGR3Y 0.0, got %v", m.RevenueCAGR3Y)
				}
			},
		},
		{
			name: "three years with cagr",
			financials: []models.Financial{
				{
					Year:    2021,
					Revenue: ns("100.0"),
					Pat:     ns("10.0"),
				},
				{
					Year:    2022,
					Revenue: ns("120.0"),
					Pat:     ns("12.0"),
				},
				{
					Year:    2023,
					Revenue: ns("144.0"), // 100 -> 120 -> 144 is 20% CAGR over 2 periods
					Pat:     ns("14.4"), // 10 -> 12 -> 14.4 is 20% CAGR over 2 periods
				},
			},
			wantErr: false,
			check: func(t *testing.T, m *Metrics) {
				expectedRevCagr := math.Pow(144.0/100.0, 1.0/2.0) - 1
				if math.Abs(m.RevenueCAGR3Y-expectedRevCagr) > 0.0001 {
					t.Errorf("expected RevenueCAGR3Y %v, got %v", expectedRevCagr, m.RevenueCAGR3Y)
				}
				
				expectedPatCagr := math.Pow(14.4/10.0, 1.0/2.0) - 1
				if math.Abs(m.PATCAGR3Y-expectedPatCagr) > 0.0001 {
					t.Errorf("expected PATCAGR3Y %v, got %v", expectedPatCagr, m.PATCAGR3Y)
				}
			},
		},
		{
			name: "zero or negative handlers",
			financials: []models.Financial{
				{
					Year:    2021,
					Revenue: ns("0.0"),
					Pat:     ns("-10.0"),
				},
				{
					Year:    2022,
					Revenue: ns("0.0"),
					Pat:     ns("-10.0"),
				},
				{
					Year:        2023,
					Revenue:     ns("0.0"),
					Pat:         ns("-5.0"),
					Equity:      ns("0.0"),
					TotalDebt:   ns("0.0"),
					TotalAssets: ns("0.0"),
					Ebitda:      ns("0.0"),
				},
			},
			wantErr: false,
			check: func(t *testing.T, m *Metrics) {
				if m.EBITDAMargin != 0.0 {
					t.Errorf("expected EBITDAMargin 0.0, got %v", m.EBITDAMargin)
				}
				if m.ROE != 0.0 {
					t.Errorf("expected ROE 0.0, got %v", m.ROE)
				}
				if m.ROCE != 0.0 {
					t.Errorf("expected ROCE 0.0, got %v", m.ROCE)
				}
				if m.RevenueCAGR3Y != 0.0 {
					t.Errorf("expected RevenueCAGR3Y 0.0, got %v", m.RevenueCAGR3Y)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateMetrics(tt.financials)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
