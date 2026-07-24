# PEOS-005 — Requirement Model

## 1. Abstract

The PEOS Requirement Model defines how required engineering intent is represented, identified, revised, governed, and related within PEOS.

This specification distinguishes a Requirement from its statement, revisions, lifecycle state, allocation, implementation, validation, satisfaction, decisions, engineering commitments, and runtime representations.

The model defines the identity, revisioned content, relationships, governance interactions, and normative applicability of Requirements as defined in Section 6.

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
State Assignment

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

A Requirement is an identifiable engineering Artifact whose Artifact Revisions may express required engineering intent for one or more defined Subjects within a defined Applicability.

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

Every Artifact Revision whose Artifact is a Requirement SHALL identify one or more engineering Subjects to which the required engineering intent represented by that Artifact Revision applies.

A Subject identifies the engineering matter, system, component, service, behavior, interface, process, information item, organization, actor, environment, or other engineering concern governed by the represented required intent.

Where multiple Subjects are identified, each Subject SHALL remain distinguishable.

The identification of multiple Subjects SHALL NOT by itself create multiple Requirements.

The represented required engineering intent SHALL make clear whether it applies:

- independently to each identified Subject;
- collectively to the identified Subjects;
- or according to another explicitly represented relationship among those Subjects.

Subject semantics SHALL NOT be inferred solely from textual ordering.

### 10.1 Subject and Requirement Identity

Requirement identity SHALL remain independent of the Subjects identified by any individual Requirement Artifact Revision.

Changing, adding, removing, or otherwise modifying an identified Subject constitutes a Requirement content change and therefore requires a new Artifact Revision.

Such a content change SHALL NOT by itself create a new Requirement identity.

A new Requirement identity is required only where applicable identity governance determines that the engineering requirement itself is no longer the same Requirement.

### 10.2 Subject and Allocation Target

A Requirement Subject SHALL remain distinct from an Allocation Target.

A Subject identifies what the required engineering intent concerns.

An Allocation Target identifies an engineering element assigned responsibility for realizing, addressing, or otherwise carrying that required intent.

A Requirement MAY concern one or more Subjects without being allocated.

A Requirement MAY be allocated to one or more Allocation Targets that differ from its identified Subjects.

Subject identification SHALL NOT imply Allocation.

Allocation SHALL NOT modify Requirement Subject content.

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
- verifiability by the future Validation Model.

This specification does not define a complete Requirement quality taxonomy or a mandatory Requirement quality assessment method.

Failure to satisfy one or more quality characteristics does not invalidate the existence of the Requirement.

However, poor Requirement quality increases engineering ambiguity and SHOULD be corrected through normal Artifact Revision procedures.

## 17. Requirement Relationships

Requirements MAY participate in one or more explicitly defined engineering relationships.

Relationships establish engineering semantics between identifiable entities.

Relationships SHALL be explicit.

Relationships SHALL NOT be inferred solely from textual similarity.

The existence of a Requirement relationship SHALL remain independent of:

- State Assignment;
- Representation;
- Validation;
- Allocation.

A Requirement relationship SHALL NOT become part of Requirement content solely because the relationship concerns that Requirement.

A relationship MAY identify a Requirement identity, an identified Requirement Artifact Revision, or both, according to the semantics of its relationship type.

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

### 17.3 Relationship Participant Level

Every explicitly represented Requirement relationship SHALL make clear whether each participant identifies:

- a Requirement identity;
- a specific Requirement Artifact Revision;
- or both.

Where relationship meaning depends upon represented engineering intent, the relationship SHALL identify the applicable Requirement Artifact Revision.

Where relationship meaning applies to the Requirement identity independently of any individual content revision, identifying the Requirement identity is sufficient.

A relationship SHALL NOT leave participant level ambiguous.

Identification of a Requirement Artifact Revision SHALL NOT create a separate Requirement Revision entity.

### 17.4 Artifact Relation Conformance

Every explicitly represented Requirement relationship defined by this specification is an Artifact Relation governed by PEOS-002.

In addition to the relationship-type-specific information defined in Sections 18 through 23, every Requirement relationship SHALL identify:

- its Relation Type;
- its provenance;
- its applicable scope where that scope is not self-evident from the identified participants and relationship semantics.

Relationship provenance SHALL identify the origin of the represented relationship fact.

Relationship provenance SHALL NOT be inferred solely from the provenance, authorship, authority, or origin of a participating Requirement or Requirement Artifact Revision.

Every Requirement Relation Type definition SHALL make explicit:

- permitted source participant types;
- permitted target participant types;
- direction;
- multiplicity;
- whether cycles are permitted;
- whether participants are identified at Requirement identity level, Requirement Artifact Revision level, or both;
- applicable integrity constraints.

A relationship SHALL conform both to PEOS-002 and to the additional semantics of its Requirement Relation Type.

Where this specification imposes a stronger constraint than the general Artifact Relation model, the stronger Requirement-specific constraint applies.

Every individual Requirement relationship SHALL be represented as one binary Artifact Relation having exactly one source participant and one target participant.

Where engineering semantics involve multiple sources, multiple targets, or more than two participants, those semantics SHALL be represented through multiple individually identifiable binary Artifact Relations.

A collection of related Artifact Relations MAY be interpreted together for an explicitly defined engineering purpose.

Such a collection SHALL NOT be interpreted as introducing a separate Relationship Set, relationship-group, hyperedge, or other PEOS entity.

Every Artifact Relation within such a collection SHALL independently satisfy the identity, provenance, participant, Relation Type, scope, and integrity requirements applicable to that relation.

The provenance requirement established by this section applies independently to every binary Requirement relationship, including every relationship participating in a collectively interpreted group.

## 18. Derivation

A Requirement Artifact Revision MAY be derived from one or more identified Requirement Artifact Revisions.

Each source-to-derived association SHALL be represented by a separate Derivation relationship.

Multiple Derivation relationships MAY share the same derived Requirement Artifact Revision as their target.

Derivation records that required engineering intent was produced through engineering reasoning using other represented required engineering intent as an input.

Every Derivation relationship SHALL identify:

- one source Requirement;
- one source Requirement Artifact Revision;
- the derived Requirement;
- the derived Requirement Artifact Revision.

Derivation SHALL preserve sufficient engineering rationale to explain how the derived required engineering intent originates from its identified sources.

A derived Requirement SHALL possess its own identity.

A derived Requirement SHALL NOT inherit the identity of a source Requirement.

Derivation does not require the derived intent to be a more precise expression of the source intent.

Derivation SHALL NOT imply:

- Refinement;
- Decomposition;
- Supersession;
- implementation;
- Allocation;
- Satisfaction.

### 18.1 Derivation Relation Type

For Derivation:

- the source participant SHALL be an identified source Requirement Artifact Revision;
- the target participant SHALL be the derived Requirement Artifact Revision;
- direction SHALL be from source to derived;
- each Derivation relationship SHALL have exactly one source Requirement Artifact Revision and exactly one target Requirement Artifact Revision;
- multiple Derivation relationships MAY share the same target Requirement Artifact Revision;
- one source Requirement Artifact Revision MAY be the source of multiple Derivation relationships;
- Derivation cycles SHALL NOT be permitted;
- both source and target SHALL be identified at Requirement Artifact Revision level.

A Requirement Artifact Revision SHALL NOT be directly or transitively derived from itself.

Every source and target Requirement SHALL remain independently identifiable.

### 18.2 Multiple Derivation

Where a Requirement Artifact Revision is derived from multiple source Requirement Artifact Revisions, each source SHALL participate through a separate Derivation relationship targeting that derived Requirement Artifact Revision.

The combined engineering contribution of those sources MAY be interpreted collectively without creating a separate relationship-group entity.

Every source relationship SHALL remain individually identifiable and inspectable.

## 19. Refinement

A Requirement Artifact Revision MAY refine one or more identified Requirement Artifact Revisions.

Refinement of multiple Requirement Artifact Revisions SHALL be represented through one separate Refinement relationship for each refined source Requirement Artifact Revision.

Refinement records that required engineering intent is expressed with increased precision, narrower interpretation, additional constraint, or greater engineering detail while remaining compatible with the refined required engineering intent within the defined scope.

Every Refinement relationship SHALL identify:

- one refined Requirement;
- one refined Requirement Artifact Revision;
- one refining Requirement;
- one refining Requirement Artifact Revision;
- the scope within which compatibility is asserted.

Refinement SHALL NOT be represented where the refining intent contradicts the refined intent within the same scope.

Refinement SHALL NOT replace the refined Requirement or Requirement Artifact Revision unless an explicit Supersession relationship separately establishes replacement.

A refining Requirement SHALL remain independently identifiable.

Refinement MAY also be supported by Derivation, but neither relationship SHALL be inferred solely from the existence of the other.

Refinement SHALL NOT imply:

- Derivation;
- Decomposition;
- Supersession;
- implementation;
- Allocation;
- Satisfaction.

### 19.1 Refinement Relation Type

For Refinement:

- the source participant SHALL be the refined Requirement Artifact Revision;
- the target participant SHALL be the refining Requirement Artifact Revision;
- direction SHALL be from refined to refining;
- each Refinement relationship SHALL have exactly one refined source Requirement Artifact Revision and exactly one refining target Requirement Artifact Revision;
- multiple Refinement relationships MAY share the same refining target Requirement Artifact Revision;
- one refined Requirement Artifact Revision MAY be the source of multiple Refinement relationships;
- Refinement cycles SHALL NOT be permitted;
- both source and target SHALL be identified at Requirement Artifact Revision level.

A Requirement Artifact Revision SHALL NOT directly or transitively refine itself.

The refining required engineering intent SHALL remain compatible with the refined required engineering intent within the defined scope.

## 20. Decomposition

A Requirement Artifact Revision MAY be decomposed into multiple subordinate Requirement Artifact Revisions.

Each parent-to-subordinate association SHALL be represented by a separate Decomposition relationship.

Multiple Decomposition relationships sharing the same parent Requirement Artifact Revision MAY collectively represent one engineering decomposition.

Decomposition records that required engineering intent represented by an identified parent Requirement Artifact Revision is partitioned into multiple independently identifiable subordinate Requirements.

Every Decomposition relationship SHALL identify:

- one parent Requirement;
- one parent Requirement Artifact Revision;
- one subordinate Requirement;
- one subordinate Requirement Artifact Revision;
- the scope of that parent-to-subordinate association.

Each subordinate Requirement SHALL possess its own Requirement identity.

The parent Requirement SHALL remain independently identifiable.

Decomposition SHALL NOT by itself establish:

- that the subordinate Requirements completely cover the parent required engineering intent;
- that the subordinate Requirements are mutually exclusive;
- that Satisfaction of every subordinate Requirement establishes Satisfaction of the parent Requirement;
- implementation structure;
- Allocation;
- Lifecycle state;
- Supersession.

### 20.1 Decomposition Relation Type

For Decomposition:

- the source participant SHALL be the parent Requirement Artifact Revision;
- the target participant SHALL be a subordinate Requirement Artifact Revision;
- direction SHALL be from parent to subordinate;
- each Decomposition relationship SHALL have exactly one parent source Requirement Artifact Revision and exactly one subordinate target Requirement Artifact Revision;
- one parent Requirement Artifact Revision MAY be the source of multiple Decomposition relationships;
- one subordinate Requirement Artifact Revision MAY be the target of more than one Decomposition relationship only where each parent relationship and its scope remain explicitly distinguishable;
- Decomposition cycles SHALL NOT be permitted;
- parent and subordinate participants SHALL be identified at Requirement Artifact Revision level.

A Requirement Artifact Revision SHALL NOT directly or transitively decompose itself.

A subordinate Requirement identity SHALL remain distinct from the parent Requirement identity.

### 20.2 Decomposition Completeness

The existence of one or more Decomposition relationships SHALL NOT by itself imply that decomposition is complete.

Where engineering completeness is required, that completeness SHALL be established only through explicitly represented engineering information.

This specification defines no structural model for representing decomposition completeness.

This specification therefore SHALL NOT be interpreted as introducing:

- a Relationship Collection;
- a Decomposition Set;
- a Completeness Assertion entity;
- or any other PEOS entity representing a group of Decomposition relationships.

The ownership, identity, lifecycle, and representation of engineering completeness remain outside the scope of this specification.

## 21. Dependency

A Requirement or Requirement Artifact Revision MAY depend upon one or more Requirements or Requirement Artifact Revisions.

Each Dependency relationship SHALL connect exactly one dependent source participant to exactly one dependency target participant.

Where one participant depends upon multiple other participants, each dependency SHALL be represented through a separate Dependency relationship.

Dependency records that the interpretation, applicability, authority, realization, or evaluation of one Requirement relies upon another identified Requirement or represented required engineering intent.

Every Dependency relationship SHALL identify:

- the dependent participant;
- one dependency participant;
- whether each participant is identified at Requirement identity level or Requirement Artifact Revision level;
- the nature of the engineering reliance;
- the applicable scope where the reliance is not universal.

Dependency SHALL remain explicitly represented.

Dependency SHALL NOT imply:

- Derivation;
- Refinement;
- Decomposition;
- Supersession;
- Allocation;
- implementation order.

Dependency SHALL NOT by itself transfer:

- Requirement Authority;
- Requirement Applicability;
- Lifecycle State;
- Satisfaction;
- identity;
- content.

### 21.1 Dependency Relation Type

For Dependency:

- the source participant SHALL be the dependent Requirement or Requirement Artifact Revision;
- the target participant SHALL be the Requirement or Requirement Artifact Revision upon which the source depends;
- direction SHALL be from dependent to dependency;
- each Dependency relationship SHALL have exactly one dependent source participant and exactly one dependency target participant;
- one dependent participant MAY depend upon one or more dependency participants;
- one dependency participant MAY be depended upon by multiple dependent participants;
- Dependency cycles MAY be represented;
- each participant SHALL explicitly identify whether it refers to Requirement identity level or Requirement Artifact Revision level.

Where a Dependency cycle exists, every participating Dependency relationship SHALL remain explicitly represented.

The existence of a Dependency cycle SHALL NOT by itself establish that the participating Requirements are invalid, unsatisfiable, or non-conforming.

An applicable PEOS Product contract MAY prohibit specified Dependency cycles or require their engineering resolution.

## 22. Conflict

A Requirement or Requirement Artifact Revision MAY participate in one or more explicitly represented binary Conflict relationships.

Each Conflict relationship connects exactly two distinct participants.

A Conflict relationship records that the required engineering intent of its two identified participants cannot be assumed to be simultaneously satisfied, implemented, or applied within a defined scope without engineering resolution.

Every Conflict relationship SHALL identify:

- one source participant;
- one target participant;
- each participating Requirement;
- each applicable Requirement Artifact Revision where the Conflict concerns represented engineering intent;
- the scope in which the Conflict exists;
- the nature of the incompatibility.

Conflict SHALL NOT be inferred solely from different wording, different classifications, different authorities, or different Lifecycle States.

Conflict SHALL NOT automatically:

- invalidate a Requirement;
- modify Requirement content;
- modify Requirement Applicability;
- modify Requirement Authority;
- withdraw a Requirement;
- reject a Requirement;
- establish Supersession.

Conflict resolution SHALL occur through applicable engineering governance.

A Decision Outcome MAY establish the selected resolution.

Resolution of a Conflict SHALL NOT remove the historical fact that the Conflict was represented unless applicable retention rules permit removal of that engineering information.

### 22.1 Conflict Relation Type

For Conflict:

- each Conflict relationship SHALL have exactly one source participant and exactly one target participant;
- source and target SHALL be distinct;
- Conflict engineering semantics SHALL be symmetric between source and target;
- one participant MAY participate in multiple Conflict relationships;
- a conflict involving more than two participants SHALL be represented through multiple binary Conflict relationships;
- Conflict cycles among three or more distinct participants MAY be represented;
- each participant SHALL explicitly identify whether it refers to Requirement identity level or Requirement Artifact Revision level.

The ordering of Conflict participants SHALL NOT imply priority, authority, preference, or resolution direction.

Where an implementation requires ordered source and target fields, that ordering SHALL be representational only and SHALL NOT alter the symmetric engineering semantics of Conflict.

A collection of binary Conflict relationships MAY collectively represent a broader multi-participant incompatibility.

Such a collection SHALL NOT create a separate Conflict entity or relationship-group entity.

Absence of a binary Conflict relationship between two participants SHALL NOT be inferred solely from both participants appearing elsewhere in the same broader incompatibility.

## 23. Supersession

A Requirement MAY supersede one or more Requirements through one or more explicitly represented Supersession relationships.

Each superseding-to-superseded association SHALL be represented through a separate binary Supersession relationship.

Each Supersession relationship establishes that the required engineering intent represented by an identified superseding Requirement Artifact Revision replaces the required engineering intent represented by one identified superseded Requirement Artifact Revision within a defined scope.

For the purposes of this specification, a governance action is an established Decision Outcome or another governance mechanism explicitly permitted by an applicable PEOS Product contract.

Governance action is a semantic role and SHALL NOT be interpreted as introducing a separate PEOS entity.

Every Supersession relationship SHALL identify:

- one superseding Requirement;
- one superseding Requirement Artifact Revision;
- one superseded Requirement;
- one superseded Requirement Artifact Revision;
- the scope within which that replacement applies;
- the governance action under which that replacement was established.

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

### 23.1 Supersession Relation Type

For Requirement Supersession:

- the source participant SHALL be the superseding Requirement Artifact Revision;
- the target participant SHALL be a superseded Requirement Artifact Revision;
- direction SHALL be from superseding to superseded;
- each Supersession relationship SHALL have exactly one superseding source Requirement Artifact Revision and exactly one superseded target Requirement Artifact Revision;
- one superseding Requirement Artifact Revision MAY be the source of multiple Supersession relationships;
- one Requirement Artifact Revision MAY be the target of more than one Supersession relationship only where the applicable scopes remain explicitly distinguishable and non-contradictory;
- Supersession cycles SHALL NOT be permitted;
- source and target SHALL be identified at Requirement Artifact Revision level while their owning Requirement identities remain identifiable.

A Requirement Artifact Revision SHALL NOT directly or transitively supersede itself.

Each Supersession relationship SHALL identify the resulting effective status or Lifecycle consequence, if any, for its superseded target within the relation scope.

Absence of a Lifecycle consequence SHALL be explicitly representable and SHALL NOT invalidate an otherwise established Supersession relationship.

### 23.2 Superseded Requirement

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

In this specification, a State Assignment refers to a lifecycle State Assignment governed by PEOS-003.

State Assignments SHALL remain separate from Artifact Revisions.

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

A State Assignment SHALL NOT by itself establish:

- which Requirement supersedes another Requirement;
- which Requirement Artifact Revisions are involved;
- the scope of replacement;
- the authority or governance action establishing replacement.

Normative replacement of required engineering intent SHALL be represented through an explicit Supersession relationship as defined by Section 23.

The State Assignment and the applicable Supersession relationship SHALL remain semantically distinct and independently inspectable.

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

### 28.4 Conflict Resolution

A Decision Outcome MAY establish an engineering resolution for an explicitly represented Requirement Conflict.

Such an Outcome MAY:

- select one Requirement over another within a defined scope;
- authorize Requirement change;
- authorize Withdrawal;
- authorize Supersession;
- establish a Waiver;
- require additional Requirements.

The Decision Outcome SHALL NOT silently rewrite or delete any Requirement or Requirement Artifact Revision participating in the Conflict.

Any resulting Requirement change, Lifecycle transition, Supersession relationship, or Waiver SHALL remain explicitly represented through its applicable PEOS semantics.

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

Historical State Assignments SHALL remain distinguishable.

The existence, participants, direction, type, scope, and engineering meaning of explicitly represented historical Requirement relationships SHALL remain inspectable where those relationships are retained as engineering information.

This requirement does not define independent revision, lifecycle, or temporal-history semantics for relationships.

Such semantics remain outside the scope of this specification as defined by Section 17.1.

Historical Decisions referencing Requirements SHALL remain inspectable.

Historical Engineering Commitments SHALL remain inspectable.

Historical Validation Records SHALL be governed by the future Validation Model.

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

A State Assignment SHALL NOT by itself establish Supersession.

Supersession SHALL identify the Requirement Artifact Revisions whose required engineering intent is replaced and the scope in which replacement applies.

Supersession SHALL NOT merge or destroy Requirement identities.

**Relationship Participants**

A Requirement relationship SHALL NOT leave ambiguous whether a participant identifies a Requirement identity or a Requirement Artifact Revision.

A relationship whose meaning depends upon represented required engineering intent SHALL identify the applicable Requirement Artifact Revision.

**Derivation**

Derivation SHALL record engineering origin and SHALL NOT be treated as synonymous with Refinement.

**Refinement**

Refinement SHALL preserve compatibility with refined required engineering intent within its defined scope.

**Decomposition**

Decomposition SHALL NOT imply completeness or Satisfaction aggregation unless those semantics are explicitly represented through applicable models.

**Dependency**

Dependency SHALL NOT transfer identity, content, Authority, Applicability, Lifecycle State, or Satisfaction.

**Conflict**

Conflict SHALL NOT by itself modify, reject, withdraw, waive, or supersede a Requirement.

Conflict representation SHALL remain binary even where multiple Conflict relationships collectively represent broader incompatibility.

Binary representation SHALL NOT alter the symmetric engineering semantics of Conflict.

**Artifact Relation Conformance**

Every Requirement relationship SHALL conform to the Artifact Relation contract established by PEOS-002.

Every Requirement relationship SHALL identify its Relation Type, provenance, participants at the required identity or Artifact Revision level, and applicable scope.

Every Requirement Relation Type SHALL define direction, multiplicity, cycle policy, participant level, and integrity constraints.

Every individual Requirement relationship SHALL contain exactly one source participant and exactly one target participant.

Engineering semantics involving multiple sources, multiple targets, or more than two participants SHALL be represented through multiple binary Requirement relationships.

A collectively interpreted group of Requirement relationships SHALL NOT create a separate PEOS entity.

**Relationship Cycles**

Derivation, Refinement, Decomposition, and Supersession cycles SHALL NOT be permitted.

Dependency and multi-participant Conflict cycles MAY be represented subject to applicable Product contract constraints.

**Relationship Provenance**

Relationship provenance SHALL remain distinct from Requirement Origin, Requirement Authority, Artifact Revision provenance, and participant authorship.

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

### 36.14 Ambiguous Relationship Participant

Representing a Requirement relationship without identifying whether a participant refers to the Requirement identity or to a specific Requirement Artifact Revision.

---

### 36.15 Derivation as Refinement

Treating every derived Requirement as necessarily being a more precise or narrower expression of its source Requirement.

---

### 36.16 Implicit Decomposition Completeness

Assuming that a set of subordinate Requirements completely covers the parent required engineering intent solely because Decomposition relationships exist.

---

### 36.17 Dependency Property Transfer

Treating a Dependency relationship as transferring Requirement Authority, Applicability, Lifecycle State, content, identity, or Satisfaction between participants.

---

### 36.18 Conflict as Automatic Invalidity

Automatically rejecting, withdrawing, waiving, or superseding a Requirement solely because a Conflict relationship exists.

### 36.19 Missing Relationship Provenance

Representing a Requirement relationship without identifying the provenance of the relationship fact.

---

### 36.20 Undefined Relationship Cycle Policy

Defining or using a Requirement Relation Type without an explicit rule stating whether cycles are permitted.

---

### 36.21 Prohibited Structural Cycle

Representing a Derivation, Refinement, Decomposition, or Supersession cycle.

---

### 36.22 Subject Cardinality Ambiguity

Identifying multiple Requirement Subjects without representing whether the required engineering intent applies independently, collectively, or according to another explicit relationship among those Subjects.

### 36.23 Multi-Source Artifact Relation

Representing one Requirement relationship with more than one source participant.

---

### 36.24 Multi-Target Artifact Relation

Representing one Requirement relationship with more than one target participant.

---

### 36.25 Reified Relationship Collection

Treating a collection of related Requirement relationships as a separate Relationship Set, Conflict, Decomposition Set, Supersession Set, hyperedge, or other PEOS entity without a normative Relationship Model defining that entity.

---

### 36.26 Non-Binary Conflict Relation

Representing one Conflict relationship with more than two participants instead of multiple binary Conflict relationships.

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
