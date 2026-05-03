package worlds

import (
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func solSystem() stars.System {
	sol := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V, Mass: 1.000, Diameter: 1.000, Temperature: 5772,
	})
	return stars.System{Primary: sol}
}

func TestGenerateCounts_GasGiants_None(t *testing.T) {
	t.Parallel()
	// Existence roll = 10 (>9), so no gas giants. Belts existence = 7 (<8) so no belts.
	// Terrestrials: 2D=7 → 5 + DM+1 (single Class V) = 6 ≥3, +D3-1 with D3=2 → +1 → 7.
	r := roller.NewScripted(10 /*GG existence*/, 7 /*belts existence*/, 7 /*terrestrials 2D*/, 2 /*D3 add*/)
	got, err := GenerateCounts(r, solSystem(), CountsOpts{})
	if err != nil {
		t.Fatalf("GenerateCounts: %v", err)
	}
	if got.GasGiants != 0 {
		t.Errorf("GasGiants = %d, want 0", got.GasGiants)
	}
}

func TestGenerateCounts_GasGiants_PresentSingleClassV(t *testing.T) {
	t.Parallel()
	// Existence = 9 (≤9 → present). Quantity 2D=7, +DM+1 (single Class V) = 8 → row 7-8 → 3 GG.
	// Belts existence = 7 (no belts).
	// Terrestrials: 2D=7 → 5 + DM+1 = 6, +D3-1 (D3=1) = 6.
	r := roller.NewScripted(9, 7, 7 /*GG quantity*/, 7 /*belts existence*/, 7 /*terrestrials*/, 1)
	got, err := GenerateCounts(r, solSystem(), CountsOpts{})
	if err != nil {
		t.Fatalf("GenerateCounts: %v", err)
	}
	if got.GasGiants != 3 {
		t.Errorf("GasGiants = %d, want 3", got.GasGiants)
	}
}

func TestGenerateCounts_GasGiants_QuantityTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		roll int
		want int
	}{
		{3, 1}, // 4-: 1
		{4, 1},
		{5, 2}, // 5-6: 2
		{6, 2},
		{7, 3}, // 7-8: 3
		{8, 3},
		{9, 4}, // 9-11: 4
		{11, 4},
		{12, 5}, // 12: 5
		{13, 6}, // 13+: 6
	}
	for _, tc := range cases {
		// Class IV primary: no single-Class-V DM, no other DMs.
		sys := stars.System{Primary: stars.Compose(stars.ComposeOpts{
			Kind: stars.KindMainSequence, SpectralType: stars.SpectralType{Letter: 'G', Subtype: 2},
			LuminosityClass: stars.IV, Mass: 1.0, Diameter: 1.0, Temperature: 5772,
		})}
		r := roller.NewScripted(5, tc.roll, 7, 7, 1)
		got, err := GenerateCounts(r, sys, CountsOpts{})
		if err != nil {
			t.Fatalf("roll %d: GenerateCounts: %v", tc.roll, err)
		}
		if got.GasGiants != tc.want {
			t.Errorf("roll %d: GasGiants = %d, want %d", tc.roll, got.GasGiants, tc.want)
		}
	}
}

func TestGenerateCounts_GasGiants_DM_BrownDwarfPrimary(t *testing.T) {
	t.Parallel()
	// Brown-dwarf primary: existence raw 14 + DMs (-2 BD, -2 post-stellar, -1 per post-stellar = -5) = 9 → present.
	// Quantity: raw 14 + DMs -5 = 9 → row 9-11 → 4 GG.
	bd := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindBrownDwarf, LuminosityClass: stars.BD, Mass: 0.05, Diameter: 0.1, Temperature: 1500,
	})
	sys := stars.System{Primary: bd}
	// Belts existence = 0, terrestrials 2D=2 (force reroll path): D3+2 with D3=5? D3 max is 3 — implementation will be added in Task 5; this test only asserts GG count, so the trailing rolls just need to not panic. Provide enough.
	r := roller.NewScripted(14, 14, 0, 2, 5)
	got, err := GenerateCounts(r, sys, CountsOpts{})
	if err != nil {
		t.Fatalf("GenerateCounts: %v", err)
	}
	if got.GasGiants != 4 {
		t.Errorf("GasGiants = %d, want 4", got.GasGiants)
	}
}
