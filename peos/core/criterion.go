package core

import (
	"encoding/json"
	"fmt"
)

// This file defines CriterionRef, the tagged union used for a Claim's
// criteria (PEOS-006 Claim Criteria), and the small combinator reference
// types it needs for criterion kinds whose owning construct (Quality
// Profile, Runtime Contract, Template) is not modeled until a later
// packet.
//
// CriterionRef is deliberately not EngineeringSubjectRef, and no
// construction path converts one into the other. Keeping them separate
// lets a future validator check, at the type level, that a Claim's
// criteria never smuggle in a second subject (PEOS-006's "a criterion is
// not a second Claim Subject" rule) and that a Requirement used as a
// Satisfaction Claim criterion is never simultaneously that Claim's
// subject (the BL-1 locked resolution).
//
// Some criterion kinds reference an owned-value element (a Quality
// Characteristic or Measure, a Runtime Assertion, a Template Parameter
// Constraint) that PEOS-000-009 scope to their owning Artifact Revision
// with a local key (PEOS Reference Meta-Model Blueprint SS5). Packet A
// does not implement those owning constructs (Quality Profile, Runtime
// Contract, Template are Packet B/F/G/H work); the combinator types below
// reference such an element only as "the owning Revision, plus the local
// key naming the element within it," using LocalKey and the Revision-ref
// types already defined in reference.go, rather than inventing a
// standalone entity type for the element itself.

// QualityElementCriterionRef references a Quality Profile Revision-owned
// element by naming its owning Quality Profile Revision and its local key
// within that Revision. Quality Profile is an Artifact (PEOS-007); this
// package does not define a dedicated QualityProfileRevisionRef type, so
// the owning Revision is referenced with the general-purpose
// ArtifactRevisionRef.
//
// The element it identifies may be any of the five Profile Revision-owned
// kinds PEOS-007 permits as Quality Claim criteria:
//
//   - Quality Characteristic (CriterionKindQualityCharacteristic);
//   - Quality Measure (CriterionKindQualityMeasure);
//   - Threshold (CriterionKindQualityThreshold);
//   - Target (CriterionKindQualityTarget);
//   - Quality Constraint (CriterionKindQualityConstraint).
//
// All five share this exact payload shape because PEOS-007 scopes all five
// the same way: each "has no independent identity outside its owning
// Profile Revision", so (Profile Revision, local key) is the complete
// reference in every case. The criterion kind, not the payload, is what
// distinguishes them -- which is also why a Threshold reference can never
// be silently read as a Characteristic reference: the discriminator is
// checked before the payload is handed back.
//
// This type is deliberately not a union over the five element kinds. Its
// local key is meaningful only within its owning Profile Revision, and
// PEOS-007 states no uniqueness rule across element kinds, so the same key
// may legitimately name a Characteristic and a Threshold in one Profile
// Revision. The owning kind therefore has to come from the criterion
// discriminator rather than from the key.
type QualityElementCriterionRef struct {
	profile ArtifactRevisionRef
	element LocalKey
}

func NewQualityElementCriterionRef(profile ArtifactRevisionRef, element LocalKey) (QualityElementCriterionRef, error) {
	if profile.IsZero() {
		return QualityElementCriterionRef{}, fmt.Errorf("core: NewQualityElementCriterionRef: %w", ErrEmptyIdentity)
	}
	if element.IsZero() {
		return QualityElementCriterionRef{}, fmt.Errorf("core: NewQualityElementCriterionRef: %w", ErrEmptyIdentity)
	}
	return QualityElementCriterionRef{profile: profile, element: element}, nil
}

func (r QualityElementCriterionRef) Profile() ArtifactRevisionRef { return r.profile }
func (r QualityElementCriterionRef) Element() LocalKey            { return r.element }
func (r QualityElementCriterionRef) IsZero() bool                 { return r.profile.IsZero() && r.element.IsZero() }

type qualityElementCriterionRefJSON struct {
	Profile ArtifactRevisionRef `json:"profile"`
	Element LocalKey            `json:"element"`
}

func (r QualityElementCriterionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(qualityElementCriterionRefJSON{Profile: r.profile, Element: r.element})
}

func (r *QualityElementCriterionRef) UnmarshalJSON(data []byte) error {
	var raw qualityElementCriterionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal QualityElementCriterionRef: %w", err)
	}
	v, err := NewQualityElementCriterionRef(raw.Profile, raw.Element)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeRuleCriterionRef references a Runtime Contract rule or a Runtime
// Assertion by naming its owning Runtime Contract Revision and its local
// key within that Revision.
type RuntimeRuleCriterionRef struct {
	contract RuntimeContractRevisionRef
	rule     LocalKey
}

func NewRuntimeRuleCriterionRef(contract RuntimeContractRevisionRef, rule LocalKey) (RuntimeRuleCriterionRef, error) {
	if contract.IsZero() {
		return RuntimeRuleCriterionRef{}, fmt.Errorf("core: NewRuntimeRuleCriterionRef: %w", ErrEmptyIdentity)
	}
	if rule.IsZero() {
		return RuntimeRuleCriterionRef{}, fmt.Errorf("core: NewRuntimeRuleCriterionRef: %w", ErrEmptyIdentity)
	}
	return RuntimeRuleCriterionRef{contract: contract, rule: rule}, nil
}

func (r RuntimeRuleCriterionRef) Contract() RuntimeContractRevisionRef { return r.contract }
func (r RuntimeRuleCriterionRef) Rule() LocalKey                       { return r.rule }
func (r RuntimeRuleCriterionRef) IsZero() bool                         { return r.contract.IsZero() && r.rule.IsZero() }

type runtimeRuleCriterionRefJSON struct {
	Contract RuntimeContractRevisionRef `json:"contract"`
	Rule     LocalKey                   `json:"rule"`
}

func (r RuntimeRuleCriterionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeRuleCriterionRefJSON{Contract: r.contract, Rule: r.rule})
}

func (r *RuntimeRuleCriterionRef) UnmarshalJSON(data []byte) error {
	var raw runtimeRuleCriterionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeRuleCriterionRef: %w", err)
	}
	v, err := NewRuntimeRuleCriterionRef(raw.Contract, raw.Rule)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// TemplateConstraintCriterionRef references a Template Parameter
// Constraint, or another Template-owned constraint, by naming its owning
// Template Artifact Revision and its local key within that Revision.
type TemplateConstraintCriterionRef struct {
	template   TemplateArtifactRevisionRef
	constraint LocalKey
}

func NewTemplateConstraintCriterionRef(template TemplateArtifactRevisionRef, constraint LocalKey) (TemplateConstraintCriterionRef, error) {
	if template.IsZero() {
		return TemplateConstraintCriterionRef{}, fmt.Errorf("core: NewTemplateConstraintCriterionRef: %w", ErrEmptyIdentity)
	}
	if constraint.IsZero() {
		return TemplateConstraintCriterionRef{}, fmt.Errorf("core: NewTemplateConstraintCriterionRef: %w", ErrEmptyIdentity)
	}
	return TemplateConstraintCriterionRef{template: template, constraint: constraint}, nil
}

func (r TemplateConstraintCriterionRef) Template() TemplateArtifactRevisionRef { return r.template }
func (r TemplateConstraintCriterionRef) Constraint() LocalKey                  { return r.constraint }
func (r TemplateConstraintCriterionRef) IsZero() bool {
	return r.template.IsZero() && r.constraint.IsZero()
}

type templateConstraintCriterionRefJSON struct {
	Template   TemplateArtifactRevisionRef `json:"template"`
	Constraint LocalKey                    `json:"constraint"`
}

func (r TemplateConstraintCriterionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(templateConstraintCriterionRefJSON{Template: r.template, Constraint: r.constraint})
}

func (r *TemplateConstraintCriterionRef) UnmarshalJSON(data []byte) error {
	var raw templateConstraintCriterionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal TemplateConstraintCriterionRef: %w", err)
	}
	v, err := NewTemplateConstraintCriterionRef(raw.Template, raw.Constraint)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ProductRuleRef references a Product contract rule by its namespaced
// vocabulary identifier. It carries no independent PEOS identity and no
// lifecycle; it is a plain, immutable descriptor. ProductRuleRef and
// ExternalRuleRef are deliberately distinct Go types, even though both
// wrap a VocabularyValue, so that a caller cannot pass one where the
// other is expected without an explicit, visible conversion at the call
// site (there is none: each has its own field name, so neither is even
// explicitly convertible to the other).
type ProductRuleRef struct{ productRuleValue VocabularyValue }

// NewProductRuleRef validates value and returns a ProductRuleRef.
func NewProductRuleRef(value VocabularyValue) (ProductRuleRef, error) {
	if value.IsZero() {
		return ProductRuleRef{}, fmt.Errorf("core: NewProductRuleRef: %w", ErrEmptyIdentity)
	}
	return ProductRuleRef{productRuleValue: value}, nil
}

func (r ProductRuleRef) VocabularyValue() VocabularyValue { return r.productRuleValue }
func (r ProductRuleRef) String() string                   { return r.productRuleValue.String() }
func (r ProductRuleRef) IsZero() bool                     { return r.productRuleValue.IsZero() }

func (r ProductRuleRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.productRuleValue)
}

func (r *ProductRuleRef) UnmarshalJSON(data []byte) error {
	var v VocabularyValue
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("core: unmarshal ProductRuleRef: %w", err)
	}
	parsed, err := NewProductRuleRef(v)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

// ExternalRuleRef references another explicitly defined, non-PEOS
// evaluation rule (for example, an external standard's criterion) by its
// namespaced vocabulary identifier. Same shape and guarantees as
// ProductRuleRef; kept as a separate type for the same reason.
type ExternalRuleRef struct{ externalRuleValue VocabularyValue }

// NewExternalRuleRef validates value and returns an ExternalRuleRef.
func NewExternalRuleRef(value VocabularyValue) (ExternalRuleRef, error) {
	if value.IsZero() {
		return ExternalRuleRef{}, fmt.Errorf("core: NewExternalRuleRef: %w", ErrEmptyIdentity)
	}
	return ExternalRuleRef{externalRuleValue: value}, nil
}

func (r ExternalRuleRef) VocabularyValue() VocabularyValue { return r.externalRuleValue }
func (r ExternalRuleRef) String() string                   { return r.externalRuleValue.String() }
func (r ExternalRuleRef) IsZero() bool                     { return r.externalRuleValue.IsZero() }

func (r ExternalRuleRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.externalRuleValue)
}

func (r *ExternalRuleRef) UnmarshalJSON(data []byte) error {
	var v VocabularyValue
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("core: unmarshal ExternalRuleRef: %w", err)
	}
	parsed, err := NewExternalRuleRef(v)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

// Known discriminator values for CriterionRef.
const (
	CriterionKindRequirement           = "requirement"
	CriterionKindRequirementRevision   = "requirement_revision"
	CriterionKindArtifact              = "artifact"
	CriterionKindArtifactRevision      = "artifact_revision"
	CriterionKindQualityCharacteristic = "quality_characteristic"
	CriterionKindQualityMeasure        = "quality_measure"
	CriterionKindQualityThreshold      = "quality_threshold"
	CriterionKindQualityTarget         = "quality_target"
	CriterionKindQualityConstraint     = "quality_constraint"
	CriterionKindRuntimeContractRule   = "runtime_contract_rule"
	CriterionKindRuntimeAssertion      = "runtime_assertion"
	CriterionKindTemplateConstraint    = "template_constraint"
	CriterionKindProductRule           = "product_rule"
	CriterionKindExternalRule          = "external_rule"
)

var knownCriterionKinds = map[string]bool{
	CriterionKindRequirement:           true,
	CriterionKindRequirementRevision:   true,
	CriterionKindArtifact:              true,
	CriterionKindArtifactRevision:      true,
	CriterionKindQualityCharacteristic: true,
	CriterionKindQualityMeasure:        true,
	CriterionKindQualityThreshold:      true,
	CriterionKindQualityTarget:         true,
	CriterionKindQualityConstraint:     true,
	CriterionKindRuntimeContractRule:   true,
	CriterionKindRuntimeAssertion:      true,
	CriterionKindTemplateConstraint:    true,
	CriterionKindProductRule:           true,
	CriterionKindExternalRule:          true,
}

// OpaqueCriterion carries the discriminator and namespaced identity of a
// CriterionRef whose kind this packet does not give a dedicated typed
// payload to.
type OpaqueCriterion struct {
	kind       string
	namespace  string
	identifier string
}

func (o OpaqueCriterion) Kind() string       { return o.kind }
func (o OpaqueCriterion) Namespace() string  { return o.namespace }
func (o OpaqueCriterion) Identifier() string { return o.identifier }

// CriterionRef is the tagged union used for a Validation/Quality/
// Compliance/Template Conformance Claim's criteria (PEOS-006 Claim
// Criteria; PEOS Reference Meta-Model Blueprint SS9). It is a distinct
// type from EngineeringSubjectRef; see the package-level comment above
// for why the separation matters.
type CriterionRef struct {
	kind  string
	known bool

	requirement           RequirementRef
	requirementRevision   RequirementArtifactRevisionRef
	artifact              ArtifactRef
	artifactRevision      ArtifactRevisionRef
	qualityCharacteristic QualityElementCriterionRef
	qualityMeasure        QualityElementCriterionRef
	qualityThreshold      QualityElementCriterionRef
	qualityTarget         QualityElementCriterionRef
	qualityConstraint     QualityElementCriterionRef
	runtimeContractRule   RuntimeRuleCriterionRef
	runtimeAssertion      RuntimeRuleCriterionRef
	templateConstraint    TemplateConstraintCriterionRef
	productRule           ProductRuleRef
	externalRule          ExternalRuleRef

	opaque OpaqueCriterion
}

func (c CriterionRef) Kind() string  { return c.kind }
func (c CriterionRef) IsKnown() bool { return c.known }
func (c CriterionRef) IsZero() bool  { return c.kind == "" }

func CriterionRefFromRequirement(ref RequirementRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromRequirement: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindRequirement, known: true, requirement: ref}, nil
}

func (c CriterionRef) AsRequirement() (RequirementRef, bool) {
	if c.kind != CriterionKindRequirement {
		return RequirementRef{}, false
	}
	return c.requirement, true
}

func CriterionRefFromRequirementRevision(ref RequirementArtifactRevisionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromRequirementRevision: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindRequirementRevision, known: true, requirementRevision: ref}, nil
}

func (c CriterionRef) AsRequirementRevision() (RequirementArtifactRevisionRef, bool) {
	if c.kind != CriterionKindRequirementRevision {
		return RequirementArtifactRevisionRef{}, false
	}
	return c.requirementRevision, true
}

func CriterionRefFromArtifact(ref ArtifactRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromArtifact: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindArtifact, known: true, artifact: ref}, nil
}

func (c CriterionRef) AsArtifact() (ArtifactRef, bool) {
	if c.kind != CriterionKindArtifact {
		return ArtifactRef{}, false
	}
	return c.artifact, true
}

func CriterionRefFromArtifactRevision(ref ArtifactRevisionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromArtifactRevision: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindArtifactRevision, known: true, artifactRevision: ref}, nil
}

func (c CriterionRef) AsArtifactRevision() (ArtifactRevisionRef, bool) {
	if c.kind != CriterionKindArtifactRevision {
		return ArtifactRevisionRef{}, false
	}
	return c.artifactRevision, true
}

func CriterionRefFromQualityCharacteristic(ref QualityElementCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromQualityCharacteristic: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindQualityCharacteristic, known: true, qualityCharacteristic: ref}, nil
}

func (c CriterionRef) AsQualityCharacteristic() (QualityElementCriterionRef, bool) {
	if c.kind != CriterionKindQualityCharacteristic {
		return QualityElementCriterionRef{}, false
	}
	return c.qualityCharacteristic, true
}

func CriterionRefFromQualityMeasure(ref QualityElementCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromQualityMeasure: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindQualityMeasure, known: true, qualityMeasure: ref}, nil
}

func (c CriterionRef) AsQualityMeasure() (QualityElementCriterionRef, bool) {
	if c.kind != CriterionKindQualityMeasure {
		return QualityElementCriterionRef{}, false
	}
	return c.qualityMeasure, true
}

// CriterionRefFromQualityThreshold constructs a CriterionRef identifying a
// Threshold owned by an exact Quality Profile Revision. PEOS-007 lists
// Threshold among the criteria a Quality Claim may cite, and a Threshold
// "has no independent identity outside its owning Profile Revision", so
// the reference names that Revision plus the Threshold's local key within
// it.
func CriterionRefFromQualityThreshold(ref QualityElementCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromQualityThreshold: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindQualityThreshold, known: true, qualityThreshold: ref}, nil
}

func (c CriterionRef) AsQualityThreshold() (QualityElementCriterionRef, bool) {
	if c.kind != CriterionKindQualityThreshold {
		return QualityElementCriterionRef{}, false
	}
	return c.qualityThreshold, true
}

// CriterionRefFromQualityTarget constructs a CriterionRef identifying a
// Target owned by an exact Quality Profile Revision.
//
// Target and Threshold are separate criterion kinds because PEOS-007
// requires it: "a Target expresses intent, while a Threshold expresses the
// boundary used for a Claim outcome. The two SHALL NOT be conflated." They
// share a payload shape but never a discriminator, so a Target can never be
// read back through AsQualityThreshold.
func CriterionRefFromQualityTarget(ref QualityElementCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromQualityTarget: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindQualityTarget, known: true, qualityTarget: ref}, nil
}

func (c CriterionRef) AsQualityTarget() (QualityElementCriterionRef, bool) {
	if c.kind != CriterionKindQualityTarget {
		return QualityElementCriterionRef{}, false
	}
	return c.qualityTarget, true
}

// CriterionRefFromQualityConstraint constructs a CriterionRef identifying a
// Quality Constraint owned by an exact Quality Profile Revision.
//
// This is the Profile Revision-owned Quality Constraint, not a Requirement.
// PEOS-007 permits a Quality Constraint to "also be represented as a
// Requirement" by explicit modeling choice; when it is, that Requirement is
// cited through CriterionKindRequirement or
// CriterionKindRequirementRevision instead. The two paths stay distinct so
// that a Profile-owned constraint is never silently treated as a
// Requirement ("Every Quality Constraint SHALL NOT be silently treated as a
// Requirement merely because it constrains engineering behavior").
func CriterionRefFromQualityConstraint(ref QualityElementCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromQualityConstraint: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindQualityConstraint, known: true, qualityConstraint: ref}, nil
}

func (c CriterionRef) AsQualityConstraint() (QualityElementCriterionRef, bool) {
	if c.kind != CriterionKindQualityConstraint {
		return QualityElementCriterionRef{}, false
	}
	return c.qualityConstraint, true
}

func CriterionRefFromRuntimeContractRule(ref RuntimeRuleCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromRuntimeContractRule: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindRuntimeContractRule, known: true, runtimeContractRule: ref}, nil
}

func (c CriterionRef) AsRuntimeContractRule() (RuntimeRuleCriterionRef, bool) {
	if c.kind != CriterionKindRuntimeContractRule {
		return RuntimeRuleCriterionRef{}, false
	}
	return c.runtimeContractRule, true
}

func CriterionRefFromRuntimeAssertion(ref RuntimeRuleCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromRuntimeAssertion: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindRuntimeAssertion, known: true, runtimeAssertion: ref}, nil
}

func (c CriterionRef) AsRuntimeAssertion() (RuntimeRuleCriterionRef, bool) {
	if c.kind != CriterionKindRuntimeAssertion {
		return RuntimeRuleCriterionRef{}, false
	}
	return c.runtimeAssertion, true
}

func CriterionRefFromTemplateConstraint(ref TemplateConstraintCriterionRef) (CriterionRef, error) {
	if ref.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromTemplateConstraint: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindTemplateConstraint, known: true, templateConstraint: ref}, nil
}

func (c CriterionRef) AsTemplateConstraint() (TemplateConstraintCriterionRef, bool) {
	if c.kind != CriterionKindTemplateConstraint {
		return TemplateConstraintCriterionRef{}, false
	}
	return c.templateConstraint, true
}

func CriterionRefFromProductRule(rule ProductRuleRef) (CriterionRef, error) {
	if rule.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromProductRule: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindProductRule, known: true, productRule: rule}, nil
}

func (c CriterionRef) AsProductRule() (ProductRuleRef, bool) {
	if c.kind != CriterionKindProductRule {
		return ProductRuleRef{}, false
	}
	return c.productRule, true
}

func CriterionRefFromExternalRule(rule ExternalRuleRef) (CriterionRef, error) {
	if rule.IsZero() {
		return CriterionRef{}, fmt.Errorf("core: CriterionRefFromExternalRule: %w", ErrInvalidPayload)
	}
	return CriterionRef{kind: CriterionKindExternalRule, known: true, externalRule: rule}, nil
}

func (c CriterionRef) AsExternalRule() (ExternalRuleRef, bool) {
	if c.kind != CriterionKindExternalRule {
		return ExternalRuleRef{}, false
	}
	return c.externalRule, true
}

// NewOpaqueCriterionRef constructs a forward-compatible CriterionRef for
// a kind this packet does not give a typed payload to.
//
// Opaque preservation supports exactly one shape: a namespaced scalar
// reference, carried as the pair (namespace, identifier). It does not
// support composite references — the same limitation documented on
// NewOpaqueEngineeringSubjectRef applies here. Notably, a future
// criterion kind shaped like this package's own
// QualityElementCriterionRef/RuntimeRuleCriterionRef/
// TemplateConstraintCriterionRef (an owning Artifact Revision reference
// plus a LocalKey) cannot round-trip through this opaque path; it
// requires an additive, dedicated kind added to this file, not just a
// caller passing more values into namespace/identifier.
//
// CriterionKindQualityThreshold, CriterionKindQualityTarget, and
// CriterionKindQualityConstraint are exactly that case, resolved: PEOS-007
// requires all three as Quality Claim criteria, all three are
// (Profile Revision, local key) composites, and all three were therefore
// added here as dedicated kinds reusing QualityElementCriterionRef rather
// than being pushed through this constructor. Adding them changed no
// existing kind, constructor signature, or wire form.
//
// A malformed or unsupported composite payload fails explicitly during
// decode rather than being silently truncated or partially accepted; see
// CriterionRef.UnmarshalJSON's default case. No silent data loss occurs.
func NewOpaqueCriterionRef(kind, namespace, identifier string) (CriterionRef, error) {
	k, err := normalizeIdentityValue(kind)
	if err != nil {
		return CriterionRef{}, fmt.Errorf("core: NewOpaqueCriterionRef: %w", err)
	}
	if knownCriterionKinds[k] {
		return CriterionRef{}, fmt.Errorf("core: NewOpaqueCriterionRef: %q is a known kind, use its typed constructor: %w", k, ErrInvalidReferenceDiscriminator)
	}
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return CriterionRef{}, fmt.Errorf("core: NewOpaqueCriterionRef: %w", err)
	}
	id, err := normalizeIdentityValue(identifier)
	if err != nil {
		return CriterionRef{}, fmt.Errorf("core: NewOpaqueCriterionRef: %w", err)
	}
	return CriterionRef{
		kind:   k,
		known:  false,
		opaque: OpaqueCriterion{kind: k, namespace: ns, identifier: id},
	}, nil
}

func (c CriterionRef) AsOpaque() (OpaqueCriterion, bool) {
	if c.known || c.kind == "" {
		return OpaqueCriterion{}, false
	}
	return c.opaque, true
}

type criterionRefEnvelope struct {
	Kind string          `json:"kind"`
	Ref  json.RawMessage `json:"ref"`
}

func (c CriterionRef) MarshalJSON() ([]byte, error) {
	if c.kind == "" {
		return nil, fmt.Errorf("core: marshal CriterionRef: %w", ErrInvalidReferenceDiscriminator)
	}
	var (
		refBytes []byte
		err      error
	)
	switch {
	case !c.known:
		refBytes, err = json.Marshal(opaqueSubjectPayloadJSON{Namespace: c.opaque.namespace, Identifier: c.opaque.identifier})
	case c.kind == CriterionKindRequirement:
		refBytes, err = json.Marshal(c.requirement)
	case c.kind == CriterionKindRequirementRevision:
		refBytes, err = json.Marshal(c.requirementRevision)
	case c.kind == CriterionKindArtifact:
		refBytes, err = json.Marshal(c.artifact)
	case c.kind == CriterionKindArtifactRevision:
		refBytes, err = json.Marshal(c.artifactRevision)
	case c.kind == CriterionKindQualityCharacteristic:
		refBytes, err = json.Marshal(c.qualityCharacteristic)
	case c.kind == CriterionKindQualityMeasure:
		refBytes, err = json.Marshal(c.qualityMeasure)
	case c.kind == CriterionKindQualityThreshold:
		refBytes, err = json.Marshal(c.qualityThreshold)
	case c.kind == CriterionKindQualityTarget:
		refBytes, err = json.Marshal(c.qualityTarget)
	case c.kind == CriterionKindQualityConstraint:
		refBytes, err = json.Marshal(c.qualityConstraint)
	case c.kind == CriterionKindRuntimeContractRule:
		refBytes, err = json.Marshal(c.runtimeContractRule)
	case c.kind == CriterionKindRuntimeAssertion:
		refBytes, err = json.Marshal(c.runtimeAssertion)
	case c.kind == CriterionKindTemplateConstraint:
		refBytes, err = json.Marshal(c.templateConstraint)
	case c.kind == CriterionKindProductRule:
		refBytes, err = json.Marshal(c.productRule)
	case c.kind == CriterionKindExternalRule:
		refBytes, err = json.Marshal(c.externalRule)
	default:
		return nil, fmt.Errorf("core: marshal CriterionRef: %w", ErrInvalidReferenceDiscriminator)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(criterionRefEnvelope{Kind: c.kind, Ref: refBytes})
}

func (c *CriterionRef) UnmarshalJSON(data []byte) error {
	var env criterionRefEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("core: unmarshal CriterionRef: %w", err)
	}
	if env.Kind == "" {
		return fmt.Errorf("core: unmarshal CriterionRef: %w", ErrInvalidReferenceDiscriminator)
	}

	var (
		result CriterionRef
		err    error
	)
	switch env.Kind {
	case CriterionKindRequirement:
		var ref RequirementRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromRequirement(ref)
		}
	case CriterionKindRequirementRevision:
		var ref RequirementArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromRequirementRevision(ref)
		}
	case CriterionKindArtifact:
		var ref ArtifactRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromArtifact(ref)
		}
	case CriterionKindArtifactRevision:
		var ref ArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromArtifactRevision(ref)
		}
	case CriterionKindQualityCharacteristic:
		var ref QualityElementCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromQualityCharacteristic(ref)
		}
	case CriterionKindQualityMeasure:
		var ref QualityElementCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromQualityMeasure(ref)
		}
	case CriterionKindQualityThreshold:
		var ref QualityElementCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromQualityThreshold(ref)
		}
	case CriterionKindQualityTarget:
		var ref QualityElementCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromQualityTarget(ref)
		}
	case CriterionKindQualityConstraint:
		var ref QualityElementCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromQualityConstraint(ref)
		}
	case CriterionKindRuntimeContractRule:
		var ref RuntimeRuleCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromRuntimeContractRule(ref)
		}
	case CriterionKindRuntimeAssertion:
		var ref RuntimeRuleCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromRuntimeAssertion(ref)
		}
	case CriterionKindTemplateConstraint:
		var ref TemplateConstraintCriterionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromTemplateConstraint(ref)
		}
	case CriterionKindProductRule:
		var ref ProductRuleRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromProductRule(ref)
		}
	case CriterionKindExternalRule:
		var ref ExternalRuleRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = CriterionRefFromExternalRule(ref)
		}
	default:
		var payload opaqueSubjectPayloadJSON
		if err = json.Unmarshal(env.Ref, &payload); err == nil {
			result, err = NewOpaqueCriterionRef(env.Kind, payload.Namespace, payload.Identifier)
		} else {
			err = fmt.Errorf("core: unmarshal CriterionRef: unrecognized kind %q with non-opaque ref: %w", env.Kind, ErrInvalidPayload)
		}
	}
	if err != nil {
		return err
	}
	*c = result
	return nil
}
