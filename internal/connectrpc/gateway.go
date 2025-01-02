package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type GatewayHandler struct{}

func NewGatewayHandler() *GatewayHandler {
	return &GatewayHandler{}
}

func (handler *GatewayHandler) CreateGateway(ctx context.Context, req *connect.Request[iotv1.CreateGatewayRequest]) (*connect.Response[iotv1.CreateGatewayResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateGateway")
	defer span.End()

	return nil, nil
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
