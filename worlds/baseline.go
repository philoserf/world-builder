package worlds

import (
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
