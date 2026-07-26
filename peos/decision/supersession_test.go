package decision

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustSupersessionID(t *testing.T, value string) SupersessionID {
	t.Helper()
	id, err := NewSupersessionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustSupersessionDecisionRef(t *testing.T, value string) core.DecisionRef {
	t.Helper()
	ref, err := core.NewDecisionRef(mustDecisionID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustTimestamp(t *testing.T, value string) core.Timestamp {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	ts, err := core.NewTimestamp(parsed)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

// --- SupersessionExtent ------------------------------------------------

func TestNewCompleteSupersessionExtent(t *testing.T) {
	e := NewCompleteSupersessionExtent()
	if !e.IsComplete() || e.IsPartial() {
		t.Errorf("IsComplete/IsPartial = %v/%v, want true/false", e.IsComplete(), e.IsPartial())
	}
	if _, ok := e.RemainingScope(); ok {
		t.Error("RemainingScope() ok = true for complete extent")
	}
}

func TestNewPartialSupersessionExtentValid(t *testing.T) {
	scope := mustScope(t, "/services/*")
	e, err := NewPartialSupersessionExtent(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !e.IsPartial() || e.IsComplete() {
		t.Errorf("IsPartial/IsComplete = %v/%v, want true/false", e.IsPartial(), e.IsComplete())
	}
	got, ok := e.RemainingScope()
	if !ok || !got.Equal(scope) {
		t.Errorf("RemainingScope() = (%v,%v), want (%v,true)", got, ok, scope)
	}
}

func TestNewPartialSupersessionExtentZeroScopeRejected(t *testing.T) {
	if _, err := NewPartialSupersessionExtent(core.Scope{}); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentEqual(t *testing.T) {
	scope := mustScope(t, "/services/*")
	a, err := NewPartialSupersessionExtent(scope)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewPartialSupersessionExtent(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !a.Equal(b) {
		t.Error("Equal(same partial) = false")
	}
	if NewCompleteSupersessionExtent().Equal(a) {
		t.Error("Equal(complete, partial) = true")
	}
	if !NewCompleteSupersessionExtent().Equal(NewCompleteSupersessionExtent()) {
		t.Error("Equal(complete, complete) = false")
	}
}

func TestSupersessionExtentString(t *testing.T) {
	if NewCompleteSupersessionExtent().String() != "complete" {
		t.Errorf("String() = %q, want %q", NewCompleteSupersessionExtent().String(), "complete")
	}
	scope := mustScope(t, "/services/*")
	e, err := NewPartialSupersessionExtent(scope)
	if err != nil {
		t.Fatal(err)
	}
	if e.String() != "partial" {
		t.Errorf("String() = %q, want %q", e.String(), "partial")
	}
}

func TestSupersessionExtentZeroBehavior(t *testing.T) {
	var e SupersessionExtent
	if !e.IsZero() {
		t.Error("zero SupersessionExtent IsZero() = false")
	}
	if e.IsComplete() || e.IsPartial() {
		t.Error("zero SupersessionExtent reports a variant")
	}
}

func TestSupersessionExtentJSONCompleteShape(t *testing.T) {
	e := NewCompleteSupersessionExtent()
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["kind"]) != `"complete"` {
		t.Errorf(`kind = %s, want "complete"`, raw["kind"])
	}
	if _, present := raw["remaining_scope"]; present {
		t.Error("remaining_scope present for complete extent")
	}
	var decoded SupersessionExtent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(e) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, e)
	}
}

func TestSupersessionExtentJSONPartialShape(t *testing.T) {
	scope := mustScope(t, "/services/*")
	e, err := NewPartialSupersessionExtent(scope)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["kind"]) != `"partial"` {
		t.Errorf(`kind = %s, want "partial"`, raw["kind"])
	}
	if _, present := raw["remaining_scope"]; !present {
		t.Error("remaining_scope absent for partial extent")
	}
	var decoded SupersessionExtent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(e) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, e)
	}
}

func TestSupersessionExtentUnknownKindRejected(t *testing.T) {
	var e SupersessionExtent
	if err := json.Unmarshal([]byte(`{"kind":"mostly"}`), &e); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentMissingKindRejected(t *testing.T) {
	var e SupersessionExtent
	if err := json.Unmarshal([]byte(`{}`), &e); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentCompleteWithRemainingScopeRejected(t *testing.T) {
	var e SupersessionExtent
	payload := `{"kind":"complete","remaining_scope":{"kind":"product-x:path","expression":"/x"}}`
	if err := json.Unmarshal([]byte(payload), &e); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentPartialWithoutRemainingScopeRejected(t *testing.T) {
	var e SupersessionExtent
	if err := json.Unmarshal([]byte(`{"kind":"partial"}`), &e); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentPartialWithNullRemainingScopeRejected(t *testing.T) {
	var e SupersessionExtent
	if err := json.Unmarshal([]byte(`{"kind":"partial","remaining_scope":null}`), &e); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentZeroMarshalRejected(t *testing.T) {
	var e SupersessionExtent
	if _, err := json.Marshal(e); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionExtentUnmarshalFailurePreservesReceiver(t *testing.T) {
	scope := mustScope(t, "/services/*")
	original, err := NewPartialSupersessionExtent(scope)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"kind":"partial"}`), &receiver); err == nil {
		t.Fatal("partial without remaining_scope accepted, want error")
	}
	if !receiver.Equal(original) {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestSupersessionExtentUnknownFieldIgnored(t *testing.T) {
	var e SupersessionExtent
	if err := json.Unmarshal([]byte(`{"kind":"complete","unknown_field":123}`), &e); err != nil {
		t.Fatal(err)
	}
}

// --- DecisionSupersession construction -----------------------------------

func baseSupersessionArgs(t *testing.T) (SupersessionID, core.DecisionRef, core.DecisionRef, core.Scope, SupersessionExtent) {
	t.Helper()
	return mustSupersessionID(t, "sup-1"),
		mustSupersessionDecisionRef(t, "dec-2"),
		mustSupersessionDecisionRef(t, "dec-1"),
		mustScope(t, "/services/billing"),
		NewCompleteSupersessionExtent()
}

func TestNewDecisionSupersessionValidWithTimeOnly(t *testing.T) {
	id, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	s, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.EffectiveAt(); !ok {
		t.Error("EffectiveAt() ok = false")
	}
	if _, ok := s.EffectiveCondition(); ok {
		t.Error("EffectiveCondition() ok = true")
	}
}

func TestNewDecisionSupersessionValidWithConditionOnly(t *testing.T) {
	id, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	s, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, core.Timestamp{}, "when the migration completes")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.EffectiveAt(); ok {
		t.Error("EffectiveAt() ok = true")
	}
	cond, ok := s.EffectiveCondition()
	if !ok || cond != "when the migration completes" {
		t.Errorf("EffectiveCondition() = (%q,%v)", cond, ok)
	}
}

func TestNewDecisionSupersessionValidWithBoth(t *testing.T) {
	id, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	s, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), "when the migration completes")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.EffectiveAt(); !ok {
		t.Error("EffectiveAt() ok = false")
	}
	if _, ok := s.EffectiveCondition(); !ok {
		t.Error("EffectiveCondition() ok = false")
	}
}

func TestNewDecisionSupersessionNeitherEffectiveRejected(t *testing.T) {
	id, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, core.Timestamp{}, ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionZeroIDRejected(t *testing.T) {
	_, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(SupersessionID{}, superseding, superseded, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionZeroSupersedingRejected(t *testing.T) {
	id, _, superseded, scope, extent := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(id, core.DecisionRef{}, superseded, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionZeroSupersededRejected(t *testing.T) {
	id, superseding, _, scope, extent := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(id, superseding, core.DecisionRef{}, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionSameDecisionRejected(t *testing.T) {
	id, superseding, _, scope, extent := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(id, superseding, superseding, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionZeroAffectedScopeRejected(t *testing.T) {
	id, superseding, superseded, _, extent := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(id, superseding, superseded, core.Scope{}, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionZeroExtentRejected(t *testing.T) {
	id, superseding, superseded, scope, _ := baseSupersessionArgs(t)
	if _, err := NewDecisionSupersession(id, superseding, superseded, scope, SupersessionExtent{}, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewDecisionSupersessionPartialExtentPreservesRemainingScope(t *testing.T) {
	id, superseding, superseded, scope, _ := baseSupersessionArgs(t)
	remaining := mustScope(t, "/services/*")
	extent, err := NewPartialSupersessionExtent(remaining)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), "")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := s.Extent().RemainingScope()
	if !ok || !got.Equal(remaining) {
		t.Errorf("RemainingScope() = (%v,%v), want (%v,true)", got, ok, remaining)
	}
}

// --- With*/Without* and accessors ---------------------------------------

func baseSupersession(t *testing.T) DecisionSupersession {
	t.Helper()
	id, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	s, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, mustTimestamp(t, "2026-07-26T10:00:00Z"), "")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestDecisionSupersessionReasonAbsentPresent(t *testing.T) {
	s := baseSupersession(t)
	if _, ok := s.Reason(); ok {
		t.Error("Reason() ok = true before WithReason")
	}
	withReason, err := s.WithReason("scheduled migration")
	if err != nil {
		t.Fatal(err)
	}
	reason, ok := withReason.Reason()
	if !ok || reason != "scheduled migration" {
		t.Errorf("Reason() = (%q,%v)", reason, ok)
	}
	if _, ok := s.Reason(); ok {
		t.Error("WithReason mutated the original receiver")
	}
}

func TestDecisionSupersessionEmptyReasonRejected(t *testing.T) {
	s := baseSupersession(t)
	if _, err := s.WithReason(""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestDecisionSupersessionWithoutReason(t *testing.T) {
	s := baseSupersession(t)
	s, err := s.WithReason("scheduled migration")
	if err != nil {
		t.Fatal(err)
	}
	cleared := s.WithoutReason()
	if _, ok := cleared.Reason(); ok {
		t.Error("Reason() ok = true after WithoutReason")
	}
	if _, ok := s.Reason(); !ok {
		t.Error("WithoutReason mutated the original receiver")
	}
}

func TestDecisionSupersessionProvenanceAbsentPresent(t *testing.T) {
	s := baseSupersession(t)
	if _, ok := s.Provenance(); ok {
		t.Error("Provenance() ok = true before WithProvenance")
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	withProv, err := s.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withProv.Provenance(); !ok {
		t.Error("Provenance() ok = false after WithProvenance")
	}
	cleared := withProv.WithoutProvenance()
	if _, ok := cleared.Provenance(); ok {
		t.Error("Provenance() ok = true after WithoutProvenance")
	}
}

func TestDecisionSupersessionZeroProvenanceRejected(t *testing.T) {
	s := baseSupersession(t)
	if _, err := s.WithProvenance(core.Provenance{}); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestDecisionSupersessionExtension(t *testing.T) {
	s := baseSupersession(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := s.WithExtension(ext)
	if !s.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestDecisionSupersessionWithMethodsPreserveUnrelatedFields(t *testing.T) {
	s := baseSupersession(t)
	withReason, err := s.WithReason("scheduled migration")
	if err != nil {
		t.Fatal(err)
	}
	if withReason.ID() != s.ID() || withReason.SupersedingDecision() != s.SupersedingDecision() ||
		withReason.SupersededDecision() != s.SupersededDecision() || !withReason.AffectedScope().Equal(s.AffectedScope()) ||
		!withReason.Extent().Equal(s.Extent()) {
		t.Error("WithReason altered unrelated fields")
	}
}

// --- JSON ----------------------------------------------------------------

func fullSupersession(t *testing.T) DecisionSupersession {
	t.Helper()
	s := baseSupersession(t)
	s, err := s.WithReason("scheduled migration")
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	s, err = s.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return s.WithExtension(ext)
}

func TestDecisionSupersessionJSONMinimumTimeOnlyRoundTrip(t *testing.T) {
	s := baseSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DecisionSupersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.ID().Equal(s.ID()) {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID(), s.ID())
	}
}

func TestDecisionSupersessionJSONMinimumConditionOnlyRoundTrip(t *testing.T) {
	id, superseding, superseded, scope, extent := baseSupersessionArgs(t)
	s, err := NewDecisionSupersession(id, superseding, superseded, scope, extent, core.Timestamp{}, "when migration completes")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DecisionSupersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	cond, ok := decoded.EffectiveCondition()
	if !ok || cond != "when migration completes" {
		t.Errorf("EffectiveCondition() = (%q,%v)", cond, ok)
	}
}

func TestDecisionSupersessionJSONFullRoundTrip(t *testing.T) {
	s := fullSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DecisionSupersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.ID().Equal(s.ID()) {
		t.Errorf("ID mismatch")
	}
	if _, ok := decoded.Reason(); !ok {
		t.Error("Reason absent after round trip")
	}
	if _, ok := decoded.Provenance(); !ok {
		t.Error("Provenance absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDecisionSupersessionJSONRequiredWireKeys(t *testing.T) {
	s := baseSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "superseding_decision", "superseded_decision", "affected_scope", "extent", "effective_at"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
}

func TestDecisionSupersessionJSONOptionalFieldsOmitted(t *testing.T) {
	s := baseSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"effective_condition", "reason", "provenance", "extension"} {
		if _, present := raw[key]; present {
			t.Errorf("optional field %q present despite not being set", key)
		}
	}
}

func supersessionBaseJSON() string {
	return `{"id":"sup-1","superseding_decision":{"decision_id":"dec-2"},"superseded_decision":{"decision_id":"dec-1"},"affected_scope":{"kind":"product-x:path","expression":"/x"},"extent":{"kind":"complete"},"effective_at":"2026-07-26T10:00:00Z"`
}

func TestDecisionSupersessionExplicitNullRejected(t *testing.T) {
	fields := []string{"effective_at", "effective_condition", "reason", "provenance", "extension"}
	for _, field := range fields {
		payload := supersessionBaseJSON() + `,"` + field + `":null}`
		var s DecisionSupersession
		if err := json.Unmarshal([]byte(payload), &s); err == nil {
			t.Errorf("field %q: explicit null accepted, want error", field)
		}
	}
}

func TestDecisionSupersessionEmptyStringsRejectedWhenPresent(t *testing.T) {
	var s DecisionSupersession
	if err := json.Unmarshal([]byte(supersessionBaseJSON()+`,"effective_condition":""}`), &s); err == nil {
		t.Error("empty effective_condition accepted, want error")
	}
	if err := json.Unmarshal([]byte(supersessionBaseJSON()+`,"reason":""}`), &s); err == nil {
		t.Error("empty reason accepted, want error")
	}
}

func TestDecisionSupersessionUnknownFieldIgnored(t *testing.T) {
	var s DecisionSupersession
	payload := supersessionBaseJSON() + `,"unknown_field":123}`
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionSupersessionZeroMarshalRejected(t *testing.T) {
	var s DecisionSupersession
	if _, err := json.Marshal(s); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestDecisionSupersessionUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullSupersession(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"id":"sup-1"}`), &receiver); err == nil {
		t.Fatal("missing required fields accepted, want error")
	}
	if !receiver.ID().Equal(original.ID()) {
		t.Error("failed Unmarshal changed receiver ID")
	}
	if _, ok := receiver.Reason(); !ok {
		t.Error("failed Unmarshal changed receiver's reason presence")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

func TestDecisionSupersessionNestedDecodeErrorsWrapSentinel(t *testing.T) {
	var s DecisionSupersession
	// Malformed nested affected_scope (missing required expression).
	payload := `{"id":"sup-1","superseding_decision":{"decision_id":"dec-2"},"superseded_decision":{"decision_id":"dec-1"},"affected_scope":{"kind":"product-x:path","expression":""},"extent":{"kind":"complete"},"effective_at":"2026-07-26T10:00:00Z"}`
	err := json.Unmarshal([]byte(payload), &s)
	if err == nil {
		t.Fatal("malformed affected_scope accepted, want error")
	}
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want wrapping %v", err, core.ErrInvalidScope)
	}
	if !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want also wrapping %v", err, ErrInvalidDecisionSupersession)
	}
}
