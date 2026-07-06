package soaphdr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/ws"
	"github.com/furrysalamander/onvif-go/internal/wssecurity"
	"github.com/furrysalamander/onvif-go/onvif/schema/env"
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

type Fault struct {
	Code   string
	Reason string
	Detail string
}

func (f *Fault) Error() string {
	s := fmt.Sprintf("soap: SOAP fault: code=%s reason=%s", f.Code, f.Reason)
	if f.Detail != "" {
		s += " detail=" + f.Detail
	}
	return s
}

func (c *Client) Do(ctx context.Context, action string, reqBody, resBody interface{}) error {
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewReader(data))
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
		return envToFault(fault)
	}
	return nil
}

func envToFault(f *env.Fault) *Fault {
	ft := &Fault{
		Code: fullCode(f.Code),
	}
	if f.Reason != nil && len(f.Reason.Texts) > 0 {
		ft.Reason = f.Reason.Texts[0].Text
	}
	if f.Detail != nil && len(f.Detail.Any) > 0 {
		var b strings.Builder
		for _, a := range f.Detail.Any {
			b.Write(a)
		}
		ft.Detail = b.String()
	}
	return ft
}

func fullCode(c *env.FaultCode) string {
	if c == nil {
		return ""
	}
	s := c.Value
	if c.Subcode != nil {
		sub := fullCode(c.Subcode)
		if sub != "" {
			s += "/" + sub
		}
	}
	return s
}

func NewWithContext(endpoint, username, password string) *Client {
	return New(endpoint, username, password)
}
