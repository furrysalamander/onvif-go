package codegen

import (
	"strings"
	"testing"

	"github.com/furrysalamander/onvif-go/internal/catalog"
	"github.com/furrysalamander/onvif-go/internal/onvifgen/parser"
)

func TestSymTab_HasFloatRange(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	p := parser.New(cat)
	loader := parser.NewLoader(p)
	for _, w := range []string{
		"onvif/ver10/device/wsdl/devicemgmt.wsdl",
		"onvif/ver10/media/wsdl/media.wsdl",
		"onvif/ver20/media/wsdl/media.wsdl",
		"onvif/ver20/ptz/wsdl/ptz.wsdl",
		"onvif/ver10/events/wsdl/event.wsdl",
		"onvif/ver20/imaging/wsdl/imaging.wsdl",
	} {
		if _, err := loader.Load(w); err != nil {
			t.Fatalf("load %s: %v", w, err)
		}
	}
	tab := NewSymTab()
	for _, m := range loader.Modules() {
		if err := tab.AddModule(m); err != nil {
			t.Fatalf("add %s: %v", m.Path, err)
		}
	}
	want := []string{"FloatRange", "DurationRange", "IntList", "IntRange"}
	for _, w := range want {
		if _, ok := tab.Lookup("http://www.onvif.org/ver10/schema", w); !ok {
			t.Errorf("missing %s from tt namespace", w)
		}
	}
	simple, complex, elements, _, _, _ := tab.GroupedByKind("http://www.onvif.org/ver10/schema")
	t.Logf("tt simple=%d complex=%d elements=%d", len(simple), len(complex), len(elements))
	for _, ct := range complex {
		if ct.Name == "FloatRange" {
			t.Logf("FloatRange in tab: %+v", ct)
		}
	}
	// Generate and confirm.
	g := &Generator{Tab: tab, EmitPackages: []string{"tt"}, OutBase: "onvif/schema"}
	files, err := g.Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for path, body := range files {
		if path == "onvif/schema/tt/complextypes_gen.go" {
			if !strings.Contains(body, "type FloatRange struct") {
				t.Errorf("FloatRange missing from generated output (len=%d)", len(body))
			} else {
				t.Logf("FloatRange emitted")
			}
		}
	}
}
