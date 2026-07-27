package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file implements ViolationTrigger and Violation, the PEOS-008
// Runtime Violation and its triggering reference. Violation is not a
// mutable Runtime Contract field, a Lifecycle State, a Decision Outcome, a
// Validation Claim, or an Incident -- PEOS-008 states all five exclusions
// explicitly and introduces no general Incident entity, so none is added
// here.

// --- ViolationTrigger ---------------------------------------------------------

type violationTriggerKind string

const (
	violationTriggerKindObservation violationTriggerKind = "observation"
	violationTriggerKindEvidence    violationTriggerKind = "evidence"
)

// ViolationTrigger is a closed two-arm union naming the exact Runtime
// Observation or Evidence Artifact Revision that triggered a Runtime
// Violation. PEOS-008 requires "the triggering Observation or Evidence" --
// exactly one of the two, never neither and never both, and never a
// generic Artifact identity or an Incident.
//
// The two arms are asymmetric in what they can certify: an Observation
// trigger identifies an immutable record that is not itself normative
// Evidence merely by existing, while an Evidence trigger identifies
// material that has already been captured or referenced through a
// PEOS-002 Evidence-role Artifact. This package performs no implicit
// conversion between the two; the caller states which kind of trigger
// actually occurred.
type ViolationTrigger struct {
	kind        violationTriggerKind
	observation core.RuntimeObservationRef
	evidence    core.EvidenceArtifactRevisionRef
}

// NewObservationTrigger validates ref and returns a ViolationTrigger
// naming the exact Runtime Observation that triggered a Violation.
func NewObservationTrigger(ref core.RuntimeObservationRef) (ViolationTrigger, error) {
	if ref.IsZero() {
		return ViolationTrigger{}, fmt.Errorf("runtime: NewObservationTrigger: %w: observation reference must not be zero", ErrInvalidRuntimeViolation)
	}
	return ViolationTrigger{kind: violationTriggerKindObservation, observation: ref}, nil
}

// NewEvidenceTrigger validates ref and returns a ViolationTrigger naming
// the exact Evidence Artifact Revision that triggered a Violation.
func NewEvidenceTrigger(ref core.EvidenceArtifactRevisionRef) (ViolationTrigger, error) {
	if ref.IsZero() {
		return ViolationTrigger{}, fmt.Errorf("runtime: NewEvidenceTrigger: %w: evidence reference must not be zero", ErrInvalidRuntimeViolation)
	}
	return ViolationTrigger{kind: violationTriggerKindEvidence, evidence: ref}, nil
}

// Kind returns t's discriminator, "observation" or "evidence". The zero
// value returns the empty string.
func (t ViolationTrigger) Kind() string { return string(t.kind) }

// Observation returns t's triggering Runtime Observation reference, and
// whether one is set (that is, whether t is the observation arm).
func (t ViolationTrigger) Observation() (core.RuntimeObservationRef, bool) {
	if t.kind != violationTriggerKindObservation {
		return core.RuntimeObservationRef{}, false
	}
	return t.observation, true
}

// Evidence returns t's triggering Evidence Artifact Revision reference,
// and whether one is set (that is, whether t is the evidence arm).
func (t ViolationTrigger) Evidence() (core.EvidenceArtifactRevisionRef, bool) {
	if t.kind != violationTriggerKindEvidence {
		return core.EvidenceArtifactRevisionRef{}, false
	}
	return t.evidence, true
}

// IsZero reports whether t is the zero value.
func (t ViolationTrigger) IsZero() bool { return t.kind == "" }

type violationTriggerJSON struct {
	Kind        string                            `json:"kind"`
	Observation *core.RuntimeObservationRef       `json:"observation,omitempty"`
	Evidence    *core.EvidenceArtifactRevisionRef `json:"evidence,omitempty"`
}

type violationTriggerUnmarshalJSON struct {
	Kind        string          `json:"kind"`
	Observation json.RawMessage `json:"observation"`
	Evidence    json.RawMessage `json:"evidence"`
}

// MarshalJSON encodes t as {"kind":"observation","observation":{...}} or
// {"kind":"evidence","evidence":{...}}.
func (t ViolationTrigger) MarshalJSON() ([]byte, error) {
	switch t.kind {
	case violationTriggerKindObservation:
		return json.Marshal(violationTriggerJSON{Kind: string(violationTriggerKindObservation), Observation: &t.observation})
	case violationTriggerKindEvidence:
		return json.Marshal(violationTriggerJSON{Kind: string(violationTriggerKindEvidence), Evidence: &t.evidence})
	default:
		return nil, fmt.Errorf("runtime: marshal ViolationTrigger: %w", ErrInvalidRuntimeViolation)
	}
}

// UnmarshalJSON decodes t from its JSON form. An unrecognized or missing
// kind, a value carrying the wrong arm's field, a value missing its own
// arm's field, and an explicit null for the present arm are all rejected.
// The receiver is left untouched unless every check passes.
func (t *ViolationTrigger) UnmarshalJSON(data []byte) error {
	var raw violationTriggerUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: %w", ErrInvalidRuntimeViolation, err)
	}
	hasObservation := len(raw.Observation) > 0 && string(raw.Observation) != "null"
	hasEvidence := len(raw.Evidence) > 0 && string(raw.Evidence) != "null"

	var (
		result ViolationTrigger
		err    error
	)
	switch raw.Kind {
	case string(violationTriggerKindObservation):
		if hasEvidence {
			return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: an observation trigger must not carry evidence", ErrInvalidRuntimeViolation)
		}
		if !hasObservation {
			return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: an observation trigger requires an observation", ErrInvalidRuntimeViolation)
		}
		var ref core.RuntimeObservationRef
		if err = json.Unmarshal(raw.Observation, &ref); err != nil {
			return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: %w", ErrInvalidRuntimeViolation, err)
		}
		if result, err = NewObservationTrigger(ref); err != nil {
			return err
		}
	case string(violationTriggerKindEvidence):
		if hasObservation {
			return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: an evidence trigger must not carry an observation", ErrInvalidRuntimeViolation)
		}
		if !hasEvidence {
			return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: an evidence trigger requires evidence", ErrInvalidRuntimeViolation)
		}
		var ref core.EvidenceArtifactRevisionRef
		if err = json.Unmarshal(raw.Evidence, &ref); err != nil {
			return fmt.Errorf("runtime: unmarshal ViolationTrigger: %w: %w", ErrInvalidRuntimeViolation, err)
		}
		if result, err = NewEvidenceTrigger(ref); err != nil {
			return err
		}
	default:
		return fmt.Errorf("runtime: unmarshal ViolationTrigger: unrecognized kind %q: %w", raw.Kind, ErrInvalidRuntimeViolation)
	}
	*t = result
	return nil
}

// --- Violation -----------------------------------------------------------------

// Violation is a PEOS-008 Runtime Violation: "an immutable record",
// "independently identifiable", which "is not an Artifact. It is not
// revisioned. It is not lifecycle-bearing."
//
// Violation carries no correction reference: PEOS-008 documents one for
// Runtime Binding and Unbinding Records only; it names no such reference
// for Violation, and no core.RuntimeViolationRef type exists to serve as
// one (see doc.go's core-impact discussion).
//
// The applicable-Waiver field is represented as an opaque, trimmed
// description, not a typed Waiver reference: PEOS-005's Waiver
// (peos/requirement.Waiver) has no independent identity or Ref type by
// design, so no exact reference to it can be constructed without
// modifying peos/requirement or peos/core, which this packet does not do.
// This is a disclosed representational limitation (RJ-1), not a silent
// omission -- Waiver applicability is otherwise repository-derived, per
// doc.go.
type Violation struct {
	id             core.RuntimeViolationID
	subject        core.RuntimeSubjectRef
	criterion      core.CriterionRef
	trigger        ViolationTrigger
	occurredAt     core.Timestamp
	classification ViolationClassification
	scope          core.Scope
	provenance     core.Provenance

	binding          core.RuntimeBindingRecordRef
	intervalEnd      core.Timestamp
	severity         ViolationSeverity
	uncertainty      string
	limitations      []string
	applicableWaiver string
	relatedClaims    []core.ValidationClaimRef
	relatedDecisions []core.DecisionRef
	extension        core.Extension
}

// newViolation is the shared constructor path both NewViolationFromObservation
// and NewViolationFromEvidence use, after each has built the appropriate
// ViolationTrigger arm.
func newViolation(
	id core.RuntimeViolationID,
	subject core.RuntimeSubjectRef,
	criterion core.CriterionRef,
	trigger ViolationTrigger,
	occurredAt core.Timestamp,
	classification ViolationClassification,
	scope core.Scope,
	provenance core.Provenance,
) (Violation, error) {
	if id.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: id must not be zero", ErrInvalidRuntimeViolation)
	}
	if subject.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: subject must not be zero", ErrInvalidRuntimeViolation)
	}
	if criterion.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: criterion must not be zero", ErrInvalidRuntimeViolation)
	}
	if trigger.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: trigger must not be zero", ErrInvalidRuntimeViolation)
	}
	if occurredAt.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: occurred-at timestamp must not be zero", ErrInvalidRuntimeViolation)
	}
	if classification.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: classification must not be zero", ErrInvalidRuntimeViolation)
	}
	if scope.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if provenance.IsZero() {
		return Violation{}, fmt.Errorf("runtime: NewViolation: %w: provenance must not be zero", ErrInvalidRuntimeViolation)
	}
	return Violation{
		id:             id,
		subject:        subject,
		criterion:      criterion,
		trigger:        trigger,
		occurredAt:     occurredAt,
		classification: classification,
		scope:          scope,
		provenance:     provenance,
	}, nil
}

// NewViolationFromObservation validates its arguments and returns a
// Violation whose trigger is the exact Runtime Observation given.
func NewViolationFromObservation(
	id core.RuntimeViolationID,
	subject core.RuntimeSubjectRef,
	criterion core.CriterionRef,
	trigger core.RuntimeObservationRef,
	occurredAt core.Timestamp,
	classification ViolationClassification,
	scope core.Scope,
	provenance core.Provenance,
) (Violation, error) {
	t, err := NewObservationTrigger(trigger)
	if err != nil {
		return Violation{}, err
	}
	return newViolation(id, subject, criterion, t, occurredAt, classification, scope, provenance)
}

// NewViolationFromEvidence validates its arguments and returns a Violation
// whose trigger is the exact Evidence Artifact Revision given.
func NewViolationFromEvidence(
	id core.RuntimeViolationID,
	subject core.RuntimeSubjectRef,
	criterion core.CriterionRef,
	trigger core.EvidenceArtifactRevisionRef,
	occurredAt core.Timestamp,
	classification ViolationClassification,
	scope core.Scope,
	provenance core.Provenance,
) (Violation, error) {
	t, err := NewEvidenceTrigger(trigger)
	if err != nil {
		return Violation{}, err
	}
	return newViolation(id, subject, criterion, t, occurredAt, classification, scope, provenance)
}

// WithBinding returns a copy of v referencing the exact Runtime Binding
// Record in force when v occurred. binding must be non-zero. Use
// WithoutBinding to clear it.
func (v Violation) WithBinding(binding core.RuntimeBindingRecordRef) (Violation, error) {
	if binding.IsZero() {
		return Violation{}, fmt.Errorf("runtime: Violation.WithBinding: %w: binding reference must not be zero", ErrInvalidRuntimeViolation)
	}
	v.binding = binding
	return v, nil
}

// WithoutBinding returns a copy of v with its Binding Record reference
// cleared.
func (v Violation) WithoutBinding() Violation {
	v.binding = core.RuntimeBindingRecordRef{}
	return v
}

// WithInterval returns a copy of v with an interval end timestamp set,
// representing "timestamp or interval" alongside the mandatory OccurredAt
// as the interval's start. end must be non-zero and must not precede
// OccurredAt. Use WithoutInterval to clear it.
func (v Violation) WithInterval(end core.Timestamp) (Violation, error) {
	if end.IsZero() {
		return Violation{}, fmt.Errorf("runtime: Violation.WithInterval: %w: interval end must not be zero", ErrInvalidRuntimeViolation)
	}
	if end.Before(v.occurredAt) {
		return Violation{}, fmt.Errorf("runtime: Violation.WithInterval: %w: interval end must not precede the occurred-at timestamp", ErrInvalidRuntimeViolation)
	}
	v.intervalEnd = end
	return v, nil
}

// WithoutInterval returns a copy of v with its interval end cleared.
func (v Violation) WithoutInterval() Violation {
	v.intervalEnd = core.Timestamp{}
	return v
}

// WithSeverity returns a copy of v with its severity set. severity must be
// non-zero. Use WithoutSeverity to clear it. PEOS-008 requires severity
// only "where applicable".
func (v Violation) WithSeverity(severity ViolationSeverity) (Violation, error) {
	if severity.IsZero() {
		return Violation{}, fmt.Errorf("runtime: Violation.WithSeverity: %w: severity must not be zero", ErrInvalidRuntimeViolation)
	}
	v.severity = severity
	return v, nil
}

// WithoutSeverity returns a copy of v with its severity cleared.
func (v Violation) WithoutSeverity() Violation {
	v.severity = ViolationSeverity{}
	return v
}

// WithUncertainty returns a copy of v with a statement of known
// uncertainty set. uncertainty must be non-empty after trimming. Use
// WithoutUncertainty to clear it.
func (v Violation) WithUncertainty(uncertainty string) (Violation, error) {
	trimmed, err := trimmedRequired("Violation.WithUncertainty", "uncertainty", uncertainty, ErrInvalidRuntimeViolation)
	if err != nil {
		return Violation{}, err
	}
	v.uncertainty = trimmed
	return v, nil
}

// WithoutUncertainty returns a copy of v with its uncertainty statement
// cleared.
func (v Violation) WithoutUncertainty() Violation {
	v.uncertainty = ""
	return v
}

// WithLimitations returns a copy of v with its known-limitations
// descriptions set to exactly the values given, in the order given. Each
// entry is trimmed and must be non-empty after trimming. Passing an empty
// or nil slice declares none.
func (v Violation) WithLimitations(limitations []string) (Violation, error) {
	cp, err := trimmedStringSlice("Violation.WithLimitations", "limitation", limitations, ErrInvalidRuntimeViolation)
	if err != nil {
		return Violation{}, err
	}
	v.limitations = cp
	return v, nil
}

// WithApplicableWaiver returns a copy of v with an opaque description of
// an applicable PEOS-005 Waiver set. description must be non-empty after
// trimming. Use WithoutApplicableWaiver to clear it.
//
// This is a trimmed description, not a typed Waiver reference: see the
// type documentation above for why one cannot be constructed without
// modifying peos/requirement or peos/core.
func (v Violation) WithApplicableWaiver(description string) (Violation, error) {
	trimmed, err := trimmedRequired("Violation.WithApplicableWaiver", "applicable waiver", description, ErrInvalidRuntimeViolation)
	if err != nil {
		return Violation{}, err
	}
	v.applicableWaiver = trimmed
	return v, nil
}

// WithoutApplicableWaiver returns a copy of v with its applicable-Waiver
// description cleared.
func (v Violation) WithoutApplicableWaiver() Violation {
	v.applicableWaiver = ""
	return v
}

// WithRelatedClaims returns a copy of v with its related Validation Claim
// references set to exactly the values given, in the order given. A
// zero-value element is rejected. Passing an empty or nil slice declares
// none.
func (v Violation) WithRelatedClaims(claims []core.ValidationClaimRef) (Violation, error) {
	for _, c := range claims {
		if c.IsZero() {
			return Violation{}, fmt.Errorf("runtime: Violation.WithRelatedClaims: %w: claim reference must not be zero", ErrInvalidRuntimeViolation)
		}
	}
	v.relatedClaims = copySlice(claims)
	return v, nil
}

// WithRelatedDecisions returns a copy of v with its related Decision
// references set to exactly the values given, in the order given. A
// zero-value element is rejected. Passing an empty or nil slice declares
// none.
func (v Violation) WithRelatedDecisions(decisions []core.DecisionRef) (Violation, error) {
	for _, d := range decisions {
		if d.IsZero() {
			return Violation{}, fmt.Errorf("runtime: Violation.WithRelatedDecisions: %w: decision reference must not be zero", ErrInvalidRuntimeViolation)
		}
	}
	v.relatedDecisions = copySlice(decisions)
	return v, nil
}

// WithExtension returns a copy of v with its extension data set.
func (v Violation) WithExtension(extension core.Extension) Violation {
	v.extension = extension
	return v
}

// WithoutExtension returns a copy of v with its extension data cleared.
func (v Violation) WithoutExtension() Violation {
	v.extension = core.Extension{}
	return v
}

// ID returns v's identity. No core.RuntimeViolationRef type exists (see
// doc.go), so unlike BindingRecord, UnbindingRecord, and Observation,
// Violation exposes no Ref method.
func (v Violation) ID() core.RuntimeViolationID { return v.id }

// Subject returns the runtime subject v applies to.
func (v Violation) Subject() core.RuntimeSubjectRef { return v.subject }

// Criterion returns the violated Requirement, Requirement Artifact
// Revision, Runtime Contract rule, Quality criterion, or Runtime Assertion.
// This package does not resolve it against any repository.
func (v Violation) Criterion() core.CriterionRef { return v.criterion }

// Trigger returns the exact Observation or Evidence that triggered v.
func (v Violation) Trigger() ViolationTrigger { return v.trigger }

// OccurredAt returns v's occurrence timestamp (or interval start).
func (v Violation) OccurredAt() core.Timestamp { return v.occurredAt }

// Classification returns v's violation classification.
func (v Violation) Classification() ViolationClassification { return v.classification }

// Scope returns v's declared scope.
func (v Violation) Scope() core.Scope { return v.scope }

// Provenance returns v's declared provenance.
func (v Violation) Provenance() core.Provenance { return v.provenance }

// Binding returns the exact Runtime Binding Record in force when v
// occurred, and whether one is set.
func (v Violation) Binding() (core.RuntimeBindingRecordRef, bool) {
	return v.binding, !v.binding.IsZero()
}

// Interval returns v's interval end timestamp, and whether one is set.
func (v Violation) Interval() (core.Timestamp, bool) {
	return v.intervalEnd, !v.intervalEnd.IsZero()
}

// Severity returns v's severity, and whether one is set.
func (v Violation) Severity() (ViolationSeverity, bool) {
	return v.severity, !v.severity.IsZero()
}

// Uncertainty returns v's statement of known uncertainty, and whether one
// is set.
func (v Violation) Uncertainty() (string, bool) { return v.uncertainty, v.uncertainty != "" }

// Limitations returns a defensive copy of v's known-limitations
// descriptions, in declaration order.
func (v Violation) Limitations() []string { return copySlice(v.limitations) }

// ApplicableWaiver returns v's opaque applicable-Waiver description, and
// whether one is set.
func (v Violation) ApplicableWaiver() (string, bool) {
	return v.applicableWaiver, v.applicableWaiver != ""
}

// RelatedClaims returns a defensive copy of v's related Validation Claim
// references, in declaration order.
func (v Violation) RelatedClaims() []core.ValidationClaimRef { return copySlice(v.relatedClaims) }

// RelatedDecisions returns a defensive copy of v's related Decision
// references, in declaration order.
func (v Violation) RelatedDecisions() []core.DecisionRef { return copySlice(v.relatedDecisions) }

// Extension returns v's extension data.
func (v Violation) Extension() core.Extension { return v.extension }

// IsZero reports whether v is the zero value.
func (v Violation) IsZero() bool {
	return v.id.IsZero() && v.subject.IsZero() && v.criterion.IsZero() && v.trigger.IsZero() &&
		v.occurredAt.IsZero() && v.classification.IsZero() && v.scope.IsZero() && v.provenance.IsZero()
}

type violationJSON struct {
	ID               core.RuntimeViolationID       `json:"id"`
	Subject          core.RuntimeSubjectRef        `json:"subject"`
	Criterion        core.CriterionRef             `json:"criterion"`
	Trigger          ViolationTrigger              `json:"trigger"`
	OccurredAt       core.Timestamp                `json:"occurred_at"`
	Classification   ViolationClassification       `json:"classification"`
	Scope            core.Scope                    `json:"scope"`
	Provenance       core.Provenance               `json:"provenance"`
	Binding          *core.RuntimeBindingRecordRef `json:"binding,omitempty"`
	IntervalEnd      *core.Timestamp               `json:"interval_end,omitempty"`
	Severity         *ViolationSeverity            `json:"severity,omitempty"`
	Uncertainty      string                        `json:"uncertainty,omitempty"`
	Limitations      []string                      `json:"limitations,omitempty"`
	ApplicableWaiver string                        `json:"applicable_waiver,omitempty"`
	RelatedClaims    []core.ValidationClaimRef     `json:"related_claims,omitempty"`
	RelatedDecisions []core.DecisionRef            `json:"related_decisions,omitempty"`
	Extension        *core.Extension               `json:"extension,omitempty"`
}

type violationUnmarshalJSON struct {
	ID               core.RuntimeViolationID   `json:"id"`
	Subject          core.RuntimeSubjectRef    `json:"subject"`
	Criterion        core.CriterionRef         `json:"criterion"`
	Trigger          ViolationTrigger          `json:"trigger"`
	OccurredAt       core.Timestamp            `json:"occurred_at"`
	Classification   ViolationClassification   `json:"classification"`
	Scope            core.Scope                `json:"scope"`
	Provenance       core.Provenance           `json:"provenance"`
	Binding          json.RawMessage           `json:"binding"`
	IntervalEnd      json.RawMessage           `json:"interval_end"`
	Severity         json.RawMessage           `json:"severity"`
	Uncertainty      json.RawMessage           `json:"uncertainty"`
	Limitations      []string                  `json:"limitations"`
	ApplicableWaiver json.RawMessage           `json:"applicable_waiver"`
	RelatedClaims    []core.ValidationClaimRef `json:"related_claims"`
	RelatedDecisions []core.DecisionRef        `json:"related_decisions"`
	Extension        *core.Extension           `json:"extension,omitempty"`
}

// MarshalJSON encodes v with its eight mandatory keys always present, plus
// whichever optional keys are set. There is no "status", "state",
// "lifecycle", "resolved", "closed", "outcome_authority", or "incident"
// key: their absence is the structural proof that a Runtime Violation
// carries only what was recorded when it occurred, never a mutable status
// or a governance/lifecycle construct.
func (v Violation) MarshalJSON() ([]byte, error) {
	if v.IsZero() {
		return nil, fmt.Errorf("runtime: marshal Violation: %w", ErrInvalidRuntimeViolation)
	}
	raw := violationJSON{
		ID:               v.id,
		Subject:          v.subject,
		Criterion:        v.criterion,
		Trigger:          v.trigger,
		OccurredAt:       v.occurredAt,
		Classification:   v.classification,
		Scope:            v.scope,
		Provenance:       v.provenance,
		Uncertainty:      v.uncertainty,
		Limitations:      v.limitations,
		ApplicableWaiver: v.applicableWaiver,
		RelatedClaims:    v.relatedClaims,
		RelatedDecisions: v.relatedDecisions,
	}
	if !v.binding.IsZero() {
		raw.Binding = &v.binding
	}
	if !v.intervalEnd.IsZero() {
		raw.IntervalEnd = &v.intervalEnd
	}
	if !v.severity.IsZero() {
		raw.Severity = &v.severity
	}
	if !v.extension.IsZero() {
		raw.Extension = &v.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes v from its JSON form, applying the same validation
// as the two constructors and each With* method. The receiver is left
// untouched unless every check passes.
func (v *Violation) UnmarshalJSON(data []byte) error {
	var raw violationUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal Violation: %w: %w", ErrInvalidRuntimeViolation, err)
	}
	result, err := newViolation(raw.ID, raw.Subject, raw.Criterion, raw.Trigger, raw.OccurredAt, raw.Classification, raw.Scope, raw.Provenance)
	if err != nil {
		return err
	}
	if len(raw.Binding) > 0 {
		if err = rejectNullRaw("Violation", "binding", raw.Binding, ErrInvalidRuntimeViolation); err != nil {
			return err
		}
		var binding core.RuntimeBindingRecordRef
		if err = json.Unmarshal(raw.Binding, &binding); err != nil {
			return fmt.Errorf("runtime: unmarshal Violation: %w: %w", ErrInvalidRuntimeViolation, err)
		}
		if result, err = result.WithBinding(binding); err != nil {
			return err
		}
	}
	if len(raw.IntervalEnd) > 0 {
		if err = rejectNullRaw("Violation", "interval end", raw.IntervalEnd, ErrInvalidRuntimeViolation); err != nil {
			return err
		}
		var end core.Timestamp
		if err = json.Unmarshal(raw.IntervalEnd, &end); err != nil {
			return fmt.Errorf("runtime: unmarshal Violation: %w: %w", ErrInvalidRuntimeViolation, err)
		}
		if result, err = result.WithInterval(end); err != nil {
			return err
		}
	}
	if len(raw.Severity) > 0 {
		if err = rejectNullRaw("Violation", "severity", raw.Severity, ErrInvalidRuntimeViolation); err != nil {
			return err
		}
		var severity ViolationSeverity
		if err = json.Unmarshal(raw.Severity, &severity); err != nil {
			return fmt.Errorf("runtime: unmarshal Violation: %w: %w", ErrInvalidRuntimeViolation, err)
		}
		if result, err = result.WithSeverity(severity); err != nil {
			return err
		}
	}
	if len(raw.Uncertainty) > 0 {
		if err = rejectNullRaw("Violation", "uncertainty", raw.Uncertainty, ErrInvalidRuntimeViolation); err != nil {
			return err
		}
		var uncertainty string
		if err = json.Unmarshal(raw.Uncertainty, &uncertainty); err != nil {
			return fmt.Errorf("runtime: unmarshal Violation: %w: %w", ErrInvalidRuntimeViolation, err)
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
	if len(raw.ApplicableWaiver) > 0 {
		if err = rejectNullRaw("Violation", "applicable waiver", raw.ApplicableWaiver, ErrInvalidRuntimeViolation); err != nil {
			return err
		}
		var description string
		if err = json.Unmarshal(raw.ApplicableWaiver, &description); err != nil {
			return fmt.Errorf("runtime: unmarshal Violation: %w: %w", ErrInvalidRuntimeViolation, err)
		}
		if result, err = result.WithApplicableWaiver(description); err != nil {
			return err
		}
	}
	if len(raw.RelatedClaims) > 0 {
		if result, err = result.WithRelatedClaims(raw.RelatedClaims); err != nil {
			return err
		}
	}
	if len(raw.RelatedDecisions) > 0 {
		if result, err = result.WithRelatedDecisions(raw.RelatedDecisions); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*v = result
	return nil
}
