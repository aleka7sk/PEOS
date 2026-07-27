package runtime

import (
	"fmt"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/validation"
)

// This file implements the PEOS-008 Compliance Claim construction helper.
//
// "A Compliance Claim is a specialization of Validation Claim, as defined
// by PEOS-006... This specification does not define a second
// runtime-specific Claim base model. A Compliance Claim exists exclusively
// as an instance of the PEOS-006 Validation Claim mechanism." There is
// therefore no ComplianceClaim type in this package -- only a constructor
// helper that delegates to validation.NewClaim with the Claim Type fixed
// to core.ClaimTypeCompliance. Every immutability, identity, correction,
// and derived-effect guarantee validation.Claim already provides applies
// unchanged; this package adds none of its own.
//
// This is the one file in the package that imports peos/validation, and it
// exists only for this helper -- no other J.2 type composes or references
// validation.Claim.

// NewComplianceClaim validates its arguments and returns a validation.Claim
// whose Claim Type is fixed to core.ClaimTypeCompliance.
//
// This is a thin delegation to validation.NewClaim, not a new construct:
// it supplies core.ClaimTypeCompliance as validation.NewClaim's claimType
// argument and otherwise changes nothing about that constructor's
// contract. Every validation.Claim invariant -- exactly one Subject,
// criteria separated from Subject, at least one Evidence citation,
// immutability, identity, and the correction/replacement/invalidation
// model -- is enforced by validation.NewClaim itself and is not
// re-implemented or weakened here.
//
// subject MAY be a runtime subject, a Runtime Contract Artifact or exact
// Revision, a deployed Artifact Revision, or any other explicitly
// identified engineering subject core.EngineeringSubjectRef can carry --
// PEOS-008 permits all of these as a Compliance Claim's Subject. criteria
// MAY include a Requirement, a Requirement Artifact Revision, a Runtime
// Contract rule, a Runtime Assertion, or a Quality Characteristic or
// Measure, all already expressible through existing core.CriterionRef
// arms; PEOS-008 additionally permits "applicable Waiver conditions" as
// criteria, which this package does not give a dedicated CriterionRef arm
// (see doc.go's RJ-1 note on Waiver's lack of independent identity).
func NewComplianceClaim(
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
	claim, err := validation.NewClaim(id, core.ClaimTypeCompliance, subject, scope, outcome, method, criteria, evidence, timestamp, provenance)
	if err != nil {
		return validation.Claim{}, fmt.Errorf("runtime: NewComplianceClaim: %w", err)
	}
	return claim, nil
}
