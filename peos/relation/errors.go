package relation

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels
// rather than comparing error values directly.
var (
	// ErrInvalidRelation is returned when a Relation is constructed from
	// a zero-value core.Provenance, or when a zero-value Relation is
	// marshaled.
	ErrInvalidRelation = errors.New("relation: relation is invalid")

	// ErrInvalidRelationType is returned when a Relation is constructed
	// from a zero-value core.RelationType.
	ErrInvalidRelationType = errors.New("relation: relation type is invalid")

	// ErrInvalidRelationSource is returned when a Relation is constructed
	// from a zero-value source core.EngineeringSubjectRef.
	ErrInvalidRelationSource = errors.New("relation: source is invalid")

	// ErrInvalidRelationTarget is returned when a Relation is constructed
	// from a zero-value target core.EngineeringSubjectRef.
	ErrInvalidRelationTarget = errors.New("relation: target is invalid")
)
