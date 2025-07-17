package casbin

import (
	"context"

	"github.com/ponix-dev/ponix/internal/telemetry"
)

// CanAccessOrganization checks if a user can perform an action on an organization
func (e *Enforcer) CanAccessOrganization(ctx context.Context, userId string, action string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanAccessOrganization")
	defer span.End()
	// Format: subject, object, action, organization
	return e.casbin.Enforce(userId, "organization", action, organizationId)
}
