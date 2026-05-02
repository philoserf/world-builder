// Package dice parses WBH dice notation strings.
//
// Supports the conventions used in the World Builder's Handbook:
//   - "2D" / "1D"        -> count of d6
//   - "D3", "d10", "d100", "2d10" -> arbitrary-sided dice
//   - "+N" / "-N"        -> integer modifier appended to any of the above
package dice

import (
	"fmt"
	"regexp"
	"strconv"
)

// Spec is the parsed form of a dice notation string.
type Spec struct {
	Count, Sides, Modifier int
}

var pattern = regexp.MustCompile(`^(\d*)[Dd](\d*)(?:([+-])(\d+))?$`)

// Parse converts a WBH dice notation string into a Spec.
//
// Examples:
//
//	Parse("2D")    -> Spec{2, 6, 0}
//	Parse("2D-7")  -> Spec{2, 6, -7}
//	Parse("D3-1")  -> Spec{1, 3, -1}
func Parse(notation string) (Spec, error) {
	m := pattern.FindStringSubmatch(notation)
	if m == nil {
		return Spec{}, fmt.Errorf("dice: invalid notation: %q", notation)
	}
	count := 1
	if m[1] != "" {
		c, err := strconv.Atoi(m[1])
		if err != nil {
			return Spec{}, fmt.Errorf("dice: bad count in %q: %w", notation, err)
		}
		count = c
	}
	sides := 6
	if m[2] != "" {
		s, err := strconv.Atoi(m[2])
		if err != nil {
			return Spec{}, fmt.Errorf("dice: bad sides in %q: %w", notation, err)
		}
		sides = s
	}
	if sides < 2 {
		return Spec{}, fmt.Errorf("dice: invalid sides (<2) in %q", notation)
	}
	modifier := 0
	if m[3] != "" {
		v, err := strconv.Atoi(m[4])
		if err != nil {
			return Spec{}, fmt.Errorf("dice: bad modifier in %q: %w", notation, err)
		}
		if m[3] == "-" {
			v = -v
		}
		modifier = v
	}
	return Spec{Count: count, Sides: sides, Modifier: modifier}, nil
}
