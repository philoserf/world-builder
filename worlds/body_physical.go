package worlds

import (
	"fmt"
	"math"

	"github.com/philoserf/world-builder/roller"
)

// BodyPhysical — physical characteristics per WBH pp.71-78.
//
// Density, Gravity, Mass, DiameterKm are 3A1 outputs and are NOT
// temperature-sensitive; they remain stable across the 5D rederive pass.
type BodyPhysical struct {
	Composition     string  // "Exotic Ice"|"Mostly Ice"|"Mostly Rock"|"Rock and Metal"|"Mostly Metal"|"Compressed Metal"
	Density         float64 // relative to Terra (1.0 = 5.514 g/cm³)
	Gravity         float64 // relative to Terra (1.0 G)
	EscapeVelocity  float64 // m/s
	OrbitalVelocity float64 // m/s at surface
	SizeProfile     string  // "S-Dkm-D-G-M", e.g. "5-8163-1.03-0.66-0.27"
}

// BodyPhysicalDMs — DM accumulators for the Composition roll, per WBH p. 71.
type BodyPhysicalDMs struct {
	SizeCode       SizeCode // for Size 0-4 DM-1 / Size 6-9 DM+1 / Size A-F DM+3
	AtHZCOOrCloser bool     // DM+1
	BeyondHZCO     int      // DM-1 per full Orbit#
	SystemAgeGyr   float64  // DM-1 if > 10 Gyr
}

func compositionDM(d BodyPhysicalDMs) int {
	dm := 0
	switch d.SizeCode {
	// "0" and "R" encode the full WBH p.71 table row but are unreachable
	// through the Stage-3 orchestrator, which early-returns for those
	// codes before building BodyPhysicalDMs; they are exercised only by
	// direct unit tests.
	case "0", "R", "S", "1", "2", "3", "4":
		dm--
	case "6", "7", "8", "9":
		dm++
	case "A", "B", "C", "D", "E", "F":
		dm += 3
	}
	if d.AtHZCOOrCloser {
		dm++
	}
	dm -= d.BeyondHZCO
	if d.SystemAgeGyr > 10 {
		dm--
	}
	return dm
}

// RollComposition rolls 2D + DMs and returns the composition column from the Terrestrial Composition table (WBH p. 71).
func RollComposition(r roller.Roller, dms BodyPhysicalDMs) (string, error) {
	base := r.Roll("2D")
	total := base + compositionDM(dms)
	switch {
	case total <= -4:
		return "Exotic Ice", nil
	case total <= -2:
		return "Mostly Ice", nil
	case total <= 6:
		return "Mostly Rock", nil
	case total <= 11:
		return "Rock and Metal", nil
	case total <= 14:
		return "Mostly Metal", nil
	default:
		return "Compressed Metal", nil
	}
}

// densityTable — WBH p. 71 Terrestrial Density: rows = 2D 2..12, columns = composition.
var densityTable = map[string][11]float64{
	//                       2D=2   3      4      5      6      7      8      9      10     11     12
	"Exotic Ice":       {0.03, 0.06, 0.09, 0.12, 0.15, 0.18, 0.21, 0.24, 0.27, 0.30, 0.33},
	"Mostly Ice":       {0.18, 0.21, 0.24, 0.27, 0.30, 0.33, 0.36, 0.39, 0.41, 0.44, 0.47},
	"Mostly Rock":      {0.50, 0.53, 0.56, 0.59, 0.62, 0.65, 0.68, 0.71, 0.74, 0.77, 0.80},
	"Rock and Metal":   {0.82, 0.85, 0.88, 0.91, 0.94, 0.97, 1.00, 1.03, 1.06, 1.09, 1.12},
	"Mostly Metal":     {1.15, 1.18, 1.21, 1.24, 1.27, 1.30, 1.33, 1.36, 1.39, 1.42, 1.45},
	"Compressed Metal": {1.50, 1.55, 1.60, 1.65, 1.70, 1.75, 1.80, 1.85, 1.90, 1.95, 2.00},
}

// RollDensity rolls 2D and returns the density value from the Terrestrial Density table column for the given composition (WBH p. 71).
func RollDensity(r roller.Roller, composition string) (float64, error) {
	roll := r.Roll("2D")
	row, ok := densityTable[composition]
	if !ok {
		return 0, fmt.Errorf("worlds: density: unknown composition %q", composition)
	}
	if roll < 2 {
		roll = 2
	}
	if roll > 12 {
		roll = 12
	}
	return row[roll-2], nil
}

// DiameterTerra is Terra's diameter in km, used as the reference for size-relative calculations (WBH p. 71).
const DiameterTerra = 12742.0

// DeriveGravity returns surface gravity in G: (Density × Diameter) / DiameterTerra.
func DeriveGravity(densityRel, diameterKm float64) float64 {
	return densityRel * diameterKm / DiameterTerra
}

// DeriveMass returns mass in M⊕: Density × (Diameter / DiameterTerra)³.
func DeriveMass(densityRel, diameterKm float64) float64 {
	ratio := diameterKm / DiameterTerra
	return densityRel * ratio * ratio * ratio
}

// DeriveEscapeVelocity returns escape velocity in m/s: √(m / (D/D⊕)) × 11,186.
func DeriveEscapeVelocity(massEarth, diameterKm float64) float64 {
	if diameterKm <= 0 {
		return 0
	}
	ratio := massEarth / (diameterKm / DiameterTerra)
	if ratio < 0 {
		return 0
	}
	return math.Sqrt(ratio) * 11186
}

// DeriveOrbitalVelocity returns surface orbital velocity in m/s: EscV / √2.
func DeriveOrbitalVelocity(escapeVelocity float64) float64 {
	return escapeVelocity / math.Sqrt(2)
}

// FormatSizeProfile formats the size profile string "S-Dkm-D-G-M".
func FormatSizeProfile(p BodyPhysical, massEarth float64, sizeCode SizeCode, diameterKm int) string {
	return fmt.Sprintf("%s-%d-%.2f-%.2f-%.2f",
		string(sizeCode), diameterKm, p.Density, p.Gravity, massEarth)
}

// GenerateBodyPhysical orchestrates the per-body physical pipeline (WBH pp. 71-72).
// The caller pre-resolves diameterKm from the Size table; this function consumes
// the diameter dice (D3, 1D, d100) from the roller script for trace alignment,
// then rolls composition and density, and derives gravity, mass, and velocities.
func GenerateBodyPhysical(r roller.Roller, sizeCode SizeCode, diameterKm int, dms BodyPhysicalDMs) (BodyPhysical, error) {
	// Consume diameter dice — values already reflected in diameterKm.
	r.Roll("D3")
	r.Roll("1D")
	r.Roll("d100")

	comp, err := RollComposition(r, dms)
	if err != nil {
		return BodyPhysical{}, err
	}
	density, err := RollDensity(r, comp)
	if err != nil {
		return BodyPhysical{}, err
	}
	p := BodyPhysical{
		Composition: comp,
		Density:     density,
	}
	p.Gravity = DeriveGravity(density, float64(diameterKm))
	mass := DeriveMass(density, float64(diameterKm))
	p.EscapeVelocity = DeriveEscapeVelocity(mass, float64(diameterKm))
	p.OrbitalVelocity = DeriveOrbitalVelocity(p.EscapeVelocity)
	p.SizeProfile = FormatSizeProfile(p, mass, sizeCode, diameterKm)
	return p, nil
}
