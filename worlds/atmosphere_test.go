package worlds

import (
	"testing"

	"wbh/roller"
)

func TestHZCOOffsetToTempRange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		orbit, hzco float64
		want        TempRange
	}{
		{1.0, 3.3, TempBoiling},   // offset -2.3
		{1.5, 3.3, TempHot},       // offset -1.8 (in -2.0..-1.01 band)
		{3.3, 3.3, TempTemperate}, // offset 0
		{4.5, 3.3, TempCold},      // offset +1.2
		{7.0, 3.3, TempFrozen},    // offset +3.7
	}
	for _, c := range cases {
		got := HZCOOffsetToTempRange(c.orbit, c.hzco)
		if got != c.want {
			t.Errorf("orbit=%v hzco=%v: got %v, want %v", c.orbit, c.hzco, got, c.want)
		}
	}
}

func TestRollAtmoCode_HZBasic(t *testing.T) {
	t.Parallel()
	// Size 5, 2D=8 → 8-7+5 = 6 → Standard
	r := roller.NewScripted(8)
	got, err := RollAtmoCode(r, SizeCode("5"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 6 {
		t.Errorf("got %d, want 6", got)
	}
}

func TestRollAtmoCode_AutomaticZero(t *testing.T) {
	t.Parallel()
	for _, s := range []SizeCode{"0", "1", "S"} {
		r := roller.NewScripted(12)
		got, err := RollAtmoCode(r, s, 0)
		if err != nil {
			t.Fatal(err)
		}
		if got != 0 {
			t.Errorf("Size %s: got %d, want 0", s, got)
		}
	}
}

func TestRollAtmoCode_ZedAabI(t *testing.T) {
	t.Parallel()
	// Size B (11) at HZCO offset -2.3, 2D=5 → 5-7+11 = 9 → Dense, Tainted
	r := roller.NewScripted(5)
	got, err := RollAtmoCode(r, SizeCode("B"), -2.3)
	if err != nil {
		t.Fatal(err)
	}
	if got != 9 {
		t.Errorf("Aab I atmo code: got %d, want 9", got)
	}
}

func TestAtmosphereCompositionLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code int
		want string
	}{
		{0, "None"},
		{6, "Standard"},
		{9, "Dense, Tainted"},
		{10, "Exotic"},
		{11, "Corrosive"},
		{17, "Gas, Hydrogen"},
	}
	for _, c := range cases {
		got := AtmosphereCompositionLabel(c.code)
		if got != c.want {
			t.Errorf("code %d: got %q, want %q", c.code, got, c.want)
		}
	}
}

func TestSizeAsInt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    SizeCode
		want int
	}{
		{"0", 0},
		{"S", 0},
		{"R", 0},
		{"", 0},
		{"1", 1},
		{"5", 5},
		{"9", 9},
		{"A", 10},
		{"B", 11},
		{"F", 15},
	}
	for _, c := range cases {
		got := SizeAsInt(c.s)
		if got != c.want {
			t.Errorf("%q: got %d, want %d", c.s, got, c.want)
		}
	}
}
