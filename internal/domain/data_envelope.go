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

// ProcessedEnvelopeStorer persists ProcessedEnvelope messages to storage.
type ProcessedEnvelopeStorer interface {
	StoreProcessedEnvelopes(ctx context.Context, envelope ...*envelopev1.ProcessedEnvelope) error
}

// DataEnvelopeManager orchestrates the ingestion and processing of data envelopes.
type DataEnvelopeManager struct {
	producer ProcessedEnvelopeProducer
	store    ProcessedEnvelopeStorer
}

// NewDataEnvelopeManager creates a new instance of DataEnvelopeService with the provided producer and store.
func NewDataEnvelopeManager(p ProcessedEnvelopeProducer, w ProcessedEnvelopeStorer) *DataEnvelopeManager {
	return &DataEnvelopeManager{
		producer: p,
		store:    w,
	}
}

// IngestDataEnvelope receives a raw data envelope, adds processing metadata, and publishes it to the producer.
func (mgr *DataEnvelopeManager) IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope) error {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestDataEnvelope")
	defer span.End()

	processedEnvelope := envelopev1.ProcessedEnvelope_builder{
		EndDeviceId: envelope.GetEndDeviceId(),
		OccurredAt:  envelope.GetOccurredAt(),
		Data:        envelope.GetData(),
		ProcessedAt: timestamppb.New(time.Now().UTC()),
	}.Build()

	return mgr.producer.ProduceProcessedEnvelope(ctx, processedEnvelope)
}

// IngestProcessedEnvelope receives a processed envelope and persists it via the writer.
// This is typically called by a message consumer after receiving from the producer.
func (mgr *DataEnvelopeManager) IngestProcessedEnvelope(ctx context.Context, envelopes ...*envelopev1.ProcessedEnvelope) error {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestProcessedEnvelope")
	defer span.End()

	err := mgr.store.StoreProcessedEnvelopes(ctx, envelopes...)
	if err != nil {
		return err
	}

	return nil
}
