package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCriterionRefRequirementKinds(t *testing.T) {
	artifactID := mustArtifactID(t, "REQ-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")

	reqRef, err := NewRequirementRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := CriterionRefFromRequirement(reqRef)
	if err != nil {
		t.Fatal(err)
	}
	if criterion.Kind() != CriterionKindRequirement {
		t.Errorf("Kind() = %q, want %q", criterion.Kind(), CriterionKindRequirement)
	}
	got, ok := criterion.AsRequirement()
	if !ok || got != reqRef {
		t.Errorf("AsRequirement() = (%v, %v), want (%v, true)", got, ok, reqRef)
	}
	if _, ok := criterion.AsRequirementRevision(); ok {
		t.Error("AsRequirementRevision() ok=true for a requirement-identity criterion")
	}

	reqRevRef, err := NewRequirementArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	revCriterion, err := CriterionRefFromRequirementRevision(reqRevRef)
	if err != nil {
		t.Fatal(err)
	}
	if revCriterion.Kind() != CriterionKindRequirementRevision {
		t.Errorf("Kind() = %q, want %q", revCriterion.Kind(), CriterionKindRequirementRevision)
	}
}

func TestCriterionRefIsDistinctFromEngineeringSubjectRef(t *testing.T) {
	// A Claim's subject (EngineeringSubjectRef) and a Claim's criterion
	// (CriterionRef) are different Go types with no construction path
	// from one to the other, even when they wrap the same underlying
	// RequirementArtifactRevisionRef. This is the mechanical enforcement
	// of "a criterion is not a second Claim Subject" (PEOS-006) and of
	// keeping a Requirement's dual role (subject vs. criterion) type-safe
	// (the BL-1 locked resolution).
	artifactID := mustArtifactID(t, "REQ-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")
	ref, err := NewRequirementArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}

	subject, err := EngineeringSubjectRefFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := CriterionRefFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}

	// The following, if uncommented, must fail to compile: there is no
	// conversion between EngineeringSubjectRef and CriterionRef.
	//   _ = CriterionRef(subject)
	//   _ = EngineeringSubjectRef(criterion)

	if subject.Kind() != SubjectKindRequirementRevision {
		t.Errorf("subject.Kind() = %q", subject.Kind())
	}
	if criterion.Kind() != CriterionKindRequirementRevision {
		t.Errorf("criterion.Kind() = %q", criterion.Kind())
	}
}

func TestCriterionRefArtifactKinds(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")

	artifactRef, err := NewArtifactRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	c, err := CriterionRefFromArtifact(artifactRef)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.AsArtifact(); !ok {
		t.Error("AsArtifact() ok=false")
	}

	artifactRevisionRef, err := NewArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := CriterionRefFromArtifactRevision(artifactRevisionRef)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c2.AsArtifactRevision(); !ok {
		t.Error("AsArtifactRevision() ok=false")
	}
}

func TestCriterionRefQualityElement(t *testing.T) {
	profileArtifactID := mustArtifactID(t, "PROFILE-1")
	profileRevisionID := mustArtifactRevisionID(t, "REV-1")
	profileRef, err := NewArtifactRevisionRef(profileArtifactID, profileRevisionID)
	if err != nil {
		t.Fatal(err)
	}
	element, err := NewLocalKey("readability")
	if err != nil {
		t.Fatal(err)
	}

	qeRef, err := NewQualityElementCriterionRef(profileRef, element)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewQualityElementCriterionRef(ArtifactRevisionRef{}, element); err == nil {
		t.Error("zero profile accepted, want error")
	}

	characteristicCriterion, err := CriterionRefFromQualityCharacteristic(qeRef)
	if err != nil {
		t.Fatal(err)
	}
	if characteristicCriterion.Kind() != CriterionKindQualityCharacteristic {
		t.Errorf("Kind() = %q", characteristicCriterion.Kind())
	}

	measureCriterion, err := CriterionRefFromQualityMeasure(qeRef)
	if err != nil {
		t.Fatal(err)
	}
	if measureCriterion.Kind() != CriterionKindQualityMeasure {
		t.Errorf("Kind() = %q", measureCriterion.Kind())
	}

	data, err := json.Marshal(characteristicCriterion)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CriterionRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, ok := decoded.AsQualityCharacteristic()
	if !ok || got != qeRef {
		t.Errorf("round trip mismatch: got (%v, %v), want (%v, true)", got, ok, qeRef)
	}
}

func TestCriterionRefRuntimeRule(t *testing.T) {
	contractArtifactID := mustArtifactID(t, "CONTRACT-1")
	contractRevisionID := mustArtifactRevisionID(t, "REV-1")
	contractRef, err := NewRuntimeContractRevisionRef(contractArtifactID, contractRevisionID)
	if err != nil {
		t.Fatal(err)
	}
	rule, err := NewLocalKey("latency-slo")
	if err != nil {
		t.Fatal(err)
	}
	rrRef, err := NewRuntimeRuleCriterionRef(contractRef, rule)
	if err != nil {
		t.Fatal(err)
	}

	ruleCriterion, err := CriterionRefFromRuntimeContractRule(rrRef)
	if err != nil {
		t.Fatal(err)
	}
	if ruleCriterion.Kind() != CriterionKindRuntimeContractRule {
		t.Errorf("Kind() = %q", ruleCriterion.Kind())
	}

	assertionCriterion, err := CriterionRefFromRuntimeAssertion(rrRef)
	if err != nil {
		t.Fatal(err)
	}
	if assertionCriterion.Kind() != CriterionKindRuntimeAssertion {
		t.Errorf("Kind() = %q", assertionCriterion.Kind())
	}
}

func TestCriterionRefTemplateConstraint(t *testing.T) {
	templateArtifactID := mustArtifactID(t, "TEMPLATE-1")
	templateRevisionID := mustArtifactRevisionID(t, "REV-1")
	templateRef, err := NewTemplateArtifactRevisionRef(templateArtifactID, templateRevisionID)
	if err != nil {
		t.Fatal(err)
	}
	constraintKey, err := NewLocalKey("max-length")
	if err != nil {
		t.Fatal(err)
	}
	tcRef, err := NewTemplateConstraintCriterionRef(templateRef, constraintKey)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := CriterionRefFromTemplateConstraint(tcRef)
	if err != nil {
		t.Fatal(err)
	}
	if criterion.Kind() != CriterionKindTemplateConstraint {
		t.Errorf("Kind() = %q", criterion.Kind())
	}
}

func TestProductRuleRef(t *testing.T) {
	if _, err := NewProductRuleRef(VocabularyValue{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("zero vocabulary: error = %v, want %v", err, ErrEmptyIdentity)
	}

	v := mustVocabularyValue(t, "acme-corp", "pricing-rule-7")
	rule, err := NewProductRuleRef(v)
	if err != nil {
		t.Fatal(err)
	}
	if rule.IsZero() {
		t.Error("valid ProductRuleRef reports IsZero() = true")
	}
	if !rule.VocabularyValue().Equal(v) {
		t.Errorf("VocabularyValue() = %v, want %v", rule.VocabularyValue(), v)
	}

	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ProductRuleRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.VocabularyValue().Equal(v) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded.VocabularyValue(), v)
	}
}

func TestExternalRuleRef(t *testing.T) {
	if _, err := NewExternalRuleRef(VocabularyValue{}); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("zero vocabulary: error = %v, want %v", err, ErrEmptyIdentity)
	}

	v := mustVocabularyValue(t, "iso-25010", "maintainability")
	rule, err := NewExternalRuleRef(v)
	if err != nil {
		t.Fatal(err)
	}
	if rule.IsZero() {
		t.Error("valid ExternalRuleRef reports IsZero() = true")
	}
	if !rule.VocabularyValue().Equal(v) {
		t.Errorf("VocabularyValue() = %v, want %v", rule.VocabularyValue(), v)
	}

	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ExternalRuleRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.VocabularyValue().Equal(v) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded.VocabularyValue(), v)
	}
}

func TestProductRuleRefAndExternalRuleRefAreNotInterchangeable(t *testing.T) {
	// ProductRuleRef and ExternalRuleRef wrap the same underlying
	// VocabularyValue shape but are distinct Go types with distinct field
	// names; the following, if uncommented, must fail to compile:
	//   var _ ProductRuleRef = ExternalRuleRef{}
	//   _ = ProductRuleRef(ExternalRuleRef{})
	v := mustVocabularyValue(t, "acme-corp", "same-looking-value")
	product, err := NewProductRuleRef(v)
	if err != nil {
		t.Fatal(err)
	}
	external, err := NewExternalRuleRef(v)
	if err != nil {
		t.Fatal(err)
	}
	if product.VocabularyValue() != external.VocabularyValue() {
		t.Error("underlying VocabularyValue differs unexpectedly")
	}
}

func TestCriterionRefProductAndExternalRules(t *testing.T) {
	productRule, err := NewProductRuleRef(mustVocabularyValue(t, "acme-corp", "pricing-rule-7"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := CriterionRefFromProductRule(productRule)
	if err != nil {
		t.Fatal(err)
	}
	if c.Kind() != CriterionKindProductRule {
		t.Errorf("Kind() = %q, want %q", c.Kind(), CriterionKindProductRule)
	}
	got, ok := c.AsProductRule()
	if !ok || got != productRule {
		t.Errorf("AsProductRule() = (%v, %v), want (%v, true)", got, ok, productRule)
	}
	if _, ok := c.AsExternalRule(); ok {
		t.Error("AsExternalRule() ok=true for a product_rule CriterionRef")
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decodedC CriterionRef
	if err := json.Unmarshal(data, &decodedC); err != nil {
		t.Fatal(err)
	}
	gotDecoded, ok := decodedC.AsProductRule()
	if !ok || gotDecoded != productRule {
		t.Errorf("round trip AsProductRule() = (%v, %v), want (%v, true)", gotDecoded, ok, productRule)
	}

	externalRule, err := NewExternalRuleRef(mustVocabularyValue(t, "iso-25010", "maintainability"))
	if err != nil {
		t.Fatal(err)
	}
	c2, err := CriterionRefFromExternalRule(externalRule)
	if err != nil {
		t.Fatal(err)
	}
	if c2.Kind() != CriterionKindExternalRule {
		t.Errorf("Kind() = %q, want %q", c2.Kind(), CriterionKindExternalRule)
	}
	got2, ok := c2.AsExternalRule()
	if !ok || got2 != externalRule {
		t.Errorf("AsExternalRule() = (%v, %v), want (%v, true)", got2, ok, externalRule)
	}
	if _, ok := c2.AsProductRule(); ok {
		t.Error("AsProductRule() ok=true for an external_rule CriterionRef")
	}

	data2, err := json.Marshal(c2)
	if err != nil {
		t.Fatal(err)
	}
	var decodedC2 CriterionRef
	if err := json.Unmarshal(data2, &decodedC2); err != nil {
		t.Fatal(err)
	}
	gotDecoded2, ok := decodedC2.AsExternalRule()
	if !ok || gotDecoded2 != externalRule {
		t.Errorf("round trip AsExternalRule() = (%v, %v), want (%v, true)", gotDecoded2, ok, externalRule)
	}
}

func TestCriterionRefOpaque(t *testing.T) {
	c, err := NewOpaqueCriterionRef("future-criterion-kind", "product-x", "rule-1")
	if err != nil {
		t.Fatal(err)
	}
	if c.IsKnown() {
		t.Error("IsKnown() = true for opaque criterion")
	}
	opaque, ok := c.AsOpaque()
	if !ok || opaque.Kind() != "future-criterion-kind" {
		t.Errorf("AsOpaque() = (%v, %v)", opaque, ok)
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CriterionRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Kind() != c.Kind() {
		t.Errorf("round trip Kind() = %q, want %q", decoded.Kind(), c.Kind())
	}
}

func TestCriterionRefOpaqueRejectsKnownKind(t *testing.T) {
	if _, err := NewOpaqueCriterionRef(CriterionKindRequirement, "ns", "id"); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
}

func TestCriterionRefInvalidDiscriminator(t *testing.T) {
	if err := json.Unmarshal([]byte(`{"kind":"","ref":{}}`), &CriterionRef{}); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Error("empty kind accepted, want error")
	}
}

func TestCriterionRefPayloadMismatch(t *testing.T) {
	err := json.Unmarshal([]byte(`{"kind":"requirement","ref":{"artifact_id":""}}`), &CriterionRef{})
	if err == nil {
		t.Error("payload mismatch accepted, want error")
	}
}
