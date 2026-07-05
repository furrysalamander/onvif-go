package parser

import (
	"encoding/xml"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// decodeDefinitions walks the children of <wsdl:definitions>.
func (p *Parser) decodeDefinitions(dec *xml.Decoder, w *ir.WSDL, file, targetNS string, prefixes map[string]string) error {
	return decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "import":
			w.Imports = append(w.Imports, ir.WSDLImport{
				Namespace: attr(se.Attr, "namespace"),
				Location:  attr(se.Attr, "location"),
			})
			return skipElement(dec)
		case "types":
			return p.decodeTypes(dec, w, file, prefixes)
		case "message":
			m, err := p.decodeMessage(se, dec, prefixes)
			if err != nil {
				return err
			}
			w.Messages = append(w.Messages, m)
			return nil
		case "portType":
			pt, err := p.decodePortType(se, dec, prefixes)
			if err != nil {
				return err
			}
			w.PortTypes = append(w.PortTypes, pt)
			return nil
		case "binding":
			b, err := p.decodeBinding(se, dec, prefixes)
			if err != nil {
				return err
			}
			w.Bindings = append(w.Bindings, b)
			return nil
		case "service":
			svc, err := p.decodeService(se, dec, prefixes)
			if err != nil {
				return err
			}
			w.Services = append(w.Services, svc)
			return nil
		default:
			// documentation is rare at top level inside definitions; skip.
			return skipElement(dec)
		}
	})
}

// decodeTypes parses <wsdl:types> which contains one or more <xs:schema>.
func (p *Parser) decodeTypes(dec *xml.Decoder, w *ir.WSDL, file string, prefixes map[string]string) error {
	return decodeUntilEnd(dec, func(se xml.StartElement) error {
		if se.Name.Local == "schema" && se.Name.Space == xsNS {
			s := &ir.Schema{}
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "targetNamespace":
					s.TargetNS = a.Value
				case "elementFormDefault":
					s.ElementFormDef = a.Value
				case "attributeFormDefault":
					s.AttributeFormDef = a.Value
				}
			}
			// Merge the enclosing definitions' prefixes into the inline
			// schema's prefix map so QNames resolve identically.
			schemaPrefixes := copyPrefixes(prefixes)
			for _, a := range se.Attr {
				if a.Name.Space == "xmlns" && a.Value != "" && a.Name.Local != "" {
					schemaPrefixes[a.Name.Local] = a.Value
				}
			}
			s.File = file
			if err := p.decodeSchema(dec, s, file, s.TargetNS, schemaPrefixes); err != nil {
				return err
			}
			w.Types = append(w.Types, s)
			return nil
		}
		return skipElement(dec)
	})
}

func copyPrefixes(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (p *Parser) decodeMessage(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Message, error) {
	m := &ir.Message{Name: attr(se.Attr, "name")}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "documentation":
			_, err := readCharData(dec)
			return err
		case "part":
			mp := ir.MessagePart{
				Name:    attr(se.Attr, "name"),
				Element: attrQ(se.Attr, "element", prefixes),
				Type:    attrQ(se.Attr, "type", prefixes),
			}
			m.Parts = append(m.Parts, mp)
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Parser) decodePortType(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.PortType, error) {
	pt := &ir.PortType{Name: attr(se.Attr, "name")}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "documentation":
			doc, err := readCharData(dec)
			if err != nil {
				return err
			}
			pt.Doc = doc
			return nil
		case "operation":
			op, err := p.decodeOperation(se, dec, prefixes)
			if err != nil {
				return err
			}
			pt.Operations = append(pt.Operations, op)
			return nil
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return pt, nil
}

func (p *Parser) decodeOperation(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Operation, error) {
	op := &ir.Operation{
		Name:           attr(se.Attr, "name"),
		ParameterOrder: attr(se.Attr, "parameterOrder"),
	}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "documentation":
			doc, err := readCharData(dec)
			if err != nil {
				return err
			}
			op.Doc = doc
			return nil
		case "input":
			op.Input = &ir.OperationMsg{
				Message: attrQ(se.Attr, "message", prefixes),
				Name:    attr(se.Attr, "name"),
			}
			return skipElement(dec)
		case "output":
			op.Output = &ir.OperationMsg{
				Message: attrQ(se.Attr, "message", prefixes),
				Name:    attr(se.Attr, "name"),
			}
			return skipElement(dec)
		case "fault":
			op.Faults = append(op.Faults, ir.OperationFault{
				Name:    attr(se.Attr, "name"),
				Message: attrQ(se.Attr, "message", prefixes),
			})
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return op, nil
}

func (p *Parser) decodeBinding(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Binding, error) {
	b := &ir.Binding{
		Name:     attr(se.Attr, "name"),
		PortType: attrQ(se.Attr, "type", prefixes),
	}
	// Determine SOAP version from the namespace of <soap:binding> children.
	soapVersion := ""
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "binding":
			// <soap:binding style=... transport=...>
			switch se.Name.Space {
			case soapNS:
				soapVersion = "soap12"
			case soapNS11:
				soapVersion = "soap11"
			}
			b.Style = attr(se.Attr, "style")
			b.Transport = attr(se.Attr, "transport")
			return skipElement(dec)
		case "operation":
			bo, err := p.decodeBindingOperation(se, dec, prefixes)
			if err != nil {
				return err
			}
			b.Operations = append(b.Operations, bo)
			return nil
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	b.Type = soapVersion
	return b, nil
}

func (p *Parser) decodeBindingOperation(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (ir.BindingOperation, error) {
	bo := ir.BindingOperation{
		Name: attr(se.Attr, "name"),
	}
	// The <soap:operation soapAction=...> may carry the action on this
	// StartElement *or* as a child. Capture from this element's attrs first.
	bo.SOAPAction = attr(se.Attr, "soapAction")
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "operation":
			if bo.SOAPAction == "" {
				bo.SOAPAction = attr(se.Attr, "soapAction")
			}
			bo.Style = attr(se.Attr, "style")
			return skipElement(dec)
		case "input":
			bo.Input = p.decodeBindingMsg(dec, prefixes)
			return nil
		case "output":
			bo.Output = p.decodeBindingMsg(dec, prefixes)
			return nil
		case "fault":
			bf := ir.BindingFault{
				Name: attr(se.Attr, "name"),
				Use:  "literal",
			}
			bo.Faults = append(bo.Faults, bf)
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return bo, err
	}
	return bo, nil
}

// decodeBindingMsg drains a binding <input>/<output> body and records the
// soap:body use attribute (almost always "literal" in ONVIF).
func (p *Parser) decodeBindingMsg(dec *xml.Decoder, prefixes map[string]string) *ir.BindingMsg {
	bm := &ir.BindingMsg{Use: "literal"}
	_ = decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "body":
			u := attr(se.Attr, "use")
			if u != "" {
				bm.Use = u
			}
			return skipElement(dec)
		case "header":
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	})
	return bm
}

func (p *Parser) decodeService(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Service, error) {
	svc := &ir.Service{Name: attr(se.Attr, "name")}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "documentation":
			doc, err := readCharData(dec)
			if err != nil {
				return err
			}
			svc.Doc = doc
			return nil
		case "port":
			sp := ir.ServicePort{
				Name:    attr(se.Attr, "name"),
				Binding: attrQ(se.Attr, "binding", prefixes),
			}
			// <soap:address location=...> inside the port.
			if err := decodeUntilEnd(dec, func(c2 xml.StartElement) error {
				if c2.Name.Local == "address" {
					sp.Address = attr(c2.Attr, "location")
				}
				return skipElement(dec)
			}); err != nil {
				return err
			}
			svc.Ports = append(svc.Ports, sp)
			return nil
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return svc, nil
}
