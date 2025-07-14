-- name: CreateOrganization :one
INSERT INTO
    organizations (id, name, status, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetOrganization :one
SELECT
    *
FROM
    organizations
WHERE
    id = $1;
