package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// emitComplexType emits the Go struct for a named complexType. Returns the
// source text (without the package/import header) or "" if the type is
// anonymous (inline in an element, handled by the element emitter).
func emitComplexType(ct *ir.ComplexType, r *resolver) string {
	if ct.Name == "" {
		return ""
	}
	goName := pascal(ct.Name)
	var b strings.Builder
	if ct.Doc != "" {
		b.WriteString(cleanComment(ct.Doc, ""))
	}
	fmt.Fprintf(&b, "type %s struct {\n", goName)

	// Content model → fields.
	var fields []fieldShape
	anySeen := false
	if ct.Content.Kind != ir.ParticleEmpty {
		_ = r.particleFields(ct.Content, &fields, &anySeen)
	}
	// simpleContent extension: emit a Value field carrying the base type as
	// chardata. complexContent extension's base type fields surface via the
	// base struct's own members; we embed the base type by its Go name when
	// we can resolve it.
	switch ct.ContentModel.Kind {
	case ir.ContentModelSimpleContentExtension, ir.ContentModelSimpleContentRestriction:
		if ct.ContentModel.Base != (ir.QName{}) {
			if base, ok := r.goTypeOfQName(ct.ContentModel.Base); ok {
				fields = append([]fieldShape{{
					GoName: "Value",
					GoType: base,
					XMLTag: ",chardata",
				}}, fields...)
			}
		}
	case ir.ContentModelComplexContentExtension:
		if ct.ContentModel.Base != (ir.QName{}) {
			if base, ok := r.goTypeOfQName(ct.ContentModel.Base); ok {
				// Embed the base type anonymously when it's a plain type
				// name. Slice/pointer base types can't be anonymous embeds
				// (Go syntax error), so give them a field name.
				if strings.HasPrefix(base, "[") || strings.HasPrefix(base, "*") {
					fields = append([]fieldShape{{
						GoName: "Base",
						GoType: base,
						XMLTag: ",omitempty",
					}}, fields...)
				} else {
					fields = append([]fieldShape{{
						GoName: "",
						GoType: base,
						XMLTag: ",omitempty",
					}}, fields...)
				}
			}
		}
	}

	// Attributes (flattening attrGroup placeholders).
	for _, a := range ct.Attributes {
		if a.Name == "@attrGroupRef" {
			_ = r.lookupAttributeGroup(a.Type.NS, a.Type.Local, &fields, map[string]bool{})
			continue
		}
		f, err := r.attributeField(a)
		if err == nil {
			fields = append(fields, f)
		}
	}
	if len(ct.AnyAttribute) > 0 && !anyAttrInFields(fields) {
		fields = append(fields, fieldShape{
			GoName: "AnyAttributes",
			GoType: "[]xml.Attr",
			XMLTag: ",any,attr,omitempty",
		})
	}

	// Emit fields with stable ordering as produced (matches XSD source order).
	for _, f := range fields {
		emitField(&b, f)
	}
	b.WriteString("}\n\n")
	return b.String()
}

// anyAttrInFields reports whether the field list already contains a
// catch-all anyAttribute slot.
func anyAttrInFields(fields []fieldShape) bool {
	for _, f := range fields {
		if strings.HasSuffix(f.XMLTag, ",attr,omitempty") || strings.HasSuffix(f.XMLTag, ",any,attr,omitempty") {
			if f.GoName == "AnyAttributes" {
				return true
			}
		}
	}
	return false
}

var baseFieldCount int

func emitStruct(b *strings.Builder, goName string, fields []fieldShape) {
	fmt.Fprintf(b, "type %s struct {\n", goName)
	baseFieldCount = 0
	for _, f := range fields {
		emitField(b, f)
	}
	b.WriteString("}\n\n")
}

func emitField(b *strings.Builder, f fieldShape) {
	if f.Doc != "" {
		b.WriteString(cleanComment(f.Doc, "\t"))
	}
	if f.GoName == "" {
		if strings.HasPrefix(f.GoType, "[") || strings.HasPrefix(f.GoType, "*") || strings.Contains(f.GoType, ".") {
			baseFieldCount++
			name := fmt.Sprintf("Base%d", baseFieldCount)
			fmt.Fprintf(b, "\t%s %s `xml:\"%s\"`\n", name, f.GoType, f.XMLTag)
			return
		}
		fmt.Fprintf(b, "\t%s `xml:\"%s\"`\n", f.GoType, f.XMLTag)
		return
	}
	fmt.Fprintf(b, "\t%s %s `xml:\"%s\"`\n", f.GoName, f.GoType, f.XMLTag)
}

// emitElement emits a top-level element wrapper. If the element references a
// named complex type, the wrapper is just an alias so callers can write
// `*tt.GetServicesRequest` etc. If the element has an inline anonymous
// complex type, emit a struct synthesised from the inline body (named
// <ElementName>).
func emitElement(e *ir.Element, r *resolver) string {
	if e.Name == "" {
		return ""
	}
	goName := pascal(e.Name)

	// Skip emitting a same-name alias when the element references a named
	// type in its own package whose Go name matches the element's: Go can't
	// represent `type Foo Foo`. Callers use the named type directly.
	if e.Type != (ir.QName{}) && e.InlineComplex == nil && e.InlineSimple == nil {
		if pkg, ok := NSPkg[e.Type.NS]; ok && pkg == r.curPkg {
			if pascal(e.Type.Local) == goName {
				return ""
			}
		}
	}

	var b strings.Builder
	if e.Doc != "" {
		b.WriteString(cleanComment(e.Doc, ""))
	}
	if e.InlineComplex != nil {
		// Synthesise a struct typedef out of the inline body.
		fmt.Fprintf(&b, "type %s struct {\n", goName)
		var fields []fieldShape
		anySeen := false
		_ = r.particleFields(e.InlineComplex.Content, &fields, &anySeen)
		for _, a := range e.InlineComplex.Attributes {
			if a.Name == "@attrGroupRef" {
				_ = r.lookupAttributeGroup(a.Type.NS, a.Type.Local, &fields, map[string]bool{})
				continue
			}
			f, err := r.attributeField(a)
			if err == nil {
				fields = append(fields, f)
			}
		}
		if len(e.InlineComplex.AnyAttribute) > 0 && !anyAttrInFields(fields) {
			fields = append(fields, fieldShape{
				GoName: "AnyAttributes",
				GoType: "[]xml.Attr",
				XMLTag: ",any,attr,omitempty",
			})
		}
		// Embed the inline complex type's named names if it has one (rare).
		if e.InlineComplex.Name != "" {
			fmt.Fprintf(&b, "\t%s\n", pascal(e.InlineComplex.Name))
		}
		for _, f := range fields {
			emitField(&b, f)
		}
		b.WriteString("}\n\n")
		return b.String()
	}
	if e.InlineSimple != nil {
		base, _, _ := simpleBaseGoType(e.InlineSimple, r)
		if base == "" {
			base = "string"
		}
		fmt.Fprintf(&b, "type %s %s\n\n", goName, base)
		return b.String()
	}
	if e.Type != (ir.QName{}) {
		goty, ok := r.goTypeOfQName(e.Type)
		if ok {
			fmt.Fprintf(&b, "type %s %s\n\n", goName, goty)
			return b.String()
		}
	}
	// Unknown shape: declare a placeholder struct so the package still builds.
	fmt.Fprintf(&b, "type %s struct{}\n\n", goName)
	return b.String()
}

// sortComplex sorts complex types by name for deterministic emission.
func sortComplex(types []*ir.ComplexType) {
	sort.SliceStable(types, func(i, j int) bool { return types[i].Name < types[j].Name })
}
