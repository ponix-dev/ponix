package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type GatewayManager interface {
	CreateGateway(ctx context.Context, gateway *iotv1.Gateway) (string, error)
}

type GatewayHandler struct {
	gatewayManager GatewayManager
}

func NewGatewayHandler(gmgr GatewayManager) *GatewayHandler {
	return &GatewayHandler{
		gatewayManager: gmgr,
	}
}

func (handler *GatewayHandler) CreateGateway(ctx context.Context, req *connect.Request[iotv1.CreateGatewayRequest]) (*connect.Response[iotv1.CreateGatewayResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateGateway")
	defer span.End()

	gateway := &iotv1.Gateway_builder{
		SystemId:        req.Msg.GetSystemId(),
		NetworkServerId: req.Msg.GetNetworkServerId(),
		Name:            req.Msg.GetName(),
	}

	id, err := handler.gatewayManager.CreateGateway(ctx, gateway.Build())
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.CreateGatewayResponse_builder{
		GatewayId: id,
	}.Build())

	return resp, nil
}
func (handler *GatewayHandler) Gateway(ctx context.Context, req *connect.Request[iotv1.GatewayRequest]) (*connect.Response[iotv1.GatewayResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "Gateway")
	defer span.End()

	return nil, nil
}
func (handler *GatewayHandler) GatewayEndDevices(ctx context.Context, req *connect.Request[iotv1.GatewayEndDevicesRequest]) (*connect.Response[iotv1.GatewayEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GatewayEndDevices")
	defer span.End()

	return nil, nil
}
