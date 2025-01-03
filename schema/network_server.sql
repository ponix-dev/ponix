-- name: CreateNetworkServer :one
INSERT INTO
    network_servers (id, system_id, name, status, iot_platform)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;
