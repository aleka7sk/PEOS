# PEOS-000 — Overview

**Category:** Normative

**Status:** Draft

**Version:** 0.1

---

# Abstract

This document defines the structure, scope, and interpretation of the PEOS Specification.

It establishes the foundational concepts used throughout the specification and defines how the remaining PEOS documents relate to one another.

Every normative document in PEOS assumes the definitions introduced here.

---

# Purpose

The purpose of PEOS is to provide an open specification for AI-assisted Product Engineering.

PEOS defines **how engineering work is structured**, not **how individual tools behave**.

It standardizes engineering concepts, artifacts, decisions, validation, and lifecycle independently of programming languages, frameworks, vendors, or AI models.

---

# Scope

PEOS specifies:

* engineering terminology;
* engineering artifacts;
* engineering lifecycle;
* engineering decisions;
* traceability between artifacts;
* validation rules;
* quality requirements;
* contracts between specifications and runtimes.

PEOS does not specify:

* programming languages;
* software architecture styles;
* deployment strategies;
* project management methodologies;
* specific AI models;
* implementation technologies.

---

# Specification Model

The PEOS Specification consists of independent normative documents.

Each document defines one responsibility.

Normative documents may reference other normative documents but must not redefine concepts already defined elsewhere.

Every concept introduced by the specification has exactly one authoritative definition.

---

# Core Concepts

## Specification

A Specification defines normative engineering rules.

Specifications describe **what must be true**, not how individual tools implement those rules.

---

## Product

A Product is the engineering objective for which artifacts, decisions, and implementations exist.

Everything within PEOS ultimately serves the Product.

---

## Artifact

An Artifact is a persistent engineering output created during product development.

Examples include requirements, architectural designs, implementation plans, source code, test plans, deployment plans, and review documents.

Artifacts are first-class entities within PEOS.

---

## Decision

A Decision records an intentional engineering choice together with its rationale.

Every significant implementation should be traceable to one or more engineering decisions.

---

## Lifecycle

A Lifecycle defines the sequence of engineering states through which a product evolves.

Lifecycle stages define progression—not organizational structure.

---

## Validation

Validation determines whether an artifact, decision, or implementation satisfies the requirements defined by the specification.

Validation evaluates conformance rather than correctness of business outcomes.

---

## Runtime

A Runtime is a software system capable of executing the PEOS Specification.

Examples include AI agents, automation tools, development assistants, and future implementations.

Runtimes implement the specification but do not define it.

---

## Template

A Template provides a standardized structure for creating artifacts.

Templates improve consistency but are not normative unless explicitly referenced by the specification.

---

# Specification Structure

The PEOS Specification is organized into independent documents.

PEOS-000 establishes the foundation.

Subsequent specifications define individual engineering domains.

Each document builds upon previously defined concepts while remaining responsible for exactly one domain.

---

# Normative Language

The following keywords indicate requirement strength.

**MUST**

Indicates an absolute requirement.

**MUST NOT**

Indicates an absolute prohibition.

**SHOULD**

Indicates a strong recommendation.

**SHOULD NOT**

Indicates that an action is generally discouraged.

**MAY**

Indicates optional behavior.

Unless explicitly stated otherwise, these keywords are interpreted as normative requirements.

---

# Conformance

An implementation is considered PEOS-compatible only if it satisfies all applicable normative requirements defined by the specification.

A runtime may extend PEOS behavior provided that such extensions do not contradict normative requirements.

---

# Document Categories

PEOS documents belong to one of two categories.

**Normative**

Documents that define engineering rules and requirements.

**Informational**

Documents that explain, demonstrate, or provide guidance without introducing normative requirements.

When a conflict exists, normative documents take precedence.

---

# Source of Truth

The `spec/` directory is the authoritative source of the PEOS Specification.

Runtime implementations, templates, tools, examples, and documentation must remain consistent with the specification.

No component outside `spec/` may redefine normative concepts.

---

# Versioning

The PEOS Specification evolves through versioned releases.

Backward compatibility, deprecation, and supersession are defined by the specification governance process.

---

# References

This document serves as the entry point for all subsequent PEOS specifications.

Every normative specification assumes the terminology and interpretation rules defined in PEOS-000.
