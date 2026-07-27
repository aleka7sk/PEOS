package core_test

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This example shows the smallest building block every PEOS package builds
// on: an Artifact with stable identity, and one immutable Artifact Revision
// recording a state of it. Nothing here is specific to any one specification
// domain — every domain package (requirement, validation, quality, runtime,
// template) composes this same pair.
func Example_artifactAndRevision() {
	artifactID, err := core.NewArtifactID("ART-1001")
	if err != nil {
		panic(err)
	}
	artifactType := core.NewArtifactType(mustVocabularyValue("acme", "widget"))

	artifact, err := core.NewArtifact(artifactID, artifactType)
	if err != nil {
		panic(err)
	}

	// Provenance records who/what produced a Revision and when. It is
	// mandatory on every ArtifactRevision.
	actor, err := core.NewActorRef("acme", "build-service")
	if err != nil {
		panic(err)
	}
	timestamp, err := core.NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor).WithRecordedAt(timestamp)

	origin, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		panic(err)
	}
	integrity, err := core.NewIntegrityIdentity(
		core.IntegrityMechanismContentAddressedReference,
		"sha256:deadbeef",
		core.IntegrityProtectedScopeContent,
	)
	if err != nil {
		panic(err)
	}

	revisionID, err := core.NewArtifactRevisionID("REV-1")
	if err != nil {
		panic(err)
	}
	revision, err := core.NewArtifactRevision(artifactID, revisionID, origin, provenance, integrity)
	if err != nil {
		panic(err)
	}

	// A typed reference to this exact Revision. This is what other
	// constructs across the SDK cite, rather than holding the Artifact or
	// Revision value directly.
	revisionRef, err := core.NewArtifactRevisionRef(artifact.ID(), revision.RevisionID())
	if err != nil {
		panic(err)
	}

	fmt.Println(artifact.Type().String())
	fmt.Println(revisionRef.IsZero())

	// Output:
	// acme:widget
	// false
}

func mustVocabularyValue(namespace, value string) core.VocabularyValue {
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		panic(err)
	}
	return v
}
