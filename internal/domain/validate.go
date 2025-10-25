package domain

import "errors"

// Validate is a function that validates protobuf messages according to their validation rules.
type Validate func(msg any) error

var (
	// ErrInvalidMessageFormat is returned when a message fails validation.
	ErrInvalidMessageFormat = errors.New("invalid message format")
)
