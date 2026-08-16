-- name: CreateSubscriptionData :one
INSERT INTO subscription_data (
    ipo_id,
    category,
    times_subscribed
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetLatestSubscription :many
-- Gets the latest entry for each category for a given IPO.
SELECT DISTINCT ON (category) *
FROM subscription_data
WHERE ipo_id = $1
ORDER BY category, recorded_at DESC;
