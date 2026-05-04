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
// multi-source IR addition).
//
// Currently populates: Luminosity, AU, ScaleHeight, Albedo, GreenhouseFactor,
// MeanK, BasicK (Task 6); AxialTiltFactor, RotationFactor, GeographicFactor,
// AtmosphericFactor, LuminosityModifier, NearAU, FarAU, HighK, LowK,
// WorstHighK, WorstLowK (Task 7); IsTwilight, TwilightK, BrightSideK,
// DarkSideK for 1:1 star-locked bodies (Task 8); ParentRadianceK plus
// elevated MeanK/HighK/LowK for moons of warm gas giants (Task 9).
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

	// Multi-source addition for moons (p.111, p.125-126): parent body's IR
	// contribution combines with stellar temperature via ⁴√(T₁⁴ + T₂⁴).
	// Pragmatic MVP threshold: skip unless parent's MeanK exceeds moon's
	// stellar-only MeanK by 30K — cold gas giants contribute negligibly.
	// Variance components (axial tilt, rotation, etc.) describe stellar
	// variability only and are NOT modified.
	if parent != nil && parent.Temperature != nil {
		tParent := parent.Temperature.MeanK
		if tParent > t.MeanK+30 {
			t.ParentRadianceK = tParent
			t.MeanK = CombineTemperatures(t.MeanK, tParent)
			t.HighK = CombineTemperatures(t.HighK, tParent)
			t.LowK = CombineTemperatures(t.LowK, tParent)
		}
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

// SunlightPortion computes the fraction of a solar day with sunlight at a
// given latitude, axial tilt, and date per WBH p.118. Returns 1.0 for polar
// day, 0.0 for polar night, otherwise the sunrise-angle fraction.
//
// Callers wanting daylight hours multiply the result by their solarDayHours.
func SunlightPortion(latDeg, axialTiltDeg, daysSinceSolstice, localYearDays float64) float64 {
	if localYearDays <= 0 {
		return 0
	}
	// Solar declination (in radians) = axial_tilt × cos(360° × date / year).
	declRad := axialTiltDeg * math.Cos(2*math.Pi*daysSinceSolstice/localYearDays) * math.Pi / 180.0

	tanLat := math.Tan(latDeg * math.Pi / 180.0)
	tanDecl := math.Tan(declRad)
	cosSunrise := -tanLat * tanDecl

	switch {
	case cosSunrise > 1:
		return 0 // polar night
	case cosSunrise < -1:
		return 1.0 // polar day
	default:
		return math.Acos(cosSunrise) / math.Pi
	}
}

// MeanByLatitude returns the annual mean temperature at a specific latitude
// per WBH p.116-117, ignoring season and time of day.
//
// Twilight worlds short-circuit to TwilightK (no meaningful latitude variation
// when one hemisphere is in perpetual day and the other in perpetual night;
// caller should use BrightSideK/DarkSideK directly for hemisphere selection).
func (t *Temperature) MeanByLatitude(latDeg float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	zoneTiltFactor := t.zoneTiltAdjustment(latDeg)
	lumMod := zoneTiltFactor / t.AtmosphericFactor
	if lumMod > 1 {
		lumMod = 1
	}
	if lumMod < 0 {
		lumMod = 0
	}
	latLum := t.Luminosity * (1 + lumMod)
	return MeanTemperatureK(latLum, t.Albedo, t.GreenhouseFactor, t.AU)
}

// zoneTiltAdjustment returns the latitude-zone-adjusted axial-tilt-equivalent
// factor per WBH p.116-117 three-zone classification (tropical / middle /
// arctic). For axial tilt ≥ 45° the middle zone disappears (Part B p.117).
func (t *Temperature) zoneTiltAdjustment(latDeg float64) float64 {
	tiltDeg := math.Asin(t.AxialTiltFactor) * 180.0 / math.Pi
	if math.IsNaN(tiltDeg) {
		// |AxialTiltFactor| > 1 — clamp.
		if t.AxialTiltFactor > 0 {
			tiltDeg = 90
		} else {
			tiltDeg = 0
		}
	}
	if latDeg < 0 {
		latDeg = -latDeg
	}
	if latDeg > 90 {
		latDeg = 90
	}

	switch {
	case latDeg <= tiltDeg:
		// Tropical zone: sin(45° - axial_tilt) replaces axial tilt factor.
		adj := 45.0 - tiltDeg
		if adj < 0 {
			adj = 0
		}
		return math.Sin(adj * math.Pi / 180.0)
	case tiltDeg >= 45 && latDeg < (90-tiltDeg):
		// Part B: no middle zone; use arctic-edge result at lat=90-tilt.
		return math.Sin((45.0 - (90.0 - tiltDeg)) * math.Pi / 180.0)
	default:
		// Middle/arctic: sin(45° - latitude).
		return math.Sin((45.0 - latDeg) * math.Pi / 180.0)
	}
}

// MeanBySeason returns the mean temperature on a specific day at a specific
// latitude, ignoring time of day, per WBH p.115.
//
// daysSinceSolstice: 0 = summer solstice in the relevant hemisphere; year/2 = winter solstice.
// localYearDays: caller decides — for moons, use parent's stellar year (moons co-orbit star with planet).
//
// Twilight worlds short-circuit to TwilightK.
func (t *Temperature) MeanBySeason(latDeg, daysSinceSolstice, localYearDays float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	if localYearDays <= 0 {
		return t.MeanByLatitude(latDeg)
	}

	// Adjusted Fractional Year per WBH p.115.
	stdYearDays := 8766.0 / 24.0 // 365.25
	lagDays := 0.1 * math.Min(stdYearDays, localYearDays)
	adjFracYear := (daysSinceSolstice - 0.1*lagDays) / localYearDays

	// Seasonal axial tilt factor: cos(adjFracYear × 360°) × stored AxialTiltFactor.
	// Positive = summer (sun higher in sky → more heat); negative = winter (less heat).
	// Apply directly as a signed luminosity modifier: the axial-tilt contribution
	// swings from +AxialTiltFactor at summer solstice to -AxialTiltFactor at winter.
	seasonalTilt := math.Cos(adjFracYear*2*math.Pi) * t.AxialTiltFactor

	lumMod := seasonalTilt / t.AtmosphericFactor
	if lumMod > 1 {
		lumMod = 1
	}
	if lumMod < -1 {
		lumMod = -1
	}
	latLum := t.Luminosity * (1 + lumMod)
	if latLum < 0 {
		latLum = 0
	}
	return MeanTemperatureK(latLum, t.Albedo, t.GreenhouseFactor, t.AU)
}

// AtMoment returns the instantaneous temperature at a specific moment per WBH p.117.
//
// Uses Method 1 (even-length days) internally; callers wanting Method 2
// precision should call SunlightPortion separately and modulate hoursSinceDawn.
//
// Hourly variation follows a cosine curve centered on the peak-heat time, which
// lags 15% of the solar day past solar noon (i.e., peak at 65% of the day from
// dawn). Dawn is the coolest point, peak is shortly after noon.
//
// Twilight worlds short-circuit to TwilightK.
func (t *Temperature) AtMoment(latDeg, daysSinceSolstice, localYearDays, hoursSinceDawn, solarDayHours float64) float64 {
	if t.IsTwilight {
		return t.TwilightK
	}
	if solarDayHours <= 0 {
		return t.MeanBySeason(latDeg, daysSinceSolstice, localYearDays)
	}

	// Seasonal contribution.
	seasonalK := t.MeanBySeason(latDeg, daysSinceSolstice, localYearDays)

	// Hourly rotation factor: cosine centered on peak-heat time (65% of solar day).
	// cos(0) = 1 at the hottest point; cos(π) = -1 at dawn (coolest).
	// peakFrac = 0.65 (noon 0.50 + lag 0.15).
	peakFrac := 0.65
	fracDay := hoursSinceDawn / solarDayHours
	hourlyRot := math.Cos(2*math.Pi*(fracDay-peakFrac)) * t.RotationFactor

	// Apply hourly rotation as a luminosity-modifier delta scaled into K via fourth-root.
	delta := hourlyRot / t.AtmosphericFactor
	scale := math.Pow(1+delta, 0.25)
	if scale <= 0 {
		scale = 0.01
	}
	return seasonalK * scale
}

// AdjustedForAltitude returns a temperature adjusted for altitude per WBH
// p.123-124. The greenhouse factor is reduced because atmospheric pressure
// drops with altitude (exp(-altitude/scale_height)). The implementation
// recomputes the equation with the modified greenhouse factor, then scales
// baseTempK by the ratio newRefK/storedMeanK to preserve any caller-applied
// scenario adjustments (latitude, season, etc.).
//
// Uses the cached t.ScaleHeight (populated from body.Atmosphere.ScaleHeight
// at GenerateTemperature time). Returns baseTempK unchanged when altitudeKm
// is zero or t.ScaleHeight is zero (e.g., vacuum world or atmosphere with
// no scale-height data).
//
// This is an approximation — the book's full altitude treatment includes
// lapse-rate and density-gradient effects deferred beyond 3A2b-temp.
func (t *Temperature) AdjustedForAltitude(baseTempK, altitudeKm float64) float64 {
	if altitudeKm <= 0 || t.ScaleHeight <= 0 {
		return baseTempK
	}
	// Pressure scales as e^(-h/H). Greenhouse factor scales with √(pressure).
	pressureRatio := math.Exp(-altitudeKm / t.ScaleHeight)
	gAtAlt := t.GreenhouseFactor * math.Sqrt(pressureRatio)

	// Recompute mean equation with modified G, using stored A/L/AU. Then scale
	// baseTempK by the ratio to preserve caller-applied scenario adjustments.
	newRefK := MeanTemperatureK(t.Luminosity, t.Albedo, gAtAlt, t.AU)
	storedRefK := MeanTemperatureK(t.Luminosity, t.Albedo, t.GreenhouseFactor, t.AU)
	if storedRefK == 0 {
		return baseTempK
	}
	return baseTempK * (newRefK / storedRefK)
}
