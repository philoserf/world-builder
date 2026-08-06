package worlds

import "github.com/philoserf/world-builder/stars"

// HZCO returns the group's Habitable Zone Centre Orbit#.
//
// Single-star group: delegates to Members[0].HZCO().
// Pair group: delegates to stars.CompositeHZCO(Members...) per WBH p. 42,
// which sums the constituent luminosities before applying the formula.
//
// Source: WBH pp. 41–42.
func (g Group) HZCO() float64 {
	if len(g.Members) == 1 {
		return g.Members[0].HZCO()
	}

	return stars.CompositeHZCO(g.Members...)
}
