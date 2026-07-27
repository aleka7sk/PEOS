package runtime

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

func mustContractRule(t *testing.T, key, text string) ContractRule {
	t.Helper()
	r, err := NewContractRule(mustLocalKey(t, key), text)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func mustVocabularyValue(t *testing.T, namespace, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
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
	ref, err := core.NewAuthorityRef("peos", "runtime-board")
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

func mustContract(t *testing.T, id string) Contract {
	t.Helper()
	c, err := NewContract(mustArtifact(t, id, ArtifactTypeRuntimeContract))
	if err != nil {
		t.Fatal(err)
	}
	return c
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

func mustRuntimeSubjectRef(t *testing.T, namespace, identifier string) core.RuntimeSubjectRef {
	t.Helper()
	ref, err := core.NewRuntimeSubjectRef(namespace, identifier)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustEnvironment(t *testing.T, value string) Environment {
	t.Helper()
	return NewEnvironment(mustVocabularyValue(t, "product", value))
}

func mustRequirementRef(t *testing.T, id string) core.RequirementRef {
	t.Helper()
	ref, err := core.NewRequirementRef(mustArtifactID(t, id))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustRequirementRevisionRef(t *testing.T, artifactID, revisionID string) core.RequirementArtifactRevisionRef {
	t.Helper()
	ref, err := core.NewRequirementArtifactRevisionRef(mustArtifactID(t, artifactID), mustArtifactRevisionID(t, revisionID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustRequirementIdentityReference(t *testing.T, id string) RequirementReference {
	t.Helper()
	ref, err := NewRequirementIdentityReference(mustRequirementRef(t, id))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// mustMinimalContractContent builds the minimal valid ContractContent: one
// identity-level Requirement reference, mandatory scalars, and no
// Assertions or optional collections.
func mustMinimalContractContent(t *testing.T) ContractContent {
	t.Helper()
	c, err := NewContractContent(
		[]RequirementReference{mustRequirementIdentityReference(t, "REQ-1")},
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustEnvironment(t, "production"),
		mustScope(t, "cluster=prod-1"),
		NewUnrestrictedContractApplicability(),
		mustProvenance(t),
		mustAuthority(t),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// --- ArtifactTypeRuntimeContract ---------------------------------------------

func TestArtifactTypeRuntimeContract(t *testing.T) {
	if ArtifactTypeRuntimeContract.IsZero() {
		t.Fatal("ArtifactTypeRuntimeContract is zero")
	}
	if got := ArtifactTypeRuntimeContract.String(); got != "peos:runtime-contract" {
		t.Errorf("String() = %q, want %q", got, "peos:runtime-contract")
	}
	if ns := ArtifactTypeRuntimeContract.Value().Namespace(); ns != core.PEOSNamespace {
		t.Errorf("namespace = %q, want %q", ns, core.PEOSNamespace)
	}
}

// --- Contract ------------------------------------------------------------------

func TestNewContract(t *testing.T) {
	if _, err := NewContract(core.Artifact{}); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero artifact: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}

	wrongType := mustArtifact(t, "ART-1", mustVocabularyValueArtifactType(t))
	if _, err := NewContract(wrongType); !errors.Is(err, ErrRuntimeContractArtifactTypeMismatch) {
		t.Errorf("wrong artifact type: error = %v, want %v", err, ErrRuntimeContractArtifactTypeMismatch)
	}

	c := mustContract(t, "ART-1")
	if c.IsZero() {
		t.Error("valid Contract reports IsZero() = true")
	}
	if c.ID() != mustArtifactID(t, "ART-1") {
		t.Error("ID() mismatch")
	}
	ref, err := c.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != c.ID() {
		t.Error("Ref() ArtifactID mismatch")
	}
}

func mustVocabularyValueArtifactType(t *testing.T) core.ArtifactType {
	t.Helper()
	return core.NewArtifactType(mustVocabularyValue(t, core.PEOSNamespace, "something-else"))
}

func TestContractJSONRoundTrip(t *testing.T) {
	c := mustContract(t, "ART-1")
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Contract
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != c.ID() {
		t.Errorf("round trip mismatch: got %v, want %v", decoded.ID(), c.ID())
	}
}

func TestContractMarshalZero(t *testing.T) {
	var c Contract
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	c := mustContract(t, "ART-1")
	originalData, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"id":"ART-2","artifact_type":"peos:something-else"}`), &c); err == nil {
		t.Fatal("wrong artifact type accepted, want error")
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
	if err := json.Unmarshal([]byte(`null`), &c); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("null: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
	if err := json.Unmarshal([]byte(`not json`), &c); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
}

// --- ContractApplicability ---------------------------------------------------

func TestContractApplicabilityUnrestricted(t *testing.T) {
	a := NewUnrestrictedContractApplicability()
	if a.IsZero() {
		t.Error("unrestricted applicability reports IsZero() = true")
	}
	if !a.IsUnrestricted() || a.IsScoped() {
		t.Error("IsUnrestricted/IsScoped mismatch")
	}
	if _, ok := a.Scope(); ok {
		t.Error("Scope() ok = true for unrestricted")
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"kind":"unrestricted"}` {
		t.Errorf("Marshal = %s", data)
	}
	var decoded ContractApplicability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != a {
		t.Error("round trip mismatch")
	}
}

func TestContractApplicabilityScoped(t *testing.T) {
	scope := mustScope(t, "cluster=prod-1")
	a, err := NewScopedContractApplicability(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !a.IsScoped() || a.IsUnrestricted() {
		t.Error("IsScoped/IsUnrestricted mismatch")
	}
	got, ok := a.Scope()
	if !ok || got != scope {
		t.Errorf("Scope() = (%v, %v), want (%v, true)", got, ok, scope)
	}

	if _, err := NewScopedContractApplicability(core.Scope{}); !errors.Is(err, ErrInvalidContractApplicability) {
		t.Errorf("zero scope: error = %v, want %v", err, ErrInvalidContractApplicability)
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContractApplicability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != a {
		t.Error("round trip mismatch")
	}
}

func TestContractApplicabilityZeroInvalid(t *testing.T) {
	var a ContractApplicability
	if !a.IsZero() {
		t.Error("zero-value ContractApplicability.IsZero() = false, want true")
	}
	if _, err := json.Marshal(a); !errors.Is(err, ErrInvalidContractApplicability) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidContractApplicability)
	}
}

func TestContractApplicabilityUnmarshalRejections(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing kind", `{}`},
		{"unknown kind", `{"kind":"bogus"}`},
		{"null", `null`},
		{"unrestricted with scope", `{"kind":"unrestricted","scope":{"kind":"peos:component","expression":"x"}}`},
		{"scoped without scope", `{"kind":"scoped"}`},
		{"scoped with null scope", `{"kind":"scoped","scope":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a ContractApplicability
			if err := json.Unmarshal([]byte(tt.json), &a); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

// --- RequirementReference ----------------------------------------------------

func TestRequirementIdentityReference(t *testing.T) {
	ref := mustRequirementRef(t, "REQ-1")
	r, err := NewRequirementIdentityReference(ref)
	if err != nil {
		t.Fatal(err)
	}
	if r.Kind() != "identity" {
		t.Errorf("Kind() = %q, want %q", r.Kind(), "identity")
	}
	got, ok := r.Identity()
	if !ok || got != ref {
		t.Errorf("Identity() = (%v, %v), want (%v, true)", got, ok, ref)
	}
	if _, ok := r.Revision(); ok {
		t.Error("Revision() ok = true for identity arm")
	}
	if r.RequirementArtifactID() != ref.ArtifactID() {
		t.Error("RequirementArtifactID() mismatch")
	}

	if _, err := NewRequirementIdentityReference(core.RequirementRef{}); !errors.Is(err, ErrInvalidRequirementReference) {
		t.Errorf("zero ref: error = %v, want %v", err, ErrInvalidRequirementReference)
	}
}

func TestRequirementRevisionReference(t *testing.T) {
	ref := mustRequirementRevisionRef(t, "REQ-1", "REV-1")
	r, err := NewRequirementRevisionReference(ref)
	if err != nil {
		t.Fatal(err)
	}
	if r.Kind() != "revision" {
		t.Errorf("Kind() = %q, want %q", r.Kind(), "revision")
	}
	got, ok := r.Revision()
	if !ok || got != ref {
		t.Errorf("Revision() = (%v, %v), want (%v, true)", got, ok, ref)
	}
	if _, ok := r.Identity(); ok {
		t.Error("Identity() ok = true for revision arm")
	}
	if r.RequirementArtifactID() != ref.ArtifactID() {
		t.Error("RequirementArtifactID() mismatch")
	}

	if _, err := NewRequirementRevisionReference(core.RequirementArtifactRevisionRef{}); !errors.Is(err, ErrInvalidRequirementReference) {
		t.Errorf("zero ref: error = %v, want %v", err, ErrInvalidRequirementReference)
	}
}

func TestRequirementReferenceJSONRoundTrip(t *testing.T) {
	identity, err := NewRequirementIdentityReference(mustRequirementRef(t, "REQ-1"))
	if err != nil {
		t.Fatal(err)
	}
	revision, err := NewRequirementRevisionReference(mustRequirementRevisionRef(t, "REQ-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}

	for name, r := range map[string]RequirementReference{"identity": identity, "revision": revision} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			var decoded RequirementReference
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != r {
				t.Errorf("round trip mismatch: got %+v, want %+v", decoded, r)
			}
		})
	}
}

func TestRequirementReferenceUnmarshalRejections(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing kind", `{}`},
		{"unknown kind", `{"kind":"bogus"}`},
		{"null", `null`},
		{"identity missing identity", `{"kind":"identity"}`},
		{"identity with null identity", `{"kind":"identity","identity":null}`},
		{"identity carrying revision", `{"kind":"identity","identity":{"artifact_id":"REQ-1"},"revision":{"artifact_id":"REQ-1","revision_id":"REV-1"}}`},
		{"revision missing revision", `{"kind":"revision"}`},
		{"revision with null revision", `{"kind":"revision","revision":null}`},
		{"revision carrying identity", `{"kind":"revision","revision":{"artifact_id":"REQ-1","revision_id":"REV-1"},"identity":{"artifact_id":"REQ-1"}}`},
		{"malformed JSON", `not json`},
		{"malformed identity payload", `{"kind":"identity","identity":123}`},
		{"malformed revision payload", `{"kind":"revision","revision":123}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r RequirementReference
			if err := json.Unmarshal([]byte(tt.json), &r); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

func TestRequirementReferenceZeroInvalid(t *testing.T) {
	var r RequirementReference
	if !r.IsZero() {
		t.Error("zero-value RequirementReference.IsZero() = false, want true")
	}
	if r.RequirementArtifactID() != (core.ArtifactID{}) {
		t.Error("RequirementArtifactID() on zero value is not zero")
	}
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRequirementReference) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRequirementReference)
	}
}

// --- ContractContent: cardinality matrix -------------------------------------

func TestContractContentRequiresAtLeastOneRequirement(t *testing.T) {
	_, err := NewContractContent(
		nil,
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustEnvironment(t, "production"),
		mustScope(t, "cluster=prod-1"),
		NewUnrestrictedContractApplicability(),
		mustProvenance(t),
		mustAuthority(t),
		nil,
	)
	if !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("empty requirements: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractContentCardinalityMatrix(t *testing.T) {
	identity := mustRequirementIdentityReference(t, "REQ-1")
	revision, err := NewRequirementRevisionReference(mustRequirementRevisionRef(t, "REQ-2", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	assertion := mustAssertion(t, "assert-1")

	base := func(requirements []RequirementReference, assertions []Assertion) (ContractContent, error) {
		return NewContractContent(
			requirements,
			mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
			mustEnvironment(t, "production"),
			mustScope(t, "cluster=prod-1"),
			NewUnrestrictedContractApplicability(),
			mustProvenance(t),
			mustAuthority(t),
			assertions,
		)
	}

	// 1. zero Requirement references -- reject.
	if _, err := base(nil, nil); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("case 1 (empty requirements): error = %v, want %v", err, ErrInvalidRuntimeContract)
	}

	// 2. one Requirement identity reference, zero Assertions -- accept.
	if _, err := base([]RequirementReference{identity}, nil); err != nil {
		t.Errorf("case 2 (identity only): unexpected error %v", err)
	}

	// 3. one Requirement Revision reference, zero Assertions -- accept.
	if _, err := base([]RequirementReference{revision}, nil); err != nil {
		t.Errorf("case 3 (revision only): unexpected error %v", err)
	}

	// 4. Requirement reference plus one Assertion -- accept.
	if _, err := base([]RequirementReference{identity}, []Assertion{assertion}); err != nil {
		t.Errorf("case 4 (requirement + assertion): unexpected error %v", err)
	}

	// 10. duplicate Assertion key -- reject.
	if _, err := base([]RequirementReference{identity}, []Assertion{assertion, assertion}); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
		t.Errorf("case 10 (duplicate assertion key): error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
	}

	// Requirement reference plus only observation requirements -- accept.
	c, err := base([]RequirementReference{identity}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.WithObservationRequirements([]ContractRule{mustContractRule(t, "observe-cpu", "observe cpu")}); err != nil {
		t.Errorf("observation-requirements-only: unexpected error %v", err)
	}

	// Requirement reference plus only classification rules -- accept.
	if _, err := c.WithViolationClassificationRules([]ContractRule{mustContractRule(t, "latency-breach", "latency breach")}); err != nil {
		t.Errorf("classification-rules-only: unexpected error %v", err)
	}

	// Requirement reference plus only waiver handling rules -- accept.
	if _, err := c.WithWaiverHandlingRules([]ContractRule{mustContractRule(t, "defer-oncall", "defer to on-call")}); err != nil {
		t.Errorf("waiver-rules-only: unexpected error %v", err)
	}

	// Requirement reference plus only enforcement expectations -- accept.
	if _, err := c.WithEnforcementExpectations([]ContractRule{mustContractRule(t, "page-oncall", "page on-call")}); err != nil {
		t.Errorf("enforcement-expectations-only: unexpected error %v", err)
	}

	// Requirement reference plus only Quality Profile Revisions -- accept.
	qpRef, err := core.NewArtifactRevisionRef(mustArtifactID(t, "PROFILE-1"), mustArtifactRevisionID(t, "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.WithQualityProfileRevisions([]core.ArtifactRevisionRef{qpRef}); err != nil {
		t.Errorf("quality-profile-revisions-only: unexpected error %v", err)
	}
}

func TestContractContentAssertionCrossKindKeyReuse(t *testing.T) {
	// Assertions are the only keyed collection Packet J.1 implements, so
	// cross-kind reuse cannot yet be exercised against a second keyed
	// collection; this test documents that a second Assertion with a
	// distinct key is accepted (uniqueness is per-collection, not global
	// beyond this one collection).
	a1 := mustAssertion(t, "assert-1")
	a2 := mustAssertion(t, "assert-2")
	_, err := NewContractContent(
		[]RequirementReference{mustRequirementIdentityReference(t, "REQ-1")},
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustEnvironment(t, "production"),
		mustScope(t, "cluster=prod-1"),
		NewUnrestrictedContractApplicability(),
		mustProvenance(t),
		mustAuthority(t),
		[]Assertion{a1, a2},
	)
	if err != nil {
		t.Errorf("two distinct assertion keys: unexpected error %v", err)
	}
}

func TestContractContentMandatoryFieldRejections(t *testing.T) {
	valid := func() (
		[]RequirementReference, core.RuntimeSubjectRef, Environment, core.Scope,
		ContractApplicability, core.Provenance, core.AuthorityRef, []Assertion,
	) {
		return []RequirementReference{mustRequirementIdentityReference(t, "REQ-1")},
			mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
			mustEnvironment(t, "production"),
			mustScope(t, "cluster=prod-1"),
			NewUnrestrictedContractApplicability(),
			mustProvenance(t),
			mustAuthority(t),
			nil
	}

	t.Run("zero subject target", func(t *testing.T) {
		reqs, _, env, scope, app, prov, auth, assertions := valid()
		if _, err := NewContractContent(reqs, core.RuntimeSubjectRef{}, env, scope, app, prov, auth, assertions); !errors.Is(err, ErrInvalidRuntimeContract) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContract)
		}
	})
	t.Run("zero environment", func(t *testing.T) {
		reqs, subj, _, scope, app, prov, auth, assertions := valid()
		if _, err := NewContractContent(reqs, subj, Environment{}, scope, app, prov, auth, assertions); !errors.Is(err, ErrInvalidRuntimeContract) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContract)
		}
	})
	t.Run("zero deployment scope", func(t *testing.T) {
		reqs, subj, env, _, app, prov, auth, assertions := valid()
		if _, err := NewContractContent(reqs, subj, env, core.Scope{}, app, prov, auth, assertions); !errors.Is(err, core.ErrInvalidScope) {
			t.Errorf("error = %v, want %v", err, core.ErrInvalidScope)
		}
	})
	t.Run("zero applicability", func(t *testing.T) {
		reqs, subj, env, scope, _, prov, auth, assertions := valid()
		if _, err := NewContractContent(reqs, subj, env, scope, ContractApplicability{}, prov, auth, assertions); !errors.Is(err, ErrInvalidContractApplicability) {
			t.Errorf("error = %v, want %v", err, ErrInvalidContractApplicability)
		}
	})
	t.Run("zero provenance", func(t *testing.T) {
		reqs, subj, env, scope, app, _, auth, assertions := valid()
		if _, err := NewContractContent(reqs, subj, env, scope, app, core.Provenance{}, auth, assertions); !errors.Is(err, ErrInvalidRuntimeContract) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContract)
		}
	})
	t.Run("zero authority", func(t *testing.T) {
		reqs, subj, env, scope, app, prov, _, assertions := valid()
		if _, err := NewContractContent(reqs, subj, env, scope, app, prov, core.AuthorityRef{}, assertions); !errors.Is(err, ErrInvalidRuntimeContract) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContract)
		}
	})
}

func TestContractContentDefensiveCopy(t *testing.T) {
	reqs := []RequirementReference{mustRequirementIdentityReference(t, "REQ-1")}
	c := mustMinimalContractContent(t)
	c2, err := c.WithObservationRequirements([]ContractRule{mustContractRule(t, "a", "a"), mustContractRule(t, "b", "b")})
	if err != nil {
		t.Fatal(err)
	}
	returned := c2.ObservationRequirements()
	returned[0] = mustContractRule(t, "mutated", "mutated")
	if c2.ObservationRequirements()[0].Key() == mustLocalKey(t, "mutated") {
		t.Error("ObservationRequirements() accessor did not return a defensive copy")
	}

	reqs[0] = RequirementReference{}
	if c.Requirements()[0].IsZero() {
		t.Error("constructor did not defensively copy requirements input")
	}
}

func TestContractContentAccessors(t *testing.T) {
	subjectTarget := mustRuntimeSubjectRef(t, "kubernetes", "pod-1")
	environment := mustEnvironment(t, "production")
	deploymentScope := mustScope(t, "cluster=prod-1")
	applicability := NewUnrestrictedContractApplicability()
	provenance := mustProvenance(t)
	authority := mustAuthority(t)
	requirement := mustRequirementIdentityReference(t, "REQ-1")
	assertion := mustAssertion(t, "assert-1")

	c, err := NewContractContent(
		[]RequirementReference{requirement},
		subjectTarget, environment, deploymentScope, applicability, provenance, authority,
		[]Assertion{assertion},
	)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithObservationRequirements([]ContractRule{mustContractRule(t, "observe-cpu", "observe cpu")})
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithViolationClassificationRules([]ContractRule{mustContractRule(t, "latency-breach", "latency breach")})
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithWaiverHandlingRules([]ContractRule{mustContractRule(t, "defer-oncall", "defer to on-call")})
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithEnforcementExpectations([]ContractRule{mustContractRule(t, "page-oncall", "page on-call")})
	if err != nil {
		t.Fatal(err)
	}
	qpRef, err := core.NewArtifactRevisionRef(mustArtifactID(t, "PROFILE-1"), mustArtifactRevisionID(t, "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithQualityProfileRevisions([]core.ArtifactRevisionRef{qpRef})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	c = c.WithExtension(ext)

	if c.SubjectTarget() != subjectTarget {
		t.Error("SubjectTarget() mismatch")
	}
	if c.Environment() != environment {
		t.Error("Environment() mismatch")
	}
	if c.DeploymentScope() != deploymentScope {
		t.Error("DeploymentScope() mismatch")
	}
	if c.Applicability() != applicability {
		t.Error("Applicability() mismatch")
	}
	if c.Provenance().IsZero() {
		t.Error("Provenance() is zero")
	}
	if c.Authority() != authority {
		t.Error("Authority() mismatch")
	}
	if len(c.Assertions()) != 1 || c.Assertions()[0].Key() != assertion.Key() {
		t.Error("Assertions() mismatch")
	}
	got, ok := c.Assertion(assertion.Key())
	if !ok || got.Key() != assertion.Key() {
		t.Error("Assertion(key) lookup failed")
	}
	if _, ok := c.Assertion(core.LocalKey{}); ok {
		t.Error("Assertion(zero key) should return ok=false")
	}
	if _, ok := c.Assertion(mustLocalKey(t, "does-not-exist")); ok {
		t.Error("Assertion(unknown key) should return ok=false")
	}
	if got := c.ObservationRequirements(); len(got) != 1 || got[0].Text() != "observe cpu" {
		t.Errorf("ObservationRequirements() = %v", got)
	}
	if got := c.ViolationClassificationRules(); len(got) != 1 || got[0].Text() != "latency breach" {
		t.Errorf("ViolationClassificationRules() = %v", got)
	}
	if got := c.WaiverHandlingRules(); len(got) != 1 || got[0].Text() != "defer to on-call" {
		t.Errorf("WaiverHandlingRules() = %v", got)
	}
	if got := c.EnforcementExpectations(); len(got) != 1 || got[0].Text() != "page on-call" {
		t.Errorf("EnforcementExpectations() = %v", got)
	}
	if got := c.QualityProfileRevisions(); len(got) != 1 || got[0] != qpRef {
		t.Errorf("QualityProfileRevisions() = %v", got)
	}
	if c.Extension().IsZero() {
		t.Error("Extension() is zero")
	}
	got2, ok := c.ContractRule(mustLocalKey(t, "latency-breach"))
	if !ok || got2.Text() != "latency breach" {
		t.Errorf("ContractRule(key) lookup failed: got (%v, %v)", got2, ok)
	}
	if _, ok := c.ContractRule(core.LocalKey{}); ok {
		t.Error("ContractRule(zero key) should return ok=false")
	}
	if _, ok := c.ContractRule(mustLocalKey(t, "does-not-exist")); ok {
		t.Error("ContractRule(unknown key) should return ok=false")
	}
}

func TestContractContentWithoutExtension(t *testing.T) {
	c := mustMinimalContractContent(t)
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	c = c.WithExtension(ext)
	if c.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
	c = c.WithoutExtension()
	if !c.Extension().IsZero() {
		t.Error("WithoutExtension did not clear extension")
	}
}

func TestContractContentZeroContractRuleRejections(t *testing.T) {
	c := mustMinimalContractContent(t)
	if _, err := c.WithObservationRequirements([]ContractRule{{}}); !errors.Is(err, ErrInvalidRuntimeContractRule) {
		t.Errorf("zero observation requirement: error = %v, want %v", err, ErrInvalidRuntimeContractRule)
	}
	if _, err := c.WithViolationClassificationRules([]ContractRule{{}}); !errors.Is(err, ErrInvalidRuntimeContractRule) {
		t.Errorf("zero classification rule: error = %v, want %v", err, ErrInvalidRuntimeContractRule)
	}
	if _, err := c.WithWaiverHandlingRules([]ContractRule{{}}); !errors.Is(err, ErrInvalidRuntimeContractRule) {
		t.Errorf("zero waiver rule: error = %v, want %v", err, ErrInvalidRuntimeContractRule)
	}
	if _, err := c.WithEnforcementExpectations([]ContractRule{{}}); !errors.Is(err, ErrInvalidRuntimeContractRule) {
		t.Errorf("zero enforcement expectation: error = %v, want %v", err, ErrInvalidRuntimeContractRule)
	}
	if _, err := c.WithQualityProfileRevisions([]core.ArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero quality profile revision: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

// --- ContractRule ------------------------------------------------------------

func TestNewContractRule(t *testing.T) {
	r, err := NewContractRule(mustLocalKey(t, "observe-cpu"), "  observe cpu utilization  ")
	if err != nil {
		t.Fatal(err)
	}
	if r.Key() != mustLocalKey(t, "observe-cpu") {
		t.Errorf("Key() = %v, want observe-cpu", r.Key())
	}
	if r.Text() != "observe cpu utilization" {
		t.Errorf("Text() = %q, want trimmed text", r.Text())
	}
	if r.IsZero() {
		t.Error("valid ContractRule reports IsZero() = true")
	}
}

func TestNewContractRuleMandatoryFieldRejections(t *testing.T) {
	t.Run("zero key", func(t *testing.T) {
		if _, err := NewContractRule(core.LocalKey{}, "text"); !errors.Is(err, ErrInvalidRuntimeContractRule) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContractRule)
		}
	})
	t.Run("empty text", func(t *testing.T) {
		if _, err := NewContractRule(mustLocalKey(t, "k"), ""); !errors.Is(err, ErrInvalidRuntimeContractRule) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContractRule)
		}
	})
	t.Run("whitespace-only text", func(t *testing.T) {
		if _, err := NewContractRule(mustLocalKey(t, "k"), "   "); !errors.Is(err, ErrInvalidRuntimeContractRule) {
			t.Errorf("error = %v, want %v", err, ErrInvalidRuntimeContractRule)
		}
	})
}

func TestContractRuleJSONRoundTrip(t *testing.T) {
	r := mustContractRule(t, "observe-cpu", "observe cpu utilization")
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContractRule
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != r {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, r)
	}
}

func TestContractRuleUnmarshalToleratesUnknownFields(t *testing.T) {
	var r ContractRule
	err := json.Unmarshal([]byte(`{"key":"observe-cpu","text":"observe cpu","category":"observation"}`), &r)
	if err != nil {
		t.Fatal(err)
	}
	if r.Key() != mustLocalKey(t, "observe-cpu") || r.Text() != "observe cpu" {
		t.Errorf("r = %+v, want key=observe-cpu text=%q", r, "observe cpu")
	}
}

func TestContractRuleMarshalZero(t *testing.T) {
	var r ContractRule
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRuntimeContractRule) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeContractRule)
	}
}

func TestContractRuleUnmarshalRejections(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing key", `{"text":"observe cpu"}`},
		{"null key", `{"key":null,"text":"observe cpu"}`},
		{"missing text", `{"key":"observe-cpu"}`},
		{"null text", `{"key":"observe-cpu","text":null}`},
		{"empty text", `{"key":"observe-cpu","text":""}`},
		{"whitespace text", `{"key":"observe-cpu","text":"   "}`},
		{"malformed JSON", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r ContractRule
			if err := json.Unmarshal([]byte(tt.json), &r); err == nil {
				t.Errorf("%s accepted, want error", tt.json)
			}
		})
	}
}

func TestContractRuleUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	r := mustContractRule(t, "observe-cpu", "observe cpu")
	before := r
	if err := json.Unmarshal([]byte(`{"key":"","text":"x"}`), &r); err == nil {
		t.Fatal("expected error")
	}
	if r != before {
		t.Errorf("receiver mutated on failed decode: got %+v, want %+v", r, before)
	}
}

// --- Runtime Contract Rule combined namespace --------------------------------

// TestContractContentRuleNamespaceCollisions exercises J3-03's corrective
// combined namespace: all four ContractRule collections
// (observationRequirements, violationClassificationRules,
// waiverHandlingRules, enforcementExpectations) share exactly one
// runtime-local key namespace, distinct from the Assertion namespace.
func TestContractContentRuleNamespaceCollisions(t *testing.T) {
	type rules struct {
		observation, classification, waiver, enforcement []ContractRule
	}
	base := func() rules {
		return rules{
			observation:    []ContractRule{mustContractRule(t, "observe-cpu", "observe cpu")},
			classification: []ContractRule{mustContractRule(t, "latency-breach", "latency breach")},
			waiver:         []ContractRule{mustContractRule(t, "defer-oncall", "defer to on-call")},
			enforcement:    []ContractRule{mustContractRule(t, "page-oncall", "page on-call")},
		}
	}
	build := func(r rules) (ContractContent, error) {
		c := mustMinimalContractContent(t)
		var err error
		if c, err = c.WithObservationRequirements(r.observation); err != nil {
			return ContractContent{}, err
		}
		if c, err = c.WithViolationClassificationRules(r.classification); err != nil {
			return ContractContent{}, err
		}
		if c, err = c.WithWaiverHandlingRules(r.waiver); err != nil {
			return ContractContent{}, err
		}
		if c, err = c.WithEnforcementExpectations(r.enforcement); err != nil {
			return ContractContent{}, err
		}
		return c, nil
	}

	t.Run("1 duplicate within observation requirements", func(t *testing.T) {
		r := base()
		r.observation = append(r.observation, mustContractRule(t, "observe-cpu", "observe cpu again"))
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("2 duplicate within violation classification rules", func(t *testing.T) {
		r := base()
		r.classification = append(r.classification, mustContractRule(t, "latency-breach", "latency breach again"))
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("3 duplicate within waiver handling rules", func(t *testing.T) {
		r := base()
		r.waiver = append(r.waiver, mustContractRule(t, "defer-oncall", "defer again"))
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("4 duplicate within enforcement expectations", func(t *testing.T) {
		r := base()
		r.enforcement = append(r.enforcement, mustContractRule(t, "page-oncall", "page again"))
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("5 observation vs classification", func(t *testing.T) {
		r := base()
		r.classification = []ContractRule{mustContractRule(t, "observe-cpu", "collides with observation")}
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("6 observation vs waiver", func(t *testing.T) {
		r := base()
		r.waiver = []ContractRule{mustContractRule(t, "observe-cpu", "collides with observation")}
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("7 observation vs enforcement", func(t *testing.T) {
		r := base()
		r.enforcement = []ContractRule{mustContractRule(t, "observe-cpu", "collides with observation")}
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("8 classification vs waiver", func(t *testing.T) {
		r := base()
		r.waiver = []ContractRule{mustContractRule(t, "latency-breach", "collides with classification")}
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("9 classification vs enforcement", func(t *testing.T) {
		r := base()
		r.enforcement = []ContractRule{mustContractRule(t, "latency-breach", "collides with classification")}
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	t.Run("10 waiver vs enforcement", func(t *testing.T) {
		r := base()
		r.enforcement = []ContractRule{mustContractRule(t, "defer-oncall", "collides with waiver")}
		if _, err := build(r); !errors.Is(err, ErrDuplicateRuntimeLocalKey) {
			t.Errorf("error = %v, want %v", err, ErrDuplicateRuntimeLocalKey)
		}
	})
	// Case 11 (ContractRule and Assertion sharing a key is accepted) is
	// covered by TestContractContentAssertionAndContractRuleShareKeyAccepted
	// below: it needs a full NewContractContent call (Assertions are a
	// constructor argument, not a With* collection) and so does not fit
	// this table's build() helper.
}

// TestContractContentAssertionAndContractRuleShareKeyAccepted documents that
// the Assertion namespace and the combined Runtime Contract Rule namespace
// are independent: the same core.LocalKey may be used once for an Assertion
// and once for a ContractRule without triggering ErrDuplicateRuntimeLocalKey.
func TestContractContentAssertionAndContractRuleShareKeyAccepted(t *testing.T) {
	a := mustAssertion(t, "shared-key")
	c, err := NewContractContent(
		[]RequirementReference{mustRequirementIdentityReference(t, "REQ-1")},
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustEnvironment(t, "production"),
		mustScope(t, "cluster=prod-1"),
		NewUnrestrictedContractApplicability(),
		mustProvenance(t),
		mustAuthority(t),
		[]Assertion{a},
	)
	if err != nil {
		t.Fatal(err)
	}
	c, err = c.WithObservationRequirements([]ContractRule{mustContractRule(t, "shared-key", "an observation requirement")})
	if err != nil {
		t.Errorf("Assertion and ContractRule sharing a key: unexpected error %v", err)
	}
	if _, ok := c.Assertion(mustLocalKey(t, "shared-key")); !ok {
		t.Error("Assertion(shared-key) lookup failed")
	}
	if _, ok := c.ContractRule(mustLocalKey(t, "shared-key")); !ok {
		t.Error("ContractRule(shared-key) lookup failed")
	}
}

func TestContractContentAssertionZeroRejected(t *testing.T) {
	_, err := NewContractContent(
		[]RequirementReference{mustRequirementIdentityReference(t, "REQ-1")},
		mustRuntimeSubjectRef(t, "kubernetes", "pod-1"),
		mustEnvironment(t, "production"),
		mustScope(t, "cluster=prod-1"),
		NewUnrestrictedContractApplicability(),
		mustProvenance(t),
		mustAuthority(t),
		[]Assertion{{}},
	)
	if !errors.Is(err, ErrInvalidRuntimeAssertion) {
		t.Errorf("zero assertion: error = %v, want %v", err, ErrInvalidRuntimeAssertion)
	}
}

// TestContractContentRuleCollectionJSONBehavior verifies, for each of the
// four ContractRule collections independently, that an absent key, an
// explicit null, and an empty array are all equivalent to "declares none",
// that one or several valid elements decode correctly, that an invalid
// element is rejected, and that the receiver is preserved on a failed
// decode -- per Packet J.2.A section 9's per-collection JSON matrix.
func TestContractContentRuleCollectionJSONBehavior(t *testing.T) {
	fields := []string{
		"observation_requirements",
		"violation_classification_rules",
		"waiver_handling_rules",
		"enforcement_expectations",
	}

	baseMap := func(t *testing.T) map[string]json.RawMessage {
		t.Helper()
		data, err := json.Marshal(mustMinimalContractContent(t))
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
			t.Run("absent", func(t *testing.T) {
				m := baseMap(t)
				delete(m, field)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				var c ContractContent
				if err := json.Unmarshal(data, &c); err != nil {
					t.Errorf("absent %s rejected: %v", field, err)
				}
			})
			t.Run("null", func(t *testing.T) {
				m := baseMap(t)
				m[field] = json.RawMessage(`null`)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				var c ContractContent
				if err := json.Unmarshal(data, &c); err != nil {
					t.Errorf("null %s rejected: %v", field, err)
				}
			})
			t.Run("empty array", func(t *testing.T) {
				m := baseMap(t)
				m[field] = json.RawMessage(`[]`)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				var c ContractContent
				if err := json.Unmarshal(data, &c); err != nil {
					t.Errorf("empty array %s rejected: %v", field, err)
				}
			})
			t.Run("one valid element", func(t *testing.T) {
				m := baseMap(t)
				m[field] = json.RawMessage(`[{"key":"k1","text":"rule one"}]`)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				var c ContractContent
				if err := json.Unmarshal(data, &c); err != nil {
					t.Fatalf("one valid element rejected: %v", err)
				}
			})
			t.Run("multiple valid elements", func(t *testing.T) {
				m := baseMap(t)
				m[field] = json.RawMessage(`[{"key":"k1","text":"rule one"},{"key":"k2","text":"rule two"}]`)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				var c ContractContent
				if err := json.Unmarshal(data, &c); err != nil {
					t.Fatalf("multiple valid elements rejected: %v", err)
				}
			})
			t.Run("invalid element", func(t *testing.T) {
				m := baseMap(t)
				m[field] = json.RawMessage(`[{"key":"","text":"rule one"}]`)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				var c ContractContent
				if err := json.Unmarshal(data, &c); err == nil {
					t.Error("invalid element accepted, want error")
				}
			})
			t.Run("receiver preservation", func(t *testing.T) {
				c := mustMinimalContractContent(t)
				m := baseMap(t)
				m[field] = json.RawMessage(`[{"key":"","text":"rule one"}]`)
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(data, &c); err == nil {
					t.Fatal("invalid element accepted, want error")
				}
				if len(c.ObservationRequirements()) != 0 || len(c.ViolationClassificationRules()) != 0 ||
					len(c.WaiverHandlingRules()) != 0 || len(c.EnforcementExpectations()) != 0 {
					t.Error("receiver mutated on failed decode")
				}
			})
		})
	}
}

func TestContractContentUnmarshalMalformed(t *testing.T) {
	var c ContractContent
	if err := json.Unmarshal([]byte(`not json`), &c); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
	if err := json.Unmarshal([]byte(`{"authority":"not an object"}`), &c); err == nil {
		t.Error("malformed authority accepted, want error")
	}
}

func TestContractContentUnmarshalPreservesReceiverOnFailure(t *testing.T) {
	c := mustMinimalContractContent(t)
	originalData, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"requirements":[]}`), &c); err == nil {
		t.Fatal("empty requirements accepted, want error")
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(originalData) {
		t.Error("failed unmarshal did not preserve receiver")
	}
}

func TestContractCoreAccessor(t *testing.T) {
	artifact := mustArtifact(t, "ART-1", ArtifactTypeRuntimeContract)
	c, err := NewContract(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if c.Core().ID() != artifact.ID() || c.Core().Type() != artifact.Type() {
		t.Error("Core() mismatch")
	}
}

func TestContractRevisionContentAccessor(t *testing.T) {
	contract := mustContract(t, "ART-1")
	revision := mustArtifactRevision(t, "ART-1", "REV-1")
	content := mustMinimalContractContent(t)
	r, err := NewContractRevision(contract, revision, content)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Content().Requirements()) != len(content.Requirements()) {
		t.Error("Content() mismatch")
	}
}

func TestContractApplicabilityKindZero(t *testing.T) {
	var a ContractApplicability
	if a.Kind() != "" {
		t.Errorf("zero-value Kind() = %q, want empty string", a.Kind())
	}
	if got := NewUnrestrictedContractApplicability().Kind(); got != "unrestricted" {
		t.Errorf("Kind() = %q, want %q", got, "unrestricted")
	}
	scoped, err := NewScopedContractApplicability(mustScope(t, "x"))
	if err != nil {
		t.Fatal(err)
	}
	if got := scoped.Kind(); got != "scoped" {
		t.Errorf("Kind() = %q, want %q", got, "scoped")
	}
}

func TestContractContentIsZero(t *testing.T) {
	var c ContractContent
	if !c.IsZero() {
		t.Error("zero-value ContractContent.IsZero() = false, want true")
	}
	if mustMinimalContractContent(t).IsZero() {
		t.Error("valid ContractContent reports IsZero() = true")
	}
}

func TestContractContentJSONRoundTrip(t *testing.T) {
	c := mustMinimalContractContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContractContent
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

func TestContractContentMarshalZero(t *testing.T) {
	var c ContractContent
	if _, err := json.Marshal(c); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractContentJSONRoundTripWithExtension(t *testing.T) {
	c := mustMinimalContractContent(t)
	ext, err := core.NewExtension().With("product", json.RawMessage(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	c = c.WithExtension(ext)

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContractContent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Extension().IsZero() {
		t.Error("decoded ContractContent lost its extension")
	}
	data2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
	}
}

func TestContractContentAuthorityMissingOrNullRejected(t *testing.T) {
	c := mustMinimalContractContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	delete(m, "authority")
	missingData, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContractContent
	if err := json.Unmarshal(missingData, &decoded); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("missing authority: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}

	m["authority"] = json.RawMessage("null")
	nullData, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(nullData, &decoded); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("null authority: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractContentNoForbiddenWireKeys(t *testing.T) {
	c := mustMinimalContractContent(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"bound", "active_deployment", "activeDeployment", "deployed",
		"compliant", "compliance", "state", "status", "lifecycle",
		"state_assignment", "relation", "source", "target", "version",
		"current", "latest", "effective", "incident", "verdict",
		"outcome_authority",
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range forbidden {
		if _, ok := m[key]; ok {
			t.Errorf("wire form contains forbidden key %q", key)
		}
	}
}

// --- ContractRevision ---------------------------------------------------------

func TestNewContractRevision(t *testing.T) {
	contract := mustContract(t, "ART-1")
	revision := mustArtifactRevision(t, "ART-1", "REV-1")
	content := mustMinimalContractContent(t)

	r, err := NewContractRevision(contract, revision, content)
	if err != nil {
		t.Fatal(err)
	}
	if r.IsZero() {
		t.Error("valid ContractRevision reports IsZero() = true")
	}
	ref, err := r.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.ArtifactID() != contract.ID() || ref.RevisionID() != revision.RevisionID() {
		t.Error("Ref() mismatch")
	}
}

func TestNewContractRevisionArtifactIDMismatch(t *testing.T) {
	contract := mustContract(t, "ART-1")
	revision := mustArtifactRevision(t, "ART-OTHER", "REV-1")
	content := mustMinimalContractContent(t)

	if _, err := NewContractRevision(contract, revision, content); !errors.Is(err, ErrRuntimeContractArtifactIDMismatch) {
		t.Errorf("error = %v, want %v", err, ErrRuntimeContractArtifactIDMismatch)
	}
}

func TestNewContractRevisionZeroValues(t *testing.T) {
	contract := mustContract(t, "ART-1")
	revision := mustArtifactRevision(t, "ART-1", "REV-1")
	content := mustMinimalContractContent(t)

	if _, err := NewContractRevision(Contract{}, revision, content); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero contract: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
	if _, err := NewContractRevision(contract, core.ArtifactRevision{}, content); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero revision: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
	if _, err := NewContractRevision(contract, revision, ContractContent{}); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero content: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractRevisionJSONRoundTrip(t *testing.T) {
	contract := mustContract(t, "ART-1")
	revision := mustArtifactRevision(t, "ART-1", "REV-1")
	content := mustMinimalContractContent(t)

	r, err := NewContractRevision(contract, revision, content)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContractRevision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Core().ArtifactID() != r.Core().ArtifactID() || decoded.Core().RevisionID() != r.Core().RevisionID() {
		t.Error("round trip mismatch")
	}
}

func TestContractRevisionMarshalZero(t *testing.T) {
	var r ContractRevision
	if _, err := json.Marshal(r); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("zero marshal: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractRevisionUnmarshalMalformed(t *testing.T) {
	var r ContractRevision
	if err := json.Unmarshal([]byte(`not json`), &r); err == nil {
		t.Error("malformed JSON accepted, want error")
	}
	if err := json.Unmarshal([]byte(`{"core":{},"content":{}}`), &r); !errors.Is(err, ErrInvalidRuntimeContract) {
		t.Errorf("empty core/content: error = %v, want %v", err, ErrInvalidRuntimeContract)
	}
}

func TestContractRevisionWireForm(t *testing.T) {
	contract := mustContract(t, "ART-1")
	revision := mustArtifactRevision(t, "ART-1", "REV-1")
	content := mustMinimalContractContent(t)

	r, err := NewContractRevision(contract, revision, content)
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
	if _, ok := m["core"]; !ok {
		t.Error(`wire form missing "core" key`)
	}
	if _, ok := m["content"]; !ok {
		t.Error(`wire form missing "content" key`)
	}
	if len(m) != 2 {
		t.Errorf("wire form has %d top-level keys, want 2", len(m))
	}
}
