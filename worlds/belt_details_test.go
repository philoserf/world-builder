package worlds

import (
	"math"
	"testing"

	"wbh/roller"
)

func TestRollBeltSpan_BasicFormula(t *testing.T) {
	t.Parallel()
	// Spread = 0.5 (Zed system), 2D = 5 → 0.5 * 5 / 10 = 0.25
	r := roller.NewScripted(5)
	got, err := RollBeltSpan(r, 0.5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-0.25) > 1e-6 {
		t.Errorf("got %v, want 0.25", got)
	}
}

func TestRollBeltSpan_AppliesGGAdjacencyDM(t *testing.T) {
	t.Parallel()
	// Spread 0.5, 2D=6, DM-1 → effective 5 → 0.5 * 5 / 10 = 0.25
	r := roller.NewScripted(6)
	got, err := RollBeltSpan(r, 0.5, -1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-0.25) > 1e-6 {
		t.Errorf("with DM-1: got %v, want 0.25", got)
	}
}

func TestRollBeltComposition_AabPI(t *testing.T) {
	t.Parallel()
	// WBH p. 74 Aab PI: composition 6-4=2 (DM-4 inside HZCO):
	//   m-type 40+1D×5  → 1D=3 → 40+15 = 55
	//   s-type 15+1D×5  → 1D=5 → 15+25 = 40
	//   c-type 1D       → 1D=2 → 2
	//   other = 100 - 55 - 40 - 2 = 3
	// Scripted dice: 2D=6 (composition row), 1D=3 (m), 1D=5 (s), 1D=2 (c).
	r := roller.NewScripted(6, 3, 5, 2)
	got, err := RollBeltComposition(r, -4)
	if err != nil {
		t.Fatal(err)
	}
	if got.MTypePct != 55 {
		t.Errorf("m-type: got %d, want 55", got.MTypePct)
	}
	if got.STypePct != 40 {
		t.Errorf("s-type: got %d, want 40", got.STypePct)
	}
	if got.CTypePct != 2 {
		t.Errorf("c-type: got %d, want 2", got.CTypePct)
	}
	if got.OtherPct != 3 {
		t.Errorf("other: got %d, want 3", got.OtherPct)
	}
}
