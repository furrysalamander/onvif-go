// Package media2 implements the ONVIF Media v2 client.
//
// Create a Client with New and call operations like GetProfiles,
// GetVideoSourceConfigurations, or GetVideoEncoderConfigurations.
package media2

import (
	"context"

	"github.com/furrysalamander/onvif-go/onvif/schema/tr2"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
)

const actionBase = "http://www.onvif.org/ver20/media/wsdl"

type Client struct {
	c *soaphdr.Client
}

func New(endpoint, username, password string) *Client {
	return &Client{c: soaphdr.New(endpoint, username, password)}
}

func (c *Client) GetVideoSourceConfigurations(ctx context.Context) (*tr2.GetVideoSourceConfigurationsResponse, error) {
	res := &tr2.GetVideoSourceConfigurationsResponse{}
	err := c.c.Do(ctx, actionBase+"/GetVideoSourceConfigurations", &tr2.GetVideoSourceConfigurations{}, res)
	return res, err
}

func (c *Client) GetProfiles(ctx context.Context) (*tr2.GetProfilesResponse, error) {
	res := &tr2.GetProfilesResponse{}
	err := c.c.Do(ctx, actionBase+"/GetProfiles", &tr2.GetProfiles{}, res)
	return res, err
}
