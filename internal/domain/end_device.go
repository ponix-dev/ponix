package domain

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

type EndDeviceRegister interface {
	RegisterEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) error
}

type EndDeviceStorer interface {
	AddEndDevice(ctx context.Context, endDevice *iotv1.EndDevice, organizationID string) error
	GetLoRaWANHardwareType(ctx context.Context, hardwareTypeID string) (*iotv1.LoRaWANHardwareData, error)
}

type EndDeviceManager struct {
	endDeviceStore    EndDeviceStorer
	endDeviceRegister EndDeviceRegister
	applicationId     string
	stringId          StringId
	validate          Validate
}

func NewEndDeviceManager(eds EndDeviceStorer, edr EndDeviceRegister, applicationId string, stringId StringId, validate Validate) *EndDeviceManager {
	return &EndDeviceManager{
		endDeviceStore:    eds,
		endDeviceRegister: edr,
		applicationId:     applicationId,
		stringId:          stringId,
		validate:          validate,
	}
}

func (mgr *EndDeviceManager) CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organizationID string) (*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	endDeviceID := mgr.stringId()

	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	endDevice, err := mgr.buildEndDeviceFromRequest(ctx, endDeviceID, createReq)
	if err != nil {
		return nil, err
	}

	// may want to split this in to function calls when there are more types
	switch endDevice.GetHardwareType() {
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
		err = mgr.endDeviceRegister.RegisterEndDevice(ctx, endDevice)
		if err != nil {
			return nil, err
		}
	}

	// Store the device in the database
	err = mgr.endDeviceStore.AddEndDevice(ctx, endDevice, organizationID)
	if err != nil {
		return nil, err
	}

	return endDevice, nil
}

// buildEndDeviceFromRequest constructs a complete EndDevice from CreateEndDeviceRequest
func (mgr *EndDeviceManager) buildEndDeviceFromRequest(ctx context.Context, endDeviceID string, createReq *iotv1.CreateEndDeviceRequest) (*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "buildEndDeviceFromRequest")
	defer span.End()

	// Create base EndDevice using builder pattern
	endDeviceBuilder := iotv1.EndDevice_builder{
		Id:           endDeviceID,
		Name:         createReq.GetName(),
		Description:  createReq.GetDescription(),
		Status:       iotv1.EndDeviceStatus_END_DEVICE_STATUS_PENDING,
		DataType:     createReq.GetDataType(),
		HardwareType: createReq.GetHardwareType(),
	}

	// Handle hardware-specific configuration
	switch createReq.GetHardwareType() {
	case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
		lorawanConfig, err := mgr.buildLoRaWANConfig(ctx, createReq)
		if err != nil {
			return nil, err
		}
		endDeviceBuilder.LorawanConfig = lorawanConfig
	default:
		return nil, stacktrace.NewStackTraceErrorf("unsupported hardware type: %v", createReq.GetHardwareType())
	}

	return endDeviceBuilder.Build(), nil
}

// buildLoRaWANConfig constructs LoRaWAN configuration with hardware data from database
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

// Placeholder functions for generating LoRaWAN identifiers
// TODO: Implement proper ID generation logic
func generateDeviceEUI() string {
	// Generate unique 64-bit device EUI (16 hex chars)
	return "0123456789ABCDEF"
}

func generateApplicationEUI() string {
	// Generate or use default 64-bit application EUI (16 hex chars)
	return "FEDCBA9876543210"
}

func generateApplicationKey() string {
	// Generate 128-bit application key (32 hex chars)
	return "00112233445566778899AABBCCDDEEFF"
}

func generateNetworkKey() string {
	// Generate 128-bit network key (32 hex chars)
	return "FFEEDDCCBBAA99887766554433221100"
}
