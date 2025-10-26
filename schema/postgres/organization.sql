-- name: CreateOrganization :one
INSERT INTO
    organizations (id, name, status, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at
RETURNING
    *;

-- name: GetOrganization :one
SELECT
    *
FROM
    organizations
WHERE
    id = $1;

-- name: GetUserOrganizationsWithDetails :many
SELECT
    o.id,
    o.name,
    o.status,
    o.created_at,
    o.updated_at
FROM
    organizations o
    INNER JOIN user_organizations uo ON o.id = uo.organization_id
WHERE
    uo.user_id = $1;