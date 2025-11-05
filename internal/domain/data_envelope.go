package domain

import (
	"context"
	"time"

	envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
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
	producer       ProcessedEnvelopeProducer
	store          ProcessedEnvelopeStorer
	endDeviceStore EndDeviceStorer
}

// NewDataEnvelopeManager creates a new instance of DataEnvelopeService with the provided producer and store.
func NewDataEnvelopeManager(
	producer ProcessedEnvelopeProducer,
	store ProcessedEnvelopeStorer,
	endDeviceStore EndDeviceStorer,
) *DataEnvelopeManager {
	return &DataEnvelopeManager{
		producer:       producer,
		store:          store,
		endDeviceStore: endDeviceStore,
	}
}

// IngestDataEnvelope receives a raw data envelope, adds processing metadata, and publishes it to the producer.
// The organizationID parameter identifies which organization owns the data.
func (mgr *DataEnvelopeManager) IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope, organizationID string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestDataEnvelope")
	defer span.End()

	// VALIDATION: Verify device exists and belongs to organization
	_, deviceOrgID, err := mgr.endDeviceStore.GetEndDeviceWithOrganization(ctx, envelope.GetEndDeviceId())
	if err != nil {
		return err
	}

	if deviceOrgID != organizationID {
		return stacktrace.NewStackTraceErrorf(
			"organization mismatch: device %s belongs to %s, but data sent for %s",
			envelope.GetEndDeviceId(),
			deviceOrgID,
			organizationID,
		)
	}

	// Build ProcessedEnvelope with validation complete
	processedEnvelope := envelopev1.ProcessedEnvelope_builder{
		OrganizationId: organizationID,
		EndDeviceId:    envelope.GetEndDeviceId(),
		OccurredAt:     envelope.GetOccurredAt(),
		Data:           envelope.GetData(),
		ProcessedAt:    timestamppb.New(time.Now().UTC()),
	}.Build()

	// Publish to NATS
	err = mgr.producer.ProduceProcessedEnvelope(ctx, processedEnvelope)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
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
