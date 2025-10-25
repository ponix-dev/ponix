package connectrpc

import (
	"context"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// EndDeviceManager handles end device business operations.
type EndDeviceManager interface {
	CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organization string) (*iotv1.EndDevice, error)
}

// EndDeviceAuthorizer checks permissions for end device operations.
type EndDeviceAuthorizer interface {
	CanCreateEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
	CanReadEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
	CanUpdateEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
	CanDeleteEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
}

// EndDeviceHandler implements Connect RPC handlers for end device operations.
type EndDeviceHandler struct {
	endDeviceManager EndDeviceManager
	authorizer       EndDeviceAuthorizer
}

// NewEndDeviceHandler creates a new EndDeviceHandler with the provided dependencies.
func NewEndDeviceHandler(edmgr EndDeviceManager, authorizer EndDeviceAuthorizer) *EndDeviceHandler {
	return &EndDeviceHandler{
		endDeviceManager: edmgr,
		authorizer:       authorizer,
	}
}

// CreateEndDevice handles RPC requests to create a new end device.
// Requires super admin privileges or device creation permission in the organization.
// Organization ID can be provided in the request or via X-Organization-ID header.
func (handler *EndDeviceHandler) CreateEndDevice(ctx context.Context, req *connect.Request[iotv1.CreateEndDeviceRequest]) (*connect.Response[iotv1.CreateEndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	// Extract organization from request
	organization := GetOrganizationFromRequest(req.Msg)
	if organization == "" {
		// Try to get from headers as fallback
		organization = req.Header().Get("X-Organization-ID")
		if organization == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization ID is required"))
		}
	}

	userId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanCreateEndDevice(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}

		allowed = can
	}

	if !allowed {
		userId, _ := domain.GetUserFromContext(ctx)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to create end devices in organization %s", userId, organization))
	}

	endDevice, err := handler.endDeviceManager.CreateEndDevice(ctx, req.Msg, organization)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.CreateEndDeviceResponse_builder{
		EndDevice: endDevice,
	}.Build())

	return resp, nil
}

// EndDevice handles RPC requests to retrieve a single end device.
// Requires super admin privileges or device read permission in the organization.
func (handler *EndDeviceHandler) EndDevice(ctx context.Context, req *connect.Request[iotv1.EndDeviceRequest]) (*connect.Response[iotv1.EndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDevice")
	defer span.End()

	userId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	// Extract organization from request
	organization := GetOrganizationFromRequest(req.Msg)
	if organization == "" {
		// Try to get from headers as fallback
		organization = req.Header().Get("X-Organization-ID")
		if organization == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization ID is required"))
		}
	}

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadEndDevice(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to read end devices in organization %s", userId, organization))
	}

	return nil, nil
}

// OrganizationEndDevices handles RPC requests to list all end devices in an organization.
// Requires super admin privileges or device read permission in the organization.
func (handler *EndDeviceHandler) OrganizationEndDevices(ctx context.Context, req *connect.Request[iotv1.OrganizationEndDevicesRequest]) (*connect.Response[iotv1.OrganizationEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "OrganizationEndDevices")
	defer span.End()

	userId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	organization := req.Msg.GetOrganizationId()

	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadEndDevice(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to read end devices in organization %s", userId, organization))
	}

	resp := connect.NewResponse(iotv1.OrganizationEndDevicesResponse_builder{}.Build())

	return resp, nil
}

// EndDeviceData handles RPC requests to retrieve time-series data for an end device.
// Requires super admin privileges or device read permission in the organization.
func (handler *EndDeviceHandler) EndDeviceData(ctx context.Context, req *connect.Request[iotv1.EndDeviceDataRequest]) (*connect.Response[iotv1.EndDeviceDataResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDeviceData")
	defer span.End()

	userId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	// Extract organization from request
	organization := GetOrganizationFromRequest(req.Msg)
	if organization == "" {
		// Try to get from headers as fallback
		organization = req.Header().Get("X-Organization-ID")
		if organization == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization ID is required"))
		}
	}

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadEndDevice(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}

		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to read end device data in organization %s", userId, organization))
	}

	resp := connect.NewResponse(iotv1.EndDeviceDataResponse_builder{}.Build())

	return resp, nil
}
