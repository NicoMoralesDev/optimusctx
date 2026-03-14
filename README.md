# optimusctx

OptimusCtx is a local-first runtime that builds and maintains persistent repository context for coding agents.

## Current status

Phase 1 is bootstrapping the Go CLI and the repository-local state model. The command surface is intentionally small while the discovery and persistence layers are being built:

- `optimusctx --help`
- `optimusctx version`
- `optimusctx init` (planned in this phase)
- `optimusctx snippet` (planned in this phase)

## Install locally

Use `go install` to build the binary without mutating any target repository:

```bash
go install ./cmd/optimusctx
```

For local development from this repository:

```bash
go run ./cmd/optimusctx --help
go run ./cmd/optimusctx version
```

## Non-invasive contract

OptimusCtx is being built with an explicit local-first, non-invasive contract:

- project state lives under `.optimusctx/` inside the target repository
- Phase 1 commands do not rewrite instruction files such as `AGENTS.md`, `CLAUDE.md`, or editor settings
- integration guidance is emitted as manual-copy output instead of automatic repository edits
