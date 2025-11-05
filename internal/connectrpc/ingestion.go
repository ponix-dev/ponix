package connectrpc

import (
	"context"
	"fmt"
	"time"

	envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DataEnvelopeManager handles device data ingestion operations
type DataEnvelopeManager interface {
	IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope, organizationID string) error
}

// IngestionHandler implements ConnectRPC handlers for device data ingestion
type IngestionHandler struct {
	envelopeManager DataEnvelopeManager
}

// NewIngestionHandler creates a new ingestion handler
func NewIngestionHandler(envelopeManager DataEnvelopeManager) *IngestionHandler {
	return &IngestionHandler{
		envelopeManager: envelopeManager,
	}
}

// IngestDeviceData handles RPC requests to ingest device telemetry data
// No authorization required for MVP (future enhancement)
func (handler *IngestionHandler) IngestDeviceData(
	ctx context.Context,
	req *connect.Request[iotv1.IngestDeviceDataRequest],
) (*connect.Response[iotv1.IngestDeviceDataResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestDeviceData")
	defer span.End()

	// Validate required fields
	if req.Msg.GetOrganizationId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization_id is required"))
	}
	if req.Msg.GetEndDeviceId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("end_device_id is required"))
	}
	if req.Msg.GetData() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("data is required"))
	}

	// Use provided occurred_at or default to now
	occurredAt := req.Msg.GetOccurredAt()
	if occurredAt == nil {
		occurredAt = timestamppb.New(time.Now().UTC())
	}

	// Build DataEnvelope
	envelope := envelopev1.DataEnvelope_builder{
		EndDeviceId: req.Msg.GetEndDeviceId(),
		OccurredAt:  occurredAt,
		Data:        req.Msg.GetData(),
	}.Build()

	// Ingest with validation (checks device exists and belongs to org)
	err := handler.envelopeManager.IngestDataEnvelope(ctx, envelope, req.Msg.GetOrganizationId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ingestion failed: %w", err))
	}

	// Return success
	return connect.NewResponse(iotv1.IngestDeviceDataResponse_builder{
		Success: true,
		Message: "Data ingested successfully",
	}.Build()), nil
}
