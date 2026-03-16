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

## Benchmark reruns and methodology verification

Phase 12 adds two benchmark-facing report surfaces on top of the persisted evidence bundle:

- `go run ./cmd/optimusctx eval benchmark export --suite <id> --attempts <n>`
- `go run ./cmd/optimusctx eval benchmark report --suite <id> --attempts <n>`

Benchmark inputs live here:

- suites: `testdata/eval/benchmarks/`
- fixtures: `testdata/eval/fixtures/`

To rerun or inspect the repeated-run methodology from the repository root, use the shipped report surfaces:

```bash
go run ./cmd/optimusctx eval benchmark export --suite go-benchmark-refresh-v1 --attempts 2
go run ./cmd/optimusctx eval benchmark report --suite go-benchmark-refresh-v1 --attempts 2
```

The human-readable report is intentionally narrow and truthful:

- timing and estimated-token comparisons are shown lane by lane
- treatment-side attribution is grouped with BNCH-02-facing labels such as Repository Map, Exact Lookup, L2 Context, and Pack Export
- estimated tokens always use the `bytes_div_4_ceiling` policy
- the report explains workflow-consumed evidence volume, not provider-billed token invoices
- rerun guidance and methodology fingerprint stay visible so reviewers can inspect the same frozen suite again

For targeted verification coverage from the repository root, use:

```bash
go test ./internal/repository ./internal/app ./internal/store/sqlite -run 'TestBenchmarkRepeatedRuns|TestBenchmarkComparisonSummary'
go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkVerificationWorkflow|TestBenchmarkRerunsDeterministic'
```

What Phase 12 reporting proves now:

- the same frozen suites can be rerun repeatedly on the same fixtures
- paired baseline and OptimusCtx arms preserve suite, arm, lane, and attempt identity
- methodology drift such as changed stop conditions is rejected as a verification failure
- workflow-speed and token-attribution evidence are persisted in `.optimusctx/db.sqlite`
- the human-readable report is rendered from the persisted/exported evidence bundle rather than bespoke terminal summaries

What the report still does not prove:

- token attribution by artifact type
- provider-billed token truth
- universal savings beyond the recorded frozen-suite attempts
- statistical significance beyond the explicit reruns you asked it to render

Treat the benchmark report as milestone evidence for the recorded suite and attempts only. If you need to review the raw machine-readable bundle, use `eval benchmark export`; if you need to inspect the operator-facing narrative, use `eval benchmark report`.

## Functional validation evidence

Phase 10 closes the functional milestone with persisted evidence, not a separate `eval report` command. The shipped workflow stays on `optimusctx eval` plus the repo-local artifact tree under `.optimusctx/eval/`.

Requirement-mapped scenario inventory:

- `EVAL-02`: `mcp-go-basic-v1`, `mcp-go-worktree-v1`
- `EVAL-03`: `cli-go-stale-v1`, `mcp-go-degraded-v1`, `mcp-go-recovery-v1`

Rerun any functional scenario from the repository root with the existing command surface:

```bash
go run ./cmd/optimusctx eval --scenario mcp-go-basic-v1
go run ./cmd/optimusctx eval --scenario mcp-go-worktree-v1
go run ./cmd/optimusctx eval --scenario cli-go-stale-v1
go run ./cmd/optimusctx eval --scenario mcp-go-degraded-v1
go run ./cmd/optimusctx eval --scenario mcp-go-recovery-v1
```

Each run persists its copied evidence under the source repository:

- run root: `.optimusctx/eval/run-<id>/`
- copied file artifacts: `.optimusctx/eval/run-<id>/artifacts/`
- per-step stdout or stderr captures: `.optimusctx/eval/run-<id>/<step-id>/`
- SQLite metadata for `eval_runs`, `eval_steps`, and `eval_artifacts`: `.optimusctx/db.sqlite`

The functional suite only reports the shipped contract:

- MCP validation uses `mcp serve`, readiness on `stderr`, `initialize`, `tools/list`, and the shipped query or ops tools
- functional failure-path validation uses stale, partially degraded, and recovered evidence from the existing CLI and MCP surfaces
- there is no MCP `doctor` tool, no MCP `watch` tool, and no new user-facing report mode

To regenerate the requirement coverage summary used in milestone verification, run the targeted tests that read persisted eval evidence:

```bash
go test ./internal/app ./internal/cli ./internal/store/sqlite -run 'TestEvalReportSummaries|TestEvalRequirementCoverageReport'
```

Those tests validate that the latest persisted scenario runs map real scenario IDs, rerun commands, and artifact roots back to `EVAL-02` and `EVAL-03`.

## Non-invasive contract

OptimusCtx is being built with an explicit local-first, non-invasive contract:

- project state lives under `.optimusctx/` inside the target repository
- Phase 1 commands do not rewrite instruction files such as `AGENTS.md`, `CLAUDE.md`, or editor settings
- integration guidance is emitted as manual-copy output instead of automatic repository edits
