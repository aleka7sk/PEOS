// Package lifecycle implements the core, structural subset of PEOS-003's
// Lifecycle Model on top of the stable peos/core package.
//
// This package implements the identity, structural, and immutable-record
// layer PEOS-003 requires: Lifecycle Definition, Lifecycle Definition
// Version, Lifecycle State, Transition Definition, State Assignment,
// Transition Record, and Lifecycle Definition Version Supersession
// (Packet E.1). It deliberately does not implement Guard evaluation,
// Effect execution, Transition Attempt workflow, Current State/State
// History resolution, or Lifecycle Migration -- see the "Deferred"
// section below.
//
// # Lifecycle Definition is not obligatorily an Artifact
//
// PEOS-003 states that "A Lifecycle Definition MAY be represented as an
// Artifact" -- permissive, not mandatory. This package therefore does not
// model Definition or DefinitionVersion via core.Artifact or
// core.ArtifactRevision. Both types carry their own generic normative
// identity (core.LifecycleDefinitionID and
// core.LifecycleDefinitionVersionID, defined in peos/core) as their
// canonical identity.
//
// Lifecycle identity is canonical. Artifact binding is optional. Since
// Packet E.1, Definition and DefinitionVersion each expose an optional
// Artifact correspondence (Definition.WithArtifact /
// DefinitionVersion.WithArtifactRevision) for Products that do choose
// Artifact representation. This binding exists specifically to satisfy
// PEOS-003's conditional requirement: "When a Lifecycle Definition is
// represented as an Artifact, its Definition Version MUST identify the
// corresponding Artifact Revision." When Artifact-backed, DefinitionVersion
// identifies the corresponding core.ArtifactRevisionRef through that
// binding. The binding does not erase or replace Lifecycle identity: a
// Definition or DefinitionVersion with no Artifact binding at all remains
// fully valid, and core.LifecycleDefinitionID / core.LifecycleDefinitionVersionID
// remain the identity these types use everywhere else in this package
// (State Assignment, Transition Record) regardless of whether a binding
// is present. See DefinitionVersion.ValidateArtifactBinding for the pure,
// local consistency check between a Definition's and a DefinitionVersion's
// bindings.
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
// # A Transition Record's endpoints and its responsible Actor
//
// PEOS-003 requires a Transition Record to identify "the source State
// Assignment" and "the resulting target State Assignment when successful".
// Both endpoints are therefore core.StateAssignmentRef values on
// TransitionRecordContent -- fromAssignment (mandatory) and
// resultingAssignment (set when successful). A StateID names a State inside
// a Definition Version and cannot say which of a Subject's Assignments to
// that State a Transition departed from, since a Subject may occupy the same
// State more than once over its history. Packet L.0.C replaced the original
// fromState StateID field with fromAssignment for that reason, changing the
// wire key from "from_state" to "source_state_assignment"; no parallel
// StateID field or compatibility alias is retained. toState remains a
// StateID, because it names the targeted State rather than an Assignment.
//
// PEOS-003 also states two obligations for a completed Transition, and this
// package satisfies them in two different places.
//
// First, "Every completed Transition MUST identify the responsible Actor or
// Runtime." This package does not duplicate the Actor inside
// TransitionRecordContent: a Transition Record is a PEOS-002 Artifact, so its
// Revision already carries core.Provenance, and that Provenance's Actor *is*
// the responsible transition Actor.
//
// Second, PEOS-003 lists "authority basis" among a Transition Record's items
// **without a qualifier** -- in direct contrast to "failure or override
// information when applicable" three items below it in the same list -- and
// states an Authority Invariant: "A completed Transition has an identifiable
// authority basis." The authority basis is stored explicitly, as
// TransitionRecordContent.Authority(), and is mandatory for a completed
// Transition. Packet L.0.C had left it optional for every outcome and
// described that as "exactly as PEOS-003 leaves it"; Packet L.0.D found that
// description inaccurate, and Packet L.0.E applies the obligation.
//
// Because neither the content nor the core.ArtifactRevision can check either
// rule alone, newTransitionRecordRevisionFromParts -- the single construction
// boundary both NewTransitionRecordRevision and UnmarshalJSON pass through --
// enforces both, rejecting a succeeded-outcome revision whose Provenance
// carries no Actor (ErrMissingResponsibleActor) or whose content carries no
// Authority (ErrMissingTransitionAuthority), each wrapped inside
// ErrInvalidTransitionRecordRevision.
//
// Both invariants stop at the succeeded outcome. PEOS-003's Target State
// Invariant -- "A completed Transition establishes an explicitly permitted
// target State" -- identifies which outcome is a completed Transition:
// validateOutcome requires a target State and a resulting State Assignment for
// succeeded and forbids both for failed, interrupted, and indeterminate. A
// Transition Attempt that was rejected, failed, interrupted, or left
// indeterminate is not a completed Transition, and PEOS-003 states neither
// obligation for one; extending either rule to those outcomes, or to a
// Product-declared outcome this package has no normative basis to constrain,
// would invent normative content. Non-succeeded outcomes therefore keep the
// optional Authority semantics they have always had.
//
// Actor and Authority remain distinct concepts and neither substitutes for the
// other -- PEOS-003: "Actor identity and transition authority are distinct."
// A record missing both surfaces the Actor cause first, in check order.
//
// **Authority derivation is not implemented.** PEOS-003 requires a Lifecycle
// Definition to define which Actors or authority classes may perform a
// Transition, so an authority basis could in principle be derived from the
// Definition Version together with the Actor. This package models neither
// authority classes nor a policy language, so no such derivation would be
// reliable; Authority() is read directly and nothing else is consulted.
//
// Applicable Evidence stays optional and is not duplicated into content:
// PEOS-003 makes it conditional on a Lifecycle Definition requiring it ("A
// Lifecycle Definition MAY require Evidence"), scopes its MUST to *Required*
// Evidence, defers retention to "the applicable contract," and assigns
// validity to PEOS-006. The Revision already carries Origin, Provenance, and
// Representations. The Provenance recorded time is not a replacement for
// attemptedAt or completedAt, which are separate, domain-meaningful times.
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
// # Lifecycle Definition Version Supersession is distinct from Artifact Supersession and Migration
//
// LifecycleDefinitionVersionSupersession (supersession.go) records the
// PEOS-003 "Supersession" fact -- that one Lifecycle Definition Version
// supersedes another -- as its own immutable, independently identified
// record. It is not Artifact Supersession (PEOS-002): PEOS-003 states
// "Artifact or Artifact Revision supersession remains governed by PEOS-002
// and applicable specialized lifecycles," and a Lifecycle Definition
// Version is not an Artifact Revision in the first place. It is also not
// Lifecycle Migration (see below): recording that a supersession occurred,
// with its effective scope and normative consequences, is a materially
// smaller and different concern than executing a migration.
//
// # Lifecycle Migration is deferred
//
// Every StateAssignment and TransitionRecordContent is pinned to an exact
// core.LifecycleDefinitionVersionRef and never reinterpreted under a
// different Definition Version by this package. This package records the
// Supersession fact (above) but does not execute Migration: it does not
// remap State, does not reinterpret historical Transitions, and tracks no
// migration progress.
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
