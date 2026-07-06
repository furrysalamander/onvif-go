// Package media implements the ONVIF Media v1 client.
//
// Create a Client with New and call operations like GetProfiles,
// GetVideoSources, or GetVideoSourceConfigurations.
package media

import (
	"github.com/furrysalamander/onvif-go/onvif/schema/trt"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
)

const actionBase = "http://www.onvif.org/ver10/media/wsdl"

type Client struct {
	c *soaphdr.Client
}

func New(endpoint, username, password string) *Client {
	return &Client{c: soaphdr.New(endpoint, username, password)}
}

func (c *Client) GetVideoSources() (*trt.GetVideoSourcesResponse, error) {
	res := &trt.GetVideoSourcesResponse{}
	err := c.c.Do(actionBase+"/GetVideoSources", &trt.GetVideoSources{}, res)
	return res, err
}

func (c *Client) GetProfiles() (*trt.GetProfilesResponse, error) {
	res := &trt.GetProfilesResponse{}
	err := c.c.Do(actionBase+"/GetProfiles", &trt.GetProfiles{}, res)
	return res, err
}
