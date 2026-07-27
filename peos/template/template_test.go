package template

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

// --- helpers -----------------------------------------------------------------

func mustArtifactID(t *testing.T, value string) core.ArtifactID {
	t.Helper()
	id, err := core.NewArtifactID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustArtifactRevisionID(t *testing.T, value string) core.ArtifactRevisionID {
	t.Helper()
	id, err := core.NewArtifactRevisionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustLocalKey(t *testing.T, value string) core.LocalKey {
	t.Helper()
	k, err := core.NewLocalKey(value)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func mustVocabularyValue(t *testing.T, namespace, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustArtifactType(t *testing.T, value string) core.ArtifactType {
	t.Helper()
	return core.NewArtifactType(mustVocabularyValue(t, core.PEOSNamespace, value))
}

func mustScope(t *testing.T, expression string) core.Scope {
	t.Helper()
	s, err := core.NewScope(mustVocabularyValue(t, core.PEOSNamespace, "component"), expression)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustOrigin(t *testing.T) core.Origin {
	t.Helper()
	o, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustProvenance(t *testing.T) core.Provenance {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	actor, err := core.NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewProvenance().WithActor(actor).WithRecordedAt(ts)
}

func mustAuthority(t *testing.T) core.AuthorityRef {
	t.Helper()
	a, err := core.NewAuthorityRef("peos", "template-board")
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustIntegrity(t *testing.T) core.IntegrityIdentity {
	t.Helper()
	i, err := core.NewIntegrityIdentity(core.IntegrityMechanismCryptographicDigest, "sha256:abc123", core.IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func mustArtifact(t *testing.T, id string, artifactType core.ArtifactType) core.Artifact {
	t.Helper()
	a, err := core.NewArtifact(mustArtifactID(t, id), artifactType)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustTemplate(t *testing.T, id string) Template {
	t.Helper()
	tmpl, err := NewTemplate(mustArtifact(t, id, ArtifactTypeTemplate))
	if err != nil {
		t.Fatal(err)
	}
	return tmpl
}

func mustArtifactRevision(t *testing.T, artifactID, revisionID string) core.ArtifactRevision {
	t.Helper()
	r, err := core.NewArtifactRevision(
		mustArtifactID(t, artifactID),
		mustArtifactRevisionID(t, revisionID),
		mustOrigin(t),
		mustProvenance(t),
		mustIntegrity(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func mustTemplateRevisionRef(t *testing.T, artifactID, revisionID string) core.TemplateArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewTemplateArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustCompatibilityDeclaration(t *testing.T) CompatibilityDeclaration {
	t.Helper()
	d, err := NewCompatibilityDeclaration(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"all required parameters supplied",
		"product contract v1",
	)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func mustParameterType(t *testing.T) ParameterType {
	t.Helper()
	pt, err := NewVocabularyParameterType(mustVocabularyValue(t, "product", "string"))
	if err != nil {
		t.Fatal(err)
	}
	return pt
}

func mustParameter(t *testing.T, key string, required bool) Parameter {
	t.Helper()
	p, err := NewParameter(mustLocalKey(t, key), mustParameterType(t), required)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func mustParameterDefault(t *testing.T, parameterKey, value string) ParameterDefault {
	t.Helper()
	d, err := NewParameterDefault(mustLocalKey(t, parameterKey), value)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func mustEvaluationPoint(t *testing.T) ConstraintEvaluationPoint {
	t.Helper()
	return NewConstraintEvaluationPoint(mustVocabularyValue(t, "product", "pre-generation"))
}

func mustFailureSemantics(t *testing.T) ConstraintFailureSemantics {
	t.Helper()
	return NewConstraintFailureSemantics(mustVocabularyValue(t, "product", "reject"))
}

func mustParameterConstraint(t *testing.T, key, parameterKey string) ParameterConstraint {
	t.Helper()
	target, err := NewParameterConstraintTarget(mustLocalKey(t, parameterKey))
	if err != nil {
		t.Fatal(err)
	}
	v, err := NewParameterConstraint(
		mustLocalKey(t, key),
		target,
		"value must be non-empty",
		mustScope(t, "env=prod"),
		mustEvaluationPoint(t),
		mustFailureSemantics(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustGeneratedContentConstraint(t *testing.T, key, descriptor string) ParameterConstraint {
	t.Helper()
	target, err := NewGeneratedContentConstraintTarget(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	v, err := NewParameterConstraint(
		mustLocalKey(t, key),
		target,
		"generated content must declare a statement",
		mustScope(t, "env=prod"),
		mustEvaluationPoint(t),
		mustFailureSemantics(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// mustMinimalTemplateContent builds the smallest valid TemplateContent: one
// permitted generated Artifact Type, expansion semantics, a compatibility
// declaration, explicit applicability, and provenance -- with zero parameters,
// zero defaults, zero constraints, and no authority.
func mustMinimalTemplateContent(t *testing.T) TemplateContent {
	t.Helper()
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand each parameter into the requirement statement",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		nil, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// --- ArtifactTypeTemplate ----------------------------------------------------

func TestArtifactTypeTemplate(t *testing.T) {
	if ArtifactTypeTemplate.IsZero() {
		t.Fatal("ArtifactTypeTemplate is zero")
	}
	if got, want := ArtifactTypeTemplate.String(), "peos:template"; got != want {
		t.Errorf("ArtifactTypeTemplate = %q, want %q", got, want)
	}
}

// --- Template ----------------------------------------------------------------

func TestNewTemplate(t *testing.T) {
	artifact := mustArtifact(t, "TPL-1", ArtifactTypeTemplate)
	tmpl, err := NewTemplate(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.IsZero() {
		t.Error("valid Template reports IsZero() = true")
	}
	if tmpl.ID() != mustArtifactID(t, "TPL-1") {
		t.Errorf("ID() = %v", tmpl.ID())
	}
	if tmpl.Core().ID() != artifact.ID() || tmpl.Core().Type() != artifact.Type() {
		t.Error("Core() mismatch")
	}
	ref, err := tmpl.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != tmpl.ID() {
		t.Error("Ref() artifact id mismatch")
	}

	var zero Template
	if !zero.IsZero() {
		t.Error("zero-value Template.IsZero() = false, want true")
	}
}

func TestNewTemplateRejections(t *testing.T) {
	if _, err := NewTemplate(core.Artifact{}); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("zero artifact: error = %v, want %v", err, ErrInvalidTemplate)
	}
	wrong := mustArtifact(t, "TPL-1", mustArtifactType(t, "requirement"))
	if _, err := NewTemplate(wrong); !errors.Is(err, ErrTemplateArtifactTypeMismatch) {
		t.Errorf("wrong artifact type: error = %v, want %v", err, ErrTemplateArtifactTypeMismatch)
	}
}

func TestTemplateJSONRoundTrip(t *testing.T) {
	tmpl := mustTemplate(t, "TPL-1")
	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Template
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != tmpl.ID() {
		t.Error("round trip lost identity")
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
}

func TestTemplateMarshalZero(t *testing.T) {
	var tmpl Template
	if _, err := json.Marshal(tmpl); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidTemplate)
	}
}

func TestTemplateUnmarshalRejectionsPreserveReceiver(t *testing.T) {
	tmpl := mustTemplate(t, "TPL-1")
	original, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatal(err)
	}

	wrongType, err := json.Marshal(mustArtifact(t, "TPL-1", mustArtifactType(t, "requirement")))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wrongType, &tmpl); !errors.Is(err, ErrTemplateArtifactTypeMismatch) {
		t.Errorf("wrong artifact type: error = %v, want %v", err, ErrTemplateArtifactTypeMismatch)
	}
	if err := json.Unmarshal([]byte(`null`), &tmpl); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("explicit null: error = %v, want %v", err, ErrInvalidTemplate)
	}
	if err := json.Unmarshal([]byte(`not json`), &tmpl); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
	after, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- TemplateApplicability ---------------------------------------------------

func TestTemplateApplicabilityUnrestricted(t *testing.T) {
	a := NewUnrestrictedTemplateApplicability()
	if a.IsZero() {
		t.Error("explicit unrestricted applicability reports IsZero() = true; it is a stated value")
	}
	if !a.IsUnrestricted() || a.IsScoped() {
		t.Error("discriminator mismatch")
	}
	if _, ok := a.Scope(); ok {
		t.Error("unrestricted applicability returned a scope")
	}
	if a.Kind() != "unrestricted" {
		t.Errorf("Kind() = %q", a.Kind())
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"kind":"unrestricted"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded TemplateApplicability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != a {
		t.Error("round trip mismatch")
	}
}

func TestTemplateApplicabilityScoped(t *testing.T) {
	scope := mustScope(t, "tenant=acme")
	a, err := NewScopedTemplateApplicability(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !a.IsScoped() || a.IsUnrestricted() {
		t.Error("discriminator mismatch")
	}
	got, ok := a.Scope()
	if !ok || got != scope {
		t.Error("Scope() mismatch")
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TemplateApplicability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != a {
		t.Error("round trip mismatch")
	}
}

func TestTemplateApplicabilityZeroInvalid(t *testing.T) {
	var a TemplateApplicability
	if !a.IsZero() {
		t.Error("zero-value TemplateApplicability.IsZero() = false, want true")
	}
	if a.Kind() != "" {
		t.Errorf("zero Kind() = %q, want empty", a.Kind())
	}
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidTemplateApplicability) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidTemplateApplicability)
	}
	if _, err := NewScopedTemplateApplicability(core.Scope{}); !errors.Is(err, ErrInvalidTemplateApplicability) {
		t.Errorf("zero scope: error = %v, want %v", err, ErrInvalidTemplateApplicability)
	}
}

func TestTemplateApplicabilityUnmarshalRejections(t *testing.T) {
	// A real, valid core.Scope wire form, so the "unrestricted carrying scope"
	// case is rejected by this package's own union rule rather than by
	// core.Scope's decoder failing first.
	scopeJSON, err := json.Marshal(mustScope(t, "tenant=acme"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		json string
	}{
		{"missing kind", `{}`},
		{"unknown kind", `{"kind":"bogus"}`},
		{"explicit null", `null`},
		{"unrestricted carrying scope", `{"kind":"unrestricted","scope":` + string(scopeJSON) + `}`},
		{"scoped missing scope", `{"kind":"scoped"}`},
		{"scoped null scope", `{"kind":"scoped","scope":null}`},
		{"scoped malformed scope", `{"kind":"scoped","scope":123}`},
		{"non-object payload", `123`},
		{"malformed JSON", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a TemplateApplicability
			if err := json.Unmarshal([]byte(tt.json), &a); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

// --- TemplateContent: cardinality --------------------------------------------

// TestTemplateContentMinimalValid proves PEOS-009's only minimum cardinality
// is >=1 generated Artifact Type: a Template Revision with zero parameters,
// zero defaults, zero constraints, zero composition and specialization
// references, and no authority is valid and round-trippable.
func TestTemplateContentMinimalValid(t *testing.T) {
	c := mustMinimalTemplateContent(t)
	if c.IsZero() {
		t.Error("minimal valid TemplateContent reports IsZero() = true")
	}
	if len(c.Parameters()) != 0 || len(c.Defaults()) != 0 || len(c.Constraints()) != 0 {
		t.Error("minimal content should declare no parameters, defaults, or constraints")
	}
	if len(c.CompositionReferences()) != 0 || len(c.SpecializationReferences()) != 0 {
		t.Error("minimal content should declare no composition or specialization references")
	}
	if _, ok := c.Authority(); ok {
		t.Error("minimal content should declare no authority")
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TemplateContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("minimal content failed to round-trip: %v", err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
}

// TestTemplateContentOptionalStateMatrix adds each optional collection and
// scalar to the minimal valid content independently, proving none of them is
// required and none interferes with another.
func TestTemplateContentOptionalStateMatrix(t *testing.T) {
	base := func() (
		[]core.ArtifactType, string, CompatibilityDeclaration, TemplateApplicability, core.Provenance,
	) {
		return []core.ArtifactType{mustArtifactType(t, "requirement")},
			"expand parameters",
			mustCompatibilityDeclaration(t),
			NewUnrestrictedTemplateApplicability(),
			mustProvenance(t)
	}

	t.Run("parameters only", func(t *testing.T) {
		at, es, cd, ap, pr := base()
		if _, err := NewTemplateContent(at, es, cd, ap, pr, []Parameter{mustParameter(t, "name", true)}, nil, nil); err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("parameters and defaults", func(t *testing.T) {
		at, es, cd, ap, pr := base()
		_, err := NewTemplateContent(at, es, cd, ap, pr,
			[]Parameter{mustParameter(t, "name", false)},
			[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
			nil,
		)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("parameters and constraints", func(t *testing.T) {
		at, es, cd, ap, pr := base()
		_, err := NewTemplateContent(at, es, cd, ap, pr,
			[]Parameter{mustParameter(t, "name", true)},
			nil,
			[]ParameterConstraint{mustParameterConstraint(t, "name-nonempty", "name")},
		)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("generated-content constraint with zero parameters", func(t *testing.T) {
		at, es, cd, ap, pr := base()
		_, err := NewTemplateContent(at, es, cd, ap, pr, nil, nil,
			[]ParameterConstraint{mustGeneratedContentConstraint(t, "stmt-present", "the requirement statement")},
		)
		if err != nil {
			t.Errorf("a generated-content constraint must not require any parameter: %v", err)
		}
	})
	t.Run("multiple generated artifact types", func(t *testing.T) {
		_, es, cd, ap, pr := base()
		types := []core.ArtifactType{mustArtifactType(t, "requirement"), mustArtifactType(t, "runtime-contract")}
		if _, err := NewTemplateContent(types, es, cd, ap, pr, nil, nil, nil); err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("composition references", func(t *testing.T) {
		c := mustMinimalTemplateContent(t)
		if _, err := c.WithCompositionReferences([]core.TemplateArtifactRevisionRef{mustTemplateRevisionRef(t, "TPL-2", "REV-1")}); err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("specialization references", func(t *testing.T) {
		c := mustMinimalTemplateContent(t)
		if _, err := c.WithSpecializationReferences([]core.TemplateArtifactRevisionRef{mustTemplateRevisionRef(t, "TPL-3", "REV-1")}); err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("authority", func(t *testing.T) {
		c := mustMinimalTemplateContent(t)
		c2, err := c.WithAuthority(mustAuthority(t))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := c2.Authority(); !ok {
			t.Error("WithAuthority did not set authority")
		}
		if _, ok := c2.WithoutAuthority().Authority(); ok {
			t.Error("WithoutAuthority did not clear authority")
		}
	})
	t.Run("extension", func(t *testing.T) {
		c := mustMinimalTemplateContent(t)
		ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
		if err != nil {
			t.Fatal(err)
		}
		c = c.WithExtension(ext)
		if c.Extension().IsZero() {
			t.Error("WithExtension did not set extension")
		}
		if !c.WithoutExtension().Extension().IsZero() {
			t.Error("WithoutExtension did not clear extension")
		}
	})
}

func TestTemplateContentMandatoryFieldRejections(t *testing.T) {
	valid := func() (
		[]core.ArtifactType, string, CompatibilityDeclaration, TemplateApplicability, core.Provenance,
	) {
		return []core.ArtifactType{mustArtifactType(t, "requirement")},
			"expand parameters",
			mustCompatibilityDeclaration(t),
			NewUnrestrictedTemplateApplicability(),
			mustProvenance(t)
	}

	t.Run("zero generated artifact types", func(t *testing.T) {
		_, es, cd, ap, pr := valid()
		if _, err := NewTemplateContent(nil, es, cd, ap, pr, nil, nil, nil); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("zero-value generated artifact type element", func(t *testing.T) {
		_, es, cd, ap, pr := valid()
		if _, err := NewTemplateContent([]core.ArtifactType{{}}, es, cd, ap, pr, nil, nil, nil); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("duplicate generated artifact type", func(t *testing.T) {
		_, es, cd, ap, pr := valid()
		types := []core.ArtifactType{mustArtifactType(t, "requirement"), mustArtifactType(t, "requirement")}
		if _, err := NewTemplateContent(types, es, cd, ap, pr, nil, nil, nil); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("empty expansion semantics", func(t *testing.T) {
		at, _, cd, ap, pr := valid()
		if _, err := NewTemplateContent(at, "   ", cd, ap, pr, nil, nil, nil); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("zero compatibility declaration", func(t *testing.T) {
		at, es, _, ap, pr := valid()
		if _, err := NewTemplateContent(at, es, CompatibilityDeclaration{}, ap, pr, nil, nil, nil); !errors.Is(err, ErrInvalidCompatibilityDeclaration) {
			t.Errorf("error = %v, want %v", err, ErrInvalidCompatibilityDeclaration)
		}
	})
	t.Run("zero applicability", func(t *testing.T) {
		at, es, cd, _, pr := valid()
		if _, err := NewTemplateContent(at, es, cd, TemplateApplicability{}, pr, nil, nil, nil); !errors.Is(err, ErrInvalidTemplateApplicability) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateApplicability)
		}
	})
	t.Run("zero provenance", func(t *testing.T) {
		at, es, cd, ap, _ := valid()
		if _, err := NewTemplateContent(at, es, cd, ap, core.Provenance{}, nil, nil, nil); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("zero-value parameter element", func(t *testing.T) {
		at, es, cd, ap, pr := valid()
		if _, err := NewTemplateContent(at, es, cd, ap, pr, []Parameter{{}}, nil, nil); !errors.Is(err, ErrInvalidTemplateParameter) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateParameter)
		}
	})
	t.Run("zero-value default element", func(t *testing.T) {
		at, es, cd, ap, pr := valid()
		if _, err := NewTemplateContent(at, es, cd, ap, pr, nil, []ParameterDefault{{}}, nil); !errors.Is(err, ErrInvalidParameterDefault) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterDefault)
		}
	})
	t.Run("zero-value constraint element", func(t *testing.T) {
		at, es, cd, ap, pr := valid()
		if _, err := NewTemplateContent(at, es, cd, ap, pr, nil, nil, []ParameterConstraint{{}}); !errors.Is(err, ErrInvalidParameterConstraint) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterConstraint)
		}
	})
	t.Run("zero authority via modifier", func(t *testing.T) {
		c := mustMinimalTemplateContent(t)
		if _, err := c.WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
}

// --- namespaces --------------------------------------------------------------

func TestTemplateContentParameterNamespace(t *testing.T) {
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true), mustParameter(t, "owner", false)},
		nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := c.Parameter(mustLocalKey(t, "name"))
	if !ok || got.Key() != mustLocalKey(t, "name") {
		t.Error("Parameter(name) lookup failed")
	}
	if !got.Required() {
		t.Error("Parameter(name).Required() = false, want true")
	}
	optional, ok := c.Parameter(mustLocalKey(t, "owner"))
	if !ok || optional.Required() {
		t.Error("Parameter(owner) should exist and be explicitly optional")
	}
	if _, ok := c.Parameter(core.LocalKey{}); ok {
		t.Error("Parameter(zero key) should return ok=false")
	}
	if _, ok := c.Parameter(mustLocalKey(t, "does-not-exist")); ok {
		t.Error("Parameter(unknown key) should return ok=false")
	}
}

func TestTemplateContentDuplicateParameterKeyRejected(t *testing.T) {
	_, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true), mustParameter(t, "name", false)},
		nil, nil,
	)
	if !errors.Is(err, ErrDuplicateTemplateLocalKey) {
		t.Errorf("duplicate parameter key: error = %v, want %v", err, ErrDuplicateTemplateLocalKey)
	}
}

func TestTemplateContentConstraintNamespace(t *testing.T) {
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true)},
		nil,
		[]ParameterConstraint{
			mustParameterConstraint(t, "name-nonempty", "name"),
			mustGeneratedContentConstraint(t, "stmt-present", "the requirement statement"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := c.Constraint(mustLocalKey(t, "name-nonempty"))
	if !ok || got.Key() != mustLocalKey(t, "name-nonempty") {
		t.Error("Constraint(name-nonempty) lookup failed")
	}
	if _, ok := c.Constraint(core.LocalKey{}); ok {
		t.Error("Constraint(zero key) should return ok=false")
	}
	if _, ok := c.Constraint(mustLocalKey(t, "does-not-exist")); ok {
		t.Error("Constraint(unknown key) should return ok=false")
	}
}

func TestTemplateContentDuplicateConstraintKeyRejected(t *testing.T) {
	_, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true)},
		nil,
		[]ParameterConstraint{
			mustParameterConstraint(t, "dup", "name"),
			mustParameterConstraint(t, "dup", "name"),
		},
	)
	if !errors.Is(err, ErrDuplicateTemplateLocalKey) {
		t.Errorf("duplicate constraint key: error = %v, want %v", err, ErrDuplicateTemplateLocalKey)
	}
}

// TestTemplateContentParameterAndConstraintMayShareKey documents that the
// parameter namespace and the constraint namespace are independent: PEOS-009
// states no cross-collection uniqueness rule, and the two are addressed by
// genuinely different reference kinds (a default or constraint target names a
// parameter; core.CriterionKindTemplateConstraint names a constraint).
func TestTemplateContentParameterAndConstraintMayShareKey(t *testing.T) {
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "shared-key", true)},
		nil,
		[]ParameterConstraint{mustParameterConstraint(t, "shared-key", "shared-key")},
	)
	if err != nil {
		t.Fatalf("parameter and constraint sharing a key: unexpected error %v", err)
	}
	p, ok := c.Parameter(mustLocalKey(t, "shared-key"))
	if !ok {
		t.Error("Parameter(shared-key) lookup failed")
	}
	v, ok := c.Constraint(mustLocalKey(t, "shared-key"))
	if !ok {
		t.Error("Constraint(shared-key) lookup failed")
	}
	if p.Key() != v.Key() {
		t.Error("the two namespaces should resolve the same key to their own value")
	}
}

// TestTemplateConstraintCriterionRefResolves is the end-to-end proof that
// core.CriterionKindTemplateConstraint has a resolvable target: a criterion
// built from a Template Revision's Ref plus a constraint's local key resolves
// back to exactly one ParameterConstraint through Constraint(key).
func TestTemplateConstraintCriterionRefResolves(t *testing.T) {
	tmpl := mustTemplate(t, "TPL-1")
	content, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true)},
		nil,
		[]ParameterConstraint{mustParameterConstraint(t, "name-nonempty", "name")},
	)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := NewTemplateRevision(tmpl, mustArtifactRevision(t, "TPL-1", "REV-1"), content)
	if err != nil {
		t.Fatal(err)
	}
	revRef, err := revision.Ref()
	if err != nil {
		t.Fatal(err)
	}

	criterionRef, err := core.NewTemplateConstraintCriterionRef(revRef, mustLocalKey(t, "name-nonempty"))
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := core.CriterionRefFromTemplateConstraint(criterionRef)
	if err != nil {
		t.Fatal(err)
	}
	if criterion.Kind() != core.CriterionKindTemplateConstraint {
		t.Fatalf("criterion kind = %q, want %q", criterion.Kind(), core.CriterionKindTemplateConstraint)
	}

	payload, ok := criterion.AsTemplateConstraint()
	if !ok {
		t.Fatal("AsTemplateConstraint() failed")
	}
	if payload.Template() != revRef {
		t.Error("criterion names a different Template Revision")
	}
	resolved, ok := revision.Content().Constraint(payload.Constraint())
	if !ok {
		t.Fatal("the criterion's local key did not resolve to a constraint")
	}
	if resolved.Key() != mustLocalKey(t, "name-nonempty") {
		t.Errorf("resolved constraint key = %v", resolved.Key())
	}
}

// --- cross-reference validation ----------------------------------------------

func TestTemplateContentDefaultCrossReferences(t *testing.T) {
	build := func(params []Parameter, defaults []ParameterDefault) (TemplateContent, error) {
		return NewTemplateContent(
			[]core.ArtifactType{mustArtifactType(t, "requirement")},
			"expand parameters",
			mustCompatibilityDeclaration(t),
			NewUnrestrictedTemplateApplicability(),
			mustProvenance(t),
			params, defaults, nil,
		)
	}

	t.Run("default resolves to a declared parameter", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", false)},
			[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
		)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("default names an undeclared parameter", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", false)},
			[]ParameterDefault{mustParameterDefault(t, "missing", "x")},
		)
		if !errors.Is(err, ErrUnknownTemplateLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrUnknownTemplateLocalKey)
		}
	})
	t.Run("two defaults for one parameter", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", false)},
			[]ParameterDefault{
				mustParameterDefault(t, "name", "first"),
				mustParameterDefault(t, "name", "second"),
			},
		)
		if !errors.Is(err, ErrInvalidParameterDefault) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterDefault)
		}
	})
	t.Run("default targeting a parameter that forbids default resolution", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", true).WithForbiddenDefaultResolution()},
			[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
		)
		if !errors.Is(err, ErrInvalidParameterDefault) {
			t.Errorf("error = %v, want %v", err, ErrInvalidParameterDefault)
		}
	})
	t.Run("permitted default resolution restores acceptance", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", true).WithForbiddenDefaultResolution().WithPermittedDefaultResolution()},
			[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
		)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
}

func TestTemplateContentConstraintCrossReferences(t *testing.T) {
	build := func(params []Parameter, constraints []ParameterConstraint) (TemplateContent, error) {
		return NewTemplateContent(
			[]core.ArtifactType{mustArtifactType(t, "requirement")},
			"expand parameters",
			mustCompatibilityDeclaration(t),
			NewUnrestrictedTemplateApplicability(),
			mustProvenance(t),
			params, nil, constraints,
		)
	}

	t.Run("parameter-targeting constraint resolves", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", true)},
			[]ParameterConstraint{mustParameterConstraint(t, "c1", "name")},
		)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
	t.Run("parameter-targeting constraint names an undeclared parameter", func(t *testing.T) {
		_, err := build(
			[]Parameter{mustParameter(t, "name", true)},
			[]ParameterConstraint{mustParameterConstraint(t, "c1", "missing")},
		)
		if !errors.Is(err, ErrUnknownTemplateLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrUnknownTemplateLocalKey)
		}
	})
	t.Run("generated-content constraint needs no parameter at all", func(t *testing.T) {
		_, err := build(nil, []ParameterConstraint{mustGeneratedContentConstraint(t, "c1", "the statement")})
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
}

func TestTemplateContentRevisionReferenceRejections(t *testing.T) {
	c := mustMinimalTemplateContent(t)
	dup := mustTemplateRevisionRef(t, "TPL-2", "REV-1")

	if _, err := c.WithCompositionReferences([]core.TemplateArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("zero composition reference: error = %v, want %v", err, ErrInvalidTemplateContent)
	}
	if _, err := c.WithCompositionReferences([]core.TemplateArtifactRevisionRef{dup, dup}); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("duplicate composition reference: error = %v, want %v", err, ErrInvalidTemplateContent)
	}
	if _, err := c.WithSpecializationReferences([]core.TemplateArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("zero specialization reference: error = %v, want %v", err, ErrInvalidTemplateContent)
	}
	if _, err := c.WithSpecializationReferences([]core.TemplateArtifactRevisionRef{dup, dup}); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("duplicate specialization reference: error = %v, want %v", err, ErrInvalidTemplateContent)
	}
}

// --- accessors and defensive copying ----------------------------------------

func TestTemplateContentAccessors(t *testing.T) {
	types := []core.ArtifactType{mustArtifactType(t, "requirement")}
	compat := mustCompatibilityDeclaration(t)
	applicability := NewUnrestrictedTemplateApplicability()
	provenance := mustProvenance(t)

	c, err := NewTemplateContent(
		types, "  expand parameters  ", compat, applicability, provenance,
		[]Parameter{mustParameter(t, "name", true)},
		[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
		[]ParameterConstraint{mustParameterConstraint(t, "c1", "name")},
	)
	if err != nil {
		t.Fatal(err)
	}

	if got := c.GeneratedArtifactTypes(); len(got) != 1 || got[0] != types[0] {
		t.Errorf("GeneratedArtifactTypes() = %v", got)
	}
	if c.ExpansionSemantics() != "expand parameters" {
		t.Errorf("ExpansionSemantics() = %q, want trimmed", c.ExpansionSemantics())
	}
	if c.Compatibility().ParameterContract() != compat.ParameterContract() {
		t.Error("Compatibility() mismatch")
	}
	if c.Applicability() != applicability {
		t.Error("Applicability() mismatch")
	}
	if c.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if got := c.Parameters(); len(got) != 1 || got[0].Key() != mustLocalKey(t, "name") {
		t.Errorf("Parameters() = %v", got)
	}
	if got := c.Defaults(); len(got) != 1 || got[0].Value() != "anonymous" {
		t.Errorf("Defaults() = %v", got)
	}
	if got := c.Constraints(); len(got) != 1 || got[0].Key() != mustLocalKey(t, "c1") {
		t.Errorf("Constraints() = %v", got)
	}
}

func TestTemplateContentDefensiveCopy(t *testing.T) {
	types := []core.ArtifactType{mustArtifactType(t, "requirement")}
	params := []Parameter{mustParameter(t, "name", true)}
	defaults := []ParameterDefault{mustParameterDefault(t, "name", "anonymous")}
	constraints := []ParameterConstraint{mustParameterConstraint(t, "c1", "name")}

	c, err := NewTemplateContent(
		types, "expand", mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(), mustProvenance(t),
		params, defaults, constraints,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Mutating the constructor inputs must not affect the stored value.
	types[0] = mustArtifactType(t, "mutated")
	params[0] = mustParameter(t, "mutated", false)
	defaults[0] = mustParameterDefault(t, "mutated", "mutated")
	constraints[0] = mustParameterConstraint(t, "mutated", "mutated")

	if c.GeneratedArtifactTypes()[0].String() == "peos:mutated" {
		t.Error("constructor did not defensively copy generated artifact types")
	}
	if c.Parameters()[0].Key() == mustLocalKey(t, "mutated") {
		t.Error("constructor did not defensively copy parameters")
	}
	if c.Defaults()[0].Parameter() == mustLocalKey(t, "mutated") {
		t.Error("constructor did not defensively copy defaults")
	}
	if c.Constraints()[0].Key() == mustLocalKey(t, "mutated") {
		t.Error("constructor did not defensively copy constraints")
	}

	// Mutating a returned slice must not affect the stored value either.
	returned := c.Parameters()
	returned[0] = mustParameter(t, "mutated-again", false)
	if c.Parameters()[0].Key() == mustLocalKey(t, "mutated-again") {
		t.Error("Parameters() accessor did not return a defensive copy")
	}

	c2, err := c.WithCompositionReferences([]core.TemplateArtifactRevisionRef{mustTemplateRevisionRef(t, "TPL-2", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	refs := c2.CompositionReferences()
	refs[0] = mustTemplateRevisionRef(t, "TPL-9", "REV-9")
	if c2.CompositionReferences()[0].ArtifactID() == mustArtifactID(t, "TPL-9") {
		t.Error("CompositionReferences() accessor did not return a defensive copy")
	}
}

// TestTemplateContentFailedModifierLeavesReceiverUnchanged confirms a rejected
// modifier returns the zero value and never mutates its receiver.
func TestTemplateContentFailedModifierLeavesReceiverUnchanged(t *testing.T) {
	c := mustMinimalTemplateContent(t)
	before, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.WithCompositionReferences([]core.TemplateArtifactRevisionRef{{}}); err == nil {
		t.Fatal("expected error")
	}
	after, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("failed modifier mutated its receiver")
	}
}

// --- TemplateContent JSON ----------------------------------------------------

func TestTemplateContentJSONRoundTrip(t *testing.T) {
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement"), mustArtifactType(t, "runtime-contract")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true), mustParameter(t, "owner", false)},
		[]ParameterDefault{mustParameterDefault(t, "owner", "unassigned")},
		[]ParameterConstraint{
			mustParameterConstraint(t, "c1", "name"),
			mustGeneratedContentConstraint(t, "c2", "the statement"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithCompositionReferences([]core.TemplateArtifactRevisionRef{mustTemplateRevisionRef(t, "TPL-2", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithSpecializationReferences([]core.TemplateArtifactRevisionRef{mustTemplateRevisionRef(t, "TPL-3", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	c = c.WithExtension(ext)

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TemplateContent
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
	if _, ok := decoded.Authority(); !ok {
		t.Error("round trip lost authority")
	}
	if len(decoded.Parameters()) != 2 || len(decoded.Defaults()) != 1 || len(decoded.Constraints()) != 2 {
		t.Error("round trip lost collection elements")
	}
}

func TestTemplateContentMarshalZero(t *testing.T) {
	var c TemplateContent
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidTemplateContent) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidTemplateContent)
	}
	if !c.IsZero() {
		t.Error("zero-value TemplateContent.IsZero() = false, want true")
	}
}

// TestTemplateContentCollectionJSONEquivalence verifies that for every
// optional collection, an absent key, an explicit null, and an empty array are
// all equivalent and all mean "declares none of this kind".
func TestTemplateContentCollectionJSONEquivalence(t *testing.T) {
	fields := []string{"parameters", "defaults", "constraints", "composition_references", "specialization_references"}

	baseMap := func(t *testing.T) map[string]json.RawMessage {
		t.Helper()
		data, err := json.Marshal(mustMinimalTemplateContent(t))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			for _, tc := range []struct{ name, value string }{
				{"absent", ""},
				{"null", "null"},
				{"empty array", "[]"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					m := baseMap(t)
					if tc.value == "" {
						delete(m, field)
					} else {
						m[field] = json.RawMessage(tc.value)
					}
					data, err := json.Marshal(m)
					if err != nil {
						t.Fatal(err)
					}
					var c TemplateContent
					if err := json.Unmarshal(data, &c); err != nil {
						t.Errorf("%s %s rejected: %v", tc.name, field, err)
					}
				})
			}
		})
	}
}

func TestTemplateContentUnmarshalMandatoryFieldRejections(t *testing.T) {
	baseMap := func(t *testing.T) map[string]json.RawMessage {
		t.Helper()
		data, err := json.Marshal(mustMinimalTemplateContent(t))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}

	mandatory := []string{
		"generated_artifact_types",
		"expansion_semantics",
		"compatibility_declaration",
		"applicability",
		"provenance",
	}
	for _, field := range mandatory {
		t.Run("absent "+field, func(t *testing.T) {
			m := baseMap(t)
			delete(m, field)
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			var c TemplateContent
			if err := json.Unmarshal(data, &c); err == nil {
				t.Errorf("absent %s accepted, want error", field)
			}
		})
		t.Run("null "+field, func(t *testing.T) {
			m := baseMap(t)
			m[field] = json.RawMessage(`null`)
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			var c TemplateContent
			if err := json.Unmarshal(data, &c); err == nil {
				t.Errorf("null %s accepted, want error", field)
			}
		})
	}
}

func TestTemplateContentUnmarshalOptionalScalarRejections(t *testing.T) {
	baseMap := func(t *testing.T) map[string]json.RawMessage {
		t.Helper()
		data, err := json.Marshal(mustMinimalTemplateContent(t))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}

	t.Run("absent authority is valid", func(t *testing.T) {
		m := baseMap(t)
		delete(m, "authority")
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var c TemplateContent
		if err := json.Unmarshal(data, &c); err != nil {
			t.Errorf("absent authority rejected: %v", err)
		}
		if _, ok := c.Authority(); ok {
			t.Error("absent authority decoded as set")
		}
	})
	t.Run("explicit null authority rejected", func(t *testing.T) {
		m := baseMap(t)
		m["authority"] = json.RawMessage(`null`)
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var c TemplateContent
		if err := json.Unmarshal(data, &c); !errors.Is(err, ErrInvalidTemplateContent) {
			t.Errorf("null authority: error = %v, want %v", err, ErrInvalidTemplateContent)
		}
	})
	t.Run("malformed authority rejected", func(t *testing.T) {
		m := baseMap(t)
		m["authority"] = json.RawMessage(`123`)
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var c TemplateContent
		if err := json.Unmarshal(data, &c); err == nil {
			t.Error("malformed authority accepted, want error")
		}
	})
}

// TestTemplateContentUnmarshalRevalidatesAggregate confirms decode converges on
// the same shared validation path as the constructor: an aggregate invariant
// violated only across collections is still caught.
func TestTemplateContentUnmarshalRevalidatesAggregate(t *testing.T) {
	valid, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", false)},
		[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	t.Run("default naming an undeclared parameter", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		mCopy["defaults"] = json.RawMessage(`[{"parameter":"missing","value":"x"}]`)
		modified, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var c TemplateContent
		if err := json.Unmarshal(modified, &c); !errors.Is(err, ErrUnknownTemplateLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrUnknownTemplateLocalKey)
		}
	})
	t.Run("duplicate parameter keys after decode", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		one, err := json.Marshal(mustParameter(t, "name", false))
		if err != nil {
			t.Fatal(err)
		}
		mCopy["parameters"] = json.RawMessage(`[` + string(one) + `,` + string(one) + `]`)
		modified, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var c TemplateContent
		if err := json.Unmarshal(modified, &c); !errors.Is(err, ErrDuplicateTemplateLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateTemplateLocalKey)
		}
	})
	t.Run("invalid collection element", func(t *testing.T) {
		mCopy := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			mCopy[k] = v
		}
		mCopy["parameters"] = json.RawMessage(`[{"key":"","parameter_type":{"kind":"vocabulary","value":"product:string"},"required":true}]`)
		modified, err := json.Marshal(mCopy)
		if err != nil {
			t.Fatal(err)
		}
		var c TemplateContent
		if err := json.Unmarshal(modified, &c); err == nil {
			t.Error("invalid parameter element accepted, want error")
		}
	})
}

func TestTemplateContentUnmarshalToleratesUnknownFields(t *testing.T) {
	data, err := json.Marshal(mustMinimalTemplateContent(t))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m["product_specific_field"] = json.RawMessage(`{"anything":true}`)
	modified, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var c TemplateContent
	if err := json.Unmarshal(modified, &c); err != nil {
		t.Errorf("unknown field rejected: %v", err)
	}
}

func TestTemplateContentUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	c := mustMinimalTemplateContent(t)
	original, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"generated_artifact_types":[]}`), &c); err == nil {
		t.Fatal("empty generated artifact types accepted, want error")
	}
	if err := json.Unmarshal([]byte(`not json`), &c); err == nil {
		t.Fatal("malformed JSON accepted, want error")
	}
	after, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- TemplateRevision --------------------------------------------------------

func TestNewTemplateRevision(t *testing.T) {
	tmpl := mustTemplate(t, "TPL-1")
	revision := mustArtifactRevision(t, "TPL-1", "REV-1")
	content := mustMinimalTemplateContent(t)

	r, err := NewTemplateRevision(tmpl, revision, content)
	if err != nil {
		t.Fatal(err)
	}
	if r.IsZero() {
		t.Error("valid TemplateRevision reports IsZero() = true")
	}
	if r.Core().ArtifactID() != revision.ArtifactID() || r.Core().RevisionID() != revision.RevisionID() {
		t.Error("Core() mismatch")
	}
	if r.Content().ExpansionSemantics() != content.ExpansionSemantics() {
		t.Error("Content() mismatch")
	}
	ref, err := r.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != mustArtifactID(t, "TPL-1") || ref.RevisionID() != mustArtifactRevisionID(t, "REV-1") {
		t.Error("Ref() mismatch")
	}

	var zero TemplateRevision
	if !zero.IsZero() {
		t.Error("zero-value TemplateRevision.IsZero() = false, want true")
	}
}

func TestNewTemplateRevisionRejections(t *testing.T) {
	tmpl := mustTemplate(t, "TPL-1")
	revision := mustArtifactRevision(t, "TPL-1", "REV-1")
	content := mustMinimalTemplateContent(t)

	if _, err := NewTemplateRevision(Template{}, revision, content); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("zero template: error = %v, want %v", err, ErrInvalidTemplate)
	}
	if _, err := NewTemplateRevision(tmpl, core.ArtifactRevision{}, content); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("zero revision: error = %v, want %v", err, ErrInvalidTemplate)
	}
	if _, err := NewTemplateRevision(tmpl, revision, TemplateContent{}); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("zero content: error = %v, want %v", err, ErrInvalidTemplate)
	}
	mismatched := mustArtifactRevision(t, "TPL-OTHER", "REV-1")
	if _, err := NewTemplateRevision(tmpl, mismatched, content); !errors.Is(err, ErrTemplateArtifactIDMismatch) {
		t.Errorf("artifact id mismatch: error = %v, want %v", err, ErrTemplateArtifactIDMismatch)
	}
}

func TestTemplateRevisionWireForm(t *testing.T) {
	r, err := NewTemplateRevision(mustTemplate(t, "TPL-1"), mustArtifactRevision(t, "TPL-1", "REV-1"), mustMinimalTemplateContent(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if len(m) != 2 {
		t.Errorf("wire form has %d keys, want exactly 2", len(m))
	}
	for _, key := range []string{"core", "content"} {
		if _, ok := m[key]; !ok {
			t.Errorf("wire form missing %q", key)
		}
	}

	var decoded TemplateRevision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
}

func TestTemplateRevisionMarshalZeroAndDecodeRejections(t *testing.T) {
	var zero TemplateRevision
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidTemplate) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidTemplate)
	}

	r, err := NewTemplateRevision(mustTemplate(t, "TPL-1"), mustArtifactRevision(t, "TPL-1", "REV-1"), mustMinimalTemplateContent(t))
	if err != nil {
		t.Fatal(err)
	}
	original, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{`{}`, `{"core":null,"content":null}`, `not json`, `null`} {
		if err := json.Unmarshal([]byte(payload), &r); err == nil {
			t.Errorf("%s accepted, want error", payload)
		}
	}
	after, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

// --- forbidden wire state ----------------------------------------------------

// TestTemplateContentNoForbiddenWireKeys is the structural proof that a
// Template Artifact Revision's content carries a declared generation contract
// only: no template body (that is core.Representation's), no generated output,
// no lifecycle, and no stored compatibility or conformance verdict.
func TestTemplateContentNoForbiddenWireKeys(t *testing.T) {
	c, err := NewTemplateContent(
		[]core.ArtifactType{mustArtifactType(t, "requirement")},
		"expand parameters",
		mustCompatibilityDeclaration(t),
		NewUnrestrictedTemplateApplicability(),
		mustProvenance(t),
		[]Parameter{mustParameter(t, "name", true)},
		[]ParameterDefault{mustParameterDefault(t, "name", "anonymous")},
		[]ParameterConstraint{mustParameterConstraint(t, "c1", "name")},
	)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithAuthority(mustAuthority(t))
	if err != nil {
		t.Fatal(err)
	}

	forbidden := []string{
		"body", "template_body", "source", "script", "expression", "rendered",
		"instance", "current", "active", "effective", "compatible", "conformant",
		"status", "state", "lifecycle", "execution", "invocation", "result",
		"generated_artifact_id", "generated_revision_id", "resolved_values",
		"outcome", "migration", "version", "template_version",
	}
	assertNoWireKeys(t, "TemplateContent", c, forbidden)
}

// assertNoWireKeys marshals v and fails for every forbidden key present at the
// top level of the resulting object. It decodes into map[string]json.RawMessage
// rather than searching the raw bytes, so a forbidden word appearing inside a
// legitimate string value cannot produce a false failure.
func assertNoWireKeys(t *testing.T, label string, v any, forbidden []string) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range forbidden {
		if _, ok := m[key]; ok {
			t.Errorf("%s wire form contains forbidden key %q", label, key)
		}
	}
}
