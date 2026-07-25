package requirement

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func TestSubjectCombinationIndependent(t *testing.T) {
	if SubjectCombinationIndependent.IsZero() {
		t.Error("SubjectCombinationIndependent reports IsZero() = true")
	}
	if got, want := SubjectCombinationIndependent.Value().String(), "peos:independent"; got != want {
		t.Errorf("Value().String() = %q, want %q", got, want)
	}
}

func TestSubjectCombinationCollective(t *testing.T) {
	if SubjectCombinationCollective.IsZero() {
		t.Error("SubjectCombinationCollective reports IsZero() = true")
	}
	if got, want := SubjectCombinationCollective.Value().String(), "peos:collective"; got != want {
		t.Errorf("Value().String() = %q, want %q", got, want)
	}
}

func TestNewSubjectCombinationCustomVocabulary(t *testing.T) {
	v := mustVocab(t, "product-x", "weighted-majority")
	sc, err := NewSubjectCombination(v)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Value() != v {
		t.Errorf("Value() = %v, want %v", sc.Value(), v)
	}
}

func TestNewSubjectCombinationZeroRejected(t *testing.T) {
	if _, err := NewSubjectCombination(core.VocabularyValue{}); !errors.Is(err, ErrInvalidSubjectCombination) {
		t.Errorf("error = %v, want %v", err, ErrInvalidSubjectCombination)
	}
}

func TestSubjectCombinationJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(SubjectCombinationCollective)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `"peos:collective"`; got != want {
		t.Errorf("Marshal output = %s, want %s", got, want)
	}
	var decoded SubjectCombination
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Value() != SubjectCombinationCollective.Value() {
		t.Errorf("round trip Value() = %v, want %v", decoded.Value(), SubjectCombinationCollective.Value())
	}
}

func TestSubjectCombinationUnmarshalJSONFailurePreservesReceiver(t *testing.T) {
	original := SubjectCombinationIndependent
	receiver := original
	if err := json.Unmarshal([]byte(`123`), &receiver); err == nil {
		t.Fatal("malformed SubjectCombination JSON accepted, want error")
	}
	if receiver.Value() != original.Value() {
		t.Errorf("failed Unmarshal changed Value(): got %v, want %v", receiver.Value(), original.Value())
	}
}

func TestSubjectCombinationZeroValue(t *testing.T) {
	var sc SubjectCombination
	if !sc.IsZero() {
		t.Error("zero-value SubjectCombination.IsZero() = false, want true")
	}
}
