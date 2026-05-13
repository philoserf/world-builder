# WBH Internal Inconsistencies

The _World Builder's Handbook_ (Lanesskog 2023) contains documented internal contradictions: tables that disagree with each other, formulas that disagree with worked examples, and footnotes that disagree with the IISS-form printed values. The project commits to one interpretation per case, encoded in code with the chosen value asserted by a test and the divergence documented here. No runtime toggles for either interpretation.

The rules:

- **One interpretation per inconsistency.** No `*Opts` field, no `t.Logf` divergence note, no flag.
- **Each entry cites both sources** (the conflicting pages or sources within the book).
- **Each entry states the chosen interpretation** and the reason for the choice.
- **Each entry names the verification target.** Where applicable, the chosen interpretation is the one that reproduces the canonical Zed Prime IISS form on WBH pp.141–142.

## The decision rule

Across the six entries below, one heuristic emerges:

> When a formula table and a worked example disagree, choose the interpretation that reproduces the canonical Zed Prime IISS form on pp.141–142. Worked examples on the IISS forms are the printed verification targets. Where the formula matches the form, follow the formula. Where the worked example matches the form (and the formula does not), follow the worked example.

This is not always "trust the table" or "trust the worked example" — it is "trust whichever interpretation makes the canonical example reproducible." The rule applies consistently and is decisive for entries 4–6; for entries 1–3 the formula and example agree on the value but disagree elsewhere.

## Inconsistency 1: WBH p.19 luminosity table vs p.42 HZCO table

**Sources.** WBH p.19 publishes stellar luminosities for spectral type/luminosity-class cells. WBH p.42 publishes HZCO (habitable-zone central orbit) values derived from those luminosities. The formula `HZCO_Orbit = AUToOrbit(sqrt(luminosity))` reproduces 83 of 88 populated p.42 cells within ±5%.

**Conflict.** Five Class VI cells diverge by more than 5% between the two tables:

| Cell  | p.19 luminosity | Formula HZCO | p.42 HZCO | Required L for p.42 |
| ----- | --------------: | -----------: | --------: | ------------------: |
| G5 VI |            0.43 |         1.85 |       2.5 |         0.72 (+67%) |
| K0 VI |            0.23 |         1.27 |       1.9 |         0.45 (+96%) |
| K5 VI |           0.083 |         0.72 |       1.3 |        0.24 (+189%) |
| M9 V  |         0.00029 |        0.043 |      0.04 |      0.00025 (−14%) |
| M9 VI |         0.00019 |        0.035 |      0.03 |      0.00014 (−26%) |

**Chosen interpretation.** **Implement the formula** (`HZCO = AUToOrbit(sqrt(luminosity))`) using p.19 luminosities as the source of truth. The five cells where p.42 disagrees are accepted as inter-table drift; the tests verify the formula against the 83 consistent cells and explicitly skip the five known-divergent cells. The skip list is hard-coded with comments citing this document.

**Why.** The formula is procedurally specified; the p.42 table is a derived display. Choosing the formula keeps the implementation honest about what HZCO means (a function of luminosity). The Class VI divergence is small in absolute orbit-units and does not affect any worked example we verify against.

## Inconsistency 2: WBH p.58 sizing table vs p.63 IISS form for Aab IV-d (Zed Prime)

**Sources.** WBH p.58 sizing-results table for Aab IV reads: `2, S, S, S, S` (five moons sized 2, S, S, S, S — `S` = small body, < 600 km). WBH p.63 IISS Class II/III form for Aab IV's Notes column reads: `1,200⊕, HZ, 200, S, S, 566*, S` — the fourth moon is the mainworld candidate **Zed Prime** at SAH 566 (Size 5, Atmosphere 6, Hydrographics 6).

**Conflict.** If d-moon were Size S as p.58 says, it could not be a mainworld candidate (Size S = small body < 600 km, no atmosphere capacity). But the form makes Zed Prime the mainworld at Size 5.

**Chosen interpretation.** **Treat p.63 (the form) as authoritative.** The d-moon's underlying SizeCode is "5".

**Why.** The book authors evidently updated the p.63 form to make Zed Prime habitable but left the p.58 sizing table unrevised. The form is the canonical verification target on pp.141–142; the sizing table appears to be a typo or stale text. The Zed Prime fixture in the harness encodes Size=5 directly.

## Inconsistency 3: WBH temperature chapter (pp.108–126) — three table-vs-text divergences

The temperature chapter contains three independent contradictions surfaced during pass-1's 3A2b-temp implementation. The project follows the formula-table interpretation in each case.

### 3a. Albedo Hyd 6+ formula

**Sources.** WBH p.110 table specifies the albedo hydrographics modifier as `+(2D − 4) × 0.03`. WBH p.111 worked example for Zed Prime writes `(2D − 3) × 0.03` and computes `(6 − 3) × 0.03 = 0.09`.

**Conflict.** The text formula `(2D − 4) × 0.03` and the worked example `(2D − 3) × 0.03` give different results for the same dice value.

**Chosen interpretation.** **Follow the table** (`(2D − 4) × 0.03`).

**Why.** To reproduce the book's stated 0.09 albedo modifier with the table formula, the harness scripts dice value `7` for the hyd modifier (gives `(7 − 4) × 0.03 = 0.09`). The dice script in the Zed fixture encodes this; the worked example's stated `(6 − 3)` is treated as a typo.

### 3b. Terra reference greenhouse factor

**Sources.** WBH uses `G = 0.36` for Terra-comparison computations. With Terra's `L = 1.0, A = 0.30, AU = 1.0`, the equation `T = 279 × ⁴√(L(1 − A)(1 + G) / d²)` gives `T = 279 × (0.952)^0.25 = 275.6 K` — not Earth's actual ~288 K.

**Conflict.** The book's stated `G = 0.36` does not reproduce real Earth's mean temperature of ~288 K. Earth's actual mean would require `G ≈ 0.62`.

**Chosen interpretation.** **Use `G = 0.36` as the book specifies.** The implementation does not "correct" the Terra reference value to match real Earth.

**Why.** The book's choice of 0.36 is a simplified-model value used consistently throughout the chapter. Substituting 0.62 would break every other Terra-relative computation. The discrepancy with real Earth is a model simplification, not an internal contradiction.

### 3c. WBH p.115 sidebar — Zed Prime WorstLow

**Sources.** WBH p.115 sidebar states Zed Prime's WorstLow temperature as 230 K. Consistent computation using Near/Far AU per the same step-9 formula as normal high/low gives 219 K.

**Conflict.** The book's stated 230 K appears to use base AU (instead of Far AU) for worst-low only — internally inconsistent with how worst-high is computed in the same sidebar.

**Chosen interpretation.** **Compute consistently using Near/Far AU** (yields 319 K worst-high, 219 K worst-low). The Zed fixture pins the computed values; the sidebar's 230 K is documented as a book-internal arithmetic drift.

**Why.** The internal consistency of the formula matters more than matching one stated number that breaks its own pattern.

## Inconsistency 4: WBH p.125 vs p.126 — ResidualSeismicStress density DM

**Sources.** WBH p.125 ResidualSeismicStress formula table:

```text
World is a moon                     DM+1
World has Size 1 or larger moons    DM+1 per moon, max DM+12
Density greater than 1.0            DM+2
Density less than 0.5               DM−1
```

WBH p.126 worked example for Zed Prime (density 1.03):

> "Zed Prime is a Size 5 moon with density 1.03 and is 6.3 billion years old: residual seismic stress is: 5 − 6.3 + 1 (for being a moon) + 1 (for density) = 0.7, rounded down to 0 prior to squaring."

**Conflict.** The table says density > 1.0 → DM+2; the worked example uses DM+1.

**Chosen interpretation.** **Follow the formula table** (density > 1.0 → DM+2).

**Why.** The table is the procedural reference; the worked example's `+1` appears to be transcription drift. Result for Zed Prime: `5 − 6.3 + 1 (moon) + 2 (density) = 1.7 → floor 1 → 1² = 1`. The book's worked example would give 0; pass-2 gives 1. Neither value is large enough to materially affect downstream calculations because TSS is dominated by tidal stress + tidal heating in any tectonically interesting world. The harness fixture asserts 1, with a comment citing the book's worked-example value of 0.

## Inconsistency 5: WBH p.131 — Compatibility worked example contains unsourced "+3"

**Sources.** WBH p.131 Compatibility Rating formula box:

```text
Compatibility Rating = 2D − Biocomplexity/2 + DMs
```

WBH p.131 worked example for Zed Prime:

> "Zed Prime has a biocomplexity rating of 5 and DM+2 for Atmosphere. A roll of 7 becomes a result of 7 + 3 − 2.5 + 2 = 9.5. rounded to 9."

**Conflict.** The worked example shows `7 + 3 − 2.5 + 2 = 9.5` but the formula box has no `+3` addend. The "+3" is unsourced — no documented DM adds to 3, and no secondary table supports the value.

**Chosen interpretation.** **Follow the formula box** (`Compatibility = floor(2D − Biocomplexity/2 + DMs)`).

**Why.** The formula box is the procedural reference. The "+3" cannot be reverse-engineered from any documented DM table or footnote. The book's printed Compatibility=9 for Zed Prime appears to be an unstated DM, a leftover from a prior edition, or a transcription error. Result: Zed Prime gets Compatibility = `7 − 5/2 + 2 = 6.5 → floor → 6`. Native Lifeform Profile becomes `"A576"` (not `"A579"` per the book).

**Verification target divergence noted.** This is the one case where the chosen interpretation does **not** reproduce the canonical IISS form value. Pass 2 accepts this divergence because the alternative (encoding the unsourced `+3` as a constant) would commit the implementation to a magic number with no procedural justification. The harness fixture asserts `"A576"`; the IISS form's `"A579"` is documented as book-internal arithmetic drift.

**Exception to the verification-target rule.** The decision rule at the top of this document says "follow whichever interpretation reproduces the canonical IISS form." This entry deliberately violates that rule, and the rule is therefore a heuristic, not a constitution. The criterion that breaks the tie: when reproducing the form would require implementing an unsourced constant, the formula wins; when reproducing the form follows from a procedural rule that is itself defensible (Inconsistency 6's "narrower band wins"), the form wins.

A referee replicating Zed Prime by hand under the formula gets `"A576"`/`Hab 7`; the form's printed values give `"A579"`/`Hab 7`. The two results differ on Compatibility but agree on Habitability — they cannot both be the verification target. The code picks the formula for Compatibility, the form for Habitability, and accepts the asymmetry. Both divergences are flagged ⚠️ in the harness; neither hides.

### Adjacent finding: WBH p.131 Compatibility table mentions atm codes G and H

The Compatibility DM table on p.131 lists "Atmosphere 0, 1, B, G, or H: DM−8". Atm codes G and H **do not exist** in the standard WBH atm system (range 0–F). They are likely a remnant from Cepheus Engine or an earlier Mongoose supplement. `RollAtmoCode` cannot produce them, so pass-2's `compatibilityAtmDM` simply omits G/H — no DM is applied because the codes can never appear. Documented in the procedure's doc-comment.

## Inconsistency 6: WBH p.132 — gravity DM bands overlap; footnote contradicts worked example

**Sources.** WBH p.132 Habitability DM table — gravity rows:

```text
Gravity less than 0.2:    DM−4
Gravity 0.2–0.7:          DM−2
Gravity 0.4–0.7:          DM−1
Gravity 0.7–0.9:          DM+1
Gravity 1.1–1.4:          DM−1
Gravity 1.4–2.0:          DM−3
Gravity greater than 2.0: DM−6
```

The bands `0.2–0.7` (DM−2) and `0.4–0.7` (DM−1) overlap on `[0.4, 0.7]`.

WBH p.132 footnote: "Assume the worst DM for gravity at the edges of a DM criteria."

WBH p.133 worked example for Zed Prime: "Gravity is only 0.66, so that warrants a DM−1."

**Conflict.** The footnote says use the worst (more-negative) DM at boundaries — gravity 0.66 falls in both bands, so the footnote prescribes DM−2. The worked example uses DM−1 (the narrower 0.4–0.7 band wins).

**Chosen interpretation.** **Follow the worked example.** The narrower band wins for the overlap zone.

**Why.** Zed Prime's printed Habitability rating on the canonical IISS form (pp.141–142) is **7**. With "use worst at edges" the formula gives 6; with "narrower band wins" the formula gives 7, matching the form. The verification target rule (introduced at the top of this document) decides: follow whichever interpretation reproduces the printed canonical value. The footnote is treated as text-vs-form drift; the form wins.

**Encoded interpretation.**

```text
g < 0.2                  → DM−4
0.2 ≤ g < 0.4            → DM−2  (residual of 0.2–0.7 after the narrower band claims [0.4, 0.7))
0.4 ≤ g < 0.7            → DM−1  (narrower band)
0.7 ≤ g ≤ 0.9            → DM+1
0.9 < g ≤ 1.1            → DM 0  (baseline, not in the table)
1.1 < g ≤ 1.4            → DM−1
1.4 < g ≤ 2.0            → DM−3
g > 2.0                  → DM−6
```

Result for Zed Prime: 10 + 0 (size 5) + 0 (atm 6) + 0 (hydro 6) + 0 (no lock) + (−2) (HighK 346 > 323) + 0 (MeanK 300) + 0 (LowK 262) + (−1) (gravity 0.66) = **7** ✓.

## Inconsistency 7: WBH p.106 atmosphere DM — pipeline ordering

WBH p.106 lists `pressure > 2.5 bar → DM−2` in the all-cases-common tidal-lock DMs. But atmospheric pressure is determined by the climate cluster (`ApplyClimatePasses`, Stage 5), which runs **after** rotation/tilt/tidal (`ApplyRotationTilt`, Stage 4). At Stage 4 evaluation time, `body.Atmosphere` is nil and the DM cannot fire — making the line dead code in any single-pass pipeline.

Reordering Stage 4 and Stage 5 isn't viable: temperature (Stage 5) reads `body.AxialTilt` and `body.Eccentricity`, both of which `ApplyTidalLockEffect` can mutate. Atmosphere and tidal lock are mutually dependent.

**Implementation: Stage-5-post re-evaluation cascade.**

1. Stage 4 (`ApplyRotationTilt`) runs tidal lock without the atmosphere DM.
2. `GenerateTidalLock` captures `body.preTidalLockSnapshot` (Eccentricity / AxialTilt / DayLength).
3. Stage 5 (`ApplyClimate`) sets `body.Atmosphere` with the rolled pressure.
4. `ApplyTidalLockReEval` walks bodies with `Atmosphere.Pressure > 2.5`, restores the snapshot, re-runs `GenerateTidalLock` (the DM now fires because `commonTidalLockDMs` sees the pressure), then clears Stage-5 output and re-runs `ApplyClimatePasses` for the affected body so atmosphere/hydrographics/temperature/geology are derived from the corrected tidal-lock outputs.

**Trade-off (acknowledged):** the dice stream consumes both the original Stage-4 tidal-lock rolls and the re-eval's rolls + climate re-run. Deterministic per seed, but a body that goes through the cascade has consumed roughly twice the dice of one that doesn't. Confined to a narrow population — `Pressure > 2.5 bar` plus a captured snapshot from Stage 4. In the 10 regression seeds only `seed_500`'s `A II` hits the cascade.

Affected code: `worlds/tidal_lock_snapshot.go`, `worlds/tidal_lock_reeval.go`, `worlds/tidal_lock.go` (`GenerateTidalLock` capture site), `worlds/generate.go` (pipeline insertion). Tests in `worlds/tidal_lock_reeval_test.go`.

## Naming gotchas (not book contradictions but worth flagging)

### WBH atmosphere code labels for D and E

WBH p.79 labels are unintuitive: **D = "Very Dense"** (2.50–10.0 bar), **E = "Low"** (0.10–0.42 bar, low-pressure oxygen). Pass-1 doc-comments occasionally guessed "Dense" for D or "Ellipsoidal" for E (the latter from a Cepheus Engine variant); both were caught in review. Pass-2 implementation must verify against the canonical label map before naming any code by letter. The full pass-2 map:

```text
0 = None        2-9 = oxygen variants
1 = Trace       A   = Exotic
                B   = Corrosive
                C   = Insidious
                D   = Very Dense
                E   = Low
                F   = Unusual
```

This map lives in `worlds/atmosphere.go::atmosphereLabels` and is the source of truth.

## How the code encodes these decisions

- Each decision is encoded with a doc-comment block citing this document by section.
- Each decision has at least one test that asserts the chosen value and documents the alternative as a comment (e.g., `// Book worked example shows "+3" giving 9; we follow the formula box (= 6) per docs/wbh-inconsistencies.md § Inconsistency 5.`).
- The Zed Prime harness fixture asserts the post-decision values, not the book's printed values, where they differ. The IISS form output is committed as a golden file with the differences explicitly documented.
- No runtime toggle exists for any of these decisions. The cuts list in `design-intent.md` includes "toggles for book inconsistencies" specifically.

## Pattern recognition — when to suspect a new inconsistency

Future implementation work will likely surface further book contradictions. The decision rule is:

1. If a formula and a worked example disagree, compute Zed Prime by hand under each interpretation.
2. Whichever interpretation reproduces the canonical IISS form value on pp.141–142 wins.
3. If neither reproduces the form value, follow the formula (procedural reference) and document the divergence as an additional entry here.
4. Append a new section to this document. Each section is canonical; memory entries for newly-discovered inconsistencies become one-liners pointing back to this document.

The discipline: log every divergence. Never silently fudge the implementation to match an example. If the IISS form is the verification target, encode that. If the form value cannot be reproduced under any reasonable interpretation, that is a significant finding — flag it in the spec and surface it in the harness `t.Logf` (not as a runtime toggle, but as a one-time acknowledgement).
