-- name: CreateGrowMediumSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status, grow_medium_id)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateTankSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status, tank_id)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateFieldSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status, field_id)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateGrowMedium :one
INSERT INTO
    grow_mediums (id, medium_type)
VALUES
    ($1, $2)
RETURNING
    *;

-- name: CreateTank :one
INSERT INTO
    tanks (id)
VALUES
    ($1)
RETURNING
    *;

-- name: CreateField :one
INSERT INTO
    fields (id)
VALUES
    ($1)
RETURNING
    *;
