# ODM Validation Checklist

Manual validation of onvif-go against ONVIF Device Manager.

## Prerequisites

- Go 1.26+
- [ONVIF Device Manager](https://sourceforge.net/projects/onvifdm/)
- Built binaries: `go build ./cmd/mockcam ./cmd/onvif-cli`

## Mock Camera Validation

1. Start mockcam: `./cmd/mockcam/mockcam`
2. In ODM, add device at the printed URL
3. Credentials: `admin` / `password` (any accepted)

| Check | Expected |
|-------|----------|
| Manufacturer | MockONVIF |
| Model | MockCam-3000 |
| Firmware | 1.0.0 |
| SerialNumber | MOCK-001 |

## Per-Service Checks

### Device Management

| Operation | Verified |
|-----------|----------|
| GetDeviceInformation | [ ] |
| GetServices | [ ] |
| GetCapabilities | [ ] |
| GetScopes | [ ] |
| GetSystemDateAndTime | [ ] |

### Media v1 / v2

| Operation | Verified |
|-----------|----------|
| GetProfiles | [ ] |
| GetVideoSources | [ ] |
| GetVideoSourceConfigurations | [ ] |

### PTZ

| Operation | Verified |
|-----------|----------|
| GetNodes | [ ] |
| GetStatus | [ ] |

### Events

| Operation | Verified |
|-----------|----------|
| GetEventProperties | [ ] |
| CreatePullPointSubscription | [ ] |
| PullMessages | [ ] |

### Imaging

| Operation | Verified |
|-----------|----------|
| GetImagingSettings | [ ] |
| GetMoveOptions | [ ] |

## Discovery

```bash
./cmd/onvif-cli/onvif-cli discover
```

| Check | Verified |
|-------|----------|
| Mock camera discovered | [ ] |
| Types include NetworkVideoTransmitter | [ ] |
| XAddrs populated | [ ] |

## Automated Tests

```bash
go test -v ./onvif/testutil/
go test -v ./onvif/devicemgmt/
go test -v ./internal/wsdiscovery/
```
