package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustStateID(t *testing.T, value string) StateID {
	t.Helper()
	id, err := NewStateID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustTransitionID(t *testing.T, value string) TransitionID {
	t.Helper()
	id, err := NewTransitionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestNewStateRequiresMeaning(t *testing.T) {
	if _, err := NewState(mustStateID(t, "draft"), ""); !errors.Is(err, ErrInvalidState) {
		t.Errorf("empty meaning: error = %v, want %v", err, ErrInvalidState)
	}
	if _, err := NewState(mustStateID(t, "draft"), "   "); !errors.Is(err, ErrInvalidState) {
		t.Errorf("whitespace-only meaning: error = %v, want %v", err, ErrInvalidState)
	}
}

func TestNewStateZeroIDRejected(t *testing.T) {
	if _, err := NewState(StateID{}, "Not yet reviewed"); !errors.Is(err, ErrInvalidStateID) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStateID)
	}
}

func TestStateOpenClassificationAcceptsProductValue(t *testing.T) {
	vocab, err := core.NewVocabularyValue("product-x", "archived")
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewState(mustStateID(t, "draft"), "Not yet reviewed")
	if err != nil {
		t.Fatal(err)
	}
	s = s.WithClassification(NewStateClassification(vocab))
	got, ok := s.Classification()
	if !ok || got.String() != "product-x:archived" {
		t.Errorf("Classification() = (%v, %v), want (product-x:archived, true)", got, ok)
	}
}

func TestStatePredeclaredClassifications(t *testing.T) {
	if StateClassificationOrdinary.String() != "peos:ordinary" {
		t.Errorf("Ordinary = %q, want %q", StateClassificationOrdinary.String(), "peos:ordinary")
	}
	if StateClassificationTerminal.String() != "peos:terminal" {
		t.Errorf("Terminal = %q, want %q", StateClassificationTerminal.String(), "peos:terminal")
	}
	if StateClassificationExceptional.String() != "peos:exceptional" {
		t.Errorf("Exceptional = %q, want %q", StateClassificationExceptional.String(), "peos:exceptional")
	}
}

func TestStateExtensionDefensiveCopy(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewState(mustStateID(t, "draft"), "Not yet reviewed")
	if err != nil {
		t.Fatal(err)
	}
	withExt := s.WithExtension(ext)
	if !s.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	got, ok := withExt.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestStateJSONRoundTrip(t *testing.T) {
	s, err := NewState(mustStateID(t, "accepted"), "Reviewed and accepted")
	if err != nil {
		t.Fatal(err)
	}
	s = s.WithClassification(StateClassificationTerminal)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded State
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != s.ID() || decoded.Meaning() != s.Meaning() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, s)
	}
	gotClass, ok := decoded.Classification()
	if !ok || !gotClass.Equal(StateClassificationTerminal) {
		t.Errorf("Classification() = (%v, %v), want (%v, true)", gotClass, ok, StateClassificationTerminal)
	}
}

func TestStateJSONMinimumOmitsOptionalFields(t *testing.T) {
	s, err := NewState(mustStateID(t, "draft"), "Not yet reviewed")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, optional := range []string{"classification", "extension"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}
}

// TestStateExplicitNullClassificationTreatedAsAbsent documents the
// deliberate difference from Relation.Scope: no PEOS-003 invariant
// distinguishes an explicitly-null classification from an absent one, so
// an explicit null is accepted and behaves identically to an absent key.
func TestStateExplicitNullClassificationTreatedAsAbsent(t *testing.T) {
	payload := `{"id":"draft","meaning":"Not yet reviewed","classification":null}`
	var s State
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Fatalf("explicit null classification unexpectedly rejected: %v", err)
	}
	if _, ok := s.Classification(); ok {
		t.Error("Classification() ok = true after decoding classification:null")
	}
}

func TestStateUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original, err := NewState(mustStateID(t, "draft"), "Not yet reviewed")
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithClassification(StateClassificationOrdinary).WithExtension(ext)

	receiver := original
	payload := `{"id":"draft","meaning":""}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("empty meaning accepted, want error")
	}
	if receiver.ID() != original.ID() || receiver.Meaning() != original.Meaning() {
		t.Error("failed Unmarshal changed receiver")
	}
	gotClass, gotOK := receiver.Classification()
	wantClass, wantOK := original.Classification()
	if gotOK != wantOK || !gotClass.Equal(wantClass) {
		t.Error("failed Unmarshal changed receiver's classification")
	}
}

func TestStateZeroMarshalRejected(t *testing.T) {
	var s State
	if _, err := json.Marshal(s); !errors.Is(err, ErrInvalidState) {
		t.Errorf("error = %v, want %v", err, ErrInvalidState)
	}
}
