package worlds

import (
	"math"
	"testing"

	"wbh/roller"
)

func TestRollBasicSiderealHours_NoDMsNoCascade(t *testing.T) {
	// (2D-2) × 4 + 2 + 1D + DMs.
	// Scripted: 2D=4, 1D=3 → (4-2)×4 + 2 + 3 = 8 + 2 + 3 = 13.
	// Result < 40 so no cascade roll consumed.
	r := roller.NewScripted(4, 3)
	got, err := RollBasicSiderealHours(r, DayLengthDMs{})
	if err != nil {
		t.Fatal(err)
	}
	want := 13.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRollBasicSiderealHours_SystemAgeDM(t *testing.T) {
	// SystemAgeGyr 6.3 → DM+3 (6.3/2 round down = 3).
	// Scripted: 2D=11, 1D=1 → (11-2)×4 + 2 + 1 + 3 = 36+2+1+3 = 42 → cascade fires.
	// Cascade: 1D=4 → < 5 → no addition → final 42.
	r := roller.NewScripted(11, 1, 4)
	got, err := RollBasicSiderealHours(r, DayLengthDMs{SystemAgeGyr: 6.3})
	if err != nil {
		t.Fatal(err)
	}
	want := 42.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRollBasicSiderealHours_CascadeAdds(t *testing.T) {
	// 2D=11, 1D=1, age DM+3 → 42. Cascade 1D=5 → add another (2D-2)×4 + 2 + 1D + DMs.
	// Second roll: 2D=4, 1D=2 → (4-2)×4 + 2 + 2 + 3 = 8+2+2+3 = 15.
	// Total now 42+15 = 57. Cascade 1D=2 → no further addition.
	r := roller.NewScripted(11, 1, 5, 4, 2, 2)
	got, err := RollBasicSiderealHours(r, DayLengthDMs{SystemAgeGyr: 6.3})
	if err != nil {
		t.Fatal(err)
	}
	want := 57.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRollBasicSiderealHours_GGOrSizeSDoublesResult(t *testing.T) {
	// GG/S × 2 doubles the final pre-cascade-aware result.
	// Per spec: "For gas giant or small body (Size 0 or S) rotation, multiply by 2 instead."
	// Scripted: 2D=4, 1D=3 → 13 → ×2 = 26.
	r := roller.NewScripted(4, 3)
	got, err := RollBasicSiderealHours(r, DayLengthDMs{IsGGOrSizeS: true})
	if err != nil {
		t.Fatal(err)
	}
	want := 26.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeYearDays_TerraExample(t *testing.T) {
	// Terra: year ≈ 8766h, sidereal ≈ 23.93h → ~365.25 solar days.
	// year_h / sidereal_h - 1 = 8766/23.93 - 1 = 365.25.
	got := ComputeYearDays(8766.0, 23.93)
	want := 365.25
	if math.Abs(got-want) > 0.5 {
		t.Errorf("got %v, want ~%v", got, want)
	}
}

func TestComputeSolarHours_TerraExample(t *testing.T) {
	// Solar day = year_h / year_days = 8766 / 365.25 = ~24h.
	got := ComputeSolarHours(8766.0, 365.25)
	want := 24.0
	if math.Abs(got-want) > 0.1 {
		t.Errorf("got %v, want ~%v", got, want)
	}
}
