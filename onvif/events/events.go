package events

import (
	"time"

	"github.com/furrysalamander/onvif-go/onvif/schema/core"
	"github.com/furrysalamander/onvif-go/onvif/schema/tev"
	"github.com/furrysalamander/onvif-go/onvif/schema/wsn"
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

func (c *Client) PullMessages(timeout time.Duration, messageLimit int) (*tev.PullMessagesResponse, error) {
	res := &tev.PullMessagesResponse{}
	err := c.c.Do(actionBase+"/PullMessages", &tev.PullMessages{
		Timeout:      &core.Duration{Minutes: int(timeout.Minutes())},
		MessageLimit: int32(messageLimit),
	}, res)
	return res, err
}

func (c *Client) Seek(utcTime time.Time, reverse bool) (*tev.SeekResponse, error) {
	res := &tev.SeekResponse{}
	req := &tev.Seek{UtcTime: utcTime}
	if reverse {
		req.Reverse = &reverse
	}
	err := c.c.Do(actionBase+"/Seek", req, res)
	return res, err
}

func (c *Client) SetSynchronizationPoint() (*tev.SetSynchronizationPointResponse, error) {
	res := &tev.SetSynchronizationPointResponse{}
	err := c.c.Do(actionBase+"/SetSynchronizationPoint", &tev.SetSynchronizationPoint{}, res)
	return res, err
}

func (c *Client) CreatePullPoint() (*wsn.CreatePullPointResponse, error) {
	res := &wsn.CreatePullPointResponse{}
	err := c.c.Do("http://docs.oasis-open.org/wsn/bw-2/CreatePullPoint", &wsn.CreatePullPoint{}, res)
	return res, err
}

func (c *Client) Renew(terminationTime time.Time) (*wsn.RenewResponse, error) {
	res := &wsn.RenewResponse{}
	err := c.c.Do("http://docs.oasis-open.org/wsn/bw-2/Renew", &wsn.Renew{
		TerminationTime: wsn.AbsoluteOrRelativeTimeType(terminationTime),
	}, res)
	return res, err
}

func (c *Client) Unsubscribe() (*wsn.UnsubscribeResponse, error) {
	res := &wsn.UnsubscribeResponse{}
	err := c.c.Do("http://docs.oasis-open.org/wsn/bw-2/Unsubscribe", &wsn.Unsubscribe{}, res)
	return res, err
}
