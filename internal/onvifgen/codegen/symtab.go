// Package codegen turns the onvifgen IR into Go source files.
//
// The generator is namespace-driven: every parsed IR Schema contributes its
// named types/elements/attributes to a symbol table keyed by (namespace,
// local-name). A separate namespace→Go-package map decides which packages get
// emitted. Callers (cmd/onvifgen) select which packages to emit via the
// Generator's EmitPackages field, so M2 can ship just tt + tds while the
// generator already knows about trt/tr2/tptz/tev/timg/...
package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// Symbol is a registry entry for one named schema construct.
type Symbol struct {
	Kind      SymbolKind
	Namespace string
	Local     string
	Simple    *ir.SimpleType
	Complex   *ir.ComplexType
	Element   *ir.Element
	Attribute *ir.Attribute
	AttrGroup *ir.AttrGroup
	Group     *ir.Group
	Owner     *ir.Module // module that defined the symbol (for diagnostics)
}

// SymbolKind enumerates the kinds of symbols the table holds.
type SymbolKind int

const (
	SymSimpleType SymbolKind = iota
	SymComplexType
	SymElement
	SymAttribute
	SymAttrGroup
	SymGroup
)

// SymTab maps (namespace, local-name, kind) to schema symbols across all
// loaded modules. Kind-keying matters because ONVIF frequently defines a
// <xs:complexType name="Foo"> alongside a sibling <xs:element name="Foo"
// type="tt:Foo">; without kind partitioning the latter would clobber the
// former in a single-key table.
type SymTab struct {
	byName []map[string]Symbol            // indexed by SymbolKind; outer key is the namespace
	byNS   []map[string]map[string]Symbol // [kind][ns][local] = Symbol (for fast iteration by NS)
}

// NewSymTab builds an empty symbol table.
func NewSymTab() *SymTab {
	const n = 6 // SymSimpleType..SymGroup
	t := &SymTab{
		byName: make([]map[string]Symbol, n),
		byNS:   make([]map[string]map[string]Symbol, n),
	}
	for i := range t.byName {
		t.byName[i] = map[string]Symbol{}
		t.byNS[i] = map[string]map[string]Symbol{}
	}
	return t
}

// AddModule registers every named top-level construct from a module's schema
// (or each inline schema of a WSDL) into the table. Anonymous inline types
// are skipped (they are emitted alongside their owning element).
func (s *SymTab) AddModule(m *ir.Module) error {
	schemas := []*ir.Schema{}
	if m.Kind == ir.ModuleSchema {
		schemas = append(schemas, m.Schema)
	} else {
		schemas = append(schemas, m.WSDL.Types...)
	}
	for _, sc := range schemas {
		if err := s.addSchema(m, sc); err != nil {
			return err
		}
	}
	return nil
}

func (s *SymTab) addSchema(m *ir.Module, sc *ir.Schema) error {
	ns := sc.TargetNS
	add := func(kind SymbolKind, name string, sym Symbol) {
		if name == "" {
			return
		}
		sym.Kind = kind
		sym.Namespace = ns
		sym.Local = name
		sym.Owner = m
		if s.byName[kind][ns+"|"+name] == (Symbol{}) || s.byName[kind][ns+"|"+name].Owner == nil {
			s.byName[kind][ns+"|"+name] = sym
		} else {
			// Later writes win (so the most recently loaded module's symbol
			// takes precedence, mirroring XSD semantics across includes).
			s.byName[kind][ns+"|"+name] = sym
		}
		bucket, ok := s.byNS[kind][ns]
		if !ok {
			bucket = map[string]Symbol{}
			s.byNS[kind][ns] = bucket
		}
		bucket[name] = sym
	}
	for _, st := range sc.SimpleTypes {
		add(SymSimpleType, st.Name, Symbol{Simple: st})
	}
	for _, ct := range sc.ComplexTypes {
		add(SymComplexType, ct.Name, Symbol{Complex: ct})
	}
	for _, e := range sc.Elements {
		add(SymElement, e.Name, Symbol{Element: e})
	}
	for _, a := range sc.Attributes {
		add(SymAttribute, a.Name, Symbol{Attribute: a})
	}
	for _, ag := range sc.AttrGroups {
		add(SymAttrGroup, ag.Name, Symbol{AttrGroup: ag})
	}
	for _, g := range sc.Groups {
		add(SymGroup, g.Name, Symbol{Group: g})
	}
	return nil
}

// Lookup returns the symbol for (ns, local) or false. It tries complex,
// simple, element, attribute, attrGroup, group in that order so the most
// type-like definition wins. When ns is empty the XMLSchema namespace is
// assumed (built-in).
func (s *SymTab) Lookup(ns, local string) (Symbol, bool) {
	if local == "" {
		return Symbol{}, false
	}
	if ns == "" {
		if isXSBuiltin(local) {
			return Symbol{Kind: SymSimpleType, Namespace: "http://www.w3.org/2001/XMLSchema", Local: local}, true
		}
		return Symbol{}, false
	}
	for _, k := range []SymbolKind{SymComplexType, SymSimpleType, SymElement, SymAttribute, SymAttrGroup, SymGroup} {
		if sym, ok := s.byName[k][ns+"|"+local]; ok && sym != (Symbol{}) {
			return sym, true
		}
	}
	return Symbol{}, false
}

// LookupKind returns the symbol of a specific kind, or false.
func (s *SymTab) LookupKind(ns, local string, k SymbolKind) (Symbol, bool) {
	if local == "" {
		return Symbol{}, false
	}
	if ns == "" {
		if isXSBuiltin(local) && k == SymSimpleType {
			return Symbol{Kind: SymSimpleType, Namespace: "http://www.w3.org/2001/XMLSchema", Local: local}, true
		}
		return Symbol{}, false
	}
	sym, ok := s.byName[k][ns+"|"+local]
	if !ok || sym == (Symbol{}) {
		return Symbol{}, false
	}
	return sym, true
}

// Namespaces returns every namespace for which at least one symbol exists,
// sorted.
func (s *SymTab) Namespaces() []string {
	seen := map[string]struct{}{}
	for _, byNS := range s.byNS {
		for ns := range byNS {
			seen[ns] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for ns := range seen {
		out = append(out, ns)
	}
	sort.Strings(out)
	return out
}

// In returns the symbols in a namespace across all kinds, sorted by name
// (ties broken by kind for stability).
func (s *SymTab) In(ns string) []Symbol {
	var out []Symbol
	for _, byNS := range s.byNS {
		bucket, ok := byNS[ns]
		if !ok {
			continue
		}
		for _, sym := range bucket {
			out = append(out, sym)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Local != out[j].Local {
			return out[i].Local < out[j].Local
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

// GroupedByKind returns namespace symbols partitioned by kind for deterministic
// iteration during emission.
func (s *SymTab) GroupedByKind(ns string) (simple []*ir.SimpleType, complex []*ir.ComplexType, elements []*ir.Element, attrs []*ir.Attribute, attrGroups []*ir.AttrGroup, groups []*ir.Group) {
	for k, byNS := range s.byNS {
		bucket, ok := byNS[ns]
		if !ok {
			continue
		}
		for _, sym := range bucket {
			switch SymbolKind(k) {
			case SymSimpleType:
				if sym.Simple != nil {
					simple = append(simple, sym.Simple)
				}
			case SymComplexType:
				if sym.Complex != nil {
					complex = append(complex, sym.Complex)
				}
			case SymElement:
				if sym.Element != nil {
					elements = append(elements, sym.Element)
				}
			case SymAttribute:
				if sym.Attribute != nil {
					attrs = append(attrs, sym.Attribute)
				}
			case SymAttrGroup:
				if sym.AttrGroup != nil {
					attrGroups = append(attrGroups, sym.AttrGroup)
				}
			case SymGroup:
				if sym.Group != nil {
					groups = append(groups, sym.Group)
				}
			}
		}
	}
	return
}

// String of a symbol for diagnostics.
func (sym Symbol) String() string {
	return fmt.Sprintf("{%s}%s (%v)", sym.Namespace, sym.Local, kindName(sym.Kind))
}

func kindName(k SymbolKind) string {
	switch k {
	case SymSimpleType:
		return "simpleType"
	case SymComplexType:
		return "complexType"
	case SymElement:
		return "element"
	case SymAttribute:
		return "attribute"
	case SymAttrGroup:
		return "attributeGroup"
	case SymGroup:
		return "group"
	}
	return "?"
}

// isXSBuiltin reports whether local is a known XSD 1.0 built-in simple type.
func isXSBuiltin(local string) bool {
	_, ok := xsBuiltinKind[local]
	return ok
}

// xsBuiltinKind maps XSD built-in type names to the generator's primitive Go
// kind. Anything not here falls back to a generic string mapping via
// goTypeForBuiltin.
var xsBuiltinKind = map[string]struct{}{
	"anyType":            {},
	"anySimpleType":      {},
	"anyURI":             {},
	"base64Binary":       {},
	"boolean":            {},
	"byte":               {},
	"date":               {},
	"dateTime":           {},
	"decimal":            {},
	"double":             {},
	"duration":           {},
	"ENTITIES":           {},
	"ENTITY":             {},
	"float":              {},
	"gDay":               {},
	"gMonth":             {},
	"gMonthDay":          {},
	"gYear":              {},
	"gYearMonth":         {},
	"hexBinary":          {},
	"ID":                 {},
	"IDREF":              {},
	"IDREFS":             {},
	"int":                {},
	"integer":            {},
	"language":           {},
	"long":               {},
	"Name":               {},
	"NCName":             {},
	"negativeInteger":    {},
	"NMTOKEN":            {},
	"NMTOKENS":           {},
	"nonNegativeInteger": {},
	"nonPositiveInteger": {},
	"normalizedString":   {},
	"NOTATION":           {},
	"positiveInteger":    {},
	"QName":              {},
	"short":              {},
	"string":             {},
	"time":               {},
	"token":              {},
	"unsignedByte":       {},
	"unsignedInt":        {},
	"unsignedLong":       {},
	"unsignedShort":      {},
}

// goTypeForBuiltin returns the Go type string for an XSD builtin. Returns
// ("string", false) for unknown builtins.
func goTypeForBuiltin(local string) (goty string, needsImport bool) {
	switch local {
	case "string", "normalizedString", "token", "language", "Name", "NCName",
		"anyURI", "ID", "IDREF", "NMTOKEN", "ENTITY", "NOTATION":
		return "string", false
	case "boolean":
		return "bool", false
	case "base64Binary", "hexBinary":
		return "[]byte", false
	case "dateTime", "date", "time", "gYear", "gMonth", "gDay", "gMonthDay", "gYearMonth":
		return "time.Time", true
	case "duration":
		return "*" + pkgIdent("core") + ".Duration", true
	case "decimal":
		return "float64", false
	case "float":
		return "float32", false
	case "double":
		return "float64", false
	case "integer", "int", "short", "byte", "nonNegativeInteger", "positiveInteger",
		"negativeInteger", "nonPositiveInteger":
		return "int32", false
	case "long", "unsignedInt", "unsignedShort", "unsignedByte", "unsignedLong":
		return "int64", false
	case "ENTITIES", "IDREFS", "NMTOKENS":
		return "[]string", false
	case "QName":
		return "*" + pkgIdent("core") + ".QName", true
	case "anyType", "anySimpleType":
		return "xml.RawMessage", true
	}
	return "string", false
}

// pkgIdent converts a package id ("tt") to its Go identifier (unchanged for
// now; kept as a single point of mutation).
func pkgIdent(id string) string { return id }

// importPath returns the full Go import path for an onvif schema subpackage.
func importPath(id string) string {
	return "github.com/furrysalamander/onvif-go/onvif/schema/" + id
}

// cleanComment collapses whitespace in xs:documentation text for emission as a
// Go doc comment. Lines are indented by indent so they read well inside struct
// comments.
func cleanComment(doc, indent string) string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return ""
	}
	var b strings.Builder
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString(indent)
		b.WriteString("// ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
