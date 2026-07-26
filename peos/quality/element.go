package quality

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file defines the seven Quality Profile Revision-owned value
// structures PEOS-007 names, plus the three quality-local vocabulary
// wrappers they need.
//
// Every type here shares one ontological status: it is Revision-owned value
// content, not an entity. None of them carries a PEOS identity, a Ref, an
// Artifact, an Artifact Revision, a revision system, a lifecycle, or its own
// provenance -- provenance belongs to the owning ProfileContent, which
// records the origin of the Profile Revision as a whole. PEOS-007's
// Profile-Owned Rule Invariant states this for Threshold, Target, Quality
// Constraint, Normalization Rule, and Aggregation Rule; its Characteristic
// Scope and Measure Scope Invariants state the equivalent for Characteristic
// and Measure. The structural absence of those members is what prevents the
// non-conforming patterns "Quality Characteristic Revision" and "Quality
// Measure Version", rather than a convention or a runtime check.
//
// Each value's core.LocalKey is meaningful only within its owning Quality
// Profile Revision. A later Profile Revision MAY reuse, remove, or
// reintroduce a key, and a key from an earlier Revision SHALL NOT be assumed
// to name the same value in a later one unless an applicable Product
// contract defines that continuity. This package never compares keys across
// Profile Revisions.
//
// A Profile-scoped value becomes a Quality Claim criterion by pairing its
// owning Profile Revision's reference with its local key. For a
// Characteristic:
//
//	revisionRef, err := profileRevision.Ref()                  // core.ArtifactRevisionRef
//	elementRef, err := core.NewQualityElementCriterionRef(revisionRef, characteristic.Key())
//	criterion, err := core.CriterionRefFromQualityCharacteristic(elementRef)
//
// and identically through core.CriterionRefFromQualityMeasure,
// CriterionRefFromQualityThreshold, CriterionRefFromQualityTarget, and
// CriterionRefFromQualityConstraint for the other four citable kinds. The
// criterion kind, not the key, is what says which collection the key is to
// be resolved in -- which is why the same key may name a Characteristic and
// a Threshold in one Revision without ambiguity.

// --- quality-local vocabulary wrappers ---------------------------------------

// Unit, Scale, and ThresholdOperator are namespaced vocabulary values
// specific to PEOS-007. They live here rather than in peos/core for the same
// reason ArtifactTypeQualityProfile does: the vocabulary belongs to the
// specification that defines the concept, and peos/core carries no PEOS-007
// concept.
//
// None of the three is a closed Go enum, and none predeclares any constant.
// PEOS-007 names no unit, no scale, and no comparison operator: its
// Non-Goals disclaim "a universal quality metric catalog" and "a specific
// scoring formula or weighting scheme". Predeclaring a set here would invent
// a catalog the specification deliberately omits, and a units or
// dimensional-analysis framework would go further still. What the unit,
// scale, and operator mean, and how a value is compared against a boundary,
// is Product-owned.
//
// All three constructors are infallible wrappers returning a bare value,
// following the seven-of-eight majority convention in core's own vocabulary
// family (NewArtifactType, NewArtifactRole, NewRelationType,
// NewValidationMethod, NewClaimType, NewClaimOutcome, NewCorrectionKind).
// They deliberately do not copy core.NewExecutionOutcome's fallible
// (value, error) shape: validating the wrapper would only re-check what
// core.NewVocabularyValue already guarantees, and the aggregate constructors
// that consume these values reject a zero one anyway, which is where the
// mandatory-ness actually lives.

// Unit is a namespaced Quality Measure unit vocabulary value (PEOS-007
// Quality Measure: a Measure SHALL identify "its unit").
type Unit struct{ value core.VocabularyValue }

// NewUnit wraps v as a Unit.
func NewUnit(v core.VocabularyValue) Unit { return Unit{value: v} }

// Value returns the underlying core.VocabularyValue.
func (u Unit) Value() core.VocabularyValue { return u.value }
func (u Unit) String() string              { return u.value.String() }
func (u Unit) IsZero() bool                { return u.value.IsZero() }

// Equal reports whether u and other carry the same vocabulary value.
func (u Unit) Equal(other Unit) bool { return u.value.Equal(other.value) }

func (u Unit) MarshalJSON() ([]byte, error) { return json.Marshal(u.value) }

func (u *Unit) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &u.value) }

// Scale is a namespaced Quality Measure scale vocabulary value (PEOS-007
// Quality Measure: a Measure SHALL identify "its scale"). It is a distinct
// Go type from Unit so that one can never be passed where the other is
// expected.
type Scale struct{ value core.VocabularyValue }

// NewScale wraps v as a Scale.
func NewScale(v core.VocabularyValue) Scale { return Scale{value: v} }

// Value returns the underlying core.VocabularyValue.
func (s Scale) Value() core.VocabularyValue { return s.value }
func (s Scale) String() string              { return s.value.String() }
func (s Scale) IsZero() bool                { return s.value.IsZero() }

// Equal reports whether s and other carry the same vocabulary value.
func (s Scale) Equal(other Scale) bool { return s.value.Equal(other.value) }

func (s Scale) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }

func (s *Scale) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &s.value) }

// ThresholdOperator is the namespaced comparison a Threshold applies between
// a measured value and its boundary. PEOS-007 defines a Threshold as "a
// boundary used for classification or for determining a Quality Claim
// outcome"; a boundary without a stated comparison direction is ambiguous,
// which is why Threshold requires one. PEOS-007 names no operator values, so
// this vocabulary is entirely open and its interpretation is Product-owned.
type ThresholdOperator struct{ value core.VocabularyValue }

// NewThresholdOperator wraps v as a ThresholdOperator.
func NewThresholdOperator(v core.VocabularyValue) ThresholdOperator {
	return ThresholdOperator{value: v}
}

// Value returns the underlying core.VocabularyValue.
func (o ThresholdOperator) Value() core.VocabularyValue { return o.value }
func (o ThresholdOperator) String() string              { return o.value.String() }
func (o ThresholdOperator) IsZero() bool                { return o.value.IsZero() }

// Equal reports whether o and other carry the same vocabulary value.
func (o ThresholdOperator) Equal(other ThresholdOperator) bool { return o.value.Equal(other.value) }

func (o ThresholdOperator) MarshalJSON() ([]byte, error) { return json.Marshal(o.value) }

func (o *ThresholdOperator) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &o.value)
}

// --- shared validation helpers -----------------------------------------------

// trimmedRequired trims value and rejects it if nothing remains, attributing
// the failure to the supplied sentinel.
func trimmedRequired(caller, label, value string, sentinel error) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("quality: %s: %w: %s must not be empty", caller, sentinel, label)
	}
	return trimmed, nil
}

// rejectNullRaw reports an error when raw is an explicit JSON null, which
// every optional single value in this package rejects rather than silently
// treating as absent.
func rejectNullRaw(caller, label string, raw json.RawMessage, sentinel error) error {
	if string(raw) == "null" {
		return fmt.Errorf("quality: unmarshal %s: %w: %s must not be null", caller, sentinel, label)
	}
	return nil
}

// --- Characteristic ----------------------------------------------------------

type characteristicKind string

const (
	characteristicKindProfile  characteristicKind = "profile"
	characteristicKindExternal characteristicKind = "external"
)

// Characteristic is a PEOS-007 Quality Characteristic: "a controlled term
// whose identity and meaning are scoped by: the exact owning Quality Profile
// Revision in which it is defined; or an exact externally referenced
// normative vocabulary."
//
// Those two scopings are the type's two arms, and they are exclusive. A
// Characteristic must be exactly one of them: a value carrying both a
// Profile-scoped term and an external vocabulary reference would have two
// competing sources of meaning, and one carrying neither would have none --
// the non-conforming pattern "Threshold with Hidden External Meaning" names
// exactly that failure for Thresholds, and the Characteristic Scope
// Invariant requires the same discipline here. Both cases are rejected at
// construction and on decode.
//
// Both arms carry a local key. The key is how the Characteristic is
// referenced from inside its owning Profile Revision (by a Measure) and from
// outside it (as a core.CriterionRef). The externally scoped arm still needs
// one for that reason: the external vocabulary supplies the
// Characteristic's *meaning* and its globally stable identity, but the key
// is what names it within this Revision's own content.
//
// A Characteristic has "no independent revision system of its own". Changing
// its meaning requires a new Profile Revision, or a new identity within the
// external vocabulary -- never an in-place edit, which is why every modifier
// here returns a copy and none of them can reach the key or the arm.
type Characteristic struct {
	kind       characteristicKind
	key        core.LocalKey
	term       string
	vocabulary core.VocabularyValue

	description string
	extension   core.Extension
}

// NewProfileCharacteristic validates key and term and returns a
// Profile-scoped Characteristic: one whose identity and meaning are scoped
// by its owning Quality Profile Revision.
//
// key must be non-zero. term must be non-empty after trimming surrounding
// whitespace; the trimmed value is stored. term is the controlled term
// itself, deliberately an opaque string: PEOS-007 defines no term vocabulary
// and its Non-Goals disclaim a metric catalog, so the term's content is
// Product-owned.
func NewProfileCharacteristic(key core.LocalKey, term string) (Characteristic, error) {
	if key.IsZero() {
		return Characteristic{}, fmt.Errorf("quality: NewProfileCharacteristic: %w: key must not be zero", ErrInvalidQualityCharacteristic)
	}
	trimmed, err := trimmedRequired("NewProfileCharacteristic", "term", term, ErrInvalidQualityCharacteristic)
	if err != nil {
		return Characteristic{}, err
	}
	return Characteristic{kind: characteristicKindProfile, key: key, term: trimmed}, nil
}

// NewExternalCharacteristic validates key and vocabulary and returns an
// externally scoped Characteristic: one whose identity and meaning come from
// "an exact externally referenced normative vocabulary".
//
// key must be non-zero, and vocabulary must be a non-zero
// core.VocabularyValue naming the exact external concept. This is the only
// case in which a Characteristic "claim[s] globally stable independent
// identity", and PEOS-007 permits it precisely because that identity is
// supplied and governed externally rather than minted here.
func NewExternalCharacteristic(key core.LocalKey, vocabulary core.VocabularyValue) (Characteristic, error) {
	if key.IsZero() {
		return Characteristic{}, fmt.Errorf("quality: NewExternalCharacteristic: %w: key must not be zero", ErrInvalidQualityCharacteristic)
	}
	if vocabulary.IsZero() {
		return Characteristic{}, fmt.Errorf("quality: NewExternalCharacteristic: %w: external vocabulary must not be zero", ErrInvalidQualityCharacteristic)
	}
	return Characteristic{kind: characteristicKindExternal, key: key, vocabulary: vocabulary}, nil
}

// WithDescription returns a copy of c with an optional human-readable
// description set. description must be non-empty after trimming; the trimmed
// value is stored. A description explains the Characteristic; it never
// replaces the term or the external vocabulary reference as the source of
// meaning.
func (c Characteristic) WithDescription(description string) (Characteristic, error) {
	trimmed, err := trimmedRequired("Characteristic.WithDescription", "description", description, ErrInvalidQualityCharacteristic)
	if err != nil {
		return Characteristic{}, err
	}
	c.description = trimmed
	return c, nil
}

// WithoutDescription returns a copy of c with its description cleared.
func (c Characteristic) WithoutDescription() Characteristic {
	c.description = ""
	return c
}

// WithExtension returns a copy of c with its extension data set. Passing the
// zero core.Extension is equivalent to declaring none, per core.Extension's
// own documented contract.
func (c Characteristic) WithExtension(extension core.Extension) Characteristic {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c Characteristic) WithoutExtension() Characteristic {
	c.extension = core.Extension{}
	return c
}

// Kind returns c's discriminator, "profile" or "external". The zero value
// returns the empty string.
func (c Characteristic) Kind() string { return string(c.kind) }

// Key returns c's profile-local key. It is meaningful only within c's owning
// Quality Profile Revision.
func (c Characteristic) Key() core.LocalKey { return c.key }

// Term returns c's Profile-scoped controlled term, and whether one is set
// (that is, whether c is the Profile-scoped arm).
func (c Characteristic) Term() (string, bool) {
	if c.kind != characteristicKindProfile {
		return "", false
	}
	return c.term, true
}

// ExternalVocabulary returns the exact external vocabulary concept scoping
// c's identity and meaning, and whether one is set (that is, whether c is
// the externally scoped arm).
func (c Characteristic) ExternalVocabulary() (core.VocabularyValue, bool) {
	if c.kind != characteristicKindExternal {
		return core.VocabularyValue{}, false
	}
	return c.vocabulary, true
}

// IsExternallyScoped reports whether c's identity and meaning come from an
// external normative vocabulary rather than from its owning Profile
// Revision.
func (c Characteristic) IsExternallyScoped() bool { return c.kind == characteristicKindExternal }

// Description returns c's optional description, and whether one is set.
func (c Characteristic) Description() (string, bool) {
	return c.description, c.description != ""
}

// Extension returns c's extension data.
func (c Characteristic) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value -- the unscoped state PEOS-007
// does not permit.
func (c Characteristic) IsZero() bool { return c.kind == "" }

type characteristicJSON struct {
	Kind        string                `json:"kind"`
	Key         core.LocalKey         `json:"key"`
	Term        string                `json:"term,omitempty"`
	Vocabulary  *core.VocabularyValue `json:"vocabulary,omitempty"`
	Description string                `json:"description,omitempty"`
	Extension   *core.Extension       `json:"extension,omitempty"`
}

// characteristicUnmarshalJSON mirrors characteristicJSON for decoding, with
// Description captured as raw bytes so an explicit null can be
// distinguished from an absent key and rejected -- the json.RawMessage probe
// technique Packet D.1 established. Term and Vocabulary need the same
// treatment for a different reason: their presence, not just their value, is
// what selects the arm, and a null must not be read as "arm present".
type characteristicUnmarshalJSON struct {
	Kind        string          `json:"kind"`
	Key         core.LocalKey   `json:"key"`
	Term        json.RawMessage `json:"term"`
	Vocabulary  json.RawMessage `json:"vocabulary"`
	Description json.RawMessage `json:"description"`
	Extension   *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"kind":"profile","key":...,"term":...} or
// {"kind":"external","key":...,"vocabulary":...}, plus description and
// extension when set.
//
// There is no "id" key, no revision key, no lifecycle key, no relation key,
// and no provenance key: their absence is the structural proof that a Quality
// Characteristic is Revision-owned value content and not an independently
// identified, independently revised entity.
func (c Characteristic) MarshalJSON() ([]byte, error) {
	raw := characteristicJSON{Key: c.key, Description: c.description}
	switch c.kind {
	case characteristicKindProfile:
		raw.Kind = string(characteristicKindProfile)
		raw.Term = c.term
	case characteristicKindExternal:
		raw.Kind = string(characteristicKindExternal)
		raw.Vocabulary = &c.vocabulary
	default:
		return nil, fmt.Errorf("quality: marshal Characteristic: %w", ErrInvalidQualityCharacteristic)
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as the two constructors and each With* method, so a decoded Characteristic
// can never be constructor-impossible. The receiver is left untouched unless
// every check passes.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - kind: a missing key, an explicit null, and an unrecognized value are
//     all rejected with ErrInvalidQualityCharacteristic. An explicit null for
//     the whole document yields an empty kind and is rejected the same way --
//     null is never read as "unscoped" or as either arm.
//   - key: a missing key leaves the field zero and is rejected with
//     ErrInvalidQualityCharacteristic. An explicit null invokes
//     core.LocalKey's own UnmarshalJSON, which fails there, so the error
//     carries core.ErrEmptyIdentity instead. Both are rejected; the sentinel
//     sets differ.
//   - term: required by, and permitted only for, the profile arm. Absent,
//     null, and empty-after-trimming are all rejected for that arm; any
//     presence at all is rejected for the external arm.
//   - vocabulary: required by, and permitted only for, the external arm,
//     with the same three rejections mirrored.
//   - description: a missing key means absent; an explicit null is rejected
//     rather than silently treated as absent.
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract.
func (c *Characteristic) UnmarshalJSON(data []byte) error {
	var raw characteristicUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal Characteristic: %w: %w", ErrInvalidQualityCharacteristic, err)
	}

	hasTerm := len(raw.Term) > 0 && string(raw.Term) != "null"
	hasVocabulary := len(raw.Vocabulary) > 0 && string(raw.Vocabulary) != "null"

	var (
		result Characteristic
		err    error
	)
	switch raw.Kind {
	case string(characteristicKindProfile):
		if hasVocabulary {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: a profile-scoped characteristic must not carry an external vocabulary", ErrInvalidQualityCharacteristic)
		}
		if !hasTerm {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: a profile-scoped characteristic requires a term", ErrInvalidQualityCharacteristic)
		}
		var term string
		if err = json.Unmarshal(raw.Term, &term); err != nil {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: %w", ErrInvalidQualityCharacteristic, err)
		}
		if result, err = NewProfileCharacteristic(raw.Key, term); err != nil {
			return err
		}
	case string(characteristicKindExternal):
		if hasTerm {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: an externally scoped characteristic must not carry a term", ErrInvalidQualityCharacteristic)
		}
		if !hasVocabulary {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: an externally scoped characteristic requires an external vocabulary", ErrInvalidQualityCharacteristic)
		}
		var vocabulary core.VocabularyValue
		if err = json.Unmarshal(raw.Vocabulary, &vocabulary); err != nil {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: %w", ErrInvalidQualityCharacteristic, err)
		}
		if result, err = NewExternalCharacteristic(raw.Key, vocabulary); err != nil {
			return err
		}
	default:
		return fmt.Errorf("quality: unmarshal Characteristic: unrecognized kind %q: %w", raw.Kind, ErrInvalidQualityCharacteristic)
	}

	if len(raw.Description) > 0 {
		if err = rejectNullRaw("Characteristic", "description", raw.Description, ErrInvalidQualityCharacteristic); err != nil {
			return err
		}
		var description string
		if err = json.Unmarshal(raw.Description, &description); err != nil {
			return fmt.Errorf("quality: unmarshal Characteristic: %w: %w", ErrInvalidQualityCharacteristic, err)
		}
		if result, err = result.WithDescription(description); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*c = result
	return nil
}

// --- Measure -----------------------------------------------------------------

// Measure is a PEOS-007 Quality Measure: it "defines how a Quality
// Characteristic is observed or computed", as "a Quality Profile
// Revision-owned definition" with "no mandatory independent Artifact
// identity and no Quality Measure Version".
//
// PEOS-007 lists eight items a Quality Measure SHALL identify. Five are
// mandatory constructor arguments here -- the Characteristic it measures, its
// unit, its scale, its method, plus the local key that names it within its
// owning Revision -- because for each of those a zero value is ambiguous
// between "unstated" and a legitimate value, so no With* call may be the
// first place it is established.
//
// The remaining three (the Evidence required as input, how uncertainty is
// handled, its valid range) plus the conditional Normalization Rule are
// optional. For each, the zero value is already an unambiguous "none
// declared" rather than a third unstated state: an empty required-Evidence
// list means none is required, and an absent uncertainty-handling or
// valid-range statement means none is declared. The Normalization Rule is
// explicitly conditional in the specification itself -- "where the measured
// value requires normalization" -- which is the same qualified/unqualified
// contrast that makes ProfileApplicability mandatory and these optional.
type Measure struct {
	key            core.LocalKey
	characteristic core.LocalKey
	unit           Unit
	scale          Scale
	method         core.ValidationMethod

	requiredEvidence    []string
	uncertaintyHandling string
	validRange          string
	normalizationRule   core.LocalKey
	extension           core.Extension
}

// NewMeasure validates its five arguments and returns a Measure with no
// required Evidence, uncertainty handling, valid range, Normalization Rule,
// or extension data. Use the With* methods to add those.
//
// All five are mandatory and must be non-zero. characteristic is the
// profile-local key of the Characteristic this Measure measures; whether that
// key actually resolves is checked by NewProfileContent, which is the only
// place the owning Revision's full Characteristic key set is known. method is
// a core.ValidationMethod -- PEOS-006's vocabulary, reused rather than
// redefined, because PEOS-007 states that "PEOS-006 owns the mechanism
// (Plan, Planned Activity, Method, Execution Record, Evidence, Claim...)"
// while PEOS-007 owns only the vocabulary and configuration.
//
// A successful call always returns a fully valid Measure: no mandatory field
// is supplied by a later With* call, so a Measure can never exist in a
// partially established state.
func NewMeasure(
	key core.LocalKey,
	characteristic core.LocalKey,
	unit Unit,
	scale Scale,
	method core.ValidationMethod,
) (Measure, error) {
	if key.IsZero() {
		return Measure{}, fmt.Errorf("quality: NewMeasure: %w: key must not be zero", ErrInvalidQualityMeasure)
	}
	if characteristic.IsZero() {
		return Measure{}, fmt.Errorf("quality: NewMeasure: %w: characteristic key must not be zero", ErrInvalidQualityMeasure)
	}
	if unit.IsZero() {
		return Measure{}, fmt.Errorf("quality: NewMeasure: %w: unit must not be zero", ErrInvalidQualityMeasure)
	}
	if scale.IsZero() {
		return Measure{}, fmt.Errorf("quality: NewMeasure: %w: scale must not be zero", ErrInvalidQualityMeasure)
	}
	if method.IsZero() {
		return Measure{}, fmt.Errorf("quality: NewMeasure: %w: validation method must not be zero", ErrInvalidQualityMeasure)
	}
	return Measure{key: key, characteristic: characteristic, unit: unit, scale: scale, method: method}, nil
}

// WithRequiredEvidence returns a copy of m with its required-Evidence
// descriptions set to exactly the values given, in the order given. Each
// entry is trimmed and must be non-empty after trimming. Passing an empty or
// nil slice declares none, which is why no WithoutRequiredEvidence exists:
// WithRequiredEvidence(nil) already expresses removal, and a second method
// would create a second validation path for the same field.
//
// These are descriptions of the Evidence a measurement is expected to rely
// upon, not citations. At configuration time that Evidence does not yet
// exist, so there is no Artifact Revision to reference; a Measurement Record
// cites the Evidence that actually materialized, using
// core.EvidenceArtifactRevisionRef inherited from PEOS-006. This mirrors
// exactly how validation.PlannedActivity.WithExpectedEvidence represents the
// same planning-time-versus-execution-time distinction, and is why an
// existing core Evidence reference type is deliberately not used here:
// PEOS-007 says a Measure identifies "the Evidence required as input", and
// requiring a resolvable Artifact Revision would make it impossible to
// configure a Profile before any Evidence exists. Defining a structured
// Evidence-kind vocabulary instead would invent an ontology PEOS-007 does not
// state.
func (m Measure) WithRequiredEvidence(descriptions []string) (Measure, error) {
	if len(descriptions) == 0 {
		m.requiredEvidence = nil
		return m, nil
	}
	cp := make([]string, len(descriptions))
	for idx, d := range descriptions {
		trimmed, err := trimmedRequired("Measure.WithRequiredEvidence", "required evidence entry", d, ErrInvalidQualityMeasure)
		if err != nil {
			return Measure{}, err
		}
		cp[idx] = trimmed
	}
	m.requiredEvidence = cp
	return m, nil
}

// WithUncertaintyHandling returns a copy of m with its statement of how
// uncertainty is handled set. handling must be non-empty after trimming; the
// trimmed value is stored. Use WithoutUncertaintyHandling to clear it.
//
// The statement is an opaque string. PEOS-007 requires a Measure to identify
// "how uncertainty is handled" without defining any uncertainty model, and
// inventing one here would be a framework the specification does not state.
func (m Measure) WithUncertaintyHandling(handling string) (Measure, error) {
	trimmed, err := trimmedRequired("Measure.WithUncertaintyHandling", "uncertainty handling", handling, ErrInvalidQualityMeasure)
	if err != nil {
		return Measure{}, err
	}
	m.uncertaintyHandling = trimmed
	return m, nil
}

// WithoutUncertaintyHandling returns a copy of m with its uncertainty
// handling cleared.
func (m Measure) WithoutUncertaintyHandling() Measure {
	m.uncertaintyHandling = ""
	return m
}

// WithValidRange returns a copy of m with its valid range set. validRange
// must be non-empty after trimming; the trimmed value is stored. Use
// WithoutValidRange to clear it.
//
// The range is an opaque string for the same reason a Threshold's boundary
// value is: PEOS-007 defines no numeric type, no unit arithmetic, and no
// range grammar, so how the range is expressed and interpreted is
// Product-owned.
func (m Measure) WithValidRange(validRange string) (Measure, error) {
	trimmed, err := trimmedRequired("Measure.WithValidRange", "valid range", validRange, ErrInvalidQualityMeasure)
	if err != nil {
		return Measure{}, err
	}
	m.validRange = trimmed
	return m, nil
}

// WithoutValidRange returns a copy of m with its valid range cleared.
func (m Measure) WithoutValidRange() Measure {
	m.validRange = ""
	return m
}

// WithNormalizationRule returns a copy of m referencing, by profile-local
// key, the Normalization Rule applicable to its measured value. key must be
// non-zero; use WithoutNormalizationRule to clear it. Whether the key
// resolves to a Normalization Rule in the same Profile Revision is checked by
// NewProfileContent.
func (m Measure) WithNormalizationRule(key core.LocalKey) (Measure, error) {
	if key.IsZero() {
		return Measure{}, fmt.Errorf("quality: Measure.WithNormalizationRule: %w: normalization rule key must not be zero", ErrInvalidQualityMeasure)
	}
	m.normalizationRule = key
	return m, nil
}

// WithoutNormalizationRule returns a copy of m with its Normalization Rule
// reference cleared.
func (m Measure) WithoutNormalizationRule() Measure {
	m.normalizationRule = core.LocalKey{}
	return m
}

// WithExtension returns a copy of m with its extension data set. Passing the
// zero core.Extension is equivalent to declaring none.
//
// Extension is where Product-specific measurement parameters belong. PEOS-007
// defines no measurement parameter model and this package introduces none.
func (m Measure) WithExtension(extension core.Extension) Measure {
	m.extension = extension
	return m
}

// WithoutExtension returns a copy of m with its extension data cleared.
func (m Measure) WithoutExtension() Measure {
	m.extension = core.Extension{}
	return m
}

// Key returns m's profile-local key.
func (m Measure) Key() core.LocalKey { return m.key }

// Characteristic returns the profile-local key of the Characteristic m
// measures. It is mandatory and therefore never absent on a valid Measure.
func (m Measure) Characteristic() core.LocalKey { return m.characteristic }

// Unit returns m's declared unit.
func (m Measure) Unit() Unit { return m.unit }

// Scale returns m's declared scale.
func (m Measure) Scale() Scale { return m.scale }

// Method returns the PEOS-006 Validation Method m uses to obtain a value.
func (m Measure) Method() core.ValidationMethod { return m.method }

// RequiredEvidence returns a defensive copy of m's required-Evidence
// descriptions, in declaration order.
func (m Measure) RequiredEvidence() []string {
	if len(m.requiredEvidence) == 0 {
		return nil
	}
	cp := make([]string, len(m.requiredEvidence))
	copy(cp, m.requiredEvidence)
	return cp
}

// UncertaintyHandling returns m's statement of how uncertainty is handled,
// and whether one is set.
func (m Measure) UncertaintyHandling() (string, bool) {
	return m.uncertaintyHandling, m.uncertaintyHandling != ""
}

// ValidRange returns m's declared valid range, and whether one is set.
func (m Measure) ValidRange() (string, bool) { return m.validRange, m.validRange != "" }

// NormalizationRule returns the profile-local key of m's applicable
// Normalization Rule, and whether one is set.
func (m Measure) NormalizationRule() (core.LocalKey, bool) {
	return m.normalizationRule, !m.normalizationRule.IsZero()
}

// Extension returns m's extension data.
func (m Measure) Extension() core.Extension { return m.extension }

// IsZero reports whether m is the zero value.
func (m Measure) IsZero() bool {
	return m.key.IsZero() && m.characteristic.IsZero() && m.unit.IsZero() &&
		m.scale.IsZero() && m.method.IsZero()
}

type measureJSON struct {
	Key                 core.LocalKey         `json:"key"`
	Characteristic      core.LocalKey         `json:"characteristic"`
	Unit                Unit                  `json:"unit"`
	Scale               Scale                 `json:"scale"`
	Method              core.ValidationMethod `json:"method"`
	RequiredEvidence    []string              `json:"required_evidence,omitempty"`
	UncertaintyHandling string                `json:"uncertainty_handling,omitempty"`
	ValidRange          string                `json:"valid_range,omitempty"`
	NormalizationRule   *core.LocalKey        `json:"normalization_rule,omitempty"`
	Extension           *core.Extension       `json:"extension,omitempty"`
}

// measureUnmarshalJSON mirrors measureJSON for decoding, capturing the three
// optional single values as raw bytes so an explicit null can be
// distinguished from an absent key and rejected. required_evidence
// deliberately does not get that treatment: absent, null, and [] all denote
// the same valid state, "none required".
type measureUnmarshalJSON struct {
	Key                 core.LocalKey         `json:"key"`
	Characteristic      core.LocalKey         `json:"characteristic"`
	Unit                Unit                  `json:"unit"`
	Scale               Scale                 `json:"scale"`
	Method              core.ValidationMethod `json:"method"`
	RequiredEvidence    []string              `json:"required_evidence"`
	UncertaintyHandling json.RawMessage       `json:"uncertainty_handling"`
	ValidRange          json.RawMessage       `json:"valid_range"`
	NormalizationRule   json.RawMessage       `json:"normalization_rule"`
	Extension           *core.Extension       `json:"extension,omitempty"`
}

// MarshalJSON encodes m with its five mandatory keys always present, plus
// whichever optional keys are set. There is no "id", revision, lifecycle,
// relation, or provenance key.
func (m Measure) MarshalJSON() ([]byte, error) {
	if m.IsZero() {
		return nil, fmt.Errorf("quality: marshal Measure: %w", ErrInvalidQualityMeasure)
	}
	raw := measureJSON{
		Key:                 m.key,
		Characteristic:      m.characteristic,
		Unit:                m.unit,
		Scale:               m.scale,
		Method:              m.method,
		UncertaintyHandling: m.uncertaintyHandling,
		ValidRange:          m.validRange,
	}
	if len(m.requiredEvidence) > 0 {
		raw.RequiredEvidence = m.requiredEvidence
	}
	if !m.normalizationRule.IsZero() {
		raw.NormalizationRule = &m.normalizationRule
	}
	if !m.extension.IsZero() {
		raw.Extension = &m.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes m from its JSON form, applying the same validation as
// NewMeasure and each With* method.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - key, characteristic, unit, scale, method: a missing key leaves the
//     field zero and is rejected with ErrInvalidQualityMeasure. An explicit
//     null invokes the nested type's own UnmarshalJSON where it has one
//     (core.LocalKey fails with core.ErrEmptyIdentity); for Unit, Scale, and
//     Method a null leaves the wrapped vocabulary value zero, which
//     NewMeasure then rejects. All are rejected; the sentinel sets differ.
//   - required_evidence: absent, explicit null, and empty array are
//     equivalent and all mean "none required".
//   - uncertainty_handling, valid_range, normalization_rule: a missing key
//     means absent; an explicit null is rejected rather than silently
//     treated as absent.
//   - extension: null is equivalent to absent.
func (m *Measure) UnmarshalJSON(data []byte) error {
	var raw measureUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal Measure: %w: %w", ErrInvalidQualityMeasure, err)
	}

	result, err := NewMeasure(raw.Key, raw.Characteristic, raw.Unit, raw.Scale, raw.Method)
	if err != nil {
		return err
	}
	if len(raw.RequiredEvidence) > 0 {
		if result, err = result.WithRequiredEvidence(raw.RequiredEvidence); err != nil {
			return err
		}
	}
	if len(raw.UncertaintyHandling) > 0 {
		if err = rejectNullRaw("Measure", "uncertainty handling", raw.UncertaintyHandling, ErrInvalidQualityMeasure); err != nil {
			return err
		}
		var handling string
		if err = json.Unmarshal(raw.UncertaintyHandling, &handling); err != nil {
			return fmt.Errorf("quality: unmarshal Measure: %w: %w", ErrInvalidQualityMeasure, err)
		}
		if result, err = result.WithUncertaintyHandling(handling); err != nil {
			return err
		}
	}
	if len(raw.ValidRange) > 0 {
		if err = rejectNullRaw("Measure", "valid range", raw.ValidRange, ErrInvalidQualityMeasure); err != nil {
			return err
		}
		var validRange string
		if err = json.Unmarshal(raw.ValidRange, &validRange); err != nil {
			return fmt.Errorf("quality: unmarshal Measure: %w: %w", ErrInvalidQualityMeasure, err)
		}
		if result, err = result.WithValidRange(validRange); err != nil {
			return err
		}
	}
	if len(raw.NormalizationRule) > 0 {
		if err = rejectNullRaw("Measure", "normalization rule", raw.NormalizationRule, ErrInvalidQualityMeasure); err != nil {
			return err
		}
		var key core.LocalKey
		if err = json.Unmarshal(raw.NormalizationRule, &key); err != nil {
			return fmt.Errorf("quality: unmarshal Measure: %w: %w", ErrInvalidQualityMeasure, err)
		}
		if result, err = result.WithNormalizationRule(key); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*m = result
	return nil
}

// --- Threshold ---------------------------------------------------------------

// Threshold is a PEOS-007 Threshold: "a Quality Profile Revision-owned value
// structure defining a boundary used for classification or for determining a
// Quality Claim outcome."
//
// A Threshold names the Measure it bounds, the comparison it applies, and the
// boundary value itself. It stores no pass/fail result, no outcome, and no
// current quality state: it defines the boundary, and a Quality Claim
// recorded against it carries the outcome. Storing an outcome here would be
// the non-conforming pattern "Mutable Quality Score" in miniature, and would
// also make a Threshold a derived-state holder, which the Derived Quality
// State Invariant forbids.
//
// The boundary value is an opaque, trimmed string. PEOS-007 defines no
// numeric type, no unit arithmetic, and no comparison grammar, and its
// Non-Goals disclaim "a specific scoring formula or weighting scheme".
// Inventing a numeric framework here -- a float, a typed quantity, an
// expression language -- would be a model the specification does not state.
// How the value is parsed and how the operator is applied are Product-owned.
type Threshold struct {
	key      core.LocalKey
	measure  core.LocalKey
	operator ThresholdOperator
	value    string

	description string
	extension   core.Extension
}

// NewThreshold validates its four arguments and returns a Threshold with no
// description and no extension data.
//
// All four are mandatory. key, measure, and operator must be non-zero, and
// value must be non-empty after trimming; the trimmed value is stored.
// Whether measure resolves to a Measure in the same Profile Revision is
// checked by NewProfileContent.
func NewThreshold(
	key core.LocalKey,
	measure core.LocalKey,
	operator ThresholdOperator,
	value string,
) (Threshold, error) {
	if key.IsZero() {
		return Threshold{}, fmt.Errorf("quality: NewThreshold: %w: key must not be zero", ErrInvalidQualityThreshold)
	}
	if measure.IsZero() {
		return Threshold{}, fmt.Errorf("quality: NewThreshold: %w: measure key must not be zero", ErrInvalidQualityThreshold)
	}
	if operator.IsZero() {
		return Threshold{}, fmt.Errorf("quality: NewThreshold: %w: operator must not be zero", ErrInvalidQualityThreshold)
	}
	trimmed, err := trimmedRequired("NewThreshold", "value", value, ErrInvalidQualityThreshold)
	if err != nil {
		return Threshold{}, err
	}
	return Threshold{key: key, measure: measure, operator: operator, value: trimmed}, nil
}

// WithDescription returns a copy of t with an optional description set.
// description must be non-empty after trimming.
func (t Threshold) WithDescription(description string) (Threshold, error) {
	trimmed, err := trimmedRequired("Threshold.WithDescription", "description", description, ErrInvalidQualityThreshold)
	if err != nil {
		return Threshold{}, err
	}
	t.description = trimmed
	return t, nil
}

// WithoutDescription returns a copy of t with its description cleared.
func (t Threshold) WithoutDescription() Threshold {
	t.description = ""
	return t
}

// WithExtension returns a copy of t with its extension data set.
func (t Threshold) WithExtension(extension core.Extension) Threshold {
	t.extension = extension
	return t
}

// WithoutExtension returns a copy of t with its extension data cleared.
func (t Threshold) WithoutExtension() Threshold {
	t.extension = core.Extension{}
	return t
}

// Key returns t's profile-local key.
func (t Threshold) Key() core.LocalKey { return t.key }

// Measure returns the profile-local key of the Measure t bounds.
func (t Threshold) Measure() core.LocalKey { return t.measure }

// Operator returns the comparison t applies between a measured value and its
// boundary.
func (t Threshold) Operator() ThresholdOperator { return t.operator }

// Value returns t's boundary value, uninterpreted.
func (t Threshold) Value() string { return t.value }

// Description returns t's optional description, and whether one is set.
func (t Threshold) Description() (string, bool) { return t.description, t.description != "" }

// Extension returns t's extension data.
func (t Threshold) Extension() core.Extension { return t.extension }

// IsZero reports whether t is the zero value.
func (t Threshold) IsZero() bool {
	return t.key.IsZero() && t.measure.IsZero() && t.operator.IsZero() && t.value == ""
}

type thresholdJSON struct {
	Key         core.LocalKey     `json:"key"`
	Measure     core.LocalKey     `json:"measure"`
	Operator    ThresholdOperator `json:"operator"`
	Value       string            `json:"value"`
	Description string            `json:"description,omitempty"`
	Extension   *core.Extension   `json:"extension,omitempty"`
}

type thresholdUnmarshalJSON struct {
	Key         core.LocalKey     `json:"key"`
	Measure     core.LocalKey     `json:"measure"`
	Operator    ThresholdOperator `json:"operator"`
	Value       string            `json:"value"`
	Description json.RawMessage   `json:"description"`
	Extension   *core.Extension   `json:"extension,omitempty"`
}

// MarshalJSON encodes t with its four mandatory keys always present. There
// is no "satisfied", "verdict", "outcome", "status", or "score" key: a
// Threshold defines a boundary and never records a result.
func (t Threshold) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return nil, fmt.Errorf("quality: marshal Threshold: %w", ErrInvalidQualityThreshold)
	}
	raw := thresholdJSON{
		Key:         t.key,
		Measure:     t.measure,
		Operator:    t.operator,
		Value:       t.value,
		Description: t.description,
	}
	if !t.extension.IsZero() {
		raw.Extension = &t.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes t from its JSON form, applying the same validation as
// NewThreshold and each With* method.
//
// Missing-versus-null behavior: key, measure, and operator are rejected when
// absent (zero value) and when explicitly null (core.LocalKey's own decode
// fails with core.ErrEmptyIdentity; a null operator leaves its vocabulary
// value zero, which NewThreshold rejects). value is rejected when absent,
// null, or whitespace-only, all yielding the empty string. description is
// absent when its key is missing and rejected when explicitly null.
// extension treats null as absent.
func (t *Threshold) UnmarshalJSON(data []byte) error {
	var raw thresholdUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal Threshold: %w: %w", ErrInvalidQualityThreshold, err)
	}
	result, err := NewThreshold(raw.Key, raw.Measure, raw.Operator, raw.Value)
	if err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		if err = rejectNullRaw("Threshold", "description", raw.Description, ErrInvalidQualityThreshold); err != nil {
			return err
		}
		var description string
		if err = json.Unmarshal(raw.Description, &description); err != nil {
			return fmt.Errorf("quality: unmarshal Threshold: %w: %w", ErrInvalidQualityThreshold, err)
		}
		if result, err = result.WithDescription(description); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*t = result
	return nil
}

// --- Target ------------------------------------------------------------------

// Target is a PEOS-007 Target: "a Quality Profile Revision-owned value
// structure defining a desired value or range."
//
// Target is a separate Go type from Threshold, never a mode or flag of one,
// because PEOS-007 requires the separation: "A Target does not itself define
// pass/fail classification; a Target expresses intent, while a Threshold
// expresses the boundary used for a Claim outcome. The two SHALL NOT be
// conflated." Two consequences are structural rather than documented. First,
// Target carries no operator: a comparison is what turns a value into a
// boundary, and a Target is not a boundary. Second, Target carries no
// classification outcome, no pass/fail field, and no "met"/"achieved" flag --
// whether a Target was reached is derived from Measurement Records and
// Quality Claims, per the Derived Quality State Invariant, never stored here.
//
// Its desired value is an opaque, trimmed string, for the same reason a
// Threshold's boundary value is. PEOS-007 says "a desired value or range"
// without defining either form, so a single uninterpreted value covers both
// and lets the Product decide the notation.
type Target struct {
	key     core.LocalKey
	measure core.LocalKey
	value   string

	description string
	extension   core.Extension
}

// NewTarget validates its three arguments and returns a Target with no
// description and no extension data.
//
// All three are mandatory. key and measure must be non-zero, and value must
// be non-empty after trimming; the trimmed value is stored. Whether measure
// resolves to a Measure in the same Profile Revision is checked by
// NewProfileContent.
func NewTarget(key core.LocalKey, measure core.LocalKey, value string) (Target, error) {
	if key.IsZero() {
		return Target{}, fmt.Errorf("quality: NewTarget: %w: key must not be zero", ErrInvalidQualityTarget)
	}
	if measure.IsZero() {
		return Target{}, fmt.Errorf("quality: NewTarget: %w: measure key must not be zero", ErrInvalidQualityTarget)
	}
	trimmed, err := trimmedRequired("NewTarget", "value", value, ErrInvalidQualityTarget)
	if err != nil {
		return Target{}, err
	}
	return Target{key: key, measure: measure, value: trimmed}, nil
}

// WithDescription returns a copy of t with an optional description set.
// description must be non-empty after trimming.
func (t Target) WithDescription(description string) (Target, error) {
	trimmed, err := trimmedRequired("Target.WithDescription", "description", description, ErrInvalidQualityTarget)
	if err != nil {
		return Target{}, err
	}
	t.description = trimmed
	return t, nil
}

// WithoutDescription returns a copy of t with its description cleared.
func (t Target) WithoutDescription() Target {
	t.description = ""
	return t
}

// WithExtension returns a copy of t with its extension data set.
func (t Target) WithExtension(extension core.Extension) Target {
	t.extension = extension
	return t
}

// WithoutExtension returns a copy of t with its extension data cleared.
func (t Target) WithoutExtension() Target {
	t.extension = core.Extension{}
	return t
}

// Key returns t's profile-local key.
func (t Target) Key() core.LocalKey { return t.key }

// Measure returns the profile-local key of the Measure t expresses intent
// for.
func (t Target) Measure() core.LocalKey { return t.measure }

// Value returns t's desired value or range, uninterpreted.
func (t Target) Value() string { return t.value }

// Description returns t's optional description, and whether one is set.
func (t Target) Description() (string, bool) { return t.description, t.description != "" }

// Extension returns t's extension data.
func (t Target) Extension() core.Extension { return t.extension }

// IsZero reports whether t is the zero value.
func (t Target) IsZero() bool {
	return t.key.IsZero() && t.measure.IsZero() && t.value == ""
}

type targetJSON struct {
	Key         core.LocalKey   `json:"key"`
	Measure     core.LocalKey   `json:"measure"`
	Value       string          `json:"value"`
	Description string          `json:"description,omitempty"`
	Extension   *core.Extension `json:"extension,omitempty"`
}

type targetUnmarshalJSON struct {
	Key         core.LocalKey   `json:"key"`
	Measure     core.LocalKey   `json:"measure"`
	Value       string          `json:"value"`
	Description json.RawMessage `json:"description"`
	Extension   *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes t with its three mandatory keys always present. There
// is deliberately no "operator" key -- that is what distinguishes a Target's
// wire form from a Threshold's -- and no "satisfied", "met", "achieved",
// "outcome", "verdict", "status", or "score" key.
func (t Target) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return nil, fmt.Errorf("quality: marshal Target: %w", ErrInvalidQualityTarget)
	}
	raw := targetJSON{
		Key:         t.key,
		Measure:     t.measure,
		Value:       t.value,
		Description: t.description,
	}
	if !t.extension.IsZero() {
		raw.Extension = &t.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes t from its JSON form, applying the same validation as
// NewTarget and each With* method. Missing-versus-null behavior matches
// Threshold's, minus the operator.
func (t *Target) UnmarshalJSON(data []byte) error {
	var raw targetUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal Target: %w: %w", ErrInvalidQualityTarget, err)
	}
	result, err := NewTarget(raw.Key, raw.Measure, raw.Value)
	if err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		if err = rejectNullRaw("Target", "description", raw.Description, ErrInvalidQualityTarget); err != nil {
			return err
		}
		var description string
		if err = json.Unmarshal(raw.Description, &description); err != nil {
			return fmt.Errorf("quality: unmarshal Target: %w: %w", ErrInvalidQualityTarget, err)
		}
		if result, err = result.WithDescription(description); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*t = result
	return nil
}

// --- Constraint --------------------------------------------------------------

// Constraint is a PEOS-007 Quality Constraint: "a normative restriction
// contained in a Quality Profile Revision."
//
// A Constraint is not a Requirement. PEOS-007 permits a Quality Constraint to
// "also be represented as a Requirement, as defined by PEOS-005, where
// persistent Requirement identity, Lifecycle, Authority, Applicability,
// Allocation, or Requirement relationships are needed", but insists that
// doing so "is a deliberate, explicit modeling choice, made per constraint,
// not an automatic consequence of this specification", and that "Every
// Quality Constraint SHALL NOT be silently treated as a Requirement merely
// because it constrains engineering behavior."
//
// This type makes that structural rather than advisory. It has no
// Requirement identity, no Lifecycle, no Authority of its own, no
// Applicability, no Allocation, and no Artifact Relation, and there is no
// conversion in either direction between Constraint and
// requirement.Requirement -- peos/quality does not import peos/requirement at
// all, so no such conversion is even expressible. A Product that wants a
// constraint to be a Requirement constructs a requirement.Requirement
// directly, as its own deliberate act, and cites it as a Requirement
// criterion.
type Constraint struct {
	key       core.LocalKey
	statement string

	description string
	extension   core.Extension
}

// NewConstraint validates key and statement and returns a Constraint with no
// description and no extension data.
//
// Both are mandatory. key must be non-zero, and statement must be non-empty
// after trimming; the trimmed value is stored. The statement is the normative
// restriction itself, an opaque string: PEOS-007 defines no constraint
// grammar or expression language, and introducing one would be a DSL the
// specification does not state.
func NewConstraint(key core.LocalKey, statement string) (Constraint, error) {
	if key.IsZero() {
		return Constraint{}, fmt.Errorf("quality: NewConstraint: %w: key must not be zero", ErrInvalidQualityConstraint)
	}
	trimmed, err := trimmedRequired("NewConstraint", "statement", statement, ErrInvalidQualityConstraint)
	if err != nil {
		return Constraint{}, err
	}
	return Constraint{key: key, statement: trimmed}, nil
}

// WithDescription returns a copy of c with an optional description set.
// description must be non-empty after trimming.
func (c Constraint) WithDescription(description string) (Constraint, error) {
	trimmed, err := trimmedRequired("Constraint.WithDescription", "description", description, ErrInvalidQualityConstraint)
	if err != nil {
		return Constraint{}, err
	}
	c.description = trimmed
	return c, nil
}

// WithoutDescription returns a copy of c with its description cleared.
func (c Constraint) WithoutDescription() Constraint {
	c.description = ""
	return c
}

// WithExtension returns a copy of c with its extension data set.
func (c Constraint) WithExtension(extension core.Extension) Constraint {
	c.extension = extension
	return c
}

// WithoutExtension returns a copy of c with its extension data cleared.
func (c Constraint) WithoutExtension() Constraint {
	c.extension = core.Extension{}
	return c
}

// Key returns c's profile-local key.
func (c Constraint) Key() core.LocalKey { return c.key }

// Statement returns c's normative restriction, uninterpreted.
func (c Constraint) Statement() string { return c.statement }

// Description returns c's optional description, and whether one is set.
func (c Constraint) Description() (string, bool) { return c.description, c.description != "" }

// Extension returns c's extension data.
func (c Constraint) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value.
func (c Constraint) IsZero() bool { return c.key.IsZero() && c.statement == "" }

type constraintJSON struct {
	Key         core.LocalKey   `json:"key"`
	Statement   string          `json:"statement"`
	Description string          `json:"description,omitempty"`
	Extension   *core.Extension `json:"extension,omitempty"`
}

type constraintUnmarshalJSON struct {
	Key         core.LocalKey   `json:"key"`
	Statement   string          `json:"statement"`
	Description json.RawMessage `json:"description"`
	Extension   *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes c with its two mandatory keys always present. There is
// no "requirement", "requirement_id", lifecycle, authority, applicability,
// allocation, or relation key: their absence is what proves a Quality
// Constraint is not silently a Requirement.
func (c Constraint) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("quality: marshal Constraint: %w", ErrInvalidQualityConstraint)
	}
	raw := constraintJSON{Key: c.key, Statement: c.statement, Description: c.description}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation as
// NewConstraint and each With* method. key is rejected when absent (zero) and
// when explicitly null (core.ErrEmptyIdentity); statement is rejected when
// absent, null, or whitespace-only; description is absent when its key is
// missing and rejected when explicitly null; extension treats null as absent.
func (c *Constraint) UnmarshalJSON(data []byte) error {
	var raw constraintUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal Constraint: %w: %w", ErrInvalidQualityConstraint, err)
	}
	result, err := NewConstraint(raw.Key, raw.Statement)
	if err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		if err = rejectNullRaw("Constraint", "description", raw.Description, ErrInvalidQualityConstraint); err != nil {
			return err
		}
		var description string
		if err = json.Unmarshal(raw.Description, &description); err != nil {
			return fmt.Errorf("quality: unmarshal Constraint: %w: %w", ErrInvalidQualityConstraint, err)
		}
		if result, err = result.WithDescription(description); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*c = result
	return nil
}

// --- NormalizationRule and AggregationRule -----------------------------------

// NormalizationRule is a PEOS-007 Normalization Rule: "a Quality Profile
// Revision-owned value structure describing how a raw Measurement Record
// value is transformed into a comparable form." It "has no independent
// identity, revision, or lifecycle outside its owning Profile Revision."
//
// The rule is carried as a mandatory description, not as an executable
// formula. PEOS-007 defines no transformation language, and its Non-Goals
// disclaim "a specific scoring formula or weighting scheme"; a formula DSL
// or an evaluation engine here would be a mechanism the specification does
// not state. This package therefore never computes a normalized value --
// interpreting and applying the rule is Product-owned.
type NormalizationRule struct {
	key         core.LocalKey
	description string
	extension   core.Extension
}

// NewNormalizationRule validates key and description and returns a
// NormalizationRule with no extension data. Both arguments are mandatory: key
// must be non-zero, and description must be non-empty after trimming; the
// trimmed value is stored. The description is the rule -- it is not an
// optional annotation, which is why there is no WithDescription or
// WithoutDescription.
func NewNormalizationRule(key core.LocalKey, description string) (NormalizationRule, error) {
	if key.IsZero() {
		return NormalizationRule{}, fmt.Errorf("quality: NewNormalizationRule: %w: key must not be zero", ErrInvalidQualityRule)
	}
	trimmed, err := trimmedRequired("NewNormalizationRule", "description", description, ErrInvalidQualityRule)
	if err != nil {
		return NormalizationRule{}, err
	}
	return NormalizationRule{key: key, description: trimmed}, nil
}

// WithExtension returns a copy of r with its extension data set.
func (r NormalizationRule) WithExtension(extension core.Extension) NormalizationRule {
	r.extension = extension
	return r
}

// WithoutExtension returns a copy of r with its extension data cleared.
func (r NormalizationRule) WithoutExtension() NormalizationRule {
	r.extension = core.Extension{}
	return r
}

// Key returns r's profile-local key.
func (r NormalizationRule) Key() core.LocalKey { return r.key }

// Description returns r's transformation description. It is mandatory and
// therefore never absent on a valid NormalizationRule.
func (r NormalizationRule) Description() string { return r.description }

// Extension returns r's extension data.
func (r NormalizationRule) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r NormalizationRule) IsZero() bool { return r.key.IsZero() && r.description == "" }

type qualityRuleJSON struct {
	Key         core.LocalKey   `json:"key"`
	Description string          `json:"description"`
	Extension   *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"key":...,"description":...} plus extension when
// set. There is no "formula", "expression", "weight", or "score" key.
func (r NormalizationRule) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("quality: marshal NormalizationRule: %w", ErrInvalidQualityRule)
	}
	raw := qualityRuleJSON{Key: r.key, Description: r.description}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes r from its JSON form, applying the same validation as
// NewNormalizationRule. key is rejected when absent (zero value) and when
// explicitly null (core.ErrEmptyIdentity from core.LocalKey's own decode);
// description is rejected when absent, null, or whitespace-only, all of which
// yield the empty string; extension treats null as absent. Because
// description is mandatory, no raw-bytes probe is needed to tell absent from
// null -- both are errors.
func (r *NormalizationRule) UnmarshalJSON(data []byte) error {
	var raw qualityRuleJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal NormalizationRule: %w: %w", ErrInvalidQualityRule, err)
	}
	result, err := NewNormalizationRule(raw.Key, raw.Description)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*r = result
	return nil
}

// AggregationRule is a PEOS-007 Aggregation Rule: "a Quality Profile
// Revision-owned value structure describing how multiple Measurement Records
// or Quality Claims are combined into a single derived view." It "has no
// independent identity, revision, or lifecycle outside its owning Profile
// Revision."
//
// PEOS-007 is explicit about what an Aggregation Rule must not become: "An
// Aggregation Rule produces a derived view. It does not itself produce a
// stored, mutable field on any Subject." This type accordingly stores no
// aggregate, no score, no cached result, and no list of the records it was
// applied to; it holds only its key and its description. A consumer needing a
// "current" aggregate "MUST compute it, on demand, from the applicable,
// non-replaced, non-invalidated Measurement Records and Quality Claims" --
// which is why this package provides no aggregation function and no scoring
// engine.
//
// AggregationRule and NormalizationRule are structurally identical and share
// one sentinel, but remain distinct Go types so that one can never be
// supplied where the other is expected. They are also kept in separate
// ProfileContent collections with separate key namespaces, and only
// NormalizationRule is referenceable from a Measure.
type AggregationRule struct {
	key         core.LocalKey
	description string
	extension   core.Extension
}

// NewAggregationRule validates key and description and returns an
// AggregationRule with no extension data. Both arguments are mandatory, with
// the same contract as NewNormalizationRule.
func NewAggregationRule(key core.LocalKey, description string) (AggregationRule, error) {
	if key.IsZero() {
		return AggregationRule{}, fmt.Errorf("quality: NewAggregationRule: %w: key must not be zero", ErrInvalidQualityRule)
	}
	trimmed, err := trimmedRequired("NewAggregationRule", "description", description, ErrInvalidQualityRule)
	if err != nil {
		return AggregationRule{}, err
	}
	return AggregationRule{key: key, description: trimmed}, nil
}

// WithExtension returns a copy of r with its extension data set.
func (r AggregationRule) WithExtension(extension core.Extension) AggregationRule {
	r.extension = extension
	return r
}

// WithoutExtension returns a copy of r with its extension data cleared.
func (r AggregationRule) WithoutExtension() AggregationRule {
	r.extension = core.Extension{}
	return r
}

// Key returns r's profile-local key.
func (r AggregationRule) Key() core.LocalKey { return r.key }

// Description returns r's combination description. It is mandatory and
// therefore never absent on a valid AggregationRule.
func (r AggregationRule) Description() string { return r.description }

// Extension returns r's extension data.
func (r AggregationRule) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r AggregationRule) IsZero() bool { return r.key.IsZero() && r.description == "" }

// MarshalJSON encodes r as {"key":...,"description":...} plus extension when
// set. There is no "formula", "weight", "score", "aggregate", "current", or
// "latest" key -- the last four would be exactly the stored derived state
// PEOS-007 forbids.
func (r AggregationRule) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("quality: marshal AggregationRule: %w", ErrInvalidQualityRule)
	}
	raw := qualityRuleJSON{Key: r.key, Description: r.description}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes r from its JSON form, applying the same validation as
// NewAggregationRule. Missing-versus-null behavior matches
// NormalizationRule.UnmarshalJSON exactly.
func (r *AggregationRule) UnmarshalJSON(data []byte) error {
	var raw qualityRuleJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal AggregationRule: %w: %w", ErrInvalidQualityRule, err)
	}
	result, err := NewAggregationRule(raw.Key, raw.Description)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*r = result
	return nil
}
