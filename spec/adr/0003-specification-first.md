# ADR 0003 — Specification-First Source of Truth

- **Status:** Accepted
- **Record created:** 2026-07-27 (Packet L.0.C), reconstructing a decision already in force
  and enforced by this repository's own process rules. The file existed but was empty.

## Context

A framework that is developed alongside its reference implementation drifts in one of two
directions: either the specification is edited to match whatever the code does, or the code
accumulates behaviour the specification never authorized. Both destroy the framework's value,
because neither artifact can then be trusted as the answer to "what is required?"

## Decision

**The specifications are the source of truth; the SDK implements them.**
`CLAUDE.md` fixes the precedence explicitly: validated business evidence, then approved
scope, then approved requirements, then approved architecture decisions, then delivery
plans, then source code, then informal notes. It states directly that "Source code is not
the primary source of truth for business intent," and that a conflict between code and
approved requirements must be *reported*, not silently resolved in the code's favour.

**No SDK type exists without a normative clause requiring it.** Where a specification names
a concept but never defines its ontology, the SDK records a deferral rather than inventing
structure. Template Migration is the clearest case: PEOS-009 states nine things every
migration must identify but never states what a Migration *is* — not an Artifact, not a
record, not a relation — so `peos/template` implements none of it and the Deferred
Architecture section records why.

**Work proceeds in packets with a fixed lifecycle.** Blueprint (read-only architecture
derivation from the specification) → Implementation → Audit (independent, read-only,
treating the implementation report as a claim to verify) → Corrective packet if needed →
Closure. `.claude/rules/peos-progress.md` governs the tracker that records this, and states
that a checkbox is a claim rather than a fact: implementation existence, passing tests,
passing quality gates, and a recorded audit verdict are each verified independently before
a packet is accepted.

**`docs/implementation-progress.md` is the authoritative implementation record.** Chat
history is not a record. Decisions, deferrals, findings, and their resolutions live in the
repository or they do not exist.

## Consequences

- Every implementation decision is traceable to a specification clause, and audits classify
  normative statements against shipped code rather than reviewing code on its own terms.
- Ambiguity in a specification is resolved by explicit reasoning that is then recorded, not
  by picking whichever reading is convenient. The `peos/relation` scope-cardinality
  resolution and the PEOS-003 completed-Transition Actor scope are both examples.
- Gaps stay visible. Nine concepts are currently deferred, each with its normative source,
  its reason, and the question that must be answered before implementation.
- The process costs more than writing code directly, and it is meant to. The alternative is
  a specification that documents whatever the code happened to do.

## Alternatives considered

- **Code-first, with specifications written afterward as documentation.** Rejected: it makes
  the specification a description rather than a contract, and removes any basis for calling
  an implementation non-conformant.
- **Treating passing tests as sufficient acceptance evidence.** Rejected, and the reason is
  recorded in the repository's own history: PEOS-002 and PEOS-003 were accepted on
  implementation progression and quality gates without a dedicated audit packet, and a later
  review found items that the audited specifications' process would have surfaced earlier.
