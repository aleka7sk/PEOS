package runtime_test

import (
	"fmt"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/runtime"
)

// This example builds a Runtime Contract, binds a runtime subject to it,
// records one Observation and one Violation derived from it, and unbinds
// the subject — the full PEOS-008 monitoring chain.
func Example_runtimeChain() {
	contractArtifact, err := core.NewArtifact(mustArtifactID("RC-1"), runtime.ArtifactTypeRuntimeContract)
	if err != nil {
		panic(err)
	}
	contract, err := runtime.NewContract(contractArtifact)
	if err != nil {
		panic(err)
	}
	contractRevisionRef, err := core.NewRuntimeContractRevisionRef(contract.Core().ID(), mustArtifactRevisionID("REV-1"))
	if err != nil {
		panic(err)
	}

	runtimeSubject, err := core.NewRuntimeSubjectRef("acme", "svc-1-prod")
	if err != nil {
		panic(err)
	}
	environment := runtime.NewEnvironment(mustVocabularyValue("acme", "production"))

	actor, err := core.NewActorRef("acme", "deploy-service")
	if err != nil {
		panic(err)
	}
	boundAt, err := core.NewTimestamp(time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	provenance := core.NewProvenance().WithActor(actor).WithRecordedAt(boundAt)

	deploymentScope, err := core.NewScope(mustVocabularyValue("acme", "service-boundary"), "SVC-1 production instance")
	if err != nil {
		panic(err)
	}

	binding, err := runtime.NewBindingRecord(
		mustRuntimeBindingRecordID("BIND-1"),
		contractRevisionRef,
		runtimeSubject,
		environment,
		deploymentScope,
		boundAt,
		actor,
		provenance,
	)
	if err != nil {
		panic(err)
	}
	bindingRef, err := binding.Ref()
	if err != nil {
		panic(err)
	}

	observation, err := runtime.NewObservation(
		mustRuntimeObservationID("OBS-1"),
		runtimeSubject,
		deploymentScope,
		environment,
		boundAt,
		"523ms p99 latency",
		"metrics-scrape",
		actor,
		provenance,
	)
	if err != nil {
		panic(err)
	}
	observation, err = observation.WithBinding(bindingRef)
	if err != nil {
		panic(err)
	}
	observationRef, err := observation.Ref()
	if err != nil {
		panic(err)
	}

	requirementRef, err := core.NewRequirementRef(mustArtifactID("REQ-LATENCY-1"))
	if err != nil {
		panic(err)
	}
	criterion, err := core.CriterionRefFromRequirement(requirementRef)
	if err != nil {
		panic(err)
	}

	violation, err := runtime.NewViolationFromObservation(
		mustRuntimeViolationID("VIOL-1"),
		runtimeSubject,
		criterion,
		observationRef,
		boundAt,
		runtime.NewViolationClassification(mustVocabularyValue("acme", "latency-breach")),
		deploymentScope,
		provenance,
	)
	if err != nil {
		panic(err)
	}

	unbinding, err := runtime.NewUnbindingRecord(
		mustRuntimeUnbindingRecordID("UNBIND-1"),
		bindingRef,
		runtimeSubject,
		boundAt,
		"instance decommissioned",
		actor,
		provenance,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(contract.IsZero())
	fmt.Println(observation.Environment().String())
	fmt.Println(violation.Classification().String())
	fmt.Println(unbinding.Reason())

	// Output:
	// false
	// acme:production
	// acme:latency-breach
	// instance decommissioned
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

func mustRuntimeBindingRecordID(value string) core.RuntimeBindingRecordID {
	id, err := core.NewRuntimeBindingRecordID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustRuntimeObservationID(value string) core.RuntimeObservationID {
	id, err := core.NewRuntimeObservationID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustRuntimeViolationID(value string) core.RuntimeViolationID {
	id, err := core.NewRuntimeViolationID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustRuntimeUnbindingRecordID(value string) core.RuntimeUnbindingRecordID {
	id, err := core.NewRuntimeUnbindingRecordID(value)
	if err != nil {
		panic(err)
	}
	return id
}
