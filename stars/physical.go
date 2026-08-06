package stars

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// gridKeys lists the WBH-tabulated subtype grid in book order.
var gridKeys = []string{
	"O0", "O5",
	"B0", "B5",
	"A0", "A5",
	"F0", "F5",
	"G0", "G5",
	"K0", "K5",
	"M0", "M5", "M9",
}

// letterOrder maps each spectral letter to its position in book order.
var letterOrder = map[SpectralLetter]int{
	'O': 0, 'B': 1, 'A': 2, 'F': 3, 'G': 4, 'K': 5, 'M': 6,
}

// gridIndex returns (lowerIdx, upperIdx, fraction) for interpolation.
// Treats the grid as a flat ordered sequence: O0 < O5 < B0 < ... < M9.
// Subtypes between two grid points yield a fraction in [0, 1].
// Exact grid points yield (idx, idx, 0).
func gridIndex(st SpectralType) (lower, upper int, frac float64, err error) {
	lo, ok := letterOrder[st.Letter]
	if !ok {
		return 0, 0, 0, fmt.Errorf("stars: unsupported letter: %c", st.Letter)
	}

	target := lo*10 + st.Subtype

	// Compute numeric position of every grid key.
	positions := make([]int, len(gridKeys))
	for i, k := range gridKeys {
		positions[i] = letterOrder[SpectralLetter(k[0])]*10 + int(k[1]-'0')
	}

	for i, p := range positions {
		if p == target {
			return i, i, 0, nil
		}
	}

	lowerIdx, upperIdx := -1, -1

	for i, p := range positions {
		if p < target {
			lowerIdx = i
		}

		if p > target && upperIdx == -1 {
			upperIdx = i
		}
	}

	if lowerIdx == -1 || upperIdx == -1 {
		return 0, 0, 0, fmt.Errorf("stars: spectral type out of grid range: %s", st)
	}

	span := positions[upperIdx] - positions[lowerIdx]

	return lowerIdx, upperIdx, float64(target-positions[lowerIdx]) / float64(span), nil
}

// InterpolateClassRow interpolates a class-keyed quantity (Mass, Diameter,
// Luminosity) from a ClassRow table. When one bracketing row lacks the
// requested class (a book "—" cell at the table boundary, e.g. K5 has
// no Class IV cell), the function falls back to the other endpoint —
// the table is treated as truncated there. Mirrors MAO()'s one-sided-
// missing policy. Both-missing still returns an error.
func InterpolateClassRow(table map[string]ClassRow, st SpectralType, lc LuminosityClass) (float64, error) {
	lo, hi, frac, err := gridIndex(st)
	if err != nil {
		return 0, err
	}

	loRow, ok := table[gridKeys[lo]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[lo])
	}

	loVal, loOK := loRow.Get(lc)
	if lo == hi {
		if !loOK {
			return 0, fmt.Errorf("stars: %s class %s missing", gridKeys[lo], lc)
		}

		return loVal, nil
	}

	hiRow, ok := table[gridKeys[hi]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[hi])
	}

	hiVal, hiOK := hiRow.Get(lc)
	switch {
	case !loOK && !hiOK:
		return 0, fmt.Errorf("stars: %s class %s missing", gridKeys[lo], lc)
	case !loOK:
		return hiVal, nil
	case !hiOK:
		return loVal, nil
	}

	return loVal + frac*(hiVal-loVal), nil
}

// InterpolateScalar interpolates a scalar-valued quantity (Temperature)
// across the spectral grid.
func InterpolateScalar(table map[string]float64, st SpectralType) (float64, error) {
	lo, hi, frac, err := gridIndex(st)
	if err != nil {
		return 0, err
	}

	loVal, ok := table[gridKeys[lo]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[lo])
	}

	if lo == hi {
		return loVal, nil
	}

	hiVal, ok := table[gridKeys[hi]]
	if !ok {
		return 0, fmt.Errorf("stars: no row for %s", gridKeys[hi])
	}

	return loVal + frac*(hiVal-loVal), nil
}

// ComputeMass returns the interpolated mass in solar units (WBH p. 17).
func ComputeMass(st SpectralType, lc LuminosityClass) (float64, error) {
	return InterpolateClassRow(StarMass, st, lc)
}

// ComputeDiameter returns the interpolated diameter in solar units (WBH p. 19).
func ComputeDiameter(st SpectralType, lc LuminosityClass) (float64, error) {
	return InterpolateClassRow(StarDiameter, st, lc)
}

// ComputeTemperature returns the interpolated surface temperature in
// Kelvin (WBH p. 17 — Temperature column).
func ComputeTemperature(st SpectralType) (float64, error) {
	return InterpolateScalar(StarTemperature, st)
}

// SolTemperatureK is the WBH p. 20 reference temperature for Sol.
const SolTemperatureK = 5772.0

// ComputeLuminosityFromTable returns the interpolated luminosity in
// solar units (WBH p. 19).
func ComputeLuminosityFromTable(st SpectralType, lc LuminosityClass) (float64, error) {
	return InterpolateClassRow(StarLuminosity, st, lc)
}

// ComputeLuminosityFromFormula returns the closed-form luminosity in
// solar units (WBH p. 20):
//
//	L/L⊙ = (D/D⊙)^2 × (T/T⊙)^4
//
// The book recommends this formula when accurate diameter and
// temperature values are available.
func ComputeLuminosityFromFormula(diameter, temperature float64) float64 {
	tRatio := temperature / SolTemperatureK

	return diameter * diameter * tRatio * tRatio * tRatio * tRatio
}

// ApplyVariance applies the WBH optional variance roll (p. 17–19).
//
// The roll is 2D-7 (range -5 to +5). The result scales linearly between
// -maxPct and +maxPct of the base value:
//
//	adjusted = base × (1 + (2D-7)/5 × maxPct)
//
// Use maxPct=0.20 for mass/diameter, 0.30 for luminosity (Class III/V).
func ApplyVariance(base float64, r roller.Roller, maxPct float64) float64 {
	deviation := r.Roll("2D-7")
	factor := 1.0 + (float64(deviation)/5.0)*maxPct

	return base * factor
}
