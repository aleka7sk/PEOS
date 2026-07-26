package validation

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- shared helpers ----------------------------------------------------------

func mustVocab(t *testing.T, namespace, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustLocalKey(t *testing.T, value string) core.LocalKey {
	t.Helper()
	k, err := core.NewLocalKey(value)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func mustArtifactID(t *testing.T, value string) core.ArtifactID {
	t.Helper()
	id, err := core.NewArtifactID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustArtifactRevisionID(t *testing.T, value string) core.ArtifactRevisionID {
	t.Helper()
	id, err := core.NewArtifactRevisionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustArtifactRevisionRef(t *testing.T, artifactID, revisionID string) core.ArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustSubject(t *testing.T, artifactID, revisionID string) core.EngineeringSubjectRef {
	t.Helper()
	subject, err := core.EngineeringSubjectRefFromArtifactRevision(mustArtifactRevisionRef(t, artifactID, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return subject
}

func mustMethod(t *testing.T, value string) core.ValidationMethod {
	t.Helper()
	return core.NewValidationMethod(mustVocab(t, core.PEOSNamespace, value))
}

func mustCriterion(t *testing.T, artifactID, revisionID string) core.CriterionRef {
	t.Helper()
	ref, err := core.NewRequirementArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	c, err := core.CriterionRefFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustAuthorityRef(t *testing.T, namespace, identifier string) core.AuthorityRef {
	t.Helper()
	ref, err := core.NewAuthorityRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustProvenance(t *testing.T) core.Provenance {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	actor, err := core.NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewProvenance().WithActor(actor).WithRecordedAt(ts)
}

func mustScope(t *testing.T, kind, expression string) core.Scope {
	t.Helper()
	s, err := core.NewScope(mustVocab(t, core.PEOSNamespace, kind), expression)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustExtension(t *testing.T, namespace, payload string) core.Extension {
	t.Helper()
	e, err := core.NewExtension().With(namespace, json.RawMessage(payload))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// mustActivity returns a minimum-valid PlannedActivity keyed "ACT-1".
func mustActivity(t *testing.T) PlannedActivity {
	t.Helper()
	return mustActivityKeyed(t, "ACT-1")
}

func mustActivityKeyed(t *testing.T, key string) PlannedActivity {
	t.Helper()
	a, err := NewPlannedActivity(
		mustLocalKey(t, key),
		mustSubject(t, "AR-42", "REV-3"),
		mustMethod(t, "test"),
		"pass iff all assertions hold",
	)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// --- construction ------------------------------------------------------------

func TestNewPlannedActivityValidMinimum(t *testing.T) {
	a := mustActivity(t)
	if a.IsZero() {
		t.Fatal("valid PlannedActivity reports IsZero")
	}
	if got := a.Key().String(); got != "ACT-1" {
		t.Errorf("Key() = %q, want ACT-1", got)
	}
	if got := a.OutcomeInterpretation(); got != "pass iff all assertions hold" {
		t.Errorf("OutcomeInterpretation() = %q", got)
	}
	if len(a.Criteria()) != 0 || len(a.ExpectedEvidence()) != 0 || len(a.Prerequisites()) != 0 || len(a.Dependencies()) != 0 {
		t.Error("minimum PlannedActivity has a non-empty optional collection")
	}
	if _, ok := a.ResponsibleRole(); ok {
		t.Error("minimum PlannedActivity reports a responsible role")
	}
	if _, ok := a.RequiredAuthority(); ok {
		t.Error("minimum PlannedActivity reports a required authority")
	}
	if _, ok := a.MethodDefinition(); ok {
		t.Error("minimum PlannedActivity reports a method definition")
	}
	if !a.Extension().IsZero() {
		t.Error("minimum PlannedActivity has extension data")
	}
}

func TestNewPlannedActivityZeroKeyRejected(t *testing.T) {
	_, err := NewPlannedActivity(core.LocalKey{}, mustSubject(t, "AR-42", "REV-3"), mustMethod(t, "test"), "x")
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestNewPlannedActivityZeroSubjectRejected(t *testing.T) {
	_, err := NewPlannedActivity(mustLocalKey(t, "ACT-1"), core.EngineeringSubjectRef{}, mustMethod(t, "test"), "x")
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestNewPlannedActivityZeroMethodRejected(t *testing.T) {
	_, err := NewPlannedActivity(mustLocalKey(t, "ACT-1"), mustSubject(t, "AR-42", "REV-3"), core.ValidationMethod{}, "x")
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestNewPlannedActivityEmptyOutcomeInterpretationRejected(t *testing.T) {
	for _, value := range []string{"", "   ", "\t\n "} {
		_, err := NewPlannedActivity(mustLocalKey(t, "ACT-1"), mustSubject(t, "AR-42", "REV-3"), mustMethod(t, "test"), value)
		if !errors.Is(err, ErrInvalidPlannedActivity) {
			t.Errorf("outcome interpretation %q: error = %v, want %v", value, err, ErrInvalidPlannedActivity)
		}
	}
}

func TestNewPlannedActivityTrimsOutcomeInterpretation(t *testing.T) {
	a, err := NewPlannedActivity(mustLocalKey(t, "ACT-1"), mustSubject(t, "AR-42", "REV-3"), mustMethod(t, "test"), "  interpret me  ")
	if err != nil {
		t.Fatal(err)
	}
	if got := a.OutcomeInterpretation(); got != "interpret me" {
		t.Errorf("OutcomeInterpretation() = %q, want %q", got, "interpret me")
	}
}

func TestNewPlannedActivityAcceptsIdentityLevelSubject(t *testing.T) {
	ref, err := core.NewRequirementRef(mustArtifactID(t, "REQ-1"))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := core.EngineeringSubjectRefFromRequirement(ref)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewPlannedActivity(mustLocalKey(t, "ACT-1"), subject, mustMethod(t, "review"), "x"); err != nil {
		t.Fatalf("identity-level subject rejected: %v", err)
	}
}

func TestPlannedActivityIsZero(t *testing.T) {
	var a PlannedActivity
	if !a.IsZero() {
		t.Error("zero PlannedActivity does not report IsZero")
	}
}

// --- modifiers ---------------------------------------------------------------

func TestPlannedActivityWithCriteria(t *testing.T) {
	a, err := mustActivity(t).WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Criteria()) != 1 {
		t.Fatalf("Criteria() length = %d, want 1", len(a.Criteria()))
	}
}

func TestPlannedActivityWithCriteriaEmptyAccepted(t *testing.T) {
	withCriteria, err := mustActivity(t).WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")})
	if err != nil {
		t.Fatal(err)
	}
	cleared, err := withCriteria.WithCriteria(nil)
	if err != nil {
		t.Fatalf("WithCriteria(nil) rejected: %v", err)
	}
	if len(cleared.Criteria()) != 0 {
		t.Error("WithCriteria(nil) did not clear criteria")
	}
	cleared2, err := withCriteria.WithCriteria([]core.CriterionRef{})
	if err != nil {
		t.Fatalf("WithCriteria(empty) rejected: %v", err)
	}
	if len(cleared2.Criteria()) != 0 {
		t.Error("WithCriteria(empty) did not clear criteria")
	}
}

// TestPlannedActivityOptionalCollectionsClearedByEmptyInput locks that each
// optional collection is cleared by an empty or nil input, which is why no
// Without* counterpart exists for any of them.
func TestPlannedActivityOptionalCollectionsClearedByEmptyInput(t *testing.T) {
	a := fullActivity(t)
	var err error
	if a, err = a.WithExpectedEvidence(nil); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithPrerequisites([]string{}); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithDependencies(nil); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithCriteria([]core.CriterionRef{}); err != nil {
		t.Fatal(err)
	}
	if len(a.ExpectedEvidence()) != 0 || len(a.Prerequisites()) != 0 || len(a.Dependencies()) != 0 || len(a.Criteria()) != 0 {
		t.Error("empty input did not clear an optional collection")
	}
	if a.IsZero() {
		t.Error("clearing optional collections invalidated the activity")
	}
}

func TestPlannedActivityMalformedDocumentRejected(t *testing.T) {
	for _, payload := range []string{`not json`, `[]`, `{"key":[]}`} {
		var a PlannedActivity
		if err := json.Unmarshal([]byte(payload), &a); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

func TestPlannedActivityJSONOptionalSingleWrongTypeRejected(t *testing.T) {
	cases := map[string]string{
		"responsible_role":   `123`,
		"required_authority": `"not-an-object"`,
		"method_definition":  `42`,
	}
	for field, value := range cases {
		t.Run(field, func(t *testing.T) {
			var a PlannedActivity
			err := json.Unmarshal([]byte(activityPayload(t, map[string]string{field: value})), &a)
			if !errors.Is(err, ErrInvalidPlannedActivity) {
				t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
			}
		})
	}
}

func TestPlannedActivityWithCriteriaZeroElementRejected(t *testing.T) {
	_, err := mustActivity(t).WithCriteria([]core.CriterionRef{{}})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestPlannedActivityWithCriteriaReplacesWholeCollection(t *testing.T) {
	a, err := mustActivity(t).WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-1", "REV-1"), mustCriterion(t, "REQ-2", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-3", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Criteria()) != 1 {
		t.Errorf("Criteria() length = %d, want 1 after replacement", len(a.Criteria()))
	}
}

func TestPlannedActivityWithCriteriaPreservesOrderAndDuplicates(t *testing.T) {
	dup := mustCriterion(t, "REQ-1", "REV-1")
	other := mustCriterion(t, "REQ-2", "REV-1")
	a, err := mustActivity(t).WithCriteria([]core.CriterionRef{dup, other, dup})
	if err != nil {
		t.Fatalf("duplicate criteria rejected: %v", err)
	}
	got := a.Criteria()
	if len(got) != 3 {
		t.Fatalf("Criteria() length = %d, want 3 (no deduplication)", len(got))
	}
	if got[0] != dup || got[1] != other || got[2] != dup {
		t.Error("Criteria() did not preserve caller order")
	}
}

func TestPlannedActivityWithExpectedEvidence(t *testing.T) {
	a, err := mustActivity(t).WithExpectedEvidence([]string{"  test report  ", "log bundle"})
	if err != nil {
		t.Fatal(err)
	}
	got := a.ExpectedEvidence()
	if len(got) != 2 || got[0] != "test report" || got[1] != "log bundle" {
		t.Errorf("ExpectedEvidence() = %#v, want trimmed [test report log bundle]", got)
	}
}

func TestPlannedActivityWithExpectedEvidenceEmptyElementRejected(t *testing.T) {
	for _, bad := range [][]string{{""}, {"ok", "   "}} {
		_, err := mustActivity(t).WithExpectedEvidence(bad)
		if !errors.Is(err, ErrInvalidPlannedActivity) {
			t.Errorf("input %#v: error = %v, want %v", bad, err, ErrInvalidPlannedActivity)
		}
	}
}

func TestPlannedActivityWithPrerequisites(t *testing.T) {
	a, err := mustActivity(t).WithPrerequisites([]string{" env provisioned "})
	if err != nil {
		t.Fatal(err)
	}
	got := a.Prerequisites()
	if len(got) != 1 || got[0] != "env provisioned" {
		t.Errorf("Prerequisites() = %#v", got)
	}
}

func TestPlannedActivityWithPrerequisitesEmptyElementRejected(t *testing.T) {
	_, err := mustActivity(t).WithPrerequisites([]string{"\t"})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestPlannedActivityWithDependencies(t *testing.T) {
	a, err := mustActivity(t).WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-0")})
	if err != nil {
		t.Fatal(err)
	}
	got := a.Dependencies()
	if len(got) != 1 || got[0].String() != "ACT-0" {
		t.Errorf("Dependencies() = %#v", got)
	}
}

func TestPlannedActivityWithDependenciesZeroKeyRejected(t *testing.T) {
	_, err := mustActivity(t).WithDependencies([]core.LocalKey{{}})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestPlannedActivityWithResponsibleRole(t *testing.T) {
	role := mustVocab(t, "acme", "qa-lead")
	a, err := mustActivity(t).WithResponsibleRole(role)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := a.ResponsibleRole()
	if !ok || !got.Equal(role) {
		t.Errorf("ResponsibleRole() = %v, %v", got, ok)
	}
	cleared := a.WithoutResponsibleRole()
	if _, ok := cleared.ResponsibleRole(); ok {
		t.Error("WithoutResponsibleRole did not clear")
	}
}

func TestPlannedActivityWithResponsibleRoleZeroRejected(t *testing.T) {
	_, err := mustActivity(t).WithResponsibleRole(core.VocabularyValue{})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestPlannedActivityWithRequiredAuthority(t *testing.T) {
	auth := mustAuthorityRef(t, "org", "safety-board")
	a, err := mustActivity(t).WithRequiredAuthority(auth)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := a.RequiredAuthority()
	if !ok || got != auth {
		t.Errorf("RequiredAuthority() = %v, %v", got, ok)
	}
	if _, ok := a.WithoutRequiredAuthority().RequiredAuthority(); ok {
		t.Error("WithoutRequiredAuthority did not clear")
	}
}

func TestPlannedActivityWithRequiredAuthorityZeroRejected(t *testing.T) {
	_, err := mustActivity(t).WithRequiredAuthority(core.AuthorityRef{})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestPlannedActivityWithMethodDefinition(t *testing.T) {
	ref := mustArtifactRevisionRef(t, "PROC-1", "REV-1")
	a, err := mustActivity(t).WithMethodDefinition(ref)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := a.MethodDefinition()
	if !ok || got != ref {
		t.Errorf("MethodDefinition() = %v, %v", got, ok)
	}
	if _, ok := a.WithoutMethodDefinition().MethodDefinition(); ok {
		t.Error("WithoutMethodDefinition did not clear")
	}
}

func TestPlannedActivityWithMethodDefinitionZeroRejected(t *testing.T) {
	_, err := mustActivity(t).WithMethodDefinition(core.ArtifactRevisionRef{})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestPlannedActivityWithExtension(t *testing.T) {
	ext := mustExtension(t, "acme", `{"timeout":30}`)
	a := mustActivity(t).WithExtension(ext)
	if a.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !a.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestPlannedActivityModifierReceiverImmutability(t *testing.T) {
	original := mustActivity(t)

	if _, err := original.WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")}); err != nil {
		t.Fatal(err)
	}
	if len(original.Criteria()) != 0 {
		t.Error("WithCriteria mutated the receiver")
	}
	if _, err := original.WithExpectedEvidence([]string{"x"}); err != nil {
		t.Fatal(err)
	}
	if len(original.ExpectedEvidence()) != 0 {
		t.Error("WithExpectedEvidence mutated the receiver")
	}
	if _, err := original.WithPrerequisites([]string{"x"}); err != nil {
		t.Fatal(err)
	}
	if len(original.Prerequisites()) != 0 {
		t.Error("WithPrerequisites mutated the receiver")
	}
	if _, err := original.WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-0")}); err != nil {
		t.Fatal(err)
	}
	if len(original.Dependencies()) != 0 {
		t.Error("WithDependencies mutated the receiver")
	}
	if _, err := original.WithResponsibleRole(mustVocab(t, "acme", "qa")); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.ResponsibleRole(); ok {
		t.Error("WithResponsibleRole mutated the receiver")
	}
	if _, err := original.WithRequiredAuthority(mustAuthorityRef(t, "org", "board")); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.RequiredAuthority(); ok {
		t.Error("WithRequiredAuthority mutated the receiver")
	}
	if _, err := original.WithMethodDefinition(mustArtifactRevisionRef(t, "PROC-1", "REV-1")); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.MethodDefinition(); ok {
		t.Error("WithMethodDefinition mutated the receiver")
	}
	_ = original.WithExtension(mustExtension(t, "acme", `{}`))
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}
}

func TestPlannedActivityFailedModifierPreservesReceiver(t *testing.T) {
	original, err := mustActivity(t).WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := original.WithCriteria([]core.CriterionRef{{}}); err == nil {
		t.Fatal("expected failure")
	}
	if len(original.Criteria()) != 1 {
		t.Error("failed WithCriteria disturbed the receiver")
	}
}

// --- defensive copying -------------------------------------------------------

func TestPlannedActivityCollectionsCopiedIn(t *testing.T) {
	criteria := []core.CriterionRef{mustCriterion(t, "REQ-1", "REV-1")}
	evidence := []string{"report"}
	prereq := []string{"env"}
	deps := []core.LocalKey{mustLocalKey(t, "ACT-0")}

	a, err := mustActivity(t).WithCriteria(criteria)
	if err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithExpectedEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithPrerequisites(prereq); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithDependencies(deps); err != nil {
		t.Fatal(err)
	}

	criteria[0] = mustCriterion(t, "REQ-999", "REV-9")
	evidence[0] = "tampered"
	prereq[0] = "tampered"
	deps[0] = mustLocalKey(t, "TAMPERED")

	if got, _ := a.Criteria()[0].AsRequirementRevision(); got.ArtifactID().String() != "REQ-1" {
		t.Error("criteria input slice was not copied in")
	}
	if a.ExpectedEvidence()[0] != "report" {
		t.Error("expected evidence input slice was not copied in")
	}
	if a.Prerequisites()[0] != "env" {
		t.Error("prerequisites input slice was not copied in")
	}
	if a.Dependencies()[0].String() != "ACT-0" {
		t.Error("dependencies input slice was not copied in")
	}
}

func TestPlannedActivityCollectionsCopiedOut(t *testing.T) {
	a, err := mustActivity(t).WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-1", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithExpectedEvidence([]string{"report"}); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithPrerequisites([]string{"env"}); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-0")}); err != nil {
		t.Fatal(err)
	}

	a.Criteria()[0] = core.CriterionRef{}
	a.ExpectedEvidence()[0] = "tampered"
	a.Prerequisites()[0] = "tampered"
	a.Dependencies()[0] = core.LocalKey{}

	if a.Criteria()[0].IsZero() {
		t.Error("Criteria() accessor did not return a defensive copy")
	}
	if a.ExpectedEvidence()[0] != "report" {
		t.Error("ExpectedEvidence() accessor did not return a defensive copy")
	}
	if a.Prerequisites()[0] != "env" {
		t.Error("Prerequisites() accessor did not return a defensive copy")
	}
	if a.Dependencies()[0].IsZero() {
		t.Error("Dependencies() accessor did not return a defensive copy")
	}
}

// --- JSON --------------------------------------------------------------------

func fullActivity(t *testing.T) PlannedActivity {
	t.Helper()
	a, err := mustActivity(t).WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")})
	if err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithExpectedEvidence([]string{"test report"}); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithPrerequisites([]string{"env provisioned"}); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-0")}); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithResponsibleRole(mustVocab(t, "acme", "qa-lead")); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithRequiredAuthority(mustAuthorityRef(t, "org", "safety-board")); err != nil {
		t.Fatal(err)
	}
	if a, err = a.WithMethodDefinition(mustArtifactRevisionRef(t, "PROC-1", "REV-1")); err != nil {
		t.Fatal(err)
	}
	return a.WithExtension(mustExtension(t, "acme", `{"timeout":30}`))
}

func TestPlannedActivityJSONMinimumKeys(t *testing.T) {
	data, err := json.Marshal(mustActivity(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"key", "subject", "method", "outcome_interpretation"}
	if len(raw) != len(want) {
		t.Errorf("minimum wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, forbidden := range []string{"relation", "id", "key_id", "revision_id", "artifact_id", "lifecycle", "state", "provenance", "type", "criteria", "extension"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("minimum wire form unexpectedly carries %q", forbidden)
		}
	}
}

func TestPlannedActivityJSONFullKeys(t *testing.T) {
	data, err := json.Marshal(fullActivity(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"key", "subject", "method", "outcome_interpretation",
		"criteria", "expected_evidence", "prerequisites", "dependencies",
		"responsible_role", "required_authority", "method_definition", "extension",
	}
	if len(raw) != len(want) {
		t.Errorf("full wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, forbidden := range []string{"relation", "id", "lifecycle", "state", "provenance", "type"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("full wire form unexpectedly carries %q", forbidden)
		}
	}
}

func TestPlannedActivityJSONRoundTrip(t *testing.T) {
	for name, original := range map[string]PlannedActivity{"minimum": mustActivity(t), "full": fullActivity(t)} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decoded PlannedActivity
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Key() != original.Key() {
				t.Error("key mismatch")
			}
			if decoded.Subject() != original.Subject() {
				t.Error("subject mismatch")
			}
			if !decoded.Method().Value().Equal(original.Method().Value()) {
				t.Error("method mismatch")
			}
			if decoded.OutcomeInterpretation() != original.OutcomeInterpretation() {
				t.Error("outcome interpretation mismatch")
			}
			if len(decoded.Criteria()) != len(original.Criteria()) {
				t.Error("criteria length mismatch")
			}
			if len(decoded.ExpectedEvidence()) != len(original.ExpectedEvidence()) {
				t.Error("expected evidence length mismatch")
			}
			if len(decoded.Prerequisites()) != len(original.Prerequisites()) {
				t.Error("prerequisites length mismatch")
			}
			if len(decoded.Dependencies()) != len(original.Dependencies()) {
				t.Error("dependencies length mismatch")
			}
			gotRole, gotOK := decoded.ResponsibleRole()
			wantRole, wantOK := original.ResponsibleRole()
			if gotOK != wantOK || !gotRole.Equal(wantRole) {
				t.Error("responsible role mismatch")
			}
			gotAuth, gotOK := decoded.RequiredAuthority()
			wantAuth, wantOK := original.RequiredAuthority()
			if gotOK != wantOK || gotAuth != wantAuth {
				t.Error("required authority mismatch")
			}
			gotDef, gotOK := decoded.MethodDefinition()
			wantDef, wantOK := original.MethodDefinition()
			if gotOK != wantOK || gotDef != wantDef {
				t.Error("method definition mismatch")
			}
			if decoded.Extension().IsZero() != original.Extension().IsZero() {
				t.Error("extension presence mismatch")
			}
		})
	}
}

func TestPlannedActivityZeroMarshalRejected(t *testing.T) {
	var a PlannedActivity
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func activityPayload(t *testing.T, overrides map[string]string) string {
	t.Helper()
	base := map[string]string{
		"key":                    `"ACT-1"`,
		"subject":                `{"kind":"artifact_revision","ref":{"artifact_id":"AR-42","revision_id":"REV-3"}}`,
		"method":                 `"peos:test"`,
		"outcome_interpretation": `"pass iff all assertions hold"`,
	}
	for k, v := range overrides {
		if v == "" {
			delete(base, k)
			continue
		}
		base[k] = v
	}
	parts := make([]string, 0, len(base))
	for k, v := range base {
		parts = append(parts, `"`+k+`":`+v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func TestPlannedActivityJSONMandatoryMissingRejected(t *testing.T) {
	for _, field := range []string{"key", "subject", "method", "outcome_interpretation"} {
		t.Run(field, func(t *testing.T) {
			var a PlannedActivity
			err := json.Unmarshal([]byte(activityPayload(t, map[string]string{field: ""})), &a)
			if !errors.Is(err, ErrInvalidPlannedActivity) {
				t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
			}
		})
	}
}

func TestPlannedActivityJSONMandatoryNullRejected(t *testing.T) {
	for _, field := range []string{"key", "subject", "method", "outcome_interpretation"} {
		t.Run(field, func(t *testing.T) {
			var a PlannedActivity
			err := json.Unmarshal([]byte(activityPayload(t, map[string]string{field: "null"})), &a)
			if err == nil {
				t.Fatal("explicit null accepted, want error")
			}
			if !errors.Is(err, ErrInvalidPlannedActivity) {
				t.Errorf("error = %v, want wrapping %v", err, ErrInvalidPlannedActivity)
			}
		})
	}
}

// TestPlannedActivityJSONOptionalCollectionsAbsentNullEmptyEquivalent locks
// the documented decision that, for an Activity's four optional
// collections, an absent key, an explicit null, and an empty array all
// denote the same valid state -- "none declared" -- because PEOS-006
// permits zero cardinality for each. A Validation Claim's criteria will
// deliberately NOT behave this way in Packet H.2.
func TestPlannedActivityJSONOptionalCollectionsAbsentNullEmptyEquivalent(t *testing.T) {
	for _, field := range []string{"criteria", "expected_evidence", "prerequisites", "dependencies"} {
		for _, value := range []string{"", "null", "[]"} {
			label := field + "=" + value
			if value == "" {
				label = field + "=absent"
			}
			t.Run(label, func(t *testing.T) {
				var a PlannedActivity
				if err := json.Unmarshal([]byte(activityPayload(t, map[string]string{field: value})), &a); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if a.IsZero() {
					t.Fatal("decoded activity is zero")
				}
				if len(a.Criteria()) != 0 || len(a.ExpectedEvidence()) != 0 || len(a.Prerequisites()) != 0 || len(a.Dependencies()) != 0 {
					t.Error("decoded activity has a non-empty optional collection")
				}
			})
		}
	}
}

func TestPlannedActivityJSONOptionalSingleNullRejected(t *testing.T) {
	for _, field := range []string{"responsible_role", "required_authority", "method_definition"} {
		t.Run(field, func(t *testing.T) {
			var a PlannedActivity
			err := json.Unmarshal([]byte(activityPayload(t, map[string]string{field: "null"})), &a)
			if !errors.Is(err, ErrInvalidPlannedActivity) {
				t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
			}
		})
	}
}

// TestPlannedActivityJSONWhitespaceCollectionElementRejected covers the two
// collection error paths that are genuinely reachable through JSON. A
// decoded core.CriterionRef and a decoded core.LocalKey can never be zero
// (each rejects an empty payload in its own UnmarshalJSON), so those two
// collections' element checks are unreachable defensive guards; a plain
// JSON string, however, decodes happily as whitespace and must be rejected
// here.
func TestPlannedActivityJSONWhitespaceCollectionElementRejected(t *testing.T) {
	for _, field := range []string{"expected_evidence", "prerequisites"} {
		t.Run(field, func(t *testing.T) {
			var a PlannedActivity
			err := json.Unmarshal([]byte(activityPayload(t, map[string]string{field: `["   "]`})), &a)
			if !errors.Is(err, ErrInvalidPlannedActivity) {
				t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
			}
		})
	}
}

func TestPlannedActivityJSONExtensionNullMeansAbsent(t *testing.T) {
	var a PlannedActivity
	if err := json.Unmarshal([]byte(activityPayload(t, map[string]string{"extension": "null"})), &a); err != nil {
		t.Fatalf("extension null rejected: %v", err)
	}
	if !a.Extension().IsZero() {
		t.Error("extension null did not decode as absent")
	}
}

func TestPlannedActivityJSONUnknownFieldIgnored(t *testing.T) {
	var a PlannedActivity
	if err := json.Unmarshal([]byte(activityPayload(t, map[string]string{"unknown_field": `"x"`})), &a); err != nil {
		t.Fatalf("unknown field rejected: %v", err)
	}
}

func TestPlannedActivityJSONNestedSentinelsPreserved(t *testing.T) {
	cases := []struct {
		name    string
		payload map[string]string
		want    error
	}{
		{"subject missing revision id", map[string]string{"subject": `{"kind":"artifact_revision","ref":{"artifact_id":"AR-42"}}`}, core.ErrMissingRevisionID},
		{"subject missing discriminator", map[string]string{"subject": `{"ref":{"artifact_id":"AR-42"}}`}, core.ErrInvalidReferenceDiscriminator},
		{"subject unknown kind with composite ref", map[string]string{"subject": `{"kind":"not_a_kind","ref":{"profile":{"artifact_id":"X","revision_id":"Y"},"element":"E"}}`}, core.ErrEmptyIdentity},
		{"criterion missing revision id", map[string]string{"criteria": `[{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7"}}]`}, core.ErrMissingRevisionID},
		{"criterion missing discriminator", map[string]string{"criteria": `[{"ref":{"artifact_id":"REQ-7"}}]`}, core.ErrInvalidReferenceDiscriminator},
		{"empty key identity", map[string]string{"key": `"   "`}, core.ErrEmptyIdentity},
		{"malformed method vocabulary", map[string]string{"method": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"malformed responsible role vocabulary", map[string]string{"responsible_role": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"required authority empty identity", map[string]string{"required_authority": `{"namespace":"org","identifier":"  "}`}, core.ErrEmptyIdentity},
		{"method definition missing revision", map[string]string{"method_definition": `{"artifact_id":"PROC-1"}`}, core.ErrMissingRevisionID},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var a PlannedActivity
			err := json.Unmarshal([]byte(activityPayload(t, tc.payload)), &a)
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want wrapping %v", err, tc.want)
			}
		})
	}
}

// TestPlannedActivityJSONOpaqueSubjectAndCriterionPreserved locks
// peos/core's forward-compatibility contract as consumed here: an
// unrecognized subject or criterion kind carrying a well-formed namespaced
// scalar payload is preserved opaquely rather than rejected, so a Plan
// written by a newer PEOS revision still decodes.
func TestPlannedActivityJSONOpaqueSubjectAndCriterionPreserved(t *testing.T) {
	payload := activityPayload(t, map[string]string{
		"subject":  `{"kind":"future_subject","ref":{"namespace":"acme","identifier":"thing-1"}}`,
		"criteria": `[{"kind":"future_criterion","ref":{"namespace":"acme","identifier":"rule-1"}}]`,
	})
	var a PlannedActivity
	if err := json.Unmarshal([]byte(payload), &a); err != nil {
		t.Fatalf("opaque subject/criterion rejected: %v", err)
	}
	if got := a.Subject().Kind(); got != "future_subject" {
		t.Errorf("Subject().Kind() = %q, want future_subject", got)
	}
	criteria := a.Criteria()
	if len(criteria) != 1 {
		t.Fatalf("Criteria() length = %d, want 1", len(criteria))
	}
	if got := criteria[0].Kind(); got != "future_criterion" {
		t.Errorf("criterion Kind() = %q, want future_criterion", got)
	}
	if criteria[0].IsKnown() {
		t.Error("opaque criterion reports IsKnown")
	}
}

func TestPlannedActivityFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := fullActivity(t)
	before := receiver.Key()
	if err := json.Unmarshal([]byte(`{"key":null}`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.Key() != before {
		t.Error("failed Unmarshal disturbed the receiver")
	}
	if len(receiver.Criteria()) != 1 {
		t.Error("failed Unmarshal disturbed the receiver's criteria")
	}
}

func TestPlannedActivityConstructorUnmarshalEquivalence(t *testing.T) {
	original := fullActivity(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded PlannedActivity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(again) {
		t.Errorf("re-marshal differs:\n%s\n%s", data, again)
	}
}

// --- structural absence ------------------------------------------------------

// TestPlannedActivityHasNoIdentityOrRelationshipAPI proves structurally
// that a Planned Validation Activity is neither independently identified
// nor a relationship, and that none of its mandatory state is reachable
// through a With*/Without* call.
func TestPlannedActivityHasNoIdentityOrRelationshipAPI(t *testing.T) {
	forbidden := []string{
		"ID", "Ref", "ArtifactID", "RevisionID", "Core",
		"Relation", "Source", "Target",
		"Lifecycle", "State", "Provenance",
		"WithKey", "WithoutKey",
		"WithSubject", "WithoutSubject",
		"WithMethod", "WithoutMethod",
		"WithOutcomeInterpretation", "WithoutOutcomeInterpretation",
		"WithoutCriteria",
		"Outcome", "Claim", "ExecutionRecord", "Evidence",
	}
	typ := reflect.TypeOf(PlannedActivity{})
	for _, name := range forbidden {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("PlannedActivity unexpectedly exposes %s", name)
		}
		if _, ok := reflect.PointerTo(typ).MethodByName(name); ok {
			t.Errorf("*PlannedActivity unexpectedly exposes %s", name)
		}
	}
}
