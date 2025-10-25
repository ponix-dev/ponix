package conf

import (
	"errors"
	"os"
)

var (
	// ErrMissingEnv is returned when a required environment variable is not set.
	ErrMissingEnv = errors.New("missing env value")
)

// FromEnv retrieves the value of an environment variable by key.
// Returns ErrMissingEnv if the environment variable is not set.
func FromEnv(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrMissingEnv
	}

	return v, nil
}
