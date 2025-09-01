package casbin

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// EndDeviceEnforcer handles end device-specific authorization
type EndDeviceEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewEndDeviceEnforcer creates a new end device enforcer
func NewEndDeviceEnforcer(enforcer *casbin.Enforcer) *EndDeviceEnforcer {
	return &EndDeviceEnforcer{
		enforcer: enforcer,
	}
}

// CanCreateEndDevice checks if a user can create end devices within an organization
func (e *EndDeviceEnforcer) CanCreateEndDevice(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanCreateEndDevice")
	defer span.End()

	return e.enforcer.Enforce(userId, "end_device", "create", organizationId)
}

// CanReadEndDevice checks if a user can read end devices within an organization
func (e *EndDeviceEnforcer) CanReadEndDevice(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanReadEndDevice")
	defer span.End()

	return e.enforcer.Enforce(userId, "end_device", "read", organizationId)
}

// CanUpdateEndDevice checks if a user can update end devices within an organization
func (e *EndDeviceEnforcer) CanUpdateEndDevice(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanUpdateEndDevice")
	defer span.End()

	return e.enforcer.Enforce(userId, "end_device", "update", organizationId)
}

// CanDeleteEndDevice checks if a user can delete end devices within an organization
func (e *EndDeviceEnforcer) CanDeleteEndDevice(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanDeleteEndDevice")
	defer span.End()

	return e.enforcer.Enforce(userId, "end_device", "delete", organizationId)
}