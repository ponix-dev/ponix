package connectrpc

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemManager interface {
	CreateSystem(ctx context.Context, system *ponixv1.System) (string, error)
	System(ctx context.Context, systemId string) (*ponixv1.System, error)
}

type SystemNSManager interface {
	SystemNetworkServers(ctx context.Context, systemId string) ([]*iotv1.NetworkServer, error)
}

type SystemGatewayManager interface {
	SystemGateways(ctx context.Context, systemId string) ([]*iotv1.Gateway, error)
}

type SystemEndDeviceManager interface {
	SystemEndDevices(ctx context.Context, systemId string) ([]*iotv1.EndDevice, error)
}

type SystemInputsManager interface {
	SystemInputs(ctx context.Context, systemId string) ([]*ponixv1.SystemInput, error)
}

type SystemHandler struct {
	systemManager  SystemManager
	nsManager      SystemNSManager
	gatewayManager SystemGatewayManager
	edManager      SystemEndDeviceManager
	siManager      SystemInputsManager
}

func NewSystemHandler(smgr SystemManager, snsmgr SystemNSManager, gmgr SystemGatewayManager, edmgr SystemEndDeviceManager, simgr SystemInputsManager) *SystemHandler {
	return &SystemHandler{
		systemManager:  smgr,
		nsManager:      snsmgr,
		gatewayManager: gmgr,
		edManager:      edmgr,
		siManager:      simgr,
	}
}

func (handler *SystemHandler) CreateSystem(ctx context.Context, req *connect.Request[ponixv1.CreateSystemRequest]) (*connect.Response[ponixv1.CreateSystemResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystem")
	defer span.End()

	system := &ponixv1.System{
		OrganizationId: req.Msg.GetOrganizationId(),
		Name:           req.Msg.GetName(),
	}

	id, err := handler.systemManager.CreateSystem(ctx, system)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&ponixv1.CreateSystemResponse{
		SystemId: id,
	})

	return resp, nil
}

func (handler *SystemHandler) System(ctx context.Context, req *connect.Request[ponixv1.SystemRequest]) (*connect.Response[ponixv1.SystemResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "System")
	defer span.End()

	system, err := handler.systemManager.System(ctx, req.Msg.SystemId)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&ponixv1.SystemResponse{
		System: system,
	})

	return resp, nil
}

func (handler *SystemHandler) SystemNetworkServers(ctx context.Context, req *connect.Request[ponixv1.SystemNetworkServersRequest]) (*connect.Response[ponixv1.SystemNetworkServersResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemNetworkServers")
	defer span.End()

	servers, err := handler.nsManager.SystemNetworkServers(ctx, req.Msg.SystemId)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&ponixv1.SystemNetworkServersResponse{
		NetworkServers: servers,
	})

	return resp, nil
}

func (handler *SystemHandler) SystemGateways(ctx context.Context, req *connect.Request[ponixv1.SystemGatewaysRequest]) (*connect.Response[ponixv1.SystemGatewaysResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemGateways")
	defer span.End()

	gateways, err := handler.gatewayManager.SystemGateways(ctx, req.Msg.SystemId)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&ponixv1.SystemGatewaysResponse{
		Gateways: gateways,
	})

	return resp, nil
}

func (handler *SystemHandler) SystemEndDevices(ctx context.Context, req *connect.Request[ponixv1.SystemEndDevicesRequest]) (*connect.Response[ponixv1.SystemEndDevicesResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemEndDevices")
	defer span.End()

	endDevices, err := handler.edManager.SystemEndDevices(ctx, req.Msg.SystemId)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&ponixv1.SystemEndDevicesResponse{
		EndDevices: endDevices,
	})

	return resp, nil
}

func (handler *SystemHandler) SystemInputs(ctx context.Context, req *connect.Request[ponixv1.SystemInputsRequest]) (*connect.Response[ponixv1.SystemInputsResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInputs")
	defer span.End()

	inputs, err := handler.siManager.SystemInputs(ctx, req.Msg.SystemId)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&ponixv1.SystemInputsResponse{
		Inputs: inputs,
	})

	return resp, nil
}
