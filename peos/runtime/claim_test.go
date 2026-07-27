package runtime

import (
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

func mustValidationMethod(t *testing.T, value string) core.ValidationMethod {
	t.Helper()
	return core.NewValidationMethod(mustVocabularyValue(t, "product", value))
}

func mustEngineeringSubjectRefFromRuntimeSubject(t *testing.T, namespace, identifier string) core.EngineeringSubjectRef {
	t.Helper()
	subject, err := core.EngineeringSubjectRefFromRuntimeSubject(mustRuntimeSubjectRef(t, namespace, identifier))
	if err != nil {
		t.Fatal(err)
	}
	return subject
}

func mustValidationClaimID(t *testing.T, value string) core.ValidationClaimID {
	t.Helper()
	id, err := core.NewValidationClaimID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestNewComplianceClaim(t *testing.T) {
	subject := mustEngineeringSubjectRefFromRuntimeSubject(t, "kubernetes", "pod-1")
	criterion := mustCriterionRef(t, "REQ-1")
	evidence := mustEvidence(t, "ART-EV-1", "REV-1")

	claim, err := NewComplianceClaim(
		mustValidationClaimID(t, "CLAIM-1"),
		subject,
		mustScope(t, "cluster=prod-1"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		[]core.CriterionRef{criterion},
		[]core.EvidenceArtifactRevisionRef{evidence},
		mustTimestampAt(t, 0),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}

	if claim.ClaimType() != core.ClaimTypeCompliance {
		t.Errorf("ClaimType() = %v, want %v", claim.ClaimType(), core.ClaimTypeCompliance)
	}
	if claim.Subject() != subject {
		t.Error("Subject() mismatch")
	}
	if claim.Scope() != mustScope(t, "cluster=prod-1") {
		t.Error("Scope() mismatch")
	}
	if claim.Outcome() != core.ClaimOutcomeSatisfied {
		t.Error("Outcome() mismatch")
	}
	if claim.Method() != mustValidationMethod(t, "runtime-observation") {
		t.Error("Method() mismatch")
	}
	criteria := claim.Criteria()
	if len(criteria) != 1 || criteria[0] != criterion {
		t.Errorf("Criteria() = %v", criteria)
	}
	claimEvidence := claim.Evidence()
	if len(claimEvidence) != 1 || claimEvidence[0] != evidence {
		t.Errorf("Evidence() = %v", claimEvidence)
	}
	if claim.Timestamp() != mustTimestampAt(t, 0) {
		t.Error("Timestamp() mismatch")
	}
	if claim.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
}

func TestNewComplianceClaimReachesValidationSentinels(t *testing.T) {
	subject := mustEngineeringSubjectRefFromRuntimeSubject(t, "kubernetes", "pod-1")
	criterion := mustCriterionRef(t, "REQ-1")
	evidence := mustEvidence(t, "ART-EV-1", "REV-1")

	if _, err := NewComplianceClaim(
		core.ValidationClaimID{},
		subject,
		mustScope(t, "cluster=prod-1"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		[]core.CriterionRef{criterion},
		[]core.EvidenceArtifactRevisionRef{evidence},
		mustTimestampAt(t, 0),
		mustProvenance(t),
	); !errors.Is(err, validation.ErrInvalidValidationClaim) {
		t.Errorf("zero id: error = %v, want %v", err, validation.ErrInvalidValidationClaim)
	}

	if _, err := NewComplianceClaim(
		mustValidationClaimID(t, "CLAIM-1"),
		core.EngineeringSubjectRef{},
		mustScope(t, "cluster=prod-1"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		[]core.CriterionRef{criterion},
		[]core.EvidenceArtifactRevisionRef{evidence},
		mustTimestampAt(t, 0),
		mustProvenance(t),
	); !errors.Is(err, validation.ErrInvalidValidationClaim) {
		t.Errorf("zero subject: error = %v, want %v", err, validation.ErrInvalidValidationClaim)
	}

	// PEOS-006 requires at least one Evidence citation for every Claim
	// Type; a Compliance Claim is no exception, since NewComplianceClaim
	// delegates to validation.NewClaim without weakening that rule.
	if _, err := NewComplianceClaim(
		mustValidationClaimID(t, "CLAIM-1"),
		subject,
		mustScope(t, "cluster=prod-1"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		[]core.CriterionRef{criterion},
		nil,
		mustTimestampAt(t, 0),
		mustProvenance(t),
	); err == nil {
		t.Error("zero evidence accepted, want error")
	}

	if _, err := NewComplianceClaim(
		mustValidationClaimID(t, "CLAIM-1"),
		subject,
		core.Scope{},
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		[]core.CriterionRef{criterion},
		[]core.EvidenceArtifactRevisionRef{evidence},
		mustTimestampAt(t, 0),
		mustProvenance(t),
	); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestNewComplianceClaimReturnsOrdinaryValidationClaim(t *testing.T) {
	subject := mustEngineeringSubjectRefFromRuntimeSubject(t, "kubernetes", "pod-1")
	claim, err := NewComplianceClaim(
		mustValidationClaimID(t, "CLAIM-1"),
		subject,
		mustScope(t, "cluster=prod-1"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		nil,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
		mustTimestampAt(t, 0),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	// The returned value is an ordinary validation.Claim: it can be
	// round-tripped through the exact same JSON path any other Claim uses,
	// with no runtime-specific envelope or discriminator.
	var _ validation.Claim = claim
}

func TestNewComplianceClaimZeroCriteriaAccepted(t *testing.T) {
	// PEOS-008 permits a Compliance Claim's criteria to include
	// applicable Waiver conditions, Runtime Contract rules, and other
	// kinds beyond Requirements; it adds no minimum criteria count, so
	// zero criteria must be accepted, mirroring validation.Claim's
	// existing behavior for non-Satisfaction, non-Conformance Claim Types.
	subject := mustEngineeringSubjectRefFromRuntimeSubject(t, "kubernetes", "pod-1")
	claim, err := NewComplianceClaim(
		mustValidationClaimID(t, "CLAIM-1"),
		subject,
		mustScope(t, "cluster=prod-1"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "runtime-observation"),
		nil,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
		mustTimestampAt(t, 0),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(claim.Criteria()) != 0 {
		t.Errorf("Criteria() = %v, want empty", claim.Criteria())
	}
}
