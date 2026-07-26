package requirement

import (
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

type requirementParticipantKind string

const (
	requirementParticipantRequirement         requirementParticipantKind = "requirement"
	requirementParticipantRequirementRevision requirementParticipantKind = "requirement_revision"
)

// RequirementParticipant is a closed union naming one participant of a
// Dependency or Conflict relationship, at either Requirement identity
// level or Requirement Artifact Revision level (PEOS-005 §21.1, §22.1:
// "each participant SHALL explicitly identify whether it refers to
// Requirement identity level or Requirement Artifact Revision level").
// Unlike Derivation, Refinement, and Decomposition -- which fix both
// participants at Requirement Artifact Revision level -- Dependency and
// Conflict permit either level independently per participant, including
// mixed pairs.
//
// This mirrors decision.InvalidationSource, the established closed
// two-arm union pattern in this repository: a dedicated Requirement-typed
// union, rather than accepting core.EngineeringSubjectRef directly, keeps
// a wrong subject kind (Decision, Template, ...) a compile error at
// Dependency's and Conflict's own constructors instead of a runtime
// failure discovered only on decode.
//
// RequirementParticipant carries no JSON methods: it never reaches the
// wire directly. Dependency and Conflict store each participant inside
// their composed relation.Relation as a core.EngineeringSubjectRef, whose
// own {"kind":...,"ref":...} envelope already expresses the level
// distinction this type models.
type RequirementParticipant struct {
	kind        requirementParticipantKind
	requirement core.RequirementRef
	revision    core.RequirementArtifactRevisionRef
}

// NewRequirementParticipantFromRequirement validates ref and returns a
// RequirementParticipant identifying it at Requirement identity level.
func NewRequirementParticipantFromRequirement(ref core.RequirementRef) (RequirementParticipant, error) {
	if ref.IsZero() {
		return RequirementParticipant{}, fmt.Errorf("requirement: NewRequirementParticipantFromRequirement: %w: requirement must not be zero", ErrInvalidRequirementRelation)
	}
	return RequirementParticipant{kind: requirementParticipantRequirement, requirement: ref}, nil
}

// NewRequirementParticipantFromRequirementRevision validates ref and
// returns a RequirementParticipant identifying it at Requirement Artifact
// Revision level.
func NewRequirementParticipantFromRequirementRevision(ref core.RequirementArtifactRevisionRef) (RequirementParticipant, error) {
	if ref.IsZero() {
		return RequirementParticipant{}, fmt.Errorf("requirement: NewRequirementParticipantFromRequirementRevision: %w: requirement artifact revision must not be zero", ErrInvalidRequirementRelation)
	}
	return RequirementParticipant{kind: requirementParticipantRequirementRevision, revision: ref}, nil
}

// Kind returns p's discriminator, "requirement" or "requirement_revision".
func (p RequirementParticipant) Kind() string { return string(p.kind) }

// IsRequirement reports whether p identifies a Requirement identity.
func (p RequirementParticipant) IsRequirement() bool {
	return p.kind == requirementParticipantRequirement
}

// IsRequirementRevision reports whether p identifies a Requirement
// Artifact Revision.
func (p RequirementParticipant) IsRequirementRevision() bool {
	return p.kind == requirementParticipantRequirementRevision
}

// AsRequirement returns p's Requirement reference, and whether p is the
// Requirement identity-level arm.
func (p RequirementParticipant) AsRequirement() (core.RequirementRef, bool) {
	if p.kind != requirementParticipantRequirement {
		return core.RequirementRef{}, false
	}
	return p.requirement, true
}

// AsRequirementRevision returns p's Requirement Artifact Revision
// reference, and whether p is the Requirement Artifact Revision-level
// arm.
func (p RequirementParticipant) AsRequirementRevision() (core.RequirementArtifactRevisionRef, bool) {
	if p.kind != requirementParticipantRequirementRevision {
		return core.RequirementArtifactRevisionRef{}, false
	}
	return p.revision, true
}

// RequirementID returns the owning Requirement identity, available at
// both participant levels: a Requirement Artifact Revision always
// belongs to exactly one Requirement identity (PEOS-002).
func (p RequirementParticipant) RequirementID() core.ArtifactID {
	if p.kind == requirementParticipantRequirement {
		return p.requirement.ArtifactID()
	}
	return p.revision.ArtifactID()
}

// IsZero reports whether p is the zero value.
func (p RequirementParticipant) IsZero() bool { return p.kind == "" }
