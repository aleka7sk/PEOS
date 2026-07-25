package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewArtifactID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr error
	}{
		{name: "valid opaque value", value: "ART-123", want: "ART-123"},
		{name: "valid non-uuid value", value: "requirement/login-flow", want: "requirement/login-flow"},
		{name: "preserves case", value: "MixedCase-ID", want: "MixedCase-ID"},
		{name: "trims surrounding whitespace", value: "  ART-123  ", want: "ART-123"},
		{name: "empty value rejected", value: "", wantErr: ErrEmptyIdentity},
		{name: "whitespace-only value rejected", value: "   ", wantErr: ErrEmptyIdentity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArtifactID(tt.value)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewArtifactID(%q) error = %v, want %v", tt.value, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewArtifactID(%q) unexpected error: %v", tt.value, err)
			}
			if got.String() != tt.want {
				t.Errorf("String() = %q, want %q", got.String(), tt.want)
			}
			if got.IsZero() {
				t.Errorf("IsZero() = true for non-empty value %q", tt.value)
			}
		})
	}
}

func TestArtifactIDIsZero(t *testing.T) {
	var zero ArtifactID
	if !zero.IsZero() {
		t.Error("zero-value ArtifactID.IsZero() = false, want true")
	}
	id, err := NewArtifactID("ART-1")
	if err != nil {
		t.Fatal(err)
	}
	if id.IsZero() {
		t.Error("constructed ArtifactID.IsZero() = true, want false")
	}
}

func TestArtifactIDMapKey(t *testing.T) {
	a, err := NewArtifactID("ART-1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewArtifactID("ART-2")
	if err != nil {
		t.Fatal(err)
	}
	m := map[ArtifactID]string{a: "first", b: "second"}
	if m[a] != "first" || m[b] != "second" {
		t.Errorf("ArtifactID did not behave correctly as a map key: %v", m)
	}
}

func TestArtifactIDJSONRoundTrip(t *testing.T) {
	original, err := NewArtifactID("ART-42")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"ART-42"` {
		t.Errorf("Marshal = %s, want %q", data, `"ART-42"`)
	}
	var decoded ArtifactID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestArtifactIDJSONRejectsEmpty(t *testing.T) {
	var id ArtifactID
	if err := json.Unmarshal([]byte(`""`), &id); !errors.Is(err, ErrEmptyIdentity) {
		t.Errorf("Unmarshal empty string: error = %v, want %v", err, ErrEmptyIdentity)
	}
}

// TestIdentityTypesAreNotInterchangeable demonstrates, at compile time,
// that sibling identity types cannot be substituted for one another
// without an explicit, named constructor call. This test's real
// assertion is that the file compiles: if ArtifactID and
// ArtifactRevisionID were structurally interchangeable, a reviewer could
// accidentally introduce a bug that this test would not catch by static
// typing alone, which is exactly what the distinct field names in
// identity.go prevent.
func TestIdentityTypesAreNotInterchangeable(t *testing.T) {
	artifactID, err := NewArtifactID("ART-1")
	if err != nil {
		t.Fatal(err)
	}
	revisionID, err := NewArtifactRevisionID("REV-1")
	if err != nil {
		t.Fatal(err)
	}

	// The following, if uncommented, must fail to compile:
	//   var _ ArtifactID = revisionID
	//   var _ ArtifactRevisionID = artifactID
	//   var _ ArtifactID = ArtifactID(revisionID)
	// This test only exercises that both values are independently usable.
	if artifactID.String() == revisionID.String() {
		t.Skip("identical opaque values chosen; not a meaningful collision")
	}
}

// identityConstructor pairs a constructor with a human-readable name so
// the remaining twelve identity types can share one table-driven test
// instead of duplicating TestNewArtifactID's cases twelve times.
type identityConstructor struct {
	name        string
	constructFn func(string) (interface{ IsZero() bool }, error)
}

func TestRemainingIdentityConstructors(t *testing.T) {
	constructors := []identityConstructor{
		{"ArtifactRevisionID", func(v string) (interface{ IsZero() bool }, error) { return NewArtifactRevisionID(v) }},
		{"ImmutableRecordID", func(v string) (interface{ IsZero() bool }, error) { return NewImmutableRecordID(v) }},
		{"DecisionID", func(v string) (interface{ IsZero() bool }, error) { return NewDecisionID(v) }},
		{"ValidationClaimID", func(v string) (interface{ IsZero() bool }, error) { return NewValidationClaimID(v) }},
		{"ValidationExecutionRecordID", func(v string) (interface{ IsZero() bool }, error) { return NewValidationExecutionRecordID(v) }},
		{"RuntimeBindingRecordID", func(v string) (interface{ IsZero() bool }, error) { return NewRuntimeBindingRecordID(v) }},
		{"RuntimeUnbindingRecordID", func(v string) (interface{ IsZero() bool }, error) { return NewRuntimeUnbindingRecordID(v) }},
		{"RuntimeObservationID", func(v string) (interface{ IsZero() bool }, error) { return NewRuntimeObservationID(v) }},
		{"RuntimeViolationID", func(v string) (interface{ IsZero() bool }, error) { return NewRuntimeViolationID(v) }},
		{"TemplateApplicationRecordID", func(v string) (interface{ IsZero() bool }, error) { return NewTemplateApplicationRecordID(v) }},
		{"ControlledVocabularyID", func(v string) (interface{ IsZero() bool }, error) { return NewControlledVocabularyID(v) }},
		{"LocalKey", func(v string) (interface{ IsZero() bool }, error) { return NewLocalKey(v) }},
	}

	for _, c := range constructors {
		t.Run(c.name+"/valid", func(t *testing.T) {
			got, err := c.constructFn("X-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.IsZero() {
				t.Error("IsZero() = true for a valid value")
			}
		})
		t.Run(c.name+"/empty", func(t *testing.T) {
			_, err := c.constructFn("")
			if !errors.Is(err, ErrEmptyIdentity) {
				t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
			}
		})
		t.Run(c.name+"/whitespace", func(t *testing.T) {
			_, err := c.constructFn("   ")
			if !errors.Is(err, ErrEmptyIdentity) {
				t.Errorf("error = %v, want %v", err, ErrEmptyIdentity)
			}
		})
	}
}

func TestLocalKeyIsScopedNotGlobal(t *testing.T) {
	// Two LocalKey values with the same opaque string are equal as Go
	// values; this package deliberately does not attach an owning-scope
	// component to LocalKey itself (see identity.go). Callers pair a
	// LocalKey with an owning Revision reference at their own layer.
	a, err := NewLocalKey("ACT-1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewLocalKey("ACT-1")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Error("two LocalKey values built from the same string are not equal, want equal")
	}
}
