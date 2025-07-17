package casbin

import (
	"context"

	"github.com/ponix-dev/ponix/internal/telemetry"
)

// CanAccessEndDevice checks if a user can perform an action on an end device within an organization
func (e *Enforcer) CanAccessEndDevice(ctx context.Context, userId string, action string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanAccessEndDevice")
	defer span.End()

	// Format: subject, object, action, organization
	return e.casbin.Enforce(userId, "end_device", action, organizationId)
}
