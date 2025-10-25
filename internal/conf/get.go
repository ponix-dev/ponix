package conf

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// GetConfig loads configuration from environment variables into the provided type T.
// It uses struct tags (env:"VAR_NAME") to map environment variables to struct fields.
// Returns an error if required environment variables are missing or invalid.
func GetConfig[T any](ctx context.Context) (T, error) {
	var cfg T
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
