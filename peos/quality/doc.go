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
//	MeasurementRecord        specializes validation.ExecutionRecord
//	Claim                    specializes validation.Claim
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
// # Measurement Record
//
// MeasurementRecord is a PEOS-007 Measurement Record: it "specializes the
// Validation Execution Record defined by PEOS-006", is "an immutable record",
// "is not an Artifact", and "has no revisions and no lifecycle."
//
// It composes validation.ExecutionRecord by named field and adds exactly the
// three fields PEOS-007 mandates and PEOS-006 has no field for -- the observed
// value, the unit, and the scale -- plus an optional quality-specific
// extension. It duplicates nothing the composed record already owns:
//
//   - identity is inherited. A Measurement Record's ID is the composed
//     record's own core.ValidationExecutionRecordID; this type declares no ID
//     field and mints no identity;
//   - subject, activity reference, method, execution outcome, timestamps,
//     actor, provenance, criteria, and Evidence are all the composed record's;
//   - correction is the composed record's. "Correction of a Measurement Record
//     creates a new Measurement Record, in accordance with PEOS-006's
//     Validation Execution Record correction rules" -- there is no PEOS-007
//     correction mechanism, and correction is never Artifact Supersession.
//
// PEOS-007 adds one invariant of its own: the composed record must cite at
// least one quality_characteristic criterion and at least one quality_measure
// criterion, in any order, because a Measurement Record SHALL identify "the
// exact Quality Characteristic and Quality Measure references applied".
// Additional criteria of any other kind -- Threshold, Target, Constraint,
// Requirement, Product or external rule -- are permitted; the specification
// states no exclusivity and no maximum. Criteria are not deduplicated, because
// PEOS-007 states no uniqueness rule for them.
//
// A Measurement Record is immutable raw history, never current state. It stores
// no score, no normalized value, no aggregate, and no "current" or "latest"
// marker. PEOS-007 permits a quality score to appear "as a value recorded on a
// Measurement Record" -- ObservedValue is that -- while forbidding one stored
// "as globally current mutable state". Normalization is not applied here
// either: a Normalization Rule is a description this package never executes.
//
// The observed value is an opaque string, for the same reason a Threshold's
// boundary value is: PEOS-007 defines no value type, and a float would exclude
// ordinal and categorical measures while a typed quantity would require the
// units framework the specification declines to define.
//
// # Quality Claim
//
// Claim is a PEOS-007 Quality Claim: "a specialization of Validation Claim, as
// defined by PEOS-006", which "exists exclusively as an instance of the
// PEOS-006 Validation Claim mechanism". The type name is Claim, not
// QualityClaim -- within this package the qualifier is already carried by the
// package name, and a QualityClaim type would read as the second Claim base
// model the non-conforming pattern "Parallel Quality Claim Base" forbids.
//
// It composes validation.Claim by named field and adds no stored field at all.
// Its Claim Type is fixed to core.ClaimTypeQuality: NewClaim supplies it, there
// is no claimType parameter and no WithClaimType, and every construction,
// adoption, modification, and decode re-verifies it. Identity, immutability,
// the absence of revisions and lifecycle, the exactly-one-Subject rule, the
// separation of criteria from subject, Evidence citation rules, the
// correction and replacement model, historical preservation, and
// derived-current-effect semantics are all inherited without redefinition,
// exactly as PEOS-007 enumerates them. Its wire form is byte-for-byte
// validation.Claim's -- no envelope, no added key, no discriminator -- so a
// quality Claim's document is readable by any PEOS-006 consumer.
//
// PEOS-007 adds one invariant: where the subject is a Requirement or a
// Requirement Artifact Revision, no criterion may name that same Requirement.
// "The criterion SHALL be a Quality Characteristic, Measure, Threshold, Profile
// rule, or other rule distinct from the Requirement subject itself; the
// Requirement SHALL NOT be its own criterion." The comparison is by Requirement
// core.ArtifactID and is cross-level, because "the same Requirement" is
// identity-level language. PEOS-007 adds no minimum criteria count, so zero
// criteria are accepted.
//
// The wrapper exists rather than a bare constructor because peos/validation is
// PEOS-006-correct in permitting that reflexive shape for every
// non-Satisfaction Claim Type, so validation.Claim.WithCriteria enforces
// PEOS-006's rules and not PEOS-007's. A construct-then-WithCriteria sequence
// could otherwise produce a represented, PEOS-007-invalid quality Claim with no
// error anywhere. Every path that can change a field the rule depends on --
// construction, adoption of a raw Claim, all ten modifiers, and JSON decoding --
// routes through one shared validator before any value is assigned.
//
// # Extract, modify, re-wrap
//
// MeasurementRecord.Record and Claim.ValidationClaim both return value copies.
// Modifying such a copy cannot mutate the wrapper it came from, but the
// modified raw value is no longer represented as a validated PEOS-007 value --
// nothing about a bare validation.ExecutionRecord asserts that it still cites a
// Characteristic and a Measure, and nothing about a bare validation.Claim
// asserts PEOS-007's Requirement rule. Re-entry requires
// NewMeasurementRecord or NewClaimFromValidationClaim, each of which re-applies
// every check. This is why neither wrapper exposes a delegating modifier for
// the composed value's criteria, subject, method, or outcome: such a modifier
// could produce a represented value that violates PEOS-007.
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
// peos/validation. peos/validation is required, not merely permitted:
// MeasurementRecord composes validation.ExecutionRecord and Claim composes
// validation.Claim, both by named field, because PEOS-007 specializes PEOS-006's
// mechanisms rather than redefining them. The direction matches the
// specification exactly -- PEOS-007 depends on PEOS-006, never the reverse --
// and peos/validation needs no change for PEOS-007 to specialize it.
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
// Packet I.1 implemented the Quality Profile side: the Artifact Type, Profile,
// ProfileRevision, ProfileApplicability, ProfileContent, the seven owned value
// structures, and the three quality-local vocabulary wrappers, plus the three
// additive core.CriterionRef arms.
//
// Packet I.2 added the two specializations of PEOS-006 mechanisms:
// MeasurementRecord and Claim, with their sentinels
// ErrInvalidMeasurementRecord and ErrInvalidQualityClaim, both of which I.1 had
// already declared as reserved so that I.2 needed no change to errors.go.
//
// No PEOS-007 packet has modified peos/validation, and none needs to.
// peos/validation deliberately accepts a reflexive Requirement on a
// non-Satisfaction Claim -- PEOS-006 carves that case out explicitly -- and its
// own test asserting so is correct and untouched. The stricter PEOS-007 rule
// lives here, which is the whole point of the wrapper.
//
// PEOS-007 is not yet closed: Packet I.3 is the consolidated audit and Packet
// I.4 the closure. Nothing in this package has an accepted audit verdict.
package quality
