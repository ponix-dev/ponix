package telemetry

import (
	"context"

	"github.com/ponix-dev/ponix/internal/runner"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func NewPropagator() propagation.TextMapPropagator {
	p := propagation.TraceContext{}
	otel.SetTextMapPropagator(p)

	return p
}

func NewTracerProvider(ctx context.Context, resource *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(otlptracehttp.WithInsecure()))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	return tp, nil
}

func TracerProviderCloser(tp *sdktrace.TracerProvider) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			return tp.Shutdown(ctx)
		}
	}
}

var tracer trace.Tracer
var tracerSet = false

func SetServiceTracer(tracerProvider trace.TracerProvider) {
	if tracer == nil {
		tracer = tracerProvider.Tracer(
			instrumentationName,
			trace.WithSchemaURL(semconv.SchemaURL),
		)
		tracerSet = true
		otel.SetTracerProvider(tracerProvider)
	}
}

func Tracer() trace.Tracer {
	if tracerSet {
		return tracer
	}

	return noop.NewTracerProvider().Tracer(instrumentationName)
}
