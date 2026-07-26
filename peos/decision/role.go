package decision

import (
	"encoding/json"
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
)

// RoleKind classifies which Decision role an Actor holds (PEOS-004
// "Decision Roles": "A Decision MAY identify one or more Actors in
// distinct roles. Decision roles MAY include: Decision Author; Decision
// Proposer; Decision Maker; Decision Approver; Decision Reviewer;
// Decision Executor; Decision Recorder; Decision Owner."). RoleKind is an
// open vocabulary: PEOS-004 :781 states outright that "The applicable
// Product contract MAY define additional roles", and the Extension Model
// (:1398-1411) explicitly names "additional Decision roles" among what a
// Product contract MAY define -- the same open-vocabulary shape as
// OutcomeKind and CommitmentEffect (outcome.go), reused here because the
// governing clause is materially identical, not merely by analogy.
type RoleKind struct{ value core.VocabularyValue }

// NewRoleKind wraps v as a RoleKind.
func NewRoleKind(v core.VocabularyValue) RoleKind { return RoleKind{value: v} }

func (k RoleKind) Value() core.VocabularyValue { return k.value }
func (k RoleKind) IsZero() bool                { return k.value.IsZero() }
func (k RoleKind) String() string              { return k.value.String() }

// Equal reports whether k and other name the same role kind value.
func (k RoleKind) Equal(other RoleKind) bool { return k.value.Equal(other.value) }

// MarshalJSON encodes k as its canonical "namespace:value" string form. A
// zero RoleKind is rejected: Role requires a non-zero RoleKind
// unconditionally, so a zero value should never reach marshaling.
func (k RoleKind) MarshalJSON() ([]byte, error) {
	if k.IsZero() {
		return nil, fmt.Errorf("decision: marshal RoleKind: %w", ErrInvalidDecisionRole)
	}
	return json.Marshal(k.value)
}

func (k *RoleKind) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &k.value)
}

var (
	RoleKindAuthor   = RoleKind{value: mustVocab("author")}
	RoleKindProposer = RoleKind{value: mustVocab("proposer")}
	RoleKindMaker    = RoleKind{value: mustVocab("maker")}
	RoleKindApprover = RoleKind{value: mustVocab("approver")}
	RoleKindReviewer = RoleKind{value: mustVocab("reviewer")}
	RoleKindExecutor = RoleKind{value: mustVocab("executor")}
	RoleKindRecorder = RoleKind{value: mustVocab("recorder")}
	RoleKindOwner    = RoleKind{value: mustVocab("owner")}
)

// Role associates an identified Actor with a single Decision role
// (PEOS-004 "Decision Roles": "A Decision MAY identify one or more
// Actors in distinct roles."; ":785 Role identity MUST NOT be inferred
// solely from document authorship or repository ownership."). Requiring
// an explicit core.ActorRef is what satisfies :785: no field on Role can
// be populated implicitly from a Decision Record's own authorship.
//
// Role is deliberately not Authority (authority.go): :751 requires
// Decision Authority to be distinguishable from authorship, facilitation,
// recommendation, implementation, and documentation, and Role's own
// vocabulary (Author, Proposer, Reviewer, Recorder, ...) names exactly
// those non-authority participations PEOS-004 keeps apart from :737's
// authority basis. Role is also not core.Provenance.Actor: Provenance
// records who produced a record, a single actor, not who held a role in
// the Decision itself, and a Decision MAY carry many Roles.
//
// A single Actor MAY hold more than one Role on the same Decision
// (PEOS-004 :783: "One Actor MAY perform multiple roles unless
// prohibited by an applicable separation-of-duties rule"); that rule, and
// its enforcement, is a Product/Runtime responsibility this package does
// not evaluate (see doc.go).
type Role struct {
	actor     core.ActorRef
	kind      RoleKind
	extension core.Extension
}

// NewRole validates actor and kind and returns a Role with no extension
// data. Both actor and kind must be non-zero.
func NewRole(actor core.ActorRef, kind RoleKind) (Role, error) {
	if actor.IsZero() {
		return Role{}, fmt.Errorf("decision: NewRole: %w: actor must not be zero", ErrInvalidDecisionRole)
	}
	if kind.IsZero() {
		return Role{}, fmt.Errorf("decision: NewRole: %w: kind must not be zero", ErrInvalidDecisionRole)
	}
	return Role{actor: actor, kind: kind}, nil
}

// WithExtension returns a copy of r with its extension data set.
func (r Role) WithExtension(extension core.Extension) Role {
	r.extension = extension
	return r
}

func (r Role) Actor() core.ActorRef      { return r.actor }
func (r Role) Kind() RoleKind            { return r.kind }
func (r Role) Extension() core.Extension { return r.extension }

// IsZero reports whether r is the zero value.
func (r Role) IsZero() bool { return r.actor.IsZero() && r.kind.IsZero() }

type roleJSON struct {
	Actor     core.ActorRef   `json:"actor"`
	Kind      RoleKind        `json:"kind"`
	Extension *core.Extension `json:"extension,omitempty"`
}

// MarshalJSON encodes r as {"actor":..., "kind":..., "extension":...},
// omitting extension when not set.
func (r Role) MarshalJSON() ([]byte, error) {
	if r.IsZero() {
		return nil, fmt.Errorf("decision: marshal Role: %w", ErrInvalidDecisionRole)
	}
	raw := roleJSON{Actor: r.actor, Kind: r.kind}
	if !r.extension.IsZero() {
		raw.Extension = &r.extension
	}
	return json.Marshal(raw)
}

type roleUnmarshalJSON struct {
	Actor     core.ActorRef   `json:"actor"`
	Kind      RoleKind        `json:"kind"`
	Extension json.RawMessage `json:"extension"`
}

// UnmarshalJSON decodes r from its JSON form, applying the same
// validation as NewRole. An explicit JSON null for "extension" is
// rejected rather than silently treated as absent.
func (r *Role) UnmarshalJSON(data []byte) error {
	var raw roleUnmarshalJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decision: unmarshal Role: %w: %w", ErrInvalidDecisionRole, err)
	}
	result, err := NewRole(raw.Actor, raw.Kind)
	if err != nil {
		return err
	}
	ext, err := decodeOptionalExtension(raw.Extension)
	if err != nil {
		return fmt.Errorf("decision: unmarshal Role: %w: %w", ErrInvalidDecisionRole, err)
	}
	if !ext.IsZero() {
		result = result.WithExtension(ext)
	}
	*r = result
	return nil
}
