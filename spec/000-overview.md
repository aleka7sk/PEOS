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

## Engineering State

Engineering State is the persistent and inspectable body of engineering knowledge associated with a Product at a given point in its evolution.

Engineering State is represented through applicable Artifacts, Decisions, relationships, validation results, and their current normative status.

Engineering State does not require a single physical representation and MUST NOT be inferred solely from source code or runtime behavior.

---

## Artifact

An Artifact is an identifiable and persistent representation of engineering knowledge, intent, evidence, or product state.

An Artifact may be created, imported, derived, or transformed during Product Engineering.

Artifacts are first-class PEOS entities and are independent of any particular file format, storage system, development tool, or runtime.

The normative properties, identity, revision, representation, and relationship model of Artifacts are defined by PEOS-002.


---

## Decision

A Decision records an intentional engineering choice together with its rationale.

Every significant implementation should be traceable to one or more engineering decisions.

---

## Lifecycle

A Lifecycle defines the valid states and state transitions through which a governed PEOS entity may evolve.

A lifecycle may apply to a Product, Artifact, Decision, Validation, or another entity explicitly governed by a normative PEOS specification.

Lifecycle definitions establish permitted evolution, not organizational structure or a mandatory linear process.

The formal lifecycle and transition model is defined by PEOS-003.


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

## Actor

An Actor is an identifiable human, automated system, organization, or Runtime that performs or authorizes an engineering action.

An Actor may create, transform, review, validate, approve, or otherwise affect Engineering State.

Actor identity and authority are distinct. Identification of an Actor does not, by itself, establish that the Actor was authorized to perform an action.

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
