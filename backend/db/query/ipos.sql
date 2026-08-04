-- name: CreateIPO :one
INSERT INTO ipos (
    name,
    exchange_type,
    sector,
    price_band_low,
    price_band_high,
    open_date,
    close_date,
    listing_date,
    status,
    source_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetIPOByName :one
SELECT * FROM ipos
WHERE name = $1 LIMIT 1;

-- name: ListIPOs :many
SELECT * FROM ipos
ORDER BY open_date DESC
LIMIT $1 OFFSET $2;

-- name: GetIPO :one
SELECT * FROM ipos
WHERE id = $1 LIMIT 1;
