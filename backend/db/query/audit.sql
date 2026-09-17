-- name: CountRecentPeersTracked :one
SELECT COUNT(DISTINCT ipo_id)
FROM peer_companies
WHERE ($1::int = 0 OR created_at >= NOW() - ($1::int * INTERVAL '1 hour'));

-- name: CountRecentGMPTracked :one
SELECT COUNT(DISTINCT ipo_id)
FROM gmp_history
WHERE ($1::int = 0 OR recorded_at >= NOW() - ($1::int * INTERVAL '1 hour'));

-- name: CountRecentSubscriptionsTracked :one
SELECT COUNT(DISTINCT ipo_id)
FROM subscription_data
WHERE ($1::int = 0 OR recorded_at >= NOW() - ($1::int * INTERVAL '1 hour'));

-- name: CountRecentValuationsTracked :one
SELECT COUNT(DISTINCT ipo_id)
FROM valuation
WHERE ($1::int = 0 OR updated_at >= NOW() - ($1::int * INTERVAL '1 hour'));
