# internal/catalog — vendored schema provenance + checksums

This directory vendors the exact WSDL/XSD artifacts that `cmd/onvifgen` parses
to generate Go code. Files are committed verbatim; their integrity is checked
at test time by `internal/catalog.TestChecksums` against `checksums.txt`.

## Editing policy

- **Do not edit** any file under `onvif/` or `external/`. They must remain
  byte-identical to the upstream artifacts.
- `catalog.xml` and `checksums.txt` are hand/generated-maintained.
- To refresh the checksums after adding a schema:
  ```bash
  cd internal/catalog
  find onvif external -type f \( -name '*.wsdl' -o -name '*.xsd' \) | sort | \
    sed 's|^\./||' | xargs sha256sum | sed 's|  \./|  |' > checksums.txt
  ```

## Provenance

### ONVIF (`onvif/`)

Source: `https://www.onvif.org/...`

| Path | Source URL |
|---|---|
| `onvif/ver10/device/wsdl/devicemgmt.wsdl` | https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl |
| `onvif/ver10/media/wsdl/media.wsdl` | https://www.onvif.org/ver10/media/wsdl/media.wsdl |
| `onvif/ver20/media/wsdl/media.wsdl` | https://www.onvif.org/ver20/media/wsdl/media.wsdl |
| `onvif/ver20/ptz/wsdl/ptz.wsdl` | https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl |
| `onvif/ver10/events/wsdl/event.wsdl` | https://www.onvif.org/ver10/events/wsdl/event.wsdl |
| `onvif/ver20/imaging/wsdl/imaging.wsdl` | https://www.onvif.org/ver20/imaging/wsdl/imaging.wsdl |
| `onvif/ver10/schema/onvif.xsd` | https://www.onvif.org/ver10/schema/onvif.xsd |
| `onvif/ver10/schema/common.xsd` | https://www.onvif.org/ver10/schema/common.xsd |

ONVIF specs are © ONVIF; use is governed by ONVIF's notice embedded at the
top of each WSDL/XSD file (no license to modify).

### OASIS support schemas (`external/oasis/`)

| Path | Source URL |
|---|---|
| `external/oasis/wsn/b-2.xsd` | https://docs.oasis-open.org/wsn/b-2.xsd |
| `external/oasis/wsn/t-1.xsd` | https://docs.oasis-open.org/wsn/t-1.xsd |
| `external/oasis/wsn/bw-2.wsdl` | https://docs.oasis-open.org/wsn/bw-2.wsdl |
| `external/oasis/wsrf/rw-2.wsdl` | https://docs.oasis-open.org/wsrf/rw-2.wsdl |
| `external/oasis/wsrf/bf-2.xsd` | https://docs.oasis-open.org/wsrf/bf-2.xsd |
| `external/oasis/wsrf/r-2.xsd` | https://docs.oasis-open.org/wsrf/r-2.xsd |

OASIS standards are governed by the OASIS IPR policy; see each file header
for copyright terms.

### W3C support schemas (`external/w3c/`)

| Path | Source URL |
|---|---|
| `external/w3c/2005/08/addressing/ws-addr.xsd` | https://www.w3.org/2005/08/addressing/ws-addr.xsd |
| `external/w3c/2005/05/xmlmime/xmlmime.xsd` | https://www.w3.org/2005/05/xmlmime |
| `external/w3c/2003/05/soap-envelope/soap-envelope.xsd` | https://www.w3.org/2003/05/soap-envelope |
| `external/w3c/2004/08/xop/include/xop-include.xsd` | https://www.w3.org/2004/08/xop/include |
| `external/w3c/2001/xml.xsd` | https://www.w3.org/2001/xml.xsd |

W3C documents are governed by the W3C Document License; see each file header.

## License — library code

The Go code in `internal/catalog/*.go` is part of `onvif-go` and is MIT
licensed (see the repository LICENSE). The vendored schema artifacts retain
their original copyright notices (embedded in each file).