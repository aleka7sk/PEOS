package runtime

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// mustCriterionRef builds a valid core.CriterionRef citing a Requirement,
// for use as an Assertion's criterion in tests that do not care which
// criterion kind is used.
func mustCriterionRef(t *testing.T, requirementID string) core.CriterionRef {
	t.Helper()
	ref, err := core.CriterionRefFromRequirement(mustRequirementRef(t, requirementID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// mustAssertion builds a minimal valid Assertion with the given key.
func mustAssertion(t *testing.T, key string) Assertion {
	t.Helper()
	a, err := NewAssertion(
		mustLocalKey(t, key),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustCriterionRef(t, "REQ-1"),
		"observed p99 latency < 200ms",
		"true",
		mustScope(t, "cluster=prod-1"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// --- Environment ---------------------------------------------------------------

func TestEnvironment(t *testing.T) {
	v := mustVocabularyValue(t, "product", "production")
	e := NewEnvironment(v)
	if e.IsZero() {
		t.Error("valid Environment reports IsZero() = true")
	}
	if e.Value() != v {
		t.Error("Value() mismatch")
	}
	if e.String() != v.String() {
		t.Error("String() mismatch")
	}
	if !e.Equal(NewEnvironment(v)) {
		t.Error("Equal() should be true for identical values")
	}
	other := NewEnvironment(mustVocabularyValue(t, "product", "staging"))
	if e.Equal(other) {
		t.Error("Equal() should be false for different values")
	}

	var zero Environment
	if !zero.IsZero() {
		t.Error("zero-value Environment.IsZero() = false, want true")
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Environment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Value() != e.Value() {
		t.Error("round trip mismatch")
	}
}

// --- ViolationClassification / ViolationSeverity -----------------------------

func TestViolationClassification(t *testing.T) {
	v := mustVocabularyValue(t, "product", "breach")
	c := NewViolationClassification(v)
	if c.IsZero() {
		t.Error("valid ViolationClassification reports IsZero() = true")
	}
	if c.Value() != v {
		t.Error("Value() mismatch")
	}
	if c.String() != v.String() {
		t.Error("String() mismatch")
	}
	if !c.Equal(NewViolationClassification(v)) {
		t.Error("Equal() should be true for identical values")
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ViolationClassification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Value() != c.Value() {
		t.Error("round trip mismatch")
	}
}

func TestViolationSeverity(t *testing.T) {
	v := mustVocabularyValue(t, "product", "critical")
	s := NewViolationSeverity(v)
	if s.IsZero() {
		t.Error("valid ViolationSeverity reports IsZero() = true")
	}
	if s.Value() != v {
		t.Error("Value() mismatch")
	}
	if s.String() != v.String() {
		t.Error("String() mismatch")
	}
	if !s.Equal(NewViolationSeverity(v)) {
		t.Error("Equal() should be true for identical values")
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ViolationSeverity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Value() != s.Value() {
		t.Error("round trip mismatch")
	}
}

// --- Assertion -----------------------------------------------------------------

func TestNewAssertion(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	if a.IsZero() {
		t.Error("valid Assertion reports IsZero() = true")
	}
	if a.Key() != mustLocalKey(t, "assert-1") {
		t.Error("Key() mismatch")
	}
	if a.EvaluationRule() != "observed p99 latency < 200ms" {
		t.Errorf("EvaluationRule() = %q", a.EvaluationRule())
	}
	if a.ExpectedResult() != "true" {
		t.Errorf("ExpectedResult() = %q", a.ExpectedResult())
	}
	if a.Subject() != mustRuntimeSubjectRef(t, "kubernetes", "pod-1") {
		t.Error("Subject() mismatch")
	}
	if a.Criterion() != mustCriterionRef(t, "REQ-1") {
		t.Error("Criterion() mismatch")
	}
	if a.Scope() != mustScope(t, "cluster=prod-1") {
		t.Error("Scope() mismatch")
	}
	if len(a.ObservationInputs()) != 0 {
		t.Error("new Assertion should have no observation inputs")
	}
	if _, ok := a.TemporalConditions(); ok {
		t.Error("new Assertion should have no temporal conditions")
	}
	if _, ok := a.UncertaintyHandling(); ok {
		t.Error("new Assertion should have no uncertainty handling")
	}
}

func TestNewAssertionMandatoryFieldRejections(t *testing.T) {
	key := mustLocalKey(t, "assert-1")
	subject := mustRuntimeSubjectRef(t, "kubernetes", "pod-1")
	criterion := mustCriterionRef(t, "REQ-1")
	rule := "rule"
	result := "true"
	scope := mustScope(t, "cluster=prod-1")

	if _, err := NewAssertion(core.LocalKey{}, subject, criterion, rule, result, scope); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("zero key: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
	if _, err := NewAssertion(key, core.RuntimeSubjectRef{}, criterion, rule, result, scope); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("zero subject: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
	if _, err := NewAssertion(key, subject, core.CriterionRef{}, rule, result, scope); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("zero criterion: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
	if _, err := NewAssertion(key, subject, criterion, "   ", result, scope); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("whitespace-only evaluation rule: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
	if _, err := NewAssertion(key, subject, criterion, rule, "", scope); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("empty expected result: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
	if _, err := NewAssertion(key, subject, criterion, rule, result, core.Scope{}); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestAssertionWithObservationInputs(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	a2, err := a.WithObservationInputs([]string{"p99 latency sample", "error rate sample"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.ObservationInputs()) != 0 {
		t.Error("original Assertion mutated by WithObservationInputs")
	}
	got := a2.ObservationInputs()
	if len(got) != 2 || got[0] != "p99 latency sample" {
		t.Errorf("ObservationInputs() = %v", got)
	}
	got[0] = "mutated"
	if a2.ObservationInputs()[0] == "mutated" {
		t.Error("ObservationInputs() accessor did not return a defensive copy")
	}

	if _, err := a.WithObservationInputs([]string{"  "}); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("whitespace-only entry: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}

	cleared, err := a2.WithObservationInputs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.ObservationInputs()) != 0 {
		t.Error("WithObservationInputs(nil) did not clear the collection")
	}
}

func TestAssertionWithTemporalConditions(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	a2, err := a.WithTemporalConditions("sustained for 5 minutes")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := a2.TemporalConditions()
	if !ok || got != "sustained for 5 minutes" {
		t.Errorf("TemporalConditions() = (%q, %v)", got, ok)
	}
	if _, ok := a.TemporalConditions(); ok {
		t.Error("original Assertion mutated by WithTemporalConditions")
	}

	if _, err := a.WithTemporalConditions("   "); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("whitespace-only: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}

	cleared := a2.WithoutTemporalConditions()
	if _, ok := cleared.TemporalConditions(); ok {
		t.Error("WithoutTemporalConditions did not clear the field")
	}
}

func TestAssertionWithUncertaintyHandling(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	a2, err := a.WithUncertaintyHandling("median of 3 samples")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := a2.UncertaintyHandling()
	if !ok || got != "median of 3 samples" {
		t.Errorf("UncertaintyHandling() = (%q, %v)", got, ok)
	}
	if _, ok := a.UncertaintyHandling(); ok {
		t.Error("original Assertion mutated by WithUncertaintyHandling")
	}

	if _, err := a.WithUncertaintyHandling(""); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("empty: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}

	cleared := a2.WithoutUncertaintyHandling()
	if _, ok := cleared.UncertaintyHandling(); ok {
		t.Error("WithoutUncertaintyHandling did not clear the field")
	}
}

func TestAssertionExtension(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	a2 := a.WithExtension(ext)
	if a2.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !a.Extension().IsZero() {
		t.Error("original Assertion mutated by WithExtension")
	}
	a3 := a2.WithoutExtension()
	if !a3.Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestAssertionJSONRoundTrip(t *testing.T) {
	a, err := mustAssertion(t, "assert-1").WithObservationInputs([]string{"sample"})
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithTemporalConditions("sustained for 5 minutes")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithUncertaintyHandling("median of 3 samples")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	a = a.WithExtension(ext)

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Assertion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Extension().IsZero() {
		t.Error("decoded Assertion lost its extension")
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
}

func TestAssertionMarshalZero(t *testing.T) {
	var a Assertion
	if !a.IsZero() {
		t.Error("zero-value Assertion.IsZero() = false, want true")
	}
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
}

func TestAssertionUnmarshalRejectsExplicitNullOptionalFields(t *testing.T) {
	base := mustAssertion(t, "assert-1")
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	m["temporal_conditions"] = json.RawMessage("null")
	withNullTemporal, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Assertion
	if err := json.Unmarshal(withNullTemporal, &decoded); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("null temporal_conditions: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
	delete(m, "temporal_conditions")

	m["uncertainty_handling"] = json.RawMessage("null")
	withNullUncertainty, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(withNullUncertainty, &decoded); !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("null uncertainty_handling: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
}

func TestAssertionUnmarshalFieldSpecificRejections(t *testing.T) {
	base := mustAssertion(t, "assert-1")
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		key   string
		value json.RawMessage
	}{
		{"whitespace-only observation input", "observation_inputs", json.RawMessage(`["  "]`)},
		{"malformed temporal_conditions type", "temporal_conditions", json.RawMessage(`123`)},
		{"whitespace-only temporal_conditions", "temporal_conditions", json.RawMessage(`"   "`)},
		{"malformed uncertainty_handling type", "uncertainty_handling", json.RawMessage(`123`)},
		{"empty uncertainty_handling", "uncertainty_handling", json.RawMessage(`""`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			mCopy[tt.key] = tt.value
			modified, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var a Assertion
			if err := json.Unmarshal(modified, &a); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}
}

func TestAssertionUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	originalData, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"key":"assert-1"}`), &a); err == nil {
		t.Fatal("incomplete assertion accepted, want error")
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
	if err := json.Unmarshal([]byte(`not json`), &a); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
}

func TestAssertionNoForbiddenWireKeys(t *testing.T) {
	a := mustAssertion(t, "assert-1")
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"outcome", "result", "satisfied", "violated", "status", "state"}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range forbidden {
		if _, ok := m[key]; ok {
			t.Errorf("wire form contains forbidden key %q", key)
		}
	}
}
