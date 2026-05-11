package worlds

import "fmt"

// BodyKind classifies a Body in the universe. Per anti-pattern A.1, the
// moon-vs-planet distinction is encoded as a Kind value, not a separate
// type or code path.
type BodyKind int

const (
	// BodyEmpty marks an orbit slot with no world assigned.
	BodyEmpty BodyKind = iota
	// BodyTerrestrial marks a terrestrial planet or moon.
	BodyTerrestrial
	// BodyGasGiant marks a gas giant.
	BodyGasGiant
	// BodyPlanetoidBelt marks a planetoid belt.
	BodyPlanetoidBelt
	// BodyMoon marks a moon. Distinct from BodyTerrestrial because some
	// procedures treat moons differently regardless of whether the
	// underlying body shape is terrestrial.
	BodyMoon
)

// Body is the unified per-body type. One shape for planets, moons, and
// belts. The moon-path silent-zero anti-pattern (anti-pattern A.1) is
// prevented at the type level — moons are walked by the same iterator
// as planets, with Kind / Parent telling per-procedure code which
// rules apply.
//
// Pointer fields are nil when the Stage that populates them does not
// apply to this body. Has*() predicates wrap the nil checks.
type Body struct {
	// Identity
	Designation string
	Kind        BodyKind
	Parent      *Body // nil for planets and belts; non-nil for moons

	// Stage 1 (placement) — for planets and belts; moons inherit via Parent.
	Group        Group // Cycle 0 uses worlds.Group; api-surface.md § Open
	Orbit        float64
	Eccentricity float64
	HZ           bool

	// Stage 2 (sizing/period)
	SizeCode       SizeCode
	DiameterKm     float64
	GGClass        GasGiantClass // NotGasGiant for non-GG bodies
	GGDiameterCode string        // GG only — eHex code matching DiameterEarth
	DiameterEarth  float64       // GG only
	MassEarth      float64
	Period         Period

	// Moon-specific (Kind == BodyMoon)
	OrbitPD     float64
	OrbitKm     float64
	PeriodHours float64
	Retrograde  bool // moon orbits its parent retrograde (anomalous slot)

	// Stage 3
	Physical *BodyPhysical
	Belt     *BeltDetails

	// Stage 4
	DayLength    *DayLength
	AxialTilt    *AxialTilt
	TidalLock    *TidalLock
	TidalEffects *SurfaceTidalEffects

	// Stage 5 — populated post-ConvergeClimate
	Atmosphere    *Atmosphere
	Hydrographics *Hydrographics
	Temperature   *Temperature

	// Stage 6
	SurfaceDistribution *SurfaceDistribution

	// Stages 7–9
	Geology      *Geology
	Biology      *Biology
	Habitability *Habitability

	// Sub-bodies (moons). Iterator descends into Children after the parent.
	Children []*Body
}

// HasPhysical reports whether body physical (composition / density /
// gravity) has been populated for this body.
func (b *Body) HasPhysical() bool { return b.Physical != nil }

// HasBelt reports whether belt-detail fields are populated.
func (b *Body) HasBelt() bool { return b.Belt != nil }

// HasDayLength reports whether the rotation pass has run for this body.
func (b *Body) HasDayLength() bool { return b.DayLength != nil }

// HasAxialTilt reports whether the axial-tilt pass has run.
func (b *Body) HasAxialTilt() bool { return b.AxialTilt != nil }

// HasTidalLock reports whether tidal-lock state is recorded for this body.
func (b *Body) HasTidalLock() bool { return b.TidalLock != nil }

// HasTidalEffects reports whether the surface tidal-effects pass has run.
func (b *Body) HasTidalEffects() bool { return b.TidalEffects != nil }

// HasAtmosphere reports whether climate convergence produced an atmosphere.
func (b *Body) HasAtmosphere() bool { return b.Atmosphere != nil }

// HasHydrographics reports whether climate convergence produced hydrographics.
func (b *Body) HasHydrographics() bool { return b.Hydrographics != nil }

// HasTemperature reports whether climate convergence produced a temperature.
func (b *Body) HasTemperature() bool { return b.Temperature != nil }

// HasSurfaceDistribution reports whether the surface-distribution pass has run.
func (b *Body) HasSurfaceDistribution() bool { return b.SurfaceDistribution != nil }

// HasGeology reports whether the geology pass has run for this body.
func (b *Body) HasGeology() bool { return b.Geology != nil }

// HasBiology reports whether the biology pass has run for this body.
func (b *Body) HasBiology() bool { return b.Biology != nil }

// HasHabitability reports whether the habitability rating is computed.
func (b *Body) HasHabitability() bool { return b.Habitability != nil }

// RenderSAH returns the 3-character SAH triplet for the IISS form.
// HZ bodies get the full triplet; non-HZ bodies render as "<Size>??".
func (b *Body) RenderSAH() string {
	size := string(b.SizeCode)
	if size == "" {
		size = "?"
	}
	if !b.HasAtmosphere() || !b.HasHydrographics() {
		return size + "??"
	}
	atmoChar := atmosphereCodeChar(b.Atmosphere.Code)
	hydroChar := fmt.Sprintf("%d", b.Hydrographics.Code)
	if b.Hydrographics.Code == 10 {
		hydroChar = "A"
	}
	return size + atmoChar + hydroChar
}

// atmosphereCodeChar lives in atmosphere_profile.go (cycle 5).

// StellarOrbit returns the body's orbit around its host star. For
// moons, that is the parent planet's orbit; for planets and belts, the
// body's own Orbit. Resolves the pass-1 confusion where moon code
// "happened to work" because some procedures aliased through dp.Orbit
// and others didn't (anti-pattern memo per docs/pass-2/api-surface.md
// § The Body trade-off).
func (b *Body) StellarOrbit() float64 {
	if b.Kind == BodyMoon && b.Parent != nil {
		return b.Parent.Orbit
	}
	return b.Orbit
}
