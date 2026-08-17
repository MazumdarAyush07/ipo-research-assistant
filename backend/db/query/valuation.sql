-- name: CreateOrUpdateValuation :one
INSERT INTO valuation (
    ipo_id,
    issue_price,
    market_cap,
    pe_ratio,
    pb_ratio,
    ev_ebitda
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (ipo_id) DO UPDATE 
SET issue_price = EXCLUDED.issue_price,
    market_cap = EXCLUDED.market_cap,
    pe_ratio = EXCLUDED.pe_ratio,
    pb_ratio = EXCLUDED.pb_ratio,
    ev_ebitda = EXCLUDED.ev_ebitda
RETURNING *;

-- name: GetValuationByIPO :one
SELECT * FROM valuation
WHERE ipo_id = $1 LIMIT 1;
