package worlds

import (
	"fmt"

	"wbh/roller"
)

// BodyPhysical — terrestrial body or moon body physical characteristics, WBH pp. 69–72.
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
