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
// (core.LifecycleDefinitionID) as its canonical identity, rather than a
// core.Artifact identity. Definition itself carries no other required
// content: every substantive part of the published lifecycle contract
// (States, Transitions, Scope, governed Subject Types, initial-state
// policy, entry Transition) belongs to DefinitionVersion, below.
//
// Definition MAY optionally carry a core.ArtifactRef (PEOS-003: "A
// Lifecycle Definition MAY be represented as an Artifact"). That binding
// is a correspondence to an Artifact identity that already exists
// elsewhere -- it does not replace or supersede core.LifecycleDefinitionID
// as this type's identity, and Definition remains fully usable, valid,
// and IsZero()-false with no binding present at all.
type Definition struct {
	id          core.LifecycleDefinitionID
	hasArtifact bool
	artifact    core.ArtifactRef
}

// NewDefinition validates id and returns a Definition with no Artifact
// binding. Use WithArtifact to add one.
func NewDefinition(id core.LifecycleDefinitionID) (Definition, error) {
	if id.IsZero() {
		return Definition{}, fmt.Errorf("lifecycle: NewDefinition: %w", ErrInvalidLifecycleDefinition)
	}
	return Definition{id: id}, nil
}

func (d Definition) ID() core.LifecycleDefinitionID { return d.id }

// WithArtifact returns a copy of d with its Artifact binding set. artifact
// must be non-zero. d's own core.LifecycleDefinitionID is unchanged and
// remains d's identity; artifact only records that d also corresponds to
// the given Artifact. Use WithoutArtifact to clear a previously set
// binding.
func (d Definition) WithArtifact(artifact core.ArtifactRef) (Definition, error) {
	if artifact.IsZero() {
		return Definition{}, fmt.Errorf("lifecycle: Definition.WithArtifact: %w: artifact must not be zero", ErrInvalidLifecycleDefinition)
	}
	d.artifact, d.hasArtifact = artifact, true
	return d, nil
}

// WithoutArtifact returns a copy of d with its Artifact binding cleared.
func (d Definition) WithoutArtifact() Definition {
	d.artifact, d.hasArtifact = core.ArtifactRef{}, false
	return d
}

// Artifact returns d's bound core.ArtifactRef, and whether one is set.
func (d Definition) Artifact() (core.ArtifactRef, bool) { return d.artifact, d.hasArtifact }

// IsZero reports whether d is the zero value.
func (d Definition) IsZero() bool { return d.id.IsZero() }

// Ref returns a core.LifecycleDefinitionRef identifying d.
func (d Definition) Ref() (core.LifecycleDefinitionRef, error) {
	return core.NewLifecycleDefinitionRef(d.id)
}

type definitionJSON struct {
	ID       core.LifecycleDefinitionID `json:"id"`
	Artifact *core.ArtifactRef          `json:"artifact,omitempty"`
}

// definitionUnmarshalJSON mirrors definitionJSON for decoding only, with
// one difference: Artifact is captured as raw, undecoded bytes rather than
// *core.ArtifactRef, so an explicit JSON null can be distinguished from an
// absent key and rejected -- the same technique Packet D.1 established
// for Relation.Scope in peos/relation, and already used elsewhere in this
// package (see assignment.go, record.go).
type definitionUnmarshalJSON struct {
	ID       core.LifecycleDefinitionID `json:"id"`
	Artifact json.RawMessage            `json:"artifact"`
}

// MarshalJSON encodes d as {"id": ..., "artifact": ...}, omitting artifact
// when not set.
func (d Definition) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal Definition: %w", ErrInvalidLifecycleDefinition)
	}
	raw := definitionJSON{ID: d.id}
	if d.hasArtifact {
		raw.Artifact = &d.artifact
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes d from its JSON form, applying the same
// validation as NewDefinition and WithArtifact. An explicit JSON null for
// "artifact" is rejected rather than silently treated as "no binding."
func (d *Definition) UnmarshalJSON(data []byte) error {
	var raw definitionUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal Definition: %w", err)
	}
	result, err := NewDefinition(raw.ID)
	if err != nil {
		return err
	}
	if len(raw.Artifact) > 0 {
		if string(raw.Artifact) == "null" {
			return fmt.Errorf("lifecycle: unmarshal Definition: %w: artifact must not be null", ErrInvalidLifecycleDefinition)
		}
		var artifact core.ArtifactRef
		if err := json.Unmarshal(raw.Artifact, &artifact); err != nil {
			return fmt.Errorf("lifecycle: unmarshal Definition: %w", err)
		}
		result, err = result.WithArtifact(artifact)
		if err != nil {
			return err
		}
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
//
// DefinitionVersion MAY optionally carry a core.ArtifactRevisionRef
// (PEOS-003: "When a Lifecycle Definition is represented as an Artifact,
// its Definition Version MUST identify the corresponding Artifact
// Revision"). That reference is a correspondence, not a replacement:
// core.LifecycleDefinitionVersionID remains v's canonical identity
// whether or not a binding is present. See ValidateArtifactBinding for
// the (optional, caller-invoked) local consistency check between a
// Definition's and a DefinitionVersion's bindings.
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

	hasArtifactRevision bool
	artifactRevision    core.ArtifactRevisionRef
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

// WithArtifactRevision returns a copy of v with its Artifact Revision
// binding set. ref must be non-zero. v's own core.LifecycleDefinitionVersionID
// and every other field are unchanged; ref only records that v also
// corresponds to the given Artifact Revision. Use WithoutArtifactRevision
// to clear a previously set binding.
func (v DefinitionVersion) WithArtifactRevision(ref core.ArtifactRevisionRef) (DefinitionVersion, error) {
	if ref.IsZero() {
		return DefinitionVersion{}, fmt.Errorf("lifecycle: DefinitionVersion.WithArtifactRevision: %w: artifact revision must not be zero", ErrInvalidLifecycleDefinitionVersion)
	}
	v.artifactRevision, v.hasArtifactRevision = ref, true
	return v, nil
}

// WithoutArtifactRevision returns a copy of v with its Artifact Revision
// binding cleared.
func (v DefinitionVersion) WithoutArtifactRevision() DefinitionVersion {
	v.artifactRevision, v.hasArtifactRevision = core.ArtifactRevisionRef{}, false
	return v
}

// ArtifactRevision returns v's bound core.ArtifactRevisionRef, and
// whether one is set.
func (v DefinitionVersion) ArtifactRevision() (core.ArtifactRevisionRef, bool) {
	return v.artifactRevision, v.hasArtifactRevision
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
	ID               core.LifecycleDefinitionVersionID `json:"id"`
	Definition       core.LifecycleDefinitionRef       `json:"definition"`
	Scope            core.Scope                        `json:"scope"`
	SubjectTypes     []core.VocabularyValue            `json:"subject_types"`
	States           []State                           `json:"states"`
	InitialStates    []StateID                         `json:"initial_states"`
	Transitions      []TransitionDefinition            `json:"transitions"`
	EntryTransition  TransitionID                      `json:"entry_transition"`
	Provenance       core.Provenance                   `json:"provenance"`
	ArtifactRevision *core.ArtifactRevisionRef         `json:"artifact_revision,omitempty"`
	Extension        *core.Extension                   `json:"extension,omitempty"`
}

// definitionVersionUnmarshalJSON mirrors definitionVersionJSON for
// decoding only, with one difference: ArtifactRevision is captured as
// raw, undecoded bytes rather than *core.ArtifactRevisionRef, so an
// explicit JSON null can be distinguished from an absent key and rejected
// -- the same technique used by Definition's own UnmarshalJSON above.
type definitionVersionUnmarshalJSON struct {
	ID               core.LifecycleDefinitionVersionID `json:"id"`
	Definition       core.LifecycleDefinitionRef       `json:"definition"`
	Scope            core.Scope                        `json:"scope"`
	SubjectTypes     []core.VocabularyValue            `json:"subject_types"`
	States           []State                           `json:"states"`
	InitialStates    []StateID                         `json:"initial_states"`
	Transitions      []TransitionDefinition            `json:"transitions"`
	EntryTransition  TransitionID                      `json:"entry_transition"`
	Provenance       core.Provenance                   `json:"provenance"`
	ArtifactRevision json.RawMessage                   `json:"artifact_revision"`
	Extension        *core.Extension                   `json:"extension,omitempty"`
}

// MarshalJSON encodes v as a flat JSON object containing every field
// listed in definitionVersionJSON, omitting artifact_revision and
// extension when not set. A DefinitionVersion with no Artifact Revision
// binding marshals identically to the pre-Packet-E.1 wire form: no
// artifact_revision key is ever written unless a binding is present.
func (v DefinitionVersion) MarshalJSON() ([]byte, error) {
	if v.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal DefinitionVersion: %w", ErrInvalidLifecycleDefinitionVersion)
	}
	raw := definitionVersionJSON{
		ID: v.id, Definition: v.definition, Scope: v.scope, SubjectTypes: v.subjectTypes,
		States: v.states, InitialStates: v.initialStates, Transitions: v.transitions,
		EntryTransition: v.entryTransition, Provenance: v.provenance,
	}
	if v.hasArtifactRevision {
		raw.ArtifactRevision = &v.artifactRevision
	}
	if !v.extension.IsZero() {
		raw.Extension = &v.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes v from its JSON form, applying the same
// validation as NewDefinitionVersion and WithArtifactRevision. Unknown
// ordinary JSON fields are ignored. An explicit JSON null for
// "artifact_revision" is rejected rather than silently treated as "no
// binding." JSON produced before this binding existed (no
// artifact_revision key at all) decodes exactly as before.
func (v *DefinitionVersion) UnmarshalJSON(data []byte) error {
	var raw definitionVersionUnmarshalJSON
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
	if len(raw.ArtifactRevision) > 0 {
		if string(raw.ArtifactRevision) == "null" {
			return fmt.Errorf("lifecycle: unmarshal DefinitionVersion: %w: artifact_revision must not be null", ErrInvalidLifecycleDefinitionVersion)
		}
		var ref core.ArtifactRevisionRef
		if err := json.Unmarshal(raw.ArtifactRevision, &ref); err != nil {
			return fmt.Errorf("lifecycle: unmarshal DefinitionVersion: %w", err)
		}
		result, err = result.WithArtifactRevision(ref)
		if err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*v = result
	return nil
}

// ValidateArtifactBinding performs a pure, local consistency check
// between v's own optional Artifact Revision binding and def's optional
// Artifact binding. It validates only the supplied values: it does not
// look up, fetch, or otherwise verify that either binding corresponds to
// a real, existing Artifact -- that remains a repository or Product
// concern.
//
// It first requires that v's parent Lifecycle Definition reference
// identifies def (v.Definition().LifecycleDefinitionID() ==
// def.ID()) -- this holds regardless of whether either side has an
// Artifact binding. It then requires that v and def agree on whether an
// Artifact-backed representation is in use at all: neither bound is
// valid, and both bound is valid only when the bound
// ArtifactRevisionRef's ArtifactID matches the bound ArtifactRef's
// ArtifactID (PEOS-003: "its Definition Version MUST identify the
// corresponding Artifact Revision"). Exactly one side bound is always
// invalid.
func (v DefinitionVersion) ValidateArtifactBinding(def Definition) error {
	if v.definition.LifecycleDefinitionID() != def.ID() {
		return fmt.Errorf("lifecycle: DefinitionVersion.ValidateArtifactBinding: %w: definition version's parent does not match the given Definition", ErrArtifactBindingMismatch)
	}

	versionArtifactRevision, versionHasBinding := v.ArtifactRevision()
	defArtifact, defHasBinding := def.Artifact()

	switch {
	case !versionHasBinding && !defHasBinding:
		return nil
	case versionHasBinding && !defHasBinding:
		return fmt.Errorf("lifecycle: DefinitionVersion.ValidateArtifactBinding: %w: version is Artifact-backed but definition is not", ErrArtifactBindingMismatch)
	case !versionHasBinding && defHasBinding:
		return fmt.Errorf("lifecycle: DefinitionVersion.ValidateArtifactBinding: %w: definition is Artifact-backed but version is not", ErrArtifactBindingMismatch)
	default:
		if versionArtifactRevision.ArtifactID() != defArtifact.ArtifactID() {
			return fmt.Errorf("lifecycle: DefinitionVersion.ValidateArtifactBinding: %w: version's artifact revision does not belong to definition's artifact", ErrArtifactBindingMismatch)
		}
		return nil
	}
}
