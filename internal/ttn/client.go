package ttn

import (
	"context"
	"crypto/tls"
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

func WitCollaboratorApiKey(key string, collab string) TTNClientOption {
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

func (ttnClient *TTNClient) CreateGateway(ctx context.Context, gateway *iotv1.Gateway) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateGateway")
	defer span.End()

	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

	req := lorawanv3.CreateGatewayRequest_builder{
		Collaborator: apiCollaborator(ttnClient.ApiCollaborator),
		Gateway:      lorawanv3.Gateway_builder{}.Build(),
	}.Build()

	_, err := ttnClient.gatewayRegistryClient.Create(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (ttnClient *TTNClient) ListGateways(ctx context.Context) ([]*iotv1.Gateway, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "ListGateways")
	defer span.End()

	ctx = setAuthorizationContext(ctx, ttnClient.ApiKey)

	req := lorawanv3.ListGatewaysRequest_builder{
		Collaborator: apiCollaborator(ttnClient.ApiCollaborator),
		FieldMask:    gatewayFieldMask(),
	}.Build()

	resp, err := ttnClient.gatewayRegistryClient.List(ctx, req)
	if err != nil {
		return nil, err
	}

	respGws := resp.GetGateways()

	gws := make([]*iotv1.Gateway, len(respGws))

	for i, rgw := range respGws {
		gw := iotv1.Gateway_builder{
			Name: rgw.GetName(),
		}.Build()

		gws[i] = gw
	}

	return gws, nil
}

func gatewayFieldMask() *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{
		Paths: []string{
			"description",
			"name",
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
