package domain

import (
	"context"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

// LoRaWANHardwareTypeStorer defines the persistence operations for LoRaWAN hardware types.
type LoRaWANHardwareTypeStorer interface {
	GetLoRaWANHardwareType(ctx context.Context, hardwareType string) (*iotv1.LoRaWANHardwareData, error)
	AddLoRaWANHardwareType(ctx context.Context, hardwareData *iotv1.LoRaWANHardwareData) error
	UpdateLoRaWANHardwareType(ctx context.Context, hardwareData *iotv1.LoRaWANHardwareData) error
	ListLoRaWANHardwareTypes(ctx context.Context) ([]*iotv1.LoRaWANHardwareData, error)
	DeleteLoRaWANHardwareType(ctx context.Context, hardwareType string) error
}

// LoRaWANManager orchestrates LoRaWAN hardware type catalog business logic.
type LoRaWANManager struct {
	hardwareTypeStore LoRaWANHardwareTypeStorer
	stringId          StringId
	validate          Validate
}

// NewLoRaWANManager creates a new instance of LoRaWANManager with the provided dependencies.
func NewLoRaWANManager(store LoRaWANHardwareTypeStorer, stringId StringId, validate Validate) *LoRaWANManager {
	return &LoRaWANManager{
		hardwareTypeStore: store,
		stringId:          stringId,
		validate:          validate,
	}
}

// CreateLoRaWANHardwareType creates a new LoRaWAN hardware type with a unique ID.
func (mgr *LoRaWANManager) CreateLoRaWANHardwareType(ctx context.Context, createReq *iotv1.CreateLoRaWANHardwareTypeRequest) (*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateLoRaWANHardwareType")
	defer span.End()

	hardwareType := mgr.stringId()

	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	// Build complete hardware data using builder pattern
	hardwareData := iotv1.LoRaWANHardwareData_builder{
		HardwareTypeId:  hardwareType,
		Name:            createReq.GetName(),
		Description:     createReq.GetDescription(),
		Manufacturer:    createReq.GetManufacturer(),
		Model:           createReq.GetModel(),
		FirmwareVersion: createReq.GetFirmwareVersion(),
		HardwareVersion: createReq.GetHardwareVersion(),
		Profile:         createReq.GetProfile(),
		LorawanVersion:  createReq.GetLorawanVersion(),
	}.Build()

	// Store in database
	err = mgr.hardwareTypeStore.AddLoRaWANHardwareType(ctx, hardwareData)
	if err != nil {
		return nil, err
	}

	return hardwareData, nil
}

// GetLoRaWANHardwareType retrieves a LoRaWAN hardware type by its ID.
func (mgr *LoRaWANManager) GetLoRaWANHardwareType(ctx context.Context, hardwareType string) (*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetLoRaWANHardwareType")
	defer span.End()

	return mgr.hardwareTypeStore.GetLoRaWANHardwareType(ctx, hardwareType)
}

// ListLoRaWANHardwareTypes retrieves all available LoRaWAN hardware types.
func (mgr *LoRaWANManager) ListLoRaWANHardwareTypes(ctx context.Context) ([]*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListLoRaWANHardwareTypes")
	defer span.End()

	return mgr.hardwareTypeStore.ListLoRaWANHardwareTypes(ctx)
}

// UpdateLoRaWANHardwareType updates an existing LoRaWAN hardware type's metadata.
func (mgr *LoRaWANManager) UpdateLoRaWANHardwareType(ctx context.Context, updateReq *iotv1.UpdateLoRaWANHardwareTypeRequest) (*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "UpdateLoRaWANHardwareType")
	defer span.End()

	err := mgr.validate(updateReq)
	if err != nil {
		return nil, err
	}

	// Use the hardware data from the request directly
	hardwareData := updateReq.GetHardwareData()
	if hardwareData == nil {
		return nil, fmt.Errorf("hardware data is required for update")
	}

	// Update in database
	err = mgr.hardwareTypeStore.UpdateLoRaWANHardwareType(ctx, hardwareData)
	if err != nil {
		return nil, err
	}

	// Fetch the updated hardware data to return
	updatedHardwareData, err := mgr.hardwareTypeStore.GetLoRaWANHardwareType(ctx, hardwareData.GetHardwareTypeId())
	if err != nil {
		return nil, err
	}

	return updatedHardwareData, nil
}

// DeleteLoRaWANHardwareType soft deletes a LoRaWAN hardware type by its ID.
func (mgr *LoRaWANManager) DeleteLoRaWANHardwareType(ctx context.Context, hardwareType string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "DeleteLoRaWANHardwareType")
	defer span.End()

	return mgr.hardwareTypeStore.DeleteLoRaWANHardwareType(ctx, hardwareType)
}
