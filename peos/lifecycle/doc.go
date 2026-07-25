// Package lifecycle implements the core, structural subset of PEOS-003's
// Lifecycle Model on top of the stable peos/core package.
//
// This packet (Packet E) implements the identity, structural, and
// immutable-record layer PEOS-003 requires: Lifecycle Definition,
// Lifecycle Definition Version, Lifecycle State, Transition Definition,
// State Assignment, and Transition Record. It deliberately does not
// implement Guard evaluation, Effect execution, Transition Attempt
// workflow, Current State/State History resolution, or Lifecycle
// Migration -- see the "Deferred" section below.
//
// # Lifecycle Definition is not obligatorily an Artifact
//
// PEOS-003 states that "A Lifecycle Definition MAY be represented as an
// Artifact" -- permissive, not mandatory. This package therefore does not
// model Definition or DefinitionVersion via core.Artifact or
// core.ArtifactRevision. Both types carry their own generic normative
// identity (core.LifecycleDefinitionID and
// core.LifecycleDefinitionVersionID, defined in peos/core). An
// Artifact-backed representation, for Products that want one, is left to
// a future additive profile built on top of this package, not assumed
// here.
//
// # Transition Record is the one exception: it is a persistent Artifact
//
// PEOS-003 explicitly and uniquely calls Transition Record "a persistent
// Artifact that records the attempted or completed application of a
// Transition," and requires it to "conform to PEOS-002." This is the
// opposite of the non-Artifact immutable-record pattern PEOS-006/007/008
// use for their own execution/observation/violation records. TransitionRecord
// therefore composes core.Artifact and core.ArtifactRevision by named
// field, exactly like peos/requirement.Requirement/Revision, and carries
// no independent, dedicated Transition Record identity type -- it uses
// core.ArtifactRef / core.ArtifactRevisionRef directly.
//
// # State Assignment is an immutable non-Artifact record
//
// Unlike Transition Record, PEOS-003 never calls a State Assignment "an
// Artifact" or requires it to conform to PEOS-002. StateAssignment is
// therefore modeled as an immutable value type carrying its own dedicated
// core.StateAssignmentID, analogous to how PEOS-006 models a Validation
// Claim.
//
// # State and Transition identity are scoped local keys, not global identities
//
// PEOS-003 requires State and Transition identifiers to be unique only
// within their owning Lifecycle Definition Version, not globally. StateID
// and TransitionID each wrap a validated core.LocalKey -- the same
// scoped-identity primitive PEOS-006's Planned Validation Activity and
// PEOS-009's Template Parameter already use for an identical scoping
// requirement -- but each is its own named Go type, not a type alias, so a
// StateID and a TransitionID (or a bare core.LocalKey from an unrelated
// construct) are never compile-time interchangeable with one another.
//
// # Initial-state policy belongs to DefinitionVersion, not State
//
// PEOS-003's "Initial State" section discusses initial-state policy
// ("one required initial State; multiple permitted initial States; a
// conditional initial-State selection rule") as a property of the
// Lifecycle Definition, not of an individual State. DefinitionVersion
// therefore owns an explicit initial-state set; State itself carries no
// "initial" flag.
//
// # Guard, Effect, and Trigger are intentionally not modeled
//
// PEOS-003 defines Guard's and Effect's normative *contract*
// (identifiable definition, determinable result, explicit failure
// consequence for Guard; a named set of possible consequences for Effect)
// but never a stable, executable expression language for either, and never
// uses the word "Trigger" as a distinct construct at all. This package
// therefore exposes no Guard, Effect, Trigger, expression interface, or
// Evaluate method -- adding one now would invent normative content PEOS-003
// does not define. TransitionDefinition carries only what PEOS-003
// unconditionally requires: identity, source States, and target States.
// Guard and Effect modeling is deferred to a later packet (informally,
// Packet E.2) once a stable expression contract exists.
//
// # Current State and State History are derived, not stored
//
// PEOS-003's own "Current State" and "State History" sections require
// both to be *determinable*, never a mutable stored field. Resolving
// either correctly also requires Product-level policy this packet does
// not yet have (same-timestamp tiebreaks, scope-sensitive composition,
// concurrent-assignment conflict handling). This package therefore exposes
// no ResolveCurrentState, ResolveHistory, CurrentState, or StateHistory
// API; it supplies only the immutable records (State Assignment,
// Transition Record) those future resolvers will read.
//
// # Lifecycle Migration is deferred
//
// Every StateAssignment and TransitionRecordContent is pinned to an exact
// core.LifecycleDefinitionVersionRef and never reinterpreted under a
// different Definition Version by this package.
//
// # Scope is owned by DefinitionVersion only
//
// core.Scope appears once, as a required DefinitionVersion field.
// StateAssignment and TransitionRecordContent carry no Scope field of
// their own: PEOS-003's own "Orthogonal States" example represents
// distinct lifecycle dimensions as distinct Lifecycle Definitions, not as
// a Scope stacked independently on every record family. A record-level
// Scope, if a Product ever needs one, is an additive normative change for
// a later packet, not part of this baseline.
package lifecycle
