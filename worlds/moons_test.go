package worlds

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
)

// TestCountMoons_PerSizeBand exercises each row of WBH p.55 Significant
// Moon Quantity table at a representative dice value.
func TestCountMoons_PerSizeBand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		parent ParentInfo
		dice   []int // one int per Roll call (summed totals)
		dms    int
		want   int
	}{
		// Size 1-2 → 1D-5: 1D=6 → 1
		{"Size 1 → 1D-5 (1D=6)", ParentInfo{SizeCode: "1"}, []int{6}, 0, 1},
		// Size 3-9 → 2D-8: 2D=10 → 2
		{"Size 5 → 2D-8 (2D=10)", ParentInfo{SizeCode: "5"}, []int{10}, 0, 2},
		// Size 3-9 → 2D-8: 2D=8 → 0 (exactly 0 → ring marker)
		{"Size 9 → 2D-8 (2D=8)", ParentInfo{SizeCode: "9"}, []int{8}, 0, 0},
		// Size A-F → 2D-6: 2D=8 → 2
		{"Size A → 2D-6 (2D=8)", ParentInfo{SizeCode: "A"}, []int{8}, 0, 2},
		// Small GG → 3D-7: 3D=12 → 5
		{"Small GG → 3D-7 (3D=12)", ParentInfo{IsGasGiant: true, GGClass: GasGiantSmall}, []int{12}, 0, 5},
		// Medium GG → 4D-6: 4D=14 → 8
		{"Med GG → 4D-6 (4D=14)", ParentInfo{IsGasGiant: true, GGClass: GasGiantMedium}, []int{14}, 0, 8},
		// Large GG → 4D-6: 4D=16 → 10
		{"Large GG → 4D-6 (4D=16)", ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge}, []int{16}, 0, 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := CountMoons(r, tc.parent, tc.dms)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != tc.want {
				t.Errorf("CountMoons = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestCountMoons_DM applies dms=-1 (one of the WBH p.55 conditions)
// and confirms it shifts the per-die total downward.
func TestCountMoons_DM(t *testing.T) {
	t.Parallel()

	// Size A → 2D-6 with dms=-1 (per-die for 2 dice = -2 total).
	// 2D=8, with dms=-1 (×2 dice = -2): 8 + (-2) - 6 = 0.
	r := roller.NewScripted(8)
	got, err := CountMoons(r, ParentInfo{SizeCode: "A"}, -1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("CountMoons with dms=-1 = %d, want 0", got)
	}

	// Small GG → 3D-7 with dms=-1 (per-die for 3 dice = -3 total).
	// 3D=12, with dms-applied: 12 + (-3) - 7 = 2.
	r = roller.NewScripted(12)
	got, err = CountMoons(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantSmall}, -1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 2 {
		t.Errorf("CountMoons Small GG with dms=-1 = %d, want 2", got)
	}
}

// TestCountMoons_NegativeReturnsZero asserts the negative-result clamp.
func TestCountMoons_NegativeReturnsZero(t *testing.T) {
	t.Parallel()
	// Size 1 → 1D-5: 1D=2 → -3 → clamped to 0.
	r := roller.NewScripted(2)
	got, err := CountMoons(r, ParentInfo{SizeCode: "1"}, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("CountMoons = %d, want 0 (negative clamped)", got)
	}
}

// TestCountMoons_SubOneSizeShortCircuit asserts that bodies with
// SizeCode "0", "R", or "S" return 0 moons immediately without
// consuming a die, since WBH p.55's Quantity table starts at Size 1-2.
func TestCountMoons_SubOneSizeShortCircuit(t *testing.T) {
	t.Parallel()

	cases := []SizeCode{"0", "R", "S"}
	for _, sc := range cases {
		t.Run(string(sc), func(t *testing.T) {
			t.Parallel()
			// Use a NewScripted with no dice — if the function consumes
			// a die, it will panic ("exhausted on Roll(...)").
			r := roller.NewScripted()
			got, err := CountMoons(r, ParentInfo{SizeCode: sc}, 0)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != 0 {
				t.Errorf("CountMoons(SizeCode=%q) = %d, want 0", sc, got)
			}
		})
	}
}

// TestCountMoons_ZedAabIV reproduces a Zed result from p.56:
// Aab IV (GLE) → form Sub=5. Use formula 4D-6 with dms=0:
// 4D=11 → 11-6 = 5.
func TestCountMoons_ZedAabIV(t *testing.T) {
	t.Parallel()
	r := roller.NewScripted(11)
	got, err := CountMoons(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge}, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 5 {
		t.Errorf("CountMoons = %d, want 5", got)
	}
}

// TestSizeMoon_FirstRollBranches covers the three top-level branches
// of WBH p.57 Significant Moon Sizing.
func TestSizeMoon_FirstRollBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		parent   ParentInfo
		dice     []int
		wantSize SizeCode
	}{
		// 1-3 → S (no second roll)
		{"first 1 → S", ParentInfo{SizeCode: "8"}, []int{1}, "S"},
		{"first 3 → S", ParentInfo{SizeCode: "8"}, []int{3}, "S"},
		// 4-5 → D3-1 (range 0(R) to 2)
		// First 1D=4, D3=1 → 0 → "R"
		{"first 4 D3=1 → R (0)", ParentInfo{SizeCode: "8"}, []int{4, 1}, "R"},
		// First 1D=4, D3=2 → 1 → "1"
		{"first 4 D3=2 → 1", ParentInfo{SizeCode: "8"}, []int{4, 2}, "1"},
		// First 1D=5, D3=3 → 2 → "2"
		{"first 5 D3=3 → 2", ParentInfo{SizeCode: "8"}, []int{5, 3}, "2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(tc.dice...)
			got, err := SizeMoon(r, tc.parent)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.SizeCode != tc.wantSize {
				t.Errorf("SizeCode = %q, want %q", got.SizeCode, tc.wantSize)
			}
		})
	}
}

// TestSizeMoon_TerrestrialFirst6 covers the "first roll 6 → Size-1-1D"
// branch for terrestrial parents.
func TestSizeMoon_TerrestrialFirst6(t *testing.T) {
	t.Parallel()

	// Parent Size 8, first 6, 1D=3 → result Size = 8-1-3 = 4.
	r := roller.NewScripted(6, 3)
	got, err := SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "4" {
		t.Errorf("SizeCode = %q, want \"4\" (8-1-3)", got.SizeCode)
	}

	// Parent Size 8, first 6, 1D=6 → 8-1-6 = 1 → Size 1.
	r = roller.NewScripted(6, 6)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "1" {
		t.Errorf("SizeCode = %q, want \"1\"", got.SizeCode)
	}

	// Parent Size 3, first 6, 1D=6 → 3-1-6 = -4 → ring "R" (Size <= 0).
	r = roller.NewScripted(6, 6)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "3"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "R" {
		t.Errorf("SizeCode for negative result = %q, want \"R\"", got.SizeCode)
	}
}

// TestSizeMoon_TerrestrialSize1Parent: any moon < parent's Size becomes
// Size S per WBH p.57.
func TestSizeMoon_TerrestrialSize1Parent(t *testing.T) {
	t.Parallel()

	// Parent Size 1, first 4, D3=2 → Size 1 (= parent, not <) stays Size 1.
	r := roller.NewScripted(4, 2)
	got, err := SizeMoon(r, ParentInfo{SizeCode: "1"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "1" {
		t.Errorf("SizeCode = %q, want \"1\" (= parent, not <)", got.SizeCode)
	}

	// Parent Size 1, first 4, D3=1 → would-be Size 0 → less than 1 → "S".
	r = roller.NewScripted(4, 1)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "1"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "S" {
		t.Errorf("SizeCode = %q, want \"S\" (Size 0 < parent Size 1)", got.SizeCode)
	}
}

// TestSizeMoon_Exactly2Less2DAdjust: when terrestrial moon initial Size
// is exactly 2 less than parent, roll 2D for adjustment.
func TestSizeMoon_Exactly2Less2DAdjust(t *testing.T) {
	t.Parallel()

	// Parent Size 8, first 6, 1D=1 → result 8-1-1=6 → exactly 2 less than 8.
	// 2D=2 → upgrade by 1 → Size 7.
	r := roller.NewScripted(6, 1, 2)
	got, err := SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "7" {
		t.Errorf("SizeCode after 2D=2 = %q, want \"7\" (parent-1)", got.SizeCode)
	}

	// 2D=12 → twin world → Size 8.
	r = roller.NewScripted(6, 1, 12)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "8" {
		t.Errorf("SizeCode after 2D=12 = %q, want \"8\" (twin world)", got.SizeCode)
	}

	// 2D=7 → keep at 6.
	r = roller.NewScripted(6, 1, 7)
	got, err = SizeMoon(r, ParentInfo{SizeCode: "8"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "6" {
		t.Errorf("SizeCode after 2D=7 = %q, want \"6\" (kept at 2-less)", got.SizeCode)
	}
}

// TestSizeMoon_GGSpecial: gas-giant parent on first 6 → GG Special Moon Sizing.
func TestSizeMoon_GGSpecial(t *testing.T) {
	t.Parallel()

	// First 6 → Special. Sub 1D=2 (1-3 branch) → 1D=4 → Size 4.
	r := roller.NewScripted(6, 2, 4)
	got, err := SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "4" {
		t.Errorf("SizeCode (1-3 branch) = %q, want \"4\"", got.SizeCode)
	}

	// First 6 → Special. Sub 1D=4 (4-5 branch) → 2D=8 → 2D-2=6 → Size 6.
	r = roller.NewScripted(6, 4, 8)
	got, err = SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "6" {
		t.Errorf("SizeCode (4-5 branch) = %q, want \"6\"", got.SizeCode)
	}

	// First 6 → Special. Sub 1D=4 → 2D=2 → 2D-2=0 → "R".
	r = roller.NewScripted(6, 4, 2)
	got, err = SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "R" {
		t.Errorf("SizeCode (4-5 → 0) = %q, want \"R\"", got.SizeCode)
	}

	// First 6 → Special. Sub 1D=6 (6 branch) → 2D=6 → 2D+4=10 → Size A.
	r = roller.NewScripted(6, 6, 6)
	got, err = SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "A" {
		t.Errorf("SizeCode (6 → 10) = %q, want \"A\"", got.SizeCode)
	}
}

// TestSizeMoon_GGSpecial_GiantMoonCascade covers the WBH p.57 footnote:
// when GG Special second roll yields Size G (16) the moon is itself a
// Small gas giant (or Medium if the cascade 2D rolls 12).
func TestSizeMoon_GGSpecial_GiantMoonCascade(t *testing.T) {
	t.Parallel()

	// First 6 → Special. Sub 1D=6 → 2D=12 → 2D+4=16 → Size G → cascade to Small GG.
	// Small GG diameter = D3+D3: D3=1, D3=2 → 3.
	// Small GG mass = 5×(1D+1): 1D=4 → 5×5=25.
	// Cascade footnote 2D check: 2D=7 (not 12) → stays Small GG.
	r := roller.NewScripted(6, 6, 12, 1, 2, 4, 7)
	got, err := SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "G" {
		t.Errorf("SizeCode = %q, want \"G\" (GG cascade sentinel)", got.SizeCode)
	}
	if got.GGClass != GasGiantSmall {
		t.Errorf("GGClass = %v, want GasGiantSmall", got.GGClass)
	}
	if got.GGDiameterCode != "3" {
		t.Errorf("GGDiameterCode = %q, want \"3\"", got.GGDiameterCode)
	}
	if got.MassEarth != 25 {
		t.Errorf("MassEarth = %v, want 25", got.MassEarth)
	}
}

// TestSizeMoon_GGSpecial_GiantMoonCascade_Medium covers the WBH p.57
// footnote: when the cascade's additional 2D rolls 12, the moon
// upgrades from Small GG to Medium GG, which re-rolls diameter
// (1D+6) and mass (20×(3D-1)).
func TestSizeMoon_GGSpecial_GiantMoonCascade_Medium(t *testing.T) {
	t.Parallel()

	// Dice script (one int per Roll call):
	//   6     SizeMoon first → GG Special branch
	//   6     gasGiantSpecialMoon first → 6 sub-branch (2D+4)
	//   12    2D+4 = 16 → cascade
	//   1, 2  Small GG diameter D3+D3 (consumed but discarded on Medium upgrade)
	//   4     Small GG mass 1D=4 → 5×(4+1)=25 (discarded)
	//   12    cascade 2D = 12 → upgrade to Medium GG
	//   4     Medium GG diameter 1D=4 → 4+6=10 → "A"
	//   9     Medium GG mass 3D=9 → 20×(9-1) = 160
	r := roller.NewScripted(6, 6, 12, 1, 2, 4, 12, 4, 9)
	got, err := SizeMoon(r, ParentInfo{IsGasGiant: true, GGClass: GasGiantLarge})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SizeCode != "G" {
		t.Errorf("SizeCode = %q, want \"G\" (GG cascade sentinel)", got.SizeCode)
	}
	if got.GGClass != GasGiantMedium {
		t.Errorf("GGClass = %v, want GasGiantMedium (cascade upgrade)", got.GGClass)
	}
	if got.GGDiameterCode != "A" {
		t.Errorf("GGDiameterCode = %q, want \"A\" (1D+6=10)", got.GGDiameterCode)
	}
	if got.MassEarth != 160 {
		t.Errorf("MassEarth = %v, want 160 (20×(9-1))", got.MassEarth)
	}
}
