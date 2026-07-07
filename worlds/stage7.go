package worlds

import (
	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// ApplyGeology applies the WBH pp.125-127 geology pass to every
// non-empty non-belt body in the universe. Stage-7 orchestrator.
//
// Per dependency-graph.md § Stage 7, partial geology (Residual + TSF
// + THF) is computed inside ApplyClimatePasses so that atmosphere and
// hydrographics are re-derived from the post-TSS Temperature across the
// two climate passes. Stage 7's remaining work is:
//
//   - HZ bodies (body.Geology already set by climate): roll
//     TectonicPlates and append.
//   - Non-HZ terrestrials and moons (body.Geology nil because climate
//     skipped them): compute the full Geology including TectonicPlates
//     (which will be 0 since they have no Hydrographics — the prereq).
//   - Gas giants: compute GGResidualHeat (no plates).
//   - Belts: skipped (Size 0).
//
// Per anti-pattern A.1, every moon is walked alongside its parent via
// AllBodiesWithParent — the parent value also distinguishes moon
// (isMoon=true) from top-level (isMoon=false) for the per-body call.
func ApplyGeology(r roller.Roller, u *Universe) error {
	sys := u.System
	for body, parent := range u.AllBodiesWithParent() {
		if body.Kind == BodyEmpty || body.SizeCode == "0" {
			continue
		}
		applyBodyGeology(bodySub(r, body, parent, "geology"), body, sys, parent != nil)
	}
	return nil
}

// applyBodyGeology dispatches Stage-7 work for one body. If
// ApplyClimatePasses already populated body.Geology, this just adds
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
// leaves TectonicPlates at 0. Called by ApplyClimatePasses inside the
// two climate passes (atm/hydro-independent) and by ApplyGeology
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
		WorldMassEarth:   body.MassOrDerived(),
	}
}

// moonTidalHeatingInputs derives TidalHeatingInputs for a moon around
// its parent planet.
func moonTidalHeatingInputs(moon, parent *Body) TidalHeatingInputs {
	return TidalHeatingInputs{
		PrimaryMassEarth: parent.MassOrDerived(),
		SizeN:            SizeAsInt(moon.SizeCode),
		Eccentricity:     moon.Eccentricity,
		DistanceMkm:      moon.OrbitKm / 1_000_000.0,
		PeriodDays:       moon.PeriodHours / 24.0,
		WorldMassEarth:   moon.MassOrDerived(),
	}
}
