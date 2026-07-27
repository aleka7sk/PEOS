package decision_test

import (
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/decision"
)

// This example records a Decision affecting an engineering subject, grounded
// in an evidence-based Basis and an explicit Authority. PEOS-004 requires a
// Decision to identify its subjects, its question, its outcome, its
// applicability, and its authority — all constructor arguments here, so no
// Decision can be built with any of them missing.
func Example_decision() {
	subjectRef, err := core.NewArtifactRef(mustArtifactID("SVC-1"))
	if err != nil {
		panic(err)
	}
	subject, err := core.EngineeringSubjectRefFromArtifact(subjectRef)
	if err != nil {
		panic(err)
	}

	evidenceRef, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID("INCIDENT-42"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}
	basis, err := decision.NewBasis([]core.EvidenceArtifactRevisionRef{evidenceRef})
	if err != nil {
		panic(err)
	}

	outcome, err := decision.NewOutcome(
		"Adopt exponential backoff for downstream retries",
		decision.CommitmentEffectEstablishes,
	)
	if err != nil {
		panic(err)
	}

	basisAuthorityRef, err := core.NewAuthorityRef("acme", "architecture-board")
	if err != nil {
		panic(err)
	}
	authority, err := decision.NewAuthority(nil, []core.AuthorityRef{basisAuthorityRef})
	if err != nil {
		panic(err)
	}

	scope, err := core.NewScope(mustVocabularyValue("acme", "service-boundary"), "SVC-1 retry policy")
	if err != nil {
		panic(err)
	}

	d, err := decision.New(
		mustDecisionID("DEC-1"),
		[]core.EngineeringSubjectRef{subject},
		"How should SVC-1 handle downstream retry storms?",
		outcome,
		scope,
		authority,
	)
	if err != nil {
		panic(err)
	}

	question, _ := d.Question()
	fmt.Println(question)
	fmt.Println(len(basis.Evidence()))

	// Output:
	// How should SVC-1 handle downstream retry storms?
	// 1
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

func mustVocabularyValue(namespace, value string) core.VocabularyValue {
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		panic(err)
	}
	return v
}

func mustDecisionID(value string) core.DecisionID {
	id, err := core.NewDecisionID(value)
	if err != nil {
		panic(err)
	}
	return id
}
