// Package imaging implements the ONVIF Imaging client.
//
// Create a Client with New and call operations like GetImagingSettings
// or GetMoveOptions.
package imaging

import (
	"context"

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

func (c *Client) GetImagingSettings(ctx context.Context, videoSourceToken string) (*timg.GetImagingSettingsResponse, error) {
	res := &timg.GetImagingSettingsResponse{}
	err := c.c.Do(ctx, actionBase+"/GetImagingSettings", &timg.GetImagingSettings{VideoSourceToken: tt.ReferenceToken(videoSourceToken)}, res)
	return res, err
}

func (c *Client) GetMoveOptions(ctx context.Context, videoSourceToken string) (*timg.GetMoveOptionsResponse, error) {
	res := &timg.GetMoveOptionsResponse{}
	err := c.c.Do(ctx, actionBase+"/GetMoveOptions", &timg.GetMoveOptions{VideoSourceToken: tt.ReferenceToken(videoSourceToken)}, res)
	return res, err
}
