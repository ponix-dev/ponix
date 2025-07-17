package casbin

import (
	"context"
	"fmt"

	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// AddSuperAdmin assigns a user global super admin privileges
func (e *Enforcer) AddSuperAdmin(ctx context.Context, userId string) error {
	_, span := telemetry.Tracer().Start(ctx, "AddSuperAdmin")
	defer span.End()

	_, err := e.casbin.AddRoleForUser(userId, "super_admin")
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add super admin role: %w", err)
	}

	return e.casbin.SavePolicy()
}

// isSuperAdmin checks if a user has super admin privileges
func (e *Enforcer) IsSuperAdmin(user string) (bool, error) {
	_, span := telemetry.Tracer().Start(context.Background(), "IsSuperAdmin")
	defer span.End()

	roles, err := e.casbin.GetRolesForUser(user)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, role := range roles {
		if role == "super_admin" {
			return true, nil
		}
	}

	return false, nil
}
