// Package crosspackage contains one worked, compiling example of the SDK's
// most important composition pattern: connecting a Requirement (peos/requirement)
// to a Validation Claim (peos/validation) without either package importing the
// other. See docs/consumer-guide.md section 7 for the narrative explanation;
// this file is the executable proof of that explanation.
package crosspackage

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/requirement"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// Example_requirementToValidationClaim builds a Requirement, then a
// Validation Plan, Execution Record, and Claim that evaluate it — the full
// PEOS-005-to-PEOS-006 chain: Requirement -> Validation Plan -> Execution
// Record -> Evidence -> Result -> Validation Claim.
//
// The composition boundary is the point of the example. peos/validation's
// own architecture forbids it from importing peos/requirement (enforced by
// peos/validation/doc_test.go's TestProductionImportBoundary and
// peos/requirement/doc_test.go's converse TestValidationDoesNotImportRequirement).
// A Requirement therefore never appears as a Go value inside peos/validation.
// Instead:
//
//   - the Requirement becomes a core.EngineeringSubjectRef via
//     core.EngineeringSubjectRefFromRequirement, when something needs to be
//     evaluated as-of an identity-level Requirement;
//   - the Requirement becomes a core.CriterionRef via
//     core.CriterionRefFromRequirement, when it is cited as evaluation
//     criteria rather than as the thing being evaluated.
//
// Both conversions happen here, in the consumer's own code -- the one place
// that is allowed to import both packages -- never inside peos/validation
// itself. No shadow struct duplicating Requirement's fields is needed:
// core.RequirementRef is the only thing that crosses the boundary, and it is
// already exactly what peos/validation's own union types accept.
//
// Repository lookup (resolving a core.RequirementRef back to a full
// requirement.Requirement, or finding "the current Claim for this subject")
// is deliberately absent from this example: PEOS assigns that to a consumer
// repository, not to the value model. See docs/consumer-guide.md section 6.
func Example_requirementToValidationClaim() {
	// --- 1. Build the Requirement Artifact and its Revision (peos/requirement) ---

	requirementArtifact, err := core.NewArtifact(mustArtifactID("REQ-1"), requirement.ArtifactTypeRequirement)
	if err != nil {
		panic(err)
	}
	req, err := requirement.New(requirementArtifact)
	if err != nil {
		panic(err)
	}

	statement, err := requirement.NewStatement("The system SHALL retry failed downstream calls with exponential backoff.")
	if err != nil {
		panic(err)
	}
	serviceArtifactRef, err := core.NewArtifactRef(mustArtifactID("SVC-1"))
	if err != nil {
		panic(err)
	}
	serviceSubject, err := core.EngineeringSubjectRefFromArtifact(serviceArtifactRef)
	if err != nil {
		panic(err)
	}
	subjectMode, err := requirement.NewSubjectCombination(mustVocabularyValue("peos", "independent"))
	if err != nil {
		panic(err)
	}
	content, err := requirement.NewContent(
		[]requirement.Statement{statement},
		[]core.EngineeringSubjectRef{serviceSubject},
		subjectMode,
		requirement.NewUnrestrictedApplicability(),
	)
	if err != nil {
		panic(err)
	}

	actor, err := core.NewActorRef("acme", "product-team")
	if err != nil {
		panic(err)
	}
	recordedAt, err := core.NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor).WithRecordedAt(recordedAt)
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
	requirementCoreRevision, err := core.NewArtifactRevision(
		requirementArtifact.ID(),
		mustArtifactRevisionID("REV-1"),
		origin,
		provenance,
		integrity,
	)
	if err != nil {
		panic(err)
	}
	requirementRevision, err := requirement.NewRevision(req, requirementCoreRevision, content)
	if err != nil {
		panic(err)
	}

	// --- 2. Cross the package boundary: Requirement -> core reference unions ---
	//
	// peos/validation never sees a requirement.Requirement value. It sees
	// only these two core types, which it already knows how to accept.

	requirementRef, err := core.NewRequirementRef(requirementArtifact.ID())
	if err != nil {
		panic(err)
	}

	// As the Claim Subject (what is being evaluated), when a Satisfaction
	// Claim is scoped to the Requirement's own identity rather than to the
	// implementing Artifact.
	requirementAsSubject, err := core.EngineeringSubjectRefFromRequirement(requirementRef)
	if err != nil {
		panic(err)
	}

	// As Claim criteria (what the Claim's actual subject is evaluated
	// against) -- the shape used below, since PEOS-006 requires a
	// Satisfaction Claim's criteria to identify the Requirement.
	requirementAsCriterion, err := core.CriterionRefFromRequirement(requirementRef)
	if err != nil {
		panic(err)
	}

	// --- 3. Validation Plan, Execution Record, Evidence, Claim (peos/validation) ---

	planArtifact, err := core.NewArtifact(mustArtifactID("VP-1"), validation.ArtifactTypeValidationPlan)
	if err != nil {
		panic(err)
	}
	plan, err := validation.NewPlan(planArtifact)
	if err != nil {
		panic(err)
	}

	method := core.NewValidationMethod(mustVocabularyValue("acme", "integration-test"))
	activity, err := validation.NewAdHocActivityReference("retry-storm-drill")
	if err != nil {
		panic(err)
	}
	qaActor, err := core.NewActorRef("acme", "qa-service")
	if err != nil {
		panic(err)
	}
	completedAt, err := core.NewTimestamp(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	executionProvenance := core.NewProvenance().WithActor(qaActor).WithRecordedAt(completedAt)

	execution, err := validation.NewExecutionRecord(
		mustValidationExecutionRecordID("VER-1"),
		activity,
		serviceSubject, // the Artifact whose behaviour was actually exercised
		method,
		core.ExecutionOutcomeCompleted,
		completedAt,
		qaActor,
		executionProvenance,
	)
	if err != nil {
		panic(err)
	}

	// Evidence: a fixed Artifact Revision the Execution Record's result rests on.
	evidenceRef, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID("DRILL-LOG-1"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}

	scope, err := core.NewScope(mustVocabularyValue("acme", "service-boundary"), "SVC-1 retry policy")
	if err != nil {
		panic(err)
	}

	// Result: the Claim's Outcome. PEOS-006 states directly that "There is
	// no separate Verdict entity" -- the outcome recorded here is the
	// complete and only representation of what was determined.
	claim, err := validation.NewClaim(
		mustValidationClaimID("CLAIM-1"),
		core.ClaimTypeSatisfaction,
		serviceSubject, // the Claim's Subject: what satisfies the Requirement
		scope,
		core.ClaimOutcomeSatisfied,
		method,
		[]core.CriterionRef{requirementAsCriterion}, // the Requirement, cited as criterion
		[]core.EvidenceArtifactRevisionRef{evidenceRef},
		completedAt,
		executionProvenance,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(requirementRevision.Content().Statements()[0].Text())
	fmt.Println(plan.IsZero())
	fmt.Println(execution.Outcome().String())
	fmt.Println(requirementAsSubject.Kind())
	fmt.Println(claim.Outcome().String())

	// Output:
	// The system SHALL retry failed downstream calls with exponential backoff.
	// false
	// peos:completed
	// requirement
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
