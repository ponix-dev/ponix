-- ===== End Devices (Generic) =====

-- name: CreateEndDevice :one
INSERT INTO end_devices (id, name, description, organization_id, status, data_type, hardware_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEndDevice :one
SELECT * FROM end_devices
WHERE id = $1;

-- name: ListEndDevicesByOrganization :many
SELECT * FROM end_devices
WHERE organization_id = $1
ORDER BY name;

-- name: UpdateEndDevice :one
UPDATE end_devices
SET name = $2, description = $3, status = $4, data_type = $5, hardware_type = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEndDevice :exec
DELETE FROM end_devices
WHERE id = $1;

-- name: GetEndDeviceWithOrganization :one
SELECT id, name, description, organization_id, status, data_type, hardware_type, created_at, updated_at
FROM end_devices
WHERE id = $1;
