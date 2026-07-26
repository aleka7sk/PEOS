package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustInvalidationID(t *testing.T, value string) InvalidationID {
	t.Helper()
	id, err := NewInvalidationID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// --- InvalidationSource ---------------------------------------------------

func TestNewInvalidationSourceFromDecisionValid(t *testing.T) {
	ref := mustSupersessionDecisionRef(t, "dec-9")
	s, err := NewInvalidationSourceFromDecision(ref)
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsDecision() || s.IsAuthority() {
		t.Errorf("IsDecision/IsAuthority = %v/%v, want true/false", s.IsDecision(), s.IsAuthority())
	}
	got, ok := s.AsDecision()
	if !ok || got != ref {
		t.Errorf("AsDecision() = (%v,%v), want (%v,true)", got, ok, ref)
	}
}

func TestNewInvalidationSourceFromAuthorityValid(t *testing.T) {
	ref := mustAuthorityRef(t, "role", "cto")
	s, err := NewInvalidationSourceFromAuthority(ref)
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsAuthority() || s.IsDecision() {
		t.Errorf("IsAuthority/IsDecision = %v/%v, want true/false", s.IsAuthority(), s.IsDecision())
	}
	got, ok := s.AsAuthority()
	if !ok || got != ref {
		t.Errorf("AsAuthority() = (%v,%v), want (%v,true)", got, ok, ref)
	}
}

func TestNewInvalidationSourceZeroDecisionRefRejected(t *testing.T) {
	if _, err := NewInvalidationSourceFromDecision(core.DecisionRef{}); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewInvalidationSourceZeroAuthorityRefRejected(t *testing.T) {
	if _, err := NewInvalidationSourceFromAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceKind(t *testing.T) {
	decSource, err := NewInvalidationSourceFromDecision(mustSupersessionDecisionRef(t, "dec-9"))
	if err != nil {
		t.Fatal(err)
	}
	if decSource.Kind() != "decision" {
		t.Errorf("Kind() = %q, want %q", decSource.Kind(), "decision")
	}
	authSource, err := NewInvalidationSourceFromAuthority(mustAuthorityRef(t, "role", "cto"))
	if err != nil {
		t.Fatal(err)
	}
	if authSource.Kind() != "authority" {
		t.Errorf("Kind() = %q, want %q", authSource.Kind(), "authority")
	}
}

func TestInvalidationSourceAsWrongArm(t *testing.T) {
	decSource, err := NewInvalidationSourceFromDecision(mustSupersessionDecisionRef(t, "dec-9"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := decSource.AsAuthority(); ok {
		t.Error("AsAuthority() ok = true on decision arm")
	}
	authSource, err := NewInvalidationSourceFromAuthority(mustAuthorityRef(t, "role", "cto"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := authSource.AsDecision(); ok {
		t.Error("AsDecision() ok = true on authority arm")
	}
}

func TestInvalidationSourceEqual(t *testing.T) {
	ref := mustSupersessionDecisionRef(t, "dec-9")
	a, err := NewInvalidationSourceFromDecision(ref)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewInvalidationSourceFromDecision(ref)
	if err != nil {
		t.Fatal(err)
	}
	authSource, err := NewInvalidationSourceFromAuthority(mustAuthorityRef(t, "role", "cto"))
	if err != nil {
		t.Fatal(err)
	}
	if !a.Equal(b) {
		t.Error("Equal(same decision arm) = false")
	}
	if a.Equal(authSource) {
		t.Error("Equal(decision arm, authority arm) = true")
	}
}

func TestInvalidationSourceJSONDecisionArmShape(t *testing.T) {
	ref := mustSupersessionDecisionRef(t, "dec-9")
	s, err := NewInvalidationSourceFromDecision(ref)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["kind"]) != `"decision"` {
		t.Errorf(`kind = %s, want "decision"`, raw["kind"])
	}
	if _, present := raw["ref"]; !present {
		t.Error("ref key missing")
	}
	var decoded InvalidationSource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(s) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, s)
	}
}

func TestInvalidationSourceJSONAuthorityArmShape(t *testing.T) {
	ref := mustAuthorityRef(t, "role", "cto")
	s, err := NewInvalidationSourceFromAuthority(ref)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["kind"]) != `"authority"` {
		t.Errorf(`kind = %s, want "authority"`, raw["kind"])
	}
	var decoded InvalidationSource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(s) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, s)
	}
}

func TestInvalidationSourceUnknownKindRejected(t *testing.T) {
	var s InvalidationSource
	payload := `{"kind":"decree","ref":{"decision_id":"dec-9"}}`
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceMissingKindRejected(t *testing.T) {
	var s InvalidationSource
	if err := json.Unmarshal([]byte(`{"ref":{"decision_id":"dec-9"}}`), &s); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceMissingRefRejected(t *testing.T) {
	var s InvalidationSource
	if err := json.Unmarshal([]byte(`{"kind":"decision"}`), &s); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceNullRefRejected(t *testing.T) {
	var s InvalidationSource
	if err := json.Unmarshal([]byte(`{"kind":"decision","ref":null}`), &s); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceMalformedNestedRefWrapsSentinel(t *testing.T) {
	var s InvalidationSource
	payload := `{"kind":"authority","ref":{"namespace":"","identifier":""}}`
	err := json.Unmarshal([]byte(payload), &s)
	if err == nil {
		t.Fatal("malformed nested ref accepted, want error")
	}
	if !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceZeroMarshalRejected(t *testing.T) {
	var s InvalidationSource
	if _, err := json.Marshal(s); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationSourceUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewInvalidationSourceFromDecision(mustSupersessionDecisionRef(t, "dec-9"))
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"kind":"decision"}`), &receiver); err == nil {
		t.Fatal("missing ref accepted, want error")
	}
	if !receiver.Equal(original) {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestInvalidationSourceUnknownFieldIgnored(t *testing.T) {
	var s InvalidationSource
	payload := `{"kind":"decision","ref":{"decision_id":"dec-9"},"unknown_field":123}`
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Fatal(err)
	}
}

// --- DecisionInvalidation construction ------------------------------------

func baseInvalidationArgs(t *testing.T) (InvalidationID, core.DecisionRef, InvalidationSource, string) {
	t.Helper()
	source, err := NewInvalidationSourceFromDecision(mustSupersessionDecisionRef(t, "dec-9"))
	if err != nil {
		t.Fatal(err)
	}
	return mustInvalidationID(t, "inv-1"), mustSupersessionDecisionRef(t, "dec-1"), source, "authority basis was withdrawn"
}

func TestNewDecisionInvalidationValidDecisionSourceWithTime(t *testing.T) {
	id, invalidated, source, reason := baseInvalidationArgs(t)
	i, err := NewDecisionInvalidation(id, invalidated, source, reason, mustTimestamp(t, "2026-07-26T10:00:00Z"), "")
	if err != nil {
		t.Fatal(err)
	}
	if i.Reason() != reason {
		t.Errorf("Reason() = %q, want %q", i.Reason(), reason)
	}
}

func TestNewDecisionInvalidationValidAuthoritySourceWithCondition(t *testing.T) {
	id, invalidated, _, reason := baseInvalidationArgs(t)
	authSource, err := NewInvalidationSourceFromAuthority(mustAuthorityRef(t, "role", "cto"))
	if err != nil {
		t.Fatal(err)
	}
	i, err := NewDecisionInvalidation(id, invalidated, authSource, reason, core.Timestamp{}, "when revocation becomes effective")
	if err != nil {
		t.Fatal(err)
	}
	cond, ok := i.EffectiveCondition()
	if !ok || cond != "when revocation becomes effective" {
		t.Errorf("EffectiveCondition() = (%q,%v)", cond, ok)
	}
}

func TestNewDecisionInvalidationValidBothEffectiveForms(t *testing.T) {
	id, invalidated, source, reason := baseInvalidationArgs(t)
	i, err := NewDecisionInvalidation(id, invalidated, source, reason, mustTimestamp(t, "2026-07-26T10:00:00Z"), "when revocation becomes effective")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := i.EffectiveAt(); !ok {
		t.Error("EffectiveAt() ok = false")
	}
	if _, ok := i.EffectiveCondition(); !ok {
		t.Error("EffectiveCondition() ok = false")
	}
}

func TestNewDecisionInvalidationNeitherEffectiveRejected(t *testing.T) {
	id, invalidated, source, reason := baseInvalidationArgs(t)
	if _, err := NewDecisionInvalidation(id, invalidated, source, reason, core.Timestamp{}, ""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewDecisionInvalidationZeroIDRejected(t *testing.T) {
	_, invalidated, source, reason := baseInvalidationArgs(t)
	if _, err := NewDecisionInvalidation(InvalidationID{}, invalidated, source, reason, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewDecisionInvalidationZeroInvalidatedDecisionRejected(t *testing.T) {
	id, _, source, reason := baseInvalidationArgs(t)
	if _, err := NewDecisionInvalidation(id, core.DecisionRef{}, source, reason, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewDecisionInvalidationZeroSourceRejected(t *testing.T) {
	id, invalidated, _, reason := baseInvalidationArgs(t)
	if _, err := NewDecisionInvalidation(id, invalidated, InvalidationSource{}, reason, mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewDecisionInvalidationEmptyReasonRejected(t *testing.T) {
	id, invalidated, source, _ := baseInvalidationArgs(t)
	if _, err := NewDecisionInvalidation(id, invalidated, source, "", mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewDecisionInvalidationWhitespaceOnlyReasonRejected(t *testing.T) {
	id, invalidated, source, _ := baseInvalidationArgs(t)
	if _, err := NewDecisionInvalidation(id, invalidated, source, "   ", mustTimestamp(t, "2026-07-26T10:00:00Z"), ""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

// --- With*/Without* and accessors ---------------------------------------

func baseInvalidation(t *testing.T) DecisionInvalidation {
	t.Helper()
	id, invalidated, source, reason := baseInvalidationArgs(t)
	i, err := NewDecisionInvalidation(id, invalidated, source, reason, mustTimestamp(t, "2026-07-26T10:00:00Z"), "")
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func TestDecisionInvalidationProvenanceAbsentPresent(t *testing.T) {
	i := baseInvalidation(t)
	if _, ok := i.Provenance(); ok {
		t.Error("Provenance() ok = true before WithProvenance")
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	withProv, err := i.WithProvenance(prov)
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
	if _, ok := i.Provenance(); ok {
		t.Error("WithProvenance mutated the original receiver")
	}
}

func TestDecisionInvalidationZeroProvenanceRejected(t *testing.T) {
	i := baseInvalidation(t)
	if _, err := i.WithProvenance(core.Provenance{}); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestDecisionInvalidationExtension(t *testing.T) {
	i := baseInvalidation(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := i.WithExtension(ext)
	if !i.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestDecisionInvalidationRequiredFieldsNeverChanged(t *testing.T) {
	i := baseInvalidation(t)
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	withProv, err := i.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	if withProv.ID() != i.ID() || withProv.InvalidatedDecision() != i.InvalidatedDecision() ||
		!withProv.Source().Equal(i.Source()) || withProv.Reason() != i.Reason() {
		t.Error("WithProvenance altered a required field")
	}
}

// --- JSON ----------------------------------------------------------------

func fullInvalidation(t *testing.T) DecisionInvalidation {
	t.Helper()
	i := baseInvalidation(t)
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	i, err := i.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return i.WithExtension(ext)
}

func TestDecisionInvalidationJSONMinimumRoundTrip(t *testing.T) {
	i := baseInvalidation(t)
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DecisionInvalidation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.ID().Equal(i.ID()) {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID(), i.ID())
	}
}

func TestDecisionInvalidationJSONFullRoundTrip(t *testing.T) {
	i := fullInvalidation(t)
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DecisionInvalidation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Reason() != i.Reason() {
		t.Errorf("Reason mismatch: got %q, want %q", decoded.Reason(), i.Reason())
	}
	if _, ok := decoded.Provenance(); !ok {
		t.Error("Provenance absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDecisionInvalidationJSONRequiredWireKeys(t *testing.T) {
	i := baseInvalidation(t)
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "invalidated_decision", "source", "reason", "effective_at"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
}

func TestDecisionInvalidationJSONOptionalFieldsOmitted(t *testing.T) {
	i := baseInvalidation(t)
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"effective_condition", "provenance", "extension"} {
		if _, present := raw[key]; present {
			t.Errorf("optional field %q present despite not being set", key)
		}
	}
}

func invalidationBaseJSON() string {
	return `{"id":"inv-1","invalidated_decision":{"decision_id":"dec-1"},"source":{"kind":"decision","ref":{"decision_id":"dec-9"}},"reason":"authority basis was withdrawn","effective_at":"2026-07-26T10:00:00Z"`
}

func TestDecisionInvalidationExplicitNullRejected(t *testing.T) {
	fields := []string{"effective_at", "effective_condition", "provenance", "extension"}
	for _, field := range fields {
		payload := invalidationBaseJSON() + `,"` + field + `":null}`
		var i DecisionInvalidation
		if err := json.Unmarshal([]byte(payload), &i); err == nil {
			t.Errorf("field %q: explicit null accepted, want error", field)
		}
	}
}

func TestDecisionInvalidationEmptyEffectiveConditionRejected(t *testing.T) {
	var i DecisionInvalidation
	if err := json.Unmarshal([]byte(invalidationBaseJSON()+`,"effective_condition":""}`), &i); err == nil {
		t.Error("empty effective_condition accepted, want error")
	}
}

func TestDecisionInvalidationUnknownFieldIgnored(t *testing.T) {
	var i DecisionInvalidation
	payload := invalidationBaseJSON() + `,"unknown_field":123}`
	if err := json.Unmarshal([]byte(payload), &i); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionInvalidationZeroMarshalRejected(t *testing.T) {
	var i DecisionInvalidation
	if _, err := json.Marshal(i); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestDecisionInvalidationUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullInvalidation(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"id":"inv-1"}`), &receiver); err == nil {
		t.Fatal("missing required fields accepted, want error")
	}
	if !receiver.ID().Equal(original.ID()) {
		t.Error("failed Unmarshal changed receiver ID")
	}
	if receiver.Reason() != original.Reason() {
		t.Error("failed Unmarshal changed receiver's reason")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

func TestDecisionInvalidationNestedDecodeErrorsWrapSentinel(t *testing.T) {
	var i DecisionInvalidation
	payload := `{"id":"inv-1","invalidated_decision":{"decision_id":"dec-1"},"source":{"kind":"authority","ref":{"namespace":"","identifier":""}},"reason":"r","effective_at":"2026-07-26T10:00:00Z"}`
	err := json.Unmarshal([]byte(payload), &i)
	if err == nil {
		t.Fatal("malformed nested source accepted, want error")
	}
	if !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidDecisionInvalidation)
	}
}
