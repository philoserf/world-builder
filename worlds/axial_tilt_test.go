package worlds

import (
	"math"
	"testing"

	"wbh/roller"
)

func TestRollBasicAxialTilt_Rows2to9(t *testing.T) {
	cases := []struct {
		name    string
		scripts []int // dice values consumed in order
		want    float64
		delta   float64
	}{
		{
			name:    "2D=2 (1D-1)/50 with 1D=1",
			scripts: []int{2, 1},
			want:    0.0, // (1-1)/50 = 0
			delta:   0.001,
		},
		{
			name:    "2D=4 (1D-1)/50 with 1D=6",
			scripts: []int{4, 6},
			want:    0.1, // (6-1)/50 = 0.1
			delta:   0.001,
		},
		{
			name:    "2D=5 (1D)/5 with 1D=3",
			scripts: []int{5, 3},
			want:    0.6, // 3/5 = 0.6
			delta:   0.001,
		},
		{
			name:    "2D=6 1D with 1D=4",
			scripts: []int{6, 4},
			want:    4.0,
			delta:   0.001,
		},
		{
			name:    "2D=7 6+1D with 1D=3",
			scripts: []int{7, 3},
			want:    9.0, // 6+3=9
			delta:   0.001,
		},
		{
			name:    "2D=8 5+1D*5 with 1D=4",
			scripts: []int{8, 4},
			want:    25.0, // 5 + 4*5 = 25
			delta:   0.001,
		},
		{
			name:    "2D=9 5+1D*5 with 1D=5",
			scripts: []int{9, 5},
			want:    30.0, // 5 + 5*5 = 30
			delta:   0.001,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.scripts...)
			got, isExtreme, err := RollBasicAxialTilt(r)
			if err != nil {
				t.Fatal(err)
			}
			if isExtreme {
				t.Fatal("did not expect extreme dispatch for 2D < 10")
			}
			if math.Abs(got-c.want) > c.delta {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestRollBasicAxialTilt_DispatchesToExtreme(t *testing.T) {
	// 2D=10 → triggers Extreme dispatch.
	r := roller.NewScripted(10)
	tilt, isExtreme, err := RollBasicAxialTilt(r)
	if err != nil {
		t.Fatal(err)
	}
	if !isExtreme {
		t.Errorf("expected extreme dispatch for 2D=10")
	}
	if tilt != 0 {
		t.Errorf("expected sentinel 0 from RollBasicAxialTilt on extreme dispatch, got %v", tilt)
	}
}

func TestRollExtremeAxialTilt_Rows(t *testing.T) {
	cases := []struct {
		name    string
		scripts []int
		want    float64
		delta   float64
	}{
		// Row 1-2: 10 + 1D × 10. Row 1 with 1D×10 second roll = 1 → 10+10=20.
		{name: "row 1 with 1D=1", scripts: []int{1, 1}, want: 20, delta: 0.001},
		{name: "row 2 with 1D=6", scripts: []int{2, 6}, want: 70, delta: 0.001},
		// Row 3: 30 + 1D × 10. 1D=4 → 30+40=70 (Zed Prime path).
		{name: "row 3 with 1D=4 (Zed)", scripts: []int{3, 4}, want: 70, delta: 0.001},
		// Row 4: 90 + 1D × 1D (two 1Ds multiplied).
		{name: "row 4 with 1D=2, 1D=3", scripts: []int{4, 2, 3}, want: 96, delta: 0.001},
		// Row 5: 180 - 1D × 1D.
		{name: "row 5 with 1D=4, 1D=4", scripts: []int{5, 4, 4}, want: 164, delta: 0.001},
		// Row 6: 120 + 1D × 10.
		{name: "row 6 with 1D=4", scripts: []int{6, 4}, want: 160, delta: 0.001},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := roller.NewScripted(c.scripts...)
			got, err := RollExtremeAxialTilt(r)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-c.want) > c.delta {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestGenerateAxialTilt_ZedPrime(t *testing.T) {
	// Per WBH p.105: Zed Prime axial tilt 73.65°.
	//   2D=10 → Extreme dispatch.
	//   Extreme 1D=3 → row 3 → 30 + 1D × 10 with 1D=4 → 70°.
	//   Linear-variance precision:
	//     extra degrees: tens (1D-1)=0 (1D=1), ones d10=3 → +3° → 73.
	//     arcminutes: tens (1D-1)=3 (1D=4), ones d10=9 → +39'.
	//   Total: 73 + 39/60 = 73.65°.
	r := roller.NewScripted(
		10,   // basic 2D triggers Extreme
		3, 4, // extreme row 3, 1D=4 → 70°
		1, 3, // degree precision: 1D=1 (tens=0), d10=3 → +3°
		4, 9, // arcminute precision: 1D=4 (tens=3), d10=9 → +39'
	)
	dp := &Body{}
	dp.Kind = BodyTerrestrial
	dp.SizeCode = "5"

	at, err := GenerateAxialTilt(r, dp)
	if err != nil {
		t.Fatal(err)
	}
	if at == nil {
		t.Fatal("expected non-nil AxialTilt")
	}
	if math.Abs(at.Degrees-73.65) > 0.05 {
		t.Errorf("Degrees: got %v, want 73.65", at.Degrees)
	}
	if math.Abs(at.BaselineDegrees-73.65) > 0.05 {
		t.Errorf("BaselineDegrees: got %v, want 73.65", at.BaselineDegrees)
	}
	if at.Retrograde {
		t.Errorf("expected prograde (tilt < 90°), got Retrograde=true")
	}
}

func TestGenerateAxialTilt_RetrogradeAbove90(t *testing.T) {
	// Force a retrograde tilt: 2D=10 → Extreme; row 4 with high product → 90 + 5×6 = 120.
	// No linear-variance for this test (ones-degrees 0, arcminutes 0).
	r := roller.NewScripted(
		10,
		4, 5, 6, // extreme row 4, 1D=5, 1D=6 → 90 + 30 = 120°
		1, 0, // degrees: 1D=1, d10=0 → +0°
		1, 0, // arcminutes: 1D=1, d10=0 → +0'
	)
	dp := &Body{SizeCode: "5"}
	dp.Kind = BodyTerrestrial
	at, err := GenerateAxialTilt(r, dp)
	if err != nil {
		t.Fatal(err)
	}
	if !at.Retrograde {
		t.Errorf("expected retrograde for tilt 120°")
	}
	if math.Abs(at.Degrees-120) > 0.05 {
		t.Errorf("Degrees: got %v, want 120", at.Degrees)
	}
}

func TestGenerateAxialTilt_NilForEmptyBody(t *testing.T) {
	r := roller.NewScripted()
	dp := &Body{}
	at, err := GenerateAxialTilt(r, dp)
	if err != nil {
		t.Fatal(err)
	}
	if at != nil {
		t.Errorf("expected nil for empty body, got %+v", at)
	}
}
