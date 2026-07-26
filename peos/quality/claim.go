package quality

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// Claim is a PEOS-007 Quality Claim: "a specialization of Validation Claim, as
// defined by PEOS-006", which "exists exclusively as an instance of the
// PEOS-006 Validation Claim mechanism".
//
// The type name is Claim, not QualityClaim. Within package quality the
// qualifier is already carried by the package name, and a QualityClaim type
// would read as a second Claim base model -- exactly the non-conforming
// pattern "Parallel Quality Claim Base" names.
//
// # Why a wrapper rather than a bare constructor
//
// PEOS-007 imposes one rule on a quality Claim that PEOS-006 does not enforce:
// where the subject is a Requirement, "the Requirement SHALL NOT be its own
// criterion". PEOS-006 is correct not to enforce it -- it deliberately permits
// that shape for every non-Satisfaction Claim Type, and
// validation.checkClaimTypeCriteria therefore falls through to its default case
// for core.ClaimTypeQuality.
//
// A constructor plus an optional validator function would leave a reachable
// hole: validation.Claim.WithCriteria re-runs only PEOS-006's rules, so
// construct-then-WithCriteria could produce a represented, PEOS-007-invalid
// quality Claim with no error anywhere. A wrapper closes it structurally.
// Every path that can change a field this rule depends on -- construction,
// adoption of a raw Claim, every modifier, and JSON decoding -- routes through
// checkQualityClaim before any value is assigned. That is the same
// constructor-completeness discipline as "no modifier establishes mandatory
// state", applied to a cross-field invariant.
//
// # Not a parallel mechanism
//
// The wrapper adds no stored field beyond the composed validation.Claim, and
// there is exactly one of each of the following, every one inherited rather
// than restated (PEOS-007: a Quality Claim inherits "identity; immutability;
// the absence of revisions; the absence of lifecycle; the requirement of
// exactly one Subject; the separation of criteria from Subject; Evidence
// citation rules; the replacement, correction, and invalidation rules defined
// by PEOS-006 ...; historical preservation; derived-current-effect
// semantics"):
//
//   - identity -- core.ValidationClaimID; this type declares no ID field;
//   - Claim Type -- core.ClaimTypeQuality, an existing core vocabulary value,
//     fixed at construction and unchangeable thereafter;
//   - subject, scope, outcome, method, criteria, timestamp, provenance;
//   - Evidence model -- core.EvidenceArtifactRevisionRef, at least one;
//   - correction model -- core.RecordCorrectionRef[core.ValidationClaimRef],
//     "never Artifact Supersession";
//   - wire form -- byte-for-byte validation.Claim's, with no envelope and no
//     added key.
//
// PEOS-007's entire contribution is stricter validation. It is not an
// Artifact, not an Artifact Revision, not revisioned, and not
// lifecycle-bearing: "A Quality Claim does not itself assign a Lifecycle State
// or State Assignment."
//
// # Extract, modify, re-wrap
//
// ValidationClaim returns a value copy. Modifying that copy cannot mutate the
// Claim it came from, but the modified raw value is no longer represented as a
// validated quality Claim -- nothing about a bare validation.Claim asserts
// PEOS-007's rule. Re-entry requires NewClaimFromValidationClaim, which
// re-applies every check.
type Claim struct {
	claim validation.Claim
}

// checkQualityClaim is the single shared PEOS-007 validation path for Claim.
// NewClaim, NewClaimFromValidationClaim, every modifier (through
// wrapQualityClaim), MarshalJSON, and UnmarshalJSON all route through it, so
// the rules cannot drift between construction, modification, encoding, and
// decoding.
//
// It enforces exactly two things:
//
//  1. The Claim Type is core.ClaimTypeQuality. This is what makes adopting a
//     raw validation.Claim safe, and what guarantees ClaimType() never returns
//     anything else.
//  2. PEOS-007's Requirement-criterion rule: "A Requirement MAY be the Subject
//     of a Quality Claim only where the Requirement itself is being evaluated
//     as an engineering artifact ... In that case the criterion SHALL be a
//     Quality Characteristic, Measure, Threshold, Profile rule, or other rule
//     distinct from the Requirement subject itself; the Requirement SHALL NOT
//     be its own criterion."
//
// Everything else about a quality Claim is PEOS-006's and is already enforced
// by validation.NewClaim: exactly one subject, at least one Evidence citation,
// non-zero mandatory fields, no zero-value criterion element.
//
// PEOS-007 adds no minimum criteria count. "criteria may include" is
// permissive, and the specification nowhere requires a quality Claim to cite
// any criterion, so zero criteria are accepted -- consistent with PEOS-006's
// "Where zero criteria are identified, the Claim's outcome SHALL be
// interpreted strictly according to its stated Validation Method and basis."
func checkQualityClaim(caller string, claim validation.Claim) error {
	if claim.IsZero() {
		return fmt.Errorf("quality: %s: %w: validation claim must not be zero", caller, ErrInvalidQualityClaim)
	}
	if claim.ClaimType() != core.ClaimTypeQuality {
		return fmt.Errorf("quality: %s: %w: claim type is %q, want %q", caller, ErrInvalidQualityClaim, claim.ClaimType().String(), core.ClaimTypeQuality.String())
	}

	subjectID, ok := subjectRequirementArtifactID(claim.Subject())
	if !ok {
		// A subject of any other kind -- an Artifact, an Artifact Revision, a
		// Decision Outcome, a Runtime Subject -- can never conflict with a
		// Requirement criterion, so any Requirement criterion is permitted.
		return nil
	}
	if slices.Contains(requirementCriterionArtifactIDs(claim.Criteria()), subjectID) {
		return fmt.Errorf("quality: %s: %w: Requirement %q must not be both the quality claim subject and one of its own criteria", caller, ErrInvalidQualityClaim, subjectID.String())
	}
	return nil
}

// requirementCriterionArtifactIDs returns the owning Requirement
// core.ArtifactID of every criterion identifying a Requirement or a Requirement
// Artifact Revision, collapsing both levels to identity so the comparison in
// checkQualityClaim is cross-level.
//
// This mirrors, rather than reuses, the equivalent helper inside
// peos/validation: that one is unexported, and PEOS-007 must not require a
// change to peos/validation to widen its visibility. The behaviour is
// deliberately identical, because "the same Requirement" is identity-level
// language in both specifications.
func requirementCriterionArtifactIDs(criteria []core.CriterionRef) []core.ArtifactID {
	var ids []core.ArtifactID
	for _, c := range criteria {
		if ref, ok := c.AsRequirement(); ok {
			ids = append(ids, ref.ArtifactID())
			continue
		}
		if ref, ok := c.AsRequirementRevision(); ok {
			ids = append(ids, ref.ArtifactID())
		}
	}
	return ids
}

// subjectRequirementArtifactID returns the owning Requirement core.ArtifactID
// of a subject identifying a Requirement or a Requirement Artifact Revision,
// and whether the subject is of either kind.
func subjectRequirementArtifactID(subject core.EngineeringSubjectRef) (core.ArtifactID, bool) {
	if ref, ok := subject.AsRequirement(); ok {
		return ref.ArtifactID(), true
	}
	if ref, ok := subject.AsRequirementRevision(); ok {
		return ref.ArtifactID(), true
	}
	return core.ArtifactID{}, false
}

// wrapQualityClaim validates claim as a PEOS-007 Quality Claim and returns the
// wrapper. It is the only place a Claim value is ever constructed, so no path
// can assign an unvalidated one.
func wrapQualityClaim(caller string, claim validation.Claim) (Claim, error) {
	if err := checkQualityClaim(caller, claim); err != nil {
		return Claim{}, err
	}
	return Claim{claim: claim}, nil
}

// NewClaim validates its nine arguments and returns a Quality Claim with no
// optional field set. Use the With* methods to add those.
//
// There is deliberately no claimType parameter: a Quality Claim's Claim Type is
// core.ClaimTypeQuality by definition, this constructor supplies it, and no
// public path can change it afterwards. Every other argument is PEOS-006's and
// is validated by validation.NewClaim under its own sentinels -- in particular
// evidence must cite at least one core.EvidenceArtifactRevisionRef, and scope
// must be non-zero (rejected with core.ErrInvalidScope).
//
// criteria is a constructor argument for the same reason it is one on
// validation.NewClaim, reinforced here: PEOS-007's Requirement-criterion rule
// couples the subject and the criteria, and the subject is immutable, so a
// Claim built without criteria and completed afterwards could pass through a
// normatively invalid state. Zero criteria are accepted -- PEOS-007 states no
// minimum.
//
// Both slices are defensively copied by validation.NewClaim.
func NewClaim(
	id core.ValidationClaimID,
	subject core.EngineeringSubjectRef,
	scope core.Scope,
	outcome core.ClaimOutcome,
	method core.ValidationMethod,
	criteria []core.CriterionRef,
	evidence []core.EvidenceArtifactRevisionRef,
	timestamp core.Timestamp,
	provenance core.Provenance,
) (Claim, error) {
	claim, err := validation.NewClaim(
		id,
		core.ClaimTypeQuality,
		subject,
		scope,
		outcome,
		method,
		criteria,
		evidence,
		timestamp,
		provenance,
	)
	if err != nil {
		return Claim{}, err
	}
	return wrapQualityClaim("NewClaim", claim)
}

// NewClaimFromValidationClaim validates an existing validation.Claim as a
// PEOS-007 Quality Claim and returns the wrapper.
//
// This is the re-entry point for a value that already exists as a
// validation.Claim: one read back from a repository, one decoded from PEOS-006
// wire form, one arriving across an integration boundary, or one deliberately
// modified as a raw Claim after extraction through ValidationClaim.
//
// It rejects a zero Claim, and any Claim whose Claim Type is not
// core.ClaimTypeQuality -- Satisfaction, Conformance, Compliance, Template
// Conformance, and any Product-defined Claim Type are all refused, because a
// quality Claim is not merely a Claim that happens to be about quality. It
// also rejects a quality Claim that violates PEOS-007's Requirement-criterion
// rule, which is precisely the shape peos/validation correctly accepts.
func NewClaimFromValidationClaim(claim validation.Claim) (Claim, error) {
	return wrapQualityClaim("NewClaimFromValidationClaim", claim)
}

// --- modifiers ---------------------------------------------------------------
//
// Every modifier follows the same three steps: delegate to the corresponding
// validation.Claim modifier, re-validate the result through checkQualityClaim,
// and only then return the new wrapper. A failure at either step returns the
// zero Claim and an error, and because the receiver is a value, it is never
// mutated.
//
// All ten return (Claim, error), including the five whose validation.Claim
// counterparts cannot fail. The uniformity is deliberate: PEOS-007 is a Draft
// at v0.1, and an added invariant touching authority, reasoning, or extension
// would otherwise force a breaking signature change. The cost, stated plainly,
// is that only WithCriteria can currently return a PEOS-007 error -- the rule
// depends on subject and criteria alone, and no modifier can reach the subject.
// Those unreachable branches are justified in coverage rather than chased.
//
// There is no WithClaimType, no WithoutCriteria, and no modifier for any
// mandatory field. The Claim Type is set once by NewClaim and re-verified on
// every wrap, so no sequence of public calls can change it.

// WithCriteria returns a copy of c with its criteria set to exactly the values
// given, in the order given, replacing any previous criteria. Passing an empty
// or nil slice declares zero criteria, which PEOS-007 permits -- which is why
// there is no WithoutCriteria: WithCriteria(nil) already expresses removal, and
// a second method would create a second validation path for one field.
//
// This is the modifier PEOS-007's Requirement-criterion rule actually
// constrains, and the reason this wrapper exists. Supplying a criteria set that
// names c's own Requirement subject is rejected with ErrInvalidQualityClaim,
// where the same call on a bare validation.Claim would succeed.
func (c Claim) WithCriteria(criteria []core.CriterionRef) (Claim, error) {
	updated, err := c.claim.WithCriteria(criteria)
	if err != nil {
		return Claim{}, err
	}
	return wrapQualityClaim("Claim.WithCriteria", updated)
}

// WithExecutionRecords returns a copy of c with the relevant Validation
// Execution Records set to exactly the values given, in the order given. These
// are the references a Measurement Record is cited by. An empty or nil slice
// declares none.
func (c Claim) WithExecutionRecords(records []core.ValidationExecutionRecordRef) (Claim, error) {
	updated, err := c.claim.WithExecutionRecords(records)
	if err != nil {
		return Claim{}, err
	}
	return wrapQualityClaim("Claim.WithExecutionRecords", updated)
}

// WithReasoning returns a copy of c with its reasoning or interpretation set.
// The value is trimmed and must be non-empty after trimming.
func (c Claim) WithReasoning(reasoning string) (Claim, error) {
	updated, err := c.claim.WithReasoning(reasoning)
	if err != nil {
		return Claim{}, err
	}
	return wrapQualityClaim("Claim.WithReasoning", updated)
}

// WithoutReasoning returns a copy of c with its reasoning cleared.
func (c Claim) WithoutReasoning() (Claim, error) {
	return wrapQualityClaim("Claim.WithoutReasoning", c.claim.WithoutReasoning())
}

// WithAuthority returns a copy of c with its authority set. authority must be
// non-zero; use WithoutAuthority to clear it.
//
// An authority records who had the right to establish the Claim. It never makes
// the Claim a governance act: treating a quality outcome as sufficient
// authority to establish or change an Engineering Commitment is the
// non-conforming pattern "Quality Outcome Used as Governance Authority", and
// PEOS-004 remains the sole owner of that decision.
func (c Claim) WithAuthority(authority core.AuthorityRef) (Claim, error) {
	updated, err := c.claim.WithAuthority(authority)
	if err != nil {
		return Claim{}, err
	}
	return wrapQualityClaim("Claim.WithAuthority", updated)
}

// WithoutAuthority returns a copy of c with its authority cleared.
func (c Claim) WithoutAuthority() (Claim, error) {
	return wrapQualityClaim("Claim.WithoutAuthority", c.claim.WithoutAuthority())
}

// WithCorrection returns a copy of c declaring that c corrects, replaces, or
// invalidates an earlier Validation Claim, identified exactly.
//
// This is PEOS-006's correction model, inherited unchanged: a minimal
// record-to-record reference, never an Artifact Relation and never Artifact
// Supersession. PEOS-007 introduces no second replacement mechanism. Which
// Claim is currently applicable remains a derived view, never a stored field.
func (c Claim) WithCorrection(correction core.RecordCorrectionRef[core.ValidationClaimRef]) (Claim, error) {
	updated, err := c.claim.WithCorrection(correction)
	if err != nil {
		return Claim{}, err
	}
	return wrapQualityClaim("Claim.WithCorrection", updated)
}

// WithoutCorrection returns a copy of c with its correction reference cleared.
func (c Claim) WithoutCorrection() (Claim, error) {
	return wrapQualityClaim("Claim.WithoutCorrection", c.claim.WithoutCorrection())
}

// WithExtension returns a copy of c with its extension data set. Passing the
// zero core.Extension is equivalent to declaring none.
func (c Claim) WithExtension(extension core.Extension) (Claim, error) {
	return wrapQualityClaim("Claim.WithExtension", c.claim.WithExtension(extension))
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c Claim) WithoutExtension() (Claim, error) {
	return wrapQualityClaim("Claim.WithoutExtension", c.claim.WithoutExtension())
}

// --- accessors ---------------------------------------------------------------

// ValidationClaim returns a value copy of c's underlying PEOS-006 Validation
// Claim.
//
// The accessor is named after the wrapped concept, following the convention
// this repository already uses twice: Core() where a core.Artifact or
// core.ArtifactRevision is wrapped, Relation() where a relation.Relation is
// wrapped. Core() would be wrong here -- validation.Claim is not a core type.
//
// The returned value is a copy: modifying it cannot mutate c. The result of
// such a modification is a bare validation.Claim, no longer represented as a
// validated Quality Claim; pass it back through NewClaimFromValidationClaim to
// re-establish the PEOS-007 guarantees.
func (c Claim) ValidationClaim() validation.Claim { return c.claim }

// ID returns c's identity, which is the composed Validation Claim's own
// core.ValidationClaimID. A Quality Claim has no separate identity space.
func (c Claim) ID() core.ValidationClaimID { return c.claim.ID() }

// Ref returns a core.ValidationClaimRef citing c. This is the reference a later
// correcting Claim targets, and the reference a PEOS-004 Decision may cite as
// part of its Decision Basis.
func (c Claim) Ref() (core.ValidationClaimRef, error) { return c.claim.Ref() }

// ClaimType returns c's Claim Type. For a valid Claim this is always
// core.ClaimTypeQuality: NewClaim supplies it, checkQualityClaim re-verifies it
// on every construction, adoption, modification, and decode, and no modifier
// can change it.
func (c Claim) ClaimType() core.ClaimType { return c.claim.ClaimType() }

// Subject returns c's single engineering Subject -- "the engineering subject
// whose quality is evaluated". PEOS-007 permits exactly one, and forbids a
// composite subject such as a Characteristic-and-Revision pair; the field is a
// single core.EngineeringSubjectRef, so a composite is unrepresentable.
func (c Claim) Subject() core.EngineeringSubjectRef { return c.claim.Subject() }

// Scope returns c's explicit scope.
func (c Claim) Scope() core.Scope { return c.claim.Scope() }

// Outcome returns c's recorded determination about its Subject. A quality score
// may appear here as an outcome attribute; it is never stored as current
// mutable state on the Subject.
func (c Claim) Outcome() core.ClaimOutcome { return c.claim.Outcome() }

// Method returns the Validation Method c applied.
func (c Claim) Method() core.ValidationMethod { return c.claim.Method() }

// Criteria returns a defensive copy of the criteria c evaluated, in
// declaration order. A Quality Characteristic, Measure, Threshold, Target, or
// Quality Constraint is always a criterion, never a second Subject -- enforced
// by the type system, since core.CriterionRef and core.EngineeringSubjectRef
// have no conversion path in either direction.
func (c Claim) Criteria() []core.CriterionRef { return c.claim.Criteria() }

// Evidence returns a defensive copy of the exact Evidence Artifact Revisions c
// relied upon. A valid Claim always cites at least one -- PEOS-006's rule,
// inherited without a parallel quality Evidence mechanism.
func (c Claim) Evidence() []core.EvidenceArtifactRevisionRef { return c.claim.Evidence() }

// Timestamp returns the time c's assertion was made.
func (c Claim) Timestamp() core.Timestamp { return c.claim.Timestamp() }

// Provenance returns c's provenance.
func (c Claim) Provenance() core.Provenance { return c.claim.Provenance() }

// ExecutionRecords returns a defensive copy of the Validation Execution Records
// relevant to c. For a Quality Claim these are typically the Measurement
// Records it interprets, cited through their inherited
// core.ValidationExecutionRecordRef.
func (c Claim) ExecutionRecords() []core.ValidationExecutionRecordRef {
	return c.claim.ExecutionRecords()
}

// Reasoning returns c's recorded reasoning or interpretation, and whether any
// is set.
func (c Claim) Reasoning() (string, bool) { return c.claim.Reasoning() }

// Authority returns c's declared authority, and whether one is set.
func (c Claim) Authority() (core.AuthorityRef, bool) { return c.claim.Authority() }

// Correction returns c's declared correction reference to an earlier Validation
// Claim, and whether one is set.
func (c Claim) Correction() (core.RecordCorrectionRef[core.ValidationClaimRef], bool) {
	return c.claim.Correction()
}

// Extension returns c's extension data.
func (c Claim) Extension() core.Extension { return c.claim.Extension() }

// IsZero reports whether c is the zero value.
func (c Claim) IsZero() bool { return c.claim.IsZero() }

// MarshalJSON encodes c as exactly the wire form of its underlying
// validation.Claim -- the same keys, the same nesting, the same bytes.
//
// There is deliberately no {"claim":{...}} envelope and no PEOS-007
// discriminator. A Quality Claim is a Validation Claim, and claim_type already
// carries "peos:quality" as ordinary content, so a wrapper key would create a
// second wire representation of one record -- which is what the non-conforming
// pattern "Parallel Quality Claim Base" forbids. One consequence is worth
// stating: a quality Claim's document is readable by any PEOS-006 consumer with
// no knowledge of PEOS-007.
//
// A zero or non-quality wrapper fails with ErrInvalidQualityClaim rather than
// producing bytes.
func (c Claim) MarshalJSON() ([]byte, error) {
	if err := checkQualityClaim("marshal Claim", c.claim); err != nil {
		return nil, err
	}
	return json.Marshal(c.claim)
}

// UnmarshalJSON decodes c from validation.Claim's wire form, then applies
// PEOS-007's own checks.
//
// The nested validation.Claim decodes first, which re-applies every PEOS-006
// rule -- including the json.RawMessage probe distinguishing an absent
// "criteria" key from an explicit null, and the at-least-one Evidence
// invariant. Then the Claim Type is required to be core.ClaimTypeQuality, and
// the Requirement-criterion rule is applied. The receiver is assigned only
// after all three stages pass, so a failed decode leaves a previously valid
// value untouched, and a decoded Claim can never be constructor-impossible.
//
// Sentinel reachability, all via errors.Is: ErrInvalidQualityClaim for a wrong
// Claim Type or a Requirement-criterion violation;
// validation.ErrInvalidValidationClaim for a PEOS-006 aggregate failure;
// core.ErrInvalidScope, core.ErrEmptyIdentity, core.ErrMissingRevisionID,
// core.ErrInvalidVocabularyValue, core.ErrInvalidPayload,
// core.ErrInvalidReferenceDiscriminator, and
// core.ErrInvalidCorrectionReference for malformed nested core values. No
// owning sentinel is ever replaced -- only wrapped.
func (c *Claim) UnmarshalJSON(data []byte) error {
	var claim validation.Claim
	if err := json.Unmarshal(data, &claim); err != nil {
		return fmt.Errorf("quality: unmarshal Claim: %w: %w", ErrInvalidQualityClaim, err)
	}
	result, err := wrapQualityClaim("unmarshal Claim", claim)
	if err != nil {
		return err
	}
	*c = result
	return nil
}
