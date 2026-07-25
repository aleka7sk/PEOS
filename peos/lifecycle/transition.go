package lifecycle

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// TransitionDefinition is an identifiable lifecycle rule permitting a
// Lifecycle Subject to move from one State to another under explicit
// conditions (PEOS-003 "Transition"), scoped to its owning Lifecycle
// Definition Version. This type models only the structural shape PEOS-003
// requires at minimum: identity, one or more source States, and one or
// more target States.
//
// Guard, Effect, Trigger, and any executable expression semantics are
// intentionally not modeled here -- see doc.go. This package therefore
// does not expose a placeholder Guard/Effect/Trigger field, interface, or
// Evaluate method; adding one now would invent normative content PEOS-003
// does not define.
type TransitionDefinition struct {
	id           TransitionID
	sourceStates []StateID
	targetStates []StateID
	extension    core.Extension
}

// NewTransitionDefinition validates id, sourceStates, and targetStates and
// returns a TransitionDefinition.
//
// Both sourceStates and targetStates must be non-empty and free of exact
// duplicates. A self-transition (the same StateID appearing in both sets)
// is permitted -- PEOS-003 explicitly allows same-State Transitions
// ("A same-State Transition MUST have explicit normative meaning"). Two or
// more TransitionDefinitions sharing the same source/target pair are also
// permitted: PEOS-003 does not cap how many distinct Transitions may
// connect the same pair of States, each with its own TransitionID.
func NewTransitionDefinition(id TransitionID, sourceStates, targetStates []StateID) (TransitionDefinition, error) {
	if id.IsZero() {
		return TransitionDefinition{}, fmt.Errorf("lifecycle: NewTransitionDefinition: %w", ErrInvalidTransitionID)
	}
	sources, err := dedupStateIDs(sourceStates)
	if err != nil {
		return TransitionDefinition{}, fmt.Errorf("lifecycle: NewTransitionDefinition: source states: %w", err)
	}
	if len(sources) == 0 {
		return TransitionDefinition{}, fmt.Errorf("lifecycle: NewTransitionDefinition: %w: source states must not be empty", ErrInvalidTransition)
	}
	targets, err := dedupStateIDs(targetStates)
	if err != nil {
		return TransitionDefinition{}, fmt.Errorf("lifecycle: NewTransitionDefinition: target states: %w", err)
	}
	if len(targets) == 0 {
		return TransitionDefinition{}, fmt.Errorf("lifecycle: NewTransitionDefinition: %w: target states must not be empty", ErrInvalidTransition)
	}
	return TransitionDefinition{id: id, sourceStates: sources, targetStates: targets}, nil
}

// dedupStateIDs returns a defensive copy of ids with declaration order
// preserved, rejecting a zero StateID or an exact duplicate.
func dedupStateIDs(ids []StateID) ([]StateID, error) {
	seen := make(map[string]bool, len(ids))
	cp := make([]StateID, 0, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			return nil, fmt.Errorf("%w: state id must not be zero", ErrInvalidStateID)
		}
		key := id.String()
		if seen[key] {
			return nil, fmt.Errorf("state %q: %w", key, ErrDuplicateStateID)
		}
		seen[key] = true
		cp = append(cp, id)
	}
	return cp, nil
}

// WithExtension returns a copy of t with its extension data set.
func (t TransitionDefinition) WithExtension(e core.Extension) TransitionDefinition {
	t.extension = e
	return t
}

func (t TransitionDefinition) ID() TransitionID { return t.id }

// SourceStates returns a defensive copy of t's permitted source States, in
// declaration order.
func (t TransitionDefinition) SourceStates() []StateID {
	if len(t.sourceStates) == 0 {
		return nil
	}
	cp := make([]StateID, len(t.sourceStates))
	copy(cp, t.sourceStates)
	return cp
}

// TargetStates returns a defensive copy of t's permitted target States, in
// declaration order.
func (t TransitionDefinition) TargetStates() []StateID {
	if len(t.targetStates) == 0 {
		return nil
	}
	cp := make([]StateID, len(t.targetStates))
	copy(cp, t.targetStates)
	return cp
}

func (t TransitionDefinition) Extension() core.Extension { return t.extension }

// IsZero reports whether t is the zero value.
func (t TransitionDefinition) IsZero() bool {
	return t.id.IsZero() && len(t.sourceStates) == 0 && len(t.targetStates) == 0
}

type transitionDefinitionJSON struct {
	ID           TransitionID    `json:"id"`
	SourceStates []StateID       `json:"source_states"`
	TargetStates []StateID       `json:"target_states"`
	Extension    *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes t as {"id":..., "source_states":[...],
// "target_states":[...], ...}, omitting extension when not set.
func (t TransitionDefinition) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal TransitionDefinition: %w", ErrInvalidTransition)
	}
	raw := transitionDefinitionJSON{ID: t.id, SourceStates: t.sourceStates, TargetStates: t.targetStates}
	if !t.extension.IsZero() {
		raw.Extension = &t.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes t from its JSON form, applying the same
// validation as NewTransitionDefinition.
func (t *TransitionDefinition) UnmarshalJSON(data []byte) error {
	var raw transitionDefinitionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal TransitionDefinition: %w", err)
	}
	result, err := NewTransitionDefinition(raw.ID, raw.SourceStates, raw.TargetStates)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*t = result
	return nil
}
