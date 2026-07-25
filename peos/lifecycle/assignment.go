package lifecycle

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// StateAssignment is an identifiable association between a Lifecycle
// Subject, a Lifecycle Definition Version, and a Lifecycle State
// (PEOS-003 "State Assignment"). It is an immutable non-Artifact record:
// PEOS-003 never calls it "an Artifact" or requires it to conform to
// PEOS-002 -- contrast TransitionRecord (record.go), which PEOS-003
// explicitly does call a persistent Artifact.
type StateAssignment struct {
	id                core.StateAssignmentID
	subject           core.LifecycleSubjectRef
	definitionVersion core.LifecycleDefinitionVersionRef
	state             StateID
	effectiveAt       core.Timestamp
	provenance        core.Provenance
	establishedBy     core.ArtifactRevisionRef
	hasAuthority      bool
	authority         core.AuthorityRef
	extension         core.Extension
}

// NewStateAssignment validates its required arguments and returns a
// StateAssignment with no Authority set. Use WithAuthority to add one.
//
// This constructor performs only local structural validation: it does not
// verify that subject or definitionVersion actually exist, that state is
// a declared member of some remotely loaded DefinitionVersion, that
// establishedBy identifies a real, existing Transition Record Revision,
// or that any authority is sufficient. Those checks require data this
// package does not fetch. See WithDefinitionVersion for an opt-in local
// membership check against a DefinitionVersion value already in hand.
func NewStateAssignment(
	id core.StateAssignmentID,
	subject core.LifecycleSubjectRef,
	definitionVersion core.LifecycleDefinitionVersionRef,
	state StateID,
	effectiveAt core.Timestamp,
	provenance core.Provenance,
	establishedBy core.ArtifactRevisionRef,
) (StateAssignment, error) {
	if id.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w", ErrInvalidStateAssignment)
	}
	if subject.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w: subject must not be zero", ErrInvalidStateAssignment)
	}
	if definitionVersion.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w: definition version must not be zero", ErrInvalidStateAssignment)
	}
	if state.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w: state must not be zero", ErrInvalidStateAssignment)
	}
	if effectiveAt.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w: effective time must not be zero", ErrInvalidStateAssignment)
	}
	if provenance.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w: provenance must not be zero", ErrInvalidStateAssignment)
	}
	if establishedBy.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: NewStateAssignment: %w: established-by revision must not be zero", ErrInvalidStateAssignment)
	}
	return StateAssignment{
		id: id, subject: subject, definitionVersion: definitionVersion, state: state,
		effectiveAt: effectiveAt, provenance: provenance, establishedBy: establishedBy,
	}, nil
}

// WithDefinitionVersion validates that a's assigned State is a declared
// member of version's own State set, and that version identifies the same
// Definition Version a already references. It performs no repository
// access: it is purely local structural validation against a
// DefinitionVersion value the caller already has in hand.
func (a StateAssignment) WithDefinitionVersion(version DefinitionVersion) (StateAssignment, error) {
	ref, err := version.Ref()
	if err != nil {
		return StateAssignment{}, fmt.Errorf("lifecycle: StateAssignment.WithDefinitionVersion: %w", err)
	}
	if ref != a.definitionVersion {
		return StateAssignment{}, fmt.Errorf("lifecycle: StateAssignment.WithDefinitionVersion: %w: definition version mismatch", ErrInvalidStateAssignment)
	}
	found := false
	for _, s := range version.States() {
		if s.ID().Equal(a.state) {
			found = true
			break
		}
	}
	if !found {
		return StateAssignment{}, fmt.Errorf("lifecycle: StateAssignment.WithDefinitionVersion: %w", ErrUnknownStateID)
	}
	return a, nil
}

// WithAuthority returns a copy of a with its authority set. authority
// must be non-zero; use WithoutAuthority to clear a previously set
// authority.
func (a StateAssignment) WithAuthority(authority core.AuthorityRef) (StateAssignment, error) {
	if authority.IsZero() {
		return StateAssignment{}, fmt.Errorf("lifecycle: StateAssignment.WithAuthority: %w: authority must not be zero", ErrInvalidStateAssignment)
	}
	a.authority, a.hasAuthority = authority, true
	return a, nil
}

// WithoutAuthority returns a copy of a with its authority cleared.
func (a StateAssignment) WithoutAuthority() StateAssignment {
	a.authority, a.hasAuthority = core.AuthorityRef{}, false
	return a
}

// WithExtension returns a copy of a with its extension data set.
func (a StateAssignment) WithExtension(e core.Extension) StateAssignment {
	a.extension = e
	return a
}

func (a StateAssignment) ID() core.StateAssignmentID        { return a.id }
func (a StateAssignment) Subject() core.LifecycleSubjectRef { return a.subject }
func (a StateAssignment) DefinitionVersion() core.LifecycleDefinitionVersionRef {
	return a.definitionVersion
}
func (a StateAssignment) State() StateID                          { return a.state }
func (a StateAssignment) EffectiveAt() core.Timestamp             { return a.effectiveAt }
func (a StateAssignment) Provenance() core.Provenance             { return a.provenance }
func (a StateAssignment) EstablishedBy() core.ArtifactRevisionRef { return a.establishedBy }

// Authority returns a's declared authority, and whether one is set.
func (a StateAssignment) Authority() (core.AuthorityRef, bool) { return a.authority, a.hasAuthority }

func (a StateAssignment) Extension() core.Extension { return a.extension }

// IsZero reports whether a is the zero value.
func (a StateAssignment) IsZero() bool { return a.id.IsZero() }

// Ref returns a core.StateAssignmentRef identifying a.
func (a StateAssignment) Ref() (core.StateAssignmentRef, error) {
	return core.NewStateAssignmentRef(a.id)
}

type stateAssignmentJSON struct {
	ID                core.StateAssignmentID             `json:"id"`
	Subject           core.LifecycleSubjectRef           `json:"subject"`
	DefinitionVersion core.LifecycleDefinitionVersionRef `json:"definition_version"`
	State             StateID                            `json:"state"`
	EffectiveAt       core.Timestamp                     `json:"effective_at"`
	Provenance        core.Provenance                    `json:"provenance"`
	EstablishedBy     core.ArtifactRevisionRef           `json:"established_by"`
	Authority         *core.AuthorityRef                 `json:"authority,omitempty"`
	Extension         *core.Extension                    `json:"extension,omitempty"`
}

// stateAssignmentUnmarshalJSON mirrors stateAssignmentJSON's field set for
// decoding only, with one difference: Authority is captured as raw,
// undecoded bytes rather than *core.AuthorityRef, so an explicit JSON null
// can be distinguished from an absent key and rejected -- the same
// technique Packet D.1 established for Relation.Scope in peos/relation.
type stateAssignmentUnmarshalJSON struct {
	ID                core.StateAssignmentID             `json:"id"`
	Subject           core.LifecycleSubjectRef           `json:"subject"`
	DefinitionVersion core.LifecycleDefinitionVersionRef `json:"definition_version"`
	State             StateID                            `json:"state"`
	EffectiveAt       core.Timestamp                     `json:"effective_at"`
	Provenance        core.Provenance                    `json:"provenance"`
	EstablishedBy     core.ArtifactRevisionRef           `json:"established_by"`
	Authority         json.RawMessage                    `json:"authority"`
	Extension         *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes a as {"id":..., "subject":..., ...}, omitting
// authority and extension when not set. No "scope" field is ever written:
// per E-13, StateAssignment carries no Scope of its own.
func (a StateAssignment) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal StateAssignment: %w", ErrInvalidStateAssignment)
	}
	raw := stateAssignmentJSON{
		ID: a.id, Subject: a.subject, DefinitionVersion: a.definitionVersion, State: a.state,
		EffectiveAt: a.effectiveAt, Provenance: a.provenance, EstablishedBy: a.establishedBy,
	}
	if a.hasAuthority {
		raw.Authority = &a.authority
	}
	if !a.extension.IsZero() {
		raw.Extension = &a.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes a from its JSON form, applying the same
// validation as NewStateAssignment and WithAuthority. An explicit JSON
// null for "authority" is rejected rather than silently treated as
// "no authority."
func (a *StateAssignment) UnmarshalJSON(data []byte) error {
	var raw stateAssignmentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal StateAssignment: %w", err)
	}
	result, err := NewStateAssignment(raw.ID, raw.Subject, raw.DefinitionVersion, raw.State, raw.EffectiveAt, raw.Provenance, raw.EstablishedBy)
	if err != nil {
		return err
	}
	if len(raw.Authority) > 0 {
		if string(raw.Authority) == "null" {
			return fmt.Errorf("lifecycle: unmarshal StateAssignment: %w: authority must not be null", ErrInvalidStateAssignment)
		}
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("lifecycle: unmarshal StateAssignment: %w", err)
		}
		result, err = result.WithAuthority(authority)
		if err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*a = result
	return nil
}
