package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustRefinement(t *testing.T) Refinement {
	t.Helper()
	r, err := NewRefinement(
		mustRequirementRevisionRef(t, "REQ-1", "REV-1"),
		mustRequirementRevisionRef(t, "REQ-2", "REV-1"),
		mustProvenance(t),
		mustScope(t, "product-x", "/services/*"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func fullRefinement(t *testing.T) Refinement {
	t.Helper()
	r := mustRefinement(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return r.WithExtension(ext)
}

// --- NewRefinement -------------------------------------------------------

func TestNewRefinementValid(t *testing.T) {
	r := mustRefinement(t)
	if r.IsZero() {
		t.Error("valid Refinement IsZero() = true")
	}
}

func TestNewRefinementZeroRefinedRejected(t *testing.T) {
	_, err := NewRefinement(core.RequirementArtifactRevisionRef{}, mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewRefinementZeroRefiningRejected(t *testing.T) {
	_, err := NewRefinement(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), core.RequirementArtifactRevisionRef{}, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewRefinementSelfRefinementRejected proves the direct
// self-refinement prohibition (PEOS-005 §19.1).
func TestNewRefinementSelfRefinementRejected(t *testing.T) {
	same := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewRefinement(same, same, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewRefinementZeroProvenanceRejected(t *testing.T) {
	_, err := NewRefinement(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), core.Provenance{}, mustScope(t, "product-x", "/x"))
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewRefinementZeroScopeRejected proves scope is mandatory for
// Refinement, unlike Derivation.
func TestNewRefinementZeroScopeRejected(t *testing.T) {
	_, err := NewRefinement(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), core.Scope{})
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewRefinementSameRequirementDifferentRevisionsAccepted proves §19
// imposes no Requirement-identity distinctness rule: a later Revision of
// a Requirement MAY refine an earlier Revision of the very same
// Requirement -- the deliberate mirror of Decomposition's rejection of
// the equivalent case.
func TestNewRefinementSameRequirementDifferentRevisionsAccepted(t *testing.T) {
	refined := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	refining := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	r, err := NewRefinement(refined, refining, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatalf("same-Requirement refinement rejected, want accepted: %v", err)
	}
	if r.Refined() != refined || r.Refining() != refining {
		t.Error("participants not preserved")
	}
}

func TestNewRefinementRelationTypeAlwaysRefinement(t *testing.T) {
	r := mustRefinement(t)
	if r.Relation().RelationType() != core.RelationTypeRefinement {
		t.Errorf("RelationType() = %v, want %v", r.Relation().RelationType(), core.RelationTypeRefinement)
	}
}

// --- With* ---------------------------------------------------------------

func TestRefinementWithExtension(t *testing.T) {
	r := mustRefinement(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := r.WithExtension(ext)
	if !r.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestRefinementWithoutExtension(t *testing.T) {
	r := mustRefinement(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := r.WithExtension(ext)
	cleared := withExt.WithoutExtension()
	if !cleared.Extension().IsZero() {
		t.Error("Extension() non-zero after WithoutExtension")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithoutExtension mutated the original receiver")
	}
}

func TestRefinementWithMethodsAreImmutable(t *testing.T) {
	r := mustRefinement(t)
	original := r
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = r.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

// --- Accessors -----------------------------------------------------------

func TestRefinementAccessors(t *testing.T) {
	refined := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	refining := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	prov := mustProvenance(t)
	scope := mustScope(t, "product-x", "/services/*")
	r, err := NewRefinement(refined, refining, prov, scope)
	if err != nil {
		t.Fatal(err)
	}
	if r.Refined() != refined {
		t.Errorf("Refined() = %v, want %v", r.Refined(), refined)
	}
	if r.Refining() != refining {
		t.Errorf("Refining() = %v, want %v", r.Refining(), refining)
	}
	if r.Scope() != scope {
		t.Errorf("Scope() = %v, want %v", r.Scope(), scope)
	}
	gotActor, _ := r.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if gotActor != wantActor {
		t.Errorf("Provenance().Actor() = %v, want %v", gotActor, wantActor)
	}
}

// TestRefinementScopeNeverAbsent proves Scope() always returns a
// non-zero value for a validly constructed Refinement -- there is no
// presence bool because presence is guaranteed.
func TestRefinementScopeNeverAbsent(t *testing.T) {
	r := mustRefinement(t)
	if r.Scope().IsZero() {
		t.Error("Scope() returned zero value on a valid Refinement")
	}
}

func TestRefinementIsZero(t *testing.T) {
	var r Refinement
	if !r.IsZero() {
		t.Error("zero Refinement IsZero() = false")
	}
	if mustRefinement(t).IsZero() {
		t.Error("valid Refinement IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestRefinementJSONLiteralWireKeys(t *testing.T) {
	r := fullRefinement(t)
	data, err := json.Marshal(r)
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
		t.Errorf("Refinement wire form has %d top-level keys, want exactly 1 (relation): %v", len(raw), raw)
	}
	var relRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["relation"], &relRaw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation_type", "source", "target", "provenance", "scope", "extension"} {
		if _, present := relRaw[key]; !present {
			t.Errorf("required nested key %q missing", key)
		}
	}
}

func TestRefinementJSONMinimumOmitsExtension(t *testing.T) {
	r := mustRefinement(t)
	data, err := json.Marshal(r)
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
		t.Error(`"scope" must always be present -- Refinement's scope is mandatory`)
	}
}

func TestRefinementJSONRoundTrip(t *testing.T) {
	r := fullRefinement(t)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Refinement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Refined() != r.Refined() || decoded.Refining() != r.Refining() {
		t.Error("participant round trip mismatch")
	}
	if decoded.Scope() != r.Scope() {
		t.Error("Scope mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestRefinementJSONMinimumRoundTrip(t *testing.T) {
	r := mustRefinement(t)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Refinement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Refined() != r.Refined() {
		t.Error("round trip mismatch")
	}
}

func TestRefinementZeroMarshalRejected(t *testing.T) {
	var r Refinement
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestRefinementJSONUnknownFieldIgnored(t *testing.T) {
	r := mustRefinement(t)
	data, err := json.Marshal(r)
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
	var decoded Refinement
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestRefinementUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullRefinement(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Refined() != original.Refined() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// --- Decode-only validation ------------------------------------------------

func refinementPayload(t *testing.T, sourceKind, targetKind, relationType string, includeScope bool) string {
	t.Helper()
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"` + sourceKind + `","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	target := `{"kind":"` + targetKind + `","ref":{"artifact_id":"REQ-2","revision_id":"REV-1"}}`
	if sourceKind == "requirement" {
		source = `{"kind":"requirement","ref":{"artifact_id":"REQ-1"}}`
	}
	if targetKind == "requirement" {
		target = `{"kind":"requirement","ref":{"artifact_id":"REQ-2"}}`
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

func TestRefinementJSONWrongRelationTypeRejected(t *testing.T) {
	payload := refinementPayload(t, "requirement_revision", "requirement_revision", "peos:derivation", true)
	var r Refinement
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestRefinementJSONSourceAtIdentityLevelRejected(t *testing.T) {
	payload := refinementPayload(t, "requirement", "requirement_revision", "peos:refinement", true)
	var r Refinement
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestRefinementJSONTargetAtIdentityLevelRejected(t *testing.T) {
	payload := refinementPayload(t, "requirement_revision", "requirement", "peos:refinement", true)
	var r Refinement
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestRefinementJSONSourceEqualsTargetRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	same := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:refinement","source":` + same + `,"target":` + same + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `}}`
	var r Refinement
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestRefinementJSONMissingScopeRejected proves scope is mandatory on
// decode, not merely on construction.
func TestRefinementJSONMissingScopeRejected(t *testing.T) {
	payload := refinementPayload(t, "requirement_revision", "requirement_revision", "peos:refinement", false)
	var r Refinement
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestRefinementJSONExplicitNullRelationRejected(t *testing.T) {
	var r Refinement
	if err := json.Unmarshal([]byte(`{"relation":null}`), &r); err == nil {
		t.Error("null relation accepted, want error")
	}
}

// --- Constructor / Unmarshal equivalence ------------------------------------

func TestRefinementConstructorUnmarshalEquivalenceSelfRefinement(t *testing.T) {
	_, ctorErr := NewRefinement(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"))
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	same := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:refinement","source":` + same + `,"target":` + same + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `}}`
	var r Refinement
	jsonErr := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

func TestRefinementConstructorUnmarshalEquivalenceMissingScope(t *testing.T) {
	_, ctorErr := NewRefinement(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), core.Scope{})
	payload := refinementPayload(t, "requirement_revision", "requirement_revision", "peos:refinement", false)
	var r Refinement
	jsonErr := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

// --- Nested sentinel preservation --------------------------------------

// TestRefinementNestedSentinelPreserved proves a malformed nested core
// reference (empty artifact_id) preserves both
// ErrInvalidRequirementRelation and the underlying core.ErrEmptyIdentity
// through errors.Is.
func TestRefinementNestedSentinelPreserved(t *testing.T) {
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
	payload := `{"relation":{"relation_type":"peos:refinement","source":` + malformedSource + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `}}`
	var r Refinement
	err = json.Unmarshal([]byte(payload), &r)
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

// --- Cross-type: Derivation and Refinement coexist --------------------------

// TestDerivationAndRefinementCoexistIndependently proves a Derivation
// and a Refinement over the same participant pair are independent
// values, neither implied by the other (PEOS-005 §19, non-conforming
// pattern §36.15).
func TestDerivationAndRefinementCoexistIndependently(t *testing.T) {
	source := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	target := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	d, err := NewDerivation(source, target, mustProvenance(t), "derived for this reason")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRefinement(source, target, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	if d.Source() != r.Refined() || d.Target() != r.Refining() {
		t.Fatal("test setup: expected same participant pair")
	}
	if d.Relation().RelationType() == r.Relation().RelationType() {
		t.Error("Derivation and Refinement must not collapse into the same relation type")
	}
}

// --- Multiple Refinements sharing one refining target -----------------------

// TestMultipleRefinementsShareRefiningTarget proves §19.1's "multiple
// Refinement relationships MAY share the same refining target
// Requirement Artifact Revision."
func TestMultipleRefinementsShareRefiningTarget(t *testing.T) {
	refining := mustRequirementRevisionRef(t, "REQ-3", "REV-1")
	r1, err := NewRefinement(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), refining, mustProvenance(t), mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewRefinement(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), refining, mustProvenance(t), mustScope(t, "product-x", "/y"))
	if err != nil {
		t.Fatal(err)
	}
	if r1.Refining() != r2.Refining() {
		t.Error("expected both refinements to share the same refining target")
	}
}
