# internal/catalog — vendored ONVIF + OASIS/W3C schemas

This directory holds the **codegen input** for `cmd/onvifgen`. Schemas are
vendored (downloaded once and committed) so the generator never needs network
access.

## Layout (planned, populated in M1)

```
internal/catalog/
  onvif/        ver10/ver20 WSDL + XSD from www.onvif.org
  external/     OASIS WS-Notification (bw-2, b-2, t-1), WSRF (rw-2),
                WS-Addressing (ws-addr.xsd), XML, SOAP
  catalog.xml   XML Catalog mapping namespace URIs to local files
  NOTICE.md     provenance + SHA-256 checksums of every vendored file
  catalog.go    Go API exposing a Resolver to onvifgen
```

## Editing policy

- **Do not edit** contents of `onvif/` or `external/` — they must remain
  verbatim copies of the upstream artifacts. Provenance + checksums in
  `NOTICE.md` are the integrity guarantee.
- `catalog.go` and `catalog.xml` are hand-maintained.