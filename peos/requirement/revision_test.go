package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewRevisionValidMatchingArtifactID(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, req.ID(), "REV-1")
	content := mustContent(t)

	rev, err := NewRevision(req, coreRev, content)
	if err != nil {
		t.Fatal(err)
	}
	if rev.IsZero() {
		t.Error("valid Revision reports IsZero() = true")
	}
	if rev.Core().ArtifactID() != req.ID() {
		t.Errorf("Core().ArtifactID() = %v, want %v", rev.Core().ArtifactID(), req.ID())
	}
}

func TestNewRevisionMismatchedArtifactID(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, mustArtifactID(t, "REQ-2"), "REV-1")
	content := mustContent(t)

	if _, err := NewRevision(req, coreRev, content); !errors.Is(err, ErrRequirementArtifactIDMismatch) {
		t.Errorf("error = %v, want %v", err, ErrRequirementArtifactIDMismatch)
	}
}

func TestNewRevisionZeroRequirement(t *testing.T) {
	coreRev := mustCoreRevision(t, mustArtifactID(t, "REQ-1"), "REV-1")
	content := mustContent(t)
	if _, err := NewRevision(Requirement{}, coreRev, content); !errors.Is(err, ErrInvalidRequirementRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRevision)
	}
}

func TestNewRevisionZeroCoreRevision(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	content := mustContent(t)
	if _, err := NewRevision(req, core.ArtifactRevision{}, content); !errors.Is(err, ErrInvalidRequirementRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRevision)
	}
}

func TestNewRevisionZeroContent(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, req.ID(), "REV-1")
	if _, err := NewRevision(req, coreRev, Content{}); !errors.Is(err, ErrMissingRequirementContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRequirementContent)
	}
}

func TestNewRevisionZeroCoreRepresentationsAccepted(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, req.ID(), "REV-1")
	content := mustContent(t)

	rev, err := NewRevision(req, coreRev, content)
	if err != nil {
		t.Fatal(err)
	}
	if reps := rev.Core().Representations(); reps != nil {
		t.Errorf("Core().Representations() = %v, want nil", reps)
	}
}

func TestRevisionJSONNestedShape(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, req.ID(), "REV-1")
	content := mustContent(t)
	rev, err := NewRevision(req, coreRev, content)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(rev)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["core"]; !present {
		t.Error("\"core\" field missing from Revision's Marshal output")
	}
	if _, present := raw["content"]; !present {
		t.Error("\"content\" field missing from Revision's Marshal output")
	}
	// The nested shape must not be flattened: core-level fields such as
	// artifact_id must not appear at the top level.
	if _, present := raw["artifact_id"]; present {
		t.Error("Revision JSON is flattened: \"artifact_id\" present at top level")
	}

	var nestedCore map[string]json.RawMessage
	if err := json.Unmarshal(raw["core"], &nestedCore); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"artifact_id", "revision_id", "origin", "provenance", "integrity"} {
		if _, present := nestedCore[field]; !present {
			t.Errorf("nested core field %q missing", field)
		}
	}
}

func TestRevisionJSONRoundTrip(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, req.ID(), "REV-1")
	content := mustContent(t)
	original, err := NewRevision(req, coreRev, content)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Revision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Core().ArtifactID() != original.Core().ArtifactID() || decoded.Core().RevisionID() != original.Core().RevisionID() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded.Core(), original.Core())
	}
	if len(decoded.Content().Statements()) != 1 || decoded.Content().Statements()[0].Text() != original.Content().Statements()[0].Text() {
		t.Errorf("round trip Content() mismatch: got %v", decoded.Content().Statements())
	}
}

func TestRevisionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	req := mustRequirement(t, "REQ-1")
	coreRev := mustCoreRevision(t, req.ID(), "REV-1")
	content := mustContent(t)
	original, err := NewRevision(req, coreRev, content)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"core":{"artifact_id":"REQ-1"}}`), &receiver); err == nil {
		t.Fatal("malformed Revision JSON accepted, want error")
	}
	if receiver.Core().ArtifactID() != original.Core().ArtifactID() || receiver.Core().RevisionID() != original.Core().RevisionID() {
		t.Errorf("failed Unmarshal changed Core(): got %+v, want %+v", receiver.Core(), original.Core())
	}
	if len(receiver.Content().Statements()) != len(original.Content().Statements()) {
		t.Errorf("failed Unmarshal changed Content(): got %v, want %v", receiver.Content(), original.Content())
	}
}

func TestRevisionZeroValue(t *testing.T) {
	var rev Revision
	if !rev.IsZero() {
		t.Error("zero-value Revision.IsZero() = false, want true")
	}
}

func TestRevisionZeroValueMarshalFails(t *testing.T) {
	var rev Revision
	if _, err := json.Marshal(rev); !errors.Is(err, ErrInvalidRequirementRevision) {
		t.Errorf("Marshal(zero Revision): error = %v, want %v", err, ErrInvalidRequirementRevision)
	}
}
