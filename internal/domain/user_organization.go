package domain

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

type OrganizationRole string

const (
	OrganizationRoleAdmin  OrganizationRole = "admin"
	OrganizationRoleMember OrganizationRole = "member"
	OrganizationRoleViewer OrganizationRole = "viewer"
)

type UserOrganizationStorer interface {
	AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error
}

type UserAuther interface {
	AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error
}

type UserOrganizationManager struct {
	userOrgStore UserOrganizationStorer
	userAuther   UserAuther
	validate     Validate
}

func NewUserOrganizationManager(userOrgStore UserOrganizationStorer, userAuther UserAuther, validate Validate) *UserOrganizationManager {
	return &UserOrganizationManager{
		userOrgStore: userOrgStore,
		userAuther:   userAuther,
		validate:     validate,
	}
}

func (mgr *UserOrganizationManager) AddDefaultAdminUser(ctx context.Context, organizationId string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "AddDefaultAdminUser")
	defer span.End()

	userId, ok := GetUserFromContext(ctx)
	if !ok {
		return stacktrace.NewStackTraceError(ErrMissingUserInContext)
	}

	orgUser := &organizationv1.OrganizationUser{
		UserId:         userId,
		OrganizationId: organizationId,
		Role:           string(OrganizationRoleAdmin),
	}

	err := mgr.userOrgStore.AddUserToOrganization(ctx, orgUser)
	if err != nil {
		return err
	}

	err = mgr.userAuther.AddUserToOrganization(ctx, orgUser)
	if err != nil {
		return err
	}

	return nil
}

func (mgr *UserOrganizationManager) AddOrganizationUser(ctx context.Context, orgUser *organizationv1.OrganizationUser) error {
	ctx, span := telemetry.Tracer().Start(ctx, "AddOrganizationUser")
	defer span.End()

	err := mgr.userOrgStore.AddUserToOrganization(ctx, orgUser)
	if err != nil {
		return err
	}

	err = mgr.userAuther.AddUserToOrganization(ctx, orgUser)
	if err != nil {
		return err
	}
	return nil
}
