package worlds

import (
	"math"

	"wbh/roller"
	"wbh/stars"
)

// RollBaselineNumber implements WBH Step 2 (pp. 44-45). The baseline
// number determines whether the system is hot, temperate, or cold, and
// drives Step 3.
func RollBaselineNumber(r roller.Roller, sys stars.System, counts Counts) (int, error) {
	return r.Roll("2D") + baselineDMs(sys, counts), nil
}

// baselineDMs computes the WBH p. 45 DM stack for Step 2.
func baselineDMs(sys stars.System, counts Counts) int {
	dm := 0
	if primaryHasCompanion(sys) {
		dm -= 2
	}
	switch sys.Primary.LuminosityClass {
	case stars.Ia, stars.Ib, stars.II:
		dm += 3
	case stars.III:
		dm += 2
	case stars.IV:
		dm++
	case stars.VI:
		dm--
	}
	if isPostStellar(sys.Primary.Kind) {
		dm -= 2
	}
	switch {
	case counts.Total < 6:
		dm -= 4
	case counts.Total <= 9:
		dm -= 3
	case counts.Total <= 12:
		dm -= 2
	case counts.Total <= 15:
		dm--
	case counts.Total <= 17:
		// 16-17: unlisted band in book → 0
	case counts.Total <= 20:
		dm++
	default: // > 20
		dm += 2
	}
	dm -= secondaryStarCount(sys)
	return dm
}

// primaryHasCompanion reports whether sys has a Companion-class star
// directly orbiting the primary.
func primaryHasCompanion(sys stars.System) bool {
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
			return true
		}
	}
	return false
}

// secondaryStarCount returns the count of non-companion stars at Close,
// Near, or Far around the primary.
func secondaryStarCount(sys stars.System) int {
	n := 0
	for _, c := range sys.Companions {
		if c.ParentIndex != -1 {
			continue
		}
		switch c.OrbitClass {
		case stars.OrbitClose, stars.OrbitNear, stars.OrbitFar:
			n++
		}
	}
	return n
}

// BaselineOrbit implements WBH Step 3 (pp. 45-46). Selects the right
// formula by comparing baselineN to totalWorlds. Snaps the result to the
// nearest available orbit (with (2D-7)/10 direction variance) when the
// formula lands inside a primary-group exclusion zone.
//
// hzco is the primary group's HZCO (use primary.HZCO()).
// Continuation Method (sub-case 3d) is out of scope.
func BaselineOrbit(
	r roller.Roller,
	primary Group,
	hzco float64,
	baselineN, totalWorlds int,
) (float64, error) {
	var orbit float64
	switch {
	case baselineN >= 1 && baselineN <= totalWorlds:
		// Sub-case 3a.
		v := r.Roll("2D")
		if hzco >= 1.0 {
			orbit = hzco + float64(v-7)/10.0
		} else {
			orbit = hzco + float64(v-7)/100.0
		}
	case baselineN < 1:
		// Sub-case 3b. minOrbit = max(MAO, HZCO).
		minOrbit := primary.MAO
		if hzco > minOrbit {
			minOrbit = hzco
		}
		v := r.Roll("2D")
		if minOrbit >= 1.0 {
			orbit = hzco - float64(baselineN) + float64(totalWorlds) + float64(v-2)/10.0
		} else {
			orbit = minOrbit - float64(baselineN)/10.0 + float64(v-2)/100.0
		}
	default:
		// Sub-case 3c (baselineN > totalWorlds).
		v := r.Roll("2D")
		firstForm := hzco - float64(baselineN) + float64(totalWorlds)
		if firstForm >= 1.0 {
			orbit = firstForm + float64(v-7)/5.0
		} else {
			orbit = hzco - (float64(baselineN)+float64(totalWorlds)+float64(v-7)/5.0)/10.0
			if orbit < 0 {
				lower := primary.MAO + float64(totalWorlds)*0.01
				if hzco-0.1 > lower {
					orbit = hzco - 0.1
				} else {
					orbit = lower
				}
			}
		}
	}
	if !primary.Contains(orbit) {
		orbit = snapToAvailable(r, primary, orbit)
	}
	return orbit, nil
}

// snapToAvailable returns the nearest in-interval orbit to want, with
// (2D-7)/10 direction variance applied per the book p. 45 narrative.
func snapToAvailable(r roller.Roller, primary Group, want float64) float64 {
	if len(primary.Intervals) == 0 {
		return want
	}
	bestDist := math.Inf(1)
	var best float64
	for _, iv := range primary.Intervals {
		var snap float64
		switch {
		case want < iv.Min:
			snap = iv.Min
		case want > iv.Max:
			snap = iv.Max
		default:
			snap = want
		}
		if d := math.Abs(snap - want); d < bestDist {
			bestDist = d
			best = snap
		}
	}
	v := r.Roll("2D")
	return best + float64(v-7)/10.0
}
