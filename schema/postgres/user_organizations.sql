-- name: AddUserToOrganization :exec
INSERT INTO user_organizations (user_id, organization_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, organization_id)
DO UPDATE SET 
    role = EXCLUDED.role,
    created_at = COALESCE(user_organizations.created_at, NOW());

-- name: RemoveUserFromOrganization :exec
DELETE FROM user_organizations 
WHERE user_id = $1 AND organization_id = $2;

-- name: GetUserOrganizations :many
SELECT organization_id, role
FROM user_organizations
WHERE user_id = $1;

-- name: GetOrganizationUsers :many
SELECT user_id, role
FROM user_organizations
WHERE organization_id = $1;

-- name: IsUserInOrganization :one
SELECT EXISTS(
    SELECT 1 FROM user_organizations 
    WHERE user_id = $1 AND organization_id = $2
) AS is_member;

-- name: UpdateUserRole :exec
INSERT INTO user_organizations (user_id, organization_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, organization_id)
DO UPDATE SET 
    role = EXCLUDED.role;