package core

import (
	"encoding/json"
	"fmt"
)

// This file defines every participant reference type PEOS-000-009
// require at the identity or Artifact Revision level, plus the two
// tagged unions built on top of them: EngineeringSubjectRef (a Claim or
// Lifecycle Subject) and, via composition, LifecycleSubjectRef.
//
// Every concrete reference type below stores exactly the identity
// components that its participant level requires: an identity-level
// reference has only the owning identity; a Revision-level reference has
// both the owning identity and the exact Revision identity. These are
// distinct Go struct types, not one struct with an optional Revision
// field, so a value cannot exist in a state where its participant level
// and its stored components disagree.
//
// Each struct below uses a field name unique to that type (see the
// comment in identity.go on why), so that no two reference types with the
// same component shape are mutually convertible via an explicit Go type
// conversion.

// ArtifactRef identifies an Artifact at the identity level (PEOS-002).
type ArtifactRef struct{ artifactRefID ArtifactID }

// NewArtifactRef validates id and returns an ArtifactRef.
func NewArtifactRef(id ArtifactID) (ArtifactRef, error) {
	if id.IsZero() {
		return ArtifactRef{}, fmt.Errorf("core: NewArtifactRef: %w", ErrEmptyIdentity)
	}
	return ArtifactRef{artifactRefID: id}, nil
}

func (r ArtifactRef) ArtifactID() ArtifactID { return r.artifactRefID }
func (r ArtifactRef) IsZero() bool           { return r.artifactRefID.IsZero() }

type artifactRefJSON struct {
	ArtifactID ArtifactID `json:"artifact_id"`
}

func (r ArtifactRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(artifactRefJSON{ArtifactID: r.artifactRefID})
}

func (r *ArtifactRef) UnmarshalJSON(data []byte) error {
	var raw artifactRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ArtifactRef: %w", err)
	}
	v, err := NewArtifactRef(raw.ArtifactID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ArtifactRevisionRef identifies an exact Artifact Revision (PEOS-002):
// the owning Artifact plus the exact Revision.
type ArtifactRevisionRef struct {
	revisionRefArtifactID ArtifactID
	revisionRefRevisionID ArtifactRevisionID
}

// NewArtifactRevisionRef validates artifactID and revisionID and returns
// an ArtifactRevisionRef.
func NewArtifactRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (ArtifactRevisionRef, error) {
	if artifactID.IsZero() {
		return ArtifactRevisionRef{}, fmt.Errorf("core: NewArtifactRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return ArtifactRevisionRef{}, fmt.Errorf("core: NewArtifactRevisionRef: %w", ErrMissingRevisionID)
	}
	return ArtifactRevisionRef{revisionRefArtifactID: artifactID, revisionRefRevisionID: revisionID}, nil
}

func (r ArtifactRevisionRef) ArtifactID() ArtifactID         { return r.revisionRefArtifactID }
func (r ArtifactRevisionRef) RevisionID() ArtifactRevisionID { return r.revisionRefRevisionID }
func (r ArtifactRevisionRef) IsZero() bool {
	return r.revisionRefArtifactID.IsZero() && r.revisionRefRevisionID.IsZero()
}

type artifactRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r ArtifactRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(artifactRevisionRefJSON{ArtifactID: r.revisionRefArtifactID, RevisionID: r.revisionRefRevisionID})
}

func (r *ArtifactRevisionRef) UnmarshalJSON(data []byte) error {
	var raw artifactRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ArtifactRevisionRef: %w", err)
	}
	v, err := NewArtifactRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RequirementRef identifies a Requirement at the identity level
// (PEOS-005). Requirement identity is an Artifact identity; this type is
// nonetheless structurally distinct from ArtifactRef so the two are never
// confused at a call site.
type RequirementRef struct{ requirementRefID ArtifactID }

func NewRequirementRef(id ArtifactID) (RequirementRef, error) {
	if id.IsZero() {
		return RequirementRef{}, fmt.Errorf("core: NewRequirementRef: %w", ErrEmptyIdentity)
	}
	return RequirementRef{requirementRefID: id}, nil
}

func (r RequirementRef) ArtifactID() ArtifactID { return r.requirementRefID }
func (r RequirementRef) IsZero() bool           { return r.requirementRefID.IsZero() }

type requirementRefJSON struct {
	ArtifactID ArtifactID `json:"artifact_id"`
}

func (r RequirementRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(requirementRefJSON{ArtifactID: r.requirementRefID})
}

func (r *RequirementRef) UnmarshalJSON(data []byte) error {
	var raw requirementRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RequirementRef: %w", err)
	}
	v, err := NewRequirementRef(raw.ArtifactID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RequirementArtifactRevisionRef identifies an exact Requirement Artifact
// Revision (PEOS-005), required whenever wording, acceptance criteria, or
// applicability being referenced is revision-specific.
type RequirementArtifactRevisionRef struct {
	requirementRevisionArtifactID ArtifactID
	requirementRevisionRevisionID ArtifactRevisionID
}

func NewRequirementArtifactRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (RequirementArtifactRevisionRef, error) {
	if artifactID.IsZero() {
		return RequirementArtifactRevisionRef{}, fmt.Errorf("core: NewRequirementArtifactRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return RequirementArtifactRevisionRef{}, fmt.Errorf("core: NewRequirementArtifactRevisionRef: %w", ErrMissingRevisionID)
	}
	return RequirementArtifactRevisionRef{requirementRevisionArtifactID: artifactID, requirementRevisionRevisionID: revisionID}, nil
}

func (r RequirementArtifactRevisionRef) ArtifactID() ArtifactID {
	return r.requirementRevisionArtifactID
}
func (r RequirementArtifactRevisionRef) RevisionID() ArtifactRevisionID {
	return r.requirementRevisionRevisionID
}
func (r RequirementArtifactRevisionRef) IsZero() bool {
	return r.requirementRevisionArtifactID.IsZero() && r.requirementRevisionRevisionID.IsZero()
}

type requirementArtifactRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r RequirementArtifactRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(requirementArtifactRevisionRefJSON{
		ArtifactID: r.requirementRevisionArtifactID,
		RevisionID: r.requirementRevisionRevisionID,
	})
}

func (r *RequirementArtifactRevisionRef) UnmarshalJSON(data []byte) error {
	var raw requirementArtifactRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RequirementArtifactRevisionRef: %w", err)
	}
	v, err := NewRequirementArtifactRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// DecisionRef identifies a Decision (PEOS-004) by its own Decision
// identity, which is distinct from the identity of any Decision Record
// Artifact that documents it.
type DecisionRef struct{ decisionRefID DecisionID }

func NewDecisionRef(id DecisionID) (DecisionRef, error) {
	if id.IsZero() {
		return DecisionRef{}, fmt.Errorf("core: NewDecisionRef: %w", ErrEmptyIdentity)
	}
	return DecisionRef{decisionRefID: id}, nil
}

func (r DecisionRef) DecisionID() DecisionID { return r.decisionRefID }
func (r DecisionRef) IsZero() bool           { return r.decisionRefID.IsZero() }

type decisionRefJSON struct {
	DecisionID DecisionID `json:"decision_id"`
}

func (r DecisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(decisionRefJSON{DecisionID: r.decisionRefID})
}

func (r *DecisionRef) UnmarshalJSON(data []byte) error {
	var raw decisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal DecisionRef: %w", err)
	}
	v, err := NewDecisionRef(raw.DecisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// DecisionOutcomeRef identifies a Decision Outcome (PEOS-004) via the
// identity of the Decision that establishes it; a Decision Outcome has no
// identity independent of its owning Decision.
type DecisionOutcomeRef struct{ decisionOutcomeRefID DecisionID }

func NewDecisionOutcomeRef(id DecisionID) (DecisionOutcomeRef, error) {
	if id.IsZero() {
		return DecisionOutcomeRef{}, fmt.Errorf("core: NewDecisionOutcomeRef: %w", ErrEmptyIdentity)
	}
	return DecisionOutcomeRef{decisionOutcomeRefID: id}, nil
}

func (r DecisionOutcomeRef) DecisionID() DecisionID { return r.decisionOutcomeRefID }
func (r DecisionOutcomeRef) IsZero() bool           { return r.decisionOutcomeRefID.IsZero() }

type decisionOutcomeRefJSON struct {
	DecisionID DecisionID `json:"decision_id"`
}

func (r DecisionOutcomeRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(decisionOutcomeRefJSON{DecisionID: r.decisionOutcomeRefID})
}

func (r *DecisionOutcomeRef) UnmarshalJSON(data []byte) error {
	var raw decisionOutcomeRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal DecisionOutcomeRef: %w", err)
	}
	v, err := NewDecisionOutcomeRef(raw.DecisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// EngineeringCommitmentRef identifies an Engineering Commitment
// (PEOS-004) via the identity of its owning Decision; PEOS-004
// deliberately does not reify Engineering Commitment as an independent
// top-level entity.
type EngineeringCommitmentRef struct{ engineeringCommitmentRefID DecisionID }

func NewEngineeringCommitmentRef(id DecisionID) (EngineeringCommitmentRef, error) {
	if id.IsZero() {
		return EngineeringCommitmentRef{}, fmt.Errorf("core: NewEngineeringCommitmentRef: %w", ErrEmptyIdentity)
	}
	return EngineeringCommitmentRef{engineeringCommitmentRefID: id}, nil
}

func (r EngineeringCommitmentRef) DecisionID() DecisionID { return r.engineeringCommitmentRefID }
func (r EngineeringCommitmentRef) IsZero() bool           { return r.engineeringCommitmentRefID.IsZero() }

type engineeringCommitmentRefJSON struct {
	DecisionID DecisionID `json:"decision_id"`
}

func (r EngineeringCommitmentRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(engineeringCommitmentRefJSON{DecisionID: r.engineeringCommitmentRefID})
}

func (r *EngineeringCommitmentRef) UnmarshalJSON(data []byte) error {
	var raw engineeringCommitmentRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal EngineeringCommitmentRef: %w", err)
	}
	v, err := NewEngineeringCommitmentRef(raw.DecisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeSubjectRef is an opaque, Product-defined reference to a runtime
// subject (PEOS-008). PEOS-008 does not require a runtime subject to be
// an Artifact, so this type never carries an ArtifactID; it is a plain
// namespaced identifier.
type RuntimeSubjectRef struct {
	runtimeSubjectNamespace  string
	runtimeSubjectIdentifier string
}

func NewRuntimeSubjectRef(namespace, identifier string) (RuntimeSubjectRef, error) {
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return RuntimeSubjectRef{}, fmt.Errorf("core: NewRuntimeSubjectRef: %w", err)
	}
	id, err := normalizeIdentityValue(identifier)
	if err != nil {
		return RuntimeSubjectRef{}, fmt.Errorf("core: NewRuntimeSubjectRef: %w", err)
	}
	return RuntimeSubjectRef{runtimeSubjectNamespace: ns, runtimeSubjectIdentifier: id}, nil
}

func (r RuntimeSubjectRef) Namespace() string  { return r.runtimeSubjectNamespace }
func (r RuntimeSubjectRef) Identifier() string { return r.runtimeSubjectIdentifier }
func (r RuntimeSubjectRef) IsZero() bool {
	return r.runtimeSubjectNamespace == "" && r.runtimeSubjectIdentifier == ""
}

type runtimeSubjectRefJSON struct {
	Namespace  string `json:"namespace"`
	Identifier string `json:"identifier"`
}

func (r RuntimeSubjectRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeSubjectRefJSON{Namespace: r.runtimeSubjectNamespace, Identifier: r.runtimeSubjectIdentifier})
}

func (r *RuntimeSubjectRef) UnmarshalJSON(data []byte) error {
	var raw runtimeSubjectRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeSubjectRef: %w", err)
	}
	v, err := NewRuntimeSubjectRef(raw.Namespace, raw.Identifier)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeContractRef identifies a Runtime Contract at the identity level
// (PEOS-008).
type RuntimeContractRef struct{ runtimeContractRefID ArtifactID }

func NewRuntimeContractRef(id ArtifactID) (RuntimeContractRef, error) {
	if id.IsZero() {
		return RuntimeContractRef{}, fmt.Errorf("core: NewRuntimeContractRef: %w", ErrEmptyIdentity)
	}
	return RuntimeContractRef{runtimeContractRefID: id}, nil
}

func (r RuntimeContractRef) ArtifactID() ArtifactID { return r.runtimeContractRefID }
func (r RuntimeContractRef) IsZero() bool           { return r.runtimeContractRefID.IsZero() }

type runtimeContractRefJSON struct {
	ArtifactID ArtifactID `json:"artifact_id"`
}

func (r RuntimeContractRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeContractRefJSON{ArtifactID: r.runtimeContractRefID})
}

func (r *RuntimeContractRef) UnmarshalJSON(data []byte) error {
	var raw runtimeContractRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeContractRef: %w", err)
	}
	v, err := NewRuntimeContractRef(raw.ArtifactID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeContractRevisionRef identifies an exact Runtime Contract
// Revision (PEOS-008). A Runtime Binding Record is required to use this
// type, never the bare RuntimeContractRef.
type RuntimeContractRevisionRef struct {
	runtimeContractRevisionArtifactID ArtifactID
	runtimeContractRevisionRevisionID ArtifactRevisionID
}

func NewRuntimeContractRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (RuntimeContractRevisionRef, error) {
	if artifactID.IsZero() {
		return RuntimeContractRevisionRef{}, fmt.Errorf("core: NewRuntimeContractRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return RuntimeContractRevisionRef{}, fmt.Errorf("core: NewRuntimeContractRevisionRef: %w", ErrMissingRevisionID)
	}
	return RuntimeContractRevisionRef{runtimeContractRevisionArtifactID: artifactID, runtimeContractRevisionRevisionID: revisionID}, nil
}

func (r RuntimeContractRevisionRef) ArtifactID() ArtifactID {
	return r.runtimeContractRevisionArtifactID
}
func (r RuntimeContractRevisionRef) RevisionID() ArtifactRevisionID {
	return r.runtimeContractRevisionRevisionID
}
func (r RuntimeContractRevisionRef) IsZero() bool {
	return r.runtimeContractRevisionArtifactID.IsZero() && r.runtimeContractRevisionRevisionID.IsZero()
}

type runtimeContractRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r RuntimeContractRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeContractRevisionRefJSON{
		ArtifactID: r.runtimeContractRevisionArtifactID,
		RevisionID: r.runtimeContractRevisionRevisionID,
	})
}

func (r *RuntimeContractRevisionRef) UnmarshalJSON(data []byte) error {
	var raw runtimeContractRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeContractRevisionRef: %w", err)
	}
	v, err := NewRuntimeContractRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeBindingRecordRef identifies an exact Runtime Binding Record
// (PEOS-008). It mirrors ValidationClaimRef and
// ValidationExecutionRecordRef: a dedicated, type-safe reference for a
// record family that also has a generic RecordRef arm
// (RecordKindRuntimeBindingRecord), used wherever the target record family
// is fixed at compile time rather than decided at runtime -- for example, a
// Runtime Unbinding Record's mandatory "exactly one Binding Record"
// reference, or a Runtime Binding Record's own optional correction
// reference (core.RecordCorrectionRef[RuntimeBindingRecordRef]).
type RuntimeBindingRecordRef struct{ runtimeBindingRecordRefID RuntimeBindingRecordID }

// NewRuntimeBindingRecordRef validates id and returns a
// RuntimeBindingRecordRef.
func NewRuntimeBindingRecordRef(id RuntimeBindingRecordID) (RuntimeBindingRecordRef, error) {
	if id.IsZero() {
		return RuntimeBindingRecordRef{}, fmt.Errorf("core: NewRuntimeBindingRecordRef: %w", ErrEmptyIdentity)
	}
	return RuntimeBindingRecordRef{runtimeBindingRecordRefID: id}, nil
}

func (r RuntimeBindingRecordRef) RecordID() RuntimeBindingRecordID {
	return r.runtimeBindingRecordRefID
}
func (r RuntimeBindingRecordRef) IsZero() bool { return r.runtimeBindingRecordRefID.IsZero() }

type runtimeBindingRecordRefJSON struct {
	RecordID RuntimeBindingRecordID `json:"record_id"`
}

func (r RuntimeBindingRecordRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeBindingRecordRefJSON{RecordID: r.runtimeBindingRecordRefID})
}

func (r *RuntimeBindingRecordRef) UnmarshalJSON(data []byte) error {
	var raw runtimeBindingRecordRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeBindingRecordRef: %w", err)
	}
	v, err := NewRuntimeBindingRecordRef(raw.RecordID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeUnbindingRecordRef identifies an exact Runtime Unbinding Record
// (PEOS-008). Same shape and purpose as RuntimeBindingRecordRef, used
// wherever a Runtime Unbinding Record is referenced at compile-time-fixed
// type -- for example, a Runtime Unbinding Record's own optional correction
// reference (core.RecordCorrectionRef[RuntimeUnbindingRecordRef]).
type RuntimeUnbindingRecordRef struct {
	runtimeUnbindingRecordRefID RuntimeUnbindingRecordID
}

// NewRuntimeUnbindingRecordRef validates id and returns a
// RuntimeUnbindingRecordRef.
func NewRuntimeUnbindingRecordRef(id RuntimeUnbindingRecordID) (RuntimeUnbindingRecordRef, error) {
	if id.IsZero() {
		return RuntimeUnbindingRecordRef{}, fmt.Errorf("core: NewRuntimeUnbindingRecordRef: %w", ErrEmptyIdentity)
	}
	return RuntimeUnbindingRecordRef{runtimeUnbindingRecordRefID: id}, nil
}

func (r RuntimeUnbindingRecordRef) RecordID() RuntimeUnbindingRecordID {
	return r.runtimeUnbindingRecordRefID
}
func (r RuntimeUnbindingRecordRef) IsZero() bool { return r.runtimeUnbindingRecordRefID.IsZero() }

type runtimeUnbindingRecordRefJSON struct {
	RecordID RuntimeUnbindingRecordID `json:"record_id"`
}

func (r RuntimeUnbindingRecordRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeUnbindingRecordRefJSON{RecordID: r.runtimeUnbindingRecordRefID})
}

func (r *RuntimeUnbindingRecordRef) UnmarshalJSON(data []byte) error {
	var raw runtimeUnbindingRecordRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeUnbindingRecordRef: %w", err)
	}
	v, err := NewRuntimeUnbindingRecordRef(raw.RecordID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// RuntimeObservationRef identifies an exact Runtime Observation
// (PEOS-008). Same shape and purpose as RuntimeBindingRecordRef, used
// wherever a Runtime Observation is referenced at compile-time-fixed type
// -- for example, a Runtime Violation's exact reference to the Observation
// that triggered it. PEOS-008 documents no correction reference for Runtime
// Observation (see correction.go), so unlike RuntimeBindingRecordRef and
// RuntimeUnbindingRecordRef, this type is never used as a
// RecordCorrectionRef type parameter.
type RuntimeObservationRef struct{ runtimeObservationRefID RuntimeObservationID }

// NewRuntimeObservationRef validates id and returns a RuntimeObservationRef.
func NewRuntimeObservationRef(id RuntimeObservationID) (RuntimeObservationRef, error) {
	if id.IsZero() {
		return RuntimeObservationRef{}, fmt.Errorf("core: NewRuntimeObservationRef: %w", ErrEmptyIdentity)
	}
	return RuntimeObservationRef{runtimeObservationRefID: id}, nil
}

func (r RuntimeObservationRef) RecordID() RuntimeObservationID { return r.runtimeObservationRefID }
func (r RuntimeObservationRef) IsZero() bool                   { return r.runtimeObservationRefID.IsZero() }

type runtimeObservationRefJSON struct {
	RecordID RuntimeObservationID `json:"record_id"`
}

func (r RuntimeObservationRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(runtimeObservationRefJSON{RecordID: r.runtimeObservationRefID})
}

func (r *RuntimeObservationRef) UnmarshalJSON(data []byte) error {
	var raw runtimeObservationRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeObservationRef: %w", err)
	}
	v, err := NewRuntimeObservationRef(raw.RecordID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// TemplateRef identifies a Template at the identity level (PEOS-009).
type TemplateRef struct{ templateRefID ArtifactID }

func NewTemplateRef(id ArtifactID) (TemplateRef, error) {
	if id.IsZero() {
		return TemplateRef{}, fmt.Errorf("core: NewTemplateRef: %w", ErrEmptyIdentity)
	}
	return TemplateRef{templateRefID: id}, nil
}

func (r TemplateRef) ArtifactID() ArtifactID { return r.templateRefID }
func (r TemplateRef) IsZero() bool           { return r.templateRefID.IsZero() }

type templateRefJSON struct {
	ArtifactID ArtifactID `json:"artifact_id"`
}

func (r TemplateRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(templateRefJSON{ArtifactID: r.templateRefID})
}

func (r *TemplateRef) UnmarshalJSON(data []byte) error {
	var raw templateRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal TemplateRef: %w", err)
	}
	v, err := NewTemplateRef(raw.ArtifactID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// TemplateArtifactRevisionRef identifies an exact Template Artifact
// Revision (PEOS-009).
type TemplateArtifactRevisionRef struct {
	templateRevisionArtifactID ArtifactID
	templateRevisionRevisionID ArtifactRevisionID
}

func NewTemplateArtifactRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (TemplateArtifactRevisionRef, error) {
	if artifactID.IsZero() {
		return TemplateArtifactRevisionRef{}, fmt.Errorf("core: NewTemplateArtifactRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return TemplateArtifactRevisionRef{}, fmt.Errorf("core: NewTemplateArtifactRevisionRef: %w", ErrMissingRevisionID)
	}
	return TemplateArtifactRevisionRef{templateRevisionArtifactID: artifactID, templateRevisionRevisionID: revisionID}, nil
}

func (r TemplateArtifactRevisionRef) ArtifactID() ArtifactID { return r.templateRevisionArtifactID }
func (r TemplateArtifactRevisionRef) RevisionID() ArtifactRevisionID {
	return r.templateRevisionRevisionID
}
func (r TemplateArtifactRevisionRef) IsZero() bool {
	return r.templateRevisionArtifactID.IsZero() && r.templateRevisionRevisionID.IsZero()
}

type templateArtifactRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r TemplateArtifactRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(templateArtifactRevisionRefJSON{
		ArtifactID: r.templateRevisionArtifactID,
		RevisionID: r.templateRevisionRevisionID,
	})
}

func (r *TemplateArtifactRevisionRef) UnmarshalJSON(data []byte) error {
	var raw templateArtifactRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal TemplateArtifactRevisionRef: %w", err)
	}
	v, err := NewTemplateArtifactRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// GeneratedArtifactRef identifies a Template-generated Artifact at the
// identity level (PEOS-009). A Generated Artifact has its own identity,
// independent of the Template that produced it.
type GeneratedArtifactRef struct{ generatedArtifactRefID ArtifactID }

func NewGeneratedArtifactRef(id ArtifactID) (GeneratedArtifactRef, error) {
	if id.IsZero() {
		return GeneratedArtifactRef{}, fmt.Errorf("core: NewGeneratedArtifactRef: %w", ErrEmptyIdentity)
	}
	return GeneratedArtifactRef{generatedArtifactRefID: id}, nil
}

func (r GeneratedArtifactRef) ArtifactID() ArtifactID { return r.generatedArtifactRefID }
func (r GeneratedArtifactRef) IsZero() bool           { return r.generatedArtifactRefID.IsZero() }

type generatedArtifactRefJSON struct {
	ArtifactID ArtifactID `json:"artifact_id"`
}

func (r GeneratedArtifactRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(generatedArtifactRefJSON{ArtifactID: r.generatedArtifactRefID})
}

func (r *GeneratedArtifactRef) UnmarshalJSON(data []byte) error {
	var raw generatedArtifactRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal GeneratedArtifactRef: %w", err)
	}
	v, err := NewGeneratedArtifactRef(raw.ArtifactID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// GeneratedArtifactRevisionRef identifies an exact Revision of a
// Template-generated Artifact (PEOS-009).
type GeneratedArtifactRevisionRef struct {
	generatedArtifactRevisionArtifactID ArtifactID
	generatedArtifactRevisionRevisionID ArtifactRevisionID
}

func NewGeneratedArtifactRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (GeneratedArtifactRevisionRef, error) {
	if artifactID.IsZero() {
		return GeneratedArtifactRevisionRef{}, fmt.Errorf("core: NewGeneratedArtifactRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return GeneratedArtifactRevisionRef{}, fmt.Errorf("core: NewGeneratedArtifactRevisionRef: %w", ErrMissingRevisionID)
	}
	return GeneratedArtifactRevisionRef{
		generatedArtifactRevisionArtifactID: artifactID,
		generatedArtifactRevisionRevisionID: revisionID,
	}, nil
}

func (r GeneratedArtifactRevisionRef) ArtifactID() ArtifactID {
	return r.generatedArtifactRevisionArtifactID
}
func (r GeneratedArtifactRevisionRef) RevisionID() ArtifactRevisionID {
	return r.generatedArtifactRevisionRevisionID
}
func (r GeneratedArtifactRevisionRef) IsZero() bool {
	return r.generatedArtifactRevisionArtifactID.IsZero() && r.generatedArtifactRevisionRevisionID.IsZero()
}

type generatedArtifactRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r GeneratedArtifactRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(generatedArtifactRevisionRefJSON{
		ArtifactID: r.generatedArtifactRevisionArtifactID,
		RevisionID: r.generatedArtifactRevisionRevisionID,
	})
}

func (r *GeneratedArtifactRevisionRef) UnmarshalJSON(data []byte) error {
	var raw generatedArtifactRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal GeneratedArtifactRevisionRef: %w", err)
	}
	v, err := NewGeneratedArtifactRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ValidationPlanRef identifies a Validation Plan at the identity level
// (PEOS-006).
type ValidationPlanRef struct{ validationPlanRefID ArtifactID }

func NewValidationPlanRef(id ArtifactID) (ValidationPlanRef, error) {
	if id.IsZero() {
		return ValidationPlanRef{}, fmt.Errorf("core: NewValidationPlanRef: %w", ErrEmptyIdentity)
	}
	return ValidationPlanRef{validationPlanRefID: id}, nil
}

func (r ValidationPlanRef) ArtifactID() ArtifactID { return r.validationPlanRefID }
func (r ValidationPlanRef) IsZero() bool           { return r.validationPlanRefID.IsZero() }

type validationPlanRefJSON struct {
	ArtifactID ArtifactID `json:"artifact_id"`
}

func (r ValidationPlanRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(validationPlanRefJSON{ArtifactID: r.validationPlanRefID})
}

func (r *ValidationPlanRef) UnmarshalJSON(data []byte) error {
	var raw validationPlanRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ValidationPlanRef: %w", err)
	}
	v, err := NewValidationPlanRef(raw.ArtifactID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ValidationPlanRevisionRef identifies an exact Validation Plan Revision
// (PEOS-006). A Planned Validation Activity's plan-local key is only
// meaningful together with a reference of this type.
type ValidationPlanRevisionRef struct {
	validationPlanRevisionArtifactID ArtifactID
	validationPlanRevisionRevisionID ArtifactRevisionID
}

func NewValidationPlanRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (ValidationPlanRevisionRef, error) {
	if artifactID.IsZero() {
		return ValidationPlanRevisionRef{}, fmt.Errorf("core: NewValidationPlanRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return ValidationPlanRevisionRef{}, fmt.Errorf("core: NewValidationPlanRevisionRef: %w", ErrMissingRevisionID)
	}
	return ValidationPlanRevisionRef{validationPlanRevisionArtifactID: artifactID, validationPlanRevisionRevisionID: revisionID}, nil
}

func (r ValidationPlanRevisionRef) ArtifactID() ArtifactID { return r.validationPlanRevisionArtifactID }
func (r ValidationPlanRevisionRef) RevisionID() ArtifactRevisionID {
	return r.validationPlanRevisionRevisionID
}
func (r ValidationPlanRevisionRef) IsZero() bool {
	return r.validationPlanRevisionArtifactID.IsZero() && r.validationPlanRevisionRevisionID.IsZero()
}

type validationPlanRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r ValidationPlanRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(validationPlanRevisionRefJSON{
		ArtifactID: r.validationPlanRevisionArtifactID,
		RevisionID: r.validationPlanRevisionRevisionID,
	})
}

func (r *ValidationPlanRevisionRef) UnmarshalJSON(data []byte) error {
	var raw validationPlanRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ValidationPlanRevisionRef: %w", err)
	}
	v, err := NewValidationPlanRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// EvidenceArtifactRevisionRef cites an exact Artifact Revision serving
// the Evidence role (PEOS-002/006). Evidence citation is always at the
// exact Revision level; there is no identity-level Evidence reference.
type EvidenceArtifactRevisionRef struct {
	evidenceArtifactID ArtifactID
	evidenceRevisionID ArtifactRevisionID
}

func NewEvidenceArtifactRevisionRef(artifactID ArtifactID, revisionID ArtifactRevisionID) (EvidenceArtifactRevisionRef, error) {
	if artifactID.IsZero() {
		return EvidenceArtifactRevisionRef{}, fmt.Errorf("core: NewEvidenceArtifactRevisionRef: %w", ErrEmptyIdentity)
	}
	if revisionID.IsZero() {
		return EvidenceArtifactRevisionRef{}, fmt.Errorf("core: NewEvidenceArtifactRevisionRef: %w", ErrMissingRevisionID)
	}
	return EvidenceArtifactRevisionRef{evidenceArtifactID: artifactID, evidenceRevisionID: revisionID}, nil
}

func (r EvidenceArtifactRevisionRef) ArtifactID() ArtifactID         { return r.evidenceArtifactID }
func (r EvidenceArtifactRevisionRef) RevisionID() ArtifactRevisionID { return r.evidenceRevisionID }
func (r EvidenceArtifactRevisionRef) IsZero() bool {
	return r.evidenceArtifactID.IsZero() && r.evidenceRevisionID.IsZero()
}

type evidenceArtifactRevisionRefJSON struct {
	ArtifactID ArtifactID         `json:"artifact_id"`
	RevisionID ArtifactRevisionID `json:"revision_id"`
}

func (r EvidenceArtifactRevisionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(evidenceArtifactRevisionRefJSON{ArtifactID: r.evidenceArtifactID, RevisionID: r.evidenceRevisionID})
}

func (r *EvidenceArtifactRevisionRef) UnmarshalJSON(data []byte) error {
	var raw evidenceArtifactRevisionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal EvidenceArtifactRevisionRef: %w", err)
	}
	v, err := NewEvidenceArtifactRevisionRef(raw.ArtifactID, raw.RevisionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ValidationClaimRef cites a Validation Claim, or any of its
// specializations, by identity (PEOS-006). This is used for citation
// fields such as a Claim's replacement reference, never for resolving
// which Claim is "current."
type ValidationClaimRef struct{ validationClaimRefID ValidationClaimID }

func NewValidationClaimRef(id ValidationClaimID) (ValidationClaimRef, error) {
	if id.IsZero() {
		return ValidationClaimRef{}, fmt.Errorf("core: NewValidationClaimRef: %w", ErrEmptyIdentity)
	}
	return ValidationClaimRef{validationClaimRefID: id}, nil
}

func (r ValidationClaimRef) ClaimID() ValidationClaimID { return r.validationClaimRefID }
func (r ValidationClaimRef) IsZero() bool               { return r.validationClaimRefID.IsZero() }

type validationClaimRefJSON struct {
	ClaimID ValidationClaimID `json:"claim_id"`
}

func (r ValidationClaimRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(validationClaimRefJSON{ClaimID: r.validationClaimRefID})
}

func (r *ValidationClaimRef) UnmarshalJSON(data []byte) error {
	var raw validationClaimRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ValidationClaimRef: %w", err)
	}
	v, err := NewValidationClaimRef(raw.ClaimID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ValidationExecutionRecordRef cites a Validation Execution Record, or
// its Measurement Record specialization, by identity (PEOS-006/007).
type ValidationExecutionRecordRef struct {
	validationExecutionRecordRefID ValidationExecutionRecordID
}

func NewValidationExecutionRecordRef(id ValidationExecutionRecordID) (ValidationExecutionRecordRef, error) {
	if id.IsZero() {
		return ValidationExecutionRecordRef{}, fmt.Errorf("core: NewValidationExecutionRecordRef: %w", ErrEmptyIdentity)
	}
	return ValidationExecutionRecordRef{validationExecutionRecordRefID: id}, nil
}

func (r ValidationExecutionRecordRef) RecordID() ValidationExecutionRecordID {
	return r.validationExecutionRecordRefID
}
func (r ValidationExecutionRecordRef) IsZero() bool { return r.validationExecutionRecordRefID.IsZero() }

type validationExecutionRecordRefJSON struct {
	RecordID ValidationExecutionRecordID `json:"record_id"`
}

func (r ValidationExecutionRecordRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(validationExecutionRecordRefJSON{RecordID: r.validationExecutionRecordRefID})
}

func (r *ValidationExecutionRecordRef) UnmarshalJSON(data []byte) error {
	var raw validationExecutionRecordRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal ValidationExecutionRecordRef: %w", err)
	}
	v, err := NewValidationExecutionRecordRef(raw.RecordID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// LifecycleDefinitionRef identifies a Lifecycle Definition at the identity
// level (PEOS-003). A Lifecycle Definition has its own normative identity,
// independent of ArtifactID (see LifecycleDefinitionID in identity.go).
type LifecycleDefinitionRef struct{ lifecycleDefinitionRefID LifecycleDefinitionID }

// NewLifecycleDefinitionRef validates id and returns a
// LifecycleDefinitionRef.
func NewLifecycleDefinitionRef(id LifecycleDefinitionID) (LifecycleDefinitionRef, error) {
	if id.IsZero() {
		return LifecycleDefinitionRef{}, fmt.Errorf("core: NewLifecycleDefinitionRef: %w", ErrEmptyIdentity)
	}
	return LifecycleDefinitionRef{lifecycleDefinitionRefID: id}, nil
}

func (r LifecycleDefinitionRef) LifecycleDefinitionID() LifecycleDefinitionID {
	return r.lifecycleDefinitionRefID
}
func (r LifecycleDefinitionRef) IsZero() bool { return r.lifecycleDefinitionRefID.IsZero() }

type lifecycleDefinitionRefJSON struct {
	LifecycleDefinitionID LifecycleDefinitionID `json:"lifecycle_definition_id"`
}

func (r LifecycleDefinitionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(lifecycleDefinitionRefJSON{LifecycleDefinitionID: r.lifecycleDefinitionRefID})
}

func (r *LifecycleDefinitionRef) UnmarshalJSON(data []byte) error {
	var raw lifecycleDefinitionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal LifecycleDefinitionRef: %w", err)
	}
	v, err := NewLifecycleDefinitionRef(raw.LifecycleDefinitionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// LifecycleDefinitionVersionRef identifies an exact Lifecycle Definition
// Version (PEOS-003): the owning Lifecycle Definition plus the exact
// Version. A Lifecycle Definition Version is not an Artifact Revision (see
// LifecycleDefinitionID's own doc comment), so this reference pairs
// LifecycleDefinitionID with LifecycleDefinitionVersionID rather than
// ArtifactID with ArtifactRevisionID.
type LifecycleDefinitionVersionRef struct {
	lifecycleDefinitionVersionRefDefinitionID LifecycleDefinitionID
	lifecycleDefinitionVersionRefVersionID    LifecycleDefinitionVersionID
}

// NewLifecycleDefinitionVersionRef validates definitionID and versionID
// and returns a LifecycleDefinitionVersionRef.
func NewLifecycleDefinitionVersionRef(definitionID LifecycleDefinitionID, versionID LifecycleDefinitionVersionID) (LifecycleDefinitionVersionRef, error) {
	if definitionID.IsZero() {
		return LifecycleDefinitionVersionRef{}, fmt.Errorf("core: NewLifecycleDefinitionVersionRef: %w", ErrEmptyIdentity)
	}
	if versionID.IsZero() {
		return LifecycleDefinitionVersionRef{}, fmt.Errorf("core: NewLifecycleDefinitionVersionRef: %w", ErrMissingRevisionID)
	}
	return LifecycleDefinitionVersionRef{
		lifecycleDefinitionVersionRefDefinitionID: definitionID,
		lifecycleDefinitionVersionRefVersionID:    versionID,
	}, nil
}

func (r LifecycleDefinitionVersionRef) LifecycleDefinitionID() LifecycleDefinitionID {
	return r.lifecycleDefinitionVersionRefDefinitionID
}
func (r LifecycleDefinitionVersionRef) VersionID() LifecycleDefinitionVersionID {
	return r.lifecycleDefinitionVersionRefVersionID
}
func (r LifecycleDefinitionVersionRef) IsZero() bool {
	return r.lifecycleDefinitionVersionRefDefinitionID.IsZero() && r.lifecycleDefinitionVersionRefVersionID.IsZero()
}

type lifecycleDefinitionVersionRefJSON struct {
	LifecycleDefinitionID        LifecycleDefinitionID        `json:"lifecycle_definition_id"`
	LifecycleDefinitionVersionID LifecycleDefinitionVersionID `json:"lifecycle_definition_version_id"`
}

func (r LifecycleDefinitionVersionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(lifecycleDefinitionVersionRefJSON{
		LifecycleDefinitionID:        r.lifecycleDefinitionVersionRefDefinitionID,
		LifecycleDefinitionVersionID: r.lifecycleDefinitionVersionRefVersionID,
	})
}

func (r *LifecycleDefinitionVersionRef) UnmarshalJSON(data []byte) error {
	var raw lifecycleDefinitionVersionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal LifecycleDefinitionVersionRef: %w", err)
	}
	v, err := NewLifecycleDefinitionVersionRef(raw.LifecycleDefinitionID, raw.LifecycleDefinitionVersionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// StateAssignmentRef identifies a State Assignment at the identity level
// (PEOS-003). State Assignment is an immutable record, not an Artifact
// (see StateAssignmentID's own doc comment).
type StateAssignmentRef struct{ stateAssignmentRefID StateAssignmentID }

// NewStateAssignmentRef validates id and returns a StateAssignmentRef.
func NewStateAssignmentRef(id StateAssignmentID) (StateAssignmentRef, error) {
	if id.IsZero() {
		return StateAssignmentRef{}, fmt.Errorf("core: NewStateAssignmentRef: %w", ErrEmptyIdentity)
	}
	return StateAssignmentRef{stateAssignmentRefID: id}, nil
}

func (r StateAssignmentRef) StateAssignmentID() StateAssignmentID { return r.stateAssignmentRefID }
func (r StateAssignmentRef) IsZero() bool                         { return r.stateAssignmentRefID.IsZero() }

type stateAssignmentRefJSON struct {
	StateAssignmentID StateAssignmentID `json:"state_assignment_id"`
}

func (r StateAssignmentRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(stateAssignmentRefJSON{StateAssignmentID: r.stateAssignmentRefID})
}

func (r *StateAssignmentRef) UnmarshalJSON(data []byte) error {
	var raw stateAssignmentRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal StateAssignmentRef: %w", err)
	}
	v, err := NewStateAssignmentRef(raw.StateAssignmentID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// LifecycleDefinitionVersionSupersessionRef identifies a Lifecycle
// Definition Version Supersession at the identity level (PEOS-003
// "Supersession"). Like State Assignment, it is an immutable record, not
// an Artifact.
type LifecycleDefinitionVersionSupersessionRef struct {
	lifecycleDefinitionVersionSupersessionRefID LifecycleDefinitionVersionSupersessionID
}

// NewLifecycleDefinitionVersionSupersessionRef validates id and returns a
// LifecycleDefinitionVersionSupersessionRef.
func NewLifecycleDefinitionVersionSupersessionRef(id LifecycleDefinitionVersionSupersessionID) (LifecycleDefinitionVersionSupersessionRef, error) {
	if id.IsZero() {
		return LifecycleDefinitionVersionSupersessionRef{}, fmt.Errorf("core: NewLifecycleDefinitionVersionSupersessionRef: %w", ErrEmptyIdentity)
	}
	return LifecycleDefinitionVersionSupersessionRef{lifecycleDefinitionVersionSupersessionRefID: id}, nil
}

func (r LifecycleDefinitionVersionSupersessionRef) SupersessionID() LifecycleDefinitionVersionSupersessionID {
	return r.lifecycleDefinitionVersionSupersessionRefID
}
func (r LifecycleDefinitionVersionSupersessionRef) IsZero() bool {
	return r.lifecycleDefinitionVersionSupersessionRefID.IsZero()
}

type lifecycleDefinitionVersionSupersessionRefJSON struct {
	SupersessionID LifecycleDefinitionVersionSupersessionID `json:"lifecycle_definition_version_supersession_id"`
}

func (r LifecycleDefinitionVersionSupersessionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(lifecycleDefinitionVersionSupersessionRefJSON{SupersessionID: r.lifecycleDefinitionVersionSupersessionRefID})
}

func (r *LifecycleDefinitionVersionSupersessionRef) UnmarshalJSON(data []byte) error {
	var raw lifecycleDefinitionVersionSupersessionRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal LifecycleDefinitionVersionSupersessionRef: %w", err)
	}
	v, err := NewLifecycleDefinitionVersionSupersessionRef(raw.SupersessionID)
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// Known discriminator values for EngineeringSubjectRef. These are plain
// strings, not a closed Go enum: EngineeringSubjectRef accepts any
// discriminator through its opaque construction path (see
// NewOpaqueEngineeringSubjectRef), and these constants only name the
// kinds this packet gives a typed payload to.
const (
	SubjectKindArtifact                  = "artifact"
	SubjectKindArtifactRevision          = "artifact_revision"
	SubjectKindRequirement               = "requirement"
	SubjectKindRequirementRevision       = "requirement_revision"
	SubjectKindDecision                  = "decision"
	SubjectKindDecisionOutcome           = "decision_outcome"
	SubjectKindEngineeringCommitment     = "engineering_commitment"
	SubjectKindRuntimeSubject            = "runtime_subject"
	SubjectKindRuntimeContract           = "runtime_contract"
	SubjectKindRuntimeContractRevision   = "runtime_contract_revision"
	SubjectKindTemplate                  = "template"
	SubjectKindTemplateRevision          = "template_revision"
	SubjectKindGeneratedArtifact         = "generated_artifact"
	SubjectKindGeneratedArtifactRevision = "generated_artifact_revision"
	SubjectKindValidationPlan            = "validation_plan"
	SubjectKindValidationPlanRevision    = "validation_plan_revision"
	SubjectKindEvidence                  = "evidence"
)

var knownSubjectKinds = map[string]bool{
	SubjectKindArtifact:                  true,
	SubjectKindArtifactRevision:          true,
	SubjectKindRequirement:               true,
	SubjectKindRequirementRevision:       true,
	SubjectKindDecision:                  true,
	SubjectKindDecisionOutcome:           true,
	SubjectKindEngineeringCommitment:     true,
	SubjectKindRuntimeSubject:            true,
	SubjectKindRuntimeContract:           true,
	SubjectKindRuntimeContractRevision:   true,
	SubjectKindTemplate:                  true,
	SubjectKindTemplateRevision:          true,
	SubjectKindGeneratedArtifact:         true,
	SubjectKindGeneratedArtifactRevision: true,
	SubjectKindValidationPlan:            true,
	SubjectKindValidationPlanRevision:    true,
	SubjectKindEvidence:                  true,
}

// OpaqueEngineeringSubject carries the discriminator and namespaced
// identity of an EngineeringSubjectRef whose kind this packet does not
// give a dedicated typed payload to. It is preserved losslessly only when
// its wire representation fits the (namespace, identifier) shape; a
// future PEOS SDK packet MAY add a dedicated typed payload for one of
// these kinds without breaking callers that only used the opaque form.
type OpaqueEngineeringSubject struct {
	kind       string
	namespace  string
	identifier string
}

func (o OpaqueEngineeringSubject) Kind() string       { return o.kind }
func (o OpaqueEngineeringSubject) Namespace() string  { return o.namespace }
func (o OpaqueEngineeringSubject) Identifier() string { return o.identifier }

// EngineeringSubjectRef is the canonical tagged union used wherever a
// PEOS construct needs exactly one engineering subject (a Validation,
// Quality, Compliance, or Template Conformance Claim subject; a Lifecycle
// Subject, via LifecycleSubjectRef; a Requirement Subject per PEOS-005
// §10, identifying the engineering matter a Requirement's required
// intent applies to). It is never constructed as an untyped string:
// every known kind carries its own strongly-typed payload, and every
// value is built through a constructor that validates the payload
// before the value exists.
//
// EngineeringSubjectRef is a distinct type from CriterionRef
// (criterion.go) even though a few kinds overlap in shape (for example,
// both can carry a RequirementArtifactRevisionRef). Keeping subject and
// criteria as separate unions lets a future validator mechanically check
// that a Requirement used as Claim criterion never also appears as that
// Claim's subject, without relying on a runtime discriminator comparison
// between two values of what would otherwise be one shared type.
type EngineeringSubjectRef struct {
	kind  string
	known bool

	artifact                  ArtifactRef
	artifactRevision          ArtifactRevisionRef
	requirement               RequirementRef
	requirementRevision       RequirementArtifactRevisionRef
	decision                  DecisionRef
	decisionOutcome           DecisionOutcomeRef
	engineeringCommitment     EngineeringCommitmentRef
	runtimeSubject            RuntimeSubjectRef
	runtimeContract           RuntimeContractRef
	runtimeContractRevision   RuntimeContractRevisionRef
	template                  TemplateRef
	templateRevision          TemplateArtifactRevisionRef
	generatedArtifact         GeneratedArtifactRef
	generatedArtifactRevision GeneratedArtifactRevisionRef
	validationPlan            ValidationPlanRef
	validationPlanRevision    ValidationPlanRevisionRef
	evidence                  EvidenceArtifactRevisionRef

	opaque OpaqueEngineeringSubject
}

// Kind returns the subject's discriminator string.
func (r EngineeringSubjectRef) Kind() string { return r.kind }

// IsKnown reports whether r carries one of this packet's typed payloads,
// as opposed to an opaque, forward-compatible unknown kind.
func (r EngineeringSubjectRef) IsKnown() bool { return r.known }

// IsZero reports whether r is the zero value.
func (r EngineeringSubjectRef) IsZero() bool { return r.kind == "" }

func EngineeringSubjectRefFromArtifact(ref ArtifactRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromArtifact: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindArtifact, known: true, artifact: ref}, nil
}

func (r EngineeringSubjectRef) AsArtifact() (ArtifactRef, bool) {
	if r.kind != SubjectKindArtifact {
		return ArtifactRef{}, false
	}
	return r.artifact, true
}

func EngineeringSubjectRefFromArtifactRevision(ref ArtifactRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromArtifactRevision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindArtifactRevision, known: true, artifactRevision: ref}, nil
}

func (r EngineeringSubjectRef) AsArtifactRevision() (ArtifactRevisionRef, bool) {
	if r.kind != SubjectKindArtifactRevision {
		return ArtifactRevisionRef{}, false
	}
	return r.artifactRevision, true
}

func EngineeringSubjectRefFromRequirement(ref RequirementRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromRequirement: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindRequirement, known: true, requirement: ref}, nil
}

func (r EngineeringSubjectRef) AsRequirement() (RequirementRef, bool) {
	if r.kind != SubjectKindRequirement {
		return RequirementRef{}, false
	}
	return r.requirement, true
}

func EngineeringSubjectRefFromRequirementRevision(ref RequirementArtifactRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromRequirementRevision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindRequirementRevision, known: true, requirementRevision: ref}, nil
}

func (r EngineeringSubjectRef) AsRequirementRevision() (RequirementArtifactRevisionRef, bool) {
	if r.kind != SubjectKindRequirementRevision {
		return RequirementArtifactRevisionRef{}, false
	}
	return r.requirementRevision, true
}

func EngineeringSubjectRefFromDecision(ref DecisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromDecision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindDecision, known: true, decision: ref}, nil
}

func (r EngineeringSubjectRef) AsDecision() (DecisionRef, bool) {
	if r.kind != SubjectKindDecision {
		return DecisionRef{}, false
	}
	return r.decision, true
}

func EngineeringSubjectRefFromDecisionOutcome(ref DecisionOutcomeRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromDecisionOutcome: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindDecisionOutcome, known: true, decisionOutcome: ref}, nil
}

func (r EngineeringSubjectRef) AsDecisionOutcome() (DecisionOutcomeRef, bool) {
	if r.kind != SubjectKindDecisionOutcome {
		return DecisionOutcomeRef{}, false
	}
	return r.decisionOutcome, true
}

func EngineeringSubjectRefFromEngineeringCommitment(ref EngineeringCommitmentRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromEngineeringCommitment: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindEngineeringCommitment, known: true, engineeringCommitment: ref}, nil
}

func (r EngineeringSubjectRef) AsEngineeringCommitment() (EngineeringCommitmentRef, bool) {
	if r.kind != SubjectKindEngineeringCommitment {
		return EngineeringCommitmentRef{}, false
	}
	return r.engineeringCommitment, true
}

func EngineeringSubjectRefFromRuntimeSubject(ref RuntimeSubjectRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromRuntimeSubject: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindRuntimeSubject, known: true, runtimeSubject: ref}, nil
}

func (r EngineeringSubjectRef) AsRuntimeSubject() (RuntimeSubjectRef, bool) {
	if r.kind != SubjectKindRuntimeSubject {
		return RuntimeSubjectRef{}, false
	}
	return r.runtimeSubject, true
}

func EngineeringSubjectRefFromRuntimeContract(ref RuntimeContractRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromRuntimeContract: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindRuntimeContract, known: true, runtimeContract: ref}, nil
}

func (r EngineeringSubjectRef) AsRuntimeContract() (RuntimeContractRef, bool) {
	if r.kind != SubjectKindRuntimeContract {
		return RuntimeContractRef{}, false
	}
	return r.runtimeContract, true
}

func EngineeringSubjectRefFromRuntimeContractRevision(ref RuntimeContractRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromRuntimeContractRevision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindRuntimeContractRevision, known: true, runtimeContractRevision: ref}, nil
}

func (r EngineeringSubjectRef) AsRuntimeContractRevision() (RuntimeContractRevisionRef, bool) {
	if r.kind != SubjectKindRuntimeContractRevision {
		return RuntimeContractRevisionRef{}, false
	}
	return r.runtimeContractRevision, true
}

func EngineeringSubjectRefFromTemplate(ref TemplateRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromTemplate: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindTemplate, known: true, template: ref}, nil
}

func (r EngineeringSubjectRef) AsTemplate() (TemplateRef, bool) {
	if r.kind != SubjectKindTemplate {
		return TemplateRef{}, false
	}
	return r.template, true
}

func EngineeringSubjectRefFromTemplateRevision(ref TemplateArtifactRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromTemplateRevision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindTemplateRevision, known: true, templateRevision: ref}, nil
}

func (r EngineeringSubjectRef) AsTemplateRevision() (TemplateArtifactRevisionRef, bool) {
	if r.kind != SubjectKindTemplateRevision {
		return TemplateArtifactRevisionRef{}, false
	}
	return r.templateRevision, true
}

func EngineeringSubjectRefFromGeneratedArtifact(ref GeneratedArtifactRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromGeneratedArtifact: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindGeneratedArtifact, known: true, generatedArtifact: ref}, nil
}

func (r EngineeringSubjectRef) AsGeneratedArtifact() (GeneratedArtifactRef, bool) {
	if r.kind != SubjectKindGeneratedArtifact {
		return GeneratedArtifactRef{}, false
	}
	return r.generatedArtifact, true
}

func EngineeringSubjectRefFromGeneratedArtifactRevision(ref GeneratedArtifactRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromGeneratedArtifactRevision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindGeneratedArtifactRevision, known: true, generatedArtifactRevision: ref}, nil
}

func (r EngineeringSubjectRef) AsGeneratedArtifactRevision() (GeneratedArtifactRevisionRef, bool) {
	if r.kind != SubjectKindGeneratedArtifactRevision {
		return GeneratedArtifactRevisionRef{}, false
	}
	return r.generatedArtifactRevision, true
}

func EngineeringSubjectRefFromValidationPlan(ref ValidationPlanRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromValidationPlan: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindValidationPlan, known: true, validationPlan: ref}, nil
}

func (r EngineeringSubjectRef) AsValidationPlan() (ValidationPlanRef, bool) {
	if r.kind != SubjectKindValidationPlan {
		return ValidationPlanRef{}, false
	}
	return r.validationPlan, true
}

func EngineeringSubjectRefFromValidationPlanRevision(ref ValidationPlanRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromValidationPlanRevision: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindValidationPlanRevision, known: true, validationPlanRevision: ref}, nil
}

func (r EngineeringSubjectRef) AsValidationPlanRevision() (ValidationPlanRevisionRef, bool) {
	if r.kind != SubjectKindValidationPlanRevision {
		return ValidationPlanRevisionRef{}, false
	}
	return r.validationPlanRevision, true
}

func EngineeringSubjectRefFromEvidence(ref EvidenceArtifactRevisionRef) (EngineeringSubjectRef, error) {
	if ref.IsZero() {
		return EngineeringSubjectRef{}, fmt.Errorf("core: EngineeringSubjectRefFromEvidence: %w", ErrInvalidPayload)
	}
	return EngineeringSubjectRef{kind: SubjectKindEvidence, known: true, evidence: ref}, nil
}

func (r EngineeringSubjectRef) AsEvidence() (EvidenceArtifactRevisionRef, bool) {
	if r.kind != SubjectKindEvidence {
		return EvidenceArtifactRevisionRef{}, false
	}
	return r.evidence, true
}

// NewOpaqueEngineeringSubjectRef constructs a forward-compatible
// EngineeringSubjectRef for a kind this packet does not give a typed
// payload to. kind must be non-empty and must not collide with one of
// this packet's known kinds (use the matching EngineeringSubjectRefFrom*
// constructor for those).
//
// Opaque preservation supports exactly one shape: a namespaced scalar
// reference, carried as the pair (namespace, identifier). It does not
// support composite references — anything that would need more than
// those two plain strings (for example, an Artifact ID paired with a
// Revision ID, composite runtime coordinates, an external system plus an
// identifier plus a revision, or a value scoped by a LocalKey against an
// owning Artifact Revision, the exact shape CriterionRef's own
// QualityElementCriterionRef/RuntimeRuleCriterionRef/
// TemplateConstraintCriterionRef combinators use). A future PEOS subject
// kind with a composite shape cannot be represented through this opaque
// path at all; it requires additive support in this package itself (a
// new known kind: a struct field, a marshal/unmarshal branch, and a
// paired constructor/accessor), not just a caller supplying more data
// through the existing (namespace, identifier) pair.
//
// A malformed or unsupported composite payload therefore fails
// explicitly during decode (see EngineeringSubjectRef.UnmarshalJSON's
// default case) rather than being accepted in a truncated or
// partially-decoded form. No silent data loss occurs: either the
// (namespace, identifier) shape round-trips exactly, or construction and
// decoding both fail with a typed error.
func NewOpaqueEngineeringSubjectRef(kind, namespace, identifier string) (EngineeringSubjectRef, error) {
	k, err := normalizeIdentityValue(kind)
	if err != nil {
		return EngineeringSubjectRef{}, fmt.Errorf("core: NewOpaqueEngineeringSubjectRef: %w", err)
	}
	if knownSubjectKinds[k] {
		return EngineeringSubjectRef{}, fmt.Errorf("core: NewOpaqueEngineeringSubjectRef: %q is a known kind, use its typed constructor: %w", k, ErrInvalidReferenceDiscriminator)
	}
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return EngineeringSubjectRef{}, fmt.Errorf("core: NewOpaqueEngineeringSubjectRef: %w", err)
	}
	id, err := normalizeIdentityValue(identifier)
	if err != nil {
		return EngineeringSubjectRef{}, fmt.Errorf("core: NewOpaqueEngineeringSubjectRef: %w", err)
	}
	return EngineeringSubjectRef{
		kind:   k,
		known:  false,
		opaque: OpaqueEngineeringSubject{kind: k, namespace: ns, identifier: id},
	}, nil
}

// AsOpaque returns r's opaque payload, and true, if r does not carry one
// of this packet's typed payloads.
func (r EngineeringSubjectRef) AsOpaque() (OpaqueEngineeringSubject, bool) {
	if r.known || r.kind == "" {
		return OpaqueEngineeringSubject{}, false
	}
	return r.opaque, true
}

type engineeringSubjectRefEnvelope struct {
	Kind string          `json:"kind"`
	Ref  json.RawMessage `json:"ref"`
}

type opaqueSubjectPayloadJSON struct {
	Namespace  string `json:"namespace"`
	Identifier string `json:"identifier"`
}

// MarshalJSON encodes r as {"kind": ..., "ref": ...}, where "ref" is the
// JSON form of whichever concrete reference type r's kind selects.
func (r EngineeringSubjectRef) MarshalJSON() ([]byte, error) {
	if r.kind == "" {
		return nil, fmt.Errorf("core: marshal EngineeringSubjectRef: %w", ErrInvalidReferenceDiscriminator)
	}
	var (
		refBytes []byte
		err      error
	)
	switch {
	case !r.known:
		refBytes, err = json.Marshal(opaqueSubjectPayloadJSON{Namespace: r.opaque.namespace, Identifier: r.opaque.identifier})
	case r.kind == SubjectKindArtifact:
		refBytes, err = json.Marshal(r.artifact)
	case r.kind == SubjectKindArtifactRevision:
		refBytes, err = json.Marshal(r.artifactRevision)
	case r.kind == SubjectKindRequirement:
		refBytes, err = json.Marshal(r.requirement)
	case r.kind == SubjectKindRequirementRevision:
		refBytes, err = json.Marshal(r.requirementRevision)
	case r.kind == SubjectKindDecision:
		refBytes, err = json.Marshal(r.decision)
	case r.kind == SubjectKindDecisionOutcome:
		refBytes, err = json.Marshal(r.decisionOutcome)
	case r.kind == SubjectKindEngineeringCommitment:
		refBytes, err = json.Marshal(r.engineeringCommitment)
	case r.kind == SubjectKindRuntimeSubject:
		refBytes, err = json.Marshal(r.runtimeSubject)
	case r.kind == SubjectKindRuntimeContract:
		refBytes, err = json.Marshal(r.runtimeContract)
	case r.kind == SubjectKindRuntimeContractRevision:
		refBytes, err = json.Marshal(r.runtimeContractRevision)
	case r.kind == SubjectKindTemplate:
		refBytes, err = json.Marshal(r.template)
	case r.kind == SubjectKindTemplateRevision:
		refBytes, err = json.Marshal(r.templateRevision)
	case r.kind == SubjectKindGeneratedArtifact:
		refBytes, err = json.Marshal(r.generatedArtifact)
	case r.kind == SubjectKindGeneratedArtifactRevision:
		refBytes, err = json.Marshal(r.generatedArtifactRevision)
	case r.kind == SubjectKindValidationPlan:
		refBytes, err = json.Marshal(r.validationPlan)
	case r.kind == SubjectKindValidationPlanRevision:
		refBytes, err = json.Marshal(r.validationPlanRevision)
	case r.kind == SubjectKindEvidence:
		refBytes, err = json.Marshal(r.evidence)
	default:
		return nil, fmt.Errorf("core: marshal EngineeringSubjectRef: %w", ErrInvalidReferenceDiscriminator)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(engineeringSubjectRefEnvelope{Kind: r.kind, Ref: refBytes})
}

// UnmarshalJSON decodes r from {"kind": ..., "ref": ...}. An unrecognized
// kind is preserved as an opaque payload only if "ref" fits the
// {"namespace": ..., "identifier": ...} shape; otherwise decoding fails,
// since this package never stores an unparsed JSON blob as a payload.
func (r *EngineeringSubjectRef) UnmarshalJSON(data []byte) error {
	var env engineeringSubjectRefEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("core: unmarshal EngineeringSubjectRef: %w", err)
	}
	if env.Kind == "" {
		return fmt.Errorf("core: unmarshal EngineeringSubjectRef: %w", ErrInvalidReferenceDiscriminator)
	}

	var (
		result EngineeringSubjectRef
		err    error
	)
	switch env.Kind {
	case SubjectKindArtifact:
		var ref ArtifactRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromArtifact(ref)
		}
	case SubjectKindArtifactRevision:
		var ref ArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromArtifactRevision(ref)
		}
	case SubjectKindRequirement:
		var ref RequirementRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromRequirement(ref)
		}
	case SubjectKindRequirementRevision:
		var ref RequirementArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromRequirementRevision(ref)
		}
	case SubjectKindDecision:
		var ref DecisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromDecision(ref)
		}
	case SubjectKindDecisionOutcome:
		var ref DecisionOutcomeRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromDecisionOutcome(ref)
		}
	case SubjectKindEngineeringCommitment:
		var ref EngineeringCommitmentRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromEngineeringCommitment(ref)
		}
	case SubjectKindRuntimeSubject:
		var ref RuntimeSubjectRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromRuntimeSubject(ref)
		}
	case SubjectKindRuntimeContract:
		var ref RuntimeContractRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromRuntimeContract(ref)
		}
	case SubjectKindRuntimeContractRevision:
		var ref RuntimeContractRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromRuntimeContractRevision(ref)
		}
	case SubjectKindTemplate:
		var ref TemplateRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromTemplate(ref)
		}
	case SubjectKindTemplateRevision:
		var ref TemplateArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromTemplateRevision(ref)
		}
	case SubjectKindGeneratedArtifact:
		var ref GeneratedArtifactRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromGeneratedArtifact(ref)
		}
	case SubjectKindGeneratedArtifactRevision:
		var ref GeneratedArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromGeneratedArtifactRevision(ref)
		}
	case SubjectKindValidationPlan:
		var ref ValidationPlanRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromValidationPlan(ref)
		}
	case SubjectKindValidationPlanRevision:
		var ref ValidationPlanRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromValidationPlanRevision(ref)
		}
	case SubjectKindEvidence:
		var ref EvidenceArtifactRevisionRef
		if err = json.Unmarshal(env.Ref, &ref); err == nil {
			result, err = EngineeringSubjectRefFromEvidence(ref)
		}
	default:
		var payload opaqueSubjectPayloadJSON
		if err = json.Unmarshal(env.Ref, &payload); err == nil {
			result, err = NewOpaqueEngineeringSubjectRef(env.Kind, payload.Namespace, payload.Identifier)
		} else {
			err = fmt.Errorf("core: unmarshal EngineeringSubjectRef: unrecognized kind %q with non-opaque ref: %w", env.Kind, ErrInvalidPayload)
		}
	}
	if err != nil {
		return err
	}
	*r = result
	return nil
}

// LifecycleSubjectRef identifies the Lifecycle Subject of a State
// Assignment or Transition Record (PEOS-003). PEOS-003 permits a
// Lifecycle Subject to be an Artifact, an Artifact Revision, a
// Requirement, a Requirement Artifact Revision, or a Decision; this
// packet exposes constructors for exactly those, plus an opaque
// constructor for a Lifecycle Subject kind a later packet models (for
// example, a Planned Validation Activity). LifecycleSubjectRef is
// implemented as a restricted view over EngineeringSubjectRef so the two
// unions never drift out of sync on their shared payload shapes; it does
// not itself carry every EngineeringSubjectRef kind, only the ones its
// own constructors below produce.
type LifecycleSubjectRef struct{ ref EngineeringSubjectRef }

func NewLifecycleSubjectRefFromArtifact(ref ArtifactRef) (LifecycleSubjectRef, error) {
	subject, err := EngineeringSubjectRefFromArtifact(ref)
	if err != nil {
		return LifecycleSubjectRef{}, err
	}
	return LifecycleSubjectRef{ref: subject}, nil
}

func NewLifecycleSubjectRefFromArtifactRevision(ref ArtifactRevisionRef) (LifecycleSubjectRef, error) {
	subject, err := EngineeringSubjectRefFromArtifactRevision(ref)
	if err != nil {
		return LifecycleSubjectRef{}, err
	}
	return LifecycleSubjectRef{ref: subject}, nil
}

func NewLifecycleSubjectRefFromRequirement(ref RequirementRef) (LifecycleSubjectRef, error) {
	subject, err := EngineeringSubjectRefFromRequirement(ref)
	if err != nil {
		return LifecycleSubjectRef{}, err
	}
	return LifecycleSubjectRef{ref: subject}, nil
}

func NewLifecycleSubjectRefFromRequirementRevision(ref RequirementArtifactRevisionRef) (LifecycleSubjectRef, error) {
	subject, err := EngineeringSubjectRefFromRequirementRevision(ref)
	if err != nil {
		return LifecycleSubjectRef{}, err
	}
	return LifecycleSubjectRef{ref: subject}, nil
}

func NewLifecycleSubjectRefFromDecision(ref DecisionRef) (LifecycleSubjectRef, error) {
	subject, err := EngineeringSubjectRefFromDecision(ref)
	if err != nil {
		return LifecycleSubjectRef{}, err
	}
	return LifecycleSubjectRef{ref: subject}, nil
}

// NewOpaqueLifecycleSubjectRef constructs a forward-compatible
// LifecycleSubjectRef for a Lifecycle Subject kind this packet does not
// give a typed payload to (for example, a Planned Validation Activity).
func NewOpaqueLifecycleSubjectRef(kind, namespace, identifier string) (LifecycleSubjectRef, error) {
	subject, err := NewOpaqueEngineeringSubjectRef(kind, namespace, identifier)
	if err != nil {
		return LifecycleSubjectRef{}, err
	}
	return LifecycleSubjectRef{ref: subject}, nil
}

// Kind returns the Lifecycle Subject's discriminator string.
func (l LifecycleSubjectRef) Kind() string { return l.ref.Kind() }

// IsZero reports whether l is the zero value.
func (l LifecycleSubjectRef) IsZero() bool { return l.ref.IsZero() }

// Subject returns the underlying EngineeringSubjectRef.
func (l LifecycleSubjectRef) Subject() EngineeringSubjectRef { return l.ref }

func (l LifecycleSubjectRef) MarshalJSON() ([]byte, error) { return l.ref.MarshalJSON() }

func (l *LifecycleSubjectRef) UnmarshalJSON(data []byte) error {
	var ref EngineeringSubjectRef
	if err := json.Unmarshal(data, &ref); err != nil {
		return fmt.Errorf("core: unmarshal LifecycleSubjectRef: %w", err)
	}
	l.ref = ref
	return nil
}
