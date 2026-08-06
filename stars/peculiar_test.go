package stars

import (
	"testing"

	"github.com/philoserf/world-builder/roller"
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

// TestRollSpecialPrimary_Special_ClassRedirects verifies the Special
// column (the cleaner Referee default per WBH p.15) routes every cell
// to a class redirect within the mainstream pp.14-146 ruleset — no
// BD/D/Peculiar primaries ever reached.
func TestRollSpecialPrimary_Special_ClassRedirects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		roll    int
		wantLC  LuminosityClass
		wantStr string
	}{
		{"row2_ClassVI", 2, VI, ""},
		{"row5_ClassVI", 5, VI, ""},
		{"row6_ClassIV", 6, IV, ""},
		{"row8_ClassIV", 8, IV, ""},
		{"row9_ClassIII", 9, III, ""},
		{"row10_ClassIII", 10, III, ""},
		{"row11_Giants", 11, "Giants", ""},
		{"row12_Giants", 12, "Giants", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.roll)

			kind, lc, err := RollSpecialPrimary(r, PeculiarPathSpecial)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			if kind != "" {
				t.Errorf("got kind %v, want empty (class redirect)", kind)
			}

			if lc != c.wantLC {
				t.Errorf("got class %v, want %v", lc, c.wantLC)
			}
		})
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

// TestGeneratePrimaryAtClass_IV_M_to_K covers a regression: the Class IV
// redirect must apply the WBH p.16 letter constraint M → K before
// rolling subtype (and computing physical values), otherwise the lookup
// hits M0/M5/M9 — none of which carry a Class IV cell — and fails with
// "M0 class IV missing". Seed 212 hits this in the 10k sweep.
func TestGeneratePrimaryAtClass_IV_M_to_K(t *testing.T) {
	t.Parallel()

	// Roll sequence (class redirects use DM+1 on the Type column, p.16):
	//  1. 2D=3 → +1 → row 4 = "M". M maps to K for Class IV per p.16.
	//  2. 2D=7 → StarSubtypeNumeric[7] = 9; K-IV-subtype>4 shift → 4 → K4
	//  3. 1D=1, D3=2 → SmallStarAge accuracy=1 → age 3 Gyr
	r := roller.NewScripted(3, 7, 1, 2)

	got, err := generatePrimaryAtClass(r, IV, GenerateOpts{Accuracy: 1})
	if err != nil {
		t.Fatalf("generatePrimaryAtClass: %v", err)
	}

	if got.LuminosityClass != IV {
		t.Errorf("LuminosityClass = %s, want IV", got.LuminosityClass)
	}

	if got.SpectralType.Letter != 'K' {
		t.Errorf("Letter = %c, want K (M→K constraint)", got.SpectralType.Letter)
	}

	if got.SpectralType.Subtype != 4 {
		t.Errorf("Subtype = %d, want 4 (9 - 5 K-IV-subtype>4 shift)", got.SpectralType.Subtype)
	}

	if got.Mass <= 0 {
		t.Errorf("Mass = %v, want > 0", got.Mass)
	}
}

// TestGeneratePrimaryAtClass_VI_F_to_G covers the parallel regression
// for Class VI: F must map to G per p.16, since F0/F5 carry no VI
// cell. Seed 6547 hits this in the 10k sweep.
func TestGeneratePrimaryAtClass_VI_F_to_G(t *testing.T) {
	t.Parallel()

	// Roll sequence (class redirects use DM+1 on the Type column, p.16):
	//  1. 2D=10 → +1 → row 11 = "F". F maps to G for Class VI per p.16.
	//  2. 2D=7 → StarSubtypeNumeric[7] = 9 → G9 (no class-IV shift for VI)
	//  3. 1D=1, D3=2 → age 3 Gyr
	r := roller.NewScripted(10, 7, 1, 2)

	got, err := generatePrimaryAtClass(r, VI, GenerateOpts{Accuracy: 1})
	if err != nil {
		t.Fatalf("generatePrimaryAtClass: %v", err)
	}

	if got.LuminosityClass != VI {
		t.Errorf("LuminosityClass = %s, want VI", got.LuminosityClass)
	}

	if got.SpectralType.Letter != 'G' {
		t.Errorf("Letter = %c, want G (F→G constraint)", got.SpectralType.Letter)
	}

	if got.SpectralType.Subtype != 9 {
		t.Errorf("Subtype = %d, want 9", got.SpectralType.Subtype)
	}

	if got.Mass <= 0 {
		t.Errorf("Mass = %v, want > 0", got.Mass)
	}
}

func TestGeneratePrimaryAtClass_III(t *testing.T) {
	t.Parallel()

	// Roll sequence (class redirects use DM+1 on the Type column, p.16):
	//  1. 2D=6 → +1 → row 7 = "K".
	//  2. 2D=7 → RollSubtype('K', III) → StarSubtypeNumeric[7] = 9, no IV clamp → K9
	//  3. 1D=1 → SmallStarAge accuracy=1 (oneD)
	//  4. D3=2 → SmallStarAge accuracy=1 (d3) → age = 1×2 + 2 − 1 = 3 Gyr
	rolls := []int{6, 7, 1, 2}
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

func TestParsePeculiarPath(t *testing.T) {
	t.Parallel()

	ok := map[string]PeculiarPath{
		"special":  PeculiarPathSpecial,
		"unusual":  PeculiarPathUnusual,
		"peculiar": PeculiarPathPeculiar,
	}
	for in, want := range ok {
		got, err := ParsePeculiarPath(in)
		if err != nil {
			t.Errorf("ParsePeculiarPath(%q) errored: %v", in, err)
		}

		if got != want {
			t.Errorf("ParsePeculiarPath(%q) = %q, want %q", in, got, want)
		}
	}

	for _, bad := range []string{"", "Special", "giant", "bogus"} {
		if _, err := ParsePeculiarPath(bad); err == nil {
			t.Errorf("ParsePeculiarPath(%q) = nil error, want error", bad)
		}
	}
}
