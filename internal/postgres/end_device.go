package postgres

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
	"github.com/rs/xid"
)

// EndDeviceStore handles database operations for end devices and LoRaWAN configurations.
type EndDeviceStore struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

// NewEndDeviceStore creates a new EndDeviceStore instance.
func NewEndDeviceStore(db *sqlc.Queries, pool *pgxpool.Pool) *EndDeviceStore {
	return &EndDeviceStore{
		db:   db,
		pool: pool,
	}
}

// AddEndDevice inserts a new end device and its associated configuration into the database.
// For LoRaWAN devices, this also creates the corresponding LoRaWAN configuration within a transaction.
func (store *EndDeviceStore) AddEndDevice(ctx context.Context, endDevice *iotv1.EndDevice, organizationID string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}
	defer tx.Rollback(ctx)

	txQueries := store.db.WithTx(tx)

	endDeviceParams := sqlc.CreateEndDeviceParams{
		ID:             endDevice.GetId(),
		Name:           endDevice.GetName(),
		Description:    pgtype.Text{String: endDevice.GetDescription(), Valid: endDevice.GetDescription() != ""},
		OrganizationID: organizationID,
		Status:         int32(endDevice.GetStatus()),
		DataType:       int32(endDevice.GetDataType()),
		HardwareType:   int32(endDevice.GetHardwareType()),
	}

	_, err = txQueries.CreateEndDevice(ctx, endDeviceParams)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	switch endDevice.GetHardwareType() {
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
		lorawanConfig := endDevice.GetLorawanConfig()
		if lorawanConfig == nil {
			return stacktrace.NewStackTraceErrorf("LoRaWAN config required for LoRaWAN device type")
		}

		lorawanParams := sqlc.CreateLoRaWANConfigParams{
			ID:               xid.New().String(),
			EndDeviceID:      endDevice.GetId(),
			DeviceEui:        lorawanConfig.GetDeviceEui(),
			ApplicationEui:   lorawanConfig.GetApplicationEui(),
			ApplicationID:    lorawanConfig.GetApplicationId(),
			ApplicationKey:   lorawanConfig.GetApplicationKey(),
			NetworkKey:       pgtype.Text{String: lorawanConfig.GetNetworkKey(), Valid: lorawanConfig.GetNetworkKey() != ""},
			ActivationMethod: int32(lorawanConfig.GetActivationMethod()),
			FrequencyPlanID:  lorawanConfig.GetFrequencyPlan(),
			HardwareTypeID:   lorawanConfig.GetHardwareData().GetHardwareTypeId(),
		}

		_, err = txQueries.CreateLoRaWANConfig(ctx, lorawanParams)
		if err != nil {
			return stacktrace.NewStackTraceError(err)
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}

// GetLoRaWANHardwareType retrieves a LoRaWAN hardware type by ID from the database.
func (store *EndDeviceStore) GetLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) (*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetLoRaWANHardwareType")
	defer span.End()

	hwType, err := store.db.GetLoRaWANHardwareType(ctx, hardwareTypeID)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	// Build LoRaWAN hardware data using builder pattern
	hardwareDataBuilder := iotv1.LoRaWANHardwareData_builder{
		HardwareTypeId:  hwType.ID,
		Name:            hwType.Name,
		Description:     hwType.Description.String,
		Manufacturer:    hwType.Manufacturer,
		Model:           hwType.Model,
		FirmwareVersion: hwType.FirmwareVersion.String,
		HardwareVersion: hwType.HardwareVersion.String,
		Profile:         hwType.Profile.String,
		LorawanVersion:  iotv1.LORAWANVersion(hwType.LorawanVersion),
	}

	return hardwareDataBuilder.Build(), nil
}

// GetCompleteLoRaWANDevice retrieves a complete LoRaWAN device with its configuration from the database.
func (store *EndDeviceStore) GetCompleteLoRaWANDevice(ctx context.Context, endDeviceID string) (*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetCompleteLoRaWANDevice")
	defer span.End()

	deviceData, err := store.db.GetCompleteLoRaWANDevice(ctx, endDeviceID)
	if err != nil {
		return nil, err
	}

	// Build complete end device using builder pattern
	endDeviceBuilder := iotv1.EndDevice_builder{
		Id:     deviceData.EndDeviceID,
		Name:   deviceData.Name,
		Status: iotv1.EndDeviceStatus(deviceData.Status),
		// TODO: Add complete LoRaWAN configuration once protobuf structure is confirmed
	}

	return endDeviceBuilder.Build(), nil
}

// ListEndDevicesByOrganization retrieves all end devices belonging to an organization from the database.
func (store *EndDeviceStore) ListEndDevicesByOrganization(ctx context.Context, organizationID string) ([]*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListEndDevicesByOrganization")
	defer span.End()

	devices, err := store.db.ListCompleteLoRaWANDevicesByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	endDevices := make([]*iotv1.EndDevice, len(devices))
	for i, device := range devices {
		// Build each end device using builder pattern
		endDeviceBuilder := iotv1.EndDevice_builder{
			Id:     device.EndDeviceID,
			Name:   device.Name,
			Status: iotv1.EndDeviceStatus(device.Status),
			// Note: This simplified version doesn't include full LoRaWAN config
			// Use GetCompleteLoRaWANDevice for full details
		}
		endDevices[i] = endDeviceBuilder.Build()
	}

	return endDevices, nil
}

// AddLoRaWANHardwareType inserts a new LoRaWAN hardware type into the database.
func (store *EndDeviceStore) AddLoRaWANHardwareType(ctx context.Context, hardwareData *iotv1.LoRaWANHardwareData) error {
	ctx, span := telemetry.Tracer().Start(ctx, "AddLoRaWANHardwareType")
	defer span.End()

	params := sqlc.CreateLoRaWANHardwareTypeParams{
		ID:              hardwareData.GetHardwareTypeId(),
		Name:            hardwareData.GetName(),
		Description:     pgtype.Text{String: hardwareData.GetDescription(), Valid: hardwareData.GetDescription() != ""},
		Manufacturer:    hardwareData.GetManufacturer(),
		Model:           hardwareData.GetModel(),
		FirmwareVersion: pgtype.Text{String: hardwareData.GetFirmwareVersion(), Valid: hardwareData.GetFirmwareVersion() != ""},
		HardwareVersion: pgtype.Text{String: hardwareData.GetHardwareVersion(), Valid: hardwareData.GetHardwareVersion() != ""},
		Profile:         pgtype.Text{String: hardwareData.GetProfile(), Valid: hardwareData.GetProfile() != ""},
		LorawanVersion:  int32(hardwareData.GetLorawanVersion()),
	}

	_, err := store.db.CreateLoRaWANHardwareType(ctx, params)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}

// ListLoRaWANHardwareTypes retrieves all available LoRaWAN hardware types from the database.
func (store *EndDeviceStore) ListLoRaWANHardwareTypes(ctx context.Context) ([]*iotv1.LoRaWANHardwareData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListLoRaWANHardwareTypes")
	defer span.End()

	hwTypes, err := store.db.ListLoRaWANHardwareTypes(ctx)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	hardwareTypes := make([]*iotv1.LoRaWANHardwareData, len(hwTypes))
	for i, hwType := range hwTypes {
		hardwareData := iotv1.LoRaWANHardwareData_builder{
			HardwareTypeId:  hwType.ID,
			Name:            hwType.Name,
			Description:     hwType.Description.String,
			Manufacturer:    hwType.Manufacturer,
			Model:           hwType.Model,
			FirmwareVersion: hwType.FirmwareVersion.String,
			HardwareVersion: hwType.HardwareVersion.String,
			Profile:         hwType.Profile.String,
			LorawanVersion:  iotv1.LORAWANVersion(hwType.LorawanVersion),
		}.Build()

		hardwareTypes[i] = hardwareData
	}

	return hardwareTypes, nil
}

// UpdateLoRaWANHardwareType updates an existing LoRaWAN hardware type in the database.
func (store *EndDeviceStore) UpdateLoRaWANHardwareType(ctx context.Context, hardwareData *iotv1.LoRaWANHardwareData) error {
	ctx, span := telemetry.Tracer().Start(ctx, "UpdateLoRaWANHardwareType")
	defer span.End()

	params := sqlc.UpdateLoRaWANHardwareTypeParams{
		ID:              hardwareData.GetHardwareTypeId(),
		Name:            hardwareData.GetName(),
		Description:     pgtype.Text{String: hardwareData.GetDescription(), Valid: hardwareData.GetDescription() != ""},
		Manufacturer:    hardwareData.GetManufacturer(),
		Model:           hardwareData.GetModel(),
		FirmwareVersion: pgtype.Text{String: hardwareData.GetFirmwareVersion(), Valid: hardwareData.GetFirmwareVersion() != ""},
		HardwareVersion: pgtype.Text{String: hardwareData.GetHardwareVersion(), Valid: hardwareData.GetHardwareVersion() != ""},
		Profile:         pgtype.Text{String: hardwareData.GetProfile(), Valid: hardwareData.GetProfile() != ""},
		LorawanVersion:  int32(hardwareData.GetLorawanVersion()),
	}

	_, err := store.db.UpdateLoRaWANHardwareType(ctx, params)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}

// DeleteLoRaWANHardwareType deletes a LoRaWAN hardware type by ID from the database.
func (store *EndDeviceStore) DeleteLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "DeleteLoRaWANHardwareType")
	defer span.End()

	err := store.db.DeleteLoRaWANHardwareType(ctx, hardwareTypeID)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}
