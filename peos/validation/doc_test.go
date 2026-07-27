package validation

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

// TestProductionImportBoundary locks this package's dependency rule: the
// production sources may import only the standard library and peos/core.
//
// peos/relation is excluded because PEOS-006 defines no Artifact Relation.
// peos/lifecycle is excluded because a validation outcome is never a
// Lifecycle State or State Assignment -- the non-import is the structural
// guarantee of that separation. peos/requirement is excluded because
// Requirement and Requirement Artifact Revision Subjects arrive through
// core.EngineeringSubjectRef and Requirement criteria through
// core.CriterionRef, so no PEOS-005 type is needed. peos/decision is
// excluded because PEOS-004 references Claims rather than the reverse, and
// authority is carried as core.AuthorityRef.
func TestProductionImportBoundary(t *testing.T) {
	const allowed = peosModulePrefix + "core"
	forbidden := []string{
		peosModulePrefix + "relation",
		peosModulePrefix + "lifecycle",
		peosModulePrefix + "requirement",
		peosModulePrefix + "decision",
	}

	byFile := parsePackageImports(t, ".", false)
	sawCore := false
	for name, paths := range byFile {
		for _, path := range paths {
			if !strings.HasPrefix(path, peosModulePrefix) {
				// Standard library (or any non-PEOS module) is permitted.
				continue
			}
			if path == allowed {
				sawCore = true
				continue
			}
			t.Errorf("%s imports %q; production sources may import only the standard library and %q", name, path, allowed)
		}
		for _, bad := range forbidden {
			for _, path := range paths {
				if path == bad {
					t.Errorf("%s imports forbidden package %q", name, bad)
				}
			}
		}
	}
	if !sawCore {
		t.Errorf("no production file imports %q; the helper may be parsing the wrong directory", allowed)
	}
}

// TestTestImportBoundary applies the same rule to this package's own test
// sources. A test that reached for peos/requirement or peos/lifecycle would
// prove the boundary is crossable in practice even if production code stays
// clean, so the tests hold themselves to it too.
func TestTestImportBoundary(t *testing.T) {
	allowed := map[string]bool{peosModulePrefix + "core": true}
	// A package's own external test files (package validation_test) legitimately
	// import the package under test -- that is the standard Go idiom for
	// example_test.go files godoc attaches to this package, not a boundary
	// violation, so self-import is allowed here.
	allowed[peosModulePrefix+"validation"] = true

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
				t.Errorf("%s imports %q; test sources may import only the standard library and peos/core", name, path)
			}
		}
	}
}

// TestRequirementDoesNotImportValidation locks the converse direction.
// PEOS-005 §30.1 keeps Requirements independent of validation outcomes, and
// PEOS-006 derives satisfaction from Claims rather than storing it on the
// Requirement. Together with TestProductionImportBoundary this is what makes
// an import cycle between the two packages inexpressible.
func TestRequirementDoesNotImportValidation(t *testing.T) {
	const self = peosModulePrefix + "validation"

	dir := filepath.Join("..", "requirement")
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("peos/requirement not present at %s: %v", dir, err)
	}

	byFile := parsePackageImports(t, dir, true)
	for name, paths := range byFile {
		for _, path := range paths {
			if path == self {
				t.Errorf("peos/requirement/%s imports %q; PEOS-005 must remain independent of validation outcomes", name, self)
			}
		}
	}
}
