package postgres

import (
	"context"

	aquaponicsv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/aquaponics/v1"
	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	soilponicsv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/soilponics/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type SystemInputStore struct {
	transactioner *pgxpool.Pool
	db            *sqlc.Queries
}

func NewSystemInputStore(db *sqlc.Queries) *SystemInputStore {
	return &SystemInputStore{
		db: db,
	}
}

func (store *SystemInputStore) CreateSystemInput(ctx context.Context, systemInput *ponixv1.SystemInput) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateSystemInput")
	defer span.End()

	switch systemInput.InputData.(type) {
	case *ponixv1.SystemInput_Field:
		return store.createField(ctx, systemInput, systemInput.GetField())
	case *ponixv1.SystemInput_GrowMedium:
		return store.createGrowMedium(ctx, systemInput, systemInput.GetGrowMedium())
	case *ponixv1.SystemInput_Tank:
		return store.createTank(ctx, systemInput, systemInput.GetTank())
	}

	return nil
}

func (store *SystemInputStore) createField(ctx context.Context, systemInput *ponixv1.SystemInput, fieldData *soilponicsv1.FieldData) error {
	tx, err := store.transactioner.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := store.db.WithTx(tx)

	_, err = qtx.CreateField(ctx, systemInput.Id)
	if err != nil {
		return err
	}

	fsparams := sqlc.CreateFieldSystemInputParams{
		ID:       systemInput.GetId(),
		Name:     systemInput.GetName(),
		SystemID: systemInput.GetSystemId(),
		Status:   int32(systemInput.GetStatus()),
	}

	_, err = qtx.CreateFieldSystemInput(ctx, fsparams)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (store *SystemInputStore) createGrowMedium(ctx context.Context, systemInput *ponixv1.SystemInput, gmData *aquaponicsv1.GrowMediumData) error {
	tx, err := store.transactioner.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := store.db.WithTx(tx)

	gmparams := sqlc.CreateGrowMediumParams{
		ID:         systemInput.Id,
		MediumType: int32(gmData.GetMediumType()),
	}

	_, err = qtx.CreateGrowMedium(ctx, gmparams)
	if err != nil {
		return err
	}

	fsparams := sqlc.CreateGrowMediumSystemInputParams{
		ID:       systemInput.GetId(),
		Name:     systemInput.GetName(),
		SystemID: systemInput.GetSystemId(),
		Status:   int32(systemInput.GetStatus()),
	}

	_, err = qtx.CreateGrowMediumSystemInput(ctx, fsparams)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (store *SystemInputStore) createTank(ctx context.Context, systemInput *ponixv1.SystemInput, tankData *aquaponicsv1.TankData) error {
	tx, err := store.transactioner.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := store.db.WithTx(tx)

	_, err = qtx.CreateTank(ctx, systemInput.Id)
	if err != nil {
		return err
	}

	fsparams := sqlc.CreateTankSystemInputParams{
		ID:       systemInput.GetId(),
		Name:     systemInput.GetName(),
		SystemID: systemInput.GetSystemId(),
		Status:   int32(systemInput.GetStatus()),
	}

	_, err = qtx.CreateTankSystemInput(ctx, fsparams)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (store *SystemInputStore) SystemInput(ctx context.Context, systemInputId string) (*ponixv1.SystemInput, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInput")
	defer span.End()

	row, err := store.db.GetSystemInput(ctx, systemInputId)
	if err != nil {
		return nil, err
	}

	sib := ponixv1.SystemInput_builder{
		Id:       row.ID,
		Name:     row.Name,
		SystemId: row.SystemID,
		Status:   ponixv1.SystemInputStatus(row.Status),
	}

	switch {
	case row.GrowMedium != (sqlc.GrowMedium{}):
		sib.GrowMedium = &aquaponicsv1.GrowMediumData{
			MediumType: aquaponicsv1.MediumType(row.GrowMedium.MediumType),
		}
	case row.Field != (sqlc.Field{}):
		sib.Field = &soilponicsv1.FieldData{}
	case row.Tank != (sqlc.Tank{}):
		sib.Tank = &aquaponicsv1.TankData{}
	}

	return sib.Build(), nil
}

func (store *SystemInputStore) SystemInputs(ctx context.Context, systemId string) ([]*ponixv1.SystemInput, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemInputs")
	defer span.End()

	rows, err := store.db.GetSystemInputs(ctx, systemId)
	if err != nil {
		return nil, err
	}

	systemInputs := make([]*ponixv1.SystemInput, len(rows))

	for i, row := range rows {
		sib := ponixv1.SystemInput_builder{
			Id:       row.ID,
			Name:     row.Name,
			SystemId: row.SystemID,
			Status:   ponixv1.SystemInputStatus(row.Status),
		}

		switch {
		case row.GrowMedium != (sqlc.GrowMedium{}):
			sib.GrowMedium = &aquaponicsv1.GrowMediumData{
				MediumType: aquaponicsv1.MediumType(row.GrowMedium.MediumType),
			}
		case row.Field != (sqlc.Field{}):
			sib.Field = &soilponicsv1.FieldData{}
		case row.Tank != (sqlc.Tank{}):
			sib.Tank = &aquaponicsv1.TankData{}
		}

		systemInputs[i] = sib.Build()
	}

	return systemInputs, nil
}
