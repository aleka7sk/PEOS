package quality

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// MeasurementRecord is a PEOS-007 Measurement Record: it "specializes the
// Validation Execution Record defined by PEOS-006", is "an immutable record",
// "is not an Artifact", and "has no revisions and no lifecycle."
//
// # Why a composing wrapper
//
// PEOS-007 requires a Measurement Record to identify "the observed value" and
// "the unit and scale". validation.ExecutionRecord has no field for any of the
// three, and no validator profile over it can add one. Putting them in
// core.Extension is exactly what relation.Relation's own documentation
// forbids -- extension data is for Product-specific content, not for
// PEOS-defined normative fields. Composition by named field is therefore the
// only available mechanism, and it is the same one requirement.Revision uses
// over core.ArtifactRevision and requirement.Derivation uses over
// relation.Relation.
//
// # What it does not introduce
//
// This is a specialization, not a parallel mechanism (non-conforming patterns
// "Parallel Quality Activity" and "Parallel Quality Evidence"). There is
// exactly one of each of the following, and every one of them is inherited
// from the composed validation.ExecutionRecord rather than restated:
//
//   - identity -- core.ValidationExecutionRecordID; this type declares no ID
//     field and mints no identity of its own;
//   - activity reference -- validation.ActivityReference, planned or ad hoc;
//   - subject, at its exact participant level;
//   - method -- core.ValidationMethod;
//   - execution outcome -- core.ExecutionOutcome;
//   - timestamps, actor, and provenance;
//   - criteria -- core.CriterionRef;
//   - Evidence -- produced and relied-upon core.EvidenceArtifactRevisionRef;
//   - correction model -- core.RecordCorrectionRef[core.ValidationExecutionRecordRef].
//     "Correction of a Measurement Record creates a new Measurement Record, in
//     accordance with PEOS-006's Validation Execution Record correction
//     rules"; there is no PEOS-007 correction mechanism, and correction is
//     never Artifact Supersession;
//   - wire form for all of the above, nested under one "record" key.
//
// MeasurementRecord adds exactly the three fields PEOS-007 mandates and
// PEOS-006 lacks, plus an optional quality-specific extension. It duplicates
// no field the composed record already owns.
//
// # Immutable raw history, never current state
//
// A Measurement Record is raw measurement history. "A Measurement Record SHALL
// NOT be mutated once recorded", and this type stores no score, no
// normalized value, no aggregate, no pass/fail, and no "current" or "latest"
// marker. PEOS-007 permits a quality score to appear "as a value recorded on a
// Measurement Record" -- that is what ObservedValue is -- but forbids storing
// one "as globally current mutable state", and requires any consumer needing a
// current value to "compute it, on demand, from the applicable, non-replaced,
// non-invalidated Measurement Records and Quality Claims". Normalization is
// likewise not applied here: a Normalization Rule is a description this
// package never executes.
//
// # Extract, modify, re-wrap
//
// Record returns a value copy. Modifying that copy -- for instance calling
// WithCriteria on it -- does not and cannot mutate the MeasurementRecord it
// came from, but the modified raw record is no longer represented as a
// validated MeasurementRecord: nothing about a bare
// validation.ExecutionRecord asserts that it still cites a Quality
// Characteristic and a Quality Measure. Re-entry requires
// NewMeasurementRecord, which re-applies every PEOS-007 check. This is why
// MeasurementRecord exposes no delegating modifier for the composed record's
// criteria, subject, method, or outcome: such a modifier could otherwise
// produce a represented MeasurementRecord that violates PEOS-007's
// SHALL-identify list.
type MeasurementRecord struct {
	record        validation.ExecutionRecord
	observedValue string
	unit          Unit
	scale         Scale
	extension     core.Extension
}

// validateMeasurementRecord is the single shared validation path for
// MeasurementRecord. NewMeasurementRecord, and therefore UnmarshalJSON and
// every wrapping operation, all route through it, so the PEOS-007 rules cannot
// drift between construction and decoding.
//
// observedValue is expected already trimmed; NewMeasurementRecord is the one
// place that trims, so there is exactly one trimming site as well as exactly
// one validation site.
func validateMeasurementRecord(
	caller string,
	record validation.ExecutionRecord,
	observedValue string,
	unit Unit,
	scale Scale,
) error {
	if record.IsZero() {
		return fmt.Errorf("quality: %s: %w: validation execution record must not be zero", caller, ErrInvalidMeasurementRecord)
	}
	if observedValue == "" {
		return fmt.Errorf("quality: %s: %w: observed value must not be empty", caller, ErrInvalidMeasurementRecord)
	}
	if unit.IsZero() {
		return fmt.Errorf("quality: %s: %w: unit must not be zero", caller, ErrInvalidMeasurementRecord)
	}
	if scale.IsZero() {
		return fmt.Errorf("quality: %s: %w: scale must not be zero", caller, ErrInvalidMeasurementRecord)
	}

	// PEOS-007: every Measurement Record SHALL identify "the exact Quality
	// Characteristic and Quality Measure references applied". Both are
	// required, in any order, and additional criteria of any other kind are
	// permitted -- PEOS-007 states no exclusivity and no maximum, and a
	// measurement taken against a Threshold, a Target, a Constraint, a
	// Requirement, or an external rule as well is a shape the specification
	// nowhere forbids.
	var hasCharacteristic, hasMeasure bool
	for _, criterion := range record.Criteria() {
		switch criterion.Kind() {
		case core.CriterionKindQualityCharacteristic:
			hasCharacteristic = true
		case core.CriterionKindQualityMeasure:
			hasMeasure = true
		}
	}
	if !hasCharacteristic {
		return fmt.Errorf("quality: %s: %w: the execution record must cite at least one %s criterion", caller, ErrInvalidMeasurementRecord, core.CriterionKindQualityCharacteristic)
	}
	if !hasMeasure {
		return fmt.Errorf("quality: %s: %w: the execution record must cite at least one %s criterion", caller, ErrInvalidMeasurementRecord, core.CriterionKindQualityMeasure)
	}

	return nil
}

// NewMeasurementRecord validates record, observedValue, unit, and scale and
// returns a MeasurementRecord with no extension data. Use WithExtension to add
// that.
//
// All four arguments are mandatory. record must be non-zero and must cite at
// least one quality_characteristic criterion and at least one quality_measure
// criterion, in any order. observedValue must be non-empty after trimming
// surrounding whitespace; the trimmed value is stored. unit and scale must
// both be non-zero.
//
// observedValue is an opaque string. PEOS-007 requires a Measurement Record to
// identify "the observed value" without defining any value type, and its
// Non-Goals disclaim "a universal quality metric catalog" and "a specific
// scoring formula or weighting scheme". A float would silently exclude
// ordinal, categorical, and interval-valued measures; a typed quantity would
// require the units framework PEOS-007 declines to define. Parsing the value,
// and interpreting it against its unit and scale, is Product-owned -- the same
// treatment core.Scope gives a scope expression.
//
// Nothing beyond the four arguments is required. PEOS-006 already governs the
// composed record's own mandatory fields, and the remaining items on PEOS-007's
// SHALL-identify list are satisfied by fields the composed record either
// mandates (subject and its exact participant level, method, timestamp,
// provenance) or makes conditional (environment "where applicable",
// uncertainty, limitations, Evidence). An ad hoc validation.ActivityReference
// is accepted, because PEOS-006 permits one and PEOS-007 adds no planning
// requirement; a Measurement Record does not have to descend from a Planned
// Validation Activity.
//
// A successful call always returns a fully valid MeasurementRecord: no
// mandatory value is supplied by a later modifier, so the record can never
// exist in a partially established state.
func NewMeasurementRecord(
	record validation.ExecutionRecord,
	observedValue string,
	unit Unit,
	scale Scale,
) (MeasurementRecord, error) {
	trimmed := strings.TrimSpace(observedValue)
	if err := validateMeasurementRecord("NewMeasurementRecord", record, trimmed, unit, scale); err != nil {
		return MeasurementRecord{}, err
	}
	return MeasurementRecord{
		record:        record,
		observedValue: trimmed,
		unit:          unit,
		scale:         scale,
	}, nil
}

// WithExtension returns a copy of m with its quality-specific extension data
// set. Passing the zero core.Extension is equivalent to declaring none, per
// core.Extension's own documented contract.
//
// This is the only modifier MeasurementRecord exposes, and it is the only one
// it can safely expose: extension data is Product-specific content that no
// PEOS-007 invariant constrains, so setting it cannot invalidate the record.
// The composed validation.ExecutionRecord carries its own separate extension;
// this one is the quality wrapper's, and neither overwrites the other.
func (m MeasurementRecord) WithExtension(extension core.Extension) MeasurementRecord {
	m.extension = extension
	return m
}

// WithoutExtension returns a copy of m with its extension data cleared.
func (m MeasurementRecord) WithoutExtension() MeasurementRecord {
	m.extension = core.Extension{}
	return m
}

// Record returns a value copy of m's underlying PEOS-006 Validation Execution
// Record.
//
// The returned value is a copy: modifying it -- including through its own
// WithCriteria, WithProducedEvidence, or any other modifier -- cannot mutate
// m. The result of such a modification is a bare validation.ExecutionRecord,
// no longer represented as a validated MeasurementRecord; pass it back through
// NewMeasurementRecord, together with an observed value, unit, and scale, to
// re-establish the PEOS-007 guarantees.
func (m MeasurementRecord) Record() validation.ExecutionRecord { return m.record }

// ID returns m's identity, which is the composed Validation Execution Record's
// own core.ValidationExecutionRecordID. A Measurement Record has no separate
// identity space.
func (m MeasurementRecord) ID() core.ValidationExecutionRecordID { return m.record.ID() }

// Ref returns a core.ValidationExecutionRecordRef citing m. This is the
// reference a Quality Claim's execution-record list carries, and the reference
// a later correcting Measurement Record targets.
func (m MeasurementRecord) Ref() (core.ValidationExecutionRecordRef, error) {
	return m.record.Ref()
}

// ObservedValue returns the value observed by the measurement, uninterpreted.
// It is mandatory and therefore never absent on a valid MeasurementRecord.
func (m MeasurementRecord) ObservedValue() string { return m.observedValue }

// Unit returns the unit the observed value is expressed in.
func (m MeasurementRecord) Unit() Unit { return m.unit }

// Scale returns the scale the observed value is expressed on.
func (m MeasurementRecord) Scale() Scale { return m.scale }

// Extension returns m's quality-specific extension data. The composed
// record's own extension is reached through Record().Extension().
func (m MeasurementRecord) Extension() core.Extension { return m.extension }

// IsZero reports whether m is the zero value.
func (m MeasurementRecord) IsZero() bool {
	return m.record.IsZero() && m.observedValue == "" && m.unit.IsZero() && m.scale.IsZero()
}

type measurementRecordJSON struct {
	Record        validation.ExecutionRecord `json:"record"`
	ObservedValue string                     `json:"observed_value"`
	Unit          Unit                       `json:"unit"`
	Scale         Scale                      `json:"scale"`
	Extension     *core.Extension            `json:"extension,omitempty"`
}

// MarshalJSON encodes m as {"record":{...},"observed_value":...,"unit":...,
// "scale":...}, plus "extension" when set.
//
// The composed record is nested under "record" rather than flattened, for the
// same reason core.ArtifactRevision documents for its own specializations: a
// nested encoding delegates the whole of PEOS-006's wire contract to
// validation.ExecutionRecord.MarshalJSON, so this type can neither drift from
// it nor collide with one of its keys.
//
// Deliberately absent: any "measurement_record_type" discriminator (there is
// no sibling union to discriminate), and any lifecycle, "state", "status",
// "current", "latest", "score", "result", "finding", "violation", "relation",
// "source", or "target" key. Their absence is the structural proof that a
// Measurement Record is immutable raw history, not derived state, not a
// lifecycle-bearing entity, and not a relationship.
func (m MeasurementRecord) MarshalJSON() ([]byte, error) {
	if err := validateMeasurementRecord("marshal MeasurementRecord", m.record, m.observedValue, m.unit, m.scale); err != nil {
		return nil, err
	}
	raw := measurementRecordJSON{
		Record:        m.record,
		ObservedValue: m.observedValue,
		Unit:          m.unit,
		Scale:         m.scale,
	}
	if !m.extension.IsZero() {
		raw.Extension = &m.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes m from its JSON form, routing through
// NewMeasurementRecord so that every PEOS-007 check is re-applied on decode. A
// decoded MeasurementRecord can never be constructor-impossible: in
// particular, a document whose nested record cites no Quality Characteristic
// or no Quality Measure criterion is rejected. The receiver is left untouched
// unless every check passes.
//
// Missing-versus-null behavior, stated exactly rather than assumed:
//
//   - record: a missing key leaves the field zero and is rejected with
//     ErrInvalidMeasurementRecord. An explicit null invokes
//     validation.ExecutionRecord's own UnmarshalJSON, which leaves its fields
//     zero and fails there with validation.ErrInvalidExecutionRecord, so that
//     sentinel is what surfaces. Both are rejected; the sentinel sets differ.
//   - observed_value: a missing key, an explicit null, and a whitespace-only
//     string all yield the empty string after trimming, all rejected with
//     ErrInvalidMeasurementRecord.
//   - unit, scale: a missing key leaves the wrapper zero. An explicit null
//     reaches the wrapper's own UnmarshalJSON, which leaves the wrapped
//     core.VocabularyValue zero. Both are rejected with
//     ErrInvalidMeasurementRecord.
//   - extension: null is equivalent to absent, per core.Extension's own
//     documented contract, and both yield the zero Extension.
//
// Unknown ordinary fields are ignored, per this repository's convention.
func (m *MeasurementRecord) UnmarshalJSON(data []byte) error {
	var raw measurementRecordJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("quality: unmarshal MeasurementRecord: %w: %w", ErrInvalidMeasurementRecord, err)
	}
	result, err := NewMeasurementRecord(raw.Record, raw.ObservedValue, raw.Unit, raw.Scale)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*m = result
	return nil
}
