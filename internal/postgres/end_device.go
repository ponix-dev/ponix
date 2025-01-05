package postgres

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type EndDeviceStore struct {
	db *sqlc.Queries
}

func NewEndDeviceStore(db *sqlc.Queries) *EndDeviceStore {
	return &EndDeviceStore{
		db: db,
	}
}

func (store *EndDeviceStore) CreateEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	params := sqlc.CreateEndDeviceParams{
		ID:              endDevice.GetId(),
		SystemID:        endDevice.GetSystemId(),
		NetworkServerID: endDevice.GetNetworkServerId(),
		SystemInputID:   endDevice.GetSystemInputId(),
		Name:            endDevice.GetName(),
		Status:          int32(endDevice.GetStatus()),
	}

	_, err := store.db.CreateEndDevice(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (store *EndDeviceStore) SystemEndDevices(ctx context.Context, systemId string) ([]*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemEndDevices")
	defer span.End()

	dbeds, err := store.db.GetSystemEndDevices(ctx, systemId)
	if err != nil {
		return nil, err
	}

	endDevices := make([]*iotv1.EndDevice, len(dbeds))

	for i, dbed := range dbeds {
		edb := iotv1.EndDevice_builder{
			Id:              dbed.ID,
			NetworkServerId: dbed.NetworkServerID,
			SystemId:        dbed.SystemID,
			Name:            dbed.Name,
			SystemInputId:   dbed.SystemInputID,
			Status:          iotv1.EndDeviceStatus(dbed.Status),
		}

		endDevices[i] = edb.Build()
	}

	return endDevices, nil
}
