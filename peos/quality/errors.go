package quality

import "errors"

// Sentinel errors are wrapped with additional context by the functions in
// this package. Callers should use errors.Is against these sentinels rather
// than comparing error values directly.
//
// The complete PEOS-007 sentinel set is declared here up front, including
// the two sentinels reserved for Packet I.2 (Measurement Record and Quality
// Claim). Declaring them together keeps the package's error taxonomy
// inspectable as a whole and means I.2 does not have to reopen this file --
// the convention Packet H.1 established for peos/validation and H.2
// confirmed by needing no change to it. Each reserved sentinel documents the
// exact concept it belongs to; neither is referenced by Packet I.1 code,
// because I.1 deliberately implements only the Quality Profile side.
//
// There is deliberately no per-field sentinel for Unit, Scale, threshold
// operator, or observed value. Each of those is a component of exactly one
// owning value structure, and a caller that receives
// ErrInvalidQualityMeasure or ErrInvalidQualityThreshold already knows which
// aggregate rejected the input; the wrapped message names the field. Adding
// a sentinel per field would multiply near-identical error identities with
// no caller able to act differently on them.
//
// Component-owned failures are never re-attributed to this package: a zero
// or malformed core.Scope surfaces core.ErrInvalidScope, an empty identity
// or local key surfaces core.ErrEmptyIdentity, a malformed vocabulary value
// surfaces core.ErrInvalidVocabularyValue, a malformed nested core reference
// surfaces core.ErrInvalidPayload or
// core.ErrInvalidReferenceDiscriminator, and a revision-level reference
// missing its revision id surfaces core.ErrMissingRevisionID. This package
// wraps such errors, adding its own context, without replacing the owning
// sentinel.
var (
	// ErrInvalidQualityProfile is the aggregate sentinel for the Quality
	// Profile Artifact and its Revision content: a zero-value core.Artifact
	// supplied to NewProfile, a zero-value core.ArtifactRevision or
	// ProfileContent supplied to NewProfileRevision, a zero-value
	// core.Provenance, an empty Characteristic or Measure list, or an invalid
	// optional ProfileContent value, plus a zero-value marshal or a failed
	// top-level decode of Profile, ProfileRevision, or ProfileContent.
	// Component-specific failures use their own sentinels instead.
	ErrInvalidQualityProfile = errors.New("quality: quality profile is invalid")

	// ErrQualityProfileArtifactTypeMismatch is returned when NewProfile
	// receives a non-zero core.Artifact whose declared Artifact Type is not
	// ArtifactTypeQualityProfile (PEOS-007: "A Quality Profile is an Artifact
	// as defined by PEOS-002"). It mirrors
	// validation.ErrValidationPlanArtifactTypeMismatch and
	// requirement.ErrRequirementArtifactTypeMismatch.
	ErrQualityProfileArtifactTypeMismatch = errors.New("quality: artifact type is not quality profile")

	// ErrQualityProfileArtifactIDMismatch is returned when a
	// ProfileRevision's core Artifact Revision refers to a different Artifact
	// than the Profile it is being paired with. It mirrors
	// validation.ErrValidationPlanArtifactIDMismatch.
	ErrQualityProfileArtifactIDMismatch = errors.New("quality: artifact id mismatch between quality profile and revision")

	// ErrInvalidProfileApplicability is returned when a ProfileApplicability
	// is left in its zero (unstated) state, is decoded with an unrecognized
	// or missing kind, is unrestricted yet carries a scope, or is scoped yet
	// carries no scope. PEOS-007 lists "its applicability conditions" among a
	// Quality Profile Revision's unqualified SHALL-identify items, so a
	// ProfileContent may never leave it unstated;
	// NewUnrestrictedProfileApplicability exists so that "no restriction" is
	// an explicit, non-zero value.
	ErrInvalidProfileApplicability = errors.New("quality: profile applicability is invalid")

	// ErrInvalidQualityCharacteristic is returned when a Characteristic is
	// constructed or decoded with a zero local key, an empty or
	// whitespace-only Profile-scoped term, a zero external vocabulary value,
	// both a term and an external vocabulary, or neither; when its optional
	// description is empty after trimming; and when a zero-value
	// Characteristic is marshaled (PEOS-007 Quality Characteristic:
	// scoped by "the exact owning Quality Profile Revision in which it is
	// defined; or an exact externally referenced normative vocabulary").
	ErrInvalidQualityCharacteristic = errors.New("quality: quality characteristic is invalid")

	// ErrInvalidQualityMeasure is returned when a Measure is constructed or
	// decoded with a zero key, a zero Characteristic key, a zero Unit, a zero
	// Scale, or a zero Validation Method; when one of its optional values is
	// supplied zero or empty; and when a zero-value Measure is marshaled
	// (PEOS-007 Quality Measure SHALL-identify list).
	ErrInvalidQualityMeasure = errors.New("quality: quality measure is invalid")

	// ErrInvalidQualityThreshold is returned when a Threshold is constructed
	// or decoded with a zero key, a zero Measure key, a zero operator, or an
	// empty boundary value; when its optional description is empty after
	// trimming; and when a zero-value Threshold is marshaled (PEOS-007
	// Threshold: "a Quality Profile Revision-owned value structure defining a
	// boundary used for classification or for determining a Quality Claim
	// outcome").
	ErrInvalidQualityThreshold = errors.New("quality: quality threshold is invalid")

	// ErrInvalidQualityTarget is returned when a Target is constructed or
	// decoded with a zero key, a zero Measure key, or an empty desired value;
	// when its optional description is empty after trimming; and when a
	// zero-value Target is marshaled. Target has its own sentinel, distinct
	// from ErrInvalidQualityThreshold, because PEOS-007 requires the two
	// concepts to stay distinct: "a Target expresses intent, while a
	// Threshold expresses the boundary used for a Claim outcome. The two
	// SHALL NOT be conflated."
	ErrInvalidQualityTarget = errors.New("quality: quality target is invalid")

	// ErrInvalidQualityConstraint is returned when a Constraint is
	// constructed or decoded with a zero key or an empty restriction
	// statement; when its optional description is empty after trimming; and
	// when a zero-value Constraint is marshaled (PEOS-007 Quality
	// Constraint: "a normative restriction contained in a Quality Profile
	// Revision").
	ErrInvalidQualityConstraint = errors.New("quality: quality constraint is invalid")

	// ErrInvalidQualityRule is the shared sentinel for both
	// NormalizationRule and AggregationRule: a zero key, an empty or
	// whitespace-only description, or a zero-value marshal. The two share one
	// sentinel because they share their entire validation contract -- PEOS-007
	// describes both as Profile Revision-owned value structures carrying a
	// description of a transformation or combination, with identical
	// identity, revision, and lifecycle rules. They remain separate Go types
	// so that one can never be supplied where the other is expected.
	ErrInvalidQualityRule = errors.New("quality: quality rule is invalid")

	// ErrDuplicateProfileLocalKey is returned when one owned-value
	// collection of a single ProfileContent contains the same core.LocalKey
	// more than once. Uniqueness is enforced per owned-value kind, not across
	// kinds: PEOS-007 states no key uniqueness rule at all, and the necessary
	// derived rule is only that a criterion citing a Characteristic (or
	// Measure, Threshold, Target, Constraint) by key must resolve to exactly
	// one of them. A Characteristic and a Threshold may therefore share a key,
	// because every reference already determines its target collection --
	// internal references by the field that holds them, and criterion
	// citations by core.CriterionRef's kind discriminator.
	//
	// The wrapped message always names the owned-value kind and the offending
	// key, so a caller need not branch on a per-kind sentinel to report which
	// collection failed.
	ErrDuplicateProfileLocalKey = errors.New("quality: duplicate profile-local key")

	// ErrUnknownProfileLocalKey is returned when an internal reference inside
	// one ProfileContent names a core.LocalKey that its expected target
	// collection does not define: a Measure's Characteristic key or
	// Normalization Rule key, or a Threshold's or Target's Measure key. A
	// profile-local key has no meaning outside its owning Profile Revision, so
	// a reference naming a key absent from the expected collection cannot
	// denote anything.
	//
	// A key that exists only in a different collection does not satisfy
	// resolution: a Threshold whose measure key names a Characteristic is
	// rejected. The wrapped message names the referring value, the referenced
	// key, and the expected target collection.
	ErrUnknownProfileLocalKey = errors.New("quality: unknown profile-local key")

	// ErrInvalidMeasurementRecord is reserved for Packet I.2. It will be the
	// aggregate sentinel for MeasurementRecord -- the PEOS-007 Measurement
	// Record, which "specializes the Validation Execution Record defined by
	// PEOS-006" by composing validation.ExecutionRecord and adding the
	// observed value, unit, and scale that PEOS-007 mandates and PEOS-006 has
	// no field for. It will cover a zero composed record, a zero or empty
	// observed value, a zero Unit or Scale, a composed record lacking the
	// required Characteristic and Measure criteria, a zero-value marshal, and
	// a failed decode. It is not used by Packet I.1.
	ErrInvalidMeasurementRecord = errors.New("quality: measurement record is invalid")

	// ErrInvalidQualityClaim is reserved for Packet I.2. It will be the
	// sentinel for the Claim wrapper that enforces PEOS-007's two additional
	// rules on a validation.Claim: that its Claim Type is
	// core.ClaimTypeQuality, and that a Quality Claim whose subject is a
	// Requirement does not cite that same Requirement as its own criterion
	// ("the Requirement SHALL NOT be its own criterion"). The wrapper exists
	// because peos/validation is PEOS-006-correct in accepting that shape for
	// a non-Satisfaction Claim, so the stricter rule must be owned here and
	// re-applied on construction, on decode, and on every modifier. It is not
	// used by Packet I.1.
	ErrInvalidQualityClaim = errors.New("quality: quality claim is invalid")
)
