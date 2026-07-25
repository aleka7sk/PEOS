# PEOS-009 — Template Contract

**Category:** Normative

**Status:** Draft

**Version:** 0.1

---

# Abstract

This specification defines the PEOS Template Contract Model.

The Template Contract Model describes how a Template generates Artifacts, how each generation is recorded, and how template conformance and compatibility are derived, without introducing a parallel revision system, a Template Instance entity, or a general relationship-collection entity.

It distinguishes:

* a Template from the Artifact Revisions it generates;
* a Template Application from the Template Application Record that persists it;
* a generated Artifact from the Template that produced it;
* the optional Generated-From relation from the Template Application Record it supplements but never replaces;
* Template Supersession, which reuses PEOS-002, from Template Lifecycle, which reuses PEOS-003.

A Template is an Artifact. A Template Application Record is an immutable record, not an Artifact. A generated Artifact is an ordinary Artifact with its own identity. A Template Conformance Claim is a specialization of the Conformance Claim defined by PEOS-006.

No Template, Template Revision, or generated Artifact owns mutable compatibility or conformance state.

---

# Purpose

The purpose of the Template Contract Model is to provide a stable and implementation-independent foundation for:

* representing a Template as an ordinary Artifact, using ordinary Artifact Revision, without a parallel Template Version mechanism;
* recording each application of a Template as an immutable, independently identifiable Template Application Record;
* generating ordinary Artifacts that retain their own identity, independent of any "Template Instance" concept;
* representing composition, specialization, and recursive use of Templates with explicit binarity and cycle policy;
* deriving template compatibility and conformance rather than storing them;
* reusing PEOS-002 Artifact Supersession and PEOS-003 Lifecycle rather than defining parallel mechanisms for Template Supersession and Template Lifecycle;
* relating Template Contract to Requirement, Validation, and Decision concerns without redefining any of them.

---

# Scope

This specification defines:

* Template and Template Artifact Revision;
* Template Representation, Parameter, Parameter Type, Default, and Constraint;
* Template Application Record and Template Application Outcome;
* Generated Artifact and the Generated-From relation;
* Template Composition, Specialization, and recursive use;
* Template Compatibility and Migration;
* Template Conformance Claim and derived Template Conformance;
* Template Lifecycle and Template Supersession, both by reuse of PEOS-003 and PEOS-002;
* the boundary between Template Contract and Requirement, Validation, and Decision concerns.

This specification does not define:

* Artifact, Artifact Revision, Artifact Relation, or Artifact Supersession structure (PEOS-002);
* Lifecycle Definitions, States, or Transitions (PEOS-003);
* Decision structure or governance outcomes (PEOS-004);
* Requirement structure (PEOS-005);
* the base Validation Claim or Conformance Claim mechanism (PEOS-006);
* Quality Characteristics, Measures, or Profiles (PEOS-007);
* Runtime Contract structure (PEOS-008);
* a general relationship-collection, group, or hyperedge model;
* a general traceability, coverage, or orphan-detection mechanism.

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
* PEOS-006 — Validation Model.

Terms defined by those specifications retain their normative meaning unless explicitly specialized here.

This specification does not redefine Artifact identity, Artifact Relation structure, Artifact Supersession, Lifecycle semantics, Decision structure, Requirement structure, or the Validation Claim mechanism.

---

# Template Model Overview

```text
Template
    is an Artifact
    uses ordinary Artifact Revision
    has no Template Version

Template Artifact Revision
    is immutable
    contains Parameters, Defaults, Constraints, and generation semantics

Template Application Record
    is an immutable record
    is independently identifiable
    identifies exact Template Artifact Revision and resolved parameter values
    identifies generated Artifact Revision where generation succeeds

Generated Artifact
    is an ordinary Artifact
    has its own identity
    is not a Template Instance

Generated-From
    is an optional binary Artifact Relation
    supplements, and does not replace, the Application Record

Template Conformance Claim
    specializes Conformance Claim
    is not an Artifact
```

The following relationships are normative:

```text
Template Artifact Revision
    is applied through
        zero or more Template Application Records

Template Application Record
    produces
        zero or one generated Artifact Revision, when successful

Generated Artifact Revision
    MAY reference, via Generated-From
        the Template Artifact Revision it was generated from

Template Artifact Revision
    MAY specialize or compose
        one or more other Template Artifact Revisions, subject to cycle prohibition

Template Conformance Claim
    concerns
        exactly one Subject drawn from generated content or Template content
```

A Template Application Record is the authoritative record of a template application. The Generated-From relation, where present, is supplementary traceability; it is never a substitute for the Application Record.

---

# Template

Template SHALL be an Artifact, as defined by PEOS-002.

Template SHALL use ordinary Artifact Revision.

This specification does not introduce `Template Version` or `Template Revision` as a revision system parallel to Artifact Revision.

The phrase "Template Revision" MAY be used only as informal shorthand for:

```text
Artifact Revision whose Artifact is a Template
```

It does not define a separate engineering entity, and it does not create a second revision mechanism.

---

# Template Artifact Revision

Every Template Artifact Revision SHALL identify:

* the represented template content;
* the generated Artifact Type or permitted Artifact Types;
* its parameters;
* parameter types;
* required parameters;
* defaults;
* constraints;
* expansion or generation semantics;
* composition or specialization references, where applicable;
* a compatibility declaration;
* applicability;
* provenance;
* authority, where required.

A Template Artifact Revision is immutable.

---

# Template Representation

Template syntax, schema, rendering language, source code, natural-language text, model, or other representation belongs to the Template Artifact Revision, in accordance with PEOS-002's Artifact Representation contract.

Representation is not Template identity.

A change of Representation format alone does not create a new template revision system; where it changes normative content, it produces an ordinary new Artifact Revision, per PEOS-002.

---

# Template Parameter

A Template Parameter is a Template Artifact Revision-owned value structure.

Every Template Parameter SHALL have a stable template-local key within its owning Template Artifact Revision.

A template-local key:

* is unique only within that exact Template Artifact Revision;
* is not an Artifact Identity;
* is not a global Template Parameter identity;
* MAY be referenced by constraints, defaults, composition mappings, and Template Application Records.

A Template Parameter has no independent parameter lifecycle or revision system of its own.

---

# Parameter Type

A Parameter Type is:

* a controlled vocabulary/type; or
* an exact reference to an externally governed normative type definition.

A Parameter Type has no mandatory independent PEOS Artifact identity.

Any change to a parameter's meaning or type requires a new Template Artifact Revision.

---

# Parameter Default

A Parameter Default is a Template Artifact Revision-owned value structure.

A default does not satisfy a required parameter where the owning Template Artifact Revision explicitly forbids default resolution for that parameter.

---

# Parameter Constraint

A Parameter Constraint is a Template Artifact Revision-owned value structure.

Every Parameter Constraint SHALL identify:

* the affected parameter or generated content;
* the rule;
* its scope;
* its evaluation point;
* its failure semantics;
* authority, where required.

---

# Template Application Record

A Template Application Record is an immutable record.

A Template Application Record is independently identifiable.

A Template Application Record is not an Artifact. It is not revisioned. It is not lifecycle-bearing.

Every Template Application Record SHALL identify:

* Application Record identity;
* the exact Template Artifact Revision applied;
* the resolved parameter values;
* the source of each resolved value (explicit input, default, or derived);
* the actor or executing system;
* authority, where required;
* timestamp;
* environment or context;
* provenance;
* execution outcome;
* the generated Artifact and exact generated Artifact Revision, where generation succeeded;
* known limitations;
* a correction, replacement, or invalidation reference, where applicable.

Correction of a Template Application Record creates a new Template Application Record.

A new record MAY explicitly `correct`, `replace`, or `invalidate` an earlier Template Application Record. Such a reference SHALL identify the earlier record exactly. The earlier record remains historically preserved.

Record replacement SHALL NOT be described using the normative term Supersession, except when explicitly explaining that PEOS-002 Artifact Supersession does not apply to Template Application Records.

---

# Template Application Outcome

Every Template Application Record SHALL identify an outcome drawn from an extensible controlled vocabulary including, at minimum:

* succeeded;
* failed;
* partially succeeded;
* interrupted;
* indeterminate.

Outcome is an attribute of the Template Application Record. There is no separate Outcome entity.

A `partially succeeded` outcome SHALL explicitly identify which outputs were generated and which were not.

---

# Generated Artifact

A generated Artifact is an ordinary PEOS Artifact, as defined by PEOS-002.

A generated Artifact Revision SHALL preserve:

* its own Artifact identity, independent of the Template's identity;
* its own immutable Artifact Revision;
* its Origin, per PEOS-002;
* its provenance;
* the exact Template Application Record that produced it;
* the exact Template Artifact Revision applied;
* its generated content.

Generated content does not share Template identity. A generated Artifact and the Template that produced it are two distinct Artifacts.

Changing generated content, after generation, follows ordinary Artifact Revision rules defined by PEOS-002; it is not a re-application of the Template unless a new Template Application Record is also created.

This specification does not create a "Template Instance" construct. A generated Artifact is simply an ordinary Artifact whose Origin and provenance cite the Template Application Record that produced it.

---

# Generated-From Relation

Generated-From is an optional Artifact Relation Type, specializing the general Artifact Relation contract defined by PEOS-002.

```text
source participant:
    the generated Artifact Revision

target participant:
    the Template Artifact Revision

direction:
    generated → template

multiplicity:
    many generated Artifact Revisions MAY reference one Template Artifact Revision

cycles:
    prohibited

identity:
    none (per PEOS-002's general Artifact Relation contract)

revision:
    none

lifecycle:
    none
```

Every Generated-From relation SHALL identify:

* its exact source;
* its exact target;
* its Relation Type;
* its scope;
* its provenance.

A Generated-From relation SHALL NOT contain:

* the full resolved parameter state;
* execution event history;
* mutable application status;
* authority history.

That information belongs exclusively to the Template Application Record.

Generated-From is supplementary traceability. It does not replace the Template Application Record, and a conformant implementation MAY omit Generated-From entirely provided the Application Record itself remains inspectable.

This specification does not define coverage, orphan detection, or a general cross-artifact traceability mechanism; those remain the responsibility of a future Traceability Model.

---

# Template Composition

Template Composition SHALL remain binary at the Artifact Relation level, in accordance with PEOS-002.

One logical multi-template composition MAY be represented through multiple binary relations.

A composition reference SHALL identify the exact Template Artifact Revision, where exact content matters.

The source participant is the composing Template Artifact Revision. The target participant is the composed Template Artifact Revision. This direction SHALL be used consistently throughout this specification.

One Template Artifact Revision MAY compose multiple other Template Artifact Revisions. One Template Artifact Revision MAY be composed by multiple other Template Artifact Revisions. Many-to-many composition semantics SHALL be represented through multiple binary Artifact Relations, never a single relation with more than one source or target.

Every Template Composition relation SHALL identify:

* its exact source (the composing Template Artifact Revision);
* its exact target (the composed Template Artifact Revision);
* participant levels;
* direction;
* multiplicity;
* cycle policy;
* scope;
* provenance;
* parameter mapping rules;
* conflict handling.

Composition cycles SHALL NOT be permitted.

Controlled runtime recursive expansion, as defined in **Recursive Template Use**, is a separate mechanism and does not relax this prohibition.

This specification does not introduce:

* a Template Collection;
* a Template Composition Set;
* a hyperedge;
* any relationship-group identity.

A composed set of Templates MAY be interpreted together for an explicitly defined engineering purpose, in accordance with PEOS-002's general rule for collections of Artifact Relations, without that collection becoming a separate PEOS entity.

---

# Template Specialization

A specialized Template SHALL preserve an explicit relation to the Template it specializes.

Every Template Specialization relation SHALL identify:

* the source Template or Template Artifact Revision;
* the target base Template or Template Artifact Revision;
* participant levels;
* inherited elements;
* overridden elements;
* compatibility effect;
* scope;
* provenance.

Specialization cycles SHALL NOT be permitted.

A specialization does not mutate the base Template Artifact Revision.

---

# Recursive Template Use

The following are distinguished:

* a structural composition cycle;
* a specialization cycle;
* runtime recursive expansion (a Template Application that itself applies a Template during execution).

Structural composition cycles and specialization cycles SHALL be prohibited.

Runtime recursive expansion MAY be permitted only where:

* it is explicitly declared by the Template Artifact Revision;
* it is bounded;
* termination conditions are defined;
* a maximum depth or equivalent control exists;
* the resulting provenance remains inspectable through Template Application Records.

Permission for controlled runtime recursive expansion SHALL NOT be interpreted as permission for structural composition or specialization cycles. The two are independent policies.

---

# Template Compatibility

Compatibility is not mutable global state.

A compatibility declaration SHALL be scoped to:

* exact Template Artifact Revisions;
* the consumer or generated Artifact Type;
* applicable constraints;
* the parameter contract;
* migration requirements, where applicable;
* the applicable Product contract.

Current compatibility is a derived interpretation, computed from the applicable compatibility declarations at query time.

The following mutable fields SHALL NOT be created:

```text
Template.compatible
TemplateRevision.compatible
```

---

# Template Migration

Migration is a governed transformation from one Template Artifact Revision to another.

Every migration SHALL identify:

* the source Template Artifact Revision;
* the target Template Artifact Revision;
* affected generated Artifacts or future applications;
* parameter mappings;
* transformation rules;
* information-loss risks;
* authority;
* provenance;
* applicable Validation requirements, per PEOS-006.

Migration does not rewrite historical Template Application Records or previously generated Artifact Revisions.

A new Template Application Record is required when a generated Artifact is regenerated from a new Template Artifact Revision. Regeneration is a new application, not a retroactive correction of the original one, unless it is explicitly recorded as a correction per **Template Application Record**.

---

# Template Conformance Claim

A Template Conformance Claim is a specialization of the Conformance Claim defined by PEOS-006.

A Template Conformance Claim inherits, without redefinition, all Validation Claim rules defined by PEOS-006.

The Subject of a Template Conformance Claim SHALL be exactly one of:

* a generated Artifact;
* a generated Artifact Revision;
* a Template Artifact;
* a Template Artifact Revision;
* another explicitly permitted engineering subject.

Criteria MAY include:

* a Template Artifact Revision;
* parameter constraints;
* a generated Artifact Type rule;
* a representation rule;
* a compatibility rule;
* an applicable Product contract;
* a Requirement Artifact Revision.

A Template Artifact Revision used as a criterion is not a second Subject.

---

# Derived Template Conformance

Template conformance is a derived view. It is not a stored field.

Template conformance SHALL be derived from:

* applicable Template Conformance Claims;
* the exact Template Artifact Revision;
* the applicable Template Application Record;
* the generated Artifact Revision;
* criteria;
* Evidence;
* Claim correction, replacement, and invalidation history;
* scope;
* authority;
* governing Product rules.

The following mutable fields SHALL NOT be created:

```text
Template.conformant
TemplateRevision.conformant
GeneratedArtifact.conformant
```

---

# Template Lifecycle

Template is an ordinary Lifecycle Subject, as defined by PEOS-003.

A State Assignment, as defined by PEOS-003, applied to a Template does not:

* create a Template Artifact Revision;
* establish Template Supersession;
* establish compatibility;
* mutate a Template Application Record;
* establish conformance.

---

# Template Supersession

Template Supersession reuses Artifact Supersession, as defined by PEOS-002. This specification does not define a separate Template Supersession entity.

Every Template Supersession relation SHALL be:

* binary;
* explicit;
* directed from the superseding subject to the superseded subject;
* cycle-prohibited;
* scoped;
* provenance-bearing;

in accordance with PEOS-002's general Artifact Supersession contract.

A newer Template Artifact Revision does not automatically imply Supersession.

Supersession does not automatically migrate generated Artifacts; migration, where required, follows the rules in **Template Migration**.

Supersession does not rewrite historical Template Application Records.

---

# Template and Requirements

A Template MAY generate Requirement Artifacts or Requirement Artifact Revisions.

Generated Requirements remain ordinary Requirements, as defined by PEOS-005, subject to all of PEOS-005's rules without exception.

Template parameter values do not become Requirement identity.

Changing generated Requirement wording, after generation, requires an ordinary Artifact Revision, per PEOS-002 and PEOS-005.

Template Application does not, by itself, establish Requirement satisfaction, Validation, Allocation, Applicability, Authority, or Lifecycle State for a generated Requirement.

Requirement satisfaction and Validation are governed by PEOS-006.

Requirement Allocation, Applicability, and Requirement Authority are governed by PEOS-005.

Lifecycle State and State Assignment are governed by PEOS-003.

---

# Template and Validation

Template Conformance Claim is the sole Claim type this specification defines, and it is a specialization of the Conformance Claim mechanism owned by PEOS-006. This specification does not define an independent Activity, Evidence, or Claim base mechanism.

A Template Application Record MAY be referenced as Evidence for a Validation Claim or Template Conformance Claim, subject to PEOS-006's Evidence qualification rules.

---

# Template and Decisions

Template generation or migration that requires governance approval SHALL use Decision semantics, as defined by PEOS-004.

A Template Application Record's outcome is not a Decision Outcome.

A Template Conformance Claim is not an approval or authorization. It records a determination; it does not itself grant governance permission.

---

# Template Invariants

A conformant implementation MUST preserve the following invariants.

## Template Artifact Invariant

Every Template is an Artifact using ordinary Artifact Revision.

## No Template Version Invariant

Template evolution uses Artifact Revision; there is no separate Template Version or Template Revision mechanism.

## Template Representation Ownership Invariant

Template representation belongs to the Template Artifact Revision, per PEOS-002; it is not a separate identity.

## Template Parameter Local-Key Invariant

Every Template Parameter has a stable key unique within its owning Template Artifact Revision, and that key does not constitute an identity outside that Revision.

## Template Application Record Immutability Invariant

A recorded Template Application Record does not change. Correction produces a new record.

## Generated Artifact Identity Invariant

A generated Artifact has its own Artifact identity, independent of the Template's identity.

## Generated-From Binarity Invariant

Every Generated-From relation has exactly one source and exactly one target, per PEOS-002.

## Generated-From Non-Identity Invariant

A Generated-From relation has no normative identity, revision, or lifecycle of its own.

## Application Record Authority Invariant

Every Template Application Record identifies its actor and, where required, its authority.

## Composition Binarity Invariant

Every Template Composition relation has exactly one source and exactly one target.

## Composition Cycle Invariant

Template Composition cycles are prohibited.

## Specialization Cycle Invariant

Template Specialization cycles are prohibited.

## Controlled Recursion Invariant

Runtime recursive Template expansion, where permitted, is bounded, terminating, and inspectable.

## Compatibility Derived-State Invariant

Template compatibility is always derived from applicable compatibility declarations; it is never a stored field.

## Migration History Preservation Invariant

Template Migration does not rewrite historical Template Application Records or previously generated Artifact Revisions.

## Template Conformance Claim Specialization Invariant

Every Template Conformance Claim is a Conformance Claim as defined by PEOS-006, without redefinition of Claim mechanics.

## Single Template Conformance Subject Invariant

Every Template Conformance Claim identifies exactly one Subject.

## Derived Template Conformance Invariant

Template conformance is always derived from applicable Template Conformance Claims; no Subject owns a mutable field representing it.

## Template Lifecycle Separation Invariant

A Template's State Assignment does not itself establish Template Supersession, compatibility, or conformance.

## Template Supersession Separation Invariant

Template Supersession reuses PEOS-002's general Artifact Supersession contract; it is not a separate entity, and a newer Template Artifact Revision does not automatically imply it.

---

# Non-Conforming Patterns

The following implementation patterns violate this specification.

## Template Version

Introducing a revision system for Template distinct from ordinary Artifact Revision.

## Parallel Template Revision System

Introducing any revision mechanism for Template, Template Parameter, or Template Application distinct from Artifact Revision and immutable-record correction.

## Template Instance

Introducing a "Template Instance" entity distinct from an ordinary generated Artifact.

## Template Parameter with Global Identity

Granting a Template Parameter an identity that survives independently of its owning Template Artifact Revision.

## Mutable Template Application Record

Modifying a recorded Template Application Record in place instead of recording a new one.

## Application Record Revision

Creating an "Application Record Revision" instead of a new Template Application Record with an explicit correction reference.

## Application Record Lifecycle

Assigning a Lifecycle State or State Assignment to a Template Application Record.

## Generated Artifact Sharing Template Identity

Representing a generated Artifact as sharing, or inheriting, the Template's own Artifact identity.

## Generated-From Relation as Application Record

Treating the Generated-From relation as a substitute for the Template Application Record.

## Parameter Values Stored Only on Generated-From Relation

Storing resolved parameter values solely on the Generated-From relation instead of on the Template Application Record.

## Template Composition Hyperedge

Representing Template Composition as a single relation with more than one source or more than one target.

## Template Collection Owning Normative State

Granting a collection of composed or specialized Templates independent identity or normative state of its own.

## Composition Cycle

Representing a Template Composition relation that is directly or transitively cyclic.

## Specialization Cycle

Representing a Template Specialization relation that is directly or transitively cyclic.

## Unbounded Recursive Expansion

Permitting runtime recursive Template expansion without declared bounds, termination conditions, or inspectable provenance.

## Mutable Template Compatibility

Storing template compatibility as a mutable field such as `Template.compatible` instead of deriving it from applicable declarations.

## Migration Rewriting History

Using Template Migration to rewrite or delete historical Template Application Records or previously generated Artifact Revisions.

## Parallel Template Claim Base

Defining a Template Conformance Claim that does not specialize PEOS-006's Conformance Claim, or that redefines Claim identity, immutability, or replacement semantics.

## Template Conformance Claim as Artifact

Representing a Template Conformance Claim as an Artifact.

## Composite Template Conformance Subject

Identifying more than one Subject on a single Template Conformance Claim.

## Mutable Template Conformance

Storing template conformance as a mutable field such as `Template.conformant` instead of deriving it from applicable Claims.

## Template Lifecycle State Establishing Supersession

Treating a Template's State Assignment as itself establishing Template Supersession.

## Newer Template Revision Implying Supersession

Assuming that creation of a newer Template Artifact Revision automatically supersedes an earlier one without an explicit Supersession relation.

## Application Outcome Used as Governance Authority

Treating a Template Application Record's outcome or a Template Conformance Claim's outcome as itself sufficient authority to establish, change, or remove an Engineering Commitment, bypassing the Decision Model.

---

# Conformance

An implementation conforms to this specification when it can represent and preserve:

* Templates as Artifacts using ordinary Artifact Revision;
* Template Parameters, Defaults, and Constraints as Template Artifact Revision-owned content with stable local keys;
* Template Application Records as immutable, independently identifiable, non-Artifact records;
* generated Artifacts as ordinary Artifacts with their own identity;
* an optional, non-identity-bearing Generated-From relation that never substitutes for the Application Record;
* Template Composition and Specialization as binary, cycle-prohibited Artifact Relations;
* derived template compatibility and conformance, with no mutable field on any Subject;
* Template Supersession by reuse of PEOS-002, and Template Lifecycle by reuse of PEOS-003, without redefinition.

A Product contract conforms to this specification when it does not contradict the defined Template semantics and does not require a mutable compatibility or conformance field on any Subject.

---

# Non-Goals

This specification does not require:

* every Artifact to be generated from a Template;
* a specific templating language or engine;
* every Template Application to produce exactly one generated Artifact;
* automatic migration of previously generated Artifacts upon Template Supersession;
* a general relationship-collection, group, or hyperedge model;
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
* PEOS-006 — Validation Model.

This document is provided a Runtime-generation foundation by:

* PEOS-008 — Runtime Contract, where a Runtime Contract MAY itself be a Template-generated Artifact.
