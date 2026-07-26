package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Decision is a PEOS-004 Decision: an immutable value with its own
// core.DecisionID, independent of any Artifact that documents it (see
// doc.go). A material semantic change to a Decision's content is a new
// Decision with a new identity, never a mutation of this value or a new
// Decision Revision -- this package defines no such revision mechanism.
//
// PEOS-004's Decision MUST-have list requires: a stable identity; a
// defined Subject or Decision Question; a Decision Outcome; an
// applicability scope; an identifiable authority requirement or basis. Of
// these, Subjects and Question are not both mandatory -- PEOS-004 permits
// either, or both, but not neither. Alternatives, Basis, Rationale, and
// Provenance are each optional: PEOS-004 either marks them SHOULD/MAY, or,
// for Provenance, never mentions them at all (see doc.go).
type Decision struct {
	id            core.DecisionID
	subjects      []core.EngineeringSubjectRef
	question      string
	outcome       Outcome
	applicability core.Scope
	authority     Authority
	alternatives  []Alternative
	basis         Basis
	rationale     string
	provenance    core.Provenance
	roles         []Role
	consequences  []Consequence
	extension     core.Extension
}

// New validates its arguments and returns a Decision with no Alternatives,
// Basis, Rationale, Provenance, or extension data. Use the With* methods to
// add those.
//
// id, outcome, applicability, and authority must each be non-zero. subjects
// and question are not both required, but at least one of a non-empty
// subjects slice or a non-empty question must be given; both MAY be given
// together. A zero-value subject among subjects is rejected.
func New(
	id core.DecisionID,
	subjects []core.EngineeringSubjectRef,
	question string,
	outcome Outcome,
	applicability core.Scope,
	authority Authority,
) (Decision, error) {
	if id.IsZero() {
		return Decision{}, fmt.Errorf("decision: New: %w", ErrInvalidDecision)
	}
	if len(subjects) == 0 && strings.TrimSpace(question) == "" {
		return Decision{}, fmt.Errorf("decision: New: %w: at least one subject or a non-empty question is required", ErrInvalidDecision)
	}
	var subs []core.EngineeringSubjectRef
	if len(subjects) > 0 {
		subs = make([]core.EngineeringSubjectRef, len(subjects))
		for idx, s := range subjects {
			if s.IsZero() {
				return Decision{}, fmt.Errorf("decision: New: %w", ErrInvalidDecisionSubject)
			}
			subs[idx] = s
		}
	}
	if outcome.IsZero() {
		return Decision{}, fmt.Errorf("decision: New: %w: outcome must not be zero", ErrInvalidDecision)
	}
	if applicability.IsZero() {
		return Decision{}, fmt.Errorf("decision: New: %w: applicability must not be zero", ErrInvalidDecision)
	}
	if authority.IsZero() {
		return Decision{}, fmt.Errorf("decision: New: %w", ErrInvalidAuthority)
	}
	return Decision{
		id:            id,
		subjects:      subs,
		question:      question,
		outcome:       outcome,
		applicability: applicability,
		authority:     authority,
	}, nil
}

// WithAlternatives returns a copy of d with its declared Alternatives set
// to exactly the values given, in the order given, replacing any previous
// Alternatives. A zero-value Alternative among alternatives is rejected.
// Calling with no arguments clears the declared Alternatives.
func (d Decision) WithAlternatives(alternatives ...Alternative) (Decision, error) {
	if len(alternatives) == 0 {
		d.alternatives = nil
		return d, nil
	}
	cp := make([]Alternative, len(alternatives))
	for idx, a := range alternatives {
		if a.IsZero() {
			return Decision{}, fmt.Errorf("decision: Decision.WithAlternatives: %w", ErrInvalidAlternative)
		}
		cp[idx] = a
	}
	d.alternatives = cp
	return d, nil
}

// WithBasis returns a copy of d with its Basis set. basis must be non-zero;
// use WithoutBasis to clear a previously set Basis.
func (d Decision) WithBasis(basis Basis) (Decision, error) {
	if basis.IsZero() {
		return Decision{}, fmt.Errorf("decision: Decision.WithBasis: %w", ErrInvalidBasis)
	}
	d.basis = basis
	return d, nil
}

// WithoutBasis returns a copy of d with its Basis cleared.
func (d Decision) WithoutBasis() Decision {
	d.basis = Basis{}
	return d
}

// WithRationale returns a copy of d with its Rationale set. rationale must
// be non-empty; use WithoutRationale to clear a previously set Rationale.
func (d Decision) WithRationale(rationale string) (Decision, error) {
	if strings.TrimSpace(rationale) == "" {
		return Decision{}, fmt.Errorf("decision: Decision.WithRationale: %w: rationale must not be empty", ErrInvalidDecision)
	}
	d.rationale = rationale
	return d, nil
}

// WithoutRationale returns a copy of d with its Rationale cleared.
func (d Decision) WithoutRationale() Decision {
	d.rationale = ""
	return d
}

// WithProvenance returns a copy of d with its Provenance set. provenance
// must be non-zero; use WithoutProvenance to clear a previously set
// Provenance.
func (d Decision) WithProvenance(provenance core.Provenance) (Decision, error) {
	if provenance.IsZero() {
		return Decision{}, fmt.Errorf("decision: Decision.WithProvenance: %w: provenance must not be zero", ErrInvalidDecision)
	}
	d.provenance = provenance
	return d, nil
}

// WithoutProvenance returns a copy of d with its Provenance cleared.
func (d Decision) WithoutProvenance() Decision {
	d.provenance = core.Provenance{}
	return d
}

// WithRoles returns a copy of d with its declared Roles set to exactly
// the values given, in the order given, replacing any previous Roles. A
// zero-value Role among roles is rejected. Calling with no arguments
// clears the declared Roles.
func (d Decision) WithRoles(roles ...Role) (Decision, error) {
	if len(roles) == 0 {
		d.roles = nil
		return d, nil
	}
	cp := make([]Role, len(roles))
	for idx, r := range roles {
		if r.IsZero() {
			return Decision{}, fmt.Errorf("decision: Decision.WithRoles: %w", ErrInvalidDecisionRole)
		}
		cp[idx] = r
	}
	d.roles = cp
	return d, nil
}

// WithConsequences returns a copy of d with its declared Consequences set
// to exactly the values given, in the order given, replacing any
// previous Consequences. A zero-value Consequence among consequences is
// rejected. Calling with no arguments clears the declared Consequences.
func (d Decision) WithConsequences(consequences ...Consequence) (Decision, error) {
	if len(consequences) == 0 {
		d.consequences = nil
		return d, nil
	}
	cp := make([]Consequence, len(consequences))
	for idx, c := range consequences {
		if c.IsZero() {
			return Decision{}, fmt.Errorf("decision: Decision.WithConsequences: %w", ErrInvalidConsequence)
		}
		cp[idx] = c
	}
	d.consequences = cp
	return d, nil
}

// WithExtension returns a copy of d with its extension data set.
func (d Decision) WithExtension(e core.Extension) Decision {
	d.extension = e
	return d
}

func (d Decision) ID() core.DecisionID { return d.id }

// Subjects returns a defensive copy of d's declared Subjects, in
// declaration order.
func (d Decision) Subjects() []core.EngineeringSubjectRef {
	if len(d.subjects) == 0 {
		return nil
	}
	cp := make([]core.EngineeringSubjectRef, len(d.subjects))
	copy(cp, d.subjects)
	return cp
}

// Question returns d's declared Decision Question, and whether one is set.
func (d Decision) Question() (string, bool) { return d.question, d.question != "" }

func (d Decision) Outcome() Outcome          { return d.outcome }
func (d Decision) Applicability() core.Scope { return d.applicability }
func (d Decision) Authority() Authority      { return d.authority }

// Alternatives returns a defensive copy of d's declared Alternatives, in
// declaration order.
func (d Decision) Alternatives() []Alternative {
	if len(d.alternatives) == 0 {
		return nil
	}
	cp := make([]Alternative, len(d.alternatives))
	copy(cp, d.alternatives)
	return cp
}

// Basis returns d's declared Basis, and whether one is set.
func (d Decision) Basis() (Basis, bool) { return d.basis, !d.basis.IsZero() }

// Rationale returns d's declared Rationale, and whether one is set.
func (d Decision) Rationale() (string, bool) { return d.rationale, d.rationale != "" }

// Provenance returns d's declared Provenance, and whether one is set.
func (d Decision) Provenance() (core.Provenance, bool) { return d.provenance, !d.provenance.IsZero() }

// Roles returns a defensive copy of d's declared Roles, in declaration
// order.
func (d Decision) Roles() []Role {
	if len(d.roles) == 0 {
		return nil
	}
	cp := make([]Role, len(d.roles))
	copy(cp, d.roles)
	return cp
}

// Consequences returns a defensive copy of d's declared Consequences, in
// declaration order.
func (d Decision) Consequences() []Consequence {
	if len(d.consequences) == 0 {
		return nil
	}
	cp := make([]Consequence, len(d.consequences))
	copy(cp, d.consequences)
	return cp
}

func (d Decision) Extension() core.Extension { return d.extension }

// IsZero reports whether d is the zero value.
func (d Decision) IsZero() bool { return d.id.IsZero() }

// Ref returns a core.DecisionRef identifying d.
func (d Decision) Ref() (core.DecisionRef, error) { return core.NewDecisionRef(d.id) }

// OutcomeRef returns a core.DecisionOutcomeRef identifying d's Outcome.
func (d Decision) OutcomeRef() (core.DecisionOutcomeRef, error) {
	return core.NewDecisionOutcomeRef(d.id)
}

type decisionJSON struct {
	ID            core.DecisionID              `json:"id"`
	Subjects      []core.EngineeringSubjectRef `json:"subjects,omitempty"`
	Question      string                       `json:"question,omitempty"`
	Outcome       Outcome                      `json:"outcome"`
	Applicability core.Scope                   `json:"applicability"`
	Authority     Authority                    `json:"authority"`
	Alternatives  []Alternative                `json:"alternatives,omitempty"`
	Basis         *Basis                       `json:"basis,omitempty"`
	Rationale     string                       `json:"rationale,omitempty"`
	Provenance    *core.Provenance             `json:"provenance,omitempty"`
	Roles         []Role                       `json:"roles,omitempty"`
	Consequences  []Consequence                `json:"consequences,omitempty"`
	Extension     *core.Extension              `json:"extension,omitempty"`
}

// MarshalJSON encodes d as {"id":..., "outcome":..., "applicability":...,
// "authority":..., ...}, omitting subjects, question, alternatives, basis,
// rationale, provenance, and extension when not set.
func (d Decision) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("decision: marshal Decision: %w", ErrInvalidDecision)
	}
	raw := decisionJSON{
		ID:            d.id,
		Outcome:       d.outcome,
		Applicability: d.applicability,
		Authority:     d.authority,
	}
	if len(d.subjects) > 0 {
		raw.Subjects = d.subjects
	}
	if d.question != "" {
		raw.Question = d.question
	}
	if len(d.alternatives) > 0 {
		raw.Alternatives = d.alternatives
	}
	if !d.basis.IsZero() {
		raw.Basis = &d.basis
	}
	if d.rationale != "" {
		raw.Rationale = d.rationale
	}
	if !d.provenance.IsZero() {
		raw.Provenance = &d.provenance
	}
	if len(d.roles) > 0 {
		raw.Roles = d.roles
	}
	if len(d.consequences) > 0 {
		raw.Consequences = d.consequences
	}
	if !d.extension.IsZero() {
		raw.Extension = &d.extension
	}
	return json.Marshal(raw)
}

type decisionUnmarshalJSON struct {
	ID            core.DecisionID `json:"id"`
	Subjects      json.RawMessage `json:"subjects"`
	Question      json.RawMessage `json:"question"`
	Outcome       Outcome         `json:"outcome"`
	Applicability core.Scope      `json:"applicability"`
	Authority     Authority       `json:"authority"`
	Alternatives  json.RawMessage `json:"alternatives"`
	Basis         json.RawMessage `json:"basis"`
	Rationale     json.RawMessage `json:"rationale"`
	Provenance    json.RawMessage `json:"provenance"`
	Roles         json.RawMessage `json:"roles"`
	Consequences  json.RawMessage `json:"consequences"`
	Extension     json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes d from its JSON form, applying the same validation
// as New and each With* method. An explicit JSON null for any optional
// field ("subjects", "question", "alternatives", "basis", "rationale",
// "provenance", "roles", "consequences", "extension") is rejected rather
// than silently treated as absent; an empty string for "question" or
// "rationale" when the key is present is likewise rejected.
func (d *Decision) UnmarshalJSON(data []byte) error {
	var raw decisionUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Decision: %w", err)
	}

	var subjects []core.EngineeringSubjectRef
	if len(raw.Subjects) > 0 {
		if string(raw.Subjects) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: subjects must not be null", ErrInvalidDecision)
		}
		if err := json.Unmarshal(raw.Subjects, &subjects); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
	}

	var question string
	if len(raw.Question) > 0 {
		if string(raw.Question) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: question must not be null", ErrInvalidDecision)
		}
		if err := json.Unmarshal(raw.Question, &question); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if question == "" {
			return fmt.Errorf("decision: unmarshal Decision: %w: question must not be empty when present", ErrInvalidDecision)
		}
	}

	result, err := New(raw.ID, subjects, question, raw.Outcome, raw.Applicability, raw.Authority)
	if err != nil {
		return err
	}

	if len(raw.Alternatives) > 0 {
		if string(raw.Alternatives) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: alternatives must not be null", ErrInvalidDecision)
		}
		var alternatives []Alternative
		if err := json.Unmarshal(raw.Alternatives, &alternatives); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if result, err = result.WithAlternatives(alternatives...); err != nil {
			return err
		}
	}

	if len(raw.Basis) > 0 {
		if string(raw.Basis) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: basis must not be null", ErrInvalidDecision)
		}
		var basis Basis
		if err := json.Unmarshal(raw.Basis, &basis); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if result, err = result.WithBasis(basis); err != nil {
			return err
		}
	}

	if len(raw.Rationale) > 0 {
		if string(raw.Rationale) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: rationale must not be null", ErrInvalidDecision)
		}
		var rationale string
		if err := json.Unmarshal(raw.Rationale, &rationale); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if result, err = result.WithRationale(rationale); err != nil {
			return err
		}
	}

	if len(raw.Provenance) > 0 {
		if string(raw.Provenance) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: provenance must not be null", ErrInvalidDecision)
		}
		var provenance core.Provenance
		if err := json.Unmarshal(raw.Provenance, &provenance); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if result, err = result.WithProvenance(provenance); err != nil {
			return err
		}
	}

	if len(raw.Roles) > 0 {
		if string(raw.Roles) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: roles must not be null", ErrInvalidDecision)
		}
		var roles []Role
		if err := json.Unmarshal(raw.Roles, &roles); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if result, err = result.WithRoles(roles...); err != nil {
			return err
		}
	}

	if len(raw.Consequences) > 0 {
		if string(raw.Consequences) == "null" {
			return fmt.Errorf("decision: unmarshal Decision: %w: consequences must not be null", ErrInvalidDecision)
		}
		var consequences []Consequence
		if err := json.Unmarshal(raw.Consequences, &consequences); err != nil {
			return fmt.Errorf("decision: unmarshal Decision: %w", err)
		}
		if result, err = result.WithConsequences(consequences...); err != nil {
			return err
		}
	}

	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Decision: %w: %w", ErrInvalidDecision, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*d = result
	return nil
}
