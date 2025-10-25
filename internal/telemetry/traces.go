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
)

// NewPropagator creates and registers a TraceContext propagator for distributed tracing
// context propagation across service boundaries.
func NewPropagator() propagation.TextMapPropagator {
	p := propagation.TraceContext{}
	otel.SetTextMapPropagator(p)

	return p
}

// NewTracerProvider creates and configures an OpenTelemetry tracer provider with OTLP
// HTTP exporter for distributed tracing. Traces are batched for efficient export.
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

// TracerProviderCloser returns a runner function that gracefully shuts down the tracer
// provider, flushing any pending trace spans before termination.
func TracerProviderCloser(tp *sdktrace.TracerProvider) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			return tp.Shutdown(ctx)
		}
	}
}

// SetServiceTracer registers the given tracer provider as the global OpenTelemetry
// tracer provider for the application.
func SetServiceTracer(tracerProvider trace.TracerProvider) {
	otel.SetTracerProvider(tracerProvider)
}

// Tracer returns an OpenTelemetry tracer configured with the instrumentation name and
// semantic conventions schema for recording trace spans.
func Tracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(
		instrumentationName,
		trace.WithSchemaURL(semconv.SchemaURL),
	)
}
