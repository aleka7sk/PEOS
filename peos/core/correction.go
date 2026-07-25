package core

import (
	"encoding/json"
	"fmt"
)

// This file defines the narrow, reusable correction primitives PEOS-006's
// Claim Correction section requires (and which the same "new record
// references an earlier one" pattern applies to every other immutable
// record family that documents a correction path): CorrectionKind
// (vocabulary.go), RecordRef, and RecordCorrectionRef[T].
//
// None of PEOS-000-009's immutable record families gain a correction
// field here. Packet A only provides the shape a later packet MAY embed
// into a specific record type (for example, a future ValidationClaim's
// "replaces" field would have type RecordCorrectionRef[ValidationClaimRef]).
// This packet does not assert that Runtime Observation or Runtime
// Violation support a correction reference; PEOS-008 does not document
// one for those two record families (see the PEOS Reference Meta-Model
// Blueprint SS16), and adding a correction field to them here would
// assert something the specifications do not.
//
// The correction kinds below are exactly the three PEOS-006 names:
// correct, replace, invalidate. This package never uses "supersede" or
// "superseded" for record correction; PEOS-002's Artifact Supersession
// is a distinct mechanism that applies to Artifacts, not to immutable
// records, and the two are not interchangeable vocabulary.

// Known discriminator values for RecordRef.
const (
	RecordKindValidationClaim           = "validation_claim"
	RecordKindValidationExecutionRecord = "validation_execution_record"
	RecordKindRuntimeBindingRecord      = "runtime_binding_record"
	RecordKindRuntimeUnbindingRecord    = "runtime_unbinding_record"
	RecordKindRuntimeObservation        = "runtime_observation"
	RecordKindRuntimeViolation          = "runtime_violation"
	RecordKindTemplateApplicationRecord = "template_application_record"
	RecordKindImmutableRecord           = "immutable_record"
)

var knownRecordKinds = map[string]bool{
	RecordKindValidationClaim:           true,
	RecordKindValidationExecutionRecord: true,
	RecordKindRuntimeBindingRecord:      true,
	RecordKindRuntimeUnbindingRecord:    true,
	RecordKindRuntimeObservation:        true,
	RecordKindRuntimeViolation:          true,
	RecordKindTemplateApplicationRecord: true,
	RecordKindImmutableRecord:           true,
}

// OpaqueRecord carries the discriminator and identity of a RecordRef
// whose kind this packet does not give a dedicated typed payload to.
type OpaqueRecord struct {
	kind       string
	namespace  string
	identifier string
}

func (o OpaqueRecord) Kind() string       { return o.kind }
func (o OpaqueRecord) Namespace() string  { return o.namespace }
func (o OpaqueRecord) Identifier() string { return o.identifier }

// RecordRef is a tagged union over the identity of any immutable record
// family this packet names an identity type for (identity.go). It exists
// so a correction reference can be represented dynamically (its record
// family decided at runtime, e.g. by a storage or serialization layer)
// as well as generically via RecordCorrectionRef[T] below (its record
// family fixed at compile time by a specific PEOS SDK packet).
type RecordRef struct {
	kind  string
	known bool

	validationClaim           ValidationClaimID
	validationExecutionRecord ValidationExecutionRecordID
	runtimeBindingRecord      RuntimeBindingRecordID
	runtimeUnbindingRecord    RuntimeUnbindingRecordID
	runtimeObservation        RuntimeObservationID
	runtimeViolation          RuntimeViolationID
	templateApplicationRecord TemplateApplicationRecordID
	immutableRecord           ImmutableRecordID

	opaque OpaqueRecord
}

func (r RecordRef) Kind() string  { return r.kind }
func (r RecordRef) IsKnown() bool { return r.known }
func (r RecordRef) IsZero() bool  { return r.kind == "" }

func RecordRefFromValidationClaim(id ValidationClaimID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromValidationClaim: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindValidationClaim, known: true, validationClaim: id}, nil
}

func (r RecordRef) AsValidationClaim() (ValidationClaimID, bool) {
	if r.kind != RecordKindValidationClaim {
		return ValidationClaimID{}, false
	}
	return r.validationClaim, true
}

func RecordRefFromValidationExecutionRecord(id ValidationExecutionRecordID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromValidationExecutionRecord: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindValidationExecutionRecord, known: true, validationExecutionRecord: id}, nil
}

func (r RecordRef) AsValidationExecutionRecord() (ValidationExecutionRecordID, bool) {
	if r.kind != RecordKindValidationExecutionRecord {
		return ValidationExecutionRecordID{}, false
	}
	return r.validationExecutionRecord, true
}

func RecordRefFromRuntimeBindingRecord(id RuntimeBindingRecordID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromRuntimeBindingRecord: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindRuntimeBindingRecord, known: true, runtimeBindingRecord: id}, nil
}

func (r RecordRef) AsRuntimeBindingRecord() (RuntimeBindingRecordID, bool) {
	if r.kind != RecordKindRuntimeBindingRecord {
		return RuntimeBindingRecordID{}, false
	}
	return r.runtimeBindingRecord, true
}

func RecordRefFromRuntimeUnbindingRecord(id RuntimeUnbindingRecordID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromRuntimeUnbindingRecord: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindRuntimeUnbindingRecord, known: true, runtimeUnbindingRecord: id}, nil
}

func (r RecordRef) AsRuntimeUnbindingRecord() (RuntimeUnbindingRecordID, bool) {
	if r.kind != RecordKindRuntimeUnbindingRecord {
		return RuntimeUnbindingRecordID{}, false
	}
	return r.runtimeUnbindingRecord, true
}

func RecordRefFromRuntimeObservation(id RuntimeObservationID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromRuntimeObservation: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindRuntimeObservation, known: true, runtimeObservation: id}, nil
}

func (r RecordRef) AsRuntimeObservation() (RuntimeObservationID, bool) {
	if r.kind != RecordKindRuntimeObservation {
		return RuntimeObservationID{}, false
	}
	return r.runtimeObservation, true
}

func RecordRefFromRuntimeViolation(id RuntimeViolationID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromRuntimeViolation: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindRuntimeViolation, known: true, runtimeViolation: id}, nil
}

func (r RecordRef) AsRuntimeViolation() (RuntimeViolationID, bool) {
	if r.kind != RecordKindRuntimeViolation {
		return RuntimeViolationID{}, false
	}
	return r.runtimeViolation, true
}

func RecordRefFromTemplateApplicationRecord(id TemplateApplicationRecordID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromTemplateApplicationRecord: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindTemplateApplicationRecord, known: true, templateApplicationRecord: id}, nil
}

func (r RecordRef) AsTemplateApplicationRecord() (TemplateApplicationRecordID, bool) {
	if r.kind != RecordKindTemplateApplicationRecord {
		return TemplateApplicationRecordID{}, false
	}
	return r.templateApplicationRecord, true
}

func RecordRefFromImmutableRecord(id ImmutableRecordID) (RecordRef, error) {
	if id.IsZero() {
		return RecordRef{}, fmt.Errorf("core: RecordRefFromImmutableRecord: %w", ErrEmptyIdentity)
	}
	return RecordRef{kind: RecordKindImmutableRecord, known: true, immutableRecord: id}, nil
}

func (r RecordRef) AsImmutableRecord() (ImmutableRecordID, bool) {
	if r.kind != RecordKindImmutableRecord {
		return ImmutableRecordID{}, false
	}
	return r.immutableRecord, true
}

// NewOpaqueRecordRef constructs a forward-compatible RecordRef for a
// record kind this packet does not name an identity type for.
func NewOpaqueRecordRef(kind, namespace, identifier string) (RecordRef, error) {
	k, err := normalizeIdentityValue(kind)
	if err != nil {
		return RecordRef{}, fmt.Errorf("core: NewOpaqueRecordRef: %w", err)
	}
	if knownRecordKinds[k] {
		return RecordRef{}, fmt.Errorf("core: NewOpaqueRecordRef: %q is a known kind, use its typed constructor: %w", k, ErrInvalidReferenceDiscriminator)
	}
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return RecordRef{}, fmt.Errorf("core: NewOpaqueRecordRef: %w", err)
	}
	id, err := normalizeIdentityValue(identifier)
	if err != nil {
		return RecordRef{}, fmt.Errorf("core: NewOpaqueRecordRef: %w", err)
	}
	return RecordRef{
		kind:   k,
		known:  false,
		opaque: OpaqueRecord{kind: k, namespace: ns, identifier: id},
	}, nil
}

func (r RecordRef) AsOpaque() (OpaqueRecord, bool) {
	if r.known || r.kind == "" {
		return OpaqueRecord{}, false
	}
	return r.opaque, true
}

type recordRefEnvelope struct {
	Kind string          `json:"kind"`
	Ref  json.RawMessage `json:"ref"`
}

func (r RecordRef) MarshalJSON() ([]byte, error) {
	if r.kind == "" {
		return nil, fmt.Errorf("core: marshal RecordRef: %w", ErrInvalidReferenceDiscriminator)
	}
	var (
		refBytes []byte
		err      error
	)
	switch {
	case !r.known:
		refBytes, err = json.Marshal(opaqueSubjectPayloadJSON{Namespace: r.opaque.namespace, Identifier: r.opaque.identifier})
	case r.kind == RecordKindValidationClaim:
		refBytes, err = json.Marshal(r.validationClaim)
	case r.kind == RecordKindValidationExecutionRecord:
		refBytes, err = json.Marshal(r.validationExecutionRecord)
	case r.kind == RecordKindRuntimeBindingRecord:
		refBytes, err = json.Marshal(r.runtimeBindingRecord)
	case r.kind == RecordKindRuntimeUnbindingRecord:
		refBytes, err = json.Marshal(r.runtimeUnbindingRecord)
	case r.kind == RecordKindRuntimeObservation:
		refBytes, err = json.Marshal(r.runtimeObservation)
	case r.kind == RecordKindRuntimeViolation:
		refBytes, err = json.Marshal(r.runtimeViolation)
	case r.kind == RecordKindTemplateApplicationRecord:
		refBytes, err = json.Marshal(r.templateApplicationRecord)
	case r.kind == RecordKindImmutableRecord:
		refBytes, err = json.Marshal(r.immutableRecord)
	default:
		return nil, fmt.Errorf("core: marshal RecordRef: %w", ErrInvalidReferenceDiscriminator)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(recordRefEnvelope{Kind: r.kind, Ref: refBytes})
}

func (r *RecordRef) UnmarshalJSON(data []byte) error {
	var env recordRefEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("core: unmarshal RecordRef: %w", err)
	}
	if env.Kind == "" {
		return fmt.Errorf("core: unmarshal RecordRef: %w", ErrInvalidReferenceDiscriminator)
	}

	var (
		result RecordRef
		err    error
	)
	switch env.Kind {
	case RecordKindValidationClaim:
		var id ValidationClaimID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromValidationClaim(id)
		}
	case RecordKindValidationExecutionRecord:
		var id ValidationExecutionRecordID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromValidationExecutionRecord(id)
		}
	case RecordKindRuntimeBindingRecord:
		var id RuntimeBindingRecordID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromRuntimeBindingRecord(id)
		}
	case RecordKindRuntimeUnbindingRecord:
		var id RuntimeUnbindingRecordID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromRuntimeUnbindingRecord(id)
		}
	case RecordKindRuntimeObservation:
		var id RuntimeObservationID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromRuntimeObservation(id)
		}
	case RecordKindRuntimeViolation:
		var id RuntimeViolationID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromRuntimeViolation(id)
		}
	case RecordKindTemplateApplicationRecord:
		var id TemplateApplicationRecordID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromTemplateApplicationRecord(id)
		}
	case RecordKindImmutableRecord:
		var id ImmutableRecordID
		if err = json.Unmarshal(env.Ref, &id); err == nil {
			result, err = RecordRefFromImmutableRecord(id)
		}
	default:
		var payload opaqueSubjectPayloadJSON
		if err = json.Unmarshal(env.Ref, &payload); err == nil {
			result, err = NewOpaqueRecordRef(env.Kind, payload.Namespace, payload.Identifier)
		} else {
			err = fmt.Errorf("core: unmarshal RecordRef: unrecognized kind %q with non-opaque ref: %w", env.Kind, ErrInvalidPayload)
		}
	}
	if err != nil {
		return err
	}
	*r = result
	return nil
}

// zeroChecker is satisfied by every identity, reference, and vocabulary
// type in this package; it is the minimal constraint RecordCorrectionRef
// needs to validate its target without knowing the target's concrete
// type.
type zeroChecker interface {
	IsZero() bool
}

// RecordCorrectionRef pairs a CorrectionKind with a strongly-typed
// reference to the earlier record being corrected, replaced, or
// invalidated. T is fixed by the caller: a future packet embedding a
// correction reference into, for example, a Validation Claim would use
// RecordCorrectionRef[ValidationClaimRef]; a caller that needs the target
// record family decided at runtime instead can instantiate
// RecordCorrectionRef[RecordRef].
//
// This type is never embedded automatically into a record; see the
// package comment at the top of this file.
type RecordCorrectionRef[T zeroChecker] struct {
	kind   CorrectionKind
	target T
}

// NewRecordCorrectionRef validates kind and target and returns a
// RecordCorrectionRef.
func NewRecordCorrectionRef[T zeroChecker](kind CorrectionKind, target T) (RecordCorrectionRef[T], error) {
	if kind.IsZero() {
		return RecordCorrectionRef[T]{}, fmt.Errorf("core: NewRecordCorrectionRef: %w", ErrInvalidCorrectionReference)
	}
	if target.IsZero() {
		return RecordCorrectionRef[T]{}, fmt.Errorf("core: NewRecordCorrectionRef: %w", ErrInvalidCorrectionReference)
	}
	return RecordCorrectionRef[T]{kind: kind, target: target}, nil
}

func (r RecordCorrectionRef[T]) Kind() CorrectionKind { return r.kind }
func (r RecordCorrectionRef[T]) Target() T            { return r.target }
func (r RecordCorrectionRef[T]) IsZero() bool         { return r.kind.IsZero() }

type recordCorrectionRefJSON[T zeroChecker] struct {
	Kind   CorrectionKind `json:"kind"`
	Target T              `json:"target"`
}

func (r RecordCorrectionRef[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(recordCorrectionRefJSON[T]{Kind: r.kind, Target: r.target})
}

func (r *RecordCorrectionRef[T]) UnmarshalJSON(data []byte) error {
	var raw recordCorrectionRefJSON[T]
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RecordCorrectionRef: %w", err)
	}
	v, err := NewRecordCorrectionRef(raw.Kind, raw.Target)
	if err != nil {
		return err
	}
	*r = v
	return nil
}
