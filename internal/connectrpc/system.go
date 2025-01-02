package connectrpc

import (
	"context"

	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (handler *SystemHandler) CreateSystem(ctx context.Context, req *connect.Request[ponixv1.CreateSystemRequest]) (*connect.Response[ponixv1.CreateSystemResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystem")
	defer span.End()

	return nil, nil
}

func (handler *SystemHandler) System(ctx context.Context, req *connect.Request[ponixv1.SystemRequest]) (*connect.Response[ponixv1.SystemResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "System")
	defer span.End()

	return nil, nil
}

func (handler *SystemHandler) SystemNetworkServers(ctx context.Context, req *connect.Request[ponixv1.SystemNetworkServersRequest]) (*connect.Response[ponixv1.SystemNetworkServersResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemNetworkServers")
	defer span.End()

	return nil, nil
}

func (handler *SystemHandler) SystemGateways(ctx context.Context, req *connect.Request[ponixv1.SystemGatewaysRequest]) (*connect.Response[ponixv1.SystemGatewaysResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemGateways")
	defer span.End()

	return nil, nil
}

func (handler *SystemHandler) SystemEndDevices(ctx context.Context, req *connect.Request[ponixv1.SystemEndDevicesRequest]) (*connect.Response[ponixv1.SystemEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemEndDevices")
	defer span.End()

	return nil, nil
}

func (handler *SystemHandler) SystemInputs(ctx context.Context, req *connect.Request[ponixv1.SystemInputsRequest]) (*connect.Response[ponixv1.SystemInputsResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInputs")
	defer span.End()

	return nil, nil
}
