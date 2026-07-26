package decision

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Alternative is a considered option within a Decision's deliberation
// (PEOS-004 "Alternative": a Decision MAY document one or more Alternatives
// that were considered; "A Decision is not required to enumerate every
// theoretically possible Alternative"). Alternative carries no identifier:
// PEOS-004 defines no Alternative identity, and its own worked example
// (Decision Outcome: "PostgreSQL is selected as the transactional system of
// record") expresses the selected option as prose within the Outcome
// statement, never as a structural reference to a specific Alternative --
// Outcome therefore carries no selected-alternative link either.
type Alternative struct {
	statement string
	note      string
	extension core.Extension
}

// NewAlternative validates statement and returns an Alternative with no
// note set.
func NewAlternative(statement string) (Alternative, error) {
	if strings.TrimSpace(statement) == "" {
		return Alternative{}, fmt.Errorf("decision: NewAlternative: %w: statement must not be empty", ErrInvalidAlternative)
	}
	return Alternative{statement: statement}, nil
}

// WithNote returns a copy of a with its note set. note must be non-empty.
func (a Alternative) WithNote(note string) (Alternative, error) {
	if strings.TrimSpace(note) == "" {
		return Alternative{}, fmt.Errorf("decision: Alternative.WithNote: %w: note must not be empty", ErrInvalidAlternative)
	}
	a.note = note
	return a, nil
}

// WithoutNote returns a copy of a with its note cleared.
func (a Alternative) WithoutNote() Alternative {
	a.note = ""
	return a
}

// WithExtension returns a copy of a with its extension data set.
func (a Alternative) WithExtension(e core.Extension) Alternative {
	a.extension = e
	return a
}

func (a Alternative) Statement() string { return a.statement }

// Note returns a's note, and whether one is set.
func (a Alternative) Note() (string, bool) { return a.note, a.note != "" }

func (a Alternative) Extension() core.Extension { return a.extension }

// IsZero reports whether a is the zero value.
func (a Alternative) IsZero() bool { return a.statement == "" }

type alternativeJSON struct {
	Statement string          `json:"statement"`
	Note      string          `json:"note,omitempty"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes a as {"statement":..., "note":..., "extension":...},
// omitting note and extension when not set.
func (a Alternative) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return nil, fmt.Errorf("decision: marshal Alternative: %w", ErrInvalidAlternative)
	}
	raw := alternativeJSON{Statement: a.statement, Note: a.note}
	if !a.extension.IsZero() {
		raw.Extension = &a.extension
	}
	return json.Marshal(raw)
}

type alternativeUnmarshalJSON struct {
	Statement string          `json:"statement"`
	Note      json.RawMessage `json:"note"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes a from its JSON form, applying the same validation
// as NewAlternative and WithNote. An explicit JSON null for "note" or
// "extension" is rejected rather than silently treated as absent.
func (a *Alternative) UnmarshalJSON(data []byte) error {
	var raw alternativeUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Alternative: %w", err)
	}
	result, err := NewAlternative(raw.Statement)
	if err != nil {
		return err
	}
	if len(raw.Note) > 0 {
		if string(raw.Note) == "null" {
			return fmt.Errorf("decision: unmarshal Alternative: %w: note must not be null", ErrInvalidAlternative)
		}
		var note string
		if err := json.Unmarshal(raw.Note, &note); err != nil {
			return fmt.Errorf("decision: unmarshal Alternative: %w", err)
		}
		if result, err = result.WithNote(note); err != nil {
			return err
		}
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Alternative: %w: %w", ErrInvalidAlternative, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*a = result
	return nil
}
