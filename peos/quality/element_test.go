package quality

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- shared helpers ----------------------------------------------------------

func mustLocalKey(t *testing.T, value string) core.LocalKey {
	t.Helper()
	k, err := core.NewLocalKey(value)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func mustVocabularyValue(t *testing.T, namespace, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustUnit(t *testing.T, value string) Unit {
	t.Helper()
	return NewUnit(mustVocabularyValue(t, "product-x", value))
}

func mustScale(t *testing.T, value string) Scale {
	t.Helper()
	return NewScale(mustVocabularyValue(t, "product-x", value))
}

func mustOperator(t *testing.T, value string) ThresholdOperator {
	t.Helper()
	return NewThresholdOperator(mustVocabularyValue(t, "product-x", value))
}

func mustValidationMethod(t *testing.T, value string) core.ValidationMethod {
	t.Helper()
	return core.NewValidationMethod(mustVocabularyValue(t, "product-x", value))
}

func mustExtension(t *testing.T) core.Extension {
	t.Helper()
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"note":"n"}`))
	if err != nil {
		t.Fatal(err)
	}
	return ext
}

func mustProfileCharacteristic(t *testing.T, key, term string) Characteristic {
	t.Helper()
	c, err := NewProfileCharacteristic(mustLocalKey(t, key), term)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustMeasure(t *testing.T, key, characteristic string) Measure {
	t.Helper()
	m, err := NewMeasure(
		mustLocalKey(t, key),
		mustLocalKey(t, characteristic),
		mustUnit(t, "millisecond"),
		mustScale(t, "ratio"),
		mustValidationMethod(t, "automated-test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func mustThreshold(t *testing.T, key, measure string) Threshold {
	t.Helper()
	v, err := NewThreshold(mustLocalKey(t, key), mustLocalKey(t, measure), mustOperator(t, "lte"), "250")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustTarget(t *testing.T, key, measure string) Target {
	t.Helper()
	v, err := NewTarget(mustLocalKey(t, key), mustLocalKey(t, measure), "120")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustConstraint(t *testing.T, key, statement string) Constraint {
	t.Helper()
	v, err := NewConstraint(mustLocalKey(t, key), statement)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustNormalizationRule(t *testing.T, key, description string) NormalizationRule {
	t.Helper()
	v, err := NewNormalizationRule(mustLocalKey(t, key), description)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustAggregationRule(t *testing.T, key, description string) AggregationRule {
	t.Helper()
	v, err := NewAggregationRule(mustLocalKey(t, key), description)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// --- vocabulary wrappers -----------------------------------------------------

func TestVocabularyWrappersRoundTripAndCompare(t *testing.T) {
	unitA := mustUnit(t, "millisecond")
	unitB := mustUnit(t, "millisecond")
	unitC := mustUnit(t, "second")

	if !unitA.Equal(unitB) {
		t.Error("Unit.Equal() = false for identical values")
	}
	if unitA.Equal(unitC) {
		t.Error("Unit.Equal() = true for different values")
	}
	if unitA.IsZero() {
		t.Error("Unit.IsZero() = true for a constructed unit")
	}
	if (Unit{}).IsZero() != true {
		t.Error("zero Unit.IsZero() = false")
	}
	if unitA.String() != "product-x:millisecond" {
		t.Errorf("Unit.String() = %q", unitA.String())
	}
	if unitA.Value() != mustVocabularyValue(t, "product-x", "millisecond") {
		t.Error("Unit.Value() mismatch")
	}

	scaleA := mustScale(t, "ratio")
	if !scaleA.Equal(mustScale(t, "ratio")) || scaleA.Equal(mustScale(t, "ordinal")) {
		t.Error("Scale.Equal() incorrect")
	}
	opA := mustOperator(t, "lte")
	if !opA.Equal(mustOperator(t, "lte")) || opA.Equal(mustOperator(t, "gte")) {
		t.Error("ThresholdOperator.Equal() incorrect")
	}
	if (Scale{}).IsZero() != true || (ThresholdOperator{}).IsZero() != true {
		t.Error("zero Scale/ThresholdOperator IsZero() = false")
	}
	if scaleA.String() != "product-x:ratio" || opA.String() != "product-x:lte" {
		t.Error("String() mismatch on Scale or ThresholdOperator")
	}
	if scaleA.Value().IsZero() || opA.Value().IsZero() {
		t.Error("Value() returned a zero vocabulary value")
	}

	// Each wrapper round-trips as a bare vocabulary string, adding no
	// envelope of its own.
	for name, pair := range map[string][2]any{
		"Unit":              {unitA, new(Unit)},
		"Scale":             {scaleA, new(Scale)},
		"ThresholdOperator": {opA, new(ThresholdOperator)},
	} {
		data, err := json.Marshal(pair[0])
		if err != nil {
			t.Fatalf("%s: marshal: %v", name, err)
		}
		if !strings.HasPrefix(string(data), `"`) {
			t.Errorf("%s wire form = %s, want a bare string", name, data)
		}
		if err := json.Unmarshal(data, pair[1]); err != nil {
			t.Fatalf("%s: unmarshal: %v", name, err)
		}
	}

	decoded := new(Unit)
	if err := json.Unmarshal([]byte(`"product-x:millisecond"`), decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(unitA) {
		t.Error("Unit did not round-trip through its wire form")
	}
}

func TestVocabularyWrappersAreDistinctTypes(t *testing.T) {
	// Unit, Scale, and ThresholdOperator wrap the same underlying value but
	// are separate Go types with no conversion path, so one can never be
	// passed where another is expected. The following, if uncommented, must
	// fail to compile:
	//   var u Unit = mustScale(t, "ratio")
	//   var s Scale = mustOperator(t, "lte")
	v := mustVocabularyValue(t, "product-x", "same")
	if NewUnit(v).String() != NewScale(v).String() {
		t.Error("wrappers of the same value disagree on String()")
	}
}

// --- Characteristic ----------------------------------------------------------

func TestNewProfileCharacteristic(t *testing.T) {
	c := mustProfileCharacteristic(t, "latency", "  Response latency  ")
	if c.Kind() != "profile" {
		t.Errorf("Kind() = %q, want %q", c.Kind(), "profile")
	}
	term, ok := c.Term()
	if !ok || term != "Response latency" {
		t.Errorf("Term() = (%q, %v), want (%q, true) with surrounding whitespace trimmed", term, ok, "Response latency")
	}
	if _, ok := c.ExternalVocabulary(); ok {
		t.Error("ExternalVocabulary() ok=true for a profile-scoped characteristic")
	}
	if c.IsExternallyScoped() {
		t.Error("IsExternallyScoped() = true for a profile-scoped characteristic")
	}
	if c.Key() != mustLocalKey(t, "latency") {
		t.Error("Key() mismatch")
	}
	if c.IsZero() {
		t.Error("IsZero() = true for a constructed characteristic")
	}
	if _, ok := c.Description(); ok {
		t.Error("Description() ok=true before one is set")
	}
	if !c.Extension().IsZero() {
		t.Error("Extension() non-zero before one is set")
	}
}

func TestNewProfileCharacteristicRejectsInvalidInput(t *testing.T) {
	if _, err := NewProfileCharacteristic(core.LocalKey{}, "term"); !errors.Is(err, ErrInvalidQualityCharacteristic) {
		t.Errorf("zero key error = %v, want %v", err, ErrInvalidQualityCharacteristic)
	}
	for _, term := range []string{"", "   ", "\t\n"} {
		if _, err := NewProfileCharacteristic(mustLocalKey(t, "k"), term); !errors.Is(err, ErrInvalidQualityCharacteristic) {
			t.Errorf("term %q accepted, want rejection", term)
		}
	}
}

func TestNewExternalCharacteristic(t *testing.T) {
	vocab := mustVocabularyValue(t, "iso", "25010-maintainability")
	c, err := NewExternalCharacteristic(mustLocalKey(t, "maintainability"), vocab)
	if err != nil {
		t.Fatal(err)
	}
	if c.Kind() != "external" {
		t.Errorf("Kind() = %q, want %q", c.Kind(), "external")
	}
	if !c.IsExternallyScoped() {
		t.Error("IsExternallyScoped() = false for an externally scoped characteristic")
	}
	got, ok := c.ExternalVocabulary()
	if !ok || got != vocab {
		t.Errorf("ExternalVocabulary() = (%v, %v), want (%v, true)", got, ok, vocab)
	}
	if _, ok := c.Term(); ok {
		t.Error("Term() ok=true for an externally scoped characteristic")
	}
	// The external arm still carries a local key: the external vocabulary
	// supplies the meaning, the key names it within this Revision.
	if c.Key().IsZero() {
		t.Error("an externally scoped characteristic must still carry a profile-local key")
	}
}

func TestNewExternalCharacteristicRejectsInvalidInput(t *testing.T) {
	if _, err := NewExternalCharacteristic(core.LocalKey{}, mustVocabularyValue(t, "iso", "x")); !errors.Is(err, ErrInvalidQualityCharacteristic) {
		t.Errorf("zero key error = %v, want %v", err, ErrInvalidQualityCharacteristic)
	}
	if _, err := NewExternalCharacteristic(mustLocalKey(t, "k"), core.VocabularyValue{}); !errors.Is(err, ErrInvalidQualityCharacteristic) {
		t.Error("zero external vocabulary accepted, want rejection")
	}
}

func TestCharacteristicArmsAreExclusive(t *testing.T) {
	// There is no constructor accepting both a term and an external
	// vocabulary, and none accepting neither: the two constructors are the
	// only construction paths, and each establishes exactly one arm. The JSON
	// path is where a document could attempt either, and both are rejected
	// there -- see TestCharacteristicJSONRejectsMalformedArms.
	profile := mustProfileCharacteristic(t, "k", "term")
	external, err := NewExternalCharacteristic(mustLocalKey(t, "k"), mustVocabularyValue(t, "iso", "x"))
	if err != nil {
		t.Fatal(err)
	}
	if profile.IsExternallyScoped() == external.IsExternallyScoped() {
		t.Error("the two arms are not distinguishable")
	}
	if _, ok := profile.Term(); !ok {
		t.Error("profile arm has no term")
	}
	if _, ok := external.ExternalVocabulary(); !ok {
		t.Error("external arm has no vocabulary")
	}
	// The zero value is neither arm and is invalid.
	var zero Characteristic
	if !zero.IsZero() || zero.Kind() != "" {
		t.Error("zero Characteristic is not recognizable as unscoped")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidQualityCharacteristic) {
		t.Errorf("zero-value marshal error = %v, want %v", err, ErrInvalidQualityCharacteristic)
	}
}

func TestCharacteristicModifiers(t *testing.T) {
	c := mustProfileCharacteristic(t, "latency", "Response latency")

	withDesc, err := c.WithDescription("  how quickly it responds  ")
	if err != nil {
		t.Fatal(err)
	}
	desc, ok := withDesc.Description()
	if !ok || desc != "how quickly it responds" {
		t.Errorf("Description() = (%q, %v)", desc, ok)
	}
	// The original is unchanged: modifiers return copies.
	if _, ok := c.Description(); ok {
		t.Error("WithDescription mutated the receiver")
	}
	if _, ok := withDesc.WithoutDescription().Description(); ok {
		t.Error("WithoutDescription did not clear the description")
	}
	for _, bad := range []string{"", "   "} {
		if _, err := c.WithDescription(bad); !errors.Is(err, ErrInvalidQualityCharacteristic) {
			t.Errorf("WithDescription(%q) accepted, want rejection", bad)
		}
	}

	ext := mustExtension(t)
	withExt := c.WithExtension(ext)
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set the extension")
	}
	if !c.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}
	if !withExt.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear the extension")
	}
}

func TestCharacteristicJSONRoundTrip(t *testing.T) {
	profile, err := mustProfileCharacteristic(t, "latency", "Response latency").WithDescription("d")
	if err != nil {
		t.Fatal(err)
	}
	profile = profile.WithExtension(mustExtension(t))

	external, err := NewExternalCharacteristic(mustLocalKey(t, "maintainability"), mustVocabularyValue(t, "iso", "25010-m"))
	if err != nil {
		t.Fatal(err)
	}

	for name, original := range map[string]Characteristic{"profile": profile, "external": external} {
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("%s: marshal: %v", name, err)
		}
		var decoded Characteristic
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("%s: unmarshal: %v", name, err)
		}
		if decoded.Kind() != original.Kind() || decoded.Key() != original.Key() {
			t.Errorf("%s: kind or key not preserved", name)
		}
		gotTerm, gotTermOK := decoded.Term()
		wantTerm, wantTermOK := original.Term()
		if gotTerm != wantTerm || gotTermOK != wantTermOK {
			t.Errorf("%s: term not preserved", name)
		}
		gotVocab, gotVocabOK := decoded.ExternalVocabulary()
		wantVocab, wantVocabOK := original.ExternalVocabulary()
		if gotVocab != wantVocab || gotVocabOK != wantVocabOK {
			t.Errorf("%s: external vocabulary not preserved", name)
		}
		gotDesc, gotDescOK := decoded.Description()
		wantDesc, wantDescOK := original.Description()
		if gotDesc != wantDesc || gotDescOK != wantDescOK {
			t.Errorf("%s: description not preserved", name)
		}
		// A second marshal must be byte-identical.
		again, err := json.Marshal(decoded)
		if err != nil {
			t.Fatal(err)
		}
		if string(again) != string(data) {
			t.Errorf("%s: round trip byte mismatch:\n got %s\nwant %s", name, again, data)
		}
	}
}

func TestCharacteristicJSONForbiddenKeysAbsent(t *testing.T) {
	c := mustProfileCharacteristic(t, "latency", "Response latency")
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysAbsent(t, data, "id", "ref", "artifact_id", "revision_id", "revision",
		"version", "lifecycle", "state", "status", "relation", "source", "target",
		"provenance", "score", "quality_score")
}

func TestCharacteristicJSONRejectsMalformedArms(t *testing.T) {
	cases := map[string]string{
		"missing kind":         `{"key":"k","term":"t"}`,
		"unknown kind":         `{"kind":"invented","key":"k","term":"t"}`,
		"null document":        `null`,
		"null kind":            `{"kind":null,"key":"k","term":"t"}`,
		"profile with vocab":   `{"kind":"profile","key":"k","term":"t","vocabulary":"iso:x"}`,
		"profile without term": `{"kind":"profile","key":"k"}`,
		"profile null term":    `{"kind":"profile","key":"k","term":null}`,
		"profile empty term":   `{"kind":"profile","key":"k","term":"   "}`,
		"external with term":   `{"kind":"external","key":"k","term":"t","vocabulary":"iso:x"}`,
		"external no vocab":    `{"kind":"external","key":"k"}`,
		"external null vocab":  `{"kind":"external","key":"k","vocabulary":null}`,
		"missing key":          `{"kind":"profile","term":"t"}`,
		"null description":     `{"kind":"profile","key":"k","term":"t","description":null}`,
		"empty description":    `{"kind":"profile","key":"k","term":"t","description":"  "}`,
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			var c Characteristic
			if err := json.Unmarshal([]byte(doc), &c); err == nil {
				t.Fatalf("accepted %s, want rejection", doc)
			} else if !errors.Is(err, ErrInvalidQualityCharacteristic) {
				t.Errorf("error = %v, want it to wrap %v", err, ErrInvalidQualityCharacteristic)
			}
			if !c.IsZero() {
				t.Error("receiver was modified by a failed decode")
			}
		})
	}

	// A null key is rejected too, but through core's own sentinel rather
	// than being re-attributed to this package.
	var c Characteristic
	err := json.Unmarshal([]byte(`{"kind":"profile","key":null,"term":"t"}`), &c)
	if err == nil {
		t.Fatal("null key accepted, want rejection")
	}
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("null key error = %v, want it to wrap core.ErrEmptyIdentity", err)
	}
}

func TestCharacteristicFailedDecodePreservesReceiver(t *testing.T) {
	original := mustProfileCharacteristic(t, "latency", "Response latency")
	c := original
	if err := json.Unmarshal([]byte(`{"kind":"nope"}`), &c); err == nil {
		t.Fatal("expected rejection")
	}
	if c.Kind() != original.Kind() || c.Key() != original.Key() {
		t.Error("a failed decode overwrote a previously valid receiver")
	}
}

// --- Measure ------------------------------------------------------------------

func TestNewMeasure(t *testing.T) {
	m := mustMeasure(t, "latency-p99", "latency")
	if m.Key() != mustLocalKey(t, "latency-p99") {
		t.Error("Key() mismatch")
	}
	if m.Characteristic() != mustLocalKey(t, "latency") {
		t.Error("Characteristic() mismatch")
	}
	if !m.Unit().Equal(mustUnit(t, "millisecond")) {
		t.Error("Unit() mismatch")
	}
	if !m.Scale().Equal(mustScale(t, "ratio")) {
		t.Error("Scale() mismatch")
	}
	if m.Method() != mustValidationMethod(t, "automated-test") {
		t.Error("Method() mismatch")
	}
	if m.IsZero() {
		t.Error("IsZero() = true for a constructed measure")
	}
	if m.RequiredEvidence() != nil {
		t.Error("RequiredEvidence() non-nil before any is set")
	}
	if _, ok := m.UncertaintyHandling(); ok {
		t.Error("UncertaintyHandling() ok=true before one is set")
	}
	if _, ok := m.ValidRange(); ok {
		t.Error("ValidRange() ok=true before one is set")
	}
	if _, ok := m.NormalizationRule(); ok {
		t.Error("NormalizationRule() ok=true before one is set")
	}
}

func TestNewMeasureRejectsZeroMandatoryFields(t *testing.T) {
	key := mustLocalKey(t, "m")
	char := mustLocalKey(t, "c")
	unit := mustUnit(t, "ms")
	scale := mustScale(t, "ratio")
	method := mustValidationMethod(t, "test")

	cases := map[string]func() (Measure, error){
		"zero key":            func() (Measure, error) { return NewMeasure(core.LocalKey{}, char, unit, scale, method) },
		"zero characteristic": func() (Measure, error) { return NewMeasure(key, core.LocalKey{}, unit, scale, method) },
		"zero unit":           func() (Measure, error) { return NewMeasure(key, char, Unit{}, scale, method) },
		"zero scale":          func() (Measure, error) { return NewMeasure(key, char, unit, Scale{}, method) },
		"zero method":         func() (Measure, error) { return NewMeasure(key, char, unit, scale, core.ValidationMethod{}) },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := fn(); !errors.Is(err, ErrInvalidQualityMeasure) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityMeasure)
			}
		})
	}
}

func TestMeasureOptionalFields(t *testing.T) {
	m := mustMeasure(t, "latency-p99", "latency")

	withEvidence, err := m.WithRequiredEvidence([]string{"  trace export  ", "load-test report"})
	if err != nil {
		t.Fatal(err)
	}
	got := withEvidence.RequiredEvidence()
	if len(got) != 2 || got[0] != "trace export" || got[1] != "load-test report" {
		t.Errorf("RequiredEvidence() = %q", got)
	}
	// Defensive copy on the way out.
	got[0] = "mutated"
	if withEvidence.RequiredEvidence()[0] != "trace export" {
		t.Error("RequiredEvidence() returned an aliased slice")
	}
	// Defensive copy on the way in.
	input := []string{"a"}
	withInput, err := m.WithRequiredEvidence(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = "mutated"
	if withInput.RequiredEvidence()[0] != "a" {
		t.Error("WithRequiredEvidence retained the caller's slice")
	}
	// nil clears; there is no WithoutRequiredEvidence.
	if withEvidence.mustWithRequiredEvidence(t, nil).RequiredEvidence() != nil {
		t.Error("WithRequiredEvidence(nil) did not clear")
	}
	if _, err := m.WithRequiredEvidence([]string{"ok", "  "}); !errors.Is(err, ErrInvalidQualityMeasure) {
		t.Error("a whitespace-only evidence entry was accepted")
	}

	withHandling, err := m.WithUncertaintyHandling("  +/- 5ms  ")
	if err != nil {
		t.Fatal(err)
	}
	if h, ok := withHandling.UncertaintyHandling(); !ok || h != "+/- 5ms" {
		t.Errorf("UncertaintyHandling() = (%q, %v)", h, ok)
	}
	if _, ok := withHandling.WithoutUncertaintyHandling().UncertaintyHandling(); ok {
		t.Error("WithoutUncertaintyHandling did not clear")
	}
	if _, err := m.WithUncertaintyHandling("  "); !errors.Is(err, ErrInvalidQualityMeasure) {
		t.Error("empty uncertainty handling accepted")
	}

	withRange, err := m.WithValidRange("0..10000")
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := withRange.ValidRange(); !ok || r != "0..10000" {
		t.Errorf("ValidRange() = (%q, %v)", r, ok)
	}
	if _, ok := withRange.WithoutValidRange().ValidRange(); ok {
		t.Error("WithoutValidRange did not clear")
	}
	if _, err := m.WithValidRange(""); !errors.Is(err, ErrInvalidQualityMeasure) {
		t.Error("empty valid range accepted")
	}

	withRule, err := m.WithNormalizationRule(mustLocalKey(t, "norm-1"))
	if err != nil {
		t.Fatal(err)
	}
	if k, ok := withRule.NormalizationRule(); !ok || k != mustLocalKey(t, "norm-1") {
		t.Errorf("NormalizationRule() = (%v, %v)", k, ok)
	}
	if _, ok := withRule.WithoutNormalizationRule().NormalizationRule(); ok {
		t.Error("WithoutNormalizationRule did not clear")
	}
	if _, err := m.WithNormalizationRule(core.LocalKey{}); !errors.Is(err, ErrInvalidQualityMeasure) {
		t.Error("zero normalization rule key accepted")
	}

	withExt := m.WithExtension(mustExtension(t))
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set the extension")
	}
	if !withExt.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear")
	}

	// None of the modifiers mutated the original.
	if len(m.RequiredEvidence()) != 0 {
		t.Error("the original measure was mutated")
	}
}

// mustWithRequiredEvidence is a small test-local helper keeping the chained
// assertion above readable.
func (m Measure) mustWithRequiredEvidence(t *testing.T, descriptions []string) Measure {
	t.Helper()
	result, err := m.WithRequiredEvidence(descriptions)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestMeasureJSONRoundTrip(t *testing.T) {
	m := mustMeasure(t, "latency-p99", "latency")
	m, err := m.WithRequiredEvidence([]string{"trace export"})
	if err != nil {
		t.Fatal(err)
	}
	if m, err = m.WithUncertaintyHandling("+/- 5ms"); err != nil {
		t.Fatal(err)
	}
	if m, err = m.WithValidRange("0..10000"); err != nil {
		t.Fatal(err)
	}
	if m, err = m.WithNormalizationRule(mustLocalKey(t, "norm-1")); err != nil {
		t.Fatal(err)
	}
	m = m.WithExtension(mustExtension(t))

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Measure
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
	}
	assertKeysAbsent(t, data, "id", "ref", "artifact_id", "revision_id", "revision",
		"version", "lifecycle", "state", "status", "relation", "source", "target",
		"provenance", "score")
}

func TestMeasureJSONMissingAndNull(t *testing.T) {
	valid := `{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}`
	var ok Measure
	if err := json.Unmarshal([]byte(valid), &ok); err != nil {
		t.Fatalf("valid document rejected: %v", err)
	}

	rejected := map[string]string{
		"missing key":            `{"characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}`,
		"missing characteristic": `{"key":"m","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}`,
		"missing unit":           `{"key":"m","characteristic":"c","scale":"product-x:ratio","method":"product-x:test"}`,
		"missing scale":          `{"key":"m","characteristic":"c","unit":"product-x:ms","method":"product-x:test"}`,
		"missing method":         `{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio"}`,
		"null unit":              `{"key":"m","characteristic":"c","unit":null,"scale":"product-x:ratio","method":"product-x:test"}`,
		"null scale":             `{"key":"m","characteristic":"c","unit":"product-x:ms","scale":null,"method":"product-x:test"}`,
		"null method":            `{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":null}`,
		"null uncertainty":       valid[:len(valid)-1] + `,"uncertainty_handling":null}`,
		"null valid range":       valid[:len(valid)-1] + `,"valid_range":null}`,
		"null normalization":     valid[:len(valid)-1] + `,"normalization_rule":null}`,
		"empty uncertainty":      valid[:len(valid)-1] + `,"uncertainty_handling":"  "}`,
		"empty evidence entry":   valid[:len(valid)-1] + `,"required_evidence":["ok","  "]}`,
	}
	for name, doc := range rejected {
		t.Run(name, func(t *testing.T) {
			var m Measure
			if err := json.Unmarshal([]byte(doc), &m); err == nil {
				t.Fatalf("accepted %s, want rejection", doc)
			} else if !errors.Is(err, ErrInvalidQualityMeasure) {
				t.Errorf("error = %v, want it to wrap %v", err, ErrInvalidQualityMeasure)
			}
			if !m.IsZero() {
				t.Error("receiver was modified by a failed decode")
			}
		})
	}

	// A null key is rejected through core's own sentinel.
	var m Measure
	err := json.Unmarshal([]byte(`{"key":null,"characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}`), &m)
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("null key error = %v, want it to wrap core.ErrEmptyIdentity", err)
	}

	// required_evidence: absent, null, and [] are all equivalent to "none".
	for _, doc := range []string{
		valid,
		valid[:len(valid)-1] + `,"required_evidence":null}`,
		valid[:len(valid)-1] + `,"required_evidence":[]}`,
	} {
		var decoded Measure
		if err := json.Unmarshal([]byte(doc), &decoded); err != nil {
			t.Fatalf("%s rejected: %v", doc, err)
		}
		if decoded.RequiredEvidence() != nil {
			t.Errorf("%s produced required evidence, want none", doc)
		}
	}

	// A zero-value marshal fails with the owning sentinel.
	if _, err := json.Marshal(Measure{}); !errors.Is(err, ErrInvalidQualityMeasure) {
		t.Errorf("zero-value marshal error = %v, want %v", err, ErrInvalidQualityMeasure)
	}
}

// --- Threshold and Target -----------------------------------------------------

func TestNewThreshold(t *testing.T) {
	th := mustThreshold(t, "latency-max", "latency-p99")
	if th.Key() != mustLocalKey(t, "latency-max") {
		t.Error("Key() mismatch")
	}
	if th.Measure() != mustLocalKey(t, "latency-p99") {
		t.Error("Measure() mismatch")
	}
	if !th.Operator().Equal(mustOperator(t, "lte")) {
		t.Error("Operator() mismatch")
	}
	if th.Value() != "250" {
		t.Errorf("Value() = %q", th.Value())
	}
	if th.IsZero() {
		t.Error("IsZero() = true for a constructed threshold")
	}
}

func TestNewThresholdRejectsInvalidInput(t *testing.T) {
	key := mustLocalKey(t, "t")
	measure := mustLocalKey(t, "m")
	op := mustOperator(t, "lte")

	cases := map[string]func() (Threshold, error){
		"zero key":      func() (Threshold, error) { return NewThreshold(core.LocalKey{}, measure, op, "1") },
		"zero measure":  func() (Threshold, error) { return NewThreshold(key, core.LocalKey{}, op, "1") },
		"zero operator": func() (Threshold, error) { return NewThreshold(key, measure, ThresholdOperator{}, "1") },
		"empty value":   func() (Threshold, error) { return NewThreshold(key, measure, op, "") },
		"blank value":   func() (Threshold, error) { return NewThreshold(key, measure, op, "   ") },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := fn(); !errors.Is(err, ErrInvalidQualityThreshold) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityThreshold)
			}
		})
	}
}

func TestNewTarget(t *testing.T) {
	tg := mustTarget(t, "latency-goal", "latency-p99")
	if tg.Key() != mustLocalKey(t, "latency-goal") {
		t.Error("Key() mismatch")
	}
	if tg.Measure() != mustLocalKey(t, "latency-p99") {
		t.Error("Measure() mismatch")
	}
	if tg.Value() != "120" {
		t.Errorf("Value() = %q", tg.Value())
	}
}

func TestNewTargetRejectsInvalidInput(t *testing.T) {
	key := mustLocalKey(t, "t")
	measure := mustLocalKey(t, "m")
	cases := map[string]func() (Target, error){
		"zero key":     func() (Target, error) { return NewTarget(core.LocalKey{}, measure, "1") },
		"zero measure": func() (Target, error) { return NewTarget(key, core.LocalKey{}, "1") },
		"empty value":  func() (Target, error) { return NewTarget(key, measure, "  ") },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := fn(); !errors.Is(err, ErrInvalidQualityTarget) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityTarget)
			}
		})
	}
}

// TestThresholdAndTargetAreNotConflated locks PEOS-007's "The two SHALL NOT
// be conflated" as a structural property rather than a documented intention.
func TestThresholdAndTargetAreNotConflated(t *testing.T) {
	th := mustThreshold(t, "k", "m")
	tg := mustTarget(t, "k", "m")

	// Separate Go types with separate sentinels. The following, if
	// uncommented, must fail to compile:
	//   var x Threshold = tg
	//   var y Target = th
	if _, err := NewThreshold(core.LocalKey{}, mustLocalKey(t, "m"), mustOperator(t, "lte"), "1"); errors.Is(err, ErrInvalidQualityTarget) {
		t.Error("a Threshold failure reported the Target sentinel")
	}
	if _, err := NewTarget(core.LocalKey{}, mustLocalKey(t, "m"), "1"); errors.Is(err, ErrInvalidQualityThreshold) {
		t.Error("a Target failure reported the Threshold sentinel")
	}

	// A Target has no operator: a comparison is what turns a value into a
	// boundary, and a Target is not a boundary. Its wire form proves it.
	targetData, err := json.Marshal(tg)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysAbsent(t, targetData, "operator")
	thresholdData, err := json.Marshal(th)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysPresent(t, thresholdData, "operator")

	// Neither stores a classification outcome or any derived quality state.
	for name, data := range map[string][]byte{"threshold": thresholdData, "target": targetData} {
		t.Run(name, func(t *testing.T) {
			assertKeysAbsent(t, data, "satisfied", "met", "achieved", "outcome",
				"verdict", "status", "state", "score", "quality_score", "pass",
				"fail", "classification", "current", "latest", "effective",
				"aggregate", "id", "ref", "revision", "version", "lifecycle",
				"relation", "provenance")
		})
	}

	// A Threshold's wire document is not decodable as a Target with its
	// operator silently dropped -- and vice versa, a Target document lacks
	// the operator a Threshold requires.
	var asTarget Target
	if err := json.Unmarshal(thresholdData, &asTarget); err != nil {
		t.Fatalf("a threshold document failed to decode as a target: %v", err)
	}
	if asTarget.Value() != th.Value() {
		t.Error("value not preserved")
	}
	// The operator is lost, which is exactly why the two are separate
	// criterion kinds in core: nothing about the payload distinguishes them,
	// only the discriminator does.
	var asThreshold Threshold
	if err := json.Unmarshal(targetData, &asThreshold); !errors.Is(err, ErrInvalidQualityThreshold) {
		t.Errorf("a target document decoded as a threshold: %v", err)
	}
}

func TestThresholdAndTargetModifiersAndJSON(t *testing.T) {
	th, err := mustThreshold(t, "k", "m").WithDescription("  boundary  ")
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := th.Description(); !ok || d != "boundary" {
		t.Errorf("Description() = (%q, %v)", d, ok)
	}
	if _, ok := th.WithoutDescription().Description(); ok {
		t.Error("WithoutDescription did not clear")
	}
	if _, err := th.WithDescription(" "); !errors.Is(err, ErrInvalidQualityThreshold) {
		t.Error("blank description accepted on Threshold")
	}
	th = th.WithExtension(mustExtension(t))
	if th.Extension().IsZero() {
		t.Error("WithExtension did not set on Threshold")
	}
	if !th.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear on Threshold")
	}

	tg, err := mustTarget(t, "k", "m").WithDescription("intent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tg.WithDescription(""); !errors.Is(err, ErrInvalidQualityTarget) {
		t.Error("blank description accepted on Target")
	}
	if _, ok := tg.WithoutDescription().Description(); ok {
		t.Error("WithoutDescription did not clear on Target")
	}
	tg = tg.WithExtension(mustExtension(t))
	if !tg.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear on Target")
	}

	for name, pair := range map[string][2]any{
		"threshold": {th, new(Threshold)},
		"target":    {tg, new(Target)},
	} {
		data, err := json.Marshal(pair[0])
		if err != nil {
			t.Fatalf("%s: marshal: %v", name, err)
		}
		if err := json.Unmarshal(data, pair[1]); err != nil {
			t.Fatalf("%s: unmarshal: %v", name, err)
		}
		again, err := json.Marshal(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if string(again) != string(data) {
			t.Errorf("%s: round trip byte mismatch:\n got %s\nwant %s", name, again, data)
		}
	}

	// Zero-value marshals fail with their own sentinels.
	if _, err := json.Marshal(Threshold{}); !errors.Is(err, ErrInvalidQualityThreshold) {
		t.Error("zero Threshold marshal did not fail with its sentinel")
	}
	if _, err := json.Marshal(Target{}); !errors.Is(err, ErrInvalidQualityTarget) {
		t.Error("zero Target marshal did not fail with its sentinel")
	}

	// Null and missing handling.
	for name, doc := range map[string]string{
		"threshold null description": `{"key":"k","measure":"m","operator":"product-x:lte","value":"1","description":null}`,
		"threshold null operator":    `{"key":"k","measure":"m","operator":null,"value":"1"}`,
		"threshold missing value":    `{"key":"k","measure":"m","operator":"product-x:lte"}`,
		"threshold null value":       `{"key":"k","measure":"m","operator":"product-x:lte","value":null}`,
	} {
		t.Run(name, func(t *testing.T) {
			var v Threshold
			if err := json.Unmarshal([]byte(doc), &v); !errors.Is(err, ErrInvalidQualityThreshold) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityThreshold)
			}
			if !v.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
	}
	for name, doc := range map[string]string{
		"target null description": `{"key":"k","measure":"m","value":"1","description":null}`,
		"target missing value":    `{"key":"k","measure":"m"}`,
		"target null value":       `{"key":"k","measure":"m","value":null}`,
	} {
		t.Run(name, func(t *testing.T) {
			var v Target
			if err := json.Unmarshal([]byte(doc), &v); !errors.Is(err, ErrInvalidQualityTarget) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityTarget)
			}
			if !v.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
	}
}

// --- Constraint ---------------------------------------------------------------

func TestNewConstraint(t *testing.T) {
	c := mustConstraint(t, "no-plaintext-secrets", "  Secrets must not be logged.  ")
	if c.Key() != mustLocalKey(t, "no-plaintext-secrets") {
		t.Error("Key() mismatch")
	}
	if c.Statement() != "Secrets must not be logged." {
		t.Errorf("Statement() = %q, want the trimmed value", c.Statement())
	}
	if c.IsZero() {
		t.Error("IsZero() = true for a constructed constraint")
	}
}

func TestNewConstraintRejectsInvalidInput(t *testing.T) {
	if _, err := NewConstraint(core.LocalKey{}, "s"); !errors.Is(err, ErrInvalidQualityConstraint) {
		t.Errorf("zero key error = %v, want %v", err, ErrInvalidQualityConstraint)
	}
	for _, s := range []string{"", "  ", "\n"} {
		if _, err := NewConstraint(mustLocalKey(t, "k"), s); !errors.Is(err, ErrInvalidQualityConstraint) {
			t.Errorf("statement %q accepted, want rejection", s)
		}
	}
}

// TestConstraintIsNotARequirement locks PEOS-007's "Every Quality Constraint
// SHALL NOT be silently treated as a Requirement" structurally: the type has
// no Requirement identity, no lifecycle, no authority, no applicability, no
// allocation, no relation, and no conversion in either direction. The absence
// of a peos/requirement import (asserted by doc_test.go) is what makes such a
// conversion inexpressible rather than merely absent.
func TestConstraintIsNotARequirement(t *testing.T) {
	c, err := mustConstraint(t, "k", "statement").WithDescription("d")
	if err != nil {
		t.Fatal(err)
	}
	c = c.WithExtension(mustExtension(t))
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysAbsent(t, data,
		"requirement", "requirement_id", "requirement_revision",
		"id", "ref", "artifact_id", "revision_id", "revision", "version",
		"lifecycle", "state", "status", "authority", "applicability",
		"allocation", "relation", "source", "target", "provenance",
		"waiver", "governance", "score")
	assertKeysPresent(t, data, "key", "statement")

	var decoded Constraint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
	}
}

func TestConstraintModifiersAndJSON(t *testing.T) {
	c := mustConstraint(t, "k", "s")
	withDesc, err := c.WithDescription("  why  ")
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := withDesc.Description(); !ok || d != "why" {
		t.Errorf("Description() = (%q, %v)", d, ok)
	}
	if _, ok := withDesc.WithoutDescription().Description(); ok {
		t.Error("WithoutDescription did not clear")
	}
	if _, err := c.WithDescription("  "); !errors.Is(err, ErrInvalidQualityConstraint) {
		t.Error("blank description accepted")
	}
	if !c.WithExtension(mustExtension(t)).WithoutExtension().Extension().IsZero() {
		t.Error("extension modifiers do not round-trip")
	}
	if _, ok := c.Description(); ok {
		t.Error("the original constraint was mutated")
	}

	if _, err := json.Marshal(Constraint{}); !errors.Is(err, ErrInvalidQualityConstraint) {
		t.Error("zero-value marshal did not fail with the owning sentinel")
	}
	for name, doc := range map[string]string{
		"missing statement": `{"key":"k"}`,
		"null statement":    `{"key":"k","statement":null}`,
		"blank statement":   `{"key":"k","statement":"  "}`,
		"null description":  `{"key":"k","statement":"s","description":null}`,
		"blank description": `{"key":"k","statement":"s","description":" "}`,
		"missing key":       `{"statement":"s"}`,
	} {
		t.Run(name, func(t *testing.T) {
			var v Constraint
			if err := json.Unmarshal([]byte(doc), &v); !errors.Is(err, ErrInvalidQualityConstraint) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityConstraint)
			}
			if !v.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
	}
}

// --- NormalizationRule and AggregationRule ------------------------------------

func TestQualityRules(t *testing.T) {
	norm := mustNormalizationRule(t, "norm-1", "  divide by the baseline  ")
	if norm.Key() != mustLocalKey(t, "norm-1") {
		t.Error("NormalizationRule.Key() mismatch")
	}
	if norm.Description() != "divide by the baseline" {
		t.Errorf("NormalizationRule.Description() = %q, want the trimmed value", norm.Description())
	}
	if norm.IsZero() {
		t.Error("NormalizationRule.IsZero() = true for a constructed rule")
	}

	agg := mustAggregationRule(t, "agg-1", "weighted mean of applicable records")
	if agg.Key() != mustLocalKey(t, "agg-1") {
		t.Error("AggregationRule.Key() mismatch")
	}
	if agg.Description() != "weighted mean of applicable records" {
		t.Error("AggregationRule.Description() mismatch")
	}

	// Description is mandatory on both, so there is no WithDescription or
	// WithoutDescription -- the description *is* the rule.
	for name, err := range map[string]error{
		"normalization zero key":   firstErr(NewNormalizationRule(core.LocalKey{}, "d")),
		"normalization blank desc": firstErr(NewNormalizationRule(mustLocalKey(t, "k"), "  ")),
		"normalization empty desc": firstErr(NewNormalizationRule(mustLocalKey(t, "k"), "")),
		"aggregation zero key":     firstErr(NewAggregationRule(core.LocalKey{}, "d")),
		"aggregation blank desc":   firstErr(NewAggregationRule(mustLocalKey(t, "k"), " \t ")),
	} {
		if !errors.Is(err, ErrInvalidQualityRule) {
			t.Errorf("%s: error = %v, want %v", name, err, ErrInvalidQualityRule)
		}
	}

	// Extension modifiers.
	if !norm.WithExtension(mustExtension(t)).WithoutExtension().Extension().IsZero() {
		t.Error("NormalizationRule extension modifiers do not round-trip")
	}
	if !agg.WithExtension(mustExtension(t)).WithoutExtension().Extension().IsZero() {
		t.Error("AggregationRule extension modifiers do not round-trip")
	}
	if !norm.Extension().IsZero() || !agg.Extension().IsZero() {
		t.Error("a rule was mutated by WithExtension")
	}
}

// TestQualityRulesStoreNoDerivedState locks PEOS-007's "An Aggregation Rule
// produces a derived view. It does not itself produce a stored, mutable field
// on any Subject." Neither rule type has any field beyond its key,
// description, and extension, and neither wire form carries a computed value.
func TestQualityRulesStoreNoDerivedState(t *testing.T) {
	norm := mustNormalizationRule(t, "n", "d").WithExtension(mustExtension(t))
	agg := mustAggregationRule(t, "a", "d").WithExtension(mustExtension(t))

	for name, pair := range map[string][2]any{
		"normalization": {norm, new(NormalizationRule)},
		"aggregation":   {agg, new(AggregationRule)},
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(pair[0])
			if err != nil {
				t.Fatal(err)
			}
			assertKeysPresent(t, data, "key", "description")
			assertKeysAbsent(t, data,
				"formula", "expression", "weight", "weights", "dsl",
				"score", "quality_score", "aggregate", "value", "result",
				"current", "latest", "effective", "records", "claims",
				"id", "ref", "revision", "version", "lifecycle", "state",
				"status", "relation", "provenance")
			if err := json.Unmarshal(data, pair[1]); err != nil {
				t.Fatal(err)
			}
			again, err := json.Marshal(pair[1])
			if err != nil {
				t.Fatal(err)
			}
			if string(again) != string(data) {
				t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
			}
		})
	}
}

func TestQualityRulesJSONRejection(t *testing.T) {
	if _, err := json.Marshal(NormalizationRule{}); !errors.Is(err, ErrInvalidQualityRule) {
		t.Error("zero NormalizationRule marshal did not fail with the owning sentinel")
	}
	if _, err := json.Marshal(AggregationRule{}); !errors.Is(err, ErrInvalidQualityRule) {
		t.Error("zero AggregationRule marshal did not fail with the owning sentinel")
	}

	for name, doc := range map[string]string{
		"missing description": `{"key":"k"}`,
		"null description":    `{"key":"k","description":null}`,
		"blank description":   `{"key":"k","description":"  "}`,
		"missing key":         `{"description":"d"}`,
	} {
		t.Run("normalization "+name, func(t *testing.T) {
			var v NormalizationRule
			if err := json.Unmarshal([]byte(doc), &v); !errors.Is(err, ErrInvalidQualityRule) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityRule)
			}
			if !v.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
		t.Run("aggregation "+name, func(t *testing.T) {
			var v AggregationRule
			if err := json.Unmarshal([]byte(doc), &v); !errors.Is(err, ErrInvalidQualityRule) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityRule)
			}
			if !v.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
	}

	// A null key surfaces core's own sentinel on both.
	var norm NormalizationRule
	if err := json.Unmarshal([]byte(`{"key":null,"description":"d"}`), &norm); !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("null key error = %v, want it to wrap core.ErrEmptyIdentity", err)
	}
}

// TestQualityRuleTypesAreDistinct records that the two rule types share a
// sentinel and a shape but never a Go type: one can never be supplied where
// the other is expected. The following, if uncommented, must fail to compile:
//
//	var n NormalizationRule = mustAggregationRule(t, "a", "d")
func TestQualityRuleTypesAreDistinct(t *testing.T) {
	norm := mustNormalizationRule(t, "same-key", "same description")
	agg := mustAggregationRule(t, "same-key", "same description")

	// Their wire forms are identical, which is exactly why the type-level
	// separation, not the encoding, is what keeps them apart.
	normData, err := json.Marshal(norm)
	if err != nil {
		t.Fatal(err)
	}
	aggData, err := json.Marshal(agg)
	if err != nil {
		t.Fatal(err)
	}
	if string(normData) != string(aggData) {
		t.Errorf("wire forms differ unexpectedly: %s vs %s", normData, aggData)
	}
}

// --- test-local assertion helpers --------------------------------------------

// assertKeysAbsent decodes data as a JSON object and fails if any of keys is
// present at the top level.
func assertKeysAbsent(t *testing.T, data []byte, keys ...string) {
	t.Helper()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("decode as object: %v (data: %s)", err, data)
	}
	for _, key := range keys {
		if _, ok := obj[key]; ok {
			t.Errorf("forbidden key %q present in %s", key, data)
		}
	}
}

// assertKeysPresent decodes data as a JSON object and fails if any of keys is
// missing at the top level.
func assertKeysPresent(t *testing.T, data []byte, keys ...string) {
	t.Helper()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("decode as object: %v (data: %s)", err, data)
	}
	for _, key := range keys {
		if _, ok := obj[key]; !ok {
			t.Errorf("required key %q missing from %s", key, data)
		}
	}
}

// firstErr discards a constructor's value and returns only its error, so a
// table of rejection cases stays readable.
func firstErr[T any](_ T, err error) error { return err }
