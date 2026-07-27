package quality

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustArtifactTypeQualityProfile() core.ArtifactType {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, "quality-profile")
	if err != nil {
		panic(err)
	}
	return core.NewArtifactType(v)
}

// ArtifactTypeQualityProfile is the PEOS-007 Quality Profile Artifact Type.
// PEOS-007 does not itself fix an exact vocabulary string for this value --
// this is an implementation choice, namespaced under core.PEOSNamespace
// because Quality Profile is a PEOS-000-009-defined Artifact Type rather
// than a Product-specific one, matching the convention
// requirement.ArtifactTypeRequirement and validation.ArtifactTypeValidationPlan
// already established. core.ArtifactType's own vocabulary remains fully open.
//
// This value deliberately lives in peos/quality, not peos/core, for the same
// reason those two live in their own packages: the Artifact Type belongs to
// the specialization that defines it.
var ArtifactTypeQualityProfile = mustArtifactTypeQualityProfile()

// --- Profile -----------------------------------------------------------------

// Profile is a PEOS-007 Quality Profile identity: a core.Artifact whose
// declared Artifact Type is ArtifactTypeQualityProfile ("A Quality Profile is
// an Artifact as defined by PEOS-002").
//
// Profile adds no field of its own. Every Quality Profile content element --
// scope, applicability, provenance, and the Characteristics, Measures,
// Thresholds, Targets, Constraints, and Rules themselves -- is Revision-owned
// content carried by ProfileContent, never Profile identity. Profile
// therefore has no Version field of any kind: "A Quality Profile uses
// ordinary Artifact Revision for all of its evolution. There is no Quality
// Profile Version distinct from Artifact Revision" (No Quality Profile
// Version Invariant; non-conforming pattern "Quality Profile Version").
//
// Profile carries no Lifecycle State and no extension of its own. A Quality
// Profile is an ordinary Artifact and MAY be governed by a Lifecycle under
// PEOS-003; that lifecycle is modeled exclusively in peos/lifecycle, which
// this package does not import. Profile also carries no quality score:
// "A quality score SHALL NOT be stored as globally current mutable state on:
// an Artifact; an Artifact Revision; a Requirement; a Quality Profile."
// Product-specific data belongs on the underlying core.Artifact's own
// extension, or on ProfileContent.
type Profile struct {
	core core.Artifact
}

// NewProfile validates artifact and returns a Profile. artifact must be
// non-zero and its Type() must equal ArtifactTypeQualityProfile.
func NewProfile(artifact core.Artifact) (Profile, error) {
	if artifact.IsZero() {
		return Profile{}, fmt.Errorf("quality: NewProfile: %w: artifact must not be zero", ErrInvalidQualityProfile)
	}
	if artifact.Type() != ArtifactTypeQualityProfile {
		return Profile{}, fmt.Errorf("quality: NewProfile: %w", ErrQualityProfileArtifactTypeMismatch)
	}
	return Profile{core: artifact}, nil
}

// Core returns the Profile's underlying core.Artifact.
func (p Profile) Core() core.Artifact { return p.core }

// ID returns the Profile's identity.
func (p Profile) ID() core.ArtifactID { return p.core.ID() }

// Ref returns a core.ArtifactRef identifying p.
//
// peos/core deliberately defines no QualityProfileRef type: core/criterion.go
// records that "this package does not define a dedicated
// QualityProfileRevisionRef type, so the owning Revision is referenced with
// the general-purpose ArtifactRevisionRef." The same reasoning applies at the
// Artifact level, so the general-purpose core.ArtifactRef is used here rather
// than introducing a PEOS-007-specific reference type into core.
func (p Profile) Ref() (core.ArtifactRef, error) {
	return core.NewArtifactRef(p.core.ID())
}

// IsZero reports whether p is the zero value.
func (p Profile) IsZero() bool { return p.core.IsZero() }

// MarshalJSON encodes p as the wire form of its underlying core.Artifact,
// with no additional envelope -- the same strategy requirement.Requirement
// and validation.Plan use. core.Artifact's own JSON already carries
// artifact_type, which both preserves and (on Unmarshal) lets NewProfile
// re-verify that the decoded value is a Quality Profile.
func (p Profile) MarshalJSON() ([]byte, error) {
	if p.IsZero() {
		return nil, fmt.Errorf("quality: marshal Profile: %w", ErrInvalidQualityProfile)
	}
	return json.Marshal(p.core)
}

// UnmarshalJSON decodes p from its JSON form, applying the same validation as
// NewProfile. An explicit JSON null decodes core.Artifact to its zero value,
// which NewProfile then rejects with ErrInvalidQualityProfile; a decoded
// Profile can never be constructor-impossible. The receiver is left untouched
// unless every check passes.
func (p *Profile) UnmarshalJSON(data []byte) error {
	var artifact core.Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return fmt.Errorf("quality: unmarshal Profile: %w: %w", ErrInvalidQualityProfile, err)
	}
	result, err := NewProfile(artifact)
	if err != nil {
		return err
	}
	*p = result
	return nil
}

// --- ProfileApplicability ----------------------------------------------------

type profileApplicabilityKind string

const (
	profileApplicabilityKindUnrestricted profileApplicabilityKind = "unrestricted"
	profileApplicabilityKindScoped       profileApplicabilityKind = "scoped"
)

// ProfileApplicability declares the conditions under which a Quality Profile
// Revision applies. PEOS-007 lists "its applicability conditions" among the
// items every Quality Profile Revision SHALL identify, with no "where
// applicable" or "where required" qualifier -- in deliberate contrast to the
// two bullets immediately preceding it (Normalization Rules and Aggregation
// Rules, both "where applicable") and the one following provenance
// ("authority, where required"). ProfileContent therefore requires it as a
// constructor argument and offers no
// WithApplicability/WithoutApplicability modifier.
//
// ProfileApplicability is a closed two-state discriminator whose zero value
// is invalid and represents a third, unstated state PEOS-007 does not permit.
// NewUnrestrictedProfileApplicability constructs "no restriction" as a
// distinct, non-zero value -- this is what makes explicit unrestricted
// applicability distinguishable from an unstated one. A *core.Scope or a bare
// (core.Scope, bool) pair cannot express that distinction, which is why
// neither is used.
//
// ProfileApplicability is deliberately not validation.PlanApplicability and
// not requirement.Applicability, and is not converted to or from either. Each
// answers a different question for a different owning specification: when a
// Requirement's intent applies (PEOS-005), when a Validation Plan Revision
// applies (PEOS-006), and when a Quality Profile Revision applies (PEOS-007).
// The shape is duplicated deliberately (an implementation choice); the
// concept is not. Reusing the PEOS-006 type would additionally couple this
// package's applicability semantics to that package's.
//
// ProfileApplicability carries no identity, no revision, no lifecycle, and no
// extension. This package does not interpret core.Scope's expression in any
// way.
type ProfileApplicability struct {
	kind  profileApplicabilityKind
	scope core.Scope
}

// NewUnrestrictedProfileApplicability returns a ProfileApplicability
// declaring explicitly that the Profile Revision's applicability is not
// restricted. The returned value is non-zero: an explicit "unrestricted" is a
// stated applicability, not an absent one.
func NewUnrestrictedProfileApplicability() ProfileApplicability {
	return ProfileApplicability{kind: profileApplicabilityKindUnrestricted}
}

// NewScopedProfileApplicability validates scope and returns a
// ProfileApplicability bound to an explicit condition expression.
func NewScopedProfileApplicability(scope core.Scope) (ProfileApplicability, error) {
	if scope.IsZero() {
		return ProfileApplicability{}, fmt.Errorf("quality: NewScopedProfileApplicability: %w: scope must not be zero", ErrInvalidProfileApplicability)
	}
	return ProfileApplicability{kind: profileApplicabilityKindScoped, scope: scope}, nil
}

// Kind returns a's discriminator, "unrestricted" or "scoped". The zero value
// returns the empty string.
func (a ProfileApplicability) Kind() string { return string(a.kind) }

// IsUnrestricted reports whether a explicitly declares unrestricted
// applicability.
func (a ProfileApplicability) IsUnrestricted() bool {
	return a.kind == profileApplicabilityKindUnrestricted
}

// IsScoped reports whether a declares a scoped applicability.
func (a ProfileApplicability) IsScoped() bool { return a.kind == profileApplicabilityKindScoped }

// Scope returns a's condition expression, and whether one is set (that is,
// whether a is the scoped variant).
func (a ProfileApplicability) Scope() (core.Scope, bool) {
	if a.kind != profileApplicabilityKindScoped {
		return core.Scope{}, false
	}
	return a.scope, true
}

// IsZero reports whether a is the zero value -- the unstated state PEOS-007
// does not permit on a valid ProfileContent.
func (a ProfileApplicability) IsZero() bool { return a.kind == "" }

type profileApplicabilityJSON struct {
	Kind  string      `json:"kind"`
	Scope *core.Scope `json:"scope,omitempty"`
}

// MarshalJSON encodes a as {"kind":"unrestricted"} or {"kind":"scoped",
// "scope":{...}}. There is no top-level type discriminator beyond this
// union's own "kind".
func (a ProfileApplicability) MarshalJSON() ([]byte, error) {
	switch a.kind {
	case profileApplicabilityKindUnrestricted:
		return json.Marshal(profileApplicabilityJSON{Kind: string(profileApplicabilityKindUnrestricted)})
	case profileApplicabilityKindScoped:
		return json.Marshal(profileApplicabilityJSON{Kind: string(profileApplicabilityKindScoped), Scope: &a.scope})
	default:
		return nil, fmt.Errorf("quality: marshal ProfileApplicability: %w", ErrInvalidProfileApplicability)
	}
}

// UnmarshalJSON decodes a from its JSON form. An unrecognized or missing
// kind, an unrestricted value carrying a scope, and a scoped value missing a
// scope are all rejected.
//
// An explicit JSON null for the whole value decodes to an empty kind and is
// therefore rejected by the default case below, with
// ErrInvalidProfileApplicability -- null is never silently accepted as
// "unstated" or as "unrestricted". A scoped value whose "scope" key is
// explicitly null is likewise rejected: encoding/json leaves the *core.Scope
// field nil for both an absent and a null "scope", and both cases are errors
// here, so the two need not be distinguished. The receiver is left untouched
// unless every check passes.
func (a *ProfileApplicability) UnmarshalJSON(data []byte) error {
	var raw profileApplicabilityJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal ProfileApplicability: %w: %w", ErrInvalidProfileApplicability, err)
	}
	var result ProfileApplicability
	switch raw.Kind {
	case string(profileApplicabilityKindUnrestricted):
		if raw.Scope != nil {
			return fmt.Errorf("quality: unmarshal ProfileApplicability: %w: unrestricted must not carry a scope", ErrInvalidProfileApplicability)
		}
		result = NewUnrestrictedProfileApplicability()
	case string(profileApplicabilityKindScoped):
		if raw.Scope == nil {
			return fmt.Errorf("quality: unmarshal ProfileApplicability: %w: scoped requires a scope", ErrInvalidProfileApplicability)
		}
		var err error
		result, err = NewScopedProfileApplicability(*raw.Scope)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("quality: unmarshal ProfileApplicability: unrecognized kind %q: %w", raw.Kind, ErrInvalidProfileApplicability)
	}
	*a = result
	return nil
}

// --- profile-local key namespaces --------------------------------------------

// Owned-value kind labels, used in duplicate- and unknown-key error messages
// so that a caller can identify which collection failed without needing a
// per-kind sentinel.
const (
	kindCharacteristic    = "characteristic"
	kindMeasure           = "measure"
	kindThreshold         = "threshold"
	kindTarget            = "target"
	kindConstraint        = "constraint"
	kindNormalizationRule = "normalization rule"
	kindAggregationRule   = "aggregation rule"
)

// addProfileLocalKey records key in set, rejecting a repeat within that one
// collection.
//
// Uniqueness is per owned-value kind. PEOS-007 states no key uniqueness rule
// at all: the word "unique" does not appear in the specification, and unlike
// PEOS-006 -- which states plan-local key uniqueness explicitly -- PEOS-007
// establishes only Revision *ownership* for its owned values. The derived rule
// implemented here is the minimum the reference model needs: a criterion
// citing a Characteristic by key must resolve to exactly one Characteristic,
// so two Characteristics may not share a key. Cross-kind uniqueness is not
// needed, because every reference already determines its target collection,
// and asserting it would enforce a rule the specification does not state.
func addProfileLocalKey(caller, kind string, set map[string]bool, key core.LocalKey) error {
	s := key.String()
	if set[s] {
		return fmt.Errorf("quality: %s: %s key %q: %w", caller, kind, s, ErrDuplicateProfileLocalKey)
	}
	set[s] = true
	return nil
}

// resolveProfileLocalKey checks that key is defined in the collection named
// by targetKind, whose key set is target.
func resolveProfileLocalKey(caller, referringKind, referringKey, targetKind string, target map[string]bool, key core.LocalKey) error {
	if target[key.String()] {
		return nil
	}
	return fmt.Errorf("quality: %s: %s %q references %s %q: %w: no %s with that key is defined in this profile revision",
		caller, referringKind, referringKey, targetKind, key.String(), ErrUnknownProfileLocalKey, targetKind)
}

// --- ProfileContent ----------------------------------------------------------

// ProfileContent is the typed normative content PEOS-007 assigns to every
// Artifact Revision whose Artifact is a Quality Profile: its scope, its
// applicability, its provenance, and the Characteristics, Measures,
// Thresholds, Targets, Constraints, Normalization Rules, and Aggregation
// Rules it defines.
//
// # Mandatory versus optional
//
// scope, applicability, and provenance are mandatory constructor arguments
// and are unreachable through any later With* call: PEOS-007 states each
// without a conditional qualifier, and an aggregate that could exist without
// any of them would be normatively incomplete. This is why ProfileContent
// exposes no WithScope, WithoutScope, WithApplicability,
// WithoutApplicability, WithProvenance, or WithoutProvenance.
//
// characteristics and measures are constructor arguments, but -- unlike
// scope, applicability, and provenance -- neither requires a non-empty
// collection. PEOS-007 lists both without a qualifier, but that wording is
// identical in form to Subjects, Thresholds, Targets, and Quality
// Constraints (also unqualified) and to Normalization and Aggregation Rules
// ("where applicable"), all six of which this package already treats as
// satisfied by an empty collection. Nothing in the specification states any
// minimum cardinality for a Quality Profile Revision's content -- the word
// "one" never appears alongside any of these seven items -- so a Revision
// that identifies an empty set of Characteristics, of Measures, or of both
// has identified that set, exactly as it does for the other five kinds.
// This repository already ships and audits this exact reading elsewhere:
// validation.PlannedActivity treats its own unqualified "criteria", "Evidence
// expected", and "execution prerequisites" items the same way.
//
// A Quality Profile Revision containing only Quality Constraints, only
// Quality Characteristics, only Normalization or Aggregation Rules, or no
// content at all is therefore normatively valid: a Quality Constraint (`:196`)
// depends on no Characteristic or Measure and is independently citable as a
// core.CriterionKindQualityConstraint criterion; a Characteristic (`:146`) is
// a complete definitional act on its own and is independently citable as
// core.CriterionKindQualityCharacteristic; and neither Normalization nor
// Aggregation Rules are referenced by anything unless a Measure or an
// Aggregation consumer chooses to. Whether a given Revision is complete
// enough to publish, apply, or approve is repository- and Product-owned, not
// a PEOS-007 value-layer concern -- the same boundary this package already
// draws for unit semantics, threshold comparison, and aggregation execution.
//
// Coherence for the collections that do reference something is preserved by
// conditional dependencies, not by a global minimum: every Measure's
// Characteristic key must resolve among characteristics; every Measure's
// optional Normalization Rule key, if present, must resolve among
// normalizationRules; every Threshold's and every Target's Measure key must
// resolve among measures. A Measure-only, Threshold-only, or Target-only
// Revision is therefore still rejected -- not by a cardinality rule on
// Characteristics or Measures, but because the reference each of them
// carries cannot resolve against an empty target collection.
//
// normalizationRules is a constructor argument even though it is optional
// content that may legitimately be empty. It has to be, because a Measure MAY
// reference a Normalization Rule by profile-local key, and that reference is
// resolved against this collection. Were the rules supplied only through a
// later WithNormalizationRules call, a Measure carrying such a reference could
// never be constructed at all: the constructor would reject it as an unknown
// key before the rules could be added, and no WithMeasures exists to attach
// the reference afterwards. Accepting it as an argument is what keeps the
// aggregate constructor-complete -- a mandatory-given-another-argument value
// must be a constructor argument. aggregationRules needs no such treatment
// because nothing references it.
//
// subjects and subjectTypes are optional. PEOS-007 lists "the Subjects or
// Subject types to which it applies" among the SHALL-identify items, but that
// obligation is discharged by the mandatory applicability, which states
// exactly when the Revision applies -- an explicitly unrestricted
// applicability says it applies everywhere, and a scoped one carries the
// condition. These two fields are a more precise optional supplement to that
// statement, the same relationship validation.PlanContent's optional
// acceptanceRules has to its Activities' mandatory outcome interpretation.
// See the package documentation for this derivation's status.
//
// # No derived state
//
// ProfileContent stores no quality score, no aggregate, no pass/fail, and no
// evaluation outcome, and it accumulates no measurement history. A
// Measurement Record references the exact Profile Revision and profile-local
// key it applied; the Profile Revision never gains a field recording that a
// measurement occurred. Any such accumulation would require mutating a
// recorded Profile Revision, and would be the non-conforming pattern "Mutable
// Quality Score". "Any consumer requiring a 'current' quality score MUST
// compute it, on demand, from the applicable, non-replaced, non-invalidated
// Measurement Records and Quality Claims."
type ProfileContent struct {
	scope         core.Scope
	applicability ProfileApplicability
	provenance    core.Provenance

	characteristics    []Characteristic
	measures           []Measure
	normalizationRules []NormalizationRule

	thresholds       []Threshold
	targets          []Target
	constraints      []Constraint
	aggregationRules []AggregationRule

	subjects     []core.EngineeringSubjectRef
	subjectTypes []core.VocabularyValue
	authority    core.AuthorityRef
	extension    core.Extension
}

// validateProfileContent is the single shared validation path for
// ProfileContent. NewProfileContent, every collection With* method, and
// UnmarshalJSON all route through it, so the per-kind uniqueness rules and the
// internal reference resolution rules cannot drift between construction,
// modification, and decoding.
//
// Resolution is order-independent in both directions: every collection's full
// key set is built before any reference is checked, so a Measure may reference
// a Normalization Rule declared later in its slice, and a Threshold may
// reference a Measure declared after it.
//
// PEOS-007 states no cycle policy and no self-reference prohibition for these
// references, and none is enforced here -- rejecting them would enforce a rule
// the specification does not state, exactly as validation.NewPlanContent
// declines to reject Activity dependency cycles. Reference semantics beyond
// resolvability are Product-owned.
func validateProfileContent(caller string, c ProfileContent) error {
	if c.scope.IsZero() {
		return fmt.Errorf("quality: %s: %w: scope must not be zero", caller, core.ErrInvalidScope)
	}
	if c.applicability.IsZero() {
		return fmt.Errorf("quality: %s: %w: applicability must be explicitly stated", caller, ErrInvalidProfileApplicability)
	}
	if c.provenance.IsZero() {
		return fmt.Errorf("quality: %s: %w: provenance must not be zero", caller, ErrInvalidQualityProfile)
	}

	characteristicKeys := make(map[string]bool, len(c.characteristics))
	for _, v := range c.characteristics {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: quality characteristic must not be zero", caller, ErrInvalidQualityCharacteristic)
		}
		if err := addProfileLocalKey(caller, kindCharacteristic, characteristicKeys, v.Key()); err != nil {
			return err
		}
	}
	measureKeys := make(map[string]bool, len(c.measures))
	for _, v := range c.measures {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: quality measure must not be zero", caller, ErrInvalidQualityMeasure)
		}
		if err := addProfileLocalKey(caller, kindMeasure, measureKeys, v.Key()); err != nil {
			return err
		}
	}
	normalizationKeys := make(map[string]bool, len(c.normalizationRules))
	for _, v := range c.normalizationRules {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: normalization rule must not be zero", caller, ErrInvalidQualityRule)
		}
		if err := addProfileLocalKey(caller, kindNormalizationRule, normalizationKeys, v.Key()); err != nil {
			return err
		}
	}
	thresholdKeys := make(map[string]bool, len(c.thresholds))
	for _, v := range c.thresholds {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: threshold must not be zero", caller, ErrInvalidQualityThreshold)
		}
		if err := addProfileLocalKey(caller, kindThreshold, thresholdKeys, v.Key()); err != nil {
			return err
		}
	}
	targetKeys := make(map[string]bool, len(c.targets))
	for _, v := range c.targets {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: target must not be zero", caller, ErrInvalidQualityTarget)
		}
		if err := addProfileLocalKey(caller, kindTarget, targetKeys, v.Key()); err != nil {
			return err
		}
	}
	constraintKeys := make(map[string]bool, len(c.constraints))
	for _, v := range c.constraints {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: quality constraint must not be zero", caller, ErrInvalidQualityConstraint)
		}
		if err := addProfileLocalKey(caller, kindConstraint, constraintKeys, v.Key()); err != nil {
			return err
		}
	}
	aggregationKeys := make(map[string]bool, len(c.aggregationRules))
	for _, v := range c.aggregationRules {
		if v.IsZero() {
			return fmt.Errorf("quality: %s: %w: aggregation rule must not be zero", caller, ErrInvalidQualityRule)
		}
		if err := addProfileLocalKey(caller, kindAggregationRule, aggregationKeys, v.Key()); err != nil {
			return err
		}
	}

	for _, s := range c.subjects {
		if s.IsZero() {
			return fmt.Errorf("quality: %s: %w: subject must not be zero", caller, ErrInvalidQualityProfile)
		}
	}
	for _, st := range c.subjectTypes {
		if st.IsZero() {
			return fmt.Errorf("quality: %s: %w: subject type must not be zero", caller, ErrInvalidQualityProfile)
		}
	}

	// Internal reference resolution, each against exactly one target
	// namespace. A key present only in a different collection does not
	// satisfy resolution.
	for _, m := range c.measures {
		if err := resolveProfileLocalKey(caller, kindMeasure, m.Key().String(), kindCharacteristic, characteristicKeys, m.Characteristic()); err != nil {
			return err
		}
		if key, ok := m.NormalizationRule(); ok {
			if err := resolveProfileLocalKey(caller, kindMeasure, m.Key().String(), kindNormalizationRule, normalizationKeys, key); err != nil {
				return err
			}
		}
	}
	for _, t := range c.thresholds {
		if err := resolveProfileLocalKey(caller, kindThreshold, t.Key().String(), kindMeasure, measureKeys, t.Measure()); err != nil {
			return err
		}
	}
	for _, t := range c.targets {
		if err := resolveProfileLocalKey(caller, kindTarget, t.Key().String(), kindMeasure, measureKeys, t.Measure()); err != nil {
			return err
		}
	}

	return nil
}

// NewProfileContent validates its arguments and returns a ProfileContent with
// no Thresholds, Targets, Constraints, Aggregation Rules, Subjects, Subject
// Types, authority, or extension data. Use the With* methods to add those.
//
// scope, applicability, and provenance must all be non-zero; applicability
// must be explicitly stated (use NewUnrestrictedProfileApplicability to
// declare an explicit absence of restriction). characteristics, measures, and
// normalizationRules may each be empty or nil -- PEOS-007 states no minimum
// cardinality for any Profile-owned collection -- but no element within a
// non-empty collection may be zero-valued. normalizationRules must be
// supplied here rather than later whenever any Measure references a
// Normalization Rule -- see the type documentation for why.
//
// Within each collection, profile-local keys must be unique. Across
// collections they need not be: the same key may name a Characteristic and a
// Threshold. Every internal reference must resolve in its own expected
// collection, order-independently.
//
// Every slice argument is defensively copied; the caller may reuse or mutate
// its own slices afterward without affecting the returned value.
func NewProfileContent(
	scope core.Scope,
	applicability ProfileApplicability,
	provenance core.Provenance,
	characteristics []Characteristic,
	measures []Measure,
	normalizationRules []NormalizationRule,
) (ProfileContent, error) {
	c := ProfileContent{
		scope:              scope,
		applicability:      applicability,
		provenance:         provenance,
		characteristics:    copySlice(characteristics),
		measures:           copySlice(measures),
		normalizationRules: copySlice(normalizationRules),
	}
	if err := validateProfileContent("NewProfileContent", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// copySlice returns a defensive copy of s, or nil when s is empty.
func copySlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	cp := make([]T, len(s))
	copy(cp, s)
	return cp
}

// WithThresholds returns a copy of c with its Thresholds set to exactly the
// values given, in the order given, replacing any previous Thresholds.
// Passing an empty or nil slice declares none, which is why there is no
// WithoutThresholds: WithThresholds(nil) already expresses removal, and a
// second method would create a second validation path for the same field.
//
// The result is re-validated through the same shared path NewProfileContent
// uses, so a duplicate Threshold key or a Threshold naming an unknown Measure
// is rejected here exactly as it would be at construction.
func (c ProfileContent) WithThresholds(thresholds []Threshold) (ProfileContent, error) {
	c.thresholds = copySlice(thresholds)
	if err := validateProfileContent("ProfileContent.WithThresholds", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithTargets returns a copy of c with its Targets set to exactly the values
// given, in the order given. Passing an empty or nil slice declares none. The
// result is re-validated through the shared path.
func (c ProfileContent) WithTargets(targets []Target) (ProfileContent, error) {
	c.targets = copySlice(targets)
	if err := validateProfileContent("ProfileContent.WithTargets", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithConstraints returns a copy of c with its Quality Constraints set to
// exactly the values given, in the order given. Passing an empty or nil slice
// declares none. The result is re-validated through the shared path.
func (c ProfileContent) WithConstraints(constraints []Constraint) (ProfileContent, error) {
	c.constraints = copySlice(constraints)
	if err := validateProfileContent("ProfileContent.WithConstraints", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithNormalizationRules returns a copy of c with its Normalization Rules set
// to exactly the values given, in the order given. Passing an empty or nil
// slice declares none.
//
// The result is re-validated through the shared path, so dropping a rule that
// one of c's Measures references is rejected rather than silently leaving a
// dangling reference.
func (c ProfileContent) WithNormalizationRules(rules []NormalizationRule) (ProfileContent, error) {
	c.normalizationRules = copySlice(rules)
	if err := validateProfileContent("ProfileContent.WithNormalizationRules", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithAggregationRules returns a copy of c with its Aggregation Rules set to
// exactly the values given, in the order given. Passing an empty or nil slice
// declares none. The result is re-validated through the shared path.
func (c ProfileContent) WithAggregationRules(rules []AggregationRule) (ProfileContent, error) {
	c.aggregationRules = copySlice(rules)
	if err := validateProfileContent("ProfileContent.WithAggregationRules", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithSubjects returns a copy of c with the exact Subjects the Profile
// Revision applies to set, in the order given. A zero-value element is
// rejected. Passing an empty or nil slice declares none, leaving the
// mandatory applicability as the sole statement of when the Revision applies.
//
// These are Subjects the Profile applies *to*. They are not Quality Claim
// subjects and they are never criteria: a Profile Revision's applicability
// list and a Claim's subject are different things, and this package provides
// no conversion between core.EngineeringSubjectRef and core.CriterionRef --
// nor could it, since those two core types have no conversion path in either
// direction.
func (c ProfileContent) WithSubjects(subjects []core.EngineeringSubjectRef) (ProfileContent, error) {
	c.subjects = copySlice(subjects)
	if err := validateProfileContent("ProfileContent.WithSubjects", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithSubjectTypes returns a copy of c with the Subject types the Profile
// Revision applies to set, in the order given. A zero-value element is
// rejected. Passing an empty or nil slice declares none.
//
// Subject types are carried as open core.VocabularyValue rather than as
// core.ArtifactType. PEOS-007 says "the Subjects or Subject types to which it
// applies" without defining any Subject-type vocabulary, and a Quality Claim
// subject need not be an Artifact -- core.EngineeringSubjectRef spans
// Requirements, Decisions, Runtime Subjects, Templates, and more. Typing this
// field as core.ArtifactType would silently assert that a Subject type is
// always an Artifact Type, which PEOS-007 does not state.
func (c ProfileContent) WithSubjectTypes(subjectTypes []core.VocabularyValue) (ProfileContent, error) {
	c.subjectTypes = copySlice(subjectTypes)
	if err := validateProfileContent("ProfileContent.WithSubjectTypes", c); err != nil {
		return ProfileContent{}, err
	}
	return c, nil
}

// WithAuthority returns a copy of c with its authority set. authority must be
// non-zero; use WithoutAuthority to clear it. PEOS-007 requires authority only
// "where required", so it is optional.
//
// An authority recorded here establishes who is accountable for this Profile
// Revision's content. It is never a Decision, and a Quality Claim recorded
// against this Profile never derives governance authority from it -- treating
// a quality outcome as governance authority is the non-conforming pattern
// "Quality Outcome Used as Governance Authority", and PEOS-004 remains the
// sole owner of Engineering Commitments.
func (c ProfileContent) WithAuthority(authority core.AuthorityRef) (ProfileContent, error) {
	if authority.IsZero() {
		return ProfileContent{}, fmt.Errorf("quality: ProfileContent.WithAuthority: %w: authority must not be zero", ErrInvalidQualityProfile)
	}
	c.authority = authority
	return c, nil
}

// WithoutAuthority returns a copy of c with its authority cleared.
func (c ProfileContent) WithoutAuthority() ProfileContent {
	c.authority = core.AuthorityRef{}
	return c
}

// WithExtension returns a copy of c with its extension data set. Passing the
// zero core.Extension is equivalent to declaring none.
func (c ProfileContent) WithExtension(extension core.Extension) ProfileContent {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c ProfileContent) WithoutExtension() ProfileContent {
	c.extension = core.Extension{}
	return c
}

// Scope returns c's declared scope. It is mandatory and therefore never
// absent on a valid ProfileContent.
func (c ProfileContent) Scope() core.Scope { return c.scope }

// Applicability returns c's declared applicability conditions. It is
// mandatory and therefore never unstated on a valid ProfileContent.
func (c ProfileContent) Applicability() ProfileApplicability { return c.applicability }

// Provenance returns c's declared provenance.
func (c ProfileContent) Provenance() core.Provenance { return c.provenance }

// Characteristics returns a defensive copy of c's Quality Characteristics, in
// declaration order. May be empty: PEOS-007 states no minimum cardinality.
func (c ProfileContent) Characteristics() []Characteristic { return copySlice(c.characteristics) }

// Measures returns a defensive copy of c's Quality Measures, in declaration
// order. May be empty: PEOS-007 states no minimum cardinality.
func (c ProfileContent) Measures() []Measure { return copySlice(c.measures) }

// Thresholds returns a defensive copy of c's Thresholds, in declaration
// order.
func (c ProfileContent) Thresholds() []Threshold { return copySlice(c.thresholds) }

// Targets returns a defensive copy of c's Targets, in declaration order.
func (c ProfileContent) Targets() []Target { return copySlice(c.targets) }

// Constraints returns a defensive copy of c's Quality Constraints, in
// declaration order.
func (c ProfileContent) Constraints() []Constraint { return copySlice(c.constraints) }

// NormalizationRules returns a defensive copy of c's Normalization Rules, in
// declaration order.
func (c ProfileContent) NormalizationRules() []NormalizationRule {
	return copySlice(c.normalizationRules)
}

// AggregationRules returns a defensive copy of c's Aggregation Rules, in
// declaration order.
func (c ProfileContent) AggregationRules() []AggregationRule {
	return copySlice(c.aggregationRules)
}

// Characteristic returns the Quality Characteristic in c whose profile-local
// key equals key, and whether one was found. The key is looked up only among
// Characteristics: because keys are unique per kind rather than across kinds,
// a lookup is always scoped to one collection, and there is deliberately no
// cross-kind lookup that could resolve a key in the wrong namespace.
func (c ProfileContent) Characteristic(key core.LocalKey) (Characteristic, bool) {
	if key.IsZero() {
		return Characteristic{}, false
	}
	for _, v := range c.characteristics {
		if v.Key() == key {
			return v, true
		}
	}
	return Characteristic{}, false
}

// Measure returns the Quality Measure in c whose profile-local key equals
// key, and whether one was found. Looked up only among Measures.
func (c ProfileContent) Measure(key core.LocalKey) (Measure, bool) {
	if key.IsZero() {
		return Measure{}, false
	}
	for _, v := range c.measures {
		if v.Key() == key {
			return v, true
		}
	}
	return Measure{}, false
}

// Threshold returns the Threshold in c whose profile-local key equals key,
// and whether one was found. Looked up only among Thresholds.
func (c ProfileContent) Threshold(key core.LocalKey) (Threshold, bool) {
	if key.IsZero() {
		return Threshold{}, false
	}
	for _, v := range c.thresholds {
		if v.Key() == key {
			return v, true
		}
	}
	return Threshold{}, false
}

// Target returns the Target in c whose profile-local key equals key, and
// whether one was found. Looked up only among Targets.
func (c ProfileContent) Target(key core.LocalKey) (Target, bool) {
	if key.IsZero() {
		return Target{}, false
	}
	for _, v := range c.targets {
		if v.Key() == key {
			return v, true
		}
	}
	return Target{}, false
}

// Constraint returns the Quality Constraint in c whose profile-local key
// equals key, and whether one was found. Looked up only among Constraints.
func (c ProfileContent) Constraint(key core.LocalKey) (Constraint, bool) {
	if key.IsZero() {
		return Constraint{}, false
	}
	for _, v := range c.constraints {
		if v.Key() == key {
			return v, true
		}
	}
	return Constraint{}, false
}

// Subjects returns a defensive copy of the exact Subjects c applies to, in
// declaration order.
func (c ProfileContent) Subjects() []core.EngineeringSubjectRef { return copySlice(c.subjects) }

// SubjectTypes returns a defensive copy of the Subject types c applies to, in
// declaration order.
func (c ProfileContent) SubjectTypes() []core.VocabularyValue { return copySlice(c.subjectTypes) }

// Authority returns c's declared authority, and whether one is set.
func (c ProfileContent) Authority() (core.AuthorityRef, bool) {
	return c.authority, !c.authority.IsZero()
}

// Extension returns c's extension data.
func (c ProfileContent) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c ProfileContent) IsZero() bool {
	return c.scope.IsZero() && c.applicability.IsZero() && c.provenance.IsZero() &&
		len(c.characteristics) == 0 && len(c.measures) == 0
}

type profileContentJSON struct {
	Scope              core.Scope                   `json:"scope"`
	Applicability      ProfileApplicability         `json:"applicability"`
	Provenance         core.Provenance              `json:"provenance"`
	Characteristics    []Characteristic             `json:"characteristics"`
	Measures           []Measure                    `json:"measures"`
	Thresholds         []Threshold                  `json:"thresholds,omitempty"`
	Targets            []Target                     `json:"targets,omitempty"`
	Constraints        []Constraint                 `json:"constraints,omitempty"`
	NormalizationRules []NormalizationRule          `json:"normalization_rules,omitempty"`
	AggregationRules   []AggregationRule            `json:"aggregation_rules,omitempty"`
	Subjects           []core.EngineeringSubjectRef `json:"subjects,omitempty"`
	SubjectTypes       []core.VocabularyValue       `json:"subject_types,omitempty"`
	Authority          *core.AuthorityRef           `json:"authority,omitempty"`
	Extension          *core.Extension              `json:"extension,omitempty"`
}

// profileContentUnmarshalJSON mirrors profileContentJSON's field set for
// decoding only, with one difference: Authority is captured as raw, undecoded
// bytes so an explicit JSON null can be distinguished from an absent key and
// rejected -- the json.RawMessage probe technique Packet D.1 established.
//
// scope, applicability, and provenance need no such treatment: an absent key
// and an explicit null both yield a zero value that validateProfileContent
// rejects, so the two cases converge on the same error and need not be told
// apart. Every collection -- characteristics and measures included -- needs
// no such treatment either, but for the opposite reason: absent, null, and []
// all denote the same valid state, "defines none of this kind", exactly as
// for thresholds, targets, constraints, normalization_rules, aggregation_rules,
// subjects, and subject_types.
type profileContentUnmarshalJSON struct {
	Scope              core.Scope                   `json:"scope"`
	Applicability      ProfileApplicability         `json:"applicability"`
	Provenance         core.Provenance              `json:"provenance"`
	Characteristics    []Characteristic             `json:"characteristics"`
	Measures           []Measure                    `json:"measures"`
	Thresholds         []Threshold                  `json:"thresholds"`
	Targets            []Target                     `json:"targets"`
	Constraints        []Constraint                 `json:"constraints"`
	NormalizationRules []NormalizationRule          `json:"normalization_rules"`
	AggregationRules   []AggregationRule            `json:"aggregation_rules"`
	Subjects           []core.EngineeringSubjectRef `json:"subjects"`
	SubjectTypes       []core.VocabularyValue       `json:"subject_types"`
	Authority          json.RawMessage              `json:"authority"`
	Extension          *core.Extension              `json:"extension,omitempty"`
}

// MarshalJSON encodes c with scope, applicability, provenance, characteristics,
// and measures always present as keys, plus whichever optional keys are set.
// Of those five, only scope, applicability, and provenance carry mandatory
// (non-empty) content -- characteristics and measures are serialized
// unconditionally (no omitempty) but may legitimately be null, exactly like
// the optional collections below them.
//
// There is no "relation", "source", "target", lifecycle, "state", "status",
// "version", "score", "quality_score", "current", "latest", "effective",
// "aggregate", "satisfied", "conformant", "compliant", "certified",
// "accepted", "basis", or "verdict" key, and no top-level PEOS type
// discriminator. Their absence is the structural proof that a Quality Profile
// Revision carries configuration only -- never a relationship, never a
// lifecycle, and never stored derived quality state.
func (c ProfileContent) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("quality: marshal ProfileContent: %w", ErrInvalidQualityProfile)
	}
	raw := profileContentJSON{
		Scope:              c.scope,
		Applicability:      c.applicability,
		Provenance:         c.provenance,
		Characteristics:    c.characteristics,
		Measures:           c.measures,
		Thresholds:         c.thresholds,
		Targets:            c.targets,
		Constraints:        c.constraints,
		NormalizationRules: c.normalizationRules,
		AggregationRules:   c.aggregationRules,
		Subjects:           c.subjects,
		SubjectTypes:       c.subjectTypes,
	}
	if !c.authority.IsZero() {
		raw.Authority = &c.authority
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation as
// NewProfileContent and each With* method, so a decoded ProfileContent can
// never be constructor-impossible. The receiver is left untouched unless every
// check passes.
//
// Every collection is assembled before validation runs, exactly once, through
// the same shared path -- not applied one modifier at a time. That matters for
// normalization_rules: a Measure may reference a Normalization Rule, so
// validating an intermediate state that had the Measures but not yet the Rules
// would reject a document that is in fact valid.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - scope, provenance: a missing key leaves the field zero and is rejected
//     through its owning sentinel (core.ErrInvalidScope,
//     ErrInvalidQualityProfile). An explicit null invokes that nested type's
//     own UnmarshalJSON, which fails there, so the error is wrapped here with
//     ErrInvalidQualityProfile instead. Both are rejected; the sentinel sets
//     differ.
//   - applicability: both a missing key and an explicit null yield an empty
//     kind, rejected with ErrInvalidProfileApplicability.
//   - characteristics, measures, thresholds, targets, constraints,
//     normalization_rules, aggregation_rules, subjects, subject_types: absent,
//     explicit null, and empty array are all equivalent and all mean "defines
//     none of this kind" -- PEOS-007 states no minimum cardinality for any of
//     them. A non-empty collection is still validated element by element, and
//     the conditional references a Measure, Threshold, or Target carries must
//     still resolve.
//   - authority: a missing key means absent; an explicit null is rejected
//     rather than silently treated as absent.
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract.
func (c *ProfileContent) UnmarshalJSON(data []byte) error {
	var raw profileContentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal ProfileContent: %w: %w", ErrInvalidQualityProfile, err)
	}

	result := ProfileContent{
		scope:              raw.Scope,
		applicability:      raw.Applicability,
		provenance:         raw.Provenance,
		characteristics:    copySlice(raw.Characteristics),
		measures:           copySlice(raw.Measures),
		normalizationRules: copySlice(raw.NormalizationRules),
		thresholds:         copySlice(raw.Thresholds),
		targets:            copySlice(raw.Targets),
		constraints:        copySlice(raw.Constraints),
		aggregationRules:   copySlice(raw.AggregationRules),
		subjects:           copySlice(raw.Subjects),
		subjectTypes:       copySlice(raw.SubjectTypes),
	}
	if err := validateProfileContent("unmarshal ProfileContent", result); err != nil {
		return err
	}

	if len(raw.Authority) > 0 {
		if string(raw.Authority) == "null" {
			return fmt.Errorf("quality: unmarshal ProfileContent: %w: authority must not be null", ErrInvalidQualityProfile)
		}
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("quality: unmarshal ProfileContent: %w: %w", ErrInvalidQualityProfile, err)
		}
		var err error
		if result, err = result.WithAuthority(authority); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*c = result
	return nil
}

// --- ProfileRevision ---------------------------------------------------------

// ProfileRevision is shorthand for "an Artifact Revision whose Artifact is a
// Quality Profile" -- not a separate PEOS entity, and not a Quality Profile
// Version. It composes core.ArtifactRevision by named field, per the
// specialized-Revision strategy core.ArtifactRevision itself documents and
// requirement.Revision and validation.PlanRevision already follow, and pairs
// it with typed ProfileContent.
//
// ProfileRevision is immutable and exposes no WithContent: "Modification of a
// Quality Profile's content constitutes a content change and SHALL create a
// new Artifact Revision in accordance with PEOS-002", so a new
// ProfileRevision is constructed rather than an existing one edited. This is
// also what makes a Characteristic's or Measure's meaning unable to "change
// silently within the same Profile Revision".
type ProfileRevision struct {
	core    core.ArtifactRevision
	content ProfileContent
}

// newProfileRevisionFromParts validates revision and content without
// reference to any Profile, and is the path both NewProfileRevision and
// UnmarshalJSON share. It cannot, and does not attempt to, check that
// revision belongs to any particular Profile -- see NewProfileRevision and
// UnmarshalJSON for why that check needs a Profile value a ProfileRevision's
// own JSON does not carry.
func newProfileRevisionFromParts(revision core.ArtifactRevision, content ProfileContent) (ProfileRevision, error) {
	if revision.IsZero() {
		return ProfileRevision{}, fmt.Errorf("%w: core revision must not be zero", ErrInvalidQualityProfile)
	}
	if content.IsZero() {
		return ProfileRevision{}, fmt.Errorf("%w: profile content must not be zero", ErrInvalidQualityProfile)
	}
	return ProfileRevision{core: revision, content: content}, nil
}

// NewProfileRevision validates profile, revision, and content and returns a
// ProfileRevision. profile and revision must both be non-zero, content must
// be non-zero, and revision.ArtifactID() must equal profile.ID().
func NewProfileRevision(profile Profile, revision core.ArtifactRevision, content ProfileContent) (ProfileRevision, error) {
	if profile.IsZero() {
		return ProfileRevision{}, fmt.Errorf("quality: NewProfileRevision: %w: profile must not be zero", ErrInvalidQualityProfile)
	}
	result, err := newProfileRevisionFromParts(revision, content)
	if err != nil {
		return ProfileRevision{}, fmt.Errorf("quality: NewProfileRevision: %w", err)
	}
	if revision.ArtifactID() != profile.ID() {
		return ProfileRevision{}, fmt.Errorf("quality: NewProfileRevision: %w", ErrQualityProfileArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the ProfileRevision's underlying core.ArtifactRevision.
func (r ProfileRevision) Core() core.ArtifactRevision { return r.core }

// Content returns the ProfileRevision's typed Quality Profile content.
func (r ProfileRevision) Content() ProfileContent { return r.content }

// Ref returns a core.ArtifactRevisionRef identifying r. This is the reference
// a Quality Claim criterion pairs with a profile-local key to name one exact
// Characteristic, Measure, Threshold, Target, or Quality Constraint -- see
// core.NewQualityElementCriterionRef.
func (r ProfileRevision) Ref() (core.ArtifactRevisionRef, error) {
	return core.NewArtifactRevisionRef(r.core.ArtifactID(), r.core.RevisionID())
}

// IsZero reports whether r is the zero value.
func (r ProfileRevision) IsZero() bool { return r.core.IsZero() && r.content.IsZero() }

type profileRevisionJSON struct {
	Core    core.ArtifactRevision `json:"core"`
	Content ProfileContent        `json:"content"`
}

// MarshalJSON encodes r as {"core":{...},"content":{...}}, per the
// nested-composition strategy core.ArtifactRevision documents.
func (r ProfileRevision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("quality: marshal ProfileRevision: %w", ErrInvalidQualityProfile)
	}
	return json.Marshal(profileRevisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core":{...},"content":{...}} JSON
// form.
//
// This reconstructs r.core and r.content via the same checks
// newProfileRevisionFromParts (and therefore NewProfileRevision) applies, but
// cannot repeat NewProfileRevision's ArtifactID-to-Profile cross-check: a
// ProfileRevision's own JSON carries only its core.ArtifactRevision (with a
// bare ArtifactID) and its ProfileContent, never a core.Artifact with an
// ArtifactType to check that ArtifactID against. This is the same limitation
// core.ArtifactRevision, requirement.Revision, and validation.PlanRevision
// already document. A ProfileRevision decoded on its own is therefore a value
// encoding of its own fields, not a self-sufficient record of "this Revision
// of that Quality Profile" without external context supplying the matching
// Profile.
func (r *ProfileRevision) UnmarshalJSON(data []byte) error {
	var raw profileRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal ProfileRevision: %w: %w", ErrInvalidQualityProfile, err)
	}
	result, err := newProfileRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return fmt.Errorf("quality: unmarshal ProfileRevision: %w", err)
	}
	*r = result
	return nil
}
