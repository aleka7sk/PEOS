package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func mustArtifactType(t *testing.T, namespace, value string) ArtifactType {
	t.Helper()
	return NewArtifactType(mustVocabularyValue(t, namespace, value))
}

func mustArtifactRole(t *testing.T, namespace, value string) ArtifactRole {
	t.Helper()
	return NewArtifactRole(mustVocabularyValue(t, namespace, value))
}

func TestNewArtifact(t *testing.T) {
	id := mustArtifactID(t, "ART-1")
	typ := mustArtifactType(t, "peos", "requirement")

	a, err := NewArtifact(id, typ)
	if err != nil {
		t.Fatal(err)
	}
	if a.IsZero() {
		t.Error("valid Artifact reports IsZero() = true")
	}
	if a.ID() != id {
		t.Errorf("ID() = %v, want %v", a.ID(), id)
	}
	if a.Type() != typ {
		t.Errorf("Type() = %v, want %v", a.Type(), typ)
	}
	if roles := a.Roles(); roles != nil {
		t.Errorf("Roles() = %v, want nil", roles)
	}
}

func TestNewArtifactZeroID(t *testing.T) {
	typ := mustArtifactType(t, "peos", "requirement")
	if _, err := NewArtifact(ArtifactID{}, typ); !errors.Is(err, ErrInvalidArtifact) || !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("error = %v, want wrapping both %v and %v", err, ErrInvalidArtifact, ErrEmptyIdentity)
	}
}

func TestNewArtifactZeroType(t *testing.T) {
	id := mustArtifactID(t, "ART-1")
	if _, err := NewArtifact(id, ArtifactType{}); !errors.Is(err, ErrInvalidArtifact) {
		t.Errorf("error = %v, want %v", err, ErrInvalidArtifact)
	}
}

func TestNewArtifactUnknownType(t *testing.T) {
	id := mustArtifactID(t, "ART-1")
	typ := mustArtifactType(t, "product-x", "custom-type")
	a, err := NewArtifact(id, typ)
	if err != nil {
		t.Fatal(err)
	}
	if a.Type() != typ {
		t.Errorf("unknown Artifact Type not preserved: got %v, want %v", a.Type(), typ)
	}
}

func TestArtifactWithRolesNone(t *testing.T) {
	a, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithRoles()
	if err != nil {
		t.Fatal(err)
	}
	if a.Roles() != nil {
		t.Errorf("Roles() = %v, want nil", a.Roles())
	}
}

func TestArtifactWithRolesOneAndMultiple(t *testing.T) {
	evidence := mustArtifactRole(t, "peos", "evidence")
	other := mustArtifactRole(t, "product-x", "custom-role")

	base, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}

	one, err := base.WithRoles(evidence)
	if err != nil {
		t.Fatal(err)
	}
	if got := one.Roles(); len(got) != 1 || got[0] != evidence {
		t.Errorf("Roles() = %v, want [%v]", got, evidence)
	}

	many, err := base.WithRoles(evidence, other)
	if err != nil {
		t.Fatal(err)
	}
	got := many.Roles()
	if len(got) != 2 || got[0] != evidence || got[1] != other {
		t.Errorf("Roles() = %v, want [%v %v] (declaration order preserved)", got, evidence, other)
	}
}

func TestArtifactWithRolesDuplicateRejected(t *testing.T) {
	evidence := mustArtifactRole(t, "peos", "evidence")
	base, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := base.WithRoles(evidence, evidence); !errors.Is(err, ErrDuplicateArtifactRole) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateArtifactRole)
	}
}

func TestArtifactWithRolesZeroRoleRejected(t *testing.T) {
	base, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := base.WithRoles(ArtifactRole{}); !errors.Is(err, ErrInvalidArtifact) {
		t.Errorf("error = %v, want %v", err, ErrInvalidArtifact)
	}
}

func TestArtifactWithRolesChainedCallsReplaceNotAccumulate(t *testing.T) {
	roleA := mustArtifactRole(t, "peos", "evidence")
	roleB := mustArtifactRole(t, "peos", "authoritative")

	base, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}

	a1, err := base.WithRoles(roleA)
	if err != nil {
		t.Fatal(err)
	}
	a2, err := a1.WithRoles(roleB)
	if err != nil {
		t.Fatal(err)
	}

	if got := a2.Roles(); len(got) != 1 || got[0] != roleB {
		t.Errorf("a2.Roles() = %v, want [%v] (second WithRoles call must replace, not accumulate)", got, roleB)
	}
	if got := a1.Roles(); len(got) != 1 || got[0] != roleA {
		t.Errorf("a1.Roles() = %v, want [%v] (a1 must remain unaffected by a2's WithRoles call)", got, roleA)
	}
	if got := base.Roles(); got != nil {
		t.Errorf("base.Roles() = %v, want nil (base must remain unaffected)", got)
	}

	a3, err := a2.WithRoles()
	if err != nil {
		t.Fatal(err)
	}
	if got := a3.Roles(); got != nil {
		t.Errorf("a3.Roles() after no-args WithRoles() = %v, want nil", got)
	}
	if got := a2.Roles(); len(got) != 1 || got[0] != roleB {
		t.Errorf("a2.Roles() = %v, want [%v] (a2 must remain unaffected by a3's clearing call)", got, roleB)
	}
}

func TestArtifactRolesDefensiveCopy(t *testing.T) {
	evidence := mustArtifactRole(t, "peos", "evidence")
	a, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	a, err = a.WithRoles(evidence)
	if err != nil {
		t.Fatal(err)
	}
	got := a.Roles()
	got[0] = mustArtifactRole(t, "product-x", "tampered")
	again := a.Roles()
	if again[0] != evidence {
		t.Errorf("mutating a Roles() result affected internal state: got %v, want %v", again[0], evidence)
	}
}

func TestArtifactWithScope(t *testing.T) {
	a, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Scope(); ok {
		t.Error("Scope() ok = true before WithScope was called")
	}

	scope, err := NewScope(mustVocabularyValue(t, "peos", "product-scope"), "product-x")
	if err != nil {
		t.Fatal(err)
	}
	a = a.WithScope(scope)
	got, ok := a.Scope()
	if !ok || !got.Equal(scope) {
		t.Errorf("Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}

	// Passing the zero Scope clears it.
	cleared := a.WithScope(Scope{})
	if _, ok := cleared.Scope(); ok {
		t.Error("Scope() ok = true after WithScope(Scope{})")
	}
}

func TestArtifactWithExtension(t *testing.T) {
	a, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	a = a.WithExtension(ext)
	got, ok := a.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestArtifactJSONRoundTrip(t *testing.T) {
	evidence := mustArtifactRole(t, "peos", "evidence")
	scope, err := NewScope(mustVocabularyValue(t, "peos", "product-scope"), "product-x")
	if err != nil {
		t.Fatal(err)
	}
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}

	original, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithRoles(evidence)
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithScope(scope).WithExtension(ext)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Artifact
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != original.ID() || decoded.Type() != original.Type() {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, original)
	}
	if got := decoded.Roles(); len(got) != 1 || got[0] != evidence {
		t.Errorf("round trip Roles() = %v, want [%v]", got, evidence)
	}
	if gotScope, ok := decoded.Scope(); !ok || !gotScope.Equal(scope) {
		t.Errorf("round trip Scope() = (%v, %v), want (%v, true)", gotScope, ok, scope)
	}
}

func TestArtifactJSONMinimalNoOptionalFields(t *testing.T) {
	original, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
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
	for _, absent := range []string{"roles", "scope", "extension"} {
		if _, present := raw[absent]; present {
			t.Errorf("%q present in JSON despite not being set", absent)
		}
	}
}

func TestArtifactJSONMalformed(t *testing.T) {
	var a Artifact
	// Empty artifact_id fails inside ArtifactID's own UnmarshalJSON before
	// NewArtifact is ever reached.
	if err := json.Unmarshal([]byte(`{"artifact_id":"", "artifact_type":"peos:requirement"}`), &a); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("empty id: error = %v, want %v", err, ErrEmptyIdentity)
	}
	// Empty artifact_type fails inside VocabularyValue's own parsing
	// (missing the required "namespace:value" separator) before
	// NewArtifact is ever reached.
	if err := json.Unmarshal([]byte(`{"artifact_id":"ART-1", "artifact_type":""}`), &a); !errors.Is(err, ErrInvalidVocabularyValue) {
		t.Errorf("empty type: error = %v, want %v", err, ErrInvalidVocabularyValue)
	}
	if err := json.Unmarshal([]byte(`{"artifact_id":"ART-1", "artifact_type":"peos:requirement", "roles":["peos:evidence","peos:evidence"]}`), &a); !errors.Is(err, ErrDuplicateArtifactRole) {
		t.Errorf("duplicate roles: error = %v, want %v", err, ErrDuplicateArtifactRole)
	}
}

func TestArtifactUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	evidence := mustArtifactRole(t, "peos", "evidence")
	original, err := NewArtifact(mustArtifactID(t, "ART-1"), mustArtifactType(t, "peos", "requirement"))
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithRoles(evidence)
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"artifact_id":"", "artifact_type":""}`), &receiver); err == nil {
		t.Fatal("malformed Artifact JSON accepted, want error")
	}
	// Artifact is not comparable with == (it holds a []ArtifactRole slice and
	// an Extension map), so every field is checked individually via its own
	// accessor.
	if receiver.ID() != original.ID() || receiver.Type() != original.Type() {
		t.Errorf("failed Unmarshal changed ID()/Type(): got (%v, %v), want (%v, %v)",
			receiver.ID(), receiver.Type(), original.ID(), original.Type())
	}
	gotRoles, wantRoles := receiver.Roles(), original.Roles()
	if len(gotRoles) != len(wantRoles) || gotRoles[0] != wantRoles[0] {
		t.Errorf("failed Unmarshal changed Roles(): got %v, want %v", gotRoles, wantRoles)
	}
}

func TestArtifactZeroValue(t *testing.T) {
	var a Artifact
	if !a.IsZero() {
		t.Error("zero-value Artifact.IsZero() = false, want true")
	}
	if !a.ID().IsZero() || !a.Type().IsZero() {
		t.Error("zero-value Artifact has non-zero ID or Type")
	}
}
