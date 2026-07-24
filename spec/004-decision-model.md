# PEOS-004 — Decision Model

## Abstract

This specification defines the PEOS Decision Model.

The Decision Model describes how identifiable engineering Decisions govern the establishment, change, rejection, deferral, authorization, prohibition, or removal of normative engineering intent.

It distinguishes:

* a Decision from its documentary representation;
* a Decision Outcome from the lifecycle State of the Decision;
* an engineering determination from the reasoning that supports it;
* authority from authorship;
* a Decision from a Lifecycle Transition;
* normative commitments from implementation actions.

A Decision is not equivalent to a document, approval, vote, recommendation, or Transition.

A Decision MAY be represented by one or more Artifacts, but its identity and normative meaning are independent of any particular representation, storage system, workflow engine, or Runtime.

---

# Purpose

The purpose of the Decision Model is to provide a stable and implementation-independent foundation for:

* identifying engineering Decisions;
* recording what was decided;
* identifying why a Decision was made;
* distinguishing evidence, assumptions, constraints, and rationale;
* establishing the scope and applicability of Decisions;
* identifying the Actors authorized to make Decisions;
* preserving rejected and deferred alternatives;
* relating Decisions to Artifacts, Requirements, Lifecycle Transitions, Validation, and Engineering State;
* superseding or invalidating Decisions without rewriting history;
* reproducing the normative context under which an engineering commitment was established.

---

# Scope

This specification defines:

* Decision;
* Decision Identity;
* Decision Subject;
* Decision Question;
* Decision Context;
* Decision Alternative;
* Decision Basis;
* Decision Outcome;
* Engineering Commitment;
* Decision Consequence;
* Decision Applicability;
* Decision Authority;
* Decision Roles;
* Decision Record;
* Decision Relationships;
* Decision Lifecycle;
* Decision integrity and invariants;
* Decision conformance requirements.

This specification does not define:

* a universal decision-making methodology;
* a mandatory voting process;
* a mandatory approval hierarchy;
* a universal set of Decision lifecycle States;
* a universal significance classification;
* a storage schema;
* a Runtime implementation;
* a user interface;
* a mandatory document format;
* an algorithm that mechanically derives Decisions from Evidence.

---

# Normative Foundation

The key words MUST, MUST NOT, REQUIRED, SHALL, SHALL NOT, SHOULD, SHOULD NOT, RECOMMENDED, MAY, and OPTIONAL in this specification are to be interpreted as normative requirements.

This specification builds upon:

* PEOS-000 — Overview;
* PEOS-001 — Core Concepts;
* PEOS-002 — Artifact Model;
* PEOS-003 — Lifecycle.

Terms defined by those specifications retain their normative meaning unless explicitly specialized here.

---

# Decision Model Overview

The Decision Model separates five concepts:

```text
Decision
    identifiable engineering determination

Decision Outcome
    proposed or established result of the Decision

Engineering Commitment
    normative intent established, changed, or removed by an applicable Decision Outcome

Decision Record
    Artifact representing the Decision

Decision Lifecycle
    applicability and governance state of the Decision
```

The following relationship is normative:

```text
Decision
    has
        proposed or established Decision Outcome

Established Decision Outcome
    MAY establish, change, reject, defer, authorize, prohibit, or remove
        Engineering Commitment
```

A proposed Decision Outcome does not have normative effect unless the applicable Decision Lifecycle and authority rules establish that effect.

An established Decision Outcome MAY result in zero, one, or multiple Engineering Commitments.

An applicable Decision MAY have normative consequences without causing an immediate Lifecycle Transition or Artifact Revision.

A Decision MUST NOT be treated as equivalent to the Artifact that records it.

---

# Decision

A Decision is an identifiable engineering determination concerning normative engineering intent for a defined Subject.

When applicable, a Decision establishes, changes, rejects, defers, authorizes, prohibits, or removes normative engineering intent according to its Decision Outcome.

A Decision MUST have:

* a stable identity;
* a defined Subject or Decision Question;
* a Decision Outcome;
* an applicability scope;
* - an identifiable authority requirement or authority basis;
* an identifiable lifecycle;
* sufficient persistent representation when required by the applicable Product contract.

A proposed Decision MAY identify the authority required to establish its Outcome before that authority has been exercised.

Every applicable Decision MUST have an established authority basis.

A Decision MAY:

* select an Alternative;
* reject one or more Alternatives;
* defer resolution;
* authorize an action or Transition;
* prohibit an action or Transition;
* accept a risk;
* impose a constraint;
* establish an Engineering Commitment;
* modify an existing Engineering Commitment;
* remove an existing Engineering Commitment;
* delegate another Decision;
* require further investigation;
* record that no action is currently required.

A discussion, recommendation, observation, or preference is not a Decision merely because it concerns a Decision Question.

A proposal becomes a Decision when it has a stable Decision identity, an identifiable proposed Decision Outcome, and is governed by an applicable Decision Lifecycle.

A proposed Decision MUST NOT be treated as establishing applicable normative engineering intent unless its Lifecycle State and authority rules make that intent applicable.

---

# Decision Identity

Every Decision MUST have a stable Decision Identifier.

A Decision Identifier MUST identify the Decision independently of:

* its Decision Record;
* its current lifecycle State;
* its storage location;
* its representation format;
* its responsible Runtime;
* its current applicability.

A Decision Identifier MUST NOT be reused for a materially different Decision.

A Decision that replaces or contradicts an earlier Decision MUST receive a distinct Decision Identifier.

Correction of a typographical or representational error MAY preserve Decision identity when the normative meaning of the Decision does not change.

Material change to the Decision Outcome, applicability, authority basis, or Engineering Commitment MUST NOT be represented as a silent correction.

---

# Decision Subject

A Decision Subject is the Entity, concern, question, scope, or Engineering State to which a Decision applies.

A Decision MUST identify at least one Decision Subject or one explicit Decision Question.

A Decision Subject MAY be:

* a Product;
* an Artifact;
* an Artifact Revision;
* a Lifecycle Subject;
* a Requirement;
* an interface;
* a component;
* a system;
* a policy;
* a risk;
* a constraint;
* another Decision;
* a defined engineering concern.

A Decision MUST NOT be assumed to apply to all Products, Artifacts, or contexts merely because its Subject is not explicitly limited.

Where the Decision Subject is broad, the Decision Applicability MUST define the effective boundaries.

---

# Decision Question

A Decision Question identifies the issue resolved, constrained, rejected, or deferred by a Decision.

A Decision Question SHOULD be explicit when the Decision:

* evaluates multiple Alternatives;
* resolves a conflict;
* accepts a risk;
* delegates authority;
* supersedes another Decision;
* affects multiple Lifecycle Subjects;
* has significant or irreversible consequences.

A Decision Question MAY be expressed as:

* a question;
* a problem statement;
* a choice to be made;
* a risk disposition;
* a policy determination;
* an authorization request.

The Decision Question MUST be distinguishable from the Decision Outcome.

Example:

```text
Decision Question:
Which database technology shall be used for transactional persistence?

Decision Outcome:
PostgreSQL is selected.
```

---

# Decision Context

Decision Context is the set of circumstances under which the Decision was made and interpreted.

Decision Context MAY include:

* organizational context;
* Product context;
* technical context;
* temporal context;
* operational context;
* regulatory context;
* market context;
* resource limitations;
* known incidents;
* prior Engineering State;
* relevant Lifecycle States;
* applicable policies.

Contextual information MUST NOT be represented as Evidence unless it supports a claim used in the Decision Basis.

A Decision Record SHOULD preserve enough Decision Context to allow a later Actor to understand the conditions under which the Decision was made.

---

# Decision Significance

Decision Significance expresses the expected engineering impact of a Decision.

A Product contract MAY define significance levels.

Decision Significance MAY consider:

* reversibility;
* cost of change;
* safety impact;
* security impact;
* regulatory impact;
* architectural scope;
* operational impact;
* number of affected Subjects;
* duration of applicability;
* uncertainty;
* dependency impact.

A significant Decision SHOULD have:

* an explicit Decision Question;
* considered Alternatives;
* an explicit Decision Basis;
* an explicit authority basis;
* a persistent Decision Record;
* identified consequences;
* identified applicability;
* identified validation or review expectations.

A Decision MUST NOT be treated as insignificant solely because it is easy to implement.

---

# Decision Alternative

A Decision Alternative is a materially distinct course of action, position, disposition, or outcome considered in relation to a Decision Question.

An Alternative MUST be distinguishable from the Decision Outcome.

An Alternative MAY represent:

* an implementation choice;
* an architectural option;
* acceptance of a risk;
* rejection of a proposal;
* deferral;
* maintenance of the current state;
* delegation;
* further investigation;
* no action.

A Decision MAY consider one or more Alternatives.

A Decision is not required to enumerate every theoretically possible Alternative.

A significant Decision SHOULD identify the materially relevant Alternatives that were considered.

---

# Selected Alternative

A selected Alternative is an Alternative adopted by the Decision Outcome.

Selection of an Alternative MUST NOT be interpreted beyond the Decision Applicability.

A selected Alternative MAY establish one or more Engineering Commitments.

Selection alone does not prove that the selected Alternative has been implemented, validated, or satisfied.

---

# Rejected Alternative

A rejected Alternative is an Alternative explicitly not adopted within the scope of a Decision.

A rejected Alternative SHOULD identify the reason for rejection when the Decision is significant.

Rejection of an Alternative:

* does not make the Alternative universally invalid;
* does not prohibit its use outside the Decision Applicability;
* does not prove that the Alternative is technically impossible;
* does not erase Evidence that supported it.

A previously rejected Alternative MAY be reconsidered by a later Decision.

---

# Deferred Alternative

A deferred Alternative is an Alternative whose resolution has been intentionally postponed.

Deferral MUST identify, when applicable:

* the unresolved question;
* the reason for deferral;
* the conditions for reconsideration;
* the responsible Actor or authority;
* the expected review point.

A deferred Decision Outcome MUST NOT be treated as acceptance or rejection.

---

# Decision Basis

Decision Basis is the inspectable set of information used to support, constrain, or explain a Decision.

Decision Basis MAY include:

* Evidence;
* assumptions;
* Requirements;
* constraints;
* prior Decisions;
* policies;
* standards;
* risks;
* trade-offs;
* estimates;
* expert judgment;
* uncertainty.

Decision Basis MUST be distinguishable from Decision Outcome.

Decision Basis does not mechanically determine the Decision.

Evidence supports a Decision but does not automatically establish Decision Authority.

A Decision MAY be validly made under uncertainty when that uncertainty is made explicit.

---

# Evidence

Evidence used in a Decision Basis MUST be identifiable.

Evidence SHOULD identify:

* its source;
* its relevant content;
* its applicability;
* its observation or production time;
* its known limitations;
* its relationship to the Decision.

Evidence MUST NOT be silently altered after being used as part of a recorded Decision Basis.

A later correction or replacement of Evidence MUST NOT retroactively rewrite the historical Decision Basis.

A later Decision MAY reconsider the earlier Decision using corrected or additional Evidence.

---

# Assumption

An Assumption is a proposition treated as true for the purpose of making a Decision without sufficient Evidence to establish it as fact.

An Assumption used materially by a Decision MUST be explicit.

An Assumption SHOULD identify:

* its scope;
* its source;
* its uncertainty;
* its expected validation condition;
* the consequence if it proves false.

An Assumption MUST NOT be represented as verified Evidence.

Invalidation of an Assumption does not automatically rewrite the Decision.

It MAY trigger review, invalidation, supersession, or another Decision.

---

# Constraint

A Constraint is a normative or practical limitation that restricts the available Alternatives or Decision Outcome.

A Constraint MAY originate from:

* a Requirement;
* a Decision;
* a contract;
* a regulation;
* a policy;
* a standard;
* available resources;
* technology limitations;
* time limitations;
* safety or security conditions.

A material Constraint SHOULD identify its normative source.

A Decision MUST NOT claim freedom to select an Alternative that violates an applicable Constraint unless the Decision also has explicit authority to waive, replace, or supersede that Constraint.

---

# Prior Decision

A Decision MAY depend on one or more prior Decisions.

Such dependency MUST identify the relationship between the Decisions.

A prior Decision MAY:

* constrain the current Decision;
* authorize the current Decision;
* establish context;
* establish an Engineering Commitment;
* be superseded by the current Decision;
* conflict with the current Decision.

A Decision MUST NOT silently override an applicable prior Decision.

---

# Rationale

Decision Rationale explains why the Decision Outcome was adopted.

Rationale MAY include:

* interpretation of Evidence;
* comparison of Alternatives;
* evaluation criteria;
* trade-offs;
* accepted disadvantages;
* risk reasoning;
* expected benefits;
* rejected reasoning;
* uncertainty.

Rationale is not equivalent to Evidence.

Rationale MAY contain judgment, interpretation, or value-based evaluation.

For significant Decisions, the Rationale SHOULD be sufficient to explain why the selected Outcome was preferred over materially relevant Alternatives.

---

# Uncertainty

Known material uncertainty in a Decision Basis MUST be explicit.

Uncertainty MAY concern:

* Evidence quality;
* assumptions;
* estimates;
* future conditions;
* implementation feasibility;
* cost;
* timing;
* impact;
* reversibility;
* risk.

A Decision MAY remain applicable despite uncertainty.

The applicable Product contract MAY require review conditions based on defined uncertainty thresholds or future events.

---

# Decision Outcome

A Decision Outcome is the proposed or established result of a Decision.

An established Decision Outcome has normative effect within its Decision Applicability.

Every Decision MUST have an identifiable proposed or established Decision Outcome.

A Decision Outcome becomes established when the applicable Decision Lifecycle and authority rules give it normative effect.

A Decision Outcome MUST be distinguishable from:

* the Decision Question;
* the Decision Basis;
* the selected Alternative;
* the Decision lifecycle State;
* implementation actions;
* Transition effects.

A Decision Outcome MAY:

* select;
* reject;
* defer;
* authorize;
* prohibit;
* constrain;
* delegate;
* accept risk;
* require investigation;
* establish an Engineering Commitment;
* change an Engineering Commitment;
* remove an Engineering Commitment;
* state that no change is required.

Example:

```text
Decision lifecycle State:
Accepted

Decision Outcome:
PostgreSQL is selected as the transactional system of record.
```

The lifecycle State indicates whether the Decision is applicable.

The Decision Outcome indicates what was decided.

---

# Engineering Commitment

An Engineering Commitment is normative engineering intent established, changed, or removed by an established Decision Outcome.

Within this specification, Engineering Commitment is a semantic component of the Decision Model and is not required to be a separate top-level PEOS Entity.

An Engineering Commitment MAY:

* require an Engineering State;
* prohibit an Engineering State;
* constrain future Artifacts or Artifact Revisions;
* require a Lifecycle Transition;
* prohibit a Lifecycle Transition;
* define an architectural rule;
* establish an implementation obligation;
* establish an operational obligation;
* establish a validation expectation;
* accept a defined risk;
* impose a review condition.

An Engineering Commitment SHOULD be expressed in a form that makes its normative effect inspectable.

Example:

```text
Decision Outcome:
PostgreSQL is selected.

Engineering Commitment:
All newly created transactional persistence components SHALL use PostgreSQL unless superseded by a later applicable Decision.
```

An established Decision Outcome MAY establish zero, one, or multiple Engineering Commitments.

An Engineering Commitment MAY also have a normative source outside a Decision, including:

* a Requirement;
* a contract;
* a regulation;
* a policy;
* a standard.

This specification defines only Engineering Commitments established, changed, or removed through Decisions.

It does not define a universal Commitment Model.

---

# Commitment Effect

A Decision MUST make clear whether it:

* establishes a new Engineering Commitment;
* changes an existing Engineering Commitment;
* removes an existing Engineering Commitment;
* rejects a proposed Engineering Commitment;
* defers establishment of an Engineering Commitment;
* leaves existing Engineering Commitments unchanged.

A Decision that changes or removes an existing Engineering Commitment MUST identify the affected prior Decision or normative source when identifiable.

A removed Engineering Commitment MUST remain historically inspectable through the Decisions and Records that established and removed it.

---

# Decision Consequence

A Decision Consequence is an expected, required, permitted, or accepted effect of a Decision Outcome.

A Consequence MAY include:

* required Artifact changes;
* required Lifecycle Transitions;
* implementation work;
* migration work;
* validation work;
* operational changes;
* accepted risks;
* new constraints;
* deprecated behavior;
* review obligations;
* follow-up Decisions.

Consequences MUST be distinguishable from completed effects.

Expected Consequences do not prove implementation.

A Decision MAY be applicable even when its Consequences have not yet been completed.

---

# Decision Applicability

Decision Applicability defines where, when, and under what conditions a Decision has normative effect.

A Decision Applicability MUST identify sufficient scope to determine whether the Decision governs a given Subject or Engineering State.

Applicability MAY include:

* Product scope;
* organizational scope;
* component scope;
* Artifact scope;
* Lifecycle scope;
* temporal scope;
* version scope;
* environmental scope;
* jurisdiction;
* deployment scope;
* conditional applicability.

A Decision MUST NOT be assumed to apply outside its explicit or deterministically derived Applicability.

Where multiple applicable Decisions conflict, the conflict MUST be resolved by an explicit authority rule, priority rule, supersession relation, or later Decision.

Repository order, filename order, creation time, or storage order MUST NOT be treated as implicit normative priority unless defined by the applicable Product contract.

---

# Decision Authority

Decision Authority is the legitimate capability of an Actor to establish a Decision for a defined scope.

Every applicable Decision MUST identify an authority basis.

An authority basis MAY originate from:

* organizational responsibility;
* delegated authority;
* policy;
* contract;
* regulation;
* ownership;
* governance rule;
* Lifecycle State;
* prior Decision.

Decision Authority MUST be distinguishable from:

* authorship;
* facilitation;
* recommendation;
* implementation;
* documentation;
* Runtime execution.

Creation of a Decision Record does not prove Decision Authority.

A Runtime does not acquire Decision Authority solely because it generated, stored, analyzed, recommended, or executed a Decision.

---

# Decision Roles

A Decision MAY identify one or more Actors in distinct roles.

Decision roles MAY include:

* Decision Author;
* Decision Proposer;
* Decision Maker;
* Decision Approver;
* Decision Reviewer;
* Decision Executor;
* Decision Recorder;
* Decision Owner.

The applicable Product contract MAY define additional roles.

One Actor MAY perform multiple roles unless prohibited by an applicable separation-of-duties rule.

Role identity MUST NOT be inferred solely from document authorship or repository ownership.

---

# Decision Maker

A Decision Maker is an Actor authorized to establish the normative effect of a Decision Outcome.

A Decision Maker MUST have authority applicable to the Decision Subject, Decision Outcome, and Decision Applicability.

A Decision Maker MAY be:

* a human;
* a group;
* an organization;
* an automated system;
* a Runtime acting under explicitly delegated authority.

Where a Runtime is permitted to act as Decision Maker, the delegation scope and authority basis MUST be inspectable.

---

# Decision Approver

A Decision Approver is an Actor whose authorization is required before a Decision becomes applicable.

Approval is not universally required.

Where approval is required:

* the approving Actor MUST be identifiable;
* the authority basis MUST be inspectable;
* the approval MUST be associated with the applicable Decision state or Transition Record;
* approval MUST NOT be inferred from silence unless explicitly permitted.

Approval does not replace the Decision Outcome.

---

# Delegation

Decision Authority MAY be delegated.

A delegation MUST identify:

* the delegating Actor or authority;
* the receiving Actor;
* the delegated scope;
* applicable constraints;
* the effective period or condition;
* revocation conditions when applicable.

A delegated Actor MUST NOT exercise authority outside the delegated scope.

Delegation MUST NOT silently transfer accountability where the applicable Product contract distinguishes authority from accountability.

---

# Separation of Duties

A Product contract MAY require separation between:

* Decision Author and Decision Approver;
* Decision Maker and Decision Executor;
* Decision Maker and Validator;
* Decision Recorder and Decision Maker;
* proposer and approver.

Where separation of duties is required, conformance MUST be deterministically inspectable.

A Runtime MUST NOT bypass a required separation-of-duties rule.

---

# Decision Record

A Decision Record is an Artifact that represents a Decision.

A Decision Record MUST conform to PEOS-002.

A Decision Record SHOULD identify:

* the Decision Identifier;
* the Decision Subject;
* the Decision Question;
* the Decision Outcome;
* the Decision Applicability;
* the Decision Basis;
* the authority basis;
* relevant Actors and roles;
* Alternatives;
* Rationale;
* Consequences;
* related Engineering Commitments;
* related Decisions;
* the applicable lifecycle information.

A Decision MAY be represented by more than one Decision Record.

Each Decision Record Revision MAY have one or more Representations according to PEOS-002.

Multiple Decision Records representing the same Decision MUST preserve the same normative Decision identity.

A Decision Record MUST NOT silently represent materially different Decision Outcomes under the same Decision Identifier.

---

# Decision Record Revision

A Decision Record Revision is an Artifact Revision of a Decision Record.

Decision Record Revisions MUST be immutable.

A new Decision Record Revision MAY:

- clarify existing content without changing its normative meaning;
- reference supplemental Evidence produced after the Decision;
- correct non-normative errors;
- provide an additional Representation;
- add traceability relationships;
- reference later Transition Records or State Assignments.

Supplemental Evidence added after the Decision MUST be distinguishable from Evidence that formed the original Decision Basis.

A Decision Record Revision MUST NOT represent later Evidence as though it had been available when the Decision was made.

Later lifecycle State MUST remain established by the applicable State Assignment and Transition Record rather than by the content of a Decision Record Revision.

A Decision Record Revision MUST NOT silently rewrite the historical Decision Outcome.

A material change to the Decision Outcome MUST be represented as:

* a new Decision;
* an explicit superseding Decision;
* an explicit invalidating Decision;
* another explicit normative relationship.

---

# Decision and Record Consistency

The normative meaning of a Decision MUST be deterministically recoverable from its applicable Decision Record or other conformant persistent representation.

Where multiple Records conflict, the applicable Product contract MUST define how the authoritative representation is identified.

Modification time, repository order, filename, or display order MUST NOT be treated as implicit authority.

A Decision Record MAY contain non-normative explanatory content.

Normative and non-normative content SHOULD be distinguishable.

---

# Decision Lifecycle

A Decision is a Lifecycle Subject.

Its lifecycle MUST conform to PEOS-003.

A Decision lifecycle expresses the governance and applicability state of the Decision.

It does not express the content of the Decision Outcome.

A Product contract MAY define a Decision lifecycle containing States such as:

* Proposed;
* Under Review;
* Accepted;
* Rejected;
* Withdrawn;
* Superseded;
* Invalidated;
* Reopened.

These State names are illustrative and are not universally required.

A Decision MUST NOT be treated as applicable solely because a Decision Record exists.

A Decision becomes applicable according to its Lifecycle Definition and applicable authority rules.

---

# Proposed Decision

A Proposed Decision expresses a candidate Decision Outcome that has not yet become applicable.

A proposal MAY include Alternatives, Rationale, Evidence, and expected Consequences.

A Proposed Decision MUST NOT be treated as establishing an applicable Engineering Commitment unless explicitly permitted by the applicable Product contract.

---

# Accepted Decision

An Accepted Decision is a Decision whose applicable lifecycle establishes that its Decision Outcome has normative effect.

Acceptance MUST NOT be confused with implementation completion.

An Accepted Decision MAY have incomplete Consequences.

An Accepted Decision MAY later be superseded, invalidated, or reopened.

---

# Rejected Decision

A Rejected Decision is a Decision proposal whose proposed Outcome did not become applicable.

Rejection MUST preserve historical inspectability.

A Rejected Decision MUST NOT be treated as though it never existed.

Rejection MAY itself be a Decision Outcome when the rejection creates normative consequences.

---

# Withdrawn Decision

A Withdrawn Decision is a proposal removed from active consideration by an authorized Actor before becoming applicable.

Withdrawal MUST NOT be treated as rejection unless defined by the applicable Lifecycle Definition.

A withdrawn proposal MAY be reintroduced as a new Decision or through an explicit reopening Transition.

---

# Superseded Decision

A Decision is superseded when another applicable Decision explicitly replaces all or part of its normative effect.

Supersession MUST identify:

* the superseding Decision;
* the superseded Decision;
* the affected Applicability;
* whether supersession is complete or partial;
* the effective condition or time.

Supersession MUST NOT delete or rewrite the superseded Decision.

A superseded Decision remains part of historical Engineering State.

Where supersession is partial, the remaining applicable scope MUST be deterministically identifiable.

---

# Invalidated Decision

A Decision is invalidated when its normative effect is withdrawn because its validity, authority, basis, or applicability is no longer accepted.

Invalidation MUST identify the invalidating authority or Decision.

Invalidation MUST preserve:

* the original Decision;
* its original Outcome;
* its historical Applicability;
* the reason for invalidation;
* the invalidation time or condition.

Invalidation MUST NOT retroactively claim that the Decision was never applicable unless the applicable Product contract explicitly defines that legal or normative effect.

---

# Reopened Decision

A Decision MAY be reopened for reconsideration.

Reopening MUST NOT silently mutate the original Decision Outcome.

Reconsideration MAY result in:

* confirmation;
* a new Decision;
* supersession;
* invalidation;
* changed Applicability;
* additional Consequences.

The relationship between the original Decision and the resulting Decision MUST be explicit.

---

# Decision Relationship

A Decision Relationship is a typed relationship between a Decision and another Entity.

A Decision Relationship MUST identify:

* its source;
* its target;
* its Relation Type;
* its normative meaning;
* its applicability when conditional.

Decision Relation Types MAY include:

* depends on;
* constrains;
* authorizes;
* prohibits;
* implements;
* supersedes;
* invalidates;
* conflicts with;
* refines;
* delegates;
* responds to;
* justifies;
* is justified by;
* affects;
* requires;
* satisfies.

Decision relationships MUST NOT be inferred solely from textual proximity or repository structure.

---

# Decision Dependency

A Decision depends on another Decision when its validity, Applicability, Outcome, or Consequences rely on that Decision.

A dependency SHOULD identify what aspect is dependent.

Invalidation or supersession of a depended-on Decision MAY require review of dependent Decisions.

Such review MUST NOT automatically change dependent Decisions unless defined by an explicit rule or later Decision.

---

# Decision Conflict

A Decision Conflict exists when two applicable Decisions establish incompatible normative intent within overlapping Applicability.

A conflict MUST identify:

* the conflicting Decisions;
* the overlapping scope;
* the incompatible Outcomes or Engineering Commitments;
* the applicable authority or priority rules.

A Runtime MUST NOT silently resolve a Decision Conflict using storage order, creation time, or implementation convenience.

Conflict resolution MUST be established by:

* explicit priority;
* authority;
* supersession;
* scope refinement;
* another Decision;
* an applicable Product contract.

---

# Decision Hierarchy

A Product contract MAY define hierarchical relationships between Decisions.

A hierarchy MAY express:

* strategic and implementation Decisions;
* policy and local Decisions;
* parent and child Decisions;
* general and specialized Decisions;
* delegated Decisions.

A subordinate Decision MUST NOT contradict an applicable superior Decision unless it has explicit authority to override, specialize, waive, or supersede it.

Hierarchy MUST NOT be inferred solely from document nesting.

---

# Decision and Transition

A Decision is not a Lifecycle Transition.

An applicable Decision establishes normative intent according to its established Decision Outcome.

A Transition changes the Lifecycle State Assignment of a Lifecycle Subject according to a Lifecycle Definition.

A Decision MAY:

* authorize a Transition;
* prohibit a Transition;
* require a Transition;
* select a target State;
* provide required Transition Evidence;
* satisfy an authority condition;
* have no immediate Transition effect.

A Transition MAY occur without a new Decision when already authorized by an applicable Lifecycle Definition, Decision, policy, or other normative source.

A Decision MUST NOT be treated as completed implementation merely because it authorizes a Transition.

A Transition Record MAY reference a Decision when that Decision authorizes, requires, constrains, or explains the Transition.

---

# Decision and Artifact Revision

A Decision MAY require creation or modification of an Artifact.

A Decision MUST NOT itself create an Artifact Revision unless an Actor or Runtime performs the corresponding Revision Operation.

An Artifact Revision MAY:

* implement a Decision;
* represent a Decision;
* provide Evidence for a Decision;
* violate a Decision;
* supersede a prior implementation of a Decision.

Existence of an Artifact Revision does not prove conformance with a Decision.

---

# Decision and Validation

A Decision MAY define or imply validation expectations.

Validation MAY assess:

* whether the Decision Basis remains valid;
* whether an Engineering Commitment has been implemented;
* whether required Consequences occurred;
* whether the Decision Applicability is correctly interpreted;
* whether an implementation conforms to the Decision Outcome;
* whether assumptions remain valid.

Validation does not create or alter the Decision unless an explicit later Decision or Lifecycle Transition does so.

A failed Validation MAY trigger:

* review;
* reopening;
* invalidation;
* supersession;
* remediation;
* another Decision.

---

# Decision and Requirement

A Requirement is not a Decision.

A Requirement establishes a required need, condition, capability, or constraint.

A Decision determines normative engineering intent in response to a question, Requirement, context, or constraint.

A Requirement MAY:

* constrain a Decision;
* require a Decision;
* be interpreted by a Decision;
* be implemented through multiple Decisions;
* conflict with a Decision;
* be waived only through authorized mechanisms.

A Decision MUST NOT silently redefine an applicable Requirement unless it has explicit authority to do so.

---

# Decision and Risk

A Decision MAY accept, avoid, mitigate, transfer, or defer a risk.

A risk-related Decision SHOULD identify:

* the risk;
* the selected disposition;
* responsible authority;
* Applicability;
* assumptions;
* accepted consequences;
* review or expiration conditions.

Risk acceptance MUST NOT be inferred merely because no mitigation was implemented.

---

# Decision Integrity

Decision integrity requires that the identity, Outcome, authority, Applicability, Basis, and history of a Decision remain inspectable.

A conformant implementation MUST preserve enough information to determine:

* what was decided;
* who or what had authority;
* what scope was affected;
* why the Decision was made;
* which Decision Outcome became applicable;
* which Decisions superseded or invalidated it;
* which Artifacts or Transitions were related to it.

Decision integrity MUST NOT depend on undocumented Runtime behavior.

---

# Decision Invariants

## Identity Invariant

Every Decision has a stable Decision Identifier.

## Outcome Invariant

Every Decision has an identifiable proposed or established Decision Outcome.

Every applicable Decision has an established Decision Outcome.

## Subject Invariant

Every Decision identifies a Decision Subject or Decision Question.

## Applicability Invariant

Every applicable Decision has determinable Applicability.

## Authority Invariant

Every applicable Decision has an inspectable authority basis.

## Record Separation Invariant

A Decision is not equivalent to its Decision Record.

## Lifecycle Separation Invariant

A Decision Outcome is not equivalent to the lifecycle State of the Decision.

## Transition Separation Invariant

A Decision is not equivalent to a Lifecycle Transition.

## Evidence Separation Invariant

Evidence, assumptions, constraints, and Rationale are distinguishable.

## History Invariant

Supersession, invalidation, rejection, or withdrawal does not erase historical Decisions.

## No Silent Mutation Invariant

Material change to an applicable Decision Outcome requires an explicit new Decision or explicit normative relationship.

## Commitment Invariant

A Decision that establishes, changes, or removes an Engineering Commitment makes that effect inspectable.

## Runtime Authority Invariant

A Runtime does not acquire Decision Authority merely by executing or representing a Decision.

## Conflict Invariant

Conflicting applicable Decisions require explicit resolution.

---

# Runtime Independence

The Decision Model is independent of any particular Runtime.

A Runtime MAY:

* create Decision proposals;
* collect Evidence;
* evaluate guards;
* assist with Alternative analysis;
* generate Decision Records;
* execute authorized actions;
* create Transition Attempts;
* validate conformance;
* detect conflicts.

A Runtime MUST NOT:

* claim Decision Authority without explicit delegation;
* silently change a Decision Outcome;
* silently expand Decision Applicability;
* infer approval from undocumented behavior;
* erase rejected or superseded Decisions;
* resolve conflicts through storage order;
* reinterpret historical Decisions using current policy without recording that reinterpretation;
* treat implementation completion as proof that a Decision was validly authorized.

---

# Storage Independence

This specification does not require any particular storage mechanism.

Decisions MAY be represented using:

* text documents;
* structured records;
* graph stores;
* relational stores;
* event stores;
* source repositories;
* distributed systems;
* external systems.

Storage implementation MUST preserve normative identity and relationships.

Repository paths, filenames, database keys, display order, and modification times MUST NOT be treated as Decision semantics unless explicitly defined by the applicable Product contract.

---

# Extension Model

A Product contract MAY extend the Decision Model by defining:

* Decision Types;
* Decision significance levels;
* Decision lifecycle profiles;
* additional Decision roles;
* authority rules;
* approval policies;
* Alternative classifications;
* Decision Relation Types;
* required Decision Record fields;
* validation requirements;
* conflict-resolution policies;
* retention policies.

Extensions MUST NOT redefine the core meaning of:

* Decision;
* Decision Outcome;
* Decision Record;
* Decision Authority;
* Decision Applicability;
* Decision lifecycle;
* Engineering Commitment;
* supersession;
* invalidation.

Extensions MUST preserve the invariants of this specification.

---

# Conformance

An implementation conforms to this specification when it can represent and preserve:

* stable Decision identity;
* Decision Subject or Decision Question;
* Decision Outcome;
* Decision Applicability;
* Decision Authority;
* Decision lifecycle;
* Decision Basis where required;
* Decision Records where required;
* Decision relationships;
* supersession and invalidation history;
* separation between Decision, Decision Record, Outcome, and Transition.

A Product contract conforms to this specification when it:

* does not contradict the defined Decision semantics;
* defines required extensions explicitly;
* preserves Decision history;
* defines authority and applicability sufficiently;
* provides deterministic conflict handling;
* does not rely on undocumented storage or Runtime behavior for normative interpretation.

A Runtime conforms to this specification when it:

* preserves immutable historical records;
* enforces applicable authority rules;
* does not silently mutate Decisions;
* preserves Decision relationships;
* distinguishes proposals from applicable Decisions;
* does not treat Decision Records as self-authorizing.

---

# Non-Goals

This specification does not require:

* every engineering discussion to become a Decision;
* every Decision to enumerate Alternatives;
* every Decision to require approval;
* every Decision to create an Engineering Commitment;
* every Decision to cause a Transition;
* every Decision to be represented by a single document;
* every Decision to use the same lifecycle;
* every Decision to be made by a human;
* consensus;
* voting;
* a specific architecture decision record format;
* a specific governance methodology.

---

# References

* PEOS-000 — Overview
* PEOS-001 — Core Concepts
* PEOS-002 — Artifact Model
* PEOS-003 — Lifecycle
