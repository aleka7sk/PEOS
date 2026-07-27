// Package template implements PEOS-009's Template Contract Model on top of
// the stable peos/core package.
//
// # Ontology
//
//	Template                    Artifact (ordinary core.Artifact)
//	TemplateRevision            Artifact Revision (ordinary core.ArtifactRevision)
//	TemplateContent             Revision-owned typed content
//	TemplateApplicability       Revision-owned closed union
//	Parameter                   Revision-owned value, locally keyed
//	ParameterType               Revision-owned closed union (vocabulary | external)
//	ParameterDefault            Revision-owned value, targets a parameter key
//	ParameterConstraint         Revision-owned value, locally keyed
//	ConstraintTarget            closed union (parameter | generated content)
//	CompatibilityDeclaration    Revision-owned value
//
// Template is the only Artifact this package defines. Everything else is
// Revision-owned content with no identity of its own: PEOS-009 grants no
// independent identity to a Template Parameter, a Parameter Default, a
// Parameter Constraint, or a compatibility declaration, and "Granting a
// Template Parameter an identity that survives independently of its owning
// Template Artifact Revision" is a named non-conforming pattern.
//
// # Template is an Artifact, and there is no Template Version
//
// "Template SHALL be an Artifact, as defined by PEOS-002. Template SHALL use
// ordinary Artifact Revision." PEOS-009 then explicitly declines to introduce
// a parallel revision system: "This specification does not introduce `Template
// Version` or `Template Revision` as a revision system parallel to Artifact
// Revision." The phrase "Template Revision" is permitted only as informal
// shorthand for "Artifact Revision whose Artifact is a Template", which is
// exactly what the TemplateRevision type is -- a named-field composition of
// core.ArtifactRevision plus typed content, following the same
// specialized-Revision strategy as requirement.Revision,
// validation.PlanRevision, quality.ProfileRevision, and
// runtime.ContractRevision.
//
// Consequently there is no Version field, no version accessor, and no wire
// key for one anywhere in this package.
//
// # The template body lives in core.Representation, not in TemplateContent
//
// This is the single most important boundary in this package, and it is easy
// to get wrong by analogy to ordinary templating libraries.
//
// PEOS-009 assigns the body to PEOS-002: "Template syntax, schema, rendering
// language, source code, natural-language text, model, or other
// representation belongs to the Template Artifact Revision, in accordance with
// PEOS-002's Artifact Representation contract", and "Representation is not
// Template identity." core.ArtifactRevision.Representations() already provides
// exactly that, with five content forms -- inline text, inline bytes, content
// address, external reference, and composed.
//
// TemplateContent therefore has no body, source, template_text, schema,
// script, expression, or renderable_content field. Adding one would duplicate
// PEOS-002's Representation contract and create two competing sources of truth
// for the same content.
//
// TemplateContent.ExpansionSemantics() is a declarative descriptor of how that
// Representation is to be expanded -- not the thing expanded, and not a
// program. PEOS-009 defines no interpolation syntax, no ordering rule, no
// conditional or loop construct, and no expression language; its Non-Goals
// disclaim "a specific templating language or engine". No renderer, evaluator,
// generator, parser, or engine of any kind exists in this package, and none is
// planned for a later packet.
//
// # TemplateContent declares a generation contract, never a generated output
//
// TemplateContent names the Artifact *Types* a generated Artifact may declare.
// It never names a generated Artifact or Revision, because PEOS-009 gives a
// generated Artifact "its own Artifact identity, independent of the Template's
// identity" and forbids "Representing a generated Artifact as sharing, or
// inheriting, the Template's own Artifact identity". A generated Artifact is an
// ordinary Artifact; this package owns none of its identity, and there is no
// Template Instance construct ("This specification does not create a 'Template
// Instance' construct").
//
// A permitted generated Artifact Type may name Requirement, Validation Plan,
// Quality Profile, Runtime Contract, or any other valid Artifact Type. Naming
// a type creates no reference to any instance of it, which is why this package
// imports none of those packages.
//
// # Two separate template-local key namespaces
//
// PEOS-009 states that a Template Parameter's key "is unique only within that
// exact Template Artifact Revision; is not an Artifact Identity; is not a
// global Template Parameter identity", and that it "MAY be referenced by
// constraints, defaults, composition mappings, and Template Application
// Records". It states no cross-collection uniqueness rule at all.
//
// This package therefore maintains two namespaces, split by reference kind
// rather than by Go collection:
//
//	parameter namespace   = TemplateContent.parameters
//	    resolved by: a ParameterDefault's parameter key,
//	                 a parameter-targeting ParameterConstraint,
//	                 and (in a later packet) composition parameter mappings
//	                 and Template Application Record resolved values.
//	    public resolver: TemplateContent.Parameter(key)
//
//	constraint namespace  = TemplateContent.constraints
//	    resolved by: core.CriterionKindTemplateConstraint, whose payload
//	                 core.TemplateConstraintCriterionRef is a
//	                 (Template Artifact Revision, LocalKey) pair.
//	    public resolver: TemplateContent.Constraint(key)
//
// A duplicate key within either namespace is rejected with
// ErrDuplicateTemplateLocalKey; one key used once by a Parameter and once by a
// ParameterConstraint is accepted, because the two are named by genuinely
// different reference kinds.
//
// This is the same principle Packet J.2.A settled for peos/runtime, applied to
// a simpler case. There, four Go collections had to share one namespace
// because core.RuntimeRuleCriterionRef carries no rule-category discriminator;
// here the two collections are addressed by different reference kinds, so they
// are genuinely independent.
//
// PEOS-009 does not explicitly require a Parameter Constraint to have a local
// key -- it requires one only for a Parameter. The key on ParameterConstraint
// is a derived structural requirement: core.CriterionKindTemplateConstraint has
// no other resolution target, so an unkeyed constraint collection would leave
// that criterion kind unresolvable. That is precisely the defect Packet J.3
// raised as J3-03 against peos/runtime's originally unkeyed Contract Rule
// collections, and keying constraints from the start avoids repeating it.
//
// # Compatibility and conformance are derived, never stored
//
// CompatibilityDeclaration holds declared inputs only. "Current compatibility
// is a derived interpretation, computed from the applicable compatibility
// declarations at query time", and both Template.compatible and
// TemplateRevision.compatible are named non-conforming patterns. There is no
// Compatible() method, no boolean, no status field, and no wire key for any of
// them.
//
// Template conformance is likewise derived, from Template Conformance Claims,
// and appears nowhere in this package: Template.conformant,
// TemplateRevision.conformant, and GeneratedArtifact.conformant are all named
// non-conforming patterns.
//
// # Product-owned interpretation
//
// This package validates structure and never interprets meaning. The following
// are opaque, trimmed strings or Product-defined vocabularies, deliberately:
//
//   - the template body itself (core.Representation's content);
//   - TemplateContent's expansion or generation semantics;
//   - a ParameterConstraint's rule -- PEOS-009 defines no constraint grammar;
//   - a ParameterConstraint's evaluation point and failure semantics, both
//     Product-defined vocabularies with no PEOS-predeclared values;
//   - a ParameterDefault's value;
//   - a ParameterType's external locator and governing authority;
//   - a CompatibilityDeclaration's parameter contract, Product contract, and
//     migration requirements;
//   - a generated-content ConstraintTarget's descriptor;
//   - core.Scope's expression, in applicability or in a constraint's scope.
//
// Encoding any of them would be a framework PEOS-009 deliberately does not
// define.
//
// # Repository responsibilities
//
// This package models values and enforces PEOS-009's structural invariants. It
// does not persist, index, query, or resolve anything across Revisions or
// against generated content. A repository built on it owns:
//
//   - storing Templates, Template Revisions, Template Application Records, and
//     generated Artifacts, and retrieving them by identity;
//   - deriving current template compatibility from applicable compatibility
//     declarations, and Template conformance from applicable Template
//     Conformance Claims;
//   - detecting transitive composition and specialization cycles. PEOS-009
//     prohibits both ("Composition cycles SHALL NOT be permitted",
//     "Specialization cycles SHALL NOT be permitted"), but a value layer sees
//     one Revision at a time and cannot detect a transitive cycle. This is the
//     same division peos/requirement already documents for Derivation,
//     Refinement, Decomposition, and Requirement Supersession cycles, and
//     PEOS-009 itself assigns cross-artifact traversal to "a future
//     Traceability Model";
//   - evaluating a ParameterConstraint's rule against whatever it names, and
//     resolving parameter values at application time;
//   - actual generation, rendering, and expansion -- all Product-owned and out
//     of scope for this package.
//
// # Package dependency boundary
//
// Production sources import only the standard library, peos/core,
// peos/relation, and peos/validation. peos/relation is required because
// PEOS-009 defines three Artifact Relation types, each carrying SHALL-identify
// state a bare relation.Relation cannot hold. peos/validation is required by
// exactly one file, claim.go, for the Template Conformance Claim helper --
// PEOS-009 defines no Claim base mechanism of its own -- and doc_test.go
// asserts that no other file imports it. Nothing imports peos/template, and
// nothing should: it is a leaf.
//
// peos/lifecycle is deliberately never imported. A Template is an ordinary
// PEOS-003 Lifecycle Subject, and a Template's State Assignment "does not
// create a Template Artifact Revision; establish Template Supersession;
// establish compatibility; mutate a Template Application Record; establish
// conformance" -- so lifecycle is modeled entirely in peos/lifecycle and no
// state field appears here.
//
// peos/requirement, peos/quality, and peos/runtime are never imported either.
// A Template may generate a Requirement, a Quality Profile, or a Runtime
// Contract, but it names only their Artifact Types, never their instances, and
// "Generated Requirements remain ordinary Requirements, as defined by PEOS-005,
// subject to all of PEOS-005's rules without exception."
//
// # One additive peos/core change
//
// Packet K.1 added exactly one thing to peos/core:
// core.TemplateApplicationRecordRef, an exact reference to a Template
// Application Record. core.TemplateApplicationRecordID and the
// RecordKindTemplateApplicationRecord arm of core.RecordRef already existed;
// the dedicated Ref did not, and it is needed both as the record's own Ref()
// and as the type parameter of
// core.RecordCorrectionRef[core.TemplateApplicationRecordRef], since PEOS-009
// documents correction, replacement, and invalidation for that record family.
//
// Everything else PEOS-009 needs from peos/core was already there:
// TemplateRef, TemplateArtifactRevisionRef, GeneratedArtifactRef,
// GeneratedArtifactRevisionRef, TemplateConstraintCriterionRef with
// CriterionKindTemplateConstraint, ClaimTypeTemplateConformance,
// RelationTypeGeneratedFrom, RelationTypeTemplateComposition,
// RelationTypeTemplateSpecialization, RelationTypeArtifactSupersession, and
// EngineeringSubjectRef arms for Template, Template Revision, generated
// Artifact, and generated Artifact Revision.
//
// # Packet scope
//
// Packet K.1 implemented the Template Artifact foundation and its
// Revision-owned content: ArtifactTypeTemplate, Template, TemplateRevision,
// TemplateContent, TemplateApplicability, Parameter, ParameterType,
// ParameterDefault, ParameterConstraint with ConstraintTarget,
// CompatibilityDeclaration, the two constraint vocabularies, and both
// template-local key namespaces with their resolvers. It also added the one
// additive peos/core reference described above.
//
// Packet K.2 (this packet) completed the PEOS-009 value layer:
//
//   - ApplicationRecord, the immutable, independently identifiable,
//     non-Artifact Template Application Record, with
//     core.RecordCorrectionRef[core.TemplateApplicationRecordRef] correction and
//     self-correction rejected;
//   - ApplicationOutcome, the template-local outcome vocabulary carrying the
//     five values PEOS-009 names at minimum, together with the
//     outcome-conditional generated-output rule the record enforces
//     structurally;
//   - ResolvedValue with its ValueSource vocabulary, and GeneratedOutput, the
//     record's two owned value structures;
//   - GeneratedFrom, Composition, and Specialization, the three typed relation
//     wrappers over relation.Relation;
//   - NewTemplateConformanceClaim, the helper delegating to
//     validation.NewClaim with core.ClaimTypeTemplateConformance.
//
// Two things about the record are worth stating here rather than leaving to
// its own doc comment. First, its generated and ungenerated outputs are
// constructor arguments rather than modifiers alone: the outcome-conditional
// rule is a cross-field invariant, and with the outputs reachable only through
// modifiers every succeeded and partially-succeeded record -- precisely the two
// outcomes PEOS-009 attaches output obligations to -- would have been
// unconstructible. Second, the Conformance Claim helper adds exactly one rule
// beyond validation.NewClaim, the >=1-criterion rule PEOS-006 states for a
// Conformance Claim, because PEOS-009 makes a Template Conformance Claim a
// specialization that "inherits, without redefinition, all Validation Claim
// rules defined by PEOS-006" and peos/validation deliberately leaves
// PEOS-007/008/009 to add their own type-specific rules in their own packets.
//
// What remains for PEOS-009: Packet K.3 (Consolidated Audit) and Packet K.4
// (Closure). No further construct is scheduled.
//
// # Deliberately not modeled by this package
//
// Template Supersession is deliberately absent permanently.
// "Template Supersession reuses Artifact Supersession, as defined by PEOS-002.
// This specification does not define a separate Template Supersession entity."
// Defining a template.Supersession type would therefore contradict PEOS-009
// directly. core.RelationTypeArtifactSupersession exists for it; a general
// PEOS-002 Artifact Supersession wrapper is peos/relation's or a future
// packet's concern, not this package's debt.
//
// # Template Migration -- deferred, ontology unspecified
//
// PEOS-009 states nine things every migration SHALL identify (source and
// target Template Artifact Revisions, affected generated Artifacts or future
// applications, parameter mappings, transformation rules, information-loss
// risks, authority, provenance, and applicable Validation requirements), and
// states that migration "does not rewrite historical Template Application
// Records or previously generated Artifact Revisions".
//
// It never states what a Migration *is*. It is not declared an Artifact, not
// declared an immutable record, and not declared an Artifact Relation; it does
// not appear among the items a Template Artifact Revision SHALL identify; and
// peos/core carries no RelationTypeTemplateMigration alongside the three
// relation types it does define for PEOS-009. Each candidate ontology --
// Revision-owned value, standalone non-identity value, or a
// repository-and-Decision-owned workflow governed by PEOS-004's "Template
// generation or migration that requires governance approval SHALL use Decision
// semantics" -- is representable, and the specification does not choose.
//
// Inventing one would assert structure PEOS-009 does not define, so this
// package defines no Migration type, no MigrationRef, no migration relation
// type, no migration JSON object, and no migration sentinel. The single
// concession is CompatibilityDeclaration.MigrationRequirements(), an opaque
// descriptor satisfying the compatibility declaration's own "migration
// requirements, where applicable" scoping item without modeling Migration
// itself.
//
// Resolving this needs either a PEOS-009 amendment assigning Migration an
// ontology, or an explicit architectural decision recorded before any packet
// implements it. The tracker records the deferral; no future packet is named
// yet.
package template
