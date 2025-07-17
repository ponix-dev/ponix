package auth

import (
	"context"
	"fmt"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// AddUserToOrganization assigns a user to a role within an organization (idempotent)
func (e *Enforcer) AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error {
	_, span := telemetry.Tracer().Start(ctx, "addUserToOrganization")
	defer span.End()

	// Create organization-specific role name
	orgRole := fmt.Sprintf("org_%s:%s", orgUser.Role, orgUser.OrganizationId)

	// Remove any existing roles for this user in this organization first
	existingRoles, err := e.casbin.GetRolesForUser(orgUser.UserId)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to get user roles: %w", err)
	}

	for _, role := range existingRoles {
		if len(role) > len(orgUser.OrganizationId) && role[len(role)-len(orgUser.OrganizationId):] == orgUser.OrganizationId {
			_, _ = e.casbin.DeleteRoleForUser(orgUser.UserId, role) // Ignore errors - might not exist
		}
	}

	// Assign user to organization-specific role (idempotent - Casbin handles duplicates)
	_, err = e.casbin.AddRoleForUser(orgUser.UserId, orgRole)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add user to organization: %w", err)
	}

	// Add organization-specific policies based on role (idempotent)
	err = e.addOrgSpecificPolicies(domain.OrganizationRole(orgUser.Role), orgUser.OrganizationId)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add organization policies: %w", err)
	}

	return e.casbin.SavePolicy()
}

// addOrgSpecificPolicies adds policies for a specific organization and role (idempotent)
func (e *Enforcer) addOrgSpecificPolicies(role domain.OrganizationRole, organization string) error {
	_, span := telemetry.Tracer().Start(context.Background(), "addOrgSpecificPolicies")
	defer span.End()

	orgRole := fmt.Sprintf("org_%s:%s", role, organization)

	// Remove existing policies for this org role first to ensure clean state
	e.casbin.RemoveFilteredPolicy(0, orgRole)

	var policies [][]string
	switch role {
	case domain.OrganizationRoleAdmin:
		policies = [][]string{
			{orgRole, "end_device", "create", organization},
			{orgRole, "end_device", "read", organization},
			{orgRole, "end_device", "update", organization},
			{orgRole, "end_device", "delete", organization},
			{orgRole, "organization", "read", organization},
			{orgRole, "organization", "update", organization},
			{orgRole, "user", "create", organization},
			{orgRole, "user", "update", organization},
			{orgRole, "user", "delete", organization},
		}
	case domain.OrganizationRoleMember:
		policies = [][]string{
			{orgRole, "end_device", "read", organization},
			{orgRole, "end_device", "update", organization},
			{orgRole, "organization", "read", organization},
		}
	case domain.OrganizationRoleViewer:
		policies = [][]string{
			{orgRole, "end_device", "read", organization},
			{orgRole, "organization", "read", organization},
		}
	default:
		return stacktrace.NewStackTraceErrorf("unknown role: %s", role)
	}

	// Add all policies for this organization role (idempotent - duplicates are handled by Casbin)
	for _, policy := range policies {
		_, _ = e.casbin.AddPolicy(policy) // Ignore errors - might already exist
	}

	return nil
}

// CanManageUsers checks if a user can manage other users within an organization
func (e *Enforcer) CanManageUsers(ctx context.Context, user string, action string, organization string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanManageUsers")
	defer span.End()

	// Format: subject, object, action, organization
	return e.casbin.Enforce(user, "user", action, organization)
}

// UpdateUserRole updates a user's role within an organization
func (e *Enforcer) UpdateUserRole(ctx context.Context, userId, organizationId, role string) error {
	_, span := telemetry.Tracer().Start(ctx, "UpdateUserRole")
	defer span.End()

	// First, remove all existing organization roles for this user in this organization
	roles, err := e.casbin.GetRolesForUser(userId)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to get user roles: %w", err)
	}

	orgPrefix := fmt.Sprintf("org_%s:", organizationId)
	for _, existingRole := range roles {
		if len(existingRole) > len(orgPrefix) && existingRole[len(existingRole)-len(organizationId):] == organizationId {
			_, err := e.casbin.DeleteRoleForUser(userId, existingRole)
			if err != nil {
				return stacktrace.NewStackTraceErrorf("failed to remove existing role: %w", err)
			}
		}
	}

	// Add the new role
	newOrgRole := fmt.Sprintf("org_%s:%s", role, organizationId)
	_, err = e.casbin.AddRoleForUser(userId, newOrgRole)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add new role: %w", err)
	}

	// Add organization-specific policies for the new role
	err = e.addOrgSpecificPolicies(domain.OrganizationRole(role), organizationId)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add organization policies: %w", err)
	}

	return e.casbin.SavePolicy()
}

// RemoveUserFromOrganization removes all user roles within an organization (idempotent)
func (e *Enforcer) RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error {
	_, span := telemetry.Tracer().Start(ctx, "RemoveUserFromOrganization")
	defer span.End()

	// Get all roles for the user
	roles, err := e.casbin.GetRolesForUser(userId)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to get user roles: %w", err)
	}

	// Remove all organization-specific roles for this user in this organization (idempotent)
	for _, role := range roles {
		if len(role) > len(organizationId) && role[len(role)-len(organizationId):] == organizationId {
			_, _ = e.casbin.DeleteRoleForUser(userId, role) // Ignore errors - might not exist
		}
	}

	return e.casbin.SavePolicy()
}
