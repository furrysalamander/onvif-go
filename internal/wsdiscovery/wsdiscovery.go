package wsdiscovery

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/furrysalamander/onvif-go/onvif/schema/env"
)

const (
	MulticastAddr = "239.255.255.250"
	Port          = 3702
	nsDiscovery   = "http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01"
	nsWSA         = "http://www.w3.org/2005/08/addressing"
)

type Probe struct {
	XMLName xml.Name `xml:"http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01 Probe"`
	Types   string   `xml:"Types,omitempty"`
	Scopes  string   `xml:"Scopes,omitempty"`
}

type ProbeMatches struct {
	XMLName    xml.Name     `xml:"http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01 ProbeMatches"`
	ProbeMatch []ProbeMatch `xml:"ProbeMatch,omitempty"`
}

type ProbeMatch struct {
	EndpointReference EndpointReference `xml:"http://www.w3.org/2005/08/addressing EndpointReference"`
	Types             string            `xml:"Types,omitempty"`
	Scopes            string            `xml:"Scopes,omitempty"`
	XAddrs            string            `xml:"XAddrs,omitempty"`
	MetadataVersion   int               `xml:"MetadataVersion"`
}

type EndpointReference struct {
	Address string `xml:"Address"`
}

type Hello struct {
	XMLName           xml.Name          `xml:"http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01 Hello"`
	EndpointReference EndpointReference `xml:"http://www.w3.org/2005/08/addressing EndpointReference"`
	Types             string            `xml:"Types,omitempty"`
	Scopes            string            `xml:"Scopes,omitempty"`
	XAddrs            string            `xml:"XAddrs,omitempty"`
	MetadataVersion   int               `xml:"MetadataVersion"`
}

type Bye struct {
	XMLName           xml.Name          `xml:"http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01 Bye"`
	EndpointReference EndpointReference `xml:"http://www.w3.org/2005/08/addressing EndpointReference"`
}

type DeviceInfo struct {
	Address         string
	Types           []string
	Scopes          []string
	XAddrs          []string
	MetadataVersion int
	Endpoint        string
}

func Discover(ctx context.Context) ([]DeviceInfo, error) {
	return discover(ctx, "")
}

func DiscoverWithTypes(ctx context.Context, types string) ([]DeviceInfo, error) {
	return discover(ctx, types)
}

func discover(ctx context.Context, types string) ([]DeviceInfo, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", MulticastAddr, Port))
	if err != nil {
		return nil, fmt.Errorf("wsdiscovery: resolve: %w", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("wsdiscovery: listen: %w", err)
	}
	defer func() { _ = conn.Close() }()

	probe := &Probe{Types: types}
	reqEnv := &env.Envelope{}
	if err := reqEnv.SetBody(probe); err != nil {
		return nil, fmt.Errorf("wsdiscovery: marshal probe: %w", err)
	}

	data, err := xml.Marshal(reqEnv)
	if err != nil {
		return nil, fmt.Errorf("wsdiscovery: marshal: %w", err)
	}

	if _, err := conn.WriteTo(data, addr); err != nil {
		return nil, fmt.Errorf("wsdiscovery: send: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetReadDeadline(deadline)
	} else {
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	}

	buf := make([]byte, 65535)
	seen := map[string]bool{}
	var devices []DeviceInfo

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				break
			}
			return devices, fmt.Errorf("wsdiscovery: read: %w", err)
		}

		var respEnv env.Envelope
		if err := xml.Unmarshal(buf[:n], &respEnv); err != nil {
			continue
		}

		if len(respEnv.Body.Any) == 0 {
			continue
		}

		var matches ProbeMatches
		if err := xml.Unmarshal(respEnv.Body.Any[0], &matches); err != nil {
			continue
		}

		for _, m := range matches.ProbeMatch {
			addr := m.EndpointReference.Address
			if seen[addr] {
				continue
			}
			seen[addr] = true

			var xaddrs []string
			for _, x := range strings.Fields(m.XAddrs) {
				xaddrs = append(xaddrs, strings.TrimSpace(x))
			}

			devices = append(devices, DeviceInfo{
				Address:         addr,
				Types:           strings.Fields(m.Types),
				Scopes:          strings.Fields(m.Scopes),
				XAddrs:          xaddrs,
				MetadataVersion: m.MetadataVersion,
				Endpoint:        first(xaddrs),
			})
		}
	}
	return devices, nil
}

func first(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}
