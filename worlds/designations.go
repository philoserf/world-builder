package worlds

import "strings"

import "fmt"

// AssignPlanetDesignations walks Detailed in arrival order (orbit order
// per group, established by 2B's PlaceWorlds) and assigns:
//
//   - Non-belt placements:    "<Group> I", "<Group> II", ... (Roman planet counter)
//   - Planetoid belts:        "<Group> PI", "<Group> PII", ... (separate belt counter)
//
// Per WBH p.53 sidebar:
//   - Planet enumeration skips belts (planet counter never advances on belt).
//   - Each new group resets both counters to I.
//   - Empty placements get no designation (left empty).
//
// Mutates DetailedPlacement.Designation in place.
func AssignPlanetDesignations(dps []DetailedPlacement) {
	currentGroup := ""
	planetN := 0
	beltN := 0
	for i := range dps {
		gd := dps[i].Group.Designation
		if gd != currentGroup {
			currentGroup = gd
			planetN = 0
			beltN = 0
		}
		switch dps[i].Body {
		case BodyEmpty:
			// No designation for empty slots.
		case BodyPlanetoidBelt:
			beltN++
			dps[i].Designation = fmt.Sprintf("%s P%s", gd, romanNumeral(beltN))
		default: // BodyTerrestrial, BodyGasGiant
			planetN++
			dps[i].Designation = fmt.Sprintf("%s %s", gd, romanNumeral(planetN))
		}
	}
}

// AssignMoonDesignations walks each DetailedPlacement.Moons in arrival
// order (closest-to-farthest from the planet, per CountMoons/SizeMoon
// emission order) and assigns "<Planet> a", "<Planet> b", ... per
// WBH p.58 sidebar.
//
// Insignificant moons are out of scope for 2C.
//
// Mutates Moon.Designation in place.
func AssignMoonDesignations(dps []DetailedPlacement) {
	for i := range dps {
		for j := range dps[i].Moons {
			letter := byte('a' + j)
			dps[i].Moons[j].Designation = fmt.Sprintf("%s %c", dps[i].Designation, letter)
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
