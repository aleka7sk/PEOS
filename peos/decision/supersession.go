package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

type supersessionExtentKind string

const (
	supersessionExtentComplete supersessionExtentKind = "complete"
	supersessionExtentPartial  supersessionExtentKind = "partial"
)

// SupersessionExtent is PEOS-004's closed complete/partial distinction
// (Superseded Decision: "Supersession MUST identify: ... whether
// supersession is complete or partial"). It is a closed two-variant
// discriminator, not an open core.VocabularyValue wrapper: PEOS-004's own
// Extension Model enumerates exactly what a Product MAY extend (Decision
// Types, significance levels, lifecycle profiles, roles, authority rules,
// approval policies, Alternative classifications, Decision Relation
// Types, required Record fields, validation requirements,
// conflict-resolution policies, retention policies), and supersession
// completeness is not on that list. The same section then states
// extensions "MUST NOT redefine the core meaning of ... supersession." A
// Product-defined third extent value would do exactly that, and would
// also make the remaining-scope invariant below unenforceable for that
// value.
//
// The partial variant carries its own required remaining scope ("Where
// supersession is partial, the remaining applicable scope MUST be
// deterministically identifiable"), so the illegal combination (a partial
// supersession with no remaining scope, or a complete supersession
// carrying one) is unrepresentable rather than merely rejected -- the
// same shape peos/requirement uses for Applicability's
// unrestricted/scoped distinction.
type SupersessionExtent struct {
	kind           supersessionExtentKind
	remainingScope core.Scope
}

// NewCompleteSupersessionExtent returns a complete SupersessionExtent.
func NewCompleteSupersessionExtent() SupersessionExtent {
	return SupersessionExtent{kind: supersessionExtentComplete}
}

// NewPartialSupersessionExtent validates remainingScope and returns a
// partial SupersessionExtent. remainingScope must be non-zero.
func NewPartialSupersessionExtent(remainingScope core.Scope) (SupersessionExtent, error) {
	if remainingScope.IsZero() {
		return SupersessionExtent{}, fmt.Errorf("decision: NewPartialSupersessionExtent: %w: remaining scope must not be zero", ErrInvalidDecisionSupersession)
	}
	return SupersessionExtent{kind: supersessionExtentPartial, remainingScope: remainingScope}, nil
}

// IsComplete reports whether e is the complete variant.
func (e SupersessionExtent) IsComplete() bool { return e.kind == supersessionExtentComplete }

// IsPartial reports whether e is the partial variant.
func (e SupersessionExtent) IsPartial() bool { return e.kind == supersessionExtentPartial }

// RemainingScope returns e's remaining scope, and whether e is the
// partial variant -- the only variant that carries one.
func (e SupersessionExtent) RemainingScope() (core.Scope, bool) {
	if e.kind != supersessionExtentPartial {
		return core.Scope{}, false
	}
	return e.remainingScope, true
}

// String returns "complete" or "partial", or "" for the zero value.
func (e SupersessionExtent) String() string { return string(e.kind) }

// Equal reports whether e and other are the same variant with, for the
// partial variant, the same remaining scope.
func (e SupersessionExtent) Equal(other SupersessionExtent) bool {
	if e.kind != other.kind {
		return false
	}
	if e.kind == supersessionExtentPartial {
		return e.remainingScope.Equal(other.remainingScope)
	}
	return true
}

// IsZero reports whether e is the zero value.
func (e SupersessionExtent) IsZero() bool { return e.kind == "" }

type supersessionExtentJSON struct {
	Kind           supersessionExtentKind `json:"kind"`
	RemainingScope *core.Scope            `json:"remaining_scope,omitempty"`
}

// MarshalJSON encodes e as {"kind":"complete"} or {"kind":"partial",
// "remaining_scope":{...}}.
func (e SupersessionExtent) MarshalJSON() ([]byte, error) {
	if e.IsZero() {
		return nil, fmt.Errorf("decision: marshal SupersessionExtent: %w", ErrInvalidDecisionSupersession)
	}
	raw := supersessionExtentJSON{Kind: e.kind}
	if e.kind == supersessionExtentPartial {
		raw.RemainingScope = &e.remainingScope
	}
	return json.Marshal(raw)
}

type supersessionExtentUnmarshalJSON struct {
	Kind           supersessionExtentKind `json:"kind"`
	RemainingScope json.RawMessage        `json:"remaining_scope"`
}

// UnmarshalJSON decodes e from its JSON form. An unrecognized or missing
// kind is rejected. A complete extent carrying "remaining_scope" is
// rejected as ambiguous; a partial extent missing "remaining_scope", or
// carrying an explicit null for it, is rejected.
func (e *SupersessionExtent) UnmarshalJSON(data []byte) error {
	var raw supersessionExtentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal SupersessionExtent: %w", err)
	}
	switch raw.Kind {
	case supersessionExtentComplete:
		if len(raw.RemainingScope) > 0 {
			return fmt.Errorf("decision: unmarshal SupersessionExtent: %w: complete extent must not carry remaining_scope", ErrInvalidDecisionSupersession)
		}
		*e = NewCompleteSupersessionExtent()
		return nil
	case supersessionExtentPartial:
		if len(raw.RemainingScope) == 0 {
			return fmt.Errorf("decision: unmarshal SupersessionExtent: %w: partial extent requires remaining_scope", ErrInvalidDecisionSupersession)
		}
		if string(raw.RemainingScope) == "null" {
			return fmt.Errorf("decision: unmarshal SupersessionExtent: %w: remaining_scope must not be null", ErrInvalidDecisionSupersession)
		}
		var scope core.Scope
		if err := json.Unmarshal(raw.RemainingScope, &scope); err != nil {
			return fmt.Errorf("decision: unmarshal SupersessionExtent: %w: %w", ErrInvalidDecisionSupersession, err)
		}
		result, err := NewPartialSupersessionExtent(scope)
		if err != nil {
			return err
		}
		*e = result
		return nil
	default:
		return fmt.Errorf("decision: unmarshal SupersessionExtent: unrecognized kind %q: %w", raw.Kind, ErrInvalidDecisionSupersession)
	}
}

// DecisionSupersession is an immutable, independently identified
// governance record establishing that one Decision supersedes another
// (PEOS-004 "Superseded Decision": "A Decision is superseded when another
// applicable Decision explicitly replaces all or part of its normative
// effect. Supersession MUST identify: the superseding Decision; the
// superseded Decision; the affected Applicability; whether supersession
// is complete or partial; the effective condition or time.").
//
// DecisionSupersession is deliberately not any of the following:
//
//   - An Artifact, a Lifecycle Subject, or an Artifact Revision: PEOS-004
//     never calls Supersession an Artifact (unlike Decision Record), and
//     it carries no lifecycle of its own.
//   - A Decision Revision: Decision has no revision mechanism (see
//     doc.go); a Supersession does not revise the superseded Decision --
//     PEOS-004 requires "Supersession MUST NOT delete or rewrite the
//     superseded Decision."
//   - A peos/relation.Relation specialization: PEOS-004 lists "supersedes"
//     as a Decision Relation Type, but Relation (peos/relation) carries no
//     extent, no effective condition/time, and no reason -- it cannot
//     express what this record must. A Product MAY additionally record a
//     Relation as an index; this record remains authoritative.
//     peos/decision does not import peos/relation.
//
// Preservation is by reference, not by duplication. Decision.Outcome and
// Decision.Applicability are constructor-only on an immutable Decision
// (see doc.go), so DecisionSupersession preserves the superseded Decision
// by naming it with core.DecisionRef rather than copying its content.
// This package does not resolve or verify that reference against any
// repository.
type DecisionSupersession struct {
	id                  SupersessionID
	supersedingDecision core.DecisionRef
	supersededDecision  core.DecisionRef
	affectedScope       core.Scope
	extent              SupersessionExtent
	effectiveAt         core.Timestamp
	effectiveCondition  string
	reason              string
	provenance          core.Provenance
	extension           core.Extension
}

// NewDecisionSupersession validates its arguments and returns a
// DecisionSupersession with no reason, provenance, or extension data. Use
// WithReason, WithProvenance, and WithExtension to add those.
//
// id, superseding, superseded, affectedScope, and extent must each be
// non-zero; superseding must differ from superseded. At least one of
// effectiveAt or effectiveCondition must be present; both MAY be present.
// effectiveAt is treated as absent when zero; effectiveCondition is
// treated as absent when empty after trimming surrounding whitespace, but
// is stored as given.
//
// A successful call always returns a fully valid record: every mandatory
// field is a required constructor argument, never a later With* call.
func NewDecisionSupersession(
	id SupersessionID,
	superseding core.DecisionRef,
	superseded core.DecisionRef,
	affectedScope core.Scope,
	extent SupersessionExtent,
	effectiveAt core.Timestamp,
	effectiveCondition string,
) (DecisionSupersession, error) {
	if id.IsZero() {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w", ErrInvalidDecisionSupersession)
	}
	if superseding.IsZero() {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w: superseding decision must not be zero", ErrInvalidDecisionSupersession)
	}
	if superseded.IsZero() {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w: superseded decision must not be zero", ErrInvalidDecisionSupersession)
	}
	if superseding == superseded {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w: superseding and superseded decision must differ", ErrInvalidDecisionSupersession)
	}
	if affectedScope.IsZero() {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w: affected scope must not be zero", ErrInvalidDecisionSupersession)
	}
	if extent.IsZero() {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w: extent must not be zero", ErrInvalidDecisionSupersession)
	}
	if effectiveAt.IsZero() && strings.TrimSpace(effectiveCondition) == "" {
		return DecisionSupersession{}, fmt.Errorf("decision: NewDecisionSupersession: %w: at least one of effective time or effective condition is required", ErrInvalidDecisionSupersession)
	}
	return DecisionSupersession{
		id:                  id,
		supersedingDecision: superseding,
		supersededDecision:  superseded,
		affectedScope:       affectedScope,
		extent:              extent,
		effectiveAt:         effectiveAt,
		effectiveCondition:  effectiveCondition,
	}, nil
}

// WithReason returns a copy of s with its reason set. reason must be
// non-empty after trimming surrounding whitespace. Use WithoutReason to
// clear a previously set reason.
func (s DecisionSupersession) WithReason(reason string) (DecisionSupersession, error) {
	if strings.TrimSpace(reason) == "" {
		return DecisionSupersession{}, fmt.Errorf("decision: DecisionSupersession.WithReason: %w: reason must not be empty", ErrInvalidDecisionSupersession)
	}
	s.reason = reason
	return s, nil
}

// WithoutReason returns a copy of s with its reason cleared.
func (s DecisionSupersession) WithoutReason() DecisionSupersession {
	s.reason = ""
	return s
}

// WithProvenance returns a copy of s with its provenance set. provenance
// must be non-zero. Use WithoutProvenance to clear a previously set
// provenance.
func (s DecisionSupersession) WithProvenance(provenance core.Provenance) (DecisionSupersession, error) {
	if provenance.IsZero() {
		return DecisionSupersession{}, fmt.Errorf("decision: DecisionSupersession.WithProvenance: %w: provenance must not be zero", ErrInvalidDecisionSupersession)
	}
	s.provenance = provenance
	return s, nil
}

// WithoutProvenance returns a copy of s with its provenance cleared.
func (s DecisionSupersession) WithoutProvenance() DecisionSupersession {
	s.provenance = core.Provenance{}
	return s
}

// WithExtension returns a copy of s with its extension data set.
func (s DecisionSupersession) WithExtension(extension core.Extension) DecisionSupersession {
	s.extension = extension
	return s
}

func (s DecisionSupersession) ID() SupersessionID                    { return s.id }
func (s DecisionSupersession) SupersedingDecision() core.DecisionRef { return s.supersedingDecision }
func (s DecisionSupersession) SupersededDecision() core.DecisionRef  { return s.supersededDecision }
func (s DecisionSupersession) AffectedScope() core.Scope             { return s.affectedScope }
func (s DecisionSupersession) Extent() SupersessionExtent            { return s.extent }

// EffectiveAt returns s's effective time, and whether one is set.
func (s DecisionSupersession) EffectiveAt() (core.Timestamp, bool) {
	return s.effectiveAt, !s.effectiveAt.IsZero()
}

// EffectiveCondition returns s's effective condition, and whether one is
// set.
func (s DecisionSupersession) EffectiveCondition() (string, bool) {
	return s.effectiveCondition, s.effectiveCondition != ""
}

// Reason returns s's declared reason, and whether one is set.
func (s DecisionSupersession) Reason() (string, bool) { return s.reason, s.reason != "" }

// Provenance returns s's declared provenance, and whether one is set.
func (s DecisionSupersession) Provenance() (core.Provenance, bool) {
	return s.provenance, !s.provenance.IsZero()
}

func (s DecisionSupersession) Extension() core.Extension { return s.extension }

// IsZero reports whether s is the zero value.
func (s DecisionSupersession) IsZero() bool { return s.id.IsZero() }

type decisionSupersessionJSON struct {
	ID                  SupersessionID     `json:"id"`
	SupersedingDecision core.DecisionRef   `json:"superseding_decision"`
	SupersededDecision  core.DecisionRef   `json:"superseded_decision"`
	AffectedScope       core.Scope         `json:"affected_scope"`
	Extent              SupersessionExtent `json:"extent"`
	EffectiveAt         *core.Timestamp    `json:"effective_at,omitempty"`
	EffectiveCondition  string             `json:"effective_condition,omitempty"`
	Reason              string             `json:"reason,omitempty"`
	Provenance          *core.Provenance   `json:"provenance,omitempty"`
	Extension           *core.Extension    `json:"extension,omitempty"`
}

// MarshalJSON encodes s as {"id":..., "superseding_decision":...,
// "superseded_decision":..., "affected_scope":..., "extent":..., ...},
// omitting effective_at, effective_condition, reason, provenance, and
// extension when not set.
func (s DecisionSupersession) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("decision: marshal DecisionSupersession: %w", ErrInvalidDecisionSupersession)
	}
	raw := decisionSupersessionJSON{
		ID:                  s.id,
		SupersedingDecision: s.supersedingDecision,
		SupersededDecision:  s.supersededDecision,
		AffectedScope:       s.affectedScope,
		Extent:              s.extent,
	}
	if !s.effectiveAt.IsZero() {
		raw.EffectiveAt = &s.effectiveAt
	}
	if s.effectiveCondition != "" {
		raw.EffectiveCondition = s.effectiveCondition
	}
	if s.reason != "" {
		raw.Reason = s.reason
	}
	if !s.provenance.IsZero() {
		raw.Provenance = &s.provenance
	}
	if !s.extension.IsZero() {
		raw.Extension = &s.extension
	}
	return json.Marshal(raw)
}

type decisionSupersessionUnmarshalJSON struct {
	ID                  SupersessionID     `json:"id"`
	SupersedingDecision core.DecisionRef   `json:"superseding_decision"`
	SupersededDecision  core.DecisionRef   `json:"superseded_decision"`
	AffectedScope       core.Scope         `json:"affected_scope"`
	Extent              SupersessionExtent `json:"extent"`
	EffectiveAt         json.RawMessage    `json:"effective_at"`
	EffectiveCondition  json.RawMessage    `json:"effective_condition"`
	Reason              json.RawMessage    `json:"reason"`
	Provenance          json.RawMessage    `json:"provenance"`
	Extension           json.RawMessage    `json:"extension"`
}

// UnmarshalJSON decodes s from its JSON form, applying the same
// validation as NewDecisionSupersession, WithReason, and WithProvenance.
// An explicit JSON null for any optional field is rejected rather than
// silently treated as absent; an empty string for effective_condition or
// reason when the key is present is likewise rejected.
func (s *DecisionSupersession) UnmarshalJSON(data []byte) error {
	var raw decisionSupersessionUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: %w", ErrInvalidDecisionSupersession, err)
	}

	var effectiveAt core.Timestamp
	if len(raw.EffectiveAt) > 0 {
		if string(raw.EffectiveAt) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: effective_at must not be null", ErrInvalidDecisionSupersession)
		}
		if err := json.Unmarshal(raw.EffectiveAt, &effectiveAt); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: %w", ErrInvalidDecisionSupersession, err)
		}
	}

	var effectiveCondition string
	if len(raw.EffectiveCondition) > 0 {
		if string(raw.EffectiveCondition) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: effective_condition must not be null", ErrInvalidDecisionSupersession)
		}
		if err := json.Unmarshal(raw.EffectiveCondition, &effectiveCondition); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: %w", ErrInvalidDecisionSupersession, err)
		}
		if effectiveCondition == "" {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: effective_condition must not be empty when present", ErrInvalidDecisionSupersession)
		}
	}

	result, err := NewDecisionSupersession(raw.ID, raw.SupersedingDecision, raw.SupersededDecision, raw.AffectedScope, raw.Extent, effectiveAt, effectiveCondition)
	if err != nil {
		return err
	}

	if len(raw.Reason) > 0 {
		if string(raw.Reason) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: reason must not be null", ErrInvalidDecisionSupersession)
		}
		var reason string
		if err := json.Unmarshal(raw.Reason, &reason); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: %w", ErrInvalidDecisionSupersession, err)
		}
		if result, err = result.WithReason(reason); err != nil {
			return err
		}
	}

	if len(raw.Provenance) > 0 {
		if string(raw.Provenance) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: provenance must not be null", ErrInvalidDecisionSupersession)
		}
		var provenance core.Provenance
		if err := json.Unmarshal(raw.Provenance, &provenance); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: %w", ErrInvalidDecisionSupersession, err)
		}
		if result, err = result.WithProvenance(provenance); err != nil {
			return err
		}
	}

	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal DecisionSupersession: %w: %w", ErrInvalidDecisionSupersession, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*s = result
	return nil
}
