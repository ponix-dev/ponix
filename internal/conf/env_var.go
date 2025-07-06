package conf

import (
	"errors"
	"os"
)

var (
	ErrMissingEnv = errors.New("missing env value")
)

func FromEnv(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrMissingEnv
	}

	return v, nil
}
