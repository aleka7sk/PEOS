package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustAuthorityRef(t *testing.T, namespace, identifier string) core.AuthorityRef {
	t.Helper()
	ref, err := core.NewAuthorityRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestNewAuthorityRequirementOnly(t *testing.T) {
	req := mustAuthorityRef(t, "role", "architect")
	a, err := NewAuthority([]core.AuthorityRef{req}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Requirements()) != 1 || len(a.Bases()) != 0 {
		t.Errorf("Requirements/Bases = %v/%v, want 1/0", a.Requirements(), a.Bases())
	}
}

func TestNewAuthorityBasisOnly(t *testing.T) {
	basis := mustAuthorityRef(t, "role", "cto")
	a, err := NewAuthority(nil, []core.AuthorityRef{basis})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Bases()) != 1 || len(a.Requirements()) != 0 {
		t.Errorf("Bases/Requirements = %v/%v, want 1/0", a.Bases(), a.Requirements())
	}
}

func TestNewAuthorityBothAccepted(t *testing.T) {
	req := mustAuthorityRef(t, "role", "architect")
	basis := mustAuthorityRef(t, "role", "cto")
	a, err := NewAuthority([]core.AuthorityRef{req}, []core.AuthorityRef{basis})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Requirements()) != 1 || len(a.Bases()) != 1 {
		t.Errorf("Requirements/Bases = %v/%v, want 1/1", a.Requirements(), a.Bases())
	}
}

func TestNewAuthorityNeitherRejected(t *testing.T) {
	if _, err := NewAuthority(nil, nil); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func TestNewAuthorityZeroRefRejected(t *testing.T) {
	if _, err := NewAuthority([]core.AuthorityRef{{}}, nil); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("zero requirement ref: error = %v, want %v", err, ErrInvalidAuthority)
	}
	if _, err := NewAuthority(nil, []core.AuthorityRef{{}}); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("zero basis ref: error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func TestAuthorityDefensiveCopies(t *testing.T) {
	req := mustAuthorityRef(t, "role", "architect")
	input := []core.AuthorityRef{req}
	a, err := NewAuthority(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = core.AuthorityRef{}
	if a.Requirements()[0].IsZero() {
		t.Error("NewAuthority did not defensively copy input slice")
	}
	got := a.Requirements()
	got[0] = core.AuthorityRef{}
	if a.Requirements()[0].IsZero() {
		t.Error("Requirements() did not defensively copy on return")
	}
}

func TestAuthorityJSONRoundTrip(t *testing.T) {
	req := mustAuthorityRef(t, "role", "architect")
	basis := mustAuthorityRef(t, "role", "cto")
	a, err := NewAuthority([]core.AuthorityRef{req}, []core.AuthorityRef{basis})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Authority
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Requirements()) != 1 || len(decoded.Bases()) != 1 {
		t.Errorf("round trip mismatch: got %+v", decoded)
	}
}

func TestAuthorityJSONMinimum(t *testing.T) {
	basis := mustAuthorityRef(t, "role", "cto")
	a, err := NewAuthority(nil, []core.AuthorityRef{basis})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["requirements"]; present {
		t.Error("requirements present despite not being set")
	}
	if _, present := raw["bases"]; !present {
		t.Error("bases absent despite being set")
	}
}

func TestAuthorityExplicitNullRejected(t *testing.T) {
	var a Authority
	if err := json.Unmarshal([]byte(`{"requirements":null,"bases":[{"namespace":"role","identifier":"cto"}]}`), &a); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("null requirements: error = %v, want %v", err, ErrInvalidAuthority)
	}
	if err := json.Unmarshal([]byte(`{"requirements":[{"namespace":"role","identifier":"cto"}],"bases":null}`), &a); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("null bases: error = %v, want %v", err, ErrInvalidAuthority)
	}
}

func TestAuthorityUnknownFieldIgnored(t *testing.T) {
	var a Authority
	payload := `{"bases":[{"namespace":"role","identifier":"cto"}],"unknown_field":123}`
	if err := json.Unmarshal([]byte(payload), &a); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorityUnmarshalFailurePreservesReceiver(t *testing.T) {
	basis := mustAuthorityRef(t, "role", "cto")
	original, err := NewAuthority(nil, []core.AuthorityRef{basis})
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if len(receiver.Bases()) != 1 {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestAuthorityZeroMarshalRejected(t *testing.T) {
	var a Authority
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidAuthority) {
		t.Errorf("error = %v, want %v", err, ErrInvalidAuthority)
	}
}
