package scoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
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
	if result.FinancialsReason == "Insufficient financial history available for scoring." {
		maxPossible -= 40
	}
	if result.ValuationReason == "No valuation data available." {
		maxPossible -= 20
	}
	if result.SubscriptionReason == "No subscription data available yet." {
		maxPossible -= 5
	}
	if result.GmpReason == "No GMP data available." {
		maxPossible -= 5
	}
	if strings.Contains(result.FinancialsReason, "PAT data not available") {
		maxPossible -= 25 // 15 for PAT growth + 10 for margin
	}
	if strings.HasPrefix(result.PromoterReason, "No AI analysis") || strings.HasPrefix(result.PromoterReason, "AI scoring") || strings.HasPrefix(result.PromoterReason, "Failed to init") {
		maxPossible -= 10
	}
	if strings.HasPrefix(result.IndustryReason, "No AI analysis") || strings.HasPrefix(result.IndustryReason, "AI scoring") || strings.HasPrefix(result.IndustryReason, "Failed to init") {
		maxPossible -= 10
	}
	if strings.HasPrefix(result.RiskReason, "No AI analysis") || strings.HasPrefix(result.RiskReason, "AI scoring") || strings.HasPrefix(result.RiskReason, "Failed to init") {
		maxPossible -= 10
	}

	if maxPossible > 0 && maxPossible < 100 {
		result.TotalScore = int((float64(rawTotal) / float64(maxPossible)) * 100)
	} else if maxPossible <= 0 {
		result.TotalScore = 0
	} else {
		result.TotalScore = rawTotal
	}
	// Cap at 100
	if result.TotalScore > 100 {
		result.TotalScore = 100
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
		IpoID:              ipoID,
		FinancialsScore:    sql.NullString{String: strconv.Itoa(res.FinancialsScore), Valid: true},
		FinancialsReason:   sql.NullString{String: res.FinancialsReason, Valid: true},
		ValuationScore:     sql.NullString{String: strconv.Itoa(res.ValuationScore), Valid: true},
		ValuationReason:    sql.NullString{String: res.ValuationReason, Valid: true},
		PromoterScore:      sql.NullString{String: strconv.Itoa(res.PromoterScore), Valid: true},
		PromoterReason:     sql.NullString{String: res.PromoterReason, Valid: true},
		IndustryScore:      sql.NullString{String: strconv.Itoa(res.IndustryScore), Valid: true},
		IndustryReason:     sql.NullString{String: res.IndustryReason, Valid: true},
		RiskScore:          sql.NullString{String: strconv.Itoa(res.RiskScore), Valid: true},
		RiskReason:         sql.NullString{String: res.RiskReason, Valid: true},
		SubscriptionScore:  sql.NullString{String: strconv.Itoa(res.SubscriptionScore), Valid: true},
		SubscriptionReason: sql.NullString{String: res.SubscriptionReason, Valid: true},
		GmpScore:           sql.NullString{String: strconv.Itoa(res.GmpScore), Valid: true},
		GmpReason:          sql.NullString{String: res.GmpReason, Valid: true},
		FinalScore:         sql.NullString{String: strconv.Itoa(res.TotalScore), Valid: true},
		Recommendation:     sql.NullString{String: res.Recommendation, Valid: true},
	}
	_, err := queries.CreateOrUpdateScore(ctx, params)
	return err
}

func scoreFinancials(financials []models.Financial) (int, string) {
	if len(financials) < 2 {
		return 0, "Insufficient financial history available for scoring."
	}

	// Sort by year ascending
	sort.Slice(financials, func(i, j int) bool {
		return financials[i].Year < financials[j].Year
	})
	latest := financials[len(financials)-1]
	previous := financials[len(financials)-2]

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

	// Check if PAT data is actually available vs just zero
	patAvailable := latest.Pat.Valid || previous.Pat.Valid

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
	if !patAvailable {
		reasons = append(reasons, "PAT data not available.")
	} else if prevPat > 0 {
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
	if !patAvailable {
		// Skip margin calculation if PAT is unavailable
	} else if latestRev > 0 {
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
		return moduleResult{0, "Failed to init AI client."}, moduleResult{0, "Failed to init AI client."}, moduleResult{0, "Failed to init AI client."}
	}
	defer client.Close()

	// Load the scoring prompt — try multiple paths so it works inside Docker and locally.
	promptStr := loadScorerPrompt()

	res, err := client.ScoreModules(ctx, aiJson, promptStr)
	if err != nil {
		log.Printf("AI scoring error: %v", err)
		return moduleResult{0, "AI scoring failed."}, moduleResult{0, "AI scoring failed."}, moduleResult{0, "AI scoring failed."}
	}

	// Guard: if Gemini returned empty reasons (which it sometimes does), retry once with a stricter prompt.
	if res.PromoterReason == "" || res.IndustryReason == "" || res.RiskReason == "" {
		log.Printf("AI scoring returned empty reasons — retrying with strict prompt.")
		strictPrompt := promptStr + "\n\nCRITICAL: Every reason field MUST be a non-empty string explaining your score in 1-2 sentences. Do NOT return empty strings."
		res2, err2 := client.ScoreModules(ctx, aiJson, strictPrompt)
		if err2 == nil && res2 != nil {
			// Merge: use retry result for any field that was originally empty
			if res.PromoterReason == "" {
				res.PromoterReason = res2.PromoterReason
				res.PromoterScore = res2.PromoterScore
			}
			if res.IndustryReason == "" {
				res.IndustryReason = res2.IndustryReason
				res.IndustryScore = res2.IndustryScore
			}
			if res.RiskReason == "" {
				res.RiskReason = res2.RiskReason
				res.RiskScore = res2.RiskScore
			}
		}
	}

	// Final guard: if reasons are still empty after retry, use descriptive fallbacks
	// so the UI always shows something meaningful rather than blank cards.
	if res.PromoterReason == "" {
		if res.PromoterScore >= 7 {
			res.PromoterReason = "Promoter background appears sound based on available DRHP data."
		} else if res.PromoterScore >= 4 {
			res.PromoterReason = "Promoter background has some concerns based on DRHP data."
		} else {
			res.PromoterReason = "Significant promoter-related concerns identified in the DRHP."
		}
	}
	if res.IndustryReason == "" {
		if res.IndustryScore >= 7 {
			res.IndustryReason = "Sector shows strong tailwinds and growth potential."
		} else if res.IndustryScore >= 4 {
			res.IndustryReason = "Sector shows moderate growth with some headwinds."
		} else {
			res.IndustryReason = "Sector faces significant challenges or declining outlook."
		}
	}
	if res.RiskReason == "" {
		if res.RiskScore >= 7 {
			res.RiskReason = "Risk profile appears clean with minimal red flags in the DRHP."
		} else if res.RiskScore >= 4 {
			res.RiskReason = "Moderate risk profile with some concerns identified."
		} else {
			res.RiskReason = "Multiple red flags or significant risks identified in the DRHP."
		}
	}

	return moduleResult{res.PromoterScore, res.PromoterReason},
		moduleResult{res.IndustryScore, res.IndustryReason},
		moduleResult{res.RiskScore, res.RiskReason}
}

// loadScorerPrompt loads the scorer prompt from known paths, falling back to a
// comprehensive embedded prompt so scoring always produces well-structured output.
func loadScorerPrompt() string {
	// Try Docker volume path first, then relative paths for local dev.
	for _, path := range []string{"/prompts/scorer_v1.md", "../prompts/scorer_v1.md", "prompts/scorer_v1.md"} {
		if b, err := os.ReadFile(path); err == nil {
			return string(b)
		}
	}
	// Embedded fallback — matches the real prompt structure so Gemini always
	// returns all three scores AND non-empty reason strings.
	return `You are an expert IPO financial analyst. Evaluate three modules based on the provided DRHP JSON analysis and return scores with detailed justifications.

Score each module from 0 to 10. You MUST write a non-empty 1-2 sentence reason for every score.

### Modules:
1. Promoter Risk (0-10): 10=experienced/clean, 5=average/minor concerns, 0=litigation/severe related-party/unqualified.
2. Industry Outlook (0-10): 10=high growth/tailwinds, 5=moderate/cyclical, 0=declining/severe regulatory risk.
3. General Risk / Red Flags (0-10): 10=clean/minimal debt/long contracts, 5=standard risks, 0=severe red flags.

Return ONLY this JSON (no markdown, no extra text):
{
  "promoter_score": <0-10>,
  "promoter_reason": "<1-2 sentences explaining the promoter score>",
  "industry_score": <0-10>,
  "industry_reason": "<1-2 sentences explaining the industry score>",
  "risk_score": <0-10>,
  "risk_reason": "<1-2 sentences explaining the risk score>"
}`
}
