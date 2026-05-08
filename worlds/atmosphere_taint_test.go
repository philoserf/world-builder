package worlds

import (
	"testing"

	"wbh/roller"
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
