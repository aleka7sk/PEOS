# Contributing to PEOS

PEOS is specification-first. The specifications under `spec/` define what is required; the
Go SDK under `peos/` implements them. Contributions are evaluated against that order.

## Before writing code

Read, in this order:

1. `docs/implementation-progress.md` — the authoritative record of what is implemented,
   what is deferred, which audits have run, and what the next action is;
2. the relevant `spec/0XX-*.md` for the concept in scope;
3. the relevant package's `doc.go`, which states the clauses it implements and the
   boundaries it deliberately stays within.

`CLAUDE.md` documents the engineering process this repository holds itself to, including
the source-of-truth hierarchy and the evidence-and-certainty vocabulary
(fact / validated assumption / hypothesis / assumption / decision / open question).
`.claude/rules/peos-progress.md` governs how the tracker is read and maintained.

## What a change needs

- **A normative basis.** A new type, field, or rule needs a specification clause requiring
  it. Where a specification names a concept without defining its ontology, record a
  deferral rather than inventing structure.
- **Tests that enforce architecture, not just behaviour.** The repository's convention is
  to assert import boundaries, union closure, forbidden wire keys, failed-decode receiver
  preservation, explicit-null rejection, and the absence of concepts PEOS excludes — not
  only that the happy path works.
- **A tracker update.** Any change to implementation state belongs in
  `docs/implementation-progress.md`. Chat history is not a record.

## Quality gates

All of these must pass before a change is proposed:

```
gofmt -l .
go build ./...
go vet ./...
go test ./... -count=1
go test ./... -race -count=1
staticcheck ./...
golangci-lint run ./...
git diff --check
```

`staticcheck` and `golangci-lint` are optional external tools; the rest need only the Go
toolchain. Package coverage should not decrease.

## Conventions worth knowing

- Types are immutable. Constructors validate; modifiers return a new value.
- A field participating in a cross-field invariant is a **constructor argument**, not a
  modifier-only field — otherwise valid states become unconstructible.
- Derived state is never stored. No `status`, `current`, `compatible`, `conformant`, or
  `satisfied` field belongs on any type.
- Error sentinels are per-construct, not per-field, and nested sentinels are wrapped so
  `errors.Is` matches either level.
- Unknown JSON fields are ignored; explicit `null` for an optional field is rejected rather
  than treated as absent.
- The SDK depends only on the standard library. Adding a third-party dependency requires
  its own recorded decision.

## Reporting a specification problem

If a project exposes a weakness in PEOS, the specification should improve rather than the
project working around it. Open the issue against the specification, describing the
engineering situation the current text cannot express or expresses ambiguously — not the
code change you would prefer.
