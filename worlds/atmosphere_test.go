package worlds

import (
	"math"
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

func TestAtmospherePressureRange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code         int
		minBar, span float64
	}{
		{0, 0, 0.0009},
		{1, 0.001, 0.089},
		{2, 0.1, 0.32},
		{6, 0.7, 0.79},
		{9, 1.5, 0.99},
		{13, 2.5, 7.5},
		{14, 0.10, 0.32},
		{10, 0, 0}, // "Varies"
	}
	for _, c := range cases {
		gotMin, gotSpan := AtmospherePressureRange(c.code)
		if math.Abs(gotMin-c.minBar) > 0.001 || math.Abs(gotSpan-c.span) > 0.01 {
			t.Errorf("code %d: got (%v, %v), want (%v, %v)", c.code, gotMin, gotSpan, c.minBar, c.span)
		}
	}
}

func TestRollTotalPressure_ZedPrime(t *testing.T) {
	t.Parallel()
	// Atmo 6: min 0.7, span 0.79
	// Book p.80: 1D-1=2 → ×5=10; 1D-1=3 → +3 = 13. Pressure = 0.7 + 0.79 × 13/30 = 1.0423.
	// Scripted: first 1D=3 (1D-1=2), second 1D=4 (1D-1=3).
	r := roller.NewScripted(3, 4)
	got, err := RollTotalPressure(r, 6)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-1.0423) > 0.01 {
		t.Errorf("got %v, want 1.0423", got)
	}
}

func TestRollOxygenFraction_AgeDMs(t *testing.T) {
	t.Parallel()
	// Verify each age band produces correct DM. 1D=5, 2D=7, 1D=1 → without DMs: 5/20 + 0 + 0 = 0.25.
	// Age > 4 → DM+1 → (5+1)/20 + 0 + 0 = 0.30.
	r := roller.NewScripted(5, 7, 1)
	got, err := RollOxygenFraction(r, 6.336) // > 4 Gyr → DM+1
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-0.30) > 0.001 {
		t.Errorf("age>4 Gyr: got %v, want 0.30", got)
	}
}

func TestScaleHeight_Terra(t *testing.T) {
	t.Parallel()
	// g=1.0, T=288K → H ≈ 8.5 km
	if got := DeriveScaleHeight(288, 1.0); math.Abs(got-8.5) > 0.1 {
		t.Errorf("Terra scale height: got %v, want 8.5", got)
	}
	// g=0.66, T=288K → H ≈ 8.5/0.66 ≈ 12.88 km (book p.82 worked example)
	if got := DeriveScaleHeight(288, 0.66); math.Abs(got-12.88) > 0.1 {
		t.Errorf("Zed scale height: got %v, want 12.88", got)
	}
}

func TestScaleHeight_Edge(t *testing.T) {
	t.Parallel()
	if got := DeriveScaleHeight(288, 0); got != 0 {
		t.Errorf("g=0: got %v, want 0", got)
	}
}

func TestRollCorrosiveInsidiousSubtype_AabI(t *testing.T) {
	t.Parallel()
	// Aab I: Size B (11), orbit 1.0, hzco 3.3, corrosive, no runaway.
	// DMs: Size 8+ → +2; Orbit < HZCO-1 → +4. Total +6.
	// 2D=7 + 6 = 13 → subtype "D".
	r := roller.NewScripted(7)
	got, err := RollCorrosiveInsidiousSubtype(r, SizeCode("B"), 1.0, 3.3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "D" {
		t.Errorf("Aab I subtype: got %q, want %q", got, "D")
	}
}

func TestRollCorrosiveInsidiousSubtype_Boundaries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		roll        int
		size        SizeCode
		orbit, hzco float64
		insidious   bool
		runaway     bool
		want        string
	}{
		{"min: 2D=2, no DM → 2", 2, "5", 3.3, 3.3, false, false, "2"},
		{"max: 2D=12 + DM+2 → 14 → E", 12, "8", 3.3, 3.3, false, false, "E"},
		{"2D=10 → A", 10, "5", 3.3, 3.3, false, false, "A"},
		{"insidious DM+2: 2D=10 + 2 = 12 → C", 10, "5", 3.3, 3.3, true, false, "C"},
		{"runaway DM+4: 2D=8 + 4 = 12 → C", 8, "5", 3.3, 3.3, false, true, "C"},
		{"size 2-4 DM-3: 2D=10 - 3 = 7 → 7", 10, "3", 3.3, 3.3, false, false, "7"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := roller.NewScripted(c.roll)
			got, err := RollCorrosiveInsidiousSubtype(r, c.size, c.orbit, c.hzco, c.insidious, c.runaway)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
