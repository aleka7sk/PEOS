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
)
