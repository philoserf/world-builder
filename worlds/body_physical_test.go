package worlds

import (
	"fmt"
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
)

func TestRollComposition_TableValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		roll int
		want string
	}{
		{-4, "Exotic Ice"}, // -4 or less
		{-3, "Mostly Ice"}, // -3 to -2
		{-2, "Mostly Ice"},
		{3, "Mostly Rock"}, // 3 to 6
		{6, "Mostly Rock"},
		{7, "Rock and Metal"}, // 7 to 11
		{11, "Rock and Metal"},
		{12, "Mostly Metal"}, // 12 to 14
		{14, "Mostly Metal"},
		{15, "Compressed Metal"}, // 15+
		{99, "Compressed Metal"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("roll=%d", c.roll), func(t *testing.T) {
			t.Parallel()

			r := roller.NewScripted(c.roll)

			got, err := RollComposition(r, BodyPhysicalDMs{})
			if err != nil {
				t.Fatalf("roll=%d: unexpected error: %v", c.roll, err)
			}

			if got != c.want {
				t.Errorf("roll=%d: got %q, want %q", c.roll, got, c.want)
			}
		})
	}
}

func TestRollComposition_AppliesSizeDM(t *testing.T) {
	t.Parallel()

	// Size A-F → DM+3 ⇒ scripted 2D=4 + 3 = 7 → "Rock and Metal"
	r := roller.NewScripted(4)

	got, err := RollComposition(r, BodyPhysicalDMs{SizeCode: SizeCode("A")})
	if err != nil {
		t.Fatal(err)
	}

	if got != "Rock and Metal" {
		t.Errorf("Size A with 2D=4 (DM+3): got %q, want %q", got, "Rock and Metal")
	}
}

func TestRollComposition_AppliesAtHZCODM(t *testing.T) {
	t.Parallel()

	// AtHZCOOrCloser → DM+1. Scripted 2D=6, no other DMs → 6+1 = 7 → "Rock and Metal".
	r := roller.NewScripted(6)

	got, err := RollComposition(r, BodyPhysicalDMs{AtHZCOOrCloser: true})
	if err != nil {
		t.Fatal(err)
	}

	if got != "Rock and Metal" {
		t.Errorf("AtHZCOOrCloser DM+1 with 2D=6: got %q, want %q", got, "Rock and Metal")
	}
}

func TestRollComposition_AppliesBeyondHZCODM(t *testing.T) {
	t.Parallel()

	// BeyondHZCO 2 → DM-2. Scripted 2D=8, no other DMs → 8-2 = 6 → "Mostly Rock".
	r := roller.NewScripted(8)

	got, err := RollComposition(r, BodyPhysicalDMs{BeyondHZCO: 2})
	if err != nil {
		t.Fatal(err)
	}

	if got != "Mostly Rock" {
		t.Errorf("BeyondHZCO=2 with 2D=8: got %q, want %q", got, "Mostly Rock")
	}
}

func TestRollComposition_AppliesSystemAgeDM(t *testing.T) {
	t.Parallel()

	// Threshold check: SystemAgeGyr=10 should NOT trigger DM-1; 11 should.
	// Use 2D=7: without DM → 7 → "Rock and Metal"; with DM-1 → 6 → "Mostly Rock".
	r1 := roller.NewScripted(7)

	got1, err := RollComposition(r1, BodyPhysicalDMs{SystemAgeGyr: 10.0})
	if err != nil {
		t.Fatal(err)
	}

	if got1 != "Rock and Metal" {
		t.Errorf("SystemAgeGyr=10.0 (no DM): got %q, want %q", got1, "Rock and Metal")
	}

	r2 := roller.NewScripted(7)

	got2, err := RollComposition(r2, BodyPhysicalDMs{SystemAgeGyr: 11.0})
	if err != nil {
		t.Fatal(err)
	}

	if got2 != "Mostly Rock" {
		t.Errorf("SystemAgeGyr=11.0 DM-1 with 2D=7: got %q, want %q", got2, "Mostly Rock")
	}
}

func TestRollDensity_RockAndMetalColumn(t *testing.T) {
	t.Parallel()

	// Terrestrial Density table p.71 Rock and Metal column:
	//   2D=2  → 0.82,  2D=7 → 0.97,  2D=12 → 1.12
	cases := []struct {
		roll int
		want float64
	}{
		{2, 0.82},
		{7, 0.97},
		{12, 1.12},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("roll=%d", c.roll), func(t *testing.T) {
			t.Parallel()

			r := roller.NewScripted(c.roll)

			got, err := RollDensity(r, "Rock and Metal")
			if err != nil {
				t.Fatalf("roll=%d: %v", c.roll, err)
			}

			if got != c.want {
				t.Errorf("roll=%d: got %v, want %v", c.roll, got, c.want)
			}
		})
	}
}

func TestDeriveGravity(t *testing.T) {
	t.Parallel()
	// Terra: density=1.0, diameter=12742 → gravity=1.0 G.
	if got := DeriveGravity(1.0, 12742); math.Abs(got-1.0) > 1e-6 {
		t.Errorf("Terra gravity: got %v, want 1.0", got)
	}
	// Zed mainworld worked example: density 1.03, diameter 8163.
	// 1.03 × (8163/12742) ≈ 0.6598 (book rounds to 0.66)
	if got := DeriveGravity(1.03, 8163); math.Abs(got-0.66) > 0.01 {
		t.Errorf("Zed gravity: got %v, want ≈0.66", got)
	}
}

func TestDeriveMass(t *testing.T) {
	t.Parallel()

	if got := DeriveMass(1.0, 12742); math.Abs(got-1.0) > 1e-6 {
		t.Errorf("Terra mass: got %v, want 1.0", got)
	}
	// Zed: 1.03 × 0.640559³ ≈ 0.2708, book rounds to 0.27
	if got := DeriveMass(1.03, 8163); math.Abs(got-0.27) > 0.01 {
		t.Errorf("Zed mass: got %v, want ≈0.27", got)
	}
}

func TestDeriveEscapeVelocity(t *testing.T) {
	t.Parallel()
	// Book p.72 worked example: m=0.27, D=8163 → EscV ≈ 7,262 m/s.
	if got := DeriveEscapeVelocity(0.27, 8163); math.Abs(got-7262) > 50 {
		t.Errorf("Zed escape vel: got %v, want ≈7262", got)
	}
	// Terra sanity: m=1.0, D=12742 → EscV ≈ 11,186 m/s.
	if got := DeriveEscapeVelocity(1.0, 12742); math.Abs(got-11186) > 5 {
		t.Errorf("Terra escape vel: got %v, want ≈11186", got)
	}
}

func TestDeriveOrbitalVelocity(t *testing.T) {
	t.Parallel()
	// EscV / √2 = 7262 / 1.41421 ≈ 5135 m/s
	if got := DeriveOrbitalVelocity(7262); math.Abs(got-5135) > 5 {
		t.Errorf("Zed orbital vel: got %v, want ≈5135", got)
	}
}

func TestFormatSizeProfile_Zed(t *testing.T) {
	t.Parallel()

	p := BodyPhysical{Density: 1.03, Gravity: 0.66}
	p.SizeProfile = FormatSizeProfile(p, 0.27, SizeCode("5"), 8163)

	want := "5-8163-1.03-0.66-0.27"
	if p.SizeProfile != want {
		t.Errorf("got %q, want %q", p.SizeProfile, want)
	}
}

func TestSol_TerraPhysicalProfile(t *testing.T) {
	t.Parallel()
	// WBH pp. 71-72 worked example for the Zed mainworld (Size 5, treated like Terra in 3A1).
	// Diameter pre-resolved to 8163 km (Size 5 base 7,200 + D3=2 (+600) + 1D=4 (+300) + d100=63).
	// Composition: 2D=10 + DM+1 (AtHZCOOrCloser DM+1; no size DM) → 11 → "Rock and Metal".
	// Density 2D=9 → index 7 → 1.03 in Rock-and-Metal column (table: 2D2=0.82 … 2D9=1.03).
	// Scripted values are pre-DM raw rolls: D3=2, 1D=4, d100=63, 2D-comp=10, 2D-density=9.
	r := roller.NewScripted(2, 4, 63, 10, 9) // D3, 1D, d100, 2D-comp, 2D-density

	got, err := GenerateBodyPhysical(r, SizeCode("5"), 8163, BodyPhysicalDMs{
		AtHZCOOrCloser: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.Composition != "Rock and Metal" {
		t.Errorf("composition: got %q, want %q", got.Composition, "Rock and Metal")
	}

	if math.Abs(got.Density-1.03) > 0.001 {
		t.Errorf("density: got %v, want 1.03", got.Density)
	}

	if math.Abs(got.Gravity-0.66) > 0.01 {
		t.Errorf("gravity: got %v, want 0.66", got.Gravity)
	}

	if got.SizeProfile != "5-8163-1.03-0.66-0.27" {
		t.Errorf("profile: got %q, want %q", got.SizeProfile, "5-8163-1.03-0.66-0.27")
	}
}
