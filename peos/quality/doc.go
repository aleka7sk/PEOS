// Package quality implements the PEOS-007 Quality Model: the vocabulary and
// configuration through which engineering quality expectations are declared,
// and the quality-specific specializations of the PEOS-006 mechanisms that
// measure and claim against them.
//
// PEOS-007 draws the boundary with PEOS-006 exactly: "PEOS-006 owns the
// mechanism (Plan, Planned Activity, Method, Execution Record, Evidence,
// Claim, replacement semantics, derivation-not-storage). PEOS-007 owns the
// vocabulary and configuration (which Characteristics, Measures, Thresholds,
// and Targets apply to which Subjects) and the resulting specialized Claim
// type." This package implements only its own side of that boundary and
// reuses PEOS-006's mechanism rather than restating it.
//
// # Ontology
//
// PEOS-007 introduces exactly one Artifact and no new identity space:
//
//	Quality Profile          Artifact (PEOS-002), ordinary Artifact Revision
//	  ProfileRevision        an Artifact Revision of that Artifact
//	    ProfileContent       the Revision's typed normative content
//	      Characteristic     Revision-owned value  (two scopings)
//	      Measure            Revision-owned value
//	      Threshold          Revision-owned value
//	      Target             Revision-owned value
//	      Constraint         Revision-owned value
//	      NormalizationRule  Revision-owned value
//	      AggregationRule    Revision-owned value
//
//	MeasurementRecord        specializes validation.ExecutionRecord   (Packet I.2)
//	Claim                    specializes validation.Claim             (Packet I.2)
//
// Profile is the only type here with a PEOS identity, and that identity is an
// ordinary core.ArtifactID -- there is no Quality Profile identity space, no
// Quality Profile Version, and no Quality Characteristic or Quality Measure
// identity of any kind.
//
// # Quality Profile Artifact semantics
//
// A Quality Profile "is an Artifact as defined by PEOS-002" and "uses
// ordinary Artifact Revision for all of its evolution. There is no Quality
// Profile Version distinct from Artifact Revision." Profile therefore wraps a
// core.Artifact and adds no field; ProfileRevision wraps a
// core.ArtifactRevision and pairs it with ProfileContent. Neither exposes a
// version, a lifecycle, a status, or a content setter: "Modification of a
// Quality Profile's content constitutes a content change and SHALL create a
// new Artifact Revision."
//
// # Revision-owned value semantics
//
// All seven owned value structures are value content, not entities. None has
// a PEOS identity, a Ref, an Artifact, an Artifact Revision, a revision
// system, a lifecycle, or its own provenance -- provenance belongs to the
// owning ProfileContent, which records the origin of the Revision as a whole.
// PEOS-007's Profile-Owned Rule Invariant states this for Threshold, Target,
// Quality Constraint, Normalization Rule, and Aggregation Rule; its
// Characteristic Scope and Measure Scope Invariants state the equivalent for
// Characteristic and Measure.
//
// Characteristic is the one owned value with two scopings, because PEOS-007
// gives it two: "the exact owning Quality Profile Revision in which it is
// defined; or an exact externally referenced normative vocabulary". They are
// exclusive -- a value carrying both would have two competing sources of
// meaning, and one carrying neither would have none.
//
// # Per-kind profile-local key namespaces
//
// Every owned value carries a core.LocalKey that is meaningful only within its
// owning Profile Revision. Keys are unique *within each collection* and may
// repeat *across* collections: one Profile Revision may define a
// Characteristic and a Threshold that share a key.
//
// This is a derived rule, and its derivation is deliberately minimal.
// PEOS-007 states no key uniqueness rule at all -- the word "unique" does not
// appear in the specification, and unlike PEOS-006, which states plan-local
// key uniqueness explicitly, PEOS-007 establishes only Revision ownership.
// Per-kind uniqueness is nonetheless necessary: a criterion citing a
// Characteristic by key must resolve to exactly one Characteristic, and two
// would make it unresolvable. Cross-kind uniqueness is not necessary, because
// every reference already determines its target collection -- internal
// references by the field that holds them, and external criterion citations by
// core.CriterionRef's kind discriminator. Asserting it would enforce a rule
// the specification does not state.
//
// # Internal reference resolution
//
// Four references inside one ProfileContent resolve by profile-local key, each
// in exactly one target collection:
//
//	Measure.Characteristic()       -> Characteristics only
//	Measure.NormalizationRule()    -> Normalization Rules only
//	Threshold.Measure()            -> Measures only
//	Target.Measure()               -> Measures only
//
// A key present only in a different collection does not satisfy resolution: a
// Threshold whose measure key names a Characteristic is rejected with
// ErrUnknownProfileLocalKey. Resolution is order-independent -- every
// collection's full key set is built before any reference is checked -- and all
// of it runs through one shared validation path used by NewProfileContent,
// every collection modifier, and ProfileContent.UnmarshalJSON, so the rules
// cannot drift between construction, modification, and decoding.
//
// PEOS-007 states no cycle policy and no self-reference prohibition for these
// references, and none is enforced, exactly as validation.NewPlanContent
// declines to reject Planned Activity dependency cycles. This package provides
// no graph or cycle API.
//
// # Citing an owned value as a Quality Claim criterion
//
// A Profile-owned value becomes a criterion by pairing its owning Profile
// Revision's reference with its local key:
//
//	revisionRef, err := profileRevision.Ref()
//	elementRef, err := core.NewQualityElementCriterionRef(revisionRef, characteristic.Key())
//	criterion, err := core.CriterionRefFromQualityCharacteristic(elementRef)
//
// PEOS-007 permits five of the seven owned kinds as criteria -- Characteristic,
// Measure, Threshold, Target, and Quality Constraint -- and peos/core has a
// dedicated criterion kind for each. Normalization and Aggregation Rules are
// not criteria; they describe how values are transformed and combined.
//
// Threshold, Target, and Quality Constraint criterion kinds were added to
// peos/core by the same packet that created this package. They were added as
// dedicated kinds rather than routed through core.NewOpaqueCriterionRef
// because a (Profile Revision, local key) composite provably cannot round-trip
// through that opaque path, as core/criterion.go documents.
//
// # No lifecycle
//
// Nothing in this package assigns, carries, or derives a Lifecycle State or
// State Assignment. "A Quality Claim does not itself assign a Lifecycle State
// or State Assignment. Lifecycle effects remain exclusively governed by
// PEOS-003." The non-import of peos/lifecycle is the structural guarantee of
// that separation, and it is what prevents the non-conforming pattern "Quality
// Outcome Used as Lifecycle State". A Lifecycle Transition MAY require a
// Quality Claim as Transition Evidence; that requirement is expressed in
// peos/lifecycle, not here.
//
// # No relation
//
// PEOS-007 defines no Artifact Relation, so this package does not import
// peos/relation and defines no relation, source, or target field anywhere. In
// particular, correction of a Measurement Record is never Artifact
// Supersession: PEOS-006's "new Claim -> affected earlier Claim" replacement
// model is inherited unchanged.
//
// # No stored derived state
//
// No type here stores a quality score, an aggregate, a pass/fail, an outcome,
// or any "current", "latest", or "effective" quality value. PEOS-007 is
// explicit: a quality score "SHALL NOT be stored as globally current mutable
// state on: an Artifact; an Artifact Revision; a Requirement; a Quality
// Profile", and "Any consumer requiring a 'current' quality score MUST compute
// it, on demand, from the applicable, non-replaced, non-invalidated Measurement
// Records and Quality Claims." An AggregationRule accordingly describes how to
// combine records; it never holds a combined result. This package provides no
// aggregation function and no scoring engine, and its wire forms carry no such
// key.
//
// Quality Evaluation is likewise not a type here. "Quality Evaluation is not an
// independent entity. It has no independent identity, no revision system, no
// lifecycle, and no mutable state of its own" -- it is the combination of a
// Profile Revision, a Planned Validation Activity, a Measurement Record,
// Evidence, and a Quality Claim, each owned where it already lives.
//
// # Product-owned interpretation
//
// PEOS-007's Non-Goals disclaim "a universal quality metric catalog" and "a
// specific scoring formula or weighting scheme". This package accordingly
// predeclares no unit, no scale, no comparison operator, and no
// Characteristic term, and it interprets none of the following:
//
//   - a Characteristic's term and an external vocabulary's meaning;
//   - a Unit, a Scale, and any relationship between them (there is no
//     dimensional-analysis or unit-conversion framework here);
//   - a ThresholdOperator, and how a measured value is compared against a
//     boundary;
//   - a Threshold's boundary value, a Target's desired value or range, and a
//     Measure's valid range -- all opaque, trimmed strings, because PEOS-007
//     defines no numeric type and no range grammar;
//   - a Measure's uncertainty handling, and its required Evidence
//     descriptions;
//   - a Quality Constraint's statement -- there is no constraint expression
//     language;
//   - a Normalization Rule's transformation and an Aggregation Rule's
//     combination -- both are descriptions, never executable formulas;
//   - core.Scope's expression, in a Profile's scope or its applicability.
//
// Each of those is Product-owned. Encoding any of them would be a framework
// PEOS-007 deliberately does not define.
//
// A Quality Constraint is also not automatically a Requirement. PEOS-007
// permits representing one as a Requirement "where persistent Requirement
// identity, Lifecycle, Authority, Applicability, Allocation, or Requirement
// relationships are needed", but only as "a deliberate, explicit modeling
// choice, made per constraint". Constraint therefore has no Requirement
// identity and no conversion to requirement.Requirement -- this package does
// not import peos/requirement, so no such conversion is expressible. A Product
// that wants a constraint to be a Requirement constructs one directly and
// cites it as a Requirement criterion.
//
// # Package dependency boundary
//
// Production sources may import only the standard library, peos/core, and
// peos/validation. peos/validation is needed by Packet I.2, where
// MeasurementRecord composes validation.ExecutionRecord and Claim composes
// validation.Claim; Packet I.1's own sources do not import it, and an unused
// dependency is not added merely to match the eventual graph.
//
// peos/relation, peos/lifecycle, peos/requirement, and peos/decision are all
// excluded, each for a reason stated above. Nothing imports peos/quality: it
// is a leaf, and the converse direction is asserted by test for every other
// package, which is what makes an import cycle inexpressible.
//
// # Repository responsibilities
//
// This package models values and enforces PEOS-007's structural invariants.
// It does not persist, index, query, or resolve anything across Profile
// Revisions. A repository built on it owns:
//
//   - storing Profiles, Profile Revisions, Measurement Records, and Quality
//     Claims, and retrieving them by identity;
//   - deciding which Profile Revision is applicable to a given Subject, by
//     interpreting scope, applicability, subjects, and subject types -- this
//     package interprets none of them;
//   - computing derived quality state on demand from applicable, non-replaced,
//     non-invalidated Measurement Records and Quality Claims, including
//     applying Normalization and Aggregation Rules;
//   - deciding whether a profile-local key in an earlier Profile Revision
//     denotes the same value as the same key in a later one. It does not by
//     default: a key "has no independent identity outside its owning Profile
//     Revision", and this package never compares keys across Revisions.
//
// # Packet scope
//
// Packet I.1 implements the Quality Profile side: the Artifact Type, Profile,
// ProfileRevision, ProfileApplicability, ProfileContent, the seven owned value
// structures, and the three quality-local vocabulary wrappers, plus the three
// additive core.CriterionRef arms.
//
// Packet I.2 will add MeasurementRecord -- composing validation.ExecutionRecord
// and adding the observed value, unit, and scale PEOS-007 mandates and PEOS-006
// has no field for -- and Claim, a thin wrapper over validation.Claim that fixes
// the Claim Type to core.ClaimTypeQuality and enforces PEOS-007's rule that a
// Requirement-subject Quality Claim may not cite that same Requirement as its
// own criterion. Their two sentinels, ErrInvalidMeasurementRecord and
// ErrInvalidQualityClaim, are already declared in errors.go so that I.2 need
// not reopen it.
//
// Neither is implemented yet. peos/validation is PEOS-006-correct in accepting
// a reflexive Requirement on a non-Satisfaction Claim, and Packet I.2 must not
// change it: the stricter PEOS-007 rule belongs here.
package quality
