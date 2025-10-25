package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/ponix-dev/ponix/internal/runner"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// NewJetStreamConsumer creates or updates a durable JetStream consumer with the given configuration.
func NewJetStreamConsumer(ctx context.Context, js jetstream.JetStream, streamName string, consumerName string, subject string) (jetstream.Consumer, error) {
	consumer, err := js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return consumer, nil
}

// JetstreamConsumer defines the interface for consuming messages from JetStream.
type JetstreamConsumer interface {
	Consume(handler jetstream.MessageHandler, opts ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)
	CachedInfo() *jetstream.ConsumerInfo
}

// ConsumerHandler wraps a JetStream consumer and message handler for processing messages.
type ConsumerHandler struct {
	jsConsumer  JetstreamConsumer
	handlerFunc JetstreamMessageHandler
}

// NewConsumerHandler creates a new consumer handler that processes messages using the provided handler function.
func NewConsumerHandler(consumer JetstreamConsumer, handler JetstreamMessageHandler) (*ConsumerHandler, error) {
	c := &ConsumerHandler{
		jsConsumer:  consumer,
		handlerFunc: handler,
	}

	return c, nil
}

// ConsumerRunner returns a runner function that starts the consumer and handles graceful shutdown.
func ConsumerRunner(handler *ConsumerHandler) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			streamInfo := handler.jsConsumer.CachedInfo()

			slog.Info(
				"starting jetstream consumer",
				slog.String("consumer", streamInfo.Name),
				slog.String("stream", streamInfo.Stream),
				slog.String("subject", streamInfo.Config.FilterSubject),
			)

			consumeCtx, err := handler.jsConsumer.Consume(ConsumerAckWrapper(handler.handlerFunc))
			if err != nil {
				return stacktrace.NewStackTraceError(err)
			}

			<-ctx.Done()

			consumeCtx.Drain()

			return nil
		}
	}
}

// JetstreamMessageHandler is a function that processes a JetStream message and returns an error if processing fails.
type JetstreamMessageHandler func(msg jetstream.Msg) error

// ConsumerAckWrapper wraps a message handler to automatically ACK successful messages and NAK failed ones.
func ConsumerAckWrapper(handlerFunc JetstreamMessageHandler) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		err := handlerFunc(msg)
		if err != nil {
			slog.Error(
				"Failed to process message",
				slog.String("subject", msg.Subject()),
				stacktrace.ErrorAttribute(err),
			)

			err = msg.Nak()
			if err != nil {
				slog.Error(
					"failed to nak message",
					slog.String("subject", msg.Subject()),
					stacktrace.ErrorAttribute(err),
				)
			}

			return
		}

		err = msg.Ack()
		if err != nil {
			slog.Error(
				"failed to ack message",
				slog.String("subject", msg.Subject()),
				stacktrace.ErrorAttribute(err),
			)
		}
	}
}
