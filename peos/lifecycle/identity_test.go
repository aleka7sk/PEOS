package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewStateID(t *testing.T) {
	if _, err := NewStateID(""); !errors.Is(err, ErrInvalidStateID) {
		t.Errorf("empty value: error = %v, want %v", err, ErrInvalidStateID)
	}
	id, err := NewStateID("draft")
	if err != nil {
		t.Fatal(err)
	}
	if id.IsZero() {
		t.Error("valid StateID reports IsZero() = true")
	}
	if id.String() != "draft" {
		t.Errorf("String() = %q, want %q", id.String(), "draft")
	}
}

func TestStateIDJSONRoundTrip(t *testing.T) {
	original, err := NewStateID("accepted")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"accepted"` {
		t.Errorf("Marshal = %s, want %q", data, `"accepted"`)
	}
	var decoded StateID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Equal(original) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestStateIDMalformedJSONRejected(t *testing.T) {
	var id StateID
	if err := json.Unmarshal([]byte(`123`), &id); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
	if err := json.Unmarshal([]byte(`""`), &id); !errors.Is(err, ErrInvalidStateID) {
		t.Errorf("empty string: error = %v, want %v", err, ErrInvalidStateID)
	}
}

func TestStateIDZeroMarshalRejected(t *testing.T) {
	var id StateID
	if _, err := json.Marshal(id); err == nil {
		t.Fatal("Marshal(zero StateID) accepted, want error")
	}
}

func TestStateIDFailedUnmarshalPreservesReceiver(t *testing.T) {
	original, err := NewStateID("accepted")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`""`), &receiver); err == nil {
		t.Fatal("empty value accepted, want error")
	}
	if !receiver.Equal(original) {
		t.Errorf("failed Unmarshal changed receiver: got %v, want %v", receiver, original)
	}
}

func TestNewTransitionID(t *testing.T) {
	if _, err := NewTransitionID(""); !errors.Is(err, ErrInvalidTransitionID) {
		t.Errorf("empty value: error = %v, want %v", err, ErrInvalidTransitionID)
	}
	id, err := NewTransitionID("submit-for-review")
	if err != nil {
		t.Fatal(err)
	}
	if id.IsZero() {
		t.Error("valid TransitionID reports IsZero() = true")
	}
	if id.String() != "submit-for-review" {
		t.Errorf("String() = %q, want %q", id.String(), "submit-for-review")
	}
}

func TestTransitionIDJSONRoundTrip(t *testing.T) {
	original, err := NewTransitionID("submit-for-review")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded TransitionID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Equal(original) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestTransitionIDMalformedJSONRejected(t *testing.T) {
	var id TransitionID
	if err := json.Unmarshal([]byte(`{}`), &id); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
}

func TestTransitionIDZeroMarshalRejected(t *testing.T) {
	var id TransitionID
	if _, err := json.Marshal(id); err == nil {
		t.Fatal("Marshal(zero TransitionID) accepted, want error")
	}
}

func TestTransitionIDFailedUnmarshalPreservesReceiver(t *testing.T) {
	original, err := NewTransitionID("submit-for-review")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`null`), &receiver); err == nil {
		t.Fatal("null accepted, want error")
	}
	if !receiver.Equal(original) {
		t.Errorf("failed Unmarshal changed receiver: got %v, want %v", receiver, original)
	}
}

// TestStateIDAndTransitionIDAreNotInterchangeable documents, at compile
// time, that StateID and TransitionID -- despite both wrapping a scoped
// local key with an identical underlying shape -- are distinct Go types.
func TestStateIDAndTransitionIDAreNotInterchangeable(t *testing.T) {
	stateID, err := NewStateID("draft")
	if err != nil {
		t.Fatal(err)
	}
	transitionID, err := NewTransitionID("draft")
	if err != nil {
		t.Fatal(err)
	}
	// The following, if uncommented, must fail to compile:
	//   var _ StateID = transitionID
	//   var _ TransitionID = stateID
	if stateID.String() != transitionID.String() {
		t.Fatal("expected identical opaque values for this compile-time-only assertion")
	}
}
