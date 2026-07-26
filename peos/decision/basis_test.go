package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustEvidenceRef(t *testing.T, artifactID, revisionID string) core.EvidenceArtifactRevisionRef {
	t.Helper()
	aid, err := core.NewArtifactID(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	rid, err := core.NewArtifactRevisionID(revisionID)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewEvidenceArtifactRevisionRef(aid, rid)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestNewBasisValidEvidence(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Evidence()) != 1 {
		t.Errorf("Evidence() len = %d, want 1", len(b.Evidence()))
	}
}

func TestNewBasisEmptyEvidenceRejected(t *testing.T) {
	if _, err := NewBasis(nil); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewBasisZeroEvidenceRefRejected(t *testing.T) {
	if _, err := NewBasis([]core.EvidenceArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisDefensiveCopies(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	input := []core.EvidenceArtifactRevisionRef{ev}
	b, err := NewBasis(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = core.EvidenceArtifactRevisionRef{}
	if b.Evidence()[0].IsZero() {
		t.Error("NewBasis did not defensively copy input")
	}
	got := b.Evidence()
	got[0] = core.EvidenceArtifactRevisionRef{}
	if b.Evidence()[0].IsZero() {
		t.Error("Evidence() did not defensively copy on return")
	}
}

func TestBasisExtensionBehavior(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := b.WithExtension(ext)
	if !b.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestBasisJSONRoundTrip(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Basis
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Evidence()) != 1 {
		t.Errorf("round trip mismatch: got %+v", decoded)
	}
}

func TestBasisExplicitNullRejected(t *testing.T) {
	var b Basis
	if err := json.Unmarshal([]byte(`{"evidence":null}`), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("null evidence: error = %v, want %v", err, ErrInvalidBasis)
	}
	valid := `{"evidence":[{"artifact_id":"ART-1","revision_id":"REV-1"}],"extension":null}`
	if err := json.Unmarshal([]byte(valid), &b); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestBasisZeroMarshalRejected(t *testing.T) {
	var b Basis
	if _, err := json.Marshal(b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisUnmarshalFailurePreservesReceiver(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	original, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if len(receiver.Evidence()) != 1 {
		t.Error("failed Unmarshal changed receiver")
	}
}
