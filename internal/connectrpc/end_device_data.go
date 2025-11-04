package connectrpc

import (
	"context"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// EndDeviceDataManager handles end device data query operations.
type EndDeviceDataManager interface {
	QueryEndDeviceData(ctx context.Context, req *iotv1.QueryEndDeviceDataRequest) (*iotv1.QueryEndDeviceDataResponse, error)
}

// EndDeviceDataHandler implements Connect RPC handlers for end device data operations.
type EndDeviceDataHandler struct {
	endDeviceDataManager EndDeviceDataManager
	authorizer           EndDeviceAuthorizer
}

// NewEndDeviceDataHandler creates a new EndDeviceDataHandler with the provided dependencies.
func NewEndDeviceDataHandler(edDataMgr EndDeviceDataManager, authorizer EndDeviceAuthorizer) *EndDeviceDataHandler {
	return &EndDeviceDataHandler{
		endDeviceDataManager: edDataMgr,
		authorizer:           authorizer,
	}
}

// QueryEndDeviceData handles RPC requests to query time-series sensor data.
// Returns Prometheus-style histogram data for visualization.
// Requires super admin privileges or device read permission in the organization.
func (handler *EndDeviceDataHandler) QueryEndDeviceData(
	ctx context.Context,
	req *connect.Request[iotv1.QueryEndDeviceDataRequest],
) (*connect.Response[iotv1.QueryEndDeviceDataResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "QueryEndDeviceData")
	defer span.End()

	// Extract user from context
	userId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	// Extract organization from request
	organizationID := req.Msg.GetOrganizationId()
	if organizationID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization ID is required"))
	}

	// Authorization check
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadEndDevice(ctx, userId, organizationID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			fmt.Errorf("user %s not authorized to read sensor data in organization %s", userId, organizationID),
		)
	}

	// Query sensor data
	response, err := handler.endDeviceDataManager.QueryEndDeviceData(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to query sensor data: %w", err))
	}

	return connect.NewResponse(response), nil
}
