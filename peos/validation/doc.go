// Package validation implements PEOS-006 (Validation Model) on top of the
// stable peos/core package (PEOS-002 Artifact Model).
//
// Statements below are labelled to keep four different kinds of claim
// distinguishable: [NORMATIVE] restates a PEOS-006 requirement;
// [IMPLEMENTATION] records a choice this SDK made where PEOS-006 leaves
// room; [PRODUCT] assigns an obligation to an applicable Product contract,
// repository, or runtime; [DEFERRED] names something a later packet owns.
//
// # Scope of this package today
//
// This package implements the PEOS-006 constructs the specification defines:
// Plan, PlanRevision, PlanContent, PlanApplicability, and PlannedActivity
// (Packet H.1), plus ActivityReference, ExecutionEvent, ExecutionRecord, and
// Claim (Packet H.2). core.ExecutionOutcome, added by H.2, lives in peos/core
// alongside its four sibling PEOS-006 vocabularies.
//
// [NORMATIVE] PEOS-006 keeps four concepts strictly distinct: "A Validation
// Plan, a Planned Validation Activity, an Execution Record, and a
// Validation Claim are four distinct concepts. None is a substitute for
// another." All four are present here as separate types, and no one of them
// can stand in for another.
//
// [DEFERRED] Not implemented here, by design: PEOS-007's Quality Claim
// specialization and Measurement Record, PEOS-008's Compliance Claim and all
// runtime interpretation, and PEOS-009's Template Conformance Claim. Each of
// those specializes mechanisms this package owns rather than adding its own
// -- PEOS-006 states that "PEOS-007 does not define an independent Activity,
// Evidence, or Claim mechanism of its own" -- so the Claim Type values exist
// in peos/core already, while any additional rules those specifications
// impose belong to their own packets and are deliberately not inferred here.
// Also deferred: the criterion-level Waiver question and Waiver attached
// conditions, both of which are PEOS-005 concerns (see "Waiver never rewrites
// a record" below).
//
// # Validation Plan is an Artifact; it has no Version system
//
// [NORMATIVE] "A Validation Plan is an Artifact as defined by PEOS-002."
// Plan is therefore a thin specialization of core.Artifact whose declared
// Artifact Type is ArtifactTypeValidationPlan, exactly as
// requirement.Requirement specializes core.Artifact for PEOS-005.
//
// [NORMATIVE] "A Validation Plan uses ordinary Artifact Revision for all of
// its evolution. There is no Validation Plan Version distinct from Artifact
// Revision." No type in this package carries a version field, and
// PlanRevision composes core.ArtifactRevision rather than defining a
// parallel revisioning mechanism (non-conforming patterns "Validation Plan
// Version" and "No Parallel Revision System Invariant").
//
// [NORMATIVE] "Modification of a Validation Plan's content constitutes a
// content change and SHALL create a new Artifact Revision." PlanRevision is
// immutable and exposes no WithContent; a changed Plan means a new
// PlanRevision carrying a new core.ArtifactRevision.
//
// [NORMATIVE] "Creation of a new Validation Plan Revision does not mutate,
// delete, or reinterpret a previous Validation Plan Revision." Execution
// history is therefore never accumulated onto a Plan Revision. A Validation
// Execution Record will reference the exact Plan Revision and plan-local
// key it executed (Packet H.2); the referenced Plan Revision gains no field
// recording that. Any "executions so far" list on PlanContent would require
// mutating a recorded Revision and is structurally absent.
//
// [IMPLEMENTATION] ArtifactTypeValidationPlan's exact vocabulary string,
// "peos:validation-plan", is a choice: PEOS-006 does not fix one. It is
// namespaced under core.PEOSNamespace because Validation Plan is a
// PEOS-defined Artifact Type, matching requirement.ArtifactTypeRequirement.
// It lives here rather than in peos/core for the same reason
// ArtifactTypeRequirement lives in peos/requirement.
//
// # Planned Validation Activity is Revision-owned and has no identity
//
// [NORMATIVE] "A Planned Validation Activity is an Artifact Revision-owned
// value structure" that "belongs to exactly one Validation Plan Revision"
// and "has no independent PEOS identity. It has no revisions and no
// lifecycle of its own; its evolution is entirely governed by the revision
// of its owning Validation Plan."
//
// PlannedActivity accordingly carries no ID type, no Ref type, no
// core.Artifact, no core.ArtifactRevision, no lifecycle field, and no
// provenance field of its own -- provenance belongs to the owning
// PlanContent, which records the origin of the Plan Revision as a whole.
// Its wire form has no "id" key and no revision identity. This structurally
// prevents the non-conforming patterns "Decision-Like Validation Activity",
// "Planned Activity with Global Identity", and "Lifecycle-Governed
// Validation Activity".
//
// [NORMATIVE] "Every Planned Validation Activity SHALL have a stable
// plan-local key," which "is unique only within the owning Validation Plan
// Revision", "does not survive as an independent identity outside that exact
// Plan Revision", "is not an Artifact Identity", and "is not a global
// Validation Activity Identity." The key is a core.LocalKey. NewPlanContent
// rejects a repeated key within one PlanContent (ErrDuplicatePlanLocalKey),
// which is the Plan-Local Key Invariant, and rejects a dependency naming a
// key no Activity in the same content defines (ErrUnknownPlanLocalKey),
// because such a key cannot denote anything.
//
// [NORMATIVE] "A new Validation Plan Revision MAY reuse, remove, or
// reintroduce a plan-local key. A plan-local key from an earlier Plan
// Revision SHALL NOT be assumed to refer to the same Planned Validation
// Activity in a later Plan Revision unless the applicable Product contract
// explicitly defines that continuity." This package never compares keys
// across Plan Revisions and offers no API that would suggest cross-Revision
// key identity. [PRODUCT] Any such continuity is a Product contract's to
// define.
//
// # An Activity has exactly one Subject; criteria are not Subjects
//
// [NORMATIVE] Every Planned Validation Activity identifies its Subject "and
// exact participant level" (Validation Subject Level Invariant).
// PlannedActivity holds exactly one core.EngineeringSubjectRef -- a single
// value, not a slice -- and participant level is carried by that type's own
// arm rather than by a separate level field, so a decoded Subject always
// states an exact level and the two can never desynchronize.
//
// [NORMATIVE] "A criterion is not a second Claim Subject" (Claim Criterion
// Separation Invariant; non-conforming pattern "Criterion Treated as
// Subject"). Criteria are core.CriterionRef values. That type and
// core.EngineeringSubjectRef are distinct Go types with no conversion path
// in either direction -- peos/core establishes this deliberately -- so a
// criterion cannot be placed in a Subject field. The guarantee is
// structural, not a runtime check.
//
// [PRODUCT] Choosing the correct level for what is actually being evaluated
// is semantic and cannot be checked here: "Where content is being
// evaluated, the Subject SHALL be identified at the Artifact Revision or
// Requirement Artifact Revision level."
//
// # Validation Method is a vocabulary value, not an Artifact identity
//
// [NORMATIVE] "A Validation Method is a controlled vocabulary/type," and it
// "does not require independent Artifact identity. There is no Validation
// Method Version." An Activity's Method is a core.ValidationMethod, an open
// namespaced vocabulary value.
//
// [NORMATIVE] "An external or PEOS Artifact MAY define detailed procedure
// content for a Validation Method." PlannedActivity's optional
// methodDefinition is a core.ArtifactRevisionRef expressing exactly that.
// It is a reference to a definition and never the Method's identity; the
// backing Artifact's own Artifact Revision is the only revisioning
// involved, so no Method revision system is introduced (non-conforming
// pattern "Validation Method Version").
//
// [NORMATIVE] "Certification and Acceptance are not Validation Methods.
// They are governance outcomes governed by PEOS-004, established through a
// Decision Outcome that MAY reference one or more Validation Claims"
// (non-conforming patterns "Certification as Validation Method" and
// "Acceptance as Validation Method"). This package predeclares no Method
// constants at all, and in particular must never predeclare certification
// or acceptance. [PRODUCT] Because the Method vocabulary is explicitly open
// ("This list is illustrative and is not a closed vocabulary"), a Product
// declaring its own certification-shaped Method value cannot be prevented
// structurally; conforming to this prohibition is a Product obligation.
//
// [IMPLEMENTATION] PEOS-006 defines no Validation Method parameter model,
// so this package defines none. Product-specific method parameters belong
// in PlannedActivity's core.Extension; inventing typed parameters would
// amount to a method DSL the specification does not define.
//
// # PlanApplicability is mandatory, explicit, and validation-local
//
// [NORMATIVE] PEOS-006 lists "applicability" among the items every
// Validation Plan Revision SHALL identify, with no qualifier -- in
// deliberate contrast to the two bullets immediately before it, which are
// qualified "where applicable" and "where required by the applicable
// Product contract". NewPlanContent therefore takes it as a mandatory
// argument, and PlanContent exposes no WithApplicability or
// WithoutApplicability.
//
// [IMPLEMENTATION] PlanApplicability is a closed two-state discriminator
// whose zero value is invalid and means unstated.
// NewUnrestrictedPlanApplicability constructs "no restriction" as a
// distinct, non-zero value, which is what makes an explicit unrestricted
// applicability distinguishable from an unstated one. A *core.Scope or a
// bare (core.Scope, bool) pair cannot express that distinction, which is
// why neither is used -- the same reasoning requirement.Applicability and
// requirement.LifecycleConsequence already apply.
//
// [IMPLEMENTATION] PlanApplicability is deliberately NOT
// requirement.Applicability, and no conversion exists between them. That
// type is PEOS-005 §11 Requirement content, describing when a Requirement's
// required engineering intent applies; this one describes when a Validation
// Plan Revision applies. They answer different questions for different
// owning specifications, and reusing the PEOS-005 type here would require
// importing peos/requirement, which this package's import boundary forbids.
// The two-arm shape is duplicated deliberately; the concept is not shared.
//
// [PRODUCT] This package never interprets core.Scope's Expression, for
// either scope or applicability.
//
// # Mandatory state is supplied to constructors, never completed later
//
// [IMPLEMENTATION] Every mandatory field of every type here is a required
// constructor argument. No With* method sets a mandatory field, and no
// Without* method exists for one. A value returned by a constructor is
// always fully valid, and no sequence of public calls can produce a
// partially established value.
//
// PlanContent therefore has no WithScope, WithoutScope, WithApplicability,
// WithoutApplicability, WithProvenance, WithoutProvenance, WithActivities,
// or WithoutActivities: all four of its constructor fields are mandatory
// aggregate state, and replacing any of them means constructing a new
// PlanContent -- which, per PEOS-006, belongs to a new Artifact Revision
// anyway. PlannedActivity likewise has no WithKey, WithSubject, WithMethod,
// or WithOutcomeInterpretation.
//
// [IMPLEMENTATION] Optional collections use a single replace-the-whole-
// collection modifier and no paired Without* method: WithCriteria(nil)
// already expresses removal, and a second method would create a second
// validation path for one field. Where absence can be invalid, this package
// omits the Without* method entirely rather than offering one that
// sometimes fails -- the convention peos/requirement already follows by
// giving Refinement, Decomposition, Conflict, Supersession, and Waiver no
// WithoutScope.
//
// # Empty collections mean "none declared", where PEOS-006 permits zero
//
// [IMPLEMENTATION] The distinction this package applies throughout: a field
// is mandatory when its zero value would be ambiguous between "unstated"
// and a legitimate value, and optional when its zero value already
// unambiguously means "none".
//
// PlanApplicability and an Activity's outcomeInterpretation are mandatory
// on that basis. An Activity's criteria, expected evidence, prerequisites,
// and dependencies are optional on the same basis: an empty collection is
// an unambiguous "none declared", there is no third unstated state to
// guard against, and PEOS-006 contemplates zero-criteria evaluation
// elsewhere ("Where zero criteria are identified, the Claim's outcome SHALL
// be interpreted strictly according to its stated Validation Method and
// basis"). core.Extension's own contract reasons identically about its zero
// value.
//
// [IMPLEMENTATION] For those four optional collections, an absent JSON key,
// an explicit null, and an empty array are treated as equivalent, because
// all three denote the same valid state. This is deliberately not how a
// Validation Claim's criteria behave, because a Claim Type may forbid the
// empty case and the three inputs therefore carry different meanings and are
// told apart -- see claim.go. Optional single-value keys here
// (responsible_role, required_authority, method_definition,
// acceptance_rules) do reject an explicit null, following
// lifecycle.StateAssignment's treatment of its optional Authority; the one
// exception is extension, where core.Extension's documented contract makes
// null equivalent to absent. Each type's UnmarshalJSON doc comment states
// its actual missing-versus-null behavior field by field rather than
// asserting the two are identical.
//
// [IMPLEMENTATION] PEOS-006 states no uniqueness rule for an Activity's
// criteria, expected evidence, or prerequisites, so this package never
// deduplicates them and preserves caller order exactly. [PRODUCT]
// Recognizing a repeated entry is a repository or Product concern.
//
// # Plan-level acceptance rules do not duplicate per-Activity interpretation
//
// [NORMATIVE] A Plan Revision SHALL identify "the acceptance or evaluation
// rules applicable to those Activities", unqualified.
//
// [IMPLEMENTATION] That obligation is discharged by PlannedActivity's
// mandatory outcomeInterpretation, which states per Activity "how its
// expected outcome is to be interpreted". PlanContent's optional
// acceptanceRules exists for Plan-wide rules belonging to no single
// Activity; it neither duplicates nor overrides per-Activity outcome
// interpretation.
//
// # Dependencies: resolution is checked, cycles are not
//
// [NORMATIVE] A Planned Validation Activity identifies "its sequencing or
// dependencies, where applicable", and a plan-local key "is used by
// sequencing, dependencies, Execution Records, required-evidence rules, and
// Claim basis to reference the Planned Validation Activity".
//
// NewPlanContent enforces that every declared dependency key resolves to an
// Activity in the same PlanContent, order-independently: the full key set is
// collected before any dependency is checked, so an Activity may depend on
// one appearing later in the list.
//
// [IMPLEMENTATION] PEOS-006 states no cycle policy and no self-reference
// prohibition for Activity dependencies. A self-dependency and a dependency
// cycle are therefore both accepted: rejecting them would enforce a rule
// the specification does not state, and PEOS-005 §21.1 shows PEOS is
// willing to permit cycles explicitly where it means to. [PRODUCT]
// Dependency and sequencing semantics, including any cycle prohibition, are
// repository- or Product-owned. This package provides no graph traversal
// and no cycle-detection API.
//
// # Execution Record and Claim are immutable records with their own identity
//
// [NORMATIVE] "A Validation Execution Record is an immutable record... is
// independently identifiable... is not an Artifact. It is not revisioned. It
// is not lifecycle-bearing." The same four properties hold for a Validation
// Claim, which additionally "is never represented as an Artifact, an Artifact
// Revision, a Requirement, or a State Assignment" and "preserves its
// historical assertion permanently."
//
// ExecutionRecord and Claim therefore carry core.ValidationExecutionRecordID
// and core.ValidationClaimID as their own identities, compose no
// core.Artifact and no core.ArtifactRevision, expose no Core accessor, and
// carry no revision, lifecycle, or status field. Immutability is structural
// rather than conventional: every field is unexported, every modifier returns
// a copy, and no modifier touches a mandatory field.
//
// [NORMATIVE] A Claim "has exactly one engineering subject and zero or more
// criteria", and "SHALL NOT identify more than one Subject." The subject field
// is a single value, not a slice, so the non-conforming "Composite Claim
// Subject" pattern is unrepresentable. Where an evaluation concerns several
// entities jointly, PEOS-006 requires multiple Claims, each with one Subject.
//
// [NORMATIVE] Criteria are never additional subjects (Claim Criterion
// Separation Invariant). This is enforced by the type system: core.CriterionRef
// and core.EngineeringSubjectRef are distinct types with no conversion path in
// either direction, so a criterion cannot be placed in a subject field.
//
// [NORMATIVE] A Claim "cites one or more Evidence Artifact Revisions", so
// evidence is mandatory with at least one citation -- a deliberate contrast
// with "zero or more Criteria" in the same normative block. Every citation is
// a core.EvidenceArtifactRevisionRef, which exists only at exact Revision
// level, so the "Evidence Without Exact Revision" pattern is unrepresentable.
//
// # Execution outcome and Claim outcome are different things
//
// [NORMATIVE] An ExecutionRecord's core.ExecutionOutcome (completed / failed /
// interrupted / indeterminate) says whether the activity ran. A Claim's
// core.ClaimOutcome (satisfied / not satisfied / inconclusive) says what was
// determined about the Subject. They are separate vocabularies over disjoint
// value sets and separate Go types; a completed execution may back a
// not-satisfied Claim, and PEOS-006 requires neither record to imply the
// other's existence at all.
//
// [NORMATIVE] Neither outcome is a Lifecycle State: execution outcomes "are
// not Lifecycle States, and they SHALL NOT be represented as Lifecycle States
// or through a State Assignment" (non-conforming pattern "Execution Outcome as
// Lifecycle State").
//
// [NORMATIVE] There is no separate Verdict entity, and no verdict field:
// "The outcome recorded on a Validation Claim is the complete and only
// representation of what was determined; it is not restated or re-derived
// through a second construct."
//
// [NORMATIVE] There is no basis field. "Claim Basis is not an independent
// opaque field distinct from the fields it groups. This specification does not
// require an additional required field named 'basis' beyond the individually
// identified method, criteria, Evidence, Execution Records, and reasoning."
// All five are present individually; "Claim Basis" remains a collective name.
//
// [NORMATIVE] No Observation, Result, Finding, or Violation entity is
// introduced. "An Observation or a Result is not a separate identity-bearing
// category distinct from Evidence. A recorded observation or measurement
// either is represented as an Evidence Artifact, or it is represented as
// content within an immutable Validation Execution Record. No third category
// is introduced." Finding and Violation appear nowhere in PEOS-006; Violation
// belongs to PEOS-008.
//
// # Satisfaction and Conformance are Claim Type values, not types
//
// [NORMATIVE] PEOS-006 defines a Satisfaction Claim as "a Validation Claim
// whose criteria identify one or more Requirements" and a Conformance Claim as
// "a Validation Claim whose criteria identify one or more" conformance rules.
// Both are the same record under a constraint, so one Go type (Claim) covers
// every specialization, discriminated by core.ClaimType. peos/core declares a
// single identity space for all of them.
//
// [IMPLEMENTATION] The Claim-Type-conditional criteria rules live in one
// internal validation path shared by NewClaim, WithCriteria, and (through
// NewClaim) UnmarshalJSON. A second copy could drift, and the rules couple two
// fields, so they are re-checked whenever either changes. This is also why
// criteria is a NewClaim argument rather than a later With* call: a
// Satisfaction Claim needs a Requirement criterion to be valid at all, so a
// Claim completed afterward would necessarily pass through an invalid state.
//
// The enforced rules, per Claim Type:
//
//   - satisfaction: at least one criterion identifying a Requirement or a
//     Requirement Artifact Revision, and -- if the Subject itself identifies a
//     Requirement or Requirement Artifact Revision -- a Requirement identity
//     differing from that of every Requirement-kind criterion. "A Requirement
//     SHALL NOT become both the Claim subject and the same Claim's criterion."
//   - conformance: at least one criterion, of any kind.
//   - quality, compliance, template-conformance, and any Product-defined Claim
//     Type: no additional rule; zero criteria are accepted.
//
// [IMPLEMENTATION] The Satisfaction identity comparison is at
// core.ArtifactID and is cross-level, because "the same Requirement" is
// identity-level language: an identity-level subject conflicts with a
// revision-level criterion of that Requirement, a revision-level subject
// conflicts with an identity-level criterion of it, and two different
// Revisions of one Requirement still conflict.
//
// [NORMATIVE] PEOS-006 carves out the converse explicitly: "This does not
// prohibit every Claim whose subject is a Requirement. A general Validation
// Claim MAY evaluate a Requirement as an engineering subject for other
// purposes, such as statement quality, completeness, consistency, or
// conformance to a Requirement-writing profile." A non-Satisfaction Claim may
// therefore name the same Requirement as both subject and criterion.
//
// [IMPLEMENTATION] There is no WithoutCriteria. WithCriteria(nil) already
// expresses removal, and a separate method would either duplicate the
// Claim-Type validation or bypass it. Removal is invalid for Satisfaction and
// Conformance, so a Without* method would be one that sometimes fails --
// which is why this package omits it, exactly as peos/requirement omits
// WithoutScope wherever scope can never legitimately be absent.
//
// # Correction is a new record, never a mutation or a Supersession
//
// [NORMATIVE] "Correction of a Validation Execution Record creates a new
// Validation Execution Record. A Validation Execution Record SHALL NOT be
// mutated once recorded." Likewise, "Correction, replacement, and invalidation
// of a Validation Claim are each represented by recording a new Validation
// Claim," and "A Validation Claim SHALL NOT be mutated once recorded."
//
// Both record families carry an optional core.RecordCorrectionRef whose
// target is the exact earlier record's typed reference. The reference lives on
// the *new* record and points backward, so no already-written record is ever
// rewritten to note that it was corrected -- "The earlier record remains
// historically preserved; it is not erased or overwritten."
//
// [NORMATIVE] This is not Artifact Supersession and not an Artifact Relation.
// PEOS-006 states the correction reference "is not an Artifact Relation", "is
// not Artifact Supersession as defined by PEOS-002", and "has no separate
// entity identity of its own", and forbids the vocabulary outright: the terms
// supersede / supersedes / superseded / Supersession "SHALL NOT be used to
// describe Claim replacement". core.CorrectionKind offers exactly correct,
// replace, and invalidate for that reason, and this package uses no other
// word (non-conforming pattern "Claim Replacement Called Artifact
// Supersession").
//
// [PRODUCT] Whether a correction target exists, whether the chronology is
// coherent, whether a correction chain is consistent, how conflicting records
// are resolved, and which record is currently applicable are all repository or
// Product concerns. This value layer holds one record and cannot see another.
// PEOS-006 is explicit that the currently applicable Claim "is derived by
// identifying the most recent Claim that has not been replaced or invalidated
// by a later Claim. This determination is a derived view; it is never stored
// as a field."
//
// # ExecutionEvent is content, not an Observation entity
//
// [IMPLEMENTATION] PEOS-006 says only that event history, "when required by
// the applicable Product contract, is an ordered sequence of observations
// recorded within the same immutable Execution Record". It enumerates no
// fields, so ExecutionEvent's shape -- a mandatory timestamp, a mandatory
// trimmed note, and an optional extension -- is an implementation choice, not
// a normative field set. Event order is the slice order, preserved verbatim,
// with no sequence number that could disagree with it.
//
// [NORMATIVE] An ExecutionEvent is the "content within an immutable
// Validation Execution Record" branch of PEOS-006's two permitted ways to
// record an observation. It is therefore not an entity: no identity, no
// reference type, no revision, no lifecycle -- and no severity, criterion,
// outcome, or actor, none of which PEOS-006 defines for event history.
//
// [PRODUCT] "An indeterminate or interrupted outcome SHALL NOT be silently
// treated as completed." That is an obligation on whoever interprets a
// recorded outcome, not a structural property this layer can enforce.
//
// # ActivityReference is a locator, not an entity
//
// [NORMATIVE] Every Execution Record identifies "its exact Planned Activity
// reference (Plan Revision and plan-local key), or an explicit ad hoc
// designation". ActivityReference is the closed two-arm union expressing that
// choice, with an invalid zero value so "unset" differs from either arm.
//
// [IMPLEMENTATION] It carries no identity, and there is deliberately no
// ActivityReferenceID or Ref type: a plan-local key "does not survive as an
// independent identity outside that exact Plan Revision", so only the pair
// (Plan Revision reference, key) resolves anything, and the ad hoc arm names
// an execution that was never planned.
//
// # Optional collection null behavior differs between the record families
//
// [IMPLEMENTATION] For PlannedActivity's and ExecutionRecord's optional
// collections, an absent JSON key, an explicit null, and an empty array are
// equivalent -- all mean "none declared" -- because PEOS-006 permits zero
// cardinality for each, so distinguishing them would carry no meaning.
//
// A Claim's criteria deliberately does not follow that rule. Absent means
// "zero criteria declared", which is valid for a general Claim Type and
// invalid for Satisfaction and Conformance; an explicit null is rejected
// outright, because a caller writing null has said something different from
// writing nothing. A json.RawMessage probe keeps the two distinguishable.
//
// A Claim's evidence keeps a plain typed field: absent, null, and empty array
// all yield an empty slice and all must fail the same one-or-more invariant,
// so the cases converge and need not be told apart.
//
// Optional single-value keys reject an explicit null throughout, following
// lifecycle.StateAssignment's treatment of its optional Authority; the one
// exception is extension, where core.Extension's documented contract makes
// null equivalent to absent. Each UnmarshalJSON doc comment states its actual
// missing-versus-null behavior field by field rather than asserting the two
// are identical.
//
// # A Claim records an assertion; it does not authorize anything
//
// [NORMATIVE] "Certification, acceptance, approval, rejection, and
// authorization are governance outcomes governed by PEOS-004, expressed
// through a Decision Outcome." A Decision Outcome "MAY rely on one or more
// Validation Claims as part of its Decision Basis", but "A Validation Claim
// does not replace a Decision Outcome where authority or governance judgment
// is required... it does not itself authorize, approve, or accept anything on
// behalf of an organization."
//
// Neither Claim nor ExecutionRecord carries a Decision Outcome reference or a
// governance action, so neither can express authorization; each carries only
// an optional core.AuthorityRef recording who had the right to establish it
// (non-conforming pattern "Validation Outcome Used as Governance Authority").
//
// [NORMATIVE] Waiver never rewrites a record. PEOS-006 defines no Waiver
// interaction at all, and no type here references requirement.Waiver. A
// Waiver's effect is on a *derived* conformance interpretation at runtime
// (PEOS-008), never on a recorded Claim or Execution Record. [DEFERRED] The
// criterion-level Waiver question and Waiver attached conditions are PEOS-005
// concerns.
//
// [PRODUCT] Also Product- or repository-owned, and deliberately unchecked
// here: verifying that a cited Evidence Artifact actually carries
// core.ArtifactRoleEvidence (a Ref carries no roles); Evidence relevance and
// temporal applicability; what a criterion means; aggregation of Claims into a
// derived satisfaction or conformance view; whether an authority is
// sufficient; and whether an Execution Record is compatible with the Plan
// Revision and plan-local key it names.
//
// # No relationships, no lifecycle, no derived state
//
// [NORMATIVE] PEOS-006 defines no Artifact Relation. No type here composes
// peos/relation.Relation, no wire form has a "relation" key, and
// peos/relation is not imported. Claim and Execution Record correction is a
// record-to-record reference that "is not an Artifact Relation" and "is not
// Artifact Supersession as defined by PEOS-002" -- see "Correction is a new
// record, never a mutation or a Supersession" above.
//
// [NORMATIVE] "A Validation Claim and a Validation Execution Record do not,
// by themselves, assign a Lifecycle State or a State Assignment," and an
// execution outcome SHALL NOT be represented as a Lifecycle State
// (Validation and Lifecycle Separation Invariant). No type here carries a
// Lifecycle State, and this package does not import peos/lifecycle -- that
// non-import is the structural guarantee, not merely a documented
// intention. [NORMATIVE] A Validation Plan is an ordinary Artifact and MAY
// be governed by a Lifecycle; that lifecycle belongs exclusively to PEOS-003
// and peos/lifecycle, and is not modeled here. A Lifecycle Transition MAY
// reference validation as Transition Evidence, which is PEOS-003 reading
// validation, never validation writing lifecycle.
//
// [NORMATIVE] Requirement satisfaction, Artifact conformance, quality
// evaluation, runtime compliance, and template conformance "are derived
// views, never stored fields", and no evaluated Subject owns a mutable
// satisfied / validated / conformant / quality / qualityScore / compliant
// field (Derived Satisfaction Invariant). No such field exists anywhere in
// this package. [PRODUCT] "A Product contract SHALL define the aggregation
// rule where more than one applicable subject, allocation, or Satisfaction
// Claim contributes to a derived Requirement-satisfaction view"; PEOS-006
// forbids defining a universal aggregation policy, and this package defines
// none.
//
// [NORMATIVE] Observation, Result, Verdict, Finding, and Violation are not
// PEOS-006 entities and none is introduced here. "An Observation or a
// Result is not a separate identity-bearing category distinct from
// Evidence... No third category is introduced," and "There is no separate
// Verdict entity." Finding and Violation appear nowhere in PEOS-006;
// Violation belongs to PEOS-008.
//
// [NORMATIVE] "Certification, acceptance, approval, rejection, and
// authorization are governance outcomes governed by PEOS-004, expressed
// through a Decision Outcome." No type here records acceptance,
// certification, or approval, and none carries a Decision Outcome
// reference or a governance action, so validation cannot express
// authorization (non-conforming pattern "Validation Outcome Used as
// Governance Authority"). [PRODUCT] A consumer treating a validation
// outcome as governance authority is outside what this layer can prevent.
//
// # Import boundary
//
// [IMPLEMENTATION] This package imports only the standard library and
// peos/core. It must never import peos/relation (PEOS-006 defines no
// Artifact Relation), peos/lifecycle (validation state is never a Lifecycle
// State), peos/requirement (Requirement and Requirement Revision Subjects
// arrive via core.EngineeringSubjectRef, and Requirement criteria via
// core.CriterionRef, so no PEOS-005 type is needed), or peos/decision
// (PEOS-004 references Claims, not the reverse; authority is carried as
// core.AuthorityRef). Conversely, peos/requirement must never import this
// package: PEOS-005 §30.1 keeps Requirements independent of validation
// outcomes. Those two rules together are what make an import cycle
// inexpressible. A test in this package parses its own source and fails if
// the boundary is crossed.
//
// [IMPLEMENTATION] Packet A had already placed almost the whole PEOS-006
// reference and vocabulary layer in peos/core -- ValidationPlanRef,
// ValidationPlanRevisionRef, EvidenceArtifactRevisionRef, ValidationClaimID
// and Ref, ValidationExecutionRecordID and Ref, CriterionRef, LocalKey,
// RecordRef, RecordCorrectionRef, CorrectionKind, ValidationMethod,
// ClaimType, ClaimOutcome, and ArtifactRoleEvidence. This package adds the
// aggregates on top of them.
//
// Packet H.1 required no change to peos/core at all. Packet H.2 added exactly
// one thing there: core.ExecutionOutcome, the fifth PEOS-006 vocabulary, which
// Packet A had left out while declaring its four siblings. It belongs beside
// them rather than here because PEOS-007 specializes the Execution Record as a
// Measurement Record and would otherwise need to reach into this package for
// the vocabulary. The addition is purely additive: no existing signature,
// field, or wire form changed.
package validation
