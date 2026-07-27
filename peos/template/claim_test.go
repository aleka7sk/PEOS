package template

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

func mustClaimID(t *testing.T, value string) core.ValidationClaimID {
	t.Helper()
	id, err := core.NewValidationClaimID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustEvidence(t *testing.T, artifactID, revisionID string) core.EvidenceArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustTemplateSubject(t *testing.T) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.EngineeringSubjectRefFromTemplateRevision(mustTemplateRevisionRef(t, "TPL-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// mustTemplateConstraintCriterion builds a criterion citing a Parameter
// Constraint by its template-local key -- the criterion kind Packet K.1 made
// resolvable through TemplateContent.Constraint(key).
func mustTemplateConstraintCriterion(t *testing.T, constraintKey string) core.CriterionRef {
	t.Helper()
	payload, err := core.NewTemplateConstraintCriterionRef(
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustLocalKey(t, constraintKey),
	)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.CriterionRefFromTemplateConstraint(payload)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// TestNewTemplateConformanceClaim confirms the helper fixes the Claim Type and
// otherwise hands back an ordinary validation.Claim.
func TestNewTemplateConformanceClaim(t *testing.T) {
	claim, err := NewTemplateConformanceClaim(
		mustClaimID(t, "CLAIM-1"),
		mustTemplateSubject(t),
		mustScope(t, "tenant=acme"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "inspection"),
		[]core.CriterionRef{mustTemplateConstraintCriterion(t, "name-nonempty")},
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if claim.ClaimType() != core.ClaimTypeTemplateConformance {
		t.Errorf("ClaimType() = %v, want %v", claim.ClaimType(), core.ClaimTypeTemplateConformance)
	}
	if claim.ID() != mustClaimID(t, "CLAIM-1") {
		t.Error("ID() mismatch")
	}
	if len(claim.Criteria()) != 1 {
		t.Error("Criteria() mismatch")
	}
}

// TestTemplateConformanceClaimAcceptsEveryPermittedSubject confirms the helper
// constrains the Subject no further than PEOS-006 already does: PEOS-009 permits
// a generated Artifact, a generated Artifact Revision, a Template Artifact, a
// Template Artifact Revision, or "another explicitly permitted engineering
// subject", and all five are expressible through existing
// core.EngineeringSubjectRef arms.
func TestTemplateConformanceClaimAcceptsEveryPermittedSubject(t *testing.T) {
	generatedArtifact, err := core.EngineeringSubjectRefFromGeneratedArtifact(mustGeneratedArtifactRef(t, "GEN-1"))
	if err != nil {
		t.Fatal(err)
	}
	generatedRevision, err := core.EngineeringSubjectRefFromGeneratedArtifactRevision(mustGeneratedRevisionRef(t, "GEN-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	templateRef, err := core.NewTemplateRef(mustArtifactID(t, "TPL-1"))
	if err != nil {
		t.Fatal(err)
	}
	templateArtifact, err := core.EngineeringSubjectRefFromTemplate(templateRef)
	if err != nil {
		t.Fatal(err)
	}
	other, err := core.NewOpaqueEngineeringSubjectRef("product-thing", "product", "thing-1")
	if err != nil {
		t.Fatal(err)
	}

	for name, subject := range map[string]core.EngineeringSubjectRef{
		"generated artifact":          generatedArtifact,
		"generated artifact revision": generatedRevision,
		"template artifact":           templateArtifact,
		"template artifact revision":  mustTemplateSubject(t),
		"another permitted subject":   other,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewTemplateConformanceClaim(
				mustClaimID(t, "CLAIM-1"),
				subject,
				mustScope(t, "tenant=acme"),
				core.ClaimOutcomeSatisfied,
				mustValidationMethod(t, "inspection"),
				[]core.CriterionRef{mustTemplateConstraintCriterion(t, "name-nonempty")},
				[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
				mustTimestamp(t),
				mustProvenance(t),
			)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}

// TestTemplateConformanceClaimInheritsValidationRules confirms the helper
// re-implements and weakens nothing: every validation.Claim invariant still
// applies, surfacing peos/validation's own sentinels.
func TestTemplateConformanceClaimInheritsValidationRules(t *testing.T) {
	valid := func() (core.ValidationClaimID, core.EngineeringSubjectRef, core.Scope, core.ClaimOutcome, core.ValidationMethod, []core.CriterionRef, []core.EvidenceArtifactRevisionRef, core.Timestamp, core.Provenance) {
		return mustClaimID(t, "CLAIM-1"),
			mustTemplateSubject(t),
			mustScope(t, "tenant=acme"),
			core.ClaimOutcomeSatisfied,
			mustValidationMethod(t, "inspection"),
			[]core.CriterionRef{mustTemplateConstraintCriterion(t, "name-nonempty")},
			[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
			mustTimestamp(t),
			mustProvenance(t)
	}

	t.Run("zero id", func(t *testing.T) {
		_, subject, scope, outcome, method, criteria, evidence, ts, prov := valid()
		if _, err := NewTemplateConformanceClaim(core.ValidationClaimID{}, subject, scope, outcome, method, criteria, evidence, ts, prov); err == nil {
			t.Error("accepted, want error")
		}
	})
	t.Run("zero subject", func(t *testing.T) {
		id, _, scope, outcome, method, criteria, evidence, ts, prov := valid()
		if _, err := NewTemplateConformanceClaim(id, core.EngineeringSubjectRef{}, scope, outcome, method, criteria, evidence, ts, prov); err == nil {
			t.Error("accepted, want error")
		}
	})
	// PEOS-006 requires a Conformance Claim to cite at least one criterion, and
	// PEOS-009 makes a Template Conformance Claim a specialization of it that
	// "inherits, without redefinition, all Validation Claim rules defined by
	// PEOS-006". validation.NewClaim keys that check on ClaimTypeConformance
	// exactly and leaves PEOS-007/008/009 to add their own, so this helper adds
	// it -- surfacing peos/validation's own sentinel for the mirrored rule.
	for _, tt := range []struct {
		name     string
		criteria []core.CriterionRef
	}{
		{"nil criteria", nil},
		{"empty criteria", []core.CriterionRef{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			id, subject, scope, outcome, method, _, evidence, ts, prov := valid()
			_, err := NewTemplateConformanceClaim(id, subject, scope, outcome, method, tt.criteria, evidence, ts, prov)
			if !errors.Is(err, validation.ErrInvalidConformanceClaim) {
				t.Errorf("error = %v, want %v", err, validation.ErrInvalidConformanceClaim)
			}
		})
	}
	t.Run("no evidence", func(t *testing.T) {
		id, subject, scope, outcome, method, criteria, _, ts, prov := valid()
		if _, err := NewTemplateConformanceClaim(id, subject, scope, outcome, method, criteria, nil, ts, prov); err == nil {
			t.Error("accepted, want error")
		}
	})

	// The error is attributed to this package but keeps peos/validation's own
	// sentinel reachable through %w.
	id, subject, scope, outcome, method, criteria, evidence, ts, _ := valid()
	_, err := NewTemplateConformanceClaim(id, subject, scope, outcome, method, criteria, evidence, ts, core.Provenance{})
	if err == nil {
		t.Fatal("zero provenance accepted, want error")
	}
	if !errors.Is(err, validation.ErrInvalidValidationClaim) {
		t.Errorf("error = %v; peos/validation's sentinel should stay reachable", err)
	}
}

// TestTemplateConformanceClaimSubjectIsSingular confirms PEOS-009's "Single
// Template Conformance Subject Invariant" needs no enforcement here:
// validation.Claim takes exactly one subject, so the "Composite Template
// Conformance Subject" non-conforming pattern is unrepresentable.
func TestTemplateConformanceClaimSubjectIsSingular(t *testing.T) {
	claim, err := NewTemplateConformanceClaim(
		mustClaimID(t, "CLAIM-1"),
		mustTemplateSubject(t),
		mustScope(t, "tenant=acme"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "inspection"),
		[]core.CriterionRef{mustTemplateConstraintCriterion(t, "name-nonempty")},
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	subject := claim.Subject()
	if subject.IsZero() {
		t.Error("Subject() is zero")
	}
	got, ok := subject.AsTemplateRevision()
	if !ok || got != mustTemplateRevisionRef(t, "TPL-1", "REV-1") {
		t.Error("Subject() did not round-trip as a Template Artifact Revision")
	}
}

// TestTemplateConformanceClaimCriterionResolvesToAConstraint is the end-to-end
// tie between the Claim helper and Packet K.1's constraint namespace: a
// criterion cited by a Template Conformance Claim resolves back to exactly one
// ParameterConstraint on the named Template Revision's content.
func TestTemplateConformanceClaimCriterionResolvesToAConstraint(t *testing.T) {
	content, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true)},
		nil,
		[]ParameterConstraint{mustParameterConstraint(t, "name-nonempty", "name")},
	)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := NewTemplateRevision(mustTemplate(t, "TPL-1"), mustArtifactRevision(t, "TPL-1", "REV-1"), content)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := NewTemplateConformanceClaim(
		mustClaimID(t, "CLAIM-1"),
		mustTemplateSubject(t),
		mustScope(t, "tenant=acme"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "inspection"),
		[]core.CriterionRef{mustTemplateConstraintCriterion(t, "name-nonempty")},
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}

	criteria := claim.Criteria()
	if len(criteria) != 1 {
		t.Fatalf("Criteria() = %d, want 1", len(criteria))
	}
	payload, ok := criteria[0].AsTemplateConstraint()
	if !ok {
		t.Fatal("the criterion is not a template constraint")
	}
	revRef, err := revision.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if payload.Template() != revRef {
		t.Error("the criterion names a different Template Revision")
	}
	resolved, ok := revision.Content().Constraint(payload.Constraint())
	if !ok {
		t.Fatal("the criterion's local key did not resolve to a constraint")
	}
	if resolved.Key() != mustLocalKey(t, "name-nonempty") {
		t.Errorf("resolved constraint key = %v", resolved.Key())
	}
}

// TestTemplateConformanceClaimIsNotAnArtifact confirms the returned value is an
// ordinary validation.Claim -- not an Artifact, which PEOS-009 names as a
// non-conforming pattern ("Template Conformance Claim as Artifact").
func TestTemplateConformanceClaimIsNotAnArtifact(t *testing.T) {
	claim, err := NewTemplateConformanceClaim(
		mustClaimID(t, "CLAIM-1"),
		mustTemplateSubject(t),
		mustScope(t, "tenant=acme"),
		core.ClaimOutcomeSatisfied,
		mustValidationMethod(t, "inspection"),
		[]core.CriterionRef{mustTemplateConstraintCriterion(t, "name-nonempty")},
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	assertNoWireKeys(t, "TemplateConformanceClaim", claim, []string{
		"artifact_type", "artifact_id", "revision", "revision_id",
		"lifecycle", "state", "status", "conformant", "compatible",
	})
}
