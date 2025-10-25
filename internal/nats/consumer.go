package nats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	Fetch(batch int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error)
	CachedInfo() *jetstream.ConsumerInfo
}

// JetstreamMessageHandler is a function that processes a JetStream message and returns an error if processing fails.
type JetstreamMessageHandler func(msgs ...jetstream.Msg) BatchResult

// ConsumerHandler wraps a JetStream consumer and message handler for processing messages.
type ConsumerHandler struct {
	jsConsumer  JetstreamConsumer
	handlerFunc JetstreamMessageHandler
	batchSize   int
	maxWait     time.Duration
}

// NewConsumerHandler creates a new consumer handler that processes messages using the provided handler function.
func NewConsumerHandler(consumer JetstreamConsumer, handler JetstreamMessageHandler, batchSize int, maxWait time.Duration) (*ConsumerHandler, error) {
	c := &ConsumerHandler{
		jsConsumer:  consumer,
		handlerFunc: handler,
		batchSize:   batchSize,
		maxWait:     maxWait,
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

			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					batch, err := handler.jsConsumer.Fetch(
						handler.batchSize,
						jetstream.FetchMaxWait(handler.maxWait),
					)
					if err != nil {
						return stacktrace.NewStackTraceError(err)
					}

					msgChan := batch.Messages()
					msgs := []jetstream.Msg{}
					for msg := range msgChan {
						msgs = append(msgs, msg)
					}

					if batch.Error() != nil {
						return stacktrace.NewStackTraceError(batch.Error())
					}

					result := handler.handlerFunc(msgs...)

					AckMessages(result.AckMsgs)

					if result.Error != nil {
						NakMessages(result.NakMsgs)
						return result.Error
					}
				}
			}
		}
	}
}

func AckMessages(msgs []jetstream.Msg) {
	for _, msg := range msgs {
		err := msg.Ack()
		if err != nil {
			slog.Error(
				"failed to ack message",
				slog.String("subject", msg.Subject()),
				stacktrace.ErrorAttribute(err),
			)
		}
	}
}

func NakMessages(msgs []jetstream.Msg) {
	for _, msg := range msgs {
		err := msg.Nak()
		if err != nil {
			slog.Error(
				"failed to nak message",
				slog.String("subject", msg.Subject()),
				stacktrace.ErrorAttribute(err),
			)
		}
	}
}
