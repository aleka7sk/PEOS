package core

import (
	"encoding/json"
	"fmt"
)

// OriginKind classifies an Artifact Revision's Origin (PEOS-002 Artifact
// Origin). PEOS-002 names known, unknown, unavailable, and reconstructed
// as origin conditions; this is an open vocabulary, not a closed Go enum
// — a Product MAY declare additional origin kinds.
type OriginKind struct{ value VocabularyValue }

// NewOriginKind wraps v as an OriginKind.
func NewOriginKind(v VocabularyValue) OriginKind { return OriginKind{value: v} }

func (k OriginKind) Value() VocabularyValue { return k.value }
func (k OriginKind) IsZero() bool           { return k.value.IsZero() }
func (k OriginKind) String() string         { return k.value.String() }
func (k OriginKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.value)
}
func (k *OriginKind) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &k.value)
}

var (
	// OriginKindKnown marks a Revision whose engineering origin is
	// identified.
	OriginKindKnown = OriginKind{value: VocabularyValue{namespace: PEOSNamespace, value: "known"}}

	// OriginKindUnknown marks a Revision whose origin has not been
	// identified.
	OriginKindUnknown = OriginKind{value: VocabularyValue{namespace: PEOSNamespace, value: "unknown"}}

	// OriginKindUnavailable marks a Revision whose origin existed but is
	// no longer available to record.
	OriginKindUnavailable = OriginKind{value: VocabularyValue{namespace: PEOSNamespace, value: "unavailable"}}

	// OriginKindReconstructed marks a Revision whose origin was
	// reconstructed rather than recorded at the time it occurred.
	OriginKindReconstructed = OriginKind{value: VocabularyValue{namespace: PEOSNamespace, value: "reconstructed"}}
)

// Origin is the PEOS-002 Artifact Origin of an Artifact Revision: the
// engineering basis explaining why the Revision exists. Origin is kept
// deliberately minimal in this packet: it does not attempt a universal
// reference union spanning Decisions, immutable records, Template
// Application Records, and every other future construct that MAY justify
// a Revision's existence. Where a structured reference to another
// engineering subject is needed, it is represented through an Artifact
// Relation or a specialization-specific field defined by a later packet,
// not by this type.
//
// PEOS-002: "When origin is unknown, unavailable, disputed, or
// reconstructed, that condition MUST be explicit." This packet enforces
// that requirement by requiring a non-empty explanatory Note whenever
// Kind is anything other than OriginKindKnown.
type Origin struct {
	kind      OriginKind
	note      string
	extension Extension
}

// NewOrigin validates kind and note and returns an Origin. note is
// required to be non-empty whenever kind is not OriginKindKnown.
func NewOrigin(kind OriginKind, note string) (Origin, error) {
	if kind.IsZero() {
		return Origin{}, fmt.Errorf("core: NewOrigin: %w: kind must not be zero", ErrInvalidOrigin)
	}
	if !kind.Value().Equal(OriginKindKnown.Value()) && note == "" {
		return Origin{}, fmt.Errorf("core: NewOrigin: %w: a non-known origin kind requires an explanatory note", ErrInvalidOrigin)
	}
	return Origin{kind: kind, note: note}, nil
}

// WithExtension returns a copy of o with its extension data set.
func (o Origin) WithExtension(extension Extension) Origin {
	o.extension = extension
	return o
}

// Kind returns the Origin's kind.
func (o Origin) Kind() OriginKind { return o.kind }

// Note returns the Origin's explanatory note, and whether one is set.
func (o Origin) Note() (string, bool) { return o.note, o.note != "" }

// Extension returns the Origin's extension data.
func (o Origin) Extension() Extension { return o.extension }

// IsZero reports whether o is the zero value.
func (o Origin) IsZero() bool { return o.kind.IsZero() && o.note == "" }

type originJSON struct {
	Kind      OriginKind `json:"kind"`
	Note      string     `json:"note,omitempty"`
	Extension *Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes o as {"kind":..., "note":..., "extension":...}.
func (o Origin) MarshalJSON() ([]byte, error) {
	raw := originJSON{Kind: o.kind, Note: o.note}
	if !o.extension.IsZero() {
		raw.Extension = &o.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes o from its JSON form, applying the same
// validation as NewOrigin.
func (o *Origin) UnmarshalJSON(data []byte) error {
	var raw originJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal Origin: %w", err)
	}
	result, err := NewOrigin(raw.Kind, raw.Note)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*o = result
	return nil
}

// IntegrityMechanism classifies the mechanism behind an Integrity
// Identity (PEOS-002 Artifact Integrity). PEOS-002 names a cryptographic
// digest, an immutable version identifier, a content-addressed
// reference, an append-only record, and a trusted repository commit as
// example mechanisms; this is an open vocabulary — no mechanism,
// including a cryptographic digest, is assumed by this package.
type IntegrityMechanism struct{ value VocabularyValue }

func NewIntegrityMechanism(v VocabularyValue) IntegrityMechanism { return IntegrityMechanism{value: v} }

func (m IntegrityMechanism) Value() VocabularyValue { return m.value }
func (m IntegrityMechanism) IsZero() bool           { return m.value.IsZero() }
func (m IntegrityMechanism) String() string         { return m.value.String() }
func (m IntegrityMechanism) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.value)
}
func (m *IntegrityMechanism) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.value)
}

var (
	IntegrityMechanismCryptographicDigest        = IntegrityMechanism{value: VocabularyValue{namespace: PEOSNamespace, value: "cryptographic-digest"}}
	IntegrityMechanismImmutableVersionIdentifier = IntegrityMechanism{value: VocabularyValue{namespace: PEOSNamespace, value: "immutable-version-identifier"}}
	IntegrityMechanismContentAddressedReference  = IntegrityMechanism{value: VocabularyValue{namespace: PEOSNamespace, value: "content-addressed-reference"}}
	IntegrityMechanismAppendOnlyRecord           = IntegrityMechanism{value: VocabularyValue{namespace: PEOSNamespace, value: "append-only-record"}}
	IntegrityMechanismTrustedRepositoryCommit    = IntegrityMechanism{value: VocabularyValue{namespace: PEOSNamespace, value: "trusted-repository-commit"}}
)

// IntegrityProtectedScope names what an Integrity Identity's protection
// covers (PEOS-002: "The scope MAY include: content; normative metadata;
// Representation; relations embedded in the Revision; provenance
// information."). Open vocabulary; a Product MAY declare additional
// protected-scope values, or a composite value of its own choosing.
type IntegrityProtectedScope struct{ value VocabularyValue }

func NewIntegrityProtectedScope(v VocabularyValue) IntegrityProtectedScope {
	return IntegrityProtectedScope{value: v}
}

func (s IntegrityProtectedScope) Value() VocabularyValue { return s.value }
func (s IntegrityProtectedScope) IsZero() bool           { return s.value.IsZero() }
func (s IntegrityProtectedScope) String() string         { return s.value.String() }
func (s IntegrityProtectedScope) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}
func (s *IntegrityProtectedScope) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.value)
}

var (
	IntegrityProtectedScopeContent        = IntegrityProtectedScope{value: VocabularyValue{namespace: PEOSNamespace, value: "content"}}
	IntegrityProtectedScopeMetadata       = IntegrityProtectedScope{value: VocabularyValue{namespace: PEOSNamespace, value: "metadata"}}
	IntegrityProtectedScopeRepresentation = IntegrityProtectedScope{value: VocabularyValue{namespace: PEOSNamespace, value: "representation"}}
	IntegrityProtectedScopeRelations      = IntegrityProtectedScope{value: VocabularyValue{namespace: PEOSNamespace, value: "relations"}}
	IntegrityProtectedScopeProvenance     = IntegrityProtectedScope{value: VocabularyValue{namespace: PEOSNamespace, value: "provenance"}}
)

// IntegrityIdentity is the PEOS-002 Artifact Integrity of an Artifact
// Revision: a declarative statement that the Revision's protected scope
// can be verified to remain the state it claims to represent.
//
// This type is declarative only. It does not compute, verify, or assume
// any specific mechanism (in particular, it does not assume SHA-256 or
// any other specific digest algorithm); Packet B has no canonical
// serialization contract for an Artifact Revision, so integrity cannot be
// calculated generically here. A later packet or a Product-specific tool
// is responsible for actually computing and verifying integrity values;
// this type only carries the resulting declaration.
type IntegrityIdentity struct {
	mechanism      IntegrityMechanism
	value          string
	protectedScope IntegrityProtectedScope
	extension      Extension
}

// NewIntegrityIdentity validates mechanism, value, and protectedScope and
// returns an IntegrityIdentity.
func NewIntegrityIdentity(mechanism IntegrityMechanism, value string, protectedScope IntegrityProtectedScope) (IntegrityIdentity, error) {
	if mechanism.IsZero() {
		return IntegrityIdentity{}, fmt.Errorf("core: NewIntegrityIdentity: %w: mechanism must not be zero", ErrInvalidIntegrityIdentity)
	}
	if value == "" {
		return IntegrityIdentity{}, fmt.Errorf("core: NewIntegrityIdentity: %w: value must not be empty", ErrInvalidIntegrityIdentity)
	}
	if protectedScope.IsZero() {
		return IntegrityIdentity{}, fmt.Errorf("core: NewIntegrityIdentity: %w: protected scope must not be zero", ErrInvalidIntegrityIdentity)
	}
	return IntegrityIdentity{mechanism: mechanism, value: value, protectedScope: protectedScope}, nil
}

// WithExtension returns a copy of i with its extension data set.
func (i IntegrityIdentity) WithExtension(extension Extension) IntegrityIdentity {
	i.extension = extension
	return i
}

func (i IntegrityIdentity) Mechanism() IntegrityMechanism           { return i.mechanism }
func (i IntegrityIdentity) Value() string                           { return i.value }
func (i IntegrityIdentity) ProtectedScope() IntegrityProtectedScope { return i.protectedScope }
func (i IntegrityIdentity) Extension() Extension                    { return i.extension }

// IsZero reports whether i is the zero value.
func (i IntegrityIdentity) IsZero() bool {
	return i.mechanism.IsZero() && i.value == "" && i.protectedScope.IsZero()
}

type integrityIdentityJSON struct {
	Mechanism      IntegrityMechanism      `json:"mechanism"`
	Value          string                  `json:"value"`
	ProtectedScope IntegrityProtectedScope `json:"protected_scope"`
	Extension      *Extension              `json:"extension,omitempty"`
}

// MarshalJSON encodes i as {"mechanism":..., "value":..., "protected_scope":..., "extension":...}.
func (i IntegrityIdentity) MarshalJSON() ([]byte, error) {
	raw := integrityIdentityJSON{Mechanism: i.mechanism, Value: i.value, ProtectedScope: i.protectedScope}
	if !i.extension.IsZero() {
		raw.Extension = &i.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes i from its JSON form, applying the same
// validation as NewIntegrityIdentity.
func (i *IntegrityIdentity) UnmarshalJSON(data []byte) error {
	var raw integrityIdentityJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal IntegrityIdentity: %w", err)
	}
	result, err := NewIntegrityIdentity(raw.Mechanism, raw.Value, raw.ProtectedScope)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*i = result
	return nil
}

// ArtifactRevision is the PEOS-002 Artifact Revision: an identifiable and
// fixed state of an Artifact. ArtifactRevision and Artifact are
// independent domain records connected only by ArtifactID (see the
// package comment on Artifact); ArtifactRevision does not reference its
// predecessor, does not carry a revision number or semantic version, and
// does not carry a status or Lifecycle state field — Revision ordering
// and ancestry are Product-defined and deferred out of this packet
// (PEOS-002 §Revision Ordering explicitly refuses to mandate a single
// ordering mechanism), and status/Lifecycle state are governed by
// PEOS-003, explicitly out of scope here.
//
// This package cannot verify that ArtifactID actually refers to a
// recorded Artifact; that cross-check is a future repository or
// validator rule (PEOS-ART-006 in the Packet B blueprint), not something
// a single in-memory ArtifactRevision value can establish on its own.
type ArtifactRevision struct {
	artifactID      ArtifactID
	revisionID      ArtifactRevisionID
	origin          Origin
	provenance      Provenance
	integrity       IntegrityIdentity
	representations []Representation
	extension       Extension
}

// NewArtifactRevision validates artifactID, revisionID, origin,
// provenance, and integrity and returns an ArtifactRevision with no
// Representations and no extension data.
func NewArtifactRevision(
	artifactID ArtifactID,
	revisionID ArtifactRevisionID,
	origin Origin,
	provenance Provenance,
	integrity IntegrityIdentity,
) (ArtifactRevision, error) {
	if artifactID.IsZero() {
		return ArtifactRevision{}, fmt.Errorf("core: NewArtifactRevision: %w: %w", ErrInvalidArtifactRevision, ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return ArtifactRevision{}, fmt.Errorf("core: NewArtifactRevision: %w: %w", ErrInvalidArtifactRevision, ErrMissingRevisionID)
	}
	if origin.IsZero() {
		return ArtifactRevision{}, fmt.Errorf("core: NewArtifactRevision: %w: origin must not be zero", ErrInvalidArtifactRevision)
	}
	if provenance.IsZero() {
		return ArtifactRevision{}, fmt.Errorf("core: NewArtifactRevision: %w: provenance must not be zero", ErrInvalidArtifactRevision)
	}
	if integrity.IsZero() {
		return ArtifactRevision{}, fmt.Errorf("core: NewArtifactRevision: %w: integrity identity must not be zero", ErrInvalidArtifactRevision)
	}
	return ArtifactRevision{
		artifactID: artifactID,
		revisionID: revisionID,
		origin:     origin,
		provenance: provenance,
		integrity:  integrity,
	}, nil
}

// WithRepresentations returns a copy of r with its Representations set to
// exactly the values given, in the order given, replacing any previous
// Representations. A zero-value Representation among reps is rejected.
func (r ArtifactRevision) WithRepresentations(reps ...Representation) (ArtifactRevision, error) {
	if len(reps) == 0 {
		r.representations = nil
		return r, nil
	}
	cp := make([]Representation, len(reps))
	for idx, rep := range reps {
		if rep.IsZero() {
			return ArtifactRevision{}, fmt.Errorf("core: ArtifactRevision.WithRepresentations: %w", ErrInvalidRepresentation)
		}
		cp[idx] = rep
	}
	r.representations = cp
	return r, nil
}

// WithExtension returns a copy of r with its extension data set.
func (r ArtifactRevision) WithExtension(extension Extension) ArtifactRevision {
	r.extension = extension
	return r
}

func (r ArtifactRevision) ArtifactID() ArtifactID         { return r.artifactID }
func (r ArtifactRevision) RevisionID() ArtifactRevisionID { return r.revisionID }
func (r ArtifactRevision) Origin() Origin                 { return r.origin }
func (r ArtifactRevision) Provenance() Provenance         { return r.provenance }
func (r ArtifactRevision) Integrity() IntegrityIdentity   { return r.integrity }

// Representations returns a defensive copy of r's Representations, in
// declaration order.
func (r ArtifactRevision) Representations() []Representation {
	if len(r.representations) == 0 {
		return nil
	}
	cp := make([]Representation, len(r.representations))
	copy(cp, r.representations)
	return cp
}

func (r ArtifactRevision) Extension() Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r ArtifactRevision) IsZero() bool {
	return r.artifactID.IsZero() && r.revisionID.IsZero()
}

type artifactRevisionJSON struct {
	ArtifactID      ArtifactID         `json:"artifact_id"`
	RevisionID      ArtifactRevisionID `json:"revision_id"`
	Origin          Origin             `json:"origin"`
	Provenance      Provenance         `json:"provenance"`
	Integrity       IntegrityIdentity  `json:"integrity"`
	Representations []Representation   `json:"representations,omitempty"`
	Extension       *Extension         `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"artifact_id":..., "revision_id":..., ...}.
func (r ArtifactRevision) MarshalJSON() ([]byte, error) {
	raw := artifactRevisionJSON{
		ArtifactID: r.artifactID,
		RevisionID: r.revisionID,
		Origin:     r.origin,
		Provenance: r.provenance,
		Integrity:  r.integrity,
	}
	if len(r.representations) > 0 {
		raw.Representations = r.representations
	}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as NewArtifactRevision and WithRepresentations.
func (r *ArtifactRevision) UnmarshalJSON(data []byte) error {
	var raw artifactRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ArtifactRevision: %w", err)
	}
	result, err := NewArtifactRevision(raw.ArtifactID, raw.RevisionID, raw.Origin, raw.Provenance, raw.Integrity)
	if err != nil {
		return err
	}
	if len(raw.Representations) > 0 {
		result, err = result.WithRepresentations(raw.Representations...)
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
