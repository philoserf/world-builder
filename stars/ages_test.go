package stars

import (
	"math"
	"testing"

	"wbh/roller"
)

func TestMainSequenceLifespan_Sol(t *testing.T) {
	got := MainSequenceLifespan(1.0)
	if math.Abs(got-10.0) > 1e-9 {
		t.Fatalf("got %v want 10.0", got)
	}
}

func TestMainSequenceLifespan_Zed(t *testing.T) {
	// WBH p. 21: 0.929^-2.5 * 10 = 12.022 Gyr.
	got := MainSequenceLifespan(0.929)
	if math.Abs(got-12.022) > 1e-2 {
		t.Fatalf("got %v want ~12.022", got)
	}
}

func TestSmallStarAge_Basic(t *testing.T) {
	// 1D=3, D3=1 -> 3*2 + 1 - 1 = 6 Gyr.
	r := roller.NewScripted(3, 1)
	got, err := SmallStarAge(r, 1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-6.0) > 1e-9 {
		t.Fatalf("got %v want 6.0", got)
	}
}

func TestSmallStarAge_Zed(t *testing.T) {
	// WBH p. 21: 1D=3, D3=2 -> 3*2 + 2 - 1 = 7 Gyr.
	r := roller.NewScripted(3, 2)
	got, err := SmallStarAge(r, 1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-7.0) > 1e-9 {
		t.Fatalf("got %v want 7.0", got)
	}
}

func TestSmallStarAge_Accuracy2(t *testing.T) {
	// WBH p. 21: 1D=3, D3=2, d10=3 -> 3*2 + 2 - 2 + 0.3 = 6.3 Gyr.
	r := roller.NewScripted(3, 2, 3)
	got, err := SmallStarAge(r, 2)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if math.Abs(got-6.3) > 1e-9 {
		t.Fatalf("got %v want 6.3", got)
	}
}

func TestSmallStarAge_InvalidAccuracy(t *testing.T) {
	r := roller.NewScripted(3, 2)
	if _, err := SmallStarAge(r, 0); err == nil {
		t.Fatal("expected error for accuracy 0")
	}
	r2 := roller.NewScripted(3, 2)
	if _, err := SmallStarAge(r2, 3); err == nil {
		t.Fatal("expected error for accuracy 3")
	}
}

func TestSubgiantLifespan_Zed(t *testing.T) {
	got := SubgiantLifespan(12.022, 0.929)
	want := 12.022 / (4 + 0.929)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestGiantLifespan_Zed(t *testing.T) {
	got := GiantLifespan(12.022, 0.929)
	mass := 0.929
	want := 12.022 / (10 * mass * mass * mass)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestFinalAgeProgenitor_WhiteDwarf(t *testing.T) {
	// WBH p. 30: progenitor mass 1.47 -> 4.635 Gyr.
	got := FinalAgeProgenitor(1.47)
	if math.Abs(got-4.635) > 1e-2 {
		t.Fatalf("got %v want ~4.635", got)
	}
}

func TestFinalAgeProgenitor_UnitMass(t *testing.T) {
	// At mass=1.0, final age = 10 * (1 + 1/5 + 1/10) = 13.0.
	got := FinalAgeProgenitor(1.0)
	if math.Abs(got-13.0) > 1e-9 {
		t.Fatalf("got %v want 13.0", got)
	}
}
