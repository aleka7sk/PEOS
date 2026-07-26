package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// mustVocab constructs a core.VocabularyValue under core.PEOSNamespace for
// this package's own predeclared open-vocabulary constants. It panics only
// on a malformed hardcoded literal supplied by this file itself, never on
// caller input -- the same guarantee core's own predeclared vocabulary
// constants have simply by being struct literals inside the core package;
// this package cannot use that literal form itself, because
// core.VocabularyValue's fields are unexported outside core.
func mustVocab(value string) core.VocabularyValue {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, value)
	if err != nil {
		panic(fmt.Sprintf("decision: invalid predeclared vocabulary value %q: %v", value, err))
	}
	return v
}

// OutcomeKind classifies the nature of a Decision Outcome (PEOS-004: an
// Outcome MAY "select an option; reject a proposal; defer a choice;
// authorize an action; prohibit an action; constrain future choices;
// delegate the decision; accept a risk; require further investigation;
// establish, change, or remove an Engineering Commitment; or leave prior
// Engineering Commitments unchanged"). OutcomeKind is an open vocabulary:
// a Product MAY declare additional kind values beyond the thirteen
// predeclared below.
type OutcomeKind struct{ value core.VocabularyValue }

// NewOutcomeKind wraps v as an OutcomeKind.
func NewOutcomeKind(v core.VocabularyValue) OutcomeKind { return OutcomeKind{value: v} }

func (k OutcomeKind) Value() core.VocabularyValue { return k.value }
func (k OutcomeKind) IsZero() bool                { return k.value.IsZero() }
func (k OutcomeKind) String() string              { return k.value.String() }

// Equal reports whether k and other name the same kind value.
func (k OutcomeKind) Equal(other OutcomeKind) bool { return k.value.Equal(other.value) }

// MarshalJSON encodes k as its canonical "namespace:value" string form. A
// zero OutcomeKind is rejected: it is never legitimate to marshal on its
// own, only as an optional field omitted entirely.
func (k OutcomeKind) MarshalJSON() ([]byte, error) {
	if k.IsZero() {
		return nil, fmt.Errorf("decision: marshal OutcomeKind: %w", ErrInvalidOutcome)
	}
	return json.Marshal(k.value)
}

func (k *OutcomeKind) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &k.value)
}

var (
	OutcomeKindSelect               = OutcomeKind{value: mustVocab("select")}
	OutcomeKindReject               = OutcomeKind{value: mustVocab("reject")}
	OutcomeKindDefer                = OutcomeKind{value: mustVocab("defer")}
	OutcomeKindAuthorize            = OutcomeKind{value: mustVocab("authorize")}
	OutcomeKindProhibit             = OutcomeKind{value: mustVocab("prohibit")}
	OutcomeKindConstrain            = OutcomeKind{value: mustVocab("constrain")}
	OutcomeKindDelegate             = OutcomeKind{value: mustVocab("delegate")}
	OutcomeKindAcceptRisk           = OutcomeKind{value: mustVocab("accept-risk")}
	OutcomeKindRequireInvestigation = OutcomeKind{value: mustVocab("require-investigation")}
	OutcomeKindEstablishCommitment  = OutcomeKind{value: mustVocab("establish-commitment")}
	OutcomeKindChangeCommitment     = OutcomeKind{value: mustVocab("change-commitment")}
	OutcomeKindRemoveCommitment     = OutcomeKind{value: mustVocab("remove-commitment")}
	OutcomeKindNoChangeRequired     = OutcomeKind{value: mustVocab("no-change-required")}
)

// CommitmentEffect names the exhaustive discriminator PEOS-004 requires
// every Decision to make clear about its effect on Engineering Commitment:
// "A Decision MUST make clear whether it: establishes; changes; removes;
// rejects; defers; or leaves unchanged an Engineering Commitment."
// CommitmentEffect is an open vocabulary: a Product MAY declare additional
// values, though PEOS-004 itself names exactly these six.
type CommitmentEffect struct{ value core.VocabularyValue }

// NewCommitmentEffect wraps v as a CommitmentEffect.
func NewCommitmentEffect(v core.VocabularyValue) CommitmentEffect { return CommitmentEffect{value: v} }

func (e CommitmentEffect) Value() core.VocabularyValue { return e.value }
func (e CommitmentEffect) IsZero() bool                { return e.value.IsZero() }
func (e CommitmentEffect) String() string              { return e.value.String() }

// Equal reports whether e and other name the same effect value.
func (e CommitmentEffect) Equal(other CommitmentEffect) bool { return e.value.Equal(other.value) }

// MarshalJSON encodes e as its canonical "namespace:value" string form. A
// zero CommitmentEffect is rejected: Outcome requires a non-zero
// CommitmentEffect unconditionally, so a zero value should never reach
// marshaling.
func (e CommitmentEffect) MarshalJSON() ([]byte, error) {
	if e.IsZero() {
		return nil, fmt.Errorf("decision: marshal CommitmentEffect: %w", ErrInvalidCommitmentEffect)
	}
	return json.Marshal(e.value)
}

func (e *CommitmentEffect) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &e.value)
}

var (
	CommitmentEffectEstablishes = CommitmentEffect{value: mustVocab("establishes")}
	CommitmentEffectChanges     = CommitmentEffect{value: mustVocab("changes")}
	CommitmentEffectRemoves     = CommitmentEffect{value: mustVocab("removes")}
	CommitmentEffectRejects     = CommitmentEffect{value: mustVocab("rejects")}
	CommitmentEffectDefers      = CommitmentEffect{value: mustVocab("defers")}
	CommitmentEffectUnchanged   = CommitmentEffect{value: mustVocab("unchanged")}
)

// Commitment is an Engineering Commitment (PEOS-004): a semantic component
// of a Decision Outcome, not an independent PEOS Entity. It carries no
// identity of its own -- core.EngineeringCommitmentRef is derived from the
// owning Decision's identity, outside this value.
type Commitment struct {
	statement string
	extension core.Extension
}

// NewCommitment validates statement and returns a Commitment.
func NewCommitment(statement string) (Commitment, error) {
	if strings.TrimSpace(statement) == "" {
		return Commitment{}, fmt.Errorf("decision: NewCommitment: %w: statement must not be empty", ErrInvalidCommitment)
	}
	return Commitment{statement: statement}, nil
}

// WithExtension returns a copy of c with its extension data set.
func (c Commitment) WithExtension(e core.Extension) Commitment {
	c.extension = e
	return c
}

func (c Commitment) Statement() string         { return c.statement }
func (c Commitment) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c Commitment) IsZero() bool { return c.statement == "" }

type commitmentJSON struct {
	Statement string          `json:"statement"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"statement":..., "extension":...}, omitting
// extension when not set.
func (c Commitment) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("decision: marshal Commitment: %w", ErrInvalidCommitment)
	}
	raw := commitmentJSON{Statement: c.statement}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

type commitmentUnmarshalJSON struct {
	Statement string          `json:"statement"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as NewCommitment. An explicit JSON null for "extension" is rejected.
func (c *Commitment) UnmarshalJSON(data []byte) error {
	var raw commitmentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Commitment: %w", err)
	}
	result, err := NewCommitment(raw.Statement)
	if err != nil {
		return err
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Commitment: %w: %w", ErrInvalidCommitment, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*c = result
	return nil
}

// Outcome is a Decision Outcome (PEOS-004): "The Decision Outcome indicates
// what was decided." Statement carries the decided content, in prose --
// including, where applicable, which Alternative was selected (PEOS-004's
// own worked example expresses selection this way, never as a structural
// reference to an Alternative; see Alternative's own doc comment).
// CommitmentEffect is the one exhaustive discriminator PEOS-004
// unconditionally requires; Kind is an additional, optional classification.
type Outcome struct {
	statement        string
	kind             OutcomeKind
	commitmentEffect CommitmentEffect
	commitments      []Commitment
	extension        core.Extension
}

// NewOutcome validates statement and effect and returns an Outcome with no
// Kind, Commitments, or extension data set.
func NewOutcome(statement string, effect CommitmentEffect) (Outcome, error) {
	if strings.TrimSpace(statement) == "" {
		return Outcome{}, fmt.Errorf("decision: NewOutcome: %w: statement must not be empty", ErrInvalidOutcome)
	}
	if effect.IsZero() {
		return Outcome{}, fmt.Errorf("decision: NewOutcome: %w: commitment effect must not be zero", ErrInvalidCommitmentEffect)
	}
	return Outcome{statement: statement, commitmentEffect: effect}, nil
}

// WithKind returns a copy of o with its Kind set. kind must be non-zero.
func (o Outcome) WithKind(kind OutcomeKind) (Outcome, error) {
	if kind.IsZero() {
		return Outcome{}, fmt.Errorf("decision: Outcome.WithKind: %w: kind must not be zero", ErrInvalidOutcome)
	}
	o.kind = kind
	return o, nil
}

// WithoutKind returns a copy of o with its Kind cleared.
func (o Outcome) WithoutKind() Outcome {
	o.kind = OutcomeKind{}
	return o
}

// WithCommitments returns a copy of o with its declared Commitments set to
// exactly the values given, in the order given, replacing any previous
// Commitments. A zero-value Commitment among commitments is rejected.
// Calling with no arguments clears the declared Commitments.
func (o Outcome) WithCommitments(commitments ...Commitment) (Outcome, error) {
	if len(commitments) == 0 {
		o.commitments = nil
		return o, nil
	}
	cp := make([]Commitment, len(commitments))
	for idx, c := range commitments {
		if c.IsZero() {
			return Outcome{}, fmt.Errorf("decision: Outcome.WithCommitments: %w", ErrInvalidCommitment)
		}
		cp[idx] = c
	}
	o.commitments = cp
	return o, nil
}

// WithExtension returns a copy of o with its extension data set.
func (o Outcome) WithExtension(e core.Extension) Outcome {
	o.extension = e
	return o
}

func (o Outcome) Statement() string { return o.statement }

// Kind returns o's declared Kind, and whether one is set.
func (o Outcome) Kind() (OutcomeKind, bool) { return o.kind, !o.kind.IsZero() }

func (o Outcome) CommitmentEffect() CommitmentEffect { return o.commitmentEffect }

// Commitments returns a defensive copy of o's declared Commitments, in
// declaration order.
func (o Outcome) Commitments() []Commitment {
	if len(o.commitments) == 0 {
		return nil
	}
	cp := make([]Commitment, len(o.commitments))
	copy(cp, o.commitments)
	return cp
}

func (o Outcome) Extension() core.Extension { return o.extension }

// IsZero reports whether o is the zero value.
func (o Outcome) IsZero() bool { return o.statement == "" && o.commitmentEffect.IsZero() }

type outcomeJSON struct {
	Statement        string           `json:"statement"`
	Kind             *OutcomeKind     `json:"kind,omitempty"`
	CommitmentEffect CommitmentEffect `json:"commitment_effect"`
	Commitments      []Commitment     `json:"commitments,omitempty"`
	Extension        *core.Extension  `json:"extension,omitempty"`
}

// MarshalJSON encodes o as {"statement":..., "commitment_effect":..., ...},
// omitting kind, commitments, and extension when not set.
func (o Outcome) MarshalJSON() ([]byte, error) {
	if o.IsZero() {
		return nil, fmt.Errorf("decision: marshal Outcome: %w", ErrInvalidOutcome)
	}
	raw := outcomeJSON{Statement: o.statement, CommitmentEffect: o.commitmentEffect}
	if !o.kind.IsZero() {
		raw.Kind = &o.kind
	}
	if len(o.commitments) > 0 {
		raw.Commitments = o.commitments
	}
	if !o.extension.IsZero() {
		raw.Extension = &o.extension
	}
	return json.Marshal(raw)
}

type outcomeUnmarshalJSON struct {
	Statement        string           `json:"statement"`
	Kind             json.RawMessage  `json:"kind"`
	CommitmentEffect CommitmentEffect `json:"commitment_effect"`
	Commitments      json.RawMessage  `json:"commitments"`
	Extension        json.RawMessage  `json:"extension"`
}

// UnmarshalJSON decodes o from its JSON form, applying the same validation
// as NewOutcome, WithKind, and WithCommitments. An explicit JSON null for
// "kind", "commitments", or "extension" is rejected rather than silently
// treated as absent.
func (o *Outcome) UnmarshalJSON(data []byte) error {
	var raw outcomeUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Outcome: %w", err)
	}
	result, err := NewOutcome(raw.Statement, raw.CommitmentEffect)
	if err != nil {
		return err
	}
	if len(raw.Kind) > 0 {
		if string(raw.Kind) == "null" {
			return fmt.Errorf("decision: unmarshal Outcome: %w: kind must not be null", ErrInvalidOutcome)
		}
		var kind OutcomeKind
		if err := json.Unmarshal(raw.Kind, &kind); err != nil {
			return fmt.Errorf("decision: unmarshal Outcome: %w", err)
		}
		if result, err = result.WithKind(kind); err != nil {
			return err
		}
	}
	if len(raw.Commitments) > 0 {
		if string(raw.Commitments) == "null" {
			return fmt.Errorf("decision: unmarshal Outcome: %w: commitments must not be null", ErrInvalidOutcome)
		}
		var commitments []Commitment
		if err := json.Unmarshal(raw.Commitments, &commitments); err != nil {
			return fmt.Errorf("decision: unmarshal Outcome: %w", err)
		}
		if result, err = result.WithCommitments(commitments...); err != nil {
			return err
		}
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Outcome: %w: %w", ErrInvalidOutcome, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*o = result
	return nil
}
