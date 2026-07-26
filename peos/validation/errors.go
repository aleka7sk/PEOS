package validation

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels
// rather than comparing error values directly.
//
// The complete PEOS-006 sentinel set is declared here up front, including
// the four sentinels reserved for Packet H.2 (Validation Execution Record
// and Validation Claim). Declaring them together keeps the package's error
// taxonomy inspectable as a whole and means H.2 does not have to reopen
// this file. Each reserved sentinel documents the exact concept it belongs
// to and the accepted architecture it will enforce; none is referenced by
// Packet H.1 code, because H.1 deliberately implements only the Validation
// Plan side.
//
// Component-owned failures are never re-attributed to this package: a
// zero or malformed core.Scope surfaces core.ErrInvalidScope, a malformed
// nested core reference surfaces core.ErrInvalidPayload or
// core.ErrInvalidReferenceDiscriminator, an empty identity value surfaces
// core.ErrEmptyIdentity, and a malformed vocabulary value surfaces
// core.ErrInvalidVocabularyValue. This package wraps such errors, adding
// its own context, without replacing the owning sentinel.
var (
	// ErrInvalidValidationPlan is the aggregate sentinel for the Validation
	// Plan side of PEOS-006: a zero-value core.Artifact supplied to NewPlan,
	// a zero-value core.ArtifactRevision or PlanContent supplied to
	// NewPlanRevision, a zero-value core.Provenance or an empty Planned
	// Activity list supplied to NewPlanContent, an invalid Plan-level
	// acceptance-rules value, and a zero-value marshal or a failed
	// top-level decode of any of Plan, PlanRevision, or PlanContent.
	// Component-specific failures use their own sentinels instead: see
	// ErrInvalidPlanApplicability, ErrInvalidPlannedActivity,
	// ErrDuplicatePlanLocalKey, ErrUnknownPlanLocalKey, and
	// core.ErrInvalidScope.
	ErrInvalidValidationPlan = errors.New("validation: validation plan is invalid")

	// ErrValidationPlanArtifactTypeMismatch is returned when NewPlan
	// receives a non-zero core.Artifact whose declared Artifact Type is not
	// ArtifactTypeValidationPlan (PEOS-006: "A Validation Plan is an
	// Artifact as defined by PEOS-002"). It mirrors
	// requirement.ErrRequirementArtifactTypeMismatch.
	ErrValidationPlanArtifactTypeMismatch = errors.New("validation: artifact type is not validation plan")

	// ErrValidationPlanArtifactIDMismatch is returned when a PlanRevision's
	// core Artifact Revision refers to a different Artifact than the Plan it
	// is being paired with. It mirrors
	// requirement.ErrRequirementArtifactIDMismatch.
	ErrValidationPlanArtifactIDMismatch = errors.New("validation: artifact id mismatch between validation plan and revision")

	// ErrInvalidPlannedActivity is returned when a PlannedActivity is
	// constructed or decoded with a zero plan-local key, a zero Subject, a
	// zero Validation Method, or an empty/whitespace-only outcome
	// interpretation; when one of its collections contains a zero-value or
	// empty element; when one of its optional single-value fields is
	// supplied zero; or when a zero-value PlannedActivity is marshaled. It
	// is also returned by NewPlanContent when the supplied Planned Activity
	// list contains a zero-value element (PEOS-006 "Planned Validation
	// Activity").
	ErrInvalidPlannedActivity = errors.New("validation: planned validation activity is invalid")

	// ErrDuplicatePlanLocalKey is returned when the Planned Validation
	// Activities of one PlanContent contain the same core.LocalKey more than
	// once (PEOS-006 Plan-Local Key Invariant: "Every Planned Validation
	// Activity has a stable key unique within its owning Validation Plan
	// Revision"; non-conforming pattern "Missing Plan-Local Key").
	ErrDuplicatePlanLocalKey = errors.New("validation: duplicate plan-local key")

	// ErrUnknownPlanLocalKey is returned when a Planned Validation
	// Activity declares a dependency on a core.LocalKey that no Activity in
	// the same PlanContent defines. A plan-local key "does not survive as an
	// independent identity outside that exact Plan Revision" (PEOS-006
	// Planned Validation Activity), so a dependency naming a key absent from
	// the same Revision cannot denote anything.
	ErrUnknownPlanLocalKey = errors.New("validation: unknown plan-local key")

	// ErrInvalidPlanApplicability is returned when a PlanApplicability is
	// left in its zero (unstated) state, is decoded with an unrecognized or
	// missing kind, is unrestricted yet carries a scope, or is scoped yet
	// carries no scope. PEOS-006 lists "applicability" among a Validation
	// Plan Revision's unqualified SHALL-identify items, so a PlanContent may
	// never leave it unstated; NewUnrestrictedPlanApplicability exists so
	// that "no restriction" is an explicit, non-zero value.
	ErrInvalidPlanApplicability = errors.New("validation: plan applicability is invalid")

	// ErrInvalidActivityReference is reserved for Packet H.2. It will be
	// returned when an ActivityReference -- the closed two-arm union naming
	// either an exact Planned Validation Activity (Plan Revision reference
	// plus plan-local key) or an explicitly identified ad hoc execution --
	// is left in its zero state, decoded with an unrecognized kind, missing
	// its Plan Revision reference or key, or given an empty ad hoc
	// designation (PEOS-006 Validation Execution Record). It is not used by
	// Packet H.1.
	ErrInvalidActivityReference = errors.New("validation: activity reference is invalid")

	// ErrInvalidExecutionRecord is reserved for Packet H.2. It will be the
	// aggregate sentinel for Validation Execution Record and its
	// ExecutionEvent content: a zero mandatory field, a started timestamp
	// after the completed timestamp, a zero-value marshal, or a failed
	// decode (PEOS-006 Validation Execution Record). It is not used by
	// Packet H.1.
	ErrInvalidExecutionRecord = errors.New("validation: validation execution record is invalid")

	// ErrInvalidValidationClaim is reserved for Packet H.2. It will be the
	// aggregate sentinel for Validation Claim: a zero mandatory field, an
	// empty Evidence list (PEOS-006 requires a Claim to cite "one or more
	// Evidence Artifact Revisions"), a zero-value core.CriterionRef element,
	// a zero-value marshal, or a failed decode. It is not used by Packet
	// H.1.
	ErrInvalidValidationClaim = errors.New("validation: validation claim is invalid")

	// ErrInvalidSatisfactionClaim is reserved for Packet H.2. It will be
	// returned for the two Claim-Type-conditional invariants a Satisfaction
	// Claim adds: no criterion identifying a Requirement or Requirement
	// Artifact Revision, and a subject whose Requirement identity equals
	// that of any Requirement criterion (PEOS-006 Satisfaction Claim: "A
	// Requirement SHALL NOT become both the Claim subject and the same
	// Claim's criterion"). It is not used by Packet H.1.
	ErrInvalidSatisfactionClaim = errors.New("validation: satisfaction claim is invalid")

	// ErrInvalidConformanceClaim is reserved for Packet H.2. It will be
	// returned when a Conformance Claim declares zero criteria: PEOS-006
	// defines a Conformance Claim as "a Validation Claim whose criteria
	// identify one or more" conformance rules, and evaluates its Subject
	// "against those criteria". No kind restriction is enforceable, because
	// that clause's permitted set ends in "or other explicitly identified
	// conformance rules" -- an open set with no closed mapping onto
	// core.CriterionRef's kinds. It is not used by Packet H.1.
	ErrInvalidConformanceClaim = errors.New("validation: conformance claim is invalid")
)
