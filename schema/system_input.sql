-- name: CreateGrowMediumSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateTankSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateFieldSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status)
VALUES
    ($1, $2, $3, $4)
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
