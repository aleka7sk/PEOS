package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- LifecycleConsequence ------------------------------------------------

func TestNewLifecycleConsequenceValid(t *testing.T) {
	c, err := NewLifecycleConsequence("Superseded within the EU scope.")
	if err != nil {
		t.Fatal(err)
	}
	if c.IsZero() {
		t.Error("valid LifecycleConsequence IsZero() = true")
	}
	if !c.IsIdentified() {
		t.Error("IsIdentified() = false")
	}
	if c.IsNone() {
		t.Error("IsNone() = true")
	}
	if c.Kind() != "identified" {
		t.Errorf("Kind() = %q, want %q", c.Kind(), "identified")
	}
}

func TestLifecycleConsequenceKindNone(t *testing.T) {
	if got := NoLifecycleConsequence().Kind(); got != "none" {
		t.Errorf("Kind() = %q, want %q", got, "none")
	}
}

func TestNewLifecycleConsequenceEmptyRejected(t *testing.T) {
	_, err := NewLifecycleConsequence("")
	if !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestNewLifecycleConsequenceWhitespaceOnlyRejected(t *testing.T) {
	_, err := NewLifecycleConsequence("   ")
	if !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestNewLifecycleConsequenceDescriptionStoredTrimmed(t *testing.T) {
	c, err := NewLifecycleConsequence("  padded  ")
	if err != nil {
		t.Fatal(err)
	}
	desc, ok := c.Description()
	if !ok || desc != "padded" {
		t.Errorf("Description() = (%q,%v), want (%q,true)", desc, ok, "padded")
	}
}

func TestNoLifecycleConsequenceValidAndNotZero(t *testing.T) {
	c := NoLifecycleConsequence()
	if c.IsZero() {
		t.Error("NoLifecycleConsequence() IsZero() = true, want false")
	}
	if !c.IsNone() {
		t.Error("IsNone() = false")
	}
	if c.IsIdentified() {
		t.Error("IsIdentified() = true")
	}
	if _, ok := c.Description(); ok {
		t.Error("Description() ok = true on the none arm")
	}
}

func TestLifecycleConsequenceZeroValueIsZero(t *testing.T) {
	var c LifecycleConsequence
	if !c.IsZero() {
		t.Error("zero LifecycleConsequence IsZero() = false")
	}
	if c.IsIdentified() || c.IsNone() {
		t.Error("zero LifecycleConsequence matches a known arm")
	}
}

func TestLifecycleConsequenceJSONIdentifiedRoundTrip(t *testing.T) {
	c, err := NewLifecycleConsequence("Superseded within the EU scope.")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LifecycleConsequence
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, _ := decoded.Description()
	want, _ := c.Description()
	if got != want {
		t.Errorf("round trip mismatch: got %q, want %q", got, want)
	}
}

func TestLifecycleConsequenceJSONNoneRoundTrip(t *testing.T) {
	c := NoLifecycleConsequence()
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["description"]; present {
		t.Error(`"description" present on the none arm`)
	}
	var decoded LifecycleConsequence
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.IsNone() {
		t.Error("round trip lost the none arm")
	}
}

func TestLifecycleConsequenceJSONUnrecognizedKindRejected(t *testing.T) {
	var c LifecycleConsequence
	payload := `{"kind":"unknown"}`
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestLifecycleConsequenceJSONNoneWithDescriptionRejected(t *testing.T) {
	var c LifecycleConsequence
	payload := `{"kind":"none","description":"unexpected"}`
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestLifecycleConsequenceJSONIdentifiedMissingDescriptionRejected(t *testing.T) {
	var c LifecycleConsequence
	payload := `{"kind":"identified"}`
	if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestLifecycleConsequenceZeroMarshalRejected(t *testing.T) {
	var c LifecycleConsequence
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestLifecycleConsequenceUnmarshalFailurePreservesReceiver(t *testing.T) {
	original, err := NewLifecycleConsequence("original")
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{"kind":"identified"}`), &receiver); err == nil {
		t.Fatal("missing description accepted, want error")
	}
	if receiver != original {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- Supersession helpers --------------------------------------------------

func mustSupersession(t *testing.T) Supersession {
	t.Helper()
	s, err := NewSupersession(
		mustRequirementRevisionRef(t, "REQ-2", "REV-1"),
		mustRequirementRevisionRef(t, "REQ-1", "REV-3"),
		mustProvenance(t),
		mustScope(t, "product-x", "/services/*"),
		mustGovernanceActionFromDecisionOutcome(t, "DEC-1"),
		NoLifecycleConsequence(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func fullSupersession(t *testing.T) Supersession {
	t.Helper()
	s := mustSupersession(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	return s.WithExtension(ext)
}

func supersessionParticipantJSON(kind, artifactID, revisionID string) string {
	if kind == "requirement" {
		return `{"kind":"requirement","ref":{"artifact_id":"` + artifactID + `"}}`
	}
	return `{"kind":"requirement_revision","ref":{"artifact_id":"` + artifactID + `","revision_id":"` + revisionID + `"}}`
}

func supersessionPayload(t *testing.T, sourceKind, sourceArtifact, sourceRevision, targetKind, targetArtifact, targetRevision, relationType string, includeScope bool, includeGovernanceAction bool, includeLifecycleConsequence bool) string {
	t.Helper()
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	source := supersessionParticipantJSON(sourceKind, sourceArtifact, sourceRevision)
	target := supersessionParticipantJSON(targetKind, targetArtifact, targetRevision)
	scopeField := ""
	if includeScope {
		scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
		if err != nil {
			t.Fatal(err)
		}
		scopeField = `,"scope":` + string(scope)
	}
	governanceField := ""
	if includeGovernanceAction {
		governanceField = `,"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}}`
	}
	consequenceField := ""
	if includeLifecycleConsequence {
		consequenceField = `,"lifecycle_consequence":{"kind":"none"}`
	}
	return `{"relation":{"relation_type":"` + relationType + `","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + scopeField + `}` + governanceField + consequenceField + `}`
}

// --- NewSupersession -----------------------------------------------------

func TestNewSupersessionValid(t *testing.T) {
	s := mustSupersession(t)
	if s.IsZero() {
		t.Error("valid Supersession IsZero() = true")
	}
}

func TestNewSupersessionZeroSupersedingRejected(t *testing.T) {
	_, err := NewSupersession(core.RequirementArtifactRevisionRef{}, mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewSupersessionZeroSupersededRejected(t *testing.T) {
	_, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), core.RequirementArtifactRevisionRef{}, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestNewSupersessionSelfSupersessionRejected proves the direct
// self-supersession prohibition (PEOS-005 §23.1).
func TestNewSupersessionSelfSupersessionRejected(t *testing.T) {
	same := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewSupersession(same, same, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewSupersessionZeroProvenanceRejected(t *testing.T) {
	_, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), core.Provenance{}, mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewSupersessionZeroScopeRejected(t *testing.T) {
	_, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), core.Scope{}, mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewSupersessionZeroGovernanceActionRejected(t *testing.T) {
	_, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), GovernanceAction{}, NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

// TestNewSupersessionZeroLifecycleConsequenceRejected proves the unstated
// LifecycleConsequence state is rejected (PEOS-005 §23.1).
func TestNewSupersessionZeroLifecycleConsequenceRejected(t *testing.T) {
	_, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), LifecycleConsequence{})
	if !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

// TestNewSupersessionNoLifecycleConsequenceAccepted proves
// NoLifecycleConsequence() is a valid, explicit declaration -- the
// opposite of the unstated zero value.
func TestNewSupersessionNoLifecycleConsequenceAccepted(t *testing.T) {
	_, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if err != nil {
		t.Errorf("NoLifecycleConsequence() rejected: %v", err)
	}
}

func TestNewSupersessionIdentifiedLifecycleConsequenceAccepted(t *testing.T) {
	consequence, err := NewLifecycleConsequence("Superseded within the EU scope.")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), consequence)
	if err != nil {
		t.Errorf("identified LifecycleConsequence rejected: %v", err)
	}
}

func TestNewSupersessionRelationTypeAlwaysRequirementSupersession(t *testing.T) {
	s := mustSupersession(t)
	if s.Relation().RelationType() != core.RelationTypeRequirementSupersession {
		t.Errorf("RelationType() = %v, want %v", s.Relation().RelationType(), core.RelationTypeRequirementSupersession)
	}
}

// --- Identity and revision semantics — headline group -----------------------

// TestNewSupersessionDistinctArtifactIDsAccepted proves the ordinary case:
// superseding and superseded name different Requirement identities.
func TestNewSupersessionDistinctArtifactIDsAccepted(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	if _, err := NewSupersession(superseding, superseded, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence()); err != nil {
		t.Errorf("distinct-ArtifactID Supersession rejected: %v", err)
	}
}

// TestNewSupersessionSameArtifactIDDifferentRevisionsAccepted proves the
// Packet G.4 blueprint's headline conclusion: PEOS-005 §23/§7/§28.3/§35
// all state identity *preservation*, never *distinctness*, so
// REQ-1/REV-2 superseding REQ-1/REV-1 is a valid Supersession -- the
// deliberate opposite of Derivation's and Decomposition's rejection of
// the equivalent case.
func TestNewSupersessionSameArtifactIDDifferentRevisionsAccepted(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	s, err := NewSupersession(superseding, superseded, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if err != nil {
		t.Errorf("same-ArtifactID different-revision Supersession rejected: %v", err)
	}
	if errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Error("unexpected ErrInvalidRequirementSupersession for same-ArtifactID Supersession")
	}
	if s.Superseding().ArtifactID() != s.Superseded().ArtifactID() {
		t.Error("expected superseding and superseded to share one Requirement identity in this test")
	}
}

// TestNewSupersessionSameIdentityStillRequiresGovernanceAction proves a
// same-identity Supersession cannot substitute for the mandatory
// governance action -- it cannot be mistaken for ordinary revision
// history (PEOS-005 §36.13).
func TestNewSupersessionSameIdentityStillRequiresGovernanceAction(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewSupersession(superseding, superseded, mustProvenance(t), mustScope(t, "product-x", "/x"), GovernanceAction{}, NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

func TestNewSupersessionSameIdentityStillRequiresScope(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewSupersession(superseding, superseded, mustProvenance(t), core.Scope{}, mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewSupersessionSameIdentityStillRequiresProvenance(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewSupersession(superseding, superseded, core.Provenance{}, mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestNewSupersessionSameIdentityStillRequiresLifecycleConsequence(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-1", "REV-2")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, err := NewSupersession(superseding, superseded, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), LifecycleConsequence{})
	if !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

// TestSupersessionDirectionLock proves Superseding() maps to the
// relation's Source() and Superseded() maps to its Target() -- the §23.1
// inversion relative to Derivation, Refinement, and Decomposition (whose
// source is the older/originating participant).
func TestSupersessionDirectionLock(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	s, err := NewSupersession(superseding, superseded, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if err != nil {
		t.Fatal(err)
	}
	sourceRef, ok := s.Relation().Source().AsRequirementRevision()
	if !ok || sourceRef != superseding {
		t.Errorf("relation.Source() = %v, want superseding %v", sourceRef, superseding)
	}
	targetRef, ok := s.Relation().Target().AsRequirementRevision()
	if !ok || targetRef != superseded {
		t.Errorf("relation.Target() = %v, want superseded %v", targetRef, superseded)
	}
	if s.Superseding() != superseding {
		t.Errorf("Superseding() = %v, want %v", s.Superseding(), superseding)
	}
	if s.Superseded() != superseded {
		t.Errorf("Superseded() = %v, want %v", s.Superseded(), superseded)
	}
}

// --- Accessors -----------------------------------------------------------

func TestSupersessionAccessors(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	superseded := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	prov := mustProvenance(t)
	scope := mustScope(t, "product-x", "/services/*")
	governance := mustGovernanceActionFromDecisionOutcome(t, "DEC-1")
	consequence := NoLifecycleConsequence()
	s, err := NewSupersession(superseding, superseded, prov, scope, governance, consequence)
	if err != nil {
		t.Fatal(err)
	}
	if s.Superseding() != superseding {
		t.Errorf("Superseding() = %v, want %v", s.Superseding(), superseding)
	}
	if s.Superseded() != superseded {
		t.Errorf("Superseded() = %v, want %v", s.Superseded(), superseded)
	}
	if s.GovernanceAction() != governance {
		t.Errorf("GovernanceAction() = %v, want %v", s.GovernanceAction(), governance)
	}
	if s.LifecycleConsequence() != consequence {
		t.Errorf("LifecycleConsequence() = %v, want %v", s.LifecycleConsequence(), consequence)
	}
	if s.Scope() != scope {
		t.Errorf("Scope() = %v, want %v", s.Scope(), scope)
	}
	gotActor, _ := s.Provenance().Actor()
	wantActor, _ := prov.Actor()
	if gotActor != wantActor {
		t.Errorf("Provenance().Actor() = %v, want %v", gotActor, wantActor)
	}
}

// TestSupersessionScopeNeverAbsent proves Scope() always returns a
// non-zero value for a validly constructed Supersession.
func TestSupersessionScopeNeverAbsent(t *testing.T) {
	s := mustSupersession(t)
	if s.Scope().IsZero() {
		t.Error("Scope() returned zero value on a valid Supersession")
	}
}

func TestSupersessionIsZero(t *testing.T) {
	var s Supersession
	if !s.IsZero() {
		t.Error("zero Supersession IsZero() = false")
	}
	if mustSupersession(t).IsZero() {
		t.Error("valid Supersession IsZero() = true")
	}
}

// --- With* ---------------------------------------------------------------

func TestSupersessionWithExtension(t *testing.T) {
	s := mustSupersession(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := s.WithExtension(ext)
	if !s.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestSupersessionWithoutExtension(t *testing.T) {
	s := mustSupersession(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := s.WithExtension(ext)
	cleared := withExt.WithoutExtension()
	if !cleared.Extension().IsZero() {
		t.Error("Extension() non-zero after WithoutExtension")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithoutExtension mutated the original receiver")
	}
}

func TestSupersessionWithMethodsAreImmutable(t *testing.T) {
	s := mustSupersession(t)
	original := s
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = s.WithExtension(ext)
	if !original.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
}

// --- JSON --------------------------------------------------------------

func TestSupersessionJSONLiteralWireKeys(t *testing.T) {
	s := fullSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation", "governance_action", "lifecycle_consequence"} {
		if _, present := raw[key]; !present {
			t.Errorf("required key %q missing", key)
		}
	}
	if len(raw) != 3 {
		t.Errorf("Supersession wire form has %d top-level keys, want exactly 3: %v", len(raw), raw)
	}
	var relRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["relation"], &relRaw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation_type", "source", "target", "provenance", "scope", "extension"} {
		if _, present := relRaw[key]; !present {
			t.Errorf("required nested key %q missing", key)
		}
	}
}

func TestSupersessionJSONMinimumOmitsExtension(t *testing.T) {
	s := mustSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var relRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["relation"], &relRaw); err != nil {
		t.Fatal(err)
	}
	if _, present := relRaw["extension"]; present {
		t.Error("extension present despite not being set")
	}
	if _, present := relRaw["scope"]; !present {
		t.Error(`"scope" must always be present -- Supersession's scope is mandatory`)
	}
}

func TestSupersessionJSONRoundTrip(t *testing.T) {
	s := fullSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Supersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Superseding() != s.Superseding() || decoded.Superseded() != s.Superseded() {
		t.Error("participant round trip mismatch")
	}
	if decoded.GovernanceAction() != s.GovernanceAction() {
		t.Error("GovernanceAction mismatch")
	}
	if decoded.LifecycleConsequence() != s.LifecycleConsequence() {
		t.Error("LifecycleConsequence mismatch")
	}
	if decoded.Scope() != s.Scope() {
		t.Error("Scope mismatch")
	}
	if decoded.Extension().IsZero() {
		t.Error("Extension absent after round trip")
	}
}

func TestSupersessionJSONMinimumRoundTrip(t *testing.T) {
	s := mustSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Supersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Superseding() != s.Superseding() {
		t.Error("round trip mismatch")
	}
}

func TestSupersessionZeroMarshalRejected(t *testing.T) {
	var s Supersession
	if _, err := json.Marshal(s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestSupersessionJSONUnknownFieldIgnored(t *testing.T) {
	s := mustSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["unknown_field"] = json.RawMessage(`123`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Supersession
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestSupersessionUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullSupersession(t)
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if receiver.Superseding() != original.Superseding() {
		t.Error("failed Unmarshal changed receiver")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

// --- Decode-only validation ------------------------------------------------

func TestSupersessionJSONWrongRelationTypeRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:derivation", true, true, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestSupersessionJSONSourceAtIdentityLevelRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestSupersessionJSONTargetAtIdentityLevelRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestSupersessionJSONWrongSubjectKindRejected proves a non-Requirement
// subject (Decision-kind) is rejected on decode.
func TestSupersessionJSONWrongSubjectKindRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	source := `{"kind":"decision","ref":{"decision_id":"DEC-2"}}`
	target := supersessionParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	payload := `{"relation":{"relation_type":"peos:requirement-supersession","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"lifecycle_consequence":{"kind":"none"}}`
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

func TestSupersessionJSONSourceEqualsTargetRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestSupersessionJSONSameArtifactIDDifferentRevisionsAccepted proves the
// §23 identity rule is not enforced on decode either.
func TestSupersessionJSONSameArtifactIDDifferentRevisionsAccepted(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-1", "REV-2", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Errorf("same-ArtifactID different-revision Supersession rejected on decode: %v", err)
	}
}

func TestSupersessionJSONMissingScopeRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", false, true, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementRelation)
	}
}

// TestSupersessionJSONMissingGovernanceActionRejected proves
// governance_action is mandatory on decode, not merely on construction.
func TestSupersessionJSONMissingGovernanceActionRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, false, true)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidGovernanceAction) {
		t.Errorf("error = %v, want %v", err, ErrInvalidGovernanceAction)
	}
}

// TestSupersessionJSONMissingLifecycleConsequenceRejected proves the
// unstated LifecycleConsequence is rejected on decode -- §23.1 requires
// explicit representation of absence, not omission.
func TestSupersessionJSONMissingLifecycleConsequenceRejected(t *testing.T) {
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, false)
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidRequirementSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRequirementSupersession)
	}
}

func TestSupersessionJSONExplicitNullRelationRejected(t *testing.T) {
	var s Supersession
	payload := `{"relation":null,"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"lifecycle_consequence":{"kind":"none"}}`
	if err := json.Unmarshal([]byte(payload), &s); err == nil {
		t.Error("null relation accepted, want error")
	}
}

func TestSupersessionJSONExplicitNullGovernanceActionRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	source := supersessionParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	target := supersessionParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	payload := `{"relation":{"relation_type":"peos:requirement-supersession","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"governance_action":null,"lifecycle_consequence":{"kind":"none"}}`
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); err == nil {
		t.Error("null governance_action accepted, want error")
	}
}

func TestSupersessionJSONExplicitNullLifecycleConsequenceRejected(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	source := supersessionParticipantJSON("requirement_revision", "REQ-2", "REV-1")
	target := supersessionParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	payload := `{"relation":{"relation_type":"peos:requirement-supersession","source":` + source + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"lifecycle_consequence":null}`
	var s Supersession
	if err := json.Unmarshal([]byte(payload), &s); err == nil {
		t.Error("null lifecycle_consequence accepted, want error")
	}
}

// --- Constructor / Unmarshal equivalence ------------------------------------

func TestSupersessionConstructorUnmarshalEquivalenceSelfSupersession(t *testing.T) {
	same := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	_, ctorErr := NewSupersession(same, same, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	payload := supersessionPayload(t, "requirement_revision", "REQ-1", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, true)
	var s Supersession
	jsonErr := json.Unmarshal([]byte(payload), &s)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

func TestSupersessionConstructorUnmarshalEquivalenceMissingScope(t *testing.T) {
	_, ctorErr := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), core.Scope{}, mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", false, true, true)
	var s Supersession
	jsonErr := json.Unmarshal([]byte(payload), &s)
	if !errors.Is(ctorErr, ErrInvalidRequirementRelation) || !errors.Is(jsonErr, ErrInvalidRequirementRelation) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementRelation)
	}
}

func TestSupersessionConstructorUnmarshalEquivalenceZeroGovernanceAction(t *testing.T) {
	_, ctorErr := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), GovernanceAction{}, NoLifecycleConsequence())
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, false, true)
	var s Supersession
	jsonErr := json.Unmarshal([]byte(payload), &s)
	if !errors.Is(ctorErr, ErrInvalidGovernanceAction) || !errors.Is(jsonErr, ErrInvalidGovernanceAction) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidGovernanceAction)
	}
}

func TestSupersessionConstructorUnmarshalEquivalenceUnstatedLifecycleConsequence(t *testing.T) {
	_, ctorErr := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), LifecycleConsequence{})
	payload := supersessionPayload(t, "requirement_revision", "REQ-2", "REV-1", "requirement_revision", "REQ-1", "REV-1", "peos:requirement-supersession", true, true, false)
	var s Supersession
	jsonErr := json.Unmarshal([]byte(payload), &s)
	if !errors.Is(ctorErr, ErrInvalidRequirementSupersession) || !errors.Is(jsonErr, ErrInvalidRequirementSupersession) {
		t.Errorf("ctorErr = %v, jsonErr = %v, want both wrapping %v", ctorErr, jsonErr, ErrInvalidRequirementSupersession)
	}
}

// --- Nested sentinel preservation --------------------------------------

func TestSupersessionNestedSentinelPreserved(t *testing.T) {
	prov, err := json.Marshal(mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	scope, err := json.Marshal(mustScope(t, "product-x", "/x"))
	if err != nil {
		t.Fatal(err)
	}
	malformedSource := `{"kind":"requirement_revision","ref":{"artifact_id":"","revision_id":"REV-1"}}`
	target := supersessionParticipantJSON("requirement_revision", "REQ-1", "REV-1")
	payload := `{"relation":{"relation_type":"peos:requirement-supersession","source":` + malformedSource + `,"target":` + target + `,"provenance":` + string(prov) + `,"scope":` + string(scope) + `},"governance_action":{"kind":"decision_outcome","ref":{"decision_id":"DEC-1"}},"lifecycle_consequence":{"kind":"none"}}`
	var s Supersession
	err = json.Unmarshal([]byte(payload), &s)
	if err == nil {
		t.Fatal("malformed nested source ref accepted, want error")
	}
	if !errors.Is(err, ErrInvalidRequirementRelation) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidRequirementRelation)
	}
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("error = %v, want also wrapping %v", err, core.ErrEmptyIdentity)
	}
}

// --- Cycle / branching local behavior ---------------------------------------

// TestSupersessionOneSupersedingMultipleSuperseded proves one superseding
// Revision MAY be the source of multiple Supersession relationships
// (PEOS-005 §23.1).
func TestSupersessionOneSupersedingMultipleSuperseded(t *testing.T) {
	superseding := mustRequirementRevisionRef(t, "REQ-3", "REV-2")
	s1, err := NewSupersession(superseding, mustRequirementRevisionRef(t, "REQ-1", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if err != nil {
		t.Fatal(err)
	}
	s2, err := NewSupersession(superseding, mustRequirementRevisionRef(t, "REQ-2", "REV-1"), mustProvenance(t), mustScope(t, "product-x", "/y"), mustGovernanceActionFromDecisionOutcome(t, "DEC-2"), NoLifecycleConsequence())
	if err != nil {
		t.Fatal(err)
	}
	if s1.Superseding() != s2.Superseding() {
		t.Error("expected both Supersessions to share the same superseding Revision")
	}
	if s1.Superseded() == s2.Superseded() {
		t.Error("expected distinct superseded Revisions")
	}
}

// TestSupersessionOneSupersededMultipleSuperseding proves one superseded
// Revision MAY be the target of more than one Supersession relationship
// (PEOS-005 §23.1); distinguishing their scopes is a repository-level
// concern this package does not enforce.
func TestSupersessionOneSupersededMultipleSuperseding(t *testing.T) {
	superseded := mustRequirementRevisionRef(t, "REQ-3", "REV-1")
	s1, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-1", "REV-1"), superseded, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence())
	if err != nil {
		t.Fatal(err)
	}
	s2, err := NewSupersession(mustRequirementRevisionRef(t, "REQ-2", "REV-1"), superseded, mustProvenance(t), mustScope(t, "product-x", "/y"), mustGovernanceActionFromDecisionOutcome(t, "DEC-2"), NoLifecycleConsequence())
	if err != nil {
		t.Fatal(err)
	}
	if s1.Superseded() != s2.Superseded() {
		t.Error("expected both Supersessions to share the same superseded Revision")
	}
}

// TestSupersessionReverseIndividuallyConstructAllowedLocally proves this
// package does not attempt transitive cycle detection: A superseding B
// and B superseding A both construct individually. A repository is
// responsible for rejecting the resulting cycle (PEOS-005 §23.1).
func TestSupersessionReverseIndividuallyConstructAllowedLocally(t *testing.T) {
	a := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	b := mustRequirementRevisionRef(t, "REQ-2", "REV-1")
	if _, err := NewSupersession(a, b, mustProvenance(t), mustScope(t, "product-x", "/x"), mustGovernanceActionFromDecisionOutcome(t, "DEC-1"), NoLifecycleConsequence()); err != nil {
		t.Errorf("A-supersedes-B rejected: %v", err)
	}
	if _, err := NewSupersession(b, a, mustProvenance(t), mustScope(t, "product-x", "/y"), mustGovernanceActionFromDecisionOutcome(t, "DEC-2"), NoLifecycleConsequence()); err != nil {
		t.Errorf("B-supersedes-A rejected: %v", err)
	}
}

// --- Absence audit -------------------------------------------------------

// TestSupersessionNoIdentityField is a structural absence audit proving
// Supersession carries no identity (PEOS-005 §17.1).
func TestSupersessionNoIdentityField(t *testing.T) {
	s := fullSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["id"]; present {
		t.Error(`unexpected "id" key present in Supersession wire form`)
	}
	if len(raw) != 3 {
		t.Errorf("Supersession wire form has %d top-level keys, want exactly 3 (relation, governance_action, lifecycle_consequence): %v", len(raw), raw)
	}
}
