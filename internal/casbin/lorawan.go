package casbin

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// LoRaWANEnforcer handles LoRaWAN hardware type authorization
type LoRaWANEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewLoRaWANEnforcer creates a new LoRaWAN enforcer
func NewLoRaWANEnforcer(enforcer *casbin.Enforcer) *LoRaWANEnforcer {
	return &LoRaWANEnforcer{
		enforcer: enforcer,
	}
}

// CanCreateLoRaWANHardwareType checks if a user can create LoRaWAN hardware types
func (e *LoRaWANEnforcer) CanCreateLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanCreateLoRaWANHardwareType")
	defer span.End()

	return e.enforcer.Enforce(userId, "lorawan_hardware_type", "create", organizationId)
}

// CanReadLoRaWANHardwareType checks if a user can read LoRaWAN hardware types
func (e *LoRaWANEnforcer) CanReadLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanReadLoRaWANHardwareType")
	defer span.End()

	return e.enforcer.Enforce(userId, "lorawan_hardware_type", "read", organizationId)
}

// CanUpdateLoRaWANHardwareType checks if a user can update LoRaWAN hardware types
func (e *LoRaWANEnforcer) CanUpdateLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanUpdateLoRaWANHardwareType")
	defer span.End()

	return e.enforcer.Enforce(userId, "lorawan_hardware_type", "update", organizationId)
}

// CanDeleteLoRaWANHardwareType checks if a user can delete LoRaWAN hardware types
func (e *LoRaWANEnforcer) CanDeleteLoRaWANHardwareType(ctx context.Context, userId string, organizationId string) (bool, error) {
	_, span := telemetry.Tracer().Start(ctx, "CanDeleteLoRaWANHardwareType")
	defer span.End()

	return e.enforcer.Enforce(userId, "lorawan_hardware_type", "delete", organizationId)
}