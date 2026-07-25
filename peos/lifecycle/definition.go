package lifecycle

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Definition is a Lifecycle Definition (PEOS-003): an identifiable
// normative contract defining the permitted states and transitions of one
// or more Lifecycle Subject Types. Per this package's own locked shape
// (see doc.go), Definition carries its own generic normative identity
// (core.LifecycleDefinitionID) rather than a core.Artifact identity.
// Definition itself carries no other content: every substantive part of
// the published lifecycle contract (States, Transitions, Scope, governed
// Subject Types, initial-state policy, entry Transition) belongs to
// DefinitionVersion, below.
type Definition struct {
	id core.LifecycleDefinitionID
}

// NewDefinition validates id and returns a Definition.
func NewDefinition(id core.LifecycleDefinitionID) (Definition, error) {
	if id.IsZero() {
		return Definition{}, fmt.Errorf("lifecycle: NewDefinition: %w", ErrInvalidLifecycleDefinition)
	}
	return Definition{id: id}, nil
}

func (d Definition) ID() core.LifecycleDefinitionID { return d.id }

// IsZero reports whether d is the zero value.
func (d Definition) IsZero() bool { return d.id.IsZero() }

// Ref returns a core.LifecycleDefinitionRef identifying d.
func (d Definition) Ref() (core.LifecycleDefinitionRef, error) {
	return core.NewLifecycleDefinitionRef(d.id)
}

type definitionJSON struct {
	ID core.LifecycleDefinitionID `json:"id"`
}

// MarshalJSON encodes d as {"id": ...}.
func (d Definition) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal Definition: %w", ErrInvalidLifecycleDefinition)
	}
	return json.Marshal(definitionJSON{ID: d.id})
}

// UnmarshalJSON decodes d from its JSON form, applying the same
// validation as NewDefinition.
func (d *Definition) UnmarshalJSON(data []byte) error {
	var raw definitionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal Definition: %w", err)
	}
	result, err := NewDefinition(raw.ID)
	if err != nil {
		return err
	}
	*d = result
	return nil
}

// DefinitionVersion is a Lifecycle Definition Version (PEOS-003): an
// immutable and identifiable state of a Lifecycle Definition, containing
// the complete published lifecycle contract -- governed Subject Types,
// States, Transitions, initial-state policy, and entry Transition.
// DefinitionVersion carries its own generic normative identity
// (core.LifecycleDefinitionVersionID) and does not compose
// core.ArtifactRevision (see doc.go).
//
// core.LifecycleDefinitionVersionID itself satisfies PEOS-003's
// requirement that a Lifecycle Definition Version have "an ordering or
// version identifier": it is the version identifier, exactly as
// core.ArtifactRevisionID already serves that role for Artifact Revisions
// without this package assuming any inherent sort order across values.
// No additional ordinal field is added.
type DefinitionVersion struct {
	id              core.LifecycleDefinitionVersionID
	definition      core.LifecycleDefinitionRef
	scope           core.Scope
	subjectTypes    []core.VocabularyValue
	states          []State
	initialStates   []StateID
	transitions     []TransitionDefinition
	entryTransition TransitionID
	provenance      core.Provenance
	extension       core.Extension
}

// NewDefinitionVersion validates every argument -- including cross-field
// consistency between states, initialStates, transitions, and
// entryTransition -- and returns a DefinitionVersion. All slice arguments
// are defensively copied; declaration order is preserved.
//
// This constructor performs only local structural validation. It does not
// perform repository lookups: it cannot and does not verify that
// definition identifies a Definition that actually exists anywhere beyond
// the reference itself.
func NewDefinitionVersion(
	id core.LifecycleDefinitionVersionID,
	definition core.LifecycleDefinitionRef,
	scope core.Scope,
	subjectTypes []core.VocabularyValue,
	states []State,
	initialStates []StateID,
	transitions []TransitionDefinition,
	entryTransition TransitionID,
	provenance core.Provenance,
) (DefinitionVersion, error) {
	if id.IsZero() {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w", ErrInvalidLifecycleDefinitionVersion)
	}
	if definition.IsZero() {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: parent definition must not be zero", ErrInvalidLifecycleDefinitionVersion)
	}
	if scope.IsZero() {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: scope must not be zero", ErrInvalidLifecycleDefinitionVersion)
	}
	if provenance.IsZero() {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: provenance must not be zero", ErrInvalidLifecycleDefinitionVersion)
	}

	if len(subjectTypes) == 0 {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: subject types must not be empty", ErrInvalidLifecycleDefinitionVersion)
	}
	subjectTypesCopy := make([]core.VocabularyValue, len(subjectTypes))
	for i, st := range subjectTypes {
		if st.IsZero() {
			return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: subject type must not be zero", ErrInvalidLifecycleDefinitionVersion)
		}
		subjectTypesCopy[i] = st
	}

	if len(states) == 0 {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: state set must not be empty", ErrInvalidLifecycleDefinitionVersion)
	}
	stateSet := make(map[string]bool, len(states))
	statesCopy := make([]State, len(states))
	for i, s := range states {
		if s.IsZero() {
			return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: state must not be zero", ErrInvalidState)
		}
		key := s.ID().String()
		if stateSet[key] {
			return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: state %q: %w", key, ErrDuplicateStateID)
		}
		stateSet[key] = true
		statesCopy[i] = s
	}

	if len(initialStates) == 0 {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: initial state set must not be empty", ErrInvalidInitialState)
	}
	initialCopy := make([]StateID, len(initialStates))
	for i, sid := range initialStates {
		if sid.IsZero() || !stateSet[sid.String()] {
			return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: initial state %q: %w", sid.String(), ErrInvalidInitialState)
		}
		initialCopy[i] = sid
	}

	if len(transitions) == 0 {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: transition set must not be empty", ErrInvalidLifecycleDefinitionVersion)
	}
	transitionSet := make(map[string]bool, len(transitions))
	transitionsCopy := make([]TransitionDefinition, len(transitions))
	for i, tr := range transitions {
		if tr.IsZero() {
			return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w: transition must not be zero", ErrInvalidTransition)
		}
		key := tr.ID().String()
		if transitionSet[key] {
			return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: transition %q: %w", key, ErrDuplicateTransitionID)
		}
		transitionSet[key] = true
		for _, sid := range tr.SourceStates() {
			if !stateSet[sid.String()] {
				return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: transition %q source state %q: %w", key, sid.String(), ErrUnknownStateID)
			}
		}
		for _, sid := range tr.TargetStates() {
			if !stateSet[sid.String()] {
				return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: transition %q target state %q: %w", key, sid.String(), ErrUnknownStateID)
			}
		}
		transitionsCopy[i] = tr
	}

	if entryTransition.IsZero() || !transitionSet[entryTransition.String()] {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: NewDefinitionVersion: %w", ErrInvalidEntryTransition)
	}

	return DefinitionVersion{
		id:              id,
		definition:      definition,
		scope:           scope,
		subjectTypes:    subjectTypesCopy,
		states:          statesCopy,
		initialStates:   initialCopy,
		transitions:     transitionsCopy,
		entryTransition: entryTransition,
		provenance:      provenance,
	}, nil
}

// WithExtension returns a copy of v with its extension data set.
func (v DefinitionVersion) WithExtension(e core.Extension) DefinitionVersion {
	v.extension = e
	return v
}

func (v DefinitionVersion) ID() core.LifecycleDefinitionVersionID   { return v.id }
func (v DefinitionVersion) Definition() core.LifecycleDefinitionRef { return v.definition }
func (v DefinitionVersion) Scope() core.Scope                       { return v.scope }

// SubjectTypes returns a defensive copy of v's governed Subject Types, in
// declaration order.
func (v DefinitionVersion) SubjectTypes() []core.VocabularyValue {
	if len(v.subjectTypes) == 0 {
		return nil
	}
	cp := make([]core.VocabularyValue, len(v.subjectTypes))
	copy(cp, v.subjectTypes)
	return cp
}

// States returns a defensive copy of v's declared States, in declaration
// order.
func (v DefinitionVersion) States() []State {
	if len(v.states) == 0 {
		return nil
	}
	cp := make([]State, len(v.states))
	copy(cp, v.states)
	return cp
}

// InitialStates returns a defensive copy of v's permitted initial States,
// in declaration order.
func (v DefinitionVersion) InitialStates() []StateID {
	if len(v.initialStates) == 0 {
		return nil
	}
	cp := make([]StateID, len(v.initialStates))
	copy(cp, v.initialStates)
	return cp
}

// Transitions returns a defensive copy of v's declared Transition
// Definitions, in declaration order.
func (v DefinitionVersion) Transitions() []TransitionDefinition {
	if len(v.transitions) == 0 {
		return nil
	}
	cp := make([]TransitionDefinition, len(v.transitions))
	copy(cp, v.transitions)
	return cp
}

func (v DefinitionVersion) EntryTransition() TransitionID { return v.entryTransition }
func (v DefinitionVersion) Provenance() core.Provenance   { return v.provenance }
func (v DefinitionVersion) Extension() core.Extension     { return v.extension }

// IsZero reports whether v is the zero value.
func (v DefinitionVersion) IsZero() bool { return v.id.IsZero() }

// Ref returns a core.LifecycleDefinitionVersionRef identifying v, scoped
// to its parent Definition.
func (v DefinitionVersion) Ref() (core.LifecycleDefinitionVersionRef, error) {
	return core.NewLifecycleDefinitionVersionRef(v.definition.LifecycleDefinitionID(), v.id)
}

type definitionVersionJSON struct {
	ID              core.LifecycleDefinitionVersionID `json:"id"`
	Definition      core.LifecycleDefinitionRef       `json:"definition"`
	Scope           core.Scope                        `json:"scope"`
	SubjectTypes    []core.VocabularyValue            `json:"subject_types"`
	States          []State                           `json:"states"`
	InitialStates   []StateID                         `json:"initial_states"`
	Transitions     []TransitionDefinition            `json:"transitions"`
	EntryTransition TransitionID                      `json:"entry_transition"`
	Provenance      core.Provenance                   `json:"provenance"`
	Extension       *core.Extension                   `json:"extension,omitempty"`
}

// MarshalJSON encodes v as a flat JSON object containing every field
// listed in definitionVersionJSON, omitting extension when not set.
func (v DefinitionVersion) MarshalJSON() ([]byte, error) {
	if v.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal DefinitionVersion: %w", ErrInvalidLifecycleDefinitionVersion)
	}
	raw := definitionVersionJSON{
		ID: v.id, Definition: v.definition, Scope: v.scope, SubjectTypes: v.subjectTypes,
		States: v.states, InitialStates: v.initialStates, Transitions: v.transitions,
		EntryTransition: v.entryTransition, Provenance: v.provenance,
	}
	if !v.extension.IsZero() {
		raw.Extension = &v.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes v from its JSON form, applying the same
// validation as NewDefinitionVersion. Unknown ordinary JSON fields are
// ignored.
func (v *DefinitionVersion) UnmarshalJSON(data []byte) error {
	var raw definitionVersionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal DefinitionVersion: %w", err)
	}
	result, err := NewDefinitionVersion(
		raw.ID, raw.Definition, raw.Scope, raw.SubjectTypes, raw.States,
		raw.InitialStates, raw.Transitions, raw.EntryTransition, raw.Provenance,
	)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*v = result
	return nil
}
