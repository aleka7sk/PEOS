# Product Engineering OS (PEOS)

> **Every implementation decision should be traceable to an approved product need.**

**Status: v1.0.0 release candidate.** All ten specifications are implemented and
accepted; no `v1.0.0` tag has been created yet. See "Release contract" below for what
that will mean once tagged.

PEOS is a set of ten normative specifications describing how software product engineering
state is modeled — artifacts, revisions, lifecycles, decisions, requirements, validation,
quality, runtime contracts, and templates — together with a Go SDK that implements those
specifications as immutable, self-validating value types.

The specifications are the product. The SDK exists to make them checkable by a compiler
rather than only by review.

---

## Quick start

```
go get github.com/aleka7sk/PEOS
```

Requires Go 1.22 or later. Zero third-party dependencies — the SDK imports only the
standard library.

```go
import "github.com/aleka7sk/PEOS/peos/core"

vocab, _ := core.NewVocabularyValue("acme", "widget")
artifactID, _ := core.NewArtifactID("ART-1001")
artifact, _ := core.NewArtifact(artifactID, core.NewArtifactType(vocab))
```

That is the smallest building block every package composes on: an Artifact with stable
identity. See `peos/core/example_test.go` for the full, compiling version — constructing
a Revision alongside it, with Provenance and a typed reference.

New here? Read **[docs/consumer-guide.md](docs/consumer-guide.md)** — it explains the SDK
to an external consumer without requiring all ten specifications first, including a
worked cross-package example (`examples/crosspackage`) showing how two packages compose
through `core` reference types without importing each other.

---

## Repository structure

```
.
├── spec/                      # PEOS-000 .. PEOS-009, the normative specifications
│   └── adr/                   # architecture decision records
├── peos/                      # the Go SDK, one package per specification domain
│   ├── core/                  # PEOS-002 identity, references, vocabularies, provenance
│   ├── relation/              # PEOS-002 Artifact Relation contract
│   ├── lifecycle/             # PEOS-003
│   ├── decision/              # PEOS-004
│   ├── requirement/           # PEOS-005
│   ├── validation/            # PEOS-006
│   ├── quality/                # PEOS-007
│   ├── runtime/                # PEOS-008
│   └── template/               # PEOS-009
├── examples/
│   └── crosspackage/           # compiling cross-package workflow example
├── docs/
│   ├── consumer-guide.md       # start here if you're new to the SDK
│   └── implementation-progress.md   # authoritative implementation tracker
├── CHANGELOG.md
├── VERSION
├── CLAUDE.md                   # engineering process rules for this repository
├── MANIFESTO.md
└── README.md
```

`templates/`, `runtimes/`, and `tools/` remain **empty placeholder directories** for
future work. `examples/` now contains one real compiling example (`crosspackage/`)
alongside three empty product-scenario placeholder subdirectories
(`mobile-app`, `saas-platform`, `simple-api`).

---

## The specifications

| Spec | Title | Package |
|---|---|---|
| PEOS-000 | Overview | — (interpretive foundation) |
| PEOS-001 | Philosophy | — (interpretive foundation) |
| PEOS-002 | Artifact Model | `peos/core`, `peos/relation` |
| PEOS-003 | Lifecycle | `peos/lifecycle` |
| PEOS-004 | Decision Model | `peos/decision` |
| PEOS-005 | Requirement Model | `peos/requirement` |
| PEOS-006 | Validation Model | `peos/validation` |
| PEOS-007 | Quality Model | `peos/quality` |
| PEOS-008 | Runtime Contract | `peos/runtime` |
| PEOS-009 | Template Contract | `peos/template` |

PEOS-000 and PEOS-001 have no dedicated package by design: they are an interpretive
foundation cited throughout every package's documentation, not a domain model.

---

## Architecture

**Specification-first.** No type exists in the SDK unless a specification clause requires
it. Where a specification defines a concept but not its ontology, the SDK records the gap
as a deliberate deferral rather than inventing structure — see *Deferrals* below.

**Immutable artifacts and revisions.** An Artifact carries stable identity; an Artifact
Revision is an immutable recorded state of it. Nothing in the SDK mutates a recorded
Revision. Records that need amendment reference an earlier record through a typed
correction reference; they never rewrite it.

**Constructors validate; values are immutable.** Every type is constructed through a
function that enforces its invariants, and every modifier returns a new value rather than
mutating the receiver. A field that participates in a cross-field invariant is a
constructor argument, not a modifier-only field, so no valid state is unconstructible.

**Derived state is never stored.** Current lifecycle state, satisfaction, conformance,
compatibility, and quality are derived views. No SDK type carries a `status`, `current`,
`compatible`, `conformant`, or `satisfied` field, and tests assert their absence from
every wire form.

**Graph analysis is out of scope.** Traversal, cycle detection, traceability coverage, and
orphan detection are assigned by PEOS-002 itself to a future Traceability Model. The SDK
holds no repository, no relation set, and no query mechanism. See
`docs/consumer-guide.md` section 6 for exactly what a repository must add.

### Package dependency direction

```
core  (foundation — imports no PEOS package)
 ├── relation      → core
 ├── lifecycle     → core
 ├── decision      → core
 ├── validation    → core
 ├── requirement   → core, relation
 ├── quality       → core, validation
 ├── runtime       → core, validation
 └── template      → core, relation, validation
```

The graph is acyclic and enforced by tests: every package carries a `doc_test.go` that
parses its own imports and fails on any dependency its rule does not permit, plus
assertions locking the converse directions.

---

## Release contract

This is the public compatibility contract for the PEOS Go SDK, effective from the
`v1.0.0` tag once created. Nothing below claims compatibility with a prior release — no
version of this SDK has been published before.

1. **The normative specifications are PEOS-000 through PEOS-009.** They are the source of
   truth for engineering semantics; this repository's `CLAUDE.md` states the full
   source-of-truth hierarchy.
2. **The Go SDK is one conformant implementation of those specifications** — not the
   specification itself. A different implementation, in a different language, could
   conform to the same ten specifications without sharing this SDK's API shape.
3. **The public v1 surface** includes: package import paths; exported types; exported
   constructors; exported methods; exported error sentinels; vocabulary constants
   (`ArtifactType*`, `RelationType*`, `RecordKind*`, `CriterionKind*`, `SubjectKind*`, and
   similar); JSON field names; JSON union discriminator values; and `core.Extension`
   semantics.
4. **Breaking changes require a new major version.** A breaking change is any change to
   an item in the public v1 surface that would fail to compile or fail to decode existing
   valid data against the new version — removing or renaming an exported symbol,
   changing a constructor's signature, renaming a JSON field or discriminator value, or
   changing a sentinel's matching behavior.
5. **Additive exported APIs are allowed in minor versions**, provided they introduce no
   new required construction step for existing types and alter no existing semantics — a
   new optional modifier, a new sentinel for a newly-added validation path, or a new
   package are all minor-version-compatible additions.
6. **Deprecated APIs remain available for at least one minor release** after being marked
   deprecated, unless a security or correctness issue makes that impossible — in which
   case the exception and its reason are recorded in CHANGELOG.md.
7. **JSON policy:**
   - existing field names and discriminator values are stable in v1;
   - an existing required field cannot become optional if doing so would weaken a
     normative invariant;
   - an existing optional field cannot become mandatory in a minor release;
   - new optional fields may be added at any time;
   - unknown fields remain tolerated — this is the SDK's permanent forward-compatibility
     default, not a temporary migration aid;
   - explicit-null rejection behavior for optional fields remains part of the wire
     contract and will not silently change to "treat null as absent";
   - Product-specific data must remain inside the documented `core.Extension` mechanism —
     it is not a place to smuggle in a PEOS concept a future specification amendment
     might define;
   - **no generic migration promise is made.** The SDK does not commit to shipping
     migration tooling between major versions; a breaking JSON change, should one ever be
     required, will be documented in CHANGELOG.md as such, not silently absorbed.
8. **Future ontology or architecture changes require either a new PEOS specification, or
   an accepted ADR** (`spec/adr/`) when the change affects implementation or release
   policy without changing the normative ontology itself. Neither this README nor any
   single packet may expand PEOS ontology unilaterally.

Two wire-format changes were made before this baseline and are recorded, not as
migrations, but as part of what v1.0.0 *is*: see CHANGELOG.md's "Pre-release corrections."

---

## Quality gates

```
gofmt -l .
go build ./...
go vet ./...
go test ./... -count=1
go test ./... -race -count=1
staticcheck ./...
golangci-lint run ./...
```

`staticcheck` and `golangci-lint` are optional external tools; the first five commands
require only the Go toolchain. All eight gates pass on the current tree.

Per-package coverage:

```
go test ./peos/<package> -cover
```

---

## Current status

All ten specifications are implemented and accepted. Nine Go packages, 2,011 test
functions, 9 compiling package-level examples plus one cross-package workflow example,
145 error sentinels, and 2,289 exported production symbols. Coverage by package: `core`
84.5%, `relation` 98.1%, `lifecycle` 87.2%, `decision` 92.0%, `requirement` 95.4%,
`validation` 97.4%, `quality` 97.4%, `runtime` 96.8%, `template` 97.7%.

`docs/implementation-progress.md` is the authoritative record of what is implemented,
what is deferred, which audits have run, and what the next action is. It supersedes this
section whenever the two disagree.

The SDK is a value layer. It deliberately provides **no** repository, storage backend,
query engine, execution engine, renderer, workflow orchestrator, or CLI. Those are
Product concerns, and several are explicitly assigned elsewhere by the specifications
themselves — see `docs/consumer-guide.md` section 6.

See **CHANGELOG.md** for what changed in this release candidate.

---

## Deferrals

The following are specified by PEOS but deliberately not implemented, each recorded with
its normative source and the question that must be resolved first. The full list, with
reasoning and the owning layer once resolved, is in `docs/consumer-guide.md` section 10
and the Deferred Architecture section of `docs/implementation-progress.md`.

- Specialized Artifact Supersession enforcement (PEOS-002) — the relation is
  representable through `relation.Relation`; the specialized scope and
  self-supersession rules are not yet enforced by a wrapper.
- Cross-relation graph traversal, cycle detection, traceability coverage, orphan
  detection (PEOS-002) — assigned by the specification to a future Traceability Model.
- Guard / Effect expression language (PEOS-003) — no stable expression contract exists
  to implement against.
- Lifecycle Migration execution and Current State / State History resolution (PEOS-003).
- Delegation (PEOS-004) — blocked on which package should own authority-grant aggregates.
- Runtime Waiver attached conditions and criterion-level subject (PEOS-008).
- Template Migration (PEOS-009) — PEOS-009 states what a migration must identify but
  never what a Migration *is*.

None of these block using the SDK for the workflows README and the consumer guide
describe.

---

## Contributing

The framework evolves through real projects. If a project exposes a weakness in PEOS, the
specification should improve rather than the project working around it.

`CLAUDE.md` documents the engineering process this repository holds itself to: a
source-of-truth hierarchy, an evidence-and-certainty vocabulary, scope control, and the
requirement that every implementation decision trace back to an approved product need.
See `CONTRIBUTING.md` for what a change needs before it is proposed.

---

## License

MIT — see [LICENSE](LICENSE).
