// Package catalog resolves WSDL/XSD imports against a vendored local catalog
// of ONVIF and OASIS/W3C support schemas. It never fetches schemas from the
// network at runtime.
//
// This file is a placeholder during M0; the real implementation arrives in M1.
package catalog

// Resolver returns the local filesystem path for a given (namespace, location)
// import, or false if unknown.
type Resolver interface {
	Resolve(namespace, location string) (path string, ok bool)
}

// NoopResolver is the M0 placeholder.
type NoopResolver struct{}

func (NoopResolver) Resolve(string, string) (string, bool) { return "", false }