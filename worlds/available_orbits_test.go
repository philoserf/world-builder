package worlds

import (
	"math"
	"testing"
)

func TestGroup_Total_SingleInterval(t *testing.T) {
	t.Parallel()

	g := Group{
		Designation: "A",
		MAO:         0.03,
		Intervals:   []Interval{{Min: 0.03, Max: 20.0}},
	}
	got := g.Total()
	want := 19.97
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Total = %v, want %v", got, want)
	}
}

func TestGroup_Total_MultiInterval(t *testing.T) {
	t.Parallel()

	// Zed Aab from WBH p. 40: 0.61–5.10, 7.10–10.10, 14.10–20.00 → 13.39.
	g := Group{
		Designation: "Aab",
		MAO:         0.61,
		Intervals: []Interval{
			{Min: 0.61, Max: 5.10},
			{Min: 7.10, Max: 10.10},
			{Min: 14.10, Max: 20.00},
		},
	}
	got := g.Total()
	want := 13.39
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Total = %v, want %v", got, want)
	}
}

func TestGroup_Total_Empty(t *testing.T) {
	t.Parallel()

	g := Group{Intervals: nil}
	if got := g.Total(); got != 0 {
		t.Errorf("Total = %v, want 0", got)
	}
}

func TestGroup_Contains(t *testing.T) {
	t.Parallel()

	g := Group{
		Intervals: []Interval{
			{Min: 0.61, Max: 5.10},
			{Min: 7.10, Max: 10.10},
		},
	}

	tests := []struct {
		orbit float64
		want  bool
	}{
		{0.5, false},
		{0.61, true},
		{3.0, true},
		{5.10, true},
		{6.0, false},
		{7.0, false},
		{7.10, true},
		{10.10, true},
		{15.0, false},
	}
	for _, tc := range tests {
		if got := g.Contains(tc.orbit); got != tc.want {
			t.Errorf("Contains(%v) = %v, want %v", tc.orbit, got, tc.want)
		}
	}
}
