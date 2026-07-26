package requirement

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels
// rather than comparing error values directly.
var (
	// ErrInvalidRequirement is returned when a Requirement is constructed
	// from a zero-value core.Artifact.
	ErrInvalidRequirement = errors.New("requirement: requirement is invalid")

	// ErrRequirementArtifactTypeMismatch is returned when a Requirement is
	// constructed from a non-zero core.Artifact whose declared Artifact
	// Type is not ArtifactTypeRequirement.
	ErrRequirementArtifactTypeMismatch = errors.New("requirement: artifact type is not requirement")

	// ErrInvalidRequirementRevision is returned when a Revision is
	// constructed from a zero-value Requirement or a zero-value
	// core.ArtifactRevision.
	ErrInvalidRequirementRevision = errors.New("requirement: requirement revision is invalid")

	// ErrRequirementArtifactIDMismatch is returned when a Revision's core
	// Artifact Revision refers to a different Artifact than the
	// Requirement it is being paired with.
	ErrRequirementArtifactIDMismatch = errors.New("requirement: artifact id mismatch between requirement and revision")

	// ErrMissingRequirementContent is returned when a Revision is
	// constructed from a zero-value Content, or when Content itself is
	// marshaled while zero.
	ErrMissingRequirementContent = errors.New("requirement: requirement content is missing")

	// ErrInvalidStatement is returned when a Statement's text is empty or
	// whitespace-only, or when Content's required Statement list is empty
	// or contains a zero-value Statement.
	ErrInvalidStatement = errors.New("requirement: statement is invalid")

	// ErrMissingRequirementSubject is returned when Content's required
	// Subject list is empty or contains a zero-value subject.
	ErrMissingRequirementSubject = errors.New("requirement: at least one subject is required")

	// ErrInvalidSubjectCombination is returned when a SubjectCombination
	// is constructed from a zero-value core.VocabularyValue, or when
	// Content is constructed with a zero-value SubjectCombination.
	ErrInvalidSubjectCombination = errors.New("requirement: subject combination is invalid")

	// ErrInvalidApplicability is returned when an Applicability value
	// fails its own construction or decoding rules, or when Content is
	// constructed with a zero-value Applicability.
	ErrInvalidApplicability = errors.New("requirement: applicability is invalid")

	// ErrInvalidOrigin is returned when an OriginRef's kind is zero or its
	// note is empty after trimming.
	ErrInvalidOrigin = errors.New("requirement: origin is invalid")

	// ErrInvalidAuthority is returned when a zero-value core.AuthorityRef
	// is supplied to Content.WithAuthorities. core.AuthorityRef validates
	// its own fields on construction; this sentinel exists only for the
	// reuse boundary at Content's own list-level check.
	ErrInvalidAuthority = errors.New("requirement: authority is invalid")

	// ErrInvalidClassification is returned when a Classification is
	// constructed from a zero-value core.VocabularyValue.
	ErrInvalidClassification = errors.New("requirement: classification is invalid")

	// ErrDuplicateClassification is returned when Content's declared
	// classifications contain the same value more than once.
	ErrDuplicateClassification = errors.New("requirement: duplicate classification")

	// ErrInvalidRationale is returned when a Rationale's text is empty or
	// whitespace-only.
	ErrInvalidRationale = errors.New("requirement: rationale is invalid")

	// ErrInvalidRequirementRelation is the shared foundation sentinel for
	// every Requirement relationship wrapper (Derivation, and later
	// Refinement, Decomposition, Dependency, Conflict, and Requirement
	// Supersession -- PEOS-005 §17-§23). It is returned for participant
	// zero-value or wrong-level errors, relation type mismatch, missing
	// provenance, non-distinct source/target, a malformed nested
	// relation.Relation, and zero-value marshal. Relationship-content
	// errors specific to one relation type (for example, Derivation's
	// rationale) use their own sentinel instead.
	ErrInvalidRequirementRelation = errors.New("requirement: requirement relation is invalid")

	// ErrInvalidDerivation is returned when a Derivation's rationale is
	// empty or whitespace-only.
	ErrInvalidDerivation = errors.New("requirement: derivation is invalid")

	// ErrInvalidDecomposition is returned when a Decomposition's
	// subordinate Requirement identity is the same as its parent
	// Requirement identity (PEOS-005 §20.1). There is no equivalent
	// ErrInvalidRefinement: Refinement has no type-specific failure mode
	// beyond the shared foundation ErrInvalidRequirementRelation already
	// covers.
	ErrInvalidDecomposition = errors.New("requirement: decomposition is invalid")

	// ErrInvalidDependency is returned when a Dependency's nature is
	// empty or whitespace-only.
	ErrInvalidDependency = errors.New("requirement: dependency is invalid")

	// ErrInvalidConflict is returned when a Conflict's nature is empty or
	// whitespace-only. Conflict's participant-distinctness requirement
	// (PEOS-005 §22.1) uses the shared ErrInvalidRequirementRelation
	// instead, since it is a participant-shape rule of the same kind as
	// checkDistinctParticipants, not content specific to Conflict.
	ErrInvalidConflict = errors.New("requirement: conflict is invalid")

	// ErrInvalidGovernanceAction is returned when a GovernanceAction is
	// constructed or decoded from a zero arm payload, or decoded with an
	// unrecognized kind discriminator or a missing/null ref. GovernanceAction
	// is not itself a relationship type (PEOS-005 §23 and §27 both define
	// it identically), so it carries its own sentinel rather than reusing
	// ErrInvalidRequirementRelation or ErrInvalidRequirementSupersession --
	// see governance.go's own doc comment.
	ErrInvalidGovernanceAction = errors.New("requirement: governance action is invalid")

	// ErrInvalidRequirementSupersession is returned for Supersession's own
	// type-specific failures: a LifecycleConsequence left in its zero
	// (unstated) state, an "identified" LifecycleConsequence with an empty
	// or whitespace-only description, or a "none" LifecycleConsequence
	// carrying a non-empty description (PEOS-005 §23.1). Supersession's
	// participant, provenance, scope, and self-supersession failures use
	// the shared ErrInvalidRequirementRelation instead.
	ErrInvalidRequirementSupersession = errors.New("requirement: requirement supersession is invalid")
)
