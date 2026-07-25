package requirement

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- shared test helpers, used across every file in this package's test suite ---

func mustVocab(t *testing.T, namespace, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
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

func mustRequirementArtifact(t *testing.T, id string) core.Artifact {
	t.Helper()
	a, err := core.NewArtifact(mustArtifactID(t, id), ArtifactTypeRequirement)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustRequirement(t *testing.T, id string) Requirement {
	t.Helper()
	r, err := New(mustRequirementArtifact(t, id))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func mustOrigin(t *testing.T) core.Origin {
	t.Helper()
	o, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustProvenance(t *testing.T) core.Provenance {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	actor, err := core.NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewProvenance().WithActor(actor).WithRecordedAt(ts)
}

func mustIntegrity(t *testing.T) core.IntegrityIdentity {
	t.Helper()
	i, err := core.NewIntegrityIdentity(core.IntegrityMechanismCryptographicDigest, "sha256:abc123", core.IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func mustCoreRevision(t *testing.T, artifactID core.ArtifactID, revisionID string) core.ArtifactRevision {
	t.Helper()
	rev, err := core.NewArtifactRevision(artifactID, mustArtifactRevisionID(t, revisionID), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

func mustSubject(t *testing.T, artifactID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewArtifactRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	sub, err := core.EngineeringSubjectRefFromArtifact(ref)
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func mustStatement(t *testing.T, text string) Statement {
	t.Helper()
	s, err := NewStatement(text)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustContent(t *testing.T) Content {
	t.Helper()
	c, err := NewContent(
		[]Statement{mustStatement(t, "The service shall retain audit records.")},
		[]core.EngineeringSubjectRef{mustSubject(t, "ART-1")},
		SubjectCombinationIndependent,
		NewUnrestrictedApplicability(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// --- Requirement ---------------------------------------------------------

func TestNewRequirementValid(t *testing.T) {
	artifact := mustRequirementArtifact(t, "REQ-1")
	r, err := New(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if r.IsZero() {
		t.Error("valid Requirement reports IsZero() = true")
	}
	if r.ID() != mustArtifactID(t, "REQ-1") {
		t.Errorf("ID() = %v, want %v", r.ID(), mustArtifactID(t, "REQ-1"))
	}
	// core.Artifact is not comparable with == (it holds a []ArtifactRole
	// slice and an Extension map), so it is checked via accessors.
	if r.Core().ID() != artifact.ID() || r.Core().Type() != artifact.Type() {
		t.Errorf("Core() = %+v, want %+v", r.Core(), artifact)
	}
}

func TestNewRequirementWrongArtifactType(t *testing.T) {
	otherType := core.NewArtifactType(mustVocab(t, "peos", "decision"))
	artifact, err := core.NewArtifact(mustArtifactID(t, "REQ-1"), otherType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := New(artifact); !errors.Is(err, ErrRequirementArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrRequirementArtifactTypeMismatch)
	}
}

func TestNewRequirementZeroArtifact(t *testing.T) {
	if _, err := New(core.Artifact{}); !errors.Is(err, ErrInvalidRequirement) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirement)
	}
}

func TestRequirementJSONRoundTrip(t *testing.T) {
	original := mustRequirement(t, "REQ-1")
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["artifact_type"]; !present {
		t.Error("artifact_type missing from Requirement's Marshal output")
	}
	var decoded Requirement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != original.ID() {
		t.Errorf("round trip ID() = %v, want %v", decoded.ID(), original.ID())
	}
}

func TestRequirementJSONRejectsWrongArtifactType(t *testing.T) {
	otherType := core.NewArtifactType(mustVocab(t, "peos", "decision"))
	artifact, err := core.NewArtifact(mustArtifactID(t, "REQ-1"), otherType)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Requirement
	if err := json.Unmarshal(data, &decoded); !errors.Is(err, ErrRequirementArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrRequirementArtifactTypeMismatch)
	}
}

func TestRequirementUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := mustRequirement(t, "REQ-1")
	receiver := original
	otherType := core.NewArtifactType(mustVocab(t, "peos", "decision"))
	artifact, err := core.NewArtifact(mustArtifactID(t, "REQ-2"), otherType)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &receiver); err == nil {
		t.Fatal("malformed Requirement JSON accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Errorf("failed Unmarshal changed ID(): got %v, want %v", receiver.ID(), original.ID())
	}
}

func TestRequirementZeroValue(t *testing.T) {
	var r Requirement
	if !r.IsZero() {
		t.Error("zero-value Requirement.IsZero() = false, want true")
	}
}

func TestRequirementZeroValueMarshalFails(t *testing.T) {
	var r Requirement
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRequirement) {
		t.Errorf("Marshal(zero Requirement): error = %v, want %v", err, ErrInvalidRequirement)
	}
}

func TestArtifactTypeRequirementValue(t *testing.T) {
	if ArtifactTypeRequirement.IsZero() {
		t.Error("ArtifactTypeRequirement reports IsZero() = true")
	}
	if got, want := ArtifactTypeRequirement.Value().String(), "peos:requirement"; got != want {
		t.Errorf("ArtifactTypeRequirement.Value().String() = %q, want %q", got, want)
	}
}
