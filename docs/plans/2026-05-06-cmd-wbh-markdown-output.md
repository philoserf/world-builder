# cmd/wbh Markdown output Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `cmd/wbh` emit a complete description of a generated star system as Markdown — all three IISS forms, in book order, under H1/H2 headings — and make it the new default `-format`.

**Architecture:** Three Markdown formatters sit beside the existing form renderers. Class 0/I and Class II/III formatters consume the existing `stars.SurveyForm` and `worlds.IISSClass23Form` structs (no duplication of `BuildSurveyForm`/`RenderIISSClass23` logic). Class IV-P reads from `*DetailedPlacement` directly (no struct exists there). A top-level `RenderSystemMarkdown` orchestrator chains the three under H2 sections. `cmd/wbh` wires the new path and flips the default `-format`.

**Tech Stack:** Go 1.26, `strings.Builder` for rendering, `go test -race ./...` via `task test`, `gofumpt` formatting.

**Spec:** `docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md`

---

## File Map

| File                                  | Action | What it holds                                                                                                                                               |
| ------------------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `stars/markdown.go`                   | create | `RenderClass0IMarkdown(SurveyForm) string` — Form 0421B-0I                                                                                                  |
| `stars/markdown_test.go`              | create | Per-form unit tests on synthetic `SurveyForm`                                                                                                               |
| `worlds/markdown.go`                  | create | `RenderClass4PMarkdown(*DetailedPlacement, stars.System, mainworldDesignation) string` (PART P + PART P.B), `RenderClass23Markdown(IISSClass23Form) string` |
| `worlds/markdown_test.go`             | create | Per-form unit tests for both formatters above                                                                                                               |
| `worlds/markdown_system.go`           | create | Top-level `RenderSystemMarkdown(SystemDetail, stars.System) string`                                                                                         |
| `worlds/markdown_system_test.go`      | create | Orchestrator test + golden-file Zed snapshot test                                                                                                           |
| `worlds/testdata/zed_markdown.golden` | create | Golden Markdown for the Zed worked-example seed                                                                                                             |
| `cmd/wbh/main.go`                     | modify | Flip default `-format` to `"markdown"`; new `case "markdown":` branch                                                                                       |
| `cmd/wbh/main_test.go`                | modify | Tests for new default + each `-format` value                                                                                                                |

---

## Conventions used by every Markdown formatter

These conventions are referenced by tasks below. Apply them everywhere.

- **Em-dash `—` for missing values.** Empty strings, zero `int`/`float` where the book uses "—" for "not applicable" (composite-row Temperature/Diameter, Stars rows where Orbit doesn't apply, etc.). Use the literal Unicode em-dash, not the ASCII `--`.
- **H3 sub-section headers** within each form's H2: `### Header`, `### Stars`, `### Orbit`, `### Comments`, etc.
- **2-column `Field | Value` tables** for grouped per-field sections. Header row `| Field | Value |` followed by separator `|---|---|`.
- **Multi-row tables** for Stars, Objects, Subordinates, Major Bodies subtable. Header row lists all columns.
- **Numeric formatting:**
  - `%.2f` for AU, Orbit
  - `%.3f` for Span, Mass, Diameter, Luminosity, Eccentricity
  - `%.0f` for Temperature (K), Diameter in km
  - `%d` for atmosphere/hydro codes, percentages, counts
  - These mirror the existing PART P / PART P.B plain-text Class IV-P renderer.
- **Nil pointer fields** (e.g., `*Habitability == nil` on a Class IV-P body): emit the H3 section header followed by a single-row `| Status | (not generated) |` table. Never silently drop the section.

---

## Task 1: `stars/markdown.go` — Class 0/I Markdown

**Files:**

- Create: `stars/markdown.go`
- Create: `stars/markdown_test.go`

This task lives entirely in the `stars` package. No `worlds` deps.

- [ ] **Step 1.1: Write the failing test**

Create `stars/markdown_test.go`:

```go
package stars

import (
	"strings"
	"testing"
)

func TestRenderClass0IMarkdown_PopulatedFields(t *testing.T) {
	form := SurveyForm{
		Sector:        "Storr",
		Location:      "0602",
		IISSDesig:     "566-837 (Zed Prime)",
		InitialSurvey: "207-568",
		LastUpdated:   "218-1061",
		SystemAgeGyr:  6.336,
		StellarCount:  5,
		Stars: []SurveyComponent{
			{Component: "Aa", Class: "G7 V", Mass: 0.929, Temperature: 5440, Diameter: 0.967, Luminosity: 0.738},
			{Component: "Ab", Class: "G8 V", Mass: 0.907, Temperature: 5360, Diameter: 0.957, Luminosity: 0.681, Orbit: 0.09, AU: 0.036, Eccentricity: 0.11, PeriodYears: 0.005},
			{Component: "Aab (A)", Class: "—", Mass: 1.836, Luminosity: 1.419, Orbit: 0.09, AU: 0.036, Eccentricity: 0.11, PeriodYears: 0.005, HZCO: 3.3},
		},
	}

	got := RenderClass0IMarkdown(form)

	expected := []string{
		"## IISS Class 0/I Survey — Form 0421B-0I",
		"### Header",
		"| Field | Value |",
		"| Sector | Storr |",
		"| Location | 0602 |",
		"| IISS Designation | 566-837 (Zed Prime) |",
		"| System Age (Gyr) | 6.336 |",
		"| Stellar Count | 5 |",
		"### Stars",
		"| Component | Class | Mass | Temperature | Diameter | Luminosity | Orbit | AU | Eccentricity | Period (y) | HZCO |",
		"| Aa | G7 V | 0.929 |",
		"| Aab (A) | — | 1.836 |",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass0IMarkdown_EmptyMetadataFields(t *testing.T) {
	// Metadata fields left blank should render as em-dash.
	form := SurveyForm{
		SystemAgeGyr: 4.5,
		StellarCount: 1,
		Stars:        []SurveyComponent{{Component: "A", Class: "G2 V", Mass: 1.0}},
	}
	got := RenderClass0IMarkdown(form)
	if !strings.Contains(got, "| Sector | — |") {
		t.Errorf("empty Sector should render as em-dash; got:\n%s", got)
	}
	if !strings.Contains(got, "| IISS Designation | — |") {
		t.Errorf("empty IISSDesig should render as em-dash; got:\n%s", got)
	}
}
```

- [ ] **Step 1.2: Run the failing test**

```bash
go test ./stars/ -run TestRenderClass0IMarkdown -v
```

Expected: compile error — `RenderClass0IMarkdown` is undefined.

- [ ] **Step 1.3: Create the implementation**

Create `stars/markdown.go`:

```go
package stars

import (
	"fmt"
	"strings"
)

// RenderClass0IMarkdown renders the IISS Class 0/I Survey form (WBH p.33,
// Form 0421B-0I) as a Markdown section. Output starts with an H2 form
// heading and contains H3 sub-sections for Header and Stars. Empty/zero
// fields render as em-dash.
func RenderClass0IMarkdown(form SurveyForm) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class 0/I Survey — Form 0421B-0I\n\n")
	writeClass0IHeader(&sb, form)
	writeClass0IStars(&sb, form.Stars)
	if form.Notes != "" {
		fmt.Fprintf(&sb, "### Notes\n\n%s\n\n", form.Notes)
	}
	if form.Comments != "" {
		fmt.Fprintf(&sb, "### Comments\n\n%s\n\n", form.Comments)
	}
	return sb.String()
}

func writeClass0IHeader(sb *strings.Builder, form SurveyForm) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|---|---|\n")
	fmt.Fprintf(sb, "| Sector | %s |\n", emDashIfEmpty(form.Sector))
	fmt.Fprintf(sb, "| Location | %s |\n", emDashIfEmpty(form.Location))
	fmt.Fprintf(sb, "| IISS Designation | %s |\n", emDashIfEmpty(form.IISSDesig))
	fmt.Fprintf(sb, "| Initial Survey | %s |\n", emDashIfEmpty(form.InitialSurvey))
	fmt.Fprintf(sb, "| Last Updated | %s |\n", emDashIfEmpty(form.LastUpdated))
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n", form.SystemAgeGyr)
	fmt.Fprintf(sb, "| Stellar Count | %d |\n", form.StellarCount)
	sb.WriteString("\n")
}

func writeClass0IStars(sb *strings.Builder, stars []SurveyComponent) {
	sb.WriteString("### Stars\n\n")
	sb.WriteString("| Component | Class | Mass | Temperature | Diameter | Luminosity | Orbit | AU | Eccentricity | Period (y) | HZCO |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, c := range stars {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			emDashIfEmpty(c.Component),
			emDashIfEmpty(c.Class),
			formatFloat(c.Mass, 3),
			formatFloatNonZero(c.Temperature, 0),
			formatFloatNonZero(c.Diameter, 3),
			formatFloat(c.Luminosity, 3),
			formatFloatNonZero(c.Orbit, 2),
			formatFloatNonZero(c.AU, 3),
			formatFloatNonZero(c.Eccentricity, 2),
			formatFloatNonZero(c.PeriodYears, 3),
			formatFloatNonZero(c.HZCO, 2),
		)
	}
	sb.WriteString("\n")
}

// emDashIfEmpty returns "—" for empty strings.
func emDashIfEmpty(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// formatFloat renders a float to the given precision; never returns em-dash.
func formatFloat(v float64, prec int) string {
	return fmt.Sprintf("%.*f", prec, v)
}

// formatFloatNonZero renders a float to the given precision, but returns
// em-dash if the value is zero. Use for fields where 0 means "not applicable"
// (e.g., Orbit on the primary's row, Temperature on a barycentre composite).
func formatFloatNonZero(v float64, prec int) string {
	if v == 0 {
		return "—"
	}
	return fmt.Sprintf("%.*f", prec, v)
}
```

- [ ] **Step 1.4: Run the tests and verify they pass**

```bash
go test ./stars/ -run TestRenderClass0IMarkdown -v
```

Expected: both tests PASS.

- [ ] **Step 1.5: Run `task check && task test` for the full project**

```bash
task check && task test
```

Expected: clean. If `task check` reports modernizer drift, stage and re-run.

- [ ] **Step 1.6: Commit**

```bash
git add stars/markdown.go stars/markdown_test.go
git commit -m "$(cat <<'EOF'
feat(stars): RenderClass0IMarkdown — Form 0421B-0I

Markdown formatter for IISS Class 0/I Survey form. Consumes the
existing SurveyForm struct (no duplication of BuildSurveyForm logic).
H2 heading, H3 Header and Stars sub-sections, multi-row Stars table.
Em-dash for empty/zero fields where the book uses "—".

Spec: docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: `worlds/markdown.go` — Class IV-P Markdown (both variants)

**Files:**

- Create: `worlds/markdown.go`
- Create: `worlds/markdown_test.go`

`RenderClass4PMarkdown` dispatches on `body.SizeCode == "0"` to the belt variant (PART P.B) or the terrestrial/moon variant (PART P).

- [ ] **Step 2.1: Write the failing tests**

Create `worlds/markdown_test.go`:

```go
package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderClass4PMarkdown_Terrestrial(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Aab IV d"
	body.SizeCode = "5"
	body.DiameterKm = 8163
	body.MassEarth = 0.27
	body.Orbit = 3.1
	body.Eccentricity = 0.10
	body.Period = Period{Hours: 7050, Years: 0.805}
	body.Group = Group{Designation: "Aab"}
	body.Physical = &BodyPhysical{Density: 1.03, Gravity: 0.66}
	body.Atmosphere = &Atmosphere{Code: 6, Pressure: 1.042, OxygenPartialPressure: 0.292, ScaleHeight: 12.88}
	body.Hydrographics = &Hydrographics{Code: 6, Percent: 62, Profile: "5"}
	sys := stars.System{Primary: stars.Star{AgeGyr: 6.336}}

	got := RenderClass4PMarkdown(body, sys, "Aab IV d")

	expected := []string{
		"## IISS Class IV-P Survey — Form 0407F-IV PART P",
		"### Header",
		"| World | Aab IV d |",
		"| Primary Object(s) | Aab |",
		"| System Age (Gyr) | 6.336 |",
		"### Orbit",
		"| AU | 1.06 |",
		"### Size",
		"| Diameter (km) | 8163 |",
		"| Density | 1.030 |",
		"| Gravity | 0.660 |",
		"### Atmosphere",
		"| Code | 6 |",
		"| Pressure (bar) | 1.042 |",
		"### Hydrographics",
		"| Code | 6 |",
		"| Coverage (%) | 62 |",
		"### Comments",
		"This is the system mainworld.",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_Belt(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Orbit = 2.7
	body.Period = Period{Hours: 5613.16}
	body.Group = Group{Designation: "Aab"}
	body.Belt = &BeltDetails{
		Span:           0.91,
		Composition:    BeltComposition{MTypePct: 25, STypePct: 55, CTypePct: 15, OtherPct: 5},
		Bulk:           4,
		ResourceRating: 8,
		SigSize1Bodies: 2,
		SigSizeSBodies: 19,
	}
	sys := stars.System{Primary: stars.Star{AgeGyr: 6.336}}

	got := RenderClass4PMarkdown(body, sys, "")

	expected := []string{
		"## IISS Class IV-P Survey — Form 0407K-IV PART P.B",
		"| World | Aab PI |",
		"| SAH/UWP | 000 |",
		"### Composition",
		"| m-type (%) | 25 |",
		"| s-type (%) | 55 |",
		"| c-type (%) | 15 |",
		"| other (%) | 5 |",
		"| Bulk | 4 |",
		"| Major Bodies (Size 1) | 2 |",
		"| Major Bodies (Size S) | 19 |",
		"### Resources",
		"| Rating | 8 |",
		"### Major Bodies",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_NilSubSections_RenderPlaceholders(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyTerrestrial
	body.Designation = "Bare Body"
	body.SizeCode = "5"
	body.Group = Group{Designation: "A"}
	// No Atmosphere, Hydrographics, Geology, Biology, Habitability, etc.
	sys := stars.System{}

	got := RenderClass4PMarkdown(body, sys, "")

	expected := []string{
		"### Atmosphere",
		"| Status | (not generated) |",
		"### Hydrographics",
		"### Temperature",
		"### Life",
		"### Habitability",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass4PMarkdown_MainworldMarker(t *testing.T) {
	body := &DetailedPlacement{}
	body.Body = BodyPlanetoidBelt
	body.Designation = "Aab PI"
	body.SizeCode = "0"
	body.Belt = &BeltDetails{}
	sys := stars.System{}

	withMarker := RenderClass4PMarkdown(body, sys, "Aab PI")
	if !strings.Contains(withMarker, "This is the system mainworld.") {
		t.Errorf("expected mainworld marker; got:\n%s", withMarker)
	}

	withoutMarker := RenderClass4PMarkdown(body, sys, "Other")
	if strings.Contains(withoutMarker, "This is the system mainworld.") {
		t.Errorf("unexpected mainworld marker in:\n%s", withoutMarker)
	}
}
```

- [ ] **Step 2.2: Run the failing tests**

```bash
go test ./worlds/ -run TestRenderClass4PMarkdown -v
```

Expected: compile error — `RenderClass4PMarkdown` is undefined.

- [ ] **Step 2.3: Create the implementation**

Create `worlds/markdown.go`:

```go
package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// RenderClass4PMarkdown renders the IISS Class IV-P Survey form for a
// single body (the system mainworld). Dispatches on SizeCode:
//   - "0"  → Form 0407K-IV PART P.B (belt variant, WBH p.139)
//   - else → Form 0407F-IV PART P (terrestrial/moon variant, WBH p.138)
//
// Returns "" if body is nil. mainworldDesignation is used only for the
// "this is the mainworld" marker in the Comments section.
func RenderClass4PMarkdown(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	if body == nil {
		return ""
	}
	if body.SizeCode == "0" {
		return renderClass4PBeltMarkdown(body, sys, mainworldDesignation)
	}
	return renderClass4PTerrestrialMarkdown(body, sys, mainworldDesignation)
}

func renderClass4PTerrestrialMarkdown(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class IV-P Survey — Form 0407F-IV PART P\n\n")
	writeClass4PHeader(&sb, body, sys)
	writeClass4POrbit(&sb, body)
	writeClass4PSize(&sb, body)
	writeClass4PAtmosphere(&sb, body)
	writeClass4PHydrographics(&sb, body)
	writeClass4PRotation(&sb, body)
	writeClass4PTemperature(&sb, body)
	writeClass4PSeismic(&sb, body)
	writeClass4PLife(&sb, body)
	writeClass4PResources(&sb, body)
	writeClass4PHabitability(&sb, body)
	writeClass4PSubordinates(&sb, body)
	writeClass4PComments(&sb, body, mainworldDesignation)
	return sb.String()
}

func renderClass4PBeltMarkdown(body *DetailedPlacement, sys stars.System, mainworldDesignation string) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class IV-P Survey — Form 0407K-IV PART P.B\n\n")
	writeClass4PBeltHeader(&sb, body, sys)
	writeClass4PBeltOrbit(&sb, body)
	writeClass4PBeltComposition(&sb, body)
	writeClass4PBeltResources(&sb, body)
	writeClass4PBeltMajorBodies(&sb, body)
	writeClass4PComments(&sb, body, mainworldDesignation)
	return sb.String()
}

// --- Terrestrial/moon helpers ---

func writeClass4PHeader(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| World | %s |\n", body.Designation)
	fmt.Fprintf(sb, "| Primary Object(s) | %s |\n", body.Group.Designation)
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n\n", sys.Primary.AgeGyr)
}

func writeClass4POrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Orbit\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| O# | %.2f |\n", body.Orbit)
	fmt.Fprintf(sb, "| AU | %.2f |\n", stars.OrbitToAU(body.Orbit))
	fmt.Fprintf(sb, "| Eccentricity | %.3f |\n", body.Eccentricity)
	fmt.Fprintf(sb, "| Period (h) | %.2f |\n\n", body.Period.Hours)
}

func writeClass4PSize(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Size\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Size Code | %s |\n", body.SizeCode)
	fmt.Fprintf(sb, "| Diameter (km) | %.0f |\n", body.DiameterKm)
	density, gravity := 0.0, 0.0
	if body.Physical != nil {
		density = body.Physical.Density
		gravity = body.Physical.Gravity
	}
	fmt.Fprintf(sb, "| Density | %.3f |\n", density)
	fmt.Fprintf(sb, "| Gravity | %.3f |\n", gravity)
	fmt.Fprintf(sb, "| Mass | %.3f |\n\n", body.MassEarth)
}

func writeClass4PAtmosphere(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Atmosphere\n\n")
	if body.Atmosphere == nil {
		writeNotGenerated(sb)
		return
	}
	a := body.Atmosphere
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Code | %d |\n", a.Code)
	fmt.Fprintf(sb, "| Pressure (bar) | %.3f |\n", a.Pressure)
	fmt.Fprintf(sb, "| O₂ (bar) | %.3f |\n", a.OxygenPartialPressure)
	fmt.Fprintf(sb, "| Scale Height | %.2f |\n\n", a.ScaleHeight)
}

func writeClass4PHydrographics(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Hydrographics\n\n")
	if body.Hydrographics == nil {
		writeNotGenerated(sb)
		return
	}
	h := body.Hydrographics
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Code | %d |\n", h.Code)
	fmt.Fprintf(sb, "| Coverage (%%) | %d |\n", h.Percent)
	fmt.Fprintf(sb, "| Profile | %s |\n\n", h.Profile)
}

func writeClass4PRotation(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Rotation\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	if body.DayLength != nil {
		fmt.Fprintf(sb, "| Sidereal (h) | %.2f |\n", body.DayLength.SiderealHours)
		fmt.Fprintf(sb, "| Solar (h) | %.2f |\n", body.DayLength.SolarHours)
		fmt.Fprintf(sb, "| Solar days/year | %.2f |\n", body.DayLength.YearDays)
	} else {
		sb.WriteString("| Day length | (not generated) |\n")
	}
	if body.AxialTilt != nil {
		fmt.Fprintf(sb, "| Axial Tilt (°) | %.2f |\n", body.AxialTilt.Degrees)
	} else {
		sb.WriteString("| Axial Tilt | (not generated) |\n")
	}
	tidalLockText := "no"
	if body.TidalLock != nil && body.TidalLock.LockRatio != "" {
		tidalLockText = body.TidalLock.LockRatio
	}
	fmt.Fprintf(sb, "| Tidal lock | %s |\n", tidalLockText)
	tidesM := 0.0
	if body.TidalEffects != nil {
		tidesM = body.TidalEffects.Total
	}
	fmt.Fprintf(sb, "| Tides (m) | %.2f |\n\n", tidesM)
}

func writeClass4PTemperature(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Temperature\n\n")
	if body.Temperature == nil {
		writeNotGenerated(sb)
		return
	}
	t := body.Temperature
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| High (K) | %.0f |\n", t.HighK)
	fmt.Fprintf(sb, "| Mean (K) | %.0f |\n", t.MeanK)
	fmt.Fprintf(sb, "| Low (K) | %.0f |\n", t.LowK)
	fmt.Fprintf(sb, "| Luminosity | %.3f |\n", t.Luminosity)
	fmt.Fprintf(sb, "| Albedo | %.2f |\n", t.Albedo)
	fmt.Fprintf(sb, "| Greenhouse | %.2f |\n\n", t.GreenhouseFactor)
}

func writeClass4PSeismic(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Seismic\n\n")
	if body.Geology == nil {
		writeNotGenerated(sb)
		return
	}
	g := body.Geology
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Total Seismic Stress | %d |\n", g.TotalSeismicStress)
	fmt.Fprintf(sb, "| Residual Stress | %d |\n", g.ResidualSeismicStress)
	fmt.Fprintf(sb, "| Tidal Stress | %d |\n", g.TidalStressFactor)
	fmt.Fprintf(sb, "| Tidal Heating | %d |\n", g.TidalHeatingFactor)
	fmt.Fprintf(sb, "| Tectonic Plates | %d |\n\n", g.TectonicPlates)
}

func writeClass4PLife(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Life\n\n")
	if body.Biology == nil {
		writeNotGenerated(sb)
		return
	}
	b := body.Biology
	sophontStr := "no"
	if b.HasNativeSophont {
		sophontStr = "yes"
	} else if b.HadExtinctSophont {
		sophontStr = "extinct"
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Biomass | %s |\n", string(eHexDigit(b.Biomass)))
	fmt.Fprintf(sb, "| Biocomplexity | %d |\n", b.Biocomplexity)
	fmt.Fprintf(sb, "| Sophonts? | %s |\n", sophontStr)
	fmt.Fprintf(sb, "| Biodiversity | %d |\n", b.Biodiversity)
	fmt.Fprintf(sb, "| Compatibility | %d |\n\n", b.Compatibility)
}

func writeClass4PResources(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Resources\n\n")
	if body.Biology == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %s |\n\n", string(eHexDigit(body.Biology.ResourceRating)))
}

func writeClass4PHabitability(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Habitability\n\n")
	if body.Habitability == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %d |\n\n", body.Habitability.Rating)
}

func writeClass4PSubordinates(sb *strings.Builder, body *DetailedPlacement) {
	if len(body.Moons) == 0 {
		return
	}
	sb.WriteString("### Subordinates\n\n")
	sb.WriteString("| Designation | Size | Diameter (km) | Orbit (km) | Eccentricity | Period (h) |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")
	for _, m := range body.Moons {
		fmt.Fprintf(sb, "| %s | %s | %.0f | %d | %.3f | %.2f |\n",
			m.Designation, m.SizeCode, m.DiameterKm, m.OrbitKm, m.Eccentricity, m.PeriodHours)
	}
	sb.WriteString("\n")
}

// --- Belt helpers ---

func writeClass4PBeltHeader(sb *strings.Builder, body *DetailedPlacement, sys stars.System) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| World | %s |\n", body.Designation)
	sb.WriteString("| SAH/UWP | 000 |\n")
	fmt.Fprintf(sb, "| Primary Object(s) | %s |\n", body.Group.Designation)
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n\n", sys.Primary.AgeGyr)
}

func writeClass4PBeltOrbit(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Orbit\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| O# | %.2f |\n", body.Orbit)
	fmt.Fprintf(sb, "| AU | %.2f |\n", stars.OrbitToAU(body.Orbit))
	if body.Belt != nil {
		fmt.Fprintf(sb, "| Span (Orbit#s) | %.3f |\n", body.Belt.Span)
	} else {
		sb.WriteString("| Span | (not available) |\n")
	}
	fmt.Fprintf(sb, "| Period (h) | %.2f |\n\n", body.Period.Hours)
}

func writeClass4PBeltComposition(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Composition\n\n")
	if body.Belt == nil {
		writeNotGenerated(sb)
		return
	}
	c := body.Belt.Composition
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| m-type (%%) | %d |\n", c.MTypePct)
	fmt.Fprintf(sb, "| s-type (%%) | %d |\n", c.STypePct)
	fmt.Fprintf(sb, "| c-type (%%) | %d |\n", c.CTypePct)
	fmt.Fprintf(sb, "| other (%%) | %d |\n", c.OtherPct)
	fmt.Fprintf(sb, "| Bulk | %d |\n", body.Belt.Bulk)
	fmt.Fprintf(sb, "| Major Bodies (Size 1) | %d |\n", body.Belt.SigSize1Bodies)
	fmt.Fprintf(sb, "| Major Bodies (Size S) | %d |\n\n", body.Belt.SigSizeSBodies)
}

func writeClass4PBeltResources(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Resources\n\n")
	if body.Belt == nil {
		writeNotGenerated(sb)
		return
	}
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Rating | %d |\n\n", body.Belt.ResourceRating)
}

func writeClass4PBeltMajorBodies(sb *strings.Builder, body *DetailedPlacement) {
	sb.WriteString("### Major Bodies\n\n")
	if body.Belt == nil {
		writeNotGenerated(sb)
		return
	}
	fmt.Fprintf(sb, "_Counts only: %d size-1 + %d size-S; per-body detail not generated._\n\n",
		body.Belt.SigSize1Bodies, body.Belt.SigSizeSBodies)
}

// --- Shared helpers ---

func writeClass4PComments(sb *strings.Builder, body *DetailedPlacement, mainworldDesignation string) {
	sb.WriteString("### Comments\n\n")
	if mainworldDesignation != "" && body.Designation == mainworldDesignation {
		sb.WriteString("This is the system mainworld.\n\n")
	}
}

func writeNotGenerated(sb *strings.Builder) {
	sb.WriteString("| Field | Value |\n|---|---|\n")
	sb.WriteString("| Status | (not generated) |\n\n")
}
```

- [ ] **Step 2.4: Run the tests and verify they pass**

```bash
go test ./worlds/ -run TestRenderClass4PMarkdown -v
```

Expected: all four tests PASS.

- [ ] **Step 2.5: Run `task check && task test`**

```bash
task check && task test
```

Expected: clean.

- [ ] **Step 2.6: Commit**

```bash
git add worlds/markdown.go worlds/markdown_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): RenderClass4PMarkdown — Forms 0407F-IV / 0407K-IV

Markdown formatter for IISS Class IV-P Survey form. Dispatches on
SizeCode == "0" to PART P.B (belt) or PART P (terrestrial/moon).
H2 form heading, H3 sub-sections per book layout, 2-column
Field|Value tables for grouped data, Subordinates table for moons.
Defensive against nil pointer fields (Atmosphere, Hydrographics, etc.)
— renders explicit "(not generated)" placeholder rather than silent
zero or skip.

Spec: docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: `worlds/markdown.go` — Class II/III Markdown

**Files:**

- Modify: `worlds/markdown.go` (append `RenderClass23Markdown`)
- Modify: `worlds/markdown_test.go` (append tests)

This task adds a new exported function to the file Task 2 created.

- [ ] **Step 3.1: Write the failing test**

Append to `worlds/markdown_test.go`:

```go
func TestRenderClass23Markdown_PopulatedFields(t *testing.T) {
	form := IISSClass23Form{
		SurveyForm: stars.SurveyForm{
			Sector:        "Storr",
			Location:      "0602",
			IISSDesig:     "566-837 (Zed Prime)",
			InitialSurvey: "207-568",
			LastUpdated:   "218-1061",
			SystemAgeGyr:  6.336,
			StellarCount:  5,
			Stars: []stars.SurveyComponent{
				{Component: "Aa", Class: "G7 V", Mass: 0.929, Temperature: 5440, Diameter: 0.967, Luminosity: 0.738},
			},
		},
		GasGiants:      4,
		PlanetoidBelts: 2,
		Terrestrials:   12,
		ClassIIIStatus: true,
		Objects: []ObjectRow{
			{Primary: "Aab", Designation: "Aab IV", Orbit: 3.1, AU: 1.06, Ecc: 0.10, PeriodStr: "0.805y", SAH: "GLE", Sub: "5", Notes: "1,200⊕, HZ"},
			{Primary: "Aab", Designation: "Aab IV d", SAH: "566*", Sub: "", Notes: ""},
		},
	}

	got := RenderClass23Markdown(form)

	expected := []string{
		"## IISS Class II/III Survey — Form 0421D-II.III",
		"### Header",
		"| IISS Designation | 566-837 (Zed Prime) |",
		"| Class III Status | yes |",
		"### Object Counts",
		"| Stellar | 5 |",
		"| Gas Giants | 4 |",
		"| Planetoid Belts | 2 |",
		"| Terrestrials | 12 |",
		"### Stars",
		"| Aa | G7 V |",
		"### Objects",
		"| Primary | Designation | Orbit | AU | Eccentricity | Period | SAH/UWP | Subs | Notes |",
		"| Aab | Aab IV |",
		"| Aab | Aab IV d |",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderClass23Markdown_ClassIIStatusRendersAsNo(t *testing.T) {
	form := IISSClass23Form{
		SurveyForm:     stars.SurveyForm{IISSDesig: "Test"},
		ClassIIIStatus: false,
	}
	got := RenderClass23Markdown(form)
	if !strings.Contains(got, "| Class III Status | no |") {
		t.Errorf("expected 'Class III Status | no'; got:\n%s", got)
	}
}
```

- [ ] **Step 3.2: Run the failing test**

```bash
go test ./worlds/ -run TestRenderClass23Markdown -v
```

Expected: compile error — `RenderClass23Markdown` is undefined.

- [ ] **Step 3.3: Append the implementation**

Append to `worlds/markdown.go`:

```go
// RenderClass23Markdown renders the IISS Class II/III Survey form (WBH
// pp.60-67, Form 0421D-II.III) as a Markdown section. Output starts
// with an H2 form heading and contains H3 sub-sections for Header,
// Object Counts, Stars, and Objects.
func RenderClass23Markdown(form IISSClass23Form) string {
	var sb strings.Builder
	sb.WriteString("## IISS Class II/III Survey — Form 0421D-II.III\n\n")
	writeClass23Header(&sb, form)
	writeClass23ObjectCounts(&sb, form)
	writeClass23Stars(&sb, form.Stars)
	writeClass23Objects(&sb, form.Objects)
	if form.Notes != "" {
		fmt.Fprintf(&sb, "### Notes\n\n%s\n\n", form.Notes)
	}
	if form.Comments != "" {
		fmt.Fprintf(&sb, "### Comments\n\n%s\n\n", form.Comments)
	}
	return sb.String()
}

func writeClass23Header(sb *strings.Builder, form IISSClass23Form) {
	sb.WriteString("### Header\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Sector | %s |\n", emDashIfEmpty23(form.Sector))
	fmt.Fprintf(sb, "| Location | %s |\n", emDashIfEmpty23(form.Location))
	fmt.Fprintf(sb, "| IISS Designation | %s |\n", emDashIfEmpty23(form.IISSDesig))
	fmt.Fprintf(sb, "| Initial Survey | %s |\n", emDashIfEmpty23(form.InitialSurvey))
	fmt.Fprintf(sb, "| Last Updated | %s |\n", emDashIfEmpty23(form.LastUpdated))
	fmt.Fprintf(sb, "| System Age (Gyr) | %.3f |\n", form.SystemAgeGyr)
	classIIIStr := "no"
	if form.ClassIIIStatus {
		classIIIStr = "yes"
	}
	fmt.Fprintf(sb, "| Class III Status | %s |\n\n", classIIIStr)
}

func writeClass23ObjectCounts(sb *strings.Builder, form IISSClass23Form) {
	sb.WriteString("### Object Counts\n\n")
	sb.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Stellar | %d |\n", form.StellarCount)
	fmt.Fprintf(sb, "| Gas Giants | %d |\n", form.GasGiants)
	fmt.Fprintf(sb, "| Planetoid Belts | %d |\n", form.PlanetoidBelts)
	fmt.Fprintf(sb, "| Terrestrials | %d |\n\n", form.Terrestrials)
}

// writeClass23Stars renders the same Stars table as Class 0/I, but per
// the spec, no shared helper is extracted; this function reads the
// same data fresh.
func writeClass23Stars(sb *strings.Builder, components []stars.SurveyComponent) {
	sb.WriteString("### Stars\n\n")
	sb.WriteString("| Component | Class | Mass | Temperature | Diameter | Luminosity | Orbit | AU | Eccentricity | Period (y) | HZCO | MAO |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, c := range components {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			emDashIfEmpty23(c.Component),
			emDashIfEmpty23(c.Class),
			fmt.Sprintf("%.3f", c.Mass),
			floatNonZero23(c.Temperature, 0),
			floatNonZero23(c.Diameter, 3),
			fmt.Sprintf("%.3f", c.Luminosity),
			floatNonZero23(c.Orbit, 2),
			floatNonZero23(c.AU, 3),
			floatNonZero23(c.Eccentricity, 2),
			floatNonZero23(c.PeriodYears, 3),
			floatNonZero23(c.HZCO, 2),
			floatNonZero23(c.MAO, 2),
		)
	}
	sb.WriteString("\n")
}

func writeClass23Objects(sb *strings.Builder, rows []ObjectRow) {
	sb.WriteString("### Objects\n\n")
	sb.WriteString("| Primary | Designation | Orbit | AU | Eccentricity | Period | SAH/UWP | Subs | Notes |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			emDashIfEmpty23(r.Primary),
			emDashIfEmpty23(r.Designation),
			floatNonZero23(r.Orbit, 2),
			floatNonZero23(r.AU, 3),
			floatNonZero23(r.Ecc, 3),
			emDashIfEmpty23(r.PeriodStr),
			emDashIfEmpty23(r.SAH),
			emDashIfEmpty23(r.Sub),
			emDashIfEmpty23(r.Notes),
		)
	}
	sb.WriteString("\n")
}

func emDashIfEmpty23(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func floatNonZero23(v float64, prec int) string {
	if v == 0 {
		return "—"
	}
	return fmt.Sprintf("%.*f", prec, v)
}
```

**Note:** The two helpers `emDashIfEmpty23` and `floatNonZero23` are package-private duplicates of the `stars/markdown.go` helpers. The spec explicitly chose this over extracting a shared package — the duplication is local (a few lines) and keeps the package boundaries clean.

- [ ] **Step 3.4: Run the tests and verify they pass**

```bash
go test ./worlds/ -run TestRenderClass23Markdown -v
```

Expected: both new tests PASS. The Task 2 tests should still pass.

- [ ] **Step 3.5: Run `task check && task test`**

```bash
task check && task test
```

Expected: clean.

- [ ] **Step 3.6: Commit**

```bash
git add worlds/markdown.go worlds/markdown_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): RenderClass23Markdown — Form 0421D-II.III

Markdown formatter for IISS Class II/III Survey form. Consumes the
existing IISSClass23Form struct (which embeds SurveyForm). H2 form
heading, H3 sub-sections for Header, Object Counts, Stars, and
Objects. Stars table is rendered fresh from the embedded
SurveyComponent rows — no shared helper with Class 0/I, per spec.

Spec: docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: `worlds/markdown_system.go` — Top-level orchestrator

**Files:**

- Create: `worlds/markdown_system.go`
- Create: `worlds/markdown_system_test.go`

The orchestrator chains the three Markdown formatters under H1/H2 headings. It builds the Class 0/I `SurveyForm` fresh via `stars.BuildSurveyForm` and reuses the Class II/III form already on `SystemDetail.Survey`.

- [ ] **Step 4.1: Write the failing tests**

Create `worlds/markdown_system_test.go`:

```go
package worlds

import (
	"strings"
	"testing"

	"wbh/stars"
)

func TestRenderSystemMarkdown_AllSectionsWhenMainworldExists(t *testing.T) {
	sd := SystemDetail{
		Survey: IISSClass23Form{
			SurveyForm: stars.SurveyForm{
				IISSDesig:    "Test System 1234",
				SystemAgeGyr: 4.5,
				StellarCount: 1,
				Stars:        []stars.SurveyComponent{{Component: "A", Class: "G2 V", Mass: 1.0, Luminosity: 1.0}},
			},
			Terrestrials: 1,
		},
		MainworldDesignation: "A III",
	}
	sd.Detailed = []DetailedPlacement{{
		Placement: Placement{},
	}}
	sd.Detailed[0].Body = BodyTerrestrial
	sd.Detailed[0].Designation = "A III"
	sd.Detailed[0].SizeCode = "5"
	sd.Detailed[0].Group = Group{Designation: "A"}

	sys := stars.System{Primary: stars.Star{
		SpectralType:    stars.SpectralType{Letter: 'G', Subtype: 2},
		LuminosityClass: stars.V,
		Mass:            1.0,
		Luminosity:      1.0,
		AgeGyr:          4.5,
	}}

	got := RenderSystemMarkdown(sd, sys)

	expected := []string{
		"# Star System: Test System 1234",
		"## IISS Class 0/I Survey — Form 0421B-0I",
		"## IISS Class II/III Survey — Form 0421D-II.III",
		"## IISS Class IV-P Survey — Form 0407F-IV PART P",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}

	// H2s must appear in book order: 0/I before II/III before IV-P.
	idx0I := strings.Index(got, "## IISS Class 0/I")
	idx23 := strings.Index(got, "## IISS Class II/III")
	idx4P := strings.Index(got, "## IISS Class IV-P")
	if !(idx0I < idx23 && idx23 < idx4P) {
		t.Errorf("H2 sections not in book order: 0/I=%d, II/III=%d, IV-P=%d", idx0I, idx23, idx4P)
	}
}

func TestRenderSystemMarkdown_OmitsClass4PWhenNoMainworld(t *testing.T) {
	sd := SystemDetail{
		Survey: IISSClass23Form{
			SurveyForm: stars.SurveyForm{IISSDesig: "Empty System", StellarCount: 1},
		},
		MainworldDesignation: "",
	}
	sys := stars.System{Primary: stars.Star{AgeGyr: 4.5}}

	got := RenderSystemMarkdown(sd, sys)

	if !strings.Contains(got, "## IISS Class 0/I Survey") {
		t.Errorf("missing Class 0/I H2; got:\n%s", got)
	}
	if !strings.Contains(got, "## IISS Class II/III Survey") {
		t.Errorf("missing Class II/III H2; got:\n%s", got)
	}
	if strings.Contains(got, "## IISS Class IV-P Survey") {
		t.Errorf("Class IV-P H2 must be absent when MainworldDesignation == \"\"; got:\n%s", got)
	}
}

func TestRenderSystemMarkdown_BeltMainworldUsesPartPB(t *testing.T) {
	sd := SystemDetail{
		Survey: IISSClass23Form{
			SurveyForm: stars.SurveyForm{IISSDesig: "Belt System", StellarCount: 1},
		},
		MainworldDesignation: "A PI",
	}
	belt := DetailedPlacement{Placement: Placement{}}
	belt.Body = BodyPlanetoidBelt
	belt.Designation = "A PI"
	belt.SizeCode = "0"
	belt.Belt = &BeltDetails{ResourceRating: 8}
	sd.Detailed = []DetailedPlacement{belt}

	sys := stars.System{Primary: stars.Star{AgeGyr: 4.5}}

	got := RenderSystemMarkdown(sd, sys)

	if !strings.Contains(got, "Form 0407K-IV PART P.B") {
		t.Errorf("expected belt-variant Form 0407K-IV PART P.B; got:\n%s", got)
	}
	if strings.Contains(got, "Form 0407F-IV PART P\n") {
		t.Errorf("should not emit terrestrial PART P for a belt mainworld; got:\n%s", got)
	}
}
```

- [ ] **Step 4.2: Run the failing tests**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown -v
```

Expected: compile error — `RenderSystemMarkdown` is undefined.

- [ ] **Step 4.3: Create the implementation**

Create `worlds/markdown_system.go`:

```go
package worlds

import (
	"fmt"
	"strings"

	"wbh/stars"
)

// RenderSystemMarkdown is the top-level Markdown renderer. Output:
//
//   - H1 system title
//   - H2 IISS Class 0/I Survey
//   - H2 IISS Class II/III Survey
//   - H2 IISS Class IV-P Survey (omitted when sd.MainworldDesignation == "")
//
// Class 0/I is read from sd.Survey.SurveyForm (already populated by
// DetailSystem); Class II/III is read from sd.Survey; Class IV-P
// resolves the mainworld body from sd.Detailed and dispatches to the
// right variant (PART P / PART P.B) based on its SizeCode.
func RenderSystemMarkdown(sd SystemDetail, sys stars.System) string {
	var sb strings.Builder
	title := sd.Survey.IISSDesig
	if title == "" {
		title = "(unnamed)"
	}
	fmt.Fprintf(&sb, "# Star System: %s\n\n", title)

	// Class 0/I — embedded in sd.Survey.
	sb.WriteString(stars.RenderClass0IMarkdown(sd.Survey.SurveyForm))

	// Class II/III — already on sd.Survey.
	sb.WriteString(RenderClass23Markdown(sd.Survey))

	// Class IV-P — only when a mainworld was picked.
	if sd.MainworldDesignation != "" {
		if mw := findMainworld(sd, sd.MainworldDesignation); mw != nil {
			sb.WriteString(RenderClass4PMarkdown(mw, sys, sd.MainworldDesignation))
		}
	}

	return sb.String()
}

// findMainworld locates the body in sd.Detailed (or its moons) whose
// Designation matches mainworld. Returns nil if not found.
func findMainworld(sd SystemDetail, mainworld string) *DetailedPlacement {
	for i := range sd.Detailed {
		if sd.Detailed[i].Designation == mainworld {
			return &sd.Detailed[i]
		}
		for j := range sd.Detailed[i].Moons {
			m := &sd.Detailed[i].Moons[j]
			if m.Designation == mainworld {
				// Synthesize a DetailedPlacement view of the moon for
				// rendering — same pattern as buildMoonPlacementView in
				// system_detail.go.
				return buildMoonPlacementView(m, &sd.Detailed[i])
			}
		}
	}
	return nil
}
```

- [ ] **Step 4.4: Run the tests and verify they pass**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown -v
```

Expected: all three tests PASS.

- [ ] **Step 4.5: Run `task check && task test`**

```bash
task check && task test
```

Expected: clean.

- [ ] **Step 4.6: Commit**

```bash
git add worlds/markdown_system.go worlds/markdown_test.go worlds/markdown_system_test.go
git commit -m "$(cat <<'EOF'
feat(worlds): RenderSystemMarkdown — top-level orchestrator

Chains Class 0/I + Class II/III + Class IV-P Markdown formatters
under H1/H2 headings in book order. Class IV-P section is omitted
when MainworldDesignation == "" (e.g., GG-only systems). Uses the
existing buildMoonPlacementView helper to synthesize a DetailedPlacement
view of moon mainworlds for rendering.

Spec: docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: `cmd/wbh/main.go` — wire the Markdown path and flip the default

**Files:**

- Modify: `cmd/wbh/main.go`
- Modify: `cmd/wbh/main_test.go` (or create if absent)

- [ ] **Step 5.1: Read the existing main.go and main_test.go**

```bash
cat cmd/wbh/main.go
cat cmd/wbh/main_test.go
```

Confirm the current default is `"json"` and the `format` switch has cases for `"json"` and `"short"`.

- [ ] **Step 5.2: Write the failing tests**

If `cmd/wbh/main_test.go` already has tests, append to it. Otherwise create:

```go
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRun_DefaultIsMarkdown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v (stderr=%q)", err, stderr.String())
	}
	out := stdout.String()
	if !strings.HasPrefix(strings.TrimLeft(out, "\n"), "# Star System:") {
		t.Errorf("expected output to start with '# Star System:'; got first 80 chars: %q", out[:min(80, len(out))])
	}
	if !strings.Contains(out, "## IISS Class 0/I Survey") {
		t.Errorf("expected Class 0/I H2 in default output")
	}
}

func TestRun_FormatMarkdownExplicit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42", "-format", "markdown"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "## IISS Class 0/I Survey") {
		t.Errorf("expected Class 0/I H2 with -format markdown")
	}
}

func TestRun_FormatJSONStillWorks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42", "-format", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Errorf("expected valid JSON; got error %v on output:\n%s", err, stdout.String())
	}
}

func TestRun_FormatShortStillWorks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42", "-format", "short"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	out := strings.TrimSpace(stdout.String())
	if strings.Contains(out, "\n") {
		t.Errorf("expected single-line short output; got:\n%s", out)
	}
}
```

- [ ] **Step 5.3: Run the failing tests**

```bash
go test ./cmd/wbh/ -v
```

Expected: at minimum `TestRun_DefaultIsMarkdown` FAILS (current default is `"json"`, output starts with `{`). Other tests may pass already if the JSON/short paths are unchanged, or fail if they regressed.

- [ ] **Step 5.4: Modify `cmd/wbh/main.go`**

Update the flag default and add a `markdown` case. The full updated `run` function:

```go
func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("wbh", flag.ContinueOnError)
	fs.SetOutput(stderr)
	seed := fs.Int64("seed", 0, "random seed (0 = time-based)")
	format := fs.String("format", "markdown", "output format: markdown | json | short")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}
	r := roller.NewSeeded(s)
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{
		WithVariance: true,
		Accuracy:     2,
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	switch *format {
	case "markdown":
		sp, err := worlds.GenerateSystemPlacement(r, sys)
		if err != nil {
			return fmt.Errorf("system placement: %w", err)
		}
		sd, err := worlds.DetailSystem(r, sys, sp, worlds.IISSClass23Header{})
		if err != nil {
			return fmt.Errorf("detail system: %w", err)
		}
		_, err = fmt.Fprint(stdout, worlds.RenderSystemMarkdown(sd, sys))
		return err
	case "json":
		form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{})
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(form)
	case "short":
		_, err := fmt.Fprintln(stdout, stars.ShortProfile(sys))
		return err
	default:
		return fmt.Errorf("unknown format: %q (want markdown, json, or short)", *format)
	}
}
```

Add `"wbh/worlds"` to the imports if not already present.

- [ ] **Step 5.5: Run the tests and verify they pass**

```bash
go test ./cmd/wbh/ -v
```

Expected: all four tests PASS.

- [ ] **Step 5.6: Run `task check && task test`**

```bash
task check && task test
```

Expected: clean.

- [ ] **Step 5.7: Commit**

```bash
git add cmd/wbh/main.go cmd/wbh/main_test.go
git commit -m "$(cat <<'EOF'
feat(cmd/wbh): -format markdown is now the default

Wires worlds.RenderSystemMarkdown through the CLI. The markdown path
runs GenerateSystemPlacement → DetailSystem → RenderSystemMarkdown
(unlike json which stays Class 0/I only). Default -format flips from
"json" to "markdown"; -format json and -format short continue to
work unchanged.

Spec: docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Golden-file Zed snapshot test

**Files:**

- Create: `worlds/testdata/zed_markdown.golden`
- Modify: `worlds/markdown_system_test.go` (append the golden test)

The Zed worked-example pipeline (used by `TestZed_FullDetail_*` tests) drives a `roller.Scripted` through the full generation. We capture that pipeline's `RenderSystemMarkdown` output as a golden file and assert against it.

- [ ] **Step 6.1: Find the existing Zed pipeline test and extract its setup**

Run `grep -rn "TestZed_FullDetail" worlds/` and read the test that builds the full Zed `SystemDetail`. The test uses something like:

```go
r := roller.NewScripted(/* the book's exact roll sequence */)
sys, _ := stars.GenerateSystem(r, ...)
sp, _ := worlds.GenerateSystemPlacement(r, sys)
sd, _ := worlds.DetailSystem(r, sys, sp, hdr)
```

- [ ] **Step 6.2: Append the golden test to `worlds/markdown_system_test.go`**

Add this test (along with the `-update` flag handling):

```go
import "flag"
import "os"

var update = flag.Bool("update", false, "update golden files")

func TestRenderSystemMarkdown_ZedGolden(t *testing.T) {
	// Reuse the Zed pipeline from the existing acceptance test.
	// The exact roller.NewScripted sequence and IISSClass23Header
	// values must match TestZed_FullDetail_3A2b's setup so the same
	// SystemDetail is produced.
	sd, sys := buildZedSystemDetail(t)

	got := RenderSystemMarkdown(sd, sys)

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
```

**`buildZedSystemDetail` helper:** This builder must produce the exact same `SystemDetail` that the existing `TestZed_FullDetail_3B-final` (or latest acceptance test) produces. Either:

- (a) Extract a shared helper in a non-`_test.go` file, OR
- (b) Look for an existing exported helper (e.g., `BuildZedSystemDetail`), OR
- (c) Inline the script sequence (large but self-contained).

Most likely (a) is cleanest: copy the body of the Zed acceptance test's setup into a non-`_test.go` package-internal helper. If the acceptance test already factored this out, just call the existing helper.

The implementer should grep `worlds/` for the Zed roller-script setup, identify the most recent worked-example test (3B-final), and reuse its helpers if any. If none exist, factor out a `buildZedSystemDetail(t *testing.T) (SystemDetail, stars.System)` helper in a new file (not as `_test.go` content if multiple tests need it) or a shared `_test.go` file via Go's test-file sharing.

- [ ] **Step 6.3: Run with `-update` to create the golden file**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update
ls -la worlds/testdata/zed_markdown.golden
```

Expected: file created, ~3-10 KB (depends on Zed's body count).

- [ ] **Step 6.4: Inspect the golden file by hand**

```bash
less worlds/testdata/zed_markdown.golden
```

Verify:

- Starts with `# Star System: 566-837 (Zed Prime)` (or whatever the Zed acceptance test sets as IISSDesig)
- Has all three H2 sections in book order
- Stars table includes rows for Aa, Ab, Aab (A), B, AB, Ca, Cb, Cab (C), ABC (or whatever Zed produces)
- Class IV-P uses Form 0407F-IV PART P (Zed Prime is a moon, not a belt)
- Tables are well-formed Markdown (every row has the right column count)

If the output looks wrong, fix the formatters and regenerate.

- [ ] **Step 6.5: Run without `-update` to confirm the test passes**

```bash
go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -v
```

Expected: PASS.

- [ ] **Step 6.6: Run `task check && task test`**

```bash
task check && task test
```

Expected: clean.

- [ ] **Step 6.7: Commit**

```bash
git add worlds/markdown_system_test.go worlds/testdata/zed_markdown.golden
# If a non-test helper file was created (e.g., worlds/zed_fixture.go), add it too.
git commit -m "$(cat <<'EOF'
test(worlds): golden-file snapshot of Zed Markdown output

Captures the full RenderSystemMarkdown output for the Zed worked
example as a golden file. Test runs against the snapshot; refresh
intentional format changes via:

    go test ./worlds/ -run TestRenderSystemMarkdown_ZedGolden -update

Data fidelity is already covered by TestZed_FullDetail_*; this test
covers rendering stability — catches accidental changes to layout,
formatting, or section ordering across the three IISS forms.

Spec: docs/specs/2026-05-06-cmd-wbh-markdown-output-design.md

Co-Authored-By: Claude Code <noreply@anthropic.com>
EOF
)"
```

---

## Final verification

- [ ] **Step F.1: Run the full project gate**

```bash
task check && task test
```

Expected: clean.

- [ ] **Step F.2: Manual smoke test**

```bash
go run ./cmd/wbh -seed 42 | head -50
```

Expected: Markdown output starting with `# Star System:`, followed by `## IISS Class 0/I Survey — Form 0421B-0I`, well-formed table rows.

```bash
go run ./cmd/wbh -seed 42 -format json | jq -r '.IISSDesig' 2>&1 | head -1
```

Expected: prints the system designation (verifies JSON path still works).

```bash
go run ./cmd/wbh -seed 42 -format short
```

Expected: a single-line short profile.

- [ ] **Step F.3: Verify the spec's success criteria**

Manually confirm against the spec's success-criteria list:

- `go run ./cmd/wbh -seed 42` (no `-format`) emits Markdown to stdout starting with `# Star System: ...` ✓
- The Zed seed pipeline produces a stable, golden-matched Markdown rendering ✓
- `cmd/wbh -format json` and `-format short` continue to produce their existing output ✓
- `task check && task test` clean ✓

Done.
