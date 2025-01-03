// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: query.sql

package sqlc

import (
	"context"
)

const getSystem = `-- name: GetSystem :one
SELECT id, organization_id, name, status FROM systems
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetSystem(ctx context.Context, id string) (System, error) {
	row := q.db.QueryRow(ctx, getSystem, id)
	var i System
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.Name,
		&i.Status,
	)
	return i, err
}
