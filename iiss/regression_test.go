package iiss_test

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"wbh/iiss"
	"wbh/worlds"
)

// updateRegression rewrites the testdata/seed_*.md snapshots with the
// current pipeline output. Set with `go test -update.regression` when
// pass-2 intentionally changes output (cycle 17+18 landed convergence
// changes, for instance — running with -update.regression captures
// the new baseline).
var updateRegression = flag.Bool("update.regression", false,
	"rewrite iiss/testdata/seed_*.md regression snapshots from current pipeline")

// TestRegression_MarkdownSeeds is a within-pass-2 regression baseline.
// For a fixed set of seeds, the current Markdown output must match the
// snapshot in iiss/testdata/. Snapshots were captured at the cycle
// they were committed; future cycles either preserve them (no
// behaviour change) or are updated explicitly via -update.regression
// when the change is reviewed.
//
// Note: this is NOT a pass-1-vs-pass-2 comparison — pass-2's design
// explicitly diverges (TSS fold, surface-distribution-after-converge,
// climate fixed-point reorder). It IS a guard against unintentional
// pass-2-vs-pass-2 drift.
func TestRegression_MarkdownSeeds(t *testing.T) {
	t.Parallel()
	for _, seed := range []int64{1, 7, 42, 100, 500} {
		t.Run("seed_"+strconv.FormatInt(seed, 10), func(t *testing.T) {
			t.Parallel()
			u, err := worlds.Generate(seed)
			if err != nil {
				if isExpectedSpecial(err) {
					t.Skipf("seed %d hits Special-Circumstances chapter (out of pass-2 scope): %v", seed, err)
				}
				t.Fatalf("seed %d: Generate: %v", seed, err)
			}
			got := iiss.MarkdownSystem(u.Detail.SystemForms)

			snapshot := filepath.Join("testdata", "seed_"+strconv.FormatInt(seed, 10)+".md")

			if *updateRegression {
				if err := os.WriteFile(snapshot, []byte(got), 0o644); err != nil {
					t.Fatalf("seed %d: write %s: %v", seed, snapshot, err)
				}
				t.Logf("seed %d: wrote %d bytes to %s", seed, len(got), snapshot)
				return
			}

			want, err := os.ReadFile(snapshot)
			if err != nil {
				t.Fatalf("seed %d: read %s: %v (use -update.regression to seed)", seed, snapshot, err)
			}
			if got != string(want) {
				t.Errorf("seed %d: Markdown output drifted from %s\n"+
					"to update snapshot after reviewing the diff:\n"+
					"  go test ./iiss/... -update.regression -run TestRegression",
					seed, snapshot)
			}
		})
	}
}

func isExpectedSpecial(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "post-stellar primary") ||
		strings.Contains(msg, "special primary") ||
		strings.Contains(msg, "Special-primary") ||
		strings.Contains(msg, "giant primary requires MAO") ||
		strings.Contains(msg, "class IV missing")
}
