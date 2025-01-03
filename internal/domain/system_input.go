package domain

import (
	"context"

	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemInputStorer interface {
	CreateSystemInput(ctx context.Context, systemInput *ponixv1.SystemInput) error
	SystemInput(ctx context.Context, systemInputId string) (*ponixv1.SystemInput, error)
	SystemInputs(ctx context.Context, systemId string) ([]*ponixv1.SystemInput, error)
}

type SystemInputManager struct {
	stringId         StringId
	validate         Validate
	systemInputStore SystemInputStorer
}

func NewSystemInputManager(ss SystemInputStorer, sid StringId, validate Validate) *SystemInputManager {
	return &SystemInputManager{
		systemInputStore: ss,
		stringId:         sid,
		validate:         validate,
	}
}

func (mgr *SystemInputManager) CreateSystemInput(ctx context.Context, systemInput *ponixv1.SystemInput) (string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystemInput")
	defer span.End()

	if systemInput.GetId() == "" {
		systemInput.SetId(mgr.stringId())
	}

	systemInput.SetStatus(ponixv1.SystemInputStatus_SYSTEM_INPUT_STATUS_PENDING)

	err := mgr.validate(systemInput)
	if err != nil {
		return "", err
	}

	err = mgr.systemInputStore.CreateSystemInput(ctx, systemInput)
	if err != nil {
		return "", err
	}

	return systemInput.Id, nil
}

func (mgr *SystemInputManager) SystemInput(ctx context.Context, systemInputId string) (*ponixv1.SystemInput, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInput")
	defer span.End()

	systemInput, err := mgr.systemInputStore.SystemInput(ctx, systemInputId)
	if err != nil {
		return nil, err
	}

	return systemInput, nil
}

func (mgr *SystemInputManager) SystemInputs(ctx context.Context, systemId string) ([]*ponixv1.SystemInput, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInputs")
	defer span.End()

	systemInputs, err := mgr.systemInputStore.SystemInputs(ctx, systemId)
	if err != nil {
		return nil, err
	}

	return systemInputs, nil
}
