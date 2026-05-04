package worlds

import (
	"math"
	"testing"
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
	removeAll, addRing := MoonRemovalCheck(34)
	if removeAll || addRing {
		t.Errorf("expected (false, false), got (%v, %v)", removeAll, addRing)
	}
	// Boundary check: 1.5 keeps; 1.49 drops
	removeAll, _ = MoonRemovalCheck(1.5)
	if removeAll {
		t.Errorf("limit=1.5: expected keep, got drop")
	}
}

func TestMoonRemovalCheck_Drop(t *testing.T) {
	t.Parallel()
	removeAll, addRing := MoonRemovalCheck(1.0)
	if !removeAll || !addRing {
		t.Errorf("expected (true, true), got (%v, %v)", removeAll, addRing)
	}
	removeAll, addRing = MoonRemovalCheck(0)
	if !removeAll || !addRing {
		t.Errorf("expected (true, true), got (%v, %v)", removeAll, addRing)
	}
}
