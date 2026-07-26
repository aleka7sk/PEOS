package decision

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels
// rather than comparing error values directly.
var (
	// ErrInvalidDecision is returned when a Decision fails to satisfy its
	// required fields or internal consistency rules.
	ErrInvalidDecision = errors.New("decision: decision is invalid")

	// ErrInvalidDecisionSubject is returned when a Decision's declared
	// Subjects contain a zero-value core.EngineeringSubjectRef.
	ErrInvalidDecisionSubject = errors.New("decision: decision subject is invalid")

	// ErrInvalidAuthority is returned when an Authority fails to satisfy
	// its required fields.
	ErrInvalidAuthority = errors.New("decision: authority is invalid")

	// ErrInvalidBasis is returned when a Basis fails to satisfy its
	// required fields.
	ErrInvalidBasis = errors.New("decision: basis is invalid")

	// ErrInvalidOutcome is returned when an Outcome fails to satisfy its
	// required fields.
	ErrInvalidOutcome = errors.New("decision: outcome is invalid")

	// ErrInvalidCommitmentEffect is returned when a CommitmentEffect is
	// zero where a non-zero value is required.
	ErrInvalidCommitmentEffect = errors.New("decision: commitment effect is invalid")

	// ErrInvalidCommitment is returned when a Commitment fails to satisfy
	// its required fields.
	ErrInvalidCommitment = errors.New("decision: commitment is invalid")

	// ErrInvalidAlternative is returned when an Alternative fails to
	// satisfy its required fields.
	ErrInvalidAlternative = errors.New("decision: alternative is invalid")

	// ErrInvalidDecisionRecord is returned when a Record fails to satisfy
	// its required fields.
	ErrInvalidDecisionRecord = errors.New("decision: decision record is invalid")

	// ErrInvalidDecisionRecordRevision is returned when a RecordRevision
	// or its Content fails to satisfy its required fields.
	ErrInvalidDecisionRecordRevision = errors.New("decision: decision record revision is invalid")

	// ErrArtifactTypeMismatch is returned when an Artifact's declared
	// Artifact Type does not match the type a specialized constructor
	// requires.
	ErrArtifactTypeMismatch = errors.New("decision: artifact type mismatch")

	// ErrArtifactIDMismatch is returned when a Revision's ArtifactID does
	// not match the ArtifactID of the specialized identity it is being
	// paired with.
	ErrArtifactIDMismatch = errors.New("decision: artifact id mismatch")

	// ErrInvalidDecisionSupersession is returned when a SupersessionID, a
	// SupersessionExtent, or a DecisionSupersession fails to satisfy its
	// required fields or internal consistency rules.
	ErrInvalidDecisionSupersession = errors.New("decision: decision supersession is invalid")

	// ErrInvalidDecisionInvalidation is returned when an InvalidationID,
	// an InvalidationSource, or a DecisionInvalidation fails to satisfy
	// its required fields or internal consistency rules.
	ErrInvalidDecisionInvalidation = errors.New("decision: decision invalidation is invalid")

	// ErrInvalidDecisionConflict is returned when a ConflictID or a
	// DecisionConflict fails to satisfy its required fields or internal
	// consistency rules.
	ErrInvalidDecisionConflict = errors.New("decision: decision conflict is invalid")

	// ErrInvalidConflictResolution is returned when a
	// ConflictResolutionID, a ResolutionMechanism, or a
	// ConflictResolution fails to satisfy its required fields.
	ErrInvalidConflictResolution = errors.New("decision: conflict resolution is invalid")

	// ErrInvalidDecisionRole is returned when a RoleKind or a Role fails
	// to satisfy its required fields.
	ErrInvalidDecisionRole = errors.New("decision: decision role is invalid")

	// ErrInvalidConsequence is returned when a Consequence fails to
	// satisfy its required fields.
	ErrInvalidConsequence = errors.New("decision: consequence is invalid")
)

// decodeOptionalExtension decodes a raw JSON "extension" field captured as
// json.RawMessage by an *UnmarshalJSON decode-only struct, treating an
// absent key (a zero-length raw message) as no extension and rejecting an
// explicit JSON null. This is shared by every type in this package that
// carries an optional core.Extension field.
func decodeOptionalExtension(raw json.RawMessage) (core.Extension, error) {
	if len(raw) == 0 {
		return core.Extension{}, nil
	}
	if string(raw) == "null" {
		return core.Extension{}, fmt.Errorf("extension must not be null")
	}
	var ext core.Extension
	if err := json.Unmarshal(raw, &ext); err != nil {
		return core.Extension{}, err
	}
	return ext, nil
}
