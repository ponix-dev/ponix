package domain

import (
	"context"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type LoRaWANHardwareTypeStorer interface {
	GetLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) (*iotv1.LoRaWANHardwareData, error)
	AddLoRaWANHardwareType(ctx context.Context, hardwareData *iotv1.LoRaWANHardwareData) error
	UpdateLoRaWANHardwareType(ctx context.Context, hardwareData *iotv1.LoRaWANHardwareData) error
	ListLoRaWANHardwareTypes(ctx context.Context) ([]*iotv1.LoRaWANHardwareData, error)
	DeleteLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) error
}

type LoRaWANManager struct {
	hardwareTypeStore LoRaWANHardwareTypeStorer
	stringId          StringId
	validate          Validate
}

func NewLoRaWANManager(store LoRaWANHardwareTypeStorer, stringId StringId, validate Validate) *LoRaWANManager {
	return &LoRaWANManager{
		hardwareTypeStore: store,
		stringId:          stringId,
		validate:          validate,
	}
}

// ===== LoRaWAN Hardware Type Management =====

// CreateLoRaWANHardwareType creates a new LoRaWAN hardware type from a request
func (mgr *LoRaWANManager) CreateLoRaWANHardwareType(ctx context.Context, createReq *iotv1.CreateLoRaWANHardwareTypeRequest) (*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateLoRaWANHardwareType")
	defer span.End()

	hardwareTypeID := mgr.stringId()

	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	// Build complete hardware data using builder pattern
	hardwareData := iotv1.LoRaWANHardwareData_builder{
		HardwareTypeId:  hardwareTypeID,
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

// GetLoRaWANHardwareType retrieves a LoRaWAN hardware type by ID
func (mgr *LoRaWANManager) GetLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) (*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetLoRaWANHardwareType")
	defer span.End()

	return mgr.hardwareTypeStore.GetLoRaWANHardwareType(ctx, hardwareTypeID)
}

// ListLoRaWANHardwareTypes lists all available LoRaWAN hardware types
func (mgr *LoRaWANManager) ListLoRaWANHardwareTypes(ctx context.Context) ([]*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListLoRaWANHardwareTypes")
	defer span.End()

	return mgr.hardwareTypeStore.ListLoRaWANHardwareTypes(ctx)
}

// UpdateLoRaWANHardwareType updates an existing LoRaWAN hardware type
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

// DeleteLoRaWANHardwareType soft deletes a LoRaWAN hardware type by ID
func (mgr *LoRaWANManager) DeleteLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "DeleteLoRaWANHardwareType")
	defer span.End()

	return mgr.hardwareTypeStore.DeleteLoRaWANHardwareType(ctx, hardwareTypeID)
}