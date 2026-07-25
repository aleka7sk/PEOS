package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestExtensionWithAndGet(t *testing.T) {
	ext := NewExtension()
	if !ext.IsZero() {
		t.Error("new Extension is not zero")
	}

	ext, err := ext.With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if ext.IsZero() {
		t.Error("Extension with an entry reports IsZero() = true")
	}

	got, ok := ext.Get("product-x")
	if !ok {
		t.Fatal("Get(\"product-x\") not found")
	}
	if string(got) != `{"a":1}` {
		t.Errorf("Get(\"product-x\") = %s, want %s", got, `{"a":1}`)
	}

	if _, ok := ext.Get("missing"); ok {
		t.Error("Get(\"missing\") found, want not found")
	}
}

func TestExtensionDuplicateNamespaceRejected(t *testing.T) {
	ext, err := NewExtension().With("product-x", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	_, err = ext.With("product-x", json.RawMessage(`{"different":true}`))
	if !errors.Is(err, ErrDuplicateExtensionNamespace) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateExtensionNamespace)
	}
}

func TestExtensionDefensiveCopyOnInput(t *testing.T) {
	payload := json.RawMessage(`{"a":1}`)
	ext, err := NewExtension().With("product-x", payload)
	if err != nil {
		t.Fatal(err)
	}
	// Mutate the caller's slice after construction; the stored copy must
	// be unaffected.
	for i := range payload {
		payload[i] = 'X'
	}
	got, _ := ext.Get("product-x")
	if string(got) != `{"a":1}` {
		t.Errorf("stored payload was affected by caller mutation: got %s", got)
	}
}

func TestExtensionDefensiveCopyOnOutput(t *testing.T) {
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := ext.Get("product-x")
	for i := range got {
		got[i] = 'X'
	}
	again, _ := ext.Get("product-x")
	if string(again) != `{"a":1}` {
		t.Errorf("internal state was affected by mutating a Get() result: got %s", again)
	}
}

func TestExtensionRejectsEmptyOrInvalidPayload(t *testing.T) {
	if _, err := NewExtension().With("product-x", nil); err == nil {
		t.Error("With(nil) succeeded, want error")
	}
	if _, err := NewExtension().With("product-x", json.RawMessage(`not json`)); err == nil {
		t.Error("With(invalid JSON) succeeded, want error")
	}
	if _, err := NewExtension().With("", json.RawMessage(`{}`)); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("With(empty namespace) error = %v, want %v", err, ErrEmptyIdentity)
	}
}

func TestExtensionJSONPreservation(t *testing.T) {
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	ext, err = ext.With("product-y", json.RawMessage(`[1,2,3]`))
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(ext)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Extension
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if got, ok := decoded.Get("product-x"); !ok || string(got) != `{"a":1}` {
		t.Errorf("product-x = %s, ok=%v", got, ok)
	}
	if got, ok := decoded.Get("product-y"); !ok || string(got) != `[1,2,3]` {
		t.Errorf("product-y = %s, ok=%v", got, ok)
	}
}

func TestExtensionNamespacesSorted(t *testing.T) {
	ext, err := NewExtension().With("zeta", json.RawMessage(`1`))
	if err != nil {
		t.Fatal(err)
	}
	ext, err = ext.With("alpha", json.RawMessage(`1`))
	if err != nil {
		t.Fatal(err)
	}
	got := ext.Namespaces()
	if len(got) != 2 || got[0] != "alpha" || got[1] != "zeta" {
		t.Errorf("Namespaces() = %v, want [alpha zeta]", got)
	}
}

func TestExtensionEmptyKeyRejectedOnUnmarshal(t *testing.T) {
	var ext Extension
	if err := json.Unmarshal([]byte(`{"":{}}`), &ext); err == nil {
		t.Error("Unmarshal with empty namespace key succeeded, want error")
	}
}
