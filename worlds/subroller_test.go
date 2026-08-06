package worlds_test

import (
	"reflect"
	"testing"

	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/worlds"
)

// C1 sub-roller property tests (docs/rebuild-spec.md § C1,
// docs/c1-subroller-plan.md).
//
// Fidelity (the book reproduced to the digit) is proven by the existing
// worked-example tests, which drive orchestrators through Scripted
// rollers and pass unchanged because Scripted.Fork is transparent. These
// tests prove the other half: every per-body suffix stage is independent
// of the shared stream's position, so reordering a stage or re-rolling
// one body cannot perturb any other body.

// generateZedWithShift runs the full Zed pipeline, optionally injecting
// `shift` throwaway rolls at the structure-prefix / per-body-suffix
// boundary (right after ApplyDetailFrontEnd). The shift simulates an
// unrelated upstream change that happens to consume extra dice from the
// shared stream. Returns the per-body slice for comparison.
func generateZedWithShift(t *testing.T, seed int64, shift int) []worlds.Body {
	t.Helper()

	r := roller.NewSeeded(seed)
	sys := composeZed()

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("seed %d: placement: %v", seed, err)
	}

	u := &worlds.Universe{System: sys, Placement: sp}

	// Structure prefix — shared stream (creates body identities).
	if err := worlds.ApplyDetailFrontEnd(r, u); err != nil {
		t.Fatalf("seed %d: ApplyDetailFrontEnd: %v", seed, err)
	}

	// Unrelated perturbation at the boundary.
	for range shift {
		r.Roll("2D")
	}

	// Per-body suffix — every stage forks its own substream.
	suffix := []struct {
		name string
		fn   func(roller.Roller, *worlds.Universe) error
	}{
		{"ApplyBodyPhysical", worlds.ApplyBodyPhysical},
		{"ApplyBeltDetails", worlds.ApplyBeltDetails},
		{"ApplyMoonRefinement", worlds.ApplyMoonRefinement},
		{"ApplyRotationTilt", worlds.ApplyRotationTilt},
		{"ApplyClimate", worlds.ApplyClimate},
		{"ApplyTidalLockReEval", worlds.ApplyTidalLockReEval},
		{"ApplyTaintTypology", worlds.ApplyTaintTypology},
		{"ApplySurfaceDistribution", worlds.ApplySurfaceDistribution},
		{"ApplyGeology", worlds.ApplyGeology},
		{"ApplyBiology", worlds.ApplyBiology},
	}
	for _, s := range suffix {
		if err := s.fn(r, u); err != nil {
			t.Fatalf("seed %d: %s: %v", seed, s.name, err)
		}
	}

	worlds.ApplyHabitability(u)

	return u.Detail.Bodies
}

// TestC1_SuffixIsolatedFromStreamPosition is the whole-suffix isolation
// theorem: for a fixed seed, injecting any number of throwaway rolls at
// the prefix/suffix boundary leaves every per-body output byte-identical,
// because each (body, family) draws from a substream keyed off the
// immutable seed. Once every suffix stage is forked, this single
// assertion covers them all — and it is exactly what makes the
// tidal-lock cascade local and full-pipeline gold fixtures survivable.
func TestC1_SuffixIsolatedFromStreamPosition(t *testing.T) {
	t.Parallel()

	for seed := range int64(40) {
		baseline := generateZedWithShift(t, seed, 0)
		for _, shift := range []int{1, 3, 7} {
			shifted := generateZedWithShift(t, seed, shift)
			if !reflect.DeepEqual(baseline, shifted) {
				t.Fatalf("seed %d: per-body output changed under a %d-roll stream shift; "+
					"a suffix stage is still position-dependent", seed, shift)
			}
		}
	}
}

// TestC1_SuffixDependsOnSeed guards against the degenerate reading of the
// isolation test: if per-body output were constant, isolation would hold
// vacuously. Different seeds must produce different systems, so the
// isolation result is a real invariance, not a constant.
func TestC1_SuffixDependsOnSeed(t *testing.T) {
	t.Parallel()

	a := generateZedWithShift(t, 0, 0)

	b := generateZedWithShift(t, 1, 0)
	if reflect.DeepEqual(a, b) {
		t.Fatal("seeds 0 and 1 produced identical systems: output is not seed-dependent, " +
			"so the isolation test proves nothing")
	}
}

// TestC1_ForkKeysAreStable pins the fork-key scheme as contract. Renaming
// or reordering family keys or the bodyForkID format silently re-seeds
// every substream and changes all output; this catches that. The mainworld
// SAH for a fixed seed is the canary. If a deliberate scheme change lands,
// update the expected value in the same commit and regenerate the Markdown
// baseline.
func TestC1_ForkKeysAreStable(t *testing.T) {
	t.Parallel()

	u, err := worlds.Generate(42)
	if err != nil {
		t.Fatalf("Generate(42): %v", err)
	}

	if u.Detail.Mainworld == nil {
		t.Fatal("seed 42: no mainworld picked")
	}

	const wantSAH = "799" // seed 42 mainworld SAH under the current fork-key scheme
	if got := u.Detail.Mainworld.RenderSAH(); got != wantSAH {
		t.Fatalf("seed 42 mainworld SAH = %q, want %q: the fork-key scheme changed "+
			"(family key renamed/reordered or bodyForkID format changed). If intended, "+
			"update wantSAH and regenerate iiss/testdata.", got, wantSAH)
	}
}
