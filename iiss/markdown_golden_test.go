package iiss_test

import (
	"testing"

	"wbh/worlds"

	"wbh/iiss"
)

// TestZed_MarkdownGolden asserts MarkdownSystem produces a stable,
// canonical output for a fixed seed. Per docs/pass-2/harness.md §
// Markdown system output, updates require explicit acknowledgement;
// the test is the canonical Markdown rendering.
//
// Goes green when iiss.MarkdownSystem and the underlying pipeline
// complete (cycle 11). Currently red — t.Skip removed at that point.
//
// Golden file path TBD: iiss/testdata/zed_markdown_golden.md. The
// helper writes the actual output to .actual on diff so the developer
// can inspect before promoting.
func TestZed_MarkdownGolden(t *testing.T) {
	t.Skip("cycle 11: pipeline + renderers incomplete; see harness.md § Markdown system output")
	t.Parallel()

	u, err := worlds.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got := iiss.MarkdownSystem(u.Detail.SystemForms)

	// TODO cycle 11: load testdata/zed_markdown_golden.md and compare;
	// emit .actual on mismatch.
	if len(got) == 0 {
		t.Fatal("MarkdownSystem produced empty output")
	}
}
