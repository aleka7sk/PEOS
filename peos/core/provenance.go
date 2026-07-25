package core

import "encoding/json"

// ActorRef is an opaque reference to whoever or whatever performed a
// recorded action (a user, a service account, an automated process). PEOS
// does not define actor identity structure; this type only carries a
// namespace and an identifier so that different actor systems (human
// accounts, CI systems, agents) can be distinguished without this package
// favoring one.
type ActorRef struct {
	namespace  string
	identifier string
}

// NewActorRef validates namespace and identifier and returns an ActorRef.
func NewActorRef(namespace, identifier string) (ActorRef, error) {
	ns, err := normalizeIdentityValue(namespace)
	if err != nil {
		return ActorRef{}, err
	}
	id, err := normalizeIdentityValue(identifier)
	if err != nil {
		return ActorRef{}, err
	}
	return ActorRef{namespace: ns, identifier: id}, nil
}

func (a ActorRef) Namespace() string  { return a.namespace }
func (a ActorRef) Identifier() string { return a.identifier }
func (a ActorRef) IsZero() bool       { return a.namespace == "" && a.identifier == "" }

type actorRefJSON struct {
	Namespace  string `json:"namespace"`
	Identifier string `json:"identifier"`
}

func (a ActorRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(actorRefJSON{Namespace: a.namespace, Identifier: a.identifier})
}

func (a *ActorRef) UnmarshalJSON(data []byte) error {
	var raw actorRefJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v, err := NewActorRef(raw.Namespace, raw.Identifier)
	if err != nil {
		return err
	}
	*a = v
	return nil
}

// Provenance is the PEOS provenance value model: a record of where a
// piece of content or an assertion came from. It is deliberately not a
// single required-fields struct, because different PEOS constructs
// require different subsets (a Validation Execution Record's provenance
// needs are not identical to an Artifact Relation's). This type only
// provides the representation; a construct-specific validator built on
// top of this package is responsible for enforcing which fields it
// requires.
//
// Provenance and AuthorityRef are deliberately separate types:
// provenance answers "where did this come from and who/what recorded
// it," while AuthorityRef (authority.go) answers "who had the right to
// approve it." A single value MAY appear in both roles for a given
// record, but this package never conflates the two concepts into one
// type.
//
// Provenance is an immutable value type; the With* methods return a
// modified copy rather than mutating the receiver.
type Provenance struct {
	source           VocabularyValue
	hasSource        bool
	actor            ActorRef
	hasActor         bool
	recordedAt       Timestamp
	hasRecordedAt    bool
	method           VocabularyValue
	hasMethod        bool
	externalSourceID string
	extension        Extension
}

// NewProvenance returns an empty Provenance with no fields set. Use the
// With* methods to populate the fields a specific construct requires.
func NewProvenance() Provenance { return Provenance{} }

// WithSource returns a copy of p with its source descriptor set.
func (p Provenance) WithSource(source VocabularyValue) Provenance {
	p.source, p.hasSource = source, true
	return p
}

// WithActor returns a copy of p with its actor reference set.
func (p Provenance) WithActor(actor ActorRef) Provenance {
	p.actor, p.hasActor = actor, true
	return p
}

// WithRecordedAt returns a copy of p with its recorded timestamp set.
func (p Provenance) WithRecordedAt(ts Timestamp) Provenance {
	p.recordedAt, p.hasRecordedAt = ts, true
	return p
}

// WithMethod returns a copy of p with its method/process descriptor set.
func (p Provenance) WithMethod(method VocabularyValue) Provenance {
	p.method, p.hasMethod = method, true
	return p
}

// WithExternalSourceID returns a copy of p with its external source
// identifier set.
func (p Provenance) WithExternalSourceID(id string) Provenance {
	p.externalSourceID = id
	return p
}

// WithExtension returns a copy of p with its extension data set.
func (p Provenance) WithExtension(ext Extension) Provenance {
	p.extension = ext
	return p
}

func (p Provenance) Source() (VocabularyValue, bool) { return p.source, p.hasSource }
func (p Provenance) Actor() (ActorRef, bool)         { return p.actor, p.hasActor }
func (p Provenance) RecordedAt() (Timestamp, bool)   { return p.recordedAt, p.hasRecordedAt }
func (p Provenance) Method() (VocabularyValue, bool) { return p.method, p.hasMethod }
func (p Provenance) ExternalSourceID() (string, bool) {
	return p.externalSourceID, p.externalSourceID != ""
}
func (p Provenance) Extension() Extension { return p.extension }

// IsZero reports whether no field of p has been set.
func (p Provenance) IsZero() bool {
	return !p.hasSource && !p.hasActor && !p.hasRecordedAt && !p.hasMethod &&
		p.externalSourceID == "" && p.extension.IsZero()
}

type provenanceJSON struct {
	Source           *VocabularyValue `json:"source,omitempty"`
	Actor            *ActorRef        `json:"actor,omitempty"`
	RecordedAt       *Timestamp       `json:"recorded_at,omitempty"`
	Method           *VocabularyValue `json:"method,omitempty"`
	ExternalSourceID string           `json:"external_source_id,omitempty"`
	Extension        *Extension       `json:"extension,omitempty"`
}

// MarshalJSON encodes only the fields of p that have been set.
func (p Provenance) MarshalJSON() ([]byte, error) {
	var raw provenanceJSON
	if p.hasSource {
		raw.Source = &p.source
	}
	if p.hasActor {
		raw.Actor = &p.actor
	}
	if p.hasRecordedAt {
		raw.RecordedAt = &p.recordedAt
	}
	if p.hasMethod {
		raw.Method = &p.method
	}
	raw.ExternalSourceID = p.externalSourceID
	if !p.extension.IsZero() {
		raw.Extension = &p.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes p from its JSON form, treating each field as
// present only if it appeared in the input.
func (p *Provenance) UnmarshalJSON(data []byte) error {
	var raw provenanceJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	result := NewProvenance()
	if raw.Source != nil {
		result = result.WithSource(*raw.Source)
	}
	if raw.Actor != nil {
		result = result.WithActor(*raw.Actor)
	}
	if raw.RecordedAt != nil {
		result = result.WithRecordedAt(*raw.RecordedAt)
	}
	if raw.Method != nil {
		result = result.WithMethod(*raw.Method)
	}
	if raw.ExternalSourceID != "" {
		result = result.WithExternalSourceID(raw.ExternalSourceID)
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*p = result
	return nil
}
