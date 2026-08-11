package services

import (
	"sort"
)

type PeerComparisonResult struct {
	MedianPE        float64 `json:"median_pe"`
	MedianPB        float64 `json:"median_pb"`
	PremiumDiscount float64 `json:"premium_discount_percent"`
	ValuationFlag   string  `json:"valuation_flag"` // Expensive, Fair, Attractive
}

func CompareIPOToPeers(ipoPE float64, peers []PeerData) PeerComparisonResult {
	if len(peers) == 0 {
		return PeerComparisonResult{}
	}

	var peValues []float64
	var pbValues []float64

	for _, p := range peers {
		if p.PE > 0 {
			peValues = append(peValues, p.PE)
		}
		if p.PB > 0 {
			pbValues = append(pbValues, p.PB)
		}
	}

	medianPE := calculateMedian(peValues)
	medianPB := calculateMedian(pbValues)

	premiumDiscount := 0.0
	flag := "Unknown"

	if ipoPE > 0 && medianPE > 0 {
		// Calculate how much more (or less) expensive the IPO is compared to peers
		premiumDiscount = ((ipoPE - medianPE) / medianPE) * 100

		if premiumDiscount > 40 {
			flag = "Expensive"
		} else if premiumDiscount > 10 {
			flag = "Fair"
		} else {
			flag = "Attractive"
		}
	}

	return PeerComparisonResult{
		MedianPE:        medianPE,
		MedianPB:        medianPB,
		PremiumDiscount: premiumDiscount,
		ValuationFlag:   flag,
	}
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sort.Float64s(values)
	n := len(values)
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2.0
	}
	return values[n/2]
}
