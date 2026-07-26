package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Consequence is an expected, required, permitted, or accepted effect of
// a Decision Outcome (PEOS-004 "Decision Consequence": "A Decision
// Consequence is an expected, required, permitted, or accepted effect of
// a Decision Outcome."). Consequence is modelled at SHOULD strength, not
// MUST: PEOS-004 :316 says only that a significant Decision "SHOULD have
// ... identified consequences", and :874 says a Decision Record "SHOULD
// identify ... Consequences" -- Consequence is absent from both the
// Conformance list (:1431-1443) and the fourteen Decision Invariants
// (:1284-1342), unlike Basis's Assumption and Uncertainty, which the
// SDK models at :456/:542's conditional MUST strength (see Basis's own
// doc comment, basis.go). It is nonetheless modelled, at that lower
// strength, for the same reason Alternative (SHOULD, :353) and rationale
// (MAY, :518) are modelled: it is a named PEOS-004 "# Scope" entry
// (:54) with its own dedicated section, not illustrative prose.
//
// :697's "Consequences MUST be distinguishable from completed effects"
// is not a requirement that Consequence be separately representable; it
// is a distinguishability constraint against a completion concept this
// package does not model at all, so the MUST holds vacuously today, and
// continues to hold structurally: Consequence carries no completion or
// status field, so a Consequence value can never be mistaken for a
// completed effect.
//
// Consequence is not Commitment (outcome.go), despite the two "MAY
// include" lists at :620 and :683 sharing several examples nearly
// verbatim (required Lifecycle Transitions / require a Lifecycle
// Transition; accepted risks / accept a defined risk; review obligations
// / impose a review condition; validation work / establish a validation
// expectation; operational changes / establish an operational
// obligation; implementation work / establish an implementation
// obligation; new constraints / constrain future Artifacts or
// Revisions). That overlap is of illustrative examples, not of
// definitions: :616 defines Engineering Commitment as normative intent
// established, changed, or removed by an established Decision Outcome;
// :681 defines Decision Consequence as an expected, required, permitted,
// or accepted effect of that Outcome. :681 explicitly admits predictions
// -- expected effects that are not themselves normative intent (for
// example, expected migration work or expected operational impact).
// Encoding such a prediction as a Commitment would misrepresent it as
// established normative intent and would violate PEOS-004 :1413's
// prohibition on extensions that "redefine the core meaning of ...
// Engineering Commitment." Consequence therefore carries no
// CommitmentEffect-equivalent verb, unlike Commitment.
//
// Consequence hangs off Decision, not Outcome: PEOS-004 :316 and :701
// ("A Decision MAY be applicable even when its Consequences have not yet
// been completed") both frame Consequence at Decision level, and keeping
// it off Outcome avoids normative intent (Commitment) and expected
// effect (Consequence) sharing one type.
//
// Consequence carries no identity, no Ref, no revision, and no
// lifecycle -- see Assumption's own doc comment (assumption.go) for the
// governing precedents (PEOS-006, PEOS-007) already applied throughout
// this package to value structures of this shape.
type Consequence struct {
	statement string
	extension core.Extension
}

// NewConsequence validates statement and returns a Consequence with no
// extension data. statement must be non-empty after trimming surrounding
// whitespace; the original value is stored unchanged.
func NewConsequence(statement string) (Consequence, error) {
	if strings.TrimSpace(statement) == "" {
		return Consequence{}, fmt.Errorf("decision: NewConsequence: %w: statement must not be empty", ErrInvalidConsequence)
	}
	return Consequence{statement: statement}, nil
}

// WithExtension returns a copy of c with its extension data set.
func (c Consequence) WithExtension(extension core.Extension) Consequence {
	c.extension = extension
	return c
}

func (c Consequence) Statement() string         { return c.statement }
func (c Consequence) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c Consequence) IsZero() bool { return c.statement == "" }

type consequenceJSON struct {
	Statement string          `json:"statement"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"statement":..., "extension":...}, omitting
// extension when not set.
func (c Consequence) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("decision: marshal Consequence: %w", ErrInvalidConsequence)
	}
	raw := consequenceJSON{Statement: c.statement}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

type consequenceUnmarshalJSON struct {
	Statement string          `json:"statement"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same
// validation as NewConsequence. An explicit JSON null for "extension" is
// rejected rather than silently treated as absent.
func (c *Consequence) UnmarshalJSON(data []byte) error {
	var raw consequenceUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Consequence: %w: %w", ErrInvalidConsequence, err)
	}
	result, err := NewConsequence(raw.Statement)
	if err != nil {
		return err
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Consequence: %w: %w", ErrInvalidConsequence, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*c = result
	return nil
}
