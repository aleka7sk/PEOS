package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Constraint is a Decision Basis component (PEOS-004 "Constraint": "A
// Constraint is a normative or practical limitation that restricts the
// available Alternatives or Decision Outcome." "A material Constraint
// SHOULD identify its normative source.").
//
// Constraint carries no origin-category vocabulary. PEOS-004:478-489's "A
// Constraint MAY originate from: a Requirement; a Decision; a contract; a
// regulation; a policy; a standard; available resources; technology
// limitations; time limitations; safety or security conditions" is
// illustrative "MAY originate from" prose, structurally identical to
// PEOS-004:737-749's authority-origin list -- which this package already
// declined to canonize (core.AuthorityRef's Kind is an open vocabulary
// with zero predeclared constants). :491 asks a material Constraint to
// identify its normative source, not the category that source falls
// into; Source is free text for exactly that reason -- a Requirement
// identifier, a contract clause, a regulation citation, a policy or
// standard name, or equivalent identifying text.
//
// Constraint is Decision-Basis-scoped: it restricts one Decision's own
// Alternatives or Outcome (:476). It is deliberately not reused for a
// future Delegation's "applicable constraints" (PEOS-004:833): a
// delegation's constraints limit a granted authority across all of that
// grantee's future Decisions -- a different axis, the same distinction
// this package already draws between Authority.requirements and
// Authority.bases (authority.go) and between requirement.OriginRef and
// core.Origin (a Revision's origin qualification).
//
// Constraint carries no identity, no Ref, no revision, and no lifecycle
// -- see Assumption's own doc comment for the governing precedents.
type Constraint struct {
	statement string
	source    string
	extension core.Extension
}

// NewConstraint validates statement and returns a Constraint with no
// source or extension data. statement must be non-empty after trimming
// surrounding whitespace; the original value is stored unchanged.
func NewConstraint(statement string) (Constraint, error) {
	if strings.TrimSpace(statement) == "" {
		return Constraint{}, fmt.Errorf("decision: NewConstraint: %w: statement must not be empty", ErrInvalidBasis)
	}
	return Constraint{statement: statement}, nil
}

// WithSource returns a copy of c with its normative source set. source
// must be non-empty after trimming surrounding whitespace; the original
// value is stored unchanged. Use WithoutSource to clear a previously set
// source.
func (c Constraint) WithSource(source string) (Constraint, error) {
	if strings.TrimSpace(source) == "" {
		return Constraint{}, fmt.Errorf("decision: Constraint.WithSource: %w: source must not be empty", ErrInvalidBasis)
	}
	c.source = source
	return c, nil
}

// WithoutSource returns a copy of c with its source cleared.
func (c Constraint) WithoutSource() Constraint {
	c.source = ""
	return c
}

// WithExtension returns a copy of c with its extension data set.
func (c Constraint) WithExtension(extension core.Extension) Constraint {
	c.extension = extension
	return c
}

func (c Constraint) Statement() string { return c.statement }

// Source returns c's declared normative source, and whether one is set.
func (c Constraint) Source() (string, bool) { return c.source, c.source != "" }

func (c Constraint) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c Constraint) IsZero() bool { return c.statement == "" }

type constraintJSON struct {
	Statement string          `json:"statement"`
	Source    string          `json:"source,omitempty"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"statement":..., "source":..., "extension":...},
// omitting source and extension when not set.
func (c Constraint) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("decision: marshal Constraint: %w", ErrInvalidBasis)
	}
	raw := constraintJSON{Statement: c.statement}
	if c.source != "" {
		raw.Source = c.source
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

type constraintUnmarshalJSON struct {
	Statement string          `json:"statement"`
	Source    json.RawMessage `json:"source"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as NewConstraint and WithSource. An explicit JSON null for "source" or
// "extension" is rejected rather than silently treated as absent.
func (c *Constraint) UnmarshalJSON(data []byte) error {
	var raw constraintUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Constraint: %w: %w", ErrInvalidBasis, err)
	}
	result, err := NewConstraint(raw.Statement)
	if err != nil {
		return err
	}
	if len(raw.Source) > 0 {
		if string(raw.Source) == "null" {
			return fmt.Errorf("decision: unmarshal Constraint: %w: source must not be null", ErrInvalidBasis)
		}
		var source string
		if err := json.Unmarshal(raw.Source, &source); err != nil {
			return fmt.Errorf("decision: unmarshal Constraint: %w: %w", ErrInvalidBasis, err)
		}
		if result, err = result.WithSource(source); err != nil {
			return err
		}
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Constraint: %w: %w", ErrInvalidBasis, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*c = result
	return nil
}
