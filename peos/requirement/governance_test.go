package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustDecisionOutcomeRef(t *testing.T, decisionID string) core.DecisionOutcomeRef {
	t.Helper()
	id, err := core.NewDecisionID(decisionID)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewDecisionOutcomeRef(id)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustProductMechanism(t *testing.T, namespace, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustGovernanceActionFromDecisionOutcome(t *testing.T, decisionID string) GovernanceAction {
	t.Helper()
	g, err := NewGovernanceActionFromDecisionOutcome(mustDecisionOutcomeRef(t, decisionID))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func mustGovernanceActionFromProductMechanism(t *testing.T, namespace, value string) GovernanceAction {
	t.Helper()
	g, err := NewGovernanceActionFromProductMechanism(mustProductMechanism(t, namespace, value))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// --- NewGovernanceActionFromDecisionOutcome ---------------------------------

func TestNewGovernanceActionFromDecisionOutcomeValid(t *testing.T) {
	g := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	if g.IsZero() {
		t.Error("valid GovernanceAction IsZero() = true")
	}
	if !g.IsDecisionOutcome() {
		t.Error("IsDecisionOutcome() = false")
	}
	if g.IsProductMechanism() {
		t.Error("IsProductMechanism() = true")
	}
	if g.Kind() != "decision_outcome" {
		t.Errorf("Kind() = %q, want %q", g.Kind(), "decision_outcome")
	}
}

func TestNewGovernanceActionFromDecisionOutcomeZeroRejected(t *testing.T) {
	_, err := NewGovernanceActionFromDecisionOutcome(core.DecisionOutcomeRef{})
	if !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

// --- NewGovernanceActionFromProductMechanism --------------------------------

func TestNewGovernanceActionFromProductMechanismValid(t *testing.T) {
	g := mustGovernanceActionFromProductMechanism(t, "acme", "change-board-approval")
	if g.IsZero() {
		t.Error("valid GovernanceAction IsZero() = true")
	}
	if !g.IsProductMechanism() {
		t.Error("IsProductMechanism() = false")
	}
	if g.IsDecisionOutcome() {
		t.Error("IsDecisionOutcome() = true")
	}
	if g.Kind() != "product_mechanism" {
		t.Errorf("Kind() = %q, want %q", g.Kind(), "product_mechanism")
	}
}

func TestNewGovernanceActionFromProductMechanismZeroRejected(t *testing.T) {
	_, err := NewGovernanceActionFromProductMechanism(core.VocabularyValue{})
	if !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

// --- Accessors ---------------------------------------------------------

func TestGovernanceActionAsDecisionOutcomeWrongArm(t *testing.T) {
	g := mustGovernanceActionFromProductMechanism(t, "acme", "change-board-approval")
	if _, ok := g.AsDecisionOutcome(); ok {
		t.Error("AsDecisionOutcome() ok = true on a product-mechanism arm")
	}
}

func TestGovernanceActionAsProductMechanismWrongArm(t *testing.T) {
	g := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	if _, ok := g.AsProductMechanism(); ok {
		t.Error("AsProductMechanism() ok = true on a decision-outcome arm")
	}
}

func TestGovernanceActionAsDecisionOutcomeCorrectArm(t *testing.T) {
	ref := mustDecisionOutcomeRef(t, "DEC-1")
	g, err := NewGovernanceActionFromDecisionOutcome(ref)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := g.AsDecisionOutcome()
	if !ok || got != ref {
		t.Errorf("AsDecisionOutcome() = (%v,%v), want (%v,true)", got, ok, ref)
	}
}

func TestGovernanceActionAsProductMechanismCorrectArm(t *testing.T) {
	value := mustProductMechanism(t, "acme", "change-board-approval")
	g, err := NewGovernanceActionFromProductMechanism(value)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := g.AsProductMechanism()
	if !ok || got != value {
		t.Errorf("AsProductMechanism() = (%v,%v), want (%v,true)", got, ok, value)
	}
}

func TestGovernanceActionIsZero(t *testing.T) {
	var g GovernanceAction
	if !g.IsZero() {
		t.Error("zero GovernanceAction IsZero() = false")
	}
	if mustGovernanceActionFromDecisionOutcome(t, "DEC-1").IsZero() {
		t.Error("valid GovernanceAction IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestGovernanceActionJSONDecisionOutcomeRoundTrip(t *testing.T) {
	g := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var kind string
	if err := json.Unmarshal(raw["kind"], &kind); err != nil {
		t.Fatal(err)
	}
	if kind != "decision_outcome" {
		t.Errorf("kind = %q, want %q", kind, "decision_outcome")
	}
	var decoded GovernanceAction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, _ := decoded.AsDecisionOutcome()
	want, _ := g.AsDecisionOutcome()
	if got != want {
		t.Errorf("round trip mismatch: got %v, want %v", got, want)
	}
}

func TestGovernanceActionJSONProductMechanismRoundTrip(t *testing.T) {
	g := mustGovernanceActionFromProductMechanism(t, "acme", "change-board-approval")
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var refStr string
	if err := json.Unmarshal(raw["ref"], &refStr); err != nil {
		t.Fatal(err)
	}
	if refStr != "acme:change-board-approval" {
		t.Errorf("ref = %q, want %q", refStr, "acme:change-board-approval")
	}
	var decoded GovernanceAction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	gotValue, _ := decoded.AsProductMechanism()
	wantValue, _ := g.AsProductMechanism()
	if gotValue != wantValue {
		t.Errorf("round trip mismatch: got %v, want %v", gotValue, wantValue)
	}
}

func TestGovernanceActionJSONUnrecognizedKindRejected(t *testing.T) {
	var g GovernanceAction
	payload := `{"kind":"authority","ref":{"decision_id":"DEC-1"}}`
	if err := json.Unmarshal([]byte(payload), &g); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestGovernanceActionJSONMissingRefRejected(t *testing.T) {
	var g GovernanceAction
	payload := `{"kind":"decision_outcome"}`
	if err := json.Unmarshal([]byte(payload), &g); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestGovernanceActionJSONNullRefRejected(t *testing.T) {
	var g GovernanceAction
	payload := `{"kind":"decision_outcome","ref":null}`
	if err := json.Unmarshal([]byte(payload), &g); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestGovernanceActionZeroMarshalRejected(t *testing.T) {
	var g GovernanceAction
	if _, err := json.Marshal(g); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestGovernanceActionUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver != original {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- Nested sentinel preservation --------------------------------------

// TestGovernanceActionNestedSentinelPreserved proves a malformed nested
// core.DecisionOutcomeRef (empty decision_id) preserves both
// ErrInvalidGovernanceAction and the underlying core.ErrEmptyIdentity
// through errors.Is.
func TestGovernanceActionNestedSentinelPreserved(t *testing.T) {
	var g GovernanceAction
	payload := `{"kind":"decision_outcome","ref":{"decision_id":""}}`
	err := json.Unmarshal([]byte(payload), &g)
	if err == nil {
		t.Fatal("malformed nested ref accepted, want error")
	}
	if !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidGovernanceAction)
	}
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("error = %v, want also wrapping %v", err, core.ErrEmptyIdentity)
	}
}

// --- Absence audit -------------------------------------------------------

// TestGovernanceActionHasNoIdentity is a structural absence audit proving
// GovernanceAction carries no identity (PEOS-005 §23: "Governance action
// is a semantic role and SHALL NOT be interpreted as introducing a
// separate PEOS entity").
func TestGovernanceActionHasNoIdentity(t *testing.T) {
	g := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["id"]; present {
		t.Error(`unexpected "id" key present in GovernanceAction wire form`)
	}
	if len(raw) != 2 {
		t.Errorf("GovernanceAction wire form has %d top-level keys, want exactly 2 (kind, ref): %v", len(raw), raw)
	}
}
