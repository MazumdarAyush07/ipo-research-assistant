package scoring

import (
	"context"
	"database/sql"
	"testing"
	
	"github.com/MazumdarAyush07/ipo-research/internal/models"
)

func TestScoreFinancials(t *testing.T) {
	tests := []struct {
		name       string
		financials []models.Financial
		wantScore  int
	}{
		{
			name:       "no history",
			financials: []models.Financial{},
			wantScore:  0,
		},
		{
			name: "strong growth and margins",
			financials: []models.Financial{
				{Year: 2025, Revenue: sql.NullString{String: "100", Valid: true}, Pat: sql.NullString{String: "10", Valid: true}},
				{Year: 2026, Revenue: sql.NullString{String: "130", Valid: true}, Pat: sql.NullString{String: "15", Valid: true}},
			},
			wantScore: 35, // 15 (Rev >20%) + 15 (Pat >20%) + 5 (Margin 11.5% >5%) = 35
		},
		{
			name: "declining financials",
			financials: []models.Financial{
				{Year: 2025, Revenue: sql.NullString{String: "100", Valid: true}, Pat: sql.NullString{String: "10", Valid: true}},
				{Year: 2026, Revenue: sql.NullString{String: "90", Valid: true}, Pat: sql.NullString{String: "8", Valid: true}},
			},
			wantScore: 5, // 0 (Rev <0) + 0 (Pat <0) + 5 (Margin 8.8% >5%) = 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, _ := scoreFinancials(tt.financials)
			if score != tt.wantScore {
				t.Errorf("scoreFinancials() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestScoreValuation(t *testing.T) {
	tests := []struct {
		name      string
		val       models.Valuation
		wantScore int
	}{
		{
			name: "missing PE",
			val:  models.Valuation{PeRatio: sql.NullString{Valid: false}},
			wantScore: 0,
		},
		{
			name: "attractive PE",
			val:  models.Valuation{PeRatio: sql.NullString{String: "15.5", Valid: true}},
			wantScore: 18,
		},
		{
			name: "moderate PE",
			val:  models.Valuation{PeRatio: sql.NullString{String: "35.0", Valid: true}},
			wantScore: 10,
		},
		{
			name: "expensive PE",
			val:  models.Valuation{PeRatio: sql.NullString{String: "60.0", Valid: true}},
			wantScore: 5,
		},
		{
			name: "negative PE",
			val:  models.Valuation{PeRatio: sql.NullString{String: "-5.0", Valid: true}},
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, _ := scoreValuation(context.Background(), tt.val, nil, "")
			if score != tt.wantScore {
				t.Errorf("scoreValuation() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestScoreSubscriptions(t *testing.T) {
	tests := []struct {
		name      string
		subs      []models.SubscriptionDatum
		wantScore int
	}{
		{
			name: "no subs",
			subs: []models.SubscriptionDatum{},
			wantScore: 0,
		},
		{
			name: "massive QIB",
			subs: []models.SubscriptionDatum{
				{Category: "QIB", TimesSubscribed: sql.NullString{String: "60.5", Valid: true}},
			},
			wantScore: 5,
		},
		{
			name: "strong QIB",
			subs: []models.SubscriptionDatum{
				{Category: "QIB", TimesSubscribed: sql.NullString{String: "15.0", Valid: true}},
			},
			wantScore: 4,
		},
		{
			name: "weak QIB",
			subs: []models.SubscriptionDatum{
				{Category: "QIB", TimesSubscribed: sql.NullString{String: "0.5", Valid: true}},
			},
			wantScore: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, _ := scoreSubscriptions(models.Ipo{}, tt.subs)
			if score != tt.wantScore {
				t.Errorf("scoreSubscriptions() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestScoreGMP(t *testing.T) {
	tests := []struct {
		name      string
		gmp       models.GmpHistory
		wantScore int
	}{
		{
			name: "missing GMP",
			gmp:  models.GmpHistory{PremiumPercent: sql.NullString{Valid: false}},
			wantScore: 0,
		},
		{
			name: "high GMP",
			gmp:  models.GmpHistory{PremiumPercent: sql.NullString{String: "65.0", Valid: true}},
			wantScore: 5,
		},
		{
			name: "moderate GMP",
			gmp:  models.GmpHistory{PremiumPercent: sql.NullString{String: "25.0", Valid: true}},
			wantScore: 3,
		},
		{
			name: "low GMP",
			gmp:  models.GmpHistory{PremiumPercent: sql.NullString{String: "10.0", Valid: true}},
			wantScore: 1,
		},
		{
			name: "negative GMP",
			gmp:  models.GmpHistory{PremiumPercent: sql.NullString{String: "-5.0", Valid: true}},
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, _ := scoreGMP(tt.gmp)
			if score != tt.wantScore {
				t.Errorf("scoreGMP() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}
