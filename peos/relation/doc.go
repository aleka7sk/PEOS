// Package relation implements PEOS-002's Artifact Relation contract on
// top of the stable peos/core package.
//
// # Relation has no PEOS identity, revision history, or lifecycle
//
// PEOS-002 §Artifact Relation states this directly: "An Artifact Relation
// SHALL NOT have normative identity, revision history, or lifecycle
// unless a future PEOS specification explicitly introduces such a
// model." PEOS-005 §17.1, the only specification that concretely defines
// relation types today, explicitly declines to be that future
// specification: "This specification does not define relationships as a
// separate category of engineering entity. The identity, revision,
// lifecycle, historical preservation, authority, and representation
// semantics of relationships are outside the scope of this
// specification." Relation is therefore not a core.Artifact
// specialization, has no ArtifactID, no RevisionID, no founding
// Revision, and no status/approval/validity-period field of any kind. It
// is a plain, immutable Go value — the same way core.Scope or
// core.Provenance are plain values, not identity-bearing records.
//
// # Relation is a binary source-to-target engineering assertion
//
// Every Relation identifies exactly one source and one target
// (PEOS-002: "Every Artifact Relation SHALL identify exactly one source
// participant and exactly one target participant"), together with a
// Relation Type and provenance. Direction is always structurally
// source-to-target, but whether that direction *carries normative
// meaning* — or whether a Relation Type is engineering-symmetric despite
// being structurally directed (PEOS-005 §22.1's Conflict is the
// canonical example: "Where an implementation requires ordered source
// and target fields, that ordering SHALL be representational only and
// SHALL NOT alter the symmetric engineering semantics of Conflict") — is
// a property of the Relation Type and its governing specification, not
// something this package interprets, stores as a boolean, or resolves by
// generating an automatic inverse record.
//
// # Typed endpoints reuse EngineeringSubjectRef
//
// Source and target are both core.EngineeringSubjectRef values. PEOS-002
// itself describes a relation subject as "an Artifact; an Artifact
// Revision; another entity explicitly permitted by a later normative
// PEOS specification" — exactly the union EngineeringSubjectRef already
// implements, including its opaque fallback for Product-specific or
// not-yet-typed subject kinds. No new endpoint type is introduced.
//
// # Packet D performs structural validation only
//
// New checks only that relationType, source, target, and provenance are
// each non-zero. It does not check source != target (some Relation
// Types forbid direct self-reference; others do not, and this is a
// semantic rule belonging to that Relation Type's own governing
// specification, not a universal rule this package can correctly
// enforce), endpoint-kind compatibility with the given Relation Type,
// Artifact-versus-Revision level compatibility (PEOS-002 §Artifact
// Supersession explicitly permits mixed-level relations "when explicitly
// permitted by a specialized normative PEOS specification" — a blanket
// prohibition here would be normatively wrong), referential existence of
// either endpoint, or duplicate/repeated relations (two structurally
// identical Relation values are simply equal values, not a uniqueness
// violation — nothing in PEOS-002 introduces a global-uniqueness rule,
// and enforcing one here would require a repository this package does
// not have).
//
// # Graph and specialized semantic validation remain external
//
// Cycle policy is explicitly per-Relation-Type, not universal (PEOS-002
// §Artifact Graph: "The Artifact Graph is not required to be globally
// acyclic. Cycle validity depends on the applicable Relation Types.").
// Cross-relation graph traversal, cycle detection, traceability
// coverage, and orphan detection are explicitly assigned to "a future
// Traceability Model" by PEOS-002 itself, and are not implemented here.
// This package holds no relation set, no repository, and no query
// mechanism — only the single, immutable Relation value type.
//
// # A Reference is not automatically a Relation
//
// A value that merely points at another value (an ArtifactRef inside
// some other construct's own content, a Requirement's Origin naming a
// source category, a Provenance's Actor reference) stays a reference as
// long as it lives inside one construct's own content, describing
// something about that construct. It becomes a Relation only when it is
// elevated to an independently identifiable, typed, directed assertion
// between two named participants, carrying its own Relation Type and its
// own provenance — separate from either participant's. Not every
// identifying field in this codebase is a hidden graph edge; most are
// ordinary references, and this package does not attempt to reify every
// reference it encounters into a Relation.
//
// # Integrity
//
// core.IntegrityProtectedScopeRelations (PEOS-002 §Artifact Integrity:
// "relations embedded in the Revision") applies when Relation values are
// embedded in, or otherwise protected by, some other construct's own
// Artifact Revision — for example, a future Revision that lists its own
// outbound relations as part of its content. Relation itself, defined by
// this package, carries no core.IntegrityIdentity and no independent
// integrity scope of its own, because it is not an Artifact Revision and
// has nothing of its own for an integrity identity to protect.
package relation
