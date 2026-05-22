// Package worlds — atmosphere/hydrographics re-derivation under real
// temperature per WBH p.79, p.81, pp.94-98, p.99, p.102 (sub-project 3A2b-rederive).
package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
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
// Mutates body in place. Called twice as part of ApplyClimate's 2-pass iteration.
//
// Mutates:
//   - Atmosphere.ScaleHeight   (re-derived from current MeanK + gravity via DeriveScaleHeight)
//   - Atmosphere.Code          (mutated to A/B/C only on the runaway-greenhouse 2-9/D/E path;
//     atm A/B/C/F+ stay unchanged on a "boiling-only" runaway fire
//     per WBH p.79; HZ bodies only)
//   - Atmosphere.Subtype+Pressure (re-rolled with DM+4 only when atm code was actually mutated;
//     boiling-only fires preserve subtype and pressure)
//   - Hydrographics.Code       (re-rolled with TempBoiling DM-6 on any runaway fire — both
//     the mutation and boiling-only paths — else current TempRange)
//   - Hydrographics.Profile    (composition tail per p.102)
//   - Atmosphere.Profile       (gas mix for exotic A/B/C/F atm with hydro > 0)
//
// No-op when body.Kind == BodyEmpty or body.Temperature == nil.
func RederiveAtmosphereHydrographics(
	r roller.Roller,
	body *Body,
	sys stars.System,
	parent *Body,
) error {
	if body.Kind == BodyEmpty || body.Temperature == nil {
		return nil
	}

	meanK := body.Temperature.MeanK
	tempRange := MeanKToTempRange(meanK)

	// 1. Atmosphere.ScaleHeight: re-derive from real meanK + gravity.
	if body.Atmosphere != nil && body.Physical != nil {
		body.Atmosphere.ScaleHeight = DeriveScaleHeight(meanK, body.Physical.Gravity)
	}

	// 1.5 (HZ-only): CheckRunawayGreenhouse — may mutate atm.Code to A/B/C
	// for atm 2-9/D/E paths, or fire boiling-only for atm A/B/C/F+ (no
	// mutation). We compare pre/post atm.Code to know which path fired
	// and only re-roll subtype/pressure when the code was mutated. Per
	// WBH p.79, the boiling-only path's "only effect" is the hydro DM-6
	// applied below.
	runawayFired := false
	if body.HZ {
		var preCode int
		if body.Atmosphere != nil {
			preCode = body.Atmosphere.Code
		}
		runawayFired = CheckRunawayGreenhouse(r, body, sys)
		if runawayFired && body.Atmosphere != nil && body.Atmosphere.Code != preCode {
			// Code was mutated (atm 2-9/D/E → A/B/C path). Re-roll subtype
			// + pressure with runawayResult=true (DM+4 to subtype).
			if err := rerollAtmSubtypeAndPressure(r, body, sys, true); err != nil {
				return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: post-runaway: %w", err)
			}
		}
	}

	// 2. Hydrographics.Code: re-roll with current TempRange's Hot/Boiling DMs.
	// If runaway fired, force TempBoiling so the DM-6 (boiling) applies instead of DM-2 (hot).
	hydroTempRange := tempRange
	if runawayFired {
		hydroTempRange = TempBoiling
	}
	if body.Atmosphere != nil && body.Hydrographics != nil {
		newHydro, err := RollHydroDigit(r, body.Atmosphere.Code, body.Atmosphere.Subtype, body.SizeCode, hydroTempRange)
		if err != nil {
			return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: hydro re-roll: %w", err)
		}
		body.Hydrographics.Code = newHydro
		// Refresh PercentRange + Percent so they stay consistent with the new Code.
		body.Hydrographics.PercentRange = HydroRange(newHydro)
		newPct, err := RefineHydroPercent(r, newHydro, body.Hydrographics.PercentRange)
		if err != nil {
			return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: hydro percent: %w", err)
		}
		body.Hydrographics.Percent = newPct
	}

	// 3. Hydrographics.Profile: derive from current code + atm + meanK.
	if body.Hydrographics != nil {
		atmCode := 0
		if body.Atmosphere != nil {
			atmCode = body.Atmosphere.Code
		}
		body.Hydrographics.Profile = DeriveHydrographicsProfile(meanK, atmCode, body.Hydrographics.Code)
	}

	// 4. Atm.Profile: re-derive gas mix for exotic atm (A/B/C) with hydro > 0.
	// RollGasMix's first parameter is the gas-mix table column letter
	// ("A"/"B"/"C"), derived from the atm CODE — not from the atm Subtype
	// (which is a digit/letter from the subtype roll). Atm code 15 (F /
	// Unusual) has no defined column; Profile stays empty for F per MVP.
	if body.Atmosphere != nil && body.Hydrographics != nil && body.Hydrographics.Code > 0 {
		columnLetter := gasMixColumnForAtmCode(body.Atmosphere.Code)
		if columnLetter != "" {
			newProfile, err := RollGasMix(r, columnLetter, "", tempRange, body.SizeCode)
			if err != nil {
				return fmt.Errorf("worlds: RederiveAtmosphereHydrographics: gas mix: %w", err)
			}
			body.Atmosphere.Profile = newProfile
		}
	}

	_ = parent // parent reserved for future moon-specific logic (e.g., parent radiance from atm composition)
	return nil
}

// gasMixColumnForAtmCode maps an atmosphere code to the column letter used
// by RollGasMix / GasMixTableLookup (WBH pp.96-98 Gas Mix tables):
//
//	10 (A) → "A"
//	11 (B) → "B"
//	12 (C) → "C"
//
// Atm code 15 (F / Unusual) has no defined column; returns "" (caller skips
// gas-mix derivation and leaves Atmosphere.Profile empty per MVP).
func gasMixColumnForAtmCode(code int) string {
	switch code {
	case 10:
		return "A"
	case 11:
		return "B"
	case 12:
		return "C"
	default:
		return ""
	}
}

// rerollAtmSubtypeAndPressure re-rolls Atmosphere.Subtype and Atmosphere.Pressure
// for codes A (10), B (11), and C (12) after CheckRunawayGreenhouse fires.
//
//   - A (10): pressure reset to 0 (WBH "Varies" → no fixed range); subtype unchanged.
//   - B (11), C (12): subtype re-rolled with runawayResult DM; pressure set to 0
//     (span = 0 in AtmospherePressureRange for both codes).
//
// Resetting the pressure for code A is critical: when runaway mutates a body
// from a high-pressure code (e.g., D/13 at 10 bar) to A (10), the original
// pressure must be cleared so that pass-2 GenerateTemperature does not inherit
// an implausibly high value and produce GreenhouseFactor ≥ 5.
//
// For atm codes other than A/B/C, no-op.
//
// Mutates body.Atmosphere.Subtype and body.Atmosphere.Pressure on success.
func rerollAtmSubtypeAndPressure(
	r roller.Roller,
	body *Body,
	sys stars.System,
	runawayResult bool,
) error {
	if body.Atmosphere == nil {
		return nil
	}
	code := body.Atmosphere.Code

	// Code A (10): "Varies" pressure — clear it; no subtype to re-roll.
	if code == 10 {
		body.Atmosphere.Pressure = 0
		return nil
	}

	if code != 11 && code != 12 { // only B and C have variable subtypes
		return nil
	}

	hzco := 0.0
	if len(body.Group.Members) > 0 {
		hzco = body.Group.HZCO()
	} else {
		hzco = sys.Primary.HZCO()
	}

	isInsidious := code == 12 // C
	newSubtype, err := RollCorrosiveInsidiousSubtype(r, body.SizeCode, body.Orbit, hzco, isInsidious, runawayResult)
	if err != nil {
		return fmt.Errorf("worlds: rerollAtmSubtypeAndPressure: subtype: %w", err)
	}
	body.Atmosphere.Subtype = newSubtype

	newPressure, err := RollTotalPressure(r, code, newSubtype)
	if err != nil {
		return fmt.Errorf("worlds: rerollAtmSubtypeAndPressure: pressure: %w", err)
	}
	body.Atmosphere.Pressure = newPressure
	return nil
}
