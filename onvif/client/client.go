// Package client provides ONVIF device discovery and the Device type.
//
// Use Discover to find ONVIF devices on the local network via WS-Discovery
// multicast probe. Once discovered, connect using the per-service client
// packages (devicemgmt, media, media2, ptz, events, imaging).
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
