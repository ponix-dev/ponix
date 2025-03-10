// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user.sql

package sqlc

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO
    users (id, organization_id, first_name, last_name, status)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    id, organization_id, first_name, last_name, status
`

type CreateUserParams struct {
	ID             string
	OrganizationID string
	FirstName      string
	LastName       string
	Status         int32
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.ID,
		arg.OrganizationID,
		arg.FirstName,
		arg.LastName,
		arg.Status,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.FirstName,
		&i.LastName,
		&i.Status,
	)
	return i, err
}
