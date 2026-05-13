package worlds

import (
	"testing"
)

func TestPreTidalLockSnapshot_RoundTrip(t *testing.T) {
	body := &Body{
		Eccentricity: 0.42,
		AxialTilt:    &AxialTilt{Degrees: 27, Retrograde: false},
		DayLength:    &DayLength{SiderealHours: 24, SolarHours: 24, YearDays: 365},
	}
	snap := CapturePreTidalLockSnapshot(body)
	// Mutate the body after capture.
	body.Eccentricity = 0.01
	body.AxialTilt.Degrees = 90
	body.AxialTilt.Retrograde = true
	body.DayLength.SiderealHours = 8766

	snap.RestoreInto(body)
	if body.Eccentricity != 0.42 {
		t.Errorf("Eccentricity not restored: %g, want 0.42", body.Eccentricity)
	}
	if body.AxialTilt == nil {
		t.Fatal("AxialTilt is nil after restore")
	}
	if body.AxialTilt.Degrees != 27 || body.AxialTilt.Retrograde {
		t.Errorf("AxialTilt not restored: %+v, want Degrees=27 Retrograde=false", body.AxialTilt)
	}
	if body.DayLength == nil {
		t.Fatal("DayLength is nil after restore")
	}
	if body.DayLength.SiderealHours != 24 {
		t.Errorf("DayLength.SiderealHours not restored: %v, want 24", body.DayLength.SiderealHours)
	}
}

func TestPreTidalLockSnapshot_NilFields(t *testing.T) {
	body := &Body{Eccentricity: 0.1} // nil DayLength, nil AxialTilt
	snap := CapturePreTidalLockSnapshot(body)
	snap.RestoreInto(body) // must not panic; body still has nil DayLength/AxialTilt
	if body.AxialTilt != nil {
		t.Errorf("AxialTilt = %v, want nil", body.AxialTilt)
	}
	if body.DayLength != nil {
		t.Errorf("DayLength = %v, want nil", body.DayLength)
	}
	if body.Eccentricity != 0.1 {
		t.Errorf("Eccentricity = %v, want 0.1", body.Eccentricity)
	}
}

func TestPreTidalLockSnapshot_IndependentOfBody(t *testing.T) {
	// Mutating snapshot fields after capture must NOT affect the body
	// (deep copy semantics for the embedded pointer fields).
	body := &Body{
		AxialTilt: &AxialTilt{Degrees: 10},
		DayLength: &DayLength{SiderealHours: 12},
	}
	snap := CapturePreTidalLockSnapshot(body)
	// Mutate the snapshot's AxialTilt and DayLength.
	if snap.AxialTilt != nil {
		snap.AxialTilt.Degrees = 99
	}
	if snap.DayLength != nil {
		snap.DayLength.SiderealHours = 99
	}
	// Body's originals must be untouched.
	if body.AxialTilt.Degrees != 10 {
		t.Errorf("body.AxialTilt.Degrees mutated to %v via snapshot", body.AxialTilt.Degrees)
	}
	if body.DayLength.SiderealHours != 12 {
		t.Errorf("body.DayLength.SiderealHours mutated to %v via snapshot", body.DayLength.SiderealHours)
	}
}
