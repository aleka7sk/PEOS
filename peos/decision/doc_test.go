package decision

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
// peos/lifecycle is excluded because a Decision's lifecycle consequence is a
// PEOS-004 value, not a PEOS-003 State Assignment; peos/requirement because a
// Decision Subject arrives through core.EngineeringSubjectRef; peos/validation
// because PEOS-004 references Claims by core reference rather than by type;
// peos/relation because PEOS-004 lists "supersedes" and "conflict" as Decision
// Record relationships it models directly, explicitly not as
// relation.Relation specializations (see supersession.go, conflict.go, and
// invalidation.go, each of which documents that choice).
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
	// A package's own external test files (package decision_test) legitimately
	// import the package under test -- that is the standard Go idiom for
	// example_test.go files godoc attaches to this package, not a boundary
	// violation, so self-import is allowed here.
	allowed[peosModulePrefix+"decision"] = true

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

// TestLifecycleDoesNotImportDecision locks the converse direction. The
// tracker's Delegation deferral turns on exactly this: an authority-grant
// aggregate placed in peos/decision would force peos/lifecycle to import it,
// inverting the dependency direction. Asserting the non-import here keeps that
// deferral's premise checkable rather than merely stated.
func TestLifecycleDoesNotImportDecision(t *testing.T) {
	assertDoesNotImport(t, filepath.Join("..", "lifecycle"), "peos/lifecycle",
		peosModulePrefix+"decision",
		"peos/lifecycle is the more foundational package and consumes core.AuthorityRef directly")
}
