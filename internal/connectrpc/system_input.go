package connectrpc

import (
	"context"

	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemInputHandler struct{}

func NewSystemInputHandler() *SystemInputHandler {
	return &SystemInputHandler{}
}

func (handler *SystemInputHandler) CreateSystemInput(ctx context.Context, req *connect.Request[ponixv1.CreateSystemInputRequest]) (*connect.Response[ponixv1.CreateSystemInputResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystemInput")
	defer span.End()

	return nil, nil
}

func (handler *SystemInputHandler) SystemInput(ctx context.Context, req *connect.Request[ponixv1.SystemInputRequest]) (*connect.Response[ponixv1.SystemInputResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInput")
	defer span.End()

	return nil, nil
}

func (handler *SystemInputHandler) SystemInputEndDevices(ctx context.Context, req *connect.Request[ponixv1.SystemInputEndDevicesRequest]) (*connect.Response[ponixv1.SystemInputEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInputEndDevices")
	defer span.End()

	return nil, nil
}
