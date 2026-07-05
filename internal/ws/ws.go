// Package ws implements SOAP 1.2 envelope construction, parsing, and the
// ONVIF-specific fault model. Placeholder during M0; real implementation
// arrives in M3.
package ws

// Envelope is a SOAP 1.2 envelope. M0 placeholder.
type Envelope struct {
	Header Header
	Body   Body
}

// Header is the SOAP header block.
type Header struct{}

// Body is the SOAP body block.
type Body struct{}
