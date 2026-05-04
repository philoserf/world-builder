// Package worlds — temperature characteristics per WBH pp.108-126.
package worlds

import (
	"math"

	"wbh/roller"
	"wbh/stars"
)

// MeanTemperatureK computes the world's mean temperature in Kelvin per WBH p.111:
//
//	T = 279 × ⁴√(L × (1 - A) × (1 + G) / d²)
//
// Per p.111 thumb-rule-two: "(1+G) should never use a factor of more than
// luminosity × 1.999 and a low no less than luminosity × 0.001". Clamp (1+G)
// to [0.001, 1.999].
func MeanTemperatureK(luminosity, albedo, greenhouse, au float64) float64 {
	if au <= 0 {
		return 0
	}
	gFactor := 1 + greenhouse
	if gFactor < 0.001 {
		gFactor = 0.001
	}
	if gFactor > 1.999 {
		gFactor = 1.999
	}
	core := luminosity * (1 - albedo) * gFactor / (au * au)
	if core <= 0 {
		return 0
	}
	return 279.0 * math.Pow(core, 0.25)
}

// CombineTemperatures combines independent temperature sources per WBH p.109:
//
//	T_total = ⁴√(T₁⁴ + T₂⁴ + …)
//
// Used to add a moon's parent-body IR contribution to its stellar temperature
// (p.125-126), and to combine separate stellar groups for a body orbiting a
// barycenter with multiple non-close-binary stars.
func CombineTemperatures(temps ...float64) float64 {
	if len(temps) == 0 {
		return 0
	}
	if len(temps) == 1 {
		return temps[0]
	}
	sumOf4ths := 0.0
	for _, t := range temps {
		sumOf4ths += math.Pow(t, 4)
	}
	return math.Pow(sumOf4ths, 0.25)
}

// basicMeanTemperatureK maps modified roll → Kelvin per WBH p.109 table.
var basicMeanTemperatureK = map[int]float64{
	0: 178, 1: 198, 2: 218, 3: 238, 4: 263, 5: 278, 6: 283, 7: 288,
	8: 293, 9: 298, 10: 313, 11: 338, 12: 388,
}

// Temperature — per-body temperature characteristics per WBH pp.108-126.
//
// Computed from currently-stored Atmosphere.Pressure, Atmosphere.ScaleHeight,
// and Hydrographics.Code. These 3A1 fields are provisional under HZCO
// temperature until 3A2b-rederive runs an iteration loop that re-derives
// them under real temperature.
type Temperature struct {
	// Headline values (all Kelvin)
	MeanK      float64 // canonical mean per equation (p.111)
	HighK      float64 // p.114 step 9 (populated by Task 7)
	LowK       float64 // p.114 step 9 (populated by Task 7)
	BasicK     float64 // basic table value (p.109) — sanity-check companion
	WorstHighK float64 // p.115 sidebar (populated by Task 7)
	WorstLowK  float64 // p.115 sidebar (populated by Task 7)

	// Equation inputs
	Luminosity       float64 // total in-group stellar luminosity (solar units)
	Albedo           float64 // 0.02..0.98 per p.110
	GreenhouseFactor float64 // ≥ 0; (1+G) clamped to [0.001, 1.999] inside MeanTemperatureK
	AU               float64 // distance from primary stellar source
	ScaleHeight      float64 // km; cached for AdjustedForAltitude (p.123-124)

	// Variance components (cached so scenario methods don't recompute; populated by Task 7)
	AxialTiltFactor    float64
	RotationFactor     float64
	GeographicFactor   float64
	AtmosphericFactor  float64
	LuminosityModifier float64
	NearAU             float64
	FarAU              float64

	// Twilight zone (only populated when body is 1:1 star-locked; Task 8)
	IsTwilight  bool
	TwilightK   float64 // band centerline = MeanK
	BrightSideK float64 // perpetual day
	DarkSideK   float64 // perpetual night

	// Multi-source addition (Task 9)
	ParentRadianceK float64 // contribution from parent body's thermal IR (0 for planets)
}

// GenerateTemperature is the per-body 3A2b-temp orchestrator. Returns nil
// (no error) for empty bodies. For a moon, parent is the parent planet's
// DetailedPlacement (its Temperature field, if populated, is read for
// multi-source IR addition once Task 9 lands).
//
// Currently populates: Luminosity, AU, ScaleHeight, Albedo, GreenhouseFactor,
// MeanK, BasicK (Task 6); AxialTiltFactor, RotationFactor, GeographicFactor,
// AtmosphericFactor, LuminosityModifier, NearAU, FarAU, HighK, LowK,
// WorstHighK, WorstLowK (Task 7); IsTwilight, TwilightK, BrightSideK,
// DarkSideK for 1:1 star-locked bodies (Task 8).
//
// Pending: parent-IR multi-source addition for moons of warm gas giants
// (Task 9 → ParentRadianceK).
func GenerateTemperature(
	r roller.Roller,
	body *DetailedPlacement,
	sys stars.System,
	parent *DetailedPlacement,
) (*Temperature, error) {
	if body.Body == BodyEmpty {
		return nil, nil
	}

	t := &Temperature{}

	// Equation inputs: stellar luminosity (sum within close-binary group),
	// AU (parent's AU for moons; otherwise own orbit converted).
	t.Luminosity = totalStellarLuminosity(sys)
	if parent != nil {
		t.AU = stars.OrbitToAU(parent.Orbit)
	} else {
		t.AU = stars.OrbitToAU(body.Orbit)
	}
	if body.Atmosphere != nil {
		t.ScaleHeight = body.Atmosphere.ScaleHeight
	}

	// Albedo + greenhouse → mean.
	t.Albedo = ComputeAlbedo(r, body, sys)
	t.GreenhouseFactor = ComputeGreenhouseFactor(r, body.Atmosphere)
	t.MeanK = MeanTemperatureK(t.Luminosity, t.Albedo, t.GreenhouseFactor, t.AU)

	// Basic table roll (sanity-check companion).
	_, t.BasicK = BasicTemperatureRoll(r, body, sys)

	// Variance components per WBH p.112-114.
	t.AxialTiltFactor = computeAxialTiltFactor(body)
	t.RotationFactor = computeRotationFactor(body)
	t.GeographicFactor = computeGeographicFactor(body)
	t.AtmosphericFactor = 1.0
	if body.Atmosphere != nil {
		t.AtmosphericFactor = 1 + body.Atmosphere.Pressure
	}

	variance := t.AxialTiltFactor + t.RotationFactor + t.GeographicFactor
	if variance < 0 {
		variance = 0
	}
	if variance > 1 {
		variance = 1
	}
	t.LuminosityModifier = variance / t.AtmosphericFactor
	if t.LuminosityModifier > 1 {
		t.LuminosityModifier = 1
	}

	// Eccentricity: moons use parent's ecc per spec.
	ecc := body.Eccentricity
	if parent != nil {
		ecc = parent.Eccentricity
	}
	t.NearAU = t.AU * (1 - ecc)
	t.FarAU = t.AU * (1 + ecc)

	// High/Low temperatures (step 9 p.114).
	highL := t.Luminosity * (1 + t.LuminosityModifier)
	lowL := t.Luminosity * (1 - t.LuminosityModifier)
	t.HighK = MeanTemperatureK(highL, t.Albedo, t.GreenhouseFactor, t.NearAU)
	t.LowK = MeanTemperatureK(lowL, t.Albedo, t.GreenhouseFactor, t.FarAU)

	// Worst case (p.115 sidebar): WorstCaseLumModifier = 1 / (1 + bar/2).
	bar := 0.0
	if body.Atmosphere != nil {
		bar = body.Atmosphere.Pressure
	}
	worstMod := 1.0 / (1.0 + bar/2.0)
	if worstMod > 1 {
		worstMod = 1
	}
	worstHighL := t.Luminosity * (1 + worstMod)
	worstLowL := t.Luminosity * (1 - worstMod)
	t.WorstHighK = MeanTemperatureK(worstHighL, t.Albedo, t.GreenhouseFactor, t.NearAU)
	t.WorstLowK = MeanTemperatureK(worstLowL, t.Albedo, t.GreenhouseFactor, t.FarAU)

	// Twilight zone branch (p.120): 1:1 star-locked planets/moons.
	// Moons locked to their parent planet (Case == MoonToPlanet) are NOT
	// twilight zones — book p.105 reserves the term for star locks.
	if body.TidalLock != nil &&
		body.TidalLock.LockRatio == "1:1" &&
		body.TidalLock.Case == TidalLockCasePlanetToStar {

		t.IsTwilight = true
		t.TwilightK = t.MeanK // band centerline (rotation factor = 0)

		// Bright side: rotation factor forced to +1.0.
		brightLumMod := (t.AxialTiltFactor + 1.0 + t.GeographicFactor) / t.AtmosphericFactor
		if brightLumMod > 1 {
			brightLumMod = 1
		}
		brightL := t.Luminosity * (1 + brightLumMod)
		t.BrightSideK = MeanTemperatureK(brightL, t.Albedo, t.GreenhouseFactor, t.NearAU)

		// Dark side: rotation factor forced to -1.0.
		darkLumMod := (t.AxialTiltFactor + (-1.0) + t.GeographicFactor) / t.AtmosphericFactor
		if darkLumMod < 0 {
			darkLumMod = 0
		}
		darkL := t.Luminosity * (1 - darkLumMod)
		t.DarkSideK = MeanTemperatureK(darkL, t.Albedo, t.GreenhouseFactor, t.FarAU)
	}

	return t, nil
}

// totalStellarLuminosity returns the summed luminosity (solar units) of the
// primary's group: primary + any close-binary mate (OrbitClass==OrbitCompanion
// && ParentIndex==-1). Mirrors 3A2a's totalStellarMass pattern from
// surface_tidal_effects.go.
func totalStellarLuminosity(sys stars.System) float64 {
	total := sys.Primary.Luminosity
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
			total += c.Star.Luminosity
		}
	}
	return total
}

// BasicTemperatureRoll rolls 2D + DMs and returns the modified roll plus the
// Kelvin value from the WBH p.109 Basic Mean Temperature table.
//
// DMs (p.109):
//   - Atmosphere DM from p.47 table (via HZRegionAtmosphereDM)
//   - +4 +1 per 0.5 Orbit# below HZCO-1 if Orbit# < HZCO-1
//   - -4 -1 per 0.5 Orbit# above HZCO+1 if Orbit# > HZCO+1
//
// Modified roll above 12: per book "another +50° per result above 12".
// Modified roll below 0: per book "another -5° per result below 0", with
// special recompute "as 1D+5" if value would be < 10K.
func BasicTemperatureRoll(r roller.Roller, body *DetailedPlacement, sys stars.System) (modifiedRoll int, kelvin float64) {
	raw := r.Roll("2D")
	dm := 0

	if body.Atmosphere != nil {
		dm += HZRegionAtmosphereDM(body.Atmosphere.Code)
	}

	hzco := sys.Primary.HZCO()
	if len(body.Group.Members) > 0 {
		hzco = body.Group.HZCO()
	}
	orbit := body.Orbit
	if orbit < hzco-1 {
		dm += 4 + int(math.Floor((hzco-1-orbit)/0.5))
	} else if orbit > hzco+1 {
		dm -= 4 + int(math.Floor((orbit-(hzco+1))/0.5))
	}

	mod := raw + dm

	switch {
	case mod >= 13:
		// Each step above 12 adds +50K to the 388K table top.
		kelvin = 388 + float64(mod-12)*50
	case mod >= 0 && mod <= 12:
		kelvin = basicMeanTemperatureK[mod]
	default:
		// mod < 0 → 178K + 5K per step below 0.
		kelvin = 178 + float64(mod)*5 // mod negative → subtracts
		if kelvin < 10 {
			// Recompute as 1D+5 per p.109 footnote.
			kelvin = float64(r.Roll("1D") + 5)
		}
		if kelvin < 3 {
			kelvin = 3
		}
	}

	return mod, kelvin
}

// computeAxialTiltFactor per WBH p.112-113.
//
// Base: sin(axial_tilt). Tilt clamped to [0°, 90°] (negative → abs; > 90 → 180-tilt).
// Short-year halving: if local year < 0.1 std year, halve (per WBH p.113 — for moons
// "local year" means the moon's orbit period around its planet, NOT the parent's stellar year).
// Long-year boost: if year > 2 std years, +0.01 per std year (max +0.25, cap factor at 1.0).
func computeAxialTiltFactor(body *DetailedPlacement) float64 {
	tilt := 0.0
	if body.AxialTilt != nil {
		tilt = body.AxialTilt.Degrees
		if tilt < 0 {
			tilt = -tilt
		}
		if tilt > 90 {
			tilt = 180 - tilt
		}
	}
	factor := math.Sin(tilt * math.Pi / 180.0)

	// Local year for halving: prefer body.Period.Years; fall back to body.Period.Hours / 8766.
	yrs := body.Period.Years
	if yrs == 0 && body.Period.Hours > 0 {
		yrs = body.Period.Hours / 8766.0
	}
	if yrs > 0 && yrs < 0.1 {
		factor /= 2
	}
	if yrs > 2 {
		boost := 0.01 * yrs
		if boost > 0.25 {
			boost = 0.25
		}
		factor += boost
		if factor > 1.0 {
			factor = 1.0
		}
	}
	return factor
}

// computeRotationFactor per WBH p.113.
//
//	Rotation Factor = √|solar_day_hours| / 50
//
// Exceptions: solar_day > 2500h → 1.0; 1:1 star-locked → 1.0.
func computeRotationFactor(body *DetailedPlacement) float64 {
	if body.DayLength == nil {
		return 0
	}
	if body.TidalLock != nil &&
		body.TidalLock.LockRatio == "1:1" &&
		body.TidalLock.Case == TidalLockCasePlanetToStar {
		return 1.0
	}
	solarH := body.DayLength.SolarHours
	if solarH < 0 {
		solarH = -solarH
	}
	if solarH > 2500 {
		return 1.0
	}
	if solarH == 0 {
		return 0
	}
	return math.Sqrt(solarH) / 50.0
}

// computeGeographicFactor per WBH p.113.
//
//	Geographic Factor = (10 - Hyd) / 20 + modifier
//
// Modifier: ±0.1 if Hyd ∈ [2,8] AND SurfaceDistribution.Description matches
// "Very Concentrated" (+0.1) or "Very Dispersed" (-0.1).
// Note: the p.100 table uses "Very Dispersed" (code 1) as the low extreme;
// the plan draft said "Very Distributed" — corrected here to match actual table values.
func computeGeographicFactor(body *DetailedPlacement) float64 {
	if body.Hydrographics == nil {
		return 0.5 // (10-0)/20 = 0.5 default for missing hydrographics
	}
	hyd := body.Hydrographics.Code
	factor := float64(10-hyd) / 20.0
	if body.SurfaceDistribution != nil && hyd >= 2 && hyd <= 8 {
		switch body.SurfaceDistribution.Description {
		case "Very Concentrated":
			factor += 0.1
		case "Very Dispersed":
			factor -= 0.1
		}
	}
	return factor
}
