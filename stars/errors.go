package stars

import (
	"errors"
	"fmt"
)

// ErrSpecialCircumstances is the umbrella sentinel for errors that
// indicate the WBH Special Circumstances chapter (pp.147+) is
// required — out of project scope per the pass-2 scope statement.
// All such errors wrap this so callers can classify them via a single
// errors.Is(err, ErrSpecialCircumstances) check.
var ErrSpecialCircumstances = errors.New("stars: special circumstances chapter required")

// ErrCompanionOfGiantMAO is returned by RollStellarOrbit when asked for
// a companion orbit under a giant primary (class Ia/Ib/II/III): that
// case needs 1D × MAO(primary) (WBH p.27), but RollStellarOrbit has only
// the luminosity class, not the full Star that MAO requires. Callers
// compute MAO(primary) and use RollCompanionOrbitOfGiant instead —
// GenerateSystem does exactly this.
var ErrCompanionOfGiantMAO = fmt.Errorf(
	"stars: companion of giant primary requires MAO (Plan 3+): %w",
	ErrSpecialCircumstances,
)
