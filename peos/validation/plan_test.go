package validation

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- Plan helpers ------------------------------------------------------------

func mustPlanArtifact(t *testing.T, id string) core.Artifact {
	t.Helper()
	a, err := core.NewArtifact(mustArtifactID(t, id), ArtifactTypeValidationPlan)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustPlan(t *testing.T) Plan {
	t.Helper()
	p, err := NewPlan(mustPlanArtifact(t, "VP-1"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func mustOrigin(t *testing.T) core.Origin {
	t.Helper()
	o, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustIntegrity(t *testing.T) core.IntegrityIdentity {
	t.Helper()
	i, err := core.NewIntegrityIdentity(core.IntegrityMechanismCryptographicDigest, "sha256:abc123", core.IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func mustCoreRevision(t *testing.T, artifactID, revisionID string) core.ArtifactRevision {
	t.Helper()
	r, err := core.NewArtifactRevision(
		mustArtifactID(t, artifactID),
		mustArtifactRevisionID(t, revisionID),
		mustOrigin(t),
		mustProvenance(t),
		mustIntegrity(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func mustPlanContent(t *testing.T) PlanContent {
	t.Helper()
	c, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{mustActivity(t)},
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustPlanRevision(t *testing.T) PlanRevision {
	t.Helper()
	r, err := NewPlanRevision(mustPlan(t), mustCoreRevision(t, "VP-1", "REV-1"), mustPlanContent(t))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// --- ArtifactTypeValidationPlan ---------------------------------------------

func TestArtifactTypeValidationPlanValue(t *testing.T) {
	if got := ArtifactTypeValidationPlan.String(); got != "peos:validation-plan" {
		t.Errorf("ArtifactTypeValidationPlan = %q, want peos:validation-plan", got)
	}
}

// --- Plan --------------------------------------------------------------------

func TestNewPlanValid(t *testing.T) {
	p := mustPlan(t)
	if p.IsZero() {
		t.Fatal("valid Plan reports IsZero")
	}
	if got := p.ID().String(); got != "VP-1" {
		t.Errorf("ID() = %q, want VP-1", got)
	}
	if p.Core().Type() != ArtifactTypeValidationPlan {
		t.Error("Core() lost the artifact type")
	}
	ref, err := p.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != p.ID() {
		t.Error("Ref() does not identify the Plan")
	}
}

func TestNewPlanZeroArtifactRejected(t *testing.T) {
	_, err := NewPlan(core.Artifact{})
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestNewPlanArtifactTypeMismatchRejected(t *testing.T) {
	other, err := core.NewArtifact(mustArtifactID(t, "VP-1"), core.NewArtifactType(mustVocab(t, core.PEOSNamespace, "requirement")))
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewPlan(other)
	if !errors.Is(err, ErrValidationPlanArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrValidationPlanArtifactTypeMismatch)
	}
}

func TestPlanIsZero(t *testing.T) {
	var p Plan
	if !p.IsZero() {
		t.Error("zero Plan does not report IsZero")
	}
}

func TestPlanJSONRoundTripAndKeys(t *testing.T) {
	data, err := json.Marshal(mustPlan(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["artifact_type"]; !ok {
		t.Error("Plan wire form lost artifact_type")
	}
	for _, forbidden := range []string{"relation", "lifecycle", "state", "version", "content"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("Plan wire form unexpectedly carries %q", forbidden)
		}
	}
	var decoded Plan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != mustPlan(t).ID() {
		t.Error("round-trip lost identity")
	}
}

func TestPlanZeroMarshalRejected(t *testing.T) {
	var p Plan
	if _, err := json.Marshal(p); !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestPlanJSONWrongTypeRejected(t *testing.T) {
	var p Plan
	err := json.Unmarshal([]byte(`{"artifact_id":"VP-1","artifact_type":"peos:requirement"}`), &p)
	if !errors.Is(err, ErrValidationPlanArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrValidationPlanArtifactTypeMismatch)
	}
}

func TestPlanJSONNullAndMissingRejected(t *testing.T) {
	for _, payload := range []string{`null`, `{}`, `{"artifact_id":"VP-1"}`, `{"artifact_type":"peos:validation-plan"}`} {
		var p Plan
		if err := json.Unmarshal([]byte(payload), &p); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

func TestPlanFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := mustPlan(t)
	if err := json.Unmarshal([]byte(`{"artifact_id":"VP-1","artifact_type":"peos:requirement"}`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.Core().Type() != ArtifactTypeValidationPlan {
		t.Error("failed Unmarshal disturbed the receiver")
	}
}

// --- PlanApplicability -------------------------------------------------------

func TestPlanApplicabilityUnrestrictedIsNonZero(t *testing.T) {
	a := NewUnrestrictedPlanApplicability()
	if a.IsZero() {
		t.Fatal("explicit unrestricted applicability reports IsZero")
	}
	if !a.IsUnrestricted() || a.IsScoped() {
		t.Error("unrestricted applicability misreports its variant")
	}
	if _, ok := a.Scope(); ok {
		t.Error("unrestricted applicability reports a scope")
	}
	if got := a.Kind(); got != "unrestricted" {
		t.Errorf("Kind() = %q, want unrestricted", got)
	}
}

func TestPlanApplicabilityZeroIsInvalid(t *testing.T) {
	var a PlanApplicability
	if !a.IsZero() {
		t.Error("zero PlanApplicability does not report IsZero")
	}
	if a.IsUnrestricted() {
		t.Error("zero PlanApplicability reports IsUnrestricted — unstated must differ from unrestricted")
	}
	if got := a.Kind(); got != "" {
		t.Errorf("zero Kind() = %q, want empty", got)
	}
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidPlanApplicability) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidPlanApplicability)
	}
}

func TestNewScopedPlanApplicabilityValid(t *testing.T) {
	scope := mustScope(t, "deployment", "region=eu")
	a, err := NewScopedPlanApplicability(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !a.IsScoped() || a.IsUnrestricted() {
		t.Error("scoped applicability misreports its variant")
	}
	got, ok := a.Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("Scope() = %v, %v", got, ok)
	}
}

func TestNewScopedPlanApplicabilityZeroScopeRejected(t *testing.T) {
	_, err := NewScopedPlanApplicability(core.Scope{})
	if !errors.Is(err, ErrInvalidPlanApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlanApplicability)
	}
}

func TestPlanApplicabilityJSONKeys(t *testing.T) {
	unrestricted, err := json.Marshal(NewUnrestrictedPlanApplicability())
	if err != nil {
		t.Fatal(err)
	}
	if string(unrestricted) != `{"kind":"unrestricted"}` {
		t.Errorf("unrestricted wire form = %s", unrestricted)
	}

	scoped, err := NewScopedPlanApplicability(mustScope(t, "deployment", "region=eu"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(scoped)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != 2 {
		t.Errorf("scoped wire form has %d keys, want 2: %v", len(raw), raw)
	}
	for _, k := range []string{"kind", "scope"} {
		if _, ok := raw[k]; !ok {
			t.Errorf("scoped wire form missing %q", k)
		}
	}
	if _, ok := raw["type"]; ok {
		t.Error("PlanApplicability must not carry a top-level type discriminator")
	}
}

func TestPlanApplicabilityJSONRoundTrip(t *testing.T) {
	scoped, err := NewScopedPlanApplicability(mustScope(t, "deployment", "region=eu"))
	if err != nil {
		t.Fatal(err)
	}
	for name, original := range map[string]PlanApplicability{
		"unrestricted": NewUnrestrictedPlanApplicability(),
		"scoped":       scoped,
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decoded PlanApplicability
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != original.Kind() {
				t.Errorf("kind mismatch: %q vs %q", decoded.Kind(), original.Kind())
			}
			gotScope, gotOK := decoded.Scope()
			wantScope, wantOK := original.Scope()
			if gotOK != wantOK || !gotScope.Equal(wantScope) {
				t.Error("scope mismatch")
			}
		})
	}
}

func TestPlanApplicabilityJSONRejections(t *testing.T) {
	cases := map[string]string{
		"unknown kind":            `{"kind":"sometimes"}`,
		"missing kind":            `{}`,
		"null document":           `null`,
		"unrestricted with scope": `{"kind":"unrestricted","scope":{"kind":"peos:deployment","expression":"region=eu"}}`,
		"scoped without scope":    `{"kind":"scoped"}`,
		"scoped with null scope":  `{"kind":"scoped","scope":null}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var a PlanApplicability
			err := json.Unmarshal([]byte(payload), &a)
			if !errors.Is(err, ErrInvalidPlanApplicability) {
				t.Errorf("error = %v, want %v", err, ErrInvalidPlanApplicability)
			}
		})
	}
}

func TestPlanApplicabilityJSONNestedScopeSentinelPreserved(t *testing.T) {
	var a PlanApplicability
	err := json.Unmarshal([]byte(`{"kind":"scoped","scope":{"kind":"peos:deployment","expression":""}}`), &a)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestPlanApplicabilityFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := NewUnrestrictedPlanApplicability()
	if err := json.Unmarshal([]byte(`{"kind":"sometimes"}`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if !receiver.IsUnrestricted() {
		t.Error("failed Unmarshal disturbed the receiver")
	}
}

// --- PlanContent -------------------------------------------------------------

func TestNewPlanContentValidUnrestricted(t *testing.T) {
	c := mustPlanContent(t)
	if c.IsZero() {
		t.Fatal("valid PlanContent reports IsZero")
	}
	if !c.Applicability().IsUnrestricted() {
		t.Error("applicability not preserved")
	}
	if len(c.Activities()) != 1 {
		t.Errorf("Activities() length = %d, want 1", len(c.Activities()))
	}
	if _, ok := c.AcceptanceRules(); ok {
		t.Error("minimum PlanContent reports acceptance rules")
	}
	if !c.Extension().IsZero() {
		t.Error("minimum PlanContent has extension data")
	}
}

func TestNewPlanContentValidScoped(t *testing.T) {
	applicability, err := NewScopedPlanApplicability(mustScope(t, "deployment", "region=eu"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewPlanContent(mustScope(t, "project", "proj=alpha"), applicability, mustProvenance(t), []PlannedActivity{mustActivity(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !c.Applicability().IsScoped() {
		t.Error("scoped applicability not preserved")
	}
}

func TestNewPlanContentZeroScopeRejected(t *testing.T) {
	_, err := NewPlanContent(core.Scope{}, NewUnrestrictedPlanApplicability(), mustProvenance(t), []PlannedActivity{mustActivity(t)})
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

// TestNewPlanContentZeroApplicabilityRejected is the headline test for the
// accepted Blueprint Amendment: PEOS-006 lists applicability among a Plan
// Revision's unqualified SHALL-identify items, so an unstated applicability
// is invalid and NewUnrestrictedPlanApplicability must be used to declare
// an explicit absence of restriction.
func TestNewPlanContentZeroApplicabilityRejected(t *testing.T) {
	_, err := NewPlanContent(mustScope(t, "project", "proj=alpha"), PlanApplicability{}, mustProvenance(t), []PlannedActivity{mustActivity(t)})
	if !errors.Is(err, ErrInvalidPlanApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlanApplicability)
	}
}

func TestNewPlanContentZeroProvenanceRejected(t *testing.T) {
	_, err := NewPlanContent(mustScope(t, "project", "proj=alpha"), NewUnrestrictedPlanApplicability(), core.Provenance{}, []PlannedActivity{mustActivity(t)})
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestNewPlanContentEmptyActivitiesRejected(t *testing.T) {
	for _, activities := range [][]PlannedActivity{nil, {}} {
		_, err := NewPlanContent(mustScope(t, "project", "proj=alpha"), NewUnrestrictedPlanApplicability(), mustProvenance(t), activities)
		if !errors.Is(err, ErrInvalidValidationPlan) {
			t.Errorf("activities %#v: error = %v, want %v", activities, err, ErrInvalidValidationPlan)
		}
	}
}

func TestNewPlanContentZeroActivityElementRejected(t *testing.T) {
	_, err := NewPlanContent(mustScope(t, "project", "proj=alpha"), NewUnrestrictedPlanApplicability(), mustProvenance(t), []PlannedActivity{{}})
	if !errors.Is(err, ErrInvalidPlannedActivity) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPlannedActivity)
	}
}

func TestNewPlanContentDuplicateLocalKeyRejected(t *testing.T) {
	_, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{mustActivityKeyed(t, "ACT-1"), mustActivityKeyed(t, "ACT-1")},
	)
	if !errors.Is(err, ErrDuplicatePlanLocalKey) {
		t.Errorf("error = %v, want %v", err, ErrDuplicatePlanLocalKey)
	}
}

func TestNewPlanContentDistinctLocalKeysAccepted(t *testing.T) {
	c, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{mustActivityKeyed(t, "ACT-1"), mustActivityKeyed(t, "ACT-2")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Activities()) != 2 {
		t.Errorf("Activities() length = %d, want 2", len(c.Activities()))
	}
}

func TestNewPlanContentUnresolvedDependencyRejected(t *testing.T) {
	dependent, err := mustActivityKeyed(t, "ACT-2").WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-MISSING")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{mustActivityKeyed(t, "ACT-1"), dependent},
	)
	if !errors.Is(err, ErrUnknownPlanLocalKey) {
		t.Errorf("error = %v, want %v", err, ErrUnknownPlanLocalKey)
	}
}

// TestNewPlanContentDependencyResolutionOrderIndependent proves the full
// key set is collected before any dependency is checked, so an Activity may
// depend on one declared later in the list.
func TestNewPlanContentDependencyResolutionOrderIndependent(t *testing.T) {
	forward, err := mustActivityKeyed(t, "ACT-1").WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-2")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{forward, mustActivityKeyed(t, "ACT-2")},
	); err != nil {
		t.Fatalf("forward dependency rejected: %v", err)
	}

	backward, err := mustActivityKeyed(t, "ACT-2").WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-1")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{mustActivityKeyed(t, "ACT-1"), backward},
	); err != nil {
		t.Fatalf("backward dependency rejected: %v", err)
	}
}

// TestNewPlanContentSelfDependencyAccepted locks the normative result:
// PEOS-006 states no cycle policy and no self-reference prohibition for
// Planned Validation Activity dependencies, so a self-dependency is
// accepted rather than rejected by intuition. Dependency semantics,
// including any cycle prohibition, are repository- or Product-owned.
func TestNewPlanContentSelfDependencyAccepted(t *testing.T) {
	selfDep, err := mustActivityKeyed(t, "ACT-1").WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-1")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{selfDep},
	); err != nil {
		t.Fatalf("self-dependency rejected, but PEOS-006 states no such rule: %v", err)
	}
}

// TestNewPlanContentDependencyCycleAccepted locks the same conclusion for a
// two-node cycle.
func TestNewPlanContentDependencyCycleAccepted(t *testing.T) {
	a1, err := mustActivityKeyed(t, "ACT-1").WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-2")})
	if err != nil {
		t.Fatal(err)
	}
	a2, err := mustActivityKeyed(t, "ACT-2").WithDependencies([]core.LocalKey{mustLocalKey(t, "ACT-1")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{a1, a2},
	); err != nil {
		t.Fatalf("dependency cycle rejected, but PEOS-006 states no such rule: %v", err)
	}
}

func TestPlanContentActivitiesDefensiveCopy(t *testing.T) {
	input := []PlannedActivity{mustActivityKeyed(t, "ACT-1")}
	c, err := NewPlanContent(mustScope(t, "project", "proj=alpha"), NewUnrestrictedPlanApplicability(), mustProvenance(t), input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = PlannedActivity{}
	if c.Activities()[0].IsZero() {
		t.Error("activities input slice was not copied in")
	}
	c.Activities()[0] = PlannedActivity{}
	if c.Activities()[0].IsZero() {
		t.Error("Activities() accessor did not return a defensive copy")
	}
}

func TestPlanContentActivityLookup(t *testing.T) {
	c, err := NewPlanContent(
		mustScope(t, "project", "proj=alpha"),
		NewUnrestrictedPlanApplicability(),
		mustProvenance(t),
		[]PlannedActivity{mustActivityKeyed(t, "ACT-1"), mustActivityKeyed(t, "ACT-2")},
	)
	if err != nil {
		t.Fatal(err)
	}
	found, ok := c.Activity(mustLocalKey(t, "ACT-2"))
	if !ok {
		t.Fatal("Activity(ACT-2) not found")
	}
	if found.Key().String() != "ACT-2" {
		t.Errorf("Activity(ACT-2).Key() = %q", found.Key().String())
	}
	if _, ok := c.Activity(mustLocalKey(t, "ACT-9")); ok {
		t.Error("Activity(ACT-9) unexpectedly found")
	}
	if _, ok := c.Activity(core.LocalKey{}); ok {
		t.Error("Activity(zero key) unexpectedly found")
	}
}

func TestPlanContentAcceptanceRules(t *testing.T) {
	c, err := mustPlanContent(t).WithAcceptanceRules("  all activities must pass  ")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := c.AcceptanceRules()
	if !ok || got != "all activities must pass" {
		t.Errorf("AcceptanceRules() = %q, %v", got, ok)
	}
	if _, ok := c.WithoutAcceptanceRules().AcceptanceRules(); ok {
		t.Error("WithoutAcceptanceRules did not clear")
	}
}

func TestPlanContentAcceptanceRulesEmptyRejected(t *testing.T) {
	for _, value := range []string{"", "   ", "\t"} {
		_, err := mustPlanContent(t).WithAcceptanceRules(value)
		if !errors.Is(err, ErrInvalidValidationPlan) {
			t.Errorf("value %q: error = %v, want %v", value, err, ErrInvalidValidationPlan)
		}
	}
}

func TestPlanContentExtension(t *testing.T) {
	c := mustPlanContent(t).WithExtension(mustExtension(t, "acme", `{"owner":"qa"}`))
	if c.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !c.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestPlanContentModifierReceiverImmutability(t *testing.T) {
	original := mustPlanContent(t)
	if _, err := original.WithAcceptanceRules("rules"); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.AcceptanceRules(); ok {
		t.Error("WithAcceptanceRules mutated the receiver")
	}
	_ = original.WithExtension(mustExtension(t, "acme", `{}`))
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}
}

func TestPlanContentIsZero(t *testing.T) {
	var c PlanContent
	if !c.IsZero() {
		t.Error("zero PlanContent does not report IsZero")
	}
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidValidationPlan)
	}
	if c.Activities() != nil {
		t.Error("zero PlanContent returns non-nil Activities()")
	}
	if _, ok := c.Activity(mustLocalKey(t, "ACT-1")); ok {
		t.Error("zero PlanContent resolved an activity")
	}
}

func TestPlanContentProvenanceAccessor(t *testing.T) {
	c := mustPlanContent(t)
	got := c.Provenance()
	if got.IsZero() {
		t.Fatal("Provenance() returned the zero value")
	}
	actor, ok := got.Actor()
	if !ok || actor.Identifier() != "svc-1" {
		t.Errorf("Provenance().Actor() = %v, %v", actor, ok)
	}
}

func TestPlanContentMalformedDocumentRejected(t *testing.T) {
	for _, payload := range []string{`not json`, `[]`, `{"scope":[]}`} {
		var c PlanContent
		if err := json.Unmarshal([]byte(payload), &c); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

func TestPlanContentJSONAcceptanceRulesWrongTypeRejected(t *testing.T) {
	var c PlanContent
	err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"acceptance_rules": `123`})), &c)
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestPlanRevisionMalformedDocumentRejected(t *testing.T) {
	for _, payload := range []string{`not json`, `[]`} {
		var r PlanRevision
		if err := json.Unmarshal([]byte(payload), &r); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

func TestPlanMalformedDocumentRejected(t *testing.T) {
	var p Plan
	if err := json.Unmarshal([]byte(`[]`), &p); !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func planContentPayload(t *testing.T, overrides map[string]string) string {
	t.Helper()
	base := map[string]string{
		"scope":         `{"kind":"peos:project","expression":"proj=alpha"}`,
		"applicability": `{"kind":"unrestricted"}`,
		"provenance":    `{"actor":{"namespace":"peos-cli","identifier":"svc-1"},"recorded_at":"2026-07-27T00:00:00Z"}`,
		"activities":    `[{"key":"ACT-1","subject":{"kind":"artifact_revision","ref":{"artifact_id":"AR-42","revision_id":"REV-3"}},"method":"peos:test","outcome_interpretation":"pass iff all assertions hold"}]`,
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

func TestPlanContentJSONKeys(t *testing.T) {
	data, err := json.Marshal(mustPlanContent(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"scope", "applicability", "provenance", "activities"}
	if len(raw) != len(want) {
		t.Errorf("minimum wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, forbidden := range []string{"relation", "lifecycle", "state", "version", "executions", "claims", "acceptance_rules", "extension"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("minimum wire form unexpectedly carries %q", forbidden)
		}
	}
}

func TestPlanContentJSONFullKeys(t *testing.T) {
	c, err := mustPlanContent(t).WithAcceptanceRules("all pass")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c.WithExtension(mustExtension(t, "acme", `{"owner":"qa"}`)))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"scope", "applicability", "provenance", "activities", "acceptance_rules", "extension"}
	if len(raw) != len(want) {
		t.Errorf("full wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
}

func TestPlanContentJSONRoundTrip(t *testing.T) {
	c, err := mustPlanContent(t).WithAcceptanceRules("all pass")
	if err != nil {
		t.Fatal(err)
	}
	original := c.WithExtension(mustExtension(t, "acme", `{"owner":"qa"}`))
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded PlanContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Scope().Equal(original.Scope()) {
		t.Error("scope mismatch")
	}
	if decoded.Applicability().Kind() != original.Applicability().Kind() {
		t.Error("applicability mismatch")
	}
	if len(decoded.Activities()) != len(original.Activities()) {
		t.Error("activities length mismatch")
	}
	gotRules, gotOK := decoded.AcceptanceRules()
	wantRules, wantOK := original.AcceptanceRules()
	if gotOK != wantOK || gotRules != wantRules {
		t.Error("acceptance rules mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("extension lost")
	}
}

func TestPlanContentJSONMandatoryMissingRejected(t *testing.T) {
	cases := map[string]error{
		"scope":         core.ErrInvalidScope,
		"applicability": ErrInvalidPlanApplicability,
		"provenance":    ErrInvalidValidationPlan,
		"activities":    ErrInvalidValidationPlan,
	}
	for field, want := range cases {
		t.Run(field, func(t *testing.T) {
			var c PlanContent
			err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{field: ""})), &c)
			if !errors.Is(err, want) {
				t.Errorf("error = %v, want %v", err, want)
			}
		})
	}
}

func TestPlanContentJSONMandatoryNullRejected(t *testing.T) {
	for _, field := range []string{"scope", "applicability", "provenance", "activities"} {
		t.Run(field, func(t *testing.T) {
			var c PlanContent
			err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{field: "null"})), &c)
			if err == nil {
				t.Fatal("explicit null accepted, want error")
			}
		})
	}
}

func TestPlanContentJSONEmptyActivitiesArrayRejected(t *testing.T) {
	var c PlanContent
	err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"activities": "[]"})), &c)
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestPlanContentJSONAcceptanceRulesNullRejected(t *testing.T) {
	var c PlanContent
	err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"acceptance_rules": "null"})), &c)
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestPlanContentJSONAcceptanceRulesEmptyRejected(t *testing.T) {
	var c PlanContent
	err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"acceptance_rules": `"   "`})), &c)
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestPlanContentJSONExtensionNullMeansAbsent(t *testing.T) {
	var c PlanContent
	if err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"extension": "null"})), &c); err != nil {
		t.Fatalf("extension null rejected: %v", err)
	}
	if !c.Extension().IsZero() {
		t.Error("extension null did not decode as absent")
	}
}

func TestPlanContentJSONDuplicateKeyAndUnresolvedDependencyRejectedOnDecode(t *testing.T) {
	dup := `[{"key":"ACT-1","subject":{"kind":"artifact_revision","ref":{"artifact_id":"AR-42","revision_id":"REV-3"}},"method":"peos:test","outcome_interpretation":"x"},{"key":"ACT-1","subject":{"kind":"artifact_revision","ref":{"artifact_id":"AR-43","revision_id":"REV-1"}},"method":"peos:test","outcome_interpretation":"y"}]`
	var c PlanContent
	if err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"activities": dup})), &c); !errors.Is(err, ErrDuplicatePlanLocalKey) {
		t.Errorf("duplicate key on decode: error = %v, want %v", err, ErrDuplicatePlanLocalKey)
	}

	unresolved := `[{"key":"ACT-1","subject":{"kind":"artifact_revision","ref":{"artifact_id":"AR-42","revision_id":"REV-3"}},"method":"peos:test","outcome_interpretation":"x","dependencies":["ACT-MISSING"]}]`
	var c2 PlanContent
	if err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"activities": unresolved})), &c2); !errors.Is(err, ErrUnknownPlanLocalKey) {
		t.Errorf("unresolved dependency on decode: error = %v, want %v", err, ErrUnknownPlanLocalKey)
	}
}

func TestPlanContentJSONNestedSentinelPreserved(t *testing.T) {
	var c PlanContent
	err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"scope": `{"kind":"peos:project","expression":""}`})), &c)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestPlanContentFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := mustPlanContent(t)
	if err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"applicability": "null"})), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if !receiver.Applicability().IsUnrestricted() {
		t.Error("failed Unmarshal disturbed the receiver")
	}
	if len(receiver.Activities()) != 1 {
		t.Error("failed Unmarshal disturbed the receiver's activities")
	}
}

func TestPlanContentJSONUnknownFieldIgnored(t *testing.T) {
	var c PlanContent
	if err := json.Unmarshal([]byte(planContentPayload(t, map[string]string{"unknown": `"x"`})), &c); err != nil {
		t.Fatalf("unknown field rejected: %v", err)
	}
}

// --- PlanRevision ------------------------------------------------------------

func TestNewPlanRevisionValid(t *testing.T) {
	r := mustPlanRevision(t)
	if r.IsZero() {
		t.Fatal("valid PlanRevision reports IsZero")
	}
	if r.Core().ArtifactID().String() != "VP-1" {
		t.Error("Core() lost the artifact id")
	}
	if r.Content().IsZero() {
		t.Error("Content() lost the content")
	}
	ref, err := r.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID().String() != "VP-1" || ref.RevisionID().String() != "REV-1" {
		t.Errorf("Ref() = %v", ref)
	}
}

func TestNewPlanRevisionZeroPlanRejected(t *testing.T) {
	_, err := NewPlanRevision(Plan{}, mustCoreRevision(t, "VP-1", "REV-1"), mustPlanContent(t))
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestNewPlanRevisionZeroRevisionRejected(t *testing.T) {
	_, err := NewPlanRevision(mustPlan(t), core.ArtifactRevision{}, mustPlanContent(t))
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestNewPlanRevisionZeroContentRejected(t *testing.T) {
	_, err := NewPlanRevision(mustPlan(t), mustCoreRevision(t, "VP-1", "REV-1"), PlanContent{})
	if !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestNewPlanRevisionArtifactIDMismatchRejected(t *testing.T) {
	_, err := NewPlanRevision(mustPlan(t), mustCoreRevision(t, "VP-OTHER", "REV-1"), mustPlanContent(t))
	if !errors.Is(err, ErrValidationPlanArtifactIDMismatch) {
		t.Errorf("error = %v, want %v", err, ErrValidationPlanArtifactIDMismatch)
	}
}

func TestPlanRevisionIsZero(t *testing.T) {
	var r PlanRevision
	if !r.IsZero() {
		t.Error("zero PlanRevision does not report IsZero")
	}
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidValidationPlan) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidValidationPlan)
	}
}

func TestPlanRevisionJSONKeysAndRoundTrip(t *testing.T) {
	original := mustPlanRevision(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"core", "content"}
	if len(raw) != len(want) {
		t.Errorf("wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, forbidden := range []string{"relation", "lifecycle", "version", "plan"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("wire form unexpectedly carries %q", forbidden)
		}
	}

	var decoded PlanRevision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Core().ArtifactID() != original.Core().ArtifactID() {
		t.Error("round-trip lost artifact id")
	}
	if len(decoded.Content().Activities()) != 1 {
		t.Error("round-trip lost activities")
	}
}

func TestPlanRevisionJSONRejections(t *testing.T) {
	valid := mustPlanRevision(t)
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	for _, field := range []string{"core", "content"} {
		t.Run("missing "+field, func(t *testing.T) {
			partial := map[string]json.RawMessage{}
			for k, v := range raw {
				if k != field {
					partial[k] = v
				}
			}
			payload, err := json.Marshal(partial)
			if err != nil {
				t.Fatal(err)
			}
			var r PlanRevision
			if err := json.Unmarshal(payload, &r); !errors.Is(err, ErrInvalidValidationPlan) {
				t.Errorf("error = %v, want %v", err, ErrInvalidValidationPlan)
			}
		})
		t.Run("null "+field, func(t *testing.T) {
			nulled := map[string]json.RawMessage{}
			for k, v := range raw {
				nulled[k] = v
			}
			nulled[field] = json.RawMessage("null")
			payload, err := json.Marshal(nulled)
			if err != nil {
				t.Fatal(err)
			}
			var r PlanRevision
			if err := json.Unmarshal(payload, &r); err == nil {
				t.Error("explicit null accepted, want error")
			}
		})
	}
}

func TestPlanRevisionJSONNestedSentinelPreserved(t *testing.T) {
	valid := mustPlanRevision(t)
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["content"] = json.RawMessage(planContentPayload(t, map[string]string{"scope": `{"kind":"peos:project","expression":""}`}))
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var r PlanRevision
	if err := json.Unmarshal(payload, &r); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
	}
}

func TestPlanRevisionFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := mustPlanRevision(t)
	if err := json.Unmarshal([]byte(`{"core":null,"content":null}`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.Core().ArtifactID().String() != "VP-1" {
		t.Error("failed Unmarshal disturbed the receiver")
	}
}

// --- structural absence ------------------------------------------------------

// TestPlanContentHasNoMandatoryStateModifiers proves that none of
// PlanContent's four mandatory constructor fields is reachable through a
// With*/Without* call, and that PlanContent exposes no execution or Claim
// concept.
func TestPlanContentHasNoMandatoryStateModifiers(t *testing.T) {
	forbidden := []string{
		"WithScope", "WithoutScope",
		"WithApplicability", "WithoutApplicability",
		"WithProvenance", "WithoutProvenance",
		"WithActivities", "WithoutActivities",
		"ExecutionRecord", "ExecutionRecords", "Claim", "Claims",
		"Outcome", "Evidence", "Executions",
		"Relation", "Lifecycle", "State",
	}
	typ := reflect.TypeOf(PlanContent{})
	for _, name := range forbidden {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("PlanContent unexpectedly exposes %s", name)
		}
		if _, ok := reflect.PointerTo(typ).MethodByName(name); ok {
			t.Errorf("*PlanContent unexpectedly exposes %s", name)
		}
	}
}

// TestPlanAndPlanRevisionHaveNoForbiddenAPI proves the Plan side introduces
// no relationship, no lifecycle, no version system, and no mutable content
// path.
func TestPlanAndPlanRevisionHaveNoForbiddenAPI(t *testing.T) {
	planForbidden := []string{"Relation", "Lifecycle", "State", "Version", "Content", "WithContent", "Revisions"}
	typ := reflect.TypeOf(Plan{})
	for _, name := range planForbidden {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("Plan unexpectedly exposes %s", name)
		}
	}

	revisionForbidden := []string{"WithContent", "WithoutContent", "Relation", "Lifecycle", "State", "Version", "Claim", "ExecutionRecord"}
	rtyp := reflect.TypeOf(PlanRevision{})
	for _, name := range revisionForbidden {
		if _, ok := rtyp.MethodByName(name); ok {
			t.Errorf("PlanRevision unexpectedly exposes %s", name)
		}
		if _, ok := reflect.PointerTo(rtyp).MethodByName(name); ok {
			t.Errorf("*PlanRevision unexpectedly exposes %s", name)
		}
	}
}

// TestPackageDeclaresNoDeferredH2Types proves Packet H.1 did not smuggle in
// any Packet H.2 concept. The reserved sentinels exist by design; the types
// must not.
func TestPackageDeclaresNoDeferredH2Types(t *testing.T) {
	// Each reserved sentinel must exist and be distinct, so H.2 needs no
	// change to errors.go.
	reserved := []error{
		ErrInvalidActivityReference,
		ErrInvalidExecutionRecord,
		ErrInvalidValidationClaim,
		ErrInvalidSatisfactionClaim,
		ErrInvalidConformanceClaim,
	}
	for i, a := range reserved {
		if a == nil {
			t.Fatalf("reserved sentinel %d is nil", i)
		}
		for j, b := range reserved {
			if i != j && errors.Is(a, b) {
				t.Errorf("reserved sentinels %d and %d are not distinct", i, j)
			}
		}
	}
}
