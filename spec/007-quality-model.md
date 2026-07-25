# PEOS-007 — Quality Model

**Category:** Normative

**Status:** Draft

**Version:** 0.1

---

# Abstract

This specification defines the PEOS Quality Model.

The Quality Model describes how engineering quality expectations are configured, measured, and claimed, without introducing a Validation mechanism independent of PEOS-006.

It distinguishes:

* a Quality Profile from the Quality Characteristics, Measures, Thresholds, Targets, and Constraints it configures;
* a Measurement Record from the Quality Claim that interprets it;
* a Quality Claim from the Validation Claim it specializes;
* a quality score from the derived view that produces it.

A Quality Claim is a Validation Claim. It is not an Artifact, and it introduces no revision, lifecycle, or base mechanism distinct from PEOS-006.

No Artifact, Artifact Revision, Requirement, or Quality Profile owns mutable quality state.

---

# Purpose

The purpose of the Quality Model is to provide a stable and implementation-independent foundation for:

* configuring which Quality Characteristics, Measures, Thresholds, Targets, and Constraints apply to which Subjects;
* recording quality measurements as immutable, non-Artifact records specializing PEOS-006's Validation Execution Record pattern;
* recording quality determinations as Quality Claims that specialize PEOS-006's Validation Claim, without redefining Claim mechanics;
* deriving quality state without mutable fields on the evaluated Subject;
* relating Quality to Requirements, Validation, and Runtime concerns without redefining any of them.

---

# Scope

This specification defines:

* Quality Profile;
* Quality Characteristic;
* Quality Measure;
* Threshold, Target, and Quality Constraint;
* Normalization Rule and Aggregation Rule;
* Measurement Record;
* Quality Claim, including its Subject and criteria;
* the boundary between this specification and PEOS-006.

This specification does not define:

* the Validation Claim, Validation Plan, Planned Validation Activity, Validation Execution Record, or Evidence mechanisms (PEOS-006);
* Artifact structure (PEOS-002);
* Requirement structure (PEOS-005);
* Runtime enforcement (PEOS-008);
* Template structure (PEOS-009);
* a mandatory quality methodology, metric catalog, or scoring algorithm.

---

# Normative Foundation

The key words MUST, MUST NOT, SHALL, SHALL NOT, SHOULD, SHOULD NOT, MAY, REQUIRED, RECOMMENDED, and OPTIONAL in this specification are to be interpreted as defined by PEOS-000.

This specification builds upon:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model;
* PEOS-005 — Requirement Model;
* PEOS-006 — Validation Model.

Terms defined by those specifications retain their normative meaning unless explicitly specialized here.

This specification does not redefine Artifact identity, Requirement structure, or the Validation Claim, Evidence, or Execution Record mechanisms defined by PEOS-006.

---

# Quality Model Overview

```text
Quality Profile
    is an Artifact
    uses ordinary Artifact Revision
    contains, within its Revision:
        Quality Characteristics
        Quality Measures
        Thresholds
        Targets
        Quality Constraints
        Normalization Rules
        Aggregation Rules

Measurement Record
    specializes Validation Execution Record
    is an immutable record
    is not an Artifact
    identifies the measured Subject, Characteristic, Measure, and observed value

Quality Claim
    specializes Validation Claim
    is an immutable record
    is not an Artifact
    has exactly one Subject
    has criteria drawn from an applicable Quality Profile Revision
```

A Quality Evaluation is not an independent entity. It is the combination of an applicable Quality Profile Revision, a Planned Validation Activity, a Measurement Record, Evidence, and a Quality Claim, exactly as those constructs are defined by PEOS-006 and specialized here.

---

# Quality Profile

A Quality Profile is an Artifact as defined by PEOS-002.

A Quality Profile uses ordinary Artifact Revision for all of its evolution. There is no Quality Profile Version distinct from Artifact Revision.

Every Quality Profile Revision SHALL identify:

* its scope;
* the Subjects or Subject types to which it applies;
* the Quality Characteristics it defines or references;
* the Quality Measures it defines or references;
* the Thresholds it defines;
* the Targets it defines;
* the Quality Constraints it defines;
* the Normalization Rules it defines, where applicable;
* the Aggregation Rules it defines, where applicable;
* its applicability conditions;
* provenance;
* authority, where required.

Modification of a Quality Profile's content constitutes a content change and SHALL create a new Artifact Revision in accordance with PEOS-002.

---

# Quality Characteristic

A Quality Characteristic is a controlled term whose identity and meaning are scoped by:

* the exact owning Quality Profile Revision in which it is defined; or
* an exact externally referenced normative vocabulary.

A Quality Characteristic has no independent revision system of its own. It does not claim globally stable independent identity unless that identity is supplied by an exact referenced external vocabulary.

A change in the meaning of a Quality Characteristic requires either a new Quality Profile Revision or a new identity within the externally governed vocabulary it references. A Quality Characteristic's meaning SHALL NOT change silently within the same Profile Revision or the same externally referenced identity.

---

# Quality Measure

A Quality Measure defines how a Quality Characteristic is observed or computed.

A Quality Measure SHALL identify:

* the Characteristic it measures;
* its unit;
* its scale;
* the method used to obtain a value;
* the Evidence required as input;
* how uncertainty is handled;
* its valid range;
* an applicable Normalization Rule, where the measured value requires normalization.

A Quality Measure is a controlled vocabulary/type or a Quality Profile Revision-owned definition. It has no mandatory independent Artifact identity and no Quality Measure Version.

---

# Threshold

A Threshold is a Quality Profile Revision-owned value structure defining a boundary used for classification or for determining a Quality Claim outcome.

A Threshold has no independent identity outside its owning Profile Revision unless supplied by an exact referenced external vocabulary.

---

# Target

A Target is a Quality Profile Revision-owned value structure defining a desired value or range.

A Target does not itself define pass/fail classification; a Target expresses intent, while a Threshold expresses the boundary used for a Claim outcome. The two SHALL NOT be conflated.

A Target has no independent identity outside its owning Profile Revision unless supplied by an exact referenced external vocabulary.

---

# Quality Constraint

A Quality Constraint is a normative restriction contained in a Quality Profile Revision.

A Quality Constraint MAY also be represented as a Requirement, as defined by PEOS-005, where persistent Requirement identity, Lifecycle, Authority, Applicability, Allocation, or Requirement relationships are needed for that constraint.

Every Quality Constraint SHALL NOT be silently treated as a Requirement merely because it constrains engineering behavior. Representing a Quality Constraint as a Requirement is a deliberate, explicit modeling choice, made per constraint, not an automatic consequence of this specification.

---

# Normalization Rule

A Normalization Rule is a Quality Profile Revision-owned value structure describing how a raw Measurement Record value is transformed into a comparable form.

A Normalization Rule has no independent identity, revision, or lifecycle outside its owning Profile Revision.

---

# Aggregation Rule

An Aggregation Rule is a Quality Profile Revision-owned value structure describing how multiple Measurement Records or Quality Claims are combined into a single derived view.

An Aggregation Rule has no independent identity, revision, or lifecycle outside its owning Profile Revision.

An Aggregation Rule produces a derived view. It does not itself produce a stored, mutable field on any Subject.

---

# Measurement Record

A Measurement Record specializes the Validation Execution Record defined by PEOS-006.

A Measurement Record is an immutable record. It is not an Artifact. It has no revisions and no lifecycle.

Every Measurement Record SHALL identify:

* the measured Subject and its exact participant level;
* the exact Quality Characteristic and Quality Measure references applied;
* the observed value;
* the unit and scale;
* the method used;
* the timestamp of measurement;
* the environment or context, where applicable;
* known uncertainty;
* known limitations;
* the Evidence Artifact Revisions relied upon;
* provenance.

Correction of a Measurement Record creates a new Measurement Record, in accordance with PEOS-006's Validation Execution Record correction rules. A Measurement Record SHALL NOT be mutated once recorded.

---

# Quality Claim

A Quality Claim is a specialization of Validation Claim, as defined by PEOS-006.

A Quality Claim inherits, without redefinition:

* identity;
* immutability;
* the absence of revisions;
* the absence of lifecycle;
* the requirement of exactly one Subject;
* the separation of criteria from Subject;
* Evidence citation rules;
* the replacement, correction, and invalidation rules defined by PEOS-006 (`new Claim → affected earlier Claim`, never Artifact Supersession);
* historical preservation;
* derived-current-effect semantics.

This specification does not define a second Claim base model, a second replacement mechanism, or a second Evidence mechanism. A Quality Claim exists exclusively as an instance of the PEOS-006 Validation Claim mechanism.

---

# Quality Claim Subject and Criteria

For a Quality Claim:

```text
subject = the evaluated Artifact or Artifact Revision

criteria may include:
    Quality Characteristic
    Quality Measure
    Threshold
    Target
    Quality Constraint
```

A Quality Claim SHALL NOT identify a composite subject such as a Characteristic-and-Revision pair. The Subject remains exactly one Artifact or Artifact Revision, as required by PEOS-006. The Characteristic, Measure, Threshold, Target, or Constraint being evaluated is a criterion, never a second Subject.

A Quality Claim's Subject MAY instead be a Requirement or Requirement Artifact Revision, when the evaluated quality expectation was expressed as a Requirement rather than as a Quality Profile entry, in accordance with the participant-level rules of PEOS-006.

---

# Quality Evaluation

Quality Evaluation is not an independent entity. It has no independent identity, no revision system, no lifecycle, and no mutable state of its own.

Quality Evaluation is the combination of:

* an applicable Quality Profile Revision;
* a Planned Validation Activity, as defined by PEOS-006;
* a Measurement Record;
* Evidence, as defined by PEOS-006 and PEOS-002;
* a Quality Claim.

An implementation SHALL NOT introduce an independent Quality Evaluation entity, an independent quality-specific Activity mechanism, or an independent quality-specific Evidence mechanism.

---

# Quality Score

A quality score MAY appear as:

* a value recorded on a Measurement Record;
* an outcome attribute of a Quality Claim;
* a derived view computed by applying an Aggregation Rule to applicable Measurement Records or Quality Claims.

A quality score SHALL NOT be stored as globally current mutable state on:

* an Artifact;
* an Artifact Revision;
* a Requirement;
* a Quality Profile.

Any consumer requiring a "current" quality score MUST compute it, on demand, from the applicable, non-replaced, non-invalidated Measurement Records and Quality Claims.

---

# Quality and Requirements

A Quality Constraint MAY be represented as a Requirement per this specification's own explicit modeling choice; PEOS-005's Requirement principles apply unchanged when it is.

PEOS-005 §16 (Requirement Quality) is not redefined by this specification. This specification supplies the measurement and claim mechanism that PEOS-005 §16 anticipates but does not itself define.

---

# Quality and Validation

This specification does not define an independent Activity, Evidence, or Claim base mechanism. Every quality-specific construct in this specification is a specialization of a construct owned by PEOS-006:

* Measurement Record specializes Validation Execution Record;
* Quality Claim specializes Validation Claim.

The boundary between PEOS-006 and PEOS-007 is exact: PEOS-006 owns the mechanism (Plan, Planned Activity, Method, Execution Record, Evidence, Claim, replacement semantics, derivation-not-storage). PEOS-007 owns the vocabulary and configuration (which Characteristics, Measures, Thresholds, and Targets apply to which Subjects) and the resulting specialized Claim type.

---

# Quality and Lifecycle

A Quality Claim does not itself assign a Lifecycle State or State Assignment. Lifecycle effects remain exclusively governed by PEOS-003.

A Lifecycle Transition MAY require a Quality Claim as Transition Evidence, in accordance with PEOS-003 and PEOS-006.

---

# Quality and Runtime

A Runtime Observation, as defined by PEOS-008, MAY serve as Evidence for a Measurement Record or a Quality Claim, subject to the Evidence qualification rules of PEOS-006.

This specification does not define Runtime Contracts or runtime enforcement. Those concerns belong exclusively to PEOS-008.

---

# Quality Invariants

A conformant implementation MUST preserve the following invariants.

## Quality Profile Artifact Invariant

Every Quality Profile is an Artifact using ordinary Artifact Revision.

## No Quality Profile Version Invariant

Quality Profile evolution uses Artifact Revision; there is no separate Quality Profile Version mechanism.

## Characteristic Scope Invariant

Every Quality Characteristic's identity and meaning are scoped by its owning Quality Profile Revision or an exact referenced external vocabulary.

## Measure Scope Invariant

Every Quality Measure's identity and meaning are scoped the same way as Quality Characteristic.

## Profile-Owned Rule Invariant

Threshold, Target, Quality Constraint, Normalization Rule, and Aggregation Rule are Quality Profile Revision-owned value structures without independent identity, revision, or lifecycle.

## Measurement Record Immutability Invariant

A recorded Measurement Record does not change. Correction produces a new record, per the Validation Execution Record correction rules of PEOS-006.

## Quality Claim Specialization Invariant

Every Quality Claim is a Validation Claim as defined by PEOS-006, without redefinition of Claim mechanics.

## Single Quality Claim Subject Invariant

Every Quality Claim identifies exactly one Subject.

## Quality Criterion Separation Invariant

A Quality Characteristic, Measure, Threshold, Target, or Constraint is always a criterion, never a second Subject.

## Derived Quality State Invariant

Quality score and quality evaluation outcome are always derived from applicable Measurement Records and Quality Claims; no Subject owns a mutable field representing them.

## Quality and Requirement Separation Invariant

A Quality Constraint is represented as a Requirement only by explicit modeling choice, never automatically.

## Quality and Validation Ownership Invariant

Measurement Record and Quality Claim specialize, and do not redefine, the Validation Execution Record and Validation Claim mechanisms owned by PEOS-006.

## No Parallel Claim Mechanism Invariant

This specification does not introduce a Claim base model, replacement mechanism, or Evidence mechanism independent of PEOS-006.

---

# Non-Conforming Patterns

The following implementation patterns violate this specification.

## Mutable Artifact Quality

Representing quality as a mutable field on an Artifact.

## Mutable Requirement Quality

Representing quality as a mutable field on a Requirement.

## Mutable Quality Score

Storing a quality score as globally current mutable state rather than deriving it on demand.

## Quality Profile Version

Introducing a revision system for Quality Profile distinct from ordinary Artifact Revision.

## Quality Characteristic Revision

Granting a Quality Characteristic an independent revision history distinct from its owning Quality Profile Revision or its referenced external vocabulary.

## Quality Measure Version

Introducing an independent revision system for Quality Measure.

## Independent Quality Evaluation Entity

Granting Quality Evaluation independent identity, revision, or lifecycle distinct from the combination of Profile, Planned Activity, Measurement Record, Evidence, and Quality Claim.

## Parallel Quality Activity

Defining a quality-specific Activity mechanism independent of PEOS-006's Planned Validation Activity.

## Parallel Quality Evidence

Defining a quality-specific Evidence mechanism independent of PEOS-002's Evidence role and PEOS-006's qualification rules.

## Parallel Quality Claim Base

Defining a Quality Claim that does not specialize Validation Claim, or that redefines Claim identity, immutability, or replacement semantics.

## Quality Claim as Artifact

Representing a Quality Claim as an Artifact.

## Composite Quality Claim Subject

Identifying more than one Subject, or a Characteristic-and-Revision pair, as the Subject of a single Quality Claim.

## Characteristic Treated as Second Subject

Using a Quality Characteristic, Measure, Threshold, or Target as though it were a second Claim Subject rather than a criterion.

## Threshold with Hidden External Meaning

Referencing a Threshold whose meaning is not scoped by an identifiable Quality Profile Revision or an exact referenced external vocabulary.

## Measurement Record Mutation

Modifying a recorded Measurement Record in place instead of recording a new one.

## Quality Outcome Used as Lifecycle State

Representing a Quality Claim's outcome as a Lifecycle State or State Assignment.

## Quality Outcome Used as Governance Authority

Treating a Quality Claim's existence or outcome as itself sufficient authority to establish, change, or remove an Engineering Commitment, bypassing the Decision Model.

---

# Conformance

An implementation conforms to this specification when it can represent and preserve:

* Quality Profiles as Artifacts using ordinary Artifact Revision;
* Quality Characteristics, Measures, Thresholds, Targets, Constraints, Normalization Rules, and Aggregation Rules as Profile Revision-owned content without independent identity, revision, or lifecycle;
* Measurement Records as immutable specializations of Validation Execution Record;
* Quality Claims as immutable specializations of Validation Claim, with exactly one Subject;
* derived quality score and quality evaluation state, with no mutable field on any evaluated Subject.

A Product contract conforms to this specification when it does not contradict the defined Quality semantics and does not require a mutable quality field on any Subject.

---

# Non-Goals

This specification does not require:

* a universal quality metric catalog;
* a specific scoring formula or weighting scheme;
* every Product to use the same Quality Profile;
* every Requirement to carry a quality expectation;
* a mandatory relationship between Quality Claims and Lifecycle Transitions;
* quality evaluation to occur through automation.

---

# References

This document depends on:

* PEOS-000 — Overview;
* PEOS-001 — Philosophy;
* PEOS-002 — Artifact Model;
* PEOS-005 — Requirement Model;
* PEOS-006 — Validation Model.

This document provides the Quality foundation for:

* PEOS-008 — Runtime Contract.
