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

func MeterProviderCloser(mp *sdkmetric.MeterProvider) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			return mp.Shutdown(ctx)
		}
	}
}

func SetServiceMeter(meterProvider metric.MeterProvider) {
	otel.SetMeterProvider(meterProvider)
}

func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(
		instrumentationName,
		metric.WithSchemaURL(semconv.SchemaURL),
	)
}
