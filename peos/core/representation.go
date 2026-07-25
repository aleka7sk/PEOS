package core

import (
	"encoding/json"
	"fmt"
)

// RepresentationRole classifies a Representation's relationship to its
// Revision's normative content (PEOS-002 Artifact Representation: "A
// Representation MUST identify... whether it is authoritative, derived,
// partial, or rendered."). PEOS-002 does not state that these four
// classifications are mutually exclusive alternatives; a Representation
// MAY reasonably be, for example, both partial and rendered at once.
// Classification is therefore represented as a duplicate-free, ordered
// set (Representation.Classification, []RepresentationRole), not a
// single enum-like field. This is an open vocabulary, not a closed Go
// enum — a Product MAY declare additional classification values.
type RepresentationRole struct{ value VocabularyValue }

// NewRepresentationRole wraps v as a RepresentationRole.
func NewRepresentationRole(v VocabularyValue) RepresentationRole {
	return RepresentationRole{value: v}
}

func (r RepresentationRole) Value() VocabularyValue { return r.value }
func (r RepresentationRole) IsZero() bool           { return r.value.IsZero() }
func (r RepresentationRole) String() string         { return r.value.String() }
func (r RepresentationRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}
func (r *RepresentationRole) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.value)
}

var (
	// RepresentationRoleAuthoritative marks a Representation as the
	// authoritative source of its Revision's normative content.
	RepresentationRoleAuthoritative = RepresentationRole{value: VocabularyValue{namespace: PEOSNamespace, value: "authoritative"}}

	// RepresentationRoleDerived marks a Representation as derived from
	// another, authoritative Representation.
	RepresentationRoleDerived = RepresentationRole{value: VocabularyValue{namespace: PEOSNamespace, value: "derived"}}

	// RepresentationRolePartial marks a Representation as incomplete
	// relative to its Revision's full normative content.
	RepresentationRolePartial = RepresentationRole{value: VocabularyValue{namespace: PEOSNamespace, value: "partial"}}

	// RepresentationRoleRendered marks a Representation as a generated or
	// rendered projection.
	RepresentationRoleRendered = RepresentationRole{value: VocabularyValue{namespace: PEOSNamespace, value: "rendered"}}
)

// Known discriminator values for RepresentationContent. Unlike the open
// vocabularies used elsewhere in this package, this set is closed: it
// names exactly the four content storage mechanisms PEOS-002 §Artifact
// Content describes (embedded, content-addressed, externally stored, or
// composed/other — mapped here to inline_bytes, inline_text,
// content_address, and external_reference), and no opaque/unknown-kind
// fallback is provided. This is a Go-level storage-mechanism
// discriminator, not an open PEOS normative vocabulary, so the
// forward-compatibility concerns that justify an opaque fallback
// elsewhere in this package (EngineeringSubjectRef, CriterionRef,
// RecordRef) do not apply here.
const (
	RepresentationContentKindInlineBytes       = "inline_bytes"
	RepresentationContentKindInlineText        = "inline_text"
	RepresentationContentKindExternalReference = "external_reference"
	RepresentationContentKindContentAddress    = "content_address"
)

type contentAddressPayload struct {
	algorithm VocabularyValue
	digest    string
}

// RepresentationContent is the tagged union of ways a Representation's
// content MAY be carried (PEOS-002 §Artifact Content): directly embedded
// bytes or text, a locator into an external system, or a content-
// addressed reference. Exactly one kind is ever populated; the type
// system makes "conflicting content forms" within one RepresentationContent
// structurally impossible.
type RepresentationContent struct {
	kind string

	inlineBytes       []byte
	inlineText        string
	externalReference string
	contentAddress    contentAddressPayload
}

// Kind returns the content's discriminator string.
func (c RepresentationContent) Kind() string { return c.kind }

// IsZero reports whether c is the zero value.
func (c RepresentationContent) IsZero() bool { return c.kind == "" }

// NewRepresentationContentFromInlineBytes validates content and returns a
// RepresentationContent carrying it directly. content is defensively
// copied.
func NewRepresentationContentFromInlineBytes(content []byte) (RepresentationContent, error) {
	if len(content) == 0 {
		return RepresentationContent{}, fmt.Errorf("core: NewRepresentationContentFromInlineBytes: %w", ErrMissingRepresentationContent)
	}
	cp := make([]byte, len(content))
	copy(cp, content)
	return RepresentationContent{kind: RepresentationContentKindInlineBytes, inlineBytes: cp}, nil
}

// AsInlineBytes returns a defensive copy of c's inline byte content, and
// true, if c's kind is inline_bytes.
func (c RepresentationContent) AsInlineBytes() ([]byte, bool) {
	if c.kind != RepresentationContentKindInlineBytes {
		return nil, false
	}
	cp := make([]byte, len(c.inlineBytes))
	copy(cp, c.inlineBytes)
	return cp, true
}

// NewRepresentationContentFromInlineText validates content and returns a
// RepresentationContent carrying it directly.
func NewRepresentationContentFromInlineText(content string) (RepresentationContent, error) {
	if content == "" {
		return RepresentationContent{}, fmt.Errorf("core: NewRepresentationContentFromInlineText: %w", ErrMissingRepresentationContent)
	}
	return RepresentationContent{kind: RepresentationContentKindInlineText, inlineText: content}, nil
}

// AsInlineText returns c's inline text content, and true, if c's kind is
// inline_text.
func (c RepresentationContent) AsInlineText() (string, bool) {
	if c.kind != RepresentationContentKindInlineText {
		return "", false
	}
	return c.inlineText, true
}

// NewRepresentationContentFromExternalReference validates locator and
// returns a RepresentationContent referencing it. PEOS-002 requires that
// a content reference "identify the intended content with enough
// precision to avoid silent resolution to a different state"; this
// package does not itself verify precision (a mutable branch name vs. an
// immutable versioned locator, for example), since doing so requires
// knowledge of the external system this package cannot have. No
// separate digest field is provided on this kind: PEOS-002 does not
// require one distinct from the locator itself, so it is deferred rather
// than added speculatively.
func NewRepresentationContentFromExternalReference(locator string) (RepresentationContent, error) {
	if locator == "" {
		return RepresentationContent{}, fmt.Errorf("core: NewRepresentationContentFromExternalReference: %w", ErrMissingRepresentationContent)
	}
	return RepresentationContent{kind: RepresentationContentKindExternalReference, externalReference: locator}, nil
}

// AsExternalReference returns c's external locator, and true, if c's kind
// is external_reference.
func (c RepresentationContent) AsExternalReference() (string, bool) {
	if c.kind != RepresentationContentKindExternalReference {
		return "", false
	}
	return c.externalReference, true
}

// NewRepresentationContentFromContentAddress validates algorithm and
// digest and returns a RepresentationContent carrying a content-addressed
// reference.
func NewRepresentationContentFromContentAddress(algorithm VocabularyValue, digest string) (RepresentationContent, error) {
	if algorithm.IsZero() {
		return RepresentationContent{}, fmt.Errorf("core: NewRepresentationContentFromContentAddress: %w: algorithm must not be zero", ErrMissingRepresentationContent)
	}
	if digest == "" {
		return RepresentationContent{}, fmt.Errorf("core: NewRepresentationContentFromContentAddress: %w: digest must not be empty", ErrMissingRepresentationContent)
	}
	return RepresentationContent{
		kind:           RepresentationContentKindContentAddress,
		contentAddress: contentAddressPayload{algorithm: algorithm, digest: digest},
	}, nil
}

// AsContentAddress returns c's algorithm and digest, and true, if c's
// kind is content_address.
func (c RepresentationContent) AsContentAddress() (algorithm VocabularyValue, digest string, ok bool) {
	if c.kind != RepresentationContentKindContentAddress {
		return VocabularyValue{}, "", false
	}
	return c.contentAddress.algorithm, c.contentAddress.digest, true
}

type representationContentEnvelope struct {
	Kind string          `json:"kind"`
	Ref  json.RawMessage `json:"ref"`
}

type contentAddressJSON struct {
	Algorithm VocabularyValue `json:"algorithm"`
	Digest    string          `json:"digest"`
}

// MarshalJSON encodes c as {"kind": ..., "ref": ...}.
func (c RepresentationContent) MarshalJSON() ([]byte, error) {
	if c.kind == "" {
		return nil, fmt.Errorf("core: marshal RepresentationContent: %w", ErrInvalidReferenceDiscriminator)
	}
	var (
		refBytes []byte
		err      error
	)
	switch c.kind {
	case RepresentationContentKindInlineBytes:
		refBytes, err = json.Marshal(c.inlineBytes)
	case RepresentationContentKindInlineText:
		refBytes, err = json.Marshal(c.inlineText)
	case RepresentationContentKindExternalReference:
		refBytes, err = json.Marshal(c.externalReference)
	case RepresentationContentKindContentAddress:
		refBytes, err = json.Marshal(contentAddressJSON{Algorithm: c.contentAddress.algorithm, Digest: c.contentAddress.digest})
	default:
		return nil, fmt.Errorf("core: marshal RepresentationContent: %w", ErrInvalidReferenceDiscriminator)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(representationContentEnvelope{Kind: c.kind, Ref: refBytes})
}

// UnmarshalJSON decodes c from {"kind": ..., "ref": ...}. An
// unrecognized kind fails explicitly; this closed union has no opaque
// fallback.
func (c *RepresentationContent) UnmarshalJSON(data []byte) error {
	var env representationContentEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("core: unmarshal RepresentationContent: %w", err)
	}
	if env.Kind == "" {
		return fmt.Errorf("core: unmarshal RepresentationContent: %w", ErrInvalidReferenceDiscriminator)
	}

	var (
		result RepresentationContent
		err    error
	)
	switch env.Kind {
	case RepresentationContentKindInlineBytes:
		var b []byte
		if err = json.Unmarshal(env.Ref, &b); err == nil {
			result, err = NewRepresentationContentFromInlineBytes(b)
		}
	case RepresentationContentKindInlineText:
		var s string
		if err = json.Unmarshal(env.Ref, &s); err == nil {
			result, err = NewRepresentationContentFromInlineText(s)
		}
	case RepresentationContentKindExternalReference:
		var s string
		if err = json.Unmarshal(env.Ref, &s); err == nil {
			result, err = NewRepresentationContentFromExternalReference(s)
		}
	case RepresentationContentKindContentAddress:
		var payload contentAddressJSON
		if err = json.Unmarshal(env.Ref, &payload); err == nil {
			result, err = NewRepresentationContentFromContentAddress(payload.Algorithm, payload.Digest)
		}
	default:
		err = fmt.Errorf("core: unmarshal RepresentationContent: unrecognized kind %q: %w", env.Kind, ErrInvalidReferenceDiscriminator)
	}
	if err != nil {
		return err
	}
	*c = result
	return nil
}

// Representation is a PEOS-002 Artifact Representation: a physical or
// logical encoding of an Artifact Revision's content. Representation is
// not an Artifact, carries no PEOS identity, and carries no reference
// back to its owning Artifact or Artifact Revision — its ownership is
// established structurally by being stored inside
// ArtifactRevision.Representations, never duplicated as a field on this
// type.
type Representation struct {
	content           RepresentationContent
	mediaType         VocabularyValue
	classification    []RepresentationRole
	hasLanguage       bool
	language          VocabularyValue
	hasTransformation bool
	transformation    VocabularyValue
	extension         Extension
}

// newRepresentationFromContent is the shared validation path every
// NewRepresentationFromX constructor delegates to, so the required-field
// and classification rules are defined exactly once.
func newRepresentationFromContent(content RepresentationContent, mediaType VocabularyValue, classification []RepresentationRole) (Representation, error) {
	if content.IsZero() {
		return Representation{}, fmt.Errorf("core: NewRepresentation: %w", ErrMissingRepresentationContent)
	}
	if mediaType.IsZero() {
		return Representation{}, fmt.Errorf("core: NewRepresentation: %w: media type must not be zero", ErrInvalidRepresentation)
	}
	if len(classification) == 0 {
		return Representation{}, fmt.Errorf("core: NewRepresentation: %w: at least one classification role is required", ErrInvalidRepresentation)
	}
	seen := make(map[string]bool, len(classification))
	deduped := make([]RepresentationRole, 0, len(classification))
	for _, role := range classification {
		if role.IsZero() {
			return Representation{}, fmt.Errorf("core: NewRepresentation: %w: classification role must not be zero", ErrInvalidRepresentation)
		}
		key := role.Value().String()
		if seen[key] {
			return Representation{}, fmt.Errorf("core: NewRepresentation: classification role %q: %w", key, ErrDuplicateRepresentationRole)
		}
		seen[key] = true
		deduped = append(deduped, role)
	}
	return Representation{content: content, mediaType: mediaType, classification: deduped}, nil
}

// NewRepresentationFromInlineBytes constructs a Representation whose
// content is directly embedded bytes.
func NewRepresentationFromInlineBytes(content []byte, mediaType VocabularyValue, classification ...RepresentationRole) (Representation, error) {
	c, err := NewRepresentationContentFromInlineBytes(content)
	if err != nil {
		return Representation{}, err
	}
	return newRepresentationFromContent(c, mediaType, classification)
}

// NewRepresentationFromInlineText constructs a Representation whose
// content is directly embedded text.
func NewRepresentationFromInlineText(content string, mediaType VocabularyValue, classification ...RepresentationRole) (Representation, error) {
	c, err := NewRepresentationContentFromInlineText(content)
	if err != nil {
		return Representation{}, err
	}
	return newRepresentationFromContent(c, mediaType, classification)
}

// NewRepresentationFromExternalReference constructs a Representation
// whose content is referenced by locator into an external system.
func NewRepresentationFromExternalReference(locator string, mediaType VocabularyValue, classification ...RepresentationRole) (Representation, error) {
	c, err := NewRepresentationContentFromExternalReference(locator)
	if err != nil {
		return Representation{}, err
	}
	return newRepresentationFromContent(c, mediaType, classification)
}

// NewRepresentationFromContentAddress constructs a Representation whose
// content is a content-addressed reference.
func NewRepresentationFromContentAddress(algorithm VocabularyValue, digest string, mediaType VocabularyValue, classification ...RepresentationRole) (Representation, error) {
	c, err := NewRepresentationContentFromContentAddress(algorithm, digest)
	if err != nil {
		return Representation{}, err
	}
	return newRepresentationFromContent(c, mediaType, classification)
}

// WithLanguage returns a copy of r with its declared language set.
// Passing the zero VocabularyValue is equivalent to leaving language
// unset.
func (r Representation) WithLanguage(language VocabularyValue) Representation {
	r.language, r.hasLanguage = language, !language.IsZero()
	return r
}

// WithTransformation returns a copy of r with its declared transformation
// description set. Passing the zero VocabularyValue is equivalent to
// leaving transformation unset.
func (r Representation) WithTransformation(transformation VocabularyValue) Representation {
	r.transformation, r.hasTransformation = transformation, !transformation.IsZero()
	return r
}

// WithExtension returns a copy of r with its extension data set.
func (r Representation) WithExtension(extension Extension) Representation {
	r.extension = extension
	return r
}

// Content returns r's content.
func (r Representation) Content() RepresentationContent { return r.content }

// MediaType returns r's declared media type.
func (r Representation) MediaType() VocabularyValue { return r.mediaType }

// Classification returns a defensive copy of r's classification roles,
// in declaration order.
func (r Representation) Classification() []RepresentationRole {
	if len(r.classification) == 0 {
		return nil
	}
	cp := make([]RepresentationRole, len(r.classification))
	copy(cp, r.classification)
	return cp
}

// Language returns r's declared language, and whether one is set.
func (r Representation) Language() (VocabularyValue, bool) { return r.language, r.hasLanguage }

// Transformation returns r's declared transformation description, and
// whether one is set.
func (r Representation) Transformation() (VocabularyValue, bool) {
	return r.transformation, r.hasTransformation
}

// Extension returns r's extension data.
func (r Representation) Extension() Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r Representation) IsZero() bool { return r.content.IsZero() && r.mediaType.IsZero() }

type representationJSON struct {
	Content        RepresentationContent `json:"content"`
	MediaType      VocabularyValue       `json:"media_type"`
	Classification []RepresentationRole  `json:"classification,omitempty"`
	Language       *VocabularyValue      `json:"language,omitempty"`
	Transformation *VocabularyValue      `json:"transformation,omitempty"`
	Extension      *Extension            `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"content":..., "media_type":..., ...}.
func (r Representation) MarshalJSON() ([]byte, error) {
	raw := representationJSON{Content: r.content, MediaType: r.mediaType}
	if len(r.classification) > 0 {
		raw.Classification = r.classification
	}
	if r.hasLanguage {
		raw.Language = &r.language
	}
	if r.hasTransformation {
		raw.Transformation = &r.transformation
	}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as the NewRepresentationFromX constructors.
func (r *Representation) UnmarshalJSON(data []byte) error {
	var raw representationJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal Representation: %w", err)
	}
	result, err := newRepresentationFromContent(raw.Content, raw.MediaType, raw.Classification)
	if err != nil {
		return err
	}
	if raw.Language != nil {
		result = result.WithLanguage(*raw.Language)
	}
	if raw.Transformation != nil {
		result = result.WithTransformation(*raw.Transformation)
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*r = result
	return nil
}
