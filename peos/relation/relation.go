package relation

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Relation is a PEOS-002 Artifact Relation: an explicit, typed, directed
// connection between exactly one source and one target engineering
// subject. See doc.go for the normative boundary this type deliberately
// stays within.
type Relation struct {
	relationType core.RelationType
	source       core.EngineeringSubjectRef
	target       core.EngineeringSubjectRef
	provenance   core.Provenance

	hasScope bool
	scope    core.Scope

	extension core.Extension
}

// New validates relationType, source, target, and provenance and returns
// a Relation with no scope and no extension data. Use WithScope and
// WithExtension to add those.
//
// New performs structural validation only: relationType, source, target,
// and provenance must each be non-zero. It does not check source != target,
// endpoint-kind compatibility with relationType, Artifact/Revision level
// compatibility, referential existence, cycles, duplicate relations, or
// inverse-relation existence — see doc.go for why each of these is
// deliberately left to a specialized validator, a future Traceability
// Model, or a repository layer, none of which exist in this package.
//
// relationType is not restricted to a known, pre-declared
// core.RelationType constant: core.RelationType is an open vocabulary,
// and this constructor accepts any non-zero value, including a
// Product-declared or otherwise unrecognized one.
func New(
	relationType core.RelationType,
	source core.EngineeringSubjectRef,
	target core.EngineeringSubjectRef,
	provenance core.Provenance,
) (Relation, error) {
	if relationType.IsZero() {
		return Relation{}, fmt.Errorf("relation: New: %w", ErrInvalidRelationType)
	}
	if source.IsZero() {
		return Relation{}, fmt.Errorf("relation: New: %w", ErrInvalidRelationSource)
	}
	if target.IsZero() {
		return Relation{}, fmt.Errorf("relation: New: %w", ErrInvalidRelationTarget)
	}
	if provenance.IsZero() {
		return Relation{}, fmt.Errorf("relation: New: %w: provenance must not be zero", ErrInvalidRelation)
	}
	return Relation{
		relationType: relationType,
		source:       source,
		target:       target,
		provenance:   provenance,
	}, nil
}

// WithScope returns a copy of r with its declared scope set. scope must
// be non-zero: unlike some other optional-scope constructs in this
// package family, a zero core.Scope is never treated as a hidden
// "unrestricted" or "absent" signal here — presence and absence are kept
// structurally distinct (see core.Scope's own documentation of this
// convention). Use WithoutScope to clear a previously set scope.
func (r Relation) WithScope(scope core.Scope) (Relation, error) {
	if scope.IsZero() {
		return Relation{}, fmt.Errorf("relation: WithScope: %w", core.ErrInvalidScope)
	}
	r.scope, r.hasScope = scope, true
	return r, nil
}

// WithoutScope returns a copy of r with its declared scope cleared.
func (r Relation) WithoutScope() Relation {
	r.scope, r.hasScope = core.Scope{}, false
	return r
}

// WithExtension returns a copy of r with its extension data set. Passing
// the zero core.Extension is equivalent to declaring no Product-specific
// extension data: core.Extension's own documentation treats the zero
// value and an Extension that would be empty identically as "no data
// present," so no separate presence flag is needed here.
//
// Extension exists only for genuine Product-specific data unrelated to
// this package's normative fields. It is not a place to smuggle in a
// PEOS-defined concept (such as a Supersession governance/authority
// reference) that a future specialized packet has not yet modeled —
// see doc.go.
func (r Relation) WithExtension(extension core.Extension) Relation {
	r.extension = extension
	return r
}

// WithoutExtension returns a copy of r with its extension data cleared.
func (r Relation) WithoutExtension() Relation {
	r.extension = core.Extension{}
	return r
}

// RelationType returns r's declared Relation Type.
func (r Relation) RelationType() core.RelationType { return r.relationType }

// Source returns r's source endpoint.
func (r Relation) Source() core.EngineeringSubjectRef { return r.source }

// Target returns r's target endpoint.
func (r Relation) Target() core.EngineeringSubjectRef { return r.target }

// Provenance returns r's provenance.
func (r Relation) Provenance() core.Provenance { return r.provenance }

// Scope returns r's declared scope, and whether one is set.
func (r Relation) Scope() (core.Scope, bool) { return r.scope, r.hasScope }

// Extension returns r's extension data. Absence of Product-specific
// extension data is reported by the returned value's own IsZero method,
// which is unambiguous for this type (see WithExtension).
func (r Relation) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r Relation) IsZero() bool {
	return r.relationType.IsZero() && r.source.IsZero() && r.target.IsZero() && r.provenance.IsZero()
}

type relationJSON struct {
	RelationType core.RelationType          `json:"relation_type"`
	Source       core.EngineeringSubjectRef `json:"source"`
	Target       core.EngineeringSubjectRef `json:"target"`
	Provenance   core.Provenance            `json:"provenance"`
	Scope        *core.Scope                `json:"scope,omitempty"`
	Extension    *core.Extension            `json:"extension,omitempty"`
}

// relationUnmarshalJSON mirrors relationJSON's field set for decoding
// only, with one difference: Scope is captured as raw, undecoded bytes
// rather than *core.Scope. A standard *core.Scope field cannot
// distinguish an absent "scope" key from one explicitly present with a
// JSON null value — encoding/json sets a pointer field to nil for both
// cases, without ever invoking core.Scope's own UnmarshalJSON. Capturing
// the raw bytes lets UnmarshalJSON below tell the two cases apart and
// reject an explicit null, per this type's own presence/absence
// contract for Scope (see WithScope).
type relationUnmarshalJSON struct {
	RelationType core.RelationType          `json:"relation_type"`
	Source       core.EngineeringSubjectRef `json:"source"`
	Target       core.EngineeringSubjectRef `json:"target"`
	Provenance   core.Provenance            `json:"provenance"`
	Scope        json.RawMessage            `json:"scope"`
	Extension    *core.Extension            `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"relation_type":..., "source":...,
// "target":..., "provenance":..., ...}, omitting scope and extension
// when not set. relation_type, source, target, and provenance are never
// omitted.
func (r Relation) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("relation: marshal Relation: %w", ErrInvalidRelation)
	}
	raw := relationJSON{
		RelationType: r.relationType,
		Source:       r.source,
		Target:       r.target,
		Provenance:   r.provenance,
	}
	if r.hasScope {
		raw.Scope = &r.scope
	}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as New, WithScope, and WithExtension. An explicit JSON
// null, or an explicit but zero-value scope object, is rejected rather
// than silently treated as "no scope."
func (r *Relation) UnmarshalJSON(data []byte) error {
	var raw relationUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("relation: unmarshal Relation: %w", err)
	}
	result, err := New(raw.RelationType, raw.Source, raw.Target, raw.Provenance)
	if err != nil {
		return err
	}
	if len(raw.Scope) > 0 {
		if string(raw.Scope) == "null" {
			return fmt.Errorf("relation: unmarshal Relation: %w: scope must not be null", core.ErrInvalidScope)
		}
		var scope core.Scope
		if err := json.Unmarshal(raw.Scope, &scope); err != nil {
			return fmt.Errorf("relation: unmarshal Relation: %w", err)
		}
		result, err = result.WithScope(scope)
		if err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*r = result
	return nil
}
