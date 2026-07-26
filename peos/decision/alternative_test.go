package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewAlternativeValidMinimum(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	if a.Statement() != "Use MySQL" {
		t.Errorf("Statement() = %q, want %q", a.Statement(), "Use MySQL")
	}
	if _, ok := a.Note(); ok {
		t.Error("Note() ok = true, want false")
	}
}

func TestNewAlternativeEmptyStatementRejected(t *testing.T) {
	if _, err := NewAlternative(""); !errors.Is(err, ErrInvalidAlternative) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAlternative)
	}
	if _, err := NewAlternative("   "); !errors.Is(err, ErrInvalidAlternative) {
		t.Errorf("whitespace: error = %v, want %v", err, ErrInvalidAlternative)
	}
}

func TestAlternativeNotePresentAbsent(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Note(); ok {
		t.Error("Note() ok = true before WithNote")
	}
	withNote, err := a.WithNote("rejected for licensing reasons")
	if err != nil {
		t.Fatal(err)
	}
	note, ok := withNote.Note()
	if !ok || note != "rejected for licensing reasons" {
		t.Errorf("Note() = (%q, %v), want (%q, true)", note, ok, "rejected for licensing reasons")
	}
	if _, ok := a.Note(); ok {
		t.Error("WithNote mutated the original receiver")
	}
}

func TestAlternativeEmptyNoteRejected(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.WithNote(""); !errors.Is(err, ErrInvalidAlternative) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAlternative)
	}
}

func TestAlternativeWithoutNote(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithNote("some note")
	if err != nil {
		t.Fatal(err)
	}
	cleared := a.WithoutNote()
	if _, ok := cleared.Note(); ok {
		t.Error("Note() ok = true after WithoutNote")
	}
	if _, ok := a.Note(); !ok {
		t.Error("WithoutNote mutated the original receiver")
	}
}

func TestAlternativeExtension(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := a.WithExtension(ext)
	if !a.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestAlternativeJSONRoundTrip(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithNote("rejected for licensing reasons")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Alternative
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != a.Statement() {
		t.Errorf("Statement mismatch: got %q, want %q", decoded.Statement(), a.Statement())
	}
	note, ok := decoded.Note()
	wantNote, wantOK := a.Note()
	if ok != wantOK || note != wantNote {
		t.Errorf("Note mismatch: got (%q,%v), want (%q,%v)", note, ok, wantNote, wantOK)
	}
}

func TestAlternativeNoteOmittedWhenAbsent(t *testing.T) {
	a, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, optional := range []string{"note", "extension"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}
}

func TestAlternativeExplicitNullRejected(t *testing.T) {
	var a Alternative
	if err := json.Unmarshal([]byte(`{"statement":"Use MySQL","note":null}`), &a); !errors.Is(err, ErrInvalidAlternative) {
		t.Errorf("null note: error = %v, want %v", err, ErrInvalidAlternative)
	}
	if err := json.Unmarshal([]byte(`{"statement":"Use MySQL","extension":null}`), &a); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestAlternativeZeroMarshalRejected(t *testing.T) {
	var a Alternative
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidAlternative) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAlternative)
	}
}

func TestAlternativeUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithNote("some note")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"statement":""}`), &receiver); err == nil {
		t.Fatal("empty statement accepted, want error")
	}
	if receiver.Statement() != original.Statement() {
		t.Error("failed Unmarshal changed receiver")
	}
	gotNote, gotOK := receiver.Note()
	wantNote, wantOK := original.Note()
	if gotOK != wantOK || gotNote != wantNote {
		t.Error("failed Unmarshal changed receiver's note")
	}
}
