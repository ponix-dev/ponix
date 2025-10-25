package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// ConnectionOption configures a NATS connection.
type ConnectionOption func(*nats.Options)

// WithURL sets the NATS server URL.
func WithURL(url string) ConnectionOption {
	return func(o *nats.Options) {
		o.Url = url
	}
}

// WithName sets the client name for the connection.
func WithName(name string) ConnectionOption {
	return func(o *nats.Options) {
		o.Name = name
	}
}

// NewConnection creates a new connection to a NATS server.
// Default URL is nats://localhost:4222 if not specified.
func NewConnection(opts ...ConnectionOption) (*nats.Conn, error) {
	options := nats.GetDefaultOptions()
	options.Url = nats.DefaultURL

	for _, opt := range opts {
		opt(&options)
	}

	nc, err := options.Connect()
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	return nc, nil
}

// NewJetStream creates a new JetStream context from a NATS connection.
func NewJetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	return js, nil
}
