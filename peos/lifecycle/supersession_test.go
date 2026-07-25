package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustSupersessionID(t *testing.T, value string) core.LifecycleDefinitionVersionSupersessionID {
	t.Helper()
	id, err := core.NewLifecycleDefinitionVersionSupersessionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustDefinitionVersionRef(t *testing.T, definitionID, versionID string) core.LifecycleDefinitionVersionRef {
	t.Helper()
	ref, err := core.NewLifecycleDefinitionVersionRef(mustLifecycleDefinitionID(t, definitionID), func() core.LifecycleDefinitionVersionID {
		id, err := core.NewLifecycleDefinitionVersionID(versionID)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}())
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// validSupersessionParts returns a minimally valid set of
// NewLifecycleDefinitionVersionSupersession arguments a test can freely
// mutate one field of.
func validSupersessionParts(t *testing.T) (
	id core.LifecycleDefinitionVersionSupersessionID,
	supersedingVersion core.LifecycleDefinitionVersionRef,
	supersededVersion core.LifecycleDefinitionVersionRef,
	scope core.Scope,
	compatibilityConsequence string,
	migrationRequirement string,
	provenance core.Provenance,
) {
	t.Helper()
	id = mustSupersessionID(t, "SUP-1001")
	supersedingVersion = mustDefinitionVersionRef(t, "LC-REVIEW-1", "V2")
	supersededVersion = mustDefinitionVersionRef(t, "LC-REVIEW-1", "V1")
	scope = mustScope(t, "product-x", "always")
	compatibilityConsequence = "V2 is backward compatible with V1: no States or Transitions were removed."
	migrationRequirement = "Subjects assigned V1 states MAY remain on V1 until explicitly migrated; no automatic migration is performed."
	provenance = mustProvenance(t)
	return
}

func mustSupersession(t *testing.T) LifecycleDefinitionVersionSupersession {
	t.Helper()
	s, err := NewLifecycleDefinitionVersionSupersession(validSupersessionParts(t))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNewLifecycleDefinitionVersionSupersessionValid(t *testing.T) {
	s := mustSupersession(t)
	if s.IsZero() {
		t.Error("valid Supersession reports IsZero() = true")
	}
}

func TestNewLifecycleDefinitionVersionSupersessionSameVersionRejected(t *testing.T) {
	id, _, _, scope, cc, mr, provenance := validSupersessionParts(t)
	same := mustDefinitionVersionRef(t, "LC-REVIEW-1", "V1")
	if _, err := NewLifecycleDefinitionVersionSupersession(id, same, same, scope, cc, mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestNewLifecycleDefinitionVersionSupersessionZeroIDRejected(t *testing.T) {
	_, supersedingVersion, supersededVersion, scope, cc, mr, provenance := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(core.LifecycleDefinitionVersionSupersessionID{}, supersedingVersion, supersededVersion, scope, cc, mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestNewLifecycleDefinitionVersionSupersessionZeroSupersedingRejected(t *testing.T) {
	id, _, supersededVersion, scope, cc, mr, provenance := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(id, core.LifecycleDefinitionVersionRef{}, supersededVersion, scope, cc, mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestNewLifecycleDefinitionVersionSupersessionZeroSupersededRejected(t *testing.T) {
	id, supersedingVersion, _, scope, cc, mr, provenance := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, core.LifecycleDefinitionVersionRef{}, scope, cc, mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestNewLifecycleDefinitionVersionSupersessionZeroScopeRejected(t *testing.T) {
	id, supersedingVersion, supersededVersion, _, cc, mr, provenance := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, core.Scope{}, cc, mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestNewLifecycleDefinitionVersionSupersessionEmptyCompatibilityConsequenceRejected(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, _, mr, provenance := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, "", mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
	if _, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, "   ", mr, provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("whitespace-only: error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

// TestNewLifecycleDefinitionVersionSupersessionNonEmptyMigrationRequirementPresent
// confirms the constructor's ordinary case: a non-empty argument records a
// present migration requirement.
func TestNewLifecycleDefinitionVersionSupersessionNonEmptyMigrationRequirementPresent(t *testing.T) {
	s := mustSupersession(t)
	_, _, _, _, _, mr, _ := validSupersessionParts(t)
	got, ok := s.MigrationRequirement()
	if !ok || got != mr {
		t.Errorf("MigrationRequirement() = (%q, %v), want (%q, true)", got, ok, mr)
	}
}

// TestNewLifecycleDefinitionVersionSupersessionEmptyMigrationRequirementIsValidAndAbsent
// confirms the source-compatibility bridge: an empty (or whitespace-only)
// migrationRequirement argument is accepted -- not rejected -- and produces
// a Supersession with no migration requirement recorded. This is a
// behavior change from the previous packet (which rejected empty as an
// error), but it does not break any previously *successful* call: every
// call that used to succeed (non-empty argument) still succeeds and still
// records the same value.
func TestNewLifecycleDefinitionVersionSupersessionEmptyMigrationRequirementIsValidAndAbsent(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatalf("empty migration requirement unexpectedly rejected: %v", err)
	}
	if _, ok := s.MigrationRequirement(); ok {
		t.Error("MigrationRequirement() ok = true after constructing with an empty argument")
	}

	s2, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "   ", provenance)
	if err != nil {
		t.Fatalf("whitespace-only migration requirement unexpectedly rejected: %v", err)
	}
	if _, ok := s2.MigrationRequirement(); ok {
		t.Error("MigrationRequirement() ok = true after constructing with a whitespace-only argument")
	}
}

func TestNewLifecycleDefinitionVersionSupersessionZeroProvenanceRejected(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, mr, _ := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, mr, core.Provenance{}); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestLifecycleDefinitionVersionSupersessionAccessors(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, mr, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, mr, provenance)
	if err != nil {
		t.Fatal(err)
	}
	if s.ID() != id {
		t.Error("ID() mismatch")
	}
	if s.SupersedingVersion() != supersedingVersion {
		t.Error("SupersedingVersion() mismatch")
	}
	if s.SupersededVersion() != supersededVersion {
		t.Error("SupersededVersion() mismatch")
	}
	if !s.Scope().Equal(scope) {
		t.Error("Scope() mismatch")
	}
	if s.CompatibilityConsequence() != cc {
		t.Error("CompatibilityConsequence() mismatch")
	}
	gotMR, gotMROK := s.MigrationRequirement()
	if !gotMROK || gotMR != mr {
		t.Errorf("MigrationRequirement() = (%q, %v), want (%q, true)", gotMR, gotMROK, mr)
	}
	gotActor, gotOK := s.Provenance().Actor()
	wantActor, wantOK := provenance.Actor()
	if gotOK != wantOK || gotActor != wantActor {
		t.Error("Provenance() mismatch")
	}
}

func TestLifecycleDefinitionVersionSupersessionExtensionDefensiveCopy(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	s := mustSupersession(t)
	withExt := s.WithExtension(ext)
	if !s.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	got, ok := withExt.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestLifecycleDefinitionVersionSupersessionRef(t *testing.T) {
	s := mustSupersession(t)
	ref, err := s.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.SupersessionID() != s.ID() {
		t.Error("Ref() component mismatch")
	}
}

func TestLifecycleDefinitionVersionSupersessionJSONRoundTripMigrationRequirementPresent(t *testing.T) {
	s := mustSupersession(t) // built with a non-empty migration requirement
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != s.ID() {
		t.Error("round trip changed ID")
	}
	if decoded.SupersedingVersion() != s.SupersedingVersion() || decoded.SupersededVersion() != s.SupersededVersion() {
		t.Error("round trip changed version references")
	}
	if decoded.CompatibilityConsequence() != s.CompatibilityConsequence() {
		t.Error("round trip changed CompatibilityConsequence")
	}
	gotMR, gotOK := decoded.MigrationRequirement()
	wantMR, wantOK := s.MigrationRequirement()
	if gotOK != wantOK || gotMR != wantMR {
		t.Errorf("MigrationRequirement() = (%q, %v), want (%q, %v)", gotMR, gotOK, wantMR, wantOK)
	}
}

// TestLifecycleDefinitionVersionSupersessionJSONRoundTripMigrationRequirementAbsent
// proves a Supersession constructed with no migration requirement at all
// round-trips through JSON with the requirement still absent afterward.
func TestLifecycleDefinitionVersionSupersessionJSONRoundTripMigrationRequirementAbsent(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded.MigrationRequirement(); ok {
		t.Error("MigrationRequirement() ok = true after round-tripping a Supersession with none set")
	}
}

// TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementKeyOmittedWhenAbsent
// confirms Marshal never writes an empty-string migration_requirement key;
// it omits the key entirely instead.
func TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementKeyOmittedWhenAbsent(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["migration_requirement"]; present {
		t.Error(`"migration_requirement" key present in JSON despite no migration requirement being set`)
	}
}

// TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementEmptyStringRejected
// confirms an explicit empty string is rejected, not silently treated as
// absent -- absence must be expressed by omitting the key, per the JSON
// contract.
func TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementEmptyStringRejected(t *testing.T) {
	validRef := `{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"}`
	payload := `{"id":"SUP-1001","superseding_version":` + validRef + `,"superseded_version":` + validRef + `,"scope":{"kind":"product-x:condition","expression":"always"},"compatibility_consequence":"c","migration_requirement":"","provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	var s LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

// TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementNullRejected
// confirms an explicit JSON null is rejected.
func TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementNullRejected(t *testing.T) {
	validRef := `{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"}`
	payload := `{"id":"SUP-1001","superseding_version":` + validRef + `,"superseded_version":` + validRef + `,"scope":{"kind":"product-x:condition","expression":"always"},"compatibility_consequence":"c","migration_requirement":null,"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	var s LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

// TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementWrongTypeRejected
// confirms a non-string JSON value for migration_requirement is rejected.
func TestLifecycleDefinitionVersionSupersessionJSONMigrationRequirementWrongTypeRejected(t *testing.T) {
	validRef := `{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"}`
	payload := `{"id":"SUP-1001","superseding_version":` + validRef + `,"superseded_version":` + validRef + `,"scope":{"kind":"product-x:condition","expression":"always"},"compatibility_consequence":"c","migration_requirement":123,"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	var s LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal([]byte(payload), &s); err == nil {
		t.Fatal("non-string migration_requirement accepted, want error")
	}
}

// TestExisting25567f1PayloadWithMigrationRequirementRemainsValid proves
// backward compatibility: a Supersession JSON payload with a non-empty
// migration_requirement key, in the exact shape produced before this
// packet, still decodes correctly.
func TestExisting25567f1PayloadWithMigrationRequirementRemainsValid(t *testing.T) {
	payload := `{
		"id": "SUP-1001",
		"superseding_version": {"lifecycle_definition_id": "LC-REVIEW-1", "lifecycle_definition_version_id": "V2"},
		"superseded_version": {"lifecycle_definition_id": "LC-REVIEW-1", "lifecycle_definition_version_id": "V1"},
		"scope": {"kind": "product-x:condition", "expression": "always"},
		"compatibility_consequence": "V2 is backward compatible with V1.",
		"migration_requirement": "Migrate subjects using the approved state mapping.",
		"provenance": {"actor": {"namespace": "peos-cli", "identifier": "svc-1"}, "recorded_at": "2026-01-01T00:00:00Z"}
	}`
	var s LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Fatalf("pre-E.1.1 JSON shape with migration_requirement unexpectedly rejected: %v", err)
	}
	got, ok := s.MigrationRequirement()
	if !ok || got != "Migrate subjects using the approved state mapping." {
		t.Errorf("MigrationRequirement() = (%q, %v), want (%q, true)", got, ok, "Migrate subjects using the approved state mapping.")
	}
}

func TestLifecycleDefinitionVersionSupersessionJSONExplicitNullRefsRejected(t *testing.T) {
	validRef := `{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"}`

	t.Run("superseding_version null", func(t *testing.T) {
		payload := `{"id":"SUP-1001","superseding_version":null,"superseded_version":` + validRef + `,"scope":{"kind":"product-x:condition","expression":"always"},"compatibility_consequence":"c","migration_requirement":"m","provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
		var s LifecycleDefinitionVersionSupersession
		if err := json.Unmarshal([]byte(payload), &s); err == nil {
			t.Fatal("explicit null superseding_version accepted, want error")
		}
	})

	t.Run("superseded_version null", func(t *testing.T) {
		payload := `{"id":"SUP-1001","superseding_version":` + validRef + `,"superseded_version":null,"scope":{"kind":"product-x:condition","expression":"always"},"compatibility_consequence":"c","migration_requirement":"m","provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
		var s LifecycleDefinitionVersionSupersession
		if err := json.Unmarshal([]byte(payload), &s); err == nil {
			t.Fatal("explicit null superseded_version accepted, want error")
		}
	})
}

func TestLifecycleDefinitionVersionSupersessionJSONExplicitNullScopeRejected(t *testing.T) {
	validRef := `{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"}`
	payload := `{"id":"SUP-1001","superseding_version":` + validRef + `,"superseded_version":` + validRef + `,"scope":null,"compatibility_consequence":"c","migration_requirement":"m","provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	var s LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal([]byte(payload), &s); err == nil {
		t.Fatal("explicit null scope accepted, want error")
	}
}

func TestLifecycleDefinitionVersionSupersessionJSONUnknownOrdinaryFieldIgnored(t *testing.T) {
	s := mustSupersession(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["bogus_field"] = json.RawMessage(`123`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LifecycleDefinitionVersionSupersession
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatalf("unknown ordinary field unexpectedly rejected: %v", err)
	}
}

func TestLifecycleDefinitionVersionSupersessionZeroMarshalRejected(t *testing.T) {
	var s LifecycleDefinitionVersionSupersession
	if _, err := json.Marshal(s); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestLifecycleDefinitionVersionSupersessionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original := mustSupersession(t).WithExtension(ext) // fully populated, including a migration requirement

	receiver := original
	payload := `{"id":"SUP-9999"}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("incomplete payload accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver's ID")
	}
	if receiver.CompatibilityConsequence() != original.CompatibilityConsequence() {
		t.Error("failed Unmarshal changed receiver's CompatibilityConsequence")
	}
	gotMR, gotMROK := receiver.MigrationRequirement()
	wantMR, wantMROK := original.MigrationRequirement()
	if gotMROK != wantMROK || gotMR != wantMR {
		t.Error("failed Unmarshal changed receiver's MigrationRequirement")
	}
	got, ok := receiver.Extension().Get("product-x")
	want, wantOK := original.Extension().Get("product-x")
	if ok != wantOK || string(got) != string(want) {
		t.Error("failed Unmarshal changed receiver's Extension")
	}
}

// TestLifecycleDefinitionVersionSupersessionUnmarshalJSONFailurePreservesReceiverOnBadMigrationRequirement
// exercises the specific failure path where every other field decodes
// successfully but migration_requirement itself is invalid (null): the
// receiver must still be left untouched, matching every other failure
// path above.
func TestLifecycleDefinitionVersionSupersessionUnmarshalJSONFailurePreservesReceiverOnBadMigrationRequirement(t *testing.T) {
	original := mustSupersession(t)
	receiver := original
	validRef := `{"lifecycle_definition_id":"LC-REVIEW-1","lifecycle_definition_version_id":"V1"}`
	payload := `{"id":"SUP-1001","superseding_version":` + validRef + `,"superseded_version":` + validRef + `,"scope":{"kind":"product-x:condition","expression":"always"},"compatibility_consequence":"c","migration_requirement":null,"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"}}}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("null migration_requirement accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver's ID")
	}
	gotMR, gotMROK := receiver.MigrationRequirement()
	wantMR, wantMROK := original.MigrationRequirement()
	if gotMROK != wantMROK || gotMR != wantMR {
		t.Error("failed Unmarshal changed receiver's MigrationRequirement")
	}
}

// --- Packet E.1.1: WithMigrationRequirement / WithoutMigrationRequirement ---

func TestWithMigrationRequirementValid(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatal(err)
	}
	withReq, err := s.WithMigrationRequirement("Migrate subjects using the approved state mapping.")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withReq.MigrationRequirement()
	if !ok || got != "Migrate subjects using the approved state mapping." {
		t.Errorf("MigrationRequirement() = (%q, %v), want present", got, ok)
	}
}

func TestWithMigrationRequirementEmptyRejected(t *testing.T) {
	s := mustSupersession(t)
	if _, err := s.WithMigrationRequirement(""); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
	if _, err := s.WithMigrationRequirement("   "); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("whitespace-only: error = %v, want %v", err, ErrInvalidLifecycleSupersession)
	}
}

func TestWithMigrationRequirementDoesNotMutateReceiver(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.WithMigrationRequirement("some requirement"); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.MigrationRequirement(); ok {
		t.Error("WithMigrationRequirement mutated the original receiver: s now has a migration requirement")
	}
}

func TestWithMigrationRequirementPreservesUnrelatedFields(t *testing.T) {
	s := mustSupersession(t)
	withReq, err := s.WithMigrationRequirement("a different requirement")
	if err != nil {
		t.Fatal(err)
	}
	if withReq.ID() != s.ID() {
		t.Error("WithMigrationRequirement changed ID")
	}
	if withReq.SupersedingVersion() != s.SupersedingVersion() || withReq.SupersededVersion() != s.SupersededVersion() {
		t.Error("WithMigrationRequirement changed version references")
	}
	if !withReq.Scope().Equal(s.Scope()) {
		t.Error("WithMigrationRequirement changed Scope")
	}
	if withReq.CompatibilityConsequence() != s.CompatibilityConsequence() {
		t.Error("WithMigrationRequirement changed CompatibilityConsequence")
	}
	gotActor, gotOK := withReq.Provenance().Actor()
	wantActor, wantOK := s.Provenance().Actor()
	if gotOK != wantOK || gotActor != wantActor {
		t.Error("WithMigrationRequirement changed Provenance")
	}
}

func TestWithoutMigrationRequirementClearsPresence(t *testing.T) {
	s := mustSupersession(t) // built with a present migration requirement
	if _, ok := s.MigrationRequirement(); !ok {
		t.Fatal("test precondition failed: s should have a migration requirement set")
	}
	cleared := s.WithoutMigrationRequirement()
	if _, ok := cleared.MigrationRequirement(); ok {
		t.Error("MigrationRequirement() ok = true after WithoutMigrationRequirement()")
	}
}

func TestWithoutMigrationRequirementAlreadyAbsentRemainsValid(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatal(err)
	}
	cleared := s.WithoutMigrationRequirement()
	if _, ok := cleared.MigrationRequirement(); ok {
		t.Error("MigrationRequirement() ok = true for a Supersession that never had one set")
	}
	if cleared.IsZero() {
		t.Error("WithoutMigrationRequirement produced a zero-value Supersession")
	}
}

func TestWithoutMigrationRequirementDoesNotMutateReceiver(t *testing.T) {
	s := mustSupersession(t)
	_ = s.WithoutMigrationRequirement()
	if _, ok := s.MigrationRequirement(); !ok {
		t.Error("WithoutMigrationRequirement mutated the original receiver: s lost its migration requirement")
	}
}

func TestWithoutMigrationRequirementPreservesUnrelatedFields(t *testing.T) {
	s := mustSupersession(t)
	cleared := s.WithoutMigrationRequirement()
	if cleared.ID() != s.ID() {
		t.Error("WithoutMigrationRequirement changed ID")
	}
	if cleared.SupersedingVersion() != s.SupersedingVersion() || cleared.SupersededVersion() != s.SupersededVersion() {
		t.Error("WithoutMigrationRequirement changed version references")
	}
	if !cleared.Scope().Equal(s.Scope()) {
		t.Error("WithoutMigrationRequirement changed Scope")
	}
	if cleared.CompatibilityConsequence() != s.CompatibilityConsequence() {
		t.Error("WithoutMigrationRequirement changed CompatibilityConsequence")
	}
}

// TestMigrationRequirementAbsentAccessor confirms the accessor's absent
// case independently of any With*/Without* call.
func TestMigrationRequirementAbsentAccessor(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	s, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := s.MigrationRequirement()
	if ok || got != "" {
		t.Errorf("MigrationRequirement() = (%q, %v), want (\"\", false)", got, ok)
	}
}
