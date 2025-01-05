package postgres

import (
	"context"

	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemStore struct {
	db *sqlc.Queries
}

func NewSystemStore(db *sqlc.Queries) *SystemStore {
	return &SystemStore{
		db: db,
	}
}

func (store *SystemStore) CreateSystem(ctx context.Context, system *ponixv1.System) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystem")
	defer span.End()

	params := sqlc.CreateSystemParams{
		ID:             system.GetId(),
		OrganizationID: system.GetOrganizationId(),
		Name:           system.GetName(),
		Status:         int32(system.GetStatus().Number()),
	}

	_, err := store.db.CreateSystem(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (store *SystemStore) System(ctx context.Context, systemId string) (*ponixv1.System, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "System")
	defer span.End()

	dbSystem, err := store.db.GetSystem(ctx, systemId)
	if err != nil {
		return nil, err
	}

	sb := ponixv1.System_builder{
		Id:             dbSystem.ID,
		Name:           dbSystem.Name,
		OrganizationId: dbSystem.OrganizationID,
		Status:         ponixv1.SystemStatus(dbSystem.Status),
	}

	return sb.Build(), nil
}
