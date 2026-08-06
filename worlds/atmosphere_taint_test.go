package worlds

import (
	"math"
	"testing"

	"github.com/philoserf/world-builder/roller"
)

func TestTaintSubtypeFromTotal_AllResults(t *testing.T) {
	cases := []struct {
		total int
		want  string
	}{
		{-3, "L"},
		{0, "L"},
		{1, "L"},
		{2, "L"},
		{3, "R"},
		{4, "B"},
		{5, "G"},
		{6, "P"},
		{7, "G"},
		{8, "S"},
		{9, "B"},
		{10, "P"},
		{11, "R"},
		{12, "H"},
		{15, "H"},
		{99, "H"},
	}
	for _, c := range cases {
		got := taintSubtypeFromTotal(c.total)
		if got != c.want {
			t.Errorf("total=%d: got %q, want %q", c.total, got, c.want)
		}
	}
}

func TestRollTaintSubtype_AtmosphereDMs(t *testing.T) {
	cases := []struct {
		atmCode int
		want    string
	}{
		{2, "S"}, // 8 + 0 = 8 → S
		{4, "P"}, // 8 - 2 = 6 → P
		{7, "S"}, // 8 + 0 = 8 → S
		{9, "P"}, // 8 + 2 = 10 → P
	}
	for _, c := range cases {
		r := roller.Fixed(8) // 2D=8 every call

		got := RollTaintSubtype(r, c.atmCode, false)
		if got != c.want {
			t.Errorf("atm %d 2D=8: got %q, want %q", c.atmCode, got, c.want)
		}
	}
}

func TestRollTaintSubtype_LHSuppressionOnNon4to9(t *testing.T) {
	// 2D=2 → L; on atm 10 (outside 4-9) → G (suppressed).
	r := roller.NewScripted(2)

	got := RollTaintSubtype(r, 10, false)
	if got != "G" {
		t.Errorf("atm 10 2D=2: got %q, want \"G\" (L suppressed)", got)
	}
	// 2D=12 → H; on atm 11 (outside 4-9) → G (suppressed).
	r = roller.NewScripted(12)

	got = RollTaintSubtype(r, 11, false)
	if got != "G" {
		t.Errorf("atm 11 2D=12: got %q, want \"G\" (H suppressed)", got)
	}
}

func TestRollTaintSubtype_LHSuppressionOnSecondOrLater(t *testing.T) {
	// 2D=2 → L on atm 7; isSecondOrLater=true → G.
	r := roller.NewScripted(2)

	got := RollTaintSubtype(r, 7, true)
	if got != "G" {
		t.Errorf("atm 7 2D=2 second roll: got %q, want \"G\" (L suppressed)", got)
	}
}

func TestRollTaintSeverity_BasicTable(t *testing.T) {
	// 2D values → expected severity per WBH p.84.
	cases := []struct {
		twoD int
		want int
	}{
		{2, 1},
		{4, 1},
		{5, 2},
		{6, 3},
		{7, 4},
		{8, 5},
		{9, 6},
		{10, 7},
		{11, 8},
		{12, 9},
	}
	for _, c := range cases {
		r := roller.NewScripted(c.twoD)

		got := RollTaintSeverity(r, "B", 4, 0) // B taint, atm 4 (no DM), no ppO2 override
		if got != c.want {
			t.Errorf("2D=%d: got severity %d, want %d", c.twoD, got, c.want)
		}
	}
}

func TestRollTaintSeverity_LowOxygenPpO2Override(t *testing.T) {
	cases := []struct {
		ppO2 float64
		want int
	}{
		{0.10, 2},  // ≥ 0.09
		{0.085, 3}, // ≥ 0.08, < 0.09
		{0.05, 8},  // < 0.08
	}
	for _, c := range cases {
		r := roller.NewScripted(99) // unused since override fires

		got := RollTaintSeverity(r, "L", 4, c.ppO2)
		if got != c.want {
			t.Errorf("L ppO2=%g: got %d, want %d", c.ppO2, got, c.want)
		}
	}
}

func TestRollTaintSeverity_HighOxygenPpO2Override(t *testing.T) {
	cases := []struct {
		ppO2 float64
		want int
	}{
		{0.55, 2}, // < 0.6
		{0.65, 7}, // [0.6, 0.7)
		{0.75, 8}, // ≥ 0.7
	}
	for _, c := range cases {
		r := roller.NewScripted(99)

		got := RollTaintSeverity(r, "H", 4, c.ppO2)
		if got != c.want {
			t.Errorf("H ppO2=%g: got %d, want %d", c.ppO2, got, c.want)
		}
	}
}

func TestRollTaintSeverity_InsidiousDM(t *testing.T) {
	// atm C (12) gets DM+6. 2D=4 + 6 = 10 → 7.
	r := roller.NewScripted(4)

	got := RollTaintSeverity(r, "B", 12, 0)
	if got != 7 {
		t.Errorf("atm C B taint 2D=4: got %d, want 7", got)
	}
}

func TestRollTaintPersistence_BasicTable(t *testing.T) {
	cases := []struct {
		twoD int
		want int
	}{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
		{6, 6},
		{7, 7},
		{8, 8},
		{9, 9},
		{12, 9},
	}
	for _, c := range cases {
		r := roller.NewScripted(c.twoD)

		got := RollTaintPersistence(r, "B", 4, 5) // atm 4 no DM, severity 5 no DM trigger
		if got != c.want {
			t.Errorf("2D=%d: got persistence %d, want %d", c.twoD, got, c.want)
		}
	}
}

func TestRollTaintPersistence_LHDM(t *testing.T) {
	// L/H taint → DM+4. 2D=2 + 4 = 6 → 6.
	r := roller.NewScripted(2)

	got := RollTaintPersistence(r, "L", 4, 5)
	if got != 6 {
		t.Errorf("L taint 2D=2 DM+4: got %d, want 6", got)
	}
}

func TestRollTaintPersistence_HighSeverityDM(t *testing.T) {
	// Severity ≥ 8 → DM+6. 2D=2 + 6 = 8 → 8.
	r := roller.NewScripted(2)

	got := RollTaintPersistence(r, "B", 4, 8)
	if got != 8 {
		t.Errorf("B taint severity 8 2D=2 DM+6: got %d, want 8", got)
	}
}

func TestRollTaintPersistence_InsidiousDM(t *testing.T) {
	// Atm C → DM+6. 2D=2 + 6 = 8 → 8.
	r := roller.NewScripted(2)

	got := RollTaintPersistence(r, "B", 12, 5)
	if got != 8 {
		t.Errorf("atm C B taint 2D=2 DM+6: got %d, want 8", got)
	}
}

func TestRollInsidiousHazard_AllResults(t *testing.T) {
	cases := []struct {
		twoD int
		want string
	}{
		{2, "B"},
		{4, "B"},
		{5, "R"},
		{6, "G"},
		{7, "G"},
		{8, "T"},
		{9, "G"},
		{10, "T"},
		{11, "R"},
		{12, "T"},
	}
	for _, c := range cases {
		r := roller.NewScripted(c.twoD)

		got := RollInsidiousHazard(r, false)
		if got != c.want {
			t.Errorf("2D=%d: got %q, want %q", c.twoD, got, c.want)
		}
	}
}

func TestRollInsidiousHazard_ExtremelyDenseDM(t *testing.T) {
	// 2D=4 + DM+2 = 6 → G (without DM would be B).
	r := roller.NewScripted(4)

	got := RollInsidiousHazard(r, true)
	if got != "G" {
		t.Errorf("extremely dense 2D=4 DM+2: got %q, want \"G\"", got)
	}
}

func TestRollAllTaints_NoPreseed_SingleTaint(t *testing.T) {
	// Atm 7, no DM. Subtype 2D=4 → B (no reroll); severity 2D=7 → 4; persistence 2D=5 → 5.
	r := roller.NewScripted(4, 7, 5)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 7, Pressure: 1.0, OxygenPartialPressure: 0.21},
	}

	taints := RollAllTaints(r, body, nil)
	if len(taints) != 1 {
		t.Fatalf("got %d taints, want 1", len(taints))
	}

	if taints[0] != (Taint{Code: "B", Severity: 4, Persistence: 5}) {
		t.Errorf("got %+v, want {B, 4, 5}", taints[0])
	}
}

func TestRollAllTaints_UnsupportedAtmosphere_NoTaints(t *testing.T) {
	r := roller.NewScripted(7, 7, 7)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 6, Pressure: 1.0, OxygenPartialPressure: 0.21},
	}

	taints := RollAllTaints(r, body, nil)
	if taints != nil {
		t.Fatalf("got %+v, want nil", taints)
	}
}

func TestRollAllTaints_PreseededL_FillsSevPers(t *testing.T) {
	// Pre-seeded L on atm 4 ppO2=0.05.
	// Severity uses ppO2 override (0.05 < 0.08) → 8 (no roll consumed).
	// Persistence: L → DM+4; severity 8 → DM+6; total +10.
	// 2D=2 + 10 = 12 → 9. (One Roll consumed for persistence.)
	// Then taint #2 subtype roll: 2D=8 + DM-2 (atm 4) = 6 → P.
	//   isSecondOrLater=true so L/H suppressed (P doesn't trigger anyway).
	// Severity for P at atm 4: 2D=7 → 4.
	// Persistence for P at atm 4 sev 4: 2D=5 → 5.
	preseeded := &Taint{Code: "L"}
	r := roller.NewScripted(
		2, // persistence for preseeded L: 2D=2 + DM+4 + DM+6 = 12 → 9
		8, // subtype #2 raw roll: 2D=8 + DM-2 = 6 → P
		7, // severity for P
		5, // persistence for P
	)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 4, Pressure: 0.5, OxygenPartialPressure: 0.05},
	}

	taints := RollAllTaints(r, body, preseeded)
	if len(taints) != 2 {
		t.Fatalf("got %d taints, want 2", len(taints))
	}

	if taints[0] != (Taint{Code: "L", Severity: 8, Persistence: 9}) {
		t.Errorf("preseeded got %+v, want {L, 8, 9}", taints[0])
	}

	if taints[1] != (Taint{Code: "P", Severity: 4, Persistence: 5}) {
		t.Errorf("rolled got %+v, want {P, 4, 5}", taints[1])
	}
}

func TestRollAllTaints_MaxThree(t *testing.T) {
	// Atm 7 (no DM). Three rolls of 10 → P, reroll, P, reroll, P, no further.
	// Each P: severity 2D=7 → 4; persistence 2D=5 → 5.
	r := roller.NewScripted(
		10, 7, 5, // taint 1: P 4 5
		10, 7, 5, // taint 2: P 4 5
		10, 7, 5, // taint 3: P 4 5 (no further reroll)
	)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 7, Pressure: 1.0, OxygenPartialPressure: 0.21},
	}

	taints := RollAllTaints(r, body, nil)
	if len(taints) != 3 {
		t.Fatalf("got %d taints, want 3 (max)", len(taints))
	}

	for i, tt := range taints {
		if tt.Code != "P" {
			t.Errorf("taint #%d: got %q, want P", i, tt.Code)
		}
	}
}

func TestRollAllTaints_LRollAdjustsPpO2(t *testing.T) {
	// Atm 4, ppO2=0.21. Subtype 2D=4 + DM-2 = 2 → L.
	// L is in 4-9 band on first roll: not suppressed. Fresh L: 1D adjust roll.
	// 1D=3 → ppO2 -= 0.03 → ppO2 becomes 0.18.
	// Severity for L at ppO2=0.18 → override: 0.18 ≥ 0.09 → 2.
	// Persistence: L → DM+4; severity 2 < 8 → no extra DM. 2D=2 + 4 = 6 → 6.
	r := roller.NewScripted(
		4, // subtype 2D=4 → L (with DM-2 = 2)
		3, // 1D for ppO2 adjust
		// severity uses override, no roll consumed
		2, // persistence 2D=2 + DM+4 = 6
	)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 4, Pressure: 0.5, OxygenPartialPressure: 0.21},
	}
	originalPress := body.Atmosphere.Pressure

	taints := RollAllTaints(r, body, nil)
	if len(taints) != 1 {
		t.Fatalf("got %d taints, want 1", len(taints))
	}

	if taints[0].Code != "L" || taints[0].Severity != 2 || taints[0].Persistence != 6 {
		t.Errorf("got %+v, want {L, 2, 6}", taints[0])
	}

	wantPpO2 := 0.21 - 0.03
	if math.Abs(body.Atmosphere.OxygenPartialPressure-wantPpO2) > 1e-9 {
		t.Errorf("ppO2 got %g, want %g", body.Atmosphere.OxygenPartialPressure, wantPpO2)
	}

	if body.Atmosphere.Pressure != originalPress {
		t.Errorf("total pressure changed to %g, want unchanged %g", body.Atmosphere.Pressure, originalPress)
	}
}

// ---- Worked-example regression tests (WBH pp.85, 88, 90) -------------------

func TestAabVd_TaintProfile_p85(t *testing.T) {
	// WBH p.85 Aab V d desert moon, atm 4, ppO2 = 0.114 bar (in band — no
	// promotion). Taint subtype rolls land on P (severity 6, persistence 3)
	// + R (severity 5, persistence 4).
	r := roller.NewScripted(
		12, // subtype #1 raw 2D=12 + DM-2 = 10 → P, reroll
		9,  // severity for P: 2D=9 → 6
		3,  // persistence for P: 2D=3 → 3
		5,  // subtype #2 raw 2D=5 + DM-2 = 3 → R
		8,  // severity for R: 2D=8 → 5
		4,  // persistence for R: 2D=4 → 4
	)
	body := &Body{
		Atmosphere: &Atmosphere{
			Code:                  4,
			Pressure:              0.544,
			OxygenPartialPressure: 0.114,
		},
	}

	taints := RollAllTaints(r, body, nil)
	if len(taints) != 2 {
		t.Fatalf("got %d taints, want 2", len(taints))
	}

	if taints[0] != (Taint{Code: "P", Severity: 6, Persistence: 3}) {
		t.Errorf("taint #1: got %+v, want {P, 6, 3}", taints[0])
	}

	if taints[1] != (Taint{Code: "R", Severity: 5, Persistence: 4}) {
		t.Errorf("taint #2: got %+v, want {R, 5, 4}", taints[1])
	}
}

func TestAabVb_ExoticIrritant_p88(t *testing.T) {
	// WBH p.88 Aab V b exotic atm A, dense subtype 9. Book example uses an
	// orbit DM not modeled in our RollTaintSubtype (book shows 9+2=11 → R);
	// our impl applies 0 DM for atm A. Script 2D=11 directly to land on R.
	// Severity for R: 2D=5 → 2 (book "2:surmountable").
	// Persistence for R: 2D=9 → 9 (book "9:constant"; severity 2 < 8 so no
	// extra DM; R isn't L/H so no L/H DM. Atm A isn't atm C so no DM+6.)
	r := roller.NewScripted(
		11, // subtype 2D=11 → R (no DMs in our impl on atm A)
		5,  // severity 2D=5 → 2
		9,  // persistence 2D=9 → 9
	)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 10, Subtype: "9", Pressure: 2.09},
	}

	taints := RollAllTaints(r, body, nil)
	if len(taints) != 1 {
		t.Fatalf("got %d taints, want 1", len(taints))
	}

	if taints[0] != (Taint{Code: "R", Severity: 2, Persistence: 9}) {
		t.Errorf("got %+v, want {R, 2, 9}", taints[0])
	}

	if body.Atmosphere.Code != 10 || body.Atmosphere.Subtype != "9" {
		t.Errorf("atmosphere mutated unexpectedly: %+v", body.Atmosphere)
	}
}

func TestAaBVI_CorrosiveProfile_p90(t *testing.T) {
	// WBH p.90 AaB VI corrosive atm B, subtype 6 (standard). Book example
	// does NOT roll an irritant — referee elects gas-mix as the corrosive
	// cause. Our impl always rolls. Test asserts structural invariants
	// without pinning specific irritant values.
	r := roller.NewScripted(
		7, // subtype 2D=7 → G (gas mix)
		7, // severity 2D=7 → 4
		7, // persistence 2D=7 → 7
	)
	body := &Body{
		Atmosphere: &Atmosphere{Code: 11, Subtype: "6", Pressure: 1.21},
	}

	taints := RollAllTaints(r, body, nil)
	if len(taints) != 1 {
		t.Fatalf("got %d taints, want 1 (always-roll policy)", len(taints))
	}

	if body.Atmosphere.Code != 11 || body.Atmosphere.Subtype != "6" {
		t.Errorf("atmosphere mutated unexpectedly: %+v", body.Atmosphere)
	}

	for _, tt := range taints {
		if tt.Code == "L" || tt.Code == "H" {
			t.Errorf("L/H not suppressed on atm B: got %+v", tt)
		}
	}
}

func TestRollAllTaints_Invariants(t *testing.T) {
	// Property test across many seeds.
	for seed := int64(1); seed <= 50; seed++ {
		r := roller.NewSeeded(seed)
		body := &Body{
			Atmosphere: &Atmosphere{Code: 7, Pressure: 1.0, OxygenPartialPressure: 0.21},
		}

		taints := RollAllTaints(r, body, nil)
		if len(taints) > 3 {
			t.Errorf("seed=%d: got %d taints, want ≤ 3", seed, len(taints))
		}

		for i, tt := range taints {
			if tt.Severity < 1 || tt.Severity > 9 {
				t.Errorf("seed=%d taint #%d: severity %d out of [1,9]", seed, i, tt.Severity)
			}

			if tt.Persistence < 2 || tt.Persistence > 9 {
				t.Errorf("seed=%d taint #%d: persistence %d out of [2,9]", seed, i, tt.Persistence)
			}

			if i > 0 && (tt.Code == "L" || tt.Code == "H") {
				t.Errorf("seed=%d taint #%d: L/H must be suppressed on 2nd/3rd rolls", seed, i)
			}
		}
	}
}

func TestRollInsidiousHazards(t *testing.T) {
	cases := []struct {
		name             string
		subtype          string
		isExtremelyDense bool
		twoD             int
		want             []Hazard
	}{
		// Subtype "6" (no D/E rule, no DM): rolled-B from 2D=4.
		{"subtype 6 single", "6", false, 4, []Hazard{{Code: "B"}}},
		// Subtype "C" (no D/E rule, DM+2 from extremely-dense): rolled-G from 2D=4 + DM+2 = 6 → G.
		{"subtype C single with DM", "C", true, 4, []Hazard{{Code: "G"}}},
		// Subtype "D" (D/E rule fires, DM+2): auto-T + rolled-G (2D=4 + DM+2 = 6 → G).
		{"subtype D auto-T plus rolled", "D", true, 4, []Hazard{{Code: "T"}, {Code: "G"}}},
		// Subtype "E" (D/E rule fires, DM+2): auto-T + rolled-B (2D=2 + DM+2 = 4 → B).
		{"subtype E auto-T plus rolled", "E", true, 2, []Hazard{{Code: "T"}, {Code: "B"}}},
		// Duplicate-T case: subtype "D", no DM, 2D=8 → T as rolled, plus auto-T.
		{"subtype D duplicate T", "D", false, 8, []Hazard{{Code: "T"}, {Code: "T"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.twoD)

			got := RollInsidiousHazards(r, c.subtype, c.isExtremelyDense)
			if len(got) != len(c.want) {
				t.Fatalf("len: got %d, want %d (got %+v)", len(got), len(c.want), got)
			}

			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("[%d]: got %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestIsExtremelyDenseSubtype(t *testing.T) {
	cases := []struct {
		subtype string
		want    bool
	}{
		{"C", true},
		{"D", true},
		{"E", true},
		{"", false},
		{"1", false},
		{"6", false},
		{"9", false},
		{"A", false},
		{"B", false},
		{"F", false},
	}
	for _, c := range cases {
		got := isExtremelyDenseSubtype(c.subtype)
		if got != c.want {
			t.Errorf("isExtremelyDenseSubtype(%q): got %v, want %v", c.subtype, got, c.want)
		}
	}
}
