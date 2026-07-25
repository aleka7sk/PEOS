package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustScope(t *testing.T, kind, expression string) core.Scope {
	t.Helper()
	s, err := core.NewScope(mustVocab(t, kind, "condition"), expression)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// --- Applicability ---------------------------------------------------------

func TestApplicabilityUnrestricted(t *testing.T) {
	a := NewUnrestrictedApplicability()
	if a.IsZero() {
		t.Error("unrestricted Applicability reports IsZero() = true")
	}
	if !a.IsUnrestricted() {
		t.Error("IsUnrestricted() = false, want true")
	}
	if _, ok := a.Scope(); ok {
		t.Error("Scope() ok = true for unrestricted Applicability")
	}
}

func TestApplicabilityScoped(t *testing.T) {
	scope := mustScope(t, "product-x", "deployment=eu")
	a, err := NewApplicabilityFromScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	if a.IsUnrestricted() {
		t.Error("IsUnrestricted() = true for scoped Applicability")
	}
	got, ok := a.Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}
}

func TestApplicabilityZeroScopeRejected(t *testing.T) {
	if _, err := NewApplicabilityFromScope(core.Scope{}); !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicability)
	}
}

func TestApplicabilityJSONRoundTripUnrestricted(t *testing.T) {
	original := NewUnrestrictedApplicability()
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"kind":"unrestricted"}`; got != want {
		t.Errorf("Marshal output = %s, want %s", got, want)
	}
	var decoded Applicability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.IsUnrestricted() {
		t.Error("round trip IsUnrestricted() = false, want true")
	}
}

func TestApplicabilityJSONRoundTripScoped(t *testing.T) {
	scope := mustScope(t, "product-x", "deployment=eu")
	original, err := NewApplicabilityFromScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Applicability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, ok := decoded.Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("round trip Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}
}

func TestApplicabilityJSONUnknownKindRejected(t *testing.T) {
	var a Applicability
	if err := json.Unmarshal([]byte(`{"kind":"bogus"}`), &a); !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicability)
	}
}

func TestApplicabilityJSONMissingKindRejected(t *testing.T) {
	var a Applicability
	if err := json.Unmarshal([]byte(`{}`), &a); !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicability)
	}
}

func TestApplicabilityJSONScopedWithoutScopeRejected(t *testing.T) {
	var a Applicability
	if err := json.Unmarshal([]byte(`{"kind":"scoped"}`), &a); !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicability)
	}
}

func TestApplicabilityJSONUnrestrictedWithScopeRejected(t *testing.T) {
	var a Applicability
	payload := `{"kind":"unrestricted","scope":{"kind":"product-x:condition","expression":"x"}}`
	if err := json.Unmarshal([]byte(payload), &a); !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicability)
	}
}

func TestApplicabilityUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := NewUnrestrictedApplicability()
	receiver := original
	if err := json.Unmarshal([]byte(`{"kind":"bogus"}`), &receiver); err == nil {
		t.Fatal("malformed Applicability JSON accepted, want error")
	}
	if receiver.IsUnrestricted() != original.IsUnrestricted() {
		t.Errorf("failed Unmarshal changed IsUnrestricted(): got %v, want %v", receiver.IsUnrestricted(), original.IsUnrestricted())
	}
}

func TestApplicabilityZeroValue(t *testing.T) {
	var a Applicability
	if !a.IsZero() {
		t.Error("zero-value Applicability.IsZero() = false, want true")
	}
}

func TestApplicabilityZeroValueMarshalFails(t *testing.T) {
	var a Applicability
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("Marshal(zero Applicability): error = %v, want %v", err, ErrInvalidApplicability)
	}
}

// --- OriginRef ---------------------------------------------------------

func TestNewOriginRefValid(t *testing.T) {
	kind := mustVocab(t, "peos", "regulatory")
	o, err := NewOriginRef(kind, "Required by applicable regulation.")
	if err != nil {
		t.Fatal(err)
	}
	if o.IsZero() {
		t.Error("valid OriginRef reports IsZero() = true")
	}
	if o.Kind() != kind {
		t.Errorf("Kind() = %v, want %v", o.Kind(), kind)
	}
	if got, want := o.Note(), "Required by applicable regulation."; got != want {
		t.Errorf("Note() = %q, want %q", got, want)
	}
}

func TestNewOriginRefZeroKindRejected(t *testing.T) {
	if _, err := NewOriginRef(core.VocabularyValue{}, "a note"); !errors.Is(err, ErrInvalidOrigin) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOrigin)
	}
}

func TestNewOriginRefEmptyNoteRejected(t *testing.T) {
	kind := mustVocab(t, "peos", "regulatory")
	if _, err := NewOriginRef(kind, ""); !errors.Is(err, ErrInvalidOrigin) {
		t.Errorf("empty note: error = %v, want %v", err, ErrInvalidOrigin)
	}
	if _, err := NewOriginRef(kind, "   "); !errors.Is(err, ErrInvalidOrigin) {
		t.Errorf("whitespace-only note: error = %v, want %v", err, ErrInvalidOrigin)
	}
}

func TestNewOriginRefCustomVocabularyAccepted(t *testing.T) {
	kind := mustVocab(t, "product-x", "customer-need")
	o, err := NewOriginRef(kind, "Requested by a key customer.")
	if err != nil {
		t.Fatal(err)
	}
	if o.Kind() != kind {
		t.Errorf("Kind() = %v, want %v", o.Kind(), kind)
	}
}

func TestOriginRefJSONRoundTrip(t *testing.T) {
	original, err := NewOriginRef(mustVocab(t, "peos", "regulatory"), "Required by applicable regulation.")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded OriginRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Kind() != original.Kind() || decoded.Note() != original.Note() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestOriginRefUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewOriginRef(mustVocab(t, "peos", "regulatory"), "pre-existing note")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"kind":"","note":""}`), &receiver); err == nil {
		t.Fatal("malformed OriginRef JSON accepted, want error")
	}
	if receiver.Kind() != original.Kind() || receiver.Note() != original.Note() {
		t.Errorf("failed Unmarshal changed receiver: got %+v, want %+v", receiver, original)
	}
}

func TestOriginRefZeroValue(t *testing.T) {
	var o OriginRef
	if !o.IsZero() {
		t.Error("zero-value OriginRef.IsZero() = false, want true")
	}
}

func TestOriginRefZeroValueMarshalFails(t *testing.T) {
	var o OriginRef
	if _, err := json.Marshal(o); !errors.Is(err, ErrInvalidOrigin) {
		t.Errorf("Marshal(zero OriginRef): error = %v, want %v", err, ErrInvalidOrigin)
	}
}

// --- Classification ---------------------------------------------------------

func TestNewClassificationValid(t *testing.T) {
	v := mustVocab(t, "product-x", "security")
	c, err := NewClassification(v)
	if err != nil {
		t.Fatal(err)
	}
	if c.IsZero() {
		t.Error("valid Classification reports IsZero() = true")
	}
	if c.Value() != v {
		t.Errorf("Value() = %v, want %v", c.Value(), v)
	}
}

func TestNewClassificationZeroRejected(t *testing.T) {
	if _, err := NewClassification(core.VocabularyValue{}); !errors.Is(err, ErrInvalidClassification) {
		t.Errorf("error = %v, want %v", err, ErrInvalidClassification)
	}
}

func TestClassificationJSONRoundTrip(t *testing.T) {
	original, err := NewClassification(mustVocab(t, "product-x", "security"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `"product-x:security"`; got != want {
		t.Errorf("Marshal output = %s, want %s", got, want)
	}
	var decoded Classification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Value() != original.Value() {
		t.Errorf("round trip Value() = %v, want %v", decoded.Value(), original.Value())
	}
}

func TestClassificationUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewClassification(mustVocab(t, "product-x", "security"))
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`123`), &receiver); err == nil {
		t.Fatal("malformed Classification JSON accepted, want error")
	}
	if receiver.Value() != original.Value() {
		t.Errorf("failed Unmarshal changed Value(): got %v, want %v", receiver.Value(), original.Value())
	}
}

func TestClassificationZeroValue(t *testing.T) {
	var c Classification
	if !c.IsZero() {
		t.Error("zero-value Classification.IsZero() = false, want true")
	}
}

func TestClassificationZeroValueMarshalFails(t *testing.T) {
	var c Classification
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidClassification) {
		t.Errorf("Marshal(zero Classification): error = %v, want %v", err, ErrInvalidClassification)
	}
}

// --- Rationale ---------------------------------------------------------

func TestNewRationaleValid(t *testing.T) {
	r, err := NewRationale("Audit records are required for investigations.")
	if err != nil {
		t.Fatal(err)
	}
	if r.IsZero() {
		t.Error("valid Rationale reports IsZero() = true")
	}
	if got, want := r.Text(), "Audit records are required for investigations."; got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}

func TestNewRationaleEmptyRejected(t *testing.T) {
	if _, err := NewRationale(""); !errors.Is(err, ErrInvalidRationale) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRationale)
	}
	if _, err := NewRationale("   \t "); !errors.Is(err, ErrInvalidRationale) {
		t.Errorf("whitespace-only: error = %v, want %v", err, ErrInvalidRationale)
	}
}

func TestNewRationaleMultilinePreserved(t *testing.T) {
	text := "First reason.\nSecond reason."
	r, err := NewRationale(text)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Text(); got != text {
		t.Errorf("Text() = %q, want %q", got, text)
	}
}

func TestRationaleJSONRoundTrip(t *testing.T) {
	original, err := NewRationale("Audit records are required for investigations.")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Rationale
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Text() != original.Text() {
		t.Errorf("round trip Text() = %q, want %q", decoded.Text(), original.Text())
	}
}

func TestRationaleUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewRationale("pre-existing rationale")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"text":""}`), &receiver); err == nil {
		t.Fatal("malformed Rationale JSON accepted, want error")
	}
	if receiver.Text() != original.Text() {
		t.Errorf("failed Unmarshal changed Text(): got %q, want %q", receiver.Text(), original.Text())
	}
}

func TestRationaleZeroValue(t *testing.T) {
	var r Rationale
	if !r.IsZero() {
		t.Error("zero-value Rationale.IsZero() = false, want true")
	}
}

func TestRationaleZeroValueMarshalFails(t *testing.T) {
	var r Rationale
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRationale) {
		t.Errorf("Marshal(zero Rationale): error = %v, want %v", err, ErrInvalidRationale)
	}
}

// --- Content ---------------------------------------------------------

func TestNewContentValidMinimum(t *testing.T) {
	c := mustContent(t)
	if c.IsZero() {
		t.Error("valid Content reports IsZero() = true")
	}
	if len(c.Statements()) != 1 {
		t.Errorf("Statements() len = %d, want 1", len(c.Statements()))
	}
	if len(c.Subjects()) != 1 {
		t.Errorf("Subjects() len = %d, want 1", len(c.Subjects()))
	}
	if c.Origins() != nil || c.Authorities() != nil || c.Classifications() != nil {
		t.Error("optional collections not nil on minimum Content")
	}
	if !c.Rationale().IsZero() {
		t.Error("Rationale() not zero on minimum Content")
	}
}

func TestContentSubjectCombinationAndApplicabilityAccessors(t *testing.T) {
	scope := mustScope(t, "product-x", "deployment=eu")
	applicability, err := NewApplicabilityFromScope(scope)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewContent(
		[]Statement{mustStatement(t, "text")},
		[]core.EngineeringSubjectRef{mustSubject(t, "ART-1")},
		SubjectCombinationCollective,
		applicability,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.SubjectCombination(); got.Value() != SubjectCombinationCollective.Value() {
		t.Errorf("SubjectCombination() = %v, want %v", got.Value(), SubjectCombinationCollective.Value())
	}
	got, ok := c.Applicability().Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("Applicability().Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}
}

func TestNewContentMultipleStatementsOrderPreserved(t *testing.T) {
	s1 := mustStatement(t, "first")
	s2 := mustStatement(t, "second")
	c, err := NewContent([]Statement{s1, s2}, []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}, SubjectCombinationIndependent, NewUnrestrictedApplicability())
	if err != nil {
		t.Fatal(err)
	}
	got := c.Statements()
	if len(got) != 2 || got[0].Text() != "first" || got[1].Text() != "second" {
		t.Errorf("Statements() = %v, want [first second] (order preserved)", got)
	}
}

func TestNewContentMultipleSubjectsOrderPreserved(t *testing.T) {
	sub1 := mustSubject(t, "ART-1")
	sub2 := mustSubject(t, "ART-2")
	c, err := NewContent([]Statement{mustStatement(t, "text")}, []core.EngineeringSubjectRef{sub1, sub2}, SubjectCombinationCollective, NewUnrestrictedApplicability())
	if err != nil {
		t.Fatal(err)
	}
	got := c.Subjects()
	if len(got) != 2 || got[0] != sub1 || got[1] != sub2 {
		t.Errorf("Subjects() order not preserved: got %v, want [%v %v]", got, sub1, sub2)
	}
}

func TestNewContentEmptyStatementsRejected(t *testing.T) {
	_, err := NewContent(nil, []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}, SubjectCombinationIndependent, NewUnrestrictedApplicability())
	if !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStatement)
	}
}

func TestNewContentZeroStatementRejected(t *testing.T) {
	_, err := NewContent([]Statement{{}}, []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}, SubjectCombinationIndependent, NewUnrestrictedApplicability())
	if !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStatement)
	}
}

func TestNewContentEmptySubjectsRejected(t *testing.T) {
	_, err := NewContent([]Statement{mustStatement(t, "text")}, nil, SubjectCombinationIndependent, NewUnrestrictedApplicability())
	if !errors.Is(err, ErrMissingRequirementSubject) {
		t.Errorf("error = %v, want %v", err, ErrMissingRequirementSubject)
	}
}

func TestNewContentZeroSubjectRejected(t *testing.T) {
	_, err := NewContent([]Statement{mustStatement(t, "text")}, []core.EngineeringSubjectRef{{}}, SubjectCombinationIndependent, NewUnrestrictedApplicability())
	if !errors.Is(err, ErrMissingRequirementSubject) {
		t.Errorf("error = %v, want %v", err, ErrMissingRequirementSubject)
	}
}

func TestNewContentZeroSubjectCombinationRejected(t *testing.T) {
	_, err := NewContent([]Statement{mustStatement(t, "text")}, []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}, SubjectCombination{}, NewUnrestrictedApplicability())
	if !errors.Is(err, ErrInvalidSubjectCombination) {
		t.Errorf("error = %v, want %v", err, ErrInvalidSubjectCombination)
	}
}

func TestNewContentZeroApplicabilityRejected(t *testing.T) {
	_, err := NewContent([]Statement{mustStatement(t, "text")}, []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}, SubjectCombinationIndependent, Applicability{})
	if !errors.Is(err, ErrInvalidApplicability) {
		t.Errorf("error = %v, want %v", err, ErrInvalidApplicability)
	}
}

func TestContentWithOriginsReplaceSemantics(t *testing.T) {
	base := mustContent(t)
	originA, err := NewOriginRef(mustVocab(t, "peos", "regulatory"), "note A")
	if err != nil {
		t.Fatal(err)
	}
	originB, err := NewOriginRef(mustVocab(t, "peos", "customer-need"), "note B")
	if err != nil {
		t.Fatal(err)
	}

	c1, err := base.WithOrigins(originA)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := c1.WithOrigins(originB)
	if err != nil {
		t.Fatal(err)
	}

	if got := c2.Origins(); len(got) != 1 || got[0] != originB {
		t.Errorf("c2.Origins() = %v, want [%v] (second WithOrigins must replace, not accumulate)", got, originB)
	}
	if got := c1.Origins(); len(got) != 1 || got[0] != originA {
		t.Errorf("c1.Origins() = %v, want [%v] (c1 must remain unaffected)", got, originA)
	}
	if got := base.Origins(); got != nil {
		t.Errorf("base.Origins() = %v, want nil (base must remain unaffected)", got)
	}
}

func TestContentWithOriginsEmptyClears(t *testing.T) {
	originA, err := NewOriginRef(mustVocab(t, "peos", "regulatory"), "note A")
	if err != nil {
		t.Fatal(err)
	}
	c1, err := mustContent(t).WithOrigins(originA)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := c1.WithOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if got := c2.Origins(); got != nil {
		t.Errorf("Origins() after no-args WithOrigins() = %v, want nil", got)
	}
	if got := c1.Origins(); len(got) != 1 {
		t.Errorf("c1.Origins() len = %d, want 1 (c1 must remain unaffected by c2's clearing call)", len(got))
	}
}

func TestContentWithOriginsZeroRejected(t *testing.T) {
	if _, err := mustContent(t).WithOrigins(OriginRef{}); !errors.Is(err, ErrInvalidOrigin) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOrigin)
	}
}

func TestContentWithAuthoritiesReplaceSemantics(t *testing.T) {
	base := mustContent(t)
	authA, err := core.NewAuthorityRef("product-x", "compliance-team")
	if err != nil {
		t.Fatal(err)
	}
	authB, err := core.NewAuthorityRef("product-x", "safety-board")
	if err != nil {
		t.Fatal(err)
	}

	c1, err := base.WithAuthorities(authA)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := c1.WithAuthorities(authB)
	if err != nil {
		t.Fatal(err)
	}

	if got := c2.Authorities(); len(got) != 1 || got[0] != authB {
		t.Errorf("c2.Authorities() = %v, want [%v]", got, authB)
	}
	if got := c1.Authorities(); len(got) != 1 || got[0] != authA {
		t.Errorf("c1.Authorities() = %v, want [%v] (c1 must remain unaffected)", got, authA)
	}
}

func TestContentWithAuthoritiesEmptyClears(t *testing.T) {
	authA, err := core.NewAuthorityRef("product-x", "compliance-team")
	if err != nil {
		t.Fatal(err)
	}
	c1, err := mustContent(t).WithAuthorities(authA)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := c1.WithAuthorities()
	if err != nil {
		t.Fatal(err)
	}
	if got := c2.Authorities(); got != nil {
		t.Errorf("Authorities() after no-args WithAuthorities() = %v, want nil", got)
	}
}

func TestContentWithAuthoritiesZeroRejected(t *testing.T) {
	if _, err := mustContent(t).WithAuthorities(core.AuthorityRef{}); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func TestContentWithClassificationsReplaceSemantics(t *testing.T) {
	base := mustContent(t)
	clsA, err := NewClassification(mustVocab(t, "product-x", "security"))
	if err != nil {
		t.Fatal(err)
	}
	clsB, err := NewClassification(mustVocab(t, "product-x", "performance"))
	if err != nil {
		t.Fatal(err)
	}

	c1, err := base.WithClassifications(clsA)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := c1.WithClassifications(clsB)
	if err != nil {
		t.Fatal(err)
	}

	if got := c2.Classifications(); len(got) != 1 || got[0] != clsB {
		t.Errorf("c2.Classifications() = %v, want [%v]", got, clsB)
	}
	if got := c1.Classifications(); len(got) != 1 || got[0] != clsA {
		t.Errorf("c1.Classifications() = %v, want [%v] (c1 must remain unaffected)", got, clsA)
	}
}

func TestContentWithClassificationsEmptyClears(t *testing.T) {
	clsA, err := NewClassification(mustVocab(t, "product-x", "security"))
	if err != nil {
		t.Fatal(err)
	}
	c1, err := mustContent(t).WithClassifications(clsA)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := c1.WithClassifications()
	if err != nil {
		t.Fatal(err)
	}
	if got := c2.Classifications(); got != nil {
		t.Errorf("Classifications() after no-args WithClassifications() = %v, want nil", got)
	}
}

func TestContentWithClassificationsDuplicateRejected(t *testing.T) {
	cls, err := NewClassification(mustVocab(t, "product-x", "security"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mustContent(t).WithClassifications(cls, cls); !errors.Is(err, ErrDuplicateClassification) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateClassification)
	}
}

func TestContentWithClassificationsZeroRejected(t *testing.T) {
	if _, err := mustContent(t).WithClassifications(Classification{}); !errors.Is(err, ErrInvalidClassification) {
		t.Errorf("error = %v, want %v", err, ErrInvalidClassification)
	}
}

func TestContentWithRationaleReplacement(t *testing.T) {
	base := mustContent(t)
	r1, err := NewRationale("first reason")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewRationale("second reason")
	if err != nil {
		t.Fatal(err)
	}
	c1 := base.WithRationale(r1)
	c2 := c1.WithRationale(r2)

	if got := c2.Rationale().Text(); got != "second reason" {
		t.Errorf("c2.Rationale().Text() = %q, want %q", got, "second reason")
	}
	if got := c1.Rationale().Text(); got != "first reason" {
		t.Errorf("c1.Rationale().Text() = %q, want %q (c1 must remain unaffected)", got, "first reason")
	}
	if !base.Rationale().IsZero() {
		t.Error("base.Rationale() not zero (base must remain unaffected)")
	}

	c3 := c2.WithRationale(Rationale{})
	if !c3.Rationale().IsZero() {
		t.Error("c3.Rationale() not zero after WithRationale(Rationale{})")
	}
}

func TestContentDefensiveCopyInputAndAccessors(t *testing.T) {
	statements := []Statement{mustStatement(t, "text")}
	subjects := []core.EngineeringSubjectRef{mustSubject(t, "ART-1")}
	c, err := NewContent(statements, subjects, SubjectCombinationIndependent, NewUnrestrictedApplicability())
	if err != nil {
		t.Fatal(err)
	}

	// Mutating the caller's input slices after construction must not
	// affect the stored Content.
	statements[0] = mustStatement(t, "tampered")
	subjects[0] = mustSubject(t, "ART-2")
	if got := c.Statements(); got[0].Text() != "text" {
		t.Errorf("input mutation affected stored Statements: got %q", got[0].Text())
	}
	if got := c.Subjects(); got[0] != mustSubject(t, "ART-1") {
		t.Errorf("input mutation affected stored Subjects: got %v", got[0])
	}

	// Mutating an accessor's result must not affect internal state.
	gotStatements := c.Statements()
	gotStatements[0] = mustStatement(t, "tampered again")
	if again := c.Statements(); again[0].Text() != "text" {
		t.Errorf("mutating a Statements() result affected internal state: got %q", again[0].Text())
	}

	gotSubjects := c.Subjects()
	gotSubjects[0] = mustSubject(t, "ART-3")
	if again := c.Subjects(); again[0] != mustSubject(t, "ART-1") {
		t.Errorf("mutating a Subjects() result affected internal state: got %v", again[0])
	}
}

func TestContentJSONMinimumRoundTrip(t *testing.T) {
	original := mustContent(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"statements", "subjects", "subject_combination", "applicability"} {
		if _, present := raw[required]; !present {
			t.Errorf("required field %q missing from Marshal output", required)
		}
	}
	for _, optional := range []string{"origins", "authorities", "classifications", "rationale"} {
		if _, present := raw[optional]; present {
			t.Errorf("optional field %q present despite not being set", optional)
		}
	}

	var decoded Content
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Statements()) != 1 || decoded.Statements()[0].Text() != original.Statements()[0].Text() {
		t.Errorf("round trip Statements() mismatch: got %v", decoded.Statements())
	}
}

func TestContentJSONFullRoundTrip(t *testing.T) {
	origin, err := NewOriginRef(mustVocab(t, "peos", "regulatory"), "Required by applicable regulation.")
	if err != nil {
		t.Fatal(err)
	}
	authority, err := core.NewAuthorityRef("product-x", "compliance-team")
	if err != nil {
		t.Fatal(err)
	}
	classification, err := NewClassification(mustVocab(t, "product-x", "security"))
	if err != nil {
		t.Fatal(err)
	}
	rationale, err := NewRationale("Audit records are required for investigations.")
	if err != nil {
		t.Fatal(err)
	}

	original := mustContent(t)
	original, err = original.WithOrigins(origin)
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithAuthorities(authority)
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithClassifications(classification)
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithRationale(rationale)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Content
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if got := decoded.Origins(); len(got) != 1 || got[0] != origin {
		t.Errorf("round trip Origins() = %v, want [%v]", got, origin)
	}
	if got := decoded.Authorities(); len(got) != 1 || got[0] != authority {
		t.Errorf("round trip Authorities() = %v, want [%v]", got, authority)
	}
	if got := decoded.Classifications(); len(got) != 1 || got[0] != classification {
		t.Errorf("round trip Classifications() = %v, want [%v]", got, classification)
	}
	if got := decoded.Rationale().Text(); got != rationale.Text() {
		t.Errorf("round trip Rationale().Text() = %q, want %q", got, rationale.Text())
	}
}

func TestContentJSONInvalidBypassRejected(t *testing.T) {
	var c Content
	// No statements, no subjects: NewContent must still run during
	// Unmarshal and reject this, even though structurally valid JSON.
	payload := `{"statements":[],"subjects":[],"subject_combination":"peos:independent","applicability":{"kind":"unrestricted"}}`
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidStatement) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStatement)
	}
}

func TestContentUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := mustContent(t)
	receiver := original
	payload := `{"statements":[],"subjects":[],"subject_combination":"peos:independent","applicability":{"kind":"unrestricted"}}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("malformed Content JSON accepted, want error")
	}
	if len(receiver.Statements()) != len(original.Statements()) || receiver.Statements()[0].Text() != original.Statements()[0].Text() {
		t.Errorf("failed Unmarshal changed Statements(): got %v, want %v", receiver.Statements(), original.Statements())
	}
}

func TestContentZeroValue(t *testing.T) {
	var c Content
	if !c.IsZero() {
		t.Error("zero-value Content.IsZero() = false, want true")
	}
}

func TestContentZeroValueMarshalFails(t *testing.T) {
	var c Content
	if _, err := json.Marshal(c); !errors.Is(err, ErrMissingRequirementContent) {
		t.Errorf("Marshal(zero Content): error = %v, want %v", err, ErrMissingRequirementContent)
	}
}

func TestContentJSONDuplicateClassificationRejected(t *testing.T) {
	original := mustContent(t)
	receiver := original

	// Hand-built JSON, not obtained through any constructor or With*
	// call, with an exact duplicate classification value.
	payload := `{
		"statements": [{"text": "The service shall retain audit records."}],
		"subjects": [{"kind": "artifact", "ref": {"artifact_id": "ART-1"}}],
		"subject_combination": "peos:independent",
		"applicability": {"kind": "unrestricted"},
		"classifications": ["product-x:security", "product-x:security"]
	}`
	err := json.Unmarshal([]byte(payload), &receiver)
	if !errors.Is(err, ErrDuplicateClassification) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateClassification)
	}

	// Receiver preservation, checked via accessors (Content holds slices
	// and is not comparable with ==).
	if len(receiver.Statements()) != len(original.Statements()) || receiver.Statements()[0].Text() != original.Statements()[0].Text() {
		t.Errorf("failed Unmarshal changed Statements(): got %v, want %v", receiver.Statements(), original.Statements())
	}
	if len(receiver.Subjects()) != len(original.Subjects()) || receiver.Subjects()[0] != original.Subjects()[0] {
		t.Errorf("failed Unmarshal changed Subjects(): got %v, want %v", receiver.Subjects(), original.Subjects())
	}
	if receiver.SubjectCombination().Value() != original.SubjectCombination().Value() {
		t.Errorf("failed Unmarshal changed SubjectCombination(): got %v, want %v", receiver.SubjectCombination().Value(), original.SubjectCombination().Value())
	}
	if receiver.Applicability().IsUnrestricted() != original.Applicability().IsUnrestricted() {
		t.Errorf("failed Unmarshal changed Applicability(): got %v, want %v", receiver.Applicability(), original.Applicability())
	}
	if got := receiver.Classifications(); got != nil {
		t.Errorf("failed Unmarshal changed Classifications(): got %v, want nil (original had none)", got)
	}
}
