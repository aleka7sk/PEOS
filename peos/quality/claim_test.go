package quality

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// --- helpers -----------------------------------------------------------------

func mustClaimID(t *testing.T, value string) core.ValidationClaimID {
	t.Helper()
	id, err := core.NewValidationClaimID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustClaimOutcome(t *testing.T, value string) core.ClaimOutcome {
	t.Helper()
	return core.NewClaimOutcome(mustVocabularyValue(t, core.PEOSNamespace, value))
}

func mustEvidence(t *testing.T, artifactID, revisionID string) core.EvidenceArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustRequirementSubject(t *testing.T, artifactID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewRequirementRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := core.EngineeringSubjectRefFromRequirement(ref)
	if err != nil {
		t.Fatal(err)
	}
	return subject
}

func mustRequirementRevisionSubject(t *testing.T, artifactID, revisionID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewRequirementArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := core.EngineeringSubjectRefFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return subject
}

func mustRequirementCriterion(t *testing.T, artifactID string) core.CriterionRef {
	t.Helper()
	ref, err := core.NewRequirementRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	c, err := core.CriterionRefFromRequirement(ref)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustRequirementRevisionCriterion(t *testing.T, artifactID, revisionID string) core.CriterionRef {
	t.Helper()
	ref, err := core.NewRequirementArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	c, err := core.CriterionRefFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// newTestClaim builds a quality Claim with the given subject and criteria,
// holding every other field constant.
func newTestClaim(t *testing.T, subject core.EngineeringSubjectRef, criteria []core.CriterionRef) (Claim, error) {
	t.Helper()
	return NewClaim(
		mustClaimID(t, "VC-1"),
		subject,
		mustScope(t, "service=checkout"),
		mustClaimOutcome(t, "satisfied"),
		mustValidationMethod(t, "automated-test"),
		criteria,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
}

// mustQualityClaim builds the canonical valid quality Claim: an Artifact
// Revision subject with a Characteristic and a Measure criterion.
func mustQualityClaim(t *testing.T) Claim {
	t.Helper()
	c, err := newTestClaim(t, mustArtifactSubject(t, "ART-1", "REV-1"), []core.CriterionRef{
		mustCharacteristicCriterion(t, "latency"),
		mustMeasureCriterion(t, "latency-p99"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// newRawValidationClaim builds a bare validation.Claim of the given Claim Type,
// bypassing the quality wrapper entirely. This is how the wrapping tests obtain
// the values NewClaimFromValidationClaim must accept or reject.
func newRawValidationClaim(t *testing.T, claimType core.ClaimType, subject core.EngineeringSubjectRef, criteria []core.CriterionRef) (validation.Claim, error) {
	t.Helper()
	return validation.NewClaim(
		mustClaimID(t, "VC-1"),
		claimType,
		subject,
		mustScope(t, "service=checkout"),
		mustClaimOutcome(t, "satisfied"),
		mustValidationMethod(t, "automated-test"),
		criteria,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
}

// --- construction ------------------------------------------------------------

func TestNewQualityClaim(t *testing.T) {
	c := mustQualityClaim(t)

	if c.ClaimType() != core.ClaimTypeQuality {
		t.Errorf("ClaimType() = %v, want %v", c.ClaimType(), core.ClaimTypeQuality)
	}
	if c.ID() != mustClaimID(t, "VC-1") {
		t.Error("ID() does not delegate")
	}
	if c.Subject() != mustArtifactSubject(t, "ART-1", "REV-1") {
		t.Error("Subject() does not delegate")
	}
	if c.Scope() != mustScope(t, "service=checkout") {
		t.Error("Scope() does not delegate")
	}
	if c.Outcome() != mustClaimOutcome(t, "satisfied") {
		t.Error("Outcome() does not delegate")
	}
	if c.Method() != mustValidationMethod(t, "automated-test") {
		t.Error("Method() does not delegate")
	}
	if len(c.Criteria()) != 2 {
		t.Errorf("Criteria() length = %d, want 2", len(c.Criteria()))
	}
	if len(c.Evidence()) != 1 {
		t.Errorf("Evidence() length = %d, want 1", len(c.Evidence()))
	}
	if c.Timestamp() != mustTimestamp(t) {
		t.Error("Timestamp() does not delegate")
	}
	if c.Provenance().IsZero() {
		t.Error("Provenance() does not delegate")
	}
	if c.ExecutionRecords() != nil {
		t.Error("ExecutionRecords() non-nil before any is set")
	}
	if _, ok := c.Reasoning(); ok {
		t.Error("Reasoning() ok=true before one is set")
	}
	if _, ok := c.Authority(); ok {
		t.Error("Authority() ok=true before one is set")
	}
	if _, ok := c.Correction(); ok {
		t.Error("Correction() ok=true before one is set")
	}
	if !c.Extension().IsZero() {
		t.Error("Extension() non-zero before one is set")
	}
	if c.IsZero() {
		t.Error("IsZero() = true for a constructed claim")
	}

	ref, err := c.Ref()
	if err != nil {
		t.Fatal(err)
	}
	inner, err := c.ValidationClaim().Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref != inner {
		t.Error("Ref() does not delegate")
	}
}

// TestNewQualityClaimAcceptsZeroCriteria records that PEOS-007 states no
// minimum criteria count: "criteria may include" is permissive, and PEOS-006
// already permits zero criteria for a non-Satisfaction, non-Conformance Claim.
func TestNewQualityClaimAcceptsZeroCriteria(t *testing.T) {
	for name, criteria := range map[string][]core.CriterionRef{
		"nil":   nil,
		"empty": {},
	} {
		t.Run(name, func(t *testing.T) {
			c, err := newTestClaim(t, mustArtifactSubject(t, "ART-1", "REV-1"), criteria)
			if err != nil {
				t.Fatalf("zero criteria rejected: %v", err)
			}
			if c.Criteria() != nil {
				t.Error("Criteria() non-nil for a zero-criteria claim")
			}
		})
	}

	// A Requirement subject with no criteria is also accepted: with no
	// criterion there is nothing for the Requirement to be its own criterion
	// against.
	if _, err := newTestClaim(t, mustRequirementSubject(t, "REQ-1"), nil); err != nil {
		t.Errorf("a Requirement-subject claim with no criteria was rejected: %v", err)
	}
}

// TestNewQualityClaimStillRequiresEvidence confirms PEOS-006's Evidence rule is
// inherited unchanged: a Quality Claim cites at least one Evidence Artifact
// Revision, and PEOS-007 introduces no parallel Evidence mechanism.
func TestNewQualityClaimStillRequiresEvidence(t *testing.T) {
	_, err := NewClaim(
		mustClaimID(t, "VC-1"),
		mustArtifactSubject(t, "ART-1", "REV-1"),
		mustScope(t, "s"),
		mustClaimOutcome(t, "satisfied"),
		mustValidationMethod(t, "automated-test"),
		nil,
		nil,
		mustTimestamp(t),
		mustProvenance(t),
	)
	if !errors.Is(err, validation.ErrInvalidValidationClaim) {
		t.Errorf("error = %v, want it to wrap validation.ErrInvalidValidationClaim", err)
	}
}

func TestNewQualityClaimPropagatesPEOS006Failures(t *testing.T) {
	// A zero scope surfaces core's own sentinel, not a re-attributed one.
	_, err := NewClaim(
		mustClaimID(t, "VC-1"),
		mustArtifactSubject(t, "ART-1", "REV-1"),
		core.Scope{},
		mustClaimOutcome(t, "satisfied"),
		mustValidationMethod(t, "automated-test"),
		nil,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope error = %v, want it to wrap core.ErrInvalidScope", err)
	}

	// A zero id surfaces PEOS-006's aggregate sentinel.
	_, err = newTestClaim(t, mustArtifactSubject(t, "ART-1", "REV-1"), nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewClaim(
		core.ValidationClaimID{},
		mustArtifactSubject(t, "ART-1", "REV-1"),
		mustScope(t, "s"),
		mustClaimOutcome(t, "satisfied"),
		mustValidationMethod(t, "automated-test"),
		nil,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
		mustTimestamp(t),
		mustProvenance(t),
	)
	if !errors.Is(err, validation.ErrInvalidValidationClaim) {
		t.Errorf("zero id error = %v, want it to wrap validation.ErrInvalidValidationClaim", err)
	}
}

// --- the PEOS-007 Requirement-criterion invariant ----------------------------

// requirementReflexivityCases enumerates every shape PEOS-007's rule rejects
// and every shape it accepts. The comparison is by Requirement core.ArtifactID
// and is cross-level, because "the same Requirement" is identity-level
// language: an identity-level subject conflicts with a revision-level criterion
// of that Requirement, a revision-level subject conflicts with an
// identity-level criterion of it, and two different Revisions of one
// Requirement still conflict.
func requirementReflexivityCases(t *testing.T) (rejected, accepted map[string]struct {
	subject  core.EngineeringSubjectRef
	criteria []core.CriterionRef
}) {
	t.Helper()
	type shape = struct {
		subject  core.EngineeringSubjectRef
		criteria []core.CriterionRef
	}

	rejected = map[string]shape{
		"REQ-1 subject, REQ-1 criterion": {
			mustRequirementSubject(t, "REQ-1"),
			[]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")},
		},
		"REQ-1 subject, REQ-1/REV-1 criterion": {
			mustRequirementSubject(t, "REQ-1"),
			[]core.CriterionRef{mustRequirementRevisionCriterion(t, "REQ-1", "REV-1")},
		},
		"REQ-1/REV-1 subject, REQ-1 criterion": {
			mustRequirementRevisionSubject(t, "REQ-1", "REV-1"),
			[]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")},
		},
		"REQ-1/REV-1 subject, REQ-1/REV-1 criterion": {
			mustRequirementRevisionSubject(t, "REQ-1", "REV-1"),
			[]core.CriterionRef{mustRequirementRevisionCriterion(t, "REQ-1", "REV-1")},
		},
		"REQ-1/REV-1 subject, REQ-1/REV-2 criterion": {
			mustRequirementRevisionSubject(t, "REQ-1", "REV-1"),
			[]core.CriterionRef{mustRequirementRevisionCriterion(t, "REQ-1", "REV-2")},
		},
		"conflict hidden among valid quality criteria": {
			mustRequirementSubject(t, "REQ-1"),
			[]core.CriterionRef{
				mustCharacteristicCriterion(t, "testability"),
				mustMeasureCriterion(t, "ambiguity-count"),
				mustRequirementCriterion(t, "REQ-1"),
			},
		},
	}

	accepted = map[string]shape{
		"REQ-1 subject, REQ-2 criterion": {
			mustRequirementSubject(t, "REQ-1"),
			[]core.CriterionRef{mustRequirementCriterion(t, "REQ-2")},
		},
		"REQ-1/REV-1 subject, REQ-2/REV-1 criterion": {
			mustRequirementRevisionSubject(t, "REQ-1", "REV-1"),
			[]core.CriterionRef{mustRequirementRevisionCriterion(t, "REQ-2", "REV-1")},
		},
		"artifact revision subject, any Requirement criterion": {
			mustArtifactSubject(t, "ART-1", "REV-1"),
			[]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")},
		},
		"artifact revision subject, same-id Requirement criterion": {
			// The subject is an Artifact Revision, not a Requirement, so even a
			// criterion sharing its ArtifactID cannot conflict: the rule is
			// about a Requirement being its own criterion.
			mustArtifactSubject(t, "REQ-1", "REV-1"),
			[]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")},
		},
		"Requirement subject, no criteria": {
			mustRequirementSubject(t, "REQ-1"),
			nil,
		},
		"Requirement subject, all five profile criterion kinds": {
			mustRequirementSubject(t, "REQ-1"),
			[]core.CriterionRef{
				mustCharacteristicCriterion(t, "testability"),
				mustMeasureCriterion(t, "ambiguity-count"),
				mustThresholdCriterion(t, "max-ambiguities"),
				mustTargetCriterion(t, "target-ambiguities"),
				mustConstraintCriterion(t, "no-passive-voice"),
			},
		},
		"Requirement subject, external rule criterion": {
			mustRequirementSubject(t, "REQ-1"),
			[]core.CriterionRef{mustExternalRuleCriterion(t)},
		},
	}
	return rejected, accepted
}

func TestQualityClaimRequirementCriterionInvariant(t *testing.T) {
	rejected, accepted := requirementReflexivityCases(t)

	for name, shape := range rejected {
		t.Run("rejected: "+name, func(t *testing.T) {
			got, err := newTestClaim(t, shape.subject, shape.criteria)
			if !errors.Is(err, ErrInvalidQualityClaim) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityClaim)
			}
			if !got.IsZero() {
				t.Error("a failed constructor returned a non-zero value")
			}
		})
	}
	for name, shape := range accepted {
		t.Run("accepted: "+name, func(t *testing.T) {
			if _, err := newTestClaim(t, shape.subject, shape.criteria); err != nil {
				t.Errorf("a valid shape was rejected: %v", err)
			}
		})
	}
}

// TestPEOS006StillAcceptsWhatPEOS007Rejects is the boundary test. peos/validation
// is PEOS-006-correct in accepting a reflexive Requirement on a quality Claim --
// PEOS-006 carves the non-Satisfaction case out explicitly, and its own
// TestNonSatisfactionClaimMayUseSameRequirementAsSubjectAndCriterion asserts it.
// The stricter rule is PEOS-007's alone, and lives here. If this test ever
// fails, peos/validation was wrongly changed to make PEOS-007 easier.
func TestPEOS006StillAcceptsWhatPEOS007Rejects(t *testing.T) {
	subject := mustRequirementSubject(t, "REQ-1")
	criteria := []core.CriterionRef{mustRequirementCriterion(t, "REQ-1")}

	raw, err := newRawValidationClaim(t, core.ClaimTypeQuality, subject, criteria)
	if err != nil {
		t.Fatalf("peos/validation rejected a shape PEOS-006 permits; peos/validation may have been modified: %v", err)
	}
	if raw.ClaimType() != core.ClaimTypeQuality {
		t.Fatal("the raw claim is not a quality claim")
	}

	// The very same value is refused by the PEOS-007 wrapper.
	if _, err := NewClaimFromValidationClaim(raw); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("the quality wrapper accepted a reflexive Requirement: %v", err)
	}
}

// --- wrapping a raw validation.Claim -----------------------------------------

func TestNewClaimFromValidationClaim(t *testing.T) {
	subject := mustArtifactSubject(t, "ART-1", "REV-1")
	criteria := []core.CriterionRef{mustCharacteristicCriterion(t, "latency")}

	raw, err := newRawValidationClaim(t, core.ClaimTypeQuality, subject, criteria)
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := NewClaimFromValidationClaim(raw)
	if err != nil {
		t.Fatalf("a valid quality validation.Claim was rejected: %v", err)
	}
	if wrapped.ID() != raw.ID() || wrapped.ClaimType() != core.ClaimTypeQuality {
		t.Error("wrapping did not preserve the claim")
	}

	// A zero Claim is rejected.
	if _, err := NewClaimFromValidationClaim(validation.Claim{}); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("zero claim error = %v, want %v", err, ErrInvalidQualityClaim)
	}
}

// TestNewClaimFromValidationClaimRejectsEveryOtherClaimType asserts that a
// quality Claim is not merely a Claim that happens to be about quality: every
// other Claim Type, including a Product-defined one, is refused.
func TestNewClaimFromValidationClaimRejectsEveryOtherClaimType(t *testing.T) {
	productDefined := core.NewClaimType(mustVocabularyValue(t, "product-x", "fitness"))

	types := map[string]core.ClaimType{
		"satisfaction":         core.ClaimTypeSatisfaction,
		"conformance":          core.ClaimTypeConformance,
		"compliance":           core.ClaimTypeCompliance,
		"template conformance": core.ClaimTypeTemplateConformance,
		"product-defined":      productDefined,
	}
	for name, claimType := range types {
		t.Run(name, func(t *testing.T) {
			// Satisfaction and Conformance have their own PEOS-006 criteria
			// rules, so each type gets criteria that satisfy them -- otherwise
			// the raw construction would fail for a PEOS-006 reason and the
			// PEOS-007 rejection would never be exercised.
			subject := mustArtifactSubject(t, "ART-1", "REV-1")
			criteria := []core.CriterionRef{mustRequirementCriterion(t, "REQ-1")}

			raw, err := newRawValidationClaim(t, claimType, subject, criteria)
			if err != nil {
				t.Fatalf("could not build a raw %s claim: %v", name, err)
			}
			got, err := NewClaimFromValidationClaim(raw)
			if !errors.Is(err, ErrInvalidQualityClaim) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityClaim)
			}
			if !got.IsZero() {
				t.Error("a rejected claim type produced a non-zero wrapper")
			}
		})
	}
}

// --- modifiers ---------------------------------------------------------------

// TestQualityClaimWithCriteriaCannotBypassTheInvariant is the regression test
// Amendment A exists for. Before the wrapper, this exact sequence -- construct a
// valid quality Claim, then replace its criteria with a set naming its own
// Requirement subject -- produced a represented, PEOS-007-invalid Claim with no
// error anywhere, because validation.Claim.WithCriteria re-runs only PEOS-006's
// rules and reaches its default case for ClaimTypeQuality.
func TestQualityClaimWithCriteriaCannotBypassTheInvariant(t *testing.T) {
	// Step 1: a valid quality Claim whose subject is a Requirement.
	c, err := newTestClaim(t, mustRequirementSubject(t, "REQ-1"), []core.CriterionRef{
		mustCharacteristicCriterion(t, "testability"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: attempt to replace the criteria with a reflexive set.
	got, err := c.WithCriteria([]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")})

	// Step 3: it must be rejected.
	if !errors.Is(err, ErrInvalidQualityClaim) {
		t.Fatalf("WithCriteria error = %v, want %v -- the modifier bypassed the PEOS-007 invariant", err, ErrInvalidQualityClaim)
	}
	if !got.IsZero() {
		t.Error("a failed modifier returned a non-zero Claim")
	}

	// Step 4: the receiver is unchanged.
	if len(c.Criteria()) != 1 {
		t.Error("the receiver's criteria changed")
	}
	if _, ok := c.Criteria()[0].AsQualityCharacteristic(); !ok {
		t.Error("the receiver's criterion was replaced")
	}

	// The same bypass through the raw Claim would succeed -- which is exactly
	// why the wrapper exposes no way to reach it.
	if _, err := c.ValidationClaim().WithCriteria([]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")}); err != nil {
		t.Errorf("peos/validation rejected the shape it is supposed to permit: %v", err)
	}
}

func TestQualityClaimWithCriteriaAcceptsValidReplacement(t *testing.T) {
	c := mustQualityClaim(t)
	updated, err := c.WithCriteria([]core.CriterionRef{
		mustThresholdCriterion(t, "latency-max"),
		mustTargetCriterion(t, "latency-goal"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Criteria()) != 2 {
		t.Errorf("Criteria() length = %d, want 2", len(updated.Criteria()))
	}
	if _, ok := updated.Criteria()[0].AsQualityThreshold(); !ok {
		t.Error("the replacement criteria were not stored")
	}
	if len(c.Criteria()) != 2 {
		t.Error("the receiver was mutated")
	}
	if _, ok := c.Criteria()[0].AsQualityCharacteristic(); !ok {
		t.Error("the receiver's criteria were replaced")
	}

	// WithCriteria(nil) clears, which PEOS-007 permits.
	cleared, err := c.WithCriteria(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Criteria() != nil {
		t.Error("WithCriteria(nil) did not clear")
	}
}

// TestQualityClaimAllModifiersDelegateAndPreserveClaimType exercises all ten
// modifiers, asserting each one produces the expected change, leaves the
// receiver untouched, and yields a value that is still a quality Claim.
func TestQualityClaimAllModifiersDelegateAndPreserveClaimType(t *testing.T) {
	base := mustQualityClaim(t)
	authority := mustAuthority(t)

	target, err := core.NewValidationClaimRef(mustClaimID(t, "VC-0"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, target)
	if err != nil {
		t.Fatal(err)
	}
	execRef, err := core.NewValidationExecutionRecordRef(mustExecutionRecordID(t, "EXEC-1"))
	if err != nil {
		t.Fatal(err)
	}

	withRecords, err := base.WithExecutionRecords([]core.ValidationExecutionRecordRef{execRef})
	if err != nil {
		t.Fatal(err)
	}
	if len(withRecords.ExecutionRecords()) != 1 {
		t.Error("WithExecutionRecords did not set")
	}
	if base.ExecutionRecords() != nil {
		t.Error("WithExecutionRecords mutated the receiver")
	}

	withReasoning, err := base.WithReasoning("  measured against the profile  ")
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := withReasoning.Reasoning(); !ok || r != "measured against the profile" {
		t.Errorf("Reasoning() = (%q, %v)", r, ok)
	}
	clearedReasoning, err := withReasoning.WithoutReasoning()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := clearedReasoning.Reasoning(); ok {
		t.Error("WithoutReasoning did not clear")
	}
	if _, ok := withReasoning.Reasoning(); !ok {
		t.Error("WithoutReasoning mutated the receiver")
	}

	withAuthority, err := base.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := withAuthority.Authority(); !ok || got != authority {
		t.Error("WithAuthority did not set")
	}
	clearedAuthority, err := withAuthority.WithoutAuthority()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := clearedAuthority.Authority(); ok {
		t.Error("WithoutAuthority did not clear")
	}

	withCorrection, err := base.WithCorrection(correction)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withCorrection.Correction()
	if !ok {
		t.Fatal("WithCorrection did not set")
	}
	if got.Target() != target {
		t.Error("the correction target did not land on the claim")
	}
	clearedCorrection, err := withCorrection.WithoutCorrection()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := clearedCorrection.Correction(); ok {
		t.Error("WithoutCorrection did not clear")
	}

	withExtension, err := base.WithExtension(mustExtension(t))
	if err != nil {
		t.Fatal(err)
	}
	if withExtension.Extension().IsZero() {
		t.Error("WithExtension did not set")
	}
	clearedExtension, err := withExtension.WithoutExtension()
	if err != nil {
		t.Fatal(err)
	}
	if !clearedExtension.Extension().IsZero() {
		t.Error("WithoutExtension did not clear")
	}
	if withExtension.Extension().IsZero() {
		t.Error("WithoutExtension mutated the receiver")
	}

	// Every result is still a quality Claim, and every one still marshals.
	results := map[string]Claim{
		"WithExecutionRecords": withRecords,
		"WithReasoning":        withReasoning,
		"WithoutReasoning":     clearedReasoning,
		"WithAuthority":        withAuthority,
		"WithoutAuthority":     clearedAuthority,
		"WithCorrection":       withCorrection,
		"WithoutCorrection":    clearedCorrection,
		"WithExtension":        withExtension,
		"WithoutExtension":     clearedExtension,
	}
	for name, result := range results {
		if result.ClaimType() != core.ClaimTypeQuality {
			t.Errorf("%s: ClaimType() = %v, want %v", name, result.ClaimType(), core.ClaimTypeQuality)
		}
		if _, err := json.Marshal(result); err != nil {
			t.Errorf("%s: result does not marshal: %v", name, err)
		}
	}
}

func TestQualityClaimModifiersPropagatePEOS006Failures(t *testing.T) {
	base := mustQualityClaim(t)

	cases := map[string]func() (Claim, error){
		"zero authority":   func() (Claim, error) { return base.WithAuthority(core.AuthorityRef{}) },
		"blank reasoning":  func() (Claim, error) { return base.WithReasoning("   ") },
		"zero correction":  func() (Claim, error) { return base.WithCorrection(core.RecordCorrectionRef[core.ValidationClaimRef]{}) },
		"zero criterion":   func() (Claim, error) { return base.WithCriteria([]core.CriterionRef{{}}) },
		"zero exec record": func() (Claim, error) { return base.WithExecutionRecords([]core.ValidationExecutionRecordRef{{}}) },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := fn()
			if !errors.Is(err, validation.ErrInvalidValidationClaim) {
				t.Errorf("error = %v, want it to wrap validation.ErrInvalidValidationClaim", err)
			}
			if !got.IsZero() {
				t.Error("a failed modifier returned a non-zero Claim")
			}
		})
	}
}

// --- extract, modify, re-wrap ------------------------------------------------

func TestQualityClaimExtractedValidationClaimIsACopy(t *testing.T) {
	c, err := newTestClaim(t, mustRequirementSubject(t, "REQ-1"), []core.CriterionRef{
		mustCharacteristicCriterion(t, "testability"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Modify the extracted raw Claim into a PEOS-007-invalid shape. PEOS-006
	// permits it, so this succeeds.
	invalid, err := c.ValidationClaim().WithCriteria([]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")})
	if err != nil {
		t.Fatalf("peos/validation rejected the shape it permits: %v", err)
	}

	// The wrapper is untouched and still valid.
	if len(c.Criteria()) != 1 {
		t.Error("modifying the extracted claim mutated the wrapper")
	}
	if _, ok := c.Criteria()[0].AsQualityCharacteristic(); !ok {
		t.Error("the wrapper's criterion changed")
	}
	if _, err := json.Marshal(c); err != nil {
		t.Errorf("the wrapper became invalid: %v", err)
	}

	// Re-entry revalidates and refuses the modified value.
	if _, err := NewClaimFromValidationClaim(invalid); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("re-entry accepted an invalid claim: %v", err)
	}

	// A legitimately modified raw Claim re-enters successfully.
	valid, err := c.ValidationClaim().WithCriteria([]core.CriterionRef{mustRequirementCriterion(t, "REQ-2")})
	if err != nil {
		t.Fatal(err)
	}
	rewrapped, err := NewClaimFromValidationClaim(valid)
	if err != nil {
		t.Fatalf("re-entry rejected a valid claim: %v", err)
	}
	if len(rewrapped.Criteria()) != 1 {
		t.Error("re-entry lost the criteria")
	}

	// Slices returned through the wrapper are defensive copies, inherited from
	// validation.Claim.
	criteria := c.Criteria()
	criteria[0] = core.CriterionRef{}
	if c.Criteria()[0].IsZero() {
		t.Error("Criteria() returned an aliased slice")
	}
	evidence := c.Evidence()
	if len(evidence) > 0 {
		evidence[0] = core.EvidenceArtifactRevisionRef{}
		if c.Evidence()[0].IsZero() {
			t.Error("Evidence() returned an aliased slice")
		}
	}
}

// --- JSON --------------------------------------------------------------------

// TestQualityClaimJSONIsExactlyValidationClaimWireForm is the core wire-contract
// assertion: PEOS-007 adds no envelope, no key, and no discriminator. A quality
// Claim's document is readable by any PEOS-006 consumer.
func TestQualityClaimJSONIsExactlyValidationClaimWireForm(t *testing.T) {
	c := mustQualityClaim(t)

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	inner, err := json.Marshal(c.ValidationClaim())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(inner) {
		t.Errorf("wire form differs from the wrapped claim's:\n got %s\nwant %s", data, inner)
	}

	assertKeysPresent(t, data, "id", "claim_type", "subject", "scope", "outcome",
		"method", "criteria", "evidence", "timestamp", "provenance")
	// No wrapper envelope and no PEOS-007 discriminator.
	assertKeysAbsent(t, data, "claim", "quality_claim", "quality", "wrapper",
		"quality_claim_type", "type", "observed_value", "unit", "scale",
		"relation", "lifecycle", "state", "status", "basis", "verdict",
		"score", "quality_score", "current", "latest", "effective")

	// claim_type is ordinary content carrying peos:quality.
	var envelope struct {
		ClaimType string `json:"claim_type"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.ClaimType != "peos:quality" {
		t.Errorf("claim_type = %q, want %q", envelope.ClaimType, "peos:quality")
	}
}

func TestQualityClaimJSONRoundTrip(t *testing.T) {
	c := mustQualityClaim(t)
	c, err := c.WithReasoning("measured against QP-1/REV-1")
	if err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithAuthority(mustAuthority(t)); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithExtension(mustExtension(t)); err != nil {
		t.Fatal(err)
	}
	execRef, err := core.NewValidationExecutionRecordRef(mustExecutionRecordID(t, "EXEC-1"))
	if err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithExecutionRecords([]core.ValidationExecutionRecordRef{execRef}); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Claim
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != c.ID() || decoded.ClaimType() != core.ClaimTypeQuality {
		t.Error("identity or claim type lost in the round trip")
	}
	if len(decoded.Criteria()) != 2 || len(decoded.Evidence()) != 1 ||
		len(decoded.ExecutionRecords()) != 1 {
		t.Error("a collection was lost in the round trip")
	}
	if _, ok := decoded.Reasoning(); !ok {
		t.Error("reasoning lost")
	}
	if _, ok := decoded.Authority(); !ok {
		t.Error("authority lost")
	}
	if decoded.Extension().IsZero() {
		t.Error("extension lost")
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
	}
}

func TestQualityClaimJSONRejection(t *testing.T) {
	valid := mustQualityClaim(t)

	// A non-quality Claim Type is rejected on decode.
	raw, err := newRawValidationClaim(t, core.ClaimTypeConformance,
		mustArtifactSubject(t, "ART-1", "REV-1"),
		[]core.CriterionRef{mustCharacteristicCriterion(t, "latency")})
	if err != nil {
		t.Fatal(err)
	}
	conformanceData, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	decoded := valid
	if err := json.Unmarshal(conformanceData, &decoded); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("error = %v, want %v", err, ErrInvalidQualityClaim)
	}
	if decoded.ID() != valid.ID() || decoded.ClaimType() != core.ClaimTypeQuality {
		t.Error("a failed decode overwrote a previously valid receiver")
	}

	// A reflexive-Requirement quality Claim is rejected on decode: the same
	// document peos/validation accepts.
	reflexive, err := newRawValidationClaim(t, core.ClaimTypeQuality,
		mustRequirementSubject(t, "REQ-1"),
		[]core.CriterionRef{mustRequirementCriterion(t, "REQ-1")})
	if err != nil {
		t.Fatal(err)
	}
	reflexiveData, err := json.Marshal(reflexive)
	if err != nil {
		t.Fatal(err)
	}
	var target validation.Claim
	if err := json.Unmarshal(reflexiveData, &target); err != nil {
		t.Fatalf("peos/validation rejected its own document: %v", err)
	}
	decoded = valid
	if err := json.Unmarshal(reflexiveData, &decoded); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("a reflexive-Requirement document decoded successfully: %v", err)
	}
	if decoded.ID() != valid.ID() {
		t.Error("a failed decode overwrote a previously valid receiver")
	}

	// Structural rejections.
	for name, doc := range map[string]string{
		"null document":    `null`,
		"empty object":     `{}`,
		"not an object":    `42`,
		"missing evidence": `{"id":"VC-1","claim_type":"peos:quality","subject":{"kind":"artifact_revision","ref":{"artifact_id":"ART-1","revision_id":"REV-1"}},"scope":{"kind":"peos:component","expression":"s"},"outcome":"peos:satisfied","method":"product-x:test","timestamp":"2026-07-27T00:00:00Z","provenance":{"actor":{"namespace":"a","identifier":"b"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			d := valid
			if err := json.Unmarshal([]byte(doc), &d); err == nil {
				t.Fatalf("accepted %s, want rejection", doc)
			}
			if d.ID() != valid.ID() {
				t.Error("a failed decode overwrote a previously valid receiver")
			}
		})
	}

	// A zero wrapper fails to marshal.
	if _, err := json.Marshal(Claim{}); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("zero-value marshal error = %v, want %v", err, ErrInvalidQualityClaim)
	}
	if (Claim{}).IsZero() != true {
		t.Error("zero Claim IsZero() = false")
	}
}

func TestQualityClaimJSONNestedSentinelReachable(t *testing.T) {
	base := `{"id":"VC-1","claim_type":"peos:quality",` +
		`"subject":{"kind":"artifact_revision","ref":{"artifact_id":"ART-1","revision_id":"REV-1"}},` +
		`"scope":%s,"outcome":"peos:satisfied","method":"product-x:test",` +
		`"criteria":%s,"evidence":%s,"timestamp":"2026-07-27T00:00:00Z",` +
		`"provenance":{"actor":{"namespace":"a","identifier":"b"}}%s}`

	validScope := `{"kind":"peos:component","expression":"s"}`
	validCriteria := `[{"kind":"quality_characteristic","ref":{"profile":{"artifact_id":"QP-1","revision_id":"REV-1"},"element":"latency"}}]`
	validEvidence := `[{"artifact_id":"EV-1","revision_id":"REV-1"}]`

	cases := map[string]struct {
		doc  string
		want []error
	}{
		"malformed scope": {
			doc:  fmtDoc(base, `{"kind":"","expression":""}`, validCriteria, validEvidence, ""),
			want: []error{core.ErrInvalidScope, core.ErrInvalidVocabularyValue},
		},
		"malformed criterion reference": {
			doc:  fmtDoc(base, validScope, `[{"kind":"quality_characteristic","ref":{"profile":{"artifact_id":"QP-1","revision_id":""},"element":"latency"}}]`, validEvidence, ""),
			want: []error{core.ErrMissingRevisionID, core.ErrEmptyIdentity},
		},
		"malformed evidence reference": {
			doc:  fmtDoc(base, validScope, validCriteria, `[{"artifact_id":"EV-1","revision_id":""}]`, ""),
			want: []error{core.ErrMissingRevisionID, core.ErrEmptyIdentity},
		},
		"malformed correction reference": {
			doc:  fmtDoc(base, validScope, validCriteria, validEvidence, `,"correction":{"kind":"","target":{"kind":"validation_claim","ref":"VC-0"}}`),
			want: []error{core.ErrInvalidCorrectionReference, core.ErrInvalidVocabularyValue, core.ErrEmptyIdentity},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var c Claim
			err := json.Unmarshal([]byte(tc.doc), &c)
			if err == nil {
				t.Fatalf("accepted %s, want rejection", tc.doc)
			}
			for _, sentinel := range tc.want {
				if errors.Is(err, sentinel) {
					return
				}
			}
			t.Errorf("error = %v, want one of the nested core sentinels %v to remain reachable", err, tc.want)
		})
	}
}

// fmtDoc fills the four %s placeholders of the claim document template.
func fmtDoc(template string, scope, criteria, evidence, extra string) string {
	out := template
	for _, v := range []string{scope, criteria, evidence, extra} {
		idx := indexOf(out, "%s")
		if idx < 0 {
			break
		}
		out = out[:idx] + v + out[idx+2:]
	}
	return out
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// --- error consistency across every entry point ------------------------------

// TestQualityClaimErrorConsistency asserts that the same invalid input is
// rejected with the same sentinel through every one of the four public entry
// points. This is the observable consequence of routing all of them through the
// single checkQualityClaim path -- and it is what makes drift between them
// impossible rather than merely unlikely.
func TestQualityClaimErrorConsistency(t *testing.T) {
	subject := mustRequirementSubject(t, "REQ-1")
	reflexive := []core.CriterionRef{mustRequirementCriterion(t, "REQ-1")}

	// 1. NewClaim.
	_, errNew := newTestClaim(t, subject, reflexive)

	// 2. NewClaimFromValidationClaim.
	raw, err := newRawValidationClaim(t, core.ClaimTypeQuality, subject, reflexive)
	if err != nil {
		t.Fatal(err)
	}
	_, errWrap := NewClaimFromValidationClaim(raw)

	// 3. WithCriteria on an already-valid Claim.
	valid, err := newTestClaim(t, subject, []core.CriterionRef{mustCharacteristicCriterion(t, "testability")})
	if err != nil {
		t.Fatal(err)
	}
	_, errModifier := valid.WithCriteria(reflexive)

	// 4. UnmarshalJSON.
	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Claim
	errJSON := json.Unmarshal(data, &decoded)

	for name, err := range map[string]error{
		"NewClaim":                    errNew,
		"NewClaimFromValidationClaim": errWrap,
		"WithCriteria":                errModifier,
		"UnmarshalJSON":               errJSON,
	} {
		if !errors.Is(err, ErrInvalidQualityClaim) {
			t.Errorf("%s: error = %v, want %v", name, err, ErrInvalidQualityClaim)
		}
	}

	// The same four entry points agree on a wrong Claim Type too, for the two
	// that can observe one.
	conformance, err := newRawValidationClaim(t, core.ClaimTypeConformance, subject, reflexive)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewClaimFromValidationClaim(conformance); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("NewClaimFromValidationClaim on a conformance claim: %v", err)
	}
	conformanceData, err := json.Marshal(conformance)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(conformanceData, &decoded); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("UnmarshalJSON on a conformance claim: %v", err)
	}
}

// TestQualityClaimTypeIsUnreachable records that no public path can change the
// Claim Type: NewClaim supplies it, there is no WithClaimType (asserted in
// doc_test.go), and every modifier's result is re-verified.
func TestQualityClaimTypeIsUnreachable(t *testing.T) {
	c := mustQualityClaim(t)
	if c.ClaimType() != core.ClaimTypeQuality {
		t.Fatal("the canonical claim is not a quality claim")
	}

	// Even a document explicitly declaring another Claim Type cannot produce a
	// wrapper with that type: the decode is refused outright.
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	obj["claim_type"] = json.RawMessage(`"peos:compliance"`)
	tampered, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Claim
	if err := json.Unmarshal(tampered, &decoded); !errors.Is(err, ErrInvalidQualityClaim) {
		t.Errorf("a tampered claim_type was accepted: %v", err)
	}
	if !decoded.IsZero() {
		t.Error("a failed decode assigned the receiver")
	}
}
