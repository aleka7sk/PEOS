package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file implements Observation, the PEOS-008 Runtime Observation:
// "an immutable record", "independently identifiable", which "is not an
// Artifact." Observation is not an invocation, an execution, an execution
// result, or a validation.ExecutionRecord specialization -- it records a
// runtime measurement or event, not the execution of a planned validation
// activity, and PEOS-008 defines no execution/invocation record of any
// kind. See doc.go for the ontology this rests on.

// Observation is a PEOS-008 Runtime Observation.
//
// Mandatory state: identity, the runtime subject, the observation
// timestamp, the observed value or event, the collection method, and the
// actor or system source, plus provenance -- for each of these, a zero
// value is ambiguous between "unstated" and a legitimate value, so none
// may be established only by a later With* call.
//
// Optional state: the exact Runtime Binding Record (where applicable), an
// interval end (representing "timestamp or time interval" alongside the
// mandatory single timestamp), a descriptor for unit/scale/event type,
// environment, known uncertainty, known limitations, and explicit Evidence
// Artifact Revision citations. For each, an empty/zero value already means
// "none declared" unambiguously, the same reasoning
// quality.Measure.RequiredEvidence and
// quality.Measure.UncertaintyHandling apply to their own optional fields.
//
// An Observation is not automatically an Evidence Artifact merely by being
// recorded: "Where a Runtime Observation must serve as normative PEOS
// Evidence, it SHALL be captured or referenced through an Artifact
// satisfying the Evidence role defined by PEOS-002." This package
// therefore only ever lets a caller cite existing
// core.EvidenceArtifactRevisionRef values explicitly; it provides no
// Observation-to-Evidence conversion that would mint an Artifact.
//
// Observation carries no correction reference: PEOS-008 documents one for
// Runtime Binding and Unbinding Records only (":282", ":314"); it names no
// such reference for Observation.
type Observation struct {
	id               core.RuntimeObservationID
	subject          core.RuntimeSubjectRef
	observedAt       core.Timestamp
	observedValue    string
	collectionMethod string
	source           core.ActorRef
	provenance       core.Provenance

	binding          core.RuntimeBindingRecordRef
	intervalEnd      core.Timestamp
	unitScaleOrEvent string
	environment      Environment
	uncertainty      string
	limitations      []string
	evidence         []core.EvidenceArtifactRevisionRef
	extension        core.Extension
}

// NewObservation validates its seven mandatory arguments and returns an
// Observation with no Binding reference, interval end, unit/scale/event
// descriptor, environment, uncertainty, limitations, Evidence, or
// extension. Use the With* methods to add those.
//
// observedValue and collectionMethod must each be non-empty after
// trimming; the trimmed values are stored. observedValue is an opaque
// string: PEOS-008 defines no typed measurement model, so a float would
// exclude event-shaped observations while a typed quantity would require a
// units framework the specification does not define. collectionMethod is
// likewise an opaque string, deliberately not core.ValidationMethod:
// PEOS-008 does not give Observation's collection method PEOS-006 Method
// ontology, unlike quality.Measure, which PEOS-007 explicitly ties to
// PEOS-006's Method vocabulary.
func NewObservation(
	id core.RuntimeObservationID,
	subject core.RuntimeSubjectRef,
	observedAt core.Timestamp,
	observedValue string,
	collectionMethod string,
	source core.ActorRef,
	provenance core.Provenance,
) (Observation, error) {
	if id.IsZero() {
		return Observation{}, fmt.Errorf("runtime: NewObservation: %w: id must not be zero", ErrInvalidRuntimeObservation)
	}
	if subject.IsZero() {
		return Observation{}, fmt.Errorf("runtime: NewObservation: %w: subject must not be zero", ErrInvalidRuntimeObservation)
	}
	if observedAt.IsZero() {
		return Observation{}, fmt.Errorf("runtime: NewObservation: %w: observed-at timestamp must not be zero", ErrInvalidRuntimeObservation)
	}
	trimmedValue, err := trimmedRequired("NewObservation", "observed value", observedValue, ErrInvalidRuntimeObservation)
	if err != nil {
		return Observation{}, err
	}
	trimmedMethod, err := trimmedRequired("NewObservation", "collection method", collectionMethod, ErrInvalidRuntimeObservation)
	if err != nil {
		return Observation{}, err
	}
	if source.IsZero() {
		return Observation{}, fmt.Errorf("runtime: NewObservation: %w: source must not be zero", ErrInvalidRuntimeObservation)
	}
	if provenance.IsZero() {
		return Observation{}, fmt.Errorf("runtime: NewObservation: %w: provenance must not be zero", ErrInvalidRuntimeObservation)
	}
	return Observation{
		id:               id,
		subject:          subject,
		observedAt:       observedAt,
		observedValue:    trimmedValue,
		collectionMethod: trimmedMethod,
		source:           source,
		provenance:       provenance,
	}, nil
}

// WithBinding returns a copy of o referencing the exact Runtime Binding
// Record it was observed under. binding must be non-zero; use
// WithoutBinding to clear it.
func (o Observation) WithBinding(binding core.RuntimeBindingRecordRef) (Observation, error) {
	if binding.IsZero() {
		return Observation{}, fmt.Errorf("runtime: Observation.WithBinding: %w: binding reference must not be zero", ErrInvalidRuntimeObservation)
	}
	o.binding = binding
	return o, nil
}

// WithoutBinding returns a copy of o with its Binding Record reference
// cleared.
func (o Observation) WithoutBinding() Observation {
	o.binding = core.RuntimeBindingRecordRef{}
	return o
}

// WithInterval returns a copy of o with an interval end timestamp set,
// representing "timestamp or time interval" alongside the mandatory
// ObservedAt as the interval's start. end must be non-zero and must not
// precede ObservedAt. Use WithoutInterval to clear it.
func (o Observation) WithInterval(end core.Timestamp) (Observation, error) {
	if end.IsZero() {
		return Observation{}, fmt.Errorf("runtime: Observation.WithInterval: %w: interval end must not be zero", ErrInvalidRuntimeObservation)
	}
	if end.Before(o.observedAt) {
		return Observation{}, fmt.Errorf("runtime: Observation.WithInterval: %w: interval end must not precede the observed-at timestamp", ErrInvalidRuntimeObservation)
	}
	o.intervalEnd = end
	return o, nil
}

// WithoutInterval returns a copy of o with its interval end cleared.
func (o Observation) WithoutInterval() Observation {
	o.intervalEnd = core.Timestamp{}
	return o
}

// WithUnitScaleOrEventType returns a copy of o with its unit, scale, or
// event type descriptor set. value must be non-empty after trimming. Use
// WithoutUnitScaleOrEventType to clear it.
//
// PEOS-008 bundles these three under a single "where applicable" item
// ("unit, scale, or event type"); this package represents them as one
// opaque descriptor rather than three typed fields, since the
// specification does not distinguish which of the three applies to any
// given Observation kind -- that distinction is Product-owned.
func (o Observation) WithUnitScaleOrEventType(value string) (Observation, error) {
	trimmed, err := trimmedRequired("Observation.WithUnitScaleOrEventType", "unit, scale, or event type", value, ErrInvalidRuntimeObservation)
	if err != nil {
		return Observation{}, err
	}
	o.unitScaleOrEvent = trimmed
	return o, nil
}

// WithoutUnitScaleOrEventType returns a copy of o with its unit/scale/
// event-type descriptor cleared.
func (o Observation) WithoutUnitScaleOrEventType() Observation {
	o.unitScaleOrEvent = ""
	return o
}

// WithEnvironment returns a copy of o with its environment or context set.
// environment must be non-zero. Use WithoutEnvironment to clear it.
func (o Observation) WithEnvironment(environment Environment) (Observation, error) {
	if environment.IsZero() {
		return Observation{}, fmt.Errorf("runtime: Observation.WithEnvironment: %w: environment must not be zero", ErrInvalidRuntimeObservation)
	}
	o.environment = environment
	return o, nil
}

// WithoutEnvironment returns a copy of o with its environment cleared.
func (o Observation) WithoutEnvironment() Observation {
	o.environment = Environment{}
	return o
}

// WithUncertainty returns a copy of o with a statement of known
// uncertainty set. uncertainty must be non-empty after trimming. Use
// WithoutUncertainty to clear it.
func (o Observation) WithUncertainty(uncertainty string) (Observation, error) {
	trimmed, err := trimmedRequired("Observation.WithUncertainty", "uncertainty", uncertainty, ErrInvalidRuntimeObservation)
	if err != nil {
		return Observation{}, err
	}
	o.uncertainty = trimmed
	return o, nil
}

// WithoutUncertainty returns a copy of o with its uncertainty statement
// cleared.
func (o Observation) WithoutUncertainty() Observation {
	o.uncertainty = ""
	return o
}

// WithLimitations returns a copy of o with its known-limitations
// descriptions set to exactly the values given, in the order given. Each
// entry is trimmed and must be non-empty after trimming. Passing an empty
// or nil slice declares none.
func (o Observation) WithLimitations(limitations []string) (Observation, error) {
	cp, err := trimmedStringSlice("Observation.WithLimitations", "limitation", limitations, ErrInvalidRuntimeObservation)
	if err != nil {
		return Observation{}, err
	}
	o.limitations = cp
	return o, nil
}

// WithEvidence returns a copy of o with the exact Evidence Artifact
// Revisions it cites set to exactly the values given, in the order given.
// A zero-value element is rejected. Passing an empty or nil slice declares
// none, which is the default: an Observation is never automatically
// Evidence merely by being recorded.
func (o Observation) WithEvidence(evidence []core.EvidenceArtifactRevisionRef) (Observation, error) {
	for _, e := range evidence {
		if e.IsZero() {
			return Observation{}, fmt.Errorf("runtime: Observation.WithEvidence: %w: evidence reference must not be zero", ErrInvalidRuntimeObservation)
		}
	}
	o.evidence = copySlice(evidence)
	return o, nil
}

// WithExtension returns a copy of o with its extension data set.
func (o Observation) WithExtension(extension core.Extension) Observation {
	o.extension = extension
	return o
}

// WithoutExtension returns a copy of o with its extension data cleared.
func (o Observation) WithoutExtension() Observation {
	o.extension = core.Extension{}
	return o
}

// ID returns o's identity.
func (o Observation) ID() core.RuntimeObservationID { return o.id }

// Ref returns a core.RuntimeObservationRef identifying o.
func (o Observation) Ref() (core.RuntimeObservationRef, error) {
	return core.NewRuntimeObservationRef(o.id)
}

// Subject returns the runtime subject o was observed on.
func (o Observation) Subject() core.RuntimeSubjectRef { return o.subject }

// ObservedAt returns o's observation timestamp (or interval start).
func (o Observation) ObservedAt() core.Timestamp { return o.observedAt }

// ObservedValue returns o's observed value or event, uninterpreted.
func (o Observation) ObservedValue() string { return o.observedValue }

// CollectionMethod returns o's collection method, uninterpreted.
func (o Observation) CollectionMethod() string { return o.collectionMethod }

// Source returns the actor or system source that produced o.
func (o Observation) Source() core.ActorRef { return o.source }

// Provenance returns o's declared provenance.
func (o Observation) Provenance() core.Provenance { return o.provenance }

// Binding returns the exact Runtime Binding Record o was observed under,
// and whether one is set.
func (o Observation) Binding() (core.RuntimeBindingRecordRef, bool) {
	return o.binding, !o.binding.IsZero()
}

// Interval returns o's interval end timestamp, and whether one is set.
func (o Observation) Interval() (core.Timestamp, bool) {
	return o.intervalEnd, !o.intervalEnd.IsZero()
}

// UnitScaleOrEventType returns o's unit, scale, or event type descriptor,
// and whether one is set.
func (o Observation) UnitScaleOrEventType() (string, bool) {
	return o.unitScaleOrEvent, o.unitScaleOrEvent != ""
}

// Environment returns o's environment or context, and whether one is set.
func (o Observation) Environment() (Environment, bool) {
	return o.environment, !o.environment.IsZero()
}

// Uncertainty returns o's statement of known uncertainty, and whether one
// is set.
func (o Observation) Uncertainty() (string, bool) { return o.uncertainty, o.uncertainty != "" }

// Limitations returns a defensive copy of o's known-limitations
// descriptions, in declaration order.
func (o Observation) Limitations() []string { return copySlice(o.limitations) }

// Evidence returns a defensive copy of o's cited Evidence Artifact
// Revisions, in declaration order. May be empty: an Observation is not
// automatically Evidence.
func (o Observation) Evidence() []core.EvidenceArtifactRevisionRef { return copySlice(o.evidence) }

// Extension returns o's extension data.
func (o Observation) Extension() core.Extension { return o.extension }

// IsZero reports whether o is the zero value.
func (o Observation) IsZero() bool {
	return o.id.IsZero() && o.subject.IsZero() && o.observedAt.IsZero() &&
		o.observedValue == "" && o.collectionMethod == "" && o.source.IsZero() && o.provenance.IsZero()
}

type observationJSON struct {
	ID               core.RuntimeObservationID          `json:"id"`
	Subject          core.RuntimeSubjectRef             `json:"subject"`
	ObservedAt       core.Timestamp                     `json:"observed_at"`
	ObservedValue    string                             `json:"observed_value"`
	CollectionMethod string                             `json:"collection_method"`
	Source           core.ActorRef                      `json:"source"`
	Provenance       core.Provenance                    `json:"provenance"`
	Binding          *core.RuntimeBindingRecordRef      `json:"binding,omitempty"`
	IntervalEnd      *core.Timestamp                    `json:"interval_end,omitempty"`
	UnitScaleOrEvent string                             `json:"unit_scale_or_event_type,omitempty"`
	Environment      *Environment                       `json:"environment,omitempty"`
	Uncertainty      string                             `json:"uncertainty,omitempty"`
	Limitations      []string                           `json:"limitations,omitempty"`
	Evidence         []core.EvidenceArtifactRevisionRef `json:"evidence,omitempty"`
	Extension        *core.Extension                    `json:"extension,omitempty"`
}

type observationUnmarshalJSON struct {
	ID               core.RuntimeObservationID          `json:"id"`
	Subject          core.RuntimeSubjectRef             `json:"subject"`
	ObservedAt       core.Timestamp                     `json:"observed_at"`
	ObservedValue    string                             `json:"observed_value"`
	CollectionMethod string                             `json:"collection_method"`
	Source           core.ActorRef                      `json:"source"`
	Provenance       core.Provenance                    `json:"provenance"`
	Binding          json.RawMessage                    `json:"binding"`
	IntervalEnd      json.RawMessage                    `json:"interval_end"`
	UnitScaleOrEvent json.RawMessage                    `json:"unit_scale_or_event_type"`
	Environment      json.RawMessage                    `json:"environment"`
	Uncertainty      json.RawMessage                    `json:"uncertainty"`
	Limitations      []string                           `json:"limitations"`
	Evidence         []core.EvidenceArtifactRevisionRef `json:"evidence"`
	Extension        *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes o with its seven mandatory keys always present, plus
// whichever optional keys are set. There is no "execution", "invocation",
// "result", "outcome", "success", or "failure" key: an Observation records
// what was observed, never the execution of an activity.
func (o Observation) MarshalJSON() ([]byte, error) {
	if o.IsZero() {
		return nil, fmt.Errorf("runtime: marshal Observation: %w", ErrInvalidRuntimeObservation)
	}
	raw := observationJSON{
		ID:               o.id,
		Subject:          o.subject,
		ObservedAt:       o.observedAt,
		ObservedValue:    o.observedValue,
		CollectionMethod: o.collectionMethod,
		Source:           o.source,
		Provenance:       o.provenance,
		UnitScaleOrEvent: o.unitScaleOrEvent,
		Uncertainty:      o.uncertainty,
		Limitations:      o.limitations,
		Evidence:         o.evidence,
	}
	if !o.binding.IsZero() {
		raw.Binding = &o.binding
	}
	if !o.intervalEnd.IsZero() {
		raw.IntervalEnd = &o.intervalEnd
	}
	if !o.environment.IsZero() {
		raw.Environment = &o.environment
	}
	if !o.extension.IsZero() {
		raw.Extension = &o.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes o from its JSON form, applying the same validation
// as NewObservation and each With* method. The receiver is left untouched
// unless every check passes.
//
// evidence: absent, explicit null, and empty array are all equivalent and
// all mean "no Evidence cited" -- the same reading as every other optional
// collection in this package.
func (o *Observation) UnmarshalJSON(data []byte) error {
	var raw observationUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal Observation: %w: %w", ErrInvalidRuntimeObservation, err)
	}
	result, err := NewObservation(raw.ID, raw.Subject, raw.ObservedAt, raw.ObservedValue, raw.CollectionMethod, raw.Source, raw.Provenance)
	if err != nil {
		return err
	}
	if len(raw.Binding) > 0 {
		if err = rejectNullRaw("Observation", "binding", raw.Binding, ErrInvalidRuntimeObservation); err != nil {
			return err
		}
		var binding core.RuntimeBindingRecordRef
		if err = json.Unmarshal(raw.Binding, &binding); err != nil {
			return fmt.Errorf("runtime: unmarshal Observation: %w: %w", ErrInvalidRuntimeObservation, err)
		}
		if result, err = result.WithBinding(binding); err != nil {
			return err
		}
	}
	if len(raw.IntervalEnd) > 0 {
		if err = rejectNullRaw("Observation", "interval end", raw.IntervalEnd, ErrInvalidRuntimeObservation); err != nil {
			return err
		}
		var end core.Timestamp
		if err = json.Unmarshal(raw.IntervalEnd, &end); err != nil {
			return fmt.Errorf("runtime: unmarshal Observation: %w: %w", ErrInvalidRuntimeObservation, err)
		}
		if result, err = result.WithInterval(end); err != nil {
			return err
		}
	}
	if len(raw.UnitScaleOrEvent) > 0 {
		if err = rejectNullRaw("Observation", "unit, scale, or event type", raw.UnitScaleOrEvent, ErrInvalidRuntimeObservation); err != nil {
			return err
		}
		var value string
		if err = json.Unmarshal(raw.UnitScaleOrEvent, &value); err != nil {
			return fmt.Errorf("runtime: unmarshal Observation: %w: %w", ErrInvalidRuntimeObservation, err)
		}
		if result, err = result.WithUnitScaleOrEventType(value); err != nil {
			return err
		}
	}
	if len(raw.Environment) > 0 {
		if err = rejectNullRaw("Observation", "environment", raw.Environment, ErrInvalidRuntimeObservation); err != nil {
			return err
		}
		var environment Environment
		if err = json.Unmarshal(raw.Environment, &environment); err != nil {
			return fmt.Errorf("runtime: unmarshal Observation: %w: %w", ErrInvalidRuntimeObservation, err)
		}
		if result, err = result.WithEnvironment(environment); err != nil {
			return err
		}
	}
	if len(raw.Uncertainty) > 0 {
		if err = rejectNullRaw("Observation", "uncertainty", raw.Uncertainty, ErrInvalidRuntimeObservation); err != nil {
			return err
		}
		var uncertainty string
		if err = json.Unmarshal(raw.Uncertainty, &uncertainty); err != nil {
			return fmt.Errorf("runtime: unmarshal Observation: %w: %w", ErrInvalidRuntimeObservation, err)
		}
		if result, err = result.WithUncertainty(uncertainty); err != nil {
			return err
		}
	}
	if len(raw.Limitations) > 0 {
		if result, err = result.WithLimitations(raw.Limitations); err != nil {
			return err
		}
	}
	if len(raw.Evidence) > 0 {
		if result, err = result.WithEvidence(raw.Evidence); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*o = result
	return nil
}
