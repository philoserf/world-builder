package worlds

import (
	"fmt"

	"wbh/roller"
)

// Hydrographics — surface liquid coverage per WBH p.99.
//
// Code is populated by 3A1 with HZCO-bucketed provisional temperature; Step 5D
// (3A2b-rederive) re-derives Code under real Temperature.MeanK and populates
// Profile (composition tail). Post-5D values are final.
// Format: "H<code>:<liquid>-100" (e.g., "H6:H2O-100", "H4:CH4-100").
// Empty for vacuum or zero-hydrographics bodies.
type Hydrographics struct {
	Code         int
	PercentRange [2]int
	Percent      int
	Profile      string // 3A2b-rederive composition tail
}

// RollHydroDigit computes the hydro digit: 2D-7 + atmoCode + DMs, capped [0, 10].
//
// DMs:
//   - Size 0 or 1 → DM-4
//   - Atmo 0, 1, or ≥10 → DM-4
//   - Hot → DM-2; Boiling → DM-6 (both ignored when atmo is D (13) or panthalassic F/7 (15/"7"))
func RollHydroDigit(r roller.Roller, atmoCode int, atmoSubtype string, sizeCode SizeCode, tempRange TempRange) (int, error) {
	roll := r.Roll("2D")
	dm := 0

	si := SizeAsInt(sizeCode)
	if si == 0 || si == 1 {
		dm -= 4
	}

	if atmoCode == 0 || atmoCode == 1 || atmoCode >= 10 {
		dm -= 4
	}

	skipTempDM := atmoCode == 13 || (atmoCode == 15 && atmoSubtype == "7")
	if !skipTempDM {
		switch tempRange {
		case TempHot:
			dm -= 2
		case TempBoiling:
			dm -= 6
		}
	}

	digit := roll - 7 + atmoCode + dm
	if digit < 0 {
		digit = 0
	}
	if digit > 10 {
		digit = 10
	}
	return digit, nil
}

// HydroRange returns the percent range [lo, hi] for a hydro digit per WBH p.99.
func HydroRange(digit int) [2]int {
	switch digit {
	case 0:
		return [2]int{0, 5}
	case 1:
		return [2]int{6, 15}
	case 2:
		return [2]int{16, 25}
	case 3:
		return [2]int{26, 35}
	case 4:
		return [2]int{36, 45}
	case 5:
		return [2]int{46, 55}
	case 6:
		return [2]int{56, 65}
	case 7:
		return [2]int{66, 75}
	case 8:
		return [2]int{76, 85}
	case 9:
		return [2]int{86, 95}
	case 10:
		return [2]int{96, 100}
	}
	return [2]int{0, 0}
}

// RefineHydroPercent refines hydro percent within a range via d10 variance per WBH p.99.
//
//   - digit 0:  pct = -4 + d10, floor at 0
//   - digit 10: pct = 96 + d10, cap at 100
//   - other:    pct = hydroRange[0] + (d10 - 1), cap at hydroRange[1]
func RefineHydroPercent(r roller.Roller, digit int, hydroRange [2]int) (int, error) {
	v := r.Roll("d10")
	switch digit {
	case 0:
		pct := -4 + v
		if pct < 0 {
			pct = 0
		}
		return pct, nil
	case 10:
		pct := 96 + v
		if pct > 100 {
			pct = 100
		}
		return pct, nil
	}
	pct := hydroRange[0] + (v - 1)
	if pct > hydroRange[1] {
		pct = hydroRange[1]
	}
	return pct, nil
}

// GenerateHydrographics orchestrates digit → range → percent variance per WBH p.99.
func GenerateHydrographics(r roller.Roller, atmo Atmosphere, sizeCode SizeCode, tempRange TempRange) (Hydrographics, error) {
	digit, err := RollHydroDigit(r, atmo.Code, atmo.Subtype, sizeCode, tempRange)
	if err != nil {
		return Hydrographics{}, err
	}
	rng := HydroRange(digit)
	pct, err := RefineHydroPercent(r, digit, rng)
	if err != nil {
		return Hydrographics{}, err
	}
	return Hydrographics{Code: digit, PercentRange: rng, Percent: pct}, nil
}

// hydroCodeChar renders a hydrographics code as its UWP character: 0-9 → "0".."9", 10 → "A".
func hydroCodeChar(code int) string {
	if code <= 9 {
		return fmt.Sprintf("%d", code)
	}
	return "A"
}

// DeriveHydrographicsProfile returns the composition tail for a body with
// the given mean temperature, atmosphere code, and hydrographics code.
//
// Format: "H<code>:<liquid>-100" — single-dominant-liquid form per spec.
// Empty for vacuum (atmCode=0), zero hydrographics (hydroCode=0), or
// when no exotic liquid fits at meanK.
//
// Selection:
//   - Atm 2-9, D, E with hydro > 0: liquid is "H2O" (water default)
//   - Atm A-C, F (10-12, 15) with hydro > 0: SelectExoticLiquid(meanK, atmCode)
//   - Otherwise: empty
func DeriveHydrographicsProfile(meanK float64, atmCode, hydroCode int) string {
	if hydroCode <= 0 || atmCode == 0 {
		return ""
	}
	var liquid string
	if isExoticAtmCode(atmCode) {
		liquid = SelectExoticLiquid(meanK, atmCode)
	} else {
		liquid = "H2O"
	}
	if liquid == "" {
		return ""
	}
	return "H" + hydroCodeChar(hydroCode) + ":" + liquid + "-100"
}
