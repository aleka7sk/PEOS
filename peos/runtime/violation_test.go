package runtime

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustViolationClassification(t *testing.T, value string) ViolationClassification {
	t.Helper()
	return NewViolationClassification(mustVocabularyValue(t, "product", value))
}

func mustViolationSeverity(t *testing.T, value string) ViolationSeverity {
	t.Helper()
	return NewViolationSeverity(mustVocabularyValue(t, "product", value))
}

func mustRuntimeObservationRef(t *testing.T, id string) core.RuntimeObservationRef {
	t.Helper()
	ref, err := core.NewRuntimeObservationRef(mustRuntimeObservationID(t, id))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// mustObservationTriggeredViolation builds a minimal valid Violation
// triggered by a Runtime Observation.
func mustObservationTriggeredViolation(t *testing.T, id string) Violation {
	t.Helper()
	v, err := NewViolationFromObservation(
		mustRuntimeViolationID(t, id),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustCriterionRef(t, "REQ-1"),
		mustRuntimeObservationRef(t, "OBS-1"),
		mustTimestampAt(t, 0),
		mustViolationClassification(t, "latency-breach"),
		mustScope(t, "cluster=prod-1"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// --- ViolationTrigger ---------------------------------------------------------

func TestNewObservationTrigger(t *testing.T) {
	ref := mustRuntimeObservationRef(t, "OBS-1")
	trig, err := NewObservationTrigger(ref)
	if err != nil {
		t.Fatal(err)
	}
	if trig.Kind() != "observation" {
		t.Errorf("Kind() = %q, want %q", trig.Kind(), "observation")
	}
	got, ok := trig.Observation()
	if !ok || got != ref {
		t.Errorf("Observation() = (%v, %v)", got, ok)
	}
	if _, ok := trig.Evidence(); ok {
		t.Error("Evidence() ok = true for observation arm")
	}
	if _, err := NewObservationTrigger(core.RuntimeObservationRef{}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero ref: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestNewEvidenceTrigger(t *testing.T) {
	ref := mustEvidence(t, "ART-EV-1", "REV-1")
	trig, err := NewEvidenceTrigger(ref)
	if err != nil {
		t.Fatal(err)
	}
	if trig.Kind() != "evidence" {
		t.Errorf("Kind() = %q, want %q", trig.Kind(), "evidence")
	}
	got, ok := trig.Evidence()
	if !ok || got != ref {
		t.Errorf("Evidence() = (%v, %v)", got, ok)
	}
	if _, ok := trig.Observation(); ok {
		t.Error("Observation() ok = true for evidence arm")
	}
	if _, err := NewEvidenceTrigger(core.EvidenceArtifactRevisionRef{}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero ref: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationTriggerZeroInvalid(t *testing.T) {
	var trig ViolationTrigger
	if !trig.IsZero() {
		t.Error("zero-value ViolationTrigger.IsZero() = false, want true")
	}
	if _, err := json.Marshal(trig); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationTriggerJSONRoundTrip(t *testing.T) {
	obsTrig, err := NewObservationTrigger(mustRuntimeObservationRef(t, "OBS-1"))
	if err != nil {
		t.Fatal(err)
	}
	evTrig, err := NewEvidenceTrigger(mustEvidence(t, "ART-EV-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	for name, trig := range map[string]ViolationTrigger{"observation": obsTrig, "evidence": evTrig} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(trig)
			if err != nil {
				t.Fatal(err)
			}
			var decoded ViolationTrigger
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != trig {
				t.Errorf("round trip mismatch: got %+v, want %+v", decoded, trig)
			}
		})
	}
}

func TestViolationTriggerUnmarshalRejections(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing kind", `{}`},
		{"unknown kind", `{"kind":"incident"}`},
		{"null", `null`},
		{"observation missing observation", `{"kind":"observation"}`},
		{"observation with null observation", `{"kind":"observation","observation":null}`},
		{"observation carrying evidence", `{"kind":"observation","observation":{"record_id":"OBS-1"},"evidence":{"artifact_id":"ART-EV-1","revision_id":"REV-1"}}`},
		{"evidence missing evidence", `{"kind":"evidence"}`},
		{"evidence with null evidence", `{"kind":"evidence","evidence":null}`},
		{"evidence carrying observation", `{"kind":"evidence","evidence":{"artifact_id":"ART-EV-1","revision_id":"REV-1"},"observation":{"record_id":"OBS-1"}}`},
		{"malformed observation payload type", `{"kind":"observation","observation":123}`},
		{"malformed evidence payload type", `{"kind":"evidence","evidence":123}`},
		{"malformed JSON", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trig ViolationTrigger
			if err := json.Unmarshal([]byte(tt.json), &trig); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

// --- Violation -----------------------------------------------------------------

func TestNewViolationFromObservation(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	if v.IsZero() {
		t.Error("valid Violation reports IsZero() = true")
	}
	if v.ID() != mustRuntimeViolationID(t, "VIOL-1") {
		t.Error("ID() mismatch")
	}
	trigger := v.Trigger()
	got, ok := trigger.Observation()
	if !ok || got != mustRuntimeObservationRef(t, "OBS-1") {
		t.Errorf("Trigger().Observation() = (%v, %v)", got, ok)
	}
	if v.Subject() != mustRuntimeSubjectRef(t, "kubernetes", "pod-1") {
		t.Error("Subject() mismatch")
	}
	if v.Criterion() != mustCriterionRef(t, "REQ-1") {
		t.Error("Criterion() mismatch")
	}
	if v.OccurredAt() != mustTimestampAt(t, 0) {
		t.Error("OccurredAt() mismatch")
	}
	if v.Classification() != mustViolationClassification(t, "latency-breach") {
		t.Error("Classification() mismatch")
	}
	if v.Scope() != mustScope(t, "cluster=prod-1") {
		t.Error("Scope() mismatch")
	}
	if v.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if !v.Extension().IsZero() {
		t.Error("new Violation should have zero extension")
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	v2 := v.WithExtension(ext)
	if v2.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !v.Extension().IsZero() {
		t.Error("original Violation mutated by WithExtension")
	}
	v3 := v2.WithoutExtension()
	if !v3.Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestNewViolationFromEvidence(t *testing.T) {
	evidence := mustEvidence(t, "ART-EV-1", "REV-1")
	v, err := NewViolationFromEvidence(
		mustRuntimeViolationID(t, "VIOL-1"),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustCriterionRef(t, "REQ-1"),
		evidence,
		mustTimestampAt(t, 0),
		mustViolationClassification(t, "latency-breach"),
		mustScope(t, "cluster=prod-1"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := v.Trigger().Evidence()
	if !ok || got != evidence {
		t.Errorf("Trigger().Evidence() = (%v, %v)", got, ok)
	}
}

func TestNewViolationZeroTriggerRejected(t *testing.T) {
	if _, err := NewViolationFromObservation(
		mustRuntimeViolationID(t, "VIOL-1"),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustCriterionRef(t, "REQ-1"),
		core.RuntimeObservationRef{},
		mustTimestampAt(t, 0),
		mustViolationClassification(t, "latency-breach"),
		mustScope(t, "cluster=prod-1"),
		mustProvenance(t),
	); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero observation trigger: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := NewViolationFromEvidence(
		mustRuntimeViolationID(t, "VIOL-1"),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustCriterionRef(t, "REQ-1"),
		core.EvidenceArtifactRevisionRef{},
		mustTimestampAt(t, 0),
		mustViolationClassification(t, "latency-breach"),
		mustScope(t, "cluster=prod-1"),
		mustProvenance(t),
	); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero evidence trigger: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestNewViolationMandatoryFieldRejections(t *testing.T) {
	id := mustRuntimeViolationID(t, "VIOL-1")
	subject := mustRuntimeSubjectRef(t, "kubernetes", "pod-1")
	criterion := mustCriterionRef(t, "REQ-1")
	trigger := mustRuntimeObservationRef(t, "OBS-1")
	occurredAt := mustTimestampAt(t, 0)
	classification := mustViolationClassification(t, "latency-breach")
	scope := mustScope(t, "cluster=prod-1")
	provenance := mustProvenance(t)

	if _, err := NewViolationFromObservation(core.RuntimeViolationID{}, subject, criterion, trigger, occurredAt, classification, scope, provenance); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero id: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := NewViolationFromObservation(id, core.RuntimeSubjectRef{}, criterion, trigger, occurredAt, classification, scope, provenance); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero subject: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := NewViolationFromObservation(id, subject, core.CriterionRef{}, trigger, occurredAt, classification, scope, provenance); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero criterion: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := NewViolationFromObservation(id, subject, criterion, trigger, core.Timestamp{}, classification, scope, provenance); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero occurred-at: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := NewViolationFromObservation(id, subject, criterion, trigger, occurredAt, ViolationClassification{}, scope, provenance); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero classification: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := NewViolationFromObservation(id, subject, criterion, trigger, occurredAt, classification, core.Scope{}, provenance); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
	if _, err := NewViolationFromObservation(id, subject, criterion, trigger, occurredAt, classification, scope, core.Provenance{}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationWithBinding(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	v2, err := v.WithBinding(bindingRef)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := v2.Binding()
	if !ok || got != bindingRef {
		t.Errorf("Binding() = (%v, %v)", got, ok)
	}
	if _, err := v.WithBinding(core.RuntimeBindingRecordRef{}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero binding: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	cleared := v2.WithoutBinding()
	if _, ok := cleared.Binding(); ok {
		t.Error("WithoutBinding did not clear the field")
	}
}

func TestViolationWithInterval(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	v2, err := v.WithInterval(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := v2.Interval()
	if !ok || got != mustTimestampAt(t, 5) {
		t.Errorf("Interval() = (%v, %v)", got, ok)
	}
	if _, err := v.WithInterval(core.Timestamp{}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero interval end: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	if _, err := v.WithInterval(mustTimestampAt(t, -1)); err == nil {
		t.Error("interval end before occurred-at accepted, want error")
	}
	cleared := v2.WithoutInterval()
	if _, ok := cleared.Interval(); ok {
		t.Error("WithoutInterval did not clear the field")
	}
}

func TestViolationWithSeverity(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	severity := mustViolationSeverity(t, "critical")
	v2, err := v.WithSeverity(severity)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := v2.Severity()
	if !ok || got != severity {
		t.Errorf("Severity() = (%v, %v)", got, ok)
	}
	if _, err := v.WithSeverity(ViolationSeverity{}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero severity: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	cleared := v2.WithoutSeverity()
	if _, ok := cleared.Severity(); ok {
		t.Error("WithoutSeverity did not clear the field")
	}
}

func TestViolationWithUncertainty(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	v2, err := v.WithUncertainty("sampling noise")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := v2.Uncertainty()
	if !ok || got != "sampling noise" {
		t.Errorf("Uncertainty() = (%q, %v)", got, ok)
	}
	if _, err := v.WithUncertainty(""); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("empty: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
	cleared := v2.WithoutUncertainty()
	if _, ok := cleared.Uncertainty(); ok {
		t.Error("WithoutUncertainty did not clear the field")
	}
}

func TestViolationWithLimitations(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	v2, err := v.WithLimitations([]string{"single sample"})
	if err != nil {
		t.Fatal(err)
	}
	got := v2.Limitations()
	if len(got) != 1 || got[0] != "single sample" {
		t.Errorf("Limitations() = %v", got)
	}
	if _, err := v.WithLimitations([]string{""}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("empty entry: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

// TestViolationHasNoApplicableWaiverAPI documents audit finding J3-04's
// resolution: Violation carries no applicable-Waiver field, accessor, or
// modifier of any kind. Waiver applicability is entirely repository-owned,
// computed as an input to derived Runtime Compliance -- see doc.go's RJ-1
// note.
func TestViolationHasNoApplicableWaiverAPI(t *testing.T) {
	forbidden := []string{
		"ApplicableWaiver", "WithApplicableWaiver", "WithoutApplicableWaiver",
	}
	assertNoMethods(t, "Violation", reflect.TypeOf(Violation{}), forbidden)
}

func TestViolationWithRelatedClaims(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	claimRef, err := core.NewValidationClaimRef(func() core.ValidationClaimID {
		id, err := core.NewValidationClaimID("CLAIM-1")
		if err != nil {
			t.Fatal(err)
		}
		return id
	}())
	if err != nil {
		t.Fatal(err)
	}
	v2, err := v.WithRelatedClaims([]core.ValidationClaimRef{claimRef})
	if err != nil {
		t.Fatal(err)
	}
	got := v2.RelatedClaims()
	if len(got) != 1 || got[0] != claimRef {
		t.Errorf("RelatedClaims() = %v", got)
	}
	if _, err := v.WithRelatedClaims([]core.ValidationClaimRef{{}}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero claim ref: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationWithRelatedDecisions(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	decisionID, err := core.NewDecisionID("DEC-1")
	if err != nil {
		t.Fatal(err)
	}
	decisionRef, err := core.NewDecisionRef(decisionID)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := v.WithRelatedDecisions([]core.DecisionRef{decisionRef})
	if err != nil {
		t.Fatal(err)
	}
	got := v2.RelatedDecisions()
	if len(got) != 1 || got[0] != decisionRef {
		t.Errorf("RelatedDecisions() = %v", got)
	}
	if _, err := v.WithRelatedDecisions([]core.DecisionRef{{}}); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero decision ref: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationMarshalZero(t *testing.T) {
	var v Violation
	if _, err := json.Marshal(v); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationJSONRoundTrip(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithBinding(bindingRef)
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithInterval(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithSeverity(mustViolationSeverity(t, "critical"))
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithUncertainty("sampling noise")
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithLimitations([]string{"single sample"})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	v = v.WithExtension(ext)

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Violation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
}

func TestViolationUnmarshalMissingTriggerRejected(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	delete(m, "trigger")
	modified, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Violation
	if err := json.Unmarshal(modified, &decoded); !errors.Is(err, ErrInvalidRuntimeViolation) {
		t.Errorf("missing trigger: error = %v, want %v", err, ErrInvalidRuntimeViolation)
	}
}

func TestViolationUnmarshalFieldSpecificRejections(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	v, err = v.WithBinding(bindingRef)
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithInterval(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithSeverity(mustViolationSeverity(t, "critical"))
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithUncertainty("sampling noise")
	if err != nil {
		t.Fatal(err)
	}
	claimRef, err := core.NewValidationClaimRef(mustValidationClaimID(t, "CLAIM-1"))
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.WithRelatedClaims([]core.ValidationClaimRef{claimRef})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var base map[string]json.RawMessage
	if err := json.Unmarshal(data, &base); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		key   string
		value json.RawMessage
	}{
		{"null binding", "binding", json.RawMessage("null")},
		{"malformed binding", "binding", json.RawMessage(`123`)},
		{"null interval_end", "interval_end", json.RawMessage("null")},
		{"malformed interval_end", "interval_end", json.RawMessage(`123`)},
		{"interval_end before occurred_at", "interval_end", json.RawMessage(`"2020-01-01T00:00:00Z"`)},
		{"null severity", "severity", json.RawMessage("null")},
		{"malformed severity", "severity", json.RawMessage(`123`)},
		{"null uncertainty", "uncertainty", json.RawMessage("null")},
		{"empty uncertainty", "uncertainty", json.RawMessage(`""`)},
		{"malformed related_claims", "related_claims", json.RawMessage(`[123]`)},
		{"zero related_claims element", "related_claims", json.RawMessage(`[{}]`)},
		{"malformed related_decisions", "related_decisions", json.RawMessage(`[123]`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(map[string]json.RawMessage, len(base))
			for k, v := range base {
				m[k] = v
			}
			m[tt.key] = tt.value
			modified, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			var decoded Violation
			if err := json.Unmarshal(modified, &decoded); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}
}

func TestViolationUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	originalData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"id":"VIOL-2"}`), &v); err == nil {
		t.Fatal("incomplete violation accepted, want error")
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
	if err := json.Unmarshal([]byte(`not json`), &v); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
}

func TestViolationNoForbiddenWireKeys(t *testing.T) {
	v := mustObservationTriggeredViolation(t, "VIOL-1")
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"status", "state", "lifecycle", "resolved", "closed",
		"outcome_authority", "incident", "compliant", "compliance",
		"applicable_waiver",
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range forbidden {
		if _, ok := m[key]; ok {
			t.Errorf("wire form contains forbidden key %q", key)
		}
	}
}
