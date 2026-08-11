-- name: InsertPeerCompany :one
INSERT INTO peer_companies (
    ipo_id, name, ticker, pe, pb, ev_ebitda, roe, market_cap
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetPeerCompaniesByIPO :many
SELECT * FROM peer_companies
WHERE ipo_id = $1
ORDER BY market_cap DESC;

-- name: DeletePeerCompaniesByIPO :exec
DELETE FROM peer_companies
WHERE ipo_id = $1;
