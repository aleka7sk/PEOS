package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustLifecycleDefinitionID(t *testing.T, value string) core.LifecycleDefinitionID {
	t.Helper()
	id, err := core.NewLifecycleDefinitionID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustDefinitionRef(t *testing.T, value string) core.LifecycleDefinitionRef {
	t.Helper()
	ref, err := core.NewLifecycleDefinitionRef(mustLifecycleDefinitionID(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustScope(t *testing.T, namespace, expression string) core.Scope {
	t.Helper()
	vocab, err := core.NewVocabularyValue(namespace, "condition")
	if err != nil {
		t.Fatal(err)
	}
	s, err := core.NewScope(vocab, expression)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustProvenance(t *testing.T) core.Provenance {
	t.Helper()
	ts, err := core.NewTimestamp(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	actor, err := core.NewActorRef("peos-cli", "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	return core.NewProvenance().WithActor(actor).WithRecordedAt(ts)
}

func mustSubjectType(t *testing.T, value string) core.VocabularyValue {
	t.Helper()
	v, err := core.NewVocabularyValue("peos", value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// validDefinitionVersionParts returns a minimally valid set of
// NewDefinitionVersion arguments a test can freely mutate one field of.
func validDefinitionVersionParts(t *testing.T) (
	id core.LifecycleDefinitionVersionID,
	definition core.LifecycleDefinitionRef,
	scope core.Scope,
	subjectTypes []core.VocabularyValue,
	states []State,
	initialStates []StateID,
	transitions []TransitionDefinition,
	entryTransition TransitionID,
	provenance core.Provenance,
) {
	t.Helper()
	id, err := core.NewLifecycleDefinitionVersionID("V1")
	if err != nil {
		t.Fatal(err)
	}
	definition = mustDefinitionRef(t, "LC-REVIEW-1")
	scope = mustScope(t, "product-x", "always")
	subjectTypes = []core.VocabularyValue{mustSubjectType(t, "requirement")}

	draft, err := NewState(mustStateID(t, "draft"), "Not yet reviewed")
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := NewState(mustStateID(t, "accepted"), "Reviewed and accepted")
	if err != nil {
		t.Fatal(err)
	}
	states = []State{draft, accepted}
	initialStates = []StateID{mustStateID(t, "draft")}

	submit, err := NewTransitionDefinition(mustTransitionID(t, "submit-for-review"), []StateID{mustStateID(t, "draft")}, []StateID{mustStateID(t, "accepted")})
	if err != nil {
		t.Fatal(err)
	}
	transitions = []TransitionDefinition{submit}
	entryTransition = mustTransitionID(t, "submit-for-review")
	provenance = mustProvenance(t)
	return
}

func mustDefinitionVersion(t *testing.T) DefinitionVersion {
	t.Helper()
	v, err := NewDefinitionVersion(validDefinitionVersionParts(t))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestNewDefinitionZeroIDRejected(t *testing.T) {
	if _, err := NewDefinition(core.LifecycleDefinitionID{}); !errors.Is(err, ErrInvalidLifecycleDefinition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinition)
	}
}

func TestDefinitionRef(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	ref, err := d.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.LifecycleDefinitionID() != d.ID() {
		t.Error("Ref() component mismatch")
	}
}

func TestDefinitionJSONRoundTrip(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Definition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != d.ID() {
		t.Error("round trip mismatch")
	}
}

func TestDefinitionZeroMarshalRejected(t *testing.T) {
	var d Definition
	if _, err := json.Marshal(d); !errors.Is(err, ErrInvalidLifecycleDefinition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinition)
	}
}

func TestNewDefinitionVersionMinimumOneStateOneTransition(t *testing.T) {
	id, definition, scope, _, _, _, _, _, provenance := validDefinitionVersionParts(t)
	v, err := NewDefinitionVersion(validDefinitionVersionParts(t))
	if err != nil {
		t.Fatalf("minimally valid DefinitionVersion unexpectedly rejected: %v", err)
	}
	if v.ID() != id {
		t.Error("ID() mismatch")
	}
	if v.Definition() != definition {
		t.Error("Definition() mismatch")
	}
	if !v.Scope().Equal(scope) {
		t.Error("Scope() mismatch")
	}
	gotActor, gotOK := v.Provenance().Actor()
	wantActor, wantOK := provenance.Actor()
	if gotOK != wantOK || gotActor != wantActor {
		t.Error("Provenance() mismatch")
	}
	if len(v.SubjectTypes()) == 0 || len(v.States()) == 0 || len(v.InitialStates()) == 0 || len(v.Transitions()) == 0 {
		t.Error("expected non-empty collections")
	}
}

func TestNewDefinitionVersionMultipleInitialStates(t *testing.T) {
	id, definition, scope, subjectTypes, states, _, transitions, entry, provenance := validDefinitionVersionParts(t)
	rejected, err := NewState(mustStateID(t, "rejected"), "Reviewed and rejected")
	if err != nil {
		t.Fatal(err)
	}
	states = append(states, rejected)
	initialStates := []StateID{mustStateID(t, "draft"), mustStateID(t, "rejected")}
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance); err != nil {
		t.Fatalf("multiple initial states unexpectedly rejected: %v", err)
	}
}

func TestNewDefinitionVersionDuplicateStateIDRejected(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance := validDefinitionVersionParts(t)
	dup, err := NewState(mustStateID(t, "draft"), "Duplicate")
	if err != nil {
		t.Fatal(err)
	}
	states = append(states, dup)
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance); !errors.Is(err, ErrDuplicateStateID) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateStateID)
	}
}

func TestNewDefinitionVersionDuplicateTransitionIDRejected(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance := validDefinitionVersionParts(t)
	dup, err := NewTransitionDefinition(mustTransitionID(t, "submit-for-review"), []StateID{mustStateID(t, "accepted")}, []StateID{mustStateID(t, "draft")})
	if err != nil {
		t.Fatal(err)
	}
	transitions = append(transitions, dup)
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance); !errors.Is(err, ErrDuplicateTransitionID) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateTransitionID)
	}
}

func TestNewDefinitionVersionUnknownInitialStateRejected(t *testing.T) {
	id, definition, scope, subjectTypes, states, _, transitions, entry, provenance := validDefinitionVersionParts(t)
	initialStates := []StateID{mustStateID(t, "does-not-exist")}
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance); !errors.Is(err, ErrInvalidInitialState) {
		t.Errorf("error = %v, want %v", err, ErrInvalidInitialState)
	}
}

func TestNewDefinitionVersionUnknownTransitionSourceRejected(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, _, entry, provenance := validDefinitionVersionParts(t)
	bad, err := NewTransitionDefinition(mustTransitionID(t, "bogus"), []StateID{mustStateID(t, "nonexistent")}, []StateID{mustStateID(t, "accepted")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, []TransitionDefinition{bad}, entry, provenance); !errors.Is(err, ErrUnknownStateID) {
		t.Errorf("error = %v, want %v", err, ErrUnknownStateID)
	}
	_ = bad
}

func TestNewDefinitionVersionUnknownTransitionTargetRejected(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, _, _, provenance := validDefinitionVersionParts(t)
	bad, err := NewTransitionDefinition(mustTransitionID(t, "submit-for-review"), []StateID{mustStateID(t, "draft")}, []StateID{mustStateID(t, "nonexistent")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, []TransitionDefinition{bad}, mustTransitionID(t, "submit-for-review"), provenance); !errors.Is(err, ErrUnknownStateID) {
		t.Errorf("error = %v, want %v", err, ErrUnknownStateID)
	}
}

func TestNewDefinitionVersionUnknownEntryTransitionRejected(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, transitions, _, provenance := validDefinitionVersionParts(t)
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, mustTransitionID(t, "does-not-exist"), provenance); !errors.Is(err, ErrInvalidEntryTransition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidEntryTransition)
	}
}

func TestNewDefinitionVersionOpenSubjectTypeAccepted(t *testing.T) {
	id, definition, scope, _, states, initialStates, transitions, entry, provenance := validDefinitionVersionParts(t)
	subjectTypes := []core.VocabularyValue{mustSubjectType(t, "totally-unknown-subject-type")}
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance); err != nil {
		t.Errorf("open subject type unexpectedly rejected: %v", err)
	}
}

func TestNewDefinitionVersionRequiresScope(t *testing.T) {
	id, definition, _, subjectTypes, states, initialStates, transitions, entry, provenance := validDefinitionVersionParts(t)
	if _, err := NewDefinitionVersion(id, definition, core.Scope{}, subjectTypes, states, initialStates, transitions, entry, provenance); !errors.Is(err, ErrInvalidLifecycleDefinitionVersion) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinitionVersion)
	}
}

func TestNewDefinitionVersionRequiresProvenance(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, transitions, entry, _ := validDefinitionVersionParts(t)
	if _, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, core.Provenance{}); !errors.Is(err, ErrInvalidLifecycleDefinitionVersion) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinitionVersion)
	}
}

func TestDefinitionVersionDefensiveCopyOfEverySlice(t *testing.T) {
	id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance := validDefinitionVersionParts(t)
	v, err := NewDefinitionVersion(id, definition, scope, subjectTypes, states, initialStates, transitions, entry, provenance)
	if err != nil {
		t.Fatal(err)
	}
	// Mutate caller-supplied inputs after construction.
	subjectTypes[0] = mustSubjectType(t, "mutated")
	states[0], _ = NewState(mustStateID(t, "mutated"), "mutated")
	initialStates[0] = mustStateID(t, "mutated")
	transitions[0] = TransitionDefinition{}

	if v.SubjectTypes()[0].String() == "peos:mutated" {
		t.Error("mutating caller's subjectTypes slice affected the constructed DefinitionVersion")
	}
	if v.States()[0].ID().String() == "mutated" {
		t.Error("mutating caller's states slice affected the constructed DefinitionVersion")
	}
	if v.InitialStates()[0].String() == "mutated" {
		t.Error("mutating caller's initialStates slice affected the constructed DefinitionVersion")
	}
	if v.Transitions()[0].IsZero() {
		t.Error("mutating caller's transitions slice affected the constructed DefinitionVersion")
	}

	// Mutate the returned slices too.
	out := v.States()
	out[0], _ = NewState(mustStateID(t, "mutated-out"), "mutated")
	if v.States()[0].ID().String() == "mutated-out" {
		t.Error("mutating a returned slice affected the DefinitionVersion's internal state")
	}
}

func TestDefinitionVersionJSONRoundTrip(t *testing.T) {
	v := mustDefinitionVersion(t)
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionVersion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != v.ID() {
		t.Error("round trip changed ID")
	}
	if len(decoded.States()) != len(v.States()) {
		t.Error("round trip changed States")
	}
	if !decoded.EntryTransition().Equal(v.EntryTransition()) {
		t.Error("round trip changed EntryTransition")
	}
}

func TestDefinitionVersionJSONUnknownOrdinaryFieldIgnored(t *testing.T) {
	v := mustDefinitionVersion(t)
	data, err := json.Marshal(v)
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
	var decoded DefinitionVersion
	if err := json.Unmarshal(patched, &decoded); err != nil {
		t.Fatalf("unknown ordinary field unexpectedly rejected: %v", err)
	}
}

func TestDefinitionVersionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := mustDefinitionVersion(t)
	receiver := original
	payload := `{"id":"V1"}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("incomplete payload accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestDefinitionVersionZeroMarshalRejected(t *testing.T) {
	var v DefinitionVersion
	if _, err := json.Marshal(v); !errors.Is(err, ErrInvalidLifecycleDefinitionVersion) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinitionVersion)
	}
}

func TestDefinitionVersionRef(t *testing.T) {
	v := mustDefinitionVersion(t)
	ref, err := v.Ref()
	if err != nil {
		t.Fatal(err)
	}
	if ref.LifecycleDefinitionID() != v.Definition().LifecycleDefinitionID() {
		t.Error("Ref() definition component mismatch")
	}
	if ref.VersionID() != v.ID() {
		t.Error("Ref() version component mismatch")
	}
}

// --- Packet E.1: Definition Artifact binding ---------------------------

func mustArtifactRef(t *testing.T, artifactID string) core.ArtifactRef {
	t.Helper()
	ref, err := core.NewArtifactRef(mustArtifactID(t, artifactID))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestDefinitionArtifactAbsentByDefault(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Artifact(); ok {
		t.Error("Artifact() ok = true for a Definition with no declared Artifact binding")
	}
}

func TestDefinitionWithArtifactValid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	artifact := mustArtifactRef(t, "LC-ARTIFACT-1")
	withArtifact, err := d.WithArtifact(artifact)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withArtifact.Artifact()
	if !ok || got != artifact {
		t.Errorf("Artifact() = (%v, %v), want (%v, true)", got, ok, artifact)
	}
}

func TestDefinitionWithArtifactZeroRejected(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.WithArtifact(core.ArtifactRef{}); !errors.Is(err, ErrInvalidLifecycleDefinition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinition)
	}
}

func TestDefinitionWithArtifactPreservesID(t *testing.T) {
	id := mustLifecycleDefinitionID(t, "LC-REVIEW-1")
	d, err := NewDefinition(id)
	if err != nil {
		t.Fatal(err)
	}
	withArtifact, err := d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	if withArtifact.ID() != id {
		t.Error("WithArtifact changed Definition's own LifecycleDefinitionID")
	}
	// d itself must remain unaffected (value receiver, returns a copy).
	if _, ok := d.Artifact(); ok {
		t.Error("WithArtifact mutated the original receiver: d now has an Artifact binding")
	}
}

func TestDefinitionWithoutArtifactClearsPresence(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	withArtifact, err := d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	cleared := withArtifact.WithoutArtifact()
	if _, ok := cleared.Artifact(); ok {
		t.Error("Artifact() ok = true after WithoutArtifact()")
	}
}

func TestDefinitionJSONRoundTripWithArtifactBinding(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Definition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID() != d.ID() {
		t.Error("round trip changed ID")
	}
	wantArtifact, _ := d.Artifact()
	gotArtifact, ok := decoded.Artifact()
	if !ok || gotArtifact != wantArtifact {
		t.Errorf("Artifact() = (%v, %v), want (%v, true)", gotArtifact, ok, wantArtifact)
	}
}

func TestDefinitionJSONArtifactKeyOmittedWithoutBinding(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["artifact"]; present {
		t.Error(`"artifact" key present in Definition JSON despite no binding being set`)
	}
}

func TestDefinitionJSONArtifactExplicitNullRejected(t *testing.T) {
	payload := `{"id":"LC-REVIEW-1","artifact":null}`
	var d Definition
	if err := json.Unmarshal([]byte(payload), &d); !errors.Is(err, ErrInvalidLifecycleDefinition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinition)
	}
}

func TestDefinitionJSONArtifactEmptyObjectRejected(t *testing.T) {
	payload := `{"id":"LC-REVIEW-1","artifact":{}}`
	var d Definition
	if err := json.Unmarshal([]byte(payload), &d); err == nil {
		t.Fatal("empty artifact object accepted, want error")
	}
}

func TestDefinitionJSONUnknownOrdinaryFieldIgnored(t *testing.T) {
	payload := `{"id":"LC-REVIEW-1","bogus_field":123}`
	var d Definition
	if err := json.Unmarshal([]byte(payload), &d); err != nil {
		t.Fatalf("unknown ordinary field unexpectedly rejected: %v", err)
	}
}

func TestDefinitionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	payload := `{"id":"LC-REVIEW-1","artifact":null}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("explicit null artifact accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver's ID")
	}
	gotArtifact, gotOK := receiver.Artifact()
	wantArtifact, wantOK := original.Artifact()
	if gotOK != wantOK || gotArtifact != wantArtifact {
		t.Error("failed Unmarshal changed receiver's Artifact binding")
	}
}

// --- Packet E.1: DefinitionVersion ArtifactRevision binding -------------

func TestDefinitionVersionArtifactRevisionAbsentByDefault(t *testing.T) {
	v := mustDefinitionVersion(t)
	if _, ok := v.ArtifactRevision(); ok {
		t.Error("ArtifactRevision() ok = true for a DefinitionVersion with no declared binding")
	}
}

func TestDefinitionVersionWithArtifactRevisionValid(t *testing.T) {
	v := mustDefinitionVersion(t)
	ref := mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1")
	withRef, err := v.WithArtifactRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := withRef.ArtifactRevision()
	if !ok || got != ref {
		t.Errorf("ArtifactRevision() = (%v, %v), want (%v, true)", got, ok, ref)
	}
}

func TestDefinitionVersionWithArtifactRevisionZeroRejected(t *testing.T) {
	v := mustDefinitionVersion(t)
	if _, err := v.WithArtifactRevision(core.ArtifactRevisionRef{}); !errors.Is(err, ErrInvalidLifecycleDefinitionVersion) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinitionVersion)
	}
}

func TestDefinitionVersionWithArtifactRevisionPreservesExistingFields(t *testing.T) {
	v := mustDefinitionVersion(t)
	withRef, err := v.WithArtifactRevision(mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	if withRef.ID() != v.ID() || withRef.Definition() != v.Definition() {
		t.Error("WithArtifactRevision changed identity fields")
	}
	if !withRef.Scope().Equal(v.Scope()) {
		t.Error("WithArtifactRevision changed Scope")
	}
	if len(withRef.States()) != len(v.States()) || len(withRef.Transitions()) != len(v.Transitions()) {
		t.Error("WithArtifactRevision changed collections")
	}
	if !withRef.EntryTransition().Equal(v.EntryTransition()) {
		t.Error("WithArtifactRevision changed EntryTransition")
	}
	// v itself must remain unaffected.
	if _, ok := v.ArtifactRevision(); ok {
		t.Error("WithArtifactRevision mutated the original receiver")
	}
}

func TestDefinitionVersionWithoutArtifactRevisionClearsPresence(t *testing.T) {
	v := mustDefinitionVersion(t)
	withRef, err := v.WithArtifactRevision(mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	cleared := withRef.WithoutArtifactRevision()
	if _, ok := cleared.ArtifactRevision(); ok {
		t.Error("ArtifactRevision() ok = true after WithoutArtifactRevision()")
	}
}

func TestDefinitionVersionJSONRoundTripWithArtifactRevision(t *testing.T) {
	v := mustDefinitionVersion(t)
	ref := mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1")
	v, err := v.WithArtifactRevision(ref)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionVersion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got, ok := decoded.ArtifactRevision()
	if !ok || got != ref {
		t.Errorf("ArtifactRevision() = (%v, %v), want (%v, true)", got, ok, ref)
	}
}

func TestDefinitionVersionJSONArtifactRevisionKeyOmittedWhenAbsent(t *testing.T) {
	v := mustDefinitionVersion(t)
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["artifact_revision"]; present {
		t.Error(`"artifact_revision" key present despite no binding being set`)
	}
}

func TestDefinitionVersionJSONArtifactRevisionExplicitNullRejected(t *testing.T) {
	v := mustDefinitionVersion(t)
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["artifact_revision"] = json.RawMessage(`null`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionVersion
	if err := json.Unmarshal(patched, &decoded); !errors.Is(err, ErrInvalidLifecycleDefinitionVersion) {
		t.Errorf("error = %v, want %v", err, ErrInvalidLifecycleDefinitionVersion)
	}
}

func TestDefinitionVersionJSONArtifactRevisionEmptyObjectRejected(t *testing.T) {
	v := mustDefinitionVersion(t)
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["artifact_revision"] = json.RawMessage(`{}`)
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionVersion
	if err := json.Unmarshal(patched, &decoded); err == nil {
		t.Fatal("empty artifact_revision object accepted, want error")
	}
}

func TestDefinitionVersionUnmarshalJSONFailurePreservesReceiverWithArtifactRevision(t *testing.T) {
	original := mustDefinitionVersion(t)
	original, err := original.WithArtifactRevision(mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	payload := `{"id":"V1"}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("incomplete payload accepted, want error")
	}
	if receiver.ID() != original.ID() {
		t.Error("failed Unmarshal changed receiver's ID")
	}
	gotRef, gotOK := receiver.ArtifactRevision()
	wantRef, wantOK := original.ArtifactRevision()
	if gotOK != wantOK || gotRef != wantRef {
		t.Error("failed Unmarshal changed receiver's ArtifactRevision binding")
	}
}

// TestExisting129dd70StyleJSONRemainsValid proves backward compatibility:
// a DefinitionVersion JSON payload with no "artifact_revision" key at all
// (the only shape that existed at commit 129dd70, before this binding was
// added) still decodes correctly, and re-marshals without introducing the
// new key.
func TestExisting129dd70StyleJSONRemainsValid(t *testing.T) {
	payload := `{
		"id": "V1",
		"definition": {"lifecycle_definition_id": "LC-REVIEW-1"},
		"scope": {"kind": "product-x:condition", "expression": "always"},
		"subject_types": ["peos:requirement"],
		"states": [
			{"id": "draft", "meaning": "Not yet reviewed"},
			{"id": "accepted", "meaning": "Reviewed and accepted"}
		],
		"initial_states": ["draft"],
		"transitions": [
			{"id": "submit-for-review", "source_states": ["draft"], "target_states": ["accepted"]}
		],
		"entry_transition": "submit-for-review",
		"provenance": {"actor": {"namespace": "peos-cli", "identifier": "svc-1"}, "recorded_at": "2026-01-01T00:00:00Z"}
	}`
	var v DefinitionVersion
	if err := json.Unmarshal([]byte(payload), &v); err != nil {
		t.Fatalf("pre-E.1 JSON shape unexpectedly rejected: %v", err)
	}
	if _, ok := v.ArtifactRevision(); ok {
		t.Error("ArtifactRevision() ok = true for a payload with no artifact_revision key")
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["artifact_revision"]; present {
		t.Error("re-marshal introduced an artifact_revision key that was not present in the original payload")
	}
}

// --- Packet E.1: Binding consistency validation -------------------------

func TestValidateArtifactBindingNeitherBoundIsValid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t)
	if err := v.ValidateArtifactBinding(d); err != nil {
		t.Errorf("neither side bound: unexpected error: %v", err)
	}
}

func TestValidateArtifactBindingBothConsistentIsValid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t)
	v, err = v.WithArtifactRevision(mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateArtifactBinding(d); err != nil {
		t.Errorf("consistent bindings: unexpected error: %v", err)
	}
}

func TestValidateArtifactBindingDefinitionOnlyIsInvalid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t)
	if err := v.ValidateArtifactBinding(d); !errors.Is(err, ErrArtifactBindingMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactBindingMismatch)
	}
}

func TestValidateArtifactBindingVersionOnlyIsInvalid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t)
	v, err = v.WithArtifactRevision(mustArtifactRevisionRef(t, "LC-ARTIFACT-1", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateArtifactBinding(d); !errors.Is(err, ErrArtifactBindingMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactBindingMismatch)
	}
}

func TestValidateArtifactBindingArtifactIDMismatchIsInvalid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t)
	v, err = v.WithArtifactRevision(mustArtifactRevisionRef(t, "LC-ARTIFACT-DIFFERENT", "REV-1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateArtifactBinding(d); !errors.Is(err, ErrArtifactBindingMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactBindingMismatch)
	}
}

func TestValidateArtifactBindingDefinitionIDMismatchIsInvalid(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-SOME-OTHER-DEFINITION"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t) // v.Definition() refers to "LC-REVIEW-1"
	if err := v.ValidateArtifactBinding(d); !errors.Is(err, ErrArtifactBindingMismatch) {
		t.Errorf("error = %v, want %v", err, ErrArtifactBindingMismatch)
	}
}

func TestValidateArtifactBindingReceiversUnchanged(t *testing.T) {
	d, err := NewDefinition(mustLifecycleDefinitionID(t, "LC-REVIEW-1"))
	if err != nil {
		t.Fatal(err)
	}
	d, err = d.WithArtifact(mustArtifactRef(t, "LC-ARTIFACT-1"))
	if err != nil {
		t.Fatal(err)
	}
	v := mustDefinitionVersion(t)
	beforeArtifact, beforeOK := d.Artifact()
	beforeID := v.ID()
	_ = v.ValidateArtifactBinding(d) // deliberately invalid (version has no binding); return value ignored
	afterArtifact, afterOK := d.Artifact()
	if beforeOK != afterOK || beforeArtifact != afterArtifact {
		t.Error("ValidateArtifactBinding mutated the Definition receiver")
	}
	if v.ID() != beforeID {
		t.Error("ValidateArtifactBinding mutated the DefinitionVersion receiver")
	}
}
