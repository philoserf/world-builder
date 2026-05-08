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
