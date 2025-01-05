package telemetry

import (
	"context"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewLogger(ctx context.Context, resource *resource.Resource, handlerName string) (*slog.Logger, error) {
	exporter, err := otlploghttp.New(ctx, otlploghttp.WithInsecure())
	if err != nil {
		return slog.Default(), err
	}

	bp := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(log.WithProcessor(bp), log.WithResource(resource))
	otelHandler := otelslog.NewHandler(handlerName, otelslog.WithLoggerProvider(provider))
	logger := slog.New(slogmulti.Fanout(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}), // then to second handler: stderr
		otelHandler,
	))

	slog.SetDefault(logger)

	return logger, nil
}
