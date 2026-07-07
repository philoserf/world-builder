package worlds

import (
	"strings"
	"testing"

	"github.com/philoserf/world-builder/iiss"
)

// TestClass4P_RingRendered asserts a ringed mainworld surfaces its ring
// on the Class IV-P form (previously only the Class II/III object-notes
// column showed it).
func TestClass4P_RingRendered(t *testing.T) {
	t.Parallel()
	u := &Universe{}
	body := &Body{Designation: "A I", Kind: BodyTerrestrial, SizeCode: "7", Ring: true}
	p := buildClass4PPlanet(u, body, true)
	if !p.Ring {
		t.Fatal("builder did not copy Body.Ring")
	}
	var b strings.Builder
	p.RenderBody(&b, iiss.FormHeader{})
	if !strings.Contains(b.String(), "planetary ring") {
		t.Errorf("Class IV-P body does not mention the ring:\n%s", b.String())
	}
}

// TestClass4P_GasGiant asserts a gas-giant PART P renders GG-appropriate
// detail (class + residual temperature) and omits the misleading
// terrestrial atmosphere/hydrographics sections.
func TestClass4P_GasGiant(t *testing.T) {
	t.Parallel()
	u := &Universe{}
	gg := &Body{
		Designation:   "A II",
		Kind:          BodyGasGiant,
		GGClass:       GasGiantMedium,
		DiameterEarth: 7.1,
		MassEarth:     280,
		Geology:       &Geology{InherentTemperatureK: 187},
	}
	p := buildClass4PPlanet(u, gg, false)
	if !p.IsGasGiant || p.GasGiantClass != "Medium" {
		t.Fatalf("GG flags wrong: IsGasGiant=%v class=%q", p.IsGasGiant, p.GasGiantClass)
	}
	var b strings.Builder
	p.RenderBody(&b, iiss.FormHeader{})
	out := b.String()
	if !strings.Contains(out, "Medium gas giant") || !strings.Contains(out, "### GAS GIANT") {
		t.Errorf("GG PART P missing gas-giant detail:\n%s", out)
	}
	if strings.Contains(out, "vacuum") || strings.Contains(out, "### HYDROGRAPHICS") {
		t.Errorf("GG PART P leaked terrestrial sections:\n%s", out)
	}
}
