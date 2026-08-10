package services

import (
	"fmt"
	"math"
	"strconv"
	"sort"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
)

type Metrics struct {
	RevenueCAGR3Y    float64 `json:"revenue_cagr_3y"`
	PATCAGR3Y        float64 `json:"pat_cagr_3y"`
	EBITDAMargin     float64 `json:"ebitda_margin"`
	PATMargin        float64 `json:"pat_margin"`
	ROE              float64 `json:"roe"`
	ROCE             float64 `json:"roce"`
	DebtToEquity     float64 `json:"debt_to_equity"`
	AssetTurnover    float64 `json:"asset_turnover"`
}

// CalculateMetrics computes financial ratios from a slice of Financial models.
// It expects the financials to be for consecutive years.
func CalculateMetrics(financials []models.Financial) (*Metrics, error) {
	if len(financials) == 0 {
		return nil, fmt.Errorf("no financial data available")
	}

	// Sort financials ascending by year (oldest to newest)
	sorted := make([]models.Financial, len(financials))
	copy(sorted, financials)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Year < sorted[j].Year
	})

	latest := sorted[len(sorted)-1]
	
	// Helper to safely parse strings to float64
	parseFloat := func(s string, valid bool) float64 {
		if !valid || s == "" {
			return 0.0
		}
		val, _ := strconv.ParseFloat(s, 64)
		return val
	}

	metrics := &Metrics{}

	// Latest Year Metrics
	rev := parseFloat(latest.Revenue.String, latest.Revenue.Valid)
	pat := parseFloat(latest.Pat.String, latest.Pat.Valid)
	ebitda := parseFloat(latest.Ebitda.String, latest.Ebitda.Valid)
	assets := parseFloat(latest.TotalAssets.String, latest.TotalAssets.Valid)
	debt := parseFloat(latest.TotalDebt.String, latest.TotalDebt.Valid)
	equity := parseFloat(latest.Equity.String, latest.Equity.Valid)

	// Margins
	if rev > 0 {
		metrics.EBITDAMargin = ebitda / rev
		metrics.PATMargin = pat / rev
	}

	if assets > 0 {
		metrics.AssetTurnover = rev / assets
	}

	// Return Ratios
	if equity > 0 {
		metrics.ROE = pat / equity
		metrics.DebtToEquity = debt / equity
	}
	
	capEmployed := equity + debt
	if capEmployed > 0 {
		metrics.ROCE = ebitda / capEmployed
	}

	// 3Y CAGR (Needs at least 3 years of data)
	if len(sorted) >= 3 {
		oldest3Y := sorted[len(sorted)-3]
		oldRev := parseFloat(oldest3Y.Revenue.String, oldest3Y.Revenue.Valid)
		oldPat := parseFloat(oldest3Y.Pat.String, oldest3Y.Pat.Valid)

		if oldRev > 0 && rev > 0 {
			metrics.RevenueCAGR3Y = math.Pow(rev/oldRev, 1.0/2.0) - 1
		}
		
		// PAT can be negative, CAGR is technically invalid if crossing 0, but we'll do simple check
		if oldPat > 0 && pat > 0 {
			metrics.PATCAGR3Y = math.Pow(pat/oldPat, 1.0/2.0) - 1
		}
	}

	return metrics, nil
}
