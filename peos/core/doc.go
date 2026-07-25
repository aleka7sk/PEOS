// Package core provides the shared value types, identities, references,
// and vocabulary primitives used across the PEOS Core Go SDK.
//
// PEOS-000 through PEOS-009 (see spec/) are the sole normative source for
// the framework this package supports. Nothing in this package introduces
// a new normative construct, redefines an existing one, or resolves a
// normative ambiguity the specifications leave open; every type here is an
// implementation-level value type whose job is to let later PEOS SDK
// packages (Artifact, Lifecycle, Decision, Requirement, Validation,
// Quality, Runtime, Template) share one non-duplicated foundation.
//
// # Constructors are the conformance path
//
// Every identity, vocabulary, timestamp, and reference type in this
// package is constructed through an explicit function (e.g. NewArtifactID,
// NewVocabularyValue) that performs the validation PEOS conformance
// requires (non-empty identity, required vocabulary components, non-zero
// timestamps, structurally consistent references). Go's type system
// cannot prevent a caller from bypassing a constructor with a struct
// literal or an explicit conversion between two structurally identical
// types; callers and validators built on top of this package MUST treat
// values obtained outside a constructor (e.g. via reflection, decoding
// into an exported field, or an explicit type conversion) as unverified
// and MUST re-validate them before relying on PEOS conformance guarantees.
//
// # References are typed by participant level
//
// An identity-level reference (e.g. ArtifactRef) and a revision-level
// reference (e.g. ArtifactRevisionRef) are distinct Go types with
// distinct fields, not one type with an optional revision field. A value
// cannot be constructed in a state where its participant level and its
// stored identity components disagree.
//
// # Subject and criteria are distinct unions
//
// EngineeringSubjectRef (the subject of a Validation/Quality/Compliance/
// Template Conformance Claim) and CriterionRef (a Claim's criteria) are
// separate tagged unions with non-overlapping constructor sets where the
// underlying PEOS specifications restrict them. This lets a future
// validator mechanically reject a Requirement used as both a Claim's
// subject and its own criterion without depending on runtime discriminator
// comparison across two otherwise-identical types.
//
// # Relation identity does not exist at the PEOS level
//
// PEOS-002 gives an Artifact Relation no normative identity. No type in
// this package, or in any later PEOS SDK package built on top of it,
// should expose a relation identifier as anything more than an
// implementation-local storage key.
//
// # Derived-state resolution is out of scope
//
// This package does not resolve "current" or "effective" values of any
// kind (current Lifecycle State, currently applicable Claim, derived
// Requirement satisfaction, current Runtime Binding, and so on). PEOS
// does not define a universal rule for selecting one record from a
// record's history, and no such rule is implemented here.
//
// # Artifact and Artifact Revision are independent records
//
// Artifact and ArtifactRevision are independent domain values connected
// only by ArtifactID; Artifact does not contain, embed, or cache any
// Revision, and has no "current Revision" field (PEOS-002 treats the
// current or applicable Revision as derived or Product-selected, never a
// stored field). Representation is Revision-owned content with no PEOS
// identity of its own; its ownership by one ArtifactRevision is
// established structurally, by being stored inside that Revision's
// Representations, not by a reference field on Representation itself. A
// combined "Artifact plus its Revisions" shape is an application-level
// export or interchange envelope, not part of this domain model.
package core
