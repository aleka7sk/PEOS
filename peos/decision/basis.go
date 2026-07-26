package decision

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Basis is the immutable value object recording a Decision's established
// evidentiary basis (PEOS-004 "Decision Basis": "Decision Basis MUST be
// distinguishable from Decision Outcome"). Basis is optional on Decision:
// PEOS-004's MUST-have list for a Decision never names Basis as
// unconditionally required. Packet F carries only Evidence; assumptions,
// constraints, and uncertainty (PEOS-004's other named Basis components)
// are deferred to Packet F.1 without requiring any change to Decision
// itself, since Basis is already its own value object.
type Basis struct {
	evidence  []core.EvidenceArtifactRevisionRef
	extension core.Extension
}

// NewBasis validates evidence and returns a Basis with no extension data.
// At least one non-zero Evidence reference is required.
func NewBasis(evidence []core.EvidenceArtifactRevisionRef) (Basis, error) {
	if len(evidence) == 0 {
		return Basis{}, fmt.Errorf("decision: NewBasis: %w: at least one evidence reference is required", ErrInvalidBasis)
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(evidence))
	for idx, ref := range evidence {
		if ref.IsZero() {
			return Basis{}, fmt.Errorf("decision: NewBasis: %w: evidence reference must not be zero", ErrInvalidBasis)
		}
		cp[idx] = ref
	}
	return Basis{evidence: cp}, nil
}

// WithExtension returns a copy of b with its extension data set.
func (b Basis) WithExtension(e core.Extension) Basis {
	b.extension = e
	return b
}

// Evidence returns a defensive copy of b's declared evidence references, in
// declaration order.
func (b Basis) Evidence() []core.EvidenceArtifactRevisionRef {
	if len(b.evidence) == 0 {
		return nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(b.evidence))
	copy(cp, b.evidence)
	return cp
}

func (b Basis) Extension() core.Extension { return b.extension }

// IsZero reports whether b is the zero value.
func (b Basis) IsZero() bool { return len(b.evidence) == 0 }

type basisJSON struct {
	Evidence  []core.EvidenceArtifactRevisionRef `json:"evidence"`
	Extension *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes b as {"evidence":[...], "extension":...}, omitting
// extension when not set.
func (b Basis) MarshalJSON() ([]byte, error) {
	if b.IsZero() {
		return nil, fmt.Errorf("decision: marshal Basis: %w", ErrInvalidBasis)
	}
	raw := basisJSON{Evidence: b.evidence}
	if !b.extension.IsZero() {
		raw.Extension = &b.extension
	}
	return json.Marshal(raw)
}

type basisUnmarshalJSON struct {
	Evidence  json.RawMessage `json:"evidence"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes b from its JSON form, applying the same validation
// as NewBasis. "evidence" is required and its explicit JSON null is
// rejected; an explicit null for "extension" is likewise rejected.
func (b *Basis) UnmarshalJSON(data []byte) error {
	var raw basisUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Basis: %w", err)
	}
	if len(raw.Evidence) == 0 {
		return fmt.Errorf("decision: unmarshal Basis: %w: evidence is required", ErrInvalidBasis)
	}
	if string(raw.Evidence) == "null" {
		return fmt.Errorf("decision: unmarshal Basis: %w: evidence must not be null", ErrInvalidBasis)
	}
	var evidence []core.EvidenceArtifactRevisionRef
	if err := json.Unmarshal(raw.Evidence, &evidence); err != nil {
		return fmt.Errorf("decision: unmarshal Basis: %w", err)
	}
	result, err := NewBasis(evidence)
	if err != nil {
		return err
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Basis: %w: %w", ErrInvalidBasis, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*b = result
	return nil
}
