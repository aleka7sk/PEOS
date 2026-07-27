package template

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels rather
// than comparing error values directly.
//
// Only the sentinels Packet K.1's shipped types actually return are declared
// here. This deliberately departs from the declare-the-whole-set-up-front
// convention Packets H.1/I.1/J.1 used: that convention produced four
// sentinels whose doc comments still described them in the future tense
// after the packet that activated them shipped, which Packet J.3.A had to
// raise as finding J3A-03 and Packet J.4 had to correct. Declaring only
// active sentinels means Packet K.2 adds its own with accurate comments the
// first time, and no comment in this file is ever stale.
//
// There is deliberately no per-field sentinel. Each field belongs to exactly
// one owning aggregate, and a caller that receives ErrInvalidTemplateParameter
// or ErrInvalidParameterConstraint already knows which aggregate rejected the
// input; the wrapped message names the field. There is also no
// generation-failure, rendering-failure, or compatibility-status sentinel of
// any kind: PEOS-009 defines no templating engine and no failure taxonomy,
// and none is invented here -- see doc.go.
//
// Component-owned failures are never re-attributed to this package: a zero or
// malformed core.Scope surfaces core.ErrInvalidScope, an empty identity or
// local key surfaces core.ErrEmptyIdentity, a malformed vocabulary value
// surfaces core.ErrInvalidVocabularyValue, and a malformed nested core
// reference surfaces core.ErrInvalidPayload or
// core.ErrInvalidReferenceDiscriminator. This package wraps such errors,
// adding its own context, without replacing the owning sentinel.
var (
	// ErrInvalidTemplate is the aggregate sentinel for the Template Artifact
	// and its Revision envelope: a zero-value core.Artifact supplied to
	// NewTemplate, a zero-value core.ArtifactRevision or TemplateContent
	// supplied to NewTemplateRevision, plus a zero-value marshal or a failed
	// top-level decode of Template or TemplateRevision.
	ErrInvalidTemplate = errors.New("template: template is invalid")

	// ErrTemplateArtifactTypeMismatch is returned when NewTemplate receives a
	// non-zero core.Artifact whose declared Artifact Type is not
	// ArtifactTypeTemplate (PEOS-009: "Template SHALL be an Artifact, as
	// defined by PEOS-002"). It mirrors
	// runtime.ErrRuntimeContractArtifactTypeMismatch,
	// quality.ErrQualityProfileArtifactTypeMismatch, and
	// validation.ErrValidationPlanArtifactTypeMismatch.
	ErrTemplateArtifactTypeMismatch = errors.New("template: artifact type is not template")

	// ErrTemplateArtifactIDMismatch is returned when a TemplateRevision's core
	// Artifact Revision refers to a different Artifact than the Template it is
	// being paired with.
	ErrTemplateArtifactIDMismatch = errors.New("template: artifact id mismatch between template and revision")

	// ErrInvalidTemplateContent is the aggregate sentinel for TemplateContent:
	// an empty or duplicate-bearing generated-Artifact-Type collection, an
	// empty expansion-semantics declaration, a zero mandatory scalar
	// (compatibility declaration, provenance), an invalid optional value, a
	// duplicate composition or specialization reference, plus a zero-value
	// marshal or a failed top-level decode. Element-specific failures use
	// their own sentinels instead.
	ErrInvalidTemplateContent = errors.New("template: template content is invalid")

	// ErrInvalidTemplateApplicability is returned when a TemplateApplicability
	// is left in its zero (unstated) state, is decoded with an unrecognized or
	// missing kind, is unrestricted yet carries a scope, or is scoped yet
	// carries no scope.
	ErrInvalidTemplateApplicability = errors.New("template: template applicability is invalid")

	// ErrInvalidTemplateParameter is returned when a Parameter is constructed
	// or decoded with a zero key, a zero parameter type, an empty description
	// after trimming, or a zero-value marshal.
	ErrInvalidTemplateParameter = errors.New("template: template parameter is invalid")

	// ErrInvalidParameterType is returned when a ParameterType is constructed
	// or decoded with a zero payload, an unrecognized or missing kind, or
	// both/neither arm present. PEOS-009 defines exactly two arms -- "a
	// controlled vocabulary/type; or an exact reference to an externally
	// governed normative type definition" -- so a ParameterType can never
	// leave its arm implicit.
	ErrInvalidParameterType = errors.New("template: parameter type is invalid")

	// ErrInvalidParameterDefault is returned when a ParameterDefault is
	// constructed or decoded with a zero target parameter key, an empty value
	// after trimming, or a zero-value marshal. A default whose target does not
	// resolve, a second default for one parameter, and a default targeting a
	// parameter that forbids default resolution are all rejected by
	// TemplateContent, which owns the aggregate view those checks need.
	ErrInvalidParameterDefault = errors.New("template: parameter default is invalid")

	// ErrInvalidParameterConstraint is returned when a ParameterConstraint is
	// constructed or decoded with a zero key, a zero or invalid affected
	// target, an empty rule after trimming, a zero evaluation point or failure
	// semantics, an invalid optional authority, or a zero-value marshal. Its
	// scope failures surface core.ErrInvalidScope instead.
	ErrInvalidParameterConstraint = errors.New("template: parameter constraint is invalid")

	// ErrInvalidConstraintTarget is returned when a ConstraintTarget is
	// constructed or decoded with a zero payload, an unrecognized or missing
	// kind, or both/neither arm present. PEOS-009 requires every Parameter
	// Constraint to identify "the affected parameter or generated content" --
	// a closed two-arm choice that is never left implicit.
	ErrInvalidConstraintTarget = errors.New("template: constraint target is invalid")

	// ErrInvalidCompatibilityDeclaration is returned when a
	// CompatibilityDeclaration is constructed or decoded with a zero
	// mandatory scalar, an empty descriptor after trimming, an empty or
	// duplicate-bearing applicable-Artifact-Type collection, or a zero-value
	// marshal.
	//
	// It is never returned to report that something is or is not compatible:
	// "Current compatibility is a derived interpretation, computed from the
	// applicable compatibility declarations at query time." This sentinel
	// reports an invalid declaration, never a compatibility verdict.
	ErrInvalidCompatibilityDeclaration = errors.New("template: compatibility declaration is invalid")

	// ErrDuplicateTemplateLocalKey is returned when a core.LocalKey repeats
	// within a single template-local namespace of one TemplateContent.
	// Uniqueness is enforced per namespace, not globally: the parameter
	// namespace is exactly TemplateContent.parameters, and the constraint
	// namespace is exactly TemplateContent.constraints. A key may therefore be
	// reused once by a Parameter and once by a ParameterConstraint (the two
	// namespaces are independent), but not twice within either.
	//
	// The wrapped message always names the namespace and the offending key, so
	// a caller need not branch on a per-namespace sentinel to report which one
	// failed.
	ErrDuplicateTemplateLocalKey = errors.New("template: duplicate template-local key")

	// ErrUnknownTemplateLocalKey is returned when an internal reference inside
	// one TemplateContent names a core.LocalKey that its expected target
	// collection does not define: a ParameterDefault or a parameter-targeting
	// ParameterConstraint naming a parameter that does not exist. Unlike
	// runtime.ErrUnknownRuntimeLocalKey, this sentinel is reachable from the
	// day it is declared -- TemplateContent resolves both reference kinds.
	ErrUnknownTemplateLocalKey = errors.New("template: unknown template-local key")
)
