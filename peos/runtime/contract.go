package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file implements the Runtime Contract Artifact side of PEOS-008: the
// Artifact Type, Contract identity, ContractApplicability, the
// RequirementReference union, ContractContent, and ContractRevision.
//
// Runtime binding, unbinding, observation, and violation -- and the
// Compliance Claim helper -- are Packet J.2, not this file. See doc.go for
// the full PEOS-008 ontology and the J.1/J.2 boundary.

// --- shared validation helpers -----------------------------------------------

// trimmedRequired trims value and rejects it if nothing remains, attributing
// the failure to the supplied sentinel. Mirrors the identical helper in
// peos/quality; duplicated rather than imported because peos/runtime does
// not import peos/quality.
func trimmedRequired(caller, label, value string, sentinel error) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("runtime: %s: %w: %s must not be empty", caller, sentinel, label)
	}
	return trimmed, nil
}

// rejectNullRaw reports an error when raw is an explicit JSON null, which
// every optional single value in this package rejects rather than silently
// treating as absent.
func rejectNullRaw(caller, label string, raw json.RawMessage, sentinel error) error {
	if string(raw) == "null" {
		return fmt.Errorf("runtime: unmarshal %s: %w: %s must not be null", caller, sentinel, label)
	}
	return nil
}

// copySlice returns a defensive copy of s, or nil when s is empty. Mirrors
// the identical helper in peos/quality.
func copySlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	cp := make([]T, len(s))
	copy(cp, s)
	return cp
}

// --- Artifact Type -------------------------------------------------------------

func mustArtifactTypeRuntimeContract() core.ArtifactType {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, "runtime-contract")
	if err != nil {
		panic(err)
	}
	return core.NewArtifactType(v)
}

// ArtifactTypeRuntimeContract is the PEOS-008 Runtime Contract Artifact
// Type. PEOS-008 does not itself fix an exact vocabulary string for this
// value -- this is an implementation choice, namespaced under
// core.PEOSNamespace because Runtime Contract is a PEOS-000-009-defined
// Artifact Type rather than a Product-specific one, matching the
// convention requirement.ArtifactTypeRequirement,
// validation.ArtifactTypeValidationPlan, and
// quality.ArtifactTypeQualityProfile already established.
//
// This value deliberately lives in peos/runtime, not peos/core, for the
// same reason those three live in their own packages: the Artifact Type
// belongs to the specialization that defines it.
var ArtifactTypeRuntimeContract = mustArtifactTypeRuntimeContract()

// --- Contract ------------------------------------------------------------------

// Contract is a PEOS-008 Runtime Contract identity: a core.Artifact whose
// declared Artifact Type is ArtifactTypeRuntimeContract ("A Runtime
// Contract is an Artifact as defined by PEOS-002").
//
// Contract adds no field of its own. Every declared enforcement element --
// governed Requirements, runtime subject, environment, scope,
// applicability, provenance, authority, and Assertions -- is Revision-owned
// content carried by ContractContent, never Contract identity. Contract
// therefore has no Version field of any kind: "A Runtime Contract SHALL use
// ordinary Artifact Revision for its own declared content evolution... A
// Runtime Contract SHALL NOT define an independent Runtime Contract
// Version distinct from Artifact Revision."
//
// Contract carries no runtime binding state, no deployment state, and no
// derived compliance of its own. Current Runtime Binding and Runtime
// Compliance are PEOS-008 derived views (see doc.go) and are never fields
// on this type or on core.Artifact. Contract also carries no Lifecycle
// State: a Runtime Contract MAY be governed by a Lifecycle under PEOS-003,
// modeled exclusively in peos/lifecycle, which this package does not
// import.
type Contract struct {
	core core.Artifact
}

// NewContract validates artifact and returns a Contract. artifact must be
// non-zero and its Type() must equal ArtifactTypeRuntimeContract.
func NewContract(artifact core.Artifact) (Contract, error) {
	if artifact.IsZero() {
		return Contract{}, fmt.Errorf("runtime: NewContract: %w: artifact must not be zero", ErrInvalidRuntimeContract)
	}
	if artifact.Type() != ArtifactTypeRuntimeContract {
		return Contract{}, fmt.Errorf("runtime: NewContract: %w", ErrRuntimeContractArtifactTypeMismatch)
	}
	return Contract{core: artifact}, nil
}

// Core returns the Contract's underlying core.Artifact.
func (c Contract) Core() core.Artifact { return c.core }

// ID returns the Contract's identity.
func (c Contract) ID() core.ArtifactID { return c.core.ID() }

// Ref returns a core.RuntimeContractRef identifying c.
func (c Contract) Ref() (core.RuntimeContractRef, error) {
	return core.NewRuntimeContractRef(c.core.ID())
}

// IsZero reports whether c is the zero value.
func (c Contract) IsZero() bool { return c.core.IsZero() }

// MarshalJSON encodes c as the wire form of its underlying core.Artifact,
// with no additional envelope -- the same strategy requirement.Requirement,
// validation.Plan, and quality.Profile use. core.Artifact's own JSON
// already carries artifact_type, which both preserves and (on Unmarshal)
// lets NewContract re-verify that the decoded value is a Runtime Contract.
func (c Contract) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("runtime: marshal Contract: %w", ErrInvalidRuntimeContract)
	}
	return json.Marshal(c.core)
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as NewContract. An explicit JSON null decodes core.Artifact to its zero
// value, which NewContract then rejects with ErrInvalidRuntimeContract; a
// decoded Contract can never be constructor-impossible. The receiver is
// left untouched unless every check passes.
func (c *Contract) UnmarshalJSON(data []byte) error {
	var artifact core.Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return fmt.Errorf("runtime: unmarshal Contract: %w: %w", ErrInvalidRuntimeContract, err)
	}
	result, err := NewContract(artifact)
	if err != nil {
		return err
	}
	*c = result
	return nil
}

// --- ContractApplicability -------------------------------------------------

type contractApplicabilityKind string

const (
	contractApplicabilityKindUnrestricted contractApplicabilityKind = "unrestricted"
	contractApplicabilityKindScoped       contractApplicabilityKind = "scoped"
)

// ContractApplicability declares the conditions under which a Runtime
// Contract Revision applies. PEOS-008 lists "its applicability" among the
// items every Runtime Contract Revision SHALL identify, with no qualifier.
// ContractContent therefore requires it as a constructor argument and
// offers no WithApplicability/WithoutApplicability modifier.
//
// ContractApplicability is a closed two-state discriminator whose zero
// value is invalid and represents a third, unstated state PEOS-008 does
// not permit. NewUnrestrictedContractApplicability constructs "no
// restriction" as a distinct, non-zero value -- this is what makes explicit
// unrestricted applicability distinguishable from an unstated one.
//
// ContractApplicability is deliberately not quality.ProfileApplicability,
// validation.PlanApplicability, or requirement.Applicability, and is not
// converted to or from any of them. Each answers the same shaped question
// for a different owning specification; the shape is duplicated
// deliberately (an implementation choice), the concept is not.
//
// ContractApplicability carries no identity, no revision, no lifecycle,
// and no extension. This package does not interpret core.Scope's
// expression in any way.
type ContractApplicability struct {
	kind  contractApplicabilityKind
	scope core.Scope
}

// NewUnrestrictedContractApplicability returns a ContractApplicability
// declaring explicitly that the Contract Revision's applicability is not
// restricted. The returned value is non-zero: an explicit "unrestricted"
// is a stated applicability, not an absent one.
func NewUnrestrictedContractApplicability() ContractApplicability {
	return ContractApplicability{kind: contractApplicabilityKindUnrestricted}
}

// NewScopedContractApplicability validates scope and returns a
// ContractApplicability bound to an explicit condition expression.
func NewScopedContractApplicability(scope core.Scope) (ContractApplicability, error) {
	if scope.IsZero() {
		return ContractApplicability{}, fmt.Errorf("runtime: NewScopedContractApplicability: %w: scope must not be zero", ErrInvalidContractApplicability)
	}
	return ContractApplicability{kind: contractApplicabilityKindScoped, scope: scope}, nil
}

// Kind returns a's discriminator, "unrestricted" or "scoped". The zero
// value returns the empty string.
func (a ContractApplicability) Kind() string { return string(a.kind) }

// IsUnrestricted reports whether a explicitly declares unrestricted
// applicability.
func (a ContractApplicability) IsUnrestricted() bool {
	return a.kind == contractApplicabilityKindUnrestricted
}

// IsScoped reports whether a declares a scoped applicability.
func (a ContractApplicability) IsScoped() bool { return a.kind == contractApplicabilityKindScoped }

// Scope returns a's condition expression, and whether one is set (that is,
// whether a is the scoped variant).
func (a ContractApplicability) Scope() (core.Scope, bool) {
	if a.kind != contractApplicabilityKindScoped {
		return core.Scope{}, false
	}
	return a.scope, true
}

// IsZero reports whether a is the zero value -- the unstated state
// PEOS-008 does not permit on a valid ContractContent.
func (a ContractApplicability) IsZero() bool { return a.kind == "" }

type contractApplicabilityJSON struct {
	Kind  string      `json:"kind"`
	Scope *core.Scope `json:"scope,omitempty"`
}

// MarshalJSON encodes a as {"kind":"unrestricted"} or {"kind":"scoped",
// "scope":{...}}. There is no top-level type discriminator beyond this
// union's own "kind".
func (a ContractApplicability) MarshalJSON() ([]byte, error) {
	switch a.kind {
	case contractApplicabilityKindUnrestricted:
		return json.Marshal(contractApplicabilityJSON{Kind: string(contractApplicabilityKindUnrestricted)})
	case contractApplicabilityKindScoped:
		return json.Marshal(contractApplicabilityJSON{Kind: string(contractApplicabilityKindScoped), Scope: &a.scope})
	default:
		return nil, fmt.Errorf("runtime: marshal ContractApplicability: %w", ErrInvalidContractApplicability)
	}
}

// UnmarshalJSON decodes a from its JSON form. An unrecognized or missing
// kind, an unrestricted value carrying a scope, and a scoped value missing
// a scope are all rejected. An explicit JSON null for the whole value
// decodes to an empty kind and is rejected the same way; a scoped value
// whose "scope" key is explicitly null is likewise rejected. The receiver
// is left untouched unless every check passes.
func (a *ContractApplicability) UnmarshalJSON(data []byte) error {
	var raw contractApplicabilityJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal ContractApplicability: %w: %w", ErrInvalidContractApplicability, err)
	}
	var result ContractApplicability
	switch raw.Kind {
	case string(contractApplicabilityKindUnrestricted):
		if raw.Scope != nil {
			return fmt.Errorf("runtime: unmarshal ContractApplicability: %w: unrestricted must not carry a scope", ErrInvalidContractApplicability)
		}
		result = NewUnrestrictedContractApplicability()
	case string(contractApplicabilityKindScoped):
		if raw.Scope == nil {
			return fmt.Errorf("runtime: unmarshal ContractApplicability: %w: scoped requires a scope", ErrInvalidContractApplicability)
		}
		var err error
		result, err = NewScopedContractApplicability(*raw.Scope)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("runtime: unmarshal ContractApplicability: unrecognized kind %q: %w", raw.Kind, ErrInvalidContractApplicability)
	}
	*a = result
	return nil
}

// --- RequirementReference ---------------------------------------------------

type requirementReferenceKind string

const (
	requirementReferenceKindIdentity requirementReferenceKind = "identity"
	requirementReferenceKindRevision requirementReferenceKind = "revision"
)

// RequirementReference is a PEOS-008 Runtime Requirement Reference: a
// closed two-arm union naming the exact Requirement identity or the exact
// Requirement Artifact Revision a Runtime Contract Revision governs.
//
// PEOS-008 requires that "every reference from a Runtime Contract to a
// Requirement or Requirement Artifact Revision SHALL preserve the exact
// participant level being referenced -- Requirement identity, or a
// specific Requirement Artifact Revision -- and SHALL NOT silently treat
// one as equivalent to the other." A RequirementReference can therefore
// never leave the participant level implicit: it is always exactly one of
// the two arms, never neither and never both, and there is no third
// "latest" or "current" arm -- an implicit-latest reference would silently
// treat identity and revision as equivalent, exactly what PEOS-008
// forbids.
//
// RequirementReference carries no Requirement Statement text. PEOS-008
// states "Requirement Statement content, as defined by PEOS-005, SHALL NOT
// be copied into mutable Runtime Contract fields" -- this type's two arms
// are core.RequirementRef and core.RequirementArtifactRevisionRef, neither
// of which carries Statement content, so the prohibition is structural
// rather than a convention this package could violate by accident.
type RequirementReference struct {
	kind     requirementReferenceKind
	identity core.RequirementRef
	revision core.RequirementArtifactRevisionRef
}

// NewRequirementIdentityReference validates ref and returns a
// RequirementReference naming the Requirement at identity level.
func NewRequirementIdentityReference(ref core.RequirementRef) (RequirementReference, error) {
	if ref.IsZero() {
		return RequirementReference{}, fmt.Errorf("runtime: NewRequirementIdentityReference: %w: reference must not be zero", ErrInvalidRequirementReference)
	}
	return RequirementReference{kind: requirementReferenceKindIdentity, identity: ref}, nil
}

// NewRequirementRevisionReference validates ref and returns a
// RequirementReference naming an exact Requirement Artifact Revision.
func NewRequirementRevisionReference(ref core.RequirementArtifactRevisionRef) (RequirementReference, error) {
	if ref.IsZero() {
		return RequirementReference{}, fmt.Errorf("runtime: NewRequirementRevisionReference: %w: reference must not be zero", ErrInvalidRequirementReference)
	}
	return RequirementReference{kind: requirementReferenceKindRevision, revision: ref}, nil
}

// Kind returns r's discriminator, "identity" or "revision". The zero value
// returns the empty string.
func (r RequirementReference) Kind() string { return string(r.kind) }

// Identity returns r's Requirement identity reference, and whether one is
// set (that is, whether r is the identity-level arm).
func (r RequirementReference) Identity() (core.RequirementRef, bool) {
	if r.kind != requirementReferenceKindIdentity {
		return core.RequirementRef{}, false
	}
	return r.identity, true
}

// Revision returns r's exact Requirement Artifact Revision reference, and
// whether one is set (that is, whether r is the revision-level arm).
func (r RequirementReference) Revision() (core.RequirementArtifactRevisionRef, bool) {
	if r.kind != requirementReferenceKindRevision {
		return core.RequirementArtifactRevisionRef{}, false
	}
	return r.revision, true
}

// RequirementArtifactID returns the ArtifactID of the Requirement r
// references, regardless of which arm r is. It returns the zero
// core.ArtifactID for a zero-value r.
func (r RequirementReference) RequirementArtifactID() core.ArtifactID {
	switch r.kind {
	case requirementReferenceKindIdentity:
		return r.identity.ArtifactID()
	case requirementReferenceKindRevision:
		return r.revision.ArtifactID()
	default:
		return core.ArtifactID{}
	}
}

// IsZero reports whether r is the zero value.
func (r RequirementReference) IsZero() bool { return r.kind == "" }

type requirementReferenceJSON struct {
	Kind     string                               `json:"kind"`
	Identity *core.RequirementRef                 `json:"identity,omitempty"`
	Revision *core.RequirementArtifactRevisionRef `json:"revision,omitempty"`
}

// requirementReferenceUnmarshalJSON mirrors requirementReferenceJSON for
// decoding, with both arms captured as raw bytes so an explicit null, or a
// value present for the wrong arm, can be distinguished and rejected --
// the json.RawMessage probe technique Packet D.1 established.
type requirementReferenceUnmarshalJSON struct {
	Kind     string          `json:"kind"`
	Identity json.RawMessage `json:"identity"`
	Revision json.RawMessage `json:"revision"`
}

// MarshalJSON encodes r as {"kind":"identity","identity":{...}} or
// {"kind":"revision","revision":{...}}. There is no third "latest" arm and
// no field carrying Requirement Statement text.
func (r RequirementReference) MarshalJSON() ([]byte, error) {
	switch r.kind {
	case requirementReferenceKindIdentity:
		return json.Marshal(requirementReferenceJSON{Kind: string(requirementReferenceKindIdentity), Identity: &r.identity})
	case requirementReferenceKindRevision:
		return json.Marshal(requirementReferenceJSON{Kind: string(requirementReferenceKindRevision), Revision: &r.revision})
	default:
		return nil, fmt.Errorf("runtime: marshal RequirementReference: %w", ErrInvalidRequirementReference)
	}
}

// UnmarshalJSON decodes r from its JSON form, applying the same validation
// as the two constructors. An unrecognized or missing kind, a value
// carrying the wrong arm's field, a value missing its own arm's field, and
// an explicit null for the present arm are all rejected. The receiver is
// left untouched unless every check passes.
func (r *RequirementReference) UnmarshalJSON(data []byte) error {
	var raw requirementReferenceUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal RequirementReference: %w: %w", ErrInvalidRequirementReference, err)
	}
	hasIdentity := len(raw.Identity) > 0 && string(raw.Identity) != "null"
	hasRevision := len(raw.Revision) > 0 && string(raw.Revision) != "null"

	var (
		result RequirementReference
		err    error
	)
	switch raw.Kind {
	case string(requirementReferenceKindIdentity):
		if hasRevision {
			return fmt.Errorf("runtime: unmarshal RequirementReference: %w: an identity-level reference must not carry a revision", ErrInvalidRequirementReference)
		}
		if !hasIdentity {
			return fmt.Errorf("runtime: unmarshal RequirementReference: %w: an identity-level reference requires an identity", ErrInvalidRequirementReference)
		}
		var ref core.RequirementRef
		if err = json.Unmarshal(raw.Identity, &ref); err != nil {
			return fmt.Errorf("runtime: unmarshal RequirementReference: %w: %w", ErrInvalidRequirementReference, err)
		}
		if result, err = NewRequirementIdentityReference(ref); err != nil {
			return err
		}
	case string(requirementReferenceKindRevision):
		if hasIdentity {
			return fmt.Errorf("runtime: unmarshal RequirementReference: %w: a revision-level reference must not carry an identity", ErrInvalidRequirementReference)
		}
		if !hasRevision {
			return fmt.Errorf("runtime: unmarshal RequirementReference: %w: a revision-level reference requires a revision", ErrInvalidRequirementReference)
		}
		var ref core.RequirementArtifactRevisionRef
		if err = json.Unmarshal(raw.Revision, &ref); err != nil {
			return fmt.Errorf("runtime: unmarshal RequirementReference: %w: %w", ErrInvalidRequirementReference, err)
		}
		if result, err = NewRequirementRevisionReference(ref); err != nil {
			return err
		}
	default:
		return fmt.Errorf("runtime: unmarshal RequirementReference: unrecognized kind %q: %w", raw.Kind, ErrInvalidRequirementReference)
	}
	*r = result
	return nil
}

// --- runtime-local key namespace ---------------------------------------------

// Runtime local-key namespaces are criterion-kind namespaces, not one
// namespace per Go collection. PEOS-008 gives a Runtime Contract Revision
// exactly two citable, locally-keyed collections of content --
// core.CriterionKindRuntimeAssertion and
// core.CriterionKindRuntimeContractRule -- and this package's key
// namespaces mirror that two-way split exactly, not the four Go
// collections the four Contract Rule categories happen to be stored in.
const (
	kindAssertion    = "assertion"
	kindContractRule = "runtime contract rule"
)

// addRuntimeLocalKey records key in set, rejecting a repeat within that one
// namespace. Uniqueness is per criterion-kind namespace, not per Go
// collection -- the same derivation quality.addProfileLocalKey documents:
// PEOS-008 states no key uniqueness rule at all, and the necessary derived
// rule is only that a criterion citing an owned value by key must resolve
// to exactly one such value within its own criterion kind.
func addRuntimeLocalKey(caller, kind string, set map[string]bool, key core.LocalKey) error {
	s := key.String()
	if set[s] {
		return fmt.Errorf("runtime: %s: %s key %q: %w", caller, kind, s, ErrDuplicateRuntimeLocalKey)
	}
	set[s] = true
	return nil
}

// --- ContractRule ----------------------------------------------------------

// ContractRule is a PEOS-008 Runtime Contract Revision-owned declarative
// rule value, used for observation requirements, violation classification
// rules, Waiver handling rules, and enforcement expectations alike. It
// carries a runtime-local key naming it within the
// core.CriterionKindRuntimeContractRule namespace (shared across all four
// categories, since core.RuntimeRuleCriterionRef's payload carries no
// rule-category discriminator of its own -- see "runtime-local key
// namespace" above) and an opaque, trimmed rule text.
//
// ContractRule carries no independent identity, no Revision, and no
// lifecycle of its own -- it is Revision-owned content, exactly like
// Assertion. It carries no rule-category field: the collection it is
// stored in (observation requirements, violation classification rules,
// Waiver handling rules, or enforcement expectations) already supplies
// that, and duplicating it inside the value would be redundant, not
// additional structure. It carries no expression language, no executable
// behaviour, and no evaluation result -- PEOS-008 defines none of these
// for any Contract Rule category, and this package does not resolve a
// ContractRule's criterion citation against any repository.
type ContractRule struct {
	key  core.LocalKey
	text string
}

// NewContractRule validates key and text and returns a ContractRule.
//
// Both are mandatory. key must be non-zero, and text must be non-empty
// after trimming; the trimmed value is stored. text is the declarative
// rule itself, an opaque string: PEOS-008 defines no rule grammar or
// expression language for any of the four categories it names.
func NewContractRule(key core.LocalKey, text string) (ContractRule, error) {
	if key.IsZero() {
		return ContractRule{}, fmt.Errorf("runtime: NewContractRule: %w: key must not be zero", ErrInvalidRuntimeContractRule)
	}
	trimmed, err := trimmedRequired("NewContractRule", "text", text, ErrInvalidRuntimeContractRule)
	if err != nil {
		return ContractRule{}, err
	}
	return ContractRule{key: key, text: trimmed}, nil
}

// Key returns r's runtime-local key. It is meaningful only within the
// combined Runtime Contract Rule namespace of its owning ContractContent.
func (r ContractRule) Key() core.LocalKey { return r.key }

// Text returns r's declarative rule text, uninterpreted.
func (r ContractRule) Text() string { return r.text }

// IsZero reports whether r is the zero value.
func (r ContractRule) IsZero() bool { return r.key.IsZero() && r.text == "" }

type contractRuleJSON struct {
	Key  core.LocalKey `json:"key"`
	Text string        `json:"text"`
}

// MarshalJSON encodes r as {"key":...,"text":...}. There is no "category"
// key -- the collection r is stored in supplies that -- and no "result",
// "outcome", "satisfied", or "evaluation" key: a ContractRule declares a
// rule and never records whether it held.
func (r ContractRule) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("runtime: marshal ContractRule: %w", ErrInvalidRuntimeContractRule)
	}
	return json.Marshal(contractRuleJSON{Key: r.key, Text: r.text})
}

// UnmarshalJSON decodes r from its JSON form, applying the same validation
// as NewContractRule. The receiver is left untouched unless every check
// passes.
func (r *ContractRule) UnmarshalJSON(data []byte) error {
	var raw contractRuleJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal ContractRule: %w: %w", ErrInvalidRuntimeContractRule, err)
	}
	result, err := NewContractRule(raw.Key, raw.Text)
	if err != nil {
		return err
	}
	*r = result
	return nil
}

// --- ContractContent ----------------------------------------------------------

// ContractContent is the typed normative content PEOS-008 assigns to every
// Artifact Revision whose Artifact is a Runtime Contract: the Requirements
// it governs, the runtime subject and scope it applies to, its
// applicability, its provenance and authority, its Assertions, and its
// optional observation, classification, Waiver-handling, enforcement, and
// Quality Profile references.
//
// # Mandatory versus optional
//
// requirements, subjectTarget, environment, deploymentScope, applicability,
// provenance, and authority are mandatory constructor arguments and are
// unreachable through any later With* call: PEOS-008 states each of these
// as a Contract Revision SHALL-identify item, without the "where required"
// or "where applicable" qualifier PEOS-008 uses elsewhere (compare
// authority here, unqualified, against Runtime Binding Record authority,
// ":278", explicitly "where required"). This is why ContractContent
// exposes no WithRequirements, WithSubjectTarget, WithEnvironment,
// WithDeploymentScope, WithApplicability, WithProvenance, or WithAuthority.
//
// requirements additionally carries the one explicit minimum cardinality
// PEOS-008 states for Contract content: "A Runtime Contract Revision SHALL
// reference one or more Requirements or Requirement Artifact Revisions
// that it governs." No other collection here has any minimum: assertions,
// observationRequirements, violationClassificationRules,
// waiverHandlingRules, enforcementExpectations, and
// qualityProfileRevisions may all be empty. PEOS-008 lists Assertions,
// observation requirements, violation classification rules, and Waiver
// handling rules without a cardinality qualifier, exactly the same
// unqualified form this repository already reads as permitting emptiness
// for validation.PlannedActivity's own "criteria", "Evidence expected", and
// "execution prerequisites"; Quality Profile Revisions are explicitly
// "where required"; and enforcement expectations are likewise unqualified.
//
// assertions is a constructor argument, unlike the four Contract Rule
// collections, only because it is validated together with them by the same
// shared path every other field uses -- not modified afterward through a
// separate, unvalidated path. All five owned-value collections (assertions
// plus the four ContractRule categories) are keyed and participate in the
// runtime-local key namespace rules documented above: assertions form the
// core.CriterionKindRuntimeAssertion namespace, and the four Contract Rule
// categories together form one combined core.CriterionKindRuntimeContractRule
// namespace, because PEOS-008 ":423"/":475" require a "Runtime Contract
// rule" to be citable as a criterion, and core.RuntimeRuleCriterionRef's
// (Revision, LocalKey) payload carries no rule-category discriminator to
// separate the four collections at the criterion level.
//
// # No derived state
//
// ContractContent stores no current binding, no deployment status, and no
// compliance outcome. Current Runtime Binding and Runtime Compliance are
// PEOS-008 derived views, computed on demand by a repository from Runtime
// Binding, Unbinding, Observation, and Violation Records and Compliance
// Claims -- never stored as a mutable field on a Runtime Contract Revision.
type ContractContent struct {
	requirements    []RequirementReference
	subjectTarget   core.RuntimeSubjectRef
	environment     Environment
	deploymentScope core.Scope
	applicability   ContractApplicability
	provenance      core.Provenance
	authority       core.AuthorityRef

	assertions []Assertion

	observationRequirements      []ContractRule
	violationClassificationRules []ContractRule
	waiverHandlingRules          []ContractRule
	enforcementExpectations      []ContractRule
	qualityProfileRevisions      []core.ArtifactRevisionRef

	extension core.Extension
}

// validateContractContent is the single shared validation path for
// ContractContent. NewContractContent, every collection With* method, and
// UnmarshalJSON all route through it, so the mandatory-field rule, the
// requirements minimum, and the Assertion key-uniqueness rule cannot drift
// between construction, modification, and decoding.
func validateContractContent(caller string, c ContractContent) error {
	if len(c.requirements) == 0 {
		return fmt.Errorf("runtime: %s: %w: at least one requirement reference is required", caller, ErrInvalidRuntimeContract)
	}
	for _, r := range c.requirements {
		if r.IsZero() {
			return fmt.Errorf("runtime: %s: %w: requirement reference must not be zero", caller, ErrInvalidRequirementReference)
		}
	}
	if c.subjectTarget.IsZero() {
		return fmt.Errorf("runtime: %s: %w: subject target must not be zero", caller, ErrInvalidRuntimeContract)
	}
	if c.environment.IsZero() {
		return fmt.Errorf("runtime: %s: %w: environment must not be zero", caller, ErrInvalidRuntimeContract)
	}
	if c.deploymentScope.IsZero() {
		return fmt.Errorf("runtime: %s: %w: deployment scope must not be zero", caller, core.ErrInvalidScope)
	}
	if c.applicability.IsZero() {
		return fmt.Errorf("runtime: %s: %w: applicability must be explicitly stated", caller, ErrInvalidContractApplicability)
	}
	if c.provenance.IsZero() {
		return fmt.Errorf("runtime: %s: %w: provenance must not be zero", caller, ErrInvalidRuntimeContract)
	}
	if c.authority.IsZero() {
		return fmt.Errorf("runtime: %s: %w: authority must not be zero", caller, ErrInvalidRuntimeContract)
	}

	assertionKeys := make(map[string]bool, len(c.assertions))
	for _, v := range c.assertions {
		if v.IsZero() {
			return fmt.Errorf("runtime: %s: %w: assertion must not be zero", caller, ErrInvalidRuntimeAssertion)
		}
		if err := addRuntimeLocalKey(caller, kindAssertion, assertionKeys, v.Key()); err != nil {
			return err
		}
	}

	// The four Contract Rule categories share one combined
	// core.CriterionKindRuntimeContractRule namespace -- see "runtime-local
	// key namespace" above -- so their keys are validated together against
	// a single set, not one set per Go collection.
	contractRuleKeys := make(map[string]bool, len(c.observationRequirements)+len(c.violationClassificationRules)+len(c.waiverHandlingRules)+len(c.enforcementExpectations))
	for _, v := range c.observationRequirements {
		if v.IsZero() {
			return fmt.Errorf("runtime: %s: %w: observation requirement must not be zero", caller, ErrInvalidRuntimeContractRule)
		}
		if err := addRuntimeLocalKey(caller, kindContractRule, contractRuleKeys, v.Key()); err != nil {
			return err
		}
	}
	for _, v := range c.violationClassificationRules {
		if v.IsZero() {
			return fmt.Errorf("runtime: %s: %w: violation classification rule must not be zero", caller, ErrInvalidRuntimeContractRule)
		}
		if err := addRuntimeLocalKey(caller, kindContractRule, contractRuleKeys, v.Key()); err != nil {
			return err
		}
	}
	for _, v := range c.waiverHandlingRules {
		if v.IsZero() {
			return fmt.Errorf("runtime: %s: %w: waiver handling rule must not be zero", caller, ErrInvalidRuntimeContractRule)
		}
		if err := addRuntimeLocalKey(caller, kindContractRule, contractRuleKeys, v.Key()); err != nil {
			return err
		}
	}
	for _, v := range c.enforcementExpectations {
		if v.IsZero() {
			return fmt.Errorf("runtime: %s: %w: enforcement expectation must not be zero", caller, ErrInvalidRuntimeContractRule)
		}
		if err := addRuntimeLocalKey(caller, kindContractRule, contractRuleKeys, v.Key()); err != nil {
			return err
		}
	}

	for _, ref := range c.qualityProfileRevisions {
		if ref.IsZero() {
			return fmt.Errorf("runtime: %s: %w: quality profile revision reference must not be zero", caller, ErrInvalidRuntimeContract)
		}
	}

	return nil
}

// NewContractContent validates its arguments and returns a ContractContent
// with no observation requirements, violation classification rules, Waiver
// handling rules, enforcement expectations, Quality Profile Revisions, or
// extension data. Use the With* methods to add those.
//
// requirements must contain at least one non-zero RequirementReference --
// PEOS-008's one explicit minimum cardinality. subjectTarget, environment,
// deploymentScope, applicability, provenance, and authority must all be
// non-zero; applicability must be explicitly stated (use
// NewUnrestrictedContractApplicability to declare an explicit absence of
// restriction). assertions may be empty or nil -- PEOS-008 states no
// minimum cardinality for it -- but no element within a non-empty
// collection may be zero-valued, and its keys must be unique within the
// collection.
//
// Every slice argument is defensively copied; the caller may reuse or
// mutate its own slices afterward without affecting the returned value.
func NewContractContent(
	requirements []RequirementReference,
	subjectTarget core.RuntimeSubjectRef,
	environment Environment,
	deploymentScope core.Scope,
	applicability ContractApplicability,
	provenance core.Provenance,
	authority core.AuthorityRef,
	assertions []Assertion,
) (ContractContent, error) {
	c := ContractContent{
		requirements:    copySlice(requirements),
		subjectTarget:   subjectTarget,
		environment:     environment,
		deploymentScope: deploymentScope,
		applicability:   applicability,
		provenance:      provenance,
		authority:       authority,
		assertions:      copySlice(assertions),
	}
	if err := validateContractContent("NewContractContent", c); err != nil {
		return ContractContent{}, err
	}
	return c, nil
}

// WithObservationRequirements returns a copy of c with its observation
// requirements set to exactly the ContractRule values given, in the order
// given. A zero-value element is rejected, and every key must be unique
// within the combined Runtime Contract Rule namespace shared with the
// other three categories (see "runtime-local key namespace" above).
// Passing an empty or nil slice declares none, which is why there is no
// WithoutObservationRequirements: WithObservationRequirements(nil) already
// expresses removal.
func (c ContractContent) WithObservationRequirements(rules []ContractRule) (ContractContent, error) {
	c.observationRequirements = copySlice(rules)
	if err := validateContractContent("ContractContent.WithObservationRequirements", c); err != nil {
		return ContractContent{}, err
	}
	return c, nil
}

// WithViolationClassificationRules returns a copy of c with its violation
// classification rules set to exactly the ContractRule values given, in
// the order given. A zero-value element is rejected, and every key must be
// unique within the combined Runtime Contract Rule namespace. Passing an
// empty or nil slice declares none.
func (c ContractContent) WithViolationClassificationRules(rules []ContractRule) (ContractContent, error) {
	c.violationClassificationRules = copySlice(rules)
	if err := validateContractContent("ContractContent.WithViolationClassificationRules", c); err != nil {
		return ContractContent{}, err
	}
	return c, nil
}

// WithWaiverHandlingRules returns a copy of c with its Waiver handling
// rules set to exactly the ContractRule values given, in the order given.
// A zero-value element is rejected, and every key must be unique within
// the combined Runtime Contract Rule namespace. Passing an empty or nil
// slice declares none.
//
// These describe how an applicable PEOS-005 Waiver is to be handled;
// PEOS-008 consumes PEOS-005 Waiver semantics without redefining them, and
// this package does not import peos/requirement (see doc.go).
func (c ContractContent) WithWaiverHandlingRules(rules []ContractRule) (ContractContent, error) {
	c.waiverHandlingRules = copySlice(rules)
	if err := validateContractContent("ContractContent.WithWaiverHandlingRules", c); err != nil {
		return ContractContent{}, err
	}
	return c, nil
}

// WithEnforcementExpectations returns a copy of c with its enforcement
// expectations set to exactly the ContractRule values given, in the order
// given. A zero-value element is rejected, and every key must be unique
// within the combined Runtime Contract Rule namespace. Passing an empty or
// nil slice declares none.
func (c ContractContent) WithEnforcementExpectations(rules []ContractRule) (ContractContent, error) {
	c.enforcementExpectations = copySlice(rules)
	if err := validateContractContent("ContractContent.WithEnforcementExpectations", c); err != nil {
		return ContractContent{}, err
	}
	return c, nil
}

// WithQualityProfileRevisions returns a copy of c with its applicable
// Quality Profile Revision references set to exactly the values given, in
// the order given. A zero-value element is rejected. Passing an empty or
// nil slice declares none -- PEOS-008 requires these only "where
// required".
//
// Applicable Quality Profile Revisions are referenced with the
// general-purpose core.ArtifactRevisionRef, not a dedicated
// QualityProfileRevisionRef type: peos/core defines none, and this package
// does not import peos/quality (see doc.go).
func (c ContractContent) WithQualityProfileRevisions(revisions []core.ArtifactRevisionRef) (ContractContent, error) {
	c.qualityProfileRevisions = copySlice(revisions)
	if err := validateContractContent("ContractContent.WithQualityProfileRevisions", c); err != nil {
		return ContractContent{}, err
	}
	return c, nil
}

// WithExtension returns a copy of c with its extension data set. Passing
// the zero core.Extension is equivalent to declaring none.
func (c ContractContent) WithExtension(extension core.Extension) ContractContent {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c ContractContent) WithoutExtension() ContractContent {
	c.extension = core.Extension{}
	return c
}

// trimmedStringSlice trims each entry of values and rejects any that is
// empty after trimming, attributing the failure to sentinel. Returns nil
// for an empty input, so that passing nil declares "none" without an
// allocation.
func trimmedStringSlice(caller, label string, values []string, sentinel error) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	cp := make([]string, len(values))
	for idx, v := range values {
		trimmed, err := trimmedRequired(caller, label+" entry", v, sentinel)
		if err != nil {
			return nil, err
		}
		cp[idx] = trimmed
	}
	return cp, nil
}

// Requirements returns a defensive copy of c's Requirement references, in
// declaration order. Never empty on a valid ContractContent.
func (c ContractContent) Requirements() []RequirementReference { return copySlice(c.requirements) }

// SubjectTarget returns c's declared runtime subject type or target. It is
// mandatory and therefore never absent on a valid ContractContent.
func (c ContractContent) SubjectTarget() core.RuntimeSubjectRef { return c.subjectTarget }

// Environment returns c's declared environment.
func (c ContractContent) Environment() Environment { return c.environment }

// DeploymentScope returns c's declared deployment scope.
func (c ContractContent) DeploymentScope() core.Scope { return c.deploymentScope }

// Applicability returns c's declared applicability conditions. It is
// mandatory and therefore never unstated on a valid ContractContent.
func (c ContractContent) Applicability() ContractApplicability { return c.applicability }

// Provenance returns c's declared provenance.
func (c ContractContent) Provenance() core.Provenance { return c.provenance }

// Authority returns c's declared authority. It is mandatory for this
// packet's reading of PEOS-008 (unqualified at ":196"; contrast Runtime
// Binding Record authority, ":278", explicitly "where required").
func (c ContractContent) Authority() core.AuthorityRef { return c.authority }

// Assertions returns a defensive copy of c's Runtime Assertions, in
// declaration order. May be empty: PEOS-008 states no minimum cardinality.
func (c ContractContent) Assertions() []Assertion { return copySlice(c.assertions) }

// Assertion returns the Assertion in c whose runtime-local key equals key,
// and whether one was found.
func (c ContractContent) Assertion(key core.LocalKey) (Assertion, bool) {
	if key.IsZero() {
		return Assertion{}, false
	}
	for _, v := range c.assertions {
		if v.Key() == key {
			return v, true
		}
	}
	return Assertion{}, false
}

// ObservationRequirements returns a defensive copy of c's observation
// requirements, in declaration order.
func (c ContractContent) ObservationRequirements() []ContractRule {
	return copySlice(c.observationRequirements)
}

// ViolationClassificationRules returns a defensive copy of c's violation
// classification rules, in declaration order.
func (c ContractContent) ViolationClassificationRules() []ContractRule {
	return copySlice(c.violationClassificationRules)
}

// WaiverHandlingRules returns a defensive copy of c's Waiver handling
// rules, in declaration order.
func (c ContractContent) WaiverHandlingRules() []ContractRule {
	return copySlice(c.waiverHandlingRules)
}

// EnforcementExpectations returns a defensive copy of c's enforcement
// expectations, in declaration order.
func (c ContractContent) EnforcementExpectations() []ContractRule {
	return copySlice(c.enforcementExpectations)
}

// ContractRule returns the ContractRule in c whose runtime-local key
// equals key, and whether one was found. The lookup spans all four
// Contract Rule categories -- observation requirements, violation
// classification rules, Waiver handling rules, and enforcement
// expectations -- because they share one combined
// core.CriterionKindRuntimeContractRule namespace (see "runtime-local key
// namespace" above), so a key resolves to at most one ContractRule across
// all four collections together. This performs local resolution only
// within an already-loaded ContractContent; it does not load another
// Revision, verify the referenced Contract Revision exists, or evaluate
// the rule.
func (c ContractContent) ContractRule(key core.LocalKey) (ContractRule, bool) {
	if key.IsZero() {
		return ContractRule{}, false
	}
	for _, categories := range [][]ContractRule{
		c.observationRequirements,
		c.violationClassificationRules,
		c.waiverHandlingRules,
		c.enforcementExpectations,
	} {
		for _, v := range categories {
			if v.Key() == key {
				return v, true
			}
		}
	}
	return ContractRule{}, false
}

// QualityProfileRevisions returns a defensive copy of c's applicable
// Quality Profile Revision references, in declaration order.
func (c ContractContent) QualityProfileRevisions() []core.ArtifactRevisionRef {
	return copySlice(c.qualityProfileRevisions)
}

// Extension returns c's extension data.
func (c ContractContent) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c ContractContent) IsZero() bool {
	return len(c.requirements) == 0 && c.subjectTarget.IsZero() && c.environment.IsZero() &&
		c.deploymentScope.IsZero() && c.applicability.IsZero() && c.provenance.IsZero() && c.authority.IsZero()
}

type contractContentJSON struct {
	Requirements                 []RequirementReference     `json:"requirements"`
	SubjectTarget                core.RuntimeSubjectRef     `json:"subject_target"`
	Environment                  Environment                `json:"environment"`
	DeploymentScope              core.Scope                 `json:"deployment_scope"`
	Applicability                ContractApplicability      `json:"applicability"`
	Provenance                   core.Provenance            `json:"provenance"`
	Authority                    core.AuthorityRef          `json:"authority"`
	Assertions                   []Assertion                `json:"assertions,omitempty"`
	ObservationRequirements      []ContractRule             `json:"observation_requirements,omitempty"`
	ViolationClassificationRules []ContractRule             `json:"violation_classification_rules,omitempty"`
	WaiverHandlingRules          []ContractRule             `json:"waiver_handling_rules,omitempty"`
	EnforcementExpectations      []ContractRule             `json:"enforcement_expectations,omitempty"`
	QualityProfileRevisions      []core.ArtifactRevisionRef `json:"quality_profile_revisions,omitempty"`
	Extension                    *core.Extension            `json:"extension,omitempty"`
}

// MarshalJSON encodes c with its seven mandatory keys always present, plus
// whichever optional keys are set.
//
// There is no "bound", "active_deployment", "deployed", "compliant",
// "compliance", "state", "status", "lifecycle", "state_assignment",
// "relation", "source", "target", "version", "current", "latest",
// "effective", "incident", "verdict", or "outcome_authority" key, and no
// top-level PEOS type discriminator. Their absence is the structural proof
// that a Runtime Contract Revision carries declared enforcement content
// only -- never a relationship, never a lifecycle, and never stored
// runtime binding or compliance state.
func (c ContractContent) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("runtime: marshal ContractContent: %w", ErrInvalidRuntimeContract)
	}
	raw := contractContentJSON{
		Requirements:                 c.requirements,
		SubjectTarget:                c.subjectTarget,
		Environment:                  c.environment,
		DeploymentScope:              c.deploymentScope,
		Applicability:                c.applicability,
		Provenance:                   c.provenance,
		Authority:                    c.authority,
		Assertions:                   c.assertions,
		ObservationRequirements:      c.observationRequirements,
		ViolationClassificationRules: c.violationClassificationRules,
		WaiverHandlingRules:          c.waiverHandlingRules,
		EnforcementExpectations:      c.enforcementExpectations,
		QualityProfileRevisions:      c.qualityProfileRevisions,
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// contractContentUnmarshalJSON mirrors contractContentJSON for decoding
// only, with Authority captured as raw bytes so an explicit JSON null can
// be distinguished from an absent key and rejected -- the json.RawMessage
// probe technique Packet D.1 established. Requirements, SubjectTarget,
// Environment, DeploymentScope, Applicability, and Provenance need no such
// treatment: an absent key and an explicit null both yield a zero value
// that validateContractContent rejects, so the two cases converge on the
// same error and need not be told apart. Every optional collection needs
// no such treatment either, but for the opposite reason: absent, null, and
// [] all denote the same valid state, "declares none of this kind".
type contractContentUnmarshalJSON struct {
	Requirements                 []RequirementReference     `json:"requirements"`
	SubjectTarget                core.RuntimeSubjectRef     `json:"subject_target"`
	Environment                  Environment                `json:"environment"`
	DeploymentScope              core.Scope                 `json:"deployment_scope"`
	Applicability                ContractApplicability      `json:"applicability"`
	Provenance                   core.Provenance            `json:"provenance"`
	Authority                    json.RawMessage            `json:"authority"`
	Assertions                   []Assertion                `json:"assertions"`
	ObservationRequirements      []ContractRule             `json:"observation_requirements"`
	ViolationClassificationRules []ContractRule             `json:"violation_classification_rules"`
	WaiverHandlingRules          []ContractRule             `json:"waiver_handling_rules"`
	EnforcementExpectations      []ContractRule             `json:"enforcement_expectations"`
	QualityProfileRevisions      []core.ArtifactRevisionRef `json:"quality_profile_revisions"`
	Extension                    *core.Extension            `json:"extension,omitempty"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as NewContractContent and each With* method, so a decoded ContractContent
// can never be constructor-impossible. The receiver is left untouched
// unless every check passes.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - requirements: absent, null, and [] are all rejected by the ≥1
//     minimum -- the same error, so the three need not be distinguished.
//   - subject_target, environment, deployment_scope, applicability,
//     provenance: a missing key leaves the field zero and is rejected
//     through its owning sentinel. An explicit null invokes that nested
//     type's own UnmarshalJSON, which fails there for subject_target,
//     environment, and deployment_scope, or yields an empty kind for
//     applicability. Both are rejected; the sentinel sets differ.
//   - authority: a missing key and an explicit null are both rejected --
//     authority is mandatory in this packet's reading, so there is no
//     "absent means unset" case to preserve.
//   - assertions, observation_requirements, violation_classification_rules,
//     waiver_handling_rules, enforcement_expectations,
//     quality_profile_revisions: absent, explicit null, and empty array are
//     all equivalent and all mean "declares none of this kind".
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract.
func (c *ContractContent) UnmarshalJSON(data []byte) error {
	var raw contractContentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal ContractContent: %w: %w", ErrInvalidRuntimeContract, err)
	}

	if len(raw.Authority) == 0 || string(raw.Authority) == "null" {
		return fmt.Errorf("runtime: unmarshal ContractContent: %w: authority must not be absent or null", ErrInvalidRuntimeContract)
	}
	var authority core.AuthorityRef
	if err := json.Unmarshal(raw.Authority, &authority); err != nil {
		return fmt.Errorf("runtime: unmarshal ContractContent: %w: %w", ErrInvalidRuntimeContract, err)
	}

	result := ContractContent{
		requirements:                 copySlice(raw.Requirements),
		subjectTarget:                raw.SubjectTarget,
		environment:                  raw.Environment,
		deploymentScope:              raw.DeploymentScope,
		applicability:                raw.Applicability,
		provenance:                   raw.Provenance,
		authority:                    authority,
		assertions:                   copySlice(raw.Assertions),
		observationRequirements:      copySlice(raw.ObservationRequirements),
		violationClassificationRules: copySlice(raw.ViolationClassificationRules),
		waiverHandlingRules:          copySlice(raw.WaiverHandlingRules),
		enforcementExpectations:      copySlice(raw.EnforcementExpectations),
		qualityProfileRevisions:      copySlice(raw.QualityProfileRevisions),
	}
	if err := validateContractContent("unmarshal ContractContent", result); err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*c = result
	return nil
}

// --- ContractRevision ---------------------------------------------------------

// ContractRevision is shorthand for "an Artifact Revision whose Artifact is
// a Runtime Contract" -- not a separate PEOS entity, and not a Runtime
// Contract Version. It composes core.ArtifactRevision by named field, per
// the specialized-Revision strategy already followed by
// requirement.Revision, validation.PlanRevision, and quality.ProfileRevision,
// and pairs it with typed ContractContent.
//
// ContractRevision is immutable and exposes no WithContent: "Modification
// of the declared content of a Runtime Contract... constitutes a content
// change and SHALL create a new Artifact Revision", so a new
// ContractRevision is constructed rather than an existing one edited.
//
// ContractRevision carries no runtime binding, deployment, or compliance
// state: "Creation of a new Runtime Contract Revision does not, by itself,
// change which Revision is currently bound to a runtime subject."
// Establishing or changing a binding is exclusively the province of a
// Runtime Binding Record (Packet J.2), never a field on this type.
type ContractRevision struct {
	core    core.ArtifactRevision
	content ContractContent
}

// newContractRevisionFromParts validates revision and content without
// reference to any Contract, and is the path both NewContractRevision and
// UnmarshalJSON share. It cannot, and does not attempt to, check that
// revision belongs to any particular Contract -- see NewContractRevision
// and UnmarshalJSON for why that check needs a Contract value a
// ContractRevision's own JSON does not carry.
func newContractRevisionFromParts(revision core.ArtifactRevision, content ContractContent) (ContractRevision, error) {
	if revision.IsZero() {
		return ContractRevision{}, fmt.Errorf("%w: core revision must not be zero", ErrInvalidRuntimeContract)
	}
	if content.IsZero() {
		return ContractRevision{}, fmt.Errorf("%w: contract content must not be zero", ErrInvalidRuntimeContract)
	}
	return ContractRevision{core: revision, content: content}, nil
}

// NewContractRevision validates contract, revision, and content and
// returns a ContractRevision. contract and revision must both be non-zero,
// content must be non-zero, and revision.ArtifactID() must equal
// contract.ID().
func NewContractRevision(contract Contract, revision core.ArtifactRevision, content ContractContent) (ContractRevision, error) {
	if contract.IsZero() {
		return ContractRevision{}, fmt.Errorf("runtime: NewContractRevision: %w: contract must not be zero", ErrInvalidRuntimeContract)
	}
	result, err := newContractRevisionFromParts(revision, content)
	if err != nil {
		return ContractRevision{}, fmt.Errorf("runtime: NewContractRevision: %w", err)
	}
	if revision.ArtifactID() != contract.ID() {
		return ContractRevision{}, fmt.Errorf("runtime: NewContractRevision: %w", ErrRuntimeContractArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the ContractRevision's underlying core.ArtifactRevision.
func (r ContractRevision) Core() core.ArtifactRevision { return r.core }

// Content returns the ContractRevision's typed Runtime Contract content.
func (r ContractRevision) Content() ContractContent { return r.content }

// Ref returns a core.RuntimeContractRevisionRef identifying r. A Runtime
// Binding Record is required to reference a Runtime Contract Revision
// using exactly this type, never the bare core.RuntimeContractRef.
func (r ContractRevision) Ref() (core.RuntimeContractRevisionRef, error) {
	return core.NewRuntimeContractRevisionRef(r.core.ArtifactID(), r.core.RevisionID())
}

// IsZero reports whether r is the zero value.
func (r ContractRevision) IsZero() bool { return r.core.IsZero() && r.content.IsZero() }

type contractRevisionJSON struct {
	Core    core.ArtifactRevision `json:"core"`
	Content ContractContent       `json:"content"`
}

// MarshalJSON encodes r as {"core":{...},"content":{...}}, per the
// nested-composition strategy core.ArtifactRevision documents.
func (r ContractRevision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("runtime: marshal ContractRevision: %w", ErrInvalidRuntimeContract)
	}
	return json.Marshal(contractRevisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core":{...},"content":{...}}
// JSON form.
//
// This reconstructs r.core and r.content via the same checks
// newContractRevisionFromParts (and therefore NewContractRevision) applies,
// but cannot repeat NewContractRevision's ArtifactID-to-Contract
// cross-check: a ContractRevision's own JSON carries only its
// core.ArtifactRevision (with a bare ArtifactID) and its ContractContent,
// never a core.Artifact with an ArtifactType to check that ArtifactID
// against. This is the same limitation core.ArtifactRevision,
// requirement.Revision, validation.PlanRevision, and
// quality.ProfileRevision already document.
func (r *ContractRevision) UnmarshalJSON(data []byte) error {
	var raw contractRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal ContractRevision: %w: %w", ErrInvalidRuntimeContract, err)
	}
	result, err := newContractRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return fmt.Errorf("runtime: unmarshal ContractRevision: %w", err)
	}
	*r = result
	return nil
}
