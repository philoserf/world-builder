# Pass 2 — Next Steps

**v1.0 shipped 2026-05-12** (tag `v1.0`, GitHub release). This doc catalogs what's actually open after v1.0 and the historical disposition of every pre-v1.0 item.

## What's actually open (post-v1.0)

- **C2 — optional referee knobs.** Four small toggles deferred from pass-2; land if vetting feedback motivates them.
- **C4 — special-object detail.** Detailed BD/D/NS/BH/Pulsar companion physics. Scope TBD.
- **#46 — long functions.** 8 callouts exceeding the 50-line guideline. Opportunistic — touch when next editing.
- **#47 — large files.** 7 callouts exceeding the 300-line guideline. Opportunistic — touch when next editing.
- **C5 — Special Circumstances chapter (WBH pp.147+).** Explicitly out of scope per CLAUDE.md; preserved here for completeness.

Details on each below.

## Historical disposition

### A. Items that required user judgement

- **A0 — pass-1-vs-pass-2 byte comparison.** Dropped pre-v1.0; the within-pass-2 Markdown regression baseline replaces it.
- **A1 — strict `ConvergeClimate` convergence.** Resolved pre-v1.0 (climate is not a fixed point; 2-pass). Renamed to `ApplyClimatePasses` in v1.0 (commit `041704d`, closed #42).
- **A2 — `stars.Group` migration to `stars/`.** Closed won't-fix as GitHub #41 on 2026-05-12. The migration has no functional benefit; the api-surface.md decision was right in principle but the cost in practice doesn't justify it for a working system. Re-open if `stars.Group` starts attracting unrelated coupling.
- **A3 — `stars.GenerateSystemOpts` cuts.** Resolved pre-v1.0 (`WithVariance` and `Accuracy` are book-fidelity load-bearing for worked-example tests, kept indefinitely).
- **A4 — belt-mainworld worked example.** Closed won't-fix as GitHub #51 on 2026-05-12. No canonical WBH source for the values; constructing the fixture would be design fiat. The PART P.B renderer works; a fixture is nice-to-have, not critical. Re-open with concrete values once a belt mainworld actually shows up.

### B. Mechanical items — all resolved pre-v1.0

- **B1** misuse-path tests, **B2** harness.md status, **B3** cmd/world-builder JSON output, **B4** property test expansion — all landed before v1.0.

### C. Strategic / scope items

- **C1 — Merge pass-2 to main.** Merged pre-v1.0 at `a65a412`. Tag `pass-2-final` marks the post-merge state; `pass-1-final` preserves the pre-merge state.
- **C2 — Optional referee knobs.** **Still open.** Four small toggles named in `design-intent.md` § Post-parity work:
  - Rare Earth Universe Variant
  - Optional any-oxygen-atm biomass floor
  - Optional Insidious DE hazard rule's optional branch
  - `-mainworld <designation>` override flag

  Each is a small individual change (an `Opts` field on the relevant `Generate*` or a `cmd/world-builder` flag). Effort: ~1 day for all four, including tests and a `cmd/world-builder -help` update. Value: medium — referee-facing. Drive by vetting feedback.

- **C3 — Notable Features Markdown block.** **Shipped in v1.0** (commit `89cbcd5`). Five sections above the IISS forms: tidal-lock zones, cold snaps, crush worlds, taint chains, mainworld habitability rationale.

- **C4 — Special-object detail (BD / D / NS / BH / Pulsar companions).** **Still open.** Currently companion special objects get minimum-useful stub values (kind/mass/age) but no detailed physics (accretion, degenerate-matter equations, jet behaviour, white-dwarf cooling curves, pulsar spin-down). Effort: unknown — depends on which objects to detail and to what depth. WBH's Special Circumstances chapter has the rules, so this overlaps with C5's scope decision. Value: variable; black-hole and neutron-star companions cause real campaign effects, white dwarfs are common in older systems.

- **C5 — Special Circumstances chapter (WBH pp.147+).** **Explicitly out of scope.** Social characteristics, government, technology, special-object population physics, etc. Effort: comparable to all of pass-1's physical-rules work. CLAUDE.md is explicit: "Do not start work in those chapters; do not add code that anticipates them." Pass-2 + v1.0 honour this.

## GitHub backlog (post-v1.0)

Two opportunistic tech-debt issues remain:

- **#46** — long functions exceed 50-line smell threshold (8 callouts in `stars/system.go`, `stars/survey.go`, `worlds/temperature.go`, etc.). Each is one cohesive WBH procedure with sequential phases — length isn't symptomatic of mixed responsibility, but extraction into named phase helpers would improve diffability. Touch opportunistically when next editing.
- **#47** — large files exceed 300-line A.7 guideline (7 callouts, mostly in `worlds/`). Same opportunistic disposition.

Neither blocks v1.x or later work.
