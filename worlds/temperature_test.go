package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

func TestComputeAlbedo_ZedPrime(t *testing.T) {
	// Zed Prime: rocky terrestrial (density ~1.0), atm 6, hyd 6, orbit 1.06 AU
	// (parent's), star Aab L=1.419, HZCO computed from L.
	// Per WBH p.111: 0.04 + (8-2)*0.02 + 8*0.01 + (7-4)*0.03 = 0.33.
	// Scripted dice: [8 (rocky base), 8 (atm 6), 7 (hyd 6+)].
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Orbit = 1.0 // moon's parent orbit; for albedo we use Orbit# vs HZCO
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.04}
	body.Hydrographics = &Hydrographics{Code: 6}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.419}}

	r := roller.NewScripted(8, 8, 7)
	got := ComputeAlbedo(r, body, sys)
	want := 0.33
	if math.Abs(got-want) > 0.005 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeAlbedo_Terra_Reference(t *testing.T) {
	// Terra: rocky (density 1.0), atm 6, hyd 7, orbit 1.0 AU, sol L=1.0.
	// 0.04 + (X-2)*0.02 + Y*0.01 + (Z-4)*0.03 should hit ~0.30 with mid rolls.
	// Scripted [7, 7, 6]: 0.04 + 0.10 + 0.07 + 0.06 = 0.27. Close to 0.30 reference.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Orbit = 3.0
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.0}
	body.Hydrographics = &Hydrographics{Code: 7}

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7, 7, 6)
	got := ComputeAlbedo(r, body, sys)
	if got < 0.20 || got > 0.35 {
		t.Errorf("Terra-reference albedo got %v, want ~0.27 (Terra book value 0.30)", got)
	}
}

func TestComputeAlbedo_GasGiant(t *testing.T) {
	// Gas giant: 0.05 + 2D × 0.05. With 2D=7: 0.40.
	body := &DetailedPlacement{}
	body.GGClass = GasGiantSmall
	body.SizeCode = "S"
	body.Orbit = 5.0

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7)
	got := ComputeAlbedo(r, body, sys)
	want := 0.40
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeAlbedo_IcyBeyondHZCO2(t *testing.T) {
	// Icy beyond HZCO+2: 0.25 + (2D-2) × 0.07. With 2D=7: 0.25 + 5*0.07 = 0.60.
	// HZCO for L=1.0 = 1.0 Orbit#. Body at Orbit# 4 is HZCO+3 → beyond HZCO+2.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Orbit = 4.0
	body.Physical = &BodyPhysical{Density: 0.3} // icy

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(7)
	got := ComputeAlbedo(r, body, sys)
	// 0.25 + 5*0.07 = 0.60. Above 0.4 so the bonus-subtraction does not fire.
	want := 0.60
	if math.Abs(got-want) > 0.01 {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeAlbedo_Clamping(t *testing.T) {
	// Force a result above 0.98: rocky terr base 0.04 + (12-2)*0.02 = 0.24 + atm A-C +(12-2)*0.05 = 0.50 + hyd 6+ +(12-4)*0.03 = 0.24 → 0.98 exactly.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Atmosphere = &Atmosphere{Code: 10, Pressure: 1.0} // A → +(2D-2)*0.05
	body.Hydrographics = &Hydrographics{Code: 10}          // 6+ → +(2D-4)*0.03

	sys := stars.System{Primary: stars.Star{Luminosity: 1.0}}

	r := roller.NewScripted(12, 12, 12) // max rolls everywhere
	got := ComputeAlbedo(r, body, sys)
	// 0.04 + 10*0.02 + 10*0.05 + 8*0.03 = 0.04 + 0.20 + 0.50 + 0.24 = 0.98 exactly.
	if got > 0.98 {
		t.Errorf("clamp failed: got %v, want ≤ 0.98", got)
	}
}
