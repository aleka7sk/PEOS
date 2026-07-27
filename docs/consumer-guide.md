# PEOS Go SDK Consumer Guide

This guide explains the SDK to a new external consumer without requiring you to read all
ten PEOS specifications first. It is a map, not a substitute — where this guide and a
specification disagree, the specification governs, and where this guide and
`docs/implementation-progress.md` disagree, the tracker governs. See README.md's "Release
contract" section for what is stable and what may still change.

## 1. What PEOS models

**The specifications are normative; the SDK is one conformant implementation.** PEOS-000
through PEOS-009 describe how software product engineering state is modeled —
independently of any programming language. This Go SDK implements that model as a set of
immutable value types. A different implementation, in a different language, could conform
to the same specifications without looking anything like this code.

**Engineering state is immutable.** Nothing in this SDK represents "the current state of
X" as a field you can overwrite. Instead, every meaningful state is a new immutable value,
and history accumulates rather than being replaced.

**An Artifact is stable identity; an Artifact Revision is one immutable recorded state of
it.** `core.Artifact` never changes once constructed except through its own limited
modifiers (roles, scope, extension — never identity or type). `core.ArtifactRevision` is
recorded once and never mutated afterward. When something needs a new state, you construct
a new Revision; the old one remains, permanently, as history.

**Not everything immutable is an Artifact.** PEOS-006 through PEOS-009 each define their
own non-Artifact immutable records — a Validation Execution Record, a Validation Claim, a
Runtime Observation, a Template Application Record. These carry their own dedicated
identity type (for example `core.ValidationClaimID`) rather than an `ArtifactID`, and they
are never revisioned: each one is recorded once, and correcting it means recording a new
one that references the old, not editing the old in place.

**References, not embedded values, connect things.** When one construct needs to point at
another, it holds a typed reference (a `core.*Ref` value) — never the other construct's
full value inline. This is what lets `peos/validation` cite a Requirement without importing
`peos/requirement` at all (see section 4).

**Repositories, not PEOS values, own persistence and derivation.** The SDK is a value
model. It has no database, no in-memory store, and no query language. Anything that
requires looking across many values — "what is the current Revision," "what is the
compliance status right now," "does this graph have a cycle" — is explicitly a consumer
repository's job, not a PEOS type's job. Section 6 goes into this in detail.

**Products own everything domain-specific the specifications don't define.** `core.Scope`,
`core.VocabularyValue`, and `core.Extension` exist so a Product can extend a controlled
vocabulary or attach its own data, without inventing a parallel field the SDK didn't
anticipate.

**What PEOS deliberately does not execute:** no execution engine, no workflow
orchestrator, no Guard/Effect expression evaluator, no template renderer, no migration
tool, no CLI. These are Product concerns. Ten specific deferrals — including these — are
listed in section 10.

## 2. Package map

```
core         → (no PEOS imports — the foundation)
relation     → core
lifecycle    → core
decision     → core
validation   → core
requirement  → core, relation
quality      → core, validation
runtime      → core, validation
template     → core, relation, validation
```

Every package's dependency rule (including the *absence* of a dependency, in both
directions) is enforced by an automated test in that package's `doc_test.go` — this is not
merely a convention, it is checked on every `go test ./...`.

| Package | Owns |
|---|---|
| `core` | Identity, references, vocabularies, provenance, the Artifact/Revision pair, and the union types (`EngineeringSubjectRef`, `CriterionRef`, `RecordRef`) every other package composes through |
| `relation` | The generic PEOS-002 Artifact Relation contract — a plain, immutable, directed assertion between two subjects |
| `lifecycle` | PEOS-003: Lifecycle Definition, Definition Version, State Assignment, Transition Record |
| `decision` | PEOS-004: Decision, Authority, Basis, Outcome, Record, and Decision-family relationships |
| `requirement` | PEOS-005: Requirement, Revision, and six relationship wrappers over `relation.Relation` |
| `validation` | PEOS-006: Validation Plan, Execution Record, Claim (Satisfaction and Conformance) |
| `quality` | PEOS-007: Quality Profile, Measurement Record (composing an Execution Record), Quality Claim |
| `runtime` | PEOS-008: Runtime Contract, Binding/Unbinding Record, Observation, Violation |
| `template` | PEOS-009: Template, Application Record, Generated-From/Composition/Specialization, Conformance Claim |

`peos/core` is the foundation: it imports no PEOS package, and every other package depends
on it directly or transitively. Nothing imports `peos/template` — it is a leaf.

## 3. Artifact and Revision fundamentals

**Identity** is `core.ArtifactID` plus `core.ArtifactType`, held by `core.Artifact`. It
never changes for the life of the Artifact.

**Immutable revisions.** `core.ArtifactRevision` carries `ArtifactID`, its own
`ArtifactRevisionID`, `Origin`, `Provenance`, `IntegrityIdentity`, and `Representations`.
Once constructed, none of these fields can change — there are no setters, only a handful
of `With*` modifiers on the mutable-looking-but-actually-copy-returning fields (like
`Extension`), and even those return a new value rather than mutating the receiver.

**Provenance** (`core.Provenance`) records who or what produced a Revision and when — an
optional `ActorRef`, an optional recorded timestamp, an optional source/method vocabulary
value. It is mandatory on every `ArtifactRevision`.

**Current-revision resolution is not the SDK's job.** PEOS does not define a universal
rule for picking "the current Revision" out of an Artifact's history — this is
explicitly repository- and Product-owned, because the right rule (latest timestamp?
latest accepted-status Revision? something scope-specific?) varies by Product. The SDK
gives you every Revision it is told about; deciding which one is "current" is yours.

**Correction, not rewriting.** When PEOS defines a correction path for a record (six
families do: `validation.Claim`, `validation.ExecutionRecord`, `quality.Claim`,
`quality.MeasurementRecord`, `runtime.BindingRecord`, `runtime.UnbindingRecord`,
`template.ApplicationRecord`), it is expressed as `core.RecordCorrectionRef[T]` — a typed
reference the *new* record carries, naming the record it corrects, replaces, or
invalidates. The old record is never edited or deleted; it stays exactly as it was, for
history's sake.

## 4. Reference model

**Typed `Ref` values are how anything points at anything else.** There are 29 of them in
`peos/core` (`ArtifactRef`, `ArtifactRevisionRef`, `RequirementRef`,
`ValidationClaimRef`, and so on). Each is a small, immutable, comparable value — nothing
more than the identity information needed to name the thing it refers to.

**Artifact-level versus Revision-level.** Many entities have both an Artifact-level
reference (naming the logical thing, across all its history) and a Revision-level
reference (naming one exact recorded state of it) — `ArtifactRef` versus
`ArtifactRevisionRef`, `RequirementRef` versus `RequirementArtifactRevisionRef`, and so on.
Whether a given construct requires the Artifact level, the Revision level, or accepts
either (via a participant union, like `template.TemplateParticipant`) is a normative
choice each specification makes explicitly — not a convenience decision left to the SDK.

**`core.EngineeringSubjectRef`** is the closed union used wherever PEOS lets something be
"about" a broad range of engineering subjects — an Artifact, an Artifact Revision, a
Requirement, a Decision Outcome, a Runtime Subject, and more (17 arms in total). This is
how, for example, a Decision names its subjects, or an Execution Record names what it
exercised.

**`core.CriterionRef`** is the closed union used specifically for Claim criteria — what a
Claim is *evaluated against*, as distinct from what it is *about* (its Subject). It has 14
arms, including a Requirement, a Quality Characteristic, a Runtime Contract rule, and a
Template constraint. `CriterionRef` is deliberately not `EngineeringSubjectRef`: keeping
them separate is what makes it structurally impossible for a Claim's criteria to smuggle
in a second Subject.

**`core.RecordRef`** is the dynamic 8-arm union over every immutable record family's
identity, for a storage or serialization layer that needs to decide the record family at
runtime rather than at compile time — the type parameter of `core.RecordCorrectionRef[T]`
is the compile-time-fixed alternative to this.

**Local-key references.** Several owned-value elements — a Quality Characteristic, a
Runtime Assertion, a Template Parameter Constraint — are named by pairing their owning
Revision's reference with a `core.LocalKey`, rather than being given independent identity.
This is how, for example, `core.QualityElementCriterionRef` cites a Characteristic that
lives inside one `ProfileContent`.

**Why cross-package composition goes through `core`.** The package graph in section 2 is
deliberately acyclic and one-directional — `peos/validation` cannot import
`peos/requirement`, `peos/quality` cannot import `peos/runtime`, and so on. Whenever two
domain packages need to refer to each other's concepts, the reference travels through one
of `core`'s union types instead of a direct type dependency. Section 7 walks through the
canonical example of this — Requirement to Validation — in full.

## 5. Construction conventions

**Mandatory fields are constructor arguments; optional fields are `With*` modifiers.**
Every constructor across all nine packages follows this rule without exception. If a field
participates in a cross-field invariant (for example, PEOS-009's rule that a succeeded
Template Application must name at least one generated output), it is a constructor
argument too — never modifier-only — so that no valid state is ever unconstructible
through the public API.

**Modifiers return a new value.** `x.WithScope(s)` never mutates `x`; it returns a new
value with the scope set, and you must capture the return value (`x, err =
x.WithScope(s)`). Nothing in this SDK exposes a pointer receiver for a value-semantics
type.

**Zero values are rejected wherever a valid value is required.** A zero `core.ArtifactID`,
a zero `core.Timestamp`, an empty required string — all produce an error, from the
constructor or the modifier that received them, immediately at the call site.

**Error matching via `errors.Is`, with nested sentinels.** Every construct has its own
top-level sentinel (for example `ErrInvalidTransitionRecordRevision`), and a specific cause
is wrapped inside it (for example `ErrMissingResponsibleActor`) rather than replacing it.
Both match with `errors.Is`:

```go
_, err := lifecycle.NewTransitionRecordRevision(record, revision, content)
if errors.Is(err, lifecycle.ErrMissingResponsibleActor) { /* the specific cause */ }
if errors.Is(err, lifecycle.ErrInvalidTransitionRecordRevision) { /* the general sentinel */ }
```

**JSON decoding preserves the receiver on failure and rejects explicit null for optional
fields consistently.** If `json.Unmarshal(data, &v)` returns an error, `v` is left exactly
as it was before the call — never partially overwritten. An explicit JSON `null` for an
optional field is treated as malformed input, not as "absent"; an absent key is what means
"absent." Unknown JSON fields are always tolerated, which is the SDK's one deliberate
forward-compatibility default (see README's JSON compatibility policy).

## 6. Repository boundary

**What your repository must provide, that PEOS deliberately does not:**

- **Persistence.** The SDK never writes anything to disk, a database, or any store. Every
  value is in-memory only until you serialize and persist it yourself.
- **Current-revision selection.** Given an Artifact's full Revision history, deciding
  which one is "current" — by timestamp, by status, by scope — is your rule to write.
- **Graph traversal and cycle detection.** `peos/relation` holds a single immutable
  `Relation` value type — no relation set, no repository, no traversal. PEOS-002 assigns
  cross-relation analysis to "a future Traceability Model."
- **Current lifecycle state.** `lifecycle.StateAssignment` gives you every assignment
  ever made; resolving "the current State" from that history (tie-breaking on
  concurrent assignments, composing across scopes) is repository policy.
- **Lookup and aggregation.** "All Claims for this Subject," "the most recent,
  non-invalidated Claim for this scope," "every Violation tied to this Binding" — all of
  these are queries over a set of PEOS values, and the SDK holds no such set.
- **Indexing and query models.** However you choose to index Artifacts, Revisions, and
  records for fast lookup is entirely up to your storage layer.

**What must never be written back into a PEOS value as a derived field:** `status`,
`current`, `compatible`, `conformant`, `satisfied`, or any field expressing a computed
verdict. Every one of these is a *derived view* over immutable records, by specification
design — PEOS-006 states directly, for example, that "There is no separate Verdict
entity." If you find yourself wanting to add a field like this to a PEOS type, that
computation belongs in your repository or aggregation layer, as its own value, not
patched onto the SDK's types.

A minimal conceptual repository interface, sketched (not code the SDK provides):

- **Artifacts and Revisions:** `PutRevision(rev)`, `RevisionsOf(artifactID) []Revision`,
  `Revision(ref ArtifactRevisionRef) (Revision, bool)`.
- **Immutable records:** `Put(record)`, `Get(ref) (record, bool)`, keyed by each family's
  own identity type.
- **Relations:** `PutRelation(rel)`, `RelationsFrom(subject)`, `RelationsTo(subject)`.
- **Lifecycle history:** `AssignmentsFor(subject) []StateAssignment`,
  `TransitionsFor(subject) []TransitionRecordRevision`.
- **Claims and Evidence:** `ClaimsFor(subject, scope) []Claim`,
  `CurrentClaim(subject, scope, criteria) (Claim, bool)` — the derived-view query PEOS-006
  explicitly assigns to you.
- **Template applications:** `ApplicationsOf(templateRevisionRef) []ApplicationRecord`.
- **Runtime bindings and observations:** `ActiveBindings(subject) []BindingRecord`,
  `ObservationsFor(binding) []Observation`.

## 7. Cross-package workflow

The complete conceptual chain: **Requirement → Validation Plan → Execution Record →
Evidence → Result → Validation Claim.**

This is the SDK's most important composition pattern, because it crosses the one boundary
that is architecturally enforced in both directions: `peos/validation` cannot import
`peos/requirement`, and `peos/requirement` cannot import `peos/validation`. A
`requirement.Requirement` Go value never appears inside `peos/validation` — not as a
field, not as a parameter, not anywhere.

Instead, a Requirement becomes one of two `core` reference types, depending on what role
it plays:

- **As a `core.EngineeringSubjectRef`**, via `core.EngineeringSubjectRefFromRequirement`,
  when the Requirement itself is the *thing being evaluated* — for example, a Decision or
  a Runtime Contract naming the Requirement as its subject.
- **As a `core.CriterionRef`**, via `core.CriterionRefFromRequirement`, when the
  Requirement is *what something else is evaluated against* — the shape used for a
  Satisfaction Claim, where PEOS-006 requires the Claim's criteria to identify the
  Requirement being satisfied, while the Claim's Subject is the Artifact or Revision doing
  the satisfying.

Both conversions happen in your code — the consumer is the one place allowed to import
both packages — never inside either package itself. **No shadow struct is required**: you
never need to redeclare a `Requirement`-shaped type inside a validation context. The
`core.RequirementRef` you already have is exactly what `validation.Claim`'s criteria slice
accepts, because it is a `core.CriterionRef` arm.

**Repository lookup stays outside this chain entirely.** Nothing above resolves a
`core.RequirementRef` back into a full `requirement.Requirement`, and nothing needs to —
the Claim only needs the reference, not the referenced value. If you later need to display
"which Requirement does this Claim satisfy," that lookup is your repository's job (section
6), not something the Claim carries.

A fully compiling version of this exact example lives at
`examples/crosspackage/workflow_example_test.go` and runs under `go test ./...`.

## 8. Lifecycle workflow

PEOS-003's Transition Record is where the SDK's construction conventions and its
composition-over-duplication discipline meet most visibly.

- **State Assignment** (`lifecycle.StateAssignment`) records that a Lifecycle Subject was
  assigned a given `StateID` under a given Definition Version, at a given time, under a
  given Provenance. It has its own identity (`core.StateAssignmentID`) and is never
  revisioned.
- **Transition Record Artifact** (`lifecycle.TransitionRecord`) is the one exception to
  the "records aren't Artifacts" default in this package: PEOS-003 explicitly calls it "a
  persistent Artifact." It composes `core.Artifact` the same way `requirement.Requirement`
  does.
- **Transition Record Revision** (`lifecycle.TransitionRecordRevision`) pairs a
  `core.ArtifactRevision` with typed `TransitionRecordContent`.
- **Source State Assignment.** `TransitionRecordContent.FromAssignment()` is a
  `core.StateAssignmentRef` — not a bare State name — because PEOS-003 requires "the
  source State Assignment" specifically, and a Subject may occupy the same State more
  than once over its history.
- **Target State** (`ToState()`, a `StateID`) names the State the Transition targeted;
  **resulting State Assignment** (`ResultingAssignment()`, a `core.StateAssignmentRef`)
  is the Assignment record that establishes it. These are two different things for a
  reason: one names a State, the other names the specific historical record of reaching
  it.
- **Responsible Actor** lives in the enclosing `core.ArtifactRevision`'s `Provenance`, not
  on `TransitionRecordContent` — because a Transition Record Revision's Provenance Actor
  *is* PEOS-003's "responsible Actor or Runtime," and duplicating it into content would
  create two sources of truth for the same fact.
- **Authority** is a separate concept from Actor ("Actor identity and transition
  authority are distinct," PEOS-003), stored explicitly on `TransitionRecordContent` via
  `Authority()`/`WithAuthority()`.
- **Succeeded versus non-succeeded outcomes.** A succeeded Transition Record is a
  completed Transition under PEOS-003's Target State Invariant, and both the Actor and
  the Authority obligations apply to it specifically. Failed, interrupted, indeterminate,
  and Product-declared outcomes are Transition Attempts, not completed Transitions, and
  neither obligation is imposed on them.

A fully compiling version of this workflow lives at
`peos/lifecycle/example_test.go`.

## 9. Correction workflows

Six record families carry a documented correction path:
`validation.ExecutionRecord`, `validation.Claim`, `quality.MeasurementRecord`,
`quality.Claim`, `runtime.BindingRecord`, `runtime.UnbindingRecord`, and
`template.ApplicationRecord`.

Each exposes `WithCorrection(core.RecordCorrectionRef[T])` and `WithoutCorrection()`. The
correction reference names three things: which kind of correction it is
(`core.CorrectionKind` — correct, replace, or invalidate, per PEOS-006's "Claim
Correction, Replacement, and Invalidation" section), and which earlier record of the same
family it points at.

The earlier record is **never mutated, deleted, or rewritten.** Recording a correction
means constructing a *new* record of the same family, with `WithCorrection` naming the one
it corrects. Both records — the original and its correction — remain permanently
inspectable. Determining "the currently applicable one" for a given subject and scope is,
again, a repository-owned derived view (section 6), never a stored field.

Two record families — `runtime.Observation` and `runtime.Violation` — deliberately carry
**no** correction reference at all, because PEOS-008's own sections defining them state no
correction, replacement, or invalidation mechanism for either.

## 10. Deferred capabilities

PEOS defines these concepts; this SDK deliberately does not implement them. Each is
recorded in `docs/implementation-progress.md`'s "Deferred Architecture" section with its
exact normative source and the open question blocking it.

| Deferral | Owning layer once resolved |
|---|---|
| Specialized Artifact Supersession enforcement (mandatory scope, self-supersession rejection) | `peos/relation` or a future sibling package — an owning-package decision is still open |
| Cross-relation graph traversal, cycle detection, traceability coverage | A future Traceability Model (PEOS-002's own words) |
| Guard / Effect / Trigger expression language | A future `peos/lifecycle` packet, once PEOS-003 or a Product contract defines a stable expression language |
| Lifecycle Migration execution | A future Runtime-layer packet |
| Current State / State History resolution | Your repository, by design — not a future SDK packet |
| Delegation (authority-grant records) | `peos/core` or a new sibling package — an owning-package decision is still open |
| Runtime Waiver attached conditions and criterion-level subject | Blocked on PEOS-005/PEOS-006 defining the missing concepts first |
| Template Migration | Blocked on PEOS-009 (or an accepted architecture decision) assigning Migration an ontology — it currently states what a migration must identify, never what a Migration *is* |

None of these block a v1 release scenario: every workflow this guide walks through is
fully representable without them. See README.md's "Deferrals" section for the short
version.

## 11. Compatibility expectations

This SDK targets a v1.0.0 release candidate. README.md's "Release contract" section is the
authoritative statement of what is public and stable, what semantic versioning means for
this project, and the JSON compatibility policy. In summary: the public v1 surface
(package paths, exported types/constructors/methods/sentinels, vocabulary constants, JSON
field names and discriminators, and extension semantics) is stable within a major version;
breaking changes require a new major version; additive APIs are fine in a minor version.
CHANGELOG.md records what changed and when, starting from this v1.0.0 baseline.

**The specifications under `spec/` remain authoritative.** This guide, and this SDK, exist
to make them checkable by a compiler — not to replace them. Where this guide simplifies or
summarizes, the specification text is the final word.
