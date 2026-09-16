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

-- name: CountIPOsWithNonZeroSubscriptions :one
SELECT COUNT(*) FROM (
    SELECT ipo_id
    FROM subscription_data
    WHERE times_subscribed IS NOT NULL AND times_subscribed > 0
    GROUP BY ipo_id
    HAVING COUNT(DISTINCT category) >= 4
) AS sub;

