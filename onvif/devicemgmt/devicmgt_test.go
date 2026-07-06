package devicemgmt

import (
	"os"
	"testing"
	"time"

	"github.com/furrysalamander/onvif-go/internal/ws"
	"github.com/furrysalamander/onvif-go/internal/wssecurity"
	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
)

const goldenDir = "../../testdata/golden/devicemgmt"

var fixedNonce = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

func TestGetDeviceInformationRequestGolden(t *testing.T) {
	ut := wssecurity.NewUsernameTokenDeterministic("admin", "password", fixedNonce, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	action := ws.NewAction("http://www.onvif.org/ver10/device/wsdl/GetDeviceInformation")

	data, err := ws.MarshalRequest("", &tds.GetDeviceInformation{}, ut, action)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	path := goldenDir + "/get_device_information_request.xml"
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
		t.Errorf("request mismatch:\n got: %s\nwant: %s", data, want)
	}
}

func TestGetDeviceInformationResponseGolden(t *testing.T) {
	path := goldenDir + "/get_device_information_response.xml"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	var res tds.GetDeviceInformationResponse
	fault, err := ws.UnmarshalResponse(data, &res)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if fault != nil {
		t.Fatalf("unexpected fault: %v", fault)
	}

	if res.Manufacturer != "TestCorp" {
		t.Errorf("Manufacturer = %q, want %q", res.Manufacturer, "TestCorp")
	}
	if res.Model != "TestCam" {
		t.Errorf("Model = %q, want %q", res.Model, "TestCam")
	}
}
