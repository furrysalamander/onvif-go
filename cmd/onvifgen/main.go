// onvifgen is the ONVIF WSDL/XSD → Go code generator.
//
// Usage:
//
//	onvifgen <catalog-dir> <out-base> <pkg> [<pkg>...]
//
// It loads each vendored WSDL/XSD via internal/catalog, parses them into the
// onvifgen IR, then emits one Go package per requested package id under
// <out-base>/<pkg>/. Output is committed to the repo and verified for drift
// by CI.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/catalog"
	"github.com/furrysalamander/onvif-go/internal/onvifgen/codegen"
	"github.com/furrysalamander/onvif-go/internal/onvifgen/parser"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: onvifgen <catalog-req-build-tag|ignored> <out-base> <pkg> [<pkg>...]")
		fmt.Fprintln(os.Stderr, "note: the first arg is retained for backward compatibility with the M0 directive; the catalog is embedded.")
		os.Exit(2)
	}
	outBase := os.Args[2]
	emitPkgs := os.Args[3:]

	cat, err := catalog.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "onvifgen: catalog.Load: %v\n", err)
		os.Exit(1)
	}
	p := parser.New(cat)
	loader := parser.NewLoader(p)

	// Walk every WSDL referenced by the ONVIF Phase-1 set so the SymTab sees
	// the full import closure (onvif.xsd, common.xsd, OASIS WS-Notification,
	// WS-Addressing, ...).
	for _, w := range phase1WSDLs {
		if _, err := loader.Load(w); err != nil {
			fmt.Fprintf(os.Stderr, "onvifgen: load %s: %v\n", w, err)
			os.Exit(1)
		}
	}

	tab := codegen.NewSymTab()
	for _, m := range loader.Modules() {
		if err := tab.AddModule(m); err != nil {
			fmt.Fprintf(os.Stderr, "onvifgen: add module %s: %v\n", m.Path, err)
			os.Exit(1)
		}
	}

	g := &codegen.Generator{Tab: tab, EmitPackages: emitPkgs, OutBase: outBase}
	files, err := g.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "onvifgen: generate: %v\n", err)
		os.Exit(1)
	}

	paths := make([]string, 0, len(files))
	for pth := range files {
		paths = append(paths, pth)
	}
	sort.Strings(paths)

	for _, pth := range paths {
		full := filepath.FromSlash(pth)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "onvifgen: mkdir %s: %v\n", filepath.Dir(full), err)
			os.Exit(1)
		}
		if err := os.WriteFile(full, []byte(files[pth]), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "onvifgen: write %s: %v\n", full, err)
			os.Exit(1)
		}
		fmt.Println("wrote", pth)
	}
}

// phase1WSDLs is the set of catalog-relative WSDL paths onvifgen loads so the
// symbol table can resolve the full import closure for any Phase-1 package.
var phase1WSDLs = []string{
	"onvif/ver10/device/wsdl/devicemgmt.wsdl",
	"onvif/ver10/media/wsdl/media.wsdl",
	"onvif/ver20/media/wsdl/media.wsdl",
	"onvif/ver20/ptz/wsdl/ptz.wsdl",
	"onvif/ver10/events/wsdl/event.wsdl",
	"onvif/ver20/imaging/wsdl/imaging.wsdl",
	"onvif/ver10/accessrules/wsdl/accessrules.wsdl",
	"onvif/ver10/actionengine.wsdl",
	"onvif/ver10/advancedsecurity/wsdl/advancedsecurity.wsdl",
	"onvif/ver10/appmgmt/wsdl/appmgmt.wsdl",
	"onvif/ver10/authenticationbehavior/wsdl/authenticationbehavior.wsdl",
	"onvif/ver10/credential/wsdl/credential.wsdl",
	"onvif/ver10/deviceio.wsdl",
	"onvif/ver10/display.wsdl",
	"onvif/ver10/pacs/accesscontrol.wsdl",
	"onvif/ver10/pacs/doorcontrol.wsdl",
	"onvif/ver10/provisioning/wsdl/provisioning.wsdl",
	"onvif/ver10/receiver.wsdl",
	"onvif/ver10/recording.wsdl",
	"onvif/ver10/replay.wsdl",
	"onvif/ver10/schedule/wsdl/schedule.wsdl",
	"onvif/ver10/search.wsdl",
	"onvif/ver10/thermal/wsdl/thermal.wsdl",
	"onvif/ver10/uplink/wsdl/uplink.wsdl",
	"onvif/ver20/analytics/wsdl/analytics.wsdl",
}

// keep strings import used in case future errors need it.
var _ = strings.TrimSpace
