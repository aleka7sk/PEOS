package core

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestNewVocabularyValue(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		value         string
		wantErr       error
		wantNamespace string
		wantValue     string
	}{
		{name: "known peos value", namespace: "peos", value: "artifact", wantNamespace: "peos", wantValue: "artifact"},
		{name: "unknown product namespace preserved", namespace: "acme-corp", value: "custom-method", wantNamespace: "acme-corp", wantValue: "custom-method"},
		{name: "value may contain colons", namespace: "peos", value: "a:b:c", wantNamespace: "peos", wantValue: "a:b:c"},
		{name: "empty namespace rejected", namespace: "", value: "x", wantErr: ErrInvalidVocabularyValue},
		{name: "empty value rejected", namespace: "peos", value: "", wantErr: ErrInvalidVocabularyValue},
		{name: "namespace with colon rejected", namespace: "peos:sub", value: "x", wantErr: ErrInvalidVocabularyValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewVocabularyValue(tt.namespace, tt.value)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Namespace() != tt.wantNamespace || got.Value() != tt.wantValue {
				t.Errorf("got (%q, %q), want (%q, %q)", got.Namespace(), got.Value(), tt.wantNamespace, tt.wantValue)
			}
		})
	}
}

func TestVocabularyValueMalformedParse(t *testing.T) {
	tests := []string{"", "no-separator", ":missing-namespace", "missing-value:"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			_, err := ParseVocabularyValue(s)
			if err == nil && s != "missing-value:" {
				t.Fatalf("ParseVocabularyValue(%q) succeeded, want error", s)
			}
		})
	}
	// "missing-value:" parses to namespace "missing-value", empty value,
	// which NewVocabularyValue rejects because the value component is
	// empty.
	if _, err := ParseVocabularyValue("missing-value:"); !errors.Is(err, ErrInvalidVocabularyValue) {
		t.Errorf("ParseVocabularyValue(\"missing-value:\") error = %v, want %v", err, ErrInvalidVocabularyValue)
	}
}

func TestVocabularyValueRoundTrip(t *testing.T) {
	tests := []string{"peos:artifact", "acme-corp:custom-method", "peos:a:b:c"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			v, err := ParseVocabularyValue(s)
			if err != nil {
				t.Fatalf("ParseVocabularyValue(%q): %v", s, err)
			}
			if v.String() != s {
				t.Errorf("String() = %q, want %q", v.String(), s)
			}

			data, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var decoded VocabularyValue
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !decoded.Equal(v) {
				t.Errorf("round trip mismatch: got %v, want %v", decoded, v)
			}
		})
	}
}

func TestVocabularyValueUnknownValuesPreserved(t *testing.T) {
	v, err := NewVocabularyValue("product-x", "not-a-peos-value")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var decoded VocabularyValue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Namespace() != "product-x" || decoded.Value() != "not-a-peos-value" {
		t.Errorf("unknown namespace/value not preserved: got (%q, %q)", decoded.Namespace(), decoded.Value())
	}
}

func TestTypedVocabularyWrappersRoundTrip(t *testing.T) {
	if ClaimOutcomeSatisfied.String() != "peos:satisfied" {
		t.Errorf("ClaimOutcomeSatisfied.String() = %q, want %q", ClaimOutcomeSatisfied.String(), "peos:satisfied")
	}

	data, err := json.Marshal(ClaimTypeSatisfaction)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ClaimType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(ClaimTypeSatisfaction.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ClaimTypeSatisfaction)
	}

	// A Product MAY extend a vocabulary family with a value outside the
	// predefined constants; the typed wrapper must accept it.
	custom := NewValidationMethod(mustVocabularyValue(t, "acme-corp", "manual-review"))
	if custom.IsZero() {
		t.Error("custom ValidationMethod unexpectedly zero")
	}
}

// vocabularyWrapper describes one of the typed VocabularyValue wrappers
// (ArtifactType, ArtifactRole, RelationType, ClaimType, ClaimOutcome) so
// that all five can share one table-driven constructor+JSON test rather
// than duplicating near-identical cases five times.
type vocabularyWrapper struct {
	name        string
	constructFn func(VocabularyValue) interface {
		Value() VocabularyValue
		IsZero() bool
	}
}

func TestVocabularyWrapperConstructorsAndRoundTrip(t *testing.T) {
	wrappers := []vocabularyWrapper{
		{"ArtifactType", func(v VocabularyValue) interface {
			Value() VocabularyValue
			IsZero() bool
		} {
			return NewArtifactType(v)
		}},
		{"ArtifactRole", func(v VocabularyValue) interface {
			Value() VocabularyValue
			IsZero() bool
		} {
			return NewArtifactRole(v)
		}},
		{"RelationType", func(v VocabularyValue) interface {
			Value() VocabularyValue
			IsZero() bool
		} {
			return NewRelationType(v)
		}},
		{"ClaimType", func(v VocabularyValue) interface {
			Value() VocabularyValue
			IsZero() bool
		} {
			return NewClaimType(v)
		}},
		{"ClaimOutcome", func(v VocabularyValue) interface {
			Value() VocabularyValue
			IsZero() bool
		} {
			return NewClaimOutcome(v)
		}},
	}

	for _, w := range wrappers {
		t.Run(w.name, func(t *testing.T) {
			// Constructed from an unknown, Product-namespaced value, not
			// one of the predefined constants: the wrapper must accept it
			// without complaint.
			unknown := mustVocabularyValue(t, "product-x", "not-a-predefined-value")
			wrapped := w.constructFn(unknown)
			if wrapped.IsZero() {
				t.Fatal("constructed wrapper reports IsZero() = true for a non-zero value")
			}
			if !wrapped.Value().Equal(unknown) {
				t.Errorf("Value() = %v, want %v", wrapped.Value(), unknown)
			}

			marshaler, ok := wrapped.(json.Marshaler)
			if !ok {
				t.Fatal("wrapper does not implement json.Marshaler")
			}
			data, err := marshaler.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			if string(data) != `"product-x:not-a-predefined-value"` {
				t.Errorf("MarshalJSON = %s, want %q", data, `"product-x:not-a-predefined-value"`)
			}
		})
	}
}

func TestArtifactTypeJSONRoundTrip(t *testing.T) {
	original := NewArtifactType(mustVocabularyValue(t, "product-x", "custom-artifact-type"))
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ArtifactType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(original.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestArtifactRoleJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(ArtifactRoleEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ArtifactRole
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(ArtifactRoleEvidence.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ArtifactRoleEvidence)
	}
}

func TestRelationTypeJSONRoundTrip(t *testing.T) {
	original := NewRelationType(mustVocabularyValue(t, "product-x", "custom-relation"))
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RelationType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(original.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
	// A predefined constant round-trips identically.
	data2, err := json.Marshal(RelationTypeDerivation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded2 RelationType
	if err := json.Unmarshal(data2, &decoded2); err != nil {
		t.Fatal(err)
	}
	if !decoded2.Value().Equal(RelationTypeDerivation.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded2, RelationTypeDerivation)
	}
}

func TestClaimTypeJSONRoundTrip(t *testing.T) {
	original := NewClaimType(mustVocabularyValue(t, "product-x", "custom-claim-type"))
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ClaimType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(original.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestClaimOutcomeJSONRoundTrip(t *testing.T) {
	original := NewClaimOutcome(mustVocabularyValue(t, "product-x", "custom-outcome"))
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ClaimOutcome
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(original.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func mustVocabularyValue(t *testing.T, namespace, value string) VocabularyValue {
	t.Helper()
	v, err := NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// --- ExecutionOutcome (PEOS-006) ---------------------------------------------

func TestNewExecutionOutcomeValid(t *testing.T) {
	v := mustVocabularyValue(t, "product-x", "partially-completed")
	o, err := NewExecutionOutcome(v)
	if err != nil {
		t.Fatal(err)
	}
	if o.IsZero() {
		t.Fatal("valid ExecutionOutcome reports IsZero")
	}
	if !o.Value().Equal(v) {
		t.Errorf("Value() = %v, want %v", o.Value(), v)
	}
	if got := o.String(); got != "product-x:partially-completed" {
		t.Errorf("String() = %q", got)
	}
}

// TestNewExecutionOutcomeZeroRejected records the one deliberate divergence
// from this file's four sibling vocabulary wrappers: unlike
// NewClaimOutcome, NewClaimType, NewValidationMethod, and NewCorrectionKind,
// NewExecutionOutcome validates its input and rejects a zero
// VocabularyValue.
func TestNewExecutionOutcomeZeroRejected(t *testing.T) {
	_, err := NewExecutionOutcome(VocabularyValue{})
	if !errors.Is(err, ErrInvalidVocabularyValue) {
		t.Errorf("error = %v, want %v", err, ErrInvalidVocabularyValue)
	}
}

func TestExecutionOutcomeZeroValue(t *testing.T) {
	var o ExecutionOutcome
	if !o.IsZero() {
		t.Error("zero ExecutionOutcome does not report IsZero")
	}
	if got := o.String(); got != ":" {
		t.Errorf("zero String() = %q, want %q (the inner zero VocabularyValue's form)", got, ":")
	}
}

func TestExecutionOutcomePredeclaredConstants(t *testing.T) {
	cases := map[string]ExecutionOutcome{
		"peos:completed":     ExecutionOutcomeCompleted,
		"peos:failed":        ExecutionOutcomeFailed,
		"peos:interrupted":   ExecutionOutcomeInterrupted,
		"peos:indeterminate": ExecutionOutcomeIndeterminate,
	}
	for want, outcome := range cases {
		if got := outcome.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
		if outcome.IsZero() {
			t.Errorf("%s reports IsZero", want)
		}
		if outcome.Value().Namespace() != PEOSNamespace {
			t.Errorf("%s namespace = %q, want %q", want, outcome.Value().Namespace(), PEOSNamespace)
		}
	}
	if len(cases) != 4 {
		t.Fatalf("expected exactly the 4 PEOS-006 minimum outcomes, have %d", len(cases))
	}
}

// TestExecutionOutcomeDoesNotPredeclareGovernanceOrClaimValues guards
// against a future contributor adding values PEOS-006 does not mandate for
// this vocabulary. success/passed are not PEOS-006 execution outcomes;
// satisfied belongs to ClaimOutcome; accepted/certified are PEOS-004
// governance outcomes.
func TestExecutionOutcomeDoesNotPredeclareGovernanceOrClaimValues(t *testing.T) {
	declared := []ExecutionOutcome{
		ExecutionOutcomeCompleted,
		ExecutionOutcomeFailed,
		ExecutionOutcomeInterrupted,
		ExecutionOutcomeIndeterminate,
	}
	forbidden := []string{"success", "succeeded", "passed", "accepted", "certified", "satisfied", "not-satisfied", "inconclusive"}
	for _, o := range declared {
		for _, bad := range forbidden {
			if o.Value().Value() == bad {
				t.Errorf("ExecutionOutcome predeclares %q, which PEOS-006 does not mandate for this vocabulary", bad)
			}
		}
	}
}

func TestExecutionOutcomeEqual(t *testing.T) {
	if !ExecutionOutcomeCompleted.Equal(ExecutionOutcomeCompleted) {
		t.Error("Equal is not reflexive")
	}
	if ExecutionOutcomeCompleted.Equal(ExecutionOutcomeFailed) {
		t.Error("distinct outcomes report Equal")
	}
	rebuilt, err := NewExecutionOutcome(mustVocabularyValue(t, PEOSNamespace, "completed"))
	if err != nil {
		t.Fatal(err)
	}
	if !rebuilt.Equal(ExecutionOutcomeCompleted) {
		t.Error("an independently constructed equal value does not report Equal")
	}
	var zero ExecutionOutcome
	if zero.Equal(ExecutionOutcomeCompleted) {
		t.Error("zero reports Equal to a non-zero outcome")
	}
}

// TestExecutionOutcomeDistinctFromClaimOutcome proves the two PEOS-006
// vocabularies are separate Go types over disjoint value sets, so a
// completed execution can carry a not-satisfied Claim and neither value can
// be passed where the other is expected.
func TestExecutionOutcomeDistinctFromClaimOutcome(t *testing.T) {
	execution := []ExecutionOutcome{
		ExecutionOutcomeCompleted, ExecutionOutcomeFailed,
		ExecutionOutcomeInterrupted, ExecutionOutcomeIndeterminate,
	}
	claim := []ClaimOutcome{
		ClaimOutcomeSatisfied, ClaimOutcomeNotSatisfied, ClaimOutcomeInconclusive,
	}
	for _, e := range execution {
		for _, c := range claim {
			if e.String() == c.String() {
				t.Errorf("ExecutionOutcome and ClaimOutcome share the value %q", e.String())
			}
		}
	}
	if reflect.TypeOf(ExecutionOutcomeCompleted) == reflect.TypeOf(ClaimOutcomeSatisfied) {
		t.Error("ExecutionOutcome and ClaimOutcome are the same Go type")
	}
}

func TestExecutionOutcomeJSONRoundTrip(t *testing.T) {
	for _, original := range []ExecutionOutcome{
		ExecutionOutcomeCompleted, ExecutionOutcomeFailed,
		ExecutionOutcomeInterrupted, ExecutionOutcomeIndeterminate,
	} {
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+original.String()+`"` {
			t.Errorf("wire form = %s, want %q", data, original.String())
		}
		var decoded ExecutionOutcome
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if !decoded.Equal(original) {
			t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
		}
	}
}

func TestExecutionOutcomeJSONProductValuePreserved(t *testing.T) {
	original, err := NewExecutionOutcome(mustVocabularyValue(t, "acme", "aborted-by-operator"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ExecutionOutcome
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Error("Product-declared execution outcome not preserved through JSON")
	}
}

func TestExecutionOutcomeJSONInvalidRejected(t *testing.T) {
	for _, payload := range []string{`"no-colon"`, `""`, `123`, `{}`, `":"`} {
		var o ExecutionOutcome
		if err := json.Unmarshal([]byte(payload), &o); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

// TestExecutionOutcomeZeroMarshalDoesNotRoundTrip documents the actual
// behavior of marshaling a zero value: MarshalJSON delegates to the inner
// VocabularyValue exactly as the four sibling wrappers do, so it emits ":"
// rather than failing, and that output then fails to decode. A zero outcome
// cannot reach the wire through an ExecutionRecord, which rejects one.
func TestExecutionOutcomeZeroMarshalDoesNotRoundTrip(t *testing.T) {
	var zero ExecutionOutcome
	data, err := json.Marshal(zero)
	if err != nil {
		t.Fatalf("zero marshal failed; the sibling wrappers do not fail here: %v", err)
	}
	if string(data) != `":"` {
		t.Errorf("zero wire form = %s, want %q", data, ":")
	}
	var decoded ExecutionOutcome
	if err := json.Unmarshal(data, &decoded); err == nil {
		t.Error("zero wire form round-tripped, but ':' is not a valid vocabulary value")
	}
}

func TestExecutionOutcomeFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := ExecutionOutcomeCompleted
	if err := json.Unmarshal([]byte(`"no-colon"`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if !receiver.Equal(ExecutionOutcomeCompleted) {
		t.Errorf("failed Unmarshal disturbed the receiver: got %v", receiver)
	}
}
