package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

type invalidationSourceKind string

const (
	invalidationSourceDecision  invalidationSourceKind = "decision"
	invalidationSourceAuthority invalidationSourceKind = "authority"
)

// InvalidationSource identifies what invalidated a Decision (PEOS-004
// "Invalidated Decision": "Invalidation MUST identify the invalidating
// authority or Decision."). PEOS-004's "or" is not stated as an exclusive
// or in the spec text; this package nonetheless chooses exactly one
// canonical source per Invalidation, as a deliberate SDK architectural
// decision rather than a specification requirement: when the source is a
// Decision, that Decision's own core.AuthorityRef is already required and
// non-zero (see Decision.Authority), so carrying a second, independent
// AuthorityRef alongside it would duplicate that authority with no way
// for this package to detect a divergence between the two. "Who recorded
// this Invalidation" remains core.Provenance's role, not this type's.
//
// core.EngineeringSubjectRef is not used for this union: none of its
// subject kinds carries an authority payload, so it cannot express the
// authority arm at all.
type InvalidationSource struct {
	kind      invalidationSourceKind
	decision  core.DecisionRef
	authority core.AuthorityRef
}

// NewInvalidationSourceFromDecision validates ref and returns an
// InvalidationSource naming the invalidating Decision.
func NewInvalidationSourceFromDecision(ref core.DecisionRef) (InvalidationSource, error) {
	if ref.IsZero() {
		return InvalidationSource{}, fmt.Errorf("decision: NewInvalidationSourceFromDecision: %w: decision reference must not be zero", ErrInvalidDecisionInvalidation)
	}
	return InvalidationSource{kind: invalidationSourceDecision, decision: ref}, nil
}

// NewInvalidationSourceFromAuthority validates ref and returns an
// InvalidationSource naming the invalidating authority.
func NewInvalidationSourceFromAuthority(ref core.AuthorityRef) (InvalidationSource, error) {
	if ref.IsZero() {
		return InvalidationSource{}, fmt.Errorf("decision: NewInvalidationSourceFromAuthority: %w: authority reference must not be zero", ErrInvalidDecisionInvalidation)
	}
	return InvalidationSource{kind: invalidationSourceAuthority, authority: ref}, nil
}

// Kind returns s's discriminator, "decision" or "authority".
func (s InvalidationSource) Kind() string { return string(s.kind) }

// IsDecision reports whether s names an invalidating Decision.
func (s InvalidationSource) IsDecision() bool { return s.kind == invalidationSourceDecision }

// IsAuthority reports whether s names an invalidating authority.
func (s InvalidationSource) IsAuthority() bool { return s.kind == invalidationSourceAuthority }

// AsDecision returns s's Decision reference, and whether s is the
// Decision arm.
func (s InvalidationSource) AsDecision() (core.DecisionRef, bool) {
	if s.kind != invalidationSourceDecision {
		return core.DecisionRef{}, false
	}
	return s.decision, true
}

// AsAuthority returns s's authority reference, and whether s is the
// authority arm.
func (s InvalidationSource) AsAuthority() (core.AuthorityRef, bool) {
	if s.kind != invalidationSourceAuthority {
		return core.AuthorityRef{}, false
	}
	return s.authority, true
}

// Equal reports whether s and other are the same arm with the same
// referenced value.
func (s InvalidationSource) Equal(other InvalidationSource) bool {
	if s.kind != other.kind {
		return false
	}
	switch s.kind {
	case invalidationSourceDecision:
		return s.decision == other.decision
	case invalidationSourceAuthority:
		return s.authority == other.authority
	default:
		return true
	}
}

// IsZero reports whether s is the zero value.
func (s InvalidationSource) IsZero() bool { return s.kind == "" }

type invalidationSourceEnvelope struct {
	Kind invalidationSourceKind `json:"kind"`
	Ref  json.RawMessage        `json:"ref"`
}

// MarshalJSON encodes s as {"kind":"decision","ref":{...}} or
// {"kind":"authority","ref":{...}}.
func (s InvalidationSource) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("decision: marshal InvalidationSource: %w", ErrInvalidDecisionInvalidation)
	}
	var (
		refBytes []byte
		err      error
	)
	switch s.kind {
	case invalidationSourceDecision:
		refBytes, err = json.Marshal(s.decision)
	case invalidationSourceAuthority:
		refBytes, err = json.Marshal(s.authority)
	default:
		return nil, fmt.Errorf("decision: marshal InvalidationSource: %w", ErrInvalidDecisionInvalidation)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(invalidationSourceEnvelope{Kind: s.kind, Ref: refBytes})
}

// UnmarshalJSON decodes s from its {"kind":...,"ref":...} JSON form,
// applying the same validation as NewInvalidationSourceFromDecision and
// NewInvalidationSourceFromAuthority. An unrecognized or missing "kind",
// a missing "ref", or an explicit null "ref" are all rejected.
func (s *InvalidationSource) UnmarshalJSON(data []byte) error {
	var env invalidationSourceEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("decision: unmarshal InvalidationSource: %w: %w", ErrInvalidDecisionInvalidation, err)
	}
	if len(env.Ref) == 0 {
		return fmt.Errorf("decision: unmarshal InvalidationSource: %w: ref is required", ErrInvalidDecisionInvalidation)
	}
	if string(env.Ref) == "null" {
		return fmt.Errorf("decision: unmarshal InvalidationSource: %w: ref must not be null", ErrInvalidDecisionInvalidation)
	}

	var (
		result InvalidationSource
		err    error
	)
	switch env.Kind {
	case invalidationSourceDecision:
		var ref core.DecisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = NewInvalidationSourceFromDecision(ref)
		}
	case invalidationSourceAuthority:
		var ref core.AuthorityRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = NewInvalidationSourceFromAuthority(ref)
		}
	default:
		return fmt.Errorf("decision: unmarshal InvalidationSource: unrecognized kind %q: %w", env.Kind, ErrInvalidDecisionInvalidation)
	}
	if err != nil {
		return fmt.Errorf("decision: unmarshal InvalidationSource: %w: %w", ErrInvalidDecisionInvalidation, err)
	}
	*s = result
	return nil
}

// DecisionInvalidation is an immutable, independently identified
// governance record establishing that a Decision's normative effect has
// been withdrawn (PEOS-004 "Invalidated Decision": "A Decision is
// invalidated when its normative effect is withdrawn because its
// validity, authority, basis, or applicability is no longer accepted.
// Invalidation MUST identify the invalidating authority or Decision.
// Invalidation MUST preserve: the original Decision; its original
// Outcome; its historical Applicability; the reason for invalidation; the
// invalidation time or condition.").
//
// DecisionInvalidation is deliberately not any of the following:
//
//   - An Artifact, a Lifecycle Subject, or an Artifact Revision.
//   - A Decision Revision or a mutation of the invalidated Decision:
//     Decision has no revision mechanism (see doc.go).
//   - A peos/relation.Relation specialization: PEOS-004 lists
//     "invalidates" as a Decision Relation Type, but Relation carries no
//     reason and no effective condition/time. A Product MAY additionally
//     record a Relation as an index; this record remains authoritative.
//   - core.RecordCorrection: invalidation is a governance act, never a
//     typo, representation, or metadata fix. core.RecordRef's own union
//     deliberately excludes DecisionID.
//
// Preservation of the original Decision, its Outcome, and its historical
// Applicability is by reference, not by duplication: Decision.Outcome and
// Decision.Applicability are constructor-only on an immutable Decision,
// so DecisionInvalidation preserves them by naming the Decision with
// core.DecisionRef. This package does not resolve or verify that
// reference against any repository.
//
// Invalidation is non-retroactive by default: this type carries no field
// asserting that the Decision was never applicable, and its absence
// means exactly that -- non-retroactive. Retroactive effect requires an
// explicit Product contract and is not modeled here.
type DecisionInvalidation struct {
	id                  InvalidationID
	invalidatedDecision core.DecisionRef
	source              InvalidationSource
	reason              string
	effectiveAt         core.Timestamp
	effectiveCondition  string
	provenance          core.Provenance
	extension           core.Extension
}

// NewDecisionInvalidation validates its arguments and returns a
// DecisionInvalidation with no provenance or extension data. Use
// WithProvenance and WithExtension to add those.
//
// id, invalidatedDecision, and source must each be non-zero; reason must
// be non-empty after trimming surrounding whitespace. At least one of
// effectiveAt or effectiveCondition must be present; both MAY be present.
//
// A successful call always returns a fully valid record: every mandatory
// field is a required constructor argument, never a later With* call.
func NewDecisionInvalidation(
	id InvalidationID,
	invalidatedDecision core.DecisionRef,
	source InvalidationSource,
	reason string,
	effectiveAt core.Timestamp,
	effectiveCondition string,
) (DecisionInvalidation, error) {
	if id.IsZero() {
		return DecisionInvalidation{}, fmt.Errorf("decision: NewDecisionInvalidation: %w", ErrInvalidDecisionInvalidation)
	}
	if invalidatedDecision.IsZero() {
		return DecisionInvalidation{}, fmt.Errorf("decision: NewDecisionInvalidation: %w: invalidated decision must not be zero", ErrInvalidDecisionInvalidation)
	}
	if source.IsZero() {
		return DecisionInvalidation{}, fmt.Errorf("decision: NewDecisionInvalidation: %w: source must not be zero", ErrInvalidDecisionInvalidation)
	}
	if strings.TrimSpace(reason) == "" {
		return DecisionInvalidation{}, fmt.Errorf("decision: NewDecisionInvalidation: %w: reason must not be empty", ErrInvalidDecisionInvalidation)
	}
	if effectiveAt.IsZero() && strings.TrimSpace(effectiveCondition) == "" {
		return DecisionInvalidation{}, fmt.Errorf("decision: NewDecisionInvalidation: %w: at least one of effective time or effective condition is required", ErrInvalidDecisionInvalidation)
	}
	return DecisionInvalidation{
		id:                  id,
		invalidatedDecision: invalidatedDecision,
		source:              source,
		reason:              reason,
		effectiveAt:         effectiveAt,
		effectiveCondition:  effectiveCondition,
	}, nil
}

// WithProvenance returns a copy of i with its provenance set. provenance
// must be non-zero. Use WithoutProvenance to clear a previously set
// provenance.
func (i DecisionInvalidation) WithProvenance(provenance core.Provenance) (DecisionInvalidation, error) {
	if provenance.IsZero() {
		return DecisionInvalidation{}, fmt.Errorf("decision: DecisionInvalidation.WithProvenance: %w: provenance must not be zero", ErrInvalidDecisionInvalidation)
	}
	i.provenance = provenance
	return i, nil
}

// WithoutProvenance returns a copy of i with its provenance cleared.
func (i DecisionInvalidation) WithoutProvenance() DecisionInvalidation {
	i.provenance = core.Provenance{}
	return i
}

// WithExtension returns a copy of i with its extension data set.
func (i DecisionInvalidation) WithExtension(extension core.Extension) DecisionInvalidation {
	i.extension = extension
	return i
}

func (i DecisionInvalidation) ID() InvalidationID                    { return i.id }
func (i DecisionInvalidation) InvalidatedDecision() core.DecisionRef { return i.invalidatedDecision }
func (i DecisionInvalidation) Source() InvalidationSource            { return i.source }

// Reason returns i's invalidation reason. Reason is required and
// therefore always non-empty on a non-zero DecisionInvalidation.
func (i DecisionInvalidation) Reason() string { return i.reason }

// EffectiveAt returns i's effective time, and whether one is set.
func (i DecisionInvalidation) EffectiveAt() (core.Timestamp, bool) {
	return i.effectiveAt, !i.effectiveAt.IsZero()
}

// EffectiveCondition returns i's effective condition, and whether one is
// set.
func (i DecisionInvalidation) EffectiveCondition() (string, bool) {
	return i.effectiveCondition, i.effectiveCondition != ""
}

// Provenance returns i's declared provenance, and whether one is set.
func (i DecisionInvalidation) Provenance() (core.Provenance, bool) {
	return i.provenance, !i.provenance.IsZero()
}

func (i DecisionInvalidation) Extension() core.Extension { return i.extension }

// IsZero reports whether i is the zero value.
func (i DecisionInvalidation) IsZero() bool { return i.id.IsZero() }

type decisionInvalidationJSON struct {
	ID                  InvalidationID     `json:"id"`
	InvalidatedDecision core.DecisionRef   `json:"invalidated_decision"`
	Source              InvalidationSource `json:"source"`
	Reason              string             `json:"reason"`
	EffectiveAt         *core.Timestamp    `json:"effective_at,omitempty"`
	EffectiveCondition  string             `json:"effective_condition,omitempty"`
	Provenance          *core.Provenance   `json:"provenance,omitempty"`
	Extension           *core.Extension    `json:"extension,omitempty"`
}

// MarshalJSON encodes i as {"id":..., "invalidated_decision":...,
// "source":..., "reason":..., ...}, omitting effective_at,
// effective_condition, provenance, and extension when not set.
func (i DecisionInvalidation) MarshalJSON() ([]byte, error) {
	if i.IsZero() {
		return nil, fmt.Errorf("decision: marshal DecisionInvalidation: %w", ErrInvalidDecisionInvalidation)
	}
	raw := decisionInvalidationJSON{
		ID:                  i.id,
		InvalidatedDecision: i.invalidatedDecision,
		Source:              i.source,
		Reason:              i.reason,
	}
	if !i.effectiveAt.IsZero() {
		raw.EffectiveAt = &i.effectiveAt
	}
	if i.effectiveCondition != "" {
		raw.EffectiveCondition = i.effectiveCondition
	}
	if !i.provenance.IsZero() {
		raw.Provenance = &i.provenance
	}
	if !i.extension.IsZero() {
		raw.Extension = &i.extension
	}
	return json.Marshal(raw)
}

type decisionInvalidationUnmarshalJSON struct {
	ID                  InvalidationID     `json:"id"`
	InvalidatedDecision core.DecisionRef   `json:"invalidated_decision"`
	Source              InvalidationSource `json:"source"`
	Reason              string             `json:"reason"`
	EffectiveAt         json.RawMessage    `json:"effective_at"`
	EffectiveCondition  json.RawMessage    `json:"effective_condition"`
	Provenance          json.RawMessage    `json:"provenance"`
	Extension           json.RawMessage    `json:"extension"`
}

// UnmarshalJSON decodes i from its JSON form, applying the same
// validation as NewDecisionInvalidation and WithProvenance. An explicit
// JSON null for any optional field is rejected rather than silently
// treated as absent; an empty string for effective_condition when the key
// is present is likewise rejected.
func (i *DecisionInvalidation) UnmarshalJSON(data []byte) error {
	var raw decisionInvalidationUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: %w", ErrInvalidDecisionInvalidation, err)
	}

	var effectiveAt core.Timestamp
	if len(raw.EffectiveAt) > 0 {
		if string(raw.EffectiveAt) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: effective_at must not be null", ErrInvalidDecisionInvalidation)
		}
		if err := json.Unmarshal(raw.EffectiveAt, &effectiveAt); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: %w", ErrInvalidDecisionInvalidation, err)
		}
	}

	var effectiveCondition string
	if len(raw.EffectiveCondition) > 0 {
		if string(raw.EffectiveCondition) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: effective_condition must not be null", ErrInvalidDecisionInvalidation)
		}
		if err := json.Unmarshal(raw.EffectiveCondition, &effectiveCondition); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: %w", ErrInvalidDecisionInvalidation, err)
		}
		if effectiveCondition == "" {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: effective_condition must not be empty when present", ErrInvalidDecisionInvalidation)
		}
	}

	result, err := NewDecisionInvalidation(raw.ID, raw.InvalidatedDecision, raw.Source, raw.Reason, effectiveAt, effectiveCondition)
	if err != nil {
		return err
	}

	if len(raw.Provenance) > 0 {
		if string(raw.Provenance) == "null" {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: provenance must not be null", ErrInvalidDecisionInvalidation)
		}
		var provenance core.Provenance
		if err := json.Unmarshal(raw.Provenance, &provenance); err != nil {
			return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: %w", ErrInvalidDecisionInvalidation, err)
		}
		if result, err = result.WithProvenance(provenance); err != nil {
			return err
		}
	}

	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal DecisionInvalidation: %w: %w", ErrInvalidDecisionInvalidation, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*i = result
	return nil
}
