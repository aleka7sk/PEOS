# PEOS-008 — Runtime Contract

**Category:** Normative

**Status:** Draft

**Version:** 0.1

---

# Abstract

This specification defines the PEOS Runtime Contract Model.

The Runtime Contract Model describes how a set of Requirements is bound to a runtime subject for enforcement, how runtime binding, observation, and violation are recorded, and how runtime compliance is derived without granting the runtime any authority over the Artifact, Requirement, Lifecycle, Decision, Validation, or Waiver models it consumes.

It distinguishes:

* a Runtime Contract from the Requirements it references;
* design-time Contract existence from runtime binding;
* runtime binding from Lifecycle activation;
* a Runtime Observation from the Evidence it may become;
* a Runtime Violation from a Decision Outcome;
* current runtime compliance from the historical record it is derived from.

A Runtime Contract is an Artifact. A Runtime Binding Record, Runtime Unbinding Record, Runtime Observation, and Runtime Violation are immutable records, not Artifacts. A Compliance Claim is a specialization of the Validation Claim defined by PEOS-006.

No runtime subject, Contract, or Contract Revision owns mutable compliance, binding, or violation state.

---

# Purpose

The purpose of the Runtime Contract Model is to provide a stable and implementation-independent foundation for:

* binding an identified set of Requirements to a runtime subject for enforcement, without duplicating Requirement machinery;
* recording runtime binding and unbinding as immutable, independently identifiable, append-only records;
* recording runtime observation and violation without granting them Artifact status or mutable subject-level fields;
* deriving runtime compliance from Compliance Claims, Violations, Observations, and applicable Waivers, rather than storing it;
* consuming Waiver semantics from PEOS-005 without redefining Waiver identity, lifecycle, revocation, or historical model;
* relating Runtime Contract to Artifact Lifecycle, Requirement Applicability, Allocation, Validation, Quality, and Decision concerns without redefining any of them.

---

# Scope

This specification defines:

* Runtime Contract and Runtime Contract Revision;
* Runtime Requirement Reference;
* Runtime Subject and Runtime Scope;
* Runtime Binding Record and Runtime Unbinding Record;
* derivation of Current Runtime Binding;
* Runtime Assertion;
* Runtime Observation and Runtime Evidence;
* Runtime Violation;
* Compliance Claim;
* derivation of runtime compliance;
* Runtime Waiver consumption;
* Runtime Remediation;
* the boundary between Runtime Contract and Artifact Lifecycle, Requirement Applicability, Allocation, Validation, Quality, and Decision concerns.

This specification does not define:

* Artifact, Artifact Revision, or Artifact Relation structure (PEOS-002);
* Lifecycle Definitions, States, or Transitions (PEOS-003);
* Decision structure or governance outcomes (PEOS-004);
* Requirement structure, Allocation, or Waiver structure (PEOS-005);
* the base Validation Claim, Validation Plan, or Evidence mechanism (PEOS-006);
* Quality Characteristics, Measures, or Profiles (PEOS-007);
* a general Incident Model;
* cross-artifact traceability paths, coverage, or orphan detection (a future Traceability Model);
* a mandatory deployment technology, infrastructure platform, or monitoring tool.

---

# Normative Foundation

The key words MUST, MUST NOT, SHALL, SHALL NOT, SHOULD, SHOULD NOT, MAY, REQUIRED, RECOMMENDED, and OPTIONAL in this specification are to be interpreted as defined by PEOS-000.

This specification builds upon:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model;
* PEOS-003 — Lifecycle;
* PEOS-004 — Decision Model;
* PEOS-005 — Requirement Model;
* PEOS-006 — Validation Model;
* PEOS-007 — Quality Model.

Terms defined by those specifications retain their normative meaning unless explicitly specialized here.

This specification does not redefine Artifact identity, Artifact Relation structure, Lifecycle semantics, Decision structure, Requirement structure, Waiver structure, or the Validation Claim mechanism.

---

# Runtime Model Overview

```text
Runtime Contract
    is an Artifact
    uses ordinary Artifact Revision
    references Requirements through Runtime Requirement References

Runtime Binding Record
    is an immutable record
    identifies exact Runtime Contract Artifact Revision and runtime subject

Runtime Unbinding Record
    is an immutable record
    references exactly one Binding Record

Current Runtime Binding
    is a derived view
    over Binding and Unbinding Record history

Runtime Observation
    is an immutable record
    may produce or reference Evidence Artifacts

Runtime Violation
    is an immutable record
    identifies the violated criterion and triggering Observation or Evidence

Compliance Claim
    specializes Validation Claim
    is not an Artifact
    has exactly one subject
```

The following relationships are normative:

```text
Runtime Contract Revision
    references
        one or more Requirements or Requirement Artifact Revisions

Runtime Binding Record
    binds
        exact Runtime Contract Artifact Revision
    to
        exact runtime subject

Runtime Unbinding Record
    terminates
        exactly one Runtime Binding Record

Runtime Observation
    supports or contradicts
        zero or more Runtime Assertions

Runtime Violation
    is triggered by
        one Observation or Evidence item

Compliance Claim
    cites
        zero or more Runtime Violations and Observations as Evidence
    concerns
        exactly one Subject
```

The following are distinct, independently inspectable conditions, and none implies another:

1. design-time existence of a Runtime Contract Artifact;
2. publication of a specific Runtime Contract Artifact Revision;
3. an optional State Assignment marking Lifecycle activation of the Runtime Contract, governed by PEOS-003;
4. a Runtime Binding Record connecting a Runtime Contract Artifact Revision to a runtime subject, governed by this specification;
5. actual runtime deployment or execution.

Lifecycle activation (item 3) and runtime binding (item 4) are not the same concern. Lifecycle activation is governed exclusively by PEOS-003. Runtime binding is governed exclusively by this specification.

---

# Runtime Contract

Runtime Contract SHALL be an Artifact, as defined by PEOS-002.

Runtime Contract SHALL use ordinary Artifact Revision. There is no Runtime Contract Version distinct from Artifact Revision.

A Runtime Contract is not a Requirement subtype. It does not duplicate Requirement Subject, Applicability, Origin, Authority, Classification, or Rationale machinery.

Every Runtime Contract Revision SHALL identify:

* the exact Requirements or Requirement Artifact Revisions it governs;
* the runtime subject type or target;
* the environment and deployment scope;
* the runtime assertions it defines, or the assertion definitions it references;
* its observation requirements;
* its violation classification rules;
* applicable Quality Profile Revisions, where required;
* applicable Waiver handling rules;
* enforcement expectations;
* provenance;
* authority;
* applicability.

A Runtime Contract Revision is immutable.

A newer Runtime Contract Revision does not mutate an earlier Revision.

---

# Runtime Contract Revision

Changes to any content listed under **Runtime Contract** constitute a content change and SHALL create a new Artifact Revision in accordance with PEOS-002.

Creation of a new Runtime Contract Revision does not, by itself, change which Revision is currently bound to a runtime subject. Binding to a specific Contract Revision is established exclusively by a Runtime Binding Record.

A Runtime Contract Revision that is not referenced by any Runtime Binding Record is a valid, inspectable, unbound Contract Revision.

---

# Runtime Requirement Reference

Requirements referenced by a Runtime Contract remain separate Requirement Artifacts, as defined by PEOS-005.

A Runtime Contract SHALL reference them through binary Artifact Relations, as defined by PEOS-002, or through explicit Revision-owned references consistent with PEOS-002's Artifact Relation contract.

Every reference SHALL preserve the exact participant level:

* Requirement identity, where revision-independent intent is deliberate;
* Requirement Artifact Revision, where exact wording or acceptance criteria govern runtime behavior.

Requirement Statement content SHALL NOT be copied into mutable Runtime Contract fields as an alternative to referencing the Requirement. The Runtime Contract references the Requirement; it does not restate or own the Requirement's content.

A Runtime Requirement Reference has no independent identity or lifecycle. It is either an Artifact Relation (per PEOS-002's general contract: exactly one source, exactly one target, no normative identity, no revisions, no lifecycle) or Revision-owned content within the Runtime Contract Revision.

---

# Runtime Subject

Runtime Subject means the concrete or logical running system, service, component, process, deployment, interface, actor, data flow, or other runtime target governed by a Runtime Contract.

Every Runtime Binding Record, Runtime Observation, Runtime Violation, and Compliance Claim SHALL identify its exact runtime subject and subject scope.

The following are distinguished and SHALL NOT be conflated:

* Artifact identity;
* Artifact Revision;
* runtime deployment;
* runtime process instance;
* infrastructure target.

A runtime subject is not itself an Artifact unless a specialized PEOS specification or Product contract explicitly represents it as one.

---

# Runtime Scope

Runtime Scope defines the boundary within which a Runtime Binding Record, Runtime Observation, Runtime Violation, or Compliance Claim applies.

Scope MAY include environment, deployment target, geographic or jurisdictional boundary, tenant, time window, or another explicitly identified constraint.

Scope SHALL be explicit wherever it is not self-evident from the identified runtime subject.

---

# Runtime Binding Record

A Runtime Binding Record is an immutable record.

A Runtime Binding Record is independently identifiable.

A Runtime Binding Record is not an Artifact. It is not revisioned. It is not lifecycle-bearing. It is not a State Assignment.

Every Runtime Binding Record SHALL identify:

* Binding Record identity;
* the exact Runtime Contract Artifact Revision bound;
* the exact runtime subject or deployment target;
* environment;
* scope;
* binding timestamp;
* deployment timestamp, where distinct from the binding timestamp;
* actor;
* authority, where required;
* provenance;
* configuration or deployment reference, where required;
* known limitations;
* any previous Binding Record it explicitly corrects, replaces, or invalidates.

A Runtime Binding Record does not mutate the Runtime Contract Revision it binds.

A Runtime Binding Record does not, by itself, establish Requirement satisfaction or runtime compliance.

A new Runtime Binding Record MAY explicitly `correct`, `replace`, or `invalidate` an earlier Runtime Binding Record. Such a reference SHALL identify the earlier record exactly. The earlier record remains historically preserved.

Record replacement SHALL NOT be described using the normative term Supersession, except when explicitly explaining that PEOS-002 Artifact Supersession does not apply to Runtime Binding Records.

---

# Runtime Unbinding Record

A Runtime Unbinding Record is an immutable record.

A Runtime Unbinding Record is independently identifiable.

A Runtime Unbinding Record references exactly one Runtime Binding Record.

A Runtime Unbinding Record is not an Artifact. It is not revisioned. It is not lifecycle-bearing.

Every Runtime Unbinding Record SHALL identify:

* Unbinding Record identity;
* the exact Runtime Binding Record affected;
* the runtime subject;
* termination timestamp;
* the reason for termination;
* actor;
* authority, where required;
* provenance;
* a correction/replacement reference, where applicable.

A Runtime Unbinding Record does not delete, erase, or rewrite its Binding Record. The Binding Record remains historically inspectable after unbinding.

---

# Current Runtime Binding

Current Runtime Binding is a derived view. It is not a stored field.

Current Runtime Binding SHALL be derived from:

* applicable Runtime Binding Records;
* applicable Runtime Unbinding Records;
* correction, replacement, and invalidation references among those records;
* the runtime subject in question;
* environment;
* scope;
* the time at which the determination is made;
* governing Product rules.

The following mutable fields SHALL NOT be created:

```text
RuntimeContract.bound
RuntimeContract.activeDeployment
ArtifactRevision.deployed
```

Lifecycle activation, as defined by PEOS-003 through a State Assignment, and runtime binding, as defined by this specification through Binding and Unbinding Records, remain distinct. A Runtime Binding Record is not a State Assignment, and a State Assignment does not itself establish or terminate a Runtime Binding.

A State Assignment MAY be an additional Product-defined eligibility condition for Current Runtime Binding, where the applicable Product contract explicitly requires it. A State Assignment is not inherently part of Binding Record history.

---

# Runtime Assertion

A Runtime Assertion is a Revision-owned rule or derived evaluation rule. It is content within the Runtime Contract Revision that defines it, or an explicitly referenced definition.

A Runtime Assertion has no required independent identity, no revisions, and no lifecycle of its own.

Every Runtime Assertion SHALL identify:

* the evaluated runtime subject;
* the Requirement or Contract criterion it evaluates;
* its observation inputs;
* its evaluation rule;
* its expected result;
* its scope;
* temporal conditions, where applicable;
* uncertainty handling, where applicable.

A Runtime Assertion SHALL NOT be modeled as an independently lifecycle-bearing entity.

---

# Runtime Observation

A Runtime Observation is an immutable record.

A Runtime Observation is independently identifiable.

A Runtime Observation is not an Artifact.

Every Runtime Observation SHALL identify:

* Observation identity;
* the runtime subject;
* the exact Runtime Binding Record, where applicable;
* timestamp or time interval;
* the observed value or event;
* unit, scale, or event type, where applicable;
* environment or context;
* collection method;
* the actor or system source;
* known uncertainty;
* known limitations;
* provenance;
* Evidence Artifact Revision references, where normative Evidence is required.

A Runtime Observation is not automatically an Evidence Artifact merely by being recorded.

Where a Runtime Observation must serve as normative PEOS Evidence, it SHALL be captured or referenced through an Artifact satisfying the Evidence role defined by PEOS-002. This specification does not redefine Evidence.

---

# Runtime Evidence

Runtime Evidence is the specific application of the PEOS-002 Evidence role, and the qualification rules of PEOS-006, to material produced by or relevant to a runtime subject.

Runtime Evidence SHALL be cited at exact Artifact Revision level, in accordance with PEOS-006.

A Runtime Observation that has not been captured or referenced as an Artifact remains an immutable record; it does not, by that fact alone, qualify as normative Evidence for a Compliance Claim or any other Validation Claim.

---

# Runtime Violation

A Runtime Violation is an immutable record.

A Runtime Violation is independently identifiable.

A Runtime Violation is not an Artifact. It is not revisioned. It is not lifecycle-bearing.

Every Runtime Violation SHALL identify:

* Violation identity;
* the runtime subject;
* the exact Runtime Binding Record, where applicable;
* the violated Requirement, Requirement Artifact Revision, Runtime Contract rule, Quality criterion, or Runtime Assertion;
* the triggering Observation or Evidence;
* timestamp or interval;
* its violation classification;
* severity, where applicable;
* scope;
* provenance;
* known uncertainty and limitations;
* an applicable Waiver, if any;
* related Claims or Decisions, where applicable.

A Runtime Violation is not:

* a mutable Runtime Contract field;
* a Lifecycle State;
* a Decision Outcome;
* a Validation Claim;
* an Incident, unless another explicitly defined model establishes an Incident construct.

This specification does not introduce a general Incident entity.

---

# Compliance Claim

A Compliance Claim is a specialization of Validation Claim, as defined by PEOS-006.

A Compliance Claim inherits, without redefinition:

* immutable record semantics;
* independent Claim identity;
* the absence of Artifact status;
* the absence of revisions;
* the absence of lifecycle;
* the requirement of exactly one Subject;
* the separation of criteria from Subject;
* exact Evidence Artifact Revision references;
* correction, replacement, and invalidation history (`new Claim → affected earlier Claim`, never Artifact Supersession);
* derived-current-effect semantics.

A Compliance Claim's Subject MAY be:

* a runtime subject;
* a Runtime Contract Artifact;
* a Runtime Contract Artifact Revision;
* a deployed Artifact Revision;
* another explicitly identified engineering subject.

A Compliance Claim's criteria MAY include:

* a Requirement;
* a Requirement Artifact Revision;
* a Runtime Contract rule;
* a Runtime Assertion;
* a Quality Characteristic;
* a Quality Measure;
* applicable Waiver conditions.

This specification does not define a second runtime-specific Claim base model. A Compliance Claim exists exclusively as an instance of the PEOS-006 Validation Claim mechanism.

---

# Derived Runtime Compliance

Runtime compliance is a derived view. It is not a stored field.

Runtime compliance SHALL be derived from:

* applicable Compliance Claims;
* applicable Runtime Violations;
* applicable Runtime Observations;
* normative Evidence;
* Claim replacement and invalidation history;
* applicable Runtime Binding and Unbinding Records;
* applicable Waivers;
* the exact Runtime Contract Revision in force;
* the exact Requirement criteria evaluated;
* time;
* environment;
* scope;
* governing Product rules.

The following mutable fields SHALL NOT be created:

```text
RuntimeContract.compliant
RuntimeContractRevision.compliant
RuntimeSubject.compliant
Deployment.compliant
```

The existence of a Runtime Violation does not automatically establish globally current non-compliance outside its explicit subject, scope, criteria, and time. Determining current compliance for a given subject, scope, and criteria requires evaluating the applicable, non-invalidated history described above at query time.

---

# Runtime Waiver

PEOS-008 SHALL consume Waiver semantics from PEOS-005. This specification does not define a new Runtime Waiver entity, and does not define Waiver identity, Waiver lifecycle, Waiver revocation, or a Waiver historical model.

A Decision Outcome MAY authorize a Waiver, in accordance with PEOS-005.

A runtime evaluation MAY consider an applicable Waiver in its derivation of compliance only where:

* the Waiver's scope matches the evaluated runtime subject and criteria;
* the Waiver's authority is valid, in accordance with PEOS-005;
* the Waiver's temporal applicability matches the time of evaluation;
* any conditions attached to the Waiver are met;
* the affected Requirement or criterion is explicitly identified by the Waiver.

A Waiver does not erase Runtime Observations or Runtime Violations. It affects only the derived interpretation of compliance, or a permitted consequence, without deleting or rewriting the historical record.

---

# Runtime Remediation

Remediation that requires a governance choice SHALL be represented through a Decision or Decision Outcome governed by PEOS-004.

A Runtime Violation MAY be referenced by:

* a Decision Basis;
* a corrective action plan;
* an Engineering Commitment;
* a follow-up Validation Plan, as defined by PEOS-006.

A Runtime Violation does not, by itself, authorize remediation. Authorization requires an applicable Decision Outcome.

---

# Runtime and Artifact Lifecycle

Requirement applicability (PEOS-005), Artifact Lifecycle State (PEOS-003), Lifecycle activation of a Runtime Contract (PEOS-003), runtime binding (this specification), runtime deployment (this specification), Allocation (PEOS-005), Validation Claim outcome (PEOS-006), conformance (PEOS-006 Conformance Claim), and runtime compliance (this specification) are nine distinct concerns. None is interchangeable with another.

Runtime binding is governed by PEOS-008. Lifecycle activation is governed by PEOS-003.

A State Assignment marking Lifecycle activation for a Runtime Contract is optional unless the applicable Product contract or Lifecycle Definition explicitly requires it.

A Runtime Binding Record does not create, imply, or replace a State Assignment. A State Assignment does not create, imply, or replace a Runtime Binding Record.

A Runtime Contract MAY be bound regardless of whether it participates in a Lifecycle, unless a governing Product rule explicitly requires a particular Lifecycle State.

Runtime deployment does not mutate an Artifact Revision. A deployed Artifact Revision remains exactly the immutable Revision it was before deployment.

A Lifecycle Transition MAY require a Runtime Binding Record or Compliance Claim as Transition Evidence, in accordance with PEOS-003 and PEOS-006. Lifecycle effects remain exclusively governed by PEOS-003.

---

# Runtime and Requirement Applicability

A Requirement referenced by a Runtime Contract retains its own Applicability, as defined by PEOS-005 §11. Runtime binding does not establish, override, or substitute for Requirement Applicability.

A Requirement that is not applicable, per PEOS-005, does not become enforceable at runtime merely because it is referenced by a bound Runtime Contract.

---

# Runtime and Allocation

Allocation, as defined by PEOS-005 §24, assigns realization or responsibility for a Requirement. Runtime binding, as defined by this specification, deploys enforcement of a Requirement to a runtime subject.

Allocation does not imply runtime binding. Runtime binding does not imply Allocation. The two SHALL remain independently inspectable.

---

# Runtime and Validation

Runtime Observation and Runtime Violation MAY serve as, or produce, Evidence for a Validation Claim, subject to the Evidence qualification rules of PEOS-006.

Compliance Claim is the sole Claim type this specification defines, and it is a specialization of the Validation Claim mechanism owned by PEOS-006. This specification does not define an independent Activity, Evidence, or Claim base mechanism.

---

# Runtime and Quality

A Runtime Contract Revision MAY reference an applicable Quality Profile Revision, as defined by PEOS-007, as part of its enforcement expectations.

A Compliance Claim's criteria MAY include a Quality Characteristic or Quality Measure. This specification does not define Quality Characteristics, Measures, or Profiles; those remain owned exclusively by PEOS-007.

---

# Runtime and Decisions

Certification, acceptance, approval, and authorization concerning runtime enforcement are governance outcomes governed by PEOS-004, expressed through a Decision Outcome.

A Runtime Violation is not a Decision Outcome. A Compliance Claim is not a Decision Outcome. Neither replaces the governance judgment that only a Decision Outcome can express.

---

# Runtime Invariants

A conformant implementation MUST preserve the following invariants.

## Runtime Contract Artifact Invariant

Every Runtime Contract is an Artifact using ordinary Artifact Revision.

## No Runtime Contract Version Invariant

Runtime Contract evolution uses Artifact Revision; there is no separate Runtime Contract Version mechanism.

## Requirement Reference Level Invariant

Every Runtime Requirement Reference identifies its exact participant level (Requirement identity or Requirement Artifact Revision).

## Runtime Binding Record Immutability Invariant

A recorded Runtime Binding Record does not change. Correction produces a new record.

## Runtime Unbinding Preservation Invariant

A Runtime Unbinding Record does not delete or rewrite the Binding Record it terminates.

## Derived Current Binding Invariant

Current Runtime Binding is always derived from Binding and Unbinding Record history; it is never a stored field.

## Runtime Binding and Lifecycle Separation Invariant

A Runtime Binding Record is never represented as a State Assignment.

## Runtime Observation Immutability Invariant

A recorded Runtime Observation does not change.

## Runtime Evidence Ownership Invariant

Runtime Evidence conforms to the Evidence role owned by PEOS-002 and the qualification rules owned by PEOS-006; it is not redefined here.

## Runtime Violation Immutability Invariant

A recorded Runtime Violation does not change.

## Compliance Claim Specialization Invariant

Every Compliance Claim is a Validation Claim as defined by PEOS-006, without redefinition of Claim mechanics.

## Single Compliance Claim Subject Invariant

Every Compliance Claim identifies exactly one Subject.

## Derived Runtime Compliance Invariant

Runtime compliance is always derived from applicable Compliance Claims, Violations, Observations, and Waivers; no Subject owns a mutable field representing it.

## Runtime Waiver Consumption Invariant

This specification consumes, and never redefines, Waiver identity, lifecycle, revocation, or historical model as owned by PEOS-005.

## Runtime and Decision Separation Invariant

A Runtime Violation or Compliance Claim never itself constitutes a Decision Outcome.

## No Mutable Runtime Truth Invariant

No runtime subject, Runtime Contract, or Runtime Contract Revision owns a mutable field representing binding, compliance, or violation state.

---

# Non-Conforming Patterns

The following implementation patterns violate this specification.

## Runtime Contract Version

Introducing a revision system for Runtime Contract distinct from ordinary Artifact Revision.

## Runtime Contract as Requirement Subtype

Modeling Runtime Contract as a specialized Requirement rather than an Artifact that references separate Requirement Artifacts.

## Requirement Statement Copied as Mutable Runtime State

Copying Requirement Statement content into a mutable Runtime Contract field instead of referencing the Requirement.

## Runtime Binding as State Assignment

Representing runtime binding through a State Assignment instead of a Runtime Binding Record.

## Mutable Binding Record

Modifying a recorded Runtime Binding Record in place instead of recording a new one.

## Mutable Current Binding Flag

Storing current runtime binding as a mutable field such as `RuntimeContract.bound` instead of deriving it from Binding and Unbinding Record history.

## Runtime Observation as Automatically Normative Evidence

Treating a Runtime Observation as normative PEOS Evidence without capturing or referencing it through an Artifact satisfying the PEOS-002 Evidence role.

## Mutable Observation

Modifying a recorded Runtime Observation in place instead of recording a new one.

## Mutable Violation

Modifying a recorded Runtime Violation in place instead of recording a new one.

## Violation as Lifecycle State

Representing a Runtime Violation as a Lifecycle State or State Assignment.

## Violation as Decision Outcome

Treating the existence of a Runtime Violation as itself a Decision Outcome or governance authorization.

## Runtime Incident Entity Introduced Implicitly

Introducing an Incident construct without an explicit, separately defined model for it.

## Parallel Runtime Claim Base

Defining a Compliance Claim that does not specialize Validation Claim, or that redefines Claim identity, immutability, or replacement semantics.

## Compliance Claim as Artifact

Representing a Compliance Claim as an Artifact.

## Composite Compliance Claim Subject

Identifying more than one Subject on a single Compliance Claim.

## Mutable Runtime Compliance

Storing runtime compliance as a mutable field such as `RuntimeContract.compliant` instead of deriving it from applicable Claims, Violations, and Observations.

## Runtime Waiver Redefinition

Defining Waiver identity, lifecycle, revocation, or historical model within PEOS-008 instead of consuming PEOS-005's definition.

## Waiver Erasing Violation History

Using an applicable Waiver to delete, rewrite, or hide a recorded Runtime Violation or Runtime Observation instead of affecting only the derived interpretation of compliance.

## Runtime Outcome Used as Governance Authority

Treating a Compliance Claim's outcome or a Runtime Violation's existence as itself sufficient authority to establish, change, or remove an Engineering Commitment, bypassing the Decision Model.

---

# Conformance

An implementation conforms to this specification when it can represent and preserve:

* Runtime Contracts as Artifacts using ordinary Artifact Revision;
* Runtime Requirement References at an exact, explicit participant level;
* Runtime Binding Records and Runtime Unbinding Records as immutable, independently identifiable, non-Artifact records;
* Runtime Observations and Runtime Violations as immutable, independently identifiable, non-Artifact records;
* Compliance Claims as immutable specializations of Validation Claim, with exactly one Subject;
* derived current binding and derived runtime compliance, with no mutable field on any Subject;
* Waiver consumption without redefinition of Waiver structure.

A Product contract conforms to this specification when it does not contradict the defined Runtime semantics and does not require a mutable binding, compliance, or violation field on any Subject.

A Runtime conforms to this specification when it does not mutate a recorded Binding Record, Unbinding Record, Observation, Violation, or Compliance Claim, and does not present derived runtime state as though it were stored authoritative state.

---

# Non-Goals

This specification does not require:

* every Requirement to be governed by a Runtime Contract;
* every runtime deployment to produce a Runtime Binding Record before this specification applies retroactively;
* a specific monitoring, observability, or enforcement technology;
* a general Incident Management Model;
* automatic remediation;
* a mandatory relationship between Runtime Violations and Lifecycle Transitions;
* a general traceability, coverage, or orphan-detection mechanism.

---

# References

This document depends on:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model;
* PEOS-003 — Lifecycle;
* PEOS-004 — Decision Model;
* PEOS-005 — Requirement Model;
* PEOS-006 — Validation Model;
* PEOS-007 — Quality Model.

This document provides the Runtime foundation for:

* PEOS-009 — Template Contract, where a Runtime Contract MAY itself be a Template-generated Artifact.
