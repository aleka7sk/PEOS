package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustConflictResolutionID(t *testing.T, value string) ConflictResolutionID {
	t.Helper()
	id, err := NewConflictResolutionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func fullConflictResolution(t *testing.T) ConflictResolution {
	t.Helper()
	r, err := NewConflictResolution(
		mustConflictResolutionID(t, "resolution-1"),
		mustConflictID(t, "conflict-1"),
		ResolutionMechanismSupersession,
		"dec-2 supersedes dec-1 within the overlapping scope",
	)
	if err != nil {
		t.Fatal(err)
	}
	r, err = r.WithResolvingDecision(mustConflictDecisionRef(t, "dec-2"))
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	r, err = r.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return r.WithExtension(ext)
}

// --- ConflictResolutionID ------------------------------------------------

func TestNewConflictResolutionIDValid(t *testing.T) {
	id, err := NewConflictResolutionID("resolution-1")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "resolution-1" {
		t.Errorf("String() = %q, want %q", id.String(), "resolution-1")
	}
}

func TestNewConflictResolutionIDEmptyRejected(t *testing.T) {
	if _, err := NewConflictResolutionID(""); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestNewConflictResolutionIDWhitespaceOnlyRejected(t *testing.T) {
	if _, err := NewConflictResolutionID("   "); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestConflictResolutionIDEqual(t *testing.T) {
	a := mustConflictResolutionID(t, "r-1")
	b := mustConflictResolutionID(t, "r-1")
	c := mustConflictResolutionID(t, "r-2")
	if !a.Equal(b) {
		t.Error("Equal() = false for identical IDs")
	}
	if a.Equal(c) {
		t.Error("Equal() = true for different IDs")
	}
}

func TestConflictResolutionIDIsZero(t *testing.T) {
	var id ConflictResolutionID
	if !id.IsZero() {
		t.Error("zero ConflictResolutionID IsZero() = false")
	}
}

func TestConflictResolutionIDUnmarshalMalformedJSONRejected(t *testing.T) {
	var id ConflictResolutionID
	if err := json.Unmarshal([]byte(`123`), &id); err == nil {
		t.Error("non-string JSON accepted, want error")
	}
}

func TestConflictResolutionIDZeroMarshalRejected(t *testing.T) {
	var id ConflictResolutionID
	if _, err := json.Marshal(id); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

// --- ResolutionMechanism ---------------------------------------------------

func TestResolutionMechanismPredeclaredNonZero(t *testing.T) {
	for _, m := range []ResolutionMechanism{
		ResolutionMechanismPriority, ResolutionMechanismAuthority, ResolutionMechanismSupersession,
		ResolutionMechanismScopeRefinement, ResolutionMechanismDecision, ResolutionMechanismProductContract,
	} {
		if m.IsZero() {
			t.Errorf("predeclared ResolutionMechanism %v is zero", m)
		}
	}
}

// TestResolutionMechanismOpenVocabularyAcceptsProductDefined proves
// ResolutionMechanism is open: a Product-defined value not among the six
// predeclared constants is accepted with no special-cased rejection.
func TestResolutionMechanismOpenVocabularyAcceptsProductDefined(t *testing.T) {
	v, err := core.NewVocabularyValue("product-x", "custom-mechanism")
	if err != nil {
		t.Fatal(err)
	}
	custom := NewResolutionMechanism(v)
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), custom, "resolved via a Product-defined mechanism")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Mechanism().Equal(custom) {
		t.Error("custom mechanism not preserved")
	}
}

func TestResolutionMechanismValueAndString(t *testing.T) {
	if ResolutionMechanismPriority.Value().IsZero() {
		t.Error("Value() returned zero value")
	}
	if ResolutionMechanismPriority.String() != "peos:priority" {
		t.Errorf("String() = %q, want %q", ResolutionMechanismPriority.String(), "peos:priority")
	}
}

func TestResolutionMechanismZeroMarshalRejected(t *testing.T) {
	var m ResolutionMechanism
	if _, err := json.Marshal(m); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

// --- NewConflictResolution -------------------------------------------------

func TestNewConflictResolutionValid(t *testing.T) {
	r := fullConflictResolution(t)
	if r.IsZero() {
		t.Error("valid ConflictResolution IsZero() = true")
	}
}

func TestNewConflictResolutionZeroIDRejected(t *testing.T) {
	_, err := NewConflictResolution(ConflictResolutionID{}, mustConflictID(t, "c-1"), ResolutionMechanismPriority, "s")
	if !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestNewConflictResolutionZeroConflictRejected(t *testing.T) {
	_, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), ConflictID{}, ResolutionMechanismPriority, "s")
	if !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestNewConflictResolutionZeroMechanismRejected(t *testing.T) {
	_, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanism{}, "s")
	if !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestNewConflictResolutionEmptyStatementRejected(t *testing.T) {
	_, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "")
	if !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestNewConflictResolutionWhitespaceStatementRejected(t *testing.T) {
	_, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "   ")
	if !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestNewConflictResolutionRawStatementPreserved(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "  padded  ")
	if err != nil {
		t.Fatal(err)
	}
	if r.Statement() != "  padded  " {
		t.Errorf("Statement() = %q, want raw preserved", r.Statement())
	}
}

// TestNewConflictResolutionValidWithEveryMechanismNoResolvingDecision
// proves resolvingDecision is a plain optional field: every mechanism,
// including Decision and Supersession, constructs successfully with no
// resolving Decision set.
func TestNewConflictResolutionValidWithEveryMechanismNoResolvingDecision(t *testing.T) {
	for _, m := range []ResolutionMechanism{
		ResolutionMechanismPriority, ResolutionMechanismAuthority, ResolutionMechanismSupersession,
		ResolutionMechanismScopeRefinement, ResolutionMechanismDecision, ResolutionMechanismProductContract,
	} {
		r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), m, "resolved")
		if err != nil {
			t.Fatalf("mechanism %v: %v", m, err)
		}
		if _, ok := r.ResolvingDecision(); ok {
			t.Errorf("mechanism %v: ResolvingDecision() ok = true, want false", m)
		}
	}
}

// --- With* / accessors ------------------------------------------------------

func TestConflictResolutionResolvingDecisionAbsentPresent(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismDecision, "s")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.ResolvingDecision(); ok {
		t.Error("ResolvingDecision() ok = true before WithResolvingDecision")
	}
	withDec, err := r.WithResolvingDecision(mustConflictDecisionRef(t, "dec-3"))
	if err != nil {
		t.Fatal(err)
	}
	dec, ok := withDec.ResolvingDecision()
	if !ok || dec.IsZero() {
		t.Errorf("ResolvingDecision() = (%v,%v)", dec, ok)
	}
	if _, ok := r.ResolvingDecision(); ok {
		t.Error("WithResolvingDecision mutated the original receiver")
	}
}

func TestConflictResolutionZeroResolvingDecisionRejected(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismDecision, "s")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.WithResolvingDecision(core.DecisionRef{}); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestConflictResolutionWithoutResolvingDecision(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismDecision, "s")
	if err != nil {
		t.Fatal(err)
	}
	r, err = r.WithResolvingDecision(mustConflictDecisionRef(t, "dec-3"))
	if err != nil {
		t.Fatal(err)
	}
	cleared := r.WithoutResolvingDecision()
	if _, ok := cleared.ResolvingDecision(); ok {
		t.Error("ResolvingDecision() ok = true after WithoutResolvingDecision")
	}
	if _, ok := r.ResolvingDecision(); !ok {
		t.Error("WithoutResolvingDecision mutated the original receiver")
	}
}

func TestConflictResolutionWithProvenance(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "s")
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	withProv, err := r.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withProv.Provenance(); !ok {
		t.Error("Provenance() ok = false after WithProvenance")
	}
	if _, ok := r.Provenance(); ok {
		t.Error("WithProvenance mutated the original receiver")
	}
}

func TestConflictResolutionZeroProvenanceRejected(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "s")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.WithProvenance(core.Provenance{}); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestConflictResolutionWithoutProvenance(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "s")
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	r, err = r.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	cleared := r.WithoutProvenance()
	if _, ok := cleared.Provenance(); ok {
		t.Error("Provenance() ok = true after WithoutProvenance")
	}
}

func TestConflictResolutionWithExtension(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "s")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := r.WithExtension(ext)
	if !r.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestConflictResolutionIDAccessor(t *testing.T) {
	r := fullConflictResolution(t)
	if r.ID().IsZero() {
		t.Error("ID() returned zero value")
	}
}

func TestConflictResolutionIsZero(t *testing.T) {
	var r ConflictResolution
	if !r.IsZero() {
		t.Error("zero ConflictResolution IsZero() = false")
	}
	if fullConflictResolution(t).IsZero() {
		t.Error("valid ConflictResolution IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestConflictResolutionJSONLiteralWireKeys(t *testing.T) {
	r := fullConflictResolution(t)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "conflict", "mechanism", "statement", "resolving_decision", "provenance", "extension"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
}

func TestConflictResolutionJSONMinimumOmitsOptionalFields(t *testing.T) {
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), mustConflictID(t, "c-1"), ResolutionMechanismPriority, "s")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"resolving_decision", "provenance", "extension"} {
		if _, present := raw[key]; present {
			t.Errorf("optional key %q present despite not being set", key)
		}
	}
}

func TestConflictResolutionJSONRoundTrip(t *testing.T) {
	r := fullConflictResolution(t)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ConflictResolution
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Statement() != r.Statement() {
		t.Error("Statement mismatch")
	}
	if !decoded.Mechanism().Equal(r.Mechanism()) {
		t.Error("Mechanism mismatch")
	}
	if _, ok := decoded.ResolvingDecision(); !ok {
		t.Error("ResolvingDecision absent after round trip")
	}
	if _, ok := decoded.Provenance(); !ok {
		t.Error("Provenance absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestConflictResolutionJSONNullRejectedPerKey(t *testing.T) {
	base := `{"id":"r-1","conflict":"c-1","mechanism":"peos:priority","statement":"s"`
	if err := json.Unmarshal([]byte(base+`,"resolving_decision":null}`), new(ConflictResolution)); err == nil {
		t.Error("null resolving_decision accepted, want error")
	}
	if err := json.Unmarshal([]byte(base+`,"provenance":null}`), new(ConflictResolution)); err == nil {
		t.Error("null provenance accepted, want error")
	}
	if err := json.Unmarshal([]byte(base+`,"extension":null}`), new(ConflictResolution)); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestConflictResolutionJSONUnknownFieldIgnored(t *testing.T) {
	payload := `{"id":"r-1","conflict":"c-1","mechanism":"peos:priority","statement":"s","unknown_field":123}`
	var r ConflictResolution
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatal(err)
	}
}

func TestConflictResolutionZeroMarshalRejected(t *testing.T) {
	var r ConflictResolution
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidConflictResolution) {
		t.Errorf("error = %v, want %v", err, ErrInvalidConflictResolution)
	}
}

func TestConflictResolutionUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullConflictResolution(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Statement() != original.Statement() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// --- Cross-type: resolution -> conflict linkage ---------------------------

func TestConflictResolutionReferencesConflictID(t *testing.T) {
	conflict := fullDecisionConflict(t)
	r, err := NewConflictResolution(mustConflictResolutionID(t, "r-1"), conflict.ID(), ResolutionMechanismSupersession, "resolved")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Conflict().Equal(conflict.ID()) {
		t.Error("ConflictResolution.Conflict() does not reference the DecisionConflict's own ID")
	}
}
