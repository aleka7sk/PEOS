package quality_test

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/quality"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// This example constructs a Quality Profile, a Measurement Record composing
// a validation.ExecutionRecord, and a Quality Claim — the validated wrapper
// PEOS-007 requires over validation.Claim, enforcing PEOS-007's stricter
// reflexive rule that validation.Claim alone does not.
func Example_qualityChain() {
	profileArtifact, err := core.NewArtifact(mustArtifactID("QP-1"), quality.ArtifactTypeQualityProfile)
	if err != nil {
		panic(err)
	}
	profile, err := quality.NewProfile(profileArtifact)
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

	method := core.NewValidationMethod(mustVocabularyValue("acme", "latency-benchmark"))
	activity, err := validation.NewAdHocActivityReference("p99-latency-run")
	if err != nil {
		panic(err)
	}
	actor, err := core.NewActorRef("acme", "perf-service")
	if err != nil {
		panic(err)
	}
	completedAt, err := core.NewTimestamp(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor).WithRecordedAt(completedAt)

	execution, err := validation.NewExecutionRecord(
		mustValidationExecutionRecordID("VER-2"),
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

	// A MeasurementRecord's underlying ExecutionRecord must cite the Quality
	// Characteristic being measured, named by its owning Profile Revision
	// plus a local key.
	profileRevisionRef, err := core.NewArtifactRevisionRef(profileArtifact.ID(), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}
	characteristicKey, err := core.NewLocalKey("p99-latency")
	if err != nil {
		panic(err)
	}
	characteristicElement, err := core.NewQualityElementCriterionRef(profileRevisionRef, characteristicKey)
	if err != nil {
		panic(err)
	}
	characteristicCriterion, err := core.CriterionRefFromQualityCharacteristic(characteristicElement)
	if err != nil {
		panic(err)
	}
	measureKey, err := core.NewLocalKey("p99-latency-measure")
	if err != nil {
		panic(err)
	}
	measureElement, err := core.NewQualityElementCriterionRef(profileRevisionRef, measureKey)
	if err != nil {
		panic(err)
	}
	measureCriterion, err := core.CriterionRefFromQualityMeasure(measureElement)
	if err != nil {
		panic(err)
	}
	execution, err = execution.WithCriteria([]core.CriterionRef{characteristicCriterion, measureCriterion})
	if err != nil {
		panic(err)
	}

	measurement, err := quality.NewMeasurementRecord(
		execution,
		"142",
		quality.NewUnit(mustVocabularyValue("acme", "milliseconds")),
		quality.NewScale(mustVocabularyValue("acme", "ratio")),
	)
	if err != nil {
		panic(err)
	}

	scope, err := core.NewScope(mustVocabularyValue("acme", "service-boundary"), "SVC-1 latency")
	if err != nil {
		panic(err)
	}
	evidenceRef, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID("BENCHMARK-LOG-1"), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}
	qualityClaim, err := quality.NewClaim(
		mustValidationClaimID("QCLAIM-1"),
		subject,
		scope,
		core.ClaimOutcomeSatisfied,
		method,
		[]core.CriterionRef{characteristicCriterion},
		[]core.EvidenceArtifactRevisionRef{evidenceRef},
		completedAt,
		provenance,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(profile.IsZero())
	fmt.Println(measurement.ObservedValue())
	fmt.Println(qualityClaim.Outcome().String())

	// Output:
	// false
	// 142
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
