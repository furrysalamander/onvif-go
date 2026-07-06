package codegen

import (
	"fmt"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// fieldShape describes a Go struct field emitted for one schema particle.
type fieldShape struct {
	GoName      string // exported field name (PascalCase)
	GoType      string // Go type expression, e.g. "string", "*tt.QName", "[]trt.VideoEncoderConfiguration"
	XMLTag      string // the `xml:"..."` tag without the surrounding backticks
	Doc         string // single-line doc carried over from <xs:documentation> (best-effort)
	IsExtension bool   // true for xs:any Extension fields
}

// resolver turns IR QNames into Go type references, threading through the
// symbol table for cross-namespace types.
type resolver struct {
	tab    *SymTab
	curPkg string // package we are emitting (for same-namespace short refs)
}

// goTypeOfQName resolves a QName to a Go type string. Built-in XSD types map
// to Go primitives; named symbols resolve to `pkg.Type` (or `pkg.Type` for the
// current package, dropped to just `TypeName` for same-package brevity — but
// to keep the generator simple we always qualify; Go's import of the same
// package is a no-op so we use the bare name for the current package).
// qualifyPkg returns expr as-is if pkg == r.curPkg, else pkg + "." + expr.
// Used to drop the `tt.` prefix when emitting the tt package itself.
func (r *resolver) qualifyPkg(pkg, expr string) string {
	if pkg == r.curPkg {
		return expr
	}
	return pkg + "." + expr
}

func (r *resolver) goTypeOfQName(q ir.QName) (string, bool) {
	if q == (ir.QName{}) {
		return "", false
	}
	if q.NS == "" || q.NS == "http://www.w3.org/2001/XMLSchema" {
		return r.goBuiltin(q.Local)
	}
	pkg, ok := NSPkg[q.NS]
	if !ok {
		return "", false
	}
	if pkg == "tt" && isCoreLocal(q.Local) {
		pkg = "core"
	}
	if r.curPkg == "tt" && isCyclePkg(pkg) {
		return r.qualifyPkg("core", "Extension"), true
	}
	return r.qualifyPkg(pkg, pascal(q.Local)), true
}

func isCyclePkg(pkg string) bool {
	switch pkg {
	case "fc", "bd":
		return true
	}
	return false
}

func isCoreLocal(local string) bool {
	switch local {
	case "QName", "Duration", "Extension", "RawXML":
		return true
	}
	return false
}

// goBuiltin returns the Go type for an XSD builtin, dropping the `core.`
// prefix when emitting the core or tt packages themselves.
func (r *resolver) goBuiltin(local string) (string, bool) {
	goty, _ := goTypeForBuiltin(local)
	if goty == "" {
		return "", false
	}
	if r.curPkg == "tt" || r.curPkg == "core" {
		goty = strings.TrimPrefix(goty, "*"+"core.")
		if strings.HasPrefix(goty, "core.") {
			goty = strings.TrimPrefix(goty, "core.")
		} else if strings.HasPrefix(goty, "*"+"core.") {
			goty = "*" + strings.TrimPrefix(goty, "*"+"core.")
		}
	}
	return goty, true
}

// elementField builds a fieldShape for an element particle (content-model ref).
func (r *resolver) elementField(er *ir.ElementRef) (fieldShape, error) {
	ty, err := r.elementGoType(er)
	if err != nil {
		return fieldShape{}, err
	}
	shape := fieldShape{
		GoName: pascal(er.Name),
		GoType: ty,
	}
	shape.XMLTag = buildElementTag(er.Name, er.MinOccurs, er.MaxOccurs)
	shape.Doc = ""
	return shape, nil
}

// elementGoType returns the Go type for an element particle, taking
// cardinality (pointer/slice) into account. The base element type may come
// from the element's `type` attribute, its `ref` attribute (points to a
// global element whose type we then resolve), or its substitutionGroup.
func (r *resolver) elementGoType(er *ir.ElementRef) (string, error) {
	base, err := r.elementBaseType(er)
	if err != nil {
		return "", err
	}
	return applyCardinality(base, er.MinOccurs, er.MaxOccurs), nil
}

// elementBaseType returns the type expression for a single occurrence of the
// element (no pointer/slice wrapping).
func (r *resolver) elementBaseType(er *ir.ElementRef) (string, error) {
	if er.Type != (ir.QName{}) {
		goty, ok := r.goTypeOfQName(er.Type)
		if !ok {
			return "", fmt.Errorf("unresolved element type %v for %q", er.Type, er.Name)
		}
		return goty, nil
	}
	if er.Ref != (ir.QName{}) {
		// Reference a global element. In Go we represent the element's wire
		// shape by the element's *type*. Look the global element up.
		sym, ok := r.tab.Lookup(er.Ref.NS, er.Ref.Local)
		if !ok || sym.Element == nil {
			return "", fmt.Errorf("unresolved element ref %v", er.Ref)
		}
		// If the global element has a named type, use it.
		if sym.Element.Type != (ir.QName{}) {
			goty, ok := r.goTypeOfQName(sym.Element.Type)
			if !ok {
				return "", fmt.Errorf("unresolved global element type %v for %v", sym.Element.Type, er.Ref)
			}
			return goty, nil
		}
		// If it has an inline complex type, we'd need a generated
		// synthesized name (FooBody). Not handled in M2 for content-model
		// refs — ONVIF Phase 1 elements all have named types.
		return "", fmt.Errorf("global element %v has no named type (inline body unsupported in content models)", er.Ref)
	}
	if er.SubstitutionGroup != (ir.QName{}) {
		return "", fmt.Errorf("substitutionGroup on %q not supported in M2", er.Name)
	}
	return "", fmt.Errorf("element %q has neither type nor ref", er.Name)
}

// applyCardinality wraps a base Go type in a pointer or slice per XSD
// minOccurs/maxOccurs semantics.
func applyCardinality(base, min, max string) string {
	switch {
	case max == "unbounded" || (max != "" && max != "1"):
		return "[]" + base
	case min == "0":
		return "*" + base
	default:
		return base
	}
}

// buildElementTag emits the `xml:"name,opt"` tag for an element particle.
func buildElementTag(name, min, max string) string {
	if max == "unbounded" || (max != "" && max != "1") {
		return name + ",omitempty"
	}
	if min == "0" {
		return name + ",omitempty"
	}
	return name
}

// attributeField builds a fieldShape for an attribute particle. AttrGroup
// references (placeholder attributes with name "@attrGroupRef") are flattened
// by the caller via lookupAttributeGroup.
func (r *resolver) attributeField(a *ir.Attribute) (fieldShape, error) {
	ty, err := r.attributeGoType(a)
	if err != nil {
		return fieldShape{}, err
	}
	tag := a.Name
	if a.Use != "required" {
		tag += ",omitempty"
	}
	return fieldShape{
		GoName: pascal(a.Name),
		GoType: ty,
		XMLTag: tag + ",attr",
	}, nil
}

func (r *resolver) attributeGoType(a *ir.Attribute) (string, error) {
	if a.SimpleType != nil && a.SimpleType.Base != (ir.QName{}) {
		// Anonymous inline restriction — use base type, lose enumerations for
		// now (M2 limitation; enumerations on attributes surface via a follow-
		// up typed re-export).
		goty, ok := r.goTypeOfQName(a.SimpleType.Base)
		if !ok {
			return "", fmt.Errorf("attribute %q inline simpleType base %v unresolved", a.Name, a.SimpleType.Base)
		}
		return goty, nil
	}
	if a.Type != (ir.QName{}) {
		goty, ok := r.goTypeOfQName(a.Type)
		if !ok {
			return "", fmt.Errorf("attribute %q type %v unresolved", a.Name, a.Type)
		}
		return goty, nil
	}
	// No type and no inline type: default to string (xs:anySimpleType).
	return "string", nil
}

// lookupAttributeGroup flattens an attrGroup placeholder attribute into the
// group's members, appending fieldShape entries to out. Recurses on nested
// attrGroup refs (same placeholder convention).
func (r *resolver) lookupAttributeGroup(refNS, refLocal string, out *[]fieldShape, seen map[string]bool) error {
	key := refNS + "|" + refLocal
	if seen[key] {
		return nil // cycle guard
	}
	seen[key] = true
	sym, ok := r.tab.Lookup(refNS, refLocal)
	if !ok || sym.AttrGroup == nil {
		return fmt.Errorf("unresolved attributeGroup {%s}%s", refNS, refLocal)
	}
	ag := sym.AttrGroup
	for _, a := range ag.Attributes {
		if a.Name == "@attrGroupRef" {
			if err := r.lookupAttributeGroup(a.Type.NS, a.Type.Local, out, seen); err != nil {
				return err
			}
			continue
		}
		f, err := r.attributeField(a)
		if err != nil {
			return err
		}
		*out = append(*out, f)
	}
	for range ag.AnyAttribute {
		*out = append(*out, fieldShape{GoName: "AnyAttributes", GoType: "[]xml.Attr", XMLTag: ",any,attr,omitempty"})
	}
	return nil
}

// anyField returns the fieldShape for an <xs:any> particle. The Go field is
// named "Any" to avoid colliding with a sibling typed element named
// "Extension" (a common ONVIF pattern). The tt.Extension type implements
// MarshalXML that ignores its passed StartElement name and writes raw child
// fragments directly, so the xml tag's "Any" name never appears on the wire.
func (r *resolver) anyField(name string) fieldShape {
	extTy := "core.Extension"
	if r.curPkg == "tt" || r.curPkg == "core" {
		extTy = "Extension"
	}
	if name == "" || name == "Extension" {
		name = "Any"
	}
	return fieldShape{
		GoName:      pascal(name),
		GoType:      extTy,
		XMLTag:      "Any,omitempty",
		IsExtension: true,
	}
}

// particleFields expands an ir.Particle into zero or more fieldShape entries.
// Sequences dump their children in order; choices dump each option as a
// separate optional pointer field with doc noting "one of".
func (r *resolver) particleFields(p ir.Particle, fields *[]fieldShape, anySeen *bool) error {
	switch p.Kind {
	case ir.ParticleEmpty:
		return nil
	case ir.ParticleElement:
		if p.Element == nil {
			return nil
		}
		f, err := r.elementField(p.Element)
		if err != nil {
			return err
		}
		*fields = append(*fields, f)
		return nil
	case ir.ParticleSequence, ir.ParticleAll:
		for _, child := range p.Seq {
			if err := r.particleFields(child, fields, anySeen); err != nil {
				return err
			}
		}
		return nil
	case ir.ParticleChoice:
		if p.Choice == nil {
			return nil
		}
		for _, child := range p.Choice.Body {
			f, err := r.choiceOptionField(child)
			if err != nil {
				return err
			}
			*fields = append(*fields, f)
		}
		return nil
	case ir.ParticleAny:
		if p.Any == nil {
			return nil
		}
		if *anySeen {
			return nil // ONVIF schemes use a single Extension slot; collapse duplicates.
		}
		*anySeen = true
		*fields = append(*fields, r.anyField("Extension"))
		return nil
	case ir.ParticleGroup:
		// Group refs (xs:group ref) — flatten the group's body inline.
		if p.Group == nil {
			return nil
		}
		sym, ok := r.tab.Lookup(p.Group.Ref.NS, p.Group.Ref.Local)
		if !ok || sym.Group == nil {
			return fmt.Errorf("unresolved group ref %v", p.Group.Ref)
		}
		return r.particleFields(sym.Group.Body, fields, anySeen)
	}
	return nil
}

// choiceOptionField emits one choice option as an optional pointer field
// (cardinality is per-option).
func (r *resolver) choiceOptionField(p ir.Particle) (fieldShape, error) {
	if p.Kind != ir.ParticleElement || p.Element == nil {
		// Nested sequences/choices inside choices are emitted by recursing via
		// particleFields; but as a single field we can only handle element
		// particles. For M2 we collapse nested composites by recursing and
		// emitting the first element field (good enough for ONVIF's handful
		// of trivial choices).
		var sub []fieldShape
		any := false
		if err := r.particleFields(p, &sub, &any); err != nil {
			return fieldShape{}, err
		}
		if len(sub) == 0 {
			return fieldShape{}, fmt.Errorf("empty choice option")
		}
		return sub[0], nil
	}
	f, err := r.elementField(p.Element)
	if err != nil {
		return fieldShape{}, err
	}
	// Force optional pointer regardless of cardinality (choice options are
	// mutually exclusive ⇒ absent unless chosen).
	if !strings.HasPrefix(f.GoType, "*") && !strings.HasPrefix(f.GoType, "[]") {
		f.GoType = "*" + f.GoType
		f.XMLTag = p.Element.Name + ",omitempty"
	}
	return f, nil
}
