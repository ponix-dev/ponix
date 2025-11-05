package domain

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// EndDeviceRegister defines the operations for registering devices with external systems.
type EndDeviceRegister interface {
	RegisterEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) error
}

// EndDeviceStorer defines the persistence operations for end devices.
type EndDeviceStorer interface {
	AddEndDevice(ctx context.Context, endDevice *iotv1.EndDevice, organizationId string) error
	GetLoRaWANHardwareType(ctx context.Context, hardwareTypeId string) (*iotv1.LoRaWANHardwareData, error)
	GetEndDeviceWithOrganization(ctx context.Context, endDeviceID string) (*iotv1.EndDevice, string, error)
}

// EndDeviceManager orchestrates end device business logic including creation and external registration.
type EndDeviceManager struct {
	endDeviceStore    EndDeviceStorer
	endDeviceRegister EndDeviceRegister
	applicationId     string
	stringId          StringId
	validate          Validate
}

// NewEndDeviceManager creates a new instance of EndDeviceManager with the provided dependencies.
func NewEndDeviceManager(eds EndDeviceStorer, edr EndDeviceRegister, applicationId string, stringId StringId, validate Validate) *EndDeviceManager {
	return &EndDeviceManager{
		endDeviceStore:    eds,
		endDeviceRegister: edr,
		applicationId:     applicationId,
		stringId:          stringId,
		validate:          validate,
	}
}

// CreateEndDevice creates a new end device, registers it with external systems if needed, and persists it.
func (mgr *EndDeviceManager) CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organizationId string) (*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	endDeviceId := mgr.stringId()

	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	endDevice, err := mgr.buildEndDeviceFromRequest(ctx, endDeviceId, createReq)
	if err != nil {
		return nil, err
	}

	// Only register with external systems for LoRaWAN devices
	switch endDevice.GetHardwareType() {
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
		err = mgr.endDeviceRegister.RegisterEndDevice(ctx, endDevice)
		if err != nil {
			return nil, stacktrace.NewStackTraceError(err)
		}
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_HTTP:
		// HTTP devices don't need external registration
		// Continue to storage
	}

	// Store the device in the database
	err = mgr.endDeviceStore.AddEndDevice(ctx, endDevice, organizationId)
	if err != nil {
		return nil, err
	}

	return endDevice, nil
}

// buildEndDeviceFromRequest constructs a complete EndDevice from the request including hardware-specific configuration.
func (mgr *EndDeviceManager) buildEndDeviceFromRequest(ctx context.Context, endDeviceId string, createReq *iotv1.CreateEndDeviceRequest) (*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "buildEndDeviceFromRequest")
	defer span.End()

	// Create base EndDevice using builder pattern
	endDeviceBuilder := iotv1.EndDevice_builder{
		Id:           endDeviceId,
		Name:         createReq.GetName(),
		Description:  createReq.GetDescription(),
		Status:       iotv1.EndDeviceStatus_END_DEVICE_STATUS_PENDING,
		HardwareType: createReq.GetHardwareType(),
		// Note: data_type is deprecated and not set
	}

	// Handle hardware-specific configuration
	switch createReq.GetHardwareType() {
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
		lorawanConfig, err := mgr.buildLoRaWANConfig(ctx, createReq)
		if err != nil {
			return nil, err
		}
		endDeviceBuilder.LorawanConfig = lorawanConfig
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_HTTP:
		// HTTP devices don't need additional configuration
		// Just validate and continue
	default:
		return nil, stacktrace.NewStackTraceErrorf("unsupported hardware type: %v", createReq.GetHardwareType())
	}

	return endDeviceBuilder.Build(), nil
}

// buildLoRaWANConfig constructs a complete LoRaWAN configuration including device identifiers, keys, and hardware data.
func (mgr *EndDeviceManager) buildLoRaWANConfig(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest) (*iotv1.LoRaWANConfig, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "buildLoRaWANConfig")
	defer span.End()

	// Fetch hardware type data from database
	hardwareData, err := mgr.endDeviceStore.GetLoRaWANHardwareType(ctx, createReq.GetHardwareTypeId())
	if err != nil {
		return nil, err
	}

	// Build LoRaWAN configuration using builder pattern
	lorawanConfigBuilder := iotv1.LoRaWANConfig_builder{
		DeviceEui:        generateDeviceEUI(),                           // Generate unique device EUI
		ApplicationEui:   generateApplicationEUI(),                      // Generate or use default app EUI
		ApplicationId:    mgr.applicationId,                             // Should come from context/config
		ApplicationKey:   generateApplicationKey(),                      // Generate 128-bit key
		NetworkKey:       generateNetworkKey(),                          // Generate 128-bit key for LoRaWAN 1.1+
		ActivationMethod: iotv1.ActivationMethod_ACTIVATION_METHOD_OTAA, // Default to OTAA
		FrequencyPlan:    string(FreqPlanUS902_928),                     // Default US frequency plan
		HardwareData:     hardwareData,
	}

	return lorawanConfigBuilder.Build(), nil
}

// generateDeviceEUI generates a unique 64-bit device EUI (16 hex characters).
// TODO: Implement proper cryptographically secure ID generation.
func generateDeviceEUI() string {
	return "0123456789ABCDEF"
}

// generateApplicationEUI generates a 64-bit application EUI (16 hex characters).
// TODO: Implement proper cryptographically secure ID generation.
func generateApplicationEUI() string {
	return "FEDCBA9876543210"
}

// generateApplicationKey generates a 128-bit application key (32 hex characters).
// TODO: Implement proper cryptographically secure key generation.
func generateApplicationKey() string {
	return "00112233445566778899AABBCCDDEEFF"
}

// generateNetworkKey generates a 128-bit network key (32 hex characters).
// TODO: Implement proper cryptographically secure key generation.
func generateNetworkKey() string {
	return "FFEEDDCCBBAA99887766554433221100"
}
