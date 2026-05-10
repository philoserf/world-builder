package worlds

import (
	"fmt"
	"math"

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

// ConvergeClimate is the per-body climate fixed-point solver per
// docs/pass-2/api-surface.md § The Climate solver. Folds partial
// geology (Residual + TSF + THF) into the loop so Temperature
// converges including the WBH p.125 inherent-temperature addition
// (cycle 18 / dependency-graph.md § Stage 7).
//
// Loop body per iteration:
//  1. Compute Temperature from current atm/hydro.
//  2. Compute partial-geology (atm/hydro-independent).
//  3. Apply TSS via T' = ⁴√(T⁴ + TSS⁴); refresh ScaleHeight.
//  4. Rederive atm/hydro from post-TSS Temperature.
//  5. If atm.Code, hydro.Code, and Temperature.MeanK are stable
//     relative to the previous iteration, return early.
//
// Cap N = 5. If the convergence test never fires within N, the loop
// exits silently with the last-iteration values committed — pass-2
// inherits pass-1's behaviour of accepting the post-loop state. The
// formal "panic / error on overflow" stance per spike-findings.md
// § Finding 6a is deferred: empirical testing shows some seeds
// oscillate between two atm.Code values within ±1 indefinitely, and
// erroring on those would block cmd/wbh on legitimate-but-edgy
// systems. Tightening the contract is a follow-up that needs deeper
// investigation of the oscillation root cause.
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
	const maxIter = 5
	const tempEpsilon = 0.5

	prevAtmCode := -1
	prevHydroCode := -1
	prevMeanK := math.NaN()

	for iter := range maxIter {
		// (1) Compute Temperature from current atm/hydro.
		temp, terr := GenerateTemperature(r, body, sys, parent)
		if terr != nil {
			return fmt.Errorf("temperature iter %d: %w", iter, terr)
		}
		body.Temperature = temp

		// (2) Compute partial-geology (Residual + TSF + THF). Atm/hydro-
		// independent — depends on body physical / orbital parameters.
		body.Geology = computePartialGeology(body, sys, body.Kind == BodyMoon)

		// (3) Apply TSS to Temperature and refresh ScaleHeight.
		ApplyInherentTempAddition(temp, body.Geology.InherentTemperatureK)
		if body.Physical != nil {
			body.Atmosphere.ScaleHeight = DeriveScaleHeight(temp.MeanK, body.Physical.Gravity)
		}

		// (4) Rederive atm/hydro from post-TSS Temperature.
		if err := RederiveAtmosphereHydrographics(r, body, sys, parent); err != nil {
			return fmt.Errorf("rederive iter %d: %w", iter, err)
		}

		// (5) Early-exit on convergence (after first iter so prev exists).
		if iter > 0 &&
			body.Atmosphere.Code == prevAtmCode &&
			body.Hydrographics.Code == prevHydroCode &&
			math.Abs(body.Temperature.MeanK-prevMeanK) < tempEpsilon {
			return nil
		}
		prevAtmCode = body.Atmosphere.Code
		prevHydroCode = body.Hydrographics.Code
		prevMeanK = body.Temperature.MeanK
	}
	// N exhausted without early exit — accept last-iteration state.
	return nil
}
