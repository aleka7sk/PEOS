package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustDependency(t *testing.T) Dependency {
	t.Helper()
	d, err := NewDependency(
		mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"),
		mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"),
		mustProvenance(t),
		"REQ-1 relies on REQ-2's audit-trail capability being available.",
	)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func fullDependency(t *testing.T) Dependency {
	t.Helper()
	d := mustDependency(t)
	d, err := d.WithScope(mustScope(t, "product-x", "/services/*"))
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return d.WithExtension(ext)
}

func dependencyParticipantJSON(kind, artifactID, revisionID string) string {
	if kind == "requirement" {
		return `{"kind":"requirement","ref":{"artifact_id":"` + artifactID + `"}}`
	}
	return `{"kind":"requirement_revision","ref":{"artifact_id":"` + artifactID + `","revision_id":"` + revisionID + `"}}`
}

func dependencyPayload(t *testing.T, sourceKind, sourceArtifact, sourceRevision, targetKind, targetArtifact, targetRevision, relationType, nature string, includeScope bool) string {
	t.Helper()
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := dependencyParticipantJSON(sourceKind, sourceArtifact, sourceRevision)
	target := dependencyParticipantJSON(targetKind, targetArtifact, targetRevision)
	scopeField := ""
	if includeScope {
		scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
		if err != nil {
			t.Fatal(err)
		}
		scopeField = `,"scope":` + string(scope)
	}
	return `{"relation":{"relation_type":"` + relationType + `","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + scopeField + `},"nature":"` + nature + `"}`
}

// --- NewDependency -------------------------------------------------------

func TestNewDependencyValid(t *testing.T) {
	d := mustDependency(t)
	if d.IsZero() {
		t.Error("valid Dependency IsZero() = true")
	}
}

func TestNewDependencyZeroDependentRejected(t *testing.T) {
	_, err := NewDependency(RequirementParticipant{}, mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewDependencyZeroDependsOnRejected(t *testing.T) {
	_, err := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), RequirementParticipant{}, mustProvenance(t), "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewDependencySelfDependencyAccepted proves PEOS-005 §21.1's explicit
// permission of Dependency cycles, of which a self-dependency is the
// degenerate case -- the deliberate opposite of Conflict's rejection of
// the equivalent case (see TestNewConflictSelfConflictRejected).
func TestNewDependencySelfDependencyAccepted(t *testing.T) {
	same := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	d, err := NewDependency(same, same, mustProvenance(t), "a requirement may rely on its own continued validity")
	if err != nil {
		t.Fatalf("self-dependency rejected: %v", err)
	}
	if d.IsZero() {
		t.Error("valid self-Dependency IsZero() = true")
	}
}

// TestNewDependencySameRequirementDifferentRevisionsAccepted proves no
// Requirement-identity distinctness rule applies to Dependency, unlike
// Derivation and Decomposition.
func TestNewDependencySameRequirementDifferentRevisionsAccepted(t *testing.T) {
	dependent := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	dependsOn := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-2")
	if _, err := NewDependency(dependent, dependsOn, mustProvenance(t), "nature"); err != nil {
		t.Errorf("same-Requirement different-revision Dependency rejected: %v", err)
	}
}

// TestNewDependencyIdentityAndRevisionLevelSameRequirementAccepted proves
// an identity-level and a revision-level participant naming the same
// Requirement MAY both participate in one Dependency.
func TestNewDependencyIdentityAndRevisionLevelSameRequirementAccepted(t *testing.T) {
	dependent := mustRequirementParticipantFromRequirement(t, "REQ-1")
	dependsOn := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	if _, err := NewDependency(dependent, dependsOn, mustProvenance(t), "nature"); err != nil {
		t.Errorf("mixed-level same-Requirement Dependency rejected: %v", err)
	}
}

// TestNewDependencyTwoNodeCycleAccepted proves a 2-node Dependency cycle
// (A depends on B, and B depends on A) is representable as two
// independent, both-valid Dependency values (PEOS-005 §21.1).
func TestNewDependencyTwoNodeCycleAccepted(t *testing.T) {
	a := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	b := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	if _, err := NewDependency(a, b, mustProvenance(t), "a on b"); err != nil {
		t.Errorf("A-depends-on-B rejected: %v", err)
	}
	if _, err := NewDependency(b, a, mustProvenance(t), "b on a"); err != nil {
		t.Errorf("B-depends-on-A rejected: %v", err)
	}
}

// TestNewDependencyAllParticipantLevelCombinations proves all four
// (identity, revision) combinations are conforming (PEOS-005 §21.1).
func TestNewDependencyAllParticipantLevelCombinations(t *testing.T) {
	identityA := mustRequirementParticipantFromRequirement(t, "REQ-1")
	identityB := mustRequirementParticipantFromRequirement(t, "REQ-2")
	revisionA := mustRequirementParticipantFromRevision(t, "REQ-3", "REV-1")
	revisionB := mustRequirementParticipantFromRevision(t, "REQ-4", "REV-1")

	cases := []struct {
		name                 string
		dependent, dependsOn RequirementParticipant
	}{
		{"identity-identity", identityA, identityB},
		{"identity-revision", identityA, revisionB},
		{"revision-identity", revisionA, identityB},
		{"revision-revision", revisionA, revisionB},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, err := NewDependency(c.dependent, c.dependsOn, mustProvenance(t), "nature")
			if err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			if d.Dependent() != c.dependent {
				t.Errorf("%s: Dependent() = %v, want %v", c.name, d.Dependent(), c.dependent)
			}
			if d.DependsOn() != c.dependsOn {
				t.Errorf("%s: DependsOn() = %v, want %v", c.name, d.DependsOn(), c.dependsOn)
			}
		})
	}
}

func TestNewDependencyZeroProvenanceRejected(t *testing.T) {
	_, err := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), core.Provenance{}, "nature")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewDependencyEmptyNatureRejected(t *testing.T) {
	_, err := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), "")
	if !errors.Is(err, ErrInvalidDependency) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDependency)
	}
}

func TestNewDependencyWhitespaceOnlyNatureRejected(t *testing.T) {
	_, err := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), "   ")
	if !errors.Is(err, ErrInvalidDependency) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDependency)
	}
}

func TestNewDependencyNatureStoredTrimmed(t *testing.T) {
	d, err := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), "  padded  ")
	if err != nil {
		t.Fatal(err)
	}
	if d.Nature() != "padded" {
		t.Errorf("Nature() = %q, want %q", d.Nature(), "padded")
	}
}

func TestNewDependencyRelationTypeAlwaysDependency(t *testing.T) {
	d := mustDependency(t)
	if d.Relation().RelationType() != core.RelationTypeDependency {
		t.Errorf("RelationType() = %v, want %v", d.Relation().RelationType(), core.RelationTypeDependency)
	}
}

// --- With* ---------------------------------------------------------------

func TestDependencyWithScope(t *testing.T) {
	d := mustDependency(t)
	if _, ok := d.Scope(); ok {
		t.Error("Scope() ok = true before WithScope")
	}
	scope := mustScope(t, "product-x", "/services/*")
	withScope, err := d.WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withScope.Scope()
	if !ok || got != scope {
		t.Errorf("Scope() = (%v,%v)", got, ok)
	}
	if _, ok := d.Scope(); ok {
		t.Error("WithScope mutated the original receiver")
	}
}

func TestDependencyZeroScopeRejected(t *testing.T) {
	d := mustDependency(t)
	if _, err := d.WithScope(core.Scope{}); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestDependencyWithoutScope(t *testing.T) {
	d := mustDependency(t)
	scope := mustScope(t, "product-x", "/services/*")
	withScope, err := d.WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	cleared := withScope.WithoutScope()
	if _, ok := cleared.Scope(); ok {
		t.Error("Scope() ok = true after WithoutScope")
	}
	if _, ok := withScope.Scope(); !ok {
		t.Error("WithoutScope mutated the original receiver")
	}
}

func TestDependencyWithExtension(t *testing.T) {
	d := mustDependency(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := d.WithExtension(ext)
	if !d.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestDependencyWithoutExtension(t *testing.T) {
	d := mustDependency(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := d.WithExtension(ext)
	cleared := withExt.WithoutExtension()
	if !cleared.Extension().IsZero() {
		t.Error("Extension() non-zero after WithoutExtension")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithoutExtension mutated the original receiver")
	}
}

func TestDependencyWithMethodsAreImmutable(t *testing.T) {
	d := mustDependency(t)
	original := d
	scope := mustScope(t, "product-x", "/services/*")
	if _, err := d.WithScope(scope); err != nil {
		t.Fatal(err)
	}
	if d.Nature() != original.Nature() {
		t.Error("WithScope mutated d")
	}
	if _, ok := d.Scope(); ok {
		t.Error("WithScope mutated d")
	}
}

// --- Accessors -----------------------------------------------------------

func TestDependencyAccessors(t *testing.T) {
	dependent := mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1")
	dependsOn := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	prov := mustProvenance(t)
	d, err := NewDependency(dependent, dependsOn, prov, "nature")
	if err != nil {
		t.Fatal(err)
	}
	if d.Dependent() != dependent {
		t.Errorf("Dependent() = %v, want %v", d.Dependent(), dependent)
	}
	if d.DependsOn() != dependsOn {
		t.Errorf("DependsOn() = %v, want %v", d.DependsOn(), dependsOn)
	}
	if d.Nature() != "nature" {
		t.Errorf("Nature() = %q, want %q", d.Nature(), "nature")
	}
	gotActor, _ := d.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if gotActor != wantActor {
		t.Errorf("Provenance().Actor() = %v, want %v", gotActor, wantActor)
	}
}

func TestDependencyIsZero(t *testing.T) {
	var d Dependency
	if !d.IsZero() {
		t.Error("zero Dependency IsZero() = false")
	}
	if mustDependency(t).IsZero() {
		t.Error("valid Dependency IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestDependencyJSONLiteralWireKeys(t *testing.T) {
	d := fullDependency(t)
	data, err := json.Marshal(d)
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
		t.Errorf("Dependency wire form has %d top-level keys, want exactly 2 (relation, nature): %v", len(raw), raw)
	}
}

func TestDependencyJSONMinimumOmitsOptionalFields(t *testing.T) {
	d := mustDependency(t)
	data, err := json.Marshal(d)
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
	for _, key := range []string{"scope", "extension"} {
		if _, present := relRaw[key]; present {
			t.Errorf("optional nested key %q present despite not being set", key)
		}
	}
}

func TestDependencyJSONRoundTrip(t *testing.T) {
	d := fullDependency(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Dependency
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Dependent() != d.Dependent() || decoded.DependsOn() != d.DependsOn() {
		t.Error("participant round trip mismatch")
	}
	if decoded.Nature() != d.Nature() {
		t.Error("Nature mismatch")
	}
	if _, ok := decoded.Scope(); !ok {
		t.Error("Scope absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDependencyJSONMinimumRoundTrip(t *testing.T) {
	d := mustDependency(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Dependency
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Dependent() != d.Dependent() {
		t.Error("round trip mismatch")
	}
}

// TestDependencyJSONMixedLevelRoundTrip proves a Dependency whose two
// participants sit at different levels round-trips correctly.
func TestDependencyJSONMixedLevelRoundTrip(t *testing.T) {
	dependent := mustRequirementParticipantFromRequirement(t, "REQ-1")
	dependsOn := mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1")
	d, err := NewDependency(dependent, dependsOn, mustProvenance(t), "nature")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Dependency
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Dependent().IsRequirement() {
		t.Error("Dependent() lost identity-level kind after round trip")
	}
	if !decoded.DependsOn().IsRequirementRevision() {
		t.Error("DependsOn() lost revision-level kind after round trip")
	}
	if decoded.Dependent() != dependent || decoded.DependsOn() != dependsOn {
		t.Error("mixed-level participant round trip mismatch")
	}
}

func TestDependencyZeroMarshalRejected(t *testing.T) {
	var d Dependency
	if _, err := json.Marshal(d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDependencyJSONUnknownFieldIgnored(t *testing.T) {
	d := mustDependency(t)
	data, err := json.Marshal(d)
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
	var decoded Dependency
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestDependencyUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullDependency(t)
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

func TestDependencyJSONWrongRelationTypeRejected(t *testing.T) {
	payload := dependencyPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:conflict", "nature", false)
	var d Dependency
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestDependencyJSONWrongParticipantKindRejected proves a subject of a
// non-Requirement kind (Decision) is rejected, not silently accepted.
func TestDependencyJSONWrongParticipantKindRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"decision","ref":{"decision_id":"DEC-1"}}`
	target := dependencyParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:dependency","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `},"nature":"nature"}`
	var d Dependency
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDependencyJSONMissingNatureRejected(t *testing.T) {
	payload := dependencyPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:dependency", "", false)
	var d Dependency
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidDependency) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDependency)
	}
}

func TestDependencyJSONExplicitNullRelationRejected(t *testing.T) {
	var d Dependency
	if err := json.Unmarshal([]byte(`{"relation":null,"nature":"n"}`), &d); err == nil {
		t.Error("null relation accepted, want error")
	}
}

func TestDependencyJSONExplicitNullNatureRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := dependencyParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	target := dependencyParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:dependency","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `},"nature":null}`
	var d Dependency
	if err := json.Unmarshal([]byte(payload), &d); err == nil {
		t.Error("null nature accepted, want error")
	}
}

// --- Constructor / Unmarshal equivalence ------------------------------------

func TestDependencyConstructorUnmarshalEquivalenceEmptyNature(t *testing.T) {
	_, ctorErr := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), mustProvenance(t), "")
	payload := dependencyPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-2", "REV-1", "peos:dependency", "", false)
	var d Dependency
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidDependency) || !errors.Is(jsonErr, ErrInvalidDependency) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidDependency)
	}
}

func TestDependencyConstructorUnmarshalEquivalenceZeroProvenance(t *testing.T) {
	_, ctorErr := NewDependency(mustRequirementParticipantFromRevision(t, "REQ-1", "REV-1"), mustRequirementParticipantFromRevision(t, "REQ-2", "REV-1"), core.Provenance{}, "nature")
	source := dependencyParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	target := dependencyParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:dependency","source":` + source + `,"target":` + target + `},"nature":"nature"}`
	var d Dependency
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

// --- Nested sentinel preservation --------------------------------------

func TestDependencyNestedSentinelPreserved(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	malformedSource := `{"kind":"requirement_revision","ref":{"artifact_id":"","revision_id":"REV-1"}}`
	target := dependencyParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	payload := `{"relation":{"relation_type":"peos:dependency","source":` + malformedSource + `,"target":` + target + `,"provenance":` + string(prov) + `},"nature":"nature"}`
	var d Dependency
	err = json.Unmarshal([]byte(payload), &d)
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

// --- Absence audit -------------------------------------------------------

func TestDependencyNoIdentityField(t *testing.T) {
	d := fullDependency(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["id"]; present {
		t.Error(`unexpected "id" key present in Dependency wire form`)
	}
}
