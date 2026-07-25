package lifecycle

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels
// rather than comparing error values directly.
var (
	// ErrInvalidLifecycleDefinition is returned when a Definition fails to
	// satisfy its required fields.
	ErrInvalidLifecycleDefinition = errors.New("lifecycle: lifecycle definition is invalid")

	// ErrInvalidLifecycleDefinitionVersion is returned when a
	// DefinitionVersion fails to satisfy its required fields or internal
	// consistency rules.
	ErrInvalidLifecycleDefinitionVersion = errors.New("lifecycle: lifecycle definition version is invalid")

	// ErrInvalidState is returned when a State fails to satisfy its
	// required fields.
	ErrInvalidState = errors.New("lifecycle: state is invalid")

	// ErrInvalidStateID is returned when a StateID is constructed or
	// decoded from an empty or malformed value, or used where a non-zero
	// StateID is required.
	ErrInvalidStateID = errors.New("lifecycle: state id is invalid")

	// ErrInvalidTransition is returned when a TransitionDefinition fails
	// to satisfy its required fields.
	ErrInvalidTransition = errors.New("lifecycle: transition definition is invalid")

	// ErrInvalidTransitionID is returned when a TransitionID is
	// constructed or decoded from an empty or malformed value, or used
	// where a non-zero TransitionID is required.
	ErrInvalidTransitionID = errors.New("lifecycle: transition id is invalid")

	// ErrDuplicateStateID is returned when a State identifier that must be
	// unique within its scope (a Definition Version's State set, or a
	// single Transition Definition's source/target sets) appears more than
	// once.
	ErrDuplicateStateID = errors.New("lifecycle: duplicate state id")

	// ErrDuplicateTransitionID is returned when a Transition identifier
	// that must be unique within its owning Definition Version appears
	// more than once.
	ErrDuplicateTransitionID = errors.New("lifecycle: duplicate transition id")

	// ErrUnknownStateID is returned when a StateID is referenced (as an
	// initial state, a Transition source/target, or a State Assignment's
	// assigned state) but is not a member of the applicable Definition
	// Version's declared State set.
	ErrUnknownStateID = errors.New("lifecycle: state id is not a member of the referenced definition version")

	// ErrInvalidInitialState is returned when a DefinitionVersion's
	// initial-state set is empty or references a StateID outside its own
	// State set.
	ErrInvalidInitialState = errors.New("lifecycle: initial state is invalid")

	// ErrInvalidEntryTransition is returned when a DefinitionVersion's
	// entry TransitionID is zero or is not a member of its own Transition
	// set.
	ErrInvalidEntryTransition = errors.New("lifecycle: entry transition is invalid")

	// ErrInvalidStateAssignment is returned when a StateAssignment fails
	// to satisfy its required fields.
	ErrInvalidStateAssignment = errors.New("lifecycle: state assignment is invalid")

	// ErrInvalidTransitionRecord is returned when a TransitionRecord's
	// underlying Artifact fails to satisfy its required fields.
	ErrInvalidTransitionRecord = errors.New("lifecycle: transition record is invalid")

	// ErrInvalidTransitionRecordRevision is returned when a
	// TransitionRecordRevision or its Content fails to satisfy its
	// required fields or internal consistency rules.
	ErrInvalidTransitionRecordRevision = errors.New("lifecycle: transition record revision is invalid")

	// ErrInvalidTransitionOutcome is returned when a TransitionOutcome is
	// zero, or when a TransitionRecordContent's fields are inconsistent
	// with its declared outcome (see record.go).
	ErrInvalidTransitionOutcome = errors.New("lifecycle: transition outcome is invalid")

	// ErrInvalidFailure is returned when a Failure fails to satisfy its
	// required fields.
	ErrInvalidFailure = errors.New("lifecycle: failure is invalid")

	// ErrArtifactTypeMismatch is returned when an Artifact's declared
	// Artifact Type does not match the type a specialized constructor
	// requires.
	ErrArtifactTypeMismatch = errors.New("lifecycle: artifact type mismatch")

	// ErrArtifactIDMismatch is returned when a Revision's ArtifactID does
	// not match the ArtifactID of the specialized identity it is being
	// paired with.
	ErrArtifactIDMismatch = errors.New("lifecycle: artifact id mismatch")

	// ErrArtifactBindingMismatch is returned by
	// DefinitionVersion.ValidateArtifactBinding when a Definition's and a
	// DefinitionVersion's optional Artifact bindings are inconsistent with
	// each other, or when the DefinitionVersion's parent reference does not
	// identify the given Definition.
	ErrArtifactBindingMismatch = errors.New("lifecycle: artifact binding mismatch")

	// ErrInvalidLifecycleSupersession is returned when a
	// LifecycleDefinitionVersionSupersession fails to satisfy its required
	// fields or internal consistency rules.
	ErrInvalidLifecycleSupersession = errors.New("lifecycle: lifecycle definition version supersession is invalid")
)
