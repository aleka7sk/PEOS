# Changelog

All notable changes to the PEOS Go SDK are recorded here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/); versioning follows
[Semantic Versioning](https://semver.org/), as detailed in README.md's "Release contract"
section.

This is the project's first tagged release. The entry below describes the v1.0.0
baseline — what a consumer gets on day one — not a migration from any prior published
version, since none exists.

## [1.0.0] — 2026-07-28

### Added

- Implementation of PEOS-000 through PEOS-009, the ten normative specifications
  describing the Product Engineering OS artifact, lifecycle, decision, requirement,
  validation, quality, runtime contract, and template contract models.
- A nine-package Go SDK (`peos/core`, `peos/relation`, `peos/lifecycle`, `peos/decision`,
  `peos/requirement`, `peos/validation`, `peos/quality`, `peos/runtime`, `peos/template`)
  implementing those specifications as immutable, self-validating value types.
- An immutable Artifact and Artifact Revision model (`core.Artifact`,
  `core.ArtifactRevision`): stable identity, immutable recorded state, mandatory
  provenance, and a typed reference layer (29 `*Ref` types, 16 identity types) that every
  cross-package citation travels through.
- Lifecycle support: Lifecycle Definition, Definition Version, State Assignment,
  Transition Record, and Lifecycle Definition Version Supersession.
- Decision support: Decision, Authority, Basis, Alternative, Outcome, Commitment, Record,
  Supersession, Invalidation, Conflict, and Conflict Resolution.
- Requirement support: Requirement, Revision, and six relationship wrappers (Derivation,
  Refinement, Decomposition, Dependency, Conflict, Supersession) plus Waiver.
- Validation support: Validation Plan, Planned Activity, Execution Record, and Validation
  Claim (covering Satisfaction and Conformance Claim Types).
- Quality support: Quality Profile (Characteristic, Measure, Threshold, Target,
  Constraint, Normalization/Aggregation Rules), Measurement Record, and Quality Claim.
- Runtime Contract support: Contract, Assertion, Contract Rule, Binding Record,
  Unbinding Record, Observation, Violation, and Compliance Claim.
- Template support: Template, Template Revision, Parameter/ParameterType/
  ParameterConstraint, Application Record, Generated-From/Composition/Specialization
  relations, and Template Conformance Claim.
- Zero third-party dependencies — the SDK imports only the Go standard library.
- Architecture boundary tests in every package (`doc_test.go`), enforcing the package
  dependency graph in both directions via AST-based import parsing, so a future change
  cannot silently introduce a forbidden or reversed dependency.
- Public JSON support on every type: stable union discriminators, uniform explicit-null
  rejection for optional fields, uniform unknown-field tolerance, and receiver
  preservation on failed decode.
- Product extension support (`core.Extension`) on essentially every type, as the
  sanctioned escape hatch for Product-specific data that is not itself a PEOS concept.
- Nine compiling package-level examples (`Example_*` functions) plus one cross-package
  workflow example (`examples/crosspackage`) demonstrating how a Requirement composes
  with Validation through `core.EngineeringSubjectRef` and `core.CriterionRef` without
  either package importing the other.
- `docs/consumer-guide.md`, a bounded guide for a new external consumer.
- Ten documented deferrals — specification concepts deliberately not implemented, each
  recorded with its normative source and the question that must be resolved first (see
  README.md's "Deferrals" section and `docs/implementation-progress.md`'s "Deferred
  Architecture" section).

### Pre-release corrections

These changes were made during architecture review, before this first tag was created.
They are part of the v1.0.0 baseline, not migrations from a previously released version —
no prior version of this SDK was ever published.

- The Transition Record source field changed from `from_state` (a `StateID`) to
  `source_state_assignment` (a `core.StateAssignmentRef`), because PEOS-003 requires a
  Transition Record to identify "the source State Assignment," which a State identifier
  alone cannot express — a Subject may occupy the same State more than once over its
  history.
- A succeeded (completed) Transition Record now requires both a responsible Actor
  (carried by the enclosing Revision's Provenance) and an identifiable authority basis
  (carried explicitly on the Transition Record's content), per PEOS-003's Actor
  obligation and Authority Invariant. Neither is duplicated into the other; both are
  distinct concepts.

This entry describes the v1.0.0 release-finalization commit and its annotated tag target;
see `docs/implementation-progress.md` for the current, exact repository state.
