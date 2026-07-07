package worlds_test

import (
	"errors"
	"testing"

	"github.com/philoserf/world-builder/stars"
	"github.com/philoserf/world-builder/worlds"
)

// TestSol_Generate is the Sol/Generate façade fixture per
// docs/harness.md § Façade end-to-end. Per
// docs/history/spike-findings.md § Finding 2, this is a Seeded shape-
// invariant fixture (not Scripted value-exact); pass-1's full-pipeline
// gold scripts were abandoned because pipeline reorders break them.
//
// Goes green when GenerateWithRoller's pipeline lands its last stage
// (cycle 11). Currently red — t.Skip removed at that point.
func TestSol_Generate(t *testing.T) {
	t.Parallel()

	good := 0
	for iter := range 100 {
		seed := int64(iter)
		u, err := worlds.Generate(seed)
		if err != nil {
			if errors.Is(err, stars.ErrSpecialCircumstances) {
				continue
			}
			t.Errorf("seed %d: unexpected error: %v", seed, err)
			continue
		}
		good++

		if u.System.Primary.Mass <= 0 {
			t.Errorf("seed %d: primary mass = %v, want > 0", seed, u.System.Primary.Mass)
		}
		if len(u.Placement.Allocations) == 0 {
			t.Errorf("seed %d: Allocations empty", seed)
		}
		if len(u.Placement.Placements) == 0 {
			t.Errorf("seed %d: Placements empty", seed)
		}
		if len(u.Detail.Class0I.Stars) == 0 {
			t.Errorf("seed %d: Class 0/I form has no Stars rows", seed)
		}
	}
	if good < 50 {
		t.Errorf("only %d / 100 seeds produced a system without post-stellar primaries (expected >= 50)", good)
	}
}

// TestZed_Generate is the Zed/Generate façade fixture per
// docs/harness.md § Façade end-to-end. Multi-star path; mirrors
// pass-1's TestZed_FullDetail_3A2b shape-invariant assertions.
//
// Note: Generate(seed) builds a fresh system from scratch (random
// stars). To exercise Zed specifically, callers compose the Zed system
// directly and feed it through the pipeline below the façade. This
// fixture exercises the façade with a free seed and asserts general
// invariants that hold for any system; per-stage Zed-specific
// fixtures live in their respective stage cycles.
//
// Goes green when GenerateWithRoller's pipeline lands its last stage.
// Currently red — t.Skip removed at that point.
func TestZed_Generate(t *testing.T) {
	t.Parallel()

	for iter := range 100 {
		seed := int64(iter)
		u, err := worlds.Generate(seed)
		if err != nil {
			// Skip post-stellar primaries (Special Circumstances).
			continue
		}

		// Every body in the iterator's range has its Kind set.
		for body := range u.AllBodies() {
			if body.Kind == worlds.BodyEmpty && body.Designation != "" {
				t.Errorf("seed %d: empty body has designation %q", seed, body.Designation)
			}
			if body.Kind != worlds.BodyEmpty && body.Designation == "" {
				t.Errorf("seed %d: non-empty body (%v) lacks designation", seed, body.Kind)
			}
		}

		// Per anti-pattern A.1 (moon-path silent-zero), every body
		// with non-empty Children has each child processed. The
		// iterator yields children alongside parents — so if any
		// child appears as a Body, it should be reachable from
		// AllBodies (no orphans).
		seen := map[string]bool{}
		for body := range u.AllBodies() {
			if body.Designation != "" {
				seen[body.Designation] = true
			}
		}
		for _, body := range u.Detail.Bodies {
			for _, child := range body.Children {
				if child.Designation != "" && !seen[child.Designation] {
					t.Errorf("seed %d: child %q not yielded by AllBodies (moon-path silent-zero?)",
						seed, child.Designation)
				}
			}
		}
	}
}

// TestGenerate_BeltsHaveDetails guards the ApplyBeltDetails gate: every
// planetoid belt in a generated system must carry Belt details (span,
// composition, resource rating). Regression for the historical
// SizeCode-"0" gate that matched no body, leaving Belt nil for every
// belt ever generated.
func TestGenerate_BeltsHaveDetails(t *testing.T) {
	t.Parallel()

	belts := 0
	for iter := range 100 {
		u, err := worlds.Generate(int64(iter))
		if err != nil {
			continue // post-stellar primaries etc. — covered elsewhere
		}
		for i := range u.Detail.Bodies {
			body := &u.Detail.Bodies[i]
			if body.Kind != worlds.BodyPlanetoidBelt {
				continue
			}
			belts++
			if body.Belt == nil {
				t.Errorf("seed %d: belt %s has nil Belt details", iter, body.Designation)
				continue
			}
			if body.Belt.Span <= 0 {
				t.Errorf("seed %d: belt %s span = %v, want > 0", iter, body.Designation, body.Belt.Span)
			}
			total := body.Belt.Composition.MTypePct + body.Belt.Composition.STypePct +
				body.Belt.Composition.CTypePct + body.Belt.Composition.OtherPct
			if total != 100 {
				t.Errorf("seed %d: belt %s composition sums to %d, want 100", iter, body.Designation, total)
			}
		}
	}
	if belts == 0 {
		t.Fatal("no belts across 100 seeds — sweep is vacuous")
	}
}

// TestGenerate_MoonOrbitsMaterialized guards the moon-refinement outputs
// that were historically computed but never assigned: OrbitKm must be
// positive for every refined moon (regression for the OrbitKm
// silent-zero), and the retrograde roll must be wired (over a seed
// sweep, at least one moon should come up retrograde — WBH p.78 gives
// outer-range moons a ~1-in-6+ chance).
func TestGenerate_MoonOrbitsMaterialized(t *testing.T) {
	t.Parallel()

	moons, retro := 0, 0
	for iter := range 100 {
		u, err := worlds.Generate(int64(iter))
		if err != nil {
			continue
		}
		for i := range u.Detail.Bodies {
			parent := &u.Detail.Bodies[i]
			for _, m := range parent.Children {
				moons++
				if m.OrbitPD > 0 && m.OrbitKm <= 0 {
					t.Errorf("seed %d: moon %s has OrbitPD %.2f but OrbitKm %v",
						iter, m.Designation, m.OrbitPD, m.OrbitKm)
				}
				if m.Retrograde {
					retro++
				}
			}
		}
	}
	if moons == 0 {
		t.Fatal("no moons across 100 seeds — sweep is vacuous")
	}
	if retro == 0 {
		t.Errorf("0 retrograde moons across %d moons — RollMoonRetrograde likely unwired", moons)
	}
}
