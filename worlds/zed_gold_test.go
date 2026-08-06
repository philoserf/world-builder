package worlds_test

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/philoserf/world-builder/iiss"
	"github.com/philoserf/world-builder/roller"
	"github.com/philoserf/world-builder/worlds"
)

// updateZedGold rewrites the Zed full-pipeline gold-master. Set with
// `go test ./worlds/ -update.zedgold` after reviewing an intentional
// output change.
var updateZedGold = flag.Bool("update.zedgold", false,
	"rewrite worlds/testdata/zed_gold.md from the current pipeline")

// zedGoldSeed pins the placement stream for the gold-master. Seed 27's
// system has a moon mainworld (Aab II d), so the fixture exercises the
// Class IV-P moon-mainworld path — the same shape as the book's Zed
// Prime — plus climate, biology, geology, and moons throughout.
const zedGoldSeed = 27

// buildZedUniverse runs the canonical Zed stars (composeZed) through the
// entire pipeline at zedGoldSeed. When swapTaintSurface is set, the two
// mutually-independent post-climate stages (taint typology and surface
// distribution) run in swapped order — used to prove reorder-survival.
func buildZedUniverse(t *testing.T, swapTaintSurface bool) *worlds.Universe {
	t.Helper()

	r := roller.NewSeeded(zedGoldSeed)
	sys := composeZed()

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("placement: %v", err)
	}

	u := &worlds.Universe{System: sys, Placement: sp}

	if err := worlds.ApplyDetailFrontEnd(r, u); err != nil {
		t.Fatalf("ApplyDetailFrontEnd: %v", err)
	}

	stages := []func(roller.Roller, *worlds.Universe) error{
		worlds.ApplyBodyPhysical,
		worlds.ApplyBeltDetails,
		worlds.ApplyMoonRefinement,
		worlds.ApplyRotationTilt,
		worlds.ApplyClimate,
		worlds.ApplyTidalLockReEval,
	}
	if swapTaintSurface {
		stages = append(stages, worlds.ApplySurfaceDistribution, worlds.ApplyTaintTypology)
	} else {
		stages = append(stages, worlds.ApplyTaintTypology, worlds.ApplySurfaceDistribution)
	}

	stages = append(stages, worlds.ApplyGeology, worlds.ApplyBiology)

	for _, fn := range stages {
		if err := fn(r, u); err != nil {
			t.Fatalf("suffix stage: %v", err)
		}
	}

	worlds.ApplyHabitability(u)
	worlds.AggregateSystem(u)
	worlds.BuildIISSForms(u)

	return u
}

// TestZed_GoldMaster is the full-pipeline gold-master for the named Zed
// system: composeZed's canonical stars run end-to-end at a pinned seed,
// rendered to the complete Markdown (all three IISS forms), and diffed
// against a committed golden file. Any pipeline change that alters a
// single field of the output fails this test with a reviewable diff.
//
// Pass 1 and pass 2 could not maintain a full-pipeline fixture like this:
// on a single shared dice stream, any stage reorder or added roll shifted
// every downstream value, so such a fixture asserted pipeline ordering,
// not system content (history/spike-findings.md § Finding 2). C1's
// per-body sub-streams make each body's output a function of (seed,
// identity, family) only — see TestZed_GoldSurvivesStageReorder — so this
// gold-master is stable against unrelated pipeline changes and is a
// trustworthy regression anchor.
func TestZed_GoldMaster(t *testing.T) {
	t.Parallel()

	u := buildZedUniverse(t, false)
	got := iiss.MarkdownClass4Survey(u.Detail.SystemForms)

	golden := filepath.Join("testdata", "zed_gold.md")
	if *updateZedGold {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("write %s: %v", golden, err)
		}

		t.Logf("wrote %d bytes to %s", len(got), golden)

		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read %s: %v (use -update.zedgold to seed)", golden, err)
	}

	if got != string(want) {
		t.Errorf("Zed gold-master drifted from %s\n"+
			"review the diff, then if intended:\n"+
			"  go test ./worlds/ -update.zedgold -run TestZed_GoldMaster", golden)
	}
}

// TestZed_GoldSurvivesStageReorder is the concrete demonstration of C1's
// dividend: the entire system is byte-identical whether the two
// mutually-independent post-climate stages (taint typology, surface
// distribution) run in canonical or swapped order. Under the old shared
// stream the swap would shift every downstream draw and change the output;
// under C1 each stage draws from its own (body, family) sub-stream, so
// order does not matter. This is what makes the gold-master above durable.
func TestZed_GoldSurvivesStageReorder(t *testing.T) {
	t.Parallel()

	canonical := buildZedUniverse(t, false)
	swapped := buildZedUniverse(t, true)

	if !reflect.DeepEqual(canonical.Detail.Bodies, swapped.Detail.Bodies) {
		t.Fatal("swapping taint/surface changed per-body output: a suffix stage is still position-dependent")
	}

	if iiss.MarkdownClass4Survey(canonical.Detail.SystemForms) != iiss.MarkdownClass4Survey(swapped.Detail.SystemForms) {
		t.Fatal("swapping taint/surface changed the rendered system")
	}
}
