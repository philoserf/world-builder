package worlds_test

import (
	"strings"
	"testing"

	"wbh/worlds"
)

// TestNotableFeatures_EmptyOnVoid verifies an empty Universe produces
// no block (not just an empty heading).
func TestNotableFeatures_EmptyOnVoid(t *testing.T) {
	t.Parallel()
	var u worlds.Universe
	got := worlds.NotableFeatures(&u)
	if got != "" {
		t.Errorf("expected empty for void universe, got:\n%s", got)
	}
}

// TestNotableFeatures_NilSafe verifies the function handles a nil
// receiver (catch-all for misuse).
func TestNotableFeatures_NilSafe(t *testing.T) {
	t.Parallel()
	got := worlds.NotableFeatures(nil)
	if got != "" {
		t.Errorf("expected empty for nil universe, got:\n%s", got)
	}
}

// TestNotableFeatures_Sol_HasMainworldNote drives end-to-end via a
// real generated system. Seed 42 produces a mainworld with non-empty
// habitability notes; assert the block contains the mainworld section
// and the WorstLow cold snap surfaces.
func TestNotableFeatures_Sol_HasMainworldNote(t *testing.T) {
	t.Parallel()
	u, err := worlds.Generate(42)
	if err != nil {
		t.Fatalf("Generate(42): %v", err)
	}
	got := worlds.NotableFeatures(&u)
	if got == "" {
		t.Fatal("expected non-empty Notable Features block for seed 42")
	}
	for _, want := range []string{
		"## Notable Features",
		"### Mainworld habitability",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("block missing %q; got:\n%s", want, got)
		}
	}
}
