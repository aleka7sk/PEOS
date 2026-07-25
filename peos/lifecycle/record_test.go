package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"

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

func baseContentParts(t *testing.T) (core.LifecycleSubjectRef, core.LifecycleDefinitionVersionRef, TransitionID, StateID, core.Timestamp) {
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
	fromState := mustStateID(t, "draft")
	attemptedAt := mustTimestamp(t, 2026, 1, 1)
	return subject, definitionVersion, transition, fromState, attemptedAt
}

func mustSucceededContent(t *testing.T) TransitionRecordContent {
	t.Helper()
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeSucceeded)
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
	if c.FromState().IsZero() {
		t.Error("FromState() is zero")
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeSucceeded)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeSucceeded)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeSucceeded)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeFailed)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeFailed)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeFailed)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeFailed)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeInterrupted)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeInterrupted)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeInterrupted)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeIndeterminate)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateOutcome(); err != nil {
		t.Errorf("indeterminate content without completed_at unexpectedly rejected: %v", err)
	}
}

func TestIndeterminateContentWithToStateRejected(t *testing.T) {
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeIndeterminate)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeIndeterminate)
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	c, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, NewTransitionOutcome(vocab))
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
	subject, definitionVersion, transition, fromState, attemptedAt := baseContentParts(t)
	content, err := NewTransitionRecordContent(subject, definitionVersion, transition, fromState, attemptedAt, TransitionOutcomeSucceeded)
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
