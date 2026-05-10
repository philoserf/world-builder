package worlds

import (
	"fmt"
	"strconv"
	"strings"

	"wbh/iiss"
	"wbh/stars"
)

// BuildIISSForms populates u.Detail.SystemForms (Class0I, Class23,
// Class4P) from the converged universe state. Pure function — no
// rolls. Called by GenerateWithRoller after AggregateSystem.
//
// Cycle-11 MVP: produces structurally-correct forms with the
// header / counts / star rows / object rows populated; cell-by-cell
// fidelity to the WBH p.35 / p.63 / pp.141-142 layouts lands in a
// post-parity sub-project.
func BuildIISSForms(u *Universe) {
	u.Detail.Class0I = buildClass0I(u)
	u.Detail.Class23 = buildClass23(u)
	u.Detail.Class4P = buildClass4P(u)
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

func buildClass23(u *Universe) iiss.Class23Form {
	c0 := buildClass0I(u)
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

func renderObjectSAH(body *Body) string {
	switch body.Kind {
	case BodyTerrestrial:
		return body.RenderSAH()
	case BodyGasGiant:
		var prefix string
		switch body.GGClass {
		case GasGiantSmall:
			prefix = "GS"
		case GasGiantMedium:
			prefix = "GM"
		case GasGiantLarge:
			prefix = "GL"
		default:
			prefix = "G"
		}
		return prefix + body.GGDiameterCode
	case BodyPlanetoidBelt:
		return "000"
	}
	return ""
}

func renderMoonSAH(m *Body, parentInHZ bool) string {
	if m.GGClass != NotGasGiant {
		var prefix string
		switch m.GGClass {
		case GasGiantSmall:
			prefix = "GS"
		case GasGiantMedium:
			prefix = "GM"
		case GasGiantLarge:
			prefix = "GL"
		}
		return prefix + m.GGDiameterCode
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
	if body.Belt != nil && body.Belt.Profile != "" {
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

func buildClass4P(u *Universe) iiss.Class4PForm {
	form := iiss.Class4PForm{
		FormHeader: buildClass0I(u).FormHeader,
	}
	mainworld := findMainworld(u)
	if mainworld == nil {
		return form
	}
	switch mainworld.Kind {
	case BodyPlanetoidBelt:
		form.Variant = iiss.Class4PBelt
		form.PartPB = &iiss.Class4PPartPB{}
	case BodyMoon:
		form.Variant = iiss.Class4PMoon
		form.PartP = &iiss.Class4PPartP{}
	default:
		form.Variant = iiss.Class4PPlanet
		form.PartP = &iiss.Class4PPartP{}
	}
	return form
}

func findMainworld(u *Universe) *Body {
	if u.Detail.MainworldDesignation == "" {
		return nil
	}
	target := u.Detail.MainworldDesignation
	for i := range u.Detail.Bodies {
		body := &u.Detail.Bodies[i]
		if body.Designation == target {
			return body
		}
		for _, child := range body.Children {
			if child.Designation == target {
				return child
			}
		}
	}
	return nil
}

// buildBodyNotes (cycle-11 simple Notes string) was replaced by
// renderObjectNotes / renderObjectSAH / renderSub for full pass-1
// fidelity in cycle 15.
