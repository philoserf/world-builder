package worlds

import (
	"fmt"

	"wbh/roller"
	"wbh/stars"
)

// ApplyClimate runs the climate fixed-point cluster (atmosphere ↔
// hydrographics ↔ temperature) for every eligible body in the
// universe. Stage-5 orchestrator per WBH pp.79, 81, 96-99, 102, 108-126.
//
// Per dependency-graph.md § Stage 5, climate is a fixed-point: pass-1
// approximated it with a 2-pass rederive after a temperature pass.
// Cycle-5 MVP replicates that pattern (initial atm/hydro → temperature
// → rederive → temperature → rederive) and treats it as converged.
// Formal N-iteration assertion is a cycle-5 follow-up.
//
// Per docs/pass-2/spike-findings.md § Finding 6a, the panic-vs-error
// stance for Seeded rollers in production is deferred to its own spec
// — current behavior is to error on convergence overflow, but cycle-5
// MVP doesn't trigger overflow because the loop count is fixed.
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
// docs/pass-2/api-surface.md § The Climate solver. Cycle-5 MVP
// implements pass-1's 2-pass rederive flow; a formal N=3 iteration
// loop with assertable convergence lands as cycle-5 follow-up.
//
// Mutates body.Atmosphere, body.Hydrographics, body.Temperature on
// return. No-op for ineligible bodies (non-HZ, atmosphereless,
// gas giants, belts).
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

	// Stage-5C: temperature pass.
	parent := body.Parent
	temp, terr := GenerateTemperature(r, body, sys, parent)
	if terr != nil {
		return fmt.Errorf("temperature: %w", terr)
	}
	body.Temperature = temp

	// Stage-5D: 2-pass rederive (atm/hydro from real temperature, then
	// recompute temperature, rederive again).
	if err := RederiveAtmosphereHydrographics(r, body, sys, parent); err != nil {
		return fmt.Errorf("rederive 1: %w", err)
	}
	temp, terr = GenerateTemperature(r, body, sys, parent)
	if terr != nil {
		return fmt.Errorf("temperature 2: %w", terr)
	}
	body.Temperature = temp
	if err := RederiveAtmosphereHydrographics(r, body, sys, parent); err != nil {
		return fmt.Errorf("rederive 2: %w", err)
	}
	return nil
}
