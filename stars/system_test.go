package stars

import (
	"errors"
	"testing"

	"wbh/roller"
)

func TestSystem_TypeFieldsZeroValue(t *testing.T) {
	var s System
	if s.PrimaryDesignation != "" {
		t.Fatalf("zero value PrimaryDesignation: %q", s.PrimaryDesignation)
	}
	if len(s.Companions) != 0 {
		t.Fatalf("zero value Companions: %v", s.Companions)
	}
}

func TestCompanionStar_TypeFieldsZeroValue(t *testing.T) {
	var c CompanionStar
	if c.Designation != "" {
		t.Fatalf("zero value Designation: %q", c.Designation)
	}
	if c.ParentIndex != 0 {
		t.Fatalf("zero value ParentIndex: %d", c.ParentIndex)
	}
}

// TestGenerateSystem_NoCompanions drives all four presence rolls below the
// 10+ threshold so GenerateSystem produces a single-star system.
//
// Roll sequence (WithVariance=true, Accuracy=1):
//  1. 2D=9  → Star Type: G
//  2. 2D=6  → Subtype: 7  (G7 V)
//  3. 2D-7=0 → mass variance (no change)
//  4. 2D-7=0 → diameter variance (no change)
//  5. 1D=3  → age 1D component
//  6. D3=2  → age D3 component  (age = 3×2+2-1 = 7 Gyr)
//  7. 2D=9  → Close presence: 9+0 = 9 < 10 → absent
//  8. 2D=9  → Near presence: absent
//  9. 2D=9  → Far presence: absent
//
// 10. 2D=9  → Primary companion presence: absent
func TestGenerateSystem_NoCompanions(t *testing.T) {
	r := roller.NewScripted(
		9, 6, // primary type (G) and subtype (7 → G7)
		0, 0, // mass/diameter variance
		3, 2, // age: 1D=3, D3=2 → 7 Gyr
		9, 9, 9, 9, // all four presence rolls below threshold
	)
	sys, err := GenerateSystem(r, GenerateSystemOpts{WithVariance: true, Accuracy: 1})
	if err != nil {
		t.Fatalf("GenerateSystem: %v", err)
	}
	if len(sys.Companions) != 0 {
		t.Fatalf("expected 0 companions, got %d", len(sys.Companions))
	}
	if sys.PrimaryDesignation != "A" {
		t.Fatalf("primary designation: got %q want A", sys.PrimaryDesignation)
	}
	if sys.Primary.SpectralType != (SpectralType{Letter: 'G', Subtype: 7}) {
		t.Fatalf("primary spectral type: got %v want G7", sys.Primary.SpectralType)
	}
	if sys.Primary.LuminosityClass != V {
		t.Fatalf("primary class: got %v want V", sys.Primary.LuminosityClass)
	}
}

// TestGenerateSystem_PrimaryWithCompanion drives a single OrbitCompanion
// companion off the primary (Twin descriptor).
//
// Roll sequence (WithVariance=true, Accuracy=1):
//
// Primary generation (rolls 1-6): same G7 V as above.
//
// Presence rolls (rolls 7-10):
//  7. 2D=9  → Close absent
//  8. 2D=9  → Near absent
//  9. 2D=9  → Far absent
//
// 10. 2D=10 → Primary companion present (10+0=10 ≥ 10)
//
// No Close/Near/Far to generate companion-presence rolls for.
//
// Companion star generation (rolls 11-13):
//  11. 2D=10 → RollNonPrimaryDescriptor (Companion col, row 10) = "Twin"
//     GenerateCompanionStar("Twin") consumes no dice.
//     Physical quantities: ComputeMass/Diameter/Temperature consume no dice.
//  12. 2D-7=0 → mass variance for Twin
//  13. 2D-7=0 → diameter variance for Twin
//
// Orbital placement (rolls 14-19):
// 14. 1D=3  → RollStellarOrbit OrbitCompanion first part (3/10=0.3)
// 15. 2D-7=0 → second part (0+0/100 = 0.3 orbit#)
// 16. 2D=7  → eccentricity (IsStar DM+2 → row 9)
// 17. 1D=3  → eccentricity second roll (0.03 + 3/100 = 0.06)
// 18. 2D=4  → inclination: 4 ≤ 6 → VeryLow
// 19. 1D=2  → VeryLow degrees: 2/2 = 1.0°
func TestGenerateSystem_PrimaryWithCompanion(t *testing.T) {
	r := roller.NewScripted(
		9, 6, // primary G7 V
		0, 0, // mass/diameter variance
		3, 2, // age → 7 Gyr
		9, 9, 9, 10, // presence: Close/Near/Far absent, primary companion present
		10,   // descriptor roll → row 10 Companion col = "Twin"
		0, 0, // Twin mass/diameter variance
		3, 0, // orbit: 1D=3, 2D-7=0 → 0.3 orbit#
		7, 3, // eccentricity: 2D=7 → row 9; 1D=3 → 0.03+3/100=0.06
		4, 2, // inclination: 2D=4 → VeryLow; 1D=2 → 2/2=1.0°
	)
	sys, err := GenerateSystem(r, GenerateSystemOpts{WithVariance: true, Accuracy: 1})
	if err != nil {
		t.Fatalf("GenerateSystem: %v", err)
	}
	if len(sys.Companions) != 1 {
		t.Fatalf("expected 1 companion, got %d", len(sys.Companions))
	}
	if sys.PrimaryDesignation != "Aa" {
		t.Fatalf("primary designation: got %q want Aa", sys.PrimaryDesignation)
	}
	comp := sys.Companions[0]
	if comp.Designation != "Ab" {
		t.Fatalf("companion designation: got %q want Ab", comp.Designation)
	}
	if comp.OrbitClass != OrbitCompanion {
		t.Fatalf("companion orbit class: got %v want OrbitCompanion", comp.OrbitClass)
	}
	if comp.Star.SpectralType != sys.Primary.SpectralType {
		t.Fatalf("twin spectral type mismatch: %v vs %v", comp.Star.SpectralType, sys.Primary.SpectralType)
	}
	if comp.OrbitNumber != 0.3 {
		t.Fatalf("orbit number: got %v want 0.3", comp.OrbitNumber)
	}
}

func TestGenerateSystem_SpecialPrimary_BD(t *testing.T) {
	// Drive: 2D type = 2 (Special) -> ErrSpecialPrimary;
	// then 2D on Unusual column = 6 -> "BD";
	// then small-star age (1D=3, D3=2) -> 7 Gyr (no progenitor for BD).
	// Then four presence rolls all below threshold (no companions).
	// Then no companion-presence sub-rolls either.
	r := roller.NewScripted(
		2,    // primary type 2D = 2 -> Special
		6,    // Unusual column 2D = 6 -> "BD"
		3, 2, // small-star age 1D=3, D3=2 -> 7 Gyr
		// Multiple Stars Presence rolls (4 total, all below 10):
		9, 9, 9, 9,
	)
	sys, err := GenerateSystem(r, GenerateSystemOpts{
		WithVariance: true,
		Accuracy:     1,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if sys.Primary.Kind != KindBrownDwarf {
		t.Fatalf("primary kind: got %v want BrownDwarf", sys.Primary.Kind)
	}
	if sys.Primary.LuminosityClass != BD {
		t.Fatalf("primary class: got %v want BD", sys.Primary.LuminosityClass)
	}
	if len(sys.Companions) != 0 {
		t.Fatalf("expected no companions, got %d", len(sys.Companions))
	}
}

func TestGenerateSystem_SpecialPrimary_D(t *testing.T) {
	// 2D=2 -> Special; Unusual 2D=8 -> "D" (white dwarf).
	// White dwarf age = small-star + progenitor: 1D=2, D3=2 (small_star=5), D3=1 (progenitor mult 3).
	// progenitor mass = 3 * 0.6 = 1.8; FinalAgeProgenitor(1.8) ≈ 2.86 Gyr.
	// Total ≈ 7.86 Gyr.
	r := roller.NewScripted(
		2,    // Type=2 -> Special
		8,    // Unusual=8 -> "D"
		2, 2, // small-star: 1D=2, D3=2 -> 5
		1, // progenitor D3=1 -> mult 3
		// Presence rolls (4 below threshold):
		9, 9, 9, 9,
	)
	sys, err := GenerateSystem(r, GenerateSystemOpts{
		WithVariance: true,
		Accuracy:     1,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if sys.Primary.Kind != KindWhiteDwarf {
		t.Fatalf("primary kind: got %v want WhiteDwarf", sys.Primary.Kind)
	}
	if sys.Primary.LuminosityClass != D {
		t.Fatalf("primary class: got %v want D", sys.Primary.LuminosityClass)
	}
	// Age should be ~7.86 Gyr (small_star=5 + progenitor=2.86).
	if sys.Primary.AgeGyr < 7.5 || sys.Primary.AgeGyr > 8.5 {
		t.Fatalf("age: got %v Gyr, want roughly 7.86", sys.Primary.AgeGyr)
	}
}

func TestGenerateSystem_SpecialPrimary_ClassRedirectErrors(t *testing.T) {
	// 2D=2 -> Special; Unusual 2D=4 -> "Class IV" -> class redirect.
	r := roller.NewScripted(2, 4)
	_, err := GenerateSystem(r, GenerateSystemOpts{Accuracy: 1})
	if err == nil {
		t.Fatal("expected error for class redirect")
	}
	if !errors.Is(err, ErrSpecialPrimaryClassRedirect) {
		t.Fatalf("expected ErrSpecialPrimaryClassRedirect, got: %v", err)
	}
}
