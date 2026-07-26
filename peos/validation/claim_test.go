package validation

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- Claim helpers -----------------------------------------------------------

func mustClaimID(t *testing.T, value string) core.ValidationClaimID {
	t.Helper()
	id, err := core.NewValidationClaimID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustClaimRef(t *testing.T, value string) core.ValidationClaimRef {
	t.Helper()
	ref, err := core.NewValidationClaimRef(mustClaimID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
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

func mustProductRuleCriterion(t *testing.T, namespace, value string) core.CriterionRef {
	t.Helper()
	rule, err := core.NewProductRuleRef(mustVocab(t, namespace, value))
	if err != nil {
		t.Fatal(err)
	}
	c, err := core.CriterionRefFromProductRule(rule)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustOpaqueCriterion(t *testing.T, kind, namespace, identifier string) core.CriterionRef {
	t.Helper()
	c, err := core.NewOpaqueCriterionRef(kind, namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustRequirementSubject(t *testing.T, artifactID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewRequirementRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	s, err := core.EngineeringSubjectRefFromRequirement(ref)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustRequirementRevisionSubject(t *testing.T, artifactID, revisionID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewRequirementArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	s, err := core.EngineeringSubjectRefFromRequirementRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// newClaimWith builds a Claim with the given Claim Type, Subject, and
// criteria, holding every other mandatory argument at a valid default.
func newClaimWith(
	t *testing.T,
	claimType core.ClaimType,
	subject core.EngineeringSubjectRef,
	criteria []core.CriterionRef,
) (Claim, error) {
	t.Helper()
	return NewClaim(
		mustClaimID(t, "VC-1"),
		claimType,
		subject,
		mustScope(t, "deployment", "region=eu"),
		core.ClaimOutcomeSatisfied,
		mustMethod(t, "test"),
		criteria,
		[]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
		mustTimestamp(t, "2026-07-27T10:20:00Z"),
		mustProvenance(t),
	)
}

// mustGeneralClaim returns a valid Quality Claim with zero criteria -- the
// simplest Claim Type that imposes no criteria rule of its own.
func mustGeneralClaim(t *testing.T) Claim {
	t.Helper()
	c, err := newClaimWith(t, core.ClaimTypeQuality, mustSubject(t, "AR-42", "REV-3"), nil)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustSatisfactionClaim(t *testing.T) Claim {
	t.Helper()
	c, err := newClaimWith(t, core.ClaimTypeSatisfaction, mustSubject(t, "AR-42", "REV-3"),
		[]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustConformanceClaim(t *testing.T) Claim {
	t.Helper()
	c, err := newClaimWith(t, core.ClaimTypeConformance, mustSubject(t, "AR-42", "REV-3"),
		[]core.CriterionRef{mustProductRuleCriterion(t, "acme", "coding-standard")})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func fullClaim(t *testing.T) Claim {
	t.Helper()
	c := mustSatisfactionClaim(t)
	var err error
	if c, err = c.WithExecutionRecords([]core.ValidationExecutionRecordRef{mustExecutionRecordRef(t, "EXR-1")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithReasoning("all three assertions passed within the declared scope"); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithAuthority(mustAuthorityRef(t, "org", "qa-board")); err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindReplace, mustClaimRef(t, "VC-0"))
	if err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithCorrection(correction); err != nil {
		t.Fatal(err)
	}
	return c.WithExtension(mustExtension(t, "acme", `{"reviewer":"r1"}`))
}

// --- construction ------------------------------------------------------------

func TestNewClaimValidGeneralWithZeroCriteria(t *testing.T) {
	c := mustGeneralClaim(t)
	if c.IsZero() {
		t.Fatal("valid Claim reports IsZero")
	}
	if len(c.Criteria()) != 0 {
		t.Error("expected zero criteria")
	}
	if len(c.Evidence()) != 1 {
		t.Errorf("Evidence() length = %d, want 1", len(c.Evidence()))
	}
	if c.ID().String() != "VC-1" {
		t.Errorf("ID() = %q", c.ID().String())
	}
	ref, err := c.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ClaimID() != c.ID() {
		t.Error("Ref() does not cite the Claim")
	}
	if !c.ClaimType().Value().Equal(core.ClaimTypeQuality.Value()) {
		t.Error("claim type not preserved")
	}
	if !c.Outcome().Value().Equal(core.ClaimOutcomeSatisfied.Value()) {
		t.Error("outcome not preserved")
	}
	if !c.Scope().Equal(mustScope(t, "deployment", "region=eu")) {
		t.Error("scope not preserved")
	}
	if c.Timestamp().IsZero() || c.Provenance().IsZero() {
		t.Error("timestamp or provenance not preserved")
	}
	if len(c.ExecutionRecords()) != 0 {
		t.Error("minimum Claim reports execution records")
	}
	if _, ok := c.Reasoning(); ok {
		t.Error("minimum Claim reports reasoning")
	}
	if _, ok := c.Authority(); ok {
		t.Error("minimum Claim reports authority")
	}
	if _, ok := c.Correction(); ok {
		t.Error("minimum Claim reports a correction")
	}
	if !c.Extension().IsZero() {
		t.Error("minimum Claim has extension data")
	}
}

func TestNewClaimMandatoryRejections(t *testing.T) {
	id := mustClaimID(t, "VC-1")
	subject := mustSubject(t, "AR-42", "REV-3")
	scope := mustScope(t, "deployment", "region=eu")
	method := mustMethod(t, "test")
	evidence := []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")}
	ts := mustTimestamp(t, "2026-07-27T10:20:00Z")
	prov := mustProvenance(t)
	ct := core.ClaimTypeQuality
	outcome := core.ClaimOutcomeSatisfied

	cases := []struct {
		name string
		call func() (Claim, error)
		want error
	}{
		{"zero id", func() (Claim, error) {
			return NewClaim(core.ValidationClaimID{}, ct, subject, scope, outcome, method, nil, evidence, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero claim type", func() (Claim, error) {
			return NewClaim(id, core.ClaimType{}, subject, scope, outcome, method, nil, evidence, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero subject", func() (Claim, error) {
			return NewClaim(id, ct, core.EngineeringSubjectRef{}, scope, outcome, method, nil, evidence, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero scope", func() (Claim, error) {
			return NewClaim(id, ct, subject, core.Scope{}, outcome, method, nil, evidence, ts, prov)
		}, core.ErrInvalidScope},
		{"zero outcome", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, core.ClaimOutcome{}, method, nil, evidence, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero method", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, outcome, core.ValidationMethod{}, nil, evidence, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero criterion element", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, outcome, method, []core.CriterionRef{{}}, evidence, ts, prov)
		}, ErrInvalidValidationClaim},
		{"empty evidence", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, outcome, method, nil, nil, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero evidence element", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, outcome, method, nil, []core.EvidenceArtifactRevisionRef{{}}, ts, prov)
		}, ErrInvalidValidationClaim},
		{"zero timestamp", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, outcome, method, nil, evidence, core.Timestamp{}, prov)
		}, ErrInvalidValidationClaim},
		{"zero provenance", func() (Claim, error) {
			return NewClaim(id, ct, subject, scope, outcome, method, nil, evidence, ts, core.Provenance{})
		}, ErrInvalidValidationClaim},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.call()
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestNewClaimAcceptsEveryPredeclaredOutcome(t *testing.T) {
	for _, outcome := range []core.ClaimOutcome{
		core.ClaimOutcomeSatisfied, core.ClaimOutcomeNotSatisfied, core.ClaimOutcomeInconclusive,
	} {
		if _, err := NewClaim(
			mustClaimID(t, "VC-1"), core.ClaimTypeQuality, mustSubject(t, "AR-42", "REV-3"),
			mustScope(t, "deployment", "region=eu"), outcome, mustMethod(t, "test"),
			nil, []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
			mustTimestamp(t, "2026-07-27T10:20:00Z"), mustProvenance(t),
		); err != nil {
			t.Errorf("outcome %s rejected: %v", outcome, err)
		}
	}
}

func TestClaimIsZero(t *testing.T) {
	var c Claim
	if !c.IsZero() {
		t.Error("zero Claim does not report IsZero")
	}
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidValidationClaim) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidValidationClaim)
	}
	if c.Criteria() != nil || c.Evidence() != nil || c.ExecutionRecords() != nil {
		t.Error("zero Claim returns non-nil collections")
	}
}

// --- Satisfaction Claim invariants -------------------------------------------

func TestSatisfactionClaimRequiresRequirementCriterion(t *testing.T) {
	subject := mustSubject(t, "AR-42", "REV-3")

	// Zero criteria.
	if _, err := newClaimWith(t, core.ClaimTypeSatisfaction, subject, nil); !errors.Is(err, ErrInvalidSatisfactionClaim) {
		t.Errorf("zero criteria: error = %v, want %v", err, ErrInvalidSatisfactionClaim)
	}
	// Criteria present, but none of Requirement kind.
	nonRequirement := []core.CriterionRef{mustProductRuleCriterion(t, "acme", "rule-1")}
	if _, err := newClaimWith(t, core.ClaimTypeSatisfaction, subject, nonRequirement); !errors.Is(err, ErrInvalidSatisfactionClaim) {
		t.Errorf("no Requirement criterion: error = %v, want %v", err, ErrInvalidSatisfactionClaim)
	}
}

func TestSatisfactionClaimAcceptsRequirementCriterionAtEitherLevel(t *testing.T) {
	subject := mustSubject(t, "AR-42", "REV-3")

	if _, err := newClaimWith(t, core.ClaimTypeSatisfaction, subject,
		[]core.CriterionRef{mustRequirementCriterion(t, "REQ-7")}); err != nil {
		t.Errorf("Requirement identity criterion rejected: %v", err)
	}
	if _, err := newClaimWith(t, core.ClaimTypeSatisfaction, subject,
		[]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")}); err != nil {
		t.Errorf("Requirement Artifact Revision criterion rejected: %v", err)
	}
	// A mixed criteria list including a non-Requirement kind is still fine, so
	// long as at least one Requirement-kind criterion is present.
	if _, err := newClaimWith(t, core.ClaimTypeSatisfaction, subject, []core.CriterionRef{
		mustProductRuleCriterion(t, "acme", "rule-1"),
		mustCriterion(t, "REQ-7", "REV-2"),
	}); err != nil {
		t.Errorf("mixed criteria rejected: %v", err)
	}
}

// TestSatisfactionClaimSubjectCriterionIdentityConflict is the headline
// invariant: PEOS-006 forbids a Requirement being both the Satisfaction
// Claim's subject and one of its own criteria, and the comparison is
// cross-level at Requirement ArtifactID.
func TestSatisfactionClaimSubjectCriterionIdentityConflict(t *testing.T) {
	cases := []struct {
		name       string
		subject    core.EngineeringSubjectRef
		criterion  core.CriterionRef
		wantReject bool
	}{
		{
			name:       "identity subject vs identity criterion, same Requirement",
			subject:    mustRequirementSubject(t, "REQ-7"),
			criterion:  mustRequirementCriterion(t, "REQ-7"),
			wantReject: true,
		},
		{
			name:       "revision subject vs revision criterion, same Requirement and revision",
			subject:    mustRequirementRevisionSubject(t, "REQ-7", "REV-2"),
			criterion:  mustCriterion(t, "REQ-7", "REV-2"),
			wantReject: true,
		},
		{
			name:       "identity subject vs revision criterion, same Requirement",
			subject:    mustRequirementSubject(t, "REQ-7"),
			criterion:  mustCriterion(t, "REQ-7", "REV-2"),
			wantReject: true,
		},
		{
			name:       "revision subject vs identity criterion, same Requirement",
			subject:    mustRequirementRevisionSubject(t, "REQ-7", "REV-2"),
			criterion:  mustRequirementCriterion(t, "REQ-7"),
			wantReject: true,
		},
		{
			name:       "revision subject vs different revision of the same Requirement",
			subject:    mustRequirementRevisionSubject(t, "REQ-7", "REV-1"),
			criterion:  mustCriterion(t, "REQ-7", "REV-2"),
			wantReject: true,
		},
		{
			name:       "identity subject vs different Requirement criterion",
			subject:    mustRequirementSubject(t, "REQ-2"),
			criterion:  mustRequirementCriterion(t, "REQ-7"),
			wantReject: false,
		},
		{
			name:       "revision subject vs different Requirement revision criterion",
			subject:    mustRequirementRevisionSubject(t, "REQ-2", "REV-1"),
			criterion:  mustCriterion(t, "REQ-7", "REV-2"),
			wantReject: false,
		},
		{
			name:       "non-Requirement subject never conflicts",
			subject:    mustSubject(t, "AR-42", "REV-3"),
			criterion:  mustCriterion(t, "REQ-7", "REV-2"),
			wantReject: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newClaimWith(t, core.ClaimTypeSatisfaction, tc.subject, []core.CriterionRef{tc.criterion})
			if tc.wantReject {
				if !errors.Is(err, ErrInvalidSatisfactionClaim) {
					t.Errorf("error = %v, want %v", err, ErrInvalidSatisfactionClaim)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected rejection: %v", err)
			}
		})
	}
}

// TestNonSatisfactionClaimMayUseSameRequirementAsSubjectAndCriterion locks
// PEOS-006's explicit carve-out: the prohibition applies specifically to
// Satisfaction Claims. A general Claim MAY evaluate a Requirement as an
// engineering subject for other purposes, such as statement quality.
func TestNonSatisfactionClaimMayUseSameRequirementAsSubjectAndCriterion(t *testing.T) {
	for _, claimType := range []core.ClaimType{
		core.ClaimTypeQuality, core.ClaimTypeCompliance, core.ClaimTypeTemplateConformance,
	} {
		t.Run(claimType.String(), func(t *testing.T) {
			if _, err := newClaimWith(t, claimType, mustRequirementSubject(t, "REQ-7"),
				[]core.CriterionRef{mustRequirementCriterion(t, "REQ-7")}); err != nil {
				t.Errorf("rejected, but PEOS-006 permits this for a non-Satisfaction Claim: %v", err)
			}
		})
	}
	// Conformance also permits it -- its only extra rule is non-emptiness.
	if _, err := newClaimWith(t, core.ClaimTypeConformance, mustRequirementSubject(t, "REQ-7"),
		[]core.CriterionRef{mustRequirementCriterion(t, "REQ-7")}); err != nil {
		t.Errorf("conformance rejected: %v", err)
	}
}

// --- Conformance Claim invariants --------------------------------------------

func TestConformanceClaimRequiresAtLeastOneCriterion(t *testing.T) {
	if _, err := newClaimWith(t, core.ClaimTypeConformance, mustSubject(t, "AR-42", "REV-3"), nil); !errors.Is(err, ErrInvalidConformanceClaim) {
		t.Errorf("zero criteria: error = %v, want %v", err, ErrInvalidConformanceClaim)
	}
}

// TestConformanceClaimImposesNoCriterionKindRestriction locks the deliberate
// asymmetry with Satisfaction: PEOS-006's permitted conformance-criterion set
// ends in "or other explicitly identified conformance rules", an open set with
// no closed mapping onto core.CriterionRef's kinds, so no kind check is
// enforceable.
func TestConformanceClaimImposesNoCriterionKindRestriction(t *testing.T) {
	cases := map[string]core.CriterionRef{
		"requirement identity":  mustRequirementCriterion(t, "REQ-7"),
		"requirement revision":  mustCriterion(t, "REQ-7", "REV-2"),
		"product rule":          mustProductRuleCriterion(t, "acme", "coding-standard"),
		"opaque forward-compat": mustOpaqueCriterion(t, "future_conformance_rule", "acme", "rule-9"),
	}
	for name, criterion := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := newClaimWith(t, core.ClaimTypeConformance, mustSubject(t, "AR-42", "REV-3"),
				[]core.CriterionRef{criterion}); err != nil {
				t.Errorf("criterion kind %q rejected: %v", criterion.Kind(), err)
			}
		})
	}
}

// --- other Claim Types -------------------------------------------------------

// TestOtherClaimTypesAcceptZeroCriteria locks that PEOS-006 imposes no
// criteria cardinality rule on Quality, Compliance, Template-Conformance, or a
// Product-defined Claim Type. PEOS-007/008/009 may add their own rules in
// their own packets; none is inferred here.
func TestOtherClaimTypesAcceptZeroCriteria(t *testing.T) {
	productDefined := core.NewClaimType(mustVocab(t, "acme", "supplier-attestation"))
	for _, claimType := range []core.ClaimType{
		core.ClaimTypeQuality,
		core.ClaimTypeCompliance,
		core.ClaimTypeTemplateConformance,
		productDefined,
	} {
		t.Run(claimType.String(), func(t *testing.T) {
			c, err := newClaimWith(t, claimType, mustSubject(t, "AR-42", "REV-3"), nil)
			if err != nil {
				t.Fatalf("zero criteria rejected: %v", err)
			}
			if len(c.Criteria()) != 0 {
				t.Error("criteria unexpectedly non-empty")
			}
		})
	}
}

// --- WithCriteria ------------------------------------------------------------

func TestClaimWithCriteriaReplacesCollection(t *testing.T) {
	c := mustSatisfactionClaim(t)
	replaced, err := c.WithCriteria([]core.CriterionRef{
		mustCriterion(t, "REQ-8", "REV-1"),
		mustProductRuleCriterion(t, "acme", "rule-1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(replaced.Criteria()) != 2 {
		t.Errorf("Criteria() length = %d, want 2", len(replaced.Criteria()))
	}
	if len(c.Criteria()) != 1 {
		t.Error("WithCriteria mutated the receiver")
	}
}

func TestClaimWithCriteriaEmptyPerClaimType(t *testing.T) {
	t.Run("general accepted", func(t *testing.T) {
		cleared, err := mustGeneralClaim(t).WithCriteria(nil)
		if err != nil {
			t.Fatalf("empty rejected for a general Claim Type: %v", err)
		}
		if len(cleared.Criteria()) != 0 {
			t.Error("criteria not cleared")
		}
	})
	t.Run("satisfaction rejected", func(t *testing.T) {
		_, err := mustSatisfactionClaim(t).WithCriteria(nil)
		if !errors.Is(err, ErrInvalidSatisfactionClaim) {
			t.Errorf("error = %v, want %v", err, ErrInvalidSatisfactionClaim)
		}
	})
	t.Run("conformance rejected", func(t *testing.T) {
		_, err := mustConformanceClaim(t).WithCriteria([]core.CriterionRef{})
		if !errors.Is(err, ErrInvalidConformanceClaim) {
			t.Errorf("error = %v, want %v", err, ErrInvalidConformanceClaim)
		}
	})
}

// TestClaimWithCriteriaSharesValidationPathWithNewClaim proves the two entry
// points cannot drift: for each invalid input, both produce the same sentinel.
func TestClaimWithCriteriaSharesValidationPathWithNewClaim(t *testing.T) {
	cases := []struct {
		name      string
		claimType core.ClaimType
		subject   core.EngineeringSubjectRef
		criteria  []core.CriterionRef
		want      error
	}{
		{"satisfaction zero criteria", core.ClaimTypeSatisfaction, mustSubject(t, "AR-42", "REV-3"), nil, ErrInvalidSatisfactionClaim},
		{"satisfaction no requirement criterion", core.ClaimTypeSatisfaction, mustSubject(t, "AR-42", "REV-3"), []core.CriterionRef{mustProductRuleCriterion(t, "acme", "r")}, ErrInvalidSatisfactionClaim},
		{"satisfaction identity conflict", core.ClaimTypeSatisfaction, mustRequirementSubject(t, "REQ-7"), []core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")}, ErrInvalidSatisfactionClaim},
		{"conformance zero criteria", core.ClaimTypeConformance, mustSubject(t, "AR-42", "REV-3"), nil, ErrInvalidConformanceClaim},
		{"zero criterion element", core.ClaimTypeQuality, mustSubject(t, "AR-42", "REV-3"), []core.CriterionRef{{}}, ErrInvalidValidationClaim},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ctorErr := newClaimWith(t, tc.claimType, tc.subject, tc.criteria)
			if !errors.Is(ctorErr, tc.want) {
				t.Errorf("NewClaim error = %v, want %v", ctorErr, tc.want)
			}

			// Build a valid Claim of the same type and subject, then apply the
			// same invalid criteria through the modifier.
			var seed []core.CriterionRef
			switch tc.claimType {
			case core.ClaimTypeSatisfaction:
				seed = []core.CriterionRef{mustRequirementCriterion(t, "REQ-OTHER")}
			case core.ClaimTypeConformance:
				seed = []core.CriterionRef{mustProductRuleCriterion(t, "acme", "seed")}
			default:
				seed = nil
			}
			valid, err := newClaimWith(t, tc.claimType, tc.subject, seed)
			if err != nil {
				t.Fatalf("could not build a valid seed Claim: %v", err)
			}
			_, modErr := valid.WithCriteria(tc.criteria)
			if !errors.Is(modErr, tc.want) {
				t.Errorf("WithCriteria error = %v, want %v", modErr, tc.want)
			}
		})
	}
}

func TestClaimWithCriteriaFailedPreservesReceiver(t *testing.T) {
	original := mustSatisfactionClaim(t)
	if _, err := original.WithCriteria(nil); err == nil {
		t.Fatal("expected failure")
	}
	if len(original.Criteria()) != 1 {
		t.Error("failed WithCriteria disturbed the receiver")
	}
}

func TestClaimCriteriaDefensivelyCopied(t *testing.T) {
	input := []core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")}
	c, err := newClaimWith(t, core.ClaimTypeSatisfaction, mustSubject(t, "AR-42", "REV-3"), input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = core.CriterionRef{}
	if c.Criteria()[0].IsZero() {
		t.Error("criteria input slice was not copied in")
	}
	c.Criteria()[0] = core.CriterionRef{}
	if c.Criteria()[0].IsZero() {
		t.Error("Criteria() accessor did not return a defensive copy")
	}
}

func TestClaimEvidenceDefensivelyCopiedAndNotDeduplicated(t *testing.T) {
	dup := mustEvidence(t, "EV-1", "REV-1")
	input := []core.EvidenceArtifactRevisionRef{dup, dup}
	c, err := NewClaim(
		mustClaimID(t, "VC-1"), core.ClaimTypeQuality, mustSubject(t, "AR-42", "REV-3"),
		mustScope(t, "deployment", "region=eu"), core.ClaimOutcomeSatisfied, mustMethod(t, "test"),
		nil, input, mustTimestamp(t, "2026-07-27T10:20:00Z"), mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Evidence()) != 2 {
		t.Errorf("Evidence() length = %d, want 2 (no deduplication)", len(c.Evidence()))
	}
	input[0] = core.EvidenceArtifactRevisionRef{}
	if c.Evidence()[0].IsZero() {
		t.Error("evidence input slice was not copied in")
	}
	c.Evidence()[0] = core.EvidenceArtifactRevisionRef{}
	if c.Evidence()[0].IsZero() {
		t.Error("Evidence() accessor did not return a defensive copy")
	}
}

// --- other modifiers ---------------------------------------------------------

func TestClaimExecutionRecordsModifier(t *testing.T) {
	c, err := mustGeneralClaim(t).WithExecutionRecords([]core.ValidationExecutionRecordRef{
		mustExecutionRecordRef(t, "EXR-1"), mustExecutionRecordRef(t, "EXR-2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.ExecutionRecords()) != 2 {
		t.Errorf("ExecutionRecords() length = %d, want 2", len(c.ExecutionRecords()))
	}
	cleared, err := c.WithExecutionRecords(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.ExecutionRecords()) != 0 {
		t.Error("empty input did not clear execution records")
	}
	if _, err := c.WithExecutionRecords([]core.ValidationExecutionRecordRef{{}}); !errors.Is(err, ErrInvalidValidationClaim) {
		t.Error("zero execution record reference accepted")
	}
	// Defensive copy out.
	c.ExecutionRecords()[0] = core.ValidationExecutionRecordRef{}
	if c.ExecutionRecords()[0].IsZero() {
		t.Error("ExecutionRecords() accessor did not return a defensive copy")
	}
}

func TestClaimReasoningModifier(t *testing.T) {
	c, err := mustGeneralClaim(t).WithReasoning("  because the evidence shows it  ")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := c.Reasoning()
	if !ok || got != "because the evidence shows it" {
		t.Errorf("Reasoning() = %q, %v", got, ok)
	}
	if _, ok := c.WithoutReasoning().Reasoning(); ok {
		t.Error("WithoutReasoning did not clear")
	}
	for _, bad := range []string{"", "  ", "\t"} {
		if _, err := mustGeneralClaim(t).WithReasoning(bad); !errors.Is(err, ErrInvalidValidationClaim) {
			t.Errorf("reasoning %q: error = %v, want %v", bad, err, ErrInvalidValidationClaim)
		}
	}
}

func TestClaimAuthorityModifier(t *testing.T) {
	auth := mustAuthorityRef(t, "org", "qa-board")
	c, err := mustGeneralClaim(t).WithAuthority(auth)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := c.Authority()
	if !ok || got != auth {
		t.Errorf("Authority() = %v, %v", got, ok)
	}
	if _, ok := c.WithoutAuthority().Authority(); ok {
		t.Error("WithoutAuthority did not clear")
	}
	if _, err := mustGeneralClaim(t).WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidValidationClaim) {
		t.Error("zero authority accepted")
	}
}

func TestClaimCorrectionModifier(t *testing.T) {
	for _, kind := range []core.CorrectionKind{core.CorrectionKindCorrect, core.CorrectionKindReplace, core.CorrectionKindInvalidate} {
		correction, err := core.NewRecordCorrectionRef(kind, mustClaimRef(t, "VC-0"))
		if err != nil {
			t.Fatal(err)
		}
		c, err := mustGeneralClaim(t).WithCorrection(correction)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := c.Correction()
		if !ok {
			t.Fatal("Correction() reported false")
		}
		if !got.Kind().Value().Equal(kind.Value()) {
			t.Errorf("kind = %v, want %v", got.Kind(), kind)
		}
		if got.Target().ClaimID().String() != "VC-0" {
			t.Errorf("target = %v", got.Target())
		}
		if _, ok := c.WithoutCorrection().Correction(); ok {
			t.Error("WithoutCorrection did not clear")
		}
	}
	if _, err := mustGeneralClaim(t).WithCorrection(core.RecordCorrectionRef[core.ValidationClaimRef]{}); !errors.Is(err, ErrInvalidValidationClaim) {
		t.Error("zero correction accepted")
	}
}

func TestClaimExtensionModifier(t *testing.T) {
	c := mustGeneralClaim(t).WithExtension(mustExtension(t, "acme", `{"k":1}`))
	if c.Extension().IsZero() {
		t.Error("WithExtension did not set")
	}
	if !c.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear")
	}
}

func TestClaimModifierReceiverImmutability(t *testing.T) {
	original := mustGeneralClaim(t)
	if _, err := original.WithExecutionRecords([]core.ValidationExecutionRecordRef{mustExecutionRecordRef(t, "EXR-1")}); err != nil {
		t.Fatal(err)
	}
	if len(original.ExecutionRecords()) != 0 {
		t.Error("WithExecutionRecords mutated the receiver")
	}
	if _, err := original.WithReasoning("x"); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Reasoning(); ok {
		t.Error("WithReasoning mutated the receiver")
	}
	if _, err := original.WithAuthority(mustAuthorityRef(t, "org", "b")); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Authority(); ok {
		t.Error("WithAuthority mutated the receiver")
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, mustClaimRef(t, "VC-0"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := original.WithCorrection(correction); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Correction(); ok {
		t.Error("WithCorrection mutated the receiver")
	}
	_ = original.WithExtension(mustExtension(t, "acme", `{}`))
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}
}

// TestClaimCorrectionDoesNotMutateOriginal proves historical preservation:
// recording a correcting Claim leaves the corrected Claim's value untouched.
func TestClaimCorrectionDoesNotMutateOriginal(t *testing.T) {
	earlier := mustGeneralClaim(t)
	before, err := json.Marshal(earlier)
	if err != nil {
		t.Fatal(err)
	}

	later, err := NewClaim(
		mustClaimID(t, "VC-2"), core.ClaimTypeQuality, mustSubject(t, "AR-42", "REV-3"),
		mustScope(t, "deployment", "region=eu"), core.ClaimOutcomeNotSatisfied, mustMethod(t, "test"),
		nil, []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-2", "REV-1")},
		mustTimestamp(t, "2026-07-27T12:00:00Z"), mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	earlierRef, err := earlier.Ref()
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindInvalidate, earlierRef)
	if err != nil {
		t.Fatal(err)
	}
	if later, err = later.WithCorrection(correction); err != nil {
		t.Fatal(err)
	}

	after, err := json.Marshal(earlier)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("recording a correction altered the corrected Claim")
	}
	if _, ok := earlier.Correction(); ok {
		t.Error("the corrected Claim gained a correction reference")
	}
	if !earlier.Outcome().Value().Equal(core.ClaimOutcomeSatisfied.Value()) {
		t.Error("the corrected Claim's outcome changed")
	}

	// The correction lives on the new Claim and points backward at the earlier
	// one -- that direction is what keeps the earlier record untouched.
	got, ok := later.Correction()
	if !ok {
		t.Fatal("the correcting Claim does not carry a correction reference")
	}
	if got.Target().ClaimID() != earlier.ID() {
		t.Errorf("correction target = %v, want %v", got.Target().ClaimID(), earlier.ID())
	}
	if !got.Kind().Value().Equal(core.CorrectionKindInvalidate.Value()) {
		t.Errorf("correction kind = %v, want invalidate", got.Kind())
	}
}

// --- JSON --------------------------------------------------------------------

func claimPayload(t *testing.T, overrides map[string]string) string {
	t.Helper()
	base := map[string]string{
		"id":         `"VC-1"`,
		"claim_type": `"peos:quality"`,
		"subject":    `{"kind":"artifact_revision","ref":{"artifact_id":"AR-42","revision_id":"REV-3"}}`,
		"scope":      `{"kind":"peos:deployment","expression":"region=eu"}`,
		"outcome":    `"peos:satisfied"`,
		"method":     `"peos:test"`,
		"evidence":   `[{"artifact_id":"EV-1","revision_id":"REV-1"}]`,
		"timestamp":  `"2026-07-27T10:20:00Z"`,
		"provenance": `{"actor":{"namespace":"peos-cli","identifier":"svc-1"},"recorded_at":"2026-07-27T00:00:00Z"}`,
	}
	for k, v := range overrides {
		if v == "" {
			delete(base, k)
			continue
		}
		base[k] = v
	}
	parts := make([]string, 0, len(base))
	for k, v := range base {
		parts = append(parts, `"`+k+`":`+v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func TestClaimJSONMinimumKeys(t *testing.T) {
	data, err := json.Marshal(mustGeneralClaim(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"id", "claim_type", "subject", "scope", "outcome", "method", "evidence", "timestamp", "provenance"}
	if len(raw) != len(want) {
		t.Errorf("minimum wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	forbidden := []string{
		"basis", "verdict", "status", "lifecycle", "state", "relation", "waiver", "type",
		"satisfied", "conformant", "compliant", "validated", "quality", "qualityScore",
		"artifact_id", "revision_id", "core", "criteria", "reasoning", "extension",
	}
	for _, k := range forbidden {
		if _, ok := raw[k]; ok {
			t.Errorf("minimum wire form unexpectedly carries %q", k)
		}
	}
}

func TestClaimJSONFullKeysAndEquivalence(t *testing.T) {
	original := fullClaim(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"id", "claim_type", "subject", "scope", "outcome", "method",
		"criteria", "evidence", "timestamp", "provenance",
		"execution_records", "reasoning", "authority", "correction", "extension",
	}
	if len(raw) != len(want) {
		t.Errorf("full wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, k := range []string{"basis", "verdict", "status", "lifecycle", "relation", "waiver", "type"} {
		if _, ok := raw[k]; ok {
			t.Errorf("full wire form unexpectedly carries %q", k)
		}
	}

	var decoded Claim
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(again) {
		t.Errorf("constructor/Unmarshal equivalence broken:\n%s\n%s", data, again)
	}
}

// TestClaimJSONCriteriaGrid is the full absent / null / [] / valid matrix for
// criteria across a general, a Satisfaction, and a Conformance Claim Type.
// Unlike PlannedActivity and ExecutionRecord, a Claim must distinguish absent
// from explicit null, and empty must fail for two Claim Types.
func TestClaimJSONCriteriaGrid(t *testing.T) {
	validRequirement := `[{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7","revision_id":"REV-2"}}]`

	cases := []struct {
		claimType string
		criteria  string
		label     string
		want      error // nil means accepted
	}{
		// General Claim Type (quality): zero criteria are fine.
		{`"peos:quality"`, "", "absent", nil},
		{`"peos:quality"`, "[]", "empty array", nil},
		{`"peos:quality"`, "null", "explicit null", ErrInvalidValidationClaim},
		{`"peos:quality"`, validRequirement, "valid", nil},

		// Satisfaction: needs at least one Requirement-kind criterion.
		{`"peos:satisfaction"`, "", "absent", ErrInvalidSatisfactionClaim},
		{`"peos:satisfaction"`, "[]", "empty array", ErrInvalidSatisfactionClaim},
		{`"peos:satisfaction"`, "null", "explicit null", ErrInvalidValidationClaim},
		{`"peos:satisfaction"`, validRequirement, "valid", nil},

		// Conformance: needs at least one criterion of any kind.
		{`"peos:conformance"`, "", "absent", ErrInvalidConformanceClaim},
		{`"peos:conformance"`, "[]", "empty array", ErrInvalidConformanceClaim},
		{`"peos:conformance"`, "null", "explicit null", ErrInvalidValidationClaim},
		{`"peos:conformance"`, validRequirement, "valid", nil},
	}
	for _, tc := range cases {
		t.Run(tc.claimType+"/"+tc.label, func(t *testing.T) {
			payload := claimPayload(t, map[string]string{
				"claim_type": tc.claimType,
				"criteria":   tc.criteria,
			})
			var c Claim
			err := json.Unmarshal([]byte(payload), &c)
			if tc.want == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

// TestClaimJSONEvidenceAbsentNullEmptyAllRejected locks that, unlike criteria,
// the three evidence inputs converge on the same one-or-more rejection.
func TestClaimJSONEvidenceAbsentNullEmptyAllRejected(t *testing.T) {
	for _, value := range []string{"", "null", "[]"} {
		label := value
		if value == "" {
			label = "absent"
		}
		t.Run(label, func(t *testing.T) {
			var c Claim
			err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"evidence": value})), &c)
			if !errors.Is(err, ErrInvalidValidationClaim) {
				t.Errorf("error = %v, want %v", err, ErrInvalidValidationClaim)
			}
		})
	}
}

func TestClaimJSONMandatoryMissingRejected(t *testing.T) {
	cases := map[string]error{
		"id":         ErrInvalidValidationClaim,
		"claim_type": ErrInvalidValidationClaim,
		"subject":    ErrInvalidValidationClaim,
		"scope":      core.ErrInvalidScope,
		"outcome":    ErrInvalidValidationClaim,
		"method":     ErrInvalidValidationClaim,
		"timestamp":  ErrInvalidValidationClaim,
		"provenance": ErrInvalidValidationClaim,
	}
	for field, want := range cases {
		t.Run(field, func(t *testing.T) {
			var c Claim
			err := json.Unmarshal([]byte(claimPayload(t, map[string]string{field: ""})), &c)
			if !errors.Is(err, want) {
				t.Errorf("error = %v, want %v", err, want)
			}
		})
	}
}

func TestClaimJSONMandatoryNullRejected(t *testing.T) {
	for _, field := range []string{"id", "claim_type", "subject", "scope", "outcome", "method", "evidence", "timestamp", "provenance"} {
		t.Run(field, func(t *testing.T) {
			var c Claim
			if err := json.Unmarshal([]byte(claimPayload(t, map[string]string{field: "null"})), &c); err == nil {
				t.Error("explicit null accepted, want error")
			}
		})
	}
}

// TestClaimJSONSatisfactionIdentityConflictRejectedOnDecode proves the
// cross-level comparison is re-run on decode, not only in the constructor.
func TestClaimJSONSatisfactionIdentityConflictRejectedOnDecode(t *testing.T) {
	cases := map[string][2]string{
		"identity subject vs revision criterion": {
			`{"kind":"requirement","ref":{"artifact_id":"REQ-7"}}`,
			`[{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7","revision_id":"REV-2"}}]`,
		},
		"revision subject vs identity criterion": {
			`{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7","revision_id":"REV-2"}}`,
			`[{"kind":"requirement","ref":{"artifact_id":"REQ-7"}}]`,
		},
		"revision subject vs different revision": {
			`{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7","revision_id":"REV-1"}}`,
			`[{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7","revision_id":"REV-2"}}]`,
		},
	}
	for name, pair := range cases {
		t.Run(name, func(t *testing.T) {
			payload := claimPayload(t, map[string]string{
				"claim_type": `"peos:satisfaction"`,
				"subject":    pair[0],
				"criteria":   pair[1],
			})
			var c Claim
			err := json.Unmarshal([]byte(payload), &c)
			if !errors.Is(err, ErrInvalidSatisfactionClaim) {
				t.Errorf("error = %v, want %v", err, ErrInvalidSatisfactionClaim)
			}
		})
	}
}

func TestClaimJSONOptionalSingleNullRejected(t *testing.T) {
	for _, field := range []string{"reasoning", "authority", "correction"} {
		t.Run(field, func(t *testing.T) {
			var c Claim
			err := json.Unmarshal([]byte(claimPayload(t, map[string]string{field: "null"})), &c)
			if !errors.Is(err, ErrInvalidValidationClaim) {
				t.Errorf("error = %v, want %v", err, ErrInvalidValidationClaim)
			}
		})
	}
}

func TestClaimJSONExecutionRecordsEquivalent(t *testing.T) {
	for _, value := range []string{"", "null", "[]"} {
		label := value
		if value == "" {
			label = "absent"
		}
		t.Run(label, func(t *testing.T) {
			var c Claim
			if err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"execution_records": value})), &c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(c.ExecutionRecords()) != 0 {
				t.Error("execution records unexpectedly non-empty")
			}
		})
	}
}

func TestClaimJSONExtensionNullMeansAbsent(t *testing.T) {
	var c Claim
	if err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"extension": "null"})), &c); err != nil {
		t.Fatalf("extension null rejected: %v", err)
	}
	if !c.Extension().IsZero() {
		t.Error("extension null did not decode as absent")
	}
}

func TestClaimJSONNestedSentinelsPreserved(t *testing.T) {
	cases := []struct {
		name    string
		payload map[string]string
		want    error
	}{
		{"subject missing revision", map[string]string{"subject": `{"kind":"artifact_revision","ref":{"artifact_id":"AR-42"}}`}, core.ErrMissingRevisionID},
		{"subject missing discriminator", map[string]string{"subject": `{"ref":{"artifact_id":"AR-42"}}`}, core.ErrInvalidReferenceDiscriminator},
		{"criterion missing revision", map[string]string{"criteria": `[{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7"}}]`}, core.ErrMissingRevisionID},
		{"criterion missing discriminator", map[string]string{"criteria": `[{"ref":{"artifact_id":"REQ-7"}}]`}, core.ErrInvalidReferenceDiscriminator},
		{"evidence missing revision", map[string]string{"evidence": `[{"artifact_id":"EV-1"}]`}, core.ErrMissingRevisionID},
		{"scope empty expression", map[string]string{"scope": `{"kind":"peos:deployment","expression":""}`}, core.ErrInvalidScope},
		{"claim_type malformed", map[string]string{"claim_type": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"outcome malformed", map[string]string{"outcome": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"method malformed", map[string]string{"method": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"id empty identity", map[string]string{"id": `"  "`}, core.ErrEmptyIdentity},
		{"authority empty identity", map[string]string{"authority": `{"namespace":"org","identifier":"  "}`}, core.ErrEmptyIdentity},
		{"execution record empty identity", map[string]string{"execution_records": `[{"record_id":"  "}]`}, core.ErrEmptyIdentity},
		{"correction empty target", map[string]string{"correction": `{"kind":"peos:replace","target":{"claim_id":"  "}}`}, core.ErrEmptyIdentity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Claim
			err := json.Unmarshal([]byte(claimPayload(t, tc.payload)), &c)
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want wrapping %v", err, tc.want)
			}
		})
	}
}

func TestClaimJSONCriteriaWrongTypeRejected(t *testing.T) {
	var c Claim
	err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"criteria": `"not-an-array"`})), &c)
	if !errors.Is(err, ErrInvalidValidationClaim) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationClaim)
	}
}

func TestClaimJSONReasoningWrongTypeAndEmptyRejected(t *testing.T) {
	for _, value := range []string{`123`, `"   "`} {
		var c Claim
		err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"reasoning": value})), &c)
		if !errors.Is(err, ErrInvalidValidationClaim) {
			t.Errorf("reasoning %s: error = %v, want %v", value, err, ErrInvalidValidationClaim)
		}
	}
}

func TestClaimJSONMalformedDocumentRejected(t *testing.T) {
	for _, payload := range []string{`not json`, `[]`, `{"id":[]}`} {
		var c Claim
		if err := json.Unmarshal([]byte(payload), &c); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

func TestClaimJSONUnknownFieldIgnored(t *testing.T) {
	var c Claim
	if err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"unknown": `"x"`})), &c); err != nil {
		t.Fatalf("unknown field rejected: %v", err)
	}
}

func TestClaimFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := fullClaim(t)
	before := receiver.ID()
	if err := json.Unmarshal([]byte(claimPayload(t, map[string]string{"outcome": "null"})), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.ID() != before {
		t.Error("failed Unmarshal disturbed the receiver")
	}
	if len(receiver.Criteria()) != 1 {
		t.Error("failed Unmarshal disturbed the receiver's criteria")
	}
}

// --- distinction between execution and claim outcomes ------------------------

// TestExecutionAndClaimOutcomesAreIndependent proves the two outcome
// vocabularies do not constrain each other: a completed execution may back a
// not-satisfied Claim.
func TestExecutionAndClaimOutcomesAreIndependent(t *testing.T) {
	record := mustExecutionRecord(t) // ExecutionOutcomeCompleted
	claim, err := NewClaim(
		mustClaimID(t, "VC-1"), core.ClaimTypeQuality, mustSubject(t, "AR-42", "REV-3"),
		mustScope(t, "deployment", "region=eu"), core.ClaimOutcomeNotSatisfied, mustMethod(t, "test"),
		nil, []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")},
		mustTimestamp(t, "2026-07-27T10:20:00Z"), mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := record.Ref()
	if err != nil {
		t.Fatal(err)
	}
	claim, err = claim.WithExecutionRecords([]core.ValidationExecutionRecordRef{ref})
	if err != nil {
		t.Fatal(err)
	}
	if !record.Outcome().Equal(core.ExecutionOutcomeCompleted) {
		t.Error("execution outcome changed")
	}
	if !claim.Outcome().Value().Equal(core.ClaimOutcomeNotSatisfied.Value()) {
		t.Error("claim outcome changed")
	}
	if reflect.TypeOf(record.Outcome()) == reflect.TypeOf(claim.Outcome()) {
		t.Error("execution and claim outcomes share a Go type")
	}
}

// --- structural absence ------------------------------------------------------

func TestClaimHasNoForbiddenAPI(t *testing.T) {
	forbidden := []string{
		"Core", "ArtifactID", "RevisionID",
		"Relation", "Source", "Target",
		"Lifecycle", "State", "Status",
		"Basis", "Verdict",
		"WithID", "WithClaimType", "WithSubject", "WithScope",
		"WithOutcome", "WithMethod", "WithoutCriteria",
		"WithEvidence", "WithTimestamp", "WithProvenance",
		"WithoutEvidence", "WithoutSubject", "WithoutScope", "WithoutOutcome",
		"Waiver", "Satisfied", "Conformant", "Compliant",
	}
	assertNoMethods(t, "Claim", reflect.TypeOf(Claim{}), forbidden)
}

// TestPackageDeclaresNoNonPEOS006Entities proves the package introduces none
// of the entities PEOS-006 explicitly refuses to define, and no per-Claim-Type
// Go type. A declared type would be reachable as a zero value of a named type
// in this package; the reflect-based method audits above cover the API surface,
// and this test documents and checks the type-level claim by constructing the
// full set of Claim Types through the one Claim type.
func TestPackageDeclaresNoNonPEOS006Entities(t *testing.T) {
	// All five PEOS-006/007/008/009 Claim specializations are values of one
	// Go type, so a single Claim can take each of them.
	claimTypes := []core.ClaimType{
		core.ClaimTypeSatisfaction, core.ClaimTypeConformance,
		core.ClaimTypeQuality, core.ClaimTypeCompliance, core.ClaimTypeTemplateConformance,
	}
	for _, ct := range claimTypes {
		criteria := []core.CriterionRef(nil)
		switch ct {
		case core.ClaimTypeSatisfaction:
			criteria = []core.CriterionRef{mustRequirementCriterion(t, "REQ-7")}
		case core.ClaimTypeConformance:
			criteria = []core.CriterionRef{mustProductRuleCriterion(t, "acme", "r")}
		}
		c, err := newClaimWith(t, ct, mustSubject(t, "AR-42", "REV-3"), criteria)
		if err != nil {
			t.Fatalf("claim type %s not representable by the single Claim type: %v", ct, err)
		}
		if got := reflect.TypeOf(c).Name(); got != "Claim" {
			t.Errorf("claim type %s produced Go type %q, want Claim", ct, got)
		}
	}
}

// TestClaimAllAccessorsReturnConstructorInputs exercises every accessor
// against known inputs, including the mandatory-value accessors not otherwise
// touched by the behavioral tests above.
func TestClaimAllAccessorsReturnConstructorInputs(t *testing.T) {
	id := mustClaimID(t, "VC-9")
	claimType := core.ClaimTypeConformance
	subject := mustSubject(t, "AR-42", "REV-3")
	scope := mustScope(t, "deployment", "region=eu")
	outcome := core.ClaimOutcomeInconclusive
	method := mustMethod(t, "review")
	criteria := []core.CriterionRef{mustProductRuleCriterion(t, "acme", "standard-1")}
	evidence := []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")}
	ts := mustTimestamp(t, "2026-07-27T10:20:00Z")
	prov := mustProvenance(t)

	c, err := NewClaim(id, claimType, subject, scope, outcome, method, criteria, evidence, ts, prov)
	if err != nil {
		t.Fatal(err)
	}

	if c.ID() != id {
		t.Error("ID() mismatch")
	}
	if !c.ClaimType().Value().Equal(claimType.Value()) {
		t.Error("ClaimType() mismatch")
	}
	if c.Subject() != subject {
		t.Error("Subject() mismatch")
	}
	if !c.Scope().Equal(scope) {
		t.Error("Scope() mismatch")
	}
	if !c.Outcome().Value().Equal(outcome.Value()) {
		t.Error("Outcome() mismatch")
	}
	if !c.Method().Value().Equal(method.Value()) {
		t.Error("Method() mismatch")
	}
	if len(c.Criteria()) != 1 || c.Criteria()[0].Kind() != core.CriterionKindProductRule {
		t.Error("Criteria() mismatch")
	}
	if len(c.Evidence()) != 1 || c.Evidence()[0].ArtifactID().String() != "EV-1" {
		t.Error("Evidence() mismatch")
	}
	if !c.Timestamp().Equal(ts) {
		t.Error("Timestamp() mismatch")
	}
	if c.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
}
