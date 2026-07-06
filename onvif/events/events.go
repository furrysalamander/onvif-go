// Package events implements the ONVIF Events client with WS-Notification
// pull-point subscription support.
//
// Create a Client with New and use CreatePullPointSubscription to subscribe,
// PullMessages to receive events, and Unsubscribe to tear down.
package events

import (
	"context"
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

func (c *Client) GetEventProperties(ctx context.Context) (*tev.GetEventPropertiesResponse, error) {
	res := &tev.GetEventPropertiesResponse{}
	err := c.c.Do(ctx, actionBase+"/GetEventProperties", &tev.GetEventProperties{}, res)
	return res, err
}

func (c *Client) CreatePullPointSubscription(ctx context.Context) (*tev.CreatePullPointSubscriptionResponse, error) {
	res := &tev.CreatePullPointSubscriptionResponse{}
	err := c.c.Do(ctx, actionBase+"/CreatePullPointSubscription", &tev.CreatePullPointSubscription{}, res)
	return res, err
}

func (c *Client) PullMessages(ctx context.Context, timeout time.Duration, messageLimit int) (*tev.PullMessagesResponse, error) {
	res := &tev.PullMessagesResponse{}
	err := c.c.Do(ctx, actionBase+"/PullMessages", &tev.PullMessages{
		Timeout:      &core.Duration{Minutes: int(timeout.Minutes())},
		MessageLimit: int32(messageLimit),
	}, res)
	return res, err
}

func (c *Client) Seek(ctx context.Context, utcTime time.Time, reverse bool) (*tev.SeekResponse, error) {
	res := &tev.SeekResponse{}
	req := &tev.Seek{UtcTime: utcTime}
	if reverse {
		req.Reverse = &reverse
	}
	err := c.c.Do(ctx, actionBase+"/Seek", req, res)
	return res, err
}

func (c *Client) SetSynchronizationPoint(ctx context.Context) (*tev.SetSynchronizationPointResponse, error) {
	res := &tev.SetSynchronizationPointResponse{}
	err := c.c.Do(ctx, actionBase+"/SetSynchronizationPoint", &tev.SetSynchronizationPoint{}, res)
	return res, err
}

func (c *Client) CreatePullPoint(ctx context.Context) (*wsn.CreatePullPointResponse, error) {
	res := &wsn.CreatePullPointResponse{}
	err := c.c.Do(ctx, "http://docs.oasis-open.org/wsn/bw-2/CreatePullPoint", &wsn.CreatePullPoint{}, res)
	return res, err
}

func (c *Client) Renew(ctx context.Context, terminationTime time.Time) (*wsn.RenewResponse, error) {
	res := &wsn.RenewResponse{}
	err := c.c.Do(ctx, "http://docs.oasis-open.org/wsn/bw-2/Renew", &wsn.Renew{
		TerminationTime: wsn.AbsoluteOrRelativeTimeType(terminationTime),
	}, res)
	return res, err
}

func (c *Client) Unsubscribe(ctx context.Context) (*wsn.UnsubscribeResponse, error) {
	res := &wsn.UnsubscribeResponse{}
	err := c.c.Do(ctx, "http://docs.oasis-open.org/wsn/bw-2/Unsubscribe", &wsn.Unsubscribe{}, res)
	return res, err
}
