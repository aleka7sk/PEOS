package core

import (
	"encoding/json"
	"fmt"
)

// AuthorityRef is a typed reference to whoever or whatever holds the
// authority behind a decision, approval, or governed action. It is a
// reference/descriptor only: this package does not define an Authority
// aggregate, does not give AuthorityRef a lifecycle, does not define
// Waiver (PEOS-005's own deferred structural model), and does not define
// what a governance outcome is. A construct that requires authority (a
// Requirement Authority element, a Claim's authority field, a Transition
// Record's actor authority) references an AuthorityRef; this package
// stops at the reference.
type AuthorityRef struct {
	namespace  string
	identifier string
	kind       VocabularyValue
}

// NewAuthorityRef validates namespace and identifier and returns an
// AuthorityRef with no kind set.
func NewAuthorityRef(namespace, identifier string) (AuthorityRef, error) {
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return AuthorityRef{}, fmt.Errorf("core: NewAuthorityRef: %w", err)
	}
	id, err := normalizeIdentityValue(identifier)
	if err != nil {
		return AuthorityRef{}, fmt.Errorf("core: NewAuthorityRef: %w", err)
	}
	return AuthorityRef{namespace: ns, identifier: id}, nil
}

// WithKind returns a copy of a with its authority kind set (for example,
// distinguishing a role-based authority from a named-individual
// authority). kind is optional; the zero AuthorityRef has no kind.
func (a AuthorityRef) WithKind(kind VocabularyValue) AuthorityRef {
	a.kind = kind
	return a
}

func (a AuthorityRef) Namespace() string  { return a.namespace }
func (a AuthorityRef) Identifier() string { return a.identifier }
func (a AuthorityRef) Kind() (VocabularyValue, bool) {
	return a.kind, !a.kind.IsZero()
}
func (a AuthorityRef) IsZero() bool { return a.namespace == "" && a.identifier == "" }

type authorityRefJSON struct {
	Namespace  string           `json:"namespace"`
	Identifier string           `json:"identifier"`
	Kind       *VocabularyValue `json:"kind,omitempty"`
}

// MarshalJSON encodes a as {"namespace": ..., "identifier": ..., "kind": ...}.
func (a AuthorityRef) MarshalJSON() ([]byte, error) {
	raw := authorityRefJSON{Namespace: a.namespace, Identifier: a.identifier}
	if !a.kind.IsZero() {
		raw.Kind = &a.kind
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes a from its JSON form.
func (a *AuthorityRef) UnmarshalJSON(data []byte) error {
	var raw authorityRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal AuthorityRef: %w", err)
	}
	v, err := NewAuthorityRef(raw.Namespace, raw.Identifier)
	if err != nil {
		return err
	}
	if raw.Kind != nil {
		v = v.WithKind(*raw.Kind)
	}
	*a = v
	return nil
}
