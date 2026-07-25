package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustStateAssignmentID(t *testing.T, value string) core.StateAssignmentID {
	t.Helper()
	id, err := core.NewStateAssignmentID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustLifecycleSubject(t *testing.T, artifactID string) core.LifecycleSubjectRef {
	t.Helper()
	id, err := core.NewArtifactID(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewArtifactRef(id)
	if err != nil {
		t.Fatal(err)
	}
	sub, err := core.NewLifecycleSubjectRefFromArtifact(ref)
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func mustArtifactRevisionRef(t *testing.T, artifactID, revisionID string) core.ArtifactRevisionRef {
	t.Helper()
	aid, err := core.NewArtifactID(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	rid, err := core.NewArtifactRevisionID(revisionID)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewArtifactRevisionRef(aid, rid)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustTimestamp(t *testing.T, year, month, day int) core.Timestamp {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func mustAuthorityRef(t *testing.T, namespace, identifier string) core.AuthorityRef {
	t.Helper()
	a, err := core.NewAuthorityRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// validStateAssignmentParts returns a minimally valid set of
// NewStateAssignment arguments.
func validStateAssignmentParts(t *testing.T) (
	id core.StateAssignmentID,
	subject core.LifecycleSubjectRef,
	definitionVersion core.LifecycleDefinitionVersionRef,
	state StateID,
	effectiveAt core.Timestamp,
	provenance core.Provenance,
	establishedBy core.ArtifactRevisionRef,
) {
	t.Helper()
	id = mustStateAssignmentID(t, "SA-1001")
	subject = mustLifecycleSubject(t, "REQ-1")
	defID := mustLifecycleDefinitionID(t, "LC-REVIEW-1")
	versionID, err := core.NewLifecycleDefinitionVersionID("V1")
	if err != nil {
		t.Fatal(err)
	}
	definitionVersion, err = core.NewLifecycleDefinitionVersionRef(defID, versionID)
	if err != nil {
		t.Fatal(err)
	}
	state = mustStateID(t, "accepted")
	effectiveAt = mustTimestamp(t, 2026, 1, 1)
	provenance = mustProvenance(t)
	establishedBy = mustArtifactRevisionRef(t, "TR-5001", "REV-1")
	return
}

func mustStateAssignment(t *testing.T) StateAssignment {
	t.Helper()
	a, err := NewStateAssignment(validStateAssignmentParts(t))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestNewStateAssignmentAllRequiredFields(t *testing.T) {
	id, subject, definitionVersion, state, effectiveAt, provenance, establishedBy := validStateAssignmentParts(t)
	a, err := NewStateAssignment(id, subject, definitionVersion, state, effectiveAt, provenance, establishedBy)
	if err != nil {
		t.Fatal(err)
	}
	if a.IsZero() {
		t.Error("valid StateAssignment reports IsZero() = true")
	}
	if a.ID() != id {
		t.Error("ID() mismatch")
	}
	if a.Subject() != subject {
		t.Error("Subject() mismatch")
	}
	if a.DefinitionVersion() != definitionVersion {
		t.Error("DefinitionVersion() mismatch")
	}
	if !a.State().Equal(state) {
		t.Error("State() mismatch")
	}
	if !a.EffectiveAt().Equal(effectiveAt) {
		t.Error("EffectiveAt() mismatch")
	}
	gotActor, gotOK := a.Provenance().Actor()
	wantActor, wantOK := provenance.Actor()
	if gotOK != wantOK || gotActor != wantActor {
		t.Error("Provenance() mismatch")
	}
	if a.EstablishedBy() != establishedBy {
		t.Error("EstablishedBy() mismatch")
	}
}

func TestNewStateAssignmentEachZeroFieldRejected(t *testing.T) {
	id, subject, definitionVersion, state, effectiveAt, provenance, establishedBy := validStateAssignmentParts(t)

	if _, err := NewStateAssignment(core.StateAssignmentID{}, subject, definitionVersion, state, effectiveAt, provenance, establishedBy); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero id: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
	if _, err := NewStateAssignment(id, core.LifecycleSubjectRef{}, definitionVersion, state, effectiveAt, provenance, establishedBy); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero subject: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
	if _, err := NewStateAssignment(id, subject, core.LifecycleDefinitionVersionRef{}, state, effectiveAt, provenance, establishedBy); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero definition version: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
	if _, err := NewStateAssignment(id, subject, definitionVersion, StateID{}, effectiveAt, provenance, establishedBy); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero state: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
	if _, err := NewStateAssignment(id, subject, definitionVersion, state, core.Timestamp{}, provenance, establishedBy); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero effective time: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
	if _, err := NewStateAssignment(id, subject, definitionVersion, state, effectiveAt, core.Provenance{}, establishedBy); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
	if _, err := NewStateAssignment(id, subject, definitionVersion, state, effectiveAt, provenance, core.ArtifactRevisionRef{}); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("zero established-by: error = %v, want %v", err, ErrInvalidStateAssignment)
	}
}

func TestStateAssignmentOptionalAuthorityAbsent(t *testing.T) {
	a := mustStateAssignment(t)
	if _, ok := a.Authority(); ok {
		t.Error("Authority() ok = true for a StateAssignment with no declared authority")
	}
}

func TestStateAssignmentOptionalAuthorityPresent(t *testing.T) {
	a := mustStateAssignment(t)
	authority := mustAuthorityRef(t, "peos-governance", "release-manager")
	withAuth, err := a.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withAuth.Authority()
	if !ok || got != authority {
		t.Errorf("Authority() = (%v, %v), want (%v, true)", got, ok, authority)
	}
}

func TestStateAssignmentAuthorityNullRejected(t *testing.T) {
	payload := `{"id":"SA-1001","subject":{"kind":"artifact","ref":{"artifact_id":"REQ-1"}},` +
		`"definition_version":{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"},` +
		`"state":"accepted","effective_at":"2026-01-01T00:00:00Z",` +
		`"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}},` +
		`"established_by":{"artifact_id":"TR-5001","revision_id":"REV-1"},"authority":null}`
	var a StateAssignment
	if err := json.Unmarshal([]byte(payload), &a); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStateAssignment)
	}
}

func TestStateAssignmentWithAuthorityReplacement(t *testing.T) {
	a := mustStateAssignment(t)
	first := mustAuthorityRef(t, "peos-governance", "release-manager")
	second := mustAuthorityRef(t, "peos-governance", "qa-lead")

	withFirst, err := a.WithAuthority(first)
	if err != nil {
		t.Fatal(err)
	}
	withSecond, err := withFirst.WithAuthority(second)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withSecond.Authority()
	if !ok || got != second {
		t.Errorf("Authority() = (%v, %v), want (%v, true)", got, ok, second)
	}
}

func TestStateAssignmentWithoutAuthority(t *testing.T) {
	a := mustStateAssignment(t)
	authority := mustAuthorityRef(t, "peos-governance", "release-manager")
	withAuth, err := a.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	cleared := withAuth.WithoutAuthority()
	if _, ok := cleared.Authority(); ok {
		t.Error("Authority() ok = true after WithoutAuthority()")
	}
}

func TestStateAssignmentIsNonArtifactIdentity(t *testing.T) {
	a := mustStateAssignment(t)
	ref, err := a.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.StateAssignmentID() != a.ID() {
		t.Error("Ref() component mismatch")
	}
	// StateAssignment carries no ArtifactID/ArtifactRevisionID field at
	// all -- there is no Core() core.Artifact accessor to call, which
	// itself documents the non-Artifact nature of this type at compile
	// time (contrast TransitionRecord.Core() in record_test.go).
}

func TestStateAssignmentJSONHasNoScopeField(t *testing.T) {
	a := mustStateAssignment(t)
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["scope"]; present {
		t.Error(`"scope" field present in StateAssignment JSON; E-13 forbids a record-level Scope`)
	}
}

func TestStateAssignmentExtensionDefensiveCopy(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	a := mustStateAssignment(t)
	withExt := a.WithExtension(ext)
	if !a.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	got, ok := withExt.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestStateAssignmentJSONRoundTrip(t *testing.T) {
	authority := mustAuthorityRef(t, "peos-governance", "release-manager")
	a, err := mustStateAssignment(t).WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded StateAssignment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != a.ID() || decoded.Subject() != a.Subject() || decoded.DefinitionVersion() != a.DefinitionVersion() {
		t.Error("round trip mismatch")
	}
	if !decoded.State().Equal(a.State()) {
		t.Error("round trip changed State")
	}
	gotAuth, ok := decoded.Authority()
	if !ok || gotAuth != authority {
		t.Errorf("Authority() = (%v, %v), want (%v, true)", gotAuth, ok, authority)
	}
}

func TestStateAssignmentUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := mustStateAssignment(t)
	receiver := original
	payload := `{"id":"SA-9999"}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("incomplete payload accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestStateAssignmentZeroMarshalRejected(t *testing.T) {
	var a StateAssignment
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidStateAssignment) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStateAssignment)
	}
}

func TestStateAssignmentWithDefinitionVersionMembership(t *testing.T) {
	v := mustDefinitionVersion(t)
	ref, err := v.Ref()
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewStateAssignment(
		mustStateAssignmentID(t, "SA-1001"),
		mustLifecycleSubject(t, "REQ-1"),
		ref,
		mustStateID(t, "accepted"),
		mustTimestamp(t, 2026, 1, 1),
		mustProvenance(t),
		mustArtifactRevisionRef(t, "TR-5001", "REV-1"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.WithDefinitionVersion(v); err != nil {
		t.Errorf("valid membership unexpectedly rejected: %v", err)
	}

	if _, err := a.WithDefinitionVersion(mustDefinitionVersionWithoutAcceptedState(t)); err == nil {
		t.Fatal("assignment referencing a state absent from the given DefinitionVersion accepted, want error")
	}
}

// mustDefinitionVersionWithoutAcceptedState builds a valid
// DefinitionVersion whose State set does not include "accepted", for
// TestStateAssignmentWithDefinitionVersionMembership's negative case.
func mustDefinitionVersionWithoutAcceptedState(t *testing.T) DefinitionVersion {
	t.Helper()
	draft, err := NewState(mustStateID(t, "draft"), "Not yet reviewed")
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewState(mustStateID(t, "other"), "Some other state")
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTransitionDefinition(mustTransitionID(t, "move"), []StateID{mustStateID(t, "draft")}, []StateID{mustStateID(t, "other")})
	if err != nil {
		t.Fatal(err)
	}
	versionID, err := core.NewLifecycleDefinitionVersionID("V2")
	if err != nil {
		t.Fatal(err)
	}
	v, err := NewDefinitionVersion(
		versionID,
		mustDefinitionRef(t, "LC-REVIEW-1"),
		mustScope(t, "product-x", "always"),
		[]core.VocabularyValue{mustSubjectType(t, "requirement")},
		[]State{draft, other},
		[]StateID{mustStateID(t, "draft")},
		[]TransitionDefinition{tr},
		mustTransitionID(t, "move"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
