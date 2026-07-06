package testutil

import (
	"context"
	"testing"

	"github.com/furrysalamander/onvif-go/onvif"
	"github.com/furrysalamander/onvif-go/onvif/schema/tt"
	"github.com/furrysalamander/onvif-go/onvif/svc/devicemgmt"
	"github.com/furrysalamander/onvif-go/onvif/svc/events"
	"github.com/furrysalamander/onvif-go/onvif/svc/imaging"
	"github.com/furrysalamander/onvif-go/onvif/svc/media"
	"github.com/furrysalamander/onvif-go/onvif/svc/ptz"
)

func boolPtr(v bool) *bool { return &v }

func TestRoundTrip_GetDeviceInformation(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := devicemgmt.NewClient(addr, "admin", "password")
		ctx := context.Background()
		info, err := c.GetDeviceInformation(ctx)
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
		ctx := context.Background()
		svcs, err := c.GetServices(ctx, false)
		if err != nil {
			t.Fatalf("GetServices: %v", err)
		}
		if len(svcs.Service) < 3 {
			t.Errorf("expected at least 3 services, got %d", len(svcs.Service))
		}
		t.Logf("round-trip GetServices: OK")
	})
}

func TestRoundTrip_DeviceFacade(t *testing.T) {
	WithMockServer(t, func(addr string) {
		d := onvif.NewDevice(addr, "admin", "password")
		ctx := context.Background()

		info, err := d.GetInfo(ctx)
		if err != nil {
			t.Fatalf("GetInfo: %v", err)
		}
		if info.Manufacturer != "MockONVIF" {
			t.Errorf("Manufacturer = %q, want MockONVIF", info.Manufacturer)
		}

		svcs, err := d.GetServices(ctx, false)
		if err != nil {
			t.Fatalf("GetServices: %v", err)
		}
		if len(svcs) < 3 {
			t.Errorf("expected at least 3 services, got %d", len(svcs))
		}

		t.Logf("round-trip Device facade: OK")
	})
}

func TestRoundTrip_Media(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := media.NewClient(addr, "admin", "password")
		ctx := context.Background()

		profiles, err := c.GetProfiles(ctx)
		if err != nil {
			t.Fatalf("GetProfiles: %v", err)
		}
		if len(profiles.Profiles) == 0 {
			t.Error("expected at least 1 profile")
		}
		t.Logf("Media.GetProfiles: %d profiles", len(profiles.Profiles))

		_, err = c.GetVideoSources(ctx)
		if err != nil {
			t.Fatalf("GetVideoSources: %v", err)
		}
		t.Logf("Media.GetVideoSources: OK")

		stream, err := c.GetStreamUri(ctx, tt.StreamSetup{}, tt.ReferenceToken("profile1"))
		if err != nil {
			t.Fatalf("GetStreamUri: %v", err)
		}
		t.Logf("Media.GetStreamUri: %s", stream.MediaUri.Uri)

		_, err = c.GetSnapshotUri(ctx, tt.ReferenceToken("profile1"))
		if err != nil {
			t.Fatalf("GetSnapshotUri: %v", err)
		}
		t.Logf("Media.GetSnapshotUri: OK")
	})
}

func TestRoundTrip_PTZ(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := ptz.NewClient(addr, "admin", "password")
		ctx := context.Background()

		nodes, err := c.GetNodes(ctx)
		if err != nil {
			t.Fatalf("GetNodes: %v", err)
		}
		t.Logf("PTZ.GetNodes: %d nodes", len(nodes.PTZNode))

		_, err = c.GetStatus(ctx, tt.ReferenceToken("profile1"))
		if err != nil {
			t.Fatalf("GetStatus: %v", err)
		}
		t.Logf("PTZ.GetStatus: OK")

		_, err = c.AbsoluteMove(ctx, tt.ReferenceToken("profile1"), tt.PTZVector{}, nil)
		if err != nil {
			t.Fatalf("AbsoluteMove: %v", err)
		}
		t.Logf("PTZ.AbsoluteMove: OK")

		_, err = c.ContinuousMove(ctx, tt.ReferenceToken("profile1"), tt.PTZSpeed{}, nil)
		if err != nil {
			t.Fatalf("ContinuousMove: %v", err)
		}
		t.Logf("PTZ.ContinuousMove: OK")

		_, err = c.Stop(ctx, tt.ReferenceToken("profile1"), boolPtr(false), boolPtr(false))
		if err != nil {
			t.Fatalf("Stop: %v", err)
		}
		t.Logf("PTZ.Stop: OK")

		_, err = c.GetPresets(ctx, tt.ReferenceToken("profile1"))
		if err != nil {
			t.Fatalf("GetPresets: %v", err)
		}
		t.Logf("PTZ.GetPresets: OK")
	})
}

func TestRoundTrip_Events(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := events.NewClient(addr, "admin", "password")
		ctx := context.Background()

		_, err := c.CreatePullPointSubscription(ctx, nil, nil)
		if err != nil {
			t.Fatalf("CreatePullPointSubscription: %v", err)
		}
		t.Logf("Events.CreatePullPointSubscription: OK")

		_, err = c.PullMessages(ctx, nil, 1024)
		if err != nil {
			t.Fatalf("PullMessages: %v", err)
		}
		t.Logf("Events.PullMessages: OK")
	})
}

func TestRoundTrip_Imaging(t *testing.T) {
	WithMockServer(t, func(addr string) {
		c := imaging.NewClient(addr, "admin", "password")
		ctx := context.Background()

		_, err := c.GetImagingSettings(ctx, tt.ReferenceToken("videosource1"))
		if err != nil {
			t.Fatalf("GetImagingSettings: %v", err)
		}
		t.Logf("Imaging.GetImagingSettings: OK")

		_, err = c.GetMoveOptions(ctx, tt.ReferenceToken("videosource1"))
		if err != nil {
			t.Fatalf("GetMoveOptions: %v", err)
		}
		t.Logf("Imaging.GetMoveOptions: OK")

		_, err = c.GetOptions(ctx, tt.ReferenceToken("videosource1"))
		if err != nil {
			t.Fatalf("GetOptions: %v", err)
		}
		t.Logf("Imaging.GetOptions: OK")
	})
}
