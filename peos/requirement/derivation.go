package requirement

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// Derivation is a Requirement Derivation relationship (PEOS-005 §18: "A
// Requirement Artifact Revision MAY be derived from one or more
// identified Requirement Artifact Revisions... Derivation records that
// required engineering intent was produced through engineering reasoning
// using other represented required engineering intent as an input.").
//
// Derivation composes relation.Relation rather than duplicating its
// fields (see relationship.go's own doc comment for the shared
// rationale): relation is the sole source of truth for relation type,
// source, target, provenance, scope, and extension. rationale is
// Derivation's own required content (§18: "Derivation SHALL preserve
// sufficient engineering rationale to explain how the derived required
// engineering intent originates from its identified sources.").
//
// Derivation carries no identity, no Ref type, no lifecycle, and no
// revision history (PEOS-005 §17.1): "A derived Requirement SHALL
// possess its own identity" (§18) refers to the derived Requirement
// itself, never to the Derivation relationship recording that fact.
//
// That same clause is also an invariant NewDerivation enforces: source
// and target MUST name different Requirement identities
// (source.ArtifactID() != target.ArtifactID()), not merely different
// Revisions. "A derived Requirement SHALL possess its own identity" and
// "A derived Requirement SHALL NOT inherit the identity of a source
// Requirement" (§18) forbid the source and derived Requirement from
// sharing one ArtifactID -- PEOS-009 (:664, :758) establishes that
// "inheriting" an identity means sharing it, the same condition PEOS-005
// §20.1 states in different words for Decomposition's parent and
// subordinate. See doc.go's "Requirement-identity distinctness" section
// for the full three-way comparison against Refinement, which imposes no
// such rule and therefore does permit two Revisions of the very same
// Requirement to serve as Refined and Refining. A later Revision of the
// same Requirement narrowing its own earlier wording is ordinary content
// change under §25, not a Derivation.
//
// Both source and target are identified at Requirement Artifact Revision
// level, never at Requirement identity level (§18.1: "both source and
// target SHALL be identified at Requirement Artifact Revision level").
// Derivation does not imply Refinement, Decomposition, Supersession,
// implementation, Allocation, or Satisfaction (§18), and is a distinct
// relation type from Refinement even though a Refinement relationship
// MAY also be supported by Derivation (§19: "neither relationship SHALL
// be inferred solely from the existence of the other" -- non-conforming
// pattern §36.15).
//
// This package does not enforce transitive acyclicity (§18.1: "A
// Requirement Artifact Revision SHALL NOT be directly or transitively
// derived from itself"): only the direct cases (source == target, and
// source's Requirement identity == target's Requirement identity) are
// checked here. Detecting a transitive cycle requires traversing every
// other Derivation relationship in a repository, which this package does
// not hold -- see doc.go.
type Derivation struct {
	relation  relation.Relation
	rationale string
}

// NewDerivation validates source, target, provenance, and rationale and
// returns a Derivation. source and target must each be non-zero and
// identify a Requirement Artifact Revision; they must differ, and their
// owning Requirement identities must also differ (PEOS-005 §18: a
// derived Requirement must possess its own identity and must not inherit
// the identity of a source Requirement). provenance must be non-zero.
// rationale must be non-empty after trimming surrounding whitespace; the
// trimmed value is stored.
//
// A successful call always returns a fully valid Derivation: every
// mandatory field is a required constructor argument, never a later
// With* call, and the relation type is always core.RelationTypeDerivation
// -- never a caller input -- so a Derivation can never be constructed
// carrying any other relation type.
func NewDerivation(
	source core.RequirementArtifactRevisionRef,
	target core.RequirementArtifactRevisionRef,
	provenance core.Provenance,
	rationale string,
) (Derivation, error) {
	sourceSubject, err := requirementRevisionParticipant(source)
	if err != nil {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w", err)
	}
	targetSubject, err := requirementRevisionParticipant(target)
	if err != nil {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w", err)
	}
	if err := checkDistinctParticipants(source, target); err != nil {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w", err)
	}
	if err := checkDistinctRequirementIdentity(source, target, ErrInvalidDerivation); err != nil {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w", err)
	}
	if provenance.IsZero() {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w: provenance must not be zero", ErrInvalidRequirementRelation)
	}
	trimmed := strings.TrimSpace(rationale)
	if trimmed == "" {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w: rationale must not be empty", ErrInvalidDerivation)
	}

	rel, err := relation.New(core.RelationTypeDerivation, sourceSubject, targetSubject, provenance)
	if err != nil {
		return Derivation{}, fmt.Errorf("requirement: NewDerivation: %w: %w", ErrInvalidRequirementRelation, err)
	}

	return Derivation{relation: rel, rationale: trimmed}, nil
}

// WithScope returns a copy of d with its declared scope set. scope must
// be non-zero. Use WithoutScope to clear a previously set scope.
func (d Derivation) WithScope(scope core.Scope) (Derivation, error) {
	rel, err := d.relation.WithScope(scope)
	if err != nil {
		return Derivation{}, fmt.Errorf("requirement: Derivation.WithScope: %w", err)
	}
	d.relation = rel
	return d, nil
}

// WithoutScope returns a copy of d with its declared scope cleared.
func (d Derivation) WithoutScope() Derivation {
	d.relation = d.relation.WithoutScope()
	return d
}

// WithExtension returns a copy of d with its extension data set.
func (d Derivation) WithExtension(extension core.Extension) Derivation {
	d.relation = d.relation.WithExtension(extension)
	return d
}

// WithoutExtension returns a copy of d with its extension data cleared.
func (d Derivation) WithoutExtension() Derivation {
	d.relation = d.relation.WithoutExtension()
	return d
}

// Source returns d's source Requirement Artifact Revision -- the
// Requirement Artifact Revision the derived intent originates from.
func (d Derivation) Source() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(d.relation.Source())
	return ref
}

// Target returns d's target Requirement Artifact Revision -- the derived
// Requirement Artifact Revision.
func (d Derivation) Target() core.RequirementArtifactRevisionRef {
	ref, _ := asRequirementRevisionParticipant(d.relation.Target())
	return ref
}

// Rationale returns d's declared engineering rationale.
func (d Derivation) Rationale() string { return d.rationale }

// Provenance returns d's declared provenance.
func (d Derivation) Provenance() core.Provenance { return d.relation.Provenance() }

// Scope returns d's declared scope, and whether one is set.
func (d Derivation) Scope() (core.Scope, bool) { return d.relation.Scope() }

// Extension returns d's extension data.
func (d Derivation) Extension() core.Extension { return d.relation.Extension() }

// Relation returns d's underlying relation.Relation.
func (d Derivation) Relation() relation.Relation { return d.relation }

// IsZero reports whether d is the zero value.
func (d Derivation) IsZero() bool { return d.relation.IsZero() && d.rationale == "" }

type derivationJSON struct {
	Relation  relation.Relation `json:"relation"`
	Rationale string            `json:"rationale"`
}

// MarshalJSON encodes d as {"relation": {...}, "rationale": ...}.
func (d Derivation) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Derivation: %w", ErrInvalidRequirementRelation)
	}
	return json.Marshal(derivationJSON{Relation: d.relation, Rationale: d.rationale})
}

// UnmarshalJSON decodes d from its nested {"relation": {...}, "rationale":
// ...} JSON form. The decoded relation's own type is revalidated against
// core.RelationTypeDerivation, and both its source and target are
// revalidated as Requirement Artifact Revision-level participants,
// before the same validation NewDerivation applies is run again -- a
// decoded Derivation can never be constructor-impossible.
func (d *Derivation) UnmarshalJSON(data []byte) error {
	var raw derivationJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Derivation: %w: %w", ErrInvalidRequirementRelation, err)
	}

	if err := checkRelationType(raw.Relation, core.RelationTypeDerivation); err != nil {
		return fmt.Errorf("requirement: unmarshal Derivation: %w", err)
	}
	source, err := asRequirementRevisionParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Derivation: %w", err)
	}
	target, err := asRequirementRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Derivation: %w", err)
	}

	result, err := NewDerivation(source, target, raw.Relation.Provenance(), raw.Rationale)
	if err != nil {
		return err
	}

	if scope, ok := raw.Relation.Scope(); ok {
		if result, err = result.WithScope(scope); err != nil {
			return err
		}
	}
	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*d = result
	return nil
}
