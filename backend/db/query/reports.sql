-- name: CreateOrUpdateReport :one
INSERT INTO reports (
    ipo_id,
    file_path,
    format
) VALUES (
    $1, $2, $3
)
ON CONFLICT (ipo_id, format) DO UPDATE
SET
    file_path = EXCLUDED.file_path,
    generated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetReportByIPO :one
SELECT * FROM reports
WHERE ipo_id = $1 AND format = $2 LIMIT 1;
