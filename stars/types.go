// Package stars implements the WBH Stars chapter (pp. 14–35).
package stars

import (
	"fmt"
)

// SpectralLetter is one of O, B, A, F, G, K, M (WBH p. 14).
type SpectralLetter rune

// SpectralType is a spectral letter plus a 0-9 subtype.
type SpectralType struct {
	Letter  SpectralLetter
	Subtype int
}

// String renders SpectralType as e.g. "G7".
func (s SpectralType) String() string {
	return fmt.Sprintf("%c%d", s.Letter, s.Subtype)
}

// ParseSpectralType parses a 2-character string like "G7" into a SpectralType.
func ParseSpectralType(s string) (SpectralType, error) {
	if len(s) != 2 {
		return SpectralType{}, fmt.Errorf("stars: spectral type must be 2 chars: %q", s)
	}
	letter := SpectralLetter(s[0])
	switch letter {
	case 'O', 'B', 'A', 'F', 'G', 'K', 'M':
		// ok
	default:
		return SpectralType{}, fmt.Errorf("stars: invalid spectral letter: %q", s)
	}
	d := s[1]
	if d < '0' || d > '9' {
		return SpectralType{}, fmt.Errorf("stars: subtype must be a digit: %q", s)
	}
	return SpectralType{Letter: letter, Subtype: int(d - '0')}, nil
}

// LuminosityClass is a stellar luminosity class (WBH p. 14).
type LuminosityClass string

// Luminosity class constants. D and BD are book conventions for white
// dwarf and brown dwarf, respectively (not formal Yerkes classes).
const (
	Ia  LuminosityClass = "Ia"
	Ib  LuminosityClass = "Ib"
	II  LuminosityClass = "II"
	III LuminosityClass = "III"
	IV  LuminosityClass = "IV"
	V   LuminosityClass = "V"
	VI  LuminosityClass = "VI"
	D   LuminosityClass = "D"
	BD  LuminosityClass = "BD"
)

// StarKind is the top-level kind of stellar/post-stellar object (WBH pp. 14–22).
type StarKind string

// StarKind constants.
const (
	KindMainSequence StarKind = "main_sequence"
	KindGiant        StarKind = "giant"
	KindSubgiant     StarKind = "subgiant"
	KindSupergiant   StarKind = "supergiant"
	KindSubdwarf     StarKind = "subdwarf"
	KindBrownDwarf   StarKind = "brown_dwarf"
	KindWhiteDwarf   StarKind = "white_dwarf"
	KindNeutronStar  StarKind = "neutron_star"
	KindBlackHole    StarKind = "black_hole"
	KindPulsar       StarKind = "pulsar"
	KindNebula       StarKind = "nebula"
	KindProtostar    StarKind = "protostar"
	KindStarCluster  StarKind = "star_cluster"
	KindAnomaly      StarKind = "anomaly"
)

// Star represents a single stellar (or stellar-remnant) object.
//
// Mass, Diameter, and Luminosity are in solar units.
// Temperature is in Kelvin. AgeGyr is in giga-years.
type Star struct {
	Kind            StarKind
	SpectralType    SpectralType
	LuminosityClass LuminosityClass
	Mass            float64
	Diameter        float64
	Temperature     float64
	Luminosity      float64
	AgeGyr          float64
}
