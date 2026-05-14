// Package worlds — Sol fidelity tests.
//
// These tests use real solar-system parameters as fixtures for the
// tidal-lock procedure. Where WBH's procedure successfully reproduces
// reality (Galilean moons, Earth's Moon, Phobos, Deimos all 1:1-locked),
// the tests pin that successful behavior. Where the procedure diverges
// from reality (Mercury 3:2, Pluto↔Charon mutual 1:1), the tests document
// the divergence as a finding rather than asserting an impossible outcome.
//
// Project-supplied fixtures (per docs/feedback_invent_text_for_presentation.md):
// the WBH book uses Zed as its threaded worked example; Sol's worked
// inputs are project supplied from real-world astronomical values.
//
// Sol body parameters used here:
//
//	Sun:     mass 1.0 M☉, age 4.6 Gyr.
//	Earth:   diameter 12,742 km → SizeCode "8"; mass 1.0 M⊕.
//	  Moon:    diameter 3,476 km → SizeCode "2"; orbit 384,400 km → OrbitPD 30.16;
//	           ecc 0.0549; period 655.7 h.
//	Mars:    diameter 6,779 km → SizeCode "4"; mass 0.107 M⊕.
//	  Phobos:  diameter 22 km → SizeCode "0"; orbit 9,377 km → OrbitPD 1.38;
//	           ecc 0.0151; period 7.66 h.
//	  Deimos:  diameter 12 km → SizeCode "0"; orbit 23,460 km → OrbitPD 3.46;
//	           ecc 0.00033; period 30.31 h.
//	Jupiter: diameter 139,820 km; mass 317.8 M⊕.
//	  Io:        diameter 3,643 km → SizeCode "2"; orbit 421,800 km → OrbitPD 3.02;
//	             ecc 0.0041; period 42.46 h.
//	  Europa:    diameter 3,122 km → SizeCode "1"; orbit 671,034 km → OrbitPD 4.80;
//	             ecc 0.0090; period 85.23 h.
//	  Ganymede:  diameter 5,268 km → SizeCode "3"; orbit 1,070,400 km → OrbitPD 7.66;
//	             ecc 0.0013; period 171.71 h.
//	  Callisto:  diameter 4,821 km → SizeCode "3"; orbit 1,882,700 km → OrbitPD 13.47;
//	             ecc 0.0074; period 400.54 h.
//
// SizeCode mapping per WBH p.69 sizing table:
//
//	"0" = 0–800 km      "5" = 8,001–9,600 km
//	"1" = 801–2,400 km  "6" = 9,601–11,200 km
//	"2" = 2,401–4,000 km "7" = 11,201–12,800 km
//	"3" = 4,001–5,600 km "8" = 12,801–14,400 km
//	"4" = 5,601–8,000 km
//
// (Sub-S = "S" = 600 km; bodies < 600 km that generate as "0" use nForSizeCode=0.)
package worlds

import (
	"math"
	"testing"

	"wbh/roller"
	"wbh/stars"
)

// TestSolFidelity_MoonToPlanet_Moon verifies that the WBH tidal-lock
// procedure, given real-world parameters for Earth's Moon, produces the
// canonical 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Moon SizeCode "2" (diameter 3,476 km); OrbitPD 30.16 (384,400 km / 12,742 km);
//	ecc 0.0549; period 655.7 h. Earth mass 1.0 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "2" → n=2 → ceil(2/3)=1 → +1
//	  ecc 0.0549 ≤ 0.1  → 0
//	  no AxialTilt pressure (Degrees=0 ≤ 3) → 0
//	  age 4.6 Gyr (not <1, not ≥5) → 0
//	  Common total: +1
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 30.16 > 20 → −floor(30.16/20) = −1
//	  parent mass 1.0 M⊕ (1 ≤ M < 10) → +2
//	  Specific total: +7
//	Grand total DM: +8
//
// Scripted dice:
//
//	2D for initial roll: 4 → adjusted = 4+8 = 12 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → no reroll; lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (655.7 h).
func TestSolFidelity_MoonToPlanet_Moon(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "2",
		OrbitPD:      30.16,
		Eccentricity: 0.0549,
	}
	parent := &Body{
		Kind:      BodyTerrestrial,
		MassEarth: 1.0,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "2",
		Eccentricity: 0.0549,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  655.7,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(4, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}

// TestSolFidelity_MoonToPlanet_Phobos verifies that the WBH tidal-lock
// procedure, given real-world parameters for Phobos, produces the
// canonical 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Phobos SizeCode "0" (diameter 22 km); OrbitPD 1.38 (9,377 km / 6,779 km);
//	ecc 0.0151; period 7.66 h. Mars mass 0.107 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "0" → n=0 → 0 (n < 1, no size DM)
//	  ecc 0.0151 ≤ 0.1  → 0
//	  no AxialTilt pressure → 0
//	  age 4.6 Gyr → 0
//	  Common total: 0
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 1.38 ≤ 20 → 0
//	  parent mass 0.107 M⊕ (< 1) → 0 (no bracket match)
//	  Specific total: +6
//	Grand total DM: +6
//
// Scripted dice:
//
//	2D for initial roll: 6 → adjusted = 6+6 = 12 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (7.66 h).
func TestSolFidelity_MoonToPlanet_Phobos(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "0",
		OrbitPD:      1.38,
		Eccentricity: 0.0151,
	}
	parent := &Body{
		Kind:      BodyTerrestrial,
		MassEarth: 0.107,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "0",
		Eccentricity: 0.0151,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  7.66,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(6, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}

// TestSolFidelity_MoonToPlanet_Deimos verifies that the WBH tidal-lock
// procedure, given real-world parameters for Deimos, produces the
// canonical 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Deimos SizeCode "0" (diameter 12 km); OrbitPD 3.46 (23,460 km / 6,779 km);
//	ecc 0.00033; period 30.31 h. Mars mass 0.107 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "0" → n=0 → 0
//	  ecc 0.00033 ≤ 0.1 → 0
//	  no AxialTilt pressure → 0
//	  age 4.6 Gyr → 0
//	  Common total: 0
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 3.46 ≤ 20 → 0
//	  parent mass 0.107 M⊕ (< 1) → 0
//	  Specific total: +6
//	Grand total DM: +6
//
// Scripted dice:
//
//	2D for initial roll: 6 → adjusted = 6+6 = 12 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (30.31 h).
func TestSolFidelity_MoonToPlanet_Deimos(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "0",
		OrbitPD:      3.46,
		Eccentricity: 0.00033,
	}
	parent := &Body{
		Kind:      BodyTerrestrial,
		MassEarth: 0.107,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "0",
		Eccentricity: 0.00033,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  30.31,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(6, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}

// TestSolFidelity_MoonToPlanet_Io verifies that the WBH tidal-lock
// procedure, given real-world parameters for Io, produces the canonical
// 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Io SizeCode "2" (diameter 3,643 km); OrbitPD 3.02 (421,800 km / 139,820 km);
//	ecc 0.0041; period 42.46 h. Jupiter mass 317.8 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "2" → n=2 → ceil(2/3)=1 → +1
//	  ecc 0.0041 ≤ 0.1 → 0
//	  no AxialTilt pressure → 0
//	  age 4.6 Gyr → 0
//	  Common total: +1
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 3.02 ≤ 20 → 0
//	  parent mass 317.8 M⊕ (100 ≤ M < 1000) → +6
//	  Specific total: +12
//	Grand total DM: +13
//
// Scripted dice:
//
//	2D for initial roll: 2 → adjusted = 2+13 = 15 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (42.46 h).
func TestSolFidelity_MoonToPlanet_Io(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "2",
		OrbitPD:      3.02,
		Eccentricity: 0.0041,
	}
	parent := &Body{
		Kind:      BodyGasGiant,
		MassEarth: 317.8,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "2",
		Eccentricity: 0.0041,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  42.46,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(2, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}

// TestSolFidelity_MoonToPlanet_Europa verifies that the WBH tidal-lock
// procedure, given real-world parameters for Europa, produces the canonical
// 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Europa SizeCode "1" (diameter 3,122 km); OrbitPD 4.80 (671,034 km / 139,820 km);
//	ecc 0.0090; period 85.23 h. Jupiter mass 317.8 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "1" → n=1 → ceil(1/3)=1 → +1
//	  ecc 0.0090 ≤ 0.1 → 0
//	  no AxialTilt pressure → 0
//	  age 4.6 Gyr → 0
//	  Common total: +1
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 4.80 ≤ 20 → 0
//	  parent mass 317.8 M⊕ (100 ≤ M < 1000) → +6
//	  Specific total: +12
//	Grand total DM: +13
//
// Scripted dice:
//
//	2D for initial roll: 2 → adjusted = 2+13 = 15 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (85.23 h).
func TestSolFidelity_MoonToPlanet_Europa(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "1",
		OrbitPD:      4.80,
		Eccentricity: 0.0090,
	}
	parent := &Body{
		Kind:      BodyGasGiant,
		MassEarth: 317.8,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "1",
		Eccentricity: 0.0090,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  85.23,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(2, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}

// TestSolFidelity_MoonToPlanet_Ganymede verifies that the WBH tidal-lock
// procedure, given real-world parameters for Ganymede, produces the
// canonical 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Ganymede SizeCode "3" (diameter 5,268 km); OrbitPD 7.66 (1,070,400 km / 139,820 km);
//	ecc 0.0013; period 171.71 h. Jupiter mass 317.8 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "3" → n=3 → ceil(3/3)=1 → +1
//	  ecc 0.0013 ≤ 0.1 → 0
//	  no AxialTilt pressure → 0
//	  age 4.6 Gyr → 0
//	  Common total: +1
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 7.66 ≤ 20 → 0
//	  parent mass 317.8 M⊕ (100 ≤ M < 1000) → +6
//	  Specific total: +12
//	Grand total DM: +13
//
// Scripted dice:
//
//	2D for initial roll: 2 → adjusted = 2+13 = 15 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (171.71 h).
func TestSolFidelity_MoonToPlanet_Ganymede(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "3",
		OrbitPD:      7.66,
		Eccentricity: 0.0013,
	}
	parent := &Body{
		Kind:      BodyGasGiant,
		MassEarth: 317.8,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "3",
		Eccentricity: 0.0013,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  171.71,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(2, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}

// TestSolFidelity_MoonToPlanet_Callisto verifies that the WBH tidal-lock
// procedure, given real-world parameters for Callisto, produces the
// canonical 1:1 lock observed in reality.
//
// Real-world values used as inputs:
//
//	Callisto SizeCode "3" (diameter 4,821 km); OrbitPD 13.47 (1,882,700 km / 139,820 km);
//	ecc 0.0074; period 400.54 h. Jupiter mass 317.8 M⊕. Sol age 4.6 Gyr.
//
// DM trace (commonTidalLockDMs + moonToPlanetDMs):
//
//	Common:
//	  SizeCode "3" → n=3 → ceil(3/3)=1 → +1
//	  ecc 0.0074 ≤ 0.1 → 0
//	  no AxialTilt pressure → 0
//	  age 4.6 Gyr → 0
//	  Common total: +1
//	MoonToPlanet specific:
//	  base → +6
//	  OrbitPD 13.47 ≤ 20 → 0
//	  parent mass 317.8 M⊕ (100 ≤ M < 1000) → +6
//	  Specific total: +12
//	Grand total DM: +13
//
// Scripted dice:
//
//	2D for initial roll: 2 → adjusted = 2+13 = 15 (≥ 12 → verification fires)
//	2D for verification: 7 (not natural 12 → lock holds as 1:1)
//
// Asserted outcome: 1:1 lock, day = period (400.54 h).
func TestSolFidelity_MoonToPlanet_Callisto(t *testing.T) {
	moonRef := &Body{
		Kind:         BodyMoon,
		SizeCode:     "3",
		OrbitPD:      13.47,
		Eccentricity: 0.0074,
	}
	parent := &Body{
		Kind:      BodyGasGiant,
		MassEarth: 317.8,
	}
	body := &Body{
		Kind:         BodyMoon,
		SizeCode:     "3",
		Eccentricity: 0.0074,
		AxialTilt:    &AxialTilt{Degrees: 0, BaselineDegrees: 0},
		DayLength:    &DayLength{SiderealHours: 24, BaselineSiderealHours: 24},
		PeriodHours:  400.54,
	}
	sys := stars.System{Primary: stars.Star{Mass: 1.0, AgeGyr: 4.6}}

	r := roller.NewScripted(2, 7)
	tl, err := GenerateTidalLock(r, body, moonRef, sys, parent, body.PeriodHours)
	if err != nil {
		t.Fatalf("GenerateTidalLock: %v", err)
	}
	if tl == nil {
		t.Fatal("expected non-nil TidalLock")
	}
	if tl.Case != TidalLockCaseMoonToPlanet {
		t.Errorf("Case = %v, want MoonToPlanet", tl.Case)
	}
	if tl.LockRatio != "1:1" {
		t.Errorf("LockRatio = %q, want 1:1", tl.LockRatio)
	}
	if math.Abs(body.DayLength.SiderealHours-body.PeriodHours) > 0.01 {
		t.Errorf("body.DayLength.SiderealHours = %v, want %v (PeriodHours)", body.DayLength.SiderealHours, body.PeriodHours)
	}
}
