package requirement_test

import (
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/requirement"
)

// This example builds a Requirement Artifact and one Revision of it, then
// relates it to a second Requirement Revision via Refinement — one of the
// six PEOS-005 relationship wrappers that compose relation.Relation rather
// than duplicating its fields.
func Example_requirement() {
	artifact, err := core.NewArtifact(mustArtifactID("REQ-1"), requirement.ArtifactTypeRequirement)
	if err != nil {
		panic(err)
	}
	r, err := requirement.New(artifact)
	if err != nil {
		panic(err)
	}

	statement, err := requirement.NewStatement("The system SHALL retry failed downstream calls with exponential backoff.")
	if err != nil {
		panic(err)
	}
	subjectRef, err := core.NewArtifactRef(mustArtifactID("SVC-1"))
	if err != nil {
		panic(err)
	}
	subject, err := core.EngineeringSubjectRefFromArtifact(subjectRef)
	if err != nil {
		panic(err)
	}
	subjectMode, err := requirement.NewSubjectCombination(mustVocabularyValue("peos", "independent"))
	if err != nil {
		panic(err)
	}

	content, err := requirement.NewContent(
		[]requirement.Statement{statement},
		[]core.EngineeringSubjectRef{subject},
		subjectMode,
		requirement.NewUnrestrictedApplicability(),
	)
	if err != nil {
		panic(err)
	}

	revisionID, err := core.NewArtifactRevisionID("REV-1")
	if err != nil {
		panic(err)
	}
	actor, err := core.NewActorRef("acme", "product-team")
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor)
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
	coreRevision, err := core.NewArtifactRevision(artifact.ID(), revisionID, origin, provenance, integrity)
	if err != nil {
		panic(err)
	}

	revision, err := requirement.NewRevision(r, coreRevision, content)
	if err != nil {
		panic(err)
	}

	// A second Requirement Revision this one refines.
	refinedArtifact, err := core.NewArtifact(mustArtifactID("REQ-0"), requirement.ArtifactTypeRequirement)
	if err != nil {
		panic(err)
	}
	refinedRevisionID, err := core.NewArtifactRevisionID("REV-1")
	if err != nil {
		panic(err)
	}
	refinedCoreRevision, err := core.NewArtifactRevision(refinedArtifact.ID(), refinedRevisionID, origin, provenance, integrity)
	if err != nil {
		panic(err)
	}
	refinedRef, err := core.NewRequirementArtifactRevisionRef(refinedArtifact.ID(), refinedCoreRevision.RevisionID())
	if err != nil {
		panic(err)
	}
	refiningRef, err := core.NewRequirementArtifactRevisionRef(artifact.ID(), coreRevision.RevisionID())
	if err != nil {
		panic(err)
	}

	scope, err := core.NewScope(mustVocabularyValue("acme", "reliability"), "retry behaviour only")
	if err != nil {
		panic(err)
	}
	refinement, err := requirement.NewRefinement(refinedRef, refiningRef, provenance, scope)
	if err != nil {
		panic(err)
	}

	fmt.Println(revision.Content().Statements()[0].Text())
	fmt.Println(refinement.Relation().RelationType() == core.RelationTypeRefinement)

	// Output:
	// The system SHALL retry failed downstream calls with exponential backoff.
	// true
}

func mustArtifactID(value string) core.ArtifactID {
	id, err := core.NewArtifactID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustVocabularyValue(namespace, value string) core.VocabularyValue {
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		panic(err)
	}
	return v
}
