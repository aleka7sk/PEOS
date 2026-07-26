// Package decision implements the core, structural subset of PEOS-004's
// Decision Model on top of the stable peos/core package.
//
// This package implements Decision, Authority, Basis, Alternative, Outcome
// (with its Engineering Commitment value), and Decision Record (with its
// Revision and Content). It deliberately does not implement Decision
// Supersession, Decision Invalidation, Decision Conflict, Delegation, or
// any approval/voting workflow -- see "Deferred" below.
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
// # Deferred
//
// Decision Supersession, Decision Invalidation, Decision Conflict,
// Delegation, structured Basis (assumptions, constraints, uncertainty),
// Decision roles, Decision relation types, a decision repository, and any
// approval or voting workflow are not implemented by this package.
package decision
