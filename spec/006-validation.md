# PEOS-006 — Validation Model

**Category:** Normative

**Status:** Draft

**Version:** 0.1

---

# Abstract

This specification defines the PEOS Validation Model.

The Validation Model describes how engineering claims about the satisfaction, conformance, and quality of an identified engineering subject are planned, executed, evidenced, and recorded.

It distinguishes:

* a Validation Plan from the Validation Activities it plans;
* a planned Validation Activity from its execution;
* an execution from the Evidence it produces or relies upon;
* Evidence from the Claim that interprets it;
* a Validation Claim from the Requirement, Artifact, Decision, or other subject it concerns;
* a Claim's recorded outcome from the subject's current derived status.

A Validation Claim is not an Artifact, a Requirement, a Decision, or a State Assignment.

No evaluated subject owns mutable satisfaction, conformance, or quality state. Every such state is a derived view over applicable Claims.

---

# Purpose

The purpose of the Validation Model is to provide a stable and implementation-independent foundation for:

* identifying what MAY be validated, and at which participant level;
* planning Validation Activities without granting them independent identity beyond their owning Plan;
* recording the execution of Validation Activities as immutable, independently identifiable records;
* qualifying which Artifacts serve as Evidence for validation purposes;
* recording Validation Claims as immutable, independently identifiable assertions;
* deriving satisfaction, conformance, and quality without mutable state on the evaluated subject;
* relating Validation to Lifecycle, Decision, Requirement, Quality, Runtime, and Template concerns without redefining any of them.

---

# Scope

This specification defines:

* Validation Subject and participant level;
* Validation Plan;
* Planned Validation Activity;
* Validation Method;
* Validation Execution Record;
* Evidence qualification for validation use;
* Validation Claim, including Satisfaction Claim and Conformance Claim;
* Claim correction, replacement, and invalidation;
* derived validation state;
* Validation invariants and non-conforming patterns.

This specification does not define:

* Requirement structure (PEOS-005);
* Decision structure (PEOS-004);
* Lifecycle structure (PEOS-003);
* Artifact structure or the Evidence role itself (PEOS-002);
* Quality Characteristics, Measures, or Profiles (PEOS-007);
* Runtime enforcement (PEOS-008);
* Template structure (PEOS-009);
* cross-artifact traceability paths, coverage, or orphan detection (a future Traceability Model);
* a mandatory validation methodology, tool, or algorithm.

---

# Normative Foundation

The key words MUST, MUST NOT, SHALL, SHALL NOT, SHOULD, SHOULD NOT, MAY, REQUIRED, RECOMMENDED, and OPTIONAL in this specification are to be interpreted as defined by PEOS-000.

This specification builds upon:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model;
* PEOS-003 — Lifecycle;
* PEOS-004 — Decision Model;
* PEOS-005 — Requirement Model.

Terms defined by those specifications retain their normative meaning unless explicitly specialized here.

This specification does not redefine Artifact identity, Artifact Relation structure, Lifecycle semantics, Decision structure, or Requirement structure.

---

# Validation Model Overview

```text
Validation Plan
    is an Artifact
    uses ordinary Artifact Revision
    contains Planned Validation Activities as Revision-owned content

Planned Validation Activity
    belongs to exactly one Validation Plan Revision
    has a stable plan-local key
    has no independent PEOS identity
    has no revisions
    has no lifecycle

Validation Execution Record
    is an immutable record
    is independently identifiable
    is not an Artifact
    represents one attempted execution
    may reference a Planned Validation Activity by exact Plan Revision and plan-local key
    may instead represent an explicitly identified ad hoc execution

Evidence
    is an Artifact role owned by PEOS-002
    is cited by exact Artifact Revision

Validation Claim
    is an immutable record
    is independently identifiable
    is not an Artifact
    has exactly one subject and zero or more criteria
    identifies its basis, Evidence, method, provenance, timestamp, and authority where required

Claim Replacement
    is a minimal record-to-record reference
    is not an Artifact Relation
    is not Artifact Supersession
```

The following relationships are normative:

```text
Validation Plan Revision
    contains
        one or more Planned Validation Activities

Planned Validation Activity
    is executed by
        zero or more Validation Execution Records

Validation Execution Record
    produces or relies upon
        zero or more Evidence Artifact Revisions

Validation Claim
    cites
        one or more Evidence Artifact Revisions
    concerns
        exactly one Subject at an exact participant level
    evaluates
        zero or more Criteria

New Validation Claim
    MAY correct, replace, or invalidate
        an earlier Validation Claim
```

A Validation Plan, a Planned Validation Activity, an Execution Record, and a Validation Claim are four distinct concepts. None is a substitute for another.

A Validation Claim is not required for every Validation Execution Record to exist, and a Validation Execution Record is not required for every Validation Claim to exist. A Claim MAY be based on Evidence produced outside any recorded Execution, provided that Evidence satisfies the qualification rules defined in this specification.

---

# Validation Subject

A Validation Subject is the exact engineering entity, at an exact participant level, that a Planned Validation Activity, a Validation Execution Record, or a Validation Claim concerns.

Permitted Validation Subjects include:

* an Artifact;
* an Artifact Revision;
* a Requirement (identity level);
* a Requirement Artifact Revision;
* a Decision;
* a Decision Outcome;
* an Engineering Commitment.

A Planned Validation Activity, a Validation Execution Record, and a Validation Claim SHALL each identify the exact participant level of their Subject.

Where content is being evaluated, the Subject SHALL be identified at the Artifact Revision or Requirement Artifact Revision level, not merely at the Artifact or Requirement identity level.

Where the evaluated concern applies to the identity independently of any individual content revision, identifying the Artifact identity or Requirement identity is sufficient.

Ambiguous phrasing such as "validates the Requirement," when revision-specific content is actually intended, SHALL NOT be used. The exact Requirement Artifact Revision SHALL be identified instead.

This specification does not define validation of Runtime occurrences directly. A Runtime Observation MAY serve as Evidence within a Validation Claim whose Subject is an Artifact, Artifact Revision, Requirement, or Requirement Artifact Revision governed by an applicable Runtime Contract, as defined by PEOS-008.

---

# Validation Plan

A Validation Plan is an Artifact as defined by PEOS-002.

A Validation Plan uses ordinary Artifact Revision for all of its evolution. There is no Validation Plan Version distinct from Artifact Revision.

Every Validation Plan Revision SHALL identify:

* its intended scope;
* the Planned Validation Activities it contains;
* the Subjects those Activities reference;
* the Validation Methods those Activities use;
* the criteria those Activities evaluate;
* the Evidence expected by those Activities;
* the acceptance or evaluation rules applicable to those Activities;
* sequencing and dependencies among its Planned Validation Activities, where applicable;
* responsible actors or roles, where required by the applicable Product contract;
* applicability;
* provenance.

Modification of a Validation Plan's content constitutes a content change and SHALL create a new Artifact Revision in accordance with PEOS-002.

Creation of a new Validation Plan Revision does not mutate, delete, or reinterpret a previous Validation Plan Revision. A Validation Execution Record or Validation Claim that referenced an earlier Plan Revision remains correctly attributed to that earlier Revision.

---

# Planned Validation Activity

A Planned Validation Activity is an Artifact Revision-owned value structure.

A Planned Validation Activity belongs to exactly one Validation Plan Revision.

A Planned Validation Activity has no independent PEOS identity. It has no revisions and no lifecycle of its own; its evolution is entirely governed by the revision of its owning Validation Plan.

Every Planned Validation Activity SHALL have a stable plan-local key.

A stable plan-local key:

* is unique only within the owning Validation Plan Revision;
* is used by sequencing, dependencies, Execution Records, required-evidence rules, and Claim basis to reference the Planned Validation Activity;
* does not survive as an independent identity outside that exact Plan Revision;
* is not an Artifact Identity;
* is not a global Validation Activity Identity.

A new Validation Plan Revision MAY reuse, remove, or reintroduce a plan-local key. A plan-local key from an earlier Plan Revision SHALL NOT be assumed to refer to the same Planned Validation Activity in a later Plan Revision unless the applicable Product contract explicitly defines that continuity.

Every Planned Validation Activity SHALL identify, at minimum:

* its Subject and exact participant level;
* its Validation Method;
* the criteria it evaluates;
* the Evidence expected to support its execution;
* its execution prerequisites;
* its sequencing or dependencies, where applicable;
* the responsible actor or role, where required;
* the authority required to establish its result as applicable, where required;
* how its expected outcome is to be interpreted.

A Planned Validation Activity does not, by itself, constitute an executed Validation. Execution is recorded exclusively by a Validation Execution Record.

---

# Validation Method

A Validation Method is a controlled vocabulary/type. It defines the semantic meaning of an approach used to plan or execute a Validation Activity.

A Validation Method has stable semantics within its governing scope (an applicable Product contract, an explicitly referenced method definition, or this specification's own illustrative vocabulary).

A Validation Method does not require independent Artifact identity. There is no Validation Method Version.

An external or PEOS Artifact MAY define detailed procedure content for a Validation Method. This specification does not require every Method Type to be its own Artifact.

Method Types include, at minimum:

* inspection;
* analysis;
* demonstration;
* test;
* review;
* verification;
* validation.

This list is illustrative and is not a closed vocabulary.

Certification and Acceptance are not Validation Methods. They are governance outcomes governed by PEOS-004, established through a Decision Outcome that MAY reference one or more Validation Claims.

---

# Validation Execution Record

A Validation Execution Record is an immutable record.

A Validation Execution Record is independently identifiable.

A Validation Execution Record is not an Artifact. It is not revisioned. It is not lifecycle-bearing.

A Validation Execution Record represents one attempted execution of a Validation Activity.

A Validation Execution Record MAY reference one Planned Validation Activity by identifying:

* the exact Validation Plan Revision;
* the plan-local key of the Planned Validation Activity within that Revision.

A Validation Execution Record MAY instead represent an explicitly identified ad hoc execution that was not planned by any Validation Plan.

Every Validation Execution Record SHALL identify, at minimum:

* its own record identity;
* its exact Planned Activity reference (Plan Revision and plan-local key), or an explicit ad hoc designation;
* its Subject and exact participant level;
* the Validation Method used;
* the criteria evaluated;
* the started timestamp, where known;
* the completed or terminated timestamp;
* its outcome;
* its event history, where required by the applicable Product contract;
* the Evidence Artifact Revisions it produced or relied upon;
* the responsible actor;
* the authority basis, where required;
* the environment or execution context;
* known limitations;
* known uncertainty;
* provenance.

Correction of a Validation Execution Record creates a new Validation Execution Record. A Validation Execution Record SHALL NOT be mutated once recorded.

A new Validation Execution Record MAY explicitly correct, replace, or invalidate an earlier Validation Execution Record. Such a reference SHALL identify the earlier record exactly. The earlier record remains historically preserved; it is not erased or overwritten.

This correction, replacement, or invalidation relationship is not Artifact Supersession. The normative term Supersession SHALL NOT be used for Validation Execution Record correction or replacement, except when explicitly explaining that Artifact Supersession does not apply.

---

# Execution Outcome and Event History

Every Validation Execution Record SHALL identify an outcome drawn from an extensible controlled vocabulary including, at minimum:

* completed;
* failed;
* interrupted;
* indeterminate.

These outcomes are values recorded on the immutable Execution Record. They are not Lifecycle States, and they SHALL NOT be represented as Lifecycle States or through a State Assignment.

Event history, when required by the applicable Product contract, is an ordered sequence of observations recorded within the same immutable Execution Record (or a series of immutable Execution Records representing successive attempts). Event history is not itself a Lifecycle State History as defined by PEOS-003, though a Lifecycle Transition MAY reference a Validation Execution Record as Transition Evidence.

An indeterminate or interrupted outcome SHALL NOT be silently treated as completed.

---

# Evidence

Evidence remains an Artifact role owned by PEOS-002. This specification does not redefine Evidence as a new entity.

An Artifact serves as Evidence when it supports or contradicts an engineering claim, in accordance with PEOS-002.

This specification defines the conditions under which an Artifact qualifies as valid Evidence for validation use, as anticipated by PEOS-002.

Validation Evidence SHALL be:

* an Artifact used in the Evidence role, as defined by PEOS-002;
* cited at the exact Artifact Revision level;
* relevant to the identified Subject and criteria of the Planned Validation Activity, Validation Execution Record, or Validation Claim it supports;
* temporally applicable to the evaluation it supports;
* sufficiently attributable to allow inspection of its origin and provenance;
* integrity-preserved, in accordance with PEOS-002;
* accompanied by known limitations and known uncertainty where material to the evaluation.

External material MUST be captured by, or normatively referenced through, an Artifact before it can serve as normative PEOS Evidence. Material that has not been captured or referenced as an Artifact SHALL NOT be cited as Validation Evidence.

The following are distinguished:

* Evidence produced by a Validation Execution Record (an output of executing a Validation Activity);
* Evidence consumed by a Validation Execution Record (an input relied upon during execution);
* Evidence cited by a Validation Claim (the basis for the Claim's outcome).

An Observation or a Result is not a separate identity-bearing category distinct from Evidence. A recorded observation or measurement either is represented as an Evidence Artifact, or it is represented as content within an immutable Validation Execution Record. No third category is introduced.

---

# Validation Claim

A Validation Claim is an immutable record.

A Validation Claim is independently identifiable.

A Validation Claim is not an Artifact. It is not revisioned. It is not lifecycle-bearing.

A Validation Claim has exactly one engineering subject and zero or more criteria.

A Validation Claim preserves its historical assertion permanently. It SHALL NOT be mutated once recorded.

Every Validation Claim SHALL identify:

* Claim identity;
* Claim type;
* exactly one Subject;
* the exact participant level of that Subject;
* explicit scope;
* zero or more criteria;
* outcome;
* the Validation Method or evaluation rule applied;
* the exact Evidence Artifact Revisions relied upon;
* relevant Validation Execution Records, where applicable;
* reasoning or interpretation necessary to connect those inputs to the outcome, where the outcome is not mechanically determined;
* provenance;
* timestamp;
* authority, where required by applicable governance.

A Validation Claim is the sole mechanism by which satisfaction, conformance, or quality assertions are recorded. No evaluated Subject owns a mutable field representing satisfaction, conformance, or quality; see **Derived Validation State**.

---

# Claim Subject and Participant Level

A Claim Subject is exactly one engineering subject.

Permitted Claim Subjects include, where semantically applicable:

* an Artifact;
* an Artifact Revision;
* a Requirement;
* a Requirement Artifact Revision;
* a Decision;
* a Decision Outcome;
* an Engineering Commitment.

Every Validation Claim SHALL state the exact participant level of its Subject (identity level or revision level), mirroring the participant-level discipline established by PEOS-005 §17.3 for Requirement relationships.

A Validation Claim SHALL NOT identify more than one Subject. Where an evaluation concerns multiple engineering entities jointly, it SHALL be represented through multiple Validation Claims, each with exactly one Subject, optionally cross-referenced through their criteria or basis.

---

# Claim Criteria

Claim criteria are zero or more explicitly identified evaluation criteria against which the Claim's Subject is evaluated.

A criterion MAY identify:

* a Requirement;
* a Requirement Artifact Revision;
* a Quality Characteristic;
* a Quality Measure;
* a threshold;
* a target;
* a Runtime Contract rule;
* a Template constraint;
* a Product contract rule;
* another explicitly defined evaluation rule.

A criterion is not a second Claim Subject. A criterion SHALL NOT be used to smuggle a composite subject into a Claim that is required to have exactly one Subject.

Where zero criteria are identified, the Claim's outcome SHALL be interpreted strictly according to its stated Validation Method and basis.

---

# Claim Outcome

Every Validation Claim SHALL identify an outcome drawn from an extensible controlled vocabulary including, at minimum:

* satisfied;
* not satisfied;
* inconclusive.

A specialized Claim Type (such as a Quality Claim or a Compliance Claim, defined by other PEOS specifications) MAY define additional outcome values, provided each additional value is unambiguously mapped to one of the general outcome semantics above.

There is no separate Verdict entity. The outcome recorded on a Validation Claim is the complete and only representation of what was determined; it is not restated or re-derived through a second construct.

---

# Claim Basis

Claim Basis is a derived or grouped view over the Validation Method or evaluation rule applied, the criteria evaluated, the Evidence relied upon, relevant Validation Execution Records, and the reasoning or interpretation necessary to connect those inputs to the Claim's outcome.

Claim Basis is not an independent opaque field distinct from the fields it groups. This specification does not require an additional required field named "basis" beyond the individually identified method, criteria, Evidence, Execution Records, and reasoning listed under **Validation Claim**.

The term Claim Basis MAY remain in use as the collective name for those inputs. It does not introduce independent Claim Basis identity, revision, or lifecycle.

Claim Basis MUST be distinguishable from Claim Outcome. Identifying the basis does not, by itself, establish the outcome; the outcome is the recorded determination reached from that basis.

---

# Claim Correction, Replacement, and Invalidation

A Validation Claim SHALL NOT be mutated once recorded.

Correction, replacement, and invalidation of a Validation Claim are each represented by recording a new Validation Claim.

A new Validation Claim MAY explicitly:

* correct an earlier Claim (the earlier Claim's basis or outcome is understood to have been mistaken);
* replace an earlier Claim (the earlier Claim remains an accurate historical assertion for its own scope and time, but a new evaluation now applies going forward);
* invalidate an earlier Claim (the earlier Claim's basis, Evidence, or authority is no longer accepted).

Such a relationship SHALL identify the earlier Claim exactly, using the following minimal record-to-record reference:

```text
new Claim → affected earlier Claim
```

This reference:

* is owned by PEOS-006;
* is not an Artifact Relation;
* is not Artifact Supersession as defined by PEOS-002;
* has no separate entity identity of its own;
* does not erase, delete, or rewrite the earlier Claim.

Earlier Claims remain historically preserved and inspectable regardless of correction, replacement, or invalidation.

The normative terms `supersede`, `supersedes`, `superseded`, and `Supersession` SHALL NOT be used to describe Claim replacement, except when explicitly explaining that PEOS-002 Artifact Supersession does not apply to Claims.

The currently applicable Claim for a given Subject, scope, and criteria is derived by identifying the most recent Claim that has not been replaced or invalidated by a later Claim. This determination is a derived view; it is never stored as a field on the evaluated Subject.

---

# Satisfaction Claim

A Satisfaction Claim is a Validation Claim whose criteria identify one or more Requirements.

A Satisfaction Claim SHALL have exactly one engineering subject.

The subject is the Artifact, Artifact Revision, runtime subject, Decision Outcome, Engineering Commitment, or other explicitly permitted engineering subject whose satisfaction of the Requirement criteria is asserted.

Each Requirement or Requirement Artifact Revision SHALL appear as a Claim criterion.

A Requirement SHALL NOT become the Claim subject merely because it supplies the required intent being evaluated.

When the wording, acceptance criteria, or applicability being evaluated are revision-specific, the criterion SHALL identify the exact Requirement Artifact Revision rather than only the Requirement identity.

A Satisfaction Claim answers whether an identified engineering subject satisfies an identified Requirement criterion within an explicit scope. It does not answer whether a Requirement is satisfied in the abstract.

A Requirement SHALL NOT become both the Claim subject and the same Claim's criterion.

Example:

```text
subject:
Artifact Revision AR-42

criterion:
Requirement Artifact Revision RR-7

outcome:
satisfied
```

Non-conforming counterexample:

```text
subject:
Requirement Artifact Revision RR-7

criterion:
Requirement Artifact Revision RR-7
```

The counterexample is non-conforming because the same Requirement Artifact Revision is used as both the evaluated subject and its own criterion.

This does not prohibit every Claim whose subject is a Requirement. A general Validation Claim MAY evaluate a Requirement as an engineering subject for other purposes, such as statement quality, completeness, consistency, or conformance to a Requirement-writing profile. The prohibition applies specifically to using a Requirement as the Satisfaction Claim subject merely to represent satisfaction of that same Requirement.

A Requirement SHALL NOT own mutable satisfaction state. Requirement satisfaction is derived exclusively from applicable Satisfaction Claims, in accordance with PEOS-005 §30.2 and §35.

## Derived Satisfaction and Aggregation

Current Requirement satisfaction requires Product-owned aggregation rules where more than one applicable subject, Allocation, or Satisfaction Claim contributes to the derived view.

This specification owns the Claim mechanism. It SHALL NOT silently define a universal aggregation policy such as any subject satisfied, all subjects satisfied, or latest subject satisfied.

A Product contract SHALL define the aggregation rule where more than one applicable subject, allocation, or Satisfaction Claim contributes to a derived Requirement-satisfaction view.

This is not general traceability coverage. This specification does not define orphan detection or completeness.

---

# Conformance Claim

A Conformance Claim is a Validation Claim whose criteria identify one or more normative specifications, Product contracts, profiles, templates, or other explicitly identified conformance rules.

A Conformance Claim evaluates exactly one Subject against those criteria.

Conformance of a Subject is derived exclusively from applicable Conformance Claims. No Subject owns a mutable `conformant` field.

---

# Derived Validation State

The following are derived views, never stored fields:

* Requirement satisfaction;
* Artifact conformance;
* quality evaluation (as specialized by PEOS-007);
* runtime compliance (as specialized by PEOS-008);
* template conformance (as specialized by PEOS-009).

Each is derived from the applicable Validation Claims (and their specializations), the Evidence those Claims cite, the scope and criteria those Claims identify, any applicable replacement or invalidation references, the authority under which those Claims were established, and any governing Product rules.

No evaluated Subject owns mutable fields such as:

```text
satisfied
validated
conformant
quality
qualityScore
compliant
```

A Validation Claim records a historical assertion and outcome. It does not establish a mutable, globally current field on its Subject. Any consumer requiring "current" status MUST query the applicable, non-replaced, non-invalidated Claim(s) for that Subject, scope, and criteria at the time of the query.

---

# Validation and Lifecycle

A Validation Claim and a Validation Execution Record do not, by themselves, assign a Lifecycle State or a State Assignment.

A Lifecycle Transition, as defined by PEOS-003, MAY require Validation Evidence or an applicable Validation Claim as a Transition Guard condition or as Transition Evidence.

Lifecycle effects remain exclusively governed by PEOS-003. This specification does not define lifecycle consequences.

Validation outcome and Lifecycle State remain independently inspectable. A Subject's current Lifecycle State does not determine, and is not determined by, the outcome of its most recent applicable Validation Claim.

---

# Validation and Decisions

Certification, acceptance, approval, rejection, and authorization are governance outcomes governed by PEOS-004, expressed through a Decision Outcome.

A Decision Outcome MAY rely on one or more Validation Claims as part of its Decision Basis.

A Validation Claim does not replace a Decision Outcome where authority or governance judgment is required. A Validation Claim records what was observed and determined by applying a Method to Evidence; it does not itself authorize, approve, or accept anything on behalf of an organization.

---

# Validation and Requirements

This specification preserves all applicable PEOS-005 principles, including:

* a Requirement is an Artifact;
* a Requirement Statement belongs to an Artifact Revision;
* Allocation is not Validation;
* Applicability is not Lifecycle;
* a Requirement does not own mutable satisfaction state;
* an Engineering Commitment remains distinct from a Requirement.

A Satisfaction Claim or Conformance Claim concerning a Requirement SHALL identify the exact participant level (Requirement identity or Requirement Artifact Revision) in accordance with PEOS-005 §17.3.

---

# Validation and Quality

PEOS-007 specializes the Validation Claim mechanism defined by this specification to produce a Quality Claim.

This specification does not define Quality Characteristics, Quality Measures, Quality Profiles, or any quality-specific evaluation vocabulary. Those concerns belong exclusively to PEOS-007.

PEOS-007 does not define an independent Activity, Evidence, or Claim mechanism of its own.

---

# Validation and Runtime

PEOS-008 specializes the Validation Claim mechanism defined by this specification to produce a Compliance Claim.

A Runtime Observation, as defined by PEOS-008, MAY serve as Evidence for a Validation Claim, subject to the Evidence qualification rules of this specification.

This specification does not define Runtime Contracts, Runtime Binding, or Runtime Violations. Those concerns belong exclusively to PEOS-008.

---

# Validation and Templates

PEOS-009 MAY reference the Conformance Claim mechanism defined by this specification for template-specific conformance reporting.

This specification does not define Template structure, Template Application, or template-specific conformance rules. Those concerns belong exclusively to PEOS-009.

---

# Validation Invariants

A conformant implementation MUST preserve the following invariants.

## Validation Subject Level Invariant

Every Planned Validation Activity, Validation Execution Record, and Validation Claim identifies its Subject at an exact, explicit participant level.

## Planned Activity Ownership Invariant

Every Planned Validation Activity belongs to exactly one Validation Plan Revision and has no PEOS identity independent of that Revision.

## Plan-Local Key Invariant

Every Planned Validation Activity has a stable key unique within its owning Validation Plan Revision, and that key does not constitute an identity outside that Revision.

## Execution Record Immutability Invariant

A recorded Validation Execution Record does not change. Correction produces a new record.

## Evidence Artifact Revision Invariant

Every Evidence citation identifies an exact Artifact Revision, and every Evidence Artifact conforms to the role defined by PEOS-002.

## Single Claim Subject Invariant

Every Validation Claim identifies exactly one Subject.

## Claim Criterion Separation Invariant

A criterion is never treated as a second Subject.

## Claim History Preservation Invariant

Correction, replacement, or invalidation of a Validation Claim does not erase, delete, or rewrite an earlier Claim.

## Claim Is Not Artifact Invariant

A Validation Claim is never represented as an Artifact, an Artifact Revision, a Requirement, or a State Assignment.

## Derived Satisfaction Invariant

Satisfaction, conformance, and quality are always derived from applicable Claims; no evaluated Subject owns a mutable field representing them.

## Validation and Lifecycle Separation Invariant

Validation outcome does not itself constitute a Lifecycle State or State Assignment.

## Validation and Decision Separation Invariant

A Validation Claim does not itself constitute a Decision Outcome, Certification, or Acceptance.

## No Parallel Revision System Invariant

Validation Plan uses Artifact Revision; Planned Validation Activity, Validation Execution Record, and Validation Claim have no independent revision system of their own.

---

# Non-Conforming Patterns

The following implementation patterns violate this specification.

## Mutable Requirement Satisfaction

Representing Requirement satisfaction as a mutable field on the Requirement or its Artifact Revision.

## Mutable Artifact Conformance

Representing Artifact conformance as a mutable field on the Artifact or its Revision.

## Claim as Artifact

Representing a Validation Claim as an Artifact, granting it Artifact identity, Artifact Revisions, or Artifact Representations.

## Claim Revision

Creating a "Claim Revision" or otherwise revising a recorded Validation Claim in place instead of recording a new Claim.

## Claim Lifecycle

Assigning a Lifecycle State or State Assignment to a Validation Claim.

## Composite Claim Subject

Identifying more than one Subject on a single Validation Claim.

## Criterion Treated as Subject

Using a criterion (a Requirement, Quality Characteristic, threshold, or other evaluation rule) as though it were a second Claim Subject.

## Decision-Like Validation Activity

Modeling a Planned Validation Activity or its execution as an independent, non-Artifact, identity-and-lifecycle-bearing entity analogous to a Decision.

## Lifecycle-Governed Validation Activity

Assigning a Lifecycle State or State Assignment to a Planned Validation Activity or a Validation Execution Record.

## Planned Activity with Global Identity

Granting a Planned Validation Activity an identity that survives independently of its owning Validation Plan Revision.

## Missing Plan-Local Key

Representing a Planned Validation Activity without a stable key unique within its owning Validation Plan Revision.

## Validation Method Version

Introducing an independent revision system for Validation Method distinct from an optional backing Artifact's own Artifact Revision.

## Validation Plan Version

Introducing a revision system for Validation Plan distinct from ordinary Artifact Revision.

## Non-Artifact Normative Evidence

Treating material that has not been captured or referenced as an Artifact as normative PEOS Evidence.

## Evidence Without Exact Revision

Citing Evidence without identifying the exact Artifact Revision relied upon.

## Mutable Execution Record

Modifying a recorded Validation Execution Record in place instead of recording a new one.

## Execution Outcome as Lifecycle State

Representing a Validation Execution Record's outcome (completed, failed, interrupted, indeterminate) as a Lifecycle State or State Assignment.

## Silent Claim Mutation

Changing a Validation Claim's recorded basis, criteria, or outcome without recording a new Claim.

## Claim Replacement Called Artifact Supersession

Describing Claim correction, replacement, or invalidation using the normative term Supersession, or representing it as an Artifact Relation.

## Certification as Validation Method

Treating Certification as a Validation Method Type rather than a governance outcome under PEOS-004.

## Acceptance as Validation Method

Treating Acceptance as a Validation Method Type rather than a governance outcome under PEOS-004.

## Validation Outcome Used as Governance Authority

Treating the existence or outcome of a Validation Claim as itself sufficient authority to establish, change, or remove an Engineering Commitment, bypassing the Decision Model.

---

# Conformance

An implementation conforms to this specification when it can represent and preserve:

* Validation Plans as Artifacts using ordinary Artifact Revision;
* Planned Validation Activities as Revision-owned content with stable plan-local keys and no independent identity;
* Validation Execution Records as immutable, independently identifiable, non-Artifact records;
* Evidence exclusively as Artifacts satisfying the PEOS-002 Evidence role, cited at exact Revision level;
* Validation Claims as immutable, independently identifiable, non-Artifact records with exactly one Subject;
* Claim correction, replacement, and invalidation as record-to-record references distinct from Artifact Supersession;
* derived satisfaction, conformance, and quality state, with no mutable field on any evaluated Subject.

A Product contract conforms to this specification when it does not contradict the defined Validation semantics and does not require a mutable satisfaction, conformance, or quality field on any Subject.

A Runtime conforms to this specification when it does not mutate a recorded Validation Execution Record or Validation Claim, and does not present derived validation state as though it were stored authoritative state.

---

# Non-Goals

This specification does not require:

* every Validation Activity to be planned in advance;
* every Validation Execution to produce a Validation Claim;
* every Validation Claim to reference a Validation Execution Record;
* a specific validation tool, algorithm, or automation framework;
* a universal set of Validation Method Types;
* a single Validation Plan per Product or per Subject;
* Validation to determine business correctness rather than conformance;
* a general traceability, coverage, or orphan-detection mechanism.

---

# References

This document depends on:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model;
* PEOS-003 — Lifecycle;
* PEOS-004 — Decision Model;
* PEOS-005 — Requirement Model.

This document provides the Validation foundation for:

* PEOS-007 — Quality Model;
* PEOS-008 — Runtime Contract;
* PEOS-009 — Template Contract.
