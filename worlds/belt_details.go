package worlds

import (
	"wbh/roller"
)

// BeltDetails — Size-0 body planetoid belt characteristics, WBH pp. 72-74.
type BeltDetails struct {
	Span           float64         // Orbit#s
	Composition    BeltComposition // m/s/c-type %
	Bulk           int
	ResourceRating int
	SigSize1Bodies int
	SigSizeSBodies int
	Profile        string // "S-CC.CC.CC.CC-B-R-#-s"
}

// BeltComposition — m/s/c/other type percentages summing to 100.
type BeltComposition struct {
	MTypePct int // metallic
	STypePct int // stony
	CTypePct int // carbonaceous/icy
	OtherPct int // peculiar / artificial / leftover
}

// RollBeltSpan computes belt span per WBH p.72: Spread × (2D) / 10.
// dms applies DM-1 (adjacent slot is a gas giant) or DM+3 (outermost slot).
// The effective roll is clamped to a minimum of 1.
func RollBeltSpan(r roller.Roller, spreadOrbits float64, dms int) (float64, error) {
	roll := r.Roll("2D")
	effective := roll + dms
	if effective < 1 {
		effective = 1
	}
	return spreadOrbits * float64(effective) / 10.0, nil
}

// beltCompositionRow encodes one row of the Belt Composition Percentages table
// from WBH p.73. Each component (m, s, c) carries a base value and a
// die-multiplier. Special cases:
//
//   - isD3=true: roll D3 and add base (base is 0 for a bare D3 cell).
//   - mult=0, isD3=false: flat constant equal to base (no die, may be 0).
//   - mult=K, isD3=false: roll 1D and return base + roll×K.
type beltCompositionRow struct {
	mBase, mMult int
	sBase, sMult int
	cBase, cMult int
	mIsD3        bool
	sIsD3        bool
	cIsD3        bool
}

// beltCompositionTable is indexed by 2D+DM, clamped to [0, 12].
// Transcribed directly from WBH p.73 "Belt Composition Percentages" table.
//
// Verified against book image (p.73):
//
//	Row 0-:  m=60+1D×5  s=1D×5     c=0
//	Row 1:   m=50+1D×5  s=5+1D×5   c=D3
//	Row 2:   m=40+1D×5  s=15+1D×5  c=1D
//	Row 3:   m=25+1D×5  s=30+1D×5  c=1D
//	Row 4:   m=15+1D×5  s=35+1D×5  c=5+1D
//	Row 5:   m=5+1D×5   s=40+1D×5  c=5+1D×2
//	Row 6:   m=1D×5     s=40+1D×5  c=1D×5
//	Row 7:   m=5+1D×2   s=35+1D×5  c=10+1D×5
//	Row 8:   m=5+1D     s=30+1D×5  c=20+1D×5
//	Row 9:   m=1D       s=15+1D×5  c=40+1D×5
//	Row 10:  m=1D       s=5+1D×5   c=50+1D×5
//	Row 11:  m=D3       s=5+1D×2   c=60+1D×5
//	Row 12+: m=0        s=1D       c=70+1D×5
var beltCompositionTable = [13]beltCompositionRow{
	// 0-: m=60+1D×5, s=1D×5, c=0
	{mBase: 60, mMult: 5, sBase: 0, sMult: 5, cBase: 0, cMult: 0},
	// 1: m=50+1D×5, s=5+1D×5, c=D3
	{mBase: 50, mMult: 5, sBase: 5, sMult: 5, cIsD3: true},
	// 2: m=40+1D×5, s=15+1D×5, c=1D
	{mBase: 40, mMult: 5, sBase: 15, sMult: 5, cBase: 0, cMult: 1},
	// 3: m=25+1D×5, s=30+1D×5, c=1D
	{mBase: 25, mMult: 5, sBase: 30, sMult: 5, cBase: 0, cMult: 1},
	// 4: m=15+1D×5, s=35+1D×5, c=5+1D
	{mBase: 15, mMult: 5, sBase: 35, sMult: 5, cBase: 5, cMult: 1},
	// 5: m=5+1D×5, s=40+1D×5, c=5+1D×2
	{mBase: 5, mMult: 5, sBase: 40, sMult: 5, cBase: 5, cMult: 2},
	// 6: m=1D×5, s=40+1D×5, c=1D×5   (confirmed: cBase=0, not 1)
	{mBase: 0, mMult: 5, sBase: 40, sMult: 5, cBase: 0, cMult: 5},
	// 7: m=5+1D×2, s=35+1D×5, c=10+1D×5
	{mBase: 5, mMult: 2, sBase: 35, sMult: 5, cBase: 10, cMult: 5},
	// 8: m=5+1D, s=30+1D×5, c=20+1D×5
	{mBase: 5, mMult: 1, sBase: 30, sMult: 5, cBase: 20, cMult: 5},
	// 9: m=1D, s=15+1D×5, c=40+1D×5
	{mBase: 0, mMult: 1, sBase: 15, sMult: 5, cBase: 40, cMult: 5},
	// 10: m=1D, s=5+1D×5, c=50+1D×5
	{mBase: 0, mMult: 1, sBase: 5, sMult: 5, cBase: 50, cMult: 5},
	// 11: m=D3, s=5+1D×2, c=60+1D×5
	{mIsD3: true, sBase: 5, sMult: 2, cBase: 60, cMult: 5},
	// 12+: m=0 (flat), s=1D, c=70+1D×5
	{mBase: 0, mMult: 0, sBase: 0, sMult: 1, cBase: 70, cMult: 5},
}

// rollComponent resolves one cell of the Belt Composition Percentages table.
//
//   - isD3=true: result = base + D3
//   - mult=0, isD3=false: result = base (flat constant, no die)
//   - mult=K, isD3=false: result = base + 1D×K
func rollComponent(r roller.Roller, base, mult int, isD3 bool) int {
	if isD3 {
		return base + r.Roll("D3")
	}
	if mult == 0 {
		return base
	}
	return base + r.Roll("1D")*mult
}

// RollBeltComposition rolls on the Belt Composition Percentages table per
// WBH p.73. dms applies DM-4 (inside HZCO) or DM+4 (beyond HZCO+2).
//
// The 2D+DM result is clamped to [0, 12] for table lookup. After rolling
// each component, overflows above 100% are removed first from m-type then
// s-type; any shortfall below 100% is allocated as "other".
func RollBeltComposition(r roller.Roller, dms int) (BeltComposition, error) {
	roll := r.Roll("2D")
	idx := roll + dms
	if idx < 0 {
		idx = 0
	}
	if idx > 12 {
		idx = 12
	}
	row := beltCompositionTable[idx]

	m := rollComponent(r, row.mBase, row.mMult, row.mIsD3)
	s := rollComponent(r, row.sBase, row.sMult, row.sIsD3)
	c := rollComponent(r, row.cBase, row.cMult, row.cIsD3)

	// WBH p.73: "If the total of m-, s-, and t-types exceed 100%, remove any
	// excess % first from m-type, then from s-type."
	total := m + s + c
	if total > 100 {
		over := total - 100
		if m >= over {
			m -= over
		} else {
			over -= m
			m = 0
			if s >= over {
				s -= over
			} else {
				s = 0
			}
		}
	}

	// Any shortfall below 100% is "other" composition.
	other := 100 - (m + s + c)
	if other < 0 {
		other = 0
	}

	return BeltComposition{MTypePct: m, STypePct: s, CTypePct: c, OtherPct: other}, nil
}
