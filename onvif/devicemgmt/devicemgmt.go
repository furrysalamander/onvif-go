package devicemgmt

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/furrysalamander/onvif-go/internal/ws"
	"github.com/furrysalamander/onvif-go/internal/wssecurity"
	"github.com/furrysalamander/onvif-go/onvif/schema/tds"
)

const actionBase = "http://www.onvif.org/ver10/device/wsdl"

type Client struct {
	Endpoint string
	Username string
	Password string
	Client   *http.Client
}

func NewClient(endpoint, username, password string) *Client {
	return &Client{
		Endpoint: endpoint,
		Username: username,
		Password: password,
		Client:   http.DefaultClient,
	}
}

func (c *Client) do(action string, reqBody interface{}, resBody interface{}) error {
	ut, err := wssecurity.NewUsernameToken(c.Username, c.Password)
	if err != nil {
		return fmt.Errorf("devicemgmt: token: %w", err)
	}
	actionHdr := ws.NewAction(action)
	to, err := ws.NewTo(c.Endpoint)
	if err != nil {
		return fmt.Errorf("devicemgmt: endpoint: %w", err)
	}

	data, err := ws.MarshalRequest("", reqBody, ut, actionHdr, to)
	if err != nil {
		return fmt.Errorf("devicemgmt: marshal: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("devicemgmt: http: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("devicemgmt: send: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("devicemgmt: read: %w", err)
	}

	fault, err := ws.UnmarshalResponse(body, resBody)
	if err != nil {
		return fmt.Errorf("devicemgmt: unmarshal: %w", err)
	}
	if fault != nil {
		return fmt.Errorf("devicemgmt: SOAP fault: code=%s reason=%v", fault.Code.Value, fault.Reason)
	}
	return nil
}

func (c *Client) GetServices() (*tds.GetServicesResponse, error) {
	res := &tds.GetServicesResponse{}
	err := c.do(actionBase+"/GetServices", &tds.GetServices{}, res)
	return res, err
}

func (c *Client) GetDeviceInformation() (*tds.GetDeviceInformationResponse, error) {
	res := &tds.GetDeviceInformationResponse{}
	err := c.do(actionBase+"/GetDeviceInformation", &tds.GetDeviceInformation{}, res)
	return res, err
}

func (c *Client) GetCapabilities() (*tds.GetCapabilitiesResponse, error) {
	res := &tds.GetCapabilitiesResponse{}
	err := c.do(actionBase+"/GetCapabilities", &tds.GetCapabilities{}, res)
	return res, err
}
