package decision

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SupersessionID is the package-local identity of a DecisionSupersession.
// No specification outside PEOS-004 references a Decision Supersession
// record, and PEOS-004 does not itself name a dedicated identity for one,
// so this identity is not promoted to peos/core -- it is meaningful only
// within this package. SupersessionID and InvalidationID deliberately do
// not share a common wrapper type (and are not core.ImmutableRecordID,
// whose doc comment describes a single shared type for record families a
// later packet MAY still assign a dedicated identity to): a shared type
// would make the two sibling identities mutually assignable, which is
// exactly what peos/core/identity.go's own field-name-per-type convention
// exists to prevent.
type SupersessionID struct{ decisionSupersessionID string }

// NewSupersessionID validates value and returns a SupersessionID.
// Surrounding whitespace is trimmed; the result is rejected if empty.
func NewSupersessionID(value string) (SupersessionID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return SupersessionID{}, fmt.Errorf("decision: NewSupersessionID: %w", ErrInvalidDecisionSupersession)
	}
	return SupersessionID{decisionSupersessionID: trimmed}, nil
}

// String returns the opaque identity value.
func (id SupersessionID) String() string { return id.decisionSupersessionID }

// IsZero reports whether id is the zero value.
func (id SupersessionID) IsZero() bool { return id.decisionSupersessionID == "" }

// Equal reports whether id and other carry the same identity value.
func (id SupersessionID) Equal(other SupersessionID) bool {
	return id.decisionSupersessionID == other.decisionSupersessionID
}

// MarshalJSON encodes id as a JSON string.
func (id SupersessionID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("decision: marshal SupersessionID: %w", ErrInvalidDecisionSupersession)
	}
	return json.Marshal(id.decisionSupersessionID)
}

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewSupersessionID.
func (id *SupersessionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("decision: unmarshal SupersessionID: %w", err)
	}
	v, err := NewSupersessionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// InvalidationID is the package-local identity of a DecisionInvalidation.
// See SupersessionID's own doc comment for why this identity is
// package-local and structurally distinct rather than a shared wrapper.
type InvalidationID struct{ decisionInvalidationID string }

// NewInvalidationID validates value and returns an InvalidationID.
// Surrounding whitespace is trimmed; the result is rejected if empty.
func NewInvalidationID(value string) (InvalidationID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return InvalidationID{}, fmt.Errorf("decision: NewInvalidationID: %w", ErrInvalidDecisionInvalidation)
	}
	return InvalidationID{decisionInvalidationID: trimmed}, nil
}

// String returns the opaque identity value.
func (id InvalidationID) String() string { return id.decisionInvalidationID }

// IsZero reports whether id is the zero value.
func (id InvalidationID) IsZero() bool { return id.decisionInvalidationID == "" }

// Equal reports whether id and other carry the same identity value.
func (id InvalidationID) Equal(other InvalidationID) bool {
	return id.decisionInvalidationID == other.decisionInvalidationID
}

// MarshalJSON encodes id as a JSON string.
func (id InvalidationID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("decision: marshal InvalidationID: %w", ErrInvalidDecisionInvalidation)
	}
	return json.Marshal(id.decisionInvalidationID)
}

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewInvalidationID.
func (id *InvalidationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("decision: unmarshal InvalidationID: %w", err)
	}
	v, err := NewInvalidationID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}
