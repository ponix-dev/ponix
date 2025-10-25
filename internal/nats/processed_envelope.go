package nats

import (
	"context"
	"fmt"

	envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
	"google.golang.org/protobuf/proto"
)

// JetstreamPublisher defines the interface for publishing messages to JetStream.
type JetstreamPublisher interface {
	Publish(ctx context.Context, subject string, payload []byte, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error)
}

// ProcessedEnvelopeProducer publishes ProcessedEnvelope messages to a NATS JetStream topic.
type ProcessedEnvelopeProducer struct {
	js      JetstreamPublisher
	subject string
}

// NewProcessedEnvelopeProducer creates a new NATS JetStream producer for ProcessedEnvelope messages.
// The producer publishes to the specified subject.
func NewProcessedEnvelopeProducer(js JetstreamPublisher, subject string) *ProcessedEnvelopeProducer {
	p := &ProcessedEnvelopeProducer{
		js:      js,
		subject: subject,
	}

	return p
}

// ProduceProcessedEnvelope publishes a ProcessedEnvelope to the configured NATS topic.
// The envelope is serialized as protobuf before publishing.
func (p *ProcessedEnvelopeProducer) ProduceProcessedEnvelope(ctx context.Context, envelope *envelopev1.ProcessedEnvelope) error {
	ctx, span := telemetry.Tracer().Start(ctx, "ProduceProcessedEnvelope")
	defer span.End()

	// Serialize the envelope to protobuf
	data, err := proto.Marshal(envelope)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	subject := fmt.Sprintf("%s.%s", p.subject, envelope.GetEndDeviceId())

	// Publish to JetStream
	_, err = p.js.Publish(ctx, subject, data)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}

// ProcessedEnvelopeIngester defines the interface for processing received ProcessedEnvelope messages.
type ProcessedEnvelopeIngester interface {
	IngestProcessedEnvelope(ctx context.Context, envelopes ...*envelopev1.ProcessedEnvelope) error
}

// NewProcessedEnvelopeMessageHandler creates a JetStream message handler that unmarshals and processes ProcessedEnvelope messages.
func NewProcessedEnvelopeMessageHandler(ingester ProcessedEnvelopeIngester) JetstreamMessageHandler {
	return func(msgs ...jetstream.Msg) BatchResult {
		//TODO: this currently doesn't support tracing between producer and consumer
		ctx, span := telemetry.Tracer().Start(context.Background(), "ProcessedEnvelopeConsumer")
		defer span.End()

		result := BatchResult{
			AckMsgs: make([]jetstream.Msg, len(msgs)),
			NakMsgs: make([]jetstream.Msg, len(msgs)),
			Error:   nil,
		}

		envelopes := make([]*envelopev1.ProcessedEnvelope, 0, len(msgs))
		for _, msg := range msgs {

			envelope := &envelopev1.ProcessedEnvelope{}
			err := proto.Unmarshal(msg.Data(), envelope)
			if err != nil {
				span.RecordError(err)
				result.Error = stacktrace.NewStackTraceError(err)
				result.NakMsgs = msgs

				return result
			}

			envelopes = append(envelopes, envelope)
		}

		err := ingester.IngestProcessedEnvelope(ctx, envelopes...)
		if err != nil {
			span.RecordError(err)
			result.Error = stacktrace.NewStackTraceError(err)
			result.NakMsgs = msgs

			return result
		}

		result.AckMsgs = msgs

		return result
	}
}
