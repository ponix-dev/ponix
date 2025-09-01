package connectrpc

import (
	"context"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type EndDeviceManager interface {
	CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organization string) (*iotv1.EndDevice, error)
}

type EndDeviceAuthorizer interface {
	CanCreateEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
	CanReadEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
	CanUpdateEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
	CanDeleteEndDevice(ctx context.Context, userId string, organizationId string) (bool, error)
}

type EndDeviceHandler struct {
	endDeviceManager EndDeviceManager
	authorizer       EndDeviceAuthorizer
}

func NewEndDeviceHandler(edmgr EndDeviceManager, authorizer EndDeviceAuthorizer) *EndDeviceHandler {
	return &EndDeviceHandler{
		endDeviceManager: edmgr,
		authorizer:       authorizer,
	}
}

func (handler *EndDeviceHandler) CreateEndDevice(ctx context.Context, req *connect.Request[iotv1.CreateEndDeviceRequest]) (*connect.Response[iotv1.CreateEndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	// TODO: we should just add this to the protobuf definition
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
