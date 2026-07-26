package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewUncertaintyValid(t *testing.T) {
	u, err := NewUncertainty("vendor pricing after 2027 is unknown")
	if err != nil {
		t.Fatal(err)
	}
	if u.Statement() != "vendor pricing after 2027 is unknown" {
		t.Errorf("Statement() = %q, want %q", u.Statement(), "vendor pricing after 2027 is unknown")
	}
}

func TestNewUncertaintyEmptyStatementRejected(t *testing.T) {
	if _, err := NewUncertainty(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewUncertaintyWhitespaceOnlyStatementRejected(t *testing.T) {
	if _, err := NewUncertainty("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestUncertaintyWithExtension(t *testing.T) {
	u, err := NewUncertainty("statement")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := u.WithExtension(ext)
	if !u.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestUncertaintyImmutability(t *testing.T) {
	u, err := NewUncertainty("statement")
	if err != nil {
		t.Fatal(err)
	}
	original := u
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = u.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

func TestUncertaintyIsZero(t *testing.T) {
	var u Uncertainty
	if !u.IsZero() {
		t.Error("zero Uncertainty IsZero() = false")
	}
	valid, err := NewUncertainty("statement")
	if err != nil {
		t.Fatal(err)
	}
	if valid.IsZero() {
		t.Error("valid Uncertainty IsZero() = true")
	}
}

func fullUncertainty(t *testing.T) Uncertainty {
	t.Helper()
	u, err := NewUncertainty("vendor pricing after 2027 is unknown")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return u.WithExtension(ext)
}

func TestUncertaintyJSONContract(t *testing.T) {
	u := fullUncertainty(t)
	data, err := json.Marshal(u)
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
	var decoded Uncertainty
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != u.Statement() {
		t.Error("Statement mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestUncertaintyJSONMinimumOmitsExtension(t *testing.T) {
	u, err := NewUncertainty("statement")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(u)
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

func TestUncertaintyNullExtensionRejected(t *testing.T) {
	var u Uncertainty
	if err := json.Unmarshal([]byte(`{"statement":"s","extension":null}`), &u); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestUncertaintyZeroMarshal(t *testing.T) {
	var u Uncertainty
	if _, err := json.Marshal(u); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestUncertaintyReceiverPreservation(t *testing.T) {
	original := fullUncertainty(t)
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

func TestUncertaintyJSONUnknownFieldIgnored(t *testing.T) {
	var u Uncertainty
	if err := json.Unmarshal([]byte(`{"statement":"s","unknown_field":123}`), &u); err != nil {
		t.Fatal(err)
	}
}

// TestUncertaintyNoConcernOrSeverityVocabulary asserts the deliberate
// absence of any concern or severity classification (F2R-09 / R4):
// Uncertainty's wire form contains exactly two keys when fully populated
// (statement, extension) -- no concern, no severity, no significance, no
// threshold, no scope, no source, no resolution condition.
func TestUncertaintyNoConcernOrSeverityVocabulary(t *testing.T) {
	u := fullUncertainty(t)
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"concern", "severity", "significance", "threshold", "scope", "source", "resolution_condition"} {
		if _, present := raw[forbidden]; present {
			t.Errorf("unexpected %q key present in Uncertainty wire form", forbidden)
		}
	}
	if len(raw) != 2 {
		t.Errorf("Uncertainty wire form has %d keys, want exactly 2 (statement, extension): %v", len(raw), raw)
	}
}
