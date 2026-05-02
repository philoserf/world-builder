package stars

import (
	"fmt"

	"wbh/roller"
)

// KindFromUnusualCell maps an Unusual-column cell from the Star Type
// Determination table (WBH p. 15) to a StarKind.
func KindFromUnusualCell(cell string) (StarKind, error) {
	switch cell {
	case "BD":
		return KindBrownDwarf, nil
	case "D":
		return KindWhiteDwarf, nil
	default:
		return "", fmt.Errorf("stars: unknown Unusual cell: %q", cell)
	}
}

// KindFromPeculiarCell maps a Peculiar-column cell from the Star Type
// Determination table (WBH p. 15) to a StarKind.
func KindFromPeculiarCell(cell string) (StarKind, error) {
	switch cell {
	case "Black Hole":
		return KindBlackHole, nil
	case "Pulsar":
		return KindPulsar, nil
	case "Neutron Star":
		return KindNeutronStar, nil
	case "Nebula":
		return KindNebula, nil
	case "Protostar":
		return KindProtostar, nil
	case "Star Cluster":
		return KindStarCluster, nil
	case "Anomaly":
		return KindAnomaly, nil
	default:
		return "", fmt.Errorf("stars: unknown Peculiar cell: %q", cell)
	}
}

// RollSpecialPrimarySimple resolves a "Special" primary roll using the
// simple Referee path described on WBH p. 16:
//
//	1D: 1-5 -> Neutron Star, 6 -> Black Hole
//
// The full Unusual/Peculiar dispatch (which can produce brown dwarfs,
// nebulae, protostars, star clusters, anomalies, and so on) lands in
// Plan 2 alongside the rest of the multi-star pipeline.
func RollSpecialPrimarySimple(r roller.Roller) (StarKind, error) {
	roll := r.Roll("1D")
	if roll == 6 {
		return KindBlackHole, nil
	}
	return KindNeutronStar, nil
}

// PeculiarPath selects the column the Referee picks for resolving a
// "Special" (2D=2) primary roll: WBH p.16 lets the Referee choose
// either the Unusual or Peculiar column.
type PeculiarPath string

// PeculiarPath column selector constants.
const (
	PeculiarPathUnusual  PeculiarPath = "unusual"
	PeculiarPathPeculiar PeculiarPath = "peculiar"
)

// RollSpecialPrimary resolves a "Special" (2D=2) primary roll by walking
// the Unusual or Peculiar column of the Star Type Determination table
// (WBH p.15). Returns a StarKind plus, for class redirects ("Class III",
// "Class IV", "Class VI"), the indicated luminosity class.
//
// Cell handling:
//   - "BD" / "D"             → corresponding StarKind (brown/white dwarf)
//   - "Black Hole" / "Pulsar" / "Neutron Star" / "Nebula" / "Protostar" /
//     "Star Cluster" / "Anomaly" → corresponding StarKind
//   - "Class III" / "Class IV" / "Class VI" → caller should re-roll on
//     the regular Star Type Determination table at the indicated class
//   - "Giants" → caller should roll a giant via RollGiantClass and a
//     fresh Type-column roll (see WBH p.16)
//   - "Peculiar" / "Unusual" → re-roll on that column
//
// For cells that require a class-specific re-roll, the function returns
// an empty StarKind ("") and a non-empty LuminosityClass result. For
// final-kind cells, returns the kind with empty class.
//
// Class redirects beyond the simple cases (e.g., the recursive "Peculiar"
// or "Unusual" cell on a re-roll) are handled by recursive calls; recursion
// depth is bounded at 5 to prevent runaway loops on malformed dice.
func RollSpecialPrimary(r roller.Roller, path PeculiarPath) (StarKind, LuminosityClass, error) {
	return rollSpecialPrimaryImpl(r, path, 0)
}

func rollSpecialPrimaryImpl(r roller.Roller, path PeculiarPath, depth int) (StarKind, LuminosityClass, error) {
	if depth > 5 {
		return "", "", fmt.Errorf("stars: special-primary dispatch recursion depth exceeded")
	}
	roll := r.Roll("2D")
	row, ok := StarTypeDetermination[roll]
	if !ok {
		return "", "", fmt.Errorf("stars: 2D out of range: %d", roll)
	}
	var cell string
	switch path {
	case PeculiarPathUnusual:
		cell = row.Unusual
	case PeculiarPathPeculiar:
		cell = row.Peculiar
	default:
		return "", "", fmt.Errorf("stars: unknown peculiar path: %q", path)
	}

	switch cell {
	case "BD":
		return KindBrownDwarf, "", nil
	case "D":
		return KindWhiteDwarf, "", nil
	case "Black Hole":
		return KindBlackHole, "", nil
	case "Pulsar":
		return KindPulsar, "", nil
	case "Neutron Star":
		return KindNeutronStar, "", nil
	case "Nebula":
		return KindNebula, "", nil
	case "Protostar":
		return KindProtostar, "", nil
	case "Star Cluster":
		return KindStarCluster, "", nil
	case "Anomaly":
		return KindAnomaly, "", nil
	case "Class III":
		return "", III, nil
	case "Class IV":
		return "", IV, nil
	case "Class VI":
		return "", VI, nil
	case "Giants":
		return "", "Giants", nil // caller should dispatch via RollGiantClass
	case "Peculiar":
		return rollSpecialPrimaryImpl(r, PeculiarPathPeculiar, depth+1)
	case "Unusual":
		return rollSpecialPrimaryImpl(r, PeculiarPathUnusual, depth+1)
	default:
		return "", "", fmt.Errorf("stars: unexpected peculiar cell: %q", cell)
	}
}
