package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustArtifactTypeValidationPlan() core.ArtifactType {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, "validation-plan")
	if err != nil {
		panic(err)
	}
	return core.NewArtifactType(v)
}

// ArtifactTypeValidationPlan is the PEOS-006 Validation Plan Artifact
// Type. PEOS-006 does not itself fix an exact vocabulary string for this
// value -- this is an implementation choice, namespaced under
// core.PEOSNamespace because Validation Plan is a PEOS-000-009-defined
// Artifact Type rather than a Product-specific one, matching the
// convention requirement.ArtifactTypeRequirement already established.
// core.ArtifactType's own vocabulary remains fully open.
//
// This value deliberately lives in peos/validation, not peos/core, for the
// same reason ArtifactTypeRequirement lives in peos/requirement: the
// Artifact Type belongs to the specialization that defines it.
var ArtifactTypeValidationPlan = mustArtifactTypeValidationPlan()

// --- Plan --------------------------------------------------------------------

// Plan is a PEOS-006 Validation Plan identity: a core.Artifact whose
// declared Artifact Type is ArtifactTypeValidationPlan ("A Validation Plan
// is an Artifact as defined by PEOS-002").
//
// Plan adds no field of its own. Every Validation Plan content element --
// scope, applicability, provenance, and the Planned Validation Activities
// themselves -- is Revision-owned content carried by PlanContent, never
// Plan identity. Plan therefore has no Version field of any kind: "A
// Validation Plan uses ordinary Artifact Revision for all of its
// evolution. There is no Validation Plan Version distinct from Artifact
// Revision" (non-conforming pattern "Validation Plan Version").
//
// Plan carries no Lifecycle State and no extension of its own. A
// Validation Plan is an ordinary Artifact and MAY be governed by a
// Lifecycle under PEOS-003; that lifecycle is modeled exclusively in
// peos/lifecycle, which this package does not import. Product-specific
// data belongs on the underlying core.Artifact's own extension, or on
// PlanContent.
type Plan struct {
	core core.Artifact
}

// NewPlan validates artifact and returns a Plan. artifact must be non-zero
// and its Type() must equal ArtifactTypeValidationPlan.
func NewPlan(artifact core.Artifact) (Plan, error) {
	if artifact.IsZero() {
		return Plan{}, fmt.Errorf("validation: NewPlan: %w", ErrInvalidValidationPlan)
	}
	if artifact.Type() != ArtifactTypeValidationPlan {
		return Plan{}, fmt.Errorf("validation: NewPlan: %w", ErrValidationPlanArtifactTypeMismatch)
	}
	return Plan{core: artifact}, nil
}

// Core returns the Plan's underlying core.Artifact.
func (p Plan) Core() core.Artifact { return p.core }

// ID returns the Plan's identity.
func (p Plan) ID() core.ArtifactID { return p.core.ID() }

// Ref returns a core.ValidationPlanRef identifying p.
func (p Plan) Ref() (core.ValidationPlanRef, error) {
	return core.NewValidationPlanRef(p.core.ID())
}

// IsZero reports whether p is the zero value.
func (p Plan) IsZero() bool { return p.core.IsZero() }

// MarshalJSON encodes p as the wire form of its underlying core.Artifact,
// with no additional envelope -- the same strategy requirement.Requirement
// uses. core.Artifact's own JSON already carries artifact_type, which both
// preserves and (on Unmarshal) lets NewPlan re-verify that the decoded
// value is a Validation Plan.
func (p Plan) MarshalJSON() ([]byte, error) {
	if p.IsZero() {
		return nil, fmt.Errorf("validation: marshal Plan: %w", ErrInvalidValidationPlan)
	}
	return json.Marshal(p.core)
}

// UnmarshalJSON decodes p from its JSON form, applying the same validation
// as NewPlan. An explicit JSON null decodes core.Artifact to its zero
// value, which NewPlan then rejects with ErrInvalidValidationPlan; a
// decoded Plan can never be constructor-impossible.
func (p *Plan) UnmarshalJSON(data []byte) error {
	var artifact core.Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return fmt.Errorf("validation: unmarshal Plan: %w: %w", ErrInvalidValidationPlan, err)
	}
	result, err := NewPlan(artifact)
	if err != nil {
		return err
	}
	*p = result
	return nil
}

// --- PlanApplicability -------------------------------------------------------

type planApplicabilityKind string

const (
	planApplicabilityKindUnrestricted planApplicabilityKind = "unrestricted"
	planApplicabilityKindScoped       planApplicabilityKind = "scoped"
)

// PlanApplicability declares the conditions under which a Validation Plan
// Revision applies. PEOS-006 lists "applicability" among the items every
// Validation Plan Revision SHALL identify, with no "where applicable" or
// "where required" qualifier -- in deliberate contrast to the two bullets
// immediately preceding it (sequencing and dependencies "where
// applicable"; responsible actors or roles "where required by the
// applicable Product contract"). PlanContent therefore requires it as a
// constructor argument and offers no WithApplicability/WithoutApplicability
// modifier.
//
// PlanApplicability is a closed two-state discriminator whose zero value
// is invalid and represents a third, unstated state PEOS-006 does not
// permit. NewUnrestrictedPlanApplicability constructs "no restriction" as a
// distinct, non-zero value -- this is what makes explicit unrestricted
// applicability distinguishable from an unstated one. A *core.Scope or a
// bare (core.Scope, bool) pair cannot express that distinction, which is
// why neither is used.
//
// PlanApplicability is deliberately NOT requirement.Applicability, and is
// not converted to or from it. That type is PEOS-005 §11 Requirement
// content, describing when a Requirement's required engineering intent
// applies; this one describes when a Validation Plan Revision applies. The
// two answer different questions for different owning specifications, and
// reusing the PEOS-005 type here would additionally require importing
// peos/requirement, which this package's import boundary forbids. The
// shape is duplicated deliberately (an implementation choice), the concept
// is not.
//
// PlanApplicability carries no identity, no revision, no lifecycle, and no
// extension. This package does not interpret core.Scope's Expression in
// any way.
type PlanApplicability struct {
	kind  planApplicabilityKind
	scope core.Scope
}

// NewUnrestrictedPlanApplicability returns a PlanApplicability declaring
// explicitly that the Plan Revision's applicability is not restricted.
// The returned value is non-zero: an explicit "unrestricted" is a stated
// applicability, not an absent one.
func NewUnrestrictedPlanApplicability() PlanApplicability {
	return PlanApplicability{kind: planApplicabilityKindUnrestricted}
}

// NewScopedPlanApplicability validates scope and returns a
// PlanApplicability bound to an explicit condition expression.
func NewScopedPlanApplicability(scope core.Scope) (PlanApplicability, error) {
	if scope.IsZero() {
		return PlanApplicability{}, fmt.Errorf("validation: NewScopedPlanApplicability: %w: scope must not be zero", ErrInvalidPlanApplicability)
	}
	return PlanApplicability{kind: planApplicabilityKindScoped, scope: scope}, nil
}

// Kind returns a's discriminator, "unrestricted" or "scoped". The zero
// value returns the empty string.
func (a PlanApplicability) Kind() string { return string(a.kind) }

// IsUnrestricted reports whether a explicitly declares unrestricted
// applicability.
func (a PlanApplicability) IsUnrestricted() bool {
	return a.kind == planApplicabilityKindUnrestricted
}

// IsScoped reports whether a declares a scoped applicability.
func (a PlanApplicability) IsScoped() bool { return a.kind == planApplicabilityKindScoped }

// Scope returns a's condition expression, and whether one is set (that is,
// whether a is the scoped variant).
func (a PlanApplicability) Scope() (core.Scope, bool) {
	if a.kind != planApplicabilityKindScoped {
		return core.Scope{}, false
	}
	return a.scope, true
}

// IsZero reports whether a is the zero value -- the unstated state
// PEOS-006 does not permit on a valid PlanContent.
func (a PlanApplicability) IsZero() bool { return a.kind == "" }

type planApplicabilityJSON struct {
	Kind  string      `json:"kind"`
	Scope *core.Scope `json:"scope,omitempty"`
}

// MarshalJSON encodes a as {"kind":"unrestricted"} or {"kind":"scoped",
// "scope":{...}}. There is no top-level type discriminator beyond this
// union's own "kind".
func (a PlanApplicability) MarshalJSON() ([]byte, error) {
	switch a.kind {
	case planApplicabilityKindUnrestricted:
		return json.Marshal(planApplicabilityJSON{Kind: string(planApplicabilityKindUnrestricted)})
	case planApplicabilityKindScoped:
		return json.Marshal(planApplicabilityJSON{Kind: string(planApplicabilityKindScoped), Scope: &a.scope})
	default:
		return nil, fmt.Errorf("validation: marshal PlanApplicability: %w", ErrInvalidPlanApplicability)
	}
}

// UnmarshalJSON decodes a from its JSON form. An unrecognized or missing
// kind, an unrestricted value carrying a scope, and a scoped value missing
// a scope are all rejected.
//
// An explicit JSON null for the whole value decodes to an empty kind and
// is therefore rejected by the default case below, with
// ErrInvalidPlanApplicability -- null is never silently accepted as
// "unstated" or as "unrestricted". A scoped value whose "scope" key is
// explicitly null is likewise rejected: encoding/json leaves the *core.Scope
// field nil for both an absent and a null "scope", and both cases are
// errors here, so the two need not be distinguished.
func (a *PlanApplicability) UnmarshalJSON(data []byte) error {
	var raw planApplicabilityJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal PlanApplicability: %w: %w", ErrInvalidPlanApplicability, err)
	}
	var result PlanApplicability
	switch raw.Kind {
	case string(planApplicabilityKindUnrestricted):
		if raw.Scope != nil {
			return fmt.Errorf("validation: unmarshal PlanApplicability: %w: unrestricted must not carry a scope", ErrInvalidPlanApplicability)
		}
		result = NewUnrestrictedPlanApplicability()
	case string(planApplicabilityKindScoped):
		if raw.Scope == nil {
			return fmt.Errorf("validation: unmarshal PlanApplicability: %w: scoped requires a scope", ErrInvalidPlanApplicability)
		}
		var err error
		result, err = NewScopedPlanApplicability(*raw.Scope)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("validation: unmarshal PlanApplicability: unrecognized kind %q: %w", raw.Kind, ErrInvalidPlanApplicability)
	}
	*a = result
	return nil
}

// --- PlanContent -------------------------------------------------------------

// PlanContent is the typed normative content PEOS-006 assigns to every
// Artifact Revision whose Artifact is a Validation Plan: its intended
// scope, its applicability, its provenance, and the Planned Validation
// Activities it contains.
//
// All four are mandatory constructor arguments of NewPlanContent, and none
// is reachable through a later With* call. PEOS-006 states each of them
// without a conditional qualifier, and an aggregate that could exist
// without any of them would be normatively incomplete. This is why
// PlanContent exposes no WithScope, WithoutScope, WithApplicability,
// WithoutApplicability, WithProvenance, WithoutProvenance, WithActivities,
// or WithoutActivities: replacing mandatory aggregate state means
// constructing a new PlanContent (and, per PEOS-006, a new Artifact
// Revision to carry it).
//
// acceptanceRules is an optional Plan-level supplement. PEOS-006 lists
// "the acceptance or evaluation rules applicable to those Activities"
// among a Plan Revision's SHALL-identify items without a qualifier, but
// that obligation is discharged by PlannedActivity's own mandatory
// outcomeInterpretation, which states per Activity "how its expected
// outcome is to be interpreted". This field exists for Plan-wide rules
// that do not belong to any single Activity; it deliberately does not
// duplicate or override per-Activity outcome interpretation.
//
// PlanContent accumulates no execution history. A Validation Execution
// Record references the exact Plan Revision and plan-local key it
// executed; the Plan Revision never gains a field recording that an
// execution occurred. Any such accumulation would require mutating a
// recorded Plan Revision, which PEOS-006 forbids: "Creation of a new
// Validation Plan Revision does not mutate, delete, or reinterpret a
// previous Validation Plan Revision."
type PlanContent struct {
	scope         core.Scope
	applicability PlanApplicability
	provenance    core.Provenance
	activities    []PlannedActivity

	acceptanceRules string
	extension       core.Extension
}

// NewPlanContent validates scope, applicability, provenance, and
// activities and returns a PlanContent with no acceptance rules and no
// extension data. Use WithAcceptanceRules and WithExtension to add those.
//
// All four arguments are mandatory. scope must be non-zero (rejected with
// core.ErrInvalidScope, its owning sentinel). applicability must be
// explicitly stated -- its zero value is rejected with
// ErrInvalidPlanApplicability; use NewUnrestrictedPlanApplicability to
// declare an explicit absence of restriction. provenance must be non-zero.
// activities must contain at least one element (PEOS-006: a Plan Revision
// "contains one or more Planned Validation Activities"), no zero-value
// element, no repeated plan-local key, and no dependency naming a key that
// no Activity in the same list defines.
//
// The activities slice is defensively copied; the caller may reuse or
// mutate its own slice afterward without affecting the returned value.
// Dependency resolution is order-independent: the whole key set is
// collected before any dependency is checked, so an Activity may declare a
// dependency on an Activity appearing later in the list.
//
// PEOS-006 states no cycle policy and no self-reference prohibition for
// Planned Validation Activity dependencies. A self-dependency (an Activity
// naming its own key) and a dependency cycle among several Activities are
// therefore both accepted here: rejecting them would enforce a rule the
// specification does not state. Dependency and sequencing semantics,
// including any cycle prohibition, are repository- or Product-owned. This
// package provides no graph or cycle API.
func NewPlanContent(
	scope core.Scope,
	applicability PlanApplicability,
	provenance core.Provenance,
	activities []PlannedActivity,
) (PlanContent, error) {
	if scope.IsZero() {
		return PlanContent{}, fmt.Errorf("validation: NewPlanContent: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if applicability.IsZero() {
		return PlanContent{}, fmt.Errorf("validation: NewPlanContent: %w: applicability must be explicitly stated", ErrInvalidPlanApplicability)
	}
	if provenance.IsZero() {
		return PlanContent{}, fmt.Errorf("validation: NewPlanContent: %w: provenance must not be zero", ErrInvalidValidationPlan)
	}
	if len(activities) == 0 {
		return PlanContent{}, fmt.Errorf("validation: NewPlanContent: %w: at least one planned validation activity is required", ErrInvalidValidationPlan)
	}

	acts := make([]PlannedActivity, len(activities))
	keys := make(map[string]bool, len(activities))
	for idx, a := range activities {
		if a.IsZero() {
			return PlanContent{}, fmt.Errorf("validation: NewPlanContent: %w: planned validation activity must not be zero", ErrInvalidPlannedActivity)
		}
		key := a.Key().String()
		if keys[key] {
			return PlanContent{}, fmt.Errorf("validation: NewPlanContent: plan-local key %q: %w", key, ErrDuplicatePlanLocalKey)
		}
		keys[key] = true
		acts[idx] = a
	}

	for _, a := range acts {
		for _, dep := range a.dependencies {
			if !keys[dep.String()] {
				return PlanContent{}, fmt.Errorf("validation: NewPlanContent: activity %q depends on %q: %w", a.Key().String(), dep.String(), ErrUnknownPlanLocalKey)
			}
		}
	}

	return PlanContent{
		scope:         scope,
		applicability: applicability,
		provenance:    provenance,
		activities:    acts,
	}, nil
}

// WithAcceptanceRules returns a copy of c with its Plan-level acceptance
// or evaluation rules set. rules must be non-empty after trimming
// surrounding whitespace; the trimmed value is stored. Use
// WithoutAcceptanceRules to clear a previously set value.
func (c PlanContent) WithAcceptanceRules(rules string) (PlanContent, error) {
	trimmed := strings.TrimSpace(rules)
	if trimmed == "" {
		return PlanContent{}, fmt.Errorf("validation: PlanContent.WithAcceptanceRules: %w: acceptance rules must not be empty", ErrInvalidValidationPlan)
	}
	c.acceptanceRules = trimmed
	return c, nil
}

// WithoutAcceptanceRules returns a copy of c with its Plan-level
// acceptance rules cleared.
func (c PlanContent) WithoutAcceptanceRules() PlanContent {
	c.acceptanceRules = ""
	return c
}

// WithExtension returns a copy of c with its extension data set. Passing
// the zero core.Extension is equivalent to declaring none, per
// core.Extension's own documented contract.
func (c PlanContent) WithExtension(extension core.Extension) PlanContent {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c PlanContent) WithoutExtension() PlanContent {
	c.extension = core.Extension{}
	return c
}

// Scope returns c's declared intended scope. It is mandatory and therefore
// never absent on a valid PlanContent.
func (c PlanContent) Scope() core.Scope { return c.scope }

// Applicability returns c's declared applicability. It is mandatory and
// therefore never unstated on a valid PlanContent.
func (c PlanContent) Applicability() PlanApplicability { return c.applicability }

// Provenance returns c's declared provenance.
func (c PlanContent) Provenance() core.Provenance { return c.provenance }

// Activities returns a defensive copy of c's Planned Validation
// Activities, in declaration order.
func (c PlanContent) Activities() []PlannedActivity {
	if len(c.activities) == 0 {
		return nil
	}
	cp := make([]PlannedActivity, len(c.activities))
	copy(cp, c.activities)
	return cp
}

// Activity returns the Planned Validation Activity in c whose plan-local
// key equals key, and whether one was found. The key is meaningful only
// within this exact Plan Revision's content.
func (c PlanContent) Activity(key core.LocalKey) (PlannedActivity, bool) {
	if key.IsZero() {
		return PlannedActivity{}, false
	}
	for _, a := range c.activities {
		if a.key == key {
			return a, true
		}
	}
	return PlannedActivity{}, false
}

// AcceptanceRules returns c's Plan-level acceptance rules, and whether any
// are set.
func (c PlanContent) AcceptanceRules() (string, bool) {
	return c.acceptanceRules, c.acceptanceRules != ""
}

// Extension returns c's extension data.
func (c PlanContent) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c PlanContent) IsZero() bool {
	return c.scope.IsZero() && c.applicability.IsZero() && c.provenance.IsZero() && len(c.activities) == 0
}

type planContentJSON struct {
	Scope           core.Scope        `json:"scope"`
	Applicability   PlanApplicability `json:"applicability"`
	Provenance      core.Provenance   `json:"provenance"`
	Activities      []PlannedActivity `json:"activities"`
	AcceptanceRules string            `json:"acceptance_rules,omitempty"`
	Extension       *core.Extension   `json:"extension,omitempty"`
}

// planContentUnmarshalJSON mirrors planContentJSON's field set for
// decoding only, with one difference: AcceptanceRules is captured as raw,
// undecoded bytes rather than a string, so an explicit JSON null can be
// distinguished from an absent key and rejected -- the technique Packet
// D.1 established for relation.Relation's Scope and
// lifecycle.StateAssignment's Authority.
//
// The four mandatory keys need no such treatment: for each of them an
// absent key and an explicit null both yield a zero value that
// NewPlanContent rejects, so the two cases converge on the same error and
// need not be told apart.
type planContentUnmarshalJSON struct {
	Scope           core.Scope        `json:"scope"`
	Applicability   PlanApplicability `json:"applicability"`
	Provenance      core.Provenance   `json:"provenance"`
	Activities      []PlannedActivity `json:"activities"`
	AcceptanceRules json.RawMessage   `json:"acceptance_rules"`
	Extension       *core.Extension   `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"scope":..., "applicability":...,
// "provenance":..., "activities":[...]}, omitting acceptance_rules and
// extension when not set. There is no "relation" key, no lifecycle key,
// and no top-level type discriminator.
func (c PlanContent) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("validation: marshal PlanContent: %w", ErrInvalidValidationPlan)
	}
	raw := planContentJSON{
		Scope:         c.scope,
		Applicability: c.applicability,
		Provenance:    c.provenance,
		Activities:    c.activities,
	}
	if c.acceptanceRules != "" {
		raw.AcceptanceRules = c.acceptanceRules
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as NewPlanContent and each With* method.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - scope, provenance: a missing key leaves the field zero and reaches
//     NewPlanContent, which rejects it through its owning sentinel
//     (core.ErrInvalidScope, ErrInvalidValidationPlan). An explicit null
//     invokes that nested type's own UnmarshalJSON, which fails there, so
//     the error is wrapped here with ErrInvalidValidationPlan instead.
//     Both are rejected; the sentinel sets differ.
//   - applicability: both a missing key and an explicit null yield an
//     empty kind, rejected with ErrInvalidPlanApplicability.
//   - activities: both a missing key and an explicit null yield an empty
//     slice, rejected with ErrInvalidValidationPlan.
//   - acceptance_rules: a missing key means absent; an explicit null is
//     rejected rather than silently treated as absent.
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract.
func (c *PlanContent) UnmarshalJSON(data []byte) error {
	var raw planContentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal PlanContent: %w: %w", ErrInvalidValidationPlan, err)
	}
	result, err := NewPlanContent(raw.Scope, raw.Applicability, raw.Provenance, raw.Activities)
	if err != nil {
		return err
	}
	if len(raw.AcceptanceRules) > 0 {
		if string(raw.AcceptanceRules) == "null" {
			return fmt.Errorf("validation: unmarshal PlanContent: %w: acceptance rules must not be null", ErrInvalidValidationPlan)
		}
		var rules string
		if err := json.Unmarshal(raw.AcceptanceRules, &rules); err != nil {
			return fmt.Errorf("validation: unmarshal PlanContent: %w: %w", ErrInvalidValidationPlan, err)
		}
		if result, err = result.WithAcceptanceRules(rules); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*c = result
	return nil
}

// --- PlanRevision ------------------------------------------------------------

// PlanRevision is shorthand for "an Artifact Revision whose Artifact is a
// Validation Plan" -- not a separate PEOS entity, and not a Validation
// Plan Version. It composes core.ArtifactRevision by named field, per the
// specialized-Revision strategy core.ArtifactRevision itself documents and
// requirement.Revision already follows, and pairs it with typed
// PlanContent.
//
// PlanRevision is immutable and exposes no WithContent: changing a
// Validation Plan's content "constitutes a content change and SHALL create
// a new Artifact Revision in accordance with PEOS-002", so a new
// PlanRevision is constructed rather than an existing one edited.
type PlanRevision struct {
	core    core.ArtifactRevision
	content PlanContent
}

// newPlanRevisionFromParts validates revision and content without
// reference to any Plan, and is the path both NewPlanRevision and
// UnmarshalJSON share. It cannot, and does not attempt to, check that
// revision belongs to any particular Plan -- see NewPlanRevision and
// UnmarshalJSON for why that check needs a Plan value a PlanRevision's own
// JSON does not carry.
func newPlanRevisionFromParts(revision core.ArtifactRevision, content PlanContent) (PlanRevision, error) {
	if revision.IsZero() {
		return PlanRevision{}, fmt.Errorf("%w: core revision must not be zero", ErrInvalidValidationPlan)
	}
	if content.IsZero() {
		return PlanRevision{}, fmt.Errorf("%w: plan content must not be zero", ErrInvalidValidationPlan)
	}
	return PlanRevision{core: revision, content: content}, nil
}

// NewPlanRevision validates plan, revision, and content and returns a
// PlanRevision. plan and revision must both be non-zero, content must be
// non-zero, and revision.ArtifactID() must equal plan.ID().
func NewPlanRevision(plan Plan, revision core.ArtifactRevision, content PlanContent) (PlanRevision, error) {
	if plan.IsZero() {
		return PlanRevision{}, fmt.Errorf("validation: NewPlanRevision: %w: plan must not be zero", ErrInvalidValidationPlan)
	}
	result, err := newPlanRevisionFromParts(revision, content)
	if err != nil {
		return PlanRevision{}, fmt.Errorf("validation: NewPlanRevision: %w", err)
	}
	if revision.ArtifactID() != plan.ID() {
		return PlanRevision{}, fmt.Errorf("validation: NewPlanRevision: %w", ErrValidationPlanArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the PlanRevision's underlying core.ArtifactRevision.
func (r PlanRevision) Core() core.ArtifactRevision { return r.core }

// Content returns the PlanRevision's typed Validation Plan content.
func (r PlanRevision) Content() PlanContent { return r.content }

// Ref returns a core.ValidationPlanRevisionRef identifying r. This is the
// reference a Validation Execution Record pairs with a plan-local key to
// name one exact Planned Validation Activity.
func (r PlanRevision) Ref() (core.ValidationPlanRevisionRef, error) {
	return core.NewValidationPlanRevisionRef(r.core.ArtifactID(), r.core.RevisionID())
}

// IsZero reports whether r is the zero value.
func (r PlanRevision) IsZero() bool { return r.core.IsZero() && r.content.IsZero() }

type planRevisionJSON struct {
	Core    core.ArtifactRevision `json:"core"`
	Content PlanContent           `json:"content"`
}

// MarshalJSON encodes r as {"core":{...},"content":{...}}, per the
// nested-composition strategy core.ArtifactRevision documents.
func (r PlanRevision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("validation: marshal PlanRevision: %w", ErrInvalidValidationPlan)
	}
	return json.Marshal(planRevisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core":{...},"content":{...}}
// JSON form.
//
// This reconstructs r.core and r.content via the same checks
// newPlanRevisionFromParts (and therefore NewPlanRevision) applies, but
// cannot repeat NewPlanRevision's ArtifactID-to-Plan cross-check: a
// PlanRevision's own JSON carries only its core.ArtifactRevision (with a
// bare ArtifactID) and its PlanContent, never a core.Artifact with an
// ArtifactType to check that ArtifactID against. This is the same
// limitation core.ArtifactRevision and requirement.Revision already
// document. A PlanRevision decoded on its own is therefore a value
// encoding of its own fields, not a self-sufficient record of "this
// Revision of that Validation Plan" without external context supplying the
// matching Plan.
func (r *PlanRevision) UnmarshalJSON(data []byte) error {
	var raw planRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal PlanRevision: %w: %w", ErrInvalidValidationPlan, err)
	}
	result, err := newPlanRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return fmt.Errorf("validation: unmarshal PlanRevision: %w", err)
	}
	*r = result
	return nil
}
