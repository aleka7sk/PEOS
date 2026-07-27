package template

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- ParameterType -----------------------------------------------------------

func TestVocabularyParameterType(t *testing.T) {
	v := mustVocabularyValue(t, "product", "string")
	pt, err := NewVocabularyParameterType(v)
	if err != nil {
		t.Fatal(err)
	}
	if pt.IsZero() {
		t.Error("valid ParameterType reports IsZero() = true")
	}
	if pt.Kind() != "vocabulary" {
		t.Errorf("Kind() = %q", pt.Kind())
	}
	got, ok := pt.Vocabulary()
	if !ok || got != v {
		t.Error("Vocabulary() mismatch")
	}
	if _, _, ok := pt.External(); ok {
		t.Error("a vocabulary ParameterType returned external state")
	}
}

func TestExternalParameterType(t *testing.T) {
	authority := mustVocabularyValue(t, "product", "json-schema-registry")
	pt, err := NewExternalParameterType(authority, "  https://types.example/person  ")
	if err != nil {
		t.Fatal(err)
	}
	if pt.Kind() != "external" {
		t.Errorf("Kind() = %q", pt.Kind())
	}
	gotAuthority, locator, ok := pt.External()
	if !ok || gotAuthority != authority {
		t.Error("External() authority mismatch")
	}
	if locator != "https://types.example/person" {
		t.Errorf("locator = %q, want trimmed", locator)
	}
	if _, ok := pt.Vocabulary(); ok {
		t.Error("an external ParameterType returned a vocabulary value")
	}
}

func TestParameterTypeRejections(t *testing.T) {
	if _, err := NewVocabularyParameterType(core.VocabularyValue{}); !errors.Is(err, ErrInvalidParameterType) {
		t.Errorf("zero vocabulary: error = %v, want %v", err, ErrInvalidParameterType)
	}
	if _, err := NewExternalParameterType(core.VocabularyValue{}, "locator"); !errors.Is(err, ErrInvalidParameterType) {
		t.Errorf("zero authority: error = %v, want %v", err, ErrInvalidParameterType)
	}
	authority := mustVocabularyValue(t, "product", "registry")
	if _, err := NewExternalParameterType(authority, ""); !errors.Is(err, ErrInvalidParameterType) {
		t.Errorf("empty locator: error = %v, want %v", err, ErrInvalidParameterType)
	}
	if _, err := NewExternalParameterType(authority, "   "); !errors.Is(err, ErrInvalidParameterType) {
		t.Errorf("whitespace-only locator: error = %v, want %v", err, ErrInvalidParameterType)
	}

	var zero ParameterType
	if !zero.IsZero() {
		t.Error("zero-value ParameterType.IsZero() = false, want true")
	}
	if zero.Kind() != "" {
		t.Errorf("zero Kind() = %q, want empty", zero.Kind())
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidParameterType) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidParameterType)
	}
}

func TestParameterTypeJSONRoundTrip(t *testing.T) {
	vocab, err := NewVocabularyParameterType(mustVocabularyValue(t, "product", "string"))
	if err != nil {
		t.Fatal(err)
	}
	external, err := NewExternalParameterType(mustVocabularyValue(t, "product", "registry"), "https://types.example/person")
	if err != nil {
		t.Fatal(err)
	}

	for name, pt := range map[string]ParameterType{"vocabulary": vocab, "external": external} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(pt)
			if err != nil {
				t.Fatal(err)
			}
			var decoded ParameterType
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != pt {
				t.Errorf("round trip mismatch: got %+v, want %+v", decoded, pt)
			}
		})
	}
}

// TestParameterTypeUnmarshalRejections covers the full closed-union matrix:
// neither arm, both arms, unknown discriminator, and a selected arm missing its
// own payload.
func TestParameterTypeUnmarshalRejections(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing kind", `{}`},
		{"unknown kind", `{"kind":"primitive","value":"product:string"}`},
		{"explicit null", `null`},
		{"vocabulary missing value", `{"kind":"vocabulary"}`},
		{"vocabulary null value", `{"kind":"vocabulary","value":null}`},
		{"vocabulary carrying authority", `{"kind":"vocabulary","value":"product:string","authority":"product:registry"}`},
		{"vocabulary carrying locator", `{"kind":"vocabulary","value":"product:string","locator":"x"}`},
		{"external missing authority", `{"kind":"external","locator":"x"}`},
		{"external null authority", `{"kind":"external","authority":null,"locator":"x"}`},
		{"external missing locator", `{"kind":"external","authority":"product:registry"}`},
		{"external empty locator", `{"kind":"external","authority":"product:registry","locator":""}`},
		{"external carrying vocabulary value", `{"kind":"external","authority":"product:registry","locator":"x","value":"product:string"}`},
		{"malformed JSON", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pt ParameterType
			if err := json.Unmarshal([]byte(tt.json), &pt); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

func TestParameterTypeUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	pt := mustParameterType(t)
	before := pt
	if err := json.Unmarshal([]byte(`{"kind":"bogus"}`), &pt); err == nil {
		t.Fatal("expected error")
	}
	if pt != before {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- Parameter ---------------------------------------------------------------

func TestNewParameter(t *testing.T) {
	key := mustLocalKey(t, "name")
	pt := mustParameterType(t)

	p, err := NewParameter(key, pt, true)
	if err != nil {
		t.Fatal(err)
	}
	if p.IsZero() {
		t.Error("valid Parameter reports IsZero() = true")
	}
	if p.Key() != key {
		t.Error("Key() mismatch")
	}
	if p.Type() != pt {
		t.Error("Type() mismatch")
	}
	if !p.Required() {
		t.Error("Required() = false, want true")
	}
	if _, ok := p.Description(); ok {
		t.Error("new Parameter should have no description")
	}
	if p.ForbidsDefaultResolution() {
		t.Error("new Parameter should permit default resolution")
	}
}

// TestNewParameterRequiredFalseIsStatedState confirms required == false is a
// stated fact, not an omission: PEOS-009 lists "required parameters" among the
// items every Template Artifact Revision SHALL identify.
func TestNewParameterRequiredFalseIsStatedState(t *testing.T) {
	p, err := NewParameter(mustLocalKey(t, "owner"), mustParameterType(t), false)
	if err != nil {
		t.Fatal(err)
	}
	if p.Required() {
		t.Error("Required() = true, want false")
	}
	if p.IsZero() {
		t.Error("an explicitly optional Parameter is still a valid, non-zero Parameter")
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	raw, ok := m["required"]
	if !ok {
		t.Fatal(`wire form omits "required"; explicitly optional would be indistinguishable from unstated`)
	}
	if string(raw) != "false" {
		t.Errorf(`"required" = %s, want false`, raw)
	}
}

func TestNewParameterRejections(t *testing.T) {
	if _, err := NewParameter(core.LocalKey{}, mustParameterType(t), true); !errors.Is(err, ErrInvalidTemplateParameter) {
		t.Errorf("zero key: error = %v, want %v", err, ErrInvalidTemplateParameter)
	}
	if _, err := NewParameter(mustLocalKey(t, "name"), ParameterType{}, true); !errors.Is(err, ErrInvalidParameterType) {
		t.Errorf("zero parameter type: error = %v, want %v", err, ErrInvalidParameterType)
	}
}

func TestParameterModifiers(t *testing.T) {
	p := mustParameter(t, "name", true)

	p2, err := p.WithDescription("  the subject's display name  ")
	if err != nil {
		t.Fatal(err)
	}
	desc, ok := p2.Description()
	if !ok || desc != "the subject's display name" {
		t.Errorf("Description() = %q, %v; want trimmed value", desc, ok)
	}
	if _, ok := p2.WithoutDescription().Description(); ok {
		t.Error("WithoutDescription did not clear the description")
	}
	if _, err := p.WithDescription("   "); !errors.Is(err, ErrInvalidTemplateParameter) {
		t.Errorf("whitespace-only description: error = %v, want %v", err, ErrInvalidTemplateParameter)
	}

	forbidden := p.WithForbiddenDefaultResolution()
	if !forbidden.ForbidsDefaultResolution() {
		t.Error("WithForbiddenDefaultResolution did not set the flag")
	}
	if forbidden.WithPermittedDefaultResolution().ForbidsDefaultResolution() {
		t.Error("WithPermittedDefaultResolution did not clear the flag")
	}
	// The original must be untouched -- every modifier returns a copy.
	if p.ForbidsDefaultResolution() {
		t.Error("WithForbiddenDefaultResolution mutated its receiver")
	}
}

func TestParameterJSONRoundTrip(t *testing.T) {
	p, err := mustParameter(t, "name", true).WithDescription("display name")
	if err != nil {
		t.Fatal(err)
	}
	p = p.WithForbiddenDefaultResolution()

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Parameter
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != p {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, p)
	}
}

func TestParameterMarshalZeroAndDecodeRejections(t *testing.T) {
	var zero Parameter
	if !zero.IsZero() {
		t.Error("zero-value Parameter.IsZero() = false, want true")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidTemplateParameter) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidTemplateParameter)
	}

	tests := []struct {
		name string
		json string
	}{
		{"missing key", `{"parameter_type":{"kind":"vocabulary","value":"product:string"},"required":true}`},
		{"empty key", `{"key":"","parameter_type":{"kind":"vocabulary","value":"product:string"},"required":true}`},
		{"missing parameter type", `{"key":"name","required":true}`},
		{"null parameter type", `{"key":"name","parameter_type":null,"required":true}`},
		{"whitespace description", `{"key":"name","parameter_type":{"kind":"vocabulary","value":"product:string"},"required":true,"description":"   "}`},
		{"malformed JSON", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parameter
			if err := json.Unmarshal([]byte(tt.json), &p); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

func TestParameterUnmarshalToleratesUnknownFields(t *testing.T) {
	var p Parameter
	payload := `{"key":"name","parameter_type":{"kind":"vocabulary","value":"product:string"},"required":true,"product_hint":"x"}`
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatal(err)
	}
	if p.Key() != mustLocalKey(t, "name") {
		t.Error("unknown-field tolerance lost the key")
	}
}

func TestParameterUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	p := mustParameter(t, "name", true)
	before := p
	if err := json.Unmarshal([]byte(`{"key":""}`), &p); err == nil {
		t.Fatal("expected error")
	}
	if p != before {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestParameterNoForbiddenWireKeys(t *testing.T) {
	p := mustParameter(t, "name", true)
	forbidden := []string{
		"id", "ref", "revision", "lifecycle", "state", "status",
		"value", "current_value", "resolved_value", "binding",
		"provenance", "authority", "scope",
	}
	assertNoWireKeys(t, "Parameter", p, forbidden)
}

// --- ParameterDefault --------------------------------------------------------

func TestNewParameterDefault(t *testing.T) {
	d, err := NewParameterDefault(mustLocalKey(t, "name"), "  anonymous  ")
	if err != nil {
		t.Fatal(err)
	}
	if d.IsZero() {
		t.Error("valid ParameterDefault reports IsZero() = true")
	}
	if d.Parameter() != mustLocalKey(t, "name") {
		t.Error("Parameter() mismatch")
	}
	if d.Value() != "anonymous" {
		t.Errorf("Value() = %q, want trimmed", d.Value())
	}
}

func TestNewParameterDefaultRejections(t *testing.T) {
	if _, err := NewParameterDefault(core.LocalKey{}, "x"); !errors.Is(err, ErrInvalidParameterDefault) {
		t.Errorf("zero parameter key: error = %v, want %v", err, ErrInvalidParameterDefault)
	}
	if _, err := NewParameterDefault(mustLocalKey(t, "name"), ""); !errors.Is(err, ErrInvalidParameterDefault) {
		t.Errorf("empty value: error = %v, want %v", err, ErrInvalidParameterDefault)
	}
	if _, err := NewParameterDefault(mustLocalKey(t, "name"), "   "); !errors.Is(err, ErrInvalidParameterDefault) {
		t.Errorf("whitespace-only value: error = %v, want %v", err, ErrInvalidParameterDefault)
	}
}

func TestParameterDefaultJSON(t *testing.T) {
	d := mustParameterDefault(t, "name", "anonymous")
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"parameter":"name","value":"anonymous"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded ParameterDefault
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != d {
		t.Error("round trip mismatch")
	}

	var zero ParameterDefault
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidParameterDefault) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidParameterDefault)
	}

	before := d
	for _, payload := range []string{`{}`, `{"parameter":"name"}`, `{"parameter":"","value":"x"}`, `null`, `not json`} {
		if err := json.Unmarshal([]byte(payload), &d); err == nil {
			t.Errorf("%s accepted, want error", payload)
		}
	}
	if d != before {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- ConstraintTarget --------------------------------------------------------

func TestConstraintTargetParameterArm(t *testing.T) {
	target, err := NewParameterConstraintTarget(mustLocalKey(t, "name"))
	if err != nil {
		t.Fatal(err)
	}
	if target.IsZero() {
		t.Error("valid ConstraintTarget reports IsZero() = true")
	}
	if target.Kind() != "parameter" {
		t.Errorf("Kind() = %q", target.Kind())
	}
	key, ok := target.Parameter()
	if !ok || key != mustLocalKey(t, "name") {
		t.Error("Parameter() mismatch")
	}
	if _, ok := target.GeneratedContent(); ok {
		t.Error("a parameter target returned generated content")
	}
}

func TestConstraintTargetGeneratedContentArm(t *testing.T) {
	target, err := NewGeneratedContentConstraintTarget("  the requirement statement  ")
	if err != nil {
		t.Fatal(err)
	}
	if target.Kind() != "generated_content" {
		t.Errorf("Kind() = %q", target.Kind())
	}
	descriptor, ok := target.GeneratedContent()
	if !ok || descriptor != "the requirement statement" {
		t.Errorf("GeneratedContent() = %q, %v; want trimmed value", descriptor, ok)
	}
	if _, ok := target.Parameter(); ok {
		t.Error("a generated-content target returned a parameter key")
	}
}

func TestConstraintTargetRejections(t *testing.T) {
	if _, err := NewParameterConstraintTarget(core.LocalKey{}); !errors.Is(err, ErrInvalidConstraintTarget) {
		t.Errorf("zero parameter key: error = %v, want %v", err, ErrInvalidConstraintTarget)
	}
	if _, err := NewGeneratedContentConstraintTarget("   "); !errors.Is(err, ErrInvalidConstraintTarget) {
		t.Errorf("whitespace-only descriptor: error = %v, want %v", err, ErrInvalidConstraintTarget)
	}

	var zero ConstraintTarget
	if !zero.IsZero() {
		t.Error("zero-value ConstraintTarget.IsZero() = false, want true")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidConstraintTarget) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidConstraintTarget)
	}
}

func TestConstraintTargetJSONRoundTrip(t *testing.T) {
	param, err := NewParameterConstraintTarget(mustLocalKey(t, "name"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := NewGeneratedContentConstraintTarget("the statement")
	if err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]ConstraintTarget{"parameter": param, "generated_content": generated} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(target)
			if err != nil {
				t.Fatal(err)
			}
			var decoded ConstraintTarget
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != target {
				t.Errorf("round trip mismatch: got %+v, want %+v", decoded, target)
			}
		})
	}
}

func TestConstraintTargetUnmarshalRejections(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing kind", `{}`},
		{"unknown kind", `{"kind":"artifact","parameter":"name"}`},
		{"explicit null", `null`},
		{"parameter missing key", `{"kind":"parameter"}`},
		{"parameter empty key", `{"kind":"parameter","parameter":""}`},
		{"parameter carrying generated content", `{"kind":"parameter","parameter":"name","generated_content":"x"}`},
		{"generated content missing descriptor", `{"kind":"generated_content"}`},
		{"generated content carrying parameter", `{"kind":"generated_content","generated_content":"x","parameter":"name"}`},
		{"malformed JSON", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target ConstraintTarget
			if err := json.Unmarshal([]byte(tt.json), &target); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

// --- ParameterConstraint -----------------------------------------------------

func TestNewParameterConstraint(t *testing.T) {
	v := mustParameterConstraint(t, "c1", "name")
	if v.IsZero() {
		t.Error("valid ParameterConstraint reports IsZero() = true")
	}
	if v.Key() != mustLocalKey(t, "c1") {
		t.Error("Key() mismatch")
	}
	if key, ok := v.Target().Parameter(); !ok || key != mustLocalKey(t, "name") {
		t.Error("Target() mismatch")
	}
	if v.Rule() != "value must be non-empty" {
		t.Errorf("Rule() = %q", v.Rule())
	}
	if v.Scope() != mustScope(t, "env=prod") {
		t.Error("Scope() mismatch")
	}
	if !v.EvaluationPoint().Equal(mustEvaluationPoint(t)) {
		t.Error("EvaluationPoint() mismatch")
	}
	if !v.FailureSemantics().Equal(mustFailureSemantics(t)) {
		t.Error("FailureSemantics() mismatch")
	}
	if _, ok := v.Authority(); ok {
		t.Error("new ParameterConstraint should have no authority")
	}
}

func TestNewParameterConstraintRejections(t *testing.T) {
	target, err := NewParameterConstraintTarget(mustLocalKey(t, "name"))
	if err != nil {
		t.Fatal(err)
	}
	key := mustLocalKey(t, "c1")
	scope := mustScope(t, "env=prod")
	ep := mustEvaluationPoint(t)
	fs := mustFailureSemantics(t)

	t.Run("zero key", func(t *testing.T) {
		if _, err := NewParameterConstraint(core.LocalKey{}, target, "rule", scope, ep, fs); !errors.Is(err, ErrInvalidParameterConstraint) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterConstraint)
		}
	})
	t.Run("zero target", func(t *testing.T) {
		if _, err := NewParameterConstraint(key, ConstraintTarget{}, "rule", scope, ep, fs); !errors.Is(err, ErrInvalidConstraintTarget) {
			t.Errorf("error = %v, want %v", err, ErrInvalidConstraintTarget)
		}
	})
	t.Run("empty rule", func(t *testing.T) {
		if _, err := NewParameterConstraint(key, target, "   ", scope, ep, fs); !errors.Is(err, ErrInvalidParameterConstraint) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterConstraint)
		}
	})
	t.Run("zero scope surfaces the core sentinel", func(t *testing.T) {
		if _, err := NewParameterConstraint(key, target, "rule", core.Scope{}, ep, fs); !errors.Is(err, core.ErrInvalidScope) {
			t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
		}
	})
	t.Run("zero evaluation point", func(t *testing.T) {
		if _, err := NewParameterConstraint(key, target, "rule", scope, ConstraintEvaluationPoint{}, fs); !errors.Is(err, ErrInvalidParameterConstraint) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterConstraint)
		}
	})
	t.Run("zero failure semantics", func(t *testing.T) {
		if _, err := NewParameterConstraint(key, target, "rule", scope, ep, ConstraintFailureSemantics{}); !errors.Is(err, ErrInvalidParameterConstraint) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterConstraint)
		}
	})
}

func TestParameterConstraintAuthority(t *testing.T) {
	v := mustParameterConstraint(t, "c1", "name")
	withAuthority, err := v.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withAuthority.Authority()
	if !ok || got != mustAuthority(t) {
		t.Error("WithAuthority did not set authority")
	}
	if _, ok := withAuthority.WithoutAuthority().Authority(); ok {
		t.Error("WithoutAuthority did not clear authority")
	}
	if _, err := v.WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidParameterConstraint) {
		t.Errorf("zero authority: error = %v, want %v", err, ErrInvalidParameterConstraint)
	}
	if _, ok := v.Authority(); ok {
		t.Error("WithAuthority mutated its receiver")
	}
}

func TestParameterConstraintJSONRoundTrip(t *testing.T) {
	v, err := mustParameterConstraint(t, "c1", "name").WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ParameterConstraint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
	if _, ok := decoded.Authority(); !ok {
		t.Error("round trip lost authority")
	}
}

func TestParameterConstraintJSONRejections(t *testing.T) {
	var zero ParameterConstraint
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidParameterConstraint) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidParameterConstraint)
	}

	base := mustParameterConstraint(t, "c1", "name")
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
		{"empty key", "key", json.RawMessage(`""`)},
		{"null target", "target", json.RawMessage(`null`)},
		{"empty rule", "rule", json.RawMessage(`""`)},
		{"whitespace rule", "rule", json.RawMessage(`"   "`)},
		{"null scope", "scope", json.RawMessage(`null`)},
		{"null evaluation point", "evaluation_point", json.RawMessage(`null`)},
		{"null failure semantics", "failure_semantics", json.RawMessage(`null`)},
		{"null authority", "authority", json.RawMessage(`null`)},
		{"malformed authority", "authority", json.RawMessage(`123`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, val := range m {
				mCopy[k] = val
			}
			mCopy[tt.key] = tt.value
			modified, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var v ParameterConstraint
			if err := json.Unmarshal(modified, &v); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}

	t.Run("absent authority is valid", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, val := range m {
			mCopy[k] = val
		}
		delete(mCopy, "authority")
		modified, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var v ParameterConstraint
		if err := json.Unmarshal(modified, &v); err != nil {
			t.Errorf("absent authority rejected: %v", err)
		}
		if _, ok := v.Authority(); ok {
			t.Error("absent authority decoded as set")
		}
	})
}

func TestParameterConstraintUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	v := mustParameterConstraint(t, "c1", "name")
	before, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"key":""}`), &v); err == nil {
		t.Fatal("expected error")
	}
	after, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestParameterConstraintNoForbiddenWireKeys(t *testing.T) {
	v := mustParameterConstraint(t, "c1", "name")
	forbidden := []string{
		"id", "ref", "revision", "lifecycle", "state", "status",
		"result", "outcome", "satisfied", "violated", "evaluated",
	}
	assertNoWireKeys(t, "ParameterConstraint", v, forbidden)
}

// --- constraint vocabularies -------------------------------------------------

func TestConstraintVocabularies(t *testing.T) {
	v := mustVocabularyValue(t, "product", "pre-generation")
	ep := NewConstraintEvaluationPoint(v)
	if ep.IsZero() {
		t.Error("valid ConstraintEvaluationPoint reports IsZero() = true")
	}
	if ep.Value() != v || ep.String() != v.String() {
		t.Error("ConstraintEvaluationPoint accessors mismatch")
	}
	if !ep.Equal(NewConstraintEvaluationPoint(v)) {
		t.Error("Equal() = false for identical values")
	}
	var zeroEP ConstraintEvaluationPoint
	if !zeroEP.IsZero() {
		t.Error("zero-value ConstraintEvaluationPoint.IsZero() = false, want true")
	}

	fs := NewConstraintFailureSemantics(v)
	if fs.IsZero() || fs.Value() != v || fs.String() != v.String() {
		t.Error("ConstraintFailureSemantics accessors mismatch")
	}
	if !fs.Equal(NewConstraintFailureSemantics(v)) {
		t.Error("Equal() = false for identical values")
	}
	var zeroFS ConstraintFailureSemantics
	if !zeroFS.IsZero() {
		t.Error("zero-value ConstraintFailureSemantics.IsZero() = false, want true")
	}

	for name, value := range map[string]any{"evaluation point": ep, "failure semantics": fs} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != `"product:pre-generation"` {
				t.Errorf("Marshal = %s", data)
			}
		})
	}

	var decodedEP ConstraintEvaluationPoint
	if err := json.Unmarshal([]byte(`"product:pre-generation"`), &decodedEP); err != nil {
		t.Fatal(err)
	}
	if !decodedEP.Equal(ep) {
		t.Error("ConstraintEvaluationPoint round trip mismatch")
	}
	if err := json.Unmarshal([]byte(`"no-namespace"`), &decodedEP); !errors.Is(err, core.ErrInvalidVocabularyValue) {
		t.Errorf("malformed vocabulary value: error = %v, want %v", err, core.ErrInvalidVocabularyValue)
	}

	var decodedFS ConstraintFailureSemantics
	if err := json.Unmarshal([]byte(`"product:pre-generation"`), &decodedFS); err != nil {
		t.Fatal(err)
	}
	if !decodedFS.Equal(fs) {
		t.Error("ConstraintFailureSemantics round trip mismatch")
	}
}

// --- CompatibilityDeclaration ------------------------------------------------

func TestNewCompatibilityDeclaration(t *testing.T) {
	types := []core.ArtifactType{mustArtifactType(t, "requirement")}
	d, err := NewCompatibilityDeclaration(types, "  all required parameters supplied  ", "  product contract v1  ")
	if err != nil {
		t.Fatal(err)
	}
	if d.IsZero() {
		t.Error("valid CompatibilityDeclaration reports IsZero() = true")
	}
	if got := d.ApplicableArtifactTypes(); len(got) != 1 || got[0] != types[0] {
		t.Errorf("ApplicableArtifactTypes() = %v", got)
	}
	if d.ParameterContract() != "all required parameters supplied" {
		t.Errorf("ParameterContract() = %q, want trimmed", d.ParameterContract())
	}
	if d.ProductContract() != "product contract v1" {
		t.Errorf("ProductContract() = %q, want trimmed", d.ProductContract())
	}
	if len(d.ApplicableRevisions()) != 0 {
		t.Error("new declaration should have no applicable revisions")
	}
	if _, ok := d.MigrationRequirements(); ok {
		t.Error("new declaration should have no migration requirements")
	}
}

func TestNewCompatibilityDeclarationRejections(t *testing.T) {
	types := []core.ArtifactType{mustArtifactType(t, "requirement")}

	if _, err := NewCompatibilityDeclaration(nil, "p", "c"); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("zero applicable artifact types: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}
	if _, err := NewCompatibilityDeclaration([]core.ArtifactType{{}}, "p", "c"); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("zero-value element: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}
	dup := []core.ArtifactType{mustArtifactType(t, "requirement"), mustArtifactType(t, "requirement")}
	if _, err := NewCompatibilityDeclaration(dup, "p", "c"); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("duplicate element: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}
	if _, err := NewCompatibilityDeclaration(types, "   ", "c"); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("empty parameter contract: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}
	if _, err := NewCompatibilityDeclaration(types, "p", "   "); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("empty product contract: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}

	var zero CompatibilityDeclaration
	if !zero.IsZero() {
		t.Error("zero-value CompatibilityDeclaration.IsZero() = false, want true")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}
}

func TestCompatibilityDeclarationModifiers(t *testing.T) {
	d := mustCompatibilityDeclaration(t)
	ref := mustTemplateRevisionRef(t, "TPL-2", "REV-1")

	d2, err := d.WithApplicableRevisions([]core.TemplateArtifactRevisionRef{ref})
	if err != nil {
		t.Fatal(err)
	}
	if got := d2.ApplicableRevisions(); len(got) != 1 || got[0] != ref {
		t.Errorf("ApplicableRevisions() = %v", got)
	}
	returned := d2.ApplicableRevisions()
	returned[0] = mustTemplateRevisionRef(t, "TPL-9", "REV-9")
	if d2.ApplicableRevisions()[0].ArtifactID() == mustArtifactID(t, "TPL-9") {
		t.Error("ApplicableRevisions() accessor did not return a defensive copy")
	}
	if _, err := d.WithApplicableRevisions([]core.TemplateArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("zero applicable revision: error = %v, want %v", err, ErrInvalidTemplateContent)
	}
	if _, err := d.WithApplicableRevisions([]core.TemplateArtifactRevisionRef{ref, ref}); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("duplicate applicable revision: error = %v, want %v", err, ErrInvalidTemplateContent)
	}

	d3, err := d.WithMigrationRequirements("  migrate v1 parameters to v2  ")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := d3.MigrationRequirements()
	if !ok || got != "migrate v1 parameters to v2" {
		t.Errorf("MigrationRequirements() = %q, %v; want trimmed value", got, ok)
	}
	if _, ok := d3.WithoutMigrationRequirements().MigrationRequirements(); ok {
		t.Error("WithoutMigrationRequirements did not clear the value")
	}
	if _, err := d.WithMigrationRequirements("   "); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
		t.Errorf("whitespace-only migration requirements: error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
	}
}

func TestCompatibilityDeclarationDefensiveCopy(t *testing.T) {
	types := []core.ArtifactType{mustArtifactType(t, "requirement")}
	d, err := NewCompatibilityDeclaration(types, "p", "c")
	if err != nil {
		t.Fatal(err)
	}
	types[0] = mustArtifactType(t, "mutated")
	if d.ApplicableArtifactTypes()[0].String() == "peos:mutated" {
		t.Error("constructor did not defensively copy applicable artifact types")
	}
	returned := d.ApplicableArtifactTypes()
	returned[0] = mustArtifactType(t, "mutated-again")
	if d.ApplicableArtifactTypes()[0].String() == "peos:mutated-again" {
		t.Error("ApplicableArtifactTypes() accessor did not return a defensive copy")
	}
}

func TestCompatibilityDeclarationJSONRoundTrip(t *testing.T) {
	d, err := mustCompatibilityDeclaration(t).WithApplicableRevisions(
		[]core.TemplateArtifactRevisionRef{mustTemplateRevisionRef(t, "TPL-2", "REV-1")},
	)
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithMigrationRequirements("migrate v1 to v2")
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CompatibilityDeclaration
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
	if _, ok := decoded.MigrationRequirements(); !ok {
		t.Error("round trip lost migration requirements")
	}
}

func TestCompatibilityDeclarationJSONRejectionsAndEquivalence(t *testing.T) {
	base, err := json.Marshal(mustCompatibilityDeclaration(t))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct{ name, key, value string }{
		{"absent applicable artifact types", "applicable_artifact_types", ""},
		{"null applicable artifact types", "applicable_artifact_types", "null"},
		{"empty applicable artifact types", "applicable_artifact_types", "[]"},
		{"absent parameter contract", "parameter_contract", ""},
		{"null parameter contract", "parameter_contract", "null"},
		{"absent product contract", "product_contract", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			if tt.value == "" {
				delete(mCopy, tt.key)
			} else {
				mCopy[tt.key] = json.RawMessage(tt.value)
			}
			data, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var d CompatibilityDeclaration
			if err := json.Unmarshal(data, &d); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}

	// applicable_revisions is optional: absent, null, and [] are equivalent.
	for _, tt := range []struct{ name, value string }{
		{"absent", ""},
		{"null", "null"},
		{"empty array", "[]"},
	} {
		t.Run("applicable_revisions "+tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			if tt.value == "" {
				delete(mCopy, "applicable_revisions")
			} else {
				mCopy["applicable_revisions"] = json.RawMessage(tt.value)
			}
			data, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var d CompatibilityDeclaration
			if err := json.Unmarshal(data, &d); err != nil {
				t.Errorf("%s applicable_revisions rejected: %v", tt.name, err)
			}
			if len(d.ApplicableRevisions()) != 0 {
				t.Error("expected no applicable revisions")
			}
		})
	}
}

// TestCompatibilityDeclarationUnmarshalRevalidatesModifiers confirms decode
// converges on the same validation the With* methods apply: an optional value
// that only the modifier could reject is still rejected when it arrives by JSON.
func TestCompatibilityDeclarationUnmarshalRevalidatesModifiers(t *testing.T) {
	base, err := json.Marshal(mustCompatibilityDeclaration(t))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}
	refJSON, err := json.Marshal(mustTemplateRevisionRef(t, "TPL-2", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		key   string
		value json.RawMessage
	}{
		{"duplicate applicable revisions", "applicable_revisions", json.RawMessage(`[` + string(refJSON) + `,` + string(refJSON) + `]`)},
		{"zero applicable revision", "applicable_revisions", json.RawMessage(`[{"artifact_id":"","revision_id":""}]`)},
		{"whitespace-only migration requirements", "migration_requirements", json.RawMessage(`"   "`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			mCopy[tt.key] = tt.value
			data, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var d CompatibilityDeclaration
			if err := json.Unmarshal(data, &d); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}

	t.Run("non-object payload", func(t *testing.T) {
		var d CompatibilityDeclaration
		if err := json.Unmarshal([]byte(`123`), &d); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
			t.Errorf("error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
		}
	})
}

// TestValueTypesRejectNonObjectPayloads confirms every Revision-owned value
// type surfaces its own sentinel when handed syntactically valid JSON of the
// wrong shape, rather than letting a raw decoding error escape unattributed.
func TestValueTypesRejectNonObjectPayloads(t *testing.T) {
	t.Run("ParameterType", func(t *testing.T) {
		var pt ParameterType
		if err := json.Unmarshal([]byte(`123`), &pt); !errors.Is(err, ErrInvalidParameterType) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterType)
		}
	})
	t.Run("Parameter", func(t *testing.T) {
		var p Parameter
		if err := json.Unmarshal([]byte(`"a string"`), &p); !errors.Is(err, ErrInvalidTemplateParameter) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateParameter)
		}
	})
	t.Run("ParameterDefault", func(t *testing.T) {
		var d ParameterDefault
		if err := json.Unmarshal([]byte(`123`), &d); !errors.Is(err, ErrInvalidParameterDefault) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterDefault)
		}
	})
	t.Run("ConstraintTarget", func(t *testing.T) {
		var target ConstraintTarget
		if err := json.Unmarshal([]byte(`123`), &target); !errors.Is(err, ErrInvalidConstraintTarget) {
			t.Errorf("error = %v, want %v", err, ErrInvalidConstraintTarget)
		}
	})
	t.Run("ParameterConstraint", func(t *testing.T) {
		var v ParameterConstraint
		if err := json.Unmarshal([]byte(`123`), &v); !errors.Is(err, ErrInvalidParameterConstraint) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterConstraint)
		}
	})
	t.Run("TemplateContent", func(t *testing.T) {
		var c TemplateContent
		if err := json.Unmarshal([]byte(`123`), &c); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("TemplateRevision", func(t *testing.T) {
		var r TemplateRevision
		if err := json.Unmarshal([]byte(`123`), &r); !errors.Is(err, ErrInvalidTemplate) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplate)
		}
	})
	t.Run("Template", func(t *testing.T) {
		var tmpl Template
		if err := json.Unmarshal([]byte(`123`), &tmpl); !errors.Is(err, ErrInvalidTemplate) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplate)
		}
	})
}

func TestCompatibilityDeclarationUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	d := mustCompatibilityDeclaration(t)
	before, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"applicable_artifact_types":[]}`), &d); err == nil {
		t.Fatal("expected error")
	}
	if err := json.Unmarshal([]byte(`not json`), &d); err == nil {
		t.Fatal("expected error")
	}
	after, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// TestCompatibilityDeclarationStoresNoVerdict is the structural proof that this
// type declares compatibility inputs and never records a compatibility result:
// "Current compatibility is a derived interpretation, computed from the
// applicable compatibility declarations at query time."
func TestCompatibilityDeclarationStoresNoVerdict(t *testing.T) {
	d := mustCompatibilityDeclaration(t)
	forbidden := []string{
		"compatible", "compatibility", "incompatible", "current", "effective",
		"status", "state", "verdict", "result", "outcome", "conformant",
	}
	assertNoWireKeys(t, "CompatibilityDeclaration", d, forbidden)
}

// --- K3-02 / K3-03 regression: compatibility scoping --------------------------

// TestCompatibilityDeclarationOwningRevisionIsImplicitScope is the K3-02
// regression guard. The owning Revision is a compatibility declaration's
// implicit scope: an empty ApplicableRevisions() means "this Revision alone",
// not "unscoped". ApplicableRevisions() names only *additional* Revisions.
func TestCompatibilityDeclarationOwningRevisionIsImplicitScope(t *testing.T) {
	d := mustCompatibilityDeclaration(t)
	if len(d.ApplicableRevisions()) != 0 {
		t.Fatal("a fresh declaration should name no additional revisions")
	}
	if d.AppliesBeyondOwningRevision() {
		t.Error("AppliesBeyondOwningRevision() = true for a declaration scoped to its owner alone")
	}

	// Such a declaration is fully valid and usable as TemplateContent's
	// mandatory compatibility declaration -- the implicit scope is complete.
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters", d, NewUnrestrictedTemplateApplicability(),
		mustProvenance(t), nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("a declaration scoped to its owning Revision alone must be valid: %v", err)
	}
	if c.Compatibility().AppliesBeyondOwningRevision() {
		t.Error("round trip through TemplateContent changed the scoping answer")
	}

	// Naming additional Revisions flips the answer and is preserved.
	extended, err := d.WithApplicableRevisions([]core.TemplateArtifactRevisionRef{
		mustTemplateRevisionRef(t, "TPL-2", "REV-1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !extended.AppliesBeyondOwningRevision() {
		t.Error("AppliesBeyondOwningRevision() = false after naming another Revision")
	}
	if len(extended.ApplicableRevisions()) != 1 {
		t.Error("the additional Revision was not recorded")
	}
	if d.AppliesBeyondOwningRevision() {
		t.Error("WithApplicableRevisions mutated its receiver")
	}
}

// TestCompatibilityDeclarationConstructionOrderForbidsMandatoryRevisions
// documents why K3-02 was resolved by the implicit-scope reading rather than by
// making applicable revisions mandatory: a CompatibilityDeclaration is built
// before the TemplateContent that holds it, which is itself built before the
// TemplateRevision that gives it an identity. At declaration time the owning
// Revision reference does not exist, so it cannot be required.
func TestCompatibilityDeclarationConstructionOrderForbidsMandatoryRevisions(t *testing.T) {
	// Step 1: the declaration exists with no Revision in sight.
	d := mustCompatibilityDeclaration(t)
	// Step 2: content is built from it -- still no Revision identity anywhere.
	content, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters", d, NewUnrestrictedTemplateApplicability(),
		mustProvenance(t), nil, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	// Step 3: only now does the Revision -- and its reference -- come into being.
	revision, err := NewTemplateRevision(mustTemplate(t, "TPL-1"), mustArtifactRevision(t, "TPL-1", "REV-1"), content)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := revision.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsZero() {
		t.Fatal("the owning Revision reference should now exist")
	}
	// The declaration reached through the Revision is the same one, still
	// naming no additional Revisions -- and now demonstrably scoped to `ref`.
	if revision.Content().Compatibility().AppliesBeyondOwningRevision() {
		t.Error("the declaration should still be scoped to its owner alone")
	}
}

// TestCompatibilityDeclarationConstraintsLiveOnTheOwningRevision is the K3-03
// regression guard. "Applicable constraints" are exactly the owning Revision's
// own ParameterConstraints, reachable through the constraint namespace; the
// declaration deliberately stores no copy and no subset, so there is no
// duplicated state to drift.
func TestCompatibilityDeclarationConstraintsLiveOnTheOwningRevision(t *testing.T) {
	content, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters", mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(), mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true)},
		nil,
		[]ParameterConstraint{mustParameterConstraint(t, "name-nonempty", "name")},
	)
	if err != nil {
		t.Fatal(err)
	}

	// The constraints scoping the declaration are the owning content's own,
	// resolvable by key through the constraint namespace.
	if got := content.Constraints(); len(got) != 1 {
		t.Fatalf("Constraints() = %d, want 1", len(got))
	}
	if _, ok := content.Constraint(mustLocalKey(t, "name-nonempty")); !ok {
		t.Error("the applicable constraint must resolve through the owning Revision's namespace")
	}

	// The declaration itself stores no constraint state at all -- no copy, no
	// subset, nothing that could drift from the owning Revision.
	assertNoMethods(t, "CompatibilityDeclaration", reflect.TypeOf(CompatibilityDeclaration{}), []string{
		"Constraints", "ApplicableConstraints", "ConstraintKeys",
		"WithConstraints", "WithApplicableConstraints",
	})
	assertNoWireKeys(t, "CompatibilityDeclaration", content.Compatibility(), []string{
		"constraints", "applicable_constraints", "constraint_keys",
	})
}
