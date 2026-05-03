package worlds

import (
	"fmt"

	"wbh/roller"
)

// AnomalyType classifies a Step 7 anomalous slot.
type AnomalyType int

const (
	// AnomalyNone marks a slot that was placed by Steps 1-6 with no anomaly.
	AnomalyNone AnomalyType = iota
	// AnomalyRandom is a randomly-placed anomalous world (WBH p. 51, type ≤7).
	AnomalyRandom
	// AnomalyEccentric is a world with extreme eccentricity (type 8).
	AnomalyEccentric
	// AnomalyInclined is a world with a high orbital inclination (type 9).
	AnomalyInclined
	// AnomalyRetrograde is a retrograde-orbit world (type 10-11).
	AnomalyRetrograde
	// AnomalyTrojan is a Trojan-point world (type 12).
	AnomalyTrojan
)

// AnomalousSlot extends Slot with anomaly metadata. AnomalyNone slots
// are wrapped Slots with no extra behaviour; non-None slots come from
// Step 7 and feed Step 9 eccentricity DMs.
type AnomalousSlot struct {
	Slot
	Anomaly        AnomalyType
	InclinationDeg float64 // for AnomalyInclined
	TrojanOf       string  // StarSlot id of parent, for AnomalyTrojan
	EccentricityDM int     // additional DM applied in Step 9
}

// AddAnomalous implements WBH Step 7 (pp. 50-51). Wraps each input slot
// as AnomalousSlot{..., Anomaly: AnomalyNone} and appends one new
// AnomalousSlot per rolled anomaly. Updates counts (each anomaly adds
// one terrestrial and one total).
//
// In multi-star systems the parent group is picked by D3 (3-group max
// per WBH structural cap). In single-star systems no group roll is
// consumed; the only group is used.
//
// Anomalous orbits clamp to [group.MAO, 20.0] with retry. Trojan picks
// the immediately-inward slot in the parent group as the target.
func AddAnomalous(
	r roller.Roller,
	slots []Slot,
	allocs []StarAllocation,
	counts Counts,
) ([]AnomalousSlot, Counts, error) {
	out := make([]AnomalousSlot, 0, len(slots))
	for _, s := range slots {
		out = append(out, AnomalousSlot{Slot: s})
	}

	n := anomalousCount(r.Roll("2D"))
	if n == 0 {
		return out, counts, nil
	}

	for i := 0; i < n; i++ {
		atype := anomalousType(r.Roll("2D"))
		// Pick parent group (multi-star uses D3; single-star uses index 0).
		parentIdx := 0
		if len(allocs) > 1 {
			parentIdx = (r.Roll("D3") - 1) % len(allocs)
		}
		parent := allocs[parentIdx].Group

		ecdm := eccentricityDMFor(atype)
		var orbit float64
		var trojanOf string
		var inclination float64
		switch atype {
		case AnomalyTrojan:
			target := pickInwardSlot(slots, parent.Designation)
			trojanOf = target.StarSlot
			orbit = target.Orbit
		default:
			orbit = rollAnomalousOrbit(r, parent.MAO)
			if atype == AnomalyInclined {
				inclination = float64(r.Roll("1D")+2)*10.0 + float64(r.Roll("d10"))
			}
		}

		out = append(out, AnomalousSlot{
			Slot: Slot{
				StarSlot: fmt.Sprintf("%c+", parent.Designation[0]),
				Group:    parent,
				Orbit:    orbit,
			},
			Anomaly:        atype,
			InclinationDeg: inclination,
			TrojanOf:       trojanOf,
			EccentricityDM: ecdm,
		})

		counts.Terrestrials++
		counts.Total++
	}
	return out, counts, nil
}

func anomalousCount(roll int) int {
	switch {
	case roll <= 9:
		return 0
	case roll == 10:
		return 1
	case roll == 11:
		return 2
	default:
		return 3
	}
}

func anomalousType(roll int) AnomalyType {
	switch {
	case roll <= 7:
		return AnomalyRandom
	case roll == 8:
		return AnomalyEccentric
	case roll == 9:
		return AnomalyInclined
	case roll <= 11:
		return AnomalyRetrograde
	default:
		return AnomalyTrojan
	}
}

func eccentricityDMFor(t AnomalyType) int {
	switch t {
	case AnomalyEccentric:
		return 5
	case AnomalyRandom, AnomalyInclined, AnomalyRetrograde, AnomalyTrojan:
		return 2
	}
	return 0
}

// rollAnomalousOrbit rolls 2D-2 + d10/10, clamped to [mao, 20.0], with
// up to 5 retries before bailing to the clamp value.
func rollAnomalousOrbit(r roller.Roller, mao float64) float64 {
	for try := 0; try < 5; try++ {
		whole := r.Roll("2D") - 2
		frac := r.Roll("d10")
		v := float64(whole) + float64(frac)/10.0
		if v >= mao && v <= 20.0 {
			return v
		}
	}
	return mao
}

// pickInwardSlot returns the highest-Orbit slot in the parent group.
// For 2B we approximate the book's "immediately-inward" rule by picking
// the outermost slot in the parent group (since the trojan inherits its
// orbit, not derives a new one).
func pickInwardSlot(slots []Slot, parentDesignation string) Slot {
	var best Slot
	bestOrbit := -1.0
	for _, s := range slots {
		if s.Group.Designation == parentDesignation && s.Orbit > bestOrbit {
			best = s
			bestOrbit = s.Orbit
		}
	}
	return best
}
