package runtime

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels rather
// than comparing error values directly.
//
// The complete PEOS-008 sentinel set is declared here up front, per the
// convention Packet H.1/I.1 established: declaring a later packet's
// sentinels ahead of time means that packet does not have to reopen this
// file. Four sentinels below --ErrInvalidRuntimeBindingRecord,
// ErrInvalidRuntimeUnbindingRecord, ErrInvalidRuntimeObservation, and
// ErrInvalidRuntimeViolation-- are reserved for Packet J.2 (Runtime Binding,
// Unbinding, Observation, and Violation records) and are not referenced by
// any Packet J.1 code, because J.1 implements only the Runtime Contract
// declaration side.
//
// There is deliberately no per-field sentinel. Each field belongs to
// exactly one owning aggregate, and a caller that receives
// ErrInvalidRuntimeAssertion or ErrInvalidContractApplicability already
// knows which aggregate rejected the input; the wrapped message names the
// field. There is also no operational-failure sentinel of any kind
// (binding-readiness, deployment-state, compliance-status): PEOS-008
// defines no runtime failure taxonomy, and none is invented here -- see
// doc.go.
//
// Component-owned failures are never re-attributed to this package: a zero
// or malformed core.Scope surfaces core.ErrInvalidScope, an empty identity
// or local key surfaces core.ErrEmptyIdentity, a malformed vocabulary value
// surfaces core.ErrInvalidVocabularyValue, and a malformed nested core
// reference surfaces core.ErrInvalidPayload or
// core.ErrInvalidReferenceDiscriminator. This package wraps such errors,
// adding its own context, without replacing the owning sentinel.
var (
	// ErrInvalidRuntimeContract is the aggregate sentinel for the Runtime
	// Contract Artifact and its Revision content: a zero-value core.Artifact
	// supplied to NewContract, a zero-value core.ArtifactRevision or
	// ContractContent supplied to NewContractRevision, an empty Requirement
	// reference collection, a zero mandatory scalar (subject target,
	// environment, deployment scope, provenance, authority), or an invalid
	// optional value, plus a zero-value marshal or a failed top-level decode
	// of Contract, ContractRevision, or ContractContent. Component-specific
	// failures use their own sentinels instead.
	ErrInvalidRuntimeContract = errors.New("runtime: runtime contract is invalid")

	// ErrRuntimeContractArtifactTypeMismatch is returned when NewContract
	// receives a non-zero core.Artifact whose declared Artifact Type is not
	// ArtifactTypeRuntimeContract (PEOS-008: "A Runtime Contract is an
	// Artifact as defined by PEOS-002"). It mirrors
	// quality.ErrQualityProfileArtifactTypeMismatch and
	// validation.ErrValidationPlanArtifactTypeMismatch.
	ErrRuntimeContractArtifactTypeMismatch = errors.New("runtime: artifact type is not runtime contract")

	// ErrRuntimeContractArtifactIDMismatch is returned when a
	// ContractRevision's core Artifact Revision refers to a different
	// Artifact than the Contract it is being paired with. It mirrors
	// quality.ErrQualityProfileArtifactIDMismatch.
	ErrRuntimeContractArtifactIDMismatch = errors.New("runtime: artifact id mismatch between runtime contract and revision")

	// ErrInvalidContractApplicability is returned when a
	// ContractApplicability is left in its zero (unstated) state, is
	// decoded with an unrecognized or missing kind, is unrestricted yet
	// carries a scope, or is scoped yet carries no scope.
	ErrInvalidContractApplicability = errors.New("runtime: contract applicability is invalid")

	// ErrInvalidRequirementReference is returned when a RequirementReference
	// is constructed or decoded with a zero payload, an unrecognized or
	// missing kind, or both/neither arm present. PEOS-008 requires that
	// "every reference... SHALL preserve the exact participant level" (:221),
	// so a RequirementReference can never leave that level implicit.
	ErrInvalidRequirementReference = errors.New("runtime: requirement reference is invalid")

	// ErrInvalidRuntimeAssertion is returned when an Assertion is
	// constructed or decoded with a zero mandatory field (key, subject,
	// criterion, evaluation rule, expected result, scope), an empty
	// observation-input entry after trimming, or a zero-value marshal.
	ErrInvalidRuntimeAssertion = errors.New("runtime: runtime assertion is invalid")

	// ErrInvalidRuntimeContractRule is the aggregate sentinel for
	// ContractRule -- the Revision-owned, locally-keyed declarative value
	// PEOS-008 requires for observation requirements, violation
	// classification rules, Waiver handling rules, and enforcement
	// expectations (Packet J.2.A / audit finding J3-03). It is returned
	// when a ContractRule is constructed or decoded with a zero key or an
	// empty rule text after trimming, or a zero-value marshal.
	//
	// This sentinel was not part of Packet J.1's originally declared set
	// because J.1 modelled all four Contract Rule categories as opaque,
	// unkeyed []string -- a reading later proven incorrect: PEOS-008
	// ":423" and ":475" require a "Runtime Contract rule" to be citable as
	// a criterion via core.CriterionKindRuntimeContractRule, whose payload
	// (core.RuntimeRuleCriterionRef) is a (Revision, LocalKey) pair with no
	// resolvable target unless at least one Contract-owned collection is
	// keyed. A genuinely new aggregate owned-value type is therefore
	// required, and it needs its own sentinel for the same reason
	// ErrInvalidRuntimeAssertion is distinct from ErrInvalidRuntimeContract.
	ErrInvalidRuntimeContractRule = errors.New("runtime: runtime contract rule is invalid")

	// ErrInvalidRuntimeBindingRecord is reserved for Packet J.2. It will be
	// the aggregate sentinel for BindingRecord -- the PEOS-008 Runtime
	// Binding Record. It is not used by Packet J.1.
	ErrInvalidRuntimeBindingRecord = errors.New("runtime: runtime binding record is invalid")

	// ErrInvalidRuntimeUnbindingRecord is reserved for Packet J.2. It will
	// be the aggregate sentinel for UnbindingRecord -- the PEOS-008 Runtime
	// Unbinding Record. It is not used by Packet J.1.
	ErrInvalidRuntimeUnbindingRecord = errors.New("runtime: runtime unbinding record is invalid")

	// ErrInvalidRuntimeObservation is reserved for Packet J.2. It will be
	// the aggregate sentinel for Observation -- the PEOS-008 Runtime
	// Observation. It is not used by Packet J.1.
	ErrInvalidRuntimeObservation = errors.New("runtime: runtime observation is invalid")

	// ErrInvalidRuntimeViolation is reserved for Packet J.2. It will be the
	// aggregate sentinel for Violation -- the PEOS-008 Runtime Violation. It
	// is not used by Packet J.1.
	ErrInvalidRuntimeViolation = errors.New("runtime: runtime violation is invalid")

	// ErrDuplicateRuntimeLocalKey is returned when a core.LocalKey repeats
	// within a single criterion-kind namespace of one ContractContent.
	// Uniqueness is enforced per criterion-kind namespace, not per Go
	// collection: the Assertion namespace (core.CriterionKindRuntimeAssertion)
	// is exactly ContractContent.assertions, but the Runtime Contract Rule
	// namespace (core.CriterionKindRuntimeContractRule) spans all four
	// Contract Rule categories combined -- observation requirements,
	// violation classification rules, Waiver handling rules, and
	// enforcement expectations -- because core.RuntimeRuleCriterionRef
	// carries no rule-category discriminator, so a key must resolve to
	// exactly one ContractRule across all four collections together. A key
	// may therefore be reused once by an Assertion and once by a
	// ContractRule (the two namespaces are independent), but not twice
	// within either namespace, and not twice across the four Contract Rule
	// categories.
	//
	// The wrapped message always names the namespace and the offending
	// key, so a caller need not branch on a per-namespace sentinel to
	// report which one failed.
	ErrDuplicateRuntimeLocalKey = errors.New("runtime: duplicate runtime-local key")

	// ErrUnknownRuntimeLocalKey is returned when an internal reference
	// inside one ContractContent names a core.LocalKey that its expected
	// target collection does not define. Packet J.1 declares this sentinel
	// but has no internal reference that uses it yet -- an Assertion's
	// criterion is not repository-resolved by this package (see doc.go) --
	// so it becomes reachable only if a later packet introduces such a
	// reference.
	ErrUnknownRuntimeLocalKey = errors.New("runtime: unknown runtime-local key")
)
