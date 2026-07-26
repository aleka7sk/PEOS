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
// type fixes its own permitted participant level(s): Derivation,
// Refinement, and Decomposition each require both participants at
// Requirement Artifact Revision level (§18.1, §19.1, §20.1), while
// Dependency and Conflict permit either level independently per
// participant, including mixed pairs (§21.1, §22.1: "each participant
// SHALL explicitly identify whether it refers to Requirement identity
// level or Requirement Artifact Revision level"). Dependency and Conflict
// express that choice with RequirementParticipant (participant.go), a
// closed union over core.RequirementRef and
// core.RequirementArtifactRevisionRef, mirroring
// decision.InvalidationSource's closed two-arm union pattern.
// core.EngineeringSubjectRefFromRequirementRevision and
// EngineeringSubjectRef.AsRequirementRevision are the two primitives that
// make the revision-only level distinction enforceable without any
// peos/core change; core.EngineeringSubjectRefFromRequirement and
// EngineeringSubjectRef.AsRequirement are the equivalent pair for the
// identity-level arm.
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
// # Mandatory vs optional scope
//
// Derivation's and Dependency's scope is optional (§18 and §21 state no
// unconditional scope requirement; §21's is conditional -- "the
// applicable scope where the reliance is not universal"), while
// Refinement's, Decomposition's, and Conflict's scope is mandatory: §19
// requires every Refinement relationship to identify "the scope within
// which compatibility is asserted," §20 requires every Decomposition
// relationship to identify "the scope of that parent-to-subordinate
// association," and §22 requires every Conflict relationship to
// identify "the scope in which the Conflict exists." Refinement,
// Decomposition, and Conflict therefore expose no WithScope or
// WithoutScope method -- WithoutScope would make an invalid, scope-less
// value reachable on a type whose scope can never be legitimately
// absent, and WithScope is unnecessary because their own constructors
// already require scope as an argument. Their Scope() accessor returns a
// bare core.Scope, not the (core.Scope, bool) shape Derivation's and
// Dependency's optional scope require: presence is guaranteed, so no
// presence flag is needed. On decode, the shared requireRelationScope
// helper (relationship.go) rejects an absent scope key explicitly,
// rather than relying on relation.Relation's own `omitempty` behavior to
// signal presence.
//
// # Derivation vs Refinement vs Decomposition vs Dependency vs Conflict
//
// Derivation (§18) records engineering origin: that required engineering
// intent was produced through engineering reasoning using other
// represented required engineering intent as an input. Refinement (§19)
// records increased precision, narrower interpretation, additional
// constraint, or greater engineering detail, while remaining compatible
// with the refined intent within an asserted scope. Decomposition (§20)
// records that one parent's required engineering intent is partitioned
// into multiple independently identified subordinate Requirements.
// Dependency (§21) records that satisfying one participant's required
// engineering intent relies on the availability, satisfaction, or
// continued validity of the other's. Conflict (§22) records that two
// participants establish incompatible required engineering intent within
// overlapping applicability, so that satisfying both simultaneously,
// within scope, is not possible. PEOS-005 explicitly rejects inferring
// any of the first three from the mere existence of another over the
// same participants (§18, §19: "neither relationship SHALL be inferred
// solely from the existence of the other" -- non-conforming pattern
// §36.15); a Derivation and a Refinement MAY validly coexist over the
// same source/target pair as two independent values with two different
// relation types.
//
// # Dependency and Conflict: direction, distinctness, and cycles
//
// Dependency's direction is semantic, not merely representational:
// Dependent depends on DependsOn (§21.1). Conflict, by contrast, is
// symmetric in meaning but representationally ordered -- §22.1 states an
// implementation's ordering "SHALL NOT imply priority, authority,
// preference, or resolution direction." ParticipantA and ParticipantB
// deliberately avoid the Source/Target naming every other relationship
// type in this package uses, matching decision.DecisionConflict's
// DecisionA/DecisionB precedent. This package does not canonicalize
// Conflict's participant order: (X, Y) and (Y, X) are both valid and are
// preserved exactly as supplied, including through JSON round-trips;
// recognizing that two stored Conflicts denote the same unordered pair
// is a repository-level concern.
//
// Dependency enforces no distinctness rule between its two participants
// -- not participant equality, and not Requirement-identity equality.
// §21.1 explicitly permits Dependency cycles, including the degenerate
// self-dependency: "Dependency cycles MAY be represented. The existence
// of a Dependency cycle SHALL NOT by itself establish that the
// participating Requirements are invalid, unsatisfiable, or
// non-conforming." A Product contract MAY additionally prohibit specific
// cycles (§21.1's own final sentence); this package does not. Conflict,
// conversely, requires its two participants to be distinct (§22: "exactly
// two distinct participants"; §22.1: "source and target SHALL be
// distinct") -- a self-conflict is rejected -- but imposes no
// Requirement-identity distinctness rule beyond that: a Requirement MAY
// conflict with a distinct Revision of the very same Requirement.
// §22.1 explicitly permits Conflict cycles among three or more distinct
// participants, and one participant MAY participate in multiple Conflict
// relationships; neither is checked by this package.
//
// # Requirement-identity distinctness: Derivation and Decomposition require it, Refinement does not
//
// PEOS-005 states an identity-distinctness requirement for two of the
// three relation types this package implements so far, using different
// wording for each, and states none at all for the third:
//
//   - Derivation (§18): "A derived Requirement SHALL possess its own
//     identity." "A derived Requirement SHALL NOT inherit the identity
//     of a source Requirement." PEOS-009 (:664, :758) is the governing
//     cross-specification precedent for what "inherit the identity"
//     means: :664 states a generated Artifact "has its own Artifact
//     identity, independent of the Template's identity," and :758 lists,
//     as the corresponding non-conforming pattern, "Representing a
//     generated Artifact as sharing, or inheriting, the Template's own
//     Artifact identity." PEOS-009 therefore equates "inheriting" an
//     identity with "sharing" it -- an equal ArtifactID. Applied to
//     §18, a Derivation whose source and target name the same
//     Requirement identity (regardless of which Revisions) shares that
//     identity, and is therefore non-conforming identity inheritance.
//   - Decomposition (§20.1): "A subordinate Requirement identity SHALL
//     remain distinct from the parent Requirement identity" -- the same
//     condition stated directly rather than through the "inherit"
//     terminology §18 uses.
//   - Refinement (§19): no equivalent clause exists. §19's own language
//     ("A refining Requirement SHALL remain independently identifiable,"
//     line 739) requires only that the refining Requirement possess its
//     own identity, not that it differ from the refined Requirement's
//     identity, and §19.1 states no distinctness rule at all.
//
// "Independently identifiable" (§18.1:709, §19:739, §20:777, §20:789) is
// not itself a distinctness clause: it appears in all three sections,
// including Decomposition, which also carries §20.1's separate explicit
// distinctness statement -- if "independently identifiable" implied
// distinctness on its own, that second statement would be redundant. It
// is common ground, not the discriminator.
//
// This package therefore rejects both a Derivation and a Decomposition
// whose two participants name different Revisions of the very same
// Requirement identity (checkDistinctRequirementIdentity, relationship.go,
// parameterized by each caller's own sentinel: ErrInvalidDerivation for
// Derivation, ErrInvalidDecomposition for Decomposition), while accepting
// the equivalent case for Refinement. This was a genuine correction to
// this package's initial Derivation implementation (Packet G.1), found
// during the normative audit that also produced this section (Packet
// G.1.1) and applied in Packet G.1.1.I: the audit concluded the rule is
// symmetric between Derivation and Decomposition, expressed in different
// PEOS-005 wording for each, and that Refinement alone is the true
// outlier -- not, as first assumed, that Decomposition alone carried the
// rule. A later Revision of a Requirement that narrows its own earlier
// wording is ordinary content change under §25, not a Derivation of a
// new Requirement from an old one; that is the substantive reason
// Refinement is permitted to relate two Revisions of one Requirement
// while Derivation and Decomposition are not.
//
// # Decomposition completeness is not modeled
//
// PEOS-005 §20.2 states that the existence of Decomposition relationships
// does not by itself establish that subordinate Requirements completely
// cover the parent's required engineering intent, that they are mutually
// exclusive, or that satisfying every subordinate establishes
// satisfaction of the parent, and explicitly forbids introducing a
// Relationship Collection, a Decomposition Set, a Completeness Assertion,
// or any other PEOS entity representing a group of Decomposition
// relationships. This package introduces none of these: a set of
// Decomposition values sharing one parent is, deliberately, nothing more
// than a set of independently valid Go values with no group type, no
// shared identity, and no completeness semantics of its own. One parent
// MAY be the source of multiple Decomposition relationships, and one
// subordinate MAY be the target of more than one (§20.1); distinguishing
// their scopes when that occurs is a repository-level concern this
// package does not check.
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
// # Deliberately excluded from Packets C, G.1, G.2, and G.3
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
// or Requirement Supersession or Waiver, each scheduled for a later
// PEOS-005 packet (Packets G.4-G.5) on top of the foundation this file
// documents. Priority, criticality, and risk level have no normative
// basis anywhere in PEOS-005 and are not modeled at all.
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
