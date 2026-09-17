package scoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/MazumdarAyush07/ipo-research/internal/ai"
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/services"
)

type ScoreResult struct {
	TotalScore     int
	Recommendation string
	
	FinancialsScore  int
	FinancialsReason string
	
	ValuationScore  int
	ValuationReason string
	
	PromoterScore  int
	PromoterReason string
	
	IndustryScore  int
	IndustryReason string
	
	RiskScore  int
	RiskReason string
	
	SubscriptionScore  int
	SubscriptionReason string
	
	GmpScore  int
	GmpReason string
}

func ScoreIPO(ctx context.Context, queries *models.Queries, peerService *services.PeerService, ipoID int64) (*ScoreResult, error) {
	result := &ScoreResult{}

	// 1. Fetch all data
	ipo, err := queries.GetIPO(ctx, ipoID)
	sector := ""
	if err == nil && ipo.Sector.Valid {
		sector = ipo.Sector.String
	}

	financials, err := queries.GetFinancialsByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch financials: %v", err)
	}

	subs, err := queries.GetLatestSubscription(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch subscriptions: %v", err)
	}

	gmp, err := queries.GetLatestGMP(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch gmp: %v", err)
	}

	aiData, err := queries.GetAIAnalysisByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch AI analysis: %v", err)
	}

	val, err := queries.GetValuationByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch valuation: %v", err)
	}

	// 2. Score Financials (0-40)
	result.FinancialsScore, result.FinancialsReason = scoreFinancials(financials)

	// 3. Score Valuation (0-20)
	result.ValuationScore, result.ValuationReason = scoreValuation(ctx, val, peerService, sector)

	// 4. Score Subscriptions (0-5)
	result.SubscriptionScore, result.SubscriptionReason = scoreSubscriptions(ipo, subs)

	// 5. Score GMP (0-5)
	result.GmpScore, result.GmpReason = scoreGMP(gmp)

	// 6. Score AI Modules (Promoter 0-10, Industry 0-10, Risk 0-10)
	if aiData.ID != 0 && aiData.RawJson.Valid {
		promoter, ind, risk := scoreAIModules(ctx, string(aiData.RawJson.RawMessage))
		result.PromoterScore = promoter.Score
		result.PromoterReason = promoter.Reason
		result.IndustryScore = ind.Score
		result.IndustryReason = ind.Reason
		result.RiskScore = risk.Score
		result.RiskReason = risk.Reason
	} else {
		result.PromoterReason = "No AI analysis available."
		result.IndustryReason = "No AI analysis available."
		result.RiskReason = "No AI analysis available."
	}

	rawTotal := result.FinancialsScore + result.ValuationScore + result.PromoterScore +
		result.IndustryScore + result.RiskScore + result.SubscriptionScore + result.GmpScore

	maxPossible := 100
	if result.SubscriptionReason == "No subscription data available yet." {
		maxPossible -= 5
	}
	if result.GmpReason == "No GMP data available." {
		maxPossible -= 5
	}

	if maxPossible < 100 && maxPossible > 0 {
		result.TotalScore = int((float64(rawTotal) / float64(maxPossible)) * 100)
	} else {
		result.TotalScore = rawTotal
	}

	if result.TotalScore >= 75 {
		result.Recommendation = "Apply"
	} else if result.TotalScore >= 50 {
		result.Recommendation = "Apply with Caution"
	} else {
		result.Recommendation = "Avoid"
	}

	// 8. Save to DB
	err = saveScore(ctx, queries, ipoID, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save score: %w", err)
	}

	return result, nil
}

func saveScore(ctx context.Context, queries *models.Queries, ipoID int64, res *ScoreResult) error {
	params := models.CreateOrUpdateScoreParams{
		IpoID:             ipoID,
		FinancialsScore:   sql.NullString{String: strconv.Itoa(res.FinancialsScore), Valid: true},
		FinancialsReason:  sql.NullString{String: res.FinancialsReason, Valid: true},
		ValuationScore:    sql.NullString{String: strconv.Itoa(res.ValuationScore), Valid: true},
		ValuationReason:   sql.NullString{String: res.ValuationReason, Valid: true},
		PromoterScore:     sql.NullString{String: strconv.Itoa(res.PromoterScore), Valid: true},
		PromoterReason:    sql.NullString{String: res.PromoterReason, Valid: true},
		IndustryScore:     sql.NullString{String: strconv.Itoa(res.IndustryScore), Valid: true},
		IndustryReason:    sql.NullString{String: res.IndustryReason, Valid: true},
		RiskScore:         sql.NullString{String: strconv.Itoa(res.RiskScore), Valid: true},
		RiskReason:        sql.NullString{String: res.RiskReason, Valid: true},
		SubscriptionScore: sql.NullString{String: strconv.Itoa(res.SubscriptionScore), Valid: true},
		GmpScore:          sql.NullString{String: strconv.Itoa(res.GmpScore), Valid: true},
		FinalScore:        sql.NullString{String: strconv.Itoa(res.TotalScore), Valid: true},
		Recommendation:    sql.NullString{String: res.Recommendation, Valid: true},
	}
	_, err := queries.CreateOrUpdateScore(ctx, params)
	return err
}

func scoreFinancials(financials []models.Financial) (int, string) {
	if len(financials) < 2 {
		return 0, "Insufficient financial history available for scoring."
	}

	// Sort by year ascending
	latest := financials[len(financials)-1]
	previous := financials[len(financials)-2]
	if latest.Year < previous.Year {
		latest, previous = previous, latest // basic swap if unsorted for 2 items
	}

	parseMoney := func(s sql.NullString) float64 {
		if !s.Valid {
			return 0
		}
		val, _ := strconv.ParseFloat(s.String, 64)
		return val
	}

	latestRev := parseMoney(latest.Revenue)
	prevRev := parseMoney(previous.Revenue)
	latestPat := parseMoney(latest.Pat)
	prevPat := parseMoney(previous.Pat)

	score := 0
	var reasons []string

	// Revenue Growth (up to 15 points)
	if prevRev > 0 {
		revGrowth := (latestRev - prevRev) / prevRev * 100
		if revGrowth > 20 {
			score += 15
			reasons = append(reasons, "Strong revenue growth (>20%).")
		} else if revGrowth > 5 {
			score += 10
			reasons = append(reasons, "Moderate revenue growth.")
		} else if revGrowth >= 0 {
			score += 5
			reasons = append(reasons, "Flat revenue.")
		} else {
			reasons = append(reasons, "Declining revenue.")
		}
	} else if latestRev > 0 {
		score += 10
		reasons = append(reasons, "Revenue generated but previous year missing.")
	}

	// PAT Growth (up to 15 points)
	if prevPat > 0 {
		patGrowth := (latestPat - prevPat) / prevPat * 100
		if patGrowth > 20 {
			score += 15
			reasons = append(reasons, "Strong profit growth (>20%).")
		} else if patGrowth > 5 {
			score += 10
			reasons = append(reasons, "Moderate profit growth.")
		} else if patGrowth >= 0 {
			score += 5
			reasons = append(reasons, "Flat profits.")
		} else {
			reasons = append(reasons, "Declining profits.")
		}
	} else if latestPat > 0 {
		score += 10
		reasons = append(reasons, "Turned profitable this year.")
	} else {
		reasons = append(reasons, "Loss-making.")
	}

	// Profit Margin (up to 10 points)
	if latestRev > 0 {
		margin := (latestPat / latestRev) * 100
		if margin > 15 {
			score += 10
			reasons = append(reasons, "Excellent profit margins (>15%).")
		} else if margin > 5 {
			score += 5
			reasons = append(reasons, "Average profit margins.")
		} else if margin > 0 {
			score += 2
			reasons = append(reasons, "Thin profit margins.")
		} else {
			reasons = append(reasons, "Negative profit margins.")
		}
	}

	return score, strings.Join(reasons, " ")
}

func scoreValuation(ctx context.Context, val models.Valuation, ps *services.PeerService, sector string) (int, string) {
	if !val.PeRatio.Valid {
		return 0, "No valuation data available."
	}
	pe, _ := strconv.ParseFloat(val.PeRatio.String, 64)
	if pe <= 0 {
		return 0, "Negative or zero P/E ratio indicates losses."
	}

	if ps != nil && sector != "" {
		peers, err := ps.FetchPeersForSector(ctx, sector)
		if err == nil && len(peers) > 0 {
			comp := services.CompareIPOToPeers(pe, peers)
			if comp.PremiumDiscount <= -20 {
				return 20, fmt.Sprintf("Highly attractive valuation: %.2f%% discount to peers (PE: %.2f vs Median: %.2f).", -comp.PremiumDiscount, pe, comp.MedianPE)
			} else if comp.PremiumDiscount <= 0 {
				return 15, fmt.Sprintf("Attractive valuation: %.2f%% discount to peers (PE: %.2f vs Median: %.2f).", -comp.PremiumDiscount, pe, comp.MedianPE)
			} else if comp.PremiumDiscount <= 20 {
				return 10, fmt.Sprintf("Fair valuation: %.2f%% premium to peers (PE: %.2f vs Median: %.2f).", comp.PremiumDiscount, pe, comp.MedianPE)
			} else {
				return 5, fmt.Sprintf("Expensive valuation: %.2f%% premium to peers (PE: %.2f vs Median: %.2f).", comp.PremiumDiscount, pe, comp.MedianPE)
			}
		}
	}

	if pe < 20 {
		return 18, fmt.Sprintf("Attractive P/E ratio of %.2f.", pe)
	} else if pe < 40 {
		return 10, fmt.Sprintf("Moderate P/E ratio of %.2f.", pe)
	}
	return 5, fmt.Sprintf("Expensive P/E ratio of %.2f.", pe)
}

func scoreSubscriptions(ipo models.Ipo, subs []models.SubscriptionDatum) (int, string) {
	if len(subs) == 0 {
		return 0, "No subscription data available yet."
	}
	
	isSME := ipo.ExchangeType.Valid && ipo.ExchangeType.String == "SME"
	
	var qibSub, niiSub, retailSub float64
	for _, s := range subs {
		cat := strings.ToUpper(s.Category)
		if !s.TimesSubscribed.Valid {
			continue
		}
		val, _ := strconv.ParseFloat(s.TimesSubscribed.String, 64)
		if strings.Contains(cat, "QIB") || strings.Contains(cat, "QUALIFIED") {
			qibSub += val
		} else if strings.Contains(cat, "NII") || strings.Contains(cat, "NON-INSTITUTIONAL") {
			niiSub += val
		} else if strings.Contains(cat, "RETAIL") {
			retailSub += val
		}
	}
	
	// If it's an SME IPO, or if QIB is 0.00 (they haven't bid yet), evaluate based on Retail & NII demand
	if isSME || qibSub == 0 {
		totalRetailNII := niiSub + retailSub
		if totalRetailNII > 200 {
			return 5, "Exceptional Retail/HNI demand (>200x combined)."
		} else if totalRetailNII > 50 {
			return 4, "Strong Retail/HNI demand (>50x combined)."
		} else if totalRetailNII > 10 {
			return 3, "Moderate Retail/HNI demand."
		} else if totalRetailNII > 1 {
			return 2, "Mild Retail/HNI demand."
		}
		if isSME {
			return 1, "Poor Retail/HNI demand."
		}
		return 1, "Poor institutional and retail demand."
	}
	
	if qibSub > 50 {
		return 5, "Exceptional QIB demand (>50x)."
	} else if qibSub > 10 {
		return 4, "Strong QIB demand (>10x)."
	} else if qibSub > 1 {
		return 2, "Moderate QIB demand."
	}
	return 1, "Poor institutional demand."
}

func scoreGMP(gmp models.GmpHistory) (int, string) {
	if !gmp.PremiumPercent.Valid {
		return 0, "No GMP data available."
	}
	premium, _ := strconv.ParseFloat(gmp.PremiumPercent.String, 64)
	if premium > 50 {
		return 5, fmt.Sprintf("High GMP premium of %.2f%%.", premium)
	} else if premium > 20 {
		return 3, fmt.Sprintf("Moderate GMP premium of %.2f%%.", premium)
	} else if premium > 0 {
		return 1, fmt.Sprintf("Low GMP premium of %.2f%%.", premium)
	}
	return 0, "Negative or zero GMP premium."
}

type moduleResult struct {
	Score  int
	Reason string
}

func scoreAIModules(ctx context.Context, aiJson string) (moduleResult, moduleResult, moduleResult) {
	client, err := ai.NewGeminiClient(ctx)
	if err != nil {
		return moduleResult{0, "Failed to init AI client"}, moduleResult{0, "Failed to init AI client"}, moduleResult{0, "Failed to init AI client"}
	}
	defer client.Close()

	promptBytes, err := os.ReadFile("/prompts/scorer_v1.md")
	if err != nil {
		// Fallback if running outside docker or prompt missing
		promptBytes = []byte("Score Promoter, Industry, and Risk out of 10 based on JSON.")
	}

	res, err := client.ScoreModules(ctx, aiJson, string(promptBytes))
	if err != nil {
		return moduleResult{0, "AI scoring failed"}, moduleResult{0, "AI scoring failed"}, moduleResult{0, "AI scoring failed"}
	}

	return moduleResult{res.PromoterScore, res.PromoterReason},
		moduleResult{res.IndustryScore, res.IndustryReason},
		moduleResult{res.RiskScore, res.RiskReason}
}
