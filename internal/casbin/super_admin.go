package casbin

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// SuperAdminEnforcer manages global super admin authorization privileges.
type SuperAdminEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewSuperAdminEnforcer creates a new super admin enforcer instance.
func NewSuperAdminEnforcer(enforcer *casbin.Enforcer) *SuperAdminEnforcer {
	return &SuperAdminEnforcer{
		enforcer: enforcer,
	}
}

// AddSuperAdmin grants a user global super admin privileges across all organizations.
func (e *SuperAdminEnforcer) AddSuperAdmin(ctx context.Context, userId string) error {
	_, span := telemetry.Tracer().Start(ctx, "AddSuperAdmin")
	defer span.End()

	_, err := e.enforcer.AddRoleForUser(userId, "super_admin")
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add super admin role: %w", err)
	}

	return e.enforcer.SavePolicy()
}

// IsSuperAdmin checks if a user has global super admin privileges.
func (e *SuperAdminEnforcer) IsSuperAdmin(user string) (bool, error) {
	_, span := telemetry.Tracer().Start(context.Background(), "IsSuperAdmin")
	defer span.End()

	roles, err := e.enforcer.GetRolesForUser(user)
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
