package casbin

import (
	"context"

	"github.com/ponix-dev/ponix/internal/telemetry"
)

// CanManageSelf checks if a user can manage their own user data
func (e *Enforcer) CanManageSelf(ctx context.Context, userId, action, targetUserId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanManageSelf")
	defer span.End()

	// Self-access check
	if userId == targetUserId {
		return true, nil
	}

	return false, nil
}
