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
// Dependency, Conflict, and Supersession -- including its two-arm participant
// union where a relation type admits more than one participant level.
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
// # Participant levels differ per relation type
//
// PEOS-009 fixes the participant level for two of the three and leaves it open
// for the third, and the difference is normative rather than incidental:
//
//	Generated-From   generated Artifact Revision -> Template Artifact Revision
//	Composition      Template Artifact Revision  -> Template Artifact Revision
//	Specialization   Template *or* Revision      -> Template *or* Revision
//
// So Generated-From and Composition take exact reference types directly, while
// Specialization takes TemplateParticipant on both sides and reports the chosen
// levels through ParticipantLevels() -- the "participant levels" item PEOS-009
// requires it to identify. Widening Generated-From or Composition, or narrowing
// Specialization, would each contradict the specification.
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
// A zero reference is rejected by the conversion itself, so there is no
// separate pre-check: duplicating it would leave the conversion's own error
// path unreachable.
func templateRevisionParticipant(ref core.TemplateArtifactRevisionRef) (core.EngineeringSubjectRef, error) {
	subject, err := core.EngineeringSubjectRefFromTemplateRevision(ref)
	if err != nil {
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: template artifact revision reference is invalid: %w", ErrInvalidTemplateRelation, err)
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

// --- TemplateParticipant -------------------------------------------------------

type templateParticipantKind string

const (
	templateParticipantKindTemplate templateParticipantKind = "template"
	templateParticipantKindRevision templateParticipantKind = "revision"
)

// TemplateParticipant is a Template Specialization participant: a closed
// two-arm union over a Template at identity level and an exact Template
// Artifact Revision.
//
// # Why both levels exist
//
// PEOS-009 requires a Template Specialization relation to identify "the source
// Template **or** Template Artifact Revision" and "the target base Template
// **or** Template Artifact Revision", and separately requires it to identify
// "participant levels". Both readings are therefore normative, and a level that
// could not vary would make the third obligation meaningless.
//
// The contrast with Template Composition is deliberate and is spelled out in
// the specification twice: "The source participant is the composing Template
// Artifact Revision. The target participant is the composed Template Artifact
// Revision", and again "its exact source (the composing Template Artifact
// Revision)". Composition is fixed at Revision level; Specialization is not.
// Composition therefore takes core.TemplateArtifactRevisionRef arguments
// directly and does not use this type.
//
// # Not a general participant type
//
// This union exists for Specialization alone. Generated-From's participants are
// fixed by PEOS-009 at generated Artifact Revision and Template Artifact
// Revision, and Composition's at Template Artifact Revision on both sides;
// neither may use this type, and widening either would contradict the
// specification.
//
// TemplateParticipant carries no identity, no revision system, and no lifecycle
// of its own -- it is a reference-shaped value, exactly like
// requirement.RequirementParticipant, the equivalent union PEOS-005 needed for
// Dependency and Conflict while its other four relation types stayed
// revision-only.
type TemplateParticipant struct {
	kind     templateParticipantKind
	template core.TemplateRef
	revision core.TemplateArtifactRevisionRef
}

// NewTemplateParticipantFromTemplate validates ref and returns a
// TemplateParticipant identifying a Template at identity level. ref must be
// non-zero.
func NewTemplateParticipantFromTemplate(ref core.TemplateRef) (TemplateParticipant, error) {
	p := TemplateParticipant{kind: templateParticipantKindTemplate, template: ref}
	if _, err := p.subject(); err != nil {
		return TemplateParticipant{}, fmt.Errorf("template: NewTemplateParticipantFromTemplate: %w", err)
	}
	return p, nil
}

// NewTemplateParticipantFromRevision validates ref and returns a
// TemplateParticipant identifying an exact Template Artifact Revision. ref must
// be non-zero.
func NewTemplateParticipantFromRevision(ref core.TemplateArtifactRevisionRef) (TemplateParticipant, error) {
	p := TemplateParticipant{kind: templateParticipantKindRevision, revision: ref}
	if _, err := p.subject(); err != nil {
		return TemplateParticipant{}, fmt.Errorf("template: NewTemplateParticipantFromRevision: %w", err)
	}
	return p, nil
}

// Kind returns p's participant level, "template" or "revision" -- the
// "participant levels" item PEOS-009 requires a Specialization relation to
// identify. The zero value returns the empty string.
func (p TemplateParticipant) Kind() string { return string(p.kind) }

// IsTemplateLevel reports whether p identifies a Template at identity level.
func (p TemplateParticipant) IsTemplateLevel() bool {
	return p.kind == templateParticipantKindTemplate
}

// IsRevisionLevel reports whether p identifies an exact Template Artifact
// Revision.
func (p TemplateParticipant) IsRevisionLevel() bool {
	return p.kind == templateParticipantKindRevision
}

// Template returns p's Template identity reference, and whether p is the
// identity-level variant.
func (p TemplateParticipant) Template() (core.TemplateRef, bool) {
	if p.kind != templateParticipantKindTemplate {
		return core.TemplateRef{}, false
	}
	return p.template, true
}

// Revision returns p's exact Template Artifact Revision reference, and whether
// p is the revision-level variant.
func (p TemplateParticipant) Revision() (core.TemplateArtifactRevisionRef, bool) {
	if p.kind != templateParticipantKindRevision {
		return core.TemplateArtifactRevisionRef{}, false
	}
	return p.revision, true
}

// ArtifactID returns the owning Template's core.ArtifactID at either level, so
// callers can compare participants across levels without unpacking the union.
// The zero value returns the zero ArtifactID.
func (p TemplateParticipant) ArtifactID() core.ArtifactID {
	switch p.kind {
	case templateParticipantKindTemplate:
		return p.template.ArtifactID()
	case templateParticipantKindRevision:
		return p.revision.ArtifactID()
	default:
		return core.ArtifactID{}
	}
}

// IsZero reports whether p is the zero value -- the unstated level PEOS-009
// does not permit on a valid Specialization.
func (p TemplateParticipant) IsZero() bool { return p.kind == "" }

// subject converts p into the relation participant subject relation.Relation
// requires, at whichever level p states.
//
// This is the union's single validation path: both constructors go through it,
// so a zero payload and an unstated level are each rejected in exactly one
// place, and no participant that fails here can ever be constructed.
func (p TemplateParticipant) subject() (core.EngineeringSubjectRef, error) {
	switch p.kind {
	case templateParticipantKindTemplate:
		subject, err := core.EngineeringSubjectRefFromTemplate(p.template)
		if err != nil {
			return core.EngineeringSubjectRef{}, fmt.Errorf("%w: %w", ErrInvalidTemplateParticipant, err)
		}
		return subject, nil
	case templateParticipantKindRevision:
		subject, err := core.EngineeringSubjectRefFromTemplateRevision(p.revision)
		if err != nil {
			return core.EngineeringSubjectRef{}, fmt.Errorf("%w: %w", ErrInvalidTemplateParticipant, err)
		}
		return subject, nil
	default:
		return core.EngineeringSubjectRef{}, fmt.Errorf("%w: participant level must be stated", ErrInvalidTemplateParticipant)
	}
}

// asTemplateParticipant recovers a TemplateParticipant from a decoded relation
// participant, accepting either level and rejecting anything else.
func asTemplateParticipant(subject core.EngineeringSubjectRef) (TemplateParticipant, error) {
	if ref, ok := subject.AsTemplateRevision(); ok {
		return NewTemplateParticipantFromRevision(ref)
	}
	if ref, ok := subject.AsTemplate(); ok {
		return NewTemplateParticipantFromTemplate(ref)
	}
	return TemplateParticipant{}, fmt.Errorf("%w: participant is neither a template nor a template artifact revision", ErrInvalidTemplateRelation)
}

type templateParticipantJSON struct {
	Kind     string                            `json:"kind"`
	Template *core.TemplateRef                 `json:"template,omitempty"`
	Revision *core.TemplateArtifactRevisionRef `json:"revision,omitempty"`
}

// MarshalJSON encodes p as {"kind":"template","template":...} or
// {"kind":"revision","revision":...}. The kind is always present: it is the
// participant level PEOS-009 requires the relation to identify, so it is stated
// on the wire rather than inferred from which payload happens to appear.
func (p TemplateParticipant) MarshalJSON() ([]byte, error) {
	switch p.kind {
	case templateParticipantKindTemplate:
		return json.Marshal(templateParticipantJSON{Kind: string(templateParticipantKindTemplate), Template: &p.template})
	case templateParticipantKindRevision:
		return json.Marshal(templateParticipantJSON{Kind: string(templateParticipantKindRevision), Revision: &p.revision})
	default:
		return nil, fmt.Errorf("template: marshal TemplateParticipant: %w", ErrInvalidTemplateParticipant)
	}
}

// UnmarshalJSON decodes p from its JSON form. A missing or unrecognized kind,
// an arm carrying the other arm's payload, and a selected arm missing its own
// payload are all rejected. The receiver is left untouched unless every check
// passes.
func (p *TemplateParticipant) UnmarshalJSON(data []byte) error {
	var raw templateParticipantJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal TemplateParticipant: %w: %w", ErrInvalidTemplateParticipant, err)
	}
	var result TemplateParticipant
	switch raw.Kind {
	case string(templateParticipantKindTemplate):
		if raw.Revision != nil {
			return fmt.Errorf("template: unmarshal TemplateParticipant: %w: a template-level participant must not carry a revision", ErrInvalidTemplateParticipant)
		}
		if raw.Template == nil {
			return fmt.Errorf("template: unmarshal TemplateParticipant: %w: a template-level participant requires a template reference", ErrInvalidTemplateParticipant)
		}
		var err error
		if result, err = NewTemplateParticipantFromTemplate(*raw.Template); err != nil {
			return err
		}
	case string(templateParticipantKindRevision):
		if raw.Template != nil {
			return fmt.Errorf("template: unmarshal TemplateParticipant: %w: a revision-level participant must not carry a template reference", ErrInvalidTemplateParticipant)
		}
		if raw.Revision == nil {
			return fmt.Errorf("template: unmarshal TemplateParticipant: %w: a revision-level participant requires a revision reference", ErrInvalidTemplateParticipant)
		}
		var err error
		if result, err = NewTemplateParticipantFromRevision(*raw.Revision); err != nil {
			return err
		}
	default:
		return fmt.Errorf("template: unmarshal TemplateParticipant: unrecognized kind %q: %w", raw.Kind, ErrInvalidTemplateParticipant)
	}
	*p = result
	return nil
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
// # Both participant levels are permitted
//
// PEOS-009 requires a Specialization to identify "the source Template **or**
// Template Artifact Revision", "the target base Template **or** Template
// Artifact Revision", and separately "participant levels". Both participants
// are therefore TemplateParticipant values, and each states its own level
// through Kind().
//
// This is deliberately unlike Composition, whose section fixes Revision level
// explicitly and twice ("The source participant is the composing Template
// Artifact Revision. The target participant is the composed Template Artifact
// Revision", and again "its exact source (the composing Template Artifact
// Revision)"). The two relation types are adjacent in the specification and are
// worded differently on purpose; collapsing Specialization to Revision level
// would make identity-level specialization unrepresentable and would leave
// "participant levels" a constant that cannot be identified. Packet K.3 raised
// exactly that as finding K3-01, and this is its correction.
//
// The two participants must differ. Equality is exact -- same level and same
// payload -- matching Composition, so two Revisions of one Template may
// specialize one another just as they may compose one another; PEOS-009
// prohibits cycles, not intra-Template specialization. Transitive cycle
// detection remains repository-owned.
type Specialization struct {
	relation            relation.Relation
	inheritedElements   string
	overriddenElements  string
	compatibilityEffect string
}

// NewSpecialization validates its arguments and returns a Specialization whose
// relation type is always core.RelationTypeTemplateSpecialization.
//
// specializing and base are TemplateParticipant values and may each be at
// either level -- Template identity or exact Template Artifact Revision -- per
// PEOS-009's "the source Template or Template Artifact Revision". Both must be
// non-zero and must differ; an identical pair is the degenerate direct cycle
// "Specialization cycles SHALL NOT be permitted" forbids. scope and provenance
// must be non-zero. inheritedElements, overriddenElements, and
// compatibilityEffect must each be non-empty after trimming; the trimmed values
// are stored and none is interpreted.
func NewSpecialization(
	specializing TemplateParticipant,
	base TemplateParticipant,
	scope core.Scope,
	provenance core.Provenance,
	inheritedElements string,
	overriddenElements string,
	compatibilityEffect string,
) (Specialization, error) {
	// subject() reports an unstated level itself, so there is no separate
	// IsZero pre-check here: duplicating it would make subject()'s own
	// completeness guard unreachable.
	source, err := specializing.subject()
	if err != nil {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: specializing participant: %w", err)
	}
	target, err := base.subject()
	if err != nil {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: base participant: %w", err)
	}
	if specializing == base {
		return Specialization{}, fmt.Errorf("template: NewSpecialization: %w: a template must not specialize itself", ErrInvalidSpecialization)
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

// Specializing returns the participant doing the specializing (s's source), at
// whichever level it states. Its Kind() is the source participant level.
func (s Specialization) Specializing() (TemplateParticipant, bool) {
	p, err := asTemplateParticipant(s.relation.Source())
	if err != nil {
		return TemplateParticipant{}, false
	}
	return p, true
}

// Base returns the base participant being specialized (s's target), at
// whichever level it states. Its Kind() is the target participant level.
func (s Specialization) Base() (TemplateParticipant, bool) {
	p, err := asTemplateParticipant(s.relation.Target())
	if err != nil {
		return TemplateParticipant{}, false
	}
	return p, true
}

// ParticipantLevels returns the source and target participant levels --
// "template" or "revision" -- which PEOS-009 requires every Template
// Specialization relation to identify alongside its participants.
func (s Specialization) ParticipantLevels() (source string, target string) {
	if p, ok := s.Specializing(); ok {
		source = p.Kind()
	}
	if p, ok := s.Base(); ok {
		target = p.Kind()
	}
	return source, target
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
	specializing, err := asTemplateParticipant(raw.Relation.Source())
	if err != nil {
		return fmt.Errorf("template: unmarshal Specialization: %w", err)
	}
	base, err := asTemplateParticipant(raw.Relation.Target())
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
