package domain

import "errors"

type Validate func(msg any) error

var (
	ErrInvalidMessageFormat = errors.New("invalid message format")
)
