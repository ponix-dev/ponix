// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: end_device.sql

package sqlc

import (
	"context"
)

const createEndDevice = `-- name: CreateEndDevice :one
INSERT INTO
    end_devices (id, system_id, network_server_id, system_input_id, name, status)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING
    id, system_id, network_server_id, system_input_id, name, status
`

type CreateEndDeviceParams struct {
	ID              string
	SystemID        string
	NetworkServerID string
	SystemInputID   string
	Name            string
	Status          int32
}

func (q *Queries) CreateEndDevice(ctx context.Context, arg CreateEndDeviceParams) (EndDevice, error) {
	row := q.db.QueryRow(ctx, createEndDevice,
		arg.ID,
		arg.SystemID,
		arg.NetworkServerID,
		arg.SystemInputID,
		arg.Name,
		arg.Status,
	)
	var i EndDevice
	err := row.Scan(
		&i.ID,
		&i.SystemID,
		&i.NetworkServerID,
		&i.SystemInputID,
		&i.Name,
		&i.Status,
	)
	return i, err
}
