package decision

import (
	"encoding/json"
	"errors"
	"testing"
)

// --- SupersessionID --------------------------------------------------------

func TestNewSupersessionIDValid(t *testing.T) {
	id, err := NewSupersessionID("sup-1")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "sup-1" {
		t.Errorf("String() = %q, want %q", id.String(), "sup-1")
	}
}

func TestNewSupersessionIDTrimsWhitespace(t *testing.T) {
	id, err := NewSupersessionID("  sup-1  ")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "sup-1" {
		t.Errorf("String() = %q, want %q", id.String(), "sup-1")
	}
}

func TestNewSupersessionIDEmptyRejected(t *testing.T) {
	if _, err := NewSupersessionID(""); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestNewSupersessionIDWhitespaceOnlyRejected(t *testing.T) {
	if _, err := NewSupersessionID("   "); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionIDEqual(t *testing.T) {
	a, err := NewSupersessionID("sup-1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewSupersessionID("sup-1")
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewSupersessionID("sup-2")
	if err != nil {
		t.Fatal(err)
	}
	if !a.Equal(b) {
		t.Error("Equal(same value) = false")
	}
	if a.Equal(c) {
		t.Error("Equal(different value) = true")
	}
}

func TestSupersessionIDIsZero(t *testing.T) {
	var id SupersessionID
	if !id.IsZero() {
		t.Error("zero SupersessionID IsZero() = false")
	}
	valid, err := NewSupersessionID("sup-1")
	if err != nil {
		t.Fatal(err)
	}
	if valid.IsZero() {
		t.Error("valid SupersessionID IsZero() = true")
	}
}

func TestSupersessionIDJSONRoundTrip(t *testing.T) {
	id, err := NewSupersessionID("sup-1")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"sup-1"` {
		t.Errorf("Marshal = %s, want %q", data, `"sup-1"`)
	}
	var decoded SupersessionID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(id) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, id)
	}
}

func TestSupersessionIDZeroMarshalRejected(t *testing.T) {
	var id SupersessionID
	if _, err := json.Marshal(id); !errors.Is(err, ErrInvalidDecisionSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionSupersession)
	}
}

func TestSupersessionIDWrongJSONTypeRejected(t *testing.T) {
	var id SupersessionID
	if err := json.Unmarshal([]byte(`123`), &id); err == nil {
		t.Error("numeric JSON accepted, want error")
	}
}

func TestSupersessionIDUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewSupersessionID("sup-1")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`""`), &receiver); err == nil {
		t.Fatal("empty string accepted, want error")
	}
	if !receiver.Equal(original) {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- InvalidationID ----------------------------------------------------

func TestNewInvalidationIDValid(t *testing.T) {
	id, err := NewInvalidationID("inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "inv-1" {
		t.Errorf("String() = %q, want %q", id.String(), "inv-1")
	}
}

func TestNewInvalidationIDTrimsWhitespace(t *testing.T) {
	id, err := NewInvalidationID("  inv-1  ")
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "inv-1" {
		t.Errorf("String() = %q, want %q", id.String(), "inv-1")
	}
}

func TestNewInvalidationIDEmptyRejected(t *testing.T) {
	if _, err := NewInvalidationID(""); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestNewInvalidationIDWhitespaceOnlyRejected(t *testing.T) {
	if _, err := NewInvalidationID("   "); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationIDEqual(t *testing.T) {
	a, err := NewInvalidationID("inv-1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewInvalidationID("inv-1")
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewInvalidationID("inv-2")
	if err != nil {
		t.Fatal(err)
	}
	if !a.Equal(b) {
		t.Error("Equal(same value) = false")
	}
	if a.Equal(c) {
		t.Error("Equal(different value) = true")
	}
}

func TestInvalidationIDIsZero(t *testing.T) {
	var id InvalidationID
	if !id.IsZero() {
		t.Error("zero InvalidationID IsZero() = false")
	}
	valid, err := NewInvalidationID("inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if valid.IsZero() {
		t.Error("valid InvalidationID IsZero() = true")
	}
}

func TestInvalidationIDJSONRoundTrip(t *testing.T) {
	id, err := NewInvalidationID("inv-1")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"inv-1"` {
		t.Errorf("Marshal = %s, want %q", data, `"inv-1"`)
	}
	var decoded InvalidationID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(id) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, id)
	}
}

func TestInvalidationIDZeroMarshalRejected(t *testing.T) {
	var id InvalidationID
	if _, err := json.Marshal(id); !errors.Is(err, ErrInvalidDecisionInvalidation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionInvalidation)
	}
}

func TestInvalidationIDWrongJSONTypeRejected(t *testing.T) {
	var id InvalidationID
	if err := json.Unmarshal([]byte(`123`), &id); err == nil {
		t.Error("numeric JSON accepted, want error")
	}
}

func TestInvalidationIDUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewInvalidationID("inv-1")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`""`), &receiver); err == nil {
		t.Fatal("empty string accepted, want error")
	}
	if !receiver.Equal(original) {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- Compile-time distinctness -------------------------------------------
//
// SupersessionID and InvalidationID intentionally do not share a field
// name, so an explicit Go type conversion between them (or with
// core.ImmutableRecordID) fails to compile -- the same guarantee
// documented on every identity type in peos/core/identity.go. There is no
// runtime test for a compile-time property; this comment, and the
// distinct field names in identity.go, are the enforcement mechanism.
