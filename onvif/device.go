package onvif

import (
	"context"
	"net/http"

	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
	"github.com/furrysalamander/onvif-go/onvif/svc/devicemgmt"
	"github.com/furrysalamander/onvif-go/onvif/svc/events"
	"github.com/furrysalamander/onvif-go/onvif/svc/imaging"
	"github.com/furrysalamander/onvif-go/onvif/svc/media"
	"github.com/furrysalamander/onvif-go/onvif/svc/media2"
	"github.com/furrysalamander/onvif-go/onvif/svc/ptz"
)

type Device struct {
	endpoint string
	soap     *soaphdr.Client

	dm  *devicemgmt.Client
	med *media.Client
	md2 *media2.Client
	pz  *ptz.Client
	ev  *events.Client
	img *imaging.Client
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
	}
}

func (d *Device) SetHTTP(h *http.Client) { d.soap.HTTP = h }

func (d *Device) DeviceMgmt() *devicemgmt.Client { return d.dm }
func (d *Device) Media() *media.Client           { return d.med }
func (d *Device) Media2() *media2.Client         { return d.md2 }
func (d *Device) PTZ() *ptz.Client               { return d.pz }
func (d *Device) Events() *events.Client         { return d.ev }
func (d *Device) Imaging() *imaging.Client       { return d.img }

func (d *Device) GetServices(ctx context.Context, includeCapability bool) ([]tds.Service, error) {
	req := &tds.GetServices{IncludeCapability: includeCapability}
	res := &tds.GetServicesResponse{}
	if err := d.soap.Do(ctx, "http://www.onvif.org/ver10/device/wsdl/GetServices", req, res); err != nil {
		return nil, err
	}
	return res.Service, nil
}
