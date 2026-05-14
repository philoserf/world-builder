package worlds

import (
	"errors"
	"strings"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/stars"
)

// TestGenerateSystemPlacement_WrappedErrorMessage asserts that errors
// returned by GenerateSystemPlacement carry a "worlds: <step>:" prefix
// identifying which step failed. We trigger an error by forcing
// AvailableOrbits to fail (post-stellar primary).
func TestGenerateSystemPlacement_WrappedErrorMessage(t *testing.T) {
	t.Parallel()

	primary := stars.Compose(stars.ComposeOpts{
		Kind: stars.KindBrownDwarf, // post-stellar primary triggers ErrPostStellarPrimaryUnsupported
	})
	sys := stars.System{Primary: primary}

	r := roller.NewSeeded(1)
	_, err := GenerateSystemPlacement(r, sys)
	if err == nil {
		t.Fatal("GenerateSystemPlacement returned nil error; expected post-stellar failure")
	}
	if !errors.Is(err, ErrPostStellarPrimaryUnsupported) {
		t.Errorf("err is not ErrPostStellarPrimaryUnsupported via Is(): %v", err)
	}
	if !strings.Contains(err.Error(), "worlds: available-orbits:") {
		t.Errorf("err.Error() = %q, want step prefix 'worlds: available-orbits:'", err.Error())
	}
}
