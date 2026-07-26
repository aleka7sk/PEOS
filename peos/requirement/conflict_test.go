package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustConflict(t *testing.T) Conflict {
	t.Helper()
	c, err := NewConflict(
		mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"),
		mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"),
		mustProvenance(t),
		mustScope(t, "product-x", "/services/*"),
		"REQ-1 and REQ-2 mandate incompatible retention periods.",
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func fullConflict(t *testing.T) Conflict {
	t.Helper()
	c := mustConflict(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return c.WithExtension(ext)
}

func conflictParticipantJSON(kind, artifactID, revisionID string) string {
	if kind == "requirement" {
		return `{"kind":"requirement","ref":{"artifact_id":"` + artifactID + `"}}`
	}
	return `{"kind":"requirement_revision","ref":{"artifact_id":"` + artifactID + `","revision_id":"` + revisionID + `"}}`
}

func conflictPayload(t *testing.T, aKind, aArtifact, aRevision, bKind, bArtifact, bRevision, relationType, nature string, includeScope bool) string {
	t.Helper()
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	a := conflictParticipantJSON(aKind, aArtifact, aRevision)
	b := conflictParticipantJSON(bKind, bArtifact, bRevision)
	scopeField := ""
	if includeScope {
		scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
		if err != nil {
			t.Fatal(err)
		}
		scopeField = `,"scope":` + string(scope)
	}
	return `{"relation":{"relation_type":"` + relationType + `","source":` + a + `,"target":` + b + `,"provenance":` + string(prov) + scopeField + `},"nature":"` + nature + `"}`
}

// --- NewConflict -----------------------------------------------------------

func TestNewConflictValid(t *testing.T) {
	c := mustConflict(t)
	if c.IsZero() {
		t.Error("valid Conflict IsZero() = true")
	}
}

func TestNewConflictZeroParticipantARejected(t *testing.T) {
	_, err := NewConflict(RequirementParticipant{}, mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewConflictZeroParticipantBRejected(t *testing.T) {
	_, err := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), RequirementParticipant{}, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewConflictSelfConflictRejected proves PEOS-005 §22/§22.1's
// distinct-participants requirement -- the deliberate opposite of
// Dependency's self-dependency acceptance (see
// TestNewDependencySelfDependencyAccepted).
func TestNewConflictSelfConflictRejected(t *testing.T) {
	same := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	_, err := NewConflict(same, same, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewConflictSameRequirementDifferentRevisionsAccepted proves no
// Requirement-identity distinctness rule applies to Conflict: only
// participant-shape distinctness is required.
func TestNewConflictSameRequirementDifferentRevisionsAccepted(t *testing.T) {
	a := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-2")
	if _, err := NewConflict(a, b, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature"); err != nil {
		t.Errorf("same-Requirement different-revision Conflict rejected: %v", err)
	}
}

// TestNewConflictAllParticipantLevelCombinations proves all four
// (identity, revision) combinations are conforming (PEOS-005 §22.1).
func TestNewConflictAllParticipantLevelCombinations(t *testing.T) {
	identityA := mustRequirementParticipantFromRequirement(t, "REQ-1")
	identityB := mustRequirementParticipantFromRequirement(t, "REQ-2")
	revisionA := mustRequirementParticipantFromRevision(t, "REQ-3", "REV-1")
	revisionB := mustRequirementParticipantFromRevision(t, "REQ-4", "REV-1")

	cases := []struct {
		name string
		a, b RequirementParticipant
	}{
		{"identity-identity", identityA, identityB},
		{"identity-revision", identityA, revisionB},
		{"revision-identity", revisionA, identityB},
		{"revision-revision", revisionA, revisionB},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conflict, err := NewConflict(c.a, c.b, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
			if err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			if conflict.ParticipantA() != c.a {
				t.Errorf("%s: ParticipantA() = %v, want %v", c.name, conflict.ParticipantA(), c.a)
			}
			if conflict.ParticipantB() != c.b {
				t.Errorf("%s: ParticipantB() = %v, want %v", c.name, conflict.ParticipantB(), c.b)
			}
		})
	}
}

func TestNewConflictZeroProvenanceRejected(t *testing.T) {
	_, err := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), core.Provenance{}, mustScope(t, "product-x", "/x"), "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewConflictZeroScopeRejected proves scope is mandatory for
// Conflict, unlike Dependency.
func TestNewConflictZeroScopeRejected(t *testing.T) {
	_, err := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), core.Scope{}, "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewConflictEmptyNatureRejected(t *testing.T) {
	_, err := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), "")
	if !errors.Is(err, ErrInvalidConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflict)
	}
}

func TestNewConflictWhitespaceOnlyNatureRejected(t *testing.T) {
	_, err := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), "   ")
	if !errors.Is(err, ErrInvalidConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflict)
	}
}

func TestNewConflictNatureStoredTrimmed(t *testing.T) {
	c, err := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), "  padded  ")
	if err != nil {
		t.Fatal(err)
	}
	if c.Nature() != "padded" {
		t.Errorf("Nature() = %q, want %q", c.Nature(), "padded")
	}
}

func TestNewConflictRelationTypeAlwaysConflict(t *testing.T) {
	c := mustConflict(t)
	if c.Relation().RelationType() != core.RelationTypeConflict {
		t.Errorf("RelationType() = %v, want %v", c.Relation().RelationType(), core.RelationTypeConflict)
	}
}

// --- Non-canonicalization --------------------------------------------------

// TestNewConflictReverseOrderNotCanonicalized proves both (A,B) and (B,A)
// construct successfully and are not canonicalized into a single
// representation (PEOS-005 §22.1: ordering "SHALL NOT imply priority,
// authority, preference, or resolution direction").
func TestNewConflictReverseOrderNotCanonicalized(t *testing.T) {
	a := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	scope := mustScope(t, "product-x", "/x")

	forward, err := NewConflict(a, b, mustProvenance(t), scope, "nature")
	if err != nil {
		t.Fatalf("forward Conflict rejected: %v", err)
	}
	reverse, err := NewConflict(b, a, mustProvenance(t), scope, "nature")
	if err != nil {
		t.Fatalf("reverse Conflict rejected: %v", err)
	}
	if forward.ParticipantA() != a || forward.ParticipantB() != b {
		t.Error("forward Conflict did not preserve supplied order")
	}
	if reverse.ParticipantA() != b || reverse.ParticipantB() != a {
		t.Error("reverse Conflict did not preserve supplied order")
	}
	if forward.ParticipantA() == reverse.ParticipantA() {
		t.Error("forward and reverse Conflicts unexpectedly share ParticipantA -- ordering was canonicalized")
	}
}

// TestConflictSamePairDifferentNatureAccepted proves one participant pair
// MAY participate in multiple Conflict relationships with different
// nature (PEOS-005 §22.1).
func TestConflictSamePairDifferentNatureAccepted(t *testing.T) {
	a := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	scope := mustScope(t, "product-x", "/x")
	c1, err := NewConflict(a, b, mustProvenance(t), scope, "retention period conflict")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := NewConflict(a, b, mustProvenance(t), scope, "access control conflict")
	if err != nil {
		t.Fatal(err)
	}
	if c1.Nature() == c2.Nature() {
		t.Error("expected distinct nature values")
	}
}

// --- With* ---------------------------------------------------------------

func TestConflictWithExtension(t *testing.T) {
	c := mustConflict(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := c.WithExtension(ext)
	if !c.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestConflictWithoutExtension(t *testing.T) {
	c := mustConflict(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := c.WithExtension(ext)
	cleared := withExt.WithoutExtension()
	if !cleared.Extension().IsZero() {
		t.Error("Extension() non-zero after WithoutExtension")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithoutExtension mutated the original receiver")
	}
}

func TestConflictWithMethodsAreImmutable(t *testing.T) {
	c := mustConflict(t)
	original := c
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

// --- Accessors -----------------------------------------------------------

func TestConflictAccessors(t *testing.T) {
	a := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	prov := mustProvenance(t)
	scope := mustScope(t, "product-x", "/services/*")
	c, err := NewConflict(a, b, prov, scope, "nature")
	if err != nil {
		t.Fatal(err)
	}
	if c.ParticipantA() != a {
		t.Errorf("ParticipantA() = %v, want %v", c.ParticipantA(), a)
	}
	if c.ParticipantB() != b {
		t.Errorf("ParticipantB() = %v, want %v", c.ParticipantB(), b)
	}
	if c.Nature() != "nature" {
		t.Errorf("Nature() = %q, want %q", c.Nature(), "nature")
	}
	if c.Scope() != scope {
		t.Errorf("Scope() = %v, want %v", c.Scope(), scope)
	}
	gotActor, _ := c.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if gotActor != wantActor {
		t.Errorf("Provenance().Actor() = %v, want %v", gotActor, wantActor)
	}
}

// TestConflictScopeNeverAbsent proves Scope() always returns a non-zero
// value for a validly constructed Conflict.
func TestConflictScopeNeverAbsent(t *testing.T) {
	c := mustConflict(t)
	if c.Scope().IsZero() {
		t.Error("Scope() returned zero value on a valid Conflict")
	}
}

func TestConflictIsZero(t *testing.T) {
	var c Conflict
	if !c.IsZero() {
		t.Error("zero Conflict IsZero() = false")
	}
	if mustConflict(t).IsZero() {
		t.Error("valid Conflict IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestConflictJSONLiteralWireKeys(t *testing.T) {
	c := fullConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation", "nature"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
	if len(raw) != 2 {
		t.Errorf("Conflict wire form has %d top-level keys, want exactly 2 (relation, nature): %v", len(raw), raw)
	}
}

func TestConflictJSONMinimumOmitsExtension(t *testing.T) {
	c := mustConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var relRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["relation"], &relRaw); err != nil {
		t.Fatal(err)
	}
	if _, present := relRaw["extension"]; present {
		t.Error("extension present despite not being set")
	}
	if _, present := relRaw["scope"]; !present {
		t.Error(`"scope" must always be present -- Conflict's scope is mandatory`)
	}
}

func TestConflictJSONRoundTrip(t *testing.T) {
	c := fullConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Conflict
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ParticipantA() != c.ParticipantA() || decoded.ParticipantB() != c.ParticipantB() {
		t.Error("participant round trip mismatch")
	}
	if decoded.Nature() != c.Nature() {
		t.Error("Nature mismatch")
	}
	if decoded.Scope() != c.Scope() {
		t.Error("Scope mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestConflictJSONMinimumRoundTrip(t *testing.T) {
	c := mustConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Conflict
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ParticipantA() != c.ParticipantA() {
		t.Error("round trip mismatch")
	}
}

// TestConflictJSONMixedLevelRoundTrip proves a Conflict whose two
// participants sit at different levels round-trips correctly.
func TestConflictJSONMixedLevelRoundTrip(t *testing.T) {
	a := mustRequirementParticipantFromRequirement(t, "REQ-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	c, err := NewConflict(a, b, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Conflict
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.ParticipantA().IsRequirement() {
		t.Error("ParticipantA() lost identity-level kind after round trip")
	}
	if !decoded.ParticipantB().IsRequirementRevision() {
		t.Error("ParticipantB() lost revision-level kind after round trip")
	}
	if decoded.ParticipantA() != a || decoded.ParticipantB() != b {
		t.Error("mixed-level participant round trip mismatch")
	}
}

func TestConflictZeroMarshalRejected(t *testing.T) {
	var c Conflict
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestConflictJSONUnknownFieldIgnored(t *testing.T) {
	c := mustConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["unknown_field"] = json.RawMessage(`123`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Conflict
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestConflictUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullConflict(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Nature() != original.Nature() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// --- Decode-only validation ------------------------------------------------

func TestConflictJSONWrongRelationTypeRejected(t *testing.T) {
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:dependency", "nature", true)
	var c Conflict
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestConflictJSONWrongParticipantKindRejected proves a subject of a
// non-Requirement kind (Decision) is rejected, not silently accepted.
func TestConflictJSONWrongParticipantKindRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	a := `{"kind":"decision","ref":{"decision_id":"DEC-1"}}`
	b := conflictParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:conflict","source":` + a + `,"target":` + b + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"nature":"nature"}`
	var c Conflict
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestConflictJSONSameParticipantRejected(t *testing.T) {
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:conflict", "nature", true)
	var c Conflict
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestConflictJSONMissingNatureRejected(t *testing.T) {
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:conflict", "", true)
	var c Conflict
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflict)
	}
}

// TestConflictJSONMissingScopeRejected proves scope is mandatory on
// decode, not merely on construction.
func TestConflictJSONMissingScopeRejected(t *testing.T) {
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:conflict", "nature", false)
	var c Conflict
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestConflictJSONExplicitNullRelationRejected(t *testing.T) {
	var c Conflict
	if err := json.Unmarshal([]byte(`{"relation":null,"nature":"n"}`), &c); err == nil {
		t.Error("null relation accepted, want error")
	}
}

func TestConflictJSONExplicitNullNatureRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	a := conflictParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	b := conflictParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:conflict","source":` + a + `,"target":` + b + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"nature":null}`
	var c Conflict
	if err := json.Unmarshal([]byte(payload), &c); err == nil {
		t.Error("null nature accepted, want error")
	}
}

// TestConflictJSONReverseOrderPreserved proves decode does not reorder or
// canonicalize a stored source/target pair.
func TestConflictJSONReverseOrderPreserved(t *testing.T) {
	a := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	reverse, err := NewConflict(b, a, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(reverse)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Conflict
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ParticipantA() != b || decoded.ParticipantB() != a {
		t.Error("decode reordered the supplied source/target pair")
	}
}

// --- Constructor / Unmarshal equivalence ------------------------------------

func TestConflictConstructorUnmarshalEquivalenceSameParticipant(t *testing.T) {
	same := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	sameParticipant, err := NewRequirementParticipantFromRequirementRevision(same)
	if err != nil {
		t.Fatal(err)
	}
	_, ctorErr := NewConflict(sameParticipant, sameParticipant, mustProvenance(t), mustScope(t, "product-x", "/x"), "nature")
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:conflict", "nature", true)
	var c Conflict
	jsonErr := json.Unmarshal([]byte(payload), &c)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

func TestConflictConstructorUnmarshalEquivalenceMissingScope(t *testing.T) {
	_, ctorErr := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), core.Scope{}, "nature")
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:conflict", "nature", false)
	var c Conflict
	jsonErr := json.Unmarshal([]byte(payload), &c)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

func TestConflictConstructorUnmarshalEquivalenceEmptyNature(t *testing.T) {
	_, ctorErr := NewConflict(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), "")
	payload := conflictPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:conflict", "", true)
	var c Conflict
	jsonErr := json.Unmarshal([]byte(payload), &c)
	if !errors.Is(ctorErr, ErrInvalidConflict) || !errors.Is(jsonErr, ErrInvalidConflict) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidConflict)
	}
}

// --- Nested sentinel preservation --------------------------------------

func TestConflictNestedSentinelPreserved(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	malformedSource := `{"kind":"requirement_revision","ref":{"artifact_id":"","revision_id":"REV-1"}}`
	target := conflictParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:conflict","source":` + malformedSource + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"nature":"nature"}`
	var c Conflict
	err = json.Unmarshal([]byte(payload), &c)
	if err == nil {
		t.Fatal("malformed nested source ref accepted, want error")
	}
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidRequirementRelation)
	}
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("error = %v, want also wrapping %v", err, core.ErrEmptyIdentity)
	}
}

// --- Absence audit (PEOS-005 §22.1) -----------------------------------------

// TestConflictNoGroupOrGraphEntity is a structural absence audit proving
// this package introduces no ConflictSet, RelationshipGroup, or other
// PEOS entity representing a group of Conflict relationships: the fully
// populated wire form of a single Conflict has exactly two top-level
// keys.
func TestConflictNoGroupOrGraphEntity(t *testing.T) {
	c := fullConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != 2 {
		t.Errorf("Conflict wire form has %d top-level keys, want exactly 2 (relation, nature): %v", len(raw), raw)
	}
}

func TestConflictNoIdentityField(t *testing.T) {
	c := fullConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["id"]; present {
		t.Error(`unexpected "id" key present in Conflict wire form`)
	}
}
