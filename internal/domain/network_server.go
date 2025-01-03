package domain

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type NetworkServerStorer interface {
	SystemNetworkServers(ctx context.Context, systemId string) ([]*iotv1.NetworkServer, error)
	CreateNetworkServer(ctx context.Context, networkServer *iotv1.NetworkServer) error
}

type NetworkServerManager struct {
	networkServerStore NetworkServerStorer
	stringId           StringId
	validate           Validate
}

func NewNetworkServerManager(nss NetworkServerStorer, stringId StringId, validate Validate) *NetworkServerManager {
	return &NetworkServerManager{
		networkServerStore: nss,
		stringId:           stringId,
		validate:           validate,
	}
}

func (mgr *NetworkServerManager) CreateNetworkServer(ctx context.Context, networkServer *iotv1.NetworkServer) (string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateNetworkServer")
	defer span.End()

	if networkServer.GetId() == "" {
		networkServer.SetId(mgr.stringId())
	}

	networkServer.SetStatus(iotv1.NetworkServerStatus_NETWORK_SERVER_STATUS_PENDING)

	err := mgr.validate(networkServer)
	if err != nil {
		return "", err
	}

	err = mgr.networkServerStore.CreateNetworkServer(ctx, networkServer)
	if err != nil {
		return "", err
	}

	return networkServer.Id, nil
}

func (mgr *NetworkServerManager) SystemNetworkServers(ctx context.Context, systemId string) ([]*iotv1.NetworkServer, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemNetworkServers")
	defer span.End()

	servers, err := mgr.networkServerStore.SystemNetworkServers(ctx, systemId)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
