package ws

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/furrysalamander/onvif-go/onvif/schema/env"
)

type ActionHeader struct {
	XMLName        xml.Name `xml:"http://www.w3.org/2005/08/addressing Action"`
	MustUnderstand string   `xml:"http://schemas.xmlsoap.org/soap/envelope/ mustUnderstand,attr"`
	Value          string   `xml:",chardata"`
}

type ToHeader struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/08/addressing To"`
	Value   string   `xml:",chardata"`
}

func NewAction(actionURI string) *ActionHeader {
	return &ActionHeader{
		MustUnderstand: "1",
		Value:          actionURI,
	}
}

func NewTo(endpoint string) (*ToHeader, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("ws: parse endpoint: %w", err)
	}
	return &ToHeader{Value: u.String()}, nil
}

func MarshalRequest(action string, body interface{}, headers ...interface{}) ([]byte, error) {
	e := &env.Envelope{}

	for _, h := range headers {
		if err := e.AddHeader(h); err != nil {
			return nil, fmt.Errorf("ws: marshal header: %w", err)
		}
	}

	if err := e.SetBody(body); err != nil {
		return nil, fmt.Errorf("ws: marshal body: %w", err)
	}

	return xml.Marshal(e)
}

func UnmarshalResponse(data []byte, body interface{}) (*env.Fault, error) {
	var e env.Envelope
	if err := xml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("ws: unmarshal envelope: %w", err)
	}
	if e.Body.Fault != nil {
		return e.Body.Fault, nil
	}
	if len(e.Body.Any) == 0 {
		return nil, fmt.Errorf("ws: empty body (no content and no fault)")
	}
	if err := e.Body.UnmarshalBody(body); err != nil {
		return nil, fmt.Errorf("ws: unmarshal body: %w", err)
	}
	return nil, nil
}
