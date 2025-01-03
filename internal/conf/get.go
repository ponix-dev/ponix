package conf

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

func GetConfig[T any](ctx context.Context) (T, error) {
	var cfg T
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
