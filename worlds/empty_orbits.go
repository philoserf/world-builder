package worlds

import "wbh/roller"

// RollEmptyOrbits implements WBH Step 4 (p. 48). Returns the number of
// extra orbital slots to insert across the system.
func RollEmptyOrbits(r roller.Roller) (int, error) {
	switch v := r.Roll("2D"); {
	case v <= 9:
		return 0, nil
	case v == 10:
		return 1, nil
	case v == 11:
		return 2, nil
	default:
		return 3, nil
	}
}
