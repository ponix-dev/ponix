package domain

import (
	"context"
	"time"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationStorer interface {
	CreateOrganization(ctx context.Context, organization *organizationv1.Organization) error
	GetOrganization(ctx context.Context, organizationID string) (*organizationv1.Organization, error)
}

type OrganizationManager struct {
	organizationStore OrganizationStorer
	stringId          StringId
	validate          Validate
}

func NewOrganizationManager(os OrganizationStorer, stringId StringId, validate Validate) *OrganizationManager {
	return &OrganizationManager{
		organizationStore: os,
		stringId:          stringId,
		validate:          validate,
	}
}

func (mgr *OrganizationManager) CreateOrganization(ctx context.Context, createReq *organizationv1.CreateOrganizationRequest) (*organizationv1.Organization, error) {
	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	organizationID := mgr.stringId()

	now := timestamppb.New(time.Now().UTC())

	organization := &organizationv1.Organization{
		Id:        organizationID,
		Name:      createReq.GetName(),
		Status:    organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = mgr.organizationStore.CreateOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (mgr *OrganizationManager) GetOrganization(ctx context.Context, organizationReq *organizationv1.OrganizationRequest) (*organizationv1.OrganizationResponse, error) {
	err := mgr.validate(organizationReq)
	if err != nil {
		return nil, err
	}

	organization, err := mgr.organizationStore.GetOrganization(ctx, organizationReq.GetOrganizationId())
	if err != nil {
		return nil, err
	}

	return &organizationv1.OrganizationResponse{
		Organization: organization,
	}, nil
}
