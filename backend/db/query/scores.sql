-- name: CreateOrUpdateScore :one
INSERT INTO scores (
    ipo_id,
    financials_score, financials_reason,
    valuation_score, valuation_reason,
    promoter_score, promoter_reason,
    industry_score, industry_reason,
    risk_score, risk_reason,
    subscription_score,
    gmp_score,
    final_score,
    recommendation,
    scored_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, CURRENT_TIMESTAMP
)
ON CONFLICT (ipo_id) DO UPDATE 
SET financials_score = EXCLUDED.financials_score,
    financials_reason = EXCLUDED.financials_reason,
    valuation_score = EXCLUDED.valuation_score,
    valuation_reason = EXCLUDED.valuation_reason,
    promoter_score = EXCLUDED.promoter_score,
    promoter_reason = EXCLUDED.promoter_reason,
    industry_score = EXCLUDED.industry_score,
    industry_reason = EXCLUDED.industry_reason,
    risk_score = EXCLUDED.risk_score,
    risk_reason = EXCLUDED.risk_reason,
    subscription_score = EXCLUDED.subscription_score,
    gmp_score = EXCLUDED.gmp_score,
    final_score = EXCLUDED.final_score,
    recommendation = EXCLUDED.recommendation,
    scored_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetScoreByIPO :one
SELECT * FROM scores
WHERE ipo_id = $1 LIMIT 1;

-- name: CountIPOsWithScores :one
SELECT COUNT(DISTINCT ipo_id) FROM scores;

