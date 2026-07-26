package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewConstraintValid(t *testing.T) {
	c, err := NewConstraint("must remain EU-resident")
	if err != nil {
		t.Fatal(err)
	}
	if c.Statement() != "must remain EU-resident" {
		t.Errorf("Statement() = %q, want %q", c.Statement(), "must remain EU-resident")
	}
}

func TestNewConstraintEmptyStatementRejected(t *testing.T) {
	if _, err := NewConstraint(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewConstraintWhitespaceOnlyStatementRejected(t *testing.T) {
	if _, err := NewConstraint("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestConstraintSourcePresentAbsent(t *testing.T) {
	c, err := NewConstraint("must remain EU-resident")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Source(); ok {
		t.Error("Source() ok = true before WithSource")
	}
	withSource, err := c.WithSource("GDPR Article 44")
	if err != nil {
		t.Fatal(err)
	}
	source, ok := withSource.Source()
	if !ok || source != "GDPR Article 44" {
		t.Errorf("Source() = (%q,%v)", source, ok)
	}
	if _, ok := c.Source(); ok {
		t.Error("WithSource mutated the original receiver")
	}
}

func TestConstraintEmptyWhitespaceSourceRejected(t *testing.T) {
	c, err := NewConstraint("statement")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.WithSource(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("empty: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := c.WithSource("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("whitespace: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestConstraintWithoutSource(t *testing.T) {
	c, err := NewConstraint("statement")
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithSource("some source")
	if err != nil {
		t.Fatal(err)
	}
	cleared := c.WithoutSource()
	if _, ok := cleared.Source(); ok {
		t.Error("Source() ok = true after WithoutSource")
	}
	if _, ok := c.Source(); !ok {
		t.Error("WithoutSource mutated the original receiver")
	}
}

func TestConstraintWithExtension(t *testing.T) {
	c, err := NewConstraint("statement")
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

func TestConstraintImmutability(t *testing.T) {
	c, err := NewConstraint("statement")
	if err != nil {
		t.Fatal(err)
	}
	original := c
	if _, err := c.WithSource("source"); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Source(); ok {
		t.Error("WithSource mutated the original receiver")
	}
}

func TestConstraintIsZero(t *testing.T) {
	var c Constraint
	if !c.IsZero() {
		t.Error("zero Constraint IsZero() = false")
	}
	valid, err := NewConstraint("statement")
	if err != nil {
		t.Fatal(err)
	}
	if valid.IsZero() {
		t.Error("valid Constraint IsZero() = true")
	}
}

func fullConstraint(t *testing.T) Constraint {
	t.Helper()
	c, err := NewConstraint("must remain EU-resident")
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithSource("GDPR Article 44")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return c.WithExtension(ext)
}

func TestConstraintJSONContract(t *testing.T) {
	c := fullConstraint(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"statement", "source", "extension"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
	var decoded Constraint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != c.Statement() {
		t.Error("Statement mismatch")
	}
	if _, ok := decoded.Source(); !ok {
		t.Error("Source absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestConstraintJSONMinimumOmitsOptionalFields(t *testing.T) {
	c, err := NewConstraint("statement")
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
	for _, key := range []string{"source", "extension"} {
		if _, present := raw[key]; present {
			t.Errorf("optional key %q present despite not being set", key)
		}
	}
}

func TestConstraintNullRejection(t *testing.T) {
	base := `{"statement":"s"`
	if err := json.Unmarshal([]byte(base+`,"source":null}`), new(Constraint)); err == nil {
		t.Error("null source accepted, want error")
	}
	if err := json.Unmarshal([]byte(base+`,"extension":null}`), new(Constraint)); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestConstraintZeroMarshal(t *testing.T) {
	var c Constraint
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestConstraintReceiverPreservation(t *testing.T) {
	original := fullConstraint(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"statement":""}`), &receiver); err == nil {
		t.Fatal("empty statement accepted, want error")
	}
	if receiver.Statement() != original.Statement() {
		t.Error("failed Unmarshal changed receiver")
	}
	if _, ok := receiver.Source(); !ok {
		t.Error("failed Unmarshal changed receiver's source")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

func TestConstraintJSONUnknownFieldIgnored(t *testing.T) {
	var c Constraint
	if err := json.Unmarshal([]byte(`{"statement":"s","unknown_field":123}`), &c); err != nil {
		t.Fatal(err)
	}
}

// TestConstraintNoOriginVocabulary asserts the deliberate absence of any
// origin-category vocabulary (F2R-08 / R4): Constraint's own doc comment
// and API contain no ConstraintOrigin type, no origin field, and no
// predeclared origin constants -- only statement, source, and extension.
func TestConstraintNoOriginVocabulary(t *testing.T) {
	c := fullConstraint(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["origin"]; present {
		t.Error(`unexpected "origin" key present in Constraint wire form`)
	}
	if len(raw) != 3 {
		t.Errorf("Constraint wire form has %d keys, want exactly 3 (statement, source, extension): %v", len(raw), raw)
	}
}
