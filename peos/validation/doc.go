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
// [DEFERRED] This package currently implements the Validation Plan side of
// PEOS-006 only, delivered as Packet H.1: Plan, PlanRevision, PlanContent,
// PlanApplicability, and PlannedActivity.
//
// It does NOT yet implement Validation Execution Record, Validation Claim
// (including Satisfaction Claim and Conformance Claim), Execution Outcome,
// ActivityReference, or Claim/Execution correction records. Those are
// Packet H.2, whose architecture is already accepted; errors.go declares
// their sentinels in advance and marks each one reserved. Nothing in this
// package should be read as implementing, or as a substitute for, a
// recorded Validation Claim.
//
// [NORMATIVE] PEOS-006 keeps four concepts strictly distinct: "A Validation
// Plan, a Planned Validation Activity, an Execution Record, and a
// Validation Claim are four distinct concepts. None is a substitute for
// another." H.1 implements the first two.
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
// Validation Claim's criteria will behave in Packet H.2, where a Claim Type
// may forbid the empty case and the three inputs therefore carry different
// meanings and must be told apart. Optional single-value keys here
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
// # No relationships, no lifecycle, no derived state
//
// [NORMATIVE] PEOS-006 defines no Artifact Relation. No type here composes
// peos/relation.Relation, no wire form has a "relation" key, and
// peos/relation is not imported. Claim correction, when Packet H.2 adds it,
// is a record-to-record reference that "is not an Artifact Relation" and
// "is not Artifact Supersession as defined by PEOS-002".
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
// [IMPLEMENTATION] Packet A had already placed the PEOS-006 reference and
// vocabulary layer in peos/core -- ValidationPlanRef,
// ValidationPlanRevisionRef, EvidenceArtifactRevisionRef, ValidationClaimID
// and Ref, ValidationExecutionRecordID and Ref, CriterionRef, LocalKey,
// RecordRef, RecordCorrectionRef, CorrectionKind, ValidationMethod,
// ClaimType, ClaimOutcome, and ArtifactRoleEvidence. This package adds only
// the aggregates on top of them and required no change to peos/core.
package validation
