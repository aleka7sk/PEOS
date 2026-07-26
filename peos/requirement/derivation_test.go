package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustRequirementRevisionRef(t *testing.T, artifactID, revisionID string) core.RequirementArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewRequirementArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustDerivation(t *testing.T) Derivation {
	t.Helper()
	d, err := NewDerivation(
		mustRequirementRevisionRef(t, "REQ-1", "REV-1"),
		mustRequirementRevisionRef(t, "REQ-2", "REV-1"),
		mustProvenance(t),
		"Derived from the audit-trail requirement to cover retention specifically.",
	)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func fullDerivation(t *testing.T) Derivation {
	t.Helper()
	d := mustDerivation(t)
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

// --- NewDerivation -----------------------------------------------------

func TestNewDerivationValid(t *testing.T) {
	d := mustDerivation(t)
	if d.IsZero() {
		t.Error("valid Derivation IsZero() = true")
	}
}

func TestNewDerivationZeroSourceRejected(t *testing.T) {
	_, err := NewDerivation(core.RequirementArtifactRevisionRef{}, mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), "rationale")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewDerivationZeroTargetRejected(t *testing.T) {
	_, err := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), core.RequirementArtifactRevisionRef{}, mustProvenance(t), "rationale")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewDerivationSelfDerivationRejected proves the direct
// self-derivation prohibition (PEOS-005 §18.1).
func TestNewDerivationSelfDerivationRejected(t *testing.T) {
	same := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewDerivation(same, same, mustProvenance(t), "rationale")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewDerivationSameRequirementDifferentRevisionsRejected proves the
// Packet G.1.1 audit's authoritative conclusion: PEOS-005 §18 requires
// "A derived Requirement SHALL possess its own identity" and "A derived
// Requirement SHALL NOT inherit the identity of a source Requirement,"
// so a source and target naming different Revisions of the very same
// Requirement identity is non-conforming identity inheritance -- the
// deliberate mirror of Decomposition's equivalent rejection, and the
// opposite of Refinement's equivalent acceptance.
func TestNewDerivationSameRequirementDifferentRevisionsRejected(t *testing.T) {
	source := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	target := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	d, err := NewDerivation(source, target, mustProvenance(t), "rationale")
	if !errors.Is(err, ErrInvalidDerivation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDerivation)
	}
	if !d.IsZero() {
		t.Error("returned Derivation is not zero on identity-inheritance rejection")
	}
}

func TestNewDerivationZeroProvenanceRejected(t *testing.T) {
	_, err := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), core.Provenance{}, "rationale")
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewDerivationEmptyRationaleRejected(t *testing.T) {
	_, err := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), "")
	if !errors.Is(err, ErrInvalidDerivation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDerivation)
	}
}

func TestNewDerivationWhitespaceOnlyRationaleRejected(t *testing.T) {
	_, err := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), "   ")
	if !errors.Is(err, ErrInvalidDerivation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDerivation)
	}
}

func TestNewDerivationRationaleStoredTrimmed(t *testing.T) {
	d, err := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), "  padded rationale  ")
	if err != nil {
		t.Fatal(err)
	}
	if d.Rationale() != "padded rationale" {
		t.Errorf("Rationale() = %q, want trimmed %q", d.Rationale(), "padded rationale")
	}
}

// TestNewDerivationRelationTypeAlwaysDerivation proves the relation type
// is never a caller input: it is fixed internally regardless of what a
// caller might otherwise expect to control.
func TestNewDerivationRelationTypeAlwaysDerivation(t *testing.T) {
	d := mustDerivation(t)
	if d.Relation().RelationType() != core.RelationTypeDerivation {
		t.Errorf("RelationType() = %v, want %v", d.Relation().RelationType(), core.RelationTypeDerivation)
	}
}

// --- With* ---------------------------------------------------------------

func TestDerivationWithScope(t *testing.T) {
	d := mustDerivation(t)
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

func TestDerivationZeroScopeRejected(t *testing.T) {
	d := mustDerivation(t)
	if _, err := d.WithScope(core.Scope{}); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestDerivationWithoutScope(t *testing.T) {
	d := mustDerivation(t)
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

func TestDerivationWithExtension(t *testing.T) {
	d := mustDerivation(t)
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

func TestDerivationWithoutExtension(t *testing.T) {
	d := mustDerivation(t)
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

func TestDerivationWithMethodsAreImmutable(t *testing.T) {
	d := mustDerivation(t)
	original := d
	scope := mustScope(t, "product-x", "/services/*")
	if _, err := d.WithScope(scope); err != nil {
		t.Fatal(err)
	}
	if d.Rationale() != original.Rationale() {
		t.Error("WithScope mutated d")
	}
	if _, ok := d.Scope(); ok {
		t.Error("WithScope mutated d")
	}
}

// --- Accessors -----------------------------------------------------------

func TestDerivationAccessors(t *testing.T) {
	source := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	target := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	prov := mustProvenance(t)
	d, err := NewDerivation(source, target, prov, "rationale")
	if err != nil {
		t.Fatal(err)
	}
	if d.Source() != source {
		t.Errorf("Source() = %v, want %v", d.Source(), source)
	}
	if d.Target() != target {
		t.Errorf("Target() = %v, want %v", d.Target(), target)
	}
	if d.Rationale() != "rationale" {
		t.Errorf("Rationale() = %q, want %q", d.Rationale(), "rationale")
	}
	// core.Provenance is not comparable with == (it holds a
	// core.Extension, which may carry a map); checked via accessors,
	// matching the convention already used for core.Artifact comparisons
	// in requirement_test.go.
	gotActor, _ := d.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if gotActor != wantActor {
		t.Errorf("Provenance().Actor() = %v, want %v", gotActor, wantActor)
	}
}

func TestDerivationIsZero(t *testing.T) {
	var d Derivation
	if !d.IsZero() {
		t.Error("zero Derivation IsZero() = false")
	}
	if mustDerivation(t).IsZero() {
		t.Error("valid Derivation IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestDerivationJSONLiteralWireKeys(t *testing.T) {
	d := fullDerivation(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation", "rationale"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
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

func TestDerivationJSONMinimumOmitsOptionalFields(t *testing.T) {
	d := mustDerivation(t)
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

func TestDerivationJSONRoundTrip(t *testing.T) {
	d := fullDerivation(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Derivation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Source() != d.Source() || decoded.Target() != d.Target() {
		t.Error("participant round trip mismatch")
	}
	if decoded.Rationale() != d.Rationale() {
		t.Error("Rationale mismatch")
	}
	if _, ok := decoded.Scope(); !ok {
		t.Error("Scope absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDerivationJSONMinimumRoundTrip(t *testing.T) {
	d := mustDerivation(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Derivation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Source() != d.Source() {
		t.Error("round trip mismatch")
	}
}

func TestDerivationZeroMarshalRejected(t *testing.T) {
	var d Derivation
	if _, err := json.Marshal(d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDerivationJSONUnknownFieldIgnored(t *testing.T) {
	d := mustDerivation(t)
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
	var decoded Derivation
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestDerivationUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullDerivation(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Rationale() != original.Rationale() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// --- Decode-only validation ------------------------------------------------

func derivationPayload(t *testing.T, sourceKind, targetKind, relationType, rationale string) string {
	t.Helper()
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	sourceRef := `{"artifact_id":"REQ-1","revision_id":"REV-1"}`
	targetRef := `{"artifact_id":"REQ-2","revision_id":"REV-1"}`
	source := `{"kind":"` + sourceKind + `","ref":` + sourceRef + `}`
	target := `{"kind":"` + targetKind + `","ref":` + targetRef + `}`
	if sourceKind == "requirement" {
		source = `{"kind":"requirement","ref":{"artifact_id":"REQ-1"}}`
	}
	if targetKind == "requirement" {
		target = `{"kind":"requirement","ref":{"artifact_id":"REQ-2"}}`
	}
	return `{"relation":{"relation_type":"` + relationType + `","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `},"rationale":"` + rationale + `"}`
}

func TestDerivationJSONWrongRelationTypeRejected(t *testing.T) {
	payload := derivationPayload(t, "requirement_revision", "requirement_revision", "peos:refinement", "rationale")
	var d Derivation
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDerivationJSONSourceAtIdentityLevelRejected(t *testing.T) {
	payload := derivationPayload(t, "requirement", "requirement_revision", "peos:derivation", "rationale")
	var d Derivation
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDerivationJSONTargetAtIdentityLevelRejected(t *testing.T) {
	payload := derivationPayload(t, "requirement_revision", "requirement", "peos:derivation", "rationale")
	var d Derivation
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestDerivationJSONSourceEqualsTargetRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	same := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + same + `,"target":` + same + `,"provenance":` + string(prov) + `},"rationale":"rationale"}`
	var d Derivation
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestDerivationJSONSameRequirementDifferentRevisionsRejected proves the
// §18 identity-distinctness rule is enforced on decode too, and that a
// previously populated receiver is left unchanged by the rejection.
func TestDerivationJSONSameRequirementDifferentRevisionsRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	target := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-2"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `},"rationale":"rationale"}`

	original := fullDerivation(t)
	receiver := original
	err = json.Unmarshal([]byte(payload), &receiver)
	if !errors.Is(err, ErrInvalidDerivation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDerivation)
	}
	if receiver.Rationale() != original.Rationale() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

func TestDerivationJSONMissingRationaleRejected(t *testing.T) {
	payload := derivationPayload(t, "requirement_revision", "requirement_revision", "peos:derivation", "")
	var d Derivation
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidDerivation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDerivation)
	}
}

func TestDerivationJSONExplicitNullRelationRejected(t *testing.T) {
	var d Derivation
	if err := json.Unmarshal([]byte(`{"relation":null,"rationale":"r"}`), &d); err == nil {
		t.Error("null relation accepted, want error")
	}
}

func TestDerivationJSONExplicitNullRationaleRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	target := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-2","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `},"rationale":null}`
	var d Derivation
	if err := json.Unmarshal([]byte(payload), &d); err == nil {
		t.Error("null rationale accepted, want error")
	}
}

// --- Constructor / Unmarshal equivalence ------------------------------------

func TestDerivationConstructorUnmarshalEquivalenceSelfDerivation(t *testing.T) {
	_, ctorErr := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), "rationale")
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	same := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + same + `,"target":` + same + `,"provenance":` + string(prov) + `},"rationale":"rationale"}`
	var d Derivation
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

// TestDerivationConstructorUnmarshalEquivalenceSameRequirementDifferentRevisions
// proves both the constructor and Unmarshal paths enforce the same §18
// identity-distinctness invariant, matching ErrInvalidDerivation on both.
func TestDerivationConstructorUnmarshalEquivalenceSameRequirementDifferentRevisions(t *testing.T) {
	_, ctorErr := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-2"), mustProvenance(t), "rationale")
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	target := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-2"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `},"rationale":"rationale"}`
	var d Derivation
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidDerivation) || !errors.Is(jsonErr, ErrInvalidDerivation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidDerivation)
	}
}

func TestDerivationConstructorUnmarshalEquivalenceEmptyRationale(t *testing.T) {
	_, ctorErr := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), "")
	payload := derivationPayload(t, "requirement_revision", "requirement_revision", "peos:derivation", "")
	var d Derivation
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidDerivation) || !errors.Is(jsonErr, ErrInvalidDerivation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidDerivation)
	}
}

func TestDerivationConstructorUnmarshalEquivalenceZeroProvenance(t *testing.T) {
	_, ctorErr := NewDerivation(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustRequirementRevisionRef(t, "REQ-2", "REV-1"), core.Provenance{}, "rationale")
	source := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`
	target := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-2","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + source + `,"target":` + target + `},"rationale":"rationale"}`
	var d Derivation
	jsonErr := json.Unmarshal([]byte(payload), &d)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

// --- Nested sentinel preservation --------------------------------------

// TestDerivationNestedSentinelPreserved proves a malformed nested
// core reference (empty artifact_id inside the source's requirement
// Artifact Revision ref) preserves both ErrInvalidRequirementRelation and
// the underlying core.ErrEmptyIdentity through errors.Is, matching the
// convention established by TestBasisNestedDecodeErrorsPreserveBothSentinels
// (peos/decision) and the Packet F.1.A/F.2/F.3.B nested-sentinel tests.
func TestDerivationNestedSentinelPreserved(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	malformedSource := `{"kind":"requirement_revision","ref":{"artifact_id":"","revision_id":"REV-1"}}`
	target := `{"kind":"requirement_revision","ref":{"artifact_id":"REQ-2","revision_id":"REV-1"}}`
	payload := `{"relation":{"relation_type":"peos:derivation","source":` + malformedSource + `,"target":` + target + `,"provenance":` + string(prov) + `},"rationale":"rationale"}`
	var d Derivation
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

// TestDerivationNoIdentityField is a structural absence audit proving
// Derivation carries no identity: the fully populated wire form has no
// "id" key, matching PEOS-005 §17.1.
func TestDerivationNoIdentityField(t *testing.T) {
	d := fullDerivation(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["id"]; present {
		t.Error(`unexpected "id" key present in Derivation wire form`)
	}
	if len(raw) != 2 {
		t.Errorf("Derivation wire form has %d top-level keys, want exactly 2 (relation, rationale): %v", len(raw), raw)
	}
}
