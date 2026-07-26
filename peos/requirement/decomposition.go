package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// Decomposition is a Requirement Decomposition relationship (PEOS-005
// §20: "A Requirement Artifact Revision MAY be decomposed into multiple
// subordinate Requirement Artifact Revisions... Decomposition records
// that required engineering intent represented by an identified parent
// Requirement Artifact Revision is partitioned into multiple
// independently identifiable subordinate Requirements.").
//
// Decomposition composes relation.Relation rather than duplicating its
// fields (see relationship.go's own doc comment for the shared
// rationale): relation is the sole source of truth for relation type,
// source, target, provenance, scope, and extension. Like Refinement,
// Decomposition adds no field of its own -- its entire type-specific
// content (the scope of the parent-to-subordinate association, §20) is
// already expressible as relation.Relation's own scope field.
//
// Scope is mandatory for Decomposition, not optional as for Derivation:
// §20 requires every Decomposition relationship to identify "the scope
// of that parent-to-subordinate association." As with Refinement, this
// type exposes no WithScope or WithoutScope, and Scope() returns a bare
// core.Scope rather than Derivation's (core.Scope, bool) shape.
//
// Decomposition carries no identity, no Ref type, no lifecycle, and no
// revision history (PEOS-005 §17.1). Parent and Subordinate are
// identified at Requirement Artifact Revision level, never at
// Requirement identity level (§20.1: "parent and subordinate
// participants SHALL be identified at Requirement Artifact Revision
// level"), with direction from parent to subordinate (§20.1).
//
// Unlike Refinement, Decomposition enforces Requirement-identity
// distinctness between its two participants: §20.1 states explicitly "A
// subordinate Requirement identity SHALL remain distinct from the parent
// Requirement identity." This is a stronger rule than the direct
// revision-level self-reference check every relationship type in this
// package applies -- it additionally rejects a parent and subordinate
// naming two different Revisions of the very same Requirement, which
// Refinement deliberately permits (see Refinement's own doc comment for
// why §19 does not impose the same rule for refined/refining).
//
// One parent Requirement Artifact Revision MAY be the source of multiple
// Decomposition relationships (§20.1), and one subordinate Requirement
// Artifact Revision MAY be the target of more than one Decomposition
// relationship (§20.1: "only where each parent relationship and its
// scope remain explicitly distinguishable" -- a cross-relationship
// condition this package cannot check locally; see doc.go). Neither
// case is rejected by this type: both are valid PEOS-005 shapes, and
// nothing in a single Decomposition value's own fields can determine
// whether a distinguishability violation exists among other
// Decomposition relationships this package does not hold.
//
// Decomposition does not by itself establish that its subordinate
// Requirements completely cover the parent's required engineering
// intent, that they are mutually exclusive, or that satisfying every
// subordinate establishes satisfaction of the parent (§20.2). PEOS-005
// §20.2 explicitly defines no structural model for decomposition
// completeness and explicitly forbids introducing a Relationship
// Collection, a Decomposition Set, a Completeness Assertion, or any
// other PEOS entity representing a group of Decomposition relationships
// -- this package introduces none of those, and a set of Decomposition
// values sharing one parent is, deliberately, nothing more than that:
// a set of independently valid Go values with no group type of its own.
//
// This package does not enforce transitive acyclicity (§20.1:
// "Decomposition cycles SHALL NOT be permitted"; "A Requirement Artifact
// Revision SHALL NOT directly or transitively decompose itself"): only
// the direct case (parent == subordinate, and parent's Requirement
// identity == subordinate's Requirement identity) is checked here.
// Detecting a transitive cycle requires traversing every other
// Decomposition relationship in a repository, which this package does
// not hold -- see doc.go.
type Decomposition struct {
	relation relation.Relation
}

// NewDecomposition validates parent, subordinate, provenance, and scope
// and returns a Decomposition. parent and subordinate must each be
// non-zero and identify a Requirement Artifact Revision; they must
// differ, and their owning Requirement identities must also differ
// (PEOS-005 §20.1). provenance must be non-zero. scope must be non-zero.
//
// A successful call always returns a fully valid Decomposition: every
// mandatory field is a required constructor argument, never a later
// With* call, and the relation type is always
// core.RelationTypeDecomposition -- never a caller input -- so a
// Decomposition can never be constructed carrying any other relation
// type.
func NewDecomposition(
	parent core.RequirementArtifactRevisionRef,
	subordinate core.RequirementArtifactRevisionRef,
	provenance core.Provenance,
	scope core.Scope,
) (Decomposition, error) {
	parentSubject, err := requirementRevisionParticipant(parent)
	if err != nil {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w", err)
	}
	subordinateSubject, err := requirementRevisionParticipant(subordinate)
	if err != nil {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w", err)
	}
	if err := checkDistinctParticipants(parent, subordinate); err != nil {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w", err)
	}
	if err := checkDistinctRequirementIdentity(parent, subordinate, ErrInvalidDecomposition); err != nil {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w", err)
	}
	if provenance.IsZero() {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w: provenance must not be zero", ErrInvalidRequirementRelation)
	}
	if scope.IsZero() {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w: scope must not be zero", ErrInvalidRequirementRelation)
	}

	rel, err := relation.New(core.RelationTypeDecomposition, parentSubject, subordinateSubject, provenance)
	if err != nil {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w: %w", ErrInvalidRequirementRelation, err)
	}
	rel, err = rel.WithScope(scope)
	if err != nil {
		return Decomposition{}, fmt.Errorf("requirement: NewDecomposition: %w: %w", ErrInvalidRequirementRelation, err)
	}

	return Decomposition{relation: rel}, nil
}

// WithExtension returns a copy of d with its extension data set.
func (d Decomposition) WithExtension(extension core.Extension) Decomposition {
	d.relation = d.relation.WithExtension(extension)
	return d
}

// WithoutExtension returns a copy of d with its extension data cleared.
func (d Decomposition) WithoutExtension() Decomposition {
	d.relation = d.relation.WithoutExtension()
	return d
}

// Parent returns d's parent Requirement Artifact Revision -- the
// Requirement Artifact Revision being decomposed.
func (d Decomposition) Parent() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(d.relation.Source())
	return ref
}

// Subordinate returns d's subordinate Requirement Artifact Revision --
// one of the Requirement Artifact Revisions Parent() is decomposed into.
func (d Decomposition) Subordinate() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(d.relation.Target())
	return ref
}

// Provenance returns d's declared provenance.
func (d Decomposition) Provenance() core.Provenance { return d.relation.Provenance() }

// Scope returns d's declared scope. Unlike Derivation's optional scope,
// Decomposition's scope is mandatory and is therefore never absent on a
// valid Decomposition.
func (d Decomposition) Scope() core.Scope {
	scope, _ := d.relation.Scope()
	return scope
}

// Extension returns d's extension data.
func (d Decomposition) Extension() core.Extension { return d.relation.Extension() }

// Relation returns d's underlying relation.Relation.
func (d Decomposition) Relation() relation.Relation { return d.relation }

// IsZero reports whether d is the zero value.
func (d Decomposition) IsZero() bool { return d.relation.IsZero() }

type decompositionJSON struct {
	Relation relation.Relation `json:"relation"`
}

// MarshalJSON encodes d as {"relation": {...}}.
func (d Decomposition) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Decomposition: %w", ErrInvalidRequirementRelation)
	}
	return json.Marshal(decompositionJSON{Relation: d.relation})
}

// UnmarshalJSON decodes d from its nested {"relation": {...}} JSON form.
// The decoded relation's own type is revalidated against
// core.RelationTypeDecomposition, its source and target are revalidated
// as Requirement Artifact Revision-level participants, and its scope is
// revalidated as present, before the same validation NewDecomposition
// applies (including the parent/subordinate identity-distinctness check)
// is run again -- a decoded Decomposition can never be
// constructor-impossible.
func (d *Decomposition) UnmarshalJSON(data []byte) error {
	var raw decompositionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Decomposition: %w: %w", ErrInvalidRequirementRelation, err)
	}

	if err := checkRelationType(raw.Relation, core.RelationTypeDecomposition); err != nil {
		return fmt.Errorf("requirement: unmarshal Decomposition: %w", err)
	}
	parent, err := asRequirementRevisionParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Decomposition: %w", err)
	}
	subordinate, err := asRequirementRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Decomposition: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Decomposition: %w", err)
	}

	result, err := NewDecomposition(parent, subordinate, raw.Relation.Provenance(), scope)
	if err != nil {
		return err
	}

	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*d = result
	return nil
}
