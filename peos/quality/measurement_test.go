package quality

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// --- helpers -----------------------------------------------------------------

func mustExecutionRecordID(t *testing.T, value string) core.ValidationExecutionRecordID {
	t.Helper()
	id, err := core.NewValidationExecutionRecordID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustTimestamp(t *testing.T) core.Timestamp {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func mustActor(t *testing.T) core.ActorRef {
	t.Helper()
	actor, err := core.NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return actor
}

func mustAdHocActivity(t *testing.T) validation.ActivityReference {
	t.Helper()
	a, err := validation.NewAdHocActivityReference("manual latency probe")
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustArtifactSubject(t *testing.T, artifactID, revisionID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := core.EngineeringSubjectRefFromArtifactRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	return subject
}

// mustProfileElementRef builds the (Profile Revision, local key) composite every
// Profile-owned criterion carries.
func mustProfileElementRef(t *testing.T, element string) core.QualityElementCriterionRef {
	t.Helper()
	profileRef, err := core.NewArtifactRevisionRef(mustArtifactID(t, "QP-1"), mustArtifactRevisionID(t, "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewQualityElementCriterionRef(profileRef, mustLocalKey(t, element))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustCharacteristicCriterion(t *testing.T, element string) core.CriterionRef {
	t.Helper()
	c, err := core.CriterionRefFromQualityCharacteristic(mustProfileElementRef(t, element))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustMeasureCriterion(t *testing.T, element string) core.CriterionRef {
	t.Helper()
	c, err := core.CriterionRefFromQualityMeasure(mustProfileElementRef(t, element))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustThresholdCriterion(t *testing.T, element string) core.CriterionRef {
	t.Helper()
	c, err := core.CriterionRefFromQualityThreshold(mustProfileElementRef(t, element))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustTargetCriterion(t *testing.T, element string) core.CriterionRef {
	t.Helper()
	c, err := core.CriterionRefFromQualityTarget(mustProfileElementRef(t, element))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustConstraintCriterion(t *testing.T, element string) core.CriterionRef {
	t.Helper()
	c, err := core.CriterionRefFromQualityConstraint(mustProfileElementRef(t, element))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustExternalRuleCriterion(t *testing.T) core.CriterionRef {
	t.Helper()
	rule, err := core.NewExternalRuleRef(mustVocabularyValue(t, "iso", "25010-1"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := core.CriterionRefFromExternalRule(rule)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// mustBareExecutionRecord builds a valid PEOS-006 Validation Execution Record
// with no criteria. It is deliberately minimal: only PEOS-006's own eight
// mandatory fields are supplied, so every test that adds criteria is exercising
// exactly the PEOS-007 rule and nothing else.
func mustBareExecutionRecord(t *testing.T) validation.ExecutionRecord {
	t.Helper()
	record, err := validation.NewExecutionRecord(
		mustExecutionRecordID(t, "EXEC-1"),
		mustAdHocActivity(t),
		mustArtifactSubject(t, "ART-1", "REV-1"),
		mustValidationMethod(t, "automated-test"),
		core.ExecutionOutcomeCompleted,
		mustTimestamp(t),
		mustActor(t),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

// mustExecutionRecordWithCriteria builds a valid record citing exactly the
// criteria given.
func mustExecutionRecordWithCriteria(t *testing.T, criteria ...core.CriterionRef) validation.ExecutionRecord {
	t.Helper()
	record, err := mustBareExecutionRecord(t).WithCriteria(criteria)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

// mustMeasurementRecord builds the canonical valid MeasurementRecord.
func mustMeasurementRecord(t *testing.T) MeasurementRecord {
	t.Helper()
	m, err := NewMeasurementRecord(
		mustExecutionRecordWithCriteria(t,
			mustCharacteristicCriterion(t, "latency"),
			mustMeasureCriterion(t, "latency-p99"),
		),
		"243",
		mustUnit(t, "millisecond"),
		mustScale(t, "ratio"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// --- construction ------------------------------------------------------------

func TestNewMeasurementRecord(t *testing.T) {
	m := mustMeasurementRecord(t)

	if m.ID() != mustExecutionRecordID(t, "EXEC-1") {
		t.Error("ID() does not delegate to the composed record")
	}
	if m.ObservedValue() != "243" {
		t.Errorf("ObservedValue() = %q", m.ObservedValue())
	}
	if !m.Unit().Equal(mustUnit(t, "millisecond")) {
		t.Error("Unit() mismatch")
	}
	if !m.Scale().Equal(mustScale(t, "ratio")) {
		t.Error("Scale() mismatch")
	}
	if !m.Extension().IsZero() {
		t.Error("Extension() non-zero before one is set")
	}
	if m.IsZero() {
		t.Error("IsZero() = true for a constructed record")
	}

	ref, err := m.Ref()
	if err != nil {
		t.Fatal(err)
	}
	recordRef, err := m.Record().Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref != recordRef {
		t.Error("Ref() does not delegate to the composed record")
	}
}

func TestNewMeasurementRecordTrimsObservedValue(t *testing.T) {
	m, err := NewMeasurementRecord(
		mustExecutionRecordWithCriteria(t,
			mustCharacteristicCriterion(t, "latency"),
			mustMeasureCriterion(t, "latency-p99"),
		),
		"  243.7  ",
		mustUnit(t, "millisecond"),
		mustScale(t, "ratio"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if m.ObservedValue() != "243.7" {
		t.Errorf("ObservedValue() = %q, want the trimmed value", m.ObservedValue())
	}
}

func TestNewMeasurementRecordRejectsInvalidMandatoryState(t *testing.T) {
	valid := mustExecutionRecordWithCriteria(t,
		mustCharacteristicCriterion(t, "latency"),
		mustMeasureCriterion(t, "latency-p99"),
	)
	unit := mustUnit(t, "millisecond")
	scale := mustScale(t, "ratio")

	cases := map[string]func() (MeasurementRecord, error){
		"zero record": func() (MeasurementRecord, error) {
			return NewMeasurementRecord(validation.ExecutionRecord{}, "1", unit, scale)
		},
		"empty observed value": func() (MeasurementRecord, error) {
			return NewMeasurementRecord(valid, "", unit, scale)
		},
		"blank observed value": func() (MeasurementRecord, error) {
			return NewMeasurementRecord(valid, "   \t\n ", unit, scale)
		},
		"zero unit": func() (MeasurementRecord, error) {
			return NewMeasurementRecord(valid, "1", Unit{}, scale)
		},
		"zero scale": func() (MeasurementRecord, error) {
			return NewMeasurementRecord(valid, "1", unit, Scale{})
		},
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := fn()
			if !errors.Is(err, ErrInvalidMeasurementRecord) {
				t.Errorf("error = %v, want %v", err, ErrInvalidMeasurementRecord)
			}
			if !got.IsZero() {
				t.Error("a failed constructor returned a non-zero value")
			}
		})
	}
}

// --- criteria invariants -----------------------------------------------------

// TestMeasurementRecordCriteriaInvariant covers PEOS-007's requirement that
// every Measurement Record identify "the exact Quality Characteristic and
// Quality Measure references applied" -- both, in any order, with additional
// criteria of other kinds permitted because the specification states no
// exclusivity and no maximum.
func TestMeasurementRecordCriteriaInvariant(t *testing.T) {
	characteristic := mustCharacteristicCriterion(t, "latency")
	measure := mustMeasureCriterion(t, "latency-p99")

	accepted := map[string][]core.CriterionRef{
		"characteristic then measure": {characteristic, measure},
		"measure then characteristic": {measure, characteristic},
		"plus a threshold":            {characteristic, measure, mustThresholdCriterion(t, "latency-max")},
		"plus a target":               {characteristic, measure, mustTargetCriterion(t, "latency-goal")},
		"plus a constraint":           {characteristic, measure, mustConstraintCriterion(t, "no-plaintext")},
		"plus an external rule":       {characteristic, measure, mustExternalRuleCriterion(t)},
		"plus all four":               {mustThresholdCriterion(t, "th"), characteristic, mustExternalRuleCriterion(t), measure, mustTargetCriterion(t, "tg"), mustConstraintCriterion(t, "co")},
		// PEOS-007 states no uniqueness rule for a record's criteria, so
		// duplicates are accepted rather than rejected -- enforcing a
		// uniqueness rule the specification does not state would be an
		// invented invariant. This matches validation.PlannedActivity, whose
		// criteria are likewise not deduplicated.
		"duplicate characteristics": {characteristic, characteristic, measure},
		"duplicate measures":        {characteristic, measure, measure},
	}
	for name, criteria := range accepted {
		t.Run("accepted: "+name, func(t *testing.T) {
			if _, err := NewMeasurementRecord(
				mustExecutionRecordWithCriteria(t, criteria...),
				"1", mustUnit(t, "ms"), mustScale(t, "ratio"),
			); err != nil {
				t.Errorf("rejected a valid criteria set: %v", err)
			}
		})
	}

	rejected := map[string][]core.CriterionRef{
		"characteristic without measure": {characteristic},
		"measure without characteristic": {measure},
		"neither":                        {mustThresholdCriterion(t, "th")},
		"only a target":                  {mustTargetCriterion(t, "tg")},
		"no criteria at all":             nil,
	}
	for name, criteria := range rejected {
		t.Run("rejected: "+name, func(t *testing.T) {
			record := mustBareExecutionRecord(t)
			if len(criteria) > 0 {
				var err error
				if record, err = record.WithCriteria(criteria); err != nil {
					t.Fatal(err)
				}
			}
			_, err := NewMeasurementRecord(record, "1", mustUnit(t, "ms"), mustScale(t, "ratio"))
			if !errors.Is(err, ErrInvalidMeasurementRecord) {
				t.Errorf("error = %v, want %v", err, ErrInvalidMeasurementRecord)
			}
		})
	}
}

// TestMeasurementRecordAcceptsAdHocActivityAndMinimalOptionalState records that
// PEOS-007 adds no planning requirement: a measurement need not descend from a
// Planned Validation Activity, and none of PEOS-006's conditional fields
// (environment, uncertainty, limitations, produced or relied-upon Evidence,
// started timestamp) is required for a Measurement Record.
func TestMeasurementRecordAcceptsAdHocActivityAndMinimalOptionalState(t *testing.T) {
	m := mustMeasurementRecord(t)
	if _, ok := m.Record().Activity().AsAdHoc(); !ok {
		t.Fatal("the canonical test record is not ad hoc; the case is not being exercised")
	}
	if _, ok := m.Record().StartedAt(); ok {
		t.Error("the canonical record unexpectedly carries a started timestamp")
	}
	if _, ok := m.Record().Environment(); ok {
		t.Error("environment unexpectedly set")
	}
	if _, ok := m.Record().Uncertainty(); ok {
		t.Error("uncertainty unexpectedly set")
	}
	if _, ok := m.Record().Limitations(); ok {
		t.Error("limitations unexpectedly set")
	}
	if m.Record().ProducedEvidence() != nil || m.Record().ReliedUponEvidence() != nil {
		t.Error("evidence unexpectedly set")
	}

	// A planned activity reference is equally acceptable.
	planRevision, err := core.NewValidationPlanRevisionRef(mustArtifactID(t, "VP-1"), mustArtifactRevisionID(t, "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	planned, err := validation.NewPlannedActivityReference(planRevision, mustLocalKey(t, "measure-latency"))
	if err != nil {
		t.Fatal(err)
	}
	record, err := validation.NewExecutionRecord(
		mustExecutionRecordID(t, "EXEC-2"), planned,
		mustArtifactSubject(t, "ART-1", "REV-1"),
		mustValidationMethod(t, "automated-test"), core.ExecutionOutcomeCompleted,
		mustTimestamp(t), mustActor(t), mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if record, err = record.WithCriteria([]core.CriterionRef{
		mustCharacteristicCriterion(t, "latency"),
		mustMeasureCriterion(t, "latency-p99"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := NewMeasurementRecord(record, "1", mustUnit(t, "ms"), mustScale(t, "ratio")); err != nil {
		t.Errorf("a planned-activity measurement record was rejected: %v", err)
	}
}

// --- immutability ------------------------------------------------------------

// TestMeasurementRecordExtractedRecordIsACopy is the structural guarantee behind
// the extract-modify-rewrap rule: Record() hands back a value, so no
// modification of it can reach back into the wrapper.
func TestMeasurementRecordExtractedRecordIsACopy(t *testing.T) {
	m := mustMeasurementRecord(t)

	extracted := m.Record()
	// Strip the criteria PEOS-007 requires. This is a legal operation on a
	// bare validation.ExecutionRecord -- PEOS-006 permits zero criteria.
	stripped, err := extracted.WithCriteria(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(stripped.Criteria()) != 0 {
		t.Fatal("WithCriteria(nil) did not strip the criteria")
	}

	// The wrapper is untouched.
	if len(m.Record().Criteria()) != 2 {
		t.Error("modifying the extracted record mutated the MeasurementRecord")
	}
	if _, err := json.Marshal(m); err != nil {
		t.Errorf("the wrapper became invalid after its extracted copy was modified: %v", err)
	}

	// And the stripped record can no longer be wrapped -- which is exactly why
	// no delegating criteria modifier is exposed.
	if _, err := NewMeasurementRecord(stripped, "1", mustUnit(t, "ms"), mustScale(t, "ratio")); !errors.Is(err, ErrInvalidMeasurementRecord) {
		t.Errorf("re-wrapping a stripped record error = %v, want %v", err, ErrInvalidMeasurementRecord)
	}

	// Mutating a slice returned through the extracted record cannot reach the
	// wrapper either -- validation.ExecutionRecord's own defensive copying is
	// inherited.
	criteria := m.Record().Criteria()
	criteria[0] = core.CriterionRef{}
	if m.Record().Criteria()[0].IsZero() {
		t.Error("Criteria() returned an aliased slice through the wrapper")
	}
}

func TestMeasurementRecordExtensionModifiers(t *testing.T) {
	m := mustMeasurementRecord(t)
	ext := mustExtension(t)

	withExt := m.WithExtension(ext)
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set the extension")
	}
	if !m.Extension().IsZero() {
		t.Error("WithExtension mutated the receiver")
	}

	cleared := withExt.WithoutExtension()
	if !cleared.Extension().IsZero() {
		t.Error("WithoutExtension did not clear the extension")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithoutExtension mutated the receiver")
	}

	// The wrapper's extension is distinct from the composed record's own.
	if !withExt.Record().Extension().IsZero() {
		t.Error("the wrapper's extension leaked onto the composed record")
	}
}

// --- JSON --------------------------------------------------------------------

func TestMeasurementRecordJSONRoundTrip(t *testing.T) {
	m := mustMeasurementRecord(t).WithExtension(mustExtension(t))

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysPresent(t, data, "record", "observed_value", "unit", "scale", "extension")
	assertKeysAbsent(t, data,
		"measurement_record_type", "type", "lifecycle", "state", "status",
		"current", "latest", "effective", "aggregate", "score", "quality_score",
		"result", "finding", "violation", "relation", "source", "target",
		"id", "subject", "criteria", "method", "outcome", "provenance")

	// The composed record is nested, not flattened: "record" holds exactly
	// validation.ExecutionRecord's own wire form.
	var envelope struct {
		Record json.RawMessage `json:"record"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	nested, err := json.Marshal(m.Record())
	if err != nil {
		t.Fatal(err)
	}
	if string(envelope.Record) != string(nested) {
		t.Errorf("nested record wire form differs:\n got %s\nwant %s", envelope.Record, nested)
	}

	var decoded MeasurementRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != m.ID() || decoded.ObservedValue() != m.ObservedValue() ||
		!decoded.Unit().Equal(m.Unit()) || !decoded.Scale().Equal(m.Scale()) {
		t.Error("a field was lost in the round trip")
	}
	if decoded.Extension().IsZero() {
		t.Error("extension lost in the round trip")
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
	}
}

func TestMeasurementRecordJSONRejection(t *testing.T) {
	valid := mustMeasurementRecord(t)
	validData, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}

	// A helper producing a document with one key replaced or removed.
	rewrite := func(t *testing.T, mutate func(map[string]json.RawMessage)) []byte {
		t.Helper()
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(validData, &obj); err != nil {
			t.Fatal(err)
		}
		mutate(obj)
		out, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	cases := map[string][]byte{
		"record missing": rewrite(t, func(o map[string]json.RawMessage) { delete(o, "record") }),
		"record null":    rewrite(t, func(o map[string]json.RawMessage) { o["record"] = json.RawMessage(`null`) }),
		"observed value missing": rewrite(t, func(o map[string]json.RawMessage) {
			delete(o, "observed_value")
		}),
		"observed value null": rewrite(t, func(o map[string]json.RawMessage) {
			o["observed_value"] = json.RawMessage(`null`)
		}),
		"observed value blank": rewrite(t, func(o map[string]json.RawMessage) {
			o["observed_value"] = json.RawMessage(`"   "`)
		}),
		"unit missing":  rewrite(t, func(o map[string]json.RawMessage) { delete(o, "unit") }),
		"unit null":     rewrite(t, func(o map[string]json.RawMessage) { o["unit"] = json.RawMessage(`null`) }),
		"scale missing": rewrite(t, func(o map[string]json.RawMessage) { delete(o, "scale") }),
		"scale null":    rewrite(t, func(o map[string]json.RawMessage) { o["scale"] = json.RawMessage(`null`) }),
		"null document": []byte(`null`),
		"empty object":  []byte(`{}`),
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			decoded := valid
			err := json.Unmarshal(doc, &decoded)
			if err == nil {
				t.Fatalf("accepted %s, want rejection", doc)
			}
			// The record-null case surfaces PEOS-006's own sentinel, which is
			// correct: the nested type owns that failure.
			if name == "record null" {
				if !errors.Is(err, validation.ErrInvalidExecutionRecord) &&
					!errors.Is(err, ErrInvalidMeasurementRecord) {
					t.Errorf("error = %v, want a validation or quality sentinel", err)
				}
			} else if !errors.Is(err, ErrInvalidMeasurementRecord) {
				t.Errorf("error = %v, want it to wrap %v", err, ErrInvalidMeasurementRecord)
			}
			if decoded.ID() != valid.ID() {
				t.Error("a failed decode overwrote a previously valid receiver")
			}
		})
	}

	// A nested record whose criteria do not satisfy PEOS-007 is rejected on
	// decode, proving UnmarshalJSON routes through the same shared path.
	stripped, err := mustBareExecutionRecord(t).WithCriteria([]core.CriterionRef{
		mustCharacteristicCriterion(t, "latency"),
	})
	if err != nil {
		t.Fatal(err)
	}
	strippedData, err := json.Marshal(stripped)
	if err != nil {
		t.Fatal(err)
	}
	doc := rewrite(t, func(o map[string]json.RawMessage) { o["record"] = strippedData })
	var decoded MeasurementRecord
	if err := json.Unmarshal(doc, &decoded); !errors.Is(err, ErrInvalidMeasurementRecord) {
		t.Errorf("a record missing its measure criterion decoded successfully: %v", err)
	}
}

func TestMeasurementRecordJSONExtensionAbsentEqualsNull(t *testing.T) {
	base := mustMeasurementRecord(t)
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysAbsent(t, data, "extension")

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	obj["extension"] = json.RawMessage(`null`)
	withNull, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MeasurementRecord
	if err := json.Unmarshal(withNull, &decoded); err != nil {
		t.Fatalf("an explicit null extension was rejected: %v", err)
	}
	if !decoded.Extension().IsZero() {
		t.Error("a null extension did not decode to the zero Extension")
	}
}

func TestMeasurementRecordJSONUnknownFieldIgnored(t *testing.T) {
	data, err := json.Marshal(mustMeasurementRecord(t))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	obj["unknown_future_key"] = json.RawMessage(`{"x":1}`)
	withExtra, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MeasurementRecord
	if err := json.Unmarshal(withExtra, &decoded); err != nil {
		t.Fatalf("an unknown ordinary field was rejected: %v", err)
	}
	if decoded.ObservedValue() != "243" {
		t.Error("an unknown ordinary field changed the decoded value")
	}
}

func TestMeasurementRecordZeroMarshalFails(t *testing.T) {
	if _, err := json.Marshal(MeasurementRecord{}); !errors.Is(err, ErrInvalidMeasurementRecord) {
		t.Errorf("zero-value marshal error = %v, want %v", err, ErrInvalidMeasurementRecord)
	}
	if (MeasurementRecord{}).IsZero() != true {
		t.Error("zero MeasurementRecord IsZero() = false")
	}
}

func TestMeasurementRecordNestedSentinelReachable(t *testing.T) {
	valid := mustMeasurementRecord(t)
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}

	// A nested record whose own subject reference is malformed must surface a
	// core sentinel, not a re-attributed quality one.
	obj["record"] = json.RawMessage(`{"id":"EXEC-9","activity":{"kind":"ad_hoc","ref":{"designation":"d"}},` +
		`"subject":{"kind":"artifact_revision","ref":{"artifact_id":"ART-1","revision_id":""}},` +
		`"method":"product-x:test","outcome":"peos:completed","completed_at":"2026-07-27T00:00:00Z",` +
		`"actor":{"namespace":"a","identifier":"b"},"provenance":{"actor":{"namespace":"a","identifier":"b"}}}`)
	doc, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MeasurementRecord
	err = json.Unmarshal(doc, &decoded)
	if err == nil {
		t.Fatal("a malformed nested subject reference was accepted")
	}
	if !errors.Is(err, core.ErrMissingRevisionID) && !errors.Is(err, core.ErrEmptyIdentity) &&
		!errors.Is(err, core.ErrInvalidPayload) {
		t.Errorf("error = %v, want a nested core sentinel to remain reachable", err)
	}
}
