package requirement

import (
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// This file holds the shared, unexported foundation every Requirement
// relationship wrapper (Derivation, and later Refinement, Decomposition,
// Dependency, Conflict, and Requirement Supersession) is built on top of
// (PEOS-005 §17, §17.1-§17.4). See doc.go for the normative rationale.
//
// Every wrapper composes relation.Relation rather than duplicating its
// fields: relation.Relation remains the sole source of truth for relation
// type, source, target, provenance, scope, and extension. A wrapper's own
// constructor always builds the inner relation.Relation itself from typed
// parts -- it never accepts a caller-supplied relation.Relation -- so a
// wrapper can never be constructed carrying a relation type other than
// the one that constructor's own type declares.

// requirementRevisionParticipant validates ref and returns the
// core.EngineeringSubjectRef identifying it at Requirement Artifact
// Revision level (PEOS-005 §17.3: relationship meaning that depends on
// represented required engineering intent must identify the applicable
// Requirement Artifact Revision, not merely the Requirement identity).
func requirementRevisionParticipant(ref core.RequirementArtifactRevisionRef) (core.EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: requirement artifact revision must not be zero", ErrInvalidRequirementRelation)
	}
	subject, err := core.EngineeringSubjectRefFromRequirementRevision(ref)
	if err != nil {
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: %w", ErrInvalidRequirementRelation, err)
	}
	return subject, nil
}

// asRequirementRevisionParticipant unwraps subject as a participant
// identified at Requirement Artifact Revision level, rejecting a
// participant identified only at Requirement identity level or at any
// other subject kind (PEOS-005 §17.3, non-conforming pattern §36.14:
// "Representing a Requirement relationship without identifying whether a
// participant refers to the Requirement identity or to a specific
// Requirement Artifact Revision").
func asRequirementRevisionParticipant(subject core.EngineeringSubjectRef) (core.RequirementArtifactRevisionRef, error) {
	ref, ok := subject.AsRequirementRevision()
	if !ok {
		return core.RequirementArtifactRevisionRef{}, fmt.Errorf("%w: participant must identify a Requirement Artifact Revision, not a Requirement identity or another subject kind", ErrInvalidRequirementRelation)
	}
	return ref, nil
}

// checkRelationType rejects r if it does not declare exactly want. This
// is the decode-time defense against a decoded relation.Relation
// carrying a relation type other than the one this wrapper's own
// constructor always assigns (see this file's own doc comment above).
func checkRelationType(r relation.Relation, want core.RelationType) error {
	if r.RelationType() != want {
		return fmt.Errorf("%w: relation type mismatch", ErrInvalidRequirementRelation)
	}
	return nil
}

// checkDistinctParticipants rejects a and b naming the same Requirement
// Artifact Revision (PEOS-005 §18.1: "A Requirement Artifact Revision
// SHALL NOT be directly ... derived from itself"; the same direct
// self-reference prohibition recurs per relation type in §19.1, §20.1,
// and §23.1). It checks only direct self-reference: transitive cycles
// are a repository-level obligation this package cannot discharge (see
// doc.go).
func checkDistinctParticipants(a, b core.RequirementArtifactRevisionRef) error {
	if a == b {
		return fmt.Errorf("%w: source and target must not be the same Requirement Artifact Revision", ErrInvalidRequirementRelation)
	}
	return nil
}

// requireRelationScope returns r's declared scope, rejecting its absence
// (PEOS-005 §19: "the scope within which compatibility is asserted";
// §20: "the scope of that parent-to-subordinate association" -- both
// mandatory, unlike Derivation's optional scope). Used by every
// relationship type whose own PEOS-005 section makes scope mandatory
// (Refinement, Decomposition, and later Conflict and Requirement
// Supersession); Derivation and Dependency do not use this helper.
func requireRelationScope(r relation.Relation) (core.Scope, error) {
	scope, ok := r.Scope()
	if !ok {
		return core.Scope{}, fmt.Errorf("%w: scope must not be absent", ErrInvalidRequirementRelation)
	}
	return scope, nil
}

// checkDistinctRequirementIdentity rejects a and b naming Requirement
// Artifact Revisions of the same Requirement identity, regardless of
// revision. This rule is stronger than checkDistinctParticipants's
// revision-level check, and applies to any relation type whose own
// PEOS-005 section requires Requirement-identity distinctness between
// its two participants -- Decomposition (§20.1: "A subordinate
// Requirement identity SHALL remain distinct from the parent Requirement
// identity") and Derivation (§18: "A derived Requirement SHALL possess
// its own identity"; "A derived Requirement SHALL NOT inherit the
// identity of a source Requirement" -- PEOS-009 :664/:758 establishes
// that inheriting an identity means sharing it, i.e. an equal ArtifactID,
// which is the same condition §20.1 states in different words; see
// Packet G.1.1's audit and doc.go's own "Requirement-identity
// distinctness" section for the full three-way comparison against
// Refinement, which imposes no such rule).
//
// sentinel is the caller's own error sentinel (ErrInvalidDerivation,
// ErrInvalidDecomposition, ...): this helper is relation-type-agnostic
// and does not itself decide which relationship's failure this is.
func checkDistinctRequirementIdentity(a, b core.RequirementArtifactRevisionRef, sentinel error) error {
	if a.ArtifactID() == b.ArtifactID() {
		return fmt.Errorf("%w: Requirement identity must not be shared between the two participants", sentinel)
	}
	return nil
}

// requirementParticipantSubject converts p into the
// core.EngineeringSubjectRef relation.Relation stores as its source or
// target, rejecting a zero participant. Used by Dependency and Conflict,
// whose participants MAY identify either Requirement identity level or
// Requirement Artifact Revision level (PEOS-005 §21.1, §22.1), unlike the
// revision-only participants requirementRevisionParticipant handles.
func requirementParticipantSubject(p RequirementParticipant) (core.EngineeringSubjectRef, error) {
	switch {
	case p.IsZero():
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: participant must not be zero", ErrInvalidRequirementRelation)
	case p.IsRequirement():
		ref, _ := p.AsRequirement()
		subject, err := core.EngineeringSubjectRefFromRequirement(ref)
		if err != nil {
			return core.EngineeringSubjectRef{}, fmt.Errorf("%w: %w", ErrInvalidRequirementRelation, err)
		}
		return subject, nil
	default:
		ref, _ := p.AsRequirementRevision()
		subject, err := core.EngineeringSubjectRefFromRequirementRevision(ref)
		if err != nil {
			return core.EngineeringSubjectRef{}, fmt.Errorf("%w: %w", ErrInvalidRequirementRelation, err)
		}
		return subject, nil
	}
}

// asRequirementParticipant unwraps subject as a RequirementParticipant,
// accepting both Requirement identity-level and Requirement Artifact
// Revision-level kinds and rejecting every other subject kind (PEOS-005
// §21.1, §22.1).
func asRequirementParticipant(subject core.EngineeringSubjectRef) (RequirementParticipant, error) {
	if ref, ok := subject.AsRequirement(); ok {
		return NewRequirementParticipantFromRequirement(ref)
	}
	if ref, ok := subject.AsRequirementRevision(); ok {
		return NewRequirementParticipantFromRequirementRevision(ref)
	}
	return RequirementParticipant{}, fmt.Errorf("%w: participant must identify a Requirement identity or a Requirement Artifact Revision, not another subject kind", ErrInvalidRequirementRelation)
}
