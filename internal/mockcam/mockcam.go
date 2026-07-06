package mockcam

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/furrysalamander/onvif-go/internal/ws"
	"github.com/furrysalamander/onvif-go/onvif/schema/env"
	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
	"github.com/furrysalamander/onvif-go/onvif/schema/tt"
	"github.com/furrysalamander/onvif-go/onvif/schema/wsa"
)

type Server struct {
	http.Server
	Addr string

	DeviceInfo   tds.GetDeviceInformationResponse
	Services     tds.GetServicesResponse
	Capabilities tds.GetCapabilitiesResponse
	Scopes       []tt.Scope
}

func New() *Server {
	mux := http.NewServeMux()
	s := &Server{
		Server: http.Server{Handler: mux},
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

type mockResponse struct {
	XMLName xml.Name
	Body    interface{}
}

func (s *Server) Listen() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.Addr = fmt.Sprintf("http://%s/onvif/device_service", ln.Addr().String())
	err = s.Server.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var e env.Envelope
	if err := xml.Unmarshal(body, &e); err != nil {
		http.Error(w, "invalid SOAP", http.StatusBadRequest)
		return
	}

	resp := s.dispatch(body, &e)
	respEnv := &env.Envelope{}
	respEnv.SetBody(resp)
	out, _ := xml.Marshal(respEnv)
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.Write(out)
}

func (s *Server) dispatch(raw []byte, e *env.Envelope) interface{} {
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

var _ wsa.AttributedURIType
var _ = time.Now()
var _ = log.Default()
var _ = ws.MarshalRequest
