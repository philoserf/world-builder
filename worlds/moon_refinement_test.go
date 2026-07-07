package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
)

func TestHillSphere_AabIV(t *testing.T) {
	t.Parallel()
	// WBH p.75: Aab IV is a GG with 14 × Size-8 diameter = 14 × 12,800 = 179,200 km.
	// au=1.06, ecc=0.10, mass=1,200⊕, stellar=1.836☉ → 0.083AU ≈ 69.37 PD.
	auResult, pd := HillSphere(1.06, 0.10, 1200, 1.836, 179200)
	if math.Abs(auResult-0.083) > 0.001 {
		t.Errorf("Hill sphere AU: got %v, want ≈0.083", auResult)
	}
	if math.Abs(pd-69.37) > 0.5 {
		t.Errorf("Hill sphere PD: got %v, want ≈69.37", pd)
	}
}

func TestHillSphereMoonLimit(t *testing.T) {
	t.Parallel()
	if got := HillSphereMoonLimit(69.37); got != 34 {
		t.Errorf("got %v, want 34", got)
	}
	if got := HillSphereMoonLimit(2.9); got != 1 {
		t.Errorf("got %v, want 1", got)
	}
	if got := HillSphereMoonLimit(0.5); got != 0 {
		t.Errorf("got %v, want 0", got)
	}
}

func TestRocheLimit(t *testing.T) {
	t.Parallel()
	// 1.22 × 12800 × ³√(1.0/0.5) = 1.22 × 12800 × 1.2599 ≈ 19,675 km
	got := RocheLimit(12800, 1.0, 0.5)
	want := 1.22 * 12800 * math.Pow(2.0, 1.0/3.0)
	if math.Abs(got-want) > 1 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRocheLimit_ZeroDensity(t *testing.T) {
	t.Parallel()
	if got := RocheLimit(12800, 1.0, 0); got != 0 {
		t.Errorf("zero moon density: got %v, want 0", got)
	}
}

func TestMoonRemovalCheck_Keep(t *testing.T) {
	t.Parallel()
	if MoonRemovalCheck(34) {
		t.Errorf("expected keep, got drop")
	}
	// Boundary check: 1.5 keeps; 1.49 drops
	if MoonRemovalCheck(1.5) {
		t.Errorf("limit=1.5: expected keep, got drop")
	}
}

func TestMoonRemovalCheck_Drop(t *testing.T) {
	t.Parallel()
	if !MoonRemovalCheck(1.0) {
		t.Errorf("limit=1.0: expected drop, got keep")
	}
	if !MoonRemovalCheck(0) {
		t.Errorf("limit=0: expected drop, got keep")
	}
}

func TestMoonOrbitRange(t *testing.T) {
	t.Parallel()
	// WBH p.77: MOR = floor(hillSphereMoonLimit) - 2
	// Aab IV: HSML=34.685 → floor=34 → MOR=32
	if got := MoonOrbitRange(34.685, 5); got != 32 {
		t.Errorf("got %d, want 32", got)
	}
	// > 200 case: clamp to 200 + nMoons
	if got := MoonOrbitRange(500, 3); got != 203 {
		t.Errorf("got %d, want 203", got)
	}
	// Negative result → floor at 0
	if got := MoonOrbitRange(1.5, 0); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestZedPrime_OrbitalPeriod(t *testing.T) {
	t.Parallel()
	// WBH p.78 worked example: Zed Prime = Aab IV-d at PD=22.
	// Parent Aab IV is a Size-8 GG with diameter multiplier 14,
	// so effectiveSize = 14 × 8 = 112. Parent mass = 1,200⊕.
	// Book result: 0.17693 × √((22×14×8)³ ÷ 1200) ≈ 624.69 hours = 26.03 days.
	period := MoonPeriodHours(22, 112, 1200)
	if math.Abs(period-624.69) > 0.5 {
		t.Errorf("Zed Prime period: got %v, want ≈624.69 hours", period)
	}
}

func TestRollMoonOrbit_Inner(t *testing.T) {
	t.Parallel()
	// MOR=32, DM+1 because MOR < 60.
	// 1D roll=2 → rng=2+1=3 → Inner. 2D roll=7 → v=7.
	// orbit = (7-2) × 32/60 + 2 = 5 × 0.5333 + 2 = 4.667
	r := roller.NewScripted(2, 7)
	orbitPD, mr := RollMoonOrbit(r, 32)
	if mr != MoonRangeInner {
		t.Errorf("range: got %v, want Inner", mr)
	}
	want := 5.0*32.0/60.0 + 2.0
	if math.Abs(orbitPD-want) > 0.01 {
		t.Errorf("orbitPD: got %v, want %v", orbitPD, want)
	}
}

func TestRollMoonOrbit_Middle(t *testing.T) {
	t.Parallel()
	// MOR=32, DM+1 because MOR < 60.
	// 1D roll=3 → rng=3+1=4 → Middle. 2D roll=9 → v=9.
	// orbit = (9-2) × 32/30 + 32/6 + 3 = 7×1.0667 + 5.333 + 3 = 15.8
	r := roller.NewScripted(3, 9)
	orbitPD, mr := RollMoonOrbit(r, 32)
	if mr != MoonRangeMiddle {
		t.Errorf("range: got %v, want Middle", mr)
	}
	want := 7.0*32.0/30.0 + float64(32)/6.0 + 3.0
	if math.Abs(orbitPD-want) > 0.01 {
		t.Errorf("orbitPD: got %v, want %v", orbitPD, want)
	}
}

func TestRollMoonOrbit_Outer(t *testing.T) {
	t.Parallel()
	// MOR=32, DM+1 because MOR < 60.
	// 1D roll=5 → rng=5+1=6 → Outer. 2D roll=8 → v=8.
	// orbit = (8-2) × 32/20 + 32/2 + 4 = 6×1.6 + 16 + 4 = 29.6
	r := roller.NewScripted(5, 8)
	orbitPD, mr := RollMoonOrbit(r, 32)
	if mr != MoonRangeOuter {
		t.Errorf("range: got %v, want Outer", mr)
	}
	want := 6.0*32.0/20.0 + float64(32)/2.0 + 4.0
	if math.Abs(orbitPD-want) > 0.01 {
		t.Errorf("orbitPD: got %v, want %v", orbitPD, want)
	}
}

func TestRollMoonRetrograde(t *testing.T) {
	t.Parallel()
	// Outer range DM+4. Roll=6 → 6+4=10 → retrograde.
	r := roller.NewScripted(6)
	if !RollMoonRetrograde(r, MoonRangeOuter, false) {
		t.Errorf("expected retrograde")
	}
	// Inner range DM-1. Roll=9 → 9-1=8 → not retrograde.
	r = roller.NewScripted(9)
	if RollMoonRetrograde(r, MoonRangeInner, false) {
		t.Errorf("expected prograde")
	}
	// exceedsMOR adds DM+6. Middle DM+1 + DM+6 = DM+7. Roll=3 → 3+7=10 → retrograde.
	r = roller.NewScripted(3)
	if !RollMoonRetrograde(r, MoonRangeMiddle, true) {
		t.Errorf("expected retrograde (exceedsMOR)")
	}
}
