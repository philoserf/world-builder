package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
)

// SurfaceDistribution — landmass concentration per WBH p.100.
type SurfaceDistribution struct {
	Code        int                  // 0..10 (clamped from 2D-2 with -1 → 0, 11+ → 10)
	Description string               // "Extremely Dispersed"|...|"Extremely Concentrated"
	Geography   FundamentalGeography // Ocean (water major) | Land (land major)
}

// FundamentalGeography indicates whether the world's major bodies are water or land.
type FundamentalGeography int

// FundamentalGeography constants — water-major (Ocean) or land-major (Land).
const (
	GeographyOcean FundamentalGeography = iota // hydro 6+, or hydro 5 with 1D 1-3
	GeographyLand                              // hydro 4-, or hydro 5 with 1D 4-6
)

// surfaceDescriptions maps the 11-row Surface Distribution table per WBH p.100.
var surfaceDescriptions = [11]string{
	0:  "Extremely Dispersed",
	1:  "Very Dispersed",
	2:  "Dispersed",
	3:  "Scattered",
	4:  "Slightly Scattered",
	5:  "Mixed",
	6:  "Slightly Skewed",
	7:  "Skewed",
	8:  "Concentrated",
	9:  "Very Concentrated",
	10: "Extremely Concentrated",
}

// RollSurfaceDistribution rolls 2D-2 and clamps to [0, 10] per WBH p.100.
func RollSurfaceDistribution(r roller.Roller) (int, error) {
	twoD := r.Roll("2D")
	code := min(max(twoD-2, 0), 10)
	return code, nil
}

// DescribeSurfaceDistribution maps Code → Description from the p.100 table.
// Out-of-range codes return "Extremely Dispersed" (0) or "Extremely Concentrated" (10).
func DescribeSurfaceDistribution(code int) string {
	if code < 0 {
		code = 0
	}
	if code > 10 {
		code = 10
	}
	return surfaceDescriptions[code]
}

// DetermineFundamentalGeography per WBH p.101:
//
//	Hydrographics 6+ → Ocean
//	Hydrographics 4- → Land
//	Hydrographics 5  → 1D, 1-3 → Ocean, 4-6 → Land
func DetermineFundamentalGeography(r roller.Roller, hydroCode int) (FundamentalGeography, error) {
	switch {
	case hydroCode >= 6:
		return GeographyOcean, nil
	case hydroCode <= 4:
		return GeographyLand, nil
	case hydroCode == 5:
		oneD := r.Roll("1D")
		if oneD <= 3 {
			return GeographyOcean, nil
		}
		return GeographyLand, nil
	}
	// Unreachable: the switch above is exhaustive over all integers
	// (hydroCode >= 6 || hydroCode <= 4 || hydroCode == 5). The error
	// return is preserved for signature consistency with peer Roll*/Generate*
	// helpers in this package.
	return GeographyLand, fmt.Errorf("worlds: DetermineFundamentalGeography: invalid hydroCode %d", hydroCode)
}

// GenerateSurfaceDistribution orchestrates the per-body pipeline. Returns nil
// (no error) for bodies without Hydrographics — Surface Feature Distribution
// is meaningless for vacuum worlds and gas giants.
func GenerateSurfaceDistribution(r roller.Roller, hydro *Hydrographics) (*SurfaceDistribution, error) {
	if hydro == nil {
		return nil, nil
	}
	code, err := RollSurfaceDistribution(r)
	if err != nil {
		return nil, fmt.Errorf("worlds: GenerateSurfaceDistribution: %w", err)
	}
	geography, err := DetermineFundamentalGeography(r, hydro.Code)
	if err != nil {
		return nil, fmt.Errorf("worlds: GenerateSurfaceDistribution: %w", err)
	}
	return &SurfaceDistribution{
		Code:        code,
		Description: DescribeSurfaceDistribution(code),
		Geography:   geography,
	}, nil
}
