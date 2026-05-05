package worlds

import (
	"testing"
)

func TestComputeResidualSeismicStress_Terra(t *testing.T) {
	// Terra: Size 8, Age 4.568, density 1.0 (no density DM), 2 moons (Size 1+).
	// Per formula: 8 - 4.568 + 2 = 5.4322 → floor 5 → 5² = 25.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Moons = []Moon{{SizeCode: "1"}, {SizeCode: "1"}}
	got := ComputeResidualSeismicStress(body, 4.568, false)
	if got != 25 {
		t.Errorf("Terra: got %d, want 25", got)
	}
}

func TestComputeResidualSeismicStress_Luna(t *testing.T) {
	// Luna: Size 2, Age 4.568, density 0.6 (between 0.5 and 1.0 → no density DM),
	// IS a moon → +1.
	// Per formula: 2 - 4.568 + 1 = -1.5 → < 1 → 0.
	body := &DetailedPlacement{}
	body.SizeCode = "2"
	body.Physical = &BodyPhysical{Density: 0.6}
	got := ComputeResidualSeismicStress(body, 4.568, true)
	if got != 0 {
		t.Errorf("Luna: got %d, want 0 (-1.5 < 1 → 0)", got)
	}
}

func TestComputeResidualSeismicStress_ZedPrime(t *testing.T) {
	// Zed Prime: Size 5, Age 6.3, density 1.03 (> 1.0 → +2), IS a moon → +1.
	// Per formula AS WRITTEN: 5 - 6.3 + 1 + 2 = 1.7 → floor 1 → 1² = 1.
	//
	// NOTE: WBH p.126 worked example shows "5 - 6.3 +1 (for being a moon) +1
	// (for density) = 0.7 → 0", which only credits +1 for density 1.03 instead
	// of +2 per the formula table. We follow the formula as written in the
	// table, not the worked example. This is a book inconsistency — log a
	// feedback memory at end of branch.
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 1.03}
	got := ComputeResidualSeismicStress(body, 6.3, true)
	if got != 1 {
		t.Errorf("Zed Prime: got %d, want 1 (formula: 5 - 6.3 + 1 + 2 = 1.7 → 1 → 1)", got)
	}
}

func TestComputeResidualSeismicStress_PreSquareClampLessThanOne(t *testing.T) {
	// Inner expression goes below 1 — verifies < 1 → 0 clamp (NOT (-4)² = 16).
	body := &DetailedPlacement{}
	body.SizeCode = "1"
	body.Physical = &BodyPhysical{Density: 1.0}
	got := ComputeResidualSeismicStress(body, 5.0, false)
	if got != 0 {
		t.Errorf("got %d, want 0 (pre-square clamp)", got)
	}
}

func TestComputeResidualSeismicStress_DensityMaxMoonDM(t *testing.T) {
	// 15 Size-1+ moons → cap at +12.
	body := &DetailedPlacement{}
	body.SizeCode = "8"
	body.Physical = &BodyPhysical{Density: 1.0}
	body.Moons = make([]Moon, 15)
	for i := range body.Moons {
		body.Moons[i].SizeCode = "1"
	}
	got := ComputeResidualSeismicStress(body, 4.0, false)
	// 8 - 4.0 + 12 = 16 → 16² = 256
	if got != 256 {
		t.Errorf("got %d, want 256 (max +12 moon DM cap)", got)
	}
}

func TestComputeResidualSeismicStress_DensityLessThanHalf(t *testing.T) {
	body := &DetailedPlacement{}
	body.SizeCode = "5"
	body.Physical = &BodyPhysical{Density: 0.4}
	got := ComputeResidualSeismicStress(body, 1.0, false)
	// 5 - 1.0 - 1 = 3 → 3² = 9
	if got != 9 {
		t.Errorf("got %d, want 9 (density < 0.5 → -1)", got)
	}
}

func TestComputeResidualSeismicStress_NilBody_Zero(t *testing.T) {
	if got := ComputeResidualSeismicStress(nil, 4.5, false); got != 0 {
		t.Errorf("got %d, want 0 (nil body)", got)
	}
}

func TestComputeTidalStressFactor_ZedPrime(t *testing.T) {
	// Zed Prime: TidalEffects.Total ≈ 30m → 30/10 = 3.
	body := &DetailedPlacement{}
	body.TidalEffects = &SurfaceTidalEffects{Total: 30.0}
	got := ComputeTidalStressFactor(body)
	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestComputeTidalStressFactor_FloorRounding(t *testing.T) {
	body := &DetailedPlacement{}
	body.TidalEffects = &SurfaceTidalEffects{Total: 39.9}
	got := ComputeTidalStressFactor(body)
	// 39.9 / 10 = 3.99 → floor → 3
	if got != 3 {
		t.Errorf("got %d, want 3 (floor)", got)
	}
}

func TestComputeTidalStressFactor_NilTidalEffects_Zero(t *testing.T) {
	body := &DetailedPlacement{}
	body.TidalEffects = nil
	if got := ComputeTidalStressFactor(body); got != 0 {
		t.Errorf("got %d, want 0 (nil TidalEffects)", got)
	}
}

func TestComputeTidalStressFactor_LessThanTen_Zero(t *testing.T) {
	body := &DetailedPlacement{}
	body.TidalEffects = &SurfaceTidalEffects{Total: 9.5}
	// 9.5 / 10 = 0.95 → floor → 0
	if got := ComputeTidalStressFactor(body); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestComputeTidalStressFactor_NilBody_Zero(t *testing.T) {
	if got := ComputeTidalStressFactor(nil); got != 0 {
		t.Errorf("got %d, want 0 (nil body)", got)
	}
}

func TestComputeTidalStressFactor_HighTotal(t *testing.T) {
	// 1000m total → 100 (high TSS, near volcanic territory).
	body := &DetailedPlacement{}
	body.TidalEffects = &SurfaceTidalEffects{Total: 1000.0}
	if got := ComputeTidalStressFactor(body); got != 100 {
		t.Errorf("got %d, want 100", got)
	}
}
