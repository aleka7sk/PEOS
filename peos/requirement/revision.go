package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Revision is shorthand for "an Artifact Revision whose Artifact is a
// Requirement" (PEOS-005 §8) — not a separate PEOS entity. It composes
// core.ArtifactRevision by named field, per the specialized-Revision
// strategy documented on core.ArtifactRevision itself, and pairs it with
// typed Requirement Content.
type Revision struct {
	core    core.ArtifactRevision
	content Content
}

// newRevisionFromParts validates revision and content without reference
// to any Requirement, and is the path both NewRevision and UnmarshalJSON
// share. It cannot, and does not attempt to, check that revision belongs
// to any particular Requirement — see NewRevision and UnmarshalJSON's own
// documentation for why that check requires a Requirement value that a
// Revision's own JSON does not carry.
func newRevisionFromParts(revision core.ArtifactRevision, content Content) (Revision, error) {
	if revision.IsZero() {
		return Revision{}, fmt.Errorf("%w: core revision must not be zero", ErrInvalidRequirementRevision)
	}
	if content.IsZero() {
		return Revision{}, fmt.Errorf("%w", ErrMissingRequirementContent)
	}
	return Revision{core: revision, content: content}, nil
}

// NewRevision validates requirement, revision, and content and returns a
// Revision. requirement and revision must both be non-zero, content must
// be non-zero, and revision.ArtifactID() must equal requirement.ID().
// revision MAY have zero Representations: typed Content is this
// Revision's authoritative normative content (see doc.go).
func NewRevision(requirement Requirement, revision core.ArtifactRevision, content Content) (Revision, error) {
	if requirement.IsZero() {
		return Revision{}, fmt.Errorf("requirement: NewRevision: %w: requirement must not be zero", ErrInvalidRequirementRevision)
	}
	result, err := newRevisionFromParts(revision, content)
	if err != nil {
		return Revision{}, err
	}
	if revision.ArtifactID() != requirement.ID() {
		return Revision{}, fmt.Errorf("requirement: NewRevision: %w", ErrRequirementArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the Revision's underlying core.ArtifactRevision.
func (r Revision) Core() core.ArtifactRevision { return r.core }

// Content returns the Revision's typed Requirement content.
func (r Revision) Content() Content { return r.content }

// IsZero reports whether r is the zero value.
func (r Revision) IsZero() bool { return r.core.IsZero() && r.content.IsZero() }

type revisionJSON struct {
	Core    core.ArtifactRevision `json:"core"`
	Content Content               `json:"content"`
}

// MarshalJSON encodes r as {"core": {...}, "content": {...}}, per the
// nested-composition strategy documented on core.ArtifactRevision.
func (r Revision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Revision: %w", ErrInvalidRequirementRevision)
	}
	return json.Marshal(revisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core": {...}, "content":
// {...}} JSON form.
//
// This reconstructs r.core and r.content via the same checks
// newRevisionFromParts (and therefore NewRevision) applies, but cannot
// repeat NewRevision's ArtifactID-to-Requirement cross-check: a Revision's
// own JSON carries only its core.ArtifactRevision (with a bare
// ArtifactID) and its Content, never a core.Artifact with an ArtifactType
// to check that ArtifactID against. This is the same limitation
// core.ArtifactRevision itself already documents ("This package cannot
// verify that ArtifactID actually refers to a recorded Artifact... that
// cross-check is a future repository or validator rule"). A Revision
// decoded on its own is therefore, like a bare Representation, a value
// encoding of its own fields — not a complete, self-sufficient record of
// "this Revision of that Requirement" without external context supplying
// the matching Requirement.
func (r *Revision) UnmarshalJSON(data []byte) error {
	var raw revisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Revision: %w", err)
	}
	result, err := newRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return err
	}
	*r = result
	return nil
}
