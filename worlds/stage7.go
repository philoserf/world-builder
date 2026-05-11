package worlds

import (
	"wbh/roller"
	"wbh/stars"
)

// ApplyGeology applies the WBH pp.125-127 geology pass to every
// non-empty non-belt body in the universe. Stage-7 orchestrator.
//
// Per dependency-graph.md § Stage 7, partial geology (Residual + TSF
// + THF) is computed inside ConvergeClimate so the post-TSS Temperature
// converges with rederived atm/hydro. Stage 7's remaining work is:
//
//   - HZ bodies (body.Geology already set by climate): roll
//     TectonicPlates and append.
//   - Non-HZ terrestrials and moons (body.Geology nil because climate
//     skipped them): compute the full Geology including TectonicPlates
//     (which will be 0 since they have no Hydrographics — the prereq).
//   - Gas giants: compute GGResidualHeat (no plates).
//   - Belts: skipped (Size 0).
//
// Per anti-pattern A.1, every moon is walked alongside its parent.
func ApplyGeology(r roller.Roller, u *Universe) error {
	sys := u.System
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty || body.SizeCode == "0" {
			continue
		}
		applyBodyGeology(r, body, sys, false)
		for _, child := range body.Children {
			applyBodyGeology(r, child, sys, true)
		}
	}
	return nil
}

// applyBodyGeology dispatches Stage-7 work for one body. If
// ConvergeClimate already populated body.Geology, this just adds
// TectonicPlates; otherwise it computes the full Geology.
func applyBodyGeology(r roller.Roller, body *Body, sys stars.System, isMoon bool) {
	if body.Kind == BodyGasGiant {
		body.Geology = &Geology{
			InherentTemperatureK: ComputeGGResidualHeat(body.MassEarth, sys.Primary.AgeGyr),
		}
		return
	}
	if body.Geology != nil {
		// Climate solver already populated TSS factors — only plates remain.
		body.Geology.TectonicPlates = RollTectonicPlates(r, body, body.Geology.TotalSeismicStress)
		return
	}
	// Non-HZ body — climate skipped it; compute the full geology now.
	g := computePartialGeology(body, sys, isMoon)
	g.TectonicPlates = RollTectonicPlates(r, body, g.TotalSeismicStress)
	body.Geology = g
}

// computePartialGeology populates a Geology with Residual + TSF + THF +
// TotalSeismicStress + InherentTemperatureK for the given body, but
// leaves TectonicPlates at 0. Called by ConvergeClimate inside the
// climate fixed-point loop (atm/hydro-independent) and by ApplyGeology
// for non-HZ bodies that didn't go through climate.
func computePartialGeology(body *Body, sys stars.System, isMoon bool) *Geology {
	g := &Geology{}
	g.ResidualSeismicStress = ComputeResidualSeismicStress(body, sys.Primary.AgeGyr, isMoon)
	g.TidalStressFactor = ComputeTidalStressFactor(body)
	if isMoon && body.Parent != nil {
		g.TidalHeatingFactor = ComputeTidalHeatingFactor(moonTidalHeatingInputs(body, body.Parent))
	} else {
		g.TidalHeatingFactor = ComputeTidalHeatingFactor(planetTidalHeatingInputs(body, sys))
	}
	g.TotalSeismicStress = g.ResidualSeismicStress + g.TidalStressFactor + g.TidalHeatingFactor
	g.InherentTemperatureK = float64(g.TotalSeismicStress)
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
