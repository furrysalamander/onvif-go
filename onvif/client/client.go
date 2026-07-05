// Package client is the ONVIF client transport and service routing layer.
// Placeholder during M0; real implementation arrives in M3 + M4.
package client

// Device is a discovered ONVIF device endpoint. M0 placeholder.
type Device struct {
	Address         string
	Scopes          []string
	Types           []string
	XAddrs          []string
	MetadataVersion string
}
