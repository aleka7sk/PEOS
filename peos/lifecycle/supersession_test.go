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

func TestNewLifecycleDefinitionVersionSupersessionEmptyMigrationRequirementRejected(t *testing.T) {
	id, supersedingVersion, supersededVersion, scope, cc, _, provenance := validSupersessionParts(t)
	if _, err := NewLifecycleDefinitionVersionSupersession(id, supersedingVersion, supersededVersion, scope, cc, "", provenance); !errors.Is(err, ErrInvalidLifecycleSupersession) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleSupersession)
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
	if s.MigrationRequirement() != mr {
		t.Error("MigrationRequirement() mismatch")
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

func TestLifecycleDefinitionVersionSupersessionJSONRoundTrip(t *testing.T) {
	s := mustSupersession(t)
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
	if decoded.MigrationRequirement() != s.MigrationRequirement() {
		t.Error("round trip changed MigrationRequirement")
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
	original := mustSupersession(t).WithExtension(ext)

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
	got, ok := receiver.Extension().Get("product-x")
	want, wantOK := original.Extension().Get("product-x")
	if ok != wantOK || string(got) != string(want) {
		t.Error("failed Unmarshal changed receiver's Extension")
	}
}
