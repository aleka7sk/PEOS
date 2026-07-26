package decision

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// ArtifactTypeDecisionRecord is the dedicated core.ArtifactType a Record's
// underlying core.Artifact MUST declare.
var ArtifactTypeDecisionRecord = core.NewArtifactType(mustVocab("decision-record"))

// Record is a Decision Record (PEOS-004): "A Decision Record is an
// Artifact that represents a Decision." Record composes core.Artifact by
// named field, exactly like peos/requirement.Requirement, rather than
// introducing a dedicated non-Artifact identity type -- unlike Decision
// itself (see doc.go).
//
// Record and Decision identity are distinct: a Decision and its Decision
// Record MAY have different identifiers (PEOS-002). The association from
// Record to Decision lives on Record itself, not on RecordContent, so that
// every Revision of a given Record is structurally guaranteed to name the
// same Decision.
type Record struct {
	core     core.Artifact
	decision core.DecisionRef
}

// NewRecord validates that artifact is non-zero and declares
// ArtifactTypeDecisionRecord, that decision is non-zero, and returns a
// Record.
func NewRecord(artifact core.Artifact, decision core.DecisionRef) (Record, error) {
	if artifact.IsZero() {
		return Record{}, fmt.Errorf("decision: NewRecord: %w", ErrInvalidDecisionRecord)
	}
	if artifact.Type() != ArtifactTypeDecisionRecord {
		return Record{}, fmt.Errorf("decision: NewRecord: %w", ErrArtifactTypeMismatch)
	}
	if decision.IsZero() {
		return Record{}, fmt.Errorf("decision: NewRecord: %w: decision reference must not be zero", ErrInvalidDecisionRecord)
	}
	return Record{core: artifact, decision: decision}, nil
}

// Core returns the Record's underlying core.Artifact.
func (r Record) Core() core.Artifact { return r.core }

// ID returns the Record's Artifact identity.
func (r Record) ID() core.ArtifactID { return r.core.ID() }

// Decision returns the core.DecisionRef this Record documents.
func (r Record) Decision() core.DecisionRef { return r.decision }

// IsZero reports whether r is the zero value.
func (r Record) IsZero() bool { return r.core.IsZero() && r.decision.IsZero() }

type recordJSON struct {
	Core     core.Artifact    `json:"core"`
	Decision core.DecisionRef `json:"decision"`
}

// MarshalJSON encodes r as {"core": {...}, "decision": {"decision_id":
// ...}}, per the nested-composition strategy documented on
// core.ArtifactRevision.
func (r Record) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("decision: marshal Record: %w", ErrInvalidDecisionRecord)
	}
	return json.Marshal(recordJSON{Core: r.core, Decision: r.decision})
}

// UnmarshalJSON decodes r from its nested {"core": {...}, "decision":
// {...}} JSON form, applying the same validation as NewRecord.
func (r *Record) UnmarshalJSON(data []byte) error {
	var raw recordJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Record: %w", err)
	}
	result, err := NewRecord(raw.Core, raw.Decision)
	if err != nil {
		return err
	}
	*r = result
	return nil
}

// RecordContent is a Decision Record Revision's typed content. It is
// deliberately near-empty: PEOS-004's Decision Record field list (§ Decision
// Record) is a SHOULD, not a MUST, and the documentary payload of a
// Decision Record -- its narrative -- belongs to a core.Representation on
// its ArtifactRevision, not to typed content (see doc.go). RecordContent
// therefore carries only Supplemental Evidence, which PEOS-004 requires be
// distinguishable from the Evidence that formed the original Decision
// Basis (Basis.Evidence). A zero-value RecordContent is valid.
type RecordContent struct {
	supplementalEvidence []core.EvidenceArtifactRevisionRef
	extension            core.Extension
}

// NewRecordContent returns a zero-value RecordContent. Use the With*
// methods to add Supplemental Evidence and extension data.
func NewRecordContent() RecordContent { return RecordContent{} }

// WithSupplementalEvidence returns a copy of c with its declared
// Supplemental Evidence set to exactly the values given, in the order
// given, replacing any previous Supplemental Evidence. A zero-value
// Evidence reference among evidence is rejected. Calling with no arguments
// clears the declared Supplemental Evidence.
func (c RecordContent) WithSupplementalEvidence(evidence ...core.EvidenceArtifactRevisionRef) (RecordContent, error) {
	if len(evidence) == 0 {
		c.supplementalEvidence = nil
		return c, nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(evidence))
	for idx, e := range evidence {
		if e.IsZero() {
			return RecordContent{}, fmt.Errorf("decision: RecordContent.WithSupplementalEvidence: %w: evidence reference must not be zero", ErrInvalidDecisionRecordRevision)
		}
		cp[idx] = e
	}
	c.supplementalEvidence = cp
	return c, nil
}

// WithExtension returns a copy of c with its extension data set.
func (c RecordContent) WithExtension(e core.Extension) RecordContent {
	c.extension = e
	return c
}

// SupplementalEvidence returns a defensive copy of c's declared
// Supplemental Evidence, in declaration order.
func (c RecordContent) SupplementalEvidence() []core.EvidenceArtifactRevisionRef {
	if len(c.supplementalEvidence) == 0 {
		return nil
	}
	cp := make([]core.EvidenceArtifactRevisionRef, len(c.supplementalEvidence))
	copy(cp, c.supplementalEvidence)
	return cp
}

func (c RecordContent) Extension() core.Extension { return c.extension }

// IsZero reports whether c is the zero value. A zero RecordContent is a
// valid, legitimate value -- see the type's own doc comment.
func (c RecordContent) IsZero() bool {
	return len(c.supplementalEvidence) == 0 && c.extension.IsZero()
}

type recordContentJSON struct {
	SupplementalEvidence []core.EvidenceArtifactRevisionRef `json:"supplemental_evidence,omitempty"`
	Extension            *core.Extension                    `json:"extension,omitempty"`
}

// MarshalJSON encodes c as {"supplemental_evidence":[...],
// "extension":...}, omitting both when not set. Unlike every other type in
// this package, a zero-value c marshals successfully, to "{}" -- see the
// type's own doc comment.
func (c RecordContent) MarshalJSON() ([]byte, error) {
	raw := recordContentJSON{}
	if len(c.supplementalEvidence) > 0 {
		raw.SupplementalEvidence = c.supplementalEvidence
	}
	if !c.extension.IsZero() {
		raw.Extension = &c.extension
	}
	return json.Marshal(raw)
}

type recordContentUnmarshalJSON struct {
	SupplementalEvidence json.RawMessage `json:"supplemental_evidence"`
	Extension            json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes c from its JSON form, applying the same validation
// as WithSupplementalEvidence. An explicit JSON null for
// "supplemental_evidence" or "extension" is rejected rather than silently
// treated as absent. An empty or entirely absent object decodes as the
// zero RecordContent.
func (c *RecordContent) UnmarshalJSON(data []byte) error {
	var raw recordContentUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal RecordContent: %w", err)
	}
	result := NewRecordContent()
	if len(raw.SupplementalEvidence) > 0 {
		if string(raw.SupplementalEvidence) == "null" {
			return fmt.Errorf("decision: unmarshal RecordContent: %w: supplemental_evidence must not be null", ErrInvalidDecisionRecordRevision)
		}
		var evidence []core.EvidenceArtifactRevisionRef
		if err := json.Unmarshal(raw.SupplementalEvidence, &evidence); err != nil {
			return fmt.Errorf("decision: unmarshal RecordContent: %w", err)
		}
		var err error
		if result, err = result.WithSupplementalEvidence(evidence...); err != nil {
			return err
		}
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal RecordContent: %w: %w", ErrInvalidDecisionRecordRevision, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*c = result
	return nil
}

// RecordRevision is shorthand for "an Artifact Revision whose Artifact is a
// Decision Record," composing core.ArtifactRevision by named field and
// pairing it with typed RecordContent, exactly like
// peos/requirement.Revision.
type RecordRevision struct {
	core    core.ArtifactRevision
	content RecordContent
}

// newRecordRevisionFromParts validates revision and content without
// reference to any Record, and is the path both NewRecordRevision and
// UnmarshalJSON share. It cannot, and does not attempt to, check that
// revision belongs to any particular Record -- see NewRecordRevision and
// UnmarshalJSON's own documentation for why that check requires a Record
// value a Revision's own JSON does not carry (the same limitation
// peos/requirement.Revision documents for itself). content MAY be zero.
func newRecordRevisionFromParts(revision core.ArtifactRevision, content RecordContent) (RecordRevision, error) {
	if revision.IsZero() {
		return RecordRevision{}, fmt.Errorf("decision: %w: core revision must not be zero", ErrInvalidDecisionRecordRevision)
	}
	return RecordRevision{core: revision, content: content}, nil
}

// NewRecordRevision validates record, revision, and content and returns a
// RecordRevision. record and revision must both be non-zero; content MAY
// be zero. revision.ArtifactID() must equal record.ID().
func NewRecordRevision(record Record, revision core.ArtifactRevision, content RecordContent) (RecordRevision, error) {
	if record.IsZero() {
		return RecordRevision{}, fmt.Errorf("decision: NewRecordRevision: %w: record must not be zero", ErrInvalidDecisionRecordRevision)
	}
	result, err := newRecordRevisionFromParts(revision, content)
	if err != nil {
		return RecordRevision{}, err
	}
	if revision.ArtifactID() != record.ID() {
		return RecordRevision{}, fmt.Errorf("decision: NewRecordRevision: %w", ErrArtifactIDMismatch)
	}
	return result, nil
}

// Core returns the Revision's underlying core.ArtifactRevision.
func (r RecordRevision) Core() core.ArtifactRevision { return r.core }

// Content returns the Revision's typed Decision Record content.
func (r RecordRevision) Content() RecordContent { return r.content }

// IsZero reports whether r is the zero value. Because a zero RecordContent
// is itself valid, r's zero-ness depends only on its core Revision.
func (r RecordRevision) IsZero() bool { return r.core.IsZero() }

// Ref returns a core.ArtifactRevisionRef identifying r.
func (r RecordRevision) Ref() (core.ArtifactRevisionRef, error) {
	return core.NewArtifactRevisionRef(r.core.ArtifactID(), r.core.RevisionID())
}

type recordRevisionJSON struct {
	Core    core.ArtifactRevision `json:"core"`
	Content RecordContent         `json:"content"`
}

// MarshalJSON encodes r as {"core": {...}, "content": {...}}, per the
// nested-composition strategy documented on core.ArtifactRevision.
// "content" is always present, even when zero (RecordContent's own
// MarshalJSON encodes a zero value as "{}").
func (r RecordRevision) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("decision: marshal RecordRevision: %w", ErrInvalidDecisionRecordRevision)
	}
	return json.Marshal(recordRevisionJSON{Core: r.core, Content: r.content})
}

// UnmarshalJSON decodes r from its nested {"core": {...}, "content": {...}}
// JSON form.
func (r *RecordRevision) UnmarshalJSON(data []byte) error {
	var raw recordRevisionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal RecordRevision: %w", err)
	}
	result, err := newRecordRevisionFromParts(raw.Core, raw.Content)
	if err != nil {
		return err
	}
	*r = result
	return nil
}
