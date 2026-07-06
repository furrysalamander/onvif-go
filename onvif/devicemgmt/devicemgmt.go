// Package devicemgmt implements the ONVIF Device Management client.
//
// Create a Client with NewClient and call operations like
// GetDeviceInformation, GetServices, GetCapabilities, or GetScopes.
package devicemgmt

import (
	"context"

	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
)

const actionBase = "http://www.onvif.org/ver10/device/wsdl"

type Client struct {
	c *soaphdr.Client
}

func NewClient(endpoint, username, password string) *Client {
	return &Client{c: soaphdr.New(endpoint, username, password)}
}

func (c *Client) GetServices(ctx context.Context) (*tds.GetServicesResponse, error) {
	res := &tds.GetServicesResponse{}
	err := c.c.Do(ctx, actionBase+"/GetServices", &tds.GetServices{}, res)
	return res, err
}

func (c *Client) GetDeviceInformation(ctx context.Context) (*tds.GetDeviceInformationResponse, error) {
	res := &tds.GetDeviceInformationResponse{}
	err := c.c.Do(ctx, actionBase+"/GetDeviceInformation", &tds.GetDeviceInformation{}, res)
	return res, err
}

func (c *Client) GetCapabilities(ctx context.Context) (*tds.GetCapabilitiesResponse, error) {
	res := &tds.GetCapabilitiesResponse{}
	err := c.c.Do(ctx, actionBase+"/GetCapabilities", &tds.GetCapabilities{}, res)
	return res, err
}
