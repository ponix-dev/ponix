-- name: CreateGateway :one
INSERT INTO
    gateways (id, system_id, network_server_id, name, status)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;
