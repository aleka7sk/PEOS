package lifecycle

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aleka7sk/PEOS/peos/core"
)

// LifecycleDefinitionVersionSupersession is an immutable, independently
// identifiable record establishing that one Lifecycle Definition Version
// supersedes another (PEOS-003 "Supersession": "A Lifecycle Definition
// Version MAY supersede another Lifecycle Definition Version. Supersession
// MUST identify: the superseding Lifecycle Definition Version; the
// superseded Lifecycle Definition Version; effective scope; compatibility
// consequences; migration requirements when applicable.").
//
// This is deliberately not any of the following:
//
//   - Artifact Supersession (PEOS-002): PEOS-003 itself states "Artifact
//     or Artifact Revision supersession remains governed by PEOS-002 and
//     applicable specialized lifecycles" -- a Lifecycle Definition Version
//     is not an Artifact Revision (see doc.go), so PEOS-002's Artifact
//     Supersession mechanism does not apply to it in the first place.
//   - Artifact Revision history: nothing here is a new Artifact Revision.
//   - Lifecycle Migration (PEOS-003 "Lifecycle Migration"): this record
//     only establishes the supersession fact and its normative
//     consequences. It does not execute a migration, remap State, track
//     migration progress, or reinterpret any historical Transition or
//     State Assignment -- those remain out of scope for this packet.
//   - A mutable field or "supersedes" pointer on DefinitionVersion:
//     DefinitionVersion remains exactly as immutable and as it was before
//     this type existed. Supersession is recorded as its own,
//     independently identified fact, carrying its own dedicated
//     core.LifecycleDefinitionVersionSupersessionID.
type LifecycleDefinitionVersionSupersession struct {
	id                       core.LifecycleDefinitionVersionSupersessionID
	supersedingVersion       core.LifecycleDefinitionVersionRef
	supersededVersion        core.LifecycleDefinitionVersionRef
	scope                    core.Scope
	compatibilityConsequence string
	migrationRequirement     string
	provenance               core.Provenance
	extension                core.Extension
}

// NewLifecycleDefinitionVersionSupersession validates its arguments and
// returns a LifecycleDefinitionVersionSupersession.
//
// compatibilityConsequence and migrationRequirement are each a required,
// non-empty textual statement rather than a controlled vocabulary or a
// structured shape: PEOS-003's Supersession section names "compatibility
// consequences" and "migration requirements" as required identification
// items but establishes no minimum controlled terms and no structured
// shape for either, so this constructor does not invent one.
//
// Note that PEOS-003's own wording for migration requirements is "when
// applicable" (conditional), but this constructor requires a non-empty
// value unconditionally, matching this package's existing precedent of
// being locally stricter than the bare minimum where a Product-neutral
// default would otherwise be ambiguous (see DefinitionVersion's required
// core.Provenance for the same kind of deliberate tightening).
//
// This constructor performs only local structural validation. It does
// not verify that supersedingVersion or supersededVersion identify
// Lifecycle Definition Versions that actually exist, does not verify
// chronological ordering between them, and does not enforce
// repository-level uniqueness -- all of that requires data this package
// does not fetch.
func NewLifecycleDefinitionVersionSupersession(
	id core.LifecycleDefinitionVersionSupersessionID,
	supersedingVersion core.LifecycleDefinitionVersionRef,
	supersededVersion core.LifecycleDefinitionVersionRef,
	scope core.Scope,
	compatibilityConsequence string,
	migrationRequirement string,
	provenance core.Provenance,
) (LifecycleDefinitionVersionSupersession, error) {
	if id.IsZero() {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w", ErrInvalidLifecycleSupersession)
	}
	if supersedingVersion.IsZero() {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: superseding version must not be zero", ErrInvalidLifecycleSupersession)
	}
	if supersededVersion.IsZero() {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: superseded version must not be zero", ErrInvalidLifecycleSupersession)
	}
	if supersedingVersion == supersededVersion {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: superseding and superseded version must differ", ErrInvalidLifecycleSupersession)
	}
	if scope.IsZero() {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: scope must not be zero", ErrInvalidLifecycleSupersession)
	}
	if strings.TrimSpace(compatibilityConsequence) == "" {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: compatibility consequence must not be empty", ErrInvalidLifecycleSupersession)
	}
	if strings.TrimSpace(migrationRequirement) == "" {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: migration requirement must not be empty", ErrInvalidLifecycleSupersession)
	}
	if provenance.IsZero() {
		return LifecycleDefinitionVersionSupersession{}, fmt.Errorf("lifecycle: NewLifecycleDefinitionVersionSupersession: %w: provenance must not be zero", ErrInvalidLifecycleSupersession)
	}
	return LifecycleDefinitionVersionSupersession{
		id:                       id,
		supersedingVersion:       supersedingVersion,
		supersededVersion:        supersededVersion,
		scope:                    scope,
		compatibilityConsequence: compatibilityConsequence,
		migrationRequirement:     migrationRequirement,
		provenance:               provenance,
	}, nil
}

// WithExtension returns a copy of s with its extension data set.
func (s LifecycleDefinitionVersionSupersession) WithExtension(e core.Extension) LifecycleDefinitionVersionSupersession {
	s.extension = e
	return s
}

func (s LifecycleDefinitionVersionSupersession) ID() core.LifecycleDefinitionVersionSupersessionID {
	return s.id
}

func (s LifecycleDefinitionVersionSupersession) SupersedingVersion() core.LifecycleDefinitionVersionRef {
	return s.supersedingVersion
}

func (s LifecycleDefinitionVersionSupersession) SupersededVersion() core.LifecycleDefinitionVersionRef {
	return s.supersededVersion
}

func (s LifecycleDefinitionVersionSupersession) Scope() core.Scope { return s.scope }

func (s LifecycleDefinitionVersionSupersession) CompatibilityConsequence() string {
	return s.compatibilityConsequence
}

func (s LifecycleDefinitionVersionSupersession) MigrationRequirement() string {
	return s.migrationRequirement
}

func (s LifecycleDefinitionVersionSupersession) Provenance() core.Provenance { return s.provenance }

func (s LifecycleDefinitionVersionSupersession) Extension() core.Extension { return s.extension }

// IsZero reports whether s is the zero value.
func (s LifecycleDefinitionVersionSupersession) IsZero() bool { return s.id.IsZero() }

// Ref returns a core.LifecycleDefinitionVersionSupersessionRef identifying s.
func (s LifecycleDefinitionVersionSupersession) Ref() (core.LifecycleDefinitionVersionSupersessionRef, error) {
	return core.NewLifecycleDefinitionVersionSupersessionRef(s.id)
}

type lifecycleDefinitionVersionSupersessionJSON struct {
	ID                       core.LifecycleDefinitionVersionSupersessionID `json:"id"`
	SupersedingVersion       core.LifecycleDefinitionVersionRef            `json:"superseding_version"`
	SupersededVersion        core.LifecycleDefinitionVersionRef            `json:"superseded_version"`
	Scope                    core.Scope                                    `json:"scope"`
	CompatibilityConsequence string                                        `json:"compatibility_consequence"`
	MigrationRequirement     string                                        `json:"migration_requirement"`
	Provenance               core.Provenance                               `json:"provenance"`
	Extension                *core.Extension                               `json:"extension,omitempty"`
}

// MarshalJSON encodes s as a flat JSON object containing every required
// field, omitting extension when not set. s is not nested as an Artifact
// or Artifact Revision -- it is an immutable record with its own
// identity, exactly like StateAssignment.
func (s LifecycleDefinitionVersionSupersession) MarshalJSON() ([]byte, error) {
	if s.IsZero() {
		return nil, fmt.Errorf("lifecycle: marshal LifecycleDefinitionVersionSupersession: %w", ErrInvalidLifecycleSupersession)
	}
	raw := lifecycleDefinitionVersionSupersessionJSON{
		ID: s.id, SupersedingVersion: s.supersedingVersion, SupersededVersion: s.supersededVersion,
		Scope: s.scope, CompatibilityConsequence: s.compatibilityConsequence,
		MigrationRequirement: s.migrationRequirement, Provenance: s.provenance,
	}
	if !s.extension.IsZero() {
		raw.Extension = &s.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes s from its JSON form, applying the same
// validation as NewLifecycleDefinitionVersionSupersession. Unknown
// ordinary JSON fields are ignored.
//
// Every required field here (SupersedingVersion, SupersededVersion,
// Scope) is decoded as its own concrete type rather than as a pointer, so
// an explicit JSON null for any of them is passed straight to that
// type's own UnmarshalJSON (encoding/json invokes a non-pointer field's
// Unmarshaler even for a null token, unlike its behavior for a *T field)
// and is rejected there by that type's own zero-value validation --
// exactly as an explicit "{}" already is. No separate raw-capture
// handling is needed for required fields; see assignment.go and
// definition.go for where that technique is used instead, for genuinely
// optional fields.
func (s *LifecycleDefinitionVersionSupersession) UnmarshalJSON(data []byte) error {
	var raw lifecycleDefinitionVersionSupersessionJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("lifecycle: unmarshal LifecycleDefinitionVersionSupersession: %w", err)
	}
	result, err := NewLifecycleDefinitionVersionSupersession(
		raw.ID, raw.SupersedingVersion, raw.SupersededVersion, raw.Scope,
		raw.CompatibilityConsequence, raw.MigrationRequirement, raw.Provenance,
	)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*s = result
	return nil
}
