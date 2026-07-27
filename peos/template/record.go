package template

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file implements the PEOS-009 Template Application Record and the two
// Revision-independent value structures it owns: ResolvedValue and
// GeneratedOutput, plus the ApplicationOutcome vocabulary.
//
// A Template Application Record "is an immutable record", "is independently
// identifiable", and "is not an Artifact. It is not revisioned. It is not
// lifecycle-bearing." It is the authoritative record of a template
// application: "A Template Application Record is the authoritative record of
// a template application. The Generated-From relation, where present, is
// supplementary traceability; it is never a substitute for the Application
// Record."

// --- ApplicationOutcome --------------------------------------------------------

// ApplicationOutcome is a namespaced vocabulary value naming how a template
// application concluded (PEOS-009 Template Application Outcome).
//
// PEOS-009 requires "an outcome drawn from an extensible controlled vocabulary
// including, at minimum: succeeded; failed; partially succeeded; interrupted;
// indeterminate." The five values below are predeclared for exactly that
// reason; the vocabulary is extensible, so NewApplicationOutcome accepts any
// valid core.VocabularyValue and a Product MAY declare its own.
//
// # Why this is not core.ExecutionOutcome
//
// core.ExecutionOutcome is PEOS-006's validation-execution vocabulary. Its
// values are "completed", "failed", "interrupted", and "indeterminate" -- it
// has no "succeeded" (only the differently-scoped "completed") and no
// "partially succeeded" at all, which is the one outcome PEOS-009 attaches an
// extra structural obligation to. Reusing it would both assert a PEOS-006 tie
// PEOS-009 does not state and require adding a PEOS-009-motivated value to a
// PEOS-006-owned vocabulary. This is a template-local wrapper instead, the
// same choice runtime.ViolationSeverity and quality.Unit made for their own
// specifications.
//
// # Not a status
//
// An outcome is a recorded historical fact about one completed application,
// not mutable status and not a lifecycle state: "Outcome is an attribute of
// the Template Application Record. There is no separate Outcome entity", and
// assigning a Lifecycle State to an Application Record is a named
// non-conforming pattern. A record's outcome never changes -- correcting it
// means recording a new record.
type ApplicationOutcome struct{ value core.VocabularyValue }

// NewApplicationOutcome wraps v as an ApplicationOutcome. The vocabulary is
// extensible, so any valid core.VocabularyValue is accepted; the five values
// PEOS-009 names at minimum are predeclared below.
func NewApplicationOutcome(v core.VocabularyValue) ApplicationOutcome {
	return ApplicationOutcome{value: v}
}

// Value returns the underlying core.VocabularyValue.
func (o ApplicationOutcome) Value() core.VocabularyValue { return o.value }
func (o ApplicationOutcome) String() string              { return o.value.String() }
func (o ApplicationOutcome) IsZero() bool                { return o.value.IsZero() }

// Equal reports whether o and other carry the same vocabulary value.
func (o ApplicationOutcome) Equal(other ApplicationOutcome) bool {
	return o.value.Equal(other.value)
}

func (o ApplicationOutcome) MarshalJSON() ([]byte, error) { return json.Marshal(o.value) }

func (o *ApplicationOutcome) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &o.value)
}

func mustApplicationOutcome(value string) ApplicationOutcome {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, value)
	if err != nil {
		panic("template: ApplicationOutcome vocabulary value is invalid: " + err.Error())
	}
	return ApplicationOutcome{value: v}
}

// The five outcomes PEOS-009 names as the vocabulary's minimum content.
var (
	// ApplicationOutcomeSucceeded marks an application that generated
	// everything it was expected to generate.
	ApplicationOutcomeSucceeded = mustApplicationOutcome("succeeded")

	// ApplicationOutcomeFailed marks an application that generated nothing.
	ApplicationOutcomeFailed = mustApplicationOutcome("failed")

	// ApplicationOutcomePartiallySucceeded marks an application that generated
	// some but not all of its expected outputs. PEOS-009 attaches an extra
	// obligation to exactly this outcome: it "SHALL explicitly identify which
	// outputs were generated and which were not", which ApplicationRecord
	// enforces structurally -- see NewApplicationRecord.
	ApplicationOutcomePartiallySucceeded = mustApplicationOutcome("partially-succeeded")

	// ApplicationOutcomeInterrupted marks an application that was interrupted
	// before concluding.
	ApplicationOutcomeInterrupted = mustApplicationOutcome("interrupted")

	// ApplicationOutcomeIndeterminate marks an application whose result could
	// not be determined.
	ApplicationOutcomeIndeterminate = mustApplicationOutcome("indeterminate")
)

// --- ValueSource ---------------------------------------------------------------

// ValueSource names where a resolved parameter value came from (PEOS-009
// Template Application Record: "the source of each resolved value (explicit
// input, default, or derived)").
//
// Unlike ApplicationOutcome, PEOS-009 states this list without an
// extensibility clause, but it also does not close it, so the type stays an
// open vocabulary wrapper with exactly the three named values predeclared.
type ValueSource struct{ value core.VocabularyValue }

// NewValueSource wraps v as a ValueSource.
func NewValueSource(v core.VocabularyValue) ValueSource { return ValueSource{value: v} }

// Value returns the underlying core.VocabularyValue.
func (s ValueSource) Value() core.VocabularyValue { return s.value }
func (s ValueSource) String() string              { return s.value.String() }
func (s ValueSource) IsZero() bool                { return s.value.IsZero() }

// Equal reports whether s and other carry the same vocabulary value.
func (s ValueSource) Equal(other ValueSource) bool { return s.value.Equal(other.value) }

func (s ValueSource) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }

func (s *ValueSource) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &s.value) }

func mustValueSource(value string) ValueSource {
	v, err := core.NewVocabularyValue(core.PEOSNamespace, value)
	if err != nil {
		panic("template: ValueSource vocabulary value is invalid: " + err.Error())
	}
	return ValueSource{value: v}
}

// The three sources PEOS-009 names.
var (
	// ValueSourceExplicitInput marks a value supplied directly by the caller.
	ValueSourceExplicitInput = mustValueSource("explicit-input")

	// ValueSourceDefault marks a value taken from the Template Revision's own
	// ParameterDefault for that parameter.
	ValueSourceDefault = mustValueSource("default")

	// ValueSourceDerived marks a value computed during application.
	ValueSourceDerived = mustValueSource("derived")
)

// --- ResolvedValue -------------------------------------------------------------

// ResolvedValue is one resolved parameter value on a Template Application
// Record, together with where it came from (PEOS-009: "the resolved parameter
// values; the source of each resolved value").
//
// # This is the only place a parameter value ever lives
//
// A Parameter on a Template Artifact Revision carries no value, no current
// value, and no binding -- it declares a parameter's key, type, and
// requiredness, and nothing about any particular application. Resolution is
// per-application and belongs here, on the immutable record of that one
// application.
//
// There is deliberately no ParameterBinding entity and no BindingRecord:
// PEOS-009 defines neither, and "Storing resolved parameter values solely on
// the Generated-From relation instead of on the Template Application Record"
// is a named non-conforming pattern -- the resolved values belong to the
// record, as plain owned values rather than as a separate identified entity.
//
// The value itself is an opaque trimmed string, for the same reason a
// ParameterDefault's is: a parameter's type is either a controlled vocabulary
// or an externally governed definition this package does not resolve, so there
// is no basis on which to validate a typed value.
//
// ResolvedValue carries no identity, no lifecycle, and no correction reference
// of its own; the owning record carries the correction reference.
type ResolvedValue struct {
	parameter core.LocalKey
	value     string
	source    ValueSource
}

// NewResolvedValue validates its three arguments and returns a ResolvedValue.
// parameter must be non-zero and names a Parameter by its template-local key
// within the applied Template Artifact Revision; value must be non-empty after
// trimming and is stored trimmed; source must be non-zero.
//
// This package does not resolve parameter against the applied Revision: an
// ApplicationRecord references its Template Artifact Revision exactly but does
// not carry that Revision's content, so confirming the key exists requires
// loading the Revision, which is repository-owned (see doc.go).
func NewResolvedValue(parameter core.LocalKey, value string, source ValueSource) (ResolvedValue, error) {
	if parameter.IsZero() {
		return ResolvedValue{}, fmt.Errorf("template: NewResolvedValue: %w: parameter key must not be zero", ErrInvalidResolvedValue)
	}
	trimmed, err := trimmedRequired("NewResolvedValue", "value", value, ErrInvalidResolvedValue)
	if err != nil {
		return ResolvedValue{}, err
	}
	if source.IsZero() {
		return ResolvedValue{}, fmt.Errorf("template: NewResolvedValue: %w: source must not be zero", ErrInvalidResolvedValue)
	}
	return ResolvedValue{parameter: parameter, value: trimmed, source: source}, nil
}

// Parameter returns the template-local key of the Parameter this value
// resolves.
func (v ResolvedValue) Parameter() core.LocalKey { return v.parameter }

// Value returns the resolved value, uninterpreted.
func (v ResolvedValue) Value() string { return v.value }

// Source returns where the value came from.
func (v ResolvedValue) Source() ValueSource { return v.source }

// IsZero reports whether v is the zero value.
func (v ResolvedValue) IsZero() bool {
	return v.parameter.IsZero() && v.value == "" && v.source.IsZero()
}

type resolvedValueJSON struct {
	Parameter core.LocalKey `json:"parameter"`
	Value     string        `json:"value"`
	Source    ValueSource   `json:"source"`
}

// MarshalJSON encodes v as {"parameter":...,"value":...,"source":...}.
func (v ResolvedValue) MarshalJSON() ([]byte, error) {
	if v.IsZero() {
		return nil, fmt.Errorf("template: marshal ResolvedValue: %w", ErrInvalidResolvedValue)
	}
	return json.Marshal(resolvedValueJSON{Parameter: v.parameter, Value: v.value, Source: v.source})
}

// UnmarshalJSON decodes v from its JSON form, applying the same validation as
// NewResolvedValue. The receiver is left untouched unless every check passes.
func (v *ResolvedValue) UnmarshalJSON(data []byte) error {
	var raw resolvedValueJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal ResolvedValue: %w: %w", ErrInvalidResolvedValue, err)
	}
	result, err := NewResolvedValue(raw.Parameter, raw.Value, raw.Source)
	if err != nil {
		return err
	}
	*v = result
	return nil
}

// --- GeneratedOutput -----------------------------------------------------------

// GeneratedOutput names one Artifact a template application generated
// (PEOS-009: "the generated Artifact and exact generated Artifact Revision,
// where generation succeeded").
//
// # References only, never content
//
// GeneratedOutput carries exactly two exact references and nothing else. It
// stores no generated Artifact, no generated Artifact Revision, no output
// payload, and no rendering result. A generated Artifact "is an ordinary PEOS
// Artifact" with "its own Artifact identity, independent of the Template's
// identity"; this package neither owns that identity nor holds a copy of the
// generated content. Representing a generated Artifact as sharing the
// Template's identity, and introducing a "Template Instance" entity, are both
// named non-conforming patterns.
//
// PEOS-009 asks for both the Artifact and the exact Revision, so both are
// mandatory here, and the Revision reference must name the same Artifact as
// the Artifact reference -- otherwise the pair would identify two different
// generated Artifacts and neither exactly.
type GeneratedOutput struct {
	artifact core.GeneratedArtifactRef
	revision core.GeneratedArtifactRevisionRef
}

// NewGeneratedOutput validates artifact and revision and returns a
// GeneratedOutput. Both must be non-zero, and revision.ArtifactID() must equal
// artifact.ArtifactID().
func NewGeneratedOutput(artifact core.GeneratedArtifactRef, revision core.GeneratedArtifactRevisionRef) (GeneratedOutput, error) {
	if artifact.IsZero() {
		return GeneratedOutput{}, fmt.Errorf("template: NewGeneratedOutput: %w: generated artifact reference must not be zero", ErrInvalidGeneratedOutput)
	}
	if revision.IsZero() {
		return GeneratedOutput{}, fmt.Errorf("template: NewGeneratedOutput: %w: generated artifact revision reference must not be zero", ErrInvalidGeneratedOutput)
	}
	if revision.ArtifactID() != artifact.ArtifactID() {
		return GeneratedOutput{}, fmt.Errorf("template: NewGeneratedOutput: %w: the generated revision names artifact %q but its companion artifact reference names %q", ErrInvalidGeneratedOutput, revision.ArtifactID().String(), artifact.ArtifactID().String())
	}
	return GeneratedOutput{artifact: artifact, revision: revision}, nil
}

// Artifact returns the exact generated Artifact reference.
func (o GeneratedOutput) Artifact() core.GeneratedArtifactRef { return o.artifact }

// Revision returns the exact generated Artifact Revision reference.
func (o GeneratedOutput) Revision() core.GeneratedArtifactRevisionRef { return o.revision }

// IsZero reports whether o is the zero value.
func (o GeneratedOutput) IsZero() bool { return o.artifact.IsZero() && o.revision.IsZero() }

type generatedOutputJSON struct {
	Artifact core.GeneratedArtifactRef         `json:"artifact"`
	Revision core.GeneratedArtifactRevisionRef `json:"revision"`
}

// MarshalJSON encodes o as {"artifact":...,"revision":...}. There is no
// "content", "payload", "rendered", or "result" key: this names what was
// generated and never holds it.
func (o GeneratedOutput) MarshalJSON() ([]byte, error) {
	if o.IsZero() {
		return nil, fmt.Errorf("template: marshal GeneratedOutput: %w", ErrInvalidGeneratedOutput)
	}
	return json.Marshal(generatedOutputJSON{Artifact: o.artifact, Revision: o.revision})
}

// UnmarshalJSON decodes o from its JSON form, applying the same validation as
// NewGeneratedOutput. The receiver is left untouched unless every check passes.
func (o *GeneratedOutput) UnmarshalJSON(data []byte) error {
	var raw generatedOutputJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal GeneratedOutput: %w: %w", ErrInvalidGeneratedOutput, err)
	}
	result, err := NewGeneratedOutput(raw.Artifact, raw.Revision)
	if err != nil {
		return err
	}
	*o = result
	return nil
}

// --- ApplicationRecord ---------------------------------------------------------

// ApplicationRecord is a PEOS-009 Template Application Record: the
// authoritative, immutable, independently identifiable record of one
// application of one Template Artifact Revision.
//
// # Not an Artifact, not revisioned, not lifecycle-bearing
//
// "A Template Application Record is not an Artifact. It is not revisioned. It
// is not lifecycle-bearing." So ApplicationRecord composes no core.Artifact
// and no core.ArtifactRevision, has no Artifact Type, no Revision field, and
// no Lifecycle State. It is independently identifiable through
// core.TemplateApplicationRecordID and references itself exactly through
// core.TemplateApplicationRecordRef -- the reference Packet K.1 added to
// peos/core for this purpose.
//
// Creating an "Application Record Revision" instead of a new record, and
// assigning a Lifecycle State to a record, are both named non-conforming
// patterns. Correction produces a new record; see WithCorrection.
//
// # Mandatory versus optional
//
// id, template, actor, appliedAt, environment, provenance, and outcome are
// mandatory constructor arguments and are unreachable through any later With*
// call: PEOS-009 states each as an Application Record SHALL-identify item
// without a qualifier. resolvedValues is also a constructor argument, because
// a record of an application is incomplete without the values that application
// resolved, and because the outcome-conditional generated-output rule below
// needs the whole record validated through one path.
//
// authority is optional ("authority, where required"), as are generated
// outputs ("where generation succeeded"), known limitations, the correction
// reference ("where applicable"), and extension data.
//
// # The outcome-conditional generated-output rule
//
// PEOS-009 attaches two structural obligations to the outcome:
//
//   - "the generated Artifact and exact generated Artifact Revision, where
//     generation succeeded" -- so a succeeded record must name at least one
//     generated output;
//   - "A `partially succeeded` outcome SHALL explicitly identify which outputs
//     were generated and which were not" -- so a partially-succeeded record
//     must name at least one generated output *and* at least one ungenerated
//     one, since identifying only one side does not identify "which were
//     generated and which were not".
//
// A failed record must name no generated output at all -- naming one would
// contradict the outcome. Interrupted and indeterminate records are
// unconstrained in both directions: PEOS-009 says nothing about what either
// generated, and an interrupted application may well have produced some
// outputs before stopping.
//
// The vocabulary is extensible, so a Product-declared outcome this package
// does not recognize is treated as unconstrained rather than rejected -- this
// package will not invent obligations for outcomes PEOS-009 never named.
type ApplicationRecord struct {
	id             core.TemplateApplicationRecordID
	template       core.TemplateArtifactRevisionRef
	actor          core.ActorRef
	appliedAt      core.Timestamp
	environment    core.VocabularyValue
	provenance     core.Provenance
	outcome        ApplicationOutcome
	resolvedValues []ResolvedValue

	authority          core.AuthorityRef
	generatedOutputs   []GeneratedOutput
	ungeneratedOutputs []string
	limitations        []string
	correction         core.RecordCorrectionRef[core.TemplateApplicationRecordRef]
	extension          core.Extension
}

// validateApplicationRecord is the single shared validation path every
// constructor, every modifier, and UnmarshalJSON converge on, so no public
// path can produce an ApplicationRecord another path would reject.
func validateApplicationRecord(caller string, r ApplicationRecord) error {
	if r.id.IsZero() {
		return fmt.Errorf("template: %s: %w: id must not be zero", caller, ErrInvalidApplicationRecord)
	}
	if r.template.IsZero() {
		return fmt.Errorf("template: %s: %w: applied template artifact revision must not be zero", caller, ErrInvalidApplicationRecord)
	}
	if r.actor.IsZero() {
		return fmt.Errorf("template: %s: %w: actor must not be zero", caller, ErrInvalidApplicationRecord)
	}
	if r.appliedAt.IsZero() {
		return fmt.Errorf("template: %s: %w: timestamp must not be zero", caller, ErrInvalidApplicationRecord)
	}
	if r.environment.IsZero() {
		return fmt.Errorf("template: %s: %w: environment must not be zero", caller, ErrInvalidApplicationRecord)
	}
	if r.provenance.IsZero() {
		return fmt.Errorf("template: %s: %w: provenance must not be zero", caller, ErrInvalidApplicationRecord)
	}
	if r.outcome.IsZero() {
		return fmt.Errorf("template: %s: %w: outcome must not be zero", caller, ErrInvalidApplicationRecord)
	}

	// Resolved values: no zero element, and no parameter resolved twice --
	// two values for one parameter would leave the resolution ambiguous.
	seenParameters := make(map[string]bool, len(r.resolvedValues))
	for _, v := range r.resolvedValues {
		if v.IsZero() {
			return fmt.Errorf("template: %s: %w: resolved value must not be zero", caller, ErrInvalidResolvedValue)
		}
		key := v.Parameter().String()
		if seenParameters[key] {
			return fmt.Errorf("template: %s: %w: parameter %q is resolved more than once", caller, ErrInvalidResolvedValue, key)
		}
		seenParameters[key] = true
	}

	// Generated outputs: no zero element, and no generated Artifact named
	// twice.
	seenOutputs := make(map[string]bool, len(r.generatedOutputs))
	for _, o := range r.generatedOutputs {
		if o.IsZero() {
			return fmt.Errorf("template: %s: %w: generated output must not be zero", caller, ErrInvalidGeneratedOutput)
		}
		key := o.Artifact().ArtifactID().String()
		if seenOutputs[key] {
			return fmt.Errorf("template: %s: %w: generated artifact %q is named more than once", caller, ErrInvalidGeneratedOutput, key)
		}
		seenOutputs[key] = true
	}

	if slices.Contains(r.ungeneratedOutputs, "") {
		return fmt.Errorf("template: %s: %w: ungenerated-output description must not be empty", caller, ErrInvalidApplicationRecord)
	}
	if slices.Contains(r.limitations, "") {
		return fmt.Errorf("template: %s: %w: limitation must not be empty", caller, ErrInvalidApplicationRecord)
	}

	if !r.correction.IsZero() && r.correction.Target().RecordID() == r.id {
		return fmt.Errorf("template: %s: %w: an application record must not correct itself", caller, core.ErrInvalidCorrectionReference)
	}

	return validateOutcomeConsistency(caller, r)
}

// validateOutcomeConsistency enforces PEOS-009's two outcome-conditional
// obligations. See ApplicationRecord's type comment for the derivation.
func validateOutcomeConsistency(caller string, r ApplicationRecord) error {
	switch {
	case r.outcome.Equal(ApplicationOutcomeSucceeded):
		if len(r.generatedOutputs) == 0 {
			return fmt.Errorf("template: %s: %w: a succeeded application must identify at least one generated output", caller, ErrInvalidApplicationRecord)
		}
		if len(r.ungeneratedOutputs) != 0 {
			return fmt.Errorf("template: %s: %w: a succeeded application must not identify an ungenerated output", caller, ErrInvalidApplicationRecord)
		}
	case r.outcome.Equal(ApplicationOutcomePartiallySucceeded):
		if len(r.generatedOutputs) == 0 {
			return fmt.Errorf("template: %s: %w: a partially succeeded application must identify which outputs were generated", caller, ErrInvalidApplicationRecord)
		}
		if len(r.ungeneratedOutputs) == 0 {
			return fmt.Errorf("template: %s: %w: a partially succeeded application must identify which outputs were not generated", caller, ErrInvalidApplicationRecord)
		}
	case r.outcome.Equal(ApplicationOutcomeFailed):
		if len(r.generatedOutputs) != 0 {
			return fmt.Errorf("template: %s: %w: a failed application must not identify a generated output", caller, ErrInvalidApplicationRecord)
		}
	}
	// Interrupted, indeterminate, and any Product-declared outcome are
	// deliberately unconstrained: PEOS-009 states no generated-output rule for
	// them, and inventing one would assert an obligation it does not define.
	return nil
}

// NewApplicationRecord validates its arguments and returns an
// ApplicationRecord with no authority, no limitations, no correction
// reference, and no extension data. Use the With* methods to add those.
//
// environment is a namespaced core.VocabularyValue rather than a free string,
// matching how this repository already models a runtime environment; PEOS-009
// requires the record to identify its "environment or context" and defines no
// vocabulary for it, so which environments exist is Product-owned.
//
// resolvedValues may be empty or nil: a parameterless Template resolves
// nothing, and PEOS-009 states no minimum. No parameter may be resolved twice.
//
// # Why the outputs are constructor arguments
//
// generatedOutputs and ungeneratedOutputs are constructor arguments rather
// than modifiers alone, for the same constructor-completeness reason
// quality.NewProfileContent takes normalizationRules and
// template.NewTemplateContent takes parameters, defaults, and constraints: the
// outcome-conditional rule (see the type comment) is a cross-field invariant
// that can only be checked with the outcome *and* both output collections
// present. With outputs reachable only through modifiers, every succeeded and
// partially-succeeded record -- precisely the two outcomes PEOS-009 attaches
// output obligations to -- would be unconstructible, because the constructor
// would reject the record before any modifier could supply what makes it
// valid. Both may be empty or nil for the outcomes that require neither.
//
// Every slice argument is defensively copied; the caller may reuse or mutate
// its own slices afterward without affecting the returned value.
func NewApplicationRecord(
	id core.TemplateApplicationRecordID,
	template core.TemplateArtifactRevisionRef,
	actor core.ActorRef,
	appliedAt core.Timestamp,
	environment core.VocabularyValue,
	provenance core.Provenance,
	outcome ApplicationOutcome,
	resolvedValues []ResolvedValue,
	generatedOutputs []GeneratedOutput,
	ungeneratedOutputs []string,
) (ApplicationRecord, error) {
	trimmedUngenerated, err := trimmedStringSlice("NewApplicationRecord", "ungenerated-output description", ungeneratedOutputs, ErrInvalidApplicationRecord)
	if err != nil {
		return ApplicationRecord{}, err
	}
	r := ApplicationRecord{
		id:                 id,
		template:           template,
		actor:              actor,
		appliedAt:          appliedAt,
		environment:        environment,
		provenance:         provenance,
		outcome:            outcome,
		resolvedValues:     copySlice(resolvedValues),
		generatedOutputs:   copySlice(generatedOutputs),
		ungeneratedOutputs: trimmedUngenerated,
	}
	if err := validateApplicationRecord("NewApplicationRecord", r); err != nil {
		return ApplicationRecord{}, err
	}
	return r, nil
}

// WithAuthority returns a copy of r with its governing authority set.
// authority must be non-zero; use WithoutAuthority to clear it. Authority is
// optional because PEOS-009 writes "authority, where required".
func (r ApplicationRecord) WithAuthority(authority core.AuthorityRef) (ApplicationRecord, error) {
	if authority.IsZero() {
		return ApplicationRecord{}, fmt.Errorf("template: ApplicationRecord.WithAuthority: %w: authority must not be zero", ErrInvalidApplicationRecord)
	}
	r.authority = authority
	if err := validateApplicationRecord("ApplicationRecord.WithAuthority", r); err != nil {
		return ApplicationRecord{}, err
	}
	return r, nil
}

// WithoutAuthority returns a copy of r with its governing authority cleared.
func (r ApplicationRecord) WithoutAuthority() ApplicationRecord {
	r.authority = core.AuthorityRef{}
	return r
}

// WithGeneratedOutputs returns a copy of r naming exactly the generated
// Artifacts and their exact Revisions given, in the order given. A zero-value
// element, or the same generated Artifact named twice, is rejected. Passing an
// empty or nil slice declares none.
//
// The result is revalidated against the outcome-conditional rule, so clearing
// the outputs of a succeeded record is rejected rather than silently accepted.
func (r ApplicationRecord) WithGeneratedOutputs(outputs []GeneratedOutput) (ApplicationRecord, error) {
	r.generatedOutputs = copySlice(outputs)
	if err := validateApplicationRecord("ApplicationRecord.WithGeneratedOutputs", r); err != nil {
		return ApplicationRecord{}, err
	}
	return r, nil
}

// WithUngeneratedOutputs returns a copy of r describing exactly the outputs
// that were expected but not generated, in the order given. Each entry is
// trimmed and must be non-empty after trimming. Passing an empty or nil slice
// declares none.
//
// These are opaque descriptors, not references: an output that was never
// generated has no Artifact identity to reference. This is what satisfies
// PEOS-009's "A `partially succeeded` outcome SHALL explicitly identify which
// outputs were generated and which were not" for the second half of that
// obligation.
func (r ApplicationRecord) WithUngeneratedOutputs(descriptions []string) (ApplicationRecord, error) {
	cp, err := trimmedStringSlice("ApplicationRecord.WithUngeneratedOutputs", "ungenerated-output description", descriptions, ErrInvalidApplicationRecord)
	if err != nil {
		return ApplicationRecord{}, err
	}
	r.ungeneratedOutputs = cp
	if err := validateApplicationRecord("ApplicationRecord.WithUngeneratedOutputs", r); err != nil {
		return ApplicationRecord{}, err
	}
	return r, nil
}

// WithLimitations returns a copy of r with its known-limitations descriptions
// set to exactly the values given, in the order given. Each entry is trimmed
// and must be non-empty after trimming. Passing an empty or nil slice declares
// none, which is why there is no WithoutLimitations.
func (r ApplicationRecord) WithLimitations(limitations []string) (ApplicationRecord, error) {
	cp, err := trimmedStringSlice("ApplicationRecord.WithLimitations", "limitation", limitations, ErrInvalidApplicationRecord)
	if err != nil {
		return ApplicationRecord{}, err
	}
	r.limitations = cp
	return r, nil
}

// WithCorrection returns a copy of r referencing an earlier ApplicationRecord
// it explicitly corrects, replaces, or invalidates. correction must be
// non-zero and must not target r's own identity -- a record cannot correct
// itself. Use WithoutCorrection to clear it.
//
// "Correction of a Template Application Record creates a new Template
// Application Record", and "The earlier record remains historically
// preserved" -- so this reference never mutates or removes the earlier record.
// It carries PEOS-006's correct/replace/invalidate vocabulary
// (core.CorrectionKind) and never PEOS-002 Artifact Supersession: "Record
// replacement SHALL NOT be described using the normative term Supersession."
func (r ApplicationRecord) WithCorrection(correction core.RecordCorrectionRef[core.TemplateApplicationRecordRef]) (ApplicationRecord, error) {
	if correction.IsZero() {
		return ApplicationRecord{}, fmt.Errorf("template: ApplicationRecord.WithCorrection: %w: correction must not be zero", core.ErrInvalidCorrectionReference)
	}
	r.correction = correction
	if err := validateApplicationRecord("ApplicationRecord.WithCorrection", r); err != nil {
		return ApplicationRecord{}, err
	}
	return r, nil
}

// WithoutCorrection returns a copy of r with its correction reference cleared.
func (r ApplicationRecord) WithoutCorrection() ApplicationRecord {
	r.correction = core.RecordCorrectionRef[core.TemplateApplicationRecordRef]{}
	return r
}

// WithExtension returns a copy of r with its extension data set.
func (r ApplicationRecord) WithExtension(extension core.Extension) ApplicationRecord {
	r.extension = extension
	return r
}

// WithoutExtension returns a copy of r with its extension data cleared.
func (r ApplicationRecord) WithoutExtension() ApplicationRecord {
	r.extension = core.Extension{}
	return r
}

// ID returns r's identity.
func (r ApplicationRecord) ID() core.TemplateApplicationRecordID { return r.id }

// Ref returns a core.TemplateApplicationRecordRef identifying r.
func (r ApplicationRecord) Ref() (core.TemplateApplicationRecordRef, error) {
	return core.NewTemplateApplicationRecordRef(r.id)
}

// Template returns the exact Template Artifact Revision r applied.
func (r ApplicationRecord) Template() core.TemplateArtifactRevisionRef { return r.template }

// Actor returns the actor or executing system that performed the application.
func (r ApplicationRecord) Actor() core.ActorRef { return r.actor }

// AppliedAt returns r's timestamp.
func (r ApplicationRecord) AppliedAt() core.Timestamp { return r.appliedAt }

// Environment returns r's declared environment or context.
func (r ApplicationRecord) Environment() core.VocabularyValue { return r.environment }

// Provenance returns r's declared provenance.
func (r ApplicationRecord) Provenance() core.Provenance { return r.provenance }

// Outcome returns how the application concluded. This is a recorded historical
// fact, never mutable status.
func (r ApplicationRecord) Outcome() ApplicationOutcome { return r.outcome }

// ResolvedValues returns a defensive copy of r's resolved parameter values, in
// declaration order. May be empty.
func (r ApplicationRecord) ResolvedValues() []ResolvedValue { return copySlice(r.resolvedValues) }

// ResolvedValue returns the ResolvedValue in r for the parameter named by key,
// and whether one was found. Construction forbids resolving one parameter
// twice, so this can never face an ambiguous match.
func (r ApplicationRecord) ResolvedValue(key core.LocalKey) (ResolvedValue, bool) {
	if key.IsZero() {
		return ResolvedValue{}, false
	}
	for _, v := range r.resolvedValues {
		if v.Parameter() == key {
			return v, true
		}
	}
	return ResolvedValue{}, false
}

// Authority returns r's governing authority, and whether one is set.
func (r ApplicationRecord) Authority() (core.AuthorityRef, bool) {
	return r.authority, !r.authority.IsZero()
}

// GeneratedOutputs returns a defensive copy of the Artifacts r generated, in
// declaration order. May be empty.
func (r ApplicationRecord) GeneratedOutputs() []GeneratedOutput {
	return copySlice(r.generatedOutputs)
}

// UngeneratedOutputs returns a defensive copy of the descriptions of outputs
// that were expected but not generated, in declaration order. May be empty.
func (r ApplicationRecord) UngeneratedOutputs() []string { return copySlice(r.ungeneratedOutputs) }

// Limitations returns a defensive copy of r's known-limitations descriptions,
// in declaration order.
func (r ApplicationRecord) Limitations() []string { return copySlice(r.limitations) }

// Correction returns the earlier record r explicitly corrects, replaces, or
// invalidates, and whether one is set.
func (r ApplicationRecord) Correction() (core.RecordCorrectionRef[core.TemplateApplicationRecordRef], bool) {
	return r.correction, !r.correction.IsZero()
}

// Extension returns r's extension data.
func (r ApplicationRecord) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r ApplicationRecord) IsZero() bool {
	return r.id.IsZero() && r.template.IsZero() && r.actor.IsZero() && r.appliedAt.IsZero() &&
		r.environment.IsZero() && r.provenance.IsZero() && r.outcome.IsZero()
}

type applicationRecordJSON struct {
	ID                 core.TemplateApplicationRecordID                             `json:"id"`
	Template           core.TemplateArtifactRevisionRef                             `json:"template"`
	Actor              core.ActorRef                                                `json:"actor"`
	AppliedAt          core.Timestamp                                               `json:"applied_at"`
	Environment        core.VocabularyValue                                         `json:"environment"`
	Provenance         core.Provenance                                              `json:"provenance"`
	Outcome            ApplicationOutcome                                           `json:"outcome"`
	ResolvedValues     []ResolvedValue                                              `json:"resolved_values,omitempty"`
	Authority          *core.AuthorityRef                                           `json:"authority,omitempty"`
	GeneratedOutputs   []GeneratedOutput                                            `json:"generated_outputs,omitempty"`
	UngeneratedOutputs []string                                                     `json:"ungenerated_outputs,omitempty"`
	Limitations        []string                                                     `json:"limitations,omitempty"`
	Correction         *core.RecordCorrectionRef[core.TemplateApplicationRecordRef] `json:"correction,omitempty"`
	Extension          *core.Extension                                              `json:"extension,omitempty"`
}

// MarshalJSON encodes r flat, with its seven mandatory scalar keys always
// present, plus whichever optional keys are set.
//
// There is no "artifact_type", "revision", "revision_id", "version",
// "lifecycle", "state", "status", "current", "active", "effective",
// "template_instance", "instance", "rendered", "output_content", or "payload"
// key. Their absence is the structural proof that an Application Record is an
// immutable non-Artifact record of one completed application -- never an
// Artifact, never revisioned, never lifecycle-bearing, and never a holder of
// generated content.
func (r ApplicationRecord) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("template: marshal ApplicationRecord: %w", ErrInvalidApplicationRecord)
	}
	raw := applicationRecordJSON{
		ID:                 r.id,
		Template:           r.template,
		Actor:              r.actor,
		AppliedAt:          r.appliedAt,
		Environment:        r.environment,
		Provenance:         r.provenance,
		Outcome:            r.outcome,
		ResolvedValues:     r.resolvedValues,
		GeneratedOutputs:   r.generatedOutputs,
		UngeneratedOutputs: r.ungeneratedOutputs,
		Limitations:        r.limitations,
	}
	if !r.authority.IsZero() {
		raw.Authority = &r.authority
	}
	if !r.correction.IsZero() {
		raw.Correction = &r.correction
	}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

// applicationRecordUnmarshalJSON mirrors applicationRecordJSON for decoding
// only, with Authority and Correction captured as raw bytes so an explicit
// JSON null can be distinguished from an absent key and rejected -- the
// json.RawMessage probe technique Packet D.1 established. The mandatory fields
// need no such treatment: an absent key and an explicit null both yield a zero
// value that validateApplicationRecord rejects. Every optional collection
// needs none either, for the opposite reason: absent, null, and [] all denote
// the same valid state.
type applicationRecordUnmarshalJSON struct {
	ID                 core.TemplateApplicationRecordID `json:"id"`
	Template           core.TemplateArtifactRevisionRef `json:"template"`
	Actor              core.ActorRef                    `json:"actor"`
	AppliedAt          core.Timestamp                   `json:"applied_at"`
	Environment        core.VocabularyValue             `json:"environment"`
	Provenance         core.Provenance                  `json:"provenance"`
	Outcome            ApplicationOutcome               `json:"outcome"`
	ResolvedValues     []ResolvedValue                  `json:"resolved_values"`
	Authority          json.RawMessage                  `json:"authority"`
	GeneratedOutputs   []GeneratedOutput                `json:"generated_outputs"`
	UngeneratedOutputs []string                         `json:"ungenerated_outputs"`
	Limitations        []string                         `json:"limitations"`
	Correction         json.RawMessage                  `json:"correction"`
	Extension          *core.Extension                  `json:"extension,omitempty"`
}

// UnmarshalJSON decodes r from its JSON form, applying the same validation as
// NewApplicationRecord and each With* method -- including the
// outcome-conditional generated-output rule -- so a decoded ApplicationRecord
// can never be constructor-impossible. The receiver is left untouched unless
// every check passes.
func (r *ApplicationRecord) UnmarshalJSON(data []byte) error {
	var raw applicationRecordUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal ApplicationRecord: %w: %w", ErrInvalidApplicationRecord, err)
	}

	result := ApplicationRecord{
		id:                 raw.ID,
		template:           raw.Template,
		actor:              raw.Actor,
		appliedAt:          raw.AppliedAt,
		environment:        raw.Environment,
		provenance:         raw.Provenance,
		outcome:            raw.Outcome,
		resolvedValues:     copySlice(raw.ResolvedValues),
		generatedOutputs:   copySlice(raw.GeneratedOutputs),
		ungeneratedOutputs: copySlice(raw.UngeneratedOutputs),
		limitations:        copySlice(raw.Limitations),
	}

	if len(raw.Authority) > 0 {
		if err := rejectNullRaw("ApplicationRecord", "authority", raw.Authority, ErrInvalidApplicationRecord); err != nil {
			return err
		}
		var authority core.AuthorityRef
		if err := json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("template: unmarshal ApplicationRecord: %w: %w", ErrInvalidApplicationRecord, err)
		}
		if authority.IsZero() {
			return fmt.Errorf("template: unmarshal ApplicationRecord: %w: authority must not be zero", ErrInvalidApplicationRecord)
		}
		result.authority = authority
	}
	if len(raw.Correction) > 0 {
		if err := rejectNullRaw("ApplicationRecord", "correction", raw.Correction, core.ErrInvalidCorrectionReference); err != nil {
			return err
		}
		var correction core.RecordCorrectionRef[core.TemplateApplicationRecordRef]
		if err := json.Unmarshal(raw.Correction, &correction); err != nil {
			return fmt.Errorf("template: unmarshal ApplicationRecord: %w: %w", ErrInvalidApplicationRecord, err)
		}
		result.correction = correction
	}

	// Route both string collections through the same trim-and-reject helper the
	// modifiers use, so a decoded record stores exactly what a constructed one
	// would and a whitespace-only entry is rejected identically.
	var err error
	if result.ungeneratedOutputs, err = trimmedStringSlice("unmarshal ApplicationRecord", "ungenerated-output description", result.ungeneratedOutputs, ErrInvalidApplicationRecord); err != nil {
		return err
	}
	if result.limitations, err = trimmedStringSlice("unmarshal ApplicationRecord", "limitation", result.limitations, ErrInvalidApplicationRecord); err != nil {
		return err
	}

	if err := validateApplicationRecord("unmarshal ApplicationRecord", result); err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}

	*r = result
	return nil
}
