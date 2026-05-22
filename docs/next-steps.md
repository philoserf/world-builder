# Next Steps (post-v1.0)

Open work, ordered most-deterministic to least.

The 50-line / 300-line guidelines from A.7 (issues #46, #47) are **triggers, not caps** — they prompt the question "is this still one concern?" rather than mandate a split. Both issues are closed as policy. Touch opportunistically when the cohesion answer changes; do not split for the line count alone.

## C2 — optional referee knobs

Four toggles named in `design-intent.md` § Post-parity work:

- Rare Earth Universe Variant
- Optional any-oxygen-atm biomass floor
- Optional Insidious DE hazard rule's optional branch
- `-mainworld <designation>` override flag

Each is an `Opts` field on the relevant `Generate*` or a `cmd/world-builder` flag. Total effort ~1 day with tests and `-help` update. Drive by vetting feedback.

## C4 — special-object companion detail

BD/D/NS/BH/Pulsar companions currently get stub values (kind/mass/age) but no detailed physics — accretion, degenerate-matter equations, jets, white-dwarf cooling, pulsar spin-down. Effort depends on which objects and what depth. Overlaps with C5's scope decision since the rules live in WBH's Special Circumstances chapter. Black-hole and neutron-star companions cause real campaign effects; white dwarfs are common in older systems.

## C5 — Special Circumstances chapter (WBH pp.147+)

Explicitly out of scope per `CLAUDE.md`. Listed for completeness. Social characteristics, government, technology, special-object population physics. Effort comparable to all of pass-1's physical-rules work.
