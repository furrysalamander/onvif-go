package parser

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// skipElement drains tokens until the end of the element that was just
// started (the caller has already consumed the StartElement).
func skipElement(dec *xml.Decoder) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			_ = t
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

// readCharData reads the text content immediately inside the current element
// until its end. It tolerates interleaved comments.
func readCharData(dec *xml.Decoder) (string, error) {
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return b.String(), nil
		}
		if err != nil {
			return b.String(), err
		}
		switch t := tok.(type) {
		case xml.CharData:
			b.Write(t)
		case xml.EndElement:
			return b.String(), nil
		case xml.StartElement:
			// Recurse and discard nested children (e.g. formatting tags).
			if err := skipElement(dec); err != nil {
				return b.String(), err
			}
		}
	}
}

// attr finds an attribute by local name (ignoring namespace).
func attr(attrs []xml.Attr, local string) string {
	for _, a := range attrs {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// attrQ finds an attribute value and parses it as a QName using prefixes.
func attrQ(attrs []xml.Attr, local string, prefixes map[string]string) ir.QName {
	v := attr(attrs, local)
	if v == "" {
		return ir.QName{}
	}
	return splitQName(v, prefixes)
}

// readAnnotation drains an <xs:annotation> element (the caller has already
// consumed its StartElement) and returns the concatenated text of any
// <xs:documentation> children.
func readAnnotation(dec *xml.Decoder) (string, error) {
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return b.String(), err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "documentation" && t.Name.Space == xsNS {
				txt, err := readCharData(dec)
				if err != nil {
					return b.String(), err
				}
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				b.WriteString(strings.TrimSpace(txt))
			} else {
				if err := skipElement(dec); err != nil {
					return b.String(), err
				}
			}
		case xml.EndElement:
			return b.String(), nil
		}
	}
}

// splitWS splits a whitespace-separated list (used for xs:union memberTypes).
func splitWS(s string) []string {
	return strings.Fields(s)
}

// decodeUntilEnd processes children of the current element until its
// </EndElement>. dispatch receives each child StartElement and is responsible
// for fully consuming it (including its end tag).
func decodeUntilEnd(dec *xml.Decoder, dispatch func(xml.StartElement) error) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if err := dispatch(t); err != nil {
				return err
			}
		case xml.EndElement:
			return nil
		}
	}
}
