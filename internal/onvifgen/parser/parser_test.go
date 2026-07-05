package parser

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/furrysalamander/onvif-go/internal/catalog"
	"github.com/furrysalamander/onvif-go/internal/onvifgen/dump"
)

var update = flag.Bool("update", false, "regenerate the golden IR dump")

// phase1WSDLs are the catalog-relative paths of every Phase-1 ONVIF WSDL.
var phase1WSDLs = []string{
	"onvif/ver10/device/wsdl/devicemgmt.wsdl",
	"onvif/ver10/media/wsdl/media.wsdl",
	"onvif/ver20/media/wsdl/media.wsdl",
	"onvif/ver20/ptz/wsdl/ptz.wsdl",
	"onvif/ver10/events/wsdl/event.wsdl",
	"onvif/ver20/imaging/wsdl/imaging.wsdl",
}

// TestParseGolden loads the entire Phase-1 WSDL set via the catalog and
// compares a stable textual summary against testdata/ir_dump.txt. Use
// `-update` to regenerate the golden file after intentional parser changes.
func TestParseGolden(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load: %v", err)
	}
	p := New(cat)
	loader := NewLoader(p)

	// Find every WSDL reachable from the catalog (including OASIS ones imported
	// by events.wsdl) so the golden reflects the full import closure.
	wsdls := append([]string{}, phase1WSDLs...)
	// Events WSDL imports two OASIS WSDLs; include their own load too via the
	// transitive walk the Loader does automatically. Adding explicit paths here
	// would duplicate cache hits harmlessly.
	for _, w := range wsdls {
		if _, err := loader.Load(w); err != nil {
			t.Fatalf("Load %s: %v", w, err)
		}
	}

	// Stable ordering: by module path.
	paths := make([]string, 0, len(loader.parsed))
	for pth := range loader.parsed {
		paths = append(paths, pth)
	}
	sort.Strings(paths)

	var out bytes.Buffer
	for _, pth := range paths {
		dump.Module(&out, loader.parsed[pth])
		out.WriteByte('\n')
	}

	goldenPath := filepath.Join("testdata", "ir_dump.txt")
	if *update {
		if err := os.WriteFile(goldenPath, out.Bytes(), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("golden updated: %s (%d bytes)", goldenPath, out.Len())
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create)", goldenPath, err)
	}
	if !bytes.Equal(bytes.TrimSpace(out.Bytes()), bytes.TrimSpace(golden)) {
		t.Fatalf("IR dump differs from golden %s\n--- have (%d bytes) ---\n%s\n--- want (%d bytes) ---\n%s",
			goldenPath, out.Len(), strings.TrimSpace(out.String()), len(golden), strings.TrimSpace(string(golden)))
	}
}

// TestParse_NoErrors is a belt-and-braces check that loading each Phase-1
// WSDL + its dependency closure succeeds without parser errors.
func TestParse_NoErrors(t *testing.T) {
	for _, w := range phase1WSDLs {
		t.Run(w, func(t *testing.T) {
			cat, err := catalog.Load()
			if err != nil {
				t.Fatalf("catalog.Load: %v", err)
			}
			loader := NewLoader(New(cat))
			if _, err := loader.Load(w); err != nil {
				t.Fatalf("Load %s: %v", w, err)
			}
		})
	}
}
