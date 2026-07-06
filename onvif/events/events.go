package events

import (
	"github.com/furrysalamander/onvif-go/onvif/schema/tev"
	"github.com/furrysalamander/onvif-go/onvif/soaphdr"
)

const actionBase = "http://www.onvif.org/ver10/events/wsdl"

type Client struct {
	c *soaphdr.Client
}

func New(endpoint, username, password string) *Client {
	return &Client{c: soaphdr.New(endpoint, username, password)}
}

func (c *Client) GetEventProperties() (*tev.GetEventPropertiesResponse, error) {
	res := &tev.GetEventPropertiesResponse{}
	err := c.c.Do(actionBase+"/GetEventProperties", &tev.GetEventProperties{}, res)
	return res, err
}

func (c *Client) CreatePullPointSubscription() (*tev.CreatePullPointSubscriptionResponse, error) {
	res := &tev.CreatePullPointSubscriptionResponse{}
	err := c.c.Do(actionBase+"/CreatePullPointSubscription", &tev.CreatePullPointSubscription{}, res)
	return res, err
}
