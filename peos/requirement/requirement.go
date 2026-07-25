package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustArtifactTypeRequirement() core.ArtifactType {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, "requirement")
	if err != nil {
		panic(err)
	}
	return core.NewArtifactType(v)
}

// ArtifactTypeRequirement is the PEOS-005 Requirement Artifact Type.
// PEOS-005 does not itself fix an exact vocabulary string for this value
// — this is a design decision, namespaced under core.PEOSNamespace
// because Requirement is a PEOS-000-009-defined Artifact Type, not a
// Product-specific one, matching the convention already established for
// values such as core.OriginKindKnown. core.ArtifactType's own vocabulary
// remains fully open: nothing in core prevents a Product from declaring a
// different value, though there is no reason to.
var ArtifactTypeRequirement = mustArtifactTypeRequirement()

// Requirement is a PEOS-005 Requirement identity (§6-§8): a core.Artifact
// whose declared Artifact Type is ArtifactTypeRequirement. Requirement
// adds no field of its own — PEOS-005 §8 requires that "the Requirement
// identity SHALL NOT contain mutable current-value properties duplicating
// content owned by an Artifact Revision," and every PEOS-005 content
// element (Statement, Subject, Applicability, Origin, Authority,
// Classification, Rationale) is Revision-owned content (see Content),
// never Requirement identity.
type Requirement struct {
	core core.Artifact
}

// New validates artifact and returns a Requirement. artifact must be
// non-zero and its Type() must equal ArtifactTypeRequirement.
func New(artifact core.Artifact) (Requirement, error) {
	if artifact.IsZero() {
		return Requirement{}, fmt.Errorf("requirement: New: %w", ErrInvalidRequirement)
	}
	if artifact.Type() != ArtifactTypeRequirement {
		return Requirement{}, fmt.Errorf("requirement: New: %w", ErrRequirementArtifactTypeMismatch)
	}
	return Requirement{core: artifact}, nil
}

// Core returns the Requirement's underlying core.Artifact.
func (r Requirement) Core() core.Artifact { return r.core }

// ID returns the Requirement's identity.
func (r Requirement) ID() core.ArtifactID { return r.core.ID() }

// IsZero reports whether r is the zero value.
func (r Requirement) IsZero() bool { return r.core.IsZero() }

// MarshalJSON encodes r as the wire form of its underlying core.Artifact,
// with no additional envelope. This is possible, and sufficient, because
// core.Artifact's own JSON already carries artifact_type, which both
// preserves and (on Unmarshal) lets New re-verify that the decoded value
// is a Requirement.
func (r Requirement) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Requirement: %w", ErrInvalidRequirement)
	}
	return json.Marshal(r.core)
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as New.
func (r *Requirement) UnmarshalJSON(data []byte) error {
	var artifact core.Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return fmt.Errorf("requirement: unmarshal Requirement: %w", err)
	}
	result, err := New(artifact)
	if err != nil {
		return err
	}
	*r = result
	return nil
}
