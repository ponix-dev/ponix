package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type EndDeviceHandler struct{}

func NewEndDeviceHandler() *EndDeviceHandler {
	return &EndDeviceHandler{}
}

func (handler *EndDeviceHandler) CreateEndDevice(ctx context.Context, req *connect.Request[iotv1.CreateEndDeviceRequest]) (*connect.Response[iotv1.CreateEndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	return nil, nil
}

func (handler *EndDeviceHandler) EndDevice(ctx context.Context, req *connect.Request[iotv1.EndDeviceRequest]) (*connect.Response[iotv1.EndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDevice")
	defer span.End()

	return nil, nil
}
