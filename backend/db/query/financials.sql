-- name: CreateOrUpdateFinancials :one
INSERT INTO financials (
    ipo_id,
    year,
    revenue,
    pat,
    ebitda,
    total_assets,
    total_debt,
    equity
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (ipo_id, year) DO UPDATE SET
    revenue = EXCLUDED.revenue,
    pat = EXCLUDED.pat,
    ebitda = EXCLUDED.ebitda,
    total_assets = EXCLUDED.total_assets,
    total_debt = EXCLUDED.total_debt,
    equity = EXCLUDED.equity
RETURNING *;

-- name: GetFinancialsByIPO :many
SELECT * FROM financials
WHERE ipo_id = $1
ORDER BY year DESC;

-- name: CountIPOsWithFinancials :one
SELECT COUNT(DISTINCT ipo_id) FROM financials;
