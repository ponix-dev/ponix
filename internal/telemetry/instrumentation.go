package telemetry

var (
	instrumentationName = ""
)

func SetInstrumentationName(name string) {
	instrumentationName = name
}
