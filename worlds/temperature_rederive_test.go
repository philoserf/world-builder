package worlds

import (
	"testing"
)

func TestSelectExoticLiquid_Water_Terra(t *testing.T) {
	// At meanK=288 with atm A (10), water (Abundance 100) wins among candidates
	// where 273 ≤ 288 ≤ 373.
	got := SelectExoticLiquid(288, 10)
	if got != "H2O" {
		t.Errorf("got %q, want H2O", got)
	}
}

func TestSelectExoticLiquid_Methane_Cold(t *testing.T) {
	// At meanK=100 with atm B (11), methane (range 91-113, Abundance 70) wins.
	got := SelectExoticLiquid(100, 11)
	if got != "CH4" {
		t.Errorf("got %q, want CH4", got)
	}
}

func TestSelectExoticLiquid_Ethane_NotMethaneAtMid(t *testing.T) {
	// At meanK=150, methane (boils at 113) is out; ethane (range 90-184, Abundance 70) wins.
	got := SelectExoticLiquid(150, 11)
	if got != "C2H6" {
		t.Errorf("got %q, want C2H6", got)
	}
}

func TestSelectExoticLiquid_NoCandidate_TooHot(t *testing.T) {
	// At meanK=2000, all candidates' boiling points are exceeded.
	got := SelectExoticLiquid(2000, 10)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestSelectExoticLiquid_NonExoticAtm_Empty(t *testing.T) {
	// Atm 6 (standard) → defensive: caller shouldn't call; return empty.
	got := SelectExoticLiquid(288, 6)
	if got != "" {
		t.Errorf("got %q, want empty (non-exotic atm)", got)
	}
}

func TestSelectExoticLiquid_TieBreakLowerBoiling(t *testing.T) {
	// Construct a meanK where two candidates tie on Abundance — verify lower
	// BoilingK wins. CH4 (91-113, 70) and C2H6 (90-184, 70) both contain 100K.
	// Lower BoilingK is CH4 (113 < 184) → CH4 wins.
	got := SelectExoticLiquid(100, 11)
	if got != "CH4" {
		t.Errorf("got %q, want CH4 (tie-break by lower BoilingK)", got)
	}
}

func TestSelectExoticLiquid_BoundaryInclusive(t *testing.T) {
	// Verify [MeltingK, BoilingK] is inclusive on both ends.
	// H2O range: melting 273, boiling 373.
	if got := SelectExoticLiquid(273, 10); got != "H2O" {
		t.Errorf("meanK=273 (= H2O melting): got %q, want H2O", got)
	}
	if got := SelectExoticLiquid(373, 10); got != "H2O" {
		t.Errorf("meanK=373 (= H2O boiling): got %q, want H2O", got)
	}
}

func TestMeanKToTempRange_Boundaries(t *testing.T) {
	cases := []struct {
		meanK float64
		want  TempRange
	}{
		{50, TempFrozen},
		{122, TempFrozen},
		{123, TempCold},
		{200, TempCold},
		{272, TempCold},
		{273, TempTemperate},
		{300, TempTemperate},
		{352, TempTemperate},
		{353, TempHot},
		{400, TempHot},
		{452, TempHot},
		{453, TempBoiling},
		{1000, TempBoiling},
	}
	for _, c := range cases {
		if got := MeanKToTempRange(c.meanK); got != c.want {
			t.Errorf("meanK=%v: got %v, want %v", c.meanK, got, c.want)
		}
	}
}
