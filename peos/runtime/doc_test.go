package runtime

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
)

const peosModulePrefix = "github.com/aleka7sk/PEOS/peos/"

// parsePackageImports parses every .go file in dir and returns, per file,
// the set of import paths it declares. Modelled on
// quality/doc_test.go's helper of the same name, itself modelled on
// validation/doc_test.go.
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
// production sources may import only the standard library, peos/core, and
// peos/validation.
//
// peos/validation is permitted, and as of Packet J.2 is actually required:
// claim.go's NewComplianceClaim delegates to validation.NewClaim, because
// PEOS-008 specializes PEOS-006's Claim mechanism rather than redefining
// it. The test asserts that at least one production file imports it, so
// the delegation cannot be quietly replaced by a re-declared local type.
// peos/relation, peos/lifecycle, peos/requirement, peos/decision, and
// peos/quality are excluded for the reasons doc.go states: Requirement
// references are Revision-owned content rather than Artifact Relations; a
// Runtime Contract's optional activation is governed exclusively by
// peos/lifecycle, never duplicated here; a Runtime Contract never becomes
// a Requirement subtype; peos/requirement.Waiver is referenced only by
// opaque description, never by importing peos/requirement; and a runtime
// outcome is never governance authority.
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
		peosModulePrefix + "quality",
	}

	byFile := parsePackageImports(t, ".", false)
	sawCore := false
	sawValidation := false
	for name, paths := range byFile {
		for _, path := range paths {
			if !strings.HasPrefix(path, peosModulePrefix) {
				// Standard library (or any non-PEOS module) is permitted.
				continue
			}
			if allowed[path] {
				switch path {
				case peosModulePrefix + "core":
					sawCore = true
				case peosModulePrefix + "validation":
					sawValidation = true
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
	if !sawValidation {
		t.Errorf("no production file imports %q; NewComplianceClaim must delegate to validation.NewClaim, not re-declare it", peosModulePrefix+"validation")
	}
}

// TestTestImportBoundary applies the same rule to this package's own test
// sources.
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

// TestNoPackageImportsRuntime locks the converse direction for every other
// package in the repository. peos/runtime is a leaf: nothing in
// PEOS-000-007 depends on PEOS-008, and PEOS-009 (Packet J.1's successor
// area) has no package yet either. Together with
// TestProductionImportBoundary this is what makes an import cycle
// inexpressible rather than merely absent.
func TestNoPackageImportsRuntime(t *testing.T) {
	const self = peosModulePrefix + "runtime"

	for _, pkg := range []string{"core", "relation", "lifecycle", "decision", "requirement", "validation", "quality"} {
		dir := filepath.Join("..", pkg)
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("peos/%s not present at %s: %v", pkg, dir, err)
			continue
		}
		byFile := parsePackageImports(t, dir, true)
		for name, paths := range byFile {
			for _, path := range paths {
				if path == self {
					t.Errorf("peos/%s/%s imports %q; PEOS-008 is a leaf and must not be depended upon by an earlier specification's package", pkg, name, self)
				}
			}
		}
	}
}

// --- structural absence ------------------------------------------------------

// forbiddenTypeNames are concepts PEOS-008 does not define, or that this
// package's ontology forbids, as an exported type. Each corresponds to a
// framing rejected by the Packet J.0 Blueprint (Section 12-19: no runtime
// interface, no execution record) or to a PEOS-008 non-conforming pattern.
var forbiddenTypeNames = []string{
	// Rejected generic runtime-interface ontology (Blueprint sections
	// 12-19): PEOS-008 defines no input/output/dependency/capability
	// model, no invocation, no execution result, no deployment platform
	// schema.
	"Input", "Output", "Dependency", "Capability",
	"Invocation", "Result", "ExecutionRecord", "Deployment",
	"Endpoint", "Provider", "Secret", "Credential",
	// Non-conforming patterns.
	"RuntimeContractVersion",
	"CurrentBinding",
	"Compliance",
	"ComplianceClaim",
	"Waiver",
	"Incident",
	"Lifecycle",
	"LifecycleState",
	"StateAssignment",
	"Relation",
	"RuntimeRelation",
}

// TestNoForbiddenTypeDeclared parses this package's production sources and
// asserts that none of the forbidden concepts is declared. Parsing the AST
// rather than using reflection is what lets the check cover types that are
// never referenced by any value in a test.
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
	sawContract := false
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
			if declName == "Contract" {
				sawContract = true
			}
			if forbidden[declName] {
				t.Errorf("%s declares forbidden identifier %q", name, declName)
			}
		}
	}
	if !sawContract {
		t.Error("no production file declares Contract; the parser may be reading the wrong directory")
	}
}

// assertNoMethods checks both the value and pointer method sets of typ for
// any of the forbidden names. A forbidden method declared with a pointer
// receiver would be just as much of a violation as one on the value.
func assertNoMethods(t *testing.T, typeName string, typ reflect.Type, forbidden []string) {
	t.Helper()
	for _, candidate := range []reflect.Type{typ, reflect.PointerTo(typ)} {
		for _, name := range forbidden {
			if _, ok := candidate.MethodByName(name); ok {
				t.Errorf("%s exposes forbidden method %s", typeName, name)
			}
		}
	}
}

// runtimeMutableStateForbiddenMethods are the seven forbidden-field names
// PEOS-008 names by exact synonym (":337-341", ":507-512"), plus their
// obvious relatives, asserted absent on every principal type in this
// package.
var runtimeMutableStateForbiddenMethods = []string{
	"Bind", "Unbind", "Deploy", "Activate",
	"SetState", "SetStatus", "SetCompliance",
	"MarkCompliant", "MarkViolated",
	"CurrentBinding", "LatestBinding", "EffectiveBinding",
	"Bound", "ActiveDeployment", "Deployed", "Compliant",
	"Lifecycle", "State", "Status", "StateAssignment",
	"Relation", "RelationType", "Source", "Target",
}

// TestContractAndRevisionExposeNoRuntimeState asserts that Contract and
// ContractRevision introduce no Runtime Contract Version mechanism, no
// lifecycle, no relation, and no mutable binding/compliance state.
func TestContractAndRevisionExposeNoRuntimeState(t *testing.T) {
	forbidden := append([]string{
		"Version", "RuntimeContractVersion", "WithVersion",
		"WithContent", "WithoutContent",
		"WithCore", "WithType",
	}, runtimeMutableStateForbiddenMethods...)

	types := map[string]reflect.Type{
		"Contract":         reflect.TypeOf(Contract{}),
		"ContractRevision": reflect.TypeOf(ContractRevision{}),
	}
	for typeName, typ := range types {
		assertNoMethods(t, typeName, typ, forbidden)
	}
}

// TestContractContentExposesNoMandatoryStateModifier asserts the
// constructor-completeness half of the contract from the other direction:
// no With* or Without* method may establish or remove ContractContent's
// mandatory state, and no method may expose runtime binding, deployment,
// or compliance state.
func TestContractContentExposesNoMandatoryStateModifier(t *testing.T) {
	forbidden := append([]string{
		"WithRequirements", "WithoutRequirements",
		"WithSubjectTarget", "WithoutSubjectTarget",
		"WithEnvironment", "WithoutEnvironment",
		"WithDeploymentScope", "WithoutDeploymentScope",
		"WithApplicability", "WithoutApplicability",
		"WithProvenance", "WithoutProvenance",
		"WithAuthority", "WithoutAuthority",
		"WithAssertions", "WithoutAssertions",
		// Removal counterparts of the optional collections: each collection
		// modifier already expresses removal by accepting nil.
		"WithoutObservationRequirements", "WithoutViolationClassificationRules",
		"WithoutWaiverHandlingRules", "WithoutEnforcementExpectations",
		"WithoutQualityProfileRevisions",
		// Identity and derived state.
		"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
	}, runtimeMutableStateForbiddenMethods...)

	typ := reflect.TypeOf(ContractContent{})
	assertNoMethods(t, "ContractContent", typ, forbidden)
}

// TestContractContentExactModifierSet locks the exact set of With*
// modifiers ContractContent exposes, so that a future change adding an
// unexpected modifier (for example, a compensating WithRequirements after
// some other change) is caught even though it is not individually
// blocklisted above.
func TestContractContentExactModifierSet(t *testing.T) {
	typ := reflect.TypeOf(ContractContent{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	want := []string{
		"WithEnforcementExpectations", "WithExtension",
		"WithObservationRequirements", "WithQualityProfileRevisions",
		"WithViolationClassificationRules", "WithWaiverHandlingRules",
		"WithoutExtension",
	}
	if !slices.Equal(modifiers, want) {
		t.Errorf("ContractContent modifiers = %v, want exactly %v", modifiers, want)
	}
}

// TestAssertionExposesNoForbiddenMethod asserts that Assertion carries no
// identity, no revision, no lifecycle, and no evaluation-outcome field:
// "A Runtime Assertion is a Runtime Contract Revision-owned rule. It has
// no required independent identity, no required independent revision, and
// no required independent lifecycle." An Assertion defines a rule; it
// never records whether the rule held.
func TestAssertionExposesNoForbiddenMethod(t *testing.T) {
	forbidden := append([]string{
		"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
		"Outcome", "Result", "Satisfied", "Verdict", "Violated",
		"WithKey", "WithoutKey",
		"WithSubject", "WithCriterion", "WithEvaluationRule", "WithExpectedResult", "WithScope",
	}, runtimeMutableStateForbiddenMethods...)
	assertNoMethods(t, "Assertion", reflect.TypeOf(Assertion{}), forbidden)

	typ := reflect.TypeOf(Assertion{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	want := []string{
		"WithExtension", "WithObservationInputs", "WithTemporalConditions",
		"WithUncertaintyHandling", "WithoutExtension", "WithoutTemporalConditions",
		"WithoutUncertaintyHandling",
	}
	if !slices.Equal(modifiers, want) {
		t.Errorf("Assertion modifiers = %v, want exactly %v", modifiers, want)
	}
}

// TestRequirementReferenceHasNoThirdArm asserts RequirementReference
// exposes exactly the two accessor arms PEOS-008 permits, with no
// "latest"/"current" implicit-resolution arm and no Statement accessor
// (Requirement Statement content must never be copied into this type).
func TestRequirementReferenceHasNoThirdArm(t *testing.T) {
	forbidden := []string{
		"Latest", "Current", "Effective",
		"Statement", "Text", "WithStatement",
		"WithIdentity", "WithRevision", "WithKind",
	}
	assertNoMethods(t, "RequirementReference", reflect.TypeOf(RequirementReference{}), forbidden)
}

// TestOwnedValueVocabularyWrappersExposeNoIdentity audits the three
// runtime-local vocabulary wrappers over their public API for any method
// that would grant them an identity, a revision, or a lifecycle.
func TestOwnedValueVocabularyWrappersExposeNoIdentity(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
		"Lifecycle", "State", "Status",
	}
	types := map[string]reflect.Type{
		"Environment":             reflect.TypeOf(Environment{}),
		"ViolationClassification": reflect.TypeOf(ViolationClassification{}),
		"ViolationSeverity":       reflect.TypeOf(ViolationSeverity{}),
	}
	for typeName, typ := range types {
		assertNoMethods(t, typeName, typ, forbidden)
	}
}

// --- Packet J.2 structural audits -------------------------------------------

// TestPacketJ2TypesNowDeclared is the positive counterpart to
// TestNoForbiddenTypeDeclared: BindingRecord, UnbindingRecord,
// Observation, ViolationTrigger, and Violation, plus their constructors,
// must now exist. "ComplianceClaim" remains in forbiddenTypeNames above --
// NewComplianceClaim returns an ordinary validation.Claim.
func TestPacketJ2TypesNowDeclared(t *testing.T) {
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
	for _, name := range []string{
		"BindingRecord", "NewBindingRecord",
		"UnbindingRecord", "NewUnbindingRecord",
		"Observation", "NewObservation",
		"ViolationTrigger", "NewObservationTrigger", "NewEvidenceTrigger",
		"Violation", "NewViolationFromObservation", "NewViolationFromEvidence",
		"NewComplianceClaim",
	} {
		if !declared[name] {
			t.Errorf("%q is not declared; Packet J.2 must implement it", name)
		}
	}
}

// TestPacketJ2AContractRuleDeclared is the positive counterpart
// documenting audit finding J3-03's resolution: ContractRule and its
// sentinel must now exist, and ErrInvalidRuntimeContractRule (not a new
// per-field sentinel) must be the aggregate error for it.
func TestPacketJ2AContractRuleDeclared(t *testing.T) {
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
	for _, name := range []string{"ContractRule", "NewContractRule", "ErrInvalidRuntimeContractRule"} {
		if !declared[name] {
			t.Errorf("%q is not declared; Packet J.2.A must implement it", name)
		}
	}
}

// TestContractRuleExposesNoForbiddenMethod audits ContractRule: no
// identity, no revision, no lifecycle, no rule-category field (the owning
// collection supplies that), and no modifier of any kind -- ContractRule
// is an immutable value with only a constructor, matching Assertion's own
// no-key-modifier discipline.
func TestContractRuleExposesNoForbiddenMethod(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "ArtifactID", "RevisionID", "Revision",
		"Lifecycle", "State", "Status",
		"Category", "Kind", "WithKey", "WithoutKey",
		"WithText", "WithoutText",
	}
	assertNoMethods(t, "ContractRule", reflect.TypeOf(ContractRule{}), forbidden)

	typ := reflect.TypeOf(ContractRule{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	if len(modifiers) != 0 {
		t.Errorf("ContractRule modifiers = %v, want none", modifiers)
	}
}

// recordForbiddenMethods are methods no PEOS-008 immutable enforcement
// record may expose: Artifact/Revision identity (none of the four records
// is an Artifact), lifecycle, relation, and every mutable/derived-state
// synonym this package forbids package-wide.
var recordForbiddenMethods = append([]string{
	"ArtifactID", "RevisionID", "Revision", "Core",
}, runtimeMutableStateForbiddenMethods...)

// TestBindingRecordExposesNoForbiddenMethod audits BindingRecord: no
// Artifact identity, no lifecycle, no mutable "current"/"bound" state, and
// no mandatory-field modifier -- only the optional-state With* methods
// PEOS-008 permits.
func TestBindingRecordExposesNoForbiddenMethod(t *testing.T) {
	forbidden := append([]string{
		"WithContractRevision", "WithSubject", "WithEnvironment",
		"WithScope", "WithBoundAt", "WithActor", "WithProvenance",
		"WithID", "WithoutID",
	}, recordForbiddenMethods...)
	assertNoMethods(t, "BindingRecord", reflect.TypeOf(BindingRecord{}), forbidden)

	typ := reflect.TypeOf(BindingRecord{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	want := []string{
		"WithAuthority", "WithConfigurationReference", "WithCorrection",
		"WithDeploymentTimestamp", "WithExtension", "WithLimitations",
		"WithoutAuthority", "WithoutConfigurationReference", "WithoutCorrection",
		"WithoutDeploymentTimestamp", "WithoutExtension",
	}
	if !slices.Equal(modifiers, want) {
		t.Errorf("BindingRecord modifiers = %v, want exactly %v", modifiers, want)
	}
}

// TestUnbindingRecordExposesNoForbiddenMethod audits UnbindingRecord: no
// Artifact identity, no lifecycle, no mutable "unbound"/"current" state.
func TestUnbindingRecordExposesNoForbiddenMethod(t *testing.T) {
	forbidden := append([]string{
		"WithBinding", "WithSubject", "WithTerminatedAt", "WithReason",
		"WithActor", "WithProvenance", "WithID", "WithoutID",
	}, recordForbiddenMethods...)
	assertNoMethods(t, "UnbindingRecord", reflect.TypeOf(UnbindingRecord{}), forbidden)

	typ := reflect.TypeOf(UnbindingRecord{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	want := []string{
		"WithAuthority", "WithCorrection", "WithExtension",
		"WithoutAuthority", "WithoutCorrection", "WithoutExtension",
	}
	if !slices.Equal(modifiers, want) {
		t.Errorf("UnbindingRecord modifiers = %v, want exactly %v", modifiers, want)
	}
}

// TestObservationExposesNoForbiddenMethod audits Observation: no Artifact
// identity, no lifecycle, no execution/outcome API, and -- critically --
// no correction API, since PEOS-008 documents none for Observation.
func TestObservationExposesNoForbiddenMethod(t *testing.T) {
	// "Source" is deliberately excluded from this test's forbidden set:
	// Observation.Source() legitimately returns "the actor or system
	// source" PEOS-008 requires, distinct from a relation endpoint. Every
	// other package-wide forbidden name still applies.
	forbidden := []string{
		"ArtifactID", "RevisionID", "Revision", "Core",
		"Bind", "Unbind", "Deploy", "Activate",
		"SetState", "SetStatus", "SetCompliance",
		"MarkCompliant", "MarkViolated",
		"CurrentBinding", "LatestBinding", "EffectiveBinding",
		"Bound", "ActiveDeployment", "Deployed", "Compliant",
		"Lifecycle", "State", "Status", "StateAssignment",
		"Relation", "RelationType", "Target",
		"WithSubject", "WithScope", "WithEnvironment", "WithObservedAt", "WithObservedValue",
		"WithCollectionMethod", "WithSource", "WithProvenance",
		"WithoutScope", "WithoutEnvironment",
		"WithID", "WithoutID",
		"WithCorrection", "WithoutCorrection",
		"Outcome", "Result", "Success", "Failure", "ExecutionOutcome",
	}
	assertNoMethods(t, "Observation", reflect.TypeOf(Observation{}), forbidden)

	typ := reflect.TypeOf(Observation{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	want := []string{
		"WithBinding", "WithEvidence", "WithExtension",
		"WithInterval", "WithLimitations", "WithUncertainty",
		"WithUnitScaleOrEventType", "WithoutBinding",
		"WithoutExtension", "WithoutInterval", "WithoutUncertainty",
		"WithoutUnitScaleOrEventType",
	}
	if !slices.Equal(modifiers, want) {
		t.Errorf("Observation modifiers = %v, want exactly %v", modifiers, want)
	}
}

// TestViolationTriggerHasExactlyTwoArms asserts ViolationTrigger exposes
// exactly the observation and evidence arms, with no Incident arm and no
// third generic-Artifact arm.
func TestViolationTriggerHasExactlyTwoArms(t *testing.T) {
	forbidden := []string{
		"Incident", "Artifact", "ArtifactRevision", "WithIncident",
	}
	assertNoMethods(t, "ViolationTrigger", reflect.TypeOf(ViolationTrigger{}), forbidden)
}

// TestViolationExposesNoForbiddenMethod audits Violation: no Artifact
// identity, no lifecycle, no mutable status/resolved/closed field, no
// Decision Outcome, no Compliance Claim value, and no correction API,
// since PEOS-008 documents none for Violation and no
// core.RuntimeViolationRef type exists to serve as a correction target.
func TestViolationExposesNoForbiddenMethod(t *testing.T) {
	forbidden := append([]string{
		"WithSubject", "WithCriterion", "WithTrigger", "WithOccurredAt",
		"WithClassification", "WithScope", "WithProvenance",
		"WithID", "WithoutID", "Ref",
		"WithCorrection", "WithoutCorrection",
		"Resolved", "Closed", "Status", "RemediationAuthority",
		"DecisionOutcome", "ComplianceClaim", "Outcome",
		"ApplicableWaiver", "WithApplicableWaiver", "WithoutApplicableWaiver",
	}, recordForbiddenMethods...)
	assertNoMethods(t, "Violation", reflect.TypeOf(Violation{}), forbidden)

	typ := reflect.TypeOf(Violation{})
	var modifiers []string
	for i := range typ.NumMethod() {
		name := typ.Method(i).Name
		if strings.HasPrefix(name, "With") {
			modifiers = append(modifiers, name)
		}
	}
	slices.Sort(modifiers)
	want := []string{
		"WithBinding", "WithExtension", "WithInterval",
		"WithLimitations", "WithRelatedClaims", "WithRelatedDecisions",
		"WithSeverity", "WithUncertainty",
		"WithoutBinding", "WithoutExtension", "WithoutInterval",
		"WithoutSeverity", "WithoutUncertainty",
	}
	if !slices.Equal(modifiers, want) {
		t.Errorf("Violation modifiers = %v, want exactly %v", modifiers, want)
	}
}

// TestComplianceClaimHelperReturnsOrdinaryValidationClaim confirms
// NewComplianceClaim's return type is validation.Claim, not a
// package-local type, by asserting the function's reflected signature.
func TestComplianceClaimHelperReturnsOrdinaryValidationClaim(t *testing.T) {
	fn := reflect.ValueOf(NewComplianceClaim)
	if fn.Type().NumOut() != 2 {
		t.Fatalf("NewComplianceClaim returns %d values, want 2", fn.Type().NumOut())
	}
	if got := fn.Type().Out(0).String(); got != "validation.Claim" {
		t.Errorf("NewComplianceClaim first result = %s, want validation.Claim", got)
	}
	if got := fn.Type().Out(1); got != reflect.TypeOf((*error)(nil)).Elem() {
		t.Errorf("NewComplianceClaim second result = %v, want error", got)
	}
}
