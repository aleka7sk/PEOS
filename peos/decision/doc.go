// Package decision implements the core, structural subset of PEOS-004's
// Decision Model on top of the stable peos/core package.
//
// This package implements Decision, Authority, Basis, Alternative, Outcome
// (with its Engineering Commitment value), Decision Record (with its
// Revision and Content), Decision Supersession, and Decision Invalidation.
// It deliberately does not implement Decision Conflict, Delegation,
// structured Basis (assumptions, constraints, uncertainty), Decision
// roles, Decision relation-type constants, or any approval/voting
// workflow -- see "Deferred" below.
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
// Decision Conflict, Delegation, structured Basis (assumptions,
// constraints, uncertainty), Decision roles, Decision Consequences,
// Decision relation-type constants, a decision repository, a supersession
// or invalidation history resolver, condition evaluation, and any
// approval or voting workflow are not implemented by this package.
package decision
