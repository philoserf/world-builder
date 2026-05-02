package stars

import (
	"errors"
	"fmt"

	"wbh/roller"
)

// ComposeOpts holds the inputs to Compose.
type ComposeOpts struct {
	Kind            StarKind
	SpectralType    SpectralType
	LuminosityClass LuminosityClass
	Mass            float64
	Diameter        float64
	Temperature     float64
	AgeGyr          float64
}

// Compose constructs a Star from explicit physical values.
//
// Luminosity is derived from the WBH p. 20 formula:
//
//	L = (D/D⊙)^2 × (T/T⊙)^4
func Compose(o ComposeOpts) Star {
	return Star{
		Kind:            o.Kind,
		SpectralType:    o.SpectralType,
		LuminosityClass: o.LuminosityClass,
		Mass:            o.Mass,
		Diameter:        o.Diameter,
		Temperature:     o.Temperature,
		Luminosity:      ComputeLuminosityFromFormula(o.Diameter, o.Temperature),
		AgeGyr:          o.AgeGyr,
	}
}

// GenerateOpts controls GenerateMainSequenceStar.
type GenerateOpts struct {
	// WithVariance applies the optional ±20% variance roll for mass and
	// diameter (WBH p. 17). Off by default. Temperature variance is
	// intentionally omitted in Plan 1: WBH p. 17 leaves the scale of
	// temperature variance to the Referee without specifying a formula.
	WithVariance bool

	// Accuracy is 1 or 2. 1 uses 1D × 2 + D3 - 1 for age; 2 adds a d10
	// for an additional digit of accuracy.
	Accuracy int
}

// ErrNonClassVPrimary is returned by GenerateMainSequenceStar when the
// primary roll yields a non-Class-V star. Plan 1 only handles Class V;
// Plan 2 introduces full class dispatch.
var ErrNonClassVPrimary = errors.New("stars: non-Class-V primary; arrives in Plan 2")

// GenerateMainSequenceStar generates a Class V main-sequence primary
// star from rolls.
//
// Roll order (consumed from the roller in this order):
//  1. 2D for Star Type Determination (Type column)
//  2. 2D for Star Subtype
//  3. (if WithVariance) 2D-7 for mass variance
//  4. (if WithVariance) 2D-7 for diameter variance
//  5. 1D for age
//  6. D3 for age
//  7. (if Accuracy == 2) d10 for age
func GenerateMainSequenceStar(r roller.Roller, opts GenerateOpts) (Star, error) {
	letter, lc, err := RollPrimaryTypeAndClass(r)
	if err != nil {
		return Star{}, err
	}
	if lc != V {
		return Star{}, fmt.Errorf("%w: got %s", ErrNonClassVPrimary, lc)
	}
	subtype, err := RollSubtype(r, letter, lc)
	if err != nil {
		return Star{}, err
	}
	st := SpectralType{Letter: letter, Subtype: subtype}

	mass, err := ComputeMass(st, lc)
	if err != nil {
		return Star{}, err
	}
	diameter, err := ComputeDiameter(st, lc)
	if err != nil {
		return Star{}, err
	}
	temperature, err := ComputeTemperature(st)
	if err != nil {
		return Star{}, err
	}

	if opts.WithVariance {
		mass = ApplyVariance(mass, r, 0.20)
		diameter = ApplyVariance(diameter, r, 0.20)
	}

	luminosity := ComputeLuminosityFromFormula(diameter, temperature)

	age, err := SmallStarAge(r, opts.Accuracy)
	if err != nil {
		return Star{}, err
	}

	return Star{
		Kind:            KindMainSequence,
		SpectralType:    st,
		LuminosityClass: lc,
		Mass:            mass,
		Diameter:        diameter,
		Temperature:     temperature,
		Luminosity:      luminosity,
		AgeGyr:          age,
	}, nil
}
