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
	for _, kind := range []OriginKind{OriginKindUnknown, OriginKindUnavailable, OriginKindDisputed, OriginKindReconstructed} {
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

func TestNewOriginKindAndString(t *testing.T) {
	v := mustVocabularyValue(t, "product-x", "migrated")
	kind := NewOriginKind(v)
	if kind.IsZero() {
		t.Error("constructed OriginKind reports IsZero() = true")
	}
	if kind.Value() != v {
		t.Errorf("Value() = %v, want %v", kind.Value(), v)
	}
	if got, want := kind.String(), "product-x:migrated"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestOriginUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewOrigin(OriginKindReconstructed, "pre-existing valid data")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"kind":"","note":""}`), &receiver); err == nil {
		t.Fatal("malformed Origin JSON accepted, want error")
	}
	// Origin is not comparable with == (it embeds Extension, which holds a
	// map), so every field is checked individually via its own accessor.
	if receiver.Kind() != original.Kind() {
		t.Errorf("failed Unmarshal changed Kind(): got %v, want %v", receiver.Kind(), original.Kind())
	}
	if gotNote, gotOK := receiver.Note(); gotNote != "pre-existing valid data" || !gotOK {
		t.Errorf("failed Unmarshal changed Note(): got (%q, %v)", gotNote, gotOK)
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

func TestNewIntegrityIdentityZeroLengthProtectedScopesRejected(t *testing.T) {
	if _, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "value"); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("no scopes given: error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
}

func TestNewIntegrityIdentityZeroProtectedScopeRejected(t *testing.T) {
	if _, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "value", IntegrityProtectedScope{}); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
	if _, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "value", IntegrityProtectedScopeContent, IntegrityProtectedScope{}); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("zero scope among valid ones: error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
}

func TestNewIntegrityIdentityMultipleProtectedScopesOrderPreserved(t *testing.T) {
	i, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123",
		IntegrityProtectedScopeContent, IntegrityProtectedScopeMetadata, IntegrityProtectedScopeRepresentation)
	if err != nil {
		t.Fatal(err)
	}
	got := i.ProtectedScopes()
	want := []IntegrityProtectedScope{IntegrityProtectedScopeContent, IntegrityProtectedScopeMetadata, IntegrityProtectedScopeRepresentation}
	if len(got) != len(want) {
		t.Fatalf("ProtectedScopes() len = %d, want %d", len(got), len(want))
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Errorf("ProtectedScopes()[%d] = %v, want %v (order not preserved)", idx, got[idx], want[idx])
		}
	}
}

func TestNewIntegrityIdentityDuplicateProtectedScopeRejected(t *testing.T) {
	_, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123",
		IntegrityProtectedScopeContent, IntegrityProtectedScopeContent)
	if !errors.Is(err, ErrDuplicateIntegrityProtectedScope) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateIntegrityProtectedScope)
	}
}

func TestNewIntegrityIdentityUnknownProtectedScopeAccepted(t *testing.T) {
	custom := NewIntegrityProtectedScope(mustVocabularyValue(t, "product-x", "custom-scope"))
	i, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123", custom)
	if err != nil {
		t.Fatal(err)
	}
	got := i.ProtectedScopes()
	if len(got) != 1 || got[0] != custom {
		t.Errorf("ProtectedScopes() = %v, want [%v]", got, custom)
	}
}

func TestIntegrityIdentityProtectedScopesDefensiveCopy(t *testing.T) {
	i, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123",
		IntegrityProtectedScopeContent, IntegrityProtectedScopeMetadata)
	if err != nil {
		t.Fatal(err)
	}
	got := i.ProtectedScopes()
	got[0] = IntegrityProtectedScopeRelations
	again := i.ProtectedScopes()
	if again[0] != IntegrityProtectedScopeContent {
		t.Errorf("mutating a ProtectedScopes() result affected internal state: got %v, want %v", again[0], IntegrityProtectedScopeContent)
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
	original, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123",
		IntegrityProtectedScopeContent, IntegrityProtectedScopeMetadata)
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
	if _, present := raw["protected_scope"]; present {
		t.Error("legacy singular protected_scope field present in Marshal output; MarshalJSON must only ever write protected_scopes")
	}
	if _, present := raw["protected_scopes"]; !present {
		t.Error("protected_scopes field missing from Marshal output")
	}

	var decoded IntegrityIdentity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Mechanism() != original.Mechanism() || decoded.Value() != original.Value() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
	got, want := decoded.ProtectedScopes(), original.ProtectedScopes()
	if len(got) != len(want) {
		t.Fatalf("round trip ProtectedScopes() len = %d, want %d", len(got), len(want))
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Errorf("round trip ProtectedScopes()[%d] = %v, want %v", idx, got[idx], want[idx])
		}
	}
}

func TestIntegrityIdentityJSONLegacyScalarCompatibility(t *testing.T) {
	var decoded IntegrityIdentity
	legacy := `{"mechanism":"peos:cryptographic-digest","value":"sha256:abc123","protected_scope":"peos:content"}`
	if err := json.Unmarshal([]byte(legacy), &decoded); err != nil {
		t.Fatalf("legacy scalar protected_scope rejected: %v", err)
	}
	got := decoded.ProtectedScopes()
	if len(got) != 1 || got[0] != IntegrityProtectedScopeContent {
		t.Errorf("ProtectedScopes() = %v, want [%v]", got, IntegrityProtectedScopeContent)
	}
}

func TestIntegrityIdentityJSONBothScalarAndListRejected(t *testing.T) {
	var decoded IntegrityIdentity
	ambiguous := `{"mechanism":"peos:cryptographic-digest","value":"sha256:abc123","protected_scope":"peos:content","protected_scopes":["peos:metadata"]}`
	if err := json.Unmarshal([]byte(ambiguous), &decoded); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
}

func TestIntegrityIdentityJSONEmptyPluralPlusLegacyRejected(t *testing.T) {
	original, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:pre-existing", IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	// protected_scopes is present as an explicit empty array, not absent.
	// Presence MUST be decided by whether the JSON key was supplied, not by
	// the array's length, so this must be detected as both fields present
	// and rejected as ambiguous, the same as a non-empty plural array would
	// be.
	ambiguous := `{"mechanism":"peos:cryptographic-digest","value":"abc","protected_scopes":[],"protected_scope":"peos:content"}`
	if err := json.Unmarshal([]byte(ambiguous), &receiver); !errors.Is(err, ErrInvalidIntegrityIdentity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidIntegrityIdentity)
	}
	if receiver.Mechanism() != original.Mechanism() || receiver.Value() != original.Value() {
		t.Errorf("failed Unmarshal changed Mechanism()/Value(): got (%v, %q), want (%v, %q)",
			receiver.Mechanism(), receiver.Value(), original.Mechanism(), original.Value())
	}
	gotScopes, wantScopes := receiver.ProtectedScopes(), original.ProtectedScopes()
	if len(gotScopes) != len(wantScopes) || gotScopes[0] != wantScopes[0] {
		t.Errorf("failed Unmarshal changed ProtectedScopes(): got %v, want %v", gotScopes, wantScopes)
	}
}

func TestIntegrityIdentityJSONDuplicatePluralScopesRejected(t *testing.T) {
	original, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:pre-existing", IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	duplicate := `{"mechanism":"peos:cryptographic-digest","value":"abc","protected_scopes":["peos:content","peos:content"]}`
	err = json.Unmarshal([]byte(duplicate), &receiver)
	// NewIntegrityIdentity's existing duplicate-scope error wraps only
	// ErrDuplicateIntegrityProtectedScope, not ErrInvalidIntegrityIdentity
	// (see artifact_revision.go's duplicate-scope check) — that wrapping
	// contract predates this fix and is unchanged by it.
	if !errors.Is(err, ErrDuplicateIntegrityProtectedScope) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateIntegrityProtectedScope)
	}
	if receiver.Mechanism() != original.Mechanism() || receiver.Value() != original.Value() {
		t.Errorf("failed Unmarshal changed Mechanism()/Value(): got (%v, %q), want (%v, %q)",
			receiver.Mechanism(), receiver.Value(), original.Mechanism(), original.Value())
	}
	gotScopes, wantScopes := receiver.ProtectedScopes(), original.ProtectedScopes()
	if len(gotScopes) != len(wantScopes) || gotScopes[0] != wantScopes[0] {
		t.Errorf("failed Unmarshal changed ProtectedScopes(): got %v, want %v", gotScopes, wantScopes)
	}
}

func TestIntegrityIdentityJSONExtensionWithPluralScopesRoundTrip(t *testing.T) {
	original, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:abc123",
		IntegrityProtectedScopeContent, IntegrityProtectedScopeMetadata)
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
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["protected_scopes"]; !present {
		t.Error("protected_scopes field missing from Marshal output")
	}
	if _, present := raw["protected_scope"]; present {
		t.Error("legacy singular protected_scope field present in Marshal output; MarshalJSON must only ever write protected_scopes")
	}

	var decoded IntegrityIdentity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Mechanism() != original.Mechanism() || decoded.Value() != original.Value() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
	gotScopes, wantScopes := decoded.ProtectedScopes(), original.ProtectedScopes()
	if len(gotScopes) != len(wantScopes) {
		t.Fatalf("round trip ProtectedScopes() len = %d, want %d", len(gotScopes), len(wantScopes))
	}
	for idx := range wantScopes {
		if gotScopes[idx] != wantScopes[idx] {
			t.Errorf("round trip ProtectedScopes()[%d] = %v, want %v", idx, gotScopes[idx], wantScopes[idx])
		}
	}
	got, ok := decoded.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("round trip Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestIntegrityIdentityUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewIntegrityIdentity(IntegrityMechanismCryptographicDigest, "sha256:pre-existing", IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"mechanism":"","value":""}`), &receiver); err == nil {
		t.Fatal("malformed IntegrityIdentity JSON accepted, want error")
	}
	if receiver.Mechanism() != original.Mechanism() || receiver.Value() != original.Value() {
		t.Errorf("failed Unmarshal changed Mechanism()/Value(): got (%v, %q), want (%v, %q)",
			receiver.Mechanism(), receiver.Value(), original.Mechanism(), original.Value())
	}
	gotScopes, wantScopes := receiver.ProtectedScopes(), original.ProtectedScopes()
	if len(gotScopes) != len(wantScopes) || gotScopes[0] != wantScopes[0] {
		t.Errorf("failed Unmarshal changed ProtectedScopes(): got %v, want %v", gotScopes, wantScopes)
	}
}

func TestNewIntegrityMechanismAndAccessors(t *testing.T) {
	v := mustVocabularyValue(t, "product-x", "custom-mechanism")
	m := NewIntegrityMechanism(v)
	if m.IsZero() {
		t.Error("constructed IntegrityMechanism reports IsZero() = true")
	}
	if m.Value() != v {
		t.Errorf("Value() = %v, want %v", m.Value(), v)
	}
	if got, want := m.String(), "product-x:custom-mechanism"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestNewIntegrityProtectedScopeAndAccessors(t *testing.T) {
	v := mustVocabularyValue(t, "product-x", "custom-scope")
	s := NewIntegrityProtectedScope(v)
	if s.IsZero() {
		t.Error("constructed IntegrityProtectedScope reports IsZero() = true")
	}
	if s.Value() != v {
		t.Errorf("Value() = %v, want %v", s.Value(), v)
	}
	if got, want := s.String(), "product-x:custom-scope"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
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

func TestArtifactRevisionOriginProvenanceIntegrityAccessors(t *testing.T) {
	origin := mustOrigin(t)
	provenance := mustProvenance(t)
	integrity := mustIntegrity(t)
	rev, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), origin, provenance, integrity)
	if err != nil {
		t.Fatal(err)
	}
	if got := rev.Origin(); got.Kind() != origin.Kind() {
		t.Errorf("Origin().Kind() = %v, want %v", got.Kind(), origin.Kind())
	}
	if got := rev.Provenance(); got.IsZero() != provenance.IsZero() {
		t.Errorf("Provenance().IsZero() = %v, want %v", got.IsZero(), provenance.IsZero())
	}
	if got := rev.Integrity(); got.Mechanism() != integrity.Mechanism() || got.Value() != integrity.Value() {
		t.Errorf("Integrity() = %+v, want %+v", got, integrity)
	}
}

func TestArtifactRevisionWithRepresentationsReplacesNotAccumulates(t *testing.T) {
	base, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	repA := mustRepresentation(t, "repA")
	repB := mustRepresentation(t, "repB")

	rev1, err := base.WithRepresentations(repA)
	if err != nil {
		t.Fatal(err)
	}
	rev2, err := rev1.WithRepresentations(repB)
	if err != nil {
		t.Fatal(err)
	}

	if got := rev2.Representations(); len(got) != 1 {
		t.Fatalf("rev2.Representations() len = %d, want 1", len(got))
	} else if text, _ := got[0].Content().AsInlineText(); text != "repB" {
		t.Errorf("rev2.Representations()[0] = %q, want %q", text, "repB")
	}

	if got := rev1.Representations(); len(got) != 1 {
		t.Fatalf("rev1.Representations() len = %d, want 1 (must be unchanged by rev2's WithRepresentations call)", len(got))
	} else if text, _ := got[0].Content().AsInlineText(); text != "repA" {
		t.Errorf("rev1.Representations()[0] = %q, want %q (rev1 must remain unaffected)", text, "repA")
	}

	if got := base.Representations(); got != nil {
		t.Errorf("base.Representations() = %v, want nil (base must remain unaffected)", got)
	}

	rev3, err := rev2.WithRepresentations()
	if err != nil {
		t.Fatal(err)
	}
	if got := rev3.Representations(); got != nil {
		t.Errorf("rev3.Representations() after no-args WithRepresentations() = %v, want nil", got)
	}
	if got := rev2.Representations(); len(got) != 1 {
		t.Errorf("rev2.Representations() len = %d, want 1 (must remain unaffected by rev3's clearing call)", len(got))
	}
}

func TestArtifactRevisionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewArtifactRevision(mustArtifactID(t, "ART-1"), mustArtifactRevisionID(t, "REV-1"), mustOrigin(t), mustProvenance(t), mustIntegrity(t))
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithRepresentations(mustRepresentation(t, "pre-existing"))
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"artifact_id":"ART-1"}`), &receiver); err == nil {
		t.Fatal("malformed ArtifactRevision JSON accepted, want error")
	}
	// ArtifactRevision is not comparable with == (it holds a []Representation
	// slice, an Extension map, and nested non-comparable Origin/IntegrityIdentity
	// values), so every field is checked individually via its own accessor.
	if receiver.ArtifactID() != original.ArtifactID() || receiver.RevisionID() != original.RevisionID() {
		t.Errorf("failed Unmarshal changed ArtifactID()/RevisionID(): got (%v, %v), want (%v, %v)",
			receiver.ArtifactID(), receiver.RevisionID(), original.ArtifactID(), original.RevisionID())
	}
	gotReps, wantReps := receiver.Representations(), original.Representations()
	if len(gotReps) != len(wantReps) {
		t.Fatalf("failed Unmarshal changed Representations() len: got %d, want %d", len(gotReps), len(wantReps))
	}
	gotText, _ := gotReps[0].Content().AsInlineText()
	wantText, _ := wantReps[0].Content().AsInlineText()
	if gotText != wantText {
		t.Errorf("failed Unmarshal changed Representations()[0] content: got %q, want %q", gotText, wantText)
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
