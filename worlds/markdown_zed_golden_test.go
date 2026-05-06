package worlds_test

import (
	"flag"
	"os"
	"testing"

	"wbh/roller"
	"wbh/worlds"
)

var update = flag.Bool("update", false, "update golden files")

// buildZedSystemDetail constructs the WBH Zed quintuple system and runs
// it through the full 2B placement + DetailSystem pipeline using a
// fixed seed. This is the deterministic input for the Markdown
// rendering snapshot test.
//
// Data fidelity for Zed is covered by the TestZed_FullDetail_* family,
// which exercise the same pipeline under property-test invariants. This
// helper only needs reproducibility, so a seeded roller is sufficient.
func buildZedSystemDetail(t *testing.T) worlds.SystemDetail {
	t.Helper()

	sys := composeZed()
	r := roller.NewSeeded(0)

	sp, err := worlds.GenerateSystemPlacement(r, sys)
	if err != nil {
		t.Fatalf("GenerateSystemPlacement: %v", err)
	}

	header := worlds.IISSClass23Header{
		SectorLocation:  "Storr | 0602",
		InitialSurvey:   "207-568",
		LastUpdated:     "218-1061",
		IISSDesignation: "Zed (system)",
	}

	sd, err := worlds.DetailSystem(r, sys, sp, header)
	if err != nil {
		t.Fatalf("DetailSystem: %v", err)
	}
	return sd
}

// TestRenderSystemMarkdown_ZedGolden snapshots the full Markdown output
// for the Zed worked example. It catches accidental changes to layout,
// formatting, or section ordering across the three IISS forms.
//
// To refresh after an intentional format change:
//
//	go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
//
// Then inspect the diff with `git diff worlds/testdata/zed_markdown.golden`.
func TestRenderSystemMarkdown_ZedGolden(t *testing.T) {
	sd := buildZedSystemDetail(t)
	sys := composeZed()

	got := worlds.RenderSystemMarkdown(sd, sys)

	goldenPath := "testdata/zed_markdown.golden"
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v (run with -update to create)", err)
	}
	if got != string(want) {
		t.Errorf("golden mismatch; run `go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update` to refresh, then inspect with `git diff`")
	}
}
