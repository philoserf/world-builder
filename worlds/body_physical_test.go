package worlds

import (
	"testing"

	"wbh/roller"
)

func TestRollComposition_TableValues(t *testing.T) {
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
		r := roller.NewScripted(c.roll)
		got, err := RollComposition(r, BodyPhysicalDMs{})
		if err != nil {
			t.Fatalf("roll=%d: unexpected error: %v", c.roll, err)
		}
		if got != c.want {
			t.Errorf("roll=%d: got %q, want %q", c.roll, got, c.want)
		}
	}
}

func TestRollComposition_AppliesSizeDM(t *testing.T) {
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

func TestRollDensity_RockAndMetalColumn(t *testing.T) {
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
		r := roller.NewScripted(c.roll)
		got, err := RollDensity(r, "Rock and Metal")
		if err != nil {
			t.Fatalf("roll=%d: %v", c.roll, err)
		}
		if got != c.want {
			t.Errorf("roll=%d: got %v, want %v", c.roll, got, c.want)
		}
	}
}
