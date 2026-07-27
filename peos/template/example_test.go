package template_test

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/template"
)

// This example builds a Template and Revision, applies it to produce one
// generated Artifact Revision, represents that generation with a
// GeneratedFrom relation, and records a Template Conformance Claim over the
// result. The outcome-conditional invariant is enforced structurally:
// generatedOutputs is a constructor argument, so a succeeded
// ApplicationRecord with no generated output cannot be built at all.
func Example_templateChain() {
	generatedType := core.NewArtifactType(mustVocabularyValue("acme", "handler"))
	declaration, err := template.NewCompatibilityDeclaration(
		[]core.ArtifactType{generatedType},
		"parameters follow semantic-version compatibility",
		"handler routing contract must remain additive across minor versions",
	)
	if err != nil {
		panic(err)
	}

	provenance := core.NewProvenance().WithActor(mustActorRef())

	content, err := template.NewTemplateContent(
		[]core.ArtifactType{generatedType},
		"renders one HTTP handler per declared route",
		declaration,
		template.NewUnrestrictedTemplateApplicability(),
		provenance,
		nil,
		nil,
		nil,
	)
	if err != nil {
		panic(err)
	}

	templateArtifact, err := core.NewArtifact(mustArtifactID("TPL-1"), template.ArtifactTypeTemplate)
	if err != nil {
		panic(err)
	}
	tpl, err := template.NewTemplate(templateArtifact)
	if err != nil {
		panic(err)
	}
	templateCoreRevision, err := core.NewArtifactRevision(
		templateArtifact.ID(),
		mustArtifactRevisionID("REV-1"),
		mustOrigin(),
		provenance,
		mustIntegrityIdentity(),
	)
	if err != nil {
		panic(err)
	}
	templateRevision, err := template.NewTemplateRevision(tpl, templateCoreRevision, content)
	if err != nil {
		panic(err)
	}
	templateRevisionRef, err := core.NewTemplateArtifactRevisionRef(templateArtifact.ID(), templateCoreRevision.RevisionID())
	if err != nil {
		panic(err)
	}

	generatedArtifactRef, err := core.NewGeneratedArtifactRef(mustArtifactID("HANDLER-1"))
	if err != nil {
		panic(err)
	}
	generatedRevisionRef, err := core.NewGeneratedArtifactRevisionRef(mustArtifactID("HANDLER-1"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}
	generatedOutput, err := template.NewGeneratedOutput(generatedArtifactRef, generatedRevisionRef)
	if err != nil {
		panic(err)
	}

	appliedAt, err := core.NewTimestamp(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	applicationEnvironment := mustVocabularyValue("acme", "generation")

	application, err := template.NewApplicationRecord(
		mustTemplateApplicationRecordID("APP-1"),
		templateRevisionRef,
		mustActorRef(),
		appliedAt,
		applicationEnvironment,
		provenance,
		template.ApplicationOutcomeSucceeded,
		nil,
		[]template.GeneratedOutput{generatedOutput},
		nil,
	)
	if err != nil {
		panic(err)
	}

	generationScope, err := core.NewScope(mustVocabularyValue("acme", "generation-scope"), "one handler per route")
	if err != nil {
		panic(err)
	}
	generatedFrom, err := template.NewGeneratedFrom(generatedRevisionRef, templateRevisionRef, generationScope, provenance)
	if err != nil {
		panic(err)
	}

	generatedSubject, err := core.EngineeringSubjectRefFromGeneratedArtifactRevision(generatedRevisionRef)
	if err != nil {
		panic(err)
	}
	conformanceScope, err := core.NewScope(mustVocabularyValue("acme", "conformance-scope"), "generated handler conforms to template")
	if err != nil {
		panic(err)
	}
	genericRevisionRef, err := core.NewArtifactRevisionRef(mustArtifactID("HANDLER-1"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}
	conformanceCriterion, err := core.CriterionRefFromArtifactRevision(genericRevisionRef)
	if err != nil {
		panic(err)
	}

	conformanceEvidence, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID("DIFF-REPORT-1"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}
	conformanceClaim, err := template.NewTemplateConformanceClaim(
		mustValidationClaimID("CONFORMANCE-1"),
		generatedSubject,
		conformanceScope,
		core.ClaimOutcomeSatisfied,
		core.NewValidationMethod(mustVocabularyValue("acme", "template-diff")),
		[]core.CriterionRef{conformanceCriterion},
		[]core.EvidenceArtifactRevisionRef{conformanceEvidence},
		appliedAt,
		provenance,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(templateRevision.IsZero())
	fmt.Println(len(application.GeneratedOutputs()))
	fmt.Println(generatedFrom.Relation().RelationType() == core.RelationTypeGeneratedFrom)
	fmt.Println(conformanceClaim.Outcome().String())

	// Output:
	// false
	// 1
	// true
	// peos:satisfied
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

func mustActorRef() core.ActorRef {
	a, err := core.NewActorRef("acme", "generator-service")
	if err != nil {
		panic(err)
	}
	return a
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

func mustTemplateApplicationRecordID(value string) core.TemplateApplicationRecordID {
	id, err := core.NewTemplateApplicationRecordID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustValidationClaimID(value string) core.ValidationClaimID {
	id, err := core.NewValidationClaimID(value)
	if err != nil {
		panic(err)
	}
	return id
}
