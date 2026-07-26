package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- OutcomeKind / CommitmentEffect (vocabulary wrappers) -----------------

func TestOutcomeKindPredeclaredValues(t *testing.T) {
	cases := []struct {
		kind OutcomeKind
		want string
	}{
		{OutcomeKindSelect, "peos:select"},
		{OutcomeKindReject, "peos:reject"},
		{OutcomeKindDefer, "peos:defer"},
		{OutcomeKindAuthorize, "peos:authorize"},
		{OutcomeKindProhibit, "peos:prohibit"},
		{OutcomeKindConstrain, "peos:constrain"},
		{OutcomeKindDelegate, "peos:delegate"},
		{OutcomeKindAcceptRisk, "peos:accept-risk"},
		{OutcomeKindRequireInvestigation, "peos:require-investigation"},
		{OutcomeKindEstablishCommitment, "peos:establish-commitment"},
		{OutcomeKindChangeCommitment, "peos:change-commitment"},
		{OutcomeKindRemoveCommitment, "peos:remove-commitment"},
		{OutcomeKindNoChangeRequired, "peos:no-change-required"},
	}
	for _, c := range cases {
		if c.kind.String() != c.want {
			t.Errorf("String() = %q, want %q", c.kind.String(), c.want)
		}
	}
}

func TestCommitmentEffectPredeclaredValues(t *testing.T) {
	cases := []struct {
		effect CommitmentEffect
		want   string
	}{
		{CommitmentEffectEstablishes, "peos:establishes"},
		{CommitmentEffectChanges, "peos:changes"},
		{CommitmentEffectRemoves, "peos:removes"},
		{CommitmentEffectRejects, "peos:rejects"},
		{CommitmentEffectDefers, "peos:defers"},
		{CommitmentEffectUnchanged, "peos:unchanged"},
	}
	for _, c := range cases {
		if c.effect.String() != c.want {
			t.Errorf("String() = %q, want %q", c.effect.String(), c.want)
		}
	}
}

func TestOutcomeKindProductDefinedValueAccepted(t *testing.T) {
	vocab, err := core.NewVocabularyValue("product-x", "escalate")
	if err != nil {
		t.Fatal(err)
	}
	k := NewOutcomeKind(vocab)
	if k.String() != "product-x:escalate" {
		t.Errorf("String() = %q, want %q", k.String(), "product-x:escalate")
	}
}

func TestOutcomeKindZeroBehavior(t *testing.T) {
	var k OutcomeKind
	if !k.IsZero() {
		t.Error("zero OutcomeKind IsZero() = false")
	}
}

func TestOutcomeKindEqual(t *testing.T) {
	if !OutcomeKindSelect.Equal(OutcomeKindSelect) {
		t.Error("Equal(self) = false")
	}
	if OutcomeKindSelect.Equal(OutcomeKindReject) {
		t.Error("Equal(different) = true")
	}
	if OutcomeKindSelect.Value().String() != "peos:select" {
		t.Errorf("Value() = %v, want peos:select", OutcomeKindSelect.Value())
	}
}

func TestCommitmentEffectEqual(t *testing.T) {
	if !CommitmentEffectEstablishes.Equal(CommitmentEffectEstablishes) {
		t.Error("Equal(self) = false")
	}
	if CommitmentEffectEstablishes.Equal(CommitmentEffectRemoves) {
		t.Error("Equal(different) = true")
	}
	if CommitmentEffectEstablishes.Value().String() != "peos:establishes" {
		t.Errorf("Value() = %v, want peos:establishes", CommitmentEffectEstablishes.Value())
	}
}

func TestOutcomeKindJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(OutcomeKindSelect)
	if err != nil {
		t.Fatal(err)
	}
	var decoded OutcomeKind
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(OutcomeKindSelect) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, OutcomeKindSelect)
	}
}

func TestOutcomeKindMalformedJSONRejected(t *testing.T) {
	var k OutcomeKind
	if err := json.Unmarshal([]byte(`"missing-colon"`), &k); err == nil {
		t.Error("malformed vocabulary string accepted, want error")
	}
}

func TestOutcomeKindZeroMarshalRejected(t *testing.T) {
	var k OutcomeKind
	if _, err := json.Marshal(k); !errors.Is(err, ErrInvalidOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOutcome)
	}
}

func TestCommitmentEffectZeroMarshalRejected(t *testing.T) {
	var e CommitmentEffect
	if _, err := json.Marshal(e); !errors.Is(err, ErrInvalidCommitmentEffect) {
		t.Errorf("error = %v, want %v", err, ErrInvalidCommitmentEffect)
	}
}

func TestOutcomeKindUnmarshalFailurePreservesReceiver(t *testing.T) {
	receiver := OutcomeKindSelect
	if err := json.Unmarshal([]byte(`"missing-colon"`), &receiver); err == nil {
		t.Fatal("malformed value accepted, want error")
	}
	if !receiver.Equal(OutcomeKindSelect) {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- Commitment ------------------------------------------------------------

func TestNewCommitmentValid(t *testing.T) {
	c, err := NewCommitment("All new services must use structured logging")
	if err != nil {
		t.Fatal(err)
	}
	if c.Statement() == "" {
		t.Error("Statement() empty")
	}
}

func TestNewCommitmentEmptyStatementRejected(t *testing.T) {
	if _, err := NewCommitment(""); !errors.Is(err, ErrInvalidCommitment) {
		t.Errorf("error = %v, want %v", err, ErrInvalidCommitment)
	}
}

func TestCommitmentExtension(t *testing.T) {
	c, err := NewCommitment("statement")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := c.WithExtension(ext)
	if !c.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestCommitmentJSON(t *testing.T) {
	c, err := NewCommitment("statement")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Commitment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != c.Statement() {
		t.Errorf("round trip mismatch: got %q, want %q", decoded.Statement(), c.Statement())
	}
}

func TestCommitmentZeroMarshalRejected(t *testing.T) {
	var c Commitment
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidCommitment) {
		t.Errorf("error = %v, want %v", err, ErrInvalidCommitment)
	}
}

func TestCommitmentUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewCommitment("statement")
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
}

// --- Outcome ---------------------------------------------------------------

func TestNewOutcomeValidMinimum(t *testing.T) {
	o, err := NewOutcome("PostgreSQL is selected.", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	if o.Statement() == "" {
		t.Error("Statement() empty")
	}
	if _, ok := o.Kind(); ok {
		t.Error("Kind() ok = true before WithKind")
	}
}

func TestNewOutcomeEmptyStatementRejected(t *testing.T) {
	if _, err := NewOutcome("", CommitmentEffectEstablishes); !errors.Is(err, ErrInvalidOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOutcome)
	}
}

func TestNewOutcomeZeroCommitmentEffectRejected(t *testing.T) {
	if _, err := NewOutcome("statement", CommitmentEffect{}); !errors.Is(err, ErrInvalidCommitmentEffect) {
		t.Errorf("error = %v, want %v", err, ErrInvalidCommitmentEffect)
	}
}

func TestOutcomeWithKindValid(t *testing.T) {
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	withKind, err := o.WithKind(OutcomeKindSelect)
	if err != nil {
		t.Fatal(err)
	}
	kind, ok := withKind.Kind()
	if !ok || !kind.Equal(OutcomeKindSelect) {
		t.Errorf("Kind() = (%v,%v), want (%v,true)", kind, ok, OutcomeKindSelect)
	}
	if _, ok := o.Kind(); ok {
		t.Error("WithKind mutated the original receiver")
	}
}

func TestOutcomeWithKindZeroRejected(t *testing.T) {
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := o.WithKind(OutcomeKind{}); !errors.Is(err, ErrInvalidOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOutcome)
	}
}

func TestOutcomeWithoutKind(t *testing.T) {
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithKind(OutcomeKindSelect)
	if err != nil {
		t.Fatal(err)
	}
	cleared := o.WithoutKind()
	if _, ok := cleared.Kind(); ok {
		t.Error("Kind() ok = true after WithoutKind")
	}
}

func TestOutcomeMultipleCommitments(t *testing.T) {
	c1, err := NewCommitment("commitment 1")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := NewCommitment("commitment 2")
	if err != nil {
		t.Fatal(err)
	}
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	withCommitments, err := o.WithCommitments(c1, c2)
	if err != nil {
		t.Fatal(err)
	}
	if len(withCommitments.Commitments()) != 2 {
		t.Errorf("Commitments() len = %d, want 2", len(withCommitments.Commitments()))
	}
	if len(o.Commitments()) != 0 {
		t.Error("WithCommitments mutated the original receiver")
	}
}

func TestOutcomeZeroCommitmentRejected(t *testing.T) {
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := o.WithCommitments(Commitment{}); !errors.Is(err, ErrInvalidCommitment) {
		t.Errorf("error = %v, want %v", err, ErrInvalidCommitment)
	}
}

func TestOutcomeDefensiveCopies(t *testing.T) {
	c1, err := NewCommitment("commitment 1")
	if err != nil {
		t.Fatal(err)
	}
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	commitments := []Commitment{c1}
	o, err = o.WithCommitments(commitments...)
	if err != nil {
		t.Fatal(err)
	}
	commitments[0] = Commitment{}
	if o.Commitments()[0].IsZero() {
		t.Error("WithCommitments did not defensively copy input")
	}
	got := o.Commitments()
	got[0] = Commitment{}
	if o.Commitments()[0].IsZero() {
		t.Error("Commitments() did not defensively copy on return")
	}
}

func TestOutcomeExtension(t *testing.T) {
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := o.WithExtension(ext)
	if !o.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestOutcomeProductDefinedKindAndEffectRoundTrip(t *testing.T) {
	kindVocab, err := core.NewVocabularyValue("product-x", "escalate")
	if err != nil {
		t.Fatal(err)
	}
	effectVocab, err := core.NewVocabularyValue("product-x", "modifies")
	if err != nil {
		t.Fatal(err)
	}
	o, err := NewOutcome("statement", NewCommitmentEffect(effectVocab))
	if err != nil {
		t.Fatal(err)
	}
	o, err = o.WithKind(NewOutcomeKind(kindVocab))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Outcome
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	kind, ok := decoded.Kind()
	if !ok || kind.String() != "product-x:escalate" {
		t.Errorf("Kind() = (%v,%v), want (product-x:escalate,true)", kind, ok)
	}
	if decoded.CommitmentEffect().String() != "product-x:modifies" {
		t.Errorf("CommitmentEffect() = %v, want product-x:modifies", decoded.CommitmentEffect())
	}
}

func TestOutcomeJSONOptionalKeyOmission(t *testing.T) {
	o, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, optional := range []string{"kind", "commitments", "extension"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}
}

func TestOutcomeExplicitNullRejected(t *testing.T) {
	var o Outcome
	base := `{"statement":"s","commitment_effect":"peos:establishes",`
	if err := json.Unmarshal([]byte(base+`"kind":null}`), &o); !errors.Is(err, ErrInvalidOutcome) {
		t.Errorf("null kind: error = %v, want %v", err, ErrInvalidOutcome)
	}
	if err := json.Unmarshal([]byte(base+`"commitments":null}`), &o); !errors.Is(err, ErrInvalidOutcome) {
		t.Errorf("null commitments: error = %v, want %v", err, ErrInvalidOutcome)
	}
	if err := json.Unmarshal([]byte(base+`"extension":null}`), &o); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestOutcomeZeroMarshalRejected(t *testing.T) {
	var o Outcome
	if _, err := json.Marshal(o); !errors.Is(err, ErrInvalidOutcome) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOutcome)
	}
}

func TestOutcomeUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewOutcome("statement", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithKind(OutcomeKindSelect)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"statement":"","commitment_effect":"peos:establishes"}`), &receiver); err == nil {
		t.Fatal("empty statement accepted, want error")
	}
	if receiver.Statement() != original.Statement() {
		t.Error("failed Unmarshal changed receiver")
	}
	gotKind, gotOK := receiver.Kind()
	wantKind, wantOK := original.Kind()
	if gotOK != wantOK || !gotKind.Equal(wantKind) {
		t.Error("failed Unmarshal changed receiver's kind")
	}
}
