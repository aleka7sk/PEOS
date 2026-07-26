package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// Refinement is a Requirement Refinement relationship (PEOS-005 §19: "A
// Requirement Artifact Revision MAY refine one or more identified
// Requirement Artifact Revisions... Refinement records that required
// engineering intent is expressed with increased precision, narrower
// interpretation, additional constraint, or greater engineering detail
// while remaining compatible with the refined required engineering
// intent within the defined scope.").
//
// Refinement composes relation.Relation rather than duplicating its
// fields (see relationship.go's own doc comment for the shared
// rationale): relation is the sole source of truth for relation type,
// source, target, provenance, scope, and extension. Unlike Derivation,
// Refinement adds no field of its own -- its entire type-specific
// content (the scope within which compatibility is asserted, §19) is
// already expressible as relation.Relation's own scope field.
//
// Scope is mandatory for Refinement, not optional as for Derivation:
// §19 requires every Refinement relationship to identify "the scope
// within which compatibility is asserted." Because scope can never be
// absent on a valid Refinement, this type exposes no WithScope or
// WithoutScope -- WithoutScope would make an invalid, scope-less value
// reachable, and WithScope is unnecessary because NewRefinement already
// requires scope as an argument. Scope() therefore returns a bare
// core.Scope, not the (core.Scope, bool) shape Derivation's optional
// scope requires.
//
// Refinement carries no identity, no Ref type, no lifecycle, and no
// revision history (PEOS-005 §17.1). Refined and Refining are
// identified at Requirement Artifact Revision level, never at
// Requirement identity level (§19.1: "both source and target SHALL be
// identified at Requirement Artifact Revision level"), with direction
// from refined to refining (§19.1).
//
// Refinement does not enforce Requirement-identity distinctness between
// Refined and Refining, unlike Decomposition's parent/subordinate pair
// (see Decomposition's own doc comment for that rule). §19's own
// language is weaker: "A refining Requirement SHALL remain independently
// identifiable" (line 739) requires only that the refining Requirement
// have its own identity -- not that it differ from the refined
// Requirement's identity -- and §19.1 states no distinctness rule at
// all, unlike §20.1's explicit "A subordinate Requirement identity SHALL
// remain distinct from the parent Requirement identity." A later
// Revision of the same Requirement MAY therefore validly refine an
// earlier Revision of itself.
//
// Refinement does not imply Derivation, Decomposition, Supersession,
// implementation, Allocation, or Satisfaction (§19). A Refinement
// relationship and a Derivation relationship MAY both exist over the
// same participant pair without either implying the other (§19:
// "Refinement MAY also be supported by Derivation, but neither
// relationship SHALL be inferred solely from the existence of the
// other" -- non-conforming pattern §36.15).
//
// This package does not enforce transitive acyclicity (§19.1: "Refinement
// cycles SHALL NOT be permitted"; "A Requirement Artifact Revision SHALL
// NOT directly or transitively refine itself"): only the direct case
// (refined == refining) is checked here. Detecting a transitive cycle
// requires traversing every other Refinement relationship in a
// repository, which this package does not hold -- see doc.go. Nor does
// this package evaluate whether the refining intent is actually
// compatible with the refined intent within the declared scope (§19,
// §19.1): that is an engineering judgment about requirement text, made
// by the applicable Product contract or its authors, not a structural
// property this package can check.
type Refinement struct {
	relation relation.Relation
}

// NewRefinement validates refined, refining, provenance, and scope and
// returns a Refinement. refined and refining must each be non-zero and
// identify a Requirement Artifact Revision; they must differ. provenance
// must be non-zero. scope must be non-zero.
//
// A successful call always returns a fully valid Refinement: every
// mandatory field is a required constructor argument, never a later
// With* call, and the relation type is always core.RelationTypeRefinement
// -- never a caller input -- so a Refinement can never be constructed
// carrying any other relation type.
func NewRefinement(
	refined core.RequirementArtifactRevisionRef,
	refining core.RequirementArtifactRevisionRef,
	provenance core.Provenance,
	scope core.Scope,
) (Refinement, error) {
	refinedSubject, err := requirementRevisionParticipant(refined)
	if err != nil {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w", err)
	}
	refiningSubject, err := requirementRevisionParticipant(refining)
	if err != nil {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w", err)
	}
	if err := checkDistinctParticipants(refined, refining); err != nil {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w", err)
	}
	if provenance.IsZero() {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w: provenance must not be zero", ErrInvalidRequirementRelation)
	}
	if scope.IsZero() {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w: scope must not be zero", ErrInvalidRequirementRelation)
	}

	rel, err := relation.New(core.RelationTypeRefinement, refinedSubject, refiningSubject, provenance)
	if err != nil {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w: %w", ErrInvalidRequirementRelation, err)
	}
	rel, err = rel.WithScope(scope)
	if err != nil {
		return Refinement{}, fmt.Errorf("requirement: NewRefinement: %w: %w", ErrInvalidRequirementRelation, err)
	}

	return Refinement{relation: rel}, nil
}

// WithExtension returns a copy of r with its extension data set.
func (r Refinement) WithExtension(extension core.Extension) Refinement {
	r.relation = r.relation.WithExtension(extension)
	return r
}

// WithoutExtension returns a copy of r with its extension data cleared.
func (r Refinement) WithoutExtension() Refinement {
	r.relation = r.relation.WithoutExtension()
	return r
}

// Refined returns r's refined Requirement Artifact Revision -- the
// Requirement Artifact Revision being refined.
func (r Refinement) Refined() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(r.relation.Source())
	return ref
}

// Refining returns r's refining Requirement Artifact Revision -- the
// Requirement Artifact Revision that refines Refined().
func (r Refinement) Refining() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(r.relation.Target())
	return ref
}

// Provenance returns r's declared provenance.
func (r Refinement) Provenance() core.Provenance { return r.relation.Provenance() }

// Scope returns r's declared scope. Unlike Derivation's optional scope,
// Refinement's scope is mandatory and is therefore never absent on a
// valid Refinement.
func (r Refinement) Scope() core.Scope {
	scope, _ := r.relation.Scope()
	return scope
}

// Extension returns r's extension data.
func (r Refinement) Extension() core.Extension { return r.relation.Extension() }

// Relation returns r's underlying relation.Relation.
func (r Refinement) Relation() relation.Relation { return r.relation }

// IsZero reports whether r is the zero value.
func (r Refinement) IsZero() bool { return r.relation.IsZero() }

type refinementJSON struct {
	Relation relation.Relation `json:"relation"`
}

// MarshalJSON encodes r as {"relation": {...}}.
func (r Refinement) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Refinement: %w", ErrInvalidRequirementRelation)
	}
	return json.Marshal(refinementJSON{Relation: r.relation})
}

// UnmarshalJSON decodes r from its nested {"relation": {...}} JSON form.
// The decoded relation's own type is revalidated against
// core.RelationTypeRefinement, its source and target are revalidated as
// Requirement Artifact Revision-level participants, and its scope is
// revalidated as present, before the same validation NewRefinement
// applies is run again -- a decoded Refinement can never be
// constructor-impossible.
func (r *Refinement) UnmarshalJSON(data []byte) error {
	var raw refinementJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Refinement: %w: %w", ErrInvalidRequirementRelation, err)
	}

	if err := checkRelationType(raw.Relation, core.RelationTypeRefinement); err != nil {
		return fmt.Errorf("requirement: unmarshal Refinement: %w", err)
	}
	refined, err := asRequirementRevisionParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Refinement: %w", err)
	}
	refining, err := asRequirementRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Refinement: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Refinement: %w", err)
	}

	result, err := NewRefinement(refined, refining, raw.Relation.Provenance(), scope)
	if err != nil {
		return err
	}

	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*r = result
	return nil
}
