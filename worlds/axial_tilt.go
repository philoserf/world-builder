package worlds

import (
	"fmt"

	"wbh/roller"
)

// AxialTilt — world's obliquity per WBH p.104.
type AxialTilt struct {
	Degrees         float64 // 0..180; >90 = retrograde
	Retrograde      bool    // 90 < tilt <= 180
	BaselineDegrees float64 // pre-lock value (preserved if 1:1 lock rerolled)
}

// RollBasicAxialTilt rolls 2D on the Basic Axial Tilt table per WBH p.104.
//
// Returns (degrees, isExtreme, err). On 2D >= 10, returns (0, true, nil)
// and the caller dispatches to RollExtremeAxialTilt without consuming
// further dice from this function.
func RollBasicAxialTilt(r roller.Roller) (degrees float64, isExtreme bool, err error) {
	twoD := r.Roll("2D")
	switch {
	case twoD >= 2 && twoD <= 4:
		// (1D-1) ÷ 50
		oneD := r.Roll("1D")
		return float64(oneD-1) / 50.0, false, nil
	case twoD == 5:
		// 1D ÷ 5
		oneD := r.Roll("1D")
		return float64(oneD) / 5.0, false, nil
	case twoD == 6:
		// 1D
		oneD := r.Roll("1D")
		return float64(oneD), false, nil
	case twoD == 7:
		// 6 + 1D
		oneD := r.Roll("1D")
		return float64(6 + oneD), false, nil
	case twoD == 8 || twoD == 9:
		// 5 + 1D × 5
		oneD := r.Roll("1D")
		return float64(5 + oneD*5), false, nil
	case twoD >= 10:
		return 0, true, nil
	}
	return 0, false, fmt.Errorf("worlds: RollBasicAxialTilt: unexpected 2D=%d", twoD)
}

// RollExtremeAxialTilt rolls 1D on the Extreme Axial Tilt table per WBH p.104.
//
//	1-2: 10 + 1D × 10  (20-70°)
//	3:   30 + 1D × 10  (40-90°)
//	4:   90 + 1D × 1D  (91-126°, retrograde — note 1D × 1D is two 1Ds multiplied)
//	5:   180 - 1D × 1D (144-179°, extreme retrograde)
//	6:   120 + 1D × 10 (130-180°, extreme retrograde with high variance)
func RollExtremeAxialTilt(r roller.Roller) (float64, error) {
	row := r.Roll("1D")
	switch row {
	case 1, 2:
		oneD := r.Roll("1D")
		return float64(10 + oneD*10), nil
	case 3:
		oneD := r.Roll("1D")
		return float64(30 + oneD*10), nil
	case 4:
		a := r.Roll("1D")
		b := r.Roll("1D")
		return float64(90 + a*b), nil
	case 5:
		a := r.Roll("1D")
		b := r.Roll("1D")
		return float64(180 - a*b), nil
	case 6:
		oneD := r.Roll("1D")
		return float64(120 + oneD*10), nil
	}
	return 0, fmt.Errorf("worlds: RollExtremeAxialTilt: unexpected row=%d", row)
}

// addAxialTiltPrecision adds linear-variance per WBH p.104:
//
//	"Linear variance for axial tilt values is appropriate and should be additive
//	 (using the result as a base) on the Extreme Axial Tilt table. Since
//	 sub-divisions of degrees are expressed as minutes and seconds, they can
//	 be added using the same procedures as day length variation."
//
// Implementation: add (0-59 degrees) + (0-59 arcminutes) using 1D-1 (tens) +
// d10 (ones) for each. Caller clamps the total to [0, 180]. Reuses
// tensOnesValue from day_length.go.
func addAxialTiltPrecision(r roller.Roller) float64 {
	extraDeg := min(
		// 0-59
		tensOnesValue(r), 59,
	)
	extraArcmin := min(
		// 0-59
		tensOnesValue(r), 59,
	)
	return float64(extraDeg) + float64(extraArcmin)/60.0
}

// GenerateAxialTilt orchestrates per-body axial tilt determination.
// Returns nil for empty bodies. Caps result at [0, 180].
//
// Linear-variance precision (extra degrees + arcminutes) is added only when
// the basic roll dispatched to the Extreme table per WBH p.104.
func GenerateAxialTilt(r roller.Roller, dp *DetailedPlacement) (*AxialTilt, error) {
	if dp.Body == BodyEmpty {
		return nil, nil
	}

	tilt, isExtreme, err := RollBasicAxialTilt(r)
	if err != nil {
		return nil, fmt.Errorf("worlds: GenerateAxialTilt: %w", err)
	}
	if isExtreme {
		tilt, err = RollExtremeAxialTilt(r)
		if err != nil {
			return nil, fmt.Errorf("worlds: GenerateAxialTilt: %w", err)
		}
		// Linear-variance precision applies to extreme-table results per p.104.
		tilt += addAxialTiltPrecision(r)
	}

	if tilt < 0 {
		tilt = 0
	}
	if tilt > 180 {
		tilt = 180
	}

	return &AxialTilt{
		Degrees:         tilt,
		Retrograde:      tilt > 90,
		BaselineDegrees: tilt,
	}, nil
}
