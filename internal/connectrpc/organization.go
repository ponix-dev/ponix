package connectrpc

import (
	"context"
	"fmt"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// OrganizationManager handles organization business operations.
type OrganizationManager interface {
	CreateOrganization(ctx context.Context, createReq *organizationv1.CreateOrganizationRequest) (*organizationv1.Organization, error)
	GetOrganization(ctx context.Context, organizationReq *organizationv1.GetOrganizationRequest) (*organizationv1.Organization, error)
	GetUserOrganizations(ctx context.Context, userId string) ([]*organizationv1.Organization, error)
}

// OrganizationAuthorizer checks permissions for organization operations.
type OrganizationAuthorizer interface {
	CanCreateOrganization(ctx context.Context, userId string) (bool, error)
	CanReadOrganization(ctx context.Context, userId string, organizationId string) (bool, error)
	CanUpdateOrganization(ctx context.Context, userId string, organizationId string) (bool, error)
	CanDeleteOrganization(ctx context.Context, userId string, organizationId string) (bool, error)
}

// OrganizationHandler implements Connect RPC handlers for organization operations.
type OrganizationHandler struct {
	organizationManager OrganizationManager
	authorizer          OrganizationAuthorizer
}

// NewOrganizationHandler creates a new OrganizationHandler with the provided dependencies.
func NewOrganizationHandler(organizationManager OrganizationManager, authorizer OrganizationAuthorizer) *OrganizationHandler {
	return &OrganizationHandler{
		organizationManager: organizationManager,
		authorizer:          authorizer,
	}
}

// CreateOrganization handles RPC requests to create a new organization.
// Requires super admin privileges or create organization permission.
func (handler *OrganizationHandler) CreateOrganization(ctx context.Context, req *connect.Request[organizationv1.CreateOrganizationRequest]) (*connect.Response[organizationv1.CreateOrganizationResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganization")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	// Authorization
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanCreateOrganization(ctx, callingUserId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}

		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to create organizations", callingUserId))
	}

	organization, err := handler.organizationManager.CreateOrganization(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	response := &organizationv1.CreateOrganizationResponse{
		OrganizationId: organization.GetId(),
		Name:           organization.GetName(),
		Status:         organization.GetStatus(),
		CreatedAt:      organization.GetCreatedAt(),
	}

	return connect.NewResponse(response), nil
}

// GetOrganization handles RPC requests to retrieve an organization by ID.
// Requires super admin privileges or read access to the organization.
func (handler *OrganizationHandler) GetOrganization(ctx context.Context, req *connect.Request[organizationv1.GetOrganizationRequest]) (*connect.Response[organizationv1.GetOrganizationResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetOrganization")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadOrganization(ctx, callingUserId, req.Msg.GetOrganizationId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to read organization %s", callingUserId, req.Msg.GetOrganizationId()))
	}

	organization, err := handler.organizationManager.GetOrganization(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	response := &organizationv1.GetOrganizationResponse{
		Organization: organization,
	}

	return connect.NewResponse(response), nil
}

// GetUserOrganizations handles RPC requests to retrieve all organizations a user belongs to.
// Users can only retrieve their own organizations unless they are a super admin.
func (handler *OrganizationHandler) GetUserOrganizations(ctx context.Context, req *connect.Request[organizationv1.GetUserOrganizationsRequest]) (*connect.Response[organizationv1.GetUserOrganizationsResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetUserOrganizations")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		// Users can only get their own organizations
		allowed = callingUserId == req.Msg.GetUserId()
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to read organizations for user %s", callingUserId, req.Msg.GetUserId()))
	}

	organizations, err := handler.organizationManager.GetUserOrganizations(ctx, req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	response := &organizationv1.GetUserOrganizationsResponse{
		Organizations: organizations,
	}

	return connect.NewResponse(response), nil
}
