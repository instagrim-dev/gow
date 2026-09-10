# AGENTS.md

`newf` is a research-search system: it treats failures as a first-class sample
space and generates frontier proposals against surviving failure invariants.
Full doctrine: `README.md` and `docs/`.

## Hard constraints (read first)

- **Never put SQL or Cobra in `internal/domain`.** SQLite lives behind `internal/store`; CLI wiring stays in `cmd/newf`.
- **Never add provider coupling to a domain package.** Provider adapters live in `internal/provider`.
- **Never silently promote epistemic status** (`Hypothesis`→`Evidence`, `Candidate`→`Established`). Preserve the weaker type; record the limitation.
- **Before you call work done:** run `go build ./...`, `go test ./...`, and `gofmt -l .` (output must be empty). Add a test at the boundary where behavior is introduced.
- **Read `README.md` and the relevant `docs/` file before changing architecture.**

## When the tree is broken

- Build/test failure in code you did not touch: **report it**; fix forward only within your slice.
- Overlapping concurrent edits: reconcile toward your task; do not revert or stash others' work.

## Why (background)

The governing loop is `failure-space -> failure invariant -> invariant break ->
partial success -> success invariant -> generalized frontier`. See `docs/`.
