package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustDecisionRecordArtifact(t *testing.T, artifactID string) core.Artifact {
	t.Helper()
	aid, err := core.NewArtifactID(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	a, err := core.NewArtifact(aid, ArtifactTypeDecisionRecord)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustTestDecisionRef(t *testing.T, value string) core.DecisionRef {
	t.Helper()
	ref, err := core.NewDecisionRef(mustDecisionID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustArtifactRevision(t *testing.T, artifactID, revisionID string) core.ArtifactRevision {
	t.Helper()
	aid, err := core.NewArtifactID(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	rid, err := core.NewArtifactRevisionID(revisionID)
	if err != nil {
		t.Fatal(err)
	}
	origin, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	integrity, err := core.NewIntegrityIdentity(core.IntegrityMechanismCryptographicDigest, "abc123", core.IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	rev, err := core.NewArtifactRevision(aid, rid, origin, prov, integrity)
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

// --- Record ------------------------------------------------------------

func TestNewRecordValid(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	decRef := mustTestDecisionRef(t, "dec-1")
	r, err := NewRecord(artifact, decRef)
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision() != decRef {
		t.Errorf("Decision() = %v, want %v", r.Decision(), decRef)
	}
}

func TestNewRecordWrongArtifactTypeRejected(t *testing.T) {
	aid, err := core.NewArtifactID("DR-1")
	if err != nil {
		t.Fatal(err)
	}
	wrongType := core.NewArtifactType(mustVocab("requirement"))
	artifact, err := core.NewArtifact(aid, wrongType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1")); !errors.Is(err, ErrArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactTypeMismatch)
	}
}

func TestNewRecordZeroArtifactRejected(t *testing.T) {
	if _, err := NewRecord(core.Artifact{}, mustTestDecisionRef(t, "dec-1")); !errors.Is(err, ErrInvalidDecisionRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecord)
	}
}

func TestNewRecordZeroDecisionRefRejected(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	if _, err := NewRecord(artifact, core.DecisionRef{}); !errors.Is(err, ErrInvalidDecisionRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecord)
	}
}

func TestRecordAccessors(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	decRef := mustTestDecisionRef(t, "dec-1")
	r, err := NewRecord(artifact, decRef)
	if err != nil {
		t.Fatal(err)
	}
	if r.ID() != artifact.ID() {
		t.Errorf("ID() = %v, want %v", r.ID(), artifact.ID())
	}
	if r.Core().ID() != artifact.ID() || r.Core().Type() != artifact.Type() {
		t.Error("Core() mismatch")
	}
}

func TestRecordNestedJSONShape(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	decRef := mustTestDecisionRef(t, "dec-1")
	r, err := NewRecord(artifact, decRef)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["core"]; !ok {
		t.Error(`"core" key missing`)
	}
	if _, ok := raw["decision"]; !ok {
		t.Error(`"decision" key missing`)
	}
	var decoded Record
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Decision() != decRef {
		t.Errorf("round trip mismatch: got %v, want %v", decoded.Decision(), decRef)
	}
}

func TestRecordZeroMarshalRejected(t *testing.T) {
	var r Record
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidDecisionRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecord)
	}
}

func TestRecordUnmarshalFailurePreservesReceiver(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	decRef := mustTestDecisionRef(t, "dec-1")
	original, err := NewRecord(artifact, decRef)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"core":{}}`), &receiver); err == nil {
		t.Fatal("missing decision accepted, want error")
	}
	if receiver.Decision() != original.Decision() {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- RecordContent -------------------------------------------------------

func TestRecordContentZeroValid(t *testing.T) {
	c := NewRecordContent()
	if !c.IsZero() {
		t.Error("NewRecordContent() IsZero() = false")
	}
}

func TestRecordContentZeroMarshalsToEmptyObject(t *testing.T) {
	c := NewRecordContent()
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("Marshal(zero) = %s, want {}", data)
	}
}

func TestRecordContentSupplementalEvidence(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-2", "REV-2")
	c := NewRecordContent()
	withEv, err := c.WithSupplementalEvidence(ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(withEv.SupplementalEvidence()) != 1 {
		t.Errorf("SupplementalEvidence() len = %d, want 1", len(withEv.SupplementalEvidence()))
	}
	if len(c.SupplementalEvidence()) != 0 {
		t.Error("WithSupplementalEvidence mutated the original receiver")
	}
}

func TestRecordContentZeroEvidenceRefRejected(t *testing.T) {
	c := NewRecordContent()
	if _, err := c.WithSupplementalEvidence(core.EvidenceArtifactRevisionRef{}); !errors.Is(err, ErrInvalidDecisionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecordRevision)
	}
}

func TestRecordContentDefensiveCopies(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-2", "REV-2")
	evidence := []core.EvidenceArtifactRevisionRef{ev}
	c := NewRecordContent()
	c, err := c.WithSupplementalEvidence(evidence...)
	if err != nil {
		t.Fatal(err)
	}
	evidence[0] = core.EvidenceArtifactRevisionRef{}
	if c.SupplementalEvidence()[0].IsZero() {
		t.Error("WithSupplementalEvidence did not defensively copy input")
	}
	got := c.SupplementalEvidence()
	got[0] = core.EvidenceArtifactRevisionRef{}
	if c.SupplementalEvidence()[0].IsZero() {
		t.Error("SupplementalEvidence() did not defensively copy on return")
	}
}

func TestRecordContentExtension(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	c := NewRecordContent().WithExtension(ext)
	if c.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestRecordContentExplicitNullRejected(t *testing.T) {
	var c RecordContent
	if err := json.Unmarshal([]byte(`{"supplemental_evidence":null}`), &c); !errors.Is(err, ErrInvalidDecisionRecordRevision) {
		t.Errorf("null supplemental_evidence: error = %v, want %v", err, ErrInvalidDecisionRecordRevision)
	}
	if err := json.Unmarshal([]byte(`{"extension":null}`), &c); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestRecordContentUnknownFieldIgnored(t *testing.T) {
	var c RecordContent
	if err := json.Unmarshal([]byte(`{"unknown_field":123}`), &c); err != nil {
		t.Fatal(err)
	}
}

func TestRecordContentUnmarshalFailurePreservesReceiver(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-2", "REV-2")
	original, err := NewRecordContent().WithSupplementalEvidence(ev)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"supplemental_evidence":null}`), &receiver); err == nil {
		t.Fatal("null supplemental_evidence accepted, want error")
	}
	if len(receiver.SupplementalEvidence()) != 1 {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- RecordRevision ------------------------------------------------------

func TestNewRecordRevisionValid(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	content := NewRecordContent()
	rr, err := NewRecordRevision(record, rev, content)
	if err != nil {
		t.Fatal(err)
	}
	if rr.IsZero() {
		t.Error("valid RecordRevision reports IsZero() = true")
	}
}

func TestNewRecordRevisionZeroRecordRejected(t *testing.T) {
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	if _, err := NewRecordRevision(Record{}, rev, NewRecordContent()); !errors.Is(err, ErrInvalidDecisionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecordRevision)
	}
}

func TestNewRecordRevisionZeroArtifactRevisionRejected(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewRecordRevision(record, core.ArtifactRevision{}, NewRecordContent()); !errors.Is(err, ErrInvalidDecisionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecordRevision)
	}
}

func TestNewRecordRevisionMatchingArtifactIDAccepted(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	if _, err := NewRecordRevision(record, rev, NewRecordContent()); err != nil {
		t.Fatal(err)
	}
}

func TestNewRecordRevisionMismatchingArtifactIDRejected(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-OTHER", "REV-1")
	if _, err := NewRecordRevision(record, rev, NewRecordContent()); !errors.Is(err, ErrArtifactIDMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactIDMismatch)
	}
}

func TestNewRecordRevisionZeroContentAccepted(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	rr, err := NewRecordRevision(record, rev, RecordContent{})
	if err != nil {
		t.Fatal(err)
	}
	if !rr.Content().IsZero() {
		t.Error("Content() not zero")
	}
}

func TestRecordRevisionNestedJSONShapeWithEmptyContent(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	rr, err := NewRecordRevision(record, rev, RecordContent{})
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
	content, ok := raw["content"]
	if !ok {
		t.Fatal(`"content" key missing`)
	}
	if string(content) != "{}" {
		t.Errorf("content = %s, want {}", content)
	}
}

func TestRecordRevisionRefCorrectness(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	rr, err := NewRecordRevision(record, rev, NewRecordContent())
	if err != nil {
		t.Fatal(err)
	}
	ref, err := rr.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != rev.ArtifactID() || ref.RevisionID() != rev.RevisionID() {
		t.Errorf("Ref() = %v, want artifact/revision matching %v/%v", ref, rev.ArtifactID(), rev.RevisionID())
	}
}

func TestRecordRevisionZeroMarshalRejected(t *testing.T) {
	var rr RecordRevision
	if _, err := json.Marshal(rr); !errors.Is(err, ErrInvalidDecisionRecordRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRecordRevision)
	}
}

func TestRecordRevisionUnmarshalFailurePreservesReceiver(t *testing.T) {
	artifact := mustDecisionRecordArtifact(t, "DR-1")
	record, err := NewRecord(artifact, mustTestDecisionRef(t, "dec-1"))
	if err != nil {
		t.Fatal(err)
	}
	rev := mustArtifactRevision(t, "DR-1", "REV-1")
	original, err := NewRecordRevision(record, rev, NewRecordContent())
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"core":{}}`), &receiver); err == nil {
		t.Fatal("zero core revision accepted, want error")
	}
	if receiver.Core().ArtifactID() != original.Core().ArtifactID() || receiver.Core().RevisionID() != original.Core().RevisionID() {
		t.Error("failed Unmarshal changed receiver")
	}
}
