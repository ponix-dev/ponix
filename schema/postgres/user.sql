-- name: CreateUser :one
INSERT INTO
    users (id, first_name, last_name, email, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    email = EXCLUDED.email,
    updated_at = EXCLUDED.updated_at
RETURNING
    *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = $1;
