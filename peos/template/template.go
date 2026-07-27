package template

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- shared helpers ------------------------------------------------------------

// trimmedRequired trims value and rejects it if nothing remains, attributing
// the failure to the supplied sentinel. Mirrors the identical helper in
// peos/runtime and peos/quality; duplicated rather than imported because
// peos/template imports neither.
func trimmedRequired(caller, label, value string, sentinel error) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("template: %s: %w: %s must not be empty", caller, sentinel, label)
	}
	return trimmed, nil
}

// rejectNullRaw reports an error when raw is an explicit JSON null, which
// every optional single value in this package rejects rather than silently
// treating as absent.
func rejectNullRaw(caller, label string, raw json.RawMessage, sentinel error) error {
	if string(raw) == "null" {
		return fmt.Errorf("template: unmarshal %s: %w: %s must not be null", caller, sentinel, label)
	}
	return nil
}

// copySlice returns a defensive copy of s, or nil when s is empty. Mirrors
// the identical helper in peos/runtime and peos/quality.
func copySlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	cp := make([]T, len(s))
	copy(cp, s)
	return cp
}

// --- template-local key namespaces ---------------------------------------------

// Template-local key namespaces are reference-kind namespaces, not one
// namespace per Go collection. PEOS-009 gives a Template Artifact Revision
// two independently referenced keyed collections -- parameters, named by
// defaults, parameter-targeting constraints, composition parameter mappings,
// and Template Application Records; and constraints, named by
// core.CriterionKindTemplateConstraint -- and this package's namespaces
// mirror that split exactly.
//
// This is the same principle Packet J.2.A settled for peos/runtime, applied
// to a simpler case: there, four Go collections had to share one namespace
// because core.RuntimeRuleCriterionRef carries no category discriminator;
// here, the two collections are named by genuinely different reference kinds,
// so they are genuinely separate namespaces.
const (
	kindParameter  = "parameter"
	kindConstraint = "constraint"
)

// addTemplateLocalKey records key in set, rejecting a repeat within that one
// namespace. Uniqueness is per namespace, not global: PEOS-009 states that a
// Template Parameter's key "is unique only within that exact Template
// Artifact Revision" and states no cross-collection rule at all, so the
// necessary derived rule is only that a reference naming an owned value by
// key resolves to exactly one such value within its own reference kind.
func addTemplateLocalKey(caller, kind string, set map[string]bool, key core.LocalKey) error {
	s := key.String()
	if set[s] {
		return fmt.Errorf("template: %s: %s key %q: %w", caller, kind, s, ErrDuplicateTemplateLocalKey)
	}
	set[s] = true
	return nil
}

// --- Artifact Type -------------------------------------------------------------

func mustArtifactTypeTemplate() core.ArtifactType {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, "template")
	if err != nil {
		panic("template: ArtifactTypeTemplate vocabulary value is invalid: " + err.Error())
	}
	return core.NewArtifactType(v)
}

// ArtifactTypeTemplate is the PEOS-009 Template Artifact Type. PEOS-009 does
// not itself fix an exact vocabulary string for this value -- this is an
// implementation choice, namespaced under core.PEOSNamespace because Template
// is a PEOS-000-009-defined Artifact Type rather than a Product-specific one,
// matching the convention requirement.ArtifactTypeRequirement,
// validation.ArtifactTypeValidationPlan, quality.ArtifactTypeQualityProfile,
// and runtime.ArtifactTypeRuntimeContract already established.
//
// There is exactly one Template Artifact Type. A generated Artifact is an
// ordinary Artifact of whatever type it declares -- Requirement, Validation
// Plan, Quality Profile, Runtime Contract, or anything else -- and never
// carries this type merely because a Template produced it.
var ArtifactTypeTemplate = mustArtifactTypeTemplate()

// --- Template ------------------------------------------------------------------

// Template is a PEOS-009 Template identity: a core.Artifact whose declared
// Artifact Type is ArtifactTypeTemplate ("Template SHALL be an Artifact, as
// defined by PEOS-002").
//
// Template adds no field of its own. Every declared generation element --
// permitted generated Artifact Types, parameters, defaults, constraints,
// expansion semantics, composition and specialization references,
// compatibility, applicability, provenance, and authority -- is Revision-owned
// content carried by TemplateContent, never Template identity.
//
// Template therefore has no Version field of any kind: "Template SHALL use
// ordinary Artifact Revision. This specification does not introduce
// `Template Version` or `Template Revision` as a revision system parallel to
// Artifact Revision." The phrase "Template Revision" is used in this package
// only as shorthand for "Artifact Revision whose Artifact is a Template",
// which is exactly what the TemplateRevision type below is.
//
// Template carries no compatibility state and no conformance state. Both are
// PEOS-009 derived views (see doc.go); Template.compatible and
// Template.conformant are named non-conforming patterns and exist nowhere in
// this package. Template also carries no Lifecycle State: a Template is an
// ordinary PEOS-003 Lifecycle Subject, modeled exclusively in peos/lifecycle,
// which this package does not import.
type Template struct {
	core core.Artifact
}

// NewTemplate validates artifact and returns a Template. artifact must be
// non-zero and its Type() must equal ArtifactTypeTemplate.
func NewTemplate(artifact core.Artifact) (Template, error) {
	if artifact.IsZero() {
		return Template{}, fmt.Errorf("template: NewTemplate: %w: artifact must not be zero", ErrInvalidTemplate)
	}
	if artifact.Type() != ArtifactTypeTemplate {
		return Template{}, fmt.Errorf("template: NewTemplate: %w", ErrTemplateArtifactTypeMismatch)
	}
	return Template{core: artifact}, nil
}

// Core returns the Template's underlying core.Artifact.
func (t Template) Core() core.Artifact { return t.core }

// ID returns the Template's identity.
func (t Template) ID() core.ArtifactID { return t.core.ID() }

// Ref returns a core.TemplateRef identifying t.
func (t Template) Ref() (core.TemplateRef, error) {
	return core.NewTemplateRef(t.core.ID())
}

// IsZero reports whether t is the zero value.
func (t Template) IsZero() bool { return t.core.IsZero() }

// MarshalJSON encodes t as the wire form of its underlying core.Artifact,
// with no additional envelope -- the same strategy requirement.Requirement,
// validation.Plan, quality.Profile, and runtime.Contract use.
// core.Artifact's own JSON already carries artifact_type, which both
// preserves and (on Unmarshal) lets NewTemplate re-verify that the decoded
// value is a Template.
func (t Template) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return nil, fmt.Errorf("template: marshal Template: %w", ErrInvalidTemplate)
	}
	return json.Marshal(t.core)
}

// UnmarshalJSON decodes t from its JSON form, applying the same validation as
// NewTemplate. An explicit JSON null decodes core.Artifact to its zero value,
// which NewTemplate then rejects with ErrInvalidTemplate; a decoded Template
// can never be constructor-impossible. The receiver is left untouched unless
// every check passes.
func (t *Template) UnmarshalJSON(data []byte) error {
	var artifact core.Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return fmt.Errorf("template: unmarshal Template: %w: %w", ErrInvalidTemplate, err)
	}
	result, err := NewTemplate(artifact)
	if err != nil {
		return err
	}
	*t = result
	return nil
}

// --- TemplateApplicability -----------------------------------------------------

type templateApplicabilityKind string

const (
	templateApplicabilityKindUnrestricted templateApplicabilityKind = "unrestricted"
	templateApplicabilityKindScoped       templateApplicabilityKind = "scoped"
)

// TemplateApplicability declares the conditions under which a Template
// Artifact Revision applies. PEOS-009 lists "applicability" among the items
// every Template Artifact Revision SHALL identify, with no qualifier.
// TemplateContent therefore requires it as a constructor argument and offers
// no WithApplicability/WithoutApplicability modifier.
//
// TemplateApplicability is a closed two-state discriminator whose zero value
// is invalid and represents a third, unstated state PEOS-009 does not permit.
// NewUnrestrictedTemplateApplicability constructs "no restriction" as a
// distinct, non-zero value -- this is what makes explicit unrestricted
// applicability distinguishable from an unstated one.
//
// Applicability is not a lifecycle state and not a compatibility result:
// PEOS-009 keeps all three separate, and a Template's State Assignment
// "does not... establish compatibility". This type interprets core.Scope's
// expression in no way.
//
// TemplateApplicability is deliberately not runtime.ContractApplicability,
// quality.ProfileApplicability, or validation.PlanApplicability, and is not
// converted to or from any of them. Each answers the same shaped question for
// a different owning specification; the shape is duplicated deliberately, the
// concept is not.
type TemplateApplicability struct {
	kind  templateApplicabilityKind
	scope core.Scope
}

// NewUnrestrictedTemplateApplicability returns a TemplateApplicability
// declaring explicitly that the Template Revision's applicability is not
// restricted. The returned value is non-zero: an explicit "unrestricted" is a
// stated applicability, not an absent one.
func NewUnrestrictedTemplateApplicability() TemplateApplicability {
	return TemplateApplicability{kind: templateApplicabilityKindUnrestricted}
}

// NewScopedTemplateApplicability validates scope and returns a
// TemplateApplicability bound to an explicit condition expression.
func NewScopedTemplateApplicability(scope core.Scope) (TemplateApplicability, error) {
	if scope.IsZero() {
		return TemplateApplicability{}, fmt.Errorf("template: NewScopedTemplateApplicability: %w: scope must not be zero", ErrInvalidTemplateApplicability)
	}
	return TemplateApplicability{kind: templateApplicabilityKindScoped, scope: scope}, nil
}

// Kind returns a's discriminator, "unrestricted" or "scoped". The zero value
// returns the empty string.
func (a TemplateApplicability) Kind() string { return string(a.kind) }

// IsUnrestricted reports whether a explicitly declares unrestricted
// applicability.
func (a TemplateApplicability) IsUnrestricted() bool {
	return a.kind == templateApplicabilityKindUnrestricted
}

// IsScoped reports whether a declares a scoped applicability.
func (a TemplateApplicability) IsScoped() bool { return a.kind == templateApplicabilityKindScoped }

// Scope returns a's condition expression, and whether one is set (that is,
// whether a is the scoped variant).
func (a TemplateApplicability) Scope() (core.Scope, bool) {
	if a.kind != templateApplicabilityKindScoped {
		return core.Scope{}, false
	}
	return a.scope, true
}

// IsZero reports whether a is the zero value -- the unstated state PEOS-009
// does not permit on a valid TemplateContent.
func (a TemplateApplicability) IsZero() bool { return a.kind == "" }

type templateApplicabilityJSON struct {
	Kind  string      `json:"kind"`
	Scope *core.Scope `json:"scope,omitempty"`
}

// MarshalJSON encodes a as {"kind":"unrestricted"} or {"kind":"scoped",
// "scope":{...}}. There is no top-level type discriminator beyond this
// union's own "kind".
func (a TemplateApplicability) MarshalJSON() ([]byte, error) {
	switch a.kind {
	case templateApplicabilityKindUnrestricted:
		return json.Marshal(templateApplicabilityJSON{Kind: string(templateApplicabilityKindUnrestricted)})
	case templateApplicabilityKindScoped:
		return json.Marshal(templateApplicabilityJSON{Kind: string(templateApplicabilityKindScoped), Scope: &a.scope})
	default:
		return nil, fmt.Errorf("template: marshal TemplateApplicability: %w", ErrInvalidTemplateApplicability)
	}
}

// UnmarshalJSON decodes a from its JSON form. An unrecognized or missing kind,
// an unrestricted value carrying a scope, and a scoped value missing a scope
// are all rejected. An explicit JSON null for the whole value decodes to an
// empty kind and is rejected the same way; a scoped value whose "scope" key is
// explicitly null is likewise rejected. The receiver is left untouched unless
// every check passes.
func (a *TemplateApplicability) UnmarshalJSON(data []byte) error {
	var raw templateApplicabilityJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal TemplateApplicability: %w: %w", ErrInvalidTemplateApplicability, err)
	}
	var result TemplateApplicability
	switch raw.Kind {
	case string(templateApplicabilityKindUnrestricted):
		if raw.Scope != nil {
			return fmt.Errorf("template: unmarshal TemplateApplicability: %w: unrestricted must not carry a scope", ErrInvalidTemplateApplicability)
		}
		result = NewUnrestrictedTemplateApplicability()
	case string(templateApplicabilityKindScoped):
		if raw.Scope == nil {
			return fmt.Errorf("template: unmarshal TemplateApplicability: %w: scoped requires a scope", ErrInvalidTemplateApplicability)
		}
		var err error
		result, err = NewScopedTemplateApplicability(*raw.Scope)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("template: unmarshal TemplateApplicability: unrecognized kind %q: %w", raw.Kind, ErrInvalidTemplateApplicability)
	}
	*a = result
	return nil
}

// --- TemplateContent -----------------------------------------------------------

// TemplateContent is the typed normative content PEOS-009 assigns to every
// Artifact Revision whose Artifact is a Template: the Artifact Types it may
// generate, its parameters, defaults and constraints, its expansion semantics,
// its composition and specialization references, its compatibility
// declaration, its applicability, and its provenance and authority.
//
// # The template body is not here
//
// TemplateContent carries no template body, source, schema, script,
// expression, or renderable-content field, and this is deliberate rather than
// an omission. PEOS-009 assigns the body elsewhere: "Template syntax, schema,
// rendering language, source code, natural-language text, model, or other
// representation belongs to the Template Artifact Revision, in accordance with
// PEOS-002's Artifact Representation contract", and "Representation is not
// Template identity." core.ArtifactRevision.Representations() already provides
// that, with five content forms (inline text, inline bytes, content address,
// external reference, composed). Adding a second body field here would
// duplicate PEOS-002's Representation contract and create two competing
// sources of truth for the same content.
//
// expansionSemantics is therefore a declarative descriptor of how the
// Representation is to be expanded -- not the thing being expanded, and not a
// program. PEOS-009 defines no interpolation syntax, no ordering rule, no
// conditional or loop construct, and no expression language, and its Non-Goals
// explicitly disclaim "a specific templating language or engine".
//
// # Mandatory versus optional
//
// generatedArtifactTypes, expansionSemantics, compatibility, applicability,
// and provenance are mandatory constructor arguments and are unreachable
// through any later With* call: PEOS-009 states each as a Template Artifact
// Revision SHALL-identify item without a qualifier. parameters, defaults, and
// constraints are also constructor arguments, for a different reason -- see
// NewTemplateContent.
//
// authority is optional, because PEOS-009 writes "authority, where required"
// -- the same explicitly qualified form PEOS-008 uses for a Runtime Binding
// Record's authority. This is the opposite of runtime.ContractContent, whose
// authority PEOS-008 states unqualified and which is therefore mandatory
// there. The two readings are both correct for their own specification, and
// PEOS-008's rule is deliberately not carried over.
//
// # No stored compatibility or conformance
//
// TemplateContent holds a compatibility *declaration* and never a
// compatibility *verdict*. "Current compatibility is a derived
// interpretation, computed from the applicable compatibility declarations at
// query time", and Template.compatible / TemplateRevision.compatible are named
// non-conforming patterns. Conformance is likewise derived from Template
// Conformance Claims and appears nowhere on this type.
type TemplateContent struct {
	generatedArtifactTypes []core.ArtifactType
	expansionSemantics     string
	compatibility          CompatibilityDeclaration
	applicability          TemplateApplicability
	provenance             core.Provenance

	parameters  []Parameter
	defaults    []ParameterDefault
	constraints []ParameterConstraint

	compositionReferences    []core.TemplateArtifactRevisionRef
	specializationReferences []core.TemplateArtifactRevisionRef
	authority                core.AuthorityRef
	extension                core.Extension
}

// validateTemplateContent is the single shared validation path every
// constructor, every modifier, and UnmarshalJSON converge on, so no public
// path can produce a TemplateContent another path would reject.
func validateTemplateContent(caller string, c TemplateContent) error {
	if len(c.generatedArtifactTypes) == 0 {
		return fmt.Errorf("template: %s: %w: at least one generated artifact type is required", caller, ErrInvalidTemplateContent)
	}
	seenTypes := make(map[string]bool, len(c.generatedArtifactTypes))
	for _, at := range c.generatedArtifactTypes {
		if at.IsZero() {
			return fmt.Errorf("template: %s: %w: generated artifact type must not be zero", caller, ErrInvalidTemplateContent)
		}
		s := at.String()
		if seenTypes[s] {
			return fmt.Errorf("template: %s: %w: generated artifact type %q is declared more than once", caller, ErrInvalidTemplateContent, s)
		}
		seenTypes[s] = true
	}
	if c.expansionSemantics == "" {
		return fmt.Errorf("template: %s: %w: expansion semantics must not be empty", caller, ErrInvalidTemplateContent)
	}
	if c.compatibility.IsZero() {
		return fmt.Errorf("template: %s: %w: compatibility declaration must not be zero", caller, ErrInvalidCompatibilityDeclaration)
	}
	if c.applicability.IsZero() {
		return fmt.Errorf("template: %s: %w: applicability must be explicitly stated", caller, ErrInvalidTemplateApplicability)
	}
	if c.provenance.IsZero() {
		return fmt.Errorf("template: %s: %w: provenance must not be zero", caller, ErrInvalidTemplateContent)
	}

	// The parameter namespace: every Parameter key is unique within it, and
	// it is the resolution target of every default and every
	// parameter-targeting constraint below.
	parameterKeys := make(map[string]bool, len(c.parameters))
	for _, p := range c.parameters {
		if p.IsZero() {
			return fmt.Errorf("template: %s: %w: parameter must not be zero", caller, ErrInvalidTemplateParameter)
		}
		if err := addTemplateLocalKey(caller, kindParameter, parameterKeys, p.Key()); err != nil {
			return err
		}
	}

	// The constraint namespace: separate from the parameter namespace, and the
	// sole resolution target of core.CriterionKindTemplateConstraint.
	constraintKeys := make(map[string]bool, len(c.constraints))
	for _, v := range c.constraints {
		if v.IsZero() {
			return fmt.Errorf("template: %s: %w: constraint must not be zero", caller, ErrInvalidParameterConstraint)
		}
		if err := addTemplateLocalKey(caller, kindConstraint, constraintKeys, v.Key()); err != nil {
			return err
		}
	}

	// Defaults resolve into the parameter namespace, at most one per
	// parameter, and never against a parameter that forbids default
	// resolution ("A default does not satisfy a required parameter where the
	// owning Template Artifact Revision explicitly forbids default resolution
	// for that parameter").
	defaulted := make(map[string]bool, len(c.defaults))
	for _, d := range c.defaults {
		if d.IsZero() {
			return fmt.Errorf("template: %s: %w: default must not be zero", caller, ErrInvalidParameterDefault)
		}
		key := d.Parameter().String()
		target, ok := c.parameter(d.Parameter())
		if !ok {
			return fmt.Errorf("template: %s: %w: default names parameter %q, which this revision does not declare", caller, ErrUnknownTemplateLocalKey, key)
		}
		if defaulted[key] {
			return fmt.Errorf("template: %s: %w: parameter %q has more than one default", caller, ErrInvalidParameterDefault, key)
		}
		defaulted[key] = true
		if target.ForbidsDefaultResolution() {
			return fmt.Errorf("template: %s: %w: parameter %q forbids default resolution", caller, ErrInvalidParameterDefault, key)
		}
	}

	// A parameter-targeting constraint resolves into the parameter namespace.
	// A generated-content-targeting constraint deliberately does not: PEOS-009
	// permits a constraint on "the affected parameter or generated content",
	// and generated content has no template-local key to resolve against.
	for _, v := range c.constraints {
		if paramKey, ok := v.Target().Parameter(); ok {
			if !parameterKeys[paramKey.String()] {
				return fmt.Errorf("template: %s: %w: constraint %q names parameter %q, which this revision does not declare", caller, ErrUnknownTemplateLocalKey, v.Key().String(), paramKey.String())
			}
		}
	}

	if err := validateRevisionRefs(caller, "composition reference", c.compositionReferences); err != nil {
		return err
	}
	if err := validateRevisionRefs(caller, "specialization reference", c.specializationReferences); err != nil {
		return err
	}
	return nil
}

// validateRevisionRefs rejects a zero-value or repeated exact Template
// Artifact Revision reference in one collection. PEOS-009 requires a
// composition reference to "identify the exact Template Artifact Revision",
// and a repeat within one collection would assert the same relationship
// twice without adding information.
func validateRevisionRefs(caller, label string, refs []core.TemplateArtifactRevisionRef) error {
	seen := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if ref.IsZero() {
			return fmt.Errorf("template: %s: %w: %s must not be zero", caller, ErrInvalidTemplateContent, label)
		}
		k := ref.ArtifactID().String() + "@" + ref.RevisionID().String()
		if seen[k] {
			return fmt.Errorf("template: %s: %w: %s %q is declared more than once", caller, ErrInvalidTemplateContent, label, k)
		}
		seen[k] = true
	}
	return nil
}

// NewTemplateContent validates its arguments and returns a TemplateContent
// with no composition or specialization references, no authority, and no
// extension data. Use the With* methods to add those.
//
// generatedArtifactTypes must contain at least one non-zero, non-repeated
// core.ArtifactType -- PEOS-009's one explicit minimum cardinality, since "the
// generated Artifact Type or permitted Artifact Types" cannot be satisfied by
// none. expansionSemantics must be non-empty after trimming; the trimmed value
// is stored. compatibility, applicability, and provenance must all be
// non-zero, and applicability must be explicitly stated (use
// NewUnrestrictedTemplateApplicability to declare an explicit absence of
// restriction).
//
// parameters, defaults, and constraints are constructor arguments rather than
// modifiers because their cross-references must all resolve through one shared
// validation path: a default names a parameter by key, and a
// parameter-targeting constraint names a parameter by key. With separate
// modifiers, a valid (parameters + defaults) pair would be unreachable --
// whichever call came first would be rejected for naming a key the other had
// not supplied yet. This is the same constructor-completeness rule Packet I.1
// established for quality.NewProfileContent's normalizationRules argument.
//
// All three may be empty or nil: PEOS-009 states no minimum cardinality for
// any of them, and a parameterless Template is coherent. Every slice argument
// is defensively copied; the caller may reuse or mutate its own slices
// afterward without affecting the returned value.
func NewTemplateContent(
	generatedArtifactTypes []core.ArtifactType,
	expansionSemantics string,
	compatibility CompatibilityDeclaration,
	applicability TemplateApplicability,
	provenance core.Provenance,
	parameters []Parameter,
	defaults []ParameterDefault,
	constraints []ParameterConstraint,
) (TemplateContent, error) {
	trimmedSemantics := strings.TrimSpace(expansionSemantics)
	c := TemplateContent{
		generatedArtifactTypes: copySlice(generatedArtifactTypes),
		expansionSemantics:     trimmedSemantics,
		compatibility:          compatibility,
		applicability:          applicability,
		provenance:             provenance,
		parameters:             copySlice(parameters),
		defaults:               copySlice(defaults),
		constraints:            copySlice(constraints),
	}
	if err := validateTemplateContent("NewTemplateContent", c); err != nil {
		return TemplateContent{}, err
	}
	return c, nil
}

// WithCompositionReferences returns a copy of c with its composition
// references set to exactly the exact Template Artifact Revision references
// given, in the order given. A zero-value or repeated element is rejected.
// Passing an empty or nil slice declares none -- PEOS-009 qualifies
// "composition or specialization references" with "where applicable".
//
// This records only the reference PEOS-009 requires a composing Revision to
// identify. The typed Template Composition relation -- with its participant
// levels, direction, multiplicity, cycle policy, parameter mapping rules, and
// conflict handling -- is a binary Artifact Relation and is Packet K.2's work,
// not a field here.
func (c TemplateContent) WithCompositionReferences(refs []core.TemplateArtifactRevisionRef) (TemplateContent, error) {
	c.compositionReferences = copySlice(refs)
	if err := validateTemplateContent("TemplateContent.WithCompositionReferences", c); err != nil {
		return TemplateContent{}, err
	}
	return c, nil
}

// WithSpecializationReferences returns a copy of c with its specialization
// references set to exactly the exact Template Artifact Revision references
// given, in the order given. A zero-value or repeated element is rejected.
// Passing an empty or nil slice declares none.
//
// As with composition, the typed Template Specialization relation -- with its
// inherited elements, overridden elements, and compatibility effect -- is
// Packet K.2's work.
func (c TemplateContent) WithSpecializationReferences(refs []core.TemplateArtifactRevisionRef) (TemplateContent, error) {
	c.specializationReferences = copySlice(refs)
	if err := validateTemplateContent("TemplateContent.WithSpecializationReferences", c); err != nil {
		return TemplateContent{}, err
	}
	return c, nil
}

// WithAuthority returns a copy of c with its governing authority set.
// authority must be non-zero; use WithoutAuthority to clear it.
//
// Authority is optional here because PEOS-009 writes "authority, where
// required" -- unlike runtime.ContractContent, where PEOS-008 states it
// unqualified and it is mandatory.
func (c TemplateContent) WithAuthority(authority core.AuthorityRef) (TemplateContent, error) {
	if authority.IsZero() {
		return TemplateContent{}, fmt.Errorf("template: TemplateContent.WithAuthority: %w: authority must not be zero", ErrInvalidTemplateContent)
	}
	c.authority = authority
	if err := validateTemplateContent("TemplateContent.WithAuthority", c); err != nil {
		return TemplateContent{}, err
	}
	return c, nil
}

// WithoutAuthority returns a copy of c with its governing authority cleared.
func (c TemplateContent) WithoutAuthority() TemplateContent {
	c.authority = core.AuthorityRef{}
	return c
}

// WithExtension returns a copy of c with its extension data set. Passing the
// zero core.Extension is equivalent to declaring none.
func (c TemplateContent) WithExtension(extension core.Extension) TemplateContent {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c TemplateContent) WithoutExtension() TemplateContent {
	c.extension = core.Extension{}
	return c
}

// GeneratedArtifactTypes returns a defensive copy of the Artifact Types c
// permits a generated Artifact to declare, in declaration order. Always
// non-empty on a valid TemplateContent.
//
// These are Artifact *Types*, never Artifact or Revision identities: PEOS-009
// gives a generated Artifact "its own Artifact identity, independent of the
// Template's identity", so this package stores no generated Artifact ID, no
// generated Revision ID, and no generated output of any kind.
func (c TemplateContent) GeneratedArtifactTypes() []core.ArtifactType {
	return copySlice(c.generatedArtifactTypes)
}

// ExpansionSemantics returns c's declared expansion or generation semantics,
// uninterpreted. See the type comment: this is a declarative descriptor, not
// the template body and not a program.
func (c TemplateContent) ExpansionSemantics() string { return c.expansionSemantics }

// Compatibility returns c's compatibility declaration. This is a declaration,
// never a verdict -- current compatibility is derived at query time.
func (c TemplateContent) Compatibility() CompatibilityDeclaration { return c.compatibility }

// Applicability returns c's declared applicability.
func (c TemplateContent) Applicability() TemplateApplicability { return c.applicability }

// Provenance returns c's declared provenance.
func (c TemplateContent) Provenance() core.Provenance { return c.provenance }

// Parameters returns a defensive copy of c's parameters, in declaration order.
// May be empty: PEOS-009 states no minimum, and a parameterless Template is
// coherent.
func (c TemplateContent) Parameters() []Parameter { return copySlice(c.parameters) }

// parameter is the unexported lookup validateTemplateContent uses while the
// aggregate is still being checked.
func (c TemplateContent) parameter(key core.LocalKey) (Parameter, bool) {
	if key.IsZero() {
		return Parameter{}, false
	}
	for _, p := range c.parameters {
		if p.Key() == key {
			return p, true
		}
	}
	return Parameter{}, false
}

// Parameter returns the Parameter in c whose template-local key equals key,
// and whether one was found. This resolves within the parameter namespace
// only, and is the resolution target of a ParameterDefault's parameter key and
// of a parameter-targeting ParameterConstraint.
//
// This performs local resolution only within an already-loaded
// TemplateContent; it does not load another Revision and does not verify that
// any repository holds one.
func (c TemplateContent) Parameter(key core.LocalKey) (Parameter, bool) {
	return c.parameter(key)
}

// Defaults returns a defensive copy of c's parameter defaults, in declaration
// order.
func (c TemplateContent) Defaults() []ParameterDefault { return copySlice(c.defaults) }

// Constraints returns a defensive copy of c's parameter constraints, in
// declaration order.
func (c TemplateContent) Constraints() []ParameterConstraint { return copySlice(c.constraints) }

// Constraint returns the ParameterConstraint in c whose template-local key
// equals key, and whether one was found. This resolves within the constraint
// namespace only, and is the sole resolution target of
// core.CriterionKindTemplateConstraint: a core.TemplateConstraintCriterionRef
// is a (Template Artifact Revision, LocalKey) pair, and this is what makes its
// LocalKey resolvable to exactly one constraint.
//
// Construction forbids duplicate constraint keys, so this can never face an
// ambiguous match. It performs local resolution only: it does not load the
// named Revision, does not verify that a repository holds it, and never
// evaluates the constraint's rule.
func (c TemplateContent) Constraint(key core.LocalKey) (ParameterConstraint, bool) {
	if key.IsZero() {
		return ParameterConstraint{}, false
	}
	for _, v := range c.constraints {
		if v.Key() == key {
			return v, true
		}
	}
	return ParameterConstraint{}, false
}

// CompositionReferences returns a defensive copy of the exact Template
// Artifact Revisions c composes, in declaration order.
func (c TemplateContent) CompositionReferences() []core.TemplateArtifactRevisionRef {
	return copySlice(c.compositionReferences)
}

// SpecializationReferences returns a defensive copy of the exact Template
// Artifact Revisions c specializes, in declaration order.
func (c TemplateContent) SpecializationReferences() []core.TemplateArtifactRevisionRef {
	return copySlice(c.specializationReferences)
}

// Authority returns c's governing authority, and whether one is set.
func (c TemplateContent) Authority() (core.AuthorityRef, bool) {
	return c.authority, !c.authority.IsZero()
}

// Extension returns c's extension data.
func (c TemplateContent) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c TemplateContent) IsZero() bool {
	return len(c.generatedArtifactTypes) == 0 && c.expansionSemantics == "" &&
		c.compatibility.IsZero() && c.applicability.IsZero() && c.provenance.IsZero()
}

type templateContentJSON struct {
	GeneratedArtifactTypes   []core.ArtifactType                `json:"generated_artifact_types"`
	ExpansionSemantics       string                             `json:"expansion_semantics"`
	CompatibilityDeclaration CompatibilityDeclaration           `json:"compatibility_declaration"`
	Applicability            TemplateApplicability              `json:"applicability"`
	Provenance               core.Provenance                    `json:"provenance"`
	Parameters               []Parameter                        `json:"parameters,omitempty"`
	Defaults                 []ParameterDefault                 `json:"defaults,omitempty"`
	Constraints              []ParameterConstraint              `json:"constraints,omitempty"`
	CompositionReferences    []core.TemplateArtifactRevisionRef `json:"composition_references,omitempty"`
	SpecializationReferences []core.TemplateArtifactRevisionRef `json:"specialization_references,omitempty"`
	Authority                *core.AuthorityRef                 `json:"authority,omitempty"`
	Extension                *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes c with its five mandatory keys always present, plus
// whichever optional keys are set.
//
// There is no "body", "template_body", "source", "script", "expression",
// "rendered", "instance", "current", "active", "effective", "compatible",
// "conformant", "status", "state", "lifecycle", "execution", "invocation",
// "result", "generated_artifact_id", "generated_revision_id",
// "resolved_values", "outcome", or "migration" key, and no top-level PEOS type
// discriminator. Their absence is the structural proof that a Template
// Artifact Revision carries a declared generation contract only -- never the
// template body (that is core.Representation's), never a generated output,
// never a lifecycle, and never a stored compatibility or conformance verdict.
func (c TemplateContent) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("template: marshal TemplateContent: %w", ErrInvalidTemplateContent)
	}
	raw := templateContentJSON{
		GeneratedArtifactTypes:   c.generatedArtifactTypes,
		ExpansionSemantics:       c.expansionSemantics,
		CompatibilityDeclaration: c.compatibility,
		Applicability:            c.applicability,
		Provenance:               c.provenance,
		Parameters:               c.parameters,
		Defaults:                 c.defaults,
		Constraints:              c.constraints,
		CompositionReferences:    c.compositionReferences,
		SpecializationReferences: c.specializationReferences,
	}
	if !c.authority.IsZero() {
		raw.Authority = &c.authority
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// templateContentUnmarshalJSON mirrors templateContentJSON for decoding only,
// with Authority captured as raw bytes so an explicit JSON null can be
// distinguished from an absent key and rejected -- the json.RawMessage probe
// technique Packet D.1 established. The mandatory fields need no such
// treatment: an absent key and an explicit null both yield a zero value that
// validateTemplateContent rejects, so the two cases converge on the same error
// and need not be told apart. Every optional collection needs no such
// treatment either, but for the opposite reason: absent, null, and [] all
// denote the same valid state, "declares none of this kind".
type templateContentUnmarshalJSON struct {
	GeneratedArtifactTypes   []core.ArtifactType                `json:"generated_artifact_types"`
	ExpansionSemantics       string                             `json:"expansion_semantics"`
	CompatibilityDeclaration CompatibilityDeclaration           `json:"compatibility_declaration"`
	Applicability            TemplateApplicability              `json:"applicability"`
	Provenance               core.Provenance                    `json:"provenance"`
	Parameters               []Parameter                        `json:"parameters"`
	Defaults                 []ParameterDefault                 `json:"defaults"`
	Constraints              []ParameterConstraint              `json:"constraints"`
	CompositionReferences    []core.TemplateArtifactRevisionRef `json:"composition_references"`
	SpecializationReferences []core.TemplateArtifactRevisionRef `json:"specialization_references"`
	Authority                json.RawMessage                    `json:"authority"`
	Extension                *core.Extension                    `json:"extension,omitempty"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation as
// NewTemplateContent and each With* method, so a decoded TemplateContent can
// never be constructor-impossible. The receiver is left untouched unless every
// check passes.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - generated_artifact_types: absent, null, and [] are all rejected by the
//     ≥1 minimum -- the same error, so the three need not be distinguished.
//   - expansion_semantics: absent, null, and "" all leave the field empty and
//     are rejected.
//   - compatibility_declaration, applicability, provenance: a missing key
//     leaves the field zero and is rejected through its owning sentinel. An
//     explicit null invokes that nested type's own UnmarshalJSON where it has
//     one, or leaves the field zero; both are rejected.
//   - authority: absent means "not declared" and is valid, because PEOS-009
//     qualifies it "where required". An explicit null is rejected -- it states
//     a value and supplies none.
//   - parameters, defaults, constraints, composition_references,
//     specialization_references: absent, explicit null, and empty array are
//     all equivalent and all mean "declares none of this kind".
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract.
func (c *TemplateContent) UnmarshalJSON(data []byte) error {
	var raw templateContentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal TemplateContent: %w: %w", ErrInvalidTemplateContent, err)
	}

	result := TemplateContent{
		generatedArtifactTypes:   copySlice(raw.GeneratedArtifactTypes),
		expansionSemantics:       strings.TrimSpace(raw.ExpansionSemantics),
		compatibility:            raw.CompatibilityDeclaration,
		applicability:            raw.Applicability,
		provenance:               raw.Provenance,
		parameters:               copySlice(raw.Parameters),
		defaults:                 copySlice(raw.Defaults),
		constraints:              copySlice(raw.Constraints),
		compositionReferences:    copySlice(raw.CompositionReferences),
		specializationReferences: copySlice(raw.SpecializationReferences),
	}
	if len(raw.Authority) > 0 {
		if err := rejectNullRaw("TemplateContent", "authority", raw.Authority, ErrInvalidTemplateContent); err != nil {
			return err
		}
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("template: unmarshal TemplateContent: %w: %w", ErrInvalidTemplateContent, err)
		}
		if authority.IsZero() {
			return fmt.Errorf("template: unmarshal TemplateContent: %w: authority must not be zero", ErrInvalidTemplateContent)
		}
		result.authority = authority
	}
	if err := validateTemplateContent("unmarshal TemplateContent", result); err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*c = result
	return nil
}

// --- TemplateRevision ----------------------------------------------------------

// TemplateRevision is shorthand for "an Artifact Revision whose Artifact is a
// Template" -- not a separate PEOS entity, and not a Template Version.
// PEOS-009 permits exactly this shorthand and nothing more: the phrase
// "Template Revision" "MAY be used only as informal shorthand for [an]
// Artifact Revision whose Artifact is a Template. It does not define a
// separate engineering entity, and it does not create a second revision
// mechanism."
//
// It composes core.ArtifactRevision by named field, per the
// specialized-Revision strategy requirement.Revision, validation.PlanRevision,
// quality.ProfileRevision, and runtime.ContractRevision already follow, and
// pairs it with typed TemplateContent.
//
// TemplateRevision is immutable and exposes no WithContent: "where it changes
// normative content, it produces an ordinary new Artifact Revision, per
// PEOS-002", so a new TemplateRevision is constructed rather than an existing
// one edited.
//
// TemplateRevision carries no compatibility state, no conformance state, and
// no generated-output state. The template body lives in this Revision's own
// core.ArtifactRevision.Representations(), not in its TemplateContent -- see
// TemplateContent's type comment.
type TemplateRevision struct {
	core    core.ArtifactRevision
	content TemplateContent
}

// newTemplateRevisionFromParts validates revision and content without
// reference to any Template, and is the path both NewTemplateRevision and
// UnmarshalJSON share. It cannot, and does not attempt to, check that revision
// belongs to any particular Template -- see NewTemplateRevision and
// UnmarshalJSON for why that check needs a Template value a TemplateRevision's
// own JSON does not carry.
func newTemplateRevisionFromParts(revision core.ArtifactRevision, content TemplateContent) (TemplateRevision, error) {
	if revision.IsZero() {
		return TemplateRevision{}, fmt.Errorf("%w: core revision must not be zero", ErrInvalidTemplate)
	}
	if content.IsZero() {
		return TemplateRevision{}, fmt.Errorf("%w: template content must not be zero", ErrInvalidTemplate)
	}
	return TemplateRevision{core: revision, content: content}, nil
}

// NewTemplateRevision validates template, revision, and content and returns a
// TemplateRevision. template and revision must both be non-zero, content must
// be non-zero, and revision.ArtifactID() must equal template.ID().
func NewTemplateRevision(template Template, revision core.ArtifactRevision, content TemplateContent) (TemplateRevision, error) {
	if template.IsZero() {
		return TemplateRevision{}, fmt.Errorf("template: NewTemplateRevision: %w: template must not be zero", ErrInvalidTemplate)
	}
	result, err := newTemplateRevisionFromParts(revision, content)
	if err != nil {
		return TemplateRevision{}, fmt.Errorf("template: NewTemplateRevision: %w", err)
	}
	if revision.ArtifactID() != template.ID() {
		return TemplateRevision{}, fmt.Errorf("template: NewTemplateRevision: %w", ErrTemplateArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the TemplateRevision's underlying core.ArtifactRevision. Its
// Representations() carry the template body.
func (r TemplateRevision) Core() core.ArtifactRevision { return r.core }

// Content returns the TemplateRevision's typed Template content.
func (r TemplateRevision) Content() TemplateContent { return r.content }

// Ref returns a core.TemplateArtifactRevisionRef identifying r. A Template
// Application Record is required to reference the Template Artifact Revision
// it applied using exactly this type, never the bare core.TemplateRef.
func (r TemplateRevision) Ref() (core.TemplateArtifactRevisionRef, error) {
	return core.NewTemplateArtifactRevisionRef(r.core.ArtifactID(), r.core.RevisionID())
}

// IsZero reports whether r is the zero value.
func (r TemplateRevision) IsZero() bool { return r.core.IsZero() && r.content.IsZero() }

type templateRevisionJSON struct {
	Core    core.ArtifactRevision `json:"core"`
	Content TemplateContent       `json:"content"`
}

// MarshalJSON encodes r as {"core":{...},"content":{...}}, per the
// nested-composition strategy core.ArtifactRevision documents.
func (r TemplateRevision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("template: marshal TemplateRevision: %w", ErrInvalidTemplate)
	}
	return json.Marshal(templateRevisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core":{...},"content":{...}} JSON
// form.
//
// This reconstructs r.core and r.content via the same checks
// newTemplateRevisionFromParts (and therefore NewTemplateRevision) applies,
// but cannot repeat NewTemplateRevision's ArtifactID-to-Template cross-check:
// a TemplateRevision's own JSON carries only its core.ArtifactRevision (with a
// bare ArtifactID) and its TemplateContent, never a core.Artifact with an
// ArtifactType to check that ArtifactID against. This is the same limitation
// core.ArtifactRevision, requirement.Revision, validation.PlanRevision,
// quality.ProfileRevision, and runtime.ContractRevision already document.
func (r *TemplateRevision) UnmarshalJSON(data []byte) error {
	var raw templateRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal TemplateRevision: %w: %w", ErrInvalidTemplate, err)
	}
	result, err := newTemplateRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return fmt.Errorf("template: unmarshal TemplateRevision: %w", err)
	}
	*r = result
	return nil
}
