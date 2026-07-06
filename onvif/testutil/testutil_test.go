package testutil

import (
	"testing"

	"github.com/furrysalamander/onvif-go/onvif/devicemgmt"
)

func TestRoundTrip_GetDeviceInformation(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := devicemgmt.NewClient(addr, "admin", "password")
		info, err := c.GetDeviceInformation()
		if err != nil {
			t.Fatalf("GetDeviceInformation: %v", err)
		}
		if info.Manufacturer != "MockONVIF" {
			t.Errorf("Manufacturer = %q, want MockONVIF", info.Manufacturer)
		}
		if info.Model != "MockCam-3000" {
			t.Errorf("Model = %q, want MockCam-3000", info.Model)
		}
		if info.FirmwareVersion != "1.0.0" {
			t.Errorf("FirmwareVersion = %q, want 1.0.0", info.FirmwareVersion)
		}
		t.Logf("round-trip GetDeviceInformation: OK")
	})
}

func TestRoundTrip_GetServices(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := devicemgmt.NewClient(addr, "admin", "password")
		svcs, err := c.GetServices()
		if err != nil {
			t.Fatalf("GetServices: %v", err)
		}
		if len(svcs.Service) < 3 {
			t.Errorf("expected at least 3 services, got %d", len(svcs.Service))
		}
		t.Logf("round-trip GetServices: OK")
	})
}
