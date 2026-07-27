package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func mustArtifactID(t *testing.T, v string) ArtifactID {
	t.Helper()
	id, err := NewArtifactID(v)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustArtifactRevisionID(t *testing.T, v string) ArtifactRevisionID {
	t.Helper()
	id, err := NewArtifactRevisionID(v)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustDecisionID(t *testing.T, v string) DecisionID {
	t.Helper()
	id, err := NewDecisionID(v)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// --- identity-level reference tests -----------------------------------

func TestArtifactRef(t *testing.T) {
	id := mustArtifactID(t, "ART-1")
	ref, err := NewArtifactRef(id)
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsZero() {
		t.Error("valid ArtifactRef reports IsZero() = true")
	}
	if ref.ArtifactID() != id {
		t.Errorf("ArtifactID() = %v, want %v", ref.ArtifactID(), id)
	}

	if _, err := NewArtifactRef(ArtifactID{}); err == nil {
		t.Error("zero ArtifactID accepted, want error")
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"artifact_id":"ART-1"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded ArtifactRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}
}

func TestArtifactRevisionRef(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")

	ref, err := NewArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
		t.Errorf("got (%v, %v), want (%v, %v)", ref.ArtifactID(), ref.RevisionID(), artifactID, revisionID)
	}

	if _, err := NewArtifactRevisionRef(ArtifactID{}, revisionID); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("missing artifact id: error = %v, want %v", err, ErrEmptyIdentity)
	}
	if _, err := NewArtifactRevisionRef(artifactID, ArtifactRevisionID{}); !errors.Is(err, ErrMissingRevisionID) {
		t.Errorf("missing revision id: error = %v, want %v", err, ErrMissingRevisionID)
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ArtifactRevisionRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}

	// An identity-level reference and a Revision-level reference are
	// structurally distinct types; ArtifactRef has no field to carry a
	// revision id at all, so "unexpected revision id" is a compile-time
	// impossibility rather than a runtime check for this pair.
	var _ = func() ArtifactRef {
		artifactOnly, err := NewArtifactRef(artifactID)
		if err != nil {
			t.Fatal(err)
		}
		return artifactOnly
	}()
}

func TestArtifactRevisionRefJSONRejectsMissingRevisionID(t *testing.T) {
	var ref ArtifactRevisionRef
	// revision_id is omitted entirely (as opposed to present-but-empty):
	// encoding/json leaves the field at its zero value without invoking
	// ArtifactRevisionID's own UnmarshalJSON, so this exercises
	// NewArtifactRevisionRef's own ErrMissingRevisionID check rather than
	// ArtifactRevisionID's ErrEmptyIdentity check.
	err := json.Unmarshal([]byte(`{"artifact_id":"ART-1"}`), &ref)
	if !errors.Is(err, ErrMissingRevisionID) {
		t.Errorf("error = %v, want %v", err, ErrMissingRevisionID)
	}

	// A present-but-empty revision_id is also rejected, one level lower
	// (by ArtifactRevisionID itself).
	err = json.Unmarshal([]byte(`{"artifact_id":"ART-1","revision_id":""}`), &ref)
	if !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}
}

func TestRequirementRefDistinctFromArtifactRef(t *testing.T) {
	id := mustArtifactID(t, "REQ-1")
	reqRef, err := NewRequirementRef(id)
	if err != nil {
		t.Fatal(err)
	}
	artRef, err := NewArtifactRef(id)
	if err != nil {
		t.Fatal(err)
	}
	// The following, if uncommented, must fail to compile because the two
	// types have different field names and are therefore not identical
	// underlying types:
	//   _ = RequirementRef(artRef)
	if reqRef.ArtifactID() != artRef.ArtifactID() {
		t.Error("underlying ArtifactID differs unexpectedly")
	}
}

func TestDecisionRefAndDependents(t *testing.T) {
	id := mustDecisionID(t, "DEC-1")

	decisionRef, err := NewDecisionRef(id)
	if err != nil {
		t.Fatal(err)
	}
	outcomeRef, err := NewDecisionOutcomeRef(id)
	if err != nil {
		t.Fatal(err)
	}
	commitmentRef, err := NewEngineeringCommitmentRef(id)
	if err != nil {
		t.Fatal(err)
	}

	if decisionRef.DecisionID() != id || outcomeRef.DecisionID() != id || commitmentRef.DecisionID() != id {
		t.Error("DecisionID() did not round-trip through Decision-family refs")
	}

	if _, err := NewDecisionRef(DecisionID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}
	if _, err := NewDecisionOutcomeRef(DecisionID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}
	if _, err := NewEngineeringCommitmentRef(DecisionID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}
}

func TestRuntimeSubjectRefIsNotAnArtifactRef(t *testing.T) {
	ref, err := NewRuntimeSubjectRef("kubernetes", "pod-abc123")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Namespace() != "kubernetes" || ref.Identifier() != "pod-abc123" {
		t.Errorf("got (%q, %q)", ref.Namespace(), ref.Identifier())
	}
	if _, err := NewRuntimeSubjectRef("", "pod-abc123"); err == nil {
		t.Error("empty namespace accepted, want error")
	}
	if _, err := NewRuntimeSubjectRef("kubernetes", ""); err == nil {
		t.Error("empty identifier accepted, want error")
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RuntimeSubjectRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}
}

func TestValidationClaimRefAndExecutionRecordRef(t *testing.T) {
	claimID, err := NewValidationClaimID("CLAIM-1")
	if err != nil {
		t.Fatal(err)
	}
	claimRef, err := NewValidationClaimRef(claimID)
	if err != nil {
		t.Fatal(err)
	}
	if claimRef.ClaimID() != claimID {
		t.Error("ClaimID() mismatch")
	}

	recordID, err := NewValidationExecutionRecordID("EXEC-1")
	if err != nil {
		t.Fatal(err)
	}
	recordRef, err := NewValidationExecutionRecordRef(recordID)
	if err != nil {
		t.Fatal(err)
	}
	if recordRef.RecordID() != recordID {
		t.Error("RecordID() mismatch")
	}
}

func TestValidationExecutionRecordRef(t *testing.T) {
	if _, err := NewValidationExecutionRecordRef(ValidationExecutionRecordID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty identity: error = %v, want %v", err, ErrEmptyIdentity)
	}

	recordID, err := NewValidationExecutionRecordID("EXEC-2")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := NewValidationExecutionRecordRef(recordID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsZero() {
		t.Error("valid ValidationExecutionRecordRef reports IsZero() = true")
	}
	var zero ValidationExecutionRecordRef
	if !zero.IsZero() {
		t.Error("zero-value ValidationExecutionRecordRef.IsZero() = false, want true")
	}
	if ref.RecordID() != recordID {
		t.Error("RecordID() mismatch")
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"record_id":"EXEC-2"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded ValidationExecutionRecordRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}
}

// revisionRefFamily captures the constructor and JSON field names shared
// by every "owning Artifact + exact Revision" reference type, so the
// remaining six families (beyond ArtifactRevisionRef, tested above in
// detail) can share one table-driven test.
func TestRevisionLevelReferenceFamilies(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")

	t.Run("RequirementArtifactRevisionRef", func(t *testing.T) {
		ref, err := NewRequirementArtifactRevisionRef(artifactID, revisionID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
			t.Error("component mismatch")
		}
		if _, err := NewRequirementArtifactRevisionRef(artifactID, ArtifactRevisionID{}); !errors.Is(err, ErrMissingRevisionID) {
			t.Errorf("error = %v, want %v", err, ErrMissingRevisionID)
		}
		roundTripJSON(t, &ref, &RequirementArtifactRevisionRef{})
	})

	t.Run("RuntimeContractRevisionRef", func(t *testing.T) {
		ref, err := NewRuntimeContractRevisionRef(artifactID, revisionID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
			t.Error("component mismatch")
		}
		if _, err := NewRuntimeContractRevisionRef(ArtifactID{}, revisionID); !errors.Is(err, ErrEmptyIdentity) {
			t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
		}
		roundTripJSON(t, &ref, &RuntimeContractRevisionRef{})
	})

	t.Run("TemplateArtifactRevisionRef", func(t *testing.T) {
		ref, err := NewTemplateArtifactRevisionRef(artifactID, revisionID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
			t.Error("component mismatch")
		}
		roundTripJSON(t, &ref, &TemplateArtifactRevisionRef{})
	})

	t.Run("GeneratedArtifactRevisionRef", func(t *testing.T) {
		ref, err := NewGeneratedArtifactRevisionRef(artifactID, revisionID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
			t.Error("component mismatch")
		}
		roundTripJSON(t, &ref, &GeneratedArtifactRevisionRef{})
	})

	t.Run("ValidationPlanRevisionRef", func(t *testing.T) {
		ref, err := NewValidationPlanRevisionRef(artifactID, revisionID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
			t.Error("component mismatch")
		}
		roundTripJSON(t, &ref, &ValidationPlanRevisionRef{})
	})

	t.Run("EvidenceArtifactRevisionRef", func(t *testing.T) {
		ref, err := NewEvidenceArtifactRevisionRef(artifactID, revisionID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID || ref.RevisionID() != revisionID {
			t.Error("component mismatch")
		}
		if _, err := NewEvidenceArtifactRevisionRef(artifactID, ArtifactRevisionID{}); !errors.Is(err, ErrMissingRevisionID) {
			t.Errorf("Evidence must always be exact-revision: error = %v, want %v", err, ErrMissingRevisionID)
		}
		roundTripJSON(t, &ref, &EvidenceArtifactRevisionRef{})
	})
}

// identityLevelRefFamily exercises the remaining single-ArtifactID
// identity-level reference types beyond ArtifactRef, tested in detail
// above.
func TestIdentityLevelReferenceFamilies(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")

	t.Run("RuntimeContractRef", func(t *testing.T) {
		ref, err := NewRuntimeContractRef(artifactID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.ArtifactID() != artifactID {
			t.Error("component mismatch")
		}
		roundTripJSON(t, &ref, &RuntimeContractRef{})
	})

	t.Run("TemplateRef", func(t *testing.T) {
		ref, err := NewTemplateRef(artifactID)
		if err != nil {
			t.Fatal(err)
		}
		roundTripJSON(t, &ref, &TemplateRef{})
	})

	t.Run("GeneratedArtifactRef", func(t *testing.T) {
		ref, err := NewGeneratedArtifactRef(artifactID)
		if err != nil {
			t.Fatal(err)
		}
		roundTripJSON(t, &ref, &GeneratedArtifactRef{})
	})

	t.Run("ValidationPlanRef", func(t *testing.T) {
		ref, err := NewValidationPlanRef(artifactID)
		if err != nil {
			t.Fatal(err)
		}
		roundTripJSON(t, &ref, &ValidationPlanRef{})
	})
}

// --- Packet E: LifecycleDefinitionRef, LifecycleDefinitionVersionRef, StateAssignmentRef ---

func mustLifecycleDefinitionID(t *testing.T, value string) LifecycleDefinitionID {
	t.Helper()
	id, err := NewLifecycleDefinitionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestLifecycleDefinitionRef(t *testing.T) {
	defID := mustLifecycleDefinitionID(t, "LC-REVIEW-1")

	ref, err := NewLifecycleDefinitionRef(defID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.LifecycleDefinitionID() != defID {
		t.Error("component mismatch")
	}
	if ref.IsZero() {
		t.Error("valid LifecycleDefinitionRef reports IsZero() = true")
	}

	if _, err := NewLifecycleDefinitionRef(LifecycleDefinitionID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}

	var zero LifecycleDefinitionRef
	if !zero.IsZero() {
		t.Error("zero-value LifecycleDefinitionRef.IsZero() = false, want true")
	}

	roundTripJSON(t, &ref, &LifecycleDefinitionRef{})

	if err := json.Unmarshal([]byte(`{"lifecycle_definition_id": ""}`), &LifecycleDefinitionRef{}); err == nil {
		t.Fatal("malformed (empty id) JSON accepted, want error")
	}

	receiver := ref
	if err := json.Unmarshal([]byte(`{"lifecycle_definition_id": ""}`), &receiver); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
	if receiver != ref {
		t.Errorf("failed Unmarshal changed receiver: got %v, want %v", receiver, ref)
	}
}

func TestLifecycleDefinitionVersionRef(t *testing.T) {
	defID := mustLifecycleDefinitionID(t, "LC-REVIEW-1")
	versionID, err := NewLifecycleDefinitionVersionID("V1")
	if err != nil {
		t.Fatal(err)
	}

	ref, err := NewLifecycleDefinitionVersionRef(defID, versionID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.LifecycleDefinitionID() != defID || ref.VersionID() != versionID {
		t.Error("component mismatch")
	}
	if ref.IsZero() {
		t.Error("valid LifecycleDefinitionVersionRef reports IsZero() = true")
	}

	if _, err := NewLifecycleDefinitionVersionRef(LifecycleDefinitionID{}, versionID); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}
	if _, err := NewLifecycleDefinitionVersionRef(defID, LifecycleDefinitionVersionID{}); !errors.Is(err, ErrMissingRevisionID) {
		t.Errorf("error = %v, want %v", err, ErrMissingRevisionID)
	}

	var zero LifecycleDefinitionVersionRef
	if !zero.IsZero() {
		t.Error("zero-value LifecycleDefinitionVersionRef.IsZero() = false, want true")
	}

	roundTripJSON(t, &ref, &LifecycleDefinitionVersionRef{})

	receiver := ref
	if err := json.Unmarshal([]byte(`{"lifecycle_definition_id": "LC-REVIEW-1"}`), &receiver); err == nil {
		t.Fatal("JSON missing version id accepted, want error")
	}
	if receiver != ref {
		t.Errorf("failed Unmarshal changed receiver: got %v, want %v", receiver, ref)
	}
}

func TestStateAssignmentRef(t *testing.T) {
	assignmentID, err := NewStateAssignmentID("SA-1001")
	if err != nil {
		t.Fatal(err)
	}

	ref, err := NewStateAssignmentRef(assignmentID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.StateAssignmentID() != assignmentID {
		t.Error("component mismatch")
	}
	if ref.IsZero() {
		t.Error("valid StateAssignmentRef reports IsZero() = true")
	}

	if _, err := NewStateAssignmentRef(StateAssignmentID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}

	var zero StateAssignmentRef
	if !zero.IsZero() {
		t.Error("zero-value StateAssignmentRef.IsZero() = false, want true")
	}

	roundTripJSON(t, &ref, &StateAssignmentRef{})

	receiver := ref
	if err := json.Unmarshal([]byte(`{"state_assignment_id": ""}`), &receiver); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
	if receiver != ref {
		t.Errorf("failed Unmarshal changed receiver: got %v, want %v", receiver, ref)
	}
}

// TestLifecycleReferenceTypesAreNotInterchangeable documents that
// LifecycleDefinitionRef, LifecycleDefinitionVersionRef, and
// StateAssignmentRef are structurally distinct from each other and from
// ArtifactRef/DecisionRef, exactly like every other reference family in
// this file.
func TestLifecycleReferenceTypesAreNotInterchangeable(t *testing.T) {
	defRef, err := NewLifecycleDefinitionRef(mustLifecycleDefinitionID(t, "LC-1"))
	if err != nil {
		t.Fatal(err)
	}
	assignmentRef, err := NewStateAssignmentRef(func() StateAssignmentID {
		id, err := NewStateAssignmentID("SA-1")
		if err != nil {
			t.Fatal(err)
		}
		return id
	}())
	if err != nil {
		t.Fatal(err)
	}
	// The following, if uncommented, must fail to compile:
	//   var _ LifecycleDefinitionRef = assignmentRef
	//   var _ StateAssignmentRef = defRef
	if defRef.LifecycleDefinitionID().String() == assignmentRef.StateAssignmentID().String() {
		t.Skip("identical opaque values chosen; not a meaningful collision")
	}
}

// --- Packet E.1: LifecycleDefinitionVersionSupersessionRef ---

func TestLifecycleDefinitionVersionSupersessionRef(t *testing.T) {
	supID, err := NewLifecycleDefinitionVersionSupersessionID("SUP-1001")
	if err != nil {
		t.Fatal(err)
	}

	ref, err := NewLifecycleDefinitionVersionSupersessionRef(supID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.SupersessionID() != supID {
		t.Error("component mismatch")
	}
	if ref.IsZero() {
		t.Error("valid LifecycleDefinitionVersionSupersessionRef reports IsZero() = true")
	}

	if _, err := NewLifecycleDefinitionVersionSupersessionRef(LifecycleDefinitionVersionSupersessionID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}

	var zero LifecycleDefinitionVersionSupersessionRef
	if !zero.IsZero() {
		t.Error("zero-value LifecycleDefinitionVersionSupersessionRef.IsZero() = false, want true")
	}

	roundTripJSON(t, &ref, &LifecycleDefinitionVersionSupersessionRef{})

	receiver := ref
	if err := json.Unmarshal([]byte(`{"lifecycle_definition_version_supersession_id": ""}`), &receiver); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
	if receiver != ref {
		t.Errorf("failed Unmarshal changed receiver: got %v, want %v", receiver, ref)
	}
}

func TestLifecycleDefinitionVersionSupersessionRefNotInterchangeable(t *testing.T) {
	supID, err := NewLifecycleDefinitionVersionSupersessionID("SUP-1")
	if err != nil {
		t.Fatal(err)
	}
	supRef, err := NewLifecycleDefinitionVersionSupersessionRef(supID)
	if err != nil {
		t.Fatal(err)
	}
	assignmentRef, err := NewStateAssignmentRef(func() StateAssignmentID {
		id, err := NewStateAssignmentID("SA-1")
		if err != nil {
			t.Fatal(err)
		}
		return id
	}())
	if err != nil {
		t.Fatal(err)
	}
	// The following, if uncommented, must fail to compile:
	//   var _ LifecycleDefinitionVersionSupersessionRef = assignmentRef
	//   var _ StateAssignmentRef = supRef
	if supRef.SupersessionID().String() == assignmentRef.StateAssignmentID().String() {
		t.Skip("identical opaque values chosen; not a meaningful collision")
	}
}

// roundTripJSON marshals src, unmarshals into dst, and fails the test if
// the two do not marshal to the same bytes (used for types without a
// convenient equality check exposed).
func roundTripJSON(t *testing.T, src, dst interface {
	json.Marshaler
	json.Unmarshaler
}) {
	t.Helper()
	data, err := src.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if err := dst.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	data2, err := dst.MarshalJSON()
	if err != nil {
		t.Fatalf("re-MarshalJSON: %v", err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip mismatch: got %s, want %s", data2, data)
	}
}

// --- EngineeringSubjectRef ----------------------------------------------

func TestEngineeringSubjectRefKnownVariants(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")
	decisionID := mustDecisionID(t, "DEC-1")

	artifactRef, err := NewArtifactRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	artifactRevisionRef, err := NewArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	requirementRef, err := NewRequirementRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	requirementRevisionRef, err := NewRequirementArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	decisionRef, err := NewDecisionRef(decisionID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		make func() (EngineeringSubjectRef, error)
		kind string
		as   func(EngineeringSubjectRef) bool
	}{
		{"artifact", func() (EngineeringSubjectRef, error) { return EngineeringSubjectRefFromArtifact(artifactRef) }, SubjectKindArtifact, func(r EngineeringSubjectRef) bool { _, ok := r.AsArtifact(); return ok }},
		{"artifact_revision", func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromArtifactRevision(artifactRevisionRef)
		}, SubjectKindArtifactRevision, func(r EngineeringSubjectRef) bool { _, ok := r.AsArtifactRevision(); return ok }},
		{"requirement", func() (EngineeringSubjectRef, error) { return EngineeringSubjectRefFromRequirement(requirementRef) }, SubjectKindRequirement, func(r EngineeringSubjectRef) bool { _, ok := r.AsRequirement(); return ok }},
		{"requirement_revision", func() (EngineeringSubjectRef, error) {
			return EngineeringSubjectRefFromRequirementRevision(requirementRevisionRef)
		}, SubjectKindRequirementRevision, func(r EngineeringSubjectRef) bool { _, ok := r.AsRequirementRevision(); return ok }},
		{"decision", func() (EngineeringSubjectRef, error) { return EngineeringSubjectRefFromDecision(decisionRef) }, SubjectKindDecision, func(r EngineeringSubjectRef) bool { _, ok := r.AsDecision(); return ok }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, err := tt.make()
			if err != nil {
				t.Fatal(err)
			}
			if subject.Kind() != tt.kind {
				t.Errorf("Kind() = %q, want %q", subject.Kind(), tt.kind)
			}
			if !subject.IsKnown() {
				t.Error("IsKnown() = false for a typed constructor result")
			}
			if !tt.as(subject) {
				t.Errorf("accessor for kind %q returned ok=false", tt.kind)
			}

			// Cross-kind accessors must not match.
			if tt.kind != SubjectKindArtifact {
				if _, ok := subject.AsArtifact(); ok {
					t.Error("AsArtifact() ok=true for a non-artifact subject")
				}
			}

			data, err := json.Marshal(subject)
			if err != nil {
				t.Fatal(err)
			}
			var decoded EngineeringSubjectRef
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != subject.Kind() {
				t.Errorf("round trip Kind() = %q, want %q", decoded.Kind(), subject.Kind())
			}
			data2, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != string(data2) {
				t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
			}
		})
	}
}

func TestEngineeringSubjectRefInvalidDiscriminator(t *testing.T) {
	if err := json.Unmarshal([]byte(`{"kind":"","ref":{}}`), &EngineeringSubjectRef{}); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("empty kind: error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
}

func TestEngineeringSubjectRefPayloadMismatch(t *testing.T) {
	// A "kind":"artifact" envelope whose "ref" does not contain a valid
	// artifact_id must fail to unmarshal, not silently succeed with a
	// zero-value payload.
	err := json.Unmarshal([]byte(`{"kind":"artifact","ref":{"artifact_id":""}}`), &EngineeringSubjectRef{})
	if err == nil {
		t.Error("payload mismatch accepted, want error")
	}
}

func TestEngineeringSubjectRefConstructorRejectsZeroPayload(t *testing.T) {
	if _, err := EngineeringSubjectRefFromArtifact(ArtifactRef{}); !errors.Is(err, ErrInvalidPayload) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPayload)
	}
}

func TestEngineeringSubjectRefOpaqueUnknownKind(t *testing.T) {
	subject, err := NewOpaqueEngineeringSubjectRef("future-kind", "product-x", "thing-1")
	if err != nil {
		t.Fatal(err)
	}
	if subject.IsKnown() {
		t.Error("IsKnown() = true for opaque subject")
	}
	opaque, ok := subject.AsOpaque()
	if !ok {
		t.Fatal("AsOpaque() ok = false")
	}
	if opaque.Kind() != "future-kind" || opaque.Namespace() != "product-x" || opaque.Identifier() != "thing-1" {
		t.Errorf("opaque payload mismatch: %+v", opaque)
	}

	data, err := json.Marshal(subject)
	if err != nil {
		t.Fatal(err)
	}
	var decoded EngineeringSubjectRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	decodedOpaque, ok := decoded.AsOpaque()
	if !ok || decodedOpaque != opaque {
		t.Errorf("round trip mismatch: got (%v, %v), want (%v, true)", decodedOpaque, ok, opaque)
	}
}

func TestEngineeringSubjectRefOpaqueRejectsKnownKind(t *testing.T) {
	if _, err := NewOpaqueEngineeringSubjectRef(SubjectKindArtifact, "ns", "id"); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
}

func TestEngineeringSubjectRefUnrecognizedKindNonOpaqueRefFails(t *testing.T) {
	// An unrecognized kind whose "ref" is not even a JSON object cannot be
	// preserved losslessly, and must fail rather than being stored as a
	// raw blob.
	err := json.Unmarshal([]byte(`{"kind":"future-kind","ref":[1,2,3]}`), &EngineeringSubjectRef{})
	if !errors.Is(err, ErrInvalidPayload) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPayload)
	}

	// An unrecognized kind whose "ref" is a JSON object but does not
	// supply a namespace/identifier is also rejected, one level lower (by
	// the opaque constructor's own empty-identity check).
	err = json.Unmarshal([]byte(`{"kind":"future-kind","ref":{"nested":{"a":1}}}`), &EngineeringSubjectRef{})
	if !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
	}
}

// --- LifecycleSubjectRef -------------------------------------------------

func TestLifecycleSubjectRefKnownVariants(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")
	decisionID := mustDecisionID(t, "DEC-1")

	artifactRef, err := NewArtifactRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	artifactRevisionRef, err := NewArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	requirementRef, err := NewRequirementRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	requirementRevisionRef, err := NewRequirementArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	decisionRef, err := NewDecisionRef(decisionID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		make func() (LifecycleSubjectRef, error)
		kind string
		as   func(EngineeringSubjectRef) bool
	}{
		{"FromArtifact", func() (LifecycleSubjectRef, error) { return NewLifecycleSubjectRefFromArtifact(artifactRef) }, SubjectKindArtifact, func(r EngineeringSubjectRef) bool { _, ok := r.AsArtifact(); return ok }},
		{"FromArtifactRevision", func() (LifecycleSubjectRef, error) {
			return NewLifecycleSubjectRefFromArtifactRevision(artifactRevisionRef)
		}, SubjectKindArtifactRevision, func(r EngineeringSubjectRef) bool { _, ok := r.AsArtifactRevision(); return ok }},
		{"FromRequirement", func() (LifecycleSubjectRef, error) { return NewLifecycleSubjectRefFromRequirement(requirementRef) }, SubjectKindRequirement, func(r EngineeringSubjectRef) bool { _, ok := r.AsRequirement(); return ok }},
		{"FromRequirementRevision", func() (LifecycleSubjectRef, error) {
			return NewLifecycleSubjectRefFromRequirementRevision(requirementRevisionRef)
		}, SubjectKindRequirementRevision, func(r EngineeringSubjectRef) bool { _, ok := r.AsRequirementRevision(); return ok }},
		{"FromDecision", func() (LifecycleSubjectRef, error) { return NewLifecycleSubjectRefFromDecision(decisionRef) }, SubjectKindDecision, func(r EngineeringSubjectRef) bool { _, ok := r.AsDecision(); return ok }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, err := tt.make()
			if err != nil {
				t.Fatal(err)
			}
			if subject.Kind() != tt.kind {
				t.Errorf("Kind() = %q, want %q", subject.Kind(), tt.kind)
			}
			if !tt.as(subject.Subject()) {
				t.Errorf("accessor for kind %q returned ok=false", tt.kind)
			}
			// A mismatched accessor (any kind other than this one) must
			// fail; Decision is a safe "other kind" probe for every case
			// except the Decision case itself, and Artifact is safe there.
			if tt.kind != SubjectKindDecision {
				if _, ok := subject.Subject().AsDecision(); ok {
					t.Error("AsDecision() ok=true for a non-decision Lifecycle Subject")
				}
			} else if _, ok := subject.Subject().AsArtifact(); ok {
				t.Error("AsArtifact() ok=true for a decision Lifecycle Subject")
			}

			data, err := json.Marshal(subject)
			if err != nil {
				t.Fatal(err)
			}
			var decoded LifecycleSubjectRef
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != subject.Kind() {
				t.Errorf("round trip Kind() = %q, want %q", decoded.Kind(), subject.Kind())
			}
			if !tt.as(decoded.Subject()) {
				t.Errorf("accessor for round-tripped kind %q returned ok=false", tt.kind)
			}
		})
	}
}

func TestLifecycleSubjectRefOpaque(t *testing.T) {
	subject, err := NewOpaqueLifecycleSubjectRef("planned-validation-activity", "plan-x", "ACT-1")
	if err != nil {
		t.Fatal(err)
	}
	if subject.IsZero() {
		t.Error("valid opaque LifecycleSubjectRef reports IsZero() = true")
	}
	opaque, ok := subject.Subject().AsOpaque()
	if !ok || opaque.Kind() != "planned-validation-activity" || opaque.Namespace() != "plan-x" || opaque.Identifier() != "ACT-1" {
		t.Errorf("AsOpaque() = (%+v, %v)", opaque, ok)
	}

	data, err := json.Marshal(subject)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LifecycleSubjectRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	decodedOpaque, ok := decoded.Subject().AsOpaque()
	if !ok || decodedOpaque != opaque {
		t.Errorf("round trip mismatch: got (%v, %v), want (%v, true)", decodedOpaque, ok, opaque)
	}
}

// --- PEOS-008 Runtime record references (Packet J.1) -------------------------
//
// RuntimeBindingRecordRef, RuntimeUnbindingRecordRef, and
// RuntimeObservationRef mirror ValidationClaimRef/ValidationExecutionRecordRef
// exactly: each wraps its record's identity type, exposes RecordID/IsZero,
// and marshals as {"record_id": "..."}. Tests below mirror
// TestValidationExecutionRecordRef's shape for each of the three.

func TestRuntimeBindingRecordRef(t *testing.T) {
	if _, err := NewRuntimeBindingRecordRef(RuntimeBindingRecordID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty identity: error = %v, want %v", err, ErrEmptyIdentity)
	}

	recordID, err := NewRuntimeBindingRecordID("BIND-1")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := NewRuntimeBindingRecordRef(recordID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsZero() {
		t.Error("valid RuntimeBindingRecordRef reports IsZero() = true")
	}
	var zero RuntimeBindingRecordRef
	if !zero.IsZero() {
		t.Error("zero-value RuntimeBindingRecordRef.IsZero() = false, want true")
	}
	if ref.RecordID() != recordID {
		t.Error("RecordID() mismatch")
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"record_id":"BIND-1"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded RuntimeBindingRecordRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}

	if err := json.Unmarshal([]byte(`{"record_id":""}`), &decoded); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty record_id: error = %v, want %v", err, ErrEmptyIdentity)
	}
	if err := json.Unmarshal([]byte(`not json`), &decoded); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
	if decoded != ref {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestRuntimeUnbindingRecordRef(t *testing.T) {
	if _, err := NewRuntimeUnbindingRecordRef(RuntimeUnbindingRecordID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty identity: error = %v, want %v", err, ErrEmptyIdentity)
	}

	recordID, err := NewRuntimeUnbindingRecordID("UNBIND-1")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := NewRuntimeUnbindingRecordRef(recordID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsZero() {
		t.Error("valid RuntimeUnbindingRecordRef reports IsZero() = true")
	}
	var zero RuntimeUnbindingRecordRef
	if !zero.IsZero() {
		t.Error("zero-value RuntimeUnbindingRecordRef.IsZero() = false, want true")
	}
	if ref.RecordID() != recordID {
		t.Error("RecordID() mismatch")
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"record_id":"UNBIND-1"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded RuntimeUnbindingRecordRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}

	if err := json.Unmarshal([]byte(`{"record_id":""}`), &decoded); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty record_id: error = %v, want %v", err, ErrEmptyIdentity)
	}
	if err := json.Unmarshal([]byte(`not json`), &decoded); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
	if decoded != ref {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestRuntimeObservationRef(t *testing.T) {
	if _, err := NewRuntimeObservationRef(RuntimeObservationID{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty identity: error = %v, want %v", err, ErrEmptyIdentity)
	}

	recordID, err := NewRuntimeObservationID("OBS-1")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := NewRuntimeObservationRef(recordID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsZero() {
		t.Error("valid RuntimeObservationRef reports IsZero() = true")
	}
	var zero RuntimeObservationRef
	if !zero.IsZero() {
		t.Error("zero-value RuntimeObservationRef.IsZero() = false, want true")
	}
	if ref.RecordID() != recordID {
		t.Error("RecordID() mismatch")
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"record_id":"OBS-1"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded RuntimeObservationRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != ref {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ref)
	}

	if err := json.Unmarshal([]byte(`{"record_id":""}`), &decoded); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty record_id: error = %v, want %v", err, ErrEmptyIdentity)
	}
	if err := json.Unmarshal([]byte(`not json`), &decoded); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
	if decoded != ref {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// TestRuntimeRecordRefsUsableAsCorrectionTarget confirms
// RuntimeBindingRecordRef and RuntimeUnbindingRecordRef satisfy
// correctionTarget (IsZero() bool + json.Marshaler), so each can
// instantiate RecordCorrectionRef[T] as PEOS-008 requires for Runtime
// Binding Record (:282) and Runtime Unbinding Record (:314) correction.
// RuntimeObservationRef deliberately is not exercised this way: PEOS-008
// documents no correction reference for Runtime Observation.
func TestRuntimeRecordRefsUsableAsCorrectionTarget(t *testing.T) {
	bindingID, err := NewRuntimeBindingRecordID("BIND-1")
	if err != nil {
		t.Fatal(err)
	}
	bindingRef, err := NewRuntimeBindingRecordRef(bindingID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewRecordCorrectionRef(CorrectionKindCorrect, bindingRef); err != nil {
		t.Fatalf("RuntimeBindingRecordRef as correction target: %v", err)
	}

	unbindingID, err := NewRuntimeUnbindingRecordID("UNBIND-1")
	if err != nil {
		t.Fatal(err)
	}
	unbindingRef, err := NewRuntimeUnbindingRecordRef(unbindingID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewRecordCorrectionRef(CorrectionKindReplace, unbindingRef); err != nil {
		t.Fatalf("RuntimeUnbindingRecordRef as correction target: %v", err)
	}
}
