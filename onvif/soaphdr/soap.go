package soaphdr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/furrysalamander/onvif-go/internal/ws"
	"github.com/furrysalamander/onvif-go/internal/wssecurity"
)

type Client struct {
	Endpoint string
	Username string
	Password string
	HTTP     *http.Client
}

func New(endpoint, username, password string) *Client {
	return &Client{
		Endpoint: endpoint,
		Username: username,
		Password: password,
		HTTP:     http.DefaultClient,
	}
}

func (c *Client) Do(action string, reqBody interface{}, resBody interface{}) error {
	ut, err := wssecurity.NewUsernameToken(c.Username, c.Password)
	if err != nil {
		return fmt.Errorf("soap: token: %w", err)
	}
	actionHdr := ws.NewAction(action)
	to, err := ws.NewTo(c.Endpoint)
	if err != nil {
		return fmt.Errorf("soap: endpoint: %w", err)
	}

	data, err := ws.MarshalRequest("", reqBody, ut, actionHdr, to)
	if err != nil {
		return fmt.Errorf("soap: marshal: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("soap: http: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return fmt.Errorf("soap: send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("soap: read: %w", err)
	}

	fault, err := ws.UnmarshalResponse(body, resBody)
	if err != nil {
		return fmt.Errorf("soap: unmarshal: %w", err)
	}
	if fault != nil {
		return fmt.Errorf("soap: SOAP fault: code=%s reason=%v", fault.Code.Value, fault.Reason)
	}
	return nil
}
