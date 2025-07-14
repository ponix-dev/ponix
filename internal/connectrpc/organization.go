package connectrpc

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type OrganizationManager interface {
	CreateOrganization(ctx context.Context, createReq *organizationv1.CreateOrganizationRequest) (*organizationv1.Organization, error)
	GetOrganization(ctx context.Context, organizationReq *organizationv1.OrganizationRequest) (*organizationv1.OrganizationResponse, error)
}

type OrganizationHandler struct {
	organizationManager OrganizationManager
}

func NewOrganizationHandler(organizationManager OrganizationManager) *OrganizationHandler {
	return &OrganizationHandler{
		organizationManager: organizationManager,
	}
}

func (handler *OrganizationHandler) CreateOrganization(ctx context.Context, req *connect.Request[organizationv1.CreateOrganizationRequest]) (*connect.Response[organizationv1.CreateOrganizationResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganization")
	defer span.End()

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

func (handler *OrganizationHandler) Organization(ctx context.Context, req *connect.Request[organizationv1.OrganizationRequest]) (*connect.Response[organizationv1.OrganizationResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "Organization")
	defer span.End()

	response, err := handler.organizationManager.GetOrganization(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(response), nil
}
