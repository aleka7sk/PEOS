package requirement

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Statement is the human-readable expression of required engineering
// intent represented by a Requirement (PEOS-005 §9). It is descriptive
// content, not an engineering identity: "The Statement SHALL NOT be
// interpreted as the Requirement itself" (§9.1).
//
// Statement is deliberately minimal: text is preserved verbatim, with
// only surrounding whitespace trimmed. No language tag, grammar, or
// parser is introduced. PEOS-005 §9.2 discusses internal consistency,
// ambiguity, and singularity of intent as qualities a Statement SHOULD
// exhibit, never as constructor-enforceable rules ("This specification
// does not define a complete Requirement quality taxonomy or a mandatory
// Requirement quality assessment method" — §16), so this type does not
// attempt to enforce them. Use of the word "shall," terminal punctuation,
// and duplicate Statement values within one Content are not restricted by
// PEOS-005 and are not restricted here.
type Statement struct {
	text string
}

// NewStatement validates text and returns a Statement. Surrounding
// whitespace is trimmed; text must be non-empty after trimming. Internal
// whitespace, multiline text, and punctuation are preserved exactly as
// given.
func NewStatement(text string) (Statement, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Statement{}, fmt.Errorf("requirement: NewStatement: %w", ErrInvalidStatement)
	}
	return Statement{text: trimmed}, nil
}

// Text returns the Statement's text.
func (s Statement) Text() string { return s.text }

// IsZero reports whether s is the zero value.
func (s Statement) IsZero() bool { return s.text == "" }

type statementJSON struct {
	Text string `json:"text"`
}

// MarshalJSON encodes s as {"text": ...}.
func (s Statement) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Statement: %w", ErrInvalidStatement)
	}
	return json.Marshal(statementJSON{Text: s.text})
}

// UnmarshalJSON decodes s from its JSON form, applying the same
// validation as NewStatement.
func (s *Statement) UnmarshalJSON(data []byte) error {
	var raw statementJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Statement: %w", err)
	}
	result, err := NewStatement(raw.Text)
	if err != nil {
		return err
	}
	*s = result
	return nil
}
