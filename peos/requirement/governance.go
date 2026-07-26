package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

type governanceActionKind string

const (
	governanceActionDecisionOutcome  governanceActionKind = "decision_outcome"
	governanceActionProductMechanism governanceActionKind = "product_mechanism"
)

// GovernanceAction identifies the governance action under which a
// Requirement Supersession's replacement was established (PEOS-005 §23:
// "a governance action is an established Decision Outcome or another
// governance mechanism explicitly permitted by an applicable PEOS
// Product contract"). §27 defines Waiver's authorizing action in
// identical terms, so this type is shared foundation for both, not
// Supersession-specific -- Packet G.5 reuses it directly.
//
// §23's own words -- "Governance action is a semantic role and SHALL NOT
// be interpreted as introducing a separate PEOS entity" -- are why
// GovernanceAction carries no identity, no lifecycle, and no Ref type: it
// is a plain immutable value, exactly like decision.InvalidationSource
// (the established closed two-arm union precedent in this repository),
// which it otherwise mirrors in shape. Unlike RequirementParticipant, it
// does carry JSON: it is a type-specific field stored alongside, not
// inside, a composed relation.Relation.
//
// The Product-mechanism arm is carried as core.VocabularyValue, the
// package-wide idiom for an open Product-defined identifier (see
// core.NewProductRuleRef and core.NewExternalRuleRef for the same shape
// used elsewhere for "a mechanism a Product contract permits").
type GovernanceAction struct {
	kind             governanceActionKind
	decisionOutcome  core.DecisionOutcomeRef
	productMechanism core.VocabularyValue
}

// NewGovernanceActionFromDecisionOutcome validates ref and returns a
// GovernanceAction naming the authorizing Decision Outcome.
func NewGovernanceActionFromDecisionOutcome(ref core.DecisionOutcomeRef) (GovernanceAction, error) {
	if ref.IsZero() {
		return GovernanceAction{}, fmt.Errorf("requirement: NewGovernanceActionFromDecisionOutcome: %w: decision outcome must not be zero", ErrInvalidGovernanceAction)
	}
	return GovernanceAction{kind: governanceActionDecisionOutcome, decisionOutcome: ref}, nil
}

// NewGovernanceActionFromProductMechanism validates value and returns a
// GovernanceAction naming a Product-contract-defined governance
// mechanism.
func NewGovernanceActionFromProductMechanism(value core.VocabularyValue) (GovernanceAction, error) {
	if value.IsZero() {
		return GovernanceAction{}, fmt.Errorf("requirement: NewGovernanceActionFromProductMechanism: %w: product mechanism must not be zero", ErrInvalidGovernanceAction)
	}
	return GovernanceAction{kind: governanceActionProductMechanism, productMechanism: value}, nil
}

// Kind returns g's discriminator, "decision_outcome" or
// "product_mechanism".
func (g GovernanceAction) Kind() string { return string(g.kind) }

// IsDecisionOutcome reports whether g names an authorizing Decision
// Outcome.
func (g GovernanceAction) IsDecisionOutcome() bool {
	return g.kind == governanceActionDecisionOutcome
}

// IsProductMechanism reports whether g names a Product-contract-defined
// governance mechanism.
func (g GovernanceAction) IsProductMechanism() bool {
	return g.kind == governanceActionProductMechanism
}

// AsDecisionOutcome returns g's Decision Outcome reference, and whether g
// is the Decision Outcome arm.
func (g GovernanceAction) AsDecisionOutcome() (core.DecisionOutcomeRef, bool) {
	if g.kind != governanceActionDecisionOutcome {
		return core.DecisionOutcomeRef{}, false
	}
	return g.decisionOutcome, true
}

// AsProductMechanism returns g's Product mechanism value, and whether g
// is the Product mechanism arm.
func (g GovernanceAction) AsProductMechanism() (core.VocabularyValue, bool) {
	if g.kind != governanceActionProductMechanism {
		return core.VocabularyValue{}, false
	}
	return g.productMechanism, true
}

// IsZero reports whether g is the zero value.
func (g GovernanceAction) IsZero() bool { return g.kind == "" }

type governanceActionJSON struct {
	Kind governanceActionKind `json:"kind"`
	Ref  json.RawMessage      `json:"ref"`
}

// MarshalJSON encodes g as {"kind":"decision_outcome","ref":{...}} or
// {"kind":"product_mechanism","ref":"namespace:value"}.
func (g GovernanceAction) MarshalJSON() ([]byte, error) {
	if g.IsZero() {
		return nil, fmt.Errorf("requirement: marshal GovernanceAction: %w", ErrInvalidGovernanceAction)
	}
	var (
		refBytes []byte
		err      error
	)
	switch g.kind {
	case governanceActionDecisionOutcome:
		refBytes, err = json.Marshal(g.decisionOutcome)
	case governanceActionProductMechanism:
		refBytes, err = json.Marshal(g.productMechanism)
	default:
		return nil, fmt.Errorf("requirement: marshal GovernanceAction: %w", ErrInvalidGovernanceAction)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(governanceActionJSON{Kind: g.kind, Ref: refBytes})
}

// UnmarshalJSON decodes g from its {"kind":...,"ref":...} JSON form,
// applying the same validation as NewGovernanceActionFromDecisionOutcome
// and NewGovernanceActionFromProductMechanism. An unrecognized or missing
// "kind", a missing "ref", or an explicit null "ref" are all rejected.
func (g *GovernanceAction) UnmarshalJSON(data []byte) error {
	var env governanceActionJSON
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("requirement: unmarshal GovernanceAction: %w: %w", ErrInvalidGovernanceAction, err)
	}
	if len(env.Ref) == 0 {
		return fmt.Errorf("requirement: unmarshal GovernanceAction: %w: ref is required", ErrInvalidGovernanceAction)
	}
	if string(env.Ref) == "null" {
		return fmt.Errorf("requirement: unmarshal GovernanceAction: %w: ref must not be null", ErrInvalidGovernanceAction)
	}

	var (
		result GovernanceAction
		err    error
	)
	switch env.Kind {
	case governanceActionDecisionOutcome:
		var ref core.DecisionOutcomeRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = NewGovernanceActionFromDecisionOutcome(ref)
		}
	case governanceActionProductMechanism:
		var value core.VocabularyValue
		if err = json.Unmarshal(env.Ref, &value); err == nil {
			result, err = NewGovernanceActionFromProductMechanism(value)
		}
	default:
		return fmt.Errorf("requirement: unmarshal GovernanceAction: unrecognized kind %q: %w", env.Kind, ErrInvalidGovernanceAction)
	}
	if err != nil {
		return fmt.Errorf("requirement: unmarshal GovernanceAction: %w: %w", ErrInvalidGovernanceAction, err)
	}
	*g = result
	return nil
}
