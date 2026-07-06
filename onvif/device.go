package onvif

import (
	"context"
	"net/http"

	"github.com/furrysalamander/onvif-go/internal/wsdiscovery"
	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
	"github.com/furrysalamander/onvif-go/onvif/schema/trt"
	"github.com/furrysalamander/onvif-go/onvif/schema/tt"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
	"github.com/furrysalamander/onvif-go/onvif/svc/accesscontrol"
	"github.com/furrysalamander/onvif-go/onvif/svc/accessrules"
	"github.com/furrysalamander/onvif-go/onvif/svc/actionengine"
	"github.com/furrysalamander/onvif-go/onvif/svc/advancedsecurity"
	"github.com/furrysalamander/onvif-go/onvif/svc/analytics"
	"github.com/furrysalamander/onvif-go/onvif/svc/appmgmt"
	"github.com/furrysalamander/onvif-go/onvif/svc/authenticationbehavior"
	"github.com/furrysalamander/onvif-go/onvif/svc/credential"
	"github.com/furrysalamander/onvif-go/onvif/svc/deviceio"
	"github.com/furrysalamander/onvif-go/onvif/svc/devicemgmt"
	"github.com/furrysalamander/onvif-go/onvif/svc/display"
	"github.com/furrysalamander/onvif-go/onvif/svc/doorcontrol"
	"github.com/furrysalamander/onvif-go/onvif/svc/events"
	"github.com/furrysalamander/onvif-go/onvif/svc/imaging"
	"github.com/furrysalamander/onvif-go/onvif/svc/media"
	"github.com/furrysalamander/onvif-go/onvif/svc/media2"
	"github.com/furrysalamander/onvif-go/onvif/svc/provisioning"
	"github.com/furrysalamander/onvif-go/onvif/svc/ptz"
	"github.com/furrysalamander/onvif-go/onvif/svc/receiver"
	"github.com/furrysalamander/onvif-go/onvif/svc/recording"
	"github.com/furrysalamander/onvif-go/onvif/svc/replay"
	"github.com/furrysalamander/onvif-go/onvif/svc/schedule"
	"github.com/furrysalamander/onvif-go/onvif/svc/search"
	"github.com/furrysalamander/onvif-go/onvif/svc/thermal"
	"github.com/furrysalamander/onvif-go/onvif/svc/uplink"
)

type Device struct {
	endpoint string
	soap     *soaphdr.Client

	dm    *devicemgmt.Client
	med   *media.Client
	md2   *media2.Client
	pz    *ptz.Client
	ev    *events.Client
	img   *imaging.Client
	aio   *deviceio.Client
	disp  *display.Client
	rec   *recording.Client
	rpl   *replay.Client
	srch  *search.Client
	rcv   *receiver.Client
	th    *thermal.Client
	an    *analytics.Client
	ae    *actionengine.Client
	app   *appmgmt.Client
	ul    *uplink.Client
	as    *advancedsecurity.Client
	cr    *credential.Client
	ar    *accessrules.Client
	ac    *accesscontrol.Client
	dc    *doorcontrol.Client
	ab    *authenticationbehavior.Client
	prov  *provisioning.Client
	sched *schedule.Client
}

func NewDevice(endpoint, username, password string) *Device {
	c := soaphdr.New(endpoint, username, password)
	return &Device{
		endpoint: endpoint,
		soap:     c,
		dm:       devicemgmt.NewClientWithTransport(c),
		med:      media.NewClientWithTransport(c),
		md2:      media2.NewClientWithTransport(c),
		pz:       ptz.NewClientWithTransport(c),
		ev:       events.NewClientWithTransport(c),
		img:      imaging.NewClientWithTransport(c),
		aio:      deviceio.NewClientWithTransport(c),
		disp:     display.NewClientWithTransport(c),
		rec:      recording.NewClientWithTransport(c),
		rpl:      replay.NewClientWithTransport(c),
		srch:     search.NewClientWithTransport(c),
		rcv:      receiver.NewClientWithTransport(c),
		th:       thermal.NewClientWithTransport(c),
		an:       analytics.NewClientWithTransport(c),
		ae:       actionengine.NewClientWithTransport(c),
		app:      appmgmt.NewClientWithTransport(c),
		ul:       uplink.NewClientWithTransport(c),
		as:       advancedsecurity.NewClientWithTransport(c),
		cr:       credential.NewClientWithTransport(c),
		ar:       accessrules.NewClientWithTransport(c),
		ac:       accesscontrol.NewClientWithTransport(c),
		dc:       doorcontrol.NewClientWithTransport(c),
		ab:       authenticationbehavior.NewClientWithTransport(c),
		prov:     provisioning.NewClientWithTransport(c),
		sched:    schedule.NewClientWithTransport(c),
	}
}

func (d *Device) SetHTTP(h *http.Client) { d.soap.HTTP = h }

func (d *Device) DeviceMgmt() *devicemgmt.Client               { return d.dm }
func (d *Device) Media() *media.Client                         { return d.med }
func (d *Device) Media2() *media2.Client                       { return d.md2 }
func (d *Device) PTZ() *ptz.Client                             { return d.pz }
func (d *Device) Events() *events.Client                       { return d.ev }
func (d *Device) Imaging() *imaging.Client                     { return d.img }
func (d *Device) DeviceIO() *deviceio.Client                   { return d.aio }
func (d *Device) Display() *display.Client                     { return d.disp }
func (d *Device) Recording() *recording.Client                 { return d.rec }
func (d *Device) Replay() *replay.Client                       { return d.rpl }
func (d *Device) Search() *search.Client                       { return d.srch }
func (d *Device) Receiver() *receiver.Client                   { return d.rcv }
func (d *Device) Thermal() *thermal.Client                     { return d.th }
func (d *Device) Analytics() *analytics.Client                 { return d.an }
func (d *Device) ActionEngine() *actionengine.Client           { return d.ae }
func (d *Device) AppMgmt() *appmgmt.Client                     { return d.app }
func (d *Device) Uplink() *uplink.Client                       { return d.ul }
func (d *Device) AdvancedSecurity() *advancedsecurity.Client   { return d.as }
func (d *Device) Credential() *credential.Client               { return d.cr }
func (d *Device) AccessRules() *accessrules.Client             { return d.ar }
func (d *Device) AccessControl() *accesscontrol.Client         { return d.ac }
func (d *Device) DoorControl() *doorcontrol.Client             { return d.dc }
func (d *Device) AuthBehavior() *authenticationbehavior.Client { return d.ab }
func (d *Device) Provisioning() *provisioning.Client           { return d.prov }
func (d *Device) Schedule() *schedule.Client                   { return d.sched }

func (d *Device) GetServices(ctx context.Context, includeCapability bool) ([]tds.Service, error) {
	req := &tds.GetServices{IncludeCapability: includeCapability}
	res := &tds.GetServicesResponse{}
	if err := d.soap.Do(ctx, "http://www.onvif.org/ver10/device/wsdl/GetServices", req, res); err != nil {
		return nil, err
	}
	return res.Service, nil
}

type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareID      string
}

func (d *Device) GetInfo(ctx context.Context) (*DeviceInfo, error) {
	di, err := d.DeviceMgmt().GetDeviceInformation(ctx)
	if err != nil {
		return nil, err
	}
	return &DeviceInfo{
		Manufacturer:    di.Manufacturer,
		Model:           di.Model,
		FirmwareVersion: di.FirmwareVersion,
		SerialNumber:    di.SerialNumber,
		HardwareID:      di.HardwareId,
	}, nil
}

func (d *Device) GetCapabilities(ctx context.Context, categories ...tt.CapabilityCategory) (*tds.GetCapabilitiesResponse, error) {
	req := &tds.GetCapabilities{Category: categories}
	res := &tds.GetCapabilitiesResponse{}
	if err := d.soap.Do(ctx, "http://www.onvif.org/ver10/device/wsdl/GetCapabilities", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Device) GetStreamURI(ctx context.Context, profileToken string, streamType tt.StreamType, protocol tt.TransportProtocol) (string, error) {
	req := &trt.GetStreamUri{
		StreamSetup: tt.StreamSetup{
			Stream:    streamType,
			Transport: tt.Transport{Protocol: protocol},
		},
		ProfileToken: tt.ReferenceToken(profileToken),
	}
	res := &trt.GetStreamUriResponse{}
	if err := d.soap.Do(ctx, "http://www.onvif.org/ver10/media/wsdl/GetStreamUri", req, res); err != nil {
		return "", err
	}
	return res.MediaUri.Uri, nil
}

func (d *Device) GetSnapshotURI(ctx context.Context, profileToken string) (string, error) {
	req := &trt.GetSnapshotUri{
		ProfileToken: tt.ReferenceToken(profileToken),
	}
	res := &trt.GetSnapshotUriResponse{}
	if err := d.soap.Do(ctx, "http://www.onvif.org/ver10/media/wsdl/GetSnapshotUri", req, res); err != nil {
		return "", err
	}
	return res.MediaUri.Uri, nil
}

type FoundDevice struct {
	Info  wsdiscovery.DeviceInfo
	XAddr string
}

func (fd *FoundDevice) Connect(username, password string) *Device {
	endpoint := fd.XAddr
	if endpoint == "" && len(fd.Info.XAddrs) > 0 {
		endpoint = fd.Info.XAddrs[0]
	}
	return NewDevice(endpoint, username, password)
}

func Discover(ctx context.Context) ([]*FoundDevice, error) {
	devices, err := wsdiscovery.Discover(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*FoundDevice, len(devices))
	for i, d := range devices {
		xaddr := ""
		if len(d.XAddrs) > 0 {
			xaddr = d.XAddrs[0]
		}
		out[i] = &FoundDevice{
			Info:  d,
			XAddr: xaddr,
		}
	}
	return out, nil
}
