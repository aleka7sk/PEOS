package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustConflictID(t *testing.T, value string) ConflictID {
	t.Helper()
	id, err := NewConflictID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustConflictDecisionRef(t *testing.T, value string) core.DecisionRef {
	t.Helper()
	ref, err := core.NewDecisionRef(mustDecisionID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func fullDecisionConflict(t *testing.T) DecisionConflict {
	t.Helper()
	c, err := NewDecisionConflict(
		mustConflictID(t, "conflict-1"),
		mustConflictDecisionRef(t, "dec-1"),
		mustConflictDecisionRef(t, "dec-2"),
		mustScope(t, "/services/*"),
		"dec-1 requires PostgreSQL; dec-2 requires MySQL for the same component",
		"dec-2 has higher priority under the architecture policy",
	)
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	c, err = c.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return c.WithExtension(ext)
}

// --- ConflictID --------------------------------------------------------

func TestNewConflictIDValid(t *testing.T) {
	id, err := NewConflictID("conflict-1")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "conflict-1" {
		t.Errorf("String() = %q, want %q", id.String(), "conflict-1")
	}
}

func TestNewConflictIDEmptyRejected(t *testing.T) {
	if _, err := NewConflictID(""); !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewConflictIDWhitespaceOnlyRejected(t *testing.T) {
	if _, err := NewConflictID("   "); !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestConflictIDTrimmed(t *testing.T) {
	id, err := NewConflictID("  conflict-1  ")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "conflict-1" {
		t.Errorf("String() = %q, want trimmed %q", id.String(), "conflict-1")
	}
}

func TestConflictIDEqual(t *testing.T) {
	a := mustConflictID(t, "conflict-1")
	b := mustConflictID(t, "conflict-1")
	c := mustConflictID(t, "conflict-2")
	if !a.Equal(b) {
		t.Error("Equal() = false for identical IDs")
	}
	if a.Equal(c) {
		t.Error("Equal() = true for different IDs")
	}
}

func TestConflictIDIsZero(t *testing.T) {
	var id ConflictID
	if !id.IsZero() {
		t.Error("zero ConflictID IsZero() = false")
	}
	if mustConflictID(t, "x").IsZero() {
		t.Error("valid ConflictID IsZero() = true")
	}
}

func TestConflictIDZeroMarshalRejected(t *testing.T) {
	var id ConflictID
	if _, err := json.Marshal(id); !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestConflictIDUnmarshalMalformedJSONRejected(t *testing.T) {
	var id ConflictID
	if err := json.Unmarshal([]byte(`123`), &id); err == nil {
		t.Error("non-string JSON accepted, want error")
	}
}

func TestConflictIDJSONRoundTrip(t *testing.T) {
	id := mustConflictID(t, "conflict-1")
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ConflictID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(id) {
		t.Error("round trip mismatch")
	}
}

// --- NewDecisionConflict ------------------------------------------------

func TestNewDecisionConflictValid(t *testing.T) {
	c := fullDecisionConflict(t)
	if c.IsZero() {
		t.Error("valid DecisionConflict IsZero() = true")
	}
}

func TestNewDecisionConflictZeroIDRejected(t *testing.T) {
	_, err := NewDecisionConflict(ConflictID{}, mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictZeroDecisionARejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), core.DecisionRef{}, mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictZeroDecisionBRejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), core.DecisionRef{}, mustScope(t, "/x"), "incompatible", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

// TestNewDecisionConflictSelfConflictRejected proves the binary-arity
// self-conflict rule: a Decision has no revision mechanism, so it cannot
// conflict with itself.
func TestNewDecisionConflictSelfConflictRejected(t *testing.T) {
	same := mustConflictDecisionRef(t, "dec-1")
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), same, same, mustScope(t, "/x"), "incompatible", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

// TestNewDecisionConflictBinaryArity proves the type has exactly two
// Decision fields (DecisionA, DecisionB) and no collection of Decisions,
// confirming binary rather than N-ary representation.
func TestNewDecisionConflictBinaryArity(t *testing.T) {
	c := fullDecisionConflict(t)
	a := c.DecisionA()
	b := c.DecisionB()
	if a.IsZero() || b.IsZero() {
		t.Fatal("DecisionA/DecisionB must both be set")
	}
	if a == b {
		t.Fatal("DecisionA and DecisionB must differ")
	}
}

func TestDecisionConflictOverlappingScope(t *testing.T) {
	c := fullDecisionConflict(t)
	if c.OverlappingScope().IsZero() {
		t.Error("OverlappingScope() returned zero value")
	}
}

func TestNewDecisionConflictZeroScopeRejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), core.Scope{}, "incompatible", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictEmptyIncompatibilityRejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictWhitespaceIncompatibilityRejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "   ", "priority")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictEmptyGoverningRuleRejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictWhitespaceGoverningRuleRejected(t *testing.T) {
	_, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "   ")
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestNewDecisionConflictRawStringsPreserved(t *testing.T) {
	c, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "  padded incompatibility  ", "  padded rule  ")
	if err != nil {
		t.Fatal(err)
	}
	if c.Incompatibility() != "  padded incompatibility  " {
		t.Errorf("Incompatibility() = %q, want raw preserved", c.Incompatibility())
	}
	if c.GoverningRule() != "  padded rule  " {
		t.Errorf("GoverningRule() = %q, want raw preserved", c.GoverningRule())
	}
}

// --- With* / accessors --------------------------------------------------

func TestDecisionConflictWithProvenance(t *testing.T) {
	c, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Provenance(); ok {
		t.Error("Provenance() ok = true before WithProvenance")
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	withProv, err := c.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := withProv.Provenance(); !ok {
		t.Error("Provenance() ok = false after WithProvenance")
	}
	if _, ok := c.Provenance(); ok {
		t.Error("WithProvenance mutated the original receiver")
	}
}

func TestDecisionConflictZeroProvenanceRejected(t *testing.T) {
	c, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.WithProvenance(core.Provenance{}); !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestDecisionConflictWithoutProvenance(t *testing.T) {
	c, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
	if err != nil {
		t.Fatal(err)
	}
	prov := core.NewProvenance().WithExternalSourceID("ext-1")
	c, err = c.WithProvenance(prov)
	if err != nil {
		t.Fatal(err)
	}
	cleared := c.WithoutProvenance()
	if _, ok := cleared.Provenance(); ok {
		t.Error("Provenance() ok = true after WithoutProvenance")
	}
	if _, ok := c.Provenance(); !ok {
		t.Error("WithoutProvenance mutated the original receiver")
	}
}

func TestDecisionConflictWithExtension(t *testing.T) {
	c, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
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

func TestDecisionConflictIsZero(t *testing.T) {
	var c DecisionConflict
	if !c.IsZero() {
		t.Error("zero DecisionConflict IsZero() = false")
	}
	if fullDecisionConflict(t).IsZero() {
		t.Error("valid DecisionConflict IsZero() = true")
	}
}

// --- JSON ----------------------------------------------------------------

func TestDecisionConflictJSONLiteralWireKeys(t *testing.T) {
	c := fullDecisionConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "decision_a", "decision_b", "overlapping_scope", "incompatibility", "governing_rule", "provenance", "extension"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
}

func TestDecisionConflictJSONMinimumOmitsOptionalFields(t *testing.T) {
	c, err := NewDecisionConflict(mustConflictID(t, "c-1"), mustConflictDecisionRef(t, "dec-1"), mustConflictDecisionRef(t, "dec-2"), mustScope(t, "/x"), "incompatible", "priority")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"provenance", "extension"} {
		if _, present := raw[key]; present {
			t.Errorf("optional key %q present despite not being set", key)
		}
	}
}

func TestDecisionConflictJSONRoundTrip(t *testing.T) {
	c := fullDecisionConflict(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DecisionConflict
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Incompatibility() != c.Incompatibility() || decoded.GoverningRule() != c.GoverningRule() {
		t.Error("round trip mismatch")
	}
	if _, ok := decoded.Provenance(); !ok {
		t.Error("Provenance absent after round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestDecisionConflictJSONNullRejectedPerKey(t *testing.T) {
	base := `{"id":"c-1","decision_a":{"decision_id":"dec-1"},"decision_b":{"decision_id":"dec-2"},"overlapping_scope":{"kind":"product-x:path","expression":"/x"},"incompatibility":"i","governing_rule":"g"`
	if err := json.Unmarshal([]byte(base+`,"provenance":null}`), new(DecisionConflict)); err == nil {
		t.Error("null provenance accepted, want error")
	}
	if err := json.Unmarshal([]byte(base+`,"extension":null}`), new(DecisionConflict)); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestDecisionConflictJSONUnknownFieldIgnored(t *testing.T) {
	base := `{"id":"c-1","decision_a":{"decision_id":"dec-1"},"decision_b":{"decision_id":"dec-2"},"overlapping_scope":{"kind":"product-x:path","expression":"/x"},"incompatibility":"i","governing_rule":"g","unknown_field":123}`
	var c DecisionConflict
	if err := json.Unmarshal([]byte(base), &c); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionConflictZeroMarshalRejected(t *testing.T) {
	var c DecisionConflict
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestDecisionConflictUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullDecisionConflict(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Incompatibility() != original.Incompatibility() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// TestDecisionConflictSelfConflictRejectedViaUnmarshal proves the
// constructor/Unmarshal equivalence for the self-conflict rule: JSON
// cannot construct a state the constructor forbids.
func TestDecisionConflictSelfConflictRejectedViaUnmarshal(t *testing.T) {
	payload := `{"id":"c-1","decision_a":{"decision_id":"dec-1"},"decision_b":{"decision_id":"dec-1"},"overlapping_scope":{"kind":"product-x:path","expression":"/x"},"incompatibility":"i","governing_rule":"g"}`
	var c DecisionConflict
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionConflict)
	}
}

func TestDecisionConflictNestedSentinelPreserved(t *testing.T) {
	payload := `{"id":"c-1","decision_a":{"decision_id":"dec-1"},"decision_b":{"decision_id":"dec-2"},"overlapping_scope":{"kind":"","expression":""},"incompatibility":"i","governing_rule":"g"}`
	var c DecisionConflict
	err := json.Unmarshal([]byte(payload), &c)
	if err == nil {
		t.Fatal("malformed nested scope accepted, want error")
	}
	if !errors.Is(err, ErrInvalidDecisionConflict) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidDecisionConflict)
	}
}
