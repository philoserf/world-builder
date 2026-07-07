# docs/

Project documentation for [world-builder](https://github.com/philoserf/world-builder). The Mongoose Traveller _World Builder's Handbook_ (Geir Lanesskog, 2023) PDF lives here as `World Builders Handbook.pdf` (gitignored — copyright); the rest is project-authored.

## Evergreen — what's authoritative

Design, reference, and living docs live at this directory's root:

- [`walkthrough.md`](walkthrough.md) — linear code tour from CLI entry through the pipeline to the rendered Markdown. Best starting point for new contributors.
- [`design-intent.md`](design-intent.md) — why the code looks the way it does. Read before proposing structural changes.
- [`api-surface.md`](api-surface.md) — every public signature with its rationale.
- [`dependency-graph.md`](dependency-graph.md) — every value, its inputs, the fixed-point clusters.
- [`anti-patterns.md`](anti-patterns.md) — don't-do-this catalog. Every entry comes from a real incident; check before introducing a new pattern.
- [`harness.md`](harness.md) — fixture catalog with status indicators.
- [`wbh-inconsistencies.md`](wbh-inconsistencies.md) — six book-internal divergences with chosen interpretations.
- [`summary.md`](summary.md) — one-page overview of what was built.
- [`next-steps.md`](next-steps.md) — living post-v1.0 open-items list.
- [`rebuild-spec.md`](rebuild-spec.md) — specification for a from-scratch third implementation (the "rebuild"), reframing both shipped passes as the exploratory first pass. Forward-looking; not yet built.

## Historical — preserved for context

[`history/`](history/) holds artifacts from the project's two implementation phases. Not authoritative; check the evergreen docs first.

- **Pass-1 archive** (`history/pass-1-specs/`, `history/pass-1-plans/`, `history/pass-1-retrospective/`): dated, chapter-numbered docs from the original implementation. Buildable via `git checkout pass-1-final`.
- **Pass-2 retrospective** (`history/lessons-learned.md`, `history/plan-clean-every-run.md`, `history/generator-error-catalog.md`, `history/spike-findings.md`, `history/allbodies-migration.md`): rebuild-era plans and retrospective notes. The evergreen docs at root are what shipped in `v1.0`.

## WBH PDF

The book sits at `docs/World Builders Handbook.pdf` and is referenced by page number throughout the codebase and these docs (e.g. "WBH p.39 MAO table", "WBH p.16 letter constraints"). Drop your own copy at that path to make the references navigable.
