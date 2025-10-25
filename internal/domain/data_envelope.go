package domain

import (
	"context"
	"time"

	envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProcessedEnvelopeProducer publishes ProcessedEnvelope messages to a message queue or stream.
type ProcessedEnvelopeProducer interface {
	ProduceProcessedEnvelope(ctx context.Context, envelope *envelopev1.ProcessedEnvelope) error
}

// ProcessedEnvelopeWriter persists ProcessedEnvelope messages to storage.
type ProcessedEnvelopeWriter interface {
	WriteProcessedEnvelope(ctx context.Context, envelope ...*envelopev1.ProcessedEnvelope) error
}

// DataEnvelopeService orchestrates the ingestion and processing of data envelopes.
type DataEnvelopeService struct {
	producer ProcessedEnvelopeProducer
	writer   ProcessedEnvelopeWriter
}

// NewDataEnvelopeService creates a new instance of DataEnvelopeService with the provided producer and writer.
func NewDataEnvelopeService(p ProcessedEnvelopeProducer, w ProcessedEnvelopeWriter) *DataEnvelopeService {
	return &DataEnvelopeService{
		producer: p,
		writer:   w,
	}
}

// IngestDataEnvelope receives a raw data envelope, adds processing metadata, and publishes it to the producer.
func (srv *DataEnvelopeService) IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope) error {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestDataEnvelope")
	defer span.End()

	processedEnvelope := envelopev1.ProcessedEnvelope_builder{
		EndDeviceId: envelope.GetEndDeviceId(),
		OccurredAt:  envelope.GetOccurredAt(),
		Data:        envelope.GetData(),
		ProcessedAt: timestamppb.New(time.Now().UTC()),
	}.Build()

	return srv.producer.ProduceProcessedEnvelope(ctx, processedEnvelope)
}

// IngestProcessedEnvelope receives a processed envelope and persists it via the writer.
// This is typically called by a message consumer after receiving from the producer.
func (srv *DataEnvelopeService) IngestProcessedEnvelope(ctx context.Context, envelopes ...*envelopev1.ProcessedEnvelope) error {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestProcessedEnvelope")
	defer span.End()

	err := srv.writer.WriteProcessedEnvelope(ctx, envelopes...)
	if err != nil {
		return err
	}

	return nil
}
