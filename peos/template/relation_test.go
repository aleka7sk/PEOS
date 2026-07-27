package template

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
	"github.com/aleka7sk/PEOS/peos/relation"
)

// --- helpers -----------------------------------------------------------------

func mustGeneratedFrom(t *testing.T) GeneratedFrom {
	t.Helper()
	g, err := NewGeneratedFrom(
		mustGeneratedRevisionRef(t, "GEN-1", "REV-1"),
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustScope(t, "tenant=acme"),
		mustProvenance(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func mustComposition(t *testing.T) Composition {
	t.Helper()
	c, err := NewComposition(
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustTemplateRevisionRef(t, "TPL-2", "REV-1"),
		mustScope(t, "tenant=acme"),
		mustProvenance(t),
		"the composing template's name parameter maps to the composed template's subject",
		"the composing template's value wins",
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustSpecialization(t *testing.T) Specialization {
	t.Helper()
	s, err := NewSpecialization(
		mustTemplateRevisionRef(t, "TPL-1", "REV-1"),
		mustTemplateRevisionRef(t, "TPL-BASE", "REV-1"),
		mustScope(t, "tenant=acme"),
		mustProvenance(t),
		"all base parameters and constraints",
		"the base statement template",
		"remains compatible with the base parameter contract",
	)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// --- GeneratedFrom -----------------------------------------------------------

func TestNewGeneratedFrom(t *testing.T) {
	g := mustGeneratedFrom(t)
	if g.IsZero() {
		t.Error("valid GeneratedFrom reports IsZero() = true")
	}
	if g.Relation().RelationType() != core.RelationTypeGeneratedFrom {
		t.Errorf("RelationType() = %v", g.Relation().RelationType())
	}

	// Direction is fixed: generated → template.
	generated, ok := g.Generated()
	if !ok || generated != mustGeneratedRevisionRef(t, "GEN-1", "REV-1") {
		t.Error("Generated() mismatch; the source must be the generated Artifact Revision")
	}
	template, ok := g.Template()
	if !ok || template != mustTemplateRevisionRef(t, "TPL-1", "REV-1") {
		t.Error("Template() mismatch; the target must be the Template Artifact Revision")
	}
	if g.Scope() != mustScope(t, "tenant=acme") {
		t.Error("Scope() mismatch")
	}
	if g.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
}

func TestNewGeneratedFromRejections(t *testing.T) {
	generated := mustGeneratedRevisionRef(t, "GEN-1", "REV-1")
	template := mustTemplateRevisionRef(t, "TPL-1", "REV-1")
	scope := mustScope(t, "tenant=acme")

	if _, err := NewGeneratedFrom(core.GeneratedArtifactRevisionRef{}, template, scope, mustProvenance(t)); !errors.Is(err, ErrInvalidGeneratedFrom) {
		t.Errorf("zero generated revision: error = %v, want %v", err, ErrInvalidGeneratedFrom)
	}
	if _, err := NewGeneratedFrom(generated, core.TemplateArtifactRevisionRef{}, scope, mustProvenance(t)); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero template revision: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	if _, err := NewGeneratedFrom(generated, template, core.Scope{}, mustProvenance(t)); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
	if _, err := NewGeneratedFrom(generated, template, scope, core.Provenance{}); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}

	var zero GeneratedFrom
	if !zero.IsZero() {
		t.Error("zero-value GeneratedFrom.IsZero() = false, want true")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidGeneratedFrom) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidGeneratedFrom)
	}
}

// TestGeneratedFromCarriesNoApplicationState is the structural proof of what
// PEOS-009 states directly: a Generated-From relation SHALL NOT contain "the
// full resolved parameter state; execution event history; mutable application
// status; authority history."
func TestGeneratedFromCarriesNoApplicationState(t *testing.T) {
	g := mustGeneratedFrom(t)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	// The wrapper adds nothing beyond the inner relation.
	if len(m) != 1 {
		t.Errorf("wire form has %d keys, want exactly 1 (relation)", len(m))
	}
	if _, ok := m["relation"]; !ok {
		t.Error(`wire form missing "relation"`)
	}
	assertNoWireKeys(t, "GeneratedFrom", g, []string{
		"resolved_values", "parameters", "outcome", "events", "event_history",
		"authority_history", "authority", "status", "application_record",
	})
}

func TestGeneratedFromJSONRoundTrip(t *testing.T) {
	g := mustGeneratedFrom(t)
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	g = g.WithExtension(ext)
	if g.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GeneratedFrom
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", data2, data)
	}
	if decoded.Extension().IsZero() {
		t.Error("round trip lost the extension")
	}
	if !g.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear the extension")
	}
}

// TestGeneratedFromRejectsWrongRelationType confirms a decoded relation of
// another type cannot masquerade as a GeneratedFrom.
func TestGeneratedFromRejectsWrongRelationType(t *testing.T) {
	c := mustComposition(t)
	data, err := json.Marshal(c.Relation())
	if err != nil {
		t.Fatal(err)
	}
	var g GeneratedFrom
	if err := json.Unmarshal([]byte(`{"relation":`+string(data)+`}`), &g); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("composition decoded as generated-from: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
}

// TestGeneratedFromRejectsWrongParticipantLevels confirms the direction and
// participant levels are fixed: a Template-to-Template pair is not a
// Generated-From, whichever way round it is.
func TestGeneratedFromRejectsWrongParticipantLevels(t *testing.T) {
	tplSubject, err := core.EngineeringSubjectRefFromTemplateRevision(mustTemplateRevisionRef(t, "TPL-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	otherTpl, err := core.EngineeringSubjectRefFromTemplateRevision(mustTemplateRevisionRef(t, "TPL-2", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	rel, err := relation.New(core.RelationTypeGeneratedFrom, tplSubject, otherTpl, mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	rel, err = rel.WithScope(mustScope(t, "tenant=acme"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(rel)
	if err != nil {
		t.Fatal(err)
	}
	var g GeneratedFrom
	if err := json.Unmarshal([]byte(`{"relation":`+string(data)+`}`), &g); !errors.Is(err, ErrInvalidGeneratedFrom) {
		t.Errorf("template source accepted: error = %v, want %v", err, ErrInvalidGeneratedFrom)
	}
}

// TestRelationWrappersRequireScopeOnDecode confirms the mandatory scope
// PEOS-009 states unqualified survives the round trip through
// relation.Relation, whose own scope is optional.
func TestRelationWrappersRequireScopeOnDecode(t *testing.T) {
	generated, err := core.EngineeringSubjectRefFromGeneratedArtifactRevision(mustGeneratedRevisionRef(t, "GEN-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	template, err := core.EngineeringSubjectRefFromTemplateRevision(mustTemplateRevisionRef(t, "TPL-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	// A scopeless relation of the right type and participants.
	rel, err := relation.New(core.RelationTypeGeneratedFrom, generated, template, mustProvenance(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(rel)
	if err != nil {
		t.Fatal(err)
	}
	var g GeneratedFrom
	if err := json.Unmarshal([]byte(`{"relation":`+string(data)+`}`), &g); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("scopeless generated-from: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
}

func TestGeneratedFromUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	g := mustGeneratedFrom(t)
	before, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{`{}`, `{"relation":null}`, `123`, `not json`} {
		if err := json.Unmarshal([]byte(payload), &g); err == nil {
			t.Errorf("%s accepted, want error", payload)
		}
	}
	after, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- Composition -------------------------------------------------------------

func TestNewComposition(t *testing.T) {
	c := mustComposition(t)
	if c.IsZero() {
		t.Error("valid Composition reports IsZero() = true")
	}
	if c.Relation().RelationType() != core.RelationTypeTemplateComposition {
		t.Errorf("RelationType() = %v", c.Relation().RelationType())
	}
	composing, ok := c.Composing()
	if !ok || composing != mustTemplateRevisionRef(t, "TPL-1", "REV-1") {
		t.Error("Composing() mismatch; the source must be the composing Revision")
	}
	composed, ok := c.Composed()
	if !ok || composed != mustTemplateRevisionRef(t, "TPL-2", "REV-1") {
		t.Error("Composed() mismatch; the target must be the composed Revision")
	}
	if c.ParameterMapping() == "" || c.ConflictHandling() == "" {
		t.Error("parameter mapping and conflict handling must be preserved")
	}
	if c.Scope() != mustScope(t, "tenant=acme") {
		t.Error("Scope() mismatch")
	}
}

func TestNewCompositionRejections(t *testing.T) {
	composing := mustTemplateRevisionRef(t, "TPL-1", "REV-1")
	composed := mustTemplateRevisionRef(t, "TPL-2", "REV-1")
	scope := mustScope(t, "tenant=acme")
	prov := mustProvenance(t)

	if _, err := NewComposition(core.TemplateArtifactRevisionRef{}, composed, scope, prov, "m", "c"); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero source: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	if _, err := NewComposition(composing, core.TemplateArtifactRevisionRef{}, scope, prov, "m", "c"); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero target: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	if _, err := NewComposition(composing, composed, core.Scope{}, prov, "m", "c"); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
	if _, err := NewComposition(composing, composed, scope, core.Provenance{}, "m", "c"); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	if _, err := NewComposition(composing, composed, scope, prov, "   ", "c"); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("empty parameter mapping: error = %v, want %v", err, ErrInvalidComposition)
	}
	if _, err := NewComposition(composing, composed, scope, prov, "m", "   "); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("empty conflict handling: error = %v, want %v", err, ErrInvalidComposition)
	}
}

// TestCompositionRejectsDegenerateCycle confirms the one cycle a value layer
// can see -- a Revision composing itself -- is rejected. Transitive cycles are
// repository-owned; see doc.go.
func TestCompositionRejectsDegenerateCycle(t *testing.T) {
	same := mustTemplateRevisionRef(t, "TPL-1", "REV-1")
	if _, err := NewComposition(same, same, mustScope(t, "tenant=acme"), mustProvenance(t), "m", "c"); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("self-composition: error = %v, want %v", err, ErrInvalidComposition)
	}

	// Two Revisions of the same Template are a distinct pair and are accepted:
	// PEOS-009 prohibits cycles, not intra-Template composition.
	other := mustTemplateRevisionRef(t, "TPL-1", "REV-2")
	if _, err := NewComposition(same, other, mustScope(t, "tenant=acme"), mustProvenance(t), "m", "c"); err != nil {
		t.Errorf("two revisions of one template: unexpected error %v", err)
	}
}

func TestCompositionJSONRoundTrip(t *testing.T) {
	c := mustComposition(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation", "parameter_mapping", "conflict_handling"} {
		if _, ok := m[key]; !ok {
			t.Errorf("wire form missing %q", key)
		}
	}
	// Multiplicity, direction, and cycle policy are properties of the relation
	// type, constant across every instance, and are documented rather than
	// stored.
	assertNoWireKeys(t, "Composition", c, []string{
		"multiplicity", "direction", "cycles", "cycle_policy", "participant_levels",
		"id", "identity", "lifecycle", "state", "status",
	})

	var decoded Composition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", data2, data)
	}
}

func TestCompositionUnmarshalRejections(t *testing.T) {
	c := mustComposition(t)
	base, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct{ name, key, value string }{
		{"empty parameter mapping", "parameter_mapping", `""`},
		{"whitespace parameter mapping", "parameter_mapping", `"   "`},
		{"absent parameter mapping", "parameter_mapping", ""},
		{"empty conflict handling", "conflict_handling", `""`},
		{"absent conflict handling", "conflict_handling", ""},
		{"null relation", "relation", `null`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			if tt.value == "" {
				delete(mCopy, tt.key)
			} else {
				mCopy[tt.key] = json.RawMessage(tt.value)
			}
			data, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var decoded Composition
			if err := json.Unmarshal(data, &decoded); err == nil {
				t.Errorf("%s accepted, want error", tt.name)
			}
		})
	}

	t.Run("specialization decoded as composition", func(t *testing.T) {
		s := mustSpecialization(t)
		relData, err := json.Marshal(s.Relation())
		if err != nil {
			t.Fatal(err)
		}
		var decoded Composition
		payload := `{"relation":` + string(relData) + `,"parameter_mapping":"m","conflict_handling":"c"}`
		if err := json.Unmarshal([]byte(payload), &decoded); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})

	t.Run("non-object payload", func(t *testing.T) {
		var decoded Composition
		if err := json.Unmarshal([]byte(`123`), &decoded); !errors.Is(err, ErrInvalidComposition) {
			t.Errorf("error = %v, want %v", err, ErrInvalidComposition)
		}
	})
}

// --- Specialization ----------------------------------------------------------

func TestNewSpecialization(t *testing.T) {
	s := mustSpecialization(t)
	if s.IsZero() {
		t.Error("valid Specialization reports IsZero() = true")
	}
	if s.Relation().RelationType() != core.RelationTypeTemplateSpecialization {
		t.Errorf("RelationType() = %v", s.Relation().RelationType())
	}
	specializing, ok := s.Specializing()
	if !ok || specializing != mustTemplateRevisionRef(t, "TPL-1", "REV-1") {
		t.Error("Specializing() mismatch; the source must be the specializing Revision")
	}
	base, ok := s.Base()
	if !ok || base != mustTemplateRevisionRef(t, "TPL-BASE", "REV-1") {
		t.Error("Base() mismatch; the target must be the base Revision")
	}
	if s.InheritedElements() == "" || s.OverriddenElements() == "" || s.CompatibilityEffect() == "" {
		t.Error("all three descriptors must be preserved")
	}
	if s.Scope() != mustScope(t, "tenant=acme") {
		t.Error("Scope() mismatch")
	}
}

func TestNewSpecializationRejections(t *testing.T) {
	specializing := mustTemplateRevisionRef(t, "TPL-1", "REV-1")
	base := mustTemplateRevisionRef(t, "TPL-BASE", "REV-1")
	scope := mustScope(t, "tenant=acme")
	prov := mustProvenance(t)

	if _, err := NewSpecialization(core.TemplateArtifactRevisionRef{}, base, scope, prov, "i", "o", "e"); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero source: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	if _, err := NewSpecialization(specializing, core.TemplateArtifactRevisionRef{}, scope, prov, "i", "o", "e"); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero target: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	if _, err := NewSpecialization(specializing, base, core.Scope{}, prov, "i", "o", "e"); !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("zero scope: error = %v, want %v", err, core.ErrInvalidScope)
	}
	if _, err := NewSpecialization(specializing, base, scope, core.Provenance{}, "i", "o", "e"); !errors.Is(err, ErrInvalidTemplateRelation) {
		t.Errorf("zero provenance: error = %v, want %v", err, ErrInvalidTemplateRelation)
	}
	for _, tt := range []struct {
		name                    string
		inherited, over, effect string
	}{
		{"empty inherited elements", "   ", "o", "e"},
		{"empty overridden elements", "i", "   ", "e"},
		{"empty compatibility effect", "i", "o", "   "},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewSpecialization(specializing, base, scope, prov, tt.inherited, tt.over, tt.effect); !errors.Is(err, ErrInvalidSpecialization) {
				t.Errorf("error = %v, want %v", err, ErrInvalidSpecialization)
			}
		})
	}
}

func TestSpecializationRejectsDegenerateCycle(t *testing.T) {
	same := mustTemplateRevisionRef(t, "TPL-1", "REV-1")
	if _, err := NewSpecialization(same, same, mustScope(t, "tenant=acme"), mustProvenance(t), "i", "o", "e"); !errors.Is(err, ErrInvalidSpecialization) {
		t.Errorf("self-specialization: error = %v, want %v", err, ErrInvalidSpecialization)
	}
}

func TestSpecializationJSONRoundTrip(t *testing.T) {
	s := mustSpecialization(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"relation", "inherited_elements", "overridden_elements", "compatibility_effect"} {
		if _, ok := m[key]; !ok {
			t.Errorf("wire form missing %q", key)
		}
	}
	// compatibility_effect declares an effect; it is never a verdict.
	assertNoWireKeys(t, "Specialization", s, []string{
		"compatible", "compatibility", "conformant", "current", "effective",
		"id", "identity", "lifecycle", "state", "status",
	})

	var decoded Specialization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", data2, data)
	}
}

func TestSpecializationUnmarshalRejections(t *testing.T) {
	base, err := json.Marshal(mustSpecialization(t))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"inherited_elements", "overridden_elements", "compatibility_effect"} {
		t.Run("absent "+key, func(t *testing.T) {
			mCopy := make(map[string]json.RawMessage, len(m))
			for k, v := range m {
				mCopy[k] = v
			}
			delete(mCopy, key)
			data, err := json.Marshal(mCopy)
			if err != nil {
				t.Fatal(err)
			}
			var s Specialization
			if err := json.Unmarshal(data, &s); !errors.Is(err, ErrInvalidSpecialization) {
				t.Errorf("error = %v, want %v", err, ErrInvalidSpecialization)
			}
		})
	}

	t.Run("composition decoded as specialization", func(t *testing.T) {
		c := mustComposition(t)
		relData, err := json.Marshal(c.Relation())
		if err != nil {
			t.Fatal(err)
		}
		var s Specialization
		payload := `{"relation":` + string(relData) + `,"inherited_elements":"i","overridden_elements":"o","compatibility_effect":"e"}`
		if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})

	t.Run("non-object payload", func(t *testing.T) {
		var s Specialization
		if err := json.Unmarshal([]byte(`123`), &s); !errors.Is(err, ErrInvalidSpecialization) {
			t.Errorf("error = %v, want %v", err, ErrInvalidSpecialization)
		}
	})
}

func TestSpecializationMarshalZeroAndExtension(t *testing.T) {
	var zero Specialization
	if !zero.IsZero() {
		t.Error("zero-value Specialization.IsZero() = false, want true")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidSpecialization) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidSpecialization)
	}

	var zeroComposition Composition
	if !zeroComposition.IsZero() {
		t.Error("zero-value Composition.IsZero() = false, want true")
	}
	if _, err := json.Marshal(zeroComposition); !errors.Is(err, ErrInvalidComposition) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidComposition)
	}

	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	s := mustSpecialization(t).WithExtension(ext)
	if s.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !s.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
	c := mustComposition(t).WithExtension(ext)
	if c.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	if !c.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

// TestRelationsHaveNoIdentity is the structural proof of PEOS-009's
// "Generated-From Non-Identity Invariant" and PEOS-002's general contract: none
// of the three relation types carries an identity, a revision, or a lifecycle,
// and neither does the relation.Relation underneath them.
func TestRelationsHaveNoIdentity(t *testing.T) {
	for label, v := range map[string]any{
		"GeneratedFrom":  mustGeneratedFrom(t),
		"Composition":    mustComposition(t),
		"Specialization": mustSpecialization(t),
	} {
		t.Run(label, func(t *testing.T) {
			assertNoWireKeys(t, label, v, []string{
				"id", "identity", "relation_id", "revision", "revision_id",
				"lifecycle", "state", "status", "approval", "validity_period",
			})
			data, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]json.RawMessage
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatal(err)
			}
			var inner map[string]json.RawMessage
			if err := json.Unmarshal(m["relation"], &inner); err != nil {
				t.Fatal(err)
			}
			for _, forbidden := range []string{"id", "identity", "relation_id", "revision", "lifecycle", "state", "status"} {
				if _, ok := inner[forbidden]; ok {
					t.Errorf("inner relation carries forbidden key %q", forbidden)
				}
			}
		})
	}
}

// TestRelationProvenanceAccessors covers the provenance accessor on both
// descriptor-bearing wrappers.
func TestRelationProvenanceAccessors(t *testing.T) {
	if mustComposition(t).Provenance().IsZero() {
		t.Error("Composition.Provenance() is zero")
	}
	if mustSpecialization(t).Provenance().IsZero() {
		t.Error("Specialization.Provenance() is zero")
	}
}

// TestCompositionAndSpecializationDecodeRejectBadParticipants covers the
// participant-level and scope revalidation both decoders perform, using a
// relation of the right type but the wrong participant levels or no scope.
func TestCompositionAndSpecializationDecodeRejectBadParticipants(t *testing.T) {
	generated, err := core.EngineeringSubjectRefFromGeneratedArtifactRevision(mustGeneratedRevisionRef(t, "GEN-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	template, err := core.EngineeringSubjectRefFromTemplateRevision(mustTemplateRevisionRef(t, "TPL-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	other, err := core.EngineeringSubjectRefFromTemplateRevision(mustTemplateRevisionRef(t, "TPL-2", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	scope := mustScope(t, "tenant=acme")

	build := func(t *testing.T, relType core.RelationType, source, target core.EngineeringSubjectRef, withScope bool) string {
		t.Helper()
		rel, err := relation.New(relType, source, target, mustProvenance(t))
		if err != nil {
			t.Fatal(err)
		}
		if withScope {
			if rel, err = rel.WithScope(scope); err != nil {
				t.Fatal(err)
			}
		}
		data, err := json.Marshal(rel)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	t.Run("composition with a generated-artifact source", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateComposition, generated, template, true)
		var c Composition
		payload := `{"relation":` + rel + `,"parameter_mapping":"m","conflict_handling":"h"}`
		if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
	t.Run("composition with a generated-artifact target", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateComposition, template, generated, true)
		var c Composition
		payload := `{"relation":` + rel + `,"parameter_mapping":"m","conflict_handling":"h"}`
		if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
	t.Run("composition without scope", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateComposition, template, other, false)
		var c Composition
		payload := `{"relation":` + rel + `,"parameter_mapping":"m","conflict_handling":"h"}`
		if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
	t.Run("composition that is a degenerate cycle", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateComposition, template, template, true)
		var c Composition
		payload := `{"relation":` + rel + `,"parameter_mapping":"m","conflict_handling":"h"}`
		if err := json.Unmarshal([]byte(payload), &c); !errors.Is(err, ErrInvalidComposition) {
			t.Errorf("error = %v, want %v", err, ErrInvalidComposition)
		}
	})
	t.Run("specialization with a generated-artifact source", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateSpecialization, generated, template, true)
		var s Specialization
		payload := `{"relation":` + rel + `,"inherited_elements":"i","overridden_elements":"o","compatibility_effect":"e"}`
		if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
	t.Run("specialization with a generated-artifact target", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateSpecialization, template, generated, true)
		var s Specialization
		payload := `{"relation":` + rel + `,"inherited_elements":"i","overridden_elements":"o","compatibility_effect":"e"}`
		if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
	t.Run("specialization without scope", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateSpecialization, template, other, false)
		var s Specialization
		payload := `{"relation":` + rel + `,"inherited_elements":"i","overridden_elements":"o","compatibility_effect":"e"}`
		if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
	t.Run("specialization that is a degenerate cycle", func(t *testing.T) {
		rel := build(t, core.RelationTypeTemplateSpecialization, template, template, true)
		var s Specialization
		payload := `{"relation":` + rel + `,"inherited_elements":"i","overridden_elements":"o","compatibility_effect":"e"}`
		if err := json.Unmarshal([]byte(payload), &s); !errors.Is(err, ErrInvalidSpecialization) {
			t.Errorf("error = %v, want %v", err, ErrInvalidSpecialization)
		}
	})
	t.Run("generated-from with a template source is rejected on decode", func(t *testing.T) {
		rel := build(t, core.RelationTypeGeneratedFrom, generated, generated, true)
		var g GeneratedFrom
		if err := json.Unmarshal([]byte(`{"relation":`+rel+`}`), &g); !errors.Is(err, ErrInvalidTemplateRelation) {
			t.Errorf("generated target: error = %v, want %v", err, ErrInvalidTemplateRelation)
		}
	})
}

// TestCompositionAndSpecializationPreserveExtensionOnDecode covers the
// extension carry-over both decoders perform.
func TestCompositionAndSpecializationPreserveExtensionOnDecode(t *testing.T) {
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}

	c := mustComposition(t).WithExtension(ext)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decodedC Composition
	if err := json.Unmarshal(data, &decodedC); err != nil {
		t.Fatal(err)
	}
	if decodedC.Extension().IsZero() {
		t.Error("Composition round trip lost the extension")
	}

	s := mustSpecialization(t).WithExtension(ext)
	data, err = json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decodedS Specialization
	if err := json.Unmarshal(data, &decodedS); err != nil {
		t.Fatal(err)
	}
	if decodedS.Extension().IsZero() {
		t.Error("Specialization round trip lost the extension")
	}
}
