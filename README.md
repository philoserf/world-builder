# wbh — World Builder's Handbook reference implementation

Go reference implementation of star-system generation procedures from
Mongoose Publishing's _World Builder's Handbook_ (Geir Lanesskog, 2023).

```bash
task test     # run tests
task check    # vet + lint + fmt check
task fmt      # apply formatting
```

The library is the artifact; the CLI is a thin wrapper. See
`docs/specs/2026-05-02-world-builder-design.md` for design rationale.

## Source

Every spec and plan in `docs/` references Mongoose Publishing's _World
Builder's Handbook_ (Geir Lanesskog, 2023) as the canonical authority.
To work with this repo locally, place a copy of the handbook PDF at
`docs/World Builders Handbook.pdf` (gitignored).

## License

MIT — see `LICENSE`. The code is licensed for reuse; the World Builder's
Handbook itself is © Mongoose Publishing and is not redistributed here.
