-- LoRaWAN CRUD queries for the new three-table structure
-- Based on TTI LoRaWAN Stack v3 and ponix protobuf definitions

-- ===== LoRaWAN Hardware Types =====

-- name: CreateLoRaWANHardwareType :one
INSERT INTO lorawan_hardware_types (id, name, description, manufacturer, model, firmware_version, hardware_version, profile, lorawan_version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetLoRaWANHardwareType :one
SELECT * FROM lorawan_hardware_types
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListLoRaWANHardwareTypes :many
SELECT * FROM lorawan_hardware_types
WHERE deleted_at IS NULL
ORDER BY manufacturer, model;

-- name: UpdateLoRaWANHardwareType :one
UPDATE lorawan_hardware_types
SET name = $2, description = $3, manufacturer = $4, model = $5, 
    firmware_version = $6, hardware_version = $7, profile = $8, lorawan_version = $9, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLoRaWANHardwareType :exec
UPDATE lorawan_hardware_types
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- ===== LoRaWAN Configurations =====

-- name: CreateLoRaWANConfig :one
INSERT INTO lorawan_configs (
    id, end_device_id, device_eui, application_eui, application_id, 
    application_key, network_key, activation_method, 
    frequency_plan_id, hardware_type_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetLoRaWANConfig :one
SELECT * FROM lorawan_configs
WHERE id = $1;

-- name: GetLoRaWANConfigByEndDevice :one
SELECT * FROM lorawan_configs
WHERE end_device_id = $1;

-- name: GetLoRaWANConfigByDeviceEUI :one
SELECT * FROM lorawan_configs
WHERE device_eui = $1;

-- name: ListLoRaWANConfigsByApplication :many
SELECT * FROM lorawan_configs
WHERE application_id = $1
ORDER BY device_eui;

-- name: UpdateLoRaWANConfig :one
UPDATE lorawan_configs
SET device_eui = $2, application_eui = $3, application_id = $4,
    application_key = $5, network_key = $6,
    activation_method = $7, frequency_plan_id = $8, hardware_type_id = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLoRaWANConfig :exec
DELETE FROM lorawan_configs
WHERE id = $1;

-- ===== Combined Queries (Joining End Device with LoRaWAN Config) =====

-- name: GetCompleteLoRaWANDevice :one
SELECT 
    ed.id as end_device_id,
    ed.name,
    ed.description,
    ed.organization_id,
    ed.status,
    ed.data_type,
    ed.hardware_type,
    ed.created_at as device_created_at,
    ed.updated_at as device_updated_at,
    lc.id as lorawan_config_id,
    lc.device_eui,
    lc.application_eui,
    lc.application_id,
    lc.application_key,
    lc.network_key,
    lc.activation_method,
    lc.frequency_plan_id,
    lc.hardware_type_id,
    lht.name as hardware_name,
    lht.manufacturer,
    lht.model,
    lht.firmware_version,
    lht.hardware_version,
    lht.lorawan_version,
    lht.profile
FROM end_devices ed
JOIN lorawan_configs lc ON ed.id = lc.end_device_id
JOIN lorawan_hardware_types lht ON lc.hardware_type_id = lht.id
WHERE ed.id = $1 AND lht.deleted_at IS NULL;

-- name: ListCompleteLoRaWANDevicesByOrganization :many
SELECT 
    ed.id as end_device_id,
    ed.name,
    ed.description,
    ed.organization_id,
    ed.status,
    ed.data_type,
    ed.hardware_type,
    lc.device_eui,
    lc.application_id,
    lht.manufacturer,
    lht.model
FROM end_devices ed
JOIN lorawan_configs lc ON ed.id = lc.end_device_id
JOIN lorawan_hardware_types lht ON lc.hardware_type_id = lht.id
WHERE ed.organization_id = $1 AND lht.deleted_at IS NULL
ORDER BY ed.name;