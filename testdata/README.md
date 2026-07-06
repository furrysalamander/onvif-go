# testdata

Test fixtures for golden-file and round-trip tests.

## Layout

```
golden/
  devicemgmt/
    get_device_information_request.xml
    get_device_information_response.xml
  wsdiscovery/
    probe_request.xml
  events/
```

## Regenerating golden files

When marshaling output changes, regenerate golden files with:

```bash
UPDATE_GOLDEN=1 go test ./onvif/devicemgmt/
UPDATE_GOLDEN=1 go test ./internal/wsdiscovery/
```

Review the diff and commit the updated golden files.
