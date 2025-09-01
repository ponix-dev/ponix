package connectrpc

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

type OrganizationUserManager interface {
	AddOrganizationUser(ctx context.Context, orgUser *organizationv1.OrganizationUser) error
	UpdateUserRole(ctx context.Context, userId, organizationId, role string) error
	RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error
}

type OrganizationUserAuthorizer interface {
	CanCreateUsers(ctx context.Context, user string, organization string) (bool, error)
	CanReadUsers(ctx context.Context, user string, organization string) (bool, error)
	CanUpdateUsers(ctx context.Context, user string, organization string) (bool, error)
	CanDeleteUsers(ctx context.Context, user string, organization string) (bool, error)
}

type OrganizationUserHandler struct {
	organizationUserManager OrganizationUserManager
	authorizer              OrganizationUserAuthorizer
}

func NewOrganizationUserHandler(organizationUserManager OrganizationUserManager, authorizer OrganizationUserAuthorizer) *OrganizationUserHandler {
	return &OrganizationUserHandler{
		organizationUserManager: organizationUserManager,
		authorizer:              authorizer,
	}
}

func (handler *OrganizationUserHandler) CreateOrganizationUser(ctx context.Context, req *connect.Request[organizationv1.CreateOrganizationUserRequest]) (*connect.Response[organizationv1.CreateOrganizationUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganizationUser")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, stacktrace.NewStackTraceErrorf("user not authenticated"))
	}

	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanCreateUsers(ctx, callingUserId, req.Msg.OrganizationId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, stacktrace.NewStackTraceErrorf("authorization check failed: %w", err))
		}

		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, stacktrace.NewStackTraceErrorf("user %s not authorized to create users in organization %s", callingUserId, req.Msg.OrganizationId))
	}

	// Create the organization user
	orgUser := &organizationv1.OrganizationUser{
		UserId:         req.Msg.UserId,
		OrganizationId: req.Msg.OrganizationId,
		Role:           req.Msg.Role,
	}

	err := handler.organizationUserManager.AddOrganizationUser(ctx, orgUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &organizationv1.CreateOrganizationUserResponse{}
	return connect.NewResponse(response), nil
}

func (handler *OrganizationUserHandler) UpdateOrganizationUserRole(ctx context.Context, req *connect.Request[organizationv1.UpdateOrganizationUserRoleRequest]) (*connect.Response[organizationv1.UpdateOrganizationUserRoleResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "UpdateOrganizationUserRole")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, stacktrace.NewStackTraceErrorf("user not authenticated"))
	}

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanUpdateUsers(ctx, callingUserId, req.Msg.OrganizationId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, stacktrace.NewStackTraceErrorf("authorization check failed: %w", err))
		}

		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, stacktrace.NewStackTraceErrorf("user %s not authorized to update users in organization %s", callingUserId, req.Msg.OrganizationId))
	}

	// Update the user role
	err := handler.organizationUserManager.UpdateUserRole(ctx, req.Msg.UserId, req.Msg.OrganizationId, req.Msg.Role)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &organizationv1.UpdateOrganizationUserRoleResponse{}
	return connect.NewResponse(response), nil
}

func (handler *OrganizationUserHandler) RemoveOrganizationUser(ctx context.Context, req *connect.Request[organizationv1.RemoveOrganizationUserRequest]) (*connect.Response[organizationv1.RemoveOrganizationUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "RemoveOrganizationUser")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, stacktrace.NewStackTraceErrorf("user not authenticated"))
	}

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanDeleteUsers(ctx, callingUserId, req.Msg.OrganizationId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, stacktrace.NewStackTraceErrorf("authorization check failed: %w", err))
		}

		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, stacktrace.NewStackTraceErrorf("user %s not authorized to delete users in organization %s", callingUserId, req.Msg.OrganizationId))
	}

	// Remove the user from organization
	err := handler.organizationUserManager.RemoveUserFromOrganization(ctx, req.Msg.UserId, req.Msg.OrganizationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &organizationv1.RemoveOrganizationUserResponse{}
	return connect.NewResponse(response), nil
}
