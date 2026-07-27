package runtime

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- helpers -----------------------------------------------------------------

func mustRuntimeBindingRecordID(t *testing.T, value string) core.RuntimeBindingRecordID {
	t.Helper()
	id, err := core.NewRuntimeBindingRecordID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustRuntimeUnbindingRecordID(t *testing.T, value string) core.RuntimeUnbindingRecordID {
	t.Helper()
	id, err := core.NewRuntimeUnbindingRecordID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustRuntimeObservationID(t *testing.T, value string) core.RuntimeObservationID {
	t.Helper()
	id, err := core.NewRuntimeObservationID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustRuntimeViolationID(t *testing.T, value string) core.RuntimeViolationID {
	t.Helper()
	id, err := core.NewRuntimeViolationID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustRuntimeContractRevisionRef(t *testing.T, artifactID, revisionID string) core.RuntimeContractRevisionRef {
	t.Helper()
	ref, err := core.NewRuntimeContractRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustTimestampAt(t *testing.T, second int) core.Timestamp {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 7, 27, 0, 0, second, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func mustActor(t *testing.T, namespace, identifier string) core.ActorRef {
	t.Helper()
	ref, err := core.NewActorRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustEvidence(t *testing.T, artifactID, revisionID string) core.EvidenceArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewEvidenceArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// mustBindingRecord builds a minimal valid BindingRecord.
func mustBindingRecord(t *testing.T, id string) BindingRecord {
	t.Helper()
	b, err := NewBindingRecord(
		mustRuntimeBindingRecordID(t, id),
		mustRuntimeContractRevisionRef(t, "ART-1", "REV-1"),
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustEnvironment(t, "production"),
		mustScope(t, "cluster=prod-1"),
		mustTimestampAt(t, 0),
		mustActor(t, "peos-cli", "svc-1"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// mustUnbindingRecord builds a minimal valid UnbindingRecord referencing
// binding.
func mustUnbindingRecord(t *testing.T, id string, binding core.RuntimeBindingRecordRef) UnbindingRecord {
	t.Helper()
	u, err := NewUnbindingRecord(
		mustRuntimeUnbindingRecordID(t, id),
		binding,
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustTimestampAt(t, 10),
		"scaled down",
		mustActor(t, "peos-cli", "svc-1"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// --- BindingRecord -------------------------------------------------------------

func TestNewBindingRecord(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	if b.IsZero() {
		t.Error("valid BindingRecord reports IsZero() = true")
	}
	if b.ID() != mustRuntimeBindingRecordID(t, "BIND-1") {
		t.Error("ID() mismatch")
	}
	ref, err := b.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.RecordID() != b.ID() {
		t.Error("Ref() mismatch")
	}
	if b.ContractRevision() != mustRuntimeContractRevisionRef(t, "ART-1", "REV-1") {
		t.Error("ContractRevision() mismatch")
	}
	if _, ok := b.DeploymentTimestamp(); ok {
		t.Error("new BindingRecord should have no deployment timestamp")
	}
	if _, ok := b.Authority(); ok {
		t.Error("new BindingRecord should have no authority")
	}
	if _, ok := b.ConfigurationReference(); ok {
		t.Error("new BindingRecord should have no configuration reference")
	}
	if len(b.Limitations()) != 0 {
		t.Error("new BindingRecord should have no limitations")
	}
	if _, ok := b.Correction(); ok {
		t.Error("new BindingRecord should have no correction")
	}
}

func TestBindingRecordAccessors(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	if b.Subject() != mustRuntimeSubjectRef(t, "kubernetes", "pod-1") {
		t.Error("Subject() mismatch")
	}
	if b.Environment() != mustEnvironment(t, "production") {
		t.Error("Environment() mismatch")
	}
	if b.Scope() != mustScope(t, "cluster=prod-1") {
		t.Error("Scope() mismatch")
	}
	if b.BoundAt() != mustTimestampAt(t, 0) {
		t.Error("BoundAt() mismatch")
	}
	if b.Actor() != mustActor(t, "peos-cli", "svc-1") {
		t.Error("Actor() mismatch")
	}
	if b.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if !b.Extension().IsZero() {
		t.Error("new BindingRecord should have zero extension")
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	b2 := b.WithExtension(ext)
	if b2.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !b.Extension().IsZero() {
		t.Error("original BindingRecord mutated by WithExtension")
	}
	b3 := b2.WithoutExtension()
	if !b3.Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestNewBindingRecordMandatoryFieldRejections(t *testing.T) {
	id := mustRuntimeBindingRecordID(t, "BIND-1")
	contractRevision := mustRuntimeContractRevisionRef(t, "ART-1", "REV-1")
	subject := mustRuntimeSubjectRef(t, "kubernetes", "pod-1")
	environment := mustEnvironment(t, "production")
	scope := mustScope(t, "cluster=prod-1")
	boundAt := mustTimestampAt(t, 0)
	actor := mustActor(t, "peos-cli", "svc-1")
	provenance := mustProvenance(t)

	if _, err := NewBindingRecord(core.RuntimeBindingRecordID{}, contractRevision, subject, environment, scope, boundAt, actor, provenance); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero id: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	if _, err := NewBindingRecord(id, core.RuntimeContractRevisionRef{}, subject, environment, scope, boundAt, actor, provenance); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero contract revision: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	if _, err := NewBindingRecord(id, contractRevision, core.RuntimeSubjectRef{}, environment, scope, boundAt, actor, provenance); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero subject: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	if _, err := NewBindingRecord(id, contractRevision, subject, Environment{}, scope, boundAt, actor, provenance); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero environment: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	if _, err := NewBindingRecord(id, contractRevision, subject, environment, core.Scope{}, boundAt, actor, provenance); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
	if _, err := NewBindingRecord(id, contractRevision, subject, environment, scope, core.Timestamp{}, actor, provenance); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero bound-at: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	if _, err := NewBindingRecord(id, contractRevision, subject, environment, scope, boundAt, core.ActorRef{}, provenance); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero actor: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	if _, err := NewBindingRecord(id, contractRevision, subject, environment, scope, boundAt, actor, core.Provenance{}); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
}

func TestBindingRecordExactContractReferenceType(t *testing.T) {
	// Structurally, NewBindingRecord's second parameter is
	// core.RuntimeContractRevisionRef -- there is no overload or coercion
	// path accepting core.RuntimeContractRef or an ArtifactID alone. This
	// test documents that guarantee by using the exact type.
	b := mustBindingRecord(t, "BIND-1")
	if b.ContractRevision().RevisionID() != mustArtifactRevisionID(t, "REV-1") {
		t.Error("ContractRevision() lost exact revision level")
	}
}

func TestBindingRecordWithDeploymentTimestamp(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	b2, err := b.WithDeploymentTimestamp(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := b2.DeploymentTimestamp()
	if !ok || got != mustTimestampAt(t, 5) {
		t.Errorf("DeploymentTimestamp() = (%v, %v)", got, ok)
	}
	if _, ok := b.DeploymentTimestamp(); ok {
		t.Error("original BindingRecord mutated by WithDeploymentTimestamp")
	}
	if _, err := b.WithDeploymentTimestamp(core.Timestamp{}); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero timestamp: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	cleared := b2.WithoutDeploymentTimestamp()
	if _, ok := cleared.DeploymentTimestamp(); ok {
		t.Error("WithoutDeploymentTimestamp did not clear the field")
	}
}

func TestBindingRecordWithAuthority(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	authority := mustAuthority(t)
	b2, err := b.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := b2.Authority()
	if !ok || got != authority {
		t.Errorf("Authority() = (%v, %v)", got, ok)
	}
	if _, err := b.WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero authority: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	cleared := b2.WithoutAuthority()
	if _, ok := cleared.Authority(); ok {
		t.Error("WithoutAuthority did not clear the field")
	}
}

func TestBindingRecordWithConfigurationReference(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	b2, err := b.WithConfigurationReference("deployment/checkout-v3")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := b2.ConfigurationReference()
	if !ok || got != "deployment/checkout-v3" {
		t.Errorf("ConfigurationReference() = (%q, %v)", got, ok)
	}
	if _, err := b.WithConfigurationReference("   "); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("whitespace-only: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	cleared := b2.WithoutConfigurationReference()
	if _, ok := cleared.ConfigurationReference(); ok {
		t.Error("WithoutConfigurationReference did not clear the field")
	}
}

func TestBindingRecordWithLimitations(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	b2, err := b.WithLimitations([]string{"reduced observability window"})
	if err != nil {
		t.Fatal(err)
	}
	got := b2.Limitations()
	if len(got) != 1 || got[0] != "reduced observability window" {
		t.Errorf("Limitations() = %v", got)
	}
	got[0] = "mutated"
	if b2.Limitations()[0] == "mutated" {
		t.Error("Limitations() accessor did not return a defensive copy")
	}
	if _, err := b.WithLimitations([]string{""}); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("empty entry: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
	cleared, err := b2.WithLimitations(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Limitations()) != 0 {
		t.Error("WithLimitations(nil) did not clear the collection")
	}
}

func TestBindingRecordCorrection(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	earlier := mustRuntimeBindingRecordID(t, "BIND-0")
	earlierRef, err := core.NewRuntimeBindingRecordRef(earlier)
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, earlierRef)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := b.WithCorrection(correction)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := b2.Correction()
	if !ok || got.Target().RecordID() != earlier {
		t.Errorf("Correction() = (%v, %v)", got, ok)
	}
	if _, err := b.WithCorrection(core.RecordCorrectionRef[core.RuntimeBindingRecordRef]{}); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("zero correction: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}
	cleared := b2.WithoutCorrection()
	if _, ok := cleared.Correction(); ok {
		t.Error("WithoutCorrection did not clear the field")
	}
}

func TestBindingRecordSelfCorrectionRejected(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	selfRef, err := core.NewRuntimeBindingRecordRef(mustRuntimeBindingRecordID(t, "BIND-1"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, selfRef)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.WithCorrection(correction); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("self-correction: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}
}

func TestBindingRecordMarshalZero(t *testing.T) {
	var b BindingRecord
	if _, err := json.Marshal(b); !errors.Is(err, ErrInvalidRuntimeBindingRecord) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeBindingRecord)
	}
}

func TestBindingRecordJSONRoundTrip(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	b, err := b.WithDeploymentTimestamp(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithConfigurationReference("deployment/checkout-v3")
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithLimitations([]string{"reduced observability"})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	b = b.WithExtension(ext)

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var decoded BindingRecord
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

func TestBindingRecordUnmarshalFieldSpecificRejections(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	b, err := b.WithDeploymentTimestamp(mustTimestampAt(t, 5))
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithConfigurationReference("deployment/checkout-v3")
	if err != nil {
		t.Fatal(err)
	}
	earlierRef, err := core.NewRuntimeBindingRecordRef(mustRuntimeBindingRecordID(t, "BIND-0"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, earlierRef)
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithCorrection(correction)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(b)
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
		{"null deployment_at", "deployment_at", json.RawMessage("null")},
		{"malformed deployment_at", "deployment_at", json.RawMessage(`123`)},
		{"null authority", "authority", json.RawMessage("null")},
		{"malformed authority", "authority", json.RawMessage(`123`)},
		{"authority decodes to zero value", "authority", json.RawMessage(`{"namespace":"","identifier":""}`)},
		{"null configuration_reference", "configuration_reference", json.RawMessage("null")},
		{"malformed configuration_reference", "configuration_reference", json.RawMessage(`123`)},
		{"whitespace-only configuration_reference", "configuration_reference", json.RawMessage(`"   "`)},
		{"limitations with empty entry", "limitations", json.RawMessage(`[""]`)},
		{"null correction", "correction", json.RawMessage("null")},
		{"malformed correction", "correction", json.RawMessage(`123`)},
		{"self-correcting correction", "correction", json.RawMessage(`{"kind":"peos:correct","target":{"record_id":"BIND-1"}}`)},
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
			var decoded BindingRecord
			if err := json.Unmarshal(modified, &decoded); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}
}

func TestBindingRecordUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	originalData, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`not json`), &b); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestBindingRecordNoForbiddenWireKeys(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"bound", "active", "active_deployment", "deployed", "current",
		"latest", "effective", "status", "state", "lifecycle", "compliant",
		"compliance", "superseded", "incident",
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

// --- UnbindingRecord -----------------------------------------------------------

func TestNewUnbindingRecord(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	if u.IsZero() {
		t.Error("valid UnbindingRecord reports IsZero() = true")
	}
	ref, err := u.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.RecordID() != u.ID() {
		t.Error("Ref() mismatch")
	}
	if u.Binding() != bindingRef {
		t.Error("Binding() mismatch")
	}
	if u.Subject() != mustRuntimeSubjectRef(t, "kubernetes", "pod-1") {
		t.Error("Subject() mismatch")
	}
	if u.TerminatedAt() != mustTimestampAt(t, 10) {
		t.Error("TerminatedAt() mismatch")
	}
	if u.Actor() != mustActor(t, "peos-cli", "svc-1") {
		t.Error("Actor() mismatch")
	}
	if u.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if !u.Extension().IsZero() {
		t.Error("new UnbindingRecord should have zero extension")
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	u2 := u.WithExtension(ext)
	if u2.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	u3 := u2.WithoutExtension()
	if !u3.Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
	if u.Reason() != "scaled down" {
		t.Errorf("Reason() = %q", u.Reason())
	}
}

func TestNewUnbindingRecordZeroBindingRejected(t *testing.T) {
	_, err := NewUnbindingRecord(
		mustRuntimeUnbindingRecordID(t, "UNBIND-1"),
		core.RuntimeBindingRecordRef{},
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustTimestampAt(t, 10),
		"reason",
		mustActor(t, "peos-cli", "svc-1"),
		mustProvenance(t),
	)
	if !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero binding: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
}

func TestNewUnbindingRecordMandatoryFieldRejections(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	subject := mustRuntimeSubjectRef(t, "kubernetes", "pod-1")
	terminatedAt := mustTimestampAt(t, 10)
	actor := mustActor(t, "peos-cli", "svc-1")
	provenance := mustProvenance(t)

	if _, err := NewUnbindingRecord(core.RuntimeUnbindingRecordID{}, bindingRef, subject, terminatedAt, "reason", actor, provenance); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero id: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
	if _, err := NewUnbindingRecord(mustRuntimeUnbindingRecordID(t, "UNBIND-1"), bindingRef, core.RuntimeSubjectRef{}, terminatedAt, "reason", actor, provenance); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero subject: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
	if _, err := NewUnbindingRecord(mustRuntimeUnbindingRecordID(t, "UNBIND-1"), bindingRef, subject, core.Timestamp{}, "reason", actor, provenance); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero terminated-at: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
	if _, err := NewUnbindingRecord(mustRuntimeUnbindingRecordID(t, "UNBIND-1"), bindingRef, subject, terminatedAt, "   ", actor, provenance); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("whitespace-only reason: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
	if _, err := NewUnbindingRecord(mustRuntimeUnbindingRecordID(t, "UNBIND-1"), bindingRef, subject, terminatedAt, "reason", core.ActorRef{}, provenance); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero actor: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
	if _, err := NewUnbindingRecord(mustRuntimeUnbindingRecordID(t, "UNBIND-1"), bindingRef, subject, terminatedAt, "reason", actor, core.Provenance{}); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
}

func TestUnbindingRecordCorrectionAndSelfCorrection(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)

	if _, err := u.WithCorrection(core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef]{}); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("zero correction: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}

	earlierRef, err := core.NewRuntimeUnbindingRecordRef(mustRuntimeUnbindingRecordID(t, "UNBIND-0"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindReplace, earlierRef)
	if err != nil {
		t.Fatal(err)
	}
	u2, err := u.WithCorrection(correction)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := u2.Correction()
	if !ok || got.Kind() != core.CorrectionKindReplace {
		t.Errorf("Correction() = (%v, %v)", got, ok)
	}

	selfRef, err := core.NewRuntimeUnbindingRecordRef(mustRuntimeUnbindingRecordID(t, "UNBIND-1"))
	if err != nil {
		t.Fatal(err)
	}
	selfCorrection, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, selfRef)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := u.WithCorrection(selfCorrection); !errors.Is(err, core.ErrInvalidCorrectionReference) {
		t.Errorf("self-correction: error = %v, want %v", err, core.ErrInvalidCorrectionReference)
	}

	cleared := u2.WithoutCorrection()
	if _, ok := cleared.Correction(); ok {
		t.Error("WithoutCorrection did not clear the field")
	}
}

func TestUnbindingRecordWithAuthority(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	authority := mustAuthority(t)
	u2, err := u.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := u2.Authority()
	if !ok || got != authority {
		t.Errorf("Authority() = (%v, %v)", got, ok)
	}
	if _, err := u.WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero authority: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
	cleared := u2.WithoutAuthority()
	if _, ok := cleared.Authority(); ok {
		t.Error("WithoutAuthority did not clear the field")
	}
}

func TestUnbindingRecordDoesNotMutateBinding(t *testing.T) {
	b := mustBindingRecord(t, "BIND-1")
	bindingRef, err := b.Ref()
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	_ = mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	after, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("creating an UnbindingRecord mutated the BindingRecord it references")
	}
}

func TestUnbindingRecordMarshalZero(t *testing.T) {
	var u UnbindingRecord
	if _, err := json.Marshal(u); !errors.Is(err, ErrInvalidRuntimeUnbindingRecord) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeUnbindingRecord)
	}
}

func TestUnbindingRecordJSONRoundTrip(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	u, err = u.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	u = u.WithExtension(ext)
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	var decoded UnbindingRecord
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

func TestUnbindingRecordUnmarshalFieldSpecificRejections(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	u, err = u.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	earlierRef, err := core.NewRuntimeUnbindingRecordRef(mustRuntimeUnbindingRecordID(t, "UNBIND-0"))
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindReplace, earlierRef)
	if err != nil {
		t.Fatal(err)
	}
	u, err = u.WithCorrection(correction)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(u)
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
		{"null authority", "authority", json.RawMessage("null")},
		{"malformed authority", "authority", json.RawMessage(`123`)},
		{"authority decodes to zero value", "authority", json.RawMessage(`{"namespace":"","identifier":""}`)},
		{"null correction", "correction", json.RawMessage("null")},
		{"malformed correction", "correction", json.RawMessage(`123`)},
		{"self-correcting correction", "correction", json.RawMessage(`{"kind":"peos:replace","target":{"record_id":"UNBIND-1"}}`)},
		{"malformed binding", "binding", json.RawMessage(`123`)},
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
			var decoded UnbindingRecord
			if err := json.Unmarshal(modified, &decoded); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}
}

func TestUnbindingRecordUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	originalData, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"reason":""}`), &u); err == nil {
		t.Fatal("empty reason accepted, want error")
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestUnbindingRecordNoForbiddenWireKeys(t *testing.T) {
	bindingRef, err := mustBindingRecord(t, "BIND-1").Ref()
	if err != nil {
		t.Fatal(err)
	}
	u := mustUnbindingRecord(t, "UNBIND-1", bindingRef)
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"unbound", "current", "latest", "effective", "status", "state"}
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
