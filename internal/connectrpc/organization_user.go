package connectrpc

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type OrganizationUserManager interface {
	AddOrganizationUser(ctx context.Context, orgUser *organizationv1.OrganizationUser) error
	UpdateUserRole(ctx context.Context, userId, organizationId, role string) error
	RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error
}

type OrganizationUserHandler struct {
	organizationUserManager OrganizationUserManager
}

func NewOrganizationUserHandler(organizationUserManager OrganizationUserManager) *OrganizationUserHandler {
	return &OrganizationUserHandler{
		organizationUserManager: organizationUserManager,
	}
}

func (handler *OrganizationUserHandler) CreateOrganizationUser(ctx context.Context, req *connect.Request[organizationv1.CreateOrganizationUserRequest]) (*connect.Response[organizationv1.CreateOrganizationUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganizationUser")
	defer span.End()

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

	// Remove the user from organization
	err := handler.organizationUserManager.RemoveUserFromOrganization(ctx, req.Msg.UserId, req.Msg.OrganizationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &organizationv1.RemoveOrganizationUserResponse{}
	return connect.NewResponse(response), nil
}
