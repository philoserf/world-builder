package worlds

import (
	"fmt"
	"testing"

	"wbh/roller"
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
