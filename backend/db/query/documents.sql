-- name: CreateDocument :one
INSERT INTO documents (
    ipo_id,
    file_path,
    type,
    downloaded_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;
