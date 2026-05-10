package worlds_test

import (
	"strings"
	"testing"

	"wbh/worlds"
)

// TestSol_Generate is the Sol/Generate façade fixture per
// docs/pass-2/harness.md § Façade end-to-end. Per
// docs/pass-2/spike-findings.md § Finding 2, this is a Seeded shape-
// invariant fixture (not Scripted value-exact); pass-1's full-pipeline
// gold scripts were abandoned because pipeline reorders break them.
//
// Goes green when GenerateWithRoller's pipeline lands its last stage
// (cycle 11). Currently red — t.Skip removed at that point.
func TestSol_Generate(t *testing.T) {
	t.Skip("cycle 11: GenerateWithRoller pipeline incomplete; see harness.md § Façade end-to-end")
	t.Parallel()

	for iter := range 100 {
		seed := int64(iter)
		u, err := worlds.Generate(seed)
		if err != nil {
			t.Fatalf("seed %d: Generate: %v", seed, err)
		}

		// Smoke-level shape — Stage 0 produces a primary star.
		if u.System.Primary.Mass <= 0 {
			t.Errorf("seed %d: primary mass = %v, want > 0", seed, u.System.Primary.Mass)
		}

		// Stage 1 produces at least one allocation and a non-empty
		// Placements slice.
		if len(u.Placement.Allocations) == 0 {
			t.Errorf("seed %d: Allocations empty", seed)
		}
		if len(u.Placement.Placements) == 0 {
			t.Errorf("seed %d: Placements empty", seed)
		}

		// Stage 10 produces an IISS Class 0/I form with at least one
		// star row.
		if len(u.Detail.Class0I.Stars) == 0 {
			t.Errorf("seed %d: Class 0/I form has no Stars rows", seed)
		}

		// Sol is single-primary; mainworld designation should be "A"
		// (when a habitable terrestrial is generated). Empty is also
		// accepted on seeds that produce no habitable bodies.
		if u.Detail.MainworldDesignation != "" && !strings.HasPrefix(u.Detail.MainworldDesignation, "A") {
			t.Errorf("seed %d: MainworldDesignation = %q, want \"A...\" or empty",
				seed, u.Detail.MainworldDesignation)
		}
	}
}

// TestZed_Generate is the Zed/Generate façade fixture per
// docs/pass-2/harness.md § Façade end-to-end. Multi-star path; mirrors
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
	t.Skip("cycle 11: GenerateWithRoller pipeline incomplete; see harness.md § Façade end-to-end")
	t.Parallel()

	for iter := range 100 {
		seed := int64(iter)
		u, err := worlds.Generate(seed)
		if err != nil {
			t.Fatalf("seed %d: Generate: %v", seed, err)
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
