package telemetry

import (
	"context"
	"time"

	"github.com/ponix-dev/ponix/internal/runner"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
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

var meter metric.Meter
var meterSet = false

func SetServiceMeter(meterProvider metric.MeterProvider) {
	if meter == nil {
		meter = meterProvider.Meter(
			instrumentationName,
			metric.WithSchemaURL(semconv.SchemaURL),
		)
		meterSet = true
		otel.SetMeterProvider(meterProvider)
	}
}

func Meter() metric.Meter {
	if meterSet {
		return meter
	}

	return noop.NewMeterProvider().Meter(instrumentationName)
}
