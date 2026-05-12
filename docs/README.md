# docs/

Project documentation for [world-builder](https://github.com/philoserf/world-builder). The Mongoose Traveller _World Builder's Handbook_ (Geir Lanesskog, 2023) PDF lives here as `World Builders Handbook.pdf` (gitignored — copyright); the rest is project-authored.

## What's authoritative

**`pass-2/`** — the active design + implementation docs. Pass-2 was an architectural rebuild that replaced pass-1; everything on `main` derives from this design. Most-load-bearing entry points:

- [`pass-2/api-surface.md`](pass-2/api-surface.md) — the public API surface and what each entry point does.
- [`pass-2/anti-patterns.md`](pass-2/anti-patterns.md) — the project's "don't-do-this" catalog. Every entry comes from a real incident; check before introducing a new pattern.
- [`pass-2/design-intent.md`](pass-2/design-intent.md) — why pass-2 looks the way it does. Read before proposing structural changes.
- [`pass-2/next-steps.md`](pass-2/next-steps.md) — the path to v1.0 and what's actually open.

Other pass-2 docs cover specific topics (dependency graph, WBH inconsistencies, lessons learned, the clean-every-run plan + sweep catalog, the harness layout). Browse `pass-2/` if you're working in that area.

## What's archive

**`pass-1/`** — historical specs / plans / retrospective from the original pass-1 implementation. Preserved for context only. Pass-2 intentionally diverges from pass-1 on multiple axes (TSS folded into climate, surface-distribution-after-converge, narrower-band-wins for gravity DM, etc.); pass-1's outputs are no longer authoritative.

If you see pass-1 referenced as the source of truth in older notes or commit messages, prefer the pass-2 doc that covers the same area.

## How to use the WBH PDF

The book sits at the project root of `docs/` and is referenced by page number throughout the codebase and these docs (e.g. "WBH p.39 MAO table", "WBH p.16 letter constraints"). Drop your own copy at `docs/World Builders Handbook.pdf` to make those references navigable.
