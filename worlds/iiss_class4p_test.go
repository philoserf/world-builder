package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderIISSClass4P_NilBody_Empty(t *testing.T) {
	got := RenderIISSClass4P(nil, stars.System{}, "")
	if got != "" {
		t.Errorf("got %q, want empty (nil body)", got)
	}
}

func TestRenderIISSClass4P_BodyEmpty_Empty(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyEmpty
	got := RenderIISSClass4P(body, stars.System{}, "")
	if got != "" {
		t.Errorf("got %q, want empty (BodyEmpty)", got)
	}
}

func TestRenderIISSClass4P_Header_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV d"
	body.SizeCode = "5"
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "IISS CLASS IV SURVEY") {
		t.Errorf("missing header: got %q", got)
	}
	if !strings.Contains(got, "Aab IV d") {
		t.Errorf("missing designation: got %q", got)
	}
}

func TestRenderIISSClass4P_OrbitSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Eccentricity = 0.10
	body.Period = Period{Hours: 365.25 * 24}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "ORBIT") {
		t.Errorf("missing ORBIT section: got %q", got)
	}
}

func TestRenderIISSClass4P_SizeSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.DiameterKm = 8163
	body.MassEarth = 0.27
	body.Physical = &BodyPhysical{Density: 1.03, Gravity: 0.66}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "SIZE") {
		t.Errorf("missing SIZE section: got %q", got)
	}
	if !strings.Contains(got, "8163") {
		t.Errorf("missing diameter 8163: got %q", got)
	}
}

func TestRenderIISSClass4P_AtmosphereSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.042, ScaleHeight: 12.88, OxygenPartialPressure: 0.292}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "ATMOSPHERE") {
		t.Errorf("missing ATMOSPHERE section: got %q", got)
	}
}

func TestRenderIISSClass4P_HydrographicsSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Hydrographics = &Hydrographics{Code: 6, Percent: 62, Profile: "H6:H2O-100"}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "HYDROGRAPHICS") {
		t.Errorf("missing HYDROGRAPHICS section: got %q", got)
	}
	if !strings.Contains(got, "62") {
		t.Errorf("missing coverage 62: got %q", got)
	}
}

func TestRenderIISSClass4P_RotationSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.DayLength = &DayLength{SiderealHours: 24, SolarHours: 24, YearDays: 365}
	body.AxialTilt = &AxialTilt{Degrees: 23.5}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "ROTATION") {
		t.Errorf("missing ROTATION section: got %q", got)
	}
}

func TestRenderIISSClass4P_TemperatureSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Temperature = &Temperature{
		MeanK: 300, HighK: 346, LowK: 262,
		Luminosity: 1.419, Albedo: 0.33, GreenhouseFactor: 0.59,
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "TEMPERATURE") {
		t.Errorf("missing TEMPERATURE section: got %q", got)
	}
	if !strings.Contains(got, "300") {
		t.Errorf("missing MeanK 300: got %q", got)
	}
}

func TestRenderIISSClass4P_SeismicSection_Present(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab III"
	body.SizeCode = "5"
	body.Geology = &Geology{
		ResidualSeismicStress: 0,
		TidalStressFactor:     3,
		TidalHeatingFactor:    14,
		TotalSeismicStress:    17,
		TectonicPlates:        4,
	}
	got := RenderIISSClass4P(body, stars.System{}, "")
	if !strings.Contains(got, "SEISMIC") {
		t.Errorf("missing SEISMIC section: got %q", got)
	}
	if !strings.Contains(got, "17") {
		t.Errorf("missing TSS 17: got %q", got)
	}
}
