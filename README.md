# Product Engineering OS (PEOS)

> **Every implementation decision should be traceable to an approved product need.**

PEOS is a set of ten normative specifications describing how software product engineering
state is modeled — artifacts, revisions, lifecycles, decisions, requirements, validation,
quality, runtime contracts, and templates — together with a Go SDK that implements those
specifications as immutable, self-validating value types.

The specifications are the product. The SDK exists to make them checkable by a compiler
rather than only by review.

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
│   ├── quality/               # PEOS-007
│   ├── runtime/               # PEOS-008
│   └── template/              # PEOS-009
├── docs/
│   └── implementation-progress.md   # authoritative implementation tracker
├── CLAUDE.md                  # engineering process rules for this repository
├── MANIFESTO.md
└── README.md
```

`templates/`, `examples/`, `runtimes/`, and `tools/` exist as **empty placeholder
directories** for future work. They contain no files today and nothing in the SDK depends
on them.

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
holds no repository, no relation set, and no query mechanism.

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

## Using the SDK

```
go get github.com/aleka7sk/PEOS
```

Requires Go 1.22 or later. The module has no third-party dependencies — the SDK imports
only the standard library.

Package documentation is the primary reference; each package's `doc.go` states the
specification clauses it implements, the boundaries it deliberately stays within, and the
reasoning behind each architectural choice:

```
go doc github.com/aleka7sk/PEOS/peos/core
go doc github.com/aleka7sk/PEOS/peos/requirement
```

---

## Running tests and quality gates

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

All ten specifications are implemented and accepted. Nine Go packages, 2,005 test
functions, 144 error sentinels, and 2,289 exported production symbols. Coverage by
package: `core` 84.4%, `relation` 98.1%, `lifecycle` 86.5%, `decision` 92.0%,
`requirement` 95.4%, `validation` 97.4%, `quality` 97.4%, `runtime` 96.8%,
`template` 97.7%.

`docs/implementation-progress.md` is the authoritative record of what is implemented,
what is deferred, which audits have run, and what the next action is. It supersedes this
section whenever the two disagree.

The SDK is a value layer. It deliberately provides **no** repository, storage backend,
query engine, execution engine, renderer, workflow orchestrator, or CLI. Those are
Product concerns, and several are explicitly assigned elsewhere by the specifications
themselves.

---

## Deferrals

The following are specified by PEOS but deliberately not implemented, each recorded with
its normative source and the question that must be resolved first. The full list, with
reasoning, is in the Deferred Architecture section of
`docs/implementation-progress.md`.

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

---

## Contributing

The framework evolves through real projects. If a project exposes a weakness in PEOS, the
specification should improve rather than the project working around it.

`CLAUDE.md` documents the engineering process this repository holds itself to: a
source-of-truth hierarchy, an evidence-and-certainty vocabulary, scope control, and the
requirement that every implementation decision trace back to an approved product need.

---

## License

MIT — see [LICENSE](LICENSE).
