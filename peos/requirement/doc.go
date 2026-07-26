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
// Refinement, Decomposition, and Supersession each require both
// participants at Requirement Artifact Revision level (§18.1, §19.1,
// §20.1, §23.1: "source and target SHALL be identified at Requirement
// Artifact Revision level while their owning Requirement identities
// remain identifiable"), while Dependency and Conflict permit either
// level independently per participant, including mixed pairs (§21.1,
// §22.1: "each participant SHALL explicitly identify whether it refers to
// Requirement identity level or Requirement Artifact Revision level").
// Dependency and Conflict express that choice with RequirementParticipant
// (participant.go), a closed union over core.RequirementRef and
// core.RequirementArtifactRevisionRef, mirroring
// decision.InvalidationSource's closed two-arm union pattern.
// RequirementParticipant is deliberately NOT used by Supersession: §23.1
// fixes both of its participants at Requirement Artifact Revision level
// with no either-level choice, so accepting a RequirementParticipant there
// would make an identity-level participant constructible in a place
// PEOS-005 forbids it.
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
// Refinement's, Decomposition's, Conflict's, and Supersession's scope is
// mandatory: §19 requires every Refinement relationship to identify "the
// scope within which compatibility is asserted," §20 requires every
// Decomposition relationship to identify "the scope of that
// parent-to-subordinate association," §22 requires every Conflict
// relationship to identify "the scope in which the Conflict exists," and
// §23 requires every Supersession relationship to identify "the scope
// within which that replacement applies." Refinement, Decomposition,
// Conflict, and Supersession therefore expose no WithScope or
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
// # Supersession is scoped replacement, not revision history
//
// Supersession's direction is inverted relative to every other
// relationship type in this package: source is the superseding (newer)
// participant, target is the superseded (older) participant (§23.1),
// whereas Derivation, Refinement, and Decomposition all put the
// originating participant at source. Superseding()/Superseded() name
// this directly, matching decision.DecisionSupersession's
// SupersedingDecision/SupersededDecision and
// lifecycle.LifecycleDefinitionVersionSupersession's
// SupersedingVersion/SupersededVersion precedents.
//
// Unlike Derivation and Decomposition, Supersession does NOT require
// Superseding and Superseded to name distinct Requirement identities:
// REQ-1/REV-2 superseding REQ-1/REV-1 is a valid Supersession. Every
// PEOS-005 clause touching Supersession's identity effect is worded as
// preservation, never distinctness -- §7 ("Requirement supersession SHALL
// NOT merge Requirement identities"), §23 ("Superseded Requirements SHALL
// retain their identities"; "Supersession SHALL NOT merge Requirement
// identities"), §28.3 ("Requirement identity SHALL remain preserved"),
// and §35 ("Supersession SHALL NOT merge or destroy Requirement
// identities") -- in contrast to §18's "SHALL NOT inherit the identity of
// a source Requirement" and §20.1's "SHALL remain distinct," both of
// which state distinctness directly (see the "Requirement-identity
// distinctness" section below for the full comparison). "Merge" denotes
// collapsing two identities into one; a Supersession whose two
// participants already share one identity merges nothing -- the identity
// count is one before and one after. §23's replacement is scoped (§23.2),
// which linear revision history alone cannot express: "REV-2 replaces
// REV-1 within scope S, while REV-1 remains applicable outside S" is
// engineering meaning §25's ordinary content-change model has no way to
// carry. checkDistinctRequirementIdentity is therefore NOT used by
// Supersession -- only checkDistinctParticipants, rejecting direct
// self-supersession (source == target).
//
// Supersession SHALL NOT be inferred solely from creation of a newer
// Artifact Revision, Statement similarity, document or identifier order,
// Lifecycle State, or archival status (§23; non-conforming pattern
// §36.13, "Implicit Supersession"). This package prevents that
// structurally rather than by convention: a Supersession value cannot be
// constructed without an explicit governance action (GovernanceAction),
// an explicit scope, and explicit provenance, none of which a bare
// Artifact Revision carries -- a same-identity Supersession still
// requires all three, exactly like a different-identity one.
//
// # Supersession records a replacement fact; it does not change Lifecycle state
//
// PEOS-005 §26.5 keeps Requirement Supersession and Lifecycle State
// Assignment deliberately separate: "Assignment of [a Superseded
// Lifecycle State] records governance state only. A State Assignment
// SHALL NOT by itself establish: which Requirement supersedes another
// Requirement; which Requirement Artifact Revisions are involved; the
// scope of replacement; the authority or governance action establishing
// replacement... The State Assignment and the applicable Supersession
// relationship SHALL remain semantically distinct and independently
// inspectable." NewSupersession only ever creates an immutable
// relationship record: it does not read, require, or validate any
// Lifecycle State, and this package does not import peos/lifecycle at
// all -- that non-import is the structural guarantee of §26.5's
// separation, not merely a documented intention.
//
// LifecycleConsequence (supersession.go) is Supersession's own declared
// "resulting effective status or Lifecycle consequence, if any" (§23.1).
// It is a closed two-state discriminator -- identified (with a mandatory,
// trimmed description) or none (an explicitly constructed, non-zero
// declaration of no consequence) -- whose zero value is invalid and
// represents a third, unstated state §23.1 does not permit ("Absence of
// a Lifecycle consequence SHALL be explicitly representable and SHALL
// NOT invalidate an otherwise established Supersession relationship").
// This mirrors Applicability's unrestricted/scoped shape (this file, and
// content.go) and decision.SupersessionExtent's complete/partial shape,
// both of which reject their own zero value for the same reason: a
// *string or (string, bool) encoding cannot distinguish
// absent-because-declared from absent-because-unset.
// LifecycleConsequence is a declaration recorded inside the relationship,
// never a Lifecycle State Assignment and never a reference to one --
// treating a Superseded Lifecycle State Assignment as sufficient to
// establish the replacement fact is exactly non-conforming pattern
// §36.12 ("Lifecycle-Only Supersession"). Where a Product wants both a
// Supersession relationship and a Lifecycle State Assignment, they are
// two independent records a higher layer may create atomically; that
// coordination is a repository/application concern, not one this value
// type performs.
//
// §23.2's applicability boundary is likewise kept out of this package:
// whether a superseded Requirement remains normatively applicable outside
// (or even within) the Supersession's scope depends on its own
// Applicability conditions, Lifecycle State, and Authority -- evaluation
// this package does not perform. A Supersession's existence never by
// itself makes its superseded participant universally non-applicable.
//
// # Requirement-identity distinctness: Derivation and Decomposition require it; Refinement, Dependency, Conflict, and Supersession do not
//
// PEOS-005 states an identity-distinctness requirement for exactly two of
// the six relation types this package implements, using different
// wording for each, and states none at all for the other four:
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
//   - Dependency (§21/§21.1) and Conflict (§22/§22.1): neither section
//     states an identity-distinctness rule of any kind. Conflict requires
//     only participant-shape distinctness (its two participants must not
//     be the same RequirementParticipant value); Dependency requires no
//     distinctness at all, permitting even self-dependency.
//   - Supersession (§23/§23.1): no distinctness clause exists here either
//     -- every clause touching Supersession's identity effect (§7, §23,
//     §28.3, §35) is worded as preservation ("SHALL NOT merge," "SHALL
//     retain," "SHALL remain preserved"), never as distinctness, in
//     contrast to §18's and §20.1's direct distinctness language. See
//     "Supersession is scoped replacement, not revision history" above
//     for the full argument.
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
// the equivalent case for Refinement, Dependency, Conflict, and
// Supersession. checkDistinctRequirementIdentity is used by exactly two
// of this package's six relationship types; the other four rely only on
// checkDistinctParticipants (or, for Dependency, no distinctness check at
// all). This was a genuine correction to this package's initial Derivation
// implementation (Packet G.1), found during the normative audit that also
// produced this section (Packet G.1.1) and applied in Packet G.1.1.I: the
// audit concluded the rule is symmetric between Derivation and
// Decomposition, expressed in different PEOS-005 wording for each, and
// that Refinement alone was the outlier among the three types then
// implemented -- Packet G.4's Supersession analysis later confirmed
// Refinement was not a special case at all, but the first instance of the
// more general rule that a section must state distinctness directly for
// it to apply; every relation type added since (Dependency, Conflict,
// Supersession) states no such clause and is treated accordingly. A later
// Revision of a Requirement that narrows its own earlier wording is
// ordinary content change under §25, not a Derivation of a new Requirement
// from an old one; that is the substantive reason Refinement (and,
// differently, Supersession) permit relating two Revisions of one
// Requirement while Derivation and Decomposition do not.
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
// # GovernanceAction is shared foundation, not Supersession-specific
//
// GovernanceAction (governance.go) identifies the governance action under
// which a Supersession's replacement was established (§23: "a governance
// action is an established Decision Outcome or another governance
// mechanism explicitly permitted by an applicable PEOS Product contract";
// "Governance action is a semantic role and SHALL NOT be interpreted as
// introducing a separate PEOS entity" -- hence no identity, no lifecycle,
// no Ref type). §27 defines Waiver's authorizing action in identical
// terms, so this type is deliberately package-level shared foundation,
// not folded into supersession.go. Packet G.5.I confirmed this reuse
// directly: Waiver (waiver.go) consumes GovernanceAction unmodified, with
// no new arm and no Waiver-specific equivalent. Unlike RequirementParticipant,
// GovernanceAction carries JSON: it is a type-specific field stored
// alongside, not inside, a composed relation.Relation -- and, for Waiver,
// alongside no relation.Relation at all, since Waiver composes none.
//
// # Waiver is a governance value record, not a relationship
//
// Waiver (waiver.go) records an authorized, scoped limitation of a
// Requirement's normative applicability (§27: "A Requirement MAY be
// waived only through applicable engineering governance... A waiver
// suspends or limits normative applicability within its defined scope.").
// It is a separate immutable governance value record -- not an Artifact,
// not an Artifact Revision, not an Artifact Relation, not a Lifecycle
// State Assignment, and not a Decision Outcome. It carries no identity of
// its own (no ArtifactID, no RevisionID, no WaiverID or WaiverRef),
// matching PEOS-008 :520's own statement that PEOS-005 "does not define
// Waiver identity, Waiver lifecycle, Waiver revocation, or a Waiver
// historical model."
//
// Waiver composes no relation.Relation, unlike this package's six
// Requirement relationship wrappers: §17.4 binds every explicitly
// represented Requirement *relationship* to PEOS-002's Artifact Relation
// contract, but §27 never calls a waiver a relationship, and §17.1
// governs relationship identity, lifecycle, and history -- not Waiver's.
// A Waiver therefore carries no "relation" JSON key, no RelationType, and
// no Source/Target: it names its single waived Requirement directly, at
// Requirement identity level (core.RequirementRef), never at Requirement
// Artifact Revision level -- §27 says "the Requirement" five times and
// never once says "Requirement Artifact Revision," in direct contrast to
// the six relationship types' own §18.1/§19.1/§20.1/§21.1/§22.1/§23.1
// participant-level clauses. RequirementParticipant (participant.go) is
// deliberately not reused here for the same reason it is not reused by
// Supersession: it exists so Dependency and Conflict can express an
// either-level choice §21.1/§22.1 explicitly require, and §27 asks no
// such question.
//
// Waiver mandates exactly four fields, one per §27/§27.1/§27.2 "SHALL"
// obligation: the waived Requirement (§27), the authority under which the
// waiver is established (§27.1, core.AuthorityRef), the governance action
// through which it was established (§27, GovernanceAction, reused
// unchanged), and the waiver's scope (§27.2, core.Scope, mandatory --
// no WithScope/WithoutScope, matching Refinement's, Decomposition's,
// Conflict's, and Supersession's mandatory-scope treatment above). All
// four are required constructor arguments; extension is the only optional
// field. Waiver deliberately carries no core.Provenance field: unlike
// this package's relationship wrappers, which inherit provenance from
// relation.Relation as PEOS-002's own Artifact Relation obligation, §27
// imposes no equivalent provenance requirement of its own, and Waiver
// has no relation.Relation to inherit one from.
//
// # Waiver is not a Requirement edit, an Applicability change, a Lifecycle state, or a Decision Outcome
//
// A waiver SHALL NOT delete the Requirement, change its identity, or
// rewrite its content (§27); structurally, Waiver holds a
// core.RequirementRef -- a reference -- never a Requirement or Revision
// value, so none of those three prohibited effects is representable here.
// This is also the structural prevention of non-conforming pattern
// §36.11 ("Representing a waiver as an informal flag or mutable
// Requirement attribute," for example Requirement{waived = true}):
// Requirement and Content gain no field for this package's Waiver support,
// since Waiver is a wholly separate record referencing a Requirement, not
// an attribute of one.
//
// Changing what a Requirement's content or Applicability says requires an
// ordinary new Artifact Revision under §25; a Waiver produces none, and
// Applicability (content.go, Revision-owned) and Waiver are independent
// types with no field or method connecting them. A Waiver likewise
// carries no Lifecycle State and does not import peos/lifecycle: §26
// governs Lifecycle exclusively there, and §27 defines no Lifecycle
// interaction for this package to model -- mirroring Supersession's own
// independence from peos/lifecycle (see "Supersession records a
// replacement fact" above). Finally, a GovernanceAction built from a
// core.DecisionOutcomeRef is an authorization *reference* the Waiver
// carries, not the Waiver itself, and not the Decision Outcome record:
// the authorizing Decision Outcome and the Waiver it authorizes remain
// two distinct records.
//
// # What Waiver deliberately does not model
//
// No waived-obligation reference: §27 contains no clause requiring an
// explicitly represented waived obligation, criterion, quality
// constraint, or rule distinct from the waived Requirement itself. The
// narrowing mechanism §27 provides is scope, not a separate obligation
// field -- §27: "A waiver suspends or limits normative applicability
// within its defined scope."
//
// No justification, reason, rationale, or compensating-control field:
// §27/§27.1/§27.2 mandate none, and §36.11's non-conformance triad is
// scope + authority + governance only. A Product needing rationale prose
// uses extension, which stays structurally separate from PEOS-owned
// fields.
//
// No temporal fields (no effective-from, expiry, duration, or review-date):
// §27 has no such clause. "Defined periods of applicability" appears only
// inside §27.2's own "Scope MAY include" list, so PEOS-005 deliberately
// carries any temporal bound inside scope's opaque expression, not as a
// separate typed field this package would need to interpret.
//
// No lifecycle, expiry, revocation, or status field: PEOS-008 :520 is
// decisive that PEOS-005 defines none of these for Waiver. Active/expired
// is a runtime evaluation derived from scope (PEOS-008's own layer);
// revoked/withdrawn is not a PEOS-005 construct -- a Product revokes by
// recording a new governance action, never by mutating the original
// Waiver value, which carries no mutator for any mandatory field.
//
// No collection or graph type: consistent with §17.1/§20.2/§22.1's
// repeated refusal to reify relationship collections, this package
// introduces no WaiverSet. A Requirement MAY carry multiple Waivers;
// recognizing duplicate, overlapping, or conflicting Waivers is a
// repository-level concern this package does not check.
//
// # Deliberately not modeled by this package
//
// PEOS-005 defines no Requirement Acceptance Criterion, Verification
// Method, or verification status/result — every "acceptance criteria" or
// "verification" mention in PEOS-005 belongs to a future Validation Model
// (§30) that evaluates Claims *about* a Requirement Revision from outside
// it; nothing of that kind is a Requirement content field. This package
// likewise does not implement Requirement Lifecycle (PEOS-003, governed
// exclusively there per §26), or Allocation (§24 -- PEOS-005 imposes no
// positive representation obligation for it; a Product MAY compose
// peos/relation.Relation with an opaque target, or use its own record).
// All six Requirement Relation Types (§17-§23) and Waiver (§27) are
// implemented elsewhere in this package (Packets G.1-G.5.I); none of
// those is deliberately excluded any longer. Priority, criticality, and
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
