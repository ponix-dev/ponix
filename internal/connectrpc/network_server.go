package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type NetworkServerHandler struct{}

func NewNetworkServerHandler() *NetworkServerHandler {
	return &NetworkServerHandler{}
}

func (handler *NetworkServerHandler) CreateNetworkServer(ctx context.Context, req *connect.Request[iotv1.CreateNetworkServerRequest]) (*connect.Response[iotv1.CreateNetworkServerResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateNetworkServer")
	defer span.End()

	return nil, nil
}

func (handler *NetworkServerHandler) NetworkServer(ctx context.Context, req *connect.Request[iotv1.NetworkServerRequest]) (*connect.Response[iotv1.NetworkServerResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "NetworkServer")
	defer span.End()

	return nil, nil
}

func (handler *NetworkServerHandler) NetworkServerGateways(ctx context.Context, req *connect.Request[iotv1.NetworkServerGatewaysRequest]) (*connect.Response[iotv1.NetworkServerGatewaysResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "NetworkServerGateways")
	defer span.End()

	return nil, nil
}

func (handler *NetworkServerHandler) NetworkServerEndDevices(ctx context.Context, req *connect.Request[iotv1.NetworkServerEndDevicesRequest]) (*connect.Response[iotv1.NetworkServerEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "NetworkServerEndDevices")
	defer span.End()

	return nil, nil
}
