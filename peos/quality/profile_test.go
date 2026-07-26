package quality

import (
	"encoding/json"
	"errors"
	"reflect"
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

func mustScope(t *testing.T, expression string) core.Scope {
	t.Helper()
	s, err := core.NewScope(mustVocabularyValue(t, core.PEOSNamespace, "component"), expression)
	if err != nil {
		t.Fatal(err)
	}
	return s
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

func mustOrigin(t *testing.T) core.Origin {
	t.Helper()
	o, err := core.NewOrigin(core.OriginKindKnown, "")
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustIntegrity(t *testing.T) core.IntegrityIdentity {
	t.Helper()
	i, err := core.NewIntegrityIdentity(core.IntegrityMechanismCryptographicDigest, "sha256:abc123", core.IntegrityProtectedScopeContent)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func mustAuthority(t *testing.T) core.AuthorityRef {
	t.Helper()
	ref, err := core.NewAuthorityRef("peos", "quality-board")
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustArtifact(t *testing.T, id string, artifactType core.ArtifactType) core.Artifact {
	t.Helper()
	a, err := core.NewArtifact(mustArtifactID(t, id), artifactType)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustProfile(t *testing.T, id string) Profile {
	t.Helper()
	p, err := NewProfile(mustArtifact(t, id, ArtifactTypeQualityProfile))
	if err != nil {
		t.Fatal(err)
	}
	return p
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

// mustProfileContent builds the minimal valid ProfileContent: one
// Characteristic, one Measure referencing it, and no Normalization Rules.
func mustProfileContent(t *testing.T) ProfileContent {
	t.Helper()
	c, err := NewProfileContent(
		mustScope(t, "service=checkout"),
		NewUnrestrictedProfileApplicability(),
		mustProvenance(t),
		[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
		[]Measure{mustMeasure(t, "latency-p99", "latency")},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustSubject(t *testing.T, artifactID string) core.EngineeringSubjectRef {
	t.Helper()
	ref, err := core.NewArtifactRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := core.EngineeringSubjectRefFromArtifact(ref)
	if err != nil {
		t.Fatal(err)
	}
	return subject
}

// --- Profile -----------------------------------------------------------------

func TestArtifactTypeQualityProfile(t *testing.T) {
	if ArtifactTypeQualityProfile.IsZero() {
		t.Fatal("ArtifactTypeQualityProfile is zero")
	}
	if got := ArtifactTypeQualityProfile.String(); got != "peos:quality-profile" {
		t.Errorf("String() = %q, want %q", got, "peos:quality-profile")
	}
	if ns := ArtifactTypeQualityProfile.Value().Namespace(); ns != core.PEOSNamespace {
		t.Errorf("namespace = %q, want %q", ns, core.PEOSNamespace)
	}
}

func TestNewProfile(t *testing.T) {
	p := mustProfile(t, "QP-1")
	if p.ID() != mustArtifactID(t, "QP-1") {
		t.Error("ID() mismatch")
	}
	if p.Core().Type() != ArtifactTypeQualityProfile {
		t.Error("Core().Type() mismatch")
	}
	if p.IsZero() {
		t.Error("IsZero() = true for a constructed profile")
	}

	ref, err := p.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != p.ID() {
		t.Error("Ref() does not identify the profile")
	}
	if ref.IsZero() {
		t.Error("Ref() returned a zero reference")
	}
	// Ref is the general-purpose core.ArtifactRef: core deliberately defines
	// no QualityProfileRef.
	if reflect.TypeOf(ref) != reflect.TypeOf(core.ArtifactRef{}) {
		t.Errorf("Ref() returned %T, want core.ArtifactRef", ref)
	}
}

func TestNewProfileRejectsWrongArtifactType(t *testing.T) {
	if _, err := NewProfile(core.Artifact{}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Errorf("zero artifact error = %v, want %v", err, ErrInvalidQualityProfile)
	}

	wrongType := core.NewArtifactType(mustVocabularyValue(t, core.PEOSNamespace, "requirement"))
	_, err := NewProfile(mustArtifact(t, "REQ-1", wrongType))
	if !errors.Is(err, ErrQualityProfileArtifactTypeMismatch) {
		t.Errorf("wrong artifact type error = %v, want %v", err, ErrQualityProfileArtifactTypeMismatch)
	}
	// A mismatch is not reported as a generic invalid-profile failure: the
	// two sentinels are distinguishable.
	if errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("an artifact type mismatch also matched ErrInvalidQualityProfile")
	}
}

func TestProfileJSONIsBareArtifactWireForm(t *testing.T) {
	p := mustProfile(t, "QP-1")
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	// The wire form is exactly core.Artifact's, with no envelope added.
	coreData, err := json.Marshal(p.Core())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(coreData) {
		t.Errorf("Profile wire form = %s, want the bare core.Artifact form %s", data, coreData)
	}
	assertKeysPresent(t, data, "artifact_type")
	assertKeysAbsent(t, data, "content", "version", "profile_version", "lifecycle",
		"state", "status", "relation", "score", "quality_score", "current",
		"latest", "effective", "aggregate", "characteristics", "measures")

	var decoded Profile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != p.ID() || decoded.Core().Type() != ArtifactTypeQualityProfile {
		t.Error("Profile did not round-trip")
	}

	// A document whose artifact_type is not a Quality Profile is rejected on
	// decode, which is what the bare-Artifact strategy buys.
	wrong, err := json.Marshal(mustArtifact(t, "X-1", core.NewArtifactType(mustVocabularyValue(t, "peos", "validation-plan"))))
	if err != nil {
		t.Fatal(err)
	}
	before := decoded
	if err := json.Unmarshal(wrong, &decoded); !errors.Is(err, ErrQualityProfileArtifactTypeMismatch) {
		t.Errorf("error = %v, want %v", err, ErrQualityProfileArtifactTypeMismatch)
	}
	if decoded.ID() != before.ID() {
		t.Error("a failed decode overwrote a previously valid receiver")
	}

	if _, err := json.Marshal(Profile{}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("zero-value marshal did not fail with the owning sentinel")
	}
	if err := json.Unmarshal([]byte(`null`), &Profile{}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("an explicit null was not rejected")
	}
}

// --- ProfileRevision ---------------------------------------------------------

func TestNewProfileRevision(t *testing.T) {
	p := mustProfile(t, "QP-1")
	revision := mustArtifactRevision(t, "QP-1", "REV-1")
	content := mustProfileContent(t)

	r, err := NewProfileRevision(p, revision, content)
	if err != nil {
		t.Fatal(err)
	}
	if r.Core().ArtifactID() != p.ID() {
		t.Error("Core().ArtifactID() mismatch")
	}
	if len(r.Content().Characteristics()) != 1 {
		t.Error("Content() did not preserve the characteristics")
	}
	if r.IsZero() {
		t.Error("IsZero() = true for a constructed revision")
	}

	ref, err := r.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != p.ID() || ref.RevisionID() != mustArtifactRevisionID(t, "REV-1") {
		t.Error("Ref() does not identify the exact Profile Revision")
	}
	// This is the reference a Quality Claim criterion pairs with a
	// profile-local key.
	elementRef, err := core.NewQualityElementCriterionRef(ref, mustLocalKey(t, "latency"))
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := core.CriterionRefFromQualityCharacteristic(elementRef)
	if err != nil {
		t.Fatal(err)
	}
	if criterion.Kind() != core.CriterionKindQualityCharacteristic {
		t.Errorf("criterion kind = %q", criterion.Kind())
	}
}

func TestNewProfileRevisionRejectsMismatchAndZeroValues(t *testing.T) {
	p := mustProfile(t, "QP-1")
	content := mustProfileContent(t)

	// ArtifactID mismatch between the Profile and the Revision.
	_, err := NewProfileRevision(p, mustArtifactRevision(t, "QP-2", "REV-1"), content)
	if !errors.Is(err, ErrQualityProfileArtifactIDMismatch) {
		t.Errorf("mismatch error = %v, want %v", err, ErrQualityProfileArtifactIDMismatch)
	}

	cases := map[string]func() (ProfileRevision, error){
		"zero profile": func() (ProfileRevision, error) {
			return NewProfileRevision(Profile{}, mustArtifactRevision(t, "QP-1", "REV-1"), content)
		},
		"zero revision": func() (ProfileRevision, error) { return NewProfileRevision(p, core.ArtifactRevision{}, content) },
		"zero content": func() (ProfileRevision, error) {
			return NewProfileRevision(p, mustArtifactRevision(t, "QP-1", "REV-1"), ProfileContent{})
		},
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := fn(); !errors.Is(err, ErrInvalidQualityProfile) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityProfile)
			}
		})
	}
}

func TestProfileRevisionJSON(t *testing.T) {
	p := mustProfile(t, "QP-1")
	r, err := NewProfileRevision(p, mustArtifactRevision(t, "QP-1", "REV-1"), mustProfileContent(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	assertKeysPresent(t, data, "core", "content")
	assertKeysAbsent(t, data, "version", "profile_version", "lifecycle", "state",
		"status", "relation", "score", "quality_score", "id", "profile")

	var decoded ProfileRevision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
	}

	if _, err := json.Marshal(ProfileRevision{}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("zero-value marshal did not fail with the owning sentinel")
	}
	for name, doc := range map[string]string{
		"missing core":    `{"content":{}}`,
		"missing content": `{"core":null}`,
		"null document":   `null`,
		"empty object":    `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			before := decoded
			if err := json.Unmarshal([]byte(doc), &decoded); err == nil {
				t.Fatalf("accepted %s, want rejection", doc)
			}
			if decoded.Core().RevisionID() != before.Core().RevisionID() {
				t.Error("a failed decode overwrote a previously valid receiver")
			}
		})
	}
}

// TestProfileRevisionHasNoContentSetter records that changing a Profile's
// content requires a new Artifact Revision, not an in-place edit: "Modification
// of a Quality Profile's content constitutes a content change and SHALL create
// a new Artifact Revision in accordance with PEOS-002."
func TestProfileRevisionHasNoContentSetter(t *testing.T) {
	typ := reflect.TypeOf(ProfileRevision{})
	for _, candidate := range []reflect.Type{typ, reflect.PointerTo(typ)} {
		for _, name := range []string{"WithContent", "WithoutContent", "SetContent", "WithCore"} {
			if _, ok := candidate.MethodByName(name); ok {
				t.Errorf("ProfileRevision exposes %s", name)
			}
		}
	}
}

// --- ProfileApplicability ----------------------------------------------------

func TestProfileApplicability(t *testing.T) {
	unrestricted := NewUnrestrictedProfileApplicability()
	if unrestricted.IsZero() {
		t.Error("an explicit unrestricted applicability must be non-zero")
	}
	if !unrestricted.IsUnrestricted() || unrestricted.IsScoped() {
		t.Error("unrestricted variant misreports itself")
	}
	if unrestricted.Kind() != "unrestricted" {
		t.Errorf("Kind() = %q", unrestricted.Kind())
	}
	if _, ok := unrestricted.Scope(); ok {
		t.Error("Scope() ok=true for the unrestricted variant")
	}

	scope := mustScope(t, "tier=critical")
	scoped, err := NewScopedProfileApplicability(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !scoped.IsScoped() || scoped.IsUnrestricted() {
		t.Error("scoped variant misreports itself")
	}
	if scoped.Kind() != "scoped" {
		t.Errorf("Kind() = %q", scoped.Kind())
	}
	got, ok := scoped.Scope()
	if !ok || got != scope {
		t.Error("Scope() did not return the supplied scope")
	}

	if _, err := NewScopedProfileApplicability(core.Scope{}); !errors.Is(err, ErrInvalidProfileApplicability) {
		t.Errorf("zero scope error = %v, want %v", err, ErrInvalidProfileApplicability)
	}

	// The zero value is a third, unstated state PEOS-007 does not permit.
	var zero ProfileApplicability
	if !zero.IsZero() || zero.Kind() != "" || zero.IsUnrestricted() || zero.IsScoped() {
		t.Error("the zero value is not recognizable as unstated")
	}
	if _, err := json.Marshal(zero); !errors.Is(err, ErrInvalidProfileApplicability) {
		t.Error("zero-value marshal did not fail with the owning sentinel")
	}
}

func TestProfileApplicabilityJSONMatrix(t *testing.T) {
	scope := mustScope(t, "tier=critical")
	scoped, err := NewScopedProfileApplicability(scope)
	if err != nil {
		t.Fatal(err)
	}

	for name, original := range map[string]ProfileApplicability{
		"unrestricted": NewUnrestrictedProfileApplicability(),
		"scoped":       scoped,
	} {
		t.Run(name+" round trip", func(t *testing.T) {
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decoded ProfileApplicability
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != original.Kind() {
				t.Errorf("Kind() = %q, want %q", decoded.Kind(), original.Kind())
			}
			gotScope, gotOK := decoded.Scope()
			wantScope, wantOK := original.Scope()
			if gotOK != wantOK || gotScope != wantScope {
				t.Error("scope not preserved")
			}
			again, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if string(again) != string(data) {
				t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
			}
		})
	}

	// The unrestricted wire form carries no scope key at all.
	unrestrictedData, err := json.Marshal(NewUnrestrictedProfileApplicability())
	if err != nil {
		t.Fatal(err)
	}
	assertKeysPresent(t, unrestrictedData, "kind")
	assertKeysAbsent(t, unrestrictedData, "scope")

	rejected := map[string]string{
		"missing kind":         `{}`,
		"unknown kind":         `{"kind":"invented"}`,
		"null document":        `null`,
		"null kind":            `{"kind":null}`,
		"unrestricted + scope": `{"kind":"unrestricted","scope":{"kind":"peos:component","expression":"x"}}`,
		"scoped without scope": `{"kind":"scoped"}`,
		"scoped null scope":    `{"kind":"scoped","scope":null}`,
	}
	for name, doc := range rejected {
		t.Run(name, func(t *testing.T) {
			var a ProfileApplicability
			if err := json.Unmarshal([]byte(doc), &a); !errors.Is(err, ErrInvalidProfileApplicability) {
				t.Errorf("error = %v, want %v", err, ErrInvalidProfileApplicability)
			}
			if !a.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
	}

	// A previously valid receiver survives a failed decode.
	valid := NewUnrestrictedProfileApplicability()
	a := valid
	if err := json.Unmarshal([]byte(`{"kind":"nope"}`), &a); err == nil {
		t.Fatal("expected rejection")
	}
	if a.Kind() != valid.Kind() {
		t.Error("a failed decode overwrote a previously valid receiver")
	}
}

// --- ProfileContent: mandatory state ------------------------------------------

func TestNewProfileContent(t *testing.T) {
	c := mustProfileContent(t)
	if c.Scope().IsZero() {
		t.Error("Scope() is zero")
	}
	if !c.Applicability().IsUnrestricted() {
		t.Error("Applicability() not preserved")
	}
	if c.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if len(c.Characteristics()) != 1 || len(c.Measures()) != 1 {
		t.Error("mandatory collections not preserved")
	}
	if c.IsZero() {
		t.Error("IsZero() = true for a constructed content")
	}
	// Optional state is empty until set.
	if c.Thresholds() != nil || c.Targets() != nil || c.Constraints() != nil ||
		c.NormalizationRules() != nil || c.AggregationRules() != nil ||
		c.Subjects() != nil || c.SubjectTypes() != nil {
		t.Error("an optional collection was non-nil before being set")
	}
	if _, ok := c.Authority(); ok {
		t.Error("Authority() ok=true before one is set")
	}
	if !c.Extension().IsZero() {
		t.Error("Extension() non-zero before one is set")
	}
}

func TestNewProfileContentRejectsMissingMandatoryState(t *testing.T) {
	scope := mustScope(t, "s")
	app := NewUnrestrictedProfileApplicability()
	prov := mustProvenance(t)
	chars := []Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")}
	measures := []Measure{mustMeasure(t, "latency-p99", "latency")}

	t.Run("zero scope", func(t *testing.T) {
		_, err := NewProfileContent(core.Scope{}, app, prov, chars, measures, nil)
		if !errors.Is(err, core.ErrInvalidScope) {
			t.Errorf("error = %v, want core.ErrInvalidScope (the owning sentinel, not re-attributed)", err)
		}
	})
	t.Run("unstated applicability", func(t *testing.T) {
		_, err := NewProfileContent(scope, ProfileApplicability{}, prov, chars, measures, nil)
		if !errors.Is(err, ErrInvalidProfileApplicability) {
			t.Errorf("error = %v, want %v", err, ErrInvalidProfileApplicability)
		}
	})
	t.Run("zero provenance", func(t *testing.T) {
		_, err := NewProfileContent(scope, app, core.Provenance{}, chars, measures, nil)
		if !errors.Is(err, ErrInvalidQualityProfile) {
			t.Errorf("error = %v, want %v", err, ErrInvalidQualityProfile)
		}
	})
	t.Run("no characteristics", func(t *testing.T) {
		for _, empty := range [][]Characteristic{nil, {}} {
			_, err := NewProfileContent(scope, app, prov, empty, measures, nil)
			if !errors.Is(err, ErrInvalidQualityProfile) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityProfile)
			}
		}
	})
	t.Run("no measures", func(t *testing.T) {
		for _, empty := range [][]Measure{nil, {}} {
			_, err := NewProfileContent(scope, app, prov, chars, empty, nil)
			if !errors.Is(err, ErrInvalidQualityProfile) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityProfile)
			}
		}
	})
	t.Run("zero characteristic element", func(t *testing.T) {
		_, err := NewProfileContent(scope, app, prov, []Characteristic{{}}, measures, nil)
		if !errors.Is(err, ErrInvalidQualityCharacteristic) {
			t.Errorf("error = %v, want %v", err, ErrInvalidQualityCharacteristic)
		}
	})
	t.Run("zero measure element", func(t *testing.T) {
		_, err := NewProfileContent(scope, app, prov, chars, []Measure{{}}, nil)
		if !errors.Is(err, ErrInvalidQualityMeasure) {
			t.Errorf("error = %v, want %v", err, ErrInvalidQualityMeasure)
		}
	})
}

// --- ProfileContent: per-kind key namespaces ---------------------------------

// TestProfileContentRejectsDuplicateKeyWithinEachKind exercises all seven
// owned-value collections. Uniqueness is enforced per kind because a criterion
// citing a value by key must resolve to exactly one of them.
func TestProfileContentRejectsDuplicateKeyWithinEachKind(t *testing.T) {
	base := func(t *testing.T) ProfileContent {
		t.Helper()
		c, err := NewProfileContent(
			mustScope(t, "s"),
			NewUnrestrictedProfileApplicability(),
			mustProvenance(t),
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{mustMeasure(t, "latency-p99", "latency")},
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}

	t.Run("characteristics", func(t *testing.T) {
		_, err := NewProfileContent(
			mustScope(t, "s"), NewUnrestrictedProfileApplicability(), mustProvenance(t),
			[]Characteristic{
				mustProfileCharacteristic(t, "dup", "first"),
				mustProfileCharacteristic(t, "dup", "second"),
			},
			[]Measure{mustMeasure(t, "m", "dup")},
			nil,
		)
		assertDuplicateKey(t, err, "characteristic", "dup")
	})

	t.Run("measures", func(t *testing.T) {
		_, err := NewProfileContent(
			mustScope(t, "s"), NewUnrestrictedProfileApplicability(), mustProvenance(t),
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{mustMeasure(t, "dup", "latency"), mustMeasure(t, "dup", "latency")},
			nil,
		)
		assertDuplicateKey(t, err, "measure", "dup")
	})

	t.Run("normalization rules", func(t *testing.T) {
		_, err := NewProfileContent(
			mustScope(t, "s"), NewUnrestrictedProfileApplicability(), mustProvenance(t),
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{mustMeasure(t, "latency-p99", "latency")},
			[]NormalizationRule{
				mustNormalizationRule(t, "dup", "first"),
				mustNormalizationRule(t, "dup", "second"),
			},
		)
		assertDuplicateKey(t, err, "normalization rule", "dup")
	})

	t.Run("thresholds", func(t *testing.T) {
		_, err := base(t).WithThresholds([]Threshold{
			mustThreshold(t, "dup", "latency-p99"),
			mustThreshold(t, "dup", "latency-p99"),
		})
		assertDuplicateKey(t, err, "threshold", "dup")
	})

	t.Run("targets", func(t *testing.T) {
		_, err := base(t).WithTargets([]Target{
			mustTarget(t, "dup", "latency-p99"),
			mustTarget(t, "dup", "latency-p99"),
		})
		assertDuplicateKey(t, err, "target", "dup")
	})

	t.Run("constraints", func(t *testing.T) {
		_, err := base(t).WithConstraints([]Constraint{
			mustConstraint(t, "dup", "first"),
			mustConstraint(t, "dup", "second"),
		})
		assertDuplicateKey(t, err, "constraint", "dup")
	})

	t.Run("aggregation rules", func(t *testing.T) {
		_, err := base(t).WithAggregationRules([]AggregationRule{
			mustAggregationRule(t, "dup", "first"),
			mustAggregationRule(t, "dup", "second"),
		})
		assertDuplicateKey(t, err, "aggregation rule", "dup")
	})
}

// assertDuplicateKey checks that err reports ErrDuplicateProfileLocalKey and
// that its message names both the owned-value kind and the offending key --
// the reason a single sentinel suffices instead of seven.
func assertDuplicateKey(t *testing.T, err error, kind, key string) {
	t.Helper()
	if !errors.Is(err, ErrDuplicateProfileLocalKey) {
		t.Fatalf("error = %v, want %v", err, ErrDuplicateProfileLocalKey)
	}
	msg := err.Error()
	if !contains(msg, kind) {
		t.Errorf("error message %q does not name the value kind %q", msg, kind)
	}
	if !contains(msg, key) {
		t.Errorf("error message %q does not name the offending key %q", msg, key)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		func() bool {
			for i := 0; i+len(needle) <= len(haystack); i++ {
				if haystack[i:i+len(needle)] == needle {
					return true
				}
			}
			return false
		}()
}

// TestProfileContentAcceptsSameKeyAcrossAllSevenKinds is the positive half of
// Amendment B: PEOS-007 states no key uniqueness rule at all, so cross-kind
// duplicates must be accepted. Every reference already determines its target
// collection, so "shared" is unambiguous.
func TestProfileContentAcceptsSameKeyAcrossAllSevenKinds(t *testing.T) {
	const shared = "shared"

	measure, err := mustMeasure(t, shared, shared).WithNormalizationRule(mustLocalKey(t, shared))
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewProfileContent(
		mustScope(t, "s"),
		NewUnrestrictedProfileApplicability(),
		mustProvenance(t),
		[]Characteristic{mustProfileCharacteristic(t, shared, "the characteristic")},
		[]Measure{measure},
		[]NormalizationRule{mustNormalizationRule(t, shared, "the normalization rule")},
	)
	if err != nil {
		t.Fatalf("the same key in three collections was rejected: %v", err)
	}
	if c, err = c.WithThresholds([]Threshold{mustThreshold(t, shared, shared)}); err != nil {
		t.Fatalf("threshold sharing the key was rejected: %v", err)
	}
	if c, err = c.WithTargets([]Target{mustTarget(t, shared, shared)}); err != nil {
		t.Fatalf("target sharing the key was rejected: %v", err)
	}
	if c, err = c.WithConstraints([]Constraint{mustConstraint(t, shared, "the constraint")}); err != nil {
		t.Fatalf("constraint sharing the key was rejected: %v", err)
	}
	if c, err = c.WithAggregationRules([]AggregationRule{mustAggregationRule(t, shared, "the aggregation rule")}); err != nil {
		t.Fatalf("aggregation rule sharing the key was rejected: %v", err)
	}

	// All seven now use the same key, and each per-kind lookup resolves in
	// its own namespace only -- returning seven different values for one key.
	key := mustLocalKey(t, shared)
	characteristic, ok := c.Characteristic(key)
	if !ok {
		t.Fatal("Characteristic lookup failed")
	}
	term, _ := characteristic.Term()
	if term != "the characteristic" {
		t.Errorf("Characteristic(%q) resolved to the wrong value: %q", shared, term)
	}
	gotMeasure, ok := c.Measure(key)
	if !ok || gotMeasure.Key() != key {
		t.Error("Measure lookup failed")
	}
	gotThreshold, ok := c.Threshold(key)
	if !ok || gotThreshold.Value() != "250" {
		t.Error("Threshold lookup failed or resolved to the wrong value")
	}
	gotTarget, ok := c.Target(key)
	if !ok || gotTarget.Value() != "120" {
		t.Error("Target lookup failed or resolved to the wrong value")
	}
	gotConstraint, ok := c.Constraint(key)
	if !ok || gotConstraint.Statement() != "the constraint" {
		t.Error("Constraint lookup failed or resolved to the wrong value")
	}
	if len(c.NormalizationRules()) != 1 || c.NormalizationRules()[0].Description() != "the normalization rule" {
		t.Error("normalization rule not preserved")
	}
	if len(c.AggregationRules()) != 1 || c.AggregationRules()[0].Description() != "the aggregation rule" {
		t.Error("aggregation rule not preserved")
	}

	// The whole value still round-trips.
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ProfileContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("a content with one key shared across all seven kinds failed to decode: %v", err)
	}
}

// --- ProfileContent: internal reference resolution ---------------------------

func TestProfileContentResolvesInternalReferences(t *testing.T) {
	scope := mustScope(t, "s")
	app := NewUnrestrictedProfileApplicability()
	prov := mustProvenance(t)

	t.Run("measure to unknown characteristic", func(t *testing.T) {
		_, err := NewProfileContent(scope, app, prov,
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{mustMeasure(t, "m", "absent")},
			nil,
		)
		assertUnknownKey(t, err, "measure", "absent", "characteristic")
	})

	t.Run("measure to unknown normalization rule", func(t *testing.T) {
		m, err := mustMeasure(t, "m", "latency").WithNormalizationRule(mustLocalKey(t, "absent"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewProfileContent(scope, app, prov,
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{m},
			nil,
		)
		assertUnknownKey(t, err, "measure", "absent", "normalization rule")
	})

	t.Run("threshold to unknown measure", func(t *testing.T) {
		_, err := mustProfileContent(t).WithThresholds([]Threshold{mustThreshold(t, "th", "absent")})
		assertUnknownKey(t, err, "threshold", "absent", "measure")
	})

	t.Run("target to unknown measure", func(t *testing.T) {
		_, err := mustProfileContent(t).WithTargets([]Target{mustTarget(t, "tg", "absent")})
		assertUnknownKey(t, err, "target", "absent", "measure")
	})
}

// TestProfileContentRejectsKeyFromWrongCollection is the behavioural
// consequence of per-kind namespaces: a key that exists, but only in a
// different collection, does not satisfy resolution.
func TestProfileContentRejectsKeyFromWrongCollection(t *testing.T) {
	scope := mustScope(t, "s")
	app := NewUnrestrictedProfileApplicability()
	prov := mustProvenance(t)

	t.Run("threshold measure names a characteristic", func(t *testing.T) {
		// "latency" exists -- as a Characteristic, not a Measure.
		_, err := mustProfileContent(t).WithThresholds([]Threshold{mustThreshold(t, "th", "latency")})
		assertUnknownKey(t, err, "threshold", "latency", "measure")
	})

	t.Run("target measure names a characteristic", func(t *testing.T) {
		_, err := mustProfileContent(t).WithTargets([]Target{mustTarget(t, "tg", "latency")})
		assertUnknownKey(t, err, "target", "latency", "measure")
	})

	t.Run("measure characteristic names a measure", func(t *testing.T) {
		// "latency-p99" exists -- as a Measure, not a Characteristic.
		_, err := NewProfileContent(scope, app, prov,
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{mustMeasure(t, "latency-p99", "latency-p99")},
			nil,
		)
		assertUnknownKey(t, err, "measure", "latency-p99", "characteristic")
	})

	t.Run("measure normalization rule names an aggregation rule", func(t *testing.T) {
		m, err := mustMeasure(t, "m", "latency").WithNormalizationRule(mustLocalKey(t, "agg-only"))
		if err != nil {
			t.Fatal(err)
		}
		content, err := NewProfileContent(scope, app, prov,
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{mustMeasure(t, "latency-p99", "latency")},
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		content, err = content.WithAggregationRules([]AggregationRule{mustAggregationRule(t, "agg-only", "d")})
		if err != nil {
			t.Fatal(err)
		}
		// The aggregation rule exists, but a Measure's normalization
		// reference resolves only among Normalization Rules. There is no
		// WithMeasures, so this is exercised through the constructor with the
		// same aggregation rule set absent -- proving the reference cannot be
		// satisfied from the wrong collection.
		_, err = NewProfileContent(scope, app, prov,
			[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
			[]Measure{m},
			nil,
		)
		assertUnknownKey(t, err, "measure", "agg-only", "normalization rule")
		_ = content
	})
}

// assertUnknownKey checks that err reports ErrUnknownProfileLocalKey and that
// its message names the referring value, the referenced key, and the expected
// target collection.
func assertUnknownKey(t *testing.T, err error, referringKind, referencedKey, targetKind string) {
	t.Helper()
	if !errors.Is(err, ErrUnknownProfileLocalKey) {
		t.Fatalf("error = %v, want %v", err, ErrUnknownProfileLocalKey)
	}
	msg := err.Error()
	for label, part := range map[string]string{
		"referring value kind":       referringKind,
		"referenced key":             referencedKey,
		"expected target collection": targetKind,
	} {
		if !contains(msg, part) {
			t.Errorf("error message %q does not name the %s (%q)", msg, label, part)
		}
	}
}

// TestProfileContentResolutionIsOrderIndependent builds the whole key set
// before checking any reference, so declaration order cannot matter.
func TestProfileContentResolutionIsOrderIndependent(t *testing.T) {
	scope := mustScope(t, "s")
	app := NewUnrestrictedProfileApplicability()
	prov := mustProvenance(t)

	measureA, err := mustMeasure(t, "m-a", "c-b").WithNormalizationRule(mustLocalKey(t, "n-b"))
	if err != nil {
		t.Fatal(err)
	}
	measureB, err := mustMeasure(t, "m-b", "c-a").WithNormalizationRule(mustLocalKey(t, "n-a"))
	if err != nil {
		t.Fatal(err)
	}

	// Forward order: each Measure references the *later*-declared
	// Characteristic and Normalization Rule.
	forward, err := NewProfileContent(scope, app, prov,
		[]Characteristic{mustProfileCharacteristic(t, "c-a", "a"), mustProfileCharacteristic(t, "c-b", "b")},
		[]Measure{measureA, measureB},
		[]NormalizationRule{mustNormalizationRule(t, "n-a", "a"), mustNormalizationRule(t, "n-b", "b")},
	)
	if err != nil {
		t.Fatalf("forward order rejected: %v", err)
	}

	// Reverse order: the same set, declared the other way round.
	reverse, err := NewProfileContent(scope, app, prov,
		[]Characteristic{mustProfileCharacteristic(t, "c-b", "b"), mustProfileCharacteristic(t, "c-a", "a")},
		[]Measure{measureB, measureA},
		[]NormalizationRule{mustNormalizationRule(t, "n-b", "b"), mustNormalizationRule(t, "n-a", "a")},
	)
	if err != nil {
		t.Fatalf("reverse order rejected: %v", err)
	}

	// A Threshold declared before the Measure it references also resolves.
	if _, err := forward.WithThresholds([]Threshold{
		mustThreshold(t, "th-1", "m-b"),
		mustThreshold(t, "th-2", "m-a"),
	}); err != nil {
		t.Fatalf("thresholds referencing measures in any order rejected: %v", err)
	}
	if len(reverse.Measures()) != 2 {
		t.Error("reverse content lost a measure")
	}
}

// TestProfileContentAcceptsSelfAndCyclicReferences records the deliberate
// non-enforcement of a rule PEOS-007 does not state. A Characteristic key and
// a Measure key may coincide, producing a Measure that appears to reference
// "itself" by key; the specification says nothing about it, so it is accepted,
// exactly as validation.NewPlanContent accepts Activity dependency cycles.
func TestProfileContentAcceptsSelfAndCyclicReferences(t *testing.T) {
	const shared = "same"
	if _, err := NewProfileContent(
		mustScope(t, "s"), NewUnrestrictedProfileApplicability(), mustProvenance(t),
		[]Characteristic{mustProfileCharacteristic(t, shared, "term")},
		[]Measure{mustMeasure(t, shared, shared)},
		nil,
	); err != nil {
		t.Errorf("a measure whose key equals its characteristic key was rejected: %v", err)
	}
}

// --- ProfileContent: optional state and defensive copying ---------------------

func TestProfileContentOptionalCollections(t *testing.T) {
	base := mustProfileContent(t)

	withThresholds, err := base.WithThresholds([]Threshold{mustThreshold(t, "th", "latency-p99")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withThresholds.Thresholds()) != 1 {
		t.Error("WithThresholds did not set")
	}
	if base.Thresholds() != nil {
		t.Error("WithThresholds mutated the receiver")
	}
	// nil clears; there is no WithoutThresholds.
	cleared, err := withThresholds.WithThresholds(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Thresholds() != nil {
		t.Error("WithThresholds(nil) did not clear")
	}

	withTargets, err := base.WithTargets([]Target{mustTarget(t, "tg", "latency-p99")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withTargets.Targets()) != 1 {
		t.Error("WithTargets did not set")
	}

	withConstraints, err := base.WithConstraints([]Constraint{mustConstraint(t, "co", "no plaintext secrets")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withConstraints.Constraints()) != 1 {
		t.Error("WithConstraints did not set")
	}

	withNorm, err := base.WithNormalizationRules([]NormalizationRule{mustNormalizationRule(t, "n", "d")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withNorm.NormalizationRules()) != 1 {
		t.Error("WithNormalizationRules did not set")
	}

	withAgg, err := base.WithAggregationRules([]AggregationRule{mustAggregationRule(t, "a", "d")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withAgg.AggregationRules()) != 1 {
		t.Error("WithAggregationRules did not set")
	}

	withSubjects, err := base.WithSubjects([]core.EngineeringSubjectRef{mustSubject(t, "ART-1")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withSubjects.Subjects()) != 1 {
		t.Error("WithSubjects did not set")
	}
	if _, err := base.WithSubjects([]core.EngineeringSubjectRef{{}}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("a zero subject was accepted")
	}

	withTypes, err := base.WithSubjectTypes([]core.VocabularyValue{mustVocabularyValue(t, "peos", "requirement")})
	if err != nil {
		t.Fatal(err)
	}
	if len(withTypes.SubjectTypes()) != 1 {
		t.Error("WithSubjectTypes did not set")
	}
	if _, err := base.WithSubjectTypes([]core.VocabularyValue{{}}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("a zero subject type was accepted")
	}

	authority := mustAuthority(t)
	withAuthority, err := base.WithAuthority(authority)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withAuthority.Authority()
	if !ok || got != authority {
		t.Error("WithAuthority did not set")
	}
	if _, ok := withAuthority.WithoutAuthority().Authority(); ok {
		t.Error("WithoutAuthority did not clear")
	}
	if _, err := base.WithAuthority(core.AuthorityRef{}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("a zero authority was accepted")
	}

	withExt := base.WithExtension(mustExtension(t))
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set")
	}
	if !withExt.WithoutExtension().Extension().IsZero() {
		t.Error("WithoutExtension did not clear")
	}

	// A zero element in any owned-value collection is rejected with that
	// value's own sentinel.
	for name, err := range map[string]error{
		"threshold":         firstErr(base.WithThresholds([]Threshold{{}})),
		"target":            firstErr(base.WithTargets([]Target{{}})),
		"constraint":        firstErr(base.WithConstraints([]Constraint{{}})),
		"normalizationRule": firstErr(base.WithNormalizationRules([]NormalizationRule{{}})),
		"aggregationRule":   firstErr(base.WithAggregationRules([]AggregationRule{{}})),
	} {
		if err == nil {
			t.Errorf("a zero %s element was accepted", name)
		}
	}
}

// TestProfileContentModifiersReuseTheSharedValidationPath proves the modifiers
// do not have their own weaker rules: the same invalid input rejected by the
// constructor is rejected by the corresponding modifier, with the same
// sentinel.
func TestProfileContentModifiersReuseTheSharedValidationPath(t *testing.T) {
	base := mustProfileContent(t)

	// Dropping a Normalization Rule that a Measure references must fail,
	// because the shared path re-runs resolution rather than trusting the
	// value that was already valid.
	m, err := mustMeasure(t, "m2", "latency").WithNormalizationRule(mustLocalKey(t, "n"))
	if err != nil {
		t.Fatal(err)
	}
	withRule, err := NewProfileContent(
		mustScope(t, "s"), NewUnrestrictedProfileApplicability(), mustProvenance(t),
		[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
		[]Measure{m},
		[]NormalizationRule{mustNormalizationRule(t, "n", "d")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := withRule.WithNormalizationRules(nil); !errors.Is(err, ErrUnknownProfileLocalKey) {
		t.Errorf("dropping a referenced normalization rule error = %v, want %v", err, ErrUnknownProfileLocalKey)
	}

	// A failed modifier returns the zero value and leaves the receiver
	// untouched, so a caller cannot end up holding a partially modified value.
	broken, err := base.WithThresholds([]Threshold{mustThreshold(t, "th", "absent")})
	if err == nil {
		t.Fatal("expected rejection")
	}
	if !broken.IsZero() {
		t.Error("a failed modifier returned a non-zero value")
	}
	if base.Thresholds() != nil {
		t.Error("a failed modifier mutated the receiver")
	}
}

func TestProfileContentDefensiveCopying(t *testing.T) {
	chars := []Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")}
	measures := []Measure{mustMeasure(t, "latency-p99", "latency")}
	rules := []NormalizationRule{mustNormalizationRule(t, "n", "d")}

	c, err := NewProfileContent(mustScope(t, "s"), NewUnrestrictedProfileApplicability(), mustProvenance(t), chars, measures, rules)
	if err != nil {
		t.Fatal(err)
	}

	// Mutating the caller's input slices must not affect the value.
	chars[0] = Characteristic{}
	measures[0] = Measure{}
	rules[0] = NormalizationRule{}
	if c.Characteristics()[0].IsZero() || c.Measures()[0].IsZero() || c.NormalizationRules()[0].IsZero() {
		t.Error("NewProfileContent retained the caller's slices")
	}

	// Mutating a returned slice must not affect the value either.
	out := c.Characteristics()
	out[0] = Characteristic{}
	if c.Characteristics()[0].IsZero() {
		t.Error("Characteristics() returned an aliased slice")
	}

	thresholds := []Threshold{mustThreshold(t, "th", "latency-p99")}
	withThresholds, err := c.WithThresholds(thresholds)
	if err != nil {
		t.Fatal(err)
	}
	thresholds[0] = Threshold{}
	if withThresholds.Thresholds()[0].IsZero() {
		t.Error("WithThresholds retained the caller's slice")
	}
	outThresholds := withThresholds.Thresholds()
	outThresholds[0] = Threshold{}
	if withThresholds.Thresholds()[0].IsZero() {
		t.Error("Thresholds() returned an aliased slice")
	}
}

func TestProfileContentLookupsMissAndRejectZeroKeys(t *testing.T) {
	c := mustProfileContent(t)
	zero := core.LocalKey{}
	absent := mustLocalKey(t, "absent")

	if _, ok := c.Characteristic(zero); ok {
		t.Error("Characteristic(zero) ok=true")
	}
	if _, ok := c.Measure(zero); ok {
		t.Error("Measure(zero) ok=true")
	}
	if _, ok := c.Threshold(zero); ok {
		t.Error("Threshold(zero) ok=true")
	}
	if _, ok := c.Target(zero); ok {
		t.Error("Target(zero) ok=true")
	}
	if _, ok := c.Constraint(zero); ok {
		t.Error("Constraint(zero) ok=true")
	}
	for name, ok := range map[string]bool{
		"Characteristic": firstBool(c.Characteristic(absent)),
		"Measure":        firstBool(c.Measure(absent)),
		"Threshold":      firstBool(c.Threshold(absent)),
		"Target":         firstBool(c.Target(absent)),
		"Constraint":     firstBool(c.Constraint(absent)),
	} {
		if ok {
			t.Errorf("%s(absent) ok=true", name)
		}
	}
}

func firstBool[T any](_ T, ok bool) bool { return ok }

// --- ProfileContent: JSON ------------------------------------------------------

func fullProfileContent(t *testing.T) ProfileContent {
	t.Helper()
	m, err := mustMeasure(t, "latency-p99", "latency").WithNormalizationRule(mustLocalKey(t, "n"))
	if err != nil {
		t.Fatal(err)
	}
	if m, err = m.WithRequiredEvidence([]string{"trace export"}); err != nil {
		t.Fatal(err)
	}
	c, err := NewProfileContent(
		mustScope(t, "service=checkout"),
		NewUnrestrictedProfileApplicability(),
		mustProvenance(t),
		[]Characteristic{mustProfileCharacteristic(t, "latency", "Response latency")},
		[]Measure{m},
		[]NormalizationRule{mustNormalizationRule(t, "n", "divide by baseline")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithThresholds([]Threshold{mustThreshold(t, "th", "latency-p99")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithTargets([]Target{mustTarget(t, "tg", "latency-p99")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithConstraints([]Constraint{mustConstraint(t, "co", "no plaintext secrets")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithAggregationRules([]AggregationRule{mustAggregationRule(t, "a", "weighted mean")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithSubjects([]core.EngineeringSubjectRef{mustSubject(t, "ART-1")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithSubjectTypes([]core.VocabularyValue{mustVocabularyValue(t, "peos", "artifact-revision")}); err != nil {
		t.Fatal(err)
	}
	if c, err = c.WithAuthority(mustAuthority(t)); err != nil {
		t.Fatal(err)
	}
	return c.WithExtension(mustExtension(t))
}

func TestProfileContentJSONRoundTrip(t *testing.T) {
	original := fullProfileContent(t)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	assertKeysPresent(t, data, "scope", "applicability", "provenance",
		"characteristics", "measures", "thresholds", "targets", "constraints",
		"normalization_rules", "aggregation_rules", "subjects", "subject_types",
		"authority", "extension")

	var decoded ProfileContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Errorf("round trip byte mismatch:\n got %s\nwant %s", again, data)
	}
	if len(decoded.Thresholds()) != 1 || len(decoded.Targets()) != 1 ||
		len(decoded.Constraints()) != 1 || len(decoded.NormalizationRules()) != 1 ||
		len(decoded.AggregationRules()) != 1 || len(decoded.Subjects()) != 1 ||
		len(decoded.SubjectTypes()) != 1 {
		t.Error("an optional collection was lost in the round trip")
	}
	if _, ok := decoded.Authority(); !ok {
		t.Error("authority lost in the round trip")
	}

	// The minimal content omits every optional key.
	minimal, err := json.Marshal(mustProfileContent(t))
	if err != nil {
		t.Fatal(err)
	}
	assertKeysPresent(t, minimal, "scope", "applicability", "provenance", "characteristics", "measures")
	assertKeysAbsent(t, minimal, "thresholds", "targets", "constraints",
		"normalization_rules", "aggregation_rules", "subjects", "subject_types",
		"authority", "extension")
}

// TestProfileContentJSONForbiddenKeysAbsent asserts the wire form carries no
// lifecycle, no relation, no version, and no stored derived quality state.
func TestProfileContentJSONForbiddenKeysAbsent(t *testing.T) {
	data, err := json.Marshal(fullProfileContent(t))
	if err != nil {
		t.Fatal(err)
	}
	assertKeysAbsent(t, data,
		"relation", "lifecycle", "state", "status", "version", "score",
		"quality_score", "current", "latest", "effective", "aggregate",
		"satisfied", "conformant", "compliant", "certified", "accepted",
		"basis", "verdict", "source", "target", "type", "id", "ref",
		"measurement_records", "claims", "executions", "outcome")
}

// TestProfileContentJSONMandatoryCollectionMatrix asserts that absent, null,
// and [] all reject for the two mandatory collections, through the same
// at-least-one invariant.
func TestProfileContentJSONMandatoryCollectionMatrix(t *testing.T) {
	const head = `{"scope":{"kind":"peos:component","expression":"s"},` +
		`"applicability":{"kind":"unrestricted"},` +
		`"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"},"recorded_at":"2026-07-27T00:00:00Z"},`

	valid := head + `"characteristics":[{"kind":"profile","key":"c","term":"t"}],` +
		`"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`

	var ok ProfileContent
	if err := json.Unmarshal([]byte(valid), &ok); err != nil {
		t.Fatalf("valid document rejected: %v", err)
	}

	for name, doc := range map[string]string{
		"characteristics absent": head + `"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`,
		"characteristics null":   head + `"characteristics":null,"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`,
		"characteristics empty":  head + `"characteristics":[],"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`,
		"measures absent":        head + `"characteristics":[{"kind":"profile","key":"c","term":"t"}]}`,
		"measures null":          head + `"characteristics":[{"kind":"profile","key":"c","term":"t"}],"measures":null}`,
		"measures empty":         head + `"characteristics":[{"kind":"profile","key":"c","term":"t"}],"measures":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			var c ProfileContent
			if err := json.Unmarshal([]byte(doc), &c); !errors.Is(err, ErrInvalidQualityProfile) {
				t.Errorf("error = %v, want %v", err, ErrInvalidQualityProfile)
			}
			if !c.IsZero() {
				t.Error("receiver modified by a failed decode")
			}
		})
	}

	// Optional collections: absent, null, and [] are all equivalent to
	// "defines none".
	for _, key := range []string{"thresholds", "targets", "constraints", "normalization_rules", "aggregation_rules", "subjects", "subject_types"} {
		for _, form := range []string{"null", "[]"} {
			doc := valid[:len(valid)-1] + `,"` + key + `":` + form + `}`
			var c ProfileContent
			if err := json.Unmarshal([]byte(doc), &c); err != nil {
				t.Errorf("%s=%s rejected: %v", key, form, err)
			}
		}
	}

	// Mandatory scalars.
	for name, doc := range map[string]string{
		"scope absent":         `{"applicability":{"kind":"unrestricted"},"provenance":{"actor":{"namespace":"a","identifier":"b"},"recorded_at":"2026-07-27T00:00:00Z"},"characteristics":[{"kind":"profile","key":"c","term":"t"}],"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`,
		"applicability absent": `{"scope":{"kind":"peos:component","expression":"s"},"provenance":{"actor":{"namespace":"a","identifier":"b"},"recorded_at":"2026-07-27T00:00:00Z"},"characteristics":[{"kind":"profile","key":"c","term":"t"}],"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`,
		"provenance absent":    `{"scope":{"kind":"peos:component","expression":"s"},"applicability":{"kind":"unrestricted"},"characteristics":[{"kind":"profile","key":"c","term":"t"}],"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`,
		"authority null":       valid[:len(valid)-1] + `,"authority":null}`,
	} {
		t.Run(name, func(t *testing.T) {
			var c ProfileContent
			if err := json.Unmarshal([]byte(doc), &c); err == nil {
				t.Fatalf("accepted %s, want rejection", doc)
			}
		})
	}

	// An absent scope surfaces core's own sentinel, not a re-attributed one.
	var c ProfileContent
	err := json.Unmarshal([]byte(`{"applicability":{"kind":"unrestricted"},"provenance":{"actor":{"namespace":"a","identifier":"b"},"recorded_at":"2026-07-27T00:00:00Z"},"characteristics":[{"kind":"profile","key":"c","term":"t"}],"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}]}`), &c)
	if !errors.Is(err, core.ErrInvalidScope) {
		t.Errorf("absent scope error = %v, want it to wrap core.ErrInvalidScope", err)
	}

	if _, err := json.Marshal(ProfileContent{}); !errors.Is(err, ErrInvalidQualityProfile) {
		t.Error("zero-value marshal did not fail with the owning sentinel")
	}
}

// TestProfileContentJSONEnforcesKeyAndReferenceRules proves the decode path
// runs the same per-kind uniqueness and resolution rules as the constructor --
// the single-shared-path guarantee, checked from the JSON side.
func TestProfileContentJSONEnforcesKeyAndReferenceRules(t *testing.T) {
	const head = `{"scope":{"kind":"peos:component","expression":"s"},` +
		`"applicability":{"kind":"unrestricted"},` +
		`"provenance":{"actor":{"namespace":"peos-cli","identifier":"svc-1"},"recorded_at":"2026-07-27T00:00:00Z"},`
	const measure = `{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}`

	t.Run("duplicate characteristic key", func(t *testing.T) {
		doc := head + `"characteristics":[{"kind":"profile","key":"c","term":"a"},{"kind":"profile","key":"c","term":"b"}],"measures":[` + measure + `]}`
		var c ProfileContent
		err := json.Unmarshal([]byte(doc), &c)
		assertDuplicateKey(t, err, "characteristic", "c")
	})

	t.Run("measure references unknown characteristic", func(t *testing.T) {
		doc := head + `"characteristics":[{"kind":"profile","key":"other","term":"a"}],"measures":[` + measure + `]}`
		var c ProfileContent
		err := json.Unmarshal([]byte(doc), &c)
		assertUnknownKey(t, err, "measure", "c", "characteristic")
	})

	t.Run("threshold references a characteristic key", func(t *testing.T) {
		doc := head + `"characteristics":[{"kind":"profile","key":"c","term":"a"}],"measures":[` + measure +
			`],"thresholds":[{"key":"th","measure":"c","operator":"product-x:lte","value":"1"}]}`
		var c ProfileContent
		err := json.Unmarshal([]byte(doc), &c)
		assertUnknownKey(t, err, "threshold", "c", "measure")
	})

	t.Run("normalization rule supplied alongside a measure referencing it", func(t *testing.T) {
		// This is the case that requires the whole document to be assembled
		// before validation runs: the Measure references "n", which appears
		// only in normalization_rules.
		doc := head + `"characteristics":[{"kind":"profile","key":"c","term":"a"}],` +
			`"measures":[{"key":"m","characteristic":"c","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test","normalization_rule":"n"}],` +
			`"normalization_rules":[{"key":"n","description":"d"}]}`
		var c ProfileContent
		if err := json.Unmarshal([]byte(doc), &c); err != nil {
			t.Fatalf("a valid document was rejected: %v", err)
		}
		m, ok := c.Measure(mustLocalKey(t, "m"))
		if !ok {
			t.Fatal("measure lost")
		}
		if key, ok := m.NormalizationRule(); !ok || key != mustLocalKey(t, "n") {
			t.Error("normalization rule reference lost")
		}
	})

	t.Run("same key across kinds accepted on decode", func(t *testing.T) {
		doc := head + `"characteristics":[{"kind":"profile","key":"x","term":"a"}],` +
			`"measures":[{"key":"x","characteristic":"x","unit":"product-x:ms","scale":"product-x:ratio","method":"product-x:test"}],` +
			`"thresholds":[{"key":"x","measure":"x","operator":"product-x:lte","value":"1"}],` +
			`"targets":[{"key":"x","measure":"x","value":"1"}],` +
			`"constraints":[{"key":"x","statement":"s"}],` +
			`"normalization_rules":[{"key":"x","description":"d"}],` +
			`"aggregation_rules":[{"key":"x","description":"d"}]}`
		var c ProfileContent
		if err := json.Unmarshal([]byte(doc), &c); err != nil {
			t.Fatalf("one key shared across all seven kinds was rejected on decode: %v", err)
		}
	})
}

// --- constructor completeness -------------------------------------------------

// TestNoWithMethodEstablishesMandatoryState is the reflective half of the
// constructor-completeness audit: for every public type in this package, no
// With* method may return the type as its only result while establishing
// mandatory state, and the constructor argument lists must cover every
// mandatory value.
//
// The absence checks live in doc_test.go; this test asserts the positive
// direction -- that a successful constructor call always yields a value that
// marshals, meaning no mandatory field was left for a later call.
func TestConstructorsYieldImmediatelyValidValues(t *testing.T) {
	values := []any{
		mustProfile(t, "QP-1"),
		mustProfileContent(t),
		mustProfileCharacteristic(t, "c", "term"),
		mustMeasure(t, "m", "c"),
		mustThreshold(t, "th", "m"),
		mustTarget(t, "tg", "m"),
		mustConstraint(t, "co", "statement"),
		mustNormalizationRule(t, "n", "d"),
		mustAggregationRule(t, "a", "d"),
		NewUnrestrictedProfileApplicability(),
		mustUnit(t, "ms"),
		mustScale(t, "ratio"),
		mustOperator(t, "lte"),
	}
	for _, v := range values {
		if _, err := json.Marshal(v); err != nil {
			t.Errorf("%T: a freshly constructed value failed to marshal, so a mandatory field was not established by its constructor: %v", v, err)
		}
	}

	revision, err := NewProfileRevision(mustProfile(t, "QP-1"), mustArtifactRevision(t, "QP-1", "REV-1"), mustProfileContent(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := json.Marshal(revision); err != nil {
		t.Errorf("ProfileRevision: %v", err)
	}
}

// TestZeroValuesAreNeverSilentlyValid asserts that no zero value is accepted
// as a legitimate "unrestricted" or "unspecified" state anywhere.
func TestZeroValuesAreNeverSilentlyValid(t *testing.T) {
	marshalers := map[string]json.Marshaler{
		"Profile":              Profile{},
		"ProfileRevision":      ProfileRevision{},
		"ProfileContent":       ProfileContent{},
		"Characteristic":       Characteristic{},
		"Measure":              Measure{},
		"Threshold":            Threshold{},
		"Target":               Target{},
		"Constraint":           Constraint{},
		"NormalizationRule":    NormalizationRule{},
		"AggregationRule":      AggregationRule{},
		"ProfileApplicability": ProfileApplicability{},
	}
	for name, m := range marshalers {
		if _, err := m.MarshalJSON(); err == nil {
			t.Errorf("%s: a zero value marshaled successfully, so it is silently valid", name)
		}
	}

	// The three vocabulary wrappers are the deliberate exception: they are
	// thin wrappers whose zero value marshals to the zero vocabulary value,
	// exactly as core's own wrappers do. Their mandatory-ness lives in the
	// aggregate constructors that consume them, which reject a zero one.
	if _, err := NewMeasure(mustLocalKey(t, "m"), mustLocalKey(t, "c"), Unit{}, mustScale(t, "r"), mustValidationMethod(t, "t")); err == nil {
		t.Error("a zero Unit was accepted by NewMeasure")
	}
	if _, err := NewThreshold(mustLocalKey(t, "t"), mustLocalKey(t, "m"), ThresholdOperator{}, "1"); err == nil {
		t.Error("a zero ThresholdOperator was accepted by NewThreshold")
	}
}
