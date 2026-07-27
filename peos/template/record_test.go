package template

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- helpers -----------------------------------------------------------------

func mustApplicationRecordID(t *testing.T, value string) core.TemplateApplicationRecordID {
	t.Helper()
	id, err := core.NewTemplateApplicationRecordID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustActor(t *testing.T) core.ActorRef {
	t.Helper()
	a, err := core.NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustTimestamp(t *testing.T) core.Timestamp {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func mustEnvironment(t *testing.T, value string) core.VocabularyValue {
	t.Helper()
	return mustVocabularyValue(t, "product", value)
}

func mustGeneratedArtifactRef(t *testing.T, id string) core.GeneratedArtifactRef {
	t.Helper()
	ref, err := core.NewGeneratedArtifactRef(mustArtifactID(t, id))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustGeneratedRevisionRef(t *testing.T, artifactID, revisionID string) core.GeneratedArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewGeneratedArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustGeneratedOutput(t *testing.T, artifactID, revisionID string) GeneratedOutput {
	t.Helper()
	o, err := NewGeneratedOutput(mustGeneratedArtifactRef(t, artifactID), mustGeneratedRevisionRef(t, artifactID, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustResolvedValue(t *testing.T, parameter, value string, source ValueSource) ResolvedValue {
	t.Helper()
	v, err := NewResolvedValue(mustLocalKey(t, parameter), value, source)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// mustApplicationRecord builds a valid indeterminate-outcome record, the
// simplest outcome that constrains generated outputs in neither direction.
func mustApplicationRecord(t *testing.T, id string) ApplicationRecord {
	t.Helper()
	r, err := NewApplicationRecord(
		mustApplicationRecordID(t, id),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t),
		mustTimestamp(t),
		mustEnvironment(t, "ci"),
		mustProvenance(t),
		ApplicationOutcomeIndeterminate,
		[]ResolvedValue{mustResolvedValue(t, "name", "acme", ValueSourceExplicitInput)},
		nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// --- ApplicationOutcome ------------------------------------------------------

// TestApplicationOutcomeVocabulary confirms the five values PEOS-009 names at
// minimum are predeclared, distinct, and namespaced under PEOS.
func TestApplicationOutcomeVocabulary(t *testing.T) {
	want := map[string]ApplicationOutcome{
		"peos:succeeded":           ApplicationOutcomeSucceeded,
		"peos:failed":              ApplicationOutcomeFailed,
		"peos:partially-succeeded": ApplicationOutcomePartiallySucceeded,
		"peos:interrupted":         ApplicationOutcomeInterrupted,
		"peos:indeterminate":       ApplicationOutcomeIndeterminate,
	}
	seen := make(map[string]bool, len(want))
	for s, outcome := range want {
		if outcome.IsZero() {
			t.Errorf("%s is zero", s)
		}
		if outcome.String() != s {
			t.Errorf("String() = %q, want %q", outcome.String(), s)
		}
		if seen[outcome.String()] {
			t.Errorf("%s is not distinct from another predeclared outcome", s)
		}
		seen[outcome.String()] = true
	}
}

// TestApplicationOutcomeIsNotExecutionOutcome documents why this vocabulary is
// template-local: PEOS-006's core.ExecutionOutcome has "completed" rather than
// "succeeded" and no "partially succeeded" at all, so it cannot express what
// PEOS-009 requires.
func TestApplicationOutcomeIsNotExecutionOutcome(t *testing.T) {
	if ApplicationOutcomeSucceeded.Value() == core.ExecutionOutcomeCompleted.Value() {
		t.Error("ApplicationOutcomeSucceeded must not be core.ExecutionOutcomeCompleted; they are different vocabularies")
	}
	for _, v := range []core.VocabularyValue{
		core.ExecutionOutcomeCompleted.Value(),
		core.ExecutionOutcomeFailed.Value(),
		core.ExecutionOutcomeInterrupted.Value(),
		core.ExecutionOutcomeIndeterminate.Value(),
	} {
		if v.String() == ApplicationOutcomePartiallySucceeded.String() {
			t.Error("core.ExecutionOutcome unexpectedly carries a partially-succeeded value")
		}
	}
}

func TestApplicationOutcomeExtensible(t *testing.T) {
	custom := NewApplicationOutcome(mustVocabularyValue(t, "product", "queued"))
	if custom.IsZero() {
		t.Error("a Product-declared outcome should be valid; the vocabulary is extensible")
	}
	if custom.Equal(ApplicationOutcomeSucceeded) {
		t.Error("distinct vocabulary values reported Equal")
	}

	data, err := json.Marshal(custom)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"product:queued"` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded ApplicationOutcome
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(custom) {
		t.Error("round trip mismatch")
	}
	if err := json.Unmarshal([]byte(`"no-namespace"`), &decoded); !errors.Is(err, core.ErrInvalidVocabularyValue) {
		t.Errorf("malformed vocabulary value: error = %v, want %v", err, core.ErrInvalidVocabularyValue)
	}

	var zero ApplicationOutcome
	if !zero.IsZero() {
		t.Error("zero-value ApplicationOutcome.IsZero() = false, want true")
	}
}

func TestValueSourceVocabulary(t *testing.T) {
	for s, source := range map[string]ValueSource{
		"peos:explicit-input": ValueSourceExplicitInput,
		"peos:default":        ValueSourceDefault,
		"peos:derived":        ValueSourceDerived,
	} {
		if source.IsZero() || source.String() != s {
			t.Errorf("ValueSource %s mismatch: %q", s, source.String())
		}
	}
	if ValueSourceDefault.Equal(ValueSourceDerived) {
		t.Error("distinct sources reported Equal")
	}
	var zero ValueSource
	if !zero.IsZero() {
		t.Error("zero-value ValueSource.IsZero() = false, want true")
	}
	custom := NewValueSource(mustVocabularyValue(t, "product", "inherited"))
	if custom.Value().String() != "product:inherited" {
		t.Errorf("Value() = %q", custom.Value().String())
	}
	data, err := json.Marshal(ValueSourceDefault)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ValueSource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(ValueSourceDefault) {
		t.Error("round trip mismatch")
	}
}

// --- ResolvedValue -----------------------------------------------------------

func TestNewResolvedValue(t *testing.T) {
	v, err := NewResolvedValue(mustLocalKey(t, "name"), "  acme  ", ValueSourceExplicitInput)
	if err != nil {
		t.Fatal(err)
	}
	if v.IsZero() {
		t.Error("valid ResolvedValue reports IsZero() = true")
	}
	if v.Parameter() != mustLocalKey(t, "name") {
		t.Error("Parameter() mismatch")
	}
	if v.Value() != "acme" {
		t.Errorf("Value() = %q, want trimmed", v.Value())
	}
	if !v.Source().Equal(ValueSourceExplicitInput) {
		t.Error("Source() mismatch")
	}
}

func TestNewResolvedValueRejections(t *testing.T) {
	key := mustLocalKey(t, "name")
	if _, err := NewResolvedValue(core.LocalKey{}, "x", ValueSourceDefault); !errors.Is(err, ErrInvalidResolvedValue) {
		t.Errorf("zero parameter key: error = %v, want %v", err, ErrInvalidResolvedValue)
	}
	if _, err := NewResolvedValue(key, "   ", ValueSourceDefault); !errors.Is(err, ErrInvalidResolvedValue) {
		t.Errorf("whitespace-only value: error = %v, want %v", err, ErrInvalidResolvedValue)
	}
	if _, err := NewResolvedValue(key, "x", ValueSource{}); !errors.Is(err, ErrInvalidResolvedValue) {
		t.Errorf("zero source: error = %v, want %v", err, ErrInvalidResolvedValue)
	}

	var zero ResolvedValue
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidResolvedValue) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidResolvedValue)
	}
}

func TestResolvedValueJSON(t *testing.T) {
	v := mustResolvedValue(t, "name", "acme", ValueSourceDefault)
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"parameter":"name","value":"acme","source":"peos:default"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded ResolvedValue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != v {
		t.Error("round trip mismatch")
	}

	before := v
	for _, payload := range []string{
		`{}`,
		`{"parameter":"name","value":"x"}`,
		`{"parameter":"","value":"x","source":"peos:default"}`,
		`{"parameter":"name","value":"","source":"peos:default"}`,
		`{"parameter":"name","value":"x","source":null}`,
		`123`,
		`not json`,
	} {
		if err := json.Unmarshal([]byte(payload), &v); err == nil {
			t.Errorf("%s accepted, want error", payload)
		}
	}
	if v != before {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- GeneratedOutput ---------------------------------------------------------

func TestNewGeneratedOutput(t *testing.T) {
	o := mustGeneratedOutput(t, "GEN-1", "REV-1")
	if o.IsZero() {
		t.Error("valid GeneratedOutput reports IsZero() = true")
	}
	if o.Artifact().ArtifactID() != mustArtifactID(t, "GEN-1") {
		t.Error("Artifact() mismatch")
	}
	if o.Revision().RevisionID() != mustArtifactRevisionID(t, "REV-1") {
		t.Error("Revision() mismatch")
	}
}

func TestNewGeneratedOutputRejections(t *testing.T) {
	artifact := mustGeneratedArtifactRef(t, "GEN-1")
	revision := mustGeneratedRevisionRef(t, "GEN-1", "REV-1")

	if _, err := NewGeneratedOutput(core.GeneratedArtifactRef{}, revision); !errors.Is(err, ErrInvalidGeneratedOutput) {
		t.Errorf("zero artifact: error = %v, want %v", err, ErrInvalidGeneratedOutput)
	}
	if _, err := NewGeneratedOutput(artifact, core.GeneratedArtifactRevisionRef{}); !errors.Is(err, ErrInvalidGeneratedOutput) {
		t.Errorf("zero revision: error = %v, want %v", err, ErrInvalidGeneratedOutput)
	}

	// The Revision must name the same Artifact as its companion reference --
	// otherwise the pair identifies two different generated Artifacts.
	mismatched := mustGeneratedRevisionRef(t, "GEN-OTHER", "REV-1")
	if _, err := NewGeneratedOutput(artifact, mismatched); !errors.Is(err, ErrInvalidGeneratedOutput) {
		t.Errorf("artifact/revision mismatch: error = %v, want %v", err, ErrInvalidGeneratedOutput)
	}

	var zero GeneratedOutput
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidGeneratedOutput) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidGeneratedOutput)
	}
}

func TestGeneratedOutputJSON(t *testing.T) {
	o := mustGeneratedOutput(t, "GEN-1", "REV-1")
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GeneratedOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != o {
		t.Error("round trip mismatch")
	}

	// It names what was generated and never holds it.
	assertNoWireKeys(t, "GeneratedOutput", o, []string{
		"content", "payload", "rendered", "result", "bytes", "body", "output",
	})

	before := o
	for _, payload := range []string{
		`{}`,
		`{"artifact":{"artifact_id":"GEN-1"}}`,
		`{"artifact":{"artifact_id":"GEN-1"},"revision":{"artifact_id":"GEN-OTHER","revision_id":"REV-1"}}`,
		`123`,
		`not json`,
	} {
		if err := json.Unmarshal([]byte(payload), &o); err == nil {
			t.Errorf("%s accepted, want error", payload)
		}
	}
	if o != before {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- ApplicationRecord: construction -----------------------------------------

func TestNewApplicationRecord(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")
	if r.IsZero() {
		t.Error("valid ApplicationRecord reports IsZero() = true")
	}
	if r.ID() != mustApplicationRecordID(t, "TAR-1") {
		t.Error("ID() mismatch")
	}
	if r.Template() != mustTemplateRevisionRef(t, "TPL-1", "REV-1") {
		t.Error("Template() mismatch")
	}
	if r.Actor() != mustActor(t) {
		t.Error("Actor() mismatch")
	}
	if r.AppliedAt() != mustTimestamp(t) {
		t.Error("AppliedAt() mismatch")
	}
	if r.Environment() != mustEnvironment(t, "ci") {
		t.Error("Environment() mismatch")
	}
	if r.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if !r.Outcome().Equal(ApplicationOutcomeIndeterminate) {
		t.Error("Outcome() mismatch")
	}
	if len(r.ResolvedValues()) != 1 {
		t.Error("ResolvedValues() mismatch")
	}
	got, ok := r.ResolvedValue(mustLocalKey(t, "name"))
	if !ok || got.Value() != "acme" {
		t.Error("ResolvedValue(name) lookup failed")
	}
	if _, ok := r.ResolvedValue(core.LocalKey{}); ok {
		t.Error("ResolvedValue(zero key) should return ok=false")
	}
	if _, ok := r.ResolvedValue(mustLocalKey(t, "missing")); ok {
		t.Error("ResolvedValue(unknown key) should return ok=false")
	}

	ref, err := r.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.RecordID() != r.ID() {
		t.Error("Ref() mismatch")
	}

	// Optional state is absent by default.
	if _, ok := r.Authority(); ok {
		t.Error("new record should have no authority")
	}
	if _, ok := r.Correction(); ok {
		t.Error("new record should have no correction reference")
	}
	if len(r.GeneratedOutputs()) != 0 || len(r.UngeneratedOutputs()) != 0 || len(r.Limitations()) != 0 {
		t.Error("new record should have no outputs or limitations")
	}
}

func TestNewApplicationRecordMandatoryFieldRejections(t *testing.T) {
	valid := func() (core.TemplateApplicationRecordID, core.TemplateArtifactRevisionRef, core.ActorRef, core.Timestamp, core.VocabularyValue, core.Provenance, ApplicationOutcome) {
		return mustApplicationRecordID(t, "TAR-1"),
			mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
			mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
			mustProvenance(t), ApplicationOutcomeIndeterminate
	}

	t.Run("zero id", func(t *testing.T) {
		_, tpl, actor, ts, env, prov, outcome := valid()
		if _, err := NewApplicationRecord(core.TemplateApplicationRecordID{}, tpl, actor, ts, env, prov, outcome, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero template revision", func(t *testing.T) {
		id, _, actor, ts, env, prov, outcome := valid()
		if _, err := NewApplicationRecord(id, core.TemplateArtifactRevisionRef{}, actor, ts, env, prov, outcome, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero actor", func(t *testing.T) {
		id, tpl, _, ts, env, prov, outcome := valid()
		if _, err := NewApplicationRecord(id, tpl, core.ActorRef{}, ts, env, prov, outcome, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero timestamp", func(t *testing.T) {
		id, tpl, actor, _, env, prov, outcome := valid()
		if _, err := NewApplicationRecord(id, tpl, actor, core.Timestamp{}, env, prov, outcome, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero environment", func(t *testing.T) {
		id, tpl, actor, ts, _, prov, outcome := valid()
		if _, err := NewApplicationRecord(id, tpl, actor, ts, core.VocabularyValue{}, prov, outcome, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero provenance", func(t *testing.T) {
		id, tpl, actor, ts, env, _, outcome := valid()
		if _, err := NewApplicationRecord(id, tpl, actor, ts, env, core.Provenance{}, outcome, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero outcome", func(t *testing.T) {
		id, tpl, actor, ts, env, prov, _ := valid()
		if _, err := NewApplicationRecord(id, tpl, actor, ts, env, prov, ApplicationOutcome{}, nil, nil, nil); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("zero-value resolved value element", func(t *testing.T) {
		id, tpl, actor, ts, env, prov, outcome := valid()
		if _, err := NewApplicationRecord(id, tpl, actor, ts, env, prov, outcome, []ResolvedValue{{}}, nil, nil); !errors.Is(err, ErrInvalidResolvedValue) {
			t.Errorf("error = %v, want %v", err, ErrInvalidResolvedValue)
		}
	})
	t.Run("one parameter resolved twice", func(t *testing.T) {
		id, tpl, actor, ts, env, prov, outcome := valid()
		values := []ResolvedValue{
			mustResolvedValue(t, "name", "first", ValueSourceExplicitInput),
			mustResolvedValue(t, "name", "second", ValueSourceDefault),
		}
		if _, err := NewApplicationRecord(id, tpl, actor, ts, env, prov, outcome, values, nil, nil); !errors.Is(err, ErrInvalidResolvedValue) {
			t.Errorf("error = %v, want %v", err, ErrInvalidResolvedValue)
		}
	})
	t.Run("zero resolved values is valid", func(t *testing.T) {
		id, tpl, actor, ts, env, prov, outcome := valid()
		if _, err := NewApplicationRecord(id, tpl, actor, ts, env, prov, outcome, nil, nil, nil); err != nil {
			t.Errorf("a parameterless template resolves nothing: unexpected error %v", err)
		}
	})
}

// --- ApplicationRecord: the outcome-conditional generated-output rule --------

// TestApplicationRecordOutcomeGeneratedOutputMatrix is the structural
// enforcement of PEOS-009's two outcome-conditional obligations: "the generated
// Artifact and exact generated Artifact Revision, where generation succeeded",
// and "A `partially succeeded` outcome SHALL explicitly identify which outputs
// were generated and which were not."
func TestApplicationRecordOutcomeGeneratedOutputMatrix(t *testing.T) {
	build := func(outcome ApplicationOutcome, generated []GeneratedOutput, ungenerated []string) error {
		r, err := NewApplicationRecord(
			mustApplicationRecordID(t, "TAR-1"),
			mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
			mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
			mustProvenance(t), outcome, nil, generated, ungenerated,
		)
		if err != nil {
			return err
		}
		if len(generated) > 0 {
			if r, err = r.WithGeneratedOutputs(generated); err != nil {
				return err
			}
		}
		if len(ungenerated) > 0 {
			if _, err = r.WithUngeneratedOutputs(ungenerated); err != nil {
				return err
			}
		}
		return nil
	}
	outputs := []GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")}
	ungenerated := []string{"the second requirement"}

	tests := []struct {
		name        string
		outcome     ApplicationOutcome
		generated   []GeneratedOutput
		ungenerated []string
		wantErr     bool
	}{
		{"succeeded with outputs", ApplicationOutcomeSucceeded, outputs, nil, false},
		{"succeeded without outputs", ApplicationOutcomeSucceeded, nil, nil, true},
		{"succeeded naming an ungenerated output", ApplicationOutcomeSucceeded, outputs, ungenerated, true},
		{"partially succeeded with both sides", ApplicationOutcomePartiallySucceeded, outputs, ungenerated, false},
		{"partially succeeded with generated only", ApplicationOutcomePartiallySucceeded, outputs, nil, true},
		{"partially succeeded with ungenerated only", ApplicationOutcomePartiallySucceeded, nil, ungenerated, true},
		{"partially succeeded with neither", ApplicationOutcomePartiallySucceeded, nil, nil, true},
		{"failed without outputs", ApplicationOutcomeFailed, nil, nil, false},
		{"failed with outputs", ApplicationOutcomeFailed, outputs, nil, true},
		{"failed naming ungenerated outputs", ApplicationOutcomeFailed, nil, ungenerated, false},
		{"interrupted without outputs", ApplicationOutcomeInterrupted, nil, nil, false},
		{"interrupted with outputs", ApplicationOutcomeInterrupted, outputs, nil, false},
		{"indeterminate without outputs", ApplicationOutcomeIndeterminate, nil, nil, false},
		{"indeterminate with outputs", ApplicationOutcomeIndeterminate, outputs, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := build(tt.outcome, tt.generated, tt.ungenerated)
			if tt.wantErr && err == nil {
				t.Error("accepted, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}

// TestApplicationRecordProductOutcomeUnconstrained confirms this package does
// not invent generated-output obligations for outcomes PEOS-009 never named:
// the vocabulary is extensible, and a Product-declared outcome is unconstrained
// in both directions rather than rejected.
func TestApplicationRecordProductOutcomeUnconstrained(t *testing.T) {
	custom := NewApplicationOutcome(mustVocabularyValue(t, "product", "queued"))
	r, err := NewApplicationRecord(
		mustApplicationRecordID(t, "TAR-1"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
		mustProvenance(t), custom, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("a Product-declared outcome with no outputs was rejected: %v", err)
	}
	if _, err := r.WithGeneratedOutputs([]GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")}); err != nil {
		t.Errorf("a Product-declared outcome with outputs was rejected: %v", err)
	}
}

// TestApplicationRecordConstructibleForEveryOutcome is the regression guard for
// the constructor-completeness defect that made generatedOutputs and
// ungeneratedOutputs constructor arguments: with outputs reachable only through
// modifiers, a succeeded or partially-succeeded record was unconstructible
// through any public path, because the constructor rejected it before a
// modifier could supply what made it valid.
func TestApplicationRecordConstructibleForEveryOutcome(t *testing.T) {
	outputs := []GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")}
	ungenerated := []string{"the second requirement"}

	for _, tt := range []struct {
		name        string
		outcome     ApplicationOutcome
		generated   []GeneratedOutput
		ungenerated []string
	}{
		{"succeeded", ApplicationOutcomeSucceeded, outputs, nil},
		{"partially succeeded", ApplicationOutcomePartiallySucceeded, outputs, ungenerated},
		{"failed", ApplicationOutcomeFailed, nil, ungenerated},
		{"interrupted", ApplicationOutcomeInterrupted, outputs, nil},
		{"indeterminate", ApplicationOutcomeIndeterminate, nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewApplicationRecord(
				mustApplicationRecordID(t, "TAR-1"),
				mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
				mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
				mustProvenance(t), tt.outcome, nil, tt.generated, tt.ungenerated,
			)
			if err != nil {
				t.Fatalf("a %s record must be constructible in one call: %v", tt.name, err)
			}
			if !r.Outcome().Equal(tt.outcome) {
				t.Error("Outcome() mismatch")
			}
		})
	}
}

// TestApplicationRecordClearingOutputsRevalidates confirms the modifiers rerun
// the outcome-conditional rule, so a succeeded record cannot be emptied of the
// generated outputs that made it valid.
func TestApplicationRecordClearingOutputsRevalidates(t *testing.T) {
	r, err := NewApplicationRecord(
		mustApplicationRecordID(t, "TAR-1"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
		mustProvenance(t), ApplicationOutcomeSucceeded, nil,
		[]GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.WithGeneratedOutputs(nil); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("clearing a succeeded record's outputs: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}
	if _, err := r.WithUngeneratedOutputs([]string{"something"}); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("naming an ungenerated output on a succeeded record: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}

	// A partially-succeeded record likewise cannot lose either side.
	p, err := NewApplicationRecord(
		mustApplicationRecordID(t, "TAR-2"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
		mustProvenance(t), ApplicationOutcomePartiallySucceeded, nil,
		[]GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")},
		[]string{"the second requirement"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.WithUngeneratedOutputs(nil); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("clearing the ungenerated side: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}
	if _, err := p.WithGeneratedOutputs(nil); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("clearing the generated side: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}
}

func TestApplicationRecordOutputRejections(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")

	if _, err := r.WithGeneratedOutputs([]GeneratedOutput{{}}); !errors.Is(err, ErrInvalidGeneratedOutput) {
		t.Errorf("zero generated output: error = %v, want %v", err, ErrInvalidGeneratedOutput)
	}
	dup := mustGeneratedOutput(t, "GEN-1", "REV-1")
	if _, err := r.WithGeneratedOutputs([]GeneratedOutput{dup, dup}); !errors.Is(err, ErrInvalidGeneratedOutput) {
		t.Errorf("duplicate generated artifact: error = %v, want %v", err, ErrInvalidGeneratedOutput)
	}
	if _, err := r.WithUngeneratedOutputs([]string{"   "}); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("whitespace-only ungenerated description: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}
	if _, err := r.WithLimitations([]string{""}); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("empty limitation: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}

	// Two different generated Artifacts are fine.
	if _, err := r.WithGeneratedOutputs([]GeneratedOutput{
		mustGeneratedOutput(t, "GEN-1", "REV-1"),
		mustGeneratedOutput(t, "GEN-2", "REV-1"),
	}); err != nil {
		t.Errorf("two distinct generated artifacts: unexpected error %v", err)
	}
}

// --- ApplicationRecord: correction -------------------------------------------

func TestApplicationRecordCorrection(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-2")
	earlier, err := core.NewTemplateApplicationRecordRef(mustApplicationRecordID(t, "TAR-1"))
	if err != nil {
		t.Fatal(err)
	}

	for _, kind := range []core.CorrectionKind{core.CorrectionKindCorrect, core.CorrectionKindReplace, core.CorrectionKindInvalidate} {
		t.Run(kind.String(), func(t *testing.T) {
			correction, err := core.NewRecordCorrectionRef(kind, earlier)
			if err != nil {
				t.Fatal(err)
			}
			corrected, err := r.WithCorrection(correction)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := corrected.Correction()
			if !ok || got.Target().RecordID() != earlier.RecordID() {
				t.Error("WithCorrection did not set the correction reference")
			}
			if got.Kind() != kind {
				t.Errorf("Kind() = %v, want %v", got.Kind(), kind)
			}
			if _, ok := corrected.WithoutCorrection().Correction(); ok {
				t.Error("WithoutCorrection did not clear the reference")
			}
		})
	}

	if _, ok := r.Correction(); ok {
		t.Error("WithCorrection mutated its receiver")
	}
}

// TestApplicationRecordSelfCorrectionRejected confirms a record cannot correct
// itself -- correction always produces a *new* record referencing an *earlier*
// one.
func TestApplicationRecordSelfCorrectionRejected(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")
	self, err := core.NewTemplateApplicationRecordRef(mustApplicationRecordID(t, "TAR-1"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, self)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.WithCorrection(correction); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("self-correction: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}

	var zero core.RecordCorrectionRef[core.TemplateApplicationRecordRef]
	if _, err := r.WithCorrection(zero); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("zero correction: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}
}

// --- ApplicationRecord: optional state ---------------------------------------

func TestApplicationRecordOptionalState(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")

	withAuthority, err := r.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withAuthority.Authority(); !ok {
		t.Error("WithAuthority did not set authority")
	}
	if _, ok := withAuthority.WithoutAuthority().Authority(); ok {
		t.Error("WithoutAuthority did not clear authority")
	}
	if _, err := r.WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("zero authority: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}

	withLimitations, err := r.WithLimitations([]string{"  sampled  "})
	if err != nil {
		t.Fatal(err)
	}
	if got := withLimitations.Limitations(); len(got) != 1 || got[0] != "sampled" {
		t.Errorf("Limitations() = %v, want trimmed", got)
	}

	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	if r.WithExtension(ext).Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !r.WithExtension(ext).WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestApplicationRecordDefensiveCopy(t *testing.T) {
	values := []ResolvedValue{mustResolvedValue(t, "name", "acme", ValueSourceExplicitInput)}
	r, err := NewApplicationRecord(
		mustApplicationRecordID(t, "TAR-1"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
		mustProvenance(t), ApplicationOutcomeIndeterminate, values, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = mustResolvedValue(t, "mutated", "mutated", ValueSourceDerived)
	if r.ResolvedValues()[0].Parameter() == mustLocalKey(t, "mutated") {
		t.Error("constructor did not defensively copy resolved values")
	}
	returned := r.ResolvedValues()
	returned[0] = mustResolvedValue(t, "mutated-again", "x", ValueSourceDerived)
	if r.ResolvedValues()[0].Parameter() == mustLocalKey(t, "mutated-again") {
		t.Error("ResolvedValues() accessor did not return a defensive copy")
	}

	outputs := []GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")}
	r2, err := r.WithGeneratedOutputs(outputs)
	if err != nil {
		t.Fatal(err)
	}
	outputs[0] = mustGeneratedOutput(t, "GEN-9", "REV-9")
	if r2.GeneratedOutputs()[0].Artifact().ArtifactID() == mustArtifactID(t, "GEN-9") {
		t.Error("modifier did not defensively copy generated outputs")
	}
	returnedOutputs := r2.GeneratedOutputs()
	returnedOutputs[0] = mustGeneratedOutput(t, "GEN-8", "REV-8")
	if r2.GeneratedOutputs()[0].Artifact().ArtifactID() == mustArtifactID(t, "GEN-8") {
		t.Error("GeneratedOutputs() accessor did not return a defensive copy")
	}
}

func TestApplicationRecordFailedModifierLeavesReceiverUnchanged(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")
	before, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.WithGeneratedOutputs([]GeneratedOutput{{}}); err == nil {
		t.Fatal("expected error")
	}
	if _, err := r.WithAuthority(core.AuthorityRef{}); err == nil {
		t.Fatal("expected error")
	}
	after, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("failed modifier mutated its receiver")
	}
}

// --- ApplicationRecord: JSON -------------------------------------------------

func TestApplicationRecordJSONRoundTrip(t *testing.T) {
	earlier, err := core.NewTemplateApplicationRecordRef(mustApplicationRecordID(t, "TAR-0"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindReplace, earlier)
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewApplicationRecord(
		mustApplicationRecordID(t, "TAR-1"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
		mustProvenance(t), ApplicationOutcomePartiallySucceeded,
		[]ResolvedValue{
			mustResolvedValue(t, "name", "acme", ValueSourceExplicitInput),
			mustResolvedValue(t, "owner", "unassigned", ValueSourceDefault),
		},
		[]GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")},
		[]string{"the second requirement"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithGeneratedOutputs([]GeneratedOutput{mustGeneratedOutput(t, "GEN-1", "REV-1")}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithUngeneratedOutputs([]string{"the second requirement"}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithAuthority(mustAuthority(t)); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithLimitations([]string{"sampled"}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithCorrection(correction); err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	r = r.WithExtension(ext)

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ApplicationRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", data2, data)
	}
	if _, ok := decoded.Correction(); !ok {
		t.Error("round trip lost the correction reference")
	}
	if _, ok := decoded.Authority(); !ok {
		t.Error("round trip lost authority")
	}
	if len(decoded.ResolvedValues()) != 2 || len(decoded.GeneratedOutputs()) != 1 || len(decoded.UngeneratedOutputs()) != 1 {
		t.Error("round trip lost collection elements")
	}
}

func TestApplicationRecordMarshalZero(t *testing.T) {
	var r ApplicationRecord
	if !r.IsZero() {
		t.Error("zero-value ApplicationRecord.IsZero() = false, want true")
	}
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}
}

func TestApplicationRecordUnmarshalRejections(t *testing.T) {
	base, err := json.Marshal(mustApplicationRecord(t, "TAR-1"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}

	mandatory := []string{"id", "template", "actor", "applied_at", "environment", "provenance", "outcome"}
	for _, field := range mandatory {
		for _, mode := range []string{"absent", "null"} {
			t.Run(mode+" "+field, func(t *testing.T) {
				mCopy := make(map[string]json.RawMessage, len(m))
				for k, v := range m {
					mCopy[k] = v
				}
				if mode == "absent" {
					delete(mCopy, field)
				} else {
					mCopy[field] = json.RawMessage(`null`)
				}
				data, err := json.Marshal(mCopy)
				if err != nil {
					t.Fatal(err)
				}
				var r ApplicationRecord
				if err := json.Unmarshal(data, &r); err == nil {
					t.Errorf("%s %s accepted, want error", mode, field)
				}
			})
		}
	}

	t.Run("null authority", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		mCopy["authority"] = json.RawMessage(`null`)
		data, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var r ApplicationRecord
		if err := json.Unmarshal(data, &r); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("null correction", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		mCopy["correction"] = json.RawMessage(`null`)
		data, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var r ApplicationRecord
		if err := json.Unmarshal(data, &r); !errors.Is(err, core.ErrInvalidCorrectionReference) {
			t.Errorf("error = %v, want %v", err, core.ErrInvalidCorrectionReference)
		}
	})
	t.Run("whitespace-only limitation", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		mCopy["limitations"] = json.RawMessage(`["   "]`)
		data, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var r ApplicationRecord
		if err := json.Unmarshal(data, &r); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("whitespace-only ungenerated output", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		mCopy["ungenerated_outputs"] = json.RawMessage(`["   "]`)
		data, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var r ApplicationRecord
		if err := json.Unmarshal(data, &r); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
	t.Run("non-object payload", func(t *testing.T) {
		var r ApplicationRecord
		if err := json.Unmarshal([]byte(`123`), &r); !errors.Is(err, ErrInvalidApplicationRecord) {
			t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
		}
	})
}

// TestApplicationRecordUnmarshalRevalidatesOutcomeRule confirms decode enforces
// the outcome-conditional rule too, so a hand-written document cannot smuggle
// in a succeeded record with no generated output.
func TestApplicationRecordUnmarshalRevalidatesOutcomeRule(t *testing.T) {
	base, err := json.Marshal(mustApplicationRecord(t, "TAR-1"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}
	m["outcome"] = json.RawMessage(`"peos:succeeded"`)
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var r ApplicationRecord
	if err := json.Unmarshal(data, &r); !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("succeeded with no generated output: error = %v, want %v", err, ErrInvalidApplicationRecord)
	}

	// A self-correcting record must be rejected on decode as well.
	m2 := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		m2[k] = v
	}
	m2["outcome"] = json.RawMessage(`"peos:indeterminate"`)
	m2["correction"] = json.RawMessage(`{"kind":"peos:correct","target":{"record_id":"TAR-1"}}`)
	data2, err := json.Marshal(m2)
	if err != nil {
		t.Fatal(err)
	}
	var r2 ApplicationRecord
	if err := json.Unmarshal(data2, &r2); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("self-correction on decode: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}
}

func TestApplicationRecordUnmarshalToleratesUnknownFieldsAndPreservesReceiver(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")
	base, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}
	m["product_specific_field"] = json.RawMessage(`{"anything":true}`)
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ApplicationRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("unknown field rejected: %v", err)
	}

	if err := json.Unmarshal([]byte(`{"id":""}`), &r); err == nil {
		t.Fatal("expected error")
	}
	if err := json.Unmarshal([]byte(`not json`), &r); err == nil {
		t.Fatal("expected error")
	}
	after, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(base) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// TestApplicationRecordNoForbiddenWireKeys is the structural proof that an
// Application Record is an immutable non-Artifact record of one completed
// application: never an Artifact, never revisioned, never lifecycle-bearing,
// and never a holder of generated content.
func TestApplicationRecordNoForbiddenWireKeys(t *testing.T) {
	r := mustApplicationRecord(t, "TAR-1")
	forbidden := []string{
		"artifact_type", "artifact_id", "revision", "revision_id", "version",
		"lifecycle", "state", "status", "current", "active", "effective",
		"template_instance", "instance", "rendered", "output_content", "payload",
		"body", "content", "compatible", "conformant", "supersession", "superseded",
	}
	assertNoWireKeys(t, "ApplicationRecord", r, forbidden)
}

// TestApplicationRecordConstructorRejectsEmptyUngeneratedDescription covers the
// constructor's own trim-and-reject pass over the ungenerated-output
// descriptions, which runs before the shared validation path.
func TestApplicationRecordConstructorRejectsEmptyUngeneratedDescription(t *testing.T) {
	_, err := NewApplicationRecord(
		mustApplicationRecordID(t, "TAR-1"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustActor(t), mustTimestamp(t), mustEnvironment(t, "ci"),
		mustProvenance(t), ApplicationOutcomeFailed, nil, nil, []string{"   "},
	)
	if !errors.Is(err, ErrInvalidApplicationRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicationRecord)
	}
}

// TestApplicationRecordDecodeRejectsMalformedOptionalRefs covers the decode
// paths for a syntactically valid but structurally wrong authority or
// correction reference.
func TestApplicationRecordDecodeRejectsMalformedOptionalRefs(t *testing.T) {
	base, err := json.Marshal(mustApplicationRecord(t, "TAR-1"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct{ name, key, value string }{
		{"malformed authority", "authority", `123`},
		{"malformed correction", "correction", `123`},
		{"correction with an empty target", "correction", `{"kind":"peos:correct","target":{"record_id":""}}`},
		// An unrecognized correction kind is deliberately NOT rejected:
		// core.NewCorrectionKind is open for forward compatibility with a
		// future specification amendment, and this package does not narrow it.
	} {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			mCopy[tt.key] = json.RawMessage(tt.value)
			data, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var r ApplicationRecord
			if err := json.Unmarshal(data, &r); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}
}
