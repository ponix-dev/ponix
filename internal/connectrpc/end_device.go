package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type EndDeviceManager interface {
	CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organization string) (*iotv1.EndDevice, error)
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

	// TODO: Extract organization  from authentication context or request headers
	// For now, using a placeholder organization
	organization := "org_placeholder_123"

	endDevice, err := handler.endDeviceManager.CreateEndDevice(ctx, req.Msg, organization)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(iotv1.CreateEndDeviceResponse_builder{
		EndDevice: endDevice,
	}.Build())

	return resp, nil
}

func (handler *EndDeviceHandler) EndDevice(ctx context.Context, req *connect.Request[iotv1.EndDeviceRequest]) (*connect.Response[iotv1.EndDeviceResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDevice")
	defer span.End()

	return nil, nil
}

func (handler *EndDeviceHandler) OrganizationEndDevices(ctx context.Context, req *connect.Request[iotv1.OrganizationEndDevicesRequest]) (*connect.Response[iotv1.OrganizationEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "OrganizationEndDevices")
	defer span.End()

	resp := connect.NewResponse(iotv1.OrganizationEndDevicesResponse_builder{}.Build())

	return resp, nil
}

func (handler *EndDeviceHandler) EndDeviceData(ctx context.Context, req *connect.Request[iotv1.EndDeviceDataRequest]) (*connect.Response[iotv1.EndDeviceDataResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDeviceData")
	defer span.End()

	resp := connect.NewResponse(iotv1.EndDeviceDataResponse_builder{}.Build())

	return resp, nil
}
