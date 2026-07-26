package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Assumption is a Decision Basis component (PEOS-004 "Assumption": "An
// Assumption is a proposition treated as true for the purpose of making a
// Decision without sufficient Evidence to establish it as fact." "An
// Assumption used materially by a Decision MUST be explicit." "An
// Assumption SHOULD identify: its scope; its source; its uncertainty; its
// expected validation condition; the consequence if it proves false.").
//
// Every optional field here traces to one item of that SHOULD-identify
// list; nothing else is modeled.
//
// Assumption.Uncertainty is a qualifier of THIS specific assumption --
// "how uncertain is this one assumption" (PEOS-004 :462). It is
// deliberately distinct from the standalone Uncertainty component on
// Basis (Basis.Uncertainties), which records an independent known
// material uncertainty fact (PEOS-004 :542: "Known material uncertainty
// in a Decision Basis MUST be explicit"; :546: "Uncertainty MAY concern:
// ... assumptions"). :462 and :546 are opposite directions of relation --
// one Assumption's own uncertainty qualifier is not the same fact as a
// Basis-level Uncertainty that happens to concern assumptions in general,
// and neither substitutes for the other. Carrying "uncertainty" in both
// places, with no way to tell which Assumption a Basis-level Uncertainty
// concerns, is exactly the kind of two-homes-for-one-fact ambiguity
// InvalidationSource's exclusive-source design (invalidation.go) rejects.
//
// Assumption carries no identity, no Ref, no revision, and no lifecycle:
// PEOS-006's Claim Basis ("does not introduce independent ... identity,
// revision, or lifecycle") and PEOS-007's Quality Constraint ("value
// structures without independent identity, revision, or lifecycle") are
// the governing precedents for every Decision Basis component.
type Assumption struct {
	statement                   string
	scope                       core.Scope
	source                      string
	uncertainty                 string
	expectedValidationCondition string
	consequenceIfFalse          string
	extension                   core.Extension
}

// NewAssumption validates statement and returns an Assumption with none of
// its optional fields set. statement must be non-empty after trimming
// surrounding whitespace; the original value is stored unchanged.
func NewAssumption(statement string) (Assumption, error) {
	if strings.TrimSpace(statement) == "" {
		return Assumption{}, fmt.Errorf("decision: NewAssumption: %w: statement must not be empty", ErrInvalidBasis)
	}
	return Assumption{statement: statement}, nil
}

// WithScope returns a copy of a with its scope set. scope must be
// non-zero. Use WithoutScope to clear a previously set scope.
func (a Assumption) WithScope(scope core.Scope) (Assumption, error) {
	if scope.IsZero() {
		return Assumption{}, fmt.Errorf("decision: Assumption.WithScope: %w: scope must not be zero", ErrInvalidBasis)
	}
	a.scope = scope
	return a, nil
}

// WithoutScope returns a copy of a with its scope cleared.
func (a Assumption) WithoutScope() Assumption {
	a.scope = core.Scope{}
	return a
}

// WithSource returns a copy of a with its source set. source must be
// non-empty after trimming surrounding whitespace; the original value is
// stored unchanged. Use WithoutSource to clear a previously set source.
func (a Assumption) WithSource(source string) (Assumption, error) {
	if strings.TrimSpace(source) == "" {
		return Assumption{}, fmt.Errorf("decision: Assumption.WithSource: %w: source must not be empty", ErrInvalidBasis)
	}
	a.source = source
	return a, nil
}

// WithoutSource returns a copy of a with its source cleared.
func (a Assumption) WithoutSource() Assumption {
	a.source = ""
	return a
}

// WithUncertainty returns a copy of a with its uncertainty qualifier set.
// uncertainty must be non-empty after trimming surrounding whitespace; the
// original value is stored unchanged. Use WithoutUncertainty to clear a
// previously set uncertainty. See Assumption's own doc comment for why
// this is distinct from Basis.Uncertainties.
func (a Assumption) WithUncertainty(uncertainty string) (Assumption, error) {
	if strings.TrimSpace(uncertainty) == "" {
		return Assumption{}, fmt.Errorf("decision: Assumption.WithUncertainty: %w: uncertainty must not be empty", ErrInvalidBasis)
	}
	a.uncertainty = uncertainty
	return a, nil
}

// WithoutUncertainty returns a copy of a with its uncertainty qualifier
// cleared.
func (a Assumption) WithoutUncertainty() Assumption {
	a.uncertainty = ""
	return a
}

// WithExpectedValidationCondition returns a copy of a with its expected
// validation condition set. condition must be non-empty after trimming
// surrounding whitespace; the original value is stored unchanged. Use
// WithoutExpectedValidationCondition to clear a previously set condition.
func (a Assumption) WithExpectedValidationCondition(condition string) (Assumption, error) {
	if strings.TrimSpace(condition) == "" {
		return Assumption{}, fmt.Errorf("decision: Assumption.WithExpectedValidationCondition: %w: expected validation condition must not be empty", ErrInvalidBasis)
	}
	a.expectedValidationCondition = condition
	return a, nil
}

// WithoutExpectedValidationCondition returns a copy of a with its expected
// validation condition cleared.
func (a Assumption) WithoutExpectedValidationCondition() Assumption {
	a.expectedValidationCondition = ""
	return a
}

// WithConsequenceIfFalse returns a copy of a with its consequence-if-false
// set. consequence must be non-empty after trimming surrounding
// whitespace; the original value is stored unchanged. Use
// WithoutConsequenceIfFalse to clear a previously set consequence.
func (a Assumption) WithConsequenceIfFalse(consequence string) (Assumption, error) {
	if strings.TrimSpace(consequence) == "" {
		return Assumption{}, fmt.Errorf("decision: Assumption.WithConsequenceIfFalse: %w: consequence must not be empty", ErrInvalidBasis)
	}
	a.consequenceIfFalse = consequence
	return a, nil
}

// WithoutConsequenceIfFalse returns a copy of a with its
// consequence-if-false cleared.
func (a Assumption) WithoutConsequenceIfFalse() Assumption {
	a.consequenceIfFalse = ""
	return a
}

// WithExtension returns a copy of a with its extension data set.
func (a Assumption) WithExtension(extension core.Extension) Assumption {
	a.extension = extension
	return a
}

func (a Assumption) Statement() string { return a.statement }

// Scope returns a's declared scope, and whether one is set.
func (a Assumption) Scope() (core.Scope, bool) { return a.scope, !a.scope.IsZero() }

// Source returns a's declared source, and whether one is set.
func (a Assumption) Source() (string, bool) { return a.source, a.source != "" }

// Uncertainty returns a's declared uncertainty qualifier, and whether one
// is set. This qualifies THIS assumption only -- see the type's own doc
// comment for the distinction from Basis.Uncertainties.
func (a Assumption) Uncertainty() (string, bool) { return a.uncertainty, a.uncertainty != "" }

// ExpectedValidationCondition returns a's declared expected validation
// condition, and whether one is set.
func (a Assumption) ExpectedValidationCondition() (string, bool) {
	return a.expectedValidationCondition, a.expectedValidationCondition != ""
}

// ConsequenceIfFalse returns a's declared consequence-if-false, and
// whether one is set.
func (a Assumption) ConsequenceIfFalse() (string, bool) {
	return a.consequenceIfFalse, a.consequenceIfFalse != ""
}

func (a Assumption) Extension() core.Extension { return a.extension }

// IsZero reports whether a is the zero value.
func (a Assumption) IsZero() bool { return a.statement == "" }

type assumptionJSON struct {
	Statement                   string          `json:"statement"`
	Scope                       *core.Scope     `json:"scope,omitempty"`
	Source                      string          `json:"source,omitempty"`
	Uncertainty                 string          `json:"uncertainty,omitempty"`
	ExpectedValidationCondition string          `json:"expected_validation_condition,omitempty"`
	ConsequenceIfFalse          string          `json:"consequence_if_false,omitempty"`
	Extension                   *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes a as {"statement":..., "scope":..., "source":...,
// "uncertainty":..., "expected_validation_condition":...,
// "consequence_if_false":..., "extension":...}, omitting every optional
// field that is not set.
func (a Assumption) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return nil, fmt.Errorf("decision: marshal Assumption: %w", ErrInvalidBasis)
	}
	raw := assumptionJSON{Statement: a.statement}
	if !a.scope.IsZero() {
		raw.Scope = &a.scope
	}
	if a.source != "" {
		raw.Source = a.source
	}
	if a.uncertainty != "" {
		raw.Uncertainty = a.uncertainty
	}
	if a.expectedValidationCondition != "" {
		raw.ExpectedValidationCondition = a.expectedValidationCondition
	}
	if a.consequenceIfFalse != "" {
		raw.ConsequenceIfFalse = a.consequenceIfFalse
	}
	if !a.extension.IsZero() {
		raw.Extension = &a.extension
	}
	return json.Marshal(raw)
}

type assumptionUnmarshalJSON struct {
	Statement                   string          `json:"statement"`
	Scope                       json.RawMessage `json:"scope"`
	Source                      json.RawMessage `json:"source"`
	Uncertainty                 json.RawMessage `json:"uncertainty"`
	ExpectedValidationCondition json.RawMessage `json:"expected_validation_condition"`
	ConsequenceIfFalse          json.RawMessage `json:"consequence_if_false"`
	Extension                   json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes a from its JSON form, applying the same validation
// as NewAssumption and every With* method. An explicit JSON null for any
// optional field is rejected rather than silently treated as absent; an
// empty string for an optional field when the key is present is likewise
// rejected.
func (a *Assumption) UnmarshalJSON(data []byte) error {
	var raw assumptionUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
	}

	result, err := NewAssumption(raw.Statement)
	if err != nil {
		return err
	}

	if len(raw.Scope) > 0 {
		if string(raw.Scope) == "null" {
			return fmt.Errorf("decision: unmarshal Assumption: %w: scope must not be null", ErrInvalidBasis)
		}
		var scope core.Scope
		if err := json.Unmarshal(raw.Scope, &scope); err != nil {
			return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
		}
		if result, err = result.WithScope(scope); err != nil {
			return err
		}
	}

	if len(raw.Source) > 0 {
		if string(raw.Source) == "null" {
			return fmt.Errorf("decision: unmarshal Assumption: %w: source must not be null", ErrInvalidBasis)
		}
		var source string
		if err := json.Unmarshal(raw.Source, &source); err != nil {
			return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
		}
		if result, err = result.WithSource(source); err != nil {
			return err
		}
	}

	if len(raw.Uncertainty) > 0 {
		if string(raw.Uncertainty) == "null" {
			return fmt.Errorf("decision: unmarshal Assumption: %w: uncertainty must not be null", ErrInvalidBasis)
		}
		var uncertainty string
		if err := json.Unmarshal(raw.Uncertainty, &uncertainty); err != nil {
			return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
		}
		if result, err = result.WithUncertainty(uncertainty); err != nil {
			return err
		}
	}

	if len(raw.ExpectedValidationCondition) > 0 {
		if string(raw.ExpectedValidationCondition) == "null" {
			return fmt.Errorf("decision: unmarshal Assumption: %w: expected_validation_condition must not be null", ErrInvalidBasis)
		}
		var condition string
		if err := json.Unmarshal(raw.ExpectedValidationCondition, &condition); err != nil {
			return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
		}
		if result, err = result.WithExpectedValidationCondition(condition); err != nil {
			return err
		}
	}

	if len(raw.ConsequenceIfFalse) > 0 {
		if string(raw.ConsequenceIfFalse) == "null" {
			return fmt.Errorf("decision: unmarshal Assumption: %w: consequence_if_false must not be null", ErrInvalidBasis)
		}
		var consequence string
		if err := json.Unmarshal(raw.ConsequenceIfFalse, &consequence); err != nil {
			return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
		}
		if result, err = result.WithConsequenceIfFalse(consequence); err != nil {
			return err
		}
	}

	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Assumption: %w: %w", ErrInvalidBasis, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}

	*a = result
	return nil
}
