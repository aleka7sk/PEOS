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

// --- PEOS-007 Quality Profile Revision-owned criterion kinds -----------------
//
// The three kinds below (Threshold, Target, Quality Constraint) were added
// additively for PEOS-007, which requires all three as Quality Claim
// criteria. They reuse QualityElementCriterionRef because PEOS-007 scopes
// all five Profile-owned element kinds identically: (owning Profile
// Revision, local key within it).

// mustQualityElementRef builds a QualityElementCriterionRef for tests.
func mustQualityElementRef(t *testing.T, artifactID, revisionID, element string) QualityElementCriterionRef {
	t.Helper()
	profileRef, err := NewArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	key, err := NewLocalKey(element)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := NewQualityElementCriterionRef(profileRef, key)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestCriterionRefQualityProfileOwnedKinds(t *testing.T) {
	ref := mustQualityElementRef(t, "PROFILE-1", "REV-1", "latency-p99")

	cases := []struct {
		name        string
		construct   func(QualityElementCriterionRef) (CriterionRef, error)
		wantKind    string
		accessor    func(CriterionRef) (QualityElementCriterionRef, bool)
		otherAccess []func(CriterionRef) (QualityElementCriterionRef, bool)
	}{
		{
			name:      "threshold",
			construct: CriterionRefFromQualityThreshold,
			wantKind:  CriterionKindQualityThreshold,
			accessor:  CriterionRef.AsQualityThreshold,
			otherAccess: []func(CriterionRef) (QualityElementCriterionRef, bool){
				CriterionRef.AsQualityTarget,
				CriterionRef.AsQualityConstraint,
				CriterionRef.AsQualityCharacteristic,
				CriterionRef.AsQualityMeasure,
			},
		},
		{
			name:      "target",
			construct: CriterionRefFromQualityTarget,
			wantKind:  CriterionKindQualityTarget,
			accessor:  CriterionRef.AsQualityTarget,
			otherAccess: []func(CriterionRef) (QualityElementCriterionRef, bool){
				CriterionRef.AsQualityThreshold,
				CriterionRef.AsQualityConstraint,
				CriterionRef.AsQualityCharacteristic,
				CriterionRef.AsQualityMeasure,
			},
		},
		{
			name:      "constraint",
			construct: CriterionRefFromQualityConstraint,
			wantKind:  CriterionKindQualityConstraint,
			accessor:  CriterionRef.AsQualityConstraint,
			otherAccess: []func(CriterionRef) (QualityElementCriterionRef, bool){
				CriterionRef.AsQualityThreshold,
				CriterionRef.AsQualityTarget,
				CriterionRef.AsQualityCharacteristic,
				CriterionRef.AsQualityMeasure,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Constructor success, exact discriminator, IsKnown.
			c, err := tc.construct(ref)
			if err != nil {
				t.Fatal(err)
			}
			if c.Kind() != tc.wantKind {
				t.Errorf("Kind() = %q, want %q", c.Kind(), tc.wantKind)
			}
			if !c.IsKnown() {
				t.Error("IsKnown() = false, want true for a typed kind")
			}
			if c.IsZero() {
				t.Error("IsZero() = true for a constructed criterion")
			}

			// Correct accessor returns the payload.
			got, ok := tc.accessor(c)
			if !ok || got != ref {
				t.Errorf("accessor = (%v, %v), want (%v, true)", got, ok, ref)
			}

			// Every other quality accessor must refuse it. This is what
			// keeps Target and Threshold unconflatable (PEOS-007: "The two
			// SHALL NOT be conflated") even though they share a payload
			// shape.
			for _, other := range tc.otherAccess {
				if _, ok := other(c); ok {
					t.Error("a sibling quality accessor accepted this criterion; kinds are conflated")
				}
			}
			if _, ok := c.AsOpaque(); ok {
				t.Error("AsOpaque() ok=true for a known kind")
			}

			// Zero payload rejected.
			if _, err := tc.construct(QualityElementCriterionRef{}); !errors.Is(err, ErrInvalidPayload) {
				t.Errorf("zero payload error = %v, want %v", err, ErrInvalidPayload)
			}

			// MarshalJSON carries the exact discriminator and the composite
			// payload.
			data, err := json.Marshal(c)
			if err != nil {
				t.Fatal(err)
			}
			var env struct {
				Kind string `json:"kind"`
				Ref  struct {
					Profile map[string]string `json:"profile"`
					Element string            `json:"element"`
				} `json:"ref"`
			}
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatal(err)
			}
			if env.Kind != tc.wantKind {
				t.Errorf("wire kind = %q, want %q", env.Kind, tc.wantKind)
			}
			if env.Ref.Element != "latency-p99" {
				t.Errorf("wire element = %q, want %q", env.Ref.Element, "latency-p99")
			}
			if env.Ref.Profile["artifact_id"] != "PROFILE-1" {
				t.Errorf("wire profile artifact_id = %q, want PROFILE-1", env.Ref.Profile["artifact_id"])
			}

			// UnmarshalJSON round-trips through the dedicated arm.
			var decoded CriterionRef
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != tc.wantKind {
				t.Errorf("round trip Kind() = %q, want %q", decoded.Kind(), tc.wantKind)
			}
			if !decoded.IsKnown() {
				t.Error("round trip IsKnown() = false; the kind decoded through the opaque path")
			}
			gotDecoded, ok := tc.accessor(decoded)
			if !ok || gotDecoded != ref {
				t.Errorf("round trip accessor = (%v, %v), want (%v, true)", gotDecoded, ok, ref)
			}

			// A missing ref is rejected: the payload decodes to its zero
			// value, which the constructor refuses.
			missing := []byte(`{"kind":"` + tc.wantKind + `"}`)
			if err := json.Unmarshal(missing, &CriterionRef{}); err == nil {
				t.Error("missing ref accepted, want error")
			}

			// An explicit null ref is rejected the same way every existing
			// composite arm rejects it: QualityElementCriterionRef's own
			// UnmarshalJSON is never reached for a JSON null (encoding/json
			// leaves the value untouched), so the zero payload reaches the
			// constructor and fails with ErrEmptyIdentity.
			null := []byte(`{"kind":"` + tc.wantKind + `","ref":null}`)
			if err := json.Unmarshal(null, &CriterionRef{}); !errors.Is(err, ErrEmptyIdentity) {
				t.Errorf("explicit null ref error = %v, want %v", err, ErrEmptyIdentity)
			}

			// A malformed nested ArtifactRevisionRef preserves the nested
			// sentinel through errors.Is rather than being re-attributed.
			malformed := []byte(`{"kind":"` + tc.wantKind +
				`","ref":{"profile":{"artifact_id":"P-1","revision_id":""},"element":"k"}}`)
			err = json.Unmarshal(malformed, &CriterionRef{})
			if err == nil {
				t.Fatal("malformed nested profile ref accepted, want error")
			}
			if !errors.Is(err, ErrMissingRevisionID) && !errors.Is(err, ErrEmptyIdentity) {
				t.Errorf("malformed nested ref error = %v, want a nested core sentinel", err)
			}

			// An unknown ordinary field is ignored, exactly as it is for
			// every existing arm.
			extra := []byte(`{"kind":"` + tc.wantKind +
				`","ref":{"profile":{"artifact_id":"PROFILE-1","revision_id":"REV-1"},"element":"latency-p99"},"unknown":1}`)
			var withExtra CriterionRef
			if err := json.Unmarshal(extra, &withExtra); err != nil {
				t.Fatalf("unknown ordinary field rejected: %v", err)
			}
			if _, ok := tc.accessor(withExtra); !ok {
				t.Error("unknown ordinary field changed the decoded arm")
			}

			// The kind is registered as known, so the opaque constructor
			// must refuse it.
			if _, err := NewOpaqueCriterionRef(tc.wantKind, "ns", "id"); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
				t.Errorf("NewOpaqueCriterionRef(%q) error = %v, want %v", tc.wantKind, err, ErrInvalidReferenceDiscriminator)
			}
		})
	}
}

// TestQualityCompositeCriteriaRequireDedicatedKinds records why the three
// kinds above had to be added to this file rather than expressed through
// NewOpaqueCriterionRef, and locks that reasoning against regression.
//
// The opaque path preserves exactly one shape: (namespace, identifier). A
// Quality Threshold/Target/Constraint reference is a composite -- an
// Artifact Revision reference (itself two identity values) plus a LocalKey
// -- so it has no faithful encoding as a namespaced scalar. Before these
// arms existed, such a criterion was unrepresentable; the two assertions
// below show both halves of that: the opaque path cannot carry the payload,
// and the dedicated arms round-trip it exactly.
func TestQualityCompositeCriteriaRequireDedicatedKinds(t *testing.T) {
	ref := mustQualityElementRef(t, "PROFILE-9", "REV-3", "availability")

	// The opaque path cannot express the composite: the kinds are now
	// registered as known, so it refuses them outright rather than
	// silently truncating the Profile Revision reference to a string.
	for _, kind := range []string{
		CriterionKindQualityThreshold,
		CriterionKindQualityTarget,
		CriterionKindQualityConstraint,
	} {
		if _, err := NewOpaqueCriterionRef(kind, "peos", "PROFILE-9/REV-3/availability"); err == nil {
			t.Errorf("NewOpaqueCriterionRef(%q) accepted a flattened composite, want rejection", kind)
		}
	}

	// A JSON document using one of these kinds with an opaque
	// (namespace, identifier) payload is rejected too: the dedicated arm
	// now owns the discriminator, and the composite payload is required.
	opaqueShaped := []byte(`{"kind":"quality_threshold","ref":{"namespace":"peos","identifier":"x"}}`)
	if err := json.Unmarshal(opaqueShaped, &CriterionRef{}); err == nil {
		t.Error("opaque-shaped payload accepted for quality_threshold, want error")
	}

	// The dedicated arms preserve the full composite across a round trip,
	// including both identity values of the owning Profile Revision.
	c, err := CriterionRefFromQualityThreshold(ref)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CriterionRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, ok := decoded.AsQualityThreshold()
	if !ok {
		t.Fatal("AsQualityThreshold() ok=false after round trip")
	}
	if got.Profile() != ref.Profile() || got.Element() != ref.Element() {
		t.Errorf("composite not preserved: got (%v, %v), want (%v, %v)",
			got.Profile(), got.Element(), ref.Profile(), ref.Element())
	}
}

// TestCriterionRefExistingArmsUnchanged locks the additive nature of the
// PEOS-007 change: every criterion kind that existed before it still
// constructs, still reports the same discriminator, and still round-trips
// through its own accessor. A regression here would mean the three new arms
// were not additive after all.
func TestCriterionRefExistingArmsUnchanged(t *testing.T) {
	artifactID := mustArtifactID(t, "ART-1")
	revisionID := mustArtifactRevisionID(t, "REV-1")
	qeRef := mustQualityElementRef(t, "PROFILE-1", "REV-1", "readability")

	reqRef, err := NewRequirementRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	reqRevRef, err := NewRequirementArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	artifactRef, err := NewArtifactRef(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	artifactRevRef, err := NewArtifactRevisionRef(artifactID, revisionID)
	if err != nil {
		t.Fatal(err)
	}
	runtimeRef, err := NewRuntimeRuleCriterionRef(
		func() RuntimeContractRevisionRef {
			r, err := NewRuntimeContractRevisionRef(mustArtifactID(t, "CONTRACT-1"), revisionID)
			if err != nil {
				t.Fatal(err)
			}
			return r
		}(),
		func() LocalKey {
			k, err := NewLocalKey("rule-1")
			if err != nil {
				t.Fatal(err)
			}
			return k
		}(),
	)
	if err != nil {
		t.Fatal(err)
	}
	templateRef, err := NewTemplateConstraintCriterionRef(
		func() TemplateArtifactRevisionRef {
			r, err := NewTemplateArtifactRevisionRef(mustArtifactID(t, "TPL-1"), revisionID)
			if err != nil {
				t.Fatal(err)
			}
			return r
		}(),
		func() LocalKey {
			k, err := NewLocalKey("param-1")
			if err != nil {
				t.Fatal(err)
			}
			return k
		}(),
	)
	if err != nil {
		t.Fatal(err)
	}
	productRule, err := NewProductRuleRef(mustVocabularyValue(t, "product-x", "rule-1"))
	if err != nil {
		t.Fatal(err)
	}
	externalRule, err := NewExternalRuleRef(mustVocabularyValue(t, "iso", "25010-1"))
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		criterion CriterionRef
		wantKind  string
	}{
		{mustCriterion(CriterionRefFromRequirement(reqRef)), CriterionKindRequirement},
		{mustCriterion(CriterionRefFromRequirementRevision(reqRevRef)), CriterionKindRequirementRevision},
		{mustCriterion(CriterionRefFromArtifact(artifactRef)), CriterionKindArtifact},
		{mustCriterion(CriterionRefFromArtifactRevision(artifactRevRef)), CriterionKindArtifactRevision},
		{mustCriterion(CriterionRefFromQualityCharacteristic(qeRef)), CriterionKindQualityCharacteristic},
		{mustCriterion(CriterionRefFromQualityMeasure(qeRef)), CriterionKindQualityMeasure},
		{mustCriterion(CriterionRefFromRuntimeContractRule(runtimeRef)), CriterionKindRuntimeContractRule},
		{mustCriterion(CriterionRefFromRuntimeAssertion(runtimeRef)), CriterionKindRuntimeAssertion},
		{mustCriterion(CriterionRefFromTemplateConstraint(templateRef)), CriterionKindTemplateConstraint},
		{mustCriterion(CriterionRefFromProductRule(productRule)), CriterionKindProductRule},
		{mustCriterion(CriterionRefFromExternalRule(externalRule)), CriterionKindExternalRule},
	}

	if len(cases) != 11 {
		t.Fatalf("expected the 11 pre-PEOS-007 criterion kinds, have %d", len(cases))
	}

	for _, tc := range cases {
		if tc.criterion.Kind() != tc.wantKind {
			t.Errorf("Kind() = %q, want %q", tc.criterion.Kind(), tc.wantKind)
		}
		data, err := json.Marshal(tc.criterion)
		if err != nil {
			t.Fatalf("%s: marshal: %v", tc.wantKind, err)
		}
		var decoded CriterionRef
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("%s: unmarshal: %v", tc.wantKind, err)
		}
		if decoded.Kind() != tc.wantKind {
			t.Errorf("%s: round trip Kind() = %q", tc.wantKind, decoded.Kind())
		}
		if !decoded.IsKnown() {
			t.Errorf("%s: round trip IsKnown() = false", tc.wantKind)
		}
		// None of the pre-existing kinds may be readable through a new
		// PEOS-007 accessor.
		if _, ok := decoded.AsQualityThreshold(); ok && tc.wantKind != CriterionKindQualityThreshold {
			t.Errorf("%s: readable as quality threshold", tc.wantKind)
		}
		if _, ok := decoded.AsQualityTarget(); ok && tc.wantKind != CriterionKindQualityTarget {
			t.Errorf("%s: readable as quality target", tc.wantKind)
		}
		if _, ok := decoded.AsQualityConstraint(); ok && tc.wantKind != CriterionKindQualityConstraint {
			t.Errorf("%s: readable as quality constraint", tc.wantKind)
		}
	}
}

// mustCriterion unwraps a CriterionRef constructor result. It panics
// instead of calling t.Fatal because Go does not permit a multi-value call
// to be mixed with other arguments, so t cannot be passed alongside it; a
// panic fails the test just as loudly and never occurs for the valid
// references this test builds.
func mustCriterion(c CriterionRef, err error) CriterionRef {
	if err != nil {
		panic(err)
	}
	return c
}
