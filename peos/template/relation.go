package template

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// This file implements the three PEOS-009 Artifact Relation types as typed
// wrappers over relation.Relation: Generated-From, Template Composition, and
// Template Specialization.
//
// # Why wrappers rather than bare relation.Relation values
//
// Each of the three carries SHALL-identify state a bare relation.Relation
// cannot hold -- Composition's "parameter mapping rules" and "conflict
// handling", Specialization's "inherited elements", "overridden elements", and
// "compatibility effect". Generated-From carries no extra state but needs its
// participant levels and direction fixed, and needs the guarantee that it
// never accumulates the state PEOS-009 forbids it. This is the same wrapper
// strategy peos/requirement uses for Derivation, Refinement, Decomposition,
// Dependency, Conflict, and Supersession.
//
// # No identity, no revision, no lifecycle
//
// PEOS-009 is explicit for Generated-From ("identity: none (per PEOS-002's
// general Artifact Relation contract); revision: none; lifecycle: none") and
// relies on PEOS-002's general contract for the other two. None of the three
// types below has an ID field, a Ref method, a revision, or a lifecycle state,
// and relation.Relation itself has none either.
//
// # Binarity and cycles
//
// Binarity is structural: relation.Relation identifies exactly one source and
// exactly one target, so "Template Composition Hyperedge" -- a relation with
// more than one source or target -- is unrepresentable rather than merely
// forbidden. Many-to-many composition is expressed as multiple binary
// relations, exactly as PEOS-009 requires.
//
// Each constructor rejects the degenerate direct cycle (source == target).
// Transitive cycle detection is repository-owned: a value layer sees one
// relation at a time and cannot know the rest of the graph. peos/requirement
// documents the identical division for its own cycle-prohibited relation types,
// and PEOS-009 itself assigns cross-artifact traversal to "a future
// Traceability Model".
//
// # Mandatory scope
//
// relation.Relation treats scope as optional (WithScope/WithoutScope), but
// PEOS-009 lists "its scope" unqualified for all three relation types, so each
// wrapper takes scope as a mandatory constructor argument, sets it on the inner
// relation immediately, and re-verifies its presence on decode. Neither
// WithScope nor WithoutScope is exposed -- a wrapper can never lose its scope.

// checkRelationType verifies that a decoded relation carries exactly the type
// its wrapper fixes, so a Composition's JSON cannot decode into a
// Specialization or vice versa.
func checkRelationType(rel relation.Relation, want core.RelationType) error {
	if rel.RelationType() != want {
		return fmt.Errorf("%w: relation type is %q, want %q", ErrInvalidTemplateRelation, rel.RelationType().String(), want.String())
	}
	return nil
}

// requireRelationScope verifies that a decoded relation carries the scope
// PEOS-009 states unqualified for every relation type it defines.
func requireRelationScope(rel relation.Relation) (core.Scope, error) {
	scope, ok := rel.Scope()
	if !ok {
		return core.Scope{}, fmt.Errorf("%w: scope must not be absent", ErrInvalidTemplateRelation)
	}
	return scope, nil
}

// templateRevisionParticipant converts an exact Template Artifact Revision
// reference into the relation participant subject relation.Relation requires.
func templateRevisionParticipant(ref core.TemplateArtifactRevisionRef) (core.EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: template artifact revision reference must not be zero", ErrInvalidTemplateRelation)
	}
	subject, err := core.EngineeringSubjectRefFromTemplateRevision(ref)
	if err != nil {
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: %w", ErrInvalidTemplateRelation, err)
	}
	return subject, nil
}

// asTemplateRevisionParticipant recovers the exact Template Artifact Revision
// reference from a decoded participant, rejecting any other participant level.
func asTemplateRevisionParticipant(subject core.EngineeringSubjectRef) (core.TemplateArtifactRevisionRef, error) {
	ref, ok := subject.AsTemplateRevision()
	if !ok {
		return core.TemplateArtifactRevisionRef{}, fmt.Errorf("%w: participant is not a template artifact revision", ErrInvalidTemplateRelation)
	}
	return ref, nil
}

// --- GeneratedFrom -------------------------------------------------------------

// GeneratedFrom is the PEOS-009 Generated-From relation: an optional,
// supplementary traceability link from a generated Artifact Revision to the
// Template Artifact Revision that produced it.
//
//	source participant: the generated Artifact Revision
//	target participant: the Template Artifact Revision
//	direction:          generated → template
//	multiplicity:       many generated Revisions MAY reference one Template Revision
//	cycles:             prohibited
//	identity:           none
//	revision:           none
//	lifecycle:          none
//
// # It never substitutes for the Application Record
//
// "Generated-From is supplementary traceability. It does not replace the
// Template Application Record, and a conformant implementation MAY omit
// Generated-From entirely provided the Application Record itself remains
// inspectable." Two named non-conforming patterns follow from that, and both
// are structurally impossible here: "Generated-From Relation as Application
// Record" and "Parameter Values Stored Only on Generated-From Relation".
//
// PEOS-009 states directly what this relation SHALL NOT contain: "the full
// resolved parameter state; execution event history; mutable application
// status; authority history." GeneratedFrom therefore adds **no field of its
// own** beyond the inner relation -- it carries no resolved values, no
// outcome, no status, no authority, and no event history. All of that belongs
// exclusively to ApplicationRecord.
type GeneratedFrom struct {
	relation relation.Relation
}

// NewGeneratedFrom validates its arguments and returns a GeneratedFrom whose
// relation type is always core.RelationTypeGeneratedFrom -- never a caller
// input, so a GeneratedFrom can never carry another relation type.
//
// generated must be a non-zero exact generated Artifact Revision reference and
// template a non-zero exact Template Artifact Revision reference; scope and
// provenance must both be non-zero. The direction is fixed: generated is the
// source, template is the target.
func NewGeneratedFrom(
	generated core.GeneratedArtifactRevisionRef,
	template core.TemplateArtifactRevisionRef,
	scope core.Scope,
	provenance core.Provenance,
) (GeneratedFrom, error) {
	if generated.IsZero() {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w: generated artifact revision reference must not be zero", ErrInvalidGeneratedFrom)
	}
	source, err := core.EngineeringSubjectRefFromGeneratedArtifactRevision(generated)
	if err != nil {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w: %w", ErrInvalidGeneratedFrom, err)
	}
	target, err := templateRevisionParticipant(template)
	if err != nil {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w", err)
	}
	if scope.IsZero() {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if provenance.IsZero() {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w: provenance must not be zero", ErrInvalidTemplateRelation)
	}

	rel, err := relation.New(core.RelationTypeGeneratedFrom, source, target, provenance)
	if err != nil {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w: %w", ErrInvalidTemplateRelation, err)
	}
	if rel, err = rel.WithScope(scope); err != nil {
		return GeneratedFrom{}, fmt.Errorf("template: NewGeneratedFrom: %w: %w", ErrInvalidTemplateRelation, err)
	}
	return GeneratedFrom{relation: rel}, nil
}

// Relation returns the underlying relation.Relation.
func (g GeneratedFrom) Relation() relation.Relation { return g.relation }

// Generated returns the exact generated Artifact Revision g's source names.
func (g GeneratedFrom) Generated() (core.GeneratedArtifactRevisionRef, bool) {
	return g.relation.Source().AsGeneratedArtifactRevision()
}

// Template returns the exact Template Artifact Revision g's target names.
func (g GeneratedFrom) Template() (core.TemplateArtifactRevisionRef, bool) {
	return g.relation.Target().AsTemplateRevision()
}

// Scope returns g's declared scope. It is mandatory and therefore never absent
// on a valid GeneratedFrom.
func (g GeneratedFrom) Scope() core.Scope {
	scope, _ := g.relation.Scope()
	return scope
}

// Provenance returns g's declared provenance.
func (g GeneratedFrom) Provenance() core.Provenance { return g.relation.Provenance() }

// WithExtension returns a copy of g with its extension data set.
func (g GeneratedFrom) WithExtension(extension core.Extension) GeneratedFrom {
	g.relation = g.relation.WithExtension(extension)
	return g
}

// WithoutExtension returns a copy of g with its extension data cleared.
func (g GeneratedFrom) WithoutExtension() GeneratedFrom {
	g.relation = g.relation.WithoutExtension()
	return g
}

// Extension returns g's extension data.
func (g GeneratedFrom) Extension() core.Extension { return g.relation.Extension() }

// IsZero reports whether g is the zero value.
func (g GeneratedFrom) IsZero() bool { return g.relation.IsZero() }

type generatedFromJSON struct {
	Relation relation.Relation `json:"relation"`
}

// MarshalJSON encodes g as {"relation":{...}}. There is no "resolved_values",
// "outcome", "status", "authority_history", or "events" key -- PEOS-009 states
// directly that a Generated-From relation SHALL NOT contain any of them.
func (g GeneratedFrom) MarshalJSON() ([]byte, error) {
	if g.IsZero() {
		return nil, fmt.Errorf("template: marshal GeneratedFrom: %w", ErrInvalidGeneratedFrom)
	}
	return json.Marshal(generatedFromJSON{Relation: g.relation})
}

// UnmarshalJSON decodes g from its nested {"relation":{...}} form. The decoded
// relation's type is revalidated against core.RelationTypeGeneratedFrom, both
// participants are revalidated at their required levels, and its scope is
// revalidated as present, before the same validation NewGeneratedFrom applies
// runs again -- a decoded GeneratedFrom can never be constructor-impossible.
func (g *GeneratedFrom) UnmarshalJSON(data []byte) error {
	var raw generatedFromJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal GeneratedFrom: %w: %w", ErrInvalidGeneratedFrom, err)
	}
	if err := checkRelationType(raw.Relation, core.RelationTypeGeneratedFrom); err != nil {
		return fmt.Errorf("template: unmarshal GeneratedFrom: %w", err)
	}
	generated, ok := raw.Relation.Source().AsGeneratedArtifactRevision()
	if !ok {
		return fmt.Errorf("template: unmarshal GeneratedFrom: %w: source is not a generated artifact revision", ErrInvalidGeneratedFrom)
	}
	template, err := asTemplateRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("template: unmarshal GeneratedFrom: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("template: unmarshal GeneratedFrom: %w", err)
	}

	result, err := NewGeneratedFrom(generated, template, scope, raw.Relation.Provenance())
	if err != nil {
		return err
	}
	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*g = result
	return nil
}

// --- Composition ---------------------------------------------------------------

// Composition is the PEOS-009 Template Composition relation: one Template
// Artifact Revision composing another.
//
//	source participant: the composing Template Artifact Revision
//	target participant: the composed Template Artifact Revision
//	direction:          composing → composed
//	multiplicity:       many-to-many, via multiple binary relations only
//	cycles:             prohibited
//
// PEOS-009 requires every Composition relation to identify ten things. Six --
// exact source, exact target, participant levels, direction, scope, provenance
// -- are carried structurally by the inner relation.Relation and by this
// wrapper's fixed participant types. Multiplicity and cycle policy are
// properties of the relation *type* rather than of any one instance, and are
// documented here rather than stored as per-instance fields: storing "cycles
// are prohibited" on each relation would record the same constant on every
// value. The remaining two -- parameter mapping rules and conflict handling --
// are per-instance state and are stored as opaque trimmed descriptors, because
// PEOS-009 defines no mapping or conflict-resolution language.
//
// # No collection entity
//
// "This specification does not introduce: a Template Collection; a Template
// Composition Set; a hyperedge; any relationship-group identity." A composed
// set of Templates is interpreted together by a repository from multiple binary
// relations; it never becomes a value in this package.
type Composition struct {
	relation         relation.Relation
	parameterMapping string
	conflictHandling string
}

// NewComposition validates its arguments and returns a Composition whose
// relation type is always core.RelationTypeTemplateComposition.
//
// composing and composed must both be non-zero exact Template Artifact
// Revision references and must differ -- an identical pair is the degenerate
// direct cycle PEOS-009's "Composition cycles SHALL NOT be permitted" forbids.
// scope and provenance must be non-zero. parameterMapping and conflictHandling
// must each be non-empty after trimming; the trimmed values are stored and
// neither is interpreted.
func NewComposition(
	composing core.TemplateArtifactRevisionRef,
	composed core.TemplateArtifactRevisionRef,
	scope core.Scope,
	provenance core.Provenance,
	parameterMapping string,
	conflictHandling string,
) (Composition, error) {
	source, err := templateRevisionParticipant(composing)
	if err != nil {
		return Composition{}, fmt.Errorf("template: NewComposition: %w", err)
	}
	target, err := templateRevisionParticipant(composed)
	if err != nil {
		return Composition{}, fmt.Errorf("template: NewComposition: %w", err)
	}
	if composing == composed {
		return Composition{}, fmt.Errorf("template: NewComposition: %w: a template artifact revision must not compose itself", ErrInvalidComposition)
	}
	if scope.IsZero() {
		return Composition{}, fmt.Errorf("template: NewComposition: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if provenance.IsZero() {
		return Composition{}, fmt.Errorf("template: NewComposition: %w: provenance must not be zero", ErrInvalidTemplateRelation)
	}
	trimmedMapping, err := trimmedRequired("NewComposition", "parameter mapping", parameterMapping, ErrInvalidComposition)
	if err != nil {
		return Composition{}, err
	}
	trimmedConflict, err := trimmedRequired("NewComposition", "conflict handling", conflictHandling, ErrInvalidComposition)
	if err != nil {
		return Composition{}, err
	}

	rel, err := relation.New(core.RelationTypeTemplateComposition, source, target, provenance)
	if err != nil {
		return Composition{}, fmt.Errorf("template: NewComposition: %w: %w", ErrInvalidTemplateRelation, err)
	}
	if rel, err = rel.WithScope(scope); err != nil {
		return Composition{}, fmt.Errorf("template: NewComposition: %w: %w", ErrInvalidTemplateRelation, err)
	}
	return Composition{relation: rel, parameterMapping: trimmedMapping, conflictHandling: trimmedConflict}, nil
}

// Relation returns the underlying relation.Relation.
func (c Composition) Relation() relation.Relation { return c.relation }

// Composing returns the exact Template Artifact Revision doing the composing
// (c's source).
func (c Composition) Composing() (core.TemplateArtifactRevisionRef, bool) {
	return c.relation.Source().AsTemplateRevision()
}

// Composed returns the exact Template Artifact Revision being composed (c's
// target).
func (c Composition) Composed() (core.TemplateArtifactRevisionRef, bool) {
	return c.relation.Target().AsTemplateRevision()
}

// ParameterMapping returns c's declared parameter mapping rules,
// uninterpreted.
func (c Composition) ParameterMapping() string { return c.parameterMapping }

// ConflictHandling returns c's declared conflict handling, uninterpreted.
func (c Composition) ConflictHandling() string { return c.conflictHandling }

// Scope returns c's declared scope. It is mandatory and therefore never absent
// on a valid Composition.
func (c Composition) Scope() core.Scope {
	scope, _ := c.relation.Scope()
	return scope
}

// Provenance returns c's declared provenance.
func (c Composition) Provenance() core.Provenance { return c.relation.Provenance() }

// WithExtension returns a copy of c with its extension data set.
func (c Composition) WithExtension(extension core.Extension) Composition {
	c.relation = c.relation.WithExtension(extension)
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c Composition) WithoutExtension() Composition {
	c.relation = c.relation.WithoutExtension()
	return c
}

// Extension returns c's extension data.
func (c Composition) Extension() core.Extension { return c.relation.Extension() }

// IsZero reports whether c is the zero value.
func (c Composition) IsZero() bool {
	return c.relation.IsZero() && c.parameterMapping == "" && c.conflictHandling == ""
}

type compositionJSON struct {
	Relation         relation.Relation `json:"relation"`
	ParameterMapping string            `json:"parameter_mapping"`
	ConflictHandling string            `json:"conflict_handling"`
}

// MarshalJSON encodes c as {"relation":{...},"parameter_mapping":...,
// "conflict_handling":...}. There is no "cycles", "multiplicity", or
// "direction" key: all three are properties of the relation type, constant
// across every instance, and are documented rather than stored.
func (c Composition) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("template: marshal Composition: %w", ErrInvalidComposition)
	}
	return json.Marshal(compositionJSON{
		Relation:         c.relation,
		ParameterMapping: c.parameterMapping,
		ConflictHandling: c.conflictHandling,
	})
}

// UnmarshalJSON decodes c from its nested JSON form, revalidating the relation
// type, both participant levels, and the mandatory scope before rerunning the
// same validation NewComposition applies.
func (c *Composition) UnmarshalJSON(data []byte) error {
	var raw compositionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal Composition: %w: %w", ErrInvalidComposition, err)
	}
	if err := checkRelationType(raw.Relation, core.RelationTypeTemplateComposition); err != nil {
		return fmt.Errorf("template: unmarshal Composition: %w", err)
	}
	composing, err := asTemplateRevisionParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("template: unmarshal Composition: %w", err)
	}
	composed, err := asTemplateRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("template: unmarshal Composition: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("template: unmarshal Composition: %w", err)
	}

	result, err := NewComposition(composing, composed, scope, raw.Relation.Provenance(), raw.ParameterMapping, raw.ConflictHandling)
	if err != nil {
		return err
	}
	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*c = result
	return nil
}

// --- Specialization ------------------------------------------------------------

// Specialization is the PEOS-009 Template Specialization relation: one Template
// Artifact Revision specializing another.
//
//	source participant: the specializing Template Artifact Revision
//	target participant: the base Template Artifact Revision
//	direction:          specializing → base
//	cycles:             prohibited
//
// "A specialized Template SHALL preserve an explicit relation to the Template
// it specializes", and "A specialization does not mutate the base Template
// Artifact Revision" -- which this type satisfies structurally, since it holds
// only references and cannot reach the base Revision's content.
//
// Of the eight things PEOS-009 requires a Specialization relation to identify,
// source, target, participant levels, scope, and provenance are carried by the
// inner relation and this wrapper's fixed participant types; inherited
// elements, overridden elements, and compatibility effect are per-instance
// state stored as opaque trimmed descriptors, since PEOS-009 defines no
// inheritance or override language.
//
// The compatibility-effect descriptor is a *declaration* of how specializing
// affects compatibility, never a compatibility verdict -- current compatibility
// remains derived at query time.
//
// PEOS-009 permits a Specialization to name "the source Template or Template
// Artifact Revision" and "the target base Template or Template Artifact
// Revision", so both participant levels appear admissible. This type fixes both
// at Revision level, because "A composition reference SHALL identify the exact
// Template Artifact Revision, where exact content matters" and inherited and
// overridden elements are exact content: which elements a specialization
// inherits or overrides is undefined against a bare Template identity whose
// content can change with every new Revision. A Product needing
// identity-level specialization records that in its own contract.
type Specialization struct {
	relation            relation.Relation
	inheritedElements   string
	overriddenElements  string
	compatibilityEffect string
}

// NewSpecialization validates its arguments and returns a Specialization whose
// relation type is always core.RelationTypeTemplateSpecialization.
//
// specializing and base must both be non-zero exact Template Artifact Revision
// references and must differ -- an identical pair is the degenerate direct
// cycle PEOS-009's "Specialization cycles SHALL NOT be permitted" forbids.
// scope and provenance must be non-zero. inheritedElements,
// overriddenElements, and compatibilityEffect must each be non-empty after
// trimming; the trimmed values are stored and none is interpreted.
func NewSpecialization(
	specializing core.TemplateArtifactRevisionRef,
	base core.TemplateArtifactRevisionRef,
	scope core.Scope,
	provenance core.Provenance,
	inheritedElements string,
	overriddenElements string,
	compatibilityEffect string,
) (Specialization, error) {
	source, err := templateRevisionParticipant(specializing)
	if err != nil {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w", err)
	}
	target, err := templateRevisionParticipant(base)
	if err != nil {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w", err)
	}
	if specializing == base {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w: a template artifact revision must not specialize itself", ErrInvalidSpecialization)
	}
	if scope.IsZero() {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if provenance.IsZero() {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w: provenance must not be zero", ErrInvalidTemplateRelation)
	}
	trimmedInherited, err := trimmedRequired("NewSpecialization", "inherited elements", inheritedElements, ErrInvalidSpecialization)
	if err != nil {
		return Specialization{}, err
	}
	trimmedOverridden, err := trimmedRequired("NewSpecialization", "overridden elements", overriddenElements, ErrInvalidSpecialization)
	if err != nil {
		return Specialization{}, err
	}
	trimmedEffect, err := trimmedRequired("NewSpecialization", "compatibility effect", compatibilityEffect, ErrInvalidSpecialization)
	if err != nil {
		return Specialization{}, err
	}

	rel, err := relation.New(core.RelationTypeTemplateSpecialization, source, target, provenance)
	if err != nil {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w: %w", ErrInvalidTemplateRelation, err)
	}
	if rel, err = rel.WithScope(scope); err != nil {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w: %w", ErrInvalidTemplateRelation, err)
	}
	return Specialization{
		relation:            rel,
		inheritedElements:   trimmedInherited,
		overriddenElements:  trimmedOverridden,
		compatibilityEffect: trimmedEffect,
	}, nil
}

// Relation returns the underlying relation.Relation.
func (s Specialization) Relation() relation.Relation { return s.relation }

// Specializing returns the exact Template Artifact Revision doing the
// specializing (s's source).
func (s Specialization) Specializing() (core.TemplateArtifactRevisionRef, bool) {
	return s.relation.Source().AsTemplateRevision()
}

// Base returns the exact base Template Artifact Revision being specialized
// (s's target).
func (s Specialization) Base() (core.TemplateArtifactRevisionRef, bool) {
	return s.relation.Target().AsTemplateRevision()
}

// InheritedElements returns s's declared inherited elements, uninterpreted.
func (s Specialization) InheritedElements() string { return s.inheritedElements }

// OverriddenElements returns s's declared overridden elements, uninterpreted.
func (s Specialization) OverriddenElements() string { return s.overriddenElements }

// CompatibilityEffect returns s's declared compatibility effect,
// uninterpreted. This is a declaration, never a compatibility verdict.
func (s Specialization) CompatibilityEffect() string { return s.compatibilityEffect }

// Scope returns s's declared scope. It is mandatory and therefore never absent
// on a valid Specialization.
func (s Specialization) Scope() core.Scope {
	scope, _ := s.relation.Scope()
	return scope
}

// Provenance returns s's declared provenance.
func (s Specialization) Provenance() core.Provenance { return s.relation.Provenance() }

// WithExtension returns a copy of s with its extension data set.
func (s Specialization) WithExtension(extension core.Extension) Specialization {
	s.relation = s.relation.WithExtension(extension)
	return s
}

// WithoutExtension returns a copy of s with its extension data cleared.
func (s Specialization) WithoutExtension() Specialization {
	s.relation = s.relation.WithoutExtension()
	return s
}

// Extension returns s's extension data.
func (s Specialization) Extension() core.Extension { return s.relation.Extension() }

// IsZero reports whether s is the zero value.
func (s Specialization) IsZero() bool {
	return s.relation.IsZero() && s.inheritedElements == "" && s.overriddenElements == "" &&
		s.compatibilityEffect == ""
}

type specializationJSON struct {
	Relation            relation.Relation `json:"relation"`
	InheritedElements   string            `json:"inherited_elements"`
	OverriddenElements  string            `json:"overridden_elements"`
	CompatibilityEffect string            `json:"compatibility_effect"`
}

// MarshalJSON encodes s as {"relation":{...},"inherited_elements":...,
// "overridden_elements":...,"compatibility_effect":...}. There is no
// "compatible" or "compatibility" verdict key -- compatibility_effect declares
// an effect, and current compatibility stays derived.
func (s Specialization) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("template: marshal Specialization: %w", ErrInvalidSpecialization)
	}
	return json.Marshal(specializationJSON{
		Relation:            s.relation,
		InheritedElements:   s.inheritedElements,
		OverriddenElements:  s.overriddenElements,
		CompatibilityEffect: s.compatibilityEffect,
	})
}

// UnmarshalJSON decodes s from its nested JSON form, revalidating the relation
// type, both participant levels, and the mandatory scope before rerunning the
// same validation NewSpecialization applies.
func (s *Specialization) UnmarshalJSON(data []byte) error {
	var raw specializationJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal Specialization: %w: %w", ErrInvalidSpecialization, err)
	}
	if err := checkRelationType(raw.Relation, core.RelationTypeTemplateSpecialization); err != nil {
		return fmt.Errorf("template: unmarshal Specialization: %w", err)
	}
	specializing, err := asTemplateRevisionParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("template: unmarshal Specialization: %w", err)
	}
	base, err := asTemplateRevisionParticipant(raw.Relation.Target())
	if err != nil {
		return fmt.Errorf("template: unmarshal Specialization: %w", err)
	}
	scope, err := requireRelationScope(raw.Relation)
	if err != nil {
		return fmt.Errorf("template: unmarshal Specialization: %w", err)
	}

	result, err := NewSpecialization(
		specializing, base, scope, raw.Relation.Provenance(),
		raw.InheritedElements, raw.OverriddenElements, raw.CompatibilityEffect,
	)
	if err != nil {
		return err
	}
	if ext := raw.Relation.Extension(); !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*s = result
	return nil
}
