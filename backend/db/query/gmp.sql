-- name: CreateGMPHistory :one
INSERT INTO gmp_history (
    ipo_id,
    gmp_amount,
    premium_percent
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetLatestGMP :one
SELECT * FROM gmp_history
WHERE ipo_id = $1
ORDER BY recorded_at DESC
LIMIT 1;
