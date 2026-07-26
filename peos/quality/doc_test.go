package quality

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const peosModulePrefix = "github.com/aleka7sk/PEOS/peos/"

// parsePackageImports parses every .go file in dir and returns, per file, the
// set of import paths it declares. It uses go/parser from the standard library
// rather than invoking the Go toolchain as a subprocess, so the check runs as
// an ordinary unit test with no external process, no network, and no
// dependency on a built binary being present. Modelled on
// validation/doc_test.go's helper of the same name.
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
// production sources may import only the standard library, peos/core, and
// peos/validation.
//
// peos/validation is permitted because PEOS-007's two specializations compose
// PEOS-006 types (Packet I.2). Packet I.1's sources do not actually import it
// yet, and this test deliberately does not require them to -- an unused
// dependency is not added to match a future graph.
//
// peos/relation is excluded because PEOS-007 defines no Artifact Relation.
// peos/lifecycle is excluded because "A Quality Claim does not itself assign a
// Lifecycle State or State Assignment" -- the non-import is the structural
// guarantee of that separation. peos/requirement is excluded because a Quality
// Constraint must never be silently treated as a Requirement, and because
// Requirement subjects and criteria arrive through core.EngineeringSubjectRef
// and core.CriterionRef, so no PEOS-005 type is needed. peos/decision is
// excluded because a quality outcome is never governance authority.
func TestProductionImportBoundary(t *testing.T) {
	allowed := map[string]bool{
		peosModulePrefix + "core":       true,
		peosModulePrefix + "validation": true,
	}
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
			if allowed[path] {
				if path == peosModulePrefix+"core" {
					sawCore = true
				}
				continue
			}
			t.Errorf("%s imports %q; production sources may import only the standard library, peos/core, and peos/validation", name, path)
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
		t.Errorf("no production file imports %q; the helper may be parsing the wrong directory", peosModulePrefix+"core")
	}
}

// TestTestImportBoundary applies the same rule to this package's own test
// sources. A test that reached for peos/requirement or peos/lifecycle would
// prove the boundary is crossable in practice even if production code stays
// clean, so the tests hold themselves to it too.
func TestTestImportBoundary(t *testing.T) {
	allowed := map[string]bool{
		peosModulePrefix + "core":       true,
		peosModulePrefix + "validation": true,
	}

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
				t.Errorf("%s imports %q; test sources may import only the standard library, peos/core, and peos/validation", name, path)
			}
		}
	}
}

// TestNoPackageImportsQuality locks the converse direction for every other
// package in the repository. peos/quality is a leaf: PEOS-007 depends on
// PEOS-002 through PEOS-006, never the reverse. Together with
// TestProductionImportBoundary this is what makes an import cycle
// inexpressible rather than merely absent.
func TestNoPackageImportsQuality(t *testing.T) {
	const self = peosModulePrefix + "quality"

	for _, pkg := range []string{"core", "validation", "requirement", "lifecycle", "decision", "relation"} {
		dir := filepath.Join("..", pkg)
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("peos/%s not present at %s: %v", pkg, dir, err)
			continue
		}
		byFile := parsePackageImports(t, dir, true)
		for name, paths := range byFile {
			for _, path := range paths {
				if path == self {
					t.Errorf("peos/%s/%s imports %q; PEOS-007 is a leaf and must not be depended upon by an earlier specification's package", pkg, name, self)
				}
			}
		}
	}
}

// --- structural absence ------------------------------------------------------

// forbiddenTypeNames are PEOS-007 concepts an implementation must not turn
// into a type. Each corresponds to a non-conforming pattern or an explicit
// prohibition; the test below asserts none is declared as an exported type in
// this package.
var forbiddenTypeNames = []string{
	"QualityProfileVersion",         // "Quality Profile Version" pattern
	"QualityCharacteristicRevision", // "Quality Characteristic Revision" pattern
	"QualityMeasureVersion",         // "Quality Measure Version" pattern
	"QualityEvaluation",             // "Independent Quality Evaluation Entity" pattern
	"MeasurementPlan",               // "Parallel Quality Activity" pattern
	"MeasurementActivity",           // "Parallel Quality Activity" pattern
	"QualityEvidence",               // "Parallel Quality Evidence" pattern
	"QualityClaim",                  // Claim specializes validation.Claim; no parallel base
	"QualityScore",                  // "Mutable Quality Score" pattern
	"QualityProfileRef",             // core deliberately defines no such reference type
	"QualityProfileRevisionRef",
	"ProfileVersion",
	"CharacteristicRevision",
	"MeasureVersion",
	"QualityState",
	"QualityStatus",
	"QualityOutcome",
	"Lifecycle",
	"LifecycleState",
	"StateAssignment",
	"Relation",
	"QualityRelation",
}

// TestNoForbiddenTypeDeclared parses this package's production sources and
// asserts that none of the forbidden PEOS-007 concept types is declared.
// Parsing the AST rather than using reflection is what lets the check cover
// types that are never referenced by any value in a test.
func TestNoForbiddenTypeDeclared(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	forbidden := make(map[string]bool, len(forbiddenTypeNames))
	for _, n := range forbiddenTypeNames {
		forbidden[n] = true
	}

	fset := token.NewFileSet()
	sawProfile := false
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for declName := range file.Scope.Objects {
			if declName == "Profile" {
				sawProfile = true
			}
			if forbidden[declName] {
				t.Errorf("%s declares forbidden identifier %q", name, declName)
			}
		}
	}
	if !sawProfile {
		t.Error("no production file declares Profile; the parser may be reading the wrong directory")
	}
}

// ownedValueForbiddenMethods are methods no Quality Profile Revision-owned
// value may expose. Each would contradict PEOS-007's Profile-Owned Rule,
// Characteristic Scope, or Measure Scope Invariant by granting the value an
// identity, a revision, a lifecycle, or a mutable key.
var ownedValueForbiddenMethods = []string{
	"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
	"Lifecycle", "State", "Status",
	"WithKey", "WithoutKey",
	"Provenance", "WithProvenance",
	"Relation", "Source", "Target",
	"Score", "QualityScore", "Outcome", "Verdict",
	"SetKey", "SetDescription",
}

// TestOwnedValuesExposeNoIdentityOrLifecycle audits all seven Profile
// Revision-owned value types plus the three vocabulary wrappers, over the
// public API via reflection, for any method that would grant them an
// identity, a revision, a lifecycle, their own provenance, a relation, or a
// stored outcome.
//
// Note that Target legitimately has a Value method and Threshold has both
// Value and Operator; those are the boundary and desired-value data
// themselves, not derived state. "Outcome", "Verdict", and "Score" are the
// forbidden ones, and none of the ten types declares any of them.
func TestOwnedValuesExposeNoIdentityOrLifecycle(t *testing.T) {
	types := map[string]reflect.Type{
		"Characteristic":    reflect.TypeOf(Characteristic{}),
		"Measure":           reflect.TypeOf(Measure{}),
		"Threshold":         reflect.TypeOf(Threshold{}),
		"Target":            reflect.TypeOf(Target{}),
		"Constraint":        reflect.TypeOf(Constraint{}),
		"NormalizationRule": reflect.TypeOf(NormalizationRule{}),
		"AggregationRule":   reflect.TypeOf(AggregationRule{}),
		"Unit":              reflect.TypeOf(Unit{}),
		"Scale":             reflect.TypeOf(Scale{}),
		"ThresholdOperator": reflect.TypeOf(ThresholdOperator{}),
	}

	for typeName, typ := range types {
		// Check both the value and pointer method sets: a forbidden method
		// declared with a pointer receiver would be just as much of a
		// violation.
		for _, candidate := range []reflect.Type{typ, reflect.PointerTo(typ)} {
			for _, forbidden := range ownedValueForbiddenMethods {
				if _, ok := candidate.MethodByName(forbidden); ok {
					t.Errorf("%s exposes forbidden method %s; a Profile Revision-owned value has no identity, revision, lifecycle, provenance, relation, or stored outcome of its own", typeName, forbidden)
				}
			}
		}
	}
}

// TestProfileContentExposesNoMandatoryStateModifier asserts the
// constructor-completeness half of the contract from the other direction: no
// With* or Without* method may establish or remove ProfileContent's mandatory
// state. Replacing mandatory aggregate state means constructing a new
// ProfileContent, and per PEOS-007 a new Artifact Revision to carry it.
func TestProfileContentExposesNoMandatoryStateModifier(t *testing.T) {
	forbidden := []string{
		"WithScope", "WithoutScope",
		"WithApplicability", "WithoutApplicability",
		"WithProvenance", "WithoutProvenance",
		"WithCharacteristics", "WithoutCharacteristics",
		"WithMeasures", "WithoutMeasures",
		// Removal counterparts of the optional collections: each collection
		// modifier already expresses removal by accepting nil, and a second
		// method would create a second validation path for one field.
		"WithoutThresholds", "WithoutTargets", "WithoutConstraints",
		"WithoutNormalizationRules", "WithoutAggregationRules",
		"WithoutSubjects", "WithoutSubjectTypes",
		// Identity and derived state.
		"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
		"Lifecycle", "State", "Status",
		"Score", "QualityScore", "Aggregate", "Current", "Latest", "Effective",
		// Artifact Relation members. "Target" is deliberately absent from
		// this list: ProfileContent.Target(key) is the by-key lookup for the
		// Target collection, not a relation endpoint. The relation concern is
		// covered by Relation/RelationType/Source, none of which exists here,
		// and by the no-"relation"-key wire-form test.
		"Relation", "RelationType", "Source",
	}
	typ := reflect.TypeOf(ProfileContent{})
	for _, candidate := range []reflect.Type{typ, reflect.PointerTo(typ)} {
		for _, name := range forbidden {
			if _, ok := candidate.MethodByName(name); ok {
				t.Errorf("ProfileContent exposes forbidden method %s", name)
			}
		}
	}
}

// TestProfileAndRevisionExposeNoVersionOrLifecycle asserts that the two
// Artifact-level types introduce no Quality Profile Version mechanism, no
// lifecycle, and no stored quality score -- the last being explicitly
// forbidden on a Quality Profile by name.
func TestProfileAndRevisionExposeNoVersionOrLifecycle(t *testing.T) {
	forbidden := []string{
		"Version", "ProfileVersion", "WithVersion",
		"Lifecycle", "State", "Status", "StateAssignment",
		"Relation", "Source", "Target",
		"Score", "QualityScore", "Current", "Latest", "Effective", "Aggregate",
		"WithContent", "WithoutContent",
		"WithCore", "WithType",
	}
	types := map[string]reflect.Type{
		"Profile":         reflect.TypeOf(Profile{}),
		"ProfileRevision": reflect.TypeOf(ProfileRevision{}),
	}
	for typeName, typ := range types {
		for _, candidate := range []reflect.Type{typ, reflect.PointerTo(typ)} {
			for _, name := range forbidden {
				if _, ok := candidate.MethodByName(name); ok {
					t.Errorf("%s exposes forbidden method %s", typeName, name)
				}
			}
		}
	}
}

// TestPacketI2TypesNotYetDeclared records the packet boundary: Packet I.1
// implements the Quality Profile side only, and neither PEOS-007
// specialization of a PEOS-006 mechanism exists yet. This test is expected to
// be removed by Packet I.2 when it adds them -- its presence is what stops a
// later reader from assuming I.1 shipped them.
func TestPacketI2TypesNotYetDeclared(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for declName := range file.Scope.Objects {
			switch declName {
			case "MeasurementRecord", "Claim", "NewMeasurementRecord", "NewClaim":
				t.Errorf("%s declares %q; Packet I.1 implements the Quality Profile side only, and Measurement Record and Quality Claim belong to Packet I.2", name, declName)
			}
		}
	}
}
