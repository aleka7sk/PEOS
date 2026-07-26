package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func fullConsequence(t *testing.T) Consequence {
	t.Helper()
	c, err := NewConsequence("Migration of existing services is expected.")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return c.WithExtension(ext)
}

func TestNewConsequenceValid(t *testing.T) {
	c, err := NewConsequence("Migration is expected.")
	if err != nil {
		t.Fatal(err)
	}
	if c.Statement() != "Migration is expected." {
		t.Errorf("Statement() = %q, want %q", c.Statement(), "Migration is expected.")
	}
}

func TestNewConsequenceEmptyStatementRejected(t *testing.T) {
	if _, err := NewConsequence(""); !errors.Is(err, ErrInvalidConsequence) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConsequence)
	}
}

func TestNewConsequenceWhitespaceOnlyStatementRejected(t *testing.T) {
	if _, err := NewConsequence("   "); !errors.Is(err, ErrInvalidConsequence) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConsequence)
	}
}

func TestNewConsequenceRawStatementPreserved(t *testing.T) {
	c, err := NewConsequence("  padded  ")
	if err != nil {
		t.Fatal(err)
	}
	if c.Statement() != "  padded  " {
		t.Errorf("Statement() = %q, want raw preserved", c.Statement())
	}
}

func TestConsequenceWithExtension(t *testing.T) {
	c, err := NewConsequence("s")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := c.WithExtension(ext)
	if !c.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestConsequenceImmutability(t *testing.T) {
	c, err := NewConsequence("s")
	if err != nil {
		t.Fatal(err)
	}
	original := c
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

func TestConsequenceIsZero(t *testing.T) {
	var c Consequence
	if !c.IsZero() {
		t.Error("zero Consequence IsZero() = false")
	}
	if fullConsequence(t).IsZero() {
		t.Error("valid Consequence IsZero() = true")
	}
}

func TestConsequenceJSONLiteralWireKeys(t *testing.T) {
	c := fullConsequence(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"statement", "extension"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
}

func TestConsequenceJSONMinimumOmitsExtension(t *testing.T) {
	c, err := NewConsequence("s")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["extension"]; present {
		t.Error("extension present despite not being set")
	}
}

func TestConsequenceJSONRoundTrip(t *testing.T) {
	c := fullConsequence(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Consequence
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != c.Statement() {
		t.Error("Statement mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestConsequenceJSONNullExtensionRejected(t *testing.T) {
	var c Consequence
	if err := json.Unmarshal([]byte(`{"statement":"s","extension":null}`), &c); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestConsequenceJSONUnknownFieldIgnored(t *testing.T) {
	var c Consequence
	if err := json.Unmarshal([]byte(`{"statement":"s","unknown_field":123}`), &c); err != nil {
		t.Fatal(err)
	}
}

func TestConsequenceZeroMarshalRejected(t *testing.T) {
	var c Consequence
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidConsequence) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConsequence)
	}
}

func TestConsequenceUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullConsequence(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"statement":""}`), &receiver); err == nil {
		t.Fatal("empty statement accepted, want error")
	}
	if receiver.Statement() != original.Statement() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// TestConsequenceNoCompletionField structurally asserts the absence of
// any completion/status field, proving PEOS-004 :697's "Consequences
// MUST be distinguishable from completed effects" holds structurally:
// the wire form has exactly two keys when fully populated.
func TestConsequenceNoCompletionField(t *testing.T) {
	c := fullConsequence(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"completed", "status", "complete", "done", "resolved"} {
		if _, present := raw[forbidden]; present {
			t.Errorf("unexpected %q key present in Consequence wire form", forbidden)
		}
	}
	if len(raw) != 2 {
		t.Errorf("Consequence wire form has %d keys, want exactly 2 (statement, extension): %v", len(raw), raw)
	}
}
