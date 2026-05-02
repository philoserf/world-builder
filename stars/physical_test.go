package stars

import (
	"math"
	"testing"
)

func TestInterpolateClassRow_GridPoint(t *testing.T) {
	// G5 V is a tabulated grid point; interpolation should return the table value.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'G', Subtype: 5}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != 0.9 {
		t.Fatalf("got %v want 0.9", got)
	}
}

func TestInterpolateClassRow_G7VMass(t *testing.T) {
	// WBH p. 17: G7 V mass = 0.9 + (2/5) * (0.8 - 0.9) = 0.86.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'G', Subtype: 7}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.86
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_G7VDiameter(t *testing.T) {
	// WBH p. 18: G7 V diameter = 0.95 + (2/5) * (0.9 - 0.95) = 0.93.
	got, err := InterpolateClassRow(StarDiameter, SpectralType{Letter: 'G', Subtype: 7}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.93
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateScalar_Temperature(t *testing.T) {
	// WBH p. 17: G7 V temperature = 5600 + 0.4 * (5200 - 5600) = 5440.
	got, err := InterpolateScalar(StarTemperature, SpectralType{Letter: 'G', Subtype: 7})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 5440.0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_A3III(t *testing.T) {
	// A3 III interpolates between A0 III (8) and A5 III (6) at 3/5.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'A', Subtype: 3}, III)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 8 + (3.0/5.0)*(6-8)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_M3V(t *testing.T) {
	// M3 V interpolates between M0 V (0.5) and M5 V (0.16) at 3/5.
	got, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'M', Subtype: 3}, V)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := 0.5 + (3.0/5.0)*(0.16-0.5)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInterpolateClassRow_MissingClass(t *testing.T) {
	// Class IV not tabulated for O0.
	_, err := InterpolateClassRow(StarMass, SpectralType{Letter: 'O', Subtype: 0}, IV)
	if err == nil {
		t.Fatal("expected error for missing class")
	}
}
