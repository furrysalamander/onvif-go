package mockcam

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/furrysalamander/onvif-go/onvif/schema/env"
	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
	"github.com/furrysalamander/onvif-go/onvif/schema/tev"
	"github.com/furrysalamander/onvif-go/onvif/schema/timg"
	"github.com/furrysalamander/onvif-go/onvif/schema/tptz"
	"github.com/furrysalamander/onvif-go/onvif/schema/trt"
	"github.com/furrysalamander/onvif-go/onvif/schema/tt"
)

type Server struct {
	http.Server
	Addr  string
	Ready chan struct{}

	DeviceInfo   tds.GetDeviceInformationResponse
	Services     tds.GetServicesResponse
	Capabilities tds.GetCapabilitiesResponse
	Scopes       []tt.Scope
}

func New() *Server {
	mux := http.NewServeMux()
	s := &Server{
		Server: http.Server{Handler: mux},
		Ready:  make(chan struct{}),
		DeviceInfo: tds.GetDeviceInformationResponse{
			Manufacturer:    "MockONVIF",
			Model:           "MockCam-3000",
			FirmwareVersion: "1.0.0",
			SerialNumber:    "MOCK-001",
			HardwareId:      "HW-MOCK-001",
		},
		Services: tds.GetServicesResponse{
			Service: []tds.Service{
				{Namespace: "http://www.onvif.org/ver10/device/wsdl", XAddr: ""},
				{Namespace: "http://www.onvif.org/ver10/media/wsdl", XAddr: ""},
				{Namespace: "http://www.onvif.org/ver20/media/wsdl", XAddr: ""},
				{Namespace: "http://www.onvif.org/ver10/events/wsdl", XAddr: ""},
				{Namespace: "http://www.onvif.org/ver20/ptz/wsdl", XAddr: ""},
				{Namespace: "http://www.onvif.org/ver20/imaging/wsdl", XAddr: ""},
			},
		},
		Capabilities: tds.GetCapabilitiesResponse{
			Capabilities: tt.Capabilities{},
		},
	}
	mux.HandleFunc("/", s.handle)
	return s
}

func (s *Server) Listen() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.Addr = fmt.Sprintf("http://%s/onvif/device_service", ln.Addr().String())
	close(s.Ready)
	err = s.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var e env.Envelope
	if err := xml.Unmarshal(body, &e); err != nil {
		http.Error(w, "invalid SOAP", http.StatusBadRequest)
		return
	}

	resp := s.dispatch(body)
	respEnv := &env.Envelope{}
	_ = respEnv.SetBody(resp)
	out, _ := xml.Marshal(respEnv)
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	_, _ = w.Write(out)
}

func (s *Server) dispatch(raw []byte) interface{} {
	bodyStr := string(raw)
	switch {
	case strings.Contains(bodyStr, "GetDeviceInformation"):
		return &s.DeviceInfo
	case strings.Contains(bodyStr, "GetServices"):
		return s.getServices()
	case strings.Contains(bodyStr, "GetCapabilities"):
		return &s.Capabilities
	case strings.Contains(bodyStr, "GetScopes"):
		return s.getScopes()
	case strings.Contains(bodyStr, "GetSystemDateAndTime"):
		return &tds.GetSystemDateAndTimeResponse{}

	case strings.Contains(bodyStr, "GetProfiles"):
		return s.getProfiles()
	case strings.Contains(bodyStr, "GetVideoSources"):
		return &trt.GetVideoSourcesResponse{}
	case strings.Contains(bodyStr, "GetStreamUri"):
		return &trt.GetStreamUriResponse{
			MediaUri: tt.MediaUri{Uri: "rtsp://127.0.0.1/stream"},
		}
	case strings.Contains(bodyStr, "GetSnapshotUri"):
		return &trt.GetSnapshotUriResponse{
			MediaUri: tt.MediaUri{Uri: "http://127.0.0.1/snapshot"},
		}

	case strings.Contains(bodyStr, "GetNodes"):
		return &tptz.GetNodesResponse{}
	case strings.Contains(bodyStr, "GetStatus"):
		return &tptz.GetStatusResponse{}
	case strings.Contains(bodyStr, "AbsoluteMove"):
		return &tptz.AbsoluteMoveResponse{}
	case strings.Contains(bodyStr, "ContinuousMove"):
		return &tptz.ContinuousMoveResponse{}
	case strings.Contains(bodyStr, "RelativeMove"):
		return &tptz.RelativeMoveResponse{}
	case strings.Contains(bodyStr, "Stop"):
		return &tptz.StopResponse{}
	case strings.Contains(bodyStr, "GeoMove"):
		return &tptz.GeoMoveResponse{}
	case strings.Contains(bodyStr, "GotoHomePosition"):
		return &tptz.GotoHomePositionResponse{}
	case strings.Contains(bodyStr, "SetPreset"):
		return &tptz.SetPresetResponse{PresetToken: "preset-001"}
	case strings.Contains(bodyStr, "GetPresets"):
		return &tptz.GetPresetsResponse{}
	case strings.Contains(bodyStr, "CreatePresetTour"):
		return &tptz.CreatePresetTourResponse{}

	case strings.Contains(bodyStr, "GetEventProperties"):
		return &tev.GetEventPropertiesResponse{}
	case strings.Contains(bodyStr, "CreatePullPointSubscription"):
		return &tev.CreatePullPointSubscriptionResponse{}
	case strings.Contains(bodyStr, "PullMessages"):
		return &tev.PullMessagesResponse{}
	case strings.Contains(bodyStr, "Seek"):
		return &tev.SeekResponse{}
	case strings.Contains(bodyStr, "SetSynchronizationPoint"):
		return &tev.SetSynchronizationPointResponse{}

	case strings.Contains(bodyStr, "GetImagingSettings"):
		return &timg.GetImagingSettingsResponse{}
	case strings.Contains(bodyStr, "GetMoveOptions"):
		return &timg.GetMoveOptionsResponse{}
	case strings.Contains(bodyStr, "GetOptions"):
		return &timg.GetOptionsResponse{}
	}
	return &env.Fault{
		Code:   &env.FaultCode{Value: "env:Sender"},
		Reason: &env.FaultReason{Texts: []env.FaultText{{Lang: "en", Text: "unknown operation"}}},
	}
}

func (s *Server) getServices() *tds.GetServicesResponse {
	svcs := s.Services
	for i := range svcs.Service {
		if svcs.Service[i].XAddr == "" {
			svcs.Service[i].XAddr = strings.Replace(s.Addr, "device_service", svcs.Service[i].Namespace, 1)
		}
	}
	return &svcs
}

func (s *Server) getScopes() *tds.GetScopesResponse {
	scopes := s.Scopes
	if len(scopes) == 0 {
		scopes = []tt.Scope{}
	}
	return &tds.GetScopesResponse{Scopes: scopes}
}

func (s *Server) getProfiles() *trt.GetProfilesResponse {
	return &trt.GetProfilesResponse{
		Profiles: []tt.Profile{
			{
				Xtoken: tt.ReferenceToken("profile-001"),
				Name:   tt.Name("MainProfile"),
			},
		},
	}
}
