package codegen

import (
	"fmt"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// emitSimpleType emits the Go form of a named SimpleType. Returns "" if the
// type cannot be represented (e.g. anonymous inline with no enumeration).
func emitSimpleType(st *ir.SimpleType, r *resolver) string {
	if st.Name == "" {
		return ""
	}
	goName := pascal(st.Name)
	var b strings.Builder
	if st.Doc != "" {
		b.WriteString(cleanComment(st.Doc, ""))
	}
	if len(st.Enumeration) > 0 {
		// Enum-style: typed string + consts.
		fmt.Fprintf(&b, "type %s string\n\n", goName)
		fmt.Fprintf(&b, "const (\n")
		for _, v := range st.Enumeration {
			fmt.Fprintf(&b, "\t%s%s %s = %q\n", goName, enumConstSuffix(v), goName, v)
		}
		fmt.Fprintf(&b, ")\n\n")
		return b.String()
	}
	// Non-enum simple type: emit as a named wrapper over its base type so the
	// type carries doc + facets but is wire-compatible with the base.
	base, _, err := simpleBaseGoType(st, r)
	if err != nil || base == "" {
		// Fall back to an alias of string to keep the symbol visible.
		fmt.Fprintf(&b, "type %s string\n\n", goName)
		return b.String()
	}
	fmt.Fprintf(&b, "type %s %s\n\n", goName, base)
	return b.String()
}

// simpleBaseGoType returns the Go type for a SimpleType's effective base
// (restriction base, or list item type, or union's first member).
func simpleBaseGoType(st *ir.SimpleType, r *resolver) (string, bool, error) {
	if st.Base != (ir.QName{}) {
		goty, ok := r.goTypeOfQName(st.Base)
		return goty, ok, nil
	}
	if st.ListItemType != (ir.QName{}) {
		goty, ok := r.goTypeOfQName(st.ListItemType)
		return "[]" + goty, ok && goty != "", nil
	}
	if len(st.UnionMembers) > 0 {
		goty, ok := r.goTypeOfQName(st.UnionMembers[0])
		return goty, ok, nil
	}
	return "", false, nil
}

// enumConstSuffix derives a PascalCase const suffix from an enumeration value.
// Numbers get a "Value" prefix; everything else is PascalCased.
func enumConstSuffix(v string) string {
	if v == "" {
		return "Empty"
	}
	if (v[0] >= '0' && v[0] <= '9') || v[0] == '-' || v[0] == '+' {
		return "Value" + sanitizeEnumIdent(v)
	}
	return sanitizeEnumIdent(v)
}

// sanitizeEnumIdent turns an arbitrary enumeration value into a valid Go
// identifier suffix, upper-casing the first rune and replacing non-identifier
// runes with underscores.
func sanitizeEnumIdent(v string) string {
	var b strings.Builder
	for i, r := range v {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= 'a' && r <= 'z':
			if i == 0 {
				b.WriteRune(r - 'a' + 'A')
			} else {
				b.WriteRune(r)
			}
		case r >= '0' && r <= '9' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
