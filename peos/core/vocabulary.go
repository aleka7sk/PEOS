package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VocabularyValue is an open, namespaced vocabulary value. PEOS-000-009
// define a number of extensible vocabularies (Artifact Type, Relation
// Type, Validation Method, Claim Outcome, and others; see the PEOS
// Reference Meta-Model Blueprint, Controlled Vocabulary Registry) that a
// Product or a future PEOS specification MAY extend with new values. This
// type never rejects an unrecognized value; it only rejects a value that
// fails to supply both a namespace and a value component.
//
// The string form is "namespace:value", split on the first colon only.
// This means Namespace MUST NOT itself contain a colon, while Value MAY
// contain colons (they are preserved verbatim in the remainder of the
// string after the first colon). This is the single parsing rule this
// package applies; it is not configurable per call site.
type VocabularyValue struct {
	namespace string
	value     string
}

// NewVocabularyValue validates namespace and value and returns a
// VocabularyValue. Both components are required and are not case-folded,
// trimmed beyond surrounding whitespace, or Unicode-normalized.
func NewVocabularyValue(namespace, value string) (VocabularyValue, error) {
	ns := strings.TrimSpace(namespace)
	val := strings.TrimSpace(value)
	if ns == "" || val == "" {
		return VocabularyValue{}, fmt.Errorf("core: NewVocabularyValue: %w", ErrInvalidVocabularyValue)
	}
	if strings.Contains(ns, ":") {
		return VocabularyValue{}, fmt.Errorf("core: NewVocabularyValue: namespace must not contain ':': %w", ErrInvalidVocabularyValue)
	}
	return VocabularyValue{namespace: ns, value: val}, nil
}

// ParseVocabularyValue parses the canonical "namespace:value" string form
// produced by VocabularyValue.String, splitting on the first colon only.
func ParseVocabularyValue(s string) (VocabularyValue, error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return VocabularyValue{}, fmt.Errorf("core: ParseVocabularyValue: missing ':' separator: %w", ErrInvalidVocabularyValue)
	}
	return NewVocabularyValue(s[:idx], s[idx+1:])
}

// Namespace returns the vocabulary's namespace component.
func (v VocabularyValue) Namespace() string { return v.namespace }

// Value returns the vocabulary's value component.
func (v VocabularyValue) Value() string { return v.value }

// IsZero reports whether v is the zero value.
func (v VocabularyValue) IsZero() bool { return v.namespace == "" && v.value == "" }

// String returns the canonical "namespace:value" form.
func (v VocabularyValue) String() string { return v.namespace + ":" + v.value }

// Equal reports whether v and other have the same namespace and value.
func (v VocabularyValue) Equal(other VocabularyValue) bool {
	return v.namespace == other.namespace && v.value == other.value
}

// MarshalJSON encodes v as its canonical "namespace:value" string form.
func (v VocabularyValue) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }

// UnmarshalJSON decodes v from its canonical "namespace:value" string
// form.
func (v *VocabularyValue) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal VocabularyValue: %w", err)
	}
	parsed, err := ParseVocabularyValue(s)
	if err != nil {
		return err
	}
	*v = parsed
	return nil
}

// PEOSNamespace is the namespace this package uses for vocabulary values
// PEOS-000-009 name directly, as opposed to values a Product contract or
// implementation introduces under its own namespace.
const PEOSNamespace = "peos"

// The typed wrappers below each carry a VocabularyValue for a specific
// PEOS vocabulary family. They exist only where mixing two different
// vocabulary families by accident is a realistic risk (for example,
// passing a ClaimOutcome where a ClaimType is expected). None of them is
// a closed Go enum: each constructor accepts any namespaced value, and
// the predefined constants below are a non-exhaustive, documented
// convenience, not the full set of legal values.

// ArtifactType is a namespaced Artifact Type vocabulary value (PEOS-002).
type ArtifactType struct{ value VocabularyValue }

// NewArtifactType wraps v as an ArtifactType.
func NewArtifactType(v VocabularyValue) ArtifactType { return ArtifactType{value: v} }

// Value returns the underlying VocabularyValue.
func (t ArtifactType) Value() VocabularyValue { return t.value }
func (t ArtifactType) IsZero() bool           { return t.value.IsZero() }
func (t ArtifactType) String() string         { return t.value.String() }
func (t ArtifactType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value)
}
func (t *ArtifactType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.value)
}

// ArtifactRole is a namespaced Artifact Role vocabulary value (PEOS-002),
// for example the Evidence role.
type ArtifactRole struct{ value VocabularyValue }

func NewArtifactRole(v VocabularyValue) ArtifactRole { return ArtifactRole{value: v} }
func (r ArtifactRole) Value() VocabularyValue        { return r.value }
func (r ArtifactRole) IsZero() bool                  { return r.value.IsZero() }
func (r ArtifactRole) String() string                { return r.value.String() }
func (r ArtifactRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}
func (r *ArtifactRole) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.value)
}

// ArtifactRoleEvidence is the PEOS-002 Evidence role.
var ArtifactRoleEvidence = ArtifactRole{value: VocabularyValue{namespace: PEOSNamespace, value: "evidence"}}

// RelationType is a namespaced Artifact Relation Type vocabulary value
// (PEOS-002 general contract; specialized per PEOS-005/008/009 relation
// families).
type RelationType struct{ value VocabularyValue }

func NewRelationType(v VocabularyValue) RelationType { return RelationType{value: v} }
func (t RelationType) Value() VocabularyValue        { return t.value }
func (t RelationType) IsZero() bool                  { return t.value.IsZero() }
func (t RelationType) String() string                { return t.value.String() }
func (t RelationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value)
}
func (t *RelationType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.value)
}

// Requirement Relation Type constants (PEOS-005 SS18-23).
var (
	RelationTypeDerivation              = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "derivation"}}
	RelationTypeRefinement              = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "refinement"}}
	RelationTypeDecomposition           = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "decomposition"}}
	RelationTypeDependency              = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "dependency"}}
	RelationTypeConflict                = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "conflict"}}
	RelationTypeRequirementSupersession = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "requirement-supersession"}}
	RelationTypeArtifactSupersession    = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "artifact-supersession"}}
	RelationTypeGeneratedFrom           = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "generated-from"}}
	RelationTypeTemplateComposition     = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "template-composition"}}
	RelationTypeTemplateSpecialization  = RelationType{value: VocabularyValue{namespace: PEOSNamespace, value: "template-specialization"}}
)

// ValidationMethod is a namespaced Validation Method vocabulary value
// (PEOS-006). It MAY additionally be Artifact-backed by a later packet;
// this type only carries the vocabulary identifier.
type ValidationMethod struct{ value VocabularyValue }

func NewValidationMethod(v VocabularyValue) ValidationMethod { return ValidationMethod{value: v} }
func (m ValidationMethod) Value() VocabularyValue            { return m.value }
func (m ValidationMethod) IsZero() bool                      { return m.value.IsZero() }
func (m ValidationMethod) String() string                    { return m.value.String() }
func (m ValidationMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.value)
}
func (m *ValidationMethod) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.value)
}

// ClaimType distinguishes Validation Claim specializations (PEOS-006
// Claim Type). This is the one vocabulary family this package treats as
// closed at the v0 baseline: PEOS-000-009 currently define exactly five
// Claim specializations, and introducing a sixth is a PEOS specification
// change, not a Product-level extension. NewClaimType therefore still
// accepts any namespaced value (so a future specification amendment does
// not require a breaking API change), but only the constants below are
// currently normative.
type ClaimType struct{ value VocabularyValue }

func NewClaimType(v VocabularyValue) ClaimType { return ClaimType{value: v} }
func (t ClaimType) Value() VocabularyValue     { return t.value }
func (t ClaimType) IsZero() bool               { return t.value.IsZero() }
func (t ClaimType) String() string             { return t.value.String() }
func (t ClaimType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value)
}
func (t *ClaimType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.value)
}

var (
	ClaimTypeSatisfaction        = ClaimType{value: VocabularyValue{namespace: PEOSNamespace, value: "satisfaction"}}
	ClaimTypeConformance         = ClaimType{value: VocabularyValue{namespace: PEOSNamespace, value: "conformance"}}
	ClaimTypeQuality             = ClaimType{value: VocabularyValue{namespace: PEOSNamespace, value: "quality"}}
	ClaimTypeCompliance          = ClaimType{value: VocabularyValue{namespace: PEOSNamespace, value: "compliance"}}
	ClaimTypeTemplateConformance = ClaimType{value: VocabularyValue{namespace: PEOSNamespace, value: "template-conformance"}}
)

// ClaimOutcome is a namespaced Claim Outcome vocabulary value (PEOS-006).
// The minimum set (satisfied / not satisfied / inconclusive) is fixed by
// PEOS-006; Product configuration MAY add further values provided they
// map unambiguously to one of those three (PEOS Reference Meta-Model
// Blueprint SS6). This type does not enforce that mapping; a future
// validator does.
type ClaimOutcome struct{ value VocabularyValue }

func NewClaimOutcome(v VocabularyValue) ClaimOutcome { return ClaimOutcome{value: v} }
func (o ClaimOutcome) Value() VocabularyValue        { return o.value }
func (o ClaimOutcome) IsZero() bool                  { return o.value.IsZero() }
func (o ClaimOutcome) String() string                { return o.value.String() }
func (o ClaimOutcome) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.value)
}
func (o *ClaimOutcome) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &o.value)
}

var (
	ClaimOutcomeSatisfied    = ClaimOutcome{value: VocabularyValue{namespace: PEOSNamespace, value: "satisfied"}}
	ClaimOutcomeNotSatisfied = ClaimOutcome{value: VocabularyValue{namespace: PEOSNamespace, value: "not-satisfied"}}
	ClaimOutcomeInconclusive = ClaimOutcome{value: VocabularyValue{namespace: PEOSNamespace, value: "inconclusive"}}
)

// CorrectionKind distinguishes correct / replace / invalidate (see
// correction.go). PEOS-006's Claim Correction section and the PEOS
// Reference Meta-Model Blueprint (SS10) name exactly these three kinds;
// this package treats the family as closed for that reason, while still
// exposing NewCorrectionKind for forward compatibility with a future
// specification amendment.
type CorrectionKind struct{ value VocabularyValue }

func NewCorrectionKind(v VocabularyValue) CorrectionKind { return CorrectionKind{value: v} }
func (k CorrectionKind) Value() VocabularyValue          { return k.value }
func (k CorrectionKind) IsZero() bool                    { return k.value.IsZero() }
func (k CorrectionKind) String() string                  { return k.value.String() }
func (k CorrectionKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.value)
}
func (k *CorrectionKind) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &k.value)
}

var (
	CorrectionKindCorrect    = CorrectionKind{value: VocabularyValue{namespace: PEOSNamespace, value: "correct"}}
	CorrectionKindReplace    = CorrectionKind{value: VocabularyValue{namespace: PEOSNamespace, value: "replace"}}
	CorrectionKindInvalidate = CorrectionKind{value: VocabularyValue{namespace: PEOSNamespace, value: "invalidate"}}
)
