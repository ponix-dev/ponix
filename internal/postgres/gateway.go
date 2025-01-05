package postgres

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type GatewayStore struct {
	db *sqlc.Queries
}

func NewGatewayStore(db *sqlc.Queries) *GatewayStore {
	return &GatewayStore{
		db: db,
	}
}
func (store *GatewayStore) CreateGateway(ctx context.Context, gateway *iotv1.Gateway) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateGateway")
	defer span.End()

	params := sqlc.CreateGatewayParams{
		ID:              gateway.GetId(),
		SystemID:        gateway.GetSystemId(),
		NetworkServerID: gateway.GetNetworkServerId(),
		Name:            gateway.GetName(),
		Status:          int32(gateway.GetStatus()),
	}

	_, err := store.db.CreateGateway(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (store *GatewayStore) SystemGateways(ctx context.Context, systemId string) ([]*iotv1.Gateway, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemGateways")
	defer span.End()

	dbgs, err := store.db.GetSystemGateways(ctx, systemId)
	if err != nil {
		return nil, err
	}

	gateways := make([]*iotv1.Gateway, len(dbgs))

	for i, dbg := range dbgs {

		gb := iotv1.Gateway_builder{
			Id:              dbg.ID,
			Name:            dbg.Name,
			NetworkServerId: dbg.NetworkServerID,
			SystemId:        dbg.SystemID,
			Status:          iotv1.GatewayStatus(dbg.Status),
		}

		gateways[i] = gb.Build()
	}

	return gateways, nil
}
