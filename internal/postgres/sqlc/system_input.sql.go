// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: system_input.sql

package sqlc

import (
	"context"
)

const createField = `-- name: CreateField :one
INSERT INTO
    fields (id)
VALUES
    ($1)
RETURNING
    id
`

func (q *Queries) CreateField(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRow(ctx, createField, id)
	err := row.Scan(&id)
	return id, err
}

const createFieldSystemInput = `-- name: CreateFieldSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    id, system_id, name, status
`

type CreateFieldSystemInputParams struct {
	ID       string
	SystemID string
	Name     string
	Status   int32
}

func (q *Queries) CreateFieldSystemInput(ctx context.Context, arg CreateFieldSystemInputParams) (SystemInput, error) {
	row := q.db.QueryRow(ctx, createFieldSystemInput,
		arg.ID,
		arg.SystemID,
		arg.Name,
		arg.Status,
	)
	var i SystemInput
	err := row.Scan(
		&i.ID,
		&i.SystemID,
		&i.Name,
		&i.Status,
	)
	return i, err
}

const createGrowMedium = `-- name: CreateGrowMedium :one
INSERT INTO
    grow_mediums (id, medium_type)
VALUES
    ($1, $2)
RETURNING
    id, medium_type
`

type CreateGrowMediumParams struct {
	ID         string
	MediumType int32
}

func (q *Queries) CreateGrowMedium(ctx context.Context, arg CreateGrowMediumParams) (GrowMedium, error) {
	row := q.db.QueryRow(ctx, createGrowMedium, arg.ID, arg.MediumType)
	var i GrowMedium
	err := row.Scan(&i.ID, &i.MediumType)
	return i, err
}

const createGrowMediumSystemInput = `-- name: CreateGrowMediumSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    id, system_id, name, status
`

type CreateGrowMediumSystemInputParams struct {
	ID       string
	SystemID string
	Name     string
	Status   int32
}

func (q *Queries) CreateGrowMediumSystemInput(ctx context.Context, arg CreateGrowMediumSystemInputParams) (SystemInput, error) {
	row := q.db.QueryRow(ctx, createGrowMediumSystemInput,
		arg.ID,
		arg.SystemID,
		arg.Name,
		arg.Status,
	)
	var i SystemInput
	err := row.Scan(
		&i.ID,
		&i.SystemID,
		&i.Name,
		&i.Status,
	)
	return i, err
}

const createTank = `-- name: CreateTank :one
INSERT INTO
    tanks (id)
VALUES
    ($1)
RETURNING
    id
`

func (q *Queries) CreateTank(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRow(ctx, createTank, id)
	err := row.Scan(&id)
	return id, err
}

const createTankSystemInput = `-- name: CreateTankSystemInput :one
INSERT INTO
    system_inputs (id, system_id, name, status)
VALUES
    ($1, $2, $3, $4)
RETURNING
    id, system_id, name, status
`

type CreateTankSystemInputParams struct {
	ID       string
	SystemID string
	Name     string
	Status   int32
}

func (q *Queries) CreateTankSystemInput(ctx context.Context, arg CreateTankSystemInputParams) (SystemInput, error) {
	row := q.db.QueryRow(ctx, createTankSystemInput,
		arg.ID,
		arg.SystemID,
		arg.Name,
		arg.Status,
	)
	var i SystemInput
	err := row.Scan(
		&i.ID,
		&i.SystemID,
		&i.Name,
		&i.Status,
	)
	return i, err
}