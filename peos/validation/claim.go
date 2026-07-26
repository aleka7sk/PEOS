package validation

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Claim is a PEOS-006 Validation Claim: "an immutable record",
// "independently identifiable", which "is not an Artifact. It is not
// revisioned. It is not lifecycle-bearing." It "has exactly one engineering
// subject and zero or more criteria", and "preserves its historical
// assertion permanently."
//
// One Go type covers every Claim specialization. Satisfaction Claim,
// Conformance Claim, Quality Claim (PEOS-007), Compliance Claim (PEOS-008),
// and Template Conformance Claim (PEOS-009) are core.ClaimType *values*, not
// distinct types: PEOS-006 defines a Satisfaction Claim as "a Validation
// Claim whose criteria identify one or more Requirements" and a Conformance
// Claim as "a Validation Claim whose criteria identify one or more"
// conformance rules -- both are the same record under a constraint. core also
// declares a single identity space for all of them
// (core.ValidationClaimID's own documentation says so).
//
// Structurally: Claim carries core.ValidationClaimID as its own identity,
// composes no core.Artifact and no core.ArtifactRevision, has no Core
// accessor, is not a Requirement, composes no relation.Relation, and carries
// no Lifecycle State, State Assignment, or status field (non-conforming
// patterns "Claim as Artifact", "Claim Revision", "Claim Lifecycle").
//
// There is no basis field. PEOS-006 is explicit: "Claim Basis is not an
// independent opaque field distinct from the fields it groups. This
// specification does not require an additional required field named 'basis'
// beyond the individually identified method, criteria, Evidence, Execution
// Records, and reasoning." Those five fields are all present individually,
// and "Claim Basis" remains only a collective name for them.
//
// There is no verdict field: "There is no separate Verdict entity. The
// outcome recorded on a Validation Claim is the complete and only
// representation of what was determined."
//
// A Claim's outcome (core.ClaimOutcome: satisfied / not satisfied /
// inconclusive) describes what was determined about the Subject. It is
// distinct from an ExecutionRecord's core.ExecutionOutcome (completed /
// failed / interrupted / indeterminate), which describes whether the activity
// ran. A completed execution may carry a not-satisfied Claim.
//
// A Claim records an assertion; it does not authorize anything.
// "Certification, acceptance, approval, rejection, and authorization are
// governance outcomes governed by PEOS-004", and "A Validation Claim ... does
// not itself authorize, approve, or accept anything on behalf of an
// organization." Claim carries no Decision Outcome reference and no
// governance action, only an optional core.AuthorityRef.
//
// Immutability is structural: every field is unexported, every modifier
// returns a copy, and no modifier touches a mandatory field. "Correction,
// replacement, and invalidation of a Validation Claim are each represented by
// recording a new Validation Claim." The correction reference on the *new*
// Claim points backward, so no already-written Claim is ever rewritten.
type Claim struct {
	id         core.ValidationClaimID
	claimType  core.ClaimType
	subject    core.EngineeringSubjectRef
	scope      core.Scope
	outcome    core.ClaimOutcome
	method     core.ValidationMethod
	criteria   []core.CriterionRef
	evidence   []core.EvidenceArtifactRevisionRef
	timestamp  core.Timestamp
	provenance core.Provenance

	executionRecords []core.ValidationExecutionRecordRef
	reasoning        string
	authority        core.AuthorityRef
	correction       core.RecordCorrectionRef[core.ValidationClaimRef]
	extension        core.Extension
}

// NewClaim validates all ten mandatory arguments, in the order listed, and
// returns a Claim with no optional field set. Use the With* methods to add
// those.
//
// criteria is a constructor argument, not a later With* call, because its
// required cardinality and validity depend on claimType: a Satisfaction Claim
// needs at least one Requirement-kind criterion and a Conformance Claim needs
// at least one criterion, so a Claim built without criteria and completed
// afterward would necessarily pass through a normatively invalid state. No
// With* sequence can make a Claim-Type/criteria pair valid atomically, which
// is why the pair is established together.
//
// evidence must contain at least one citation for every Claim Type: PEOS-006
// states that a Claim "cites one or more Evidence Artifact Revisions",
// deliberately contrasting with "zero or more Criteria" in the same
// normative block.
//
// Both slices are defensively copied.
func NewClaim(
	id core.ValidationClaimID,
	claimType core.ClaimType,
	subject core.EngineeringSubjectRef,
	scope core.Scope,
	outcome core.ClaimOutcome,
	method core.ValidationMethod,
	criteria []core.CriterionRef,
	evidence []core.EvidenceArtifactRevisionRef,
	timestamp core.Timestamp,
	provenance core.Provenance,
) (Claim, error) {
	if id.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: claim id must not be zero", ErrInvalidValidationClaim)
	}
	if claimType.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: claim type must not be zero", ErrInvalidValidationClaim)
	}
	if subject.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: subject must not be zero", ErrInvalidValidationClaim)
	}
	if scope.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if outcome.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: claim outcome must not be zero", ErrInvalidValidationClaim)
	}
	if method.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: validation method must not be zero", ErrInvalidValidationClaim)
	}
	criteriaCopy, err := copyCriteria("NewClaim", criteria, ErrInvalidValidationClaim)
	if err != nil {
		return Claim{}, err
	}
	if len(evidence) == 0 {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: at least one evidence artifact revision must be cited", ErrInvalidValidationClaim)
	}
	evidenceCopy, err := copyEvidence("NewClaim", evidence, ErrInvalidValidationClaim)
	if err != nil {
		return Claim{}, err
	}
	if timestamp.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: timestamp must not be zero", ErrInvalidValidationClaim)
	}
	if provenance.IsZero() {
		return Claim{}, fmt.Errorf("validation: NewClaim: %w: provenance must not be zero", ErrInvalidValidationClaim)
	}
	if err := checkClaimTypeCriteria("NewClaim", claimType, subject, criteriaCopy); err != nil {
		return Claim{}, err
	}

	return Claim{
		id:         id,
		claimType:  claimType,
		subject:    subject,
		scope:      scope,
		outcome:    outcome,
		method:     method,
		criteria:   criteriaCopy,
		evidence:   evidenceCopy,
		timestamp:  timestamp,
		provenance: provenance,
	}, nil
}

// checkClaimTypeCriteria is the single Claim-Type-conditional validation
// path, shared by NewClaim, WithCriteria, and (through NewClaim)
// UnmarshalJSON. Having exactly one implementation is deliberate: a second
// copy could drift, and the criteria rules couple two fields, so they must be
// re-checked whenever either changes.
//
// Satisfaction (PEOS-006 Satisfaction Claim):
//
//   - at least one criterion must identify a Requirement or a Requirement
//     Artifact Revision -- "A Satisfaction Claim is a Validation Claim whose
//     criteria identify one or more Requirements", reinforced by "Each
//     Requirement or Requirement Artifact Revision SHALL appear as a Claim
//     criterion";
//   - if the Subject itself identifies a Requirement or Requirement Artifact
//     Revision, its Requirement identity must differ from that of every
//     Requirement-kind criterion -- "A Requirement SHALL NOT become both the
//     Claim subject and the same Claim's criterion", with the specification's
//     own non-conforming counterexample using the same Requirement Artifact
//     Revision in both positions.
//
// The identity comparison is at core.ArtifactID and is cross-level, because
// "the same Requirement" is identity-level language: an identity-level
// subject conflicts with a revision-level criterion of that Requirement, a
// revision-level subject conflicts with an identity-level criterion of it, and
// two different Revisions of one Requirement still conflict.
//
// Conformance (PEOS-006 Conformance Claim): at least one criterion is
// required -- "a Validation Claim whose criteria identify one or more"
// conformance rules, which "evaluates exactly one Subject against those
// criteria". No criterion-kind restriction is enforceable, because that
// clause's permitted set ends in "or other explicitly identified conformance
// rules"; a Product rule, an external rule, and an opaque forward-compatible
// criterion are all acceptable.
//
// Quality, Compliance, Template Conformance, and any Product-defined Claim
// Type: PEOS-006 imposes no additional criteria rule, so zero criteria are
// accepted. PEOS-007/008/009 may add their own rules in their own packets;
// none is inferred here.
//
// Every other Claim Type accepts zero criteria, per "Where zero criteria are
// identified, the Claim's outcome SHALL be interpreted strictly according to
// its stated Validation Method and basis."
func checkClaimTypeCriteria(
	caller string,
	claimType core.ClaimType,
	subject core.EngineeringSubjectRef,
	criteria []core.CriterionRef,
) error {
	switch claimType {
	case core.ClaimTypeSatisfaction:
		requirementCriteria := requirementCriterionArtifactIDs(criteria)
		if len(requirementCriteria) == 0 {
			return fmt.Errorf("validation: %s: %w: a satisfaction claim requires at least one criterion identifying a Requirement or Requirement Artifact Revision", caller, ErrInvalidSatisfactionClaim)
		}
		subjectID, ok := subjectRequirementArtifactID(subject)
		if !ok {
			return nil
		}
		if slices.Contains(requirementCriteria, subjectID) {
			return fmt.Errorf("validation: %s: %w: Requirement %q must not be both the claim subject and one of its own criteria", caller, ErrInvalidSatisfactionClaim, subjectID.String())
		}
		return nil

	case core.ClaimTypeConformance:
		if len(criteria) == 0 {
			return fmt.Errorf("validation: %s: %w: a conformance claim requires at least one criterion", caller, ErrInvalidConformanceClaim)
		}
		return nil

	default:
		return nil
	}
}

// requirementCriterionArtifactIDs returns the owning Requirement
// core.ArtifactID of every criterion that identifies a Requirement or a
// Requirement Artifact Revision, collapsing both levels to identity so the
// comparison in checkClaimTypeCriteria is cross-level.
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

// subjectRequirementArtifactID returns the owning Requirement
// core.ArtifactID of a Subject that identifies a Requirement or a Requirement
// Artifact Revision, and whether the Subject is of either kind. A Subject of
// any other kind (an Artifact Revision, a Decision Outcome, an Engineering
// Commitment, and so on) never conflicts with a Requirement criterion.
func subjectRequirementArtifactID(subject core.EngineeringSubjectRef) (core.ArtifactID, bool) {
	if ref, ok := subject.AsRequirement(); ok {
		return ref.ArtifactID(), true
	}
	if ref, ok := subject.AsRequirementRevision(); ok {
		return ref.ArtifactID(), true
	}
	return core.ArtifactID{}, false
}

// WithCriteria returns a copy of c with its evaluation criteria set to
// exactly the values given, in the order given, replacing any previous
// criteria. A zero-value core.CriterionRef is rejected, and the full
// Claim-Type-conditional validation is re-run against c's own Claim Type and
// Subject through the same path NewClaim uses.
//
// Passing an empty or nil slice declares zero criteria. That is valid for a
// general Claim Type, and invalid for both Satisfaction and Conformance --
// which is exactly why there is no WithoutCriteria method: WithCriteria(nil)
// already expresses removal, and a separate method would either duplicate
// this validation or bypass it.
//
// The receiver is never mutated, including when validation fails.
func (c Claim) WithCriteria(criteria []core.CriterionRef) (Claim, error) {
	cp, err := copyCriteria("Claim.WithCriteria", criteria, ErrInvalidValidationClaim)
	if err != nil {
		return Claim{}, err
	}
	if err := checkClaimTypeCriteria("Claim.WithCriteria", c.claimType, c.subject, cp); err != nil {
		return Claim{}, err
	}
	c.criteria = cp
	return c, nil
}

// WithExecutionRecords returns a copy of c with the relevant Validation
// Execution Records set to exactly the values given, in the order given. A
// zero-value reference is rejected. An empty or nil slice declares none:
// PEOS-006 requires these only "where applicable", and explicitly does not
// require "every Validation Claim to reference a Validation Execution
// Record".
func (c Claim) WithExecutionRecords(records []core.ValidationExecutionRecordRef) (Claim, error) {
	if len(records) == 0 {
		c.executionRecords = nil
		return c, nil
	}
	cp := make([]core.ValidationExecutionRecordRef, len(records))
	for idx, r := range records {
		if r.IsZero() {
			return Claim{}, fmt.Errorf("validation: Claim.WithExecutionRecords: %w: execution record reference must not be zero", ErrInvalidValidationClaim)
		}
		cp[idx] = r
	}
	c.executionRecords = cp
	return c, nil
}

// WithReasoning returns a copy of c with its reasoning or interpretation set.
// The value is trimmed and must be non-empty after trimming.
//
// PEOS-006 requires reasoning only "where the outcome is not mechanically
// determined". Whether a given outcome was mechanically determined is a
// semantic judgment this layer cannot make, so the field is optional and the
// obligation belongs to the Product contract.
func (c Claim) WithReasoning(reasoning string) (Claim, error) {
	trimmed, err := requireTrimmed("Claim.WithReasoning", "reasoning", reasoning, ErrInvalidValidationClaim)
	if err != nil {
		return Claim{}, err
	}
	c.reasoning = trimmed
	return c, nil
}

// WithoutReasoning returns a copy of c with its reasoning cleared.
func (c Claim) WithoutReasoning() Claim {
	c.reasoning = ""
	return c
}

// WithAuthority returns a copy of c with its authority set. PEOS-006 requires
// this only "where required by applicable governance", so it is optional.
// authority must be non-zero; use WithoutAuthority to clear it.
//
// An authority records who had the right to establish the Claim. It does not
// make the Claim a governance act: acceptance and certification remain
// PEOS-004 Decision Outcomes.
func (c Claim) WithAuthority(authority core.AuthorityRef) (Claim, error) {
	if authority.IsZero() {
		return Claim{}, fmt.Errorf("validation: Claim.WithAuthority: %w: authority must not be zero", ErrInvalidValidationClaim)
	}
	c.authority = authority
	return c, nil
}

// WithoutAuthority returns a copy of c with its authority cleared.
func (c Claim) WithoutAuthority() Claim {
	c.authority = core.AuthorityRef{}
	return c
}

// WithCorrection returns a copy of c declaring that c corrects, replaces, or
// invalidates an earlier Validation Claim, identified exactly.
//
// PEOS-006 defines this as "a minimal record-to-record reference" that "is
// owned by PEOS-006", "is not an Artifact Relation", "is not Artifact
// Supersession as defined by PEOS-002", "has no separate entity identity of
// its own", and "does not erase, delete, or rewrite the earlier Claim". The
// normative terms supersede/superseded/Supersession "SHALL NOT be used to
// describe Claim replacement", which is why core.CorrectionKind offers only
// correct, replace, and invalidate.
//
// Which Claim is currently applicable "is derived by identifying the most
// recent Claim that has not been replaced or invalidated by a later Claim.
// This determination is a derived view; it is never stored as a field" --
// neither on the Subject nor here.
func (c Claim) WithCorrection(correction core.RecordCorrectionRef[core.ValidationClaimRef]) (Claim, error) {
	if correction.IsZero() {
		return Claim{}, fmt.Errorf("validation: Claim.WithCorrection: %w: correction must not be zero", ErrInvalidValidationClaim)
	}
	if correction.Target().IsZero() {
		return Claim{}, fmt.Errorf("validation: Claim.WithCorrection: %w: correction target must not be zero", ErrInvalidValidationClaim)
	}
	c.correction = correction
	return c, nil
}

// WithoutCorrection returns a copy of c with its correction reference
// cleared.
func (c Claim) WithoutCorrection() Claim {
	c.correction = core.RecordCorrectionRef[core.ValidationClaimRef]{}
	return c
}

// WithExtension returns a copy of c with its extension data set.
func (c Claim) WithExtension(extension core.Extension) Claim {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c Claim) WithoutExtension() Claim {
	c.extension = core.Extension{}
	return c
}

// ID returns c's own Claim identity.
func (c Claim) ID() core.ValidationClaimID { return c.id }

// Ref returns a core.ValidationClaimRef citing c. This is the reference a
// later Claim's correction targets, and the reference a PEOS-004 Decision may
// cite as part of its Decision Basis.
func (c Claim) Ref() (core.ValidationClaimRef, error) {
	return core.NewValidationClaimRef(c.id)
}

// ClaimType returns c's Claim Type. Satisfaction, Conformance, Quality,
// Compliance, and Template Conformance are values of this vocabulary, not
// separate Go types.
func (c Claim) ClaimType() core.ClaimType { return c.claimType }

// Subject returns c's single engineering Subject. PEOS-006 permits exactly
// one, and the field is a single value rather than a slice, so a composite
// Subject is unrepresentable. Its participant level is carried by the
// returned core.EngineeringSubjectRef's own arm.
func (c Claim) Subject() core.EngineeringSubjectRef { return c.subject }

// Scope returns c's explicit scope. It is mandatory and therefore never
// absent on a valid Claim.
func (c Claim) Scope() core.Scope { return c.scope }

// Outcome returns c's recorded determination about its Subject. This is "the
// complete and only representation of what was determined"; there is no
// separate verdict.
func (c Claim) Outcome() core.ClaimOutcome { return c.outcome }

// Method returns the Validation Method or evaluation rule c applied.
func (c Claim) Method() core.ValidationMethod { return c.method }

// Criteria returns a defensive copy of the criteria c evaluated, in
// declaration order. A criterion is never a second Subject.
func (c Claim) Criteria() []core.CriterionRef {
	if len(c.criteria) == 0 {
		return nil
	}
	cp := make([]core.CriterionRef, len(c.criteria))
	copy(cp, c.criteria)
	return cp
}

// Evidence returns a defensive copy of the exact Evidence Artifact Revisions
// c relied upon, in declaration order. A valid Claim always cites at least
// one.
func (c Claim) Evidence() []core.EvidenceArtifactRevisionRef {
	return copyEvidenceOut(c.evidence)
}

// Timestamp returns the time c's assertion was made. This is a
// domain-meaningful time, distinct from the recorded-at time c's provenance
// may carry.
func (c Claim) Timestamp() core.Timestamp { return c.timestamp }

// Provenance returns c's provenance.
func (c Claim) Provenance() core.Provenance { return c.provenance }

// ExecutionRecords returns a defensive copy of the Validation Execution
// Records relevant to c, in declaration order.
func (c Claim) ExecutionRecords() []core.ValidationExecutionRecordRef {
	if len(c.executionRecords) == 0 {
		return nil
	}
	cp := make([]core.ValidationExecutionRecordRef, len(c.executionRecords))
	copy(cp, c.executionRecords)
	return cp
}

// Reasoning returns c's recorded reasoning or interpretation, and whether any
// is set.
func (c Claim) Reasoning() (string, bool) { return c.reasoning, c.reasoning != "" }

// Authority returns c's declared authority, and whether one is set.
func (c Claim) Authority() (core.AuthorityRef, bool) {
	return c.authority, !c.authority.IsZero()
}

// Correction returns c's declared correction reference to an earlier
// Validation Claim, and whether one is set.
func (c Claim) Correction() (core.RecordCorrectionRef[core.ValidationClaimRef], bool) {
	return c.correction, !c.correction.IsZero()
}

// Extension returns c's extension data.
func (c Claim) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c Claim) IsZero() bool {
	return c.id.IsZero() && c.claimType.IsZero() && c.subject.IsZero() && c.scope.IsZero() &&
		c.outcome.IsZero() && c.method.IsZero() && len(c.evidence) == 0 &&
		c.timestamp.IsZero() && c.provenance.IsZero()
}

type claimJSON struct {
	ID         core.ValidationClaimID             `json:"id"`
	ClaimType  core.ClaimType                     `json:"claim_type"`
	Subject    core.EngineeringSubjectRef         `json:"subject"`
	Scope      core.Scope                         `json:"scope"`
	Outcome    core.ClaimOutcome                  `json:"outcome"`
	Method     core.ValidationMethod              `json:"method"`
	Criteria   []core.CriterionRef                `json:"criteria,omitempty"`
	Evidence   []core.EvidenceArtifactRevisionRef `json:"evidence"`
	Timestamp  core.Timestamp                     `json:"timestamp"`
	Provenance core.Provenance                    `json:"provenance"`

	ExecutionRecords []core.ValidationExecutionRecordRef                `json:"execution_records,omitempty"`
	Reasoning        string                                             `json:"reasoning,omitempty"`
	Authority        *core.AuthorityRef                                 `json:"authority,omitempty"`
	Correction       *core.RecordCorrectionRef[core.ValidationClaimRef] `json:"correction,omitempty"`
	Extension        *core.Extension                                    `json:"extension,omitempty"`
}

// claimUnmarshalJSON mirrors claimJSON for decoding only, with one
// significant difference from the rest of this package: Criteria is captured
// as raw, undecoded bytes.
//
// For a Claim, absent / explicit null / empty array do NOT all mean the same
// thing, so they must be distinguishable. An absent "criteria" key means
// "zero criteria declared", which is valid for a general Claim Type and
// invalid for Satisfaction and Conformance. An explicit null is rejected
// outright rather than silently reinterpreted as zero criteria, because a
// caller writing null has said something different from writing nothing.
// This is the deliberate contrast with PlannedActivity's and
// ExecutionRecord's optional collections, where every Claim Type permits zero
// and the three inputs therefore carry no distinct meaning.
//
// Evidence keeps a plain typed field: an absent key, an explicit null, and an
// empty array all yield an empty slice, and all three must be rejected by the
// same "one or more" invariant, so the cases converge and need not be told
// apart. The optional single-value keys are captured raw so an explicit null
// can be rejected.
type claimUnmarshalJSON struct {
	ID         core.ValidationClaimID             `json:"id"`
	ClaimType  core.ClaimType                     `json:"claim_type"`
	Subject    core.EngineeringSubjectRef         `json:"subject"`
	Scope      core.Scope                         `json:"scope"`
	Outcome    core.ClaimOutcome                  `json:"outcome"`
	Method     core.ValidationMethod              `json:"method"`
	Criteria   json.RawMessage                    `json:"criteria"`
	Evidence   []core.EvidenceArtifactRevisionRef `json:"evidence"`
	Timestamp  core.Timestamp                     `json:"timestamp"`
	Provenance core.Provenance                    `json:"provenance"`

	ExecutionRecords []core.ValidationExecutionRecordRef `json:"execution_records"`
	Reasoning        json.RawMessage                     `json:"reasoning"`
	Authority        json.RawMessage                     `json:"authority"`
	Correction       json.RawMessage                     `json:"correction"`
	Extension        *core.Extension                     `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"id":..., "claim_type":..., "subject":...,
// "scope":..., "outcome":..., "method":..., "evidence":[...],
// "timestamp":..., "provenance":...}, plus whichever optional keys are set.
// criteria is omitted when empty.
//
// claim_type is ordinary content, not an envelope discriminator: there is no
// top-level polymorphic type key, because there is no sibling union to
// discriminate -- every Claim specialization is this same record.
//
// Deliberately absent from the wire form: "relation", "basis", "verdict",
// "status", any lifecycle or state key, any Artifact or Revision identity,
// any waiver key, and any satisfied / conformant / compliant boolean. Their
// absence is the structural proof of PEOS-006's Claim Is Not Artifact,
// Derived Satisfaction, and Validation-and-Lifecycle-Separation invariants.
func (c Claim) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("validation: marshal Claim: %w", ErrInvalidValidationClaim)
	}
	raw := claimJSON{
		ID:         c.id,
		ClaimType:  c.claimType,
		Subject:    c.subject,
		Scope:      c.scope,
		Outcome:    c.outcome,
		Method:     c.method,
		Evidence:   c.evidence,
		Timestamp:  c.timestamp,
		Provenance: c.provenance,
		Reasoning:  c.reasoning,
	}
	if len(c.criteria) > 0 {
		raw.Criteria = c.criteria
	}
	if len(c.executionRecords) > 0 {
		raw.ExecutionRecords = c.executionRecords
	}
	if !c.authority.IsZero() {
		raw.Authority = &c.authority
	}
	if !c.correction.IsZero() {
		raw.Correction = &c.correction
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, routing through NewClaim so
// that every Claim-Type-conditional rule is re-applied on decode -- a
// decoded Satisfaction Claim repeats the cross-level Requirement ArtifactID
// comparison, and a decoded Conformance Claim repeats the non-empty criteria
// rule. A decoded Claim can never be constructor-impossible.
//
// Missing-versus-null, stated exactly rather than assumed:
//
//   - id, claim_type, subject, scope, outcome, method, timestamp,
//     provenance: a missing key leaves the field zero and reaches NewClaim,
//     which rejects it through its owning sentinel (core.ErrInvalidScope for
//     scope, ErrInvalidValidationClaim for the rest). An explicit null
//     instead invokes that nested type's own UnmarshalJSON, which fails
//     there, so the error is wrapped here with ErrInvalidValidationClaim plus
//     that type's own sentinel. Both are rejected; the sentinel sets differ.
//   - criteria: absent means zero criteria declared -- valid for a general
//     Claim Type, rejected for Satisfaction and Conformance. An explicit null
//     is rejected outright. An empty array behaves as absent.
//   - evidence: absent, explicit null, and empty array all fail the
//     one-or-more invariant with ErrInvalidValidationClaim.
//   - execution_records: absent, explicit null, and empty array are
//     equivalent, all meaning "none declared".
//   - reasoning, authority, correction: a missing key means absent; an
//     explicit null is rejected.
//   - extension: null is equivalent to absent, per core.Extension's contract.
func (c *Claim) UnmarshalJSON(data []byte) error {
	var raw claimUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validation: unmarshal Claim: %w: %w", ErrInvalidValidationClaim, err)
	}

	if err := rejectNull("Claim", "criteria", raw.Criteria, ErrInvalidValidationClaim); err != nil {
		return err
	}
	var criteria []core.CriterionRef
	if len(raw.Criteria) > 0 {
		if err := json.Unmarshal(raw.Criteria, &criteria); err != nil {
			return fmt.Errorf("validation: unmarshal Claim: criteria: %w: %w", ErrInvalidValidationClaim, err)
		}
	}

	result, err := NewClaim(
		raw.ID, raw.ClaimType, raw.Subject, raw.Scope, raw.Outcome, raw.Method,
		criteria, raw.Evidence, raw.Timestamp, raw.Provenance,
	)
	if err != nil {
		return err
	}

	if len(raw.ExecutionRecords) > 0 {
		if result, err = result.WithExecutionRecords(raw.ExecutionRecords); err != nil {
			return err
		}
	}
	if err := rejectNull("Claim", "reasoning", raw.Reasoning, ErrInvalidValidationClaim); err != nil {
		return err
	}
	if len(raw.Reasoning) > 0 {
		var reasoning string
		if err := json.Unmarshal(raw.Reasoning, &reasoning); err != nil {
			return fmt.Errorf("validation: unmarshal Claim: reasoning: %w: %w", ErrInvalidValidationClaim, err)
		}
		if result, err = result.WithReasoning(reasoning); err != nil {
			return err
		}
	}
	if err := rejectNull("Claim", "authority", raw.Authority, ErrInvalidValidationClaim); err != nil {
		return err
	}
	if len(raw.Authority) > 0 {
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("validation: unmarshal Claim: %w: %w", ErrInvalidValidationClaim, err)
		}
		if result, err = result.WithAuthority(authority); err != nil {
			return err
		}
	}
	if err := rejectNull("Claim", "correction", raw.Correction, ErrInvalidValidationClaim); err != nil {
		return err
	}
	if len(raw.Correction) > 0 {
		var correction core.RecordCorrectionRef[core.ValidationClaimRef]
		if err := json.Unmarshal(raw.Correction, &correction); err != nil {
			return fmt.Errorf("validation: unmarshal Claim: %w: %w", ErrInvalidValidationClaim, err)
		}
		if result, err = result.WithCorrection(correction); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*c = result
	return nil
}
