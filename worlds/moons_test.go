package worlds

import (
	"testing"

	"wbh/roller"
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
