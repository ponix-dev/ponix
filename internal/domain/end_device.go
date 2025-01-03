package domain

import (
	"context"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type EndDeviceStorer interface {
	CreateEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) error
	SystemEndDevices(ctx context.Context, systemId string) ([]*iotv1.EndDevice, error)
}

type EndDeviceManager struct {
	endDeviceStore EndDeviceStorer
	stringId       StringId
}

func NewEndDeviceManager(eds EndDeviceStorer, stringId StringId) *EndDeviceManager {
	return &EndDeviceManager{
		endDeviceStore: eds,
		stringId:       stringId,
	}
}

func (mgr *EndDeviceManager) CreateEndDevice(ctx context.Context, endDevice *iotv1.EndDevice) (string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateEndDevice")
	defer span.End()

	if endDevice.GetId() == "" {
		endDevice.SetId(mgr.stringId())
	}

	endDevice.Status = iotv1.EndDeviceStatus_END_DEVICE_STATUS_PENDING

	err := mgr.endDeviceStore.CreateEndDevice(ctx, endDevice)
	if err != nil {
		return "", err
	}

	return endDevice.Id, nil
}

func (mgr *EndDeviceManager) SystemEndDevices(ctx context.Context, systemId string) ([]*iotv1.EndDevice, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "SystemEndDevices")
	defer span.End()

	servers, err := mgr.endDeviceStore.SystemEndDevices(ctx, systemId)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
