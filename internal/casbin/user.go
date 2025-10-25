package casbin

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// UserEnforcer manages authorization for user self-access operations.
type UserEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewUserEnforcer creates a new user enforcer instance.
func NewUserEnforcer(enforcer *casbin.Enforcer) *UserEnforcer {
	return &UserEnforcer{
		enforcer: enforcer,
	}
}

// CanReadSelf checks if a user has permission to read their own user data.
func (e *UserEnforcer) CanReadSelf(ctx context.Context, userId, targetUserId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanReadSelf")
	defer span.End()

	// Self-access check
	return userId == targetUserId, nil
}

// CanUpdateSelf checks if a user has permission to update their own user data.
func (e *UserEnforcer) CanUpdateSelf(ctx context.Context, userId, targetUserId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanUpdateSelf")
	defer span.End()

	// Self-access check
	return userId == targetUserId, nil
}
