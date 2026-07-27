package validation_test

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// This example constructs a Validation Plan, an Execution Record against an
// ad hoc activity, and a Validation Claim citing one Evidence reference —
// the PEOS-006 chain: Plan, Execution Record, Evidence, Result, Claim. The
// "Result" is the Claim's Outcome; PEOS-006 states directly that "There is
// no separate Verdict entity."
func Example_validationChain() {
	planArtifact, err := core.NewArtifact(mustArtifactID("VP-1"), validation.ArtifactTypeValidationPlan)
	if err != nil {
		panic(err)
	}
	plan, err := validation.NewPlan(planArtifact)
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

	method := core.NewValidationMethod(mustVocabularyValue("acme", "integration-test"))
	activity, err := validation.NewAdHocActivityReference("retry-storm-drill")
	if err != nil {
		panic(err)
	}

	actor, err := core.NewActorRef("acme", "qa-service")
	if err != nil {
		panic(err)
	}
	completedAt, err := core.NewTimestamp(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor).WithRecordedAt(completedAt)

	execution, err := validation.NewExecutionRecord(
		mustValidationExecutionRecordID("VER-1"),
		activity,
		subject,
		method,
		core.ExecutionOutcomeCompleted,
		completedAt,
		actor,
		provenance,
	)
	if err != nil {
		panic(err)
	}

	evidenceRef, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID("DRILL-LOG-1"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}

	scope, err := core.NewScope(mustVocabularyValue("acme", "service-boundary"), "SVC-1 retry policy")
	if err != nil {
		panic(err)
	}

	// A Satisfaction Claim's criteria SHALL identify one or more
	// Requirements (PEOS-006). The Requirement itself lives in
	// peos/requirement; peos/validation never imports that package, so the
	// citation travels through core.CriterionRef instead — the same
	// composition pattern the full cross-package example demonstrates.
	requirementRef, err := core.NewRequirementRef(mustArtifactID("REQ-1"))
	if err != nil {
		panic(err)
	}
	criterion, err := core.CriterionRefFromRequirement(requirementRef)
	if err != nil {
		panic(err)
	}

	claim, err := validation.NewClaim(
		mustValidationClaimID("CLAIM-1"),
		core.ClaimTypeSatisfaction,
		subject,
		scope,
		core.ClaimOutcomeSatisfied, // the Result: PEOS-006 has no separate Verdict entity
		method,
		[]core.CriterionRef{criterion},
		[]core.EvidenceArtifactRevisionRef{evidenceRef},
		completedAt,
		provenance,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(plan.IsZero())
	fmt.Println(execution.Outcome().String())
	fmt.Println(claim.Outcome().String())

	// Output:
	// false
	// peos:completed
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

func mustValidationExecutionRecordID(value string) core.ValidationExecutionRecordID {
	id, err := core.NewValidationExecutionRecordID(value)
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
