package postgres

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type NetworkServerStore struct {
	db *sqlc.Queries
}

func NewNetworkServerStore(db *sqlc.Queries) *NetworkServerStore {
	return &NetworkServerStore{
		db: db,
	}
}

func (store *NetworkServerStore) CreateNetworkServer(ctx context.Context, networkServer *iotv1.NetworkServer) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateNetworkServer")
	defer span.End()

	params := sqlc.CreateNetworkServerParams{
		ID:          networkServer.GetId(),
		SystemID:    networkServer.GetSystemId(),
		Name:        networkServer.GetName(),
		Status:      int32(networkServer.GetStatus()),
		IotPlatform: int32(networkServer.GetIotPlatform()),
	}

	_, err := store.db.CreateNetworkServer(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (store *NetworkServerStore) SystemNetworkServers(ctx context.Context, systemId string) ([]*iotv1.NetworkServer, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemNetworkServers")
	defer span.End()

	dbnss, err := store.db.GetSystemNetworkServers(ctx, systemId)
	if err != nil {
		return nil, err
	}

	networkServers := make([]*iotv1.NetworkServer, len(dbnss))

	for i, dbns := range dbnss {
		nsb := iotv1.NetworkServer_builder{
			Id:          dbns.ID,
			Name:        dbns.Name,
			SystemId:    dbns.SystemID,
			IotPlatform: iotv1.IOTPlatform(dbns.IotPlatform),
			Status:      iotv1.NetworkServerStatus(dbns.Status),
		}

		networkServers[i] = nsb.Build()
	}

	return networkServers, nil
}
