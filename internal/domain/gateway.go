package domain

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type GatewayStorer interface {
	CreateGateway(ctx context.Context, gateway *iotv1.Gateway) error
	SystemGateways(ctx context.Context, systemId string) ([]*iotv1.Gateway, error)
}

type GatewayManager struct {
	gatewayStore GatewayStorer
	stringId     StringId
	validate     Validate
}

func NewGatewayManager(gs GatewayStorer, stringId StringId, validate Validate) *GatewayManager {
	return &GatewayManager{
		gatewayStore: gs,
		stringId:     stringId,
		validate:     validate,
	}
}

func (mgr *GatewayManager) CreateGateway(ctx context.Context, gateway *iotv1.Gateway) (string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateGateway")
	defer span.End()

	if gateway.GetId() == "" {
		gateway.SetId(mgr.stringId())
	}

	gateway.Status = iotv1.GatewayStatus_GATEWAY_STATUS_PENDING

	err := mgr.validate(gateway)
	if err != nil {
		return "", err
	}

	err = mgr.gatewayStore.CreateGateway(ctx, gateway)
	if err != nil {
		return "", err
	}

	return gateway.Id, nil
}

func (mgr *GatewayManager) SystemGateways(ctx context.Context, systemId string) ([]*iotv1.Gateway, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemGateways")
	defer span.End()

	servers, err := mgr.gatewayStore.SystemGateways(ctx, systemId)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
