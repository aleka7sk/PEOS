package template

import (
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// This file implements the PEOS-009 Template Conformance Claim construction
// helper.
//
// "A Template Conformance Claim is a specialization of the Conformance Claim
// defined by PEOS-006. A Template Conformance Claim inherits, without
// redefinition, all Validation Claim rules defined by PEOS-006." PEOS-009 adds
// that it "is the sole Claim type this specification defines... This
// specification does not define an independent Activity, Evidence, or Claim
// base mechanism."
//
// There is therefore no TemplateConformanceClaim type in this package -- only
// a constructor helper that delegates to validation.NewClaim with the Claim
// Type fixed to core.ClaimTypeTemplateConformance. Defining a wrapper type
// would risk exactly what PEOS-009 names as a non-conforming pattern:
// "Defining a Template Conformance Claim that does not specialize PEOS-006's
// Conformance Claim, or that redefines Claim identity, immutability, or
// replacement semantics." Representing one as an Artifact is a second named
// non-conforming pattern, and returning validation.Claim -- which is not an
// Artifact -- makes that unrepresentable too.
//
// This is the one file in the package that imports peos/validation, and it
// exists only for this helper. This mirrors runtime/claim.go exactly, which
// does the same for PEOS-008's Compliance Claim.

// NewTemplateConformanceClaim validates its arguments and returns a
// validation.Claim whose Claim Type is fixed to
// core.ClaimTypeTemplateConformance.
//
// This is a thin delegation to validation.NewClaim, not a new construct: it
// supplies core.ClaimTypeTemplateConformance as validation.NewClaim's
// claimType argument and otherwise changes nothing about that constructor's
// contract. Every validation.Claim invariant -- exactly one Subject, criteria
// separated from Subject, at least one Evidence citation, immutability,
// identity, and the correction/replacement/invalidation model -- is enforced
// by validation.NewClaim itself and is not re-implemented or weakened here.
//
// PEOS-009's "Single Template Conformance Subject Invariant" ("Every Template
// Conformance Claim identifies exactly one Subject") needs no enforcement in
// this package: validation.Claim already takes exactly one
// core.EngineeringSubjectRef, so the named non-conforming pattern "Composite
// Template Conformance Subject" is structurally unrepresentable.
//
// subject SHALL be exactly one of: a generated Artifact; a generated Artifact
// Revision; a Template Artifact; a Template Artifact Revision; or another
// explicitly permitted engineering subject. All five are already expressible
// through existing core.EngineeringSubjectRef arms
// (EngineeringSubjectRefFromGeneratedArtifact,
// ...FromGeneratedArtifactRevision, ...FromTemplate, ...FromTemplateRevision,
// and NewOpaqueEngineeringSubjectRef), so this helper constrains the subject
// no further than PEOS-006 already does -- narrowing it here would reject the
// "another explicitly permitted engineering subject" arm PEOS-009 allows.
//
// criteria MAY include a Template Artifact Revision, parameter constraints, a
// generated Artifact Type rule, a representation rule, a compatibility rule, an
// applicable Product contract, or a Requirement Artifact Revision. A parameter
// constraint is cited through core.CriterionKindTemplateConstraint, whose
// payload resolves against TemplateContent.Constraint(key) -- see doc.go. Note
// that "A Template Artifact Revision used as a criterion is not a second
// Subject", which validation.Claim's own separation of subject from criteria
// already guarantees.
//
// # The one rule this helper adds
//
// criteria must contain at least one criterion. This is the single rule the
// helper enforces beyond validation.NewClaim, and it is inherited rather than
// invented: PEOS-006 requires a Conformance Claim to cite at least one
// criterion, and PEOS-009 states that "A Template Conformance Claim is a
// specialization of the Conformance Claim defined by PEOS-006" which
// "inherits, without redefinition, all Validation Claim rules defined by
// PEOS-006."
//
// validation.NewClaim does not apply that rule here on its own, because it
// keys the check on core.ClaimTypeConformance exactly and its own
// checkClaimTypeCriteria documents the reason: "PEOS-007/008/009 may add their
// own rules in their own packets; none is inferred here." This is PEOS-009's
// packet, and this is its rule. Omitting it would produce a Template
// Conformance Claim that does not behave as a Conformance Claim -- precisely
// the "Parallel Template Claim Base" non-conforming pattern.
//
// The error surfaces validation.ErrInvalidConformanceClaim, the same sentinel
// peos/validation raises for the rule this one mirrors, so a caller need not
// distinguish where the Claim came from to handle it.
//
// # This helper computes nothing about conformance
//
// "Template conformance is a derived view. It is not a stored field", derived
// by a repository from applicable Claims, the exact Template Artifact
// Revision, the applicable Application Record, the generated Artifact
// Revision, criteria, Evidence, Claim correction history, scope, authority,
// and governing Product rules.
func NewTemplateConformanceClaim(
	id core.ValidationClaimID,
	subject core.EngineeringSubjectRef,
	scope core.Scope,
	outcome core.ClaimOutcome,
	method core.ValidationMethod,
	criteria []core.CriterionRef,
	evidence []core.EvidenceArtifactRevisionRef,
	timestamp core.Timestamp,
	provenance core.Provenance,
) (validation.Claim, error) {
	if len(criteria) == 0 {
		return validation.Claim{}, fmt.Errorf("template: NewTemplateConformanceClaim: %w: a template conformance claim requires at least one criterion", validation.ErrInvalidConformanceClaim)
	}
	claim, err := validation.NewClaim(id, core.ClaimTypeTemplateConformance, subject, scope, outcome, method, criteria, evidence, timestamp, provenance)
	if err != nil {
		return validation.Claim{}, fmt.Errorf("template: NewTemplateConformanceClaim: %w", err)
	}
	return claim, nil
}
