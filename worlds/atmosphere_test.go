package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
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
	got, err := RollTotalPressure(r, 6, "")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-1.0423) > 0.01 {
		t.Errorf("got %v, want 1.0423", got)
	}
}

func TestRollTotalPressure_AtmBCWithSubtype(t *testing.T) {
	cases := []struct {
		name    string
		atmCode int
		subtype string
		// Roll values for the formula's two 1D rolls.
		// scale = ((a-1)*5 + (b-1)) / 30; pressure = min + span*scale.
		a, b int
		want float64
	}{
		// Subtype 6 (Standard): min=0.70, span=0.79.
		// (1,1) → scale=0 → 0.70.
		{"atm B subtype 6 min", 11, "6", 1, 1, 0.70},
		// (6,6) → scale=1 → 0.70+0.79 = 1.49.
		{"atm B subtype 6 max", 11, "6", 6, 6, 1.49},
		// Subtype C: min=10, span=90.
		{"atm C subtype C min", 12, "C", 1, 1, 10},
		{"atm C subtype C max", 12, "C", 6, 6, 100},
		// Subtype E: min=1000, span=9000.
		{"atm C subtype E min", 12, "E", 1, 1, 1000},
		{"atm C subtype E max", 12, "E", 6, 6, 10000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.a, c.b)
			got, err := RollTotalPressure(r, c.atmCode, c.subtype)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if math.Abs(got-c.want) > 1e-9 {
				t.Errorf("got %g, want %g", got, c.want)
			}
		})
	}
}

func TestRollTotalPressure_AtmBCEmptySubtype(t *testing.T) {
	// Empty subtype on atm 11/12 falls back to (0, 0): no rolls consumed,
	// returns 0. This preserves legacy "Varies" behavior for callers that
	// don't have a subtype yet.
	r := roller.NewScripted() // no rolls scripted; if any are consumed, panic.
	got, err := RollTotalPressure(r, 11, "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("got %g, want 0", got)
	}
}

func TestRollTotalPressure_RegularCodeIgnoresSubtype(t *testing.T) {
	// For atm codes outside 11/12, the subtype parameter is ignored.
	// Atm 6 (Standard): min=0.70, span=0.79. (1,1) → 0.70.
	r := roller.NewScripted(1, 1)
	got, err := RollTotalPressure(r, 6, "C") // subtype "C" ignored on atm 6
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0.70 {
		t.Errorf("got %g, want 0.70", got)
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

func TestCorrosiveInsidiousPressureRange(t *testing.T) {
	cases := []struct {
		subtype  string
		wantMin  float64
		wantSpan float64
	}{
		{"1", 0.1, 0.32},
		{"2", 0.1, 0.32},
		{"3", 0.1, 0.32},
		{"4", 0.43, 0.27},
		{"5", 0.43, 0.27},
		{"6", 0.70, 0.79},
		{"7", 0.70, 0.79},
		{"8", 1.50, 0.99},
		{"9", 1.50, 0.99},
		{"A", 2.50, 7.50},
		{"B", 2.50, 7.50},
		{"C", 10, 90},
		{"D", 100, 900},
		{"E", 1000, 9000},
		{"", 0, 0},
		{"0", 0, 0},
		{"Z", 0, 0},
	}
	for _, c := range cases {
		gotMin, gotSpan := corrosiveInsidiousPressureRange(c.subtype)
		if gotMin != c.wantMin || gotSpan != c.wantSpan {
			t.Errorf("corrosiveInsidiousPressureRange(%q): got (%g, %g), want (%g, %g)",
				c.subtype, gotMin, gotSpan, c.wantMin, c.wantSpan)
		}
	}
}

// TestRollAtmoCode_UpperClamp asserts the 0-15 (0-F) world-atmosphere
// clamp: codes 16 (G) and 17 (H) are gas-giant rows of the p.79 table
// and must never be produced for worlds (docs/wbh-inconsistencies.md
// documents this as a package invariant).
func TestRollAtmoCode_UpperClamp(t *testing.T) {
	t.Parallel()
	// Size F (15) with max 2D=12: unclamped 12-7+15 = 20 → must clamp to 15.
	got, err := RollAtmoCode(roller.NewScripted(12), SizeCode("F"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 15 {
		t.Errorf("RollAtmoCode(Size F, 2D=12) = %d, want 15 (clamped)", got)
	}
	// Size B (11) with 2D=12: 12-7+11 = 16 (G) → clamp to 15.
	got, err = RollAtmoCode(roller.NewScripted(12), SizeCode("B"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 15 {
		t.Errorf("RollAtmoCode(Size B, 2D=12) = %d, want 15 (clamped)", got)
	}
}

// TestAtmosphereCodes_TableConsistency asserts the single p.79 table
// transcription is complete (codes 0-17) and its chars match the UWP
// hex convention.
func TestAtmosphereCodes_TableConsistency(t *testing.T) {
	t.Parallel()
	wantChars := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H"}
	for code, want := range wantChars {
		entry, ok := atmosphereCodes[code]
		if !ok {
			t.Errorf("atmosphereCodes missing code %d", code)
			continue
		}
		if entry.Char != want {
			t.Errorf("atmosphereCodes[%d].Char = %q, want %q", code, entry.Char, want)
		}
		if entry.Name == "" {
			t.Errorf("atmosphereCodes[%d].Name is empty", code)
		}
	}
	if len(atmosphereCodes) != len(wantChars) {
		t.Errorf("atmosphereCodes has %d entries, want %d", len(atmosphereCodes), len(wantChars))
	}
}

// TestRollExoticSubtype covers the WBH p.85 Exotic Atmosphere Subtype
// table: the 2D→code mapping (including the 13→A / 14+→B wrap), the four
// DMs, and the low-end clamp. Scripted rolls feed the 2D result directly.
func TestRollExoticSubtype(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		roll    int
		size    SizeCode
		orbit   float64
		hzco    float64
		runaway bool
		want    string
	}{
		// No DMs (size 5, orbit == hzco): total == roll.
		{"7 → 7", 7, "5", 3.0, 3.0, false, "7"},
		{"9 → 9", 9, "5", 3.0, 3.0, false, "9"},
		{"10 → A", 10, "5", 3.0, 3.0, false, "A"},
		{"11 → B", 11, "5", 3.0, 3.0, false, "B"},
		{"12 → C", 12, "5", 3.0, 3.0, false, "C"},
		// Size 2-4 → DM-2: 2D=12, size 3 → 10 → A.
		{"size3 DM-2: 12→10→A", 12, "3", 3.0, 3.0, false, "A"},
		// Orbit < HZCO-1 → DM-2: orbit 1.0, hzco 3.0 → 10-2 → 8.
		{"cold DM-2: 10→8", 10, "5", 1.0, 3.0, false, "8"},
		// Orbit > HZCO+2 → DM+2: orbit 6.0, hzco 3.0 → 8+2 → 10 → A.
		{"far DM+2: 8→10→A", 8, "5", 6.0, 3.0, false, "A"},
		// Runaway → DM+4: 2D=8 → 12 → C.
		{"runaway DM+4: 8→12→C", 8, "5", 3.0, 3.0, true, "C"},
		// Low clamp: size 3 (DM-2) + cold (DM-2), 2D=2 → -2 → "2".
		{"low clamp → 2", 2, "3", 1.0, 3.0, false, "2"},
		// High wrap: 2D=12, far (+2), runaway (+4) → 18 → 14+ → B.
		{"high wrap → B", 12, "5", 6.0, 3.0, true, "B"},
		// Exactly 13 wraps to A: 2D=11, far (+2) → 13 → A.
		{"13 wrap → A", 11, "5", 6.0, 3.0, false, "A"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.roll)
			got, err := RollExoticSubtype(r, c.size, c.orbit, c.hzco, c.runaway)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// TestRollTotalPressure_AtmAWithSubtype covers exotic (code 10 / A) pressure
// derived from the WBH p.85 subtype-keyed range, mirroring the atm B/C path.
func TestRollTotalPressure_AtmAWithSubtype(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		subtype string
		a, b    int
		want    float64
	}{
		// Subtype 6 (Standard): min=0.70, span=0.79. (1,1)→0.70; (6,6)→1.49.
		{"A subtype 6 min", "6", 1, 1, 0.70},
		{"A subtype 6 max", "6", 6, 6, 1.49},
		// Subtype C (Very Dense): min=2.50, span=7.50. (1,1)→2.50; (6,6)→10.0.
		{"A subtype C min", "C", 1, 1, 2.50},
		{"A subtype C max", "C", 6, 6, 10.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.a, c.b)
			got, err := RollTotalPressure(r, 10, c.subtype)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if math.Abs(got-c.want) > 1e-9 {
				t.Errorf("got %g, want %g", got, c.want)
			}
		})
	}
}

// TestRollTotalPressure_AtmAEmptySubtype asserts an exotic code with no
// subtype still falls back to (0, 0): no rolls consumed, returns 0.
func TestRollTotalPressure_AtmAEmptySubtype(t *testing.T) {
	t.Parallel()
	r := roller.NewScripted() // panics if any roll is consumed
	got, err := RollTotalPressure(r, 10, "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != 0 {
		t.Errorf("got %g, want 0", got)
	}
}
