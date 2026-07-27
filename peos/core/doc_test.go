package core

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const peosModulePrefix = "github.com/aleka7sk/PEOS/peos/"

// parsePackageImports parses every .go file in dir and returns, per file,
// the set of import paths it declares. It uses go/parser from the standard
// library rather than invoking the Go toolchain as a subprocess, so the
// check runs as an ordinary unit test with no external process, no network,
// and no dependency on a built binary being present.
//
// This mirrors the helper peos/validation, peos/quality, peos/runtime, and
// peos/template already carry. Go cannot share an unexported test helper
// across package boundaries, so each package keeps its own copy; the
// duplication is the established repository convention rather than an
// oversight.
func parsePackageImports(t *testing.T, dir string, includeTests bool) map[string][]string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	result := make(map[string][]string)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		isTest := strings.HasSuffix(name, "_test.go")
		if isTest && !includeTests {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		paths := make([]string, 0, len(file.Imports))
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquote import %s: %v", name, spec.Path.Value, err)
			}
			paths = append(paths, path)
		}
		result[name] = paths
	}
	if len(result) == 0 {
		t.Fatalf("no .go files found in %s", dir)
	}
	return result
}

// TestProductionImportBoundary locks the single most load-bearing structural
// fact in this module's dependency graph: peos/core imports no PEOS package
// at all. Every other package depends on core, so a single import here in any
// direction would make a cycle expressible. Packet L.0.C added this test
// because the property was true but unenforced -- nothing would have failed if
// a later packet had introduced such an import.
func TestProductionImportBoundary(t *testing.T) {
	allowed := map[string]bool{}

	byFile := parsePackageImports(t, ".", false)
	saw := map[string]bool{}
	for name, paths := range byFile {
		for _, path := range paths {
			if !strings.HasPrefix(path, peosModulePrefix) {
				// Standard library (or any non-PEOS module) is permitted.
				continue
			}
			if allowed[path] {
				saw[path] = true
				continue
			}
			t.Errorf("%s imports %q, which this package's dependency rule does not permit", name, path)
		}
	}
}

// TestTestImportBoundary holds this package's own tests to the same rule. A
// core test that reached for a downstream package would prove the boundary is
// crossable in practice even while production sources stay clean.
func TestTestImportBoundary(t *testing.T) {
	allowed := map[string]bool{}

	byFile := parsePackageImports(t, ".", true)
	for name, paths := range byFile {
		if !strings.HasSuffix(name, "_test.go") {
			continue
		}
		for _, path := range paths {
			if !strings.HasPrefix(path, peosModulePrefix) {
				continue
			}
			if !allowed[path] {
				t.Errorf("%s imports %q; test sources are held to the same dependency rule as production sources", name, path)
			}
		}
	}
}

// TestNoPeosImportsAnywhereInCore states the same rule as a single, direct
// assertion, independent of the table-driven form above, so the guarantee
// survives a refactor of either test.
func TestNoPeosImportsAnywhereInCore(t *testing.T) {
	byFile := parsePackageImports(t, ".", true)
	for name, paths := range byFile {
		for _, path := range paths {
			if strings.HasPrefix(path, peosModulePrefix) {
				t.Errorf("%s imports %q; peos/core is the foundation and must depend on no PEOS package", name, path)
			}
		}
	}
}
