package stars

import (
	"testing"

	"wbh/roller"
)

func TestKindFromUnusualCell(t *testing.T) {
	cases := map[string]StarKind{
		"BD": KindBrownDwarf,
		"D":  KindWhiteDwarf,
	}
	for cell, want := range cases {
		got, err := KindFromUnusualCell(cell)
		if err != nil {
			t.Fatalf("%s error: %v", cell, err)
		}
		if got != want {
			t.Fatalf("%s = %v want %v", cell, got, want)
		}
	}
}

func TestKindFromPeculiarCell(t *testing.T) {
	cases := map[string]StarKind{
		"Black Hole":   KindBlackHole,
		"Pulsar":       KindPulsar,
		"Neutron Star": KindNeutronStar,
		"Nebula":       KindNebula,
		"Protostar":    KindProtostar,
		"Star Cluster": KindStarCluster,
		"Anomaly":      KindAnomaly,
	}
	for cell, want := range cases {
		got, err := KindFromPeculiarCell(cell)
		if err != nil {
			t.Fatalf("%s error: %v", cell, err)
		}
		if got != want {
			t.Fatalf("%s = %v want %v", cell, got, want)
		}
	}
}

func TestRollSpecialPrimary_Simple(t *testing.T) {
	// 1D=3 -> Neutron Star, 1D=6 -> Black Hole.
	r := roller.NewScripted(3)
	got, err := RollSpecialPrimarySimple(r)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != KindNeutronStar {
		t.Fatalf("got %v want neutron star", got)
	}

	r2 := roller.NewScripted(6)
	got2, err := RollSpecialPrimarySimple(r2)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got2 != KindBlackHole {
		t.Fatalf("got %v want black hole", got2)
	}
}

func TestRollSpecialPrimary_Unusual_BD(t *testing.T) {
	// Unusual column at row 5 = "BD".
	r := roller.NewScripted(5)
	kind, _, err := RollSpecialPrimary(r, PeculiarPathUnusual)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if kind != KindBrownDwarf {
		t.Fatalf("got %v want BrownDwarf", kind)
	}
}

func TestRollSpecialPrimary_Unusual_D(t *testing.T) {
	// Unusual column at row 8 = "D".
	r := roller.NewScripted(8)
	kind, _, err := RollSpecialPrimary(r, PeculiarPathUnusual)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if kind != KindWhiteDwarf {
		t.Fatalf("got %v want WhiteDwarf", kind)
	}
}

func TestRollSpecialPrimary_Peculiar_BlackHole(t *testing.T) {
	// Peculiar column at row 2 = "Black Hole".
	r := roller.NewScripted(2)
	kind, _, err := RollSpecialPrimary(r, PeculiarPathPeculiar)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if kind != KindBlackHole {
		t.Fatalf("got %v want BlackHole", kind)
	}
}

func TestRollSpecialPrimary_Peculiar_Anomaly(t *testing.T) {
	// Peculiar column at row 11 = "Anomaly".
	r := roller.NewScripted(11)
	kind, _, err := RollSpecialPrimary(r, PeculiarPathPeculiar)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if kind != KindAnomaly {
		t.Fatalf("got %v want Anomaly", kind)
	}
}

func TestRollSpecialPrimary_Unusual_ClassRedirect(t *testing.T) {
	// Unusual column at row 4 = "Class IV".
	r := roller.NewScripted(4)
	kind, lc, err := RollSpecialPrimary(r, PeculiarPathUnusual)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if kind != "" {
		t.Fatalf("got kind %v, want empty (class redirect)", kind)
	}
	if lc != IV {
		t.Fatalf("got class %v, want IV", lc)
	}
}

func TestRollSpecialPrimary_PeculiarRecursion(t *testing.T) {
	// Unusual column at row 2 = "Peculiar" -> recurse on Peculiar column.
	// Recursive 2D=11 -> Peculiar column = "Anomaly".
	r := roller.NewScripted(2, 11)
	kind, _, err := RollSpecialPrimary(r, PeculiarPathUnusual)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if kind != KindAnomaly {
		t.Fatalf("got %v want Anomaly", kind)
	}
}

func TestGeneratePrimaryAtClass_III(t *testing.T) {
	t.Parallel()

	// Roll sequence:
	//  1. 2D=7 → RollPrimaryTypeAndClass → Type column row 7 = "K"
	//  2. 2D=7 → RollSubtype('K', III) → StarSubtypeNumeric[7] = 9, no IV clamp → K9
	//  3. 1D=1 → SmallStarAge accuracy=1 (oneD)
	//  4. D3=2 → SmallStarAge accuracy=1 (d3) → age = 1×2 + 2 − 1 = 3 Gyr
	rolls := []int{7, 7, 1, 2}
	r := roller.NewScripted(rolls...)
	got, err := generatePrimaryAtClass(r, III, GenerateOpts{Accuracy: 1})
	if err != nil {
		t.Fatalf("generatePrimaryAtClass: %v", err)
	}
	if got.LuminosityClass != III {
		t.Errorf("LuminosityClass = %s, want III", got.LuminosityClass)
	}
	if got.SpectralType.Letter != 'K' {
		t.Errorf("Letter = %c, want K", got.SpectralType.Letter)
	}
	if got.SpectralType.Subtype != 9 {
		t.Errorf("Subtype = %d, want 9 (StarSubtypeNumeric[7] = 9, no IV clamp at Class III)", got.SpectralType.Subtype)
	}
	if got.Kind != KindGiant {
		t.Errorf("Kind = %s, want %s", got.Kind, KindGiant)
	}
	if got.Mass <= 0 {
		t.Errorf("Mass = %v, want > 0", got.Mass)
	}
	if got.Diameter <= 0 {
		t.Errorf("Diameter = %v, want > 0", got.Diameter)
	}
	if got.Temperature <= 0 {
		t.Errorf("Temperature = %v, want > 0", got.Temperature)
	}
	if got.Luminosity <= 0 {
		t.Errorf("Luminosity = %v, want > 0", got.Luminosity)
	}
}
