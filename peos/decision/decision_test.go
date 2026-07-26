package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustDecisionID(t *testing.T, value string) core.DecisionID {
	t.Helper()
	id, err := core.NewDecisionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustSubject(t *testing.T, decisionID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewDecisionRef(mustDecisionID(t, decisionID))
	if err != nil {
		t.Fatal(err)
	}
	sub, err := core.EngineeringSubjectRefFromDecision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func mustScope(t *testing.T, expression string) core.Scope {
	t.Helper()
	kind, err := core.NewVocabularyValue("product-x", "path")
	if err != nil {
		t.Fatal(err)
	}
	scope, err := core.NewScope(kind, expression)
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func mustTestAuthority(t *testing.T) Authority {
	t.Helper()
	basis, err := core.NewAuthorityRef("role", "cto")
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewAuthority(nil, []core.AuthorityRef{basis})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustTestOutcome(t *testing.T) Outcome {
	t.Helper()
	o, err := NewOutcome("PostgreSQL is selected.", CommitmentEffectEstablishes)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestNewDecisionSubjectOnlyAccepted(t *testing.T) {
	d, err := New(mustDecisionID(t, "dec-1"), []core.EngineeringSubjectRef{mustSubject(t, "dec-other")}, "", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Subjects()) != 1 {
		t.Errorf("Subjects() len = %d, want 1", len(d.Subjects()))
	}
}

func TestNewDecisionQuestionOnlyAccepted(t *testing.T) {
	d, err := New(mustDecisionID(t, "dec-1"), nil, "Which database should be used?", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	q, ok := d.Question()
	if !ok || q != "Which database should be used?" {
		t.Errorf("Question() = (%q,%v)", q, ok)
	}
}

func TestNewDecisionBothAccepted(t *testing.T) {
	_, err := New(mustDecisionID(t, "dec-1"), []core.EngineeringSubjectRef{mustSubject(t, "dec-other")}, "Which database?", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewDecisionNeitherRejected(t *testing.T) {
	if _, err := New(mustDecisionID(t, "dec-1"), nil, "", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t)); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestNewDecisionZeroIDRejected(t *testing.T) {
	if _, err := New(core.DecisionID{}, nil, "question", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t)); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestNewDecisionZeroSubjectRejected(t *testing.T) {
	if _, err := New(mustDecisionID(t, "dec-1"), []core.EngineeringSubjectRef{{}}, "question", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t)); !errors.Is(err, ErrInvalidDecisionSubject) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSubject)
	}
}

func TestNewDecisionZeroOutcomeRejected(t *testing.T) {
	if _, err := New(mustDecisionID(t, "dec-1"), nil, "question", Outcome{}, mustScope(t, "/services/*"), mustTestAuthority(t)); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestNewDecisionZeroScopeRejected(t *testing.T) {
	if _, err := New(mustDecisionID(t, "dec-1"), nil, "question", mustTestOutcome(t), core.Scope{}, mustTestAuthority(t)); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestNewDecisionZeroAuthorityRejected(t *testing.T) {
	if _, err := New(mustDecisionID(t, "dec-1"), nil, "question", mustTestOutcome(t), mustScope(t, "/services/*"), Authority{}); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func baseDecision(t *testing.T) Decision {
	t.Helper()
	d, err := New(mustDecisionID(t, "dec-1"), nil, "Which database should be used?", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestDecisionAlternativesAbsentPresent(t *testing.T) {
	d := baseDecision(t)
	if len(d.Alternatives()) != 0 {
		t.Error("Alternatives() non-empty before WithAlternatives")
	}
	alt, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	withAlt, err := d.WithAlternatives(alt)
	if err != nil {
		t.Fatal(err)
	}
	if len(withAlt.Alternatives()) != 1 {
		t.Errorf("Alternatives() len = %d, want 1", len(withAlt.Alternatives()))
	}
	if len(d.Alternatives()) != 0 {
		t.Error("WithAlternatives mutated the original receiver")
	}
}

func TestDecisionZeroAlternativeRejected(t *testing.T) {
	d := baseDecision(t)
	if _, err := d.WithAlternatives(Alternative{}); !errors.Is(err, ErrInvalidAlternative) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAlternative)
	}
}

func TestDecisionBasisAbsentPresent(t *testing.T) {
	d := baseDecision(t)
	if _, ok := d.Basis(); ok {
		t.Error("Basis() ok = true before WithBasis")
	}
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	basis, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	withBasis, err := d.WithBasis(basis)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withBasis.Basis(); !ok {
		t.Error("Basis() ok = false after WithBasis")
	}
	cleared := withBasis.WithoutBasis()
	if _, ok := cleared.Basis(); ok {
		t.Error("Basis() ok = true after WithoutBasis")
	}
}

func TestDecisionZeroBasisRejected(t *testing.T) {
	d := baseDecision(t)
	if _, err := d.WithBasis(Basis{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestDecisionRationaleAbsentPresent(t *testing.T) {
	d := baseDecision(t)
	if _, ok := d.Rationale(); ok {
		t.Error("Rationale() ok = true before WithRationale")
	}
	withRationale, err := d.WithRationale("Chose PostgreSQL for JSONB support.")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := withRationale.Rationale()
	if !ok || r == "" {
		t.Errorf("Rationale() = (%q,%v)", r, ok)
	}
	cleared := withRationale.WithoutRationale()
	if _, ok := cleared.Rationale(); ok {
		t.Error("Rationale() ok = true after WithoutRationale")
	}
}

func TestDecisionEmptyRationaleRejected(t *testing.T) {
	d := baseDecision(t)
	if _, err := d.WithRationale(""); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestDecisionProvenanceAbsentPresent(t *testing.T) {
	d := baseDecision(t)
	if _, ok := d.Provenance(); ok {
		t.Error("Provenance() ok = true before WithProvenance")
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	withProv, err := d.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withProv.Provenance(); !ok {
		t.Error("Provenance() ok = false after WithProvenance")
	}
	cleared := withProv.WithoutProvenance()
	if _, ok := cleared.Provenance(); ok {
		t.Error("Provenance() ok = true after WithoutProvenance")
	}
}

func TestDecisionZeroProvenanceRejected(t *testing.T) {
	d := baseDecision(t)
	if _, err := d.WithProvenance(core.Provenance{}); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestDecisionWithMethodsAreImmutable(t *testing.T) {
	d := baseDecision(t)
	original := d
	alt, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.WithAlternatives(alt)
	if err != nil {
		t.Fatal(err)
	}
	if d.ID() != original.ID() || len(d.Alternatives()) != len(original.Alternatives()) {
		t.Error("WithAlternatives mutated d")
	}
}

func TestDecisionSlicesDefensivelyCopied(t *testing.T) {
	sub := mustSubject(t, "dec-other")
	subjects := []core.EngineeringSubjectRef{sub}
	d, err := New(mustDecisionID(t, "dec-1"), subjects, "", mustTestOutcome(t), mustScope(t, "/services/*"), mustTestAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	subjects[0] = core.EngineeringSubjectRef{}
	if d.Subjects()[0].IsZero() {
		t.Error("New did not defensively copy subjects input")
	}
	got := d.Subjects()
	got[0] = core.EngineeringSubjectRef{}
	if d.Subjects()[0].IsZero() {
		t.Error("Subjects() did not defensively copy on return")
	}
}

func TestDecisionCoreAccessors(t *testing.T) {
	outcome := mustTestOutcome(t)
	scope := mustScope(t, "/services/*")
	authority := mustTestAuthority(t)
	d, err := New(mustDecisionID(t, "dec-1"), nil, "question", outcome, scope, authority)
	if err != nil {
		t.Fatal(err)
	}
	if d.Outcome().Statement() != outcome.Statement() {
		t.Errorf("Outcome() = %v, want %v", d.Outcome(), outcome)
	}
	if !d.Applicability().Equal(scope) {
		t.Errorf("Applicability() = %v, want %v", d.Applicability(), scope)
	}
	if len(d.Authority().Bases()) != len(authority.Bases()) {
		t.Errorf("Authority() = %v, want %v", d.Authority(), authority)
	}
}

func TestDecisionRefCorrectness(t *testing.T) {
	d := baseDecision(t)
	ref, err := d.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.DecisionID() != d.ID() {
		t.Errorf("Ref().DecisionID() = %v, want %v", ref.DecisionID(), d.ID())
	}
}

func TestDecisionOutcomeRefCorrectness(t *testing.T) {
	d := baseDecision(t)
	ref, err := d.OutcomeRef()
	if err != nil {
		t.Fatal(err)
	}
	if ref.DecisionID() != d.ID() {
		t.Errorf("OutcomeRef().DecisionID() = %v, want %v", ref.DecisionID(), d.ID())
	}
}

func fullDecision(t *testing.T) Decision {
	t.Helper()
	d := baseDecision(t)
	alt, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithAlternatives(alt)
	if err != nil {
		t.Fatal(err)
	}
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	basis, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithBasis(basis)
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithRationale("Chose PostgreSQL for JSONB support.")
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	d, err = d.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return d.WithExtension(ext)
}

func TestDecisionJSONFullRoundTrip(t *testing.T) {
	d := fullDecision(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Decision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != d.ID() {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID(), d.ID())
	}
	q, _ := decoded.Question()
	wantQ, _ := d.Question()
	if q != wantQ {
		t.Errorf("Question mismatch: got %q, want %q", q, wantQ)
	}
	if len(decoded.Alternatives()) != len(d.Alternatives()) {
		t.Errorf("Alternatives mismatch")
	}
	if _, ok := decoded.Basis(); !ok {
		t.Error("Basis absent after round trip")
	}
	if _, ok := decoded.Rationale(); !ok {
		t.Error("Rationale absent after round trip")
	}
	if _, ok := decoded.Provenance(); !ok {
		t.Error("Provenance absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDecisionJSONMinimumRoundTrip(t *testing.T) {
	d := baseDecision(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Decision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != d.ID() {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID(), d.ID())
	}
}

func TestDecisionJSONOptionalKeysOmitted(t *testing.T) {
	d := baseDecision(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, optional := range []string{"subjects", "alternatives", "basis", "rationale", "provenance", "extension"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}
}

func TestDecisionExplicitNullRejectedForEveryOptionalField(t *testing.T) {
	base := `{"id":"dec-1","outcome":{"statement":"s","commitment_effect":"peos:establishes"},"applicability":{"kind":"product-x:path","expression":"/x"},"authority":{"bases":[{"namespace":"role","identifier":"cto"}]}`
	fields := []string{"subjects", "question", "alternatives", "basis", "rationale", "provenance", "extension"}
	for _, field := range fields {
		payload := base + `,"` + field + `":null}`
		var d Decision
		if err := json.Unmarshal([]byte(payload), &d); err == nil {
			t.Errorf("field %q: explicit null accepted, want error", field)
		}
	}
}

func TestDecisionEmptyOptionalStringRejectedWhenPresent(t *testing.T) {
	base := `{"id":"dec-1","outcome":{"statement":"s","commitment_effect":"peos:establishes"},"applicability":{"kind":"product-x:path","expression":"/x"},"authority":{"bases":[{"namespace":"role","identifier":"cto"}]}`
	var d Decision
	if err := json.Unmarshal([]byte(base+`,"question":""}`), &d); err == nil {
		t.Error("empty question accepted, want error")
	}
	if err := json.Unmarshal([]byte(base+`,"subjects":[{"kind":"decision","ref":{"decision_id":"other"}}],"rationale":""}`), &d); err == nil {
		t.Error("empty rationale accepted, want error")
	}
}

func TestDecisionUnknownFieldIgnored(t *testing.T) {
	base := `{"id":"dec-1","outcome":{"statement":"s","commitment_effect":"peos:establishes"},"applicability":{"kind":"product-x:path","expression":"/x"},"authority":{"bases":[{"namespace":"role","identifier":"cto"}]},"question":"q","unknown_field":123}`
	var d Decision
	if err := json.Unmarshal([]byte(base), &d); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionZeroMarshalRejected(t *testing.T) {
	var d Decision
	if _, err := json.Marshal(d); !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecision)
	}
}

func TestDecisionUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullDecision(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{"id":"dec-1"}`), &receiver); err == nil {
		t.Fatal("missing required fields accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver ID")
	}
	if _, ok := receiver.Basis(); !ok {
		t.Error("failed Unmarshal changed receiver's basis presence")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}
