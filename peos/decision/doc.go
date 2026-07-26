// Package decision implements the core, structural subset of PEOS-004's
// Decision Model on top of the stable peos/core package.
//
// This package implements Decision, Authority, Basis (with its Evidence,
// Assumption, Constraint, and Uncertainty components), Alternative,
// Outcome (with its Engineering Commitment value), Decision Record (with
// its Revision and Content), Decision Supersession, Decision
// Invalidation, Decision Conflict, Conflict Resolution, Decision Roles,
// and Decision Consequences. It deliberately does not implement
// Delegation, Decision relation-type constants, or any approval/voting
// workflow -- see "Deferred" below.
//
// # Decision is not an Artifact
//
// PEOS-004 states "A Decision MUST NOT be treated as equivalent to the
// Artifact that records it" (the Record Separation Invariant), and PEOS-002
// states "A Decision MUST NOT be reduced to an Artifact Type." Decision
// therefore carries its own core.DecisionID, independent of core.ArtifactID,
// and is never composed from core.Artifact.
//
// # Decision Record is an Artifact
//
// PEOS-004 states "A Decision Record is an Artifact that represents a
// Decision." Record therefore composes core.Artifact by named field,
// exactly like peos/requirement.Requirement, and declares
// ArtifactTypeDecisionRecord. A Decision and its Decision Record MAY have
// different identifiers (PEOS-002); Record carries its core.DecisionRef on
// the Record itself, not on RecordContent, so every Revision of a given
// Record is structurally guaranteed to name the same Decision.
//
// # Decision has no revision mechanism
//
// Unlike a Requirement or a Decision Record, a Decision is not an Artifact
// and is never revised in place. A material semantic change to a Decision's
// content is a new Decision, with its own core.DecisionID -- not a new
// Decision Revision. This package therefore defines no DecisionRevision
// type. A Decision Record's Artifact Revisions document a Decision's
// content over time (for example, as a Decision's own recognition and
// wording matures before an authority basis is established), not a
// Decision's own mutation.
//
// # Decision lifecycle state is out of scope
//
// PEOS-004 states "A Decision is a Lifecycle Subject" (PEOS-003), and
// core.NewLifecycleSubjectRefFromDecision already exists for that purpose.
// This package defines no Decision state field and no state vocabulary:
// Decision lifecycle state is governed entirely by peos/lifecycle. Decision
// Outcome (Outcome, this package) is a distinct concept from lifecycle
// state -- it records what was decided, not whether the Decision is
// currently applicable.
//
// # Engineering Commitment is a value, not an Entity
//
// PEOS-004 states "Engineering Commitment is a semantic component of the
// Decision Model and is not required to be a separate top-level PEOS
// Entity." Commitment therefore carries no identity of its own;
// core.EngineeringCommitmentRef is derived from the owning Decision's
// identity, outside this package's Commitment value.
//
// # Deliberate choices
//
// Provenance is optional on Decision: PEOS-004 never mentions provenance,
// unlike PEOS-002's requirement for every Artifact Relation. Alternative
// has no key: PEOS-004 defines no Alternative identity, and its own worked
// example expresses a selected option as prose in Outcome.Statement, not as
// a structural reference to a specific Alternative -- Outcome therefore
// also carries no selected-alternative link. RecordContent has no
// narrative field: the documentary payload of a Decision Record belongs to
// a core.Representation on its ArtifactRevision, not to typed content: the
// Decision Record field list in PEOS-004 is a SHOULD, not a MUST, and a
// zero-value RecordContent is valid.
//
// # Decision Basis may combine Evidence, Assumptions, Constraints, and Uncertainties
//
// A Basis may consist of any non-empty combination of its four content
// collections (PEOS-004:406-419's own "Decision Basis MAY include" list):
// Evidence, Assumptions, Constraints, and Uncertainties. Evidence alone,
// or a Basis carrying only Assumptions with no Evidence at all, are both
// conformant -- nothing in PEOS-004 privileges Evidence over the other
// three. Assumption, Constraint, and Uncertainty are typed value
// structures, not free-form strings, because their normative field
// carriage cannot fit in a single string: see each type's own doc comment
// for the exact PEOS-004 clause behind every field it exposes.
//
// Every optional field on Assumption, Constraint, and Uncertainty traces
// to an explicit PEOS-004 "SHOULD identify" clause -- Assumption's five
// (:458-464), Constraint's one (:491), and Uncertainty's none (:544's "MAY
// concern" list carries no identify obligation at all) -- and nothing
// beyond those clauses is modeled. In particular, neither Constraint nor
// Uncertainty carries an origin- or concern-category vocabulary:
// PEOS-004:478-489's "Constraint MAY originate from" list and :544-555's
// "Uncertainty MAY concern" list are structurally identical to
// PEOS-004:737-749's authority-origin list, which this package already
// declines to canonize (core.AuthorityRef's Kind is an open vocabulary
// with zero predeclared constants); a Product needing that classification
// carries it in Extension.
//
// Assumption.Uncertainty and Basis.Uncertainties are two distinct
// directions of relation, not the same fact recorded twice.
// Assumption.Uncertainty (PEOS-004:462, "An Assumption SHOULD identify...
// its uncertainty") qualifies exactly the one Assumption that carries it.
// Basis.Uncertainties (PEOS-004:542, "Known material uncertainty in a
// Decision Basis MUST be explicit") records a standalone known material
// uncertainty fact, independent of any particular Assumption -- even
// though :546 permits such a fact to "concern... assumptions" in general,
// that permission does not identify which Assumption it concerns, so it
// cannot discharge :462's per-Assumption obligation. Both may be present
// without duplication, and neither substitutes for the other.
//
// Constraint is Decision-Basis-scoped: PEOS-004:476 defines it as
// restricting "the available Alternatives or Decision Outcome" of the one
// Decision whose Basis carries it. It is not reused for a future
// Delegation's "applicable constraints" (PEOS-004:833): a delegation's
// constraints would limit a granted authority across all of that
// grantee's future Decisions -- a different axis, the same distinction
// this package already draws between Authority.requirements and
// Authority.bases.
//
// None of Assumption, Constraint, or Uncertainty carries identity, a Ref,
// a revision, or a lifecycle: PEOS-006's Claim Basis ("does not introduce
// independent Claim Basis identity, revision, or lifecycle") and
// PEOS-007's Quality Constraint ("value structures without independent
// identity, revision, or lifecycle") are the governing precedents.
//
// The Basis final-state validity guarantee -- that NewBasis, NewBasisFrom,
// every collection With* mutator, and UnmarshalJSON return a fully valid
// Basis whenever they return a nil error -- does not extend to the
// pre-existing, no-error-return WithExtension method: calling
// Basis{}.WithExtension(ext) on the zero Basis yields a value that still
// reports IsZero()==true. Extension alone is Product-specific data, not
// Decision Basis content in the PEOS-004 sense, so it does not by itself
// satisfy the "at least one of Evidence, Assumptions, Constraints, or
// Uncertainties" requirement. This is a pre-existing Packet F API edge,
// unchanged by Packet F.2.
//
// # Decision Supersession and Decision Invalidation are dedicated immutable governance records
//
// DecisionSupersession and DecisionInvalidation are each independently
// identified, immutable governance records -- not Artifacts, not
// Lifecycle Subjects, not Artifact Revisions, not Decision Revisions (see
// above), and not peos/relation.Relation specializations. PEOS-004 lists
// "supersedes" and "invalidates" among its Decision Relation Types, but
// Relation carries no extent, no effective condition/time, and no reason,
// so it cannot express what these records must; a Product MAY
// additionally record a Relation as an index over either fact, but that
// Relation is never authoritative and this package does not import
// peos/relation to create one.
//
// # Supersession extent is closed
//
// SupersessionExtent is a closed complete/partial distinction, not an
// open vocabulary: PEOS-004's Extension Model does not list supersession
// completeness among what a Product MAY extend, and explicitly forbids
// extensions from redefining the core meaning of supersession. A partial
// extent carries its own required, deterministically identifiable
// remaining scope; a complete extent carries none.
//
// # Effective condition is normative text, not an executable predicate
//
// DecisionSupersession.EffectiveCondition and
// DecisionInvalidation.EffectiveCondition are plain normative condition
// statements. This package defines no condition language and no
// evaluator; evaluating a condition against runtime or repository state
// is a later, deliberately deferred concern.
//
// # Preservation is by reference, not by duplication
//
// Both records preserve the original Decision's Outcome and Applicability
// by naming the Decision with core.DecisionRef, relying on Decision's own
// constructor-only Outcome and Applicability fields (see above) rather
// than copying their content. This package does not resolve or verify
// that reference against any repository: the guarantee that the named
// Decision is still the one originally superseded or invalidated is a
// repository-layer obligation this package does not check, the same
// limitation peos/lifecycle's own Supersession record documents for its
// superseding/superseded references.
//
// # Invalidation is non-retroactive by default
//
// DecisionInvalidation carries no field asserting that the invalidated
// Decision was never applicable. Its absence means exactly that:
// non-retroactive. PEOS-004 permits retroactive effect only when an
// applicable Product contract explicitly defines it; this package does
// not model that profile.
//
// # InvalidationSource chooses one canonical source
//
// PEOS-004 requires Invalidation to "identify the invalidating authority
// or Decision," and that "or" is not stated as an exclusive or in the
// specification text. InvalidationSource nonetheless accepts exactly one
// of the two, as a deliberate SDK architectural decision, not a
// specification requirement: when the source is a Decision, that
// Decision's own Authority is already required and non-zero, so carrying
// a second, independent AuthorityRef alongside it would duplicate that
// authority with no way for this package to detect a divergence between
// the two.
//
// # Decision Invalidation is not core.RecordCorrection
//
// Invalidation is a governance act -- a withdrawal of normative effect --
// never a typo correction, a representation fix, a metadata correction,
// or a non-normative clarification. core.RecordRef's own correction union
// deliberately excludes DecisionID; this package does not reuse
// core.RecordCorrection or core.CorrectionKindInvalidate for Decision
// Invalidation.
//
// # Decision Conflict is binary and independently identified
//
// DecisionConflict represents PEOS-004's Decision Conflict ("A Decision
// Conflict exists when two applicable Decisions establish incompatible
// normative intent within overlapping Applicability") as a binary
// relationship, not an N-ary one: :1117's own definitional clause is
// "when two applicable Decisions ...". A jointly-unsatisfiable triple of
// Decisions is three pairwise conflicts, each independently identified;
// PEOS-004 defines no N-ary conflict concept for this package to model.
//
// DecisionConflict carries its own ConflictID, not a natural key derived
// from the two conflicting Decisions and their overlapping scope: the
// same pair of Decisions MAY conflict in the same scope over two
// materially different incompatible Outcomes (:1123), so the natural key
// is insufficient. Independent identity is also what the Conflict
// Invariant (:1342, "Conflicting applicable Decisions require explicit
// resolution") requires structurally: the invariant presupposes a
// conflict that exists while unresolved, and a conflict recorded only
// inside its own resolution could never be observed as unresolved.
//
// governingRule is a required field, not optional, and this is a
// deliberate boundary: it is what makes DecisionConflict the analyzed
// conflict :1121 defines (all four of its MUSTs satisfied), as distinct
// from a Runtime's raw pre-analysis detection (:1360's permitted "detect
// conflicts" activity). A raw detection is represented as a
// peos/relation.Relation carrying core.RelationTypeConflict -- this
// package does not import peos/relation to create one, for the same
// reason DecisionSupersession does not: Relation cannot carry :1123 or
// :1124, has no identity for a resolution to reference, and does not
// reject source == target (see DecisionSupersession's own doc comment).
//
// # Conflict Resolution is a separate, optional-linked record
//
// ConflictResolution records that a specific DecisionConflict has been
// closed (:1342; :1128's six mechanisms). It is a separate immutable
// value, never a mutation of DecisionConflict: there is no "resolved
// bool" or "status" field anywhere in this package's Conflict model, so
// a DecisionConflict without a corresponding ConflictResolution is, by
// construction, unresolved.
//
// ResolutionMechanism is an open core.VocabularyValue wrapper with six
// predeclared PEOS-namespace constants drawn directly from :1128. This
// is a deliberate departure from the closed, two-variant
// SupersessionExtent: PEOS-004's Extension Model (:1398-1411) explicitly
// names "conflict-resolution policies" among what a Product contract MAY
// define, and :1128's own sixth arm is itself "an applicable Product
// contract" -- the governing distinction (see SupersessionExtent's own
// doc comment) is whether the Extension Model lists the axis, and here
// it does.
//
// resolvingDecision is a plain optional field, not conditionally
// required by mechanism. An earlier design required it exactly when
// mechanism was the Decision or Supersession constant; that rule was
// rejected during adversarial review because ResolutionMechanism is
// open -- a Product-defined mechanism value that is, in substance,
// decision-based would silently escape a rule keyed to two hardcoded
// predeclared constants, producing an invariant that presents as
// complete but is not. The identification burden instead rests entirely
// on the required statement field, consistent with :1128's six
// mechanisms being categories, not identifications.
//
// # Decision Roles use an open RoleKind vocabulary
//
// Role associates an identified core.ActorRef with a single Decision
// role (:768's "Decision Author; Decision Proposer; Decision Maker;
// Decision Approver; Decision Reviewer; Decision Executor; Decision
// Recorder; Decision Owner"). RoleKind is modeled open, with those eight
// values predeclared, because :781 states outright "The applicable
// Product contract MAY define additional roles" and the Extension Model
// separately lists "additional Decision roles" -- unlike Constraint's
// origin list and Uncertainty's concern list (Packet F.2), which carry
// no such explicit extensibility permission and are therefore left
// unmodeled entirely rather than opened.
//
// Requiring an explicit core.ActorRef on every Role is what satisfies
// :785 ("Role identity MUST NOT be inferred solely from document
// authorship or repository ownership"): no field on Role can be
// populated implicitly from a Decision Record's own authorship. Role is
// not Authority: :751 requires Decision Authority to stay distinguishable
// from authorship, facilitation, recommendation, implementation, and
// documentation, and Role's vocabulary names exactly those
// non-authority participations. Role is also not core.Provenance.Actor,
// which records who produced a record -- one actor -- not who held a
// role in the Decision, of which there MAY be many (:783).
//
// Decision Roles and Decision Maker/Approver are modeled at SHOULD /
// conditional-MUST strength, not unconditional MUST: :815's "the
// approving Actor MUST be identifiable" is conditional on approval being
// required, and Roles are absent from both the Conformance list
// (:1431-1443) and the fourteen Decision Invariants (:1284-1342).
// Separation-of-duties enforcement (:845-855, "A Product contract MAY
// require...") is a Product/Runtime responsibility this package does not
// evaluate -- see "Runtime and Product boundaries" below.
//
// # Decision Consequence is distinct from Engineering Commitment
//
// Consequence is modeled at SHOULD strength: :316 says only that a
// significant Decision "SHOULD have ... identified consequences", and
// :874 says a Decision Record "SHOULD identify ... Consequences".
// Consequence is absent from both the Conformance list and the fourteen
// Decision Invariants, unlike Basis's Assumption and Uncertainty, which
// this package models at :456/:542's conditional MUST strength (see
// Basis's own doc comment). It is nonetheless modeled, at that lower
// strength, for the same reason Alternative (SHOULD, :353) and rationale
// (MAY, :518) are modeled: it is a named PEOS-004 "# Scope" entry (:54)
// with its own dedicated section, not illustrative prose.
//
// :697's "Consequences MUST be distinguishable from completed effects"
// is not a requirement that Consequence be separately representable; it
// is a distinguishability constraint against a completion concept this
// package does not model at all, so the MUST holds vacuously today, and
// continues to hold structurally: Consequence carries no completion or
// status field, so a Consequence value can never be mistaken for a
// completed effect.
//
// Consequence is not Commitment, despite :620 and :683's "MAY include"
// lists sharing several examples nearly verbatim (required Lifecycle
// Transitions / require a Lifecycle Transition; accepted risks / accept
// a defined risk; review obligations / impose a review condition;
// validation work / establish a validation expectation; operational
// changes / establish an operational obligation; implementation work /
// establish an implementation obligation; new constraints / constrain
// future Artifacts or Revisions). That overlap is of illustrative
// examples, not of definitions: :616 defines Engineering Commitment as
// normative intent established, changed, or removed by an established
// Decision Outcome; :681 defines Decision Consequence as an expected,
// required, permitted, or accepted effect of that Outcome. :681
// explicitly admits predictions -- expected effects that are not
// themselves normative intent. Encoding such a prediction as a
// Commitment would misrepresent it as established normative intent and
// would violate :1413's prohibition on extensions that "redefine the
// core meaning of ... Engineering Commitment." Consequence therefore
// carries no CommitmentEffect-equivalent verb, unlike Commitment, and
// hangs off Decision rather than Outcome, following :316 and :701 ("A
// Decision MAY be applicable even when its Consequences have not yet
// been completed").
//
// # Runtime and Product boundaries
//
// This package does not evaluate scope overlap, priority, or authority
// to determine whether two Decisions actually conflict (:1121's own
// analysis), does not detect conflicts (:1360's permitted Runtime
// activity), does not evaluate separation-of-duties rules (:845-855),
// and does not define a significance classification (:300, "A Product
// contract MAY define significance levels") or a Decision Context
// representation beyond what core.Representation on a Decision Record
// Revision already provides (:275-292, all MAY). These remain Runtime or
// Product-contract responsibilities this package's value types support
// but do not perform.
//
// # Delegation is deferred to a future authority-grant packet
//
// Decision Authority delegation (:824-839: "A delegation MUST identify:
// the delegating Actor or authority; the receiving Actor; the delegated
// scope; applicable constraints; the effective period or condition;
// revocation conditions when applicable") is not implemented by this
// package, on architectural grounds rather than for lack of a
// representable shape.
//
// Every standalone aggregate this package defines names at least one
// Decision -- Record, DecisionSupersession, DecisionInvalidation,
// DecisionConflict -- and every non-standalone type (Authority, Basis,
// Outcome, Alternative, Commitment, Role, Consequence) is embedded
// inside Decision. A delegation record would be neither: it grants
// authority, after which the grantee makes arbitrarily many future
// Decisions unrelated to any single one this package could name.
//
// Authority is demonstrably cross-cutting in this repository, not
// Decision-local: peos/lifecycle already consumes core.AuthorityRef (in
// StateAssignment and TransitionRecordContent) and does not import
// peos/decision. PEOS-003 :627 ("Authority delegated to a Runtime MUST
// be explicit or determinable through an applicable contract") means
// the moment a lifecycle record made under delegated authority must name
// its delegation, peos/lifecycle would need to import peos/decision --
// inverting dependency direction, since lifecycle is the more
// foundational model -- or duplicate the record. peos/core also
// explicitly declines to define an Authority aggregate today
// (core.AuthorityRef's own doc comment: "this package does not define
// an Authority aggregate, does not give AuthorityRef a lifecycle ...
// this package stops at the reference"); a delegation record with
// identity, scope, effective period, constraints, and revocation
// conditions is exactly such an aggregate, and building it inside
// peos/decision would pre-empt that open question in the wrong package.
//
// Decision-side conformance does not require Delegation: :743 lists
// "delegated authority" among the nine MAY-origins of an authority
// basis, core.AuthorityRef already carries an open Kind for that
// purpose, the Authority Invariant (:1306) is met today by Authority,
// and Delegation is absent from both the Conformance list and all
// fourteen Decision Invariants.
//
// The missing type is an authority-grant record: identity; delegator
// (core.ActorRef or core.AuthorityRef); delegate core.ActorRef;
// delegated scope; effective period or condition; applicable
// constraints; revocation conditions. A Product-defined Extension is not
// a substitute: :828 imposes six MUST-identify items an opaque blob
// cannot enforce, and an extension cannot be referenced by
// peos/lifecycle's own authority fields. The owning future packet must
// first decide which package owns authority-grant aggregates -- peos/core
// (admitting the aggregate it currently declines) or a new sibling
// package dedicated to governance grants -- and whether core.AuthorityRef
// gains a link to the delegation that produced it. It cannot be
// peos/decision, because peos/lifecycle must be able to reach the record
// without importing peos/decision.
//
// # Deferred
//
// Delegation (see above), Decision relation-type constants, a decision
// repository, a supersession or invalidation history resolver, condition
// evaluation, conflict detection and scope-overlap evaluation, and any
// approval or voting workflow are not implemented by this package.
package decision
