# Next Steps (post-v1.0)

Open work, ordered most-deterministic to least.

## #46 — long functions

Eight callouts exceed the 50-line guideline (`stars/system.go`, `stars/survey.go`, `worlds/temperature.go`, others). Each is one cohesive WBH procedure with sequential phases; length isn't mixed responsibility, but extracting named phase helpers would improve diffability. Touch opportunistically.

## #47 — large files

Seven callouts exceed the 300-line A.7 guideline, mostly in `worlds/`. Same opportunistic disposition as #46.

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
