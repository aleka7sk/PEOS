package requirement

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// Waiver is an authorized, scoped, explicitly represented limitation of a
// Requirement's normative applicability (PEOS-005 §27: "A Requirement MAY
// be waived only through applicable engineering governance... A waiver
// suspends or limits normative applicability within its defined scope.").
//
// Waiver is a separate immutable governance value record, not an Artifact,
// not an Artifact Revision, not an Artifact Relation, not a Lifecycle
// State Assignment, and not a Decision Outcome. It carries no identity of
// its own -- no ArtifactID, no RevisionID, no WaiverID or WaiverRef --
// matching PEOS-008 :520's own statement that PEOS-005 "does not define
// Waiver identity, Waiver lifecycle, Waiver revocation, or a Waiver
// historical model." Unlike this package's Requirement relationship
// wrappers (Derivation, Refinement, Decomposition, Dependency, Conflict,
// Supersession), Waiver composes no relation.Relation: §17.4 binds every
// explicitly represented Requirement *relationship* to PEOS-002's Artifact
// Relation contract, but §27 never calls a waiver a relationship, and §17.1
// governs relationship identity/lifecycle/history, not Waiver's. A Waiver
// therefore carries no "relation" key, no RelationType, and no
// Source/Target -- it names its single waived Requirement directly.
//
// A waiver SHALL NOT delete the Requirement, change its identity, or
// rewrite its content (§27); structurally, Waiver holds a
// core.RequirementRef -- a reference -- never a Requirement or Revision
// value, so none of those three prohibited effects is even representable
// here. Changing what a Requirement's content or Applicability says
// requires an ordinary new Artifact Revision under §25; a Waiver produces
// none. A Waiver likewise carries no Lifecycle State and does not import
// peos/lifecycle: §26 governs Lifecycle exclusively there, and §27 defines
// no Lifecycle interaction of its own for this package to model.
//
// Waiver's subject is fixed at Requirement identity level
// (core.RequirementRef), never Requirement Artifact Revision level: §27
// says "the Requirement" five times and never once says "Requirement
// Artifact Revision," in direct contrast to the six relationship types'
// own §18.1/§19.1/§20.1/§21.1/§22.1/§23.1 clauses, each of which is
// explicit about participant level. RequirementParticipant (participant.go)
// is deliberately not reused here: it exists so Dependency and Conflict
// can express an either-level choice §21.1/§22.1 explicitly require, and
// §27 asks no such question. RequirementParticipant also carries no JSON
// support today, and Waiver must not modify participant.go to add it.
//
// Waiver mandates exactly four fields, one per §27/§27.1/§27.2 "SHALL"
// obligation: the waived Requirement (§27), the authority under which the
// waiver is established (§27.1), the governance action through which it
// was established (§27, reusing GovernanceAction unchanged -- see
// governance.go's own doc comment, which predicted this reuse), and the
// waiver's scope (§27.2). All four are required constructor arguments;
// none is ever supplied later through a With* call, so a Waiver can never
// exist in a partially-established state. extension is the only optional
// field, for Product-specific data outside PEOS-owned fields.
//
// See doc.go's "Waiver is a governance value record, not a relationship"
// and "What Waiver deliberately does not model" sections for the full
// ontology and for the normative basis behind each field this type omits.
type Waiver struct {
	requirement      core.RequirementRef
	authority        core.AuthorityRef
	governanceAction GovernanceAction
	scope            core.Scope

	extension core.Extension
}

// NewWaiver validates requirement, authority, governanceAction, and scope
// and returns a Waiver. All four are mandatory and are rejected when zero:
// requirement (the waived Requirement, §27) with ErrInvalidWaiver,
// authority (§27.1) with ErrInvalidAuthority, governanceAction (§27) with
// ErrInvalidGovernanceAction, and scope (§27.2) with core.ErrInvalidScope.
//
// A successful call always returns a fully valid Waiver: every mandatory
// field is a required constructor argument, never a later With* call.
func NewWaiver(
	requirement core.RequirementRef,
	authority core.AuthorityRef,
	governanceAction GovernanceAction,
	scope core.Scope,
) (Waiver, error) {
	if requirement.IsZero() {
		return Waiver{}, fmt.Errorf("requirement: NewWaiver: %w: requirement must not be zero", ErrInvalidWaiver)
	}
	if authority.IsZero() {
		return Waiver{}, fmt.Errorf("requirement: NewWaiver: %w: authority must not be zero", ErrInvalidAuthority)
	}
	if governanceAction.IsZero() {
		return Waiver{}, fmt.Errorf("requirement: NewWaiver: %w: governance action must not be zero", ErrInvalidGovernanceAction)
	}
	if scope.IsZero() {
		return Waiver{}, fmt.Errorf("requirement: NewWaiver: %w: scope must not be zero", core.ErrInvalidScope)
	}
	return Waiver{
		requirement:      requirement,
		authority:        authority,
		governanceAction: governanceAction,
		scope:            scope,
	}, nil
}

// WithExtension returns a copy of w with its extension data set.
func (w Waiver) WithExtension(extension core.Extension) Waiver {
	w.extension = extension
	return w
}

// WithoutExtension returns a copy of w with its extension data cleared.
func (w Waiver) WithoutExtension() Waiver {
	w.extension = core.Extension{}
	return w
}

// Requirement returns w's waived Requirement.
func (w Waiver) Requirement() core.RequirementRef { return w.requirement }

// Authority returns w's declared authority (§27.1).
func (w Waiver) Authority() core.AuthorityRef { return w.authority }

// GovernanceAction returns w's declared governance action (§27).
func (w Waiver) GovernanceAction() GovernanceAction { return w.governanceAction }

// Scope returns w's declared scope (§27.2). Waiver's scope is mandatory
// and is therefore never absent on a valid Waiver.
func (w Waiver) Scope() core.Scope { return w.scope }

// Extension returns w's extension data.
func (w Waiver) Extension() core.Extension { return w.extension }

// IsZero reports whether w is the zero value.
func (w Waiver) IsZero() bool {
	return w.requirement.IsZero() && w.authority.IsZero() && w.governanceAction.IsZero() && w.scope.IsZero()
}

type waiverJSON struct {
	Requirement      core.RequirementRef `json:"requirement"`
	Authority        core.AuthorityRef   `json:"authority"`
	GovernanceAction GovernanceAction    `json:"governance_action"`
	Scope            core.Scope          `json:"scope"`
	Extension        *core.Extension     `json:"extension,omitempty"`
}

// MarshalJSON encodes w as {"requirement":..., "authority":...,
// "governance_action":..., "scope":..., "extension":...}, omitting
// extension when not set. requirement, authority, governance_action, and
// scope are never omitted. There is no "relation" key and no "type"
// discriminator -- a Waiver is not a relationship.
func (w Waiver) MarshalJSON() ([]byte, error) {
	if w.IsZero() {
		return nil, fmt.Errorf("requirement: marshal Waiver: %w", ErrInvalidWaiver)
	}
	raw := waiverJSON{
		Requirement:      w.requirement,
		Authority:        w.authority,
		GovernanceAction: w.governanceAction,
		Scope:            w.scope,
	}
	if !w.extension.IsZero() {
		raw.Extension = &w.extension
	}
	return json.Marshal(raw)
}

// UnmarshalJSON decodes w from its JSON form, applying the same
// validation as NewWaiver. An explicit JSON null for requirement,
// authority, governance_action, or scope decodes to that field's own
// zero value and is then rejected by NewWaiver through its owning
// sentinel; a missing key behaves identically. A decoded Waiver can
// never be constructor-impossible.
func (w *Waiver) UnmarshalJSON(data []byte) error {
	var raw waiverJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("requirement: unmarshal Waiver: %w: %w", ErrInvalidWaiver, err)
	}
	result, err := NewWaiver(raw.Requirement, raw.Authority, raw.GovernanceAction, raw.Scope)
	if err != nil {
		return err
	}
	if raw.Extension != nil {
		result = result.WithExtension(*raw.Extension)
	}
	*w = result
	return nil
}
