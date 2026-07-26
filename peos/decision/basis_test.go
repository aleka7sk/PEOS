package decision

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aleka7sk/PEOS/peos/core"
)

func mustEvidenceRef(t *testing.T, artifactID, revisionID string) core.EvidenceArtifactRevisionRef {
	t.Helper()
	aid, err := core.NewArtifactID(artifactID)
	if err != nil {
		t.Fatal(err)
	}
	rid, err := core.NewArtifactRevisionID(revisionID)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewEvidenceArtifactRevisionRef(aid, rid)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func mustAssumption(t *testing.T, statement string) Assumption {
	t.Helper()
	a, err := NewAssumption(statement)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustConstraint(t *testing.T, statement string) Constraint {
	t.Helper()
	c, err := NewConstraint(statement)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func mustUncertainty(t *testing.T, statement string) Uncertainty {
	t.Helper()
	u, err := NewUncertainty(statement)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// --- NewBasis (existing behavior unchanged) -------------------------------

func TestNewBasisValidEvidence(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Evidence()) != 1 {
		t.Errorf("Evidence() len = %d, want 1", len(b.Evidence()))
	}
}

func TestNewBasisEmptyEvidenceRejected(t *testing.T) {
	if _, err := NewBasis(nil); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := NewBasis([]core.EvidenceArtifactRevisionRef{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("empty slice: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewBasisZeroEvidenceRefRejected(t *testing.T) {
	if _, err := NewBasis([]core.EvidenceArtifactRevisionRef{{}}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisDefensiveCopies(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	input := []core.EvidenceArtifactRevisionRef{ev}
	b, err := NewBasis(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = core.EvidenceArtifactRevisionRef{}
	if b.Evidence()[0].IsZero() {
		t.Error("NewBasis did not defensively copy input")
	}
	got := b.Evidence()
	got[0] = core.EvidenceArtifactRevisionRef{}
	if b.Evidence()[0].IsZero() {
		t.Error("Evidence() did not defensively copy on return")
	}
}

func TestBasisExtensionBehavior(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	withExt := b.WithExtension(ext)
	if !b.Extension().IsZero() {
		t.Error("WithExtension mutated the original receiver")
	}
	if withExt.Extension().IsZero() {
		t.Error("WithExtension did not set extension")
	}
}

func TestBasisJSONRoundTrip(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Basis
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Evidence()) != 1 {
		t.Errorf("round trip mismatch: got %+v", decoded)
	}
}

func TestBasisExplicitNullRejected(t *testing.T) {
	var b Basis
	if err := json.Unmarshal([]byte(`{"evidence":null}`), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("null evidence: error = %v, want %v", err, ErrInvalidBasis)
	}
	valid := `{"evidence":[{"artifact_id":"ART-1","revision_id":"REV-1"}],"extension":null}`
	if err := json.Unmarshal([]byte(valid), &b); err == nil {
		t.Error("null extension accepted, want error")
	}
	if err := json.Unmarshal([]byte(`{"assumptions":null}`), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("null assumptions: error = %v, want %v", err, ErrInvalidBasis)
	}
	if err := json.Unmarshal([]byte(`{"constraints":null}`), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("null constraints: error = %v, want %v", err, ErrInvalidBasis)
	}
	if err := json.Unmarshal([]byte(`{"uncertainties":null}`), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("null uncertainties: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisZeroMarshalRejected(t *testing.T) {
	var b Basis
	if _, err := json.Marshal(b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisUnmarshalFailurePreservesReceiver(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	original, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if len(receiver.Evidence()) != 1 {
		t.Error("failed Unmarshal changed receiver")
	}
}

// --- NewBasisFrom ----------------------------------------------------------

func TestNewBasisFromEvidenceOnly(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasisFrom([]core.EvidenceArtifactRevisionRef{ev}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Evidence()) != 1 {
		t.Error("evidence not preserved")
	}
}

func TestNewBasisFromAssumptionOnly(t *testing.T) {
	b, err := NewBasisFrom(nil, []Assumption{mustAssumption(t, "traffic stays low")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Assumptions()) != 1 {
		t.Error("assumptions not preserved")
	}
	if b.IsZero() {
		t.Error("assumption-only Basis reports IsZero() = true")
	}
}

func TestNewBasisFromConstraintOnly(t *testing.T) {
	b, err := NewBasisFrom(nil, nil, []Constraint{mustConstraint(t, "must remain EU-resident")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Constraints()) != 1 {
		t.Error("constraints not preserved")
	}
	if b.IsZero() {
		t.Error("constraint-only Basis reports IsZero() = true")
	}
}

func TestNewBasisFromUncertaintyOnly(t *testing.T) {
	b, err := NewBasisFrom(nil, nil, nil, []Uncertainty{mustUncertainty(t, "vendor pricing unknown")})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Uncertainties()) != 1 {
		t.Error("uncertainties not preserved")
	}
	if b.IsZero() {
		t.Error("uncertainty-only Basis reports IsZero() = true")
	}
}

func TestNewBasisFromMixed(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasisFrom(
		[]core.EvidenceArtifactRevisionRef{ev},
		[]Assumption{mustAssumption(t, "a")},
		[]Constraint{mustConstraint(t, "c")},
		[]Uncertainty{mustUncertainty(t, "u")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Evidence()) != 1 || len(b.Assumptions()) != 1 || len(b.Constraints()) != 1 || len(b.Uncertainties()) != 1 {
		t.Errorf("mixed Basis did not preserve all four collections: %+v", b)
	}
}

func TestNewBasisFromAllEmptyRejected(t *testing.T) {
	if _, err := NewBasisFrom(nil, nil, nil, nil); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := NewBasisFrom(
		[]core.EvidenceArtifactRevisionRef{}, []Assumption{}, []Constraint{}, []Uncertainty{},
	); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("all-empty slices: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewBasisFromZeroElementRejectedPerCollection(t *testing.T) {
	if _, err := NewBasisFrom([]core.EvidenceArtifactRevisionRef{{}}, nil, nil, nil); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("zero evidence: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := NewBasisFrom(nil, []Assumption{{}}, nil, nil); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("zero assumption: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := NewBasisFrom(nil, nil, []Constraint{{}}, nil); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("zero constraint: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := NewBasisFrom(nil, nil, nil, []Uncertainty{{}}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("zero uncertainty: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestNewBasisFromDefensiveInputCopies(t *testing.T) {
	assumptions := []Assumption{mustAssumption(t, "a")}
	constraints := []Constraint{mustConstraint(t, "c")}
	uncertainties := []Uncertainty{mustUncertainty(t, "u")}
	ev := []core.EvidenceArtifactRevisionRef{mustEvidenceRef(t, "ART-1", "REV-1")}

	b, err := NewBasisFrom(ev, assumptions, constraints, uncertainties)
	if err != nil {
		t.Fatal(err)
	}
	ev[0] = core.EvidenceArtifactRevisionRef{}
	assumptions[0] = Assumption{}
	constraints[0] = Constraint{}
	uncertainties[0] = Uncertainty{}

	if b.Evidence()[0].IsZero() || b.Assumptions()[0].IsZero() || b.Constraints()[0].IsZero() || b.Uncertainties()[0].IsZero() {
		t.Error("NewBasisFrom did not defensively copy input slices")
	}
}

// --- Basis mutators (R1) ----------------------------------------------------

func fullMixedBasis(t *testing.T) Basis {
	t.Helper()
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasisFrom(
		[]core.EvidenceArtifactRevisionRef{ev},
		[]Assumption{mustAssumption(t, "a")},
		[]Constraint{mustConstraint(t, "c")},
		[]Uncertainty{mustUncertainty(t, "u")},
	)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBasisWithAssumptionsReplace(t *testing.T) {
	b := fullMixedBasis(t)
	other := mustAssumption(t, "other assumption")
	updated, err := b.WithAssumptions(other)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Assumptions()) != 1 || updated.Assumptions()[0].Statement() != "other assumption" {
		t.Errorf("WithAssumptions did not replace: %+v", updated.Assumptions())
	}
}

func TestBasisClearAssumptionsWhileEvidenceRemains(t *testing.T) {
	b := fullMixedBasis(t)
	// isolate to evidence + assumptions only, to test the specific pairing
	b, err := b.WithConstraints()
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithUncertainties()
	if err != nil {
		t.Fatal(err)
	}
	updated, err := b.WithAssumptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Assumptions()) != 0 {
		t.Error("assumptions not cleared")
	}
	if updated.IsZero() {
		t.Error("clearing assumptions while evidence remains produced IsZero()==true")
	}
}

func TestBasisClearAssumptionsWhileConstraintsRemain(t *testing.T) {
	b := fullMixedBasis(t)
	b, err := b.WithEvidence()
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.WithUncertainties()
	if err != nil {
		t.Fatal(err)
	}
	updated, err := b.WithAssumptions()
	if err != nil {
		t.Fatal(err)
	}
	if updated.IsZero() {
		t.Error("clearing assumptions while constraints remain produced IsZero()==true")
	}
}

func TestBasisClearAssumptionsWhenLastRejected(t *testing.T) {
	b, err := NewBasisFrom(nil, []Assumption{mustAssumption(t, "only content")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	receiver := b
	if _, err := b.WithAssumptions(); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
	if len(receiver.Assumptions()) != 1 {
		t.Error("failed WithAssumptions mutated receiver")
	}
}

func TestBasisClearEvidenceWhileUncertaintyRemains(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasisFrom([]core.EvidenceArtifactRevisionRef{ev}, nil, nil, []Uncertainty{mustUncertainty(t, "u")})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := b.WithEvidence()
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Evidence()) != 0 {
		t.Error("evidence not cleared")
	}
	if updated.IsZero() {
		t.Error("clearing evidence while uncertainty remains produced IsZero()==true")
	}
}

func TestBasisClearEvidenceWhenLastRejected(t *testing.T) {
	ev := mustEvidenceRef(t, "ART-1", "REV-1")
	b, err := NewBasis([]core.EvidenceArtifactRevisionRef{ev})
	if err != nil {
		t.Fatal(err)
	}
	receiver := b
	if _, err := b.WithEvidence(); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
	if len(receiver.Evidence()) != 1 {
		t.Error("failed WithEvidence mutated receiver")
	}
}

func TestBasisClearConstraintsWhenLastRejected(t *testing.T) {
	b, err := NewBasisFrom(nil, nil, []Constraint{mustConstraint(t, "only content")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	receiver := b
	if _, err := b.WithConstraints(); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
	if len(receiver.Constraints()) != 1 {
		t.Error("failed WithConstraints mutated receiver")
	}
}

func TestBasisClearUncertaintiesWhenLastRejected(t *testing.T) {
	b, err := NewBasisFrom(nil, nil, nil, []Uncertainty{mustUncertainty(t, "only content")})
	if err != nil {
		t.Fatal(err)
	}
	receiver := b
	if _, err := b.WithUncertainties(); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
	if len(receiver.Uncertainties()) != 1 {
		t.Error("failed WithUncertainties mutated receiver")
	}
}

func TestBasisMutatorZeroElementRejected(t *testing.T) {
	b := fullMixedBasis(t)
	if _, err := b.WithEvidence(core.EvidenceArtifactRevisionRef{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithEvidence zero element: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := b.WithAssumptions(Assumption{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithAssumptions zero element: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := b.WithConstraints(Constraint{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithConstraints zero element: error = %v, want %v", err, ErrInvalidBasis)
	}
	if _, err := b.WithUncertainties(Uncertainty{}); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("WithUncertainties zero element: error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisMutatorFailurePreservesReceiverAndUnrelatedFields(t *testing.T) {
	b := fullMixedBasis(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	b = b.WithExtension(ext)

	if _, err := b.WithAssumptions(Assumption{}); !errors.Is(err, ErrInvalidBasis) {
		t.Fatalf("expected error, got %v", err)
	}
	// b (the original receiver variable) must be untouched: value semantics
	// already guarantee this, but assert observable state explicitly.
	if len(b.Evidence()) != 1 || len(b.Assumptions()) != 1 || len(b.Constraints()) != 1 || len(b.Uncertainties()) != 1 {
		t.Error("failed mutator changed receiver's collections")
	}
	if b.Extension().IsZero() {
		t.Error("failed mutator changed receiver's extension")
	}
}

func TestBasisMutatorInputSliceCopied(t *testing.T) {
	b := fullMixedBasis(t)
	assumptions := []Assumption{mustAssumption(t, "fresh")}
	updated, err := b.WithAssumptions(assumptions...)
	if err != nil {
		t.Fatal(err)
	}
	assumptions[0] = Assumption{}
	if updated.Assumptions()[0].IsZero() {
		t.Error("WithAssumptions did not defensively copy input")
	}
}

// --- Basis accessors ---------------------------------------------------

func TestBasisAccessorsDefensiveCopies(t *testing.T) {
	b := fullMixedBasis(t)

	ev := b.Evidence()
	ev[0] = core.EvidenceArtifactRevisionRef{}
	if b.Evidence()[0].IsZero() {
		t.Error("Evidence() did not defensively copy on return")
	}

	as := b.Assumptions()
	as[0] = Assumption{}
	if b.Assumptions()[0].IsZero() {
		t.Error("Assumptions() did not defensively copy on return")
	}

	cs := b.Constraints()
	cs[0] = Constraint{}
	if b.Constraints()[0].IsZero() {
		t.Error("Constraints() did not defensively copy on return")
	}

	us := b.Uncertainties()
	us[0] = Uncertainty{}
	if b.Uncertainties()[0].IsZero() {
		t.Error("Uncertainties() did not defensively copy on return")
	}
}

// --- IsZero semantics + Decision integration --------------------------

func TestBasisIsZeroPerCategory(t *testing.T) {
	var zero Basis
	if !zero.IsZero() {
		t.Error("zero Basis IsZero() = false")
	}

	evOnly, err := NewBasis([]core.EvidenceArtifactRevisionRef{mustEvidenceRef(t, "ART-1", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	if evOnly.IsZero() {
		t.Error("evidence-only Basis IsZero() = true")
	}

	asOnly, err := NewBasisFrom(nil, []Assumption{mustAssumption(t, "a")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if asOnly.IsZero() {
		t.Error("assumption-only Basis IsZero() = true")
	}

	csOnly, err := NewBasisFrom(nil, nil, []Constraint{mustConstraint(t, "c")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if csOnly.IsZero() {
		t.Error("constraint-only Basis IsZero() = true")
	}

	usOnly, err := NewBasisFrom(nil, nil, nil, []Uncertainty{mustUncertainty(t, "u")})
	if err != nil {
		t.Fatal(err)
	}
	if usOnly.IsZero() {
		t.Error("uncertainty-only Basis IsZero() = true")
	}
}

// TestBasisExtensionOnlyZeroReceiverEdge documents and locks in the
// pre-existing Packet F edge described on Basis.WithExtension: calling
// WithExtension on the zero Basis does not create valid content. This is
// intentional, not a bug -- see F2-09 / WithExtension's own doc comment.
func TestBasisExtensionOnlyZeroReceiverEdge(t *testing.T) {
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	extOnly := Basis{}.WithExtension(ext)
	if !extOnly.IsZero() {
		t.Error("extension-only Basis (via zero-receiver WithExtension) IsZero() = false, want true")
	}
	if _, err := json.Marshal(extOnly); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("Marshal error = %v, want %v", err, ErrInvalidBasis)
	}

	d := baseDecision(t)
	if _, err := d.WithBasis(extOnly); err == nil {
		t.Error("Decision.WithBasis accepted an extension-only Basis, want error")
	}
}

func TestDecisionWithBasisAcceptsEachSingleCategoryBasis(t *testing.T) {
	d := baseDecision(t)

	evOnly, err := NewBasis([]core.EvidenceArtifactRevisionRef{mustEvidenceRef(t, "ART-1", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.WithBasis(evOnly); err != nil {
		t.Errorf("evidence-only Basis rejected by Decision.WithBasis: %v", err)
	}

	asOnly, err := NewBasisFrom(nil, []Assumption{mustAssumption(t, "a")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	withAs, err := d.WithBasis(asOnly)
	if err != nil {
		t.Fatalf("assumption-only Basis rejected by Decision.WithBasis: %v", err)
	}
	gotBasis, ok := withAs.Basis()
	if !ok {
		t.Fatal("Decision.Basis() ok = false after WithBasis(assumption-only)")
	}
	if len(gotBasis.Assumptions()) != 1 {
		t.Error("assumption-only Basis not preserved through Decision.WithBasis/Basis()")
	}

	csOnly, err := NewBasisFrom(nil, nil, []Constraint{mustConstraint(t, "c")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.WithBasis(csOnly); err != nil {
		t.Errorf("constraint-only Basis rejected by Decision.WithBasis: %v", err)
	}

	usOnly, err := NewBasisFrom(nil, nil, nil, []Uncertainty{mustUncertainty(t, "u")})
	if err != nil {
		t.Fatal(err)
	}
	withUs, err := d.WithBasis(usOnly)
	if err != nil {
		t.Fatalf("uncertainty-only Basis rejected by Decision.WithBasis: %v", err)
	}
	gotBasis2, ok := withUs.Basis()
	if !ok {
		t.Fatal("Decision.Basis() ok = false after WithBasis(uncertainty-only)")
	}
	if len(gotBasis2.Uncertainties()) != 1 {
		t.Error("uncertainty-only Basis not preserved through Decision.WithBasis/Basis()")
	}
}

func TestDecisionJSONRoundTripPreservesEachSingleCategoryBasis(t *testing.T) {
	d := baseDecision(t)

	asOnly, err := NewBasisFrom(nil, []Assumption{mustAssumption(t, "assumption-only content")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	withAs, err := d.WithBasis(asOnly)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(withAs)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Decision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	decodedBasis, ok := decoded.Basis()
	if !ok {
		t.Fatal("assumption-only Basis was silently omitted from Decision JSON round trip")
	}
	if len(decodedBasis.Assumptions()) != 1 || decodedBasis.Assumptions()[0].Statement() != "assumption-only content" {
		t.Errorf("assumption-only Basis content lost in round trip: %+v", decodedBasis.Assumptions())
	}

	usOnly, err := NewBasisFrom(nil, nil, nil, []Uncertainty{mustUncertainty(t, "uncertainty-only content")})
	if err != nil {
		t.Fatal(err)
	}
	withUs, err := d.WithBasis(usOnly)
	if err != nil {
		t.Fatal(err)
	}
	data2, err := json.Marshal(withUs)
	if err != nil {
		t.Fatal(err)
	}
	var decoded2 Decision
	if err := json.Unmarshal(data2, &decoded2); err != nil {
		t.Fatal(err)
	}
	decodedBasis2, ok := decoded2.Basis()
	if !ok {
		t.Fatal("uncertainty-only Basis was silently omitted from Decision JSON round trip")
	}
	if len(decodedBasis2.Uncertainties()) != 1 {
		t.Error("uncertainty-only Basis content lost in round trip")
	}
}

// --- Basis JSON: minimum per category, mixed, and error paths ----------

func TestBasisJSONMinimumPerCategory(t *testing.T) {
	cases := []string{
		`{"evidence":[{"artifact_id":"ART-1","revision_id":"REV-1"}]}`,
		`{"assumptions":[{"statement":"a"}]}`,
		`{"constraints":[{"statement":"c"}]}`,
		`{"uncertainties":[{"statement":"u"}]}`,
	}
	for _, payload := range cases {
		var b Basis
		if err := json.Unmarshal([]byte(payload), &b); err != nil {
			t.Errorf("payload %s: unexpected error %v", payload, err)
		}
		if b.IsZero() {
			t.Errorf("payload %s: decoded Basis IsZero() = true", payload)
		}
	}
}

func TestBasisJSONMixed(t *testing.T) {
	payload := `{
		"evidence": [{"artifact_id":"ART-1","revision_id":"REV-1"}],
		"assumptions": [{"statement":"a"}],
		"constraints": [{"statement":"c"}],
		"uncertainties": [{"statement":"u"}]
	}`
	var b Basis
	if err := json.Unmarshal([]byte(payload), &b); err != nil {
		t.Fatal(err)
	}
	if len(b.Evidence()) != 1 || len(b.Assumptions()) != 1 || len(b.Constraints()) != 1 || len(b.Uncertainties()) != 1 {
		t.Errorf("mixed JSON did not populate all four collections: %+v", b)
	}
}

func TestBasisJSONAllAbsentRejected(t *testing.T) {
	var b Basis
	if err := json.Unmarshal([]byte(`{}`), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisJSONAllEmptyArraysRejected(t *testing.T) {
	var b Basis
	payload := `{"evidence":[],"assumptions":[],"constraints":[],"uncertainties":[]}`
	if err := json.Unmarshal([]byte(payload), &b); !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBasis)
	}
}

func TestBasisJSONZeroElementsRejected(t *testing.T) {
	cases := []string{
		`{"evidence":[{"artifact_id":"","revision_id":""}]}`,
		`{"assumptions":[{"statement":""}]}`,
		`{"constraints":[{"statement":""}]}`,
		`{"uncertainties":[{"statement":""}]}`,
	}
	for _, payload := range cases {
		var b Basis
		if err := json.Unmarshal([]byte(payload), &b); err == nil {
			t.Errorf("payload %s: accepted, want error", payload)
		}
	}
}

func TestBasisJSONUnknownFieldIgnored(t *testing.T) {
	var b Basis
	payload := `{"assumptions":[{"statement":"a"}],"unknown_field":123}`
	if err := json.Unmarshal([]byte(payload), &b); err != nil {
		t.Fatal(err)
	}
}

func TestBasisJSONRequiredLiteralKeysWhenPopulated(t *testing.T) {
	b := fullMixedBasis(t)
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"evidence", "assumptions", "constraints", "uncertainties"} {
		if _, present := raw[key]; !present {
			t.Errorf("populated key %q missing from wire form", key)
		}
	}
}

func TestBasisJSONOmitEmptyBehavior(t *testing.T) {
	evOnly, err := NewBasis([]core.EvidenceArtifactRevisionRef{mustEvidenceRef(t, "ART-1", "REV-1")})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(evOnly)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"assumptions", "constraints", "uncertainties", "extension"} {
		if _, present := raw[key]; present {
			t.Errorf("empty key %q present in wire form", key)
		}
	}
}

func TestBasisFullUnmarshalFailurePreservesReceiver(t *testing.T) {
	original := fullMixedBasis(t)
	ext, err := core.NewExtension().With("product-x", json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	original = original.WithExtension(ext)

	receiver := original
	if err := json.Unmarshal([]byte(`{}`), &receiver); err == nil {
		t.Fatal("empty object accepted, want error")
	}
	if len(receiver.Evidence()) != 1 || len(receiver.Assumptions()) != 1 ||
		len(receiver.Constraints()) != 1 || len(receiver.Uncertainties()) != 1 {
		t.Error("failed Unmarshal changed receiver's collections")
	}
	if receiver.Extension().IsZero() {
		t.Error("failed Unmarshal changed receiver's extension")
	}
}

func TestBasisNestedDecodeErrorsPreserveBothSentinels(t *testing.T) {
	var b Basis
	// Malformed nested evidence ref (missing artifact_id).
	payload := `{"evidence":[{"artifact_id":"","revision_id":"REV-1"}]}`
	err := json.Unmarshal([]byte(payload), &b)
	if err == nil {
		t.Fatal("malformed evidence accepted, want error")
	}
	if !errors.Is(err, ErrInvalidBasis) {
		t.Errorf("error = %v, want wrapping %v", err, ErrInvalidBasis)
	}
	if !errors.Is(err, core.ErrEmptyIdentity) {
		t.Errorf("error = %v, want also wrapping %v", err, core.ErrEmptyIdentity)
	}
}
