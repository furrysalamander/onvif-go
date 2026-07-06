package client

import (
	"context"

	"github.com/furrysalamander/onvif-go/internal/wsdiscovery"
)

type Device = wsdiscovery.DeviceInfo

func Discover(ctx context.Context) ([]Device, error) {
	return wsdiscovery.Discover(ctx)
}

func DiscoverWithTypes(ctx context.Context, types string) ([]Device, error) {
	return wsdiscovery.DiscoverWithTypes(ctx, types)
}
