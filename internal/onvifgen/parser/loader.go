package parser

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

// Loader parses the catalog's WSDLs + their transitive XSD imports into a set
// of IR modules, caching by catalog path so each file is parsed at most once.
type Loader struct {
	p       *Parser
	parsed  map[string]*ir.Module // keyed by catalog path
	loading map[string]bool       // cycle guard
}

// NewLoader builds a Loader over the given catalog-backed parser.
func NewLoader(p *Parser) *Loader {
	return &Loader{
		p:       p,
		parsed:  map[string]*ir.Module{},
		loading: map[string]bool{},
	}
}

// Load opens the WSDL at the catalog-relative path wFile, parses it and all
// transitive dependencies, and returns the module for wFile.
func (l *Loader) Load(wFile string) (*ir.Module, error) {
	if m, ok := l.parsed[wFile]; ok {
		return m, nil
	}
	if l.loading[wFile] {
		return nil, fmt.Errorf("loader: cyclic import of %s", wFile)
	}
	l.loading[wFile] = true
	defer delete(l.loading, wFile)

	data, err := fs.ReadFile(l.p.Load().FS(), wFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", wFile, err)
	}
	// Detect WSDL vs XSD by extension.
	kind := classify(wFile)
	if kind == ir.ModuleWSDL {
		w, err := l.p.ParseWSDL(wFile, filepath.Dir(wFile), strings.NewReader(string(data)))
		if err != nil {
			return nil, err
		}
		if err := l.loadWSDLImports(w); err != nil {
			return nil, err
		}
		m := &ir.Module{Path: wFile, Kind: ir.ModuleWSDL, TargetNS: w.TargetNS, WSDL: w}
		l.parsed[wFile] = m
		return m, nil
	}
	s, err := l.p.ParseSchema(wFile, filepath.Dir(wFile), strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	if err := l.loadSchemaImports(s); err != nil {
		return nil, err
	}
	m := &ir.Module{Path: wFile, Kind: ir.ModuleSchema, TargetNS: s.TargetNS, Schema: s}
	l.parsed[wFile] = m
	return m, nil
}

func (l *Loader) loadWSDLImports(w *ir.WSDL) error {
	for _, imp := range w.Imports {
		if imp.Location == "" {
			continue
		}
		p, ok := l.resolve("", imp.Location, filepath.Dir(w.File))
		if !ok {
			continue // best-effort: skip unresolvable WSDL imports
		}
		if _, err := l.Load(p); err != nil {
			return fmt.Errorf("wsdl import %s: %w", p, err)
		}
	}
	for _, s := range w.Types {
		if err := l.loadSchemaImports(s); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadSchemaImports(s *ir.Schema) error {
	for _, inc := range s.Includes {
		p, ok := l.resolve("", inc.SchemaLocation, filepath.Dir(s.File))
		if !ok {
			continue
		}
		if _, err := l.Load(p); err != nil {
			return fmt.Errorf("schema include %s: %w", p, err)
		}
	}
	for _, imp := range s.Imports {
		p, ok := l.resolve(imp.Namespace, imp.SchemaLocation, filepath.Dir(s.File))
		if !ok {
			continue
		}
		if _, err := l.Load(p); err != nil {
			return fmt.Errorf("schema import %s: %w", p, err)
		}
	}
	return nil
}

// resolve turns a (namespace, location) reference into a catalog-relative
// path. Location may be relative (joined against importerDir), an http(s)
// URL, or empty (resolution by namespace only).
func (l *Loader) resolve(namespace, location, importerDir string) (string, bool) {
	cat := l.p.Load()
	if location == "" {
		// Namespace-only import (some WSDLs omit schemaLocation).
		if p, ok := cat.Resolve(namespace, ""); ok {
			return p, true
		}
		return "", false
	}
	if isURL(location) {
		// First try catalog namespace/SystemID resolution against the URL,
		// then against the namespace.
		if p, ok := cat.Resolve(namespace, location); ok {
			return p, true
		}
		// Pure namespace last resort.
		if p, ok := cat.Resolve(namespace, ""); ok {
			return p, true
		}
		return "", false
	}
	// Relative path: join against importerDir and clean.
	joined := path.Clean(path.Join(importerDir, location))
	// Verify the file exists in the catalog FS.
	if _, err := fs.Stat(cat.FS(), joined); err == nil {
		return joined, true
	}
	return "", false
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func classify(file string) ir.ModuleKind {
	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".wsdl" {
		return ir.ModuleWSDL
	}
	return ir.ModuleSchema
}
