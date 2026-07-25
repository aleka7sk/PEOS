package core

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func mustProvenance(t *testing.T) Provenance {
	t.Helper()
	ts, err := NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	actor, err := NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return NewProvenance().WithActor(actor).WithRecordedAt(ts)
}

func mustOrigin(t *testing.T) Origin {
	t.Helper()
	o, err := NewOrigin(OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustIntegrity(t *testing.T) IntegrityIdentity {
	t.Helper()
	i, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123", IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func mustRepresentation(t *testing.T, text string) Representation {
	t.Helper()
	rep, err := NewRepresentationFromInlineText(text, mustVocabularyValue(t, "mime", "text/markdown"), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

// --- Origin ---------------------------------------------------------------

func TestNewOriginKnown(t *testing.T) {
	o, err := NewOrigin(OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	if o.IsZero() {
		t.Error("valid Origin reports IsZero() = true")
	}
	if o.Kind() != OriginKindKnown {
		t.Errorf("Kind() = %v, want %v", o.Kind(), OriginKindKnown)
	}
	if _, ok := o.Note(); ok {
		t.Error("Note() ok = true for an empty note")
	}
}

func TestNewOriginNonKnownRequiresNote(t *testing.T) {
	for _, kind := range []OriginKind{OriginKindUnknown, OriginKindUnavailable, OriginKindReconstructed} {
		if _, err := NewOrigin(kind, ""); !errors.Is(err, ErrInvalidOrigin) {
			t.Errorf("kind %v with empty note: error = %v, want %v", kind, err, ErrInvalidOrigin)
		}
		o, err := NewOrigin(kind, "explanation")
		if err != nil {
			t.Errorf("kind %v with note: unexpected error: %v", kind, err)
		}
		if note, ok := o.Note(); !ok || note != "explanation" {
			t.Errorf("Note() = (%q, %v), want (\"explanation\", true)", note, ok)
		}
	}
}

func TestNewOriginZeroKindRejected(t *testing.T) {
	if _, err := NewOrigin(OriginKind{}, "note"); !errors.Is(err, ErrInvalidOrigin) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOrigin)
	}
}

func TestOriginUnknownVocabularyPreserved(t *testing.T) {
	custom := NewOriginKind(mustVocabularyValue(t, "product-x", "custom-kind"))
	o, err := NewOrigin(custom, "explanation required for a non-standard kind")
	if err != nil {
		t.Fatal(err)
	}
	if o.Kind() != custom {
		t.Errorf("Kind() = %v, want %v", o.Kind(), custom)
	}
}

func TestOriginWithExtension(t *testing.T) {
	o := mustOrigin(t)
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	o = o.WithExtension(ext)
	got, ok := o.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestOriginJSONRoundTrip(t *testing.T) {
	original, err := NewOrigin(OriginKindReconstructed, "reconstructed from a backup")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Origin
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Kind() != original.Kind() {
		t.Errorf("round trip Kind() = %v, want %v", decoded.Kind(), original.Kind())
	}
	if note, ok := decoded.Note(); !ok || note != "reconstructed from a backup" {
		t.Errorf("round trip Note() = (%q, %v)", note, ok)
	}
}

func TestOriginZeroValue(t *testing.T) {
	var o Origin
	if !o.IsZero() {
		t.Error("zero-value Origin.IsZero() = false, want true")
	}
}

// --- IntegrityIdentity ------------------------------------------------------

func TestNewIntegrityIdentityValidUnknownMechanism(t *testing.T) {
	mechanism := NewIntegrityMechanism(mustVocabularyValue(t, "product-x", "custom-mechanism"))
	i, err := NewIntegrityIdentity(mechanism, "opaque-value", IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	if i.Mechanism() != mechanism {
		t.Errorf("Mechanism() = %v, want %v", i.Mechanism(), mechanism)
	}
}

func TestNewIntegrityIdentityZeroMechanism(t *testing.T) {
	if _, err := NewIntegrityIdentity(IntegrityMechanism{}, "value", IntegrityProtectedScopeContent); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
}

func TestNewIntegrityIdentityEmptyValue(t *testing.T) {
	if _, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "", IntegrityProtectedScopeContent); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
}

func TestNewIntegrityIdentityZeroProtectedScope(t *testing.T) {
	if _, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "value", IntegrityProtectedScope{}); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
}

func TestIntegrityIdentityWithExtension(t *testing.T) {
	i := mustIntegrity(t)
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	i = i.WithExtension(ext)
	got, ok := i.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestIntegrityIdentityJSONRoundTrip(t *testing.T) {
	original := mustIntegrity(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded IntegrityIdentity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Mechanism() != original.Mechanism() || decoded.Value() != original.Value() || decoded.ProtectedScope() != original.ProtectedScope() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestIntegrityIdentityZeroValue(t *testing.T) {
	var i IntegrityIdentity
	if !i.IsZero() {
		t.Error("zero-value IntegrityIdentity.IsZero() = false, want true")
	}
}

// --- ArtifactRevision --------------------------------------------------------

func TestNewArtifactRevisionZeroRepresentations(t *testing.T) {
	rev, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	if rev.IsZero() {
		t.Error("valid ArtifactRevision reports IsZero() = true")
	}
	if reps := rev.Representations(); reps != nil {
		t.Errorf("Representations() = %v, want nil", reps)
	}
}

func TestNewArtifactRevisionMultipleRepresentations(t *testing.T) {
	rev, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	rep1 := mustRepresentation(t, "first")
	rep2 := mustRepresentation(t, "second")
	rev, err = rev.WithRepresentations(rep1, rep2)
	if err != nil {
		t.Fatal(err)
	}
	got := rev.Representations()
	if len(got) != 2 {
		t.Fatalf("Representations() len = %d, want 2", len(got))
	}
	text1, _ := got[0].Content().AsInlineText()
	text2, _ := got[1].Content().AsInlineText()
	if text1 != "first" || text2 != "second" {
		t.Errorf("Representations() order = (%q, %q), want (\"first\", \"second\")", text1, text2)
	}
}

func TestNewArtifactRevisionZeroArtifactID(t *testing.T) {
	if _, err := NewArtifactRevision(ArtifactID{}, mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t)); !errors.Is(err, ErrInvalidArtifactRevision) || !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want wrapping both %v and %v", err, ErrInvalidArtifactRevision, ErrEmptyIdentity)
	}
}

func TestNewArtifactRevisionZeroRevisionID(t *testing.T) {
	if _, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), ArtifactRevisionID{}, mustOrigin(t), mustProvenance(t), mustIntegrity(t)); !errors.Is(err, ErrInvalidArtifactRevision) || !errors.Is(err, ErrMissingRevisionID) {
		t.Errorf("error = %v, want wrapping both %v and %v", err, ErrInvalidArtifactRevision, ErrMissingRevisionID)
	}
}

func TestNewArtifactRevisionZeroOrigin(t *testing.T) {
	if _, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), Origin{}, mustProvenance(t), mustIntegrity(t)); !errors.Is(err, ErrInvalidArtifactRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidArtifactRevision)
	}
}

func TestNewArtifactRevisionZeroProvenance(t *testing.T) {
	if _, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), Provenance{}, mustIntegrity(t)); !errors.Is(err, ErrInvalidArtifactRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidArtifactRevision)
	}
}

func TestNewArtifactRevisionZeroIntegrity(t *testing.T) {
	if _, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), IntegrityIdentity{}); !errors.Is(err, ErrInvalidArtifactRevision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidArtifactRevision)
	}
}

func TestArtifactRevisionRepresentationsDefensiveCopy(t *testing.T) {
	rev, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	rev, err = rev.WithRepresentations(mustRepresentation(t, "original"))
	if err != nil {
		t.Fatal(err)
	}
	got := rev.Representations()
	got[0] = mustRepresentation(t, "tampered")
	again := rev.Representations()
	text, _ := again[0].Content().AsInlineText()
	if text != "original" {
		t.Errorf("mutating a Representations() result affected internal state: got %q, want %q", text, "original")
	}
}

func TestArtifactRevisionWithRepresentationsRejectsZeroValue(t *testing.T) {
	rev, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rev.WithRepresentations(Representation{}); !errors.Is(err, ErrInvalidRepresentation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRepresentation)
	}
}

func TestArtifactRevisionWithExtension(t *testing.T) {
	rev, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	rev = rev.WithExtension(ext)
	got, ok := rev.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestArtifactRevisionJSONRoundTrip(t *testing.T) {
	original, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithRepresentations(mustRepresentation(t, "content"))
	if err != nil {
		t.Fatal(err)
	}
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithExtension(ext)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ArtifactRevision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ArtifactID() != original.ArtifactID() || decoded.RevisionID() != original.RevisionID() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
	if len(decoded.Representations()) != 1 {
		t.Errorf("round trip Representations() len = %d, want 1", len(decoded.Representations()))
	}
}

func TestArtifactRevisionJSONMalformed(t *testing.T) {
	var rev ArtifactRevision
	err := json.Unmarshal([]byte(`{"artifact_id":"ART-1"}`), &rev)
	if !errors.Is(err, ErrInvalidArtifactRevision) {
		t.Errorf("missing required fields: error = %v, want %v", err, ErrInvalidArtifactRevision)
	}
}

func TestArtifactRevisionJSONHasNoOrderingStatusOrCurrentFields(t *testing.T) {
	original, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
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
	for _, forbidden := range []string{
		"current_revision_id", "status", "lifecycle_state",
		"predecessors", "revision_number", "semantic_version",
	} {
		if _, present := raw[forbidden]; present {
			t.Errorf("forbidden field %q present in ArtifactRevision JSON", forbidden)
		}
	}
}

func TestArtifactRevisionZeroValue(t *testing.T) {
	var rev ArtifactRevision
	if !rev.IsZero() {
		t.Error("zero-value ArtifactRevision.IsZero() = false, want true")
	}
}
