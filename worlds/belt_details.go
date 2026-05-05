package worlds

import (
	"fmt"

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
	effective := max(roll+dms, 1)
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
	idx := max(roll+dms, 0)
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
	other := max(100-(m+s+c), 0)

	return BeltComposition{MTypePct: m, STypePct: s, CTypePct: c, OtherPct: other}, nil
}

// RollBeltBulk implements WBH p.73 Belt Bulk = 2D+2 + DMs.
// System Age contributes DM = -(Gyr ÷ 2 floor); composition contributes DM = +(cType% ÷ 10 floor).
// Result < 1 is clamped to 1.
//
// NOTE: WBH dice notation "2D2" means 2D + 2 (range 4-14). The correct
// notation for this parser is "2D+2", which Roll returns as the post-modifier sum.
func RollBeltBulk(r roller.Roller, ageGyr float64, comp BeltComposition) (int, error) {
	roll := r.Roll("2D+2")
	dms := -int(ageGyr/2) + comp.CTypePct/10
	bulk := max(roll+dms, 1)
	return bulk, nil
}

// RollResourceRating implements WBH p.73: 2D-7 + Bulk + (mType%÷10) - (cType%÷10).
// Rating < 2 is treated as 2; rating > 12 is capped at 12.
func RollResourceRating(r roller.Roller, bulk int, comp BeltComposition) (int, error) {
	roll := r.Roll("2D")
	rating := max(roll-7+bulk+comp.MTypePct/10-comp.CTypePct/10, 2)
	if rating > 12 {
		rating = 12
	}
	return rating, nil
}

// RollSigSize1Bodies implements WBH p.73: 2D-12 + Bulk + DMs.
// DMs: beltOrbit ≥ hzco+3 → DM+2; span < 0.1 → DM-4.
// Negative results are clamped to 0.
func RollSigSize1Bodies(r roller.Roller, bulk int, beltOrbit, hzco, span float64) (int, error) {
	roll := r.Roll("2D")
	dms := 0
	if beltOrbit >= hzco+3 {
		dms += 2
	}
	if span < 0.1 {
		dms -= 4
	}
	count := max(roll-12+bulk+dms, 0)
	return count, nil
}

// RollSigSizeSBodies implements WBH p.73: 2D-10 + (DM+1) × (Bulk+1).
// DMs: hzco+2 ≤ beltOrbit < hzco+3 → DM+1; beltOrbit ≥ hzco+3 → DM+3; span > 1.0 → DM+1.
// Negative results are clamped to 0. If span < 0.1, result is halved (round up).
func RollSigSizeSBodies(r roller.Roller, bulk int, beltOrbit, hzco, span float64) (int, error) {
	roll := r.Roll("2D")
	dm := 0
	if beltOrbit >= hzco+2 && beltOrbit < hzco+3 {
		dm++
	}
	if beltOrbit >= hzco+3 {
		dm += 3
	}
	if span > 1.0 {
		dm++
	}
	count := max(roll-10+(dm+1)*(bulk+1), 0)
	if span < 0.1 {
		count = (count + 1) / 2
	}
	return count, nil
}

// GenerateBeltDetails orchestrates the per-belt pipeline per WBH pp.72-74.
// Caller supplies pre-resolved geometric parameters (orbit, spread, hzco, age) and
// neighbor flags (adjacentToGG triggers DM-1 on span; outermostSlot triggers DM+3).
func GenerateBeltDetails(r roller.Roller, beltOrbit, spreadOrbits, hzco, ageGyr float64, adjacentToGG, outermostSlot bool) (BeltDetails, error) {
	spanDM := 0
	if adjacentToGG {
		spanDM = -1
	}
	if outermostSlot {
		spanDM = 3
	}
	span, err := RollBeltSpan(r, spreadOrbits, spanDM)
	if err != nil {
		return BeltDetails{}, err
	}

	compDM := 0
	if beltOrbit < hzco {
		compDM = -4
	}
	if beltOrbit > hzco+2 {
		compDM = 4
	}
	comp, err := RollBeltComposition(r, compDM)
	if err != nil {
		return BeltDetails{}, err
	}

	bulk, err := RollBeltBulk(r, ageGyr, comp)
	if err != nil {
		return BeltDetails{}, err
	}

	resources, err := RollResourceRating(r, bulk, comp)
	if err != nil {
		return BeltDetails{}, err
	}

	size1, err := RollSigSize1Bodies(r, bulk, beltOrbit, hzco, span)
	if err != nil {
		return BeltDetails{}, err
	}

	sizeS, err := RollSigSizeSBodies(r, bulk, beltOrbit, hzco, span)
	if err != nil {
		return BeltDetails{}, err
	}

	belt := BeltDetails{
		Span: span, Composition: comp, Bulk: bulk,
		ResourceRating: resources, SigSize1Bodies: size1, SigSizeSBodies: sizeS,
	}
	belt.Profile = FormatBeltProfile(belt)
	return belt, nil
}

// FormatBeltProfile renders the belt-profile shorthand "S-CC.CC.CC.CC-B-R-#-s" per WBH p.74.
// Resource rating uses hex letters A/B/C for 10/11/12.
func FormatBeltProfile(b BeltDetails) string {
	resourceStr := fmt.Sprintf("%d", b.ResourceRating)
	switch b.ResourceRating {
	case 10:
		resourceStr = "A"
	case 11:
		resourceStr = "B"
	case 12:
		resourceStr = "C"
	}
	return fmt.Sprintf(
		"%g-%02d.%02d.%02d.%02d-%d-%s-%d-%d",
		b.Span,
		b.Composition.MTypePct, b.Composition.STypePct,
		b.Composition.CTypePct, b.Composition.OtherPct,
		b.Bulk,
		resourceStr,
		b.SigSize1Bodies, b.SigSizeSBodies,
	)
}
