package catalog

import (
	"io"
	"testing"
)

func TestLoad_ResolvesOnvifNamespaces(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	cases := []struct {
		ns, loc, want string
	}{
		{"http://www.onvif.org/ver10/device/wsdl", "", "onvif/ver10/device/wsdl/devicemgmt.wsdl"},
		{"http://www.onvif.org/ver10/schema", "", "onvif/ver10/schema/onvif.xsd"},
		{"http://docs.oasis-open.org/wsn/b-2", "", "external/oasis/wsn/b-2.xsd"},
		{"", "http://docs.oasis-open.org/wsn/b-2.xsd", "external/oasis/wsn/b-2.xsd"},
		{"http://www.w3.org/2005/08/addressing", "", "external/w3c/2005/08/addressing/ws-addr.xsd"},
		{"", "http://www.w3.org/2001/xml.xsd", "external/w3c/2001/xml.xsd"},
		{"", "https://www.w3.org/2003/05/soap-envelope", "external/w3c/2003/05/soap-envelope/soap-envelope.xsd"},
	}
	for _, tc := range cases {
		got, ok := c.Resolve(tc.ns, tc.loc)
		if !ok {
			t.Errorf("Resolve(%q,%q): not found", tc.ns, tc.loc)
			continue
		}
		if got != tc.want {
			t.Errorf("Resolve(%q,%q) = %q want %q", tc.ns, tc.loc, got, tc.want)
		}
	}
}

func TestLoad_OpenReadsFile(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	rc, err := c.Open("onvif/ver10/schema/common.xsd")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("read empty file")
	}
}

func TestResolve_UnknownReturnsFalse(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := c.Resolve("http://example.invalid", ""); ok {
		t.Fatal("expected not found")
	}
}

// Relative schemaLocation paths inside the vendored tree (e.g. onvif.xsd
// includes "common.xsd" relative to its own directory) must be resolvable by
// the generator by joining against the importer's directory, not by the
// catalog. Resolve() should report false for raw relative paths so callers
// know to do that join.
func TestResolve_RelativePathNotInCatalog(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := c.Resolve("", "common.xsd"); ok {
		t.Fatal("relative path should not resolve via catalog")
	}
}
