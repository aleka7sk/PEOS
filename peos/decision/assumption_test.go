package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewAssumptionValid(t *testing.T) {
	a, err := NewAssumption("traffic stays under 1k rps")
	if err != nil {
		t.Fatal(err)
	}
	if a.Statement() != "traffic stays under 1k rps" {
		t.Errorf("Statement() = %q, want %q", a.Statement(), "traffic stays under 1k rps")
	}
}

func TestNewAssumptionEmptyStatementRejected(t *testing.T) {
	if _, err := NewAssumption(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewAssumptionWhitespaceOnlyStatementRejected(t *testing.T) {
	if _, err := NewAssumption("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewAssumptionRawStatementPreserved(t *testing.T) {
	a, err := NewAssumption("  padded statement  ")
	if err != nil {
		t.Fatal(err)
	}
	if a.Statement() != "  padded statement  " {
		t.Errorf("Statement() = %q, want the original padded value preserved", a.Statement())
	}
}

func mustBasisScope(t *testing.T, expression string) core.Scope {
	t.Helper()
	kind, err := core.NewVocabularyValue("product-x", "path")
	if err != nil {
		t.Fatal(err)
	}
	scope, err := core.NewScope(kind, expression)
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func fullAssumption(t *testing.T) Assumption {
	t.Helper()
	a, err := NewAssumption("traffic stays under 1k rps")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithScope(mustBasisScope(t, "/services/billing"))
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithSource("2026-Q2 capacity study")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithUncertainty("based on a single quarter of data")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithExpectedValidationCondition("re-measure after Q4 peak")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithConsequenceIfFalse("billing tier must be resized")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return a.WithExtension(ext)
}

func TestAssumptionAllFiveOptionalFields(t *testing.T) {
	a := fullAssumption(t)

	scope, ok := a.Scope()
	if !ok || !scope.Equal(mustBasisScope(t, "/services/billing")) {
		t.Errorf("Scope() = (%v,%v)", scope, ok)
	}
	source, ok := a.Source()
	if !ok || source != "2026-Q2 capacity study" {
		t.Errorf("Source() = (%q,%v)", source, ok)
	}
	uncertainty, ok := a.Uncertainty()
	if !ok || uncertainty != "based on a single quarter of data" {
		t.Errorf("Uncertainty() = (%q,%v)", uncertainty, ok)
	}
	cond, ok := a.ExpectedValidationCondition()
	if !ok || cond != "re-measure after Q4 peak" {
		t.Errorf("ExpectedValidationCondition() = (%q,%v)", cond, ok)
	}
	consequence, ok := a.ConsequenceIfFalse()
	if !ok || consequence != "billing tier must be resized" {
		t.Errorf("ConsequenceIfFalse() = (%q,%v)", consequence, ok)
	}
}

func TestAssumptionOptionalFieldsAbsentByDefault(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Scope(); ok {
		t.Error("Scope() ok = true before WithScope")
	}
	if _, ok := a.Source(); ok {
		t.Error("Source() ok = true before WithSource")
	}
	if _, ok := a.Uncertainty(); ok {
		t.Error("Uncertainty() ok = true before WithUncertainty")
	}
	if _, ok := a.ExpectedValidationCondition(); ok {
		t.Error("ExpectedValidationCondition() ok = true before With")
	}
	if _, ok := a.ConsequenceIfFalse(); ok {
		t.Error("ConsequenceIfFalse() ok = true before With")
	}
}

func TestAssumptionZeroScopeRejected(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.WithScope(core.Scope{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestAssumptionEmptyOptionalStringsRejected(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.WithSource(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithSource empty: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithSource("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithSource whitespace: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithUncertainty(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithUncertainty empty: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithUncertainty("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithUncertainty whitespace: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithExpectedValidationCondition(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithExpectedValidationCondition empty: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithExpectedValidationCondition("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithExpectedValidationCondition whitespace: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithConsequenceIfFalse(""); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithConsequenceIfFalse empty: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := a.WithConsequenceIfFalse("   "); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithConsequenceIfFalse whitespace: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestAssumptionWithoutMethods(t *testing.T) {
	a := fullAssumption(t)

	if _, ok := a.WithoutScope().Scope(); ok {
		t.Error("Scope() ok = true after WithoutScope")
	}
	if _, ok := a.WithoutSource().Source(); ok {
		t.Error("Source() ok = true after WithoutSource")
	}
	if _, ok := a.WithoutUncertainty().Uncertainty(); ok {
		t.Error("Uncertainty() ok = true after WithoutUncertainty")
	}
	if _, ok := a.WithoutExpectedValidationCondition().ExpectedValidationCondition(); ok {
		t.Error("ExpectedValidationCondition() ok = true after Without")
	}
	if _, ok := a.WithoutConsequenceIfFalse().ConsequenceIfFalse(); ok {
		t.Error("ConsequenceIfFalse() ok = true after Without")
	}

	// original receiver must remain fully populated
	if _, ok := a.Scope(); !ok {
		t.Error("WithoutScope mutated the original receiver")
	}
}

func TestAssumptionWithExtension(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := a.WithExtension(ext)
	if !a.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestAssumptionImmutability(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	original := a
	if _, err := a.WithSource("some source"); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Source(); ok {
		t.Error("WithSource mutated the original receiver")
	}
}

func TestAssumptionIsZero(t *testing.T) {
	var a Assumption
	if !a.IsZero() {
		t.Error("zero Assumption IsZero() = false")
	}
	valid, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	if valid.IsZero() {
		t.Error("valid Assumption IsZero() = true")
	}
}

func TestAssumptionJSONMinimumRoundTrip(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Assumption
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != "statement" {
		t.Errorf("Statement() = %q, want %q", decoded.Statement(), "statement")
	}
}

func TestAssumptionJSONFullRoundTrip(t *testing.T) {
	a := fullAssumption(t)
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Assumption
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != a.Statement() {
		t.Errorf("Statement mismatch")
	}
	if _, ok := decoded.Scope(); !ok {
		t.Error("Scope absent after round trip")
	}
	if _, ok := decoded.Source(); !ok {
		t.Error("Source absent after round trip")
	}
	if _, ok := decoded.Uncertainty(); !ok {
		t.Error("Uncertainty absent after round trip")
	}
	if _, ok := decoded.ExpectedValidationCondition(); !ok {
		t.Error("ExpectedValidationCondition absent after round trip")
	}
	if _, ok := decoded.ConsequenceIfFalse(); !ok {
		t.Error("ConsequenceIfFalse absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestAssumptionJSONLiteralWireKeys(t *testing.T) {
	a := fullAssumption(t)
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"statement", "scope", "source", "uncertainty",
		"expected_validation_condition", "consequence_if_false", "extension",
	} {
		if _, present := raw[key]; !present {
			t.Errorf("required wire key %q missing", key)
		}
	}
}

func TestAssumptionJSONOptionalOmission(t *testing.T) {
	a, err := NewAssumption("statement")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"scope", "source", "uncertainty",
		"expected_validation_condition", "consequence_if_false", "extension",
	} {
		if _, present := raw[key]; present {
			t.Errorf("optional key %q present despite not being set", key)
		}
	}
}

func TestAssumptionExplicitNullRejectedPerOptionalKey(t *testing.T) {
	base := `{"statement":"s"`
	fields := []string{"scope", "source", "uncertainty", "expected_validation_condition", "consequence_if_false", "extension"}
	for _, field := range fields {
		payload := base + `,"` + field + `":null}`
		var a Assumption
		if err := json.Unmarshal([]byte(payload), &a); err == nil {
			t.Errorf("field %q: explicit null accepted, want error", field)
		}
	}
}

func TestAssumptionEmptyStringPerOptionalKeyRejected(t *testing.T) {
	base := `{"statement":"s"`
	fields := []string{"source", "uncertainty", "expected_validation_condition", "consequence_if_false"}
	for _, field := range fields {
		payload := base + `,"` + field + `":""}`
		var a Assumption
		if err := json.Unmarshal([]byte(payload), &a); err == nil {
			t.Errorf("field %q: empty string accepted, want error", field)
		}
	}
}

func TestAssumptionMalformedNestedScopePreservesBothSentinels(t *testing.T) {
	var a Assumption
	// Valid vocabulary kind but empty expression: triggers core.Scope's own
	// ErrInvalidScope specifically (as opposed to a vocabulary-parse error).
	payload := `{"statement":"s","scope":{"kind":"product-x:path","expression":""}}`
	err := json.Unmarshal([]byte(payload), &a)
	if err == nil {
		t.Fatal("malformed scope accepted, want error")
	}
	if !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidBasis)
	}
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want also wrapping %v", err, core.ErrInvalidScope)
	}
}

func TestAssumptionJSONUnknownFieldIgnored(t *testing.T) {
	var a Assumption
	if err := json.Unmarshal([]byte(`{"statement":"s","unknown_field":123}`), &a); err != nil {
		t.Fatal(err)
	}
}

func TestAssumptionZeroMarshalRejected(t *testing.T) {
	var a Assumption
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestAssumptionUnmarshalFailurePreservesPopulatedReceiver(t *testing.T) {
	original := fullAssumption(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"statement":""}`), &receiver); err == nil {
		t.Fatal("empty statement accepted, want error")
	}
	if receiver.Statement() != original.Statement() {
		t.Error("failed Unmarshal changed receiver's statement")
	}
	if _, ok := receiver.Scope(); !ok {
		t.Error("failed Unmarshal changed receiver's scope presence")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}
