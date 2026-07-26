package core

import (
	"encoding/json"
	"testing"
)

// TestEngineeringSubjectRefAllKnownKindsRoundTrip exercises every branch
// of EngineeringSubjectRef's marshal/unmarshal switch statements, beyond
// the five kinds already covered in detail by
// TestEngineeringSubjectRefKnownVariants. Because those switches are
// large and hand-written, a mismatched case constant in an untested
// branch would otherwise be invisible.
func TestEngineeringSubjectRefAllKnownKindsRoundTrip(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")
	decisionID := mustDecisionID(t, "DEC-1")

	decisionOutcomeRef, err := NewDecisionOutcomeRef(decisionID)
	if err != nil {
		t.Fatal(err)
	}
	engineeringCommitmentRef, err := NewEngineeringCommitmentRef(decisionID)
	if err != nil {
		t.Fatal(err)
	}
	runtimeSubjectRef, err := NewRuntimeSubjectRef("kubernetes", "pod-1")
	if err != nil {
		t.Fatal(err)
	}
	runtimeContractRef, err := NewRuntimeContractRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	runtimeContractRevisionRef, err := NewRuntimeContractRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	templateRef, err := NewTemplateRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	templateRevisionRef, err := NewTemplateArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	generatedArtifactRef, err := NewGeneratedArtifactRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	generatedArtifactRevisionRef, err := NewGeneratedArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	validationPlanRef, err := NewValidationPlanRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	validationPlanRevisionRef, err := NewValidationPlanRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	evidenceRef, err := NewEvidenceArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		kind string
		make func() (EngineeringSubjectRef, error)
		as   func(EngineeringSubjectRef) bool
	}{
		{SubjectKindDecisionOutcome, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromDecisionOutcome(decisionOutcomeRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsDecisionOutcome(); return ok }},
		{SubjectKindEngineeringCommitment, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromEngineeringCommitment(engineeringCommitmentRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsEngineeringCommitment(); return ok }},
		{SubjectKindRuntimeSubject, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromRuntimeSubject(runtimeSubjectRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsRuntimeSubject(); return ok }},
		{SubjectKindRuntimeContract, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromRuntimeContract(runtimeContractRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsRuntimeContract(); return ok }},
		{SubjectKindRuntimeContractRevision, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromRuntimeContractRevision(runtimeContractRevisionRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsRuntimeContractRevision(); return ok }},
		{SubjectKindTemplate, func() (EngineeringSubjectRef, error) { return EngineeringSubjectRefFromTemplate(templateRef) }, func(r EngineeringSubjectRef) bool { _, ok := r.AsTemplate(); return ok }},
		{SubjectKindTemplateRevision, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromTemplateRevision(templateRevisionRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsTemplateRevision(); return ok }},
		{SubjectKindGeneratedArtifact, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromGeneratedArtifact(generatedArtifactRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsGeneratedArtifact(); return ok }},
		{SubjectKindGeneratedArtifactRevision, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromGeneratedArtifactRevision(generatedArtifactRevisionRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsGeneratedArtifactRevision(); return ok }},
		{SubjectKindValidationPlan, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromValidationPlan(validationPlanRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsValidationPlan(); return ok }},
		{SubjectKindValidationPlanRevision, func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromValidationPlanRevision(validationPlanRevisionRef)
		}, func(r EngineeringSubjectRef) bool { _, ok := r.AsValidationPlanRevision(); return ok }},
		{SubjectKindEvidence, func() (EngineeringSubjectRef, error) { return EngineeringSubjectRefFromEvidence(evidenceRef) }, func(r EngineeringSubjectRef) bool { _, ok := r.AsEvidence(); return ok }},
	}

	seenKinds := map[string]bool{
		SubjectKindArtifact:            true,
		SubjectKindArtifactRevision:    true,
		SubjectKindRequirement:         true,
		SubjectKindRequirementRevision: true,
		SubjectKindDecision:            true,
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			seenKinds[tt.kind] = true
			subject, err := tt.make()
			if err != nil {
				t.Fatal(err)
			}
			if subject.Kind() != tt.kind {
				t.Errorf("Kind() = %q, want %q", subject.Kind(), tt.kind)
			}
			if !tt.as(subject) {
				t.Errorf("accessor for kind %q returned ok=false", tt.kind)
			}

			data, err := json.Marshal(subject)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var decoded EngineeringSubjectRef
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !tt.as(decoded) {
				t.Errorf("accessor for round-tripped kind %q returned ok=false", tt.kind)
			}
			data2, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != string(data2) {
				t.Errorf("round trip byte mismatch for kind %q: got %s, want %s", tt.kind, data2, data)
			}
		})
	}

	for kind := range knownSubjectKinds {
		if !seenKinds[kind] {
			t.Errorf("known subject kind %q is not exercised by any test in this package", kind)
		}
	}
}

// TestRecordRefAllKnownKindsRoundTrip exercises every branch of
// RecordRef's marshal/unmarshal switch statements, beyond
// validation_claim and runtime_violation, already covered by
// TestRecordRefKnownKinds.
func TestRecordRefAllKnownKindsRoundTrip(t *testing.T) {
	execID, err := NewValidationExecutionRecordID("EXEC-1")
	if err != nil {
		t.Fatal(err)
	}
	bindingID, err := NewRuntimeBindingRecordID("BIND-1")
	if err != nil {
		t.Fatal(err)
	}
	unbindingID, err := NewRuntimeUnbindingRecordID("UNBIND-1")
	if err != nil {
		t.Fatal(err)
	}
	observationID, err := NewRuntimeObservationID("OBS-1")
	if err != nil {
		t.Fatal(err)
	}
	applicationID, err := NewTemplateApplicationRecordID("APP-1")
	if err != nil {
		t.Fatal(err)
	}
	immutableID, err := NewImmutableRecordID("REC-1")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		kind string
		make func() (RecordRef, error)
		as   func(RecordRef) bool
	}{
		{RecordKindValidationExecutionRecord, func() (RecordRef, error) { return RecordRefFromValidationExecutionRecord(execID) }, func(r RecordRef) bool { _, ok := r.AsValidationExecutionRecord(); return ok }},
		{RecordKindRuntimeBindingRecord, func() (RecordRef, error) { return RecordRefFromRuntimeBindingRecord(bindingID) }, func(r RecordRef) bool { _, ok := r.AsRuntimeBindingRecord(); return ok }},
		{RecordKindRuntimeUnbindingRecord, func() (RecordRef, error) { return RecordRefFromRuntimeUnbindingRecord(unbindingID) }, func(r RecordRef) bool { _, ok := r.AsRuntimeUnbindingRecord(); return ok }},
		{RecordKindRuntimeObservation, func() (RecordRef, error) { return RecordRefFromRuntimeObservation(observationID) }, func(r RecordRef) bool { _, ok := r.AsRuntimeObservation(); return ok }},
		{RecordKindTemplateApplicationRecord, func() (RecordRef, error) { return RecordRefFromTemplateApplicationRecord(applicationID) }, func(r RecordRef) bool { _, ok := r.AsTemplateApplicationRecord(); return ok }},
		{RecordKindImmutableRecord, func() (RecordRef, error) { return RecordRefFromImmutableRecord(immutableID) }, func(r RecordRef) bool { _, ok := r.AsImmutableRecord(); return ok }},
	}

	seenKinds := map[string]bool{
		RecordKindValidationClaim:  true,
		RecordKindRuntimeViolation: true,
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			seenKinds[tt.kind] = true
			ref, err := tt.make()
			if err != nil {
				t.Fatal(err)
			}
			if ref.Kind() != tt.kind {
				t.Errorf("Kind() = %q, want %q", ref.Kind(), tt.kind)
			}
			if !tt.as(ref) {
				t.Errorf("accessor for kind %q returned ok=false", tt.kind)
			}

			data, err := json.Marshal(ref)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var decoded RecordRef
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !tt.as(decoded) {
				t.Errorf("accessor for round-tripped kind %q returned ok=false", tt.kind)
			}
		})
	}

	for kind := range knownRecordKinds {
		if !seenKinds[kind] {
			t.Errorf("known record kind %q is not exercised by any test in this package", kind)
		}
	}
}

// TestCriterionRefAllKnownKindsExercised documents, via the same
// completeness check used above, that every CriterionRef known kind is
// covered somewhere in criterion_test.go.
func TestCriterionRefAllKnownKindsExercised(t *testing.T) {
	exercised := map[string]bool{
		CriterionKindRequirement:           true,
		CriterionKindRequirementRevision:   true,
		CriterionKindArtifact:              true,
		CriterionKindArtifactRevision:      true,
		CriterionKindQualityCharacteristic: true,
		CriterionKindQualityMeasure:        true,
		// The three kinds below were added additively by Packet I.1 for
		// PEOS-007 and are exercised by
		// TestCriterionRefQualityProfileOwnedKinds and
		// TestQualityCompositeCriteriaRequireDedicatedKinds.
		CriterionKindQualityThreshold:    true,
		CriterionKindQualityTarget:       true,
		CriterionKindQualityConstraint:   true,
		CriterionKindRuntimeContractRule: true,
		CriterionKindRuntimeAssertion:    true,
		CriterionKindTemplateConstraint:  true,
		CriterionKindProductRule:         true,
		CriterionKindExternalRule:        true,
	}
	for kind := range knownCriterionKinds {
		if !exercised[kind] {
			t.Errorf("known criterion kind %q is not exercised by any test in this package", kind)
		}
	}
}
