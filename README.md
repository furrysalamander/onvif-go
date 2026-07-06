# onvif-go

[![Go Reference](https://pkg.go.dev/badge/github.com/furrysalamander/onvif-go.svg)](https://pkg.go.dev/github.com/furrysalamander/onvif-go)

A pure-Go, specification-driven implementation of the
[ONVIF](https://www.onvif.org/) network video interface — for both ONVIF
**clients** (talk to real cameras) and **servers** (mock cameras for testing,
and the building blocks for future real-camera firmware use).

This project is built from the ONVIF WSDL/XSD schemas (vendored locally) using
a custom Go code generator. The generated types are committed to the repository
so consumers can `go get` the library without running the generator.

## Installation

```bash
go get github.com/furrysalamander/onvif-go
```

Requires Go 1.26+.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/furrysalamander/onvif-go/onvif/client"
	"github.com/furrysalamander/onvif-go/onvif/devicemgmt"
)

func main() {
	// 1. Discover cameras on the local network
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := client.Discover(ctx)
	if err != nil {
		panic(err)
	}
	for _, d := range devices {
		fmt.Printf("Found: %s (%s)\n", d.Endpoint, d.Types)
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return
	}

	// 2. Connect to the first device
	c := devicemgmt.NewClient(devices[0].Endpoint, "admin", "password")

	// 3. Get device information
	info, err := c.GetDeviceInformation()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Manufacturer: %s\n", info.Manufacturer)
	fmt.Printf("Model:        %s\n", info.Model)
	fmt.Printf("Firmware:     %s\n", info.FirmwareVersion)

	// 4. Get capabilities
	caps, err := c.GetCapabilities()
	if err != nil {
		panic(err)
	}
	_ = caps
}
```

See [examples/](examples/) for more complete programs.

## Services (Phase 1)

| Service | Spec | Package |
|---|---|---|
| Device Management | ver10 | `onvif/devicemgmt` |
| Media v1 | ver10 | `onvif/media` |
| Media v2 | ver20 | `onvif/media2` |
| PTZ | ver20 | `onvif/ptz` |
| Events (WS-Notification) | ver10 | `onvif/events` |
| Imaging | ver20 | `onvif/imaging` |

## Commands

| Command | Purpose |
|---|---|
| `cmd/onvifgen` | Code generator — run via `go generate ./...` |
| `cmd/onvif-cli` | Ad-hoc ONVIF client (discover + device info) |
| `cmd/mockcam` | Mock ONVIF camera server for testing |

## Layout

```
cmd/                  onvifgen (codegen), onvif-cli (client), mockcam (mock server)
internal/             ws (SOAP), wsdiscovery, wssecurity, catalog (vendored XSDs)
onvif/                public API: per-service client facades + generated schema
testdata/             golden XML, fixtures
docs/                 validation checklists (ODM)
examples/             usage examples
```

## Status

See [PROGRESS.md](PROGRESS.md) for the milestone plan and current status.

## Licensing

MIT — see [LICENSE](LICENSE). The vendored ONVIF WSDL/XSD schemas remain
property of their respective copyright holders under ONVIF's own notice; see
`internal/catalog/NOTICE.md`.
