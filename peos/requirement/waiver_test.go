package requirement

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustAuthorityRef(t *testing.T, namespace, identifier string) core.AuthorityRef {
	t.Helper()
	ref, err := core.NewAuthorityRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustExtension(t *testing.T, namespace string, payload string) core.Extension {
	t.Helper()
	ext, err := core.NewExtension().With(namespace, json.RawMessage(payload))
	if err != nil {
		t.Fatal(err)
	}
	return ext
}

func mustWaiver(t *testing.T) Waiver {
	t.Helper()
	w, err := NewWaiver(
		mustRequirementRef(t, "REQ-1"),
		mustAuthorityRef(t, "org", "safety-board"),
		mustGovernanceActionFromDecisionOutcome(t, "DEC-1"),
		mustScope(t, "peos", "region=eu"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// --- Construction ---------------------------------------------------------

func TestNewWaiverValidMinimum(t *testing.T) {
	w := mustWaiver(t)
	if w.IsZero() {
		t.Error("valid Waiver IsZero() = true")
	}
	if !w.Extension().IsZero() {
		t.Error("minimum Waiver has non-zero Extension")
	}
}

func TestNewWaiverValidFullWithExtension(t *testing.T) {
	w := mustWaiver(t).WithExtension(mustExtension(t, "product-x", `{"rationale":"known gap, board-approved"}`))
	if w.Extension().IsZero() {
		t.Error("Extension() is zero after WithExtension")
	}
}

func TestNewWaiverZeroRequirementRejected(t *testing.T) {
	_, err := NewWaiver(
		core.RequirementRef{},
		mustAuthorityRef(t, "org", "safety-board"),
		mustGovernanceActionFromDecisionOutcome(t, "DEC-1"),
		mustScope(t, "peos", "region=eu"),
	)
	if !errors.Is(err, ErrInvalidWaiver) {
		t.Errorf("error = %v, want %v", err, ErrInvalidWaiver)
	}
}

func TestNewWaiverZeroAuthorityRejected(t *testing.T) {
	_, err := NewWaiver(
		mustRequirementRef(t, "REQ-1"),
		core.AuthorityRef{},
		mustGovernanceActionFromDecisionOutcome(t, "DEC-1"),
		mustScope(t, "peos", "region=eu"),
	)
	if !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func TestNewWaiverZeroGovernanceActionRejected(t *testing.T) {
	_, err := NewWaiver(
		mustRequirementRef(t, "REQ-1"),
		mustAuthorityRef(t, "org", "safety-board"),
		GovernanceAction{},
		mustScope(t, "peos", "region=eu"),
	)
	if !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestNewWaiverZeroScopeRejected(t *testing.T) {
	_, err := NewWaiver(
		mustRequirementRef(t, "REQ-1"),
		mustAuthorityRef(t, "org", "safety-board"),
		mustGovernanceActionFromDecisionOutcome(t, "DEC-1"),
		core.Scope{},
	)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestNewWaiverAcceptsProductMechanismGovernanceAction(t *testing.T) {
	w, err := NewWaiver(
		mustRequirementRef(t, "REQ-1"),
		mustAuthorityRef(t, "org", "safety-board"),
		mustGovernanceActionFromProductMechanism(t, "acme", "change-board-approval"),
		mustScope(t, "peos", "region=eu"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !w.GovernanceAction().IsProductMechanism() {
		t.Error("GovernanceAction() is not the product-mechanism arm")
	}
}

func TestNewWaiverAcceptsProductNamespacedScopeKind(t *testing.T) {
	w, err := NewWaiver(
		mustRequirementRef(t, "REQ-1"),
		mustAuthorityRef(t, "org", "safety-board"),
		mustGovernanceActionFromDecisionOutcome(t, "DEC-1"),
		mustScope(t, "acme", "site=plant-3"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if w.Scope().Expression() != "site=plant-3" {
		t.Errorf("Scope().Expression() = %q, want %q", w.Scope().Expression(), "site=plant-3")
	}
}

// --- Accessors and value semantics -----------------------------------------

func TestWaiverAccessorsReturnConstructorInputs(t *testing.T) {
	requirement := mustRequirementRef(t, "REQ-1")
	authority := mustAuthorityRef(t, "org", "safety-board")
	governanceAction := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	scope := mustScope(t, "peos", "region=eu")

	w, err := NewWaiver(requirement, authority, governanceAction, scope)
	if err != nil {
		t.Fatal(err)
	}
	if w.Requirement() != requirement {
		t.Errorf("Requirement() = %v, want %v", w.Requirement(), requirement)
	}
	if w.Authority() != authority {
		t.Errorf("Authority() = %v, want %v", w.Authority(), authority)
	}
	got, _ := w.GovernanceAction().AsDecisionOutcome()
	want, _ := governanceAction.AsDecisionOutcome()
	if got != want {
		t.Errorf("GovernanceAction() mismatch: got %v, want %v", got, want)
	}
	if !w.Scope().Equal(scope) {
		t.Errorf("Scope() = %v, want %v", w.Scope(), scope)
	}
}

func TestWaiverIsZero(t *testing.T) {
	var w Waiver
	if !w.IsZero() {
		t.Error("zero Waiver IsZero() = false")
	}
	if mustWaiver(t).IsZero() {
		t.Error("valid Waiver IsZero() = true")
	}
}

// --- Modifiers --------------------------------------------------------------

func TestWaiverWithExtensionSets(t *testing.T) {
	ext := mustExtension(t, "product-x", `{"rationale":"known gap"}`)
	w := mustWaiver(t).WithExtension(ext)
	if w.Extension().IsZero() {
		t.Error("Extension() is zero after WithExtension")
	}
}

func TestWaiverWithoutExtensionClears(t *testing.T) {
	ext := mustExtension(t, "product-x", `{"rationale":"known gap"}`)
	w := mustWaiver(t).WithExtension(ext).WithoutExtension()
	if !w.Extension().IsZero() {
		t.Error("Extension() is non-zero after WithoutExtension")
	}
}

func TestWaiverWithExtensionReplacementSemantics(t *testing.T) {
	first := mustExtension(t, "product-x", `{"rationale":"first"}`)
	second := mustExtension(t, "product-y", `{"rationale":"second"}`)
	w := mustWaiver(t).WithExtension(first).WithExtension(second)
	if _, ok := w.Extension().Get("product-x"); ok {
		t.Error("Extension() still carries the replaced namespace")
	}
	if _, ok := w.Extension().Get("product-y"); !ok {
		t.Error("Extension() missing the replacement namespace")
	}
}

func TestWaiverWithExtensionReceiverImmutability(t *testing.T) {
	original := mustWaiver(t)
	_ = original.WithExtension(mustExtension(t, "product-x", `{"rationale":"known gap"}`))
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}
}

// --- JSON ---------------------------------------------------------------

func TestWaiverJSONMinimumRoundTrip(t *testing.T) {
	w := mustWaiver(t)
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Waiver
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Requirement() != w.Requirement() {
		t.Errorf("Requirement() = %v, want %v", decoded.Requirement(), w.Requirement())
	}
	if decoded.Authority() != w.Authority() {
		t.Errorf("Authority() = %v, want %v", decoded.Authority(), w.Authority())
	}
	if !decoded.Scope().Equal(w.Scope()) {
		t.Errorf("Scope() = %v, want %v", decoded.Scope(), w.Scope())
	}
}

func TestWaiverJSONFullRoundTripWithExtension(t *testing.T) {
	w := mustWaiver(t).WithExtension(mustExtension(t, "product-x", `{"rationale":"known gap"}`))
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Waiver
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	payload, ok := decoded.Extension().Get("product-x")
	if !ok {
		t.Fatal("decoded Extension missing product-x namespace")
	}
	if string(payload) != `{"rationale":"known gap"}` {
		t.Errorf("decoded Extension payload = %s, want %s", payload, `{"rationale":"known gap"}`)
	}
}

func TestWaiverJSONTopLevelKeys(t *testing.T) {
	w := mustWaiver(t).WithExtension(mustExtension(t, "product-x", `{"rationale":"known gap"}`))
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"requirement", "authority", "governance_action", "scope", "extension"}
	if len(raw) != len(want) {
		t.Errorf("Waiver wire form has %d top-level keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, key := range want {
		if _, present := raw[key]; !present {
			t.Errorf("Waiver wire form missing key %q", key)
		}
	}
	forbidden := []string{"relation", "type", "id", "provenance", "lifecycle_consequence"}
	for _, key := range forbidden {
		if _, present := raw[key]; present {
			t.Errorf("Waiver wire form unexpectedly carries key %q", key)
		}
	}
}

func TestWaiverJSONExtensionOmittedWhenAbsent(t *testing.T) {
	data, err := json.Marshal(mustWaiver(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["extension"]; present {
		t.Error(`"extension" key present despite absent Extension`)
	}
}

func TestWaiverZeroMarshalRejected(t *testing.T) {
	var w Waiver
	if _, err := json.Marshal(w); !errors.Is(err, ErrInvalidWaiver) {
		t.Errorf("error = %v, want %v", err, ErrInvalidWaiver)
	}
}

func TestWaiverJSONMissingRequirementRejected(t *testing.T) {
	payload := `{"authority":{"namespace":"org","identifier":"safety-board"},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"scope":{"kind":"peos:deployment","expression":"region=eu"}}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); !errors.Is(err, ErrInvalidWaiver) {
		t.Errorf("error = %v, want %v", err, ErrInvalidWaiver)
	}
}

func TestWaiverJSONMissingAuthorityRejected(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"scope":{"kind":"peos:deployment","expression":"region=eu"}}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func TestWaiverJSONMissingGovernanceActionRejected(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"authority":{"namespace":"org","identifier":"safety-board"},"scope":{"kind":"peos:deployment","expression":"region=eu"}}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestWaiverJSONMissingScopeRejected(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"authority":{"namespace":"org","identifier":"safety-board"},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}}}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

// TestWaiverJSONExplicitNullMandatoryFieldsRejected proves that an explicit
// JSON null for any of Waiver's four mandatory keys is rejected rather than
// silently treated as absent. requirement and authority surface
// ErrInvalidWaiver (their nested types have no "null" literal check of
// their own, so the zero value they decode to is caught by NewWaiver's own
// validation, which for requirement uses ErrInvalidWaiver directly);
// governance_action and scope surface their own owning sentinel, since
// GovernanceAction and core.Scope reject a missing/null "ref" or
// zero-value payload from inside their own UnmarshalJSON.
func TestWaiverJSONExplicitNullMandatoryFieldsRejected(t *testing.T) {
	base := map[string]string{
		"requirement":       `{"artifact_id":"REQ-1"}`,
		"authority":         `{"namespace":"org","identifier":"safety-board"}`,
		"governance_action": `{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}}`,
		"scope":             `{"kind":"peos:deployment","expression":"region=eu"}`,
	}
	cases := []struct {
		field string
		want  error
	}{
		{"requirement", ErrInvalidWaiver},
		{"authority", ErrInvalidWaiver},
		{"governance_action", ErrInvalidGovernanceAction},
		{"scope", core.ErrInvalidScope},
	}
	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			payload := "{"
			first := true
			for key, value := range base {
				if !first {
					payload += ","
				}
				first = false
				if key == tc.field {
					payload += `"` + key + `":null`
				} else {
					payload += `"` + key + `":` + value
				}
			}
			payload += "}"
			var w Waiver
			err := json.Unmarshal([]byte(payload), &w)
			if err == nil {
				t.Fatal("explicit null accepted, want error")
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want wrapping %v", err, tc.want)
			}
		})
	}
}

func TestWaiverJSONExtensionNullMeansAbsent(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"authority":{"namespace":"org","identifier":"safety-board"},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"scope":{"kind":"peos:deployment","expression":"region=eu"},"extension":null}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); err != nil {
		t.Fatal(err)
	}
	if !w.Extension().IsZero() {
		t.Error("null extension decoded to non-zero Extension")
	}
}

func TestWaiverJSONUnknownFieldIgnored(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"authority":{"namespace":"org","identifier":"safety-board"},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"scope":{"kind":"peos:deployment","expression":"region=eu"},"unknown_field":"ignored"}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); err != nil {
		t.Fatal(err)
	}
	if w.IsZero() {
		t.Error("valid payload with unknown field decoded to zero Waiver")
	}
}

func TestWaiverJSONMalformedGovernanceActionRejected(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"authority":{"namespace":"org","identifier":"safety-board"},"governance_action":{"kind":"unrecognized","ref":{"decision_id":"DEC-1"}},"scope":{"kind":"peos:deployment","expression":"region=eu"}}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestWaiverJSONMalformedScopeRejected(t *testing.T) {
	payload := `{"requirement":{"artifact_id":"REQ-1"},"authority":{"namespace":"org","identifier":"safety-board"},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"scope":{"kind":"peos:deployment","expression":""}}`
	var w Waiver
	if err := json.Unmarshal([]byte(payload), &w); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestWaiverUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := mustWaiver(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Requirement() != original.Requirement() {
		t.Error("failed Unmarshal changed receiver's Requirement")
	}
	if receiver.Authority() != original.Authority() {
		t.Error("failed Unmarshal changed receiver's Authority")
	}
	if !receiver.Scope().Equal(original.Scope()) {
		t.Error("failed Unmarshal changed receiver's Scope")
	}
	if receiver.Extension().IsZero() != original.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's Extension presence")
	}
}

func TestWaiverConstructorUnmarshalEquivalence(t *testing.T) {
	constructed := mustWaiver(t)
	data, err := json.Marshal(constructed)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Waiver
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.IsZero() {
		t.Error("decoded Waiver is zero")
	}
}

// --- Cardinality ------------------------------------------------------------

func TestWaiverCardinalityTwoWaiversSameSubjectBothConstruct(t *testing.T) {
	subject := mustRequirementRef(t, "REQ-1")
	first, err := NewWaiver(subject, mustAuthorityRef(t, "org", "safety-board"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), mustScope(t, "peos", "region=eu"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewWaiver(subject, mustAuthorityRef(t, "org", "change-board"), mustGovernanceActionFromDecisionOutcome(t, "DEC-2"), mustScope(t, "peos", "region=us"))
	if err != nil {
		t.Fatal(err)
	}
	if first.Scope().Equal(second.Scope()) {
		t.Error("test setup produced identical scopes, cannot distinguish the two Waivers")
	}
	if first.Requirement() != second.Requirement() {
		t.Error("both Waivers should name the same subject Requirement")
	}
}

// --- Absence audit -----------------------------------------------------

// TestWaiverHasNoRelationshipOrIdentityMethods is a structural absence
// audit proving Waiver carries no identity, no lifecycle, no provenance,
// and is not a relationship: PEOS-005 §27 defines none of these, and
// PEOS-008 :520 confirms PEOS-005 defines no Waiver identity, lifecycle,
// revocation, or historical model. Note Provenance is deliberately absent
// too, per the accepted correction to this packet's architecture (Waiver
// carries no core.Provenance field).
func TestWaiverHasNoRelationshipOrIdentityMethods(t *testing.T) {
	forbidden := map[string]bool{
		"Relation":             true,
		"Source":               true,
		"Target":               true,
		"RelationType":         true,
		"ArtifactID":           true,
		"RevisionID":           true,
		"ID":                   true,
		"Ref":                  true,
		"LifecycleConsequence": true,
		"Provenance":           true,
	}
	typ := reflect.TypeOf(Waiver{})
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		if forbidden[name] {
			t.Errorf("Waiver has forbidden method %q", name)
		}
	}
}

// TestWaiverJSONHasNoProvenanceKey proves the wire form carries no
// "provenance" key, matching the accepted correction removing
// core.Provenance from this type entirely.
func TestWaiverJSONHasNoProvenanceKey(t *testing.T) {
	data, err := json.Marshal(mustWaiver(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["provenance"]; present {
		t.Error(`unexpected "provenance" key present in Waiver wire form`)
	}
}
