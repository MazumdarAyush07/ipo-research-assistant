-- name: CreateOrUpdateAIAnalysis :one
INSERT INTO ai_analysis (
    ipo_id,
    business_model,
    moat,
    promoter_risk,
    legal_cases,
    customer_concentration,
    debt_assessment,
    key_risks,
    red_flags,
    management_assumptions,
    industry_outlook,
    raw_json
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
ON CONFLICT (ipo_id) DO UPDATE SET
    business_model = EXCLUDED.business_model,
    moat = EXCLUDED.moat,
    promoter_risk = EXCLUDED.promoter_risk,
    legal_cases = EXCLUDED.legal_cases,
    customer_concentration = EXCLUDED.customer_concentration,
    debt_assessment = EXCLUDED.debt_assessment,
    key_risks = EXCLUDED.key_risks,
    red_flags = EXCLUDED.red_flags,
    management_assumptions = EXCLUDED.management_assumptions,
    industry_outlook = EXCLUDED.industry_outlook,
    raw_json = EXCLUDED.raw_json
RETURNING *;

-- name: GetAIAnalysisByIPO :one
SELECT * FROM ai_analysis
WHERE ipo_id = $1;

-- name: CountIPOsWithAIAnalysis :one
SELECT COUNT(*) FROM ai_analysis WHERE red_flags IS NOT NULL;

-- name: CountParsedIPOs :one
SELECT COUNT(*) FROM ai_analysis WHERE raw_json IS NOT NULL;

