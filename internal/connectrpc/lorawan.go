package connectrpc

import (
	"context"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type LoRaWANHardwareTypeManager interface {
	CreateLoRaWANHardwareType(ctx context.Context, createReq *iotv1.CreateLoRaWANHardwareTypeRequest) (*iotv1.LoRaWANHardwareData, error)
	GetLoRaWANHardwareType(ctx context.Context, hardwareType string) (*iotv1.LoRaWANHardwareData, error)
	ListLoRaWANHardwareTypes(ctx context.Context) ([]*iotv1.LoRaWANHardwareData, error)
	UpdateLoRaWANHardwareType(ctx context.Context, updateReq *iotv1.UpdateLoRaWANHardwareTypeRequest) (*iotv1.LoRaWANHardwareData, error)
	DeleteLoRaWANHardwareType(ctx context.Context, hardwareType string) error
}

type LoRaWANAuthorizer interface {
	CanCreateLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error)
	CanReadLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error)
	CanUpdateLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error)
	CanDeleteLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error)
}

type LoRaWANHandler struct {
	hardwareTypeManager LoRaWANHardwareTypeManager
	authorizer          LoRaWANAuthorizer
}

func NewLoRaWANHandler(htMgr LoRaWANHardwareTypeManager, authorizer LoRaWANAuthorizer) *LoRaWANHandler {
	return &LoRaWANHandler{
		hardwareTypeManager: htMgr,
		authorizer:          authorizer,
	}
}

func (handler *LoRaWANHandler) CreateLoRaWANHardwareType(ctx context.Context, req *connect.Request[iotv1.CreateLoRaWANHardwareTypeRequest]) (*connect.Response[iotv1.CreateLoRaWANHardwareTypeResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateLoRaWANHardwareType")
	defer span.End()

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
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

		can, err := handler.authorizer.CanCreateLoRaWANHardwareType(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user not authorized to create LoRaWAN hardware types"))
	}

	hardwareData, err := handler.hardwareTypeManager.CreateLoRaWANHardwareType(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.CreateLoRaWANHardwareTypeResponse_builder{
		HardwareData: hardwareData,
	}.Build())

	return resp, nil
}

func (handler *LoRaWANHandler) GetLoRaWANHardwareType(ctx context.Context, req *connect.Request[iotv1.GetLoRaWANHardwareTypeRequest]) (*connect.Response[iotv1.GetLoRaWANHardwareTypeResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetLoRaWANHardwareType")
	defer span.End()

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
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

		can, err := handler.authorizer.CanReadLoRaWANHardwareType(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user not authorized to read LoRaWAN hardware types"))
	}

	hardwareData, err := handler.hardwareTypeManager.GetLoRaWANHardwareType(ctx, req.Msg.GetHardwareTypeId())
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.GetLoRaWANHardwareTypeResponse_builder{
		HardwareData: hardwareData,
	}.Build())

	return resp, nil
}

func (handler *LoRaWANHandler) ListLoRaWANHardwareTypes(ctx context.Context, req *connect.Request[iotv1.ListLoRaWANHardwareTypesRequest]) (*connect.Response[iotv1.ListLoRaWANHardwareTypesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListLoRaWANHardwareTypes")
	defer span.End()

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
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

		can, err := handler.authorizer.CanReadLoRaWANHardwareType(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user not authorized to read LoRaWAN hardware types"))
	}

	hardwareTypes, err := handler.hardwareTypeManager.ListLoRaWANHardwareTypes(ctx)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.ListLoRaWANHardwareTypesResponse_builder{
		HardwareTypes: hardwareTypes,
	}.Build())

	return resp, nil
}

func (handler *LoRaWANHandler) UpdateLoRaWANHardwareType(ctx context.Context, req *connect.Request[iotv1.UpdateLoRaWANHardwareTypeRequest]) (*connect.Response[iotv1.UpdateLoRaWANHardwareTypeResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "UpdateLoRaWANHardwareType")
	defer span.End()

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
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

		can, err := handler.authorizer.CanUpdateLoRaWANHardwareType(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user not authorized to update LoRaWAN hardware types"))
	}

	_, err := handler.hardwareTypeManager.UpdateLoRaWANHardwareType(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.UpdateLoRaWANHardwareTypeResponse_builder{}.Build())

	return resp, nil
}

func (handler *LoRaWANHandler) DeleteLoRaWANHardwareType(ctx context.Context, req *connect.Request[iotv1.DeleteLoRaWANHardwareTypeRequest]) (*connect.Response[iotv1.DeleteLoRaWANHardwareTypeResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "DeleteLoRaWANHardwareType")
	defer span.End()

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
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

		can, err := handler.authorizer.CanDeleteLoRaWANHardwareType(ctx, userId, organization)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user not authorized to delete LoRaWAN hardware types"))
	}

	err := handler.hardwareTypeManager.DeleteLoRaWANHardwareType(ctx, req.Msg.GetHardwareTypeId())
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.DeleteLoRaWANHardwareTypeResponse_builder{}.Build())

	return resp, nil
}
