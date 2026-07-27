package lifecycle_test

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/lifecycle"
)

// This example builds a State Assignment and a succeeded Transition Record
// Revision that departed from it. It demonstrates PEOS-003's two obligations
// for a completed Transition: a responsible Actor (carried by the enclosing
// core.ArtifactRevision's Provenance) and an identifiable authority basis
// (carried explicitly on TransitionRecordContent) — and the fact that both
// are enforced even though neither is duplicated into the other.
func Example_lifecycleTransition() {
	subject, err := core.NewLifecycleSubjectRefFromArtifact(mustArtifactRef("REQ-1"))
	if err != nil {
		panic(err)
	}
	definitionVersion, err := core.NewLifecycleDefinitionVersionRef(
		mustLifecycleDefinitionID("LC-REVIEW"),
		mustLifecycleDefinitionVersionID("V1"),
	)
	if err != nil {
		panic(err)
	}
	actor, err := core.NewActorRef("acme", "review-service")
	if err != nil {
		panic(err)
	}
	at, err := core.NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor).WithRecordedAt(at)

	// The source State Assignment: PEOS-003 requires a Transition Record to
	// identify "the source State Assignment", not merely a State name, since
	// a Subject may occupy the same State more than once over its history.
	fromAssignment, err := lifecycle.NewStateAssignment(
		mustStateAssignmentID("SA-1"),
		subject,
		definitionVersion,
		mustStateID("draft"),
		at,
		provenance,
		mustArtifactRevisionRef("TR-9001", "REV-0"),
	)
	if err != nil {
		panic(err)
	}
	fromRef, err := fromAssignment.Ref()
	if err != nil {
		panic(err)
	}

	// The resulting State Assignment a successful Transition establishes.
	resultingAssignment, err := lifecycle.NewStateAssignment(
		mustStateAssignmentID("SA-2"),
		subject,
		definitionVersion,
		mustStateID("accepted"),
		at,
		provenance,
		mustArtifactRevisionRef("TR-9001", "REV-1"),
	)
	if err != nil {
		panic(err)
	}
	resultingRef, err := resultingAssignment.Ref()
	if err != nil {
		panic(err)
	}

	content, err := lifecycle.NewTransitionRecordContent(
		subject,
		definitionVersion,
		mustTransitionID("submit-for-review"),
		fromRef,
		at,
		lifecycle.TransitionOutcomeSucceeded,
	)
	if err != nil {
		panic(err)
	}
	if content, err = content.WithToState(mustStateID("accepted")); err != nil {
		panic(err)
	}
	if content, err = content.WithCompletedAt(at); err != nil {
		panic(err)
	}
	if content, err = content.WithResultingAssignment(resultingRef); err != nil {
		panic(err)
	}
	// PEOS-003's Authority Invariant ("A completed Transition has an
	// identifiable authority basis") requires this for the succeeded
	// outcome; it is a distinct concept from the responsible Actor above.
	authority, err := core.NewAuthorityRef("acme", "review-board")
	if err != nil {
		panic(err)
	}
	if content, err = content.WithAuthority(authority); err != nil {
		panic(err)
	}

	transitionRecord, err := lifecycle.NewTransitionRecord(mustTransitionRecordArtifact("TR-9001"))
	if err != nil {
		panic(err)
	}
	revision, err := core.NewArtifactRevision(
		mustArtifactID("TR-9001"),
		mustArtifactRevisionID("REV-1"),
		mustOrigin(),
		provenance, // carries the responsible Actor
		mustIntegrityIdentity(),
	)
	if err != nil {
		panic(err)
	}

	transitionRecordRevision, err := lifecycle.NewTransitionRecordRevision(transitionRecord, revision, content)
	if err != nil {
		panic(err)
	}

	fmt.Println(transitionRecordRevision.Content().FromAssignment() == fromRef)
	_, hasActor := transitionRecordRevision.Core().Provenance().Actor()
	fmt.Println(hasActor)
	_, hasAuthority := transitionRecordRevision.Content().Authority()
	fmt.Println(hasAuthority)

	// Output:
	// true
	// true
	// true
}

func mustArtifactID(value string) core.ArtifactID {
	id, err := core.NewArtifactID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustArtifactRevisionID(value string) core.ArtifactRevisionID {
	id, err := core.NewArtifactRevisionID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustArtifactRef(artifactID string) core.ArtifactRef {
	ref, err := core.NewArtifactRef(mustArtifactID(artifactID))
	if err != nil {
		panic(err)
	}
	return ref
}

func mustArtifactRevisionRef(artifactID, revisionID string) core.ArtifactRevisionRef {
	ref, err := core.NewArtifactRevisionRef(mustArtifactID(artifactID), mustArtifactRevisionID(revisionID))
	if err != nil {
		panic(err)
	}
	return ref
}

func mustOrigin() core.Origin {
	o, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		panic(err)
	}
	return o
}

func mustIntegrityIdentity() core.IntegrityIdentity {
	i, err := core.NewIntegrityIdentity(
		core.IntegrityMechanismContentAddressedReference,
		"sha256:deadbeef",
		core.IntegrityProtectedScopeContent,
	)
	if err != nil {
		panic(err)
	}
	return i
}

func mustLifecycleDefinitionID(value string) core.LifecycleDefinitionID {
	id, err := core.NewLifecycleDefinitionID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustLifecycleDefinitionVersionID(value string) core.LifecycleDefinitionVersionID {
	id, err := core.NewLifecycleDefinitionVersionID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustStateAssignmentID(value string) core.StateAssignmentID {
	id, err := core.NewStateAssignmentID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustStateID(value string) lifecycle.StateID {
	id, err := lifecycle.NewStateID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustTransitionID(value string) lifecycle.TransitionID {
	id, err := lifecycle.NewTransitionID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustTransitionRecordArtifact(artifactID string) core.Artifact {
	a, err := core.NewArtifact(mustArtifactID(artifactID), lifecycle.ArtifactTypeTransitionRecord)
	if err != nil {
		panic(err)
	}
	return a
}
