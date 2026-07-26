package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// ConflictResolutionID is the package-local identity of a
// ConflictResolution. See SupersessionID's own doc comment
// (identity.go) for why this identity is package-local and structurally
// distinct rather than a shared wrapper.
type ConflictResolutionID struct{ conflictResolutionID string }

// NewConflictResolutionID validates value and returns a
// ConflictResolutionID. Surrounding whitespace is trimmed; the result is
// rejected if empty.
func NewConflictResolutionID(value string) (ConflictResolutionID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ConflictResolutionID{}, fmt.Errorf("decision: NewConflictResolutionID: %w", ErrInvalidConflictResolution)
	}
	return ConflictResolutionID{conflictResolutionID: trimmed}, nil
}

// String returns the opaque identity value.
func (id ConflictResolutionID) String() string { return id.conflictResolutionID }

// IsZero reports whether id is the zero value.
func (id ConflictResolutionID) IsZero() bool { return id.conflictResolutionID == "" }

// Equal reports whether id and other carry the same identity value.
func (id ConflictResolutionID) Equal(other ConflictResolutionID) bool {
	return id.conflictResolutionID == other.conflictResolutionID
}

// MarshalJSON encodes id as a JSON string.
func (id ConflictResolutionID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("decision: marshal ConflictResolutionID: %w", ErrInvalidConflictResolution)
	}
	return json.Marshal(id.conflictResolutionID)
}

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewConflictResolutionID.
func (id *ConflictResolutionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("decision: unmarshal ConflictResolutionID: %w", err)
	}
	v, err := NewConflictResolutionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// ResolutionMechanism names how a Decision Conflict was resolved
// (PEOS-004 "Decision Conflict": "Conflict resolution MUST be
// established by: explicit priority; authority; supersession; scope
// refinement; another Decision; an applicable Product contract.").
// ResolutionMechanism is an open vocabulary: a Product MAY declare
// additional mechanism values beyond the six predeclared below, because
// the Extension Model (PEOS-004 :1398-1411) explicitly names
// "conflict-resolution policies" among what a Product contract MAY
// define, and :1128's own sixth arm is itself "an applicable Product
// contract" -- the same open-vocabulary shape as OutcomeKind and
// CommitmentEffect (outcome.go), and the opposite of the closed,
// two-variant SupersessionExtent (supersession.go), whose completeness
// axis the Extension Model does not list.
type ResolutionMechanism struct{ value core.VocabularyValue }

// NewResolutionMechanism wraps v as a ResolutionMechanism.
func NewResolutionMechanism(v core.VocabularyValue) ResolutionMechanism {
	return ResolutionMechanism{value: v}
}

func (m ResolutionMechanism) Value() core.VocabularyValue { return m.value }
func (m ResolutionMechanism) IsZero() bool                { return m.value.IsZero() }
func (m ResolutionMechanism) String() string              { return m.value.String() }

// Equal reports whether m and other name the same mechanism value.
func (m ResolutionMechanism) Equal(other ResolutionMechanism) bool {
	return m.value.Equal(other.value)
}

// MarshalJSON encodes m as its canonical "namespace:value" string form. A
// zero ResolutionMechanism is rejected: ConflictResolution requires a
// non-zero ResolutionMechanism unconditionally, so a zero value should
// never reach marshaling.
func (m ResolutionMechanism) MarshalJSON() ([]byte, error) {
	if m.IsZero() {
		return nil, fmt.Errorf("decision: marshal ResolutionMechanism: %w", ErrInvalidConflictResolution)
	}
	return json.Marshal(m.value)
}

func (m *ResolutionMechanism) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.value)
}

var (
	ResolutionMechanismPriority        = ResolutionMechanism{value: mustVocab("priority")}
	ResolutionMechanismAuthority       = ResolutionMechanism{value: mustVocab("authority")}
	ResolutionMechanismSupersession    = ResolutionMechanism{value: mustVocab("supersession")}
	ResolutionMechanismScopeRefinement = ResolutionMechanism{value: mustVocab("scope-refinement")}
	ResolutionMechanismDecision        = ResolutionMechanism{value: mustVocab("decision")}
	ResolutionMechanismProductContract = ResolutionMechanism{value: mustVocab("product-contract")}
)

// ConflictResolution is an immutable, independently identified governance
// record establishing that a specific DecisionConflict has been resolved
// (PEOS-004 "Conflict Invariant" :1342: "Conflicting applicable Decisions
// require explicit resolution."; "Decision Conflict": "Conflict
// resolution MUST be established by: explicit priority; authority;
// supersession; scope refinement; another Decision; an applicable
// Product contract.").
//
// ConflictResolution is a separate value, not a mutation of
// DecisionConflict: PEOS-004 never describes a conflict as being
// "resolved in place", and this package's own convention never mutates
// an already-issued governance record (see DecisionSupersession,
// DecisionInvalidation). There is no "resolved bool" or "status" field
// anywhere in this package's Conflict model; a DecisionConflict without a
// corresponding ConflictResolution is, by construction, unresolved.
//
// conflict identifies the specific DecisionConflict being closed.
// Referencing the two conflicting Decisions directly instead of the
// conflict's own identity was considered and rejected: the same pair of
// Decisions MAY be in more than one conflict over the same overlapping
// scope (see ConflictID's own doc comment), so only ConflictID
// unambiguously names which conflict this resolution closes.
//
// mechanism alone does not identify what actually happened: PEOS-004's
// six mechanisms are categories ("resolved by explicit priority"), not
// identifications of which priority, which authority, or which scope
// refinement applied. statement is required for that reason, and is what
// makes :1342's "explicit resolution" inspectable.
//
// resolvingDecision is optional, not conditionally required on
// mechanism. An earlier draft of this design required it exactly when
// mechanism was ResolutionMechanismDecision or
// ResolutionMechanismSupersession; that rule was rejected because
// ResolutionMechanism is an open vocabulary (see its own doc comment) --
// a Product-defined mechanism value that is, in substance, decision-based
// would silently escape a rule keyed to two hardcoded predeclared
// constants, producing an invariant that presents as complete but is
// not. The identification burden instead rests entirely on statement.
//
// ConflictResolution carries no effectiveAt or effectiveCondition field:
// unlike Supersession (PEOS-004 :1021, an explicit MUST for effective
// time/condition) and Invalidation (:1043, likewise), :1128 imposes no
// such requirement on conflict resolution. A Product needing an
// effective-time semantics for its resolutions carries it in extension.
type ConflictResolution struct {
	id                ConflictResolutionID
	conflict          ConflictID
	mechanism         ResolutionMechanism
	statement         string
	resolvingDecision core.DecisionRef
	provenance        core.Provenance
	extension         core.Extension
}

// NewConflictResolution validates its arguments and returns a
// ConflictResolution with no resolving Decision, provenance, or
// extension data. Use WithResolvingDecision, WithProvenance, and
// WithExtension to add those.
//
// id, conflict, and mechanism must each be non-zero. statement must be
// non-empty after trimming surrounding whitespace; it is stored as
// given.
//
// A successful call always returns a fully valid record: every mandatory
// field is a required constructor argument, never a later With* call.
func NewConflictResolution(
	id ConflictResolutionID,
	conflict ConflictID,
	mechanism ResolutionMechanism,
	statement string,
) (ConflictResolution, error) {
	if id.IsZero() {
		return ConflictResolution{}, fmt.Errorf("decision: NewConflictResolution: %w", ErrInvalidConflictResolution)
	}
	if conflict.IsZero() {
		return ConflictResolution{}, fmt.Errorf("decision: NewConflictResolution: %w: conflict must not be zero", ErrInvalidConflictResolution)
	}
	if mechanism.IsZero() {
		return ConflictResolution{}, fmt.Errorf("decision: NewConflictResolution: %w: mechanism must not be zero", ErrInvalidConflictResolution)
	}
	if strings.TrimSpace(statement) == "" {
		return ConflictResolution{}, fmt.Errorf("decision: NewConflictResolution: %w: statement must not be empty", ErrInvalidConflictResolution)
	}
	return ConflictResolution{
		id:        id,
		conflict:  conflict,
		mechanism: mechanism,
		statement: statement,
	}, nil
}

// WithResolvingDecision returns a copy of r with its resolving Decision
// set. resolvingDecision must be non-zero. Use WithoutResolvingDecision
// to clear a previously set resolving Decision.
func (r ConflictResolution) WithResolvingDecision(resolvingDecision core.DecisionRef) (ConflictResolution, error) {
	if resolvingDecision.IsZero() {
		return ConflictResolution{}, fmt.Errorf("decision: ConflictResolution.WithResolvingDecision: %w: resolving decision must not be zero", ErrInvalidConflictResolution)
	}
	r.resolvingDecision = resolvingDecision
	return r, nil
}

// WithoutResolvingDecision returns a copy of r with its resolving
// Decision cleared.
func (r ConflictResolution) WithoutResolvingDecision() ConflictResolution {
	r.resolvingDecision = core.DecisionRef{}
	return r
}

// WithProvenance returns a copy of r with its provenance set. provenance
// must be non-zero. Use WithoutProvenance to clear a previously set
// provenance.
func (r ConflictResolution) WithProvenance(provenance core.Provenance) (ConflictResolution, error) {
	if provenance.IsZero() {
		return ConflictResolution{}, fmt.Errorf("decision: ConflictResolution.WithProvenance: %w: provenance must not be zero", ErrInvalidConflictResolution)
	}
	r.provenance = provenance
	return r, nil
}

// WithoutProvenance returns a copy of r with its provenance cleared.
func (r ConflictResolution) WithoutProvenance() ConflictResolution {
	r.provenance = core.Provenance{}
	return r
}

// WithExtension returns a copy of r with its extension data set.
func (r ConflictResolution) WithExtension(extension core.Extension) ConflictResolution {
	r.extension = extension
	return r
}

func (r ConflictResolution) ID() ConflictResolutionID       { return r.id }
func (r ConflictResolution) Conflict() ConflictID           { return r.conflict }
func (r ConflictResolution) Mechanism() ResolutionMechanism { return r.mechanism }
func (r ConflictResolution) Statement() string              { return r.statement }

// ResolvingDecision returns r's declared resolving Decision, and whether
// one is set.
func (r ConflictResolution) ResolvingDecision() (core.DecisionRef, bool) {
	return r.resolvingDecision, !r.resolvingDecision.IsZero()
}

// Provenance returns r's declared provenance, and whether one is set.
func (r ConflictResolution) Provenance() (core.Provenance, bool) {
	return r.provenance, !r.provenance.IsZero()
}

func (r ConflictResolution) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r ConflictResolution) IsZero() bool { return r.id.IsZero() }

type conflictResolutionJSON struct {
	ID                ConflictResolutionID `json:"id"`
	Conflict          ConflictID           `json:"conflict"`
	Mechanism         ResolutionMechanism  `json:"mechanism"`
	Statement         string               `json:"statement"`
	ResolvingDecision *core.DecisionRef    `json:"resolving_decision,omitempty"`
	Provenance        *core.Provenance     `json:"provenance,omitempty"`
	Extension         *core.Extension      `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"id":..., "conflict":..., "mechanism":...,
// "statement":..., "resolving_decision":..., "provenance":...,
// "extension":...}, omitting resolving_decision, provenance, and
// extension when not set.
func (r ConflictResolution) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("decision: marshal ConflictResolution: %w", ErrInvalidConflictResolution)
	}
	raw := conflictResolutionJSON{
		ID:        r.id,
		Conflict:  r.conflict,
		Mechanism: r.mechanism,
		Statement: r.statement,
	}
	if !r.resolvingDecision.IsZero() {
		raw.ResolvingDecision = &r.resolvingDecision
	}
	if !r.provenance.IsZero() {
		raw.Provenance = &r.provenance
	}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

type conflictResolutionUnmarshalJSON struct {
	ID                ConflictResolutionID `json:"id"`
	Conflict          ConflictID           `json:"conflict"`
	Mechanism         ResolutionMechanism  `json:"mechanism"`
	Statement         string               `json:"statement"`
	ResolvingDecision json.RawMessage      `json:"resolving_decision"`
	Provenance        json.RawMessage      `json:"provenance"`
	Extension         json.RawMessage      `json:"extension"`
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as NewConflictResolution, WithResolvingDecision, and
// WithProvenance. An explicit JSON null for "resolving_decision",
// "provenance", or "extension" is rejected rather than silently treated
// as absent.
func (r *ConflictResolution) UnmarshalJSON(data []byte) error {
	var raw conflictResolutionUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal ConflictResolution: %w: %w", ErrInvalidConflictResolution, err)
	}
	result, err := NewConflictResolution(raw.ID, raw.Conflict, raw.Mechanism, raw.Statement)
	if err != nil {
		return err
	}
	if len(raw.ResolvingDecision) > 0 {
		if string(raw.ResolvingDecision) == "null" {
			return fmt.Errorf("decision: unmarshal ConflictResolution: %w: resolving_decision must not be null", ErrInvalidConflictResolution)
		}
		var resolvingDecision core.DecisionRef
		if err := json.Unmarshal(raw.ResolvingDecision, &resolvingDecision); err != nil {
			return fmt.Errorf("decision: unmarshal ConflictResolution: %w: %w", ErrInvalidConflictResolution, err)
		}
		if result, err = result.WithResolvingDecision(resolvingDecision); err != nil {
			return err
		}
	}
	if len(raw.Provenance) > 0 {
		if string(raw.Provenance) == "null" {
			return fmt.Errorf("decision: unmarshal ConflictResolution: %w: provenance must not be null", ErrInvalidConflictResolution)
		}
		var provenance core.Provenance
		if err := json.Unmarshal(raw.Provenance, &provenance); err != nil {
			return fmt.Errorf("decision: unmarshal ConflictResolution: %w: %w", ErrInvalidConflictResolution, err)
		}
		if result, err = result.WithProvenance(provenance); err != nil {
			return err
		}
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal ConflictResolution: %w: %w", ErrInvalidConflictResolution, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*r = result
	return nil
}
