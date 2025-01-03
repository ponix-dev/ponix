// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type EndDevice struct {
	ID              string
	SystemID        string
	NetworkServerID string
	SystemInputID   string
	Name            string
	Status          int32
}

type Field struct {
	ID string
}

type Gateway struct {
	ID              string
	SystemID        string
	NetworkServerID string
	Name            string
	Status          int32
}

type GrowMedium struct {
	ID         string
	MediumType pgtype.Int4
}

type NetworkServer struct {
	ID          string
	SystemID    string
	Name        string
	Status      int32
	IotPlatform int32
}

type System struct {
	ID             string
	OrganizationID string
	Name           string
	Status         int32
}

type SystemInput struct {
	ID           string
	SystemID     string
	Status       int32
	GrowMediumID pgtype.Text
	TankID       pgtype.Text
	FieldID      pgtype.Text
}

type Tank struct {
	ID string
}