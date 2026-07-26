package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// ConflictID is the package-local identity of a DecisionConflict. No
// specification outside PEOS-004 references a Decision Conflict record,
// and PEOS-004 does not itself name a dedicated identity for one, so this
// identity is not promoted to peos/core -- see SupersessionID's own doc
// comment (identity.go) for why a package-local identity is structurally
// distinct rather than a shared wrapper.
//
// A surrogate identity is required here, not merely convenient: the
// natural key (decisionA, decisionB, overlappingScope) is insufficient,
// because the same two Decisions MAY conflict in the same scope over two
// materially different incompatible Outcomes (PEOS-004 :1123).
type ConflictID struct{ decisionConflictID string }

// NewConflictID validates value and returns a ConflictID. Surrounding
// whitespace is trimmed; the result is rejected if empty.
func NewConflictID(value string) (ConflictID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ConflictID{}, fmt.Errorf("decision: NewConflictID: %w", ErrInvalidDecisionConflict)
	}
	return ConflictID{decisionConflictID: trimmed}, nil
}

// String returns the opaque identity value.
func (id ConflictID) String() string { return id.decisionConflictID }

// IsZero reports whether id is the zero value.
func (id ConflictID) IsZero() bool { return id.decisionConflictID == "" }

// Equal reports whether id and other carry the same identity value.
func (id ConflictID) Equal(other ConflictID) bool {
	return id.decisionConflictID == other.decisionConflictID
}

// MarshalJSON encodes id as a JSON string.
func (id ConflictID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("decision: marshal ConflictID: %w", ErrInvalidDecisionConflict)
	}
	return json.Marshal(id.decisionConflictID)
}

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewConflictID.
func (id *ConflictID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("decision: unmarshal ConflictID: %w", err)
	}
	v, err := NewConflictID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// DecisionConflict is an immutable, independently identified governance
// record establishing that two applicable Decisions establish
// incompatible normative intent within overlapping Applicability
// (PEOS-004 "Decision Conflict": "A Decision Conflict exists when two
// applicable Decisions establish incompatible normative intent within
// overlapping Applicability. A conflict MUST identify: the conflicting
// Decisions; the overlapping scope; the incompatible Outcomes or
// Engineering Commitments; the applicable authority or priority rules.").
//
// DecisionConflict is binary, not N-ary: PEOS-004's own definitional
// clause is "when two applicable Decisions ...". A jointly-unsatisfiable
// triple of Decisions is three pairwise conflicts, each independently
// identified; PEOS-004 does not define an N-ary conflict concept.
// decisionA and decisionB must differ -- a Decision has no revision
// mechanism (see doc.go), so it cannot conflict with itself.
//
// DecisionConflict carries independent identity (ConflictID) because
// PEOS-004's Conflict Invariant (":1342 Conflicting applicable Decisions
// require explicit resolution") presupposes a conflict that exists while
// unresolved, and because ":1126 A Runtime MUST NOT silently resolve a
// Decision Conflict" requires the unresolved state itself to be
// inspectable. A conflict recorded only inside its own resolution could
// never be observed as unresolved, making the invariant unrepresentable.
//
// governingRule is required, not optional: this is what makes
// DecisionConflict the analyzed conflict PEOS-004 :1121 defines, as
// opposed to a Runtime's raw pre-analysis detection (:1360, "detect
// conflicts" is a permitted Runtime activity). A Runtime's raw detection
// is represented as a peos/relation.Relation carrying
// core.RelationTypeConflict; peos/decision does not import peos/relation
// (see DecisionSupersession's own doc comment for the identical argument
// against reusing Relation for a record with its own required fields,
// identity, and no source/target-equality check).
//
// DecisionConflict is deliberately not any of the following:
//
//   - An Artifact or a Lifecycle Subject: PEOS-004 never calls Conflict
//     an Artifact (unlike Decision Record, :861), and it names no
//     Conflict lifecycle States anywhere.
//   - A peos/relation.Relation specialization: Relation cannot carry
//     :1123 (the incompatible Outcomes or Engineering Commitments) or
//     :1124 (the applicable authority or priority rules), has no
//     identity for a ConflictResolution to reference, and does not
//     reject source == target.
//
// incompatibility is a statement, not a structural reference:
// core.DecisionOutcomeRef and core.EngineeringCommitmentRef are both
// derived deterministically from core.DecisionID (see their own doc
// comments in peos/core/reference.go), so referencing them would carry
// no information beyond decisionA and decisionB already carry. :1123
// asks what is incompatible about the two Decisions' Outcomes or
// Commitments; prose is the only lossless carrier available to this
// package.
type DecisionConflict struct {
	id               ConflictID
	decisionA        core.DecisionRef
	decisionB        core.DecisionRef
	overlappingScope core.Scope
	incompatibility  string
	governingRule    string
	provenance       core.Provenance
	extension        core.Extension
}

// NewDecisionConflict validates its arguments and returns a
// DecisionConflict with no provenance or extension data. Use
// WithProvenance and WithExtension to add those.
//
// id, decisionA, decisionB, overlappingScope must each be non-zero;
// decisionA and decisionB must differ. incompatibility and governingRule
// must each be non-empty after trimming surrounding whitespace; both are
// stored as given.
//
// A successful call always returns a fully valid record: every mandatory
// field is a required constructor argument, never a later With* call.
func NewDecisionConflict(
	id ConflictID,
	decisionA core.DecisionRef,
	decisionB core.DecisionRef,
	overlappingScope core.Scope,
	incompatibility string,
	governingRule string,
) (DecisionConflict, error) {
	if id.IsZero() {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w", ErrInvalidDecisionConflict)
	}
	if decisionA.IsZero() {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w: decision A must not be zero", ErrInvalidDecisionConflict)
	}
	if decisionB.IsZero() {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w: decision B must not be zero", ErrInvalidDecisionConflict)
	}
	if decisionA == decisionB {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w: decision A and decision B must differ", ErrInvalidDecisionConflict)
	}
	if overlappingScope.IsZero() {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w: overlapping scope must not be zero", ErrInvalidDecisionConflict)
	}
	if strings.TrimSpace(incompatibility) == "" {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w: incompatibility must not be empty", ErrInvalidDecisionConflict)
	}
	if strings.TrimSpace(governingRule) == "" {
		return DecisionConflict{}, fmt.Errorf("decision: NewDecisionConflict: %w: governing rule must not be empty", ErrInvalidDecisionConflict)
	}
	return DecisionConflict{
		id:               id,
		decisionA:        decisionA,
		decisionB:        decisionB,
		overlappingScope: overlappingScope,
		incompatibility:  incompatibility,
		governingRule:    governingRule,
	}, nil
}

// WithProvenance returns a copy of c with its provenance set. provenance
// must be non-zero. Use WithoutProvenance to clear a previously set
// provenance.
func (c DecisionConflict) WithProvenance(provenance core.Provenance) (DecisionConflict, error) {
	if provenance.IsZero() {
		return DecisionConflict{}, fmt.Errorf("decision: DecisionConflict.WithProvenance: %w: provenance must not be zero", ErrInvalidDecisionConflict)
	}
	c.provenance = provenance
	return c, nil
}

// WithoutProvenance returns a copy of c with its provenance cleared.
func (c DecisionConflict) WithoutProvenance() DecisionConflict {
	c.provenance = core.Provenance{}
	return c
}

// WithExtension returns a copy of c with its extension data set.
func (c DecisionConflict) WithExtension(extension core.Extension) DecisionConflict {
	c.extension = extension
	return c
}

func (c DecisionConflict) ID() ConflictID { return c.id }

// DecisionA returns one of the two conflicting Decisions. DecisionA and
// DecisionB are an unordered pair; no normative meaning attaches to which
// Decision is A and which is B.
func (c DecisionConflict) DecisionA() core.DecisionRef { return c.decisionA }

// DecisionB returns the other of the two conflicting Decisions.
func (c DecisionConflict) DecisionB() core.DecisionRef { return c.decisionB }

func (c DecisionConflict) OverlappingScope() core.Scope { return c.overlappingScope }
func (c DecisionConflict) Incompatibility() string      { return c.incompatibility }
func (c DecisionConflict) GoverningRule() string        { return c.governingRule }

// Provenance returns c's declared provenance, and whether one is set.
func (c DecisionConflict) Provenance() (core.Provenance, bool) {
	return c.provenance, !c.provenance.IsZero()
}

func (c DecisionConflict) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c DecisionConflict) IsZero() bool { return c.id.IsZero() }

type decisionConflictJSON struct {
	ID               ConflictID       `json:"id"`
	DecisionA        core.DecisionRef `json:"decision_a"`
	DecisionB        core.DecisionRef `json:"decision_b"`
	OverlappingScope core.Scope       `json:"overlapping_scope"`
	Incompatibility  string           `json:"incompatibility"`
	GoverningRule    string           `json:"governing_rule"`
	Provenance       *core.Provenance `json:"provenance,omitempty"`
	Extension        *core.Extension  `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"id":..., "decision_a":..., "decision_b":...,
// "overlapping_scope":..., "incompatibility":..., "governing_rule":...,
// "provenance":..., "extension":...}, omitting provenance and extension
// when not set.
func (c DecisionConflict) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("decision: marshal DecisionConflict: %w", ErrInvalidDecisionConflict)
	}
	raw := decisionConflictJSON{
		ID:               c.id,
		DecisionA:        c.decisionA,
		DecisionB:        c.decisionB,
		OverlappingScope: c.overlappingScope,
		Incompatibility:  c.incompatibility,
		GoverningRule:    c.governingRule,
	}
	if !c.provenance.IsZero() {
		raw.Provenance = &c.provenance
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

type decisionConflictUnmarshalJSON struct {
	ID               ConflictID       `json:"id"`
	DecisionA        core.DecisionRef `json:"decision_a"`
	DecisionB        core.DecisionRef `json:"decision_b"`
	OverlappingScope core.Scope       `json:"overlapping_scope"`
	Incompatibility  string           `json:"incompatibility"`
	GoverningRule    string           `json:"governing_rule"`
	Provenance       json.RawMessage  `json:"provenance"`
	Extension        json.RawMessage  `json:"extension"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same
// validation as NewDecisionConflict and WithProvenance. An explicit JSON
// null for "provenance" or "extension" is rejected rather than silently
// treated as absent.
func (c *DecisionConflict) UnmarshalJSON(data []byte) error {
	var raw decisionConflictUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal DecisionConflict: %w: %w", ErrInvalidDecisionConflict, err)
	}
	result, err := NewDecisionConflict(raw.ID, raw.DecisionA, raw.DecisionB, raw.OverlappingScope, raw.Incompatibility, raw.GoverningRule)
	if err != nil {
		return err
	}
	if len(raw.Provenance) > 0 {
		if string(raw.Provenance) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionConflict: %w: provenance must not be null", ErrInvalidDecisionConflict)
		}
		var provenance core.Provenance
		if err := json.Unmarshal(raw.Provenance, &provenance); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionConflict: %w: %w", ErrInvalidDecisionConflict, err)
		}
		if result, err = result.WithProvenance(provenance); err != nil {
			return err
		}
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal DecisionConflict: %w: %w", ErrInvalidDecisionConflict, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*c = result
	return nil
}
