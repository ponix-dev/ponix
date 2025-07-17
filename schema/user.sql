-- name: CreateUser :one
INSERT INTO
    users (id, first_name, last_name, email, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = $1;
