package imaging

import (
	"github.com/furrysalamander/onvif-go/onvif/schema/timg"
	"github.com/furrysalamander/onvif-go/onvif/schema/tt"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
)

const actionBase = "http://www.onvif.org/ver20/imaging/wsdl"

type Client struct {
	c *soaphdr.Client
}

func New(endpoint, username, password string) *Client {
	return &Client{c: soaphdr.New(endpoint, username, password)}
}

func (c *Client) GetImagingSettings(videoSourceToken string) (*timg.GetImagingSettingsResponse, error) {
	res := &timg.GetImagingSettingsResponse{}
	err := c.c.Do(actionBase+"/GetImagingSettings", &timg.GetImagingSettings{VideoSourceToken: tt.ReferenceToken(videoSourceToken)}, res)
	return res, err
}

func (c *Client) GetMoveOptions(videoSourceToken string) (*timg.GetMoveOptionsResponse, error) {
	res := &timg.GetMoveOptionsResponse{}
	err := c.c.Do(actionBase+"/GetMoveOptions", &timg.GetMoveOptions{VideoSourceToken: tt.ReferenceToken(videoSourceToken)}, res)
	return res, err
}
