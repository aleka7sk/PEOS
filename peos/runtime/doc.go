// Package runtime implements the PEOS-008 Runtime Contract model: how a set
// of Requirements is bound to a runtime subject for enforcement, and how
// binding, observation, violation, and compliance are recorded immutably
// and derived rather than stored.
//
// # What PEOS-008 is not
//
// PEOS-008 does not define a generic runtime interface, deployment
// manifest, orchestrator, scheduler, container platform, workflow engine,
// or infrastructure API. Its Non-Goals explicitly disclaim "a mandatory
// deployment technology, infrastructure platform, or monitoring tool" and
// "a specific monitoring, observability, or enforcement technology". This
// package accordingly defines no runtime input, output, dependency,
// capability, invocation, execution result, or provider/endpoint/secret
// model. Inventing any of those would assert an ontology the specification
// does not state -- see the PEOS-008 Runtime Contract Architecture
// Blueprint (Packet J.0) for the full analysis of why the generic
// runtime-interface framing does not fit this specification.
//
// # Ontology
//
// PEOS-008 introduces exactly one Artifact:
//
//	Runtime Contract              Artifact (PEOS-002), ordinary Artifact Revision
//	  ContractRevision             an Artifact Revision of that Artifact
//	    ContractContent            the Revision's typed declared content
//	      RequirementReference     Revision-owned value (two participant levels)
//	      Assertion                Revision-owned value
//	      ContractRule             Revision-owned value (four categories, one namespace)
//
//	BindingRecord                 immutable record, not an Artifact
//	UnbindingRecord                immutable record, not an Artifact
//	Observation                    immutable record, not an Artifact
//	ViolationTrigger               closed two-arm union (Observation | Evidence)
//	Violation                      immutable record, not an Artifact
//	Compliance Claim               core.ClaimTypeCompliance value on validation.Claim (no dedicated type)
//
//	Current Runtime Binding        derived view, never stored              (repository-owned)
//	Runtime Compliance             derived view, never stored              (repository-owned)
//
// Contract is the only type in this package with a PEOS identity, and that
// identity is an ordinary core.ArtifactID -- there is no Runtime Contract
// Version, and no independent identity for a Runtime Requirement
// Reference, a Runtime Assertion, or a Runtime Contract Rule.
//
// # Runtime Contract Artifact semantics
//
// A Runtime Contract "is an Artifact as defined by PEOS-002" and uses
// ordinary Artifact Revision for its own declared content evolution --
// there is no Runtime Contract Version distinct from Artifact Revision.
// Contract therefore wraps a core.Artifact and adds no field;
// ContractRevision wraps a core.ArtifactRevision and pairs it with
// ContractContent. Neither exposes a version, a lifecycle, a status, or a
// content setter: modifying declared content constitutes a content change
// and creates a new Artifact Revision.
//
// # Declared content versus runtime binding
//
// PEOS-008 draws this boundary explicitly: Contract existence, Revision
// publication, optional Lifecycle activation, a Runtime Binding Record,
// and actual deployment are five distinct, independently inspectable
// conditions, and none implies another. This package models only the
// first two (Contract, ContractRevision/ContractContent); binding is a
// separate immutable record (BindingRecord, UnbindingRecord), never
// Revision content, and creating a new Contract Revision does not, by
// itself, change which Revision is currently bound to a runtime subject.
//
// # The one explicit cardinality minimum
//
// PEOS-008 states exactly one minimum for Contract content: "A Runtime
// Contract Revision SHALL reference one or more Requirements or
// Requirement Artifact Revisions that it governs." NewContractContent and
// every collection modifier enforce len(requirements) >= 1 through the
// single shared validation path validateContractContent.
//
// No other collection has any minimum. Assertions, the four ContractRule
// categories (observation requirements, violation classification rules,
// Waiver handling rules, enforcement expectations), and applicable Quality
// Profile Revisions may all be empty -- PEOS-008 lists each without a
// cardinality qualifier (Quality Profile Revisions are explicitly "where
// required"), the same unqualified form this repository already reads as
// permitting emptiness for validation.PlannedActivity's own "criteria",
// "Evidence expected", and "execution prerequisites" items, and for
// quality.ProfileContent's seven owned-value collections after Packet
// I.3.B's correction. Packet J.2.A, which introduced ContractRule in place
// of the four categories' original opaque []string representation,
// introduced no new aggregate minimum either.
//
// # Exact Requirement participant level
//
// Every RequirementReference is exactly one of two arms -- Requirement
// identity (core.RequirementRef) or an exact Requirement Artifact Revision
// (core.RequirementArtifactRevisionRef) -- never neither, never both, and
// never a third "latest" arm. PEOS-008 requires this explicitly: "every
// reference... SHALL preserve the exact participant level being
// referenced... and SHALL NOT silently treat one as equivalent to the
// other." Requirement Statement content is never copied into a
// RequirementReference: its two arms carry no Statement field, so the
// prohibition is structural.
//
// # Runtime-local key namespace
//
// Runtime local-key namespaces are criterion-kind namespaces, not one
// namespace per Go collection. PEOS-008 gives a Runtime Contract Revision
// exactly two citable, locally-keyed collections of content, matching the
// two dedicated core.CriterionRef arms peos/core already carries --
// core.CriterionKindRuntimeAssertion and
// core.CriterionKindRuntimeContractRule, both taking a
// core.RuntimeRuleCriterionRef payload of (Revision, LocalKey) with no
// further discriminator -- and this package's key namespaces mirror that
// two-way split exactly:
//
//   - the Assertion namespace is exactly ContractContent.assertions;
//   - the Runtime Contract Rule namespace spans all four ContractRule
//     categories combined -- observation requirements, violation
//     classification rules, Waiver handling rules, and enforcement
//     expectations -- because core.RuntimeRuleCriterionRef's payload
//     carries no rule-category discriminator to tell the four categories
//     apart at the criterion level, so a key must resolve to at most one
//     ContractRule across all four collections together, not once per
//     collection.
//
// Uniqueness is enforced *within* each of these two namespaces; PEOS-008
// states no key uniqueness rule at all, so this is a minimal derived
// rule -- the same derivation quality.addProfileLocalKey documents for its
// own per-kind namespaces: a criterion citing an owned value by key must
// resolve to exactly one such value. The two namespaces are independent of
// each other, so the same key may be used once by an Assertion and once by
// a ContractRule without conflict (the criterion kind, not the key,
// determines which namespace is consulted) -- but a key may not repeat
// twice within the Assertion namespace, nor twice across any two of the
// four Contract Rule categories.
//
// This package does not perform repository resolution of an Assertion's
// or a ContractRule's core.CriterionRef: whether the Requirement, Runtime
// Contract rule, Quality element, or external rule it names actually
// exists is repository-owned. ContractContent.ContractRule(key) and
// ContractContent.Assertion(key) perform only local resolution within an
// already-loaded ContractContent.
//
// Before Packet J.2.A, the four Contract Rule categories were represented
// as opaque []string with no key of any kind, which left
// core.CriterionKindRuntimeContractRule -- already shipped in peos/core --
// with no resolvable target anywhere in this package (audit finding
// J3-03). ContractRule and the combined namespace above are the
// correction.
//
// # No execution or invocation record
//
// PEOS-008 defines no runtime invocation, execution request, or execution
// result. Actual runtime execution, scheduling, retries, and timeouts are
// entirely Product-owned and out of scope. This package therefore does not
// reuse validation.ExecutionRecord for anything: Observation records a
// runtime measurement or event, not the execution of a planned validation
// activity, and conflating the two would assert an ontology PEOS-008 does
// not state.
//
// # No lifecycle duplication
//
// Nothing in this package assigns, carries, or derives a Lifecycle State
// or State Assignment. A Runtime Contract's optional activation under
// PEOS-003 is governed exclusively by peos/lifecycle, which this package
// does not import. Runtime binding readiness is never a Lifecycle State,
// and a Runtime Binding Record is never a State Assignment.
//
// # No relation
//
// PEOS-008's Requirement references are Revision-owned content, not
// Artifact Relations: ":219" offers a choice between binary Artifact
// Relations and explicit Revision-owned references, and this package uses
// the latter, so it does not import peos/relation. There is no source,
// target, or relation field anywhere in this package.
//
// # No stored derived state
//
// ContractContent stores no current binding, no deployment status, and no
// compliance outcome. Current Runtime Binding and Runtime Compliance are
// PEOS-008 derived views, computed on demand by a repository from Runtime
// Binding, Unbinding, Observation, and Violation Records and Compliance
// Claims -- never a mutable field on a Runtime Contract Revision. PEOS-008
// names three forbidden fields by this exact pattern:
// RuntimeContract.bound, RuntimeContract.activeDeployment, and
// ArtifactRevision.deployed; none of them, nor any synonym, exists on any
// type in this package.
//
// # Product-owned interpretation
//
// This package predeclares no environment, no violation classification,
// no violation severity, and no evaluation rule language, and it
// interprets none of the following:
//
//   - an Environment's meaning (a Kubernetes namespace, a cloud region, a
//     deployment tier, or anything else);
//   - a runtime subject's identifier syntax (core.RuntimeSubjectRef is an
//     opaque, Product-defined namespaced identifier);
//   - an Assertion's evaluation rule and expected result -- both opaque,
//     trimmed strings, because PEOS-008 defines no expression language;
//   - a ContractRule's text, across all four categories (observation
//     requirements, violation classification rules, Waiver handling
//     rules, enforcement expectations) -- opaque, trimmed strings;
//   - core.Scope's expression, in a Contract's deployment scope, an
//     Assertion's scope, or applicability.
//
// Each of those is Product-owned. Encoding any of them would be a
// framework PEOS-008 deliberately does not define.
//
// # Package dependency boundary
//
// Production sources may import only the standard library, peos/core, and
// peos/validation. peos/validation is imported by exactly one file,
// claim.go, for the Compliance Claim construction helper -- no other type
// in this package composes or references validation.Claim. peos/relation,
// peos/lifecycle, peos/requirement, peos/decision, and peos/quality are
// all excluded, each for a reason stated above. Nothing imports
// peos/runtime: it is a leaf.
//
// # Core impact
//
// peos/core already carried nearly the entire PEOS-008 identity, reference,
// subject, record, and criterion layer before this package existed:
// RuntimeSubjectRef, RuntimeContractRef, RuntimeContractRevisionRef; the
// four Runtime*ID types; CriterionKindRuntimeContractRule and
// CriterionKindRuntimeAssertion with their RuntimeRuleCriterionRef payload;
// RecordKindRuntimeBindingRecord, RecordKindRuntimeUnbindingRecord,
// RecordKindRuntimeObservation, and RecordKindRuntimeViolation with their
// RecordRef arms; SubjectKindRuntimeSubject, SubjectKindRuntimeContract,
// and SubjectKindRuntimeContractRevision with their EngineeringSubjectRef
// arms; and core.ClaimTypeCompliance. Packet J.1 added exactly one
// additive core change: RuntimeBindingRecordRef, RuntimeUnbindingRecordRef,
// and RuntimeObservationRef, dedicated per-record reference types
// mirroring ValidationClaimRef/ValidationExecutionRecordRef, needed
// wherever a PEOS-008 record must reference another record of a
// compile-time-fixed family (an Unbinding Record's mandatory "exactly one
// Binding Record" reference, or a Binding/Unbinding Record's own optional
// correction reference). RuntimeViolationRef was deliberately not added:
// nothing in PEOS-008's value layer references a Violation by exact type.
//
// # Repository responsibilities
//
// This package models values and enforces PEOS-008's structural
// invariants. It does not persist, index, query, or resolve anything
// across Contract Revisions or against runtime state. A repository built
// on it owns:
//
//   - storing Contracts, Contract Revisions, Binding/Unbinding Records,
//     Observations, Violations, and Compliance Claims, and retrieving them
//     by identity;
//   - deriving Current Runtime Binding from Binding and Unbinding Record
//     history, and Runtime Compliance from applicable, non-replaced,
//     non-invalidated Compliance Claims, Violations, and Waivers;
//   - evaluating an Assertion's or a ContractRule's criterion against
//     whatever it names, and evaluating PEOS-005 Waiver applicability
//     (scope, authority, temporal, and -- where PEOS-005 comes to define
//     them -- attached conditions) as an input to derived Runtime
//     Compliance -- this package stores no Waiver reference of any kind
//     (see RJ-1 below);
//   - actual runtime execution, deployment, scheduling, and monitoring,
//     all Product-owned and out of scope for this package.
//
// # Packet scope
//
// Packet J.1 implemented the Runtime Contract Artifact foundation:
// ArtifactTypeRuntimeContract, Contract, ContractApplicability,
// RequirementReference, ContractContent, ContractRevision, Assertion, and
// the three runtime-local vocabulary wrappers (Environment,
// ViolationClassification, ViolationSeverity), plus the additive
// peos/core change described above.
//
// Packet J.2 added the four immutable enforcement records and the
// Compliance Claim helper: BindingRecord and UnbindingRecord
// (correction-bearing, via core.RecordCorrectionRef[core.RuntimeBindingRecordRef]
// and core.RecordCorrectionRef[core.RuntimeUnbindingRecordRef]
// respectively); Observation (no correction reference; explicit,
// non-automatic Evidence citation via []core.EvidenceArtifactRevisionRef);
// ViolationTrigger (a closed two-arm union naming the exact Observation or
// Evidence that triggered a Violation) and Violation itself (no correction
// reference); and NewComplianceClaim, which delegates to
// validation.NewClaim with the Claim Type fixed to core.ClaimTypeCompliance
// and returns an ordinary validation.Claim -- no ComplianceClaim type
// exists, because PEOS-008 imposes no rule on a Compliance Claim beyond
// what PEOS-006 already enforces. All four sentinels J.1 reserved for this
// work (ErrInvalidRuntimeBindingRecord, ErrInvalidRuntimeUnbindingRecord,
// ErrInvalidRuntimeObservation, ErrInvalidRuntimeViolation) became active.
//
// Packet J.3's independent, read-only Consolidated Audit returned AUDIT
// PASS WITH CORRECTIVE PACKET REQUIRED: 0 BLOCKER, 3 MAJOR, 2 MINOR, 5
// NOTE. Packet J.2.A (this update) implemented the three MAJOR and two
// MINOR corrections:
//
//   - J3-01: Observation had no scope field at all, though PEOS-008 ":236"
//     requires every Runtime Binding Record, Runtime Observation, Runtime
//     Violation, and Compliance Claim to identify "its exact runtime
//     subject and subject scope". Observation now takes a mandatory
//     core.Scope constructor argument and exposes Scope().
//   - J3-02: Observation's environment was optional though PEOS-008
//     ":384" states "environment or context" unqualified, in the same
//     SHALL-identify list as two genuinely conditional items, and
//     BindingRecord's identically-unqualified environment (":273") is
//     mandatory in this same package. Environment is now a mandatory
//     Observation constructor argument; WithEnvironment and
//     WithoutEnvironment no longer exist.
//   - J3-03: core.CriterionKindRuntimeContractRule had no resolvable
//     target, because none of the four Contract Rule categories were
//     keyed. ContractRule (a new Revision-owned value, with its own
//     sentinel ErrInvalidRuntimeContractRule) now carries a core.LocalKey,
//     and all four categories share one combined criterion-kind namespace
//     -- see "Runtime-local key namespace" above.
//   - J3-04 (RJ-1): Violation.applicableWaiver -- an opaque, unresolvable
//     description standing in for an "identify" obligation PEOS-005's
//     Waiver has no identity to satisfy -- has been removed outright, not
//     renamed or replaced. Waiver applicability remains exactly what
//     PEOS-008 ":531" says it is: an input to derived Runtime Compliance,
//     computed by a repository, never a stored value-layer field. This
//     does not reduce PEOS-008 conformance: the removed field never
//     satisfied ":429" as an identification in the first place.
//   - J3-05: the tracker previously recorded J.1 and J.2 as "uncommitted"
//     with no commit hash; both are now recorded with their exact hashes.
//
// RJ-2 (whether Contract Revision authority, unqualified at ":196", is
// genuinely mandatory) was independently re-derived by Packet J.3 and
// closed in favor of the shipped behavior: PEOS-008 itself qualifies
// authority "where required" for both Binding Record (":278") and
// Unbinding Record (":312"), so the Contract Revision's own unqualified
// authority is a deliberate intra-document contrast, not an oversight.
// Authority remains mandatory; no change was made.
//
// Packet J.3.A (a focused re-audit of exactly these five corrections) is
// the next step; Packet J.4 is the closure.
package runtime
