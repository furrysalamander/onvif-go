# onvif-go

A pure-Go, specification-driven implementation of the
[ONVIF](https://www.onvif.org/) network video interface — for both ONVIF
**clients** (talk to real cameras) and **servers** (mock cameras for testing,
and the building blocks for future real-camera firmware use).

This project is built from the ONVIF WSDL/XSD schemas (vendored locally) using
a custom Go code generator. The generated types are committed to the repository
so consumers can `go get` the library without running the generator.

## Status

See [PROGRESS.md](PROGRESS.md) for the milestone plan and current status.

## Services (Phase 1)

| Service | Spec | Package |
|---|---|---|
| Device Management | ver10 | `onvif/devicemgmt` |
| Media v1 | ver10 | `onvif/media` |
| Media v2 | ver20 | `onvif/media2` |
| PTZ | ver10/ver20 | `onvif/ptz` |
| Events (WS-Notification) | ver10 | `onvif/events` |
| Imaging | ver10/ver20 | `onvif/imaging` |

## Layout

```
cmd/                  onvifgen (codegen), onvif-cli (client), mockcam (mock server)
internal/             ws (SOAP), wsdiscovery, wssecurity, catalog (vendored XSDs)
onvif/                public API: per-service client facades + generated schema
testdata/             golden XML, fixtures
```

## Licensing

MIT — see [LICENSE](LICENSE). The vendored ONVIF WSDL/XSD schemas remain
property of their respective copyright holders under ONVIF's own notice; see
`internal/catalog/NOTICE.md`.