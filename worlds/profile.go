package worlds

import (
	"fmt"
	"strconv"
	"strings"
)

// ShortProfile renders the WBH p.58 short Planetary System Profile:
// "G-P-T-N-S" where G=gas giants, P=belts, T=terrestrials, N=baseline
// number (floored at 0), S=system spread (one decimal).
//
// Note: collides by name with stars.ShortProfile (Class 0/I survey).
// Different profiles for different purposes; package qualifier disambiguates.
func ShortProfile(sd SystemDetail) string {
	n := max(0, sd.BaselineN)
	return fmt.Sprintf(
		"%d-%d-%d-%d-%s",
		sd.Counts.GasGiants,
		sd.Counts.PlanetoidBelts,
		sd.Counts.Terrestrials,
		n,
		formatSpread(sd.SystemSpread),
	)
}

// LongProfile renders the WBH p.58 long Planetary System Profile:
// "St-N-W-W-W...-S:St-N-W-W...-S:..." per star.
//
// Slot codes (in orbit order): G (gas giant), P (belt), T (terrestrial).
// M, GM, PM reserved for mainworld variants; 2C never emits (no picker).
func LongProfile(sd SystemDetail) string {
	type starSegment struct {
		designation string
		baselineN   int
		codes       []string
	}
	segments := []starSegment{}
	groupIdx := map[string]int{}

	for _, alloc := range sd.Allocations {
		groupIdx[alloc.Group.Designation] = len(segments)
		segments = append(segments, starSegment{
			designation: alloc.Group.Designation,
			baselineN:   alloc.BaselineN,
		})
	}

	for _, dp := range sd.Detailed {
		gd := dp.Group.Designation
		idx, ok := groupIdx[gd]
		if !ok {
			continue
		}
		var code string
		switch dp.Body {
		case BodyTerrestrial:
			code = "T"
		case BodyGasGiant:
			code = "G"
		case BodyPlanetoidBelt:
			code = "P"
		case BodyEmpty:
			continue
		default:
			continue
		}
		segments[idx].codes = append(segments[idx].codes, code)
	}

	parts := make([]string, 0, len(segments))
	for _, seg := range segments {
		fields := []string{seg.designation, strconv.Itoa(seg.baselineN)}
		fields = append(fields, seg.codes...)
		fields = append(fields, formatSpread(sd.SystemSpread))
		parts = append(parts, strings.Join(fields, "-"))
	}
	return strings.Join(parts, ":")
}

// formatSpread renders a spread value to one decimal place.
func formatSpread(s float64) string {
	return strconv.FormatFloat(s, 'f', 1, 64)
}
