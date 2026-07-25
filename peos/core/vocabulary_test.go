package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewVocabularyValue(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		value         string
		wantErr       error
		wantNamespace string
		wantValue     string
	}{
		{name: "known peos value", namespace: "peos", value: "artifact", wantNamespace: "peos", wantValue: "artifact"},
		{name: "unknown product namespace preserved", namespace: "acme-corp", value: "custom-method", wantNamespace: "acme-corp", wantValue: "custom-method"},
		{name: "value may contain colons", namespace: "peos", value: "a:b:c", wantNamespace: "peos", wantValue: "a:b:c"},
		{name: "empty namespace rejected", namespace: "", value: "x", wantErr: ErrInvalidVocabularyValue},
		{name: "empty value rejected", namespace: "peos", value: "", wantErr: ErrInvalidVocabularyValue},
		{name: "namespace with colon rejected", namespace: "peos:sub", value: "x", wantErr: ErrInvalidVocabularyValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewVocabularyValue(tt.namespace, tt.value)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Namespace() != tt.wantNamespace || got.Value() != tt.wantValue {
				t.Errorf("got (%q, %q), want (%q, %q)", got.Namespace(), got.Value(), tt.wantNamespace, tt.wantValue)
			}
		})
	}
}

func TestVocabularyValueMalformedParse(t *testing.T) {
	tests := []string{"", "no-separator", ":missing-namespace", "missing-value:"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			_, err := ParseVocabularyValue(s)
			if err == nil && s != "missing-value:" {
				t.Fatalf("ParseVocabularyValue(%q) succeeded, want error", s)
			}
		})
	}
	// "missing-value:" parses to namespace "missing-value", empty value,
	// which NewVocabularyValue rejects because the value component is
	// empty.
	if _, err := ParseVocabularyValue("missing-value:"); !errors.Is(err, ErrInvalidVocabularyValue) {
		t.Errorf("ParseVocabularyValue(\"missing-value:\") error = %v, want %v", err, ErrInvalidVocabularyValue)
	}
}

func TestVocabularyValueRoundTrip(t *testing.T) {
	tests := []string{"peos:artifact", "acme-corp:custom-method", "peos:a:b:c"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			v, err := ParseVocabularyValue(s)
			if err != nil {
				t.Fatalf("ParseVocabularyValue(%q): %v", s, err)
			}
			if v.String() != s {
				t.Errorf("String() = %q, want %q", v.String(), s)
			}

			data, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var decoded VocabularyValue
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !decoded.Equal(v) {
				t.Errorf("round trip mismatch: got %v, want %v", decoded, v)
			}
		})
	}
}

func TestVocabularyValueUnknownValuesPreserved(t *testing.T) {
	v, err := NewVocabularyValue("product-x", "not-a-peos-value")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var decoded VocabularyValue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Namespace() != "product-x" || decoded.Value() != "not-a-peos-value" {
		t.Errorf("unknown namespace/value not preserved: got (%q, %q)", decoded.Namespace(), decoded.Value())
	}
}

func TestTypedVocabularyWrappersRoundTrip(t *testing.T) {
	if ClaimOutcomeSatisfied.String() != "peos:satisfied" {
		t.Errorf("ClaimOutcomeSatisfied.String() = %q, want %q", ClaimOutcomeSatisfied.String(), "peos:satisfied")
	}

	data, err := json.Marshal(ClaimTypeSatisfaction)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ClaimType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Value().Equal(ClaimTypeSatisfaction.Value()) {
		t.Errorf("round trip mismatch: got %v, want %v", decoded, ClaimTypeSatisfaction)
	}

	// A Product MAY extend a vocabulary family with a value outside the
	// predefined constants; the typed wrapper must accept it.
	custom := NewValidationMethod(mustVocabularyValue(t, "acme-corp", "manual-review"))
	if custom.IsZero() {
		t.Error("custom ValidationMethod unexpectedly zero")
	}
}

func mustVocabularyValue(t *testing.T, namespace, value string) VocabularyValue {
	t.Helper()
	v, err := NewVocabularyValue(namespace, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
