# PEOS-005 — Requirement Model

## 1. Abstract

The PEOS Requirement Model defines how required engineering intent is represented, identified, revised, governed, and related within PEOS.

This specification distinguishes a Requirement from its statement, revisions, lifecycle state, allocation, implementation, validation, satisfaction, decisions, engineering commitments, and runtime representations.

The model defines Requirements as identifiable engineering Artifacts whose required intent may become normatively applicable according to Lifecycle, authority, and Applicability.

This specification establishes structural semantics only.

It does not define implementation strategies, validation methods, verification activities, risk evaluation, workflow engines, storage formats, repository layouts, or runtime behavior.

## 2. Purpose

The purpose of this specification is to establish a consistent, inspectable, and implementation-independent model for representing required engineering intent.

This specification defines:

- Requirement identity;
- Requirement content;
- Requirement Statement;
- Requirement Subject;
- Requirement Applicability;
- Requirement Origin;
- Requirement Authority;
- Requirement classification;
- Requirement relationships;
- Requirement allocation;
- Requirement change;
- Requirement lifecycle interaction;
- Requirement supersession;
- Requirement waiver;
- relationships between Requirements and other PEOS concepts.

This specification ensures that Requirements remain stable engineering identities independent of revisions, lifecycle states, storage mechanisms, implementation technologies, runtime environments, and evaluation activities.

## 3. Scope

This specification applies to every Requirement represented within PEOS regardless of engineering discipline, engineering domain, product type, repository implementation, storage technology, runtime platform, organizational process, or tooling environment.

This specification governs:

- identification of Requirements;
- normative structure of Requirement content;
- Requirement evolution;
- Requirement traceability;
- Requirement relationships;
- Requirement applicability;
- Requirement authority;
- interactions between Requirements and other PEOS concepts.

This specification does not define:

- verification activities;
- validation activities;
- test procedures;
- evidence evaluation;
- implementation planning;
- project management;
- workflow execution;
- scheduling;
- configuration management;
- risk management;
- repository implementation.

Those concerns are defined by other PEOS specifications.

## 4. Normative Foundation

This specification extends:

- PEOS-000 — Overview;
- PEOS-001 — Core Concepts;
- PEOS-002 — Artifact Model;
- PEOS-003 — Lifecycle Model;
- PEOS-004 — Decision Model.

The concepts defined by those specifications are normative.

If any conflict appears between this specification and a previously established PEOS specification, the definitions established by the earlier specification take precedence unless explicitly superseded by a future PEOS specification.

This specification introduces no exception to the normative semantics established by PEOS-002, PEOS-003, or PEOS-004.

## 5. Requirement Model Overview

```
   Artifact
   └── Requirement

Requirement
    is
        Artifact
    participates in
        Requirement Relationships

Artifact Revision
    whose Artifact is Requirement
    contains
        Requirement Statement
        Requirement Subject
        Requirement Applicability
        Requirement Origin
        Requirement Authority
        Requirement Classification
        Requirement Rationale

Requirement Lifecycle
governed by
Lifecycle State Assignment

Decision Outcome
MAY establish, approve, interpret, waive,
supersede, or remove normative applicability
associated with required engineering intent

Requirement Allocation
assigns realization or responsibility scope

Future Validation Model
evaluates explicit Claims concerning
Requirement Artifact Revisions

Future Risk Model
evaluates engineering risk associated with
Requirements and their satisfaction
```

A Requirement is an identifiable engineering Artifact expressing required engineering intent for a defined Subject and Applicability.

A Requirement remains distinct from:

- its Requirement Statement;
- its Artifact Revision;
- its Lifecycle State;
- its Representation;
- its Allocation;
- its implementation;
- its validation;
- its satisfaction;
- its associated Decisions;
- its associated Engineering Commitments.

A Requirement MAY exist regardless of whether it is currently applicable, implemented, validated, satisfied, allocated, or approved.

## 6. Requirement

A Requirement is an identifiable engineering Artifact concerning required engineering intent for one or more defined Subjects within a defined Applicability.

A Requirement represents what is required.

A Requirement does not represent:

- implementation;
- implementation planning;
- engineering decisions;
- engineering commitments;
- allocation;
- evidence;
- validation;
- verification;
- lifecycle history;
- runtime execution.

A Requirement SHALL preserve its identity independently of changes to:

- its Statement;
- its Artifact Revision;
- its Representation;
- its Lifecycle State;
- its Allocation;
- its storage location;
- its implementation;
- its satisfaction status.

A Requirement MAY exist in proposed, approved, rejected, withdrawn, superseded, obsolete, or other lifecycle states defined by an applicable Lifecycle Definition.

The existence of a Requirement does not imply that its required intent is currently applicable.

Normative applicability depends upon:

- satisfied Requirement Applicability conditions;
- a Lifecycle State that permits normative effect;
- Requirement Authority where required by applicable governance.

Applicable Product contracts and organizational rules MAY define the governance under which these conditions are interpreted and enforced.

## 7. Requirement Identity

Every Requirement SHALL possess a stable identity.

Requirement identity SHALL remain independent of:

- current wording;
- document structure;
- repository organization;
- file path;
- filename;
- storage technology;
- Representation;
- Artifact Revision;
- Lifecycle State;
- implementation;
- evaluation result.

Requirement identity SHALL NOT be derived solely from:

- Requirement Statement;
- document numbering;
- paragraph numbering;
- document ordering;
- repository location;
- runtime identifier.

Requirement identity SHALL remain historically inspectable.

A new Artifact Revision SHALL NOT create a new Requirement identity.

A lifecycle transition SHALL NOT create a new Requirement identity.

Requirement supersession SHALL NOT merge Requirement identities.

## 8. Requirement as Artifact

Every Requirement is an Artifact as defined by PEOS-002.

All normative rules governing Artifacts apply equally to Requirements unless explicitly overridden by this specification.

Requirement content SHALL be represented through Artifact Revisions.

Every content element describing the engineering meaning of a Requirement SHALL belong to an Artifact Revision rather than directly to the Requirement identity.

Requirement content includes, but is not limited to:

- Requirement Statements;
- Requirement Subjects;
- Requirement Applicability;
- Requirement Origins;
- Requirement Authority;
- Requirement Classifications;
- Requirement Rationale.

The Requirement identity SHALL NOT contain mutable current-value properties duplicating content owned by an Artifact Revision.

References to a Requirement defining, possessing, or including such content SHALL be interpreted as references to the applicable Artifact Revision of that Requirement.

Requirement Representations belong to Artifact Revisions rather than directly to the Requirement.

Changes to Requirement content SHALL create a new Artifact Revision according to PEOS-002.

Lifecycle changes SHALL be represented according to PEOS-003.

A Requirement SHALL NOT introduce an independent revision mechanism parallel to Artifact Revision.

The informal expression:

> Requirement Revision

MAY be used as shorthand for:

> Artifact Revision whose Artifact is a Requirement.

It does not define a separate engineering entity.

## 9. Requirement Statement

Every Artifact Revision whose Artifact is a Requirement SHALL contain at least one Requirement Statement.

A Requirement Statement is the human-readable expression of the engineering intent represented by a Requirement.

A Requirement Statement is descriptive content.

A Requirement Statement is not an engineering identity.

A Requirement Statement SHALL belong to exactly one Artifact Revision.

The existence of a Requirement identity does not by itself imply the existence of a complete or approved Requirement Statement.

A Requirement SHALL NOT own Requirement Statements directly.

Different Artifact Revisions of the same Requirement MAY contain different Requirement Statements.

Modification of a Requirement Statement constitutes a content change and SHALL create a new Artifact Revision in accordance with PEOS-002.

A Requirement Statement SHALL NOT possess an independent lifecycle, identifier, revision history, applicability, authority, or relationships separate from its enclosing Artifact Revision.

A Requirement Statement MAY be represented using one or more Representations as defined by PEOS-002.

### 9.1 Requirement Statement Semantics

A Requirement Statement expresses the required engineering intent represented by the Requirement.

The Requirement Statement communicates the intent.

The Requirement defines the identity.

The Statement SHALL NOT be interpreted as the Requirement itself.

Different Statements MAY express the same Requirement.

Equivalent wording SHALL NOT create a new Requirement.

Improved wording SHALL NOT create a new Requirement.

Formatting changes SHALL NOT create a new Requirement.

### 9.2 Statement Consistency

Every Requirement Statement SHALL be internally consistent.

A Requirement Statement SHALL avoid ambiguity to the greatest extent reasonably achievable.

A Requirement Statement SHALL avoid expressing multiple unrelated normative intents within a single Statement.

Where multiple independent normative intents exist, they SHOULD be represented as distinct Requirements unless their separation would alter the intended engineering semantics.

## 10. Requirement Subject

Every Artifact Revision whose Artifact is a Requirement SHALL define the engineering Subject to which the required intent represented by that Artifact Revision applies.

The Subject identifies what is constrained, required, prohibited, or otherwise normatively governed.

Requirement Subject is Requirement content owned by the Artifact Revision in which it is defined.

The Subject MAY be:

- an Artifact;
- a Product;
- a Component;
- a Service;
- an Interface;
- a Process;
- a Capability;
- a Configuration Item;
- another engineering concept defined by PEOS.

The Subject SHALL be identifiable.

The Subject SHALL NOT be inferred solely from informal wording where explicit identification is reasonably possible.

### 10.1 Subject Independence

Requirement identity SHALL remain independent of its Subject.

Changing the Subject constitutes a content change.

Changing the Subject SHALL create a new Artifact Revision.

Changing the Subject SHALL NOT create a new Requirement unless engineering identity changes.

### 10.2 Subject vs Allocation Target

Requirement Subject and Allocation Target are distinct concepts.

The Subject defines what the Requirement governs.

Allocation defines responsibility or realization assignment.

Allocation SHALL NOT redefine the Subject.

Changing Allocation SHALL NOT change the Subject.

## 11. Requirement Applicability

Every Artifact Revision whose Artifact is a Requirement SHALL define the conditions under which the required engineering intent represented by that Artifact Revision applies.

Applicability determines the engineering circumstances in which the Requirement is normatively effective.

Requirement Applicability is Requirement content owned by the Artifact Revision in which it is defined.

Applicability MAY be:

- explicit;
- inherited;
- deterministically derived.

Applicability SHALL be objectively determinable.

Applicability SHALL NOT depend upon subjective interpretation.

Applicability SHALL NOT depend upon implementation status.

Applicability SHALL remain independent of validation results.

Applicability SHALL remain independent of allocation.

### 11.1 Applicable Requirement

A Requirement becomes applicable only when its Applicability conditions are satisfied, its Lifecycle permits normative effect, and any Authority required by applicable governance has been established.

Applicability SHALL NOT be inferred solely from the existence of the Requirement.

Applicability SHALL NOT imply implementation.

Applicability SHALL NOT imply satisfaction.

Applicability SHALL NOT imply validation.

Applicability SHALL NOT imply conformance.

### 11.2 Applicability Change

Changing Applicability constitutes a Requirement content change.

Such change SHALL create a new Artifact Revision.

Applicability change SHALL NOT create a new Requirement identity unless engineering intent changes fundamentally.

## 12. Requirement Origin

Every Artifact Revision whose Artifact is a Requirement MAY define one or more Requirement Origins.

Requirement Origin identifies the engineering source from which the Requirement was derived or motivated.

Origin describes provenance.

Requirement Origin is Requirement content owned by the Artifact Revision in which it is defined.

Origin does not establish authority.

Origin does not itself constitute a Decision, Decision Outcome, Engineering Commitment, or normative approval.

A reference to a Decision as an Origin records provenance only.

Where a Decision Outcome establishes a Requirement, the Decision relationship and the Requirement Origin SHALL remain semantically distinguishable.

Possible Origins include:

- customer needs;
- contractual obligations;
- regulations;
- standards;
- engineering analysis;
- risk analysis;
- safety analysis;
- stakeholder requests;
- previously established Requirements;
- Decisions.

### 12.1 Origin Independence

Origin SHALL remain independent from:

- Authority;
- Lifecycle;
- Applicability;
- Validation;
- Allocation.

Changing Origin SHALL constitute a content change.

Changing Origin SHALL create a new Artifact Revision.

## 13. Requirement Authority

Requirement Authority identifies the authority under which the Requirement becomes normatively binding.

Requirement Authority is Requirement content owned by the Artifact Revision in which it is defined.

Authority establishes normative legitimacy.

Authority does not establish origin.

Authority does not establish authorship.

Authority does not establish ownership.

Authority SHALL be explicitly identifiable whenever normative effect depends upon organizational governance.

Multiple Authorities MAY jointly provide the normative authority represented by the same Requirement Artifact Revision.

### 13.1 Authority vs Origin

Origin answers:

> Where did this Requirement come from?

Authority answers:

> Under whose authority is this Requirement normatively binding?

These concepts SHALL NOT be treated as interchangeable.

### 13.2 Authority Change

Changing Authority constitutes a Requirement content change.

Authority change SHALL create a new Artifact Revision.

Authority change SHALL NOT create a new Requirement identity.

## 14. Requirement Classification

An Artifact Revision whose Artifact is a Requirement MAY contain one or more Requirement Classifications.

Requirement Classification is Requirement content owned by the Artifact Revision in which it is defined.

Classification provides organizational structure.

Classification SHALL NOT affect:

- identity;
- applicability;
- authority;
- lifecycle;
- revision semantics.

Possible classifications include:

- Functional;
- Non-functional;
- Safety;
- Security;
- Regulatory;
- Performance;
- Reliability;
- Interface;
- Operational.

This specification does not standardize classification taxonomies.

## 15. Requirement Rationale

An Artifact Revision whose Artifact is a Requirement MAY include documented engineering rationale.

Requirement Rationale is Requirement content owned by the Artifact Revision in which it is defined.

Rationale explains why the Requirement exists.

Rationale SHALL NOT modify normative meaning.

Rationale SHALL NOT replace the Requirement Statement.

Changing Rationale constitutes a content change.

Such change SHALL create a new Artifact Revision.

## 16. Requirement Quality

Every Requirement Statement SHOULD exhibit characteristics supporting objective engineering interpretation.

Examples of desirable characteristics include:

- clarity;
- consistency;
- necessity;
- singularity of intent;
- objective interpretability;
- traceability;
- verifiability by future Validation models.

This specification does not define a complete Requirement quality taxonomy or a mandatory Requirement quality assessment method.

Failure to satisfy one or more quality characteristics does not invalidate the existence of the Requirement.

However, poor Requirement quality increases engineering ambiguity and SHOULD be corrected through normal Artifact Revision procedures.

## 17. Requirement Relationships

Requirements MAY participate in one or more explicitly defined engineering relationships.

Relationships establish engineering semantics between identifiable entities.

Relationships SHALL be explicit.

Relationships SHALL NOT be inferred solely from textual similarity.

Relationships SHALL remain independent of:

- Lifecycle;
- Representation;
- Artifact Revision;
- Validation;
- Allocation.

Relationship semantics SHALL be determined by their defined relationship type.

### 17.1 Relationship Representation

This specification defines the engineering semantics of relationships involving Requirements.

This specification does not define relationships as a separate category of engineering entity.

The identity, revision, lifecycle, historical preservation, authority, and representation semantics of relationships are outside the scope of this specification.

A conforming implementation SHALL preserve sufficient information to distinguish explicitly represented relationships and their engineering meaning.

Changing or removing a relationship SHALL NOT by itself modify the identity or content of either related Requirement.

A future PEOS specification MAY define a common model for engineering relationships.

### 17.2 Relationship Direction

Relationships MAY be directional.

Where direction exists, reversing direction changes engineering meaning.

Relationship direction SHALL NOT be ignored.

Relationship semantics SHALL define the meaning of both source and target.

## 18. Derivation

A Requirement MAY be derived from one or more Requirements.

Derivation expresses engineering refinement of engineering intent.

Derivation SHALL preserve traceable engineering rationale.

A derived Requirement SHALL possess its own identity.

A derived Requirement SHALL NOT inherit the identity of its source Requirement.

Derivation SHALL NOT imply decomposition.

Derivation SHALL NOT imply implementation.

Derivation SHALL NOT imply allocation.

Derivation SHALL NOT imply satisfaction.

### 18.1 Multiple Derivation

A Requirement MAY be derived from multiple source Requirements.

Such derivation SHALL remain explicitly represented.

Derived engineering intent SHALL remain inspectable.

## 19. Refinement

A Requirement MAY refine another Requirement.

Refinement expresses increased engineering precision while preserving compatible engineering intent.

Refinement SHALL NOT replace the refined Requirement unless explicitly superseded.

A refined Requirement SHALL remain independently identifiable.

Refinement SHALL NOT imply implementation.

Refinement SHALL NOT imply decomposition.

Refinement SHALL NOT imply satisfaction.

## 20. Decomposition

A Requirement MAY be decomposed into multiple subordinate Requirements.

Decomposition divides engineering intent into smaller independently identifiable Requirements.

Each decomposed Requirement SHALL possess its own Requirement identity.

The parent Requirement SHALL remain independently identifiable.

Decomposition SHALL NOT imply that the parent Requirement is satisfied when every child Requirement is satisfied.

Satisfaction semantics are defined by the Validation Model.

### 20.1 Decomposition Completeness

This specification does not require decomposition to be complete.

Engineering completeness SHALL be determined by applicable engineering processes rather than by the existence of decomposition relationships.

## 21. Dependency

A Requirement MAY depend upon one or more Requirements.

Dependency expresses engineering reliance.

Dependency SHALL remain explicitly represented.

Dependency SHALL NOT imply derivation.

Dependency SHALL NOT imply refinement.

Dependency SHALL NOT imply allocation.

Dependency SHALL NOT imply implementation order.

## 22. Conflict

Two or more Requirements MAY conflict.

Conflict indicates that simultaneous normative satisfaction cannot be assumed without engineering resolution.

Conflict SHALL remain explicitly represented.

Conflict SHALL NOT automatically invalidate any Requirement.

Conflict SHALL NOT automatically withdraw any Requirement.

Conflict resolution SHALL occur through applicable engineering governance.

## 23. Supersession

A Requirement MAY supersede one or more Requirements through an explicitly represented Supersession relationship.

Supersession establishes that the required engineering intent represented by an identified superseding Requirement Artifact Revision replaces the required engineering intent represented by one or more identified superseded Requirement Artifact Revisions within a defined scope.

Every Supersession relationship SHALL identify:

- the superseding Requirement;
- the superseding Requirement Artifact Revision;
- each superseded Requirement;
- each superseded Requirement Artifact Revision;
- the scope within which replacement applies;
- the governance action under which Supersession was established.

Supersession SHALL NOT be inferred solely from:

- creation of a newer Artifact Revision;
- similarity of Requirement Statements;
- document order;
- identifier order;
- Lifecycle State;
- archival status.

Superseded Requirements SHALL retain their identities.

Supersession SHALL preserve engineering history.

Supersession SHALL NOT merge Requirement identities.

Supersession SHALL NOT delete historical Artifact Revisions.

Supersession SHALL remain explicitly represented.

A Decision Outcome MAY authorize Supersession as defined by Section 28.3.

### 23.1 Superseded Requirement

A superseded Requirement remains an identifiable Requirement.

A Supersession relationship does not erase the Requirement, its Artifact Revisions, or its engineering history.

A superseded Requirement SHALL NOT be interpreted as universally non-applicable solely because a Supersession relationship exists.

Its normative applicability SHALL be determined according to:

- the scope of the applicable Supersession relationship;
- its Requirement Applicability conditions;
- its Lifecycle State;
- Requirement Authority where required by applicable governance.

Superseded Requirements and their Artifact Revisions MAY remain referenced by historical Artifacts, Decisions, Engineering Commitments, Validation Records, and other engineering information.

## 24. Requirement Allocation

Requirements MAY be allocated to engineering entities.

Allocation establishes realization responsibility or engineering assignment.

Allocation SHALL remain independent of Requirement identity.

Allocation SHALL remain independent of Applicability.

Allocation SHALL remain independent of Validation.

Allocation SHALL remain independent of Satisfaction.

Allocation SHALL NOT redefine engineering intent.

Allocation SHALL NOT modify the Requirement Statement.

Changing Allocation SHALL NOT require creation of a new Requirement.

### 24.1 Allocation Targets

Allocation targets MAY include:

- Products;
- Components;
- Services;
- Organizations;
- Teams;
- Processes;
- Interfaces;
- Configuration Items;
- other engineering entities.

This specification does not standardize Allocation target taxonomies.

### 24.2 Allocation Semantics

Allocation expresses engineering assignment.

Allocation SHALL NOT imply:

- implementation completion;
- implementation correctness;
- implementation approval;
- Requirement satisfaction;
- Requirement verification;
- Requirement validation.

Allocation only establishes an engineering relationship.

## 25. Requirement Change

A Requirement evolves through controlled engineering change.

Requirement change SHALL preserve Requirement identity unless a fundamentally different engineering intent is established.

Requirement change SHALL be represented by Artifact Revisions as defined by PEOS-002.

Requirement change SHALL NOT be represented by Lifecycle transitions.

Requirement change SHALL NOT be represented by Allocation changes.

Requirement change SHALL NOT be represented by Validation results.

Requirement change SHALL preserve historical inspectability.

### 25.1 Material Change

A material change is any change that modifies the engineering meaning represented by a Requirement Artifact Revision.

Material changes include, but are not limited to, changes to:

- Requirement Statement;
- Subject;
- Applicability;
- Authority;
- Origin;
- normative constraints;
- required behavior;
- required qualities;
- thresholds;
- limits;
- conditions;
- units of measure;
- normative strength;
- engineering rationale where that rationale affects interpretation of required intent.

Every material change SHALL create a new Artifact Revision.

A change SHALL be classified according to its effect on engineering meaning rather than according to the field, representation, or storage location in which the change occurs.

### 25.2 Non-Material Change

Non-material changes do not modify the engineering content of the Requirement.

Examples include:

- lifecycle transitions;
- representation improvements;
- formatting corrections;
- metadata corrections not affecting engineering meaning.

Non-material changes SHALL NOT create a new Requirement identity.

Whether a new Artifact Revision is required is governed by PEOS-002.

## 26. Requirement Lifecycle

Requirement lifecycle SHALL be governed exclusively by PEOS-003.

This specification defines no independent lifecycle mechanism.

Lifecycle State Assignments SHALL remain separate from Artifact Revisions.

Requirement lifecycle SHALL remain independent of:

- implementation;
- validation;
- satisfaction;
- allocation;
- runtime execution.

Lifecycle SHALL determine governance state.

Lifecycle SHALL NOT determine engineering correctness.

### 26.1 Proposed Requirement

An applicable Lifecycle Definition MAY define a Proposed Lifecycle State.

A Requirement assigned such a state remains a valid Requirement identity.

A Proposed Requirement MAY participate in:

- traceability;
- derivation;
- refinement;
- decomposition;
- Decisions;
- engineering discussion.

Assignment of a Proposed Lifecycle State SHALL NOT by itself establish normative applicability.

### 26.2 Applicable Requirement

Normative applicability depends upon all of the following:

- satisfied Requirement Applicability conditions;
- a Lifecycle State that permits normative effect;
- Requirement Authority where required by applicable governance.

No individual factor is sufficient by itself.

The absence of a separately represented Requirement Authority SHALL NOT prevent normative applicability where applicable governance does not require such Authority.

A Requirement SHALL NOT become applicable solely because it exists.

### 26.3 Rejected Requirement

An applicable Lifecycle Definition MAY define a Rejected Lifecycle State.

Assignment of such a state SHALL NOT destroy Requirement identity or Requirement Artifact Revisions.

Rejected Requirements SHALL remain historically inspectable.

Rejected Requirements and their Artifact Revisions MAY remain referenced by historical Decisions, Engineering Commitments, Validation Records, and other engineering information.

Any later Lifecycle transition SHALL be governed exclusively by the applicable Lifecycle Definition and PEOS-003.

### 26.4 Withdrawn Requirement

An applicable Lifecycle Definition MAY define a Withdrawn Lifecycle State.

While a Requirement is assigned a Lifecycle State whose semantics prohibit normative effect, the Requirement SHALL NOT possess normative applicability.

Withdrawal SHALL NOT erase Requirement identity, Requirement Artifact Revisions, Lifecycle history, or other engineering history.

Withdrawal SHALL NOT by itself invalidate historical Decisions, Engineering Commitments, Validation Records, or engineering actions established while the Requirement possessed normative applicability.

Any later Lifecycle transition SHALL be governed exclusively by the applicable Lifecycle Definition and PEOS-003.

### 26.5 Superseded Requirement

An applicable Lifecycle Definition MAY define a Lifecycle State indicating that a Requirement has been superseded.

Assignment of such a Lifecycle State records governance state only.

A Lifecycle State Assignment SHALL NOT by itself establish:

- which Requirement supersedes another Requirement;
- which Requirement Artifact Revisions are involved;
- the scope of replacement;
- the authority or governance action establishing replacement.

Normative replacement of required engineering intent SHALL be represented through an explicit Supersession relationship as defined by Section 23.

The Lifecycle State Assignment and the applicable Supersession relationship SHALL remain semantically distinct and independently inspectable.

## 27. Waiver

A Requirement MAY be waived only through applicable engineering governance.

A waiver SHALL be authorized by an established Decision Outcome or by another governance mechanism explicitly permitted by an applicable PEOS Product contract.

A waiver suspends or limits normative applicability within its defined scope.

A waiver SHALL NOT delete the Requirement.

A waiver SHALL NOT change Requirement identity.

A waiver SHALL NOT rewrite Requirement content.

Waivers SHALL remain explicitly represented.

A waiver SHALL identify the governance action through which it was established.

### 27.1 Waiver Authority

Every waiver SHALL identify the authority under which it is established.

Where a waiver is established by a Decision Outcome, the Decision and applicable Outcome SHALL remain explicitly traceable.

A waiver established without authority required by applicable governance SHALL NOT possess normative effect.

### 27.2 Waiver Scope

A waiver SHALL define its scope.

Scope MAY include:

- products;
- configurations;
- deployments;
- projects;
- operational contexts;
- defined periods of applicability.

Waivers SHALL NOT be interpreted beyond their explicitly established scope.

## 28. Requirement and Decision

Requirements and Decisions represent different engineering concepts.

A Requirement defines required engineering intent.

A Decision establishes an engineering determination.

Neither concept SHALL replace the other.

### 28.1 Decision Establishing Requirement

A Decision Outcome MAY establish one or more Requirements.

Such establishment SHALL preserve independent Requirement identities.

The Decision SHALL NOT become the Requirement.

The Requirement SHALL NOT become the Decision.

### 28.2 Decision Interpreting Requirement

A Decision MAY establish authoritative interpretation of an existing Requirement.

Interpretation SHALL NOT automatically modify Requirement content.

Modification of engineering intent SHALL require a new Artifact Revision.

### 28.3 Decision Superseding Requirement

A Decision Outcome MAY authorize Requirement supersession.

The supersession SHALL remain explicitly represented.

Requirement identity SHALL remain preserved.

## 29. Requirement and Engineering Commitment

Requirements and Engineering Commitments are distinct engineering entities.

Requirements define required engineering intent.

Engineering Commitments define normatively established obligations created through Decision Outcomes.

Engineering Commitments MAY reference Requirements.

Engineering Commitments SHALL NOT replace Requirements.

Requirements SHALL NOT replace Engineering Commitments.

### 29.1 Requirement Supporting Commitment

An Engineering Commitment MAY reference one or more Requirements.

Such references SHALL preserve independent identities.

Commitments SHALL remain independently governed by the Decision Model.

### 29.2 Commitment Independence

Modification of an Engineering Commitment SHALL NOT modify a Requirement.

Modification of a Requirement SHALL NOT modify an Engineering Commitment.

Any relationship between the two SHALL remain explicitly represented.

## 30. Requirement and Validation

Validation is outside the scope of this specification.

This specification defines only the structural boundary between Requirement content and future Validation concepts.

The future Validation Model SHALL evaluate explicit engineering claims concerning identified Requirement Artifact Revisions.

Such claims MAY include Satisfaction Claims and Conformance Claims.

A Validation activity, result, claim, or evidence item SHALL NOT become part of Requirement content solely because it concerns a Requirement.

Requirements and their Artifact Revisions SHALL remain independent of Validation outcomes.

### 30.1 Validation Independence

Validation SHALL NOT determine:

- Requirement identity;
- Requirement content;
- Requirement Applicability;
- Requirement Authority;
- Requirement existence.

Validation evaluates explicit engineering claims concerning identified Requirement Artifact Revisions.

Validation does not define required engineering intent.

### 30.2 Satisfaction

Requirement satisfaction is not defined by this specification.

A future Validation Model SHALL define Satisfaction Claims and the evaluation of those Claims.

A Satisfaction Claim SHALL identify the Requirement Artifact Revision whose required engineering intent is being evaluated.

A Requirement SHALL NOT possess mutable satisfaction state.

An Artifact Revision SHALL NOT possess mutable satisfaction state.

This specification intentionally defines no property equivalent to:

```
Requirement.satisfied = true
```

or:

```
ArtifactRevision.satisfied = true
```

Satisfaction SHALL be represented through explicit Validation concepts rather than mutable Requirement or Artifact Revision attributes.

## 31. Requirement and Risk

Risk assessment is outside the scope of this specification.

Requirements and Risks are independent engineering entities.

A Requirement MAY reference one or more Risks.

A Risk MAY reference one or more Requirements.

Neither entity SHALL derive its identity from the other.

Modification of a Risk SHALL NOT modify a Requirement.

Modification of a Requirement SHALL NOT modify a Risk.

Risk evaluation SHALL be defined by a future PEOS Risk specification.

## 32. Requirement and Runtime

Requirements exist independently of runtime systems.

Runtime behavior SHALL NOT determine Requirement identity.

Runtime behavior SHALL NOT determine Requirement applicability.

Runtime behavior SHALL NOT determine Requirement authority.

Runtime behavior MAY provide evidence relevant to future Validation activities.

Runtime systems SHALL NOT be treated as the authoritative source of Requirement semantics.

## 33. Historical Preservation

Requirement history SHALL remain inspectable.

Historical Artifact Revisions SHALL remain distinguishable.

Historical Lifecycle State Assignments SHALL remain distinguishable.

The existence, participants, direction, type, scope, and engineering meaning of explicitly represented historical Requirement relationships SHALL remain inspectable where those relationships are retained as engineering information.

This requirement does not define independent revision, lifecycle, or temporal-history semantics for relationships.

Such semantics remain outside the scope of this specification as defined by Section 17.1.

Historical Decisions referencing Requirements SHALL remain inspectable.

Historical Engineering Commitments SHALL remain inspectable.

Historical Validation Records SHALL be governed by the Validation Model.

Historical preservation SHALL NOT require historical Requirements to remain applicable.

## 34. Conformance

An implementation claiming conformance to this specification SHALL satisfy all mandatory requirements defined using the keywords SHALL and SHALL NOT.

Optional capabilities defined using MAY are not required for conformance.

Recommended behavior expressed using SHOULD is not mandatory but SHOULD be followed unless a documented engineering rationale justifies deviation.

Conformance SHALL be evaluated independently of implementation technology, repository structure, storage mechanism, programming language, runtime platform, or organizational process.

### 34.1 Minimum Conformance

A conforming implementation SHALL, at minimum:

- represent Requirements as Artifacts;
- preserve stable Requirement identities;
- represent Requirement content through Artifact Revisions;
- govern lifecycle through PEOS-003;
- distinguish Requirement Statements from Requirements;
- distinguish Requirement Authority from Requirement Origin;
- distinguish Applicability from Lifecycle;
- preserve historical revisions;
- preserve historical lifecycle transitions;
- preserve explicitly represented Requirement relationships and their engineering meaning.

## 35. Invariants

The following invariants SHALL always hold.

**Identity**

A Requirement SHALL possess exactly one stable identity.

Identity SHALL survive:

- content revision;
- lifecycle transition;
- representation change;
- allocation change;
- validation;
- implementation.

**Revision**

Requirement content SHALL evolve exclusively through Artifact Revisions.

Lifecycle transitions SHALL NOT replace Artifact Revisions.

**Lifecycle**

Lifecycle SHALL remain independent of:

- validation;
- implementation;
- allocation;
- satisfaction.

**Supersession**

A Lifecycle State Assignment SHALL NOT by itself establish Supersession.

Supersession SHALL identify the Requirement Artifact Revisions whose required engineering intent is replaced and the scope in which replacement applies.

Supersession SHALL NOT merge or destroy Requirement identities.

**Statement**

Requirement Statement SHALL NOT define Requirement identity.

**Applicability**

Applicability SHALL remain independent from:

- implementation;
- validation;
- allocation.

**Allocation**

Allocation SHALL NOT imply:

- implementation;
- correctness;
- satisfaction;
- validation.

**Validation**

Validation SHALL evaluate explicit Claims concerning identified Requirement Artifact Revisions.

Validation SHALL NOT define Requirement identity or Requirement content.

Satisfaction and Conformance SHALL NOT be represented as mutable Requirement or Artifact Revision state.

**Decision**

Decisions SHALL remain independent of Requirements.

Neither entity SHALL replace the other.

**Engineering Commitment**

Engineering Commitments SHALL remain independent of Requirements.

**Waiver**

A waiver SHALL NOT delete, rewrite, or replace the Requirement to which it applies.

A waiver SHALL possess normative effect only through applicable engineering governance and only within its defined scope.

**Historical Preservation**

Historical engineering information SHALL remain inspectable.

## 36. Non-Conforming Patterns

The following implementation patterns violate this specification.

### 36.1 Requirement Equals Statement

Treating a Requirement Statement as the Requirement itself.

### 36.2 Mutable Satisfaction

Representing Requirement satisfaction as mutable Requirement state.

Example:

```
Requirement {
satisfied = true
}
```

### 36.3 Lifecycle as Revision

Representing engineering content change using lifecycle transitions.

### 36.4 Revision as Lifecycle

Representing lifecycle transitions by creating new Requirement Artifact Revisions.

### 36.5 Allocation as Satisfaction

Treating allocation as evidence that a Requirement has been satisfied.

### 36.6 Decision Equals Requirement

Treating Decisions as Requirements.

### 36.7 Requirement Equals Engineering Commitment

Treating Engineering Commitments as Requirements.

### 36.8 Authority Equals Origin

Treating Authority and Origin as interchangeable concepts.

### 36.9 Validation Changes Requirement

Allowing validation results to modify Requirement semantics.

### 36.10 Runtime Defines Requirement

Allowing runtime behavior to determine Requirement identity or normative meaning.

### 36.11 Ungoverned Waiver

Representing a waiver as an informal flag or mutable Requirement attribute without explicit scope, authority, and applicable engineering governance.

Example:

```
Requirement {
waived = true
}
```

### 36.12 Lifecycle-Only Supersession

Treating assignment of a Superseded Lifecycle State as sufficient to establish which Requirement, Requirement Artifact Revision, and scope replace another required engineering intent.

### 36.13 Implicit Supersession

Inferring Supersession solely from a newer Artifact Revision, document order, identifier order, Statement similarity, archival status, or Lifecycle State without an explicitly represented Supersession relationship.

## 37. Summary

This specification defines Requirements as identifiable engineering Artifacts expressing required engineering intent.

This specification defines only the structural semantics of Requirements and their interaction with other PEOS concepts.

Requirement identity is preserved independently of revisions, lifecycle states, representations, allocation, implementation, validation, and runtime behavior.

Requirement content evolves through Artifact Revisions.

Requirement governance is performed through the Lifecycle Model.

Requirements remain distinct from:

- Requirement Statements;
- Decisions;
- Engineering Commitments;
- Validation;
- Satisfaction;
- Allocation;
- Risks;
- Runtime behavior.

This specification intentionally separates engineering intent from governance, implementation, validation, and operational execution.

Future PEOS specifications extend this model by defining validation, risk, configuration, traceability, and other engineering concerns while preserving the invariants established by this specification.
