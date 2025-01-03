package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type EndDeviceManager interface {
	CreateEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) (string, error)
}

type EndDeviceHandler struct {
	endDeviceManager EndDeviceManager
}

func NewEndDeviceHandler(edmgr EndDeviceManager) *EndDeviceHandler {
	return &EndDeviceHandler{
		endDeviceManager: edmgr,
	}
}

func (handler *EndDeviceHandler) CreateEndDevice(ctx context.Context, req *connect.Request[iotv1.CreateEndDeviceRequest]) (*connect.Response[iotv1.CreateEndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	endDevice := &iotv1.EndDevice{
		NetworkServerId: req.Msg.GetNetworkServerId(),
		SystemId:        req.Msg.GetSystemId(),
		Name:            req.Msg.GetName(),
	}

	id, err := handler.endDeviceManager.CreateEndDevice(ctx, endDevice)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&iotv1.CreateEndDeviceResponse{
		EndDeviceId: id,
	})

	return resp, nil
}

func (handler *EndDeviceHandler) EndDevice(ctx context.Context, req *connect.Request[iotv1.EndDeviceRequest]) (*connect.Response[iotv1.EndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDevice")
	defer span.End()

	return nil, nil
}
