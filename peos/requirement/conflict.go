package requirement

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// Conflict is a Requirement Conflict relationship (PEOS-005 §22: "Two
// Requirements or Requirement Artifact Revisions MAY be identified as
// being in Conflict when they establish incompatible required
// engineering intent within overlapping applicability... Conflict records
// that satisfying both participants simultaneously, within the identified
// scope, is not possible.").
//
// Conflict composes relation.Relation rather than duplicating its fields
// (see relationship.go's own doc comment for the shared rationale):
// relation is the sole source of truth for relation type, source, target,
// provenance, scope, and extension. nature is Conflict's own required
// content (§22: the nature of the incompatibility).
//
// As with Dependency, each of ParticipantA and ParticipantB MAY
// independently identify either a Requirement identity or a specific
// Requirement Artifact Revision (§22.1, identical wording to §21.1).
//
// Conflict is symmetric in meaning but representationally ordered: §22.1
// states an implementation's participant ordering "SHALL NOT imply
// priority, authority, preference, or resolution direction." ParticipantA
// and ParticipantB deliberately avoid the Source/Target naming every other
// relationship type in this package uses, matching
// decision.DecisionConflict's DecisionA/DecisionB precedent for the same
// reason. This package does not canonicalize participant order: a
// Conflict constructed as (X, Y) and one constructed as (Y, X) are both
// valid and are preserved exactly as supplied, including through JSON
// round-trips. Recognizing that two stored Conflicts denote the same
// unordered pair is a repository-level concern this package does not
// perform.
//
// Scope is mandatory for Conflict, as for Refinement and Decomposition:
// §22 requires every Conflict relationship to identify "the scope in
// which the Conflict exists." This type exposes no WithScope or
// WithoutScope, and Scope() returns a bare core.Scope rather than
// Dependency's (core.Scope, bool) shape.
//
// Conflict carries no identity, no Ref type, no lifecycle, and no
// revision history (PEOS-005 §17.1).
//
// Unlike Dependency, Conflict requires its two participants to be
// distinct (§22: "exactly two distinct participants"; §22.1: "source and
// target SHALL be distinct") -- a self-conflict is rejected. This
// distinctness check compares the two RequirementParticipant values
// directly (participant shape, not Requirement identity), matching
// checkDistinctParticipants's own participant-shape rule for the
// revision-only relationship types; no Requirement-identity distinctness
// rule applies here (a Requirement MAY conflict with a distinct Revision
// of the very same Requirement -- for example, an amended Revision
// conflicting with an as-yet-unwithdrawn earlier one). §22.1 explicitly
// permits Conflict cycles among three or more distinct participants, and
// one participant MAY participate in multiple Conflict relationships
// (§22.1) -- neither is checked by this package; see doc.go.
type Conflict struct {
	relation relation.Relation
	nature   string
}

// NewConflict validates participantA, participantB, provenance, scope,
// and nature and returns a Conflict. participantA and participantB must
// each be non-zero and must differ from one another. provenance must be
// non-zero. scope must be non-zero. nature must be non-empty after
// trimming surrounding whitespace; the trimmed value is stored.
//
// A successful call always returns a fully valid Conflict: every
// mandatory field is a required constructor argument, never a later
// With* call, and the relation type is always core.RelationTypeConflict
// -- never a caller input -- so a Conflict can never be constructed
// carrying any other relation type.
func NewConflict(
	participantA RequirementParticipant,
	participantB RequirementParticipant,
	provenance core.Provenance,
	scope core.Scope,
	nature string,
) (Conflict, error) {
	subjectA, err := requirementParticipantSubject(participantA)
	if err != nil {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w", err)
	}
	subjectB, err := requirementParticipantSubject(participantB)
	if err != nil {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w", err)
	}
	if participantA == participantB {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w: the two participants must not be the same", ErrInvalidRequirementRelation)
	}
	if provenance.IsZero() {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w: provenance must not be zero", ErrInvalidRequirementRelation)
	}
	trimmed := strings.TrimSpace(nature)
	if trimmed == "" {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w: nature must not be empty", ErrInvalidConflict)
	}
	if scope.IsZero() {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w: scope must not be zero", ErrInvalidRequirementRelation)
	}

	rel, err := relation.New(core.RelationTypeConflict, subjectA, subjectB, provenance)
	if err != nil {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w: %w", ErrInvalidRequirementRelation, err)
	}
	rel, err = rel.WithScope(scope)
	if err != nil {
		return Conflict{}, fmt.Errorf("requirement: NewConflict: %w: %w", ErrInvalidRequirementRelation, err)
	}

	return Conflict{relation: rel, nature: trimmed}, nil
}

// WithExtension returns a copy of c with its extension data set.
func (c Conflict) WithExtension(extension core.Extension) Conflict {
	c.relation = c.relation.WithExtension(extension)
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c Conflict) WithoutExtension() Conflict {
	c.relation = c.relation.WithoutExtension()
	return c
}

// ParticipantA returns one of c's two conflicting participants.
// ParticipantA and ParticipantB are an unordered pair as a matter of
// meaning; the SDK preserves the caller's supplied ordering exactly
// rather than canonicalizing it (see this type's own doc comment).
func (c Conflict) ParticipantA() RequirementParticipant {
	p, _ := asRequirementParticipant(c.relation.Source())
	return p
}

// ParticipantB returns the other of c's two conflicting participants.
func (c Conflict) ParticipantB() RequirementParticipant {
	p, _ := asRequirementParticipant(c.relation.Target())
	return p
}

// Nature returns c's declared nature of incompatibility.
func (c Conflict) Nature() string { return c.nature }

// Provenance returns c's declared provenance.
func (c Conflict) Provenance() core.Provenance { return c.relation.Provenance() }

// Scope returns c's declared scope. Unlike Dependency's optional scope,
// Conflict's scope is mandatory and is therefore never absent on a valid
// Conflict.
func (c Conflict) Scope() core.Scope {
	scope, _ := c.relation.Scope()
	return scope
}

// Extension returns c's extension data.
func (c Conflict) Extension() core.Extension { return c.relation.Extension() }

// Relation returns c's underlying relation.Relation.
func (c Conflict) Relation() relation.Relation { return c.relation }

// IsZero reports whether c is the zero value.
func (c Conflict) IsZero() bool { return c.relation.IsZero() && c.nature == "" }

type conflictJSON struct {
	Relation relation.Relation `json:"relation"`
	Nature   string            `json:"nature"`
}

// MarshalJSON encodes c as {"relation": {...}, "nature": ...}.
func (c Conflict) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Conflict: %w", ErrInvalidRequirementRelation)
	}
	return json.Marshal(conflictJSON{Relation: c.relation, Nature: c.nature})
}

// UnmarshalJSON decodes c from its nested {"relation": {...}, "nature":
// ...} JSON form. The decoded relation's own type is revalidated against
// core.RelationTypeConflict, its source and target are revalidated as
// Requirement participants at either permitted level, and its scope is
// revalidated as present, before the same validation NewConflict applies
// (including participant distinctness) is run again -- a decoded Conflict
// can never be constructor-impossible. Decoding does not reorder or
// otherwise canonicalize the supplied source/target pair.
func (c *Conflict) UnmarshalJSON(data []byte) error {
	var raw conflictJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Conflict: %w: %w", ErrInvalidRequirementRelation, err)
	}

	if err := checkRelationType(raw.Relation, core.RelationTypeConflict); err != nil {
		return fmt.Errorf("requirement: unmarshal Conflict: %w", err)
	}
	participantA, err := asRequirementParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Conflict: %w", err)
	}
	participantB, err := asRequirementParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Conflict: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("requirement: unmarshal Conflict: %w", err)
	}

	result, err := NewConflict(participantA, participantB, raw.Relation.Provenance(), scope, raw.Nature)
	if err != nil {
		return err
	}

	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*c = result
	return nil
}
