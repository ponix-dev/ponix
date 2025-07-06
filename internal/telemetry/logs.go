package telemetry

import (
	"context"
	"log/slog"
	"os"

	"github.com/ponix-dev/ponix/internal/runner"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func LogFatal(logger *slog.Logger, msg string, err error) {
	logger.Error(msg, err)
	os.Exit(1)
}

func defaultLoggerShutdown(ctx context.Context) error {
	return nil
}

var loggerShutdown = defaultLoggerShutdown

func LoggerProviderCloser() runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			return loggerShutdown(ctx)
		}
	}
}

func NewLogger(ctx context.Context, resource *resource.Resource, handlerName string) (*slog.Logger, error) {
	exporter, err := otlploghttp.New(ctx, otlploghttp.WithInsecure())
	if err != nil {
		return slog.Default(), err
	}

	bp := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(log.WithProcessor(bp), log.WithResource(resource))
	loggerShutdown = provider.Shutdown

	otelHandler := stacktrace.NewOtelHandlerWrapper(otelslog.NewHandler(handlerName, otelslog.WithLoggerProvider(provider)))
	logger := slog.New(slogmulti.Fanout(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{ReplaceAttr: stacktrace.ReplaceAttr}), // then to second handler: stderr
		otelHandler,
	))

	return logger, nil
}

func SetLogger(logger *slog.Logger) {
	slog.SetDefault(logger)
}

func Logger() *slog.Logger {
	return slog.Default()
}
