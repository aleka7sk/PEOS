package requirement

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// Dependency is a Requirement Dependency relationship (PEOS-005 §21: "A
// Requirement or Requirement Artifact Revision MAY depend upon one or
// more Requirements or Requirement Artifact Revisions... Dependency
// records that satisfying the dependent required engineering intent
// relies on the availability, satisfaction, or continued validity of the
// depended-upon required engineering intent.").
//
// Dependency composes relation.Relation rather than duplicating its
// fields (see relationship.go's own doc comment for the shared
// rationale): relation is the sole source of truth for relation type,
// source, target, provenance, scope, and extension. nature is
// Dependency's own required content (§21: the nature of the engineering
// reliance).
//
// Unlike Derivation, Refinement, and Decomposition, each of Dependent and
// DependsOn MAY independently identify either a Requirement identity or a
// specific Requirement Artifact Revision (§21.1: "each participant SHALL
// explicitly identify whether it refers to Requirement identity level or
// Requirement Artifact Revision level") -- RequirementParticipant is the
// closed union expressing that choice (see participant.go). Direction is
// semantic, not merely representational: dependent depends on dependsOn
// (§21.1).
//
// Scope is optional for Dependency, as for Derivation: §21 states scope
// only conditionally, "the applicable scope where the reliance is not
// universal." WithScope and WithoutScope exist for exactly that reason,
// and Scope() returns the (core.Scope, bool) shape Derivation's optional
// scope also uses.
//
// Dependency carries no identity, no Ref type, no lifecycle, and no
// revision history (PEOS-005 §17.1).
//
// Dependency enforces no distinctness rule of any kind between its two
// participants -- not participant equality, and not Requirement-identity
// equality. §21 and §21.1 state no such rule, and §21.1 explicitly
// permits Dependency cycles: "Dependency cycles MAY be represented. The
// existence of a Dependency cycle SHALL NOT by itself establish that the
// participating Requirements are invalid, unsatisfiable, or
// non-conforming." A self-dependency is the degenerate one-node cycle and
// is therefore a valid Dependency, not an error -- the deliberate opposite
// of Derivation's, Refinement's, and Decomposition's self-reference
// rejection. A Product contract MAY additionally prohibit specific
// Dependency cycles (§21.1's own final sentence); this package does not.
type Dependency struct {
	relation relation.Relation
	nature   string
}

// NewDependency validates dependent, dependsOn, provenance, and nature
// and returns a Dependency. dependent and dependsOn must each be
// non-zero; no distinctness rule applies to them (see this type's own
// doc comment). provenance must be non-zero. nature must be non-empty
// after trimming surrounding whitespace; the trimmed value is stored.
//
// A successful call always returns a fully valid Dependency: every
// mandatory field is a required constructor argument, never a later
// With* call, and the relation type is always core.RelationTypeDependency
// -- never a caller input -- so a Dependency can never be constructed
// carrying any other relation type.
func NewDependency(
	dependent RequirementParticipant,
	dependsOn RequirementParticipant,
	provenance core.Provenance,
	nature string,
) (Dependency, error) {
	dependentSubject, err := requirementParticipantSubject(dependent)
	if err != nil {
		return Dependency{}, fmt.Errorf("requirement: NewDependency: %w", err)
	}
	dependsOnSubject, err := requirementParticipantSubject(dependsOn)
	if err != nil {
		return Dependency{}, fmt.Errorf("requirement: NewDependency: %w", err)
	}
	if provenance.IsZero() {
		return Dependency{}, fmt.Errorf("requirement: NewDependency: %w: provenance must not be zero", ErrInvalidRequirementRelation)
	}
	trimmed := strings.TrimSpace(nature)
	if trimmed == "" {
		return Dependency{}, fmt.Errorf("requirement: NewDependency: %w: nature must not be empty", ErrInvalidDependency)
	}

	rel, err := relation.New(core.RelationTypeDependency, dependentSubject, dependsOnSubject, provenance)
	if err != nil {
		return Dependency{}, fmt.Errorf("requirement: NewDependency: %w: %w", ErrInvalidRequirementRelation, err)
	}

	return Dependency{relation: rel, nature: trimmed}, nil
}

// WithScope returns a copy of d with its declared scope set. scope must
// be non-zero. Use WithoutScope to clear a previously set scope.
func (d Dependency) WithScope(scope core.Scope) (Dependency, error) {
	rel, err := d.relation.WithScope(scope)
	if err != nil {
		return Dependency{}, fmt.Errorf("requirement: Dependency.WithScope: %w", err)
	}
	d.relation = rel
	return d, nil
}

// WithoutScope returns a copy of d with its declared scope cleared.
func (d Dependency) WithoutScope() Dependency {
	d.relation = d.relation.WithoutScope()
	return d
}

// WithExtension returns a copy of d with its extension data set.
func (d Dependency) WithExtension(extension core.Extension) Dependency {
	d.relation = d.relation.WithExtension(extension)
	return d
}

// WithoutExtension returns a copy of d with its extension data cleared.
func (d Dependency) WithoutExtension() Dependency {
	d.relation = d.relation.WithoutExtension()
	return d
}

// Dependent returns d's dependent participant -- the Requirement or
// Requirement Artifact Revision whose satisfaction relies on DependsOn().
func (d Dependency) Dependent() RequirementParticipant {
	p, _ := asRequirementParticipant(d.relation.Source())
	return p
}

// DependsOn returns d's depended-upon participant.
func (d Dependency) DependsOn() RequirementParticipant {
	p, _ := asRequirementParticipant(d.relation.Target())
	return p
}

// Nature returns d's declared nature of engineering reliance.
func (d Dependency) Nature() string { return d.nature }

// Provenance returns d's declared provenance.
func (d Dependency) Provenance() core.Provenance { return d.relation.Provenance() }

// Scope returns d's declared scope, and whether one is set.
func (d Dependency) Scope() (core.Scope, bool) { return d.relation.Scope() }

// Extension returns d's extension data.
func (d Dependency) Extension() core.Extension { return d.relation.Extension() }

// Relation returns d's underlying relation.Relation.
func (d Dependency) Relation() relation.Relation { return d.relation }

// IsZero reports whether d is the zero value.
func (d Dependency) IsZero() bool { return d.relation.IsZero() && d.nature == "" }

type dependencyJSON struct {
	Relation relation.Relation `json:"relation"`
	Nature   string            `json:"nature"`
}

// MarshalJSON encodes d as {"relation": {...}, "nature": ...}.
func (d Dependency) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Dependency: %w", ErrInvalidRequirementRelation)
	}
	return json.Marshal(dependencyJSON{Relation: d.relation, Nature: d.nature})
}

// UnmarshalJSON decodes d from its nested {"relation": {...}, "nature":
// ...} JSON form. The decoded relation's own type is revalidated against
// core.RelationTypeDependency, and both its source and target are
// revalidated as Requirement participants at either permitted level,
// before the same validation NewDependency applies is run again -- a
// decoded Dependency can never be constructor-impossible.
func (d *Dependency) UnmarshalJSON(data []byte) error {
	var raw dependencyJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Dependency: %w: %w", ErrInvalidRequirementRelation, err)
	}

	if err := checkRelationType(raw.Relation, core.RelationTypeDependency); err != nil {
		return fmt.Errorf("requirement: unmarshal Dependency: %w", err)
	}
	dependent, err := asRequirementParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Dependency: %w", err)
	}
	dependsOn, err := asRequirementParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Dependency: %w", err)
	}

	result, err := NewDependency(dependent, dependsOn, raw.Relation.Provenance(), raw.Nature)
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
