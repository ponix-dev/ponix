package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

var (
	streams = []StreamSetup{
		{
			Name:        "processed_envelopes",
			Description: "Stream processed envelopes",
			Subjects:    []string{"processed_envelopes.>"},
		},
	}
)

// StreamSetup defines configuration for a JetStream stream.
type StreamSetup struct {
	Name        string
	Description string
	Subjects    []string
}

// SetupJetStream ensures that all required JetStream streams are created.
// Call this before starting any consumers to ensure the required streams are available.
func SetupJetStream(ctx context.Context, js jetstream.JetStream) error {
	for _, setup := range streams {
		err := EnsureStream(ctx, js, setup.Name, setup.Subjects)
		if err != nil {
			return err
		}
	}

	return nil
}

// EnsureStream creates or updates a JetStream stream if it doesn't exist.
func EnsureStream(ctx context.Context, js jetstream.JetStream, streamName string, subjects []string) error {
	_, err := js.Stream(ctx, streamName)
	if err == nil {
		return nil
	}

	// Stream doesn't exist, create it
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	return nil
}
