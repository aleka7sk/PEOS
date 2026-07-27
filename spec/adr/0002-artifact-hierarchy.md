# ADR 0002 — Artifact / Immutable Revision Hierarchy

- **Status:** Accepted
- **Record created:** 2026-07-27 (Packet L.0.C), reconstructing a decision already in force.
  The file existed but was empty; this record states only what PEOS-002 mandates and what
  the SDK demonstrably implements.

## Context

Every PEOS domain needs to answer the same question: what does it mean for engineering
content to change? Two shapes are available — a mutable record updated in place, or a
stable identity with an append-only series of immutable recorded states.

PEOS-002 settles it. An Artifact carries stable identity; an Artifact Revision is an
immutable recorded state of that Artifact. PEOS-002's Immutability Invariant, History
Preservation Invariant, and Revision Identity Invariant all depend on that split, and its
Artifact Supersession section states explicitly that "Creation of a newer Artifact Revision
SHALL NOT automatically imply that an earlier Artifact Revision is superseded, invalid,
withdrawn, or non-applicable."

## Decision

**Identity and recorded state are separate types throughout the SDK.**
`core.Artifact` holds `ArtifactID` and `ArtifactType`. `core.ArtifactRevision` holds the
artifact ID, its own revision ID, origin, provenance, integrity identity, and
representations. Neither exposes a setter.

**Every specialized domain that PEOS calls an Artifact follows the same pair.**
`requirement.Requirement`/`Revision`, `validation.Plan`/`PlanRevision`,
`quality.Profile`/`ProfileRevision`, `runtime.Contract`/`ContractRevision`,
`template.Template`/`TemplateRevision`, and `lifecycle.TransitionRecord`/
`TransitionRecordRevision` each compose the core pair by named field and pair it with a
typed `*Content` value carrying the domain's own fields.

**Constructs PEOS does not call Artifacts are not forced into the hierarchy.**
State Assignments, Validation Execution Records and Claims, Measurement Records, Runtime
Binding/Unbinding/Observation/Violation records, and Template Application Records are
immutable, independently identified, non-revisioned values. They carry their own identity
types rather than an `ArtifactID`, because PEOS-006 through PEOS-009 explicitly call them
records rather than Artifacts.

**Amendment never rewrites.** Records that PEOS gives a correction path use
`core.RecordCorrectionRef[T]` to point at the record they correct, replace, or invalidate.
No type in the SDK mutates a recorded value.

## Consequences

- Historical state is preserved structurally rather than by convention. There is no code
  path that edits a Revision, so History Preservation cannot be violated by accident.
- Content is owned by the Revision, not the Artifact. A Revision-owned `*Content` value is
  the authoritative content for that recorded state, which is why compatibility,
  applicability, and parameter declarations all live on the Revision side.
- "Which Revision is current" is not answerable by the SDK. PEOS assigns no universal rule
  for selecting one record from a history, so derived-state resolution is deliberately out
  of scope — see the Deferred Architecture section of
  `docs/implementation-progress.md`.
- The Artifact Representation contract stays on `core.ArtifactRevision`. This is why, for
  example, a Template's body is not a field of `template.TemplateContent`.

## Alternatives considered

- **A single mutable Artifact type with a version counter.** Rejected: it makes the
  Immutability and History Preservation Invariants unenforceable, and PEOS-002 states that a
  newer Revision must not by itself imply anything about an earlier one.
- **Forcing every immutable record into the Artifact/Revision pair.** Rejected: PEOS-006
  through PEOS-009 describe their records as non-Artifact and non-revisioned, and giving
  them revision histories would assert structure the specifications do not define.
