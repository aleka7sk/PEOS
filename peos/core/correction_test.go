package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCorrectionKindConstants(t *testing.T) {
	if CorrectionKindCorrect.Value().Value() != "correct" {
		t.Errorf("CorrectionKindCorrect = %v", CorrectionKindCorrect)
	}
	if CorrectionKindReplace.Value().Value() != "replace" {
		t.Errorf("CorrectionKindReplace = %v", CorrectionKindReplace)
	}
	if CorrectionKindInvalidate.Value().Value() != "invalidate" {
		t.Errorf("CorrectionKindInvalidate = %v", CorrectionKindInvalidate)
	}

	// This package never spells record correction as "supersede" /
	// "superseded"; that vocabulary belongs to PEOS-002 Artifact
	// Supersession, a distinct mechanism for Artifacts, not records.
	for _, k := range []CorrectionKind{CorrectionKindCorrect, CorrectionKindReplace, CorrectionKindInvalidate} {
		if k.Value().Value() == "supersede" || k.Value().Value() == "supersedes" || k.Value().Value() == "superseded" {
			t.Errorf("CorrectionKind %v uses forbidden Supersession vocabulary", k)
		}
	}
}

func TestRecordRefKnownKinds(t *testing.T) {
	claimID, err := NewValidationClaimID("CLAIM-1")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := RecordRefFromValidationClaim(claimID)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Kind() != RecordKindValidationClaim {
		t.Errorf("Kind() = %q, want %q", ref.Kind(), RecordKindValidationClaim)
	}
	got, ok := ref.AsValidationClaim()
	if !ok || got != claimID {
		t.Errorf("AsValidationClaim() = (%v, %v), want (%v, true)", got, ok, claimID)
	}
	if _, ok := ref.AsRuntimeViolation(); ok {
		t.Error("AsRuntimeViolation() ok=true for a validation_claim RecordRef")
	}

	violationID, err := NewRuntimeViolationID("VIOL-1")
	if err != nil {
		t.Fatal(err)
	}
	violationRef, err := RecordRefFromRuntimeViolation(violationID)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(violationRef)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RecordRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	gotViolation, ok := decoded.AsRuntimeViolation()
	if !ok || gotViolation != violationID {
		t.Errorf("round trip mismatch: got (%v, %v), want (%v, true)", gotViolation, ok, violationID)
	}
}

func TestRecordRefOpaque(t *testing.T) {
	ref, err := NewOpaqueRecordRef("future-record-kind", "product-x", "rec-1")
	if err != nil {
		t.Fatal(err)
	}
	if ref.IsKnown() {
		t.Error("IsKnown() = true for opaque record ref")
	}
	if _, err := NewOpaqueRecordRef(RecordKindValidationClaim, "ns", "id"); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
}

func TestRecordCorrectionRefTypedTarget(t *testing.T) {
	claimID, err := NewValidationClaimID("CLAIM-2")
	if err != nil {
		t.Fatal(err)
	}
	target, err := NewValidationClaimRef(claimID)
	if err != nil {
		t.Fatal(err)
	}

	correction, err := NewRecordCorrectionRef(CorrectionKindReplace, target)
	if err != nil {
		t.Fatal(err)
	}
	if correction.Kind() != CorrectionKindReplace {
		t.Errorf("Kind() = %v, want %v", correction.Kind(), CorrectionKindReplace)
	}
	if correction.Target() != target {
		t.Errorf("Target() = %v, want %v", correction.Target(), target)
	}

	data, err := json.Marshal(correction)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RecordCorrectionRef[ValidationClaimRef]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Target() != target {
		t.Errorf("round trip Target() = %v, want %v", decoded.Target(), target)
	}
}

func TestRecordCorrectionRefDynamicTarget(t *testing.T) {
	// RecordCorrectionRef[RecordRef] is the type-erased form, useful when
	// the corrected record's family is decided at runtime rather than
	// fixed by a specific PEOS SDK packet's struct field.
	violationID, err := NewRuntimeViolationID("VIOL-2")
	if err != nil {
		t.Fatal(err)
	}
	target, err := RecordRefFromRuntimeViolation(violationID)
	if err != nil {
		t.Fatal(err)
	}
	correction, err := NewRecordCorrectionRef(CorrectionKindCorrect, target)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := correction.Target().AsRuntimeViolation()
	if !ok || got != violationID {
		t.Errorf("Target().AsRuntimeViolation() = (%v, %v), want (%v, true)", got, ok, violationID)
	}
}

func TestRecordCorrectionRefRejectsZeroKindOrTarget(t *testing.T) {
	claimID, err := NewValidationClaimID("CLAIM-3")
	if err != nil {
		t.Fatal(err)
	}
	target, err := NewValidationClaimRef(claimID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewRecordCorrectionRef(CorrectionKind{}, target); !errors.Is(err, ErrInvalidCorrectionReference) {
		t.Errorf("zero kind: error = %v, want %v", err, ErrInvalidCorrectionReference)
	}
	if _, err := NewRecordCorrectionRef(CorrectionKindCorrect, ValidationClaimRef{}); !errors.Is(err, ErrInvalidCorrectionReference) {
		t.Errorf("zero target: error = %v, want %v", err, ErrInvalidCorrectionReference)
	}
}

func TestRecordCorrectionRefUnsupportedValueRejectedByCorrectionKindItself(t *testing.T) {
	// CorrectionKind is technically an open vocabulary wrapper (a future
	// PEOS specification amendment might add a fourth kind); this test
	// documents that an arbitrary namespaced value is accepted by
	// NewCorrectionKind itself (no closed Go enum), while
	// RecordCorrectionRef only rejects the zero value, not "unknown"
	// values. Enforcing the current three-kind closed set is a future
	// validator's job, not this package's.
	arbitrary := NewCorrectionKind(mustVocabularyValue(t, "peos", "supersede"))
	claimID, err := NewValidationClaimID("CLAIM-4")
	if err != nil {
		t.Fatal(err)
	}
	target, err := NewValidationClaimRef(claimID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewRecordCorrectionRef(arbitrary, target); err != nil {
		t.Errorf("unexpected error for a non-zero, non-standard CorrectionKind: %v", err)
	}
}
