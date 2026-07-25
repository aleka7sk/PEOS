package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// SubjectCombination makes explicit how a Requirement's declared
// Subjects, taken together, relate to its required engineering intent
// (PEOS-005 §10: "The represented required engineering intent SHALL make
// clear whether it applies independently to each identified Subject;
// collectively to the identified Subjects; or according to another
// explicitly represented relationship among those Subjects." — see also
// non-conforming pattern §36.22, "Subject Cardinality Ambiguity"). It is
// an open vocabulary, not a closed Go enum: PEOS-005 leaves "another
// explicitly represented relationship" open-ended, so a Product MAY
// declare additional combination values beyond independent/collective.
//
// One SubjectCombination value applies to the entire ordered Subject
// sequence in a Content value. Packet C does not implement boolean
// subject expressions, per-subject operators, nested subject groups, or a
// relationship graph between individual Subjects — SubjectCombination
// names one whole-sequence relationship, nothing finer-grained.
type SubjectCombination struct{ value core.VocabularyValue }

// NewSubjectCombination validates value and returns a SubjectCombination.
func NewSubjectCombination(value core.VocabularyValue) (SubjectCombination, error) {
	if value.IsZero() {
		return SubjectCombination{}, fmt.Errorf("requirement: NewSubjectCombination: %w", ErrInvalidSubjectCombination)
	}
	return SubjectCombination{value: value}, nil
}

// Value returns the underlying vocabulary value.
func (s SubjectCombination) Value() core.VocabularyValue { return s.value }

// IsZero reports whether s is the zero value.
func (s SubjectCombination) IsZero() bool { return s.value.IsZero() }

func mustSubjectCombination(value string) SubjectCombination {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, value)
	if err != nil {
		panic(err)
	}
	sc, err := NewSubjectCombination(v)
	if err != nil {
		panic(err)
	}
	return sc
}

var (
	// SubjectCombinationIndependent marks a Requirement's required
	// engineering intent as applying independently to each identified
	// Subject.
	SubjectCombinationIndependent = mustSubjectCombination("independent")

	// SubjectCombinationCollective marks a Requirement's required
	// engineering intent as applying collectively to the identified
	// Subjects taken together.
	SubjectCombinationCollective = mustSubjectCombination("collective")
)

// MarshalJSON encodes s as its canonical "namespace:value" string form.
func (s SubjectCombination) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("requirement: marshal SubjectCombination: %w", ErrInvalidSubjectCombination)
	}
	return json.Marshal(s.value)
}

// UnmarshalJSON decodes s from its canonical "namespace:value" string
// form, applying the same validation as NewSubjectCombination.
func (s *SubjectCombination) UnmarshalJSON(data []byte) error {
	var v core.VocabularyValue
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("requirement: unmarshal SubjectCombination: %w", err)
	}
	result, err := NewSubjectCombination(v)
	if err != nil {
		return err
	}
	*s = result
	return nil
}
