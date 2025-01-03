-- name: CreateEndDevice :one
INSERT INTO
    end_devices (id, system_id, network_server_id, system_input_id, name, status)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING
    *;
