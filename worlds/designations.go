package worlds

import (
	"fmt"
	"strings"
)

// AssignPlanetDesignations walks bodies in arrival order (orbit order
// per group, established by Stage 1's PlaceWorlds) and assigns:
//
//   - Non-belt placements:    "<Group> I", "<Group> II", ... (Roman planet counter)
//   - Planetoid belts:        "<Group> PI", "<Group> PII", ... (separate belt counter)
//
// Per WBH p.53 sidebar:
//   - Planet enumeration skips belts (planet counter never advances on belt).
//   - Each new group resets both counters to I.
//   - Empty placements get no designation (left empty).
//
// Mutates Body.Designation in place.
func AssignPlanetDesignations(bodies []Body) {
	currentGroup := ""
	planetN := 0
	beltN := 0
	for i := range bodies {
		gd := bodies[i].Group.Designation
		if gd != currentGroup {
			currentGroup = gd
			planetN = 0
			beltN = 0
		}
		switch bodies[i].Kind {
		case BodyEmpty:
			// No designation for empty slots.
		case BodyPlanetoidBelt:
			beltN++
			bodies[i].Designation = fmt.Sprintf("%s P%s", gd, romanNumeral(beltN))
		default: // BodyTerrestrial, BodyGasGiant
			planetN++
			bodies[i].Designation = fmt.Sprintf("%s %s", gd, romanNumeral(planetN))
		}
	}
}

// AssignMoonDesignations walks each body's Children in arrival order
// (closest-to-farthest from the parent, per CountMoons/SizeMoon
// emission order) and assigns "<Parent> a", "<Parent> b", ... per
// WBH p.58 sidebar.
//
// Insignificant moons are out of scope.
//
// Mutates each child Body.Designation in place via the pointer.
func AssignMoonDesignations(bodies []Body) {
	for i := range bodies {
		for j := range bodies[i].Children {
			letter := byte('a' + j)
			bodies[i].Children[j].Designation = fmt.Sprintf("%s %c", bodies[i].Designation, letter)
		}
	}
}

// romanNumeral returns the Roman-numeral representation of n for n in
// [1, 39]. WBH planet/belt enumeration never approaches that bound
// (Zed has 8 worlds in Aab, the largest in book examples).
func romanNumeral(n int) string {
	if n < 1 {
		return ""
	}
	values := []int{10, 9, 5, 4, 1}
	symbols := []string{"X", "IX", "V", "IV", "I"}
	var out strings.Builder
	for i, v := range values {
		for n >= v {
			out.WriteString(symbols[i])
			n -= v
		}
	}
	return out.String()
}

// MarkHZ sets Body.HZ when the body's orbit lies within HZCO ± 1.0 of
// the body's host group, per WBH p.58: "The habitable zone is generally
// considered to be +/ 1.0 Orbit#s from the HZCO."
//
// Uses Group.HZCO() from Stage 1 (group_hzco.go).
//
// Mutates Body.HZ in place. Empty bodies are skipped.
func MarkHZ(bodies []Body) {
	for i := range bodies {
		if bodies[i].Kind == BodyEmpty {
			continue
		}
		hzco := bodies[i].Group.HZCO()
		o := bodies[i].Orbit
		if o >= hzco-1.0 && o <= hzco+1.0 {
			bodies[i].HZ = true
		}
	}
}
