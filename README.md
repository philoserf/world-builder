# wbh — World Builder's Handbook reference implementation

Go reference implementation of star-system generation procedures from
Mongoose Publishing's _World Builder's Handbook_ (Geir Lanesskog, 2023).

```bash
just test     # run tests
just check    # vet + lint + fmt check
just fmt      # apply formatting
```

The library is the artifact; the CLI is a thin wrapper. See
`docs/superpowers/specs/2026-05-02-world-builder-design.md` at the repo
root for design rationale.
