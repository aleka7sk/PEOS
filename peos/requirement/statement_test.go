package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewStatementValid(t *testing.T) {
	s, err := NewStatement("The service shall retain audit records.")
	if err != nil {
		t.Fatal(err)
	}
	if s.IsZero() {
		t.Error("valid Statement reports IsZero() = true")
	}
	if got, want := s.Text(), "The service shall retain audit records."; got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}

func TestNewStatementTrimsSurroundingWhitespace(t *testing.T) {
	s, err := NewStatement("  \t hello world \n  ")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := s.Text(), "hello world"; got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}

func TestNewStatementEmptyRejected(t *testing.T) {
	if _, err := NewStatement(""); !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStatement)
	}
}

func TestNewStatementWhitespaceOnlyRejected(t *testing.T) {
	if _, err := NewStatement("   \t\n  "); !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStatement)
	}
}

func TestNewStatementMultilinePreserved(t *testing.T) {
	text := "Line one.\nLine two.\n\nLine four."
	s, err := NewStatement(text)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Text(); got != text {
		t.Errorf("Text() = %q, want %q (multiline internal structure must be preserved)", got, text)
	}
}

func TestNewStatementPunctuationAndWordingPreserved(t *testing.T) {
	text := "The system SHALL NOT allow unauthenticated access; see also RFC-1234."
	s, err := NewStatement(text)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Text(); got != text {
		t.Errorf("Text() = %q, want %q (exact wording/punctuation must be preserved)", got, text)
	}
}

func TestNewStatementDoesNotRequireShall(t *testing.T) {
	if _, err := NewStatement("Audit records are retained for five years."); err != nil {
		t.Errorf("unexpected error for a Statement without the word \"shall\": %v", err)
	}
}

func TestStatementJSONRoundTrip(t *testing.T) {
	original, err := NewStatement("The service shall retain audit records.")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["text"]; !present {
		t.Error("text field missing from Marshal output")
	}
	var decoded Statement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Text() != original.Text() {
		t.Errorf("round trip Text() = %q, want %q", decoded.Text(), original.Text())
	}
}

func TestStatementJSONMalformed(t *testing.T) {
	var s Statement
	if err := json.Unmarshal([]byte(`{"text":""}`), &s); !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("empty text: error = %v, want %v", err, ErrInvalidStatement)
	}
	if err := json.Unmarshal([]byte(`{"text":"   "}`), &s); !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("whitespace-only text: error = %v, want %v", err, ErrInvalidStatement)
	}
}

func TestStatementUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewStatement("pre-existing valid statement")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"text":""}`), &receiver); err == nil {
		t.Fatal("malformed Statement JSON accepted, want error")
	}
	if receiver.Text() != original.Text() {
		t.Errorf("failed Unmarshal changed Text(): got %q, want %q", receiver.Text(), original.Text())
	}
}

func TestStatementZeroValue(t *testing.T) {
	var s Statement
	if !s.IsZero() {
		t.Error("zero-value Statement.IsZero() = false, want true")
	}
}

func TestStatementDuplicatesNotForbidden(t *testing.T) {
	a, err := NewStatement("Duplicate wording")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewStatement("Duplicate wording")
	if err != nil {
		t.Fatal(err)
	}
	subjects := []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}
	if _, err := NewContent([]Statement{a, b}, subjects, SubjectCombinationIndependent, NewUnrestrictedApplicability()); err != nil {
		t.Fatalf("duplicate Statement values within one Content unexpectedly rejected: %v", err)
	}
}
