package ttn

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"buf.build/gen/go/thethingsindustries/lorawan-stack/grpc/go/ttn/lorawan/v3/lorawanv3grpc"
	lorawanv3 "buf.build/gen/go/thethingsindustries/lorawan-stack/protocolbuffers/go/ttn/lorawan/v3"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type TTNRegion string

type Application struct {
	ID          string
	Name        string
	Description string
}

const (
	TTNRegionEU    TTNRegion = "eu1"
	TTNRegionNam   TTNRegion = "nam1"
	TTNRegionLocal TTNRegion = "local"
)

func (region TTNRegion) String() string {
	return string(region)
}

type TTNClient struct {
	ServerName               string
	Region                   TTNRegion
	ApiKey                   string
	ApiCollaborator          string
	IdentityServerAddress    string
	GatewayServerAddress     string
	NetworkServerAddress     string
	ApplicationServerAddress string
	JoinServerAddress        string
	grpcConns                map[string]*grpc.ClientConn
	appRegistryClient        lorawanv3grpc.ApplicationRegistryClient
	gatewayRegistryClient    lorawanv3grpc.GatewayRegistryClient
	endDeviceRegistryClient  lorawanv3grpc.EndDeviceRegistryClient
}

type TTNClientOption func(*TTNClient)

func WithServerName(name string) TTNClientOption {
	return func(t *TTNClient) {
		t.ServerName = name
	}
}

func WithRegion(region TTNRegion) TTNClientOption {
	return func(t *TTNClient) {
		t.Region = region
	}
}

func WithCollaboratorApiKey(key string, collab string) TTNClientOption {
	return func(t *TTNClient) {
		t.ApiKey = key
		t.ApiCollaborator = collab
	}
}

// TODO: implement address options

func NewTTNClient(opts ...TTNClientOption) (*TTNClient, error) {
	ttnClient := &TTNClient{
		Region:    TTNRegionLocal,
		grpcConns: map[string]*grpc.ClientConn{},
	}

	for _, opt := range opts {
		opt(ttnClient)
	}

	if ttnClient.ServerName == "" {
		return nil, errors.New("ttn client requires a server name")
	}

	if ttnClient.ApiKey == "" {
		return nil, errors.New("ttn client requires an api key")
	}

	var err error
	switch ttnClient.Region {
	case TTNRegionLocal:
		err = configureLocalTTNRegionClient(ttnClient)
	case TTNRegionNam:
		err = configureNAMTTNClient(ttnClient)
	case TTNRegionEU:
		err = configureEUTTNClient(ttnClient)
	default:
		return nil, errors.New("ttn client does not support provided region")
	}
	if err != nil {
		return nil, err
	}

	ttnClient.appRegistryClient = lorawanv3grpc.NewApplicationRegistryClient(ttnClient.grpcConns[ttnClient.IdentityServerAddress])
	ttnClient.gatewayRegistryClient = lorawanv3grpc.NewGatewayRegistryClient(ttnClient.grpcConns[ttnClient.IdentityServerAddress])
	ttnClient.endDeviceRegistryClient = lorawanv3grpc.NewEndDeviceRegistryClient(ttnClient.grpcConns[ttnClient.IdentityServerAddress])
	return ttnClient, nil
}

func configureLocalTTNRegionClient(ttnClient *TTNClient) error {
	if ttnClient.IdentityServerAddress == "" ||
		ttnClient.ApplicationServerAddress == "" ||
		ttnClient.GatewayServerAddress == "" ||
		ttnClient.JoinServerAddress == "" ||
		ttnClient.NetworkServerAddress == "" {
		return errors.New("local ttn client requires all server addresses set")
	}

	//TODO: implement this when better understood what is needed
	return errors.New("local client currently not supported")
}

func configureNAMTTNClient(ttnClient *TTNClient) error {
	euAddress := formatTTNCloudAddress(ttnClient.ServerName, TTNRegionEU)
	namAddress := formatTTNCloudAddress(ttnClient.ServerName, ttnClient.Region)

	ttnClient.IdentityServerAddress = euAddress
	ttnClient.ApplicationServerAddress = namAddress
	ttnClient.GatewayServerAddress = namAddress
	ttnClient.JoinServerAddress = namAddress
	ttnClient.NetworkServerAddress = namAddress

	euConn, err := grpcConn(euAddress)
	if err != nil {
		return err
	}

	ttnClient.grpcConns[euAddress] = euConn

	namConn, err := grpcConn(namAddress)
	if err != nil {
		return err
	}

	ttnClient.grpcConns[namAddress] = namConn

	return nil
}

func configureEUTTNClient(ttnClient *TTNClient) error {
	euAddress := formatTTNCloudAddress(ttnClient.ServerName, ttnClient.Region)

	ttnClient.IdentityServerAddress = euAddress
	ttnClient.ApplicationServerAddress = euAddress
	ttnClient.GatewayServerAddress = euAddress
	ttnClient.JoinServerAddress = euAddress
	ttnClient.NetworkServerAddress = euAddress

	euConn, err := grpcConn(euAddress)
	if err != nil {
		return err
	}

	ttnClient.grpcConns[euAddress] = euConn

	return nil
}

func formatTTNCloudAddress(serverName string, region TTNRegion) string {
	return fmt.Sprintf("%s.%s.cloud.thethings.industries", serverName, region.String())
}

func grpcConn(address string) (conn *grpc.ClientConn, err error) {
	return grpc.NewClient(address, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
}

// func (ttnClient *TTNClient) CreateGateway(ctx context.Context, gateway *iotv1.Gateway) error {
// 	ctx, span := telemetry.Tracer().Start(ctx, "CreateGateway")
// 	defer span.End()

// 	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

// 	req := lorawanv3.CreateGatewayRequest_builder{
// 		Collaborator: apiCollaborator(ttnClient.ApiCollaborator),
// 		Gateway:      lorawanv3.Gateway_builder{}.Build(),
// 	}.Build()

// 	_, err := ttnClient.gatewayRegistryClient.Create(ctx, req)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (ttnClient *TTNClient) ListGateways(ctx context.Context) ([]*iotv1.Gateway, error) {
// 	ctx, span := telemetry.Tracer().Start(ctx, "ListGateways")
// 	defer span.End()

// 	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

// 	req := lorawanv3.ListGatewaysRequest_builder{
// 		Collaborator: apiCollaborator(ttnClient.ApiCollaborator),
// 		FieldMask:    gatewayFieldMask(),
// 	}.Build()

// 	resp, err := ttnClient.gatewayRegistryClient.List(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	respGws := resp.GetGateways()

// 	gws := make([]*iotv1.Gateway, len(respGws))

// 	for i, rgw := range respGws {
// 		gw := iotv1.Gateway_builder{
// 			Name: rgw.GetName(),
// 		}.Build()

// 		gws[i] = gw
// 	}

// 	return gws, nil
// }

func (ttnClient *TTNClient) RegisterEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

	// Extract LoRaWAN configuration from the end device
	lorawanConfig := endDevice.GetLorawanConfig()
	if lorawanConfig == nil {
		return fmt.Errorf("LoRaWAN configuration is required for TTN registration")
	}

	// Build TTN end device identifiers
	endDeviceIds := lorawanv3.EndDeviceIdentifiers_builder{
		ApplicationIds: lorawanv3.ApplicationIdentifiers_builder{
			ApplicationId: lorawanConfig.GetApplicationId(),
		}.Build(),
		DeviceId: endDevice.GetId(),
		DevEui:   parseEUI(lorawanConfig.GetDeviceEui()),
		JoinEui:  parseEUI(lorawanConfig.GetApplicationEui()),
	}.Build()

	// Build root keys for OTAA
	var rootKeys *lorawanv3.RootKeys
	if lorawanConfig.GetActivationMethod() == iotv1.ActivationMethod_ACTIVATION_METHOD_OTAA {
		rootKeysBuilder := lorawanv3.RootKeys_builder{
			AppKey: lorawanv3.KeyEnvelope_builder{
				Key: parseKey(lorawanConfig.GetApplicationKey()),
			}.Build(),
		}

		// Add network key for LoRaWAN 1.1+
		if lorawanConfig.GetNetworkKey() != "" {
			rootKeysBuilder.NwkKey = lorawanv3.KeyEnvelope_builder{
				Key: parseKey(lorawanConfig.GetNetworkKey()),
			}.Build()
		}

		rootKeys = rootKeysBuilder.Build()
	}

	// Build the complete TTN end device
	ttnEndDevice := lorawanv3.EndDevice_builder{
		Ids:                      endDeviceIds,
		Name:                     endDevice.GetName(),
		Description:              endDevice.GetDescription(),
		NetworkServerAddress:     ttnClient.NetworkServerAddress,
		ApplicationServerAddress: ttnClient.ApplicationServerAddress,
		JoinServerAddress:        ttnClient.JoinServerAddress,

		// LoRaWAN version configuration
		LorawanVersion: convertLoRaWANVersion(lorawanConfig.GetHardwareData().GetLorawanVersion()),

		// Frequency plan
		FrequencyPlanId: lorawanConfig.GetFrequencyPlan(),

		// Root keys for OTAA
		RootKeys: rootKeys,

		// Hardware and version info from hardware data
		VersionIds: buildVersionIds(lorawanConfig.GetHardwareData(), lorawanConfig.GetFrequencyPlan()),

		// Additional attributes
		Attributes: buildDeviceAttributes(endDevice, lorawanConfig),
	}.Build()

	// Create the registration request
	createRequest := lorawanv3.CreateEndDeviceRequest_builder{
		EndDevice: ttnEndDevice,
	}.Build()

	_, err := ttnClient.endDeviceRegistryClient.Create(ctx, createRequest)
	if err != nil {
		return fmt.Errorf("failed to register end device with TTN: %w", err)
	}

	return nil
}

func (ttnClient *TTNClient) ListEndDevices(ctx context.Context, applicationID string) ([]*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListEndDevices")
	defer span.End()

	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

	req := lorawanv3.ListEndDevicesRequest_builder{
		ApplicationIds: lorawanv3.ApplicationIdentifiers_builder{
			ApplicationId: applicationID,
		}.Build(),
		FieldMask: endDeviceFieldMask(),
	}.Build()

	resp, err := ttnClient.endDeviceRegistryClient.List(ctx, req)
	if err != nil {
		return nil, err
	}

	respDevices := resp.GetEndDevices()

	devices := make([]*iotv1.EndDevice, len(respDevices))

	for i, rDevice := range respDevices {
		device := iotv1.EndDevice_builder{
			Id:   rDevice.GetIds().GetDeviceId(),
			Name: rDevice.GetName(),
		}.Build()

		devices[i] = device
	}

	return devices, nil
}

func (ttnClient *TTNClient) ListApplications(ctx context.Context) ([]*Application, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListApplications")
	defer span.End()

	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

	req := lorawanv3.ListApplicationsRequest_builder{
		Collaborator: apiCollaborator(ttnClient.ApiCollaborator),
		FieldMask:    applicationFieldMask(),
	}.Build()

	resp, err := ttnClient.appRegistryClient.List(ctx, req)
	if err != nil {
		return nil, err
	}

	respApps := resp.GetApplications()

	apps := make([]*Application, len(respApps))

	for i, rApp := range respApps {
		app := &Application{
			ID:          rApp.GetIds().GetApplicationId(),
			Name:        rApp.GetName(),
			Description: rApp.GetDescription(),
		}

		apps[i] = app
	}

	return apps, nil
}

func gatewayFieldMask() *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{
		Paths: []string{
			"description",
			"name",
		},
	}
}

func endDeviceFieldMask() *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{
		Paths: []string{
			"ids",
			"name",
			"description",
		},
	}
}

func applicationFieldMask() *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{
		Paths: []string{
			"ids",
			"name",
			"description",
		},
	}
}

func setAuthorizationContext(ctx context.Context, key string) context.Context {
	md := metadata.Pairs(
		"authorization", fmt.Sprintf("Bearer %s", key),
	)

	return metadata.NewOutgoingContext(ctx, md)
}

func apiCollaborator(collab string) *lorawanv3.OrganizationOrUserIdentifiers {
	return lorawanv3.OrganizationOrUserIdentifiers_builder{
		UserIds: lorawanv3.UserIdentifiers_builder{
			UserId: collab,
		}.Build(),
	}.Build()
}

// parseEUI converts a hex string to an 8-byte EUI
func parseEUI(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil || len(bytes) != 8 {
		// Return zero EUI if parsing fails
		return make([]byte, 8)
	}
	return bytes
}

// parseKey converts a hex string to a 16-byte key
func parseKey(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil || len(bytes) != 16 {
		// Return zero key if parsing fails
		return make([]byte, 16)
	}
	return bytes
}

// convertLoRaWANVersion converts our protobuf enum to TTN's LoRaWAN version
func convertLoRaWANVersion(version iotv1.LORAWANVersion) lorawanv3.MACVersion {
	switch version {
	case iotv1.LORAWANVersion_LORAWAN_VERSION_1_0_0:
		return lorawanv3.MACVersion_MAC_V1_0
	case iotv1.LORAWANVersion_LORAWAN_VERSION_1_0_1:
		return lorawanv3.MACVersion_MAC_V1_0_1
	case iotv1.LORAWANVersion_LORAWAN_VERSION_1_0_2:
		return lorawanv3.MACVersion_MAC_V1_0_2
	case iotv1.LORAWANVersion_LORAWAN_VERSION_1_0_3:
		return lorawanv3.MACVersion_MAC_V1_0_3
	case iotv1.LORAWANVersion_LORAWAN_VERSION_1_0_4:
		return lorawanv3.MACVersion_MAC_V1_0_4
	case iotv1.LORAWANVersion_LORAWAN_VERSION_1_1_0:
		return lorawanv3.MACVersion_MAC_V1_1
	default:
		return lorawanv3.MACVersion_MAC_V1_0_3 // Default to 1.0.3
	}
}

// buildVersionIds creates version identifiers from hardware data
func buildVersionIds(hardwareData *iotv1.LoRaWANHardwareData, frequencyPlan string) *lorawanv3.EndDeviceVersionIdentifiers {
	if hardwareData == nil {
		return nil
	}

	return lorawanv3.EndDeviceVersionIdentifiers_builder{
		BrandId:         hardwareData.GetManufacturer(),
		ModelId:         hardwareData.GetModel(),
		HardwareVersion: hardwareData.GetHardwareVersion(),
		FirmwareVersion: hardwareData.GetFirmwareVersion(),
		BandId:          frequencyPlan,
	}.Build()
}

// buildDeviceAttributes creates device attributes from EndDevice and LoRaWAN config
func buildDeviceAttributes(endDevice *iotv1.EndDevice, lorawanConfig *iotv1.LoRaWANConfig) map[string]string {
	attributes := make(map[string]string)

	// Add basic device attributes
	if endDevice.GetDescription() != "" {
		attributes["description"] = endDevice.GetDescription()
	}

	// Add hardware information
	if hardwareData := lorawanConfig.GetHardwareData(); hardwareData != nil {
		if hardwareData.GetManufacturer() != "" {
			attributes["manufacturer"] = hardwareData.GetManufacturer()
		}
		if hardwareData.GetModel() != "" {
			attributes["model"] = hardwareData.GetModel()
		}
		if hardwareData.GetProfile() != "" {
			attributes["profile"] = hardwareData.GetProfile()
		}
	}

	// Add activation method
	switch lorawanConfig.GetActivationMethod() {
	case iotv1.ActivationMethod_ACTIVATION_METHOD_OTAA:
		attributes["activation_method"] = "OTAA"
	case iotv1.ActivationMethod_ACTIVATION_METHOD_ABP:
		attributes["activation_method"] = "ABP"
	}

	return attributes
}
