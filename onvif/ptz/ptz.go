package ptz

import (
	"github.com/furrysalamander/onvif-go/onvif/schema/tptz"
	"github.com/furrysalamander/onvif-go/onvif/schema/tt"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
)

const actionBase = "http://www.onvif.org/ver20/ptz/wsdl"

type Client struct {
	c *soaphdr.Client
}

func New(endpoint, username, password string) *Client {
	return &Client{c: soaphdr.New(endpoint, username, password)}
}

func (c *Client) GetNodes() (*tptz.GetNodesResponse, error) {
	res := &tptz.GetNodesResponse{}
	err := c.c.Do(actionBase+"/GetNodes", &tptz.GetNodes{}, res)
	return res, err
}

func (c *Client) GetStatus(profileToken string) (*tptz.GetStatusResponse, error) {
	res := &tptz.GetStatusResponse{}
	err := c.c.Do(actionBase+"/GetStatus", &tptz.GetStatus{ProfileToken: tt.ReferenceToken(profileToken)}, res)
	return res, err
}
