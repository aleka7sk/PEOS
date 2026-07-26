package validation

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- H.2 shared helpers ------------------------------------------------------

func mustTimestamp(t *testing.T, value string) core.Timestamp {
	t.Helper()
	ts, err := core.ParseTimestamp(value)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func mustActorRef(t *testing.T, namespace, identifier string) core.ActorRef {
	t.Helper()
	a, err := core.NewActorRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustExecutionRecordID(t *testing.T, value string) core.ValidationExecutionRecordID {
	t.Helper()
	id, err := core.NewValidationExecutionRecordID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustExecutionRecordRef(t *testing.T, value string) core.ValidationExecutionRecordRef {
	t.Helper()
	ref, err := core.NewValidationExecutionRecordRef(mustExecutionRecordID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustPlanRevisionRef(t *testing.T, artifactID, revisionID string) core.ValidationPlanRevisionRef {
	t.Helper()
	ref, err := core.NewValidationPlanRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
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

func mustPlannedRef(t *testing.T) ActivityReference {
	t.Helper()
	a, err := NewPlannedActivityReference(mustPlanRevisionRef(t, "VP-1", "REV-4"), mustLocalKey(t, "ACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustAdHocRef(t *testing.T) ActivityReference {
	t.Helper()
	a, err := NewAdHocActivityReference("exploratory smoke run")
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustExecutionEvent(t *testing.T, at, note string) ExecutionEvent {
	t.Helper()
	e, err := NewExecutionEvent(mustTimestamp(t, at), note)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func mustExecutionCorrection(t *testing.T, kind core.CorrectionKind, target string) core.RecordCorrectionRef[core.ValidationExecutionRecordRef] {
	t.Helper()
	c, err := core.NewRecordCorrectionRef(kind, mustExecutionRecordRef(t, target))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustExecutionRecord(t *testing.T) ExecutionRecord {
	t.Helper()
	r, err := NewExecutionRecord(
		mustExecutionRecordID(t, "EXR-1"),
		mustPlannedRef(t),
		mustSubject(t, "AR-42", "REV-3"),
		mustMethod(t, "test"),
		core.ExecutionOutcomeCompleted,
		mustTimestamp(t, "2026-07-27T10:15:00Z"),
		mustActorRef(t, "ci", "runner-9"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func fullExecutionRecord(t *testing.T) ExecutionRecord {
	t.Helper()
	r := mustExecutionRecord(t)
	var err error
	if r, err = r.WithStartedAt(mustTimestamp(t, "2026-07-27T10:00:00Z")); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-7", "REV-2")}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithEvents([]ExecutionEvent{mustExecutionEvent(t, "2026-07-27T10:05:00Z", "step 1 ok")}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithProducedEvidence([]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithReliedUponEvidence([]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-2", "REV-1")}); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithAuthority(mustAuthorityRef(t, "org", "qa-board")); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithEnvironment("ci-linux-amd64"); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithLimitations("network mocked"); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithUncertainty("timing jitter under load"); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithCorrection(mustExecutionCorrection(t, core.CorrectionKindReplace, "EXR-0")); err != nil {
		t.Fatal(err)
	}
	return r.WithExtension(mustExtension(t, "acme", `{"runner":"9"}`))
}

// --- ActivityReference -------------------------------------------------------

func TestNewPlannedActivityReferenceValid(t *testing.T) {
	a := mustPlannedRef(t)
	if a.IsZero() {
		t.Fatal("valid planned ActivityReference reports IsZero")
	}
	if got := a.Kind(); got != "planned" {
		t.Errorf("Kind() = %q, want planned", got)
	}
	rev, key, ok := a.AsPlanned()
	if !ok {
		t.Fatal("AsPlanned reported false")
	}
	if rev.ArtifactID().String() != "VP-1" || rev.RevisionID().String() != "REV-4" || key.String() != "ACT-1" {
		t.Errorf("AsPlanned() = %v, %v", rev, key)
	}
	if _, ok := a.AsAdHoc(); ok {
		t.Error("planned arm reported AsAdHoc")
	}
}

func TestNewAdHocActivityReferenceValid(t *testing.T) {
	a := mustAdHocRef(t)
	if got := a.Kind(); got != "ad_hoc" {
		t.Errorf("Kind() = %q, want ad_hoc", got)
	}
	designation, ok := a.AsAdHoc()
	if !ok || designation != "exploratory smoke run" {
		t.Errorf("AsAdHoc() = %q, %v", designation, ok)
	}
	if _, _, ok := a.AsPlanned(); ok {
		t.Error("ad hoc arm reported AsPlanned")
	}
}

func TestNewAdHocActivityReferenceTrims(t *testing.T) {
	a, err := NewAdHocActivityReference("  spot check  ")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := a.AsAdHoc()
	if got != "spot check" {
		t.Errorf("designation = %q, want %q", got, "spot check")
	}
}

func TestNewPlannedActivityReferenceIncompleteRejected(t *testing.T) {
	if _, err := NewPlannedActivityReference(core.ValidationPlanRevisionRef{}, mustLocalKey(t, "ACT-1")); !errors.Is(err, ErrInvalidActivityReference) {
		t.Errorf("zero plan revision: error = %v, want %v", err, ErrInvalidActivityReference)
	}
	if _, err := NewPlannedActivityReference(mustPlanRevisionRef(t, "VP-1", "REV-4"), core.LocalKey{}); !errors.Is(err, ErrInvalidActivityReference) {
		t.Errorf("zero key: error = %v, want %v", err, ErrInvalidActivityReference)
	}
}

func TestNewAdHocActivityReferenceEmptyRejected(t *testing.T) {
	for _, value := range []string{"", "   ", "\t\n"} {
		if _, err := NewAdHocActivityReference(value); !errors.Is(err, ErrInvalidActivityReference) {
			t.Errorf("designation %q: error = %v, want %v", value, err, ErrInvalidActivityReference)
		}
	}
}

func TestActivityReferenceZeroInvalid(t *testing.T) {
	var a ActivityReference
	if !a.IsZero() {
		t.Error("zero ActivityReference does not report IsZero")
	}
	if got := a.Kind(); got != "" {
		t.Errorf("zero Kind() = %q, want empty", got)
	}
	if _, _, ok := a.AsPlanned(); ok {
		t.Error("zero reported AsPlanned")
	}
	if _, ok := a.AsAdHoc(); ok {
		t.Error("zero reported AsAdHoc")
	}
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidActivityReference) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidActivityReference)
	}
}

func TestActivityReferenceJSONKeys(t *testing.T) {
	planned, err := json.Marshal(mustPlannedRef(t))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"ACT-1"}}`
	if string(planned) != want {
		t.Errorf("planned wire form =\n%s\nwant\n%s", planned, want)
	}

	adHoc, err := json.Marshal(mustAdHocRef(t))
	if err != nil {
		t.Fatal(err)
	}
	wantAdHoc := `{"kind":"ad_hoc","ref":{"designation":"exploratory smoke run"}}`
	if string(adHoc) != wantAdHoc {
		t.Errorf("ad hoc wire form =\n%s\nwant\n%s", adHoc, wantAdHoc)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(planned, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != 2 {
		t.Errorf("envelope has %d keys, want 2", len(raw))
	}
	if _, ok := raw["type"]; ok {
		t.Error("ActivityReference must not carry a top-level type discriminator")
	}
}

func TestActivityReferenceJSONRoundTrip(t *testing.T) {
	for name, original := range map[string]ActivityReference{
		"planned": mustPlannedRef(t),
		"ad_hoc":  mustAdHocRef(t),
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decoded ActivityReference
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != original.Kind() {
				t.Errorf("kind mismatch: %q vs %q", decoded.Kind(), original.Kind())
			}
			gotRev, gotKey, gotOK := decoded.AsPlanned()
			wantRev, wantKey, wantOK := original.AsPlanned()
			if gotOK != wantOK || gotRev != wantRev || gotKey != wantKey {
				t.Error("planned arm mismatch")
			}
			gotD, gotOK := decoded.AsAdHoc()
			wantD, wantOK := original.AsAdHoc()
			if gotOK != wantOK || gotD != wantD {
				t.Error("ad hoc arm mismatch")
			}
		})
	}
}

func TestActivityReferenceJSONRejections(t *testing.T) {
	cases := map[string]string{
		"unknown kind":            `{"kind":"guessed","ref":{"designation":"x"}}`,
		"missing kind":            `{"ref":{"designation":"x"}}`,
		"missing ref":             `{"kind":"ad_hoc"}`,
		"null ref":                `{"kind":"ad_hoc","ref":null}`,
		"null document":           `null`,
		"planned missing key":     `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"}}}`,
		"planned missing plan":    `{"kind":"planned","ref":{"key":"ACT-1"}}`,
		"planned empty key":       `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"  "}}`,
		"ad hoc empty":            `{"kind":"ad_hoc","ref":{"designation":"   "}}`,
		"ad hoc missing":          `{"kind":"ad_hoc","ref":{}}`,
		"planned rev missing rev": `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1"},"key":"ACT-1"}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var a ActivityReference
			err := json.Unmarshal([]byte(payload), &a)
			if err == nil {
				t.Fatal("payload accepted, want error")
			}
			if !errors.Is(err, ErrInvalidActivityReference) {
				t.Errorf("error = %v, want wrapping %v", err, ErrInvalidActivityReference)
			}
		})
	}
}

// TestActivityReferenceJSONCrossArmContaminationIgnored proves neither arm can
// pick up the other's state: extra keys are simply unknown fields for the arm
// being decoded.
func TestActivityReferenceJSONCrossArmContaminationIgnored(t *testing.T) {
	var planned ActivityReference
	payload := `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"ACT-1","designation":"sneaky"}}`
	if err := json.Unmarshal([]byte(payload), &planned); err != nil {
		t.Fatal(err)
	}
	if _, ok := planned.AsAdHoc(); ok {
		t.Error("planned arm absorbed ad hoc designation")
	}

	var adHoc ActivityReference
	payload2 := `{"kind":"ad_hoc","ref":{"designation":"real","plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"ACT-1"}}`
	if err := json.Unmarshal([]byte(payload2), &adHoc); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := adHoc.AsPlanned(); ok {
		t.Error("ad hoc arm absorbed planned state")
	}
	if d, _ := adHoc.AsAdHoc(); d != "real" {
		t.Errorf("designation = %q", d)
	}
}

func TestActivityReferenceJSONNestedSentinelPreserved(t *testing.T) {
	var a ActivityReference
	err := json.Unmarshal([]byte(`{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"   "}}`), &a)
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("error = %v, want wrapping %v", err, core.ErrEmptyIdentity)
	}
}

func TestActivityReferenceFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := mustPlannedRef(t)
	if err := json.Unmarshal([]byte(`{"kind":"guessed","ref":{}}`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.Kind() != "planned" {
		t.Error("failed Unmarshal disturbed the receiver")
	}
}

// --- ExecutionEvent ----------------------------------------------------------

func TestNewExecutionEventValid(t *testing.T) {
	e := mustExecutionEvent(t, "2026-07-27T10:05:00Z", "  step 1 ok  ")
	if e.IsZero() {
		t.Fatal("valid ExecutionEvent reports IsZero")
	}
	if e.Note() != "step 1 ok" {
		t.Errorf("Note() = %q, want trimmed", e.Note())
	}
	if e.At().IsZero() {
		t.Error("At() is zero")
	}
	if !e.Extension().IsZero() {
		t.Error("new event carries extension data")
	}
}

func TestNewExecutionEventRejections(t *testing.T) {
	if _, err := NewExecutionEvent(core.Timestamp{}, "note"); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero timestamp: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
	for _, note := range []string{"", "   ", "\t"} {
		if _, err := NewExecutionEvent(mustTimestamp(t, "2026-07-27T10:00:00Z"), note); !errors.Is(err, ErrInvalidExecutionRecord) {
			t.Errorf("note %q: error = %v, want %v", note, err, ErrInvalidExecutionRecord)
		}
	}
}

func TestExecutionEventExtensionModifiers(t *testing.T) {
	original := mustExecutionEvent(t, "2026-07-27T10:05:00Z", "note")
	withExt := original.WithExtension(mustExtension(t, "acme", `{"k":1}`))
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set")
	}
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}
	if !withExt.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear")
	}
}

func TestExecutionEventZeroAndJSON(t *testing.T) {
	var zero ExecutionEvent
	if !zero.IsZero() {
		t.Error("zero ExecutionEvent does not report IsZero")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidExecutionRecord)
	}

	original := mustExecutionEvent(t, "2026-07-27T10:05:00Z", "note").WithExtension(mustExtension(t, "acme", `{"k":1}`))
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"at", "note", "extension"}
	if len(raw) != len(want) {
		t.Errorf("wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, forbidden := range []string{"id", "severity", "criterion", "outcome", "actor", "provenance", "lifecycle", "state"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("ExecutionEvent unexpectedly carries %q", forbidden)
		}
	}

	var decoded ExecutionEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.At().Equal(original.At()) || decoded.Note() != original.Note() {
		t.Error("round trip mismatch")
	}
}

func TestExecutionEventJSONRejections(t *testing.T) {
	for name, payload := range map[string]string{
		"missing at":   `{"note":"x"}`,
		"null at":      `{"at":null,"note":"x"}`,
		"missing note": `{"at":"2026-07-27T10:00:00Z"}`,
		"null note":    `{"at":"2026-07-27T10:00:00Z","note":null}`,
		"empty note":   `{"at":"2026-07-27T10:00:00Z","note":"   "}`,
		"not json":     `[]`,
	} {
		t.Run(name, func(t *testing.T) {
			var e ExecutionEvent
			if err := json.Unmarshal([]byte(payload), &e); !errors.Is(err, ErrInvalidExecutionRecord) {
				t.Errorf("error = %v, want %v", err, ErrInvalidExecutionRecord)
			}
		})
	}
}

func TestExecutionEventFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := mustExecutionEvent(t, "2026-07-27T10:05:00Z", "original")
	if err := json.Unmarshal([]byte(`{"at":null,"note":"x"}`), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.Note() != "original" {
		t.Error("failed Unmarshal disturbed the receiver")
	}
}

// --- ExecutionRecord construction --------------------------------------------

func TestNewExecutionRecordValidMinimum(t *testing.T) {
	r := mustExecutionRecord(t)
	if r.IsZero() {
		t.Fatal("valid ExecutionRecord reports IsZero")
	}
	if r.ID().String() != "EXR-1" {
		t.Errorf("ID() = %q", r.ID().String())
	}
	ref, err := r.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.RecordID() != r.ID() {
		t.Error("Ref() does not cite the record")
	}
	if r.Activity().Kind() != "planned" {
		t.Error("activity not preserved")
	}
	if !r.Outcome().Equal(core.ExecutionOutcomeCompleted) {
		t.Error("outcome not preserved")
	}
	if r.Actor().Identifier() != "runner-9" {
		t.Error("actor not preserved")
	}
	if _, ok := r.StartedAt(); ok {
		t.Error("minimum record reports a started timestamp")
	}
	for _, n := range [][]any{{len(r.Criteria())}, {len(r.Events())}, {len(r.ProducedEvidence())}, {len(r.ReliedUponEvidence())}} {
		if n[0].(int) != 0 {
			t.Error("minimum record has a non-empty optional collection")
		}
	}
	if _, ok := r.Authority(); ok {
		t.Error("minimum record reports authority")
	}
	if _, ok := r.Environment(); ok {
		t.Error("minimum record reports environment")
	}
	if _, ok := r.Limitations(); ok {
		t.Error("minimum record reports limitations")
	}
	if _, ok := r.Uncertainty(); ok {
		t.Error("minimum record reports uncertainty")
	}
	if _, ok := r.Correction(); ok {
		t.Error("minimum record reports a correction")
	}
	if !r.Extension().IsZero() {
		t.Error("minimum record has extension data")
	}
}

func TestNewExecutionRecordMandatoryRejections(t *testing.T) {
	id := mustExecutionRecordID(t, "EXR-1")
	activity := mustPlannedRef(t)
	subject := mustSubject(t, "AR-42", "REV-3")
	method := mustMethod(t, "test")
	outcome := core.ExecutionOutcomeCompleted
	completed := mustTimestamp(t, "2026-07-27T10:15:00Z")
	actor := mustActorRef(t, "ci", "runner-9")
	prov := mustProvenance(t)

	cases := []struct {
		name string
		call func() (ExecutionRecord, error)
		want error
	}{
		{"zero id", func() (ExecutionRecord, error) {
			return NewExecutionRecord(core.ValidationExecutionRecordID{}, activity, subject, method, outcome, completed, actor, prov)
		}, ErrInvalidExecutionRecord},
		{"zero activity", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, ActivityReference{}, subject, method, outcome, completed, actor, prov)
		}, ErrInvalidActivityReference},
		{"zero subject", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, activity, core.EngineeringSubjectRef{}, method, outcome, completed, actor, prov)
		}, ErrInvalidExecutionRecord},
		{"zero method", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, activity, subject, core.ValidationMethod{}, outcome, completed, actor, prov)
		}, ErrInvalidExecutionRecord},
		{"zero outcome", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, activity, subject, method, core.ExecutionOutcome{}, completed, actor, prov)
		}, ErrInvalidExecutionRecord},
		{"zero completedAt", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, activity, subject, method, outcome, core.Timestamp{}, actor, prov)
		}, ErrInvalidExecutionRecord},
		{"zero actor", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, activity, subject, method, outcome, completed, core.ActorRef{}, prov)
		}, ErrInvalidExecutionRecord},
		{"zero provenance", func() (ExecutionRecord, error) {
			return NewExecutionRecord(id, activity, subject, method, outcome, completed, actor, core.Provenance{})
		}, ErrInvalidExecutionRecord},
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

func TestNewExecutionRecordAcceptsAdHocActivity(t *testing.T) {
	r, err := NewExecutionRecord(
		mustExecutionRecordID(t, "EXR-2"), mustAdHocRef(t),
		mustSubject(t, "AR-42", "REV-3"), mustMethod(t, "inspection"),
		core.ExecutionOutcomeIndeterminate, mustTimestamp(t, "2026-07-27T11:00:00Z"),
		mustActorRef(t, "person", "reviewer-1"), mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Activity().AsAdHoc(); !ok {
		t.Error("ad hoc activity not preserved")
	}
}

func TestExecutionRecordAcceptsEveryPredeclaredOutcome(t *testing.T) {
	for _, outcome := range []core.ExecutionOutcome{
		core.ExecutionOutcomeCompleted, core.ExecutionOutcomeFailed,
		core.ExecutionOutcomeInterrupted, core.ExecutionOutcomeIndeterminate,
	} {
		if _, err := NewExecutionRecord(
			mustExecutionRecordID(t, "EXR-1"), mustPlannedRef(t),
			mustSubject(t, "AR-42", "REV-3"), mustMethod(t, "test"),
			outcome, mustTimestamp(t, "2026-07-27T10:15:00Z"),
			mustActorRef(t, "ci", "runner-9"), mustProvenance(t),
		); err != nil {
			t.Errorf("outcome %s rejected: %v", outcome, err)
		}
	}
}

func TestExecutionRecordIsZero(t *testing.T) {
	var r ExecutionRecord
	if !r.IsZero() {
		t.Error("zero ExecutionRecord does not report IsZero")
	}
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero marshal error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
	if r.Criteria() != nil || r.Events() != nil || r.ProducedEvidence() != nil || r.ReliedUponEvidence() != nil {
		t.Error("zero record returns non-nil collections")
	}
}

// --- temporal rule -----------------------------------------------------------

func TestExecutionRecordStartedAtTemporalRule(t *testing.T) {
	base := mustExecutionRecord(t) // completedAt 10:15

	if _, err := base.WithStartedAt(mustTimestamp(t, "2026-07-27T10:00:00Z")); err != nil {
		t.Errorf("started before completed rejected: %v", err)
	}
	// Equality is permitted: PEOS-006 imposes no minimum duration.
	if _, err := base.WithStartedAt(mustTimestamp(t, "2026-07-27T10:15:00Z")); err != nil {
		t.Errorf("started == completed rejected: %v", err)
	}
	if _, err := base.WithStartedAt(mustTimestamp(t, "2026-07-27T10:16:00Z")); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("started after completed: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
	if _, err := base.WithStartedAt(core.Timestamp{}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero started: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}

	withStart, err := base.WithStartedAt(mustTimestamp(t, "2026-07-27T10:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withStart.StartedAt()
	if !ok || !got.Equal(mustTimestamp(t, "2026-07-27T10:00:00Z")) {
		t.Error("StartedAt() did not round-trip")
	}
	if _, ok := withStart.WithoutStartedAt().StartedAt(); ok {
		t.Error("WithoutStartedAt did not clear")
	}
}

// --- ExecutionRecord modifiers -----------------------------------------------

func TestExecutionRecordCollectionModifiers(t *testing.T) {
	base := mustExecutionRecord(t)

	if _, err := base.WithCriteria([]core.CriterionRef{{}}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero criterion: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
	if _, err := base.WithEvents([]ExecutionEvent{{}}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero event: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
	if _, err := base.WithProducedEvidence([]core.EvidenceArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero produced evidence: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
	if _, err := base.WithReliedUponEvidence([]core.EvidenceArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero relied-upon evidence: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}

	// Empty input clears, and is accepted for all four.
	full := fullExecutionRecord(t)
	var err error
	if full, err = full.WithCriteria(nil); err != nil {
		t.Fatal(err)
	}
	if full, err = full.WithEvents([]ExecutionEvent{}); err != nil {
		t.Fatal(err)
	}
	if full, err = full.WithProducedEvidence(nil); err != nil {
		t.Fatal(err)
	}
	if full, err = full.WithReliedUponEvidence([]core.EvidenceArtifactRevisionRef{}); err != nil {
		t.Fatal(err)
	}
	if len(full.Criteria()) != 0 || len(full.Events()) != 0 || len(full.ProducedEvidence()) != 0 || len(full.ReliedUponEvidence()) != 0 {
		t.Error("empty input did not clear a collection")
	}
}

func TestExecutionRecordEvidenceNotDeduplicated(t *testing.T) {
	dup := mustEvidence(t, "EV-1", "REV-1")
	r, err := mustExecutionRecord(t).WithProducedEvidence([]core.EvidenceArtifactRevisionRef{dup, dup})
	if err != nil {
		t.Fatalf("duplicate evidence rejected: %v", err)
	}
	if len(r.ProducedEvidence()) != 2 {
		t.Errorf("ProducedEvidence() length = %d, want 2 (no deduplication)", len(r.ProducedEvidence()))
	}
}

func TestExecutionRecordStringModifiers(t *testing.T) {
	base := mustExecutionRecord(t)
	cases := []struct {
		name string
		set  func(ExecutionRecord, string) (ExecutionRecord, error)
		get  func(ExecutionRecord) (string, bool)
		clr  func(ExecutionRecord) ExecutionRecord
	}{
		{"environment", ExecutionRecord.WithEnvironment, ExecutionRecord.Environment, ExecutionRecord.WithoutEnvironment},
		{"limitations", ExecutionRecord.WithLimitations, ExecutionRecord.Limitations, ExecutionRecord.WithoutLimitations},
		{"uncertainty", ExecutionRecord.WithUncertainty, ExecutionRecord.Uncertainty, ExecutionRecord.WithoutUncertainty},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set, err := tc.set(base, "  value  ")
			if err != nil {
				t.Fatal(err)
			}
			got, ok := tc.get(set)
			if !ok || got != "value" {
				t.Errorf("get = %q, %v; want trimmed value", got, ok)
			}
			if _, ok := tc.get(tc.clr(set)); ok {
				t.Error("clear did not clear")
			}
			for _, bad := range []string{"", "   ", "\t"} {
				if _, err := tc.set(base, bad); !errors.Is(err, ErrInvalidExecutionRecord) {
					t.Errorf("value %q: error = %v, want %v", bad, err, ErrInvalidExecutionRecord)
				}
			}
		})
	}
}

func TestExecutionRecordAuthorityModifier(t *testing.T) {
	auth := mustAuthorityRef(t, "org", "qa-board")
	r, err := mustExecutionRecord(t).WithAuthority(auth)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := r.Authority()
	if !ok || got != auth {
		t.Errorf("Authority() = %v, %v", got, ok)
	}
	if _, ok := r.WithoutAuthority().Authority(); ok {
		t.Error("WithoutAuthority did not clear")
	}
	if _, err := mustExecutionRecord(t).WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("zero authority: error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
}

func TestExecutionRecordCorrectionModifier(t *testing.T) {
	for _, kind := range []core.CorrectionKind{core.CorrectionKindCorrect, core.CorrectionKindReplace, core.CorrectionKindInvalidate} {
		correction := mustExecutionCorrection(t, kind, "EXR-0")
		r, err := mustExecutionRecord(t).WithCorrection(correction)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := r.Correction()
		if !ok {
			t.Fatal("Correction() reported false")
		}
		if !got.Kind().Value().Equal(kind.Value()) {
			t.Errorf("correction kind = %v, want %v", got.Kind(), kind)
		}
		if got.Target().RecordID().String() != "EXR-0" {
			t.Errorf("correction target = %v", got.Target())
		}
		if _, ok := r.WithoutCorrection().Correction(); ok {
			t.Error("WithoutCorrection did not clear")
		}
	}
	if _, err := mustExecutionRecord(t).WithCorrection(core.RecordCorrectionRef[core.ValidationExecutionRecordRef]{}); !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Error("zero correction accepted")
	}
}

func TestExecutionRecordExtensionModifier(t *testing.T) {
	r := mustExecutionRecord(t).WithExtension(mustExtension(t, "acme", `{"k":1}`))
	if r.Extension().IsZero() {
		t.Error("WithExtension did not set")
	}
	if !r.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear")
	}
}

func TestExecutionRecordModifierReceiverImmutability(t *testing.T) {
	original := mustExecutionRecord(t)

	if _, err := original.WithStartedAt(mustTimestamp(t, "2026-07-27T10:00:00Z")); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.StartedAt(); ok {
		t.Error("WithStartedAt mutated the receiver")
	}
	if _, err := original.WithCriteria([]core.CriterionRef{mustCriterion(t, "REQ-1", "REV-1")}); err != nil {
		t.Fatal(err)
	}
	if len(original.Criteria()) != 0 {
		t.Error("WithCriteria mutated the receiver")
	}
	if _, err := original.WithEvents([]ExecutionEvent{mustExecutionEvent(t, "2026-07-27T10:05:00Z", "x")}); err != nil {
		t.Fatal(err)
	}
	if len(original.Events()) != 0 {
		t.Error("WithEvents mutated the receiver")
	}
	if _, err := original.WithProducedEvidence([]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")}); err != nil {
		t.Fatal(err)
	}
	if len(original.ProducedEvidence()) != 0 {
		t.Error("WithProducedEvidence mutated the receiver")
	}
	if _, err := original.WithReliedUponEvidence([]core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-2", "REV-1")}); err != nil {
		t.Fatal(err)
	}
	if len(original.ReliedUponEvidence()) != 0 {
		t.Error("WithReliedUponEvidence mutated the receiver")
	}
	if _, err := original.WithAuthority(mustAuthorityRef(t, "org", "board")); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Authority(); ok {
		t.Error("WithAuthority mutated the receiver")
	}
	if _, err := original.WithEnvironment("env"); err != nil {
		t.Fatal(err)
	}
	if _, ok := original.Environment(); ok {
		t.Error("WithEnvironment mutated the receiver")
	}
	if _, err := original.WithCorrection(mustExecutionCorrection(t, core.CorrectionKindCorrect, "EXR-0")); err != nil {
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

func TestExecutionRecordCollectionsDefensivelyCopied(t *testing.T) {
	criteria := []core.CriterionRef{mustCriterion(t, "REQ-1", "REV-1")}
	events := []ExecutionEvent{mustExecutionEvent(t, "2026-07-27T10:05:00Z", "original")}
	produced := []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-1", "REV-1")}
	relied := []core.EvidenceArtifactRevisionRef{mustEvidence(t, "EV-2", "REV-1")}

	r, err := mustExecutionRecord(t).WithCriteria(criteria)
	if err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithEvents(events); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithProducedEvidence(produced); err != nil {
		t.Fatal(err)
	}
	if r, err = r.WithReliedUponEvidence(relied); err != nil {
		t.Fatal(err)
	}

	criteria[0] = core.CriterionRef{}
	events[0] = ExecutionEvent{}
	produced[0] = core.EvidenceArtifactRevisionRef{}
	relied[0] = core.EvidenceArtifactRevisionRef{}

	if r.Criteria()[0].IsZero() || r.Events()[0].IsZero() || r.ProducedEvidence()[0].IsZero() || r.ReliedUponEvidence()[0].IsZero() {
		t.Error("an input slice was not copied in")
	}

	r.Criteria()[0] = core.CriterionRef{}
	r.Events()[0] = ExecutionEvent{}
	r.ProducedEvidence()[0] = core.EvidenceArtifactRevisionRef{}
	r.ReliedUponEvidence()[0] = core.EvidenceArtifactRevisionRef{}

	if r.Criteria()[0].IsZero() || r.Events()[0].IsZero() || r.ProducedEvidence()[0].IsZero() || r.ReliedUponEvidence()[0].IsZero() {
		t.Error("an accessor did not return a defensive copy")
	}
}

// TestExecutionRecordCorrectionDoesNotMutateOriginal proves the historical
// preservation property structurally: recording a correcting record leaves the
// corrected record's own value untouched, because the correction reference
// lives on the new record and points backward.
func TestExecutionRecordCorrectionDoesNotMutateOriginal(t *testing.T) {
	earlier := mustExecutionRecord(t)
	earlierSnapshot, err := json.Marshal(earlier)
	if err != nil {
		t.Fatal(err)
	}

	later, err := NewExecutionRecord(
		mustExecutionRecordID(t, "EXR-2"), mustPlannedRef(t),
		mustSubject(t, "AR-42", "REV-3"), mustMethod(t, "test"),
		core.ExecutionOutcomeCompleted, mustTimestamp(t, "2026-07-27T12:00:00Z"),
		mustActorRef(t, "ci", "runner-9"), mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	earlierRef, err := earlier.Ref()
	if err != nil {
		t.Fatal(err)
	}
	correction, err := core.NewRecordCorrectionRef(core.CorrectionKindCorrect, earlierRef)
	if err != nil {
		t.Fatal(err)
	}
	if later, err = later.WithCorrection(correction); err != nil {
		t.Fatal(err)
	}

	afterSnapshot, err := json.Marshal(earlier)
	if err != nil {
		t.Fatal(err)
	}
	if string(earlierSnapshot) != string(afterSnapshot) {
		t.Error("recording a correction altered the corrected record")
	}
	if _, ok := earlier.Correction(); ok {
		t.Error("the corrected record gained a correction reference")
	}
	got, ok := later.Correction()
	if !ok || got.Target().RecordID() != earlier.ID() {
		t.Error("the correcting record does not point at the earlier record")
	}
}

// --- ExecutionRecord JSON ----------------------------------------------------

func TestExecutionRecordJSONMinimumKeys(t *testing.T) {
	data, err := json.Marshal(mustExecutionRecord(t))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{"id", "activity", "subject", "method", "outcome", "completed_at", "actor", "provenance"}
	if len(raw) != len(want) {
		t.Errorf("minimum wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	forbidden := []string{
		"relation", "type", "artifact_id", "revision_id", "core",
		"lifecycle", "state", "status", "basis", "verdict",
		"started_at", "criteria", "events", "extension",
	}
	for _, k := range forbidden {
		if _, ok := raw[k]; ok {
			t.Errorf("minimum wire form unexpectedly carries %q", k)
		}
	}
}

func TestExecutionRecordJSONFullKeysAndRoundTrip(t *testing.T) {
	original := fullExecutionRecord(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"id", "activity", "subject", "method", "outcome", "completed_at", "actor", "provenance",
		"started_at", "criteria", "events", "produced_evidence", "relied_upon_evidence",
		"authority", "environment", "limitations", "uncertainty", "correction", "extension",
	}
	if len(raw) != len(want) {
		t.Errorf("full wire form has %d keys, want %d: %v", len(raw), len(want), raw)
	}
	for _, k := range want {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	for _, k := range []string{"relation", "type", "lifecycle", "state", "status", "basis", "verdict"} {
		if _, ok := raw[k]; ok {
			t.Errorf("full wire form unexpectedly carries %q", k)
		}
	}

	var decoded ExecutionRecord
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

func executionPayload(t *testing.T, overrides map[string]string) string {
	t.Helper()
	base := map[string]string{
		"id":           `"EXR-1"`,
		"activity":     `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"ACT-1"}}`,
		"subject":      `{"kind":"artifact_revision","ref":{"artifact_id":"AR-42","revision_id":"REV-3"}}`,
		"method":       `"peos:test"`,
		"outcome":      `"peos:completed"`,
		"completed_at": `"2026-07-27T10:15:00Z"`,
		"actor":        `{"namespace":"ci","identifier":"runner-9"}`,
		"provenance":   `{"actor":{"namespace":"peos-cli","identifier":"svc-1"},"recorded_at":"2026-07-27T00:00:00Z"}`,
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

func TestExecutionRecordJSONMandatoryMissingRejected(t *testing.T) {
	cases := map[string]error{
		"id":           ErrInvalidExecutionRecord,
		"activity":     ErrInvalidActivityReference,
		"subject":      ErrInvalidExecutionRecord,
		"method":       ErrInvalidExecutionRecord,
		"outcome":      ErrInvalidExecutionRecord,
		"completed_at": ErrInvalidExecutionRecord,
		"actor":        ErrInvalidExecutionRecord,
		"provenance":   ErrInvalidExecutionRecord,
	}
	for field, want := range cases {
		t.Run(field, func(t *testing.T) {
			var r ExecutionRecord
			err := json.Unmarshal([]byte(executionPayload(t, map[string]string{field: ""})), &r)
			if !errors.Is(err, want) {
				t.Errorf("error = %v, want %v", err, want)
			}
		})
	}
}

func TestExecutionRecordJSONMandatoryNullRejected(t *testing.T) {
	for _, field := range []string{"id", "activity", "subject", "method", "outcome", "completed_at", "actor", "provenance"} {
		t.Run(field, func(t *testing.T) {
			var r ExecutionRecord
			if err := json.Unmarshal([]byte(executionPayload(t, map[string]string{field: "null"})), &r); err == nil {
				t.Error("explicit null accepted, want error")
			}
		})
	}
}

// TestExecutionRecordJSONOptionalCollectionsEquivalent locks the documented
// decision that, for an ExecutionRecord's four optional collections, absent /
// null / [] all mean "none declared" -- PEOS-006 permits zero cardinality for
// each. This deliberately differs from Claim.criteria.
func TestExecutionRecordJSONOptionalCollectionsEquivalent(t *testing.T) {
	for _, field := range []string{"criteria", "events", "produced_evidence", "relied_upon_evidence"} {
		for _, value := range []string{"", "null", "[]"} {
			label := field + "=" + value
			if value == "" {
				label = field + "=absent"
			}
			t.Run(label, func(t *testing.T) {
				var r ExecutionRecord
				if err := json.Unmarshal([]byte(executionPayload(t, map[string]string{field: value})), &r); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(r.Criteria()) != 0 || len(r.Events()) != 0 || len(r.ProducedEvidence()) != 0 || len(r.ReliedUponEvidence()) != 0 {
					t.Error("decoded record has a non-empty optional collection")
				}
			})
		}
	}
}

func TestExecutionRecordJSONOptionalSingleNullRejected(t *testing.T) {
	for _, field := range []string{"started_at", "authority", "environment", "limitations", "uncertainty", "correction"} {
		t.Run(field, func(t *testing.T) {
			var r ExecutionRecord
			err := json.Unmarshal([]byte(executionPayload(t, map[string]string{field: "null"})), &r)
			if !errors.Is(err, ErrInvalidExecutionRecord) {
				t.Errorf("error = %v, want %v", err, ErrInvalidExecutionRecord)
			}
		})
	}
}

func TestExecutionRecordJSONExtensionNullMeansAbsent(t *testing.T) {
	var r ExecutionRecord
	if err := json.Unmarshal([]byte(executionPayload(t, map[string]string{"extension": "null"})), &r); err != nil {
		t.Fatalf("extension null rejected: %v", err)
	}
	if !r.Extension().IsZero() {
		t.Error("extension null did not decode as absent")
	}
}

func TestExecutionRecordJSONStartedAfterCompletedRejectedOnDecode(t *testing.T) {
	var r ExecutionRecord
	err := json.Unmarshal([]byte(executionPayload(t, map[string]string{"started_at": `"2026-07-27T11:00:00Z"`})), &r)
	if !errors.Is(err, ErrInvalidExecutionRecord) {
		t.Errorf("error = %v, want %v", err, ErrInvalidExecutionRecord)
	}
}

func TestExecutionRecordJSONNestedSentinelsPreserved(t *testing.T) {
	cases := []struct {
		name    string
		payload map[string]string
		want    error
	}{
		{"subject missing revision", map[string]string{"subject": `{"kind":"artifact_revision","ref":{"artifact_id":"AR-42"}}`}, core.ErrMissingRevisionID},
		{"subject missing discriminator", map[string]string{"subject": `{"ref":{"artifact_id":"AR-42"}}`}, core.ErrInvalidReferenceDiscriminator},
		{"method malformed vocabulary", map[string]string{"method": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"outcome malformed vocabulary", map[string]string{"outcome": `"no-colon"`}, core.ErrInvalidVocabularyValue},
		{"id empty identity", map[string]string{"id": `"  "`}, core.ErrEmptyIdentity},
		{"criterion missing revision", map[string]string{"criteria": `[{"kind":"requirement_revision","ref":{"artifact_id":"REQ-7"}}]`}, core.ErrMissingRevisionID},
		{"produced evidence missing revision", map[string]string{"produced_evidence": `[{"artifact_id":"EV-1"}]`}, core.ErrMissingRevisionID},
		{"authority empty identity", map[string]string{"authority": `{"namespace":"org","identifier":"  "}`}, core.ErrEmptyIdentity},
		{"correction malformed kind", map[string]string{"correction": `{"kind":"no-colon","target":{"record_id":"EXR-0"}}`}, core.ErrInvalidVocabularyValue},
		{"correction empty target", map[string]string{"correction": `{"kind":"peos:replace","target":{"record_id":"  "}}`}, core.ErrEmptyIdentity},
		{"activity nested empty key", map[string]string{"activity": `{"kind":"planned","ref":{"plan_revision":{"artifact_id":"VP-1","revision_id":"REV-4"},"key":"  "}}`}, core.ErrEmptyIdentity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r ExecutionRecord
			err := json.Unmarshal([]byte(executionPayload(t, tc.payload)), &r)
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want wrapping %v", err, tc.want)
			}
		})
	}
}

func TestExecutionRecordJSONStringFieldWrongTypeRejected(t *testing.T) {
	for _, field := range []string{"environment", "limitations", "uncertainty"} {
		t.Run(field, func(t *testing.T) {
			var r ExecutionRecord
			err := json.Unmarshal([]byte(executionPayload(t, map[string]string{field: `123`})), &r)
			if !errors.Is(err, ErrInvalidExecutionRecord) {
				t.Errorf("error = %v, want %v", err, ErrInvalidExecutionRecord)
			}
		})
	}
}

func TestExecutionRecordJSONEmptyStringFieldRejected(t *testing.T) {
	for _, field := range []string{"environment", "limitations", "uncertainty"} {
		t.Run(field, func(t *testing.T) {
			var r ExecutionRecord
			err := json.Unmarshal([]byte(executionPayload(t, map[string]string{field: `"   "`})), &r)
			if !errors.Is(err, ErrInvalidExecutionRecord) {
				t.Errorf("error = %v, want %v", err, ErrInvalidExecutionRecord)
			}
		})
	}
}

func TestExecutionRecordJSONMalformedDocumentRejected(t *testing.T) {
	for _, payload := range []string{`not json`, `[]`, `{"id":[]}`} {
		var r ExecutionRecord
		if err := json.Unmarshal([]byte(payload), &r); err == nil {
			t.Errorf("payload %s accepted, want error", payload)
		}
	}
}

func TestExecutionRecordJSONUnknownFieldIgnored(t *testing.T) {
	var r ExecutionRecord
	if err := json.Unmarshal([]byte(executionPayload(t, map[string]string{"unknown": `"x"`})), &r); err != nil {
		t.Fatalf("unknown field rejected: %v", err)
	}
}

func TestExecutionRecordFailedUnmarshalPreservesReceiver(t *testing.T) {
	receiver := fullExecutionRecord(t)
	before := receiver.ID()
	if err := json.Unmarshal([]byte(executionPayload(t, map[string]string{"outcome": "null"})), &receiver); err == nil {
		t.Fatal("expected failure")
	}
	if receiver.ID() != before {
		t.Error("failed Unmarshal disturbed the receiver")
	}
	if len(receiver.Criteria()) != 1 {
		t.Error("failed Unmarshal disturbed the receiver's criteria")
	}
}

func TestExecutionRecordJSONAdHocRoundTrip(t *testing.T) {
	payload := executionPayload(t, map[string]string{
		"activity": `{"kind":"ad_hoc","ref":{"designation":"spot check"}}`,
	})
	var r ExecutionRecord
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatal(err)
	}
	got, ok := r.Activity().AsAdHoc()
	if !ok || got != "spot check" {
		t.Errorf("ad hoc designation = %q, %v", got, ok)
	}
}

// --- structural absence ------------------------------------------------------

func TestExecutionRecordHasNoForbiddenAPI(t *testing.T) {
	forbidden := []string{
		"Core", "ArtifactID", "RevisionID",
		"Relation", "Source", "Target",
		"Lifecycle", "State", "Status",
		"WithID", "WithActivity", "WithSubject", "WithMethod",
		"WithOutcome", "WithCompletedAt", "WithActor", "WithProvenance",
		"WithoutActivity", "WithoutSubject", "WithoutMethod", "WithoutOutcome",
		"Basis", "Verdict", "Claim", "Waiver",
	}
	assertNoMethods(t, "ExecutionRecord", reflect.TypeOf(ExecutionRecord{}), forbidden)
}

func TestActivityReferenceAndExecutionEventHaveNoIdentityAPI(t *testing.T) {
	forbidden := []string{"ID", "Ref", "ArtifactID", "RevisionID", "Lifecycle", "State", "Provenance", "Core"}
	assertNoMethods(t, "ActivityReference", reflect.TypeOf(ActivityReference{}), forbidden)
	assertNoMethods(t, "ExecutionEvent", reflect.TypeOf(ExecutionEvent{}), forbidden)
	// ExecutionEvent must additionally not grow observation-shaped fields.
	assertNoMethods(t, "ExecutionEvent", reflect.TypeOf(ExecutionEvent{}), []string{"Severity", "Criterion", "Outcome", "Actor"})
}

func assertNoMethods(t *testing.T, label string, typ reflect.Type, forbidden []string) {
	t.Helper()
	for _, name := range forbidden {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("%s unexpectedly exposes %s", label, name)
		}
		if _, ok := reflect.PointerTo(typ).MethodByName(name); ok {
			t.Errorf("*%s unexpectedly exposes %s", label, name)
		}
	}
}

// TestExecutionOutcomeIsNotLifecycleShaped guards PEOS-006's requirement that
// an execution outcome is never a Lifecycle State or State Assignment: the
// field's type is a core vocabulary value, and the record exposes no
// lifecycle-shaped accessor.
func TestExecutionOutcomeIsNotLifecycleShaped(t *testing.T) {
	r := mustExecutionRecord(t)
	outcomeType := reflect.TypeOf(r.Outcome())
	if outcomeType != reflect.TypeOf(core.ExecutionOutcome{}) {
		t.Fatalf("Outcome() type = %v, want core.ExecutionOutcome", outcomeType)
	}
	if got := reflect.TypeOf(r).PkgPath(); !strings.HasSuffix(got, "/peos/validation") {
		t.Errorf("ExecutionRecord package = %q", got)
	}
	// Sanity: the outcome vocabulary carries no timestamped assignment shape.
	assertNoMethods(t, "core.ExecutionOutcome", outcomeType, []string{"EffectiveAt", "Subject", "DefinitionVersion", "Authority"})
}

var _ = time.Now // keep the time import meaningful if helpers change

// TestExecutionRecordAllAccessorsReturnConstructorInputs exercises every
// accessor against known inputs, including the four mandatory-value accessors
// not otherwise touched by the behavioral tests above.
func TestExecutionRecordAllAccessorsReturnConstructorInputs(t *testing.T) {
	id := mustExecutionRecordID(t, "EXR-9")
	activity := mustPlannedRef(t)
	subject := mustSubject(t, "AR-42", "REV-3")
	method := mustMethod(t, "analysis")
	outcome := core.ExecutionOutcomeFailed
	completed := mustTimestamp(t, "2026-07-27T10:15:00Z")
	actor := mustActorRef(t, "ci", "runner-9")
	prov := mustProvenance(t)

	r, err := NewExecutionRecord(id, activity, subject, method, outcome, completed, actor, prov)
	if err != nil {
		t.Fatal(err)
	}

	if r.ID() != id {
		t.Error("ID() mismatch")
	}
	if r.Activity().Kind() != activity.Kind() {
		t.Error("Activity() mismatch")
	}
	if r.Subject() != subject {
		t.Error("Subject() mismatch")
	}
	if !r.Method().Value().Equal(method.Value()) {
		t.Error("Method() mismatch")
	}
	if !r.Outcome().Equal(outcome) {
		t.Error("Outcome() mismatch")
	}
	if !r.CompletedAt().Equal(completed) {
		t.Error("CompletedAt() mismatch")
	}
	if r.Actor() != actor {
		t.Error("Actor() mismatch")
	}
	if r.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	gotActor, ok := r.Provenance().Actor()
	if !ok || gotActor.Identifier() != "svc-1" {
		t.Errorf("Provenance().Actor() = %v, %v", gotActor, ok)
	}
}
