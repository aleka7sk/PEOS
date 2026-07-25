package lifecycle

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// TransitionOutcome is an open vocabulary wrapper naming the result of a
// Transition Attempt (PEOS-003 "Failed Transition", "Interrupted and
// Indeterminate Transition", and successful completion throughout this
// package's own record content). This package predeclares the same
// four-value minimum vocabulary PEOS-006 uses for the directly analogous
// Validation Execution Record outcome (completed/failed/interrupted/
// indeterminate), renamed here to "succeeded" to match PEOS-003's own
// successful-Transition language. A Product MAY declare additional
// outcome values; see TransitionRecordContent's own validation for how an
// unknown outcome is handled.
type TransitionOutcome struct{ value core.VocabularyValue }

// NewTransitionOutcome wraps v as a TransitionOutcome.
func NewTransitionOutcome(v core.VocabularyValue) TransitionOutcome {
	return TransitionOutcome{value: v}
}

func (o TransitionOutcome) Value() core.VocabularyValue { return o.value }
func (o TransitionOutcome) IsZero() bool                { return o.value.IsZero() }
func (o TransitionOutcome) String() string              { return o.value.String() }

// Equal reports whether o and other name the same outcome value.
func (o TransitionOutcome) Equal(other TransitionOutcome) bool { return o.value.Equal(other.value) }

func (o TransitionOutcome) MarshalJSON() ([]byte, error) { return json.Marshal(o.value) }

func (o *TransitionOutcome) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &o.value)
}

var (
	// TransitionOutcomeSucceeded marks a Transition Attempt that
	// established its target State.
	TransitionOutcomeSucceeded = TransitionOutcome{value: mustVocab("succeeded")}

	// TransitionOutcomeFailed marks a Failed Transition (PEOS-003).
	TransitionOutcomeFailed = TransitionOutcome{value: mustVocab("failed")}

	// TransitionOutcomeInterrupted marks an Interrupted Transition
	// (PEOS-003).
	TransitionOutcomeInterrupted = TransitionOutcome{value: mustVocab("interrupted")}

	// TransitionOutcomeIndeterminate marks an Indeterminate Transition
	// (PEOS-003).
	TransitionOutcomeIndeterminate = TransitionOutcome{value: mustVocab("indeterminate")}
)

// Failure is the smallest immutable value PEOS-003 grounds for a Failed
// Transition: "Failure information SHOULD identify: the failed step; the
// observed reason..." It is not a generic error/exception model. Failure
// carries only the reason PEOS-003 requires and the failed step/phase
// PEOS-003 names as an example, plus an Extension for anything else a
// Product needs.
type Failure struct {
	reason    string
	step      string
	extension core.Extension
}

// NewFailure validates reason and returns a Failure with no step set. Use
// WithStep to add one.
func NewFailure(reason string) (Failure, error) {
	if strings.TrimSpace(reason) == "" {
		return Failure{}, fmt.Errorf("lifecycle: NewFailure: %w: reason must not be empty", ErrInvalidFailure)
	}
	return Failure{reason: reason}, nil
}

// WithStep returns a copy of f with its failed step/phase set.
func (f Failure) WithStep(step string) Failure {
	f.step = step
	return f
}

// WithExtension returns a copy of f with its extension data set.
func (f Failure) WithExtension(e core.Extension) Failure {
	f.extension = e
	return f
}

func (f Failure) Reason() string { return f.reason }

// Step returns f's declared failed step/phase, and whether one is set.
func (f Failure) Step() (string, bool) { return f.step, f.step != "" }

func (f Failure) Extension() core.Extension { return f.extension }

// IsZero reports whether f is the zero value.
func (f Failure) IsZero() bool { return f.reason == "" }

type failureJSON struct {
	Reason    string          `json:"reason"`
	Step      string          `json:"step,omitempty"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes f as {"reason":..., "step":..., ...}, omitting step
// and extension when not set.
func (f Failure) MarshalJSON() ([]byte, error) {
	if f.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal Failure: %w", ErrInvalidFailure)
	}
	raw := failureJSON{Reason: f.reason, Step: f.step}
	if !f.extension.IsZero() {
		raw.Extension = &f.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes f from its JSON form, applying the same
// validation as NewFailure.
func (f *Failure) UnmarshalJSON(data []byte) error {
	var raw failureJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal Failure: %w", err)
	}
	result, err := NewFailure(raw.Reason)
	if err != nil {
		return err
	}
	result = result.WithStep(raw.Step)
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*f = result
	return nil
}

// ArtifactTypeTransitionRecord is the dedicated core.ArtifactType a
// TransitionRecord's underlying core.Artifact MUST declare.
var ArtifactTypeTransitionRecord = core.NewArtifactType(mustVocab("transition-record"))

// TransitionRecord is a Transition Record (PEOS-003): "a persistent
// Artifact that records the attempted or completed application of a
// Transition to a Lifecycle Subject." PEOS-003 explicitly calls this
// construct an Artifact conforming to PEOS-002 -- unlike StateAssignment
// (assignment.go), and unlike PEOS-006/007/008's non-Artifact immutable
// record families. TransitionRecord therefore composes core.Artifact by
// named field, exactly like peos/requirement.Requirement composes
// core.Artifact, rather than introducing a dedicated non-Artifact
// identity type.
type TransitionRecord struct {
	core core.Artifact
}

// NewTransitionRecord validates that artifact is non-zero and declares
// ArtifactTypeTransitionRecord, and returns a TransitionRecord.
func NewTransitionRecord(artifact core.Artifact) (TransitionRecord, error) {
	if artifact.IsZero() {
		return TransitionRecord{}, fmt.Errorf("lifecycle: NewTransitionRecord: %w", ErrInvalidTransitionRecord)
	}
	if artifact.Type() != ArtifactTypeTransitionRecord {
		return TransitionRecord{}, fmt.Errorf("lifecycle: NewTransitionRecord: %w", ErrArtifactTypeMismatch)
	}
	return TransitionRecord{core: artifact}, nil
}

// Core returns the TransitionRecord's underlying core.Artifact.
func (r TransitionRecord) Core() core.Artifact { return r.core }

// ID returns the TransitionRecord's Artifact identity.
func (r TransitionRecord) ID() core.ArtifactID { return r.core.ID() }

// IsZero reports whether r is the zero value.
func (r TransitionRecord) IsZero() bool { return r.core.IsZero() }

type transitionRecordJSON struct {
	Core core.Artifact `json:"core"`
}

// MarshalJSON encodes r as {"core": {...}}, per the nested-composition
// strategy documented on core.ArtifactRevision and already used by
// peos/requirement.
func (r TransitionRecord) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal TransitionRecord: %w", ErrInvalidTransitionRecord)
	}
	return json.Marshal(transitionRecordJSON{Core: r.core})
}

// UnmarshalJSON decodes r from its nested {"core": {...}} JSON form,
// applying the same validation as NewTransitionRecord.
func (r *TransitionRecord) UnmarshalJSON(data []byte) error {
	var raw transitionRecordJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal TransitionRecord: %w", err)
	}
	result, err := NewTransitionRecord(raw.Core)
	if err != nil {
		return err
	}
	*r = result
	return nil
}

// TransitionRecordContent is a Transition Record Revision's typed,
// authoritative content, analogous to peos/requirement.Content. It
// carries only the fields PEOS-003 directly requires for a Transition
// Record, plus the outcome-specific fields §3.9/§3.10 of the Packet E
// Blueprint derived from PEOS-003's own text. Guard evaluation evidence
// and Effect result structure are intentionally not modeled -- see doc.go.
//
// core.Provenance is carried only through the enclosing
// TransitionRecordRevision's own core.ArtifactRevision.Provenance(); this
// type does not duplicate Actor or recorded time. attempted_at and
// completed_at are separate, domain-meaningful times distinct from that
// Provenance's recorded time.
type TransitionRecordContent struct {
	subject           core.LifecycleSubjectRef
	definitionVersion core.LifecycleDefinitionVersionRef
	transition        TransitionID
	fromState         StateID

	hasToState bool
	toState    StateID

	attemptedAt core.Timestamp

	hasCompletedAt bool
	completedAt    core.Timestamp

	outcome TransitionOutcome

	hasFailure bool
	failure    Failure

	hasResultingAssignment bool
	resultingAssignment    core.StateAssignmentRef

	hasAuthority bool
	authority    core.AuthorityRef

	extension core.Extension
}

// NewTransitionRecordContent validates its required arguments and returns
// a TransitionRecordContent with no optional fields set. Use the With*
// methods to add ToState, CompletedAt, Failure, ResultingAssignment,
// Authority, and Extension. The combination of optional fields required
// or forbidden by a given outcome (see validateOutcome) is enforced when
// the content is marshaled, and when it is embedded into a
// TransitionRecordRevision -- not incrementally by each With* call, since
// the applicable rule depends on which outcome is set, which may arrive
// before or after the other fields in a With*-chain.
func NewTransitionRecordContent(
	subject core.LifecycleSubjectRef,
	definitionVersion core.LifecycleDefinitionVersionRef,
	transition TransitionID,
	fromState StateID,
	attemptedAt core.Timestamp,
	outcome TransitionOutcome,
) (TransitionRecordContent, error) {
	if subject.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: NewTransitionRecordContent: %w: subject must not be zero", ErrInvalidTransitionRecordRevision)
	}
	if definitionVersion.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: NewTransitionRecordContent: %w: definition version must not be zero", ErrInvalidTransitionRecordRevision)
	}
	if transition.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: NewTransitionRecordContent: %w", ErrInvalidTransitionID)
	}
	if fromState.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: NewTransitionRecordContent: %w", ErrInvalidStateID)
	}
	if attemptedAt.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: NewTransitionRecordContent: %w: attempted-at must not be zero", ErrInvalidTransitionRecordRevision)
	}
	if outcome.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: NewTransitionRecordContent: %w: outcome must not be zero", ErrInvalidTransitionOutcome)
	}
	return TransitionRecordContent{
		subject: subject, definitionVersion: definitionVersion, transition: transition,
		fromState: fromState, attemptedAt: attemptedAt, outcome: outcome,
	}, nil
}

func (c TransitionRecordContent) WithToState(s StateID) (TransitionRecordContent, error) {
	if s.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: TransitionRecordContent.WithToState: %w", ErrInvalidStateID)
	}
	c.toState, c.hasToState = s, true
	return c, nil
}

func (c TransitionRecordContent) WithCompletedAt(ts core.Timestamp) (TransitionRecordContent, error) {
	if ts.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: TransitionRecordContent.WithCompletedAt: %w: completed-at must not be zero", ErrInvalidTransitionRecordRevision)
	}
	c.completedAt, c.hasCompletedAt = ts, true
	return c, nil
}

func (c TransitionRecordContent) WithFailure(f Failure) (TransitionRecordContent, error) {
	if f.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: TransitionRecordContent.WithFailure: %w", ErrInvalidFailure)
	}
	c.failure, c.hasFailure = f, true
	return c, nil
}

func (c TransitionRecordContent) WithResultingAssignment(ref core.StateAssignmentRef) (TransitionRecordContent, error) {
	if ref.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: TransitionRecordContent.WithResultingAssignment: %w", ErrInvalidStateAssignment)
	}
	c.resultingAssignment, c.hasResultingAssignment = ref, true
	return c, nil
}

func (c TransitionRecordContent) WithAuthority(a core.AuthorityRef) (TransitionRecordContent, error) {
	if a.IsZero() {
		return TransitionRecordContent{}, fmt.Errorf("lifecycle: TransitionRecordContent.WithAuthority: %w: authority must not be zero", ErrInvalidTransitionRecordRevision)
	}
	c.authority, c.hasAuthority = a, true
	return c, nil
}

func (c TransitionRecordContent) WithExtension(e core.Extension) TransitionRecordContent {
	c.extension = e
	return c
}

func (c TransitionRecordContent) Subject() core.LifecycleSubjectRef { return c.subject }
func (c TransitionRecordContent) DefinitionVersion() core.LifecycleDefinitionVersionRef {
	return c.definitionVersion
}
func (c TransitionRecordContent) Transition() TransitionID { return c.transition }
func (c TransitionRecordContent) FromState() StateID       { return c.fromState }

// ToState returns c's resulting target State, and whether one is set.
func (c TransitionRecordContent) ToState() (StateID, bool) { return c.toState, c.hasToState }

func (c TransitionRecordContent) AttemptedAt() core.Timestamp { return c.attemptedAt }

// CompletedAt returns c's completion time, and whether one is set.
func (c TransitionRecordContent) CompletedAt() (core.Timestamp, bool) {
	return c.completedAt, c.hasCompletedAt
}

func (c TransitionRecordContent) Outcome() TransitionOutcome { return c.outcome }

// Failure returns c's failure detail, and whether one is set.
func (c TransitionRecordContent) Failure() (Failure, bool) { return c.failure, c.hasFailure }

// ResultingAssignment returns the State Assignment c's successful
// Transition established, and whether one is set.
func (c TransitionRecordContent) ResultingAssignment() (core.StateAssignmentRef, bool) {
	return c.resultingAssignment, c.hasResultingAssignment
}

// Authority returns c's declared authority, and whether one is set.
func (c TransitionRecordContent) Authority() (core.AuthorityRef, bool) {
	return c.authority, c.hasAuthority
}

func (c TransitionRecordContent) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c TransitionRecordContent) IsZero() bool {
	return c.subject.IsZero() && c.definitionVersion.IsZero() && c.transition.IsZero() &&
		c.fromState.IsZero() && c.attemptedAt.IsZero() && c.outcome.IsZero()
}

// validateOutcome enforces the field combination PEOS-003 implies for
// each of this package's four predeclared outcomes (see doc.go and the
// Packet E Blueprint §3.10). An outcome this package does not predeclare
// is treated generically: only the fields NewTransitionRecordContent
// already requires apply, and no additional combination is enforced --
// this package does not invent universal rules for a Product-defined
// outcome it has no normative basis to constrain.
func (c TransitionRecordContent) validateOutcome() error {
	switch {
	case c.outcome.Equal(TransitionOutcomeSucceeded):
		if !c.hasToState {
			return fmt.Errorf("%w: succeeded outcome requires to_state", ErrInvalidTransitionOutcome)
		}
		if !c.hasCompletedAt {
			return fmt.Errorf("%w: succeeded outcome requires completed_at", ErrInvalidTransitionOutcome)
		}
		if !c.hasResultingAssignment {
			return fmt.Errorf("%w: succeeded outcome requires resulting_state_assignment", ErrInvalidTransitionOutcome)
		}
		if c.hasFailure {
			return fmt.Errorf("%w: succeeded outcome must not carry failure detail", ErrInvalidTransitionOutcome)
		}
	case c.outcome.Equal(TransitionOutcomeFailed):
		if !c.hasCompletedAt {
			return fmt.Errorf("%w: failed outcome requires completed_at", ErrInvalidTransitionOutcome)
		}
		if !c.hasFailure {
			return fmt.Errorf("%w: failed outcome requires failure detail", ErrInvalidTransitionOutcome)
		}
		if c.hasToState {
			return fmt.Errorf("%w: failed outcome must not carry to_state", ErrInvalidTransitionOutcome)
		}
		if c.hasResultingAssignment {
			return fmt.Errorf("%w: failed outcome must not carry resulting_state_assignment", ErrInvalidTransitionOutcome)
		}
	case c.outcome.Equal(TransitionOutcomeInterrupted):
		if !c.hasCompletedAt {
			return fmt.Errorf("%w: interrupted outcome requires completed_at", ErrInvalidTransitionOutcome)
		}
		if c.hasToState {
			return fmt.Errorf("%w: interrupted outcome must not carry to_state", ErrInvalidTransitionOutcome)
		}
		if c.hasResultingAssignment {
			return fmt.Errorf("%w: interrupted outcome must not carry resulting_state_assignment", ErrInvalidTransitionOutcome)
		}
		if c.hasFailure {
			return fmt.Errorf("%w: interrupted outcome must not carry failure detail", ErrInvalidTransitionOutcome)
		}
	case c.outcome.Equal(TransitionOutcomeIndeterminate):
		if c.hasToState {
			return fmt.Errorf("%w: indeterminate outcome must not carry to_state", ErrInvalidTransitionOutcome)
		}
		if c.hasResultingAssignment {
			return fmt.Errorf("%w: indeterminate outcome must not carry resulting_state_assignment", ErrInvalidTransitionOutcome)
		}
	default:
		// Unknown Product-defined outcome: only the generic fields
		// NewTransitionRecordContent already required apply.
	}
	return nil
}

type transitionRecordContentJSON struct {
	Subject             core.LifecycleSubjectRef           `json:"subject"`
	DefinitionVersion   core.LifecycleDefinitionVersionRef `json:"definition_version"`
	Transition          TransitionID                       `json:"transition"`
	FromState           StateID                            `json:"from_state"`
	ToState             *StateID                           `json:"to_state,omitempty"`
	AttemptedAt         core.Timestamp                     `json:"attempted_at"`
	CompletedAt         *core.Timestamp                    `json:"completed_at,omitempty"`
	Outcome             TransitionOutcome                  `json:"outcome"`
	Failure             *Failure                           `json:"failure,omitempty"`
	ResultingAssignment *core.StateAssignmentRef           `json:"resulting_state_assignment,omitempty"`
	Authority           *core.AuthorityRef                 `json:"authority,omitempty"`
	Extension           *core.Extension                    `json:"extension,omitempty"`
}

// transitionRecordContentUnmarshalJSON mirrors transitionRecordContentJSON
// for decoding only: ToState, CompletedAt, Failure, ResultingAssignment,
// and Authority are each captured as raw, undecoded bytes so an explicit
// JSON null can be distinguished from an absent key and rejected for each
// -- the same technique Packet D.1 established for Relation.Scope.
type transitionRecordContentUnmarshalJSON struct {
	Subject             core.LifecycleSubjectRef           `json:"subject"`
	DefinitionVersion   core.LifecycleDefinitionVersionRef `json:"definition_version"`
	Transition          TransitionID                       `json:"transition"`
	FromState           StateID                            `json:"from_state"`
	ToState             json.RawMessage                    `json:"to_state"`
	AttemptedAt         core.Timestamp                     `json:"attempted_at"`
	CompletedAt         json.RawMessage                    `json:"completed_at"`
	Outcome             TransitionOutcome                  `json:"outcome"`
	Failure             json.RawMessage                    `json:"failure"`
	ResultingAssignment json.RawMessage                    `json:"resulting_state_assignment"`
	Authority           json.RawMessage                    `json:"authority"`
	Extension           *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"subject":..., "outcome":..., ...}, omitting
// every optional field that is not set, and rejects a zero-value or
// outcome-inconsistent c (see validateOutcome).
func (c TransitionRecordContent) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal TransitionRecordContent: %w", ErrInvalidTransitionRecordRevision)
	}
	if err := c.validateOutcome(); err != nil {
		return nil, fmt.Errorf("lifecycle: marshal TransitionRecordContent: %w", err)
	}
	raw := transitionRecordContentJSON{
		Subject: c.subject, DefinitionVersion: c.definitionVersion, Transition: c.transition,
		FromState: c.fromState, AttemptedAt: c.attemptedAt, Outcome: c.outcome,
	}
	if c.hasToState {
		raw.ToState = &c.toState
	}
	if c.hasCompletedAt {
		raw.CompletedAt = &c.completedAt
	}
	if c.hasFailure {
		raw.Failure = &c.failure
	}
	if c.hasResultingAssignment {
		raw.ResultingAssignment = &c.resultingAssignment
	}
	if c.hasAuthority {
		raw.Authority = &c.authority
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, applying the same
// validation as NewTransitionRecordContent, the With* methods, and
// validateOutcome. An explicit JSON null for to_state, completed_at,
// failure, resulting_state_assignment, or authority is rejected rather
// than silently treated as absent.
func (c *TransitionRecordContent) UnmarshalJSON(data []byte) error {
	var raw transitionRecordContentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
	}
	result, err := NewTransitionRecordContent(raw.Subject, raw.DefinitionVersion, raw.Transition, raw.FromState, raw.AttemptedAt, raw.Outcome)
	if err != nil {
		return err
	}

	if len(raw.ToState) > 0 {
		if string(raw.ToState) == "null" {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w: to_state must not be null", ErrInvalidStateID)
		}
		var s StateID
		if err := json.Unmarshal(raw.ToState, &s); err != nil {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
		}
		if result, err = result.WithToState(s); err != nil {
			return err
		}
	}
	if len(raw.CompletedAt) > 0 {
		if string(raw.CompletedAt) == "null" {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w: completed_at must not be null", ErrInvalidTransitionRecordRevision)
		}
		var ts core.Timestamp
		if err := json.Unmarshal(raw.CompletedAt, &ts); err != nil {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
		}
		if result, err = result.WithCompletedAt(ts); err != nil {
			return err
		}
	}
	if len(raw.Failure) > 0 {
		if string(raw.Failure) == "null" {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w: failure must not be null", ErrInvalidFailure)
		}
		var f Failure
		if err := json.Unmarshal(raw.Failure, &f); err != nil {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
		}
		if result, err = result.WithFailure(f); err != nil {
			return err
		}
	}
	if len(raw.ResultingAssignment) > 0 {
		if string(raw.ResultingAssignment) == "null" {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w: resulting_state_assignment must not be null", ErrInvalidStateAssignment)
		}
		var ref core.StateAssignmentRef
		if err := json.Unmarshal(raw.ResultingAssignment, &ref); err != nil {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
		}
		if result, err = result.WithResultingAssignment(ref); err != nil {
			return err
		}
	}
	if len(raw.Authority) > 0 {
		if string(raw.Authority) == "null" {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w: authority must not be null", ErrInvalidTransitionRecordRevision)
		}
		var a core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &a); err != nil {
			return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
		}
		if result, err = result.WithAuthority(a); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	if err := result.validateOutcome(); err != nil {
		return fmt.Errorf("lifecycle: unmarshal TransitionRecordContent: %w", err)
	}
	*c = result
	return nil
}

// TransitionRecordRevision is a Transition Record Revision: shorthand for
// "an Artifact Revision whose Artifact is a Transition Record," composing
// core.ArtifactRevision by named field and pairing it with typed
// TransitionRecordContent, exactly like peos/requirement.Revision.
type TransitionRecordRevision struct {
	core    core.ArtifactRevision
	content TransitionRecordContent
}

// newTransitionRecordRevisionFromParts validates revision and content
// without reference to any TransitionRecord, and is the path both
// NewTransitionRecordRevision and UnmarshalJSON share. It cannot, and does
// not attempt to, check that revision belongs to any particular
// TransitionRecord -- see NewTransitionRecordRevision and UnmarshalJSON's
// own documentation for why that check requires a TransitionRecord value
// a Revision's own JSON does not carry (the same limitation
// peos/requirement.Revision documents for itself).
func newTransitionRecordRevisionFromParts(revision core.ArtifactRevision, content TransitionRecordContent) (TransitionRecordRevision, error) {
	if revision.IsZero() {
		return TransitionRecordRevision{}, fmt.Errorf("lifecycle: %w: core revision must not be zero", ErrInvalidTransitionRecordRevision)
	}
	if content.IsZero() {
		return TransitionRecordRevision{}, fmt.Errorf("lifecycle: %w: content must not be zero", ErrInvalidTransitionRecordRevision)
	}
	if err := content.validateOutcome(); err != nil {
		return TransitionRecordRevision{}, fmt.Errorf("lifecycle: %w", err)
	}
	return TransitionRecordRevision{core: revision, content: content}, nil
}

// NewTransitionRecordRevision validates record, revision, and content and
// returns a TransitionRecordRevision. record and revision must both be
// non-zero, content must be non-zero and outcome-consistent, and
// revision.ArtifactID() must equal record.ID().
func NewTransitionRecordRevision(record TransitionRecord, revision core.ArtifactRevision, content TransitionRecordContent) (TransitionRecordRevision, error) {
	if record.IsZero() {
		return TransitionRecordRevision{}, fmt.Errorf("lifecycle: NewTransitionRecordRevision: %w: transition record must not be zero", ErrInvalidTransitionRecordRevision)
	}
	result, err := newTransitionRecordRevisionFromParts(revision, content)
	if err != nil {
		return TransitionRecordRevision{}, err
	}
	if revision.ArtifactID() != record.ID() {
		return TransitionRecordRevision{}, fmt.Errorf("lifecycle: NewTransitionRecordRevision: %w", ErrArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the Revision's underlying core.ArtifactRevision.
func (r TransitionRecordRevision) Core() core.ArtifactRevision { return r.core }

// Content returns the Revision's typed Transition Record content.
func (r TransitionRecordRevision) Content() TransitionRecordContent { return r.content }

// IsZero reports whether r is the zero value.
func (r TransitionRecordRevision) IsZero() bool { return r.core.IsZero() && r.content.IsZero() }

// Ref returns a core.ArtifactRevisionRef identifying r, suitable for use
// as a StateAssignment's established_by reference.
func (r TransitionRecordRevision) Ref() (core.ArtifactRevisionRef, error) {
	return core.NewArtifactRevisionRef(r.core.ArtifactID(), r.core.RevisionID())
}

type transitionRecordRevisionJSON struct {
	Core    core.ArtifactRevision   `json:"core"`
	Content TransitionRecordContent `json:"content"`
}

// MarshalJSON encodes r as {"core": {...}, "content": {...}}, per the
// nested-composition strategy documented on core.ArtifactRevision.
func (r TransitionRecordRevision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal TransitionRecordRevision: %w", ErrInvalidTransitionRecordRevision)
	}
	return json.Marshal(transitionRecordRevisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core": {...}, "content":
// {...}} JSON form.
func (r *TransitionRecordRevision) UnmarshalJSON(data []byte) error {
	var raw transitionRecordRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal TransitionRecordRevision: %w", err)
	}
	result, err := newTransitionRecordRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return err
	}
	*r = result
	return nil
}
