package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- ActivityReference -------------------------------------------------------

type activityReferenceKind string

const (
	activityReferenceKindPlanned activityReferenceKind = "planned"
	activityReferenceKindAdHoc   activityReferenceKind = "ad_hoc"
)

// ActivityReference names what an ExecutionRecord executed. PEOS-006
// requires every Validation Execution Record to identify "its exact Planned
// Activity reference (Plan Revision and plan-local key), or an explicit ad
// hoc designation", and permits a record to "instead represent an explicitly
// identified ad hoc execution that was not planned by any Validation Plan".
//
// ActivityReference is a closed two-arm union expressing exactly that
// choice, with an invalid zero value so that "unset" is distinguishable from
// either arm. A *core.ValidationPlanRevisionRef plus a bare designation
// string could not express that distinction -- an empty designation would be
// indistinguishable from "ad hoc, unnamed", which PEOS-006's "explicitly
// identified" forbids.
//
// ActivityReference is a **locator, not an entity**. It carries no identity
// of its own and there is deliberately no ActivityReferenceID or
// ActivityReferenceRef type: the planned arm's core.LocalKey "does not
// survive as an independent identity outside that exact Plan Revision"
// (PEOS-006 Planned Validation Activity), and the pair (Plan Revision
// reference, plan-local key) is the only thing that resolves a Planned
// Validation Activity at all. The ad hoc arm names an execution that was
// never planned, so there is nothing for it to point at.
type ActivityReference struct {
	kind        activityReferenceKind
	planRevison core.ValidationPlanRevisionRef
	key         core.LocalKey
	designation string
}

// NewPlannedActivityReference validates planRevision and key and returns an
// ActivityReference naming one exact Planned Validation Activity. Both
// arguments are mandatory: a plan-local key is meaningless without the exact
// Plan Revision that scopes it.
func NewPlannedActivityReference(planRevision core.ValidationPlanRevisionRef, key core.LocalKey) (ActivityReference, error) {
	if planRevision.IsZero() {
		return ActivityReference{}, fmt.Errorf("validation: NewPlannedActivityReference: %w: plan revision must not be zero", ErrInvalidActivityReference)
	}
	if key.IsZero() {
		return ActivityReference{}, fmt.Errorf("validation: NewPlannedActivityReference: %w: plan-local key must not be zero", ErrInvalidActivityReference)
	}
	return ActivityReference{kind: activityReferenceKindPlanned, planRevison: planRevision, key: key}, nil
}

// NewAdHocActivityReference validates designation and returns an
// ActivityReference naming an execution that no Validation Plan planned.
// designation must be non-empty after trimming surrounding whitespace; the
// trimmed value is stored. PEOS-006 requires such an execution to be
// "explicitly identified", which an empty designation would not be.
func NewAdHocActivityReference(designation string) (ActivityReference, error) {
	trimmed := strings.TrimSpace(designation)
	if trimmed == "" {
		return ActivityReference{}, fmt.Errorf("validation: NewAdHocActivityReference: %w: designation must not be empty", ErrInvalidActivityReference)
	}
	return ActivityReference{kind: activityReferenceKindAdHoc, designation: trimmed}, nil
}

// Kind returns a's discriminator, "planned" or "ad_hoc". The zero value
// returns the empty string.
func (a ActivityReference) Kind() string { return string(a.kind) }

// AsPlanned returns a's Plan Revision reference and plan-local key, and
// whether a is the planned arm.
func (a ActivityReference) AsPlanned() (core.ValidationPlanRevisionRef, core.LocalKey, bool) {
	if a.kind != activityReferenceKindPlanned {
		return core.ValidationPlanRevisionRef{}, core.LocalKey{}, false
	}
	return a.planRevison, a.key, true
}

// AsAdHoc returns a's ad hoc designation, and whether a is the ad hoc arm.
func (a ActivityReference) AsAdHoc() (string, bool) {
	if a.kind != activityReferenceKindAdHoc {
		return "", false
	}
	return a.designation, true
}

// IsZero reports whether a is the zero value -- the unstated state PEOS-006
// does not permit on a valid ExecutionRecord.
func (a ActivityReference) IsZero() bool { return a.kind == "" }

type plannedActivityReferenceJSON struct {
	PlanRevision core.ValidationPlanRevisionRef `json:"plan_revision"`
	Key          core.LocalKey                  `json:"key"`
}

type adHocActivityReferenceJSON struct {
	Designation string `json:"designation"`
}

type activityReferenceEnvelope struct {
	Kind activityReferenceKind `json:"kind"`
	Ref  json.RawMessage       `json:"ref"`
}

// MarshalJSON encodes a as {"kind":"planned","ref":{"plan_revision":{...},
// "key":...}} or {"kind":"ad_hoc","ref":{"designation":...}}, reusing the
// {kind, ref} envelope core.EngineeringSubjectRef and core.CriterionRef
// already use for their own unions. There is no top-level PEOS type
// discriminator beyond this union's own "kind".
func (a ActivityReference) MarshalJSON() ([]byte, error) {
	var (
		refBytes []byte
		err      error
	)
	switch a.kind {
	case activityReferenceKindPlanned:
		refBytes, err = json.Marshal(plannedActivityReferenceJSON{PlanRevision: a.planRevison, Key: a.key})
	case activityReferenceKindAdHoc:
		refBytes, err = json.Marshal(adHocActivityReferenceJSON{Designation: a.designation})
	default:
		return nil, fmt.Errorf("validation: marshal ActivityReference: %w", ErrInvalidActivityReference)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(activityReferenceEnvelope{Kind: a.kind, Ref: refBytes})
}

// UnmarshalJSON decodes a from its {kind, ref} JSON form, applying the same
// validation as NewPlannedActivityReference and NewAdHocActivityReference.
//
// An unrecognized or missing kind, a missing or explicitly null "ref", a
// planned arm missing either field, and an empty ad hoc designation are all
// rejected. Cross-arm contamination cannot survive: each arm decodes into
// its own wire struct, so a planned payload's extra "designation" key and an
// ad hoc payload's extra "plan_revision" key are both ignored as unknown
// fields for that arm and can never populate the other arm's state.
func (a *ActivityReference) UnmarshalJSON(data []byte) error {
	var env activityReferenceEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("validation: unmarshal ActivityReference: %w: %w", ErrInvalidActivityReference, err)
	}
	if len(env.Ref) == 0 {
		return fmt.Errorf("validation: unmarshal ActivityReference: %w: ref is required", ErrInvalidActivityReference)
	}
	if string(env.Ref) == "null" {
		return fmt.Errorf("validation: unmarshal ActivityReference: %w: ref must not be null", ErrInvalidActivityReference)
	}

	var (
		result ActivityReference
		err    error
	)
	switch env.Kind {
	case activityReferenceKindPlanned:
		var raw plannedActivityReferenceJSON
		if err = json.Unmarshal(env.Ref, &raw); err == nil {
			result, err = NewPlannedActivityReference(raw.PlanRevision, raw.Key)
		}
	case activityReferenceKindAdHoc:
		var raw adHocActivityReferenceJSON
		if err = json.Unmarshal(env.Ref, &raw); err == nil {
			result, err = NewAdHocActivityReference(raw.Designation)
		}
	default:
		return fmt.Errorf("validation: unmarshal ActivityReference: unrecognized kind %q: %w", env.Kind, ErrInvalidActivityReference)
	}
	if err != nil {
		// Wrap both this type's own sentinel and whatever the nested core
		// type produced, so errors.Is succeeds against either -- the
		// convention core.GovernanceAction and this package's own
		// PlannedActivity already follow. Wrapping only the nested error
		// would make an invalid ActivityReference undetectable via
		// ErrInvalidActivityReference.
		return fmt.Errorf("validation: unmarshal ActivityReference: %w: %w", ErrInvalidActivityReference, err)
	}
	*a = result
	return nil
}

// --- ExecutionEvent ----------------------------------------------------------

// ExecutionEvent is one entry in an ExecutionRecord's ordered event history.
//
// [IMPLEMENTATION] PEOS-006 says only that event history, "when required by
// the applicable Product contract, is an ordered sequence of observations
// recorded within the same immutable Execution Record". It enumerates no
// fields, so this minimal shape -- a mandatory timestamp, a mandatory
// trimmed note, and an optional extension -- is an implementation choice,
// not a normative field set. A Product needing more structure uses the
// extension.
//
// ExecutionEvent is deliberately **not a PEOS Observation entity**, because
// PEOS-006 defines none: "An Observation or a Result is not a separate
// identity-bearing category distinct from Evidence. A recorded observation
// or measurement either is represented as an Evidence Artifact, or it is
// represented as content within an immutable Validation Execution Record. No
// third category is introduced." An ExecutionEvent is that second option --
// content inside the record -- and therefore carries no identity, no
// reference type, no revision, and no lifecycle. It also carries no
// severity, no criterion reference, no outcome, and no actor: PEOS-006
// defines none of those for event history, and the record as a whole already
// carries its own outcome and actor.
//
// Event order is the slice order in ExecutionRecord.Events, preserved
// verbatim. There is no sequence number, because a second ordering mechanism
// could disagree with the slice.
type ExecutionEvent struct {
	at        core.Timestamp
	note      string
	extension core.Extension
}

// NewExecutionEvent validates at and note and returns an ExecutionEvent with
// no extension data. at must be non-zero; note must be non-empty after
// trimming surrounding whitespace, and the trimmed value is stored.
func NewExecutionEvent(at core.Timestamp, note string) (ExecutionEvent, error) {
	if at.IsZero() {
		return ExecutionEvent{}, fmt.Errorf("validation: NewExecutionEvent: %w: event timestamp must not be zero", ErrInvalidExecutionRecord)
	}
	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return ExecutionEvent{}, fmt.Errorf("validation: NewExecutionEvent: %w: event note must not be empty", ErrInvalidExecutionRecord)
	}
	return ExecutionEvent{at: at, note: trimmed}, nil
}

// WithExtension returns a copy of e with its extension data set.
func (e ExecutionEvent) WithExtension(extension core.Extension) ExecutionEvent {
	e.extension = extension
	return e
}

// WithoutExtension returns a copy of e with its extension data cleared.
func (e ExecutionEvent) WithoutExtension() ExecutionEvent {
	e.extension = core.Extension{}
	return e
}

// At returns the time e was observed.
func (e ExecutionEvent) At() core.Timestamp { return e.at }

// Note returns e's recorded note.
func (e ExecutionEvent) Note() string { return e.note }

// Extension returns e's extension data.
func (e ExecutionEvent) Extension() core.Extension { return e.extension }

// IsZero reports whether e is the zero value.
func (e ExecutionEvent) IsZero() bool { return e.at.IsZero() && e.note == "" }

type executionEventJSON struct {
	At        core.Timestamp  `json:"at"`
	Note      string          `json:"note"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes e as {"at":..., "note":...}, plus extension when set.
func (e ExecutionEvent) MarshalJSON() ([]byte, error) {
	if e.IsZero() {
		return nil, fmt.Errorf("validation: marshal ExecutionEvent: %w", ErrInvalidExecutionRecord)
	}
	raw := executionEventJSON{At: e.at, Note: e.note}
	if !e.extension.IsZero() {
		raw.Extension = &e.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes e from its JSON form, applying the same validation
// as NewExecutionEvent.
//
// Missing-versus-null: a missing "at" leaves the timestamp zero and is
// rejected by NewExecutionEvent; an explicit null invokes core.Timestamp's
// own UnmarshalJSON, which fails there, so the error is wrapped here with
// ErrInvalidExecutionRecord. A missing or null "note" both yield the empty
// string and are rejected. extension null is equivalent to absent, per
// core.Extension's contract.
func (e *ExecutionEvent) UnmarshalJSON(data []byte) error {
	var raw executionEventJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal ExecutionEvent: %w: %w", ErrInvalidExecutionRecord, err)
	}
	result, err := NewExecutionEvent(raw.At, raw.Note)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*e = result
	return nil
}

// --- ExecutionRecord ---------------------------------------------------------

// ExecutionRecord is a PEOS-006 Validation Execution Record: "an immutable
// record", "independently identifiable", which "is not an Artifact. It is
// not revisioned. It is not lifecycle-bearing." It "represents one attempted
// execution of a Validation Activity".
//
// Structurally: it carries core.ValidationExecutionRecordID as its own
// identity, composes no core.Artifact and no core.ArtifactRevision, has no
// Core accessor, composes no relation.Relation (PEOS-006 defines no Artifact
// Relation), and carries no Lifecycle State, State Assignment, or status
// field. Its outcome is a core.ExecutionOutcome vocabulary value, which
// PEOS-006 requires to be exactly that: outcomes "are not Lifecycle States,
// and they SHALL NOT be represented as Lifecycle States or through a State
// Assignment" (non-conforming patterns "Mutable Execution Record",
// "Execution Outcome as Lifecycle State", "Lifecycle-Governed Validation
// Activity").
//
// It is not a Validation Claim. An execution outcome says whether the
// activity ran; a Claim outcome says what was determined about a Subject. A
// completed execution may carry a not-satisfied Claim, and PEOS-006 requires
// neither to imply the other: "A Validation Claim is not required for every
// Validation Execution Record to exist, and a Validation Execution Record is
// not required for every Validation Claim to exist."
//
// Immutability is structural: every field is unexported, every modifier
// returns a copy, and no modifier touches a mandatory field. "Correction of a
// Validation Execution Record creates a new Validation Execution Record. A
// Validation Execution Record SHALL NOT be mutated once recorded." The
// correction field on the *new* record points backward at the earlier one,
// so no already-written record is ever rewritten to record that it was
// corrected.
type ExecutionRecord struct {
	id          core.ValidationExecutionRecordID
	activity    ActivityReference
	subject     core.EngineeringSubjectRef
	method      core.ValidationMethod
	outcome     core.ExecutionOutcome
	completedAt core.Timestamp
	actor       core.ActorRef
	provenance  core.Provenance

	startedAt          core.Timestamp
	criteria           []core.CriterionRef
	events             []ExecutionEvent
	producedEvidence   []core.EvidenceArtifactRevisionRef
	reliedUponEvidence []core.EvidenceArtifactRevisionRef
	authority          core.AuthorityRef
	environment        string
	limitations        string
	uncertainty        string
	correction         core.RecordCorrectionRef[core.ValidationExecutionRecordRef]
	extension          core.Extension
}

// NewExecutionRecord validates its eight mandatory arguments, in the order
// listed, and returns an ExecutionRecord with no optional field set. Use the
// With* methods to add those.
//
// All eight are mandatory per PEOS-006's own "SHALL identify, at minimum"
// list: its own record identity; its exact Planned Activity reference or an
// explicit ad hoc designation; its Subject and exact participant level; the
// Validation Method used; its outcome; the completed or terminated
// timestamp; the responsible actor; and provenance.
//
// completedAt is mandatory while startedAt is not, because PEOS-006
// qualifies only the latter ("the started timestamp, where known") and
// leaves "the completed or terminated timestamp" unqualified -- an
// interrupted or failed execution still terminated at some point.
//
// A successful call always returns a fully valid ExecutionRecord: no
// mandatory field is reachable through a later With* call.
func NewExecutionRecord(
	id core.ValidationExecutionRecordID,
	activity ActivityReference,
	subject core.EngineeringSubjectRef,
	method core.ValidationMethod,
	outcome core.ExecutionOutcome,
	completedAt core.Timestamp,
	actor core.ActorRef,
	provenance core.Provenance,
) (ExecutionRecord, error) {
	if id.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: record id must not be zero", ErrInvalidExecutionRecord)
	}
	if activity.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: activity reference must be explicitly stated", ErrInvalidActivityReference)
	}
	if subject.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: subject must not be zero", ErrInvalidExecutionRecord)
	}
	if method.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: validation method must not be zero", ErrInvalidExecutionRecord)
	}
	if outcome.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: execution outcome must not be zero", ErrInvalidExecutionRecord)
	}
	if completedAt.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: completed timestamp must not be zero", ErrInvalidExecutionRecord)
	}
	if actor.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: responsible actor must not be zero", ErrInvalidExecutionRecord)
	}
	if provenance.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: NewExecutionRecord: %w: provenance must not be zero", ErrInvalidExecutionRecord)
	}
	return ExecutionRecord{
		id:          id,
		activity:    activity,
		subject:     subject,
		method:      method,
		outcome:     outcome,
		completedAt: completedAt,
		actor:       actor,
		provenance:  provenance,
	}, nil
}

// WithStartedAt returns a copy of r with its started timestamp set. at must
// be non-zero and must not be after r's completed timestamp; equality is
// permitted, since PEOS-006 imposes no minimum duration on an execution. No
// other temporal semantics are inferred -- PEOS-006 defines none. Use
// WithoutStartedAt to clear it.
func (r ExecutionRecord) WithStartedAt(at core.Timestamp) (ExecutionRecord, error) {
	if at.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: ExecutionRecord.WithStartedAt: %w: started timestamp must not be zero", ErrInvalidExecutionRecord)
	}
	if at.After(r.completedAt) {
		return ExecutionRecord{}, fmt.Errorf("validation: ExecutionRecord.WithStartedAt: %w: started timestamp must not be after the completed timestamp", ErrInvalidExecutionRecord)
	}
	r.startedAt = at
	return r, nil
}

// WithoutStartedAt returns a copy of r with its started timestamp cleared.
func (r ExecutionRecord) WithoutStartedAt() ExecutionRecord {
	r.startedAt = core.Timestamp{}
	return r
}

// WithCriteria returns a copy of r with the criteria it evaluated set to
// exactly the values given, in the order given. A zero-value
// core.CriterionRef is rejected. An empty or nil slice declares none, which
// PEOS-006 permits.
func (r ExecutionRecord) WithCriteria(criteria []core.CriterionRef) (ExecutionRecord, error) {
	cp, err := copyCriteria("ExecutionRecord.WithCriteria", criteria, ErrInvalidExecutionRecord)
	if err != nil {
		return ExecutionRecord{}, err
	}
	r.criteria = cp
	return r, nil
}

// WithEvents returns a copy of r with its ordered event history set to
// exactly the values given, in the order given. A zero-value ExecutionEvent
// is rejected. An empty or nil slice declares none; PEOS-006 requires event
// history only "where required by the applicable Product contract".
func (r ExecutionRecord) WithEvents(events []ExecutionEvent) (ExecutionRecord, error) {
	if len(events) == 0 {
		r.events = nil
		return r, nil
	}
	cp := make([]ExecutionEvent, len(events))
	for idx, e := range events {
		if e.IsZero() {
			return ExecutionRecord{}, fmt.Errorf("validation: ExecutionRecord.WithEvents: %w: event must not be zero", ErrInvalidExecutionRecord)
		}
		cp[idx] = e
	}
	r.events = cp
	return r, nil
}

// WithProducedEvidence returns a copy of r with the Evidence Artifact
// Revisions its execution produced set to exactly the values given, in the
// order given. PEOS-006 distinguishes Evidence "produced by" an execution
// from Evidence "consumed by" it, which is why this and
// WithReliedUponEvidence are separate.
func (r ExecutionRecord) WithProducedEvidence(evidence []core.EvidenceArtifactRevisionRef) (ExecutionRecord, error) {
	cp, err := copyEvidence("ExecutionRecord.WithProducedEvidence", evidence, ErrInvalidExecutionRecord)
	if err != nil {
		return ExecutionRecord{}, err
	}
	r.producedEvidence = cp
	return r, nil
}

// WithReliedUponEvidence returns a copy of r with the Evidence Artifact
// Revisions its execution relied upon as input set to exactly the values
// given, in the order given.
func (r ExecutionRecord) WithReliedUponEvidence(evidence []core.EvidenceArtifactRevisionRef) (ExecutionRecord, error) {
	cp, err := copyEvidence("ExecutionRecord.WithReliedUponEvidence", evidence, ErrInvalidExecutionRecord)
	if err != nil {
		return ExecutionRecord{}, err
	}
	r.reliedUponEvidence = cp
	return r, nil
}

// WithAuthority returns a copy of r with its authority basis set. PEOS-006
// requires this only "where required", so it is optional. authority must be
// non-zero; use WithoutAuthority to clear it.
func (r ExecutionRecord) WithAuthority(authority core.AuthorityRef) (ExecutionRecord, error) {
	if authority.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: ExecutionRecord.WithAuthority: %w: authority must not be zero", ErrInvalidExecutionRecord)
	}
	r.authority = authority
	return r, nil
}

// WithoutAuthority returns a copy of r with its authority basis cleared.
func (r ExecutionRecord) WithoutAuthority() ExecutionRecord {
	r.authority = core.AuthorityRef{}
	return r
}

// WithEnvironment returns a copy of r with its environment or execution
// context set. The value is trimmed and must be non-empty after trimming.
func (r ExecutionRecord) WithEnvironment(environment string) (ExecutionRecord, error) {
	trimmed, err := requireTrimmed("ExecutionRecord.WithEnvironment", "environment", environment, ErrInvalidExecutionRecord)
	if err != nil {
		return ExecutionRecord{}, err
	}
	r.environment = trimmed
	return r, nil
}

// WithoutEnvironment returns a copy of r with its environment cleared.
func (r ExecutionRecord) WithoutEnvironment() ExecutionRecord {
	r.environment = ""
	return r
}

// WithLimitations returns a copy of r with its known limitations set. The
// value is trimmed and must be non-empty after trimming.
func (r ExecutionRecord) WithLimitations(limitations string) (ExecutionRecord, error) {
	trimmed, err := requireTrimmed("ExecutionRecord.WithLimitations", "limitations", limitations, ErrInvalidExecutionRecord)
	if err != nil {
		return ExecutionRecord{}, err
	}
	r.limitations = trimmed
	return r, nil
}

// WithoutLimitations returns a copy of r with its known limitations cleared.
func (r ExecutionRecord) WithoutLimitations() ExecutionRecord {
	r.limitations = ""
	return r
}

// WithUncertainty returns a copy of r with its known uncertainty set. The
// value is trimmed and must be non-empty after trimming.
func (r ExecutionRecord) WithUncertainty(uncertainty string) (ExecutionRecord, error) {
	trimmed, err := requireTrimmed("ExecutionRecord.WithUncertainty", "uncertainty", uncertainty, ErrInvalidExecutionRecord)
	if err != nil {
		return ExecutionRecord{}, err
	}
	r.uncertainty = trimmed
	return r, nil
}

// WithoutUncertainty returns a copy of r with its known uncertainty cleared.
func (r ExecutionRecord) WithoutUncertainty() ExecutionRecord {
	r.uncertainty = ""
	return r
}

// WithCorrection returns a copy of r declaring that r corrects, replaces, or
// invalidates an earlier Validation Execution Record, identified exactly.
//
// This is not Artifact Supersession and is not an Artifact Relation:
// PEOS-006 states that such a relationship "is not Artifact Supersession" and
// that "The normative term Supersession SHALL NOT be used for Validation
// Execution Record correction or replacement". core.CorrectionKind offers
// only correct, replace, and invalidate for exactly that reason. Setting a
// correction reference mutates nothing: the earlier record "remains
// historically preserved; it is not erased or overwritten".
//
// Whether the target exists, whether the chronology is coherent, and whether
// a correction chain is consistent are repository concerns -- this value
// layer holds one record and cannot see the other.
func (r ExecutionRecord) WithCorrection(correction core.RecordCorrectionRef[core.ValidationExecutionRecordRef]) (ExecutionRecord, error) {
	if correction.IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: ExecutionRecord.WithCorrection: %w: correction must not be zero", ErrInvalidExecutionRecord)
	}
	if correction.Target().IsZero() {
		return ExecutionRecord{}, fmt.Errorf("validation: ExecutionRecord.WithCorrection: %w: correction target must not be zero", ErrInvalidExecutionRecord)
	}
	r.correction = correction
	return r, nil
}

// WithoutCorrection returns a copy of r with its correction reference
// cleared.
func (r ExecutionRecord) WithoutCorrection() ExecutionRecord {
	r.correction = core.RecordCorrectionRef[core.ValidationExecutionRecordRef]{}
	return r
}

// WithExtension returns a copy of r with its extension data set.
func (r ExecutionRecord) WithExtension(extension core.Extension) ExecutionRecord {
	r.extension = extension
	return r
}

// WithoutExtension returns a copy of r with its extension data cleared.
func (r ExecutionRecord) WithoutExtension() ExecutionRecord {
	r.extension = core.Extension{}
	return r
}

// ID returns r's own record identity.
func (r ExecutionRecord) ID() core.ValidationExecutionRecordID { return r.id }

// Ref returns a core.ValidationExecutionRecordRef citing r. This is the
// reference a Validation Claim uses to name a relevant execution, and the
// reference a later record's correction targets.
func (r ExecutionRecord) Ref() (core.ValidationExecutionRecordRef, error) {
	return core.NewValidationExecutionRecordRef(r.id)
}

// Activity returns what r executed: an exact Planned Validation Activity, or
// an explicit ad hoc designation.
func (r ExecutionRecord) Activity() ActivityReference { return r.activity }

// Subject returns r's single Validation Subject. Its participant level is
// carried by the returned core.EngineeringSubjectRef's own arm.
func (r ExecutionRecord) Subject() core.EngineeringSubjectRef { return r.subject }

// Method returns the Validation Method r used.
func (r ExecutionRecord) Method() core.ValidationMethod { return r.method }

// Outcome returns r's execution outcome. This describes whether the activity
// ran, never what was determined about the Subject -- see Claim.Outcome.
func (r ExecutionRecord) Outcome() core.ExecutionOutcome { return r.outcome }

// CompletedAt returns the time r's execution completed or terminated.
func (r ExecutionRecord) CompletedAt() core.Timestamp { return r.completedAt }

// Actor returns r's responsible actor.
func (r ExecutionRecord) Actor() core.ActorRef { return r.actor }

// Provenance returns r's provenance.
func (r ExecutionRecord) Provenance() core.Provenance { return r.provenance }

// StartedAt returns the time r's execution started, and whether it is known.
func (r ExecutionRecord) StartedAt() (core.Timestamp, bool) {
	return r.startedAt, !r.startedAt.IsZero()
}

// Criteria returns a defensive copy of the criteria r evaluated, in
// declaration order.
func (r ExecutionRecord) Criteria() []core.CriterionRef {
	if len(r.criteria) == 0 {
		return nil
	}
	cp := make([]core.CriterionRef, len(r.criteria))
	copy(cp, r.criteria)
	return cp
}

// Events returns a defensive copy of r's ordered event history. Slice order
// is the event order.
func (r ExecutionRecord) Events() []ExecutionEvent {
	if len(r.events) == 0 {
		return nil
	}
	cp := make([]ExecutionEvent, len(r.events))
	copy(cp, r.events)
	return cp
}

// ProducedEvidence returns a defensive copy of the Evidence Artifact
// Revisions r's execution produced.
func (r ExecutionRecord) ProducedEvidence() []core.EvidenceArtifactRevisionRef {
	return copyEvidenceOut(r.producedEvidence)
}

// ReliedUponEvidence returns a defensive copy of the Evidence Artifact
// Revisions r's execution relied upon.
func (r ExecutionRecord) ReliedUponEvidence() []core.EvidenceArtifactRevisionRef {
	return copyEvidenceOut(r.reliedUponEvidence)
}

// Authority returns r's declared authority basis, and whether one is set.
func (r ExecutionRecord) Authority() (core.AuthorityRef, bool) {
	return r.authority, !r.authority.IsZero()
}

// Environment returns r's declared environment or execution context, and
// whether one is set.
func (r ExecutionRecord) Environment() (string, bool) {
	return r.environment, r.environment != ""
}

// Limitations returns r's declared known limitations, and whether any are
// set.
func (r ExecutionRecord) Limitations() (string, bool) {
	return r.limitations, r.limitations != ""
}

// Uncertainty returns r's declared known uncertainty, and whether any is
// set.
func (r ExecutionRecord) Uncertainty() (string, bool) {
	return r.uncertainty, r.uncertainty != ""
}

// Correction returns r's declared correction reference to an earlier
// Validation Execution Record, and whether one is set.
func (r ExecutionRecord) Correction() (core.RecordCorrectionRef[core.ValidationExecutionRecordRef], bool) {
	return r.correction, !r.correction.IsZero()
}

// Extension returns r's extension data.
func (r ExecutionRecord) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r ExecutionRecord) IsZero() bool {
	return r.id.IsZero() && r.activity.IsZero() && r.subject.IsZero() && r.method.IsZero() &&
		r.outcome.IsZero() && r.completedAt.IsZero() && r.actor.IsZero() && r.provenance.IsZero()
}

type executionRecordJSON struct {
	ID          core.ValidationExecutionRecordID `json:"id"`
	Activity    ActivityReference                `json:"activity"`
	Subject     core.EngineeringSubjectRef       `json:"subject"`
	Method      core.ValidationMethod            `json:"method"`
	Outcome     core.ExecutionOutcome            `json:"outcome"`
	CompletedAt core.Timestamp                   `json:"completed_at"`
	Actor       core.ActorRef                    `json:"actor"`
	Provenance  core.Provenance                  `json:"provenance"`

	StartedAt          *core.Timestamp                                              `json:"started_at,omitempty"`
	Criteria           []core.CriterionRef                                          `json:"criteria,omitempty"`
	Events             []ExecutionEvent                                             `json:"events,omitempty"`
	ProducedEvidence   []core.EvidenceArtifactRevisionRef                           `json:"produced_evidence,omitempty"`
	ReliedUponEvidence []core.EvidenceArtifactRevisionRef                           `json:"relied_upon_evidence,omitempty"`
	Authority          *core.AuthorityRef                                           `json:"authority,omitempty"`
	Environment        string                                                       `json:"environment,omitempty"`
	Limitations        string                                                       `json:"limitations,omitempty"`
	Uncertainty        string                                                       `json:"uncertainty,omitempty"`
	Correction         *core.RecordCorrectionRef[core.ValidationExecutionRecordRef] `json:"correction,omitempty"`
	Extension          *core.Extension                                              `json:"extension,omitempty"`
}

// executionRecordUnmarshalJSON mirrors executionRecordJSON for decoding
// only, capturing every optional single-value key as raw bytes so an
// explicit JSON null can be distinguished from an absent key and rejected --
// the technique Packet D.1 established for relation.Relation's Scope and
// Packet H.1 reused for PlannedActivity's optional single values.
//
// The four optional collections deliberately keep plain typed fields: for
// criteria, events, produced_evidence, and relied_upon_evidence, an absent
// key, an explicit null, and an empty array all denote the same valid
// state -- "none declared" -- because PEOS-006 permits zero cardinality for
// each ("zero or more Evidence Artifact Revisions"; event history only
// "where required by the applicable Product contract"). Distinguishing them
// would create a difference with no semantic content. This is the same
// decision Packet H.1 made and documented for PlannedActivity's optional
// collections, and it is deliberately NOT how Claim's criteria and evidence
// behave -- see claim.go.
type executionRecordUnmarshalJSON struct {
	ID          core.ValidationExecutionRecordID `json:"id"`
	Activity    ActivityReference                `json:"activity"`
	Subject     core.EngineeringSubjectRef       `json:"subject"`
	Method      core.ValidationMethod            `json:"method"`
	Outcome     core.ExecutionOutcome            `json:"outcome"`
	CompletedAt core.Timestamp                   `json:"completed_at"`
	Actor       core.ActorRef                    `json:"actor"`
	Provenance  core.Provenance                  `json:"provenance"`

	StartedAt          json.RawMessage                    `json:"started_at"`
	Criteria           []core.CriterionRef                `json:"criteria"`
	Events             []ExecutionEvent                   `json:"events"`
	ProducedEvidence   []core.EvidenceArtifactRevisionRef `json:"produced_evidence"`
	ReliedUponEvidence []core.EvidenceArtifactRevisionRef `json:"relied_upon_evidence"`
	Authority          json.RawMessage                    `json:"authority"`
	Environment        json.RawMessage                    `json:"environment"`
	Limitations        json.RawMessage                    `json:"limitations"`
	Uncertainty        json.RawMessage                    `json:"uncertainty"`
	Correction         json.RawMessage                    `json:"correction"`
	Extension          *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"id":..., "activity":..., "subject":...,
// "method":..., "outcome":..., "completed_at":..., "actor":...,
// "provenance":...}, plus whichever optional keys are set.
//
// There is no top-level type discriminator, no "relation" envelope, no
// Artifact or Revision identity, and no lifecycle or status key -- their
// absence is the structural proof that a Validation Execution Record is
// neither an Artifact, a relationship, nor a lifecycle-bearing entity.
func (r ExecutionRecord) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("validation: marshal ExecutionRecord: %w", ErrInvalidExecutionRecord)
	}
	raw := executionRecordJSON{
		ID:          r.id,
		Activity:    r.activity,
		Subject:     r.subject,
		Method:      r.method,
		Outcome:     r.outcome,
		CompletedAt: r.completedAt,
		Actor:       r.actor,
		Provenance:  r.provenance,
		Environment: r.environment,
		Limitations: r.limitations,
		Uncertainty: r.uncertainty,
	}
	if !r.startedAt.IsZero() {
		raw.StartedAt = &r.startedAt
	}
	if len(r.criteria) > 0 {
		raw.Criteria = r.criteria
	}
	if len(r.events) > 0 {
		raw.Events = r.events
	}
	if len(r.producedEvidence) > 0 {
		raw.ProducedEvidence = r.producedEvidence
	}
	if len(r.reliedUponEvidence) > 0 {
		raw.ReliedUponEvidence = r.reliedUponEvidence
	}
	if !r.authority.IsZero() {
		raw.Authority = &r.authority
	}
	if !r.correction.IsZero() {
		raw.Correction = &r.correction
	}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes r from its JSON form, applying the same validation
// as NewExecutionRecord and each With* method, so a decoded record can never
// be constructor-impossible -- including the started/completed ordering
// rule.
//
// Missing-versus-null, stated exactly rather than assumed:
//
//   - id, subject, method, outcome, completed_at, actor, provenance: a
//     missing key leaves the field zero and reaches NewExecutionRecord,
//     which rejects it with ErrInvalidExecutionRecord. An explicit null
//     instead invokes that nested type's own UnmarshalJSON, which fails
//     there, so the error is wrapped here with ErrInvalidExecutionRecord
//     plus that type's own sentinel. Both are rejected; the sentinel sets
//     differ.
//   - activity: a missing key yields a zero ActivityReference, rejected with
//     ErrInvalidActivityReference; an explicit null is rejected by
//     ActivityReference's own UnmarshalJSON.
//   - criteria, events, produced_evidence, relied_upon_evidence: absent,
//     explicit null, and empty array are equivalent, all meaning "none
//     declared".
//   - started_at, authority, environment, limitations, uncertainty,
//     correction: a missing key means absent; an explicit null is rejected
//     rather than silently treated as absent.
//   - extension: null is equivalent to absent, per core.Extension's contract.
func (r *ExecutionRecord) UnmarshalJSON(data []byte) error {
	var raw executionRecordUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal ExecutionRecord: %w: %w", ErrInvalidExecutionRecord, err)
	}

	result, err := NewExecutionRecord(raw.ID, raw.Activity, raw.Subject, raw.Method, raw.Outcome, raw.CompletedAt, raw.Actor, raw.Provenance)
	if err != nil {
		return err
	}

	if err := rejectNull("ExecutionRecord", "started_at", raw.StartedAt, ErrInvalidExecutionRecord); err != nil {
		return err
	}
	if len(raw.StartedAt) > 0 {
		var at core.Timestamp
		if err := json.Unmarshal(raw.StartedAt, &at); err != nil {
			return fmt.Errorf("validation: unmarshal ExecutionRecord: %w: %w", ErrInvalidExecutionRecord, err)
		}
		if result, err = result.WithStartedAt(at); err != nil {
			return err
		}
	}
	if len(raw.Criteria) > 0 {
		if result, err = result.WithCriteria(raw.Criteria); err != nil {
			return err
		}
	}
	if len(raw.Events) > 0 {
		if result, err = result.WithEvents(raw.Events); err != nil {
			return err
		}
	}
	if len(raw.ProducedEvidence) > 0 {
		if result, err = result.WithProducedEvidence(raw.ProducedEvidence); err != nil {
			return err
		}
	}
	if len(raw.ReliedUponEvidence) > 0 {
		if result, err = result.WithReliedUponEvidence(raw.ReliedUponEvidence); err != nil {
			return err
		}
	}
	if err := rejectNull("ExecutionRecord", "authority", raw.Authority, ErrInvalidExecutionRecord); err != nil {
		return err
	}
	if len(raw.Authority) > 0 {
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("validation: unmarshal ExecutionRecord: %w: %w", ErrInvalidExecutionRecord, err)
		}
		if result, err = result.WithAuthority(authority); err != nil {
			return err
		}
	}
	for _, field := range []struct {
		name string
		raw  json.RawMessage
		set  func(ExecutionRecord, string) (ExecutionRecord, error)
	}{
		{"environment", raw.Environment, ExecutionRecord.WithEnvironment},
		{"limitations", raw.Limitations, ExecutionRecord.WithLimitations},
		{"uncertainty", raw.Uncertainty, ExecutionRecord.WithUncertainty},
	} {
		if err := rejectNull("ExecutionRecord", field.name, field.raw, ErrInvalidExecutionRecord); err != nil {
			return err
		}
		if len(field.raw) == 0 {
			continue
		}
		var value string
		if err := json.Unmarshal(field.raw, &value); err != nil {
			return fmt.Errorf("validation: unmarshal ExecutionRecord: %s: %w: %w", field.name, ErrInvalidExecutionRecord, err)
		}
		if result, err = field.set(result, value); err != nil {
			return err
		}
	}
	if err := rejectNull("ExecutionRecord", "correction", raw.Correction, ErrInvalidExecutionRecord); err != nil {
		return err
	}
	if len(raw.Correction) > 0 {
		var correction core.RecordCorrectionRef[core.ValidationExecutionRecordRef]
		if err := json.Unmarshal(raw.Correction, &correction); err != nil {
			return fmt.Errorf("validation: unmarshal ExecutionRecord: %w: %w", ErrInvalidExecutionRecord, err)
		}
		if result, err = result.WithCorrection(correction); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*r = result
	return nil
}

// --- shared helpers ----------------------------------------------------------

// rejectNull reports an error when raw is an explicit JSON null, which every
// optional single-value key in this package rejects rather than silently
// treating as absent. An absent key arrives as a zero-length raw message and
// is permitted.
func rejectNull(typeName, field string, raw json.RawMessage, sentinel error) error {
	if len(raw) > 0 && string(raw) == "null" {
		return fmt.Errorf("validation: unmarshal %s: %w: %s must not be null", typeName, sentinel, field)
	}
	return nil
}

// requireTrimmed trims value and rejects it when empty after trimming.
func requireTrimmed(caller, label, value string, sentinel error) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("validation: %s: %w: %s must not be empty", caller, sentinel, label)
	}
	return trimmed, nil
}

// copyCriteria validates and copies a core.CriterionRef slice, rejecting a
// zero-value element. An empty or nil input returns nil, meaning "none
// declared".
func copyCriteria(caller string, criteria []core.CriterionRef, sentinel error) ([]core.CriterionRef, error) {
	if len(criteria) == 0 {
		return nil, nil
	}
	cp := make([]core.CriterionRef, len(criteria))
	for idx, c := range criteria {
		if c.IsZero() {
			return nil, fmt.Errorf("validation: %s: %w: criterion must not be zero", caller, sentinel)
		}
		cp[idx] = c
	}
	return cp, nil
}

// copyEvidence validates and copies a core.EvidenceArtifactRevisionRef
// slice, rejecting a zero-value element. Evidence citations are never
// deduplicated: PEOS-006 states no uniqueness rule, so recognizing a
// repeated citation is a repository concern, and caller order is preserved.
func copyEvidence(caller string, evidence []core.EvidenceArtifactRevisionRef, sentinel error) ([]core.EvidenceArtifactRevisionRef, error) {
	if len(evidence) == 0 {
		return nil, nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(evidence))
	for idx, e := range evidence {
		if e.IsZero() {
			return nil, fmt.Errorf("validation: %s: %w: evidence citation must not be zero", caller, sentinel)
		}
		cp[idx] = e
	}
	return cp, nil
}

// copyEvidenceOut returns a defensive copy of an evidence slice.
func copyEvidenceOut(evidence []core.EvidenceArtifactRevisionRef) []core.EvidenceArtifactRevisionRef {
	if len(evidence) == 0 {
		return nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(evidence))
	copy(cp, evidence)
	return cp
}
