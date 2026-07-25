package core

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels
// rather than comparing error values directly.
var (
	// ErrEmptyIdentity is returned when an identity constructor receives a
	// value that is empty after trimming surrounding whitespace.
	ErrEmptyIdentity = errors.New("core: identity value must not be empty")

	// ErrInvalidVocabularyValue is returned when a namespaced vocabulary
	// value is missing its namespace or its value component.
	ErrInvalidVocabularyValue = errors.New("core: vocabulary value is invalid")

	// ErrInvalidReferenceDiscriminator is returned when a reference is
	// decoded or constructed with a discriminator that does not match any
	// known reference kind, in a context where an unknown kind cannot be
	// preserved opaquely.
	ErrInvalidReferenceDiscriminator = errors.New("core: reference discriminator is invalid")

	// ErrMissingRevisionID is returned when a revision-level reference is
	// constructed or decoded without a revision identity component.
	ErrMissingRevisionID = errors.New("core: revision-level reference requires a revision id")

	// ErrUnexpectedRevisionID is returned when an identity-level reference
	// is constructed or decoded with a revision identity component present.
	ErrUnexpectedRevisionID = errors.New("core: identity-level reference must not carry a revision id")

	// ErrInvalidTimestamp is returned when a timestamp value is zero, or
	// otherwise fails PEOS timestamp validation.
	ErrInvalidTimestamp = errors.New("core: timestamp is invalid")

	// ErrInvalidScope is returned when a Scope is constructed without a
	// valid kind or with an empty expression where one is required.
	ErrInvalidScope = errors.New("core: scope is invalid")

	// ErrDuplicateExtensionNamespace is returned when an extension
	// container is constructed with the same namespace supplied more than
	// once.
	ErrDuplicateExtensionNamespace = errors.New("core: duplicate extension namespace")

	// ErrInvalidCorrectionReference is returned when a correction
	// reference is constructed with an unknown correction kind, or with a
	// record reference that fails validation.
	ErrInvalidCorrectionReference = errors.New("core: correction reference is invalid")

	// ErrInvalidPayload is returned when a tagged union (EngineeringSubjectRef
	// or CriterionRef) is decoded with a discriminator whose required
	// payload fields are missing or inconsistent.
	ErrInvalidPayload = errors.New("core: payload does not match discriminator")

	// ErrInvalidArtifact is returned when an Artifact fails to satisfy its
	// required fields. It is wrapped alongside a more specific sentinel
	// (for example ErrEmptyIdentity) where one applies, so callers may
	// check either the general or the specific condition.
	ErrInvalidArtifact = errors.New("core: artifact is invalid")

	// ErrInvalidArtifactRevision is returned when an Artifact Revision
	// fails to satisfy its required fields.
	ErrInvalidArtifactRevision = errors.New("core: artifact revision is invalid")

	// ErrInvalidRepresentation is returned when a Representation fails to
	// satisfy its required fields.
	ErrInvalidRepresentation = errors.New("core: representation is invalid")

	// ErrInvalidOrigin is returned when an Origin fails to satisfy its
	// required fields.
	ErrInvalidOrigin = errors.New("core: origin is invalid")

	// ErrInvalidIntegrityIdentity is returned when an Integrity Identity
	// fails to satisfy its required fields.
	ErrInvalidIntegrityIdentity = errors.New("core: integrity identity is invalid")

	// ErrDuplicateArtifactRole is returned when an Artifact's declared
	// roles contain the same role more than once.
	ErrDuplicateArtifactRole = errors.New("core: duplicate artifact role")

	// ErrDuplicateRepresentationRole is returned when a Representation's
	// classification contains the same role more than once.
	ErrDuplicateRepresentationRole = errors.New("core: duplicate representation role")

	// ErrMissingRepresentationContent is returned when a Representation
	// Content constructor receives an empty or zero-value payload.
	ErrMissingRepresentationContent = errors.New("core: representation content is missing")

	// ErrInvalidRepresentationComponent is returned when a
	// RepresentationComponentRef fails to satisfy its required fields, or
	// when composed Representation Content receives a zero-value
	// component.
	ErrInvalidRepresentationComponent = errors.New("core: representation component is invalid")

	// ErrDuplicateIntegrityProtectedScope is returned when an Integrity
	// Identity's declared protected scopes contain the same scope more
	// than once.
	ErrDuplicateIntegrityProtectedScope = errors.New("core: duplicate integrity protected scope")
)
