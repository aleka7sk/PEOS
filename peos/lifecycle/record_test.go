package lifecycle

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustArtifactID(t *testing.T, value string) core.ArtifactID {
	t.Helper()
	id, err := core.NewArtifactID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustOrigin(t *testing.T) core.Origin {
	t.Helper()
	o, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustIntegrity(t *testing.T) core.IntegrityIdentity {
	t.Helper()
	i, err := core.NewIntegrityIdentity(core.IntegrityMechanismContentAddressedReference, "sha256:deadbeef", core.IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func mustTransitionRecordArtifact(t *testing.T, artifactID string) core.Artifact {
	t.Helper()
	a, err := core.NewArtifact(mustArtifactID(t, artifactID), ArtifactTypeTransitionRecord)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustTransitionRecordArtifactRevision(t *testing.T, artifactID, revisionID string) core.ArtifactRevision {
	t.Helper()
	rev, err := core.NewArtifactRevision(mustArtifactID(t, artifactID), func() core.ArtifactRevisionID {
		id, err := core.NewArtifactRevisionID(revisionID)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}(), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

// --- Failure ---------------------------------------------------------

func TestNewFailureRequiresReason(t *testing.T) {
	if _, err := NewFailure(""); !errors.Is(err, ErrInvalidFailure) {
		t.Errorf("error = %v, want %v", err, ErrInvalidFailure)
	}
}

func TestFailureJSONRoundTrip(t *testing.T) {
	f, err := NewFailure("guard not satisfied")
	if err != nil {
		t.Fatal(err)
	}
	f = f.WithStep("submit-for-review")
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Failure
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Reason() != f.Reason() {
		t.Error("round trip changed Reason")
	}
	gotStep, ok := decoded.Step()
	wantStep, wantOK := f.Step()
	if ok != wantOK || gotStep != wantStep {
		t.Error("round trip changed Step")
	}
}

func TestFailureExtension(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	f := mustFailure(t).WithExtension(ext)
	got, ok := f.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestFailureZeroMarshalRejected(t *testing.T) {
	var f Failure
	if _, err := json.Marshal(f); !errors.Is(err, ErrInvalidFailure) {
		t.Errorf("error = %v, want %v", err, ErrInvalidFailure)
	}
}

// --- TransitionOutcome -------------------------------------------------

func TestTransitionOutcomePredeclaredValues(t *testing.T) {
	if TransitionOutcomeSucceeded.String() != "peos:succeeded" {
		t.Errorf("Succeeded = %q", TransitionOutcomeSucceeded.String())
	}
	if TransitionOutcomeFailed.String() != "peos:failed" {
		t.Errorf("Failed = %q", TransitionOutcomeFailed.String())
	}
	if TransitionOutcomeInterrupted.String() != "peos:interrupted" {
		t.Errorf("Interrupted = %q", TransitionOutcomeInterrupted.String())
	}
	if TransitionOutcomeIndeterminate.String() != "peos:indeterminate" {
		t.Errorf("Indeterminate = %q", TransitionOutcomeIndeterminate.String())
	}
}

// --- TransitionRecord (Artifact identity wrapper) ----------------------

func TestNewTransitionRecordCorrectArtifactTypeAccepted(t *testing.T) {
	if _, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001")); err != nil {
		t.Fatalf("correct ArtifactType unexpectedly rejected: %v", err)
	}
}

func TestNewTransitionRecordWrongArtifactTypeRejected(t *testing.T) {
	// Build an Artifact with an unrelated, valid ArtifactType instead of
	// ArtifactTypeTransitionRecord.
	wrongType := core.NewArtifactType(mustVocab("requirement"))
	artifact, err := core.NewArtifact(mustArtifactID(t, "TR-5001"), wrongType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTransitionRecord(artifact); !errors.Is(err, ErrArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactTypeMismatch)
	}
}

func TestNewTransitionRecordZeroArtifactRejected(t *testing.T) {
	if _, err := NewTransitionRecord(core.Artifact{}); !errors.Is(err, ErrInvalidTransitionRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionRecord)
	}
}

func TestTransitionRecordCore(t *testing.T) {
	artifact := mustTransitionRecordArtifact(t, "TR-5001")
	r, err := NewTransitionRecord(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if r.Core().ID() != artifact.ID() {
		t.Error("Core() mismatch")
	}
}

func TestTransitionRecordJSONRoundTrip(t *testing.T) {
	r, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != r.ID() {
		t.Error("round trip changed ID")
	}
}

func TestTransitionRecordZeroMarshalRejected(t *testing.T) {
	var r TransitionRecord
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidTransitionRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionRecord)
	}
}

// --- TransitionRecordRevision -------------------------------------------

func TestNewTransitionRecordRevisionArtifactIDMatch(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-5001", "REV-1")
	content := mustSucceededContent(t)
	if _, err := NewTransitionRecordRevision(record, revision, content); err != nil {
		t.Fatalf("matching ArtifactID unexpectedly rejected: %v", err)
	}
}

func TestNewTransitionRecordRevisionArtifactIDMismatchRejected(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-DIFFERENT", "REV-1")
	content := mustSucceededContent(t)
	if _, err := NewTransitionRecordRevision(record, revision, content); !errors.Is(err, ErrArtifactIDMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactIDMismatch)
	}
}

func TestNewTransitionRecordRevisionZeroCoreRevisionRejected(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTransitionRecordRevision(record, core.ArtifactRevision{}, mustSucceededContent(t)); !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionRecordRevision)
	}
}

func TestNewTransitionRecordRevisionZeroContentRejected(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-5001", "REV-1")
	if _, err := NewTransitionRecordRevision(record, revision, TransitionRecordContent{}); !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionRecordRevision)
	}
}

func TestTransitionRecordRevisionNestedJSONComposition(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-5001", "REV-1")
	content := mustSucceededContent(t)
	rr, err := NewTransitionRecordRevision(record, revision, content)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(rr)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["core"]; !ok {
		t.Error(`nested "core" key missing from TransitionRecordRevision JSON`)
	}
	if _, ok := raw["content"]; !ok {
		t.Error(`nested "content" key missing from TransitionRecordRevision JSON`)
	}

	var decoded TransitionRecordRevision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Core().ArtifactID() != rr.Core().ArtifactID() {
		t.Error("round trip changed Core")
	}
}

func TestTransitionRecordRevisionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-5001", "REV-1")
	original, err := NewTransitionRecordRevision(record, revision, mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"core":{}}`), &receiver); err == nil {
		t.Fatal("zero core revision accepted, want error")
	}
	if receiver.Core().ArtifactID() != original.Core().ArtifactID() {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestTransitionRecordRevisionRef(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-5001", "REV-1")
	rr, err := NewTransitionRecordRevision(record, revision, mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	ref, err := rr.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != revision.ArtifactID() || ref.RevisionID() != revision.RevisionID() {
		t.Error("Ref() component mismatch")
	}
}

func TestTransitionRecordRevisionZeroMarshalRejected(t *testing.T) {
	var rr TransitionRecordRevision
	if _, err := json.Marshal(rr); !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionRecordRevision)
	}
}

// --- TransitionRecordContent outcome rules ------------------------------

func mustResultingAssignmentRef(t *testing.T) core.StateAssignmentRef {
	t.Helper()
	ref, err := core.NewStateAssignmentRef(mustStateAssignmentID(t, "SA-1001"))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustStateAssignmentRef(t *testing.T, value string) core.StateAssignmentRef {
	t.Helper()
	ref, err := core.NewStateAssignmentRef(mustStateAssignmentID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func baseContentParts(t *testing.T) (core.LifecycleSubjectRef, core.LifecycleDefinitionVersionRef, TransitionID, core.StateAssignmentRef, core.Timestamp) {
	t.Helper()
	subject := mustLifecycleSubject(t, "REQ-1")
	defID := mustLifecycleDefinitionID(t, "LC-REVIEW-1")
	versionID, err := core.NewLifecycleDefinitionVersionID("V1")
	if err != nil {
		t.Fatal(err)
	}
	definitionVersion, err := core.NewLifecycleDefinitionVersionRef(defID, versionID)
	if err != nil {
		t.Fatal(err)
	}
	transition := mustTransitionID(t, "submit-for-review")
	fromAssignment := mustStateAssignmentRef(t, "SA-DRAFT-1")
	attemptedAt := mustTimestamp(t, 2026, 1, 1)
	return subject, definitionVersion, transition, fromAssignment, attemptedAt
}

func mustSucceededContent(t *testing.T) TransitionRecordContent {
	t.Helper()
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeSucceeded)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithToState(mustStateID(t, "accepted"))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t))
	if err != nil {
		t.Fatal(err)
	}
	// A succeeded outcome requires an authority basis (PEOS-003's Authority
	// Invariant, enforced since Packet L.0.E), so the canonical valid
	// succeeded fixture carries one. Use mustSucceededContentWithoutAuthority
	// for the negative cases.
	c, err = c.WithAuthority(mustAuthorityRef(t, "peos", "review-board"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// mustSucceededContentWithoutAuthority builds an otherwise-valid succeeded
// content carrying no authority basis, for the negative authority cases.
func mustSucceededContentWithoutAuthority(t *testing.T) TransitionRecordContent {
	t.Helper()
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeSucceeded)
	if err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithToState(mustStateID(t, "accepted")); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2)); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t)); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Authority(); ok {
		t.Fatal("fixture unexpectedly carries authority")
	}
	return c
}

func TestSucceededContentCompleteAccepted(t *testing.T) {
	c := mustSucceededContent(t)
	if err := c.validateOutcome(); err != nil {
		t.Errorf("complete succeeded content unexpectedly rejected: %v", err)
	}
	if c.Subject().IsZero() {
		t.Error("Subject() is zero")
	}
	if c.DefinitionVersion().IsZero() {
		t.Error("DefinitionVersion() is zero")
	}
	if c.FromAssignment().IsZero() {
		t.Error("FromAssignment() is zero")
	}
	if c.AttemptedAt().IsZero() {
		t.Error("AttemptedAt() is zero")
	}
	if _, ok := c.CompletedAt(); !ok {
		t.Error("CompletedAt() ok = false for a succeeded content with completed_at set")
	}
	if _, ok := c.Failure(); ok {
		t.Error("Failure() ok = true for a succeeded content")
	}
	authority := mustAuthorityRef(t, "peos-governance", "release-manager")
	withAuth, err := c.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withAuth.Authority()
	if !ok || got != authority {
		t.Errorf("Authority() = (%v, %v), want (%v, true)", got, ok, authority)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := c.WithExtension(ext)
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension data")
	}
	if !c.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

func TestSucceededContentWithoutToStateRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeSucceeded)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestSucceededContentWithoutCompletedAtRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeSucceeded)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithToState(mustStateID(t, "accepted"))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestSucceededContentWithoutResultingAssignmentRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeSucceeded)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithToState(mustStateID(t, "accepted"))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func mustFailure(t *testing.T) Failure {
	t.Helper()
	f, err := NewFailure("guard not satisfied")
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFailedContentWithFailureAccepted(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeFailed)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithFailure(mustFailure(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); err != nil {
		t.Errorf("valid failed content unexpectedly rejected: %v", err)
	}
}

func TestFailedContentWithoutFailureRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeFailed)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestFailedContentWithToStateRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeFailed)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithFailure(mustFailure(t))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithToState(mustStateID(t, "draft"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestFailedContentWithResultingAssignmentRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeFailed)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithFailure(mustFailure(t))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestInterruptedContentAccepted(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeInterrupted)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); err != nil {
		t.Errorf("valid interrupted content unexpectedly rejected: %v", err)
	}
}

func TestInterruptedContentWithToStateRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeInterrupted)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithToState(mustStateID(t, "accepted"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestInterruptedContentWithResultingAssignmentRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeInterrupted)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestIndeterminateContentWithoutCompletedAtAccepted(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeIndeterminate)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); err != nil {
		t.Errorf("indeterminate content without completed_at unexpectedly rejected: %v", err)
	}
}

func TestIndeterminateContentWithToStateRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeIndeterminate)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithToState(mustStateID(t, "accepted"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestIndeterminateContentWithResultingAssignmentRejected(t *testing.T) {
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeIndeterminate)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithResultingAssignment(mustResultingAssignmentRef(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); !errors.Is(err, ErrInvalidTransitionOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionOutcome)
	}
}

func TestUnknownProductOutcomeAcceptedUnderGenericRules(t *testing.T) {
	vocab, err := core.NewVocabularyValue("product-x", "escalated")
	if err != nil {
		t.Fatal(err)
	}
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, NewTransitionOutcome(vocab))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); err != nil {
		t.Errorf("unknown Product-defined outcome unexpectedly rejected under generic rules: %v", err)
	}
}

// --- TransitionRecordContent JSON ---------------------------------------

func TestTransitionRecordContentJSONRoundTrip(t *testing.T) {
	c := mustSucceededContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionRecordContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	gotTo, ok := decoded.ToState()
	wantTo, _ := c.ToState()
	if !ok || !gotTo.Equal(wantTo) {
		t.Error("round trip changed ToState")
	}
	if !decoded.Outcome().Equal(c.Outcome()) {
		t.Error("round trip changed Outcome")
	}
}

func TestTransitionRecordContentToStateNullRejected(t *testing.T) {
	c := mustSucceededContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["to_state"] = json.RawMessage(`null`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionRecordContent
	if err := json.Unmarshal(patched, &decoded); !errors.Is(err, ErrInvalidStateID) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStateID)
	}
}

func TestTransitionRecordContentAuthorityNullRejected(t *testing.T) {
	c := mustSucceededContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["authority"] = json.RawMessage(`null`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionRecordContent
	if err := json.Unmarshal(patched, &decoded); err == nil {
		t.Fatal("explicit null authority accepted, want error")
	}
}

func TestTransitionRecordContentUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := mustSucceededContent(t)
	receiver := original
	payload := `{"subject":{"kind":"artifact","ref":{"artifact_id":"REQ-1"}}}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("incomplete payload accepted, want error")
	}
	if !receiver.Transition().Equal(original.Transition()) {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestTransitionRecordContentZeroMarshalRejected(t *testing.T) {
	var c TransitionRecordContent
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionRecordRevision)
	}
}

// --- Cross-record behavior (in-memory, no repository) -------------------

func TestCrossRecordConstructionAndRoundTrip(t *testing.T) {
	// Build a Transition Record Artifact/Revision first.
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-5001"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordArtifactRevision(t, "TR-5001", "REV-1")
	revisionRef, err := core.NewArtifactRevisionRef(revision.ArtifactID(), revision.RevisionID())
	if err != nil {
		t.Fatal(err)
	}

	// A StateAssignment referencing that Transition Record Revision.
	assignment, err := NewStateAssignment(
		mustStateAssignmentID(t, "SA-1001"),
		mustLifecycleSubject(t, "REQ-1"),
		func() core.LifecycleDefinitionVersionRef {
			ref, err := core.NewLifecycleDefinitionVersionRef(mustLifecycleDefinitionID(t, "LC-REVIEW-1"), func() core.LifecycleDefinitionVersionID {
				id, err := core.NewLifecycleDefinitionVersionID("V1")
				if err != nil {
					t.Fatal(err)
				}
				return id
			}())
			if err != nil {
				t.Fatal(err)
			}
			return ref
		}(),
		mustStateID(t, "accepted"),
		mustTimestamp(t, 2026, 1, 2),
		mustProvenance(t),
		revisionRef,
	)
	if err != nil {
		t.Fatal(err)
	}
	assignmentRef, err := assignment.Ref()
	if err != nil {
		t.Fatal(err)
	}

	// A successful TransitionRecordContent referencing that StateAssignment.
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	content, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, TransitionOutcomeSucceeded)
	if err != nil {
		t.Fatal(err)
	}
	content, err = content.WithToState(mustStateID(t, "accepted"))
	if err != nil {
		t.Fatal(err)
	}
	content, err = content.WithCompletedAt(mustTimestamp(t, 2026, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	content, err = content.WithResultingAssignment(assignmentRef)
	if err != nil {
		t.Fatal(err)
	}
	content, err = content.WithAuthority(mustAuthorityRef(t, "peos", "review-board"))
	if err != nil {
		t.Fatal(err)
	}

	recordRevision, err := NewTransitionRecordRevision(record, revision, content)
	if err != nil {
		t.Fatal(err)
	}

	// Every construct independently round-trips through JSON.
	for name, v := range map[string]interface{ MarshalJSON() ([]byte, error) }{
		"assignment":     assignment,
		"recordRevision": recordRevision,
	} {
		data, err := v.MarshalJSON()
		if err != nil {
			t.Fatalf("%s: Marshal: %v", name, err)
		}
		if len(data) == 0 {
			t.Errorf("%s: empty marshal output", name)
		}
	}

	gotAssignmentRef, ok := recordRevision.Content().ResultingAssignment()
	if !ok || gotAssignmentRef != assignmentRef {
		t.Errorf("ResultingAssignment() = (%v, %v), want (%v, true)", gotAssignmentRef, ok, assignmentRef)
	}
}

// --- Packet L.0.C: source State Assignment endpoint ----------------------

func mustProvenanceWithoutActor(t *testing.T) core.Provenance {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return core.NewProvenance().WithRecordedAt(ts)
}

func mustTransitionRecordRevisionWithProvenance(t *testing.T, artifactID, revisionID string, prov core.Provenance) core.ArtifactRevision {
	t.Helper()
	revID, err := core.NewArtifactRevisionID(revisionID)
	if err != nil {
		t.Fatal(err)
	}
	rev, err := core.NewArtifactRevision(mustArtifactID(t, artifactID), revID, mustOrigin(t), prov, mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

// mustOutcomeContent builds a content value for outcome, satisfying exactly
// the field combination validateOutcome requires for it.
func mustOutcomeContent(t *testing.T, outcome TransitionOutcome) TransitionRecordContent {
	t.Helper()
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, outcome)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case outcome.Equal(TransitionOutcomeSucceeded):
		return mustSucceededContent(t)
	case outcome.Equal(TransitionOutcomeFailed):
		if c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2)); err != nil {
			t.Fatal(err)
		}
		f, err := NewFailure("guard rejected")
		if err != nil {
			t.Fatal(err)
		}
		if c, err = c.WithFailure(f); err != nil {
			t.Fatal(err)
		}
	case outcome.Equal(TransitionOutcomeInterrupted):
		if c, err = c.WithCompletedAt(mustTimestamp(t, 2026, 1, 2)); err != nil {
			t.Fatal(err)
		}
	}
	return c
}

func TestTransitionRecordContentFromAssignmentIsAStateAssignmentRef(t *testing.T) {
	c := mustSucceededContent(t)
	if c.FromAssignment().IsZero() {
		t.Fatal("FromAssignment() is zero")
	}
	// The source endpoint must be a State Assignment reference, not a State
	// identifier: PEOS-003 requires "the source State Assignment".
	want := mustStateAssignmentRef(t, "SA-DRAFT-1")
	if c.FromAssignment() != want {
		t.Errorf("FromAssignment() = %v, want %v", c.FromAssignment(), want)
	}
}

func TestTransitionRecordContentNoFromStateMethodRemains(t *testing.T) {
	// Packet L.0.C removed the parallel StateID endpoint entirely; no
	// deprecated accessor may survive alongside FromAssignment.
	rt := reflect.TypeOf(TransitionRecordContent{})
	for _, name := range []string{"FromState", "SourceState"} {
		if _, ok := rt.MethodByName(name); ok {
			t.Errorf("TransitionRecordContent still exposes %s", name)
		}
	}
}

func TestTransitionRecordContentZeroSourceAssignmentRejected(t *testing.T) {
	subject, definitionVersion, transition, _, attemptedAt := baseContentParts(t)
	_, err := NewTransitionRecordContent(subject, definitionVersion, transition, core.StateAssignmentRef{}, attemptedAt, TransitionOutcomeSucceeded)
	if !errors.Is(err, ErrInvalidStateAssignment) {
		t.Fatalf("err = %v, want ErrInvalidStateAssignment", err)
	}
}

func TestTransitionRecordContentSourceAssignmentJSONRoundTrip(t *testing.T) {
	c := mustSucceededContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"source_state_assignment"`) {
		t.Fatalf("wire form missing source_state_assignment key: %s", data)
	}
	if strings.Contains(string(data), `"from_state"`) {
		t.Fatalf("wire form still carries the removed from_state key: %s", data)
	}
	var decoded TransitionRecordContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.FromAssignment() != c.FromAssignment() {
		t.Errorf("FromAssignment() = %v, want %v", decoded.FromAssignment(), c.FromAssignment())
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip not byte-identical:\n got %s\nwant %s", again, data)
	}
}

func TestTransitionRecordContentSourceAssignmentExplicitNullRejected(t *testing.T) {
	data, err := json.Marshal(mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["source_state_assignment"] = json.RawMessage("null")
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionRecordContent
	if err := json.Unmarshal(payload, &decoded); err == nil {
		t.Fatal("explicit null source_state_assignment accepted")
	}
	if !decoded.IsZero() {
		t.Error("receiver mutated by a failed decode")
	}
}

func TestTransitionRecordContentFailedDecodePreservesReceiver(t *testing.T) {
	original := mustSucceededContent(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"subject":{"kind":"artifact","ref":{"artifact_id":"REQ-9"}}}`), &receiver); err == nil {
		t.Fatal("decode of an incomplete payload succeeded")
	}
	if receiver.FromAssignment() != original.FromAssignment() || !receiver.Outcome().Equal(original.Outcome()) {
		t.Error("receiver mutated by a failed decode")
	}
}

// --- Packet L.0.C: responsible Actor invariant ---------------------------

func TestCompletedTransitionWithActorSucceeds(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-6001"))
	if err != nil {
		t.Fatal(err)
	}
	// mustProvenance carries an Actor.
	revision := mustTransitionRecordArtifactRevision(t, "TR-6001", "REV-1")
	if _, err := NewTransitionRecordRevision(record, revision, mustSucceededContent(t)); err != nil {
		t.Fatalf("completed transition with an Actor rejected: %v", err)
	}
}

func TestCompletedTransitionWithoutActorFails(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-6002"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordRevisionWithProvenance(t, "TR-6002", "REV-1", mustProvenanceWithoutActor(t))
	_, err = NewTransitionRecordRevision(record, revision, mustSucceededContent(t))
	if !errors.Is(err, ErrMissingResponsibleActor) {
		t.Fatalf("err = %v, want ErrMissingResponsibleActor", err)
	}
	if !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Error("ErrMissingResponsibleActor must be wrapped inside ErrInvalidTransitionRecordRevision")
	}
}

func TestCompletedTransitionAuthorityWithoutActorStillFails(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-6003"))
	if err != nil {
		t.Fatal(err)
	}
	content, err := mustSucceededContent(t).WithAuthority(mustAuthorityRef(t, "peos", "review-board"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := content.Authority(); !ok {
		t.Fatal("authority not set")
	}
	revision := mustTransitionRecordRevisionWithProvenance(t, "TR-6003", "REV-1", mustProvenanceWithoutActor(t))
	// PEOS-003: "Actor identity and transition authority are distinct."
	// Authority must not satisfy the Actor obligation.
	if _, err := NewTransitionRecordRevision(record, revision, content); !errors.Is(err, ErrMissingResponsibleActor) {
		t.Fatalf("err = %v, want ErrMissingResponsibleActor", err)
	}
}

func TestCompletedTransitionActorWithoutAuthorityFails(t *testing.T) {
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-6004"))
	if err != nil {
		t.Fatal(err)
	}
	content := mustSucceededContentWithoutAuthority(t)
	// mustTransitionRecordArtifactRevision carries an Actor, so the Actor
	// obligation is satisfied and only the authority basis is missing.
	revision := mustTransitionRecordArtifactRevision(t, "TR-6004", "REV-1")
	_, err = NewTransitionRecordRevision(record, revision, content)
	if !errors.Is(err, ErrMissingTransitionAuthority) {
		t.Fatalf("err = %v, want ErrMissingTransitionAuthority", err)
	}
	if !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Error("ErrMissingTransitionAuthority must be wrapped inside ErrInvalidTransitionRecordRevision")
	}
	if errors.Is(err, ErrMissingResponsibleActor) {
		t.Error("an Actor was present; the Actor sentinel must not be reported")
	}
}

func TestNonCompletedTransitionsDoNotRequireActor(t *testing.T) {
	for _, outcome := range []TransitionOutcome{
		TransitionOutcomeFailed,
		TransitionOutcomeInterrupted,
		TransitionOutcomeIndeterminate,
	} {
		t.Run(outcome.String(), func(t *testing.T) {
			record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-7001"))
			if err != nil {
				t.Fatal(err)
			}
			revision := mustTransitionRecordRevisionWithProvenance(t, "TR-7001", "REV-1", mustProvenanceWithoutActor(t))
			// PEOS-003 states the Actor obligation only for a completed
			// Transition; a Transition Attempt that failed, was interrupted,
			// or is indeterminate must not inherit an invented requirement.
			if _, err := NewTransitionRecordRevision(record, revision, mustOutcomeContent(t, outcome)); err != nil {
				t.Fatalf("%s outcome wrongly required an Actor: %v", outcome, err)
			}
		})
	}
}

func TestProductDefinedOutcomeDoesNotRequireActor(t *testing.T) {
	vocab, err := core.NewVocabularyValue("acme", "compensated")
	if err != nil {
		t.Fatal(err)
	}
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	content, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, NewTransitionOutcome(vocab))
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-7002"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordRevisionWithProvenance(t, "TR-7002", "REV-1", mustProvenanceWithoutActor(t))
	if _, err := NewTransitionRecordRevision(record, revision, content); err != nil {
		t.Fatalf("product-defined outcome wrongly required an Actor: %v", err)
	}
}

func TestTransitionRecordRevisionDecodeEnforcesResponsibleActor(t *testing.T) {
	// The invariant must hold on the decode path too, not only on the
	// constructor path -- both route through
	// newTransitionRecordRevisionFromParts.
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-7003"))
	if err != nil {
		t.Fatal(err)
	}
	good := mustTransitionRecordArtifactRevision(t, "TR-7003", "REV-1")
	rr, err := NewTransitionRecordRevision(record, good, mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(rr)
	if err != nil {
		t.Fatal(err)
	}

	bad := mustTransitionRecordRevisionWithProvenance(t, "TR-7003", "REV-1", mustProvenanceWithoutActor(t))
	badRevision, err := json.Marshal(bad)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["core"] = badRevision
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}

	receiver := rr
	if err := json.Unmarshal(payload, &receiver); !errors.Is(err, ErrMissingResponsibleActor) {
		t.Fatalf("err = %v, want ErrMissingResponsibleActor", err)
	}
	if _, ok := receiver.Core().Provenance().Actor(); !ok {
		t.Error("receiver mutated by a failed decode")
	}
}

// --- Packet L.0.E: completed-transition authority basis ------------------

// succeededRevisionParts returns a Transition Record and a Revision whose
// Provenance carries an Actor, for the authority matrix below.
func succeededRevisionParts(t *testing.T, artifactID string, withActor bool) (TransitionRecord, core.ArtifactRevision) {
	t.Helper()
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	if withActor {
		return record, mustTransitionRecordArtifactRevision(t, artifactID, "REV-1")
	}
	return record, mustTransitionRecordRevisionWithProvenance(t, artifactID, "REV-1", mustProvenanceWithoutActor(t))
}

// TestSucceededTransitionRequiresActorAndAuthority covers the full 2x2 matrix
// of the two independent obligations PEOS-003 states for a completed
// Transition, and asserts that each failure names its own cause.
func TestSucceededTransitionRequiresActorAndAuthority(t *testing.T) {
	cases := []struct {
		name          string
		actor         bool
		authority     bool
		wantErr       error
		wantNotErr    error
		wantSucceeded bool
	}{
		{name: "actor and authority", actor: true, authority: true, wantSucceeded: true},
		{name: "actor without authority", actor: true, authority: false, wantErr: ErrMissingTransitionAuthority, wantNotErr: ErrMissingResponsibleActor},
		{name: "authority without actor", actor: false, authority: true, wantErr: ErrMissingResponsibleActor, wantNotErr: ErrMissingTransitionAuthority},
		{name: "neither", actor: false, authority: false, wantErr: ErrMissingResponsibleActor},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record, revision := succeededRevisionParts(t, "TR-AUTH-"+string(rune('A'+i)), tc.actor)
			content := mustSucceededContentWithoutAuthority(t)
			if tc.authority {
				content = mustSucceededContent(t)
				if _, ok := content.Authority(); !ok {
					t.Fatal("fixture lost its authority")
				}
			}

			_, err := NewTransitionRecordRevision(record, revision, content)
			if tc.wantSucceeded {
				if err != nil {
					t.Fatalf("valid succeeded revision rejected: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			// Both specific sentinels must remain matchable through the
			// general revision-level sentinel.
			if !errors.Is(err, ErrInvalidTransitionRecordRevision) {
				t.Error("cause must be wrapped inside ErrInvalidTransitionRecordRevision")
			}
			if tc.wantNotErr != nil && errors.Is(err, tc.wantNotErr) {
				t.Errorf("err also matched %v, which is satisfied in this case", tc.wantNotErr)
			}
		})
	}
}

func TestNonSucceededOutcomesDoNotRequireAuthority(t *testing.T) {
	for _, outcome := range []TransitionOutcome{
		TransitionOutcomeFailed,
		TransitionOutcomeInterrupted,
		TransitionOutcomeIndeterminate,
	} {
		t.Run(outcome.String(), func(t *testing.T) {
			record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-AUTH-N"))
			if err != nil {
				t.Fatal(err)
			}
			// No Actor and no Authority: neither obligation applies, because
			// PEOS-003 states both only for a completed Transition.
			revision := mustTransitionRecordRevisionWithProvenance(t, "TR-AUTH-N", "REV-1", mustProvenanceWithoutActor(t))
			content := mustOutcomeContent(t, outcome)
			if _, ok := content.Authority(); ok {
				t.Fatal("fixture unexpectedly carries authority")
			}
			if _, err := NewTransitionRecordRevision(record, revision, content); err != nil {
				t.Fatalf("%s outcome wrongly required an authority basis: %v", outcome, err)
			}
		})
	}
}

func TestProductDefinedOutcomeDoesNotRequireAuthority(t *testing.T) {
	vocab, err := core.NewVocabularyValue("acme", "compensated")
	if err != nil {
		t.Fatal(err)
	}
	subject, definitionVersion, transition, fromAssignment, attemptedAt := baseContentParts(t)
	content, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromAssignment, attemptedAt, NewTransitionOutcome(vocab))
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewTransitionRecord(mustTransitionRecordArtifact(t, "TR-AUTH-P"))
	if err != nil {
		t.Fatal(err)
	}
	revision := mustTransitionRecordRevisionWithProvenance(t, "TR-AUTH-P", "REV-1", mustProvenanceWithoutActor(t))
	// A Product-declared outcome follows the documented rule: this package has
	// no normative basis to constrain it, so neither obligation is imposed.
	if _, err := NewTransitionRecordRevision(record, revision, content); err != nil {
		t.Fatalf("product-defined outcome wrongly required an authority basis: %v", err)
	}
}

func TestTransitionRecordRevisionDecodeRejectsSucceededWithoutAuthority(t *testing.T) {
	record, revision := succeededRevisionParts(t, "TR-AUTH-D", true)
	valid, err := NewTransitionRecordRevision(record, revision, mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}

	// Strip the authority key from the encoded content and re-decode.
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(data, &outer); err != nil {
		t.Fatal(err)
	}
	var inner map[string]json.RawMessage
	if err := json.Unmarshal(outer["content"], &inner); err != nil {
		t.Fatal(err)
	}
	if _, ok := inner["authority"]; !ok {
		t.Fatal("valid succeeded content did not encode an authority key")
	}
	delete(inner, "authority")
	if outer["content"], err = json.Marshal(inner); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(outer)
	if err != nil {
		t.Fatal(err)
	}

	receiver := valid
	err = json.Unmarshal(payload, &receiver)
	if !errors.Is(err, ErrMissingTransitionAuthority) {
		t.Fatalf("err = %v, want ErrMissingTransitionAuthority", err)
	}
	if !errors.Is(err, ErrInvalidTransitionRecordRevision) {
		t.Error("decode cause must be wrapped inside ErrInvalidTransitionRecordRevision")
	}
	if _, ok := receiver.Content().Authority(); !ok {
		t.Error("receiver mutated by a failed decode")
	}
}

// TestConstructorAndDecodePathsAgreeOnAuthority proves the invariant is not
// enforced on only one path: the same content/revision pair must be accepted
// or rejected identically whether it is constructed or decoded.
func TestConstructorAndDecodePathsAgreeOnAuthority(t *testing.T) {
	record, revision := succeededRevisionParts(t, "TR-AUTH-E", true)

	// A valid pair encodes, decodes, and round-trips byte-identically with its
	// authority retained.
	valid, err := NewTransitionRecordRevision(record, revision, mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionRecordRevision
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("constructor accepted but decode rejected: %v", err)
	}
	gotAuthority, ok := decoded.Content().Authority()
	if !ok {
		t.Fatal("authority lost through the round trip")
	}
	wantAuthority, _ := valid.Content().Authority()
	if gotAuthority != wantAuthority {
		t.Errorf("Authority() = %v, want %v", gotAuthority, wantAuthority)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(encoded) {
		t.Errorf("round trip not byte-identical:\n got %s\nwant %s", again, encoded)
	}

	// An invalid pair is rejected by the constructor, and the equivalent
	// payload is rejected by the decoder with the same sentinel.
	_, ctorErr := NewTransitionRecordRevision(record, revision, mustSucceededContentWithoutAuthority(t))
	if !errors.Is(ctorErr, ErrMissingTransitionAuthority) {
		t.Fatalf("constructor err = %v, want ErrMissingTransitionAuthority", ctorErr)
	}
}

func TestTransitionRecordContentAuthorityExplicitNullRejected(t *testing.T) {
	// authority has an optional wire representation (omitted when absent), so
	// an explicit null is a malformed payload rather than "absent" -- the same
	// treatment every other optional field in this content receives. It is
	// rejected before the succeeded-outcome authority invariant is reached.
	data, err := json.Marshal(mustSucceededContent(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["authority"] = json.RawMessage("null")
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	original := mustSucceededContent(t)
	receiver := original
	if err := json.Unmarshal(payload, &receiver); err == nil {
		t.Fatal("explicit null authority accepted")
	}
	got, ok := receiver.Authority()
	want, _ := original.Authority()
	if !ok || got != want {
		t.Error("receiver mutated by a failed decode")
	}
}
