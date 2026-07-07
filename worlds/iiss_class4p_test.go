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
	p := buildClass4PPlanet(u, body)
	if !p.Ring {
		t.Fatal("builder did not copy Body.Ring")
	}
	var b strings.Builder
	p.RenderBody(&b, iiss.FormHeader{})
	if !strings.Contains(b.String(), "planetary ring") {
		t.Errorf("Class IV-P body does not mention the ring:\n%s", b.String())
	}
}
