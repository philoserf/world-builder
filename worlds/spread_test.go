package worlds

import (
	"math"
	"testing"

	"wbh/stars"
)

func TestSpread_BaseFormula(t *testing.T) {
	t.Parallel()
	primary := Group{MAO: 0.61, Intervals: []Interval{{Min: 0.61, Max: 20.0}}}
	// (3.1 - 0.61) / 5 = 0.498
	got := Spread(primary, 11, 3.1, 5, 3)
	if math.Abs(got-0.498) > 0.005 {
		t.Errorf("Spread = %v, want 0.498", got)
	}
}

func TestSpread_BaselineNLessThan1_TreatedAs1(t *testing.T) {
	t.Parallel()
	primary := Group{MAO: 0.5, Intervals: []Interval{{Min: 0.5, Max: 20.0}}}
	// baselineN = 0 → treat as 1. (10.0 - 0.5)/1 = 9.5.
	// primaryAllocated=2 keeps outermost (0.5 + 2×9.5 = 19.5) under the cap
	// so we isolate the "treat as 1" rule from the Maximum Spread cap.
	got := Spread(primary, 2, 10.0, 0, 1)
	if math.Abs(got-9.5) > 0.05 {
		t.Errorf("Spread = %v, want 9.5", got)
	}
}

func TestSpread_MaxCapApplied(t *testing.T) {
	t.Parallel()
	// Construct a case where base spread would push outermost > 20.
	// Primary intervals: [0.1, 20.0] → Total 19.9.
	// baselineN=1 → base spread = (19.0 - 0.1)/1 = 18.9.
	// Outermost from MAO at base spread for 5 allocated worlds: 0.1 + 5×18.9 = 94.6 > 20 → cap.
	// Cap = 19.9 / (5 + 3 stars) = 2.4875.
	primary := Group{MAO: 0.1, Intervals: []Interval{{Min: 0.1, Max: 20.0}}}
	got := Spread(primary, 5, 19.0, 1, 3)
	if math.Abs(got-2.4875) > 0.01 {
		t.Errorf("Spread = %v, want ~2.49 (cap applied)", got)
	}
}

func TestMaximumSecondarySpread(t *testing.T) {
	t.Parallel()
	sec := Group{MAO: 0.74, Intervals: []Interval{{Min: 0.74, Max: 7.10}}}
	// (7.10 - 0.74) / (5 + 1) = 1.06
	got := MaximumSecondarySpread(sec, 5)
	if math.Abs(got-1.06) > 0.01 {
		t.Errorf("MaximumSecondarySpread = %v, want 1.06", got)
	}
	_ = stars.V // silence unused-import if Compose isn't called
}
