package lifecycle

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// StateID is a Lifecycle State's identifier, stable and unique within its
// owning Lifecycle Definition Version (PEOS-003 "State Identity"). It
// wraps a validated core.LocalKey rather than aliasing it: a StateID, a
// TransitionID (transition.go), and a bare core.LocalKey from an unrelated
// construct all share the same scoped-local-key wire shape but are never
// compile-time interchangeable with one another, because each is its own
// named struct type.
type StateID struct{ key core.LocalKey }

// NewStateID validates value and returns a StateID.
func NewStateID(value string) (StateID, error) {
	k, err := core.NewLocalKey(value)
	if err != nil {
		return StateID{}, fmt.Errorf("lifecycle: NewStateID: %w: %w", ErrInvalidStateID, err)
	}
	return StateID{key: k}, nil
}

// String returns the opaque identity value.
func (id StateID) String() string { return id.key.String() }

// IsZero reports whether id is the zero value.
func (id StateID) IsZero() bool { return id.key.IsZero() }

// Equal reports whether id and other identify the same State within a
// shared owning Definition Version. This package does not compare StateID
// values across different Definition Versions; see core.LocalKey's own
// documentation of this scoping limitation.
func (id StateID) Equal(other StateID) bool { return id.key == other.key }

// MarshalJSON encodes id as a JSON string.
func (id StateID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal StateID: %w", ErrInvalidStateID)
	}
	return json.Marshal(id.key.String())
}

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewStateID.
func (id *StateID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("lifecycle: unmarshal StateID: %w", err)
	}
	v, err := NewStateID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// TransitionID is a Transition Definition's identifier, stable and unique
// within its owning Lifecycle Definition Version (PEOS-003 "Transition
// Identity"). See StateID's own comment for why this wraps core.LocalKey
// rather than aliasing it.
type TransitionID struct{ key core.LocalKey }

// NewTransitionID validates value and returns a TransitionID.
func NewTransitionID(value string) (TransitionID, error) {
	k, err := core.NewLocalKey(value)
	if err != nil {
		return TransitionID{}, fmt.Errorf("lifecycle: NewTransitionID: %w: %w", ErrInvalidTransitionID, err)
	}
	return TransitionID{key: k}, nil
}

// String returns the opaque identity value.
func (id TransitionID) String() string { return id.key.String() }

// IsZero reports whether id is the zero value.
func (id TransitionID) IsZero() bool { return id.key.IsZero() }

// Equal reports whether id and other identify the same Transition
// Definition within a shared owning Definition Version.
func (id TransitionID) Equal(other TransitionID) bool { return id.key == other.key }

// MarshalJSON encodes id as a JSON string.
func (id TransitionID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal TransitionID: %w", ErrInvalidTransitionID)
	}
	return json.Marshal(id.key.String())
}

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewTransitionID.
func (id *TransitionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("lifecycle: unmarshal TransitionID: %w", err)
	}
	v, err := NewTransitionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}
