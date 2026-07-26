package decision

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Authority is the immutable value object recording a Decision's authority
// requirement and/or authority basis (PEOS-004: a Decision MUST have "an
// identifiable authority requirement or authority basis"; "Every
// applicable Decision MUST have an established authority basis"; a
// proposed Decision "MAY identify the authority required to establish its
// Outcome before that authority has been exercised").
//
// requirements and bases are kept as two separate slices, not one slice
// distinguished by core.AuthorityRef.WithKind, because Kind distinguishes
// the *nature* of an authority holder (a role-based authority from a
// named-individual authority) -- an axis orthogonal to whether that
// authority has been exercised, which is what requirements versus bases
// distinguishes here.
type Authority struct {
	requirements []core.AuthorityRef
	bases        []core.AuthorityRef
}

// NewAuthority validates requirements and bases and returns an Authority.
// At least one non-zero core.AuthorityRef is required across both slices
// combined; a zero-value ref in either slice is rejected.
func NewAuthority(requirements, bases []core.AuthorityRef) (Authority, error) {
	reqs, err := copyValidAuthorityRefs(requirements)
	if err != nil {
		return Authority{}, fmt.Errorf("decision: NewAuthority: %w", err)
	}
	bs, err := copyValidAuthorityRefs(bases)
	if err != nil {
		return Authority{}, fmt.Errorf("decision: NewAuthority: %w", err)
	}
	if len(reqs) == 0 && len(bs) == 0 {
		return Authority{}, fmt.Errorf("decision: NewAuthority: %w: at least one requirement or basis is required", ErrInvalidAuthority)
	}
	return Authority{requirements: reqs, bases: bs}, nil
}

func copyValidAuthorityRefs(refs []core.AuthorityRef) ([]core.AuthorityRef, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	cp := make([]core.AuthorityRef, len(refs))
	for idx, ref := range refs {
		if ref.IsZero() {
			return nil, fmt.Errorf("%w: authority ref must not be zero", ErrInvalidAuthority)
		}
		cp[idx] = ref
	}
	return cp, nil
}

// Requirements returns a defensive copy of a's declared authority
// requirements, in declaration order.
func (a Authority) Requirements() []core.AuthorityRef {
	if len(a.requirements) == 0 {
		return nil
	}
	cp := make([]core.AuthorityRef, len(a.requirements))
	copy(cp, a.requirements)
	return cp
}

// Bases returns a defensive copy of a's declared authority bases, in
// declaration order.
func (a Authority) Bases() []core.AuthorityRef {
	if len(a.bases) == 0 {
		return nil
	}
	cp := make([]core.AuthorityRef, len(a.bases))
	copy(cp, a.bases)
	return cp
}

// IsZero reports whether a is the zero value.
func (a Authority) IsZero() bool { return len(a.requirements) == 0 && len(a.bases) == 0 }

type authorityJSON struct {
	Requirements []core.AuthorityRef `json:"requirements,omitempty"`
	Bases        []core.AuthorityRef `json:"bases,omitempty"`
}

// MarshalJSON encodes a as {"requirements":[...], "bases":[...]}, omitting
// either key when empty.
func (a Authority) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return nil, fmt.Errorf("decision: marshal Authority: %w", ErrInvalidAuthority)
	}
	raw := authorityJSON{}
	if len(a.requirements) > 0 {
		raw.Requirements = a.requirements
	}
	if len(a.bases) > 0 {
		raw.Bases = a.bases
	}
	return json.Marshal(raw)
}

type authorityUnmarshalJSON struct {
	Requirements json.RawMessage `json:"requirements"`
	Bases        json.RawMessage `json:"bases"`
}

// UnmarshalJSON decodes a from its JSON form, applying the same validation
// as NewAuthority. An explicit JSON null for either "requirements" or
// "bases" is rejected; an absent key or an empty array both decode as no
// entries for that side.
func (a *Authority) UnmarshalJSON(data []byte) error {
	var raw authorityUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Authority: %w", err)
	}
	var requirements []core.AuthorityRef
	if len(raw.Requirements) > 0 {
		if string(raw.Requirements) == "null" {
			return fmt.Errorf("decision: unmarshal Authority: %w: requirements must not be null", ErrInvalidAuthority)
		}
		if err := json.Unmarshal(raw.Requirements, &requirements); err != nil {
			return fmt.Errorf("decision: unmarshal Authority: %w", err)
		}
	}
	var bases []core.AuthorityRef
	if len(raw.Bases) > 0 {
		if string(raw.Bases) == "null" {
			return fmt.Errorf("decision: unmarshal Authority: %w: bases must not be null", ErrInvalidAuthority)
		}
		if err := json.Unmarshal(raw.Bases, &bases); err != nil {
			return fmt.Errorf("decision: unmarshal Authority: %w", err)
		}
	}
	result, err := NewAuthority(requirements, bases)
	if err != nil {
		return err
	}
	*a = result
	return nil
}
