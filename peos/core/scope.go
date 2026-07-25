package core

import (
	"encoding/json"
	"fmt"
)

// Scope is a representation-independent scope value, used wherever a PEOS
// construct requires an explicit scope (Artifact Relation, Artifact
// Supersession, Requirement relation types, Validation Claim, and so on).
//
// This package does not define a query language or interpret Expression
// in any way; Scope only carries a discriminator (Kind) and an opaque
// expression, so that two different scope languages (for example, a
// simple path expression versus a structured filter) can coexist without
// this package favoring one. Interpreting Expression is the
// responsibility of the construct that owns the Scope.
//
// A constructed Scope always has both a valid Kind and a non-empty
// Expression; there is no such thing as a "present but empty" Scope from
// this constructor. Where a PEOS construct permits omitting scope
// entirely, that construct represents the absence with a *Scope field
// left nil (or an equivalent optionality mechanism) at its own layer,
// not with a zero-value Scope.
type Scope struct {
	kind       VocabularyValue
	expression string
}

// NewScope validates kind and expression and returns a Scope. kind must
// be a non-zero VocabularyValue and expression must be non-empty.
func NewScope(kind VocabularyValue, expression string) (Scope, error) {
	if kind.IsZero() {
		return Scope{}, fmt.Errorf("core: NewScope: missing kind: %w", ErrInvalidScope)
	}
	if expression == "" {
		return Scope{}, fmt.Errorf("core: NewScope: missing expression: %w", ErrInvalidScope)
	}
	return Scope{kind: kind, expression: expression}, nil
}

// Kind returns the scope's discriminator.
func (s Scope) Kind() VocabularyValue { return s.kind }

// Expression returns the scope's opaque expression, uninterpreted.
func (s Scope) Expression() string { return s.expression }

// IsZero reports whether s is the zero value.
func (s Scope) IsZero() bool { return s.kind.IsZero() && s.expression == "" }

// Equal reports whether s and other have the same kind and expression.
// This is a canonical, literal comparison; it does not evaluate whether
// two differently-worded expressions denote the same scope.
func (s Scope) Equal(other Scope) bool {
	return s.kind.Equal(other.kind) && s.expression == other.expression
}

type scopeJSON struct {
	Kind       VocabularyValue `json:"kind"`
	Expression string          `json:"expression"`
}

// MarshalJSON encodes s as {"kind": ..., "expression": ...}.
func (s Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopeJSON{Kind: s.kind, Expression: s.expression})
}

// UnmarshalJSON decodes s from {"kind": ..., "expression": ...}, applying
// the same validation as NewScope.
func (s *Scope) UnmarshalJSON(data []byte) error {
	var raw scopeJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal Scope: %w", err)
	}
	parsed, err := NewScope(raw.Kind, raw.Expression)
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}
