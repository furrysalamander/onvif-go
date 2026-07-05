package catalog

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestChecksums verifies that every vendored schema file matches the SHA-256
// recorded in checksums.txt. This guards against accidental edits to the
// codegen input.
//
//go:embed checksums.txt
var checksumsRaw string

func TestChecksums(t *testing.T) {
	want := parseChecksums(t, checksumsRaw)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	fsys := c.FS()

	var got []string
	walkErr := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".wsdl") && !strings.HasSuffix(p, ".xsd") {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		sum := sha256.Sum256(data)
		got = append(got, hex.EncodeToString(sum[:])+"  "+filepath.ToSlash(p))
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk: %v", walkErr)
	}

	sort.Strings(got)
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("checksum count mismatch: have %d want %d\nhave:\n%s\nwant:\n%s",
			len(got), len(want), strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("checksum mismatch at index %d:\n have %s\n want %s", i, got[i], want[i])
		}
	}
}

func parseChecksums(t *testing.T, raw string) []string {
	t.Helper()
	var lines []string
	for _, l := range strings.Split(raw, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		// Normalise: "<hex>  <path>".
		fields := strings.Fields(l)
		if len(fields) != 2 {
			t.Fatalf("bad checksum line: %q", l)
		}
		lines = append(lines, fields[0]+"  "+fields[1])
	}
	return lines
}
