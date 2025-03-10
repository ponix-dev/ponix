// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

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
	MediumType int32
}

type NetworkServer struct {
	ID          string
	SystemID    string
	Name        string
	Status      int32
	IotPlatform int32
}

type Organization struct {
	ID     string
	Name   string
	Status int32
}

type System struct {
	ID             string
	OrganizationID string
	Name           string
	Status         int32
}

type SystemInput struct {
	ID       string
	SystemID string
	Name     string
	Status   int32
}

type Tank struct {
	ID string
}

type User struct {
	ID             string
	OrganizationID string
	FirstName      string
	LastName       string
	Status         int32
}
