// Package decision implements the core, structural subset of PEOS-004's
// Decision Model on top of the stable peos/core package.
//
// This package implements Decision, Authority, Basis (with its Evidence,
// Assumption, Constraint, and Uncertainty components), Alternative,
// Outcome (with its Engineering Commitment value), Decision Record (with
// its Revision and Content), Decision Supersession, and Decision
// Invalidation. It deliberately does not implement Decision Conflict,
// Delegation, Decision roles, Decision relation-type constants, or any
// approval/voting workflow -- see "Deferred" below.
//
// # Decision is not an Artifact
//
// PEOS-004 states "A Decision MUST NOT be treated as equivalent to the
// Artifact that records it" (the Record Separation Invariant), and PEOS-002
// states "A Decision MUST NOT be reduced to an Artifact Type." Decision
// therefore carries its own core.DecisionID, independent of core.ArtifactID,
// and is never composed from core.Artifact.
//
// # Decision Record is an Artifact
//
// PEOS-004 states "A Decision Record is an Artifact that represents a
// Decision." Record therefore composes core.Artifact by named field,
// exactly like peos/requirement.Requirement, and declares
// ArtifactTypeDecisionRecord. A Decision and its Decision Record MAY have
// different identifiers (PEOS-002); Record carries its core.DecisionRef on
// the Record itself, not on RecordContent, so every Revision of a given
// Record is structurally guaranteed to name the same Decision.
//
// # Decision has no revision mechanism
//
// Unlike a Requirement or a Decision Record, a Decision is not an Artifact
// and is never revised in place. A material semantic change to a Decision's
// content is a new Decision, with its own core.DecisionID -- not a new
// Decision Revision. This package therefore defines no DecisionRevision
// type. A Decision Record's Artifact Revisions document a Decision's
// content over time (for example, as a Decision's own recognition and
// wording matures before an authority basis is established), not a
// Decision's own mutation.
//
// # Decision lifecycle state is out of scope
//
// PEOS-004 states "A Decision is a Lifecycle Subject" (PEOS-003), and
// core.NewLifecycleSubjectRefFromDecision already exists for that purpose.
// This package defines no Decision state field and no state vocabulary:
// Decision lifecycle state is governed entirely by peos/lifecycle. Decision
// Outcome (Outcome, this package) is a distinct concept from lifecycle
// state -- it records what was decided, not whether the Decision is
// currently applicable.
//
// # Engineering Commitment is a value, not an Entity
//
// PEOS-004 states "Engineering Commitment is a semantic component of the
// Decision Model and is not required to be a separate top-level PEOS
// Entity." Commitment therefore carries no identity of its own;
// core.EngineeringCommitmentRef is derived from the owning Decision's
// identity, outside this package's Commitment value.
//
// # Deliberate choices
//
// Provenance is optional on Decision: PEOS-004 never mentions provenance,
// unlike PEOS-002's requirement for every Artifact Relation. Alternative
// has no key: PEOS-004 defines no Alternative identity, and its own worked
// example expresses a selected option as prose in Outcome.Statement, not as
// a structural reference to a specific Alternative -- Outcome therefore
// also carries no selected-alternative link. RecordContent has no
// narrative field: the documentary payload of a Decision Record belongs to
// a core.Representation on its ArtifactRevision, not to typed content: the
// Decision Record field list in PEOS-004 is a SHOULD, not a MUST, and a
// zero-value RecordContent is valid.
//
// # Decision Basis may combine Evidence, Assumptions, Constraints, and Uncertainties
//
// A Basis may consist of any non-empty combination of its four content
// collections (PEOS-004:406-419's own "Decision Basis MAY include" list):
// Evidence, Assumptions, Constraints, and Uncertainties. Evidence alone,
// or a Basis carrying only Assumptions with no Evidence at all, are both
// conformant -- nothing in PEOS-004 privileges Evidence over the other
// three. Assumption, Constraint, and Uncertainty are typed value
// structures, not free-form strings, because their normative field
// carriage cannot fit in a single string: see each type's own doc comment
// for the exact PEOS-004 clause behind every field it exposes.
//
// Every optional field on Assumption, Constraint, and Uncertainty traces
// to an explicit PEOS-004 "SHOULD identify" clause -- Assumption's five
// (:458-464), Constraint's one (:491), and Uncertainty's none (:544's "MAY
// concern" list carries no identify obligation at all) -- and nothing
// beyond those clauses is modeled. In particular, neither Constraint nor
// Uncertainty carries an origin- or concern-category vocabulary:
// PEOS-004:478-489's "Constraint MAY originate from" list and :544-555's
// "Uncertainty MAY concern" list are structurally identical to
// PEOS-004:737-749's authority-origin list, which this package already
// declines to canonize (core.AuthorityRef's Kind is an open vocabulary
// with zero predeclared constants); a Product needing that classification
// carries it in Extension.
//
// Assumption.Uncertainty and Basis.Uncertainties are two distinct
// directions of relation, not the same fact recorded twice.
// Assumption.Uncertainty (PEOS-004:462, "An Assumption SHOULD identify...
// its uncertainty") qualifies exactly the one Assumption that carries it.
// Basis.Uncertainties (PEOS-004:542, "Known material uncertainty in a
// Decision Basis MUST be explicit") records a standalone known material
// uncertainty fact, independent of any particular Assumption -- even
// though :546 permits such a fact to "concern... assumptions" in general,
// that permission does not identify which Assumption it concerns, so it
// cannot discharge :462's per-Assumption obligation. Both may be present
// without duplication, and neither substitutes for the other.
//
// Constraint is Decision-Basis-scoped: PEOS-004:476 defines it as
// restricting "the available Alternatives or Decision Outcome" of the one
// Decision whose Basis carries it. It is not reused for a future
// Delegation's "applicable constraints" (PEOS-004:833): a delegation's
// constraints would limit a granted authority across all of that
// grantee's future Decisions -- a different axis, the same distinction
// this package already draws between Authority.requirements and
// Authority.bases.
//
// None of Assumption, Constraint, or Uncertainty carries identity, a Ref,
// a revision, or a lifecycle: PEOS-006's Claim Basis ("does not introduce
// independent Claim Basis identity, revision, or lifecycle") and
// PEOS-007's Quality Constraint ("value structures without independent
// identity, revision, or lifecycle") are the governing precedents.
//
// The Basis final-state validity guarantee -- that NewBasis, NewBasisFrom,
// every collection With* mutator, and UnmarshalJSON return a fully valid
// Basis whenever they return a nil error -- does not extend to the
// pre-existing, no-error-return WithExtension method: calling
// Basis{}.WithExtension(ext) on the zero Basis yields a value that still
// reports IsZero()==true. Extension alone is Product-specific data, not
// Decision Basis content in the PEOS-004 sense, so it does not by itself
// satisfy the "at least one of Evidence, Assumptions, Constraints, or
// Uncertainties" requirement. This is a pre-existing Packet F API edge,
// unchanged by Packet F.2.
//
// # Decision Supersession and Decision Invalidation are dedicated immutable governance records
//
// DecisionSupersession and DecisionInvalidation are each independently
// identified, immutable governance records -- not Artifacts, not
// Lifecycle Subjects, not Artifact Revisions, not Decision Revisions (see
// above), and not peos/relation.Relation specializations. PEOS-004 lists
// "supersedes" and "invalidates" among its Decision Relation Types, but
// Relation carries no extent, no effective condition/time, and no reason,
// so it cannot express what these records must; a Product MAY
// additionally record a Relation as an index over either fact, but that
// Relation is never authoritative and this package does not import
// peos/relation to create one.
//
// # Supersession extent is closed
//
// SupersessionExtent is a closed complete/partial distinction, not an
// open vocabulary: PEOS-004's Extension Model does not list supersession
// completeness among what a Product MAY extend, and explicitly forbids
// extensions from redefining the core meaning of supersession. A partial
// extent carries its own required, deterministically identifiable
// remaining scope; a complete extent carries none.
//
// # Effective condition is normative text, not an executable predicate
//
// DecisionSupersession.EffectiveCondition and
// DecisionInvalidation.EffectiveCondition are plain normative condition
// statements. This package defines no condition language and no
// evaluator; evaluating a condition against runtime or repository state
// is a later, deliberately deferred concern.
//
// # Preservation is by reference, not by duplication
//
// Both records preserve the original Decision's Outcome and Applicability
// by naming the Decision with core.DecisionRef, relying on Decision's own
// constructor-only Outcome and Applicability fields (see above) rather
// than copying their content. This package does not resolve or verify
// that reference against any repository: the guarantee that the named
// Decision is still the one originally superseded or invalidated is a
// repository-layer obligation this package does not check, the same
// limitation peos/lifecycle's own Supersession record documents for its
// superseding/superseded references.
//
// # Invalidation is non-retroactive by default
//
// DecisionInvalidation carries no field asserting that the invalidated
// Decision was never applicable. Its absence means exactly that:
// non-retroactive. PEOS-004 permits retroactive effect only when an
// applicable Product contract explicitly defines it; this package does
// not model that profile.
//
// # InvalidationSource chooses one canonical source
//
// PEOS-004 requires Invalidation to "identify the invalidating authority
// or Decision," and that "or" is not stated as an exclusive or in the
// specification text. InvalidationSource nonetheless accepts exactly one
// of the two, as a deliberate SDK architectural decision, not a
// specification requirement: when the source is a Decision, that
// Decision's own Authority is already required and non-zero, so carrying
// a second, independent AuthorityRef alongside it would duplicate that
// authority with no way for this package to detect a divergence between
// the two.
//
// # Decision Invalidation is not core.RecordCorrection
//
// Invalidation is a governance act -- a withdrawal of normative effect --
// never a typo correction, a representation fix, a metadata correction,
// or a non-normative clarification. core.RecordRef's own correction union
// deliberately excludes DecisionID; this package does not reuse
// core.RecordCorrection or core.CorrectionKindInvalidate for Decision
// Invalidation.
//
// # Deferred
//
// Decision Conflict, Delegation, Decision roles, Decision Consequences,
// Decision relation-type constants, a decision repository, a supersession
// or invalidation history resolver, condition evaluation, and any
// approval or voting workflow are not implemented by this package.
package decision
