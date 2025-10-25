package domain

import (
	"context"
	"time"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DefaultAdminer handles adding the default admin user to newly created organizations.
type DefaultAdminer interface {
	AddDefaultAdminUser(ctx context.Context, organizationId string) error
}

// OrganizationStorer defines the persistence operations for organizations.
type OrganizationStorer interface {
	CreateOrganization(ctx context.Context, organization *organizationv1.Organization) error
	GetOrganization(ctx context.Context, organizationId string) (*organizationv1.Organization, error)
	GetUserOrganizationsWithDetails(ctx context.Context, userId string) ([]*organizationv1.Organization, error)
}

// OrganizationManager orchestrates organization-related business logic.
type OrganizationManager struct {
	organizationStore OrganizationStorer
	stringId          StringId
	validate          Validate
	defaultAdminer    DefaultAdminer
}

// NewOrganizationManager creates a new instance of OrganizationManager with the provided dependencies.
func NewOrganizationManager(os OrganizationStorer, stringId StringId, validate Validate, defaultAdminer DefaultAdminer) *OrganizationManager {
	return &OrganizationManager{
		organizationStore: os,
		stringId:          stringId,
		validate:          validate,
		defaultAdminer:    defaultAdminer,
	}
}

// CreateOrganization creates a new organization with a unique ID and automatically adds
// the requesting user as an admin of the organization.
func (mgr *OrganizationManager) CreateOrganization(ctx context.Context, createReq *organizationv1.CreateOrganizationRequest) (*organizationv1.Organization, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganization")
	defer span.End()

	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	organizationId := mgr.stringId()

	now := timestamppb.New(time.Now().UTC())

	organization := &organizationv1.Organization{
		Id:        organizationId,
		Name:      createReq.GetName(),
		Status:    organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = mgr.organizationStore.CreateOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}

	err = mgr.defaultAdminer.AddDefaultAdminUser(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// GetOrganization retrieves an organization by its ID.
func (mgr *OrganizationManager) GetOrganization(ctx context.Context, organizationReq *organizationv1.GetOrganizationRequest) (*organizationv1.Organization, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetOrganization")
	defer span.End()

	err := mgr.validate(organizationReq)
	if err != nil {
		return nil, err
	}

	organization, err := mgr.organizationStore.GetOrganization(ctx, organizationReq.GetOrganizationId())
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// GetUserOrganizations retrieves all organizations that a user belongs to.
func (mgr *OrganizationManager) GetUserOrganizations(ctx context.Context, userId string) ([]*organizationv1.Organization, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetUserOrganizations")
	defer span.End()

	organizations, err := mgr.organizationStore.GetUserOrganizationsWithDetails(ctx, userId)
	if err != nil {
		return nil, err
	}

	return organizations, nil
}
