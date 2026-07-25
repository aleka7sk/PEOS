package core

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProvenancePartialFields(t *testing.T) {
	p := NewProvenance()
	if !p.IsZero() {
		t.Error("empty Provenance reports IsZero() = false")
	}

	ts, err := NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	p = p.WithRecordedAt(ts)

	if _, ok := p.Source(); ok {
		t.Error("Source() ok = true for unset field")
	}
	if got, ok := p.RecordedAt(); !ok || !got.Equal(ts) {
		t.Errorf("RecordedAt() = (%v, %v), want (%v, true)", got, ok, ts)
	}
	if p.IsZero() {
		t.Error("Provenance with one field set reports IsZero() = true")
	}
}

func TestProvenanceJSONOmitsUnsetFields(t *testing.T) {
	actor, err := NewActorRef("peos-cli", "svc-account-1")
	if err != nil {
		t.Fatal(err)
	}
	p := NewProvenance().WithActor(actor)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["recorded_at"]; present {
		t.Error("recorded_at present in JSON despite being unset")
	}
	if _, present := raw["actor"]; !present {
		t.Error("actor missing from JSON despite being set")
	}
}

func TestProvenanceJSONRoundTrip(t *testing.T) {
	actor, err := NewActorRef("peos-cli", "svc-account-1")
	if err != nil {
		t.Fatal(err)
	}
	ts, err := NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	source := mustVocabularyValue(t, "peos", "manual-entry")
	original := NewProvenance().WithActor(actor).WithRecordedAt(ts).WithSource(source)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Provenance
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	gotActor, ok := decoded.Actor()
	if !ok || gotActor != actor {
		t.Errorf("Actor() = (%v, %v), want (%v, true)", gotActor, ok, actor)
	}
	gotSource, ok := decoded.Source()
	if !ok || !gotSource.Equal(source) {
		t.Errorf("Source() = (%v, %v), want (%v, true)", gotSource, ok, source)
	}
}

func TestNewAuthorityRef(t *testing.T) {
	a, err := NewAuthorityRef("peos-role", "architecture-lead")
	if err != nil {
		t.Fatal(err)
	}
	if a.IsZero() {
		t.Error("valid AuthorityRef reports IsZero() = true")
	}
	if _, ok := a.Kind(); ok {
		t.Error("Kind() ok = true before WithKind was called")
	}

	kind := mustVocabularyValue(t, "peos", "role-based")
	a = a.WithKind(kind)
	got, ok := a.Kind()
	if !ok || !got.Equal(kind) {
		t.Errorf("Kind() = (%v, %v), want (%v, true)", got, ok, kind)
	}
}

func TestAuthorityRefJSONRoundTrip(t *testing.T) {
	kind := mustVocabularyValue(t, "peos", "role-based")
	original, err := NewAuthorityRef("peos-role", "architecture-lead")
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithKind(kind)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded AuthorityRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Namespace() != original.Namespace() || decoded.Identifier() != original.Identifier() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
	gotKind, ok := decoded.Kind()
	if !ok || !gotKind.Equal(kind) {
		t.Errorf("Kind() = (%v, %v), want (%v, true)", gotKind, ok, kind)
	}
}

func TestNewAuthorityRefRejectsEmpty(t *testing.T) {
	if _, err := NewAuthorityRef("", "x"); err == nil {
		t.Error("empty namespace accepted, want error")
	}
	if _, err := NewAuthorityRef("x", ""); err == nil {
		t.Error("empty identifier accepted, want error")
	}
}
