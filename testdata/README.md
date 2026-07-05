# testdata

Test fixtures used by the test suite.

## Layout (planned)

- `golden/` — golden XML files for request marshal / response demarshal /
  fault assertions, refreshed with `go test -run TestGolden -update`.
- `fixtures/` — canned ONVIF responses captured from real devices or built by
  hand from the spec.

Populated from M3 onward.