package wsdiscovery

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/furrysalamander/onvif-go/onvif/schema/env"
)

const goldenDir = "../../testdata/golden/wsdiscovery"

func TestProbeMarshalGolden(t *testing.T) {
	e := &env.Envelope{}
	if err := e.SetBody(&Probe{Types: "dn:NetworkVideoTransmitter"}); err != nil {
		t.Fatal(err)
	}

	data, err := xml.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}

	path := goldenDir + "/probe_request.xml"
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote golden: %s", path)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run with UPDATE_GOLDEN=1 to regenerate)", err)
	}
	if string(data) != string(want) {
		t.Errorf("probe marshal mismatch:\n got: %s\nwant: %s", data, want)
	}
}
