# ADR 0001 — Framework and SDK Implementation Language

- **Status:** Accepted
- **Record created:** 2026-07-27 (Packet L.0.C), reconstructing a decision already in force
  and observable throughout the repository. The file existed but was empty; this record
  states only what the repository and its rules already demonstrate.

## Context

PEOS separates two language questions that are easy to conflate:

1. the natural language in which the framework's normative documents are written;
2. the programming language in which the reference SDK is implemented.

PEOS-000 states that PEOS "standardizes engineering concepts, artifacts, decisions,
validation, and lifecycle independently of programming languages, frameworks, vendors, or
AI models," and lists programming languages among the things outside its scope. So the
second question is explicitly *not* a specification concern — it is an implementation
decision this repository owns, and therefore one that belongs in an ADR rather than in a
specification.

## Decision

**Framework documentation is written in English.** `CLAUDE.md`'s Language Policy states
this directly: PEOS framework documentation must be in English, while project-specific
artifacts use whatever language suits their stakeholders, and widely accepted software
engineering terms stay in English where translating them would reduce clarity.

**The reference SDK is implemented in Go**, module `github.com/aleka7sk/PEOS`, targeting
Go 1.22, with no third-party dependencies. Every package under `peos/` imports only the
standard library and other `peos/` packages.

## Consequences

- A specification remains implementable in any language; Go is the reference
  implementation, not part of the contract. PEOS-000's independence claim stays true.
- Go's value semantics, unexported struct fields, and constructor functions map directly
  onto the SDK's central discipline: immutable values whose invariants are enforced at
  construction and which cannot be mutated afterward. A language without that combination
  would require defensive copying or freezing conventions the SDK does not need.
- The zero-dependency constraint keeps the SDK auditable — the quality gates
  (`go vet`, `staticcheck`, `golangci-lint`, `-race`) cover the entire shipped surface.
- Translating a specification is permitted, but the English text remains normative. A
  translation that changes the meaning of business terminology is a defect, per
  `CLAUDE.md`.

## Alternatives considered

- **Specifying an implementation language normatively in PEOS-000.** Rejected: it would
  contradict PEOS-000's own scope statement and would make every non-Go implementation
  non-conformant for a reason unrelated to engineering semantics.
- **Allowing framework documentation in multiple languages.** Rejected for normative text:
  the specifications' precision depends on consistent RFC-2119 usage, and divergence
  between translations would create contradictions with no resolution rule.
