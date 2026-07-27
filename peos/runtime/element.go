package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file defines the runtime-local vocabulary wrappers PEOS-008 needs
// (Environment, ViolationClassification, ViolationSeverity) and the one
// Contract Revision-owned value structure Packet J.1 implements: Assertion.
//
// Every type here shares one ontological status: it is Revision-owned
// value content or an open vocabulary value, never an entity. Assertion
// carries no PEOS identity, no Ref, no Artifact, no Artifact Revision, no
// revision system, and no lifecycle: "A Runtime Assertion is a Runtime
// Contract Revision-owned rule. It has no required independent identity,
// no required independent revision, and no required independent
// lifecycle."

// --- runtime-local vocabulary wrappers ---------------------------------------

// Environment is a namespaced runtime environment vocabulary value
// (PEOS-008 Runtime Contract Revision: "the environment and deployment
// scope to which it applies"; Runtime Binding Record, Runtime Observation:
// "the environment").
//
// Environment is not a closed Go enum and predeclares no constant.
// PEOS-008 names no environment vocabulary of its own -- its Non-Goals
// disclaim "a mandatory deployment technology, infrastructure platform, or
// monitoring tool" -- so what an environment value means (a Kubernetes
// namespace, a cloud region, a deployment tier, or anything else) is
// entirely Product-owned. The constructor is an infallible wrapper
// returning a bare value, following the majority convention in core's own
// vocabulary family and in quality.Unit/Scale/ThresholdOperator: validating
// the wrapper would only re-check what core.NewVocabularyValue already
// guarantees, and the aggregate constructors that consume Environment
// reject a zero one anyway.
type Environment struct{ value core.VocabularyValue }

// NewEnvironment wraps v as an Environment.
func NewEnvironment(v core.VocabularyValue) Environment { return Environment{value: v} }

// Value returns the underlying core.VocabularyValue.
func (e Environment) Value() core.VocabularyValue { return e.value }
func (e Environment) String() string              { return e.value.String() }
func (e Environment) IsZero() bool                { return e.value.IsZero() }

// Equal reports whether e and other carry the same vocabulary value.
func (e Environment) Equal(other Environment) bool { return e.value.Equal(other.value) }

func (e Environment) MarshalJSON() ([]byte, error) { return json.Marshal(e.value) }

func (e *Environment) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &e.value) }

// ViolationClassification is a namespaced Runtime Violation classification
// vocabulary value (PEOS-008 Runtime Violation: "its violation
// classification"; Runtime Contract Revision: "its violation
// classification rules"). It is a distinct Go type from ViolationSeverity
// so that one can never be passed where the other is expected.
//
// PEOS-008 names no classification vocabulary of its own; what a
// classification means, and the rules by which an observed condition maps
// to one, are Product-owned and declared as the opaque
// ContractContent.ViolationClassificationRules() descriptions. This type
// is declared in Packet J.1 alongside the rest of the package foundation,
// though it is not consumed by any J.1 field -- Runtime Violation itself is
// Packet J.2.
type ViolationClassification struct{ value core.VocabularyValue }

// NewViolationClassification wraps v as a ViolationClassification.
func NewViolationClassification(v core.VocabularyValue) ViolationClassification {
	return ViolationClassification{value: v}
}

// Value returns the underlying core.VocabularyValue.
func (c ViolationClassification) Value() core.VocabularyValue { return c.value }
func (c ViolationClassification) String() string              { return c.value.String() }
func (c ViolationClassification) IsZero() bool                { return c.value.IsZero() }

// Equal reports whether c and other carry the same vocabulary value.
func (c ViolationClassification) Equal(other ViolationClassification) bool {
	return c.value.Equal(other.value)
}

func (c ViolationClassification) MarshalJSON() ([]byte, error) { return json.Marshal(c.value) }

func (c *ViolationClassification) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.value)
}

// ViolationSeverity is a namespaced Runtime Violation severity vocabulary
// value (PEOS-008 Runtime Violation: "its severity, where applicable").
// PEOS-008 names no severity vocabulary of its own; interpretation is
// Product-owned. Declared here for the same reason as
// ViolationClassification -- not consumed until Packet J.2.
type ViolationSeverity struct{ value core.VocabularyValue }

// NewViolationSeverity wraps v as a ViolationSeverity.
func NewViolationSeverity(v core.VocabularyValue) ViolationSeverity {
	return ViolationSeverity{value: v}
}

// Value returns the underlying core.VocabularyValue.
func (s ViolationSeverity) Value() core.VocabularyValue { return s.value }
func (s ViolationSeverity) String() string              { return s.value.String() }
func (s ViolationSeverity) IsZero() bool                { return s.value.IsZero() }

// Equal reports whether s and other carry the same vocabulary value.
func (s ViolationSeverity) Equal(other ViolationSeverity) bool { return s.value.Equal(other.value) }

func (s ViolationSeverity) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }

func (s *ViolationSeverity) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.value)
}

// --- Assertion -----------------------------------------------------------------

// Assertion is a PEOS-008 Runtime Assertion: "a Runtime Contract
// Revision-owned rule evaluated against a runtime subject." It defines the
// criterion it evaluates, its observation inputs, its evaluation rule, and
// its expected result -- never the evaluation outcome itself, which is
// Packet J.2's Runtime Observation and Runtime Violation, recorded
// separately and never stored back onto the Assertion that defined them.
//
// PEOS-008 lists eight items a Runtime Assertion SHALL identify. Six are
// mandatory constructor arguments here -- the runtime-local key that names
// it, the runtime subject it evaluates, the criterion it evaluates, its
// evaluation rule, its expected result, and its scope -- because for each
// of those a zero value is ambiguous between "unstated" and a legitimate
// value.
//
// The remaining two (its observation inputs, and any temporal conditions
// or uncertainty handling it declares) are optional. An empty
// observation-input collection means none is declared as required input,
// and an absent temporal-conditions or uncertainty-handling statement
// means none is declared -- the same relationship
// quality.Measure.RequiredEvidence and
// quality.Measure.UncertaintyHandling have to their own owning aggregate.
//
// Its evaluation rule and expected result are opaque, trimmed strings.
// PEOS-008 defines no expression language, no compiled rule, and no
// evaluation engine; introducing one would be a DSL the specification does
// not state. Observation inputs are planning-time descriptions of the
// observations an evaluation is expected to consume, not citations of
// Runtime Observations that do not yet exist at declaration time -- the
// same planning-time-versus-execution-time distinction
// quality.Measure.WithRequiredEvidence documents for its own required
// Evidence descriptions.
//
// Assertion has no independent revision system of its own: changing its
// meaning requires a new Runtime Contract Revision, never an in-place
// edit, which is why every modifier here returns a copy and none of them
// can reach the key.
type Assertion struct {
	key       core.LocalKey
	subject   core.RuntimeSubjectRef
	criterion core.CriterionRef

	evaluationRule string
	expectedResult string
	scope          core.Scope

	observationInputs   []string
	temporalConditions  string
	uncertaintyHandling string
	extension           core.Extension
}

// NewAssertion validates its six arguments and returns an Assertion with no
// observation inputs, temporal conditions, uncertainty handling, or
// extension data. Use the With* methods to add those.
//
// All six are mandatory and must be non-zero, except evaluationRule and
// expectedResult, which must be non-empty after trimming; the trimmed
// values are stored. This package does not perform repository resolution
// of criterion: whether the Requirement, Runtime Contract rule, Quality
// element, or external rule it names actually exists is repository-owned
// (see doc.go).
func NewAssertion(
	key core.LocalKey,
	subject core.RuntimeSubjectRef,
	criterion core.CriterionRef,
	evaluationRule string,
	expectedResult string,
	scope core.Scope,
) (Assertion, error) {
	if key.IsZero() {
		return Assertion{}, fmt.Errorf("runtime: NewAssertion: %w: key must not be zero", ErrInvalidRuntimeAssertion)
	}
	if subject.IsZero() {
		return Assertion{}, fmt.Errorf("runtime: NewAssertion: %w: subject must not be zero", ErrInvalidRuntimeAssertion)
	}
	if criterion.IsZero() {
		return Assertion{}, fmt.Errorf("runtime: NewAssertion: %w: criterion must not be zero", ErrInvalidRuntimeAssertion)
	}
	trimmedRule, err := trimmedRequired("NewAssertion", "evaluation rule", evaluationRule, ErrInvalidRuntimeAssertion)
	if err != nil {
		return Assertion{}, err
	}
	trimmedResult, err := trimmedRequired("NewAssertion", "expected result", expectedResult, ErrInvalidRuntimeAssertion)
	if err != nil {
		return Assertion{}, err
	}
	if scope.IsZero() {
		return Assertion{}, fmt.Errorf("runtime: NewAssertion: %w: scope must not be zero", core.ErrInvalidScope)
	}
	return Assertion{
		key:            key,
		subject:        subject,
		criterion:      criterion,
		evaluationRule: trimmedRule,
		expectedResult: trimmedResult,
		scope:          scope,
	}, nil
}

// WithObservationInputs returns a copy of a with its observation-input
// descriptions set to exactly the values given, in the order given. Each
// entry is trimmed and must be non-empty after trimming. Passing an empty
// or nil slice declares none, which is why there is no
// WithoutObservationInputs: WithObservationInputs(nil) already expresses
// removal.
func (a Assertion) WithObservationInputs(descriptions []string) (Assertion, error) {
	cp, err := trimmedStringSlice("Assertion.WithObservationInputs", "observation input", descriptions, ErrInvalidRuntimeAssertion)
	if err != nil {
		return Assertion{}, err
	}
	a.observationInputs = cp
	return a, nil
}

// WithTemporalConditions returns a copy of a with its temporal conditions
// statement set. conditions must be non-empty after trimming; the trimmed
// value is stored. Use WithoutTemporalConditions to clear it.
//
// The statement is an opaque string. PEOS-008 permits a Runtime Assertion
// to declare temporal conditions without defining any temporal expression
// language, and inventing one here would be a framework the specification
// does not state.
func (a Assertion) WithTemporalConditions(conditions string) (Assertion, error) {
	trimmed, err := trimmedRequired("Assertion.WithTemporalConditions", "temporal conditions", conditions, ErrInvalidRuntimeAssertion)
	if err != nil {
		return Assertion{}, err
	}
	a.temporalConditions = trimmed
	return a, nil
}

// WithoutTemporalConditions returns a copy of a with its temporal
// conditions statement cleared.
func (a Assertion) WithoutTemporalConditions() Assertion {
	a.temporalConditions = ""
	return a
}

// WithUncertaintyHandling returns a copy of a with its statement of how
// uncertainty is handled set. handling must be non-empty after trimming;
// the trimmed value is stored. Use WithoutUncertaintyHandling to clear it.
func (a Assertion) WithUncertaintyHandling(handling string) (Assertion, error) {
	trimmed, err := trimmedRequired("Assertion.WithUncertaintyHandling", "uncertainty handling", handling, ErrInvalidRuntimeAssertion)
	if err != nil {
		return Assertion{}, err
	}
	a.uncertaintyHandling = trimmed
	return a, nil
}

// WithoutUncertaintyHandling returns a copy of a with its uncertainty
// handling statement cleared.
func (a Assertion) WithoutUncertaintyHandling() Assertion {
	a.uncertaintyHandling = ""
	return a
}

// WithExtension returns a copy of a with its extension data set. Passing
// the zero core.Extension is equivalent to declaring none.
func (a Assertion) WithExtension(extension core.Extension) Assertion {
	a.extension = extension
	return a
}

// WithoutExtension returns a copy of a with its extension data cleared.
func (a Assertion) WithoutExtension() Assertion {
	a.extension = core.Extension{}
	return a
}

// Key returns a's runtime-local key.
func (a Assertion) Key() core.LocalKey { return a.key }

// Subject returns the runtime subject a evaluates against.
func (a Assertion) Subject() core.RuntimeSubjectRef { return a.subject }

// Criterion returns the criterion a evaluates. This package does not
// resolve it against any repository; see doc.go.
func (a Assertion) Criterion() core.CriterionRef { return a.criterion }

// EvaluationRule returns a's evaluation rule, uninterpreted.
func (a Assertion) EvaluationRule() string { return a.evaluationRule }

// ExpectedResult returns a's expected result, uninterpreted.
func (a Assertion) ExpectedResult() string { return a.expectedResult }

// Scope returns a's declared scope.
func (a Assertion) Scope() core.Scope { return a.scope }

// ObservationInputs returns a defensive copy of a's observation-input
// descriptions, in declaration order. May be empty: PEOS-008 states no
// minimum cardinality.
func (a Assertion) ObservationInputs() []string { return copySlice(a.observationInputs) }

// TemporalConditions returns a's temporal conditions statement, and
// whether one is set.
func (a Assertion) TemporalConditions() (string, bool) {
	return a.temporalConditions, a.temporalConditions != ""
}

// UncertaintyHandling returns a's statement of how uncertainty is handled,
// and whether one is set.
func (a Assertion) UncertaintyHandling() (string, bool) {
	return a.uncertaintyHandling, a.uncertaintyHandling != ""
}

// Extension returns a's extension data.
func (a Assertion) Extension() core.Extension { return a.extension }

// IsZero reports whether a is the zero value.
func (a Assertion) IsZero() bool {
	return a.key.IsZero() && a.subject.IsZero() && a.criterion.IsZero() &&
		a.evaluationRule == "" && a.expectedResult == "" && a.scope.IsZero()
}

type assertionJSON struct {
	Key                 core.LocalKey          `json:"key"`
	Subject             core.RuntimeSubjectRef `json:"subject"`
	Criterion           core.CriterionRef      `json:"criterion"`
	EvaluationRule      string                 `json:"evaluation_rule"`
	ExpectedResult      string                 `json:"expected_result"`
	Scope               core.Scope             `json:"scope"`
	ObservationInputs   []string               `json:"observation_inputs,omitempty"`
	TemporalConditions  string                 `json:"temporal_conditions,omitempty"`
	UncertaintyHandling string                 `json:"uncertainty_handling,omitempty"`
	Extension           *core.Extension        `json:"extension,omitempty"`
}

// assertionUnmarshalJSON mirrors assertionJSON for decoding, with
// TemporalConditions and UncertaintyHandling captured as raw bytes so an
// explicit null can be distinguished from an absent key and rejected --
// the json.RawMessage probe technique Packet D.1 established.
// ObservationInputs needs no such treatment: absent, null, and [] all
// denote the same valid state, "none declared".
type assertionUnmarshalJSON struct {
	Key                 core.LocalKey          `json:"key"`
	Subject             core.RuntimeSubjectRef `json:"subject"`
	Criterion           core.CriterionRef      `json:"criterion"`
	EvaluationRule      string                 `json:"evaluation_rule"`
	ExpectedResult      string                 `json:"expected_result"`
	Scope               core.Scope             `json:"scope"`
	ObservationInputs   []string               `json:"observation_inputs"`
	TemporalConditions  json.RawMessage        `json:"temporal_conditions"`
	UncertaintyHandling json.RawMessage        `json:"uncertainty_handling"`
	Extension           *core.Extension        `json:"extension,omitempty"`
}

// MarshalJSON encodes a with its six mandatory keys always present, plus
// whichever optional keys are set. There is no "outcome", "result",
// "satisfied", "violated", "status", or "state" key: an Assertion defines
// a rule and never records an evaluation outcome.
func (a Assertion) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return nil, fmt.Errorf("runtime: marshal Assertion: %w", ErrInvalidRuntimeAssertion)
	}
	raw := assertionJSON{
		Key:                 a.key,
		Subject:             a.subject,
		Criterion:           a.criterion,
		EvaluationRule:      a.evaluationRule,
		ExpectedResult:      a.expectedResult,
		Scope:               a.scope,
		ObservationInputs:   a.observationInputs,
		TemporalConditions:  a.temporalConditions,
		UncertaintyHandling: a.uncertaintyHandling,
	}
	if !a.extension.IsZero() {
		raw.Extension = &a.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes a from its JSON form, applying the same validation
// as NewAssertion and each With* method. The receiver is left untouched
// unless every check passes.
func (a *Assertion) UnmarshalJSON(data []byte) error {
	var raw assertionUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal Assertion: %w: %w", ErrInvalidRuntimeAssertion, err)
	}
	result, err := NewAssertion(raw.Key, raw.Subject, raw.Criterion, raw.EvaluationRule, raw.ExpectedResult, raw.Scope)
	if err != nil {
		return err
	}
	if len(raw.ObservationInputs) > 0 {
		if result, err = result.WithObservationInputs(raw.ObservationInputs); err != nil {
			return err
		}
	}
	if len(raw.TemporalConditions) > 0 {
		if err = rejectNullRaw("Assertion", "temporal conditions", raw.TemporalConditions, ErrInvalidRuntimeAssertion); err != nil {
			return err
		}
		var conditions string
		if err = json.Unmarshal(raw.TemporalConditions, &conditions); err != nil {
			return fmt.Errorf("runtime: unmarshal Assertion: %w: %w", ErrInvalidRuntimeAssertion, err)
		}
		if result, err = result.WithTemporalConditions(conditions); err != nil {
			return err
		}
	}
	if len(raw.UncertaintyHandling) > 0 {
		if err = rejectNullRaw("Assertion", "uncertainty handling", raw.UncertaintyHandling, ErrInvalidRuntimeAssertion); err != nil {
			return err
		}
		var handling string
		if err = json.Unmarshal(raw.UncertaintyHandling, &handling); err != nil {
			return fmt.Errorf("runtime: unmarshal Assertion: %w: %w", ErrInvalidRuntimeAssertion, err)
		}
		if result, err = result.WithUncertaintyHandling(handling); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*a = result
	return nil
}
