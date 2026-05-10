package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// ApplyGeology applies the WBH pp.125-127 geology pass to every
// non-empty non-belt body in the universe. Stage-7 orchestrator.
//
// Sub-stages per body:
//
//  1. Geology (residual seismic + tidal stress + tidal heating + GG
//     residual heat for gas giants).
//  2. Apply inherent temperature addition to Temperature in place.
//  3. Recompute Atmosphere.ScaleHeight against the post-TSS mean
//     temperature.
//  4. Roll tectonic plates (terrestrial bodies with hydro ≥ 1).
//
// Per dependency-graph.md § Stage 7, pass-2's design folds the TSS-
// temperature back-edge into the climate solver. Cycle-7 MVP keeps
// pass-1's flow (TSS computed post-climate, applied forward to the
// already-converged temperature) — the formal fold-in is deferred.
//
// Per anti-pattern A.1, every moon is walked alongside its parent.
func ApplyGeology(r roller.Roller, u *Universe) error {
	sys := u.System
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty || body.SizeCode == "0" {
			continue
		}
		body.Geology = computeBodyGeology(r, body, sys, false)
		applyInherentTemp(body)

		for _, child := range body.Children {
			child.Geology = computeBodyGeology(r, child, sys, true)
			applyInherentTemp(child)
		}
	}
	return nil
}

// applyInherentTemp updates Temperature.MeanK in place via the WBH
// p.125 formula T' = ⁴√(T⁴ + TSS⁴), then refreshes Atmosphere.ScaleHeight
// to reflect the new mean temperature.
func applyInherentTemp(body *Body) {
	if body.Temperature == nil || body.Geology == nil {
		return
	}
	ApplyInherentTempAddition(body.Temperature, body.Geology.InherentTemperatureK)
	if body.Atmosphere != nil && body.Physical != nil {
		body.Atmosphere.ScaleHeight = DeriveScaleHeight(body.Temperature.MeanK, body.Physical.Gravity)
	}
}

// computeBodyGeology populates a Geology for the given body. isMoon
// hints residual seismic stress (moon = +1 DM per WBH p.125).
func computeBodyGeology(r roller.Roller, body *Body, sys stars.System, isMoon bool) *Geology {
	g := &Geology{}
	if body.Kind == BodyGasGiant {
		g.InherentTemperatureK = ComputeGGResidualHeat(body.MassEarth, sys.Primary.AgeGyr)
		return g
	}
	g.ResidualSeismicStress = ComputeResidualSeismicStress(body, sys.Primary.AgeGyr, isMoon)
	g.TidalStressFactor = ComputeTidalStressFactor(body)
	if isMoon && body.Parent != nil {
		g.TidalHeatingFactor = ComputeTidalHeatingFactor(moonTidalHeatingInputs(body, body.Parent))
	} else {
		g.TidalHeatingFactor = ComputeTidalHeatingFactor(planetTidalHeatingInputs(body, sys))
	}
	g.TotalSeismicStress = g.ResidualSeismicStress + g.TidalStressFactor + g.TidalHeatingFactor
	g.InherentTemperatureK = float64(g.TotalSeismicStress)
	g.TectonicPlates = RollTectonicPlates(r, body, g.TotalSeismicStress)
	return g
}

// planetTidalHeatingInputs derives TidalHeatingInputs for a planet
// around its primary star.
func planetTidalHeatingInputs(body *Body, sys stars.System) TidalHeatingInputs {
	const auMkm = 149.6
	const solarMassEarth = 332946.0
	return TidalHeatingInputs{
		PrimaryMassEarth: sys.Primary.Mass * solarMassEarth,
		SizeN:            SizeAsInt(body.SizeCode),
		Eccentricity:     body.Eccentricity,
		DistanceMkm:      stars.OrbitToAU(body.Orbit) * auMkm,
		PeriodDays:       body.Period.Hours / 24.0,
		WorldMassEarth:   bodyMassEarth(body),
	}
}

// moonTidalHeatingInputs derives TidalHeatingInputs for a moon around
// its parent planet.
func moonTidalHeatingInputs(moon, parent *Body) TidalHeatingInputs {
	return TidalHeatingInputs{
		PrimaryMassEarth: bodyMassEarth(parent),
		SizeN:            SizeAsInt(moon.SizeCode),
		Eccentricity:     moon.Eccentricity,
		DistanceMkm:      moon.OrbitKm / 1_000_000.0,
		PeriodDays:       moon.PeriodHours / 24.0,
		WorldMassEarth:   bodyMassEarth(moon),
	}
}

// bodyMassEarth returns body.MassEarth if populated, else falls back
// to DeriveMass(density, diameter). Stage 2 sets MassEarth for gas
// giants only — terrestrial bodies derive it from their physical
// composition + diameter via Stage 3.
func bodyMassEarth(body *Body) float64 {
	if body == nil {
		return 0
	}
	if body.MassEarth != 0 {
		return body.MassEarth
	}
	if body.Physical != nil {
		return DeriveMass(body.Physical.Density, body.DiameterKm)
	}
	return 0
}
