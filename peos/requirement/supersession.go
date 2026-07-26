package requirement

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

type lifecycleConsequenceKind string

const (
	lifecycleConsequenceIdentified lifecycleConsequenceKind = "identified"
	lifecycleConsequenceNone       lifecycleConsequenceKind = "none"
)

// LifecycleConsequence is a Supersession's declared resulting effective
// status or Lifecycle consequence for its superseded target (PEOS-005
// §23.1: "Each Supersession relationship SHALL identify the resulting
// effective status or Lifecycle consequence, if any, for its superseded
// target within the relation scope. Absence of a Lifecycle consequence
// SHALL be explicitly representable and SHALL NOT invalidate an
// otherwise established Supersession relationship.").
//
// LifecycleConsequence is a closed two-state discriminator, exactly like
// Applicability's unrestricted/scoped distinction (see content.go's own
// doc comment) and decision.SupersessionExtent's complete/partial
// distinction: its zero value is invalid and represents a third,
// unstated state that §23.1 does not permit. NoLifecycleConsequence
// explicitly constructs the "no consequence" state as a distinct,
// non-zero value -- this is what makes "absence SHALL be explicitly
// representable" enforceable. A *string or (string, bool) encoding
// cannot distinguish absent-because-declared from absent-because-unset,
// which is why neither is used here.
//
// Description is prose, not a typed Lifecycle State reference: PEOS-005
// §26.5 governs the relationship between a Lifecycle State Assignment
// and a Supersession relationship, and this package deliberately does
// not depend on peos/lifecycle (see doc.go and supersession.go's own doc
// comment on Supersession). Embedding a state-shaped value here would
// blur exactly the boundary §26.5 requires to remain distinct.
type LifecycleConsequence struct {
	kind        lifecycleConsequenceKind
	description string
}

// NewLifecycleConsequence validates description and returns a
// LifecycleConsequence declaring an identified consequence. description
// must be non-empty after trimming surrounding whitespace; the trimmed
// value is stored.
func NewLifecycleConsequence(description string) (LifecycleConsequence, error) {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return LifecycleConsequence{}, fmt.Errorf("requirement: NewLifecycleConsequence: %w: description must not be empty", ErrInvalidRequirementSupersession)
	}
	return LifecycleConsequence{kind: lifecycleConsequenceIdentified, description: trimmed}, nil
}

// NoLifecycleConsequence returns a LifecycleConsequence explicitly
// declaring that no Lifecycle consequence results from the Supersession
// (PEOS-005 §23.1's own permitted "if any" / explicit-absence case).
func NoLifecycleConsequence() LifecycleConsequence {
	return LifecycleConsequence{kind: lifecycleConsequenceNone}
}

// Kind returns c's discriminator, "identified" or "none".
func (c LifecycleConsequence) Kind() string { return string(c.kind) }

// IsIdentified reports whether c declares an identified consequence.
func (c LifecycleConsequence) IsIdentified() bool {
	return c.kind == lifecycleConsequenceIdentified
}

// IsNone reports whether c explicitly declares no consequence.
func (c LifecycleConsequence) IsNone() bool { return c.kind == lifecycleConsequenceNone }

// Description returns c's declared consequence description, and whether
// one is set (that is, whether c is the identified variant).
func (c LifecycleConsequence) Description() (string, bool) {
	if c.kind != lifecycleConsequenceIdentified {
		return "", false
	}
	return c.description, true
}

// IsZero reports whether c is the zero value -- the unstated state
// PEOS-005 §23.1 does not permit on a valid Supersession.
func (c LifecycleConsequence) IsZero() bool { return c.kind == "" }

type lifecycleConsequenceJSON struct {
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

// MarshalJSON encodes c as {"kind":"identified","description":...} or
// {"kind":"none"}.
func (c LifecycleConsequence) MarshalJSON() ([]byte, error) {
	switch c.kind {
	case lifecycleConsequenceIdentified:
		return json.Marshal(lifecycleConsequenceJSON{Kind: string(lifecycleConsequenceIdentified), Description: c.description})
	case lifecycleConsequenceNone:
		return json.Marshal(lifecycleConsequenceJSON{Kind: string(lifecycleConsequenceNone)})
	default:
		return nil, fmt.Errorf("requirement: marshal LifecycleConsequence: %w", ErrInvalidRequirementSupersession)
	}
}

// UnmarshalJSON decodes c from its JSON form. An unrecognized or missing
// kind, an identified value missing (or carrying an empty/whitespace-only)
// description, or a none value carrying a description are all rejected.
func (c *LifecycleConsequence) UnmarshalJSON(data []byte) error {
	var raw lifecycleConsequenceJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal LifecycleConsequence: %w", err)
	}
	var result LifecycleConsequence
	switch raw.Kind {
	case string(lifecycleConsequenceIdentified):
		var err error
		result, err = NewLifecycleConsequence(raw.Description)
		if err != nil {
			return err
		}
	case string(lifecycleConsequenceNone):
		if raw.Description != "" {
			return fmt.Errorf("requirement: unmarshal LifecycleConsequence: %w: none must not carry a description", ErrInvalidRequirementSupersession)
		}
		result = NoLifecycleConsequence()
	default:
		return fmt.Errorf("requirement: unmarshal LifecycleConsequence: unrecognized kind %q: %w", raw.Kind, ErrInvalidRequirementSupersession)
	}
	*c = result
	return nil
}

// Supersession is a Requirement Supersession relationship (PEOS-005 §23:
// "A Requirement MAY supersede one or more Requirements through one or
// more explicitly represented Supersession relationships... Each
// Supersession relationship establishes that the required engineering
// intent represented by an identified superseding Requirement Artifact
// Revision replaces the required engineering intent represented by one
// identified superseded Requirement Artifact Revision within a defined
// scope.").
//
// Supersession composes relation.Relation rather than duplicating its
// fields (see relationship.go's own doc comment for the shared
// rationale): relation is the sole source of truth for relation type,
// source, target, provenance, scope, and extension. governanceAction and
// lifecycleConsequence are Supersession's own required content, both
// mandated by §23's and §23.1's own bullet lists.
//
// Both Superseding and Superseded are identified at Requirement Artifact
// Revision level (§23.1: "source and target SHALL be identified at
// Requirement Artifact Revision level while their owning Requirement
// identities remain identifiable"), never at Requirement identity level
// -- RequirementParticipant (used by Dependency and Conflict, whose
// participants may sit at either level) is deliberately not used here.
//
// Direction is inverted relative to every other relationship type this
// package implements: source is the superseding (newer) participant,
// target is the superseded (older) participant (§23.1). Derivation,
// Refinement, and Decomposition all put the originating participant at
// source; Supersession does not.
//
// Unlike Derivation and Decomposition, Supersession does NOT require its
// two participants to name distinct Requirement identities:
// REQ-1/REV-2 superseding REQ-1/REV-1 is a valid Supersession. Every
// PEOS-005 clause touching Supersession's identity effect is worded as
// preservation, not distinctness -- §7 ("Requirement supersession SHALL
// NOT merge Requirement identities"), §23 ("Superseded Requirements
// SHALL retain their identities"; "Supersession SHALL NOT merge
// Requirement identities"), §28.3 ("Requirement identity SHALL remain
// preserved"), and §35 ("Supersession SHALL NOT merge or destroy
// Requirement identities") -- in contrast to §18's "SHALL NOT inherit
// the identity of a source Requirement" and §20.1's "SHALL remain
// distinct," both of which state distinctness directly. "Merge" denotes
// collapsing two identities into one; a Supersession whose two
// participants already share one identity merges nothing. §23's
// replacement is scoped (§23.2), which linear revision history alone
// cannot express -- a later Revision narrowing its own earlier wording
// within a defined scope is exactly what this type is for.
//
// Supersession SHALL NOT be inferred solely from creation of a newer
// Artifact Revision, Statement similarity, document or identifier order,
// Lifecycle State, or archival status (§23, non-conforming pattern
// §36.13): a Supersession value cannot be constructed without an
// explicit governance action, an explicit scope, and explicit
// provenance, none of which a bare Artifact Revision carries.
//
// Supersession is deliberately independent of Lifecycle. §26.5: "A State
// Assignment SHALL NOT by itself establish: which Requirement supersedes
// another Requirement; which Requirement Artifact Revisions are
// involved; the scope of replacement; the authority or governance action
// establishing replacement... The State Assignment and the applicable
// Supersession relationship SHALL remain semantically distinct and
// independently inspectable." NewSupersession only ever creates an
// immutable relationship record: it does not read, validate, or require
// any Lifecycle State, and this package does not import peos/lifecycle.
// lifecycleConsequence is a declaration recorded inside the relationship,
// not a Lifecycle State Assignment and not a reference to one -- see
// non-conforming pattern §36.12 (treating a Superseded Lifecycle State
// Assignment as sufficient to establish the replacement fact).
//
// This package does not enforce transitive acyclicity (§23.1:
// "Supersession cycles SHALL NOT be permitted"; "A Requirement Artifact
// Revision SHALL NOT directly or transitively supersede itself"): only
// the direct case (source == target) is checked here. Detecting a
// transitive cycle requires traversing every other Supersession
// relationship in a repository, which this package does not hold -- see
// doc.go.
type Supersession struct {
	relation             relation.Relation
	governanceAction     GovernanceAction
	lifecycleConsequence LifecycleConsequence
}

// NewSupersession validates superseding, superseded, provenance, scope,
// governanceAction, and lifecycleConsequence and returns a Supersession.
// superseding and superseded must each be non-zero and identify a
// Requirement Artifact Revision; they must differ (direct
// self-supersession is rejected, but the two MAY share one Requirement
// identity -- see this type's own doc comment). provenance must be
// non-zero. scope must be non-zero. governanceAction must be non-zero.
// lifecycleConsequence must be explicitly stated -- its zero (unstated)
// value is rejected; use NoLifecycleConsequence to declare an explicit
// absence of consequence.
//
// A successful call always returns a fully valid Supersession: every
// mandatory field is a required constructor argument, never a later
// With* call, and the relation type is always
// core.RelationTypeRequirementSupersession -- never a caller input -- so
// a Supersession can never be constructed carrying any other relation
// type.
func NewSupersession(
	superseding core.RequirementArtifactRevisionRef,
	superseded core.RequirementArtifactRevisionRef,
	provenance core.Provenance,
	scope core.Scope,
	governanceAction GovernanceAction,
	lifecycleConsequence LifecycleConsequence,
) (Supersession, error) {
	supersedingSubject, err := requirementRevisionParticipant(superseding)
	if err != nil {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w", err)
	}
	supersededSubject, err := requirementRevisionParticipant(superseded)
	if err != nil {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w", err)
	}
	if err := checkDistinctParticipants(superseding, superseded); err != nil {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w", err)
	}
	if provenance.IsZero() {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w: provenance must not be zero", ErrInvalidRequirementRelation)
	}
	if governanceAction.IsZero() {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w: governance action must not be zero", ErrInvalidGovernanceAction)
	}
	if lifecycleConsequence.IsZero() {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w: lifecycle consequence must be explicitly stated", ErrInvalidRequirementSupersession)
	}
	if scope.IsZero() {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w: scope must not be zero", ErrInvalidRequirementRelation)
	}

	rel, err := relation.New(core.RelationTypeRequirementSupersession, supersedingSubject, supersededSubject, provenance)
	if err != nil {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w: %w", ErrInvalidRequirementRelation, err)
	}
	rel, err = rel.WithScope(scope)
	if err != nil {
		return Supersession{}, fmt.Errorf("requirement: NewSupersession: %w: %w", ErrInvalidRequirementRelation, err)
	}

	return Supersession{relation: rel, governanceAction: governanceAction, lifecycleConsequence: lifecycleConsequence}, nil
}

// WithExtension returns a copy of s with its extension data set.
func (s Supersession) WithExtension(extension core.Extension) Supersession {
	s.relation = s.relation.WithExtension(extension)
	return s
}

// WithoutExtension returns a copy of s with its extension data cleared.
func (s Supersession) WithoutExtension() Supersession {
	s.relation = s.relation.WithoutExtension()
	return s
}

// Superseding returns s's superseding Requirement Artifact Revision --
// the Revision whose required engineering intent replaces Superseded().
func (s Supersession) Superseding() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(s.relation.Source())
	return ref
}

// Superseded returns s's superseded Requirement Artifact Revision -- the
// Revision whose required engineering intent Superseding() replaces.
func (s Supersession) Superseded() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(s.relation.Target())
	return ref
}

// GovernanceAction returns s's declared governance action.
func (s Supersession) GovernanceAction() GovernanceAction { return s.governanceAction }

// LifecycleConsequence returns s's declared Lifecycle consequence.
func (s Supersession) LifecycleConsequence() LifecycleConsequence { return s.lifecycleConsequence }

// Provenance returns s's declared provenance.
func (s Supersession) Provenance() core.Provenance { return s.relation.Provenance() }

// Scope returns s's declared scope. Supersession's scope is mandatory
// and is therefore never absent on a valid Supersession.
func (s Supersession) Scope() core.Scope {
	scope, _ := s.relation.Scope()
	return scope
}

// Extension returns s's extension data.
func (s Supersession) Extension() core.Extension { return s.relation.Extension() }

// Relation returns s's underlying relation.Relation.
func (s Supersession) Relation() relation.Relation { return s.relation }

// IsZero reports whether s is the zero value.
func (s Supersession) IsZero() bool {
	return s.relation.IsZero() && s.governanceAction.IsZero() && s.lifecycleConsequence.IsZero()
}

type supersessionJSON struct {
	Relation             relation.Relation    `json:"relation"`
	GovernanceAction     GovernanceAction     `json:"governance_action"`
	LifecycleConsequence LifecycleConsequence `json:"lifecycle_consequence"`
}

// MarshalJSON encodes s as {"relation": {...}, "governance_action":
// {...}, "lifecycle_consequence": {...}}.
func (s Supersession) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Supersession: %w", ErrInvalidRequirementRelation)
	}
	return json.Marshal(supersessionJSON{
		Relation:             s.relation,
		GovernanceAction:     s.governanceAction,
		LifecycleConsequence: s.lifecycleConsequence,
	})
}

// UnmarshalJSON decodes s from its nested {"relation": {...},
// "governance_action": {...}, "lifecycle_consequence": {...}} JSON form.
// The decoded relation's own type is revalidated against
// core.RelationTypeRequirementSupersession, its source and target are
// revalidated as Requirement Artifact Revision-level participants, and
// its scope is revalidated as present, before the same validation
// NewSupersession applies is run again -- a decoded Supersession can
// never be constructor-impossible.
func (s *Supersession) UnmarshalJSON(data []byte) error {
	var raw supersessionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Supersession: %w: %w", ErrInvalidRequirementRelation, err)
	}

	if err := checkRelationType(raw.Relation, core.RelationTypeRequirementSupersession); err != nil {
		return fmt.Errorf("requirement: unmarshal Supersession: %w", err)
	}
	superseding, err := asRequirementRevisionParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Supersession: %w", err)
	}
	superseded, err := asRequirementRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Supersession: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Supersession: %w", err)
	}

	result, err := NewSupersession(superseding, superseded, raw.Relation.Provenance(), scope, raw.GovernanceAction, raw.LifecycleConsequence)
	if err != nil {
		return err
	}

	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*s = result
	return nil
}
