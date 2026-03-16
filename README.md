# optimusctx

OptimusCtx is a local-first runtime that builds and maintains persistent repository context for coding agents.

## Current status

The current command surface covers repository bootstrap, refresh, diagnostics, export, MCP serving, eval harness runs, and watch-mode support:

- `optimusctx --help`
- `optimusctx version`
- `optimusctx doctor`
- `optimusctx eval --scenario <id>`
- `optimusctx init`
- `optimusctx install`
- `optimusctx mcp serve`
- `optimusctx pack export`
- `optimusctx refresh`
- `optimusctx snippet`
- `optimusctx watch`

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

## Evaluation harness

Phase 9 adds one committed, rerunnable evaluation path that uses versioned fixtures and versioned scenario definitions instead of hidden manual setup.

Evaluation inputs live in this repository:

- fixtures: `testdata/eval/fixtures/`
- scenarios: `testdata/eval/scenarios/`

Each scenario names a stable scenario ID, the fixture repository version to materialize, and the ordered CLI steps to execute. Reruns always reconstruct a fresh temp workspace from the fixture tree before running the scenario, so you do not hand-edit prior state between runs.

Run a scenario by ID from the repository root:

```bash
go run ./cmd/optimusctx eval --scenario <scenario-id>
```

Example:

```bash
go run ./cmd/optimusctx eval --scenario persist
```

Or run an explicit scenario file:

```bash
go run ./cmd/optimusctx eval --scenario-file testdata/eval/scenarios/persist.json
```

The current Phase 9 harness is intentionally narrow:

- fixtures and scenarios are versioned inputs checked into the repo
- `eval` executes the real shipped CLI boundary inside the materialized workspace
- reruns rebuild fixture state automatically instead of depending on leftover `.optimusctx` directories or manual repository edits
- evidence for each run is persisted under the source repository at `.optimusctx/eval/run-<id>/`
- SQLite metadata for runs, steps, and copied artifacts is stored in `.optimusctx/db.sqlite`

This is the foundation for later validation and benchmark phases. Phase 10 can add more scenario depth on the same harness instead of replacing it with ad hoc scripts.

## Non-invasive contract

OptimusCtx is being built with an explicit local-first, non-invasive contract:

- project state lives under `.optimusctx/` inside the target repository
- Phase 1 commands do not rewrite instruction files such as `AGENTS.md`, `CLAUDE.md`, or editor settings
- integration guidance is emitted as manual-copy output instead of automatic repository edits
