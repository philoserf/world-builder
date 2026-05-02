package stars

import "wbh/roller"

// OrbitClass is the WBH p.23 orbit class for a non-primary star.
type OrbitClass string

// OrbitClass constants (WBH p.23).
const (
	OrbitClose     OrbitClass = "close"
	OrbitNear      OrbitClass = "near"
	OrbitFar       OrbitClass = "far"
	OrbitCompanion OrbitClass = "companion"
)

// PresenceDM returns the WBH p.23 DM applied to all multi-star presence
// rolls for a system whose primary has the given properties.
func PresenceDM(primary Star) int {
	// Special-object primaries: BD, D, and post-stellar all -1.
	switch primary.Kind {
	case KindBrownDwarf, KindWhiteDwarf, KindPulsar, KindNeutronStar, KindBlackHole:
		return -1
	}
	// Class Ia/Ib/II/III/IV: +1.
	switch primary.LuminosityClass {
	case Ia, Ib, II, III, IV:
		return 1
	}
	// Class V or VI primary by spectral letter:
	switch primary.LuminosityClass {
	case V, VI:
		switch primary.SpectralType.Letter {
		case 'O', 'B', 'A', 'F':
			return 1
		case 'M':
			return -1
		}
	}
	return 0
}

// RollPresence rolls 2D + PresenceDM(primary) and returns true if the
// result meets the WBH p.23 threshold (10+).
//
// Class Ia/Ib/II/III primaries cannot have Close secondaries (WBH p.23);
// in those cases the function returns false without consuming a roll.
func RollPresence(r roller.Roller, primary Star, oc OrbitClass) bool {
	if oc == OrbitClose {
		switch primary.LuminosityClass {
		case Ia, Ib, II, III:
			return false
		}
	}
	natural := r.Roll("2D")
	return natural+PresenceDM(primary) >= MultipleStarsPresenceThreshold
}
