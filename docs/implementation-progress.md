# PEOS Implementation Progress

- **Last updated:** 2026-07-26 (Packet G.1.1.I — Derivation identity distinctness correction)
- **Current specification:** PEOS-005 — Requirement Model
- **Current packet:** Packet G (Blueprint) → G.A (Challenge Review) → G.B (Consolidated Blueprint) → G.B.1 (tracker sync) → G.1 (Requirement Relationship Foundation + Derivation) → G.2 (Refinement + Decomposition) → G.1.1 (Derivation Identity Inheritance Audit) → **G.1.1.I (Derivation Identity Distinctness Correction) are all complete and accepted.** Packet G.3 has **not** started.
- **Overall status:** 5 of 10 specifications have a Go SDK package (`peos/core`, `peos/requirement`, `peos/relation`, `peos/lifecycle`, `peos/decision`). PEOS-004 is fully complete and accepted, zero open findings, one deliberate architectural deferral (Delegation). PEOS-005 is in progress: architecture is fully resolved (Packet G.B); G.1 (Derivation, §18), G.2 (Refinement §19, Decomposition §20), and the G.1.1/G.1.1.I audit-and-correction pair are all complete. **Derivation now requires distinct Requirement ArtifactIDs between source and target** (source.ArtifactID() != target.ArtifactID()), correcting a genuine conformance gap in the original G.1 implementation: PEOS-005 §18's "SHALL NOT inherit the identity of a source Requirement" is the same rule §20.1 states for Decomposition in different words (PEOS-009 :664/:758 is the cross-spec terminology precedent establishing "inherit" = "share" an identity). Refinement is unaffected and remains the sole type permitting two Revisions of one Requirement identity to participate together (§19 states no distinctness rule). No new sentinel was introduced — both `ErrInvalidDerivation` and `ErrInvalidDecomposition` already existed. G.3–G.5 have not started. PEOS-006 through PEOS-009 have no implementation package yet.
- **Exact next action:** **Packet G.3 — Dependency + Conflict Blueprint** (PEOS-005 §21, §21.1, §22, §22.1). Recommended model: Claude Opus. Recommended mode: Plan. See "Current Next Action" below.

This file is the single authoritative source for PEOS implementation status. Chat history, prior reports, and memory are not authoritative — this file, verified against repository evidence, is. See `.claude/rules/peos-progress.md` for the maintenance rule.

---

## Status Legend

Standard Markdown checkboxes (`[x]` / `[ ]`) cannot express every state a PEOS packet actually occupies, so this tracker uses an extended set:

| Symbol | Meaning |
|---|---|
| `[x]` | **Complete and accepted** — implementation exists, required tests exist and pass, required quality gates pass, and any required audit has an accepted verdict (ACCEPT / ACCEPTED / COMPLETE), with no open BLOCKER or MAJOR finding. Open MINOR/NOTE findings do not block this state, but must stay listed until fixed or explicitly waived. |
| `[ ]` | **Not started or incomplete** — no implementation exists, or existing implementation does not yet satisfy its own packet's scope. |
| `[~]` | **In progress** — implementation exists but a required step (tests, quality gates, or audit) has not yet run or has not yet returned a verdict. Includes the specific case "Implemented — audit pending." |
| `[!]` | **Blocked** — an accepted audit or review returned a BLOCKER or MAJOR finding; a corrective packet is required before further work in that area proceeds. |
| `[-]` | **Deferred by architecture** — a concept PEOS defines, but which this SDK deliberately does not implement yet, for a stated architectural reason (not oversight, not lack of time). Never use `[x]` for a deferred concept. |

---

## Specification Progress

| Spec | Status | Package / location | Completed packets | Remaining packets | Audit state | Evidence |
|---|---|---|---|---|---|---|
| PEOS-000 — Overview | `[x]` | No dedicated package; interpretive foundation cited throughout `peos/core` and every other package's doc comments | N/A (definitional document, not a domain model) | None identified | N/A | `peos/core/doc.go:1-9` frames every package against "PEOS-000 through PEOS-009 (see spec/)"; no normative construct of its own to implement |
| PEOS-001 — Philosophy | `[x]` | Same as above | N/A (foundational axioms, not a domain model) | None identified | N/A | Same as above; philosophy is realized as the constructor-validates-invariants pattern used package-wide, not as its own type set |
| PEOS-002 — Artifact Model | `[x]` | `peos/core` (`Artifact`, `ArtifactRevision`, `Representation`, `ArtifactID`, `ArtifactRevisionRef`, `ArtifactType`) — Packet B / B.1; `peos/relation` (`Relation`) — Packet D / D.1 | Packet B (Artifact + Revision + Representation), Packet B.1 (integrity/representation refinement), Packet D (Relation), Packet D.1 (Relation structural validation) | Cross-relation graph traversal, cycle detection, traceability coverage, orphan detection — explicitly assigned by PEOS-002 itself to "a future Traceability Model" | No formal audit packet recorded in this history; accepted by direct commit progression (`eaafa23`→`258bb66`→`aebadd0`, `1753354`) with passing tests at each step | `go test ./peos/core ./peos/relation` pass; coverage core 80.8%, relation 98.1% (verified 2026-07-26); `peos/relation/doc.go` documents the graph/traceability deferral explicitly |
| PEOS-003 — Lifecycle | `[x]` | `peos/lifecycle` (`LifecycleDefinition`, `LifecycleDefinitionVersion`, `StateAssignment`, `TransitionRecord`, `LifecycleDefinitionVersionSupersession`) — Packet E / E.1 | Packet E (Definition, Version, State Assignment, Transition Record), Packet E.1 (Definition Version Supersession) | Guard/Effect/Trigger expression language (Packet E.2, "once a stable expression contract exists"); Current State / State History derivation; Lifecycle Migration execution | No formal audit packet recorded; accepted by commit progression (`129dd70`, `25567f1`, `1807fc2`) with passing tests | `go test ./peos/lifecycle` passes; coverage 85.9% (verified 2026-07-26); `peos/lifecycle/doc.go` documents all three deferrals explicitly with named future-packet placeholders |
| PEOS-004 — Decision Model | `[x]` | `peos/decision` (`Decision`, `Authority`, `Basis`+`Assumption`/`Constraint`/`Uncertainty`, `Alternative`, `Outcome`+`Commitment`, `Record`, `DecisionSupersession`, `DecisionInvalidation`, `DecisionConflict`, `ConflictResolution`, `Role`, `Consequence`) — Packets F, F.1, F.2, F.2.A, F.3, F.3 audit, F.3.B | See full "Packet Progress" table below — Packets F through F.3.B (including all sub-audits) are all complete and accepted, zero open findings | **Delegation** — deferred by architecture, see "Deferred Architecture" below | F.2.A: **ACCEPT** (1 MINOR, fixed in F.3). F.3 post-implementation audit: **ACCEPTED WITH MINOR FINDINGS** (2 MINOR, 0 BLOCKER/MAJOR). F.3.B corrective test packet: **both MINOR findings closed** (MINOR-1 fixed with new nested-sentinel tests; MINOR-2 accepted as non-blocking, no action taken) | `go test ./peos/decision` passes, 434 tests, 0 failures; coverage unchanged at ~91.8%; `staticcheck`/`golangci-lint` clean; `peos/decision/doc.go` documents final status for every PEOS-004 concept including the Delegation deferral with exact missing type and owning-package question |
| PEOS-005 — Requirement Model | `[~]` | `peos/requirement` (`Requirement`, `Revision`, `Content`, `Subject`, `Statement` — Packet C; `Derivation` — Packet G.1, **corrected by G.1.1.I**; `Refinement`, `Decomposition` — Packet G.2) + `peos/relation` (dependency, introduced by G.1). **Authoritative architecture plan: Packet G.B** | Packet C (§6–§16); Packet G/G.A/G.B (architecture); Packet G.1 (§17, §17.1–§17.4, §18, §18.1–§18.2 — foundation + Derivation); Packet G.2 (§19, §19.1, §20, §20.1–§20.2 — Refinement + Decomposition); **Packet G.1.1/G.1.1.I (Derivation §18 identity-distinctness audit and correction, complete and accepted)** | **G.3 — Dependency + Conflict; G.4 — Requirement Supersession; G.5 — Waiver.** No Allocation implementation packet (Allocation is Product-owned) | Packet G.A: **BLUEPRINT APPROVED WITH REQUIRED REVISIONS** (2 MAJOR, 7 MINOR, 2 NOTE, all incorporated by G.B). Packet G.B: **PACKET G BLUEPRINT REVISED AND ACCEPTED.** Packets G.1 and G.2: implemented per their Blueprints. **Packet G.1.1: DERIVATION ARTIFACT ID DISTINCTNESS REQUIRED** (normative audit; genuine conformance gap found in G.1). **Packet G.1.1.I: correction applied and verified**, no architectural deviation, all quality gates pass | `go test ./peos/requirement` passes, coverage unchanged at 95.1% (verified 2026-07-26; test-only + doc-comment addition, no new production statements beyond one generalized helper signature). **Minimum conformance (§34.1): still 9 of 10** — Dependency, Conflict, and Requirement Supersession remain (G.3–G.4) before §17–§23 is fully closed. **Requirement-identity distinctness is now required by both Derivation (`ErrInvalidDerivation`, §18) and Decomposition (`ErrInvalidDecomposition`, §20.1); Refinement alone is exempt (§19 imposes no such rule)** — corrected from G.2's original two-way framing, which mistakenly treated Decomposition as the sole type enforcing it. No §18 identity question remains open |
| PEOS-006 — Validation | `[ ]` | None | None | Everything — Claim, Validation Plan, Validation Execution Record, Validation Method | N/A | No `peos/validation` (or similarly named) directory exists. Spec text only (`spec/006-validation.md`, 36928 bytes) |
| PEOS-007 — Quality Model | `[ ]` | None | None | Everything — Quality Profile, Quality Element, Quality Constraint | N/A | No implementation package exists. Spec text only (`spec/007-quality-model.md`) |
| PEOS-008 — Runtime Contract | `[ ]` | None | None | Everything — Runtime Contract, Runtime Binding, Runtime Observation, Runtime Violation | N/A | No implementation package exists. Spec text only (`spec/008-runtime-contract.md`); commit `3498d52 "PEOS-008,009"` (2026-07-25) only edited spec text, not code |
| PEOS-009 — Template Contract | `[ ]` | None | None | Everything — Template, Template Application Record, Template Constraint | N/A | No implementation package exists. Spec text only (`spec/009-template-contract.md`) |

---

## PEOS-005 Binding Architecture Summary

The full binding architecture lives in **Packet G.B** (the consolidated Blueprint); this is
a pointer summary only, kept here so that Packets G.1–G.5 cannot silently reopen accepted
decisions. Do not treat this summary as a substitute for reading Packet G.B in full before
starting any G.1–G.5 work.

- **Dependency direction:** `peos/requirement → peos/relation → peos/core`. No cycle. `peos/requirement` does not import `peos/decision`.
- **§17–§23 architecture:** typed Requirement-specific wrappers around `relation.Relation` — not standalone records, not Requirement-aware validation inside `peos/relation`.
- Constructors accept typed semantic parts only and build the inner `relation.Relation` themselves; they **never** accept a pre-built `relation.Relation`.
- Requirement relationships have **no independent identity** (no ID field, no Ref type, no lifecycle, no revision history, no `id` JSON key) — per PEOS-005 §17.1.
- Requirement Supersession's lifecycle consequence has **exactly two** mandatory-explicit semantic states (consequence identified / explicitly no consequence); no conformant "unstated" state exists; a nil/optional pointer is an insufficient encoding.
- Requirement Supersession's governance action is a closed two-arm union (`core.DecisionOutcomeRef` | opaque Product-contract mechanism) — **not** `core.AuthorityRef`.
- **Waiver has no identity** — a value record owned by `peos/requirement`, not an Artifact Relation, no lifecycle.
- **Allocation is Product-owned** (`PROD`) — no SDK implementation debt, no implementation packet.

---

## Packet Progress

Execution order, as reconstructed from `git log --reverse` and each package's own doc-comment cross-references (commit messages are informal and do not always match packet letters 1:1 — see the Evidence column for the disambiguating commit).

| Packet | Status | Scope | Implementation commit | Audit result | Remaining findings | Next dependency |
|---|---|---|---|---|---|---|
| A | `[x]` | `peos/core` foundational identity/reference/vocabulary primitives, pre-dating Artifact | Earliest `peos/core` commits, prior to `eaafa23` (exact boundary not separately tagged) | Not formally audited as a named packet; superseded/extended by every later core commit | None open | — |
| B | `[x]` | `peos/core` Artifact, ArtifactRevision, Representation | `eaafa23`, `258bb66` | Not formally audited; accepted by commit progression + passing tests | None open | — |
| B.1 | `[x]` | ArtifactRevision integrity/representation refinement | `aebadd0` | Not formally audited | None open | — |
| C | `[~]` | `peos/requirement` — Requirement, Revision, Content, Subject, Statement (§6–§16) | `a044d46`, `158eca6` | Audited by Packet G/G.A/G.B: §6–§16 confirmed HAVE, zero deviation from specification found; §34.1 minimum conformance 9/10, gap is §17–§23 only | Two stale `doc.go` statements (Artifact Relations; Waiver) — confirmed by Packet G.B, corrections assigned to **Packet G.1**, not yet applied | Packet G.1 |
| G | `[x]` | Initial PEOS-005 Blueprint: first conformance matrix + roadmap draft | Architecture only, no code | Superseded in full by Packet G.B following Packet G.A's Challenge Review — **not an active competing Blueprint** | None open (all findings folded into G.B) | — |
| G.A | `[x]` | Adversarial Challenge Review of the Packet G Blueprint | Architecture only, no code | **BLUEPRINT APPROVED WITH REQUIRED REVISIONS** — 2 MAJOR (Allocation misclassified as DEFER; Supersession consequence miscounted as 3 states), 7 MINOR, 2 NOTE | All 11 findings incorporated by Packet G.B | Packet G.B |
| G.B | `[x]` | Consolidated, authoritative PEOS-005 Blueprint incorporating every Packet G.A finding | Architecture only, no code | **PACKET G BLUEPRINT REVISED AND ACCEPTED** | None open | Packet G.1 |
| G.B.1 | `[x]` | Progress tracker synchronization: bring `docs/implementation-progress.md` in line with Packet G.B (this update) | `docs/implementation-progress.md` only, no code, no tests | Self-contained documentation-only corrective packet; validated by `git diff --check` and manual consistency review | None open | Packet G.1 |
| G.1 | `[x]` | Requirement Relationship Foundation + Derivation Blueprint (§17, §17.1–§17.4, §18, §18.1–§18.2) | Architecture only, no code | Blueprint complete: full wrapper architecture, constructor contract, JSON contract, error model, test matrix, validation boundary matrix — see Packet G.1.I for implementation result | None open | Packet G.1.I |
| G.1.I | `[x]` | Requirement Relationship Foundation + Derivation implementation, following the Packet G.1 Blueprint literally (no redesign) | New: `peos/requirement/relationship.go`, `derivation.go`, `derivation_test.go`. Modified: `errors.go` (2 sentinels: `ErrInvalidRequirementRelation`, `ErrInvalidDerivation`), `doc.go` (both stale statements corrected + new Requirement Relationship section) | Implemented exactly per Blueprint; zero architectural deviation. All quality gates pass (`gofmt`, `go build`, `go vet`, `go test` ×2 incl. `-race`, `staticcheck`, `golangci-lint`, `git diff --check`); 32 new Derivation tests, 0 failures; package coverage 94.6% (up from 94.2%) | None open | Packet G.2 |
| G.2 | `[x]` | Refinement + Decomposition Blueprint (§19, §19.1, §20, §20.1–§20.2) | Architecture only, no code | Blueprint complete: reused G.1 foundation unchanged, specified mandatory scope, Decomposition's §20.1 identity-distinctness rule, §20.2 no-completeness-model constraint, JSON contract, error model, test matrix, validation boundary matrix — see Packet G.2.I for implementation result | None open | Packet G.2.I |
| G.2.I | `[x]` | Refinement + Decomposition implementation, following the Packet G.2 Blueprint literally (no redesign; G.1's foundation and Derivation untouched) | New: `peos/requirement/refinement.go`, `decomposition.go`, `refinement_test.go`, `decomposition_test.go`. Modified: `relationship.go` (2 new helpers, additive only: `requireRelationScope`, `checkDistinctRequirementIdentity`), `errors.go` (1 new sentinel: `ErrInvalidDecomposition`; no `ErrInvalidRefinement`), `doc.go` | Implemented exactly per Blueprint; zero architectural deviation. All quality gates pass (`gofmt`, `go build`, `go vet`, `go test` ×2 incl. `-race`, `staticcheck`, `golangci-lint`, `git diff --check`); 62 new tests (30 Refinement, 32 Decomposition), 0 failures; package coverage 95.1% (up from 94.6%); `derivation.go`/`derivation_test.go` verified untouched (`git diff --stat`); absence audit confirms no `DecompositionSet`/`RelationshipCollection`/`CompletenessAssertion` symbol exists | None open | Packet G.3 |
| G.1.1 | `[x]` | Derivation Identity Inheritance Audit — normative-analysis-only re-examination of whether PEOS-005 §18 requires `source.ArtifactID() != target.ArtifactID()` for Derivation | Read-only, no code, no tracker update performed within the audit itself | **DERIVATION ARTIFACT ID DISTINCTNESS REQUIRED.** Decisive evidence: PEOS-009 (:664, :758) establishes that "inherit an identity" means "share an identity," settling §18:681's "SHALL NOT inherit the identity of a source Requirement" as requiring distinct ArtifactIDs — the same rule §20.1 states for Decomposition in different words. Refinement remains the genuine outlier (no clause in §19) | Confirmed a real conformance gap in the shipped G.1 implementation, not merely a documentation gap | Packet G.1.1.I |
| G.1.1.I | `[x]` | Derivation Identity Distinctness Correction — implements the G.1.1 audit's conclusion | Modified: `relationship.go` (generalized `checkDistinctRequirementIdentity` to accept a caller-supplied sentinel, relation-type-agnostic, no behavior change to its Decomposition call site beyond the added argument), `derivation.go` (`NewDerivation` now also rejects `source.ArtifactID() == target.ArtifactID()`, reusing the existing `ErrInvalidDerivation` sentinel — no new sentinel), `derivation_test.go` (+4 tests: constructor rejection, JSON decode rejection with receiver-preservation check, constructor/Unmarshal equivalence, plus the existing self-derivation test continues to pass unchanged), `doc.go` (corrected the "Decomposition alone enforces identity distinctness" section to the accurate three-way pattern, citing PEOS-009). `refinement.go`/`refinement_test.go` and `errors.go` confirmed untouched | Zero architectural deviation; all quality gates pass (`gofmt`, `go build`, `go vet`, `go test` ×2 incl. `-race`, `staticcheck`, `golangci-lint`, `git diff --check`); all pre-existing Derivation/Refinement/Decomposition tests continue to pass unchanged, confirming Refinement's asymmetric acceptance and Decomposition's existing rejection both still hold | None open | Packet G.3 |
| G.3 | `[ ]` | Dependency + Conflict (§21, §21.1, §22, §22.1) | Not started | — | — | G.1 (satisfied — `relationship.go` foundation, including the new `requireRelationScope` helper, ready for reuse) |
| G.4 | `[ ]` | Requirement Supersession (§23, §23.1–§23.2, §26.5, §36.12–§36.13) | Not started | — | — | G.1 |
| G.5 | `[ ]` | Waiver (§27, §27.1–§27.2, §36.11) — no-identity value record, independent of G.1–G.4 | Not started | — | — | — |
| D | `[x]` | `peos/relation` — Relation (structural validation only) | `1753354` | Not formally audited as a named packet | None open | — |
| D.1 | `[x]` | Relation structural-validation refinement (scope, self-reference policy) | Folded into `1753354`; see `peos/relation/doc.go:46` | N/A | None open | — |
| E | `[x]` | `peos/lifecycle` — Definition, DefinitionVersion, StateAssignment, TransitionRecord | `129dd70` | Not formally audited | None open | — |
| E.1 | `[x]` | LifecycleDefinitionVersionSupersession | `25567f1`, `1807fc2` | Not formally audited | None open | Packet E.2 (Guard/Effect expression language) remains `[-]` deferred, not started |
| F | `[x]` | `peos/decision` initial — Decision, Authority, Basis (evidence-only), Alternative, Outcome+Commitment, Record | `d32b7ab` | Accepted (superseded by F.1/F.2/F.3 without reopening core decisions) | None open | — |
| F.1 | `[x]` | DecisionSupersession, DecisionInvalidation, package-local identity types | `17a1b22` | Accepted; F.1.A audit referenced by F.2's own doc comments as having flagged one MINOR test-hardening item (X-06), closed within F.1/F.2 test files | None open | — |
| F.2 | `[x]` | Basis widened to Evidence/Assumption/Constraint/Uncertainty; `NewBasisFrom`; 4 collection mutators | `e0785cb` | **F.2.A Audit: ACCEPT.** 1 MINOR finding (defensive-copy test gap on `WithEvidence`/`WithConstraints`/`WithUncertainties`) | **Resolved** — the 3 missing tests were added during F.3 implementation (`TestBasisWithEvidenceInputSliceCopied`, `TestBasisWithConstraintsInputSliceCopied`, `TestBasisWithUncertaintiesInputSliceCopied`, all in `basis_test.go`, verified present and passing) | — |
| F.3 | `[x]` | DecisionConflict, ConflictResolution, Role, Consequence; Decision integration; Delegation deferral recorded | `b361c92` | **Post-implementation audit: ACCEPTED WITH MINOR FINDINGS** (performed 2026-07-26; 0 BLOCKER, 0 MAJOR, 2 MINOR) | **Resolved** — see Packet F.3.B row below for how both were closed | — |
| F.3.B | `[x]` | Corrective test-only packet: nested-sentinel test hardening for `DecisionConflict`, `ConflictResolution`, `Role` (no production code changed — all three sentinel chains were already correct; only test assertions were incomplete) | Uncommitted at time of this update (working tree changes to `conflict_test.go`, `resolution_test.go`, `role_test.go`) | Self-contained corrective packet; validated by full quality gate re-run (`gofmt`, `go build`, `go vet`, `go test` ×2 incl. `-race`, `staticcheck`, `golangci-lint`, all clean) | **MINOR-1 (from F.3): fixed.** Added `TestDecisionConflictNestedDecisionRefSentinelPreserved`, `TestConflictResolutionNestedResolvingDecisionSentinelPreserved`, `TestRoleNestedActorSentinelPreserved`; strengthened the existing `TestDecisionConflictNestedSentinelPreserved` (previously asserted only `ErrInvalidDecisionConflict` against a payload that actually produced `core.ErrInvalidVocabularyValue`, not `core.ErrInvalidScope` — corrected the payload to isolate `core.ErrInvalidScope` and added that assertion). All four tests now assert both the `peos/decision` sentinel and the underlying `peos/core` sentinel via `errors.Is`. **MINOR-2 (from F.3): accepted as non-blocking, no action** — `ConflictID`/`ConflictResolutionID` remain in `conflict.go`/`resolution.go` per explicit instruction not to move them into `identity.go` | — |

---

## Deferred Architecture

Every intentionally deferred PEOS concept, tracked independently of packet completion. **A deferred item is never `[x]`.**

### `[-]` Delegation (PEOS-004)

- **Missing concept:** An authority-grant record — identity; delegator (`core.ActorRef` | `core.AuthorityRef`); delegate `core.ActorRef`; delegated scope; effective period or condition; applicable constraints; revocation conditions.
- **Normative source:** PEOS-004 :824-839 ("A delegation MUST identify: the delegating Actor or authority; the receiving Actor; the delegated scope; applicable constraints; the effective period or condition; revocation conditions when applicable.")
- **Reason for deferral:** Architectural, not representational. Every standalone aggregate in `peos/decision` names at least one Decision; a delegation record would be the package's only orphan aggregate with no Decision-scoped consumer. `peos/lifecycle` already consumes `core.AuthorityRef` (`StateAssignment.authority`, `TransitionRecordContent.authority`) without importing `peos/decision`; PEOS-003 :627 requires delegated Runtime authority to be inspectable, which would force `peos/lifecycle` to import `peos/decision` (inverting dependency direction, since lifecycle is more foundational) if the delegation record lived there. `peos/core/authority.go` explicitly declines to define an Authority aggregate today ("this package does not define an Authority aggregate, does not give AuthorityRef a lifecycle ... this package stops at the reference") — a delegation record with identity/scope/effective-period/constraints/revocation *is* such an aggregate.
- **Ownership or dependency that must be resolved:** Which package owns authority-grant aggregates — `peos/core` (admitting the aggregate it currently declines) or a new sibling package dedicated to governance grants — and whether `core.AuthorityRef` gains a link to the delegation that produced it.
- **Future packet responsible:** Not yet named/numbered. Must be scoped as its own packet once the owning-package question above is resolved; cannot be folded into a `peos/decision` packet.

### `[-]` Guard / Effect / Trigger expression language (PEOS-003, informally "Packet E.2")

- **Missing concept:** An executable expression language and `Evaluate` interface for Guard (identifiable definition, determinable result, explicit failure consequence) and Effect (named set of possible consequences).
- **Normative source:** PEOS-003 defines Guard's and Effect's normative *contract* but never a stable, executable expression language for either, and never uses "Trigger" as a distinct construct at all.
- **Reason for deferral:** `peos/lifecycle/doc.go` states directly: modeling one now "would invent normative content PEOS-003 does not define." `TransitionDefinition` carries only what PEOS-003 unconditionally requires (identity, source States, target States).
- **Ownership or dependency that must be resolved:** A stable, PEOS-defined (or Product-contract-defined) expression contract must exist before an `Evaluate`-shaped interface can be added without inventing normative content.
- **Future packet responsible:** Packet E.2 (informal name already used in `peos/lifecycle/doc.go:93`).

### `[-]` Lifecycle Migration execution (PEOS-003)

- **Missing concept:** Remapping State under a different Lifecycle Definition Version, reinterpreting historical Transitions, tracking migration progress.
- **Normative source:** Referenced implicitly by PEOS-003's Lifecycle Definition Version Supersession material; `peos/lifecycle` records the Supersession *fact* only.
- **Reason for deferral:** `peos/lifecycle/doc.go` ("Lifecycle Migration is deferred") — every `StateAssignment`/`TransitionRecordContent` is pinned to an exact `core.LifecycleDefinitionVersionRef` and never reinterpreted; this package does not execute Migration.
- **Ownership or dependency that must be resolved:** A Migration execution model — almost certainly a Runtime-layer concern rather than a `peos/lifecycle` domain-value concern, but not yet decided.
- **Future packet responsible:** Not yet named.

### `[-]` Current State / State History resolution (PEOS-003)

- **Missing concept:** Deriving the "current" or "effective" Lifecycle State of a subject from its history of State Assignments and Transition Records.
- **Normative source:** PEOS-003, cross-referenced by `peos/core/doc.go` ("Derived-state resolution is out of scope... PEOS does not define a universal rule for selecting one record from a record's history").
- **Reason for deferral:** No universal PEOS rule exists to derive this; it is repository/Runtime-dependent by design.
- **Ownership or dependency that must be resolved:** A repository or Runtime layer, not a domain-value package.
- **Future packet responsible:** Not yet named; explicitly out of scope for any `peos/*` domain-value package per current architecture.

### `[-]` Cross-relation graph traversal, cycle detection, traceability coverage, orphan detection (PEOS-002)

- **Missing concept:** Any multi-Relation analysis — traversal, cycle validity per Relation Type, coverage/orphan detection.
- **Normative source:** PEOS-002 §Artifact Graph explicitly assigns this to "a future Traceability Model," and states cycle validity is per-Relation-Type, not universal.
- **Reason for deferral:** `peos/relation/doc.go` — this package holds no relation set, no repository, no query mechanism, only the single immutable `Relation` value type.
- **Ownership or dependency that must be resolved:** A future Traceability Model package (PEOS-002's own words), which does not yet have a spec number or package name.
- **Future packet responsible:** Not yet named; blocked on PEOS itself defining the Traceability Model.

### Resolved — PEOS-005 Requirement Relations / Allocation / Waiver / Acceptance Criterion

**No PEOS-005 concept remains deferred as of Packet G.B.** This entry previously listed
Requirement Relations, Allocation, Waiver, and Acceptance Criterion as blocked on
dependencies that "should be re-checked." Packet G.B re-checked them and resolved every
one — none is architecture debt any longer:

- **Requirement Relations (Derivation/Refinement/Decomposition/Dependency/Conflict/Supersession, §17.4):** not deferred — scheduled as **Packets G.1–G.4**. The original blocker ("Artifact Relations not yet implemented in core") was already false at the time this entry was written; `peos/relation` (Packet D) existed.
- **Waiver (§27):** not deferred — scheduled as **Packet G.5**. The original blocker ("requires a Decision Outcome that does not yet exist") was already false; `peos/decision.Outcome` and `core.DecisionOutcomeRef` existed.
- **Allocation (§24):** not deferred, and never requires an implementation packet — reclassified **`PROD`**. §24 contains no positive "SHALL identify" obligation (only prohibitions and independence rules), is permissive ("Requirements MAY be allocated"), and is absent from §34.1 minimum conformance. A Product may compose `relation.Relation` with an opaque target, or use its own record.
- **Acceptance Criterion / Verification Method:** correctly out of PEOS-005 scope — owned by the future Validation Model (PEOS-006), which has no PEOS-005 implementation debt attached to it.

Lifecycle integration (§26) was never a deferral: §26 states PEOS-005 lifecycle is
"governed exclusively by PEOS-003," and `peos/lifecycle` already owns it.

---

## Quality Gates

All results below were re-executed directly in this task (Packet F.3.B) against the working tree on 2026-07-26, after the nested-sentinel test hardening; none are carried over from memory or an unverified prior report.

| Gate | Result | Evidence |
|---|---|---|
| `gofmt -l peos/decision` | **PASS** (clean) | No output |
| `go build ./...` | **PASS** | Exit 0 |
| `go vet ./...` | **PASS** | Exit 0 |
| `go test ./peos/decision -count=1` | **PASS** | `ok`, 434 tests (up from 431 pre-F.3.B) |
| `go test ./... -count=1` | **PASS** | `core`, `decision`, `lifecycle`, `relation`, `requirement` all `ok` |
| `go test ./... -race -count=1` | **PASS** | All 5 packages `ok` under the race detector |
| `staticcheck ./...` | **PASS** (clean) | No output |
| `golangci-lint run ./...` | **PASS** (clean) | No output |
| `git diff --check` | **PASS** (clean) | No output |
| Package coverage | `core` 80.8% · `decision` ~91.8% (unchanged — test-only addition, no new production statements) · `lifecycle` 85.9% · `relation` 98.1% · `requirement` 94.2% | `go test ./peos/<pkg> -coverprofile=...` per package |
| Architecture audit (Packet F.3) | **APPROVED** | Adversarial Challenge Review + Revised Architecture — 3 required revisions applied (`ConflictResolution.resolvingDecision` made optional, `Consequence` reclassified MUST→SHOULD, Delegation removed) before implementation began |
| Implementation audit (Packet F.3) | **ACCEPTED WITH MINOR FINDINGS** → **now fully resolved by F.3.B** | Post-implementation audit found 0 BLOCKER, 0 MAJOR, 2 MINOR; both closed (see Packet Progress table, F.3.B row) |
| Architecture audit (Packet G) | **BLUEPRINT APPROVED WITH REQUIRED REVISIONS** → **fully resolved by Packet G.B** | Packet G.A Challenge Review found 2 MAJOR, 7 MINOR, 2 NOTE; all incorporated into the Packet G.B consolidated Blueprint, none outstanding |
| `gofmt -l peos/requirement` (Packet G.1.I) | **PASS** (clean) | No output |
| `go test ./peos/requirement -count=1` (Packet G.1.I) | **PASS** | `ok`, includes 32 new `Test Derivation*` tests, 0 failures |
| `go test ./... -count=1` (Packet G.1.I) | **PASS** | `core`, `decision`, `lifecycle`, `relation`, `requirement` all `ok` |
| `go test ./... -race -count=1` (Packet G.1.I) | **PASS** | All 5 packages `ok` under the race detector |
| `staticcheck ./...` / `golangci-lint run ./...` (Packet G.1.I) | **PASS** (clean) | No output |
| `git diff --check` (Packet G.1.I) | **PASS** (clean) | No output |
| Package coverage (Packet G.1.I) | `requirement` 94.6% (up from 94.2%) | `go test ./peos/requirement -coverprofile=...` |
| Implementation compliance (Packet G.1.I) | **Zero architectural deviation** from the Packet G.1 Blueprint | Wrapper shape, constructor signature, JSON envelope, sentinel names, and decode-revalidation flow all implemented exactly as specified |
| `gofmt -l peos/requirement` (Packet G.2.I) | **PASS** (clean) | No output |
| `go test ./... -count=1` (Packet G.2.I) | **PASS** | `core`, `decision`, `lifecycle`, `relation`, `requirement` all `ok`, incl. 62 new `TestRefinement*`/`TestDecomposition*` tests, 0 failures |
| `go test ./... -race -count=1` (Packet G.2.I) | **PASS** | All 5 packages `ok` under the race detector |
| `staticcheck ./...` / `golangci-lint run ./...` (Packet G.2.I) | **PASS** (clean) | No output |
| `git diff --check` (Packet G.2.I) | **PASS** (clean) | No output |
| Package coverage (Packet G.2.I) | `requirement` 95.1% (up from 94.6%) | `go test ./peos/requirement -coverprofile=...` |
| Implementation compliance (Packet G.2.I) | **Zero architectural deviation** from the Packet G.2 Blueprint | Mandatory-scope handling, §20.1 identity-distinctness check, sentinel count (1 new, no `ErrInvalidRefinement`), and JSON envelope all implemented exactly as specified; `derivation.go`/`derivation_test.go` confirmed untouched |
| `go test ./peos/requirement -run 'TestDerivation\|TestNewDerivation\|TestRefinement\|TestNewRefinement\|TestDecomposition\|TestNewDecomposition' -count=1` (Packet G.1.1.I) | **PASS** | All pre-existing tests continue to pass unchanged; +4 new Derivation tests for the corrected invariant |
| `go test ./... -count=1` / `go test ./... -race -count=1` (Packet G.1.1.I) | **PASS** | All 5 packages `ok`, incl. race detector |
| `staticcheck ./...` / `golangci-lint run ./...` (Packet G.1.1.I) | **PASS** (clean) | No output |
| `git diff --check` (Packet G.1.1.I) | **PASS** (clean) | No output |
| No new sentinel (Packet G.1.1.I) | **CONFIRMED** | `errors.go` absent from the packet's diff entirely — both `ErrInvalidDerivation` and `ErrInvalidDecomposition` reused |
| Scope discipline (Packet G.1.1.I) | **CONFIRMED** | `refinement.go`/`refinement_test.go` diff empty (verified via `git diff --stat`); `decomposition.go` diff contains only the one required call-site argument addition |
| Implementation compliance (Packet G.1.1.I) | **Zero architectural deviation** from the Packet G.1.1 audit's conclusion | Generalized helper, corrected constructor, decode-only inherits the fix automatically (no duplicated check), documentation rewritten to the accurate three-way pattern |

**Documentation-only note (Packets G.A, G.B, G.B.1):** none of these three packets changed
any Go file. G.A and G.B were pure architecture review/consolidation with no repository
writes at all; G.B.1 changed only `docs/implementation-progress.md`. Packet G.1 (Blueprint)
was likewise architecture-only. **Packet G.1.I is the first code change since Packet F.3.B**
— it is a real implementation packet, and the full quality-gate suite above was run fresh
against it, not carried over.

---

## Current Next Action

**There is exactly one primary next action:**

> **Packet G.3 — Dependency + Conflict Blueprint** (PEOS-005 §21, §21.1, §22, §22.1)
>
> Design the two either-participant-level relationship types on top of the Packet G.1/G.2
> foundation (`peos/requirement/relationship.go`'s shared helpers, including the G.2-added
> `requireRelationScope` for Conflict's mandatory scope; the `ErrInvalidRequirementRelation`
> sentinel; the nested JSON envelope shape; the decode-revalidation pattern — reuse all of
> these, do not redesign them). Unlike every relationship type shipped so far, both
> Dependency and Conflict permit participants at *either* Requirement identity level or
> Requirement Artifact Revision level, and both permit cycles (§21.1 explicitly; §22.1 for
> multi-participant Conflict cycles) rather than prohibiting them. Dependency requires no
> source≠target rule (§21/§21.1 state none — do not add one). Conflict requires a mandatory
> scope, a mandatory "nature of incompatibility" field, symmetric engineering semantics
> despite an ordered stored representation (§22.1: "that ordering SHALL be representational
> only and SHALL NOT alter the symmetric engineering semantics of Conflict"), and a
> source≠target check (§22.1, unlike Dependency).
>
> This packet was briefly gated behind the Packet G.1.1/G.1.1.I audit-and-correction pair,
> which resolved an open normative question about Derivation's §18 identity rule before G.3
> could safely reuse the shared relationship foundation. That gate is now closed: G.1.1.I is
> complete and accepted, and Packet G.3 is unblocked.

- **Recommended model:** Claude Opus
- **Recommended mode:** Plan (blueprint phase); exit Plan Mode only once the Dependency/Conflict shape is approved
- **Expected result:** A Packet G.3 Blueprint at the same level of detail as Packets G.1's and G.2's, ready for direct implementation with no further architectural decisions

**Historical / superseded next actions — do not act on these:** every next action recorded
in this section before Packet G.1.1.I (the original "Packet G — conformance-matrix audit
phase," "Packet G.1 — … Blueprint," "Packet G.2 — … Blueprint," and the intervening
G.1.1/G.1.1.I audit-and-correction gate) is now complete and superseded. **Packet G.3,
above, is the sole current next action in this document, and it is unblocked.**
