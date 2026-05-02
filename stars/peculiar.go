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
