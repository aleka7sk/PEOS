package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// normalizeIdentityValue applies the single, uniform normalization rule
// used by every identity constructor in this file: surrounding whitespace
// is trimmed, and the result is rejected if empty. No case folding and no
// Unicode normalization is applied, so the caller's exact value (aside
// from surrounding whitespace) is preserved.
func normalizeIdentityValue(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ErrEmptyIdentity
	}
	return trimmed, nil
}

// Each identity type below is an opaque, caller-supplied string carried in
// a struct with a field name unique to that type. PEOS-000-009 never
// mandates a specific identifier encoding (UUID or otherwise); the
// constructors below accept any non-empty value and generate nothing
// themselves.
//
// Field names are deliberately distinct per type. Go permits an explicit
// conversion between two named struct types only when their underlying
// types are identical, which requires identical field names in addition
// to identical field types. Giving every identity type its own field name
// means an explicit conversion such as ArtifactID(someRevisionID) fails to
// compile, not only an implicit assignment.

// ArtifactID is the PEOS Artifact identity (PEOS-002 Artifact Identity).
// It is also the identity used by every Artifact subtype that does not
// define its own identity class (Requirement, Validation Plan, Quality
// Profile, Runtime Contract, Template, Generated Artifact, Decision
// Record, Lifecycle Definition).
type ArtifactID struct{ artifactID string }

// ArtifactRevisionID is the PEOS Artifact Revision identity (PEOS-002
// Artifact Revision Identity). It is only meaningful together with the
// ArtifactID of its owning Artifact; this package does not construct a
// bare ArtifactRevisionID as a complete reference on its own (see
// ArtifactRevisionRef).
type ArtifactRevisionID struct{ artifactRevisionID string }

// ImmutableRecordID is the shared identity type for immutable record
// families that this packet does not yet assign a dedicated identity type
// (for example, Measurement Record). A later packet MAY introduce a
// dedicated identity type for one of these without breaking this packet's
// contract, because every dedicated identity type defined here is
// structurally distinct from ImmutableRecordID and from every other
// identity type. State Assignment (PEOS-003) has its own dedicated
// StateAssignmentID below; Transition Record (PEOS-003) is a persistent
// Artifact per PEOS-003's own text ("A Transition Record is a persistent
// Artifact... A Transition Record MUST conform to PEOS-002") and therefore
// uses core.ArtifactID / core.ArtifactRevisionID, not this type or a
// dedicated ID of its own.
type ImmutableRecordID struct{ immutableRecordID string }

// LifecycleDefinitionID is the identity of a Lifecycle Definition
// (PEOS-003). A Lifecycle Definition has its own normative identity,
// independent of core.ArtifactID: PEOS-003 permits, but does not require,
// representing a Lifecycle Definition as an Artifact ("A Lifecycle
// Definition MAY be represented as an Artifact"), so this identity type
// does not assume Artifact backing. An Artifact-backed representation is
// left to a future additive profile.
type LifecycleDefinitionID struct{ lifecycleDefinitionID string }

// LifecycleDefinitionVersionID is the identity of a Lifecycle Definition
// Version (PEOS-003), meaningful only together with the
// LifecycleDefinitionID of its parent Definition (see
// LifecycleDefinitionVersionRef in reference.go). PEOS-003 requires every
// Lifecycle Definition Version to have "an ordering or version
// identifier"; this identity type itself satisfies that requirement's
// "version identifier" alternative, exactly as ArtifactRevisionID already
// does for Artifact Revisions without this package assuming any inherent
// sort order (PEOS-002 §Revision Ordering: "An implementation MUST NOT
// assume that Revision Identifiers are inherently sortable").
type LifecycleDefinitionVersionID struct{ lifecycleDefinitionVersionID string }

// StateAssignmentID is the identity of a State Assignment (PEOS-003): an
// immutable record, not an Artifact -- PEOS-003 never calls a State
// Assignment "an Artifact" or requires it to conform to PEOS-002, unlike
// Transition Record (see ImmutableRecordID's own comment above).
type StateAssignmentID struct{ stateAssignmentID string }

// DecisionID is the PEOS Decision identity (PEOS-004). It is distinct
// from the identity of the Decision Record Artifact that documents the
// Decision; PEOS-004 permits these identifiers to differ.
type DecisionID struct{ decisionID string }

// ValidationClaimID is the identity of a Validation Claim (PEOS-006) and
// every one of its specializations: Satisfaction Claim, Conformance
// Claim, Quality Claim (PEOS-007), Compliance Claim (PEOS-008), and
// Template Conformance Claim (PEOS-009). PEOS-000-009 does not define a
// separate identity space per Claim specialization.
type ValidationClaimID struct{ validationClaimID string }

// ValidationExecutionRecordID is the identity of a Validation Execution
// Record (PEOS-006), including its Measurement Record specialization
// (PEOS-007).
type ValidationExecutionRecordID struct{ validationExecutionRecordID string }

// RuntimeBindingRecordID is the identity of a Runtime Binding Record
// (PEOS-008).
type RuntimeBindingRecordID struct{ runtimeBindingRecordID string }

// RuntimeUnbindingRecordID is the identity of a Runtime Unbinding Record
// (PEOS-008).
type RuntimeUnbindingRecordID struct{ runtimeUnbindingRecordID string }

// RuntimeObservationID is the identity of a Runtime Observation
// (PEOS-008).
type RuntimeObservationID struct{ runtimeObservationID string }

// RuntimeViolationID is the identity of a Runtime Violation (PEOS-008).
type RuntimeViolationID struct{ runtimeViolationID string }

// TemplateApplicationRecordID is the identity of a Template Application
// Record (PEOS-009).
type TemplateApplicationRecordID struct{ templateApplicationRecordID string }

// ControlledVocabularyID identifies a registry entry for a controlled
// vocabulary that is independently identified beyond its namespaced
// value (see VocabularyValue in vocabulary.go). It is distinct from a
// VocabularyValue: a VocabularyValue is the namespace:value pair itself,
// while a ControlledVocabularyID identifies one specific registry record
// about such a value where PEOS or a Product contract tracks one.
type ControlledVocabularyID struct{ controlledVocabularyID string }

// LocalKey is the identity used by Artifact-Revision-owned value
// structures that PEOS scopes only to their exact owning Revision (the
// Planned Validation Activity plan-local key of PEOS-006, and the
// Template Parameter template-local key of PEOS-009). A LocalKey is
// deliberately not comparable across two different owning Revisions by
// this type alone; the owning Revision reference must be carried
// alongside it by the construct that uses it.
type LocalKey struct{ localKey string }

// NewArtifactID validates value and returns an ArtifactID.
func NewArtifactID(value string) (ArtifactID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return ArtifactID{}, fmt.Errorf("core: NewArtifactID: %w", err)
	}
	return ArtifactID{artifactID: v}, nil
}

// String returns the opaque identity value.
func (id ArtifactID) String() string { return id.artifactID }

// IsZero reports whether id is the zero value.
func (id ArtifactID) IsZero() bool { return id.artifactID == "" }

// MarshalJSON encodes id as a JSON string.
func (id ArtifactID) MarshalJSON() ([]byte, error) { return json.Marshal(id.artifactID) }

// UnmarshalJSON decodes id from a JSON string, applying the same
// validation as NewArtifactID.
func (id *ArtifactID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal ArtifactID: %w", err)
	}
	v, err := NewArtifactID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewArtifactRevisionID validates value and returns an ArtifactRevisionID.
func NewArtifactRevisionID(value string) (ArtifactRevisionID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return ArtifactRevisionID{}, fmt.Errorf("core: NewArtifactRevisionID: %w", err)
	}
	return ArtifactRevisionID{artifactRevisionID: v}, nil
}

func (id ArtifactRevisionID) String() string { return id.artifactRevisionID }
func (id ArtifactRevisionID) IsZero() bool   { return id.artifactRevisionID == "" }

func (id ArtifactRevisionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.artifactRevisionID)
}

func (id *ArtifactRevisionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal ArtifactRevisionID: %w", err)
	}
	v, err := NewArtifactRevisionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewImmutableRecordID validates value and returns an ImmutableRecordID.
func NewImmutableRecordID(value string) (ImmutableRecordID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return ImmutableRecordID{}, fmt.Errorf("core: NewImmutableRecordID: %w", err)
	}
	return ImmutableRecordID{immutableRecordID: v}, nil
}

func (id ImmutableRecordID) String() string { return id.immutableRecordID }
func (id ImmutableRecordID) IsZero() bool   { return id.immutableRecordID == "" }

func (id ImmutableRecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.immutableRecordID)
}

func (id *ImmutableRecordID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal ImmutableRecordID: %w", err)
	}
	v, err := NewImmutableRecordID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewDecisionID validates value and returns a DecisionID.
func NewDecisionID(value string) (DecisionID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return DecisionID{}, fmt.Errorf("core: NewDecisionID: %w", err)
	}
	return DecisionID{decisionID: v}, nil
}

func (id DecisionID) String() string { return id.decisionID }
func (id DecisionID) IsZero() bool   { return id.decisionID == "" }

func (id DecisionID) MarshalJSON() ([]byte, error) { return json.Marshal(id.decisionID) }

func (id *DecisionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal DecisionID: %w", err)
	}
	v, err := NewDecisionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewValidationClaimID validates value and returns a ValidationClaimID.
func NewValidationClaimID(value string) (ValidationClaimID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return ValidationClaimID{}, fmt.Errorf("core: NewValidationClaimID: %w", err)
	}
	return ValidationClaimID{validationClaimID: v}, nil
}

func (id ValidationClaimID) String() string { return id.validationClaimID }
func (id ValidationClaimID) IsZero() bool   { return id.validationClaimID == "" }

func (id ValidationClaimID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.validationClaimID)
}

func (id *ValidationClaimID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal ValidationClaimID: %w", err)
	}
	v, err := NewValidationClaimID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewValidationExecutionRecordID validates value and returns a
// ValidationExecutionRecordID.
func NewValidationExecutionRecordID(value string) (ValidationExecutionRecordID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return ValidationExecutionRecordID{}, fmt.Errorf("core: NewValidationExecutionRecordID: %w", err)
	}
	return ValidationExecutionRecordID{validationExecutionRecordID: v}, nil
}

func (id ValidationExecutionRecordID) String() string { return id.validationExecutionRecordID }
func (id ValidationExecutionRecordID) IsZero() bool   { return id.validationExecutionRecordID == "" }

func (id ValidationExecutionRecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.validationExecutionRecordID)
}

func (id *ValidationExecutionRecordID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal ValidationExecutionRecordID: %w", err)
	}
	v, err := NewValidationExecutionRecordID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewRuntimeBindingRecordID validates value and returns a
// RuntimeBindingRecordID.
func NewRuntimeBindingRecordID(value string) (RuntimeBindingRecordID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return RuntimeBindingRecordID{}, fmt.Errorf("core: NewRuntimeBindingRecordID: %w", err)
	}
	return RuntimeBindingRecordID{runtimeBindingRecordID: v}, nil
}

func (id RuntimeBindingRecordID) String() string { return id.runtimeBindingRecordID }
func (id RuntimeBindingRecordID) IsZero() bool   { return id.runtimeBindingRecordID == "" }

func (id RuntimeBindingRecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.runtimeBindingRecordID)
}

func (id *RuntimeBindingRecordID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeBindingRecordID: %w", err)
	}
	v, err := NewRuntimeBindingRecordID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewRuntimeUnbindingRecordID validates value and returns a
// RuntimeUnbindingRecordID.
func NewRuntimeUnbindingRecordID(value string) (RuntimeUnbindingRecordID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return RuntimeUnbindingRecordID{}, fmt.Errorf("core: NewRuntimeUnbindingRecordID: %w", err)
	}
	return RuntimeUnbindingRecordID{runtimeUnbindingRecordID: v}, nil
}

func (id RuntimeUnbindingRecordID) String() string { return id.runtimeUnbindingRecordID }
func (id RuntimeUnbindingRecordID) IsZero() bool   { return id.runtimeUnbindingRecordID == "" }

func (id RuntimeUnbindingRecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.runtimeUnbindingRecordID)
}

func (id *RuntimeUnbindingRecordID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeUnbindingRecordID: %w", err)
	}
	v, err := NewRuntimeUnbindingRecordID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewRuntimeObservationID validates value and returns a
// RuntimeObservationID.
func NewRuntimeObservationID(value string) (RuntimeObservationID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return RuntimeObservationID{}, fmt.Errorf("core: NewRuntimeObservationID: %w", err)
	}
	return RuntimeObservationID{runtimeObservationID: v}, nil
}

func (id RuntimeObservationID) String() string { return id.runtimeObservationID }
func (id RuntimeObservationID) IsZero() bool   { return id.runtimeObservationID == "" }

func (id RuntimeObservationID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.runtimeObservationID)
}

func (id *RuntimeObservationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeObservationID: %w", err)
	}
	v, err := NewRuntimeObservationID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewRuntimeViolationID validates value and returns a RuntimeViolationID.
func NewRuntimeViolationID(value string) (RuntimeViolationID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return RuntimeViolationID{}, fmt.Errorf("core: NewRuntimeViolationID: %w", err)
	}
	return RuntimeViolationID{runtimeViolationID: v}, nil
}

func (id RuntimeViolationID) String() string { return id.runtimeViolationID }
func (id RuntimeViolationID) IsZero() bool   { return id.runtimeViolationID == "" }

func (id RuntimeViolationID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.runtimeViolationID)
}

func (id *RuntimeViolationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal RuntimeViolationID: %w", err)
	}
	v, err := NewRuntimeViolationID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewTemplateApplicationRecordID validates value and returns a
// TemplateApplicationRecordID.
func NewTemplateApplicationRecordID(value string) (TemplateApplicationRecordID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return TemplateApplicationRecordID{}, fmt.Errorf("core: NewTemplateApplicationRecordID: %w", err)
	}
	return TemplateApplicationRecordID{templateApplicationRecordID: v}, nil
}

func (id TemplateApplicationRecordID) String() string { return id.templateApplicationRecordID }
func (id TemplateApplicationRecordID) IsZero() bool   { return id.templateApplicationRecordID == "" }

func (id TemplateApplicationRecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.templateApplicationRecordID)
}

func (id *TemplateApplicationRecordID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal TemplateApplicationRecordID: %w", err)
	}
	v, err := NewTemplateApplicationRecordID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewControlledVocabularyID validates value and returns a
// ControlledVocabularyID.
func NewControlledVocabularyID(value string) (ControlledVocabularyID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return ControlledVocabularyID{}, fmt.Errorf("core: NewControlledVocabularyID: %w", err)
	}
	return ControlledVocabularyID{controlledVocabularyID: v}, nil
}

func (id ControlledVocabularyID) String() string { return id.controlledVocabularyID }
func (id ControlledVocabularyID) IsZero() bool   { return id.controlledVocabularyID == "" }

func (id ControlledVocabularyID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.controlledVocabularyID)
}

func (id *ControlledVocabularyID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal ControlledVocabularyID: %w", err)
	}
	v, err := NewControlledVocabularyID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewLocalKey validates value and returns a LocalKey. A LocalKey is only
// meaningful together with a reference to its owning Artifact Revision;
// this type alone carries no owning-scope information.
func NewLocalKey(value string) (LocalKey, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return LocalKey{}, fmt.Errorf("core: NewLocalKey: %w", err)
	}
	return LocalKey{localKey: v}, nil
}

func (k LocalKey) String() string { return k.localKey }
func (k LocalKey) IsZero() bool   { return k.localKey == "" }

func (k LocalKey) MarshalJSON() ([]byte, error) { return json.Marshal(k.localKey) }

func (k *LocalKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal LocalKey: %w", err)
	}
	v, err := NewLocalKey(s)
	if err != nil {
		return err
	}
	*k = v
	return nil
}

// NewLifecycleDefinitionID validates value and returns a
// LifecycleDefinitionID.
func NewLifecycleDefinitionID(value string) (LifecycleDefinitionID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return LifecycleDefinitionID{}, fmt.Errorf("core: NewLifecycleDefinitionID: %w", err)
	}
	return LifecycleDefinitionID{lifecycleDefinitionID: v}, nil
}

func (id LifecycleDefinitionID) String() string { return id.lifecycleDefinitionID }
func (id LifecycleDefinitionID) IsZero() bool   { return id.lifecycleDefinitionID == "" }

func (id LifecycleDefinitionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.lifecycleDefinitionID)
}

func (id *LifecycleDefinitionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal LifecycleDefinitionID: %w", err)
	}
	v, err := NewLifecycleDefinitionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewLifecycleDefinitionVersionID validates value and returns a
// LifecycleDefinitionVersionID.
func NewLifecycleDefinitionVersionID(value string) (LifecycleDefinitionVersionID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return LifecycleDefinitionVersionID{}, fmt.Errorf("core: NewLifecycleDefinitionVersionID: %w", err)
	}
	return LifecycleDefinitionVersionID{lifecycleDefinitionVersionID: v}, nil
}

func (id LifecycleDefinitionVersionID) String() string { return id.lifecycleDefinitionVersionID }
func (id LifecycleDefinitionVersionID) IsZero() bool   { return id.lifecycleDefinitionVersionID == "" }

func (id LifecycleDefinitionVersionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.lifecycleDefinitionVersionID)
}

func (id *LifecycleDefinitionVersionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal LifecycleDefinitionVersionID: %w", err)
	}
	v, err := NewLifecycleDefinitionVersionID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// NewStateAssignmentID validates value and returns a StateAssignmentID.
func NewStateAssignmentID(value string) (StateAssignmentID, error) {
	v, err := normalizeIdentityValue(value)
	if err != nil {
		return StateAssignmentID{}, fmt.Errorf("core: NewStateAssignmentID: %w", err)
	}
	return StateAssignmentID{stateAssignmentID: v}, nil
}

func (id StateAssignmentID) String() string { return id.stateAssignmentID }
func (id StateAssignmentID) IsZero() bool   { return id.stateAssignmentID == "" }

func (id StateAssignmentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.stateAssignmentID)
}

func (id *StateAssignmentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("core: unmarshal StateAssignmentID: %w", err)
	}
	v, err := NewStateAssignmentID(s)
	if err != nil {
		return err
	}
	*id = v
	return nil
}
