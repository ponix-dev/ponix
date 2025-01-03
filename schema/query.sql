-- name: GetSystem :one
SELECT * FROM systems
WHERE id = $1 LIMIT 1;
