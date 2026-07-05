// Package catalog resolves WSDL/XSD imports against a vendored local catalog
// of ONVIF and OASIS/W3C support schemas. It never fetches schemas from the
// network at runtime.
package catalog

import (
	"bytes"
	"embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
)

//go:embed catalog.xml onvif external
var fsys embed.FS

// Root is the embed FS root directory for the catalog.
const Root = "github.com/furrysalamander/onvif-go/internal/catalog"

// Resolver resolves an import to a path inside the catalog's embed FS.
type Resolver interface {
	// Resolve returns the path (relative to the catalog root, e.g.
	// "onvif/ver10/schema/onvif.xsd") for the given (namespace, location)
	// pair, or ok=false if no catalog entry matches.
	Resolve(namespace, location string) (catalogPath string, ok bool)
}

// NoopResolver resolves nothing. Useful for tests that don't need catalog
// support.
type NoopResolver struct{}

// Resolve implements Resolver. It always reports the import as unknown.
func (NoopResolver) Resolve(string, string) (string, bool) { return "", false }

// Catalog is a loaded XML Catalog that can resolve imports.
type Catalog struct {
	uris   map[string]string // namespace-or-URL -> catalog path
	system map[string]string // systemId (schemaLocation) -> catalog path
}

// Load parses the embedded catalog.xml and returns a Catalog.
func Load() (*Catalog, error) {
	data, err := fs.ReadFile(fsys, "catalog.xml")
	if err != nil {
		return nil, fmt.Errorf("catalog: read catalog.xml: %w", err)
	}
	return Parse(data)
}

// Parse builds a Catalog from raw catalog.xml bytes (mainly for tests).
func Parse(data []byte) (*Catalog, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var c Catalog
	c.uris = make(map[string]string)
	c.system = make(map[string]string)
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("catalog: decode: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "uri":
			name := attrValue(se.Attr, "name")
			uri := attrValue(se.Attr, "uri")
			if name == "" || uri == "" {
				continue
			}
			c.uris[name] = path.Clean(uri)
		case "system":
			systemID := attrValue(se.Attr, "systemId")
			uri := attrValue(se.Attr, "uri")
			if systemID == "" || uri == "" {
				continue
			}
			c.system[systemID] = path.Clean(uri)
		}
	}
	return &c, nil
}

// Resolve implements Resolver. Resolution order:
//  1. exact systemId match against <system> entries (full schemaLocation URL)
//  2. exact namespace/URL match against <uri> entries
//  3. location-as-namespace match against <uri> entries (some WSDLs omit a
//     namespace and only give a URL; some schemas use the URL as the name)
func (c *Catalog) Resolve(namespace, location string) (string, bool) {
	if location != "" {
		if p, ok := c.system[location]; ok {
			return p, true
		}
		if p, ok := c.uris[location]; ok {
			return p, true
		}
	}
	if namespace != "" {
		if p, ok := c.uris[namespace]; ok {
			return p, true
		}
	}
	return "", false
}

// Open returns a reader for the vendored file at the resolved catalog path.
func (c *Catalog) Open(catalogPath string) (io.ReadCloser, error) {
	f, err := fsys.Open(catalogPath)
	if err != nil {
		return nil, fmt.Errorf("catalog: open %s: %w", catalogPath, err)
	}
	return f, nil
}

// FS returns the underlying embed filesystem (for callers that need to walk
// the catalog or read files directly).
func (c *Catalog) FS() fs.FS { return fsys }

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
