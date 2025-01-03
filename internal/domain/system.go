package domain

import (
	"context"

	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemStorer interface {
	CreateSystem(ctx context.Context, system *ponixv1.System) error
	System(ctx context.Context, systemId string) (*ponixv1.System, error)
}

type SystemManager struct {
	stringId    StringId
	systemStore SystemStorer
}

func NewSystemManager(ss SystemStorer, sid StringId) *SystemManager {
	return &SystemManager{
		systemStore: ss,
		stringId:    sid,
	}
}

func (mgr *SystemManager) CreateSystem(ctx context.Context, system *ponixv1.System) (string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystem")
	defer span.End()

	if system.GetId() == "" {
		system.SetId(mgr.stringId())
	}

	system.Status = ponixv1.SystemStatus_SYSTEM_STATUS_PENDING

	err := mgr.systemStore.CreateSystem(ctx, system)
	if err != nil {
		return "", err
	}

	return system.Id, nil
}

func (mgr *SystemManager) System(ctx context.Context, systemId string) (*ponixv1.System, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "System")
	defer span.End()

	system, err := mgr.systemStore.System(ctx, systemId)
	if err != nil {
		return nil, err
	}

	return system, nil
}
