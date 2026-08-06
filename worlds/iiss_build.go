package worlds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/world-builder/iiss"
	"github.com/philoserf/world-builder/stars"
)

// BuildIISSForms populates u.Detail.SystemForms (Class0I, Class23,
// Class4P) from the fully-populated universe state per WBH p.35 / p.63 /
// pp.141-142. Pure function — no rolls. Called by GenerateWithRoller
// after AggregateSystem.
func BuildIISSForms(u *Universe) {
	c0 := buildClass0I(u)
	u.Detail.Class0I = c0
	u.Detail.Class23 = buildClass23(u, c0)
	u.Detail.Class4PForms = buildClass4PForms(u, c0.FormHeader)
	u.Detail.Census = iiss.SystemCensus{
		BaselineNumber: u.Placement.BaselineN,
		BaselineOrbit:  u.Placement.BaselineOrbit,
		Spread:         u.Placement.SystemSpread,
		EmptyOrbits:    u.Placement.EmptyOrbits,
	}
	u.Detail.NotableFeatures = NotableFeatures(u)
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

	// Post-fill MAO on Stars rows from the Stage-1 star groups —
	// stars.BuildSurveyForm leaves MAO=0 for the Class 0/I header.
	// The groups come from u.Placement (already computed by
	// GenerateSystemPlacement via AvailableOrbits, which fails the
	// whole generation on error), so no recompute and no swallowed
	// error path here.
	groups := make([]Group, 0, len(u.Placement.Allocations))
	for _, alloc := range u.Placement.Allocations {
		groups = append(groups, alloc.Group)
	}

	fillStarsMAO(form.Stars, groups, u.System)

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
func fillStarsMAO(rows []iiss.Class0IStarRow, groups []Group, sys stars.System) {
	maoByGroup := map[string]float64{}
	for _, g := range groups {
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

// ggPrefix returns the SAH prefix for a gas giant:
// "GS" (Small), "GM" (Medium), "GL" (Large). NotGasGiant returns the
// 1-letter "G" — defensive fallback only; callers should gate by
// GGClass and this path should not fire in normal pipeline flow.
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
		parts = append(parts, formatMass(body.MassEarth)+"⊕")
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

	if body.Ring {
		parts = append(parts, "ring")
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

// buildClass4PForms builds one Class IV-P part per non-empty body, in
// AllBodies() order (ascending orbit, moons after their parent) — the
// per-world detail for the whole system. Belts get PART P.B; terrestrials,
// moons, and gas giants get PART P. The auto-picked mainworld is marked
// via IsMainworld on its own part.
func buildClass4PForms(u *Universe, header iiss.FormHeader) []iiss.Class4PForm {
	var forms []iiss.Class4PForm

	for body := range u.AllBodies() {
		if body.Kind == BodyEmpty {
			continue
		}

		isMW := body == u.Detail.Mainworld

		form := iiss.Class4PForm{FormHeader: header, Designation: body.Designation}
		switch body.Kind {
		case BodyPlanetoidBelt:
			form.Variant = iiss.Class4PBelt
			form.PartPB = buildClass4PBelt(u, body, isMW)
		case BodyMoon:
			form.Variant = iiss.Class4PMoon
			form.PartP = buildClass4PPlanet(u, body, isMW)
		default: // BodyTerrestrial, BodyGasGiant
			form.Variant = iiss.Class4PPlanet
			form.PartP = buildClass4PPlanet(u, body, isMW)
		}

		forms = append(forms, form)
	}

	return forms
}
