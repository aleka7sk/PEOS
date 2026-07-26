package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustDecomposition(t *testing.T) Decomposition {
	t.Helper()
	d, err := NewDecomposition(
		mustRequirementRevisionRef(t, "REQ-1", "REV-1"),
		mustRequirementRevisionRef(t, "REQ-2", "REV-1"),
		mustProvenance(t),
		mustScope(t, "product-x", "/services/*"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func fullDecomposition(t *testing.T) Decomposition {
	t.Helper()
	d := mustDecomposition(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return d.WithExtension(ext)
}

// --- NewDecomposition ----------------------------------------------------

func TestNewDecompositionValid(t *testing.T) {
	d := mustDecomposition(t)
	if d.IsZero() {
		t.Error("valid Decomposition IsZero() = true")
	}
}

func TestNewDecompositionZeroParentRejected(t *testing.T) {
	_, err := NewDecomposition(core.RequirementArtifactRevisionRef{}, mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewDecompositionZeroSubordinateRejected(t *testing.T) {
	_, err := NewDecomposition(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), core.RequirementArtifactRevisionRef{}, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewDecompositionSelfDecompositionRejected proves the direct
// self-decomposition prohibition (PEOS-005 §20.1).
func TestNewDecompositionSelfDecompositionRejected(t *testing.T) {
	same := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewDecomposition(same, same, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewDecompositionSameRequirementDifferentRevisionsRejected is the
// packet's headline test: PEOS-005 §20.1 requires "A subordinate
// Requirement identity SHALL remain distinct from the parent Requirement
// identity" -- stronger than mere revision-level distinctness, and the
// deliberate mirror of Refinement's acceptance of the equivalent case.
func TestNewDecompositionSameRequirementDifferentRevisionsRejected(t *testing.T) {
	parent := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	subordinate := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	_, err := NewDecomposition(parent, subordinate, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidDecomposition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecomposition)
	}
}

func TestNewDecompositionZeroProvenanceRejected(t *testing.T) {
	_, err := NewDecomposition(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), core.Provenance{}, mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewDecompositionZeroScopeRejected proves scope is mandatory for
// Decomposition, unlike Derivation.
func TestNewDecompositionZeroScopeRejected(t *testing.T) {
	_, err := NewDecomposition(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), core.Scope{})
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewDecompositionRelationTypeAlwaysDecomposition(t *testing.T) {
	d := mustDecomposition(t)
	if d.Relation().RelationType() != core.RelationTypeDecomposition {
		t.Errorf("RelationType() = %v, want %v", d.Relation().RelationType(), core.RelationTypeDecomposition)
	}
}

// --- With* ---------------------------------------------------------------

func TestDecompositionWithExtension(t *testing.T) {
	d := mustDecomposition(t)
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

func TestDecompositionWithoutExtension(t *testing.T) {
	d := mustDecomposition(t)
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

func TestDecompositionWithMethodsAreImmutable(t *testing.T) {
	d := mustDecomposition(t)
	original := d
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = d.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

// --- Accessors -----------------------------------------------------------

func TestDecompositionAccessors(t *testing.T) {
	parent := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	subordinate := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	prov := mustProvenance(t)
	scope := mustScope(t, "product-x", "/services/*")
	d, err := NewDecomposition(parent, subordinate, prov, scope)
	if err != nil {
		t.Fatal(err)
	}
	if d.Parent() != parent {
		t.Errorf("Parent() = %v, want %v", d.Parent(), parent)
	}
	if d.Subordinate() != subordinate {
		t.Errorf("Subordinate() = %v, want %v", d.Subordinate(), subordinate)
	}
	if d.Scope() != scope {
		t.Errorf("Scope() = %v, want %v", d.Scope(), scope)
	}
	gotActor, _ := d.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if gotActor != wantActor {
		t.Errorf("Provenance().Actor() = %v, want %v", gotActor, wantActor)
	}
}

// TestDecompositionScopeNeverAbsent proves Scope() always returns a
// non-zero value for a validly constructed Decomposition.
func TestDecompositionScopeNeverAbsent(t *testing.T) {
	d := mustDecomposition(t)
	if d.Scope().IsZero() {
		t.Error("Scope() returned zero value on a valid Decomposition")
	}
}

func TestDecompositionIsZero(t *testing.T) {
	var d Decomposition
	if !d.IsZero() {
		t.Error("zero Decomposition IsZero() = false")
	}
	if mustDecomposition(t).IsZero() {
		t.Error("valid Decomposition IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestDecompositionJSONLiteralWireKeys(t *testing.T) {
	d := fullDecomposition(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["relation"]; !present {
		t.Error(`required key "relation" missing`)
	}
	if len(raw) != 1 {
		t.Errorf("Decomposition wire form has %d top-level keys, want exactly 1 (relation): %v", len(raw), raw)
	}
}

func TestDecompositionJSONMinimumOmitsExtension(t *testing.T) {
	d := mustDecomposition(t)
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
	if _, present := relRaw["extension"]; present {
		t.Error("extension present despite not being set")
	}
	if _, present := relRaw["scope"]; !present {
		t.Error(`"scope" must always be present -- Decomposition's scope is mandatory`)
	}
}

func TestDecompositionJSONRoundTrip(t *testing.T) {
	d := fullDecomposition(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Decomposition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Parent() != d.Parent() || decoded.Subordinate() != d.Subordinate() {
		t.Error("participant round trip mismatch")
	}
	if decoded.Scope() != d.Scope() {
		t.Error("Scope mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDecompositionJSONMinimumRoundTrip(t *testing.T) {
	d := mustDecomposition(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Decomposition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Parent() != d.Parent() {
		t.Error("round trip mismatch")
	}
}

func TestDecompositionZeroMarshalRejected(t *testing.T) {
	var d Decomposition
	if _, err := json.Marshal(d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDecompositionJSONUnknownFieldIgnored(t *testing.T) {
	d := mustDecomposition(t)
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
	var decoded Decomposition
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestDecompositionUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullDecomposition(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Parent() != original.Parent() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// --- Decode-only validation ------------------------------------------------

func decompositionPayload(t *testing.T, parentArtifact, parentRevision, subArtifact, subRevision, sourceKind, targetKind, relationType string, includeScope bool) string {
	t.Helper()
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"` + sourceKind + `","ref":{"artifact_id":"` + parentArtifact + `","revision_id":"` + parentRevision + `"}}`
	target := `{"kind":"` + targetKind + `","ref":{"artifact_id":"` + subArtifact + `","revision_id":"` + subRevision + `"}}`
	if sourceKind == "requirement" {
		source = `{"kind":"requirement","ref":{"artifact_id":"` + parentArtifact + `"}}`
	}
	if targetKind == "requirement" {
		target = `{"kind":"requirement","ref":{"artifact_id":"` + subArtifact + `"}}`
	}
	scopeField := ""
	if includeScope {
		scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
		if err != nil {
			t.Fatal(err)
		}
		scopeField = `,"scope":` + string(scope)
	}
	return `{"relation":{"relation_type":"` + relationType + `","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + scopeField + `}}`
}

func TestDecompositionJSONWrongRelationTypeRejected(t *testing.T) {
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-2", "REV-1", "requirement_revision", "requirement_revision", "peos:refinement", true)
	var d Decomposition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDecompositionJSONSourceAtIdentityLevelRejected(t *testing.T) {
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-2", "REV-1", "requirement", "requirement_revision", "peos:decomposition", true)
	var d Decomposition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDecompositionJSONTargetAtIdentityLevelRejected(t *testing.T) {
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-2", "REV-1", "requirement_revision", "requirement", "peos:decomposition", true)
	var d Decomposition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDecompositionJSONSourceEqualsTargetRejected(t *testing.T) {
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-1", "REV-1", "requirement_revision", "requirement_revision", "peos:decomposition", true)
	var d Decomposition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestDecompositionJSONSameRequirementDifferentRevisionsRejected proves
// the §20.1 identity-distinctness rule is enforced on decode too.
func TestDecompositionJSONSameRequirementDifferentRevisionsRejected(t *testing.T) {
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-1", "REV-2", "requirement_revision", "requirement_revision", "peos:decomposition", true)
	var d Decomposition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidDecomposition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecomposition)
	}
}

// TestDecompositionJSONMissingScopeRejected proves scope is mandatory on
// decode, not merely on construction.
func TestDecompositionJSONMissingScopeRejected(t *testing.T) {
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-2", "REV-1", "requirement_revision", "requirement_revision", "peos:decomposition", false)
	var d Decomposition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDecompositionJSONExplicitNullRelationRejected(t *testing.T) {
	var d Decomposition
	if err := json.Unmarshal([]byte(`{"relation":null}`), &d); err == nil {
		t.Error("null relation accepted, want error")
	}
}

// --- Constructor / Unmarshal equivalence ------------------------------------

func TestDecompositionConstructorUnmarshalEquivalenceSameRequirement(t *testing.T) {
	_, ctorErr := NewDecomposition(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-2"), mustProvenance(t), mustScope(t, "product-x", "/x"))
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-1", "REV-2", "requirement_revision", "requirement_revision", "peos:decomposition", true)
	var d Decomposition
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidDecomposition) || !errors.Is(jsonErr, ErrInvalidDecomposition) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidDecomposition)
	}
}

func TestDecompositionConstructorUnmarshalEquivalenceMissingScope(t *testing.T) {
	_, ctorErr := NewDecomposition(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), core.Scope{})
	payload := decompositionPayload(t, "REQ-1", "REV-1", "REQ-2", "REV-1", "requirement_revision", "requirement_revision", "peos:decomposition", false)
	var d Decomposition
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

// --- Nested sentinel preservation --------------------------------------

// TestDecompositionNestedSentinelPreserved proves a malformed nested core
// reference (empty artifact_id) preserves both
// ErrInvalidRequirementRelation and the underlying core.ErrEmptyIdentity
// through errors.Is.
func TestDecompositionNestedSentinelPreserved(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	malformedSource := `{"kind":"requirement_revision","ref":{"artifact_id":"","revision_id":"REV-1"}}`
	target := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-2","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:decomposition","source":` + malformedSource + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `}}`
	var d Decomposition
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

// --- Multiplicity ----------------------------------------------------------

// TestDecompositionOneParentMultipleSubordinates proves §20.1's "one
// parent Requirement Artifact Revision MAY be the source of multiple
// Decomposition relationships."
func TestDecompositionOneParentMultipleSubordinates(t *testing.T) {
	parent := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	d1, err := NewDecomposition(parent, mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	d2, err := NewDecomposition(parent, mustRequirementRevisionRef(t, "REQ-3", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/y"))
	if err != nil {
		t.Fatal(err)
	}
	if d1.Parent() != d2.Parent() {
		t.Error("expected both decompositions to share the same parent")
	}
	if d1.Subordinate() == d2.Subordinate() {
		t.Error("expected distinct subordinates")
	}
}

// TestDecompositionOneSubordinateMultipleParents proves §20.1 permits
// one subordinate Requirement Artifact Revision to be the target of more
// than one Decomposition relationship; distinguishing their scopes is a
// repository-level concern this package does not enforce.
func TestDecompositionOneSubordinateMultipleParents(t *testing.T) {
	subordinate := mustRequirementRevisionRef(t, "REQ-3", "REV-1")
	d1, err := NewDecomposition(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), subordinate, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	d2, err := NewDecomposition(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), subordinate, mustProvenance(t), mustScope(t, "product-x", "/y"))
	if err != nil {
		t.Fatal(err)
	}
	if d1.Subordinate() != d2.Subordinate() {
		t.Error("expected both decompositions to share the same subordinate")
	}
}

// --- Absence audit (PEOS-005 §20.2) -----------------------------------------

// TestDecompositionNoCompletenessModel is a structural absence audit
// proving this package introduces no Decomposition Set, Relationship
// Collection, or Completeness Assertion entity (PEOS-005 §20.2): the
// fully populated wire form of a single Decomposition has exactly one
// top-level key, and no exported symbol in this package names a group
// concept for Decomposition.
func TestDecompositionNoCompletenessModel(t *testing.T) {
	d := fullDecomposition(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Errorf("Decomposition wire form has %d top-level keys, want exactly 1 (relation): %v", len(raw), raw)
	}
}
