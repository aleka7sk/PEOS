package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustRoleActor(t *testing.T, identifier string) core.ActorRef {
	t.Helper()
	a, err := core.NewActorRef("user", identifier)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func fullRole(t *testing.T) Role {
	t.Helper()
	r, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return r.WithExtension(ext)
}

// --- RoleKind ------------------------------------------------------------

func TestRoleKindPredeclaredNonZero(t *testing.T) {
	for _, k := range []RoleKind{
		RoleKindAuthor, RoleKindProposer, RoleKindMaker, RoleKindApprover,
		RoleKindReviewer, RoleKindExecutor, RoleKindRecorder, RoleKindOwner,
	} {
		if k.IsZero() {
			t.Errorf("predeclared RoleKind %v is zero", k)
		}
	}
}

// TestRoleKindOpenVocabularyAcceptsProductDefined proves RoleKind is
// open: PEOS-004 :781 permits a Product contract to define additional
// roles beyond the eight predeclared here.
func TestRoleKindOpenVocabularyAcceptsProductDefined(t *testing.T) {
	v, err := core.NewVocabularyValue("product-x", "custom-role")
	if err != nil {
		t.Fatal(err)
	}
	custom := NewRoleKind(v)
	r, err := NewRole(mustRoleActor(t, "alice"), custom)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Kind().Equal(custom) {
		t.Error("custom role kind not preserved")
	}
}

func TestRoleKindValueAndString(t *testing.T) {
	if RoleKindApprover.Value().IsZero() {
		t.Error("Value() returned zero value")
	}
	if RoleKindApprover.String() != "peos:approver" {
		t.Errorf("String() = %q, want %q", RoleKindApprover.String(), "peos:approver")
	}
}

func TestRoleKindZeroMarshalRejected(t *testing.T) {
	var k RoleKind
	if _, err := json.Marshal(k); !errors.Is(err, ErrInvalidDecisionRole) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRole)
	}
}

// --- NewRole ---------------------------------------------------------------

func TestNewRoleValid(t *testing.T) {
	r := fullRole(t)
	if r.IsZero() {
		t.Error("valid Role IsZero() = true")
	}
}

func TestNewRoleZeroActorRejected(t *testing.T) {
	_, err := NewRole(core.ActorRef{}, RoleKindApprover)
	if !errors.Is(err, ErrInvalidDecisionRole) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRole)
	}
}

func TestNewRoleZeroKindRejected(t *testing.T) {
	_, err := NewRole(mustRoleActor(t, "alice"), RoleKind{})
	if !errors.Is(err, ErrInvalidDecisionRole) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRole)
	}
}

// --- With* / accessors ------------------------------------------------------

func TestRoleWithExtension(t *testing.T) {
	r, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
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

func TestRoleImmutability(t *testing.T) {
	r, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	original := r
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = r.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

func TestRoleIsZero(t *testing.T) {
	var r Role
	if !r.IsZero() {
		t.Error("zero Role IsZero() = false")
	}
	if fullRole(t).IsZero() {
		t.Error("valid Role IsZero() = true")
	}
}

// --- JSON --------------------------------------------------------------

func TestRoleJSONLiteralWireKeys(t *testing.T) {
	r := fullRole(t)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"actor", "kind", "extension"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
}

func TestRoleJSONMinimumOmitsExtension(t *testing.T) {
	r, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
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
	if _, present := raw["extension"]; present {
		t.Error("extension present despite not being set")
	}
}

func TestRoleJSONRoundTrip(t *testing.T) {
	r := fullRole(t)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Role
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Actor() != r.Actor() {
		t.Error("Actor mismatch")
	}
	if !decoded.Kind().Equal(r.Kind()) {
		t.Error("Kind mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestRoleJSONNullExtensionRejected(t *testing.T) {
	payload := `{"actor":{"namespace":"user","identifier":"alice"},"kind":"peos:approver","extension":null}`
	var r Role
	if err := json.Unmarshal([]byte(payload), &r); err == nil {
		t.Error("null extension accepted, want error")
	}
}

func TestRoleJSONUnknownFieldIgnored(t *testing.T) {
	payload := `{"actor":{"namespace":"user","identifier":"alice"},"kind":"peos:approver","unknown_field":123}`
	var r Role
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		t.Fatal(err)
	}
}

func TestRoleZeroMarshalRejected(t *testing.T) {
	var r Role
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidDecisionRole) {
		t.Errorf("error = %v, want %v", err, ErrInvalidDecisionRole)
	}
}

func TestRoleUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullRole(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Actor() != original.Actor() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// TestRoleActorIdentityNotInferredFromKind proves Role identity is
// carried by an explicit core.ActorRef (PEOS-004 :785), not derived from
// RoleKind or any other field.
func TestRoleActorIdentityNotInferredFromKind(t *testing.T) {
	r1, err := NewRole(mustRoleActor(t, "alice"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewRole(mustRoleActor(t, "bob"), RoleKindApprover)
	if err != nil {
		t.Fatal(err)
	}
	if r1.Actor() == r2.Actor() {
		t.Error("distinct actors compared equal")
	}
}
