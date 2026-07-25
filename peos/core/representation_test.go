package core

import (
	"encoding/json"
	"errors"
	"testing"
)

func mustMediaType(t *testing.T) VocabularyValue {
	t.Helper()
	return mustVocabularyValue(t, "mime", "text/markdown")
}

// --- RepresentationContent --------------------------------------------------

func TestRepresentationContentInlineBytes(t *testing.T) {
	c, err := NewRepresentationContentFromInlineBytes([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Kind() != RepresentationContentKindInlineBytes {
		t.Errorf("Kind() = %q, want %q", c.Kind(), RepresentationContentKindInlineBytes)
	}
	got, ok := c.AsInlineBytes()
	if !ok || string(got) != "hello" {
		t.Errorf("AsInlineBytes() = (%s, %v)", got, ok)
	}
	if _, ok := c.AsInlineText(); ok {
		t.Error("AsInlineText() ok=true for inline_bytes content")
	}
}

func TestRepresentationContentInlineBytesDefensiveCopy(t *testing.T) {
	input := []byte("hello")
	c, err := NewRepresentationContentFromInlineBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 'X'
	got, _ := c.AsInlineBytes()
	if string(got) != "hello" {
		t.Errorf("stored content affected by caller mutation: got %s", got)
	}
	got[0] = 'Y'
	again, _ := c.AsInlineBytes()
	if string(again) != "hello" {
		t.Errorf("internal state affected by mutating an AsInlineBytes() result: got %s", again)
	}
}

func TestRepresentationContentInlineBytesEmptyRejected(t *testing.T) {
	if _, err := NewRepresentationContentFromInlineBytes(nil); !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
	if _, err := NewRepresentationContentFromInlineBytes([]byte{}); !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
}

func TestRepresentationContentInlineText(t *testing.T) {
	c, err := NewRepresentationContentFromInlineText("hello")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := c.AsInlineText()
	if !ok || got != "hello" {
		t.Errorf("AsInlineText() = (%q, %v)", got, ok)
	}
}

func TestRepresentationContentInlineTextEmptyRejected(t *testing.T) {
	if _, err := NewRepresentationContentFromInlineText(""); !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
}

func TestRepresentationContentExternalReference(t *testing.T) {
	c, err := NewRepresentationContentFromExternalReference("https://example.com/artifact/42?rev=7")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := c.AsExternalReference()
	if !ok || got != "https://example.com/artifact/42?rev=7" {
		t.Errorf("AsExternalReference() = (%q, %v)", got, ok)
	}
}

func TestRepresentationContentExternalReferenceEmptyRejected(t *testing.T) {
	if _, err := NewRepresentationContentFromExternalReference(""); !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
}

func TestRepresentationContentContentAddress(t *testing.T) {
	algo := mustVocabularyValue(t, "peos", "sha256")
	c, err := NewRepresentationContentFromContentAddress(algo, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	gotAlgo, gotDigest, ok := c.AsContentAddress()
	if !ok || gotAlgo != algo || gotDigest != "abc123" {
		t.Errorf("AsContentAddress() = (%v, %q, %v)", gotAlgo, gotDigest, ok)
	}
}

func TestRepresentationContentContentAddressZeroAlgorithmRejected(t *testing.T) {
	if _, err := NewRepresentationContentFromContentAddress(VocabularyValue{}, "abc123"); !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
}

func TestRepresentationContentContentAddressEmptyDigestRejected(t *testing.T) {
	algo := mustVocabularyValue(t, "peos", "sha256")
	if _, err := NewRepresentationContentFromContentAddress(algo, ""); !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
}

func TestRepresentationContentJSONRoundTripAllKinds(t *testing.T) {
	algo := mustVocabularyValue(t, "peos", "sha256")
	tests := []struct {
		name string
		make func() (RepresentationContent, error)
	}{
		{"inline_bytes", func() (RepresentationContent, error) { return NewRepresentationContentFromInlineBytes([]byte("hello")) }},
		{"inline_text", func() (RepresentationContent, error) { return NewRepresentationContentFromInlineText("hello") }},
		{"external_reference", func() (RepresentationContent, error) {
			return NewRepresentationContentFromExternalReference("https://example.com/x")
		}},
		{"content_address", func() (RepresentationContent, error) {
			return NewRepresentationContentFromContentAddress(algo, "abc123")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original, err := tt.make()
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decoded RepresentationContent
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Kind() != original.Kind() {
				t.Errorf("round trip Kind() = %q, want %q", decoded.Kind(), original.Kind())
			}
			data2, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != string(data2) {
				t.Errorf("round trip byte mismatch: got %s, want %s", data2, data)
			}
		})
	}
}

func TestRepresentationContentMalformedDiscriminator(t *testing.T) {
	var c RepresentationContent
	if err := json.Unmarshal([]byte(`{"kind":"","ref":""}`), &c); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("empty kind: error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
	if err := json.Unmarshal([]byte(`{"kind":"not_a_real_kind","ref":""}`), &c); !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("unrecognized kind: error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
}

func TestRepresentationContentPayloadMismatch(t *testing.T) {
	var c RepresentationContent
	err := json.Unmarshal([]byte(`{"kind":"content_address","ref":"not-an-object"}`), &c)
	if err == nil {
		t.Error("payload mismatch accepted, want error")
	}
}

func TestRepresentationContentZeroValue(t *testing.T) {
	var c RepresentationContent
	if !c.IsZero() {
		t.Error("zero-value RepresentationContent.IsZero() = false, want true")
	}
}

// --- Representation ----------------------------------------------------------

func TestNewRepresentationFromInlineBytes(t *testing.T) {
	rep, err := NewRepresentationFromInlineBytes([]byte("hello"), mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	if rep.IsZero() {
		t.Error("valid Representation reports IsZero() = true")
	}
	got, ok := rep.Content().AsInlineBytes()
	if !ok || string(got) != "hello" {
		t.Errorf("Content().AsInlineBytes() = (%s, %v)", got, ok)
	}
}

func TestNewRepresentationFromInlineText(t *testing.T) {
	rep, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := rep.Content().AsInlineText()
	if !ok || got != "hello" {
		t.Errorf("Content().AsInlineText() = (%q, %v)", got, ok)
	}
}

func TestNewRepresentationFromExternalReference(t *testing.T) {
	rep, err := NewRepresentationFromExternalReference("https://example.com/x", mustMediaType(t), RepresentationRoleDerived)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := rep.Content().AsExternalReference()
	if !ok || got != "https://example.com/x" {
		t.Errorf("Content().AsExternalReference() = (%q, %v)", got, ok)
	}
}

func TestNewRepresentationFromContentAddress(t *testing.T) {
	algo := mustVocabularyValue(t, "peos", "sha256")
	rep, err := NewRepresentationFromContentAddress(algo, "abc123", mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	gotAlgo, gotDigest, ok := rep.Content().AsContentAddress()
	if !ok || gotAlgo != algo || gotDigest != "abc123" {
		t.Errorf("Content().AsContentAddress() = (%v, %q, %v)", gotAlgo, gotDigest, ok)
	}
}

func TestNewRepresentationRequiresMediaType(t *testing.T) {
	if _, err := NewRepresentationFromInlineText("hello", VocabularyValue{}, RepresentationRoleAuthoritative); !errors.Is(err, ErrInvalidRepresentation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRepresentation)
	}
}

func TestNewRepresentationRequiresAtLeastOneClassification(t *testing.T) {
	if _, err := NewRepresentationFromInlineText("hello", mustMediaType(t)); !errors.Is(err, ErrInvalidRepresentation) {
		t.Errorf("error = %v, want %v", err, ErrInvalidRepresentation)
	}
}

func TestRepresentationClassificationCombinable(t *testing.T) {
	// PEOS-002 does not state that authoritative/derived/partial/rendered
	// are mutually exclusive; a Representation MAY combine partial and
	// rendered, for example.
	rep, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRolePartial, RepresentationRoleRendered)
	if err != nil {
		t.Fatal(err)
	}
	got := rep.Classification()
	if len(got) != 2 || got[0] != RepresentationRolePartial || got[1] != RepresentationRoleRendered {
		t.Errorf("Classification() = %v, want [%v %v] (order preserved)", got, RepresentationRolePartial, RepresentationRoleRendered)
	}
}

func TestRepresentationClassificationDuplicateRejected(t *testing.T) {
	if _, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative, RepresentationRoleAuthoritative); !errors.Is(err, ErrDuplicateRepresentationRole) {
		t.Errorf("error = %v, want %v", err, ErrDuplicateRepresentationRole)
	}
}

func TestRepresentationClassificationDefensiveCopy(t *testing.T) {
	rep, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	got := rep.Classification()
	got[0] = RepresentationRoleDerived
	again := rep.Classification()
	if again[0] != RepresentationRoleAuthoritative {
		t.Errorf("mutating a Classification() result affected internal state: got %v", again[0])
	}
}

func TestRepresentationWithLanguageAndTransformation(t *testing.T) {
	rep, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := rep.Language(); ok {
		t.Error("Language() ok = true before WithLanguage was called")
	}
	lang := mustVocabularyValue(t, "peos", "en")
	rep = rep.WithLanguage(lang)
	if got, ok := rep.Language(); !ok || got != lang {
		t.Errorf("Language() = (%v, %v), want (%v, true)", got, ok, lang)
	}

	transform := mustVocabularyValue(t, "peos", "render-to-html")
	rep = rep.WithTransformation(transform)
	if got, ok := rep.Transformation(); !ok || got != transform {
		t.Errorf("Transformation() = (%v, %v), want (%v, true)", got, ok, transform)
	}
}

func TestRepresentationWithExtension(t *testing.T) {
	rep, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	rep = rep.WithExtension(ext)
	got, ok := rep.Extension().Get("product-x")
	if !ok || string(got) != `{"a":1}` {
		t.Errorf("Extension().Get(\"product-x\") = (%s, %v)", got, ok)
	}
}

func TestRepresentationJSONRoundTrip(t *testing.T) {
	lang := mustVocabularyValue(t, "peos", "en")
	ext, err := NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative, RepresentationRoleRendered)
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithLanguage(lang).WithExtension(ext)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Representation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	gotText, ok := decoded.Content().AsInlineText()
	if !ok || gotText != "hello" {
		t.Errorf("round trip Content() = (%q, %v)", gotText, ok)
	}
	if decoded.MediaType() != original.MediaType() {
		t.Errorf("round trip MediaType() = %v, want %v", decoded.MediaType(), original.MediaType())
	}
	gotClassification := decoded.Classification()
	if len(gotClassification) != 2 || gotClassification[0] != RepresentationRoleAuthoritative || gotClassification[1] != RepresentationRoleRendered {
		t.Errorf("round trip Classification() = %v", gotClassification)
	}
	if gotLang, ok := decoded.Language(); !ok || gotLang != lang {
		t.Errorf("round trip Language() = (%v, %v), want (%v, true)", gotLang, ok, lang)
	}
}

func TestRepresentationJSONMalformedDiscriminator(t *testing.T) {
	var rep Representation
	err := json.Unmarshal([]byte(`{"content":{"kind":"not_a_real_kind","ref":""},"media_type":"mime:text/markdown","classification":["peos:authoritative"]}`), &rep)
	if !errors.Is(err, ErrInvalidReferenceDiscriminator) {
		t.Errorf("error = %v, want %v", err, ErrInvalidReferenceDiscriminator)
	}
}

func TestRepresentationJSONPayloadMismatch(t *testing.T) {
	var rep Representation
	err := json.Unmarshal([]byte(`{"content":{"kind":"inline_text","ref":""},"media_type":"mime:text/markdown","classification":["peos:authoritative"]}`), &rep)
	if !errors.Is(err, ErrMissingRepresentationContent) {
		t.Errorf("error = %v, want %v", err, ErrMissingRepresentationContent)
	}
}

func TestRepresentationJSONHasNoIdentityOrOwnershipFields(t *testing.T) {
	original, err := NewRepresentationFromInlineText("hello", mustMediaType(t), RepresentationRoleAuthoritative)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"artifact_id", "revision_id", "artifact_revision_id", "revision", "local_key", "id", "status", "lifecycle_state",
	} {
		if _, present := raw[forbidden]; present {
			t.Errorf("forbidden field %q present in Representation JSON", forbidden)
		}
	}
}

func TestRepresentationZeroValue(t *testing.T) {
	var rep Representation
	if !rep.IsZero() {
		t.Error("zero-value Representation.IsZero() = false, want true")
	}
}
