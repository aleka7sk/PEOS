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
