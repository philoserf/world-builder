package worlds

import (
	"fmt"

	"github.com/philoserf/world-builder/stars"
)

// SurfaceTidalEffects holds the computed tidal amplitude for a body,
// broken into named contributions. All values are in metres.
type SurfaceTidalEffects struct {
	// Components lists each tidal source with its contribution.
	Components []TidalComponent
	// Total is the sum of all component Meters values.
	Total float64
}

// TidalComponent is one source's contribution to surface tides.
type TidalComponent struct {
	// Source is a human-readable label: "planet <designation>",
	// "star <designation>", or "moon <designation>".
	Source string
	Meters float64
}

// StarTide returns the tidal amplitude (metres) on a planet from a star,
// per WBH p.107:
//
//	Amplitude = StarMass × PlanetSize / (32 × AU³)
//
// starMassSol is the star's mass in solar units; planetSizeN is the
// numeric planet Size (0-15); auFromStar is the distance to the star in AU.
func StarTide(starMassSol float64, planetSizeN int, auFromStar float64) float64 {
	if auFromStar == 0 {
		return 0
	}
	return starMassSol * float64(planetSizeN) / (32.0 * auFromStar * auFromStar * auFromStar)
}

// MoonTideOnPlanet returns the tidal amplitude (metres) that a moon
// exerts on its parent planet, per WBH p.108:
//
//	Amplitude = MoonMass × PlanetSize / (3.2 × (OrbitKm/1,000,000)³)
//
// moonMassEarth is the moon's mass in Earth units; planetSizeN is the
// numeric planet Size (0-15); orbitKm is the moon's orbital distance in km.
func MoonTideOnPlanet(moonMassEarth float64, planetSizeN, orbitKm int) float64 {
	distMkm := float64(orbitKm) / 1_000_000.0
	if distMkm == 0 {
		return 0
	}
	return moonMassEarth * float64(planetSizeN) / (3.2 * distMkm * distMkm * distMkm)
}

// PlanetTideOnMoon returns the tidal amplitude (metres) that a parent
// planet exerts on one of its moons, per WBH p.108:
//
//	Amplitude = PlanetMass × MoonSize / (3.2 × (OrbitKm/1,000,000)³)
//
// planetMassEarth is the parent planet's mass in Earth units; moonSizeN is
// the numeric moon Size (0-15); orbitKm is the moon's orbital distance in km.
func PlanetTideOnMoon(planetMassEarth float64, moonSizeN, orbitKm int) float64 {
	distMkm := float64(orbitKm) / 1_000_000.0
	if distMkm == 0 {
		return 0
	}
	return planetMassEarth * float64(moonSizeN) / (3.2 * distMkm * distMkm * distMkm)
}

// GenerateSurfaceTidalEffects computes the total surface tidal amplitude
// (in metres) for a body per WBH pp.107-108.
//
// body is the target body (planet or moon); moonRef is non-nil when body
// is a moon orbiting parentPlanet; sys provides the stellar context; and
// parentPlanet is non-nil when body is a moon (it is the planet the moon
// orbits).
//
// Three tidal sources are summed:
//  1. Parent planet → moon (PlanetTideOnMoon), when body is a moon and the
//     moon is not 1:1 locked to its parent (a 1:1 lock fixes the tidal bulge,
//     producing no surface amplitude per WBH p.108).
//  2. Stars → body (StarTide), summing mass within each close-binary group.
//  3. Sibling moons → body (MoonTideOnPlanet), using each sibling moon's
//     mass and orbit (moon-to-moon tides not yet implemented — Referee fiat per p.108).
//
// Returns an error if required fields are zero or missing.
func GenerateSurfaceTidalEffects(
	body *Body,
	moonRef *Body,
	sys stars.System,
	parentPlanet *Body,
) (*SurfaceTidalEffects, error) {
	if body == nil {
		return nil, fmt.Errorf("worlds: GenerateSurfaceTidalEffects: body is nil")
	}
	// Empty orbit slots have no tidal surface; skip per spec.
	if body.Kind == BodyEmpty {
		return nil, nil
	}

	bodySizeN := nForSizeCode(body.SizeCode)
	var components []TidalComponent

	// ── 1. Parent planet → moon ───────────────────────────────────────────
	// Applied when body is a moon (moonRef != nil) and parentPlanet is provided.
	// Skipped when the moon is 1:1 tidally locked: a fixed bulge yields no surface tide.
	isOneToOneLocked := moonRef != nil && moonRef.TidalLock != nil && moonRef.TidalLock.LockRatio == "1:1"
	if moonRef != nil && parentPlanet != nil && !isOneToOneLocked {
		planetMass := parentPlanet.MassEarth
		if moonRef.OrbitKm > 0 && planetMass > 0 {
			tide := PlanetTideOnMoon(planetMass, bodySizeN, int(moonRef.OrbitKm))
			label := "planet"
			if parentPlanet.Designation != "" {
				label = fmt.Sprintf("planet %s", parentPlanet.Designation)
			}
			components = append(components, TidalComponent{Source: label, Meters: tide})
		}
	}

	// ── 2. Stars → body ───────────────────────────────────────────────────
	// Per WBH p.107-108: for each group of stars, sum the masses of the
	// primary and any OrbitCompanion-class (close-binary) members with
	// ParentIndex == -1, then compute a single StarTide at the group AU.
	//
	// The planet's orbit in AU is used as the distance to all star groups,
	// matching the WBH worked example ("from the two relatively distant
	// suns … at 1.06 AU").
	auFromStar := body.Orbit
	if moonRef != nil && parentPlanet != nil {
		// Moon orbits a planet; use the planet's stellar orbit.
		auFromStar = parentPlanet.Orbit
	}

	// Sum primary group: primary mass + OrbitCompanion-class mates (sub-AU pairs
	// like Aab) sharing ParentIndex == -1. OrbitClose-class companions (0.05-0.5 AU)
	// are NOT summed; they form their own group below. ParentIndex values are slice
	// indices into sys.Companions as produced by stars.GenerateSystem.
	primaryGroupMass := sys.Primary.Mass
	for _, c := range sys.Companions {
		if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
			primaryGroupMass += c.Star.Mass
		}
	}
	primaryLabel := sys.PrimaryDesignation
	if primaryLabel == "" {
		primaryLabel = "primary"
	}
	if primaryGroupMass > 0 {
		tide := StarTide(primaryGroupMass, bodySizeN, auFromStar)
		components = append(components, TidalComponent{
			Source: fmt.Sprintf("star %s", primaryLabel),
			Meters: tide,
		})
	}

	// Other companion groups (non-close-binary companions: Near, Far, etc.).
	// Each such companion forms its own group with its own close-binary subset.
	for i, c := range sys.Companions {
		// Skip companions that are themselves close-binary members of the primary group.
		if c.OrbitClass == stars.OrbitCompanion && c.ParentIndex == -1 {
			continue
		}
		// Also skip sub-companions (those orbiting another companion, ParentIndex >= 0)
		// because they will be summed into their parent companion's group below.
		if c.ParentIndex >= 0 {
			continue
		}
		// This companion is the "root" of its own group.
		groupMass := c.Star.Mass
		// Add any OrbitCompanion members parenting from this companion.
		for _, sub := range sys.Companions {
			if sub.OrbitClass == stars.OrbitCompanion && sub.ParentIndex == i {
				groupMass += sub.Star.Mass
			}
		}
		label := c.Designation
		if label == "" {
			label = fmt.Sprintf("companion%d", i)
		}
		tide := StarTide(groupMass, bodySizeN, auFromStar)
		components = append(components, TidalComponent{
			Source: fmt.Sprintf("star %s", label),
			Meters: tide,
		})
	}

	// ── 3. Sibling moon → body ────────────────────────────────────────────
	// Moon-to-moon tides: Referee fiat per WBH p.108 — not yet implemented.
	// (Sibling moon contributions are negligible in most cases.)

	// ── Total ─────────────────────────────────────────────────────────────
	total := 0.0
	for _, c := range components {
		total += c.Meters
	}

	return &SurfaceTidalEffects{
		Components: components,
		Total:      total,
	}, nil
}
