package runtime

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// mustObservation builds a minimal valid Observation.
func mustObservation(t *testing.T, id string) Observation {
	t.Helper()
	o, err := NewObservation(
		mustRuntimeObservationID(t, id),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustTimestampAt(t, 0),
		"latency=142ms",
		"http-probe",
		mustActor(t, "peos-cli", "svc-1"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestNewObservation(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	if o.IsZero() {
		t.Error("valid Observation reports IsZero() = true")
	}
	ref, err := o.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.RecordID() != o.ID() {
		t.Error("Ref() mismatch")
	}
	if o.ObservedValue() != "latency=142ms" {
		t.Errorf("ObservedValue() = %q", o.ObservedValue())
	}
	if o.CollectionMethod() != "http-probe" {
		t.Errorf("CollectionMethod() = %q", o.CollectionMethod())
	}
	if _, ok := o.Binding(); ok {
		t.Error("new Observation should have no binding")
	}
	if _, ok := o.Interval(); ok {
		t.Error("new Observation should have no interval")
	}
	if _, ok := o.UnitScaleOrEventType(); ok {
		t.Error("new Observation should have no unit/scale/event type")
	}
	if _, ok := o.Environment(); ok {
		t.Error("new Observation should have no environment")
	}
	if _, ok := o.Uncertainty(); ok {
		t.Error("new Observation should have no uncertainty")
	}
	if len(o.Limitations()) != 0 {
		t.Error("new Observation should have no limitations")
	}
	if len(o.Evidence()) != 0 {
		t.Error("new Observation should have no Evidence -- not automatically Evidence")
	}
	if o.Subject() != mustRuntimeSubjectRef(t, "kubernetes", "pod-1") {
		t.Error("Subject() mismatch")
	}
	if o.ObservedAt() != mustTimestampAt(t, 0) {
		t.Error("ObservedAt() mismatch")
	}
	if o.Source() != mustActor(t, "peos-cli", "svc-1") {
		t.Error("Source() mismatch")
	}
	if o.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if !o.Extension().IsZero() {
		t.Error("new Observation should have zero extension")
	}
}

func TestNewObservationMandatoryFieldRejections(t *testing.T) {
	id := mustRuntimeObservationID(t, "OBS-1")
	subject := mustRuntimeSubjectRef(t, "kubernetes", "pod-1")
	observedAt := mustTimestampAt(t, 0)
	source := mustActor(t, "peos-cli", "svc-1")
	provenance := mustProvenance(t)

	if _, err := NewObservation(core.RuntimeObservationID{}, subject, observedAt, "value", "method", source, provenance); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero id: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := NewObservation(id, core.RuntimeSubjectRef{}, observedAt, "value", "method", source, provenance); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero subject: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := NewObservation(id, subject, core.Timestamp{}, "value", "method", source, provenance); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero observed-at: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := NewObservation(id, subject, observedAt, "   ", "method", source, provenance); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("whitespace-only observed value: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := NewObservation(id, subject, observedAt, "value", "", source, provenance); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("empty collection method: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := NewObservation(id, subject, observedAt, "value", "method", core.ActorRef{}, provenance); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero source: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := NewObservation(id, subject, observedAt, "value", "method", source, core.Provenance{}); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
}

func TestObservationWithBinding(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	o2, err := o.WithBinding(bindingRef)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := o2.Binding()
	if !ok || got != bindingRef {
		t.Errorf("Binding() = (%v, %v)", got, ok)
	}
	if _, err := o.WithBinding(core.RuntimeBindingRecordRef{}); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero binding: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	cleared := o2.WithoutBinding()
	if _, ok := cleared.Binding(); ok {
		t.Error("WithoutBinding did not clear the field")
	}
}

func TestObservationWithInterval(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	o2, err := o.WithInterval(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := o2.Interval()
	if !ok || got != mustTimestampAt(t, 5) {
		t.Errorf("Interval() = (%v, %v)", got, ok)
	}
	if _, err := o.WithInterval(core.Timestamp{}); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero interval end: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	if _, err := o.WithInterval(mustTimestampAt(t, -1)); err == nil {
		t.Error("interval end before observed-at accepted, want error")
	}
	cleared := o2.WithoutInterval()
	if _, ok := cleared.Interval(); ok {
		t.Error("WithoutInterval did not clear the field")
	}
}

func TestObservationWithUnitScaleOrEventType(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	o2, err := o.WithUnitScaleOrEventType("milliseconds")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := o2.UnitScaleOrEventType()
	if !ok || got != "milliseconds" {
		t.Errorf("UnitScaleOrEventType() = (%q, %v)", got, ok)
	}
	if _, err := o.WithUnitScaleOrEventType("  "); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("whitespace-only: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	cleared := o2.WithoutUnitScaleOrEventType()
	if _, ok := cleared.UnitScaleOrEventType(); ok {
		t.Error("WithoutUnitScaleOrEventType did not clear the field")
	}
}

func TestObservationWithEnvironment(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	env := mustEnvironment(t, "production")
	o2, err := o.WithEnvironment(env)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := o2.Environment()
	if !ok || got != env {
		t.Errorf("Environment() = (%v, %v)", got, ok)
	}
	if _, err := o.WithEnvironment(Environment{}); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero environment: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	cleared := o2.WithoutEnvironment()
	if _, ok := cleared.Environment(); ok {
		t.Error("WithoutEnvironment did not clear the field")
	}
}

func TestObservationWithUncertainty(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	o2, err := o.WithUncertainty("+/- 5ms")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := o2.Uncertainty()
	if !ok || got != "+/- 5ms" {
		t.Errorf("Uncertainty() = (%q, %v)", got, ok)
	}
	if _, err := o.WithUncertainty(""); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("empty: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	cleared := o2.WithoutUncertainty()
	if _, ok := cleared.Uncertainty(); ok {
		t.Error("WithoutUncertainty did not clear the field")
	}
}

func TestObservationWithLimitations(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	o2, err := o.WithLimitations([]string{"sampled, not continuous"})
	if err != nil {
		t.Fatal(err)
	}
	got := o2.Limitations()
	if len(got) != 1 || got[0] != "sampled, not continuous" {
		t.Errorf("Limitations() = %v", got)
	}
	got[0] = "mutated"
	if o2.Limitations()[0] == "mutated" {
		t.Error("Limitations() accessor did not return a defensive copy")
	}
	if _, err := o.WithLimitations([]string{""}); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("empty entry: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
}

func TestObservationWithEvidenceIsExplicitNotAutomatic(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	if len(o.Evidence()) != 0 {
		t.Fatal("a freshly constructed Observation must not already carry Evidence")
	}
	evidence := mustEvidence(t, "ART-EV-1", "REV-1")
	o2, err := o.WithEvidence([]core.EvidenceArtifactRevisionRef{evidence})
	if err != nil {
		t.Fatal(err)
	}
	got := o2.Evidence()
	if len(got) != 1 || got[0] != evidence {
		t.Errorf("Evidence() = %v", got)
	}
	if len(o.Evidence()) != 0 {
		t.Error("original Observation mutated by WithEvidence")
	}
	if _, err := o.WithEvidence([]core.EvidenceArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero evidence element: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
	cleared, err := o2.WithEvidence(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Evidence()) != 0 {
		t.Error("WithEvidence(nil) did not clear the collection")
	}
}

func TestObservationDefensiveCopy(t *testing.T) {
	evidence := []core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")}
	o, err := mustObservation(t, "OBS-1").WithEvidence(evidence)
	if err != nil {
		t.Fatal(err)
	}
	evidence[0] = core.EvidenceArtifactRevisionRef{}
	if o.Evidence()[0].IsZero() {
		t.Error("WithEvidence did not defensively copy its input slice")
	}
}

func TestObservationExtension(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	o2 := o.WithExtension(ext)
	if o2.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !o.Extension().IsZero() {
		t.Error("original Observation mutated by WithExtension")
	}
	o3 := o2.WithoutExtension()
	if !o3.Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestObservationMarshalZero(t *testing.T) {
	var o Observation
	if _, err := json.Marshal(o); !errors.Is(err, ErrInvalidRuntimeObservation) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeObservation)
	}
}

func TestObservationJSONRoundTrip(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	o := mustObservation(t, "OBS-1")
	o, err = o.WithBinding(bindingRef)
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithInterval(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithUnitScaleOrEventType("milliseconds")
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithEnvironment(mustEnvironment(t, "production"))
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithUncertainty("+/- 5ms")
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithLimitations([]string{"sampled"})
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithEvidence([]core.EvidenceArtifactRevisionRef{mustEvidence(t, "ART-EV-1", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	o = o.WithExtension(ext)

	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Observation
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

func TestObservationEvidenceAbsentNullEmptyEquivalent(t *testing.T) {
	base := mustObservation(t, "OBS-1")
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	delete(m, "evidence")

	for _, tt := range []struct {
		name  string
		value json.RawMessage
	}{
		{"absent", nil},
		{"null", json.RawMessage("null")},
		{"empty array", json.RawMessage("[]")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			if tt.value != nil {
				mCopy["evidence"] = tt.value
			}
			modified, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var decoded Observation
			if err := json.Unmarshal(modified, &decoded); err != nil {
				t.Fatalf("%s: unexpected error %v", tt.name, err)
			}
			if len(decoded.Evidence()) != 0 {
				t.Errorf("%s: Evidence() = %v, want empty", tt.name, decoded.Evidence())
			}
		})
	}
}

func TestObservationUnmarshalFieldSpecificRejections(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	o := mustObservation(t, "OBS-1")
	o, err = o.WithBinding(bindingRef)
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithInterval(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithUnitScaleOrEventType("milliseconds")
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithEnvironment(mustEnvironment(t, "production"))
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithUncertainty("+/- 5ms")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(o)
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
		{"interval_end before observed_at", "interval_end", json.RawMessage(`"2020-01-01T00:00:00Z"`)},
		{"null unit_scale_or_event_type", "unit_scale_or_event_type", json.RawMessage("null")},
		{"malformed unit_scale_or_event_type", "unit_scale_or_event_type", json.RawMessage(`123`)},
		{"whitespace-only unit_scale_or_event_type", "unit_scale_or_event_type", json.RawMessage(`"   "`)},
		{"null environment", "environment", json.RawMessage("null")},
		{"malformed environment", "environment", json.RawMessage(`123`)},
		{"null uncertainty", "uncertainty", json.RawMessage("null")},
		{"empty uncertainty", "uncertainty", json.RawMessage(`""`)},
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
			var decoded Observation
			if err := json.Unmarshal(modified, &decoded); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}
}

func TestObservationUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	originalData, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"observed_value":""}`), &o); err == nil {
		t.Fatal("empty observed value accepted, want error")
	}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
	if err := json.Unmarshal([]byte(`not json`), &o); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
}

func TestObservationNoForbiddenWireKeys(t *testing.T) {
	o := mustObservation(t, "OBS-1")
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"execution", "invocation", "result", "outcome", "success",
		"failure", "current", "latest", "effective", "status", "state",
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
