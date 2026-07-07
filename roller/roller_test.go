package roller

import (
	"testing"
)

func TestSeeded_Deterministic(t *testing.T) {
	a := NewSeeded(42)
	b := NewSeeded(42)
	for i := range 20 {
		ra, rb := a.Roll("2D"), b.Roll("2D")
		if ra != rb {
			t.Fatalf("seeded rollers diverged at i=%d: %d vs %d", i, ra, rb)
		}
	}
}

func TestSeeded_2DInRange(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("2D")
		if v < 2 || v > 12 {
			t.Fatalf("2D out of range: %d", v)
		}
	}
}

func TestSeeded_Modifier(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("2D-7")
		if v < -5 || v > 5 {
			t.Fatalf("2D-7 out of range: %d", v)
		}
	}
}

func TestSeeded_D10(t *testing.T) {
	r := NewSeeded(1)
	for range 200 {
		v := r.Roll("d10")
		if v < 1 || v > 10 {
			t.Fatalf("d10 out of range: %d", v)
		}
	}
}

func TestScripted_Order(t *testing.T) {
	r := NewScripted(7, 9, 11)
	if got := r.Roll("2D"); got != 7 {
		t.Fatalf("first roll = %d, want 7", got)
	}
	if got := r.Roll("2D"); got != 9 {
		t.Fatalf("second roll = %d, want 9", got)
	}
	if got := r.Roll("2D"); got != 11 {
		t.Fatalf("third roll = %d, want 11", got)
	}
}

func TestScripted_PanicsOnExhaustion(t *testing.T) {
	r := NewScripted(5)
	r.Roll("2D")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on exhausted Scripted")
		}
	}()
	r.Roll("2D")
}

func TestFixed_AlwaysSame(t *testing.T) {
	r := Fixed(8)
	if r.Roll("2D") != 8 {
		t.Fatal("Fixed(8).Roll(\"2D\") != 8")
	}
	if r.Roll("1D") != 8 {
		t.Fatal("Fixed(8).Roll(\"1D\") != 8")
	}
	if r.Roll("d100") != 8 {
		t.Fatal("Fixed(8).Roll(\"d100\") != 8")
	}
}

func TestRollerInterface(_ *testing.T) {
	var _ Roller = NewSeeded(1)
	var _ Roller = NewScripted(1)
	var _ Roller = Fixed(1)
}

// draw pulls n 2D results from r into a slice.
func draw(r Roller, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = r.Roll("2D")
	}
	return out
}

func eq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSeededFork_Reproducible — same parent seed and key yield the same
// child stream. This is what makes a (seed, body, family) substream
// stable across runs.
func TestSeededFork_Reproducible(t *testing.T) {
	a := NewSeeded(42).Fork("A II").Fork("climate")
	b := NewSeeded(42).Fork("A II").Fork("climate")
	if !eq(draw(a, 30), draw(b, 30)) {
		t.Fatal("same seed+key produced different child streams")
	}
}

// TestSeededFork_Disjoint — different keys yield independent streams.
// Different bodies (and different families of one body) must not share
// dice.
func TestSeededFork_Disjoint(t *testing.T) {
	root := NewSeeded(42)
	byBody := eq(draw(root.Fork("A II").Fork("climate"), 30),
		draw(root.Fork("B II").Fork("climate"), 30))
	if byBody {
		t.Fatal("different body keys produced identical streams")
	}
	byFamily := eq(draw(root.Fork("A II").Fork("climate"), 30),
		draw(root.Fork("A II").Fork("rotation-tilt"), 30))
	if byFamily {
		t.Fatal("different family keys produced identical streams")
	}
	// The cascade's re-eval family must differ from the first-pass family.
	byReeval := eq(draw(root.Fork("A II").Fork("climate"), 30),
		draw(root.Fork("A II").Fork("climate-reeval"), 30))
	if byReeval {
		t.Fatal("climate and climate-reeval produced identical streams")
	}
}

// TestSeededFork_PositionIndependent — the child stream is derived from
// the immutable construction seed, so a fork is identical whether taken
// before or after the parent has been rolled. This is the property that
// makes per-body output invariant to how much earlier stages consumed.
func TestSeededFork_PositionIndependent(t *testing.T) {
	early := NewSeeded(42)
	forkEarly := draw(early.Fork("A II").Fork("climate"), 30)

	late := NewSeeded(42)
	_ = draw(late, 137) // consume an arbitrary amount from the parent first
	forkLate := draw(late.Fork("A II").Fork("climate"), 30)

	if !eq(forkEarly, forkLate) {
		t.Fatal("fork stream depends on parent's roll position; must depend only on the immutable seed")
	}
}

// TestSeededFork_DoesNotConsumeParent — forking must not advance the
// parent's own stream, so a stage that forks takes zero draws from the
// shared stream.
func TestSeededFork_DoesNotConsumeParent(t *testing.T) {
	withFork := NewSeeded(42)
	_ = withFork.Fork("A II").Fork("climate")
	_ = withFork.Fork("B II").Fork("biology")
	afterForks := draw(withFork, 30)

	noFork := NewSeeded(42)
	baseline := draw(noFork, 30)

	if !eq(afterForks, baseline) {
		t.Fatal("Fork consumed draws from the parent stream")
	}
}

// TestScriptedFork_Transparent — a Scripted fork shares the one flat
// result sequence, consumed in call order. This is why worked-example
// fixtures that feed a narrated dice list are unaffected by per-body
// forking in the orchestrators.
func TestScriptedFork_Transparent(t *testing.T) {
	r := NewScripted(1, 2, 3, 4)
	a := r.Fork("x")
	b := r.Fork("y")
	// Interleaved draws across forks pull from the single shared sequence.
	if got := []int{a.Roll("2D"), b.Roll("2D"), a.Roll("2D"), b.Roll("2D")}; !eq(got, []int{1, 2, 3, 4}) {
		t.Fatalf("scripted forks did not share one sequence in call order: %v", got)
	}
}

// TestFixedFork_Constant — a pinned value stays pinned across the fork
// tree.
func TestFixedFork_Constant(t *testing.T) {
	if Fixed(8).Fork("anything").Roll("2D") != 8 {
		t.Fatal("Fixed.Fork changed the pinned value")
	}
}
