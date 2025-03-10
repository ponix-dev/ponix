// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: gateway.sql

package sqlc

import (
	"context"
)

const createGateway = `-- name: CreateGateway :one
INSERT INTO
    gateways (id, system_id, network_server_id, name, status)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    id, system_id, network_server_id, name, status
`

type CreateGatewayParams struct {
	ID              string
	SystemID        string
	NetworkServerID string
	Name            string
	Status          int32
}

func (q *Queries) CreateGateway(ctx context.Context, arg CreateGatewayParams) (Gateway, error) {
	row := q.db.QueryRow(ctx, createGateway,
		arg.ID,
		arg.SystemID,
		arg.NetworkServerID,
		arg.Name,
		arg.Status,
	)
	var i Gateway
	err := row.Scan(
		&i.ID,
		&i.SystemID,
		&i.NetworkServerID,
		&i.Name,
		&i.Status,
	)
	return i, err
}
