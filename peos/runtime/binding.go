package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file implements the two immutable, independently identifiable,
// non-Artifact records PEOS-008 uses to establish and terminate a runtime
// binding: BindingRecord and UnbindingRecord. Neither wraps core.Artifact
// or core.ArtifactRevision, neither carries lifecycle state, and neither is
// ever rewritten -- see doc.go for the full record ontology and the
// derived-view boundary (Current Runtime Binding is never a stored field
// on either type).

// --- BindingRecord -------------------------------------------------------------

// BindingRecord is a PEOS-008 Runtime Binding Record: "an immutable
// record", "independently identifiable", "not an Artifact. It is not
// revisioned. It is not lifecycle-bearing. It is not a State Assignment."
//
// A BindingRecord "does not mutate the Runtime Contract Revision it
// binds" and "does not, by itself, establish Requirement satisfaction or
// runtime compliance" -- both are structural consequences of BindingRecord
// exposing no modifier that reaches ContractRevision, and of this package
// storing no derived compliance anywhere. Establishing which binding is
// "current" for a runtime subject is a repository-owned derived view (see
// doc.go); no field here is named or shaped to answer that question.
type BindingRecord struct {
	id               core.RuntimeBindingRecordID
	contractRevision core.RuntimeContractRevisionRef
	subject          core.RuntimeSubjectRef
	environment      Environment
	scope            core.Scope
	boundAt          core.Timestamp
	actor            core.ActorRef
	provenance       core.Provenance

	deploymentAt     core.Timestamp
	authority        core.AuthorityRef
	configurationRef string
	limitations      []string
	correction       core.RecordCorrectionRef[core.RuntimeBindingRecordRef]
	extension        core.Extension
}

// NewBindingRecord validates its eight mandatory arguments and returns a
// BindingRecord with no deployment timestamp, authority, configuration
// reference, limitations, correction, or extension. Use the With* methods
// to add those.
//
// contractRevision must name the exact bound Runtime Contract Revision --
// core.RuntimeContractRevisionRef, never the bare core.RuntimeContractRef
// or an implicit "latest Revision" -- because PEOS-008 requires "the exact
// Runtime Contract Artifact Revision bound".
func NewBindingRecord(
	id core.RuntimeBindingRecordID,
	contractRevision core.RuntimeContractRevisionRef,
	subject core.RuntimeSubjectRef,
	environment Environment,
	scope core.Scope,
	boundAt core.Timestamp,
	actor core.ActorRef,
	provenance core.Provenance,
) (BindingRecord, error) {
	if id.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: id must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	if contractRevision.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: contract revision must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	if subject.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: subject must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	if environment.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: environment must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	if scope.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if boundAt.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: bound-at timestamp must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	if actor.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: actor must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	if provenance.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: NewBindingRecord: %w: provenance must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	return BindingRecord{
		id:               id,
		contractRevision: contractRevision,
		subject:          subject,
		environment:      environment,
		scope:            scope,
		boundAt:          boundAt,
		actor:            actor,
		provenance:       provenance,
	}, nil
}

// WithDeploymentTimestamp returns a copy of b with an optional deployment
// timestamp set, distinct from the binding timestamp. timestamp must be
// non-zero; use WithoutDeploymentTimestamp to clear it.
func (b BindingRecord) WithDeploymentTimestamp(timestamp core.Timestamp) (BindingRecord, error) {
	if timestamp.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: BindingRecord.WithDeploymentTimestamp: %w: timestamp must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	b.deploymentAt = timestamp
	return b, nil
}

// WithoutDeploymentTimestamp returns a copy of b with its deployment
// timestamp cleared.
func (b BindingRecord) WithoutDeploymentTimestamp() BindingRecord {
	b.deploymentAt = core.Timestamp{}
	return b
}

// WithAuthority returns a copy of b with its authority set. authority must
// be non-zero; use WithoutAuthority to clear it. PEOS-008 requires
// authority only "where required".
func (b BindingRecord) WithAuthority(authority core.AuthorityRef) (BindingRecord, error) {
	if authority.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: BindingRecord.WithAuthority: %w: authority must not be zero", ErrInvalidRuntimeBindingRecord)
	}
	b.authority = authority
	return b, nil
}

// WithoutAuthority returns a copy of b with its authority cleared.
func (b BindingRecord) WithoutAuthority() BindingRecord {
	b.authority = core.AuthorityRef{}
	return b
}

// WithConfigurationReference returns a copy of b with an optional
// configuration or deployment reference set. reference must be non-empty
// after trimming; the trimmed value is stored. Use
// WithoutConfigurationReference to clear it.
//
// PEOS-008 requires this "where required" but defines no provider or
// deployment-platform schema for it. A validated, trimmed, opaque string is
// the smallest representation that satisfies the specification without
// inventing a ProviderRef, EndpointRef, DeploymentRef, SecretRef, or
// CredentialRef -- none of which PEOS-008 names, and none of which this
// package stores secret material in.
func (b BindingRecord) WithConfigurationReference(reference string) (BindingRecord, error) {
	trimmed, err := trimmedRequired("BindingRecord.WithConfigurationReference", "configuration reference", reference, ErrInvalidRuntimeBindingRecord)
	if err != nil {
		return BindingRecord{}, err
	}
	b.configurationRef = trimmed
	return b, nil
}

// WithoutConfigurationReference returns a copy of b with its configuration
// or deployment reference cleared.
func (b BindingRecord) WithoutConfigurationReference() BindingRecord {
	b.configurationRef = ""
	return b
}

// WithLimitations returns a copy of b with its known-limitations
// descriptions set to exactly the values given, in the order given. Each
// entry is trimmed and must be non-empty after trimming. Passing an empty
// or nil slice declares none, which is why there is no
// WithoutLimitations: WithLimitations(nil) already expresses removal.
func (b BindingRecord) WithLimitations(limitations []string) (BindingRecord, error) {
	cp, err := trimmedStringSlice("BindingRecord.WithLimitations", "limitation", limitations, ErrInvalidRuntimeBindingRecord)
	if err != nil {
		return BindingRecord{}, err
	}
	b.limitations = cp
	return b, nil
}

// WithCorrection returns a copy of b referencing an earlier BindingRecord
// it explicitly corrects, replaces, or invalidates. correction must be
// non-zero and must not target b's own identity -- a Binding Record cannot
// correct itself. Use WithoutCorrection to clear it.
//
// "Record replacement SHALL NOT be described using the normative term
// Supersession" -- this correction reference is PEOS-006's correct/
// replace/invalidate vocabulary (core.CorrectionKind), never PEOS-002
// Artifact Supersession, and it never mutates or removes the earlier
// record: "The earlier record remains historically preserved."
func (b BindingRecord) WithCorrection(correction core.RecordCorrectionRef[core.RuntimeBindingRecordRef]) (BindingRecord, error) {
	if correction.IsZero() {
		return BindingRecord{}, fmt.Errorf("runtime: BindingRecord.WithCorrection: %w: correction must not be zero", core.ErrInvalidCorrectionReference)
	}
	if !b.id.IsZero() && correction.Target().RecordID() == b.id {
		return BindingRecord{}, fmt.Errorf("runtime: BindingRecord.WithCorrection: %w: a binding record must not correct itself", core.ErrInvalidCorrectionReference)
	}
	b.correction = correction
	return b, nil
}

// WithoutCorrection returns a copy of b with its correction reference
// cleared.
func (b BindingRecord) WithoutCorrection() BindingRecord {
	b.correction = core.RecordCorrectionRef[core.RuntimeBindingRecordRef]{}
	return b
}

// WithExtension returns a copy of b with its extension data set. Passing
// the zero core.Extension is equivalent to declaring none.
func (b BindingRecord) WithExtension(extension core.Extension) BindingRecord {
	b.extension = extension
	return b
}

// WithoutExtension returns a copy of b with its extension data cleared.
func (b BindingRecord) WithoutExtension() BindingRecord {
	b.extension = core.Extension{}
	return b
}

// ID returns b's identity.
func (b BindingRecord) ID() core.RuntimeBindingRecordID { return b.id }

// Ref returns a core.RuntimeBindingRecordRef identifying b.
func (b BindingRecord) Ref() (core.RuntimeBindingRecordRef, error) {
	return core.NewRuntimeBindingRecordRef(b.id)
}

// ContractRevision returns the exact Runtime Contract Revision b binds.
func (b BindingRecord) ContractRevision() core.RuntimeContractRevisionRef { return b.contractRevision }

// Subject returns the exact runtime subject or deployment target b binds
// to.
func (b BindingRecord) Subject() core.RuntimeSubjectRef { return b.subject }

// Environment returns b's declared environment.
func (b BindingRecord) Environment() Environment { return b.environment }

// Scope returns b's declared scope.
func (b BindingRecord) Scope() core.Scope { return b.scope }

// BoundAt returns b's binding timestamp.
func (b BindingRecord) BoundAt() core.Timestamp { return b.boundAt }

// Actor returns the actor who established b.
func (b BindingRecord) Actor() core.ActorRef { return b.actor }

// Provenance returns b's declared provenance.
func (b BindingRecord) Provenance() core.Provenance { return b.provenance }

// DeploymentTimestamp returns b's deployment timestamp, and whether one is
// set, where distinct from BoundAt.
func (b BindingRecord) DeploymentTimestamp() (core.Timestamp, bool) {
	return b.deploymentAt, !b.deploymentAt.IsZero()
}

// Authority returns b's declared authority, and whether one is set.
func (b BindingRecord) Authority() (core.AuthorityRef, bool) {
	return b.authority, !b.authority.IsZero()
}

// ConfigurationReference returns b's configuration or deployment
// reference, and whether one is set.
func (b BindingRecord) ConfigurationReference() (string, bool) {
	return b.configurationRef, b.configurationRef != ""
}

// Limitations returns a defensive copy of b's known-limitations
// descriptions, in declaration order.
func (b BindingRecord) Limitations() []string { return copySlice(b.limitations) }

// Correction returns b's correction reference, and whether one is set.
func (b BindingRecord) Correction() (core.RecordCorrectionRef[core.RuntimeBindingRecordRef], bool) {
	return b.correction, !b.correction.IsZero()
}

// Extension returns b's extension data.
func (b BindingRecord) Extension() core.Extension { return b.extension }

// IsZero reports whether b is the zero value.
func (b BindingRecord) IsZero() bool {
	return b.id.IsZero() && b.contractRevision.IsZero() && b.subject.IsZero() &&
		b.environment.IsZero() && b.scope.IsZero() && b.boundAt.IsZero() &&
		b.actor.IsZero() && b.provenance.IsZero()
}

type bindingRecordJSON struct {
	ID               core.RuntimeBindingRecordID                             `json:"id"`
	ContractRevision core.RuntimeContractRevisionRef                         `json:"contract_revision"`
	Subject          core.RuntimeSubjectRef                                  `json:"subject"`
	Environment      Environment                                             `json:"environment"`
	Scope            core.Scope                                              `json:"scope"`
	BoundAt          core.Timestamp                                          `json:"bound_at"`
	Actor            core.ActorRef                                           `json:"actor"`
	Provenance       core.Provenance                                         `json:"provenance"`
	DeploymentAt     *core.Timestamp                                         `json:"deployment_at,omitempty"`
	Authority        *core.AuthorityRef                                      `json:"authority,omitempty"`
	ConfigurationRef string                                                  `json:"configuration_reference,omitempty"`
	Limitations      []string                                                `json:"limitations,omitempty"`
	Correction       *core.RecordCorrectionRef[core.RuntimeBindingRecordRef] `json:"correction,omitempty"`
	Extension        *core.Extension                                         `json:"extension,omitempty"`
}

// bindingRecordUnmarshalJSON mirrors bindingRecordJSON for decoding, with
// every optional single value captured as raw bytes so an explicit null
// can be distinguished from an absent key and rejected -- the
// json.RawMessage probe technique Packet D.1 established.
type bindingRecordUnmarshalJSON struct {
	ID               core.RuntimeBindingRecordID     `json:"id"`
	ContractRevision core.RuntimeContractRevisionRef `json:"contract_revision"`
	Subject          core.RuntimeSubjectRef          `json:"subject"`
	Environment      Environment                     `json:"environment"`
	Scope            core.Scope                      `json:"scope"`
	BoundAt          core.Timestamp                  `json:"bound_at"`
	Actor            core.ActorRef                   `json:"actor"`
	Provenance       core.Provenance                 `json:"provenance"`
	DeploymentAt     json.RawMessage                 `json:"deployment_at"`
	Authority        json.RawMessage                 `json:"authority"`
	ConfigurationRef json.RawMessage                 `json:"configuration_reference"`
	Limitations      []string                        `json:"limitations"`
	Correction       json.RawMessage                 `json:"correction"`
	Extension        *core.Extension                 `json:"extension,omitempty"`
}

// MarshalJSON encodes b with its eight mandatory keys always present, plus
// whichever optional keys are set.
//
// There is no "bound", "unbound", "active", "current", "latest",
// "effective", "status", "state", "lifecycle", "compliant", "compliance",
// or "superseded" key: their absence is the structural proof that a
// Runtime Binding Record carries only what was recorded at binding time,
// never derived or current state.
func (b BindingRecord) MarshalJSON() ([]byte, error) {
	if b.IsZero() {
		return nil, fmt.Errorf("runtime: marshal BindingRecord: %w", ErrInvalidRuntimeBindingRecord)
	}
	raw := bindingRecordJSON{
		ID:               b.id,
		ContractRevision: b.contractRevision,
		Subject:          b.subject,
		Environment:      b.environment,
		Scope:            b.scope,
		BoundAt:          b.boundAt,
		Actor:            b.actor,
		Provenance:       b.provenance,
		ConfigurationRef: b.configurationRef,
		Limitations:      b.limitations,
	}
	if !b.deploymentAt.IsZero() {
		raw.DeploymentAt = &b.deploymentAt
	}
	if !b.authority.IsZero() {
		raw.Authority = &b.authority
	}
	if !b.correction.IsZero() {
		raw.Correction = &b.correction
	}
	if !b.extension.IsZero() {
		raw.Extension = &b.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes b from its JSON form, applying the same validation
// as NewBindingRecord and each With* method. The receiver is left
// untouched unless every check passes.
func (b *BindingRecord) UnmarshalJSON(data []byte) error {
	var raw bindingRecordUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal BindingRecord: %w: %w", ErrInvalidRuntimeBindingRecord, err)
	}
	result, err := NewBindingRecord(raw.ID, raw.ContractRevision, raw.Subject, raw.Environment, raw.Scope, raw.BoundAt, raw.Actor, raw.Provenance)
	if err != nil {
		return err
	}
	if len(raw.DeploymentAt) > 0 {
		if err = rejectNullRaw("BindingRecord", "deployment timestamp", raw.DeploymentAt, ErrInvalidRuntimeBindingRecord); err != nil {
			return err
		}
		var ts core.Timestamp
		if err = json.Unmarshal(raw.DeploymentAt, &ts); err != nil {
			return fmt.Errorf("runtime: unmarshal BindingRecord: %w: %w", ErrInvalidRuntimeBindingRecord, err)
		}
		if result, err = result.WithDeploymentTimestamp(ts); err != nil {
			return err
		}
	}
	if len(raw.Authority) > 0 {
		if err = rejectNullRaw("BindingRecord", "authority", raw.Authority, ErrInvalidRuntimeBindingRecord); err != nil {
			return err
		}
		var authority core.AuthorityRef
		if err = json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("runtime: unmarshal BindingRecord: %w: %w", ErrInvalidRuntimeBindingRecord, err)
		}
		if result, err = result.WithAuthority(authority); err != nil {
			return err
		}
	}
	if len(raw.ConfigurationRef) > 0 {
		if err = rejectNullRaw("BindingRecord", "configuration reference", raw.ConfigurationRef, ErrInvalidRuntimeBindingRecord); err != nil {
			return err
		}
		var reference string
		if err = json.Unmarshal(raw.ConfigurationRef, &reference); err != nil {
			return fmt.Errorf("runtime: unmarshal BindingRecord: %w: %w", ErrInvalidRuntimeBindingRecord, err)
		}
		if result, err = result.WithConfigurationReference(reference); err != nil {
			return err
		}
	}
	if len(raw.Limitations) > 0 {
		if result, err = result.WithLimitations(raw.Limitations); err != nil {
			return err
		}
	}
	if len(raw.Correction) > 0 {
		if err = rejectNullRaw("BindingRecord", "correction", raw.Correction, core.ErrInvalidCorrectionReference); err != nil {
			return err
		}
		var correction core.RecordCorrectionRef[core.RuntimeBindingRecordRef]
		if err = json.Unmarshal(raw.Correction, &correction); err != nil {
			return fmt.Errorf("runtime: unmarshal BindingRecord: %w: %w", ErrInvalidRuntimeBindingRecord, err)
		}
		if result, err = result.WithCorrection(correction); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*b = result
	return nil
}

// --- UnbindingRecord -----------------------------------------------------------

// UnbindingRecord is a PEOS-008 Runtime Unbinding Record: "an immutable
// record", "independently identifiable", which "references exactly one
// Runtime Binding Record" and "is not an Artifact. It is not revisioned.
// It is not lifecycle-bearing."
//
// An UnbindingRecord "does not delete, erase, or rewrite its Binding
// Record. The Binding Record remains historically inspectable after
// unbinding." This package enforces that structurally: UnbindingRecord
// holds a reference to the BindingRecord it terminates, never the
// BindingRecord value itself, and BindingRecord exposes no "unbound" or
// "active" field an UnbindingRecord could flip.
type UnbindingRecord struct {
	id           core.RuntimeUnbindingRecordID
	binding      core.RuntimeBindingRecordRef
	subject      core.RuntimeSubjectRef
	terminatedAt core.Timestamp
	reason       string
	actor        core.ActorRef
	provenance   core.Provenance

	authority  core.AuthorityRef
	correction core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef]
	extension  core.Extension
}

// NewUnbindingRecord validates its seven mandatory arguments and returns
// an UnbindingRecord with no authority, correction, or extension. Use the
// With* methods to add those.
//
// binding must name the exact Runtime Binding Record affected --
// core.RuntimeBindingRecordRef -- because PEOS-008 requires "the exact
// Runtime Binding Record affected", never a generic or implicit reference.
// reason must be non-empty after trimming; the trimmed value is stored.
func NewUnbindingRecord(
	id core.RuntimeUnbindingRecordID,
	binding core.RuntimeBindingRecordRef,
	subject core.RuntimeSubjectRef,
	terminatedAt core.Timestamp,
	reason string,
	actor core.ActorRef,
	provenance core.Provenance,
) (UnbindingRecord, error) {
	if id.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: NewUnbindingRecord: %w: id must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	if binding.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: NewUnbindingRecord: %w: binding record reference must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	if subject.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: NewUnbindingRecord: %w: subject must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	if terminatedAt.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: NewUnbindingRecord: %w: terminated-at timestamp must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	trimmedReason, err := trimmedRequired("NewUnbindingRecord", "reason", reason, ErrInvalidRuntimeUnbindingRecord)
	if err != nil {
		return UnbindingRecord{}, err
	}
	if actor.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: NewUnbindingRecord: %w: actor must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	if provenance.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: NewUnbindingRecord: %w: provenance must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	return UnbindingRecord{
		id:           id,
		binding:      binding,
		subject:      subject,
		terminatedAt: terminatedAt,
		reason:       trimmedReason,
		actor:        actor,
		provenance:   provenance,
	}, nil
}

// WithAuthority returns a copy of u with its authority set. authority must
// be non-zero; use WithoutAuthority to clear it.
func (u UnbindingRecord) WithAuthority(authority core.AuthorityRef) (UnbindingRecord, error) {
	if authority.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: UnbindingRecord.WithAuthority: %w: authority must not be zero", ErrInvalidRuntimeUnbindingRecord)
	}
	u.authority = authority
	return u, nil
}

// WithoutAuthority returns a copy of u with its authority cleared.
func (u UnbindingRecord) WithoutAuthority() UnbindingRecord {
	u.authority = core.AuthorityRef{}
	return u
}

// WithCorrection returns a copy of u referencing an earlier
// UnbindingRecord it explicitly corrects, replaces, or invalidates.
// correction must be non-zero and must not target u's own identity. Use
// WithoutCorrection to clear it.
func (u UnbindingRecord) WithCorrection(correction core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef]) (UnbindingRecord, error) {
	if correction.IsZero() {
		return UnbindingRecord{}, fmt.Errorf("runtime: UnbindingRecord.WithCorrection: %w: correction must not be zero", core.ErrInvalidCorrectionReference)
	}
	if !u.id.IsZero() && correction.Target().RecordID() == u.id {
		return UnbindingRecord{}, fmt.Errorf("runtime: UnbindingRecord.WithCorrection: %w: an unbinding record must not correct itself", core.ErrInvalidCorrectionReference)
	}
	u.correction = correction
	return u, nil
}

// WithoutCorrection returns a copy of u with its correction reference
// cleared.
func (u UnbindingRecord) WithoutCorrection() UnbindingRecord {
	u.correction = core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef]{}
	return u
}

// WithExtension returns a copy of u with its extension data set.
func (u UnbindingRecord) WithExtension(extension core.Extension) UnbindingRecord {
	u.extension = extension
	return u
}

// WithoutExtension returns a copy of u with its extension data cleared.
func (u UnbindingRecord) WithoutExtension() UnbindingRecord {
	u.extension = core.Extension{}
	return u
}

// ID returns u's identity.
func (u UnbindingRecord) ID() core.RuntimeUnbindingRecordID { return u.id }

// Ref returns a core.RuntimeUnbindingRecordRef identifying u.
func (u UnbindingRecord) Ref() (core.RuntimeUnbindingRecordRef, error) {
	return core.NewRuntimeUnbindingRecordRef(u.id)
}

// Binding returns the exact Runtime Binding Record u terminates.
func (u UnbindingRecord) Binding() core.RuntimeBindingRecordRef { return u.binding }

// Subject returns the runtime subject u applies to.
func (u UnbindingRecord) Subject() core.RuntimeSubjectRef { return u.subject }

// TerminatedAt returns u's termination timestamp.
func (u UnbindingRecord) TerminatedAt() core.Timestamp { return u.terminatedAt }

// Reason returns u's reason for termination, uninterpreted.
func (u UnbindingRecord) Reason() string { return u.reason }

// Actor returns the actor who recorded u.
func (u UnbindingRecord) Actor() core.ActorRef { return u.actor }

// Provenance returns u's declared provenance.
func (u UnbindingRecord) Provenance() core.Provenance { return u.provenance }

// Authority returns u's declared authority, and whether one is set.
func (u UnbindingRecord) Authority() (core.AuthorityRef, bool) {
	return u.authority, !u.authority.IsZero()
}

// Correction returns u's correction reference, and whether one is set.
func (u UnbindingRecord) Correction() (core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef], bool) {
	return u.correction, !u.correction.IsZero()
}

// Extension returns u's extension data.
func (u UnbindingRecord) Extension() core.Extension { return u.extension }

// IsZero reports whether u is the zero value.
func (u UnbindingRecord) IsZero() bool {
	return u.id.IsZero() && u.binding.IsZero() && u.subject.IsZero() &&
		u.terminatedAt.IsZero() && u.reason == "" && u.actor.IsZero() && u.provenance.IsZero()
}

type unbindingRecordJSON struct {
	ID           core.RuntimeUnbindingRecordID                             `json:"id"`
	Binding      core.RuntimeBindingRecordRef                              `json:"binding"`
	Subject      core.RuntimeSubjectRef                                    `json:"subject"`
	TerminatedAt core.Timestamp                                            `json:"terminated_at"`
	Reason       string                                                    `json:"reason"`
	Actor        core.ActorRef                                             `json:"actor"`
	Provenance   core.Provenance                                           `json:"provenance"`
	Authority    *core.AuthorityRef                                        `json:"authority,omitempty"`
	Correction   *core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef] `json:"correction,omitempty"`
	Extension    *core.Extension                                           `json:"extension,omitempty"`
}

type unbindingRecordUnmarshalJSON struct {
	ID           core.RuntimeUnbindingRecordID `json:"id"`
	Binding      core.RuntimeBindingRecordRef  `json:"binding"`
	Subject      core.RuntimeSubjectRef        `json:"subject"`
	TerminatedAt core.Timestamp                `json:"terminated_at"`
	Reason       string                        `json:"reason"`
	Actor        core.ActorRef                 `json:"actor"`
	Provenance   core.Provenance               `json:"provenance"`
	Authority    json.RawMessage               `json:"authority"`
	Correction   json.RawMessage               `json:"correction"`
	Extension    *core.Extension               `json:"extension,omitempty"`
}

// MarshalJSON encodes u with its seven mandatory keys always present, plus
// whichever optional keys are set. There is no "unbound", "current",
// "latest", or "effective" key.
func (u UnbindingRecord) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return nil, fmt.Errorf("runtime: marshal UnbindingRecord: %w", ErrInvalidRuntimeUnbindingRecord)
	}
	raw := unbindingRecordJSON{
		ID:           u.id,
		Binding:      u.binding,
		Subject:      u.subject,
		TerminatedAt: u.terminatedAt,
		Reason:       u.reason,
		Actor:        u.actor,
		Provenance:   u.provenance,
	}
	if !u.authority.IsZero() {
		raw.Authority = &u.authority
	}
	if !u.correction.IsZero() {
		raw.Correction = &u.correction
	}
	if !u.extension.IsZero() {
		raw.Extension = &u.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes u from its JSON form, applying the same validation
// as NewUnbindingRecord and each With* method. The receiver is left
// untouched unless every check passes.
func (u *UnbindingRecord) UnmarshalJSON(data []byte) error {
	var raw unbindingRecordUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("runtime: unmarshal UnbindingRecord: %w: %w", ErrInvalidRuntimeUnbindingRecord, err)
	}
	result, err := NewUnbindingRecord(raw.ID, raw.Binding, raw.Subject, raw.TerminatedAt, raw.Reason, raw.Actor, raw.Provenance)
	if err != nil {
		return err
	}
	if len(raw.Authority) > 0 {
		if err = rejectNullRaw("UnbindingRecord", "authority", raw.Authority, ErrInvalidRuntimeUnbindingRecord); err != nil {
			return err
		}
		var authority core.AuthorityRef
		if err = json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("runtime: unmarshal UnbindingRecord: %w: %w", ErrInvalidRuntimeUnbindingRecord, err)
		}
		if result, err = result.WithAuthority(authority); err != nil {
			return err
		}
	}
	if len(raw.Correction) > 0 {
		if err = rejectNullRaw("UnbindingRecord", "correction", raw.Correction, core.ErrInvalidCorrectionReference); err != nil {
			return err
		}
		var correction core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef]
		if err = json.Unmarshal(raw.Correction, &correction); err != nil {
			return fmt.Errorf("runtime: unmarshal UnbindingRecord: %w: %w", ErrInvalidRuntimeUnbindingRecord, err)
		}
		if result, err = result.WithCorrection(correction); err != nil {
			return err
		}
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*u = result
	return nil
}
