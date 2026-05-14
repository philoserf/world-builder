package worlds

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
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
		{10, 4},
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
	// Trailing rolls: belts existence = 0 (no belts), terrestrials 2D=2
	// (forces reroll path), D3=3. This test asserts only GasGiants; the
	// trailing values just satisfy the scripted roller for the rest of
	// the GenerateCounts pipeline.
	r := roller.NewScripted(14, 14, 0, 2, 3)
	got, err := GenerateCounts(r, sys, CountsOpts{})
	if err != nil {
		t.Fatalf("GenerateCounts: %v", err)
	}
	if got.GasGiants != 4 {
		t.Errorf("GasGiants = %d, want 4", got.GasGiants)
	}
}

func TestGenerateCounts_Belts_QuantityTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		roll int
		want int
	}{
		{3, 1}, // 6-: 1
		{6, 1},
		{7, 2}, // 7-11: 2
		{11, 2},
		{12, 3}, // 12+: 3
		{15, 3},
	}
	sys := solSystem() // single G2 V, no companions, no belt DMs
	for _, tc := range cases {
		// GG existence = 10 (>9 → no GG → no GG-present DM on belts).
		// Belts existence raw = 8 (≥8, no DMs apply).
		// Belt quantity raw = tc.roll, no DMs.
		// Terrestrials: trailing rolls forward-compatible.
		r := roller.NewScripted(10, 8, tc.roll, 7, 1)
		got, err := GenerateCounts(r, sys, CountsOpts{})
		if err != nil {
			t.Fatalf("roll %d: %v", tc.roll, err)
		}
		if got.PlanetoidBelts != tc.want {
			t.Errorf("roll %d: PlanetoidBelts = %d, want %d", tc.roll, got.PlanetoidBelts, tc.want)
		}
	}
}

func TestGenerateCounts_Belts_DM_GasGiantsPresent(t *testing.T) {
	t.Parallel()
	// Sol single G2 V system. GG existence raw=5 (≤9 → present). GG quantity raw=7 + DM+1
	// (single-Class-V) = 8 → row 7-8 → 3 GGs.
	//
	// Belts existence: Sol primary is NOT Special Circumstances, so existence DMs = 0.
	// Raw must be ≥8 to be present. Use raw 8.
	//
	// Belt quantity: DMs = +1 (GGs present, 3 of them).
	// Raw 5 + 1 = 6 → row 6- → 1 belt.
	//
	// Terrestrials: trailing rolls forward-compatible.
	r := roller.NewScripted(5, 7, 8, 5, 7, 1)
	got, err := GenerateCounts(r, solSystem(), CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.PlanetoidBelts != 1 {
		t.Errorf("PlanetoidBelts = %d, want 1", got.PlanetoidBelts)
	}
}

func TestGenerateCounts_Belts_DM_PostStellarPrimary(t *testing.T) {
	t.Parallel()
	// White-dwarf primary, solo system. Belt existence DMs:
	//   - post-stellar primary (flat): +1
	//   - per-post-stellar-object (count includes primary): +1
	//   = +2 total.
	// Existence raw 6 + 2 = 8 → present.
	// Belt quantity DMs: same +2 (no GGs present, single-star → no 2+-stars DM).
	// Quantity raw 4 + 2 = 6 → row 6- → 1 belt.
	//
	// GG: existence raw 14 + DMs (BD=0, post-stellar primary=-2, per-post-stellar=-1)
	//   = 14 - 3 = 11 > 9 → no GGs.
	//
	// Terrestrials: trailing rolls forward-compatible.
	wd := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindWhiteDwarf, LuminosityClass: stars.D, Mass: 0.6, Diameter: 0.013, Temperature: 8000,
	})
	sys := stars.System{Primary: wd}
	r := roller.NewScripted(14, 6, 4, 7, 1)
	got, err := GenerateCounts(r, sys, CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.PlanetoidBelts != 1 {
		t.Errorf("PlanetoidBelts = %d, want 1 (post-stellar primary +2 DM)", got.PlanetoidBelts)
	}
}

func TestGenerateCounts_Belts_DM_Protostar(t *testing.T) {
	t.Parallel()
	proto := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindProtostar, LuminosityClass: stars.V, Mass: 1.0, Diameter: 1.0, Temperature: 4000, AgeGyr: 0.001,
	})
	sys := stars.System{Primary: proto}
	// Protostar primary IS Special Circumstances. DMs apply to both rolls.
	// GG existence raw 10 + DMs (protostar primary = 0 GG DM since GG DMs are unrelated;
	// single-Class-V doesn't apply since this is Protostar; no other GG DMs apply) = 10 > 9 → no GGs.
	//
	// Belts existence: Special Circumstances DMs = +3 (protostar) + +2 (primordial age 0.001) = +5.
	// Raw 5 + 5 = 10 ≥8 → present.
	//
	// Belt quantity: same DMs +5. Raw 5 + 5 = 10 → row 7-11 → 2 belts.
	//
	// Terrestrials: trailing.
	r := roller.NewScripted(10, 5, 5, 7, 1)
	got, err := GenerateCounts(r, sys, CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.PlanetoidBelts != 2 {
		t.Errorf("PlanetoidBelts = %d, want 2", got.PlanetoidBelts)
	}
}

func TestGenerateCounts_Terrestrials_LowReroll(t *testing.T) {
	t.Parallel()
	// Sol single G2 V. No DMs on terrestrials.
	// 2D = 4 → 4-2 = 2 (< 3) → reroll D3+2; D3=2 → result 4.
	r := roller.NewScripted(10 /*GG existence none*/, 7 /*belts existence none*/, 4 /*terrestrials raw*/, 2 /*D3 reroll*/)
	got, err := GenerateCounts(r, solSystem(), CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Terrestrials != 4 {
		t.Errorf("Terrestrials = %d, want 4", got.Terrestrials)
	}
}

func TestGenerateCounts_Terrestrials_HighAdd(t *testing.T) {
	t.Parallel()
	// 2D = 8 → 8-2 = 6 (≥3) → add D3-1 with D3=3 → +2 → 8.
	r := roller.NewScripted(10, 7, 8, 3)
	got, err := GenerateCounts(r, solSystem(), CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Terrestrials != 8 {
		t.Errorf("Terrestrials = %d, want 8", got.Terrestrials)
	}
}

func TestGenerateCounts_Terrestrials_PostStellarDM(t *testing.T) {
	t.Parallel()
	// Brown-dwarf primary: terrestrials DM-1 (post-stellar count = 1, the primary).
	// 2D = 8 → 8-2-1 = 5 (≥3) → add D3-1 with D3=3 → +2 → 7.
	bd := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindBrownDwarf, LuminosityClass: stars.BD, Mass: 0.05, Diameter: 0.1, Temperature: 1500,
	})
	// GG existence: raw 14 + DMs(-2 BD, -2 post-stellar primary, -1 per post-stellar count) = -5 → 9 → present.
	// GG quantity: raw 14 + DMs -5 = 9 → 4 GGs.
	// Belts existence: raw 0 + DMs(+1 post-stellar primary, +1 per-count) = +2 → 2 < 8 → not present.
	// Terrestrials: 2D=8, D3=3.
	r := roller.NewScripted(14, 14, 0, 8, 3)
	got, err := GenerateCounts(r, stars.System{Primary: bd}, CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Terrestrials != 7 {
		t.Errorf("Terrestrials = %d, want 7", got.Terrestrials)
	}
}

func TestGenerateCounts_Total(t *testing.T) {
	t.Parallel()
	// Sol single G2 V.
	// GG existence raw 5 (no existence DMs) = 5 ≤ 9 → present.
	// GG quantity raw 7 + DM+1 (single-Class-V) = 8 → row 7-8 → 3 GGs.
	// Belts existence raw 8 (no existence DMs) = 8 → present.
	// Belts quantity raw 5 + DM+1 (GGs present) = 6 → row 6- → 1 belt.
	// Terrestrials 2D=8 → 6 (≥3) + D3-1 (D3=2 → +1) → 7.
	// Total = 3+1+7 = 11.
	r := roller.NewScripted(5, 7, 8, 5, 8, 2)
	got, err := GenerateCounts(r, solSystem(), CountsOpts{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Total != got.GasGiants+got.PlanetoidBelts+got.Terrestrials {
		t.Errorf("Total = %d, sum = %d", got.Total, got.GasGiants+got.PlanetoidBelts+got.Terrestrials)
	}
	if got.Total != 11 {
		t.Errorf("Total = %d, want 11", got.Total)
	}
}
