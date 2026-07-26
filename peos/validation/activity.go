package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// PlannedActivity is a PEOS-006 Planned Validation Activity: "an Artifact
// Revision-owned value structure" belonging to exactly one Validation Plan
// Revision.
//
// PlannedActivity has no independent PEOS identity, no revisions, and no
// lifecycle of its own: "its evolution is entirely governed by the revision
// of its owning Validation Plan." Structurally, it carries no ID type, no
// Ref type, no core.Artifact, no core.ArtifactRevision, and no provenance
// field -- provenance belongs to the owning PlanContent, which records the
// origin of the Plan Revision as a whole. Non-conforming patterns
// "Decision-Like Validation Activity", "Planned Activity with Global
// Identity", and "Lifecycle-Governed Validation Activity" are prevented
// this way rather than by convention.
//
// Its key is a core.LocalKey and is stable only within its owning
// Validation Plan Revision: it "does not survive as an independent identity
// outside that exact Plan Revision", "is not an Artifact Identity", and "is
// not a global Validation Activity Identity". A later Plan Revision MAY
// reuse, remove, or reintroduce a key, and a key from an earlier Plan
// Revision SHALL NOT be assumed to refer to the same Planned Validation
// Activity in a later one unless an applicable Product contract explicitly
// defines that continuity. This package never compares keys across Plan
// Revisions.
//
// A PlannedActivity has exactly one Subject, at an exact participant level.
// Its criteria are the evaluation rules the Subject is evaluated against;
// they are never additional Subjects. That separation is enforced by the
// type system, not by a check: core.CriterionRef and
// core.EngineeringSubjectRef are distinct types with no conversion path in
// either direction, so a criterion cannot be placed in the subject field.
//
// The four fields NewPlannedActivity requires are the ones PEOS-006 states
// unconditionally for every Planned Validation Activity and whose zero
// value would be ambiguous. The optional collections (criteria, expected
// evidence, prerequisites, dependencies) are also listed unconditionally,
// but an empty collection is already an unambiguous "none declared" --
// there is no third, unstated state to guard against -- and PEOS-006
// contemplates zero-criteria evaluation elsewhere ("Where zero criteria are
// identified, the Claim's outcome SHALL be interpreted strictly according
// to its stated Validation Method and basis"). Sequencing/dependencies,
// responsible actor or role, and required authority are each explicitly
// qualified ("where applicable", "where required") and are optional for
// that reason.
//
// A PlannedActivity may be constructed and held before it is attached to
// any PlanContent. Dependency resolution -- checking that every declared
// dependency names an Activity in the same Plan Revision -- is therefore
// performed by NewPlanContent, not here. This type provides no graph
// traversal and no cycle detection.
type PlannedActivity struct {
	key                   core.LocalKey
	subject               core.EngineeringSubjectRef
	method                core.ValidationMethod
	outcomeInterpretation string

	criteria          []core.CriterionRef
	expectedEvidence  []string
	prerequisites     []string
	dependencies      []core.LocalKey
	responsibleRole   core.VocabularyValue
	requiredAuthority core.AuthorityRef
	methodDefinition  core.ArtifactRevisionRef
	extension         core.Extension
}

// NewPlannedActivity validates key, subject, method, and
// outcomeInterpretation and returns a PlannedActivity with no criteria,
// expected evidence, prerequisites, dependencies, responsible role,
// required authority, method definition, or extension data. Use the With*
// methods to add those.
//
// All four arguments are mandatory. key, subject, and method must each be
// non-zero. outcomeInterpretation must be non-empty after trimming
// surrounding whitespace; the trimmed value is stored (PEOS-006: an
// Activity SHALL identify "how its expected outcome is to be
// interpreted").
//
// A successful call always returns a fully valid PlannedActivity: no
// mandatory field is supplied by a later With* call, so an Activity can
// never exist in a partially established state.
func NewPlannedActivity(
	key core.LocalKey,
	subject core.EngineeringSubjectRef,
	method core.ValidationMethod,
	outcomeInterpretation string,
) (PlannedActivity, error) {
	if key.IsZero() {
		return PlannedActivity{}, fmt.Errorf("validation: NewPlannedActivity: %w: plan-local key must not be zero", ErrInvalidPlannedActivity)
	}
	if subject.IsZero() {
		return PlannedActivity{}, fmt.Errorf("validation: NewPlannedActivity: %w: subject must not be zero", ErrInvalidPlannedActivity)
	}
	if method.IsZero() {
		return PlannedActivity{}, fmt.Errorf("validation: NewPlannedActivity: %w: validation method must not be zero", ErrInvalidPlannedActivity)
	}
	trimmed := strings.TrimSpace(outcomeInterpretation)
	if trimmed == "" {
		return PlannedActivity{}, fmt.Errorf("validation: NewPlannedActivity: %w: outcome interpretation must not be empty", ErrInvalidPlannedActivity)
	}
	return PlannedActivity{
		key:                   key,
		subject:               subject,
		method:                method,
		outcomeInterpretation: trimmed,
	}, nil
}

// copyLocalKeys validates and copies a core.LocalKey slice.
func copyLocalKeys(caller string, keys []core.LocalKey) ([]core.LocalKey, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	cp := make([]core.LocalKey, len(keys))
	for idx, k := range keys {
		if k.IsZero() {
			return nil, fmt.Errorf("validation: %s: %w: plan-local key must not be zero", caller, ErrInvalidPlannedActivity)
		}
		cp[idx] = k
	}
	return cp, nil
}

// copyTrimmedStrings validates and copies a string slice, trimming each
// element and rejecting any that is empty after trimming.
func copyTrimmedStrings(caller, label string, values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	cp := make([]string, len(values))
	for idx, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, fmt.Errorf("validation: %s: %w: %s entry must not be empty", caller, ErrInvalidPlannedActivity, label)
		}
		cp[idx] = trimmed
	}
	return cp, nil
}

// WithCriteria returns a copy of a with its evaluation criteria set to
// exactly the values given, in the order given, replacing any previous
// criteria. A zero-value core.CriterionRef among criteria is rejected.
// Passing an empty or nil slice declares zero criteria, which PEOS-006
// permits for a Planned Validation Activity -- this is why no
// WithoutCriteria method exists: WithCriteria(nil) already expresses
// removal, and a second method would create a second validation path for
// the same field.
//
// Criteria are not deduplicated. PEOS-006 states no uniqueness rule for an
// Activity's criteria, so recognizing a repeated criterion is a
// repository- or Product-owned concern; caller order is preserved exactly.
func (a PlannedActivity) WithCriteria(criteria []core.CriterionRef) (PlannedActivity, error) {
	if len(criteria) == 0 {
		a.criteria = nil
		return a, nil
	}
	cp := make([]core.CriterionRef, len(criteria))
	for idx, c := range criteria {
		if c.IsZero() {
			return PlannedActivity{}, fmt.Errorf("validation: PlannedActivity.WithCriteria: %w: criterion must not be zero", ErrInvalidPlannedActivity)
		}
		cp[idx] = c
	}
	a.criteria = cp
	return a, nil
}

// WithExpectedEvidence returns a copy of a with its expected-evidence
// descriptions set to exactly the values given, in the order given. Each
// entry is trimmed and must be non-empty after trimming. Passing an empty
// or nil slice declares none.
//
// These are descriptions of the Evidence an execution is expected to
// produce or rely upon, not citations: at planning time that Evidence does
// not yet exist, so there is no Artifact Revision to reference. A
// Validation Execution Record cites the Evidence that actually
// materialized, using core.EvidenceArtifactRevisionRef (Packet H.2).
func (a PlannedActivity) WithExpectedEvidence(descriptions []string) (PlannedActivity, error) {
	cp, err := copyTrimmedStrings("PlannedActivity.WithExpectedEvidence", "expected evidence", descriptions)
	if err != nil {
		return PlannedActivity{}, err
	}
	a.expectedEvidence = cp
	return a, nil
}

// WithPrerequisites returns a copy of a with its execution prerequisites
// set to exactly the values given, in the order given. Each entry is
// trimmed and must be non-empty after trimming. Passing an empty or nil
// slice declares none.
func (a PlannedActivity) WithPrerequisites(prerequisites []string) (PlannedActivity, error) {
	cp, err := copyTrimmedStrings("PlannedActivity.WithPrerequisites", "prerequisite", prerequisites)
	if err != nil {
		return PlannedActivity{}, err
	}
	a.prerequisites = cp
	return a, nil
}

// WithDependencies returns a copy of a with its declared dependencies set
// to exactly the plan-local keys given, in the order given. A zero-value
// core.LocalKey among dependencies is rejected. Passing an empty or nil
// slice declares none.
//
// Whether each key actually resolves to another Activity is checked by
// NewPlanContent, which is the only place the full key set of a Plan
// Revision is known. PEOS-006 states no cycle policy and no
// self-reference prohibition for Activity dependencies, so neither is
// rejected here or there.
func (a PlannedActivity) WithDependencies(dependencies []core.LocalKey) (PlannedActivity, error) {
	cp, err := copyLocalKeys("PlannedActivity.WithDependencies", dependencies)
	if err != nil {
		return PlannedActivity{}, err
	}
	a.dependencies = cp
	return a, nil
}

// WithResponsibleRole returns a copy of a with its responsible actor or
// role set. role must be non-zero; use WithoutResponsibleRole to clear it.
// PEOS-006 requires this only "where required", so it is optional.
func (a PlannedActivity) WithResponsibleRole(role core.VocabularyValue) (PlannedActivity, error) {
	if role.IsZero() {
		return PlannedActivity{}, fmt.Errorf("validation: PlannedActivity.WithResponsibleRole: %w: responsible role must not be zero", ErrInvalidPlannedActivity)
	}
	a.responsibleRole = role
	return a, nil
}

// WithoutResponsibleRole returns a copy of a with its responsible role
// cleared.
func (a PlannedActivity) WithoutResponsibleRole() PlannedActivity {
	a.responsibleRole = core.VocabularyValue{}
	return a
}

// WithRequiredAuthority returns a copy of a with the authority required to
// establish its result as applicable set. authority must be non-zero; use
// WithoutRequiredAuthority to clear it. PEOS-006 requires this only "where
// required", so it is optional.
func (a PlannedActivity) WithRequiredAuthority(authority core.AuthorityRef) (PlannedActivity, error) {
	if authority.IsZero() {
		return PlannedActivity{}, fmt.Errorf("validation: PlannedActivity.WithRequiredAuthority: %w: required authority must not be zero", ErrInvalidPlannedActivity)
	}
	a.requiredAuthority = authority
	return a, nil
}

// WithoutRequiredAuthority returns a copy of a with its required authority
// cleared.
func (a PlannedActivity) WithoutRequiredAuthority() PlannedActivity {
	a.requiredAuthority = core.AuthorityRef{}
	return a
}

// WithMethodDefinition returns a copy of a with an optional reference to
// the exact Artifact Revision defining its Validation Method's detailed
// procedure content. ref must be non-zero; use WithoutMethodDefinition to
// clear it.
//
// PEOS-006 permits, but does not require, an Artifact to define a Method's
// procedure: "An external or PEOS Artifact MAY define detailed procedure
// content for a Validation Method. This specification does not require
// every Method Type to be its own Artifact." This reference never becomes
// the Method's identity, and it introduces no Method revision system --
// the backing Artifact's own Artifact Revision is the only revisioning
// involved (non-conforming pattern "Validation Method Version"). The Method
// itself remains the core.ValidationMethod vocabulary value.
func (a PlannedActivity) WithMethodDefinition(ref core.ArtifactRevisionRef) (PlannedActivity, error) {
	if ref.IsZero() {
		return PlannedActivity{}, fmt.Errorf("validation: PlannedActivity.WithMethodDefinition: %w: method definition must not be zero", ErrInvalidPlannedActivity)
	}
	a.methodDefinition = ref
	return a, nil
}

// WithoutMethodDefinition returns a copy of a with its method-definition
// reference cleared.
func (a PlannedActivity) WithoutMethodDefinition() PlannedActivity {
	a.methodDefinition = core.ArtifactRevisionRef{}
	return a
}

// WithExtension returns a copy of a with its extension data set. Passing
// the zero core.Extension is equivalent to declaring none, per
// core.Extension's own documented contract.
//
// Extension is where Product-specific Validation Method parameters belong.
// PEOS-006 defines no method parameter model, and this package introduces
// none: doing so would amount to a method DSL the specification does not
// define.
func (a PlannedActivity) WithExtension(extension core.Extension) PlannedActivity {
	a.extension = extension
	return a
}

// WithoutExtension returns a copy of a with its extension data cleared.
func (a PlannedActivity) WithoutExtension() PlannedActivity {
	a.extension = core.Extension{}
	return a
}

// Key returns a's plan-local key. It is meaningful only within a's owning
// Validation Plan Revision.
func (a PlannedActivity) Key() core.LocalKey { return a.key }

// Subject returns a's single Validation Subject. Its participant level is
// carried by the returned core.EngineeringSubjectRef's own arm, so a valid
// Activity always identifies an exact level.
func (a PlannedActivity) Subject() core.EngineeringSubjectRef { return a.subject }

// Method returns a's declared Validation Method.
func (a PlannedActivity) Method() core.ValidationMethod { return a.method }

// OutcomeInterpretation returns a's declared statement of how its expected
// outcome is to be interpreted.
func (a PlannedActivity) OutcomeInterpretation() string { return a.outcomeInterpretation }

// Criteria returns a defensive copy of a's evaluation criteria, in
// declaration order.
func (a PlannedActivity) Criteria() []core.CriterionRef {
	if len(a.criteria) == 0 {
		return nil
	}
	cp := make([]core.CriterionRef, len(a.criteria))
	copy(cp, a.criteria)
	return cp
}

// ExpectedEvidence returns a defensive copy of a's expected-evidence
// descriptions, in declaration order.
func (a PlannedActivity) ExpectedEvidence() []string {
	if len(a.expectedEvidence) == 0 {
		return nil
	}
	cp := make([]string, len(a.expectedEvidence))
	copy(cp, a.expectedEvidence)
	return cp
}

// Prerequisites returns a defensive copy of a's execution prerequisites,
// in declaration order.
func (a PlannedActivity) Prerequisites() []string {
	if len(a.prerequisites) == 0 {
		return nil
	}
	cp := make([]string, len(a.prerequisites))
	copy(cp, a.prerequisites)
	return cp
}

// Dependencies returns a defensive copy of a's declared dependency keys,
// in declaration order.
func (a PlannedActivity) Dependencies() []core.LocalKey {
	if len(a.dependencies) == 0 {
		return nil
	}
	cp := make([]core.LocalKey, len(a.dependencies))
	copy(cp, a.dependencies)
	return cp
}

// ResponsibleRole returns a's declared responsible actor or role, and
// whether one is set.
func (a PlannedActivity) ResponsibleRole() (core.VocabularyValue, bool) {
	return a.responsibleRole, !a.responsibleRole.IsZero()
}

// RequiredAuthority returns the authority a declares as required to
// establish its result as applicable, and whether one is set.
func (a PlannedActivity) RequiredAuthority() (core.AuthorityRef, bool) {
	return a.requiredAuthority, !a.requiredAuthority.IsZero()
}

// MethodDefinition returns a's optional method-definition Artifact
// Revision reference, and whether one is set.
func (a PlannedActivity) MethodDefinition() (core.ArtifactRevisionRef, bool) {
	return a.methodDefinition, !a.methodDefinition.IsZero()
}

// Extension returns a's extension data.
func (a PlannedActivity) Extension() core.Extension { return a.extension }

// IsZero reports whether a is the zero value.
func (a PlannedActivity) IsZero() bool {
	return a.key.IsZero() && a.subject.IsZero() && a.method.IsZero() && a.outcomeInterpretation == ""
}

type plannedActivityJSON struct {
	Key                   core.LocalKey              `json:"key"`
	Subject               core.EngineeringSubjectRef `json:"subject"`
	Method                core.ValidationMethod      `json:"method"`
	OutcomeInterpretation string                     `json:"outcome_interpretation"`
	Criteria              []core.CriterionRef        `json:"criteria,omitempty"`
	ExpectedEvidence      []string                   `json:"expected_evidence,omitempty"`
	Prerequisites         []string                   `json:"prerequisites,omitempty"`
	Dependencies          []core.LocalKey            `json:"dependencies,omitempty"`
	ResponsibleRole       *core.VocabularyValue      `json:"responsible_role,omitempty"`
	RequiredAuthority     *core.AuthorityRef         `json:"required_authority,omitempty"`
	MethodDefinition      *core.ArtifactRevisionRef  `json:"method_definition,omitempty"`
	Extension             *core.Extension            `json:"extension,omitempty"`
}

// plannedActivityUnmarshalJSON mirrors plannedActivityJSON's field set for
// decoding only, with three differences: ResponsibleRole,
// RequiredAuthority, and MethodDefinition are captured as raw, undecoded
// bytes, so an explicit JSON null can be distinguished from an absent key
// and rejected -- the technique Packet D.1 established for
// relation.Relation's Scope and lifecycle.StateAssignment's Authority.
//
// The optional collections deliberately do not use this treatment: for
// criteria, expected_evidence, prerequisites, and dependencies, an absent
// key, an explicit null, and an empty array all denote the same valid
// state -- "none declared" -- because PEOS-006 permits zero cardinality
// for each. Distinguishing them would create a difference with no
// semantic content. This is not the case for a Validation Claim's
// criteria, where a Claim Type may forbid the empty case and the three
// inputs therefore need to be told apart; that is Packet H.2's concern.
type plannedActivityUnmarshalJSON struct {
	Key                   core.LocalKey              `json:"key"`
	Subject               core.EngineeringSubjectRef `json:"subject"`
	Method                core.ValidationMethod      `json:"method"`
	OutcomeInterpretation string                     `json:"outcome_interpretation"`
	Criteria              []core.CriterionRef        `json:"criteria"`
	ExpectedEvidence      []string                   `json:"expected_evidence"`
	Prerequisites         []string                   `json:"prerequisites"`
	Dependencies          []core.LocalKey            `json:"dependencies"`
	ResponsibleRole       json.RawMessage            `json:"responsible_role"`
	RequiredAuthority     json.RawMessage            `json:"required_authority"`
	MethodDefinition      json.RawMessage            `json:"method_definition"`
	Extension             *core.Extension            `json:"extension,omitempty"`
}

// MarshalJSON encodes a as {"key":..., "subject":..., "method":...,
// "outcome_interpretation":...}, plus whichever optional keys are set. key,
// subject, method, and outcome_interpretation are never omitted.
//
// There is no "relation" key, no lifecycle key, no "id" or revision
// identity of any kind, and no top-level type discriminator -- their
// absence is the structural proof that a Planned Validation Activity is
// neither a relationship nor an independently identified entity.
func (a PlannedActivity) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return nil, fmt.Errorf("validation: marshal PlannedActivity: %w", ErrInvalidPlannedActivity)
	}
	raw := plannedActivityJSON{
		Key:                   a.key,
		Subject:               a.subject,
		Method:                a.method,
		OutcomeInterpretation: a.outcomeInterpretation,
	}
	if len(a.criteria) > 0 {
		raw.Criteria = a.criteria
	}
	if len(a.expectedEvidence) > 0 {
		raw.ExpectedEvidence = a.expectedEvidence
	}
	if len(a.prerequisites) > 0 {
		raw.Prerequisites = a.prerequisites
	}
	if len(a.dependencies) > 0 {
		raw.Dependencies = a.dependencies
	}
	if !a.responsibleRole.IsZero() {
		raw.ResponsibleRole = &a.responsibleRole
	}
	if !a.requiredAuthority.IsZero() {
		raw.RequiredAuthority = &a.requiredAuthority
	}
	if !a.methodDefinition.IsZero() {
		raw.MethodDefinition = &a.methodDefinition
	}
	if !a.extension.IsZero() {
		raw.Extension = &a.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes a from its JSON form, applying the same validation
// as NewPlannedActivity and each With* method, so a decoded
// PlannedActivity can never be constructor-impossible.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - key, subject, method: a missing key leaves the field zero and
//     reaches NewPlannedActivity, which rejects it with
//     ErrInvalidPlannedActivity. An explicit null instead invokes that
//     nested type's own UnmarshalJSON, which fails there, so the error is
//     wrapped here with ErrInvalidPlannedActivity plus that type's own
//     sentinel (for example core.ErrEmptyIdentity for key). Both are
//     rejected; the sentinel sets differ.
//   - outcome_interpretation: both a missing key and an explicit null
//     yield the empty string, rejected with ErrInvalidPlannedActivity.
//   - criteria, expected_evidence, prerequisites, dependencies: absent,
//     explicit null, and empty array are equivalent and all mean "none
//     declared" (see plannedActivityUnmarshalJSON's own comment).
//   - responsible_role, required_authority, method_definition: a missing
//     key means absent; an explicit null is rejected rather than silently
//     treated as absent.
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract.
func (a *PlannedActivity) UnmarshalJSON(data []byte) error {
	var raw plannedActivityUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal PlannedActivity: %w: %w", ErrInvalidPlannedActivity, err)
	}

	result, err := NewPlannedActivity(raw.Key, raw.Subject, raw.Method, raw.OutcomeInterpretation)
	if err != nil {
		return err
	}
	if len(raw.Criteria) > 0 {
		if result, err = result.WithCriteria(raw.Criteria); err != nil {
			return err
		}
	}
	if len(raw.ExpectedEvidence) > 0 {
		if result, err = result.WithExpectedEvidence(raw.ExpectedEvidence); err != nil {
			return err
		}
	}
	if len(raw.Prerequisites) > 0 {
		if result, err = result.WithPrerequisites(raw.Prerequisites); err != nil {
			return err
		}
	}
	if len(raw.Dependencies) > 0 {
		if result, err = result.WithDependencies(raw.Dependencies); err != nil {
			return err
		}
	}
	if len(raw.ResponsibleRole) > 0 {
		if string(raw.ResponsibleRole) == "null" {
			return fmt.Errorf("validation: unmarshal PlannedActivity: %w: responsible role must not be null", ErrInvalidPlannedActivity)
		}
		var role core.VocabularyValue
		if err := json.Unmarshal(raw.ResponsibleRole, &role); err != nil {
			return fmt.Errorf("validation: unmarshal PlannedActivity: %w: %w", ErrInvalidPlannedActivity, err)
		}
		if result, err = result.WithResponsibleRole(role); err != nil {
			return err
		}
	}
	if len(raw.RequiredAuthority) > 0 {
		if string(raw.RequiredAuthority) == "null" {
			return fmt.Errorf("validation: unmarshal PlannedActivity: %w: required authority must not be null", ErrInvalidPlannedActivity)
		}
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.RequiredAuthority, &authority); err != nil {
			return fmt.Errorf("validation: unmarshal PlannedActivity: %w: %w", ErrInvalidPlannedActivity, err)
		}
		if result, err = result.WithRequiredAuthority(authority); err != nil {
			return err
		}
	}
	if len(raw.MethodDefinition) > 0 {
		if string(raw.MethodDefinition) == "null" {
			return fmt.Errorf("validation: unmarshal PlannedActivity: %w: method definition must not be null", ErrInvalidPlannedActivity)
		}
		var ref core.ArtifactRevisionRef
		if err := json.Unmarshal(raw.MethodDefinition, &ref); err != nil {
			return fmt.Errorf("validation: unmarshal PlannedActivity: %w: %w", ErrInvalidPlannedActivity, err)
		}
		if result, err = result.WithMethodDefinition(ref); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*a = result
	return nil
}
