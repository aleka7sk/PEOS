package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewScope(t *testing.T) {
	kind := mustVocabularyValue(t, "peos", "artifact-path")

	if _, err := NewScope(VocabularyValue{}, "some-expression"); !errors.Is(err, ErrInvalidScope) {
		t.Errorf("missing kind: error = %v, want %v", err, ErrInvalidScope)
	}
	if _, err := NewScope(kind, ""); !errors.Is(err, ErrInvalidScope) {
		t.Errorf("missing expression: error = %v, want %v", err, ErrInvalidScope)
	}

	s, err := NewScope(kind, "requirements/*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.IsZero() {
		t.Error("valid Scope reports IsZero() = true")
	}
	if s.Expression() != "requirements/*" {
		t.Errorf("Expression() = %q, want %q", s.Expression(), "requirements/*")
	}
}

func TestScopeEqual(t *testing.T) {
	kind := mustVocabularyValue(t, "peos", "artifact-path")
	a, err := NewScope(kind, "x")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewScope(kind, "x")
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewScope(kind, "y")
	if err != nil {
		t.Fatal(err)
	}
	if !a.Equal(b) {
		t.Error("identical scopes not reported equal")
	}
	if a.Equal(c) {
		t.Error("differing scopes reported equal")
	}
}

func TestScopeJSONRoundTrip(t *testing.T) {
	kind := mustVocabularyValue(t, "peos", "artifact-path")
	original, err := NewScope(kind, "requirements/*")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Scope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestScopeJSONRejectsMissingFields(t *testing.T) {
	if err := json.Unmarshal([]byte(`{"kind":"peos:artifact-path","expression":""}`), &Scope{}); err == nil {
		t.Error("Unmarshal with empty expression succeeded, want error")
	}
	if err := json.Unmarshal([]byte(`{"expression":"x"}`), &Scope{}); err == nil {
		t.Error("Unmarshal with missing kind succeeded, want error")
	}
}
