package telemetry

import (
	"context"
	"time"

	"github.com/ponix-dev/ponix/internal/runner"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	metricExportInterval = time.Second * time.Duration(5)
)

// NewMeterProvider creates and configures an OpenTelemetry meter provider with OTLP
// HTTP exporter for metrics collection. Metrics are exported periodically based on the
// configured interval.
func NewMeterProvider(ctx context.Context, resource *resource.Resource) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricExportInterval))),
	)

	return provider, nil
}

// MeterProviderCloser returns a runner function that gracefully shuts down the meter
// provider, flushing any pending metrics before termination.
func MeterProviderCloser(mp *sdkmetric.MeterProvider) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			return mp.Shutdown(ctx)
		}
	}
}

// SetServiceMeter registers the given meter provider as the global OpenTelemetry
// meter provider for the application.
func SetServiceMeter(meterProvider metric.MeterProvider) {
	otel.SetMeterProvider(meterProvider)
}

// Meter returns an OpenTelemetry meter configured with the instrumentation name and
// semantic conventions schema for recording metrics.
func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(
		instrumentationName,
		metric.WithSchemaURL(semconv.SchemaURL),
	)
}
