package stars

import (
	"errors"
	"fmt"

	"wbh/roller"
)

// ErrSpecialPrimary is returned by RollPrimaryTypeAndClass when the
// initial Type-column roll is 2 ("Special"). Callers should dispatch
// through peculiar.go to resolve the special-object branch.
var ErrSpecialPrimary = errors.New("stars: special primary; dispatch through peculiar")

var validLetters = map[SpectralLetter]struct{}{
	'O': {}, 'B': {}, 'A': {}, 'F': {}, 'G': {}, 'K': {}, 'M': {},
}

// RollPrimaryTypeAndClass rolls for the spectral letter and luminosity
// class of a primary star (WBH p. 15).
//
// A roll of 12 redirects to the Hot column with a fresh 2D roll.
// A roll of 2 returns ErrSpecialPrimary so the caller can route through
// the Special / Unusual / Peculiar dispatch in peculiar.go.
func RollPrimaryTypeAndClass(r roller.Roller) (SpectralLetter, LuminosityClass, error) {
	first := r.Roll("2D")
	row, ok := StarTypeDetermination[first]
	if !ok {
		return 0, "", fmt.Errorf("stars: 2D out of range: %d", first)
	}
	cell := row.Type

	if cell == "Hot" {
		second := r.Roll("2D")
		hotRow, ok := StarTypeDetermination[second]
		if !ok {
			return 0, "", fmt.Errorf("stars: 2D out of range: %d", second)
		}
		cell = hotRow.Hot
	}

	if cell == "Special" {
		return 0, "", ErrSpecialPrimary
	}

	if len(cell) != 1 {
		return 0, "", fmt.Errorf("stars: unexpected primary type cell: %q", cell)
	}
	letter := SpectralLetter(cell[0])
	if _, ok := validLetters[letter]; !ok {
		return 0, "", fmt.Errorf("stars: invalid letter from table: %c", letter)
	}
	return letter, V, nil
}

// RollSubtype rolls Star Subtype (WBH p. 16) and returns 0-9.
//
// Primary M-type stars use the M column; all others use the Numeric
// column. K-type Class IV stars have any subtype > 4 reduced by 5.
func RollSubtype(r roller.Roller, letter SpectralLetter, lc LuminosityClass) (int, error) {
	if _, ok := validLetters[letter]; !ok {
		return 0, fmt.Errorf("stars: invalid letter: %c", letter)
	}
	roll := r.Roll("2D")
	var sub int
	if letter == 'M' {
		s, ok := StarSubtypeMType[roll]
		if !ok {
			return 0, fmt.Errorf("stars: 2D out of range: %d", roll)
		}
		sub = s
	} else {
		s, ok := StarSubtypeNumeric[roll]
		if !ok {
			return 0, fmt.Errorf("stars: 2D out of range: %d", roll)
		}
		sub = s
	}
	if letter == 'K' && lc == IV && sub > 4 {
		sub -= 5
	}
	return sub, nil
}
