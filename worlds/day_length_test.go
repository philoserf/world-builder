package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
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

func TestAddMinuteSecondPrecision(t *testing.T) {
	// 1D-1 for tens, d10 for ones. Range 0-59 for minutes; 0-59 for seconds.
	// Scripted: 1D=3 → tens-min=2; d10=2 → ones-min=2 → minutes=22.
	// Scripted: 1D=2 → tens-sec=1; d10=5 → ones-sec=5 → seconds=15.
	// Total addition: 22/60 + 15/3600 = 0.3667 + 0.00417 = 0.37083 hours.
	r := roller.NewScripted(3, 2, 2, 5)
	got := addMinuteSecondPrecision(r)

	want := 22.0/60.0 + 15.0/3600.0
	if math.Abs(got-want) > 0.0001 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGenerateDayLength_ZedPrimeSidereal(t *testing.T) {
	// Zed Prime: moon of gas giant, around binary Aab (system age 6.3 Gyr).
	// Per p.103 Zed example:
	//   2D=11, 1D=1, age DM+3 → (11-2)×4 + 2 + 1 + 3 = 42 hours.
	//   Cascade 1D=4 → no addition.
	//   Precision: minutes 22, seconds 15 → 22/60 + 15/3600 hours added.
	//   Final: 42 + 22/60 + 15/3600 = 42.37083... h.
	//
	// Year (computed by 2C): 0.805 yr × 8766 h/yr = 7056.63 h.
	// year_days = 7056.63 / 42.37 - 1 ≈ 165.548.
	// solar_hours = 7056.63 / 165.548 ≈ 42.626.
	r := roller.NewScripted(
		11, 1, // basic 2D, 1D
		4,    // cascade 1D
		3, 2, // minutes 1D, d10
		2, 5, // seconds 1D, d10
	)
	dp := &Body{
		Period: Period{Years: 0.805, Days: 0.805 * 365.25, Hours: 0.805 * 8766},
	}
	dp.Kind = BodyTerrestrial
	dp.SizeCode = "5"

	sys := stars.System{Primary: stars.Star{AgeGyr: 6.3}}

	dl, err := GenerateDayLength(r, dp, sys)
	if err != nil {
		t.Fatal(err)
	}

	if dl == nil {
		t.Fatal("expected non-nil DayLength")
	}

	// Pre-effect baseline.
	wantBaseline := 42.37083
	if math.Abs(dl.BaselineSiderealHours-wantBaseline) > 0.001 {
		t.Errorf("BaselineSiderealHours: got %v, want %v", dl.BaselineSiderealHours, wantBaseline)
	}
	// SiderealHours (pre-tidal-lock — same as baseline at this stage).
	if math.Abs(dl.SiderealHours-wantBaseline) > 0.001 {
		t.Errorf("SiderealHours: got %v, want %v", dl.SiderealHours, wantBaseline)
	}
	// YearDays per p.104 example.
	wantYearDays := 7056.63/42.37083 - 1
	if math.Abs(dl.YearDays-wantYearDays) > 0.5 {
		t.Errorf("YearDays: got %v, want %v", dl.YearDays, wantYearDays)
	}
	// SolarHours per p.104 example.
	wantSolar := 7056.63 / wantYearDays
	if math.Abs(dl.SolarHours-wantSolar) > 0.5 {
		t.Errorf("SolarHours: got %v, want %v", dl.SolarHours, wantSolar)
	}
}
