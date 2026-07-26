package requirement

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustRequirementRef(t *testing.T, artifactID string) core.RequirementRef {
	t.Helper()
	ref, err := core.NewRequirementRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustRequirementParticipantFromRequirement(t *testing.T, artifactID string) RequirementParticipant {
	t.Helper()
	p, err := NewRequirementParticipantFromRequirement(mustRequirementRef(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func mustRequirementParticipantFromRevision(t *testing.T, artifactID, revisionID string) RequirementParticipant {
	t.Helper()
	p, err := NewRequirementParticipantFromRequirementRevision(mustRequirementRevisionRef(t, artifactID, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNewRequirementParticipantFromRequirementValid(t *testing.T) {
	p := mustRequirementParticipantFromRequirement(t, "REQ-1")
	if p.IsZero() {
		t.Error("valid RequirementParticipant IsZero() = true")
	}
	if !p.IsRequirement() {
		t.Error("IsRequirement() = false")
	}
	if p.IsRequirementRevision() {
		t.Error("IsRequirementRevision() = true")
	}
	if p.Kind() != "requirement" {
		t.Errorf("Kind() = %q, want %q", p.Kind(), "requirement")
	}
}

func TestNewRequirementParticipantFromRequirementZeroRejected(t *testing.T) {
	_, err := NewRequirementParticipantFromRequirement(core.RequirementRef{})
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewRequirementParticipantFromRequirementRevisionValid(t *testing.T) {
	p := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	if p.IsZero() {
		t.Error("valid RequirementParticipant IsZero() = true")
	}
	if !p.IsRequirementRevision() {
		t.Error("IsRequirementRevision() = false")
	}
	if p.IsRequirement() {
		t.Error("IsRequirement() = true")
	}
	if p.Kind() != "requirement_revision" {
		t.Errorf("Kind() = %q, want %q", p.Kind(), "requirement_revision")
	}
}

func TestNewRequirementParticipantFromRequirementRevisionZeroRejected(t *testing.T) {
	_, err := NewRequirementParticipantFromRequirementRevision(core.RequirementArtifactRevisionRef{})
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestRequirementParticipantAsRequirementWrongArm(t *testing.T) {
	p := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	if _, ok := p.AsRequirement(); ok {
		t.Error("AsRequirement() ok = true on a revision-level participant")
	}
}

func TestRequirementParticipantAsRequirementRevisionWrongArm(t *testing.T) {
	p := mustRequirementParticipantFromRequirement(t, "REQ-1")
	if _, ok := p.AsRequirementRevision(); ok {
		t.Error("AsRequirementRevision() ok = true on an identity-level participant")
	}
}

func TestRequirementParticipantAsRequirementCorrectArm(t *testing.T) {
	ref := mustRequirementRef(t, "REQ-1")
	p, err := NewRequirementParticipantFromRequirement(ref)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := p.AsRequirement()
	if !ok || got != ref {
		t.Errorf("AsRequirement() = (%v,%v), want (%v,true)", got, ok, ref)
	}
}

func TestRequirementParticipantAsRequirementRevisionCorrectArm(t *testing.T) {
	ref := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	p, err := NewRequirementParticipantFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := p.AsRequirementRevision()
	if !ok || got != ref {
		t.Errorf("AsRequirementRevision() = (%v,%v), want (%v,true)", got, ok, ref)
	}
}

// TestRequirementParticipantRequirementIDBothLevels proves RequirementID
// returns the owning Requirement identity at both participant levels.
func TestRequirementParticipantRequirementIDBothLevels(t *testing.T) {
	identityLevel := mustRequirementParticipantFromRequirement(t, "REQ-1")
	revisionLevel := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	if identityLevel.RequirementID() != revisionLevel.RequirementID() {
		t.Errorf("RequirementID() mismatch: %v vs %v", identityLevel.RequirementID(), revisionLevel.RequirementID())
	}
	other := mustRequirementParticipantFromRequirement(t, "REQ-2")
	if identityLevel.RequirementID() == other.RequirementID() {
		t.Error("RequirementID() equal for distinct Requirement identities")
	}
}

func TestRequirementParticipantIsZero(t *testing.T) {
	var p RequirementParticipant
	if !p.IsZero() {
		t.Error("zero RequirementParticipant IsZero() = false")
	}
	if mustRequirementParticipantFromRequirement(t, "REQ-1").IsZero() {
		t.Error("valid RequirementParticipant IsZero() = true")
	}
}

// TestRequirementParticipantComparable proves RequirementParticipant is
// comparable with ==, giving distinctness checks for free.
func TestRequirementParticipantComparable(t *testing.T) {
	a := mustRequirementParticipantFromRequirement(t, "REQ-1")
	b := mustRequirementParticipantFromRequirement(t, "REQ-1")
	c := mustRequirementParticipantFromRequirement(t, "REQ-2")
	if a != b {
		t.Error("equal participants compared unequal")
	}
	if a == c {
		t.Error("distinct participants compared equal")
	}
	revA := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	revB := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	if revA != revB {
		t.Error("equal revision-level participants compared unequal")
	}
	if a == revA {
		t.Error("identity-level and revision-level participants over the same Requirement compared equal")
	}
}

// TestRequirementParticipantNoJSONMethods is a structural absence audit
// proving RequirementParticipant never reaches the wire directly (see
// this type's own doc comment in participant.go).
func TestRequirementParticipantNoJSONMethods(t *testing.T) {
	typ := reflect.TypeOf(RequirementParticipant{})
	for _, name := range []string{"MarshalJSON", "UnmarshalJSON"} {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("RequirementParticipant unexpectedly implements %s", name)
		}
		if _, ok := reflect.PointerTo(typ).MethodByName(name); ok {
			t.Errorf("*RequirementParticipant unexpectedly implements %s", name)
		}
	}
}
