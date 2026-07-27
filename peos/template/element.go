package template

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// This file implements the Template Artifact Revision-owned value structures
// PEOS-009 defines: Parameter, ParameterType, ParameterDefault,
// ParameterConstraint (with its ConstraintTarget union and its two local
// vocabularies), and CompatibilityDeclaration.
//
// Every one of them is Revision-owned content. None has an identity, a Ref, a
// revision system, or a lifecycle of its own: "A Template Parameter has no
// independent parameter lifecycle or revision system of its own", and
// "Granting a Template Parameter an identity that survives independently of
// its owning Template Artifact Revision" is a named non-conforming pattern.
// Changing any of their meanings requires a new Template Artifact Revision --
// "Any change to a parameter's meaning or type requires a new Template
// Artifact Revision" -- which is why every modifier in this file returns a
// copy and none of them can reach a key.

// --- ConstraintEvaluationPoint / ConstraintFailureSemantics --------------------

// ConstraintEvaluationPoint is a namespaced vocabulary value naming when a
// Parameter Constraint is evaluated (PEOS-009 Parameter Constraint: "its
// evaluation point").
//
// PEOS-009 names no evaluation-point vocabulary of its own, so this type
// predeclares none: what evaluation points exist, and what each means, is
// Product-owned. It is a distinct Go type from ConstraintFailureSemantics so
// that one can never be passed where the other is expected.
//
// This is deliberately not core.ExecutionOutcome or any other PEOS-006
// vocabulary. PEOS-006's execution vocabulary describes the outcome of a
// validation execution, which is a different concept that merely shares some
// English words; reusing it would assert a PEOS-006 tie PEOS-009 does not
// state.
type ConstraintEvaluationPoint struct{ value core.VocabularyValue }

// NewConstraintEvaluationPoint wraps v as a ConstraintEvaluationPoint.
func NewConstraintEvaluationPoint(v core.VocabularyValue) ConstraintEvaluationPoint {
	return ConstraintEvaluationPoint{value: v}
}

// Value returns the underlying core.VocabularyValue.
func (p ConstraintEvaluationPoint) Value() core.VocabularyValue { return p.value }
func (p ConstraintEvaluationPoint) String() string              { return p.value.String() }
func (p ConstraintEvaluationPoint) IsZero() bool                { return p.value.IsZero() }

// Equal reports whether p and other carry the same vocabulary value.
func (p ConstraintEvaluationPoint) Equal(other ConstraintEvaluationPoint) bool {
	return p.value.Equal(other.value)
}

func (p ConstraintEvaluationPoint) MarshalJSON() ([]byte, error) { return json.Marshal(p.value) }

func (p *ConstraintEvaluationPoint) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.value)
}

// ConstraintFailureSemantics is a namespaced vocabulary value naming what
// happens when a Parameter Constraint is not satisfied (PEOS-009 Parameter
// Constraint: "its failure semantics").
//
// As with ConstraintEvaluationPoint, PEOS-009 names no vocabulary of its own,
// so none is predeclared here, and this is not any PEOS-006 outcome
// vocabulary. It is a distinct Go type from ConstraintEvaluationPoint.
//
// Declaring failure semantics is not enforcing them: this package never
// evaluates a constraint and never acts on a failure. Both are repository- and
// Product-owned.
type ConstraintFailureSemantics struct{ value core.VocabularyValue }

// NewConstraintFailureSemantics wraps v as a ConstraintFailureSemantics.
func NewConstraintFailureSemantics(v core.VocabularyValue) ConstraintFailureSemantics {
	return ConstraintFailureSemantics{value: v}
}

// Value returns the underlying core.VocabularyValue.
func (s ConstraintFailureSemantics) Value() core.VocabularyValue { return s.value }
func (s ConstraintFailureSemantics) String() string              { return s.value.String() }
func (s ConstraintFailureSemantics) IsZero() bool                { return s.value.IsZero() }

// Equal reports whether s and other carry the same vocabulary value.
func (s ConstraintFailureSemantics) Equal(other ConstraintFailureSemantics) bool {
	return s.value.Equal(other.value)
}

func (s ConstraintFailureSemantics) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }

func (s *ConstraintFailureSemantics) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.value)
}

// --- ParameterType -------------------------------------------------------------

type parameterTypeKind string

const (
	parameterTypeKindVocabulary parameterTypeKind = "vocabulary"
	parameterTypeKindExternal   parameterTypeKind = "external"
)

// ParameterType is a PEOS-009 Parameter Type: a closed two-arm union over
// exactly what the specification permits -- "A Parameter Type is: a controlled
// vocabulary/type; or an exact reference to an externally governed normative
// type definition."
//
// The zero value is invalid and represents a third, unstated state PEOS-009
// does not permit; exactly one arm is always populated on a valid value.
//
// # Why the external arm is an opaque locator, not a core reference
//
// The external arm is a validated opaque external reference (a non-empty
// trimmed locator plus a namespaced authority vocabulary naming the governing
// scheme), not core.ArtifactRevisionRef. PEOS-009 says the type definition is
// "externally governed" and adds that "A Parameter Type has no mandatory
// independent PEOS Artifact identity" -- so assuming every external normative
// type definition is a PEOS Artifact Revision would assert PEOS identity the
// specification explicitly declines to require. Introducing a new core
// reference for it would be worse: it would mint a PEOS identity concept for
// something PEOS-009 places outside PEOS.
//
// A Product whose external types *are* PEOS Artifact Revisions can express
// that through the locator plus its own authority vocabulary; the repository
// layer interprets both. This package never resolves, fetches, or validates
// the referenced definition.
//
// # What this is not
//
// There is no primitive-type enum, no JSON Schema, no Go type name, no
// expression type, and no executable validator. PEOS-009 defines none of them,
// and a parameter's *value* is never stored here in any case -- resolved values
// belong to the Template Application Record (Packet K.2).
type ParameterType struct {
	kind parameterTypeKind

	vocabulary core.VocabularyValue

	externalAuthority core.VocabularyValue
	externalLocator   string
}

// NewVocabularyParameterType validates v and returns a ParameterType naming a
// controlled vocabulary type. v must be non-zero.
func NewVocabularyParameterType(v core.VocabularyValue) (ParameterType, error) {
	if v.IsZero() {
		return ParameterType{}, fmt.Errorf("template: NewVocabularyParameterType: %w: vocabulary value must not be zero", ErrInvalidParameterType)
	}
	return ParameterType{kind: parameterTypeKindVocabulary, vocabulary: v}, nil
}

// NewExternalParameterType validates authority and locator and returns a
// ParameterType naming an exact, externally governed normative type
// definition. authority must be non-zero and names the governing scheme;
// locator must be non-empty after trimming and is stored trimmed. Neither is
// interpreted by this package.
func NewExternalParameterType(authority core.VocabularyValue, locator string) (ParameterType, error) {
	if authority.IsZero() {
		return ParameterType{}, fmt.Errorf("template: NewExternalParameterType: %w: authority must not be zero", ErrInvalidParameterType)
	}
	trimmed, err := trimmedRequired("NewExternalParameterType", "locator", locator, ErrInvalidParameterType)
	if err != nil {
		return ParameterType{}, err
	}
	return ParameterType{kind: parameterTypeKindExternal, externalAuthority: authority, externalLocator: trimmed}, nil
}

// Kind returns t's discriminator, "vocabulary" or "external". The zero value
// returns the empty string.
func (t ParameterType) Kind() string { return string(t.kind) }

// Vocabulary returns t's controlled vocabulary type, and whether t is the
// vocabulary variant.
func (t ParameterType) Vocabulary() (core.VocabularyValue, bool) {
	if t.kind != parameterTypeKindVocabulary {
		return core.VocabularyValue{}, false
	}
	return t.vocabulary, true
}

// External returns t's external governing authority and locator, and whether t
// is the external variant.
func (t ParameterType) External() (authority core.VocabularyValue, locator string, ok bool) {
	if t.kind != parameterTypeKindExternal {
		return core.VocabularyValue{}, "", false
	}
	return t.externalAuthority, t.externalLocator, true
}

// IsZero reports whether t is the zero value -- the unstated state PEOS-009
// does not permit on a valid Parameter.
func (t ParameterType) IsZero() bool { return t.kind == "" }

type parameterTypeJSON struct {
	Kind      string                `json:"kind"`
	Value     *core.VocabularyValue `json:"value,omitempty"`
	Authority *core.VocabularyValue `json:"authority,omitempty"`
	Locator   string                `json:"locator,omitempty"`
}

// MarshalJSON encodes t as {"kind":"vocabulary","value":...} or
// {"kind":"external","authority":...,"locator":...}.
func (t ParameterType) MarshalJSON() ([]byte, error) {
	switch t.kind {
	case parameterTypeKindVocabulary:
		return json.Marshal(parameterTypeJSON{Kind: string(parameterTypeKindVocabulary), Value: &t.vocabulary})
	case parameterTypeKindExternal:
		return json.Marshal(parameterTypeJSON{
			Kind:      string(parameterTypeKindExternal),
			Authority: &t.externalAuthority,
			Locator:   t.externalLocator,
		})
	default:
		return nil, fmt.Errorf("template: marshal ParameterType: %w", ErrInvalidParameterType)
	}
}

// UnmarshalJSON decodes t from its JSON form. A missing or unrecognized kind,
// a vocabulary value carrying external fields, an external value carrying a
// vocabulary value, and a selected arm missing its own payload are all
// rejected. An explicit JSON null for the whole value decodes to an empty kind
// and is rejected the same way. The receiver is left untouched unless every
// check passes.
func (t *ParameterType) UnmarshalJSON(data []byte) error {
	var raw parameterTypeJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal ParameterType: %w: %w", ErrInvalidParameterType, err)
	}
	var result ParameterType
	switch raw.Kind {
	case string(parameterTypeKindVocabulary):
		if raw.Authority != nil || raw.Locator != "" {
			return fmt.Errorf("template: unmarshal ParameterType: %w: a vocabulary type must not carry external fields", ErrInvalidParameterType)
		}
		if raw.Value == nil {
			return fmt.Errorf("template: unmarshal ParameterType: %w: a vocabulary type requires a value", ErrInvalidParameterType)
		}
		var err error
		if result, err = NewVocabularyParameterType(*raw.Value); err != nil {
			return err
		}
	case string(parameterTypeKindExternal):
		if raw.Value != nil {
			return fmt.Errorf("template: unmarshal ParameterType: %w: an external type must not carry a vocabulary value", ErrInvalidParameterType)
		}
		if raw.Authority == nil {
			return fmt.Errorf("template: unmarshal ParameterType: %w: an external type requires an authority", ErrInvalidParameterType)
		}
		var err error
		if result, err = NewExternalParameterType(*raw.Authority, raw.Locator); err != nil {
			return err
		}
	default:
		return fmt.Errorf("template: unmarshal ParameterType: unrecognized kind %q: %w", raw.Kind, ErrInvalidParameterType)
	}
	*t = result
	return nil
}

// --- Parameter -----------------------------------------------------------------

// Parameter is a PEOS-009 Template Parameter: a Template Artifact Revision-owned
// value structure carrying a stable template-local key and a Parameter Type.
//
// # No identity of its own
//
// Parameter has no ID, no Ref, no revision, and no lifecycle. PEOS-009 is
// explicit that its key "is unique only within that exact Template Artifact
// Revision; is not an Artifact Identity; is not a global Template Parameter
// identity", and that "A Template Parameter has no independent parameter
// lifecycle or revision system of its own." Granting one is a named
// non-conforming pattern.
//
// It also carries no provenance, no authority, and no scope: PEOS-009 states
// none of the three for a parameter, and adding them would assert governance
// structure the specification does not define. (A Parameter *Constraint* does
// carry scope and optional authority, because PEOS-009 states both for it.)
//
// # No value of any kind
//
// Parameter carries no value, no current value, no resolved value, and no
// binding. Resolved parameter values belong to the Template Application Record
// ("the resolved parameter values; the source of each resolved value"), which
// is Packet K.2's work; storing them here would make the Revision mutable in
// everything but name.
//
// required is a mandatory constructor argument rather than an optional flag:
// PEOS-009 lists "required parameters" among the items every Template Artifact
// Revision SHALL identify, so required == false is a stated fact about the
// parameter, not an omitted value.
type Parameter struct {
	key           core.LocalKey
	parameterType ParameterType
	required      bool

	description              string
	forbidsDefaultResolution bool
}

// NewParameter validates key, parameterType, and required and returns a
// Parameter with no description and with default resolution permitted. Use the
// With* methods to change those.
//
// key must be non-zero and parameterType must be non-zero. required is
// mandatory in both directions -- see the type comment.
func NewParameter(key core.LocalKey, parameterType ParameterType, required bool) (Parameter, error) {
	if key.IsZero() {
		return Parameter{}, fmt.Errorf("template: NewParameter: %w: key must not be zero", ErrInvalidTemplateParameter)
	}
	if parameterType.IsZero() {
		return Parameter{}, fmt.Errorf("template: NewParameter: %w: parameter type must not be zero", ErrInvalidParameterType)
	}
	return Parameter{key: key, parameterType: parameterType, required: required}, nil
}

// WithDescription returns a copy of p with its human-readable description set.
// description must be non-empty after trimming; the trimmed value is stored.
// Use WithoutDescription to clear it.
func (p Parameter) WithDescription(description string) (Parameter, error) {
	trimmed, err := trimmedRequired("Parameter.WithDescription", "description", description, ErrInvalidTemplateParameter)
	if err != nil {
		return Parameter{}, err
	}
	p.description = trimmed
	return p, nil
}

// WithoutDescription returns a copy of p with its description cleared.
func (p Parameter) WithoutDescription() Parameter {
	p.description = ""
	return p
}

// WithForbiddenDefaultResolution returns a copy of p declaring that a default
// does not satisfy it, per PEOS-009: "A default does not satisfy a required
// parameter where the owning Template Artifact Revision explicitly forbids
// default resolution for that parameter."
//
// A TemplateContent carrying a ParameterDefault that targets such a parameter
// is rejected, so this flag and a default for the same key are mutually
// exclusive by construction.
func (p Parameter) WithForbiddenDefaultResolution() Parameter {
	p.forbidsDefaultResolution = true
	return p
}

// WithPermittedDefaultResolution returns a copy of p declaring that default
// resolution is permitted for it -- the constructor default.
func (p Parameter) WithPermittedDefaultResolution() Parameter {
	p.forbidsDefaultResolution = false
	return p
}

// Key returns p's template-local key. It is meaningful only within the
// parameter namespace of its owning TemplateContent.
func (p Parameter) Key() core.LocalKey { return p.key }

// Type returns p's Parameter Type.
func (p Parameter) Type() ParameterType { return p.parameterType }

// Required reports whether p must be supplied. This is stated state, never an
// absence: a false value means "explicitly optional".
func (p Parameter) Required() bool { return p.required }

// Description returns p's human-readable description, and whether one is set.
func (p Parameter) Description() (string, bool) { return p.description, p.description != "" }

// ForbidsDefaultResolution reports whether the owning Revision forbids a
// default from satisfying p.
func (p Parameter) ForbidsDefaultResolution() bool { return p.forbidsDefaultResolution }

// IsZero reports whether p is the zero value.
func (p Parameter) IsZero() bool { return p.key.IsZero() && p.parameterType.IsZero() }

type parameterJSON struct {
	Key                      core.LocalKey `json:"key"`
	ParameterType            ParameterType `json:"parameter_type"`
	Required                 bool          `json:"required"`
	Description              string        `json:"description,omitempty"`
	ForbidsDefaultResolution bool          `json:"forbids_default_resolution,omitempty"`
}

// MarshalJSON encodes p as {"key":...,"parameter_type":...,"required":...},
// plus whichever optional keys are set.
//
// "required" is always present, even when false: it is mandatory stated state,
// and omitting it would make "explicitly optional" indistinguishable from
// "unstated". There is no "value", "current_value", "resolved_value", or
// "binding" key -- a Parameter never carries a value.
func (p Parameter) MarshalJSON() ([]byte, error) {
	if p.IsZero() {
		return nil, fmt.Errorf("template: marshal Parameter: %w", ErrInvalidTemplateParameter)
	}
	return json.Marshal(parameterJSON{
		Key:                      p.key,
		ParameterType:            p.parameterType,
		Required:                 p.required,
		Description:              p.description,
		ForbidsDefaultResolution: p.forbidsDefaultResolution,
	})
}

// UnmarshalJSON decodes p from its JSON form, applying the same validation as
// NewParameter and each With* method. An absent "required" decodes to false,
// which is a valid stated value. The receiver is left untouched unless every
// check passes.
func (p *Parameter) UnmarshalJSON(data []byte) error {
	var raw parameterJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal Parameter: %w: %w", ErrInvalidTemplateParameter, err)
	}
	result, err := NewParameter(raw.Key, raw.ParameterType, raw.Required)
	if err != nil {
		return err
	}
	if raw.Description != "" {
		if result, err = result.WithDescription(raw.Description); err != nil {
			return err
		}
	}
	if raw.ForbidsDefaultResolution {
		result = result.WithForbiddenDefaultResolution()
	}
	*p = result
	return nil
}

// --- ParameterDefault ----------------------------------------------------------

// ParameterDefault is a PEOS-009 Parameter Default: a Template Artifact
// Revision-owned value structure naming the parameter it defaults and the
// default value itself.
//
// The default value is an opaque, trimmed string, not arbitrary JSON. PEOS-009
// says nothing about a default's representation beyond that it is a
// Revision-owned value structure, and a parameter's type is itself either a
// controlled vocabulary or an externally governed definition this package does
// not resolve -- so this package has no basis on which to validate a typed
// default and no warrant to accept an arbitrary structure. A Product needing
// structure carries it in the string under its own contract, exactly as
// Assertion evaluation rules and ContractRule text do in peos/runtime.
//
// ParameterDefault carries no identity and no key of its own: it is identified
// by the parameter it targets, and TemplateContent rejects a second default
// for one parameter. It also carries no resolved value -- a default is a
// declared fallback, not a resolution; resolution is recorded on the Template
// Application Record (Packet K.2).
//
// Whether the named parameter exists, and whether it forbids default
// resolution, are checked by TemplateContent, which owns the aggregate view
// those checks require.
type ParameterDefault struct {
	parameter core.LocalKey
	value     string
}

// NewParameterDefault validates parameter and value and returns a
// ParameterDefault. parameter must be non-zero and names a Parameter by its
// template-local key; value must be non-empty after trimming and is stored
// trimmed.
func NewParameterDefault(parameter core.LocalKey, value string) (ParameterDefault, error) {
	if parameter.IsZero() {
		return ParameterDefault{}, fmt.Errorf("template: NewParameterDefault: %w: parameter key must not be zero", ErrInvalidParameterDefault)
	}
	trimmed, err := trimmedRequired("NewParameterDefault", "value", value, ErrInvalidParameterDefault)
	if err != nil {
		return ParameterDefault{}, err
	}
	return ParameterDefault{parameter: parameter, value: trimmed}, nil
}

// Parameter returns the template-local key of the Parameter d defaults.
func (d ParameterDefault) Parameter() core.LocalKey { return d.parameter }

// Value returns d's declared default value, uninterpreted.
func (d ParameterDefault) Value() string { return d.value }

// IsZero reports whether d is the zero value.
func (d ParameterDefault) IsZero() bool { return d.parameter.IsZero() && d.value == "" }

type parameterDefaultJSON struct {
	Parameter core.LocalKey `json:"parameter"`
	Value     string        `json:"value"`
}

// MarshalJSON encodes d as {"parameter":...,"value":...}. There is no
// "resolved", "applied", or "effective" key: a default is declared, never
// resolved, here.
func (d ParameterDefault) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("template: marshal ParameterDefault: %w", ErrInvalidParameterDefault)
	}
	return json.Marshal(parameterDefaultJSON{Parameter: d.parameter, Value: d.value})
}

// UnmarshalJSON decodes d from its JSON form, applying the same validation as
// NewParameterDefault. The receiver is left untouched unless every check
// passes.
func (d *ParameterDefault) UnmarshalJSON(data []byte) error {
	var raw parameterDefaultJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal ParameterDefault: %w: %w", ErrInvalidParameterDefault, err)
	}
	result, err := NewParameterDefault(raw.Parameter, raw.Value)
	if err != nil {
		return err
	}
	*d = result
	return nil
}

// --- ConstraintTarget ----------------------------------------------------------

type constraintTargetKind string

const (
	constraintTargetKindParameter        constraintTargetKind = "parameter"
	constraintTargetKindGeneratedContent constraintTargetKind = "generated_content"
)

// ConstraintTarget is what a Parameter Constraint affects: a closed two-arm
// union over exactly what PEOS-009 permits -- every Parameter Constraint SHALL
// identify "the affected parameter or generated content".
//
// The parameter arm names a Parameter by its template-local key, and
// TemplateContent confirms that key resolves. The generated-content arm is a
// non-empty trimmed descriptor of the generated content the constraint
// affects, and deliberately carries **no** generated Artifact identity: a
// generated Artifact "has its own Artifact identity, independent of the
// Template's identity", and it does not exist at declaration time, so it has
// nothing this Revision could name exactly. The descriptor is Product-owned
// and never resolved by this package.
//
// The generated-content arm is therefore not a GeneratedArtifactRef, not a
// GeneratedArtifactRevisionRef, not a current output, not an output store, and
// not an execution result -- all of which would either invent identity for
// something unborn or store generated state on the Revision.
//
// The zero value is invalid; exactly one arm is populated on a valid value.
type ConstraintTarget struct {
	kind             constraintTargetKind
	parameter        core.LocalKey
	generatedContent string
}

// NewParameterConstraintTarget validates parameter and returns a
// ConstraintTarget naming a Parameter by its template-local key. parameter
// must be non-zero.
func NewParameterConstraintTarget(parameter core.LocalKey) (ConstraintTarget, error) {
	if parameter.IsZero() {
		return ConstraintTarget{}, fmt.Errorf("template: NewParameterConstraintTarget: %w: parameter key must not be zero", ErrInvalidConstraintTarget)
	}
	return ConstraintTarget{kind: constraintTargetKindParameter, parameter: parameter}, nil
}

// NewGeneratedContentConstraintTarget validates descriptor and returns a
// ConstraintTarget naming affected generated content. descriptor must be
// non-empty after trimming and is stored trimmed; it is never interpreted.
func NewGeneratedContentConstraintTarget(descriptor string) (ConstraintTarget, error) {
	trimmed, err := trimmedRequired("NewGeneratedContentConstraintTarget", "descriptor", descriptor, ErrInvalidConstraintTarget)
	if err != nil {
		return ConstraintTarget{}, err
	}
	return ConstraintTarget{kind: constraintTargetKindGeneratedContent, generatedContent: trimmed}, nil
}

// Kind returns t's discriminator, "parameter" or "generated_content". The zero
// value returns the empty string.
func (t ConstraintTarget) Kind() string { return string(t.kind) }

// Parameter returns the template-local key of the Parameter t affects, and
// whether t is the parameter variant.
func (t ConstraintTarget) Parameter() (core.LocalKey, bool) {
	if t.kind != constraintTargetKindParameter {
		return core.LocalKey{}, false
	}
	return t.parameter, true
}

// GeneratedContent returns t's generated-content descriptor, and whether t is
// the generated-content variant.
func (t ConstraintTarget) GeneratedContent() (string, bool) {
	if t.kind != constraintTargetKindGeneratedContent {
		return "", false
	}
	return t.generatedContent, true
}

// IsZero reports whether t is the zero value.
func (t ConstraintTarget) IsZero() bool { return t.kind == "" }

type constraintTargetJSON struct {
	Kind             string         `json:"kind"`
	Parameter        *core.LocalKey `json:"parameter,omitempty"`
	GeneratedContent string         `json:"generated_content,omitempty"`
}

// MarshalJSON encodes t as {"kind":"parameter","parameter":...} or
// {"kind":"generated_content","generated_content":...}.
func (t ConstraintTarget) MarshalJSON() ([]byte, error) {
	switch t.kind {
	case constraintTargetKindParameter:
		return json.Marshal(constraintTargetJSON{Kind: string(constraintTargetKindParameter), Parameter: &t.parameter})
	case constraintTargetKindGeneratedContent:
		return json.Marshal(constraintTargetJSON{Kind: string(constraintTargetKindGeneratedContent), GeneratedContent: t.generatedContent})
	default:
		return nil, fmt.Errorf("template: marshal ConstraintTarget: %w", ErrInvalidConstraintTarget)
	}
}

// UnmarshalJSON decodes t from its JSON form. A missing or unrecognized kind,
// an arm carrying the other arm's payload, and a selected arm missing its own
// payload are all rejected. The receiver is left untouched unless every check
// passes.
func (t *ConstraintTarget) UnmarshalJSON(data []byte) error {
	var raw constraintTargetJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal ConstraintTarget: %w: %w", ErrInvalidConstraintTarget, err)
	}
	var result ConstraintTarget
	switch raw.Kind {
	case string(constraintTargetKindParameter):
		if raw.GeneratedContent != "" {
			return fmt.Errorf("template: unmarshal ConstraintTarget: %w: a parameter target must not carry generated content", ErrInvalidConstraintTarget)
		}
		if raw.Parameter == nil {
			return fmt.Errorf("template: unmarshal ConstraintTarget: %w: a parameter target requires a parameter key", ErrInvalidConstraintTarget)
		}
		var err error
		if result, err = NewParameterConstraintTarget(*raw.Parameter); err != nil {
			return err
		}
	case string(constraintTargetKindGeneratedContent):
		if raw.Parameter != nil {
			return fmt.Errorf("template: unmarshal ConstraintTarget: %w: a generated-content target must not carry a parameter key", ErrInvalidConstraintTarget)
		}
		var err error
		if result, err = NewGeneratedContentConstraintTarget(raw.GeneratedContent); err != nil {
			return err
		}
	default:
		return fmt.Errorf("template: unmarshal ConstraintTarget: unrecognized kind %q: %w", raw.Kind, ErrInvalidConstraintTarget)
	}
	*t = result
	return nil
}

// --- ParameterConstraint -------------------------------------------------------

// ParameterConstraint is a PEOS-009 Parameter Constraint: a Template Artifact
// Revision-owned value structure identifying "the affected parameter or
// generated content; the rule; its scope; its evaluation point; its failure
// semantics; authority, where required."
//
// # Why it carries a template-local key
//
// PEOS-009 states a local-key requirement for a Template Parameter but not,
// explicitly, for a Parameter Constraint. The key here is nonetheless
// mandatory, as a derived structural requirement:
// core.TemplateConstraintCriterionRef is a (Template Artifact Revision,
// LocalKey) pair whose documented purpose is to reference "a Template
// Parameter Constraint, or another Template-owned constraint... by naming its
// owning Template Artifact Revision and its local key within that Revision",
// and core.CriterionKindTemplateConstraint has no other resolution target. An
// unkeyed constraint would leave that criterion kind unresolvable -- exactly
// the defect Packet J.3 raised as J3-03 against peos/runtime's originally
// unkeyed Contract Rule collections. Keying constraints from the start avoids
// repeating it.
//
// The constraint namespace is separate from the parameter namespace, so one
// key may be used once by a Parameter and once by a ParameterConstraint.
//
// # Scope, authority, and the rule
//
// scope is mandatory: PEOS-009 lists "its scope" unqualified. authority is
// optional: PEOS-009 writes "authority, where required". The rule is an
// opaque, trimmed string -- PEOS-009 defines no constraint grammar and no
// expression language, and its Non-Goals disclaim a templating engine, so this
// package stores the rule and never parses, compiles, or evaluates it.
//
// ParameterConstraint has no identity, no revision, no lifecycle, and records
// no evaluation result: it declares a constraint and never records whether it
// held.
type ParameterConstraint struct {
	key              core.LocalKey
	target           ConstraintTarget
	rule             string
	scope            core.Scope
	evaluationPoint  ConstraintEvaluationPoint
	failureSemantics ConstraintFailureSemantics

	authority core.AuthorityRef
}

// NewParameterConstraint validates its six mandatory arguments and returns a
// ParameterConstraint with no authority. Use WithAuthority to add one.
//
// key, target, scope, evaluationPoint, and failureSemantics must all be
// non-zero; rule must be non-empty after trimming and is stored trimmed. This
// package does not resolve target: whether a parameter-targeting constraint's
// key names a declared Parameter is checked by TemplateContent, which owns the
// aggregate view that check requires.
func NewParameterConstraint(
	key core.LocalKey,
	target ConstraintTarget,
	rule string,
	scope core.Scope,
	evaluationPoint ConstraintEvaluationPoint,
	failureSemantics ConstraintFailureSemantics,
) (ParameterConstraint, error) {
	if key.IsZero() {
		return ParameterConstraint{}, fmt.Errorf("template: NewParameterConstraint: %w: key must not be zero", ErrInvalidParameterConstraint)
	}
	if target.IsZero() {
		return ParameterConstraint{}, fmt.Errorf("template: NewParameterConstraint: %w: target must not be zero", ErrInvalidConstraintTarget)
	}
	trimmedRule, err := trimmedRequired("NewParameterConstraint", "rule", rule, ErrInvalidParameterConstraint)
	if err != nil {
		return ParameterConstraint{}, err
	}
	if scope.IsZero() {
		return ParameterConstraint{}, fmt.Errorf("template: NewParameterConstraint: %w: scope must not be zero", core.ErrInvalidScope)
	}
	if evaluationPoint.IsZero() {
		return ParameterConstraint{}, fmt.Errorf("template: NewParameterConstraint: %w: evaluation point must not be zero", ErrInvalidParameterConstraint)
	}
	if failureSemantics.IsZero() {
		return ParameterConstraint{}, fmt.Errorf("template: NewParameterConstraint: %w: failure semantics must not be zero", ErrInvalidParameterConstraint)
	}
	return ParameterConstraint{
		key:              key,
		target:           target,
		rule:             trimmedRule,
		scope:            scope,
		evaluationPoint:  evaluationPoint,
		failureSemantics: failureSemantics,
	}, nil
}

// WithAuthority returns a copy of v with its governing authority set.
// authority must be non-zero; use WithoutAuthority to clear it.
func (v ParameterConstraint) WithAuthority(authority core.AuthorityRef) (ParameterConstraint, error) {
	if authority.IsZero() {
		return ParameterConstraint{}, fmt.Errorf("template: ParameterConstraint.WithAuthority: %w: authority must not be zero", ErrInvalidParameterConstraint)
	}
	v.authority = authority
	return v, nil
}

// WithoutAuthority returns a copy of v with its governing authority cleared.
func (v ParameterConstraint) WithoutAuthority() ParameterConstraint {
	v.authority = core.AuthorityRef{}
	return v
}

// Key returns v's template-local key. It is meaningful only within the
// constraint namespace of its owning TemplateContent, and is what
// core.TemplateConstraintCriterionRef resolves against.
func (v ParameterConstraint) Key() core.LocalKey { return v.key }

// Target returns what v affects.
func (v ParameterConstraint) Target() ConstraintTarget { return v.target }

// Rule returns v's declarative rule, uninterpreted.
func (v ParameterConstraint) Rule() string { return v.rule }

// Scope returns v's declared scope.
func (v ParameterConstraint) Scope() core.Scope { return v.scope }

// EvaluationPoint returns when v is evaluated.
func (v ParameterConstraint) EvaluationPoint() ConstraintEvaluationPoint { return v.evaluationPoint }

// FailureSemantics returns what happens when v is not satisfied.
func (v ParameterConstraint) FailureSemantics() ConstraintFailureSemantics {
	return v.failureSemantics
}

// Authority returns v's governing authority, and whether one is set.
func (v ParameterConstraint) Authority() (core.AuthorityRef, bool) {
	return v.authority, !v.authority.IsZero()
}

// IsZero reports whether v is the zero value.
func (v ParameterConstraint) IsZero() bool {
	return v.key.IsZero() && v.target.IsZero() && v.rule == "" && v.scope.IsZero() &&
		v.evaluationPoint.IsZero() && v.failureSemantics.IsZero()
}

type parameterConstraintJSON struct {
	Key              core.LocalKey              `json:"key"`
	Target           ConstraintTarget           `json:"target"`
	Rule             string                     `json:"rule"`
	Scope            core.Scope                 `json:"scope"`
	EvaluationPoint  ConstraintEvaluationPoint  `json:"evaluation_point"`
	FailureSemantics ConstraintFailureSemantics `json:"failure_semantics"`
	Authority        *core.AuthorityRef         `json:"authority,omitempty"`
}

// MarshalJSON encodes v with its six mandatory keys always present, plus
// authority when set. There is no "result", "outcome", "satisfied",
// "violated", or "evaluated" key: a constraint declares a rule and never
// records whether it held.
func (v ParameterConstraint) MarshalJSON() ([]byte, error) {
	if v.IsZero() {
		return nil, fmt.Errorf("template: marshal ParameterConstraint: %w", ErrInvalidParameterConstraint)
	}
	raw := parameterConstraintJSON{
		Key:              v.key,
		Target:           v.target,
		Rule:             v.rule,
		Scope:            v.scope,
		EvaluationPoint:  v.evaluationPoint,
		FailureSemantics: v.failureSemantics,
	}
	if !v.authority.IsZero() {
		raw.Authority = &v.authority
	}
	return json.Marshal(raw)
}

type parameterConstraintUnmarshalJSON struct {
	Key              core.LocalKey              `json:"key"`
	Target           ConstraintTarget           `json:"target"`
	Rule             string                     `json:"rule"`
	Scope            core.Scope                 `json:"scope"`
	EvaluationPoint  ConstraintEvaluationPoint  `json:"evaluation_point"`
	FailureSemantics ConstraintFailureSemantics `json:"failure_semantics"`
	Authority        json.RawMessage            `json:"authority"`
}

// UnmarshalJSON decodes v from its JSON form, applying the same validation as
// NewParameterConstraint and WithAuthority. An absent authority means "not
// declared" and is valid; an explicit null is rejected. The receiver is left
// untouched unless every check passes.
func (v *ParameterConstraint) UnmarshalJSON(data []byte) error {
	var raw parameterConstraintUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal ParameterConstraint: %w: %w", ErrInvalidParameterConstraint, err)
	}
	result, err := NewParameterConstraint(raw.Key, raw.Target, raw.Rule, raw.Scope, raw.EvaluationPoint, raw.FailureSemantics)
	if err != nil {
		return err
	}
	if len(raw.Authority) > 0 {
		if err = rejectNullRaw("ParameterConstraint", "authority", raw.Authority, ErrInvalidParameterConstraint); err != nil {
			return err
		}
		var authority core.AuthorityRef
		if err = json.Unmarshal(raw.Authority, &authority); err != nil {
			return fmt.Errorf("template: unmarshal ParameterConstraint: %w: %w", ErrInvalidParameterConstraint, err)
		}
		if result, err = result.WithAuthority(authority); err != nil {
			return err
		}
	}
	*v = result
	return nil
}

// --- CompatibilityDeclaration --------------------------------------------------

// CompatibilityDeclaration is a PEOS-009 compatibility declaration. PEOS-009
// requires it to be "scoped to: exact Template Artifact Revisions; the
// consumer or generated Artifact Type; applicable constraints; the parameter
// contract; migration requirements, where applicable; the applicable Product
// contract."
//
// # A declaration, never a verdict
//
// This type holds only declared inputs. "Current compatibility is a derived
// interpretation, computed from the applicable compatibility declarations at
// query time", and Template.compatible / TemplateRevision.compatible are named
// non-conforming patterns. There is therefore no Compatible() method, no
// boolean, no status, no CurrentCompatibility, no EffectiveCompatibility, and
// no repository lookup anywhere in this package.
//
// # Field mapping
//
// applicableRevisions maps "exact Template Artifact Revisions" -- the exact
// Revisions this declaration speaks about; empty means the declaration is
// scoped to its own owning Revision alone, which is the common case and the
// only one statable at declaration time without naming another Revision.
// applicableArtifactTypes maps "the consumer or generated Artifact Type".
// parameterContract maps "the parameter contract" as an opaque trimmed
// descriptor. productContract maps "the applicable Product contract", also an
// opaque trimmed descriptor: a Product contract has no PEOS identity type, and
// inventing a core reference for it would mint PEOS identity for a
// Product-owned concept. migrationRequirements maps "migration requirements,
// where applicable" -- optional, and an opaque descriptor rather than a typed
// Migration, because PEOS-009 assigns Migration no ontology at all (see
// doc.go's deferral note). "Applicable constraints" are the owning Revision's
// own ParameterConstraints, already keyed and resolvable there, so this type
// does not duplicate them.
type CompatibilityDeclaration struct {
	applicableArtifactTypes []core.ArtifactType
	parameterContract       string
	productContract         string

	applicableRevisions   []core.TemplateArtifactRevisionRef
	migrationRequirements string
}

// NewCompatibilityDeclaration validates its arguments and returns a
// CompatibilityDeclaration with no applicable-Revision references and no
// migration requirements. Use the With* methods to add those.
//
// applicableArtifactTypes must contain at least one non-zero, non-repeated
// core.ArtifactType -- PEOS-009 requires the declaration to be scoped to "the
// consumer or generated Artifact Type", which none cannot satisfy.
// parameterContract and productContract must each be non-empty after
// trimming; the trimmed values are stored and neither is interpreted.
func NewCompatibilityDeclaration(
	applicableArtifactTypes []core.ArtifactType,
	parameterContract string,
	productContract string,
) (CompatibilityDeclaration, error) {
	if len(applicableArtifactTypes) == 0 {
		return CompatibilityDeclaration{}, fmt.Errorf("template: NewCompatibilityDeclaration: %w: at least one applicable artifact type is required", ErrInvalidCompatibilityDeclaration)
	}
	seen := make(map[string]bool, len(applicableArtifactTypes))
	for _, at := range applicableArtifactTypes {
		if at.IsZero() {
			return CompatibilityDeclaration{}, fmt.Errorf("template: NewCompatibilityDeclaration: %w: applicable artifact type must not be zero", ErrInvalidCompatibilityDeclaration)
		}
		s := at.String()
		if seen[s] {
			return CompatibilityDeclaration{}, fmt.Errorf("template: NewCompatibilityDeclaration: %w: applicable artifact type %q is declared more than once", ErrInvalidCompatibilityDeclaration, s)
		}
		seen[s] = true
	}
	trimmedParameter, err := trimmedRequired("NewCompatibilityDeclaration", "parameter contract", parameterContract, ErrInvalidCompatibilityDeclaration)
	if err != nil {
		return CompatibilityDeclaration{}, err
	}
	trimmedProduct, err := trimmedRequired("NewCompatibilityDeclaration", "product contract", productContract, ErrInvalidCompatibilityDeclaration)
	if err != nil {
		return CompatibilityDeclaration{}, err
	}
	return CompatibilityDeclaration{
		applicableArtifactTypes: copySlice(applicableArtifactTypes),
		parameterContract:       trimmedParameter,
		productContract:         trimmedProduct,
	}, nil
}

// WithApplicableRevisions returns a copy of d scoped to exactly the exact
// Template Artifact Revisions given, in the order given. A zero-value or
// repeated element is rejected. Passing an empty or nil slice scopes the
// declaration to its own owning Revision alone.
func (d CompatibilityDeclaration) WithApplicableRevisions(refs []core.TemplateArtifactRevisionRef) (CompatibilityDeclaration, error) {
	if err := validateRevisionRefs("CompatibilityDeclaration.WithApplicableRevisions", "applicable revision", refs); err != nil {
		return CompatibilityDeclaration{}, err
	}
	d.applicableRevisions = copySlice(refs)
	return d, nil
}

// WithMigrationRequirements returns a copy of d with its migration
// requirements set. requirements must be non-empty after trimming; the trimmed
// value is stored. Use WithoutMigrationRequirements to clear it.
//
// This is an opaque descriptor, not a typed Migration: PEOS-009 states nine
// things every migration SHALL identify but never says what a Migration *is*
// ontologically, so this package declines to invent one (see doc.go).
func (d CompatibilityDeclaration) WithMigrationRequirements(requirements string) (CompatibilityDeclaration, error) {
	trimmed, err := trimmedRequired("CompatibilityDeclaration.WithMigrationRequirements", "migration requirements", requirements, ErrInvalidCompatibilityDeclaration)
	if err != nil {
		return CompatibilityDeclaration{}, err
	}
	d.migrationRequirements = trimmed
	return d, nil
}

// WithoutMigrationRequirements returns a copy of d with its migration
// requirements cleared.
func (d CompatibilityDeclaration) WithoutMigrationRequirements() CompatibilityDeclaration {
	d.migrationRequirements = ""
	return d
}

// ApplicableArtifactTypes returns a defensive copy of the consumer or
// generated Artifact Types d is scoped to, in declaration order. Always
// non-empty on a valid declaration.
func (d CompatibilityDeclaration) ApplicableArtifactTypes() []core.ArtifactType {
	return copySlice(d.applicableArtifactTypes)
}

// ParameterContract returns d's declared parameter contract, uninterpreted.
func (d CompatibilityDeclaration) ParameterContract() string { return d.parameterContract }

// ProductContract returns d's applicable Product contract, uninterpreted.
func (d CompatibilityDeclaration) ProductContract() string { return d.productContract }

// ApplicableRevisions returns a defensive copy of the exact Template Artifact
// Revisions d is scoped to, in declaration order. May be empty.
func (d CompatibilityDeclaration) ApplicableRevisions() []core.TemplateArtifactRevisionRef {
	return copySlice(d.applicableRevisions)
}

// MigrationRequirements returns d's migration requirements, and whether any
// are set.
func (d CompatibilityDeclaration) MigrationRequirements() (string, bool) {
	return d.migrationRequirements, d.migrationRequirements != ""
}

// IsZero reports whether d is the zero value.
func (d CompatibilityDeclaration) IsZero() bool {
	return len(d.applicableArtifactTypes) == 0 && d.parameterContract == "" && d.productContract == ""
}

type compatibilityDeclarationJSON struct {
	ApplicableArtifactTypes []core.ArtifactType                `json:"applicable_artifact_types"`
	ParameterContract       string                             `json:"parameter_contract"`
	ProductContract         string                             `json:"product_contract"`
	ApplicableRevisions     []core.TemplateArtifactRevisionRef `json:"applicable_revisions,omitempty"`
	MigrationRequirements   string                             `json:"migration_requirements,omitempty"`
}

// MarshalJSON encodes d with its three mandatory keys always present, plus
// whichever optional keys are set. There is no "compatible", "compatibility",
// "current", "effective", or "status" key -- this is a declaration, and
// current compatibility is derived at query time.
func (d CompatibilityDeclaration) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, fmt.Errorf("template: marshal CompatibilityDeclaration: %w", ErrInvalidCompatibilityDeclaration)
	}
	return json.Marshal(compatibilityDeclarationJSON{
		ApplicableArtifactTypes: d.applicableArtifactTypes,
		ParameterContract:       d.parameterContract,
		ProductContract:         d.productContract,
		ApplicableRevisions:     d.applicableRevisions,
		MigrationRequirements:   d.migrationRequirements,
	})
}

// UnmarshalJSON decodes d from its JSON form, applying the same validation as
// NewCompatibilityDeclaration and each With* method. For
// applicable_revisions, absent, explicit null, and [] are all equivalent. The
// receiver is left untouched unless every check passes.
func (d *CompatibilityDeclaration) UnmarshalJSON(data []byte) error {
	var raw compatibilityDeclarationJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("template: unmarshal CompatibilityDeclaration: %w: %w", ErrInvalidCompatibilityDeclaration, err)
	}
	result, err := NewCompatibilityDeclaration(raw.ApplicableArtifactTypes, raw.ParameterContract, raw.ProductContract)
	if err != nil {
		return err
	}
	if len(raw.ApplicableRevisions) > 0 {
		if result, err = result.WithApplicableRevisions(raw.ApplicableRevisions); err != nil {
			return err
		}
	}
	if raw.MigrationRequirements != "" {
		if result, err = result.WithMigrationRequirements(raw.MigrationRequirements); err != nil {
			return err
		}
	}
	*d = result
	return nil
}
