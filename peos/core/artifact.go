package core

import (
	"encoding/json"
	"fmt"
)

// Artifact is the PEOS Artifact identity (PEOS-002 Artifact): a stable,
// persistent logical engineering identity, independent of any particular
// Revision, Representation, or storage location.
//
// Artifact and ArtifactRevision are independent domain records connected
// only by ArtifactID. Artifact does not contain, embed, reference, or
// cache any Revision, and has no "current Revision" concept: PEOS-002
// treats the current or applicable Revision as derived or Product-
// selected, never a field owned by Artifact (see doc.go). A combined
// "Artifact plus its Revisions" shape is an application-level export or
// interchange envelope, not part of this domain model.
//
// Artifact carries no provenance field of its own in this packet.
//
// PEOS-002 §Artifact Creation requires Artifact creation to "record or
// enable the recording of... initial provenance," distinct from the
// provenance PEOS-002 also requires on every recorded Artifact Revision.
// This is a real requirement this package does not claim to be exempt
// from — Packet B does not assert that PEOS-002 requires no provenance
// at Artifact-creation time. In practice:
//
//   - When an Artifact's identity and its first Revision are recorded
//     together (the common case), that founding Revision's own required,
//     non-zero Provenance (see ArtifactRevision) may serve as the
//     practical creation record: the same act that establishes the first
//     Revision is, in that case, also the act that establishes the
//     Artifact's participation in normative Engineering State.
//   - When an Artifact's identity is deliberately recorded before its
//     first Revision (PEOS-002 explicitly permits this: "An Artifact MAY
//     exist before its first Revision is formally recorded, but such an
//     Artifact MUST NOT be treated as reproducible or validated"), no
//     mechanism in this package retains that creation-time provenance.
//     The creation operation, or a surrounding persistence or process
//     layer outside this package, is responsible for retaining initial
//     provenance until it is represented by a founding Revision or
//     another conformant persistent record. This package does not
//     introduce a repository, service, or process type to do this —
//     that responsibility is deliberately left to a later packet or to
//     the calling application.
//
// Artifact also carries no external identifiers, title, description,
// authority, status, or Lifecycle state — see artifact.go's package-level
// rationale in the Packet B blueprint: external identifiers are a Product
// convenience deferred out of this packet; title/description are either
// Type-specific normative content (belongs in a Representation) or
// operational metadata (belongs in Extension), a boundary PEOS-002
// explicitly assigns to "the applicable Artifact Type or contract," not
// to this package; authority is specialization-specific; status and
// Lifecycle state are governed by PEOS-003 and explicitly out of scope
// for this packet.
type Artifact struct {
	id        ArtifactID
	typ       ArtifactType
	roles     []ArtifactRole
	hasScope  bool
	scope     Scope
	extension Extension
}

// NewArtifact validates id and artifactType and returns an Artifact with
// no roles, no scope, and no extension data. Use WithRoles, WithScope,
// and WithExtension to add those.
func NewArtifact(id ArtifactID, artifactType ArtifactType) (Artifact, error) {
	if id.IsZero() {
		return Artifact{}, fmt.Errorf("core: NewArtifact: %w: %w", ErrInvalidArtifact, ErrEmptyIdentity)
	}
	if artifactType.IsZero() {
		return Artifact{}, fmt.Errorf("core: NewArtifact: %w: artifact type must not be zero", ErrInvalidArtifact)
	}
	return Artifact{id: id, typ: artifactType}, nil
}

// WithRoles returns a copy of a with its declared roles set to exactly
// the roles given, in the order given. Calling WithRoles again replaces
// the previous roles entirely; it does not accumulate across calls. An
// exact duplicate role within one call is rejected. A zero-value role
// within roles is rejected. Declaration order is preserved.
func (a Artifact) WithRoles(roles ...ArtifactRole) (Artifact, error) {
	if len(roles) == 0 {
		a.roles = nil
		return a, nil
	}
	seen := make(map[string]bool, len(roles))
	deduped := make([]ArtifactRole, 0, len(roles))
	for _, role := range roles {
		if role.IsZero() {
			return Artifact{}, fmt.Errorf("core: Artifact.WithRoles: %w: role must not be zero", ErrInvalidArtifact)
		}
		key := role.Value().String()
		if seen[key] {
			return Artifact{}, fmt.Errorf("core: Artifact.WithRoles: role %q: %w", key, ErrDuplicateArtifactRole)
		}
		seen[key] = true
		deduped = append(deduped, role)
	}
	a.roles = deduped
	return a, nil
}

// WithScope returns a copy of a with its declared scope set. Passing the
// zero Scope is equivalent to leaving scope unset.
func (a Artifact) WithScope(scope Scope) Artifact {
	a.scope, a.hasScope = scope, !scope.IsZero()
	return a
}

// WithExtension returns a copy of a with its extension data set.
func (a Artifact) WithExtension(extension Extension) Artifact {
	a.extension = extension
	return a
}

// ID returns the Artifact's identity.
func (a Artifact) ID() ArtifactID { return a.id }

// Type returns the Artifact's declared Artifact Type.
func (a Artifact) Type() ArtifactType { return a.typ }

// Roles returns a defensive copy of the Artifact's declared roles, in
// declaration order.
func (a Artifact) Roles() []ArtifactRole {
	if len(a.roles) == 0 {
		return nil
	}
	cp := make([]ArtifactRole, len(a.roles))
	copy(cp, a.roles)
	return cp
}

// Scope returns the Artifact's declared scope, and whether one is set.
func (a Artifact) Scope() (Scope, bool) { return a.scope, a.hasScope }

// Extension returns the Artifact's extension data.
func (a Artifact) Extension() Extension { return a.extension }

// IsZero reports whether a is the zero value.
func (a Artifact) IsZero() bool { return a.id.IsZero() && a.typ.IsZero() }

type artifactJSON struct {
	ID        ArtifactID     `json:"artifact_id"`
	Type      ArtifactType   `json:"artifact_type"`
	Roles     []ArtifactRole `json:"roles,omitempty"`
	Scope     *Scope         `json:"scope,omitempty"`
	Extension *Extension     `json:"extension,omitempty"`
}

// MarshalJSON encodes a as {"artifact_id":..., "artifact_type":..., ...},
// omitting roles, scope, and extension when not set.
func (a Artifact) MarshalJSON() ([]byte, error) {
	raw := artifactJSON{ID: a.id, Type: a.typ}
	if len(a.roles) > 0 {
		raw.Roles = a.roles
	}
	if a.hasScope {
		raw.Scope = &a.scope
	}
	if !a.extension.IsZero() {
		raw.Extension = &a.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes a from its JSON form, applying the same
// validation as NewArtifact and WithRoles.
func (a *Artifact) UnmarshalJSON(data []byte) error {
	var raw artifactJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("core: unmarshal Artifact: %w", err)
	}
	result, err := NewArtifact(raw.ID, raw.Type)
	if err != nil {
		return err
	}
	if len(raw.Roles) > 0 {
		result, err = result.WithRoles(raw.Roles...)
		if err != nil {
			return err
		}
	}
	if raw.Scope != nil {
		result = result.WithScope(*raw.Scope)
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*a = result
	return nil
}
