package relation

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- shared test helpers ---

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

func mustArtifactSubject(t *testing.T, artifactID string) core.EngineeringSubjectRef {
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

func mustArtifactRevisionSubject(t *testing.T, artifactID, revisionID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	sub, err := core.EngineeringSubjectRefFromArtifactRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func mustOpaqueSubject(t *testing.T, kind, namespace, identifier string) core.EngineeringSubjectRef {
	t.Helper()
	sub, err := core.NewOpaqueEngineeringSubjectRef(kind, namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return sub
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

func mustScope(t *testing.T, namespace, expression string) core.Scope {
	t.Helper()
	s, err := core.NewScope(mustVocab(t, namespace, "condition"), expression)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustRelation(t *testing.T) Relation {
	t.Helper()
	r, err := New(
		core.RelationTypeDerivation,
		mustArtifactRevisionSubject(t, "REQ-1", "REV-1"),
		mustArtifactRevisionSubject(t, "REQ-2", "REV-1"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// --- Construction ---------------------------------------------------------

func TestNewValidArtifactToArtifact(t *testing.T) {
	source := mustArtifactSubject(t, "ART-1")
	target := mustArtifactSubject(t, "ART-2")
	r, err := New(core.RelationTypeArtifactSupersession, source, target, mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.IsZero() {
		t.Error("valid Relation reports IsZero() = true")
	}
	if r.Source() != source || r.Target() != target {
		t.Errorf("Source()/Target() = %v/%v, want %v/%v", r.Source(), r.Target(), source, target)
	}
}

func TestNewValidRevisionToRevision(t *testing.T) {
	source := mustArtifactRevisionSubject(t, "REQ-1", "REV-1")
	target := mustArtifactRevisionSubject(t, "REQ-2", "REV-1")
	r, err := New(core.RelationTypeDerivation, source, target, mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Source() != source || r.Target() != target {
		t.Errorf("Source()/Target() = %v/%v, want %v/%v", r.Source(), r.Target(), source, target)
	}
}

func TestNewValidMixedEndpoint(t *testing.T) {
	source := mustArtifactRevisionSubject(t, "TPL-1", "REV-1")
	target := mustArtifactSubject(t, "GEN-1")
	r, err := New(core.RelationTypeGeneratedFrom, source, target, mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Source() != source || r.Target() != target {
		t.Errorf("Source()/Target() = %v/%v, want %v/%v", r.Source(), r.Target(), source, target)
	}
}

func TestNewValidOpaqueEndpoint(t *testing.T) {
	source := mustOpaqueSubject(t, "product-widget", "product-x", "widget-1")
	target := mustArtifactSubject(t, "ART-1")
	relType := core.NewRelationType(mustVocab(t, "product-x", "widget-uses"))
	r, err := New(relType, source, target, mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Source() != source {
		t.Errorf("Source() = %v, want %v", r.Source(), source)
	}
}

func TestNewZeroRelationTypeRejected(t *testing.T) {
	_, err := New(core.RelationType{}, mustArtifactSubject(t, "ART-1"), mustArtifactSubject(t, "ART-2"), mustProvenance(t))
	if !errors.Is(err, ErrInvalidRelationType) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelationType)
	}
}

func TestNewZeroSourceRejected(t *testing.T) {
	_, err := New(core.RelationTypeDependency, core.EngineeringSubjectRef{}, mustArtifactSubject(t, "ART-2"), mustProvenance(t))
	if !errors.Is(err, ErrInvalidRelationSource) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelationSource)
	}
}

func TestNewZeroTargetRejected(t *testing.T) {
	_, err := New(core.RelationTypeDependency, mustArtifactSubject(t, "ART-1"), core.EngineeringSubjectRef{}, mustProvenance(t))
	if !errors.Is(err, ErrInvalidRelationTarget) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelationTarget)
	}
}

func TestNewZeroProvenanceRejected(t *testing.T) {
	_, err := New(core.RelationTypeDependency, mustArtifactSubject(t, "ART-1"), mustArtifactSubject(t, "ART-2"), core.Provenance{})
	if !errors.Is(err, ErrInvalidRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelation)
	}
}

func TestNewCustomRelationTypeAccepted(t *testing.T) {
	relType := core.NewRelationType(mustVocab(t, "product-x", "custom-relation"))
	r, err := New(relType, mustArtifactSubject(t, "ART-1"), mustArtifactSubject(t, "ART-2"), mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.RelationType() != relType {
		t.Errorf("RelationType() = %v, want %v", r.RelationType(), relType)
	}
}

func TestNewSourceEqualsTargetAcceptedStructurally(t *testing.T) {
	subject := mustArtifactSubject(t, "ART-1")
	r, err := New(core.RelationTypeDependency, subject, subject, mustProvenance(t))
	if err != nil {
		t.Fatalf("self-relation unexpectedly rejected at the generic Packet D level: %v", err)
	}
	if r.Source() != r.Target() {
		t.Error("Source() != Target() for a deliberately constructed self-relation")
	}
}

// --- Accessors ---------------------------------------------------------

func TestAccessorsReturnConstructorValues(t *testing.T) {
	relType := core.RelationTypeConflict
	source := mustArtifactSubject(t, "ART-1")
	target := mustArtifactSubject(t, "ART-2")
	prov := mustProvenance(t)
	r, err := New(relType, source, target, prov)
	if err != nil {
		t.Fatal(err)
	}
	if r.RelationType() != relType {
		t.Errorf("RelationType() = %v, want %v", r.RelationType(), relType)
	}
	if r.Source() != source {
		t.Errorf("Source() = %v, want %v", r.Source(), source)
	}
	if r.Target() != target {
		t.Errorf("Target() = %v, want %v", r.Target(), target)
	}
	gotActor, ok := r.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if !ok || gotActor != wantActor {
		t.Errorf("Provenance().Actor() = (%v, %v), want (%v, true)", gotActor, ok, wantActor)
	}
}

func TestZeroRelationIsZero(t *testing.T) {
	var r Relation
	if !r.IsZero() {
		t.Error("zero-value Relation.IsZero() = false, want true")
	}
}

func TestValidRelationIsNotZero(t *testing.T) {
	if mustRelation(t).IsZero() {
		t.Error("valid Relation reports IsZero() = true")
	}
}

// --- Scope ---------------------------------------------------------

func TestWithScopeValid(t *testing.T) {
	base := mustRelation(t)
	scope := mustScope(t, "product-x", "deployment=eu")
	r, err := base.WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := r.Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}
}

func TestWithScopeZeroRejected(t *testing.T) {
	base := mustRelation(t)
	if _, err := base.WithScope(core.Scope{}); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestWithoutScopeClears(t *testing.T) {
	base := mustRelation(t)
	scope := mustScope(t, "product-x", "deployment=eu")
	withScope, err := base.WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	cleared := withScope.WithoutScope()
	if _, ok := cleared.Scope(); ok {
		t.Error("Scope() ok = true after WithoutScope()")
	}
}

func TestWithScopeDoesNotMutateOldRelation(t *testing.T) {
	base := mustRelation(t)
	if _, ok := base.Scope(); ok {
		t.Fatal("base unexpectedly already has a scope")
	}
	scope := mustScope(t, "product-x", "deployment=eu")
	withScope, err := base.WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := base.Scope(); ok {
		t.Error("WithScope mutated the original receiver: base now has a scope")
	}
	if _, ok := withScope.Scope(); !ok {
		t.Error("withScope has no scope after WithScope")
	}
}

func TestScopeDistinguishesAbsenceFromPresence(t *testing.T) {
	base := mustRelation(t)
	if _, ok := base.Scope(); ok {
		t.Error("Scope() ok = true for a Relation with no declared scope")
	}
	withScope, err := base.WithScope(mustScope(t, "product-x", "x"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withScope.Scope(); !ok {
		t.Error("Scope() ok = false for a Relation with a declared scope")
	}
}

// --- Extension ---------------------------------------------------------

func TestWithExtensionValid(t *testing.T) {
	base := mustRelation(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	r := base.WithExtension(ext)
	got, ok := r.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestWithExtensionReplaceSemantics(t *testing.T) {
	base := mustRelation(t)
	ext1, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	ext2, err := core.NewExtension().With("product-y", json.RawMessage(`{"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	r1 := base.WithExtension(ext1)
	r2 := r1.WithExtension(ext2)

	if _, ok := r2.Extension().Get("product-x"); ok {
		t.Error("r2 unexpectedly still has product-x extension data after replacement")
	}
	if got, ok := r2.Extension().Get("product-y"); !ok || string(got) != `{"b":2}` {
		t.Errorf("r2 Extension().Get(\"product-y\") = (%s, %v)", got, ok)
	}
	if got, ok := r1.Extension().Get("product-x"); !ok || string(got) != `{"a":1}` {
		t.Errorf("r1.Extension().Get(\"product-x\") = (%s, %v), r1 must remain unaffected by r2's WithExtension call", got, ok)
	}
}

func TestWithoutExtensionClears(t *testing.T) {
	base := mustRelation(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := base.WithExtension(ext)
	cleared := withExt.WithoutExtension()
	if !cleared.Extension().IsZero() {
		t.Error("Extension().IsZero() = false after WithoutExtension()")
	}
}

func TestWithExtensionReceiverImmutability(t *testing.T) {
	base := mustRelation(t)
	if !base.Extension().IsZero() {
		t.Fatal("base unexpectedly already has extension data")
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = base.WithExtension(ext)
	if !base.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver: base now has extension data")
	}
}

func TestExtensionAbsenceDistinguishable(t *testing.T) {
	base := mustRelation(t)
	if !base.Extension().IsZero() {
		t.Error("Extension().IsZero() = false for a Relation with no declared extension")
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := base.WithExtension(ext)
	if withExt.Extension().IsZero() {
		t.Error("Extension().IsZero() = true for a Relation with declared extension data")
	}
}

// --- JSON ---------------------------------------------------------

func TestRelationJSONRoundTripArtifactToArtifact(t *testing.T) {
	original, err := New(core.RelationTypeArtifactSupersession, mustArtifactSubject(t, "ART-1"), mustArtifactSubject(t, "ART-2"), mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Relation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.RelationType() != original.RelationType() || decoded.Source() != original.Source() || decoded.Target() != original.Target() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestRelationJSONRoundTripRevisionToRevision(t *testing.T) {
	original, err := New(core.RelationTypeDerivation, mustArtifactRevisionSubject(t, "REQ-1", "REV-1"), mustArtifactRevisionSubject(t, "REQ-2", "REV-1"), mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Relation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Source() != original.Source() || decoded.Target() != original.Target() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestRelationJSONRoundTripMixedEndpoint(t *testing.T) {
	original, err := New(core.RelationTypeGeneratedFrom, mustArtifactRevisionSubject(t, "TPL-1", "REV-1"), mustArtifactSubject(t, "GEN-1"), mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Relation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Source() != original.Source() || decoded.Target() != original.Target() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestRelationJSONRoundTripCustomRelationType(t *testing.T) {
	relType := core.NewRelationType(mustVocab(t, "product-x", "custom-relation"))
	original, err := New(relType, mustArtifactSubject(t, "ART-1"), mustArtifactSubject(t, "ART-2"), mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var relTypeStr string
	if err := json.Unmarshal(raw["relation_type"], &relTypeStr); err != nil {
		t.Fatal(err)
	}
	if relTypeStr != "product-x:custom-relation" {
		t.Errorf("relation_type = %q, want %q", relTypeStr, "product-x:custom-relation")
	}
	var decoded Relation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.RelationType() != relType {
		t.Errorf("round trip RelationType() = %v, want %v", decoded.RelationType(), relType)
	}
}

func TestRelationJSONRoundTripWithScope(t *testing.T) {
	scope := mustScope(t, "product-x", "deployment=eu")
	original, err := mustRelation(t).WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Relation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, ok := decoded.Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("round trip Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}
}

func TestRelationJSONRoundTripWithExtension(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original := mustRelation(t).WithExtension(ext)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Relation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, ok := decoded.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("round trip Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestRelationJSONMinimumOmitsScopeAndExtension(t *testing.T) {
	original := mustRelation(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"relation_type", "source", "target", "provenance"} {
		if _, present := raw[required]; !present {
			t.Errorf("required field %q missing from Marshal output", required)
		}
	}
	for _, optional := range []string{"scope", "extension"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}
}

func TestRelationZeroValueMarshalFails(t *testing.T) {
	var r Relation
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRelation) {
		t.Errorf("Marshal(zero Relation): error = %v, want %v", err, ErrInvalidRelation)
	}
}

func TestRelationJSONMissingRelationTypeRejected(t *testing.T) {
	var r Relation
	payload := `{"source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{}}`
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRelationType) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelationType)
	}
}

func TestRelationJSONMissingSourceRejected(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"peos:dependency","target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{}}`
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRelationSource) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelationSource)
	}
}

func TestRelationJSONMissingTargetRejected(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"provenance":{}}`
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRelationTarget) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelationTarget)
	}
}

func TestRelationJSONMissingProvenanceRejected(t *testing.T) {
	// core.Provenance.MarshalJSON/UnmarshalJSON always round-trips a
	// well-formed (possibly empty) JSON object; an entirely absent
	// "provenance" key decodes to the zero Provenance, which New then
	// correctly rejects as zero -- verified directly, not assumed.
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}}}`
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, ErrInvalidRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRelation)
	}
}

func TestRelationJSONMalformedRelationTypeRejected(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"no-colon-here","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{}}`
	err := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(err, core.ErrInvalidVocabularyValue) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidVocabularyValue)
	}
}

func TestRelationJSONInvalidSourceDiscriminatorRejected(t *testing.T) {
	// An empty "kind" is rejected directly by EngineeringSubjectRef's own
	// UnmarshalJSON with ErrInvalidReferenceDiscriminator. A structurally
	// unrecognized non-empty kind (e.g. "not_a_real_kind") is instead
	// routed through core's opaque-subject fallback and fails there with
	// a different, core-owned identity error -- that fallback behavior
	// belongs to core.EngineeringSubjectRef, not to this package, so this
	// test exercises the direct, unambiguous discriminator-invalid case.
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"","ref":{}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	err := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(err, core.ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidReferenceDiscriminator)
	}
}

func TestRelationJSONInvalidTargetDiscriminatorRejected(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"","ref":{}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	err := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(err, core.ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidReferenceDiscriminator)
	}
}

func TestRelationJSONScopeNullRejected(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}},"scope":null}`
	// An explicit "scope": null is distinct from an absent "scope" key
	// and must be rejected, not silently treated as "no scope" --
	// see TestRelationJSONScopeAbsentMeansNoScope for the absent case.
	err := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestRelationJSONScopeAbsentMeansNoScope(t *testing.T) {
	var r Relation
	// Hand-built minimum-valid JSON with no "scope" key at all.
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatalf("absent scope key unexpectedly rejected: %v", err)
	}
	if _, ok := r.Scope(); ok {
		t.Error("Scope() ok = true after decoding a payload with no scope key")
	}
}

func TestRelationJSONScopeNullPreservesReceiver(t *testing.T) {
	scope := mustScope(t, "product-x", "deployment=eu")
	original, err := mustRelation(t).WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithExtension(ext)

	receiver := original
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}},"scope":null}`
	if err := json.Unmarshal([]byte(payload), &receiver); !errors.Is(err, core.ErrInvalidScope) {
		t.Fatalf("error = %v, want %v", err, core.ErrInvalidScope)
	}

	// Relation is not comparable with == (it holds core.Extension, which
	// embeds a map), so every field is checked individually via its own
	// accessor.
	if receiver.RelationType() != original.RelationType() {
		t.Errorf("failed Unmarshal changed RelationType(): got %v, want %v", receiver.RelationType(), original.RelationType())
	}
	if receiver.Source() != original.Source() {
		t.Errorf("failed Unmarshal changed Source(): got %v, want %v", receiver.Source(), original.Source())
	}
	if receiver.Target() != original.Target() {
		t.Errorf("failed Unmarshal changed Target(): got %v, want %v", receiver.Target(), original.Target())
	}
	gotActor, gotOK := receiver.Provenance().Actor()
	wantActor, wantOK := original.Provenance().Actor()
	if gotActor != wantActor || gotOK != wantOK {
		t.Errorf("failed Unmarshal changed Provenance(): got (%v, %v), want (%v, %v)", gotActor, gotOK, wantActor, wantOK)
	}
	gotScope, gotScopeOK := receiver.Scope()
	wantScope, wantScopeOK := original.Scope()
	if gotScopeOK != wantScopeOK || !gotScope.Equal(wantScope) {
		t.Errorf("failed Unmarshal changed Scope(): got (%v, %v), want (%v, %v)", gotScope, gotScopeOK, wantScope, wantScopeOK)
	}
	gotExt, gotExtOK := receiver.Extension().Get("product-x")
	wantExt, wantExtOK := original.Extension().Get("product-x")
	if gotExtOK != wantExtOK || string(gotExt) != string(wantExt) {
		t.Errorf("failed Unmarshal changed Extension(): got (%s, %v), want (%s, %v)", gotExt, gotExtOK, wantExt, wantExtOK)
	}
}

func TestRelationJSONScopeEmptyObjectRejected(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}},"scope":{}}`
	err := json.Unmarshal([]byte(payload), &r)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestRelationJSONWholePayloadNullRejected(t *testing.T) {
	original := mustRelation(t)
	receiver := original
	if err := json.Unmarshal([]byte(`null`), &receiver); err == nil {
		t.Fatal("whole-payload null accepted, want error")
	}
	// encoding/json's default behavior for null into a struct pointer is
	// a silent no-op (no error, receiver untouched) UNLESS the type
	// itself defines UnmarshalJSON, in which case UnmarshalJSON receives
	// the literal bytes "null" and is responsible for its own behavior;
	// this is asserted directly above, not assumed.
}

func TestRelationUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := mustRelation(t)
	scope := mustScope(t, "product-x", "deployment=eu")
	original, err := original.WithScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithExtension(ext)

	receiver := original
	payload := `{"relation_type":"","source":{},"target":{},"provenance":{}}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("malformed Relation JSON accepted, want error")
	}

	// Relation is not comparable with == (it holds core.Extension, which
	// embeds a map), so every field is checked individually via its own
	// accessor.
	if receiver.RelationType() != original.RelationType() {
		t.Errorf("failed Unmarshal changed RelationType(): got %v, want %v", receiver.RelationType(), original.RelationType())
	}
	if receiver.Source() != original.Source() {
		t.Errorf("failed Unmarshal changed Source(): got %v, want %v", receiver.Source(), original.Source())
	}
	if receiver.Target() != original.Target() {
		t.Errorf("failed Unmarshal changed Target(): got %v, want %v", receiver.Target(), original.Target())
	}
	gotActor, gotOK := receiver.Provenance().Actor()
	wantActor, wantOK := original.Provenance().Actor()
	if gotActor != wantActor || gotOK != wantOK {
		t.Errorf("failed Unmarshal changed Provenance(): got (%v, %v), want (%v, %v)", gotActor, gotOK, wantActor, wantOK)
	}
	gotScope, gotScopeOK := receiver.Scope()
	wantScope, wantScopeOK := original.Scope()
	if gotScopeOK != wantScopeOK || (gotScopeOK && !gotScope.Equal(wantScope)) {
		t.Errorf("failed Unmarshal changed Scope(): got (%v, %v), want (%v, %v)", gotScope, gotScopeOK, wantScope, wantScopeOK)
	}
	gotExt, gotExtOK := receiver.Extension().Get("product-x")
	wantExt, wantExtOK := original.Extension().Get("product-x")
	if gotExtOK != wantExtOK || string(gotExt) != string(wantExt) {
		t.Errorf("failed Unmarshal changed Extension(): got (%s, %v), want (%s, %v)", gotExt, gotExtOK, wantExt, wantExtOK)
	}
}

func TestRelationJSONUnknownOrdinaryFieldIgnored(t *testing.T) {
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{"kind":"artifact","ref":{"artifact_id":"ART-2"}},"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}},"bogus_field":123}`
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatalf("unknown ordinary field unexpectedly rejected: %v", err)
	}
}

func TestRelationJSONNoConstructorBypass(t *testing.T) {
	// Hand-built JSON with a zero-value target ref must still be
	// rejected -- there is no path that bypasses New's validation.
	var r Relation
	payload := `{"relation_type":"peos:dependency","source":{"kind":"artifact","ref":{"artifact_id":"ART-1"}},"target":{},"provenance":{}}`
	if err := json.Unmarshal([]byte(payload), &r); !errors.Is(err, core.ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidReferenceDiscriminator)
	}
}

// --- Open semantics ---------------------------------------------------------

func TestOpenSemanticsUnknownRelationTypeAccepted(t *testing.T) {
	relType := core.NewRelationType(mustVocab(t, "product-x", "totally-unknown-relation"))
	if _, err := New(relType, mustArtifactSubject(t, "ART-1"), mustArtifactSubject(t, "ART-2"), mustProvenance(t)); err != nil {
		t.Errorf("unknown but structurally valid RelationType unexpectedly rejected: %v", err)
	}
}

func TestOpenSemanticsSourceEqualsTargetAccepted(t *testing.T) {
	subject := mustArtifactRevisionSubject(t, "REQ-1", "REV-1")
	if _, err := New(core.RelationTypeDependency, subject, subject, mustProvenance(t)); err != nil {
		t.Errorf("source == target unexpectedly rejected at the generic level: %v", err)
	}
}

func TestOpenSemanticsMixedEndpointKindsAccepted(t *testing.T) {
	source := mustArtifactSubject(t, "ART-1")
	target := mustArtifactRevisionSubject(t, "ART-2", "REV-1")
	if _, err := New(core.RelationTypeArtifactSupersession, source, target, mustProvenance(t)); err != nil {
		t.Errorf("mixed Artifact/ArtifactRevision endpoint kinds unexpectedly rejected: %v", err)
	}
}
