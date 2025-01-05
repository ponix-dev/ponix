-- name: CreateSystem :one
INSERT INTO
    systems (id, organization_id, name, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetSystem :one
SELECT
    *
FROM
    systems
WHERE
    id = $1 LIMIT 1;


-- name: GetSystemNetworkServers :many
SELECT
    *
FROM
    network_servers
WHERE
    system_id = $1;

-- name: GetSystemGateways :many
SELECT
    *
FROM
    gateways
WHERE
    system_id = $1;

-- name: GetSystemEndDevices :many
SELECT
    *
FROM
    end_devices
WHERE
    system_id = $1;

-- name: GetSystemInput :one
SELECT
    system_inputs.name,
    system_inputs.id,
    system_inputs.system_id,
    system_inputs.status,
    sqlc.embed(fields),
    sqlc.embed(tanks),
    sqlc.embed(grow_mediums)
FROM
    system_inputs
    JOIN grow_mediums ON system_inputs.grow_medium_id = grow_mediums.id
    JOIN tanks ON system_inputs.tank_id = tanks.id
    JOIN fields ON system_inputs.field_id = fields.id
WHERE
    system_inputs.id = $1 LIMIT 1;


-- name: GetSystemInputs :many
SELECT
    system_inputs.name,
    system_inputs.id,
    system_inputs.system_id,
    system_inputs.status,
    sqlc.embed(fields),
    sqlc.embed(tanks),
    sqlc.embed(grow_mediums)
FROM
    system_inputs
    JOIN grow_mediums ON system_inputs.grow_medium_id = grow_mediums.id
    JOIN tanks ON system_inputs.tank_id = tanks.id
    JOIN fields ON system_inputs.field_id = fields.id
WHERE
    system_id = $1;
