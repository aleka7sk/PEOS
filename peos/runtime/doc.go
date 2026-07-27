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
//
//	BindingRecord                 immutable record, not an Artifact       (Packet J.2)
//	UnbindingRecord                immutable record, not an Artifact       (Packet J.2)
//	Observation                    immutable record, not an Artifact       (Packet J.2)
//	Violation                      immutable record, not an Artifact       (Packet J.2)
//	Compliance Claim               core.ClaimTypeCompliance value on validation.Claim (Packet J.2)
//
//	Current Runtime Binding        derived view, never stored              (repository-owned)
//	Runtime Compliance             derived view, never stored              (repository-owned)
//
// Contract is the only type in this package with a PEOS identity, and that
// identity is an ordinary core.ArtifactID -- there is no Runtime Contract
// Version, and no independent identity for a Runtime Requirement
// Reference or a Runtime Assertion.
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
// separate immutable record (Packet J.2), never Revision content, and
// creating a new Contract Revision does not, by itself, change which
// Revision is currently bound to a runtime subject.
//
// # The one explicit cardinality minimum
//
// PEOS-008 states exactly one minimum for Contract content: "A Runtime
// Contract Revision SHALL reference one or more Requirements or
// Requirement Artifact Revisions that it governs." NewContractContent and
// every collection modifier enforce len(requirements) >= 1 through the
// single shared validation path validateContractContent.
//
// No other collection has any minimum. Assertions, observation
// requirements, violation classification rules, Waiver handling rules,
// enforcement expectations, and applicable Quality Profile Revisions may
// all be empty -- PEOS-008 lists each without a cardinality qualifier
// (Quality Profile Revisions are explicitly "where required"), the same
// unqualified form this repository already reads as permitting emptiness
// for validation.PlannedActivity's own "criteria", "Evidence expected",
// and "execution prerequisites" items, and for quality.ProfileContent's
// seven owned-value collections after Packet I.3.B's correction.
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
// Assertion is the one owned-value collection this package keys, using
// core.LocalKey. Keys are unique within the Assertions collection;
// PEOS-008 states no key uniqueness rule at all, so this is a minimal
// derived rule -- the same derivation quality.addProfileLocalKey documents
// for its own per-kind namespaces: a criterion citing an Assertion by key
// must resolve to exactly one Assertion. This package does not perform
// repository resolution of an Assertion's core.CriterionRef: whether the
// Requirement, Runtime Contract rule, Quality element, or external rule it
// names actually exists is repository-owned.
//
// # No execution or invocation record
//
// PEOS-008 defines no runtime invocation, execution request, or execution
// result. Actual runtime execution, scheduling, retries, and timeouts are
// entirely Product-owned and out of scope. This package therefore does not
// reuse validation.ExecutionRecord for anything: a future Runtime
// Observation (Packet J.2) records a runtime measurement or event, not the
// execution of a planned validation activity, and conflating the two would
// assert an ontology PEOS-008 does not state.
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
//   - observation requirements, violation classification rules, Waiver
//     handling rules, and enforcement expectations -- all opaque, trimmed
//     string descriptions;
//   - core.Scope's expression, in a Contract's deployment scope, an
//     Assertion's scope, or applicability.
//
// Each of those is Product-owned. Encoding any of them would be a
// framework PEOS-008 deliberately does not define.
//
// # Package dependency boundary
//
// During Packet J.1, production sources may import only the standard
// library and peos/core. peos/validation is not imported in J.1, because
// Compliance Claim construction -- the one PEOS-008 concept that touches
// PEOS-006's mechanism -- is Packet J.2's, not this file's.
// peos/relation, peos/lifecycle, peos/requirement, peos/decision, and
// peos/quality are all excluded, each for a reason stated above. Nothing
// imports peos/runtime: it is a leaf.
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
//   - evaluating an Assertion's criterion against whatever it names, and
//     evaluating PEOS-005 Waiver applicability (scope, authority,
//     temporal, and -- where PEOS-005 comes to define them -- attached
//     conditions);
//   - actual runtime execution, deployment, scheduling, and monitoring,
//     all Product-owned and out of scope for this package.
//
// # Packet scope
//
// Packet J.1 (this update) implemented the Runtime Contract Artifact
// foundation: ArtifactTypeRuntimeContract, Contract, ContractApplicability,
// RequirementReference, ContractContent, ContractRevision, Assertion, and
// the three runtime-local vocabulary wrappers (Environment,
// ViolationClassification, ViolationSeverity), plus the additive
// peos/core change described above.
//
// Packet J.2 will add BindingRecord, UnbindingRecord, Observation,
// Violation, and the Compliance Claim construction helper (which returns a
// validation.Claim with core.ClaimTypeCompliance -- there is no dedicated
// ComplianceClaim type, because PEOS-008 imposes no rule on a Compliance
// Claim beyond what PEOS-006 already enforces). J.1 declares four
// aggregate sentinels reserved for that work
// (ErrInvalidRuntimeBindingRecord, ErrInvalidRuntimeUnbindingRecord,
// ErrInvalidRuntimeObservation, ErrInvalidRuntimeViolation) so J.2 does not
// have to reopen errors.go, following the convention Packet H.1/I.1
// established.
//
// No PEOS-008 packet has an accepted audit verdict yet; Packet J.3 is the
// consolidated audit and Packet J.4 the closure.
package runtime
