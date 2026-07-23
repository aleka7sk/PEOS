# Product Engineering OS (PEOS)

> **Build products, not prompts.**

Product Engineering OS (PEOS) is an open specification for building software products with AI through a structured engineering process rather than ad-hoc prompting.

Instead of asking an AI model to immediately generate code, PEOS guides every project through a sequence of validated engineering stages—from understanding the problem to delivering production-ready software.

The goal is to make AI an engineering partner rather than a code generator.

---

# Why PEOS?

Modern AI can generate impressive amounts of code.

However, many software projects fail long before implementation because of:

* unclear problems;
* missing requirements;
* incorrect assumptions;
* premature architecture decisions;
* scope creep;
* lack of traceability;
* undocumented business rules.

Most AI workflows optimize for writing code.

PEOS optimizes for building successful products.

---

# Philosophy

PEOS is built around a simple principle:

> **Every line of code should trace back to a validated business need.**

The engineering process is treated as a pipeline where each stage produces artifacts consumed by the next stage.

```
Idea
    ↓
Discovery
    ↓
Research
    ↓
Product Definition
    ↓
Requirements
    ↓
Architecture
    ↓
Planning
    ↓
Implementation
    ↓
Review
    ↓
Release
```

Skipping stages is considered technical debt.

---

# Core Principles

## Problem before solution

Never design a solution before understanding the problem.

## Evidence over assumptions

Assumptions are explicitly marked until validated.

## Documents before code

Documentation defines the product.

Code implements the documentation.

## Small validated decisions

Large projects emerge from many validated decisions instead of one large design.

## Traceability

Every feature should be traceable to:

* a business goal;
* a user need;
* a requirement;
* an implementation task;
* production code.

---

# Repository Structure

```
.
├── docs/
├── templates/
├── examples/
├── .claude/
│   ├── skills/
│   ├── agents/
│   ├── rules/
│   └── settings.json
├── CLAUDE.md
└── README.md
```

---

# Workflow

PEOS divides software development into dedicated engineering stages.

Each stage is implemented as a Claude Code Skill.

Every Skill:

* has a single responsibility;
* consumes specific artifacts;
* produces specific artifacts;
* defines clear completion criteria;
* avoids making decisions outside its scope.

---

# Current Status

PEOS is currently under active development.

Version: **0.1.0**

The first reference project is an Excel-first CRM platform used to validate the framework against a real-world business workflow.

Future versions will be validated against additional domains, including SaaS products, mobile applications, marketplaces, and internal enterprise systems.

---

# Contributing

The framework evolves through real projects.

If a project exposes weaknesses in PEOS, the framework should improve rather than forcing projects to adapt.

Continuous refinement is a design goal, not a maintenance task.

---

# License

MIT License.
