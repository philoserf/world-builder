package worlds

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
)

func TestRollHydroDigit_TerraEarth(t *testing.T) {
	t.Parallel()
	// Size 8, atmo 6, temperate, no special DMs.
	// 2D=8 + (-7) + 6 = 7 → digit 7
	r := roller.NewScripted(8)
	got, err := RollHydroDigit(r, 6, "", SizeCode("8"), TempTemperate)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Errorf("got %d, want 7", got)
	}
}

func TestRollHydroDigit_AppliesSize0DM(t *testing.T) {
	t.Parallel()
	// Size 0, atmo 6, 2D=12 → 12-7+6-4 = 7
	r := roller.NewScripted(12)
	got, err := RollHydroDigit(r, 6, "", SizeCode("0"), TempTemperate)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Errorf("got %d, want 7", got)
	}
}

func TestRollHydroDigit_AppliesAtmoExtremesDM(t *testing.T) {
	t.Parallel()
	// Atmo 10 (Exotic A): DM-4. 2D=12 + (-7) + 10 - 4 = 11 → cap 10.
	r := roller.NewScripted(12)
	got, err := RollHydroDigit(r, 10, "", SizeCode("8"), TempTemperate)
	if err != nil {
		t.Fatal(err)
	}
	if got != 10 {
		t.Errorf("atmo A: got %d, want 10", got)
	}
	// Atmo 0: DM-4. 2D=12 - 7 + 0 - 4 = 1
	r = roller.NewScripted(12)
	got, err = RollHydroDigit(r, 0, "", SizeCode("8"), TempTemperate)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Errorf("atmo 0: got %d, want 1", got)
	}
}

func TestRollHydroDigit_HotTempDM(t *testing.T) {
	t.Parallel()
	// Hot temp DM-2. Atmo 6 (no exotic mask). 2D=10 + (-7) + 6 - 2 = 7
	r := roller.NewScripted(10)
	got, err := RollHydroDigit(r, 6, "", SizeCode("8"), TempHot)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Errorf("hot: got %d, want 7", got)
	}
	// Atmo D (13): hot DM ignored. 2D=10 + (-7) + 13 = 16 → cap 10
	r = roller.NewScripted(10)
	got, err = RollHydroDigit(r, 13, "", SizeCode("8"), TempHot)
	if err != nil {
		t.Fatal(err)
	}
	if got != 10 {
		t.Errorf("atmo D hot: got %d, want 10", got)
	}
}

func TestHydroRange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		digit int
		want  [2]int
	}{
		{0, [2]int{0, 5}},
		{1, [2]int{6, 15}},
		{6, [2]int{56, 65}},
		{9, [2]int{86, 95}},
		{10, [2]int{96, 100}},
	}
	for _, c := range cases {
		got := HydroRange(c.digit)
		if got != c.want {
			t.Errorf("digit %d: got %v, want %v", c.digit, got, c.want)
		}
	}
}

func TestRefineHydroPercent_Mid(t *testing.T) {
	t.Parallel()
	// Digit 6 → range [56, 65]. d10=5 → 56 + (5-1) = 60.
	r := roller.NewScripted(5)
	got, err := RefineHydroPercent(r, 6, [2]int{56, 65})
	if err != nil {
		t.Fatal(err)
	}
	if got != 60 {
		t.Errorf("got %d, want 60", got)
	}
}

func TestRefineHydroPercent_DigitZero(t *testing.T) {
	t.Parallel()
	// Digit 0: -4 + d10. d10=2 → -2 → 0; d10=8 → 4
	r := roller.NewScripted(2)
	got, err := RefineHydroPercent(r, 0, [2]int{0, 5})
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Errorf("d10=2: got %d, want 0", got)
	}
	r = roller.NewScripted(8)
	got, err = RefineHydroPercent(r, 0, [2]int{0, 5})
	if err != nil {
		t.Fatal(err)
	}
	if got != 4 {
		t.Errorf("d10=8: got %d, want 4", got)
	}
}

func TestRefineHydroPercent_DigitTen(t *testing.T) {
	t.Parallel()
	// Digit 10: 96 + d10. d10=10 → 106 → cap 100
	r := roller.NewScripted(10)
	got, err := RefineHydroPercent(r, 10, [2]int{96, 100})
	if err != nil {
		t.Fatal(err)
	}
	if got != 100 {
		t.Errorf("d10=10: got %d, want 100", got)
	}
}
