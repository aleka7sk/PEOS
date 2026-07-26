package decision

import (
	"encoding/json"
	"errors"
	"strings"
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

func TestDecisionAlternativesDefensiveCopies(t *testing.T) {
	alt, err := NewAlternative("Use MySQL")
	if err != nil {
		t.Fatal(err)
	}
	d := baseDecision(t)
	alternatives := []Alternative{alt}
	d, err = d.WithAlternatives(alternatives...)
	if err != nil {
		t.Fatal(err)
	}
	alternatives[0] = Alternative{}
	if d.Alternatives()[0].IsZero() {
		t.Error("WithAlternatives did not defensively copy input")
	}
	got := d.Alternatives()
	got[0] = Alternative{}
	if d.Alternatives()[0].IsZero() {
		t.Error("Alternatives() did not defensively copy on return")
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
	role, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithRoles(role)
	if err != nil {
		t.Fatal(err)
	}
	consequence, err := NewConsequence("Migration of existing services is expected.")
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithConsequences(consequence)
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
	if len(decoded.Roles()) != len(d.Roles()) {
		t.Errorf("Roles mismatch: got %d, want %d", len(decoded.Roles()), len(d.Roles()))
	}
	if len(decoded.Consequences()) != len(d.Consequences()) {
		t.Errorf("Consequences mismatch: got %d, want %d", len(decoded.Consequences()), len(d.Consequences()))
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
	for _, optional := range []string{"subjects", "alternatives", "basis", "rationale", "provenance", "roles", "consequences", "extension"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}
}

func TestDecisionExplicitNullRejectedForEveryOptionalField(t *testing.T) {
	base := `{"id":"dec-1","outcome":{"statement":"s","commitment_effect":"peos:establishes"},"applicability":{"kind":"product-x:path","expression":"/x"},"authority":{"bases":[{"namespace":"role","identifier":"cto"}]}`
	fields := []string{"subjects", "question", "alternatives", "basis", "rationale", "provenance", "roles", "consequences", "extension"}
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
	if len(receiver.Roles()) != len(original.Roles()) {
		t.Error("failed Unmarshal changed receiver's roles")
	}
	if len(receiver.Consequences()) != len(original.Consequences()) {
		t.Error("failed Unmarshal changed receiver's consequences")
	}
}

// --- Roles ---------------------------------------------------------------

func TestDecisionRolesAbsentPresent(t *testing.T) {
	d := baseDecision(t)
	if len(d.Roles()) != 0 {
		t.Error("Roles() non-empty before WithRoles")
	}
	role, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	withRole, err := d.WithRoles(role)
	if err != nil {
		t.Fatal(err)
	}
	if len(withRole.Roles()) != 1 {
		t.Errorf("Roles() len = %d, want 1", len(withRole.Roles()))
	}
	if len(d.Roles()) != 0 {
		t.Error("WithRoles mutated the original receiver")
	}
}

func TestDecisionZeroRoleRejected(t *testing.T) {
	d := baseDecision(t)
	if _, err := d.WithRoles(Role{}); !errors.Is(err, ErrInvalidDecisionRole) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRole)
	}
}

func TestDecisionRolesClearedByEmptyCall(t *testing.T) {
	role, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	d := baseDecision(t)
	d, err = d.WithRoles(role)
	if err != nil {
		t.Fatal(err)
	}
	cleared, err := d.WithRoles()
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Roles()) != 0 {
		t.Error("WithRoles() with no arguments did not clear Roles")
	}
}

func TestDecisionRolesDefensiveCopies(t *testing.T) {
	role, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	d := baseDecision(t)
	roles := []Role{role}
	d, err = d.WithRoles(roles...)
	if err != nil {
		t.Fatal(err)
	}
	roles[0] = Role{}
	if d.Roles()[0].IsZero() {
		t.Error("WithRoles did not defensively copy input")
	}
	got := d.Roles()
	got[0] = Role{}
	if d.Roles()[0].IsZero() {
		t.Error("Roles() did not defensively copy on return")
	}
}

// TestDecisionRolesMultipleRolesSameActor proves PEOS-004 :783: one
// Actor MAY perform multiple roles on the same Decision.
func TestDecisionRolesMultipleRolesSameActor(t *testing.T) {
	actor := mustRoleActor(t, "alice")
	author, err := NewRole(actor, RoleKindAuthor)
	if err != nil {
		t.Fatal(err)
	}
	approver, err := NewRole(actor, RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	d := baseDecision(t)
	d, err = d.WithRoles(author, approver)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Roles()) != 2 {
		t.Errorf("Roles() len = %d, want 2", len(d.Roles()))
	}
}

// --- Consequences ----------------------------------------------------------

func TestDecisionConsequencesAbsentPresent(t *testing.T) {
	d := baseDecision(t)
	if len(d.Consequences()) != 0 {
		t.Error("Consequences() non-empty before WithConsequences")
	}
	consequence, err := NewConsequence("Migration is expected.")
	if err != nil {
		t.Fatal(err)
	}
	withConsequence, err := d.WithConsequences(consequence)
	if err != nil {
		t.Fatal(err)
	}
	if len(withConsequence.Consequences()) != 1 {
		t.Errorf("Consequences() len = %d, want 1", len(withConsequence.Consequences()))
	}
	if len(d.Consequences()) != 0 {
		t.Error("WithConsequences mutated the original receiver")
	}
}

func TestDecisionZeroConsequenceRejected(t *testing.T) {
	d := baseDecision(t)
	if _, err := d.WithConsequences(Consequence{}); !errors.Is(err, ErrInvalidConsequence) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConsequence)
	}
}

func TestDecisionConsequencesClearedByEmptyCall(t *testing.T) {
	consequence, err := NewConsequence("Migration is expected.")
	if err != nil {
		t.Fatal(err)
	}
	d := baseDecision(t)
	d, err = d.WithConsequences(consequence)
	if err != nil {
		t.Fatal(err)
	}
	cleared, err := d.WithConsequences()
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Consequences()) != 0 {
		t.Error("WithConsequences() with no arguments did not clear Consequences")
	}
}

func TestDecisionConsequencesDefensiveCopies(t *testing.T) {
	consequence, err := NewConsequence("Migration is expected.")
	if err != nil {
		t.Fatal(err)
	}
	d := baseDecision(t)
	consequences := []Consequence{consequence}
	d, err = d.WithConsequences(consequences...)
	if err != nil {
		t.Fatal(err)
	}
	consequences[0] = Consequence{}
	if d.Consequences()[0].IsZero() {
		t.Error("WithConsequences did not defensively copy input")
	}
	got := d.Consequences()
	got[0] = Consequence{}
	if d.Consequences()[0].IsZero() {
		t.Error("Consequences() did not defensively copy on return")
	}
}

// --- Absence audit (Packet F.3 architectural deferral / rejections) -------

// TestDecisionNoDelegationSymbolLeaked is a documentation/API audit
// proving Delegation was not accidentally introduced: Decision exposes
// no Delegation-named method, and no Delegation-named JSON key appears
// on a fully populated Decision's wire form.
func TestDecisionNoDelegationSymbolLeaked(t *testing.T) {
	d := fullDecision(t)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for key := range raw {
		if strings.Contains(strings.ToLower(key), "delegat") {
			t.Errorf("unexpected delegation-related key %q in Decision wire form", key)
		}
	}
}

// TestDecisionConsequenceHasNoResolvedOrStatusField is a structural
// absence audit for the Conflict/Resolution "no resolved bool, no
// status field anywhere" architectural rule, exercised at the Decision
// integration boundary where a Consequence is embedded.
func TestDecisionConsequenceHasNoResolvedOrStatusField(t *testing.T) {
	consequence, err := NewConsequence("Migration is expected.")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(consequence)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["resolved"]; present {
		t.Error(`unexpected "resolved" key present in Consequence wire form`)
	}
	if _, present := raw["status"]; present {
		t.Error(`unexpected "status" key present in Consequence wire form`)
	}
}
