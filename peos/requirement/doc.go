// Package requirement implements PEOS-005 (Requirement Model) on top of
// the stable peos/core package (PEOS-002 Artifact Model).
//
// # Requirement is a specialization of core.Artifact, not a new identity mechanism
//
// A Requirement (PEOS-005 §6-§8) is an ordinary core.Artifact whose
// declared Artifact Type is ArtifactTypeRequirement. It introduces no
// identity of its own, no revision-history field, no lifecycle state, and
// no mutable current-value property duplicating content owned by an
// Artifact Revision (PEOS-005 §8's own prohibition). "Requirement
// Revision" is PEOS-005's own informal shorthand for "an Artifact
// Revision whose Artifact is a Requirement" (§8) — Revision in this
// package composes core.ArtifactRevision by named field, exactly per that
// shorthand, and is not a separate PEOS entity.
//
// # Typed Content is authoritative; core.Representation is optional
//
// Requirement content (Statement, Subject, Applicability, Origin,
// Authority, Classification, Rationale — PEOS-005 §8's own enumeration)
// is carried directly, in typed Go form, as Revision.Content. This typed
// content is this Revision's authoritative normative content in the
// PEOS-002 §Artifact Content sense; core.ArtifactRevision.Representations
// are optional rendered or alternate encodings of that same content (for
// example, a Markdown rendering of a Statement — PEOS-005 §9's own "MAY
// be represented using one or more Representations"). This package does
// not automatically generate a Representation from Content, and does not
// compare or reconcile Content against any Representation that happens to
// be present. A Requirement Revision MAY validly carry zero
// Representations; typed Content alone already satisfies PEOS-002's
// Artifact Content requirement.
//
// # Requirement Relationships compose relation.Relation
//
// PEOS-005 §17.4 makes every explicitly represented Requirement
// relationship "an Artifact Relation governed by PEOS-002," and this
// package's relationship wrappers (Derivation; later Refinement,
// Decomposition, Dependency, Conflict, and Requirement Supersession)
// compose peos/relation.Relation for exactly that reason, rather than
// defining a standalone record: a standalone record would restate
// PEOS-002's own Artifact Relation contract in a second place, creating
// two competing sources of truth for what an Artifact Relation is.
// peos/relation remains the sole source of truth for relation type,
// source, target, provenance, scope, and extension; each wrapper adds
// only the typed content its own PEOS-005 section requires beyond that
// (for example, Derivation's rationale).
//
// A wrapper's constructor always builds its inner relation.Relation
// itself from typed parts, and never accepts a caller-supplied
// relation.Relation: this is what makes it structurally impossible to
// construct a wrapper whose inner relation carries a different relation
// type than the one that wrapper's own type declares. UnmarshalJSON
// revalidates the decoded relation's type and participant levels for the
// same reason before applying the same validation the constructor does.
//
// PEOS-005 §17.1 gives Requirement relationships no independent identity,
// revision, lifecycle, or historical-preservation semantics of their own
// ("The identity, revision, lifecycle, historical preservation,
// authority, and representation semantics of relationships are outside
// the scope of this specification"). No relationship wrapper in this
// package carries an ID field, a Ref type, a lifecycle, or a revision
// history, and none is marshaled with an "id" JSON key -- this is safe
// specifically because peos/relation.Relation itself already carries none
// of these either (see peos/relation/doc.go).
//
// PEOS-005 §17.3 requires every relationship to make clear whether each
// participant identifies a Requirement identity or a specific Requirement
// Artifact Revision (non-conforming pattern §36.14). Each relationship
// type fixes its own permitted participant level(s): Derivation requires
// both participants at Requirement Artifact Revision level (§18.1).
// core.EngineeringSubjectRefFromRequirementRevision and
// EngineeringSubjectRef.AsRequirementRevision are the two primitives that
// make this level distinction enforceable without any peos/core change.
//
// This package does not enforce transitive cycle prohibitions (Derivation,
// Refinement, Decomposition, and Requirement Supersession cycles are all
// disallowed by PEOS-005, but only at the transitive, cross-relationship
// level): doing so would require traversing every other relationship of
// the same type across a repository this package does not hold, matching
// peos/relation/doc.go's own stance that "Cross-relation graph traversal,
// cycle detection, traceability coverage, and orphan detection" are
// explicitly assigned elsewhere. Only the direct case (a relationship's
// own source equal to its own target) is checked locally.
//
// # Derivation is deliberately not identity-bearing, unlike DecisionSupersession
//
// peos/decision.DecisionSupersession carries its own identity
// (SupersessionID), while this package's Derivation (and its sibling
// Requirement relationship wrappers) carry none. This is not an
// inconsistency between the two packages: it is required by their
// respective specifications. PEOS-004 :1080 lists "supersedes" only as
// one Decision Relation Type a Product "MAY include" among several
// listed as prose, and peos/decision.DecisionSupersession therefore
// needed its own identity so that a ConflictResolution (or another later
// record) could reference a *specific* supersession fact -- see
// peos/decision/supersession.go's own doc comment. PEOS-005 §17.4, by
// contrast, unconditionally makes every Requirement relationship an
// Artifact Relation, and §17.1 unconditionally forbids relationship
// identity. The two packages reach different, and equally correct,
// conclusions because they are answering to different normative text.
//
// # Deliberately excluded from Packet C and Packet G.1
//
// PEOS-005 defines no Requirement Acceptance Criterion, Verification
// Method, or verification status/result — every "acceptance criteria" or
// "verification" mention in PEOS-005 belongs to a future Validation Model
// (§30) that evaluates Claims *about* a Requirement Revision from outside
// it; nothing of that kind is a Requirement content field. This package
// likewise does not implement Requirement Lifecycle (PEOS-003, governed
// exclusively there per §26), Allocation (§24 -- PEOS-005 imposes no
// positive representation obligation for it; a Product MAY compose
// peos/relation.Relation with an opaque target, or use its own record),
// or Requirement Refinement, Decomposition, Dependency, Conflict,
// Requirement Supersession, or Waiver, each scheduled for a later PEOS-005
// packet (Packets G.2-G.5) on top of the foundation this file documents.
// Priority, criticality, and risk level have no normative basis anywhere
// in PEOS-005 and are not modeled at all.
//
// # Integrity scope
//
// core.IntegrityProtectedScopeContent already covers typed Requirement
// Content: PEOS-002 defines Artifact Content as "the engineering
// information represented by an Artifact Revision," and PEOS-005 §8
// defines Requirement content as precisely that information, specialized.
// No new protected scope is introduced by this package.
//
// # Subject combination is whole-sequence only
//
// SubjectCombination names one relationship for an entire ordered Subject
// sequence (independent, collective, or an open Product-defined value).
// This package does not implement boolean subject expressions,
// per-subject operators, nested subject groups, or a relationship graph
// between individual Subjects — PEOS-005 §10 asks only that the
// combination be made explicit and unambiguous for the sequence as a
// whole (see non-conforming pattern §36.22).
package requirement
