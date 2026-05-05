// Package worlds — atmosphere/hydrographics re-derivation under real
// temperature per WBH p.79, p.81, pp.94-98, p.99, p.102 (sub-project 3A2b-rederive).
package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// MeanKToTempRange buckets a real mean temperature in Kelvin into the same
// TempRange bands 3A1's HZCOOffsetToTempRange used (WBH pp.94-98 keying):
//
//	≥ 453K → Boiling
//	353-453K → Hot
//	273-353K → Temperate
//	123-273K → Cold
//	< 123K → Frozen
func MeanKToTempRange(meanK float64) TempRange {
	switch {
	case meanK >= 453:
		return TempBoiling
	case meanK >= 353:
		return TempHot
	case meanK >= 273:
		return TempTemperate
	case meanK >= 123:
		return TempCold
	default:
		return TempFrozen
	}
}

// RederiveAtmosphereHydrographics re-derives 3A1's temperature-sensitive
// Atmosphere/Hydrographics fields under the body's current Temperature.MeanK.
// Mutates body in place. Called twice as part of Step 5D's 2-pass iteration.
//
// Currently mutates (Task 6 baseline; later tasks add more):
//   - Atmosphere.ScaleHeight   (re-derived from current MeanK + gravity via DeriveScaleHeight)
//   - Hydrographics.Code       (re-rolled with current TempRange's Hot/Boiling DMs)
//   - Hydrographics.Profile    (composition tail per p.102)
//
// No-op when body.Body == BodyEmpty or body.Temperature == nil.
//
// Pending in subsequent tasks of this sub-project:
//   - Atm.Subtype + .Pressure re-roll for B/C atm (Task 7 helper, Task 9 wiring)
//   - Atm.Profile re-derive via RollGasMix for exotic atm (Task 8)
//   - CheckRunawayGreenhouse integration with hydro DM-6 override (Task 9)
func RederiveAtmosphereHydrographics(
	r roller.Roller,
	body *DetailedPlacement,
	sys stars.System,
	parent *DetailedPlacement,
) error {
	if body.Body == BodyEmpty || body.Temperature == nil {
		return nil
	}

	meanK := body.Temperature.MeanK
	tempRange := MeanKToTempRange(meanK)

	// 1. Atmosphere.ScaleHeight: re-derive from real meanK + gravity.
	if body.Atmosphere != nil && body.Physical != nil {
		body.Atmosphere.ScaleHeight = DeriveScaleHeight(meanK, body.Physical.Gravity)
	}

	// 2. Hydrographics.Code: re-roll with current TempRange's Hot/Boiling DMs.
	if body.Atmosphere != nil && body.Hydrographics != nil {
		newHydro, err := RollHydroDigit(r, body.Atmosphere.Code, body.Atmosphere.Subtype, body.SizeCode, tempRange)
		if err != nil {
			return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: hydro re-roll: %w", err)
		}
		body.Hydrographics.Code = newHydro
	}

	// 3. Hydrographics.Profile: derive from current code + atm + meanK.
	if body.Hydrographics != nil {
		atmCode := 0
		if body.Atmosphere != nil {
			atmCode = body.Atmosphere.Code
		}
		body.Hydrographics.Profile = DeriveHydrographicsProfile(meanK, atmCode, body.Hydrographics.Code)
	}

	_ = parent // parent unused at Task 6; wired in Task 9 for moon-specific paths
	_ = sys    // sys unused at Task 6; wired in Task 9 for runaway greenhouse
	return nil
}
