package connectrpc

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type OrganizationHandler struct{}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{}
}

func (handler *OrganizationHandler) CreateOrganization(ctx context.Context, req *connect.Request[organizationv1.CreateOrganizationRequest]) (*connect.Response[organizationv1.CreateOrganizationResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganization")
	defer span.End()

	return nil, nil
}

func (handler *OrganizationHandler) Organization(ctx context.Context, req *connect.Request[organizationv1.OrganizationRequest]) (*connect.Response[organizationv1.OrganizationResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "Organization")
	defer span.End()

	return nil, nil
}

func (handler *OrganizationHandler) OrganizationUsers(ctx context.Context, req *connect.Request[organizationv1.OrganizationUsersRequest]) (*connect.Response[organizationv1.OrganizationUsersResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "OrganizationUsers")
	defer span.End()

	return nil, nil
}

func (handler *OrganizationHandler) OrganizationSystems(ctx context.Context, req *connect.Request[organizationv1.OrganizationSystemsRequest]) (*connect.Response[organizationv1.OrganizationSystemsResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "OrganizationSystems")
	defer span.End()

	return nil, nil
}
