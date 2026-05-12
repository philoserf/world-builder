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

// ErrCompanionOfGiantMAO indicates a companion of a giant primary
// (class Ia/Ib/II/III) requires a Plan 3+ MAO computation that is not
// yet implemented.
var ErrCompanionOfGiantMAO = fmt.Errorf(
	"stars: companion of giant primary requires MAO (Plan 3+): %w",
	ErrSpecialCircumstances,
)
