package decision

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Basis is the immutable value object recording a Decision's established
// basis (PEOS-004 "Decision Basis": "Decision Basis MUST be
// distinguishable from Decision Outcome"). Basis is optional on Decision:
// PEOS-004's MUST-have list for a Decision never names Basis as
// unconditionally required.
//
// A Basis may consist of any non-empty combination of its four content
// collections -- Evidence, Assumptions, Constraints, and Uncertainties
// (PEOS-004 :406-419's own "Decision Basis MAY include" list). Evidence
// alone, an evidence-free Basis of only Assumptions, or any other
// combination are all conformant; nothing requires Evidence specifically.
// See Assumption, Constraint, and Uncertainty for the normative basis of
// each collection's element type.
//
// None of the four component types carries identity, a Ref, a revision,
// or a lifecycle: PEOS-006's Claim Basis ("does not introduce independent
// Claim Basis identity, revision, or lifecycle") and PEOS-007's Quality
// Constraint ("value structures without independent identity, revision,
// or lifecycle") are the governing precedents for every Decision Basis
// component.
type Basis struct {
	evidence      []core.EvidenceArtifactRevisionRef
	assumptions   []Assumption
	constraints   []Constraint
	uncertainties []Uncertainty
	extension     core.Extension
}

func copyValidEvidence(refs []core.EvidenceArtifactRevisionRef) ([]core.EvidenceArtifactRevisionRef, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(refs))
	for idx, ref := range refs {
		if ref.IsZero() {
			return nil, fmt.Errorf("%w: evidence reference must not be zero", ErrInvalidBasis)
		}
		cp[idx] = ref
	}
	return cp, nil
}

func copyValidAssumptions(items []Assumption) ([]Assumption, error) {
	if len(items) == 0 {
		return nil, nil
	}
	cp := make([]Assumption, len(items))
	for idx, a := range items {
		if a.IsZero() {
			return nil, fmt.Errorf("%w: assumption must not be zero", ErrInvalidBasis)
		}
		cp[idx] = a
	}
	return cp, nil
}

func copyValidConstraints(items []Constraint) ([]Constraint, error) {
	if len(items) == 0 {
		return nil, nil
	}
	cp := make([]Constraint, len(items))
	for idx, c := range items {
		if c.IsZero() {
			return nil, fmt.Errorf("%w: constraint must not be zero", ErrInvalidBasis)
		}
		cp[idx] = c
	}
	return cp, nil
}

func copyValidUncertainties(items []Uncertainty) ([]Uncertainty, error) {
	if len(items) == 0 {
		return nil, nil
	}
	cp := make([]Uncertainty, len(items))
	for idx, u := range items {
		if u.IsZero() {
			return nil, fmt.Errorf("%w: uncertainty must not be zero", ErrInvalidBasis)
		}
		cp[idx] = u
	}
	return cp, nil
}

// newBasisFromParts is the single final-state validator for every public
// entry point that constructs, mutates, or decodes a Basis: NewBasis,
// NewBasisFrom, WithEvidence, WithAssumptions, WithConstraints,
// WithUncertainties, and UnmarshalJSON. It enforces exactly two rules --
// at least one of the four content collections must be non-empty, and
// every element within every collection must be non-zero -- and
// defensively copies all four input slices into the result. A caller
// that receives a nil error from any of those entry points is guaranteed
// a Basis satisfying both rules; there is no path to a staged-invalid
// Basis.
//
// extension is passed through uninspected: it is not a content
// collection and cannot violate either rule. See WithExtension's own doc
// comment for the one pre-existing exception this implies.
func newBasisFromParts(
	evidence []core.EvidenceArtifactRevisionRef,
	assumptions []Assumption,
	constraints []Constraint,
	uncertainties []Uncertainty,
	extension core.Extension,
) (Basis, error) {
	ev, err := copyValidEvidence(evidence)
	if err != nil {
		return Basis{}, fmt.Errorf("decision: %w", err)
	}
	as, err := copyValidAssumptions(assumptions)
	if err != nil {
		return Basis{}, fmt.Errorf("decision: %w", err)
	}
	cs, err := copyValidConstraints(constraints)
	if err != nil {
		return Basis{}, fmt.Errorf("decision: %w", err)
	}
	us, err := copyValidUncertainties(uncertainties)
	if err != nil {
		return Basis{}, fmt.Errorf("decision: %w", err)
	}
	if len(ev) == 0 && len(as) == 0 && len(cs) == 0 && len(us) == 0 {
		return Basis{}, fmt.Errorf("decision: %w: at least one of evidence, assumptions, constraints, or uncertainties is required", ErrInvalidBasis)
	}
	return Basis{evidence: ev, assumptions: as, constraints: cs, uncertainties: us, extension: extension}, nil
}

// NewBasis validates evidence and returns a Basis with no assumptions,
// constraints, uncertainties, or extension data. At least one non-zero
// Evidence reference is required. This is the evidence-grounded
// convenience constructor; use NewBasisFrom to construct a Basis that is
// not evidence-only.
func NewBasis(evidence []core.EvidenceArtifactRevisionRef) (Basis, error) {
	return newBasisFromParts(evidence, nil, nil, nil, core.Extension{})
}

// NewBasisFrom validates evidence, assumptions, constraints, and
// uncertainties and returns a Basis with no extension data. At least one
// of the four collections must be non-empty; every element within every
// non-empty collection must be non-zero. Use NewBasis instead when the
// Basis consists of evidence alone.
func NewBasisFrom(
	evidence []core.EvidenceArtifactRevisionRef,
	assumptions []Assumption,
	constraints []Constraint,
	uncertainties []Uncertainty,
) (Basis, error) {
	return newBasisFromParts(evidence, assumptions, constraints, uncertainties, core.Extension{})
}

// WithEvidence returns a copy of b with its declared evidence references
// set to exactly the values given, in the order given, replacing any
// previous evidence. A zero-value reference among evidence is rejected.
// Calling with no arguments clears the declared evidence, but only when
// at least one of b's other three collections remains non-empty;
// clearing the last remaining content is rejected.
func (b Basis) WithEvidence(evidence ...core.EvidenceArtifactRevisionRef) (Basis, error) {
	return newBasisFromParts(evidence, b.assumptions, b.constraints, b.uncertainties, b.extension)
}

// WithAssumptions returns a copy of b with its declared Assumptions set
// to exactly the values given, in the order given, replacing any
// previous Assumptions. A zero-value Assumption among assumptions is
// rejected. Calling with no arguments clears the declared Assumptions,
// but only when at least one of b's other three collections remains
// non-empty; clearing the last remaining content is rejected.
func (b Basis) WithAssumptions(assumptions ...Assumption) (Basis, error) {
	return newBasisFromParts(b.evidence, assumptions, b.constraints, b.uncertainties, b.extension)
}

// WithConstraints returns a copy of b with its declared Constraints set
// to exactly the values given, in the order given, replacing any
// previous Constraints. A zero-value Constraint among constraints is
// rejected. Calling with no arguments clears the declared Constraints,
// but only when at least one of b's other three collections remains
// non-empty; clearing the last remaining content is rejected.
func (b Basis) WithConstraints(constraints ...Constraint) (Basis, error) {
	return newBasisFromParts(b.evidence, b.assumptions, constraints, b.uncertainties, b.extension)
}

// WithUncertainties returns a copy of b with its declared Uncertainties
// set to exactly the values given, in the order given, replacing any
// previous Uncertainties. A zero-value Uncertainty among uncertainties is
// rejected. Calling with no arguments clears the declared Uncertainties,
// but only when at least one of b's other three collections remains
// non-empty; clearing the last remaining content is rejected.
func (b Basis) WithUncertainties(uncertainties ...Uncertainty) (Basis, error) {
	return newBasisFromParts(b.evidence, b.assumptions, b.constraints, uncertainties, b.extension)
}

// WithExtension returns a copy of b with its extension data set.
//
// This method does not participate in the final-state validity guarantee
// documented on newBasisFromParts: unlike NewBasis, NewBasisFrom, and the
// four collection mutators above, WithExtension has no error return and
// cannot reject a call that would leave b content-less. Calling
// Basis{}.WithExtension(ext) on the zero Basis returns a value that still
// reports IsZero()==true, cannot Marshal, and is rejected by
// Decision.WithBasis: extension alone does not constitute Decision Basis
// content. This is a pre-existing Packet F API edge, preserved unchanged
// by Packet F.2 -- not a defect introduced here, and not silently
// normalized away.
func (b Basis) WithExtension(extension core.Extension) Basis {
	b.extension = extension
	return b
}

// Evidence returns a defensive copy of b's declared evidence references,
// in declaration order.
func (b Basis) Evidence() []core.EvidenceArtifactRevisionRef {
	if len(b.evidence) == 0 {
		return nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(b.evidence))
	copy(cp, b.evidence)
	return cp
}

// Assumptions returns a defensive copy of b's declared Assumptions, in
// declaration order.
func (b Basis) Assumptions() []Assumption {
	if len(b.assumptions) == 0 {
		return nil
	}
	cp := make([]Assumption, len(b.assumptions))
	copy(cp, b.assumptions)
	return cp
}

// Constraints returns a defensive copy of b's declared Constraints, in
// declaration order.
func (b Basis) Constraints() []Constraint {
	if len(b.constraints) == 0 {
		return nil
	}
	cp := make([]Constraint, len(b.constraints))
	copy(cp, b.constraints)
	return cp
}

// Uncertainties returns a defensive copy of b's declared Uncertainties,
// in declaration order.
func (b Basis) Uncertainties() []Uncertainty {
	if len(b.uncertainties) == 0 {
		return nil
	}
	cp := make([]Uncertainty, len(b.uncertainties))
	copy(cp, b.uncertainties)
	return cp
}

func (b Basis) Extension() core.Extension { return b.extension }

// IsZero reports whether b is the zero value: every one of its four
// content collections is empty. extension does not affect IsZero -- see
// WithExtension's own doc comment for why an extension-only value is
// still considered zero.
func (b Basis) IsZero() bool {
	return len(b.evidence) == 0 && len(b.assumptions) == 0 &&
		len(b.constraints) == 0 && len(b.uncertainties) == 0
}

type basisJSON struct {
	Evidence      []core.EvidenceArtifactRevisionRef `json:"evidence,omitempty"`
	Assumptions   []Assumption                       `json:"assumptions,omitempty"`
	Constraints   []Constraint                       `json:"constraints,omitempty"`
	Uncertainties []Uncertainty                      `json:"uncertainties,omitempty"`
	Extension     *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes b as {"evidence":[...], "assumptions":[...],
// "constraints":[...], "uncertainties":[...], "extension":...}, omitting
// every key whose collection is empty or, for extension, unset.
func (b Basis) MarshalJSON() ([]byte, error) {
	if b.IsZero() {
		return nil, fmt.Errorf("decision: marshal Basis: %w", ErrInvalidBasis)
	}
	raw := basisJSON{}
	if len(b.evidence) > 0 {
		raw.Evidence = b.evidence
	}
	if len(b.assumptions) > 0 {
		raw.Assumptions = b.assumptions
	}
	if len(b.constraints) > 0 {
		raw.Constraints = b.constraints
	}
	if len(b.uncertainties) > 0 {
		raw.Uncertainties = b.uncertainties
	}
	if !b.extension.IsZero() {
		raw.Extension = &b.extension
	}
	return json.Marshal(raw)
}

type basisUnmarshalJSON struct {
	Evidence      json.RawMessage `json:"evidence"`
	Assumptions   json.RawMessage `json:"assumptions"`
	Constraints   json.RawMessage `json:"constraints"`
	Uncertainties json.RawMessage `json:"uncertainties"`
	Extension     json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes b from its JSON form, applying the same
// validation as newBasisFromParts. An explicit JSON null for any of the
// five keys is rejected rather than silently treated as absent; an
// absent key or an empty array both decode as no entries for that
// collection. UnmarshalJSON builds local slices for all four collections
// and routes them, in one call, through newBasisFromParts -- there is no
// separate staged validation pass.
func (b *Basis) UnmarshalJSON(data []byte) error {
	var raw basisUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
	}

	var evidence []core.EvidenceArtifactRevisionRef
	if len(raw.Evidence) > 0 {
		if string(raw.Evidence) == "null" {
			return fmt.Errorf("decision: unmarshal Basis: %w: evidence must not be null", ErrInvalidBasis)
		}
		if err := json.Unmarshal(raw.Evidence, &evidence); err != nil {
			return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
		}
	}

	var assumptions []Assumption
	if len(raw.Assumptions) > 0 {
		if string(raw.Assumptions) == "null" {
			return fmt.Errorf("decision: unmarshal Basis: %w: assumptions must not be null", ErrInvalidBasis)
		}
		if err := json.Unmarshal(raw.Assumptions, &assumptions); err != nil {
			return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
		}
	}

	var constraints []Constraint
	if len(raw.Constraints) > 0 {
		if string(raw.Constraints) == "null" {
			return fmt.Errorf("decision: unmarshal Basis: %w: constraints must not be null", ErrInvalidBasis)
		}
		if err := json.Unmarshal(raw.Constraints, &constraints); err != nil {
			return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
		}
	}

	var uncertainties []Uncertainty
	if len(raw.Uncertainties) > 0 {
		if string(raw.Uncertainties) == "null" {
			return fmt.Errorf("decision: unmarshal Basis: %w: uncertainties must not be null", ErrInvalidBasis)
		}
		if err := json.Unmarshal(raw.Uncertainties, &uncertainties); err != nil {
			return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
		}
	}

	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
	}

	result, err := newBasisFromParts(evidence, assumptions, constraints, uncertainties, ext)
	if err != nil {
		return err
	}

	*b = result
	return nil
}
