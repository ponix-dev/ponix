-- name: CreateUser :one
INSERT INTO
    users (id, organization_id, first_name, last_name, status)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;
