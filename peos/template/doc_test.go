package template

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/aleka7sk/PEOS/peos/validation"
)

const peosModulePrefix = "github.com/aleka7sk/PEOS/peos/"

// parsePackageImports parses every .go file in dir and returns, per file, the
// set of import paths it declares. Modelled on runtime/doc_test.go's helper of
// the same name, itself modelled on quality/doc_test.go.
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

// TestProductionImportBoundary locks this package's final dependency rule:
// production sources import only the standard library, peos/core,
// peos/relation, and peos/validation.
//
// peos/relation is required because PEOS-009 defines three Artifact Relation
// types (Generated-From, Template Composition, Template Specialization), each
// carrying SHALL-identify state a bare relation.Relation cannot hold.
// peos/validation is required by exactly one file, claim.go, for the Template
// Conformance Claim helper -- PEOS-009 defines no Claim base mechanism of its
// own, so the helper delegates to validation.NewClaim.
//
// peos/lifecycle is forbidden permanently: a Template is an ordinary PEOS-003
// Lifecycle Subject, modeled entirely in peos/lifecycle, and a Template's State
// Assignment establishes neither Supersession, compatibility, nor conformance.
// peos/requirement, peos/quality, and peos/runtime are forbidden because a
// Template names only the Artifact *Types* it may generate, never an instance
// of any of them.
func TestProductionImportBoundary(t *testing.T) {
	allowed := map[string]bool{
		peosModulePrefix + "core":       true,
		peosModulePrefix + "relation":   true,
		peosModulePrefix + "validation": true,
	}
	forbidden := []string{
		peosModulePrefix + "lifecycle",
		peosModulePrefix + "decision",
		peosModulePrefix + "requirement",
		peosModulePrefix + "quality",
		peosModulePrefix + "runtime",
	}

	byFile := parsePackageImports(t, ".", false)
	seen := make(map[string]bool)
	for name, paths := range byFile {
		for _, path := range paths {
			if !strings.HasPrefix(path, peosModulePrefix) {
				// Standard library (or any non-PEOS module) is permitted.
				continue
			}
			if allowed[path] {
				seen[path] = true
				continue
			}
			t.Errorf("%s imports %q; production sources may import only the standard library, peos/core, peos/relation, and peos/validation", name, path)
		}
		for _, bad := range forbidden {
			for _, path := range paths {
				if path == bad {
					t.Errorf("%s imports forbidden package %q", name, bad)
				}
			}
		}
	}
	for path := range allowed {
		if !seen[path] {
			t.Errorf("no production file imports %q; the boundary permits it, so an unused permission should be removed rather than left standing", path)
		}
	}

	// peos/validation is permitted for one reason only. Locking it to claim.go
	// keeps the Conformance Claim helper from becoming a general licence for
	// this package to compose validation types.
	for name, paths := range byFile {
		if name == "claim.go" {
			continue
		}
		for _, path := range paths {
			if path == peosModulePrefix+"validation" {
				t.Errorf("%s imports peos/validation; only claim.go may, and only for the Template Conformance Claim helper", name)
			}
		}
	}
}

// TestTestImportBoundary confirms the test sources stay inside the same
// boundary, so a test cannot quietly establish a dependency the production code
// is forbidden to have.
func TestTestImportBoundary(t *testing.T) {
	allowed := map[string]bool{
		peosModulePrefix + "core":       true,
		peosModulePrefix + "relation":   true,
		peosModulePrefix + "validation": true,
	}

	// A package's own external test files (package template_test) legitimately
	// import the package under test -- that is the standard Go idiom for
	// example_test.go files godoc attaches to this package, not a boundary
	// violation, so self-import is allowed here.
	allowed[peosModulePrefix+"template"] = true
	for name, paths := range parsePackageImports(t, ".", true) {
		for _, path := range paths {
			if !strings.HasPrefix(path, peosModulePrefix) {
				continue
			}
			if !allowed[path] {
				t.Errorf("%s imports %q; test sources may import only the standard library, peos/core, peos/relation, and peos/validation", name, path)
			}
		}
	}
}

// TestNothingImportsTemplate confirms peos/template is a leaf: no other package
// in this module imports it, and none should.
func TestNothingImportsTemplate(t *testing.T) {
	root := filepath.Join("..", "..")
	templatePath := peosModulePrefix + "template"

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Skip this package's own files.
		if dir := filepath.Dir(path); filepath.Base(dir) == "template" {
			return nil
		}
		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", path, parseErr)
		}
		for _, spec := range file.Imports {
			imported, unquoteErr := strconv.Unquote(spec.Path.Value)
			if unquoteErr != nil {
				t.Fatalf("%s: unquote import %s: %v", path, spec.Path.Value, unquoteErr)
			}
			if imported == templatePath {
				t.Errorf("%s imports %q; peos/template must remain a leaf", path, templatePath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}

// declaredNames returns every top-level name declared by this package's
// production sources.
func declaredNames(t *testing.T) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	declared := make(map[string]bool)
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
			declared[declName] = true
		}
	}
	return declared
}

// TestPacketK1TypesDeclared is the positive counterpart to
// TestPacketK2TypesNotYetDeclared: every construct Packet K.1 is responsible
// for must exist.
func TestPacketK1TypesDeclared(t *testing.T) {
	declared := declaredNames(t)
	for _, name := range []string{
		"ArtifactTypeTemplate",
		"Template", "NewTemplate",
		"TemplateRevision", "NewTemplateRevision",
		"TemplateContent", "NewTemplateContent",
		"TemplateApplicability", "NewUnrestrictedTemplateApplicability", "NewScopedTemplateApplicability",
		"Parameter", "NewParameter",
		"ParameterType", "NewVocabularyParameterType", "NewExternalParameterType",
		"ParameterDefault", "NewParameterDefault",
		"ParameterConstraint", "NewParameterConstraint",
		"ConstraintTarget", "NewParameterConstraintTarget", "NewGeneratedContentConstraintTarget",
		"ConstraintEvaluationPoint", "NewConstraintEvaluationPoint",
		"ConstraintFailureSemantics", "NewConstraintFailureSemantics",
		"CompatibilityDeclaration", "NewCompatibilityDeclaration",
		"ErrInvalidTemplate", "ErrTemplateArtifactTypeMismatch", "ErrTemplateArtifactIDMismatch",
		"ErrInvalidTemplateContent", "ErrInvalidTemplateApplicability",
		"ErrInvalidTemplateParameter", "ErrInvalidParameterType",
		"ErrInvalidParameterDefault", "ErrInvalidParameterConstraint",
		"ErrInvalidConstraintTarget", "ErrInvalidCompatibilityDeclaration",
		"ErrDuplicateTemplateLocalKey", "ErrUnknownTemplateLocalKey",
	} {
		if !declared[name] {
			t.Errorf("%q is not declared; Packet K.1 must implement it", name)
		}
	}
}

// TestPacketK2TypesDeclared is the positive counterpart that replaced Packet
// K.1's TestPacketK2TypesNotYetDeclared, following the precedent Packet I.2 set
// when it replaced I.1's equivalent absence test: every construct Packet K.2 is
// responsible for must now exist.
func TestPacketK2TypesDeclared(t *testing.T) {
	declared := declaredNames(t)
	for _, name := range []string{
		// The Template Application Record family.
		"ApplicationRecord", "NewApplicationRecord",
		"ApplicationOutcome", "NewApplicationOutcome",
		"ApplicationOutcomeSucceeded", "ApplicationOutcomeFailed",
		"ApplicationOutcomePartiallySucceeded", "ApplicationOutcomeInterrupted",
		"ApplicationOutcomeIndeterminate",
		"ResolvedValue", "NewResolvedValue",
		"ValueSource", "NewValueSource",
		"ValueSourceExplicitInput", "ValueSourceDefault", "ValueSourceDerived",
		"GeneratedOutput", "NewGeneratedOutput",
		// The three relation wrappers.
		"GeneratedFrom", "NewGeneratedFrom",
		"Composition", "NewComposition",
		"Specialization", "NewSpecialization",
		// The Conformance Claim helper.
		"NewTemplateConformanceClaim",
		// K.2 sentinels.
		"ErrInvalidApplicationRecord", "ErrInvalidResolvedValue",
		"ErrInvalidGeneratedOutput", "ErrInvalidTemplateRelation",
		"ErrInvalidGeneratedFrom", "ErrInvalidComposition", "ErrInvalidSpecialization",
	} {
		if !declared[name] {
			t.Errorf("%q is not declared; Packet K.2 must implement it", name)
		}
	}
}

// TestNoTemplateConformanceClaimType asserts the Conformance Claim helper did
// not bring a wrapper type with it. PEOS-009 names "Defining a Template
// Conformance Claim that does not specialize PEOS-006's Conformance Claim, or
// that redefines Claim identity, immutability, or replacement semantics" and
// "Representing a Template Conformance Claim as an Artifact" as non-conforming
// patterns; returning a bare validation.Claim makes both unrepresentable.
func TestNoTemplateConformanceClaimType(t *testing.T) {
	declared := declaredNames(t)
	for _, name := range []string{
		"TemplateConformanceClaim", "ConformanceClaim", "Claim",
		"NewClaim", "TemplateClaim",
	} {
		if declared[name] {
			t.Errorf("%q is declared; PEOS-009 defines no Claim base mechanism and this package must define no Claim type", name)
		}
	}
}

// TestNewTemplateConformanceClaimReturnsValidationClaim locks the helper's
// return type structurally: it must hand back peos/validation's own Claim, not
// a package-local type that merely resembles it.
func TestNewTemplateConformanceClaimReturnsValidationClaim(t *testing.T) {
	fnType := reflect.TypeOf(NewTemplateConformanceClaim)
	if fnType.NumOut() != 2 {
		t.Fatalf("NewTemplateConformanceClaim returns %d values, want 2", fnType.NumOut())
	}
	got := fnType.Out(0)
	want := reflect.TypeOf(validation.Claim{})
	if got != want {
		t.Errorf("NewTemplateConformanceClaim returns %v, want %v", got, want)
	}
	if got.PkgPath() != peosModulePrefix+"validation" {
		t.Errorf("return type belongs to %q, want %q", got.PkgPath(), peosModulePrefix+"validation")
	}
}

// TestForbiddenOntologyAbsent asserts that the constructs PEOS-009 forbids
// outright, or that would import a templating engine this specification
// disclaims, exist nowhere in this package -- in any packet.
func TestForbiddenOntologyAbsent(t *testing.T) {
	declared := declaredNames(t)
	for _, name := range []string{
		// Named non-conforming patterns.
		"TemplateVersion", "TemplateInstance", "ApplicationRecordRevision",
		"TemplateCollection", "TemplateCompositionSet",
		// Execution / generation machinery PEOS-009 does not define.
		"TemplateExecution", "TemplateInvocation", "TemplateResult",
		"GeneratedArtifactStore", "RenderedTemplate", "Renderer", "Engine",
		"Script", "Expression", "Function", "Loop", "Conditional",
		// Derived state that must never be stored.
		"CurrentTemplate", "ActiveTemplate", "EffectiveTemplate",
		"TemplateCompatibility", "CurrentCompatibility", "EffectiveCompatibility",
		"TemplateConformance", "DerivedTemplateConformance",
		// Deliberately absent permanently: PEOS-009 defines no separate
		// Template Supersession entity, and Migration has no stated ontology.
		"TemplateSupersession", "Supersession",
		"Migration", "TemplateMigration", "MigrationRecord", "MigrationRef",
		// A parameter binding is not an entity: resolved values are plain
		// owned values on the Application Record.
		"ParameterBinding", "Binding", "BindingRecord", "NewBindingRecord",
		// Runtime/deployment concepts belonging to other specifications.
		"Deployment", "Provider", "Endpoint", "RuntimeState", "ComplianceState",
		"Secret", "Credential",
	} {
		if declared[name] {
			t.Errorf("%q is declared; PEOS-009 does not define it and this package must not either", name)
		}
	}
}

// assertNoMethods fails for every forbidden method name present on typ.
func assertNoMethods(t *testing.T, label string, typ reflect.Type, forbidden []string) {
	t.Helper()
	for _, name := range forbidden {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("%s exposes forbidden method %s()", label, name)
		}
	}
	ptr := reflect.PointerTo(typ)
	for _, name := range forbidden {
		if _, ok := ptr.MethodByName(name); ok {
			t.Errorf("*%s exposes forbidden method %s()", label, name)
		}
	}
}

// exactModifierSet returns every exported With*-prefixed method on typ, sorted.
func exactModifierSet(typ reflect.Type) []string {
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	return modifiers
}

func TestTemplateExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"Version", "TemplateVersion", "WithVersion",
		"Status", "State", "Lifecycle", "StateAssignment",
		"Current", "Active", "Effective",
		"Compatible", "Compatibility", "Conformant", "Conformance",
		"Render", "Expand", "Generate", "Apply", "Instantiate",
		"WithContent", "Content",
	}
	assertNoMethods(t, "Template", reflect.TypeOf(Template{}), forbidden)

	// Template is a pure identity wrapper: it has no modifier at all.
	if got := exactModifierSet(reflect.TypeOf(Template{})); len(got) != 0 {
		t.Errorf("Template modifiers = %v, want none", got)
	}
}

func TestTemplateRevisionExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"WithContent", "Version", "TemplateVersion", "WithVersion",
		"Status", "State", "Lifecycle",
		"Current", "Active", "Effective",
		"Compatible", "Compatibility", "Conformant", "Conformance",
		"Render", "Expand", "Generate", "Apply", "Instantiate",
		"GeneratedArtifact", "GeneratedArtifactRevision", "Outputs",
	}
	assertNoMethods(t, "TemplateRevision", reflect.TypeOf(TemplateRevision{}), forbidden)

	if got := exactModifierSet(reflect.TypeOf(TemplateRevision{})); len(got) != 0 {
		t.Errorf("TemplateRevision modifiers = %v, want none", got)
	}
}

// TestTemplateContentExposesNoForbiddenMethod locks the crucial body boundary:
// TemplateContent must expose no accessor for a template body, source, script,
// or expression, because PEOS-009 assigns all of that to
// core.ArtifactRevision.Representations().
func TestTemplateContentExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		// The body belongs to core.Representation.
		"Body", "TemplateBody", "Source", "Schema", "Script", "Expression",
		"RenderableContent", "Representation", "Representations",
		"WithBody", "WithSource", "WithSchema",
		// Derived state that must never be stored.
		"Compatible", "Conformant", "Current", "Active", "Effective",
		"Status", "State", "Lifecycle",
		// Generated output belongs to the Application Record and to the
		// generated Artifact itself.
		"GeneratedArtifact", "GeneratedArtifactRevision", "Outputs",
		"ResolvedValues", "Outcome",
		// Mandatory state has no modifier.
		"WithGeneratedArtifactTypes", "WithExpansionSemantics",
		"WithCompatibility", "WithApplicability", "WithProvenance",
		"WithParameters", "WithDefaults", "WithConstraints",
		// No engine.
		"Render", "Expand", "Generate", "Apply", "Instantiate", "Evaluate",
		// Migration has no stated ontology.
		"Migration", "Migrations", "WithMigration",
	}
	assertNoMethods(t, "TemplateContent", reflect.TypeOf(TemplateContent{}), forbidden)

	want := []string{
		"WithAuthority",
		"WithCompositionReferences",
		"WithExtension",
		"WithSpecializationReferences",
		"WithoutAuthority",
		"WithoutExtension",
	}
	if got := exactModifierSet(reflect.TypeOf(TemplateContent{})); !slices.Equal(got, want) {
		t.Errorf("TemplateContent modifiers = %v, want %v", got, want)
	}
}

func TestParameterExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		// No identity of its own.
		"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
		"Lifecycle", "State", "Status",
		// No value of any kind -- resolved values belong to the Application
		// Record.
		"Value", "CurrentValue", "ResolvedValue", "Binding", "Bind",
		"WithValue", "WithResolvedValue",
		// PEOS-009 states none of these for a parameter.
		"Provenance", "Authority", "Scope",
		"WithProvenance", "WithAuthority", "WithScope",
		// The key is immutable.
		"WithKey", "WithoutKey", "WithType", "WithRequired",
	}
	assertNoMethods(t, "Parameter", reflect.TypeOf(Parameter{}), forbidden)

	want := []string{
		"WithDescription",
		"WithForbiddenDefaultResolution",
		"WithPermittedDefaultResolution",
		"WithoutDescription",
	}
	if got := exactModifierSet(reflect.TypeOf(Parameter{})); !slices.Equal(got, want) {
		t.Errorf("Parameter modifiers = %v, want %v", got, want)
	}
}

func TestParameterTypeExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "Revision", "Lifecycle", "State", "Status",
		"ArtifactRevision", "GoType", "Primitive", "JSONSchema", "Schema",
		"Validate", "Validator", "Evaluate",
		"WithVocabulary", "WithExternal",
	}
	assertNoMethods(t, "ParameterType", reflect.TypeOf(ParameterType{}), forbidden)

	if got := exactModifierSet(reflect.TypeOf(ParameterType{})); len(got) != 0 {
		t.Errorf("ParameterType modifiers = %v, want none", got)
	}
}

func TestParameterDefaultExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "Key", "Revision", "Lifecycle", "State", "Status",
		"Resolved", "Applied", "Effective", "Resolve",
		"WithParameter", "WithValue",
	}
	assertNoMethods(t, "ParameterDefault", reflect.TypeOf(ParameterDefault{}), forbidden)

	if got := exactModifierSet(reflect.TypeOf(ParameterDefault{})); len(got) != 0 {
		t.Errorf("ParameterDefault modifiers = %v, want none", got)
	}
}

func TestParameterConstraintExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "Revision", "Lifecycle", "State", "Status",
		// A constraint declares a rule and never records whether it held.
		"Result", "Outcome", "Satisfied", "Violated", "Evaluated",
		"Evaluate", "Check", "Apply",
		// Immutable state has no modifier.
		"WithKey", "WithTarget", "WithRule", "WithScope",
		"WithEvaluationPoint", "WithFailureSemantics",
	}
	assertNoMethods(t, "ParameterConstraint", reflect.TypeOf(ParameterConstraint{}), forbidden)

	want := []string{"WithAuthority", "WithoutAuthority"}
	if got := exactModifierSet(reflect.TypeOf(ParameterConstraint{})); !slices.Equal(got, want) {
		t.Errorf("ParameterConstraint modifiers = %v, want %v", got, want)
	}
}

func TestConstraintTargetExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "Revision", "Lifecycle", "State", "Status",
		// The generated-content arm must never carry generated identity.
		"GeneratedArtifactID", "GeneratedArtifactRevisionID",
		"GeneratedArtifactRef", "GeneratedArtifactRevisionRef",
		"Artifact", "ArtifactRevision", "Output", "Result",
		"WithParameter", "WithGeneratedContent",
	}
	assertNoMethods(t, "ConstraintTarget", reflect.TypeOf(ConstraintTarget{}), forbidden)

	if got := exactModifierSet(reflect.TypeOf(ConstraintTarget{})); len(got) != 0 {
		t.Errorf("ConstraintTarget modifiers = %v, want none", got)
	}
}

// TestCompatibilityDeclarationExposesNoForbiddenMethod locks the
// declaration-not-verdict boundary: PEOS-009 names Template.compatible and
// TemplateRevision.compatible as non-conforming patterns, and current
// compatibility is always derived at query time.
func TestCompatibilityDeclarationExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"Compatible", "IsCompatible", "Incompatible", "Compatibility",
		"Current", "CurrentCompatibility", "Effective", "EffectiveCompatibility",
		"Status", "State", "Verdict", "Result", "Outcome", "Evaluate", "Resolve",
		"ID", "Ref", "Revision", "Lifecycle",
		// Migration has no stated ontology; only an opaque descriptor exists.
		"Migration", "Migrations",
	}
	assertNoMethods(t, "CompatibilityDeclaration", reflect.TypeOf(CompatibilityDeclaration{}), forbidden)

	want := []string{
		"WithApplicableRevisions",
		"WithMigrationRequirements",
		"WithoutMigrationRequirements",
	}
	if got := exactModifierSet(reflect.TypeOf(CompatibilityDeclaration{})); !slices.Equal(got, want) {
		t.Errorf("CompatibilityDeclaration modifiers = %v, want %v", got, want)
	}
}

// TestApplicationRecordExposesNoForbiddenMethod locks the record's ontology: it
// is an immutable non-Artifact record, "not revisioned", "not lifecycle-bearing",
// and never a holder of generated content.
func TestApplicationRecordExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		// Not an Artifact, not revisioned.
		"ArtifactID", "ArtifactType", "Type", "Revision", "RevisionID",
		"Version", "WithVersion", "Core",
		// Not lifecycle-bearing.
		"Lifecycle", "State", "Status", "StateAssignment",
		"Current", "Active", "Effective",
		// Immutable: mandatory state has no modifier, and an outcome is a
		// recorded fact that never changes.
		"WithID", "WithTemplate", "WithActor", "WithAppliedAt",
		"WithEnvironment", "WithProvenance", "WithOutcome", "WithResolvedValues",
		// Never holds generated content or runs anything.
		"Content", "Payload", "Rendered", "Output", "Render", "Apply",
		"Generate", "Execute", "Instantiate",
		// Conformance and compatibility are derived elsewhere.
		"Conformant", "Conformance", "Compatible", "Compatibility",
	}
	assertNoMethods(t, "ApplicationRecord", reflect.TypeOf(ApplicationRecord{}), forbidden)

	want := []string{
		"WithAuthority",
		"WithCorrection",
		"WithExtension",
		"WithGeneratedOutputs",
		"WithLimitations",
		"WithUngeneratedOutputs",
		"WithoutAuthority",
		"WithoutCorrection",
		"WithoutExtension",
	}
	if got := exactModifierSet(reflect.TypeOf(ApplicationRecord{})); !slices.Equal(got, want) {
		t.Errorf("ApplicationRecord modifiers = %v, want %v", got, want)
	}
}

func TestResolvedValueAndGeneratedOutputExposeNoForbiddenMethod(t *testing.T) {
	assertNoMethods(t, "ResolvedValue", reflect.TypeOf(ResolvedValue{}), []string{
		"ID", "Ref", "Revision", "Lifecycle", "State", "Status",
		"Bind", "Binding", "Resolve", "WithValue", "WithSource", "WithParameter",
	})
	if got := exactModifierSet(reflect.TypeOf(ResolvedValue{})); len(got) != 0 {
		t.Errorf("ResolvedValue modifiers = %v, want none", got)
	}

	assertNoMethods(t, "GeneratedOutput", reflect.TypeOf(GeneratedOutput{}), []string{
		"ID", "Ref", "Lifecycle", "State", "Status",
		// It names what was generated and never holds it.
		"Content", "Payload", "Rendered", "Result", "Bytes",
		"WithArtifact", "WithRevision",
	})
	if got := exactModifierSet(reflect.TypeOf(GeneratedOutput{})); len(got) != 0 {
		t.Errorf("GeneratedOutput modifiers = %v, want none", got)
	}
}

// TestRelationWrappersExposeNoForbiddenMethod locks all three relation types:
// none has identity, revision, or lifecycle, none can lose its mandatory scope,
// and none evaluates or executes anything.
func TestRelationWrappersExposeNoForbiddenMethod(t *testing.T) {
	shared := []string{
		// PEOS-002/PEOS-009: no identity, revision, or lifecycle.
		"ID", "Ref", "RelationID", "Revision", "RevisionID",
		"Lifecycle", "State", "Status", "Approval", "Valid", "ValidityPeriod",
		// Scope is mandatory: it can never be set to something else or cleared.
		"WithScope", "WithoutScope",
		// No execution, no graph traversal, no cycle detection here.
		"Render", "Expand", "Generate", "Apply", "Evaluate", "Resolve",
		"Cycles", "Graph", "Traverse", "Transitive",
		// Multiplicity/direction/cycle policy are properties of the relation
		// type, not per-instance state.
		"Multiplicity", "Direction", "CyclePolicy",
	}

	// ParticipantLevels is forbidden on exactly the two relation types whose
	// participant levels PEOS-009 fixes -- Generated-From (generated Artifact
	// Revision → Template Artifact Revision) and Composition (Revision on both
	// sides). For those, a per-instance level accessor would report a constant.
	// Specialization is the opposite case: PEOS-009 permits "the source Template
	// **or** Template Artifact Revision" and separately requires the relation to
	// identify "participant levels", so there the accessor is required, not
	// forbidden -- see TestSpecializationSupportsBothParticipantLevels.
	levelFixed := append(slices.Clone(shared), "ParticipantLevels")

	assertNoMethods(t, "GeneratedFrom", reflect.TypeOf(GeneratedFrom{}), append(slices.Clone(levelFixed), []string{
		// PEOS-009 states directly what a Generated-From SHALL NOT contain.
		"ResolvedValues", "ResolvedValue", "Outcome", "Events", "EventHistory",
		"AuthorityHistory", "Authority", "ApplicationRecord",
	}...))
	assertNoMethods(t, "Composition", reflect.TypeOf(Composition{}), levelFixed)
	assertNoMethods(t, "Specialization", reflect.TypeOf(Specialization{}), append(slices.Clone(shared), []string{
		// The compatibility effect is declared, never computed.
		"Compatible", "IsCompatible", "Compatibility",
	}...))

	want := []string{"WithExtension", "WithoutExtension"}
	for label, typ := range map[string]reflect.Type{
		"GeneratedFrom":  reflect.TypeOf(GeneratedFrom{}),
		"Composition":    reflect.TypeOf(Composition{}),
		"Specialization": reflect.TypeOf(Specialization{}),
	} {
		if got := exactModifierSet(typ); !slices.Equal(got, want) {
			t.Errorf("%s modifiers = %v, want %v", label, got, want)
		}
	}
}

// TestTemplateApplicabilityExposesNoForbiddenMethod confirms applicability is
// not conflated with lifecycle state or a compatibility result.
func TestTemplateApplicabilityExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "Revision", "Lifecycle", "State", "Status",
		"Compatible", "Conformant", "Current", "Active", "Effective",
		"WithScope", "WithoutScope", "WithKind",
	}
	assertNoMethods(t, "TemplateApplicability", reflect.TypeOf(TemplateApplicability{}), forbidden)

	if got := exactModifierSet(reflect.TypeOf(TemplateApplicability{})); len(got) != 0 {
		t.Errorf("TemplateApplicability modifiers = %v, want none", got)
	}
}
