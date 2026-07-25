package lifecycle

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// mustVocab constructs a core.VocabularyValue under core.PEOSNamespace for
// this package's own predeclared open-vocabulary constants. It panics only
// on a malformed hardcoded literal supplied by this file itself, never on
// caller input -- the same guarantee core's own predeclared vocabulary
// constants (e.g. core.RelationTypeDependency) have simply by being struct
// literals inside the core package; this package cannot use that literal
// form itself, because core.VocabularyValue's fields are unexported
// outside core.
func mustVocab(value string) core.VocabularyValue {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, value)
	if err != nil {
		panic(fmt.Sprintf("lifecycle: invalid predeclared vocabulary value %q: %v", value, err))
	}
	return v
}

// StateClassification classifies a Lifecycle State's structural role
// within its Definition Version (PEOS-003: "A Lifecycle State SHOULD
// define... whether the State is initial, terminal, exceptional, or
// ordinary"). StateClassification is an open vocabulary: a Product MAY
// declare additional classification values beyond the three predeclared
// below.
//
// This package deliberately predeclares no "initial" classification
// value. Initial-state membership is owned by DefinitionVersion, not by an
// individual State (see NewDefinitionVersion in definition.go and PEOS-003
// "Initial State", which discusses initial-state policy as a property of
// the Lifecycle Definition).
//
// Terminal classification is declarative only. This package does not
// enforce a blanket prohibition on outgoing Transitions from a State
// classified terminal: PEOS-003 permits exceptional recovery from a
// Terminal State when a Lifecycle Definition explicitly allows it, so a
// structural rejection here would be normatively wrong. Enforcing that a
// terminal State has no *ordinary* outgoing Transition, if a Product wants
// that check, is Product policy applied on top of this package's data, not
// a rule this package bakes in.
type StateClassification struct{ value core.VocabularyValue }

// NewStateClassification wraps v as a StateClassification.
func NewStateClassification(v core.VocabularyValue) StateClassification {
	return StateClassification{value: v}
}

func (c StateClassification) Value() core.VocabularyValue { return c.value }
func (c StateClassification) IsZero() bool                { return c.value.IsZero() }
func (c StateClassification) String() string              { return c.value.String() }

func (c StateClassification) Equal(other StateClassification) bool {
	return c.value.Equal(other.value)
}

func (c StateClassification) MarshalJSON() ([]byte, error) { return json.Marshal(c.value) }

func (c *StateClassification) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.value)
}

var (
	// StateClassificationOrdinary marks a State that is neither terminal
	// nor exceptional.
	StateClassificationOrdinary = StateClassification{value: mustVocab("ordinary")}

	// StateClassificationTerminal marks a State from which no ordinary
	// outgoing Transition is expected (PEOS-003 "Terminal State").
	StateClassificationTerminal = StateClassification{value: mustVocab("terminal")}

	// StateClassificationExceptional marks a State representing a
	// condition outside normal successful lifecycle progression
	// (PEOS-003 "Exceptional State").
	StateClassificationExceptional = StateClassification{value: mustVocab("exceptional")}
)

// State is an identifiable normative condition that may be assigned to a
// Lifecycle Subject (PEOS-003 "Lifecycle State"), scoped to its owning
// Lifecycle Definition Version. State is a Definition-Version-owned value
// structure: it has no Artifact identity, no revisions, and no lifecycle
// of its own.
type State struct {
	id             StateID
	meaning        string
	classification StateClassification
	extension      core.Extension
}

// NewState validates id and meaning and returns a State with no
// classification set. meaning is PEOS-003's required "explicit semantic
// meaning" -- PEOS-003's own example ("a State named `accepted` is
// insufficient unless the Lifecycle Definition specifies what acceptance
// means") makes a bare name insufficient, so meaning must be non-empty.
// Use WithClassification to declare whether the State is ordinary,
// terminal, exceptional, or a Product-declared classification.
func NewState(id StateID, meaning string) (State, error) {
	if id.IsZero() {
		return State{}, fmt.Errorf("lifecycle: NewState: %w", ErrInvalidStateID)
	}
	if strings.TrimSpace(meaning) == "" {
		return State{}, fmt.Errorf("lifecycle: NewState: %w: meaning must not be empty", ErrInvalidState)
	}
	return State{id: id, meaning: meaning}, nil
}

// WithClassification returns a copy of s with its classification set.
func (s State) WithClassification(c StateClassification) State {
	s.classification = c
	return s
}

// WithExtension returns a copy of s with its extension data set.
func (s State) WithExtension(e core.Extension) State {
	s.extension = e
	return s
}

func (s State) ID() StateID     { return s.id }
func (s State) Meaning() string { return s.meaning }

// Classification returns s's declared classification, and whether one is
// set.
func (s State) Classification() (StateClassification, bool) {
	return s.classification, !s.classification.IsZero()
}

func (s State) Extension() core.Extension { return s.extension }

// IsZero reports whether s is the zero value.
func (s State) IsZero() bool { return s.id.IsZero() && s.meaning == "" }

type stateJSON struct {
	ID             StateID              `json:"id"`
	Meaning        string               `json:"meaning"`
	Classification *StateClassification `json:"classification,omitempty"`
	Extension      *core.Extension      `json:"extension,omitempty"`
}

// MarshalJSON encodes s as {"id":..., "meaning":..., ...}, omitting
// classification and extension when not set.
func (s State) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal State: %w", ErrInvalidState)
	}
	raw := stateJSON{ID: s.id, Meaning: s.meaning}
	if !s.classification.IsZero() {
		raw.Classification = &s.classification
	}
	if !s.extension.IsZero() {
		raw.Extension = &s.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes s from its JSON form, applying the same
// validation as NewState. An explicit JSON null for "classification" is
// treated the same as an absent key (both leave s.Classification()'s ok
// return false): unlike Scope in peos/relation, no PEOS-003 invariant
// distinguishes an explicitly-null classification from an absent one, so
// this package does not add a presence-tracking flag or reject null here.
func (s *State) UnmarshalJSON(data []byte) error {
	var raw stateJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal State: %w", err)
	}
	result, err := NewState(raw.ID, raw.Meaning)
	if err != nil {
		return err
	}
	if raw.Classification != nil {
		result = result.WithClassification(*raw.Classification)
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*s = result
	return nil
}
