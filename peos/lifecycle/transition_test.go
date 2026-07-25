package lifecycle

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestNewTransitionDefinitionMultipleSourcesAndTargets(t *testing.T) {
	sources := []StateID{mustStateID(t, "draft"), mustStateID(t, "in-review")}
	targets := []StateID{mustStateID(t, "accepted"), mustStateID(t, "rejected")}
	tr, err := NewTransitionDefinition(mustTransitionID(t, "resolve"), sources, targets)
	if err != nil {
		t.Fatal(err)
	}
	if len(tr.SourceStates()) != 2 || len(tr.TargetStates()) != 2 {
		t.Errorf("SourceStates()/TargetStates() = %d/%d, want 2/2", len(tr.SourceStates()), len(tr.TargetStates()))
	}
}

func TestNewTransitionDefinitionSelfTransitionAccepted(t *testing.T) {
	draft := mustStateID(t, "draft")
	if _, err := NewTransitionDefinition(mustTransitionID(t, "save-draft"), []StateID{draft}, []StateID{draft}); err != nil {
		t.Errorf("self-transition unexpectedly rejected: %v", err)
	}
}

func TestMultipleTransitionsBetweenSameStatePairAccepted(t *testing.T) {
	draft := mustStateID(t, "draft")
	accepted := mustStateID(t, "accepted")
	t1, err := NewTransitionDefinition(mustTransitionID(t, "fast-track"), []StateID{draft}, []StateID{accepted})
	if err != nil {
		t.Fatal(err)
	}
	t2, err := NewTransitionDefinition(mustTransitionID(t, "standard-track"), []StateID{draft}, []StateID{accepted})
	if err != nil {
		t.Fatal(err)
	}
	if t1.ID().Equal(t2.ID()) {
		t.Fatal("expected distinct Transition IDs")
	}
}

func TestNewTransitionDefinitionEmptySourceRejected(t *testing.T) {
	if _, err := NewTransitionDefinition(mustTransitionID(t, "t1"), nil, []StateID{mustStateID(t, "accepted")}); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransition)
	}
}

func TestNewTransitionDefinitionEmptyTargetRejected(t *testing.T) {
	if _, err := NewTransitionDefinition(mustTransitionID(t, "t1"), []StateID{mustStateID(t, "draft")}, nil); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransition)
	}
}

func TestNewTransitionDefinitionZeroIDRejected(t *testing.T) {
	if _, err := NewTransitionDefinition(TransitionID{}, []StateID{mustStateID(t, "draft")}, []StateID{mustStateID(t, "accepted")}); !errors.Is(err, ErrInvalidTransitionID) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransitionID)
	}
}

func TestNewTransitionDefinitionDuplicateSourceRejected(t *testing.T) {
	draft := mustStateID(t, "draft")
	if _, err := NewTransitionDefinition(mustTransitionID(t, "t1"), []StateID{draft, draft}, []StateID{mustStateID(t, "accepted")}); !errors.Is(err, ErrDuplicateStateID) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateStateID)
	}
}

func TestNewTransitionDefinitionDuplicateTargetRejected(t *testing.T) {
	accepted := mustStateID(t, "accepted")
	if _, err := NewTransitionDefinition(mustTransitionID(t, "t1"), []StateID{mustStateID(t, "draft")}, []StateID{accepted, accepted}); !errors.Is(err, ErrDuplicateStateID) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateStateID)
	}
}

func TestTransitionDefinitionDefensiveCopies(t *testing.T) {
	sources := []StateID{mustStateID(t, "draft")}
	tr, err := NewTransitionDefinition(mustTransitionID(t, "t1"), sources, []StateID{mustStateID(t, "accepted")})
	if err != nil {
		t.Fatal(err)
	}
	sources[0] = mustStateID(t, "mutated")
	if tr.SourceStates()[0].String() != "draft" {
		t.Error("mutating caller's input slice affected the constructed TransitionDefinition")
	}
	got := tr.SourceStates()
	got[0] = mustStateID(t, "mutated-output")
	if tr.SourceStates()[0].String() != "draft" {
		t.Error("mutating the returned slice affected the TransitionDefinition's internal state")
	}
}

func TestTransitionDefinitionJSONRoundTrip(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTransitionDefinition(mustTransitionID(t, "submit"), []StateID{mustStateID(t, "draft")}, []StateID{mustStateID(t, "accepted")})
	if err != nil {
		t.Fatal(err)
	}
	tr = tr.WithExtension(ext)

	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TransitionDefinition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.ID().Equal(tr.ID()) {
		t.Error("round trip changed ID")
	}
	if len(decoded.SourceStates()) != 1 || !decoded.SourceStates()[0].Equal(tr.SourceStates()[0]) {
		t.Error("round trip changed SourceStates")
	}
	if len(decoded.TargetStates()) != 1 || !decoded.TargetStates()[0].Equal(tr.TargetStates()[0]) {
		t.Error("round trip changed TargetStates")
	}
	got, ok := decoded.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestTransitionDefinitionUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original, err := NewTransitionDefinition(mustTransitionID(t, "submit"), []StateID{mustStateID(t, "draft")}, []StateID{mustStateID(t, "accepted")})
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	payload := `{"id":"submit","source_states":[],"target_states":["accepted"]}`
	if err := json.Unmarshal([]byte(payload), &receiver); err == nil {
		t.Fatal("empty source_states accepted, want error")
	}
	if !receiver.ID().Equal(original.ID()) {
		t.Error("failed Unmarshal changed receiver")
	}
}

func TestTransitionDefinitionZeroMarshalRejected(t *testing.T) {
	var tr TransitionDefinition
	if _, err := json.Marshal(tr); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("error = %v, want %v", err, ErrInvalidTransition)
	}
}
