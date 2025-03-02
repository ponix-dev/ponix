package telemetry

var (
	instrumentationName = "github.com/ponix-dev/ponix"
)

func SetInstrumentationName(name string) {
	instrumentationName = name
}
