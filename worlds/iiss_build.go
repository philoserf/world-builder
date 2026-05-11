package worlds

import (
	"fmt"
	"strconv"
	"strings"

	"wbh/iiss"
	"wbh/stars"
)

// BuildIISSForms populates u.Detail.SystemForms (Class0I, Class23,
// Class4P) from the converged universe state per WBH p.35 / p.63 /
// pp.141-142. Pure function — no rolls. Called by GenerateWithRoller
// after AggregateSystem.
func BuildIISSForms(u *Universe) {
	c0 := buildClass0I(u)
	u.Detail.Class0I = c0
	u.Detail.Class23 = buildClass23(u, c0)
	u.Detail.Class4P = buildClass4P(u, c0.FormHeader)
}

func buildClass0I(u *Universe) iiss.Class0IForm {
	// Delegate to pass-1's stars.BuildSurveyForm for full companion +
	// composite-barycentre fidelity, then translate to the iiss/
	// boundary type.
	meta := stars.SurveyMetadata{
		Sector:      "—",
		Location:    "—",
		Designation: u.Detail.MainworldDesignation,
	}
	sf := stars.BuildSurveyForm(u.System, meta)
	form := iiss.Class0IForm{
		FormHeader: iiss.FormHeader{
			SystemName:    u.System.PrimaryDesignation,
			Sector:        sf.Sector,
			Location:      sf.Location,
			IISSDesig:     sf.IISSDesig,
			InitialSurvey: sf.InitialSurvey,
			LastUpdated:   sf.LastUpdated,
		},
		SystemAgeGyr: sf.SystemAgeGyr,
		StellarCount: sf.StellarCount,
	}
	for _, c := range sf.Stars {
		form.Stars = append(form.Stars, iiss.Class0IStarRow{
			Component:    c.Component,
			Class:        c.Class,
			Mass:         c.Mass,
			Diameter:     c.Diameter,
			Temperature:  c.Temperature,
			Luminosity:   c.Luminosity,
			Orbit:        c.Orbit,
			AU:           c.AU,
			Eccentricity: c.Eccentricity,
			PeriodYears:  c.PeriodYears,
			HZCO:         c.HZCO,
			MAO:          c.MAO,
		})
	}
	return form
}

func buildClass23(u *Universe, c0 iiss.Class0IForm) iiss.Class23Form {
	form := iiss.Class23Form{
		FormHeader:   c0.FormHeader,
		SystemAgeGyr: c0.SystemAgeGyr,
		StellarCount: c0.StellarCount,
		Stars:        c0.Stars,
		Counts: iiss.Class23Counts{
			GasGiants:      u.Placement.Counts.GasGiants,
			PlanetoidBelts: u.Placement.Counts.PlanetoidBelts,
			Terrestrials:   u.Placement.Counts.Terrestrials,
			Total:          u.Placement.Counts.Total,
		},
	}

	// Post-fill MAO on Stars rows from the AvailableOrbits result —
	// stars.BuildSurveyForm leaves MAO=0 for the Class 0/I header.
	if avail, err := AvailableOrbits(u.System); err == nil {
		fillStarsMAO(form.Stars, avail, u.System)
	}

	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Kind == BodyEmpty {
			continue
		}
		form.Objects = append(form.Objects, iiss.Class23Object{
			Primary:     body.Group.Designation,
			Designation: body.Designation,
			Orbit:       body.Orbit,
			AU:          stars.OrbitToAU(body.Orbit),
			Ecc:         body.Eccentricity,
			PeriodStr:   formatPeriod(body.Period),
			SAH:         renderObjectSAH(body),
			Sub:         renderSub(body),
			Notes:       renderObjectNotes(body),
		})
		for _, child := range body.Children {
			form.Objects = append(form.Objects, iiss.Class23Object{
				Primary:     body.Group.Designation,
				Designation: child.Designation,
				SAH:         renderMoonSAH(child, body.HZ),
				Sub:         "",
			})
		}
	}
	return form
}

// fillStarsMAO walks Stars rows and copies MAO from AvailableOrbits.
// Composite rows ("Aab", "AB", "ABC") get MAO from the corresponding
// outer companion's AU.
func fillStarsMAO(rows []iiss.Class0IStarRow, avail Result, sys stars.System) {
	maoByGroup := map[string]float64{}
	for _, g := range avail.Groups {
		maoByGroup[g.Designation] = g.MAO
	}
	composeMAO := map[string]float64{}
	for _, c := range sys.Companions {
		switch c.OrbitClass {
		case stars.OrbitClose, stars.OrbitNear:
			if c.AU > composeMAO["AB"] {
				composeMAO["AB"] = c.AU
			}
		case stars.OrbitFar:
			composeMAO["ABC"] = c.AU
		}
	}
	for i, row := range rows {
		key := row.Component
		if idx := strings.Index(key, " ("); idx >= 0 {
			key = key[:idx]
		}
		if mao, ok := maoByGroup[key]; ok {
			rows[i].MAO = mao
			continue
		}
		if mao, ok := composeMAO[key]; ok {
			rows[i].MAO = mao
		}
	}
}

// ggPrefix returns the 2-letter SAH prefix for a gas giant.
// NotGasGiant → "G" (defensive — callers should gate by GGClass).
func ggPrefix(c GasGiantClass) string {
	switch c {
	case GasGiantSmall:
		return "GS"
	case GasGiantMedium:
		return "GM"
	case GasGiantLarge:
		return "GL"
	}
	return "G"
}

func renderObjectSAH(body *Body) string {
	switch body.Kind {
	case BodyTerrestrial:
		return body.RenderSAH()
	case BodyGasGiant:
		return ggPrefix(body.GGClass) + body.GGDiameterCode
	case BodyPlanetoidBelt:
		return "000"
	}
	return ""
}

func renderMoonSAH(m *Body, parentInHZ bool) string {
	if m.GGClass != NotGasGiant {
		return ggPrefix(m.GGClass) + m.GGDiameterCode
	}
	if parentInHZ {
		return string(m.SizeCode) + "??"
	}
	return string(m.SizeCode)
}

func renderSub(body *Body) string {
	if body.Kind == BodyPlanetoidBelt {
		return "?"
	}
	if len(body.Children) == 0 {
		return "0"
	}
	return strconv.Itoa(len(body.Children))
}

func renderObjectNotes(body *Body) string {
	parts := []string{}
	if body.Kind == BodyGasGiant {
		parts = append(parts, fmt.Sprintf("%s⊕", formatMass(body.MassEarth)))
	}
	if body.HZ {
		parts = append(parts, "HZ")
	}
	if len(body.Children) > 0 {
		moonSAH := make([]string, 0, len(body.Children))
		for _, child := range body.Children {
			moonSAH = append(moonSAH, renderMoonSAH(child, body.HZ))
		}
		parts = append(parts, strings.Join(moonSAH, ", "))
	}
	if body.HasBelt() && body.Belt.Profile != "" {
		parts = append(parts, body.Belt.Profile)
	}
	return strings.Join(parts, ", ")
}

func formatMass(m float64) string {
	n := int(m + 0.5)
	s := strconv.Itoa(n)
	if n < 1000 {
		return s
	}
	out := []byte(s)
	for i := len(out) - 3; i > 0; i -= 3 {
		out = append(out[:i], append([]byte{','}, out[i:]...)...)
	}
	return string(out)
}

func formatPeriod(p Period) string {
	if p.Years > 0 && p.Years < 0.05 {
		return fmt.Sprintf("%.3fd", p.Days)
	}
	if p.Years >= 1000 {
		return formatMass(p.Years) + "y"
	}
	return fmt.Sprintf("%.3fy", p.Years)
}

func buildClass4P(u *Universe, header iiss.FormHeader) iiss.Class4PForm {
	form := iiss.Class4PForm{FormHeader: header}
	mainworld := u.Detail.Mainworld
	if mainworld == nil {
		return form
	}
	switch mainworld.Kind {
	case BodyPlanetoidBelt:
		form.Variant = iiss.Class4PBelt
		form.PartPB = buildClass4PPartPB(u, mainworld)
	case BodyMoon:
		form.Variant = iiss.Class4PMoon
		form.PartP = buildClass4PPartP(u, mainworld)
	default:
		form.Variant = iiss.Class4PPlanet
		form.PartP = buildClass4PPartP(u, mainworld)
	}
	return form
}

func buildClass4PPartP(u *Universe, body *Body) *iiss.Class4PPartP {
	p := &iiss.Class4PPartP{
		Designation:  body.Designation,
		SystemAgeGyr: u.System.Primary.AgeGyr,
		OrbitNumber:  body.Orbit,
		AU:           stars.OrbitToAU(body.Orbit),
		Eccentricity: body.Eccentricity,
		PeriodHours:  body.Period.Hours,
		DiameterKm:   body.DiameterKm,
		MassEarth:    body.MassEarth,
		IsMainworld:  true,
	}
	if body.HasPhysical() {
		p.Density = body.Physical.Density
		p.Gravity = body.Physical.Gravity
	}
	if body.HasAtmosphere() {
		atm := body.Atmosphere
		p.Atmosphere = &iiss.Class4PAtmosphere{
			Code:                  atm.Code,
			Pressure:              atm.Pressure,
			OxygenPartialPressure: atm.OxygenPartialPressure,
			ScaleHeight:           atm.ScaleHeight,
			ProfileShorthand:      FormatAtmoProfileShorthand(*atm, atm.Profile),
		}
	}
	if body.HasHydrographics() {
		hydro := body.Hydrographics
		p.Hydrographics = &iiss.Class4PHydrographics{
			Code:    hydro.Code,
			Percent: hydro.Percent,
			Profile: hydro.Profile,
		}
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

func buildClass4PPartPB(u *Universe, body *Body) *iiss.Class4PPartPB {
	pb := &iiss.Class4PPartPB{
		Designation:  body.Designation,
		PrimaryGroup: body.Group.Designation,
		SystemAgeGyr: u.System.Primary.AgeGyr,
		OrbitNumber:  body.Orbit,
		AU:           stars.OrbitToAU(body.Orbit),
		PeriodHours:  body.Period.Hours,
		IsMainworld:  true,
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
