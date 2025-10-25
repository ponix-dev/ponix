package casbin

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// OrganizationAccessEnforcer manages authorization for organization-level operations.
type OrganizationAccessEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewOrganizationAccessEnforcer creates a new organization access enforcer instance.
func NewOrganizationAccessEnforcer(enforcer *casbin.Enforcer) *OrganizationAccessEnforcer {
	return &OrganizationAccessEnforcer{
		enforcer: enforcer,
	}
}

// CanCreateOrganization checks if a user has permission to create new organizations.
func (e *OrganizationAccessEnforcer) CanCreateOrganization(ctx context.Context, userId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanCreateOrganization")
	defer span.End()
	// Organization creation is typically a system-level permission, not org-specific
	return e.enforcer.Enforce(userId, "organization", "create", "*")
}

// CanReadOrganization checks if a user has permission to read organization details.
func (e *OrganizationAccessEnforcer) CanReadOrganization(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanReadOrganization")
	defer span.End()
	return e.enforcer.Enforce(userId, "organization", "read", organizationId)
}

// CanUpdateOrganization checks if a user has permission to update organization details.
func (e *OrganizationAccessEnforcer) CanUpdateOrganization(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanUpdateOrganization")
	defer span.End()
	return e.enforcer.Enforce(userId, "organization", "update", organizationId)
}

// CanDeleteOrganization checks if a user has permission to delete an organization.
func (e *OrganizationAccessEnforcer) CanDeleteOrganization(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanDeleteOrganization")
	defer span.End()
	return e.enforcer.Enforce(userId, "organization", "delete", organizationId)
}
