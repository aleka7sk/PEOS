package requirement

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- Applicability -----------------------------------------------------------

type applicabilityKind string

const (
	applicabilityKindUnrestricted applicabilityKind = "unrestricted"
	applicabilityKindScoped       applicabilityKind = "scoped"
)

// Applicability defines the conditions under which a Requirement's
// required engineering intent applies (PEOS-005 §11). It is required on
// every Content and unambiguously distinguishes "unrestricted" (no
// condition) from "explicit scoped condition" — it deliberately does not
// model "applies nowhere": PEOS-005 §11 frames Applicability as
// conditions under which intent applies, never as an enumerable target
// set that could be empty.
//
// This package does not interpret core.Scope's Expression in any way,
// consistent with core.Scope's own documented philosophy: "Applicability
// SHALL be objectively determinable... SHALL NOT depend upon subjective
// interpretation" (§11) is a requirement on the applicable Product
// contract's evaluation of Expression, not on this package.
type Applicability struct {
	kind  applicabilityKind
	scope core.Scope
}

// NewUnrestrictedApplicability returns an Applicability with no
// restricting condition.
func NewUnrestrictedApplicability() Applicability {
	return Applicability{kind: applicabilityKindUnrestricted}
}

// NewApplicabilityFromScope validates scope and returns an Applicability
// bound to an explicit condition expression.
func NewApplicabilityFromScope(scope core.Scope) (Applicability, error) {
	if scope.IsZero() {
		return Applicability{}, fmt.Errorf("requirement: NewApplicabilityFromScope: %w: scope must not be zero", ErrInvalidApplicability)
	}
	return Applicability{kind: applicabilityKindScoped, scope: scope}, nil
}

// IsZero reports whether a is the zero value.
func (a Applicability) IsZero() bool { return a.kind == "" }

// IsUnrestricted reports whether a represents an unrestricted
// Applicability.
func (a Applicability) IsUnrestricted() bool { return a.kind == applicabilityKindUnrestricted }

// Scope returns a's condition expression, and whether one is set (that
// is, whether a is the scoped variant).
func (a Applicability) Scope() (core.Scope, bool) {
	if a.kind != applicabilityKindScoped {
		return core.Scope{}, false
	}
	return a.scope, true
}

type applicabilityJSON struct {
	Kind  string      `json:"kind"`
	Scope *core.Scope `json:"scope,omitempty"`
}

// MarshalJSON encodes a as {"kind": "unrestricted"} or {"kind": "scoped",
// "scope": {...}}.
func (a Applicability) MarshalJSON() ([]byte, error) {
	switch a.kind {
	case applicabilityKindUnrestricted:
		return json.Marshal(applicabilityJSON{Kind: string(applicabilityKindUnrestricted)})
	case applicabilityKindScoped:
		return json.Marshal(applicabilityJSON{Kind: string(applicabilityKindScoped), Scope: &a.scope})
	default:
		return nil, fmt.Errorf("requirement: marshal Applicability: %w", ErrInvalidApplicability)
	}
}

// UnmarshalJSON decodes a from its JSON form. An unrecognized or missing
// kind, an unrestricted value carrying a scope, or a scoped value missing
// a scope are all rejected.
func (a *Applicability) UnmarshalJSON(data []byte) error {
	var raw applicabilityJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Applicability: %w", err)
	}
	var result Applicability
	switch raw.Kind {
	case string(applicabilityKindUnrestricted):
		if raw.Scope != nil {
			return fmt.Errorf("requirement: unmarshal Applicability: %w: unrestricted must not carry a scope", ErrInvalidApplicability)
		}
		result = NewUnrestrictedApplicability()
	case string(applicabilityKindScoped):
		if raw.Scope == nil {
			return fmt.Errorf("requirement: unmarshal Applicability: %w: scoped requires a scope", ErrInvalidApplicability)
		}
		var err error
		result, err = NewApplicabilityFromScope(*raw.Scope)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("requirement: unmarshal Applicability: unrecognized kind %q: %w", raw.Kind, ErrInvalidApplicability)
	}
	*a = result
	return nil
}

// --- OriginRef -----------------------------------------------------------

// OriginRef names one engineering source category for a Requirement,
// with a free-text descriptive note (PEOS-005 §12: "Requirement Origin
// identifies the engineering source from which the Requirement was
// derived or motivated... Possible Origins include: customer needs;
// contractual obligations; regulations; standards; engineering analysis;
// risk analysis; safety analysis; stakeholder requests; previously
// established Requirements; Decisions").
//
// OriginRef is deliberately not core.Origin/core.OriginKind reused: those
// model the known/unknown/disputed/reconstructed *qualification* of a
// generic Artifact Revision's origin (see core.OriginKind's own doc
// comment, which explicitly defers "a structured origin-basis model...
// to a later Artifact Relation model or a specialization-specific
// field" — this is that field, for Requirement specifically).
//
// OriginRef is also deliberately not a structured reference to a specific
// Decision or Requirement instance: PEOS-005 §12 warns "A reference to a
// Decision as an Origin records provenance only... the Decision
// relationship and the Requirement Origin SHALL remain semantically
// distinguishable" — a real Decision-establishes-Requirement link belongs
// to a future Artifact Relation, never to Origin content.
type OriginRef struct {
	kind core.VocabularyValue
	note string
}

// NewOriginRef validates kind and note and returns an OriginRef. kind
// must be non-zero; note must be non-empty after trimming surrounding
// whitespace.
func NewOriginRef(kind core.VocabularyValue, note string) (OriginRef, error) {
	if kind.IsZero() {
		return OriginRef{}, fmt.Errorf("requirement: NewOriginRef: %w: kind must not be zero", ErrInvalidOrigin)
	}
	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return OriginRef{}, fmt.Errorf("requirement: NewOriginRef: %w: note must not be empty", ErrInvalidOrigin)
	}
	return OriginRef{kind: kind, note: trimmed}, nil
}

// Kind returns the origin's source-category vocabulary value.
func (o OriginRef) Kind() core.VocabularyValue { return o.kind }

// Note returns the origin's descriptive note.
func (o OriginRef) Note() string { return o.note }

// IsZero reports whether o is the zero value.
func (o OriginRef) IsZero() bool { return o.kind.IsZero() && o.note == "" }

type originRefJSON struct {
	Kind core.VocabularyValue `json:"kind"`
	Note string               `json:"note"`
}

// MarshalJSON encodes o as {"kind": ..., "note": ...}.
func (o OriginRef) MarshalJSON() ([]byte, error) {
	if o.IsZero() {
		return nil, fmt.Errorf("requirement: marshal OriginRef: %w", ErrInvalidOrigin)
	}
	return json.Marshal(originRefJSON{Kind: o.kind, Note: o.note})
}

// UnmarshalJSON decodes o from its JSON form, applying the same
// validation as NewOriginRef.
func (o *OriginRef) UnmarshalJSON(data []byte) error {
	var raw originRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal OriginRef: %w", err)
	}
	result, err := NewOriginRef(raw.Kind, raw.Note)
	if err != nil {
		return err
	}
	*o = result
	return nil
}

// --- Classification -----------------------------------------------------------

// Classification is an open-vocabulary organizational tag (PEOS-005 §14:
// "Classification provides organizational structure... This
// specification does not standardize classification taxonomies").
// Functional, Non-functional, Safety, Security, Regulatory, Performance,
// Reliability, Interface, and Operational are named only as examples in
// PEOS-005, never fixed as normative vocabulary — no constant for any of
// them is pre-declared here.
//
// Classification is not core.ArtifactRole: PEOS-005 places Classification
// at the Revision level ("Requirement Classification is Requirement
// content owned by the Artifact Revision in which it is defined"), while
// ArtifactRole is Artifact-identity-level and stable across every
// Revision of that Artifact — reusing ArtifactRole would collapse a
// per-Revision concept into a per-Artifact one.
type Classification struct{ value core.VocabularyValue }

// NewClassification validates value and returns a Classification.
func NewClassification(value core.VocabularyValue) (Classification, error) {
	if value.IsZero() {
		return Classification{}, fmt.Errorf("requirement: NewClassification: %w", ErrInvalidClassification)
	}
	return Classification{value: value}, nil
}

// Value returns the underlying vocabulary value.
func (c Classification) Value() core.VocabularyValue { return c.value }

// IsZero reports whether c is the zero value.
func (c Classification) IsZero() bool { return c.value.IsZero() }

// MarshalJSON encodes c as its canonical "namespace:value" string form.
func (c Classification) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Classification: %w", ErrInvalidClassification)
	}
	return json.Marshal(c.value)
}

// UnmarshalJSON decodes c from its canonical "namespace:value" string
// form, applying the same validation as NewClassification.
func (c *Classification) UnmarshalJSON(data []byte) error {
	var v core.VocabularyValue
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("requirement: unmarshal Classification: %w", err)
	}
	result, err := NewClassification(v)
	if err != nil {
		return err
	}
	*c = result
	return nil
}

// --- Rationale -----------------------------------------------------------

// Rationale is optional, informative, non-normative free text explaining
// why a Requirement exists (PEOS-005 §15: "Rationale explains why the
// Requirement exists... Rationale SHALL NOT modify normative meaning...
// SHALL NOT replace the Requirement Statement").
//
// Rationale is distinct from core.Provenance (which answers who/when/how
// a Revision was recorded, not why the Requirement exists), from a
// Decision, from an Artifact Relation, and from a Representation — none
// of those mechanisms is a substitute for this optional descriptive
// field, and this field is not a substitute for any of them.
type Rationale struct{ text string }

// NewRationale validates text and returns a Rationale. Surrounding
// whitespace is trimmed; text must be non-empty after trimming.
func NewRationale(text string) (Rationale, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Rationale{}, fmt.Errorf("requirement: NewRationale: %w", ErrInvalidRationale)
	}
	return Rationale{text: trimmed}, nil
}

// Text returns the rationale's text.
func (r Rationale) Text() string { return r.text }

// IsZero reports whether r is the zero value: the absence of rationale.
func (r Rationale) IsZero() bool { return r.text == "" }

type rationaleJSON struct {
	Text string `json:"text"`
}

// MarshalJSON encodes r as {"text": ...}.
func (r Rationale) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Rationale: %w", ErrInvalidRationale)
	}
	return json.Marshal(rationaleJSON{Text: r.text})
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as NewRationale.
func (r *Rationale) UnmarshalJSON(data []byte) error {
	var raw rationaleJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Rationale: %w", err)
	}
	result, err := NewRationale(raw.Text)
	if err != nil {
		return err
	}
	*r = result
	return nil
}

// --- Content -----------------------------------------------------------

// Content is the typed normative content PEOS-005 §8 assigns to every
// Artifact Revision whose Artifact is a Requirement: Statement, Subject,
// Applicability, Origin, Authority, Classification, and Rationale.
type Content struct {
	statements      []Statement
	subjects        []core.EngineeringSubjectRef
	subjectMode     SubjectCombination
	applicability   Applicability
	origins         []OriginRef
	authorities     []core.AuthorityRef
	classifications []Classification
	rationale       Rationale
}

// NewContent validates statements, subjects, subjectMode, and
// applicability and returns a Content with no Origins, Authorities,
// Classifications, or Rationale. Use the With* methods to add those.
//
// At least one Statement and one Subject are required (PEOS-005 §9, §10);
// a zero-value Statement or Subject among them is rejected. subjectMode
// and applicability must both be non-zero.
func NewContent(
	statements []Statement,
	subjects []core.EngineeringSubjectRef,
	subjectMode SubjectCombination,
	applicability Applicability,
) (Content, error) {
	if len(statements) == 0 {
		return Content{}, fmt.Errorf("requirement: NewContent: %w: at least one statement is required", ErrInvalidStatement)
	}
	stmts := make([]Statement, len(statements))
	for idx, s := range statements {
		if s.IsZero() {
			return Content{}, fmt.Errorf("requirement: NewContent: %w: statement must not be zero", ErrInvalidStatement)
		}
		stmts[idx] = s
	}

	if len(subjects) == 0 {
		return Content{}, fmt.Errorf("requirement: NewContent: %w", ErrMissingRequirementSubject)
	}
	subs := make([]core.EngineeringSubjectRef, len(subjects))
	for idx, sub := range subjects {
		if sub.IsZero() {
			return Content{}, fmt.Errorf("requirement: NewContent: %w: subject must not be zero", ErrMissingRequirementSubject)
		}
		subs[idx] = sub
	}

	if subjectMode.IsZero() {
		return Content{}, fmt.Errorf("requirement: NewContent: %w", ErrInvalidSubjectCombination)
	}
	if applicability.IsZero() {
		return Content{}, fmt.Errorf("requirement: NewContent: %w", ErrInvalidApplicability)
	}

	return Content{
		statements:    stmts,
		subjects:      subs,
		subjectMode:   subjectMode,
		applicability: applicability,
	}, nil
}

// WithOrigins returns a copy of c with its declared Origins set to
// exactly the values given, in the order given, replacing any previous
// Origins. A zero-value OriginRef among origins is rejected. Calling with
// no arguments clears the declared Origins.
func (c Content) WithOrigins(origins ...OriginRef) (Content, error) {
	if len(origins) == 0 {
		c.origins = nil
		return c, nil
	}
	cp := make([]OriginRef, len(origins))
	for idx, o := range origins {
		if o.IsZero() {
			return Content{}, fmt.Errorf("requirement: Content.WithOrigins: %w", ErrInvalidOrigin)
		}
		cp[idx] = o
	}
	c.origins = cp
	return c, nil
}

// WithAuthorities returns a copy of c with its declared Authorities set
// to exactly the values given, in the order given, replacing any previous
// Authorities. A zero-value core.AuthorityRef among authorities is
// rejected. Calling with no arguments clears the declared Authorities.
// Authorities are not deduplicated: PEOS-005 does not describe multiple
// Authorities as a set requiring uniqueness ("Multiple Authorities MAY
// jointly provide the normative authority represented by the same
// Requirement Artifact Revision" — §13).
func (c Content) WithAuthorities(authorities ...core.AuthorityRef) (Content, error) {
	if len(authorities) == 0 {
		c.authorities = nil
		return c, nil
	}
	cp := make([]core.AuthorityRef, len(authorities))
	for idx, a := range authorities {
		if a.IsZero() {
			return Content{}, fmt.Errorf("requirement: Content.WithAuthorities: %w", ErrInvalidAuthority)
		}
		cp[idx] = a
	}
	c.authorities = cp
	return c, nil
}

// WithClassifications returns a copy of c with its declared
// Classifications set to exactly the values given, in the order of first
// appearance, replacing any previous Classifications. A zero-value
// Classification, or an exact duplicate, among classifications is
// rejected. Calling with no arguments clears the declared Classifications.
func (c Content) WithClassifications(classifications ...Classification) (Content, error) {
	if len(classifications) == 0 {
		c.classifications = nil
		return c, nil
	}
	seen := make(map[string]bool, len(classifications))
	cp := make([]Classification, 0, len(classifications))
	for _, cl := range classifications {
		if cl.IsZero() {
			return Content{}, fmt.Errorf("requirement: Content.WithClassifications: %w", ErrInvalidClassification)
		}
		key := cl.Value().String()
		if seen[key] {
			return Content{}, fmt.Errorf("requirement: Content.WithClassifications: classification %q: %w", key, ErrDuplicateClassification)
		}
		seen[key] = true
		cp = append(cp, cl)
	}
	c.classifications = cp
	return c, nil
}

// WithRationale returns a copy of c with its declared Rationale set.
// Passing the zero Rationale is equivalent to declaring no rationale.
func (c Content) WithRationale(rationale Rationale) Content {
	c.rationale = rationale
	return c
}

// Statements returns a defensive copy of c's declared Statements, in
// declaration order.
func (c Content) Statements() []Statement {
	if len(c.statements) == 0 {
		return nil
	}
	cp := make([]Statement, len(c.statements))
	copy(cp, c.statements)
	return cp
}

// Subjects returns a defensive copy of c's declared Subjects, in
// declaration order.
func (c Content) Subjects() []core.EngineeringSubjectRef {
	if len(c.subjects) == 0 {
		return nil
	}
	cp := make([]core.EngineeringSubjectRef, len(c.subjects))
	copy(cp, c.subjects)
	return cp
}

// SubjectCombination returns c's declared subject combination mode.
func (c Content) SubjectCombination() SubjectCombination { return c.subjectMode }

// Applicability returns c's declared Applicability.
func (c Content) Applicability() Applicability { return c.applicability }

// Origins returns a defensive copy of c's declared Origins, in
// declaration order.
func (c Content) Origins() []OriginRef {
	if len(c.origins) == 0 {
		return nil
	}
	cp := make([]OriginRef, len(c.origins))
	copy(cp, c.origins)
	return cp
}

// Authorities returns a defensive copy of c's declared Authorities, in
// declaration order.
func (c Content) Authorities() []core.AuthorityRef {
	if len(c.authorities) == 0 {
		return nil
	}
	cp := make([]core.AuthorityRef, len(c.authorities))
	copy(cp, c.authorities)
	return cp
}

// Classifications returns a defensive copy of c's declared
// Classifications, in order of first appearance.
func (c Content) Classifications() []Classification {
	if len(c.classifications) == 0 {
		return nil
	}
	cp := make([]Classification, len(c.classifications))
	copy(cp, c.classifications)
	return cp
}

// Rationale returns c's declared Rationale, the zero Rationale if none
// was declared.
func (c Content) Rationale() Rationale { return c.rationale }

// IsZero reports whether c is the zero value.
func (c Content) IsZero() bool {
	return len(c.statements) == 0 && len(c.subjects) == 0 && c.subjectMode.IsZero() && c.applicability.IsZero()
}

type contentJSON struct {
	Statements      []Statement                  `json:"statements"`
	Subjects        []core.EngineeringSubjectRef `json:"subjects"`
	SubjectMode     SubjectCombination           `json:"subject_combination"`
	Applicability   Applicability                `json:"applicability"`
	Origins         []OriginRef                  `json:"origins,omitempty"`
	Authorities     []core.AuthorityRef          `json:"authorities,omitempty"`
	Classifications []Classification             `json:"classifications,omitempty"`
	Rationale       *Rationale                   `json:"rationale,omitempty"`
}

// MarshalJSON encodes c as {"statements":..., "subjects":...,
// "subject_combination":..., "applicability":..., ...}, omitting Origins,
// Authorities, Classifications, and Rationale when not set.
func (c Content) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Content: %w", ErrMissingRequirementContent)
	}
	raw := contentJSON{
		Statements:    c.statements,
		Subjects:      c.subjects,
		SubjectMode:   c.subjectMode,
		Applicability: c.applicability,
	}
	if len(c.origins) > 0 {
		raw.Origins = c.origins
	}
	if len(c.authorities) > 0 {
		raw.Authorities = c.authorities
	}
	if len(c.classifications) > 0 {
		raw.Classifications = c.classifications
	}
	if !c.rationale.IsZero() {
		raw.Rationale = &c.rationale
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes c from its JSON form, applying the same
// validation as NewContent and each With* method.
func (c *Content) UnmarshalJSON(data []byte) error {
	var raw contentJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Content: %w", err)
	}
	result, err := NewContent(raw.Statements, raw.Subjects, raw.SubjectMode, raw.Applicability)
	if err != nil {
		return err
	}
	if len(raw.Origins) > 0 {
		result, err = result.WithOrigins(raw.Origins...)
		if err != nil {
			return err
		}
	}
	if len(raw.Authorities) > 0 {
		result, err = result.WithAuthorities(raw.Authorities...)
		if err != nil {
			return err
		}
	}
	if len(raw.Classifications) > 0 {
		result, err = result.WithClassifications(raw.Classifications...)
		if err != nil {
			return err
		}
	}
	if raw.Rationale != nil {
		result = result.WithRationale(*raw.Rationale)
	}
	*c = result
	return nil
}
