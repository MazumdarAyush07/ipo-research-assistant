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

-- name: GetActiveIPOs :many
SELECT * FROM ipos
WHERE status IN ('ACTIVE', 'UPCOMING')
ORDER BY open_date ASC;

-- name: UpdateIPO :one
UPDATE ipos
SET 
    exchange_type = $2,
    open_date = $3,
    close_date = $4,
    status = $5
WHERE id = $1
RETURNING *;

-- name: UpdateIPODetails :one
UPDATE ipos
SET 
    sector = $2,
    price_band_low = $3,
    price_band_high = $4,
    listing_date = $5
WHERE id = $1
RETURNING *;

-- name: CountIPOs :one
SELECT COUNT(*) FROM ipos;

-- name: UpdateIPOSector :one
UPDATE ipos
SET sector = $2
WHERE id = $1
RETURNING *;
