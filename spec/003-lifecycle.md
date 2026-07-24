# PEOS-003 — Lifecycle

**Category:** Normative

**Status:** Draft

**Version:** 0.1

---

# Abstract

This document defines the normative Lifecycle Model of PEOS.

It establishes how governed engineering subjects evolve through explicit states and valid transitions while preserving history, authority, evidence, and applicable invariants.

A PEOS Lifecycle is not a mandatory development methodology or a fixed sequence of organizational stages.

It is a contract governing the permitted evolution of an identifiable subject.

---

# Purpose

The purpose of the Lifecycle Model is to define:

* Lifecycle Definitions;
* Lifecycle Subjects;
* Lifecycle States;
* State Assignments;
* Transitions;
* Transition Guards;
* transition authority;
* transition evidence;
* Transition Records;
* transition effects;
* state history;
* failed and interrupted transitions;
* compensating transitions;
* lifecycle composition;
* lifecycle specialization;
* lifecycle integrity;
* extension and conformance requirements.

This document provides a common lifecycle foundation for Artifacts, Artifact Revisions, Decisions, Validations, Products, and other subjects explicitly governed by a normative PEOS specification.

---

# Scope

This specification applies whenever a PEOS subject:

* has an explicit normative state;
* may change from one normative state to another;
* requires conditions for such a change;
* requires persistent evidence of its evolution;
* is governed by lifecycle-dependent invariants.

Examples include:

* an Artifact Revision progressing from draft to accepted;
* a Decision progressing from proposed to accepted or rejected;
* a Validation progressing from pending to passed or failed;
* a Product progressing through defined readiness states;
* an Artifact becoming deprecated or withdrawn.

This specification does not require every PEOS subject to use the same states or transitions.

It defines the common model through which specialized lifecycles are expressed.

---

# Normative Foundation

This document depends on:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model.

The terms Product, Engineering State, Artifact, Artifact Revision, Decision, Validation, Runtime, Actor, Contract, Invariant, Transformation, Evidence, and Artifact Relation are used according to their normative definitions in those specifications.

This document defines lifecycle and transition semantics.

It does not redefine Artifact identity, Artifact Revision immutability, Decision structure, traceability completeness, Validation semantics, or Runtime behavior.

---

# Lifecycle Model Overview

The Lifecycle Model distinguishes between:

* a Lifecycle Definition, which defines permitted evolution;
* a Lifecycle Subject, whose state is governed;
* a Lifecycle State, which has explicit normative meaning;
* a State Assignment, which associates a Subject with a State;
* a Transition, which defines a permitted state change;
* a Transition Attempt, which requests or executes a change;
* a Transition Record, which persistently records the result.

Conceptually:

```text
Lifecycle Definition
    governs one or more Subject Types
    defines States
    defines Transitions
    defines Guards
    defines authority requirements
    defines evidence requirements
    defines lifecycle invariants

Lifecycle Subject
    has an identifiable current State Assignment
    evolves only through permitted Transitions
    preserves State History

Transition
    has source State
    has target State
    has Guards
    may require Evidence
    may require authorized Actors
    produces a Transition Record
```

A Lifecycle defines permitted evolution.

A Transition defines a permitted change.

A Transition Attempt requests or executes that change.

A Transition Record preserves what occurred.

---

# Lifecycle Definition

A Lifecycle Definition is an identifiable normative contract defining the permitted states and transitions of one or more Lifecycle Subject Types.

A Lifecycle Definition Version is an immutable and identifiable state of a Lifecycle Definition.

When a Lifecycle Definition is represented as an Artifact, its Definition Version MUST identify the corresponding Artifact Revision.

Every published Lifecycle Definition MUST have at least one identifiable Lifecycle Definition Version.

A mutable working copy of a Lifecycle Definition is not a Lifecycle Definition Version.

Material change to the normative states, transitions, guards, authority requirements, effects, or invariants of a Lifecycle Definition MUST produce a new Lifecycle Definition Version.

A Lifecycle Definition MUST define an entry Transition through which a Subject enters the lifecycle.

Every Lifecycle Definition MUST have:

* a stable identifier;
* an explicit scope;
* one or more governed Subject Types;
* one or more Lifecycle States;
* an initial-state policy;
* an entry Transition;
* a Transition set;
* applicable lifecycle invariants;
* an identifiable Lifecycle Definition Version.

A Lifecycle Definition SHOULD identify:

* terminal States;
* exceptional States;
* transition authority requirements;
* evidence requirements;
* concurrency rules;
* extension points;
* compatibility rules.

A Lifecycle Definition MUST NOT rely solely on visual diagrams or informal naming.

Its normative states, transitions, conditions, and effects MUST be inspectable.

A Lifecycle Definition MAY be represented as an Artifact.

When a Lifecycle Definition is represented as an Artifact, the Lifecycle Definition Version used for a Transition MUST identify the corresponding Artifact Revision.

A mutable reference to the latest Lifecycle Definition Version MUST NOT replace the specific Lifecycle Definition Version used by a historical Transition.

---

# Lifecycle Subject

A Lifecycle Subject is an identifiable entity whose normative state is governed by a Lifecycle Definition.

A Lifecycle Subject MAY be:

* a Product;
* an Artifact;
* an Artifact Revision;
* a Decision;
* a Validation;
* another subject explicitly permitted by a normative PEOS specification.

Every governed Lifecycle Subject MUST have:

* a stable identity;
* an identifiable Subject Type;
* an applicable Lifecycle Definition;
* an identifiable current State Assignment, unless it has not yet entered the lifecycle;
* persistent State History after its first completed transition.

The same entity MAY participate in more than one Lifecycle when the lifecycles govern distinct concerns.

For example, an Artifact Revision MAY participate in:

* a review lifecycle;
* a security classification lifecycle;
* a publication lifecycle.

Such lifecycles MUST be distinguishable.

Their State Assignments MUST NOT be silently collapsed into one ambiguous status.

---

# Lifecycle State

A Lifecycle State is an identifiable normative condition that may be assigned to a Lifecycle Subject.

Every Lifecycle State definition MUST have:

* a stable State Identifier within its Lifecycle Definition;
* an explicit semantic meaning;
* an applicability scope;
* any invariants that must hold while a Subject occupies that State;
* any normative guarantees established by that State.

A State name alone does not define a conformant Lifecycle State.

For example, a State named `accepted` is insufficient unless the Lifecycle Definition specifies what acceptance means, which authority may establish it, and which guarantees or constraints follow from it.

A Lifecycle State SHOULD define:

* permitted incoming Transitions;
* permitted outgoing Transitions;
* applicability consequences;
* required retained Evidence;
* whether the State is initial, terminal, exceptional, or ordinary.

A Lifecycle State MUST NOT silently derive its normative meaning from:

* a file location;
* a branch name;
* a database column name;
* a user-interface label;
* an undocumented Runtime convention.

Such mechanisms MAY represent State, but they do not define its semantics.

---

# State Identity

State Identifiers MUST be unique within a Lifecycle Definition.

State identity MUST remain stable within a published Lifecycle Definition Version.

A State Identifier MUST NOT be reused for a State with materially different normative meaning within the same Lifecycle Definition Version.

A later Lifecycle Definition Version MAY change State semantics only when the compatibility and migration consequences are explicit.

Textual labels MAY change without changing State identity when normative meaning remains unchanged.

Implementations MUST NOT determine State equivalence solely from display labels.

---

# Initial State

A Lifecycle Definition MUST define one or more permitted initial States.

A Subject enters a lifecycle through an entry Transition from an unassigned condition to one of the permitted initial States.

The unassigned condition represents the absence of a State Assignment.

It is not an ordinary Lifecycle State and MUST NOT be treated as part of the Lifecycle Definition's State set.

A Lifecycle Definition MAY define:

* one required initial State;
* multiple permitted initial States;
* a conditional initial-State selection rule.

A Subject MUST NOT be treated as occupying an initial State merely because it exists.

Initial assignment MUST be established through a completed entry Transition and recorded by its Transition Record.

When initial State selection depends on conditions or Evidence, those conditions MUST be inspectable.

---

# Terminal State

A Terminal State is a State from which no ordinary outgoing Transition is permitted.

A Lifecycle Definition MAY define zero, one, or multiple Terminal States.

Examples may include:

* completed;
* rejected;
* withdrawn;
* invalidated;
* retired.

A Terminal State does not imply deletion of the Subject or its history.

A Lifecycle Definition MAY permit exceptional recovery, reopening, migration, or compensation from a Terminal State.

Such behavior MUST be explicit and MUST NOT be represented as an ordinary transition when it has exceptional semantics.

---

# Exceptional State

An Exceptional State represents a condition outside normal successful lifecycle progression.

Examples may include:

* transition-failed;
* blocked;
* suspended;
* integrity-failed;
* disputed.

A Lifecycle Definition MAY use Exceptional States when a failed or interrupted operation has a persistent normative consequence.

Failure to complete a Transition does not automatically require assigning an Exceptional State.

The Lifecycle Definition MUST state whether failure:

* leaves the Subject in its source State;
* assigns an Exceptional State;
* requires compensation;
* requires manual resolution;
* produces no State change.

---

# State Assignment

A State Assignment is an identifiable association between:

* a Lifecycle Subject;
* a Lifecycle Definition;
* a Lifecycle State;
* an effective point or interval;
* the Transition Record that established it.

A State Assignment represents lifecycle status.

It is distinct from Artifact Revision content.

An Artifact Revision MAY remain immutable while its lifecycle State Assignment changes through recorded Transitions.

A State Assignment MUST NOT modify the normative content of an Artifact Revision.

When a state change also requires a change to Artifact content, the content change MUST produce a new Artifact Revision according to PEOS-002.

A State Assignment MUST identify the specific Lifecycle Definition Version under which it was established.

A recorded State Assignment MUST be immutable.

A new State Assignment supersedes the previous applicable State Assignment within the same lifecycle dimension.

The previous State Assignment remains part of State History.

---

# Current State

The Current State is the State Assignment currently applicable to a Lifecycle Subject within a particular Lifecycle.

Current State MUST be determinable from persistent lifecycle information.

Current State MUST NOT depend solely on transient Runtime memory.

A conformant implementation MUST distinguish:

* current State;
* previous States;
* requested target State;
* partially executed Transition;
* failed Transition;
* historical State that is no longer applicable.

When multiple concurrent State Assignments are allowed, the Lifecycle Definition MUST define their composition semantics.

---

# State History

State History is the ordered persistent record of State Assignments and Transition Records for a Lifecycle Subject.

State History MUST preserve:

* the States previously assigned;
* the Transition used for each change;
* the responsible Actor or Runtime;
* transition time or ordering;
* applicable evidence references;
* the Lifecycle Definition Version used;
* failure or compensation information when applicable.

State History MUST NOT be rewritten solely to present a simplified or successful narrative.

Invalid, failed, reverted, compensated, or superseded Transition Records MUST remain distinguishable when retention permits.

A corrected presentation MAY be produced, but it MUST NOT replace the underlying historical record.

---

# Transition

A Transition is an identifiable lifecycle rule permitting a Lifecycle Subject to move from one State to another under explicit conditions.

Every Transition definition MUST identify:

* a stable Transition Identifier;
* one or more permitted source States;
* one or more target States;
* applicable Subject Types;
* Transition Guards;
* transition authority requirements;
* required Evidence, when applicable;
* transition effects;
* failure behavior.

A Transition SHOULD identify:

* required inputs;
* optional inputs;
* generated outputs;
* idempotency semantics;
* concurrency behavior;
* compensation behavior;
* applicable time constraints.

A Transition MUST NOT be treated as permitted solely because a Runtime is technically capable of changing a stored status value.

---

# Transition Identity

A Transition Identifier MUST be unique within its Lifecycle Definition.

Transition identity MUST remain stable within a published Lifecycle Definition Version.

A Transition Identifier MUST NOT be reused for materially different semantics within the same Lifecycle Definition Version.

A Transition Record MUST identify the exact Transition definition used.

A human-readable Transition name does not replace Transition identity.

---

# Source and Target States

Every Transition MUST define at least one source State and at least one target State.

The source State establishes the State from which the Transition is permitted.

The target State establishes the State that becomes applicable after successful completion.

A Transition MAY:

* have one source and one target;
* permit multiple source States;
* select among multiple target States according to an explicit result;
* preserve the same State while producing other lifecycle effects.

A same-State Transition MUST have explicit normative meaning and MUST NOT be used merely to conceal an unrecorded operation.

A Runtime MUST verify that the actual current State is an allowed source State before completing a Transition.

---

# Transition Guard

A Transition Guard is a condition that MUST be satisfied before a Transition may complete successfully.

A Guard MAY evaluate:

* current State;
* Subject properties;
* Artifact Relations;
* Evidence;
* Validation results;
* Decisions;
* Actor identity;
* Actor authority;
* time constraints;
* applicable invariants;
* external conditions permitted by the Lifecycle Definition.

Every Guard MUST have:

* an identifiable definition;
* a determinable evaluation result;
* an explicit failure consequence.

A Guard result SHOULD distinguish:

* satisfied;
* not satisfied;
* not evaluated;
* indeterminate;
* evaluation failed.

An indeterminate or failed Guard MUST NOT be treated as satisfied unless the Lifecycle Definition explicitly permits a qualified override.

A Guard override, when permitted, MUST be authorized, justified, and persistently recorded.

---

# Entry and Exit Conditions

An Entry Condition is an invariant or requirement that must hold when a Subject enters a State.

An Exit Condition is an invariant or requirement that must hold before a Subject leaves a State.

Entry and Exit Conditions MAY be represented as Transition Guards.

When defined separately, their relationship to applicable Transitions MUST be explicit.

Successful completion of a Transition MUST establish all mandatory Entry Conditions of the target State.

A Transition MUST NOT complete when it would place the Subject in a target State whose mandatory invariants are known to be false.

---

# Required Input

A Required Input is information needed to evaluate or execute a Transition.

Required Inputs MAY include:

* Artifact Revisions;
* Decisions;
* Evidence;
* Actor-provided values;
* Validation results;
* external observations;
* references to other Lifecycle Subjects.

A Required Input MUST be identified with enough precision to reproduce or inspect the transition when required by the applicable contract.

A mutable reference is insufficient when a fixed historical input is necessary.

The existence of an input does not establish its validity or sufficiency.

---

# Transition Evidence

Transition Evidence is Evidence used to support completion, rejection, failure, override, or compensation of a Transition.

A Lifecycle Definition MAY require Evidence proving that:

* Guards were satisfied;
* review occurred;
* authorization was granted;
* Validation passed;
* required work was completed;
* an external condition existed;
* an exceptional override was justified.

Required Transition Evidence MUST be identifiable before the Transition is considered complete.

Transition Evidence SHOULD reference fixed Artifact Revisions when reproducibility is required.

The Lifecycle Model does not determine whether Evidence is valid or sufficient.

Those semantics are defined by PEOS-006.

---

# Actor and Authority

Every completed Transition MUST identify the responsible Actor or Runtime.

Actor identity and transition authority are distinct.

A Lifecycle Definition MUST define which Actors or authority classes may:

* request a Transition;
* approve a Transition;
* execute a Transition;
* override a Guard;
* compensate or reverse a Transition;
* resolve a failed Transition.

One Actor MAY perform multiple roles when permitted by the applicable contract.

A Lifecycle Definition MAY require separation of duties.

A Runtime MUST NOT infer authority solely from technical access.

An unauthorized Transition MUST NOT be represented as conformant merely because its state mutation succeeded.

Authority delegated to a Runtime MUST be explicit or determinable through an applicable contract.

---

# Transition Effect

A Transition Effect is a persistent or observable consequence of successful Transition completion.

Effects MAY include:

* creation of a new State Assignment;
* supersession of the previously applicable State Assignment;
* creation of a Transition Record;
* creation of Artifact Relations;
* requirement of subsequent Validation;
* activation or removal of applicability;
* generation of notifications or operational outputs.

Normative Effects MUST be identified by the Transition definition.

A Transition MUST NOT be considered complete until its mandatory normative Effects have been established.

Non-normative operational side effects MAY occur after completion, provided that their failure does not falsely change the normative result.

When a mandatory Effect fails, the Transition MUST follow its defined failure or compensation behavior.

---

# Transition Attempt

A Transition Attempt is an identifiable request or execution process intended to apply a Transition to a Lifecycle Subject.

A Transition Attempt MUST identify:

* the Subject;
* the requested Transition;
* the expected source State;
* the initiating Actor or Runtime;
* the Lifecycle Definition Version;
* the attempt time or ordering.

A Transition Attempt MAY result in:

* successful completion;
* rejection;
* Guard failure;
* authorization failure;
* execution failure;
* interruption;
* indeterminate outcome;
* compensation.

A Transition Attempt is not equivalent to a completed Transition.

A requested or partially executed Transition MUST NOT cause the target State to be presented as current before completion requirements are satisfied.

---

# Transition Record

A Transition Record is a persistent Artifact that records the attempted or completed application of a Transition to a Lifecycle Subject.

Every completed Transition MUST produce or be represented by a Transition Record.

A Transition Record MUST identify:

* the Lifecycle Subject;
* the Lifecycle Definition Version;
* the Transition definition;
* the source State Assignment;
* the resulting target State Assignment when successful;
* the responsible Actor or Runtime;
* the transition time or ordering;
* the result;
* applicable Evidence;
* Guard evaluation outcomes;
* authority basis;
* failure or override information when applicable.

A Transition Record MUST conform to PEOS-002.

Once recorded as part of persistent Engineering State, a Transition Record MUST be immutable.

Correction of erroneous transition information MUST create a new corrective Artifact or Record rather than silently modifying the original Transition Record.

A Transition Record MAY represent an unsuccessful Transition Attempt when preserving the failure is required by the applicable Lifecycle Definition or conformance profile.

---

# Atomicity

A Lifecycle Definition MUST define the normative atomicity boundary of each Transition.

At minimum, a successful Transition MUST NOT leave the system in a state where:

* the target State is declared current;
* but the required Transition Record is absent;
* or mandatory target-State invariants are not established.

An implementation MAY use transactional, append-only, event-based, distributed, or compensating mechanisms.

The physical implementation is not prescribed.

The observable Engineering State MUST remain consistent with the declared transition result.

When atomic completion cannot be guaranteed, the implementation MUST expose an intermediate, pending, or indeterminate condition rather than falsely asserting success.

---

# Idempotency and Duplicate Attempts

A Lifecycle Definition SHOULD state whether a Transition is idempotent.

Repeated execution of an idempotent Transition with the same effective inputs MUST NOT create conflicting State history.

A Runtime MUST distinguish:

* a retried Transition Attempt;
* a duplicate request;
* a new independent Transition;
* a replay of historical events.

Duplicate execution MUST NOT silently create multiple contradictory Current States.

When duplicate Transition Records are retained, their relationship and effective result MUST be determinable.

---

# Concurrency

A Lifecycle Definition MUST define or constrain concurrent Transition behavior when more than one Transition may be attempted for the same Subject.

A conformant implementation MUST prevent undetected lost updates to lifecycle State.

Before completing a Transition, the implementation MUST verify that the source State on which the attempt was based is still applicable.

If the source State changed concurrently, the Transition MUST:

* be rejected;
* be reevaluated;
* be retried under a new attempt;
* follow another explicitly defined concurrency rule.

A Transition MUST NOT overwrite a concurrent valid State change without preserving the conflict or resolution history.

---

# Failed Transition

A Failed Transition is a Transition Attempt that did not establish its target State.

A failed Transition MUST NOT cause the target State to be presented as current.

The applicable Lifecycle Definition MUST determine whether failure:

* leaves the source State unchanged;
* creates an Exceptional State;
* requires compensation;
* requires retry;
* requires manual resolution.

Failure information SHOULD identify:

* the failed step;
* the observed reason;
* completed effects;
* incomplete effects;
* affected Evidence;
* recovery requirements.

Failure of a non-normative side effect does not necessarily make the normative Transition fail.

The distinction MUST be explicit.

---

# Interrupted and Indeterminate Transition

An Interrupted Transition is an attempt whose execution stopped before a conclusive result was established.

An Indeterminate Transition is an attempt for which the system cannot establish whether mandatory completion effects occurred.

A Runtime MUST NOT automatically interpret an indeterminate result as success.

The Lifecycle Definition or Runtime Contract MUST define recovery behavior, which MAY include:

* state reconciliation;
* integrity inspection;
* retry;
* compensation;
* manual resolution;
* creation of an Exceptional State.

Resolution of an indeterminate Transition MUST be persistent and traceable.

---

# Reversal

A Reversal is a new Transition that restores a previous lifecycle State when such restoration is semantically valid.

A Reversal does not erase the original Transition.

A reversal Transition MUST:

* be explicitly defined;
* identify the State to which the Subject returns;
* preserve intervening State History;
* satisfy its own Guards and authority requirements;
* produce its own Transition Record.

Returning to a State does not imply that the entire Engineering State is identical to its earlier historical condition.

A Lifecycle Definition MUST NOT represent arbitrary history deletion as reversal.

---

# Compensating Transition

A Compensating Transition addresses the effects of an earlier completed or partially completed Transition without claiming that the earlier Transition never occurred.

A Compensating Transition MAY:

* restore a prior applicable State;
* establish a new recovery State;
* invalidate produced outputs;
* create corrective Artifacts;
* initiate additional remediation.

A Compensating Transition MUST:

* identify the Transition or Attempt being compensated;
* preserve original history;
* state the intended compensation semantics;
* produce a Transition Record.

Compensation is not equivalent to mutation or deletion of historical records.

---

# Supersession

A Lifecycle Definition Version MAY supersede another Lifecycle Definition Version.

Supersession MUST identify:

* the superseding Lifecycle Definition Version;
* the superseded Lifecycle Definition Version;
* effective scope;
* compatibility consequences;
* migration requirements when applicable.

Historical Transitions MUST remain associated with the Lifecycle Definition Version under which they occurred.

A new Lifecycle Definition Version MUST NOT retroactively reinterpret completed historical Transitions without an explicit migration or reinterpretation record.

Artifact or Artifact Revision supersession remains governed by PEOS-002 and applicable specialized lifecycles.

---

# Lifecycle Migration

Lifecycle Migration is the controlled adoption of a different Lifecycle Definition or Lifecycle Definition Version for an existing Subject.

Migration MUST define:

* the source Lifecycle Definition Version;
* the target Lifecycle Definition Version;
* State mapping;
* invalid or unmappable States;
* authority requirements;
* Evidence requirements;
* historical preservation;
* transition or migration records.

A State with the same textual label in two Lifecycle Definitions MUST NOT automatically be assumed equivalent.

Migration MUST preserve the original State History.

A migration MAY produce a new current State Assignment without pretending that a normal domain Transition occurred.

Such an assignment MUST be identifiable as migration-derived.

---

# Lifecycle Composition

Lifecycle Composition combines multiple Lifecycle Definitions or lifecycle dimensions for a Subject.

Composition MAY be used when independent concerns must evolve separately.

For example:

```text
Review Lifecycle:
Draft → Reviewed → Accepted

Publication Lifecycle:
Private → Internal → Public
```

A composed lifecycle model MUST define:

* the participating Lifecycle Definitions;
* whether their States are independent;
* cross-lifecycle Guards;
* prohibited State combinations;
* coordinated Transitions;
* conflict-resolution behavior.

Composition MUST NOT create an ambiguous single State label from multiple independent State Assignments.

---

# Orthogonal States

A Subject MAY occupy multiple orthogonal States when each State belongs to a distinct lifecycle dimension.

For example, a Subject MAY be:

```text
Review State: Accepted
Publication State: Internal
Security State: Restricted
```

Orthogonal State Assignments MUST:

* identify their Lifecycle Definition;
* remain independently inspectable;
* obey cross-lifecycle constraints;
* avoid contradictory applicability claims.

A Runtime MUST NOT flatten orthogonal States into one undocumented composite status.

---

# Lifecycle Specialization

A specialized Lifecycle Definition MAY refine another Lifecycle Definition.

Specialization MAY:

* add States;
* add Transitions;
* strengthen Guards;
* strengthen authority requirements;
* add Evidence requirements;
* add invariants;
* restrict permitted paths.

A specialization MUST NOT weaken inherited normative requirements while claiming full conformance to the parent Lifecycle Definition.

When a specialization changes inherited semantics, it MUST be treated as a distinct Lifecycle Definition rather than a conformant specialization.

The relationship to the parent definition MUST be explicit.

---

# Lifecycle Extension

A PEOS implementation MAY define extension lifecycles, States, Guards, and Transitions.

An extension MUST:

* have an identifiable namespace or ownership boundary;
* preserve applicable PEOS invariants;
* identify its governed Subject Types;
* define its normative semantics;
* remain distinguishable from normative PEOS definitions;
* avoid identifier collisions.

An extension MUST NOT silently redefine a normative State or Transition with different meaning.

An implementation MAY support nonconformant lifecycle behavior, but it MUST explicitly identify that behavior as nonconformant or partially conformant.

---

# Lifecycle and Artifact Revision

Artifact Revision and lifecycle State MUST remain conceptually distinct.

An Artifact Revision represents a fixed engineering content state.

A lifecycle State represents the normative condition or applicability of a Lifecycle Subject.

Changing lifecycle State does not, by itself, require creation of a new Artifact Revision.

Changing Artifact content or normative meaning requires a new Artifact Revision according to PEOS-002.

For example:

```text
Artifact Revision R7
    content remains immutable

State History
    Draft
    Reviewed
    Accepted
```

The State Assignments and Transition Records change over time.

Revision R7 does not.

When lifecycle State is embedded inside Revision content or normative metadata, changing that embedded State would require a new Revision.

Implementations SHOULD therefore represent frequently changing lifecycle State separately from immutable Revision content unless the applicable contract intentionally couples them.

---

# Lifecycle and Engineering State

Lifecycle Transitions modify Engineering State by changing which State Assignments are currently applicable.

A Transition MAY also affect:

* Artifact applicability;
* Decision applicability;
* Validation requirements;
* traceability requirements;
* quality claims;
* Runtime permissions.

A lifecycle State MUST NOT be interpreted as proof that all broader Product goals or business outcomes are correct.

For example, an `accepted` Artifact may satisfy its acceptance contract while the Product itself remains incomplete or unsuccessful.

---

# Lifecycle Invariants

A conformant implementation MUST preserve the following invariants.

## Definition Identity Invariant

Every governed lifecycle has an identifiable Lifecycle Definition and Lifecycle Definition Version.

## Subject Identity Invariant

Every State Assignment and Transition Record identifies its Lifecycle Subject.

## State Meaning Invariant

Every Lifecycle State has explicit normative meaning.

## Current State Invariant

The Current State of a governed Subject is determinable from persistent lifecycle information.

## Transition Identity Invariant

Every Transition has an identifiable definition.

## Source State Invariant

A Transition completes only from a permitted source State.

## Target State Invariant

A completed Transition establishes an explicitly permitted target State.

## Guard Invariant

A Transition does not complete unless all mandatory Guards are satisfied or an explicitly permitted override is recorded.

## Authority Invariant

A completed Transition has an identifiable authority basis.

## Evidence Invariant

Required Transition Evidence is identifiable and retained according to the applicable contract.

## Record Invariant

Every completed Transition has an immutable Transition Record.

## History Preservation Invariant

Transition, reversal, compensation, migration, and failure do not silently rewrite State History.

## Atomicity Invariant

A target State is not declared current without the mandatory normative effects of the Transition.

## Failure Integrity Invariant

An unsuccessful or indeterminate Transition is not silently represented as successful.

## Revision Separation Invariant

Lifecycle State changes do not mutate recorded Artifact Revision content.

## Runtime Independence Invariant

Lifecycle semantics do not depend on undocumented behavior of a specific Runtime.

## Extension Invariant

Extension lifecycles preserve the applicable requirements of this specification.

---

# Runtime Independence

The Lifecycle Model is independent of any specific Runtime.

A Runtime MAY:

* evaluate Guards;
* request Transitions;
* execute Transitions;
* verify authority;
* collect Evidence;
* create Transition Records;
* derive Current State;
* detect concurrency conflicts;
* perform compensation;
* migrate Lifecycle Definitions.

A Runtime MUST NOT:

* redefine State or Transition semantics;
* bypass mandatory Guards while claiming conformance;
* infer authority solely from technical capability;
* declare target State before mandatory completion effects;
* mutate historical Transition Records;
* erase failed or compensating history;
* conflate Artifact Revision changes with lifecycle State changes;
* reinterpret historical Transitions using a different Lifecycle Definition Version without recording that reinterpretation.

Runtime-specific execution obligations are defined by PEOS-008.

---

# Storage Independence

PEOS does not require Lifecycle information to be stored:

* in a single database;
* in Artifact metadata;
* in a Git repository;
* in an event log;
* in files;
* in a PEOS-specific service.

A conformant implementation MAY use:

* append-only events;
* transactional records;
* immutable Artifacts;
* relational state;
* distributed logs;
* another storage model.

The implementation MUST preserve:

* Subject identity;
* State identity;
* Transition identity;
* Current State determinability;
* State History;
* Transition Record integrity;
* applicable ordering;
* concurrency safety.

Physical storage independence does not weaken normative lifecycle requirements.

---

# Conformance

An implementation conforms to this specification when every lifecycle it presents as PEOS-conformant satisfies all applicable requirements of this document.

Conformance requires, at minimum:

* an identifiable Lifecycle Definition;
* explicit State semantics;
* identifiable Lifecycle Subjects;
* persistent Current State;
* explicit Transitions;
* Guard enforcement;
* authority evaluation;
* required Evidence handling;
* immutable Transition Records;
* preserved State History;
* defined failure behavior;
* separation of lifecycle State from immutable Artifact Revision content;
* identifiable extension boundaries.

A tool MAY conform only for selected Subject Types or Lifecycle Definitions.

Its conformance statement MUST identify that scope.

An implementation MUST NOT claim full Lifecycle conformance when it:

* mutates state without persistent Transition Records;
* cannot identify the Lifecycle Definition used;
* cannot preserve historical State Assignments;
* treats failed Transitions as successful;
* relies entirely on undocumented status labels;
* permits mandatory Guards to be silently bypassed.

---

# Non-Goals

This document does not define:

* one universal lifecycle for all Products;
* one mandatory set of lifecycle States;
* a mandatory linear engineering process;
* project management stages;
* organizational reporting structures;
* issue tracker workflows;
* source-control branching strategies;
* deployment environments;
* Artifact content schemas;
* Decision lifecycle details;
* Validation algorithms;
* quality metrics;
* Runtime APIs;
* user-interface behavior;
* authorization infrastructure;
* physical transaction mechanisms.

These concerns are defined by specialized PEOS specifications, implementation profiles, Product contracts, or Runtime implementations.

---

# References

This document depends on:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model.

This document provides lifecycle foundations for:

* PEOS-004 — Decision Model;
* PEOS-005 — Traceability;
* PEOS-006 — Validation;
* PEOS-007 — Quality Model;
* PEOS-008 — Runtime Contract;
* PEOS-009 — Template Contract.