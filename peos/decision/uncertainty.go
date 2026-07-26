package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Uncertainty is a Decision Basis component (PEOS-004 "Uncertainty":
// "Known material uncertainty in a Decision Basis MUST be explicit.").
//
// Uncertainty carries no concern vocabulary and no severity, significance,
// or threshold field. PEOS-004:544-555's "Uncertainty MAY concern:
// Evidence quality; assumptions; estimates; future conditions;
// implementation feasibility; cost; timing; impact; reversibility; risk"
// is a "MAY concern" list with no accompanying "SHOULD identify" or MUST
// obligation -- unlike Assumption (:458-464) and Constraint (:491),
// PEOS-004 imposes no identify requirement on Uncertainty beyond
// explicitness itself, so this type models nothing beyond the statement
// that makes the uncertainty explicit. PEOS-004:559 assigns uncertainty
// thresholds to "the applicable Product contract", not to this type; a
// Product needing severity or concern classification carries it in
// Extension, which is genuine Product-specific data, not a PEOS-defined
// concept this package declined to model.
//
// Uncertainty is a standalone known-material-uncertainty fact, distinct
// from Assumption's own optional uncertainty qualifier
// (Assumption.Uncertainty) -- see Assumption's own doc comment for the
// two directions of relation.
//
// Uncertainty carries no identity, no Ref, no revision, and no lifecycle
// -- see Assumption's own doc comment for the governing precedents.
type Uncertainty struct {
	statement string
	extension core.Extension
}

// NewUncertainty validates statement and returns an Uncertainty with no
// extension data. statement must be non-empty after trimming surrounding
// whitespace; the original value is stored unchanged.
func NewUncertainty(statement string) (Uncertainty, error) {
	if strings.TrimSpace(statement) == "" {
		return Uncertainty{}, fmt.Errorf("decision: NewUncertainty: %w: statement must not be empty", ErrInvalidBasis)
	}
	return Uncertainty{statement: statement}, nil
}

// WithExtension returns a copy of u with its extension data set.
func (u Uncertainty) WithExtension(extension core.Extension) Uncertainty {
	u.extension = extension
	return u
}

func (u Uncertainty) Statement() string { return u.statement }

func (u Uncertainty) Extension() core.Extension { return u.extension }

// IsZero reports whether u is the zero value.
func (u Uncertainty) IsZero() bool { return u.statement == "" }

type uncertaintyJSON struct {
	Statement string          `json:"statement"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes u as {"statement":..., "extension":...}, omitting
// extension when not set.
func (u Uncertainty) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return nil, fmt.Errorf("decision: marshal Uncertainty: %w", ErrInvalidBasis)
	}
	raw := uncertaintyJSON{Statement: u.statement}
	if !u.extension.IsZero() {
		raw.Extension = &u.extension
	}
	return json.Marshal(raw)
}

type uncertaintyUnmarshalJSON struct {
	Statement string          `json:"statement"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes u from its JSON form, applying the same validation
// as NewUncertainty. An explicit JSON null for "extension" is rejected
// rather than silently treated as absent.
func (u *Uncertainty) UnmarshalJSON(data []byte) error {
	var raw uncertaintyUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Uncertainty: %w: %w", ErrInvalidBasis, err)
	}
	result, err := NewUncertainty(raw.Statement)
	if err != nil {
		return err
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Uncertainty: %w: %w", ErrInvalidBasis, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*u = result
	return nil
}
