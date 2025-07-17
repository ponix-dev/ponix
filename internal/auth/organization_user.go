package auth

import (
	"context"
	"fmt"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// AddUserToOrganization assigns a user to a role within an organization
func (e *Enforcer) AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error {
	_, span := telemetry.Tracer().Start(ctx, "addUserToOrganization")
	defer span.End()

	// Create organization-specific role name
	orgRole := fmt.Sprintf("org_%s:%s", orgUser.Role, orgUser.OrganizationId)

	// Assign user to organization-specific role
	_, err := e.casbin.AddRoleForUser(orgUser.UserId, orgRole)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add user to organization: %w", err)
	}

	// Add organization-specific policies based on role
	err = e.addOrgSpecificPolicies(domain.OrganizationRole(orgUser.Role), orgUser.OrganizationId)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add organization policies: %w", err)
	}

	return e.casbin.SavePolicy()
}

// addOrgSpecificPolicies adds policies for a specific organization and role
func (e *Enforcer) addOrgSpecificPolicies(role domain.OrganizationRole, organization string) error {
	_, span := telemetry.Tracer().Start(context.Background(), "addOrgSpecificPolicies")
	defer span.End()

	orgRole := fmt.Sprintf("org_%s:%s", role, organization)

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

	// Add all policies for this organization role
	for _, policy := range policies {
		_, err := e.casbin.AddPolicy(policy)
		if err != nil {
			return stacktrace.NewStackTraceErrorf("failed to add policy %v: %w", policy, err)
		}
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

// RemoveUserFromOrganization removes a user's role within an organization
func (e *Enforcer) RemoveUserFromOrganization(ctx context.Context, user, role, organization string) error {
	_, span := telemetry.Tracer().Start(ctx, "RemoveUserFromOrganization")
	defer span.End()

	orgRole := fmt.Sprintf("org_%s:%s", role, organization)
	_, err := e.casbin.DeleteRoleForUser(user, orgRole)
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to remove user from organization: %w", err)
	}

	return e.casbin.SavePolicy()
}
