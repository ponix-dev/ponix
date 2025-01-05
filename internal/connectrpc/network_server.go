package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type NetworkServerManager interface {
	CreateNetworkServer(ctx context.Context, networkServer *iotv1.NetworkServer) (string, error)
}

type NetworkServerHandler struct {
	networkServerManager NetworkServerManager
}

func NewNetworkServerHandler(nsmgr NetworkServerManager) *NetworkServerHandler {
	return &NetworkServerHandler{
		networkServerManager: nsmgr,
	}
}

func (handler *NetworkServerHandler) CreateNetworkServer(ctx context.Context, req *connect.Request[iotv1.CreateNetworkServerRequest]) (*connect.Response[iotv1.CreateNetworkServerResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateNetworkServer")
	defer span.End()

	ns := &iotv1.NetworkServer_builder{
		SystemId:    req.Msg.GetSystemId(),
		Name:        req.Msg.GetName(),
		IotPlatform: req.Msg.GetIotPlatform(),
	}

	id, err := handler.networkServerManager.CreateNetworkServer(ctx, ns.Build())
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.CreateNetworkServerResponse_builder{
		NetworkServerId: id,
	}.Build())

	return resp, nil
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
