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
// # Deliberately excluded from Packet C
//
// PEOS-005 defines no Requirement Acceptance Criterion, Verification
// Method, or verification status/result — every "acceptance criteria" or
// "verification" mention in PEOS-005 belongs to a future Validation Model
// (§30) that evaluates Claims *about* a Requirement Revision from outside
// it; nothing of that kind is a Requirement content field. This package
// likewise does not implement Requirement Lifecycle (PEOS-003, governed
// exclusively there per §26), Requirement Relations (Derivation,
// Refinement, Decomposition, Dependency, Conflict, Supersession — each an
// Artifact Relation per §17.4, and Artifact Relations are not yet
// implemented in core), Allocation (§24), or Waiver (§27, requires a
// Decision Outcome that does not yet exist). Priority, criticality, and
// risk level have no normative basis anywhere in PEOS-005 and are not
// modeled at all.
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
