package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// ApplyClimate runs the climate fixed-point cluster (atmosphere ↔
// hydrographics ↔ temperature, with partial geology folded in) for
// every eligible body in the universe. Stage-5 orchestrator per WBH
// pp.79, 81, 96-99, 102, 108-126.
//
// Per dependency-graph.md § Stage 7, partial-geology (Residual + TSF +
// THF) is computed inside the climate loop so the post-TSS Temperature
// converges with the rederived atm/hydro. Tectonic plates and GG
// residual heat are forward-only post-climate (Stage 7).
//
// Per anti-pattern A.1, every HZ-planet moon is walked alongside its
// parent.
func ApplyClimate(r roller.Roller, u *Universe) error {
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if err := ConvergeClimate(r, body, u.System); err != nil {
			return fmt.Errorf("worlds: stage5 climate %s: %w", body.Designation, err)
		}
		// Walk moons of HZ planets only (per pass-1 5A).
		if !body.HZ {
			continue
		}
		for _, child := range body.Children {
			if err := ConvergeClimate(r, child, u.System); err != nil {
				return fmt.Errorf("worlds: stage5 moon climate %s: %w", child.Designation, err)
			}
		}
	}
	return nil
}

// initialAtmosphere rolls the atmosphere code, subtype, pressure,
// oxygen partial pressure, and scale-height seed using the HZ-offset
// proxy temperature. Returns (atmo, true) when the body is eligible
// (HZ-orbit terrestrial with size code), (Atmosphere{}, false) otherwise.
func initialAtmosphere(r roller.Roller, body *Body, ageGyr float64) (Atmosphere, bool, error) {
	if body == nil || body.Kind == BodyEmpty {
		return Atmosphere{}, false, nil
	}
	host := body
	if body.Kind == BodyMoon && body.Parent != nil {
		host = body.Parent
	}
	if !host.HZ {
		return Atmosphere{}, false, nil
	}
	if body.GGClass != NotGasGiant {
		return Atmosphere{}, false, nil
	}
	switch body.SizeCode {
	case "", "0", "R":
		return Atmosphere{}, false, nil
	}
	hzco := host.Group.HZCO()
	offset := host.Orbit - hzco
	atmoCode, err := RollAtmoCode(r, body.SizeCode, offset)
	if err != nil {
		return Atmosphere{}, false, fmt.Errorf("atmo code: %w", err)
	}
	atmo := Atmosphere{Code: atmoCode}
	if atmoCode == 11 || atmoCode == 12 {
		st, serr := RollCorrosiveInsidiousSubtype(r, body.SizeCode, host.Orbit, hzco, atmoCode == 12, false)
		if serr != nil {
			return Atmosphere{}, false, fmt.Errorf("atmo subtype: %w", serr)
		}
		atmo.Subtype = st
	}
	press, perr := RollTotalPressure(r, atmoCode, atmo.Subtype)
	if perr != nil {
		return Atmosphere{}, false, fmt.Errorf("pressure: %w", perr)
	}
	atmo.Pressure = press
	if atmoCode >= 2 && atmoCode <= 9 {
		frac, ferr := RollOxygenFraction(r, ageGyr)
		if ferr != nil {
			return Atmosphere{}, false, fmt.Errorf("oxygen: %w", ferr)
		}
		atmo.OxygenPartialPressure = frac * press
	}
	if body.Physical != nil {
		meanT := tempRangeMidpointK(HZCOOffsetToTempRange(host.Orbit, hzco))
		atmo.ScaleHeight = DeriveScaleHeight(meanT, body.Physical.Gravity)
	}
	return atmo, true, nil
}

// tempRangeMidpointK returns a representative mean temperature in
// Kelvin for the given TempRange band. Used for scale-height
// calculation in initial atmosphere derivation, before the full
// temperature roll.
func tempRangeMidpointK(t TempRange) float64 {
	switch t {
	case TempBoiling:
		return 600
	case TempHot:
		return 400
	case TempTemperate:
		return 313
	case TempCold:
		return 200
	case TempFrozen:
		return 100
	}
	return 288
}

// ConvergeClimate is the per-body climate solver. Folds partial
// geology (Residual + TSF + THF) into each pass so Temperature
// includes the WBH p.125 inherent-temperature addition (cycle 18 /
// dependency-graph.md § Stage 7).
//
// Each pass:
//
//  1. Compute Temperature from current atm/hydro.
//  2. Compute partial-geology (atm/hydro-independent).
//  3. Apply TSS via T' = ⁴√(T⁴ + TSS⁴); refresh ScaleHeight.
//  4. Rederive atm/hydro from post-TSS Temperature.
//
// Runs exactly 2 passes (matching pass-1's behaviour). The name
// "ConvergeClimate" is a misnomer inherited from the original
// pass-2 design — empirical investigation (post-cycle-17) showed
// the climate cluster is NOT a fixed point in the strict mathematical
// sense: RederiveAtmosphereHydrographics calls RollHydroDigit, which
// consumes fresh dice from the Roller each call. Each iteration is a
// fresh stochastic sample, not a convergence step. The dice script
// drives the outcome; there is no fixed point to find.
//
// Pass-1 ran exactly 2 rederive passes and trusted the second sample.
// Pass-2 inherits that pragmatic choice. The earlier N=5 iteration
// loop with early-exit (cycle 17) was an over-engineering of a
// stochastic-sampling pattern as if it were deterministic.
//
// No-op for ineligible bodies (non-HZ, atmosphereless, gas giants,
// belts). For HZ bodies, body.Geology is also populated with the
// final TSS factors (without TectonicPlates — that's Stage 7).
func ConvergeClimate(r roller.Roller, body *Body, sys stars.System) error {
	atmo, eligible, err := initialAtmosphere(r, body, sys.Primary.AgeGyr)
	if err != nil {
		return err
	}
	if !eligible {
		return nil
	}
	body.Atmosphere = &atmo

	host := body
	if body.Kind == BodyMoon && body.Parent != nil {
		host = body.Parent
	}
	hzco := host.Group.HZCO()
	tempRange := HZCOOffsetToTempRange(host.Orbit, hzco)

	hydro, herr := GenerateHydrographics(r, atmo, body.SizeCode, tempRange)
	if herr != nil {
		return fmt.Errorf("hydro: %w", herr)
	}
	body.Hydrographics = &hydro

	parent := body.Parent

	// Pass 1.
	if err := climatePass(r, body, sys, parent, 1); err != nil {
		return err
	}
	// Pass 2 (final — matches pass-1's behaviour).
	if err := climatePass(r, body, sys, parent, 2); err != nil {
		return err
	}
	return nil
}

// climatePass runs one temp + partial-geology + TSS-apply + rederive
// pass. pass is the 1-based pass index, used for error context only.
func climatePass(r roller.Roller, body *Body, sys stars.System, parent *Body, pass int) error {
	temp, err := GenerateTemperature(r, body, sys, parent)
	if err != nil {
		return fmt.Errorf("temperature pass %d: %w", pass, err)
	}
	body.Temperature = temp

	body.Geology = computePartialGeology(body, sys, body.Kind == BodyMoon)

	ApplyInherentTempAddition(temp, body.Geology.InherentTemperatureK)
	if body.Physical != nil {
		body.Atmosphere.ScaleHeight = DeriveScaleHeight(temp.MeanK, body.Physical.Gravity)
	}

	if err := RederiveAtmosphereHydrographics(r, body, sys, parent); err != nil {
		return fmt.Errorf("rederive pass %d: %w", pass, err)
	}
	return nil
}
