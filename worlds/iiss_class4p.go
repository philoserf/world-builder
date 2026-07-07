package worlds

import (
	"github.com/philoserf/world-builder/iiss"
	"github.com/philoserf/world-builder/stars"
)

// buildClass4PPlanet builds the planet/moon mainworld view (WBH p.138
// Form 0407F-IV PART P) from the generated body. The struct and its
// Markdown renderer live in iiss/; this is the worlds→iiss boundary that
// reads the universe and fills the form.
func buildClass4PPlanet(u *Universe, body *Body, isMainworld bool) *iiss.Class4PPartP {
	p := &iiss.Class4PPartP{
		Designation:  body.Designation,
		SystemAgeGyr: u.System.Primary.AgeGyr,
		OrbitNumber:  body.Orbit,
		AU:           stars.OrbitToAU(body.Orbit),
		Eccentricity: body.Eccentricity,
		PeriodHours:  body.Period.Hours,
		DiameterKm:   body.DiameterKm,
		MassEarth:    body.MassEarth,
		Ring:         body.Ring,
		RingCentrePD: body.RingCentrePD,
		RingSpanPD:   body.RingSpanPD,
		IsMainworld:  isMainworld,
	}
	if body.Kind == BodyMoon {
		// Body.Orbit and Body.Period are star-relative and zero for
		// moons — use the parent's stellar orbit for the AU context and
		// the moon's own parent-relative orbit and period.
		p.OrbitNumber = body.StellarOrbit()
		p.AU = stars.OrbitToAU(body.StellarOrbit())
		p.PeriodHours = body.PeriodHours
		p.MoonOrbitKm = body.OrbitKm
		if body.Parent != nil {
			p.ParentDesignation = body.Parent.Designation
		}
	}
	if body.HasPhysical() {
		p.Composition = body.Physical.Composition
		p.Density = body.Physical.Density
		p.Gravity = body.Physical.Gravity
		p.EscapeVelocity = body.Physical.EscapeVelocity
		p.SizeProfile = body.Physical.SizeProfile
	}
	if body.Kind == BodyGasGiant {
		// Gas giants have no BodyPhysical/atmosphere/hydro/life; their SIZE
		// is class + Earth-diameters, and their only temperature is the
		// WBH p.125 residual heat (stored in Geology.InherentTemperatureK).
		p.IsGasGiant = true
		p.GasGiantClass = ggClassLabel(body.GGClass)
		p.DiameterEarth = body.DiameterEarth
		p.DiameterKm = body.DiameterEarth * DiameterTerra
		if body.HasGeology() {
			p.ResidualTempK = body.Geology.InherentTemperatureK
		}
	}
	if body.HasAtmosphere() {
		atm := body.Atmosphere
		a := &iiss.Class4PAtmosphere{
			Code:                  atm.Code,
			Subtype:               atm.Subtype,
			Pressure:              atm.Pressure,
			OxygenPartialPressure: atm.OxygenPartialPressure,
			ScaleHeight:           atm.ScaleHeight,
			ProfileShorthand:      FormatAtmoProfileShorthand(*atm, atm.Profile),
		}
		for _, t := range atm.Taints {
			a.Taints = append(a.Taints, iiss.Class4PTaint{
				Code: t.Code, Severity: t.Severity, Persistence: t.Persistence,
			})
		}
		for _, h := range atm.InsidiousHazards {
			a.Hazards = append(a.Hazards, h.Code)
		}
		p.Atmosphere = a
	}
	if body.HasHydrographics() {
		hydro := body.Hydrographics
		h := &iiss.Class4PHydrographics{
			Code:    hydro.Code,
			Percent: hydro.Percent,
			Profile: hydro.Profile,
		}
		if body.HasSurfaceDistribution() {
			h.Distribution = body.SurfaceDistribution.Description
			h.Geography = geographyLabel(body.SurfaceDistribution.Geography)
		}
		p.Hydrographics = h
	}
	if body.HasDayLength() {
		p.SiderealHours = body.DayLength.SiderealHours
		p.SolarHours = body.DayLength.SolarHours
		p.SolarDaysPerYear = body.DayLength.YearDays
	}
	if body.HasAxialTilt() {
		p.AxialTiltDeg = body.AxialTilt.Degrees
	}
	p.TidalLockRatio = "no"
	if body.HasTidalLock() && body.TidalLock.LockRatio != "" {
		p.TidalLockRatio = body.TidalLock.LockRatio
	}
	if body.HasTidalEffects() {
		p.TidesMeters = body.TidalEffects.Total
	}
	if body.HasTemperature() {
		t := body.Temperature
		lowK := t.LowK
		if lowK <= 0 && t.MeanK > 0 {
			lowK = -1 // sentinel: render as "—"
		}
		p.Temperature = &iiss.Class4PTemperature{
			HighK:            t.HighK,
			MeanK:            t.MeanK,
			LowK:             lowK,
			Luminosity:       t.Luminosity,
			Albedo:           t.Albedo,
			GreenhouseFactor: t.GreenhouseFactor,
		}
	}
	if body.HasGeology() {
		g := body.Geology
		p.Seismic = &iiss.Class4PSeismic{
			TotalSeismicStress:    g.TotalSeismicStress,
			ResidualSeismicStress: g.ResidualSeismicStress,
			TidalStressFactor:     g.TidalStressFactor,
			TidalHeatingFactor:    g.TidalHeatingFactor,
			TectonicPlates:        g.TectonicPlates,
		}
	}
	if body.HasBiology() {
		bio := body.Biology
		p.Life = &iiss.Class4PLife{
			Biomass:        bio.Biomass,
			Biocomplexity:  bio.Biocomplexity,
			HasSophont:     bio.HasNativeSophont,
			HadExtinct:     bio.HadExtinctSophont,
			Biodiversity:   bio.Biodiversity,
			Compatibility:  bio.Compatibility,
			ResourceRating: bio.ResourceRating,
		}
	}
	if body.HasHabitability() {
		p.HabitabilityRating = body.Habitability.Rating
		p.HabitabilityNotes = body.Habitability.Notes
	}
	p.Subordinates = make([]iiss.Class4PSubordinate, 0, len(body.Children))
	for _, child := range body.Children {
		p.Subordinates = append(p.Subordinates, iiss.Class4PSubordinate{
			Designation:  child.Designation,
			SizeCode:     string(child.SizeCode),
			DiameterKm:   child.DiameterKm,
			OrbitKm:      int(child.OrbitKm),
			Eccentricity: child.Eccentricity,
			PeriodHours:  child.PeriodHours,
		})
	}
	return p
}

// buildClass4PBelt builds the belt-mainworld view (WBH p.139 FORM
// 0407K-IV PART P.B). The struct and its Markdown renderer live in iiss/.
func buildClass4PBelt(u *Universe, body *Body, isMainworld bool) *iiss.Class4PPartPB {
	pb := &iiss.Class4PPartPB{
		Designation:  body.Designation,
		PrimaryGroup: body.Group.Designation,
		SystemAgeGyr: u.System.Primary.AgeGyr,
		OrbitNumber:  body.Orbit,
		AU:           stars.OrbitToAU(body.Orbit),
		PeriodHours:  body.Period.Hours,
		IsMainworld:  isMainworld,
	}
	if body.HasBelt() {
		pb.SpanOrbits = body.Belt.Span
		pb.MTypePct = body.Belt.Composition.MTypePct
		pb.STypePct = body.Belt.Composition.STypePct
		pb.CTypePct = body.Belt.Composition.CTypePct
		pb.OtherPct = body.Belt.Composition.OtherPct
		pb.Bulk = body.Belt.Bulk
		pb.SigSize1Bodies = body.Belt.SigSize1Bodies
		pb.SigSizeSBodies = body.Belt.SigSizeSBodies
		pb.ResourceRating = body.Belt.ResourceRating
	}
	return pb
}

// geographyLabel renders the fundamental-geography enum for the Class IV-P
// hydrographics block.
func geographyLabel(g FundamentalGeography) string {
	if g == GeographyOcean {
		return "Ocean"
	}
	return "Land"
}

// ggClassLabel renders the gas-giant size class (WBH p.55) for the
// Class IV-P size block.
func ggClassLabel(c GasGiantClass) string {
	switch c {
	case GasGiantSmall:
		return "Small"
	case GasGiantMedium:
		return "Medium"
	case GasGiantLarge:
		return "Large"
	}
	return ""
}
