package parser

import (
	"encoding/xml"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// decodeSchema walks the children of <xs:schema>.
func (p *Parser) decodeSchema(dec *xml.Decoder, s *ir.Schema, file, targetNS string, prefixes map[string]string) error {
	return decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "import":
			s.Imports = append(s.Imports, ir.SchemaImport{
				Namespace:      attr(se.Attr, "namespace"),
				SchemaLocation: attr(se.Attr, "schemaLocation"),
			})
			return skipElement(dec)
		case "include":
			s.Includes = append(s.Includes, ir.SchemaInclude{
				SchemaLocation: attr(se.Attr, "schemaLocation"),
			})
			return skipElement(dec)
		case "simpleType":
			st, err := p.decodeSimpleType(se, dec, prefixes)
			if err != nil {
				return err
			}
			s.SimpleTypes = append(s.SimpleTypes, st)
			return nil
		case "complexType":
			ct, err := p.decodeComplexType(se, dec, prefixes)
			if err != nil {
				return err
			}
			s.ComplexTypes = append(s.ComplexTypes, ct)
			return nil
		case "element":
			e, err := p.decodeGlobalElement(se, dec, prefixes)
			if err != nil {
				return err
			}
			s.Elements = append(s.Elements, e)
			return nil
		case "attribute":
			a, err := p.decodeAttribute(se, dec, prefixes)
			if err != nil {
				return err
			}
			s.Attributes = append(s.Attributes, a)
			return nil
		case "attributeGroup":
			ag, err := p.decodeAttrGroup(se, dec, prefixes)
			if err != nil {
				return err
			}
			s.AttrGroups = append(s.AttrGroups, ag)
			return nil
		case "group":
			g, err := p.decodeGroup(se, dec, prefixes)
			if err != nil {
				return err
			}
			s.Groups = append(s.Groups, g)
			return nil
		default:
			return skipElement(dec)
		}
	})
}

// decodeSimpleType reads an <xs:simpleType>.
func (p *Parser) decodeSimpleType(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.SimpleType, error) {
	st := &ir.SimpleType{Name: attr(se.Attr, "name")}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "annotation":
			doc, err := readAnnotation(dec)
			if err != nil {
				return err
			}
			st.Doc = doc
			return nil
		case "restriction":
			st.Base = attrQ(se.Attr, "base", prefixes)
			return p.decodeRestriction(dec, prefixes, st)
		case "list":
			st.ListItemType = attrQ(se.Attr, "itemType", prefixes)
			return skipElement(dec)
		case "union":
			mt := attr(se.Attr, "memberTypes")
			for _, m := range splitWS(mt) {
				st.UnionMembers = append(st.UnionMembers, splitQName(m, prefixes))
			}
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return st, nil
}

// decodeRestriction fills facets on st.
func (p *Parser) decodeRestriction(dec *xml.Decoder, prefixes map[string]string, st *ir.SimpleType) error {
	return decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "enumeration":
			st.Enumeration = append(st.Enumeration, attr(se.Attr, "value"))
			return skipElement(dec)
		case "pattern":
			st.Pattern = attr(se.Attr, "value")
			return skipElement(dec)
		case "length":
			st.Length = attr(se.Attr, "value")
			return skipElement(dec)
		case "minLength":
			st.MinLength = attr(se.Attr, "value")
			return skipElement(dec)
		case "maxLength":
			st.MaxLength = attr(se.Attr, "value")
			return skipElement(dec)
		case "minInclusive":
			st.MinInclusive = attr(se.Attr, "value")
			return skipElement(dec)
		case "maxInclusive":
			st.MaxInclusive = attr(se.Attr, "value")
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	})
}

// decodeComplexType reads an <xs:complexType>.
func (p *Parser) decodeComplexType(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.ComplexType, error) {
	ct := &ir.ComplexType{
		Name:     attr(se.Attr, "name"),
		Abstract: attr(se.Attr, "abstract") == "true",
		Mixed:    attr(se.Attr, "mixed") == "true",
	}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "annotation":
			doc, err := readAnnotation(dec)
			if err != nil {
				return err
			}
			ct.Doc = doc
			return nil
		case "sequence":
			seq, err := p.decodeSequence(dec, prefixes)
			if err != nil {
				return err
			}
			ct.Content = seq
			return nil
		case "choice":
			ch, err := p.decodeChoice(dec, prefixes)
			if err != nil {
				return err
			}
			ct.Content = ch
			return nil
		case "all":
			all, err := p.decodeAll(dec, prefixes)
			if err != nil {
				return err
			}
			ct.Content = all
			return nil
		case "any":
			ct.Content = ir.Particle{Kind: ir.ParticleAny, Any: p.decodeAny(se)}
			return skipElement(dec)
		case "attribute":
			a, err := p.decodeAttribute(se, dec, prefixes)
			if err != nil {
				return err
			}
			ct.Attributes = append(ct.Attributes, a)
			return nil
		case "attributeGroup":
			ref := attrQ(se.Attr, "ref", prefixes)
			ct.Attributes = append(ct.Attributes, &ir.Attribute{Name: "@attrGroupRef", Type: ref})
			return skipElement(dec)
		case "anyAttribute":
			ct.AnyAttribute = append(ct.AnyAttribute, p.decodeAnyAttribute(se))
			return skipElement(dec)
		case "complexContent":
			return p.decodeComplexContent(dec, prefixes, ct)
		case "simpleContent":
			return p.decodeSimpleContent(dec, prefixes, ct)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return ct, nil
}

// decodeGlobalElement reads a top-level <xs:element>.
func (p *Parser) decodeGlobalElement(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Element, error) {
	e := &ir.Element{
		Name:              attr(se.Attr, "name"),
		Nillable:          attr(se.Attr, "nillable") == "true",
		Abstract:          attr(se.Attr, "abstract") == "true",
		SubstitutionGroup: attrQ(se.Attr, "substitutionGroup", prefixes),
	}
	if t := attr(se.Attr, "type"); t != "" {
		e.Type = splitQName(t, prefixes)
	}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "annotation":
			doc, err := readAnnotation(dec)
			if err != nil {
				return err
			}
			e.Doc = doc
			return nil
		case "complexType":
			ct, err := p.decodeComplexType(se, dec, prefixes)
			if err != nil {
				return err
			}
			e.InlineComplex = ct
			return nil
		case "simpleType":
			st, err := p.decodeSimpleType(se, dec, prefixes)
			if err != nil {
				return err
			}
			e.InlineSimple = st
			return nil
		case "key", "keyref", "unique":
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return e, nil
}

func (p *Parser) decodeAttribute(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Attribute, error) {
	a := &ir.Attribute{
		Name:    attr(se.Attr, "name"),
		Type:    attrQ(se.Attr, "type", prefixes),
		Use:     attr(se.Attr, "use"),
		Default: attr(se.Attr, "default"),
		Fixed:   attr(se.Attr, "fixed"),
	}
	if a.Use == "" {
		a.Use = "optional"
	}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "annotation":
			doc, err := readAnnotation(dec)
			if err != nil {
				return err
			}
			a.Doc = doc
			return nil
		case "simpleType":
			st, err := p.decodeSimpleType(se, dec, prefixes)
			if err != nil {
				return err
			}
			a.SimpleType = st
			return nil
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return a, nil
}

func (p *Parser) decodeAttrGroup(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.AttrGroup, error) {
	ag := &ir.AttrGroup{Name: attr(se.Attr, "name")}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "annotation":
			doc, err := readAnnotation(dec)
			if err != nil {
				return err
			}
			ag.Doc = doc
			return nil
		case "attribute":
			a, err := p.decodeAttribute(se, dec, prefixes)
			if err != nil {
				return err
			}
			ag.Attributes = append(ag.Attributes, a)
			return nil
		case "anyAttribute":
			ag.AnyAttribute = append(ag.AnyAttribute, p.decodeAnyAttribute(se))
			return skipElement(dec)
		case "attributeGroup":
			ref := attrQ(se.Attr, "ref", prefixes)
			ag.Attributes = append(ag.Attributes, &ir.Attribute{Name: "@attrGroupRef", Type: ref})
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return ag, nil
}

func (p *Parser) decodeGroup(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (*ir.Group, error) {
	g := &ir.Group{Name: attr(se.Attr, "name")}
	if err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "annotation":
			doc, err := readAnnotation(dec)
			if err != nil {
				return err
			}
			g.Doc = doc
			return nil
		case "sequence":
			seq, err := p.decodeSequence(dec, prefixes)
			if err != nil {
				return err
			}
			g.Body = seq
			return nil
		case "choice":
			ch, err := p.decodeChoice(dec, prefixes)
			if err != nil {
				return err
			}
			g.Body = ch
			return nil
		case "all":
			all, err := p.decodeAll(dec, prefixes)
			if err != nil {
				return err
			}
			g.Body = all
			return nil
		default:
			return skipElement(dec)
		}
	}); err != nil {
		return nil, err
	}
	return g, nil
}

// decodeSequence reads <xs:sequence>. The returned Particle.Kind is
// ParticleSequence; Seq holds the ordered child particles.
func (p *Parser) decodeSequence(dec *xml.Decoder, prefixes map[string]string) (ir.Particle, error) {
	out := ir.Particle{Kind: ir.ParticleSequence}
	err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "element":
			ref, err := p.decodeLocalElementSE(se, dec, prefixes)
			if err != nil {
				return err
			}
			out.Seq = append(out.Seq, ref)
			return nil
		case "choice":
			ch, err := p.decodeChoice(dec, prefixes)
			if err != nil {
				return err
			}
			out.Seq = append(out.Seq, ch)
			return nil
		case "sequence":
			inner, err := p.decodeSequence(dec, prefixes)
			if err != nil {
				return err
			}
			out.Seq = append(out.Seq, inner)
			return nil
		case "any":
			out.Seq = append(out.Seq, ir.Particle{Kind: ir.ParticleAny, Any: p.decodeAny(se)})
			return skipElement(dec)
		case "group":
			out.Seq = append(out.Seq, p.decodeGroupRef(se, prefixes))
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	})
	return out, err
}

func (p *Parser) decodeChoice(dec *xml.Decoder, prefixes map[string]string) (ir.Particle, error) {
	ch := &ir.Choice{}
	out := ir.Particle{Kind: ir.ParticleChoice, Choice: ch}
	for {
		tok, err := dec.Token()
		if err != nil {
			return out, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "element":
				ref, err := p.decodeLocalElementSE(t, dec, prefixes)
				if err != nil {
					return out, err
				}
				ch.Body = append(ch.Body, ref)
			case "choice":
				inner, err := p.decodeChoice(dec, prefixes)
				if err != nil {
					return out, err
				}
				ch.Body = append(ch.Body, inner)
			case "sequence":
				inner, err := p.decodeSequence(dec, prefixes)
				if err != nil {
					return out, err
				}
				ch.Body = append(ch.Body, inner)
			case "any":
				ch.Body = append(ch.Body, ir.Particle{Kind: ir.ParticleAny, Any: p.decodeAny(t)})
				if err := skipElement(dec); err != nil {
					return out, err
				}
			case "group":
				ch.Body = append(ch.Body, p.decodeGroupRef(t, prefixes))
				if err := skipElement(dec); err != nil {
					return out, err
				}
			default:
				if err := skipElement(dec); err != nil {
					return out, err
				}
			}
		case xml.EndElement:
			return out, nil
		}
	}
}

func (p *Parser) decodeAll(dec *xml.Decoder, prefixes map[string]string) (ir.Particle, error) {
	out := ir.Particle{Kind: ir.ParticleAll}
	err := decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "element":
			ref, err := p.decodeLocalElementSE(se, dec, prefixes)
			if err != nil {
				return err
			}
			out.Seq = append(out.Seq, ref)
			return nil
		case "any":
			out.Seq = append(out.Seq, ir.Particle{Kind: ir.ParticleAny, Any: p.decodeAny(se)})
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	})
	return out, err
}

// decodeLocalElementSE captures the StartElement attributes for an element
// particle then consumes any inline body (annotation / inline type is uncommon
// inside content models but allowed).
func (p *Parser) decodeLocalElementSE(se xml.StartElement, dec *xml.Decoder, prefixes map[string]string) (ir.Particle, error) {
	er := &ir.ElementRef{
		Name:              attr(se.Attr, "name"),
		Type:              attrQ(se.Attr, "type", prefixes),
		Ref:               attrQ(se.Attr, "ref", prefixes),
		MinOccurs:         normOccurs(attr(se.Attr, "minOccurs")),
		MaxOccurs:         normOrUnbounded(attr(se.Attr, "maxOccurs")),
		Nillable:          attr(se.Attr, "nillable") == "true",
		SubstitutionGroup: attrQ(se.Attr, "substitutionGroup", prefixes),
	}
	// Consume the body. Inline types inside content models are unusual but
	// legal; ignore them for the IR (the global element the ref points to
	// carries the type definition).
	if err := skipElement(dec); err != nil {
		return ir.Particle{}, err
	}
	return ir.Particle{Kind: ir.ParticleElement, Element: er, Min: er.MinOccurs, Max: er.MaxOccurs}, nil
}

func (p *Parser) decodeGroupRef(se xml.StartElement, prefixes map[string]string) ir.Particle {
	return ir.Particle{
		Kind: ir.ParticleGroup,
		Group: &ir.GroupRef{
			Ref:       attrQ(se.Attr, "ref", prefixes),
			MinOccurs: normOccurs(attr(se.Attr, "minOccurs")),
			MaxOccurs: normOrUnbounded(attr(se.Attr, "maxOccurs")),
		},
	}
}

func (p *Parser) decodeAny(se xml.StartElement) *ir.Any {
	return &ir.Any{
		Min:             normOccurs(attr(se.Attr, "minOccurs")),
		Max:             normOrUnbounded(attr(se.Attr, "maxOccurs")),
		Namespace:       attr(se.Attr, "namespace"),
		ProcessContents: attr(se.Attr, "processContents"),
	}
}

func (p *Parser) decodeAnyAttribute(se xml.StartElement) ir.AnyAttribute {
	return ir.AnyAttribute{
		Namespace:       attr(se.Attr, "namespace"),
		ProcessContents: attr(se.Attr, "processContents"),
	}
}

func normOrUnbounded(v string) string {
	if v == "" {
		return "1"
	}
	return v
}

func (p *Parser) decodeComplexContent(dec *xml.Decoder, prefixes map[string]string, ct *ir.ComplexType) error {
	return decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "extension":
			ct.ContentModel = ir.ContentModel{Kind: ir.ContentModelComplexContentExtension, Base: attrQ(se.Attr, "base", prefixes)}
			return decodeUntilEnd(dec, func(c2 xml.StartElement) error {
				switch c2.Name.Local {
				case "sequence":
					seq, err := p.decodeSequence(dec, prefixes)
					if err != nil {
						return err
					}
					ct.Content = seq
					return nil
				case "choice":
					ch, err := p.decodeChoice(dec, prefixes)
					if err != nil {
						return err
					}
					ct.Content = ch
					return nil
				case "all":
					all, err := p.decodeAll(dec, prefixes)
					if err != nil {
						return err
					}
					ct.Content = all
					return nil
				case "attribute":
					a, err := p.decodeAttribute(c2, dec, prefixes)
					if err != nil {
						return err
					}
					ct.Attributes = append(ct.Attributes, a)
					return nil
				case "anyAttribute":
					ct.AnyAttribute = append(ct.AnyAttribute, p.decodeAnyAttribute(c2))
					return skipElement(dec)
				case "annotation":
					doc, err := readAnnotation(dec)
					if err != nil {
						return err
					}
					if ct.Doc == "" {
						ct.Doc = doc
					}
					return nil
				default:
					return skipElement(dec)
				}
			})
		case "restriction":
			ct.ContentModel = ir.ContentModel{Kind: ir.ContentModelComplexContentRestriction, Base: attrQ(se.Attr, "base", prefixes)}
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	})
}

func (p *Parser) decodeSimpleContent(dec *xml.Decoder, prefixes map[string]string, ct *ir.ComplexType) error {
	return decodeUntilEnd(dec, func(se xml.StartElement) error {
		switch se.Name.Local {
		case "extension":
			ct.ContentModel = ir.ContentModel{Kind: ir.ContentModelSimpleContentExtension, Base: attrQ(se.Attr, "base", prefixes)}
			return decodeUntilEnd(dec, func(c2 xml.StartElement) error {
				switch c2.Name.Local {
				case "attribute":
					a, err := p.decodeAttribute(c2, dec, prefixes)
					if err != nil {
						return err
					}
					ct.Attributes = append(ct.Attributes, a)
					return nil
				case "anyAttribute":
					ct.AnyAttribute = append(ct.AnyAttribute, p.decodeAnyAttribute(c2))
					return skipElement(dec)
				case "annotation":
					doc, err := readAnnotation(dec)
					if err != nil {
						return err
					}
					if ct.Doc == "" {
						ct.Doc = doc
					}
					return nil
				default:
					return skipElement(dec)
				}
			})
		case "restriction":
			ct.ContentModel = ir.ContentModel{Kind: ir.ContentModelSimpleContentRestriction, Base: attrQ(se.Attr, "base", prefixes)}
			return skipElement(dec)
		default:
			return skipElement(dec)
		}
	})
}
