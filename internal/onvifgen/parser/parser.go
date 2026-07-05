// Package parser turns the vendored WSDL/XSD files into the onvifgen IR.
package parser

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/catalog"
	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// Parser parses WSDL and XSD documents into IR modules. It does NOT resolve
// cross-module references; the Loader does that after collecting all modules.
type Parser struct {
	cat *catalog.Catalog
}

// New returns a Parser that resolves relative imports against the catalog.
func New(cat *catalog.Catalog) *Parser { return &Parser{cat: cat} }

// Load returns the catalog this parser uses for relative-import resolution.
func (p *Parser) Load() *catalog.Catalog { return p.cat }

// ParseWSDL parses a WSDL document from r. file is the catalog-relative path
// (used for diagnostics + IR metadata); fromImporterDir is the directory of
// the document that referenced this WSDL, used to resolve relative
// schemaLocation/location paths. Root callers pass "".
func (p *Parser) ParseWSDL(file, fromImporterDir string, r io.Reader) (*ir.WSDL, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	w := &ir.WSDL{File: file}
	// Track the document's own target namespace and inline prefix map to
	// resolve QNames against the catalog.
	var targetNS string
	prefixes := map[string]string{}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("token %s: %w", file, err)
		}
		if t, ok := tok.(xml.StartElement); ok {
			if t.Name.Local == "definitions" && (t.Name.Space == wsdlNS || t.Name.Space == wsdlNS11) {
				for _, a := range t.Attr {
					switch {
					case a.Name.Local == "targetNamespace" && a.Value != "":
						targetNS = a.Value
					case a.Name.Space == "xmlns" && a.Value != "" && a.Name.Local != "":
						prefixes[a.Name.Local] = a.Value
					}
				}
				w.TargetNS = targetNS
				if err := p.decodeDefinitions(dec, w, file, targetNS, prefixes); err != nil {
					return nil, err
				}
			}
		}
	}
	return w, nil
}

// ParseSchema parses an XSD document from r.
func (p *Parser) ParseSchema(file, fromImporterDir string, r io.Reader) (*ir.Schema, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	prefixes := map[string]string{}
	targetNS := ""
	// Find the <xs:schema> root and capture its targetNamespace + prefixes.
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("token %s: %w", file, err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "schema" && se.Name.Space == xsNS {
			for _, a := range se.Attr {
				switch {
				case a.Name.Local == "targetNamespace" && a.Value != "":
					targetNS = a.Value
				case a.Name.Space == "xmlns" && a.Value != "" && a.Name.Local != "":
					prefixes[a.Name.Local] = a.Value
				case a.Name.Local == "elementFormDefault":
					// captured below
				}
			}
			s := &ir.Schema{File: file, TargetNS: targetNS}
			for _, a := range se.Attr {
				if a.Name.Local == "elementFormDefault" {
					s.ElementFormDef = a.Value
				}
				if a.Name.Local == "attributeFormDefault" {
					s.AttributeFormDef = a.Value
				}
			}
			if err := p.decodeSchema(dec, s, file, targetNS, prefixes); err != nil {
				return nil, err
			}
			return s, nil
		}
	}
	return nil, fmt.Errorf("%s: no <xs:schema> root", file)
}

const (
	wsdlNS   = "http://schemas.xmlsoap.org/wsdl/"
	wsdlNS11 = "http://schemas.xmlsoap.org/wsdl/"
	xsNS     = "http://www.w3.org/2001/XMLSchema"
	soapNS   = "http://schemas.xmlsoap.org/wsdl/soap12/"
	soapNS11 = "http://schemas.xmlsoap.org/wsdl/soap/"
)

func splitQName(v string, prefixes map[string]string) ir.QName {
	v = strings.TrimSpace(v)
	if v == "" {
		return ir.QName{}
	}
	idx := strings.Index(v, ":")
	if idx < 0 {
		// No prefix — either XMLSchema builtin (xs:string etc, but those come
		// with prefixes in practice) or a no-prefix same-ns reference.
		return ir.QName{Local: v}
	}
	pfx, local := v[:idx], v[idx+1:]
	if pfx == "xs" || pfx == "xsd" {
		return ir.QName{NS: xsNS, Local: local}
	}
	ns := prefixes[pfx]
	return ir.QName{NS: ns, Local: local}
}

func normOccurs(v string) string {
	if v == "" {
		return "1"
	}
	return v
}
