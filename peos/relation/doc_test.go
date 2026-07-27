package relation

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

// assertDoesNotImport fails if any .go file under dir imports target. It is
// how this package locks the directions that would otherwise only be true by
// habit.
func assertDoesNotImport(t *testing.T, dir, pkgLabel, target, why string) {
	t.Helper()

	if _, err := os.Stat(dir); err != nil {
		t.Skipf("%s not present at %s: %v", pkgLabel, dir, err)
	}
	byFile := parsePackageImports(t, dir, true)
	for name, paths := range byFile {
		for _, path := range paths {
			if path == target {
				t.Errorf("%s/%s imports %q; %s", pkgLabel, name, target, why)
			}
		}
	}
}

// TestProductionImportBoundary locks this package's dependency rule: the
// production sources may import only the standard library and peos/core.
//
// PEOS-002 assigns Relation no specialized semantics -- cycle policy,
// endpoint-kind compatibility, and graph traversal all belong to a Relation
// Type's own governing specification or to a future Traceability Model (see
// doc.go). Importing any specialized package here would contradict that
// boundary and would invert the dependency direction every specialized
// wrapper relies on.
func TestProductionImportBoundary(t *testing.T) {
	allowed := map[string]bool{peosModulePrefix + "core": true}

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
	if !saw[peosModulePrefix+"core"] {
		t.Errorf("no production file imports %q; the helper may be parsing the wrong directory", peosModulePrefix+"core")
	}
}

// TestTestImportBoundary holds this package's own tests to the same rule.
func TestTestImportBoundary(t *testing.T) {
	allowed := map[string]bool{peosModulePrefix + "core": true}

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

// TestCoreDoesNotImportRelation locks the converse direction. peos/core must
// remain the foundation beneath peos/relation, not a peer of it; this is what
// makes a cycle between the two inexpressible.
func TestCoreDoesNotImportRelation(t *testing.T) {
	assertDoesNotImport(t, filepath.Join("..", "core"), "peos/core",
		peosModulePrefix+"relation",
		"peos/core is the foundation beneath peos/relation and must not depend on it")
}
