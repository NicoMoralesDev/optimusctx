# optimusctx

OptimusCtx is a local-first runtime that builds and maintains persistent repository context for coding agents.

## Current status

Phase 2 now includes the repository-local bootstrap and manual refresh path. The command surface is still intentionally small while extraction and query features are built:

- `optimusctx --help`
- `optimusctx version`
- `optimusctx init`
- `optimusctx refresh`
- `optimusctx snippet`

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

The supported local install path for Phase 2 is `go install ./cmd/optimusctx`. Local development can also use `go run ./cmd/optimusctx ...` from this repository. npm or `npx` packaging is not part of the current product scope.

## Smoke test in a fresh temp repository

The reproducible verification path is a disposable Git repository, not the mutable `optimusctx` checkout itself.

```bash
go install ./cmd/optimusctx

tmpdir="$(mktemp -d)"
cd "$tmpdir"
git init

cat <<'EOF' > main.go
package main

func main() {}
EOF

git add main.go
git commit -m "baseline"

optimusctx init
```

Expected results:

- `.optimusctx/` is created under the temp repository
- `optimusctx init` reports the repository root, state directory, refresh generation, and `fresh` freshness

To verify incremental refresh behavior, mutate tracked files in the temp repo and run:

```bash
printf '\nfunc refreshed() {}\n' >> main.go
cat <<'EOF' > helper.go
package main
EOF
optimusctx refresh
```

This fixture-style flow matches the automated integration tests. Running count-based UAT inside the actively changing `optimusctx` worktree is not a reliable way to validate no-op or mutation counts.

## Non-invasive contract

OptimusCtx is being built with an explicit local-first, non-invasive contract:

- project state lives under `.optimusctx/` inside the target repository
- Phase 1 commands do not rewrite instruction files such as `AGENTS.md`, `CLAUDE.md`, or editor settings
- integration guidance is emitted as manual-copy output instead of automatic repository edits
