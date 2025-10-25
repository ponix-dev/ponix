package telemetry

var (
	instrumentationName = "github.com/ponix-dev/ponix"
)

// SetInstrumentationName configures the instrumentation name used for OpenTelemetry
// meters and tracers across the telemetry package.
func SetInstrumentationName(name string) {
	instrumentationName = name
}
