package domain

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// OrganizationRole represents the role a user has within an organization.
type OrganizationRole string

const (
	// OrganizationRoleAdmin grants full access to manage the organization and its resources.
	OrganizationRoleAdmin OrganizationRole = "admin"
	// OrganizationRoleMember grants read and update access to resources.
	OrganizationRoleMember OrganizationRole = "member"
	// OrganizationRoleViewer grants read-only access to resources.
	OrganizationRoleViewer OrganizationRole = "viewer"
)

// UserOrganizationStorer defines the persistence operations for user-organization relationships.
type UserOrganizationStorer interface {
	AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error
	UpdateUserRole(ctx context.Context, userId, organizationId, role string) error
	RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error
}

// UserAuther defines the authorization operations for user-organization relationships.
type UserAuther interface {
	AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error
	UpdateUserRole(ctx context.Context, userId, organizationId, role string) error
	RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error
}

// UserOrganizationManager orchestrates user-organization relationship business logic.
type UserOrganizationManager struct {
	userOrgStore UserOrganizationStorer
	userAuther   UserAuther
	validate     Validate
}

// NewUserOrganizationManager creates a new instance of UserOrganizationManager with the provided dependencies.
func NewUserOrganizationManager(userOrgStore UserOrganizationStorer, userAuther UserAuther, validate Validate) *UserOrganizationManager {
	return &UserOrganizationManager{
		userOrgStore: userOrgStore,
		userAuther:   userAuther,
		validate:     validate,
	}
}

// AddDefaultAdminUser adds the user from context as an admin to the specified organization.
// This is typically called when a new organization is created.
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

// AddOrganizationUser adds a user to an organization with the specified role and updates authorization policies.
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

// UpdateUserRole changes a user's role within an organization and updates authorization policies accordingly.
func (mgr *UserOrganizationManager) UpdateUserRole(ctx context.Context, userId, organizationId, role string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "UpdateUserRole")
	defer span.End()

	err := mgr.userOrgStore.UpdateUserRole(ctx, userId, organizationId, role)
	if err != nil {
		return err
	}

	err = mgr.userAuther.UpdateUserRole(ctx, userId, organizationId, role)
	if err != nil {
		return err
	}

	return nil
}

// RemoveUserFromOrganization removes a user from an organization and revokes their authorization policies.
func (mgr *UserOrganizationManager) RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "RemoveUserFromOrganization")
	defer span.End()

	err := mgr.userOrgStore.RemoveUserFromOrganization(ctx, userId, organizationId)
	if err != nil {
		return err
	}

	err = mgr.userAuther.RemoveUserFromOrganization(ctx, userId, organizationId)
	if err != nil {
		return err
	}

	return nil
}
