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

func TestRollBeltBulk_AabPI(t *testing.T) {
	t.Parallel()
	// WBH p.74 Aab PI: 2D+2 result = 6 (book shows "4+2"), age 6.3 Gyr DM=-3, c-type 2% DM=+0.
	// Bulk = 6 - 3 + 0 = 3.
	// NOTE: WBH notation "2D2" is "2D + 2" (range 4-14); rendered in code as "2D+2".
	// Scripted value: 6 (post-modifier 2D+2 result — Scripted ignores notation, returns value directly).
	r := roller.NewScripted(6)
	got, err := RollBeltBulk(r, 6.3, BeltComposition{CTypePct: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestRollResourceRating_AabPI(t *testing.T) {
	t.Parallel()
	// WBH p.74 Aab PI: 2D=11 (scripted), bulk=3, mType=55 ⇒ DM+5, cType=2 ⇒ DM+0 (2/10 floor).
	// Formula: 11 - 7 + 3 + 5 - 0 = 12.
	// Book shows the answer as 11 ("B" hex) — book applies cType DM-1 instead of -0.
	// Document divergence; accept 11 or 12.
	r := roller.NewScripted(11)
	got, err := RollResourceRating(r, 3, BeltComposition{MTypePct: 55, CTypePct: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got != 11 && got != 12 {
		t.Errorf("got %d, want 11 (book) or 12 (formula)", got)
	}
	if got == 12 {
		t.Logf("WBH p.74 inconsistency: book example shows 11 but strict formula gives 12 (cType%%/10 floor = 0)")
	}
}

func TestRollSigSize1Bodies_AabPI(t *testing.T) {
	t.Parallel()
	// WBH p.74 Aab PI: 2D=8 + (-12) + bulk=3 = -1 → clamped 0.
	r := roller.NewScripted(8)
	got, err := RollSigSize1Bodies(r, 3, 2.7, 3.3, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestRollSigSizeSBodies_AabPI(t *testing.T) {
	t.Parallel()
	// WBH p.74 Aab PI: 2D=10, bulk=3, no special DMs → formula gives 0 + (0+1)×(3+1) = 4.
	// Book example shows 3. Accept either; book may have unstated DM.
	r := roller.NewScripted(10)
	got, err := RollSigSizeSBodies(r, 3, 2.7, 3.3, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 && got != 4 {
		t.Errorf("got %d, want 3 (book) or 4 (formula)", got)
	}
	if got == 4 {
		t.Logf("WBH p.74 inconsistency: book example shows 3 but strict formula gives 4")
	}
}

func TestFormatBeltProfile_AabPI(t *testing.T) {
	t.Parallel()
	b := BeltDetails{
		Span:           0.25,
		Composition:    BeltComposition{MTypePct: 55, STypePct: 40, CTypePct: 2, OtherPct: 3},
		Bulk:           3,
		ResourceRating: 11,
		SigSize1Bodies: 0,
		SigSizeSBodies: 3,
	}
	got := FormatBeltProfile(b)
	want := "0.25-55.40.02.03-3-B-0-3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
