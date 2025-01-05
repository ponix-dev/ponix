package connectrpc

import (
	"context"

	aquaponicsv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/aquaponics/v1"
	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	soilponicsv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/soilponics/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemInputManager interface {
	CreateSystemInput(ctx context.Context, systemInput *ponixv1.SystemInput) (string, error)
}

type SystemInputHandler struct {
	systemInputManager SystemInputManager
}

func NewSystemInputHandler(simgr SystemInputManager) *SystemInputHandler {
	return &SystemInputHandler{
		systemInputManager: simgr,
	}
}

func (handler *SystemInputHandler) CreateSystemInput(ctx context.Context, req *connect.Request[ponixv1.CreateSystemInputRequest]) (*connect.Response[ponixv1.CreateSystemInputResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystemInput")
	defer span.End()

	si := &ponixv1.SystemInput_builder{
		Name:     req.Msg.GetName(),
		SystemId: req.Msg.GetSystemId(),
	}

	switch req.Msg.GetInputData().(type) {
	case *ponixv1.CreateSystemInputRequest_Field:
		si.Field = &soilponicsv1.FieldData{}
	case *ponixv1.CreateSystemInputRequest_GrowMedium:
		si.GrowMedium = &aquaponicsv1.GrowMediumData{
			MediumType: req.Msg.GetGrowMedium().GetMediumType(),
		}
	case *ponixv1.CreateSystemInputRequest_Tank:
		si.Tank = &aquaponicsv1.TankData{}
	}

	inputId, err := handler.systemInputManager.CreateSystemInput(ctx, si.Build())
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(ponixv1.CreateSystemInputResponse_builder{
		SystemInputId: inputId,
	}.Build())

	return resp, nil
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
