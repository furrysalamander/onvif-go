// Package dump renders an IR module set as deterministic text for golden-file
// comparison and human inspection during generator development.
package dump

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// Module dumps one module's high-level shape. It is deliberately summary -
// enough to catch parser regressions without being so verbose that golden
// updates churn on every cosmetic change.
func Module(w io.Writer, m *ir.Module) {
	fmt.Fprintf(w, "=== %s ===\n", m.Path)
	switch m.Kind {
	case ir.ModuleWSDL:
		dumpWSDL(w, m.WSDL)
	case ir.ModuleSchema:
		dumpSchema(w, m.Schema)
	}
}

func dumpWSDL(w io.Writer, wd *ir.WSDL) {
	fmt.Fprintf(w, "kind: wsdl\n")
	fmt.Fprintf(w, "targetNS: %s\n", wd.TargetNS)
	if len(wd.Imports) > 0 {
		fmt.Fprintln(w, "imports:")
		for _, i := range wd.Imports {
			fmt.Fprintf(w, "  - ns=%s loc=%s\n", i.Namespace, i.Location)
		}
	}
	fmt.Fprintf(w, "types: %d inline schema(s)\n", len(wd.Types))
	for _, s := range wd.Types {
		fmt.Fprintf(w, "  schema targetNS=%s simple=%d complex=%d elements=%d attributes=%d attrGroups=%d groups=%d\n",
			s.TargetNS, len(s.SimpleTypes), len(s.ComplexTypes), len(s.Elements), len(s.Attributes), len(s.AttrGroups), len(s.Groups))
		if len(s.Imports) > 0 {
			for _, imp := range s.Imports {
				fmt.Fprintf(w, "    import ns=%s loc=%s\n", imp.Namespace, imp.SchemaLocation)
			}
		}
		if len(s.Includes) > 0 {
			for _, inc := range s.Includes {
				fmt.Fprintf(w, "    include loc=%s\n", inc.SchemaLocation)
			}
		}
	}
	fmt.Fprintf(w, "messages: %d\n", len(wd.Messages))
	// Operations are the headline artifact; list them sorted with I/O/F counts.
	for _, pt := range wd.PortTypes {
		fmt.Fprintf(w, "portType %q operations=%d\n", pt.Name, len(pt.Operations))
		ops := make([]*ir.Operation, len(pt.Operations))
		copy(ops, pt.Operations)
		sort.SliceStable(ops, func(i, j int) bool { return ops[i].Name < ops[j].Name })
		for _, op := range ops {
			in, out := "-", "-"
			if op.Input != nil {
				in = qnameOrDash(op.Input.Message)
			}
			if op.Output != nil {
				out = qnameOrDash(op.Output.Message)
			}
			fmt.Fprintf(w, "  op %q in=%s out=%s faults=%d action=%s\n", op.Name, in, out, len(op.Faults), soapActionFor(wd, op.Name))
		}
	}
	for _, b := range wd.Bindings {
		fmt.Fprintf(w, "binding %q portType=%s soap=%s style=%s ops=%d\n", b.Name, b.PortType, b.Type, b.Style, len(b.Operations))
	}
	for _, svc := range wd.Services {
		fmt.Fprintf(w, "service %q\n", svc.Name)
		for _, p := range svc.Ports {
			fmt.Fprintf(w, "  port %q binding=%s addr=%s\n", p.Name, p.Binding, p.Address)
		}
	}
}

// soapActionFor returns the SOAP action recorded for the named operation in
// the WSDL's first binding, or "". ONVIF SOAP actions use the target
// namespace + "/" + op name and the binding captures them; we synthesise the
// canonical fallback if none recorded.
func soapActionFor(w *ir.WSDL, opName string) string {
	for _, b := range w.Bindings {
		for _, bo := range b.Operations {
			if bo.Name == opName && bo.SOAPAction != "" {
				return bo.SOAPAction
			}
		}
	}
	return w.TargetNS + "/" + opName
}

func dumpSchema(w io.Writer, s *ir.Schema) {
	fmt.Fprintf(w, "kind: schema\n")
	fmt.Fprintf(w, "targetNS: %s\n", s.TargetNS)
	if len(s.Imports) > 0 {
		fmt.Fprintln(w, "imports:")
		for _, i := range s.Imports {
			fmt.Fprintf(w, "  - ns=%s loc=%s\n", i.Namespace, i.SchemaLocation)
		}
	}
	if len(s.Includes) > 0 {
		fmt.Fprintln(w, "includes:")
		for _, i := range s.Includes {
			fmt.Fprintf(w, "  - loc=%s\n", i.SchemaLocation)
		}
	}
	fmt.Fprintf(w, "simpleTypes: %d\n", len(s.SimpleTypes))
	fmt.Fprintf(w, "complexTypes: %d\n", len(s.ComplexTypes))
	fmt.Fprintf(w, "elements: %d\n", len(s.Elements))
	fmt.Fprintf(w, "attributes: %d\n", len(s.Attributes))
	fmt.Fprintf(w, "attrGroups: %d\n", len(s.AttrGroups))
	fmt.Fprintf(w, "groups: %d\n", len(s.Groups))
	// List a sample of enumerations + attribute-group counts so the golden
	// catches restriction/enumeration regressions.
	enumCount := 0
	for _, st := range s.SimpleTypes {
		enumCount += len(st.Enumeration)
	}
	fmt.Fprintf(w, "enumerations: %d total\n", enumCount)
	anyCount := 0
	for _, ct := range s.ComplexTypes {
		anyCount += countAny(ct.Content)
		anyCount += len(ct.AnyAttribute)
	}
	fmt.Fprintf(w, "wildcards: %d total (xs:any + anyAttribute)\n", anyCount)
}

// countAny counts <xs:any> particles in a content model subtree.
func countAny(p ir.Particle) int {
	switch p.Kind {
	case ir.ParticleAny:
		return 1
	case ir.ParticleSequence, ir.ParticleAll:
		n := 0
		for _, c := range p.Seq {
			n += countAny(c)
		}
		return n
	case ir.ParticleChoice:
		n := 0
		for _, c := range p.Choice.Body {
			n += countAny(c)
		}
		return n
	}
	return 0
}

func qnameOrDash(q ir.QName) string {
	if q == (ir.QName{}) {
		return "-"
	}
	return q.String()
}

// Lines returns the dump as a single string (used by the golden test).
func Lines(m *ir.Module) string {
	var b strings.Builder
	Module(&b, m)
	return b.String()
}
